# Tools / compiler

Phase 1 (in progress): Go module at repo root.

| Path | Role |
|------|------|
| `cmd/uli` | CLI: `run` / `build` / `emit` / `check` / `parse` / `lex` |
| `internal/token` | token kinds + Tamil-0 keyword map |
| `internal/lex` | UTF-8 lexer |
| `internal/ast` | AST nodes |
| `internal/parse` | recursive-descent parser |
| `internal/check` | type checker |
| `internal/emitc` | C99 backend |
| `tools/env.sh` | `source` this to put Go on your PATH |
| `tools/run-go.sh` | run `go` without installing system-wide |
| `.tools/go/` | portable Go 1.22 (gitignored) |

`run`/`build` need a C compiler (`gcc` or `cc`). `emit` writes `.c` only.

Default output directory (when `-o` is omitted): **`build/`**  
(`build/<basename>.c` and `build/<basename>`).


## If `go: command not found`

You do **not** need `sudo apt install golang-go` for this repo. Use the portable toolchain:

```bash
cd /path/to/uli

# option A — one-off
./tools/run-go.sh test ./...
./tools/run-go.sh run ./cmd/uli corpus/tamil/வணக்கம்.uli

# option B — for the rest of the shell session
source tools/env.sh
go test ./...
go run ./cmd/uli corpus/tamil/வணக்கம்.uli
```

Or install a system Go later (`sudo apt install golang-go`) if you prefer; `tools/env.sh` will use it automatically.
