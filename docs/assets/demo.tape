# primkit demo — session-resilient state for AI agents
# Run: vhs docs/assets/demo.tape

Output docs/assets/demo.gif
Output docs/assets/demo.mp4

Set Shell bash
Set FontFamily "Menlo"
Set FontSize 14
Set Width 1200
Set Height 600
Set LetterSpacing 0
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 30ms
Set CursorBlink false

# ── Hidden setup ──────────────────────────────────────────────
Hide
Type "export TASKPRIM_DB=/tmp/primkit-demo-task.db"
Enter
Type "export STATEPRIM_DB=/tmp/primkit-demo-state.db"
Enter
Type "rm -f $TASKPRIM_DB $STATEPRIM_DB"
Enter
Sleep 500ms
Type 'export PS1="$ "'
Enter
Type "clear"
Enter
Sleep 500ms
Show

# ── Scene 1 — Agent works on a task ──────────────────────────

Sleep 500ms

Type "# primkit: CLI-native state that survives session death"
Enter
Sleep 2s

Type " "
Enter
Sleep 500ms

Type "# step 1: agent creates a task"
Enter
Sleep 1s
Type 'taskprim add "migrate users table to new schema" --list db-work'
Sleep 500ms
Enter
Sleep 3s

Type " "
Enter
Sleep 500ms

Type "# step 2: agent records progress"
Enter
Sleep 1s
Type "taskprim edit $(taskprim list --list db-work -f quiet) --context '4/8 steps done, fk ordering blocked'"
Sleep 500ms
Enter
Sleep 3s

Type " "
Enter
Sleep 500ms

Type "# step 3: agent stores structured context (blockers, tables, notes)"
Enter
Sleep 1s

# stateprim set — hidden because VHS can't render JSON double quotes in Type
Hide
Type "bash docs/assets/demo-setup.sh"
Enter
Sleep 2s
Show

Sleep 2s

Type " "
Enter
Sleep 500ms

# ── Scene 2 — Session dies ───────────────────────────────────

Type "# === process killed. session gone. ==="
Enter
Sleep 4s

# One clear — this IS the session death. Fresh terminal.
Hide
Type "clear"
Enter
Sleep 500ms
Show

Sleep 2s

# ── Scene 3 — New session, full resumption ───────────────────

Type "# new session. different process. what survived?"
Enter
Sleep 2s

Type " "
Enter
Sleep 500ms

Type "# check: are tasks still there?"
Enter
Sleep 1s
Type "taskprim list --list db-work --state open"
Sleep 500ms
Enter
Sleep 4s

Type " "
Enter
Sleep 500ms

Type "# check: does it remember the progress?"
Enter
Sleep 1s
Type "taskprim get $(taskprim list --list db-work -f quiet)"
Sleep 500ms
Enter
Sleep 5s

Type " "
Enter
Sleep 500ms

Type "# check: is the structured context intact?"
Enter
Sleep 1s
Type "stateprim get migration context -f json"
Sleep 500ms
Enter
Sleep 7s

Type " "
Enter
Sleep 500ms

# ── Scene 4 — Complete and close ─────────────────────────────

Type "# everything survived. finish the task."
Enter
Sleep 1s
Type "taskprim done $(taskprim list --list db-work -f quiet)"
Sleep 500ms
Enter
Sleep 3s

Type " "
Enter
Sleep 500ms

Type "# zero setup. zero infrastructure. it just works."
Enter
Sleep 4s
