package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/primkit/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCaptureAndGetEntity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	body := "Edge computing is key for real-time inference."
	url := "https://example.com/article"
	entity := &model.Entity{
		Type:   "article",
		Title:  "On-Device LLMs",
		Body:   &body,
		URL:    &url,
		Source: "test",
	}

	err := s.CaptureEntity(ctx, entity, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
	assert.True(t, len(entity.ID) > 2 && entity.ID[:2] == "e_")

	// Get it back.
	got, err := s.GetEntity(ctx, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.ID, got.ID)
	assert.Equal(t, "article", got.Type)
	assert.Equal(t, "On-Device LLMs", got.Title)
	assert.Equal(t, "Edge computing is key for real-time inference.", *got.Body)
	assert.Equal(t, "https://example.com/article", *got.URL)
	assert.Equal(t, "test", got.Source)
}

func TestCaptureWithEmbedding(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entity := &model.Entity{
		Type:   "thought",
		Title:  "Embeddings are useful",
		Source: "test",
	}
	embedding := []float32{0.1, 0.2, 0.3, 0.4}

	err := s.CaptureEntity(ctx, entity, embedding)
	require.NoError(t, err)

	// Verify vector was stored.
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM entity_vectors WHERE entity_id = ?", entity.ID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestCaptureValidation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Missing type.
	err := s.CaptureEntity(ctx, &model.Entity{Title: "x", Source: "test"}, nil)
	assert.Error(t, err)

	// Missing title.
	err = s.CaptureEntity(ctx, &model.Entity{Type: "x", Source: "test"}, nil)
	assert.Error(t, err)

	// Missing source.
	err = s.CaptureEntity(ctx, &model.Entity{Type: "x", Title: "y"}, nil)
	assert.Error(t, err)
}

func TestUpdateEntity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entity := &model.Entity{Type: "article", Title: "Original", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, entity, nil))

	newTitle := "Updated Title"
	newBody := "New body content"
	err := s.UpdateEntity(ctx, entity.ID, &model.EntityUpdate{
		Title: &newTitle,
		Body:  &newBody,
	})
	require.NoError(t, err)

	got, err := s.GetEntity(ctx, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", got.Title)
	assert.Equal(t, "New body content", *got.Body)
}

func TestUpdateEntityNotFound(t *testing.T) {
	s := newTestStore(t)
	title := "x"
	err := s.UpdateEntity(context.Background(), "e_nonexistent", &model.EntityUpdate{Title: &title})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteEntity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entity := &model.Entity{Type: "thought", Title: "Temporary", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, entity, nil))

	err := s.DeleteEntity(ctx, entity.ID)
	require.NoError(t, err)

	_, err = s.GetEntity(ctx, entity.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteEntityNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.DeleteEntity(context.Background(), "e_nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteEntityCascadesEdges(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "Entity 1", Source: "test"}
	e2 := &model.Entity{Type: "b", Title: "Entity 2", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{
		SourceID:     e1.ID,
		TargetID:     e2.ID,
		Relationship: "relates_to",
	}))

	// Delete e1 → edge should be gone.
	require.NoError(t, s.DeleteEntity(ctx, e1.ID))

	var edgeCount int
	s.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&edgeCount)
	assert.Equal(t, 0, edgeCount)
}

func TestEdgeCRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "A", Source: "test"}
	e2 := &model.Entity{Type: "b", Title: "B", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))

	// Create edge.
	edgeCtx := "They are related because..."
	edge := &model.Edge{
		SourceID:     e1.ID,
		TargetID:     e2.ID,
		Relationship: "relates_to",
		Context:      &edgeCtx,
	}
	require.NoError(t, s.CreateEdge(ctx, edge))
	assert.Equal(t, 1.0, edge.Weight)

	// Duplicate should fail.
	err := s.CreateEdge(ctx, &model.Edge{
		SourceID:     e1.ID,
		TargetID:     e2.ID,
		Relationship: "relates_to",
	})
	assert.ErrorIs(t, err, ErrEdgeExists)

	// Different relationship should succeed.
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{
		SourceID:     e1.ID,
		TargetID:     e2.ID,
		Relationship: "extends",
	}))

	// Strengthen.
	require.NoError(t, s.StrengthenEdge(ctx, e1.ID, e2.ID, "relates_to"))
	got, err := s.GetEntity(ctx, e1.ID)
	require.NoError(t, err)
	for _, e := range got.Edges {
		if e.Relationship == "relates_to" {
			assert.Equal(t, 2.0, e.Weight)
		}
	}

	// Update edge.
	newCtx := "Updated context"
	require.NoError(t, s.UpdateEdge(ctx, e1.ID, e2.ID, "relates_to", &model.EdgeUpdate{
		Context: &newCtx,
	}))

	// Delete edge.
	require.NoError(t, s.DeleteEdge(ctx, e1.ID, e2.ID, "relates_to"))
	err = s.DeleteEdge(ctx, e1.ID, e2.ID, "relates_to")
	assert.ErrorIs(t, err, ErrEdgeNotFound)
}

func TestEdgeValidation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "A", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))

	// Self-edge.
	err := s.CreateEdge(ctx, &model.Edge{
		SourceID: e1.ID, TargetID: e1.ID, Relationship: "relates_to",
	})
	assert.Error(t, err)

	// Missing fields.
	err = s.CreateEdge(ctx, &model.Edge{SourceID: e1.ID, Relationship: "x"})
	assert.Error(t, err)
}

func TestSearchFTS(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	body1 := "Privacy preserving inference on mobile devices"
	body2 := "Building REST APIs with Go"
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "On-Device Privacy", Body: &body1, Source: "test",
	}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "Go REST APIs", Body: &body2, Source: "test",
	}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "thought", Title: "Privacy is important", Source: "test",
	}, nil))

	results, err := s.SearchFTS(ctx, "privacy", nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	// Type filter.
	results, err = s.SearchFTS(ctx, "privacy", &model.SearchFilter{Type: "article"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "article", results[0].Entity.Type)
}

func TestSearchVector(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "Similar A", Source: "test"}
	e2 := &model.Entity{Type: "b", Title: "Similar B", Source: "test"}
	e3 := &model.Entity{Type: "a", Title: "Different C", Source: "test"}

	require.NoError(t, s.CaptureEntity(ctx, e1, []float32{1.0, 0.0, 0.0}))
	require.NoError(t, s.CaptureEntity(ctx, e2, []float32{0.9, 0.1, 0.0}))
	require.NoError(t, s.CaptureEntity(ctx, e3, []float32{0.0, 0.0, 1.0}))

	// Search for something similar to e1.
	results, err := s.SearchVector(ctx, []float32{1.0, 0.0, 0.0}, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 2)
	// First result should be e1 (exact match).
	assert.Equal(t, e1.ID, results[0].Entity.ID)
}

func TestSearchHybrid(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	body := "Machine learning on embedded devices"
	e1 := &model.Entity{Type: "article", Title: "Edge ML", Body: &body, Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, []float32{0.5, 0.5, 0.0}))

	e2 := &model.Entity{Type: "article", Title: "Cloud ML", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e2, []float32{0.4, 0.6, 0.0}))

	// Hybrid search should work even if FTS finds different results than vector.
	results, err := s.SearchHybrid(ctx, "embedded devices", []float32{0.5, 0.5, 0.0}, nil)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0)
}

func TestSearchHybridFallsBackToFTS(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	body := "Testing hybrid search"
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "Hybrid Search", Body: &body, Source: "test",
	}, nil))

	// No embedding → falls back to FTS only.
	results, err := s.SearchHybrid(ctx, "hybrid search", nil, nil)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0)
}

func TestRelated(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create a graph: A → B → C
	a := &model.Entity{Type: "a", Title: "A", Source: "test"}
	b := &model.Entity{Type: "b", Title: "B", Source: "test"}
	c := &model.Entity{Type: "c", Title: "C", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, a, nil))
	require.NoError(t, s.CaptureEntity(ctx, b, nil))
	require.NoError(t, s.CaptureEntity(ctx, c, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: a.ID, TargetID: b.ID, Relationship: "extends"}))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: b.ID, TargetID: c.ID, Relationship: "relates_to"}))

	// Depth 1 from A → should find B.
	results, err := s.Related(ctx, a.ID, &model.TraversalOpts{Depth: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, b.ID, results[0].Entity.ID)

	// Depth 2 from A → should find B and C.
	results, err = s.Related(ctx, a.ID, &model.TraversalOpts{Depth: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))

	// Relationship filter.
	results, err = s.Related(ctx, a.ID, &model.TraversalOpts{Depth: 2, Relationship: "extends"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(results)) // Only B via "extends".
}

func TestRelatedCyclePrevention(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Cycle: A → B → C → A.
	a := &model.Entity{Type: "a", Title: "A", Source: "test"}
	b := &model.Entity{Type: "b", Title: "B", Source: "test"}
	c := &model.Entity{Type: "c", Title: "C", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, a, nil))
	require.NoError(t, s.CaptureEntity(ctx, b, nil))
	require.NoError(t, s.CaptureEntity(ctx, c, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: a.ID, TargetID: b.ID, Relationship: "r"}))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: b.ID, TargetID: c.ID, Relationship: "r"}))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: c.ID, TargetID: a.ID, Relationship: "r"}))

	// Should not infinite loop.
	results, err := s.Related(ctx, a.ID, &model.TraversalOpts{Depth: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, len(results)) // B and C, not A again.
}

func TestDiscoverOrphans(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	orphan := &model.Entity{Type: "thought", Title: "Alone", Source: "test"}
	connected := &model.Entity{Type: "article", Title: "Connected", Source: "test"}
	other := &model.Entity{Type: "article", Title: "Other", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, orphan, nil))
	require.NoError(t, s.CaptureEntity(ctx, connected, nil))
	require.NoError(t, s.CaptureEntity(ctx, other, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{
		SourceID: connected.ID, TargetID: other.ID, Relationship: "relates_to",
	}))

	report, err := s.Discover(ctx, &model.DiscoverOpts{Orphans: true})
	require.NoError(t, err)
	assert.Equal(t, 1, len(report.Orphans))
	assert.Equal(t, orphan.ID, report.Orphans[0].ID)
}

func TestDiscoverWeakEdges(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "A", Source: "test"}
	e2 := &model.Entity{Type: "b", Title: "B", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))

	// Edge with no context.
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{
		SourceID: e1.ID, TargetID: e2.ID, Relationship: "similar_to",
	}))

	report, err := s.Discover(ctx, &model.DiscoverOpts{WeakEdges: true})
	require.NoError(t, err)
	assert.Equal(t, 1, len(report.WeakEdges))
}

func TestListTypes(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{Type: "article", Title: "A1", Source: "test"}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{Type: "article", Title: "A2", Source: "test"}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{Type: "thought", Title: "T1", Source: "test"}, nil))

	types, err := s.ListTypes(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(types))
	assert.Equal(t, "article", types[0].Type)
	assert.Equal(t, 2, types[0].Count)
}

func TestListRelationships(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "a", Title: "A", Source: "test"}
	e2 := &model.Entity{Type: "b", Title: "B", Source: "test"}
	e3 := &model.Entity{Type: "c", Title: "C", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))
	require.NoError(t, s.CaptureEntity(ctx, e3, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: e1.ID, TargetID: e2.ID, Relationship: "relates_to"}))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: e2.ID, TargetID: e3.ID, Relationship: "relates_to"}))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: e1.ID, TargetID: e3.ID, Relationship: "extends"}))

	rels, err := s.ListRelationships(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(rels))
}

func TestStats(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "article", Title: "A", Source: "test"}
	e2 := &model.Entity{Type: "thought", Title: "B", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, []float32{0.1}))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))

	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: e1.ID, TargetID: e2.ID, Relationship: "r"}))

	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.EntityCount)
	assert.Equal(t, 1, stats.EdgeCount)
	assert.Equal(t, 1, stats.VectorCount)
	assert.Equal(t, 0, stats.OrphanCount)
	assert.Equal(t, 2, stats.TypeCount)
}

func TestExportImport(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e1 := &model.Entity{Type: "article", Title: "A", Source: "test"}
	e2 := &model.Entity{Type: "thought", Title: "B", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, e1, nil))
	require.NoError(t, s.CaptureEntity(ctx, e2, nil))
	require.NoError(t, s.CreateEdge(ctx, &model.Edge{SourceID: e1.ID, TargetID: e2.ID, Relationship: "extends"}))

	// Export.
	data, err := s.ExportEntities(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, len(data.Entities))
	assert.Equal(t, 1, len(data.Edges))

	// Import into a fresh store.
	s2 := newTestStore(t)
	err = s2.ImportEntities(ctx, data)
	require.NoError(t, err)

	stats, err := s2.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.EntityCount)
	assert.Equal(t, 1, stats.EdgeCount)
}

func TestExportFilterByType(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{Type: "article", Title: "A", Source: "test"}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{Type: "thought", Title: "T", Source: "test"}, nil))

	data, err := s.ExportEntities(ctx, &model.ExportFilter{Type: "article"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(data.Entities))
	assert.Equal(t, "article", data.Entities[0].Type)
}

func TestEntityWithProperties(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entity := &model.Entity{
		Type:       "concept",
		Title:      "With Props",
		Source:     "test",
		Properties: json.RawMessage(`{"domain":"ai","confidence":"high"}`),
	}
	require.NoError(t, s.CaptureEntity(ctx, entity, nil))

	got, err := s.GetEntity(ctx, entity.ID)
	require.NoError(t, err)
	assert.JSONEq(t, `{"domain":"ai","confidence":"high"}`, string(got.Properties))
}

func TestCosineDistance(t *testing.T) {
	// Identical vectors → distance 0.
	assert.InDelta(t, 0.0, cosineDistance([]float32{1, 0, 0}, []float32{1, 0, 0}), 0.001)

	// Orthogonal vectors → distance 1.
	assert.InDelta(t, 1.0, cosineDistance([]float32{1, 0, 0}, []float32{0, 1, 0}), 0.001)

	// Opposite vectors → distance 2.
	assert.InDelta(t, 2.0, cosineDistance([]float32{1, 0, 0}, []float32{-1, 0, 0}), 0.001)
}

func TestFloat32Roundtrip(t *testing.T) {
	original := []float32{0.1, 0.2, 0.3, -0.5, 1.0}
	bytes := float32sToBytes(original)
	restored := bytesToFloat32s(bytes)
	assert.Equal(t, original, restored)
}

// ---------------------------------------------------------------------------
// Embedding metadata tests
// ---------------------------------------------------------------------------

func TestGetEmbeddingMetaEmpty(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetEmbeddingMeta(context.Background())
	assert.ErrorIs(t, err, ErrNoEmbeddingMeta)
}

func TestSetAndGetEmbeddingMeta(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	meta := &model.EmbeddingMeta{
		Provider:   "openai",
		Model:      "text-embedding-3-small",
		Dimensions: 1536,
	}
	require.NoError(t, s.SetEmbeddingMeta(ctx, meta))

	got, err := s.GetEmbeddingMeta(ctx)
	require.NoError(t, err)
	assert.Equal(t, "openai", got.Provider)
	assert.Equal(t, "text-embedding-3-small", got.Model)
	assert.Equal(t, 1536, got.Dimensions)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestSetEmbeddingMetaOverwrites(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetEmbeddingMeta(ctx, &model.EmbeddingMeta{
		Provider: "openai", Model: "text-embedding-3-small", Dimensions: 1536,
	}))
	require.NoError(t, s.SetEmbeddingMeta(ctx, &model.EmbeddingMeta{
		Provider: "gemini", Model: "text-embedding-004", Dimensions: 768,
	}))

	got, err := s.GetEmbeddingMeta(ctx)
	require.NoError(t, err)
	assert.Equal(t, "gemini", got.Provider)
	assert.Equal(t, "text-embedding-004", got.Model)
	assert.Equal(t, 768, got.Dimensions)
}

func TestCheckEmbeddingMetaNoMeta(t *testing.T) {
	s := newTestStore(t)
	// No metadata set → should pass (first embed will set it).
	err := s.CheckEmbeddingMeta(context.Background(), "openai", "text-embedding-3-small", 1536)
	assert.NoError(t, err)
}

func TestCheckEmbeddingMetaMatch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetEmbeddingMeta(ctx, &model.EmbeddingMeta{
		Provider: "openai", Model: "text-embedding-3-small", Dimensions: 1536,
	}))

	// Same provider and model → pass.
	err := s.CheckEmbeddingMeta(ctx, "openai", "text-embedding-3-small", 1536)
	assert.NoError(t, err)
}

func TestCheckEmbeddingMetaMismatch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetEmbeddingMeta(ctx, &model.EmbeddingMeta{
		Provider: "openai", Model: "text-embedding-3-small", Dimensions: 1536,
	}))

	// Different provider → mismatch.
	err := s.CheckEmbeddingMeta(ctx, "gemini", "text-embedding-004", 768)
	assert.ErrorIs(t, err, ErrEmbeddingMismatch)

	// Same provider, different model → mismatch.
	err = s.CheckEmbeddingMeta(ctx, "openai", "text-embedding-3-large", 3072)
	assert.ErrorIs(t, err, ErrEmbeddingMismatch)
}

func TestEnsureEmbeddingMetaSetsOnce(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// First call sets metadata.
	require.NoError(t, s.EnsureEmbeddingMeta(ctx, "openai", "text-embedding-3-small", 1536))

	got, err := s.GetEmbeddingMeta(ctx)
	require.NoError(t, err)
	assert.Equal(t, "openai", got.Provider)

	// Second call is a no-op (doesn't overwrite).
	require.NoError(t, s.EnsureEmbeddingMeta(ctx, "gemini", "text-embedding-004", 768))

	got, err = s.GetEmbeddingMeta(ctx)
	require.NoError(t, err)
	assert.Equal(t, "openai", got.Provider) // Still openai.
}

func TestStripVectors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create entity with embedding.
	entity := &model.Entity{Type: "concept", Title: "Test", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, entity, []float32{0.1, 0.2, 0.3}))

	// Set metadata.
	require.NoError(t, s.SetEmbeddingMeta(ctx, &model.EmbeddingMeta{
		Provider: "openai", Model: "text-embedding-3-small", Dimensions: 1536,
	}))

	// Verify vectors and meta exist.
	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.VectorCount)

	_, err = s.GetEmbeddingMeta(ctx)
	require.NoError(t, err)

	// Strip.
	require.NoError(t, s.StripVectors(ctx))

	// Verify everything is gone.
	stats, err = s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.VectorCount)

	_, err = s.GetEmbeddingMeta(ctx)
	assert.ErrorIs(t, err, ErrNoEmbeddingMeta)

	// Entity still exists.
	assert.Equal(t, 1, stats.EntityCount)
}

func TestUpdateEntityVector(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create entity without embedding.
	entity := &model.Entity{Type: "concept", Title: "Test", Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, entity, nil))

	stats, _ := s.Stats(ctx)
	assert.Equal(t, 0, stats.VectorCount)

	// Add vector.
	require.NoError(t, s.UpdateEntityVector(ctx, entity.ID, []float32{0.1, 0.2, 0.3}))

	stats, _ = s.Stats(ctx)
	assert.Equal(t, 1, stats.VectorCount)

	// Update vector (replace).
	require.NoError(t, s.UpdateEntityVector(ctx, entity.ID, []float32{0.4, 0.5, 0.6}))

	stats, _ = s.Stats(ctx)
	assert.Equal(t, 1, stats.VectorCount) // Still 1, not 2.

	// Vector search should find it with the new embedding.
	results, err := s.SearchVector(ctx, []float32{0.4, 0.5, 0.6}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	assert.Equal(t, entity.ID, results[0].Entity.ID)
}

func TestStripVectorsEmptyDB(t *testing.T) {
	s := newTestStore(t)
	// Should not fail on empty database.
	require.NoError(t, s.StripVectors(context.Background()))
}

func TestFTSWorksAfterStripVectors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	body := "knowledge graph with vector embeddings"
	entity := &model.Entity{Type: "article", Title: "Graph Embeddings", Body: &body, Source: "test"}
	require.NoError(t, s.CaptureEntity(ctx, entity, []float32{0.1, 0.2, 0.3}))

	require.NoError(t, s.StripVectors(ctx))

	// FTS should still work.
	results, err := s.SearchFTS(ctx, "knowledge graph", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, entity.ID, results[0].Entity.ID)
}
