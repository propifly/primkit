package model

import "fmt"

// State represents the lifecycle state of a task. The state machine is
// intentionally simple: tasks start open, and can be completed (done) or
// dropped (killed). Both terminal states are immutable — once resolved,
// a task's state cannot change (except via explicit restore).
type State string

const (
	StateOpen   State = "open"
	StateDone   State = "done"
	StateKilled State = "killed"
)

// validTransitions defines the allowed state transitions. Each entry maps a
// source state to the set of states it can transition to.
//
//	open → done    (task completed)
//	open → killed  (task dropped, requires a reason)
//
// There is no done→killed or killed→done — those are different exit paths,
// not a progression. Restoring a resolved task back to open is handled
// separately (not a state machine transition).
var validTransitions = map[State][]State{
	StateOpen: {StateDone, StateKilled},
}

// CanTransition reports whether moving from the current state to the target
// state is a valid transition.
func CanTransition(from, to State) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// Transition validates a state transition and returns an error if it's not
// allowed. The error message explains what went wrong in human-readable terms.
func Transition(from, to State) error {
	if from == to {
		return fmt.Errorf("task is already %s", from)
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("cannot transition from %s to %s", from, to)
	}
	return nil
}

// IsTerminal reports whether a state is terminal (done or killed).
// Terminal states cannot be changed through normal state transitions.
func IsTerminal(s State) bool {
	return s == StateDone || s == StateKilled
}
