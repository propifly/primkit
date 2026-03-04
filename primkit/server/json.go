// Package server provides a lightweight HTTP server scaffold for primkit
// primitives. It includes JSON helpers, middleware, and graceful shutdown.
//
// Both taskprim and stateprim use this package for their serve mode.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse is the standard JSON error format returned by all API endpoints.
// Machine-readable code + human-readable message.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Decode reads a JSON request body into the target struct. It enforces that the
// body is valid JSON and rejects unknown fields (to catch typos in field names).
// Returns an error suitable for display to the client.
func Decode(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// Respond writes a JSON response with the given HTTP status code.
// Sets Content-Type to application/json automatically.
func Respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	}
}

// RespondError writes a JSON error response. Use this instead of http.Error
// to maintain a consistent error format across all endpoints.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	Respond(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}
