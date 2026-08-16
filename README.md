# Aram (அறம்)

Aram is a **Go-inspired** programming language with **semantic Tamil** keywords.
It is free to diverge from Go. The long-term goal is a quality compiler for
Linux (x86-64 first, then i386), with NASM as a later backend.

**Phase 0:** complete (Tamil-0 frozen 2026-08-01).  
**Phase 1:** active C backend through v0.62; NASM remains later.

| Item | Value |
|------|--------|
| Language | Aram (அறம்) |
| Host compiler language | Go |
| Source extension | `.aram` |
| Keywords | Semantic Tamil |
| Semantics | Go-inspired, free to diverge |
| Language subset | **v0.62** |
| Early backend | Emit C (reuse `gcc`/`clang`) |
| Later backend | NASM → Linux ELF (64-bit, then 32-bit) |
| Status | Experimental research compiler (Linux + C backend) |
| License | [Apache License 2.0](LICENSE) |

## First program

```text
corpus/tamil/வணக்கம்.aram
```

## Layout

```text
notes/           design decisions, roadmap, Unicode, Tamil fonts
grammar/go/      frozen Go reference (inspiration, not a checklist)
grammar/tamil/   Aram grammar, keywords, construct cards
corpus/tamil/    example and test programs
stdlib/          compiler stdlib (வலை, பரிமாற்றம், தரவுத்தளம்)
tools/           experiments (empty in Phase 0)
```

## Editor setup (Tamil glyphs)

Aram source uses Tamil script. If you see **squares** instead of letters, install/use
**Noto Sans Tamil** and set a font fallback — see [`notes/fonts-tamil.md`](notes/fonts-tamil.md).

## Status

- Phase 0 complete — Tamil-0 **frozen** (2026-08-01). See `notes/design-decisions.md`.
- Phase 1: lexer + parser + typecheck + **C emit/run** through **v0.62** on Linux.
- Suitable for **public experimental** use if framed as early/research; not a polished production language yet.
- Next major backend: NASM x86-64. Optional leftovers live in [`notes/deferred.md`](notes/deferred.md).

```bash
source tools/env.sh   # if needed
go test ./...
go run ./cmd/aram check corpus/tamil/வணக்கம்.aram
go run ./cmd/aram run corpus/tamil/வணக்கம்.aram
# → build/வணக்கம்.c and build/வணக்கம் (override with -o)
# C builds default to -O2; pass --debug or -O0 for an unoptimized build.
```

See [`tools/README.md`](tools/README.md) for the portable Go toolchain.

## Developer documentation

Open [`docs/index.html`](docs/index.html) for the comprehensive static HTML
reference: getting started, language syntax, concurrency/runtime, standard
library APIs (TCP/UDP, HTTP, and SQL/`தரவுத்தளம்`), compiler tooling,
implementation notes, and misc extras (license and Tamil typing tutors
without phonetic input).

## License

Aram is licensed under the [Apache License, Version 2.0](LICENSE).
See [`NOTICE`](NOTICE) for copyright attribution.

## Working with AI agents

Prefer feeding: construct card(s) + keyword slice + a few corpus examples +
explicit subset (`Tamil-0`), not the entire Go grammar.
