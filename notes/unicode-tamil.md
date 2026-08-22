# Unicode / Tamil source notes

## Encoding

- Source files are **UTF-8**.
- Plan: normalize identifiers to **NFC** before comparison (TBD at lexer implementation).
- Digits in Tamil-0: ASCII `0`–`9` only.

## Letters

- Tamil letters (Unicode Tamil block, roughly U+0B80–U+0BFF) are valid in identifiers.
- Latin letters remain valid (interop, gradual adoption, tooling).
- `_` allowed in identifiers (Go-like).

## Keywords

- Keywords are exact Unicode strings (semantic Tamil), reserved, not usable as identifiers.
- See `grammar/tamil/keywords.yaml`.

## Filenames

- Official extension: **`.uli`** (ASCII).
- Basenames may be Tamil: `வணக்கம்.uli`.
- Do not use `.அ` or `.நிரலுளி` as the official extension.

## Fonts / glyphs (editors)

If Tamil shows as empty **squares**, the file is usually fine — the editor font lacks Tamil glyphs.
Use **Noto Sans Tamil UI** / **Noto Sans Tamil** as a fallback. Full guide: [`fonts-tamil.md`](fonts-tamil.md).

## Open questions

1. Case folding: Tamil has no case; Latin identifiers — case-sensitive (Go-like)?

## Resolved

- Combining marks in identifiers: **yes** (needed for Tamil orthography; lexer allows Unicode marks). Frozen with Phase 1 lexer 2026-08-01.
- Tamil digits in numeric literals: **yes** (Tamil-0.31); identifiers still Arabic digits only.