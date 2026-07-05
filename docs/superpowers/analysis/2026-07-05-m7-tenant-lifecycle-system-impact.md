# System-impact analysis — M7 Tenant Lifecycle Kernel

**Date:** 2026-07-05
**Intent (one line):** Tenant onboarding path (replace manual seed), tenant data export, tenant erasure (GDPR crypto-shredding vs immutable audit hash-chain skeleton), plus F7.4 RLS-truth sweep (flip CI DB role to NOSUPERUSER+NOBYPASSRLS+non-owner so tenant-isolation tests stop being false-green).
**Work type:** feature (milestone-scale: F7.1–F7.4; adds capabilities + one cross-cutting orchestration boundary, but births no new bounded-context module unless the ADR elects one)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

> Governing spec: `docs/superpowers/milestones/global-maximum-remediation/mission.md` §7 M7. Finding 15 + canon T-003 (`wiki/standards/backend-canon.md:181` "Data lifecycle": retention, soft/hard-delete, GDPR-class erasure, backups **with tested restore**). Runtime facts verified against code 2026-07-05 (fact-verification sweep, 5 anchors below).

---

## 1. Classify & own

- **Work type:** feature-class milestone. Three product features (onboarding, export, erasure) + one infra sweep (F7.4). No new bounded-context module is *required*; the ADR (F7.1) may elect a thin `tenantlifecycle` orchestrator — flagged below, decided there.
- **Owning module(s):**
  - **Onboarding (F7.2):** `iam` — owns `metaldocs.tenants` (bootstrap only today), `iam_users`, `iam_user_roles`, `iam_groups`, role/grant seeding. A tenant is an identity/access aggregate; onboarding = create tenant row + seed the first admin principal + grants. Natural home is an `iam` application service.
  - **Export/Erasure (F7.3):** **cross-cutting orchestration** over 34 `tenant_id` tables spread across ~9 modules (audit, iam, documents, templates, controlleddocuments, approval, notifications, render-fanout, taxonomy/tokens). No single module owns "all tenant data". Owner = a **tenant-lifecycle orchestrator** that fans out to a per-module `TenantDataPort` (export + erase). See §2 (global-maximum) and §6.
  - **Audit skeleton (F7.3):** `audit` — owns the immutability contract the erasure must not break; exposes how PII (in `payload`) is shredded while the hash-chain skeleton stays valid.
  - **F7.4 RLS sweep:** `internal/platform/db` + migrations + `tests/integration/testdb` — the CI/dev role attributes and table ownership are infra, not a module.
- **Explicitly NOT owning:**
  - `distribution` — training-acknowledgment / obligated-reader is out-of-module (per finding 15) and not tenant-lifecycle.
  - `search` — derived, rebuildable read model (canon §2.6); not authoritative tenant data, rebuilt post-erasure, not in erasure scope as a source.
  - `security` — provides crypto primitives (per-tenant key) but does not *own* the lifecycle; consumer of, not owner of, the shred key.
- **Cross-module edges (with direction):** `tenantlifecycle-orchestrator → {audit, iam, documents, templates, controlleddocuments, approval, notifications, render, taxonomy}` — each edge MUST go through that module's published `TenantDataPort` (export/erase for its own tables), **never** the orchestrator reaching into another module's tables/SQL (invariant 6). `iam-onboarding → authz` (seed grants), `→ audit` (record onboarding event).
- **Ambiguity?** Export/erasure owner is genuinely cross-cutting — but this is **resolvable by design**, not AS-3: the orchestrator + per-module port pattern is the only invariant-6-compatible shape. The ADR ratifies it. Recorded as a Yellow locked-constraint, not a hard-stop.

## 2. Foundation verdict

- **Base you'd build on:**
  - Tenancy chokepoint is **sound and recent** (M3, 2026-07-03): `TxRunner` auto-seeds tenant/actor GUCs; FORCE RLS on 33 tables with a uniform `tenant_isolation` policy; tenant-namespaced blob keys (`tenants/{tenant_id}/…`, `documents/application/keys.go:12`). This is the correct-by-construction base for both export (scan by tenant) and erasure (shred by tenant).
  - Audit immutability is **sound**: REVOKE UPDATE/DELETE/TRUNCATE (`0266_audit_events_hardening.sql:119`), real hash chain (`audit_event_row_hash` over prev_hash+row fields, serialized by `pg_advisory_xact_lock`, `writer.go:59`). PII is isolated to the `payload` jsonb column — the skeleton (id, actor_id, action, resource_type/id, tenant_id, hashes) is non-PII structural.
  - Onboarding base is **absent, not legacy**: no tenant-create API; `metaldocs.tenants` is bootstrap-SQL only. Building the path is net-new, not patching a workaround.
- **Sound, or legacy/patch/workaround?** Sound base. Two real defects to fix, neither a "patch to optimize inside":
  1. **F7.4 owner-bypass trap (real).** Dev/CI `metaldocs_app` = SUPERUSER+BYPASSRLS, so FORCE RLS is inert in dev → isolation tests false-green. **Additionally: table OWNER is not explicitly set** in the baseline (all `Owner: -`), and FORCE RLS never applies to the table owner. Flipping only NOSUPERUSER+NOBYPASSRLS is **insufficient** if the CI role still owns the tables — the role must be a **non-owner**. This is the exact trap the mission flags (§7 F7.4 "Trap"). Global-max fix = make the existing integration suite the census under a prod-faithful non-owner role.
  2. **`approval_signoffs` sole-trigger isolation (real, newly surfaced).** It carries `actor_tenant_id` (not `tenant_id`) and has **NO FORCE RLS / NO tenant_isolation policy** — isolation rests only on the `enforce_signoff_tenant_consistent()` trigger. Under the F7.4 role flip this is exactly the class the sweep exists to surface. **Global-maximum, not symptom-patch:** do not special-case it; the role flip + census must catch it, and it gets either a real RLS policy or a documented, lint-guarded exception (decided in the validation-contract).
- **Global-maximum structure for export/erasure:** a **per-module `TenantDataPort`** (each module exports/erases its own tenant-scoped rows behind its published interface) + a thin **orchestrator/saga** that fans out — NOT a god-query enumerating 34 tables from one place (that would violate invariant 6 and rot on every new tenant table). Trade-off: more interfaces to define now vs. correct-by-construction boundary + each module owns its own erasure semantics (esp. audit's crypto-shred). **No AS-2** — we are not optimizing inside a patch; we are building the kernel finding 15 says is missing.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes** | Onboarding, export, erasure are each capability-gated (new caps, §4). Erasure is high-privilege → `system_admin`-class + explicit cap; never "admin role can". Tripwire arms generated via M2. | `authz.Require(ctx,tx,cap,area)`; M2 registry→arm generation |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** (if API path) | New routes (onboarding, export request, erasure request) added to `api/openapi/v1/openapi.yaml` first, then regen per module `cfg.yaml`+`gen.go`. Runbook-only path (per ADR) touches no contract. | oapi-codegen; strict JSON decode |
| Multi-tenant pooled (`tenant_id` / GUC / 404) | **Yes** (core) | Export/erasure scope = the 34 `tenant_id` tables; every scan carries an explicit `tenant_id` predicate (correct-by-construction, as M6 F6.4 did). Cross-tenant erasure request → 404. GUCs via M3 chokepoint. **Erasure operates ON tenant data — must not itself leak cross-tenant.** | `tenant.FromContext`; `authz.SeedTxIdentity`; M3 TxRunner auto-seed |
| Async = transactional outbox | **Yes (likely)** | Export artifact build + erasure fan-out are long-running / external (blob writes) → enqueue in business tx, idempotent consumer, never inline network in handler tx. Reuse River (M5 base) for the job. | Outbox repo `render/fanout/staging_outbox.go:29`; River periodic/tx enqueue (M5) |
| DB enforces invariants (triggers/constraints) | **Yes** | Audit append-only grant stays (erasure MUST NOT UPDATE/DELETE audit rows — it shreds the per-tenant key, not the rows). F7.4: FORCE RLS becomes genuinely active under non-owner role. `approval_signoffs` gap closed at DB level or lint-guarded exception. | REVOKE grants; FORCE RLS policies; `ck_*` constraints |
| Cross-module via published interface only | **Yes (load-bearing)** | Orchestrator fans out to each module's `TenantDataPort` — never reaches into another module's tables. This is the whole reason the orchestrator pattern (not a god-query) is mandatory. | module `domain/port.go` / `application/ports.go` |

**AS-1?** None. The audit-immutability × GDPR-erasure tension is **resolved, not violated**, by crypto-shredding: PII (`payload` + any PII in a per-tenant encrypted envelope) becomes unrecoverable when the per-tenant key is destroyed, while the immutable hash-chain skeleton (and row_hash integrity) stays byte-intact → chain validation stays GREEN. The ADR (F7.1) makes this the explicit decided strategy.

## 4. Capability wiring

**Not N/A** — M7 adds capabilities. Registry size currently **35** (`TestCapabilityRegistrySize` const want=35, `model_test.go:96`, targeted-verified 2026-07-05). Provisional new caps (final names decided in ADR/brainstorming): `tenant.onboard`, `tenant.export`, `tenant.erase` (erasure is the highest-privilege — likely `system_admin`-held + explicit cap). Walk per cap:

1. **const + `validCapabilities`** — declare each in `iam/domain/model.go` consts + registry (`:90`/`:134`).
2. **scope classify** — `capability_scope.go:36`. Onboarding/export/erase are tenant-lifecycle-global → classify deliberately (likely `ScopeArea`/system-level, not per-area tenant scope; the ADR decides — a tenant-create cap logically pre-exists the tenant's areas).
3. **tier-1 route→cap** — map each new route in `permissions.go` (unmapped route = silent privilege escalation).
4. **tier-2 in-tx `authz.Require`** — after `SeedTxIdentity`, pattern `templates/application/create.go:67`.
5. **seed grants** — `db/reference-data/0001_product_reference_data.sql:17`; erasure to `system_admin` only.
6. **DB tripwire** — `ck_cap_format`/`ck_cap_not_legacy` accept new names; **M2 generation** derives the trigger arms from the registry (no hand-sync).
7. **guard tests green** — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`.
8. **bump `TestCapabilityRegistrySize`** — 35 → 35 + N (N = number of caps landed). Mandatory manual edit; ADR-justified.
9. **CI capability-coherence (REQ-AUTHZ-5)** — const/classify/tier-1/seed/test agree; M2 drift check green.
10. **H-PRE-1** — erasure/export record audit events; any `authz.Require` (audit-recording) stays **off** the audit-writer's advisory-lock tx. Erasure fan-out especially: never nest an authz-recording read inside a lock-holding shred tx.

## 5. Module wiring

**Conditionally N/A.** If the ADR elects a thin `tenantlifecycle` orchestrator module, walk the birth checklist (`module-wiring.md`): folders → domain + `port.go` → application + `ports.go` (consumes each module's `TenantDataPort`) → NO own tables (orchestrator holds no state beyond an export/erasure job row, which may live in `iam` or a `tenant_lifecycle_jobs` table) → delivery Handler + RegisterRoutes → api `cfg.yaml`+`gen.go` → OpenAPI tag → composition-root wiring → migration → `wiki/modules/tenantlifecycle.md` + `-tech-debt.md` + index. **Default recommendation:** host onboarding in `iam`; host export/erasure orchestration in a **small new `tenantlifecycle` package OR an `iam` application service** — decided in F7.1 ADR. If hosted in `iam`, §5 is N/A (no new module).

## 6. Frameworks to reuse, not reinvent

- `TxRunner` (`Do`/`DoReadOnly`, `runner.go:21`) — every export scan / erasure tx; nil-tx rejected.
- `tenant.FromContext` (`context.go:27`) — never hand-thread tenant id.
- `authz.SeedTxIdentity` (`context.go:48`) + `authz.Require` (`authz.go:76`) — gate every lifecycle op.
- `problem.New`/`Write` (`problem.go:77`) — RFC 9457 for all errors (cross-tenant → 404 problem, forbidden → 403).
- `audit.NewEvent`/`RecordTx` — onboarding, export, **and erasure** are auditable events (erasure especially must leave an audit skeleton entry recording *that* erasure happened — the tombstone).
- Outbox repo (`staging_outbox.go:29`) + **River (M5 base)** — export-artifact build & erasure fan-out as idempotent async jobs; no inline network in handler tx.
- Blob key namespacing (`documents/application/keys.go:12`, `tenants/{tenant_id}/…`) — export artifact written under the tenant prefix; erasure enumerates by prefix.
- Per-tenant crypto key — **check `internal/modules/security` first**; if no per-tenant key envelope exists, that is a genuinely new cross-cutting primitive → design it as a platform/security framework (the ADR's core decision), **not** an inline one-off.
- `testdb.Open` + factory (`tests/integration/testdb/`) — F7.4 role flip lives here; isolation tests reuse `SeedWithCaps`, `Qualified`.
- `strictjson` (documents-private today) — reuse/promote, don't re-inline.

## 7. Contract & data

- **OpenAPI-first:** if API path chosen — add `POST /tenants` (onboard), `POST /tenants/{id}/export` + artifact retrieval, `POST /tenants/{id}/erasure` (request/confirm) to `openapi.yaml`, new `tenant-lifecycle` tag, then regen. Runbook path (per ADR) = no contract change; operator-driven with the same application service underneath.
- **Migration:** possible `tenant_lifecycle_jobs` (or `tenant_erasure_requests`) table (`tenant_id`, status, requested_by, key-shred marker) — FORCE RLS + policy from birth. **Per-tenant encryption key** storage (security-owned). F7.4 migration: role attribute flip + table-ownership reassignment to a non-owner app role (expand/contract — never break live). `approval_signoffs`: add FORCE RLS + `tenant_isolation` policy on `actor_tenant_id`, OR an ADR-recorded lint-guarded exception.
- **Destructive change?** Erasure is *the* destructive operation — but by design it destroys the **key**, not audit rows (expand/contract not applicable to the shred itself; the audit skeleton is never mutated). F7.4 role flip is expand/contract: create non-owner role → reassign ownership → flip attributes → verify suite green, never a single breaking step.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory, `//go:build integration`, R1–R4 discipline. F7.4 makes the **existing suite the census** — no bespoke harness.
- **QA gates that apply:** **multi-tenant isolation** (core — export/erasure must not leak cross-tenant; F7.4 makes these real), **authz** (new caps + tripwire arms), **DB-invariant** (audit append-only survives erasure; FORCE RLS active under non-owner role; `approval_signoffs` closed), **async/idempotency** (export/erasure jobs idempotent), **contract** (if API path), **docs**. Not-touched gates marked N/A in the milestone.
- **F7.4 negative+positive proof (mandatory):** under NOSUPERUSER+NOBYPASSRLS+**non-owner** role — a sole-RLS tenant query **leaks (RED) under wrong tenant / blocked under right** for real; every query the flip surfaces RED gets an explicit `tenant_id` predicate; a lint blocks new sole-RLS tenant-scoped reads/writes.
- **Live QA drive (D4, runtime-visible):** `.\scripts\start-api.ps1 -Build`; drive **onboarding → login as the new tenant's admin → a capability-gated action**; capture proof. Erasure: drive export→erase→re-validate audit chain GREEN.
- **Evidence shape:** `go build ./...`, targeted `go test -run`, `go test -tags=integration` (targeted; full suite NOT run locally — 20-min box constraint, bounded defer recorded), `.\scripts\check-system-runnable.ps1`, live-drive proof. Commands + outcomes + QA disposition + bounded defers before closure.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/iam.md` (+ `-tech-debt.md`) for onboarding; `wiki/modules/audit.md` for the crypto-shred/erasure-tombstone contract; new `wiki/modules/tenantlifecycle.md` **iff** the ADR elects the module. Refresh `Last verified` stamps. `wiki/architecture/tenant-context.md` for the F7.4 role posture. `backend-canon.md:181` T-003 "Data lifecycle" self-flag → point to the shipped kernel.
- **REQ IDs cited:** tenancy/RLS REQs in `wiki/architecture/backend-target-architecture.md` (ADR 0022 Phase-5 §7, ADR 0027 async posture); audit immutability REQs; capability REQ-AUTHZ-5. (Exact IDs pinned in the validation-contract.)
- **ADR required? YES (mandated by D7).** New ADR: **Tenant Lifecycle & Erasure Strategy** — MUST explicitly decide crypto-shredding (per-tenant key destruction) as the GDPR-erasure answer, the immutable audit skeleton that survives, the export artifact shape, the onboarding path (API vs runbook), the export/erasure orchestration boundary (orchestrator + per-module `TenantDataPort`), and the F7.4 CI-role/ownership posture. Accepted **before** the milestone plan (F7.1). May reference/annotate ADR 0022 Phase-5 & ADR 0027 (tenancy); does not supersede them.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to `milestone`/brainstorming; ADR mandatory (D7) and carries the decided erasure strategy + orchestration boundary. No AS-1/AS-2/AS-3 unresolved.
- **Open hard-stops:** none. (Export/erasure ownership is cross-cutting but design-resolved via orchestrator + per-module ports — recorded as a locked constraint, not AS-3. Audit×GDPR tension resolved by crypto-shred — not AS-1.)
- **Locked constraints handed to brainstorming / the ADR:**
  1. **Erasure = crypto-shred the per-tenant key; NEVER UPDATE/DELETE audit rows.** Audit hash-chain skeleton stays byte-intact; chain validation stays GREEN post-erasure. PII target = `audit_events.payload` + any PII behind the per-tenant encryption envelope.
  2. **Export/erasure = orchestrator + per-module `TenantDataPort`**, never a god-query over 34 tables. Invariant-6 load-bearing.
  3. **New caps ⇒ bump `TestCapabilityRegistrySize` (35 → 35+N)** + generate tripwire arms via M2; erasure cap is `system_admin`-class.
  4. **F7.4: CI/dev role must be NOSUPERUSER + NOBYPASSRLS + NON-OWNER.** Owner-bypass trap: reassign table ownership away from the app role, else FORCE RLS stays inert. Existing integration suite = the census; every surfaced-RED query gets an explicit `tenant_id` predicate; lint blocks new sole-RLS tenant reads/writes.
  5. **`approval_signoffs` gap (newly surfaced):** carries `actor_tenant_id`, NO FORCE RLS, trigger-only isolation. Under the F7.4 flip it must either gain a real `tenant_isolation` policy or an ADR-recorded, lint-guarded exception. Global-max: no special-case; the census catches it.
  6. **Contract-first if API path; onboarding must reach login → capability-gated action end-to-end** (F7.2 acceptance) with live-drive proof.
  7. **Async lifecycle jobs on River (M5 base), idempotent, transactional-outbox** — no inline network in handler tx; H-PRE-1 off-tx for any audit-recording read in the erasure fan-out.
