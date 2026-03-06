package docgen

import (
	"fmt"
	"strings"
)

// RenderCommandTable generates a markdown table for a prim's commands.
// Only the table rows and header are generated — no surrounding markdown.
func RenderCommandTable(meta PrimMeta) string {
	var sb strings.Builder
	sb.WriteString("| Command | Synopsis | Flags |\n")
	sb.WriteString("|---------|----------|-------|\n")
	for _, cmd := range meta.Commands {
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n",
			cmd.Name, cmd.Synopsis, renderFlags(cmd.Flags)))
	}
	// Trim trailing newline so the anchor suffix sits on its own line cleanly.
	return strings.TrimRight(sb.String(), "\n")
}

// renderFlags formats a slice of FlagMeta into a single markdown table cell string.
// Required flags are marked with *(required)*. Flags with non-trivial defaults
// show their default value. Flags are rendered in the order provided.
func renderFlags(flags []FlagMeta) string {
	if len(flags) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(flags))
	for _, f := range flags {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("`--%s`", f.Name))
		switch {
		case f.Required:
			sb.WriteString(" *(required)*")
		case isTrivialDefault(f.Default):
			// no default annotation
		default:
			sb.WriteString(fmt.Sprintf(" (default: `%s`)", f.Default))
		}
		if f.Usage != "" {
			sb.WriteString(fmt.Sprintf(" — %s", f.Usage))
		}
		parts = append(parts, sb.String())
	}
	return strings.Join(parts, "; ")
}

// isTrivialDefault returns true for default values that add no useful information
// to the documentation (empty string, false, 0).
func isTrivialDefault(d string) bool {
	return d == "" || d == "false" || d == "0"
}
