# M3 — Approval kernel extraction — EVIDENCE

**Unit:** ROADMAP 3.1 (approval-remediation M3) · **Branch:** `claude/nice-wu-353cd4`
**Spec:** `docs/superpowers/specs/2026-07-12-m3-approval-kernel-extraction-plan.md`
**P0 gate:** `docs/superpowers/analysis/2026-07-12-approval-kernel-extraction-system-impact.md` (🟡 Yellow)
**Ratification:** all 4 items APPROVED by operator via hub (as recommended) — logged below.

## Ratification log
- 2026-07-12 — ESCALATION sent (commit ffe604c6 base). ACK: R1 additive route-admin contract; R2(a)
  thin template entry points, retire parallel path; R3 3-phase relocate-then-generalize; R4 count
  in-flight first then hard-cutover/drain. openapi edits authorized within R1/R2 shapes.

## Dispatch ledger (HARNESS §4.4)

| Slice | Dispatch | Implementer | Reviewer | Gates | Commit | Status |
|-------|----------|-------------|----------|-------|--------|--------|
| P1.S1 relocate tree + imports | 2026-07-12 | sonnet | sonnet ACCEPT | go build+vet green; staged rename 0/0 byte-identical | b37c46d0 | DONE |
| P1.S2 re-port audit edge | 2026-07-12 | main (verify) | — | boundary-lint GREEN | b37c46d0 | VOID — no real violation |
| P1.S3 composition + codegen | 2026-07-12 | main (config align) | self+gates | api-lint 0 violations; unit tests green | 092a79d2 | DONE |
| P1.S4 supersede ADR 0072 + guard | 2026-07-12 | main | negative-plant proof | boundary GREEN; plant RED→revert-clean→GREEN | 7f407646 | DONE |
| P2.S1 migration + backfill | 2026-07-12 | sonnet | sonnet ACCEPT-WITH-NITS | build+vet; 8 testMig0296 GREEN (canonical); api-lint 0; check-db-bootstrap PASS | 82b897f1 | DONE |
| P2.S2 domain generalize | 2026-07-12 | sonnet | sonnet ACCEPT-WITH-NITS | byte-equal doc path; version-copy col-order CONFIRMED (indep + reviewer) | fe581164 | DONE |
| P2.S3 route-admin contract delta | 2026-07-12 | sonnet | sonnet ACCEPT | additive-only diff proven; regen-clean; byte-equal default | 9062f169 | DONE |
| P3.S1 in-flight count | — | main/haiku | — | count + query recorded | — | pending |
| P3.S2 template entry points | — | sonnet | sonnet (indep) | tier-1 caps, kernel wire | — | pending |
| P3.S3 config→route migration | — | sonnet | sonnet (indep) | cutover rule applied | — | pending |
| P3.S4 retire parallel path | — | sonnet | sonnet (indep) | contract diff | — | pending |

## Gate results (fill per slice)

### P1.S2 finding — audit edge was a FALSE POSITIVE
The earlier-flagged `audit/delivery/http/handler.go → approval/http/router` edge is a COMMENT
reference (lines 3/74/76/80), NOT a real Go import. `sed -n '/^import (/,/^)/p'` on the audit
handler confirms NO approval import. `check-module-boundaries.ps1` → `[module-boundaries] OK` on
the relocated tree: the layer allow-list (`domain`/`application`/`api`) already covers every real
cross-module edge (documents↔approval, jobs→approval — all on allowed layers). No re-port required.
The stale `documents/approval` nested-family exception in the guard is now DEAD config (path no
longer exists) → cleaned in P1.S4 with negative-plant proof.

### P1.S3 gates
- `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → **0 violation(s)** (was 16:
  8 tripwire-allowlist entries stale at old path + 8 new-path violations unmatched; fixed by path
  rewrite in `scripts/api-lint/tripwire-allowlist.txt`).
- `go test ./internal/modules/{approval,documents,templates,audit}/...` → all **ok** (unit).
- `go build ./...` + `go vet ./...` → green.
- Pre-existing (NOT introduced here, out of ladder): `check-module-contract-sync.ps1 -Module approval`
  reports route-enumeration DRIFT; my diff to that script is paths-only and the grepped files
  (router.go/openapi/frontend) are byte-identical to main, so the drift predates this work
  (router.go has 0 occurrences of the literal `/api/v1/approval/inbox` pattern the check greps).

### P1.S4 gates — boundary guard realignment + negative-plant proof
1. Realigned guard on real tree → `[module-boundaries] OK`.
2. Plant `_ "metaldocs/internal/modules/approval/infrastructure"` in
   `internal/modules/jobs/stuck_instance_watchdog/job.go` (external module) → guard **FAIL**, names
   exactly `... job.go -> metaldocs/internal/modules/approval/infrastructure`.
3. `git checkout` → `git diff --exit-code` CLEAN; guard `[module-boundaries] OK` again.
- ADR 0082 written (supersedes ADR 0072 ruling (a)); ADR 0072 header stamped superseded.

### Phase-1 close — integration suite (L1)
- Run 1 (commit 7f407646): 11 failing = 9 accepted RED + **2 NEW regressions** I introduced.
  - 9 accepted (unchanged): controlleddocuments `TestTenantIsolation_SequenceCounters_CrossTenant`
    ×1 (E-PROD-2 document_profiles_pkey); jobs/approval_sla_surfacer ×4 (status ambiguous);
    scenarios ×3 (`TestGrantAreaMembershipFn`, `TestGrantAreaMembershipIdempotent`,
    `TestTriggerBypassBlocked`); tenantdata `TestTenantDataPortCoverage` ×1
    (approval_delegations/review_verdicts unregistered — pre-existing gap).
  - 2 NEW (FIXED, commit dc45f360): `TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx`
    in `tests/integration/scenarios/tx_ownership_test.go` — walk dir built from split
    `filepath.Join(root,"internal","modules","documents","approval",...)` args, so P1.S1's
    string-grep import sweep could not see it; the moved dir vanished. Repointed to
    `internal/modules/approval/{infrastructure,http}` (tx-ownership invariant guards — repaired,
    not deleted). `go vet -tags=integration ./tests/integration/scenarios/` clean.
- Run 2/3 (commit dc45f360): **INVALID** — Postgres OOM-crashed mid-suite (3 back-to-back heavy
  integration runs); every failure was `FATAL: the database system is in recovery mode (SQLSTATE
  57P03)` / `driver: bad connection`, NOT assertion failures. Root cause: **429 orphan
  `metaldocs_test_*` databases** left by the killed runs bloated WAL crash-recovery (`resetting
  unlogged relations` 600s+). Not a code regression.
- **Run 4 — CLEAN CLOSE (commit dc45f360):** waited for recovery (ready after ~605s), dropped all
  429 orphan test DBs (→ 0), ran canonical `test-integration.ps1` once. Full failure set =
  **exactly 9 tests / 4 pkgs, byte-match to accepted baseline, zero NEW:**
  - `controlleddocuments/application` `TestTenantIsolation_SequenceCounters_CrossTenant` ×1 (E-PROD-2)
    — re-ran (13.2s, not cached), identical accepted failure.
  - `jobs/approval_sla_surfacer` ×4 (`FullTick`, `Writer_TenantSeed`, `Idempotent_SecondRunNoOp`,
    `AlertOnly`) — error `column reference "status" is ambiguous (SQLSTATE 42702)`, a **pre-existing
    query defect**. Decisive proof of behavior-neutrality: this package imports `approval/domain`
    (repathed → build-cache invalidated → **forced re-run**) and reproduced its *identical*
    pre-existing error against the moved import path.
  - `scenarios` ×3 (`TestGrantAreaMembershipFn` invalid area_code, `TestGrantAreaMembershipIdempotent`,
    `TestTriggerBypassBlocked` session_replication_role/BYPASSRLS dev-env).
  - `tenantdata` `TestTenantDataPortCoverage` ×1 (`approval_delegations`/`approval_review_verdicts`
    unregistered ports — pre-existing gap, predates this work).
  - tx_ownership regressions (`TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx`) —
    **gone** (P1.S1 fix confirmed). `test-integration.ps1` exit 1 = throws on any RED incl. accepted.

**PHASE 1 (pure relocate) — CLOSED.** Boundary-lint GREEN + negative-plant; byte-equal behavior
proven (every re-compiled dependent reproduces its exact pre-existing accepted-RED; zero new). ADR
0072 ruling (a) superseded by ADR 0082 here.

## Phase 2 — generalize (subject_kind, subject_key)

### P2.S1 gates — migration 0296 (expand phase)
- Files: `db/migrations/0296_approval_subject_generalization.sql` + `tests/integration/migrations/migration_0296_test.go` (8 tests).
- `go build ./...` + `go vet -tags=integration ./tests/integration/migrations/...` → green.
- `.\scripts\test-integration.ps1 -Package ./tests/integration/migrations/... -Run TestMigration0296`
  → **PASS** (canonical runner, DATABASE_URL derived from .env — never hand-set). 8/8 GREEN.
  Coverage: backfill values both tables, CHECK reject (23514), unique reject (23505),
  partial-index allows inactive dup (route-versioning regression guard), template-subject
  insertable both tables, kept-index existence.
- `api-lint -strict` → **0 violations** (DB-only slice, no contract change).
- `check-db-bootstrap.ps1` → forward-migration execution on clean bootstrap (fresh volume).
- Implementer runtime-truth correction: baseline `approval_routes_tenant_profile_key` was already
  dropped by 0287 route-versioning → migration keeps the ACTUAL post-0287 constraints
  (`approval_routes_active_profile_uq`/`_profile_version_uq`). `ux_approval_routes_tenant_subject`
  made PARTIAL `WHERE active` to match (superseded rows share subject_key).
- Independent reviewer: **ACCEPT-WITH-NITS**, zero must-fix correctness defects. Nit-1 (compat-trigger
  removal not tracked) FIXED — debt now recorded in migration header + P2.S2 plan line. Nit-2
  (index existence-only assertion) left: partial/unique semantics covered by dedicated tests.
- Expand-phase compat shim: `default_approval_subject()` BEFORE-INSERT trigger backfills subject
  cols from legacy document cols when omitted → existing Go/testdb INSERTs work under new NOT NULL
  without a Go cutover. **Contract-phase debt (drop in P2.S2) tracked.**

### P2.S2 gates — domain Subject(kind,key) + explicit persistence (commit fe581164)
- New: `internal/modules/approval/domain/subject.go` (`SubjectKind` enum document|template, `Subject{Kind,Key}`,
  `NewDocumentSubject`, `Validate`/`Equal`/`String`) + `subject_test.go` (value object + Route/Instance
  projection invariants). `Route`/`Instance` gain `Subject` field.
- Production repo now writes `subject_kind`/`subject_key` EXPLICITLY on all 3 INSERT paths:
  `InsertInstance` (from `inst.Subject`, zero-value fallback → `NewDocumentSubject(document_id)`),
  route create (`route_admin_service.go` ~211), route version-copy/supersede (SQL `INSERT ... SELECT`
  copies source row's own subject cols). Compat trigger now a **production no-op** (still backstops testdb).
- Read hydration: `Subject` DERIVED in Go from legacy col (`NewDocumentSubject(profile_code/document_id)`)
  at all Route/Instance hydration sites — lower risk, provably equivalent for document rows
  (backfill set subject_key=document_id/profile_code). Diverges only if non-document subject_key ever
  differs → safe this phase (document-only); revisit at P3 template rows.
- Gates: `go build`/`go vet` clean · `go test ./internal/modules/approval/...` green ·
  consumers (documents, templates) green · full `go test ./...` (non-integration) no FAIL ·
  canonical `test-integration.ps1 -Package ./internal/modules/approval/...` PASS + `./tests/integration/approval/...`
  PASS (document submit/signoff byte-equal) · api-lint **0** · `check-module-boundaries.ps1` **OK**.
  No accepted-RED baseline test touched/newly broken.
- Reviewer **ACCEPT-WITH-NITS**, zero must-fix. Highest-risk item (version-copy `INSERT...SELECT`
  column order) CONFIRMED correct by BOTH an independent orchestrator read and the reviewer
  (8-col positional match, copies source route's own subject). Two nits (both "document-only-safe,
  breaks at P3 templates"): InsertInstance zero-Subject fallback + derive-from-legacy hydration —
  now tracked as explicit P3.S2 must-close items in the plan.

### P2.S3 gates — route-admin subject contract delta (commit 9062f169)
- **R1 additive-only, zero breaking change** — `git show 9062f169 -- api/openapi/v1/openapi.yaml`:
  `CreateRouteRequest.required` unchanged `[profile_code, name, stages]`;
  `RouteSummary.required` unchanged `[id, name, tenant_id, profile_code, active, version, stages,
  created_at, updated_at]`. `subject_kind` (enum document|template) + `subject_key` added as OPTIONAL
  props outside both `required` lists. No field removed, no new required entry, no type/pattern/enum
  narrowing on any existing field. **Reviewer determination: NO breaking change.**
- **Regen-clean** — `go generate ./internal/modules/approval/api/...` reproduces committed `api.gen.go`
  byte-for-byte (working tree clean after regen). New fields pointer-typed/optional
  (`SubjectKind *CreateRouteRequestSubjectKind`, `SubjectKey *string`). Not hand-edited.
- **Byte-equal default** — `resolveCreateRouteSubject` (route_admin_service.go): both fields absent →
  `Subject=(document, profile_code)`, identical persisted row to pre-slice. Proven two ways:
  fake-driver INSERT-arg capture unit test + real-Postgres `SELECT subject_kind, subject_key`
  integration test (`route_admin_service_subject_integration_test.go`). `profile_code` stays required
  throughout (contract, input struct, INSERT).
- **Enum validation at both layers** — HTTP `http/contracts/route.go` `CreateRouteRequest.Validate()`
  (reject unknown → 400) + domain `domain/subject.go` `Subject.Validate()` (`ErrInvalidSubjectKind`).
  Hand-written `contracts` package + generated `api.gen.go` both carry the fields, verified in sync
  (module decodes via hand-written contracts; gen types drive the strict-server route guard only).
- **Scope guard held** — grep of full commit diff: NO template routes, NO template governance, NO new
  capability, NO G1-policy change, NO P3.S4 retired-path work. `subject_kind=template` accepted +
  persisted faithfully with zero governance (Phase-3 wiring deferred).
- Gates (reviewer ran fresh, all green): `go build`/`go vet` clean · `go test
  ./internal/modules/approval/...` (8 pkgs) ok · consumers documents+templates ok · full `go test ./...`
  no FAIL · api-lint **0** · `check-module-boundaries.ps1` **OK** · canonical
  `test-integration.ps1 -Package ./internal/modules/approval/...` PASS, zero NEW failures (approval-only
  scope, none of the 9 baseline-RED pkgs exercised).
- Reviewer **ACCEPT**, zero must-fix, zero nits.

### Phase-2 close
- P2.S1 (DB expand + backfill + compat trigger) · P2.S2 (domain Subject + explicit persistence) ·
  P2.S3 (route-admin additive contract) all DONE + committed. Existing document routes/instances
  byte-equal; contract diff additive-only; kernel now keyed by `(subject_kind, subject_key)` with
  document as the projection. Ready for Phase 3 (templates onto kernel).

## Baseline (pre-work)
- Accepted RED on main: exactly 9 tests / 4 pkgs (E-PROD-1..5: sla_surfacer ×4, controlleddocuments
  cross-tenant sequence ×1, scenarios ×3, tenantdata ×1). Bar for every slice: zero NEW failures.
- approval subtree: 164 Go files. Coupling: 2 inbound production files (documents→approval), 24
  outbound (approval→documents domain/application), 1 true re-port (audit→approval/http/router),
  3 external consumers on allowed layers (audit, jobs/approval_sla_surfacer, jobs/stuck_instance_watchdog).

## Defers / notes
- E-PROD-2 (document_profiles PK) untouched — operator decision pending.

## HS-1
- Operator sign-off gate: pending (milestone close).
