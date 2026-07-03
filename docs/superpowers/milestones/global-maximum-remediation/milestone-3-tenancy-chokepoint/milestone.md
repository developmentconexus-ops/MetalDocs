# Milestone 3 — Tenancy enforcement chokepoint

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M3)
> **Status:** **in-progress** (2026-07-03). Authored **before any feature began**. Commits local, not pushed.

> This file is a **spec**, authored up front. It says **what** M3 is, **which features** it contains,
> **what each feature implements**, and **what gets validated**. It contains **no execution steps** —
> the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone QA
> (`qa/milestone-qa.md`) validates the milestone against *this* document and the binding
> `validation-contract.md` (D4). Drift between implementation and that contract is **HS-7**.

## Objective

Convert **multi-tenant isolation** from ~62 hand-written acts of discipline (`authz.SeedTxIdentity`
called manually at every application-service write site) into a **structurally-enforced** property, and
extend the RLS backstop that today protects only the **synchronous API binary** to also cover the
**async worker/jobs fleet** — closing the review's sharpest multi-tenancy finding (Dimension 4): *the
NULL-permissive RLS policy means `metaldocs-worker` and `metaldocs-jobs` run with **zero** RLS backstop;
async isolation rests entirely on ~229 hand-written `tenant_id` predicates, and one bad join in a future
worker query is a silent cross-tenant leak with no gate.*

The bar this milestone advances: review **Dimension 4 (Multi-tenancy)** moves from **DEBT** toward
**CONFIRMED**, and cross-cutting finding #2 ("`SeedTxIdentity` called manually at ~85 sites; forgetting
one is silently absorbed by the NULL-permissive RLS policy") gets the same single-source-of-truth +
gate treatment that already made module boundaries and capability naming incident-free.

Three concrete moves, each contract-locked before implementation (D4):
1. **F3.1** — auto-seed the tenant/actor GUCs at the **TxRunner chokepoint** (`internal/platform/db`)
   from a platform-layer request-identity carrier, so the API binary's every transaction is RLS-backed
   by construction; collapse the manual `SeedTxIdentity` sites to **zero outside an explicit allowlist**.
2. **F3.2** — give the async fleet a **real RLS backstop**: seed the claimed row's tenant GUC in each
   **single-tenant processing transaction** (per ADR 0054 rule 2, already-accepted), so `FORCE ROW LEVEL
   SECURITY` catches a missing/wrong predicate; add a **compensating lint** that guards the boundary
   between tenant-scoped processing and the sanctioned cross-tenant *claim*/system-table paths; prove it
   with a **negative RLS integration test**.
3. **F3.3** — document the deliberate NULL-permissive design + the sync-vs-async posture + the new
   enforcement in **ADR 0027** and the wiki (today it lives only in a migration comment).

Source findings: review §Dimension 4; §Cross-cutting item 2; §Priorities P1.2.

## Discovered runtime truth (recorded before implementation — HS-6 surface, not silent expansion)

Census + tx-boundary investigation while authoring this spec (2026-07-03; full detail in
`validation-contract.md`) established ground truth that **refines the mission's per-feature sizing**.
Recorded here so any expansion is visible to the validator and operator, not absorbed silently:

1. **The TxRunner chokepoint already exists but does not seed.** `internal/platform/db/runner.go`
   (`TxRunner.Do` / `DoReadOnly`) is the single begin/commit chokepoint; its doc comment *references*
   `authz.SeedTxIdentity` but the seeding is still performed manually **inside the callback** at **62
   non-test call sites** (grep census, worktree copies excluded). So F3.1 is "make the existing
   chokepoint seed," not "build a chokepoint."

2. **Identity is already in the request context, but threaded as explicit args.** Application services
   receive `tenantID`/`actorID` as **function parameters** (from command/request payloads) and pass them
   to `SeedTxIdentity`. The authenticated identity originates in ctx (`tenant.FromContext`,
   `iamdomain.UserIDFromContext`) at the HTTP edge. For the chokepoint to seed transparently it must read
   identity **from ctx** — which requires a **platform-layer identity carrier** (`internal/platform/...`)
   holding tenant **and** actor, because `internal/platform/db` may not import `internal/modules/iam`
   (module-boundary guard). `platform/tenant` already carries tenant; it must be extended (or a sibling
   `platform/identity` added) to also carry actor.

3. **Two migration edges make "census = 0" a census-with-allowlist, not a naive zero:**
   - **Actor ≠ ctx-actor sites.** A few sites seed an actor taken from a *stored* value
     (`grantedByActor(membership.GrantedBy)`, `assignedBy`, `d.CreatedBy`) rather than the current ctx
     actor. Where that value is provably the current actor, the site collapses; where it is a distinct
     semantic, the site is **allowlisted** (keeps an explicit seed) — never silently changed.
   - **Cross-tenant API `Do` paths.** Any platform-admin request path (ADR 0021) that legitimately reads
     across tenants inside a request tx must **not** be auto-seeded to a single tenant. Those are
     enumerated and allowlisted (auto-seed skipped / explicit system posture). The acceptance is
     therefore *"grep census = 0 **or** explicit allowlist"* (mission F3.1).

4. **Async processing is single-tenant per tx; only claim/system paths are cross-tenant (ADR 0054).**
   The five business-processing handlers — materialize, pdf, scheduled-publish, notifications-fanout, and
   the staging-outbox per-row processing — each know their tenant from the message payload **before** the
   write and touch exactly **one** tenant's rows per tx. The tenant-unscoped steps are the **claim**
   queries (`FOR UPDATE SKIP LOCKED`, ADR 0054 rule 1), the stuck-watchdog cross-tenant **list**, and
   three **system tables with no `tenant_id` column** (`idempotency_keys`, `job_leases`, and the audit
   scan). So F3.2 seeding **completes ADR 0054 rule 2** ("everything done *with* a claimed row MUST be
   tenant-scoped; tx-local GUCs per item") — it is enforcing an already-accepted decision, **not** a
   redesign (no HS-2).

5. **RLS needs only the tenant GUC; async system work has no human actor.** The `tenant_isolation`
   policy reads `metaldocs.tenant_id` only. `SeedTxIdentity` hard-requires a **non-empty actor** as well
   (for `authz.Require`/audit). Async system paths (scheduler/janitor) legitimately have no actor and use
   `authz.BypassSystem`. Therefore F3.2 needs a **tenant-only seed primitive** (set `metaldocs.tenant_id`
   without demanding an actor) — the full `SeedTxIdentity` is wrong for the async backstop.

6. **Two async handlers run raw multi-statement `ExecContext` with no surrounding tx** (pdf,
   notifications-fanout). `set_config(..., true)` is **transaction-local**; without a tx the GUC does not
   carry to the next statement. Seeding those safely requires **wrapping their writes in a single tx**
   (small, local, also improves atomicity) — session-level GUCs on a pooled connection are forbidden
   (leak to the next borrower). This is in-boundary handler-local work, not a redesign.

None of the above crosses a feature boundary into redesign (no HS-2): F3.1 makes an existing chokepoint
seed from an existing identity source; F3.2 completes ADR 0054 rule 2 with a tenant-only seed + a
boundary lint; F3.3 is documentation. The RLS policy itself is **not weakened** — it stays NULL-permissive
(so GUC-less system scans keep working); the milestone *adds* seeding so the FORCE-RLS backstop actually
engages where a tenant is known.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-txrunner-autoseed` | A **platform-layer request-identity carrier** (tenant + actor in ctx, set by the API authn middleware) that `internal/platform/db.TxRunner` reads to **auto-seed** `metaldocs.tenant_id` + `metaldocs.actor_id` at tx begin (both `Do` and `DoReadOnly`; RO-safe `SET LOCAL`) **when identity is present**, and no-op when absent (system paths, preserving NULL-permissive behavior). **Collapse** the 62 manual `SeedTxIdentity` sites to the chokepoint; any residual site is on an **explicit allowlist** with a recorded reason. A **blocking CI lint** (`scripts/api-lint`, Go AST, same framework as the authz lints) that fails when a `SeedTxIdentity` call appears **outside the chokepoint and outside the allowlist** (census-drift guard). | Grep census of manual `SeedTxIdentity` outside the chokepoint = **0 or allowlisted** (allowlist enumerated with reasons); the census-drift lint is **RED** on a synthetic new manual seed site (captured), **GREEN** on the clean tree; the API binary's transactions are RLS-seeded by construction (shown by a chokepoint unit/integration proof); the existing `tenant_isolation_test.go` suites stay **green**; `go build ./...` green. |
| F3.2 | `f3.2-async-rls-backstop` | A **tenant-only seed primitive** (`authz.SeedTxTenant(ctx, tx, tenantID)` — sets `metaldocs.tenant_id` only, no actor requirement) called at the start of every **single-tenant async processing tx** (materialize, pdf, scheduled-publish, notifications-fanout, staging-outbox per-row processing); pdf + notifications-fanout writes **wrapped in a tx** so the GUC carries. A **blocking compensating lint** (`scripts/api-lint`, Go AST) over the worker/jobs binaries that flags any DB write not inside a tenant-seeded tx **unless** it is on the sanctioned cross-tenant/system allowlist (outbox **claim** steps per ADR 0054, stuck-watchdog list, `idempotency_keys`, `job_leases`, audit scan). The RLS policy is **unchanged** (still NULL-permissive). | **Negative RLS proof (integration):** a worker-style tx that seeds tenant A then issues a write whose predicate targets tenant B's row is **blocked by FORCE RLS** (0 rows / policy denial) — captured, labeled real-DB (testdb factory), and shown **passing (leak) pre-fix / blocked post-fix**. The compensating lint is **RED** on a synthetic worker write outside a seeded tx and off-allowlist (captured), **GREEN** on the clean tree. All async handlers still function (scheduled-publish + fanout targeted drives green). `go build ./...` green. |
| F3.3 | `f3.3-adr0027-amendment` | Amend **ADR 0027** (RLS adoption) with: the NULL-permissive design + its rationale (GUC-less system scans must see all rows); the **sync-vs-async asymmetry** and how M3 closes it (chokepoint autoseed + async tenant-seed + boundary lint); cross-reference **ADR 0054** rule 2 as now-enforced. Refresh the wiki tenancy page(s) with the same posture + the new enforcement surface. No code-behavior change. | ADR 0027 contains a new dated amendment stating the asymmetry, its rationale, and the new enforcement (chokepoint autoseed + async seed + both lints + negative-proof test); the wiki tenancy doc(s) match runtime truth (no stale "async has no backstop" or "seed manually at 85 sites" claims); `wiki-curator` pass clean; ADR/​wiki cross-refs (0027 ↔ 0054 ↔ this milestone) resolve. |

For each feature, "what to validate" is objectively checkable — a gate that fails on a negative fixture
and passes on the clean tree, with captured command output as evidence (positive + negative proof per
D4). F3.2 additionally requires **integration proof** on an RLS-enforced DB (the negative cross-tenant
write) — the only test class that can pin an RLS backstop (application tests are sqlmock and cannot
exercise the policy).

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M3:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", each demonstrated
   **failing-then-passing from clean state** (negative fixture/probe RED → clean tree GREEN) with real
   captured output. F3.2's negative RLS proof is shown **leaking pre-fix → blocked post-fix** on a real
   RLS-enforced DB (labeled real-DB, not sqlmock).
2. **Validation-contract conformance (D4)** — implementation is checked against `validation-contract.md`
   **section by section**, including the **exact expected RLS behavior per binary** (api / worker / jobs),
   the **chokepoint seeding contract**, the **allowlist** contents, and the **negative-proof test shape**.
   Any divergence is **HS-7** (fix code to the contract, or re-open the contract WITH operator approval —
   never silently adjust the contract to match code).
3. **Workflow-class QA** — backend-api multi-tenant class. QA re-runs the deterministic gates from a
   clean tree: `go run ./scripts/api-lint ...` (both new rules blocking, zero live violations),
   `go test ./scripts/api-lint/...`, `go build ./...`, and the **targeted** tenant-isolation +
   negative-RLS integration drives (`go test -tags integration -run 'Tenant|RLS|Isolation' ./...`) —
   **not** the full 20-min suite (mission §10).
4. **Regression** — M0/M1/M2 gates still pass; the 5 authz CI lints + the M2 tripwire lints stay green;
   `TestCapabilityRegistrySize` unchanged; cross-tenant 404 suite green; no route/contract shape regresses.
5. **Root-cause check** — the manual-seed discipline class is **structurally** closed: seeding is now a
   property of the chokepoint (api) + a per-message primitive (async), guarded by two blocking lints, not
   62 acts of discipline. The RLS backstop is proven **live** for async (negative cross-tenant write
   blocked), not asserted. The RLS policy is **not weakened** to make anything pass.
6. **No unplanned scope** — anything implemented beyond this list is recorded with rationale. The
   allowlist entries (actor≠ctx sites; cross-tenant API paths; sanctioned async claim/system paths) and
   any coverage limitation of the lints are recorded bounded defers with triggers.

## Dependencies & constraints

- **Depends on:** M0 + M1 + M2 passed and committed. F3.1 operates on the current `TxRunner`
  (`internal/platform/db/runner.go`) and the request-identity carriers (`internal/platform/tenant`,
  `internal/modules/iam/domain/context.go`). F3.2 operates on the current async handlers and the
  `tenant_isolation` RLS policy (`db/baseline/0001_current_schema.sql`, NULL-permissive) + ADR 0054.
- **Multi-tenant invariant is non-negotiable:** every tenant table carries `tenant_id`; **tx-local GUCs
  only** (`set_config(..., true)`); tenant-namespaced blob keys; cross-tenant URL → 404. **Do NOT weaken
  RLS** — the policy stays NULL-permissive; the milestone *adds* seeding, it does not relax the policy.
- **H-PRE-1 advisory-lock constraint holds:** seeding is `SET LOCAL` (a config write, **not** an
  authz-recording read), so it does not put an `authz.Require` read inside a lock-holding atomic tx.
  F3.2 seeds the tenant GUC at tx begin, **before** any `FOR UPDATE`/advisory lock, and adds **no**
  `authz.Require` to any locked path.
- **Module boundaries:** `internal/platform/db` must **not** import `internal/modules/iam` — the identity
  the chokepoint reads comes from a **platform-layer** carrier (guarded by `check-module-boundaries.ps1`).
- **CI-truth:** every gate this milestone adds is **blocking** by construction (api-lint `main.go`: bound
  by CI, not discipline). No reported-only tier.
- **Targeted tests only:** the full integration suite is **not** run (20+ min box, mission §10). F3.2's
  RLS proof is scoped with `-run` filters. If the box cannot run `-tags integration` locally, the drive
  is authored + the block recorded as a bounded defer with the run-trigger (M1/M2 env-risk precedent).
- **Model policy:** sonnet implement/review; haiku mechanical sweeps; **never fable** workers; ≤15
  concurrent. Subagent-driven implementation; the main session orchestrates, reviews between features,
  and commits — it does **not** implement inline.
- **Commit after verified work** (standing auth); **never push**; **never commit `docs/release/`**;
  plans dir is gitignored (never force-add).

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M4 and no push without approval. |
| HS-2 | A fix implies redesign outside its boundary — e.g. the async backstop would require **per-tenant claim loops** (contradicting ADR 0054), or making RLS non-NULL-permissive (a load-bearing policy change breaking system scans), or the chokepoint autoseed would require `platform/db` to import `iam` (boundary violation with no platform carrier). Stop; report; do not patch across the boundary. The platform identity carrier + tenant-only async seed + boundary lint exist precisely to stay inside the boundary. |
| HS-3 | A prerequisite fails from clean state — `go build ./...` red, an M0/M1/M2 gate regressed, or the RLS-enforced test DB not applyable. Repair the prerequisite first; rerun; resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery beyond the recorded runtime-truth edges (actor≠ctx allowlist; cross-tenant API paths; the raw-exec handler wrapping). Stop; surface; replan. |
| HS-7 (mission) | Implementation deviates from `validation-contract.md` (esp. the per-binary RLS-behavior table, the seeding contract, the allowlist, and the negative-proof shape) — fix code to the contract, or re-open the contract WITH operator approval; never silently adjust the contract to match the code. |
