---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: owner-questions
  role: primary
  status: draft
---

# Owner Questions

## Mode Context

This run is in intake mode. The current source is the Primkit prose brief, which already defines Primkit as local infrastructure primitives for AI agents and the humans working with them.

These questions do not ask what Primkit is. They ask for owner judgment on the choices the brief leaves open: who should feel most addressed, what promise should lead, what personality should dominate, what kinds of brands Primkit should or should not resemble, and which first-use contexts should shape direction work.

## Questions For The Owner

1. When Primkit speaks to its market, who should feel like the first audience?
   - Agentic coding environment builders.
   - Teams running multi-agent workflows.
   - Individual developers who want durable local state.
   - Another priority audience you have in mind.

2. If only one idea can lead the first impression, which should it be?
   - State that survives beyond the session.
   - Local infrastructure primitives for agents.
   - Four small CLIs backed by SQLite.
   - No hosted memory or always-on server.
   - Commands and files that can be inspected, scripted, and replayed.

3. Which personality trait should carry the most weight?
   - Local-first.
   - Sharp.
   - Durable.
   - Agent-native.

4. Which trait should be kept in check so the brand does not become unbalanced?
   - Too local-first could feel small or niche.
   - Too sharp could feel cold or aggressive.
   - Too durable could feel heavy or old-fashioned.
   - Too agent-native could feel trend-chasing or narrow.

5. When you say Primkit should not feel abstract, cute, or overbuilt, what is the most important line not to cross?
   - Avoid abstract category language that hides the practical tool.
   - Avoid mascot-like or playful expression.
   - Avoid enterprise design-system polish that makes it feel heavier than it is.
   - Avoid another specific pattern you have seen elsewhere.

6. What brands, developer tools, docs sites, or product experiences do you admire for Primkit's direction?
   - Name examples for tone, clarity, restraint, confidence, utility, or visual feel.
   - A reference can be admired for one narrow reason even if the whole brand is not a model.

7. What brands, developer tools, docs sites, or product experiences should Primkit clearly avoid resembling?
   - Name examples that feel too vague, too cute, too corporate, too crypto-coded, too academic, too sterile, or otherwise wrong.

8. Should Primkit feel more like a quiet utility, a serious infrastructure layer, or a sharp new developer category?
   - Choose the closest center of gravity, even if the final brand blends more than one.

9. How technical should the first layer of brand language feel?
   - Plain and outcome-led, with technical details one step later.
   - Technical from the start, because the right audience values specificity.
   - Balanced: clear promise first, concrete implementation proof immediately after.

10. What should Primkit make a capable developer feel in the first minute?
    - Relief that local agent state can be made durable.
    - Trust that the system is simple and inspectable.
    - Curiosity about composing the four primitives.
    - Confidence that this belongs in serious agent workflows.

11. Which application should most shape the first round of brand direction?
    - README and documentation.
    - CLI help and command examples.
    - Website or landing page.
    - Diagrams explaining the four primitives.
    - Repository, package, or launch materials.

12. If downstream work has to choose between being more memorable and being more plain, which side should it favor?
    - More memorable, as long as it stays practical.
    - More plain, as long as it does not become generic.
    - It depends on the context; specify where memorability matters most.

13. Are there words or phrases that should become Primkit signature language?
    - The brief already includes strong candidates such as "state that survives the session," "local infrastructure primitives," and "no server to run."
    - The owner call is whether any of these should become core brand language, or whether a different phrase should carry the brand.

14. Are there words, metaphors, or category labels Primkit should avoid?
    - For example: memory, agents, primitives, infrastructure, orchestration, database, framework, platform, toolkit, or operating system.
    - This is about brand connotation, not technical accuracy.

15. What unresolved judgment call would you not want downstream brand work to guess?
    - Use this to name any taste, audience, positioning, or credibility issue that would materially change the direction.

## Owner Answers

These answers incorporate the live `primkit-site` landing page as current brand
context. The site should be treated as real evidence of current positioning,
copy, visual language, and product proof, even though its design tokens are
explicitly labeled provisional.

Existing site and asset references to carry forward:

- Sibling repo root:
  `/Users/avergara/development/propifly/primkit-site`
- Landing page:
  `/Users/avergara/development/propifly/primkit-site/src/app/(home)/page.tsx`
- Terminal demo:
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
- Palette continuity: brand navy `#181848`, navy ink `#14152b`, brand cyan `#00d8f0`, AA-safe cyan `#0b7285`, cool paper `#fbfcfe`, soft paper `#f3f6fb`, wash `#e9edf6`.

1. Primary audience: developers and teams already using AI coding agents who need durable local state across sessions.

   Secondary audiences: agentic coding environment builders and multi-agent workflow operators.

   Primkit should not speak only to infrastructure builders. The landing page speaks to someone already running Claude Code, Cursor, Codex, shell loops, or cron jobs and looking for state that outlives a session.

2. The lead first impression should be: "Give your agents state that survives the session."

   Support it immediately with: "Four standalone CLIs. Four SQLite files. No server, SDK, API key, Postgres, Redis, or daemon."

   This is the strongest existing site line and should remain the main public promise unless downstream work finds a clearly stronger variant.

3. The dominant personality trait should be durable.

   A close second is explicit.

   The brand should feel like explicit local durability: commands, files, SQLite, inspectable state, and proof through real terminal examples.

4. The trait to keep in check is agent-native.

   Primkit should speak directly to agent workflows, but it should not feel like disposable AI hype. The durable local infrastructure story should outlast the current agent tooling cycle.

5. The most important line not to cross is becoming abstract or platform-y.

   Always explain Primkit through concrete commands, local files, SQLite-backed primitives, and agent workflows.

   The robot mark can stay, but it should be used as a restrained technical mark, not as a mascot-led voice or cute character system.

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
   - The four-question problem framing: what was I doing, did we already do this, what did we learn, what still needs to run.
   - The robot mark handling cube/database-like blocks in navy and cyan.
   - The interim design-token palette derived from that robot mark.

7. Anti-references:

   - Hosted AI memory platforms.
   - Enterprise "agent orchestration platform" branding.
   - Cute mascot-first developer tools.
   - Crypto/web3 infrastructure aesthetics.
   - Abstract AI productivity tools that promise magic instead of showing commands.
   - Generic developer SaaS pages with inflated claims and vague diagrams.

8. Primkit should feel like a serious local infrastructure layer with the ergonomics of a quiet CLI utility.

   It should not be a loud new category, a tiny CLI toy, or an enterprise platform. The right center of gravity is serious, local, inspectable, and calm.

9. The first layer of language should be balanced: clear promise first, concrete implementation proof immediately after.

   Preferred pattern: "State that survives the session. Four standalone CLIs, each backed by its own SQLite file. No daemon."

   Avoid opening with implementation details alone, but never let the promise sit without proof.

10. Primkit should make a capable developer feel trust that this is simple, inspectable, and real.

    Secondary feelings: relief that agent state can persist without a hosted memory system, and confidence that the tools compose in real shell workflows.

11. The first round of brand direction should prioritize the website landing page and docs.

    Secondary priority: CLI help and command examples.

    The site is already the main public brand surface, and the terminal demos are load-bearing proof. README and docs still matter, but the live site should shape the first visual and verbal system.

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

15. Do not let downstream work guess that Primkit should become more playful because of the robot mark.

    The robot is a technical symbol for agents handling durable blocks of state, not a mascot voice.

    Do not guess that Primkit is a general agent framework. It is the persistence layer underneath agents: explicit local tools, SQLite files, shell commands, and inspectable state.

    Do not discard the existing navy/cyan visual system casually. The current palette is provisional, but it is grounded in the robot mark and implemented in the site. Downstream work may refine it, but should treat a major visual departure as an intentional evolution that requires a clear reason.

    Compact owner stance: Primkit should feel like serious local infrastructure for agent workflows: durable, explicit, inspectable, and CLI-native. Keep the robot as a restrained technical mark, lead with state that survives the session, prove everything with real commands and SQLite files, preserve or deliberately evolve the current navy/cyan system, and avoid mascot cuteness, hosted-memory magic, and enterprise platform language.

## How Answers Are Used

Answers to audience and positioning questions will shape `brand-foundation`: audience priority, lead promise, message hierarchy, and the strategic center of gravity.

Answers to personality, language, application, and memorability questions will shape `brand-directions`: voice options, visual direction prompts, example copy, diagram tone, and first-use brand applications.

Answers about admired brands, disliked brands, anti-references, signature language, and avoided words will route through `brand-decisions`: they become owner verdicts that downstream direction work should follow rather than reinterpret.

Unanswered questions should remain open decisions. Downstream work can still proceed from the brief, but it should not silently invent owner taste where these answers would change the outcome.
