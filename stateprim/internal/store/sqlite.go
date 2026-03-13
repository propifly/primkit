package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/stateprim/internal/model"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLiteStore implements Store using an embedded SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// New opens the database at the given path, runs any pending migrations,
// and returns a ready-to-use store.
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

// NewFromDB wraps an existing *sql.DB connection. Useful for tests.
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
// Set (upsert)
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Set(ctx context.Context, record *model.Record) error {
	if err := record.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()

	// Check if an existing record is immutable before upserting.
	var existing struct {
		immutable bool
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT immutable FROM records WHERE namespace = ? AND key = ?`,
		record.Namespace, record.Key,
	).Scan(&existing.immutable)
	if err == nil && existing.immutable {
		return ErrImmutable
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (namespace, key) DO UPDATE SET
		   value = excluded.value,
		   updated_at = excluded.updated_at`,
		record.Namespace, record.Key, string(record.Value), record.Immutable,
		now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upserting record: %w", err)
	}

	record.UpdatedAt = now
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}

	return nil
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Get(ctx context.Context, namespace, key string) (*model.Record, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT namespace, key, value, immutable, created_at, updated_at
		 FROM records WHERE namespace = ? AND key = ?`,
		namespace, key,
	)
	return scanRecord(row)
}

// ---------------------------------------------------------------------------
// Has (existence check)
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Has(ctx context.Context, namespace, key string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM records WHERE namespace = ? AND key = ?`,
		namespace, key,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking existence: %w", err)
	}
	return count > 0, nil
}

// ---------------------------------------------------------------------------
// SetIfNew (atomic create-if-not-exists)
// ---------------------------------------------------------------------------

func (s *SQLiteStore) SetIfNew(ctx context.Context, record *model.Record) error {
	if err := record.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()
	ts := now.Format(time.RFC3339Nano)

	result, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		record.Namespace, record.Key, string(record.Value), record.Immutable, ts, ts,
	)
	if err != nil {
		return fmt.Errorf("set-if-new: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAlreadyExists
	}

	record.CreatedAt = now
	record.UpdatedAt = now
	return nil
}

// ---------------------------------------------------------------------------
// Append (immutable, auto-key)
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Append(ctx context.Context, namespace string, value []byte) (*model.Record, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	// Validate that value is valid JSON.
	if !json.Valid(value) {
		return nil, fmt.Errorf("value must be valid JSON")
	}

	now := time.Now().UTC()
	key := generateAppendKey(now)
	ts := now.Format(time.RFC3339Nano)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, TRUE, ?, ?)`,
		namespace, key, string(value), ts, ts,
	)
	if err != nil {
		return nil, fmt.Errorf("appending record: %w", err)
	}

	return &model.Record{
		Namespace: namespace,
		Key:       key,
		Value:     json.RawMessage(value),
		Immutable: true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Delete(ctx context.Context, namespace, key string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM records WHERE namespace = ? AND key = ?`,
		namespace, key,
	)
	if err != nil {
		return fmt.Errorf("deleting record: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Query
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Query(ctx context.Context, filter *model.QueryFilter) ([]*model.Record, int, error) {
	query := strings.Builder{}
	args := []interface{}{}

	if filter.CountOnly {
		query.WriteString(`SELECT COUNT(*) FROM records WHERE namespace = ?`)
	} else {
		query.WriteString(`SELECT namespace, key, value, immutable, created_at, updated_at
		                    FROM records WHERE namespace = ?`)
	}
	args = append(args, filter.Namespace)

	if filter.KeyPrefix != "" {
		query.WriteString(` AND key LIKE ?`)
		args = append(args, filter.KeyPrefix+"%")
	}

	if filter.Since > 0 {
		cutoff := time.Now().UTC().Add(-filter.Since).Format(time.RFC3339Nano)
		query.WriteString(` AND updated_at >= ?`)
		args = append(args, cutoff)
	}

	if filter.CountOnly {
		var count int
		err := s.db.QueryRowContext(ctx, query.String(), args...).Scan(&count)
		if err != nil {
			return nil, 0, fmt.Errorf("counting records: %w", err)
		}
		return nil, count, nil
	}

	query.WriteString(` ORDER BY created_at DESC`)

	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying records: %w", err)
	}
	defer rows.Close()

	var records []*model.Record
	for rows.Next() {
		r, err := scanRecordRows(rows)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, r)
	}
	if records == nil {
		records = []*model.Record{}
	}

	return records, len(records), nil
}

// ---------------------------------------------------------------------------
// Purge
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Purge(ctx context.Context, namespace, olderThan string) (int, error) {
	d, err := parseDuration(olderThan)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}

	cutoff := time.Now().UTC().Add(-d).Format(time.RFC3339Nano)
	result, execErr := s.db.ExecContext(ctx,
		`DELETE FROM records WHERE namespace = ? AND updated_at < ?`,
		namespace, cutoff,
	)
	if execErr != nil {
		return 0, fmt.Errorf("purging records: %w", execErr)
	}

	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// ---------------------------------------------------------------------------
// ListNamespaces
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT namespace, COUNT(*) as count FROM records
		 GROUP BY namespace ORDER BY namespace`)
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []model.NamespaceInfo
	for rows.Next() {
		var ns model.NamespaceInfo
		if err := rows.Scan(&ns.Namespace, &ns.Count); err != nil {
			return nil, fmt.Errorf("scanning namespace: %w", err)
		}
		namespaces = append(namespaces, ns)
	}
	if namespaces == nil {
		namespaces = []model.NamespaceInfo{}
	}

	return namespaces, nil
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func (s *SQLiteStore) Stats(ctx context.Context) (*model.Stats, error) {
	var stats model.Stats
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COUNT(DISTINCT namespace) FROM records`,
	).Scan(&stats.TotalRecords, &stats.TotalNamespaces)
	if err != nil {
		return nil, fmt.Errorf("fetching stats: %w", err)
	}
	return &stats, nil
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ExportRecords(ctx context.Context, namespace string) ([]*model.Record, error) {
	query := `SELECT namespace, key, value, immutable, created_at, updated_at FROM records`
	args := []interface{}{}

	if namespace != "" {
		query += ` WHERE namespace = ?`
		args = append(args, namespace)
	}
	query += ` ORDER BY namespace, key`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exporting records: %w", err)
	}
	defer rows.Close()

	var records []*model.Record
	for rows.Next() {
		r, err := scanRecordRows(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	if records == nil {
		records = []*model.Record{}
	}

	return records, nil
}

func (s *SQLiteStore) ImportRecords(ctx context.Context, records []*model.Record) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range records {
		_, err := stmt.ExecContext(ctx,
			r.Namespace, r.Key, string(r.Value), r.Immutable,
			r.CreatedAt.Format(time.RFC3339Nano),
			r.UpdatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("importing record %s/%s: %w", r.Namespace, r.Key, err)
		}
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// scanRecord scans a single record from a *sql.Row.
func scanRecord(row *sql.Row) (*model.Record, error) {
	var r model.Record
	var value string
	var createdAt, updatedAt string

	err := row.Scan(&r.Namespace, &r.Key, &value, &r.Immutable, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning record: %w", err)
	}

	r.Value = json.RawMessage(value)
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &r, nil
}

// scanRecordRows scans a single record from *sql.Rows.
func scanRecordRows(rows *sql.Rows) (*model.Record, error) {
	var r model.Record
	var value string
	var createdAt, updatedAt string

	err := rows.Scan(&r.Namespace, &r.Key, &value, &r.Immutable, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning record: %w", err)
	}

	r.Value = json.RawMessage(value)
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &r, nil
}

// generateAppendKey creates a key for append records: a_<timestamp>_<rand>.
// The timestamp portion makes keys monotonically ordered; the random suffix
// prevents collisions.
func generateAppendKey(t time.Time) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("a_%s_%s",
		t.Format("20060102T150405"),
		hex.EncodeToString(b),
	)
}

// parseDuration handles both Go-style durations (24h, 30m) and short-form
// day notation (7d, 30d).
func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
