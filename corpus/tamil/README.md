# Tamil corpus

Example and (later) test programs for Aram. Subset: **Tamil-0 (frozen 2026-08-01)**.

Corpus programs are part of the frozen snapshot; treat changes as subset bumps unless fixing comments/docs only.

## Valid

| File | Covers |
|------|--------|
| `வணக்கம்.aram` | package, entry `தொடக்கம்`, `பதிப்பி` string |
| `எண்கணிதம்.aram` | `மாறி`, `:=`, `=`, arithmetic `+ - * / %` |
| `நிபந்தனை.aram` | `எனில்` / `இல்லையேல்`, `மெய்` / comparisons |
| `கூட்டு.aram` | helper `செயல்பாடு`, params, `திருப்பு`, call |
| `சுழல்.aram` | Tamil-0.1 | `சுழல்` C-style + while-like loops |
| `பல்கோப்பு/` | Tamil-0.9 | same-package multi-file (`aram பல்கோப்பு`) |
| `கொணர்_எடு/` | Tamil-0.10 | `கொணர் "கணிதம்"` + `கணிதம்.கூட்டு` |
| `வலை_எதிரொலி/` | Tamil-0.53 | TCP echo via `கொணர் "வலை"` |
| `வலைபரிமாற்றம்/` | Tamil-0.56 | HTTP request headers + exact-path mux |
| `வலைபரிமாற்றம்_html/` | Tamil-0.54 | HTTP HTML response |
| `வலை_பிழை/` | Tamil-0.55 | TCP error result (`*வலை.பிழை`) |

## Invalid (`invalid/`)

| File | Why invalid |
|------|-------------|
| `தொகுப்பு_இல்லை.aram` | missing `தொகுப்பு` clause |
| `அறிவிக்கப்படாத_ஒதுக்கீடு.aram` | `=` without prior `மாறி` / `:=` (syntax OK; semantic error later) |
| `பழைய_else.aram` | uses `இல்லையெனில்` (IDENT); not parse-fatal — style/semantic later |
| `முழுமையற்ற_ஒதுக்கீடு.aram` | `=` with missing expression (parse error) |

## Conventions

- One idea per file where practical.
- Basenames may be Tamil; extension is always `.aram`.
- ASCII fallback names are for docs only (not checked in).
- Invalid programs are expected to fail once the checker/parser exists.
- Comments use Go style (`//` line, `/* */` block). Each file includes a Go-equivalent sketch in a leading `//` comment.
- Tamil must render in the editor: see [`notes/fonts-tamil.md`](../../notes/fonts-tamil.md) (Noto Sans Tamil).

When adding a program, update this table and prefer matching a construct card under `grammar/tamil/constructs/`.
