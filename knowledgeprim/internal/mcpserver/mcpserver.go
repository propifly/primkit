// Package mcpserver exposes knowledgeprim operations as MCP tools.
// 14 tools covering entities, edges, search, traversal, and discovery.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsvr "github.com/mark3labs/mcp-go/server"

	"github.com/propifly/primkit/knowledgeprim/internal/embed"
	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
)

// New creates an MCP server with all knowledgeprim tools registered.
func New(s store.Store, embedder embed.Embedder, version string) *mcpsvr.MCPServer {
	srv := mcpsvr.NewMCPServer(
		"knowledgeprim",
		version,
		mcpsvr.WithToolCapabilities(false),
	)

	registerCaptureTool(srv, s, embedder)
	registerSearchTool(srv, s, embedder)
	registerGetTool(srv, s)
	registerRelatedTool(srv, s)
	registerConnectTool(srv, s)
	registerStrengthenTool(srv, s)
	registerEdgeEditTool(srv, s)
	registerDisconnectTool(srv, s)
	registerEditTool(srv, s)
	registerDeleteTool(srv, s)
	registerDiscoverTool(srv, s)
	registerTypesTool(srv, s)
	registerRelationshipsTool(srv, s)
	registerStatsTool(srv, s)

	return srv
}

// ---------------------------------------------------------------------------
// Tool registration
// ---------------------------------------------------------------------------

func registerCaptureTool(srv *mcpsvr.MCPServer, s store.Store, embedder embed.Embedder) {
	tool := mcp.NewTool("knowledgeprim_capture",
		mcp.WithDescription("Capture a knowledge entity. Auto-embeds if configured."),
		mcp.WithString("type", mcp.Required(), mcp.Description("Entity type (article, thought, concept, pattern, etc.)")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Entity title")),
		mcp.WithString("body", mcp.Description("Entity body text")),
		mcp.WithString("url", mcp.Description("Source URL")),
		mcp.WithString("source", mcp.Description("Who captured this (defaults to 'mcp')")),
		mcp.WithString("properties", mcp.Description("JSON properties string")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		entity := &model.Entity{
			Type:   typ,
			Title:  title,
			Source: req.GetString("source", "mcp"),
		}
		if body := req.GetString("body", ""); body != "" {
			entity.Body = &body
		}
		if url := req.GetString("url", ""); url != "" {
			entity.URL = &url
		}
		if props := req.GetString("properties", ""); props != "" {
			entity.Properties = json.RawMessage(props)
		}

		var embedding []float32
		if embedder != nil {
			embedding, _ = embedder.Embed(ctx, entity.EmbeddingText())
		}

		if err := s.CaptureEntity(ctx, entity, embedding); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("capturing: %s", err)), nil
		}

		return jsonResult(entity)
	})
}

func registerSearchTool(srv *mcpsvr.MCPServer, s store.Store, embedder embed.Embedder) {
	tool := mcp.NewTool("knowledgeprim_search",
		mcp.WithDescription("Search the knowledge graph. Hybrid: FTS5 + vector + RRF."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("type", mcp.Description("Filter by entity type")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
		mcp.WithString("mode", mcp.Description("Search mode: hybrid, fts, vector"), mcp.Enum("hybrid", "fts", "vector")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		filter := &model.SearchFilter{
			Type:  req.GetString("type", ""),
			Limit: int(req.GetFloat("limit", 20)),
		}
		mode := req.GetString("mode", "hybrid")

		var results []*model.SearchResult

		switch mode {
		case "fts":
			results, err = s.SearchFTS(ctx, query, filter)
		case "vector":
			if embedder == nil {
				return mcp.NewToolResultError("vector search requires embedding provider"), nil
			}
			embedding, embErr := embedder.Embed(ctx, query)
			if embErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("embedding: %s", embErr)), nil
			}
			results, err = s.SearchVector(ctx, embedding, filter)
		default:
			var embedding []float32
			if embedder != nil {
				embedding, _ = embedder.Embed(ctx, query)
			}
			results, err = s.SearchHybrid(ctx, query, embedding, filter)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("searching: %s", err)), nil
		}

		return jsonResult(results)
	})
}

func registerGetTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_get",
		mcp.WithDescription("Get a single entity by ID with its edges."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID (e.g., e_abc123)")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		entity, err := s.GetEntity(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("getting entity: %s", err)), nil
		}
		return jsonResult(entity)
	})
}

func registerRelatedTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_related",
		mcp.WithDescription("Graph traversal from an entity."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Starting entity ID")),
		mcp.WithNumber("depth", mcp.Description("Traversal depth (default 1)")),
		mcp.WithString("relationship", mcp.Description("Filter by relationship type")),
		mcp.WithString("direction", mcp.Description("Edge direction: outgoing, incoming, both"), mcp.Enum("outgoing", "incoming", "both")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		opts := &model.TraversalOpts{
			Depth:        int(req.GetFloat("depth", 1)),
			Relationship: req.GetString("relationship", ""),
			Direction:    req.GetString("direction", "both"),
		}
		results, err := s.Related(ctx, id, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("traversing: %s", err)), nil
		}
		return jsonResult(results)
	})
}

func registerConnectTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_connect",
		mcp.WithDescription("Create an edge between two entities."),
		mcp.WithString("source_id", mcp.Required(), mcp.Description("Source entity ID")),
		mcp.WithString("target_id", mcp.Required(), mcp.Description("Target entity ID")),
		mcp.WithString("relationship", mcp.Required(), mcp.Description("Relationship type")),
		mcp.WithString("context", mcp.Description("Why this connection exists")),
		mcp.WithNumber("weight", mcp.Description("Edge weight (default 1.0)")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sourceID, err := req.RequireString("source_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		targetID, err := req.RequireString("target_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		rel, err := req.RequireString("relationship")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		edge := &model.Edge{
			SourceID:     sourceID,
			TargetID:     targetID,
			Relationship: rel,
			Weight:       req.GetFloat("weight", 1.0),
		}
		if c := req.GetString("context", ""); c != "" {
			edge.Context = &c
		}

		if err := s.CreateEdge(ctx, edge); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("connecting: %s", err)), nil
		}
		return jsonResult(edge)
	})
}

func registerStrengthenTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_strengthen",
		mcp.WithDescription("Increment an edge's weight."),
		mcp.WithString("source_id", mcp.Required(), mcp.Description("Source entity ID")),
		mcp.WithString("target_id", mcp.Required(), mcp.Description("Target entity ID")),
		mcp.WithString("relationship", mcp.Required(), mcp.Description("Relationship type")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		src, _ := req.RequireString("source_id")
		tgt, _ := req.RequireString("target_id")
		rel, _ := req.RequireString("relationship")
		if err := s.StrengthenEdge(ctx, src, tgt, rel); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("strengthening: %s", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Strengthened %s → %s [%s]", src, tgt, rel)), nil
	})
}

func registerEdgeEditTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_edge_edit",
		mcp.WithDescription("Update an edge's context or weight."),
		mcp.WithString("source_id", mcp.Required(), mcp.Description("Source entity ID")),
		mcp.WithString("target_id", mcp.Required(), mcp.Description("Target entity ID")),
		mcp.WithString("relationship", mcp.Required(), mcp.Description("Relationship type")),
		mcp.WithString("context", mcp.Description("New context")),
		mcp.WithNumber("weight", mcp.Description("New weight")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		src, _ := req.RequireString("source_id")
		tgt, _ := req.RequireString("target_id")
		rel, _ := req.RequireString("relationship")
		update := &model.EdgeUpdate{}
		if c := req.GetString("context", ""); c != "" {
			update.Context = &c
		}
		if w := req.GetFloat("weight", 0); w > 0 {
			update.Weight = &w
		}
		if err := s.UpdateEdge(ctx, src, tgt, rel, update); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("updating edge: %s", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Updated edge %s → %s [%s]", src, tgt, rel)), nil
	})
}

func registerDisconnectTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_disconnect",
		mcp.WithDescription("Remove an edge between two entities."),
		mcp.WithString("source_id", mcp.Required(), mcp.Description("Source entity ID")),
		mcp.WithString("target_id", mcp.Required(), mcp.Description("Target entity ID")),
		mcp.WithString("relationship", mcp.Required(), mcp.Description("Relationship type")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		src, _ := req.RequireString("source_id")
		tgt, _ := req.RequireString("target_id")
		rel, _ := req.RequireString("relationship")
		if err := s.DeleteEdge(ctx, src, tgt, rel); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("disconnecting: %s", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Disconnected %s → %s [%s]", src, tgt, rel)), nil
	})
}

func registerEditTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_edit",
		mcp.WithDescription("Update entity fields."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("body", mcp.Description("New body")),
		mcp.WithString("properties", mcp.Description("JSON properties")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		update := &model.EntityUpdate{}
		if t := req.GetString("title", ""); t != "" {
			update.Title = &t
		}
		if b := req.GetString("body", ""); b != "" {
			update.Body = &b
		}
		if p := req.GetString("properties", ""); p != "" {
			raw := json.RawMessage(p)
			update.Properties = &raw
		}
		if err := s.UpdateEntity(ctx, id, update); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("updating: %s", err)), nil
		}
		entity, err := s.GetEntity(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching: %s", err)), nil
		}
		return jsonResult(entity)
	})
}

func registerDeleteTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_delete",
		mcp.WithDescription("Delete an entity and its edges."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID to delete")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := s.DeleteEntity(ctx, id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("deleting: %s", err)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Deleted %s", id)), nil
	})
}

func registerDiscoverTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_discover",
		mcp.WithDescription("Run discovery: orphans, clusters, bridges, temporal, weak-edges."),
		mcp.WithBoolean("orphans", mcp.Description("Find entities with no edges")),
		mcp.WithBoolean("clusters", mcp.Description("Find densely connected groups")),
		mcp.WithBoolean("bridges", mcp.Description("Find cross-cluster connectors")),
		mcp.WithBoolean("temporal", mcp.Description("Show type distribution over time")),
		mcp.WithBoolean("weak_edges", mcp.Description("Find edges with no context")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		opts := &model.DiscoverOpts{
			Orphans:   req.GetBool("orphans", false),
			Clusters:  req.GetBool("clusters", false),
			Bridges:   req.GetBool("bridges", false),
			Temporal:  req.GetBool("temporal", false),
			WeakEdges: req.GetBool("weak_edges", false),
		}
		if !opts.Orphans && !opts.Clusters && !opts.Bridges && !opts.Temporal && !opts.WeakEdges {
			opts = model.DiscoverAll()
		}
		report, err := s.Discover(ctx, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("discovering: %s", err)), nil
		}
		return jsonResult(report)
	})
}

func registerTypesTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_types",
		mcp.WithDescription("List all entity types with counts."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		types, err := s.ListTypes(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing types: %s", err)), nil
		}
		return jsonResult(types)
	})
}

func registerRelationshipsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_relationships",
		mcp.WithDescription("List all relationship types with counts."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rels, err := s.ListRelationships(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing relationships: %s", err)), nil
		}
		return jsonResult(rels)
	})
}

func registerStatsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("knowledgeprim_stats",
		mcp.WithDescription("Get aggregate knowledge graph statistics."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := s.Stats(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching stats: %s", err)), nil
		}
		return jsonResult(stats)
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serializing: %s", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
