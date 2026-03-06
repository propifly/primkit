#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

CHECK_FLAG=""
if [[ "${1:-}" == "--check" ]]; then
  CHECK_FLAG="--check"
fi

echo "Extracting command metadata..."
for PRIM in queueprim taskprim stateprim knowledgeprim; do
  echo "  $PRIM"
  (cd "$ROOT/$PRIM" && go run ./cmd/docgen) > "$TMPDIR/$PRIM.json"
done

INPUTS="$TMPDIR/queueprim.json,$TMPDIR/taskprim.json,$TMPDIR/stateprim.json,$TMPDIR/knowledgeprim.json"

echo "Updating docs/agent-reference.md..."
(cd "$ROOT/primkit" && go run ./cmd/docupdater --inputs="$INPUTS" --doc="$ROOT/docs/agent-reference.md" $CHECK_FLAG)

echo "Done."
