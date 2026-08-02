# Tamil fonts and glyphs (editors & launch)

Aram source is **UTF-8 Tamil**. If an editor or terminal shows **empty boxes (“tofu” / squares)** instead of letters, the text is usually fine — the **font cannot render Tamil glyphs**.

This matters for:

- Developing Aram (reading `.aram` corpus and keywords)
- End users writing Aram
- Future IDE / syntax-highlight extensions we ship
- Docs, websites, and screenshots at language launch

## Recommended fonts

Prefer **Noto** Tamil faces (free, complete, good shaping for conjuncts):

| Font | Use |
|------|-----|
| **Noto Sans Tamil UI** | First choice for UI / editors |
| **Noto Sans Tamil** | Fallback |
| **Noto Serif Tamil** | Optional for docs / PDF |

On Debian/Ubuntu these often come from packages such as `fonts-noto-core` or `fonts-noto`.

Check installed Tamil fonts:

```bash
fc-list :lang=ta | head
```

## Why monospace alone is not enough

Typical coding fonts (Consolas, Menlo, many “Mono” faces) cover Latin well but **omit the Tamil block** (roughly U+0B80–U+0BFF) or lack proper **complex-script shaping**.

Tamil needs shaping for:

- Combining vowel signs (e.g. `ெ` U+0BC6, `ு` U+0BC1)
- Virama / pulli `்` (U+0BCD) and consonant clusters  
  Example identifiers: `பெருக்கு`, `வகுப்பு`

Without a Tamil-capable font in the fallback chain, those code points render as squares even though the file is valid UTF-8 NFC.

## Cursor / VS Code settings

Put a Tamil font **after** your preferred monospace so Latin stays monospace and Tamil falls back correctly:

```json
{
  "editor.fontFamily": "monospace, 'Noto Sans Tamil UI', 'Noto Sans Tamil'"
}
```

User settings path (Linux): `~/.config/Cursor/User/settings.json`  
(VS Code: `~/.config/Code/User/settings.json`)

After changing fonts: **Developer: Reload Window**, then reopen the `.aram` file.

### Workspace tip (optional later)

When we ship an Aram VS Code/Cursor extension, recommend or set:

```json
{
  "[aram]": {
    "editor.fontFamily": "monospace, 'Noto Sans Tamil UI', 'Noto Sans Tamil'"
  }
}
```

(Requires a language id `aram` for `.aram` files.)

## Terminal (program output like `வணக்கம், அறம்!`)

Terminals are **cell grids**. Tamil glyphs are often wider than one Latin cell, so a monospace primary font + thin Tamil fallback looks **squeezed** and **lighter** even when readable.

### Best practical choice (you already have these)

| Priority | Font | Why |
|----------|------|-----|
| 1 | **Noto Sans Tamil UI** Regular/Medium | Clearest Tamil shaping for UI/terminal fallback |
| 2 | **Noto Sans Tamil** Regular/Medium | Same family; avoid Condensed/Light/Thin |
| Avoid for Tamil | Noto *Condensed* / *Light* / *Thin* | Causes the “squeezed + faint” look |

There is no widely used true **Tamil monospace**. So:

- Keep a coding mono for ASCII (`JetBrains Mono`, `Noto Sans Mono`, …)
- Let Tamil fall back to **Noto Sans Tamil UI** (Regular/Medium, not Light)

### Cursor integrated terminal

In `~/.config/Cursor/User/settings.json`:

```json
{
  "terminal.integrated.fontFamily": "JetBrains Mono, 'Noto Sans Tamil UI', 'Noto Sans Tamil'",
  "terminal.integrated.fontSize": 15,
  "terminal.integrated.fontWeight": "normal",
  "terminal.integrated.lineHeight": 1.35
}
```

Tips:

- Bump **fontSize** (14→16) and **lineHeight** (~1.3–1.4) — biggest readability win
- If Tamil still looks **cramped**: raise `terminal.integrated.letterSpacing` (e.g. `1.2`) and `lineHeight` (`1.5`), and put **Noto Sans Tamil UI** first in `fontFamily` (Tamil uses its native widths; Latin becomes proportional)
- Prefer **Medium** weight (`fontWeight` `500`) over Light/Thin
- Reload the window after changing settings

### GNOME Terminal / VTE (external)

1. Preferences → Profile → Text
2. Custom font: e.g. `JetBrains Mono Regular` 13–15
3. Ensure system fallback can reach Noto Tamil (fontconfig), **or** temporarily set the profile font to **Noto Sans Tamil UI** when demoing Aram output (Latin will be proportional)

Optional fontconfig tweak (`~/.config/fontconfig/fonts.conf`) so monospace stacks Tamil better:

```xml
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <alias>
    <family>monospace</family>
    <prefer>
      <family>JetBrains Mono</family>
      <family>Noto Sans Mono</family>
      <family>Noto Sans Tamil UI</family>
      <family>Noto Sans Tamil</family>
    </prefer>
  </alias>
</fontconfig>
```

Then: `fc-cache -fv` and restart the terminal.

### Terminals with better complex-script shaping

If VTE still looks cramped: **Kitty**, **WezTerm**, or **GNOME Console** often shape Tamil more cleanly than older terminals. Alacritty is weaker for complex scripts.

### What will *not* fully fix it

- Only increasing mono size without a Tamil fallback
- Using Condensed Tamil faces
- Expecting perfect alignment of Tamil inside strict monospace columns

For Aram demos, slightly larger terminal font + Noto Sans Tamil UI is the recommended setup.

### Other editors / web

- **Vim/Neovim GUI:** configure `guifont` with a Tamil-capable face; in terminal Vim, the **terminal** font matters.
- **JetBrains:** Settings → Editor → Font → fallback; add Noto Sans Tamil UI.
- **Browsers / docs site:** CSS e.g. `system-ui, "Noto Sans Tamil UI", "Noto Sans Tamil", sans-serif`

## File encoding checklist (not a font issue)

| Check | Expectation |
|-------|-------------|
| Encoding | UTF-8 |
| BOM | Prefer **no** BOM |
| Normalization | Identifiers: plan **NFC** (see `unicode-tamil.md`) |
| Extension | `.aram` (ASCII) |

If `hexdump` / Python shows correct `U+0Bxx` code points but the UI shows squares → **font**.  
If code points are wrong or `` → **encoding/corruption**.

## Launch / packaging notes

When announcing or distributing Aram:

1. **README “Editor setup”** — link this doc; require or recommend Noto Tamil.
2. **Installer / devcontainer** — install `fonts-noto` (or equivalent) on Linux images used for demos/CI screenshots.
3. **Extension marketplace page** — call out Tamil font requirement; avoid screenshots that look broken on tofu systems.
4. **CI** — do not treat “glyph render” as a compiler test; optionally add a docs smoke check that sample strings are NFC UTF-8.
5. **Do not** change language keywords to ASCII to work around missing fonts; fix the environment.

## Related

- `notes/unicode-tamil.md` — encoding, identifiers, digits
- `corpus/tamil/` — sample programs with Tamil identifiers and keywords
