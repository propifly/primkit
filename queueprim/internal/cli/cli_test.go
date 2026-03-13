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
	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestStore(t *testing.T) store.Store {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

// execCmd runs a queueprim command against an in-memory store and returns
// captured stdout and any error.
func execCmd(t *testing.T, s store.Store, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	// Override PersistentPreRunE to inject our test store.
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := context.WithValue(cmd.Context(), storeKey, s)
		cmd.SetContext(ctx)
		return nil
	}

	err := root.Execute()
	return buf.String(), err
}

// seedJob enqueues a job and returns it. Convenience for tests that need
// pre-existing data.
func seedJob(t *testing.T, s store.Store, queue string, payload interface{}, opts ...func(*model.Job)) *model.Job {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	job := &model.Job{
		Queue:   queue,
		Payload: raw,
	}
	for _, opt := range opts {
		opt(job)
	}
	require.NoError(t, s.EnqueueJob(context.Background(), job))
	return job
}

// claimJob dequeues a job and returns it.
func claimJob(t *testing.T, s store.Store, queue string) *model.Job {
	t.Helper()
	job, err := s.DequeueJob(context.Background(), queue, "test-worker", "", 0)
	require.NoError(t, err)
	return job
}

// ---------------------------------------------------------------------------
// enqueue command
// ---------------------------------------------------------------------------

func TestEnqueue_Basic(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "enqueue", "infra/fixes", `{"summary":"test"}`)
	require.NoError(t, err)
	assert.Contains(t, out, "q_")
	assert.Contains(t, out, "infra/fixes")
}

func TestEnqueue_WithFlags(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "enqueue", "infra/fixes", `{"detail":"test"}`,
		"--type", "ssh_fail",
		"--priority", "high",
		"--max-retries", "2",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "infra/fixes")

	// Verify stored values.
	jobs, _ := s.ListJobs(context.Background(), &model.Filter{Queue: "infra/fixes"})
	require.Len(t, jobs, 1)
	assert.Equal(t, "ssh_fail", jobs[0].Type)
	assert.Equal(t, model.PriorityHigh, jobs[0].Priority)
	assert.Equal(t, 2, jobs[0].MaxRetries)
}

func TestEnqueue_JSON(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "enqueue", "q", `{"k":"v"}`, "--format", "json")
	require.NoError(t, err)

	var job model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &job))
	assert.True(t, strings.HasPrefix(job.ID, "q_"))
	assert.Equal(t, model.StatusPending, job.Status)
}

func TestEnqueue_Quiet(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "enqueue", "q", `{}`, "--format", "quiet")
	require.NoError(t, err)
	id := strings.TrimSpace(out)
	assert.True(t, strings.HasPrefix(id, "q_"), "expected job ID, got %q", id)
}

func TestEnqueue_InvalidJSON(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "enqueue", "q", "not-json")
	require.Error(t, err)
}

func TestEnqueue_MissingArgs(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "enqueue", "q")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// complete command
// ---------------------------------------------------------------------------

func TestComplete_Basic(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", map[string]string{"k": "v"})
	claimJob(t, s, "q")

	out, err := execCmd(t, s, "complete", job.ID)
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)
	assert.Contains(t, out, "done")
}

func TestComplete_WithOutput(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)
	claimJob(t, s, "q")

	out, err := execCmd(t, s, "complete", job.ID, "--output", `{"result":"ok"}`, "--format", "json")
	require.NoError(t, err)

	var got model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, model.StatusDone, got.Status)
	assert.NotNil(t, got.Output)
}

func TestComplete_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "complete", "q_notexist")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// fail command
// ---------------------------------------------------------------------------

func TestFail_ToDead(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)
	claimJob(t, s, "q")

	out, err := execCmd(t, s, "fail", job.ID, "--reason", "broke", "--dead")
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)

	got, _ := s.GetJob(context.Background(), job.ID)
	assert.Equal(t, model.StatusDead, got.Status)
}

func TestFail_Retryable(t *testing.T) {
	s := newTestStore(t)
	job := &model.Job{Queue: "q", Payload: []byte(`{}`), MaxRetries: 2}
	require.NoError(t, s.EnqueueJob(context.Background(), job))
	claimJob(t, s, "q")

	out, err := execCmd(t, s, "fail", job.ID, "--reason", "transient")
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)

	got, _ := s.GetJob(context.Background(), job.ID)
	assert.Equal(t, model.StatusPending, got.Status)
}

// ---------------------------------------------------------------------------
// release command
// ---------------------------------------------------------------------------

func TestRelease_Basic(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)
	claimJob(t, s, "q")

	out, err := execCmd(t, s, "release", job.ID)
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)

	got, _ := s.GetJob(context.Background(), job.ID)
	assert.Equal(t, model.StatusPending, got.Status)
}

// ---------------------------------------------------------------------------
// extend command
// ---------------------------------------------------------------------------

func TestExtend_Basic(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)
	claimed := claimJob(t, s, "q")

	out, err := execCmd(t, s, "extend", job.ID, "--by", "45m")
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)

	got, _ := s.GetJob(context.Background(), job.ID)
	assert.True(t, got.VisibleAfter.After(claimed.VisibleAfter))
}

// ---------------------------------------------------------------------------
// peek command
// ---------------------------------------------------------------------------

func TestPeek_ShowsJob(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "infra/fixes", map[string]string{"detail": "test"})

	out, err := execCmd(t, s, "peek", "infra/fixes")
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)

	// Job should still be pending (peek doesn't claim).
	got, _ := s.GetJob(context.Background(), job.ID)
	assert.Equal(t, model.StatusPending, got.Status)
}

func TestPeek_EmptyQueue(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "peek", "empty/queue")
	require.NoError(t, err)
	assert.Contains(t, out, "empty")
}

// ---------------------------------------------------------------------------
// list command
// ---------------------------------------------------------------------------

func TestList_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "No jobs found")
}

func TestList_ShowsJobs(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q1", `{"a":1}`)
	seedJob(t, s, "q2", `{"b":2}`)

	out, err := execCmd(t, s, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "q1")
	assert.Contains(t, out, "q2")
}

func TestList_FilterByQueue(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "infra/fixes", `{}`)
	seedJob(t, s, "review/code", `{}`)

	out, err := execCmd(t, s, "list", "--queue", "infra/fixes")
	require.NoError(t, err)
	assert.Contains(t, out, "infra/fixes")
	assert.NotContains(t, out, "review/code")
}

func TestList_FilterByStatus(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)
	seedJob(t, s, "q", `{}`)
	claimJob(t, s, "q")
	_ = s.CompleteJob(context.Background(), claimJob(t, s, "q").ID, nil)
	_ = job // suppress unused

	out, err := execCmd(t, s, "list", "--status", "done")
	require.NoError(t, err)
	assert.Contains(t, out, "done")
}

func TestList_JSONFormat(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q", map[string]string{"key": "val"})

	out, err := execCmd(t, s, "list", "--format", "json")
	require.NoError(t, err)

	var jobs []*model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &jobs))
	require.Len(t, jobs, 1)
	assert.Equal(t, "q", jobs[0].Queue)
}

// ---------------------------------------------------------------------------
// get command
// ---------------------------------------------------------------------------

func TestGet_ExistingJob(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{}`)

	out, err := execCmd(t, s, "get", job.ID)
	require.NoError(t, err)
	assert.Contains(t, out, job.ID)
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "get", "q_notexist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGet_JSON(t *testing.T) {
	s := newTestStore(t)
	job := seedJob(t, s, "q", `{"data":"test"}`)

	out, err := execCmd(t, s, "get", job.ID, "--format", "json")
	require.NoError(t, err)

	var got model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, job.ID, got.ID)
}

// ---------------------------------------------------------------------------
// queues command
// ---------------------------------------------------------------------------

func TestQueues_ShowsCounts(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "infra/fixes", `{}`)
	seedJob(t, s, "infra/fixes", `{}`)
	seedJob(t, s, "review/code", `{}`)

	out, err := execCmd(t, s, "queues")
	require.NoError(t, err)
	assert.Contains(t, out, "infra/fixes")
	assert.Contains(t, out, "review/code")
}

func TestQueues_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "queues")
	require.NoError(t, err)
	assert.Contains(t, out, "No queues found")
}

func TestQueues_JSON(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "myqueue", `{}`)

	out, err := execCmd(t, s, "queues", "--format", "json")
	require.NoError(t, err)

	var queues []model.QueueInfo
	require.NoError(t, json.Unmarshal([]byte(out), &queues))
	require.Len(t, queues, 1)
	assert.Equal(t, "myqueue", queues[0].Queue)
	assert.Equal(t, 1, queues[0].Pending)
}

// ---------------------------------------------------------------------------
// stats command
// ---------------------------------------------------------------------------

func TestStats_ShowsCounts(t *testing.T) {
	s := newTestStore(t)
	j1 := seedJob(t, s, "q", `{}`)
	seedJob(t, s, "q", `{}`)

	claimedJ1 := claimJob(t, s, "q")
	require.NoError(t, s.CompleteJob(context.Background(), claimedJ1.ID, nil))
	_ = j1

	out, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Pending: 1")
	assert.Contains(t, out, "Done:    1")
}

func TestStats_Empty(t *testing.T) {
	s := newTestStore(t)
	out, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Pending: 0")
}

func TestStats_JSON(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q", `{}`)

	out, err := execCmd(t, s, "stats", "--format", "json")
	require.NoError(t, err)

	var stats model.Stats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.Equal(t, 1, stats.TotalPending)
}

// ---------------------------------------------------------------------------
// purge command
// ---------------------------------------------------------------------------

func TestPurge_RequiresStatus(t *testing.T) {
	s := newTestStore(t)
	_, err := execCmd(t, s, "purge", "q")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestPurge_DoneJobs(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q", `{}`)
	claimed := claimJob(t, s, "q")
	require.NoError(t, s.CompleteJob(context.Background(), claimed.ID, nil))

	out, err := execCmd(t, s, "purge", "q", "--status", "done")
	require.NoError(t, err)
	assert.Contains(t, out, "Deleted 1 job(s)")
}

// ---------------------------------------------------------------------------
// export / import
// ---------------------------------------------------------------------------

func TestExport_AllJobs(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q1", `{"a":1}`)
	seedJob(t, s, "q2", `{"b":2}`)

	out, err := execCmd(t, s, "export")
	require.NoError(t, err)

	var jobs []*model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &jobs))
	assert.Len(t, jobs, 2)
}

func TestExport_FilterByQueue(t *testing.T) {
	s := newTestStore(t)
	seedJob(t, s, "q1", `{}`)
	seedJob(t, s, "q2", `{}`)

	out, err := execCmd(t, s, "export", "--queue", "q1")
	require.NoError(t, err)

	var jobs []*model.Job
	require.NoError(t, json.Unmarshal([]byte(out), &jobs))
	require.Len(t, jobs, 1)
	assert.Equal(t, "q1", jobs[0].Queue)
}

// ---------------------------------------------------------------------------
// helpers tests
// ---------------------------------------------------------------------------

func TestParseDurationFlag_Days(t *testing.T) {
	d, err := parseDurationFlag("7d")
	require.NoError(t, err)
	assert.Equal(t, 7*24*60*60, int(d.Seconds()))
}

func TestParseDurationFlag_Hours(t *testing.T) {
	d, err := parseDurationFlag("24h")
	require.NoError(t, err)
	assert.Equal(t, 24*60*60, int(d.Seconds()))
}

func TestParseDurationFlag_Invalid(t *testing.T) {
	_, err := parseDurationFlag("abc")
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

// ---------------------------------------------------------------------------
// Integration: full lifecycle via CLI
// ---------------------------------------------------------------------------

func TestLifecycle_EnqueueDequeueComplete(t *testing.T) {
	s := newTestStore(t)

	// Enqueue a job.
	out, err := execCmd(t, s, "enqueue", "infra/fixes",
		`{"agent":"clawson","summary":"SSH down"}`,
		"--type", "ssh_fail",
		"--priority", "high",
		"--format", "quiet",
	)
	require.NoError(t, err)
	jobID := strings.TrimSpace(out)
	assert.True(t, strings.HasPrefix(jobID, "q_"))

	// Job should appear in list.
	listOut, err := execCmd(t, s, "list", "--queue", "infra/fixes")
	require.NoError(t, err)
	assert.Contains(t, listOut, jobID)

	// Claim it directly via store.
	claimed := claimJob(t, s, "infra/fixes")
	assert.Equal(t, jobID, claimed.ID)
	assert.Equal(t, model.StatusClaimed, claimed.Status)

	// Complete it via CLI.
	_, err = execCmd(t, s, "complete", jobID, "--output", `{"fix":"applied"}`)
	require.NoError(t, err)

	// Stats should reflect completion.
	statsOut, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, statsOut, "Done:    1")
	assert.Contains(t, statsOut, "Pending: 0")
}

func TestLifecycle_FailAndDead(t *testing.T) {
	s := newTestStore(t)

	_, err := execCmd(t, s, "enqueue", "q", `{}`, "--format", "quiet")
	require.NoError(t, err)

	// Claim and fail → dead (one-shot default).
	claimed := claimJob(t, s, "q")
	_, err = execCmd(t, s, "fail", claimed.ID, "--reason", "escalate", "--dead")
	require.NoError(t, err)

	statsOut, err := execCmd(t, s, "stats")
	require.NoError(t, err)
	assert.Contains(t, statsOut, "Dead:    1")
}

// TestDBPathFromConfig verifies storage.db from the config file is used
// when --db and QUEUEPRIM_DB are both unset.
func TestDBPathFromConfig(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queueprim-test.db")
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("storage:\n  db: " + dbPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0o644))

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configPath, "list"})
	err := root.Execute()
	require.NoError(t, err, "command should succeed using DB path from config")

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "DB file should be created at path from config")
}
