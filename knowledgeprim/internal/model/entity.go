package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Entity is a typed knowledge node in the graph.
type Entity struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Body       *string         `json:"body,omitempty"`
	URL        *string         `json:"url,omitempty"`
	Source     string          `json:"source"`
	Properties json.RawMessage `json:"properties,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Edges      []*Edge         `json:"edges,omitempty"`
}

// Validate checks required fields on an entity.
// Does NOT check ID — the store assigns it.
func (e *Entity) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("type is required")
	}
	if e.Title == "" {
		return fmt.Errorf("title is required")
	}
	if e.Source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

// EmbeddingText returns the text to send to an embedding provider:
// title + "\n\n" + body (if body is present), or just title.
func (e *Entity) EmbeddingText() string {
	if e.Body != nil && *e.Body != "" {
		return e.Title + "\n\n" + *e.Body
	}
	return e.Title
}

// EntityUpdate holds the optional fields for updating an entity.
// Only non-nil fields are applied.
type EntityUpdate struct {
	Title      *string          `json:"title,omitempty"`
	Body       *string          `json:"body,omitempty"`
	URL        *string          `json:"url,omitempty"`
	Properties *json.RawMessage `json:"properties,omitempty"`
}

// Edge is a weighted, contextualized connection between two entities.
type Edge struct {
	SourceID     string    `json:"source_id"`
	TargetID     string    `json:"target_id"`
	Relationship string    `json:"relationship"`
	Weight       float64   `json:"weight"`
	Context      *string   `json:"context,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidateEdge checks required fields on an edge.
func (e *Edge) Validate() error {
	if e.SourceID == "" {
		return fmt.Errorf("source_id is required")
	}
	if e.TargetID == "" {
		return fmt.Errorf("target_id is required")
	}
	if e.Relationship == "" {
		return fmt.Errorf("relationship is required")
	}
	if e.SourceID == e.TargetID {
		return fmt.Errorf("self-edges are not allowed")
	}
	return nil
}

// EdgeUpdate holds optional fields for updating an edge.
type EdgeUpdate struct {
	Context *string  `json:"context,omitempty"`
	Weight  *float64 `json:"weight,omitempty"`
}

// SearchFilter controls search result filtering.
type SearchFilter struct {
	Type  string `json:"type,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// DefaultSearchLimit is the default number of results returned by search.
const DefaultSearchLimit = 20

// SearchResult is a single result from a search operation.
type SearchResult struct {
	Entity *Entity `json:"entity"`
	Score  float64 `json:"score"`
}

// TraversalOpts controls graph traversal behavior.
type TraversalOpts struct {
	Depth        int     `json:"depth,omitempty"`
	Relationship string  `json:"relationship,omitempty"`
	Direction    string  `json:"direction,omitempty"` // "outgoing", "incoming", "both"
	MinWeight    float64 `json:"min_weight,omitempty"`
}

// TraversalResult is a single entity found during graph traversal.
type TraversalResult struct {
	Entity       *Entity `json:"entity"`
	Relationship string  `json:"relationship"`
	Direction    string  `json:"direction"` // "outgoing" or "incoming"
	Depth        int     `json:"depth"`
	Weight       float64 `json:"weight"`
}

// DiscoverOpts controls which discovery operations to run.
type DiscoverOpts struct {
	Orphans   bool `json:"orphans,omitempty"`
	Clusters  bool `json:"clusters,omitempty"`
	Bridges   bool `json:"bridges,omitempty"`
	Temporal  bool `json:"temporal,omitempty"`
	WeakEdges bool `json:"weak_edges,omitempty"`
}

// DiscoverAll returns opts with all discovery operations enabled.
func DiscoverAll() *DiscoverOpts {
	return &DiscoverOpts{
		Orphans:   true,
		Clusters:  true,
		Bridges:   true,
		Temporal:  true,
		WeakEdges: true,
	}
}

// DiscoverReport contains the results of discovery operations.
type DiscoverReport struct {
	Orphans   []*Entity       `json:"orphans,omitempty"`
	Clusters  []Cluster       `json:"clusters,omitempty"`
	Bridges   []*BridgeEntity `json:"bridges,omitempty"`
	Temporal  []TemporalGroup `json:"temporal,omitempty"`
	WeakEdges []*Edge         `json:"weak_edges,omitempty"`
}

// Cluster is a group of densely connected entities.
type Cluster struct {
	Entities []*Entity `json:"entities"`
	Size     int       `json:"size"`
}

// BridgeEntity is an entity that connects otherwise separate clusters.
type BridgeEntity struct {
	Entity     *Entity `json:"entity"`
	EdgeCount  int     `json:"edge_count"`
	ClusterIDs []int   `json:"cluster_ids,omitempty"`
}

// TemporalGroup groups entities by type and time period.
type TemporalGroup struct {
	Period   string    `json:"period"` // e.g., "2024-W01", "2024-01"
	Type     string    `json:"type"`
	Count    int       `json:"count"`
	Entities []*Entity `json:"entities,omitempty"`
}

// TypeCount is a type with its entity count.
type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// RelationshipCount is a relationship type with its edge count.
type RelationshipCount struct {
	Relationship string `json:"relationship"`
	Count        int    `json:"count"`
}

// Stats contains aggregate statistics about the knowledge graph.
type Stats struct {
	EntityCount int    `json:"entity_count"`
	EdgeCount   int    `json:"edge_count"`
	VectorCount int    `json:"vector_count"`
	OrphanCount int    `json:"orphan_count"`
	TypeCount   int    `json:"type_count"`
	DBSize      int64  `json:"db_size_bytes"`
	DBPath      string `json:"db_path,omitempty"`
}

// ExportFilter controls which entities to export.
type ExportFilter struct {
	Type string `json:"type,omitempty"`
}

// ExportData is the full export format for import/export.
type ExportData struct {
	Entities []*Entity `json:"entities"`
	Edges    []*Edge   `json:"edges"`
}
