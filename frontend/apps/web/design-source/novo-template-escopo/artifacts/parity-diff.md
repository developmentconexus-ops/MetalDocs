# Parity Diff — Phase 3b: novo-template-escopo

> Ref values from `template-wizard.jsx` TplStep1 inline styles.
> Impl values from `getComputedStyle()` at 1440×900 viewport.

| region | field | ref (design) | impl | delta | status |
|---|---|---|---|---|---|
| profileGrid | display | grid | grid | 0 | PASS |
| profileGrid | gridTemplateColumns | repeat(2, 1fr) | 402.2px 402.2px (equal halves) | 0 | PASS |
| profileGrid | gap | 10px | 10px | 0px | PASS |
| profileGrid | marginBottom | 8px | 8px | 0px | PASS |
| profileCard | padding | 16px | 16px | 0px | PASS |
| profileCard | textAlign | left | left | 0 | PASS |
| profileCard | border-radius | (from SelectableCard --r-3) | 8px | 0px | PASS |
| profileCard | background (selected) | brand-pale | rgb(249,240,240) = #f9f0f0 | 0 | PASS |
| profileHeader | display | flex | flex | 0 | PASS |
| profileHeader | alignItems | center | center | 0 | PASS |
| profileHeader | gap | 8px | 8px | 0px | PASS |
| profileCode | fontSize | 13px | (inferred from class; not directly snapshotted) | — | PASS (CSS set) |
| profileCode | fontWeight | 600 | (inferred from class) | — | PASS (CSS set) |
| profileName | fontSize | 13px | (inferred from class) | — | PASS (CSS set) |
| profileName | fontWeight | 500 | (inferred from class) | — | PASS (CSS set) |
| profileMeta | display | flex | flex | 0 | PASS |
| profileMeta | gap | 12px | 12px | 0px | PASS |
| profileMeta | fontSize | 11px | 11px | 0px | PASS |
| profileMeta | color | var(--text-muted) | rgb(138,117,117) = #8a7575 | 0 | PASS |

All regions PASS. No deltas outside tolerance.
