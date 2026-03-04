// Package mcpserver exposes stateprim operations as MCP tools. Any
// MCP-compatible client (Claude Desktop, Cursor, etc.) can discover and
// call these tools via the Model Context Protocol.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsvr "github.com/mark3labs/mcp-go/server"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/propifly/primkit/stateprim/internal/store"
)

// New creates an MCP server with all stateprim tools registered.
func New(s store.Store, version string) *mcpsvr.MCPServer {
	srv := mcpsvr.NewMCPServer(
		"stateprim",
		version,
		mcpsvr.WithToolCapabilities(true),
	)

	registerSetTool(srv, s)
	registerGetTool(srv, s)
	registerHasTool(srv, s)
	registerSetIfNewTool(srv, s)
	registerAppendTool(srv, s)
	registerDeleteTool(srv, s)
	registerQueryTool(srv, s)
	registerPurgeTool(srv, s)
	registerNamespacesTool(srv, s)
	registerStatsTool(srv, s)

	return srv
}

func registerSetTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_set",
		mcp.WithDescription("Create or update a record (upsert). Returns ErrImmutable if the existing record is immutable."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace for the record")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key within the namespace")),
		mcp.WithString("value", mcp.Required(), mcp.Description("JSON value")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		val, err := req.RequireString("value")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		r := &model.Record{
			Namespace: ns,
			Key:       key,
			Value:     json.RawMessage(val),
		}
		if err := s.Set(ctx, r); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(r)
	})
}

func registerGetTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_get",
		mcp.WithDescription("Retrieve a record by namespace and key."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		r, err := s.Get(ctx, ns, key)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(r)
	})
}

func registerHasTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_has",
		mcp.WithDescription("Check if a key exists in a namespace."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		exists, err := s.Has(ctx, ns, key)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]bool{"exists": exists})
	})
}

func registerSetIfNewTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_set_if_new",
		mcp.WithDescription("Create a record only if the key doesn't already exist. Returns error if key exists."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key")),
		mcp.WithString("value", mcp.Required(), mcp.Description("JSON value")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		val, err := req.RequireString("value")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		r := &model.Record{
			Namespace: ns,
			Key:       key,
			Value:     json.RawMessage(val),
		}
		if err := s.SetIfNew(ctx, r); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(r)
	})
}

func registerAppendTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_append",
		mcp.WithDescription("Append an immutable record with an auto-generated key."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("value", mcp.Required(), mcp.Description("JSON value")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		val, err := req.RequireString("value")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		r, err := s.Append(ctx, ns, []byte(val))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(r)
	})
}

func registerDeleteTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_delete",
		mcp.WithDescription("Delete a record by namespace and key."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Key")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = s.Delete(ctx, ns, key)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]string{"status": "ok"})
	})
}

func registerQueryTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_query",
		mcp.WithDescription("Query records in a namespace with optional filters."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("prefix", mcp.Description("Filter by key prefix")),
		mcp.WithString("since", mcp.Description("Only records updated within duration (e.g., 24h, 7d)")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		filter := &model.QueryFilter{
			Namespace: ns,
			KeyPrefix: req.GetString("prefix", ""),
		}

		if since := req.GetString("since", ""); since != "" {
			d, err := parseDuration(since)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid since: %v", err)), nil
			}
			filter.Since = d
		}

		records, _, err := s.Query(ctx, filter)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(records)
	})
}

func registerPurgeTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_purge",
		mcp.WithDescription("Delete records older than a duration in a namespace."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("older_than", mcp.Required(), mcp.Description("Duration (e.g., 24h, 7d)")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns, err := req.RequireString("namespace")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		olderThan, err := req.RequireString("older_than")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		deleted, err := s.Purge(ctx, ns, olderThan)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]int{"deleted": deleted})
	})
}

func registerNamespacesTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_namespaces",
		mcp.WithDescription("List all namespaces with record counts."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nss, err := s.ListNamespaces(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(nss)
	})
}

func registerStatsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("stateprim_stats",
		mcp.WithDescription("Show aggregate record statistics."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := s.Stats(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(stats)
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// jsonResult serializes a value to indented JSON and wraps it in a tool result.
func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling result: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// parseDuration handles both Go-style durations and day notation (7d, 30d).
func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
