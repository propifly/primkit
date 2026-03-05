// Package cli implements the stateprim command-line interface using cobra.
// Each subcommand (set, get, query, etc.) lives in its own file and follows
// a consistent pattern: parse flags → call store → format output.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/propifly/primkit/primkit/config"
	"github.com/propifly/primkit/primkit/replicate"
	"github.com/propifly/primkit/stateprim/internal/store"
	"github.com/spf13/cobra"
)

// contextKey is unexported to prevent collisions with other packages.
type contextKey string

const storeKey contextKey = "store"
const configKey contextKey = "config"
const dbPathKey contextKey = "dbpath"
const replicatorKey contextKey = "replicator"

// storeFromContext extracts the Store from the command's context.
// Panics if the store is not set — this should never happen because
// PersistentPreRunE sets it before any command runs.
func storeFromContext(ctx context.Context) store.Store {
	s, ok := ctx.Value(storeKey).(store.Store)
	if !ok {
		panic("store not found in context — PersistentPreRunE did not run")
	}
	return s
}

// NewRootCmd creates the top-level stateprim command with all subcommands.
//
// Global flags:
//   - --db: path to SQLite database (overrides config and env var)
//   - --config: path to config file
//   - --format: output format (table, json, quiet)
func NewRootCmd() *cobra.Command {
	var (
		dbPath     string
		configPath string
		format     string
	)

	root := &cobra.Command{
		Use:   "stateprim",
		Short: "Operational state persistence for agents",
		Long: `stateprim is a state persistence primitive for AI agents and automation.

Three access patterns share the same underlying model:
  - Key-value state: set/get/update for current state
  - Dedup lookups: has/set-if-new for existence checks
  - Append records: immutable, timestamped log entries

Three interfaces are available: CLI (default), HTTP server (serve), and
MCP server (mcp). All share the same SQLite database.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database")
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	root.PersistentFlags().StringVarP(&format, "format", "f", "table", "output format: table, json, quiet")

	// PersistentPreRunE runs before every subcommand. It opens the database
	// and injects the store into the command's context. This is lazy — if
	// you run "stateprim --help", the database is never opened.
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if dbPath == "" {
			dbPath = os.Getenv("STATEPRIM_DB")
		}

		// Load config for replication settings and DB path fallback.
		cfg, err := config.LoadWithEnvOverrides(configPath, "STATEPRIM")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if dbPath == "" {
			dbPath = cfg.Storage.DB
		}
		if dbPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("determining home directory: %w", err)
			}
			dbPath = filepath.Join(home, ".stateprim", "default.db")
		}

		// If replication is enabled and the local DB doesn't exist, restore
		// from the replica before opening.
		var rep *replicate.Replicator
		if cfg.Storage.Replicate.Enabled {
			logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
			rep = replicate.New(cfg.Storage.Replicate, logger)
			if err := rep.RestoreIfNeeded(cmd.Context(), dbPath); err != nil {
				return fmt.Errorf("restoring from replica: %w", err)
			}
		}

		s, err := store.New(dbPath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}

		// Start replication after the DB is open. This runs for every command
		// so that all writes are replicated regardless of interface.
		if rep != nil {
			if err := rep.Start(dbPath); err != nil {
				return fmt.Errorf("starting replication: %w", err)
			}
		}

		ctx := context.WithValue(cmd.Context(), storeKey, store.Store(s))
		ctx = context.WithValue(ctx, configKey, cfg)
		ctx = context.WithValue(ctx, dbPathKey, dbPath)
		ctx = context.WithValue(ctx, replicatorKey, rep)
		cmd.SetContext(ctx)

		return nil
	}

	// PersistentPostRunE runs after every subcommand. It gracefully stops
	// Litestream replication with a final sync.
	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		rep, _ := cmd.Context().Value(replicatorKey).(*replicate.Replicator)
		if rep != nil {
			if err := rep.Stop(); err != nil {
				return fmt.Errorf("stopping replication: %w", err)
			}
		}
		return nil
	}

	// Register all subcommands.
	root.AddCommand(
		newSetCmd(),
		newGetCmd(),
		newHasCmd(),
		newSetIfNewCmd(),
		newAppendCmd(),
		newDeleteCmd(),
		newQueryCmd(),
		newPurgeCmd(),
		newNamespacesCmd(),
		newStatsCmd(),
		newExportCmd(),
		newImportCmd(),
		newServeCmd(),
		newMCPCmd(),
		newRestoreCmd(),
	)

	return root
}

// configFromContext extracts the Config from the command's context.
func configFromContext(ctx context.Context) *config.Config {
	cfg, _ := ctx.Value(configKey).(*config.Config)
	return cfg
}

// dbPathFromContext extracts the database path from the command's context.
func dbPathFromContext(ctx context.Context) string {
	p, _ := ctx.Value(dbPathKey).(string)
	return p
}
