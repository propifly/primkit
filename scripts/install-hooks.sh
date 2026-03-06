#!/usr/bin/env bash
# install-hooks.sh — Installs git hooks for primkit development.
# Run once after cloning: bash scripts/install-hooks.sh
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
HOOK="$ROOT/.git/hooks/pre-commit"

cat > "$HOOK" << 'HOOK_EOF'
#!/usr/bin/env bash
# Pre-commit hook: verify docs are up to date before committing.
set -euo pipefail
ROOT=$(git rev-parse --show-toplevel)
if ! bash "$ROOT/scripts/docgen.sh" --check 2>&1; then
  echo ""
  echo "  Docs are out of date. Run: make docs"
  echo "  Then re-stage and commit."
  exit 1
fi
HOOK_EOF

chmod +x "$HOOK"
echo "Installed pre-commit hook at $HOOK"
echo "Run 'bash scripts/install-hooks.sh' after cloning to enable."
