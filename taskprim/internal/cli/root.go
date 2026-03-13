// Package cli implements the taskprim command-line interface using cobra.
// Each subcommand (add, list, done, etc.) lives in its own file and follows
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
	"github.com/propifly/primkit/taskprim/internal/store"
	"github.com/spf13/cobra"
)

// contextKey is unexported to prevent collisions with other packages.
type contextKey string

const (
	storeKey      contextKey = "store"
	configKey     contextKey = "config"
	dbPathKey     contextKey = "dbpath"
	replicatorKey contextKey = "replicator"
)

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

// NewRootCmd creates the top-level taskprim command with all subcommands.
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
		Use:   "taskprim",
		Short: "Agent-native task management",
		Long: `taskprim is a task management primitive for AI agents and humans.

Tasks have an explicit lifecycle (open → done | killed), belong to lists,
carry freeform labels, and support per-agent seen-tracking.

Three interfaces are available: CLI (default), HTTP server (serve), and
MCP server (mcp). All share the same SQLite database.`,
		SilenceUsage:  true, // Don't print usage on errors — it clutters the output.
		SilenceErrors: true, // We handle error printing ourselves in main().
		Version:       resolveVersion(),
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database")
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	root.PersistentFlags().StringVarP(&format, "format", "f", "table", "output format: table, json, quiet")

	// PersistentPreRunE runs before every subcommand. It opens the database
	// and injects the store into the command's context. This is lazy — if
	// you run "taskprim --help", the database is never opened.
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Only consult the env var when no config file was specified.
		// When --config is given the file is authoritative; the env var must
		// not stomp per-agent configuration in multi-agent deployments.
		if dbPath == "" && configPath == "" {
			dbPath = os.Getenv("TASKPRIM_DB")
		}

		// Load config for replication settings and DB path fallback.
		// When configPath != "" env overrides are skipped inside this call.
		cfg, err := config.LoadWithEnvOverrides(configPath, "TASKPRIM")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Resolve database path:
		//   With --config:    --db flag → storage.db from config file → default
		//   Without --config: --db flag → TASKPRIM_DB env var → default
		if dbPath == "" {
			dbPath = cfg.Storage.DB
		}
		if dbPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("determining home directory: %w", err)
			}
			dbPath = filepath.Join(home, ".taskprim", "default.db")
		}

		// If replication is enabled and the local DB doesn't exist, restore
		// from the replica before opening. This lets CLI commands and serve
		// mode seamlessly recover from a previous deployment's backup.
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
		// (CLI, serve, mcp) so that all writes are replicated regardless of
		// which interface the agent uses.
		if rep != nil {
			if err := rep.Start(dbPath); err != nil {
				return fmt.Errorf("starting replication: %w", err)
			}
		}

		// Inject store, config, DB path, and replicator into context.
		ctx := context.WithValue(cmd.Context(), storeKey, store.Store(s))
		ctx = context.WithValue(ctx, configKey, cfg)
		ctx = context.WithValue(ctx, dbPathKey, dbPath)
		ctx = context.WithValue(ctx, replicatorKey, rep)
		cmd.SetContext(ctx)

		return nil
	}

	// PersistentPostRunE runs after every subcommand. It gracefully stops
	// Litestream replication with a final sync so that even short-lived CLI
	// commands push their WAL changes to the remote replica.
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
		newVersionCmd(),
		newAddCmd(),
		newListCmd(),
		newGetCmd(),
		newDoneCmd(),
		newKillCmd(),
		newEditCmd(),
		newSeenCmd(),
		newLabelsCmd(),
		newListsCmd(),
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

// defaultList returns the default list name from the TASKPRIM_LIST env var.
// If not set, returns "default".
func defaultList() string {
	if list := os.Getenv("TASKPRIM_LIST"); list != "" {
		return list
	}
	return "default"
}
