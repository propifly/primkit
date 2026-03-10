package store

import (
	"context"
	"testing"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Unit tests: sanitizeFTS5Query (pure function)
// ---------------------------------------------------------------------------

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain term", input: "privacy", want: "privacy"},
		{name: "multiple plain terms", input: "privacy devices", want: "privacy devices"},
		{name: "hyphenated term", input: "agent-first", want: `"agent-first"`},
		{name: "multiple hyphens", input: "real-time-inference", want: `"real-time-inference"`},
		{name: "mixed plain and hyphenated", input: "the agent-first approach", want: `the "agent-first" approach`},
		{name: "colon (column syntax)", input: "title:privacy", want: `"title:privacy"`},
		{name: "asterisk (prefix)", input: "priv*", want: `"priv*"`},
		{name: "caret (initial token)", input: "^privacy", want: `"^privacy"`},
		{name: "parentheses", input: "(privacy)", want: `"(privacy)"`},
		{name: "curly braces (NEAR group)", input: "{privacy}", want: `"{privacy}"`},
		{name: "keyword AND", input: "privacy AND devices", want: `privacy "AND" devices`},
		{name: "keyword OR", input: "privacy OR devices", want: `privacy "OR" devices`},
		{name: "keyword NOT", input: "NOT privacy", want: `"NOT" privacy`},
		{name: "keyword NEAR", input: "privacy NEAR devices", want: `privacy "NEAR" devices`},
		{name: "keyword lowercase not quoted", input: "privacy and devices", want: "privacy and devices"},
		{name: "already quoted", input: `"agent-first"`, want: `"agent-first"`},
		{name: "empty string", input: "", want: ""},
		{name: "whitespace only", input: "   ", want: ""},
		{name: "embedded double quotes", input: `ag"ent-first`, want: `"agent-first"`},
		{name: "multiple special tokens", input: "agent-first real-time title:test", want: `"agent-first" "real-time" "title:test"`},
		{name: "leading trailing whitespace", input: "  agent-first  ", want: `"agent-first"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFTS5Query(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Regression tests: SearchFTS with special characters (integration)
// ---------------------------------------------------------------------------

func seedSearchEntities(t *testing.T, s *SQLiteStore) {
	t.Helper()
	ctx := context.Background()

	body1 := "An agent-first design pattern for building resilient systems"
	body2 := "Real-time inference on edge devices using on-device models"
	body3 := "Building REST APIs with Go and testing strategies"

	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "Agent-First Architecture", Body: &body1, Source: "test",
	}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "Real-Time Edge ML", Body: &body2, Source: "test",
	}, nil))
	require.NoError(t, s.CaptureEntity(ctx, &model.Entity{
		Type: "article", Title: "Go REST APIs", Body: &body3, Source: "test",
	}, nil))
}

func TestSearchFTS_HyphenatedTerm(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	results, err := s.SearchFTS(context.Background(), "agent-first", nil)
	require.NoError(t, err, "SearchFTS should not error on hyphenated term")
	assert.Greater(t, len(results), 0, "should find entities matching agent-first")
}

func TestSearchFTS_MultipleHyphenatedTerms(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	results, err := s.SearchFTS(context.Background(), "real-time on-device", nil)
	require.NoError(t, err, "SearchFTS should not error on multiple hyphenated terms")
	assert.Greater(t, len(results), 0)
}

func TestSearchFTS_ColonInQuery(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	// This should not be interpreted as column:term syntax.
	_, err := s.SearchFTS(context.Background(), "title:architecture", nil)
	assert.NoError(t, err, "colon in query should not cause column-syntax error")
}

func TestSearchFTS_FTS5Keywords(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	// "AND" as a literal term should not break the query.
	_, err := s.SearchFTS(context.Background(), "agent AND first", nil)
	assert.NoError(t, err, "AND keyword should be quoted")

	_, err = s.SearchFTS(context.Background(), "NOT rest", nil)
	assert.NoError(t, err, "NOT keyword should be quoted")

	_, err = s.SearchFTS(context.Background(), "rest OR edge", nil)
	assert.NoError(t, err, "OR keyword should be quoted")
}

func TestSearchFTS_EmptyQuery(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	results, err := s.SearchFTS(context.Background(), "", nil)
	assert.NoError(t, err)
	assert.Nil(t, results)

	results, err = s.SearchFTS(context.Background(), "   ", nil)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchFTS_AsteriskInQuery(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	// Prefix wildcard should be quoted to prevent FTS5 prefix interpretation.
	_, err := s.SearchFTS(context.Background(), "agent*", nil)
	assert.NoError(t, err)
}

func TestSearchHybrid_HyphenatedTerm(t *testing.T) {
	s := newTestStore(t)
	seedSearchEntities(t, s)

	// Hybrid search also goes through SearchFTS — verify it works end to end.
	results, err := s.SearchHybrid(context.Background(), "agent-first", nil, nil)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0)
}
