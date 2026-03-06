#!/usr/bin/env bash
# sync-agent-docs.sh — Package the docs an agent needs from primkit into a zip.
#
# Usage:
#   bash scripts/sync-agent-docs.sh [--out DIR]
#
# Output: dist/primkit-agent-docs-<version>.zip
# Structure inside the zip:
#   reference.md                        ← agent command reference (auto-generated)
#   configuration.md
#   knowledgeprim/knowledgeprim-guide.md
#   queueprim/queueprim-guide.md
#
# Drop the zip wherever an agent needs it and unzip.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="$REPO_ROOT/dist"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out) OUT_DIR="$2"; shift 2 ;;
    *) echo "usage: sync-agent-docs.sh [--out DIR]" >&2; exit 1 ;;
  esac
done

VERSION="$(git -C "$REPO_ROOT" describe --tags --exact-match 2>/dev/null || git -C "$REPO_ROOT" describe --tags 2>/dev/null || echo "dev")"
ZIP_NAME="primkit-agent-docs-${VERSION}.zip"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

PKG="$TMP/primkit-agent-docs"
mkdir -p "$PKG/knowledgeprim" "$PKG/queueprim"

cp "$REPO_ROOT/docs/agent-reference.md" "$PKG/reference.md"
cp "$REPO_ROOT/docs/configuration.md"   "$PKG/configuration.md"
cp "$REPO_ROOT/docs/knowledgeprim.md"   "$PKG/knowledgeprim/knowledgeprim-guide.md"
cp "$REPO_ROOT/docs/queueprim.md"       "$PKG/queueprim/queueprim-guide.md"

mkdir -p "$OUT_DIR"
(cd "$TMP" && zip -r "$OUT_DIR/$ZIP_NAME" primkit-agent-docs/ -x "*.DS_Store")

echo "$OUT_DIR/$ZIP_NAME"
