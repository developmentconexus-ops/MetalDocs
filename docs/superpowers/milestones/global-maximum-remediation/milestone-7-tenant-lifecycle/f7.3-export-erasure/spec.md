# Feature F7.3 — Spec: Tenant export + crypto-shred erasure

> **Milestone:** 7 — Tenant Lifecycle Kernel · **Folder:** `f7.3-export-erasure`
> **Status:** Approved (pre-code) 2026-07-05
> **Design of record:** ADR 0070 (Accepted) decisions 3–5 (export artifact, crypto-shred, iam orchestrator + per-module ports) + `../validation-contract.md` §2, §3, §5 (binding, HS-7). Depth = **Implemented** (operator-elected, F7.1 interview).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | All shape decisions? | Locked in F7.1 (ADR 0070, 5 operator decisions) + contract §2/§3. No new questions. |
| 2 | Runtime-truth deltas found at planning | (a) `governance_events` has **no owning Go module** (trigger-written) → its `TenantDataPort` rows are served by the **audit module's port** (closest semantic owner — event records; same read-only shape). Recorded here, enforced by coverage test. (b) `audit_events.payload` is **plaintext today**; no crypto primitive exists anywhere → DEK/KEK envelope is net-new (ADR 0070 d4 anticipated). Shreddability of audit payloads holds for tenants onboarded **with a DEK** (post-F7.3); pre-existing plaintext rows are a **documented bounded limitation** (backfill would mutate immutable rows — forbidden). |

## Consumer contract (FIRST)

- **Consumers:**
  1. **Operator tooling** — `POST /api/v1/tenants/{tenantId}/export` (cap `tenant.export`) → **202** `{job_id}`; artifact lands under the tenant blob prefix. `POST /api/v1/tenants/{tenantId}/erase` (cap `tenant.erase`, **system_admin-only** seed) → **202** `{job_id}`. Unknown tenant → **404**; missing cap → 403; unauthenticated → 401; already-erased tenant → 409. RFC 9457.
  2. **An auditor** — export artifact must be self-evidencing: row manifest (every in-scope table, per-table count), blob manifest (key + content hash), **completeness header** (full census with included/empty/out-of-scope status per table).
  3. **A data subject (GDPR)** — post-erasure, in-scope PII unrecoverable; audit chain still validates GREEN.
  4. **F7.4 / future modules** — `TenantDataPort` coverage test goes RED when a module gains a `tenant_id` table without registering a port.
- **Source of truth:** `api/openapi/v1/openapi.yaml` (authored here: `exportTenant`, `eraseTenant`, tag `iam`) + contract §2/§3.

## What this feature implements

1. **Crypto envelope (net-new).** Platform primitives `internal/platform/crypto/` (AES-256-GCM seal/open, wrap/unwrap — pure, unit-tested). Security-module service (`internal/modules/security/`): `tenant_keys` table (`tenant_id uuid PK/FK, wrapped_dek bytea, created_at, destroyed_at timestamptz NULL` — tenant-scoped ⇒ `FORCE RLS` + tenant policy from birth, milestone constraint). KEK from env (`METALDOCS_TENANT_KEK`, 32-byte; **never logged/printed**; `.env.example` documented). Published port `TenantCrypto`: `ProvisionTenantKeyTx(ctx, tx, tenantID)`, `EncryptForTenant(ctx, tenantID, plaintext) → envelope`, `DecryptForTenant(ctx, tenantID, envelope)`, `DestroyTenantKeyTx(ctx, tx, tenantID)` (sets `destroyed_at`, zeroes `wrapped_dek`). Destroyed key ⇒ decrypt fails forever = crypto-shred. F7.2's `TenantKeyProvisioner` no-op seam is replaced by the real adapter (wired composition root; iam never imports security internals).
2. **Audit payload envelope.** Audit writer (`audit/infrastructure/postgres/writer.go RecordTx`): when the event's tenant has an **active DEK**, `payload` is stored as envelope JSON `{"enc":"aesgcm.v1","data":"<b64>"}`; no DEK → plaintext (legacy behavior, unchanged rows). Hash chain hashes whatever string is written — chain semantics untouched. Audit **reader** decrypts envelopes for display while the DEK is alive; after shred it returns the envelope marker (redacted). Crypto port injected at composition root (audit → security published port only).
3. **`TenantDataPort` (invariant-6-safe fan-out).** Interface (hosted `internal/modules/iam/domain/tenantlifecycle.go` or sibling): `Module() string`, `Tables() []string`, `ExportTenantData(ctx, tenantID) ([]TableExport, error)` (`TableExport{Table string, Rows []json.RawMessage}`), `EraseTenantData(ctx, tx, tenantID) (map[string]int64, error)` (rows deleted per table; **explicit `tenant_id` predicate in every statement** — M6 F6.4 idiom). Implementations per owning module, own tables only: documents, controlleddocuments, templates, iam, auth, notifications*, render, jobs (idempotency_keys), audit (audit_export_jobs; + read-only export of `audit_events` skeleton view + `governance_events` per interview #2a; **EraseTenantData on audit_events shreds nothing row-wise — key destruction does the work; it deletes NOTHING**), taxonomy/tokens (token_dictionary_entries), distribution/tenant_plans per actual ownership (implementers anchor at repo truth; the coverage test is the arbiter). Registration at composition root into the orchestrator.
4. **Coverage test (anti-rot).** Integration test queries the schema census (every table with a `tenant_id`/`actor_tenant_id` column, both schemas) and asserts the union of all registered ports' `Tables()` + the explicit out-of-scope allowlist (`tenants` root, `tenant_keys` handled by security shred step, `outbox_events` non-tenant) covers it exactly. New tenant table without a port ⇒ RED.
5. **Export (cap `tenant.export`).** `POST /tenants/{tenantId}/export` handler: tier-1 cap, in-tx `authz.Require`, 404 on unknown tenant, INSERT `tenant_lifecycle_jobs` row (kind `export`) + outbox event in same tx → **202**. River worker (metaldocs-jobs): fan-out `ExportTenantData` across ports, blob manifest via `VerifiedStore.ListTenantObjects` (key + content hash), completeness header (census status per table), one artifact JSON written to `tenants/{id}/exports/tenant-export-{job}.json`, job row → `ready`. Idempotent (job-id keyed; re-run overwrites same key).
6. **Erasure (cap `tenant.erase`, system_admin-only seed).** `POST /tenants/{tenantId}/erase` → 202 (same enqueue shape, kind `erase`; 409 if tenant already erased). River worker order: (1) fan-out `EraseTenantData` per port (each in seeded-GUC tx, explicit predicates); (2) delete blobs under `tenants/{id}/…`; (3) `DestroyTenantKeyTx` (crypto-shred — audit payload ciphertext now unrecoverable); (4) tombstone `metaldocs.tenants` row: UPDATE `name='erased', slug='erased-{tenantId}', erased_at=now()` (new nullable column; row **never deleted** — audit FK anchor); (5) **erasure-tombstone audit event** (`tenant.erased`, actor + tenant + timestamp) recorded **after** shred, **off any lock-holding tx** (H-PRE-1); it is written plaintext (its own DEK is gone — by design the tombstone survives). Audit rows: **never UPDATE/DELETE** (0 mutations — asserted in test).
7. **`tenant_lifecycle_jobs` table** (new migration): `id, tenant_id, kind ('export'|'erase'), status, requested_by, object_key, error, created_at, completed_at` + `tenant_id` + FORCE RLS + policy; **tripwire arm** (M2 registry → gen-tripwire → next migration): INSERT requires `tenant.export` OR `tenant.erase` (arm cap-array). Registry 19 → 20 arms.
8. **Capabilities `tenant.export` + `tenant.erase`** — full 10-touchpoint walk each (contract §5): consts+registry (**36 → 38**), ScopeTenant, catalog pt-BR, tier-1 rules, tier-2 Requires, seed grants (`tenant.export` → system_admin; `tenant.erase` → system_admin ONLY), generated arm (item 7), guard tests, REQ-AUTHZ-5, H-PRE-1.

## Non-goals (mandatory)

- No KEK rotation / re-wrap tooling (contract §6.4 declared defer).
- No backfill-encryption of pre-F7.3 plaintext audit payloads (would mutate immutable rows — forbidden; documented limitation).
- No UI; no scheduled/automatic erasure; no partial (per-user) erasure — tenant-granularity only.
- No restore/import of an export artifact.
- No RLS role flip (F7.4).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof | Real vs fixture |
|---|---|---|
| Crypto primitives correct (seal/open, wrap/unwrap, tamper fails) | `internal/platform/crypto` unit tests | real (pure) |
| Registry 38 + scope + catalog guards green | `go test ./internal/modules/iam/domain/` (RED 36≠38 first) | real |
| Contract regen + gates (oasdiff, redocly, shape lint) | commands, 0 violations | real |
| Coverage test: census fully ported; synthetic gap goes RED | `TestTenantDataPortCoverage` (+ deliberately-unregistered negative) | real (Postgres schema) |
| Export completeness | `TestTenantExport_CompleteArtifact` — seed tenant across ≥3 modules → export → row manifest covers every in-scope table with rows; blob manifest lists tenant blobs; completeness header covers full census | real |
| Erasure: PII unrecoverable + chain GREEN + rows intact | `TestTenantErasure_ChainStaysGreen` — onboard (DEK) → audit events with known payload → erase → `ValidateIntegrity` **0 issues**; payload decrypt **fails**; `count(*)` of tenant audit rows unchanged; 0 UPDATEs (xmin/row compare or before/after byte-equality on skeleton) | real |
| Erasure isolation | `TestTenantErasure_DoesNotTouchOtherTenant` — tenant B rows byte-identical post-erasure of A | real |
| Tripwire arm negative | `tenant_lifecycle_jobs` INSERT without asserted cap → P0001; M2 parity/drift lints 0 | real (live SQL, F7.2 precedent) |
| Async correctness | job enqueued via outbox in handler tx (no inline network); worker idempotent (re-run same job id safe) | real |
| **Live QA drive** | export a live tenant → 202 → artifact appears under tenant prefix; erase a disposable tenant → 202 → DB proof (rows gone, tombstone, audit skeleton intact, chain validator green) | real |

TDD: failing test first per task. Live-drive + evidence.md per contract §6.

## ADR needed?

- [x] ADR 0070 (Accepted) covers all durable decisions. Interview #2a/#2b recorded here + evidence; no new ADR (no MUST-deviation).
