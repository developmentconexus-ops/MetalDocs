# Screen Definition of Done

> **Last verified:** 2026-06-21
> **Scope:** The binding close gate for every in-scope user-facing screen in the
> `frontend-screen-completion` program (milestones M1–M5). A screen is **not done** until every
> criterion below is met **on record**.
> **Governing program:** [`mission.md`](../../docs/superpowers/milestones/frontend-screen-completion/mission.md)
> (decisions **D2** = this gate, **D3** = the cut list).

## When to use

Cite this doc from every M1–M5 screen feature's `spec.md`/`evidence.md`. The per-milestone
`milestone-validator` checks each screen feature against it. This is the **WHAT-counts-as-done**
contract; the runtime functional pass it composes lives in
[`screen-qa-checklist.md`](screen-qa-checklist.md) (not duplicated here).

## The gate (D2) — all criteria mandatory

A screen counts done only when **all** of the following are true and evidenced in the feature's
`evidence.md`:

| # | Criterion | Objective proof |
|---|-----------|-----------------|
| 1 | **Real API data** — no mock/illustrative data | `grep -rE "MOCK_\|illustrative\|em breve" <feature dir>` = **0**; the screen renders from a real query hook against a live endpoint |
| 2 | **Redesign design-system tokens** — no ad-hoc styling | styling uses the settled tokens/primitives; no new inline-style blocks or hard-coded colors introduced (consume, don't redesign) |
| 3 | **`frontend-screen-reviewer` APPROVE on record** | the visual/parity reviewer's verdict is **APPROVE** (vs `design-source/<slug>` + tokens + primitive contracts), captured in the feature evidence |
| 4 | **`frontend-code-reviewer` APPROVE on record** | the architecture/maintainability reviewer's verdict is **APPROVE** (structure, decomposition, state design, type safety), captured in the feature evidence |
| 5 | **Tests green** | `make test` (vitest) green and `npm run build` clean; new tests use the canonical framework for their class |
| 6 | **Runtime functional pass** | the screen passes [`screen-qa-checklist.md`](screen-qa-checklist.md) for the acting role, with runtime proof recorded |

> **Both reviewers, both APPROVE, on record.** A single reviewer APPROVE is **not** sufficient (D2).
> A "looks right" or "works" claim without the two on-record verdicts + green tests is a FAIL.

### Backend-blocked screens (D1)

When a screen is blocked by a missing endpoint (Distribuição fanout/coverage M2, Notifications M3, any
Publicado-stub backend M4), the endpoint is built as its **own** feature to the Grade-A bar
(contract-first, ADR, `api-lint -strict` = 0, all 6 CI guards green, integration-tested) **before** the
screen's criterion 1 can be met. The screen does not "count done" on a stubbed endpoint.

## Evidence expectation

Each screen feature's `evidence.md` records: the route exercised, acting role, the real query/endpoint
proof, the two reviewer verdicts (APPROVE), the `make test` / `npm run build` results, and any bounded
defer with a written trigger.

## CUT registry (D3)

These slugs are **CUT** — never built. No route, no `NOTES.md`, no product intent. The router must not
mount them; a screen feature must never target them.

| Slug | Disposition | Rationale (D3) |
|------|-------------|----------------|
| `alternativas-inicio-caixa` | **CUT** | No route, no NOTES, no product intent. Not a gap to close. |
| `catalogo-slots` | **CUT** | No route, no NOTES, no product intent. Not a gap to close. |
| `biblioteca` | **Not cut — already shipped** | Delivered as `LibraryPage` (`features/documents/pages/LibraryPage`). Listed here only to prevent it being mistaken for a gap. |

Re-measure (must stay true every milestone):
`grep -rEn "alternativas-inicio-caixa|catalogo-slots" frontend/apps/web/src` → **0**.

## Related

- [screen-qa-checklist.md](screen-qa-checklist.md) — the runtime functional checklist criterion 6 composes
- [qa-operating-system.md](qa-operating-system.md) — the umbrella QA discipline
- [test-discipline.md](test-discipline.md) — canonical test-framework rule for criterion 5
- [screen-redesign-tracker.md](../implementation/screen-redesign-tracker.md) — per-screen status resume doc
