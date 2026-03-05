// Package api implements the queueprim HTTP API. It translates HTTP requests
// into store operations and formats responses as JSON. Every handler follows
// the same pattern: parse input → call store → respond.
//
// Routes:
//
//	POST   /v1/jobs                        → enqueue
//	GET    /v1/jobs                        → list jobs (query params)
//	GET    /v1/jobs/:id                    → get single job
//
//	POST   /v1/queues/:queue/dequeue       → atomic claim
//	POST   /v1/jobs/:id/complete           → mark done
//	POST   /v1/jobs/:id/fail               → mark failed
//	POST   /v1/jobs/:id/release            → unclaim
//	POST   /v1/jobs/:id/extend             → extend visibility timeout
//
//	GET    /v1/queues                      → list queues with counts
//	GET    /v1/stats                       → aggregate statistics
//
//	DELETE /v1/queues/:queue               → purge (with ?status=done&older_than=7d)
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/propifly/primkit/queueprim/internal/store"
)

// Handler holds the HTTP handlers for the queueprim API.
type Handler struct {
	store  store.Store
	logger *slog.Logger
}

// New creates an API handler backed by the given store.
func New(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// Router returns an http.Handler with all queueprim routes registered.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/jobs", h.enqueueJob)
	mux.HandleFunc("GET /v1/jobs", h.listJobs)
	mux.HandleFunc("GET /v1/jobs/{id}", h.getJob)

	// Queue names may contain slashes (e.g., "infra/fixes"), so we use a prefix
	// match rather than a {queue} wildcard (which stops at the first slash).
	mux.HandleFunc("POST /v1/queues/", h.dequeueJob)
	mux.HandleFunc("POST /v1/jobs/{id}/complete", h.completeJob)
	mux.HandleFunc("POST /v1/jobs/{id}/fail", h.failJob)
	mux.HandleFunc("POST /v1/jobs/{id}/release", h.releaseJob)
	mux.HandleFunc("POST /v1/jobs/{id}/extend", h.extendJob)

	mux.HandleFunc("GET /v1/queues", h.listQueues)
	mux.HandleFunc("GET /v1/stats", h.getStats)

	mux.HandleFunc("DELETE /v1/queues/{queue}", h.purgeQueue)

	return mux
}

// ---------------------------------------------------------------------------
// POST /v1/jobs — enqueue
// ---------------------------------------------------------------------------

type enqueueRequest struct {
	Queue      string          `json:"queue"`
	Type       string          `json:"type"`
	Priority   string          `json:"priority"`
	Payload    json.RawMessage `json:"payload"`
	MaxRetries int             `json:"max_retries"`
	Delay      string          `json:"delay"` // e.g. "5m", "1h"
}

func (h *Handler) enqueueJob(w http.ResponseWriter, r *http.Request) {
	var req enqueueRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	job := &model.Job{
		Queue:      req.Queue,
		Type:       req.Type,
		Priority:   model.Priority(req.Priority),
		Payload:    req.Payload,
		MaxRetries: req.MaxRetries,
	}

	if req.Delay != "" {
		d, err := time.ParseDuration(req.Delay)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_DELAY",
				fmt.Sprintf("invalid delay %q: %v", req.Delay, err))
			return
		}
		job.VisibleAfter = time.Now().UTC().Add(d)
	}

	if err := h.store.EnqueueJob(r.Context(), job); err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusCreated, job)
}

// ---------------------------------------------------------------------------
// GET /v1/jobs — list jobs
// ---------------------------------------------------------------------------

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := &model.Filter{
		Queue: q.Get("queue"),
		Type:  q.Get("type"),
	}

	if s := q.Get("status"); s != "" {
		st := model.Status(s)
		filter.Status = &st
	}

	if ot := q.Get("older_than"); ot != "" {
		d, err := time.ParseDuration(ot)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_PARAM",
				fmt.Sprintf("invalid older_than %q: %v", ot, err))
			return
		}
		filter.OlderThan = d
	}

	jobs, err := h.store.ListJobs(r.Context(), filter)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	if jobs == nil {
		jobs = []*model.Job{}
	}
	server.Respond(w, http.StatusOK, jobs)
}

// ---------------------------------------------------------------------------
// GET /v1/jobs/{id} — get single job
// ---------------------------------------------------------------------------

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := h.store.GetJob(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, job)
}

// ---------------------------------------------------------------------------
// POST /v1/queues/{queue}/dequeue — atomic claim
// ---------------------------------------------------------------------------

type dequeueRequest struct {
	Worker  string `json:"worker"`
	Timeout string `json:"timeout"` // e.g. "30m"
	Type    string `json:"type"`
}

func (h *Handler) dequeueJob(w http.ResponseWriter, r *http.Request) {
	// Queue names may contain slashes, so we extract from the full URL path.
	queue := extractQueueFromPath(r.URL.Path, "/v1/queues/", "/dequeue")
	if queue == "" {
		server.RespondError(w, http.StatusBadRequest, "INVALID_PATH",
			"path must be /v1/queues/<queue>/dequeue")
		return
	}

	var req dequeueRequest
	// Body is optional — defaults are fine.
	if r.ContentLength > 0 {
		if err := server.Decode(r, &req); err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
	}

	// Worker defaults to the authenticated key name, then "api".
	if req.Worker == "" {
		req.Worker = auth.KeyNameFromContext(r.Context())
		if req.Worker == "" {
			req.Worker = "api"
		}
	}

	var timeout time.Duration
	if req.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(req.Timeout)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_TIMEOUT",
				fmt.Sprintf("invalid timeout %q: %v", req.Timeout, err))
			return
		}
	}

	job, err := h.store.DequeueJob(r.Context(), queue, req.Worker, req.Type, timeout)
	if err != nil {
		if errors.Is(err, store.ErrEmpty) {
			// 204 No Content: queue is available but empty.
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, job)
}

// extractQueueFromPath extracts the queue name from paths like
// /v1/queues/infra/fixes/dequeue, where the queue name contains slashes.
func extractQueueFromPath(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	path = strings.TrimPrefix(path, prefix)
	if !strings.HasSuffix(path, suffix) {
		return ""
	}
	return strings.TrimSuffix(path, suffix)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/complete
// ---------------------------------------------------------------------------

type completeRequest struct {
	Output json.RawMessage `json:"output"`
}

func (h *Handler) completeJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req completeRequest
	if r.ContentLength > 0 {
		if err := server.Decode(r, &req); err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
	}

	if err := h.store.CompleteJob(r.Context(), id, req.Output); err != nil {
		h.handleStoreError(w, err)
		return
	}

	job, _ := h.store.GetJob(r.Context(), id)
	server.Respond(w, http.StatusOK, job)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/fail
// ---------------------------------------------------------------------------

type failRequest struct {
	Reason string `json:"reason"`
	Dead   bool   `json:"dead"` // force dead-letter regardless of retries
}

func (h *Handler) failJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req failRequest
	if r.ContentLength > 0 {
		if err := server.Decode(r, &req); err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
	}

	if err := h.store.FailJob(r.Context(), id, req.Reason, req.Dead); err != nil {
		h.handleStoreError(w, err)
		return
	}

	job, _ := h.store.GetJob(r.Context(), id)
	server.Respond(w, http.StatusOK, job)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/release
// ---------------------------------------------------------------------------

func (h *Handler) releaseJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.store.ReleaseJob(r.Context(), id); err != nil {
		h.handleStoreError(w, err)
		return
	}

	job, _ := h.store.GetJob(r.Context(), id)
	server.Respond(w, http.StatusOK, job)
}

// ---------------------------------------------------------------------------
// POST /v1/jobs/{id}/extend
// ---------------------------------------------------------------------------

type extendRequest struct {
	By string `json:"by"` // duration string, e.g. "30m"
}

func (h *Handler) extendJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req extendRequest
	if r.ContentLength > 0 {
		if err := server.Decode(r, &req); err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
	}

	by := 30 * time.Minute
	if req.By != "" {
		var err error
		by, err = time.ParseDuration(req.By)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_DURATION",
				fmt.Sprintf("invalid duration %q: %v", req.By, err))
			return
		}
	}

	if err := h.store.ExtendJob(r.Context(), id, by); err != nil {
		h.handleStoreError(w, err)
		return
	}

	job, _ := h.store.GetJob(r.Context(), id)
	server.Respond(w, http.StatusOK, job)
}

// ---------------------------------------------------------------------------
// GET /v1/queues
// ---------------------------------------------------------------------------

func (h *Handler) listQueues(w http.ResponseWriter, r *http.Request) {
	queues, err := h.store.ListQueues(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	if queues == nil {
		queues = []model.QueueInfo{}
	}
	server.Respond(w, http.StatusOK, queues)
}

// ---------------------------------------------------------------------------
// GET /v1/stats
// ---------------------------------------------------------------------------

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.Stats(r.Context())
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, stats)
}

// ---------------------------------------------------------------------------
// DELETE /v1/queues/{queue}
// ---------------------------------------------------------------------------

func (h *Handler) purgeQueue(w http.ResponseWriter, r *http.Request) {
	prefix := "/v1/queues/"
	queue := strings.TrimPrefix(r.URL.Path, prefix)
	if queue == "" {
		server.RespondError(w, http.StatusBadRequest, "MISSING_QUEUE", "queue name is required")
		return
	}

	q := r.URL.Query()
	statusStr := q.Get("status")
	if statusStr == "" {
		server.RespondError(w, http.StatusBadRequest, "MISSING_STATUS", "status query parameter is required")
		return
	}

	var olderThan time.Duration
	if ot := q.Get("older_than"); ot != "" {
		var err error
		olderThan, err = time.ParseDuration(ot)
		if err != nil {
			server.RespondError(w, http.StatusBadRequest, "INVALID_PARAM",
				fmt.Sprintf("invalid older_than %q: %v", ot, err))
			return
		}
	}

	n, err := h.store.PurgeJobs(r.Context(), queue, model.Status(statusStr), olderThan)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}
	server.Respond(w, http.StatusOK, map[string]int{"deleted": n})
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func (h *Handler) handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrValidation):
		server.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, store.ErrNotFound):
		server.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, store.ErrInvalidTransition):
		server.RespondError(w, http.StatusConflict, "INVALID_TRANSITION", err.Error())
	case errors.Is(err, store.ErrEmpty):
		w.WriteHeader(http.StatusNoContent)
	default:
		h.logger.Error("store error", "error", err)
		server.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
