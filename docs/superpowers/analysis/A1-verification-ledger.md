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

## 1a. The verify entry point

`go run ./tools/verify --profile={fast|changed|pr|full|release}` is the single
implementation of every deterministic check in this repo. CI jobs call it and
nothing else, so "green locally" and "green in CI" are the same claim. Job
names are unchanged, so the ruleset's required-check names are unaffected by
the rewiring.

- `--list` prints what each profile contains.
- `--audit` reports registry entries with no CI job and exits non-zero. A
  check that runs on laptops but not on PRs is advice, not a control — the
  same failure mode as an unwired script. Its first run found one:
  `go vet -tags integration`, now wired.
- `--only=id,id` runs specific checks; `--profile=changed` runs the `pr` set
  filtered by the diff.
- `--ci-job=file.yml:job` narrows a selection to the checks that declare that
  job as their owner (§6.7). A CI job passes its own name, so which job runs a
  check is decided by the registry, not by the workflow.
- `--guard-fixtures` feeds every guard its negative fixture and requires a
  non-zero exit (§6.1).

Three properties are deliberate:

**Checks are argv, not shell strings.** A quoting bug cannot silently change
what ran.

**A check that cannot run is reported as SKIP with its reason, and the reasons
are reprinted at the end of the run.** A check that vanishes when its
precondition is missing is an inert control, and a hole that scrolled off the
top of the output is a hole nobody sees.

**Toolchain versions are preflighted against `go.mod` and `.nvmrc`.** This
repo standardizes Node 26.3.0 in development and CI; Go remains pinned by
`go.mod` and CI's setup. A run on the wrong toolchain is not a verification of
what CI will do, and verify says so before it starts.

Measured: `--profile=fast` is 15 checks in ~51s wall clock.

## 1b. A required check may not carry a `paths:` filter

GitHub evaluates required status checks by name. A required check that never
reports is not treated as "not applicable" — it is treated as **pending**, and
the PR cannot merge. So a path-filtered workflow whose jobs are required
deadlocks every PR that falls outside its filter.

That is not merely a mechanical constraint; it is the same rule this axis is
built on, seen from the other side:

> **A control that sometimes does not fire is not a gate.**

An unwired script never fires. A path-filtered required check fires for some
diffs and not others, and the diffs it skips are exactly the ones nobody
chose deliberately — they are whatever the glob happened to miss. A reviewer
reading a green PR cannot tell the difference between "this check passed" and
"this check was not run," which makes green unfalsifiable.

Diff-scoping is still worth having; it just belongs one level down. The job
always runs, and `tools/verify` decides what is relevant from the `Paths`
field on each registry entry (`--profile=changed`). The saving is the same,
and the job still reports either way.

Applied on 2026-08-07: `api-contract.yml`, `lint.yml` and `fe-ci.yml` lost
their `paths:` filters. Report-tier workflows that will never be required
(`ci`, `perf`, `e2e-coverage-gate`, `supply-chain`, `req-traceability`,
`openapi-breaking`) keep theirs. **Promoting any of those to tier 1 requires
removing its filter in the same change** — otherwise the promotion silently
converts the check from "advisory" to "blocks unrelated PRs forever."

## 2. Check tiers

The `main` ruleset promotes tier-1 to `required`. Tier-2 runs on every PR and
reports, but cannot block, because its failure does not depend on the diff.

The earlier version of this section listed *workflow* names — `Go Lint`,
`api-contract`, `lint`, `Supply Chain`. A required status check is matched by
**job** name, and not one of those strings exists as a check context, so that
table could not have been applied as written. The names below are copied from
PR #96's actual check list.

### Tier 1 — required on `main`

Live as ruleset `main` (id `20560142`), exported to `.github/rulesets/main.json`.
Every entry was observed green on PR #96 before promotion; none was promoted on
the assumption that it would pass.

| Check (context) | Workflow:job | Enforcement |
|---|---|---|
| `golangci-lint` | `golangci-lint.yml:golangci-lint` | new findings only (`--new-from-patch`) |
| `cilint — custom Go linters` | `invariants.yml:cilint` | baseline ratchet (§3.1) |
| `gofmt + go vet + staticcheck` | `invariants.yml:staticcheck` | absolute — zero findings |
| `Migration monotonicity check` | `invariants.yml:migration-gapless` | absolute — monotonic |
| `api-design-system-lint` | `api-contract.yml` | absolute |
| `backend-codegen-drift` | `api-contract.yml` | absolute |
| `contract-sync` | `api-contract.yml` | absolute |
| `frontend-codegen-drift` | `api-contract.yml` | absolute |
| `openapi-lint` | `api-contract.yml` | absolute |
| `problem-codes-freshness` | `api-contract.yml` | absolute |
| `spec-base-path-gate` | `api-contract.yml` | absolute |
| `eslint` | `lint.yml:eslint` | absolute |
| `css-token-discipline` | `lint.yml` | absolute |
| `eigenpal-selector-pin` | `lint.yml` | absolute |
| `web-typecheck-test` | `fe-ci.yml` | absolute — typecheck + unit tests |
| `node` | `ci.yml:node` | absolute — docx-v2 typecheck/build/bundle guard |
| `check` | `governance-check.yml:check` | absolute — three proxy rules (§5) |
| `wiki module/tech-debt tally sync` | `governance-check.yml:wiki-tally` | absolute |
| `unit` | `test-smoke.yml:unit` | absolute — `go build ./...` + `go test ./...` |
| `smoke` | `test-smoke.yml:smoke` | absolute |
| `gitleaks` | `secret-scan.yml` | allowlist (§3.2) |

### Tier 2 — reports, not required

| Check | Why it cannot gate | Closes when |
|---|---|---|
| `hardening` | red on D-17 | `tests/contract` has Go files again, or the gate stops asking for them |
| `conformance` | red on D-13 | 13 test-discipline violations ported |
| `gate` | red on D-1; also `paths:`-filtered | evidence landed **and** filter removed |
| `DB schema dictionary coverage` | red on D-5 | 10 tables documented |
| `E2E smoke (approval flows)` | `webServer` exits 1 — no application stack in CI | stack provisioned (D-15) |
| `Perf suite (reduced — PR gate)` | no postgres service, no migrations, `PERF_DATABASE_URL` undefined | stack provisioned (D-16) |
| `full` | `test-full.yml` now runs on PRs but has not yet produced one complete result | first green run observed |
| `Axe baseline`, `Coverage map` | green, but `e2e-coverage-gate.yml` is `paths:`-filtered (§1b) | filter removed |
| `openapi-breaking`, `Supply Chain` | `paths:`-filtered (§1b) | filter removed |
| `docx-v2 PR must not touch CK5 paths` | job-level `if:` makes it report `skipped`, which satisfies a required check without running | condition reworked so it always evaluates |

A tier-2 entry is a promise, not a parking space. Each has a closing condition;
when it is met the check moves to tier 1 and its row is deleted from this table.

### Bypass

The ruleset has **zero** bypass actors. That is deliberate and it has a cost:
renaming a required job without updating the ruleset in the same commit will
deadlock every open PR, because the old context never reports and a
never-reporting check is pending forever. Recovery procedure is in
`.github/rulesets/README.md`.

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
| D-10 | `react-router-dom` 7.18.0 (GHSA-qwww-vcr4-c8h2). Fix is 8.3.0 — a major routing-API migration touching `AppRouter` and every route module. Advisory affects the unstable RSC APIs only. | `frontend/apps/web/package.json` | frontend axis | migrated to v8, or advisory formally accepted as not-applicable |
| D-11 | `pdfjs-dist` 5.7.284 (GHSA-hq66-cqwq-w95j). Fix is 6.2.108, a major bump changing worker loading and `GlobalWorkerOptions`; the pinned `react-pdf`@9.2.1 is not certified against it. | `frontend/apps/web/package.json` (devDependency) | frontend axis | viewer integration ported to pdfjs 6 |
| D-12 | `js-yaml` 4.1.1, `postcss` 8.5.15, `fast-uri` 2.4.0/3.1.2 — three advisories, one root cause: all are purely transitive with no entry in any of this repo's 9 manifests, and `pnpm update` cannot target such a dependency. Verified empirically: it left each at its vulnerable version while producing broad unrelated lockfile churn. `@redocly/openapi-core` pins `js-yaml` to the exact string `4.1.1`, so even a matching range would not help. | attempted and reverted during A1 | security axis | a `pnpm.overrides` block is introduced — a first-time manifest-pattern decision this repo has never made, which is why A1 did not make it unilaterally |
| D-5 | 10 baseline tables have no `wiki/database/tables` page: `audit_export_jobs`, `materialize_dispatch_outbox`, `tenant_keys`, `tenant_lifecycle_jobs`, `tenant_plans`, `token_dictionary_entries`, `approval_delegations`, `approval_review_verdicts`, `approval_route_stage_selectors`, `release_generations` | `governance-check / db-dictionary-coverage` job | docs axis | all 10 documented; job → tier 1 |
| D-6 | ~~module §11 tallies disagree with registers~~ | — | — | **CLOSED** — fixed in A1, job is tier 1 |
| D-13 | `check-test-discipline.sh` reports 13 violations, so `module-boundaries.yml:conformance` is red on every PR today. R1 (raw `set_config('metaldocs.asserted_caps', ...)` instead of the fixture seam) x8, R4 (raw `documents` SQL from approval tests) x4, R3 x1. Found by `verify --profile=fast`; it was not on the inherited scoreboard because nobody had run the job's second step locally. | `bash scripts/check-test-discipline.sh` | test-discipline axis | violations ported to the canonical fixture seam |
| D-9 | `wiki-tally-check.ps1` check 2 (missing-ADR) passes silently when the module doc omits the "Decisions without ADR link: N" line. 11 of 16 modules omit it, so the check is live for 5. A guard that a doc can opt out of by deleting a line is not a guard. | `wiki-tally-check.ps1:130` `if ($null -ne $adrStated -and ...)` | docs axis | line required in all 16 docs, then the `$null` escape deleted |
| D-7 | `check-db-bootstrap.ps1` asserts a real, uncovered invariant (baseline ledger marker + 6 critical tables) but is pinned to `docker exec metaldocs-postgres` | script source | ops axis | rewritten against a GH Actions `postgres:16` service, then wired |
| D-14 | The registry has no dependency edges. `docx-v2-test` reads `dist/meta.json` produced by `docx-v2-build`, and `verify` runs a single invocation's checks concurrently, so CI enforces the order by splitting into two invocations while a local `--profile=pr` can still race. | `tools/verify/registry.go` (`docx-v2-test`), `ci.yml:node` | A1, if it recurs; otherwise the verifier axis | a `Needs`-style edge exists, or no check has a producer/consumer relationship |
| D-15 | `E2E smoke (approval flows)` cannot pass: Playwright's `webServer` exits 1 because CI provisions no application stack. The cascade is worse than the failure — the next step, `axe diff vs baseline`, then ran anyway and died on `ENOENT: axe-report.json`, so the run reports two failures with one cause. | run 31192663852 | ops axis | stack provisioned; also make the axe step depend on the suite having produced a report |
| D-16 | `Perf suite (reduced — PR gate)` has no `services: postgres`, runs no migrations, and reads `DATABASE_URL` from `secrets.PERF_DATABASE_URL`, which this repository does not define. Supersedes D-2: the harness's exit code was never the problem. A1 fixed only the disguise — the readiness loop was `curl -sf … && break \|\| sleep 2` with no exit, so a dead backend still produced a green step and k6 printed a full percentile table over 300 failed requests and 0 bytes received. | run 31192664532; `perf.yml:31` | ops axis | postgres service + migrations + a defined DB URL, then one green run |
| D-17 | `phase3-hardening-gate` runs `contract-baseline.ps1`, which runs `go test ./tests/contract`. That directory holds a `.gitkeep` and nothing else — the tests were deleted in `dc0572f6` and the gate that consumes them was not. It fails with `no Go files in .../tests/contract`. **Process note:** I promoted `hardening` to required on the assumption it would go green once the cilint fix landed. It never was green; the cilint failure was masking this one. Promotion must follow an observed green, never a predicted one — that is the same "a control that fires into a red baseline is absent" error, made against my own work. | run 31194517804; `scripts/contract-baseline.ps1:42` | test-discipline axis | contract coverage restored, or the gate's target corrected |
| D-18 | `Axe baseline — no critical violations` is **inert**, and its own workflow header asserts the opposite ("the axe-baseline-check job below it are real automated checks, not advisory", `e2e-coverage-gate.yml:18`). Both of its steps read only `frontend/apps/web/e2e/axe-baseline.json` — the list of *accepted* violations — and that file is `[]`, 3 bytes. It never reads an axe report. It can therefore only fail if someone deliberately writes a critical violation into the accept-list; a real critical violation in the app passes it. Actual axe enforcement lives in the `axe diff vs baseline` step of the e2e suite, which is dead per D-15. Net: accessibility is unguarded while two green checks say otherwise. | `e2e-coverage-gate.yml:88-120`; `axe-baseline.json` = `[]` | frontend/a11y axis | the job consumes a produced axe report, not just the accept-list |
| D-19 | `phase3-hardening-gate` is a spent milestone gate: of its four steps, two duplicate existing checks, one targets a deleted suite (D-17), and the fourth disables the only vulnerability scanner it has. `go test ./...` = `test-smoke:unit`'s `go-test-unit`; `check-module-boundaries.ps1` = `module-boundaries:conformance`'s `module-boundaries` check; `contract-baseline.ps1` → empty `tests/contract`; `security-baseline.ps1` is invoked with `-SkipGovulncheck` because `phase3-hardening-gate.ps1:3` defaults `$SkipGovulncheck = $true`. The name refers to a "phase 3" that ended. | `scripts/phase3-hardening-gate.ps1:3,56,63,69,90` | A1 | resolved per §7 |
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
| kin-openapi + grpc critical/high CVEs | bumped to v0.144.0 / v1.82.1; staticcheck repo-wide 0 findings after |
| `tools/verify` introduced 6 golangci-lint findings | fixed in the code, not excluded from the linter. `selectChecks` decomposed; `readTrimmed(path)` → `nvmrcVersion()` reading a literal (G304 gone outright); `--base` validated against `refPattern`; all subprocess construction collapsed to one `command()` chokepoint. The repo went from zero lint suppressions to exactly one, at a chokepoint, with the invariant it relies on enforced in code rather than asserted in a comment. |
| `tools/cilint` path normalization was a no-op on Linux | `filepath.ToSlash` IS the identity function where the separator is already `/`, so a Windows-written baseline path stayed backslashed in CI and two tests passed only on the platform that did not need them. Replaced with an explicit `toForwardSlash`. This was "green locally, red in CI" living inside the verifier itself. |
| `build:docx-v2` ran AFTER the tests consuming its output | `bundle-guard.test.ts` reads `dist/meta.json`, so with the build last the guard could only ever fail with "missing" — it had never guarded anything. Build is now its own check, run in a prior verify invocation. Ordering hole recorded as D-14. |
| 5 gitleaks findings untriaged | all 5 classified, none a credential. Two protocol values exempted as a **class** (`Idempotency-Key`, `Sec-WebSocket-Key` — the latter explicitly not an authenticator, RFC 6455 §1.3) because QA evidence docs capture raw HTTP and will keep producing them. Two blob-storage object keys exempted per-**instance** in `.gitleaksignore`, because `"key":"…"` in JSON is also exactly what a real leaked API key looks like. The CI step gained `-v`; it previously reported `leaks found: 5` and named none. |
| duplicate `gitleaks` check name on every PR | `secret-scan.yml` fired on unscoped `push` **and** `pull_request`, producing two runs sharing one check name — a required check matched by name would have had no defined referent. Push scoped to `main`. Found by reading PR #96's own check list. |
| 96 Go files not gofmt-clean; nothing enforced gofmt | swept with the toolchain `go.mod` pins (1.25.0 via `GOTOOLCHAIN`), which produced the same list as local 1.26 — real drift, not a version artifact. `git diff -w` is import re-ordering, blank-line collapse and wrapping only; build + vet pass. Gated by a new `gofmt` check. The script fails closed on an empty file list (the first sweep blew the argv limit and reported a green that meant gofmt never ran) and includes untracked files (without that, a new file passes locally and fails in CI on commit — found because the negative fixture did not fire on its first attempt). |
| revive finding on an exported var with no doc comment | `CodeValidationSubmitChoiceRequired` had none, and neither did its two siblings. Pre-existing, but invisible: `golangci-lint` runs `--new-from-patch`, so a finding only surfaces once its line enters a diff — and the gofmt realignment put it there. Fixed by writing the three comments, not by excluding the file. A linter scoped to new lines will keep surfacing old debt this way; that is the ratchet working, not a regression. |
| `ci.yml` gated the docx-v2 bundle guard behind a `paths:` filter that did not match `package.json` or the lockfile | so the change most likely to break "the server bundle stays framework-free" — adding a dependency — was the one change that did not run the guard. Filter removed; `node` is now required on every PR. The redundant `go` job (`go build ./...` + 7 hand-listed packages, all covered by `test-smoke:unit`'s `go test ./...`) was deleted as duplicate compute. |
| perf readiness loop could not fail | `curl -sf … && break \|\| sleep 2` with no exit, so a backend that never started still produced a green step and k6 benchmarked a dead port — emitting a full percentile table over 300 failed requests and 0 bytes received. Output shaped like a measurement that measured nothing. Now fails at the cause. The job stays red for D-16; A1 removed the disguise, not the defect. |
| `main` had no ruleset and no branch protection of any kind | direct push, force-push and deletion all permitted; **every green check in this repo was advisory**. Ruleset `main` (id `20560142`) now requires a PR, blocks force-push and deletion, requires thread resolution, and requires 22 status checks — each one observed green on PR #96 before promotion. Verified live: PR #96 reports `mergeable: MERGEABLE, state: BLOCKED`. Exported to `.github/rulesets/main.json`, with the README stating plainly that GitHub does not read that file. |

## 6. Amendment log

| Date | Change |
|---|---|
| 2026-08-07 | Created. Tiers set, cilint + gitleaks + Grype baselines recorded, D-1…D-4 deferred. |
| 2026-08-07 | Orphan scripts triaged (§4a). D-5…D-9 added, D-6 closed same day, `wiki-tally` → tier 1. |
| 2026-08-07 | kin-openapi + grpc bumped; D-10…D-12 recorded for the five advisories the stop rule refused to force. staticcheck now 0 findings repo-wide → tier 1. |
| 2026-08-07 | All 17 deterministic CI jobs rewired to `go run ./tools/verify --only=…` (item 5 complete). §1b added: a required check may not carry a `paths:` filter. `api-contract`/`lint`/`fe-ci` filters removed; `test-full` moved to `pull_request` (item 4). PR #96 opened as the evidence run — the first PR to exercise the whole check surface, and it immediately exposed the duplicate `gitleaks` name, the Linux-only cilint failure and the docx-v2 build ordering. gofmt gated; D-14 recorded. |
| 2026-08-07 | **Item 6 complete.** `main` ruleset created (id `20560142`) with 22 required checks and zero bypass actors; blocking verified live against PR #96. §2 rewritten from real evidence — the previous tier-1 table listed *workflow* names (`Go Lint`, `api-contract`, `lint`, `Supply Chain`), none of which exists as a check context, so it could not have been applied as written. `ci.yml` filter removed and its `node` job promoted. D-15/D-16 recorded; D-16 supersedes D-2. |

## 6. Phase 1 amendment — #87/A1 verifier spine (2026-08-09)

This section records what Phase 1 changed, and what it deliberately did not.

### 6.1 Negative-fixture spine (A1.2)

`go run ./tools/verify --guard-fixtures` feeds each guard a tree of
deliberately bad input and requires a non-zero exit. It runs the check's **own
`Argv`**, not a library call — the property under test is that the command CI
executes says no, which is not the same claim as "the guard's helper functions
are unit-tested". Fixtures live in `scripts/testdata/guard-fixtures/<check>/`
with a trailing `.txt` on every source file, because this repo's own guards
walk `testdata/`.

19 guards carry a fixture; 19 blocking checks carry a one-line
`FixtureWaiver` naming why they do not (third-party tools, test suites that
are their own evidence, and three transitional ones). Audit rule **A7** makes
that coverage blocking: a `pr`-profile check with neither is a finding.

The positive half of the property is not fixtured. Every one of these checks
runs against this repository on every PR, and this repository is valid input —
a synthetic good case would prove strictly less.

Run: 19 fixtures, 0 failed, 1m6s wall (parallel). Wired as the
`guard-fixtures` check in `ci.yml:verify`.

### 6.2 Reachability (A1.4)

`gosec` and `govulncheck` ran only in `nightly.yml:security-scan`, which no
`needs:` edge connects to `ci.yml:required`. They gated nothing. Both are now
`pr`-blocking with `CIJob: ci.yml:verify` — a repoint inside the existing
required closure, so no branch-ruleset change is involved.

Promotion was gated on a live run, and the run contradicted the prior triage
note: gosec v2.28.0 reported a live G705 finding whose suppression was written
in golangci-lint's `//nolint:gosec` syntax, which standalone gosec never reads.
Converted to a gosec-native `#nosec G705 -- reason`; re-run clean.

### 6.3 Toolchain pinning (A1.3)

gosec `@v2.28.0`, govulncheck `@v1.6.0`, `smoke.yml`'s checkout SHA-pinned, and
audit rule **A9** to keep it that way (workflow `uses:` must be a SHA; a check
`Argv` may not fetch `@latest`). A9 was written first and observed firing on
exactly those three gaps.

### 6.4 `release` profile (A1.1)

`release` = `full` minus the three checks whose subject is a PR diff
(`oasdiff-breaking`, `governance-diff-rules`, `migration-gapless`) — on a tag
they can only fail for the wrong reason. Membership is by exclusion, so a new
check is in `release` by default.

`release.yml` now runs it before publishing the SBOM. Until this change a tag
could be pushed and an SBOM published with **no** verification behind it:
`ci.yml` triggers on `pull_request` only.

### 6.5 golangci-lint is a registry check (A1.1)

It ran as a bare `golangci-lint-action` step, so `--audit` could not see it and
`verify --profile=pr` did not run it — two definitions of "verified".
`ci.yml:lint-go` now installs the pinned binary (v2.11.4, the exact patch the
Action's `version: v2.11` was resolving to) and calls the verifier.

Registering it exposed the second half of the same problem: `ci.yml:verify`
runs `--profile=changed`, which selected **every** `pr` check, including one
owned by another job with another job's binaries installed. The first CI run of
this branch failed on `golangci-lint is not on PATH`. The fix is `--ci-job=`
(§6.8), not dropping the check out of `pr` — a profile that omits a check which
blocks a PR is the exact lie A1.1 exists to remove.

### 6.6 A4.0 — foreign-SQL *writes*

`hgcrossmodule` matched reads only (`FROM`/`JOIN`); `UPDATE documents SET ...`
produced zero findings. It now matches `UPDATE`, `INSERT INTO` and
`DELETE FROM` as well, in the same analyzer — not a second scanner.

The census: **11 findings across 10 sites**, recorded in
`tools/cilint/baseline.json` with `owner: "#93/A4"`. They are **not fixed
here**. Porting a cross-module write to the owning module's application service
is #93/A4's property; repairing it inside A1 would be another axis's debt
cleaned silently to keep a new guard green.

Mandatory negative fixture:
`scripts/testdata/guard-fixtures/arch-lint/` contains a synthetic
`approval -> UPDATE documents`, and the guard fails on it.

### 6.7 `--ci-job=` — CIJob becomes executable (A1.1)

`Check.CIJob` was documentation read only by `--audit`. It is a selector now:
`--ci-job=ci.yml:verify` narrows a profile to the checks that declare that job
as their owner, and `--audit` mirrors the same narrowing when it reads a
workflow's `run:` block, so the audit describes the selection the command
actually makes.

This is what lets the two claims coexist: a profile is the honest full answer
to "what blocks a PR" (so a local `--profile=pr` is the same definition CI
uses), while CI still splits that set across jobs that install different
prerequisites. An unknown job name is an error, exactly as an unknown
`--only=` ID is.

### 6.8 First live catch by the fixture spine — `module-imports`

The first CI run of this branch reported, in the same job:

```
PASS  module-imports              1.5s
FAIL  module-imports   exited 0 on bad input — the guard does not guard
```

`scripts/check-module-boundaries.ps1` appended a literal `\` to the repo root
before comparing paths. Under pwsh on Linux the separator is `/`, so no file
matched the root prefix, every file kept its absolute path, failed the
`^internal/modules/` match and was skipped: **zero files inspected, exit 0**,
on the only platform that gates a merge. The guard had been decorative in CI
for as long as it had run there, and every green `module-imports` status on a
Linux runner was green over nothing.

Fixed by normalising both sides to forward slashes, plus a fail-closed
inspected-file counter so "inspected nothing" and "found nothing" stop sharing
an exit code. This is the A1.2 property paying for itself on its first run:
no unit test of the script's internals would have reported it, because the
helper logic is correct — the *command* was not.

### 6.9 Explicitly not closed

- `deploy/compose/docker-compose.yml` pins `minio/minio` and `minio/mc` at
  `:latest`. Same class as A1.3, different axis — nothing in CI or the verifier
  executes that file.
- The four-inventory drift entry (ME-13 in
  `docs/engineering/mechanical-enforcement-register.md`) still stands: A7–A9
  make more drift visible, they do not make it unrepresentable. The global
  maximum remains ROADMAP §4 row 4.7, "generated CI manifest".
