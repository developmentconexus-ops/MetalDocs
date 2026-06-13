# ADR 0027 — RLS Adoption Sequencing + auth_identities Tenant-Global by Design

> **Status:** Accepted (executed in full by Wave Z, 2026-06-13) — originally Accepted 2026-06-11 (binding decision D-3 of the backend professionalization design spec). The three-tier *sequencing* below is now historical: Wave Z Z-2 (operator override of D-3) collapsed all three tiers into a single migration (`db/migrations/0237_rls_all_tenant_tables.sql`), enabling ENABLE+FORCE RLS + the NULL-permissive `tenant_isolation` policy on **all 27 remaining tenant-scoped tables** at once (Tier 2 iam_users + Tier 3 external-tenant tables included). The by-design `auth_identities` decision is unchanged.
> **Last verified:** 2026-06-13
>
> **Current reality (2026-06):** RLS is live on every tenant-scoped table (29 total = 2 from 0234 + 27 from 0237). The "first external tenant" / RF-6 triggers below never fired — Wave Z executed the full program ahead of them. `metaldocs.user_process_areas` is a VIEW over `public.user_process_areas` (the base table IS covered); views cannot carry RLS, so the census of 28 `tenant_id`-bearing relations minus that view = 27 base tables in 0237. NOSUPERUSER probe (Wave Z): GUC-unset→all rows, GUC=A→only A, GUC=B→only B, verified live on `iam_users` + `documents`.
> **Scope:** Two related decisions: (1) `auth_identities` has no `tenant_id` by deliberate design; (2) the sequencing and rationale for Row-Level Security adoption across the MetalDocs schema. Closes tech-debt item T-008 as by-design. Documents the partial-coverage RLS model and its trigger conditions.
> **Out of scope:** The two-tier authz model itself (ADR 0007); capability coherence (ADR 0022); the specific SQL for the Wave 2.3 migration (executed in item 2.3 using the `current_setting('metaldocs.tenant_id', true)` GUC pattern verified here).
> **Key files:**
> - `db/baseline/0001_current_schema.sql:914` — `auth_identities` table definition (no `tenant_id` column)
> - `internal/modules/iam/authz/context.go:60` — `SeedTxIdentity`: `set_config('metaldocs.tenant_id', $1, true)` + `set_config('metaldocs.actor_id', $2, true)` — the transaction-local GUC pattern reused by RLS policies
> - `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql:101` — `current_setting('metaldocs.asserted_caps', true)` — same GUC family; precedent for the pattern
> - `wiki/concepts/authz-tiers.md` — two-tier model; `metaldocs.tenant_id` GUC context
> - `wiki/backend/stage2-evaluation.md` — ADR-A paragraph (F-12 verdict row)

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
