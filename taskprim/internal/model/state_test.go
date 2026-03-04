package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanTransition_ValidPaths(t *testing.T) {
	tests := []struct {
		from State
		to   State
	}{
		{StateOpen, StateDone},
		{StateOpen, StateKilled},
	}

	for _, tt := range tests {
		assert.True(t, CanTransition(tt.from, tt.to),
			"%s → %s should be valid", tt.from, tt.to)
	}
}

func TestCanTransition_InvalidPaths(t *testing.T) {
	tests := []struct {
		from State
		to   State
		desc string
	}{
		{StateDone, StateKilled, "done → killed: different terminal states don't interconvert"},
		{StateKilled, StateDone, "killed → done: different terminal states don't interconvert"},
		{StateDone, StateOpen, "done → open: not a normal transition (use restore)"},
		{StateKilled, StateOpen, "killed → open: not a normal transition (use restore)"},
	}

	for _, tt := range tests {
		assert.False(t, CanTransition(tt.from, tt.to),
			"%s → %s should be invalid: %s", tt.from, tt.to, tt.desc)
	}
}

func TestTransition_SameState(t *testing.T) {
	err := Transition(StateOpen, StateOpen)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already open")
}

func TestTransition_ValidReturnsNil(t *testing.T) {
	assert.NoError(t, Transition(StateOpen, StateDone))
	assert.NoError(t, Transition(StateOpen, StateKilled))
}

func TestTransition_InvalidReturnsError(t *testing.T) {
	err := Transition(StateDone, StateKilled)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transition")
}

func TestIsTerminal(t *testing.T) {
	assert.False(t, IsTerminal(StateOpen), "open is not terminal")
	assert.True(t, IsTerminal(StateDone), "done is terminal")
	assert.True(t, IsTerminal(StateKilled), "killed is terminal")
}
