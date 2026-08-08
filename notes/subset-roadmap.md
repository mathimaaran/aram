# Subset roadmap

Grow the language only when the previous subset has a closed loop:
**grammar card → corpus → (later) parse → typecheck → backend → run**.

## Tamil-0 (first closed loop) — **FROZEN 2026-08-01**

Goal: one runnable greeting and simple integer programs via **C backend**.

Spec artifacts are frozen: `grammar/tamil/keywords.yaml`, `ebnf.md`, construct cards, corpus.
Implementation (Phase 1) may begin against this snapshot.

In scope:

- `தொகுப்பு` (package) + entry name `தொடக்கம்` (start/main, conventional identifier)
- `செயல்பாடு` (func) with no params / `முழுஎண்` (int) params later in Tamil-0.1
- integer literals, `+ - * /`, comparisons
- `மாறி` (var) and Go-style `:=` short declaration
- `எனில்` / `இல்லையேல்` (if / else)
- `திருப்பு` (return)
- `பதிப்பி` (print) as a builtin or tiny runtime helper
- string literals (for `பதிப்பி` only)

Out of scope for Tamil-0:

- floats, arrays, structs, pointers
- packages/imports beyond a single file
- methods, interfaces, generics
- concurrency, defer, panic
- NASM backend

## After Tamil-0 — Tamil-0.1 … 0.6 done

- [x] `சுழல்` / loops (Go-style single `for`)
- [x] `முறி` / break
- [x] `தொடர்` / continue
- [x] Strings as values (`சரம்`) — concat `+`, compare `==`/`!=`
- [x] Go-like slices (`[]T`, index, `நீளம்`)
- [x] `ஒவ்வொரு` (range) over slices
- [x] `சேர்` (append), `xs[i:j]`, range over `சரம்`, string arena
- [x] Named structs (`வகை` / `அமைப்பு`) — no pointers yet
- [x] Pointers (`*T`, `&x`, `*p`; auto-deref on `*struct`)
- [x] Methods on struct / pointer receivers
- [x] Same-package multi-file (`aram <dir>` / multiple `.aram`)
- [x] Cross-package `கொணர்` (import)
- [x] Export / privacy (`வெளி`)
- [x] `இன்மை` (nil) + pointer `==` / `!=`
- [x] Richer struct fields (`*T`, nested, `[]T`)
- [x] `திசைவி` / `மற்றபடி` (switch / default)
- [x] Init before `எனில்` / `திசைவி`
- [x] Struct `==` / `!=`
- [x] `பதிப்பி` whole structs
- [x] Import aliases (`கொணர்`)
- [x] Multi-`சேர்` / `xs[i:j:k]`
- [x] Positional struct literals
- [x] Nested slices (`[][]T`)
- [x] `ஆக்கு` / `நகல்` / `திறன்` (make / copy / cap)
- [x] Multiple return values
- [x] Fixed arrays (`[N]T`)
- [x] `[]struct` / `[]*T`
- [x] Named results
- [x] Naked `திருப்பு`
- [x] Parallel assign / swap
- [x] Type alias + defined `வகை`
- [x] Floats (`மிதவைஎண்`)
- [x] Tamil digits in literals
- [x] Unused imports + cycle path messages
- [x] Conversions `T(x)`
- [x] Conversion extras + mixed int/float
- [x] Maps (`அகராதி` / `நீக்கு`)
- [x] Map literals
- [x] `தள்ளிவை` (defer)
- [x] Richer map keys (comparable)
- [x] `இருமி8` / `இருமி32` + `சரம்` ↔ `[]` conversions
- [x] Function generics MVP (`[யா]`, monomorphize)
- [x] Generics polish (multi-param, `[]யா`, import)
- [x] Bare zero-arg `தள்ளிவை` sugar
- [x] Defer method bind (pointer `&x` capture)
- [x] Function types + method / package function values
- [x] Function literals / closures (capture by reference)
- [x] Go 1.22+ per-iteration `சுழல்` / `ஒவ்வொரு` vars

## Later growth

**Harden the language first** (before native backends):

1. Other small gaps in [`deferred.md`](deferred.md)

**Native backends (after language harden):**

5. NASM x86-64
6. i386 Linux
7. Larger Go-inspired features only if they earn their place
   (interfaces, richer generics/constraints, …)

Inventory: [`deferred.md`](deferred.md).

## Status legend (construct cards)

`planned` → `specified` → `lexed` → `parsed` → `typed` → `codegen`
