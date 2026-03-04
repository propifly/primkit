package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	mcpsvr "github.com/mark3labs/mcp-go/server"
	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/propifly/primkit/taskprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------

// newTestServer creates an MCP server backed by an in-memory SQLite store.
func newTestServer(t *testing.T) (*mcpsvr.MCPServer, store.Store) {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	srv := New(s, "test")
	return srv, s
}

// callTool invokes an MCP tool by name with the given arguments via JSON-RPC.
// Returns the result content text and any error.
func callTool(t *testing.T, srv *mcpsvr.MCPServer, name string, args map[string]interface{}) (string, bool) {
	t.Helper()

	// Build a JSON-RPC request for tools/call.
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	reqJSON, err := json.Marshal(request)
	require.NoError(t, err)

	resp := srv.HandleMessage(context.Background(), reqJSON)

	// Parse the response.
	respJSON, err := json.Marshal(resp)
	require.NoError(t, err)

	var result struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(respJSON, &result))

	if result.Error != nil {
		return result.Error.Message, true
	}

	if len(result.Result.Content) == 0 {
		return "", result.Result.IsError
	}

	return result.Result.Content[0].Text, result.Result.IsError
}

// callToolJSON invokes a tool and unmarshals the result text as JSON.
func callToolJSON(t *testing.T, srv *mcpsvr.MCPServer, name string, args map[string]interface{}, target interface{}) {
	t.Helper()
	text, isErr := callTool(t, srv, name, args)
	require.False(t, isErr, "tool returned error: %s", text)
	require.NoError(t, json.Unmarshal([]byte(text), target), "unmarshal failed for: %s", text)
}

// seedTask adds a task directly to the store.
func seedTask(t *testing.T, s store.Store, what, list, source string, labels ...string) *model.Task {
	t.Helper()
	task := &model.Task{What: what, List: list, Source: source, Labels: labels}
	require.NoError(t, s.CreateTask(context.Background(), task))
	return task
}

// --------------------------------------------------------------------
// Tool listing
// --------------------------------------------------------------------

func TestToolsRegistered(t *testing.T) {
	srv, _ := newTestServer(t)

	// List tools via JSON-RPC.
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}
	reqJSON, _ := json.Marshal(request)
	resp := srv.HandleMessage(context.Background(), reqJSON)
	respJSON, _ := json.Marshal(resp)

	var result struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(respJSON, &result))

	// Verify all 11 tools are registered.
	toolNames := make(map[string]bool)
	for _, tool := range result.Result.Tools {
		toolNames[tool.Name] = true
	}
	expected := []string{
		"taskprim_add", "taskprim_list", "taskprim_get",
		"taskprim_done", "taskprim_kill", "taskprim_edit",
		"taskprim_seen", "taskprim_label_clear", "taskprim_labels",
		"taskprim_lists", "taskprim_stats",
	}
	for _, name := range expected {
		assert.True(t, toolNames[name], "missing tool: %s", name)
	}
}

// --------------------------------------------------------------------
// taskprim_add
// --------------------------------------------------------------------

func TestAdd(t *testing.T) {
	srv, _ := newTestServer(t)

	var task model.Task
	callToolJSON(t, srv, "taskprim_add", map[string]interface{}{
		"what":   "Buy groceries",
		"list":   "personal",
		"labels": []string{"errand"},
	}, &task)

	assert.Equal(t, "Buy groceries", task.What)
	assert.Equal(t, "personal", task.List)
	assert.Equal(t, model.StateOpen, task.State)
	assert.Contains(t, task.Labels, "errand")
	assert.NotEmpty(t, task.ID)
}

func TestAdd_MissingRequired(t *testing.T) {
	srv, _ := newTestServer(t)

	// Missing "what" — should return error.
	_, isErr := callTool(t, srv, "taskprim_add", map[string]interface{}{
		"list": "work",
	})
	assert.True(t, isErr)
}

// --------------------------------------------------------------------
// taskprim_list
// --------------------------------------------------------------------

func TestList(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "Task A", "work", "cli")
	seedTask(t, s, "Task B", "personal", "cli")

	var tasks []*model.Task
	callToolJSON(t, srv, "taskprim_list", map[string]interface{}{}, &tasks)
	assert.Len(t, tasks, 2)
}

func TestList_FilterByList(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "Work", "work", "cli")
	seedTask(t, s, "Personal", "personal", "cli")

	var tasks []*model.Task
	callToolJSON(t, srv, "taskprim_list", map[string]interface{}{
		"list": "work",
	}, &tasks)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Work", tasks[0].What)
}

func TestList_FilterByLabel(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "Urgent", "work", "cli", "urgent")
	seedTask(t, s, "Normal", "work", "cli")

	var tasks []*model.Task
	callToolJSON(t, srv, "taskprim_list", map[string]interface{}{
		"labels": []string{"urgent"},
	}, &tasks)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Urgent", tasks[0].What)
}

// --------------------------------------------------------------------
// taskprim_get
// --------------------------------------------------------------------

func TestGet(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Get me", "work", "cli")

	var got model.Task
	callToolJSON(t, srv, "taskprim_get", map[string]interface{}{
		"id": task.ID,
	}, &got)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, "Get me", got.What)
}

func TestGet_NotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	text, isErr := callTool(t, srv, "taskprim_get", map[string]interface{}{
		"id": "t_nonexistent",
	})
	assert.True(t, isErr)
	assert.Contains(t, text, "not found")
}

// --------------------------------------------------------------------
// taskprim_done
// --------------------------------------------------------------------

func TestDone(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Complete me", "work", "cli")

	var got model.Task
	callToolJSON(t, srv, "taskprim_done", map[string]interface{}{
		"id": task.ID,
	}, &got)
	assert.Equal(t, model.StateDone, got.State)
}

func TestDone_AlreadyDone(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Already", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	text, isErr := callTool(t, srv, "taskprim_done", map[string]interface{}{
		"id": task.ID,
	})
	assert.True(t, isErr)
	assert.Contains(t, text, "completing task")
}

// --------------------------------------------------------------------
// taskprim_kill
// --------------------------------------------------------------------

func TestKill(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Kill me", "work", "cli")

	var got model.Task
	callToolJSON(t, srv, "taskprim_kill", map[string]interface{}{
		"id":     task.ID,
		"reason": "no longer needed",
	}, &got)
	assert.Equal(t, model.StateKilled, got.State)
	require.NotNil(t, got.ResolvedReason)
	assert.Equal(t, "no longer needed", *got.ResolvedReason)
}

func TestKill_MissingReason(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Kill no reason", "work", "cli")

	_, isErr := callTool(t, srv, "taskprim_kill", map[string]interface{}{
		"id": task.ID,
	})
	assert.True(t, isErr)
}

// --------------------------------------------------------------------
// taskprim_edit
// --------------------------------------------------------------------

func TestEdit(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Old description", "work", "cli")

	var got model.Task
	callToolJSON(t, srv, "taskprim_edit", map[string]interface{}{
		"id":   task.ID,
		"what": "New description",
	}, &got)
	assert.Equal(t, "New description", got.What)
}

func TestEdit_Labels(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Labeled", "work", "cli", "keep", "remove")

	var got model.Task
	callToolJSON(t, srv, "taskprim_edit", map[string]interface{}{
		"id":         task.ID,
		"add_labels": []string{"new"},
		"del_labels": []string{"remove"},
	}, &got)
	assert.Contains(t, got.Labels, "keep")
	assert.Contains(t, got.Labels, "new")
	assert.NotContains(t, got.Labels, "remove")
}

// --------------------------------------------------------------------
// taskprim_seen
// --------------------------------------------------------------------

func TestSeen_ByTaskIDs(t *testing.T) {
	srv, s := newTestServer(t)
	t1 := seedTask(t, s, "A", "work", "cli")
	t2 := seedTask(t, s, "B", "work", "cli")

	text, isErr := callTool(t, srv, "taskprim_seen", map[string]interface{}{
		"agent":    "johanna",
		"task_ids": []string{t1.ID, t2.ID},
	})
	assert.False(t, isErr)
	assert.Contains(t, text, "2 task(s)")
}

func TestSeen_ByList(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "A", "ops", "cli")

	text, isErr := callTool(t, srv, "taskprim_seen", map[string]interface{}{
		"agent": "johanna",
		"list":  "ops",
	})
	assert.False(t, isErr)
	assert.Contains(t, text, "ops")
}

func TestSeen_NoInput(t *testing.T) {
	srv, _ := newTestServer(t)

	text, isErr := callTool(t, srv, "taskprim_seen", map[string]interface{}{
		"agent": "johanna",
	})
	assert.True(t, isErr)
	assert.Contains(t, text, "provide either")
}

// --------------------------------------------------------------------
// taskprim_label_clear
// --------------------------------------------------------------------

func TestLabelClear(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "A", "work", "cli", "today")
	seedTask(t, s, "B", "work", "cli", "today")

	text, isErr := callTool(t, srv, "taskprim_label_clear", map[string]interface{}{
		"label": "today",
	})
	assert.False(t, isErr)
	assert.Contains(t, text, "2 task(s)")
}

// --------------------------------------------------------------------
// taskprim_labels
// --------------------------------------------------------------------

func TestLabels(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "A", "work", "cli", "bug")
	seedTask(t, s, "B", "work", "cli", "bug", "urgent")

	var labels []model.LabelCount
	callToolJSON(t, srv, "taskprim_labels", map[string]interface{}{}, &labels)
	assert.Len(t, labels, 2)
}

// --------------------------------------------------------------------
// taskprim_lists
// --------------------------------------------------------------------

func TestLists(t *testing.T) {
	srv, s := newTestServer(t)
	seedTask(t, s, "A", "work", "cli")
	seedTask(t, s, "B", "personal", "cli")

	var lists []model.ListInfo
	callToolJSON(t, srv, "taskprim_lists", map[string]interface{}{}, &lists)
	assert.Len(t, lists, 2)
}

// --------------------------------------------------------------------
// taskprim_stats
// --------------------------------------------------------------------

func TestStats(t *testing.T) {
	srv, s := newTestServer(t)
	task := seedTask(t, s, "Open", "work", "cli")
	seedTask(t, s, "Also open", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	var stats model.Stats
	callToolJSON(t, srv, "taskprim_stats", map[string]interface{}{}, &stats)
	assert.Equal(t, 1, stats.TotalOpen)
	assert.Equal(t, 1, stats.TotalDone)
}

// --------------------------------------------------------------------
// Integration: full lifecycle via MCP
// --------------------------------------------------------------------

func TestLifecycle(t *testing.T) {
	srv, _ := newTestServer(t)

	// Create a task.
	var created model.Task
	callToolJSON(t, srv, "taskprim_add", map[string]interface{}{
		"what":   "Lifecycle task",
		"list":   "work",
		"labels": []string{"test"},
	}, &created)
	require.NotEmpty(t, created.ID)

	// Get it back.
	var fetched model.Task
	callToolJSON(t, srv, "taskprim_get", map[string]interface{}{
		"id": created.ID,
	}, &fetched)
	assert.Equal(t, "Lifecycle task", fetched.What)

	// Edit it.
	var edited model.Task
	callToolJSON(t, srv, "taskprim_edit", map[string]interface{}{
		"id":   created.ID,
		"what": "Updated lifecycle task",
	}, &edited)
	assert.Equal(t, "Updated lifecycle task", edited.What)

	// Mark seen.
	text, isErr := callTool(t, srv, "taskprim_seen", map[string]interface{}{
		"agent":    "test-agent",
		"task_ids": []string{created.ID},
	})
	assert.False(t, isErr)
	assert.Contains(t, text, "1 task(s)")

	// Mark done.
	var done model.Task
	callToolJSON(t, srv, "taskprim_done", map[string]interface{}{
		"id": created.ID,
	}, &done)
	assert.Equal(t, model.StateDone, done.State)

	// Stats should show 1 done.
	var stats model.Stats
	callToolJSON(t, srv, "taskprim_stats", map[string]interface{}{}, &stats)
	assert.Equal(t, 1, stats.TotalDone)
	assert.Equal(t, 0, stats.TotalOpen)
}

// Suppress unused import warning.
var _ = fmt.Sprint
