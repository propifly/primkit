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
	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/propifly/primkit/stateprim/internal/store"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore creates a fresh in-memory store for each test.
func newTestStore(t *testing.T) store.Store {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

// execCmd creates a root command, injects the test store, and runs the given args.
// Returns captured stdout.
func execCmd(t *testing.T, s store.Store, args ...string) string {
	t.Helper()
	root := NewRootCmd()
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := context.WithValue(cmd.Context(), storeKey, s)
		cmd.SetContext(ctx)
		return nil
	}

	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	require.NoError(t, err)
	return buf.String()
}

// execCmdErr is like execCmd but expects an error.
func execCmdErr(t *testing.T, s store.Store, args ...string) string {
	t.Helper()
	root := NewRootCmd()
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := context.WithValue(cmd.Context(), storeKey, s)
		cmd.SetContext(ctx)
		return nil
	}

	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	require.Error(t, err)
	return err.Error()
}

// seedRecord inserts a record into the store directly.
func seedRecord(t *testing.T, s store.Store, namespace, key, value string) {
	t.Helper()
	r := &model.Record{
		Namespace: namespace,
		Key:       key,
		Value:     json.RawMessage(value),
	}
	require.NoError(t, s.Set(context.Background(), r))
}

// ---------------------------------------------------------------------------
// set
// ---------------------------------------------------------------------------

func TestSet_Basic(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "set", "config", "theme", `"dark"`)
	assert.Contains(t, out, "config")
	assert.Contains(t, out, "theme")
}

func TestSet_JSON(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "set", "config", "theme", `"dark"`, "--format", "json")
	var r model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &r))
	assert.Equal(t, "config", r.Namespace)
	assert.Equal(t, "theme", r.Key)
	assert.Equal(t, `"dark"`, string(r.Value))
}

func TestSet_InvalidJSON(t *testing.T) {
	s := newTestStore(t)
	errMsg := execCmdErr(t, s, "set", "ns", "key", "not json")
	assert.Contains(t, errMsg, "valid JSON")
}

// ---------------------------------------------------------------------------
// get
// ---------------------------------------------------------------------------

func TestGet_Existing(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "config", "theme", `"dark"`)

	out := execCmd(t, s, "get", "config", "theme")
	assert.Contains(t, out, `"dark"`)
}

func TestGet_JSON(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "config", "theme", `"dark"`)

	out := execCmd(t, s, "get", "config", "theme", "--format", "json")
	var r model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &r))
	assert.Equal(t, `"dark"`, string(r.Value))
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	errMsg := execCmdErr(t, s, "get", "ns", "missing")
	assert.Contains(t, errMsg, "not found")
}

// ---------------------------------------------------------------------------
// has
// ---------------------------------------------------------------------------

func TestHas_Exists(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "key", `{}`)

	out := execCmd(t, s, "has", "ns", "key")
	assert.Contains(t, out, "yes")
}

func TestHas_NotExists(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "has", "ns", "missing")
	assert.Contains(t, out, "no")
}

func TestHas_JSON(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "key", `{}`)

	out := execCmd(t, s, "has", "ns", "key", "--format", "json")
	var result map[string]bool
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result["exists"])
}

// ---------------------------------------------------------------------------
// set-if-new
// ---------------------------------------------------------------------------

func TestSetIfNew_New(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "set-if-new", "dedup", "msg:abc", `{"sent":true}`)
	assert.Contains(t, out, "dedup")
	assert.Contains(t, out, "msg:abc")
}

func TestSetIfNew_AlreadyExists(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "dedup", "msg:abc", `{"sent":true}`)

	errMsg := execCmdErr(t, s, "set-if-new", "dedup", "msg:abc", `{"sent":false}`)
	assert.Contains(t, errMsg, "already exists")
}

// ---------------------------------------------------------------------------
// append
// ---------------------------------------------------------------------------

func TestAppend_Basic(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "append", "audit", `{"action":"login"}`)
	assert.Contains(t, out, "audit")
	assert.Contains(t, out, "a_") // auto-generated key prefix
}

func TestAppend_JSON(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "append", "audit", `{"action":"login"}`, "--format", "json")
	var r model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &r))
	assert.Equal(t, "audit", r.Namespace)
	assert.True(t, r.Immutable)
	assert.Contains(t, r.Key, "a_")
}

func TestAppend_InvalidJSON(t *testing.T) {
	s := newTestStore(t)
	errMsg := execCmdErr(t, s, "append", "audit", "not json")
	assert.Contains(t, errMsg, "valid JSON")
}

// ---------------------------------------------------------------------------
// delete
// ---------------------------------------------------------------------------

func TestDelete_Existing(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "key", `{}`)

	out := execCmd(t, s, "delete", "ns", "key")
	assert.Contains(t, out, "Deleted")
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStore(t)
	errMsg := execCmdErr(t, s, "delete", "ns", "missing")
	assert.Contains(t, errMsg, "not found")
}

// ---------------------------------------------------------------------------
// query
// ---------------------------------------------------------------------------

func TestQuery_AllInNamespace(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "k1", `{"a":1}`)
	seedRecord(t, s, "ns", "k2", `{"b":2}`)
	seedRecord(t, s, "other", "k3", `{"c":3}`)

	out := execCmd(t, s, "query", "ns")
	assert.Contains(t, out, "k1")
	assert.Contains(t, out, "k2")
	assert.NotContains(t, out, "k3")
}

func TestQuery_WithPrefix(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "cfg:theme", `"dark"`)
	seedRecord(t, s, "ns", "cfg:lang", `"en"`)
	seedRecord(t, s, "ns", "user:name", `"alice"`)

	out := execCmd(t, s, "query", "ns", "--prefix", "cfg:")
	assert.Contains(t, out, "cfg:theme")
	assert.Contains(t, out, "cfg:lang")
	assert.NotContains(t, out, "user:name")
}

func TestQuery_CountOnly(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "k1", `{}`)
	seedRecord(t, s, "ns", "k2", `{}`)

	out := execCmd(t, s, "query", "ns", "--count")
	assert.Equal(t, "2\n", out)
}

func TestQuery_CountOnly_JSON(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "k1", `{}`)

	out := execCmd(t, s, "query", "ns", "--count", "--format", "json")
	var result map[string]int
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 1, result["count"])
}

func TestQuery_Empty(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "query", "empty")
	assert.Contains(t, out, "No records found")
}

// ---------------------------------------------------------------------------
// purge
// ---------------------------------------------------------------------------

func TestPurge_Basic(t *testing.T) {
	s := newTestStore(t)
	// Seed records — all very recent, so purge with "1s" won't delete them.
	seedRecord(t, s, "ns", "k1", `{}`)

	out := execCmd(t, s, "purge", "ns", "24h")
	assert.Contains(t, out, "Purged 0 record(s)")
}

func TestPurge_JSON(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "purge", "ns", "24h", "--format", "json")
	var result map[string]int
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, 0, result["deleted"])
}

// ---------------------------------------------------------------------------
// namespaces
// ---------------------------------------------------------------------------

func TestNamespaces_Basic(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "alpha", "k1", `{}`)
	seedRecord(t, s, "alpha", "k2", `{}`)
	seedRecord(t, s, "beta", "k1", `{}`)

	out := execCmd(t, s, "namespaces")
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
}

func TestNamespaces_JSON(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns1", "k1", `{}`)

	out := execCmd(t, s, "namespaces", "--format", "json")
	var nss []model.NamespaceInfo
	require.NoError(t, json.Unmarshal([]byte(out), &nss))
	assert.Len(t, nss, 1)
	assert.Equal(t, "ns1", nss[0].Namespace)
}

func TestNamespaces_Empty(t *testing.T) {
	s := newTestStore(t)
	out := execCmd(t, s, "namespaces")
	assert.Contains(t, out, "No namespaces found")
}

// ---------------------------------------------------------------------------
// stats
// ---------------------------------------------------------------------------

func TestStats_Basic(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns1", "k1", `{}`)
	seedRecord(t, s, "ns2", "k2", `{}`)

	out := execCmd(t, s, "stats")
	assert.Contains(t, out, "Records:    2")
	assert.Contains(t, out, "Namespaces: 2")
}

func TestStats_JSON(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "k1", `{}`)

	out := execCmd(t, s, "stats", "--format", "json")
	var stats model.Stats
	require.NoError(t, json.Unmarshal([]byte(out), &stats))
	assert.Equal(t, 1, stats.TotalRecords)
	assert.Equal(t, 1, stats.TotalNamespaces)
}

// ---------------------------------------------------------------------------
// export / import
// ---------------------------------------------------------------------------

func TestExport_All(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns1", "k1", `{"a":1}`)
	seedRecord(t, s, "ns2", "k2", `{"b":2}`)

	out := execCmd(t, s, "export")
	var records []*model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	assert.Len(t, records, 2)
}

func TestExport_FilteredNamespace(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns1", "k1", `{}`)
	seedRecord(t, s, "ns2", "k2", `{}`)

	out := execCmd(t, s, "export", "--namespace", "ns1")
	var records []*model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	assert.Len(t, records, 1)
	assert.Equal(t, "ns1", records[0].Namespace)
}

func TestImport_FromFile(t *testing.T) {
	s := newTestStore(t)

	// Write a JSON file to import.
	data := `[{"namespace":"imported","key":"k1","value":{"x":1},"immutable":false,"created_at":"2026-03-03T10:00:00Z","updated_at":"2026-03-03T10:00:00Z"}]`
	tmpFile, err := os.CreateTemp("", "stateprim-import-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(data)
	require.NoError(t, err)
	tmpFile.Close()

	out := execCmd(t, s, "import", "--file", tmpFile.Name())
	assert.Contains(t, out, "Imported 1 record(s)")

	// Verify the record is there.
	verifyOut := execCmd(t, s, "get", "imported", "k1", "--format", "json")
	assert.Contains(t, verifyOut, `"imported"`)
}

// ---------------------------------------------------------------------------
// format helpers
// ---------------------------------------------------------------------------

func TestOutputRecord_Quiet(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "key", `{}`)

	out := execCmd(t, s, "get", "ns", "key", "--format", "quiet")
	assert.Equal(t, "ns/key\n", out)
}

func TestOutputRecords_Quiet(t *testing.T) {
	s := newTestStore(t)
	seedRecord(t, s, "ns", "k1", `{}`)
	seedRecord(t, s, "ns", "k2", `{}`)

	out := execCmd(t, s, "query", "ns", "--format", "quiet")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Len(t, lines, 2)
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "1234567...", truncate("1234567890123", 10))
}

// ---------------------------------------------------------------------------
// Integration lifecycle
// ---------------------------------------------------------------------------

func TestLifecycle_SetGetDeleteHas(t *testing.T) {
	s := newTestStore(t)

	// Set.
	execCmd(t, s, "set", "config", "theme", `"light"`)

	// Get.
	out := execCmd(t, s, "get", "config", "theme")
	assert.Contains(t, out, `"light"`)

	// Has.
	out = execCmd(t, s, "has", "config", "theme")
	assert.Contains(t, out, "yes")

	// Update.
	execCmd(t, s, "set", "config", "theme", `"dark"`)
	out = execCmd(t, s, "get", "config", "theme")
	assert.Contains(t, out, `"dark"`)

	// Delete.
	execCmd(t, s, "delete", "config", "theme")

	// Has after delete.
	out = execCmd(t, s, "has", "config", "theme")
	assert.Contains(t, out, "no")
}

func TestLifecycle_AppendQueryStats(t *testing.T) {
	s := newTestStore(t)

	// Append multiple.
	execCmd(t, s, "append", "audit", `{"action":"login"}`)
	execCmd(t, s, "append", "audit", `{"action":"update"}`)
	execCmd(t, s, "append", "audit", `{"action":"logout"}`)

	// Query.
	out := execCmd(t, s, "query", "audit", "--format", "json")
	var records []*model.Record
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	assert.Len(t, records, 3)

	// Count.
	countOut := execCmd(t, s, "query", "audit", "--count")
	assert.Equal(t, "3\n", countOut)

	// Stats.
	statsOut := execCmd(t, s, "stats")
	assert.Contains(t, statsOut, "Records:    3")
	assert.Contains(t, statsOut, "Namespaces: 1")

	// Namespaces.
	nsOut := execCmd(t, s, "namespaces")
	assert.Contains(t, nsOut, "audit")
}

// ---------------------------------------------------------------------------
// Regression tests
// ---------------------------------------------------------------------------

// TestDBPathFromConfig verifies that storage.db in the YAML config file is
// used as the database path when --db and STATEPRIM_DB are both unset.
//
// Regression for: config was loaded after the DB path fallback chain, so
// cfg.Storage.DB was never consulted and the hardcoded default always won.
func TestDBPathFromConfig(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "stateprim-test.db")
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("storage:\n  db: " + dbPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configPath, "stats"})
	err := root.Execute()
	require.NoError(t, err, "command should succeed using DB path from config")

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "DB file should be created at path from config, not the default")
}
