package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/propifly/primkit/taskprim/internal/store"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------

// newTestStore creates an in-memory SQLite store for testing. Uses the same
// approach as the store package tests — real SQLite, real SQL, real migrations.
func newTestStore(t *testing.T) store.Store {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

// execCmd runs a CLI command with the given args against an in-memory store.
// Returns the captured stdout and any error from execution.
func execCmd(t *testing.T, s store.Store, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	// Override PersistentPreRunE to inject our test store instead of opening
	// a real database. This keeps tests fast and isolated.
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := context.WithValue(cmd.Context(), storeKey, s)
		cmd.SetContext(ctx)
		return nil
	}

	err := root.Execute()
	return buf.String(), err
}

// seedTask adds a task to the store and returns it. Convenience for tests
// that need pre-existing data.
func seedTask(t *testing.T, s store.Store, what, list, source string, labels ...string) *model.Task {
	t.Helper()
	task := &model.Task{
		What:   what,
		List:   list,
		Source: source,
		Labels: labels,
	}
	err := s.CreateTask(context.Background(), task)
	require.NoError(t, err)
	return task
}

// --------------------------------------------------------------------
// add command
// --------------------------------------------------------------------

func TestAdd_BasicTask(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "add", "Buy groceries")
	require.NoError(t, err)

	// Table output should show the task ID and description.
	assert.Contains(t, out, "Buy groceries")
	assert.Contains(t, out, "t_")
}

func TestAdd_WithAllFlags(t *testing.T) {
	s := newTestStore(t)

	// Create a parent task first (foreign key constraint requires it to exist).
	parent := seedTask(t, s, "Parent task", "ops", "cli")

	out, err := execCmd(t, s, "add", "Deploy service",
		"--list", "ops",
		"--source", "johanna",
		"--label", "infra,urgent",
		"--waiting-on", "DNS propagation",
		"--context", "Waiting for TTL expiry",
		"--parent", parent.ID,
	)
	require.NoError(t, err)

	assert.Contains(t, out, "Deploy service")
	assert.Contains(t, out, "ops")

	// Verify the task was persisted with all fields. The parent is also in
	// the list, so we filter for the child by checking source.
	tasks, _ := s.ListTasks(context.Background(), &model.Filter{List: "ops", Source: "johanna"})
	require.Len(t, tasks, 1)
	assert.Equal(t, "johanna", tasks[0].Source)
	assert.ElementsMatch(t, []string{"infra", "urgent"}, tasks[0].Labels)
	require.NotNil(t, tasks[0].WaitingOn)
	assert.Equal(t, "DNS propagation", *tasks[0].WaitingOn)
	require.NotNil(t, tasks[0].Context)
	assert.Equal(t, "Waiting for TTL expiry", *tasks[0].Context)
	require.NotNil(t, tasks[0].ParentID)
	assert.Equal(t, parent.ID, *tasks[0].ParentID)
}

func TestAdd_JSON(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "add", "JSON task", "--format", "json")
	require.NoError(t, err)

	var task model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &task))
	assert.Equal(t, "JSON task", task.What)
	assert.True(t, strings.HasPrefix(task.ID, "t_"))
}

func TestAdd_Quiet(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "add", "Quiet task", "--format", "quiet")
	require.NoError(t, err)

	// Quiet mode outputs only the task ID.
	id := strings.TrimSpace(out)
	assert.True(t, strings.HasPrefix(id, "t_"), "expected task ID, got: %q", id)
	assert.NotContains(t, out, "Quiet task")
}

func TestAdd_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "add")
	require.Error(t, err)
}

func TestAdd_MultiWordDescription(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "add", "Reply", "to", "the", "accountant")
	require.NoError(t, err)

	// Multiple positional args are joined into the task description.
	assert.Contains(t, out, "Reply to the accountant")
}

// --------------------------------------------------------------------
// list command
// --------------------------------------------------------------------

func TestList_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "No tasks found")
}

func TestList_ShowsTasks(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "First task", "work", "cli")
	seedTask(t, s, "Second task", "work", "cli")

	out, err := execCmd(t, s, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "First task")
	assert.Contains(t, out, "Second task")
}

func TestList_FilterByList(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Work task", "work", "cli")
	seedTask(t, s, "Personal task", "personal", "cli")

	out, err := execCmd(t, s, "list", "--list", "personal")
	require.NoError(t, err)
	assert.Contains(t, out, "Personal task")
	assert.NotContains(t, out, "Work task")
}

func TestList_FilterByState(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Done task", "work", "cli")
	seedTask(t, s, "Open task", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	out, err := execCmd(t, s, "list", "--state", "done")
	require.NoError(t, err)
	assert.Contains(t, out, "Done task")
	assert.NotContains(t, out, "Open task")
}

func TestList_FilterByLabel(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Urgent item", "work", "cli", "urgent")
	seedTask(t, s, "Normal item", "work", "cli")

	out, err := execCmd(t, s, "list", "--label", "urgent")
	require.NoError(t, err)
	assert.Contains(t, out, "Urgent item")
	assert.NotContains(t, out, "Normal item")
}

func TestList_JSONFormat(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Test task", "work", "cli")

	out, err := execCmd(t, s, "list", "--format", "json")
	require.NoError(t, err)

	var tasks []*model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &tasks))
	require.Len(t, tasks, 1)
	assert.Equal(t, "Test task", tasks[0].What)
}

func TestList_QuietFormat(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Quiet list", "work", "cli")

	out, err := execCmd(t, s, "list", "--format", "quiet")
	require.NoError(t, err)
	assert.Contains(t, out, task.ID)
	assert.NotContains(t, out, "Quiet list")
}

// --------------------------------------------------------------------
// get command
// --------------------------------------------------------------------

func TestGet_ExistingTask(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Get me", "work", "cli", "important")

	out, err := execCmd(t, s, "get", task.ID)
	require.NoError(t, err)
	assert.Contains(t, out, task.ID)
	assert.Contains(t, out, "Get me")
	assert.Contains(t, out, "important")
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "get", "t_nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGet_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "get")
	require.Error(t, err)
}

func TestGet_JSON(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "JSON get", "work", "cli")

	out, err := execCmd(t, s, "get", task.ID, "--format", "json")
	require.NoError(t, err)

	var got model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, task.ID, got.ID)
}

// --------------------------------------------------------------------
// done command
// --------------------------------------------------------------------

func TestDone_SingleTask(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Finish this", "work", "cli")

	out, err := execCmd(t, s, "done", task.ID)
	require.NoError(t, err)
	assert.Contains(t, out, "Done: "+task.ID)

	// Verify state changed.
	got, _ := s.GetTask(context.Background(), task.ID)
	assert.Equal(t, model.StateDone, got.State)
}

func TestDone_MultipleTasks(t *testing.T) {
	s := newTestStore(t)
	t1 := seedTask(t, s, "Task one", "work", "cli")
	t2 := seedTask(t, s, "Task two", "work", "cli")

	out, err := execCmd(t, s, "done", t1.ID, t2.ID)
	require.NoError(t, err)
	assert.Contains(t, out, "Done: "+t1.ID)
	assert.Contains(t, out, "Done: "+t2.ID)
}

func TestDone_AlreadyDone(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Already done", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	_, err := execCmd(t, s, "done", task.ID)
	require.Error(t, err)
}

func TestDone_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "done", "t_nonexistent")
	require.Error(t, err)
}

func TestDone_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "done")
	require.Error(t, err)
}

// --------------------------------------------------------------------
// kill command
// --------------------------------------------------------------------

func TestKill_WithReason(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Kill this", "work", "cli")

	out, err := execCmd(t, s, "kill", task.ID, "--reason", "no longer needed")
	require.NoError(t, err)
	assert.Contains(t, out, "Killed: "+task.ID)
	assert.Contains(t, out, "no longer needed")

	// Verify state and reason.
	got, _ := s.GetTask(context.Background(), task.ID)
	assert.Equal(t, model.StateKilled, got.State)
	require.NotNil(t, got.ResolvedReason)
	assert.Equal(t, "no longer needed", *got.ResolvedReason)
}

func TestKill_MissingReason(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Kill no reason", "work", "cli")

	_, err := execCmd(t, s, "kill", task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--reason is required")
}

func TestKill_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "kill", "t_nonexistent", "--reason", "test")
	require.Error(t, err)
}

func TestKill_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "kill")
	require.Error(t, err)
}

// --------------------------------------------------------------------
// edit command
// --------------------------------------------------------------------

func TestEdit_UpdateWhat(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Old description", "work", "cli")

	out, err := execCmd(t, s, "edit", task.ID, "--what", "New description")
	require.NoError(t, err)
	assert.Contains(t, out, "New description")

	got, _ := s.GetTask(context.Background(), task.ID)
	assert.Equal(t, "New description", got.What)
}

func TestEdit_MoveList(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Move me", "old-list", "cli")

	out, err := execCmd(t, s, "edit", task.ID, "--list", "new-list")
	require.NoError(t, err)
	assert.Contains(t, out, "new-list")

	got, _ := s.GetTask(context.Background(), task.ID)
	assert.Equal(t, "new-list", got.List)
}

func TestEdit_AddAndRemoveLabels(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Labeled task", "work", "cli", "keep", "remove")

	out, err := execCmd(t, s, "edit", task.ID, "--add-label", "new", "--del-label", "remove")
	require.NoError(t, err)
	assert.Contains(t, out, task.ID)

	got, _ := s.GetTask(context.Background(), task.ID)
	assert.Contains(t, got.Labels, "keep")
	assert.Contains(t, got.Labels, "new")
	assert.NotContains(t, got.Labels, "remove")
}

func TestEdit_NoFlags(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Unchanged", "work", "cli")

	_, err := execCmd(t, s, "edit", task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fields specified")
}

func TestEdit_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "edit", "t_nonexistent", "--what", "test")
	require.Error(t, err)
}

func TestEdit_SetWaitingOn(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Blocked", "work", "cli")

	_, err := execCmd(t, s, "edit", task.ID, "--waiting-on", "approval from lead")
	require.NoError(t, err)

	got, _ := s.GetTask(context.Background(), task.ID)
	require.NotNil(t, got.WaitingOn)
	assert.Equal(t, "approval from lead", *got.WaitingOn)
}

func TestEdit_JSON(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Edit JSON", "work", "cli")

	out, err := execCmd(t, s, "edit", task.ID, "--what", "Updated", "--format", "json")
	require.NoError(t, err)

	var got model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "Updated", got.What)
}

// --------------------------------------------------------------------
// seen command
// --------------------------------------------------------------------

func TestSeen_SpecificTasks(t *testing.T) {
	s := newTestStore(t)
	t1 := seedTask(t, s, "See this", "work", "cli")
	t2 := seedTask(t, s, "And this", "work", "cli")

	out, err := execCmd(t, s, "seen", "johanna", t1.ID, t2.ID)
	require.NoError(t, err)
	assert.Contains(t, out, "Marked 2 task(s) as seen by johanna")
}

func TestSeen_AllInList(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Task A", "ops", "cli")
	seedTask(t, s, "Task B", "ops", "cli")

	out, err := execCmd(t, s, "seen", "johanna", "--list", "ops")
	require.NoError(t, err)
	assert.Contains(t, out, "Marked all open tasks")
	assert.Contains(t, out, "johanna")
}

func TestSeen_NoTasksOrList(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "seen", "johanna")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specify task IDs or --list")
}

func TestSeen_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "seen")
	require.Error(t, err)
}

// --------------------------------------------------------------------
// labels command
// --------------------------------------------------------------------

func TestLabels_ShowsCounts(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Task A", "work", "cli", "bug")
	seedTask(t, s, "Task B", "work", "cli", "bug", "urgent")

	out, err := execCmd(t, s, "labels")
	require.NoError(t, err)
	assert.Contains(t, out, "bug")
	assert.Contains(t, out, "urgent")
}

func TestLabels_FilterByList(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Work bug", "work", "cli", "bug")
	seedTask(t, s, "Personal bug", "personal", "cli", "bug")

	out, err := execCmd(t, s, "labels", "--list", "work")
	require.NoError(t, err)
	assert.Contains(t, out, "bug")
}

func TestLabels_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "labels")
	require.NoError(t, err)
	assert.Contains(t, out, "No labels found")
}

func TestLabels_JSON(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Labeled", "work", "cli", "feature")

	out, err := execCmd(t, s, "labels", "--format", "json")
	require.NoError(t, err)

	var labels []model.LabelCount
	require.NoError(t, json.Unmarshal([]byte(out), &labels))
	require.Len(t, labels, 1)
	assert.Equal(t, "feature", labels[0].Label)
}

// --------------------------------------------------------------------
// labels clear subcommand
// --------------------------------------------------------------------

func TestLabelClear(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Remove label", "work", "cli", "today")
	seedTask(t, s, "Also has label", "work", "cli", "today")

	out, err := execCmd(t, s, "labels", "clear", "today")
	require.NoError(t, err)
	assert.Contains(t, out, `Cleared "today" from 2 task(s)`)
}

func TestLabelClear_FilterByList(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Work", "work", "cli", "today")
	seedTask(t, s, "Personal", "personal", "cli", "today")

	out, err := execCmd(t, s, "labels", "clear", "today", "--list", "work")
	require.NoError(t, err)
	assert.Contains(t, out, `Cleared "today" from 1 task(s)`)
}

func TestLabelClear_NoArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "labels", "clear")
	require.Error(t, err)
}

// --------------------------------------------------------------------
// lists command
// --------------------------------------------------------------------

func TestLists_ShowsCounts(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Open task", "work", "cli")
	seedTask(t, s, "Another open", "work", "cli")
	seedTask(t, s, "Personal", "personal", "cli")
	s.DoneTask(context.Background(), task.ID)

	out, err := execCmd(t, s, "lists")
	require.NoError(t, err)
	assert.Contains(t, out, "work")
	assert.Contains(t, out, "personal")
}

func TestLists_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "lists")
	require.NoError(t, err)
	assert.Contains(t, out, "No lists found")
}

func TestLists_JSON(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Task", "mylist", "cli")

	out, err := execCmd(t, s, "lists", "--format", "json")
	require.NoError(t, err)

	var lists []model.ListInfo
	require.NoError(t, json.Unmarshal([]byte(out), &lists))
	require.Len(t, lists, 1)
	assert.Equal(t, "mylist", lists[0].Name)
}

// --------------------------------------------------------------------
// stats command
// --------------------------------------------------------------------

func TestStats_ShowsCounts(t *testing.T) {
	s := newTestStore(t)
	t1 := seedTask(t, s, "Open", "work", "cli")
	t2 := seedTask(t, s, "Will be done", "work", "cli")
	seedTask(t, s, "Also open", "work", "cli")
	s.DoneTask(context.Background(), t1.ID)
	s.KillTask(context.Background(), t2.ID, "testing")

	out, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Open:   1")
	assert.Contains(t, out, "Done:   1")
	assert.Contains(t, out, "Killed: 1")
	assert.Contains(t, out, "Total:  3")
}

func TestStats_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Open:   0")
	assert.Contains(t, out, "Total:  0")
}

func TestStats_JSON(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Task", "work", "cli")

	out, err := execCmd(t, s, "stats", "--format", "json")
	require.NoError(t, err)

	var stats model.Stats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.Equal(t, 1, stats.TotalOpen)
}

// --------------------------------------------------------------------
// export command
// --------------------------------------------------------------------

func TestExport_AllTasks(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Export me", "work", "cli", "label1")
	seedTask(t, s, "And me", "personal", "cli")

	out, err := execCmd(t, s, "export")
	require.NoError(t, err)

	var tasks []*model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &tasks))
	assert.Len(t, tasks, 2)
}

func TestExport_FilterByList(t *testing.T) {
	s := newTestStore(t)
	seedTask(t, s, "Work task", "work", "cli")
	seedTask(t, s, "Personal task", "personal", "cli")

	out, err := execCmd(t, s, "export", "--list", "work")
	require.NoError(t, err)

	var tasks []*model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &tasks))
	require.Len(t, tasks, 1)
	assert.Equal(t, "Work task", tasks[0].What)
}

func TestExport_FilterByState(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s, "Done task", "work", "cli")
	seedTask(t, s, "Open task", "work", "cli")
	s.DoneTask(context.Background(), task.ID)

	out, err := execCmd(t, s, "export", "--state", "done")
	require.NoError(t, err)

	var tasks []*model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &tasks))
	require.Len(t, tasks, 1)
	assert.Equal(t, "Done task", tasks[0].What)
}

func TestExport_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "export")
	require.NoError(t, err)

	var tasks []*model.Task
	require.NoError(t, json.Unmarshal([]byte(out), &tasks))
	assert.Empty(t, tasks)
}

// --------------------------------------------------------------------
// import command
// --------------------------------------------------------------------

func TestImport_FromFile(t *testing.T) {
	s := newTestStore(t)

	// First export some tasks.
	seedTask(t, s, "Importable", "work", "cli", "tag")

	exportOut, err := execCmd(t, s, "export")
	require.NoError(t, err)

	// Write export to a temp file.
	tmpFile := t.TempDir() + "/export.json"
	require.NoError(t, writeFile(tmpFile, exportOut))

	// Create a fresh store and import into it.
	s2 := newTestStore(t)
	out, err := execCmd(t, s2, "import", "--file", tmpFile)
	require.NoError(t, err)
	assert.Contains(t, out, "Imported 1 task(s)")
}

func TestImport_MissingFile(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "import", "--file", "/tmp/nonexistent-taskprim-test.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "opening file")
}

// writeFile is a tiny test helper to write a string to a file.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// --------------------------------------------------------------------
// format helpers
// --------------------------------------------------------------------

func TestParseDuration_Days(t *testing.T) {
	d, err := parseDuration("7d")
	require.NoError(t, err)
	assert.Equal(t, 7*24*60*60, int(d.Seconds()))
}

func TestParseDuration_Hours(t *testing.T) {
	d, err := parseDuration("24h")
	require.NoError(t, err)
	assert.Equal(t, 24*60*60, int(d.Seconds()))
}

func TestParseDuration_Invalid(t *testing.T) {
	_, err := parseDuration("abc")
	require.Error(t, err)
}

func TestTruncate_Short(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
}

func TestTruncate_Long(t *testing.T) {
	result := truncate("this is a very long string that exceeds the limit", 20)
	assert.Equal(t, "this is a very lo...", result)
	assert.Len(t, result, 20)
}

func TestSplitCSV(t *testing.T) {
	// Handles both comma-separated and repeated flags.
	assert.Equal(t, []string{"a", "b", "c"}, splitCSV([]string{"a,b", "c"}))
	assert.Equal(t, []string{"x"}, splitCSV([]string{"x"}))
	assert.Empty(t, splitCSV([]string{""}))
}

// --------------------------------------------------------------------
// Integration: full lifecycle via CLI
// --------------------------------------------------------------------

func TestLifecycle_AddListDoneKill(t *testing.T) {
	s := newTestStore(t)

	// Add two tasks.
	out1, err := execCmd(t, s, "add", "Task to complete", "--list", "work", "--format", "quiet")
	require.NoError(t, err)
	id1 := strings.TrimSpace(out1)

	out2, err := execCmd(t, s, "add", "Task to kill", "--list", "work", "--format", "quiet")
	require.NoError(t, err)
	id2 := strings.TrimSpace(out2)

	// List should show both.
	listOut, err := execCmd(t, s, "list", "--list", "work")
	require.NoError(t, err)
	assert.Contains(t, listOut, "Task to complete")
	assert.Contains(t, listOut, "Task to kill")

	// Complete the first.
	_, err = execCmd(t, s, "done", id1)
	require.NoError(t, err)

	// Kill the second.
	_, err = execCmd(t, s, "kill", id2, "--reason", "changed mind")
	require.NoError(t, err)

	// Stats should reflect the changes.
	statsOut, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, statsOut, "Open:   0")
	assert.Contains(t, statsOut, "Done:   1")
	assert.Contains(t, statsOut, "Killed: 1")
}

func TestLifecycle_SeenTracking(t *testing.T) {
	s := newTestStore(t)

	// Add tasks.
	execCmd(t, s, "add", "Task A", "--list", "ops", "--format", "quiet")
	execCmd(t, s, "add", "Task B", "--list", "ops", "--format", "quiet")

	// Agent hasn't seen anything — unseen filter returns both.
	listOut, err := execCmd(t, s, "list", "--unseen-by", "agent1", "--format", "json")
	require.NoError(t, err)
	var unseen []*model.Task
	require.NoError(t, json.Unmarshal([]byte(listOut), &unseen))
	assert.Len(t, unseen, 2)

	// Mark all seen.
	_, err = execCmd(t, s, "seen", "agent1", "--list", "ops")
	require.NoError(t, err)

	// Now unseen returns empty.
	listOut2, err := execCmd(t, s, "list", "--unseen-by", "agent1", "--format", "json")
	require.NoError(t, err)
	var unseen2 []*model.Task
	require.NoError(t, json.Unmarshal([]byte(listOut2), &unseen2))
	assert.Empty(t, unseen2)
}

// ---------------------------------------------------------------------------
// Regression tests
// ---------------------------------------------------------------------------

// TestDBPathFromConfig verifies that storage.db in the YAML config file is
// used as the database path when --db and TASKPRIM_DB are both unset.
//
// Regression for: config was loaded after the DB path fallback chain, so
// cfg.Storage.DB was never consulted and the hardcoded default always won.
func TestDBPathFromConfig(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "taskprim-test.db")
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("storage:\n  db: " + dbPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configPath, "list"})
	err := root.Execute()
	require.NoError(t, err, "command should succeed using DB path from config")

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "DB file should be created at path from config, not the default")
}
