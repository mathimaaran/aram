# Tamil-0 freeze notice

**Frozen:** 2026-08-01  
**Subset:** Tamil-0  
**Next:** Phase 1 implementation (`cmd/aram` lexer → parser → C emit)

## Immutable for Tamil-0 (unless bumping subset)

- `grammar/tamil/keywords.yaml`
- `grammar/tamil/ebnf.md`
- Construct cards under `grammar/tamil/constructs/`
- Valid/invalid programs under `corpus/tamil/` (behavioral surface)

## Allowed without bump

- Typo fixes in comments/docs
- Adding *more* corpus files that still obey frozen EBNF (prefer documenting in README)
- Implementation code under `tools/` / future Go module

## Breaking change process

1. Propose Tamil-0.1 (or later) in `notes/design-decisions.md`
2. Update keywords + EBNF + constructs + corpus together
3. Note migration impact for any existing examples
