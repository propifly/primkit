package model

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// idAlphabet uses lowercase alphanumeric characters only. Avoids ambiguous
	// characters (l/1, O/0) and keeps IDs URL-safe and double-click selectable.
	idAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

	// idLength of 12 characters gives ~62 bits of entropy — negligible collision
	// probability at any realistic queue volume.
	idLength = 12

	// idPrefix makes job IDs visually distinguishable and greppable.
	// "q_" in logs or scripts immediately identifies a queueprim job ID.
	idPrefix = "q_"
)

// NewJobID generates a unique job identifier like "q_a7x3k9m2p4qr".
func NewJobID() string {
	id, err := gonanoid.Generate(idAlphabet, idLength)
	if err != nil {
		// gonanoid only errors if the alphabet is empty or length is negative,
		// neither of which can happen with our constants.
		panic("unexpected error generating job ID: " + err.Error())
	}
	return idPrefix + id
}
