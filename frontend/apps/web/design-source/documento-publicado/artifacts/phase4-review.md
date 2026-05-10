# Screen Review: documento-publicado

**Implementation:** `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx` + `DocumentPublishedPage.module.css`
**Design source:** `frontend/apps/web/design-source/documento-publicado/`
**Visual comparison:** ✅ 1440 + 375 — computed-style numerical parity (`parity-diff.md`, 99 fields). Screenshot files deferred (preview tool limitation, see IMPLEMENTATION.md OQ-3).
**Pass:** Second pass (2026-05-08), all findings addressed.
**Verdict:** APPROVE WITH NITS

## Critical

None remaining.

- ~~CUT KPI cells "Próxima revisão" + "Páginas" rendered in DOM~~ — **RESOLVED**: removed JSX blocks at `DocumentPublishedPage.tsx:231–243`, `kpiStrip` grid updated to `repeat(2, 1fr)`, mobile rule simplified.

## Major

None remaining.

### Visual / numerical parity
- ~~Stale parity-diff row for SectionKicker~~ — **RESOLVED**: row updated, summary bullet rewritten, status PASS.

### Architecture
None.

### Tokens / primitive drift
None.

### NOTES.md compliance
All Keep/Cut/Defer decisions honored. CUT KPIs removed; DEFER items render as `—` placeholders or `aria-disabled` buttons.

### Error UX
None. No mutations on this page. Clipboard handler now sets `linkCopied` only inside `.then()` (false-positive removed).

### A11y
None. Pin buttons `aria-label` + `aria-pressed`, breadcrumb labelled nav, error `role="alert"`, loading `role="status" aria-live="polite"`, `prefers-reduced-motion` covered in timeline.

### Responsive
None. 375 verified: no overflow, hero card hidden, KPI strip 2-col grid intact.

### Iron-Law artifacts
- `parity-diff.md`: 99 fields, all regions. Stale row repaired, deviation row reclassified.
- `leakage-probe.md`: extended with VersionTimeline pin button + replyRow probes (both clean).
- `token-coverage.txt`: empty (no token bypass).
- `IMPLEMENTATION.md`: OQ-3 documents screenshot tooling limitation.

## Minor

- ~~Clipboard false-positive~~ — **RESOLVED**: `setLinkCopied(true)` moved into `.then()` branch.
- ~~Timeline pin justify-content unresolved~~ — **RESOLVED**: in-source comment marks ACCEPTED-DEVIATION (`DocumentVersionTimeline.module.css:73`); parity-diff status updated.

## Open structural decision (not a defect)

`parity-diff.md` — VersionTimeline detail panel: design renders inline meta row (flex/gap/mt/fs:11/no bg); impl renders full filled card (bg `--bg-soft`, br 6px, padding 16/18). Both verified live. Documented as decision-required Keep-extension. No automatic action — team decides whether to strip card chrome or accept extension.

## Confirmed fixes (cumulative across both passes)

| ID | Item | Status |
|---|---|---|
| C-1 | `canInitiateRevision` missing `system_admin` role | RESOLVED |
| C-2 (2nd pass) | CUT KPIs "Próxima revisão" + "Páginas" in DOM | RESOLVED |
| M-1 | Phase 3b screenshot files missing | DOCUMENTED (tooling limit) |
| M-2 | parity-diff missing 4 regions | RESOLVED (+46 fields) |
| M-3 | typeLabel placeholder absent | RESOLVED |
| M-4 | SectionKicker font-weight 600 vs design 500 | RESOLVED |
| M-5 (2nd pass) | parity-diff stale SectionKicker row | RESOLVED |
| m-1 | leakage-probe gaps (pin button, replyRow) | RESOLVED |
| m-2 | `documentId!` non-null assertion | RESOLVED |
| m-3 (2nd pass) | Timeline pin justify-content unannotated | RESOLVED |
| m-4 (2nd pass) | Clipboard `setLinkCopied` false-positive | RESOLVED |

## What's good

- Token coverage airtight: zero raw hex/rgb in CSS Module.
- TanStack Query + `QK.*` keys + `apiFetch` only — architecture clean.
- A11y: aria-labels on pins, focus-visible rings, reduced-motion guard.
- Mobile 375 numerically passes: no overflow, all breakpoints fire.

## Iron-Law cross-check

- Phase 0 audit signed: ✅
- Phase 1 worksheet complete: ✅
- Phase 2 primitive audit verified: ✅
- Phase 3a DOM diff approved: ✅
- Phase 3b: parity-diff covers all regions ✅; leakage-probe covers form els ✅; token-coverage empty ✅; numbers match reality ✅
- Phase 4 behavior trace present: ✅
- Phase 4.5 reviewer pass: ✅ (this report; second pass clean)
- Open Questions Log resolved: ✅ (3 rows, all closed)
