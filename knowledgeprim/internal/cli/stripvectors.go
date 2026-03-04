package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStripVectorsCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "strip-vectors",
		Short: "Remove all vector embeddings and embedding metadata",
		Long: `Deletes all vector embeddings and the embedding metadata from the database.
The database reverts to FTS5-only search mode. Entities and edges are not
affected.

Use this when you want to start fresh with a different embedding provider,
or when FTS5 search is sufficient and you don't need vector search.

Requires --confirm flag (destructive operation).

  knowledgeprim strip-vectors --confirm
  knowledgeprim strip-vectors --confirm --db seo-expert.db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("strip-vectors permanently deletes all embeddings; pass --confirm to proceed")
			}

			s := storeFromContext(cmd.Context())

			// Get stats before stripping.
			stats, err := s.Stats(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting stats: %w", err)
			}
			vectorCount := stats.VectorCount

			if err := s.StripVectors(cmd.Context()); err != nil {
				return fmt.Errorf("stripping vectors: %w", err)
			}

			if getFormat(cmd) == "json" {
				return writeJSON(cmd.OutOrStdout(), map[string]interface{}{
					"vectors_removed": vectorCount,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Stripped %d vectors and embedding metadata.\n", vectorCount)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm destructive operation")

	return cmd
}
