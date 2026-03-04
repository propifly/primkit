package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Server wraps the standard http.Server with graceful shutdown and logging.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New creates a Server that listens on the given port with the provided handler.
// The handler should already have middleware applied via Chain().
func New(port int, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

// ListenAndServe starts the HTTP server and blocks until the context is cancelled.
// When the context is cancelled, it initiates a graceful shutdown, giving in-flight
// requests up to 10 seconds to complete before forcefully closing.
func (s *Server) ListenAndServe(ctx context.Context) error {
	// Start the server in a goroutine so we can listen for context cancellation.
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for either the context to be cancelled or the server to fail.
	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		s.logger.Info("server stopped")
		return nil
	}
}
