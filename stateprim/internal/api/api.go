// Package api implements the stateprim HTTP API. All endpoints live under /v1/
// and operate on the shared Store interface.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/propifly/primkit/primkit/server"
	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/propifly/primkit/stateprim/internal/store"
)

// Handler holds the dependencies for all HTTP handlers.
type Handler struct {
	store  store.Store
	logger *slog.Logger
}

// New creates a new API handler.
func New(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// Router returns an http.Handler with all routes registered.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Record operations.
	mux.HandleFunc("POST /v1/records", h.handleSet)
	mux.HandleFunc("GET /v1/records/{namespace}/{key}", h.handleGet)
	mux.HandleFunc("DELETE /v1/records/{namespace}/{key}", h.handleDelete)
	mux.HandleFunc("POST /v1/records/{namespace}/set-if-new", h.handleSetIfNew)
	mux.HandleFunc("POST /v1/records/{namespace}/append", h.handleAppend)
	mux.HandleFunc("GET /v1/records/{namespace}/has/{key}", h.handleHas)

	// Query.
	mux.HandleFunc("GET /v1/records/{namespace}", h.handleQuery)

	// Lifecycle.
	mux.HandleFunc("POST /v1/records/{namespace}/purge", h.handlePurge)
	mux.HandleFunc("GET /v1/namespaces", h.handleNamespaces)
	mux.HandleFunc("GET /v1/stats", h.handleStats)

	// Import/Export.
	mux.HandleFunc("GET /v1/export", h.handleExport)
	mux.HandleFunc("POST /v1/import", h.handleImport)

	return mux
}

// ---------------------------------------------------------------------------
// Record operations
// ---------------------------------------------------------------------------

func (h *Handler) handleSet(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Namespace string          `json:"namespace"`
		Key       string          `json:"key"`
		Value     json.RawMessage `json:"value"`
		Immutable bool            `json:"immutable"`
	}
	if err := server.Decode(r, &body); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	rec := &model.Record{
		Namespace: body.Namespace,
		Key:       body.Key,
		Value:     body.Value,
		Immutable: body.Immutable,
	}
	if err := h.store.Set(r.Context(), rec); err != nil {
		h.handleStoreError(w, err)
		return
	}

	server.Respond(w, http.StatusOK, rec)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	rec, err := h.store.Get(r.Context(), r.PathValue("namespace"), r.PathValue("key"))
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, rec)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	err := h.store.Delete(r.Context(), r.PathValue("namespace"), r.PathValue("key"))
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleSetIfNew(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := server.Decode(r, &body); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	rec := &model.Record{
		Namespace: r.PathValue("namespace"),
		Key:       body.Key,
		Value:     body.Value,
	}
	if err := h.store.SetIfNew(r.Context(), rec); err != nil {
		h.handleStoreError(w, err)
		return
	}

	server.Respond(w, http.StatusCreated, rec)
}

func (h *Handler) handleAppend(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Value json.RawMessage `json:"value"`
	}
	if err := server.Decode(r, &body); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	rec, err := h.store.Append(r.Context(), r.PathValue("namespace"), body.Value)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	server.Respond(w, http.StatusCreated, rec)
}

func (h *Handler) handleHas(w http.ResponseWriter, r *http.Request) {
	exists, err := h.store.Has(r.Context(), r.PathValue("namespace"), r.PathValue("key"))
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, map[string]bool{"exists": exists})
}

// ---------------------------------------------------------------------------
// Query
// ---------------------------------------------------------------------------

func (h *Handler) handleQuery(w http.ResponseWriter, r *http.Request) {
	filter := &model.QueryFilter{
		Namespace: r.PathValue("namespace"),
		KeyPrefix: r.URL.Query().Get("prefix"),
		CountOnly: r.URL.Query().Get("count_only") == "true",
	}

	if since := r.URL.Query().Get("since"); since != "" {
		d, err := parseDuration(since)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_DURATION", err.Error())
			return
		}
		filter.Since = d
	}

	records, count, err := h.store.Query(r.Context(), filter)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	if filter.CountOnly {
		server.Respond(w, http.StatusOK, map[string]int{"count": count})
		return
	}

	server.Respond(w, http.StatusOK, records)
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

func (h *Handler) handlePurge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OlderThan string `json:"older_than"`
	}
	if err := server.Decode(r, &body); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	deleted, err := h.store.Purge(r.Context(), r.PathValue("namespace"), body.OlderThan)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]int{"deleted": deleted})
}

func (h *Handler) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	nss, err := h.store.ListNamespaces(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, nss)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.Stats(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, stats)
}

// ---------------------------------------------------------------------------
// Import / Export
// ---------------------------------------------------------------------------

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	records, err := h.store.ExportRecords(r.Context(), namespace)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, records)
}

func (h *Handler) handleImport(w http.ResponseWriter, r *http.Request) {
	var records []*model.Record
	if err := server.Decode(r, &records); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.store.ImportRecords(r.Context(), records); err != nil {
		h.handleStoreError(w, err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]int{"imported": len(records)})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (h *Handler) handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		server.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, store.ErrAlreadyExists):
		server.RespondError(w, http.StatusConflict, "ALREADY_EXISTS", err.Error())
	case errors.Is(err, store.ErrImmutable):
		server.RespondError(w, http.StatusConflict, "IMMUTABLE", err.Error())
	default:
		h.logger.Error("store error", "error", err)
		server.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}
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
