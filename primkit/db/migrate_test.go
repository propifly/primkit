package db

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/migrations/*.sql
var testMigrations embed.FS

func TestMigrate_AppliesSchema(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	err = Migrate(db, testMigrations, "testdata/migrations")
	require.NoError(t, err)

	// Verify the table created by 001_create_test.sql exists.
	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='test_items'",
	).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test_items", name)

	// Verify the column added by 002_add_column.sql exists by inserting a row.
	_, err = db.Exec("INSERT INTO test_items (id, name, priority) VALUES (1, 'test', 5)")
	assert.NoError(t, err, "priority column from migration 002 should exist")
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	// Apply migrations twice — the second run should be a no-op.
	err = Migrate(db, testMigrations, "testdata/migrations")
	require.NoError(t, err)

	err = Migrate(db, testMigrations, "testdata/migrations")
	require.NoError(t, err, "running migrations twice should not error")

	// Verify exactly 2 migrations were recorded.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "should have exactly 2 migration records")
}

func TestMigrate_Order(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	err = Migrate(db, testMigrations, "testdata/migrations")
	require.NoError(t, err)

	// Verify migrations were applied in order by checking the _migrations table.
	rows, err := db.Query("SELECT version FROM _migrations ORDER BY version")
	require.NoError(t, err)
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		require.NoError(t, rows.Scan(&v))
		versions = append(versions, v)
	}

	assert.Equal(t, []string{"001_create_test.sql", "002_add_column.sql"}, versions)
}
