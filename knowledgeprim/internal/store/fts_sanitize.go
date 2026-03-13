package store

import "strings"

// sanitizeFTS5Query escapes user input so it is safe for SQLite FTS5 MATCH.
// Tokens containing special FTS5 characters (hyphens, colons, etc.) are
// wrapped in double quotes. FTS5 keywords (AND, OR, NOT, NEAR) are also
// quoted to prevent them from being interpreted as operators.
func sanitizeFTS5Query(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// FTS5 keywords that must be quoted when used as search terms.
	keywords := map[string]bool{
		"AND":  true,
		"OR":   true,
		"NOT":  true,
		"NEAR": true,
	}

	tokens := strings.Fields(query)
	for i, tok := range tokens {
		needsQuote := keywords[tok]
		if !needsQuote {
			for _, ch := range tok {
				// Quote any token with non-alphanumeric characters. FTS5 only
				// guarantees that letters, digits, and underscores are safe in
				// bare (unquoted) tokens.
				if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') &&
					(ch < '0' || ch > '9') && ch != '_' {
					needsQuote = true
					break
				}
			}
		}

		if needsQuote {
			// Strip double quotes (prevent breaking out) and control
			// characters (null bytes etc. cause FTS5 unterminated string
			// errors even inside quotes).
			clean := strings.Map(func(r rune) rune {
				if r == '"' || r < 0x20 {
					return -1 // drop
				}
				return r
			}, tok)
			if clean != "" {
				tokens[i] = `"` + clean + `"`
			} else {
				// Token was nothing but quotes/control chars — drop it.
				tokens[i] = ""
			}
		}
	}

	// Rebuild, dropping any empty tokens from quote-only inputs.
	var out []string
	for _, tok := range tokens {
		if tok != "" {
			out = append(out, tok)
		}
	}
	return strings.Join(out, " ")
}
