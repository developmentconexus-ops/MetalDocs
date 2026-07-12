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
| P2.S1 migration + backfill | — | sonnet | sonnet (indep) | testdb backfill | — | pending |
| P2.S2 domain generalize | — | sonnet | sonnet (indep) | byte-equal doc path | — | pending |
| P2.S3 route-admin contract delta | — | sonnet | sonnet (indep) | additive-only diff | — | pending |
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
