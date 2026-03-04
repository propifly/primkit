package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/propifly/primkit/primkit/replicate"
	"github.com/spf13/cobra"
)

func newRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restore database from a replica",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFromContext(cmd.Context())
			dbPath := dbPathFromContext(cmd.Context())
			logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

			if !cfg.Storage.Replicate.Enabled {
				return fmt.Errorf("replication is not enabled in the config")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Restoring database from %s/%s to %s...\n",
				cfg.Storage.Replicate.Bucket, cfg.Storage.Replicate.Path, dbPath)

			rep := replicate.New(cfg.Storage.Replicate, logger)
			if err := rep.Restore(cmd.Context(), dbPath); err != nil {
				return fmt.Errorf("restore failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Restore complete.")
			return nil
		},
	}
}
