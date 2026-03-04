// Package api implements the taskprim HTTP API. It translates HTTP requests
// into store operations and formats responses as JSON. Every handler follows
// the same pattern: parse input → call store → respond.
//
// The API is versioned under /v1/ and uses standard HTTP methods:
//   - POST   /v1/tasks           → create task
//   - GET    /v1/tasks           → list tasks (with query param filters)
//   - GET    /v1/tasks/:id       → get single task
//   - PATCH  /v1/tasks/:id       → update task fields
//   - POST   /v1/tasks/:id/done  → mark done
//   - POST   /v1/tasks/:id/kill  → mark killed (requires reason)
//   - POST   /v1/seen/:agent     → mark tasks as seen
//   - POST   /v1/labels/:name/clear → remove label from tasks
//   - GET    /v1/labels          → list labels with counts
//   - GET    /v1/lists           → list all lists with counts
//   - GET    /v1/stats           → aggregate statistics
package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/propifly/primkit/primkit/auth"
	"github.com/propifly/primkit/primkit/server"
	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/propifly/primkit/taskprim/internal/store"
)

// Handler holds the HTTP handlers for the taskprim API. It wraps a store
// and logger that are shared across all endpoints.
type Handler struct {
	store  store.Store
	logger *slog.Logger
}

// New creates an API handler backed by the given store.
func New(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// Router returns an http.Handler with all taskprim routes registered.
// The caller is responsible for applying middleware (auth, logging, etc.)
// before passing this to the server.
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Task CRUD.
	mux.HandleFunc("POST /v1/tasks", h.createTask)
	mux.HandleFunc("GET /v1/tasks", h.listTasks)
	mux.HandleFunc("GET /v1/tasks/{id}", h.getTask)
	mux.HandleFunc("PATCH /v1/tasks/{id}", h.updateTask)
	mux.HandleFunc("POST /v1/tasks/{id}/done", h.doneTask)
	mux.HandleFunc("POST /v1/tasks/{id}/kill", h.killTask)

	// Seen tracking.
	mux.HandleFunc("POST /v1/seen/{agent}", h.markSeen)

	// Labels and lists.
	mux.HandleFunc("POST /v1/labels/{name}/clear", h.clearLabel)
	mux.HandleFunc("GET /v1/labels", h.listLabels)
	mux.HandleFunc("GET /v1/lists", h.listLists)
	mux.HandleFunc("GET /v1/stats", h.stats)

	return mux
}

// ---------------------------------------------------------------------------
// Task endpoints
// ---------------------------------------------------------------------------

// createTaskRequest is the JSON body for POST /v1/tasks.
type createTaskRequest struct {
	List      string   `json:"list"`
	What      string   `json:"what"`
	Labels    []string `json:"labels"`
	WaitingOn *string  `json:"waiting_on"`
	ParentID  *string  `json:"parent_id"`
	Context   *string  `json:"context"`
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	// Use the authenticated key name as the source, falling back to "api".
	source := auth.KeyNameFromContext(r.Context())
	if source == "" {
		source = "api"
	}

	task := &model.Task{
		List:      req.List,
		What:      req.What,
		Source:    source,
		Labels:   req.Labels,
		WaitingOn: req.WaitingOn,
		ParentID:  req.ParentID,
		Context:   req.Context,
	}

	if err := h.store.CreateTask(r.Context(), task); err != nil {
		h.handleStoreError(w, "creating task", err)
		return
	}

	server.Respond(w, http.StatusCreated, task)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	filter, err := filterFromQuery(r)
	if err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_FILTER", err.Error())
		return
	}

	tasks, err := h.store.ListTasks(r.Context(), filter)
	if err != nil {
		h.handleStoreError(w, "listing tasks", err)
		return
	}

	server.Respond(w, http.StatusOK, tasks)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "getting task", err)
		return
	}

	server.Respond(w, http.StatusOK, task)
}

// updateTaskRequest is the JSON body for PATCH /v1/tasks/:id. All fields
// are optional — only fields present in the JSON are updated.
type updateTaskRequest struct {
	What      *string  `json:"what"`
	List      *string  `json:"list"`
	WaitingOn *string  `json:"waiting_on"`
	Context   *string  `json:"context"`
	ParentID  *string  `json:"parent_id"`
	AddLabels []string `json:"add_labels"`
	DelLabels []string `json:"del_labels"`
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateTaskRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	update := &model.TaskUpdate{
		What:      req.What,
		List:      req.List,
		WaitingOn: req.WaitingOn,
		Context:   req.Context,
		ParentID:  req.ParentID,
		AddLabels: req.AddLabels,
		DelLabels: req.DelLabels,
	}

	if err := h.store.UpdateTask(r.Context(), id, update); err != nil {
		h.handleStoreError(w, "updating task", err)
		return
	}

	// Return the updated task so the client sees the new state.
	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "fetching updated task", err)
		return
	}

	server.Respond(w, http.StatusOK, task)
}

func (h *Handler) doneTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.store.DoneTask(r.Context(), id); err != nil {
		h.handleStoreError(w, "completing task", err)
		return
	}

	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "fetching completed task", err)
		return
	}

	server.Respond(w, http.StatusOK, task)
}

// killTaskRequest is the JSON body for POST /v1/tasks/:id/kill.
type killTaskRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) killTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req killTaskRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.Reason == "" {
		server.RespondError(w, http.StatusBadRequest, "REASON_REQUIRED",
			"resolved_reason is required when killing a task")
		return
	}

	if err := h.store.KillTask(r.Context(), id, req.Reason); err != nil {
		h.handleStoreError(w, "killing task", err)
		return
	}

	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		h.handleStoreError(w, "fetching killed task", err)
		return
	}

	server.Respond(w, http.StatusOK, task)
}

// ---------------------------------------------------------------------------
// Seen tracking
// ---------------------------------------------------------------------------

// markSeenRequest is the JSON body for POST /v1/seen/:agent. The client
// sends either task_ids (mark specific tasks) or list (mark all open in list).
type markSeenRequest struct {
	TaskIDs []string `json:"task_ids"`
	List    string   `json:"list"`
}

func (h *Handler) markSeen(w http.ResponseWriter, r *http.Request) {
	agent := r.PathValue("agent")

	var req markSeenRequest
	if err := server.Decode(r, &req); err != nil {
		server.RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if req.List != "" {
		if err := h.store.MarkAllSeen(r.Context(), agent, req.List); err != nil {
			h.handleStoreError(w, "marking all seen", err)
			return
		}
		server.Respond(w, http.StatusOK, map[string]string{
			"status": "ok",
			"agent":  agent,
			"list":   req.List,
		})
	} else if len(req.TaskIDs) > 0 {
		if err := h.store.MarkSeen(r.Context(), agent, req.TaskIDs); err != nil {
			h.handleStoreError(w, "marking seen", err)
			return
		}
		server.Respond(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"agent":  agent,
			"count":  len(req.TaskIDs),
		})
	} else {
		server.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			"provide either task_ids or list")
		return
	}
}

// ---------------------------------------------------------------------------
// Labels, lists, stats
// ---------------------------------------------------------------------------

func (h *Handler) clearLabel(w http.ResponseWriter, r *http.Request) {
	label := r.PathValue("name")
	list := r.URL.Query().Get("list")

	count, err := h.store.ClearLabel(r.Context(), label, list)
	if err != nil {
		h.handleStoreError(w, "clearing label", err)
		return
	}

	server.Respond(w, http.StatusOK, map[string]interface{}{
		"label":   label,
		"cleared": count,
	})
}

func (h *Handler) listLabels(w http.ResponseWriter, r *http.Request) {
	list := r.URL.Query().Get("list")

	labels, err := h.store.ListLabels(r.Context(), list)
	if err != nil {
		h.handleStoreError(w, "listing labels", err)
		return
	}

	server.Respond(w, http.StatusOK, labels)
}

func (h *Handler) listLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.store.ListLists(r.Context())
	if err != nil {
		h.handleStoreError(w, "listing lists", err)
		return
	}

	server.Respond(w, http.StatusOK, lists)
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

// filterFromQuery builds a Filter from URL query parameters. This maps the
// query param names from the design spec to the model.Filter fields.
func filterFromQuery(r *http.Request) (*model.Filter, error) {
	q := r.URL.Query()
	filter := &model.Filter{}

	filter.List = q.Get("list")
	filter.Source = q.Get("source")
	filter.UnseenBy = q.Get("unseen_by")
	filter.SeenBy = q.Get("seen_by")

	if state := q.Get("state"); state != "" {
		s := model.State(state)
		filter.State = &s
	}

	// Labels: single "label" param or comma-separated "labels" param.
	if label := q.Get("label"); label != "" {
		filter.Labels = []string{label}
	}
	if labels := q.Get("labels"); labels != "" {
		filter.Labels = strings.Split(labels, ",")
	}

	if q.Get("waiting") == "true" {
		w := true
		filter.Waiting = &w
	}

	if since := q.Get("since"); since != "" {
		d, err := parseDuration(since)
		if err != nil {
			return nil, err
		}
		filter.Since = d
	}

	if stale := q.Get("stale"); stale != "" {
		d, err := parseDuration(stale)
		if err != nil {
			return nil, err
		}
		filter.Stale = d
	}

	// parent=none means only top-level tasks (no parent).
	if parent := q.Get("parent"); parent != "" {
		if parent == "none" {
			empty := ""
			filter.ParentID = &empty
		} else {
			filter.ParentID = &parent
		}
	}

	return filter, nil
}

// parseDuration handles both Go-style durations (24h, 30m) and short-form
// day notation (7d, 30d) which Go's time.ParseDuration doesn't support.
func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}

// handleStoreError translates store-level errors into appropriate HTTP
// responses. Known sentinel errors get specific status codes; everything
// else is a 500.
func (h *Handler) handleStoreError(w http.ResponseWriter, action string, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		server.RespondError(w, http.StatusNotFound, "NOT_FOUND", "task not found")
	case errors.Is(err, store.ErrInvalidTransition):
		server.RespondError(w, http.StatusConflict, "INVALID_TRANSITION", err.Error())
	default:
		h.logger.Error(action, "error", err)
		server.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"an internal error occurred")
	}
}
