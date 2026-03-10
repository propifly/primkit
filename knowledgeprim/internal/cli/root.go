// Package cli implements the knowledgeprim command-line interface using cobra.
// Each subcommand lives in its own file and follows the pattern:
// parse flags → call store → format output.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/propifly/primkit/knowledgeprim/internal/embed"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/propifly/primkit/primkit/config"
	"github.com/propifly/primkit/primkit/replicate"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// contextKey is unexported to prevent collisions with other packages.
type contextKey string

const storeKey contextKey = "store"
const configKey contextKey = "config"
const dbPathKey contextKey = "dbpath"
const replicatorKey contextKey = "replicator"
const embedderKey contextKey = "embedder"
const embedCfgKey contextKey = "embedcfg"

// storeFromContext extracts the Store from the command's context.
func storeFromContext(ctx context.Context) store.Store {
	s, ok := ctx.Value(storeKey).(store.Store)
	if !ok {
		panic("store not found in context — PersistentPreRunE did not run")
	}
	return s
}

// embedderFromContext extracts the Embedder from the command's context.
// Returns nil if no embedding provider is configured.
func embedderFromContext(ctx context.Context) embed.Embedder {
	e, _ := ctx.Value(embedderKey).(embed.Embedder)
	return e
}

// NewRootCmd creates the top-level knowledgeprim command with all subcommands.
func NewRootCmd() *cobra.Command {
	var (
		dbPath     string
		configPath string
		format     string
	)

	root := &cobra.Command{
		Use:   "knowledgeprim",
		Short: "Agent-native knowledge graph primitive",
		Long: `knowledgeprim is a knowledge graph primitive for AI agents. It stores
entities, weighted edges, and vector embeddings in an embedded SQLite database.

Entities are typed knowledge nodes (articles, thoughts, concepts, patterns).
Edges are weighted, contextualized connections between entities. Hybrid search
combines FTS5 keyword matching with vector similarity via Reciprocal Rank Fusion.

Three interfaces: CLI (default), HTTP server (serve), MCP server (mcp).`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to SQLite database")
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	root.PersistentFlags().StringVarP(&format, "format", "f", "text", "output format: text, json")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Only consult the env var when no config file was specified.
		// When --config is given the file is authoritative; the env var must
		// not stomp per-agent configuration in multi-agent deployments.
		if dbPath == "" && configPath == "" {
			dbPath = os.Getenv("KNOWLEDGEPRIM_DB")
		}

		// Load config for replication settings and DB path fallback.
		// When configPath != "" env overrides are skipped inside this call.
		cfg, err := config.LoadWithEnvOverrides(configPath, "KNOWLEDGEPRIM")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Resolve database path:
		//   With --config:    --db flag → storage.db from config file → default
		//   Without --config: --db flag → KNOWLEDGEPRIM_DB env var → default
		if dbPath == "" {
			dbPath = cfg.Storage.DB
		}
		if dbPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("determining home directory: %w", err)
			}
			dbPath = filepath.Join(home, ".knowledgeprim", "default.db")
		}

		// Replication: restore if needed.
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

		// Initialize embedding provider from config file or env vars (optional).
		var embedder embed.Embedder
		embedCfg := loadEmbedConfig(configPath)
		if embedCfg.Provider != "" {
			embedder, err = embed.New(embedCfg)
			if err != nil {
				return fmt.Errorf("initializing embedding provider: %w", err)
			}
		}

		// Inject into context.
		ctx := context.WithValue(cmd.Context(), storeKey, store.Store(s))
		ctx = context.WithValue(ctx, configKey, cfg)
		ctx = context.WithValue(ctx, dbPathKey, dbPath)
		ctx = context.WithValue(ctx, replicatorKey, rep)
		ctx = context.WithValue(ctx, embedderKey, embedder)
		ctx = context.WithValue(ctx, embedCfgKey, embedCfg)
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

	// Register all subcommands.
	root.AddCommand(
		newVersionCmd(),
		newCaptureCmd(),
		newSearchCmd(),
		newRelatedCmd(),
		newConnectCmd(),
		newStrengthenCmd(),
		newEdgeEditCmd(),
		newDisconnectCmd(),
		newGetCmd(),
		newEditCmd(),
		newDeleteCmd(),
		newDiscoverCmd(),
		newTypesCmd(),
		newRelationshipsCmd(),
		newStatsCmd(),
		newExportCmd(),
		newImportCmd(),
		newReEmbedCmd(),
		newStripVectorsCmd(),
		newServeCmd(),
		newMCPCmd(),
		newRestoreCmd(),
	)

	return root
}

func configFromContext(ctx context.Context) *config.Config {
	cfg, _ := ctx.Value(configKey).(*config.Config)
	return cfg
}

func dbPathFromContext(ctx context.Context) string {
	p, _ := ctx.Value(dbPathKey).(string)
	return p
}

func defaultSource() string {
	if src := os.Getenv("KNOWLEDGEPRIM_SOURCE"); src != "" {
		return src
	}
	return "cli"
}

// extendedConfig is a superset of the shared config, adding knowledgeprim-
// specific sections (embedding, auto_connect). We parse the YAML separately
// because the shared config.Config struct cannot be extended without coupling.
type extendedConfig struct {
	Embedding   embed.Config    `yaml:"embedding"`
	AutoConnect autoConnectConf `yaml:"auto_connect"`
}

type autoConnectConf struct {
	Enabled        *bool   `yaml:"enabled"`
	Threshold      float64 `yaml:"threshold"`
	MaxConnections int     `yaml:"max_connections"`
}

// loadEmbedConfig reads embedding configuration from the YAML config file
// and from environment variables. Env vars override file values.
func loadEmbedConfig(configPath string) embed.Config {
	var ext extendedConfig

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			data = config.InterpolateEnvVars(data)
			_ = yaml.Unmarshal(data, &ext)
		}
	}

	cfg := ext.Embedding

	// Environment variable overrides.
	if v := os.Getenv("KNOWLEDGEPRIM_EMBEDDING_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("KNOWLEDGEPRIM_EMBEDDING_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("KNOWLEDGEPRIM_EMBEDDING_DIMENSIONS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			cfg.Dimensions = d
		}
	}
	if v := os.Getenv("KNOWLEDGEPRIM_EMBEDDING_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("KNOWLEDGEPRIM_EMBEDDING_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}

	return cfg
}
