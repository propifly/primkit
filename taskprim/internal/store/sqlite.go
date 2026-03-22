package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/taskprim/internal/model"
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

// NewFromDB wraps an existing *sql.DB connection. Useful for tests that want
// to control the database lifecycle (e.g., in-memory databases).
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
// CreateTask
// ---------------------------------------------------------------------------

func (s *SQLiteStore) CreateTask(ctx context.Context, task *model.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	task.ID = model.NewTaskID()
	task.State = model.StateOpen
	task.Created = time.Now().UTC()
	task.Updated = task.Created

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO tasks (id, list, what, source, state, waiting_on, parent_id, context, created, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.List, task.What, task.Source, task.State,
		task.WaitingOn, task.ParentID, task.Context,
		task.Created, task.Updated,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	if err := insertLabels(ctx, tx, task.ID, task.Labels); err != nil {
		return err
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// GetTask
// ---------------------------------------------------------------------------

func (s *SQLiteStore) GetTask(ctx context.Context, id string) (*model.Task, error) {
	task, err := scanTask(s.db.QueryRowContext(ctx, `
		SELECT id, list, what, source, state, waiting_on, parent_id, context,
		       created, updated, resolved_at, resolved_reason
		FROM tasks WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	task.Labels = labels

	return task, nil
}

// ---------------------------------------------------------------------------
// ListTasks
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListTasks(ctx context.Context, filter *model.Filter) ([]*model.Task, error) {
	query, args := buildListQuery(filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer rows.Close()
	return scanTasksWithLabels(ctx, s.db, rows)
}

// buildListQuery constructs a SELECT query with WHERE clauses based on the
// filter. Each filter field appends a condition and a parameter placeholder.
func buildListQuery(filter *model.Filter) (string, []interface{}) {
	var (
		conditions []string
		args       []interface{}
		joins      []string
	)

	if filter != nil {
		if filter.List != "" {
			conditions = append(conditions, "t.list = ?")
			args = append(args, filter.List)
		}
		if filter.State != nil {
			conditions = append(conditions, "t.state = ?")
			args = append(args, string(*filter.State))
		}
		if filter.Source != "" {
			conditions = append(conditions, "t.source = ?")
			args = append(args, filter.Source)
		}
		if filter.ParentID != nil {
			if *filter.ParentID == "" {
				conditions = append(conditions, "t.parent_id IS NULL")
			} else {
				conditions = append(conditions, "t.parent_id = ?")
				args = append(args, *filter.ParentID)
			}
		}
		if filter.Waiting != nil {
			if *filter.Waiting {
				conditions = append(conditions, "t.waiting_on IS NOT NULL")
			} else {
				conditions = append(conditions, "t.waiting_on IS NULL")
			}
		}
		if filter.Stale > 0 {
			cutoff := time.Now().UTC().Add(-filter.Stale)
			conditions = append(conditions, "t.updated < ?")
			args = append(args, cutoff)
			// Stale only applies to open tasks.
			conditions = append(conditions, "t.state = 'open'")
		}

		// Label filter: tasks must have ALL specified labels.
		for _, label := range filter.Labels {
			joins = append(joins, fmt.Sprintf(
				"JOIN task_labels tl_%d ON tl_%d.task_id = t.id AND tl_%d.label = ?",
				len(joins), len(joins), len(joins),
			))
			args = append(args, label)
		}

		// Unseen filter: tasks this agent hasn't seen, or that were updated
		// after the agent last saw them.
		if filter.UnseenBy != "" {
			joins = append(joins, "LEFT JOIN seen s_unseen ON s_unseen.task_id = t.id AND s_unseen.agent = ?")
			args = append(args, filter.UnseenBy)
			conditions = append(conditions, "(s_unseen.seen_at IS NULL OR s_unseen.seen_at < t.updated)")
			conditions = append(conditions, "t.state = 'open'")
		}

		// Seen-by filter: tasks this agent has seen within a time window.
		if filter.SeenBy != "" {
			joins = append(joins, "JOIN seen s_seen ON s_seen.task_id = t.id AND s_seen.agent = ?")
			args = append(args, filter.SeenBy)
			if filter.Since > 0 {
				cutoff := time.Now().UTC().Add(-filter.Since)
				conditions = append(conditions, "s_seen.seen_at >= ?")
				args = append(args, cutoff)
			}
		}
	}

	query := "SELECT t.id, t.list, t.what, t.source, t.state, t.waiting_on, t.parent_id, t.context, t.created, t.updated, t.resolved_at, t.resolved_reason FROM tasks t"
	for _, j := range joins {
		query += " " + j
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY t.created DESC"

	return query, args
}

// ---------------------------------------------------------------------------
// UpdateTask
// ---------------------------------------------------------------------------

func (s *SQLiteStore) UpdateTask(ctx context.Context, id string, update *model.TaskUpdate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Verify the task exists and is open.
	var currentState string
	err = tx.QueryRowContext(ctx, "SELECT state FROM tasks WHERE id = ?", id).Scan(&currentState)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking task: %w", err)
	}

	// Build a dynamic UPDATE statement from the non-nil fields.
	var setClauses []string
	var setArgs []interface{}

	if update.What != nil {
		setClauses = append(setClauses, "what = ?")
		setArgs = append(setArgs, *update.What)
	}
	if update.List != nil {
		setClauses = append(setClauses, "list = ?")
		setArgs = append(setArgs, *update.List)
	}
	if update.WaitingOn != nil {
		if *update.WaitingOn == "" {
			setClauses = append(setClauses, "waiting_on = NULL")
		} else {
			setClauses = append(setClauses, "waiting_on = ?")
			setArgs = append(setArgs, *update.WaitingOn)
		}
	}
	if update.Context != nil {
		setClauses = append(setClauses, "context = ?")
		setArgs = append(setArgs, *update.Context)
	}
	if update.ParentID != nil {
		if *update.ParentID == "" {
			setClauses = append(setClauses, "parent_id = NULL")
		} else {
			setClauses = append(setClauses, "parent_id = ?")
			setArgs = append(setArgs, *update.ParentID)
		}
	}

	// Always update the timestamp if any field changed.
	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated = ?")
		setArgs = append(setArgs, time.Now().UTC())
		setArgs = append(setArgs, id)

		_, err = tx.ExecContext(ctx,
			"UPDATE tasks SET "+strings.Join(setClauses, ", ")+" WHERE id = ?",
			setArgs...,
		)
		if err != nil {
			return fmt.Errorf("updating task: %w", err)
		}
	}

	// Handle label additions.
	if len(update.AddLabels) > 0 {
		if err := insertLabels(ctx, tx, id, update.AddLabels); err != nil {
			return err
		}
		// Touch the updated timestamp if only labels changed.
		if len(setClauses) == 0 {
			if _, err := tx.ExecContext(ctx, "UPDATE tasks SET updated = ? WHERE id = ?", time.Now().UTC(), id); err != nil {
				return fmt.Errorf("updating timestamp: %w", err)
			}
		}
	}

	// Handle label removals.
	for _, label := range update.DelLabels {
		if _, err := tx.ExecContext(ctx, "DELETE FROM task_labels WHERE task_id = ? AND label = ?", id, label); err != nil {
			return fmt.Errorf("removing label %q: %w", label, err)
		}
	}
	if len(update.DelLabels) > 0 && len(setClauses) == 0 && len(update.AddLabels) == 0 {
		if _, err := tx.ExecContext(ctx, "UPDATE tasks SET updated = ? WHERE id = ?", time.Now().UTC(), id); err != nil {
			return fmt.Errorf("updating timestamp: %w", err)
		}
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// DoneTask / KillTask
// ---------------------------------------------------------------------------

func (s *SQLiteStore) DoneTask(ctx context.Context, id string) error {
	return s.resolveTask(ctx, id, model.StateDone, "")
}

func (s *SQLiteStore) KillTask(ctx context.Context, id, reason string) error {
	if reason == "" {
		return fmt.Errorf("reason is required when killing a task")
	}
	return s.resolveTask(ctx, id, model.StateKilled, reason)
}

func (s *SQLiteStore) resolveTask(ctx context.Context, id string, to model.State, reason string) error {
	var currentState model.State
	err := s.db.QueryRowContext(ctx, "SELECT state FROM tasks WHERE id = ?", id).Scan(&currentState)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking task: %w", err)
	}

	if err := model.Transition(currentState, to); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTransition, err)
	}

	now := time.Now().UTC()
	var resolvedReason *string
	if reason != "" {
		resolvedReason = &reason
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE tasks SET state = ?, resolved_at = ?, resolved_reason = ?, updated = ?
		WHERE id = ?`,
		to, now, resolvedReason, now, id,
	)
	if err != nil {
		return fmt.Errorf("resolving task: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Seen tracking
// ---------------------------------------------------------------------------

func (s *SQLiteStore) MarkSeen(ctx context.Context, agent string, taskIDs []string) error {
	if agent == "" {
		return fmt.Errorf("agent name is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	for _, taskID := range taskIDs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO seen (agent, task_id, seen_at) VALUES (?, ?, ?)
			ON CONFLICT(agent, task_id) DO UPDATE SET seen_at = excluded.seen_at`,
			agent, taskID, now,
		)
		if err != nil {
			return fmt.Errorf("marking task %s as seen: %w", taskID, err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) MarkAllSeen(ctx context.Context, agent, list string) error {
	if agent == "" {
		return fmt.Errorf("agent name is required")
	}
	if list == "" {
		return fmt.Errorf("list is required for MarkAllSeen")
	}

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO seen (agent, task_id, seen_at)
		SELECT ?, id, ? FROM tasks WHERE list = ? AND state = 'open'
		ON CONFLICT(agent, task_id) DO UPDATE SET seen_at = excluded.seen_at`,
		agent, now, list,
	)
	if err != nil {
		return fmt.Errorf("marking all tasks as seen: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Labels, Lists, Stats
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ListLabels(ctx context.Context, list string) ([]model.LabelCount, error) {
	query := `
		SELECT tl.label, COUNT(*) as count
		FROM task_labels tl
		JOIN tasks t ON t.id = tl.task_id
		WHERE t.state = 'open'`
	var args []interface{}

	if list != "" {
		query += " AND t.list = ?"
		args = append(args, list)
	}
	query += " GROUP BY tl.label ORDER BY count DESC, tl.label"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying labels: %w", err)
	}
	defer rows.Close()

	var labels []model.LabelCount
	for rows.Next() {
		var lc model.LabelCount
		if err := rows.Scan(&lc.Label, &lc.Count); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, lc)
	}
	return labels, rows.Err()
}

func (s *SQLiteStore) ClearLabel(ctx context.Context, label, list string) (int, error) {
	var result sql.Result
	var err error

	if list != "" {
		result, err = s.db.ExecContext(ctx, `
			DELETE FROM task_labels WHERE label = ? AND task_id IN (
				SELECT id FROM tasks WHERE list = ?
			)`, label, list)
	} else {
		result, err = s.db.ExecContext(ctx, "DELETE FROM task_labels WHERE label = ?", label)
	}
	if err != nil {
		return 0, fmt.Errorf("clearing label: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

func (s *SQLiteStore) ListLists(ctx context.Context) ([]model.ListInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT list,
		       SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END) as open_count,
		       SUM(CASE WHEN state = 'done' THEN 1 ELSE 0 END) as done_count,
		       SUM(CASE WHEN state = 'killed' THEN 1 ELSE 0 END) as killed_count
		FROM tasks
		GROUP BY list
		ORDER BY list`)
	if err != nil {
		return nil, fmt.Errorf("querying lists: %w", err)
	}
	defer rows.Close()

	var lists []model.ListInfo
	for rows.Next() {
		var li model.ListInfo
		if err := rows.Scan(&li.Name, &li.Open, &li.Done, &li.Killed); err != nil {
			return nil, fmt.Errorf("scanning list: %w", err)
		}
		lists = append(lists, li)
	}
	return lists, rows.Err()
}

func (s *SQLiteStore) Stats(ctx context.Context) (*model.Stats, error) {
	var stats model.Stats
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN state = 'open' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN state = 'done' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN state = 'killed' THEN 1 ELSE 0 END), 0)
		FROM tasks`,
	).Scan(&stats.TotalOpen, &stats.TotalDone, &stats.TotalKilled)
	if err != nil {
		return nil, fmt.Errorf("querying stats: %w", err)
	}
	return &stats, nil
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func (s *SQLiteStore) ExportTasks(ctx context.Context, filter *model.Filter) ([]*model.Task, error) {
	tasks, err := s.ListTasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err := populateDependsOnIDs(ctx, s.db, tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *SQLiteStore) ImportTasks(ctx context.Context, tasks []*model.Task) error {
	orderedTasks, err := sortTasksForImport(tasks)
	if err != nil {
		return fmt.Errorf("sorting imported tasks: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, task := range orderedTasks {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO tasks (id, list, what, source, state, waiting_on, parent_id, context, created, updated, resolved_at, resolved_reason)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			task.ID, task.List, task.What, task.Source, task.State,
			task.WaitingOn, task.ParentID, task.Context,
			task.Created, task.Updated, task.ResolvedAt, task.ResolvedReason,
		)
		if err != nil {
			return fmt.Errorf("importing task %s: %w", task.ID, err)
		}

		if err := insertLabels(ctx, tx, task.ID, task.Labels); err != nil {
			return fmt.Errorf("importing labels for %s: %w", task.ID, err)
		}
	}

	for _, task := range tasks {
		if err := insertTaskDeps(ctx, tx, task.ID, task.DependsOnIDs); err != nil {
			return fmt.Errorf("importing deps for %s: %w", task.ID, err)
		}
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Dependencies
// ---------------------------------------------------------------------------

func (s *SQLiteStore) AddDep(ctx context.Context, taskID, dependsOnID string) error {
	if taskID == dependsOnID {
		return ErrSelfDependency
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Verify both tasks exist and taskID is open.
	var taskState string
	err = tx.QueryRowContext(ctx, "SELECT state FROM tasks WHERE id = ?", taskID).Scan(&taskState)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking task: %w", err)
	}
	if taskState != string(model.StateOpen) {
		return ErrTaskResolved
	}

	var depExists bool
	err = tx.QueryRowContext(ctx, "SELECT 1 FROM tasks WHERE id = ?", dependsOnID).Scan(&depExists)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking dependency: %w", err)
	}

	// Cycle detection: walk transitive deps of dependsOnID; reject if taskID appears.
	var cycleFound bool
	err = tx.QueryRowContext(ctx, `
		WITH RECURSIVE chain(id) AS (
			SELECT depends_on FROM task_deps WHERE task_id = ?
			UNION
			SELECT d.depends_on FROM task_deps d JOIN chain c ON d.task_id = c.id
		)
		SELECT 1 FROM chain WHERE id = ? LIMIT 1`, dependsOnID, taskID).Scan(&cycleFound)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("cycle detection: %w", err)
	}
	if cycleFound {
		return ErrCyclicDependency
	}

	_, err = tx.ExecContext(ctx,
		"INSERT OR IGNORE INTO task_deps (task_id, depends_on) VALUES (?, ?)",
		taskID, dependsOnID)
	if err != nil {
		return fmt.Errorf("inserting dep: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) RemoveDep(ctx context.Context, taskID, dependsOnID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM task_deps WHERE task_id = ? AND depends_on = ?",
		taskID, dependsOnID)
	if err != nil {
		return fmt.Errorf("removing dep: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrDepNotFound
	}
	return nil
}

func (s *SQLiteStore) Deps(ctx context.Context, taskID string) ([]*model.Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.list, t.what, t.source, t.state, t.waiting_on, t.parent_id, t.context,
		       t.created, t.updated, t.resolved_at, t.resolved_reason
		FROM tasks t
		JOIN task_deps d ON d.depends_on = t.id
		WHERE d.task_id = ?
		ORDER BY t.created DESC`, taskID)
	if err != nil {
		return nil, fmt.Errorf("querying deps: %w", err)
	}
	defer rows.Close()
	return scanTasksWithLabels(ctx, s.db, rows)
}

func (s *SQLiteStore) Dependents(ctx context.Context, taskID string) ([]*model.Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.list, t.what, t.source, t.state, t.waiting_on, t.parent_id, t.context,
		       t.created, t.updated, t.resolved_at, t.resolved_reason
		FROM tasks t
		JOIN task_deps d ON d.task_id = t.id
		WHERE d.depends_on = ?
		ORDER BY t.created DESC`, taskID)
	if err != nil {
		return nil, fmt.Errorf("querying dependents: %w", err)
	}
	defer rows.Close()
	return scanTasksWithLabels(ctx, s.db, rows)
}

func (s *SQLiteStore) Frontier(ctx context.Context, list string) ([]*model.Task, error) {
	query := `
			SELECT t.id, t.list, t.what, t.source, t.state, t.waiting_on, t.parent_id, t.context,
			       t.created, t.updated, t.resolved_at, t.resolved_reason
			FROM tasks t
			WHERE t.state = 'open'
			  AND t.waiting_on IS NULL
			  AND NOT EXISTS (
			      SELECT 1 FROM task_deps d
			      JOIN tasks dep ON dep.id = d.depends_on
			      WHERE d.task_id = t.id
			        AND dep.state = 'open'
		  )`
	var args []interface{}
	if list != "" {
		query += " AND t.list = ?"
		args = append(args, list)
	}
	query += " ORDER BY t.created DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying frontier: %w", err)
	}
	defer rows.Close()
	return scanTasksWithLabels(ctx, s.db, rows)
}

func (s *SQLiteStore) DepEdges(ctx context.Context, list string) ([]model.DepEdge, error) {
	query := "SELECT d.task_id, d.depends_on FROM task_deps d"
	var args []interface{}
	if list != "" {
		query += " JOIN tasks t ON t.id = d.task_id WHERE t.list = ?"
		args = append(args, list)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying dep edges: %w", err)
	}
	defer rows.Close()

	var edges []model.DepEdge
	for rows.Next() {
		var e model.DepEdge
		if err := rows.Scan(&e.TaskID, &e.DependsOn); err != nil {
			return nil, fmt.Errorf("scanning dep edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// scanTasksWithLabels reads all tasks from rows and loads their labels.
func scanTasksWithLabels(ctx context.Context, database *sql.DB, rows *sql.Rows) ([]*model.Task, error) {
	var tasks []*model.Task
	for rows.Next() {
		task, err := scanTaskFromRows(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tasks: %w", err)
	}
	for _, task := range tasks {
		labels, err := getLabels(ctx, database, task.ID)
		if err != nil {
			return nil, err
		}
		task.Labels = labels
	}
	return tasks, nil
}

func populateDependsOnIDs(ctx context.Context, database *sql.DB, tasks []*model.Task) error {
	for _, task := range tasks {
		dependsOnIDs, err := getDependsOnIDs(ctx, database, task.ID)
		if err != nil {
			return err
		}
		task.DependsOnIDs = dependsOnIDs
	}
	return nil
}

func getDependsOnIDs(ctx context.Context, database *sql.DB, taskID string) ([]string, error) {
	rows, err := database.QueryContext(ctx,
		"SELECT depends_on FROM task_deps WHERE task_id = ? ORDER BY depends_on", taskID)
	if err != nil {
		return nil, fmt.Errorf("querying dependency ids: %w", err)
	}
	defer rows.Close()

	var dependsOnIDs []string
	for rows.Next() {
		var dependsOnID string
		if err := rows.Scan(&dependsOnID); err != nil {
			return nil, fmt.Errorf("scanning dependency id: %w", err)
		}
		dependsOnIDs = append(dependsOnIDs, dependsOnID)
	}
	return dependsOnIDs, rows.Err()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func sortTasksForImport(tasks []*model.Task) ([]*model.Task, error) {
	taskByID := make(map[string]*model.Task, len(tasks))
	for _, task := range tasks {
		if _, exists := taskByID[task.ID]; exists {
			return nil, fmt.Errorf("duplicate task id %s", task.ID)
		}
		taskByID[task.ID] = task
	}

	ordered := make([]*model.Task, 0, len(tasks))
	visiting := make(map[string]bool, len(tasks))
	visited := make(map[string]bool, len(tasks))

	var visit func(task *model.Task) error
	visit = func(task *model.Task) error {
		if visited[task.ID] {
			return nil
		}
		if visiting[task.ID] {
			return fmt.Errorf("parent cycle detected at %s", task.ID)
		}

		visiting[task.ID] = true
		if task.ParentID != nil && *task.ParentID != "" {
			if parent, ok := taskByID[*task.ParentID]; ok {
				if err := visit(parent); err != nil {
					return err
				}
			}
		}
		visiting[task.ID] = false
		visited[task.ID] = true
		ordered = append(ordered, task)
		return nil
	}

	for _, task := range tasks {
		if err := visit(task); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}

// insertLabels adds label rows for a task within an existing transaction.
// Duplicates are silently ignored (INSERT OR IGNORE).
func insertLabels(ctx context.Context, tx *sql.Tx, taskID string, labels []string) error {
	for _, label := range labels {
		_, err := tx.ExecContext(ctx,
			"INSERT OR IGNORE INTO task_labels (task_id, label) VALUES (?, ?)",
			taskID, label,
		)
		if err != nil {
			return fmt.Errorf("inserting label %q: %w", label, err)
		}
	}
	return nil
}

func insertTaskDeps(ctx context.Context, tx *sql.Tx, taskID string, dependsOnIDs []string) error {
	for _, dependsOnID := range dependsOnIDs {
		_, err := tx.ExecContext(ctx,
			"INSERT OR IGNORE INTO task_deps (task_id, depends_on) VALUES (?, ?)",
			taskID, dependsOnID,
		)
		if err != nil {
			return fmt.Errorf("inserting dependency %q: %w", dependsOnID, err)
		}
	}
	return nil
}

// getLabels retrieves all labels for a task.
func getLabels(ctx context.Context, database *sql.DB, taskID string) ([]string, error) {
	rows, err := database.QueryContext(ctx,
		"SELECT label FROM task_labels WHERE task_id = ? ORDER BY label", taskID)
	if err != nil {
		return nil, fmt.Errorf("querying labels: %w", err)
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, label)
	}
	return labels, rows.Err()
}

// scanTask reads a single task row (without labels).
func scanTask(row *sql.Row) (*model.Task, error) {
	var t model.Task
	err := row.Scan(
		&t.ID, &t.List, &t.What, &t.Source, &t.State,
		&t.WaitingOn, &t.ParentID, &t.Context,
		&t.Created, &t.Updated, &t.ResolvedAt, &t.ResolvedReason,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning task: %w", err)
	}
	return &t, nil
}

// scanTaskFromRows reads a task from a *sql.Rows cursor (without labels).
func scanTaskFromRows(rows *sql.Rows) (*model.Task, error) {
	var t model.Task
	err := rows.Scan(
		&t.ID, &t.List, &t.What, &t.Source, &t.State,
		&t.WaitingOn, &t.ParentID, &t.Context,
		&t.Created, &t.Updated, &t.ResolvedAt, &t.ResolvedReason,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning task: %w", err)
	}
	return &t, nil
}
