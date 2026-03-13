package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/queueprim/internal/model"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLiteStore implements Store using an embedded SQLite database.
type SQLiteStore struct {
	db *sql.DB
	// dequeueMu serializes DequeueJob calls within a process. SQLite with WAL
	// mode serializes writes at the database level, but we hold this mutex
	// across the SELECT+UPDATE pair to prevent two goroutines from both
	// reading the same pending job before either write commits.
	dequeueMu sync.Mutex
}

// New opens the database at the given path, runs pending migrations, and
// returns a ready-to-use store.
func New(dbPath string) (*SQLiteStore, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	if err := db.Migrate(database, migrations, "migrations"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &SQLiteStore{db: database}, nil
}

// NewFromDB wraps an existing *sql.DB connection. Useful for tests that
// want to control the database lifecycle (e.g., in-memory databases).
func NewFromDB(database *sql.DB) (*SQLiteStore, error) {
	if err := db.Migrate(database, migrations, "migrations"); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &SQLiteStore{db: database}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ---------------------------------------------------------------------------
// EnqueueJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) EnqueueJob(ctx context.Context, job *model.Job) error {
	if err := job.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}

	job.ID = model.NewJobID()
	job.Status = model.StatusPending
	job.AttemptCount = 0

	if job.Priority == "" {
		job.Priority = model.PriorityNormal
	}
	rank := model.PriorityRank(job.Priority)

	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now
	// If no delay was set, make the job immediately available.
	if job.VisibleAfter.IsZero() {
		job.VisibleAfter = now
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO jobs (
			id, queue, type, priority, priority_rank, payload, status,
			visible_after, attempt_count, max_retries, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Queue, nullString(job.Type), string(job.Priority), rank,
		string(job.Payload), string(job.Status),
		job.VisibleAfter.UTC(), job.AttemptCount, job.MaxRetries,
		job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting job: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// DequeueJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) DequeueJob(ctx context.Context, queue, worker, jobType string, timeout time.Duration) (*model.Job, error) {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	if worker == "" {
		worker = "unknown"
	}

	s.dequeueMu.Lock()
	defer s.dequeueMu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// SELECT the next available job: pending, visible, ordered by priority then FIFO.
	query := `
		SELECT id FROM jobs
		WHERE queue = ? AND status = 'pending' AND visible_after <= ?`
	args := []interface{}{queue, time.Now().UTC()}
	if jobType != "" {
		query += " AND type = ?"
		args = append(args, jobType)
	}
	query += " ORDER BY priority_rank ASC, created_at ASC LIMIT 1"

	var id string
	err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, ErrEmpty
	}
	if err != nil {
		return nil, fmt.Errorf("selecting job: %w", err)
	}

	// Atomically claim the job within the same transaction.
	now := time.Now().UTC()
	visibleAfter := now.Add(timeout)
	_, err = tx.ExecContext(ctx, `
		UPDATE jobs
		SET status = 'claimed', claimed_by = ?, claimed_at = ?,
		    visible_after = ?, attempt_count = attempt_count + 1, updated_at = ?
		WHERE id = ?`,
		worker, now, visibleAfter, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("claiming job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing dequeue: %w", err)
	}

	return s.GetJob(ctx, id)
}

// ---------------------------------------------------------------------------
// CompleteJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) CompleteJob(ctx context.Context, id string, output []byte) error {
	var currentStatus string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = ?", id).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking job: %w", err)
	}
	if currentStatus != string(model.StatusClaimed) {
		return fmt.Errorf("%w: cannot complete a job with status %q", ErrInvalidTransition, currentStatus)
	}

	now := time.Now().UTC()
	var outputStr *string
	if len(output) > 0 {
		s := string(output)
		outputStr = &s
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = 'done', completed_at = ?, output = ?, updated_at = ?
		WHERE id = ?`,
		now, outputStr, now, id,
	)
	if err != nil {
		return fmt.Errorf("completing job: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// FailJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) FailJob(ctx context.Context, id, reason string, forceDeadLetter bool) error {
	var currentStatus string
	var attemptCount, maxRetries int
	err := s.db.QueryRowContext(ctx,
		"SELECT status, attempt_count, max_retries FROM jobs WHERE id = ?", id,
	).Scan(&currentStatus, &attemptCount, &maxRetries)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking job: %w", err)
	}
	if currentStatus != string(model.StatusClaimed) {
		return fmt.Errorf("%w: cannot fail a job with status %q", ErrInvalidTransition, currentStatus)
	}

	now := time.Now().UTC()

	// Determine next status: retry or dead.
	nextStatus := model.StatusDead
	var nextVisibleAfter interface{}
	if !forceDeadLetter && attemptCount <= maxRetries {
		// Retries remain — return to pending.
		nextStatus = model.StatusPending
		nextVisibleAfter = now
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = ?, failure_reason = ?, completed_at = ?,
		    claimed_by = NULL, claimed_at = NULL, visible_after = COALESCE(?, visible_after),
		    updated_at = ?
		WHERE id = ?`,
		string(nextStatus), reason, now, nextVisibleAfter, now, id,
	)
	if err != nil {
		return fmt.Errorf("failing job: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ReleaseJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ReleaseJob(ctx context.Context, id string) error {
	var currentStatus string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM jobs WHERE id = ?", id).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking job: %w", err)
	}
	if currentStatus != string(model.StatusClaimed) {
		return fmt.Errorf("%w: cannot release a job with status %q", ErrInvalidTransition, currentStatus)
	}

	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = 'pending', claimed_by = NULL, claimed_at = NULL,
		    visible_after = ?, updated_at = ?
		WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("releasing job: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ExtendJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ExtendJob(ctx context.Context, id string, by time.Duration) error {
	if by <= 0 {
		return fmt.Errorf("extension duration must be positive")
	}

	var currentStatus string
	var visibleAfter time.Time
	err := s.db.QueryRowContext(ctx,
		"SELECT status, visible_after FROM jobs WHERE id = ?", id,
	).Scan(&currentStatus, &visibleAfter)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking job: %w", err)
	}
	if currentStatus != string(model.StatusClaimed) {
		return fmt.Errorf("%w: cannot extend a job with status %q", ErrInvalidTransition, currentStatus)
	}

	newVisible := visibleAfter.Add(by)
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		"UPDATE jobs SET visible_after = ?, updated_at = ? WHERE id = ?",
		newVisible.UTC(), now, id,
	)
	if err != nil {
		return fmt.Errorf("extending job: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PeekJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) PeekJob(ctx context.Context, queue string) (*model.Job, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM jobs
		WHERE queue = ? AND status = 'pending' AND visible_after <= ?
		ORDER BY priority_rank ASC, created_at ASC LIMIT 1`,
		queue, time.Now().UTC(),
	).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, ErrEmpty
	}
	if err != nil {
		return nil, fmt.Errorf("peeking queue: %w", err)
	}
	return s.GetJob(ctx, id)
}

// ---------------------------------------------------------------------------
// GetJob
// ---------------------------------------------------------------------------

func (s *SQLiteStore) GetJob(ctx context.Context, id string) (*model.Job, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, queue, type, priority, payload, status,
		       claimed_by, claimed_at, visible_after, completed_at,
		       output, failure_reason, attempt_count, max_retries,
		       created_at, updated_at
		FROM jobs WHERE id = ?`, id)
	return scanJob(row)
}

// ---------------------------------------------------------------------------
// ListJobs
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListJobs(ctx context.Context, filter *model.Filter) ([]*model.Job, error) {
	query, args := buildListQuery(filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*model.Job
	for rows.Next() {
		j, err := scanJobFromRows(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating jobs: %w", err)
	}
	return jobs, nil
}

func buildListQuery(filter *model.Filter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.Queue != "" {
			conditions = append(conditions, "queue = ?")
			args = append(args, filter.Queue)
		}
		if filter.Status != nil {
			conditions = append(conditions, "status = ?")
			args = append(args, string(*filter.Status))
		}
		if filter.Type != "" {
			conditions = append(conditions, "type = ?")
			args = append(args, filter.Type)
		}
		if filter.OlderThan > 0 {
			cutoff := time.Now().UTC().Add(-filter.OlderThan)
			conditions = append(conditions, "created_at < ?")
			args = append(args, cutoff)
		}
	}

	query := `SELECT id, queue, type, priority, payload, status,
		       claimed_by, claimed_at, visible_after, completed_at,
		       output, failure_reason, attempt_count, max_retries,
		       created_at, updated_at FROM jobs`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"
	return query, args
}

// ---------------------------------------------------------------------------
// ListQueues
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListQueues(ctx context.Context) ([]model.QueueInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT queue,
		       SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'claimed' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'done'    THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'failed'  THEN 1 ELSE 0 END),
		       SUM(CASE WHEN status = 'dead'    THEN 1 ELSE 0 END)
		FROM jobs
		GROUP BY queue
		ORDER BY queue`)
	if err != nil {
		return nil, fmt.Errorf("querying queues: %w", err)
	}
	defer rows.Close()

	var queues []model.QueueInfo
	for rows.Next() {
		var qi model.QueueInfo
		if err := rows.Scan(&qi.Queue, &qi.Pending, &qi.Claimed, &qi.Done, &qi.Failed, &qi.Dead); err != nil {
			return nil, fmt.Errorf("scanning queue info: %w", err)
		}
		queues = append(queues, qi)
	}
	return queues, rows.Err()
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Stats(ctx context.Context) (*model.Stats, error) {
	var stats model.Stats
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'claimed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'done'    THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'failed'  THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'dead'    THEN 1 ELSE 0 END), 0)
		FROM jobs`,
	).Scan(
		&stats.TotalPending, &stats.TotalClaimed,
		&stats.TotalDone, &stats.TotalFailed, &stats.TotalDead,
	)
	if err != nil {
		return nil, fmt.Errorf("querying stats: %w", err)
	}
	return &stats, nil
}

// ---------------------------------------------------------------------------
// PurgeJobs
// ---------------------------------------------------------------------------

func (s *SQLiteStore) PurgeJobs(ctx context.Context, queue string, status model.Status, olderThan time.Duration) (int, error) {
	query := "DELETE FROM jobs WHERE queue = ? AND status = ?"
	args := []interface{}{queue, string(status)}

	if olderThan > 0 {
		cutoff := time.Now().UTC().Add(-olderThan)
		query += " AND created_at < ?"
		args = append(args, cutoff)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("purging jobs: %w", err)
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ExportJobs(ctx context.Context, queue string) ([]*model.Job, error) {
	filter := &model.Filter{}
	if queue != "" {
		filter.Queue = queue
	}
	return s.ListJobs(ctx, filter)
}

func (s *SQLiteStore) ImportJobs(ctx context.Context, jobs []*model.Job) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, j := range jobs {
		rank := model.PriorityRank(j.Priority)
		var outputStr *string
		if len(j.Output) > 0 {
			s := string(j.Output)
			outputStr = &s
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO jobs (
				id, queue, type, priority, priority_rank, payload, status,
				claimed_by, claimed_at, visible_after, completed_at,
				output, failure_reason, attempt_count, max_retries,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			j.ID, j.Queue, nullString(j.Type), string(j.Priority), rank,
			string(j.Payload), string(j.Status),
			j.ClaimedBy, j.ClaimedAt, j.VisibleAfter.UTC(), j.CompletedAt,
			outputStr, j.FailureReason, j.AttemptCount, j.MaxRetries,
			j.CreatedAt, j.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("importing job %s: %w", j.ID, err)
		}
	}
	return tx.Commit()
}

// ---------------------------------------------------------------------------
// SweepExpiredClaims
// ---------------------------------------------------------------------------

// SweepExpiredClaims releases claimed jobs whose visibility timeout has
// expired, returning them to pending. This is the crash recovery mechanism:
// if a worker dies holding a claim, the job is automatically re-queued.
func (s *SQLiteStore) SweepExpiredClaims(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = 'pending', claimed_by = NULL, claimed_at = NULL,
		    visible_after = ?, updated_at = ?
		WHERE status = 'claimed' AND visible_after < ?`,
		now, now, now)
	if err != nil {
		return 0, fmt.Errorf("sweeping expired claims: %w", err)
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// scanJob reads a job from a *sql.Row, returning ErrNotFound for missing rows.
func scanJob(row *sql.Row) (*model.Job, error) {
	var j model.Job
	var jobType sql.NullString
	var claimedBy sql.NullString
	var claimedAt sql.NullTime
	var completedAt sql.NullTime
	var output sql.NullString
	var failureReason sql.NullString
	var payloadStr string

	err := row.Scan(
		&j.ID, &j.Queue, &jobType, &j.Priority, &payloadStr, &j.Status,
		&claimedBy, &claimedAt, &j.VisibleAfter, &completedAt,
		&output, &failureReason, &j.AttemptCount, &j.MaxRetries,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning job: %w", err)
	}

	j.Payload = []byte(payloadStr)
	if jobType.Valid {
		j.Type = jobType.String
	}
	if claimedBy.Valid {
		j.ClaimedBy = &claimedBy.String
	}
	if claimedAt.Valid {
		t := claimedAt.Time
		j.ClaimedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		j.CompletedAt = &t
	}
	if output.Valid {
		j.Output = []byte(output.String)
	}
	if failureReason.Valid {
		j.FailureReason = &failureReason.String
	}
	return &j, nil
}

// scanJobFromRows reads a job from a *sql.Rows cursor.
func scanJobFromRows(rows *sql.Rows) (*model.Job, error) {
	var j model.Job
	var jobType sql.NullString
	var claimedBy sql.NullString
	var claimedAt sql.NullTime
	var completedAt sql.NullTime
	var output sql.NullString
	var failureReason sql.NullString
	var payloadStr string

	err := rows.Scan(
		&j.ID, &j.Queue, &jobType, &j.Priority, &payloadStr, &j.Status,
		&claimedBy, &claimedAt, &j.VisibleAfter, &completedAt,
		&output, &failureReason, &j.AttemptCount, &j.MaxRetries,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning job: %w", err)
	}

	j.Payload = []byte(payloadStr)
	if jobType.Valid {
		j.Type = jobType.String
	}
	if claimedBy.Valid {
		j.ClaimedBy = &claimedBy.String
	}
	if claimedAt.Valid {
		t := claimedAt.Time
		j.ClaimedAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		j.CompletedAt = &t
	}
	if output.Valid {
		j.Output = []byte(output.String)
	}
	if failureReason.Valid {
		j.FailureReason = &failureReason.String
	}
	return &j, nil
}

// nullString converts an empty Go string to a SQL NULL. Useful for optional
// TEXT columns like type, claimed_by, etc.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
