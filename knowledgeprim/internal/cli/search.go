package cli

import (
	"fmt"
	"strings"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		entityType string
		limit      int
		mode       string
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the knowledge graph",
		Long: `Searches entities using hybrid retrieval (FTS5 + vector + RRF).

  knowledgeprim search "privacy preserving inference"
  knowledgeprim search "Litestream" --mode fts
  knowledgeprim search "embedded replication" --mode vector --type article`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			embedder := embedderFromContext(cmd.Context())
			query := strings.Join(args, " ")

			filter := &model.SearchFilter{
				Type:  entityType,
				Limit: limit,
			}

			var results []*model.SearchResult
			var err error

			switch mode {
			case "fts":
				results, err = s.SearchFTS(cmd.Context(), query, filter)
			case "vector":
				if embedder == nil {
					return fmt.Errorf("vector search requires an embedding provider (configure in config.yaml)")
				}
				embedding, embErr := embedder.Embed(cmd.Context(), query)
				if embErr != nil {
					return fmt.Errorf("embedding query: %w", embErr)
				}
				results, err = s.SearchVector(cmd.Context(), embedding, filter)
			default: // hybrid
				var embedding []float32
				if embedder != nil {
					embedding, _ = embedder.Embed(cmd.Context(), query)
				}
				results, err = s.SearchHybrid(cmd.Context(), query, embedding, filter)
			}

			if err != nil {
				return fmt.Errorf("searching: %w", err)
			}

			return outputSearchResults(cmd, results)
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "filter by entity type")
	cmd.Flags().IntVar(&limit, "limit", 20, "max results")
	cmd.Flags().StringVar(&mode, "mode", "hybrid", "search mode: hybrid, fts, vector")

	return cmd
}
