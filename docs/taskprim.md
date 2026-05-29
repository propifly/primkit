# taskprim Guide

taskprim is a task-tracking primitive for agents and the humans they work with. Tasks have an explicit lifecycle, belong to named lists, carry freeform labels, and record who created them. On top of that flat model, taskprim adds two coordination layers that make a list usable by several agents at once: per-agent *seen-tracking* (what's new since I last looked) and a structural *dependency graph* with cycle detection and a frontier query ("what can I work on next?").

This guide covers the concepts you need to use taskprim well: the lifecycle and why it's one-way, the difference between `waiting_on` and dependencies, how the dependency graph and frontier work, and how seen-tracking lets multiple agents share one list without stepping on each other. For installation and quick start, see the [README](../README.md). For every command and flag, see the [agent reference](agent-reference.md).

taskprim is deliberately a *dumb primitive*: it stores tasks and edges and answers queries about them. It does not schedule, prioritize, or plan — that logic belongs to the agent. Keeping this boundary in mind explains most of taskprim's design.

---

## The Task Lifecycle

Every task moves through a small, one-way state machine:

```
          done
  open ──────────► done
    │
    │  kill(reason)
    └──────────► killed
```

A task starts `open` and ends either `done` (completed) or `killed` (abandoned). Both transitions are terminal — there is no reopen. If work needs to resume, create a new task. This one-way rule keeps history honest: a `done` task is a fact that happened, not a mutable status.

Three things are required at creation: the description, the `--list` it belongs to, and `--source` (who created it):

```bash
taskprim add "Deploy v2 to staging" --list ops --source johanna
```

Resolve a task when its work is finished or dropped:

```bash
taskprim done t_x7k2m9p4n1
taskprim kill t_x7k2m9p4n1 --reason "superseded by the v3 plan"
```

`kill` requires a `--reason` — an abandoned task without a recorded reason is a future mystery. Both `done` and `kill` are idempotent: calling them on an already-resolved task is a no-op, so retries are safe.

---

## Lists and Labels

Two independent ways to organize tasks:

- **Lists** are the primary partition. A task belongs to exactly one list (`--list ops`). Lists are arbitrary strings — use them for teams, projects, or agents. `taskprim lists` enumerates them, and almost every command takes `--list` to scope its work.
- **Labels** are freeform cross-cutting tags. A task can carry many (`--label deploy --label urgent`, or `--label deploy,urgent`). Filtering by multiple labels uses AND logic — `list --label deploy --label urgent` returns only tasks with both.

```bash
taskprim add "Rotate TLS certs" --list ops --label security --label urgent
taskprim list --list ops --label security --state open
```

Pick lists for *where work lives* and labels for *what work is about*. The two compose.

---

## Two Kinds of Blocking: waiting_on vs Dependencies

This is the distinction that trips people up, and getting it right is what makes taskprim's graph useful.

A task can be blocked in two fundamentally different ways:

- **`waiting_on`** is freeform text describing an *external* blocker — something taskprim doesn't track: "client approval", "CI pipeline green", "Q3 budget sign-off". It's a human-readable note, not a link.
- **A dependency** is a *structural* edge to another task taskprim does track. "Task B depends on task A" means B should not start until A is resolved.

They coexist because they answer different questions. `waiting_on` answers "why is this stuck on something outside the system?" Dependencies answer "what is the execution order within this body of work?" A single task can have both: it depends on task A *and* is waiting on a client signature.

```bash
# External blocker — freeform note
taskprim add "Publish the press release" --list pr --waiting-on "legal sign-off"

# Structural edge — B depends on A
taskprim dep add t_taskB t_taskA
```

If you find yourself encoding a task ID inside `waiting_on` text, that's a dependency — use `dep add` instead, so the frontier query understands it.

---

## The Dependency Graph and the Frontier

Dependencies form a directed graph. taskprim stores the edges, enforces a few invariants, and answers graph queries — nothing more.

**Adding and removing edges.** `dep add <task> <depends-on>` records that the first task depends on the second:

```bash
# "Deploy" depends on "Run migrations"
taskprim dep add t_deploy t_migrate
```

taskprim rejects edges that would create a cycle (detected by walking the existing graph) and edges from a task to itself. You *can* depend on an already-`done` task — that dependency is simply satisfied immediately.

**Inspecting the graph.** Two complementary views:

```bash
taskprim dep ls  t_deploy   # what t_deploy depends on (its prerequisites)
taskprim deps-of t_migrate  # what depends on t_migrate (its dependents)
```

**The frontier.** The payoff is `frontier`: the set of `open` tasks whose dependencies are *all* resolved (or that have none). It's the direct answer to "what can I work on right now?"

```bash
taskprim frontier --list ops --format json
```

Crucially, **resolving a task does not automatically unblock its dependents.** taskprim won't change any other task's state when you call `done`. That's deliberate — automatic cascading is planner logic, and taskprim is a primitive. The agent stays in control: resolve a task, then re-query the frontier to discover newly available work.

```bash
taskprim done t_migrate
taskprim frontier --list ops --format json   # t_deploy now appears
```

This "resolve, then re-query" loop is the canonical way an agent drives a body of dependent work to completion.

---

## Subtasks Are Not Dependencies

`--parent` and dependencies are orthogonal, and conflating them is a common mistake:

- **`--parent`** groups work hierarchically — a subtask is *part of* a larger task.
- **A dependency** gates execution — one task must finish before another starts.

A subtask does not automatically depend on its parent or its siblings. If one subtask must run before another, add an explicit `dep add`. Use parents to express structure and dependencies to express order; a task can have both.

```bash
taskprim add "Write migration"  --list ops --parent t_dbwork
taskprim add "Review migration" --list ops --parent t_dbwork
# Review depends on Write — express that explicitly
taskprim dep add t_review t_write
```

---

## Coordinating Multiple Agents: Seen-Tracking

When several agents share a list, each needs to know *what's new since it last looked* — without re-reading everything. taskprim tracks this per agent.

An agent marks tasks as seen, then later asks for what it hasn't seen:

```bash
# Agent "johanna" marks all open tasks in a list as seen
taskprim seen johanna --list ops

# Later: what has appeared that johanna hasn't seen?
taskprim list --unseen-by johanna --format json

# Or: what has johanna seen in the last day?
taskprim list --seen-by johanna --since 24h --format json
```

Seen-tracking is per-agent and additive — one agent marking a task seen doesn't affect another's view. This is how a watcher agent can poll a shared list and react only to genuinely new work.

---

## Finding Forgotten Work

Two filters surface tasks that may have stalled:

```bash
# Tasks explicitly blocked on something external
taskprim list --list ops --waiting --format json

# Tasks not touched in over a week (possibly abandoned)
taskprim list --list ops --stale 7d --format json
```

`stats` gives the high-level counts for a quick health read:

```bash
taskprim stats --format json   # { "total_open": 12, "total_done": 45, "total_killed": 3 }
```

Rising `total_open` against flat `total_done` is the same signal it is anywhere: work is arriving faster than it's being resolved.
