package docgen

import (
	"fmt"
	"os"
	"regexp"
)

const (
	anchorStartFmt = "<!-- docgen:start:%s:commands -->"
	anchorEndFmt   = "<!-- docgen:end:%s:commands -->"
)

// UpdateDoc reads docPath, replaces anchored command tables for each prim in metas,
// and writes the result back. If check is true it reports drift without writing.
func UpdateDoc(docPath string, metas []PrimMeta, check bool) error {
	raw, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", docPath, err)
	}

	content := string(raw)
	updated := content

	for _, meta := range metas {
		table := RenderCommandTable(meta)
		updated, err = ReplaceAnchored(updated, meta.Name, table)
		if err != nil {
			return fmt.Errorf("updating %s section: %w", meta.Name, err)
		}
	}

	if check {
		if updated != content {
			return fmt.Errorf("docs are out of date — run 'make docs' to regenerate")
		}
		return nil
	}

	return os.WriteFile(docPath, []byte(updated), 0o644)
}

// ReplaceAnchored replaces the content between the docgen anchor comments for primName.
// Returns an error if the start anchor is not found in content.
func ReplaceAnchored(content, primName, newContent string) (string, error) {
	start := fmt.Sprintf(anchorStartFmt, primName)
	end := fmt.Sprintf(anchorEndFmt, primName)

	pattern := regexp.MustCompile(
		regexp.QuoteMeta(start) + `[\s\S]*?` + regexp.QuoteMeta(end),
	)

	if !pattern.MatchString(content) {
		return "", fmt.Errorf("anchor %q not found in document", start)
	}

	replacement := start + "\n" + newContent + "\n" + end
	return pattern.ReplaceAllString(content, replacement), nil
}
