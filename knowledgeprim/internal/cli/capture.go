package cli

import (
	"encoding/json"
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/spf13/cobra"
)

func newCaptureCmd() *cobra.Command {
	var (
		entityType    string
		title         string
		body          string
		url           string
		source        string
		properties    string
		noAutoConnect bool
		threshold     float64
		force         bool
	)

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a knowledge entity",
		Long: `Creates a new entity in the knowledge graph. If an embedding provider is
configured, the entity is automatically embedded and similar entities are
connected via auto-connect.

  knowledgeprim capture --type article --title "On-Device LLMs" --body "..."
  knowledgeprim capture --type thought --title "Edge compute is the real moat"
  knowledgeprim capture --type article --title "..." --no-auto-connect`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			embedder := embedderFromContext(cmd.Context())

			entity := &model.Entity{
				Type:   entityType,
				Title:  title,
				Source: source,
			}
			if body != "" {
				entity.Body = &body
			}
			if url != "" {
				entity.URL = &url
			}
			if properties != "" {
				entity.Properties = json.RawMessage(properties)
			}

			// Generate embedding if provider is configured.
			var embedding []float32
			if embedder != nil {
				// Check embedding model compatibility unless --force.
				if !force {
					if err := checkEmbeddingCompat(cmd, s, embedder); err != nil {
						return err
					}
				}

				var err error
				embedding, err = embedder.Embed(cmd.Context(), entity.EmbeddingText())
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: embedding failed: %v\n", err)
				}
			}

			if err := s.CaptureEntity(cmd.Context(), entity, embedding); err != nil {
				return fmt.Errorf("capturing entity: %w", err)
			}

			// Set embedding metadata on first successful embed.
			if len(embedding) > 0 && embedder != nil {
				sqlStore, ok := s.(*store.SQLiteStore)
				if ok {
					_ = sqlStore.EnsureEmbeddingMeta(cmd.Context(), embedder.Provider(), embedder.Model(), embedder.Dimensions())
				}
			}

			// Auto-connect: find similar entities and create edges.
			if !noAutoConnect && len(embedding) > 0 {
				autoConnect(cmd, s, entity.ID, embedding, threshold)
			}

			return outputEntity(cmd, entity)
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "entity type (article, thought, concept, pattern, etc.)")
	cmd.Flags().StringVar(&title, "title", "", "entity title")
	cmd.Flags().StringVar(&body, "body", "", "entity body text")
	cmd.Flags().StringVar(&url, "url", "", "source URL")
	cmd.Flags().StringVar(&source, "source", defaultSource(), "who captured this")
	cmd.Flags().StringVar(&properties, "properties", "", "JSON properties")
	cmd.Flags().BoolVar(&noAutoConnect, "no-auto-connect", false, "skip auto-connect")
	cmd.Flags().Float64Var(&threshold, "threshold", 0.35, "auto-connect cosine distance threshold")
	cmd.Flags().BoolVar(&force, "force", false, "bypass embedding model mismatch check")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("title")

	return cmd
}

// autoConnect finds semantically similar entities and creates edges.
func autoConnect(cmd *cobra.Command, s store.Store, entityID string, embedding []float32, threshold float64) {
	results, err := s.SearchVector(cmd.Context(), embedding, &model.SearchFilter{Limit: 10})
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: auto-connect search failed: %v\n", err)
		return
	}

	connected := 0
	for _, r := range results {
		if r.Entity.ID == entityID {
			continue // Skip self.
		}
		distance := 1.0 - r.Score
		if distance > threshold {
			continue // Too dissimilar.
		}

		ctx := fmt.Sprintf("auto-connected: cosine distance %.2f", distance)
		edge := &model.Edge{
			SourceID:     entityID,
			TargetID:     r.Entity.ID,
			Relationship: "similar_to",
			Weight:       1.0,
			Context:      &ctx,
		}
		if err := s.CreateEdge(cmd.Context(), edge); err != nil {
			continue // Edge might already exist.
		}
		connected++
		if connected >= 10 {
			break
		}
	}

	if connected > 0 && getFormat(cmd) != "json" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Auto-connected to %d similar entities.\n", connected)
	}
}

// checkEmbeddingCompat validates that the current embedder matches the db's
// stored embedding metadata. Shared by capture and search commands.
func checkEmbeddingCompat(cmd *cobra.Command, s store.Store, embedder interface {
	Provider() string
	Model() string
	Dimensions() int
}) error {
	sqlStore, ok := s.(*store.SQLiteStore)
	if !ok {
		return nil // Non-SQLite stores don't support metadata checks.
	}
	return sqlStore.CheckEmbeddingMeta(cmd.Context(), embedder.Provider(), embedder.Model(), embedder.Dimensions())
}
