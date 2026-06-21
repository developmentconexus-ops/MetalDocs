# Feature F1.1 — Spec

> **Milestone:** 1 — Dashboard real data  ·  **Folder:** `f1.1-dashboard-stats-wire`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *re-scope contract locked via the milestone HS interview.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (`plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

The contract ambiguity here was the **producer gap** (designed cards have no data source). It was
resolved with the operator at milestone-spec time (the HS contract gate), not invented. That dialog
IS this feature's interview; persisted below.

| # | Question | Answer |
|---|----------|--------|
| 1 | The 3 hero "§ SEU PULSO" cards (Aprovados esta semana / Devolvidos aguardando / Tempo médio por decisão) have no backend producer. `/iam/kpi` is security KPIs (wrong domain); `/documents/stats` returns only `by_status`+`by_area` counts (no time window, no timing). How should F1.1's contract be set? | **Operator (2026-06-21):** re-scope the cards to what `/documents/stats` truthfully serves — honest `by_status` counts. Do **not** build an approval-throughput endpoint (refused for M1); do **not** fabricate a window/average. |
| 2 | `by_status` has no literal "devolvido/returned" key. Enum is `draft, under_review, approved, rejected, scheduled, published, obsolete, superseded, archived`. Which keys back the re-scoped cards? | **Resolved from the enum (source of truth `documents/domain/model.go` + generated `DocumentSummaryStatus`):** Aprovados = `approved`; Em revisão / devolvidos = `under_review` + `rejected` (rejected is the returned-to-author state); Publicados = `published`. Missing keys count as 0 (do not invent). |
| 3 | Is the existing "Aguardando você" hero pill (approval inbox total) in scope? | **No** — already live via `useDashboardInboxQuery`. Untouched. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `frontend/apps/web/src/features/dashboard/pages/DashboardPage.tsx` — the hero
  `heroStats` pills row (the 3 non-"Aguardando você" pills) **and** the sidebar "§ SEU PULSO"
  `statRow` list. Both currently map over the `MOCK_STATS` constant.
- **Contract (the shape the consumer needs):** an array of three derived stat items
  `{ label: string; value: string; sub: string; color }`, computed from the live producer:
  - `Aprovados` — value = `by_status.approved ?? 0` — sub `no acervo`.
  - `Em revisão` — value = `(by_status.under_review ?? 0) + (by_status.rejected ?? 0)` — sub `aguardam ajuste`.
  - `Publicados` — value = `by_status.published ?? 0` — sub `no acervo`.
  - Loading → `…`; error/empty → `—` (honest non-value, the existing placeholder), never a fabricated number.
- **Producer (must match consumer):** `GET /api/v1/documents/stats` →
  `DocumentStatsResponse { by_status: { [k: string]: number }, by_area: { [k: string]: number } }`.
- **Source of truth for the contract:** generated `components['schemas']['DocumentStatsResponse']`
  (`frontend/apps/web/src/lib/api-types/index.d.ts`); existing FE client `fetchLibraryStats()`
  (`features/documents/api/library.ts`); status enum `documents/domain/model.go` /
  generated `DocumentSummaryStatus`.

## What this feature implements

Delete the `MOCK_STATS` constant from `DashboardPage.tsx`. Add a dashboard stats query
(`useDashboardStatsQuery`, mirroring `useDashboardInboxQuery`: `useQuery` + `QK.documents.stats()`
+ `fetchLibraryStats`). Derive the three stat items from `by_status` per the contract above and
render them in **both** the hero pills row and the "§ SEU PULSO" sidebar list, with loading/`…`
and error/`—` states. No backend change.

## Non-goals (mandatory)

- No new or modified backend endpoint; no Go change of any kind.
- No "esta semana" time-window and no "tempo médio por decisão" average — no producer exists; out of scope.
- No change to the live "Aguardando você" pill or the pending-approvals list.
- No `by_area` rendering (available but not part of these cards).
- No dashboard layout/visual redesign beyond swapping the data source of the existing pills/rows.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `MOCK_STATS` gone from the dashboard feature | `grep -rnE "MOCK_STATS" frontend/apps/web/src/features/dashboard` → exit 1 / 0 matches | real |
| Stats query hook maps `by_status` → 3 items correctly (approved; under_review+rejected; published; missing→0) | new `useDashboardStatsQuery` / derivation unit test passes against a fixtured `DocumentStatsResponse` | fixture (labeled) |
| Cards render live values at runtime (not `—` when data present) | preview: load `/`, assert the 3 pills show numeric values from a seeded/real stats response | real |
| Type-safe consumer shape | `npm run build` (tsc) exit 0 — consumer uses generated `DocumentStatsResponse`, no hand-rolled shape | real |
| Suite green | `make test` (dashboard + changed scope) green; no new failures vs baseline | real |
| Both reviewer agents APPROVE (D2) | `frontend-screen-reviewer` + `frontend-code-reviewer` verdicts on record in `evidence.md` | real |

> TDD: the derivation/query-hook test is written failing first, then implemented to green. The
> hook test is fixture-backed (labeled); the runtime render check is the real-provider proof.

## ADR needed?

- [x] No durable architectural decision — skip. The re-scope is a product/contract choice recorded
  in `milestone.md` + this interview; it sets no reusable architectural rule. (If a future milestone
  builds the approval-throughput endpoint, that is ADR-worthy then.)
