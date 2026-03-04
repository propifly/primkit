package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/propifly/primkit/knowledgeprim/internal/api"
	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		Long: `Starts knowledgeprim in HTTP server mode.

  knowledgeprim serve --port 8092
  knowledgeprim serve --port 8092 --config config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			embedder := embedderFromContext(cmd.Context())
			cfg := configFromContext(cmd.Context())
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			if !cmd.Flags().Changed("port") && cfg.Server.Port != 0 {
				port = cfg.Server.Port
			}

			var keys []auth.KeyEntry
			for _, k := range cfg.Auth.Keys {
				keys = append(keys, auth.KeyEntry{
					Key:  k.Key,
					Name: k.Name,
				})
			}
			validator := auth.NewValidator(keys)

			handler := api.New(s, embedder, logger)
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

			srv := server.New(port, wrapped, logger)
			logger.Info("knowledgeprim serve starting", "port", port)
			return srv.ListenAndServe(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8092, "HTTP port")

	return cmd
}
