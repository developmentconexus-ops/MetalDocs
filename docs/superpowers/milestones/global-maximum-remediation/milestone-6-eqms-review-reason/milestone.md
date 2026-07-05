# Milestone 6 — eQMS periodic review/expiry + structured reason-for-change

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M6)
> **Status:** in-progress (validator re-dispatch pending) — F6.1 gate done (🟡 Yellow, 93cd6114). **F6.2 + F6.3 implemented** (T1–T8b, `b8a32144`…`7b3f0f82`), live QA drive GREEN. **milestone-validator run 1 (fc3057d4): FAIL** — 2 blocking findings on the F6.2 surfacer (§4 silent divergence: unseeded all-tenant sweep + test asserting the inverse of §4.3; `review_surfaced_at` marker unconsumed) + 1 non-blocking (triple-authored due predicate). **HS-4 fix-feature `f6.4-surfacer-contract-and-consumer` executed** (`7d398d92`,`400714ab`,`bf9eadaf`,`d91d3bb8`): surfacer **conformed** to §4.2/§4.3 (per-tenant `SeedTxTenant` + explicit `tenant_id` predicate, correct-by-construction isolation); review-due filter now **reads the surfaced marker** (worklist) + DTO exposure; due-core predicate **single-sourced**; and **the full authored-not-executed integration suite was executed on real Postgres — every M6 DB proof `--- PASS`** (surfacer isolation/idempotency, mark-reviewed OCC/isolation, CHECK, tripwire-negative P0001, reason-persist/reason-on-audit, schedule-publish effective/review fields). **milestone-validator re-dispatch pending.** Not pushed. M7 only after validator PASS + operator HS-1 approval, fresh session.
> **Authored:** 2026-07-04 — *before any implementation feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates the
> milestone against *this* document. The full expected behaviors are pinned in
> `validation-contract.md` (D4), authored and committed **before** the first
> implementation feature (F6.2) begins.
>
> **D7 pre-design gate (done before this spec):** the `developing-new-work` gate returned
> **🟡 Yellow** (`../../../analysis/2026-07-04-m6-eqms-review-reason-system-impact.md`,
> committed 93cd6114). HS-8 (Red ⇒ design blocked) did **not** fire. Per mission D7, M6
> requires the gate but **not** an up-front ADR (only M5/M7 do) — however the
> capability-registry bump convention (`model_test.go:96`, *"bump only via ADR"*) makes an
> ADR required *in practice* for the new `document.review` capability; it is authored **with**
> F6.2, not before the milestone.

## Objective

**After this milestone, MetalDocs enforces the two ISO-core eQMS document-control obligations
the 2026-07-03 review flagged as missing (finding 14): periodic review/expiry, and structured
reason-for-change.** Concretely, observable at close:

1. A governed document revision carries an **effective date, an expiry date, a review-due
   date, and a last-reviewed date** (`effective_from`/`effective_to` **reused** — the columns
   already exist on `public.documents` but were semantically unwired; only `review_due_at` /
   `last_reviewed_at` are genuinely new). Expiry > effective and review-due sanity are enforced
   by **DB CHECK constraints**, not app code alone.
2. Documents whose `review_due_at ≤ now()` are **surfaced by a River periodic job** on the
   leader-elected `metaldocs-jobs` binary (the M5 consolidated base), reading through a **new
   documents published read-port** (`jobs → documents`, never raw SQL) — idempotent,
   tenant-seeded (M3 backstop). No hand-rolled scheduler.
3. A **capability-gated mark-reviewed workflow** exists: a new `document.review` capability
   (registry 34→35, ADR-recorded), tier-1 route→cap mapped, tier-2 `authz.Require` in-tx, the
   documents UPDATE **tripwire arm generated from the registry via M2** (never hand-typed), and
   the transition routed through the **M4 unified transition function** (no scattered `if
   status !=` checks). Reasoned in capabilities, never roles.
4. Revision creation captures a **structured reason-for-change** — a contract field (+ optional
   category enum), not the free-text `revision_title` — at `SubmitRevisionForReview`, carried
   into the **audit trail** via the published `audit.Writer.RecordTx` port **in the business
   tx** (21 CFR Part 11 attributable change reason). Required at the API for REV≥1 (friendly
   first line), nullable in DB for legacy rows (expand/contract).
5. **Contract-first throughout** — every new route/field lands in `api/openapi/v1/openapi.yaml`
   first, then BE (`gen.go`) + FE types regenerate; zero hand-edits to generated code. A named
   FE/consumer consumes each new shape.

**Quality bar moved:** the 2026-07-03 review's dimension-6 product gap ("no effective-date
distinct from publish-date, no periodic review/expiry, no structured reason-for-change —
free-text `revision_title` only") is closed. Re-measured by: contract + generated code + FE
consumer present; a scheduled-surfacing River-job proof; a **live drive** of a due-review cycle
(set review-due → surfacer flags → capability-gated mark-reviewed) and a structured
reason-for-change capture at revision submit showing on the audit trail.

Coherent slice: exactly the two ISO-core eQMS product gaps of finding 14, no more. Training
acknowledgment / obligated-reader tracking is **out of scope by decision** (distribution owns
it). It follows the async-consolidation milestone (M5), whose River periodic-job base F6.2
rides, and precedes tenant-lifecycle (M7).

## Appetite

- **Appetite:** the 3 features below (F6.1 gate already done). Two product features on existing
  modules — no new bounded-context module. One new capability, one ADR, one forward migration
  (expand-only, all-nullable columns), contract-first route/field additions.
- **Rabbit holes (do not chase):**
  - **Training acknowledgment / obligated-reader read-tracking** — `distribution` module owns
    it; finding 14 explicitly scopes it out ("training acknowledgment legitimately
    out-of-module"). Reason: wrong module, out of finding scope.
  - **Adding parallel `published_at`-vs-`effective_date` columns** — `effective_from`/
    `effective_to` already exist on `public.documents`; the global-maximum move is to **wire the
    existing columns**, not duplicate them. Reason: would create a redundant column family.
  - **A bespoke review scheduler / cron in the app** — F6.2 surfacing is a River periodic job on
    the M5 base (`maintenance.PeriodicJobs()` pattern). Reason: re-introducing a hand-rolled
    scheduler is exactly what M5 retired.
  - **Reworking `controlleddocuments` lifecycle** — CD identity/lifecycle (`active|obsolete|
    superseded`) is untouched; review/expiry/reason attach to the **published Document
    revision**. Reason: not CD-owned; scope creep into another module.
  - **Free-text-into-structured migration of historical `revision_title`** — F6.3 captures
    reason-for-change going forward (REV≥1); legacy rows stay NULL. Reason: backfilling intent
    onto historical rows is fabrication, not data.
  - **Any push to origin.** Never, this milestone (mission §2, §10).

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F6.1 | `f6.1-gate` | *(done before this spec — D7)* `developing-new-work` system-impact analysis for the two eQMS product gaps: owning modules (documents + documents/approval + jobs + iam; controlleddocuments explicitly NOT owning), invariant walk, capability-wiring walk for `document.review`, locked-constraints list. **Consumer:** the milestone planner + the M6 validator consume the gate verdict + the 9 locked constraints as the design rails. | Gate artifact committed (93cd6114), verdict **Green/Yellow** (is Yellow); no unresolved AS-1/AS-2/AS-3; the 9 locked constraints carried verbatim into this spec + `validation-contract.md`. Committed before any F6.2 implementation. |
| F6.2 | `f6.2-periodic-review` | Review/expiry model + scheduled surfacing + capability-gated review workflow. **Migration:** `+ review_due_at`, `+ last_reviewed_at` timestamptz NULL on `public.documents`; **reuse** existing `effective_from`/`effective_to` as effective/expiry; CHECK(s): `effective_to > effective_from`, review-due sanity. **Capability:** new `document.review` (all 10 touchpoints; registry **34→35**; tripwire arm **generated via M2**, never hand-typed) + **one ADR** (review model + effective-vs-publish semantics + capability + surfacing-via-River). **Surfacer:** a River periodic job in `jobs` reading due docs via a **new documents published read-port** (`ReviewDueReader.ListDueForReview`), idempotent, tenant-seeded (M3). **Workflow:** a contract-first mark-reviewed route routed through the **M4 unified transition function**, tier-1+tier-2 authz, sets `last_reviewed_at` + next `review_due_at`. **Contract:** review-due/effective/expiry fields on document response DTO(s) + a review-due list filter + the mark-reviewed op in `openapi.yaml`; regen BE+FE. **Consumer:** the FE documents view consumes the new response fields + review-due filter; the surfacer job consumes the read-port. | Contract + generated code + FE consumer present (`oasdiff` M1 green; pin tests for new shapes). Migration applies; CHECKs reject bad rows (expiry≤effective; bad review-due) at the **DB**, not just app. Capability reachable **only** with `document.review`; tripwire fires without the arm/assert (negative proof); **registry size 35**; M2 drift check green; ADR Accepted + indexed. Surfacer integration proof (testdb): a `review_due_at ≤ now()` doc is surfaced on a scheduled tick, **idempotent** (twice → once), tenant-isolated (cross-tenant → 0 rows). mark-reviewed rejected from illegal status by the DB guard + M4 transition fn. All matches `validation-contract.md` §2–§4. |
| F6.3 | `f6.3-reason-for-change` | Structured reason-for-change at revision creation. **Contract:** `reason_for_change` (+ optional `reason_category` enum) on the **submit-revision** request schema in `openapi.yaml`; regen BE+FE. **Migration:** `+ reason_for_change text NULL` (+ optional `reason_category text NULL` with CHECK enum) on `public.documents` (expand-only). **Capture:** field set at `SubmitService.SubmitRevisionForReview` (`submit_service.go:43` `SubmitRequest`) — **not** free-text `revision_title`; written into the **audit trail** via `audit.Writer.RecordTx` **in the business tx** (payload JSON carries `reason_for_change` + category). Required at API for REV≥1 (friendly first line), nullable in DB for legacy. **Consumer:** the FE revision-submit form consumes the new field; the audit reader observes the reason-carrying event. | Contract + pin tests for the new request field(s) (`oasdiff` green; generated `SubmitRequest` carries the field). Revision-creation drive (testdb + live) shows **structured capture**: submitting a revision with `reason_for_change` persists it on `public.documents` **and** emits one audit event whose payload carries the reason (+ category). REV≥1 submit **without** the field is rejected at the API (friendly 422/problem+json); legacy NULL rows tolerated. No reuse of `revision_title` for the structured reason. Matches `validation-contract.md` §5. |
| F6.4 *(HS-4 fix-feature)* | `f6.4-surfacer-contract-and-consumer` | *(opened by milestone-validator FAIL, run 1)* Make the shipped F6.2 surfacer **conform** to binding contract §4 and give its side effect a consumer — no contract re-open (conform was feasible). **D1:** surfacer does a cross-tenant system read then **per-tenant** `SeedTxTenant` + tenant-scoped `MarkSurfaced` (§4.2/§4.3), with an **explicit `tenant_id` predicate** on both review-due queries (RLS backstop; correct-by-construction isolation, provable under the BYPASSRLS dev role). **D2:** the `review_due=true` filter **reads** `review_surfaced_at` (worklist) + the field is exposed on the document DTO; mark-reviewed auto-expels. **D3:** the "due-core" predicate is single-sourced. **Gate#7:** execute the authored-not-executed integration suite on real Postgres. | §4.3 cross-tenant **isolation** proof (seed A → B untouched) green — matches the contract, not its inverse; idempotency green under the per-tenant model; `review_due=true` returns the surfaced set + mark-reviewed expels; `review_surfaced_at` on the DTO (pin); one `dueCorePredicate` const referenced by all 3 sites; **the full M6 DB suite executed on real Postgres, every test `--- PASS`**; build/vet/registry/tripwire/api-lint (no new violation) green. Matches `validation-contract.md` §4 + §6(d) + `qa/milestone-qa.md` C7 fix-reqs 1–4. |

For each feature, "what to validate" is objectively checkable: a named integration test that
passes, a contract-diff that is clean, a negative authz proof that the tripwire fires, a DB
CHECK that rejects a bad row, an audit event whose payload carries the field. No "works" /
"looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it
judges and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), per
the binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`.
For M6:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored. F6.2 matches `validation-contract.md`
   §2–§4 (review/expiry model + CHECKs; capability 10-touchpoint + registry 35 + M2-generated
   tripwire arm; River surfacer idempotency + tenant isolation; mark-reviewed via M4 transition
   fn); F6.3 matches §5 (structured field, audit-trail capture via `RecordTx`, REV≥1 required /
   legacy nullable).
2. **Workflow-class QA checklist** — `wiki/quality/qa-operating-system.md` contract + authz +
   DB-invariant + async/idempotency + multi-tenant-isolation subsets; `wiki/quality/
   test-discipline.md` (testdb factory for every proof; targeted `-run` only — the full
   integration suite is NOT run locally, 20-min box).
3. **Regression** — M0–M5 gates still pass; the transactional-outbox invariant, the M3 RLS
   backstop (per-message seed in the surfacer), the M2 write-tripwire (now including
   `document.review`), and the M4 unified transition function are not regressed; `go build
   ./...` green; **all 4 binaries build**; `.\scripts\check-system-runnable.ps1` green.
4. **Quality-bar / root-cause check** — review/expiry **reuses** `effective_from`/`effective_to`
   (no duplicate column family); surfacing is a **River periodic job** (no hand-rolled
   scheduler); the capability tripwire arm is **generated via M2** (not hand-typed); the
   mark-reviewed transition goes through the **M4 unified function** (not a scattered guard);
   reason-for-change is a **structured field on the audit trail** (not free-text
   `revision_title`). Each is the root fix, not a symptom patch.
5. **No unplanned scope** — anything beyond F6.1–F6.3 recorded with rationale; the rabbit-hole
   list above is the scope-drift baseline. Training acknowledgment stays out (distribution).
6. **Live QA drive (runtime-visible milestone, mission §7 M6)** — `.\scripts\start-api.ps1
   -Build`; drive a **due-review cycle** (set `review_due_at` → surfacer flags the doc →
   capability-gated mark-reviewed sets `last_reviewed_at` + next due) **and** a **structured
   reason-for-change capture** at revision submit; capture proof (network/logs/DB row + audit
   event). Evidence in the closing feature's `evidence.md`.

**HS-7 (mission-specific):** the implementation is compared **section-by-section** against the
committed `validation-contract.md`. Any drift = stop; fix the code to the contract, or re-open
the contract **with operator approval** — never silently edit the contract to match the code.
This binds hardest on the **contract-first** clause: any forced change to an *existing* generated
shape stops and surfaces (never a silent hand-edit of generated code or the spec's meaning).

## Dependencies & constraints

- **Depends on:** M0–M5 (committed; M5 passed operator HS-1 2026-07-04). Specifically: **M2**
  (tripwire arms generated from the capability registry — the `document.review` arm rides it,
  not a hand-edit), **M3** (TxRunner GUC auto-seed + async backstop — the surfacer seeds
  per-message identity), **M4** (unified 9-status transition function — mark-reviewed routes
  through it), **M5** (River periodic-job base — the surfacer is one). River v0.37.1 deployed
  (`metaldocs-jobs`). The `audit.Writer.RecordTx` published port (stable).
- **Quality goals (ranked):** **contract-truth > invariant-preservation > reuse-over-addition.**
  (1) Every route/field is spec-first and the generated code + a named consumer match it — no
  producer-before-contract. (2) All 6 non-negotiables hold (authz=capabilities, contract-first,
  tenancy, async-outbox, DB-enforces, cross-module-via-port) — proven, not asserted. (3) Wire
  existing `effective_from`/`effective_to` + M2/M4/M5 machinery rather than adding parallel
  structures. The validator uses this rank on any trade-off.
- **Architectural constraints (hard rules the validator can fail on):**
  - **AuthZ = capabilities, never roles** — `document.review` walked through all 10 touchpoints;
    **registry bumped 34→35 with an ADR**; tripwire arm **generated via M2**, never a
    hand-edited `TEXT[]`; a negative proof shows the tripwire fires without the arm/assert.
  - **Contract-first** — no route/field exists in Go that isn't in `openapi.yaml`; regen only;
    zero hand-edits to generated. **HS-7** on any forced change to an existing generated shape.
  - **Multi-tenant pooled** — new columns on the already-tenant-scoped `documents` table; the
    surfacer seeds identity in the jobs binary (M3 backstop); cross-tenant read → 0 rows / 404.
  - **Async = transactional outbox** — F6.2 surfacing is a River periodic job; any side effect
    (notification) enqueues via the outbox, never an inline network call in the job tx.
  - **DB enforces invariants** — `effective_to > effective_from` + review-due sanity as CHECK
    constraints; mark-reviewed illegal-status rejected by a DB guard consistent with the M4
    transition fn, not app code alone.
  - **Cross-module via published interface only** — `jobs → documents` review-due read through a
    new documents published read-port; `documents/approval → audit` via `Writer.RecordTx`. No
    module reaches into another's tables.
  - **Reuse existing `effective_from`/`effective_to`** — runtime-verify the current publish-path
    wiring of `effective_from` before design; do not add duplicate published/effective columns.
  - **Route mark-reviewed through the M4 unified transition function** — no scattered lifecycle
    `if status !=` checks.
  - **H-PRE-1 (LIVE)** — the mark-reviewed `authz.Require` recording read must not sit inside a
    lock-holding atomic tx; the path takes no advisory lock (satisfied by construction — keep it
    off any lock).
- **Risks (named, with disposition):**
  - *Contract drift on an existing generated shape* (adding fields to the document DTO / submit
    request forces a breaking change elsewhere): mitigation — `oasdiff` (M1) gate + HS-7 stop;
    all additions are additive/optional at the wire (new fields nullable, reason required only at
    the handler for REV≥1). Surface, never silently hand-edit.
  - *Capability-registry bump without ADR* (guard `model_test.go:96` fails, or the bump is done
    as a symptom patch): mitigation — the ADR is a **gated deliverable of F6.2**; the M2 drift
    check + registry-size test are the backstop.
  - *Surfacer not idempotent* (a due doc surfaced twice on overlapping ticks): mitigation — the
    idempotency proof is a **gate** (run twice → surface once); design the read-port +
    enqueue/flag to be idempotent by construction (River elector single-runner + `ON CONFLICT`
    on any enqueue).
  - *Reason-for-change captured as free-text* (regressing into `revision_title`): mitigation —
    F6.3 acceptance explicitly forbids reusing `revision_title`; the structured field + audit
    payload assertion is the gate.
  - *Full integration suite can't run on the box* (20-min): mitigation — targeted `-run` only;
    bounded defer recorded with a trigger (run on CI/capable box before program close), per
    M1–M5 precedent.
- **Test discipline:** testdb factory for every proof (real Postgres for the surfacer + DB CHECK
  + audit-event proofs, not sqlmock); targeted `-run` filters; full suite NOT run locally;
  bounded defers in `evidence.md` with triggers.
- **Model policy:** implement via subagents (`superpowers:subagent-driven-development`; sonnet
  implement/review, haiku mechanical, never fable, ≤15 concurrent); main session
  orchestrates/reviews/commits.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored
  (never force-add).**

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M7, no merge/push without approval. |
| HS-2 | A fix implies redesign outside M6's boundary — e.g. wiring `effective_from` to a distinct effective date forces a change to the M4 publish-race transition contract, or the review workflow needs a cross-module change to `controlleddocuments` lifecycle. Stop, report the boundary + minimum prerequisite, no symptom-patch. |
| HS-3 | A prerequisite boundary fails (build / all-4-binaries / system-runnable / migration apply / contract regen alignment) — repair first, rerun, resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery mid-milestone (e.g. review/expiry turns out to belong on the CD slot not the revision, or a second reason-capture path is found) — stop, surface, replan before continuing. |
| HS-7 | Implementation deviates from the committed `validation-contract.md` — fix code to contract, or re-open contract with operator approval; never silently adjust the contract. Binds hardest on contract-first (no silent generated-code / spec edits). |
| HS-8 | *(already cleared)* the M6 `developing-new-work` gate returned Yellow, not Red — design was not blocked. Recorded for the audit trail. |
