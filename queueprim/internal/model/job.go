// Package model defines the domain types for queueprim. The central entity is
// Job — a unit of work dispatched by a producer agent and consumed by a worker
// agent with atomic claim semantics and automatic crash recovery.
package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Priority controls dequeue ordering. High-priority jobs are returned before
// normal, normal before low. Within a priority level, ordering is FIFO.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// PriorityRank converts a Priority to its integer rank for ORDER BY.
// Lower numbers sort first (high=0, normal=1, low=2).
func PriorityRank(p Priority) int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityLow:
		return 2
	default:
		return 1 // normal and unknown default to normal
	}
}

// Status represents the lifecycle state of a job.
type Status string

const (
	StatusPending Status = "pending" // available for dequeue
	StatusClaimed Status = "claimed" // exclusively held by a worker
	StatusDone    Status = "done"    // terminal success
	StatusFailed  Status = "failed"  // worker reported failure; may retry
	StatusDead    Status = "dead"    // terminal failure; requires manual inspection
)

// Job is the central domain object. Every field maps directly to the SQLite
// schema. The Payload and Output fields are arbitrary JSON stored as raw bytes.
type Job struct {
	ID            string          `json:"id"`
	Queue         string          `json:"queue"`
	Type          string          `json:"type,omitempty"`
	Priority      Priority        `json:"priority"`
	Payload       json.RawMessage `json:"payload"`
	Status        Status          `json:"status"`
	ClaimedBy     *string         `json:"claimed_by,omitempty"`
	ClaimedAt     *time.Time      `json:"claimed_at,omitempty"`
	VisibleAfter  time.Time       `json:"visible_after"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Output        json.RawMessage `json:"output,omitempty"`
	FailureReason *string         `json:"failure_reason,omitempty"`
	AttemptCount  int             `json:"attempt_count"`
	MaxRetries    int             `json:"max_retries"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Validate checks that a job has all required fields before creation.
// It does not check the ID — that is assigned by the store.
func (j *Job) Validate() error {
	if j.Queue == "" {
		return fmt.Errorf("queue is required")
	}
	if len(j.Payload) == 0 {
		return fmt.Errorf("payload is required")
	}
	if !json.Valid(j.Payload) {
		return fmt.Errorf("payload must be valid JSON")
	}
	switch j.Priority {
	case PriorityHigh, PriorityNormal, PriorityLow, "":
		// ok; empty defaults to normal
	default:
		return fmt.Errorf("priority must be high, normal, or low")
	}
	if j.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be >= 0")
	}
	return nil
}

// Filter specifies which jobs to return from ListJobs. All fields are optional;
// an empty filter returns all jobs across all queues.
type Filter struct {
	Queue     string  // only jobs in this queue
	Status    *Status // only jobs in this status (nil = all)
	Type      string  // only jobs of this type
	OlderThan time.Duration // only jobs created/updated before now-OlderThan
}

// QueueInfo provides summary counts for a single named queue.
type QueueInfo struct {
	Queue   string `json:"queue"`
	Pending int    `json:"pending"`
	Claimed int    `json:"claimed"`
	Done    int    `json:"done"`
	Failed  int    `json:"failed"`
	Dead    int    `json:"dead"`
}

// Stats provides aggregate counts across all jobs in all queues.
type Stats struct {
	TotalPending int `json:"total_pending"`
	TotalClaimed int `json:"total_claimed"`
	TotalDone    int `json:"total_done"`
	TotalFailed  int `json:"total_failed"`
	TotalDead    int `json:"total_dead"`
}
