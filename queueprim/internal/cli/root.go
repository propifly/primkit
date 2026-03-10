// Package cli implements the queueprim command-line interface using cobra.
// Each subcommand lives in its own file and follows the pattern:
// parse flags → call store → format output.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/propifly/primkit/primkit/config"
	"github.com/propifly/primkit/primkit/replicate"
	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/spf13/cobra"
)

type contextKey string

const (
	storeKey      contextKey = "store"
	configKey     contextKey = "config"
	dbPathKey     contextKey = "dbpath"
	replicatorKey contextKey = "replicator"
)

func storeFromContext(ctx context.Context) store.Store {
	s, ok := ctx.Value(storeKey).(store.Store)
	if !ok {
		panic("store not found in context — PersistentPreRunE did not run")
	}
	return s
}

func configFromContext(ctx context.Context) *config.Config {
	cfg, _ := ctx.Value(configKey).(*config.Config)
	return cfg
}

func dbPathFromContext(ctx context.Context) string {
	p, _ := ctx.Value(dbPathKey).(string)
	return p
}

// NewRootCmd creates the top-level queueprim command with all subcommands.
func NewRootCmd() *cobra.Command {
	var (
		dbPath     string
		configPath string
		format     string
	)

	root := &cobra.Command{
		Use:   "queueprim",
		Short: "Agent-native work queue primitive",
		Long: `queueprim is a work queue primitive for AI agents.

Jobs have an explicit lifecycle (pending → claimed → done | failed | dead),
belong to named queues, and support atomic dequeue with visibility timeouts
for automatic crash recovery.

Three interfaces: CLI (default), HTTP server (serve), MCP server (mcp).
All share the same SQLite database.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database")
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	root.PersistentFlags().StringVarP(&format, "format", "f", "table", "output format: table, json, quiet")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if dbPath == "" {
			dbPath = os.Getenv("QUEUEPRIM_DB")
		}

		cfg, err := config.LoadWithEnvOverrides(configPath, "QUEUEPRIM")
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
			dbPath = filepath.Join(home, ".queueprim", "default.db")
		}

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

	root.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		rep, _ := cmd.Context().Value(replicatorKey).(*replicate.Replicator)
		if rep != nil {
			if err := rep.Stop(); err != nil {
				return fmt.Errorf("stopping replication: %w", err)
			}
		}
		return nil
	}

	root.AddCommand(
		newVersionCmd(),
		newEnqueueCmd(),
		newDequeueCmd(),
		newCompleteCmd(),
		newFailCmd(),
		newReleaseCmd(),
		newExtendCmd(),
		newPeekCmd(),
		newListCmd(),
		newGetCmd(),
		newQueuesCmd(),
		newStatsCmd(),
		newPurgeCmd(),
		newExportCmd(),
		newImportCmd(),
		newServeCmd(),
		newMCPCmd(),
		newRestoreCmd(),
	)

	return root
}
