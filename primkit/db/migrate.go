package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

// Migrate applies all unapplied SQL migrations from an embedded filesystem.
//
// Migrations are SQL files under the given directory, named with a numeric prefix
// for ordering (e.g., "001_initial.sql", "002_add_index.sql"). They are applied
// in lexicographic order, which means the numeric prefix controls the sequence.
//
// Applied migrations are tracked in a _migrations table. Running Migrate multiple
// times with the same set of files is safe — already-applied migrations are skipped.
//
// Each migration runs in its own transaction: if a migration fails, it rolls back
// without affecting previously applied migrations.
//
// Usage:
//
//	//go:embed migrations/*.sql
//	var migrations embed.FS
//
//	err := db.Migrate(database, migrations, "migrations")
func Migrate(database *sql.DB, fs embed.FS, dir string) error {
	// Ensure the tracking table exists. This is idempotent.
	if _, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("creating _migrations table: %w", err)
	}

	// Read all SQL files from the embedded filesystem.
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %q: %w", dir, err)
	}

	// Collect and sort migration files. The numeric prefix (e.g., "001_") ensures
	// they're applied in the correct order.
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	// Apply each migration that hasn't been applied yet.
	for _, name := range names {
		applied, err := isMigrationApplied(database, name)
		if err != nil {
			return fmt.Errorf("checking migration %q: %w", name, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := fs.ReadFile(dir + "/" + name)
		if err != nil {
			return fmt.Errorf("reading migration %q: %w", name, err)
		}

		if err := applyMigration(database, name, string(sqlBytes)); err != nil {
			return fmt.Errorf("applying migration %q: %w", name, err)
		}
	}

	return nil
}

// isMigrationApplied checks whether a migration has already been recorded
// in the _migrations table.
func isMigrationApplied(database *sql.DB, version string) (bool, error) {
	var count int
	err := database.QueryRow(
		"SELECT COUNT(*) FROM _migrations WHERE version = ?", version,
	).Scan(&count)
	return count > 0, err
}

// applyMigration runs a single migration's SQL within a transaction, then records
// it in the _migrations table. If the SQL fails, the transaction rolls back and
// the migration is not marked as applied.
func applyMigration(database *sql.DB, version, sqlContent string) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() // No-op if committed.

	if _, err := tx.Exec(sqlContent); err != nil {
		return fmt.Errorf("executing SQL: %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO _migrations (version) VALUES (?)", version,
	); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return tx.Commit()
}
