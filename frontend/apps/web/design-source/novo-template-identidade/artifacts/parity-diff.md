# Parity-diff · novo-template-identidade

Reference: `design-source` preview port 4181 `/novo-template-identidade/novo-template-identidade.html`.
Implementation: `metaldocs-web` preview port 4174 `/templates/new?step=2` (generic scope).

| Region | Field | Reference | Implementation | Delta | Status |
|---|---|---|---|---|---|
| h2 | font-size | 20px | 22px | +2 | accepted — uses canonical `.h2` global; design HTML used scoped override |
| h2 | line-height | 25px | 27.5px | +2.5 | derived from font-size — accepted with above |
| h2 | margin-bottom | 20px | 6px | -14 | accepted — design HTML had inline `mb`; canonical `.h2` ships 6px |
| Name input | font-size | 14px | 14px | 0 | ✓ |
| Name input | height | 38px | 38px | 0 | ✓ |
| Name input | padding | 0 10px | 8px 12px | diff | accepted — global `input:not(...)` rule (`0.5rem 0.75rem`) overrides; text vertical-centers correctly in 38px box |
| Name input | border color | rgb(212,194,194) | rgb(230,220,220) | diff | accepted — uses `var(--border)` canonical token; ref design HTML used standalone shade |
| Textarea | font-size | 13px | 13px | 0 | ✓ |
| Textarea | padding | 10px | 10px | 0 | ✓ |
| Textarea | height | derived from rows=3 | **32px → 72px (after fix)** | resolved | **fixed** — `.input { height:32px }` global clobbered; added `height:auto; min-height:72px` to `.descriptionInput` |
| Textarea | border color | rgb(212,194,194) | rgb(212,194,194) | 0 | ✓ |
| Code preview value | font-size | 26px | 26px | 0 | ✓ |
| Code preview value | font-weight | 600 | 600 | 0 | ✓ |
| Code preview value | color | rgb(107,31,42) `var(--brand)` | rgb(107,31,42) | 0 | ✓ |
| Code preview card | border | dashed `var(--brand-soft)` | 0.8px dashed rgb(139,46,58) | 0 | ✓ |
| Code preview card | background | `var(--brand-pale)` | rgb(249,240,240) | 0 | ✓ |
| Code preview card | border-radius | `var(--r-3)` | 8px | 0 | ✓ |
| Recap box | background | `var(--surface-2)` | rgb(250,246,246) | 0 | ✓ |
| Recap box | border | `var(--border)` | rgb(230,220,220) | system-token | accepted — same shade as name input |
| Recap box | radius | `var(--r-2)` | 6px | 0 | ✓ |
| Recap box | padding | 10px var(--sp-3) | 10px 12px | 0 | ✓ |

## Verdict

One real defect found and fixed (textarea height leakage). Remaining deltas are all system-level token alignments — the impl uses canonical MetalDocs tokens (`var(--border)`, `.h2`, `.input`) which are the source of truth; the design HTML used standalone shades that drifted from current tokens. No further fixes needed.
