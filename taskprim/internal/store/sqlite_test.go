package store

import (
	"context"
	"testing"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore creates a fresh in-memory store for each test. The store is
// automatically closed when the test ends.
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	store, err := NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

// seedTask creates a task with sensible defaults. Override fields as needed.
func seedTask(t *testing.T, s *SQLiteStore, overrides ...func(*model.Task)) *model.Task {
	t.Helper()
	task := &model.Task{
		List:   "andres",
		What:   "test task",
		Source: "test",
	}
	for _, fn := range overrides {
		fn(task)
	}
	require.NoError(t, s.CreateTask(context.Background(), task))
	return task
}

// ---------------------------------------------------------------------------
// CreateTask
// ---------------------------------------------------------------------------

func TestCreateTask(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := &model.Task{
		List:   "andres",
		What:   "reply to accountant",
		Source: "johanna",
		Labels: []string{"financial", "tax"},
	}
	err := s.CreateTask(ctx, task)
	require.NoError(t, err)

	assert.NotEmpty(t, task.ID, "ID should be assigned")
	assert.Equal(t, model.StateOpen, task.State, "state should be open")
	assert.False(t, task.Created.IsZero(), "created should be set")
	assert.False(t, task.Updated.IsZero(), "updated should be set")
}

func TestCreateTask_Validation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.CreateTask(ctx, &model.Task{})
	assert.Error(t, err, "empty task should fail validation")
}

func TestCreateTask_WithOptionalFields(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	waitingOn := "accountant"
	taskCtx := "Waiting for Q4 numbers"

	task := &model.Task{
		List:      "andres",
		What:      "review Q4 filing",
		Source:    "johanna",
		WaitingOn: &waitingOn,
		Context:   &taskCtx,
		Labels:    []string{"financial"},
	}
	require.NoError(t, s.CreateTask(ctx, task))

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "accountant", *got.WaitingOn)
	assert.Equal(t, "Waiting for Q4 numbers", *got.Context)
	assert.Equal(t, []string{"financial"}, got.Labels)
}

// ---------------------------------------------------------------------------
// GetTask
// ---------------------------------------------------------------------------

func TestGetTask(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	created := seedTask(t, s, func(task *model.Task) {
		task.What = "test get"
		task.Labels = []string{"a", "b"}
	})

	got, err := s.GetTask(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "test get", got.What)
	assert.Equal(t, []string{"a", "b"}, got.Labels)
}

func TestGetTask_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetTask(context.Background(), "t_nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

// ---------------------------------------------------------------------------
// ListTasks
// ---------------------------------------------------------------------------

func TestListTasks_NoFilter(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.What = "first" })
	seedTask(t, s, func(task *model.Task) { task.What = "second" })

	tasks, err := s.ListTasks(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestListTasks_ByList(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.List = "andres" })
	seedTask(t, s, func(task *model.Task) { task.List = "johanna" })
	seedTask(t, s, func(task *model.Task) { task.List = "andres" })

	tasks, err := s.ListTasks(ctx, &model.Filter{List: "andres"})
	require.NoError(t, err)
	assert.Len(t, tasks, 2, "should return only tasks in 'andres' list")
}

func TestListTasks_ByState(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	seedTask(t, s)
	require.NoError(t, s.DoneTask(ctx, task.ID))

	state := model.StateDone
	tasks, err := s.ListTasks(ctx, &model.Filter{State: &state})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, model.StateDone, tasks[0].State)
}

func TestListTasks_ByLabel(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"financial", "tax"} })
	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"financial"} })
	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"home"} })

	// Filter for tasks with the "financial" label.
	tasks, err := s.ListTasks(ctx, &model.Filter{Labels: []string{"financial"}})
	require.NoError(t, err)
	assert.Len(t, tasks, 2, "two tasks have the 'financial' label")

	// Filter for tasks with BOTH "financial" AND "tax" labels.
	tasks, err = s.ListTasks(ctx, &model.Filter{Labels: []string{"financial", "tax"}})
	require.NoError(t, err)
	assert.Len(t, tasks, 1, "one task has both labels")
}

func TestListTasks_ByWaiting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	waitingOn := "someone"
	seedTask(t, s, func(task *model.Task) { task.WaitingOn = &waitingOn })
	seedTask(t, s) // no waiting_on

	waiting := true
	tasks, err := s.ListTasks(ctx, &model.Filter{Waiting: &waiting})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.NotNil(t, tasks[0].WaitingOn)
}

func TestListTasks_Unseen(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task1 := seedTask(t, s, func(task *model.Task) { task.What = "seen" })
	seedTask(t, s, func(task *model.Task) { task.What = "unseen" })

	// Mark task1 as seen by johanna.
	require.NoError(t, s.MarkSeen(ctx, "johanna", []string{task1.ID}))

	tasks, err := s.ListTasks(ctx, &model.Filter{UnseenBy: "johanna"})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "unseen", tasks[0].What)
}

func TestListTasks_Stale(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create a task and manually backdate its updated timestamp.
	task := seedTask(t, s)
	staleTime := time.Now().UTC().Add(-8 * 24 * time.Hour) // 8 days ago
	_, err := s.db.ExecContext(ctx, "UPDATE tasks SET updated = ? WHERE id = ?", staleTime, task.ID)
	require.NoError(t, err)

	seedTask(t, s) // Fresh task.

	tasks, err := s.ListTasks(ctx, &model.Filter{Stale: 7 * 24 * time.Hour})
	require.NoError(t, err)
	assert.Len(t, tasks, 1, "should return only the stale task")
}

// ---------------------------------------------------------------------------
// UpdateTask
// ---------------------------------------------------------------------------

func TestUpdateTask(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	newWhat := "updated description"
	err := s.UpdateTask(ctx, task.ID, &model.TaskUpdate{What: &newWhat})
	require.NoError(t, err)

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated description", got.What)
	assert.True(t, got.Updated.After(task.Updated) || got.Updated.Equal(task.Updated),
		"updated timestamp should advance")
}

func TestUpdateTask_Labels(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s, func(task *model.Task) { task.Labels = []string{"a", "b"} })

	err := s.UpdateTask(ctx, task.ID, &model.TaskUpdate{
		AddLabels: []string{"c"},
		DelLabels: []string{"a"},
	})
	require.NoError(t, err)

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"b", "c"}, got.Labels, "should have b and c, not a")
}

func TestUpdateTask_ClearWaitingOn(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	waitingOn := "someone"
	task := seedTask(t, s, func(task *model.Task) { task.WaitingOn = &waitingOn })

	clear := ""
	err := s.UpdateTask(ctx, task.ID, &model.TaskUpdate{WaitingOn: &clear})
	require.NoError(t, err)

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Nil(t, got.WaitingOn, "waiting_on should be cleared")
}

func TestUpdateTask_NotFound(t *testing.T) {
	s := newTestStore(t)
	newWhat := "whatever"
	err := s.UpdateTask(context.Background(), "t_nonexistent", &model.TaskUpdate{What: &newWhat})
	assert.ErrorIs(t, err, ErrNotFound)
}

// ---------------------------------------------------------------------------
// DoneTask / KillTask
// ---------------------------------------------------------------------------

func TestDoneTask(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	require.NoError(t, s.DoneTask(ctx, task.ID))

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StateDone, got.State)
	assert.NotNil(t, got.ResolvedAt)
}

func TestDoneTask_AlreadyDone(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	require.NoError(t, s.DoneTask(ctx, task.ID))

	err := s.DoneTask(ctx, task.ID)
	assert.ErrorIs(t, err, ErrInvalidTransition, "completing an already-done task should fail")
}

func TestDoneTask_NotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.DoneTask(context.Background(), "t_nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestKillTask(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	require.NoError(t, s.KillTask(ctx, task.ID, "no longer relevant"))

	got, err := s.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StateKilled, got.State)
	assert.Equal(t, "no longer relevant", *got.ResolvedReason)
	assert.NotNil(t, got.ResolvedAt)
}

func TestKillTask_RequiresReason(t *testing.T) {
	s := newTestStore(t)
	task := seedTask(t, s)
	err := s.KillTask(context.Background(), task.ID, "")
	assert.Error(t, err, "killing without a reason should fail")
}

func TestKillTask_AlreadyKilled(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	require.NoError(t, s.KillTask(ctx, task.ID, "reason"))

	err := s.KillTask(ctx, task.ID, "another reason")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

// ---------------------------------------------------------------------------
// Seen tracking
// ---------------------------------------------------------------------------

func TestMarkSeen(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	err := s.MarkSeen(ctx, "johanna", []string{task.ID})
	require.NoError(t, err)

	// The task should now NOT appear in unseen results for johanna.
	tasks, err := s.ListTasks(ctx, &model.Filter{UnseenBy: "johanna"})
	require.NoError(t, err)
	assert.Len(t, tasks, 0)
}

func TestMarkSeen_UpdatedTaskBecomesUnseen(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s)
	require.NoError(t, s.MarkSeen(ctx, "johanna", []string{task.ID}))

	// Update the task — it should become "unseen" again.
	newWhat := "updated"
	require.NoError(t, s.UpdateTask(ctx, task.ID, &model.TaskUpdate{What: &newWhat}))

	tasks, err := s.ListTasks(ctx, &model.Filter{UnseenBy: "johanna"})
	require.NoError(t, err)
	assert.Len(t, tasks, 1, "updated task should appear as unseen")
}

func TestMarkAllSeen(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.List = "andres" })
	seedTask(t, s, func(task *model.Task) { task.List = "andres" })
	seedTask(t, s, func(task *model.Task) { task.List = "johanna" })

	require.NoError(t, s.MarkAllSeen(ctx, "johanna", "andres"))

	tasks, err := s.ListTasks(ctx, &model.Filter{UnseenBy: "johanna"})
	require.NoError(t, err)
	// Only the task in "johanna" list should be unseen (tasks in "andres" were marked).
	assert.Len(t, tasks, 1, "johanna list task should still be unseen")
}

// ---------------------------------------------------------------------------
// Labels, Lists, Stats
// ---------------------------------------------------------------------------

func TestListLabels(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"financial", "tax"} })
	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"financial"} })

	labels, err := s.ListLabels(ctx, "")
	require.NoError(t, err)
	assert.Len(t, labels, 2)

	// financial should have count 2 (sorted by count DESC).
	assert.Equal(t, "financial", labels[0].Label)
	assert.Equal(t, 2, labels[0].Count)
}

func TestListLabels_FilterByList(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) {
		task.List = "andres"
		task.Labels = []string{"financial"}
	})
	seedTask(t, s, func(task *model.Task) {
		task.List = "johanna"
		task.Labels = []string{"ops"}
	})

	labels, err := s.ListLabels(ctx, "andres")
	require.NoError(t, err)
	assert.Len(t, labels, 1)
	assert.Equal(t, "financial", labels[0].Label)
}

func TestClearLabel(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"today", "financial"} })
	seedTask(t, s, func(task *model.Task) { task.Labels = []string{"today"} })

	cleared, err := s.ClearLabel(ctx, "today", "")
	require.NoError(t, err)
	assert.Equal(t, 2, cleared)

	labels, err := s.ListLabels(ctx, "")
	require.NoError(t, err)
	// Only "financial" should remain.
	assert.Len(t, labels, 1)
	assert.Equal(t, "financial", labels[0].Label)
}

func TestListLists(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	task := seedTask(t, s, func(task *model.Task) { task.List = "andres" })
	seedTask(t, s, func(task *model.Task) { task.List = "andres" })
	seedTask(t, s, func(task *model.Task) { task.List = "johanna" })
	require.NoError(t, s.DoneTask(ctx, task.ID))

	lists, err := s.ListLists(ctx)
	require.NoError(t, err)
	assert.Len(t, lists, 2)

	// Find the andres list.
	for _, l := range lists {
		if l.Name == "andres" {
			assert.Equal(t, 1, l.Open)
			assert.Equal(t, 1, l.Done)
		}
	}
}

func TestStats(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s)
	seedTask(t, s)
	task3 := seedTask(t, s)
	require.NoError(t, s.DoneTask(ctx, task3.ID))

	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalOpen)
	assert.Equal(t, 1, stats.TotalDone)
	assert.Equal(t, 0, stats.TotalKilled)
}

func TestStats_EmptyDB(t *testing.T) {
	s := newTestStore(t)
	stats, err := s.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalOpen)
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func TestExportImport_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedTask(t, s, func(task *model.Task) {
		task.What = "exported task"
		task.Labels = []string{"important"}
	})

	exported, err := s.ExportTasks(ctx, nil)
	require.NoError(t, err)
	require.Len(t, exported, 1)

	// Import into a fresh store.
	s2 := newTestStore(t)
	require.NoError(t, s2.ImportTasks(ctx, exported))

	tasks, err := s2.ListTasks(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "exported task", tasks[0].What)
	assert.Equal(t, []string{"important"}, tasks[0].Labels)
}
