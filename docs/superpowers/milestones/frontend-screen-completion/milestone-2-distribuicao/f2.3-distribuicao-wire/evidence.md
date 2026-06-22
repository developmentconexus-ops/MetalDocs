# Feature F2.3 — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.3-distribuicao-wire`  ·  **Closed:** 2026-06-22
> **Spec:** `spec.md` (Validation Gate V1–V10 this proves against) · **Plan:** `plan.md`.
> Closed only when every row is filled with **real, honestly-labeled** output. Where a runtime-with-live-API
> screenshot could not be captured (no full stack up this session), that is stated as a **bounded gap**, not
> passed off as proven.

## What was implemented

`DocumentDistributionPage` (`/documents/:documentId/distribution`) now renders the **real obligated-reader
denominator** from the three F2.2 endpoints; the **numerator** (read/ack/timeline/per-recipient status) is
replaced by a single honest **"tracking pending"** disclosure; the 4 hero CTAs stay deferred-with-trigger;
the illustrative scaffolding is deleted at root.

- **API layer** `features/documents/api/distribution.ts` — `getDistributionSummary`,
  `listDistributionRecipients({cursor,limit})`, `getDistributionCoverage`; type aliases re-exported from the
  generated `components['schemas']['Distribution*']` (no hand-rolled shapes).
- **Query hooks** `queries/useDistributionSummaryQuery.ts`, `useDistributionRecipientsQuery.ts`
  (`useInfiniteQuery`, `getNextPageParam` ← `page.next_cursor`/`has_more`), `useDistributionCoverageQuery.ts`;
  keyed under `QK.documents.distribution.*`.
- **Honest disclosure** `components/distribution/TrackingPendingNote.tsx` (+`.module.css`) — one shared
  `role="note"` atom, **zero numbers**, links the parked mission. Used by DonutCard, DistributionFacts,
  TimelineCard, KPIStrip (every numerator surface).
- **Denominator components refactored to props** — `RecipientsCard` (items + name + origin chip from
  `source`/`area_name`; "Carregar mais" keyset pagination), `CoverageByArea` (area_name + total), `KPIStrip`
  (`totalTargets`). Status tabs / read-ack columns / bars dropped.
- **Page** wires the three hooks; removed `IllustrativeBlock`, the `Dados ilustrativos · Em breve` watermark,
  the `Em breve` hero badge, the "números ilustrativos" banner; honest loading/empty/error; per-section error
  states (a single sub-query failure no longer nukes the page); errored total renders `"—"`, never a fabricated 0.
- **Root deletion** — `lib/distributionMeta.ts` (`MOCK_DISTRIBUTION` + all mock interfaces) deleted; zero importers.

### Commit list
```
9f799a0a  docs(M2/F2.3): approved spec + plan (distribuicao-wire)
e06f8415  feat(M2/F2.3): wire DocumentDistributionPage to live denominator endpoints
<this>     docs(M2/F2.3): close evidence — distribuicao wired
```

## Verification (real output)

| # | Criterion | Command | Result | Real vs fixture |
|---|-----------|---------|--------|-----------------|
| V1 | Illustrative literals gone at root | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" …/DocumentDistributionPage.tsx`; `grep -rn MOCK_DISTRIBUTION src/features/documents` | both **0 matches** (exit 1) | real |
| V2 | Denominator renders live | `useDistribution.test.tsx` V2a/V2b/V2c (typed fixtures of the generated schemas) — total, recipient name+origin chip, by-area totals render | PASS | fixture (typed to generated contract); **runtime-with-live-API = bounded gap, see below** |
| V3 | Numerator is honest tracking-pending | `useDistribution.test.tsx` V3 + reviewer audit — `TrackingPendingNote` carries no number; `totalTargets` error→`"—"` not 0 | PASS | real (static) + fixture |
| V4 | CTAs deferred-with-trigger | `useDistribution.test.tsx` V4 — 4 CTAs `aria-disabled`, `title` → parked mission, no handlers | PASS | real |
| V5 | Typed cursor pagination | `useDistribution.test.tsx` V5/V5b — "Carregar mais" appends page 2 via `next_cursor`; button gone when `has_more=false` | PASS | fixture |
| V6 | Hook/page test green | `npx vitest run useDistribution.test.tsx DocumentDistributionPage.test.tsx` | **16/16 PASS** | real |
| V7 | Type safety / generated types only | `npx tsc --noEmit` (web) | **exit 0** | real |
| V8 | FE suite holds at baseline | `npx vitest run` (full) | 4 files / 36 tests fail — **all pre-existing**, proven identical on clean HEAD (see below); `DocumentDistributionPage.test.tsx` 6/6, distribution 16/16 | real |
| V9 | Both reviewers APPROVE | `frontend-screen-reviewer` + `frontend-code-reviewer` | **both APPROVE** (after one REQUEST-CHANGES → fix cycle) | real |
| V10 | No out-of-scope change | `git show --stat e06f8415` | only `features/documents/**` + `lib/queryKeys.ts`; no backend, no type regen, no shared primitive/token, no M0/M1 | real |

### V8 — FE suite failures are pre-existing (proven, not assumed)
The full web vitest run shows **4 failed files / 36 failed tests**: `approval/pages/InboxPage.test.tsx`,
`documents/pages/DocumentEditorPage.test.tsx`, `documents/__tests__/DocumentEditorPage.test.tsx`,
`templates/api/__tests__/templates.create.test.ts`. **None** touch distribution and **none** import any F2.3
module. Proven pre-existing by stashing the F2.3 working tree (`git stash push -u`) and re-running those four
files on **clean HEAD** → **identical 36 failed / 1 passed**. They are the known FE junction-drift / legacy-
scaffold baseline (memory: `fe-node-modules-junction-drift`, `legacy-test-deletion`), not an F2.3 regression.
The F2.3 target test (`DocumentDistributionPage.test.tsx`) is green (6/6); the new suite is 16/16.

## Review disposition

- **`frontend-code-reviewer`:** first pass **APPROVE-WITH-NITS** conditioned on 2 Majors — M1 (coarse page-level
  `distError` early-return nuked a loaded recipient list on any one sub-query failure + made per-section error
  props dead) and M2 (dead exported `RecipientStatus`/`RECIPIENT_STATUS_TONE`). Both fixed; re-review **APPROVE**
  (adversarially confirmed no fabricated/coerced 0; the multi-failure test asserts the new per-section behavior).
- **`frontend-screen-reviewer`:** first pass **REQUEST CHANGES** — 3 Majors: (1) `TrackingPendingNote.module.css`
  referenced **8 undefined CSS custom properties** (`--space-*`, `--radius-md`, `--text-*`, `--font-semibold`,
  `--leading-relaxed`) → rendered unstyled at runtime on every numerator surface; (2) "Carregar mais" reused the
  28×28 icon-button class → clipped text; (3) `CoverageByArea` `.areaRow` grid had 3 columns for 2 children. All
  fixed (verified token mapping against `tokens.css`: `--sp-*`, `--r-3`, `--font-size-sm/xs`, literal `600`/`1.5`;
  `--text-faint` confirmed to exist); re-review **APPROVE**.
- Controller independently verified the undefined-token claim against `tokens.css` before dispatching the fix,
  re-verified every fix (tsc 0, 16/16, greps 0, dead `.paginationBtn` removed), and proved the V8 baseline.

## Bounded gaps / defers

| Gap / defer | Why bounded | Trigger / owner |
|-------------|-------------|-----------------|
| **V2/V3 runtime-with-live-API screenshot** | No full backend stack (API :8081 down; screen-reviewer hit the auth wall). Logic is covered by the 16 typed-fixture tests + static review; the contract is the same one F2.2 serves live (gate green). | Operator HS-1 review can spot-check at runtime; or capture on next stack-up. The producer is proven (F2.2 evidence). |
| Pre-existing FE suite failures (Inbox/Editor/Templates) | Outside F2.3 scope; proven identical on clean HEAD. Known junction-drift / legacy-scaffold baseline. | Tracked by `fe-node-modules-junction-drift` / `legacy-test-deletion`; not this feature. |
| Numerator UI (donut/timeline/per-recipient read-ack), status-tab filtering, the 4 actions | Explicit non-goal — no producer exists (HS-6). Rendered as honest tracking-pending / deferred-with-trigger. | Parked: `wiki/backlog/document-distribution-mission.md`. |
| E2E Playwright for the pagination flow | Spec V6 accepts vitest; covered against fixtures. | Optional follow-up cleanup PR. |
