#!/usr/bin/env bash
# Run go with the portable toolchain if system go is missing.
# Usage (from repo root):  ./tools/run-go.sh test ./...
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if command -v go >/dev/null 2>&1; then
  exec go "$@"
fi
GO_BIN="$ROOT/.tools/go/bin/go"
if [[ ! -x "$GO_BIN" ]]; then
  echo "go not found. Install golang, or use the portable toolchain:" >&2
  echo "  $ROOT/.tools/go/bin/go" >&2
  exit 127
fi
exec "$GO_BIN" "$@"
