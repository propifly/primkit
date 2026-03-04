// Package store defines the persistence contract for stateprim and provides
// a SQLite implementation. All consumers (CLI, API, MCP) depend on the
// Store interface.
package store

import (
	"context"
	"errors"

	"github.com/propifly/primkit/stateprim/internal/model"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrAlreadyExists is returned by SetIfNew when the key already exists.
var ErrAlreadyExists = errors.New("record already exists")

// ErrImmutable is returned when attempting to update an append record.
var ErrImmutable = errors.New("record is immutable")

// Store is the persistence contract for stateprim.
type Store interface {
	// Set creates or updates a record (upsert). Returns ErrImmutable if the
	// existing record is immutable (append record).
	Set(ctx context.Context, record *model.Record) error

	// Get retrieves a single record by namespace and key.
	// Returns ErrNotFound if the record doesn't exist.
	Get(ctx context.Context, namespace, key string) (*model.Record, error)

	// Has checks whether a key exists in a namespace.
	Has(ctx context.Context, namespace, key string) (bool, error)

	// SetIfNew creates a record only if the key doesn't exist in the namespace.
	// Returns ErrAlreadyExists if the key already exists.
	SetIfNew(ctx context.Context, record *model.Record) error

	// Append creates an immutable record with an auto-generated key.
	// The key and timestamps are set by the store.
	Append(ctx context.Context, namespace string, value []byte) (*model.Record, error)

	// Delete removes a record by namespace and key.
	// Returns ErrNotFound if the record doesn't exist.
	Delete(ctx context.Context, namespace, key string) error

	// Query returns records matching the filter. If CountOnly is set on the
	// filter, only the count is populated in the return value.
	Query(ctx context.Context, filter *model.QueryFilter) ([]*model.Record, int, error)

	// Purge deletes records older than the specified duration in a namespace.
	// Returns the number of records deleted.
	Purge(ctx context.Context, namespace string, olderThan string) (int, error)

	// ListNamespaces returns all namespaces with their record counts.
	ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error)

	// Stats returns aggregate counts.
	Stats(ctx context.Context) (*model.Stats, error)

	// ExportRecords returns all records, optionally filtered by namespace.
	ExportRecords(ctx context.Context, namespace string) ([]*model.Record, error)

	// ImportRecords bulk-inserts records, preserving their original keys.
	ImportRecords(ctx context.Context, records []*model.Record) error

	// Close releases the database connection.
	Close() error
}
