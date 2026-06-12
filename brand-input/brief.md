# Primkit Brand Brief

## Product Summary

Primkit is a family of local infrastructure primitives for AI agents and the
humans working with them. It provides four small Go CLIs backed by embedded
SQLite: taskprim for task lifecycle and dependency frontiers, stateprim for
durable key-value state and append logs, knowledgeprim for searchable knowledge
graphs, and queueprim for persistent work queues.

The product is for agentic coding environments, multi-agent workflows, and
developers who need state that survives beyond one chat session or process. It
keeps the operating model explicit: agents write to local files through commands
that can be inspected, scripted, and replayed, instead of depending on hosted
memory or an always-on server.

The current public promise is practical and infrastructure-focused: state that
survives the session; four CLIs, four SQLite files, no server to run. The brand
should feel local-first, sharp, durable, and agent-native without becoming
abstract, cute, or overbuilt.

## Current Brand Evidence

This second source-map pass should treat the live `primkit-site` repository as
current brand evidence, not as an unrelated implementation detail. The site
contains the current public landing page, terminal proof demos, robot assets,
social preview, and interim design tokens. The tokens are explicitly provisional,
but they are still the active value-of-record for the site's current brand
expression and should be preserved or deliberately evolved rather than ignored.

The sibling site repo lives outside this Primkit product repo, so these paths are
absolute filesystem paths:

- Site repo root:
  `/Users/avergara/development/propifly/primkit-site`
- Landing page:
  `/Users/avergara/development/propifly/primkit-site/src/app/(home)/page.tsx`
- Hero terminal demo:
  `/Users/avergara/development/propifly/primkit-site/src/app/(home)/hero-terminal.tsx`
- Site navigation and hero copy:
  `/Users/avergara/development/propifly/primkit-site/docs-navigation.json`
- Robot mark:
  `/Users/avergara/development/propifly/primkit-site/public/logo.png`
- Social preview with robot and tagline:
  `/Users/avergara/development/propifly/primkit-site/public/social-preview.png`
- Interim design tokens:
  `/Users/avergara/development/propifly/primkit-site/.design/design-tokens.tokens.json`
- Implemented token CSS:
  `/Users/avergara/development/propifly/primkit-site/src/app/globals.css`
- Site design note:
  `/Users/avergara/development/propifly/primkit-site/.design/design.md`

Observed current site language:

- Hero title: "Give your agents state that survives the session."
- Hero support: "Four standalone CLIs - tasks, operational state, knowledge, and queues - each persisting to its own SQLite file. Install once, call from any shell or agent loop. No Postgres, no Redis, no daemon."
- Problem heading: "The agent resets. The work shouldn't."
- Composition heading: "Four tools, one agent loop."
- Why-Primkit heading: "Explicit, not automatic. A file, not a fleet."
- Proof mode: real terminal examples for taskprim, stateprim, knowledgeprim, and queueprim.
- Trust strip: MIT licensed; pure Go, single binaries; works with Claude Code, Cursor, Codex, OpenClaw, and Hermes via CLI, MCP, or HTTP.
- Closing promise: "Stop re-explaining your project every session."

Observed current visual language:

- The robot mark is a blocky navy robot handling cube/database-like blocks with
  cyan glow accents.
- The social preview uses the robot, dark navy background, white `primkit`
  wordmark, and cyan line: "Task management and state persistence for AI
  agents."
- The interim token source says the palette is derived from the robot mark:
  brand navy `#181848`, navy ink `#14152b`, brand cyan `#00d8f0`, AA-safe cyan
  `#0b7285`, cool paper `#fbfcfe`, soft paper `#f3f6fb`, wash `#e9edf6`.
- Cyan should be used for fills, focus, and small marks, not body text; the
  token source uses `#0b7285` for AA-safe accent text on light surfaces.

## Owner Answers From First Source-Map Pass

1. Primary audience: developers and teams already using AI coding agents who
   need durable local state across sessions.

   Secondary audiences: agentic coding environment builders and multi-agent
   workflow operators.

   Primkit should not speak only to infrastructure builders. The landing page
   speaks to someone already running Claude Code, Cursor, Codex, shell loops, or
   cron jobs and looking for state that outlives a session.

2. The lead first impression should be: "Give your agents state that survives
   the session."

   Support it immediately with: "Four standalone CLIs. Four SQLite files. No
   server, SDK, API key, Postgres, Redis, or daemon."

   This is the strongest existing site line and should remain the main public
   promise unless downstream work finds a clearly stronger variant.

3. The dominant personality trait should be durable.

   A close second is explicit.

   The brand should feel like explicit local durability: commands, files,
   SQLite, inspectable state, and proof through real terminal examples.

4. The trait to keep in check is agent-native.

   Primkit should speak directly to agent workflows, but it should not feel like
   disposable AI hype. The durable local infrastructure story should outlast the
   current agent tooling cycle.

5. The most important line not to cross is becoming abstract or platform-y.

   Always explain Primkit through concrete commands, local files,
   SQLite-backed primitives, and agent workflows.

   The robot mark can stay, but it should be used as a restrained technical
   mark, not as a mascot-led voice or cute character system.

6. Admired references:

   - SQLite for durability, trust, and boring-in-the-best-way infrastructure.
   - Git for local-first inspectability and durable project memory.
   - ripgrep for sharp CLI utility and plain developer credibility.
   - Tailscale docs for practical technical clarity.
   - Fly.io for confident infrastructure copy without enterprise weight.
   - Unix tool philosophy for small tools that compose.

   Existing Primkit references to preserve and evaluate:

   - The current landing-page hero promise.
   - The tabbed terminal demo using real command output.
   - The four-question problem framing: what was I doing, did we already do
     this, what did we learn, what still needs to run.
   - The robot mark handling cube/database-like blocks in navy and cyan.
   - The interim design-token palette derived from that robot mark.

7. Anti-references:

   - Hosted AI memory platforms.
   - Enterprise "agent orchestration platform" branding.
   - Cute mascot-first developer tools.
   - Crypto/web3 infrastructure aesthetics.
   - Abstract AI productivity tools that promise magic instead of showing
     commands.
   - Generic developer SaaS pages with inflated claims and vague diagrams.

8. Primkit should feel like a serious local infrastructure layer with the
   ergonomics of a quiet CLI utility.

   It should not be a loud new category, a tiny CLI toy, or an enterprise
   platform. The right center of gravity is serious, local, inspectable, and
   calm.

9. The first layer of language should be balanced: clear promise first, concrete
   implementation proof immediately after.

   Preferred pattern: "State that survives the session. Four standalone CLIs,
   each backed by its own SQLite file. No daemon."

   Avoid opening with implementation details alone, but never let the promise
   sit without proof.

10. Primkit should make a capable developer feel trust that this is simple,
    inspectable, and real.

    Secondary feelings: relief that agent state can persist without a hosted
    memory system, and confidence that the tools compose in real shell
    workflows.

11. The first round of brand direction should prioritize the website landing
    page and docs.

    Secondary priority: CLI help and command examples.

    The site is already the main public brand surface, and the terminal demos
    are load-bearing proof. README and docs still matter, but the live site
    should shape the first visual and verbal system.

12. Favor plainness, but keep the sharp existing phrases.

    Strong current phrases to preserve or refine:

    - "Give your agents state that survives the session."
    - "The agent resets. The work shouldn't."
    - "Explicit, not automatic. A file, not a fleet."
    - "Four tools, one agent loop."
    - "Composable like any Unix tool."

    Primkit should be memorable through precision, not cleverness.

13. Signature language:

    - "State that survives the session."
    - "The agent resets. The work shouldn't."
    - "Four standalone CLIs. Four SQLite files. No daemon."
    - "Explicit, not automatic. A file, not a fleet."
    - "Composable like any Unix tool."
    - "Persistent state for AI agents."

    Best primary line: "Give your agents state that survives the session."

14. Avoid these as primary category labels:

    - platform
    - operating system
    - orchestration
    - framework
    - hosted memory
    - autonomous memory
    - AI workspace
    - agent brain

    Terms that are allowed when precise:

    - primitives
    - infrastructure
    - persistent state
    - local state
    - durable state
    - task frontier
    - knowledge graph
    - queue

15. Do not let downstream work guess that Primkit should become more playful
    because of the robot mark.

    The robot is a technical symbol for agents handling durable blocks of state,
    not a mascot voice.

    Do not guess that Primkit is a general agent framework. It is the
    persistence layer underneath agents: explicit local tools, SQLite files,
    shell commands, and inspectable state.

    Do not discard the existing navy/cyan visual system casually. The current
    palette is provisional, but it is grounded in the robot mark and implemented
    in the site. Downstream work may refine it, but should treat a major visual
    departure as an intentional evolution that requires a clear reason.

    Compact owner stance: Primkit should feel like serious local infrastructure
    for agent workflows: durable, explicit, inspectable, and CLI-native. Keep
    the robot as a restrained technical mark, lead with state that survives the
    session, prove everything with real commands and SQLite files, preserve or
    deliberately evolve the current navy/cyan system, and avoid mascot cuteness,
    hosted-memory magic, and enterprise platform language.

## Source-Map Instructions For This Pass

Treat this as archaeology mode, not thin intake mode. The original prose brief,
the current live site, the robot PNGs, the social preview, the interim token
file, the implemented CSS token mapping, and the owner answers are all supplied
brand material.

The site evidence is not final identity by default. Classify it carefully:

- Use `Confirmed` for owner decisions stated above and for product facts already
  in the brief.
- Use `De facto` for current public site behavior, site copy, robot asset usage,
  and implemented token values that are observable but still labeled
  provisional.
- Use `Missing` where production-ready identity still needs downstream owner
  approval, including final voice system, final visual identity, source-grade
  asset package, licensing/ownership, and application rules.
- Do not invent new visual directions in source-map. Carry existing evidence
  forward so `brand-foundation`, `brand-directions`, `brand-decisions`,
  `brand-verbal-identity`, and `brand-visual-identity` can make explicit choices.
