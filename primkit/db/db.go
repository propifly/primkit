// Package db provides SQLite database management for primkit primitives.
// It handles opening databases with the correct pragmas (WAL mode, foreign keys)
// and running schema migrations from embedded SQL files.
//
// Both taskprim and stateprim use this package. Each primitive embeds its own
// migration SQL files and passes them to Migrate().
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Pure Go SQLite driver — no CGo, easy cross-compilation.
)

// Open creates or opens a SQLite database at the given path with production-ready
// settings. It creates parent directories if they don't exist.
//
// The following pragmas are applied:
//   - journal_mode=WAL: allows concurrent readers while writing, required for
//     Litestream replication and serve mode (multiple HTTP requests).
//   - foreign_keys=ON: enforces referential integrity (e.g., task_labels → tasks).
//   - busy_timeout=5000: waits up to 5 seconds for locks instead of failing
//     immediately, which prevents "database is locked" errors under contention.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database %q: %w", path, err)
	}

	// Verify the connection is alive before applying pragmas.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting %q: %w", p, err)
		}
	}

	return db, nil
}

// OpenInMemory opens an in-memory SQLite database with the same pragmas as Open.
// Useful for tests that need a fresh, isolated database.
func OpenInMemory() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("opening in-memory database: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting %q: %w", p, err)
		}
	}

	return db, nil
}
