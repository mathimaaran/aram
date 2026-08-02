#!/usr/bin/env bash
# Source from repo root:  source tools/env.sh
# Prefers system `go` if present; otherwise uses .tools/go (portable toolchain).

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if command -v go >/dev/null 2>&1; then
  echo "using system go: $(command -v go) ($(go version))"
else
  LOCAL_GO="$ROOT/.tools/go/bin"
  if [[ ! -x "$LOCAL_GO/go" ]]; then
    echo "error: go not found on PATH and $LOCAL_GO/go missing" >&2
    echo "install Go, or restore the portable toolchain under .tools/go/" >&2
    return 1 2>/dev/null || exit 1
  fi
  export PATH="$LOCAL_GO:$PATH"
  echo "using portable go: $LOCAL_GO/go ($(go version))"
fi

export GOPATH="${GOPATH:-$HOME/go}"
