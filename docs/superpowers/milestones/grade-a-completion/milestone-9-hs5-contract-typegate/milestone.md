# Milestone 9 — Contract Type-Gate Closure (HS-5, 5th-miss remediation)

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** in-progress — authored 2026-06-20 (operator approved Option A on the post-M8 5th-miss surface).
> **Authored:** 2026-06-20 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone
> QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Close the **post-M8 terminal re-audit** gap (`wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m8.md`,
**VERDICT: FAIL**, mission-validator corroborated `../qa/mission-validation.md`) and reach the mission §8
Grade-A terminal bar. Post-M8, 9 of 10 dimensions are ≥ A− and H-G = 0; the **single** miss is:

- **Contract / API = B+** (< A−), held below the bar by **3 skeptic-confirmed Majors** in
  `internal/modules/documents/delivery/http/handler.go` — all untyped `map[string]string` response
  literals on public, spec-declared routes.
- **Honest H-D = 3 by §8 intent** — the F8.6 `noresponsemap` analyzer is scoped to `map[string]any`
  only (`isMapStringAnyLiteral` returns false for `map[string]string`), so these 3 literals pass the
  mechanical gate silently.

The bar moved is the §8 four-check terminal acceptance. The criterion that proves it: a fresh post-M9
re-audit + `mission-validator` returns **VERDICT: PASS** — Contract/API ≥ A−, 0 confirmed Majors,
H-D = 0 by intent (analyzer now flags `map[string]<T>` of any value type).

**Root cause this milestone fixes (why a 5th miss happened):** the recurring pattern is that each gate
names *one shape* and the next independent read finds the *adjacent shape* just outside scope — first the
path scope (post-M7: `iam/presence`, `observability`), then the type scope (post-M8: `map[string]string`
in `documents`). M9 fixes the 3 instances **and** widens the analyzer from `map[string]any`-only to any
`map[string]<T>` response literal reaching a 2xx writer, so the evasion **class** is closed mechanically,
not just the instances (operator decision 2026-06-20: option A — targeted fixes + one-type-wider gate).

## Appetite & rabbit holes

- **Appetite:** one focused milestone — 4 bounded features in a single handler file + one analyzer +
  codegen regen + one spec amendment. This is **not** the full codegen-first StrictServerInterface
  migration (operator weighed and declined that in M7/M8 — A− was reached without it; §8 is a
  grade/Major/class bar, not a framework).
- **Rabbit holes (named refusals):**
  - **No** repo-wide StrictServerInterface rewire — only the 3 non-compliant `documents` routes change.
  - **No** new FE feature work beyond the codegen-type regen the F9.3 spec amendment requires.
  - **No** new product scope, **no** migrations, **no** touching the 9 dimensions already ≥ A−.
  - **No** re-litigating the skeptic-downgraded Minors (Security `CreateDocumentTx`, the two
    code-quality nits) — out of scope; they did not gate.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F9.1 | `f9.1-duplicate-typed` | `duplicateDocument` (`handler.go:674`) emits the already-generated `documentsapi.DocumentCreateResult` typed body instead of `map[string]string{document_id,initial_revision_id,session_id}`. | Handler constructs no response `map[string]string`; returns `DocumentCreateResult`; wire JSON keys/values byte-identical to pre-change (test); 201 preserved; build+test green. |
| F9.2 | `f9.2-comments-typed` | Comment list/create/update (`handler.go:1122,1159,1193`) emit the generated `documentsapi.DocumentCommentResponse` instead of the package-local `commentResponse`; reconcile the field shapes (`id` UUID, `content` typed node array, timestamps) to the generated contract. | The three handlers serialize `DocumentCommentResponse`; local `commentResponse` removed (or proven unused); wire shape matches the generated type / FE codegen; contract test passes; build+test green. |
| F9.3 | `f9.3-revision-url-contract` | Resolve the `signedRevisionURL` (`handler.go:1105`) contract: **consumer truth = 200 + `{url}` JSON** (FE `DocumentEditorPage.tsx:95` does `apiFetch<{url?:string}>`; it does NOT follow a redirect). Amend OpenAPI `getDocumentRevisionUrl` from `302/no-body` to `200` with a typed `RevisionUrlResponse {url: string}`; handler emits the generated typed struct; regen BE+FE codegen. | OpenAPI declares 200 + typed body; handler emits the generated type (no `map[string]string`); status 200 preserved (matches live FE); FE codegen regenerated & committed; `wiki/modules/documents.md:258` stale "Aligned" note corrected; build+test green. |
| F9.4 | `f9.4-noresponsemap-widen` | Widen `tools/cilint` `noresponsemap` to flag **any** `map[string]<T>` composite literal (not just `map[string]any`) reaching a 2xx body writer on a registered-route package; update `api-contract.md` §5b to state the widened type scope. | Analyzer flags a `map[string]string` response literal in a unit fixture (test in `tools/cilint/internal/analyzers`); `go run ./tools/cilint ./...` exits 0 at HEAD only **after** F9.1–F9.3 land; §5b documents the widened scope; pre-existing allowlist/exemptions (health probes, metrics dynamic leaves, command-input/recordAudit/domain-mirror maps) still pass. |

For each feature, "what to validate" is objectively checkable: a grep that returns empty, a route that
responds with the contracted shape, a test that passes, a build that is clean.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M9 it enforces:

1. **Per-feature acceptance** — every feature meets its declared "what to validate"; each feature's
   consumer contract (`spec.md`) is honored (producer matches the generated type / FE consumer; wire
   shapes unchanged where claimed, changed exactly as declared for F9.3).
2. **Workflow-class QA** — `wiki/quality/backend-api-checklist.md` (typed-body, error-envelope,
   contract-first regen order for F9.3) + close-out.
3. **Regression** — M0–M8 still pass; full `go build ./...` + `go test -count=1 ./...` green; FE codegen
   drift-clean.
4. **Quality-bar / root-cause check** — the 3 post-M8 Majors confirmed fixed at root (typed bodies, not
   symptom-patched); the `noresponsemap` analyzer demonstrably catches `map[string]string` response
   literals now; Contract/API re-graded ≥ A−; H-D = 0 by intent.
5. **No unplanned scope** — anything beyond F9.1–F9.4 recorded with rationale.

> M9's milestone-validator is the per-milestone gate. The program's **terminal** acceptance is separate:
> after HS-1 approval, the main session re-runs the 10-dimension post-M9 re-audit and dispatches the
> `mission-validator` against §8. Grade-A is the operator's declaration on that PASS.

## Dependencies & constraints

- Depends on: M0–M8 passed (9/10 dims ≥ A−, H-G = 0; widened §5b path-scope gate + `noresponsemap` analyzer).
- Audited HEAD baseline: code `58dea742` (post-M8).
- Constraints respected: **no migrations**; F9.3 follows contract-first regen order (OpenAPI → BE codegen →
  FE codegen); ADR 0012 hand-rolled typed structs remain valid for non-codegen modules (not in scope here —
  all 3 fixes use existing generated `documentsapi` types); platform stays module-import-free.
- F9.3 amends a governing contract (OpenAPI shape) → operator-visible; record the spec change + rationale
  (runtime/consumer truth over the stale 302 declaration).
- F9.4 amends a governing contract (§5b type scope) → operator-visible; record the scope change.

## Applicable hard-stops

- **HS-1** — milestone boundary operator gate; no terminal re-audit and no merge without approval.
- **HS-2** — if any fix implies redesign outside its feature boundary (e.g. F9.2's field reconciliation
  forces a domain-model change, or F9.3 forces an auth/redirect-flow change), stop and report the boundary
  + minimum prerequisite plan; do not symptom-patch.
- **HS-3** — if a prerequisite (build/runnable/codegen regen) fails, repair it, rerun, then resume.
- **HS-4** — validator FAIL → open the named fix feature, re-run its lifecycle, re-dispatch validator.
- **HS-5** — **this milestone is the HS-5 remediation for the 5th Contract/API miss.** If the post-M9
  terminal re-audit *still* misses (6th miss), do **not** open M10 by default: STOP and surface to the
  operator — the targeted-fix + gate-widen approach failed for a 6th time; the choice is the full
  codegen-first StrictServerInterface rewire or a §8 re-scope/accept. Escalate as a hard HS-2 boundary.
- **HS-6** — scope drift / off-plan discovery → stop, surface, replan before continuing.
