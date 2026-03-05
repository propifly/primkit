// Package store defines the persistence contract for queueprim and provides
// a SQLite implementation. Every consumer (CLI, API, MCP) depends on the
// Store interface, keeping the persistence layer swappable and testable.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/propifly/primkit/queueprim/internal/model"
)

// ErrNotFound is returned when a requested job does not exist.
var ErrNotFound = errors.New("job not found")

// ErrEmpty is returned by DequeueJob when the queue has no available jobs.
// This is not an error condition — callers use it to detect an empty queue
// and exit 0 or loop when waiting. Use errors.Is(err, ErrEmpty) to check.
var ErrEmpty = errors.New("queue is empty")

// ErrInvalidTransition is returned when an operation is not valid for the
// job's current status (e.g., completing an already-done job).
var ErrInvalidTransition = errors.New("invalid status transition")

// ErrValidation is returned when job fields fail validation (e.g., missing queue).
var ErrValidation = errors.New("validation error")

// Store is the persistence contract for queueprim. All methods accept a context
// for cancellation and timeout propagation.
type Store interface {
	// EnqueueJob persists a new job. The job's ID, Status, AttemptCount,
	// CreatedAt, and UpdatedAt fields are set by the store. The caller must
	// set Queue, Payload, and optionally Type, Priority, MaxRetries, and
	// VisibleAfter (for delayed jobs).
	EnqueueJob(ctx context.Context, job *model.Job) error

	// DequeueJob atomically claims the next available job in the given queue.
	// "Available" means status=pending AND visible_after <= now. If type is
	// non-empty, only jobs of that type are considered. The claimed job's
	// status is set to "claimed", claimed_by is set to worker, and
	// visible_after is extended by timeout (default 30m if zero).
	//
	// Returns ErrEmpty if no jobs are available.
	DequeueJob(ctx context.Context, queue, worker, jobType string, timeout time.Duration) (*model.Job, error)

	// CompleteJob marks a claimed job as done. If output is non-nil, it is
	// stored as the job's result payload. Returns ErrNotFound or ErrInvalidTransition.
	CompleteJob(ctx context.Context, id string, output []byte) error

	// FailJob marks a claimed job as failed. If forceDeadLetter is true, the
	// job is moved to "dead" regardless of remaining retries. Otherwise, if
	// attempt_count <= max_retries, the job returns to "pending"; if retries
	// are exhausted, it moves to "dead". Returns ErrNotFound or ErrInvalidTransition.
	FailJob(ctx context.Context, id, reason string, forceDeadLetter bool) error

	// ReleaseJob returns a claimed job to "pending" immediately, clearing the
	// claim fields. Returns ErrNotFound or ErrInvalidTransition.
	ReleaseJob(ctx context.Context, id string) error

	// ExtendJob extends the visibility timeout of a claimed job by the given
	// duration. Used by workers on long-running jobs to prevent auto-release.
	// Returns ErrNotFound or ErrInvalidTransition.
	ExtendJob(ctx context.Context, id string, by time.Duration) error

	// PeekJob returns the next job that would be returned by DequeueJob,
	// without claiming it. Returns ErrEmpty if no jobs are available.
	PeekJob(ctx context.Context, queue string) (*model.Job, error)

	// GetJob retrieves a single job by ID. Returns ErrNotFound if it doesn't exist.
	GetJob(ctx context.Context, id string) (*model.Job, error)

	// ListJobs returns jobs matching the filter. An empty filter returns all
	// jobs. Results are ordered by created_at descending.
	ListJobs(ctx context.Context, filter *model.Filter) ([]*model.Job, error)

	// ListQueues returns all named queues with their job counts by status.
	ListQueues(ctx context.Context) ([]model.QueueInfo, error)

	// Stats returns aggregate counts across all jobs in all queues.
	Stats(ctx context.Context) (*model.Stats, error)

	// PurgeJobs deletes jobs matching the given queue, status, and age criteria.
	// If olderThan is zero, all jobs matching queue+status are deleted.
	// Returns the number of deleted jobs.
	PurgeJobs(ctx context.Context, queue string, status model.Status, olderThan time.Duration) (int, error)

	// ExportJobs returns all jobs in a queue as a slice, suitable for JSON export.
	// If queue is empty, all jobs are exported.
	ExportJobs(ctx context.Context, queue string) ([]*model.Job, error)

	// ImportJobs bulk-inserts jobs, preserving their original IDs. Used to
	// restore from an export. Runs in a single transaction.
	ImportJobs(ctx context.Context, jobs []*model.Job) error

	// SweepExpiredClaims releases claimed jobs whose visible_after timestamp
	// has passed. Returns the number of jobs released.
	SweepExpiredClaims(ctx context.Context) (int, error)

	// Close releases the database connection.
	Close() error
}
