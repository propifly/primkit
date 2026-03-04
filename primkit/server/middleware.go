package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

// contextKey is an unexported type for context keys to prevent collisions
// with keys defined in other packages.
type contextKey string

const requestIDKey contextKey = "request_id"

// RequestIDFromContext extracts the request ID from the context, returning
// an empty string if none is set.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// Chain applies middleware in order. The first middleware listed wraps outermost,
// meaning it executes first on the request and last on the response.
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse order so the first middleware in the list wraps outermost.
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Logging logs each request's method, path, status code, and duration.
// The request ID (if present) is included for correlation.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter to capture the status code.
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", RequestIDFromContext(r.Context()),
			)
		})
	}
}

// Recovery catches panics in downstream handlers and returns a 500 error
// instead of crashing the server. The panic is logged with a stack trace.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"error", rec,
						"method", r.Method,
						"path", r.URL.Path,
					)
					RespondError(w, http.StatusInternalServerError,
						"INTERNAL_ERROR", "an internal error occurred")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID generates a unique ID for each request and adds it to the context
// and the X-Request-ID response header. If the request already has an
// X-Request-ID header, it is reused (for distributed tracing).
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = generateRequestID()
			}

			w.Header().Set("X-Request-ID", id)
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// generateRequestID creates a random 8-byte hex string (16 characters).
func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
