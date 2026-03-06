// Package docgen provides types and utilities for generating primkit documentation
// from Cobra command metadata. This package has no external dependencies beyond
// the standard library — Cobra-specific extraction lives in each prim's cmd/docgen.
package docgen

// PrimMeta holds extracted metadata for one primitive.
type PrimMeta struct {
	Name     string    `json:"name"`
	Commands []CmdMeta `json:"commands"`
}

// CmdMeta holds metadata for a single command (possibly nested, e.g. "label clear").
type CmdMeta struct {
	Name     string     `json:"name"`     // full command path, e.g. "label clear"
	Synopsis string     `json:"synopsis"` // full usage string, e.g. "label clear <id> <label>"
	Short    string     `json:"short"`
	Flags    []FlagMeta `json:"flags"`
}

// FlagMeta holds metadata for a single flag.
type FlagMeta struct {
	Name     string `json:"name"`
	Usage    string `json:"usage"`
	Default  string `json:"default"`
	Required bool   `json:"required"` // true if MarkFlagRequired was called
}
