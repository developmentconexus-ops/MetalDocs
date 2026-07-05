# Feature F7.3 — Plan: Tenant export + crypto-shred erasure

> **Spec:** `./spec.md` (approved). Execution: subagent-driven (sonnet impl+review, main orchestrates/reviews/commits). TDD per task.

## Task graph

```
A (caps+arm+jobs-table migration) ──┐
B (crypto: platform + security)  ───┼──> E (contract+handlers+orchestrator+River) ──> F (integration proofs + live drive)
D (TenantDataPorts + coverage)   ───┘
C (audit payload envelope) — after B, parallel with D/E-start
```

## Tasks

### Task A — capabilities + tenant_lifecycle_jobs + tripwire arm (sonnet)
- `iam/domain`: `CapTenantExport`, `CapTenantErase` + validCapabilities (RED 36≠38 first), `capability_scope.go` ScopeTenant, `catalog.go` pt-BR, `model_test.go` 36→38.
- `permissions.go`: tier-1 rules `POST /api/v1/tenants/{tenantId}/export|erase` (pathPrefix/regex per existing param-route idiom), VisibilityPermissionGuarded.
- `0001_product_reference_data.sql`: system_admin grants for both caps (erase = system_admin ONLY — no other role).
- Migration `0278_tenant_lifecycle_jobs.sql`: table per spec item 7 + `erased_at timestamptz NULL` column on `metaldocs.tenants` + FORCE RLS + tenant policy on new table (copy idiom from latest RLS-bearing migration).
- `internal/platform/tripwire/arms.go`: arm #20 `tenant_lifecycle_jobs`/INSERT caps `[tenant.export, tenant.erase]`; render.go WHEN branch; regen via `cmd/gen-tripwire` → migration `0279_...`; arms_test 19→20, golden path; api-lint path const. Apply 0278+0279 to dev DB (container psql).
- Gates: `go build ./...`, iam domain tests, TRIPWIRE lints, live P0001 negative SQL proof.

### Task B — crypto framework (sonnet)
- `internal/platform/crypto/envelope.go`: AES-256-GCM `Seal/Open`, DEK `Wrap/Unwrap` under KEK, envelope JSON `{"enc":"aesgcm.v1","data":...}` encode/parse. Unit tests first (round-trip, tamper, wrong-key, destroyed).
- Migration `0280_tenant_keys.sql` (security module owns): table per spec + FORCE RLS + policy. Apply to dev.
- `internal/modules/security/`: `TenantCrypto` published port + postgres adapter (Provision/Encrypt/Decrypt/DestroyTx; wrapped_dek zeroed + destroyed_at on destroy; decrypt after destroy → typed `ErrKeyDestroyed`). KEK from `METALDOCS_TENANT_KEK` env (config plumbing, `.env.example` entry, never logged). Missing/short KEK → fail-fast at boot ONLY if feature used; nil-safe wiring like F7.2.
- Wire real `TenantKeyProvisioner` adapter into F7.2 seam (main.go composition root).
- Gates: unit tests green, build, onboarding still live-green (provisioning row appears).

### Task C — audit payload envelope (sonnet, after B)
- Audit writer `RecordTx`: inject optional `TenantCrypto`; if active DEK for event tenant → store envelope JSON as payload; else plaintext. Chain untouched.
- Audit reader (list/export paths): decrypt envelopes when DEK alive; `ErrKeyDestroyed` → return redacted marker `{"redacted":"crypto-shredded"}`.
- Composition-root wiring only (no audit→security internal import).
- Tests: integration — record under DEK tenant → payload column is envelope, read returns plaintext; post-destroy read returns redacted; chain validator green on mixed plaintext/ciphertext history.

### Task D — TenantDataPort + per-module impls + coverage test (sonnet; may split D1/D2)
- Interface + `TableExport` types (iam domain or `internal/platform/tenantdata` if import cycles bite — implementer decides, records why).
- Ports per owning module covering the full tenant-table census (anchor at repo truth; audit port serves governance_events read-only + audit_events skeleton export, deletes NOTHING). Every SQL statement carries explicit `tenant_id = $n` predicate.
- `TestTenantDataPortCoverage`: schema census vs registered `Tables()` union + allowlist; synthetic-gap negative.
- Registration slice at composition root (shared by api main + jobs main).

### Task E — contract + handlers + orchestrator + River worker (sonnet, after A+B+D)
- OpenAPI: `POST /tenants/{tenantId}/export`, `POST /tenants/{tenantId}/erase` (202 `{job_id}`, 401/403/404/409 problems, tag iam, Portuguese summaries) + regen; oasdiff/redocly/shape-lint clean. HS-7: contract authored from spec, never bent to code.
- `iam/application/tenant_lifecycle_service.go`: enqueue (Require cap in-tx → 404 unknown tenant → 409 erased → INSERT job + outbox event same tx) and run-side orchestration (export fan-out+artifact; erasure order per spec item 6, tombstone UPDATE, H-PRE-1 audit event off-tx).
- Handlers in `iam/delivery/http/tenant_handler.go` (extend), router wiring, main.go + metaldocs-jobs main.go (River worker registration per fanout_worker precedent).
- Artifact writer via VerifiedStore (`ListTenantObjects` manifest, artifact under tenant prefix).
- Gates: build, targeted unit tests, contract gates.

### Task F — integration proofs + live QA drive (sonnet test author; live drive = main session)
- `TestTenantExport_CompleteArtifact`, `TestTenantErasure_ChainStaysGreen`, `TestTenantErasure_DoesNotTouchOtherTenant` (canonical testdb framework; compile-verified, runs deferred per box constraint → bounded defer).
- Live drive (main): rebuild, onboard disposable tenant, seed some data, export → artifact proof; erase → DB proofs (rows gone, blobs gone, key destroyed, tombstone, audit skeleton + chain GREEN via validator/SQL).
- evidence.md.

## Review discipline

Orchestrator reviews every task diff pre-commit (F7.2 caught 4 defects this way). Commit per task-group. Never push.
