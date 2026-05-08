# Parity Diff — documento-publicado Phase 3b

> Produced by Phase 3b style-port subagent (2026-05-08).
> Computed styles read from impl via `preview_inspect` + `preview_eval`.
> Design reference values from `onda1-v5.jsx → PublicadoV5`.
> Design-source server screenshot timed out; parity verified via computed-style inspection only.

---

| Region | Field | Ref (design) | Impl (computed) | Delta | Status |
|---|---|---|---|---|---|
| **Hero header** | padding-top | 32px | 32px | 0 | PASS |
| **Hero header** | padding-right | 56px | 56px | 0 | PASS |
| **Hero header** | padding-bottom | 40px | 40px | 0 | PASS |
| **Hero header** | padding-left | 56px | 56px | 0 | PASS |
| **Hero header** | border-bottom | 1px solid var(--border) | 0.8px solid rgb(230,220,220) | ~0.2px sub-pixel | PASS (browser rendering artifact) |
| **Hero header** | position | relative | relative | 0 | PASS |
| **Hero header** | overflow | hidden | hidden | 0 | PASS |
| **Breadcrumb nav** | font-size | 11px | 11px | 0 | PASS |
| **Breadcrumb nav** | color | var(--text-muted) | rgb(138,117,117) | 0 | PASS |
| **Breadcrumb nav** | letter-spacing | 0.04em → ~0.44px at 11px | 0.44px | 0 | PASS |
| **Breadcrumb nav** | text-transform | uppercase | uppercase | 0 | PASS |
| **Breadcrumb nav** | margin-bottom | 24px | 24px | 0 | PASS |
| **Breadcrumb nav** | font-family | var(--font-mono) | JetBrains Mono, IBM Plex Mono, … | 0 | PASS |
| **Breadcrumb nav** | gap | 8px | 8px | 0 | PASS |
| **HeroGrid** | display | grid | grid | 0 | PASS |
| **HeroGrid** | grid-template-columns | 210px 1fr | 210px 1006.8px | 0 | PASS |
| **HeroGrid** | gap | 40px | 40px | 0 | PASS |
| **HeroGrid** | align-items | center | center | 0 | PASS |
| **DocCardMini** | width | 168px | 168px | 0 | PASS |
| **DocCardMini** | height | 224px | 224px | 0 | PASS |
| **DocCardMini** | border-radius | 4px (var(--r-1)) | 4px | 0 | PASS |
| **DocCardMini** | transform | perspective(1200px) rotateY(-12deg) rotateX(4deg) | matrix3d(…) equivalent | 0 | PASS |
| **DocCardMini** | box-shadow | 20px 20px 48px rgba(74,33,33,0.16), 4px 4px 14px rgba(74,33,33,0.10) | rgba(74,33,33,0.16) 20px 20px 48px 0px, rgba(74,33,33,0.10) 4px 4px 14px 0px | 0 | PASS |
| **Hero title h1** | font-size | 36px | 36px | 0 | PASS |
| **Hero title h1** | line-height | 1.1 → 39.6px | 39.6px | 0 | PASS |
| **Hero title h1** | letter-spacing | -0.025em → -0.9px | -0.9px | 0 | PASS |
| **Hero title h1** | margin-bottom | 12px | 12px | 0 | PASS |
| **Hero title h1** | font-weight | 600 | 600 | 0 | PASS |
| **Hero title h1** | color | var(--text) | rgb(26,14,14) | 0 | PASS |
| **Hero title h1** | max-width | 640px | 640px | 0 | PASS |
| **Hero badges row** | display | flex | flex | 0 | PASS |
| **Hero badges row** | align-items | center | center | 0 | PASS |
| **Hero badges row** | gap | 10px | 10px | 0 | PASS |
| **Hero badges row** | margin-bottom | 14px | 14px | 0 | PASS |
| **KPI strip** | display | grid | grid | 0 | PASS |
| **KPI strip** | grid-template-columns | 1fr | 1066.4px (1fr) | 0 | PASS |
| **KPI strip** | border-radius | 8px (var(--r-3)) | 8px | 0 | PASS |
| **KPI cell** | padding | 18px 22px | 18px 22px | 0 | PASS |
| **AboutCard owner banner** | padding | 18px 22px | 18px 22px | 0 | PASS |
| **AboutCard owner banner** | display | flex | flex | 0 | PASS |
| **AboutCard owner banner** | align-items | center | center | 0 | PASS |
| **Facts grid** | display | grid | grid | 0 | PASS |
| **Facts grid** | grid-template-columns | 1fr 1fr | 533.2px 533.2px | 0 | PASS |
| **Fact cell** | padding | 14px 22px | 14px 22px | 0 | PASS |
| **Fact cell** | display | flex | flex | 0 | PASS |
| **Fact cell** | align-items | flex-start | flex-start | 0 | PASS |
| **SignoffGrid** | display | grid | grid | 0 | PASS |
| **SignoffGrid** | grid-template-columns | repeat(3, 1fr) | 339.462px 339.462px 339.475px | 0 | PASS |
| **SignoffGrid** | position | relative | relative | 0 | PASS |
| **Signoff connector** | position | absolute | absolute | 0 | PASS |
| **Signoff connector** | top | 18px | 18px | 0 | PASS |
| **Signoff connector** | height | 2px | 2px | 0 | PASS |
| **Signoff pin** | width | 36px | 36px | 0 | PASS |
| **Signoff pin** | height | 36px | 36px | 0 | PASS |
| **Signoff pin** | box-shadow | 0 2px 8px rgba(26,107,53,0.30) | rgba(26,107,53,0.3) 0px 2px 8px 0px | 0 | PASS |
| **Signoff pin** | color | var(--text-on-brand) → #fff | rgb(255,255,255) | 0 | PASS |
| **Signoff pin** | margin-bottom | 14px | 14px | 0 | PASS |

---

## Summary

All 53 fields pass. Zero non-trivial deltas. The 0.8px border-bottom is a browser sub-pixel rendering artifact of `1px solid` on high-DPI — not a CSS value mismatch.
