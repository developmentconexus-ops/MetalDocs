# Feature F2.3 — Plan (distribuicao-wire)

> **Spec:** `./spec.md` (the contract this plan executes). Plan = **how**; spec = **what/proven**.
> **Subagent model:** Sonnet 4.6 (operator directive 2026-06-21). Controller (this session) does not
> implement — it dispatches a fresh implementer, then the two frontend reviewer agents, then closes.
> **Discipline:** test-first where a test is meaningful (query hooks); root-removal of mock, not flag-hiding.

## Boundary / invariants (carry into every subagent prompt)

- **Frozen consumer contract** — generated types `Distribution*` in `lib/api-types/index.d.ts`. Reference
  `components['schemas']['Distribution*']`; never hand-roll a duplicate shape. **No type regen.**
- **No backend change.** No new shared primitive/token (HS-2). No M0/M1 surface. No action-layer behavior (HS-6).
- **Numerator = honest tracking-pending**, never a fabricated number. Mock numerator data is **deleted**, not repurposed.
- **CTAs stay `aria-disabled` deferred-with-trigger** → `wiki/backlog/document-distribution-mission.md`.
- Touch only `frontend/apps/web/src/features/documents/**` + `lib/queryKeys.ts` + the new api/query/test files.
- Commits allowed after verified work; **no push**, **no merge**.

## Tasks

### T0 — Prereq checkpoint (controller, before dispatch)
- FE dev server starts; `DocumentDistributionPage` route reachable; generated `Distribution*` types present.
- Baseline `DocumentDistributionPage.test.tsx` green; `npm run typecheck` (web) green.
- Confirm the three F2.2 endpoints are reachable in a running stack **or** that hook tests will use fixtures (HS-3 — if the route/contract is broken, repair first). **Gate:** all green → dispatch T1+.

### T1 — API layer + query keys
- New `frontend/apps/web/src/features/documents/api/distribution.ts`:
  - `getDistributionSummary(id): Promise<DistributionSummaryResponse>` → `apiFetch('/api/v1/documents/{id}/distribution')`.
  - `listDistributionRecipients(id, { cursor?, limit? }): Promise<DistributionRecipientsResponse>` (querystring-encode cursor/limit).
  - `getDistributionCoverage(id): Promise<DistributionAreaCoverage[]>`.
  - Export the type aliases from `components['schemas']`.
- `lib/queryKeys.ts`: add `QK.documents.distribution = { summary(id), recipients(id, params), coverage(id) }`.

### T2 — RED: hook/page test against fixtures
- New test (`queries/__tests__/useDistribution*.test.tsx` and/or extend `DocumentDistributionPage.test.tsx`):
  fixtured responses typed to the generated schemas. Assert: loading→data; live total/recipients/by-area
  render; **numerator tracking-pending state present** (query by accessible text); CTAs `aria-disabled`;
  **"Carregar mais"** appends page 2 via `next_cursor`; empty-state (`total_targets:0`, `items:[]`) honest;
  error→retry. Run → **RED** (hooks/components not built yet). Capture the failing output.

### T3 — GREEN: query hooks
- `useDistributionSummaryQuery(id)` — `useQuery`, `enabled: Boolean(id)`.
- `useDistributionRecipientsQuery(id, { limit? })` — `useInfiniteQuery`, `getNextPageParam` from `page.next_cursor`/`has_more`.
- `useDistributionCoverageQuery(id)` — `useQuery`.
- Run test → **GREEN** for the hook-level assertions.

### T4 — Shared tracking-pending component
- New presentational `components/distribution/TrackingPendingNote.tsx` (+ `.module.css`): heading +
  one-line honest disclosure (PT), link/note to the parked mission. No numbers. Tokens only. Reused by
  every numerator surface so the disclosure is identical and the old literals vanish.

### T5 — Refactor denominator components to props
- `RecipientsCard` → accepts `items: DistributionRecipient[]`, `hasMore`, `onLoadMore`, loading/error;
  renders **name + origin chip** (from `source`/`area_name`); drop status tabs + read/ack/when columns.
- `CoverageByArea` → accepts `rows: DistributionAreaCoverage[]`; renders area_name + total (denominator
  only — drop read/ack bars).
- Total surface (`KPIStrip`/`DistributionFacts`) → accepts `totalTargets`; numerator KPIs replaced by
  `TrackingPendingNote` (or removed where they were pure numerator).
- `DonutCard`, `TimelineCard` → replaced by `TrackingPendingNote` at the page level (numerator-only).

### T6 — Wire the page
- `DocumentDistributionPage.tsx`: call the three hooks; remove `IllustrativeBlock`, the
  `Dados ilustrativos · Em breve` watermark, the `Em breve` hero badge, the "números… ilustrativos"
  banner. Wire denominator surfaces live; render numerator surfaces via `TrackingPendingNote`; keep the
  4 CTAs `aria-disabled` with a trigger note. Honest loading/empty/error per spec §8.

### T7 — Root-delete mock
- Remove `MOCK_DISTRIBUTION` + the now-dead mock interfaces/`RECIPIENT_TABS`/numerator types from
  `distributionMeta.ts` (keep only what a live surface still uses, if anything). Grep `MOCK_DISTRIBUTION`
  over `src/features/documents` = 0.

### T8 — Validation gate (controller verifies live)
- `grep -nE "Dados ilustrativos|MOCK_DISTRIBUTION|Em breve" …/DocumentDistributionPage.tsx` = 0 (V1).
- `npm run typecheck` (web) clean (V7); `make test` / web vitest green (V6/V8).
- Preview runtime: denominator live (V2), numerator tracking-pending (V3), CTAs disabled (V4),
  "Carregar mais" paginates (V5) — **screenshots on record**.

### T9 — Dual frontend review (separation of powers)
- Dispatch **`frontend-screen-reviewer`** (visual/architectural parity) and **`frontend-code-reviewer`**
  (code/maintainability). Fix-loop a fresh implementer until **both APPROVE** (V9). Controller
  independently re-verifies any RED→GREEN claim.

### T10 — Close
- Write `evidence.md` (real captured output per V1–V10, both reviewer dispositions, screenshots/labels,
  bounded defers). Commit `docs(M2/F2.3): close evidence — distribuicao wired`. **No push.**
- Then (M2 last feature) → dispatch **`milestone-validator`**; on PASS present HS-1 operator gate.

## Risks
- **R1 — components are mock-internal.** Refactor to props is the bulk of the work; keep each component's
  visual contract (design parity) intact — that is exactly what `frontend-screen-reviewer` guards.
- **R2 — recipient list has no total count.** Cursor-only; UX is "Carregar mais", not numbered pages.
- **R3 — numerator deletion tempts a fake `0`.** Mitigation: `TrackingPendingNote` carries no number; V3 grep guards.
- **R4 — running stack for V2/V3 runtime proof.** If the full stack isn't up, seed a doc with grants or
  use the preview against a live API; hook tests (V6) cover the logic regardless. HS-3 if the route is broken.
