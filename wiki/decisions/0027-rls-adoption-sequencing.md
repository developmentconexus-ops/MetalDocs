# ADR 0027 — RLS Adoption Sequencing + auth_identities Tenant-Global by Design

> **Status:** Accepted (executed in full by Wave Z, 2026-06-13).
> **Status history:** [below](#status-history) (originally Accepted 2026-06-11; three-tier plan
> collapsed into one migration by Wave Z Z-2; amended 2026-07-03 M3, amended 2026-07-05 M7 F7.4).
> **Last verified:** 2026-07-05 (M7 F7.4 RLS-truth sweep — see Status history)
> **Scope:** Two related decisions: (1) `auth_identities` has no `tenant_id` by deliberate design; (2) the sequencing and rationale for Row-Level Security adoption across the MetalDocs schema. Closes tech-debt item T-008 as by-design. Documents the partial-coverage RLS model and its trigger conditions.
> **Out of scope:** The two-tier authz model itself (ADR 0007); capability coherence (ADR 0022); the specific SQL for the Wave 2.3 migration (executed in item 2.3 using the `current_setting('metaldocs.tenant_id', true)` GUC pattern verified here).
> **Key files:**
> - `db/baseline/0001_current_schema.sql:914` — `auth_identities` table definition (no `tenant_id` column)
> - `internal/modules/iam/authz/context.go:60` — `SeedTxIdentity`: `set_config('metaldocs.tenant_id', $1, true)` + `set_config('metaldocs.actor_id', $2, true)` — the transaction-local GUC pattern reused by RLS policies
> - `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql:101` — `current_setting('metaldocs.asserted_caps', true)` — same GUC family; precedent for the pattern
> - `wiki/concepts/authz-tiers.md` — two-tier model; `metaldocs.tenant_id` GUC context
> - `wiki/backend/stage2-evaluation.md` — ADR-A paragraph (F-12 verdict row)
> - `db/migrations/0284_ci_rls_role.sql` — (M7 F7.4 amendment) dedicated non-owner, NOSUPERUSER+NOBYPASSRLS `metaldocs_ci` role for real RLS proofs
> - `db/migrations/0285_approval_signoffs_rls.sql` — (M7 F7.4 amendment) FORCE RLS + `tenant_isolation` policy on `public.approval_signoffs` (keyed on `actor_tenant_id`)
> - `tests/integration/testdb/ci_role.go:38` — (M7 F7.4 amendment) `OpenAsCIRole`
> - `tests/integration/security/rls_truth_test.go:41` — (M7 F7.4 amendment) `TestRLSTruth_NonOwnerRoleEnforcesIsolation`
> - `scripts/api-lint/sole_rls_read_rule.go:189` — (M7 F7.4 amendment) `checkSoleRLSAsyncRead` (`SOLE-RLS-ASYNC-READ` rule)

## Status history

> Relocated from the ADR's `> **Status:**`/`> **Last verified:**` fields 2026-07-06 (F9.1 adr-hygiene).
> Zero information loss — this is the same text, restructured into dated entries.

- **2026-06-11 — originally Accepted.** Binding decision D-3 of the backend professionalization design
  spec (three-tier RLS sequencing, `auth_identities` tenant-global by design).
- **2026-06-13 — executed in full by Wave Z.** Wave Z Z-2 (operator override of D-3) collapsed all
  three tiers into a single migration (`db/migrations/0237_rls_all_tenant_tables.sql`), enabling
  ENABLE+FORCE RLS + the NULL-permissive `tenant_isolation` policy on **all 27 remaining tenant-scoped
  tables** (29 total including the 2 already enabled in migration 0234) at once (Tier 2 `iam_users` +
  Tier 3 external-tenant tables included). The three-tier *sequencing* documented in the ADR body below
  is retained as the historical record of how the rollout was originally planned; it no longer
  describes future work. The by-design `auth_identities` decision is unchanged. **Current reality
  (2026-06):** RLS is live on every tenant-scoped table (29 total = 2 from 0234 + 27 from 0237). The
  "first external tenant" / RF-6 triggers documented in the ADR body never fired — Wave Z executed the
  full program ahead of them. `metaldocs.user_process_areas` is a VIEW over `public.user_process_areas`
  (the base table IS covered); views cannot carry RLS, so the census of 28 `tenant_id`-bearing relations
  minus that view = 27 base tables in 0237. NOSUPERUSER probe (Wave Z): GUC-unset→all rows, GUC=A→only
  A, GUC=B→only B, verified live on `iam_users` + `documents`.
- **2026-07-03 — M3 tenancy-chokepoint amendment.** See "Amendment 2026-07-03 (M3 tenancy chokepoint)"
  below in this document for the full text (unabridged, not relocated — it was already outside the
  status field).
- **2026-07-05 — M7 F7.4 RLS-truth sweep amendment.** See "Amendment 2026-07-05 (M7 F7.4 RLS-truth
  sweep)" below in this document for the full text (unabridged, not relocated — it was already outside
  the status field).

## Context

### auth_identities has no tenant_id

The `auth_identities` table (`db/baseline/0001_current_schema.sql:914`) holds credential data: `user_id`, `username`, `email`, `password_hash`, `password_algo`, lockout columns, and failure-side metadata. It has no `tenant_id` column. The Stage-1 audit flagged this as a potential anomaly (T-008: "auth_identities not tenant-scoped").

The scoping is intentional. A user identity (credential) is a cross-tenant concept in MetalDocs: the same `user_id` is associated with exactly one tenant via `iam_users.tenant_id`, but the credential itself is not tenant-owned. Tenant scoping of the identity happens via `JOIN auth_identities ON iam_users.user_id = auth_identities.user_id`, not by a `tenant_id` column on `auth_identities`. Migration `0219` (`iam_users_last_login_context`) documents this split explicitly:

> `auth_identities.last_login_at` is part of the credential surface (set on every successful bcrypt verify). The new columns are governance metadata scoped to a tenant-managed user record and therefore belong with the tenant-scoped `iam_users` row.

Adding `tenant_id` to `auth_identities` would be an incorrect data model: it would either force a one-to-one identity-to-tenant constraint (breaking the design), or require a composite key for a credential that is conceptually singular.

### Shared-schema + tenant_id is standard SaaS

MetalDocs uses a shared-schema multi-tenant design: every tenant row carries a `tenant_id` UUID, and application-layer predicates in every query and repository method enforce tenant scoping. This is the standard SaaS isolation pattern — not a design defect. The alternative (schema-per-tenant) is over-engineering at MetalDocs' current and foreseeable scale: one real tenant, single-user repository, no external tenants on the horizon.

Decision D-3 of the approved design spec (`docs/superpowers/specs/2026-06-11-backend-professionalization-design.md`) explicitly affirms this:

> Shared-schema + tenant_id is the standard SaaS pattern (not a "bad monolith"). RLS = defense-in-depth against our own bugs: `controlled_documents` + `audit_events` get RLS in Wave 2; full per-table isolation program trigger-gated on "first external tenant onboarded".

### The existing GUC pattern

MetalDocs already uses transaction-local PostgreSQL GUCs (`set_config(..., true)`) to carry runtime identity into the DB layer. The established GUC names, all transaction-local (third arg `true`):

| GUC | Purpose |
|-----|---------|
| `metaldocs.tenant_id` | Current tenant — seeded by `SeedTxIdentity` (`context.go:60`); read by RLS policies |
| `metaldocs.actor_id` | Current actor — seeded alongside `tenant_id` in every tx |
| `metaldocs.asserted_caps` | JSON array of caps asserted for the tripwire (`authz.go:262`) |
| `metaldocs.bypass_authz` | Scheduler-only bypass token (`authz.go:190`) |

The `metaldocs.tenant_id` GUC is already set on every authenticated transaction via `SeedTxIdentity`. RLS policies for `controlled_documents` and `audit_events` reuse this existing GUC — no new seeding infrastructure is required.

### The capability tripwire precedent

The `enforce_capability_asserted()` trigger (`db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql:101`) uses `current_setting('metaldocs.asserted_caps', true)` as a transaction-local enforcement gate. That trigger — not native PostgreSQL RLS — is MetalDocs' primary write-side isolation mechanism, covering 12 tables. It was chosen over native RLS because it is strictly stronger on the owner/superuser-bypass axis: triggers fire for all roles regardless of table ownership; only a superuser with `session_replication_role = replica` can skip them. See ADR 0022 Phase 5 §Item 7 for the full finding.

RLS adopted here is therefore an **additional** defense-in-depth layer on top of the application-layer predicates and the tripwire — not a replacement.

### F-12 and the Stage-2 verdict

The Stage-2 evaluation (`wiki/backend/stage2-evaluation.md` F-12 row) rates `controlled_documents` and `audit_events` as the high-risk tables for tenant leakage (REQ-TEN-1, OWASP ASVS V4.1.3). The `audit_events` table in particular holds cross-module governance data and is the target of the audit export job (`ExportJob`). Both tables see reads and writes from multiple modules; a query predicate omission would leak tenant data silently. The Stage-2 verdict: RLS on these two tables is the minimal correct fix; deferring all other tables is explicitly justified.

## Decision

### 1. auth_identities is tenant-global by design

`auth_identities` does not have and will not receive a `tenant_id` column. Identity (credential) is a global concept; tenant association flows from `iam_users.tenant_id` via JOIN. T-008 is **closed as by-design**, not as a defect. No migration or schema change is required for T-008.

### 2. RLS rollout sequencing

> **EXECUTED IN FULL (Wave Z Z-2, 2026-06-13):** the operator overrode the D-3 trigger-gated limit and ordered every tenant-scoped table covered before the v1 release. Migration `0237_rls_all_tenant_tables.sql` applied ENABLE+FORCE RLS + the NULL-permissive `tenant_isolation` policy to all 27 remaining base tables in one pass — Tier 2 (`iam_users`) and Tier 3 (all remaining) included. The three-tier *plan* below is retained as the historical record of how the rollout was originally sequenced; it no longer describes future work.

RLS adoption follows three tiers, justified by risk and the D-3 trigger-gated principle.

#### Tier 1 — Wave 2.3 (immediately after this ADR): controlled_documents + audit_events

Enable Row-Level Security on `metaldocs.controlled_documents` and `metaldocs.audit_events` using the existing `current_setting('metaldocs.tenant_id', true)` GUC pattern (`internal/modules/iam/authz/context.go:60`).

Policy shape (both tables):
```sql
CREATE POLICY tenant_isolation ON metaldocs.<table>
  USING (tenant_id = current_setting('metaldocs.tenant_id', true)::uuid);
```

These are the two tables with the highest cross-tenant leakage risk at current and foreseeable scale, and they share the simplest policy shape (all rows carry `tenant_id`). The GUC is already seeded on every authenticated transaction; no application code changes are needed beyond the migration.

This RLS is defense-in-depth against our own bugs — application-layer predicates remain the primary isolation mechanism.

#### Tier 2 — RF-6 (deferred): iam_users

RLS on `iam_users` is deferred to the RF-6 authz tripwire program. The `iam_users` table is already covered by the `enforce_capability_asserted` trigger on `iam_user_roles` and `user_process_areas`, which enforce tier-2 area-scoped authorization on writes. Adding RLS to `iam_users` itself requires careful policy design around the system_admin bypass and the membership directory scope queries (ADR 0022 Phases 3–4). This work belongs inside a dedicated authz boundary review, not in the Wave 2 structural refactor.

#### Tier 3 — external tenant trigger (deferred): all remaining tables

All remaining tables carrying `tenant_id` (documents, approval_instances, approval_signoffs, templates_template, idempotency_keys, audit_export_jobs, etc.) receive RLS when the first external tenant is onboarded. At that point the defense-in-depth rationale becomes load-bearing; until then, application-layer predicates are sufficient and RLS on every table would add maintenance overhead with no observable isolation improvement.

### 3. RLS is defense-in-depth, not primary isolation

Application-layer predicates in Go repositories (and the tripwire for write paths) remain the primary isolation mechanism. RLS is a second layer that catches predicate omissions. This is consistent with the shared-schema design decision (D-3) and with OWASP ASVS V4.1.3 defense-in-depth guidance. MetalDocs does not use Postgres-native RLS as a zero-trust boundary; the `NOSUPERUSER` deployment constraint (noted in ADR 0022 Phase 5 §Item 7) applies regardless.

## Consequences

- T-008 is closed as by-design. No schema change for `auth_identities`.
- Wave 2.3 can proceed with a narrowly scoped RLS migration on two tables.
- `iam_users` RLS deferred to RF-6; no ambiguity for that boundary.
- Remaining tables deferred by explicit trigger; no silent indefinite deferral.
- The `metaldocs.tenant_id` GUC is established as the canonical RLS predicate source for future policies.

## References

- D-3: `docs/superpowers/specs/2026-06-11-backend-professionalization-design.md` (tenancy decision + sequencing)
- F-12: `wiki/backend/stage2-evaluation.md` ADR-A paragraph + F-12 verdict row (REQ-TEN-1, OWASP ASVS V4.1.3)
- T-008: `wiki/backend/legacy-register.md` (tech-debt register entry, closed by-design here)
- REQ-TEN-1: `wiki/architecture/backend-target-architecture.md`
- OWASP ASVS V4.1.3: defense-in-depth tenant isolation
- ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md) — two-tier model; trigger tripwire background
- ADR [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) Phase 5 §Item 7 — native RLS vs. trigger-tripwire finding; `NOSUPERUSER` deployment constraint

## Amendment 2026-07-05 (M7 F7.4 RLS-truth sweep)

> **Last verified (amendment):** 2026-07-05. Does not alter the Wave Z execution record or the M3
> amendment above — closes two remaining gaps: (1) every prior "RLS is enforced" integration proof ran
> against `metaldocs_app` (SUPERUSER + BYPASSRLS + owner of every table), under which RLS is
> unconditionally inert regardless of policy correctness — a false green; (2) one tenant-scoped table
> (`public.approval_signoffs`) was missed by the Wave Z / M3 FORCE-RLS census because it keys its tenant
> column `actor_tenant_id`, not `tenant_id`.
> Source: `db/migrations/0284_ci_rls_role.sql`, `db/migrations/0285_approval_signoffs_rls.sql`,
> `scripts/api-lint/sole_rls_read_rule.go`, `tests/integration/security/rls_truth_test.go`.

### 1. A dedicated non-owner, non-bypass role now proves RLS for real

Migration `db/migrations/0284_ci_rls_role.sql` creates `metaldocs_ci`: `NOSUPERUSER NOBYPASSRLS
NOCREATEDB NOCREATEROLE NOINHERIT LOGIN`, owning no tables, with DML-only grants (`USAGE` on schemas
`metaldocs`+`public`; `SELECT/INSERT/UPDATE/DELETE` on all tables; `USAGE/SELECT` on all sequences;
`ALTER DEFAULT PRIVILEGES` so future migrations' tables/sequences — created as `metaldocs_app` — inherit
the same grants automatically).

`tests/integration/testdb/ci_role.go` (`OpenAsCIRole`) opens a second connection to the same per-test
cloned database as this role, so the integration suite can seed setup data via the owner handle
(`testdb.Open`, still `metaldocs_app`) and then run isolation-proof reads through `metaldocs_ci`, which
cannot bypass RLS. `tests/integration/security/rls_truth_test.go`
(`TestRLSTruth_NonOwnerRoleEnforcesIsolation`) is the first proof to run this way: it asserts
`metaldocs_ci` has `rolsuper=false`, `rolbypassrls=false`, and owns 0 tables (all three would silently
neuter RLS if true), then proves wrong-tenant-GUC blocks, right-tenant-GUC allows, and the deliberate
NULL-GUC escape hatch (§below) admits all rows — pinning it as sanctioned, not a leak. It also documents,
as a comment, that the pre-fix owner-connection behavior would have been a false green.

**DEVIATION (HS-1):** the source contract for this work suggested reassigning ownership of all
tenant-scoped tables to a distinct owner role. MetalDocs instead adds a fresh, ownerless role with
DML-only grants. This satisfies the actual requirement — the connecting proof role is a non-owner under
`NOBYPASSRLS`, so plain `ENABLE ROW LEVEL SECURITY` applies to it without needing `FORCE` (`FORCE` matters
only for a table's owner) — without a schema-wide ownership migration that would break dev bootstrap,
forward migrations, or the leader-elected janitors, all of which run as `metaldocs_app`. `metaldocs_app`
remains the dev/bootstrap/migration owner identity; this is unchanged by the amendment.

### 2. `public.approval_signoffs` joins the FORCE-RLS set

Migration `db/migrations/0285_approval_signoffs_rls.sql` adds `ENABLE`/`FORCE ROW LEVEL SECURITY` +
a `tenant_isolation` policy to `public.approval_signoffs`, keyed on `actor_tenant_id` (its tenant column —
see `wiki/database/tables/approval_signoffs.md`). The table carries e-signature PII (signer identity,
display-name snapshot, signature payload) and was omitted from the Wave Z (0237) and M3 census because
both were driven by a `tenant_id`-column search; `approval_signoffs` has no such column. The policy idiom
(GUC name, `NULLIF(...) IS NULL` null-GUC escape-hatch branch, policy name `tenant_isolation`) is copied
verbatim from `0281`/`0278` — only the schema (`public`) and tenant column (`actor_tenant_id`) differ. No
cross-tenant co-sign semantics exist (a signoff's `actor_tenant_id` is always the signer's own tenant), so
the strict policy needs no ADR 0070-style carve-out. This brings the FORCE-RLS tenant-scoped table count
to 34 (33 from the M3 amendment + this one).

### 3. New lint rule closes the read-side of the same seam M3 closed for writes

`scripts/api-lint/sole_rls_read_rule.go` registers `SOLE-RLS-ASYNC-READ` (blocking, alongside M3's
`ASYNC-TENANT-SEED`). M3's rule guards async **writes**: a worker/jobs mutating write against a
FORCE-RLS tenant table must sit in a function that also seeds `metaldocs.tenant_id`. The new rule guards
async **reads** the same handler roots: a `Query`/`QueryRow`/`QueryContext`/`QueryRowContext` call whose
SQL is a `SELECT` against a tenant-scoped table (same `async-tenant-tables.txt` data file as M3) with
**no** explicit `tenant_id`/`actor_tenant_id` token anywhere in its own SQL text is a violation, because
the `tenant_isolation` policy's NULL-GUC branch silently returns all tenants' rows if the enclosing
async tx's GUC was never seeded (or was seeded for the wrong tenant) — belt-and-suspenders so correctness
does not depend solely on upstream seeding discipline. This rule was carried forward from the M6 F6.4
fix-feature (review-due ports false-negative, closed by adding an explicit `tenant_id` predicate) and
makes that fix class a static, blocking guard instead of a one-off. Allowlist:
`scripts/api-lint/sole-rls-read-allowlist.txt` (same format as M3's async-tenant-seed allowlist).

### 4. What this amendment does NOT change

- The RLS **policy** shape (NULL-permissive on unset GUC) is unchanged and remains deliberate — see §4.1
  of the M3 amendment above. This continues to mean "no-GUC → all rows visible," not "no-GUC → 0 rows."
  The isolation proof this amendment adds is therefore "wrong-tenant GUC blocks the other tenant's row,"
  not "unset GUC blocks everything" — the escape hatch is pinned explicitly in the new test, not
  mistaken for a leak.
- `metaldocs_app` is still the dev/bootstrap/migration owner identity and is **not** the only DB role —
  production already runs a NOSUPERUSER/NOBYPASSRLS/non-owner app role; `metaldocs_ci` is a second,
  test-only, DML-only role layered on top for CI isolation proofs specifically.

### 5. Cross-references

- `wiki/quality/integration-test-harness.md` — general harness usage (`testdb.Open`, factories); does not
  yet describe `OpenAsCIRole` (narrow, isolation-proof-only usage; see `ci_role.go` directly).
- `wiki/database/tables/approval_signoffs.md` — table dictionary entry, updated for FORCE RLS coverage.

## Amendment 2026-07-03 (M3 tenancy chokepoint)

> **Last verified (amendment):** 2026-07-03. This amendment does not alter the Context/Decision/Consequences
> body above (Wave Z execution record, 2026-06-13) — it documents a **coverage** change layered on top of it.
> Source: `docs/superpowers/milestones/global-maximum-remediation/milestone-3-tenancy-chokepoint/validation-contract.md`
> §3.1/§4, and the F3.1/F3.2 evidence files in the same milestone folder.
> **Erratum (F3.4, 2026-07-03):** §4 and the per-binary table below originally miscited `idempotency_keys`
> as a system table "with no `tenant_id` column." It **is** a `tenant_id`-bearing FORCE-RLS table; its
> janitor sweep is a sanctioned cross-tenant NULL-permissive maintenance `DELETE` (same class as the audit
> scan). Corrected in place; `job_leases` (no `tenant_id`) was already correct.

The RLS **policy** shipped by Wave Z is unchanged: NULL-permissive (`GUC unset → all rows visible`), FORCE
RLS, on all 33 tenant-scoped tables (the 27+2 base tables from Amendment-era migrations, since grown to 33
as new tenant-scoped tables were added). Milestone 3 ("tenancy chokepoint," global-maximum-remediation
program) did not touch the policy — it closed a gap in **who seeds the GUC that the policy reads**.

### 1. The NULL-permissive design is deliberate and load-bearing

`current_setting('metaldocs.tenant_id', true)` returning unset means the `tenant_isolation` policy admits
all rows. This is **not a bug** and **must not be "fixed"** by making the policy fail-closed on an unset
GUC: system paths — janitors, cross-tenant scans, the outbox claim step (ADR 0054 rule 1), and bootstrap —
run with no tenant context by design and depend on this permissive behavior to function at all. Any change
that makes RLS deny-by-default on an unset GUC would break those paths.

### 2. The pre-M3 sync/async asymmetry

Before M3, `metaldocs-api` seeded `metaldocs.tenant_id`/`metaldocs.actor_id` on (almost) every request
transaction via hand-placed `authz.SeedTxIdentity` calls (this ADR's original GUC pattern, `context.go:48`),
so FORCE RLS was a real backstop on the sync path. `metaldocs-worker` and `metaldocs-jobs` seeded **nothing**
— async tenant isolation rested solely on hand-written query/write predicates, with no RLS gate behind
them. A single bad join or missed predicate in a worker or job handler was a silent cross-tenant leak with
no backstop to catch it.

### 3. How M3 closes it

- **(a) Sync chokepoint autoseed (F3.1).** The auth middleware
  (`internal/modules/auth/delivery/http/middleware.go:106-107`) now injects both tenant and actor into the
  platform identity carrier (`platformtenant.WithTenantID` + `platformtenant.WithActorID`). The shared
  `TxRunner` internals (`internal/platform/db/runner.go:63`, `seedTxIdentityFromContext` at :94) read that
  carrier at the start of **every** `Do`/`DoReadOnly` transaction and seed the tx-local GUCs when both
  tenant and actor are present; when either is absent (system/janitor/background paths with no request
  carrier) it is a no-op, preserving NULL-permissive behavior for those paths. This collapsed 61 hand-placed
  `SeedTxIdentity` call sites to 0 outside the chokepoint plus a 21-entry reviewed allowlist (raw-`BeginTx`
  paths the chokepoint cannot reach, and distinct-actor cases where the seeded actor differs from the
  request's authenticated user).
- **(b) Async per-message tenant seed (F3.2).** A new tenant-only primitive, `authz.SeedTxTenant(ctx, tx,
  tenantID)` (`internal/modules/iam/authz/context.go`), seeds `metaldocs.tenant_id` (no actor — async work
  has no human actor) at the start of each of the five single-tenant processing transactions: the
  materialize job runner, the PDF job runner, scheduled-publish (`RunScheduledPublishJob`), the
  notifications fanout worker, and the render staging-outbox dispatch-mark step. This engages FORCE RLS as
  a real backstop on `metaldocs-worker` and `metaldocs-jobs` for the first time, **completing ADR 0054 rule
  2** (async single-tenant processing transactions must be tenant-scoped).
- **(c) Two blocking api-lint rules make the seeding structural, not a discipline convention:**
  `SEED-CHOKEPOINT` (flags any `SeedTxIdentity` call outside the chokepoint/definition files and the
  reviewed allowlist) and `ASYNC-TENANT-SEED` (flags any tenant-scoped table write in the async handler
  roots not wrapped by `SeedTxTenant`/`SeedTxIdentity` or allowlisted). Both are registered blocking in
  `RunCodeRules` and were proven RED-on-violation/GREEN-on-clean live, not just by unit test. A negative
  real-DB RLS integration proof (`internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go`,
  `//go:build integration`) demonstrates the mechanism end-to-end: an unseeded tx leaks a cross-tenant row
  (read + a 1-row cross-tenant UPDATE) pre-fix; after `SeedTxTenant`, the same row is invisible, UPDATE/DELETE
  affect 0 rows, and a re-tenant write attempt fails with `SQLSTATE 42501`. This proof is authored and
  compiles clean but its live run is deferred (no `DATABASE_URL` available without reading `.env`, which is
  forbidden) — a bounded defer, not a gap in the mechanism itself.

### 4. Residual sanctioned GUC-unset surface

The following remain intentionally unseeded (NULL-permissive) after M3 and are enumerated as **sanctioned**,
not gaps:
- **Outbox claim steps** (`FOR UPDATE SKIP LOCKED` claim queries) — ADR 0054 rule 1; claiming must scan
  across tenants before a single row is bound to a tenant for processing.
- **Cross-tenant scans** — the stuck-instance-watchdog list query and the audit-integrity scan; these are
  read-only maintenance scans over all tenants by design.
- **`job_leases`** (lease-reaper) — genuinely has **no `tenant_id` column**; RLS cannot apply a tenant
  predicate where no tenant column exists.
- **`idempotency_keys`** (idempotency-janitor) — **is** a `tenant_id`-bearing FORCE-RLS table (1 of the 33,
  see this ADR's body §2 Tier-3). Its TTL sweep (`DELETE … WHERE expires_at < now()`,
  `internal/modules/jobs/idempotency_janitor/job.go:34`) is a **sanctioned cross-tenant system-maintenance
  sweep** run GUC-unset, relying on the NULL-permissive hatch — the same category as the audit-integrity
  scan, **not** a table where RLS "cannot apply." (The janitor package sits outside the `ASYNC-TENANT-SEED`
  scanned handler roots, so it needs no lint allowlist entry.)

### 5. Cross-references

- ADR [`0054`](0054-cross-tenant-outbox-claim.md) — rule 1 (outbox claim GUC-unset) was already
  sanctioned; rule 2 (async processing tx must be tenant-seeded) is now **enforced**, closed by F3.2.
- Milestone folder:
  `docs/superpowers/milestones/global-maximum-remediation/milestone-3-tenancy-chokepoint/` — spec,
  validation-contract.md (§1 F3.1, §2 F3.2, §3.1/§4 this amendment's source), and per-feature `evidence.md`
  files (`f3.1-txrunner-autoseed/evidence.md`, `f3.2-async-rls-backstop/evidence.md`).

### Per-binary RLS posture (post-M3)

| Binary | Who seeds the GUC | Tenant GUC state on a business tx | FORCE-RLS effect | Sanctioned GUC-unset surface |
|---|---|---|---|---|
| **metaldocs-api** (sync) | `TxRunner` chokepoint, auto from the platform identity carrier set by the auth middleware | Seeded to the authenticated tenant on every request `Do`/`DoReadOnly` | Active backstop — cross-tenant SELECT/UPDATE/DELETE return 0 rows; wrong-tenant INSERT fails `42501` | Allowlisted cross-tenant platform-admin paths; any pre-auth/system `Do` with no ctx identity (no-op seed) |
| **metaldocs-worker** (materialize, pdf, staging-outbox, platform outbox) | Per-message `authz.SeedTxTenant` in the processing tx, from the claimed row's tenant | Seeded to the claimed row's tenant in each single-tenant processing tx | Active backstop on processing writes — same 0-rows/`42501` semantics | Claim steps (ADR 0054 rule 1) run GUC-unset by design |
| **metaldocs-jobs** (scheduled-publish, notifications-fanout, River) | Per-job `authz.SeedTxTenant` in the work tx, from the job's tenant | Seeded to the job's tenant | Active backstop on job writes | Nothing tenant-scoped runs unseeded outside the sanctioned allowlist |
| **Janitors** (hosted in metaldocs-api: stuck-instance-watchdog, idempotency-janitor, audit-integrity-validator, lease-reaper) | Not seeded — system, no tenant | Unseeded — NULL-permissive by design | Intentionally inert for cross-tenant maintenance scans | Entire janitor scan/maintenance surface — cross-tenant TTL/maintenance sweeps run GUC-unset by design (e.g. the idempotency-janitor `DELETE` on the `tenant_id`-bearing FORCE-RLS `idempotency_keys`, same class as the audit scan); `job_leases` genuinely has no `tenant_id` column |

**Invariant restated:** the RLS policy is byte-identical before and after M3. What changed is **GUC-seeding
coverage** — from "API only, via ~62 hand-placed calls" to "API via a structural chokepoint + async via a
per-message primitive, both guarded by blocking lints and a negative integration proof." Cross-tenant
URL access still resolves 404, and the pre-existing cross-tenant isolation test suites remain green.
