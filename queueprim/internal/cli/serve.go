package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
	"github.com/propifly/primkit/queueprim/internal/api"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Long: `Starts queueprim in HTTP server mode. The API is served under /v1/ and
requires Bearer token authentication when auth keys are configured.

A background sweeper runs every minute to release expired claims.

  queueprim serve --port 8093
  queueprim serve --port 8093 --config /etc/queueprim/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			cfg := configFromContext(cmd.Context())
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			if !cmd.Flags().Changed("port") && cfg.Server.Port != 0 {
				port = cfg.Server.Port
			}

			var keys []auth.KeyEntry
			for _, k := range cfg.Auth.Keys {
				keys = append(keys, auth.KeyEntry{Key: k.Key, Name: k.Name})
			}
			validator := auth.NewValidator(keys)

			handler := api.New(s, logger)
			wrapped := server.Chain(
				handler.Router(),
				server.RequestID(),
				server.Logging(logger),
				server.Recovery(logger),
				validator.Middleware(),
			)

			ctx, cancel := signal.NotifyContext(
				context.Background(),
				syscall.SIGINT, syscall.SIGTERM,
			)
			defer cancel()

			// Start background sweeper for expired claims.
			go runSweeper(ctx, s, logger)

			srv := server.New(port, wrapped, logger)
			logger.Info("queueprim serve starting", "port", port)
			return srv.ListenAndServe(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8093, "HTTP port to listen on")

	return cmd
}

// runSweeper runs the expired-claim sweeper every minute until ctx is done.
func runSweeper(ctx context.Context, s interface {
	SweepExpiredClaims(context.Context) (int, error)
}, logger *slog.Logger,
) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := s.SweepExpiredClaims(ctx)
			if err != nil {
				logger.Error("sweeper error", "error", err)
			} else if n > 0 {
				logger.Info("sweeper released expired claims", "count", n)
			}
		}
	}
}
