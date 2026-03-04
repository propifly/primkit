// Package store defines the Store interface for knowledgeprim's persistence
// layer. All consumers (CLI, API, MCP) depend on this interface, not the
// concrete SQLite implementation.
package store

import (
	"context"
	"errors"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
)

// Store is the persistence interface for the knowledge graph. It covers
// entity CRUD, edge management, search (FTS + vector + hybrid), graph
// traversal, and discovery operations.
type Store interface {
	// Entity operations.
	CaptureEntity(ctx context.Context, entity *model.Entity, embedding []float32) error
	GetEntity(ctx context.Context, id string) (*model.Entity, error)
	UpdateEntity(ctx context.Context, id string, update *model.EntityUpdate) error
	DeleteEntity(ctx context.Context, id string) error

	// Edge operations.
	CreateEdge(ctx context.Context, edge *model.Edge) error
	UpdateEdge(ctx context.Context, source, target, relationship string, update *model.EdgeUpdate) error
	StrengthenEdge(ctx context.Context, source, target, relationship string) error
	DeleteEdge(ctx context.Context, source, target, relationship string) error

	// Search.
	SearchFTS(ctx context.Context, query string, filter *model.SearchFilter) ([]*model.SearchResult, error)
	SearchVector(ctx context.Context, embedding []float32, filter *model.SearchFilter) ([]*model.SearchResult, error)
	SearchHybrid(ctx context.Context, query string, embedding []float32, filter *model.SearchFilter) ([]*model.SearchResult, error)

	// Graph traversal.
	Related(ctx context.Context, id string, opts *model.TraversalOpts) ([]*model.TraversalResult, error)

	// Discovery.
	Discover(ctx context.Context, opts *model.DiscoverOpts) (*model.DiscoverReport, error)

	// Aggregates.
	ListTypes(ctx context.Context) ([]model.TypeCount, error)
	ListRelationships(ctx context.Context) ([]model.RelationshipCount, error)
	Stats(ctx context.Context) (*model.Stats, error)

	// Export/import.
	ExportEntities(ctx context.Context, filter *model.ExportFilter) (*model.ExportData, error)
	ImportEntities(ctx context.Context, data *model.ExportData) error

	// Embedding metadata.
	GetEmbeddingMeta(ctx context.Context) (*model.EmbeddingMeta, error)
	SetEmbeddingMeta(ctx context.Context, meta *model.EmbeddingMeta) error

	// Embedding maintenance.
	StripVectors(ctx context.Context) error
	UpdateEntityVector(ctx context.Context, entityID string, embedding []float32) error

	// Lifecycle.
	Close() error
}

// Error sentinels.
var (
	ErrNotFound           = errors.New("entity not found")
	ErrEdgeNotFound       = errors.New("edge not found")
	ErrEdgeExists         = errors.New("edge already exists")
	ErrInvalidEntity      = errors.New("invalid entity")
	ErrEmbeddingMismatch  = errors.New("embedding model mismatch")
	ErrNoEmbeddingMeta    = errors.New("no embedding metadata")
)
