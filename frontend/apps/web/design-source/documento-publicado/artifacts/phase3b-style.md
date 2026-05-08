# Phase 3b — Style Port Artifact

**Screen:** documento-publicado  
**Phase:** 3b — Style port  
**Date:** 2026-05-08  
**Status:** COMPLETE — pending user screenshot review

---

## What was done

### Files changed

| File | Change |
|---|---|
| `src/styles/tokens.css` | Added 4 new tokens (shadow-doc-card, shadow-success, shadow-danger, overlay-white-65) |
| `src/features/documents/pages/DocumentPublishedPage.tsx` | Changed button classNames from CSS Module classes to global `.btn` system |
| `src/features/documents/pages/DocumentPublishedPage.module.css` | Filled all CSS rules from Phase 3a skeleton with design values |

### New tokens added to tokens.css

- `--shadow-doc-card` — 3D perspective card drop shadow (matched from design rgba values)
- `--shadow-success` — green glow for signoff pins
- `--shadow-danger` — red glow for obsolete stamp
- `--overlay-white-65` — frosted white overlay for obsolete stamp background

### Button class change (TSX)

Removed CSS Module `.btnPrimary`, `.btnSecondary`, `.btnGhost` classes (which were empty in Phase 3a). Replaced with global btn system classes:
- "Visualizar documento" → `className="btn btn-primary btn-lg"`
- "Iniciar revisão" → `className="btn"`
- "Copiar link" → `className="btn btn-ghost"`

### CSS Module — regions ported

1. Root (`.root`) — flex: 1, overflow: auto, relative positioning for overlay
2. Hero (`.hero`) — gradient bg, padding, border, grid overlay
3. Breadcrumb (`.breadcrumb`) — mono font, uppercase, muted color, gap
4. DocCardMini (`.docCard`) — 3D perspective transform, shadow, gradient header
5. Hero content — badges row, vigente badge, code chip, type label, title
6. Content area (`.content`) — padding 32/56/80, max-width 1180px, flex column gap
7. KPI strip — single column grid, surface/border card
8. Section + section head — kicker/title typography
9. AboutCard — owner banner (surface-2 bg), facts 2-col grid, fact cells
10. SignoffPipeline — 3-col grid, connector line (success color), pins with success shadow
11. ObsoleteBanner — absolute overlay, hidden by default, rotated stamp

---

## Token mapping decisions

| Design value | Token used | Rationale |
|---|---|---|
| `#fff` (card header text) | `var(--text-on-brand)` | Existing token `#ffffff` for text on brand backgrounds |
| `color: var(--surface)` on pin | `var(--text-on-brand)` | Same — white text on success green bg |
| `20px 20px 48px rgba(74,33,33,0.16), 4px 4px 14px rgba(74,33,33,0.10)` | `var(--shadow-doc-card)` | New shadow token (brand-tinted, heavy depth) |
| `0 2px 8px rgba(26,107,53,0.30)` | `var(--shadow-success)` | New shadow token (success-tinted glow) |
| `0 8px 32px rgba(200,54,74,0.20)` | `var(--shadow-danger)` | New shadow token (danger-tinted glow) |
| `rgba(255,255,255,0.65)` (stamp bg) | `var(--overlay-white-65)` | New overlay token — cannot use `var(--surface)` (would be opaque) |

---

## Spacing approach

Spacing tokens available: `--sp-1` (4px) through `--sp-9` (56px). All padding/margin/gap values mapped to `rem` equivalents since the spacing tokens are in `px` and the design values given are also in `px`. Used `rem` throughout the CSS Module for scalability.

Micro-values with no token (2px connector height, 5px dot size) used raw `px` — exempt per hard rule (sub-token micro-values).

---

## TypeScript check

Pre-existing error count: 12  
Post-change error count: 12  
New errors introduced: 0  
Status: PASS

---

## Token coverage

Raw hex/rgb check: PASS (no raw values — all via `var(--token)`)

---

## Screenshots taken

| Viewport | Impl | Design ref |
|---|---|---|
| 1440px | Taken (see visual inspection) | Design server timed out |
| 1024px | Taken | Design server timed out |
| 375px | Taken (hero grid overflows — no responsive breakpoints in Phase 3b scope) | Design server timed out |

Note: Design source server (`port 4181`) consistently timed out on screenshot tool calls. Parity was verified via computed-style inspection (`preview_inspect` + `preview_eval`) against spec values. All 53 inspected fields pass. See `parity-diff.md`.

---

## Outstanding items for Phase 3c

- Wire `obsoleteBanner` display condition: `display: none` → `display: flex` when `status === 'obsolete'`
- `useParams()` integration for real document data
- Mobile responsive breakpoint for heroGrid (not in Phase 3b scope — no mobile breakpoints in design values provided)
- User visual sign-off on screenshots (hard rule: cannot be marked by subagent)
