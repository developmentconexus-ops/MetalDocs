# A1 — Verification Ledger

**Status:** live. Amended whenever a check changes tier or a baseline moves.
**Owner:** axis A1 (`docs/superpowers/analysis/A1-handoff.md`, issue #87).

## 1. What "green" means here

A1 does not fix the defects its checks reveal. Fixing them is A2–A9. But a
check that reports a known-red condition cannot be made `required`, and a check
that is not `required` is not a control. The resolution:

> **Green = no violation outside the recorded baseline.**

Every red condition A1 inherits is either
(a) **recorded** — captured in a machine-readable baseline the check enforces as
a ratchet, so new instances fail the build while known ones do not; or
(b) **deferred** — listed in §4 with a date, an owning axis, and the condition
that closes it.

A red condition that is neither recorded nor deferred is a bug in this ledger.

**A control that fires into a red baseline is absent.** That is the rule this
document exists to satisfy. A baseline is the mechanism that turns a
permanently-red check into a live control today, at the cost of admitting what
is already broken — in writing, with an owner.

Every baseline in §3 is **transitional**. Each names the global maximum (an
empty baseline) and the milestone that deletes it. A baseline with no exit is a
suppression wearing a ratchet's clothes.

## 2. Check tiers

The `main` ruleset promotes tier-1 to `required`. Tier-2 runs on every PR and
reports, but cannot block, because its failure does not depend on the diff.

### Tier 1 — required on `main`

| Check | Workflow | Enforcement |
|---|---|---|
| `Go Lint` | `golangci-lint.yml` | absolute — zero findings |
| `api-contract` | `api-contract.yml` | absolute — spec/codegen/problem-code parity |
| `CI Invariants — cilint` | `invariants.yml` | baseline ratchet (§3.1) |
| `CI Invariants — staticcheck` | `invariants.yml` | absolute — zero findings |
| `CI Invariants — migrations` | `invariants.yml` | absolute — monotonic |
| `lint` (eslint + css tokens + eigenpal pin) | `lint.yml` | absolute — zero findings |
| `governance-check / wiki-tally` | `governance-check.yml` | absolute — 16-module sweep, green |
| `fe-ci` | `fe-ci.yml` | absolute — typecheck + unit tests |
| `secret-scan` | `secret-scan.yml` | allowlist (§3.2) |
| `Supply Chain` | `supply-chain.yml` | severity gate (§3.3) |
| `test-full` | `test-full.yml` | absolute — must move pre-merge (A1 item 4) |

### Tier 2 — reports, not required

| Check | Why it cannot gate | Closes when |
|---|---|---|
| `E2E Coverage Gate` | needs `E2E_DATABASE_URL`; the secret does not exist | secret provisioned |
| `Perf Benchmarks` | harness exits 99 unconditionally | exit code fixed |
| `req-traceability` | 4 MUST REQs have no test evidence (§4) | evidence landed |
| `governance-check / db-dictionary-coverage` | red on D-5 | 10 tables documented |

A tier-2 entry is a promise, not a parking space. Each has a closing condition;
when it is met the check moves to tier 1 and its row is deleted from this table.

## 3. Recorded baselines

### 3.1 cilint — `tools/cilint/baseline.json`

102 findings across 35 keys. Key = `analyzer + file + message`; never a line
number, so unrelated edits above a finding do not invalidate it.

The ratchet fires **in both directions**:
- a key not in the baseline → exit 1 (new violation)
- a count above the baseline → exit 1 (regression)
- a count *below* the baseline → exit 1, stale-baseline message (a fixed
  violation must be removed from the record, or the baseline silently
  re-admits it later)

SARIF still reports every finding to code scanning. The baseline governs the
exit code only — it hides nothing from view.

| Analyzer | Keys | Findings | Owning axis |
|---|---|---|---|
| `hgcrossmodule` | 34 | 101 | module-boundary axis; deleted by M3 |
| `platformboundary` | 1 | 1 | `internal/platform/tripwire/arms.go` — the path is embedded in api-lint parity rules and GMR validation contracts; moving it needs an ADR |

**Global maximum:** `entries: []`. **Deleted by:** M3.

### 3.2 secret-scan — `.gitleaks.toml`

Documented allowlist entries only. Each entry names the finding, the reason it
is not a live secret, and the evidence. Scanning from a cutoff commit was
rejected: a scanner that cannot see history is an inert control by another name.

The CI step must pass `-v`. A gate that fails without naming what failed is its
own defect — the pre-A1 step reported only `leaks found: 5`.

### 3.3 Supply Chain — Grype

Gate severity: **critical + high**. Medium and low report without blocking.

Stop rule for A1: a bump that needs application code changes is **not** made
here. It is recorded in §4 and handed to the owning axis. A1 bumps versions; it
does not port code.

## 4. Deferred defects

| # | Defect | Evidence | Owning axis | Closes when |
|---|---|---|---|---|
| D-1 | 4 MUST REQs have no test evidence | `req-traceability` job output | traceability axis | each MUST cites a live test |
| D-2 | Perf harness exits 99 unconditionally | `perf.yml` run log | ops axis | harness returns a real exit code |
| D-3 | `E2E_DATABASE_URL` secret absent | `e2e-coverage-gate.yml` skips | ops axis | secret provisioned |
| D-4 | Medium/low CVEs unbumped | Grype SARIF | supply-chain axis | bumped or accepted with rationale |
| D-5 | 10 baseline tables have no `wiki/database/tables` page: `audit_export_jobs`, `materialize_dispatch_outbox`, `tenant_keys`, `tenant_lifecycle_jobs`, `tenant_plans`, `token_dictionary_entries`, `approval_delegations`, `approval_review_verdicts`, `approval_route_stage_selectors`, `release_generations` | `governance-check / db-dictionary-coverage` job | docs axis | all 10 documented; job → tier 1 |
| D-6 | ~~module §11 tallies disagree with registers~~ | — | — | **CLOSED** — fixed in A1, job is tier 1 |
| D-9 | `wiki-tally-check.ps1` check 2 (missing-ADR) passes silently when the module doc omits the "Decisions without ADR link: N" line. 11 of 16 modules omit it, so the check is live for 5. A guard that a doc can opt out of by deleting a line is not a guard. | `wiki-tally-check.ps1:130` `if ($null -ne $adrStated -and ...)` | docs axis | line required in all 16 docs, then the `$null` escape deleted |
| D-7 | `check-db-bootstrap.ps1` asserts a real, uncovered invariant (baseline ledger marker + 6 critical tables) but is pinned to `docker exec metaldocs-postgres` | script source | ops axis | rewritten against a GH Actions `postgres:16` service, then wired |
| D-8 | **Nothing anywhere verifies that a backup can be restored.** `run-backup-restore-gate.ps1` is a manual runbook step with no execution log proving a cadence | `wiki/runbooks/backup-restore.md:260-277` | ops axis | scheduled weekly DR drill against disposable CI Postgres with CI-only credentials |

D-8 is not documentation debt. It is the only entry here where the untested
thing is a recovery path, and the failure mode is discovering it at the moment
it is needed. Recommend it be scheduled ahead of the rest of its axis.

## 4a. Scripts triaged (A1 item 3)

"An unwired script is a lie." Eight orphans, resolved:

| Script | Verdict | Where |
|---|---|---|
| `check-eigenpal-selector-pin.sh` | **WIRED** — tier 1, green today | `lint.yml / eigenpal-selector-pin`. Enforces the half of ADR 0046 that was ratified and never enforced: eslint guarded the import boundary, nothing guarded the selector pin. |
| `wiki-tally-check.ps1` | **WIRED** — tier 1, green | `governance-check.yml / wiki-tally`. Required a `-All` sweep mode first: `-Module` was mandatory, so a job that just called the script would have failed on a missing-parameter error — a check that reports red without ever checking anything. The sweep then found **2** drifting modules (`controlled-documents`, `templates`), not the 1 a hand-sampled subset found. Both reconciled: registers are the primary data, §11 is a derived tally, so the tally moved. Hole recorded as D-9. |
| `check-db-dictionary-coverage.ps1` | **WIRED** — tier 2, red on D-5 | `governance-check.yml / db-dictionary-coverage` |
| `check-release-v2-names.ps1` | **DELETED** | Its own last output flagged 1071 of 1123 hits "unexpected", including permanent names (`internal/platform/docgenv2`). The V2 rename it policed finished; `v2` is now architecture, not debt. Its stale report `docs/release/v2-name-inventory.md` deleted with it. |
| `check-system-runnable.ps1` | **KEEP** — dev tool | Named in `CLAUDE.md`. Needs live API + MinIO + login. Not an orphan; a documented ritual. |
| `run-backup-restore-gate.ps1` | **KEEP** — dev tool, but see D-8 | `wiki/runbooks/backup-restore.md` |
| `check-baseline-equivalence.ps1` | **KEEP** — dev tool | `wiki/database/migration-policy.md:49` calls it mandatory before a baseline fold. Fires at fold time only; CI-porting means building two databases by two paths inside one runner. Lower priority than D-7. |
| `check-db-bootstrap.ps1` | **DEFERRED** — D-7 | Real invariant, needs a rewrite before it is CI-eligible. |

A "KEEP as dev tool" is only honest if the tool is reachable from a document
someone actually reads. All three keeps are cited above; none is an orphan.

## 5. Closed during A1

| Condition | Resolution |
|---|---|
| `problem-codes.md` stale | regenerated; 144 codes current |
| 14 missing doc comments | written, incl. 8 Noop ports rewritten to state they are not store-backed |
| `TxAuthzSeam` nil accepted silently | boot-fatal panic in all 3 constructors |
| `tenantdata/registry` under `internal/platform` | moved to `internal/composition`; `platformboundary` 15 → 1 |
| `cilint:allow-legacy` inline escape hatch | deleted; two canonical constants instead; `legacyvocab` 4 → 0 |
| SA5011 nil deref in `metaldocs-api` main | real bug — `riverBundle` dereferenced outside its own nil guard |
| 8 templates problem codes declared but never returned | wired to match approval verbatim; statuses unchanged, codes now specific |
| 7 hardcoded `go-version: 1.25.x` | `go-version-file: go.mod` |
| CSS off-palette teal button | `var(--brand)` |
| `crypto.subtle` cross-realm test failure | narrow `digest` spy, cause documented |
| axe accessibility diff terminated with `\|\| true` | removed. The workflow header called these "real automated checks, not advisory"; the swallow made that false. Artifact upload is `if: always()`, so nothing is lost. |
| 3 orphan scripts unwired, 1 rotted | wired / deleted per §4a |
| 2 module docs' §11 tallies disagreed with their registers | reconciled; `wiki-tally` sweep green across 16 modules |

## 6. Amendment log

| Date | Change |
|---|---|
| 2026-08-07 | Created. Tiers set, cilint + gitleaks + Grype baselines recorded, D-1…D-4 deferred. |
