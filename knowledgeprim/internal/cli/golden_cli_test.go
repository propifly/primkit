package cli

import (
	"context"
	"testing"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Seed helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

func seedEntity(t *testing.T, s store.Store, typ, title string, body *string) *model.Entity {
	t.Helper()
	ctx := context.Background()
	entity := &model.Entity{
		Type:   typ,
		Title:  title,
		Body:   body,
		Source: "test",
	}
	err := s.CaptureEntity(ctx, entity, nil)
	require.NoError(t, err)
	return entity
}

// ---------------------------------------------------------------------------
// Golden: capture
// ---------------------------------------------------------------------------

func TestGolden_Capture(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "text", format: "text"},
		{name: "json", format: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			out := goldenExecCmd(t, s, "capture",
				"--type", "thought",
				"--title", "Edge compute is the real moat",
				"--source", "test",
				"-f", tt.format,
			)
			assertGolden(t, "capture_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// Golden: search
// ---------------------------------------------------------------------------

func TestGolden_Search(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			if tt.seed {
				seedEntity(t, s, "thought", "LLMs will commoditize", strPtr("Large language models becoming commodities"))
				seedEntity(t, s, "article", "On-Device LLMs", strPtr("Running large language models on edge devices"))
			}
			out := goldenExecCmd(t, s, "search", "language", "--mode", "fts", "-f", tt.format)
			assertGolden(t, "search_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// Golden: get
// ---------------------------------------------------------------------------

func TestGolden_Get(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "text", format: "text"},
		{name: "json", format: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			entity := seedEntity(t, s, "concept", "Retrieval augmented generation", strPtr("Combining search with LLM generation"))
			out := goldenExecCmd(t, s, "get", entity.ID, "-f", tt.format)
			assertGolden(t, "get_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// Golden: types
// ---------------------------------------------------------------------------

func TestGolden_Types(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			if tt.seed {
				seedEntity(t, s, "thought", "First thought", nil)
				seedEntity(t, s, "thought", "Second thought", nil)
				seedEntity(t, s, "article", "An article", nil)
			}
			out := goldenExecCmd(t, s, "types", "-f", tt.format)
			assertGolden(t, "types_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// Golden: relationships
// ---------------------------------------------------------------------------

func TestGolden_Relationships(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			if tt.seed {
				e1 := seedEntity(t, s, "concept", "Edge computing", nil)
				e2 := seedEntity(t, s, "concept", "On-device inference", nil)
				ctx := context.Background()
				edgeCtx := "conceptual relationship"
				err := s.CreateEdge(ctx, &model.Edge{
					SourceID:     e1.ID,
					TargetID:     e2.ID,
					Relationship: "related_to",
					Weight:       1.0,
					Context:      &edgeCtx,
				})
				require.NoError(t, err)
			}
			out := goldenExecCmd(t, s, "relationships", "-f", tt.format)
			assertGolden(t, "relationships_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// Golden: stats
// ---------------------------------------------------------------------------

func TestGolden_Stats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty_text", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goldenTestStore(t)
			if tt.seed {
				e1 := seedEntity(t, s, "thought", "Latency matters", nil)
				e2 := seedEntity(t, s, "article", "On local inference", strPtr("Running models locally"))
				ctx := context.Background()
				err := s.CreateEdge(ctx, &model.Edge{
					SourceID:     e1.ID,
					TargetID:     e2.ID,
					Relationship: "supports",
					Weight:       1.0,
				})
				require.NoError(t, err)
			}
			out := goldenExecCmd(t, s, "stats", "-f", tt.format)
			assertGolden(t, "stats_"+tt.name, out)
		})
	}
}
