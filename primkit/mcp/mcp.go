// Package mcp provides a scaffold for MCP (Model Context Protocol) servers.
// This is a thin wrapper that will be fleshed out in Phase 6 (taskprim MCP)
// and Phase 11 (stateprim MCP).
//
// MCP allows AI agents to discover and call tools exposed by primkit primitives.
// Two transports are supported:
//   - stdio: for local use (no auth, agent runs on same machine)
//   - SSE: for remote use (auth required, agent connects over HTTP)
package mcp

// Server is a placeholder for the MCP server implementation.
// It will wrap mcp-go and provide tool registration for each primitive.
type Server struct {
	name    string
	version string
}

// New creates an MCP server with the given name and version.
// The name appears in the MCP handshake (e.g., "taskprim", "stateprim").
func New(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
	}
}
