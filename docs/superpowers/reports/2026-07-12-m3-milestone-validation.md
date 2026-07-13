# M3 (approval kernel extraction) — Milestone-close VALIDATION

**Judge:** independent milestone-validator subagent (fresh session; no implementation authorship).
**Unit:** approval-remediation M3 / ROADMAP 3.1.
**Branch:** `claude/nice-wu-353cd4` · **Worktree:** `C:\Users\leandro.theodoro\Documents\MetalDocs\.claude\worktrees\nice-wu-353cd4`.
**Diff range judged:** `cd53c658..HEAD` (46 commits, HEAD `6eae06e8`).
**Date:** 2026-07-12 (RE-JUDGE #2 after fix commit `6eae06e8`).

> Separation of powers honored: this judge ran gates from clean state, reviewed the aggregate diff,
> and wrote ONLY this verdict file. No source/spec/evidence/status was edited or flipped. `.env` was
> copied for the DB-backed runs and removed after — never printed, staged, or committed.

---

## C7 VERDICT — **PASS** (milestone may close)

The re-judge #1 (HEAD `f702414a`) returned FAIL for exactly one reason: 5 stale integration tests in
`tests/integration/migrations` (`migration_0296_test.go` / `migration_0297_test.go`) hardcoded
`document.submit` as the asserted tripwire cap regardless of `subject_kind`, broken by the 0299/0300
subject-discrimination slices (ADR 0083). That regression is now FIXED by commit `6eae06e8`
(test-only, 2 files). The delta from the prior judged HEAD (`f702414a..HEAD`) is that single commit and
nothing else. I re-ran the migrations suite myself from clean state (cache cleared) and it is GREEN;
approval + templates integration GREEN; build/vet/api-lint/module-boundaries GREEN. The aggregate
`cd53c658..HEAD` diff is coherent, the kernel is code-strong, and the intentional legacy
template-approval coexistence is correctly ratified (Option A, ROADMAP 3.1a) — not a defect. Every
remaining RED in the shared-dev-DB integration surface is confirmed **out of the M3 commit range** and
attributable to M6/M7 migration drift, not M3. **All C1–C6 pass; fail-closed criteria met.**

---

## Inputs loaded (C0)

- Milestone plan `docs/superpowers/specs/2026-07-12-m3-approval-kernel-extraction-plan.md`.
- Evidence ledger `docs/superpowers/reports/2026-07-12-m3-kernel-extraction-evidence.md`.
- ADR 0082 (`wiki/decisions/0082-approval-kernel-extraction.md`) incl. §Transitional coexistence.
- ADR 0083 (`wiki/decisions/0083-subject-discriminated-capability-tripwire.md`).
- ROADMAP (`docs/superpowers/ROADMAP.md`) rows 3.1 + 3.1a.
- Migrations `db/migrations/0296..0300`; `internal/platform/tripwire/arms.go`; kernel routes
  `internal/modules/templates/delivery/http/routes_approval_kernel.go`; `apps/api/cmd/metaldocs-api/permissions.go`.
- Fix commit `6eae06e8` (test-only; `tests/integration/migrations/migration_0296_test.go` +
  `migration_0297_test.go`).
- Aggregate `git diff cd53c658..HEAD` (237 files, +8740/-861).

> Note: `.claude/skills/milestone/references/milestone-end-validation.md` is NOT present in this
> worktree (only `developing-new-work`, `gitnexus`, `harness-hub` skills exist). Executed the C1–C7
> dimensions per the dispatching task's explicit binding gate ladder and done-definition.

---

## C1 — Per-feature spec/plan conformance + consumer contract + non-goals — **CONFIRMED**

Unchanged from re-judge #1 and re-verified against the final tree:

- **Extraction (ADR 0082):** `documents/approval` promoted to top-level `internal/modules/approval`
  (15th module). `check-module-boundaries.ps1 → [module-boundaries] OK`; approval has zero non-test
  import of `templates`.
- **Subject generalization:** `(subject_kind, subject_key)` across migrations 0296/0297/0298; two-level
  TEXT keying implemented in domain + repo.
- **Subject-discriminated tripwire (ADR 0083):** `arms.go` splits arm #1 (approval_instances INSERT →
  document→`document.submit` / template→`template.submit`, `WhenColumn=subject_kind`) and arm #2
  (approval_signoffs INSERT parent-lookup → document→`document.signoff` / template→`template.approve`).
  Rendered migrations 0299 (NEW-column CASE) + 0300 (parent-lookup CASE), disjoint arrays, fail-closed
  `ELSE` P0001 as a BEFORE INSERT row trigger. Capability sets never unioned — security property held.
- **Additive template routes (contract-first):** `POST /templates/{id}/versions/{n}/submit-for-approval`
  + `/signoff`; thin delegations to the kernel; tier-1 gated (`CapTemplateSubmit` / `CapTemplateApprove`).
  Legacy triad untouched.
- **Legacy coexistence (operator Option A, 2026-07-12):** ADR 0082 §"Transitional coexistence" +
  ROADMAP row 3.1a record the retained legacy role-based template-approval path and its deferred
  retirement. Per binding context, this is INTENTIONAL and ratified — **not** a milestone failure.

## C2 — Re-run each gate from clean state — **PASS**

Re-ran all gates myself (cache cleared for the fix-target suite; not trusting any transcript).
Postgres `metaldocs-postgres` Up 4h healthy; integration connects as `metaldocs_ci`.

| Gate | Command | Result |
|------|---------|--------|
| build | `go build ./...` | **exit 0** clean |
| vet | `go vet ./...` | **exit 0** clean |
| api-lint | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | **0 violation(s)** (exit 0) |
| boundaries | `check-module-boundaries.ps1` | **[module-boundaries] OK** (exit 0) |
| **integration — migrations** | `go clean -testcache; test-integration.ps1 -Package './tests/integration/migrations/...'` | **PASS — `ok metaldocs/tests/integration/migrations 11.252s`** (fresh, NOT cached) |
| integration — approval | `test-integration.ps1 -Package './internal/modules/approval/...'` | **PASS** — 8 subpkgs ok (application 22.3s, http 4.9s, infrastructure 14.6s, jobs 10.8s, signature, idempotency, contracts, domain) |
| integration — templates | `test-integration.ps1 -Package './internal/modules/templates/...'` | **PASS** — 5 subpkgs ok |

**(a) migrations GREEN — sole prior FAIL cause cleared.** The exact 5 tests that were RED in re-judge
#1 now pass against the fresh per-pid `testdb` clone (schema current, 0299/0300 applied). Confirmed on
a non-cached run after `go clean -testcache`.

**(b) approval + templates GREEN** — the live submit/signoff paths that exercise the 0299/0300
discriminated tripwire pass end-to-end (not fixture-only): approval `application`/`http`/`infrastructure`
suites drive real-DB inserts through the discriminator arms.

## C3 — Senior review of the aggregate diff — **CONFIRMED (coherence gap CLOSED)**

- The re-judge #1 finding was an intra-milestone break: the 0299/0300 tripwire slice silently
  invalidated the 0296/0297 migration tests, undetected because the full migrations suite was not
  re-run after the tripwire work. The fix commit `6eae06e8` closes exactly that gap, and I verified the
  full migrations package runs green from clean state — the "one feature breaks another" condition no
  longer holds.
- **Fix quality (senior review):** test-only, 2 files, no production change. It replaces the hardcoded
  `document.submit` constant in `insertInstanceWithSubject` / `insertInstanceFull` with
  `capForSubject(subjectKind)` (document→`document.submit`, template→`template.submit`; never unioned,
  mirroring `arms.go` #1a/#1b), and re-expresses the invalid-`subject_kind` case via
  `assertSubjectDiscriminatorFailClosed` — asserting the trigger's fail-closed P0001 ("no discriminated
  arm") which now correctly fires BEFORE the CHECK constraint as a BEFORE INSERT row trigger. The change
  is well-documented, models the real DB precedence, and does not weaken any invariant proof
  (`approval_routes` is not tripwire-gated, so its CHECK assertions are untouched). No split-brain, no
  dead code, no scope drift.
- The relocate remains a disciplined `git mv` + import rewrite; tripwire arm model is well-abstracted;
  api-lint PARITY/DRIFT pin Go↔SQL↔DB.

## C4 — Workflow-class QA + regression vs prior milestones — **PASS**

- **M3-owned suites are all green** (migrations, approval, templates). The milestone's own integration
  gate no longer ships RED.
- **Pre-existing / out-of-M3-range failures (NOT close blockers), attribution verified by git:**
  - `tests/integration/iam` — **0 commits** in `cd53c658..HEAD` (`git log --oneline cd53c658..HEAD -- tests/integration/iam` empty). M7 drift (dup-key seeding, missing `tenant_keys` relation, missing `erased_at` col).
  - `tests/integration/tenantdata` — **0 commits** in range. `TestTenantDataPortCoverage` flags approval_delegations/approval_review_verdicts missing TenantDataPort (accepted baseline).
  - `tests/integration/scenarios` — 2 in-range commits, but they touch ONLY `legacy_absent_test.go` (relocate import) and `tx_ownership_test.go` (P1 guard fix), both intended M3 work and both passing. The actual RED tests live in untouched files: `TestConcurrencyScenarios`/`TestOutbox_*` fail on `column "governance_class" of relation "document_profiles" does not exist (SQLSTATE 42703)` — M6 migration 0295 absent on the shared dev DB the raw-`seedTx`/`DATABASE_URL` tests use; `TestGrantAreaMembershipFn`/`TestGrantAreaMembershipIdempotent`/`TestTriggerBypassBlocked` are the accepted-baseline members. None are M3-attributable.
  - Root cause of the split: M3-owned suites use the fresh per-pid `testdb.Open` clone (current schema, all migrations applied) and pass; the raw-DSN packages hit the shared dev DB that is behind on M6/M7 — pre-existing, pre-disclosed environment drift, orthogonal to M3.
- No regression to prior-milestone behavior is attributable to M3.

## C5 — Quality-bar re-measure + "could it be built better" — **PASS**

- Root-cause discipline in the kernel is strong (ADR 0083 rejects the match-one union as a security
  regression and builds the discriminator — root cause, not symptom).
- The re-judge #1 verification-discipline gap (full integration surface not re-run after the
  authz-last-line change) has been acted on: the breaking slice's own tests were realigned and the
  package re-run green. The fix corrects the test to the real DB contract rather than weakening the
  trigger — the invariant is still proven (fail-closed P0001 asserted explicitly).

## C6 — Forbidden-list — **NO HIT**

- **suite-not-actually-green-at-close:** RESOLVED. The milestone's own migrations integration gate is
  GREEN at HEAD on a fresh, non-cached run; approval + templates green; per-feature acceptance maps to
  real-DB evidence. The prior "green rests on a stale pre-slice run" condition no longer applies.
- Not hit: fixture-as-real (live real-DB proofs), guessed-contract, self-judged-close, scope-drift
  (legacy coexistence is ratified + recorded, not drift), symptom-patch (fix corrects the test to the
  real contract; kernel fixes are root-cause).

---

## Housekeeping note (non-blocking, for the main session — I did NOT act on it)

The working tree is dirty: modified wiki docs (`wiki/modules/approval.md`, `templates.md`,
`documents.md`, `wiki/workflows/approval.md`, etc.) and the evidence ledger are uncommitted, and this
verdict file was previously untracked. These are documentation/evidence artifacts, not source or
contract changes, and do not affect the code/function verdict. The main session should commit them as
part of the close action (I do not commit or flip status — separation of powers).

---

**C7: PASS.** The sole prior FAIL cause (5 stale `tests/integration/migrations` assertions) is fixed by
the test-only commit `6eae06e8`; I re-verified migrations + approval + templates integration GREEN from
clean state, and build/vet/api-lint/module-boundaries GREEN. The aggregate `cd53c658..HEAD` diff is
coherent and contract-clean, the legacy coexistence is operator-ratified (not a defect), and all
remaining integration RED is confirmed out of the M3 commit range (shared-dev-DB M6/M7 drift). The
milestone may close on the main session's action.
