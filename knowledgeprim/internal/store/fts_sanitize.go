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
		// Already quoted — leave as-is.
		if len(tok) >= 2 && tok[0] == '"' && tok[len(tok)-1] == '"' {
			continue
		}

		needsQuote := keywords[tok]
		if !needsQuote {
			for _, ch := range tok {
				switch ch {
				case '-', ':', '*', '^', '(', ')', '{', '}':
					needsQuote = true
				}
				if needsQuote {
					break
				}
			}
		}

		if needsQuote {
			// Strip any embedded double quotes to prevent breaking out.
			clean := strings.ReplaceAll(tok, `"`, "")
			if clean != "" {
				tokens[i] = `"` + clean + `"`
			}
		}
	}

	return strings.Join(tokens, " ")
}
