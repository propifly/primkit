---
hatchable:
  domain: brand-foundation
  root: canon
  lane: brand
  artifact: promise-and-principles
  role: primary
  status: draft
---

# Primkit Promise And Principles

## Brand Promise

`Confirmed`: Primkit gives AI agents local durable state that survives the session.

The promise is practical: agents and the humans working with them can keep task, state, knowledge, and queue context across resets, shell loops, and multi-session work in local workflows. Primkit does this through four small Go CLIs backed by embedded SQLite, not through a hosted memory layer or opaque agent platform.

Scoped short-form promise:

> Give your agents state that survives the session.

Promise boundary:

- `Confirmed`: Primkit provides local infrastructure primitives for AI agents and humans working with them.
- `Confirmed`: Primkit supports durable task, state, knowledge, and queue context through `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`.
- `Confirmed`: The operating model is local, explicit, inspectable, scriptable, replayable, CLI-native, and proof-backed.
- `Confirmed`: The core promise does not require a server, SDK, API key, Postgres, Redis, or daemon.
- `Rejected`: The promise must not become an unsupported AI productivity claim, a hosted memory claim, a general agent framework claim, or a broad orchestration/platform claim.

Classification basis:

- This downstream slice does not include dependency decision files or source-map artifacts, so no external decision ID is treated as evidence here.
- The confirmed promise is supported only by the in-scope product facts stated above: Primkit is four small Go CLIs, the primitives are `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`, and the state model is local, explicit, inspectable, scriptable, replayable, CLI-native, and SQLite-backed.
- The rejected boundary is supported only as a scope guard for this artifact: do not turn the promise into claims about hosted memory, broad orchestration, general agent frameworks, or unsupported AI productivity benefits without separate authoritative proof.

## Values

| Value | Classification | Meaning |
| --- | --- | --- |
| Local | `Confirmed` | Primkit's center of gravity is local files, local commands, and SQLite-backed state. |
| Durable | `Confirmed` | Work survives agent sessions instead of depending on conversational memory alone. |
| Explicit | `Confirmed` | State and workflow actions are visible as commands, files, and named primitives. |
| Inspectable | `Confirmed` | Humans and agents can inspect what exists rather than trust an opaque memory layer. |
| Scriptable | `Confirmed` | Primkit fits shell, CLI, and agent-loop automation without becoming a hidden service. |
| Replayable | `Confirmed` | Work is reconstructable from persisted state and command-level evidence. |
| CLI-native | `Confirmed` | Terminal use and command composition are first-class parts of the brand strategy, not secondary wrappers. |
| Proof-backed | `Confirmed` | Claims are paired with concrete commands, SQLite files, terminal examples, or equivalent product proof. |

These values are strategy facts. They set decision boundaries for downstream work, but they do not specify execution systems, production asset rules, or surface-specific application rules.

## Principles

| Principle | Classification | Practical rule |
| --- | --- | --- |
| Lead with surviving state. | `Confirmed` | Start from the fact that agents can keep local state across sessions. |
| Prove the promise immediately. | `Confirmed` | Support the promise with commands, SQLite-backed primitives, and concrete product evidence. |
| Name the primitives. | `Confirmed` | Use `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim` when the claim depends on product architecture. |
| Keep the category narrow. | `Confirmed` | Frame Primkit as local infrastructure primitives, not as a platform, framework, workspace, agent brain, or hosted memory product. |
| Keep state inspectable. | `Confirmed` | Prefer claims that show where state lives and how humans or agents can inspect it. |
| Keep work scriptable. | `Confirmed` | Treat CLI use, shell composition, and explicit commands as core product behavior. |
| Keep claims replayable. | `Confirmed` | Avoid claims that cannot be checked again through files, commands, or recorded product behavior. |
| Reject unsupported benefit claims. | `Rejected` | Do not imply productivity, autonomy, intelligence, or memory benefits unless they are tied to concrete local primitives and evidence. |

These principles are strategic guardrails, not approved production application rules for website, docs, README, CLI help, demos, social previews, product descriptions, or future product surfaces.

## Behavioral Implications

- Downstream work can use the promise as a strategic priority, but this artifact does not approve where or how any production surface must present it.
- Terminology stays evidence-bound: terms such as primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue are approved descriptors only when accurate, not final copy instructions for a specific surface.
- Claims about breadth, compatibility, license, packaging, or production readiness are not promoted from current site copy into approved brand claims without proof.
- Current trust-strip claims stay proof-required unless a later owner verdict or authoritative source-map capture confirms them.
- Unsupported claims become missing decisions, not softer versions of the promise.
- Rejected category labels are hard constraints, not alternate messaging options.
- Future surfaces still need separate approval for the amount, placement, and format of product evidence; this artifact only establishes that the promise should remain evidence-bound.
- No website, docs, README, CLI help, demos, social preview, product description, or future product surface application rule is approved here; those instructions remain missing decisions.

Classification basis:

- These implications are derived from the promise, values, and principles in this artifact, not from external decision IDs.
- Trust-strip, unsupported-benefit, category-label, breadth, compatibility, license, packaging, and production-readiness claims remain proof-required unless a later source-map capture or owner verdict supplies authoritative support.

## Missing Decisions

`Missing`: Final production copy rules are not approved. Current phrases and source-listed signature lines can inform downstream work, but they are not all canonical brand copy.

Support boundary:

- This slice includes no final production copy source or owner approval, so production copy rules remain missing.

`Missing`: Production application rules are not approved for website, docs, README, CLI help, demos, social previews, product descriptions, or future product surfaces.

Support boundary:

- This slice includes no production application source-map or owner approval for website, docs, README, CLI help, demos, social previews, product descriptions, or future product surfaces.
- Surface-specific evidence depth, examples, sequencing, and credibility proof requirements are part of those missing production application rules.

`Missing`: The exact first-round audience emphasis between individual developers and teams remains unresolved. The confirmed audience includes both.

Support boundary:

- This slice includes no audience-priority decision that resolves whether to emphasize individual developers or teams first.

`Missing`: Trust-strip claims and other current site claims need proof or owner confirmation before they become approved brand claims.

Support boundary:

- This slice includes no authoritative trust-strip proof or owner verdict, so these claims cannot become approved brand claims here.

Scope note: this run's declared output scope is limited, so missing decisions are recorded here rather than emitted as sibling decision files.
