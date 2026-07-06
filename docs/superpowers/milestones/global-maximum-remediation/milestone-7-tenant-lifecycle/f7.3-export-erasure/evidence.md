# Feature F7.3 — Evidence: Tenant export + crypto-shred erasure

> **Milestone:** 7 — Tenant Lifecycle Kernel · **Feature:** `f7.3-export-erasure` · **Closed:** 2026-07-05
> **Contract:** `spec.md` (consumer contract + Validation Gate) · `../validation-contract.md` §2, §3, §5.
> **Depth:** Implemented (operator-elected).

## What was implemented

By outcome (producer matches the consumer contract in `spec.md`):

1. **Crypto envelope (net-new)** — platform AES-256-GCM primitives + security-module `TenantCryptoService` (`tenant_keys`: `tenant_id uuid PK/FK, wrapped_dek bytea, created_at, destroyed_at`; KEK from `METALDOCS_TENANT_KEK`, never logged/printed). Published `TenantCrypto` port: provision / encrypt / decrypt / destroy. `d9d21719` (Task B).
2. **Audit payload envelope** — `RecordTx` seals payload under the event tenant's active DEK (`{"enc":"aesgcm.v1",...}`); no DEK / destroyed DEK → plaintext fall-through; reader returns redacted marker after shred. Hash chain hashes the stored string (chain semantics untouched). `d90181e5` (Task C).
3. **`TenantDataPort` fan-out** — interface + per-owning-module ports (own tables only, explicit `tenant_id` predicate every statement) + schema-census coverage test (anti-rot: new tenant table without a port → RED). `a9a5e1fc` (Task D).
4. **Export** (`POST /tenants/{tenant_id}/export`, cap `tenant.export`) → 202 `{job_id}`; jobs worker fans out `ExportTenantData`, writes one artifact to `tenants/{id}/exports/tenant-export-{job}.json` with row/blob manifests + completeness census header. `7ce40546` (Task E).
5. **Erasure** (`POST /tenants/{tenant_id}/erase`, cap `tenant.erase`, system_admin-only seed) → 202; jobs worker 3-phase: (1) per-port `EraseTenantData` fan-out under seeded erasure GUC + scheduler bypass; (2) blob delete; (3) `DestroyTenantKeyTx` crypto-shred + `tenants` tombstone (`name='erased', slug='erased-{id}'`, row never deleted = audit FK anchor) + `tenant.erased` audit event. `7ce40546` (Task E).
6. **Caps `tenant.export`/`tenant.erase`** (registry 36→38, ScopeTenant, catalog, tier-1/tier-2, seed grants, generated tripwire arm on `tenant_lifecycle_jobs`), lifecycle jobs table + FORCE RLS. `3ff94200` (Task A).
7. **Prerequisite + real-run repairs** — jobs binary real `TxRunner` wiring + lifecycle-ledger retention on erasure (`d26979fd`); tripwire 0283 class-defect fix (BEFORE DELETE trigger must `RETURN OLD`, not NULL — silent row-op cancellation, `af0b9e81`); api-lint conformance back to 0 (`48af6d6f`); first-real-run iam-suite test repairs (`49a9fed0`).
8. **Live-QA erasure-completeness fixes (this session)** — two defects the live erase drive surfaced, both closed in `e9a73057`:
   - **auth_identities survived erasure.** The auth `TenantDataPort` erased only `auth_sessions`; `auth_identities` (credential rows: identifier + password hash) has no `tenant_id` column and was declared out of scope. Fix: user_id-join `DELETE FROM metaldocs.auth_identities WHERE user_id IN (SELECT user_id FROM metaldocs.iam_users WHERE tenant_id=$1)`, run before the iam port deletes `iam_users` (auth precedes iam in `eraseOrder`). Not added to `Tables()` (census keys on `tenant_id` columns); still excluded from export (credential material never exported).
   - **`tenant.onboarded` audit payload landed PLAINTEXT despite crypto being wired.** Onboarding provisions the tenant key and records the audit event in ONE tx; the audit writer's key lookup read via the POOL, which cannot see the uncommitted `tenant_keys` row → `ErrKeyNotFound` → plaintext fall-through forever, leaving PII readable after crypto-shred. Fix: tx-aware seal chain (`WrappedDEKTx` / `EncryptForTenantTx` / `PayloadCrypto.EncryptForTenantTx`) threads the `*sql.Tx` so the key read hits the same open tx. Nil-tx path unchanged; the `ErrKeyNotFound`/`ErrKeyDestroyed`→plaintext fall-through (tombstone tenants, spec item 6.5) preserved.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Crypto primitives | `go test ./internal/platform/crypto/...` | ok (seal/open, wrap/unwrap, tamper-fail) | real (pure) |
| Security/audit/auth units | `go test ./internal/modules/security/... ./internal/modules/audit/... ./internal/modules/auth/...` | all `ok` | real |
| Tx-read key visibility (regression, unit) | `TestTenantCryptoService_EncryptForTenantTx_ReadsThroughTxOnly` | PASS — pool path misses tx-only key (`ErrKeyNotFound`), Tx path resolves + round-trips | real (sqlmock) |
| Static (build + vet) | `go build ./...`; `go vet ./...`; `go vet -tags=integration ./tests/...` | clean, 0 output | — |
| Full iam integration suite (real DB) | `go test -tags=integration -count=1 ./tests/integration/iam/...` **×2 back-to-back** | `ok 178.4s` then `ok 172.9s` — back-to-back green proves the deterministic-slug cleanup fix (no leak) | real (Postgres) |
| Audit integration suite (real DB) | `go test -tags=integration -count=1 ./tests/integration/audit/...` | `ok 173.8s` | real (Postgres) |
| Onboard payload sealed (fix b, integration) | `-run TestOnboardTenant_AuditPayloadSealedWhenCryptoWired` | `--- PASS` — stored `tenant.onboarded` payload is an envelope, not plaintext | real (Postgres) |
| Erasure chain + auth_identities (fix a, integration) | `-run TestTenantErasure_ChainStaysGreen` | `--- PASS` — seeds a real `auth_identities` row, asserts 0 survive post-erase; audit `ValidateIntegrity` green; skeleton rows intact | real (Postgres) |
| **Live QA drive** (HTTP → jobs worker → DB, rebuilt binaries w/ KEK) | onboard→export→erase on disposable tenant `4855dba9…` | see transcript below | real (running API :8081 + jobs worker) |

### Live QA drive transcript (2026-07-05, rebuilt binaries, `METALDOCS_TENANT_KEK` set)

1. `POST /api/v1/auth/login {admin}` → **200**, session cookie captured (Origin header; CSRF origin guard).
2. `POST /api/v1/tenants {slug:erase-livefix-61481}` → **201** `tenant_id=4855dba9-a613-478b-8076-f6869be69997`.
3. **Fix (b) proof, immediately post-onboard:** `audit_events` `tenant.onboarded` payload = `{"enc": "aesgcm.v1", …}`, `is_envelope=t`. (Pre-fix this landed plaintext `{"name":…,"slug":…,"admin_user_id":…}`.) `auth_identities`=1, `tenant_keys` present+live.
4. `POST /tenants/{id}/export` → **202** `{job_id:f0c32edd…}`; job polled → `status=ready`, `object_key` set; MinIO artifact `tenants/4855dba9…/exports/tenant-export-f0c32edd….json` present (12K).
5. `POST /tenants/{id}/erase` → **202** `{job_id:581a2f53…}`; job polled → `status=ready`, `requested_by='erased'` (ledger scrubbed).
6. **Erasure-completeness DB proofs (post-erase):**
   - **Fix (a):** `auth_identities` surviving for tenant's users = **0** (was 1).
   - Data erased: `iam_users=0, iam_user_roles=0, tenant_plans=0, governance_events=0, auth_sessions=0`.
   - **Crypto-shred:** `tenant_keys` `dek_len=0, destroyed=t`. `tenant.onboarded` payload **still an envelope (`t`)** but the DEK is destroyed → PII permanently undecryptable (crypto-shredded). `tenant.erased` payload plaintext ids-only (`f`) — the tombstone survives by design.
   - Skeleton retained: `tenants` row `name='erased', slug='erased-4855dba…'` (never deleted — audit FK anchor).
   - Ledger retained: both `export`+`erase` `tenant_lifecycle_jobs` rows present, `requested_by='erased'`.
   - MinIO export blob under the tenant prefix = **GONE** (phase-2 blob delete).

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Crypto primitives correct | yes | crypto unit tests |
| Registry 38 + scope + catalog guards green | yes | iam/domain tests (Task A `3ff94200`) |
| Contract regen + gates (oasdiff/redocly/shape) | yes | `48af6d6f` api-lint 0 violations |
| Coverage test: census fully ported; synthetic gap RED | yes | `TestTenantDataPortCoverage` (real Postgres schema) |
| Export completeness | yes | `TestTenantExport_CompleteArtifact` + live artifact w/ manifests |
| Erasure: PII unrecoverable + chain GREEN + rows intact | yes | `TestTenantErasure_ChainStaysGreen` + live crypto-shred proof (dek_len=0, envelope undecryptable, skeleton intact) |
| Erasure isolation | yes | `TestTenantErasure_DoesNotTouchOtherTenant` (dev tenant `ffffffff…` untouched across drive) |
| Tripwire arm negative | yes | Task A live SQL + M2 parity/drift lints 0 |
| Async correctness | yes | enqueue via outbox in handler tx; worker idempotent; live 202→ready both jobs |
| Live QA drive | yes | transcript above (onboard→export→erase, rebuilt binaries) |
| **auth_identities erased** (live-found) | yes | Fix (a): surviving=0 |
| **onboarded payload sealed** (live-found) | yes | Fix (b): is_envelope=t live + integration regression |

## Review disposition

- **Spec-compliance:** the two live-QA findings were genuine contract gaps against spec §3 ("post-erasure, in-scope PII unrecoverable") and §2 audit-payload sealing — not scope creep. Both fixed at root (auth port erase step; tx-aware key read), not patched around. Onboarding tx ordering left unchanged (sealing moved inside the tx-aware path, not the event build).
- **Code-quality (subagent reviewer, `caveman:cavecrew-reviewer`):** confirmed join-delete ordering vs `eraseOrder` correct, idempotent on retry (0-row second pass), tx-visibility read genuinely hits the tx, plaintext fall-through preserved. One yellow: the integration test's `runAuditCryptoAdapter` mapped ALL errors to plaintext instead of only `ErrKeyNotFound`/`ErrKeyDestroyed` — fixed to match the composition-root adapters (real DB errors now propagate).
- **Orchestrator inline fixes (disclosed):** test-hygiene root cause of a real-run flake — `cleanupTenant` did a plain `DELETE FROM tenants`, but a crypto-wired onboard test's `tenant_keys` FK-child blocked it (error-discarded) → leaked the deterministic slug → next-run collision. Fixed to delete the FK-child `tenant_keys` before the tenant. Verified by back-to-back suite runs.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Pre-F7.3 plaintext audit rows not backfill-encrypted | Backfill would mutate immutable hash-chained rows (forbidden). Only tenants onboarded with a DEK (post-F7.3) get shreddable payloads. | Documented limitation (spec interview #2b); revisit only with an audit re-baseline |
| KEK rotation / re-wrap tooling | Contract §6.4 declared defer; single KEK from env sufficient for v1 | Product decision |
| RLS role flip (metaldocs_app NOSUPERUSER+NOBYPASSRLS) | Belongs to F7.4 (RLS-truth sweep) — erasure tests currently run under a superuser role so FORCE RLS is inert in dev | F7.4 |
