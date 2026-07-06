# Milestone 7 — Tenant Lifecycle Kernel

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` §7 M7
> **Status:** All features (F7.1–F7.4) closed with evidence — awaiting milestone-validator (Phase 4). Main flips to PASS only on the validator's verdict.
> **Authored:** 2026-07-05 — *before any feature in this milestone began (F7.1 gate+ADR excepted — F7.1 is the gate itself and completed first per D7).*

> This file is a **spec**, authored up front. It says **what** M7 is, **which features** it contains,
> **what each feature implements**, and **what gets validated**. No execution steps — the "how" lives
> in each feature's `plan.md`. The close QA (`qa/milestone-qa.md`) validates M7 against *this* file and
> the binding `validation-contract.md` (D4).

## Objective

Turn tenant onboarding, export, and erasure from **absent** (tenants exist only via manual seed; no
export/erasure) into **designed, capability-gated, DB-enforced kernel capabilities** — resolving the
audit-immutability × GDPR-erasure tension by crypto-shredding (finding 15, canon T-003
`wiki/standards/backend-canon.md:181`). **Additionally** close the M6-carried tenancy-truth defect
(F7.4): the dev/CI DB role is SUPERUSER+BYPASSRLS so FORCE RLS is inert and tenant-isolation tests are
false-green — flip the role to prod-parity (NOSUPERUSER+NOBYPASSRLS+non-owner) so the existing
integration suite genuinely exercises RLS.

**Design of record:** ADR 0070 (`wiki/decisions/0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md`, Accepted 2026-07-05)
+ gate analysis (`../../../analysis/2026-07-05-m7-tenant-lifecycle-system-impact.md`, Yellow).

**Bars this milestone moves:**
- **Kernel gap closed** — a tenant is reachable end-to-end from an onboarding path (login → capability-gated action), no manual seed.
- **GDPR-erasure ⊥ audit-immutability resolved** — erasure shreds in-scope PII while the audit hash chain validates **GREEN** (criterion: post-erasure chain-validation test passes; audit rows unmutated).
- **Tenant-isolation tests become real** — the integration suite runs under a non-bypassing, non-owner role; every previously-false-green isolation test now genuinely exercises RLS (criterion: a sole-RLS query leaks under the wrong tenant / is blocked under the right one, for real).

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F7.1 | `f7.1-gate-and-adr` | **DONE.** developing-new-work system-impact analysis (Yellow) + ADR 0070 (Accepted) deciding erasure = crypto-shred + immutable skeleton, onboarding = API, F7.3 = implemented, orchestrator = iam service. | Gate artifact committed, verdict Green/Yellow (🟡 Yellow, no AS-1/2/3); ADR 0070 Accepted, cited by this milestone. **Both committed (14c631c6, d4a90bde).** |
| F7.2 | `f7.2-onboarding` | Contract-first `POST /tenants` → `iam` `OnboardTenant` service (one tx: insert `metaldocs.tenants`, seed first admin `iam_users` + role/capability grants, provision per-tenant crypto key, audit-record). New capability `tenant.onboard` (system_admin-class), tripwire arm generated via M2, registry bump. | New tenant created via the API is reachable **end-to-end**: login as its seeded admin → a capability-gated action succeeds; cross-tenant onboarding/access → 404. Contract regen clean; capability 10-touchpoint complete (registry size test bumped, tripwire negative proof). **Live QA drive** captures the onboarding→login→action path. |
| F7.3 | `f7.3-export-erasure` | Per-module `TenantDataPort` (`ExportTenantData`/`EraseTenantData`) on every module owning `tenant_id` tables + coverage test asserting each is registered; `iam` orchestrator fans out. **Export**: complete tenant-scoped artifact (row manifest + blob manifest under `tenants/{id}/…`). **Erasure**: crypto-shred the per-tenant key + erasure-tombstone audit event; audit rows never mutated. New caps `tenant.export`, `tenant.erase` (erase = system_admin-only). Async on River (M5) via outbox. **Implemented** (mission "implemented preferred", operator-elected). | Export produces a **complete** tenant-scoped artifact (every in-scope `tenant_id` table + tenant blobs represented; out-of-scope set explicitly enumerated in the contract). Erasure demonstrably renders in-scope PII unrecoverable (key destroyed) **while the audit hash-chain validation runs GREEN post-erasure** (integration test: erase → re-validate chain). Coverage test RED if a tenant-table-owning module lacks a port. Caps wired (10-touchpoint), H-PRE-1 off-tx in the fan-out. |
| F7.4 | `f7.4-rls-truth-sweep` | **DONE.** Dedicated non-owner `metaldocs_ci` role (NOSUPERUSER+NOBYPASSRLS, owns nothing, DML-only — deviation-from-literal-ownership-reassignment, HS-1) so the integration suite exercises RLS for real. `approval_signoffs` FORCE RLS + `tenant_isolation` policy on `actor_tenant_id` (0285). `SOLE-RLS-ASYNC-READ` lint (0 on tree; async reads complement M3's write rule). §4.5 negative+positive proof live-green under the non-owner role. Census = lint static coverage (0 RED) + live proof; no-GUC-0-rows reconciled to wrong-GUC-blocks (ratified M3 null-GUC-bypass idiom, HS-1). | CI role is NOSUPERUSER+NOBYPASSRLS+non-owner; the full integration suite runs **green under it** (every prior false-green isolation test now genuinely exercises RLS). **Negative+positive proof:** a sole-RLS query **leaks under the wrong tenant / is blocked under the right one**, for real, under the non-bypassing role. Every flip-surfaced query carries an explicit `tenant_id` predicate; the lint fails a synthetic new sole-RLS query. `approval_signoffs` gap closed. Prod-parity: CI role matches the prod deployment constraint. **Committed (9b3d7a8d spec+plan, 6e971e73 impl, a7a5a23f evidence). Two §4.2/§4.5 deviations surfaced for HS-1.** |

For each feature, "what to validate" is objectively checkable — a test passes, a route returns the
contracted shape, a build is clean, a runtime behavior is observed. No "works"/"looks right".

## Milestone validation definition

Close gate run by the **`milestone-validator` subagent** (separation of powers — it judges and writes
`qa/milestone-qa.md`; the main session flips status only on its PASS), per the binding C1–C7 checklist
(`.claude/skills/milestone/references/milestone-end-validation.md`). For M7:

1. **Per-feature acceptance** — every feature meets its "what to validate"; each feature's consumer
   contract (`spec.md`) honored (producer matches consumer). Checked **section-by-section against
   `validation-contract.md`** (D4) — any divergence is **HS-7**.
2. **Workflow-class QA** — backend-api checklist (onboarding route, contract regen), multi-tenant
   isolation checklist (F7.3 export/erasure no cross-tenant leak; F7.4 real RLS), DB-invariant
   checklist (audit append-only survives erasure; FORCE RLS active under non-owner role),
   authz checklist (3 new caps + generated tripwire arms), docs checklist (ADR 0070, wiki stamps).
3. **Regression** — M0–M6 gates still pass (esp. M2 tripwire generation/drift, M3 tenancy chokepoint,
   M4 versioning kernel, M5 River async, M6 review surfacer). F7.4's role flip must not regress any
   prior milestone's suite — it makes them *more* real, not broken.
4. **Root-cause check** — F7.4 fixes the **role** (correct-by-construction census), not individual
   symptom queries hand-hunted; erasure resolves the tension by crypto-shred, not by weakening audit
   immutability. Confirmed fixed, not symptom-patched (HS-2).
5. **No unplanned scope** — anything beyond F7.1–F7.4 recorded with rationale.
6. **Live QA drive** (runtime-visible milestone) — `.\scripts\start-api.ps1 -Build`, drive
   onboarding → login → capability-gated action, proof captured (D4 live-drive requirement).

## Dependencies & constraints

- **Depends on:** M2 (capability→tripwire-arm generation — 3 new caps use it), M3 (TxRunner GUC
  auto-seed + FORCE RLS chokepoint — F7.4 flips the role that makes it real), M5 (River async base —
  export/erasure jobs), M6 (F7.4 carried from it; the explicit-`tenant_id`-predicate idiom from F6.4).
- **Constraints respected:** contract-first (openapi.yaml + regen; never hand-add a route); every new
  tenant table carries `tenant_id` + FORCE RLS from birth; cross-tenant URL → 404 not 403; audit
  append-only grant + hash chain **stay intact** (erasure shreds keys, never mutates audit rows);
  async = transactional outbox on River (no inline network in handler tx); H-PRE-1 (no audit-recording
  read inside a lock-holding tx) holds in the erasure fan-out; capabilities-never-roles (ADR 0022);
  F7.4 role/ownership change is **expand/contract** (never break the live schema in one step);
  `.env` never read/printed/committed; PowerShell startup only; full integration suite NOT run locally
  (20-min box — targeted `-run` filters, bounded defers recorded); commits local, **never push**.

## Applicable hard-stops

- **HS-1** (milestone boundary) — after validator PASS, STOP; operator review gate; no M8, no merge without approval.
- **HS-2** (fix implies redesign outside boundary) — e.g. if per-tenant crypto-key management demands a change to the audit writer's tx model, or F7.4 ownership reassignment demands a schema-wide owner migration beyond the app role — stop, report boundary + minimum prerequisite, no symptom-patch.
- **HS-3** (prerequisite boundary fails) — build/runnable/auth-session/route/contract truth broken → repair first, rerun checkpoint, resume.
- **HS-4** (validator FAIL) — open the named fix feature, re-run lifecycle, re-dispatch validator.
- **HS-6** (scope drift / off-plan discovery) — e.g. more sole-RLS tables than the census predicted, or a tenant table with no clear owning module for its `TenantDataPort` — stop, surface, replan.
- **HS-7** (impl deviates from `validation-contract.md`) — fix the code to the contract, or re-open the contract WITH operator approval; never silently edit the contract to match code.
- **HS-8** (developing-new-work gate Red) — N/A, gate returned Yellow (F7.1 done).
