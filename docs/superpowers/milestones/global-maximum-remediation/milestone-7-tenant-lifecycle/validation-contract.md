# M7 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M7 (tenant lifecycle: onboarding, export,
> crypto-shred erasure + F7.4 RLS-truth sweep)
> **Authored:** 2026-07-05, **before any F7.2+ implementation** (mission D4). Committed before the
> first code change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7).
> **Design rails locked before this contract (D7):** gate 🟡 Yellow
> (`../../../analysis/2026-07-05-m7-tenant-lifecycle-system-impact.md`, 14c631c6) + **ADR 0070**
> (`wiki/decisions/0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md`, Accepted,
> d4a90bde). This contract is the concrete enumeration ADR 0070's decisions imply.
>
> **Load-bearing clauses:** the **§0.1 34-table tenant inventory**, the **§3.2 in/out-of-erasure-scope
> table**, the **§3.3 crypto-shred mechanism + §3.4 audit-chain-stays-GREEN proof**, the **§4 F7.4
> role/ownership/census + negative-proof**, and the **§5 three-capability 10-touchpoint table
> (registry 35→38, arms generated via M2)**.

---

## 0. Runtime-truth basis (facts this contract is built on)

All claims traced to source 2026-07-05 (targeted-verify sweep). Runtime truth beats docs (CLAUDE.md).
If any anchor has moved at implementation time, the code wins and this section is re-stamped (HS-7 if
it changes an acceptance bar).

### 0.1 Tenant data inventory — the 34 `tenant_id` tables (export scope = ALL; erasure scope = §3.2)

**`metaldocs` schema (15):** `audit_events` (tenant_id **text**), `audit_export_jobs`, `auth_sessions`,
`document_process_areas`, `document_profiles`, `iam_group_members`, `iam_groups`, `iam_user_roles`,
`iam_users`, `idempotency_keys`, `materialize_dispatch_outbox`, `notifications`, `pdf_dispatch_outbox`,
`tenant_plans`, `token_dictionary_entries`.

**`public` schema (19):** `approval_instances`, `approval_routes`, `approval_signoffs`
(**`actor_tenant_id`**, no `tenant_id`), `cd_sequence_counters`, `controlled_document_area_grants`,
`controlled_document_user_grants`, `controlled_documents`, `document_comments`, `document_exports`,
`document_placeholder_values`, `documents`, `editor_sessions`, `governance_events`,
`template_audit_log`, `templates`, `templates_audit_log`, `templates_template`,
`templates_template_version`, `user_process_areas`.

Plus the tenant root row: `metaldocs.tenants` (`id`, `name`, `slug`, `created_at`, `updated_at`,
`db/baseline/0001_current_schema.sql:1601`). Blob namespace: `tenants/{tenant_id}/…`
(`internal/modules/documents/application/keys.go:12,19`).

> **Contract rule:** the F7.3 export coverage test enumerates these 34 tables (+ tenants row + blob
> prefix) as the completeness census. If a migration adds a 35th `tenant_id` table before M7 closes,
> the census test must catch it (owning module must register its `TenantDataPort` — §2.2). A new
> tenant table with no port = coverage test RED (HS-6 if owner ambiguous).

### 0.2 FORCE RLS census — 33 of 34 covered; `approval_signoffs` is the gap

33 tables carry `FORCE ROW LEVEL SECURITY` + an identical `tenant_isolation` policy:
`USING (NULLIF(current_setting('metaldocs.tenant_id', true),'') IS NULL OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true),''))`
(exemplar `db/baseline/0001_current_schema.sql:4558`). **`public.approval_signoffs` has NO FORCE RLS
and NO policy** — isolation rests solely on the `enforce_signoff_tenant_consistent()` trigger
(`db/baseline/0001_current_schema.sql:3956`). This is the F7.4 gap (§4.4).

### 0.3 Audit immutability (erasure must not break this)

- Append-only grant: `REVOKE UPDATE, DELETE, TRUNCATE ON metaldocs.audit_events FROM metaldocs_app`
  (`0266_audit_events_hardening.sql:119`).
- Hash chain: `row_hash = metaldocs.audit_event_row_hash(prev_hash, …)`, serialized by
  `pg_advisory_xact_lock(90120260513004)` (`writer.go:59`).
- DDL columns (`0001_current_schema.sql:898`): `id`, `occurred_at`, `actor_id`, `action`,
  `resource_type`, `resource_id`, `payload jsonb`, `trace_id`, `tenant_id`, `audit_sequence`,
  `prev_hash`, `row_hash`. **PII surface = `payload` (jsonb, ≤65536B cap).** **Skeleton (non-PII,
  survives erasure) = `id`, `occurred_at`, `actor_id`, `action`, `resource_type`, `resource_id`,
  `trace_id`, `tenant_id`, `audit_sequence`, `prev_hash`, `row_hash`.**

### 0.4 DB role facts (F7.4 basis)

Dev/CI `metaldocs_app` = **SUPERUSER + BYPASSRLS** → FORCE RLS inert in dev (documented in
`templates/repository/tenant_id_rls_integration_test.go:29-32`,
`documents/repository/review_due_reader_integration_test.go:157-159`). Prod = NOSUPERUSER+NOBYPASSRLS
(ADR 0022 Phase-5 / ADR 0027 — prod safe). Table ownership is **not explicitly set** in baseline
(all `Owner: -`) → **owner-bypass trap**: FORCE RLS never applies to the owner, so the CI role must be
a **non-owner** (§4.2).

### 0.5 Onboarding today

No programmatic tenant creation. `metaldocs.tenants` populated only by bootstrap SQL / dev-seed. No
`POST /tenants` in `api/openapi/v1/openapi.yaml`. Capability registry size = **35**
(`iam/domain/model_test.go:96`).

---

## 1. F7.2 — Onboarding contract

### 1.1 Route (contract-first)
- **New:** `POST /tenants` added to `api/openapi/v1/openapi.yaml` under a `tenants` (or
  `tenant-lifecycle`) tag; regenerated via the owning module `cfg.yaml`+`gen.go`. **No hand-added Go
  route.** Request: tenant `name` + `slug` + first-admin identity (email/display name). Response:
  created tenant id + admin principal id (201) or RFC 9457 problem (409 slug conflict, 403 missing cap).
- Tier-1: route→capability `tenant.onboard` mapped in `permissions.go` (unmapped = silent escalation).

### 1.2 Service (`iam` `OnboardTenant`, one `TxRunner.Do` tx)
In a single business tx: (a) insert `metaldocs.tenants`; (b) seed the first admin `iam_users` +
`iam_user_roles` grants for the tenant-admin capability set; (c) provision the per-tenant crypto key
(§3.3 envelope) — key material never logged; (d) `audit.RecordTx` an `tenant.onboarded` event.
Tier-2 `authz.Require(ctx, tx, tenant.onboard, area)` after `SeedTxIdentity`. H-PRE-1: the audit record
stays off any lock-holding sub-tx.

### 1.3 End-state (acceptance — the F7.2 bar)
A tenant created via `POST /tenants` is reachable **end-to-end**:
1. its seeded admin can **authenticate** (login issues a session for the new tenant);
2. that admin can perform **at least one capability-gated action** (e.g. create a document / list a
   tenant-scoped resource) that returns the contracted shape;
3. **cross-tenant isolation holds** — the new tenant cannot see another tenant's data; a cross-tenant
   URL returns **404** (not 403).
This is proven by an integration test **and** the mandatory live QA drive (§6.2).

---

## 2. F7.3 — Export contract

### 2.1 Artifact shape
`ExportTenantData(tenantID)` orchestrated across all per-module ports produces **one complete
tenant-scoped artifact**:
- a **row manifest**: for every in-scope `tenant_id` table (§0.1), the tenant's rows serialized
  (JSON), grouped by module/table, with a per-table row count;
- a **blob manifest**: every object under `tenants/{tenantID}/…` (documents, revisions, exports)
  listed with key + content hash;
- a **completeness header**: the 34-table census with per-table "included / empty / out-of-scope"
  status, so an auditor can see nothing was silently dropped.
Long-running/external → built as an idempotent River job (ADR 0067) via the transactional outbox;
never inline in the handler tx. Capability `tenant.export`. Written back under the tenant blob prefix.

### 2.2 Per-module `TenantDataPort` + coverage test (the anti-rot guard)
Every module owning ≥1 `tenant_id` table publishes a `TenantDataPort` with `ExportTenantData` and
`EraseTenantData` (own tables only — invariant 6). A **coverage test** asserts every module owning a
`tenant_id` table (derived from the §0.1 census / a schema query) has a registered port; **RED if a
tenant-table-owning module lacks one.** The `iam` orchestrator calls each registered port.

### 2.3 Export acceptance (the F7.3 export bar)
Export of a seeded tenant yields an artifact whose row manifest covers **every in-scope table that has
rows**, whose blob manifest covers every tenant blob, and whose completeness header accounts for all 34
tables. Proven by an integration test that seeds a tenant across multiple modules, exports, and asserts
the manifest is complete (no in-scope table missing).

---

## 3. F7.3 — Erasure contract

### 3.1 Mechanism summary
Erasure = **destroy the per-tenant crypto key**, not delete audit rows. PII stored as ciphertext
behind the key becomes unrecoverable; the audit skeleton + hash chain stay byte-intact. An
**erasure-tombstone** audit event records that erasure occurred. Capability `tenant.erase`
(**system_admin-only**). Async River job via outbox; H-PRE-1 off-tx for any audit-recording read in
the fan-out.

### 3.2 In / out of erasure scope (EXACT — load-bearing)

| Table / asset | Erasure disposition | Why |
|---|---|---|
| `documents`, `document_comments`, `document_placeholder_values`, `document_exports`, `document_process_areas`, `document_profiles`, `controlled_documents`, `controlled_document_area_grants`, `controlled_document_user_grants`, `cd_sequence_counters`, `templates`, `templates_template`, `templates_template_version`, `template_audit_log`, `templates_audit_log`, `editor_sessions`, `document_placeholder_values`, `token_dictionary_entries`, `user_process_areas` | **IN — crypto-shred PII + delete rows** (business content; PII fields shredded via key destruction, rows removed by the owning module's `EraseTenantData`) | Tenant business/authoring data; no immutability grant. |
| `iam_users`, `iam_user_roles`, `iam_groups`, `iam_group_members` | **IN — crypto-shred PII (name/email) + delete rows** | Personal identity data — core GDPR subject. |
| `auth_sessions`, `idempotency_keys`, `notifications`, `pdf_dispatch_outbox`, `materialize_dispatch_outbox`, `audit_export_jobs` | **IN — delete rows** (operational/ephemeral; TTL'd) | No long-term PII value; hard-delete is safe and simplest. |
| `approval_instances`, `approval_routes`, `approval_signoffs`, `governance_events` | **IN — crypto-shred PII in payload + delete rows**, EXCEPT signoff attributions that are audit-equivalent (see note) | Approval records may carry attributable PII; shred + delete unless a regulated-retention rule pins them (flag in evidence if so). |
| **`audit_events`** | **IN (PII only) — crypto-shred `payload`; rows NEVER updated/deleted; skeleton stays** | Immutable append-only grant (§0.3). Erasure = payload PII unrecoverable via key destruction; chain stays GREEN. |
| **`metaldocs.tenants` root row** | **OUT of delete — mark erased (tombstone); NOT deleted** | Referenced by `audit_events.tenant_id`; deleting it would orphan the immutable skeleton. Tombstone = erased-at marker + shredded name/slug. |
| **Blobs** `tenants/{id}/…` | **IN — delete objects** (or shred if envelope-encrypted at rest) | Tenant document bytes; removed from the object store. |
| **`tenant_plans`** | **IN — delete row** | Billing/plan association; no retained PII. |

> **Out-of-erasure-scope, stated explicitly:** the audit **skeleton** columns (§0.3), the tenants
> tombstone row, and the immutable hash-chain linkage. Anything not in the 34-table census + tenants
> row + blob prefix is out of tenant-erasure scope by construction (nothing else is tenant-scoped).

### 3.3 Crypto-shred mechanism
- A **per-tenant data key (DEK)** wrapped by a service KEK (`internal/modules/security` — reuse an
  existing key-envelope primitive if present; else build it as a platform/security framework, ADR 0070
  decision 4). Provisioned at onboarding (§1.2c).
- PII-at-rest that must survive-but-be-shreddable (notably `audit_events.payload`) is stored encrypted
  under the tenant DEK.
- Erasure destroys the tenant DEK (and its wrapped copy) → all ciphertext under it is unrecoverable.
- The erasure-tombstone audit event is written **after** shred, recording actor + tenant + timestamp.

### 3.4 Audit-chain-stays-GREEN proof (the F7.3 erasure bar — non-negotiable)
An integration test: seed a tenant with audit history → run erasure → **re-run the audit hash-chain
validation** (the janitor/validator path) → assert **GREEN** (no broken links; every `row_hash`
recomputes; no row was UPDATE/DELETE'd). Additionally assert: the tenant's PII (e.g. a known payload
value) is **unrecoverable** post-shred (decrypt fails / returns ciphertext). Erasure that mutates any
audit row, or that leaves PII recoverable, is a **FAIL**.

### 3.5 Erasure isolation
Erasure of tenant A must not touch tenant B's data (every `EraseTenantData` carries the explicit
`tenant_id` predicate; runs under the tenant GUC). Cross-tenant erasure request → 404.

---

## 4. F7.4 — RLS-truth sweep contract

### 4.1 Role attributes
The dev/CI app DB role used by the integration suite is **NOSUPERUSER + NOBYPASSRLS**. (Prod already
is; this makes CI prod-faithful.)

### 4.2 Non-owner requirement (owner-bypass trap)
The CI role must **not own** the tenant tables — FORCE RLS never applies to a table's owner. Table
ownership is reassigned to a distinct owner role (e.g. a `metaldocs_owner`/`postgres` owner) while the
app role connects as a **non-owner** with only DML grants. Verified: `SELECT` on a tenant table under
the app role WITHOUT a tenant GUC returns **0 rows** (policy active), not all rows (bypass).

### 4.3 Census + predicate fixes (correct-by-construction, not grep-hunt)
Run the existing integration suite under the non-bypassing non-owner role. Every tenant-scoped query
that relied **solely** on RLS flips its isolation assertion **RED**. Each is fixed with an **explicit
`tenant_id` predicate** (the M6 F6.4 idiom). The suite is the census — the set of surfaced queries is
enumerated in F7.4 `evidence.md`. A **lint** (AST/grep guard, following the M2/M3 lint pattern) blocks
new tenant-scoped reads/writes that lack an explicit tenant predicate; negative proof: it fails a
synthetic sole-RLS query.

### 4.4 `approval_signoffs` disposition
Add **FORCE ROW LEVEL SECURITY + a `tenant_isolation` policy on `actor_tenant_id`** (parity with the
other 33). **If** the signoff semantics genuinely require cross-tenant co-sign (making a strict policy
wrong), record the exception in an ADR 0070 amendment + a lint carve-out — **no silent special-case**
(HS-2 if this forces a redesign). Default expectation: the policy is added.

### 4.5 Negative + positive proof (the F7.4 bar — non-negotiable)
Under the NOSUPERUSER+NOBYPASSRLS+non-owner role, an integration test proves, **for real**: a
tenant-scoped query **leaks (returns another tenant's row) when run under the wrong tenant GUC / and is
blocked (0 rows) under the right tenant GUC** — i.e. RLS is genuinely active, not bypassed. The full
targeted integration suite runs **green** under the new role. Prior false-green isolation tests now
genuinely exercise RLS.

---

## 5. Capability wiring (3 new caps — registry 35 → 38)

| # touchpoint | `tenant.onboard` | `tenant.export` | `tenant.erase` |
|---|---|---|---|
| 1 const + `validCapabilities` | declare in `iam/domain/model.go` | ″ | ″ |
| 2 scope classify | `capability_scope.go` — system/lifecycle scope (decided in impl; not per-area) | ″ | ″ |
| 3 tier-1 route→cap | `permissions.go` `POST /tenants` | export route | erase route |
| 4 tier-2 `authz.Require` in-tx | in `OnboardTenant` | in export orchestrator | in erase orchestrator |
| 5 seed grants | tenant-admin/system_admin | system_admin (+ ops?) | **system_admin ONLY** |
| 6 DB tripwire arm | **generated from registry via M2** (no hand-sync) | ″ | ″ |
| 7 guard tests green | `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` | ″ | ″ |
| 8 `TestCapabilityRegistrySize` | **35 → 38** (one manual edit, ADR 0070-gated) | (same bump) | (same bump) |
| 9 CI capability-coherence (REQ-AUTHZ-5) | const/classify/tier-1/seed/test agree | ″ | ″ |
| 10 H-PRE-1 off-tx | audit record off lock-tx | ″ | erase fan-out audit off lock-tx |

**Negative tripwire proof:** adding a synthetic asserted capability without a generated arm fails CI
(M2 drift check), not runtime — re-confirmed for the M7 caps.

---

## 6. Gates, live drive, evidence shape

### 6.1 QA gates that apply
authz (5-surface coherence + tripwire arms), contract (openapi regen clean; oasdiff/shape lints green),
multi-tenant isolation (F7.3 no cross-tenant leak; F7.4 real RLS negative+positive), DB-invariant
(audit append-only survives erasure; FORCE RLS active under non-owner role; `approval_signoffs`
closed), docs (ADR 0070 cited; `wiki/modules/iam.md`, `audit.md`, `tenant-context.md` updated + `Last
verified` re-stamped).

### 6.2 Live QA drive (mandatory — runtime-visible milestone, D4)
`.\scripts\start-api.ps1 -Build`; drive **`POST /tenants` → login as the new tenant's admin → a
capability-gated action**; capture request/response + logs as proof. Erasure/export driven via
integration test + (if wired to a route) a live drive; the audit-chain-GREEN post-erasure proof (§3.4)
captured.

### 6.3 Evidence shape (per feature, CLAUDE.md "evidence before closure")
Commands + real output: `go build ./...`; targeted `go test -run` + `go test -tags=integration -run`
(full suite NOT run locally — 20-min box; targeted filters; bounded defers recorded with triggers);
`.\scripts\check-system-runnable.ps1`; contract regen diff; capability-coherence lint; the F7.4
negative+positive RLS proof; the §3.4 audit-green proof; live-drive capture. Fixture-vs-real labeled.
Review/QA disposition recorded. No bare "done".

### 6.4 Bounded defers (declared up front)
- Full integration suite is not run locally end-to-end (box constraint) — F7.4 asserts the **targeted**
  isolation + tenancy suites green under the new role; a full-suite run is a bounded defer with the
  trigger "CI executes it under the flipped role".
- Per-tenant crypto-key **rotation** (beyond provision + destroy) is out of scope — trigger: a
  key-rotation policy milestone. M7 provisions at onboarding and destroys at erasure only.
