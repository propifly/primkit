package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/propifly/primkit/taskprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------

// newTestHandler creates an API handler backed by an in-memory SQLite store.
func newTestHandler(t *testing.T) (*Handler, store.Store) {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(s, logger)
	return h, s
}

// doRequest sends an HTTP request against the API handler and returns the
// response recorder. Body is optional — pass nil for GET requests.
func doRequest(t *testing.T, h *Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	h.Router().ServeHTTP(rr, req)
	return rr
}

// decodeBody unmarshals the response body into the target.
func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), target))
}

// seedTask adds a task directly to the store.
func seedTask(t *testing.T, s store.Store, what, list, source string, labels ...string) *model.Task {
	t.Helper()
	task := &model.Task{
		What:   what,
		List:   list,
		Source: source,
		Labels: labels,
	}
	require.NoError(t, s.CreateTask(context.Background(), task))
	return task
}

// --------------------------------------------------------------------
// POST /v1/tasks — create task
// --------------------------------------------------------------------

func TestCreateTask(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]interface{}{
		"list": "ops",
		"what": "Deploy v2",
	}
	rr := doRequest(t, h, "POST", "/v1/tasks", body)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var task model.Task
	decodeBody(t, rr, &task)
	assert.Equal(t, "Deploy v2", task.What)
	assert.Equal(t, "ops", task.List)
	assert.Equal(t, model.StateOpen, task.State)
	assert.NotEmpty(t, task.ID)
}

func TestCreateTask_WithLabels(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]interface{}{
		"list":   "work",
		"what":   "Labeled task",
		"labels": []string{"urgent", "frontend"},
	}
	rr := doRequest(t, h, "POST", "/v1/tasks", body)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var task model.Task
	decodeBody(t, rr, &task)
	assert.ElementsMatch(t, []string{"urgent", "frontend"}, task.Labels)
}

func TestCreateTask_InvalidJSON(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/v1/tasks", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()
	h.Router().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --------------------------------------------------------------------
// GET /v1/tasks — list tasks
// --------------------------------------------------------------------

func TestListTasks(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Task A", "work", "cli")
	seedTask(t, s, "Task B", "personal", "cli")

	rr := doRequest(t, h, "GET", "/v1/tasks", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tasks []*model.Task
	decodeBody(t, rr, &tasks)
	assert.Len(t, tasks, 2)
}

func TestListTasks_FilterByList(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Work task", "work", "cli")
	seedTask(t, s, "Personal task", "personal", "cli")

	rr := doRequest(t, h, "GET", "/v1/tasks?list=work", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tasks []*model.Task
	decodeBody(t, rr, &tasks)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Work task", tasks[0].What)
}

func TestListTasks_FilterByState(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Done task", "work", "cli")
	seedTask(t, s, "Open task", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	rr := doRequest(t, h, "GET", "/v1/tasks?state=done", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tasks []*model.Task
	decodeBody(t, rr, &tasks)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Done task", tasks[0].What)
}

func TestListTasks_FilterByLabel(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Urgent", "work", "cli", "urgent")
	seedTask(t, s, "Normal", "work", "cli")

	rr := doRequest(t, h, "GET", "/v1/tasks?label=urgent", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tasks []*model.Task
	decodeBody(t, rr, &tasks)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Urgent", tasks[0].What)
}

func TestListTasks_Empty(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := doRequest(t, h, "GET", "/v1/tasks", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var tasks []*model.Task
	decodeBody(t, rr, &tasks)
	assert.Empty(t, tasks)
}

// --------------------------------------------------------------------
// GET /v1/tasks/:id — get task
// --------------------------------------------------------------------

func TestGetTask(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Get me", "work", "cli", "tag")

	rr := doRequest(t, h, "GET", "/v1/tasks/"+task.ID, nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var got model.Task
	decodeBody(t, rr, &got)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, "Get me", got.What)
	assert.Contains(t, got.Labels, "tag")
}

func TestGetTask_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := doRequest(t, h, "GET", "/v1/tasks/t_nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --------------------------------------------------------------------
// PATCH /v1/tasks/:id — update task
// --------------------------------------------------------------------

func TestUpdateTask(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Old what", "work", "cli")

	newWhat := "New what"
	body := map[string]interface{}{
		"what": newWhat,
	}
	rr := doRequest(t, h, "PATCH", "/v1/tasks/"+task.ID, body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var got model.Task
	decodeBody(t, rr, &got)
	assert.Equal(t, "New what", got.What)
}

func TestUpdateTask_AddLabels(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Labeled", "work", "cli", "existing")

	body := map[string]interface{}{
		"add_labels": []string{"new"},
	}
	rr := doRequest(t, h, "PATCH", "/v1/tasks/"+task.ID, body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var got model.Task
	decodeBody(t, rr, &got)
	assert.Contains(t, got.Labels, "existing")
	assert.Contains(t, got.Labels, "new")
}

func TestUpdateTask_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]interface{}{"what": "test"}
	rr := doRequest(t, h, "PATCH", "/v1/tasks/t_nonexistent", body)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --------------------------------------------------------------------
// POST /v1/tasks/:id/done — mark done
// --------------------------------------------------------------------

func TestDoneTask(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Complete me", "work", "cli")

	rr := doRequest(t, h, "POST", "/v1/tasks/"+task.ID+"/done", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var got model.Task
	decodeBody(t, rr, &got)
	assert.Equal(t, model.StateDone, got.State)
	assert.NotNil(t, got.ResolvedAt)
}

func TestDoneTask_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := doRequest(t, h, "POST", "/v1/tasks/t_nonexistent/done", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDoneTask_AlreadyDone(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Already done", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	rr := doRequest(t, h, "POST", "/v1/tasks/"+task.ID+"/done", nil)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// --------------------------------------------------------------------
// POST /v1/tasks/:id/kill — mark killed
// --------------------------------------------------------------------

func TestKillTask(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Kill me", "work", "cli")

	body := map[string]interface{}{"reason": "no longer relevant"}
	rr := doRequest(t, h, "POST", "/v1/tasks/"+task.ID+"/kill", body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var got model.Task
	decodeBody(t, rr, &got)
	assert.Equal(t, model.StateKilled, got.State)
	require.NotNil(t, got.ResolvedReason)
	assert.Equal(t, "no longer relevant", *got.ResolvedReason)
}

func TestKillTask_MissingReason(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Kill no reason", "work", "cli")

	body := map[string]interface{}{}
	rr := doRequest(t, h, "POST", "/v1/tasks/"+task.ID+"/kill", body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp map[string]string
	decodeBody(t, rr, &errResp)
	assert.Equal(t, "REASON_REQUIRED", errResp["code"])
}

func TestKillTask_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]interface{}{"reason": "test"}
	rr := doRequest(t, h, "POST", "/v1/tasks/t_nonexistent/kill", body)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --------------------------------------------------------------------
// POST /v1/seen/:agent — mark seen
// --------------------------------------------------------------------

func TestMarkSeen_ByTaskIDs(t *testing.T) {
	h, s := newTestHandler(t)
	t1 := seedTask(t, s, "Task A", "work", "cli")
	t2 := seedTask(t, s, "Task B", "work", "cli")

	body := map[string]interface{}{
		"task_ids": []string{t1.ID, t2.ID},
	}
	rr := doRequest(t, h, "POST", "/v1/seen/johanna", body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	decodeBody(t, rr, &resp)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, float64(2), resp["count"])
}

func TestMarkSeen_ByList(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Task A", "ops", "cli")
	seedTask(t, s, "Task B", "ops", "cli")

	body := map[string]interface{}{
		"list": "ops",
	}
	rr := doRequest(t, h, "POST", "/v1/seen/johanna", body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	decodeBody(t, rr, &resp)
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "ops", resp["list"])
}

func TestMarkSeen_NoInput(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]interface{}{}
	rr := doRequest(t, h, "POST", "/v1/seen/johanna", body)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --------------------------------------------------------------------
// POST /v1/labels/:name/clear — clear label
// --------------------------------------------------------------------

func TestClearLabel(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Task A", "work", "cli", "today")
	seedTask(t, s, "Task B", "work", "cli", "today")

	rr := doRequest(t, h, "POST", "/v1/labels/today/clear", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	decodeBody(t, rr, &resp)
	assert.Equal(t, "today", resp["label"])
	assert.Equal(t, float64(2), resp["cleared"])
}

func TestClearLabel_FilterByList(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Work", "work", "cli", "today")
	seedTask(t, s, "Personal", "personal", "cli", "today")

	rr := doRequest(t, h, "POST", "/v1/labels/today/clear?list=work", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	decodeBody(t, rr, &resp)
	assert.Equal(t, float64(1), resp["cleared"])
}

// --------------------------------------------------------------------
// GET /v1/labels — list labels
// --------------------------------------------------------------------

func TestListLabels(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Task A", "work", "cli", "bug")
	seedTask(t, s, "Task B", "work", "cli", "bug", "urgent")

	rr := doRequest(t, h, "GET", "/v1/labels", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var labels []model.LabelCount
	decodeBody(t, rr, &labels)
	assert.Len(t, labels, 2)
}

func TestListLabels_FilterByList(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "Work", "work", "cli", "bug")
	seedTask(t, s, "Personal", "personal", "cli", "feature")

	rr := doRequest(t, h, "GET", "/v1/labels?list=work", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var labels []model.LabelCount
	decodeBody(t, rr, &labels)
	require.Len(t, labels, 1)
	assert.Equal(t, "bug", labels[0].Label)
}

// --------------------------------------------------------------------
// GET /v1/lists — list all lists
// --------------------------------------------------------------------

func TestListLists(t *testing.T) {
	h, s := newTestHandler(t)
	seedTask(t, s, "A", "work", "cli")
	seedTask(t, s, "B", "personal", "cli")

	rr := doRequest(t, h, "GET", "/v1/lists", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var lists []model.ListInfo
	decodeBody(t, rr, &lists)
	assert.Len(t, lists, 2)
}

// --------------------------------------------------------------------
// GET /v1/stats — aggregate stats
// --------------------------------------------------------------------

func TestStats(t *testing.T) {
	h, s := newTestHandler(t)
	task := seedTask(t, s, "Open", "work", "cli")
	seedTask(t, s, "Also open", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	rr := doRequest(t, h, "GET", "/v1/stats", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats model.Stats
	decodeBody(t, rr, &stats)
	assert.Equal(t, 1, stats.TotalOpen)
	assert.Equal(t, 1, stats.TotalDone)
}

// --------------------------------------------------------------------
// Error handling
// --------------------------------------------------------------------

func TestErrorResponse_Format(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := doRequest(t, h, "GET", "/v1/tasks/t_nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]string
	decodeBody(t, rr, &resp)
	assert.Equal(t, "NOT_FOUND", resp["code"])
	assert.Contains(t, resp["error"], "not found")
}

// --------------------------------------------------------------------
// Integration: full lifecycle via HTTP
// --------------------------------------------------------------------

func TestLifecycle_CreateListDoneKill(t *testing.T) {
	h, _ := newTestHandler(t)

	// Create two tasks.
	rr1 := doRequest(t, h, "POST", "/v1/tasks", map[string]interface{}{
		"list": "work", "what": "Task to complete",
	})
	require.Equal(t, http.StatusCreated, rr1.Code)
	var task1 model.Task
	decodeBody(t, rr1, &task1)

	rr2 := doRequest(t, h, "POST", "/v1/tasks", map[string]interface{}{
		"list": "work", "what": "Task to kill",
	})
	require.Equal(t, http.StatusCreated, rr2.Code)
	var task2 model.Task
	decodeBody(t, rr2, &task2)

	// List should show both.
	rrList := doRequest(t, h, "GET", "/v1/tasks?list=work", nil)
	var tasks []*model.Task
	decodeBody(t, rrList, &tasks)
	assert.Len(t, tasks, 2)

	// Complete the first.
	rrDone := doRequest(t, h, "POST", fmt.Sprintf("/v1/tasks/%s/done", task1.ID), nil)
	assert.Equal(t, http.StatusOK, rrDone.Code)

	// Kill the second.
	rrKill := doRequest(t, h, "POST", fmt.Sprintf("/v1/tasks/%s/kill", task2.ID),
		map[string]interface{}{"reason": "changed mind"})
	assert.Equal(t, http.StatusOK, rrKill.Code)

	// Stats should reflect the changes.
	rrStats := doRequest(t, h, "GET", "/v1/stats", nil)
	var stats model.Stats
	decodeBody(t, rrStats, &stats)
	assert.Equal(t, 0, stats.TotalOpen)
	assert.Equal(t, 1, stats.TotalDone)
	assert.Equal(t, 1, stats.TotalKilled)
}

func TestLifecycle_SeenTracking(t *testing.T) {
	h, _ := newTestHandler(t)

	// Create tasks.
	doRequest(t, h, "POST", "/v1/tasks", map[string]interface{}{
		"list": "ops", "what": "Task A",
	})
	doRequest(t, h, "POST", "/v1/tasks", map[string]interface{}{
		"list": "ops", "what": "Task B",
	})

	// Unseen by agent — returns both.
	rrUnseen := doRequest(t, h, "GET", "/v1/tasks?unseen_by=agent1", nil)
	var unseen []*model.Task
	decodeBody(t, rrUnseen, &unseen)
	assert.Len(t, unseen, 2)

	// Mark all seen.
	doRequest(t, h, "POST", "/v1/seen/agent1", map[string]interface{}{
		"list": "ops",
	})

	// Now unseen returns empty.
	rrUnseen2 := doRequest(t, h, "GET", "/v1/tasks?unseen_by=agent1", nil)
	var unseen2 []*model.Task
	decodeBody(t, rrUnseen2, &unseen2)
	assert.Empty(t, unseen2)
}
