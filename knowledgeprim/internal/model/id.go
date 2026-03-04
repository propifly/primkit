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
	// alphabet, which is more than sufficient for entity IDs.
	idLength = 12

	// idPrefix makes entity IDs visually distinguishable and greppable.
	// When you see "e_" in logs or scripts, you know it's a knowledge entity.
	idPrefix = "e_"
)

// NewEntityID generates a unique entity identifier like "e_a7x3k9m2p4qr".
func NewEntityID() string {
	id, err := gonanoid.Generate(idAlphabet, idLength)
	if err != nil {
		// gonanoid only errors if the alphabet is empty or length is negative,
		// neither of which can happen with our constants.
		panic("unexpected error generating entity ID: " + err.Error())
	}
	return idPrefix + id
}
