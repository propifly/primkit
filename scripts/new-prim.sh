#!/usr/bin/env bash
# new-prim.sh — Scaffold a new prim and register it everywhere.
#
# Usage: bash scripts/new-prim.sh <name>
# Example: bash scripts/new-prim.sh fooprim
#
# After running this script, the new prim will pass make check-registration.
# See the printed instructions at the end for manual follow-up steps.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)

# ── Args ───────────────────────────────────────────────────────────────────────
NAME="${1:-}"
if [[ -z "$NAME" ]]; then
  echo "Usage: bash scripts/new-prim.sh <name>"
  echo "Example: bash scripts/new-prim.sh fooprim"
  exit 1
fi

# Validate name: lowercase letters and digits only, must end in 'prim'
if ! [[ "$NAME" =~ ^[a-z][a-z0-9]*prim$ ]]; then
  echo "Error: prim name must be lowercase alphanumeric and end in 'prim' (e.g. fooprim)"
  exit 1
fi

PRIM_DIR="$ROOT/$NAME"
MODULE="github.com/propifly/primkit/$NAME"

echo "Scaffolding new prim: $NAME"
echo "  Module:    $MODULE"
echo "  Directory: $PRIM_DIR"
echo

# ── Directory structure ────────────────────────────────────────────────────────
if [[ -d "$PRIM_DIR" ]]; then
  echo "  [SKIP] Directory $NAME/ already exists — skipping file creation"
else
  echo "  [CREATE] $NAME/ directory structure"
  mkdir -p "$PRIM_DIR/cmd/$NAME"
  mkdir -p "$PRIM_DIR/cmd/docgen"
  mkdir -p "$PRIM_DIR/internal/cli"

  # ── go.mod ──────────────────────────────────────────────────────────────────
  cat > "$PRIM_DIR/go.mod" << EOF
module $MODULE

go 1.26

require (
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	modernc.org/sqlite v1.46.1
)
EOF

  # ── cmd/<name>/main.go ──────────────────────────────────────────────────────
  cat > "$PRIM_DIR/cmd/$NAME/main.go" << EOF
package main

import (
	"fmt"
	"os"

	"$MODULE/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
EOF

  # ── cmd/docgen/main.go ──────────────────────────────────────────────────────
  # Pattern mirrors queueprim/cmd/docgen/main.go with name changed.
  cat > "$PRIM_DIR/cmd/docgen/main.go" << EOF
// Command docgen extracts $NAME command metadata and writes it as JSON to stdout.
// Used by scripts/docgen.sh to generate the commands table in docs/agent-reference.md.
// Local types mirror primkit/docgen.{Prim,Cmd,Flag}Meta — JSON field names must match.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"$MODULE/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type primMeta struct {
	Name     string    \`json:"name"\`
	Commands []cmdMeta \`json:"commands"\`
}

type cmdMeta struct {
	Name     string     \`json:"name"\`
	Synopsis string     \`json:"synopsis"\`
	Short    string     \`json:"short"\`
	Flags    []flagMeta \`json:"flags"\`
}

type flagMeta struct {
	Name     string \`json:"name"\`
	Usage    string \`json:"usage"\`
	Default  string \`json:"default"\`
	Required bool   \`json:"required"\`
}

var skipCommands = map[string]bool{
	"serve":      true,
	"mcp":        true,
	"completion": true,
	"help":       true,
}

func main() {
	root := cli.NewRootCmd()
	meta := primMeta{
		Name:     "$NAME",
		Commands: extractCommands(root.Commands(), ""),
	}
	if err := json.NewEncoder(os.Stdout).Encode(meta); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func extractCommands(cmds []*cobra.Command, prefix string) []cmdMeta {
	var result []cmdMeta
	for _, cmd := range cmds {
		if skipCommands[cmd.Name()] || cmd.Hidden {
			continue
		}
		fullName := cmd.Name()
		if prefix != "" {
			fullName = prefix + " " + cmd.Name()
		}
		synopsis := buildSynopsis(cmd.Use, fullName)
		if cmd.RunE != nil || cmd.Run != nil {
			m := cmdMeta{
				Name:     fullName,
				Synopsis: synopsis,
				Short:    cmd.Short,
			}
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Hidden {
					return
				}
				required := len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0
				m.Flags = append(m.Flags, flagMeta{
					Name:     f.Name,
					Usage:    f.Usage,
					Default:  f.DefValue,
					Required: required,
				})
			})
			result = append(result, m)
		}
		result = append(result, extractCommands(cmd.Commands(), fullName)...)
	}
	return result
}

func buildSynopsis(use, fullName string) string {
	parts := strings.SplitN(use, " ", 2)
	if len(parts) <= 1 || parts[1] == "" {
		return fullName
	}
	return fullName + " " + parts[1]
}
EOF

  # ── internal/cli/root.go ────────────────────────────────────────────────────
  cat > "$PRIM_DIR/internal/cli/root.go" << EOF
// Package cli implements the $NAME command-line interface using cobra.
// Each subcommand lives in its own file and follows the pattern:
// parse flags → call store → format output.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the top-level $NAME command with all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "$NAME",
		Short:         "TODO: short description of $NAME",
		Long:          \`TODO: long description of $NAME.\`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("db", "", "path to SQLite database")
	root.PersistentFlags().String("config", "", "path to config file")
	root.PersistentFlags().StringP("format", "f", "text", "output format: text, json")

	root.AddCommand(
		newServeCmd(),
		newMCPCmd(),
		newRestoreCmd(),
	)

	return root
}
EOF

  # ── internal/cli/serve.go ───────────────────────────────────────────────────
  cat > "$PRIM_DIR/internal/cli/serve.go" << EOF
package cli

import (
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement HTTP server
			return nil
		},
	}
	cmd.Flags().IntP("port", "p", 8095, "HTTP port to listen on")
	return cmd
}
EOF

  # ── internal/cli/mcp.go ─────────────────────────────────────────────────────
  cat > "$PRIM_DIR/internal/cli/mcp.go" << EOF
package cli

import (
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP (Model Context Protocol) server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement MCP server
			return nil
		},
	}
	cmd.Flags().StringP("transport", "t", "stdio", "transport: stdio or sse")
	cmd.Flags().IntP("port", "p", 8096, "port for SSE transport")
	return cmd
}
EOF

  # ── internal/cli/restore.go ─────────────────────────────────────────────────
  cat > "$PRIM_DIR/internal/cli/restore.go" << EOF
package cli

import (
	"github.com/spf13/cobra"
)

func newRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restore database from a replica",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement restore
			return nil
		},
	}
}
EOF

fi

# ── go.work: add ./name ────────────────────────────────────────────────────────
if grep -qF "./$NAME" "$ROOT/go.work"; then
  echo "  [SKIP] go.work already contains ./$NAME"
else
  echo "  [UPDATE] go.work — adding ./$NAME"
  # Insert before the closing paren of the use block
  # Use a temp file approach compatible with macOS sed
  TMPFILE=$(mktemp)
  while IFS= read -r line; do
    if [[ "$line" == ")" ]]; then
      printf '\t./%s\n' "$NAME" >> "$TMPFILE"
    fi
    printf '%s\n' "$line" >> "$TMPFILE"
  done < "$ROOT/go.work"
  mv "$TMPFILE" "$ROOT/go.work"
fi

# ── Makefile: add lines to each target ────────────────────────────────────────
# We detect the last existing prim entry in each target and append after it.
# Strategy: find the last `cd <existingprim> && go build` line per target and
# insert the new line after it, using Python for reliable multi-line editing
# on both macOS and Linux.

add_to_makefile() {
  local target_keyword="$1"   # e.g. "go build"
  local new_line="$2"         # full line to insert, e.g. "\tcd fooprim && go build ..."

  if grep -qF "cd $NAME && $target_keyword" "$ROOT/Makefile"; then
    echo "  [SKIP] Makefile $target_keyword — already has $NAME"
    return
  fi

  echo "  [UPDATE] Makefile $target_keyword target — adding $NAME"

  python3 - "$ROOT/Makefile" "$target_keyword" "$new_line" << 'PYEOF'
import sys, re

makefile_path = sys.argv[1]
keyword = sys.argv[2]
new_line = sys.argv[3]

with open(makefile_path) as f:
    content = f.read()

lines = content.split('\n')
# Find the last line containing the keyword (in a recipe context)
last_idx = -1
for i, line in enumerate(lines):
    if keyword in line and line.startswith('\t'):
        last_idx = i

if last_idx == -1:
    print(f"WARNING: could not find '{keyword}' in Makefile — skipping", file=sys.stderr)
    sys.exit(0)

lines.insert(last_idx + 1, new_line)
with open(makefile_path, 'w') as f:
    f.write('\n'.join(lines))
PYEOF
}

# build target: insert after last `go build` line
add_to_makefile "go build" "	cd $NAME && go build -o ../bin/$NAME ./cmd/$NAME"

# build-pi target: insert after last `GOOS=linux GOARCH=arm64 go build` line
if grep -qF "cd $NAME && GOOS=linux" "$ROOT/Makefile"; then
  echo "  [SKIP] Makefile build-pi — already has $NAME"
else
  echo "  [UPDATE] Makefile build-pi target — adding $NAME"
  python3 - "$ROOT/Makefile" "GOARCH=arm64 go build" \
    "	cd $NAME && GOOS=linux GOARCH=arm64 go build -o ../bin/${NAME}-linux-arm64 ./cmd/$NAME" << 'PYEOF'
import sys
makefile_path = sys.argv[1]
keyword = sys.argv[2]
new_line = sys.argv[3]
with open(makefile_path) as f:
    content = f.read()
lines = content.split('\n')
last_idx = -1
for i, line in enumerate(lines):
    if keyword in line and line.startswith('\t'):
        last_idx = i
if last_idx == -1:
    print(f"WARNING: could not find '{keyword}' in Makefile — skipping", file=sys.stderr)
    import sys as _sys; _sys.exit(0)
lines.insert(last_idx + 1, new_line)
with open(makefile_path, 'w') as f:
    f.write('\n'.join(lines))
PYEOF
fi

# test target
add_to_makefile "go test" "	cd $NAME && go test -v -race -count=1 ./..."

# lint target
add_to_makefile "golangci-lint run" "	cd $NAME && golangci-lint run ./..."

# fmt target
add_to_makefile "gofumpt" "	cd $NAME && gofumpt -w ."

# tidy target
# Note: queueprim has special logic to strip primkit from go.mod (see its tidy entry).
# New prims don't import primkit initially, so we use a plain go mod tidy.
add_to_makefile "go mod tidy" "	cd $NAME && go mod tidy"

# ── scripts/docgen.sh: add to for loop ────────────────────────────────────────
if grep -q "\b${NAME}\b" "$ROOT/scripts/docgen.sh"; then
  echo "  [SKIP] scripts/docgen.sh — already has $NAME"
else
  echo "  [UPDATE] scripts/docgen.sh — adding $NAME to for loop and INPUTS"
  python3 - "$ROOT/scripts/docgen.sh" "$NAME" << 'PYEOF'
import sys, re

path = sys.argv[1]
name = sys.argv[2]

with open(path) as f:
    content = f.read()

# Add to `for PRIM in ...` line
content = re.sub(
    r'(for PRIM in\b[^\n]+)',
    lambda m: m.group(0) + ' ' + name,
    content
)

# Add to INPUTS line: find last .json entry and add after it
content = re.sub(
    r'(INPUTS="[^"]+?)(")',
    lambda m: m.group(1) + ',$TMPDIR/' + name + '.json' + m.group(2),
    content
)

with open(path, 'w') as f:
    f.write(content)
PYEOF
fi

# ── docs/agent-reference.md: add section before ## Error Patterns ─────────────
if grep -qF "<!-- docgen:start:${NAME}:commands -->" "$ROOT/docs/agent-reference.md"; then
  echo "  [SKIP] docs/agent-reference.md — already has $NAME section"
else
  echo "  [UPDATE] docs/agent-reference.md — inserting $NAME section"
  python3 - "$ROOT/docs/agent-reference.md" "$NAME" << 'PYEOF'
import sys

path = sys.argv[1]
name = sys.argv[2]

with open(path) as f:
    content = f.read()

new_section = f"""## {name}

TODO: describe {name}.

### Commands

<!-- docgen:start:{name}:commands -->
| Command | Synopsis | Flags |
|---------|----------|-------|
<!-- docgen:end:{name}:commands -->

---

"""

# Insert before ## Error Patterns
if '## Error Patterns' not in content:
    print("WARNING: '## Error Patterns' not found in agent-reference.md — appending section at end", file=sys.stderr)
    content = content.rstrip() + '\n\n' + new_section
else:
    content = content.replace('## Error Patterns', new_section + '## Error Patterns', 1)

with open(path, 'w') as f:
    f.write(content)
PYEOF
fi

# ── Done ───────────────────────────────────────────────────────────────────────
echo
echo "Scaffold complete for $NAME."
echo
echo "========================================================================="
echo "Next steps (manual):"
echo "========================================================================="
echo ""
echo "1. Implement commands in $NAME/internal/cli/"
echo "   Add a store (internal/store/), model (internal/model/), and wire"
echo "   them into root.go following the pattern in taskprim or queueprim."
echo ""
echo "2. Run: make tidy"
echo "   This will download dependencies and update go.work.sum."
echo ""
echo "3. Run: make all"
echo "   Fails until the implementation compiles and tests pass."
echo "   make check-registration should pass immediately."
echo ""
echo "4. Update these files manually (narrative docs, not machine-checkable):"
echo "   - README.md         — add $NAME to the primitives table"
echo "   - llms.txt          — add $NAME entry"
echo "   - REPOS.md          — update top-level structure column"
echo ""
echo "5. Run: make docs"
echo "   After implementing commands so the auto-generated table is populated."
echo "========================================================================="
