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
	"strings"

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
	if err := validateDBPath(path); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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

// validateDBPath guards against paths that carry embedded environment-variable
// content — a class of misconfiguration where shell KEY=VALUE assignments (or
// mis-parsed .env lines) end up concatenated into the database file path.
//
// Two checks are enforced:
//
//  1. Null bytes: forbidden on all major filesystems; their presence in a path
//     indicates catastrophic config corruption and would create unusable files.
//
//  2. '=' in the filename (base) component: an equals sign should never appear
//     in a database filename. When it does it almost always means an env var
//     value leaked in — e.g. KNOWLEDGEPRIM_DB was set to
//     "/path/.knowledge.dbX_ACCESS_TOKEN=abc" because a .env parser concatenated
//     two lines, or a shell script omitted quotes around a variable expansion.
//     Catching this early produces a clear error instead of silently creating a
//     file whose name contains a secret token.
//
// Directory components are not checked for '=' because some legitimate paths
// on certain systems use it (e.g. Nix store paths). Only the final filename
// is validated.
func validateDBPath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("invalid database path: contains a null byte — check your *PRIM_DB env var or storage.db config value")
	}
	if base := filepath.Base(path); strings.Contains(base, "=") {
		return fmt.Errorf(
			"invalid database path %q: filename contains '=' which suggests an env var value leaked into the path\n"+
				"  hint: check that KNOWLEDGEPRIM_DB / TASKPRIM_DB / STATEPRIM_DB / QUEUEPRIM_DB\n"+
				"        or storage.db in your config file contains only the file path, not extra key=value content",
			path,
		)
	}
	return nil
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
