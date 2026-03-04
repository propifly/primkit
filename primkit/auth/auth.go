// Package auth provides API key validation middleware for primkit primitives.
//
// In CLI mode, no authentication is needed — the filesystem is the auth layer.
// In serve and MCP SSE modes, every request must include a valid API key in the
// Authorization header: "Authorization: Bearer <key>".
//
// Keys are configured in config.yaml and mapped to human-readable names. The name
// is used as the request source (e.g., "johanna" created this task via the API).
package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/propifly/primkit/primkit/server"
)

// contextKey is unexported to prevent collisions with other packages' context keys.
type contextKey string

const keyNameKey contextKey = "auth_key_name"

// KeyNameFromContext returns the name associated with the authenticated API key.
// Returns empty string if no auth was performed (CLI mode).
func KeyNameFromContext(ctx context.Context) string {
	if name, ok := ctx.Value(keyNameKey).(string); ok {
		return name
	}
	return ""
}

// KeyEntry maps an API key to its metadata.
type KeyEntry struct {
	Key  string
	Name string
}

// Validator checks API keys against a configured set. When no keys are configured,
// all requests pass through (open mode for local development).
type Validator struct {
	keys []KeyEntry
}

// NewValidator creates a validator from a list of key entries. If keys is empty,
// the validator operates in open mode (all requests allowed).
func NewValidator(keys []KeyEntry) *Validator {
	return &Validator{keys: keys}
}

// Middleware returns HTTP middleware that validates the Authorization header.
//
// If no keys are configured, all requests are allowed through (open mode). This
// makes local CLI-only usage frictionless — just don't configure any keys.
//
// When keys are configured, the middleware:
//  1. Extracts the Bearer token from the Authorization header
//  2. Compares it against all configured keys using constant-time comparison
//  3. If matched, adds the key's name to the request context
//  4. If not matched, returns 401 Unauthorized
func (v *Validator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Open mode: no keys configured means no auth required.
			if len(v.keys) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			token := extractBearerToken(r)
			if token == "" {
				server.RespondError(w, http.StatusUnauthorized,
					"AUTH_REQUIRED", "missing Authorization header")
				return
			}

			entry, ok := v.validateKey(token)
			if !ok {
				server.RespondError(w, http.StatusUnauthorized,
					"INVALID_KEY", "invalid API key")
				return
			}

			// Stash the key name in context so handlers can identify the caller.
			ctx := context.WithValue(r.Context(), keyNameKey, entry.Name)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken pulls the token from "Authorization: Bearer <token>".
// Returns empty string if the header is missing or malformed.
func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// validateKey checks a token against all configured keys using constant-time
// comparison to prevent timing attacks. Returns the matching entry and true
// if found, or zero value and false if not.
func (v *Validator) validateKey(token string) (KeyEntry, bool) {
	for _, entry := range v.keys {
		if subtle.ConstantTimeCompare([]byte(token), []byte(entry.Key)) == 1 {
			return entry, true
		}
	}
	return KeyEntry{}, false
}
