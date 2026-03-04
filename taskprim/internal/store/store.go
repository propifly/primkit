// Package store defines the persistence contract for taskprim and provides
// a SQLite implementation. Every consumer (CLI, API, MCP) depends on the
// Store interface, which keeps the persistence layer swappable and testable.
package store

import (
	"context"
	"errors"

	"github.com/propifly/primkit/taskprim/internal/model"
)

// ErrNotFound is returned when a requested task does not exist.
var ErrNotFound = errors.New("task not found")

// ErrInvalidTransition is returned when a state transition is not allowed.
var ErrInvalidTransition = errors.New("invalid state transition")

// Store is the persistence contract for taskprim. All methods accept a context
// for cancellation and timeout propagation.
type Store interface {
	// CreateTask persists a new task. The task's ID, State, Created, and Updated
	// fields are set by the store. The caller must set List, What, and Source.
	CreateTask(ctx context.Context, task *model.Task) error

	// GetTask retrieves a single task by ID, including its labels.
	// Returns ErrNotFound if the task doesn't exist.
	GetTask(ctx context.Context, id string) (*model.Task, error)

	// ListTasks returns all tasks matching the filter. An empty filter returns
	// all tasks. Results are ordered by created timestamp (newest first).
	ListTasks(ctx context.Context, filter *model.Filter) ([]*model.Task, error)

	// UpdateTask applies partial updates to a task's mutable fields.
	// Returns ErrNotFound if the task doesn't exist.
	UpdateTask(ctx context.Context, id string, update *model.TaskUpdate) error

	// DoneTask marks a task as completed. Sets state to "done" and resolved_at
	// to the current time. Returns ErrNotFound or ErrInvalidTransition.
	DoneTask(ctx context.Context, id string) error

	// KillTask marks a task as killed with a required reason explaining why it
	// was dropped. Returns ErrNotFound or ErrInvalidTransition.
	KillTask(ctx context.Context, id string, reason string) error

	// MarkSeen records that an agent has seen specific tasks. Uses upsert
	// semantics — calling MarkSeen twice for the same (agent, task) updates
	// the seen_at timestamp.
	MarkSeen(ctx context.Context, agent string, taskIDs []string) error

	// MarkAllSeen marks all open tasks in a list as seen by the agent.
	MarkAllSeen(ctx context.Context, agent string, list string) error

	// ListLabels returns all labels with the count of open tasks carrying each.
	// If list is non-empty, only labels on tasks in that list are counted.
	ListLabels(ctx context.Context, list string) ([]model.LabelCount, error)

	// ClearLabel removes a label from all tasks. If list is non-empty, only
	// tasks in that list are affected. Returns the number of tasks modified.
	ClearLabel(ctx context.Context, label, list string) (int, error)

	// ListLists returns all lists with their task counts by state.
	ListLists(ctx context.Context) ([]model.ListInfo, error)

	// Stats returns aggregate counts across all tasks.
	Stats(ctx context.Context) (*model.Stats, error)

	// ExportTasks returns tasks matching the filter in a format suitable for
	// JSON export. Same as ListTasks but guarantees all fields are populated.
	ExportTasks(ctx context.Context, filter *model.Filter) ([]*model.Task, error)

	// ImportTasks bulk-inserts tasks, preserving their original IDs. Used to
	// restore from an export. Runs in a single transaction.
	ImportTasks(ctx context.Context, tasks []*model.Task) error

	// Close releases the database connection.
	Close() error
}
