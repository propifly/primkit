# /ship — Land a fix or feature to main

Run this command when you have a fix or feature ready to merge into `main`.
It walks through the full primkit pre-merge checklist and lands the change.

## When to use

Use `/ship` any time you've finished a fix, feature, or doc change and need to:
- Validate everything builds, tests, and docs are consistent
- Update the CHANGELOG
- Merge to main cleanly
- Confirm CI goes green on main

## Workflow (execute every step in order)

### 1. Branch check

Confirm you are on a **feature branch**, not `main`:

```bash
git branch --show-current
```

If on `main`, stop and ask the user which branch to work from.

### 2. Verify clean working tree or stage pending changes

```bash
git status
```

If there are unstaged changes relevant to the fix, stage and commit them before proceeding.

### 3. Run the full pre-merge suite

```bash
make all
```

This runs (in order): `tidy`, `fmt`, `lint`, `test`, `build`, `docs-check`, `check-registration`.

- If **any step fails**, fix it before continuing. Do not skip.
- `docs-check` will fail if command tables in `docs/agent-reference.md` are out of date — run `make docs` to regenerate, then re-run `make all`.
- `check-registration` will fail if a prim is missing from the Makefile, docgen loop, agent-reference, or cmd/docgen — fix each `FAIL` line before continuing.

### 4. Update CHANGELOG.md

Open `CHANGELOG.md` and add entries to the `[Unreleased]` section.

**What belongs in the changelog:**

| Section | What goes here |
|---------|---------------|
| `### Added` | New user-facing features, new prims, new CLI commands/flags |
| `### Fixed` | Bug fixes, correctness fixes, security fixes visible to operators |
| `### Changed` | Behaviour changes, breaking changes, default value changes, model updates |
| `### Deprecated` | Things being removed in a future release |

**What does NOT belong:** internal refactors with no operator-visible effect, CI-only changes, test-only changes, typo fixes in comments. Contributor tooling (scripts, doc-gen, check-registration) belongs under `### Changed` because it affects anyone working on the repo.

Format: `**scope**: description`. Scope is the prim name (`taskprim`, `knowledgeprim`, etc.), `all prims`, or a topic (`contributor tooling`, `security`).

Example:
```markdown
### Fixed

- **all prims**: `db.Open()` now rejects database paths whose filename
  contains `=`, catching the common misconfiguration where a `.env` parser
  concatenates a `KEY=VALUE` assignment onto the DB path and silently writes
  a token into the filesystem as part of a filename.
```

### 5. Update docs (if applicable)

If the change affects user-visible behaviour, commands, flags, or configuration:

- `docs/agent-reference.md` — command tables are auto-generated; run `make docs` if you changed CLI flags
- `docs/queueprim.md`, `docs/configuration.md`, etc. — hand-edit if behaviour documented there changed
- `README.md` / `llms.txt` — only if the top-level feature list changed

If nothing user-visible changed, skip this step.

### 6. Final commit on the branch

Stage everything and commit:

```bash
git add <files>
git commit -m "<type>: <summary>

<body>

Assisted-By: Claude <model>"
```

Commit type: `feat`, `fix`, `docs`, `chore`, `refactor`.
Keep the summary under 72 characters.
Body explains *why*, not *what*.

### 7. Push branch and check for conflicts

```bash
git push origin <branch>
git fetch origin main
git log HEAD..origin/main --oneline   # should be empty (no new commits on main)
git log origin/main..HEAD --oneline   # your commits to be merged
```

If `origin/main` has moved ahead, rebase:

```bash
git rebase origin/main
# resolve any conflicts, then:
git push --force-with-lease origin <branch>
```

### 8. Merge to main

Fast-forward merge (no merge commit):

```bash
git checkout main
git merge --ff-only <branch>
git push origin main
```

If `--ff-only` fails (diverged), go back to step 7 and rebase first.

### 9. Watch CI on main

```bash
gh run list --branch main --limit 3 --repo propifly/primkit
```

Get the run ID of the new push and watch it:

```bash
gh run view <run-id> --repo propifly/primkit
```

Poll every 30–60 seconds until all jobs show ✓. If any job fails:

1. Read the failure output: `gh run view <run-id> --log-failed --repo propifly/primkit`
2. Fix the issue on main directly (for trivial one-liners) or on a new branch (for anything non-trivial)
3. Push and watch again

**Common CI failures and fixes:**

| Symptom | Fix |
|---------|-----|
| `invalid version: git ls-remote … propifly/primkit` | `make tidy` left `primkit` as a direct dep in a prim's go.mod. Run `make tidy` — the Makefile strips it. |
| `docs are not up to date` | Run `make docs`, commit the updated `docs/agent-reference.md`. |
| `check-registration FAIL` | A prim is missing a registration point. Run `make check-registration` locally, fix each FAIL. |
| `go vet` / compile error | Fix in code, re-run `make all`. |

### 10. Done

Confirm the latest `main` commit matches what you merged and CI is green. Report status to the user.
