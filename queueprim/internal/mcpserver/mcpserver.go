// Package mcpserver exposes queueprim operations as MCP (Model Context Protocol)
// tools. This allows any MCP-compatible client (Claude Desktop, Cursor, agent
// frameworks) to interact with queueprim natively.
//
// 11 tools are registered, covering the full job lifecycle:
//
//	queueprim_enqueue   — enqueue a new job
//	queueprim_dequeue   — atomically claim the next job
//	queueprim_complete  — mark a claimed job as done
//	queueprim_fail      — mark a claimed job as failed
//	queueprim_release   — return a claimed job to pending
//	queueprim_extend    — extend a claimed job's visibility timeout
//	queueprim_peek      — inspect next job without claiming
//	queueprim_list      — list jobs with filters
//	queueprim_get       — get a specific job
//	queueprim_queues    — list all queues with counts
//	queueprim_stats     — aggregate statistics
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsvr "github.com/mark3labs/mcp-go/server"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/propifly/primkit/queueprim/internal/store"
)

// New creates an MCP server with all queueprim tools registered.
func New(s store.Store, version string) *mcpsvr.MCPServer {
	srv := mcpsvr.NewMCPServer(
		"queueprim",
		version,
		mcpsvr.WithToolCapabilities(false),
	)

	registerEnqueueTool(srv, s)
	registerDequeueTool(srv, s)
	registerCompleteTool(srv, s)
	registerFailTool(srv, s)
	registerReleaseTool(srv, s)
	registerExtendTool(srv, s)
	registerPeekTool(srv, s)
	registerListTool(srv, s)
	registerGetTool(srv, s)
	registerQueuesTool(srv, s)
	registerStatsTool(srv, s)

	return srv
}

// ---------------------------------------------------------------------------
// queueprim_enqueue
// ---------------------------------------------------------------------------

func registerEnqueueTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_enqueue",
		mcp.WithDescription("Enqueue a new job into a named queue."),
		mcp.WithString("queue", mcp.Required(), mcp.Description("Queue name, e.g. infra/fixes")),
		mcp.WithString("payload", mcp.Required(), mcp.Description("Arbitrary JSON payload string")),
		mcp.WithString("type", mcp.Description("Job type category for workers, e.g. ssh_auth_fail")),
		mcp.WithString("priority", mcp.Description("high, normal (default), or low")),
		mcp.WithNumber("max_retries", mcp.Description("Max retries before dead-letter (default 0 = one-shot)")),
		mcp.WithString("delay", mcp.Description("Delay before job is visible, e.g. 5m, 1h")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queue := mcp.ParseString(req, "queue", "")
		payloadStr := mcp.ParseString(req, "payload", "")
		jobType := mcp.ParseString(req, "type", "")
		priority := mcp.ParseString(req, "priority", "")
		maxRetries := int(req.GetFloat("max_retries", 0))
		delayStr := mcp.ParseString(req, "delay", "")

		if queue == "" {
			return mcp.NewToolResultError("queue is required"), nil
		}
		if payloadStr == "" {
			return mcp.NewToolResultError("payload is required"), nil
		}

		job := &model.Job{
			Queue:      queue,
			Type:       jobType,
			Priority:   model.Priority(priority),
			Payload:    json.RawMessage(payloadStr),
			MaxRetries: maxRetries,
		}

		if delayStr != "" {
			d, err := time.ParseDuration(delayStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid delay %q: %v", delayStr, err)), nil
			}
			job.VisibleAfter = time.Now().UTC().Add(d)
		}

		if err := s.EnqueueJob(ctx, job); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_dequeue
// ---------------------------------------------------------------------------

func registerDequeueTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_dequeue",
		mcp.WithDescription("Atomically claim the next available job in a queue. Returns the claimed job or an empty result if the queue has no available jobs."),
		mcp.WithString("queue", mcp.Required(), mcp.Description("Queue name to dequeue from")),
		mcp.WithString("worker", mcp.Description("Worker agent name (for claimed_by tracking)")),
		mcp.WithString("timeout", mcp.Description("Visibility timeout, e.g. 30m (default 30m)")),
		mcp.WithString("type", mcp.Description("Only claim jobs of this type")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queue := mcp.ParseString(req, "queue", "")
		worker := mcp.ParseString(req, "worker", "mcp-agent")
		timeoutStr := mcp.ParseString(req, "timeout", "30m")
		jobType := mcp.ParseString(req, "type", "")

		if queue == "" {
			return mcp.NewToolResultError("queue is required"), nil
		}

		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid timeout %q: %v", timeoutStr, err)), nil
		}

		job, err := s.DequeueJob(ctx, queue, worker, jobType, timeout)
		if err != nil {
			if err == store.ErrEmpty {
				return mcp.NewToolResultText("queue is empty — no jobs available"), nil
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_complete
// ---------------------------------------------------------------------------

func registerCompleteTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_complete",
		mcp.WithDescription("Mark a claimed job as done."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Job ID (q_...)")),
		mcp.WithString("output", mcp.Description("Optional JSON output payload from the worker")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(req, "id", "")
		outputStr := mcp.ParseString(req, "output", "")

		if id == "" {
			return mcp.NewToolResultError("id is required"), nil
		}

		var output []byte
		if outputStr != "" {
			output = []byte(outputStr)
		}

		if err := s.CompleteJob(ctx, id, output); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job, err := s.GetJob(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_fail
// ---------------------------------------------------------------------------

func registerFailTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_fail",
		mcp.WithDescription("Mark a claimed job as failed. The job retries (if retries remain) or moves to dead."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
		mcp.WithString("reason", mcp.Description("Human-readable failure reason")),
		mcp.WithBoolean("dead", mcp.Description("Force to dead-letter regardless of retry count")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(req, "id", "")
		reason := mcp.ParseString(req, "reason", "")
		dead := req.GetBool("dead", false)

		if id == "" {
			return mcp.NewToolResultError("id is required"), nil
		}

		if err := s.FailJob(ctx, id, reason, dead); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job, err := s.GetJob(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_release
// ---------------------------------------------------------------------------

func registerReleaseTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_release",
		mcp.WithDescription("Return a claimed job to pending immediately (unclaim)."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(req, "id", "")
		if id == "" {
			return mcp.NewToolResultError("id is required"), nil
		}

		if err := s.ReleaseJob(ctx, id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job, err := s.GetJob(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_extend
// ---------------------------------------------------------------------------

func registerExtendTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_extend",
		mcp.WithDescription("Extend a claimed job's visibility timeout to prevent auto-release."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Job ID")),
		mcp.WithString("by", mcp.Description("Extension duration, e.g. 30m (default 30m)")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(req, "id", "")
		byStr := mcp.ParseString(req, "by", "30m")

		if id == "" {
			return mcp.NewToolResultError("id is required"), nil
		}

		by, err := time.ParseDuration(byStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid duration %q: %v", byStr, err)), nil
		}

		if err := s.ExtendJob(ctx, id, by); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job, err := s.GetJob(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_peek
// ---------------------------------------------------------------------------

func registerPeekTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_peek",
		mcp.WithDescription("Inspect the next job that would be dequeued, without claiming it."),
		mcp.WithString("queue", mcp.Required(), mcp.Description("Queue name")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queue := mcp.ParseString(req, "queue", "")
		if queue == "" {
			return mcp.NewToolResultError("queue is required"), nil
		}

		job, err := s.PeekJob(ctx, queue)
		if err != nil {
			if err == store.ErrEmpty {
				return mcp.NewToolResultText("queue is empty"), nil
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_list
// ---------------------------------------------------------------------------

func registerListTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_list",
		mcp.WithDescription("List jobs with optional filters."),
		mcp.WithString("queue", mcp.Description("Filter to this queue")),
		mcp.WithString("status", mcp.Description("pending, claimed, done, failed, or dead")),
		mcp.WithString("type", mcp.Description("Filter by job type")),
		mcp.WithString("older_than", mcp.Description("Only jobs created before now-duration, e.g. 1h")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filter := &model.Filter{
			Queue: mcp.ParseString(req, "queue", ""),
			Type:  mcp.ParseString(req, "type", ""),
		}

		if s := mcp.ParseString(req, "status", ""); s != "" {
			st := model.Status(s)
			filter.Status = &st
		}

		if ot := mcp.ParseString(req, "older_than", ""); ot != "" {
			d, err := time.ParseDuration(ot)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid older_than %q: %v", ot, err)), nil
			}
			filter.OlderThan = d
		}

		jobs, err := s.ListJobs(ctx, filter)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(jobs) == 0 {
			return mcp.NewToolResultText("no jobs found"), nil
		}

		data, _ := json.MarshalIndent(jobs, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_get
// ---------------------------------------------------------------------------

func registerGetTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_get",
		mcp.WithDescription("Get a specific job by ID."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Job ID (q_...)")),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := mcp.ParseString(req, "id", "")
		if id == "" {
			return mcp.NewToolResultError("id is required"), nil
		}

		job, err := s.GetJob(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(job, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_queues
// ---------------------------------------------------------------------------

func registerQueuesTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_queues",
		mcp.WithDescription("List all queues with job counts by status."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queues, err := s.ListQueues(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(queues) == 0 {
			return mcp.NewToolResultText("no queues found"), nil
		}

		data, _ := json.MarshalIndent(queues, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

// ---------------------------------------------------------------------------
// queueprim_stats
// ---------------------------------------------------------------------------

func registerStatsTool(srv *mcpsvr.MCPServer, s store.Store) {
	tool := mcp.NewTool("queueprim_stats",
		mcp.WithDescription("Get aggregate job statistics across all queues."),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := s.Stats(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, _ := json.MarshalIndent(stats, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}
