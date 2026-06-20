# Milestone 8 — Grade-A Contract & Boundary Completion (HS-5, 4th-miss closure)

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** PASS — milestone-validator VERDICT PASS 2026-06-20 (`qa/milestone-qa.md`); all 6 features closed (F8.1–F8.6). HS-1 operator gate pending; post-M8 terminal re-audit + mission-validator pending.
> **Authored:** 2026-06-20 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone
> QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Close the **post-M7 terminal re-audit** gap (`wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m7.md`,
**VERDICT: FAIL**, all 4 §8 checks unmet) and reach the mission §8 Grade-A terminal bar:

- **Contract / API ≥ A−** and **Composition / observability ≥ A−** (both graded B+ at post-M7; each holds one live Major).
- **Sessions / auth lifecycle ≥ A−** (regressed to B+ on a confirmed CWE-613 Major).
- **0 skeptic-confirmed Critical/Major** (5 confirmed at post-M7).
- **Honest H-D = 0** — zero `map[string]any` response literals on *any* public route, not just `internal/modules/*/delivery/http/`.
- **Honest H-G = 0** — zero cross-module/cross-schema raw SQL against another module's owned table.

The bar moved is the §8 four-check terminal acceptance. The criterion that proves it: a fresh post-M8
re-audit + `mission-validator` returns **VERDICT: PASS** (4/4).

**Root cause this milestone fixes (why a 4th miss happened):** the §5b H-D gate and the H-G greps are
**scoped narrower than §8 intent** — H-D greps only `internal/modules/*/delivery/http/`, H-G greps only
the two IAM tables. Two response literals (`iam/presence`, `platform/observability/metrics`) and one
cross-schema reach (`search`→taxonomy `document_profiles`) sit just outside the scopes and survived every
prior bounded sweep. M8 fixes the instances **and** widens the gate to the true public surface so the
class is closed and non-regressable (operator decision 2026-06-20: option A — typed-everywhere + gate-scope honesty).

## Appetite & rabbit holes

- **Appetite:** one focused milestone — 6 bounded features, no open-ended rewrite. This is **not** the full
  codegen-first StrictServerInterface migration; it is targeted typed-body + port + gate-honesty closure
  (the §8 bar is grade/Major/class-count, not a framework).
- **Rabbit holes (named refusals):**
  - **No** repo-wide StrictServerInterface rewire of auth/search (no codegen pipeline there; out of appetite — ADR 0012 hand-rolled typed structs remain valid).
  - **No** fully-typed metrics sub-objects / provider-interface churn — `runtime`/`scheduler`/`db_pool` stay map-backed but **declared** `additionalProperties:true` (operator decision 2026-06-20).
  - **No** new product scope, **no** FE feature work, **no** migrations.
  - **No** session-token-rotation redesign — F8.4 is scoped to deactivation enforcement only.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F8.1 | `f8.1-presence-typed-body` | Presence snapshot route emits the already-generated `iamapi.PresenceSnapshotResponse` typed body instead of `map[string]any{"items":…}`. | `iam/presence/handler.go:83` no longer constructs a response `map[string]any`; handler returns `PresenceSnapshotResponse`; wire bytes byte-identical to pre-change (test); `go build`+`go test` green. |
| F8.2 | `f8.2-metrics-typed-envelope` | `/api/v1/metrics` top-level envelope typed (platform-local `MetricsResponse`/`MetricItem`, REQ-TOP-2 module-import-free); OpenAPI `MetricsResponse` declares `scheduler`+`db_pool` as `additionalProperties:true`; FE types regenerated. | `observability/http.go` emits no top-level response `map[string]any`; OpenAPI declares all emitted top-level keys; FE codegen regenerated & committed; wire shape unchanged (test). |
| F8.3 | `f8.3-search-taxonomy-port` | Add taxonomy `FamilyCodeResolver` read-port (batch resolve profile-code→family-code honoring sentinel-tenant fallback); search drops both raw `metaldocs.document_profiles` subqueries and resolves via the injected port. | `search/.../v2documents/reader.go` contains no `metaldocs.document_profiles` SQL; search imports a taxonomy port; family projection + family filter results byte-identical (test); H-G honest = 0. |
| F8.4 | `f8.4-deactivation-session-enforcement` | Deactivation revokes live sessions (revoke-on-`is_active=false`) **and** `ResolveSession`/`buildCurrentUser` re-checks `identity.IsActive` fail-closed. | Deactivating a user with a live session → that session is rejected at next resolve (test); inactive identity → resolve fails closed (test); change-password/reset revocation unchanged; build+test green. |
| F8.5 | `f8.5-problem-json-405` | Global middleware converts stdlib `text/plain` 405 (and 404) from method-routed `ServeMux` patterns into `application/problem+json`, preserving `Allow`. | `DELETE` on a GET-only route returns `application/problem+json` + correct `Allow` header (test); intra-module 405 envelope now consistent; build+test green. |
| F8.6 | `f8.6-gate-scope-widening` | Widen `api-contract.md` §5b Part A/B grep scope and the mission §8 H-D/H-G definitions to the whole public-route surface; add a `tools/cilint` (or grep) CI guard that fails on any response-literal `map[string]any` on a registered route. | §5b + §8 updated to the true surface; CI guard present and failing-on-violation (demonstrated); re-running the widened greps at HEAD returns 0/0 after F8.1–F8.5. |

For each feature, "what to validate" is objectively checkable: a grep that returns empty, a route that
responds with the contracted shape and content-type, a test that passes, a build that is clean.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M8 it enforces:

1. **Per-feature acceptance** — every feature meets its declared "what to validate"; each feature's
   consumer contract (`spec.md`) is honored (producer matches consumer; wire shapes unchanged where claimed).
2. **Workflow-class QA** — `wiki/quality/backend-api-checklist.md` (contract/typed-body, error-envelope,
   module-boundary) + `wiki/quality/qa-operating-system.md` close-out.
3. **Regression** — M0–M7 still pass their gates; full `go build ./...` + `go test -count=1 ./...` green.
4. **Quality-bar / root-cause check** — the honest H-D and H-G classes re-measured **with the widened
   §5b/§8 scope** = 0/0; the 5 post-M7 Majors confirmed fixed at root (not symptom-patched); Contract/API,
   Composition, Sessions each re-graded ≥ A−.
5. **No unplanned scope** — anything beyond F8.1–F8.6 recorded with rationale.

> M8's milestone-validator is the per-milestone gate. The program's **terminal** acceptance is separate:
> after HS-1 approval, the main session re-runs the 10-dimension post-M8 re-audit and dispatches the
> `mission-validator` against §8 (4/4). Grade-A is the operator's declaration on that PASS.

## Dependencies & constraints

- Depends on: M0–M7 passed (typed-body parity baseline; honest two-part H-D gate in `api-contract.md` §5b).
- Audited HEAD baseline: code `dadb8275` (post-M7).
- Constraints respected: **no migrations**; reads stay live; **platform stays module-import-free**
  (REQ-TOP-2 — metrics types are platform-local, not module-generated); contract-first regen order for
  F8.2 (OpenAPI → BE/FE codegen); advisory-lock hazard rules (H-PRE-1) for F8.4 session writes; ADR 0012
  hand-rolled typed structs remain the sanctioned pattern for non-codegen modules.
- F8.3 introduces a durable decision (new taxonomy port consumed cross-module) → **ADR required**.
- F8.6 amends a governing contract (§5b / mission §8) → operator-visible; record the scope change.

## Applicable hard-stops

- **HS-1** — milestone boundary operator gate; no terminal re-audit and no merge without approval.
- **HS-2** — if any fix implies redesign outside its feature boundary (e.g. F8.3 forces a taxonomy
  schema change, or F8.2 forces provider-interface churn beyond the envelope), stop and report the
  boundary + minimum prerequisite plan; do not symptom-patch.
- **HS-3** — if a prerequisite (build/runnable/contract regen) fails, repair it, rerun, then resume.
- **HS-4** — validator FAIL → open the named fix feature, re-run its lifecycle, re-dispatch validator.
- **HS-5** — **this milestone is the HS-5 remediation for the 4th Contract/API miss.** If the post-M8
  terminal re-audit *still* misses (5th miss), do **not** open M9 by default: STOP and surface to the
  operator — the typed-everywhere + gate-widen approach failed; the choice is the full codegen-first
  rewire or a §8 re-scope.
- **HS-6** — scope drift / off-plan discovery → stop, surface, replan before continuing.
