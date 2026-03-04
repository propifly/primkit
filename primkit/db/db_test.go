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
