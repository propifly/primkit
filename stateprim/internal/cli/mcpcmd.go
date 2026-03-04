package cli

import (
	"fmt"

	mcpsvr "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/propifly/primkit/stateprim/internal/mcpserver"
)

func newMCPCmd() *cobra.Command {
	var (
		transport string
		port      int
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP (Model Context Protocol) server",
		Long: `Starts stateprim as an MCP server, exposing all state operations as tools
that any MCP-compatible client can call (Claude Desktop, Cursor, etc.).

Two transports:
  stdio — for local use (no auth, agent runs on same machine)
  sse   — for remote use (auth required, agent connects over HTTP)

  stateprim mcp --transport stdio
  stateprim mcp --transport sse --port 8092`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			srv := mcpserver.New(s, "0.1.0")

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
	cmd.Flags().IntVarP(&port, "port", "p", 8092, "port for SSE transport")

	return cmd
}
