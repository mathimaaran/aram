# Aram — design decisions

Locked decisions for Phase 0. Change only with an explicit note and date.

## Identity

| Decision | Choice | Locked |
|----------|--------|--------|
| Language name | **Aram** (அறம்) — virtue / rightness | 2026-07-25 |
| Source extension | `.aram` (Latin). Tamil may appear in basenames. | 2026-07-25 |
| Hello program | `வணக்கம்.aram` (ASCII fallback: `vanakkam.aram`) | 2026-07-25 |
| Keyword style | **Semantic Tamil** (meaning, not English phonetics) | 2026-07-25 |
| Relation to Go | **Go-inspired, free to diverge** | 2026-07-25 |
| Host language (compiler) | **Go** | 2026-07-25 |

## Architecture (frontend-first)

| Decision | Choice | Locked |
|----------|--------|--------|
| Product focus | Tamil grammar / UX for end users | 2026-07-25 |
| Early backend | Emit **C**, compile with `gcc`/`clang` | 2026-07-25 |
| Later backend | Emit **NASM**, assemble/link for Linux | 2026-07-25 |
| First CPU target | x86-64 Linux (SysV); i386 later | 2026-07-25 |
| Parser approach | Hand-written recursive descent (learning + control) | 2026-07-25 |
| IR | Small custom IR when NASM backend begins; not required for C backend | 2026-07-25 |

## Language surface (defaults until revised)

| Decision | Choice |
|----------|--------|
| Digits in source | Arabic `0`–`9` primary; Tamil `௦`–`௯` OK in numeric literals (0.31) |
| Identifiers | Tamil and/or Latin letters; NFC normalization planned |
| Strings | UTF-8; Tamil string literals allowed |
| Operators / braces | Keep ASCII operators and `{}()[]` for Phase 0–1 |
| Semicolons | Go-like automatic semicolon insertion (lexer; Tamil-0.7) |
| Goroutines / channels | Tamil-0.48 on C (pthread) |
| GC | Tamil-0.52 STW conservative mark-sweep on C (native heap) |

## Non-goals (for now)

- Go source compatibility
- Full Go standard library parity
- Self-hosting
- Windows / macOS as primary targets
- Phonetic Tamil keywords

## Naming package

```text
Full name:  Aram (அறம்)
Binary:     aram
Extension:  .aram
Hello:      வணக்கம்.aram
```

## Tamil-0.1 (2026-08-01)

Adds Go-style loops with keyword **சுழல்**, **முறி** (`break`), and **தொடர்** (`continue`).
See `grammar/tamil/constructs/for.yaml`, `break.yaml`, `continue.yaml`, and `corpus/tamil/சுழல்.aram`.

## Tamil-0.2 (2026-08-01)

Adds first-class strings: type keyword **சரம்**, variables/params/returns, `+` concatenation,
and `==` / `!=` content comparison. C backend uses `const char *`, `strcmp`, and a small
`aram_concat` helper (allocated; no free in Tamil-0.2).
See `grammar/tamil/constructs/string.yaml` and `corpus/tamil/சரம்.aram`.

## Tamil-0.3 (2026-08-01)

Adds Go-like slices: type `[]T` (T = முழுஎண் | நிலை | சரம்), composite literals `[]T{…}`,
index get/set, and builtin **நீளம்** (`len`). C backend uses fat pointers; out-of-bounds aborts.
No `ஒவ்வொரு`, append, or `xs[i:j]` yet.
See `grammar/tamil/constructs/slice.yaml` and `corpus/tamil/பட்டியல்.aram`.

## Tamil-0.4 (2026-08-01)

Adds **ஒவ்வொரு** (range) over slices inside `சுழல்`, Go-style:
`சுழல் i, v := ஒவ்வொரு xs { … }`, index-only, and blank `_`.
See `grammar/tamil/constructs/range.yaml` and `corpus/tamil/ஒவ்வொரு.aram`.

## Tamil-0.5 (2026-08-01)

- **சேர்** — append one element to a slice (`append`)
- **SliceExpr** — `xs[i:j]` / `xs[i:]` / `xs[:j]` / `xs[:]` on slices and `சரம்`
- **ஒவ்வொரு** over `சரம்` — byte index + rune (`முழுஎண்`), like Go
- **நீளம்** on `சரம்` — byte length
- **String/slice arena** — concat and growth allocate from an arena freed at end of `main`

See `grammar/tamil/constructs/append.yaml`, `slice-expr.yaml`, and `corpus/tamil/சேர்.aram`.

## Tamil-0.6 (2026-08-01)

Named structs: **வகை** (`type`) and **அமைப்பு** (`struct`), keyed literals, field `.` select/assign.
Field types in 0.6: முழுஎண், நிலை, சரம் only. No pointers or methods yet.
See `grammar/tamil/constructs/struct.yaml` and `corpus/tamil/அமைப்பு.aram`.

## Tamil-0.7 (2026-08-01)

Go-like pointers: `*T`, `&x`, `*p`; auto-deref field select on `*அமைப்பு`.
`T` may be முழுஎண், நிலை, சரம், named struct, or another pointer.
Struct field types stay scalars (no `*T` fields yet). No methods, no `nil` keyword.
See `grammar/tamil/constructs/pointer.yaml` and `corpus/tamil/சுட்டி.aram`.

## Tamil-0.8 (2026-08-01)

Methods: `செயல்பாடு (பெயர் T|*T) பெயர்(…)`, calls `X.பெயர்(…)`.
Named struct receivers only; Go-like addressability for pointer receivers.
See `grammar/tamil/constructs/method.yaml` and `corpus/tamil/முறை.aram`.

## Tamil-0.9 (2026-08-01)

Same-package multi-file: several `.aram` files with `தொகுப்பு தொடக்கம்`
share types/funcs/methods. CLI: `aram <dir>` or multiple paths.
Cross-package import keyword reserved: **கொணர்** (“bring”; verb — not noun இறக்குமதி).
See `grammar/tamil/constructs/multifile.yaml` and `corpus/tamil/பல்கோப்பு/`.

## Tamil-0.10 (2026-08-01)

**கொணர்** `"rel/path"` imports another package directory. Qualifier is the imported
`தொகுப்பு` name (`கணிதம்.கூட்டு`). All top-level names are visible (no privacy yet).
See `grammar/tamil/constructs/import.yaml` and `corpus/tamil/கொணர்_எடு/`.

## Tamil-0.11 (2026-08-01)

Opt-in export with **`வெளி`** (outside). Default package-private; same package sees
all names. Importers only see `வெளி` funcs/types/fields/methods. Replaces Go’s
capitalization rule (Tamil has no case). See `grammar/tamil/constructs/export.yaml`.

## Tamil-0.12 (2026-08-01)

Nil keyword **`இன்மை`** (absence / non-existence) plus pointer `==` / `!=`.
Untyped until used in a `*T` context; emits as C `NULL`. See
`grammar/tamil/constructs/nil.yaml`.

## Tamil-0.13 (2026-08-01)

Richer struct fields: nested named structs, `*T` (incl. recursive), and
`[]T` element scalars. Value-recursive structs rejected. See
`grammar/tamil/constructs/struct-fields.yaml`.

## Tamil-0.14 (2026-08-01)

Switch: **`திசைவி`** (director / router) + **`மற்றபடி`** (default). Case
clauses reuse **`எனில்`** (no separate case keyword). No fallthrough.
See `grammar/tamil/constructs/switch.yaml`.

## Tamil-0.15 (2026-08-01)

Go-style init before **`எனில்`** / **`திசைவி`**: `எனில் x := …; cond { … }`.
Scoped across body, cases, and `இல்லையேல்`. See `if-init.yaml`.

## Tamil-0.16 (2026-08-01)

Struct **`==` / `!=`**: field-wise, Go comparable rule (no slice fields).
See `grammar/tamil/constructs/struct-eq.yaml`.

## Tamil-0.17 (2026-08-01)

**`பதிப்பி`** of named structs: debug format `Type{field: value, …}`.
See `grammar/tamil/constructs/print-struct.yaml`.

## Tamil-0.18 (2026-08-01)

**`கொணர்` aliases**: `கொணர் க "கணிதம்"` then `க.கூட்டு`. Blank `_`
supported; no dot-import. See `grammar/tamil/constructs/import-alias.yaml`.

## Tamil-0.19 (2026-08-01)

Multi-**`சேர்`** and three-index slices **`xs[i:j:k]`** (Go rules). See
`grammar/tamil/constructs/slice-grow.yaml`.

## Tamil-0.20 (2026-08-01)

Positional struct literals `T{v1, v2, …}` (Go rules: no mix with keyed;
exact field count). See `grammar/tamil/constructs/struct-lit-pos.yaml`.

## Tamil-0.21 (2026-08-01)

Nested slices `[][]T` (and deeper) for leaf element types. See
`grammar/tamil/constructs/nested-slice.yaml`.

## Tamil-0.22 (2026-08-01)

**`ஆக்கு`** (make), **`நகல்`** (copy), **`திறன்`** (cap) for slices. See
`grammar/tamil/constructs/make-copy.yaml`.

## Tamil-0.23 (2026-08-01)

Multiple return values `(T, U)` with unpack via `:=` / `=`. See
`grammar/tamil/constructs/multi-return.yaml`.

## Tamil-0.24 (2026-08-01)

Fixed arrays `[N]T` (value types; sliceable to `[]T`). See
`grammar/tamil/constructs/array.yaml`.

## Tamil-0.25 (2026-08-01)

Slices of structs and pointers: `[]T`, `[]*T`. See
`grammar/tamil/constructs/slice-struct.yaml`.

## Tamil-0.26 (2026-08-01)

Named results `(ஈ முழுஎண், மீ முழுஎண்)`. See
`grammar/tamil/constructs/named-results.yaml`.

## Tamil-0.27 (2026-08-01)

Naked `திருப்பு` returns named result locals. See
`grammar/tamil/constructs/naked-return.yaml`.

## Tamil-0.28 (2026-08-01)

Parallel assign / swap `அ, ஆ = ஆ, அ`. See
`grammar/tamil/constructs/parallel-assign.yaml`.

## Tamil-0.29 (2026-08-01)

`வகை T = U` aliases and `வகை T U` defined types. See
`grammar/tamil/constructs/type-alias.yaml`.

## Tamil-0.30 (2026-08-01)

Float type keyword **மிதவைஎண்** (C `double`). See
`grammar/tamil/constructs/float.yaml`.

## Tamil-0.31 (2026-08-01)

Tamil digits `௦`–`௯` allowed in numeric literals (idents still Arabic). See
`grammar/tamil/constructs/tamil-digits.yaml`.

## Tamil-0.32 (2026-08-01)

Unused non-blank imports are errors; import cycles print `A → B → A`. See
`grammar/tamil/constructs/import-unused.yaml`.

## Tamil-0.33 (2026-08-01)

Conversions `T(x)`: defined ↔ underlying, `முழுஎண்` ↔ `மிதவைஎண்`. See
`grammar/tamil/constructs/conversion.yaml`.

## Tamil-0.34 (2026-08-01)

Bool/string/pointer conversions; mixed typed int/float arithmetic promotes to
`மிதவைஎண்`. See `grammar/tamil/constructs/conversion-extra.yaml`.

## Tamil-0.35 (2026-08-01)

Maps: type keyword **அகராதி**, builtin **நீக்கு**. Keys: `முழுஎண்`, `நிலை`,
`சரம்` (and defined types with those underlyings). Nil maps compare only to
`இன்மை`; assignment into a nil map aborts. C backend uses arena open-addressing
tables (handle is a pointer). See `grammar/tamil/constructs/map.yaml`.

## Tamil-0.36 (2026-08-01)

Map literals `அகராதி[K]V{k: v}` (keyed elements only; `{}` is non-nil empty).
`KeyValueExpr.Key` widened to any expression (struct field keys remain idents).
See `grammar/tamil/constructs/map-literal.yaml`.

## Tamil-0.37 (2026-08-01)

Defer keyword **தள்ளிவை** (“put aside”). CallExpr only; args evaluated now,
call runs LIFO at function return. C: arena-backed thunk stack + epilogue.
See `grammar/tamil/constructs/defer.yaml`.

## Tamil-0.38 (2026-08-01)

Map keys: any comparable type (floats, pointers, comparable structs/arrays).
See `grammar/tamil/constructs/map-keys.yaml`.

## Tamil-0.39 (2026-08-01)

`இருமி8` (byte/uint8) and `இருமி32` (rune/int32), plus `சரம்` ↔
`[]இருமி8` / `[]இருமி32`. Naming: இருமி = binary unit + bit width (not பைட்
or எழுத்து). See `grammar/tamil/constructs/byte-rune.yaml`.

## Tamil-0.40 (2026-08-01)

Function type parameters (unconstrained): `செயல்பாடு அடையாளம்[யா](x யா) யா`.
Conventional param name **யா**; no `any` keyword. Inference at calls;
optional explicit `அடையாளம்[முழுஎண்](x)`. C: monomorphize. Interfaces /
constraints / generic types deferred. See `constructs/generics.yaml`.

## Tamil-0.41 (2026-08-01)

Generics polish: multiple type params, `[]யா`, `f[T,U](…)`, exported
generics across `கொணர்`. See `constructs/generics.yaml`.

## Tamil-0.42 (2026-08-01)

Bare zero-arg defer sugar: `தள்ளிவை முடி` ≡ `தள்ளிவை முடி()`.
Conversions still rejected. See `constructs/defer.yaml`.

## Tamil-0.43 (2026-08-01)

Defer method bind: value receivers copied at `தள்ளிவை`; pointer
receivers on addressable values save `&x` (mutations before return are
visible). Applies to bare and `p.m()` forms. See `defer.yaml`.

## Tamil-0.44 (2026-08-01)

Function types (`செயல்பாடு(…) …` in type position) and first-class
method values / package function values. Receiver bind for method
values matches 0.43 defer. No function literals or interfaces in 0.44;
method expressions arrived in 0.47. See `constructs/func-value.yaml`.

## Tamil-0.45 (2026-08-01)

Function literals / closures: `செயல்பாடு(params) [results] { body }` as
expressions. Free variables captured by reference (Go-like); C arena
promotes captured locals. Builds on 0.44 `aram_fn`. See
`constructs/closure.yaml`.

## Tamil-0.46 (2026-08-01)

Go 1.22+ per-iteration loop variables: `சுழல்` / `ஒவ்வொரு` vars declared
with `:=` are distinct each iteration, so closures capture the correct
iteration values. See `constructs/closure.yaml`.

## Tamil-0.47 (2026-08-08)

Method expressions `T.M` and `(*T).M` as unbound function values on the
`aram_fn` path. Method-set rules match Go (value receivers on `T`; value
and pointer on `*T`). See `constructs/method-expr.yaml`.

## Tamil-0.48 (2026-08-08)

Concurrency on the **C** backend (pthread), not deferred to NASM:
`இழை` (go), `தடம்` (chan), `தடத்தேர்வு`/`எனில்`/`மற்றபடி` (select),
`மூடு` (close), ASCII `<-`. Select cases reuse `எனில்` (like `திசைவி`
and Go `case`); provisional `சூழல்` was dropped as too close to `சுழல்`.
Buffered and unbuffered channels; directional types. See
`constructs/goroutine.yaml`.

Parser hardening (same day): after a syntax error inside `{`…`}` loops,
always make progress (`advancePastError`) or stay on a sync token
(`}` / EOF / `;`). Select/switch/struct invalid corpus + 2s hang tests
guard against unbounded parse loops.

## Tamil-0.49 (2026-08-13)

Panic / recover: **அலறு** / **மீள்**, C `setjmp`/`longjmp` + `தள்ளிவை`
unwind (`__thread` state per `இழை`). Reserved builtins like `மூடு`.
MVP: `அலறு` takes `சரம்`; `மீள்()` returns `சரம்` (`""` if none).
Rejected **பீதி** (≈ பதிப்பி), **வெருளி** (≈ வெளி), **எச்சரி** (warn ≠ abort).
See `constructs/panic.yaml`.

## Tamil-0.50 (2026-08-13)

Panic / concurrency polish (no new keywords):
- `அலறு` accepts primitives (`முழுஎண்`, `மிதவைஎண்`, `நிலை`, `இருமி8`,
  `இருமி32`) and stringifies them; `மீள்` still returns `சரம்`.
- Uncaught `அலறு` dumps a function-name stack, then aborts.
- Panic / recover is per-`இழை` (recover in one thread does not see
  another).
- Unbuffered `தடத்தேர்வு` send only succeeds if a receiver is waiting
  (Go select); default is taken otherwise.
See `constructs/panic.yaml`, `goroutine.yaml`.

## Tamil-0.51 (2026-08-13)

`ஒவ்வொரு` over `தடம்`: at most one variable (the element); loop receives
until `மூடு` (and drain). Send-only channels are an error. Bare
`சுழல் ஒவ்வொரு x` is allowed (also for slices/maps). See `range.yaml`.

## Tamil-0.52 (2026-08-13)

Native heap + GC on the C backend (no new keywords):
- Replace bump arena with malloc-backed stop-the-world conservative
  mark-sweep (`aram_alloc`).
- Safe points: poll at function entry / loop back-edges; park around
  `தடம்` / `தடத்தேர்வு` cond-wait.
- Thread registry + Linux `pthread_getattr_np` stack bounds.
- `இழை` register/unregister; `aram_arena_alloc` is a wrapper.
See `constructs/gc.yaml`. HTTP / sockets still later.

## Priority (2026-08-01)

Harden the language (other gaps) **before** starting the NASM x86-64
backend. See `subset-roadmap.md`.

## Tamil-0 freeze

| Item | Value |
|------|--------|
| Status | **Frozen** |
| Date | 2026-08-01 |
| Artifacts | `grammar/tamil/keywords.yaml`, `grammar/tamil/ebnf.md`, `grammar/tamil/constructs/*`, `corpus/tamil/*` |
| Change policy | No casual keyword/EBNF edits; bump to Tamil-0.1+ for breaking changes |

Tamil-0 is the implementation target for Phase 1 (lexer → parser → C backend).

Phase 1 progress: lexer, parser, typecheck, C emit (`emit`/`build`/`run`). NASM still later.

## Keyword locks (Tamil-0)

| Role | Tamil | Locked |
|------|--------|--------|
| package | தொகுப்பு | 2026-08-01 freeze |
| var | மாறி | 2026-07-25 |
| int type | முழுஎண் | 2026-08-01 |
| bool type | நிலை | 2026-08-01 freeze |
| true / false | மெய் / பொய் | 2026-08-01 freeze |
| short declare | `:=` (Go-style, with மாறி) | 2026-07-25 |
| if / else | எனில் / இல்லையேல் | 2026-07-25 |
| return | திருப்பு | 2026-07-25 |
| print | பதிப்பி | 2026-07-25 |
| func | செயல்பாடு | 2026-07-25 |
| entry name | தொடக்கம் (identifier, not keyword) | 2026-07-25 |
| for | சுழல் | Tamil-0.1 |
| break | முறி | Tamil-0.1 |
| continue | தொடர் | Tamil-0.1 |
| string type | சரம் | Tamil-0.2 |
| slice type | `[]T` (ASCII brackets) | Tamil-0.3 |
| len | நீளம் | Tamil-0.3 |
| range | ஒவ்வொரு | Tamil-0.4 (slices); 0.5 also சரம் |
| append | சேர் | Tamil-0.5 |
| type | வகை | Tamil-0.6 |
| struct | அமைப்பு | Tamil-0.6 |
| pointer ops | `*` / `&` (ASCII) | Tamil-0.7 |
| methods | receiver after `செயல்பாடு` | Tamil-0.8 |
| multi-file package | same `தொகுப்பு` name, no import | Tamil-0.9 |
| import | கொணர் | Tamil-0.10 |
| default (later) | மற்றபடி | reserved |
| map | அகராதி | Tamil-0.35 |
| delete | நீக்கு | Tamil-0.35 |
| defer | தள்ளிவை | Tamil-0.37 |
| panic | அலறு | Tamil-0.49 |
| recover | மீள் | Tamil-0.49 |
