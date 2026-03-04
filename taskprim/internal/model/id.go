package model

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// idAlphabet uses lowercase alphanumeric characters only. This avoids
	// ambiguous characters (l/1, O/0) and keeps IDs URL-safe and easy to
	// select with a double-click in terminals.
	idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

	// idLength of 12 characters gives ~62 bits of entropy with a 36-character
	// alphabet, which is more than sufficient for task IDs (collision probability
	// is negligible at any realistic task volume).
	idLength = 12

	// idPrefix makes task IDs visually distinguishable and greppable.
	// When you see "t_" in logs or scripts, you know it's a task ID.
	idPrefix = "t_"
)

// NewTaskID generates a unique task identifier like "t_a7x3k9m2p4qr".
func NewTaskID() string {
	id, err := gonanoid.Generate(idAlphabet, idLength)
	if err != nil {
		// gonanoid only errors if the alphabet is empty or length is negative,
		// neither of which can happen with our constants.
		panic("unexpected error generating task ID: " + err.Error())
	}
	return idPrefix + id
}
