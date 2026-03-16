package model

import (
	"fmt"
	"time"
)

// Task is the central domain object. Every field maps directly to the SQLite
// schema. A task belongs to exactly one list, can have zero or more labels,
// and follows a simple lifecycle: open → done | killed.
type Task struct {
	ID             string     `json:"id"`
	List           string     `json:"list"`
	What           string     `json:"what"`
	Source         string     `json:"source"`
	State          State      `json:"state"`
	WaitingOn      *string    `json:"waiting_on,omitempty"`
	ParentID       *string    `json:"parent_id,omitempty"`
	Context        *string    `json:"context,omitempty"`
	Labels         []string   `json:"labels,omitempty"`
	Created        time.Time  `json:"created"`
	Updated        time.Time  `json:"updated"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolvedReason *string    `json:"resolved_reason,omitempty"`
}

// Validate checks that a task has all required fields before creation.
// It does not check the ID field — that's assigned by the store.
func (t *Task) Validate() error {
	if t.List == "" {
		return fmt.Errorf("list is required")
	}
	if t.What == "" {
		return fmt.Errorf("what is required")
	}
	if t.Source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

// TaskUpdate holds the fields that can be changed via the edit command.
// All fields are optional — nil means "don't change this field."
type TaskUpdate struct {
	What      *string  `json:"what,omitempty"`
	List      *string  `json:"list,omitempty"`
	WaitingOn *string  `json:"waiting_on,omitempty"`
	Context   *string  `json:"context,omitempty"`
	ParentID  *string  `json:"parent_id,omitempty"`
	AddLabels []string `json:"add_labels,omitempty"`
	DelLabels []string `json:"del_labels,omitempty"`
}

// ClearWaitingOn is a sentinel value for TaskUpdate.WaitingOn to clear the field.
// Use: update.WaitingOn = &model.ClearWaitingOn
var ClearWaitingOn = ""

// Filter specifies which tasks to return from ListTasks. All fields are optional
// — an empty filter returns all tasks.
type Filter struct {
	List     string   // Only tasks in this list.
	State    *State   // Only tasks in this state (nil = all states).
	Labels   []string // Tasks that have ALL of these labels.
	Source   string   // Only tasks from this source.
	ParentID *string  // Only subtasks of this parent (nil = no filter, empty string = root tasks only).
	Waiting  *bool    // true = only tasks with waiting_on set, false = only without.

	// Seen-tracking filters. These are mutually exclusive.
	UnseenBy string        // Tasks not seen by this agent (or updated since last seen).
	SeenBy   string        // Tasks seen by this agent...
	Since    time.Duration // ...within this duration (used with SeenBy).

	// Stale filter: tasks not updated within this duration.
	Stale time.Duration
}

// LabelCount pairs a label with the number of open tasks that carry it.
type LabelCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// ListInfo provides summary counts for a single list.
type ListInfo struct {
	Name   string `json:"name"`
	Open   int    `json:"open"`
	Done   int    `json:"done"`
	Killed int    `json:"killed"`
}

// Stats provides aggregate counts across all tasks.
type Stats struct {
	TotalOpen   int `json:"total_open"`
	TotalDone   int `json:"total_done"`
	TotalKilled int `json:"total_killed"`
}

// DepEdge represents a single dependency relationship: TaskID depends on DependsOn.
type DepEdge struct {
	TaskID    string `json:"task_id"`
	DependsOn string `json:"depends_on"`
}
