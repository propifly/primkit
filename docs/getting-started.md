# Getting Started

Install one binary, run three commands, and watch an agent write state that
outlives the terminal. primkit needs no server, no account, and no config file —
each primitive auto-creates its SQLite database on first use. This page gets you
to durable state in about a minute; the per-primitive guides and the
[agent reference](agent-reference.md) cover everything after that.

## Let your agent install it

The lowest-friction path: paste this into Claude Code, Cursor, Codex, or any
agent with shell access. It reads the docs, installs the primitives, and
verifies them — no manual steps on your end.

```
Read https://github.com/propifly/primkit/blob/main/README.md and install primkit for me:
the four primitives (taskprim, stateprim, knowledgeprim, queueprim) via go install, or the
pre-built binaries if Go isn't available. Then run a quick command to confirm each works.
```

Prefer to do it by hand? The manual steps are below.

## Install

The fastest path is `go install` (requires [Go 1.26+](https://go.dev/dl/)).
Install only the primitives you need:

```bash
go install github.com/propifly/primkit/taskprim/cmd/taskprim@latest
go install github.com/propifly/primkit/stateprim/cmd/stateprim@latest
go install github.com/propifly/primkit/knowledgeprim/cmd/knowledgeprim@latest
go install github.com/propifly/primkit/queueprim/cmd/queueprim@latest
```

Prefer a pre-built binary? Download one for your platform from
[the latest release](https://github.com/propifly/primkit/releases/latest) and
move it onto your `PATH`. The [README](../README.md) lists every install option,
including `curl` one-liners and building from source.

## Your first durable state

These three commands create a task, query what's open, and mark it done — the
full lifecycle, all from the shell. The database is created automatically at
`~/.taskprim/default.db` the first time you run `add`.

```bash
# 1. Create a task (source = who created it, typically an agent name)
taskprim add "Deploy v2 to staging" --list ops --source agent

# 2. List what's open
taskprim list --list ops --state open

# 3. Mark it done (use the ID printed in step 2)
taskprim done t_abc123
```

Close the terminal, open a new one, and run step 2 again — the state is still
there. That persistence across sessions is the whole point.

Every command speaks JSON for programmatic callers:

```bash
taskprim list --list ops --state open --format json
```

## Wire it into an agent

An agent with shell access uses primkit like any Unix tool — no SDK to import,
no framework to adopt. To expose the primitives over the Model Context Protocol
(for IDE agents such as Cursor or Claude Desktop), run a primitive in `mcp` mode:

```bash
taskprim mcp --transport stdio
```

The [README](../README.md) has ready-to-paste `.mcp.json` and
`claude_desktop_config.json` snippets for all four primitives.

## Where to go next

You now have one primitive running. The other three follow the same shape — a
single binary, its own SQLite file, CLI by default:

- **[taskprim Guide](taskprim.md)** — task lifecycle, lists and labels,
  dependency graphs, and the "what can I work on next?" frontier.
- **[stateprim Guide](stateprim.md)** — namespaced key-value state, dedup
  checks, and append-only logs.
- **[knowledgeprim Guide](knowledgeprim.md)** — a knowledge graph with hybrid
  full-text + vector search.
- **[queueprim Guide](queueprim.md)** — persistent work queues with priority,
  retries, and dead-letter handling.

When you want primitives to compose into agents that investigate open questions
and accumulate knowledge across sessions — not just react once — read
[Building Curious Agents](curiosity-architecture.md). For the exact command and
JSON surface an agent calls, see the [agent reference](agent-reference.md).
```

