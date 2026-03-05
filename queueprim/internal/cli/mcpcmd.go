package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	mcpsvr "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/propifly/primkit/queueprim/internal/mcpserver"
)

func newMCPCmd() *cobra.Command {
	var (
		transport string
		port      int
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP (Model Context Protocol) server",
		Long: `Starts queueprim as an MCP server, exposing all queue operations as tools
that any MCP-compatible client can call (Claude Desktop, Cursor, agents).

Two transports:
  stdio — for local use (no auth, agent runs on same machine)
  sse   — for remote use (auth required, agent connects over HTTP)

  queueprim mcp --transport stdio
  queueprim mcp --transport sse --port 8094`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

			srv := mcpserver.New(s, "0.1.0")

			// Start sweeper for serve/mcp long-running modes.
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			go runSweeper(ctx, s, logger)

			switch transport {
			case "stdio":
				fmt.Fprintln(cmd.OutOrStdout(), "Starting MCP server (stdio transport)...")
				return mcpsvr.ServeStdio(srv)
			case "sse":
				addr := fmt.Sprintf(":%d", port)
				fmt.Fprintf(cmd.OutOrStdout(), "Starting MCP server (SSE transport) on %s...\n", addr)
				return mcpsvr.NewSSEServer(srv).Start(addr)
			default:
				return fmt.Errorf("unknown transport %q (use stdio or sse)", transport)
			}
		},
	}

	cmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "transport: stdio or sse")
	cmd.Flags().IntVarP(&port, "port", "p", 8094, "port for SSE transport")

	return cmd
}
