package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "deep", "test.db")

	db, err := Open(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// The file and its parent directories should exist.
	_, err = os.Stat(dbPath)
	assert.NoError(t, err, "database file should be created")
}

func TestOpen_WALMode(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "wal.db"))
	require.NoError(t, err)
	defer db.Close()

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "journal_mode should be WAL")
}

func TestOpen_ForeignKeys(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "fk.db"))
	require.NoError(t, err)
	defer db.Close()

	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	require.NoError(t, err)
	assert.Equal(t, 1, fk, "foreign_keys should be enabled")
}

func TestOpen_BusyTimeout(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "timeout.db"))
	require.NoError(t, err)
	defer db.Close()

	var timeout int
	err = db.QueryRow("PRAGMA busy_timeout").Scan(&timeout)
	require.NoError(t, err)
	assert.Equal(t, 5000, timeout, "busy_timeout should be 5000ms")
}

func TestOpenInMemory(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	// Verify we can execute SQL.
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	assert.NoError(t, err, "should be able to create tables in in-memory database")
}

// TestValidateDBPath covers the env-var-leakage guard added to Open().
func TestValidateDBPath(t *testing.T) {
	t.Run("valid paths pass", func(t *testing.T) {
		valid := []string{
			"/tmp/test.db",
			"/home/user/.taskprim/default.db",
			"/data/workspace-johanna/.knowledge.db",
			"relative/path.db",
			"db.db",
		}
		for _, p := range valid {
			assert.NoError(t, validateDBPath(p), "valid path %q should pass", p)
		}
	})

	t.Run("null byte rejected", func(t *testing.T) {
		err := validateDBPath("/tmp/test\x00.db")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "null byte")
	})

	t.Run("equals in filename rejected — env var leakage pattern", func(t *testing.T) {
		// Reproduces: .openclaw/workspace-johanna/.knowledge.dbX_ACCESS_TOKEN=abc
		// caused by an env parser concatenating two .env lines without a separator.
		err := validateDBPath("/data/workspace-johanna/.knowledge.dbX_ACCESS_TOKEN=abc123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "'='")
	})

	t.Run("equals in filename rejected — simple case", func(t *testing.T) {
		err := validateDBPath("/tmp/db=leaked.db")
		require.Error(t, err)
	})

	t.Run("equals in directory component allowed", func(t *testing.T) {
		// Nix store paths and some CI systems use = in directory names.
		// Only the base filename is checked, not parent directories.
		err := validateDBPath("/nix/store/abc=def/bin/test.db")
		assert.NoError(t, err)
	})

	t.Run("error message includes actionable hint", func(t *testing.T) {
		err := validateDBPath("/tmp/.knowledge.dbSECRET_TOKEN=hunter2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hint:")
		assert.Contains(t, err.Error(), "KNOWLEDGEPRIM_DB")
	})

	t.Run("Open rejects path with equals in filename", func(t *testing.T) {
		_, err := Open("/tmp/.knowledge.dbTOKEN=leaked")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "'='")
	})
}
