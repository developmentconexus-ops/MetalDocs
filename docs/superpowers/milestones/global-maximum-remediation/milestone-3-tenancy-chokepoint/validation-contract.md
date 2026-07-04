# M3 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M3 (tenancy enforcement chokepoint)
> **Authored:** 2026-07-03, **before any implementation** (mission D4). Committed before the first code change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7). The **§4 per-binary RLS-behavior table**, the **§1 chokepoint seeding
> contract**, the **§1.4 allowlist**, and the **§2.5 negative-proof shape** are the load-bearing clauses.
>
> **Erratum 2026-07-03 (HS-7 re-open, operator-approved — feature F3.4).** The milestone-validator (M3 QA)
> found a false runtime-truth premise in §0.3/§2.4/§4: `idempotency_keys` was described as a system table
> "with no `tenant_id` column / RLS structurally N/A." Source (`db/baseline/0001_current_schema.sql:1330,1347`)
> shows it **has** `tenant_id uuid NOT NULL` + FORCE RLS + the `tenant_isolation` policy (it is 1 of the 33
> FORCE tables). Its idempotency-janitor TTL `DELETE` is a **sanctioned cross-tenant system-maintenance
> sweep** run GUC-unset under the NULL-permissive hatch (same class as the audit scan), not a table where
> RLS cannot apply. `job_leases` genuinely has no `tenant_id` (that claim was correct). §0.3, §2.4, §4 were
> corrected in place under this operator-approved re-open; no acceptance bar changed. See
> `f3.4-idempotency-keys-rls-truth/`.

---

## 0. Runtime-truth basis (the census + policy facts this contract is built on)

All claims below are traced to source at authoring time (2026-07-03), not to prior doc comments.

### 0.1 The RLS policy (unchanged by this milestone)

- **33 tenant-scoped tables** carry `ENABLE ROW LEVEL SECURITY` + `FORCE ROW LEVEL SECURITY`
  (`db/baseline/0001_current_schema.sql`; `FORCE` count = 33). `FORCE` means the **table owner is also
  subject** to the policy (the app does not connect as a superuser — the `NOSUPERUSER` deployment
  constraint of ADR 0027 / ADR 0022 Phase 5 applies; a superuser would bypass RLS entirely).
- Every policy is named `tenant_isolation` and has the **identical NULL-permissive shape**, with **no
  explicit `FOR` clause (⇒ `FOR ALL`) and no explicit `WITH CHECK`**:
  ```sql
  CREATE POLICY tenant_isolation ON <t> USING (
    (NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL)
    OR (tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid)
  );
  ```
- **Postgres semantics that make the negative proof exact:** for a `FOR ALL` policy with `WITH CHECK`
  omitted, the `USING` expression is **also used as the `WITH CHECK`**. Therefore, with the GUC seeded to
  tenant **A**:
  - **SELECT / UPDATE / DELETE** of a tenant-**B** row → the row is **invisible** (USING false) → **0 rows
    affected** (a cross-tenant read/mutation silently touches nothing — the leak is prevented).
  - **INSERT** (or `UPDATE ... SET tenant_id`) producing a tenant-**B** row → **WITH CHECK (= USING)
    fails** → hard error **`new row violates row-level security policy for table "<t>"`** (SQLSTATE
    `42501`).
  - **GUC unset / empty** → `USING` = TRUE → all rows visible, all writes allowed (**NULL-permissive**:
    this is the deliberate escape hatch for GUC-less system scans — **must not be removed**).
- **This milestone does NOT alter any policy, `ENABLE`, `FORCE`, or the NULL-permissive shape.** It only
  changes **which transactions seed the `metaldocs.tenant_id` GUC**, so the FORCE-RLS backstop engages
  where a tenant is known. No new migration is required for F3.1/F3.2 (Go-only); F3.3 is docs-only.

### 0.2 The seeding primitives + chokepoint

- `internal/platform/db/runner.go` — `TxRunner` (`Do`, `DoReadOnly`) is the **single** begin/commit
  chokepoint. It does **not** seed today; seeding is manual inside each callback.
- `internal/modules/iam/authz/context.go:48` — `SeedTxIdentity(ctx, tx, tenantID, actorID)` sets **both**
  `metaldocs.tenant_id` and `metaldocs.actor_id` via `set_config(..., true)` (tx-local) and
  **hard-requires both non-empty** (returns `ErrTenantContextMissing` / `ErrActorContextMissing`).
- Request identity in ctx: `internal/platform/tenant/context.go` (`WithTenantID`/`FromContext`, tenant
  only) and `internal/modules/iam/domain/context.go` (`WithAuthContext`/`UserIDFromContext`, actor).
- **Census:** **62** non-test `SeedTxIdentity(` call sites (worktree copies excluded), across
  `documents` (approval + repository + application), `controlleddocuments`, `templates`, `taxonomy`,
  `iam/infrastructure`, `tokens`. This is the set F3.1 collapses.

### 0.3 The async fleet (F3.2 target set)

| Async unit | Tx? | Tenant known before write? | One tenant per tx? | F3.2 disposition |
|---|---|---|---|---|
| materialize (`internal/platform/worker/materialize_job_runner.go`) | yes (`BeginTx`) | yes (payload) | yes | **seed tenant GUC** |
| pdf (`internal/platform/worker/pdf_job_runner.go`) | **no** (raw `ExecContext`) | yes (payload) | yes | **wrap in tx + seed** |
| scheduled-publish (`internal/modules/documents/approval/jobs/scheduled_publish_job.go`) | yes (`runner.Do`) | yes (River args) | yes | **seed tenant GUC** |
| notifications-fanout (`internal/modules/notifications/infrastructure/fanout_worker.go`) | **no** (raw `ExecContext`) | yes (event args) | yes | **wrap in tx + seed** |
| staging-outbox **processing** (`internal/modules/render/fanout/staging_outbox_worker.go`, per claimed row) | per-row | yes (claimed `OutboxRow.TenantID`) | yes (1 row) | **seed tenant GUC in processing tx** |
| **Sanctioned cross-tenant / system (stay GUC-unset — allowlisted):** platform outbox **claim** (`internal/platform/messaging/outbox/postgres/consumer.go`), staging-outbox **claim** (`internal/modules/render/fanout/staging_outbox.go`, ADR 0054 rule 1), stuck-watchdog cross-tenant **list**, `idempotency_keys` janitor (**`tenant_id`-bearing FORCE-RLS table; cross-tenant TTL sweep under the NULL-permissive hatch, same class as the audit scan**), `job_leases` lease-reaper (**genuinely no `tenant_id` column**), audit-integrity scan | — | — (cross-tenant by design) | — | **no seed; on the §2.4 allowlist** |

**ADR 0054 already mandates the F3.2 seam** (rule 2: "everything the consumer does *with* a claimed row
… MUST be scoped to that row's `tenant_id`; tx-local GUCs per item; never mix rows from different tenants
inside one business transaction"). F3.2 **enforces** rule 2; it does not invent it (no HS-2).

---

## 1. F3.1 — chokepoint auto-seed + census-drift lint (the sync/API backstop)

### 1.1 The platform identity carrier (source of truth for the chokepoint)

- A **platform-layer** ctx carrier holds **both** tenant and actor. It is set **once** by the API authn
  middleware (the same place that today sets `tenant.WithTenantID` + `iamdomain.WithAuthContext`).
- **Boundary rule (binding):** `internal/platform/db` MUST read identity **only** from a
  `internal/platform/...` package — it may **not** import `internal/modules/iam` (guarded by
  `check-module-boundaries.ps1`). Implementation may (a) extend `internal/platform/tenant` to also carry
  actor, or (b) add a sibling `internal/platform/identity`. The chosen shape is recorded in `plan.md`; the
  carrier exposes a tenant getter and an actor getter returning `("", false)`-style absence.

### 1.2 The chokepoint seeding contract (exact behavior — binding)

For **both** `TxRunner.Do` and `TxRunner.DoReadOnly`, at transaction begin, **before** invoking the
callback `fn`:

1. Read `(tenantID, ok_t)` and `(actorID, ok_a)` from the platform carrier on `ctx`.
2. **If both present and non-empty** → seed `metaldocs.tenant_id` **and** `metaldocs.actor_id` via
   `set_config(..., true)` (tx-local), exactly as `SeedTxIdentity` does. The seed uses the **same SQL**;
   `SET LOCAL` is RO-safe, so `DoReadOnly` seeds too (RLS then also backstops read paths — defense in
   depth; this is intended and must not regress cross-tenant 404 behavior).
3. **If either is absent** → **no-op** (do not error). The tx runs GUC-unset → NULL-permissive RLS lets
   it through, **identical to today's system-path behavior**. This is what keeps GUC-less janitor/system
   `Do` calls working.
4. The chokepoint seed must occur **before** any `authz.Require` in `fn` (Require reads the GUCs) and
   **before** any `FOR UPDATE`/advisory lock (H-PRE-1: the seed is a `SET LOCAL`, not an authz-recording
   read — this ordering must be preserved; **no** `authz.Require` is added to any locked path by F3.1).

**Positive behavior proof:** a chokepoint-level test shows that after `Do` with an identity-bearing ctx,
`current_setting('metaldocs.tenant_id')` == the ctx tenant inside `fn`, and with an identity-less ctx the
GUC is empty.

### 1.3 Collapsing the 62 manual sites

- Every manual `SeedTxIdentity(ctx, tx, …)` **inside a `TxRunner.Do`/`DoReadOnly` callback** whose seeded
  identity is **provably the ctx identity** is **removed** (the chokepoint now seeds it).
- **Behavior-preservation rule (binding):** removing a site must not change the seeded values. Where a
  site today seeds an actor/tenant **derived from ctx**, removal is safe. Where a site seeds a **distinct
  semantic** (see §1.4 allowlist), it is **kept** (allowlisted), not removed.
- After collapse, the census of `SeedTxIdentity(` calls **outside the chokepoint and outside the
  allowlist** is **0**.

### 1.4 The allowlist (binding — the "or allowlisted" half of the acceptance)

A short, explicit, reasoned allowlist file/const enumerates every retained non-chokepoint identity-seed.
Each entry states file:line + reason. Permitted categories, and **only** these:

- **A. Actor ≠ ctx-actor.** A site that must seed an actor that is **not** the current ctx actor (a stored
  `GrantedBy`/`assignedBy`/`CreatedBy` used as the authz actor). Each such site is verified from source to
  be a genuine distinct-actor semantic (not merely a renamed ctx actor) — otherwise it is collapsed, not
  allowlisted.
- **B. Cross-tenant API `Do` path.** A platform-admin request path (ADR 0021) that legitimately operates
  across tenants inside a request tx and therefore must **not** be auto-seeded to a single tenant. Such
  paths either keep an explicit posture or run GUC-unset by design; enumerated with reason.
- **C. The async tenant-only seeds of F3.2** (`SeedTxTenant`) — these are **not** `SeedTxIdentity` and are
  a different primitive; they are covered by §2, not counted against the F3.1 census, but listed for
  completeness.

If the true allowlist is **empty** (every site collapses cleanly), that is the ideal outcome and the
census is a literal 0. The allowlist exists so that a legitimate exception is **recorded**, never so that
the census is quietly padded.

### 1.5 The census-drift lint (blocking, `scripts/api-lint`)

- **Rule name:** `SEED-CHOKEPOINT` (or as recorded in `plan.md`), registered in the existing blocking
  `api-design-system-lint` job — **no reported-only tier**.
- **Expected:** AST-scan for `authz.SeedTxIdentity(` call expressions. Any occurrence **outside** the
  chokepoint file(s) and **outside** the declared allowlist (and outside `_test.go`) → **violation**,
  naming file:line.
- **POSITIVE proof:** clean tree post-collapse → **0 violations**.
- **NEGATIVE proof (required, captured):** add a throwaway manual `SeedTxIdentity(ctx, tx, …)` in an
  application service (not allowlisted) → lint **RED**, naming the file:line. Remove it → **GREEN**.
- **Exit criteria:** rule registered + blocking; clean-tree GREEN; synthetic manual-seed RED captured.

### 1.6 F3.1 exit criteria (all required)

Platform identity carrier exists (tenant+actor, set by authn middleware, no boundary violation) ·
chokepoint seeds per §1.2 (both `Do`/`DoReadOnly`, present→seed / absent→no-op, before Require & locks) ·
positive behavior proof captured · 62 manual sites collapsed to **0 outside chokepoint+allowlist** ·
allowlist enumerated with reasons (categories A/B only) · `SEED-CHOKEPOINT` lint registered, blocking,
GREEN clean / RED on synthetic · existing tenant-isolation suites green · `go build ./...` green ·
cross-tenant 404 behavior unchanged.

---

## 2. F3.2 — async RLS backstop (tenant-only seed + compensating lint + negative proof)

### 2.1 The tenant-only seed primitive (binding)

- **New primitive** `authz.SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error` — sets
  **only** `metaldocs.tenant_id` via `set_config('metaldocs.tenant_id', $1, true)`; requires `tenantID`
  non-empty; does **not** set or require an actor.
- **Rationale (binding):** RLS reads only `metaldocs.tenant_id`; async system work has **no human actor**
  and uses `authz.BypassSystem` for the write-tripwire. The full `SeedTxIdentity` (actor-required) is the
  wrong primitive for the async backstop. `SeedTxTenant` is the **minimum** that engages the RLS backstop.
- The async paths that also need the tripwire bypass keep calling `authz.BypassSystem` as they do today;
  `SeedTxTenant` is **additive** and orthogonal (RLS backstop vs. write-tripwire are separate gates).

### 2.2 Seed sites (binding — the five single-tenant processing txs)

`SeedTxTenant(ctx, tx, <row/payload tenant_id>)` is called at the **start** of the processing tx (before
any lock or write) in:

1. materialize job runner (existing `BeginTx` tx).
2. pdf job runner — **wrap** its write(s) in a `BeginTx`/`runner.Do` tx, then seed.
3. scheduled-publish job (existing `runner.Do` tx) — seed with the River-args tenant; before the
   `FOR UPDATE` (H-PRE-1 preserved; no `authz.Require` added).
4. notifications-fanout worker — **wrap** its insert loop in a tx, then seed.
5. staging-outbox **processing** step (per claimed row) — seed with `OutboxRow.TenantID` in the
   processing tx. The **claim** step stays GUC-unset (§2.4).

**Constraint:** a seeded processing tx must touch **exactly one** tenant's rows (ADR 0054 rule 2). If any
handler is found to mix tenants in one tx, that is an **HS-2 surface** (stop + report), not a silent
workaround.

### 2.3 The compensating lint (blocking, `scripts/api-lint`)

- **Rule name:** `ASYNC-TENANT-SEED` (or as recorded in `plan.md`), blocking, no reported-only tier.
- **Expected:** AST-scan the worker/jobs code paths (the async binaries' handler packages). For a DB
  **write** (`Exec`/`ExecContext` with INSERT/UPDATE/DELETE) against a **tenant-scoped table** (a table in
  the RLS-covered set), require that it executes inside a tx that has called `SeedTxTenant`
  (or `SeedTxIdentity`) — **unless** the write's enclosing function/site is on the **§2.4 sanctioned
  allowlist**. Fail otherwise, naming file:line + table.
- **Coverage scope (documented limitation, not a gap):** the lint is **function/handler-local** (same AST
  technique as the M2 `TRIPWIRE-ARM-DRIFT` and `authz-area-scope-binding` lints). Cross-file claim→process
  edges are covered by the **allowlist + the negative integration proof**, not by call-graph analysis
  (a larger effort — recorded bounded defer, HS-2 boundary).
- **POSITIVE proof:** clean tree → **0 violations** (the five seeded handlers pass; the allowlisted
  claim/system paths are exempt).
- **NEGATIVE proof (required, captured):** add a throwaway worker write against a tenant-scoped table in
  an unseeded, non-allowlisted site → lint **RED**. Remove it → **GREEN**.

### 2.4 The sanctioned cross-tenant / system allowlist (binding)

Explicit, reasoned allowlist (file/const). Only these categories:

- **Outbox claim steps** (ADR 0054 rule 1) — platform outbox consumer + staging-outbox claim. Cross-tenant
  `FOR UPDATE SKIP LOCKED`; tenancy enforced one row later at processing time (§2.2).
- **Cross-tenant scans** with no per-row mutation — stuck-instance-watchdog list step (its per-instance
  **action** goes through `runner.Do` and is seeded/allowlisted per its tenant).
- **`job_leases` (lease-reaper)** — genuinely has **no `tenant_id` column**; RLS is structurally N/A
  (nothing to scope); recorded as such.
- **`idempotency_keys` (idempotency-janitor)** — **is** a `tenant_id`-bearing FORCE-RLS table (1 of the 33);
  its TTL `DELETE … WHERE expires_at < now()` (`internal/modules/jobs/idempotency_janitor/job.go:34`) is a
  **sanctioned cross-tenant system-maintenance sweep** run GUC-unset under the NULL-permissive hatch — same
  category as the audit-integrity scan, **not** a table where RLS "cannot apply." Its janitor package is
  outside the `ASYNC-TENANT-SEED` scanned handler roots, so no lint allowlist entry is required.
- **Audit-integrity scan** — read-only, system-wide by design.

### 2.5 The negative RLS proof (integration — the load-bearing acceptance)

The **only** test class that can pin the RLS backstop (application tests are sqlmock). Authored
`//go:build integration`, testdb factory (RLS-enforced, `NOSUPERUSER`-equivalent connection), mirroring
the existing tenant-isolation drives. **Shape (binding):**

- **Setup:** two tenants A and B, each with a `documents` (or other RLS-covered table) row.
- **NEGATIVE-before / leak demonstration:** in a tx with **no** tenant GUC seeded (today's worker
  behavior), a cross-tenant `SELECT`/`UPDATE` against tenant B's row **succeeds/sees the row** — proving
  the pre-fix leak surface (captured; this is the "async has zero backstop" evidence).
- **POSITIVE-after / backstop engaged:** in a tx that calls `SeedTxTenant(ctx, tx, A)` first (the F3.2
  behavior), the **same** cross-tenant access to tenant B's row is **blocked**:
  - `SELECT`/`UPDATE`/`DELETE` of B's row → **0 rows** (invisible), **and**
  - `INSERT`/`UPDATE` producing a B-tenant row → **error** `new row violates row-level security policy`
    (SQLSTATE `42501`).
- Both captured, labeled **real-DB (testdb factory), not sqlmock**.
- **Run scope:** targeted `go test -tags integration -run 'RLS|Tenant|Isolation' ./...` only — never the
  full suite (mission §10). If the box cannot run `-tags integration`, the drive is **authored** and the
  run recorded as a **bounded defer** with the run-trigger (M1/M2 env-risk precedent).

### 2.6 F3.2 exit criteria (all required)

`SeedTxTenant` primitive exists (tenant-only, no actor) · seeded at the 5 single-tenant processing txs
(pdf + notifications-fanout wrapped in a tx) · no handler mixes tenants in one tx (else HS-2) ·
`ASYNC-TENANT-SEED` lint registered, blocking, GREEN clean / RED on synthetic unseeded worker write ·
sanctioned allowlist enumerated (§2.4) · negative RLS integration proof captured (leak pre-seed → blocked
post-seed, real-DB) · scheduled-publish + fanout still function (targeted drives green) · `go build ./...`
green · H-PRE-1 preserved (seed before locks; no `authz.Require` added to locked paths).

---

## 3. F3.3 — ADR 0027 + wiki amendment

### 3.1 ADR 0027 amendment (binding content)

Append a **dated amendment** (do not rewrite history) to `wiki/decisions/0027-rls-adoption-sequencing.md`
stating:

1. **The NULL-permissive design is deliberate** and load-bearing: `GUC unset → all rows visible` exists so
   GUC-less system/scan paths (janitors, outbox claim per ADR 0054, bootstrap) work without a tenant
   context. It is **not** a bug and **must not** be removed.
2. **The sync↔async asymmetry that existed pre-M3:** the API binary seeded the GUC (so FORCE RLS was a
   real backstop), but `metaldocs-worker`/`metaldocs-jobs` seeded **nothing** — async isolation rested
   solely on hand predicates, and one bad worker join was a silent cross-tenant leak with no gate.
3. **How M3 closes it:** (a) the `TxRunner` chokepoint auto-seeds tenant+actor from the platform identity
   carrier for the API binary (§1); (b) async single-tenant processing txs seed the claimed row's tenant
   via `SeedTxTenant`, engaging FORCE RLS for worker/jobs (§2), **completing ADR 0054 rule 2**; (c) two
   blocking lints (`SEED-CHOKEPOINT`, `ASYNC-TENANT-SEED`) plus the negative RLS integration proof make
   the seeding a **structural** property, not discipline.
4. **The residual GUC-unset surface** (outbox claim, cross-tenant scans, `idempotency_keys`/`job_leases`
   system tables) is enumerated as **sanctioned** and cross-referenced to ADR 0054.
5. Cross-reference **ADR 0054** (its rule 2 is now enforced) and the M3 milestone folder.

### 3.2 Wiki

- Update the tenancy wiki page(s) (`wiki/architecture/*tenan*`, `wiki/backend/*`, `wiki/concepts/authz-tiers.md`
  where they describe seeding/RLS) so **no stale claim survives**: no "async has no RLS backstop", no "seed
  manually at ~85 sites", no "RLS only on controlled_documents + audit_events". Reflect: chokepoint
  autoseed, async tenant-seed, the two lints, and the per-binary posture (§4).
- `wiki-curator` pass clean (stamps refreshed for touched docs; file:line anchors resolve).

### 3.3 F3.3 exit criteria

ADR 0027 carries the dated amendment with all five points · wiki tenancy docs match runtime truth ·
0027 ↔ 0054 ↔ M3 cross-refs resolve · `wiki-curator` clean · **no code-behavior change** in this feature.

---

## 4. ★ Expected RLS behavior per binary (the mission's explicit D4 requirement — binding)

| Binary | Who sets the GUC (post-M3) | Tenant GUC state on a business tx | FORCE-RLS effect | GUC-unset surface (sanctioned) |
|---|---|---|---|---|
| **metaldocs-api** (sync) | `TxRunner` chokepoint, auto from platform identity carrier (authn middleware) | **Seeded** to the authenticated tenant on every request `Do`/`DoReadOnly` | **Active backstop** — cross-tenant SELECT/UPDATE/DELETE → 0 rows; wrong-tenant INSERT → `42501` | Allowlisted cross-tenant platform-admin paths (§1.4-B) + any pre-auth/system `Do` with no ctx identity (no-op seed, NULL-permissive) |
| **metaldocs-worker** (async: materialize, pdf, staging-outbox, platform outbox) | Per-message `SeedTxTenant` in the processing tx (claimed row's tenant) | **Seeded** to the claimed row's tenant in each single-tenant processing tx | **Active backstop** on processing writes — same 0-rows / `42501` semantics | **Claim** steps (ADR 0054 rule 1) run GUC-unset by design (allowlisted §2.4) |
| **metaldocs-jobs** (async: scheduled-publish, notifications-fanout, River) | Per-job `SeedTxTenant` in the work tx (River-args / event tenant) | **Seeded** to the job's tenant | **Active backstop** on job writes | Nothing tenant-scoped runs unseeded except the §2.4 allowlist |
| **(janitors, hosted in metaldocs-api)** | Not seeded (system, no tenant) | **Unseeded** — NULL-permissive by design | Intentionally inert (cross-tenant maintenance); guarded by the §2.4 allowlist. Cross-tenant TTL/maintenance sweeps run GUC-unset by design — e.g. the idempotency-janitor `DELETE` on the `tenant_id`-bearing FORCE-RLS `idempotency_keys` (same class as the audit scan); `job_leases` genuinely has no `tenant_id` | Entire janitor scan/maintenance surface (sanctioned) |

**Invariant restated:** the RLS **policy** is byte-identical before and after M3 (NULL-permissive, FORCE,
33 tables). What changes is **GUC seeding coverage**: from "API only, by 62 manual acts" to "API by
chokepoint + async by per-message primitive, guarded by two blocking lints + a negative integration
proof." Cross-tenant URL → 404 and the existing cross-tenant isolation suites remain green.

---

## 5. Cross-feature constraints (bind all three features)

- **Do NOT weaken RLS:** no policy/`ENABLE`/`FORCE`/NULL-permissive change; **no** `WITH CHECK` removal or
  `FOR` narrowing. M3 adds seeding only.
- **tx-local GUCs only:** `set_config(..., true)` exclusively; **never** session-level (`false`) on a
  pooled connection (leaks to the next borrower).
- **Module boundary:** `internal/platform/db` reads identity only from `internal/platform/*`; never imports
  `internal/modules/iam` (guarded).
- **H-PRE-1:** seeding is a `SET LOCAL` config write, not an authz-recording read; seed **before** any
  lock; add **no** `authz.Require` to any lock-holding tx.
- **Registry/primitive discipline:** async uses `SeedTxTenant` (tenant-only); the actor-required
  `SeedTxIdentity` is **not** used for the async backstop.
- **Blocking CI:** both new lints fail the build on any violation (no reported-only tier).
- **Targeted tests only:** no full integration suite; `-run 'RLS|Tenant|Isolation'` + the new negative
  proof; bounded defer if the box cannot run `-tags integration`.
- **Separation of powers:** implementation via subagents (sonnet implement/review, haiku mechanical, never
  fable, ≤15 concurrent); main session orchestrates/reviews/commits; the `milestone-validator` judges and
  writes `qa/milestone-qa.md`; the main session flips status only on PASS.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored.**

## 6. Bounded defers (recorded, with triggers)

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Cross-file claim→process drift coverage via call-graph analysis (both lints are handler-local) | Call-graph resolution is a larger static-analysis effort (HS-2 boundary); the allowlist + negative integration proof cover the async class meanwhile | Post-mission static-analysis hardening |
| Integration-drive execution if the local box cannot run `-tags integration` | 20-min box constraint / env (mission §10) | Run on CI or a capable box before program close-out; drives authored regardless |
| Add explicit `WITH CHECK` to policies (currently implied by `FOR ALL`) for clarity | Behavior already correct (USING reused as WITH CHECK); explicit clause is cosmetic hardening, out of M3's "no policy change" scope | M9 governance-hygiene, if desired |
| Session-level identity carrier for worker/jobs (parallel to the API middleware) so async could also auto-seed at a chokepoint | Async tenant is per-message (claimed row), not per-connection — a middleware-style carrier does not fit the claim→process shape; explicit `SeedTxTenant` is the correct seam | Revisit under M5 River consolidation (single job primitive) |
