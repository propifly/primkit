// Package api implements the knowledgeprim HTTP API.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/propifly/primkit/knowledgeprim/internal/embed"
	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
)

// Handler holds the HTTP handlers for the knowledgeprim API.
type Handler struct {
	store    store.Store
	embedder embed.Embedder
	logger   *slog.Logger
}

// New creates an API handler.
func New(s store.Store, embedder embed.Embedder, logger *slog.Logger) *Handler {
	return &Handler{store: s, embedder: embedder, logger: logger}
}

// Router returns an http.Handler with all routes registered.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Entity CRUD.
	mux.HandleFunc("POST /v1/entities", h.captureEntity)
	mux.HandleFunc("GET /v1/entities/{id}", h.getEntity)
	mux.HandleFunc("PATCH /v1/entities/{id}", h.updateEntity)
	mux.HandleFunc("DELETE /v1/entities/{id}", h.deleteEntity)

	// Search.
	mux.HandleFunc("GET /v1/search", h.search)

	// Graph traversal.
	mux.HandleFunc("GET /v1/entities/{id}/related", h.related)

	// Edge operations.
	mux.HandleFunc("POST /v1/edges", h.createEdge)
	mux.HandleFunc("PATCH /v1/edges/{source}/{target}/{relationship}", h.updateEdge)
	mux.HandleFunc("POST /v1/edges/{source}/{target}/{relationship}/strengthen", h.strengthenEdge)
	mux.HandleFunc("DELETE /v1/edges/{source}/{target}/{relationship}", h.deleteEdge)

	// Discovery and aggregates.
	mux.HandleFunc("GET /v1/discover", h.discover)
	mux.HandleFunc("GET /v1/types", h.listTypes)
	mux.HandleFunc("GET /v1/relationships", h.listRelationships)
	mux.HandleFunc("GET /v1/stats", h.stats)

	return mux
}

// ---------------------------------------------------------------------------
// Entity endpoints
// ---------------------------------------------------------------------------

type captureRequest struct {
	Type       string           `json:"type"`
	Title      string           `json:"title"`
	Body       *string          `json:"body"`
	URL        *string          `json:"url"`
	Source     string           `json:"source"`
	Properties *json.RawMessage `json:"properties"`
}

func (h *Handler) captureEntity(w http.ResponseWriter, r *http.Request) {
	var req captureRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	source := req.Source
	if source == "" {
		source = auth.KeyNameFromContext(r.Context())
		if source == "" {
			source = "api"
		}
	}

	entity := &model.Entity{
		Type:   req.Type,
		Title:  req.Title,
		Body:   req.Body,
		URL:    req.URL,
		Source: source,
	}
	if req.Properties != nil {
		entity.Properties = *req.Properties
	}

	var embedding []float32
	if h.embedder != nil {
		var err error
		embedding, err = h.embedder.Embed(r.Context(), entity.EmbeddingText())
		if err != nil {
			h.logger.Warn("embedding failed", "error", err)
		}
	}

	if err := h.store.CaptureEntity(r.Context(), entity, embedding); err != nil {
		h.handleStoreError(w, "capturing entity", err)
		return
	}

	server.Respond(w, http.StatusCreated, entity)
}

func (h *Handler) getEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	entity, err := h.store.GetEntity(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "getting entity", err)
		return
	}

	server.Respond(w, http.StatusOK, entity)
}

type updateEntityRequest struct {
	Title      *string          `json:"title"`
	Body       *string          `json:"body"`
	URL        *string          `json:"url"`
	Properties *json.RawMessage `json:"properties"`
}

func (h *Handler) updateEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateEntityRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	update := &model.EntityUpdate{
		Title:      req.Title,
		Body:       req.Body,
		URL:        req.URL,
		Properties: req.Properties,
	}

	if err := h.store.UpdateEntity(r.Context(), id, update); err != nil {
		h.handleStoreError(w, "updating entity", err)
		return
	}

	entity, err := h.store.GetEntity(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "fetching entity", err)
		return
	}

	server.Respond(w, http.StatusOK, entity)
}

func (h *Handler) deleteEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.store.DeleteEntity(r.Context(), id); err != nil {
		h.handleStoreError(w, "deleting entity", err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("q")
	if query == "" {
		server.RespondError(w, http.StatusBadRequest, "QUERY_REQUIRED", "q parameter is required")
		return
	}

	filter := &model.SearchFilter{
		Type: q.Get("type"),
	}
	if limit := q.Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			filter.Limit = n
		}
	}

	mode := q.Get("mode")
	if mode == "" {
		mode = "hybrid"
	}

	var results []*model.SearchResult
	var err error

	switch mode {
	case "fts":
		results, err = h.store.SearchFTS(r.Context(), query, filter)
	case "vector":
		if h.embedder == nil {
			server.RespondError(w, http.StatusBadRequest, "NO_EMBEDDER", "vector search requires embedding provider")
			return
		}
		embedding, embErr := h.embedder.Embed(r.Context(), query)
		if embErr != nil {
			server.RespondError(w, http.StatusInternalServerError, "EMBEDDING_ERROR", embErr.Error())
			return
		}
		results, err = h.store.SearchVector(r.Context(), embedding, filter)
	default:
		var embedding []float32
		if h.embedder != nil {
			embedding, _ = h.embedder.Embed(r.Context(), query)
		}
		results, err = h.store.SearchHybrid(r.Context(), query, embedding, filter)
	}

	if err != nil {
		h.handleStoreError(w, "searching", err)
		return
	}

	server.Respond(w, http.StatusOK, results)
}

// ---------------------------------------------------------------------------
// Graph traversal
// ---------------------------------------------------------------------------

func (h *Handler) related(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	q := r.URL.Query()

	opts := &model.TraversalOpts{
		Depth:        1,
		Direction:    "both",
		Relationship: q.Get("relationship"),
	}
	if d := q.Get("depth"); d != "" {
		if n, err := strconv.Atoi(d); err == nil {
			opts.Depth = n
		}
	}
	if dir := q.Get("direction"); dir != "" {
		opts.Direction = dir
	}

	results, err := h.store.Related(r.Context(), id, opts)
	if err != nil {
		h.handleStoreError(w, "traversing", err)
		return
	}

	server.Respond(w, http.StatusOK, results)
}

// ---------------------------------------------------------------------------
// Edge endpoints
// ---------------------------------------------------------------------------

type createEdgeRequest struct {
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	Relationship string  `json:"relationship"`
	Weight       float64 `json:"weight"`
	Context      *string `json:"context"`
}

func (h *Handler) createEdge(w http.ResponseWriter, r *http.Request) {
	var req createEdgeRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	edge := &model.Edge{
		SourceID:     req.SourceID,
		TargetID:     req.TargetID,
		Relationship: req.Relationship,
		Weight:       req.Weight,
		Context:      req.Context,
	}

	if err := h.store.CreateEdge(r.Context(), edge); err != nil {
		h.handleStoreError(w, "creating edge", err)
		return
	}

	server.Respond(w, http.StatusCreated, edge)
}

type updateEdgeRequest struct {
	Context *string  `json:"context"`
	Weight  *float64 `json:"weight"`
}

func (h *Handler) updateEdge(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	target := r.PathValue("target")
	rel := r.PathValue("relationship")

	var req updateEdgeRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	update := &model.EdgeUpdate{
		Context: req.Context,
		Weight:  req.Weight,
	}

	if err := h.store.UpdateEdge(r.Context(), source, target, rel, update); err != nil {
		h.handleStoreError(w, "updating edge", err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) strengthenEdge(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	target := r.PathValue("target")
	rel := r.PathValue("relationship")

	if err := h.store.StrengthenEdge(r.Context(), source, target, rel); err != nil {
		h.handleStoreError(w, "strengthening edge", err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]string{"status": "strengthened"})
}

func (h *Handler) deleteEdge(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	target := r.PathValue("target")
	rel := r.PathValue("relationship")

	if err := h.store.DeleteEdge(r.Context(), source, target, rel); err != nil {
		h.handleStoreError(w, "deleting edge", err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---------------------------------------------------------------------------
// Discovery and aggregates
// ---------------------------------------------------------------------------

func (h *Handler) discover(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	opts := &model.DiscoverOpts{}

	if q.Get("orphans") == "true" {
		opts.Orphans = true
	}
	if q.Get("clusters") == "true" {
		opts.Clusters = true
	}
	if q.Get("bridges") == "true" {
		opts.Bridges = true
	}
	if q.Get("temporal") == "true" {
		opts.Temporal = true
	}
	if q.Get("weak_edges") == "true" {
		opts.WeakEdges = true
	}

	// If nothing specified, run all.
	if !opts.Orphans && !opts.Clusters && !opts.Bridges && !opts.Temporal && !opts.WeakEdges {
		opts = model.DiscoverAll()
	}

	report, err := h.store.Discover(r.Context(), opts)
	if err != nil {
		h.handleStoreError(w, "discovering", err)
		return
	}

	server.Respond(w, http.StatusOK, report)
}

func (h *Handler) listTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.store.ListTypes(r.Context())
	if err != nil {
		h.handleStoreError(w, "listing types", err)
		return
	}
	server.Respond(w, http.StatusOK, types)
}

func (h *Handler) listRelationships(w http.ResponseWriter, r *http.Request) {
	rels, err := h.store.ListRelationships(r.Context())
	if err != nil {
		h.handleStoreError(w, "listing relationships", err)
		return
	}
	server.Respond(w, http.StatusOK, rels)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.Stats(r.Context())
	if err != nil {
		h.handleStoreError(w, "fetching stats", err)
		return
	}
	server.Respond(w, http.StatusOK, stats)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *Handler) handleStoreError(w http.ResponseWriter, action string, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		server.RespondError(w, http.StatusNotFound, "NOT_FOUND", "entity not found")
	case errors.Is(err, store.ErrEdgeNotFound):
		server.RespondError(w, http.StatusNotFound, "EDGE_NOT_FOUND", "edge not found")
	case errors.Is(err, store.ErrEdgeExists):
		server.RespondError(w, http.StatusConflict, "EDGE_EXISTS", "edge already exists")
	case errors.Is(err, store.ErrInvalidEntity):
		server.RespondError(w, http.StatusBadRequest, "INVALID_ENTITY", err.Error())
	default:
		h.logger.Error(action, "error", err)
		server.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"an internal error occurred")
	}
}
