#!/usr/bin/env bash
# check-registration.sh — Verifies every prim registered in go.work is wired
# into all the right places: Makefile targets, docgen.sh, agent-reference.md,
# and cmd/docgen/main.go.
#
# Usage: bash scripts/check-registration.sh
# Exit:  0 if all checks pass, 1 if any fail.
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)

# ── Color support ──────────────────────────────────────────────────────────────
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' BOLD='' RESET=''
fi

PASS=0
FAIL=0

pass() { echo -e "  ${GREEN}PASS${RESET}  $1"; PASS=$((PASS + 1)); }
fail() { echo -e "  ${RED}FAIL${RESET}  $1"; FAIL=$((FAIL + 1)); }

# ── Parse prims from go.work ───────────────────────────────────────────────────
# Extract lines like `\t./queueprim` from the use (...) block, strip leading
# whitespace and ./ prefix, then exclude `primkit` (shared lib, not a prim).
PRIMS=()
while IFS=$'\t ' read -r token rest; do
  # token is the first word after stripping leading tabs/spaces
  if [[ "$token" == ./* ]]; then
    name="${token#./}"
    if [[ "$name" != "primkit" ]]; then
      PRIMS+=("$name")
    fi
  fi
done < "$ROOT/go.work"

if [[ ${#PRIMS[@]} -eq 0 ]]; then
  echo -e "${RED}ERROR:${RESET} No prims found in go.work — is the file correct?"
  exit 1
fi

echo -e "${BOLD}Checking registration for: ${PRIMS[*]}${RESET}"
echo

# ── Helper: check that a file contains a fixed string ─────────────────────────
file_contains() {
  local file="$1" pattern="$2"
  grep -qF "$pattern" "$file" 2>/dev/null
}

# ── Per-prim checks ────────────────────────────────────────────────────────────
for PRIM in "${PRIMS[@]}"; do
  echo -e "${BOLD}${PRIM}${RESET}"

  # 1. Makefile: build target
  if file_contains "$ROOT/Makefile" "cd ${PRIM} && go build"; then
    pass "Makefile build target"
  else
    fail "Makefile build target — missing: cd ${PRIM} && go build"
  fi

  # 2. Makefile: test target
  if file_contains "$ROOT/Makefile" "cd ${PRIM} && go test"; then
    pass "Makefile test target"
  else
    fail "Makefile test target — missing: cd ${PRIM} && go test"
  fi

  # 3. Makefile: lint target
  if file_contains "$ROOT/Makefile" "cd ${PRIM} && golangci-lint run"; then
    pass "Makefile lint target"
  else
    fail "Makefile lint target — missing: cd ${PRIM} && golangci-lint run"
  fi

  # 4. Makefile: fmt target
  if file_contains "$ROOT/Makefile" "cd ${PRIM} && gofumpt"; then
    pass "Makefile fmt target"
  else
    fail "Makefile fmt target — missing: cd ${PRIM} && gofumpt"
  fi

  # 5. Makefile: tidy target
  # The tidy target uses a for-loop: "for prim in taskprim stateprim ..."
  # Check that the prim appears on the for-loop line inside the tidy target.
  if grep -q "^	for prim in.*\b${PRIM}\b" "$ROOT/Makefile"; then
    pass "Makefile tidy target"
  else
    fail "Makefile tidy target — ${PRIM} missing from tidy for-loop in Makefile"
  fi

  # 6. scripts/docgen.sh: prim in for loop
  if grep -q "for PRIM in .*\b${PRIM}\b" "$ROOT/scripts/docgen.sh" 2>/dev/null; then
    pass "scripts/docgen.sh loop"
  else
    fail "scripts/docgen.sh loop — ${PRIM} missing from 'for PRIM in ...' line"
  fi

  # 7. docs/agent-reference.md: docgen start anchor
  if file_contains "$ROOT/docs/agent-reference.md" "<!-- docgen:start:${PRIM}:commands -->"; then
    pass "docs/agent-reference.md anchor"
  else
    fail "docs/agent-reference.md anchor — missing: <!-- docgen:start:${PRIM}:commands -->"
  fi

  # 8. cmd/docgen/main.go exists
  if [[ -f "$ROOT/${PRIM}/cmd/docgen/main.go" ]]; then
    pass "${PRIM}/cmd/docgen/main.go exists"
  else
    fail "${PRIM}/cmd/docgen/main.go — file not found"
  fi

  echo
done

# ── Summary ────────────────────────────────────────────────────────────────────
TOTAL=$((PASS + FAIL))
echo -e "${BOLD}Summary:${RESET} ${GREEN}${PASS}${RESET} passed, ${RED}${FAIL}${RESET} failed (${TOTAL} total)"

if [[ $FAIL -gt 0 ]]; then
  echo -e "${YELLOW}Run 'bash scripts/new-prim.sh <name>' to scaffold a new prim with all registrations.${RESET}"
  exit 1
fi

exit 0
