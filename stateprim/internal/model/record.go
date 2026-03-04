// Package model defines the core data types for stateprim. A Record is a
// namespaced key-value pair with a JSON payload. Three access patterns share
// the same underlying model:
//
//   - Key-value state: set/get/update for current state
//   - Dedup lookups: has/set-if-new for existence checks
//   - Append records: immutable, timestamped log entries
package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Record is the core entity — a namespaced key with a JSON value. The same
// struct serves all three access patterns; the immutable flag distinguishes
// append records from regular key-value entries.
type Record struct {
	Namespace string          `json:"namespace"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	Immutable bool            `json:"immutable"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Validate checks that a record has the minimum required fields.
func (r *Record) Validate() error {
	if r.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if r.Key == "" {
		return fmt.Errorf("key is required")
	}
	if len(r.Value) == 0 {
		return fmt.Errorf("value is required")
	}
	return nil
}

// NamespaceInfo holds a namespace name and the number of records it contains.
type NamespaceInfo struct {
	Namespace string `json:"namespace"`
	Count     int    `json:"count"`
}

// Stats holds aggregate counts across all namespaces.
type Stats struct {
	TotalRecords    int `json:"total_records"`
	TotalNamespaces int `json:"total_namespaces"`
}

// QueryFilter holds parameters for querying records within a namespace.
type QueryFilter struct {
	Namespace string        // Required: which namespace to query.
	Since     time.Duration // Only records created/updated within this window.
	KeyPrefix string        // Only records whose key starts with this prefix.
	CountOnly bool          // If true, return count instead of records.
}
