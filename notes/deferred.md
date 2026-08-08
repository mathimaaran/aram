# Deferred capabilities

Natural enhancements kept out of the current subset so each Tamil-0.x stay small and closed-loop.
**Ordered next work** stays in [`subset-roadmap.md`](subset-roadmap.md); this file is the inventory of *what* we still want, with sources.

Update when a subset ships (strike or move items) or when a construct card records a new cut.

## Likely next (harden language first; NASM later)

| Item | Notes | Source |
|------|--------|--------|
| NASM x86-64 → i386 | After language harden | roadmap |
| Interfaces | Only if earned | method.yaml |
| Generic types / constraints | After interfaces earn a place | generics.yaml |

Shipped: 0.11–0.44 (through method / function values).

## Struct & types

| Item | Notes | Source |
|------|--------|--------|
| ~~string ↔ []byte / []rune~~ | Tamil-0.39 (`இருமி8` / `இருமி32`) | byte-rune.yaml |

## Control & functions

| Item | Notes | Source |
|------|--------|--------|
| ~~bare zero-arg defer~~ | Tamil-0.42 | defer.yaml |
| ~~defer method bind (`&x`)~~ | Tamil-0.43 | defer.yaml |
| ~~first-class method / function values~~ | Tamil-0.44 | func-value.yaml |
| defer closures / function literals | still later | defer.yaml |

## Packages, runtime, backends

| Item | Notes | Source |
|------|--------|--------|
| GC / native heap | Arena only while on C | design-decisions |
| Goroutines / channels | | design-decisions |
| panic / recover | | Tamil-0 out of scope |
| NASM + custom IR | IR when NASM begins | design-decisions |
| Windows / macOS primary targets | Linux-first | Non-goals |

## Larger Go-inspired (only if earned)

Interfaces, richer generics (types/constraints), concurrency, full stdlib parity, Go source compatibility, self-hosting, phonetic Tamil keywords.

See **Non-goals** in [`design-decisions.md`](design-decisions.md).

## How to promote an item

1. Propose the subset in `design-decisions.md` (dated).
2. Lock keywords / EBNF / construct card.
3. Add corpus, then lex → parse → check → emit → run.
4. Check the item off here and on `subset-roadmap.md`.
