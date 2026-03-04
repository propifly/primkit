package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskID_Format(t *testing.T) {
	id := NewTaskID()

	assert.True(t, strings.HasPrefix(id, "t_"),
		"task ID should start with t_ prefix, got: %s", id)

	// Total length: 2 (prefix) + 12 (random) = 14
	assert.Len(t, id, 14, "task ID should be 14 characters total")
}

func TestNewTaskID_AlphanumericOnly(t *testing.T) {
	id := NewTaskID()
	suffix := id[2:] // Strip the "t_" prefix.

	for _, c := range suffix {
		assert.True(t,
			(c >= '0' && c <= '9') || (c >= 'a' && c <= 'z'),
			"task ID should only contain lowercase alphanumeric chars, got: %c", c)
	}
}

func TestNewTaskID_Unique(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := NewTaskID()
		assert.False(t, seen[id], "generated duplicate ID: %s", id)
		seen[id] = true
	}
}
