# Milestone 3 — Category C: published active-membership view + consumption

> **Program:** backend-module-boundary-hardening  ·  **Governing spec:** `../mission.md` (§5 rows 12–14, §7 M3)
> **Status:** Validator **PASS** 2026-06-21 (`qa/milestone-qa.md`) — **HS-1 operator gate pending**. All 3 features closed: F3.1 (view `fe181f34`), F3.2 (CD C1+C2 `764a9d08`), F3.3 (approval C3 `c1b654d6`). Not merged, not pushed. Awaiting operator approval before M4.
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*
> Test-PG DSN: `postgres://metaldocs:metaldocs@localhost:5434/metaldocs?sslmode=disable` (ephemeral; never `.env`).
> Integration parity needs this live + `-tags integration`; if down, mark the step **not-run (HS-3)** — never false-green.

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no execution steps**
> — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone QA
> (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Eliminate the **3 Category-C authz-visibility membership reads** (census C1, C2, C3) where a non-owner module
reaches into iam's `user_process_areas` to evaluate active-now area membership inside a set-based SQL
predicate. After this milestone, iam publishes a **purpose-built, versioned active-membership view**
`metaldocs.v_active_user_areas` that encodes **exactly** the Model-A active-now predicate `effective_to IS
NULL` (ADR 0037 D1), and the three consumers JOIN/read that published view instead of the base table:

- **C1** `controlleddocuments/infrastructure/repository.go:150` — `ListControlledDocuments` restricted-visibility EXISTS leg.
- **C2** `controlleddocuments/infrastructure/repository.go:492` — `CanRead` restricted-visibility EXISTS leg.
- **C3** `documents/approval/repository/postgres_approval_repository.go:1136` — `ResolveEligibleActors`, read **inside the caller's lock-holding tx** (H-PRE-1 site).

**Why a view, not a Go read-port (mechanism C-α, ADR-0039 D3(a)):** these are *set-based* `EXISTS` /
`SELECT … WHERE` predicates living **inside** list / visibility / routing SQL. A per-row Go membership call
would be an N+1 regression and, for C3, would brush the H-PRE-1 advisory-lock hazard. A published view keeps
the read set-based and `SELECT`-only, so it is **tx-structure-neutral** — only the object name changes (base
table → view), the caller's tx/lock structure is untouched (ADR-0039 D5).

**Bar moved:** the Category-C class → **0 remaining** `user_process_areas` entries in the `hgcrossmodule`
guard's `hgPendingRemediation` ledger. After M3 the only ledger entries left are the **C4 / search** rows
(M4). `go run ./tools/cilint ./...` stays exit 0 throughout.

This is a **seam** change, **not** a logic change (mission §2 Non-Goals; D6 parity). **No behavior, visibility,
or authz semantics change.** This is the highest-risk milestone for accidental authz drift — every ported
site's parity test must prove the published-view membership set is **exactly** the set the old raw join
returned, on a real Postgres, **before** the raw SQL is deleted.

### Critical contract fact discovered consumer-first (informs F3.1's view shape)

The view shape was read **from the three consumers**, not invented (mission §7 names a working title only):

- **Columns.** C1/C2 correlate on `(tenant_id, user_id, area_code)`. **C3 additionally filters by `role`** and
  projects `user_id`. The view therefore exposes **`(tenant_id, user_id, area_code, role)`** — `role` is
  required by C3. The temporal columns (`effective_from`, `effective_to`) are **deliberately not exposed**:
  encoding active-now *is the view's contract*, so no consumer should touch the raw interval again.
- **"Active" predicate divergence (load-bearing).** C1/C2 already gate on `effective_to IS NULL` (canonical
  Model-A, ADR 0037 D1). **C3 currently uses the Model-B *interval* form** `effective_from <= now() AND
  (effective_to IS NULL OR effective_to > now())`. Under MetalDocs' Model-A schema (ADR 0037: write path
  always inserts `effective_from = now()`, `effective_to = NULL`; revoke stamps a *past* `effective_to`;
  `effective_to > now()` is the empty set; the partial unique indexes *define* active = `effective_to IS
  NULL`), the two predicates select the **identical set**. Repointing C3 at the `effective_to IS NULL` view
  therefore (a) preserves its result set exactly — proven by parity, including a seeded *revoked* row both
  forms exclude — and (b) incidentally realigns C3 from the ADR-0037-discouraged Model-B leak to the
  canonical predicate. This is **not** an unjustified behavior change; it is the seam change with a parity
  proof that the sets coincide. F3.1 records this in its spec; F3.3's parity gate is the lock.
- **`metaldocs.user_process_areas` is already a view — but NOT a published contract.** The object the
  consumers read today (`metaldocs.user_process_areas`, baseline `0001_current_schema.sql:1653`) is a **1:1
  passthrough** schema-mirror over `public.user_process_areas` (all columns, no active predicate), part of the
  metaldocs schema-exposure layer — not a deliberate, versioned read contract. The `hgcrossmodule` guard
  rightly treats reads of it as base-table reads (it strips the `metaldocs.`/`public.` prefix and maps the
  bare `user_process_areas` token). M3's `v_active_user_areas` is the **purpose-built published contract**
  ADR-0039 D3(a) means; the existing passthrough view is left untouched (other RLS/exposure code depends on it).

## Features

Order is intentional: **F3.1 publishes the view first** (the producer the two consumer features depend on);
F3.2 (CD: two off-tx sites) and F3.3 (approval: one in-tx, H-PRE-1) are then mutually independent consumers.
Each consumer feature is parity-before-delete per site.

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-membership-view` | **iam** publishes `metaldocs.v_active_user_areas` via a forward migration in `db/migrations/`, projecting `(tenant_id, user_id, area_code, role)` from `public.user_process_areas` **WHERE `effective_to IS NULL`** — encoding exactly the ADR 0037 D1 active-now predicate, no interval reinterpretation. ADR-0039 D3(a)/D4 already names this view as the iam published contract; F3.1 records the view as the durable artifact (extend/annotate ADR-0039, no new ADR unless a durable decision emerges). No consumer is repointed in F3.1. | Migration applies cleanly on the test PG (full `db/baseline` + `db/migrations` bootstrap green via the `testdb` harness); a **view-vs-base parity query** proves `SELECT tenant_id,user_id,area_code,role FROM metaldocs.v_active_user_areas` == `… FROM public.user_process_areas WHERE effective_to IS NULL` over a seeded set incl. an active row, a revoked (past `effective_to`) row, and multiple roles; `go build ./...` green; cilint guard unchanged (no ledger edit in F3.1 — the view name is unknown to the guard, hence compliant). |
| F3.2 | `f3.2-cd-consume-view` | **CD** repoints the two restricted-visibility membership EXISTS legs — `repository.go:150` (`ListControlledDocuments`) and `:492` (`CanRead`) — to JOIN/read `metaldocs.v_active_user_areas` instead of `user_process_areas`, dropping the now-redundant `effective_to IS NULL` clause (the view encodes it). Set-based SQL preserved (no per-row Go loop). The CD-owned grant-table legs (`controlled_document_area_grants`, `controlled_document_user_grants`) are **unchanged** — only the foreign membership leg moves. | Per-site **parity test** (raw-join result == view-join result) for **list** (`ListControlledDocuments` across company / restricted-with-area-grant / restricted-with-user-grant / owner / no-access scopes) and **CanRead** (same scopes, true/false) green **BEFORE** the raw SQL is deleted; `go build`/`go test ./...` green; `go run ./tools/cilint ./...` exit 0 with the `controlleddocuments/infrastructure/repository.go` × `user_process_areas` entry **removed** from `hgPendingRemediation`; `git grep` shows 0 `user_process_areas` reads remain under `controlleddocuments/`. |
| F3.3 | `f3.3-approval-consume-view` | **approval** repoints `postgres_approval_repository.go:1136` `ResolveEligibleActors(ctx, tx, …)` to read `metaldocs.v_active_user_areas` **inside the existing caller-supplied tx**, filtering `WHERE tenant_id=$1 AND area_code=$2 AND role=$3` — dropping the Model-B interval predicate (the view encodes active-now). The read stays a **plain, non-recording `SELECT` on the caller's `tx`** (H-PRE-1 / ADR-0039 D5): no authz-recording call is added inside the lock-holding tx; tx/lock structure is byte-for-byte unchanged except the object name. | Per-site **parity test** (raw `ResolveEligibleActors` result == view-based result) across present/absent/multi-role/revoked-row/wrong-tenant cases, executed **on a `*sql.Tx`**, green **BEFORE** deletion; `go build`/`go test ./...` (+ integration) green; `go run ./tools/cilint ./...` exit 0 with the `documents/approval/repository/postgres_approval_repository.go` × `user_process_areas` entry **removed** from `hgPendingRemediation`; `git grep` shows 0 `user_process_areas` reads remain under `documents/approval/`; reviewer confirms **no authz-recording read added inside the lock-holding tx** and the read still runs on the caller's `tx`. |

**Ledger-drain checklist item (applies to F3.2 and F3.3 — every feature that drains a ledger row):** after
removing an `hgPendingRemediation` entry, **run `go test ./tools/cilint/...`** and re-green it. The cilint
unit suite's `TestHGCrossModule_Negative_PendingBaseline` fixture currently points at **C1+C2**
(`controlleddocuments/infrastructure/repository.go` × `user_process_areas`) as its "still-pending" example;
when F3.2 drains that row the fixture MUST be **realigned to a still-pending entry** (e.g. a remaining C4
`search/infrastructure/v2documents/reader.go` row) and the suite re-greened **in the same feature**. M2
shipped an undisclosed cilint-suite FAIL by skipping exactly this step — it is a binding per-feature gate
here, audited at the milestone gate (C3 regression).

For each feature, "what to validate" is objectively checkable: a named parity test green before a named raw
read is deleted, a clean build/test, a guard exit code + ledger delta, a `git grep` returning 0, and (cilint)
a green unit suite.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and writes
`qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding C1–C7 checklist
in `.claude/skills/milestone/references/milestone-end-validation.md`. For M3 it enforces:

1. **Per-feature acceptance** — F3.1–F3.3 each meet every cell of their "what to validate", and each
   consumer feature's **consumer contract** (`spec.md`) was honored: the view shape was read **from the
   consumer call sites** (the `role` column exists because C3 needs it; the active predicate is exactly
   `effective_to IS NULL`), the view is **iam-owned** (published in a `db/migrations/` migration), and each
   consumer reads the **view** — never `user_process_areas`.
2. **Workflow-class QA checklist** — persistence / migration (`wiki/database/index.md`, `wiki/quality/*`
   persistence checklist: forward-only migration, view DDL, idempotency where applicable) + module-boundaries
   (DDD ownership: iam publishes, consumers read the contract) + **authz-invariant (H-PRE-1 / ADR 0022)** for
   C3. Test-framework discipline for the new parity tests (canonical `testdb` fixture framework, ADR 0034).
3. **Regression** — M0, M1, M2 still pass their gates: ADR-0039 intact; M1 constants intact; M2's read-ports
   intact and the 9 B/N1 entries still drained; the `hgPendingRemediation` ledger contains **only** the C4 /
   search rows after M3 (both `user_process_areas` C-rows removed, no C4 row touched); `go run ./tools/cilint
   ./...` exit 0 on the full tree throughout; **`go test ./tools/cilint/...` green** (the
   `TestHGCrossModule_Negative_PendingBaseline` realign landed).
4. **Quality-bar / root-cause check** — the Category-C class is re-measured: `git grep` of
   `user_process_areas` under `controlleddocuments/` and `documents/approval/` returns **0** raw reads, and
   each was replaced by a read of the **iam-published view** (root cause = consumer re-deriving iam's
   membership predicate against iam's base table), **not** by a consumer-local copy of the membership query or
   a parallel passthrough view. The C3 predicate change is justified-and-parity-proven (set-equality under
   Model A), not a silent visibility change.
5. **No unplanned scope** — only the 3 named membership reads move; the CD-owned grant-table legs and the
   surrounding visibility/list/routing Go shape are untouched; the C4 / search reads are untouched (M4); the
   existing `metaldocs.user_process_areas` passthrough view is untouched; no schema change beyond the one
   `v_active_user_areas` view migration. Any other touched file is drift, recorded with rationale.
6. **Parity-before-delete (D6) audit** — for every deleted raw read, its parity test exists, asserts
   raw-path == view-path equality (set-equality for the membership predicate), and the feature's
   `evidence.md` shows it **green before** the deletion commit (or, if the test PG :5434 was unavailable, the
   integration parity step is marked **not-run / HS-3** explicitly — never false-green).
7. **H-PRE-1 / authz-invariant (C3)** — the validator confirms `ResolveEligibleActors` still reads on the
   caller's `tx`, the read is a plain non-recording `SELECT` (no `authz_access_log` / governance write added
   inside the lock-holding tx), and the tx/lock structure is unchanged except the object name.

## Dependencies & constraints

- **Depends on:** M0 (ADR-0039 D3(a)/D4/D5 define the published-view mechanism + the `v_active_user_areas`
  contract + the H-PRE-1 rule; the `hgcrossmodule` guard + ledger this milestone drains) and ADR 0037 (the
  Model-A active-now predicate the view encodes). Addresses census rows **C1, C2, C3** exactly. M2 read-ports
  are independent (no overlap).
- **Quality goals (ranked):** **1. correctness / parity / no authz drift** (the membership set is *exactly*
  preserved — D6 non-negotiable; this is the authz-visibility milestone) > **2. boundary integrity** (iam is
  the single home for the active-membership contract; consumers read the view) > **3. H-PRE-1 safety**
  (C3 read stays non-recording, tx-structure-neutral) > **4. simplicity** (smallest view that satisfies all
  three consumers; no speculative columns — `role` is the only addition beyond the correlation keys) >
  performance (set-based SQL preserved; one view indirection accepted — the view is over the same indexed
  base table).
- **Architectural constraints (hard rules):**
  - **One migration, one view.** M3 adds exactly `metaldocs.v_active_user_areas` (forward-only migration in
    `db/migrations/`, next sequence number). No other schema change; the existing passthrough view and the
    base table are untouched.
  - **View encodes exactly `effective_to IS NULL`** (ADR 0037 D1 / ADR-0039 D4). **No** interval
    reinterpretation (`effective_to > now()` stays refuted per ADR 0037 D2). As-of / history reads are out of
    scope and keep their parameterized interval form.
  - **Published-view mechanism (ADR-0039 D3(a)):** consumers JOIN/read the view; they name **no** iam base
    table. The view name `v_active_user_areas` is unknown to the `hgcrossmodule` guard's owned-table map, so
    reading it is compliant (not a base-table token) — no guard code change is needed for the *view read*;
    only the **ledger entries for the drained base-table reads** are removed.
  - **Set-based, no N+1:** the membership predicate stays inside the consumer's SQL as a JOIN/EXISTS against
    the view — never a per-row Go membership loop.
  - **HS-PRE-1 (C3, hard):** the in-tx `ResolveEligibleActors` read stays a plain **non-recording `SELECT`**
    on the caller's `tx`. **No** authz-recording read may be placed inside the lock-holding atomic tx. A
    `SELECT`-only view satisfies this for free (object name change only).
  - **Cilint-suite re-green on every drain:** draining a ledger row that the
    `TestHGCrossModule_Negative_PendingBaseline` fixture points at requires realigning that fixture to a
    still-pending row **in the same feature** and re-greening `go test ./tools/cilint/...`.
  - **Test discipline:** new parity tests use the canonical `testdb` fixture framework (ADR 0034) and run
    `-tags integration` against the test PG :5434; if down, mark **not-run (HS-3)**, never false-green.
- **Risks (named):**
  - *Accidental authz / visibility drift* — the whole milestone's hazard. A view that returns a different
    membership set than the old join silently changes who can see / approve documents. **Mitigation:** every
    consumer feature's parity test seeds active + revoked + multi-role + wrong-tenant rows and asserts
    **set-equality** raw-vs-view before any deletion; F3.1's view-vs-base parity proves the view itself.
  - *C3 predicate divergence (Model-B interval vs `effective_to IS NULL`)* — repointing C3 changes its
    written predicate. **Mitigation:** ADR 0037 establishes the sets coincide under Model A; the F3.3 parity
    test proves it empirically incl. a seeded revoked row both forms exclude. Documented in F3.1 + F3.3 specs.
  - *H-PRE-1 regression on C3* — re-running the read off-tx, or adding a recording call, would break the
    advisory-lock invariant. **Mitigation:** the port stays a `tx`-bound non-recording `SELECT`; reviewer +
    validator C7 confirm tx/lock structure unchanged.
  - *Migration not applied in the test harness* — a parity test could false-pass against an old schema.
    **Mitigation:** F3.1 verifies the `testdb` bootstrap applies `db/migrations` (it does — `db.go`
    `ApplyCuratedBootstrap` globs the dir) and the view exists before any consumer parity test runs.
  - *Test PG :5434 down* → integration parity not runnable. **Mitigation:** HS-3 — mark the integration step
    not-run explicitly in evidence; never false-green a skipped parity test.
  - *A consumer turns out to need a cross-module API redesign* (not a contained view-consume). **Mitigation:**
    HS-2 — surface the boundary + minimum prerequisite plan; do not symptom-patch. (Not expected — all three
    are contained membership-leg swaps; the larger search predicate redesign is deliberately M4.)

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | **The M3 boundary itself.** On validator PASS, the main session flips status and **stops** for operator review. No M4 start, no merge, no push, without explicit approval. |
| HS-2 | A consumer repoint turns out to require a cross-module **API redesign** beyond a contained view-consume (a shared-API change, a consumer-side contract reshape, a new write path). **Stop**; surface the boundary + minimum prerequisite plan; do not symptom-patch. (Not expected — C4/search is the redesign-risk item and is M4.) |
| HS-3 | A prerequisite boundary fails: `go build ./...` red, or the test PG :5434 unavailable for an integration parity test. Repair / **note the gap explicitly** (mark the parity step not-run), rerun the checkpoint, resume. **Never false-green a skipped parity test.** |
| HS-4 | The `milestone-validator` returns FAIL (a parity gap, an authz-set drift, an H-PRE-1 violation, an unported C row, a guard/ledger regression, an un-realigned cilint fixture, scope drift). Open the named fix feature; re-run its lifecycle; re-dispatch the validator. |
| HS-6 | A Category-C `user_process_areas` consumer the census missed surfaces mid-milestone and changes M3's shape. **Stop**; surface to the operator; replan before continuing. |
| HS-PRE-1 | A port would place an **authz-recording** read inside the lock-holding atomic tx (the C3 `ResolveEligibleActors` site). **Forbidden.** The view read stays a plain non-recording `SELECT` on the caller's `tx`. |
