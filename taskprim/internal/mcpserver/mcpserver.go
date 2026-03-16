// Package mcpserver exposes taskprim operations as MCP (Model Context Protocol)
// tools. This allows any MCP-compatible client (Claude Desktop, Cursor, agent
// frameworks) to interact with taskprim natively.
//
// 11 tools are registered, covering the full task lifecycle:
//
//	taskprim_add          — create a new task
//	taskprim_list         — list/query tasks with filters
//	taskprim_get          — get a single task by ID
//	taskprim_done         — mark a task as done
//	taskprim_kill         — mark a task as killed (reason required)
//	taskprim_edit         — update task fields
//	taskprim_seen         — mark tasks as seen by an agent
//	taskprim_label_clear  — remove a label from tasks
//	taskprim_labels       — list labels with counts
//	taskprim_lists        — list all lists with counts
//	taskprim_stats        — aggregate statistics
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsvr "github.com/mark3labs/mcp-go/server"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/propifly/primkit/taskprim/internal/store"
)

// New creates an MCP server with all taskprim tools registered. The server
// name and version appear in the MCP handshake.
func New(s store.Store, version string) *mcpsvr.MCPServer {
	srv := mcpsvr.NewMCPServer(
		"taskprim",
		version,
		mcpsvr.WithToolCapabilities(false),
	)

	// Register all tools.
	registerAddTool(srv, s)
	registerListTool(srv, s)
	registerGetTool(srv, s)
	registerDoneTool(srv, s)
	registerKillTool(srv, s)
	registerEditTool(srv, s)
	registerSeenTool(srv, s)
	registerLabelClearTool(srv, s)
	registerLabelsTool(srv, s)
	registerListsTool(srv, s)
	registerStatsTool(srv, s)
	registerDepAddTool(srv, s)
	registerDepRemoveTool(srv, s)
	registerDepsTool(srv, s)
	registerDependentsTool(srv, s)
	registerFrontierTool(srv, s)

	return srv
}

// ---------------------------------------------------------------------------
// Tool registration
// ---------------------------------------------------------------------------

func registerAddTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_add",
		mcp.WithDescription("Create a new task. Returns the created task as JSON."),
		mcp.WithString("what",
			mcp.Required(),
			mcp.Description("Task description — what needs to be done"),
		),
		mcp.WithString("list",
			mcp.Required(),
			mcp.Description("Which list (actor/queue) this task belongs to"),
		),
		mcp.WithString("source",
			mcp.Description("Who created this task (defaults to 'mcp')"),
		),
		mcp.WithArray("labels",
			mcp.Description("Freeform labels for categorization"),
			mcp.WithStringItems(),
		),
		mcp.WithString("waiting_on",
			mcp.Description("What this task is blocked on"),
		),
		mcp.WithString("parent_id",
			mcp.Description("Parent task ID for subtasks"),
		),
		mcp.WithString("context",
			mcp.Description("Additional context or notes"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		what, err := req.RequireString("what")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		list, err := req.RequireString("list")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		source := req.GetString("source", "mcp")
		labels := req.GetStringSlice("labels", nil)
		waitingOn := optionalString(req, "waiting_on")
		parentID := optionalString(req, "parent_id")
		taskCtx := optionalString(req, "context")

		task := &model.Task{
			What:      what,
			List:      list,
			Source:    source,
			Labels:    labels,
			WaitingOn: waitingOn,
			ParentID:  parentID,
			Context:   taskCtx,
		}

		if err := s.CreateTask(ctx, task); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("creating task: %s", err)), nil
		}

		return jsonResult(task)
	})
}

func registerListTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_list",
		mcp.WithDescription("List tasks matching filters. Returns a JSON array of tasks."),
		mcp.WithString("list",
			mcp.Description("Filter by list name"),
		),
		mcp.WithString("state",
			mcp.Description("Filter by state"),
			mcp.Enum("open", "done", "killed"),
		),
		mcp.WithArray("labels",
			mcp.Description("Filter by labels (AND logic)"),
			mcp.WithStringItems(),
		),
		mcp.WithString("source",
			mcp.Description("Filter by source"),
		),
		mcp.WithBoolean("waiting",
			mcp.Description("Only tasks with waiting_on set"),
		),
		mcp.WithString("unseen_by",
			mcp.Description("Tasks not seen by this agent"),
		),
		mcp.WithString("seen_by",
			mcp.Description("Tasks seen by this agent (use with since)"),
		),
		mcp.WithString("since",
			mcp.Description("Time window for seen_by (e.g., 24h, 7d)"),
		),
		mcp.WithString("stale",
			mcp.Description("Tasks not updated within duration (e.g., 7d)"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filter := &model.Filter{
			List:     req.GetString("list", ""),
			Source:   req.GetString("source", ""),
			UnseenBy: req.GetString("unseen_by", ""),
			SeenBy:   req.GetString("seen_by", ""),
		}

		if state := req.GetString("state", ""); state != "" {
			s := model.State(state)
			filter.State = &s
		}
		if labels := req.GetStringSlice("labels", nil); len(labels) > 0 {
			filter.Labels = labels
		}
		if waiting := req.GetBool("waiting", false); waiting {
			w := true
			filter.Waiting = &w
		}
		if since := req.GetString("since", ""); since != "" {
			d, err := parseDuration(since)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid since: %s", err)), nil
			}
			filter.Since = d
		}
		if stale := req.GetString("stale", ""); stale != "" {
			d, err := parseDuration(stale)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid stale: %s", err)), nil
			}
			filter.Stale = d
		}

		tasks, err := s.ListTasks(ctx, filter)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing tasks: %s", err)), nil
		}

		return jsonResult(tasks)
	})
}

func registerGetTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_get",
		mcp.WithDescription("Get a single task by ID. Returns the task as JSON."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID (e.g., t_abc123)"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		task, err := s.GetTask(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("getting task: %s", err)), nil
		}

		return jsonResult(task)
	})
}

func registerDoneTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_done",
		mcp.WithDescription("Mark a task as done. Returns the updated task."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to mark as done"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := s.DoneTask(ctx, id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("completing task: %s", err)), nil
		}

		task, err := s.GetTask(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching task: %s", err)), nil
		}

		return jsonResult(task)
	})
}

func registerKillTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_kill",
		mcp.WithDescription("Mark a task as killed (dropped). Requires a reason explaining why."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to kill"),
		),
		mcp.WithString("reason",
			mcp.Required(),
			mcp.Description("Why this task is being dropped"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		reason, err := req.RequireString("reason")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := s.KillTask(ctx, id, reason); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("killing task: %s", err)), nil
		}

		task, err := s.GetTask(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching task: %s", err)), nil
		}

		return jsonResult(task)
	})
}

func registerEditTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_edit",
		mcp.WithDescription("Update task fields. Only specified fields are changed. Returns the updated task."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to edit"),
		),
		mcp.WithString("what",
			mcp.Description("New task description"),
		),
		mcp.WithString("list",
			mcp.Description("Move to a different list"),
		),
		mcp.WithString("waiting_on",
			mcp.Description("Set or clear (empty string) the waiting_on field"),
		),
		mcp.WithString("context",
			mcp.Description("Update context notes"),
		),
		mcp.WithString("parent_id",
			mcp.Description("Set or clear parent task ID"),
		),
		mcp.WithArray("add_labels",
			mcp.Description("Labels to add"),
			mcp.WithStringItems(),
		),
		mcp.WithArray("del_labels",
			mcp.Description("Labels to remove"),
			mcp.WithStringItems(),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		update := &model.TaskUpdate{
			What:      optionalString(req, "what"),
			List:      optionalString(req, "list"),
			WaitingOn: optionalString(req, "waiting_on"),
			Context:   optionalString(req, "context"),
			ParentID:  optionalString(req, "parent_id"),
			AddLabels: req.GetStringSlice("add_labels", nil),
			DelLabels: req.GetStringSlice("del_labels", nil),
		}

		if err := s.UpdateTask(ctx, id, update); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("updating task: %s", err)), nil
		}

		task, err := s.GetTask(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching task: %s", err)), nil
		}

		return jsonResult(task)
	})
}

func registerSeenTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_seen",
		mcp.WithDescription("Mark tasks as seen by an agent. Provide either task_ids or list."),
		mcp.WithString("agent",
			mcp.Required(),
			mcp.Description("Agent name (e.g., 'johanna')"),
		),
		mcp.WithArray("task_ids",
			mcp.Description("Specific task IDs to mark as seen"),
			mcp.WithStringItems(),
		),
		mcp.WithString("list",
			mcp.Description("Mark all open tasks in this list as seen"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agent, err := req.RequireString("agent")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		list := req.GetString("list", "")
		taskIDs := req.GetStringSlice("task_ids", nil)

		if list != "" {
			if err := s.MarkAllSeen(ctx, agent, list); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marking all seen: %s", err)), nil
			}
			return mcp.NewToolResultText(
				fmt.Sprintf("Marked all open tasks in %q as seen by %s", list, agent),
			), nil
		}

		if len(taskIDs) > 0 {
			if err := s.MarkSeen(ctx, agent, taskIDs); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marking seen: %s", err)), nil
			}
			return mcp.NewToolResultText(
				fmt.Sprintf("Marked %d task(s) as seen by %s", len(taskIDs), agent),
			), nil
		}

		return mcp.NewToolResultError("provide either task_ids or list"), nil
	})
}

func registerLabelClearTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_label_clear",
		mcp.WithDescription("Remove a label from all tasks. Optionally scope to a specific list."),
		mcp.WithString("label",
			mcp.Required(),
			mcp.Description("Label to remove"),
		),
		mcp.WithString("list",
			mcp.Description("Only clear from tasks in this list"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		list := req.GetString("list", "")

		count, err := s.ClearLabel(ctx, label, list)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("clearing label: %s", err)), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Cleared %q from %d task(s)", label, count),
		), nil
	})
}

func registerLabelsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_labels",
		mcp.WithDescription("List all labels with open task counts."),
		mcp.WithString("list",
			mcp.Description("Only labels from tasks in this list"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		list := req.GetString("list", "")

		labels, err := s.ListLabels(ctx, list)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing labels: %s", err)), nil
		}

		return jsonResult(labels)
	})
}

func registerListsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_lists",
		mcp.WithDescription("List all lists with task counts by state."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lists, err := s.ListLists(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing lists: %s", err)), nil
		}

		return jsonResult(lists)
	})
}

func registerStatsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_stats",
		mcp.WithDescription("Get aggregate task statistics (open, done, killed counts)."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := s.Stats(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("fetching stats: %s", err)), nil
		}

		return jsonResult(stats)
	})
}

func registerDepAddTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_dep_add",
		mcp.WithDescription("Add a dependency edge between two tasks. task_id depends on depends_on."),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task that has the dependency"),
		),
		mcp.WithString("depends_on",
			mcp.Required(),
			mcp.Description("The task it depends on"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dependsOn, err := req.RequireString("depends_on")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := s.AddDep(ctx, taskID, dependsOn); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("adding dep: %s", err)), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Added: %s depends on %s", taskID, dependsOn),
		), nil
	})
}

func registerDepRemoveTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_dep_remove",
		mcp.WithDescription("Remove a dependency edge between two tasks."),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task that has the dependency"),
		),
		mcp.WithString("depends_on",
			mcp.Required(),
			mcp.Description("The dependency to remove"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dependsOn, err := req.RequireString("depends_on")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := s.RemoveDep(ctx, taskID, dependsOn); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("removing dep: %s", err)), nil
		}

		return mcp.NewToolResultText(
			fmt.Sprintf("Removed: %s no longer depends on %s", taskID, dependsOn),
		), nil
	})
}

func registerDepsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_deps",
		mcp.WithDescription("List what a task depends on. Returns a JSON array of tasks."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to list dependencies for"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tasks, err := s.Deps(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing deps: %s", err)), nil
		}

		return jsonResult(tasks)
	})
}

func registerDependentsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_dependents",
		mcp.WithDescription("List what depends on a task (reverse lookup). Returns a JSON array of tasks."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID to find dependents for"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tasks, err := s.Dependents(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing dependents: %s", err)), nil
		}

		return jsonResult(tasks)
	})
}

func registerFrontierTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("taskprim_frontier",
		mcp.WithDescription("List tasks ready for execution — open tasks with all deps resolved or no deps."),
		mcp.WithString("list",
			mcp.Description("Filter by list name"),
		),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		list := req.GetString("list", "")

		tasks, err := s.Frontier(ctx, list)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("querying frontier: %s", err)), nil
		}

		return jsonResult(tasks)
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// optionalString returns a *string if the key is present and non-empty in the
// request arguments, nil otherwise. This maps MCP's "all params are optional"
// model to Go's pointer-based optionality.
func optionalString(req mcp.CallToolRequest, key string) *string {
	v := req.GetString(key, "")
	if v == "" {
		return nil
	}
	return &v
}

// jsonResult serializes a value to indented JSON and returns it as a text result.
// MCP tool results are text-based, so we format as JSON for structured data.
func jsonResult(v interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serializing result: %s", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
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
