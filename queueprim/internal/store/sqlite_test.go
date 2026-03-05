package store_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore opens an in-memory SQLite database and returns a store ready
// for testing. Each call gets its own isolated database.
func newTestStore(t *testing.T) store.Store {
	t.Helper()
	rawDB, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(rawDB)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

// newJob is a helper to build a valid Job for testing.
func newJob(queue string, payload interface{}) *model.Job {
	raw, _ := json.Marshal(payload)
	return &model.Job{
		Queue:   queue,
		Payload: raw,
	}
}

// ---------------------------------------------------------------------------
// EnqueueJob
// ---------------------------------------------------------------------------

func TestEnqueueJob_Basic(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("test/queue", map[string]string{"key": "val"})
	require.NoError(t, s.EnqueueJob(ctx, job))

	assert.NotEmpty(t, job.ID)
	assert.True(t, len(job.ID) > 2 && job.ID[:2] == "q_", "ID should have q_ prefix")
	assert.Equal(t, model.StatusPending, job.Status)
	assert.Equal(t, model.PriorityNormal, job.Priority)
	assert.Equal(t, 0, job.AttemptCount)
	assert.False(t, job.CreatedAt.IsZero())
}

func TestEnqueueJob_WithPriorityAndType(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := &model.Job{
		Queue:      "infra/fixes",
		Type:       "ssh_auth_fail",
		Priority:   model.PriorityHigh,
		Payload:    json.RawMessage(`{"summary":"test"}`),
		MaxRetries: 3,
	}
	require.NoError(t, s.EnqueueJob(ctx, job))

	fetched, err := s.GetJob(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, "ssh_auth_fail", fetched.Type)
	assert.Equal(t, model.PriorityHigh, fetched.Priority)
	assert.Equal(t, 3, fetched.MaxRetries)
}

func TestEnqueueJob_WithDelay(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	future := time.Now().UTC().Add(10 * time.Minute)
	job := &model.Job{
		Queue:        "test",
		Payload:      json.RawMessage(`{}`),
		VisibleAfter: future,
	}
	require.NoError(t, s.EnqueueJob(ctx, job))

	// Job should NOT be dequeued because it's in the future.
	_, err := s.DequeueJob(ctx, "test", "worker", "", 0)
	assert.ErrorIs(t, err, store.ErrEmpty, "delayed job should not be dequeued yet")
}

func TestEnqueueJob_Validation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		job     *model.Job
		wantErr string
	}{
		{
			name:    "missing queue",
			job:     &model.Job{Payload: json.RawMessage(`{}`)},
			wantErr: "queue is required",
		},
		{
			name:    "missing payload",
			job:     &model.Job{Queue: "q"},
			wantErr: "payload is required",
		},
		{
			name:    "invalid payload json",
			job:     &model.Job{Queue: "q", Payload: json.RawMessage(`not json`)},
			wantErr: "payload must be valid JSON",
		},
		{
			name:    "invalid priority",
			job:     &model.Job{Queue: "q", Payload: json.RawMessage(`{}`), Priority: "urgent"},
			wantErr: "priority must be high, normal, or low",
		},
		{
			name:    "negative max_retries",
			job:     &model.Job{Queue: "q", Payload: json.RawMessage(`{}`), MaxRetries: -1},
			wantErr: "max_retries must be >= 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := s.EnqueueJob(ctx, tc.job)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// ---------------------------------------------------------------------------
// DequeueJob
// ---------------------------------------------------------------------------

func TestDequeueJob_Basic(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("test/queue", map[string]string{"work": "unit1"})
	require.NoError(t, s.EnqueueJob(ctx, job))

	claimed, err := s.DequeueJob(ctx, "test/queue", "worker-1", "", 0)
	require.NoError(t, err)
	require.NotNil(t, claimed)

	assert.Equal(t, job.ID, claimed.ID)
	assert.Equal(t, model.StatusClaimed, claimed.Status)
	assert.NotNil(t, claimed.ClaimedBy)
	assert.Equal(t, "worker-1", *claimed.ClaimedBy)
	assert.NotNil(t, claimed.ClaimedAt)
	assert.Equal(t, 1, claimed.AttemptCount)
	assert.True(t, claimed.VisibleAfter.After(time.Now()), "visible_after should be in the future")
}

func TestDequeueJob_EmptyQueue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.DequeueJob(ctx, "empty/queue", "worker", "", 0)
	assert.ErrorIs(t, err, store.ErrEmpty)
}

func TestDequeueJob_AtomicClaim(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Enqueue one job.
	job := newJob("test/queue", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	// Two workers dequeue concurrently.
	type result struct {
		job *model.Job
		err error
	}
	ch := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func(workerID int) {
			j, err := s.DequeueJob(ctx, "test/queue", "worker", "", 0)
			ch <- result{j, err}
		}(i)
	}

	r1, r2 := <-ch, <-ch

	// Exactly one worker gets the job, the other gets ErrEmpty.
	got := 0
	empty := 0
	for _, r := range []result{r1, r2} {
		if r.err == nil {
			got++
		} else if r.err == store.ErrEmpty {
			empty++
		}
	}
	assert.Equal(t, 1, got, "exactly one worker should claim the job")
	assert.Equal(t, 1, empty, "the other worker should see an empty queue")
}

func TestDequeueJob_PriorityOrder(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	low := &model.Job{Queue: "q", Priority: model.PriorityLow, Payload: json.RawMessage(`{"p":"low"}`)}
	normal := &model.Job{Queue: "q", Priority: model.PriorityNormal, Payload: json.RawMessage(`{"p":"normal"}`)}
	high := &model.Job{Queue: "q", Priority: model.PriorityHigh, Payload: json.RawMessage(`{"p":"high"}`)}

	require.NoError(t, s.EnqueueJob(ctx, low))
	require.NoError(t, s.EnqueueJob(ctx, normal))
	require.NoError(t, s.EnqueueJob(ctx, high))

	// Should get high first.
	j1, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, model.PriorityHigh, j1.Priority)

	// Then normal.
	j2, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, model.PriorityNormal, j2.Priority)

	// Then low.
	j3, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, model.PriorityLow, j3.Priority)
}

func TestDequeueJob_FIFOWithinPriority(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := &model.Job{Queue: "q", Payload: json.RawMessage(`{"seq":1}`)}
	j2 := &model.Job{Queue: "q", Payload: json.RawMessage(`{"seq":2}`)}
	j3 := &model.Job{Queue: "q", Payload: json.RawMessage(`{"seq":3}`)}

	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))
	require.NoError(t, s.EnqueueJob(ctx, j3))

	d1, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	d2, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	d3, _ := s.DequeueJob(ctx, "q", "w", "", 0)

	assert.Equal(t, j1.ID, d1.ID)
	assert.Equal(t, j2.ID, d2.ID)
	assert.Equal(t, j3.ID, d3.ID)
}

func TestDequeueJob_TypeFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	a := &model.Job{Queue: "q", Type: "typeA", Payload: json.RawMessage(`{}`)}
	b := &model.Job{Queue: "q", Type: "typeB", Payload: json.RawMessage(`{}`)}
	require.NoError(t, s.EnqueueJob(ctx, a))
	require.NoError(t, s.EnqueueJob(ctx, b))

	claimed, err := s.DequeueJob(ctx, "q", "w", "typeB", 0)
	require.NoError(t, err)
	assert.Equal(t, "typeB", claimed.Type)
}

// ---------------------------------------------------------------------------
// CompleteJob
// ---------------------------------------------------------------------------

func TestCompleteJob(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)

	output := json.RawMessage(`{"result":"ok"}`)
	require.NoError(t, s.CompleteJob(ctx, claimed.ID, output))

	done, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusDone, done.Status)
	assert.NotNil(t, done.CompletedAt)
	assert.JSONEq(t, `{"result":"ok"}`, string(done.Output))
}

func TestCompleteJob_NotClaimed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	err := s.CompleteJob(ctx, job.ID, nil)
	assert.ErrorIs(t, err, store.ErrInvalidTransition)
}

func TestCompleteJob_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.CompleteJob(ctx, "q_nonexistent", nil)
	assert.ErrorIs(t, err, store.ErrNotFound)
}

// ---------------------------------------------------------------------------
// FailJob
// ---------------------------------------------------------------------------

func TestFailJob_ToDeadWhenNoRetries(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := &model.Job{Queue: "q", Payload: json.RawMessage(`{}`), MaxRetries: 0}
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)

	require.NoError(t, s.FailJob(ctx, claimed.ID, "timed out", false))

	dead, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	// attempt_count=1, max_retries=0 → dead
	assert.Equal(t, model.StatusDead, dead.Status)
	assert.Equal(t, "timed out", *dead.FailureReason)
}

func TestFailJob_RetryWhenRetriesRemain(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := &model.Job{Queue: "q", Payload: json.RawMessage(`{}`), MaxRetries: 2}
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, 1, claimed.AttemptCount)

	require.NoError(t, s.FailJob(ctx, claimed.ID, "first failure", false))

	retried, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	// attempt_count=1, max_retries=2 → pending
	assert.Equal(t, model.StatusPending, retried.Status)

	// Claim and fail again.
	claimed2, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, 2, claimed2.AttemptCount)
	require.NoError(t, s.FailJob(ctx, claimed2.ID, "second failure", false))

	retried2, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, retried2.Status)

	// Claim and fail a third time — now exhausted.
	claimed3, err := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, err)
	assert.Equal(t, 3, claimed3.AttemptCount)
	require.NoError(t, s.FailJob(ctx, claimed3.ID, "third failure", false))

	dead, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusDead, dead.Status)
}

func TestFailJob_ForceDeadLetter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := &model.Job{Queue: "q", Payload: json.RawMessage(`{}`), MaxRetries: 10}
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, _ := s.DequeueJob(ctx, "q", "w", "", 0)

	// Force to dead even though retries remain.
	require.NoError(t, s.FailJob(ctx, claimed.ID, "escalate", true))

	dead, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusDead, dead.Status)
}

// ---------------------------------------------------------------------------
// ReleaseJob
// ---------------------------------------------------------------------------

func TestReleaseJob(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, _ := s.DequeueJob(ctx, "q", "w", "", 0)

	require.NoError(t, s.ReleaseJob(ctx, claimed.ID))

	released, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, released.Status)
	assert.Nil(t, released.ClaimedBy)
	assert.Nil(t, released.ClaimedAt)

	// Should be dequeued again.
	reclaimed, err := s.DequeueJob(ctx, "q", "w2", "", 0)
	require.NoError(t, err)
	assert.Equal(t, claimed.ID, reclaimed.ID)
}

func TestReleaseJob_NotClaimed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	err := s.ReleaseJob(ctx, job.ID)
	assert.ErrorIs(t, err, store.ErrInvalidTransition)
}

// ---------------------------------------------------------------------------
// ExtendJob
// ---------------------------------------------------------------------------

func TestExtendJob(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))
	claimed, _ := s.DequeueJob(ctx, "q", "w", "", 30*time.Minute)
	originalVisible := claimed.VisibleAfter

	require.NoError(t, s.ExtendJob(ctx, claimed.ID, 45*time.Minute))

	extended, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.True(t, extended.VisibleAfter.After(originalVisible),
		"visible_after should be later after extension")
}

func TestExtendJob_NotClaimed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	err := s.ExtendJob(ctx, job.ID, 30*time.Minute)
	assert.ErrorIs(t, err, store.ErrInvalidTransition)
}

// ---------------------------------------------------------------------------
// PeekJob
// ---------------------------------------------------------------------------

func TestPeekJob(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	peeked, err := s.PeekJob(ctx, "q")
	require.NoError(t, err)
	assert.Equal(t, job.ID, peeked.ID)
	assert.Equal(t, model.StatusPending, peeked.Status, "peek should not change status")

	// Peeking again returns the same job.
	peeked2, err := s.PeekJob(ctx, "q")
	require.NoError(t, err)
	assert.Equal(t, job.ID, peeked2.ID)
}

func TestPeekJob_EmptyQueue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.PeekJob(ctx, "empty")
	assert.ErrorIs(t, err, store.ErrEmpty)
}

// ---------------------------------------------------------------------------
// GetJob
// ---------------------------------------------------------------------------

func TestGetJob_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.GetJob(ctx, "q_doesnotexist")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

// ---------------------------------------------------------------------------
// ListJobs
// ---------------------------------------------------------------------------

func TestListJobs_Filter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := &model.Job{Queue: "q1", Type: "typeA", Payload: json.RawMessage(`{}`)}
	j2 := &model.Job{Queue: "q2", Type: "typeB", Payload: json.RawMessage(`{}`)}
	j3 := &model.Job{Queue: "q1", Type: "typeB", Payload: json.RawMessage(`{}`)}
	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))
	require.NoError(t, s.EnqueueJob(ctx, j3))

	// Filter by queue.
	all, err := s.ListJobs(ctx, &model.Filter{Queue: "q1"})
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Filter by type.
	byType, err := s.ListJobs(ctx, &model.Filter{Type: "typeB"})
	require.NoError(t, err)
	assert.Len(t, byType, 2)

	// Filter by queue + type.
	combo, err := s.ListJobs(ctx, &model.Filter{Queue: "q1", Type: "typeB"})
	require.NoError(t, err)
	assert.Len(t, combo, 1)
	assert.Equal(t, j3.ID, combo[0].ID)
}

func TestListJobs_ByStatus(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := newJob("q", `{}`)
	j2 := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))

	claimed, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, s.CompleteJob(ctx, claimed.ID, nil))

	statusDone := model.StatusDone
	done, err := s.ListJobs(ctx, &model.Filter{Status: &statusDone})
	require.NoError(t, err)
	assert.Len(t, done, 1)
	assert.Equal(t, model.StatusDone, done[0].Status)
}

// ---------------------------------------------------------------------------
// ListQueues
// ---------------------------------------------------------------------------

func TestListQueues(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := &model.Job{Queue: "infra/fixes", Payload: json.RawMessage(`{}`)}
	j2 := &model.Job{Queue: "infra/fixes", Payload: json.RawMessage(`{}`)}
	j3 := &model.Job{Queue: "review/code", Payload: json.RawMessage(`{}`)}
	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))
	require.NoError(t, s.EnqueueJob(ctx, j3))

	// Complete one job in infra/fixes.
	claimed, _ := s.DequeueJob(ctx, "infra/fixes", "w", "", 0)
	require.NoError(t, s.CompleteJob(ctx, claimed.ID, nil))

	queues, err := s.ListQueues(ctx)
	require.NoError(t, err)
	require.Len(t, queues, 2)

	infraQueue := queues[0]
	assert.Equal(t, "infra/fixes", infraQueue.Queue)
	assert.Equal(t, 1, infraQueue.Pending)
	assert.Equal(t, 1, infraQueue.Done)

	reviewQueue := queues[1]
	assert.Equal(t, "review/code", reviewQueue.Queue)
	assert.Equal(t, 1, reviewQueue.Pending)
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func TestStats(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := newJob("q", `{}`)
	j2 := newJob("q", `{}`)
	j3 := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))
	require.NoError(t, s.EnqueueJob(ctx, j3))

	// Claim and complete j1.
	claimed1, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, s.CompleteJob(ctx, claimed1.ID, nil))

	// Claim and fail j2 (one-shot → dead).
	claimed2, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, s.FailJob(ctx, claimed2.ID, "error", false))

	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalPending)
	assert.Equal(t, 0, stats.TotalClaimed)
	assert.Equal(t, 1, stats.TotalDone)
	assert.Equal(t, 0, stats.TotalFailed)
	assert.Equal(t, 1, stats.TotalDead)
}

// ---------------------------------------------------------------------------
// PurgeJobs
// ---------------------------------------------------------------------------

func TestPurgeJobs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	j1 := newJob("q", `{}`)
	j2 := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, j1))
	require.NoError(t, s.EnqueueJob(ctx, j2))

	claimed1, _ := s.DequeueJob(ctx, "q", "w", "", 0)
	require.NoError(t, s.CompleteJob(ctx, claimed1.ID, nil))

	// Purge done jobs.
	n, err := s.PurgeJobs(ctx, "q", model.StatusDone, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	// j2 should still exist.
	_, err = s.GetJob(ctx, j2.ID)
	require.NoError(t, err)

	// j1 should be gone.
	_, err = s.GetJob(ctx, j1.ID)
	assert.ErrorIs(t, err, store.ErrNotFound)
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func TestExportImport(t *testing.T) {
	s1 := newTestStore(t)
	s2 := newTestStore(t)
	ctx := context.Background()

	// Create two jobs in s1.
	j1 := &model.Job{Queue: "q", Type: "typeA", Priority: model.PriorityHigh, Payload: json.RawMessage(`{"k":"v1"}`)}
	j2 := &model.Job{Queue: "q", Priority: model.PriorityNormal, Payload: json.RawMessage(`{"k":"v2"}`)}
	require.NoError(t, s1.EnqueueJob(ctx, j1))
	require.NoError(t, s1.EnqueueJob(ctx, j2))

	// Export from s1.
	exported, err := s1.ExportJobs(ctx, "q")
	require.NoError(t, err)
	require.Len(t, exported, 2)

	// Import into s2.
	require.NoError(t, s2.ImportJobs(ctx, exported))

	// Verify s2 has the same jobs.
	jobs, err := s2.ListJobs(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, jobs, 2)
}

// ---------------------------------------------------------------------------
// SweepExpiredClaims
// ---------------------------------------------------------------------------

func TestSweepExpiredClaims(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	job := newJob("q", `{}`)
	require.NoError(t, s.EnqueueJob(ctx, job))

	// Dequeue with a very short timeout, then manually expire it.
	claimed, err := s.DequeueJob(ctx, "q", "w", "", 1*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, model.StatusClaimed, claimed.Status)

	// Wait for the claim to expire (timeout was 1ms).
	time.Sleep(5 * time.Millisecond)

	n, err := s.SweepExpiredClaims(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "one expired claim should be swept")

	swept, err := s.GetJob(ctx, claimed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, swept.Status)
	assert.Nil(t, swept.ClaimedBy)

	// Job should be dequeued again.
	reclaimed, err := s.DequeueJob(ctx, "q", "w2", "", 0)
	require.NoError(t, err)
	assert.Equal(t, claimed.ID, reclaimed.ID)
}
