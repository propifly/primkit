package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/spf13/cobra"
)

func newReEmbedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "re-embed",
		Short: "Re-embed all entities with the current embedding provider",
		Long: `Iterates all entities, generates new vector embeddings with the currently
configured embedding provider, replaces all existing vectors, and updates
the embedding metadata. Use this when switching embedding providers.

Requires a configured embedding provider (config.yaml or environment variables).

  knowledgeprim re-embed
  knowledgeprim re-embed --db seo-expert.db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			embedder := embedderFromContext(cmd.Context())

			if embedder == nil {
				return fmt.Errorf("re-embed requires a configured embedding provider")
			}

			sqlStore, ok := s.(*store.SQLiteStore)
			if !ok {
				return fmt.Errorf("re-embed is only supported with SQLite stores")
			}

			// Export all entities to iterate them.
			data, err := s.ExportEntities(cmd.Context(), nil)
			if err != nil {
				return fmt.Errorf("loading entities: %w", err)
			}

			total := len(data.Entities)
			if total == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No entities to re-embed.")
				return nil
			}

			isJSON := getFormat(cmd) == "json"
			embedded := 0
			failed := 0

			for i, entity := range data.Entities {
				text := entity.EmbeddingText()
				embedding, err := embedder.Embed(cmd.Context(), text)
				if err != nil {
					failed++
					if !isJSON {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to embed %s: %v\n", entity.ID, err)
					}
					continue
				}

				if err := s.UpdateEntityVector(cmd.Context(), entity.ID, embedding); err != nil {
					failed++
					if !isJSON {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to store vector for %s: %v\n", entity.ID, err)
					}
					continue
				}

				embedded++
				if !isJSON && (i+1)%10 == 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "Progress: %d/%d entities\n", i+1, total)
				}
			}

			// Update embedding metadata.
			if err := sqlStore.SetEmbeddingMeta(cmd.Context(), &model.EmbeddingMeta{
				Provider:   embedder.Provider(),
				Model:      embedder.Model(),
				Dimensions: embedder.Dimensions(),
			}); err != nil {
				return fmt.Errorf("updating embedding metadata: %w", err)
			}

			if isJSON {
				return writeJSON(cmd.OutOrStdout(), map[string]interface{}{
					"total":    total,
					"embedded": embedded,
					"failed":   failed,
					"provider": embedder.Provider(),
					"model":    embedder.Model(),
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Re-embedded %d/%d entities (%d failed) with %s/%s.\n",
				embedded, total, failed, embedder.Provider(), embedder.Model())
			return nil
		},
	}

	return cmd
}
