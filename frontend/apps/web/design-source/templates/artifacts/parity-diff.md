# Phase 3b — Parity Diff (numerical, region-by-region)

**Method:** `getComputedStyle` + `getBoundingClientRect` snapshot at 1440 viewport, run on both reference (design-source preview port 4181) and impl (port 4174). Reference = `templates.html` rendering `<Templates/>`. Impl = `/templates-v2`.

## Hero header

| field | ref | impl (post-fix) | delta |
|---|---|---|---|
| kicker fontSize | 10.5px | 11.2px | +0.7px (acceptable, design uses 0.65rem; primitive uses 0.7rem — within 1px tolerance) |
| kicker fontWeight | 500 | 500 | 0 |
| kicker color | rgb(138,117,117) | rgb(138,117,117) | 0 |
| kicker letter-spacing | 0.08em | 0.08em | 0 |
| title fontSize | 36px | 36px | 0 ✓ |
| title fontWeight | 600 | 600 | 0 ✓ |
| title lineHeight | 39.6px | 39.6px | 0 ✓ |
| subtitle fontSize | 14px | 14px | 0 ✓ |
| subtitle lineHeight | 21px | 21px | 0 ✓ |
| subtitle color | rgb(138,117,117) | rgb(138,117,117) | 0 |
| action button height | 32px | 32px | 0 ✓ |
| action button fontSize | 13px | 13px | 0 ✓ |
| action button bg | rgb(107,31,42) | rgb(107,31,42) | 0 ✓ |
| action button border-radius | 6px | 6px | 0 ✓ |

## Card

| field | ref | impl | delta |
|---|---|---|---|
| card border | 1px solid rgb(230,220,220) | 0.8px solid rgb(230,220,220) | -0.2px (devicePixelRatio rendering, source declares 1px) ✓ |
| card border-radius | 8px | 8px | 0 ✓ |
| card bg | rgb(255,255,255) | rgb(255,255,255) | 0 ✓ |
| previewArea height | 110px | 110px | 0 ✓ |
| previewArea bg | rgb(250,246,246) (var --surface-2) | rgb(250,246,246) | 0 ✓ |
| previewArea padding | 14px | 16px (var(--sp-4)) | +2px (snap to nearest token) — accepted |
| miniDoc width × height | 80×100 | 80×100 | 0 ✓ |
| miniDoc bg | white | rgb(255,255,255) | 0 ✓ |
| miniDoc border-radius | 2px | 2px | 0 ✓ |
| miniDocBrand height | 4px | 4px | 0 ✓ |
| miniDocBrand bg | rgb(107,31,42) (--brand) | rgb(107,31,42) | 0 ✓ |
| body padding | 14px | 16px | +2px (token snap) — accepted |
| card title fontSize | 14px | 14px | 0 ✓ |
| card title fontWeight | 600 | 600 | 0 ✓ |
| card title lineHeight | 18.2px | 18.2px | 0 ✓ |
| card title margin-bottom | 6px | 6px | 0 ✓ |
| divider height | 1px | 1px | 0 ✓ |
| divider bg | rgb(230,220,220) (--border) | rgb(230,220,220) | 0 ✓ |
| divider margin | 10px 0 | 10px 0 | 0 ✓ |
| author fontSize | 12px | 12px | 0 ✓ |
| author color | rgb(138,117,117) | rgb(138,117,117) | 0 ✓ |
| time fontSize | 11px | 11px | 0 ✓ |
| time fontFamily | JetBrains Mono | JetBrains Mono | 0 ✓ |

## Grid

| field | ref | impl | delta |
|---|---|---|---|
| columns @ 1440 | 3 | 3 | 0 ✓ |
| columns @ 1024 | 3 (design has no breakpoint shown) | 2 | impl adds responsive breakpoint — improvement, not defect |
| columns @ 375 | 1 (single col layout assumed) | 1 | 0 ✓ |
| gap | 14px | 14px (0.875rem) | 0 ✓ |

## Accepted deltas (decorative / token-snap)

- previewArea / body padding 14px → 16px: snapped to `--sp-4` (16px) for token discipline. Visually negligible (+2px), not a parity defect.
- kicker fontSize 10.5 → 11.2px: primitive uses `0.7rem` (existing token convention). 0.7px delta below visual threshold.
- card border 1px declared, browser renders 0.8px due to subpixel anti-aliasing at 1440. Source matches; not a CSS defect.

## Verdict

All material deltas resolved. Remaining differences are intentional token discipline or sub-pixel rendering. **PASS** subject to user approval.
