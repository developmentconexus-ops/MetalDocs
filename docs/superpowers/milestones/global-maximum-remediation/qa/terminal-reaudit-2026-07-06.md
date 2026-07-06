# Terminal Re-Audit — Global Maximum Remediation

**Date:** 2026-07-06  
**Method:** 10 independent adversarial sonnet reviewers (one per dimension), fresh agents with no shared context; findings synthesized by main session. Same 10 dimension charters as the 2026-07-03 baseline review (`docs/superpowers/analysis/2026-07-03-final-architecture-global-maximum-review.md`).  
**Verdict vocabulary:** CONFIRMED (global maximum, prior finding closed) / DEBT (finding partially or fully open) / RE-LITIGATE (design questionable) / NEW ISSUE (not in original review)  
**Purpose:** §8 Terminal Acceptance input for `mission-validator`.

---

## Scorecard

| # | Dimension | 2026-07-03 | 2026-07-06 | Delta |
|---|-----------|-----------|-----------|-------|
| 1 | Module structure & boundaries | CONFIRMED/DEBT | **CONFIRMED** | ✓ All 3 findings closed |
| 2 | Authorization (ADR 0022) | CONFIRMED | **CONFIRMED** | ✓ Tripwire generation + divergences closed |
| 3 | Contract-first API | DEBT | **CONFIRMED** | ✓ All 4 governance gaps closed + CI |
| 4 | Multi-tenancy | DEBT | **CONFIRMED** | ✓ SeedTxTenant + CI role + lifecycle |
| 5 | Async architecture | DEBT | **CONFIRMED** | ✓ River-only; retention; H-PRE-1 LIVE |
| 6 | Versioning kernel | DEBT | **CONFIRMED** | ✓ Unified state fn; race test; eQMS features |
| 7 | DB invariant enforcer | CONFIRMED/DEBT | **CONFIRMED** | ✓ Tripwire generated + structural CI |
| 8 | Testing & QA | CONFIRMED/DEBT | **DEBT** | − Discipline gate RED; 2 defers = absent features |
| 9 | Observability & ops | Yellow | **DEBT** | − /metrics not port-isolated (same port as API) |
| 10 | Decision governance | CONFIRMED/DEBT | **DEBT** | − ADR status rule no CI gate |

**Summary: 7 CONFIRMED, 3 DEBT. No new RE-LITIGATE. No new dimension RED.**

---

## Dimension detail

### 1. Module structure & boundaries — CONFIRMED

All three 2026-07-03 findings independently closed from current repo state:

- **CLAUDE.md inventory** (CLOSED): Line 34 lists exactly 14 modules matching `ls internal/modules/` — no `docs`, `tokens` present, approval nested-exception explicitly called out with ADR 0072 citation.
- **repository/ layer split** (CLOSED): `find internal/modules -type d -name repository` → zero results. `documents/infrastructure` and `templates/infrastructure` both exist. Rename fully complete.
- **documents/approval visibility** (CLOSED): `wiki/decisions/0072-approval-nested-exception-and-boundary-model.md` is substantive (~160 lines) with real reasoning: quantified import-edge inventory, import-cycle resolution narrative, negative-plant/revert proof, empty debt table, alternatives considered. `scripts/check-module-boundaries.ps1` exit code 0; allow-model is non-vacuous (explicit layer whitelist, empty debtAllowList, approval identity folded with correct external-surface enforcement). One test-only import (`approval/infrastructure` → `iam/infrastructure/postgres`) noted but guard explicitly scopes to non-test files — not a violation.

No new issues introduced.

### 2. Authorization (ADR 0022) — CONFIRMED

- **Tripwire generation** (CLOSED): `internal/platform/tripwire/arms.go:57` defines `TripwireArms` (Go slice, capability consts). `internal/platform/tripwire/render.go:18` `RenderMigration()` derives SQL byte-for-byte. `cmd/gen-tripwire/main.go` is the CLI. Latest migration `0283_tripwire_delete_return_old.sql` + `0279_tenant_lifecycle_jobs_tripwire_export_erase_cap.sql` carry "machine-generated from `internal/platform/tripwire`" provenance headers. No hand-typed TEXT[] literal independent of registry.
- **CI drift gate** (CLOSED): `scripts/api-lint/tripwire_arm_rules.go` implements `TRIPWIRE-ARM-PARITY` (byte-equal regenerated migration check) + `TRIPWIRE-ARM-DRIFT` (AST-walk — asserted cap must be an arm member). Wired blocking in `.github/workflows/api-contract.yml:99-100` (`-strict`, no continue-on-error). Structural gate, same class as `check-module-boundaries.ps1`.
- **Capability divergences** (CLOSED): `forceReleaseDocumentSession` was a planning-doc shorthand — actual code uses `CapMembershipManage` consistently at both tiers; regression pin in `permissions_test.go:690-709`. `approval-route-management` → `CapRouteManage` consistent both tiers, same pin.
- **TestCapabilityRegistrySize** (CONFIRMED): Asserts 38 capabilities, PASSES live (`go test ./internal/modules/iam/... -run TestCapabilityRegistrySize -count=1 → PASS`).

No new issues introduced.

### 3. Contract-first API — CONFIRMED

All four 2026-07-03 governance gaps closed with blocking CI:

- **oasdiff breaking-change gate** (CLOSED): `.github/workflows/openapi-breaking.yml` triggers on PR for `api/openapi/v1/openapi.yaml`, runs `oasdiff breaking --fail-on ERR`, no continue-on-error. Genuinely blocking.
- **Nullable⇒required shape lint** (CLOSED): `scripts/api-lint/shape_rules.go:60-180` rule `SHAPE-NULLABLE-NOT-REQUIRED` — flags nullable-not-required, explicitly tied to 9f86828b bug class. Wired blocking in `api-contract.yml:77-100`.
- **contract-sync CI gate** (CLOSED): `api-contract.yml:102-124` job `contract-sync` runs `check-contract-sync-all.ps1`, no continue-on-error. Approval deferred per F9.5 ADR 0072 — documented bounded defer, not a gap.
- **Redocly struct rule** (CLOSED): `redocly.yaml:10` → `struct: error` (not disabled). Zero errors, zero suppressions (burned down from 133 in commit `7517082c`).
- **No M9 contract drift** (CONFIRMED): Last commit touching `api/openapi/` is pre-M9 (`48af6d6f`). M9 commits touch zero files under `api/openapi/`.

### 4. Multi-tenancy — CONFIRMED

All five items verified from current code:

- **SeedTxTenant in async binaries** (CLOSED): Real call sites confirmed in worker (`pdf_job_runner.go:161`, `materialize_job_runner.go:85`, `fanout_worker.go:64`, `staging_outbox.go:146`) and jobs (`stuck_instance_watchdog/job.go:174`, `document_review_surfacer/job.go:158`, `scheduler_service.go:62`) binaries.
- **RLS policy + compensating backstop** (CLOSED): Policy still deliberately NULL-permissive (by design, per ADR 0027 amendment); compensating backstop is structural — `SeedTxTenant` at all async write roots + two blocking api-lint rules (`ASYNC-TENANT-SEED`, `SOLE-RLS-ASYNC-READ`). Sanctioned unseeded surfaces explicitly enumerated.
- **CI role NOSUPERUSER/NOBYPASSRLS** (CLOSED): `db/migrations/0284_ci_rls_role.sql:46` creates `metaldocs_ci` as `NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN`, 0 owned tables. `tests/integration/security/rls_truth_test.go:41` self-verifies role properties at runtime — not just migration text.
- **Tenant lifecycle** (CLOSED): Real implementations in `iam/application/tenant_lifecycle_service.go`, `onboard_tenant_service.go`, `iam/infrastructure/postgres/tenant_data_port.go`. Crypto-shred: `internal/modules/security` DEK/KEK framework. `TestTenantErasure_ChainStaysGreen` and `onboard_tenant_e2e_test.go` are real integration tests.
- **M9 migration cleanliness** (CONFIRMED): `git diff --name-only e93ea5b6..HEAD -- db/migrations` → empty. M9 commits are docs/governance only.

### 5. Async architecture — CONFIRMED

- **River-only** (CLOSED): `db/migrations/0273_drop_job_leases.sql` applies and DROPs `metaldocs.job_leases` + `acquire_lease`/`heartbeat_lease`/`release_lease`/`assert_lease_epoch`. No live Postgres-lease scheduler code found. `.claude/worktrees/` stale agent worktrees are gitignored, not main-branch code. Minor doc-drift: `db/baseline/0001_current_schema.sql` still shows dropped objects (predates 0273) — cosmetic, not runtime.
- **River periodic jobs** (CLOSED, with precision note): `internal/modules/jobs/maintenance/periodic.go` registers 4 jobs (stuck-instance-watchdog, idempotency-janitor, audit-integrity-validator, document-review-surfacer). The 5th job (outbox retention) lives in `internal/modules/render/fanout/retention/periodic.go:26-34` and is appended at binary entry points (`metaldocs-api/main.go:663`, `metaldocs-jobs/main.go:146`). Runtime behavior correct; "all 5 in periodic.go" claim is a documentation imprecision, not a functional gap.
- **Outbox retention** (CLOSED): `staging_outbox.go:202-224` `PurgeDispatched` — real bounded-batch DELETE, 7-day retention, 5000-row batches, 10 max iterations; wired in `metaldocs-jobs`.
- **Fanout race test** (CLOSED): `fanout_worker_race_integration_test.go` — genuine concurrent goroutines, closed-channel barrier, real Postgres via testdb; pass confirmed via M5 QA log (not independently re-executed — no live DB in this environment).
- **H-PRE-1 LIVE** (CLOSED): ADR 0067 carries an explicit dated erratum correcting the "retired" false claim; invariant-checklist.md:61-68 explicitly states H-PRE-1 LIVE with the M5 disambiguation. No current doc claims retirement.
- **M9 commits** (CONFIRMED): M9 GMR commits touch only `docs/superpowers/milestones/global-maximum-remediation/` — zero async runtime code.

### 6. Versioning kernel — CONFIRMED

- **Unified state machine** (CLOSED): `internal/modules/documents/domain/state.go:21-49` `CanTransitionDocumentStatus` — one exhaustive switch. Note: actual count is 8 statuses (not 9; `rejected` deliberately removed in migration 0272 / commit `52188636`, `archived` is a soft flag). "9-status" label in prior milestone docs is imprecise but design is intentional and documented. `TestCanTransitionDocumentStatus` + `TestCanTransitionDocumentStatus_DBTriggerParity` both PASS.
- **Publish race test** (CLOSED): `publish_race_integration_test.go` — two real goroutines, closed-channel barrier, 4 subtests, real Postgres via testdb; SKIPs correctly without DB (no fake pass).
- **Periodic review/expiry** (CLOSED): `db/migrations/0274_document_review_and_reason.sql` write path; `document_review_surfacer/job.go` worklist consumer confirmed active (commit `400714ab`).
- **Structured reason-for-change** (CLOSED, additive by design): `ReasonCategory` (5 values, DB CHECK constraint in migration 0274:58-60) added alongside `RevisionTitle` — additive, not replacement. Matches ISO/Part-11 pattern (structured enum + free-text detail).
- **ADR 0066** (CLOSED): `wiki/decisions/0066-optimistic-concurrency-transport-split.md` Accepted, ratifies If-Match (documents) / lock_version (templates) split as intentional. Still live in code.
- **M9 clean** (CONFIRMED): No M9 commits touch migrations or documents/domain.

### 7. DB as invariant enforcer — CONFIRMED

- **Tripwire arms generated** (CLOSED): `arms.go:57` → `render.go:18 RenderMigration()` → `cmd/gen-tripwire/main.go` → committed migrations with "machine-generated" provenance headers. No hand-typed TEXT[] independent of registry.
- **CI drift gate structural** (CLOSED): `scripts/api-lint/tripwire_arm_rules.go` implements `TRIPWIRE-ARM-PARITY` (byte-equal regen check) + `TRIPWIRE-ARM-DRIFT` (AST walk). Blocking in CI. Same class as module-boundaries guard — structural, not pinned-value.
- **Audit hash-chain window** (STILL OPEN — named DEBT, correctly tracked): `internal/modules/audit/infrastructure/postgres/writer.go:69` `auditIntegrityValidationWindow = 10000` unchanged. Correctly named in `wiki/modules/audit-tech-debt.md:108` T-013 with explicit rationale. Not silently dropped.
- **Forward-only migrations** (CLOSED): No down-migration files; convention unchanged.

### 8. Testing & QA — DEBT

Tooling is real (not theater), but two concrete deficiencies remain:

- **Traceability gate** (CLOSED as tool): `go run ./scripts/req-trace` genuinely exits 1 on uncovered MUSTs; gate is real. **However:** REQ-SEARCH-1 ("reindex procedure exists and tested") has zero reindex code anywhere in the repo — no implementation at all, not just untested. REQ-SEC-3 ("OWASP ASVS review checklist") has zero ASVS references outside the requirement text. These are **absent features honestly disclosed**, not "coverage gaps." Erratum E1's "governance defer" framing is accurate but understates that these are unimplemented MUSTs.
- **check-test-discipline.sh RED at HEAD** (NEW ISSUE): 4 violations at HEAD: `sequence_test.go:57,123` (R1), `job_integration_test.go:186` (R4), `tenant_id_rls_integration_test.go:148` (R2). These predate M9 and were not introduced by it. M9's governance-hygiene closure did not surface that this CI gate is currently failing.
- **Legacy test policy** (CLOSED): `wiki/quality/legacy-test-policy.md` — concrete repair-vs-delete decision tree with 4 named trigger classes and 3 real worked examples. Not hand-wavy.
- **t.Parallel expansion** (CLOSED): Sampled files use per-test sqlmock instances or static source — no shared mutable state. Integration-tier revert (concurrent DROP DATABASE burst) honestly documented, not rubber-stamped.
- **REQ-AUTHN-1 defer** (CLOSED as honest): Runtime uses bcrypt cost-12 (`auth/application/service.go:38-39,175,229,1025`), not Argon2id. Defer is accurate, not a dodge.

**Why DEBT:** check-test-discipline.sh failure at HEAD is a concrete CI gate regression (pre-existing, not M9-introduced, but unresolved at program close); 2 of 4 E1 defers name absent features, not just unenforced coverage.

### 9. Observability & ops — DEBT

5 of 6 original Yellow blockers genuinely resolved:

- **Dockerfiles** (CLOSED): `deploy/docker/{api,worker,jobs}.Dockerfile` exist; compose references resolve; DEPLOY.md declares Compose as v1 target, K8s archived.
- **Redis GCRA limiter** (CLOSED): `internal/platform/ratelimit/redis_store.go:18-28` GCRA via `go-redis/redis_rate/v10`; compose wires `METALDOCS_RATELIMIT_STORE=redis`; ADR 0071; live 2-replica QA proof in M8 evidence.
- **/metrics format** (CLOSED): `internal/platform/observability/prometheus.go` — genuine Prometheus exposition format, confirmed live.
- **X-Trace-Id correlation** (CLOSED): `internal/platform/observability/http.go` sets header from resolved trace id; live triple match (header/log/span) in M8 QA.
- **Backup/restore runbook** (CLOSED): `wiki/runbooks/backup-restore.md` — multi-section runbook with roles/prereqs/restore-warnings/RPO-RTO; one real execution evidence (dump 535KB, restore 70 relations, row-count validation 127==127).
- **/metrics port isolation** (STILL OPEN — NEW ISSUE): `apps/api/cmd/metaldocs-api/main.go:851-873` mounts `/metrics` on the same `server.Addr` as the public API (same port, `APP_PORT`). Docker Compose publishes that port directly to host. The only isolation is nginx not proxying `/metrics`, but the API container port is independently published. If `APP_PORT` is exposed beyond localhost in any prod deploy, `/metrics` is publicly scrapeable with zero authentication — not port-isolated.

**Why DEBT (not Yellow):** More improved than the original Yellow, but one specific claim from M8 ("infra-only port") was inaccurate. Real pre-production gap before first paying customer.

**Recommendation:** bind `/metrics` to a separate listener, or explicitly document the firewall/reverse-proxy mitigation requirement in DEPLOY.md (currently unstated).

### 10. Decision governance — DEBT

- **ADR 0022 mega-ADR** (CLOSED): `wiki/decisions/0022-authz-capability-coherence.md:3-4` status block is 2 lines / ~180 chars. Execution history relocated to `wiki/decisions/0022-execution-history.md` (134 lines, all 13 phases, zero information loss).
- **ADR 0013 Superseded stamp** (CLOSED WITH DISPOSITION): `wiki/decisions/0013-template-revision-labels.md:3` now reads `Accepted (amended by 0052 — version-creation trigger; REV labels + persisted revision_number unchanged)`. F9.1 runtime research correctly determined 0013 was never actually superseded — 0052 only amends the version-creation trigger; 0013's mechanism is still live. Canonical vocabulary `Accepted (amended by NNNN)` is correct here.
- **Going-forward enforcement** (STILL OPEN): All 65 ADR status fields today comply with the ≤3-line/≤400-char rule (sweep confirmed 0 violations). But `documentation-governance.md:33-34` explicitly states CI wiring of this sweep is "optional future extension, not required." No CI job exists enforcing this. A new mega-ADR can appear tomorrow with nothing structural preventing it.
- **ADR body substance** (CLOSED): M9 commits only touched status/header fields on ADRs — no decision substance rewritten.
- **Last-verified stamps** (CLOSED): 17 files stamped 2026-07-06; 4 independently spot-checked anchors all exact (`infrastructure/postgres.go:88`, `mappers.go:18`, `capability_service.go:31/48`, `authz/context.go:49/62`).

**Why DEBT:** Convention established, one-time cleanup done, but no structural gate prevents regressing. The whole mission's cross-cutting theme is "convention without a gate is not global-maximum" — this dimension ends with exactly that residual for its own rule.

**Low-cost fix:** wire the existing sweep one-liner into `governance-check.yml` CI job.

---

## Out-of-scope findings confirmed excluded

The three out-of-scope findings from the 2026-07-03 review remain excluded-by-decision with triggers intact:
- **Training acknowledgment** (distribution module, not documents module) — trigger: eQMS Phase 2 scope decision
- **C4 diagrams fragmentary** — trigger: pre-v1 documentation investment decision
- **Threat model / SLO-capacity targets** — trigger: named as backlog items, acceptable pre-v1

No evidence these were re-introduced or addressed by the mission work.

---

## CI gates installed by this mission

| Gate | Installed by | Status at HEAD |
|------|-------------|----------------|
| `openapi-breaking.yml` — oasdiff breaking-change | M1 F1.1 | GREEN |
| `api-contract.yml` — nullable-not-required lint | M1 F1.2 | GREEN |
| `api-contract.yml` — contract-sync | M1 F1.3 | GREEN |
| `.github/workflows/module-boundaries.yml` | M0/pre-mission | GREEN |
| `api-contract.yml` — TRIPWIRE-ARM-PARITY/DRIFT lints | M2 | GREEN |
| `api-contract.yml` — ASYNC-TENANT-SEED/SOLE-RLS-ASYNC-READ lints | M3 | GREEN |
| `.github/workflows/req-traceability.yml` — REQ-ID gate | M9 F9.2 | GREEN (exits 1 on drift; anti-rot PASS) |
| `check-test-discipline.sh` — test framework gate | Pre-mission | **RED** (4 pre-existing violations at HEAD) |

---

## New issues surfaced (not in 2026-07-03 review)

| # | Issue | Severity | Dim |
|---|-------|----------|-----|
| N1 | `/metrics` on same port as public API — not network-isolated, auth-bypass only | Pre-production | 9 |
| N2 | `check-test-discipline.sh` RED at HEAD — 4 pre-existing violations | Pre-production | 8 |
| N3 | `db/baseline/0001_current_schema.sql` stale — still shows dropped `job_leases`/`acquire_lease` | Low / cosmetic | 5 |
| N4 | ADR status-field sweep has no CI gate | Low | 10 |

---

## Cross-cutting — the original meta-defect class

The 2026-07-03 review identified 7 hand-synced conventions without generators or gates. Status at program close:

| # | Convention | Status |
|---|-----------|--------|
| 1 | Tripwire cap-arms vs Go registry | **GATE INSTALLED** (M2) |
| 2 | SeedTxTenant at ~85 sites | **GATE INSTALLED** (M3 ASYNC-TENANT-SEED lint) |
| 3 | oasdiff / nullable lint / contract-sync CI | **GATE INSTALLED** (M1) |
| 4 | Nullable-not-required shape lint | **GATE INSTALLED** (M1) |
| 5 | ESLint FE feature-boundary rule | **GATE INSTALLED** (M1) |
| 6 | REQ-ID traceability automation | **GATE INSTALLED** (M9 F9.2; 4 defers named+ledgered) |
| 7 | Wiki file:line anchor CI checker | **NOT INSTALLED** (stamps/curator loop only; no CI anchor checker) |

6 of 7 hand-synced conventions now have structural gates. Item 7 (wiki anchor CI checker) was noted in the 2026-07-03 review as DEBT and remains so — no mission feature targeted it.

---

## Bottom line (for mission-validator)

The mission substantially achieved its intent: 7 of 10 dimensions are now genuinely CONFIRMED (up from 0 CONFIRMED, 5 DEBT, 2 CONFIRMED/DEBT, 1 Yellow, 0 Red at baseline). The core architecture decisions remain at or near global maximum. The 6 principal hand-sync debt items all received structural gates.

Three dimensions remain DEBT:
- **Dim 8 (Testing/QA):** check-test-discipline.sh gate RED at HEAD from pre-existing violations (not introduced by mission); 2 of 4 §8 defers are absent features (REQ-SEARCH-1, REQ-SEC-3) — honestly disclosed per Erratum E1.
- **Dim 9 (Observability):** `/metrics` rides the same port as the public API — only nginx routing prevents public exposure in prod; DEPLOY.md does not document required firewall/reverse-proxy mitigation.
- **Dim 10 (Governance):** ADR status-field rule has no CI gate — convention established, one-time cleanup done, structural enforcement deferred.

Whether these three DEBTs satisfy mission.md §8(a) ("CONFIRMED on every in-scope dimension") is the mission-validator's judgment call.
