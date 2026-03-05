package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/queueprim/internal/api"
	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHandler returns an api.Handler backed by a fresh in-memory store.
func newTestHandler(t *testing.T) (*api.Handler, store.Store) {
	t.Helper()
	rawDB, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(rawDB)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return api.New(s, logger), s
}

// do is a test helper that executes an HTTP request against the handler.
func do(t *testing.T, handler http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(buf.Len())
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// parseJSON decodes the response body into v.
func parseJSON(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	require.NoError(t, json.NewDecoder(rr.Body).Decode(v))
}

// enqueueTestJob is a test helper that enqueues a job via the API and returns its ID.
func enqueueTestJob(t *testing.T, h http.Handler, queue string) string {
	t.Helper()
	rr := do(t, h, "POST", "/v1/jobs", map[string]interface{}{
		"queue":   queue,
		"payload": map[string]string{"test": "data"},
	})
	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	var job model.Job
	parseJSON(t, rr, &job)
	return job.ID
}

// ---------------------------------------------------------------------------
// POST /v1/jobs
// ---------------------------------------------------------------------------

func TestEnqueueJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "POST", "/v1/jobs", map[string]interface{}{
		"queue":       "infra/fixes",
		"type":        "ssh_fail",
		"priority":    "high",
		"payload":     map[string]string{"summary": "test"},
		"max_retries": 2,
	})

	assert.Equal(t, http.StatusCreated, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "infra/fixes", job.Queue)
	assert.Equal(t, "ssh_fail", job.Type)
	assert.Equal(t, model.PriorityHigh, job.Priority)
	assert.Equal(t, model.StatusPending, job.Status)
	assert.Equal(t, 2, job.MaxRetries)
}

func TestEnqueueJob_WithDelay(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "POST", "/v1/jobs", map[string]interface{}{
		"queue":   "q",
		"payload": map[string]string{"k": "v"},
		"delay":   "5m",
	})
	assert.Equal(t, http.StatusCreated, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.True(t, job.VisibleAfter.After(time.Now()), "delayed job should be invisible")
}

func TestEnqueueJob_InvalidPayload(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "POST", "/v1/jobs", map[string]interface{}{
		// missing queue
		"payload": map[string]string{"k": "v"},
	})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ---------------------------------------------------------------------------
// GET /v1/jobs
// ---------------------------------------------------------------------------

func TestListJobs(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	enqueueTestJob(t, handler, "q1")
	enqueueTestJob(t, handler, "q2")
	enqueueTestJob(t, handler, "q1")

	// All jobs.
	rr := do(t, handler, "GET", "/v1/jobs", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	var all []model.Job
	parseJSON(t, rr, &all)
	assert.Len(t, all, 3)

	// Filter by queue.
	rr = do(t, handler, "GET", "/v1/jobs?queue=q1", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	var q1Jobs []model.Job
	parseJSON(t, rr, &q1Jobs)
	assert.Len(t, q1Jobs, 2)
}

// ---------------------------------------------------------------------------
// GET /v1/jobs/{id}
// ---------------------------------------------------------------------------

func TestGetJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "GET", "/v1/jobs/"+id, nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	var job model.Job
	parseJSON(t, rr, &job)
	assert.Equal(t, id, job.ID)
}

func TestGetJob_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "GET", "/v1/jobs/q_notexist", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---------------------------------------------------------------------------
// POST /v1/queues/{queue}/dequeue
// ---------------------------------------------------------------------------

func TestDequeueJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	enqueueTestJob(t, handler, "infra/fixes")

	rr := do(t, handler, "POST", "/v1/queues/infra/fixes/dequeue", map[string]interface{}{
		"worker":  "raphael",
		"timeout": "30m",
	})
	assert.Equal(t, http.StatusOK, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.Equal(t, model.StatusClaimed, job.Status)
	assert.Equal(t, "raphael", *job.ClaimedBy)
}

func TestDequeueJob_EmptyQueue(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "POST", "/v1/queues/empty/queue/dequeue", nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/complete
// ---------------------------------------------------------------------------

func TestCompleteJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	// Dequeue first.
	rr := do(t, handler, "POST", "/v1/queues/q/dequeue", map[string]interface{}{"worker": "w"})
	require.Equal(t, http.StatusOK, rr.Code)

	rr = do(t, handler, "POST", "/v1/jobs/"+id+"/complete", map[string]interface{}{
		"output": map[string]string{"result": "ok"},
	})
	assert.Equal(t, http.StatusOK, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.Equal(t, model.StatusDone, job.Status)
	assert.NotNil(t, job.CompletedAt)
}

func TestCompleteJob_NotClaimed(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "POST", "/v1/jobs/"+id+"/complete", nil)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/fail
// ---------------------------------------------------------------------------

func TestFailJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "POST", "/v1/queues/q/dequeue", map[string]interface{}{"worker": "w"})
	require.Equal(t, http.StatusOK, rr.Code)

	rr = do(t, handler, "POST", "/v1/jobs/"+id+"/fail", map[string]interface{}{
		"reason": "something broke",
		"dead":   true,
	})
	assert.Equal(t, http.StatusOK, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.Equal(t, model.StatusDead, job.Status)
	assert.Equal(t, "something broke", *job.FailureReason)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/release
// ---------------------------------------------------------------------------

func TestReleaseJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "POST", "/v1/queues/q/dequeue", map[string]interface{}{"worker": "w"})
	require.Equal(t, http.StatusOK, rr.Code)

	rr = do(t, handler, "POST", "/v1/jobs/"+id+"/release", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var job model.Job
	parseJSON(t, rr, &job)
	assert.Equal(t, model.StatusPending, job.Status)
	assert.Nil(t, job.ClaimedBy)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/extend
// ---------------------------------------------------------------------------

func TestExtendJob(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	id := enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "POST", "/v1/queues/q/dequeue", map[string]interface{}{"worker": "w"})
	require.Equal(t, http.StatusOK, rr.Code)
	var claimed model.Job
	parseJSON(t, rr, &claimed)

	rr = do(t, handler, "POST", "/v1/jobs/"+id+"/extend", map[string]interface{}{"by": "45m"})
	assert.Equal(t, http.StatusOK, rr.Code)

	var extended model.Job
	parseJSON(t, rr, &extended)
	assert.True(t, extended.VisibleAfter.After(claimed.VisibleAfter))
}

// ---------------------------------------------------------------------------
// GET /v1/queues
// ---------------------------------------------------------------------------

func TestListQueues(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	enqueueTestJob(t, handler, "infra/fixes")
	enqueueTestJob(t, handler, "infra/fixes")
	enqueueTestJob(t, handler, "review/code")

	rr := do(t, handler, "GET", "/v1/queues", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var queues []model.QueueInfo
	parseJSON(t, rr, &queues)
	assert.Len(t, queues, 2)
}

// ---------------------------------------------------------------------------
// GET /v1/stats
// ---------------------------------------------------------------------------

func TestGetStats(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	enqueueTestJob(t, handler, "q")
	enqueueTestJob(t, handler, "q")

	rr := do(t, handler, "GET", "/v1/stats", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats model.Stats
	parseJSON(t, rr, &stats)
	assert.Equal(t, 2, stats.TotalPending)
}

// ---------------------------------------------------------------------------
// DELETE /v1/queues/{queue}
// ---------------------------------------------------------------------------

func TestPurgeQueue(t *testing.T) {
	h, s := newTestHandler(t)
	handler := h.Router()
	ctx := context.Background()

	id := enqueueTestJob(t, handler, "q")

	// Dequeue and complete it.
	rr := do(t, handler, "POST", "/v1/queues/q/dequeue", nil)
	require.Equal(t, http.StatusOK, rr.Code)
	_ = s.CompleteJob(ctx, id, nil)

	rr = do(t, handler, "DELETE", "/v1/queues/q?status=done", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]int
	parseJSON(t, rr, &result)
	assert.Equal(t, 1, result["deleted"])
}

// Verify ErrEmpty is handled as 204 in handler, not a server error.
func TestDequeueReturns204NotError(t *testing.T) {
	h, _ := newTestHandler(t)
	handler := h.Router()

	rr := do(t, handler, "POST", "/v1/queues/missing/queue/dequeue", nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())
}

// Make sure errors.Is works correctly with store sentinel errors.
func TestStoreErrIs(t *testing.T) {
	assert.True(t, errors.Is(store.ErrNotFound, store.ErrNotFound))
	assert.True(t, errors.Is(store.ErrEmpty, store.ErrEmpty))
	assert.True(t, errors.Is(store.ErrInvalidTransition, store.ErrInvalidTransition))
}
