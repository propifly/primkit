package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
	"github.com/propifly/primkit/taskprim/internal/api"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Long: `Starts taskprim in HTTP server mode. The API is served under /v1/ and
requires Bearer token authentication when auth keys are configured.

When replication is enabled in the config, Litestream continuously streams
WAL frames to the configured object storage (S3, R2, B2).

  taskprim serve --port 8090
  taskprim serve --port 8090 --config /etc/taskprim/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			cfg := configFromContext(cmd.Context())
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			// Override port from config if not explicitly set on the command line.
			if !cmd.Flags().Changed("port") && cfg.Server.Port != 0 {
				port = cfg.Server.Port
			}

			// Build auth validator from configured keys.
			var keys []auth.KeyEntry
			for _, k := range cfg.Auth.Keys {
				keys = append(keys, auth.KeyEntry{
					Key:  k.Key,
					Name: k.Name,
				})
			}
			validator := auth.NewValidator(keys)

			// Create the API handler and apply middleware.
			handler := api.New(s, logger)
			wrapped := server.Chain(
				handler.Router(),
				server.RequestID(),
				server.Logging(logger),
				server.Recovery(logger),
				validator.Middleware(),
			)

			// Start server with signal-based graceful shutdown.
			ctx, cancel := signal.NotifyContext(
				context.Background(),
				syscall.SIGINT, syscall.SIGTERM,
			)
			defer cancel()

			srv := server.New(port, wrapped, logger)
			logger.Info("taskprim serve starting", "port", port)
			return srv.ListenAndServe(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8090, "HTTP port to listen on")

	return cmd
}
