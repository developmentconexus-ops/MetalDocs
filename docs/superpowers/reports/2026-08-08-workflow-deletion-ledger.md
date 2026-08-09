# Workflow deletion ledger — Task 12 Step 1

Scope: prove every control in the 18 workflow files slated for deletion has a named,
**verified** successor before `git rm` runs. Each row below was checked two ways: (1) the
claimed successor location was read directly (`ci.yml`, `nightly.yml`, `release.yml`,
`tools/verify/registry.go`), not inferred from the plan text; (2) where a registry ID is the
successor, its `CIJob` field and `Argv` were read to confirm the check really runs the same
command against the same paths.

Spec reference: `docs/superpowers/specs/2026-08-07-ci-restructure-design.md` (note: the task
brief names `2026-08-08-ci-restructure-design.md`; the actual file on disk is dated
`2026-08-07`. Using the file that exists.) §4.1 (target `ci.yml` table + 27/2/29 count),
§4.4 (`nightly.yml`/`release.yml`), §4.5 (deletions + six-control table).

**Headline result: the reconciliation does NOT close against spec §4.1's stated 27+2=29 /
11 counts.** See §"Reconciliation" at the end for the exact discrepancy — this report does
not fudge the numbers to make them fit. The per-file ledger below is the ground truth the
reconciliation is built from.

---

## 1. `api-contract.yml` (155 lines, 6 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `backend-codegen-drift` | `go generate ./...` + diff | registry `codegen-drift-backend` → `ci.yml:verify` | Yes — `registry.go:203-211`, `Argv: bash scripts/check-codegen-drift-backend.sh`, same diff logic |
| `frontend-codegen-drift` | `pnpm run gen:api` + diff | registry `codegen-drift-frontend` → `ci.yml:verify` | Yes — `registry.go:213-220` |
| `openapi-lint` | redocly lint v1 spec | registry `openapi-lint-v1` → `ci.yml:verify` | Yes — `registry.go:223-237` |
| `openapi-lint` | redocly lint internal-e2e spec | registry `openapi-lint-e2e` → `ci.yml:verify` | Yes — `registry.go:240-257` |
| `spec-base-path-gate` | `api-lint -only PATH-BASE-PREFIX` v1 | registry `api-lint-base-path-v1` → **DELETED** (six-control #2, spec §4.5) | Yes — evidence in spec matches `scripts/api-lint/main.go:21,64-67`; `-only` is a filter not a mode |
| `spec-base-path-gate` | `api-lint -only PATH-BASE-PREFIX` internal-e2e | registry `api-lint-base-path-e2e` → `ci.yml:verify` | Yes — `registry.go:165-174`, survives per spec §4.5 note ("`-strict` never touches internal-e2e.yaml") |
| `api-design-system-lint` | `api-lint -strict` | registry `api-lint-strict` → `ci.yml:verify` | Yes — `registry.go:176-182` |
| `api-design-system-lint` | `api-lint` selftest (`go test ./scripts/api-lint/...`) | registry `api-lint-selftest` → `ci.yml:verify` | Yes — `registry.go:184-190` |
| `problem-codes-freshness` | `problem-codes-dump -check` | registry `problem-codes-fresh` → `ci.yml:verify` | Yes — `registry.go:137-148` |
| `contract-sync` | `check-contract-sync-all.ps1` | registry `contract-sync` → `ci.yml:verify` | Yes — `registry.go:192-201` |

`ci.yml:verify`'s own steps ("Install oasdiff", "Fetch base ref", "Materialize base-branch
spec") are copied verbatim from what was this job's prerequisite scaffolding — confirmed by
reading `ci.yml:38-70`, whose comments say exactly that.

## 2. `e2e-coverage-gate.yml` (194 lines, 3 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `coverage-map-check` | "Check all invariants have ≥1 spec ID" (grep) | registry `invariant-coverage-map` → `ci.yml:verify` | Yes — `registry.go:388-397`, `Argv: bash scripts/check-invariant-coverage-map.sh` |
| `coverage-map-check` | "Check PR template checkbox present" | **DELETED** (six-control #5, spec §4.5) — self-documented as "a reminder, not a coverage guarantee" | Yes — matches file's own header comment |
| `axe-baseline-check` | baseline reviewer-field + critical-violation checks | `nightly.yml:axe` | Yes — read `nightly.yml:251-298`; steps are byte-identical, only the `needs: coverage-map-check` dependency is dropped (job doesn't exist in nightly.yml) |
| `e2e-smoke` | Playwright approval flows + axe diff | `nightly.yml:e2e` | Yes — read `nightly.yml:126-249`; `go run ./cmd/api` (broken — that path does not exist) is fixed to `go run ./apps/api/cmd/metaldocs-api` at `nightly.yml:166`, matching spec §4.5's "must be fixed rather than deleted" ruling |

## 3. `fe-ci.yml` (34 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `web-typecheck-test` | `verify --only=fe-typecheck,fe-test` | registry `fe-typecheck`, `fe-test` → `ci.yml:verify` | Yes — `registry.go:520-541` |

## 4. `golangci-lint.yml` (27 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `golangci-lint` | golangci-lint-action, `only-new-issues: true` | `ci.yml:lint-go` (non-registry) | Yes — `ci.yml:186-216`. Note: `only-new-issues: true` is **still present** in `ci.yml:lint-go` today (line 215) — Step 4 of the plan (drop it) has not run yet, consistent with this being Step 1 only. The job is explicitly labelled TRANSITIONAL in `ci.yml:201-211`. |

## 5. `governance-check.yml` (83 lines, 4 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `check` | `check-governance.ps1` | registry `governance-diff-rules` → `ci.yml:verify` | Yes — `registry.go:375-386`, `Argv: pwsh ... check-governance.ps1`, no Paths (repo-wide, matches original) |
| `check` | `verify --only=adr-status` | registry `adr-status` → `ci.yml:verify` | Yes — `registry.go:334-342` |
| `wiki-tally` | `verify --only=wiki-tally` | registry `wiki-tally` → `ci.yml:verify` | Yes — `registry.go:344-352` |
| `db-dictionary-coverage` | `verify --only=db-dictionary` | registry `db-dictionary` → `ci.yml:verify` | Yes — `registry.go:354-362` |
| `docx-v2-isolation` | CK5-path guard, `if:` gated on PR title/branch name | **DELETED** (six-control #4, spec §4.5) — fires only when the PR self-identifies | Yes — confirmed the `if:` condition at line 69 matches spec's citation `governance-check.yml:69` exactly |

## 6. `invariants.yml` (130 lines, 3 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `cilint` | `go run ./tools/cilint --sarif` | registry `cilint` → `ci.yml:verify` | Yes — `registry.go:116-121`, `Argv: go run ./tools/cilint ./...` |
| `cilint` | "Upload SARIF" (`continue-on-error: true`, advisory telemetry) | **NO SUCCESSOR** — dropped silently | Not evidenced anywhere as intentional. Low severity: the step was itself `continue-on-error: true` and non-blocking, so no gate is lost, only a SARIF/Security-tab artifact. Flagging per the brief's "do not fudge" instruction rather than omitting it. |
| `migration-gapless` | gapless-sequence + no-historical-edit check (raw bash) | registry `migration-gapless` → `ci.yml:verify` | Yes — `registry.go:364-373`, `Argv: bash scripts/check-migration-gapless.sh`, `Needs: needsGitDepth`. **This is the control the brief calls out as "the one nearly lost before" — confirmed present and correctly targeted.** |
| `staticcheck` | `verify --only=gofmt,go-vet,go-vet-integration,staticcheck` (4 checks in 1 step) | registry `gofmt`, `go-vet`, `go-vet-integration` → `ci.yml:verify`; registry `staticcheck` → **DELETED** (separate from the six-control table — this is the 3rd registry deletion named in brief Step 3 / spec §4.6, distinct from `go-build`/`api-lint-base-path-v1`) | Yes — `registry.go:87-132`. Deletion evidence: §4.6 — golangci-lint becomes the sole umbrella, `.golangci.yml` already enables `staticcheck`, so standalone `staticcheck` is dropped for redundancy, not brokenness. |

## 7. `lint.yml` (60 lines, 3 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `eslint` | `verify --only=fe-eslint` | registry `fe-eslint` → `ci.yml:verify` | Yes — `registry.go:486-497` |
| `css-token-discipline` | `verify --only=css-token-discipline` | registry `css-token-discipline` → `ci.yml:verify` | Yes — `registry.go:499-508` |
| `eigenpal-selector-pin` | `verify --only=eigenpal-selector-pin` | registry `eigenpal-selector-pin` → `ci.yml:verify` | Yes — `registry.go:510-519` |

## 8. `module-boundaries.yml` (21 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `conformance` | `verify --only=module-boundaries,test-discipline,test-discipline-selftest` | registry `module-boundaries`, `test-discipline`, `test-discipline-selftest` → `ci.yml:verify` | Yes — `registry.go:279-305` |

## 9. `openapi-breaking.yml` (53 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `oasdiff-breaking` | install oasdiff, materialize base spec, `oasdiff breaking ... --fail-on ERR` | registry `oasdiff-breaking` → `ci.yml:verify` | Yes — `registry.go:260-275`. The base-spec materialization moved into `ci.yml:verify`'s own prerequisite steps (`ci.yml:65-70`), matching the registry comment "base-branch spec... materialized by a workflow prerequisite step (`ci.yml:lint-contract`...)" — job name in that comment is stale (says `lint-contract`, actual job is `verify`) but the step itself is present and correct. |

## 10. `perf.yml` (127 lines, 2 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `perf-reduced` (PR-gated, `KNOWN RED`) | k6 submit/signoff/publish, reduced | **NO SUCCESSOR** (by design) | Yes — `nightly.yml:22-26` comment states this explicitly: "its pull_request-gated perf-reduced sibling has no successor here: nightly has no pull_request trigger to reduce for." Consistent with spec §4.4's rationale for moving perf off `pull_request` entirely (a check "guaranteed red on every approval-touching PR"). |
| `perf-full` (push/manual) | k6 full suite + issue-on-failure | `nightly.yml:perf` | Yes — `nightly.yml:13-125`, moved verbatim per its own comment, still `KNOWN RED` (no postgres service, `secrets.PERF_DATABASE_URL` undefined) |

## 11. `phase3-hardening-gate.yml` (28 lines, 1 job) — WHOLE FILE, NO SUCCESSOR

| Job | Step | Successor | Verified |
|---|---|---|---|
| `hardening` | install gosec/govulncheck, run `scripts/phase3-hardening-gate.ps1` | **NONE** — file dies with no successor (spec §4.5) | Yes — spec cites "D-17/D-19 — two of four steps duplicate other checks, one targets a suite deleted in `dc0572f6`, the fourth disables its only vulnerability scanner by default parameter." `scripts/phase3-hardening-gate.ps1` is also in the plan's explicit script-deletion list. |

## 12. `release-readiness.yml` (55 lines, 1 job) — WHOLE FILE, NO SUCCESSOR

| Job | Step | Successor | Verified |
|---|---|---|---|
| `readiness` | `workflow_dispatch` wrapper around `scripts/phase3-release-readiness.ps1`, writes to `non_git/` | **NONE** — file dies with no successor (spec §4.5) | Yes — spec: "an operator runbook wearing a workflow costume." `scripts/release-readiness.ps1` named in plan's script-deletion list (note: the actual script invoked is `scripts/phase3-release-readiness.ps1`, not `scripts/release-readiness.ps1` — a naming mismatch between the plan's Step 2 `git rm` list and the workflow's real `run:` line; flagging since Step 2 will otherwise miss the real file). |

## 13. `req-traceability.yml` (44 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `gate` | `verify --only=req-trace-selftest,req-trace` | registry `req-trace-selftest`, `req-trace` → `ci.yml:verify` | Yes — `registry.go:636-659` |

## 14. `secret-scan.yml` (36 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `gitleaks` | full-history gitleaks scan via Docker | `ci.yml:security` (non-registry) | Yes — `ci.yml:146-161`, same image tag `v8.24.3`, same `--redact -v --exit-code 1` flags |

## 15. `supply-chain.yml` (67 lines, 3 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `sbom` (tag-triggered) | anchore/sbom-action, upload artifact | `release.yml:sbom` | Yes — `release.yml:10-27`, moved verbatim, redundant `if: startsWith(github.ref, 'refs/tags/')` guard correctly dropped since the workflow trigger already scopes to tags |
| `cve-scan` | anchore/scan-action (grype) + SARIF upload | `ci.yml:security` (non-registry) | Yes — `ci.yml:162-184`, same action, same `severity-cutoff: high`, plus a fork-PR SARIF-upload guard added |
| `dependabot-label` | applies `needs-staging-soak-7d` label | **DELETED** (six-control #6, spec §4.5) — no staging environment exists, nothing enforces the label | Yes — `smoke.yml` (kept file) confirms no staging environment claim |

## 16. `test-full.yml` (40 lines, 1 job)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `full` | `verify --only=go-test-integration` against postgres service | registry `go-test-integration` → `ci.yml:test-integration` | Yes — `registry.go:608-632`, `CIJob: "ci.yml:test-integration"`; `ci.yml:97-144` confirms the same postgres service block, `--require-infra`, 20-minute timeout, and per spec §4.2 now gated behind `needs: [verify]` (staged, not deferred to nightly) |

## 17. `test-nightly.yml` (48 lines, 1 job) — **GAP: NO SUCCESSOR FOUND**

| Job | Step | Successor | Verified |
|---|---|---|---|
| `nightly` | `go test -tags integration -count=1 -race -timeout 3600s ./tests/... ./internal/... ./apps/...` with `INTEGRATION_STRESS_N=500` | **NONE FOUND** | Checked. `nightly.yml` (421 lines, read in full) has exactly 4 jobs: `perf`, `e2e`, `axe`, `security-scan`. None of them runs the stress suite, none set `INTEGRATION_STRESS_N`, and a repo-wide grep for `stress`/`cross-platform` outside this one file returns zero matches. Spec §4.4's prose lists `nightly.yml`'s scope as "Perf, e2e, axe, cross-platform, gosec/govulncheck" — "cross-platform" does not match this job's actual content (a Linux-only stress run, not a cross-platform matrix) and no cross-platform job exists anywhere in the repo either. This looks like the stress suite was silently dropped in the actual `nightly.yml` implementation despite spec prose implying something should have landed here. |
| `nightly` | "Open issue on failure" | **NONE FOUND** — same gap | Same as above |

This is a genuine, unresolved control loss and is called out explicitly per the task's hard
gate. It should be adjudicated before Step 2 deletes this file — either the stress job needs
a home in `nightly.yml`, or its removal needs to be an evidenced, named decision the way the
six controls in spec §4.5 are.

## 18. `test-smoke.yml` (48 lines, 2 jobs)

| Job | Step | Successor | Verified |
|---|---|---|---|
| `unit` | `verify --only=go-build,go-test-unit` | `go-build` → **DELETED** (six-control #1, spec §4.5); `go-test-unit` → `ci.yml:verify` | Yes — `registry.go:80-85` (go-build), `registry.go:601-606` (go-test-unit). Deletion evidence: `go test ./...` and `go vet ./...` both already compile every package; `go build ./...` links nothing extra. |
| `smoke` | `go test -run "TestTriggerBypass\|TestMembership\|TestSchemaLockdown\|TestLegacy\|TestE2E" ./tests/integration/scenarios/` | **DELETED** (six-control #3, spec §4.5) — hand-listed regex subsumed by `go-test-integration`'s superset of paths | Yes — the exact regex matches spec's citation |

---

## Reconciliation

### What spec §4.1 claims

> 7 + 5 + 4 + 5 + 1 + 4 + 1 = **27 placed, 2 deleted** (totalling **29**). Non-registry
> controls: **11 placed.**

### What `tools/verify/registry.go` actually contains, today

`grep -c 'ID:' tools/verify/registry.go` → **43** distinct registry IDs (verified by listing
all 43 `ID:` values). Every one of the 43 already carries a `CIJob` value (mostly
`ci.yml:verify`, two `ci.yml:test-integration`/`nightly.yml:security-scan`), which means Step
3's "retarget every CIJob" is largely already done on disk — Step 3 as scoped by the plan is
really "delete `go-build`/`api-lint-base-path-v1`/`staticcheck`," not a from-scratch
retargeting.

**43 ≠ 29.** The gap is 14 registry IDs. Walking spec §4.1's own table against the actual
registry shows why:

**Table 4.1's "11 non-registry controls" column is itself wrong** — 7 of its 11 rows name
controls that already have a `tools/verify` registry ID (confirmed by reading each entry):

| Spec's non-registry label | Actual registry ID(s) | CIJob |
|---|---|---|
| `openapi-lint` (redocly) | `openapi-lint-v1` **and** `openapi-lint-e2e` (2 IDs, spec's table counts this as 1 row) | `ci.yml:verify` |
| `backend-codegen-drift` | `codegen-drift-backend` | `ci.yml:verify` |
| `frontend-codegen-drift` | `codegen-drift-frontend` | `ci.yml:verify` |
| `oasdiff-breaking` (listed again here, separately from lint-contract's registry column) | `oasdiff-breaking` | `ci.yml:verify` |
| `Migration monotonicity check` | `migration-gapless` | `ci.yml:verify` |
| `check-governance.ps1` | `governance-diff-rules` | `ci.yml:verify` |
| invariant→spec-ID coverage grep | `invariant-coverage-map` | `ci.yml:verify` |

Only 4 of the 11 rows are genuinely non-registry, workflow-native controls with no
`tools/verify` ID at all: `dorny/paths-filter` (new — see below), `golangci-lint`,
`gitleaks`, `grype`.

**And 6 more registry IDs are absent from the table under any name, in either column**:
`test-discipline-selftest`, `testdb-bypass-guard`, `gosec`, `govulncheck`,
`required-gate-selftest`, `verify-audit`. Of these, `test-discipline-selftest` **is** sourced
from one of the 18 deleted files (`module-boundaries.yml`, confirmed in ledger row 8 above);
the other five are pre-existing/self-referential registry entries not sourced from any of the
18 files under deletion, so they are correctly out of this ledger's per-file scope — but they
still inflate the registry.go total past 29, so any reconciliation that stops at "27+2=29"
undercounts the live registry by construction.

**Arithmetic that closes:** 27 (table) + 2 (six-control registry deletions) + 8 (registry IDs
hiding inside the mislabeled 7 "non-registry" rows, since `openapi-lint` covers 2 IDs) + 6
(registry IDs absent from the table entirely) = **43**. That matches the real registry.go
count exactly — the 29 in spec §4.1 is not wrong about which controls survive, it is wrong
about which bucket (registry vs. non-registry) 7 of them belong in, and it omits 6 more
entirely.

### Non-registry controls actually sourced from the 18 deleted files

Excluding the 7 misclassified rows above (which are registry-backed) and `dorny/paths-filter`
(new infrastructure — `ci.yml` has no `changes` job and no dorny action anywhere; `ci.yml`'s
own header comment explains the design pivoted to registry-`Paths`-based diff-scoping instead
of a job-level path filter, so this row in spec's table describes a design that was not built,
not a control that was lost), the genuine non-registry successors found by reading the 18
files and their successors are:

1. `golangci-lint` → `ci.yml:lint-go`
2. `gitleaks` → `ci.yml:security`
3. `grype` (cve-scan) → `ci.yml:security`
4. `sbom` → `release.yml:sbom`
5. `perf-full` → `nightly.yml:perf`
6. `e2e-smoke` → `nightly.yml:e2e`
7. `axe-baseline-check` → `nightly.yml:axe`

**7, not 11.**

### Bottom line

The ledger **does not reconcile** to spec §4.1's stated 27 registry-placed + 2 deleted = 29 /
11 non-registry numbers. It reconciles to a different, evidenced split:
- **32 distinct registry IDs** placed at `ci.yml:verify` or `ci.yml:test-integration`,
  sourced directly from the 18 files under deletion (full list: `codegen-drift-backend`,
  `codegen-drift-frontend`, `openapi-lint-v1`, `openapi-lint-e2e`, `api-lint-base-path-e2e`,
  `api-lint-selftest`, `api-lint-strict`, `problem-codes-fresh`, `contract-sync`,
  `invariant-coverage-map`, `fe-typecheck`, `fe-test`, `adr-status`, `governance-diff-rules`,
  `wiki-tally`, `db-dictionary`, `cilint`, `migration-gapless`, `gofmt`, `go-vet`,
  `go-vet-integration`, `fe-eslint`, `css-token-discipline`, `eigenpal-selector-pin`,
  `module-boundaries`, `test-discipline`, `test-discipline-selftest`, `oasdiff-breaking`,
  `req-trace-selftest`, `req-trace`, `go-test-integration`, `go-test-unit`).
- **3 registry IDs deleted**, not 2: `go-build` and `api-lint-base-path-v1` (six-control
  table, spec §4.5) **plus** `staticcheck` (spec §4.6 / plan Step 3's own instruction —
  outside the six-control table but still a registry deletion this task carries out).
- **7 non-registry controls** placed (`golangci-lint`, `gitleaks`, `grype`, `sbom`,
  `perf-full`→`nightly:perf`, `e2e-smoke`→`nightly:e2e`, `axe-baseline-check`→`nightly:axe`),
  not 11.
- **2 whole files with no successor by design**: `phase3-hardening-gate.yml`,
  `release-readiness.yml` (spec §4.5, evidenced).
- **1 job with no successor by design**: `perf.yml`'s `perf-reduced` (evidenced in
  `nightly.yml`'s own comment).
- **1 silently dropped advisory step**: `invariants.yml:cilint`'s SARIF upload (low severity —
  was itself non-blocking, but not named anywhere as an intentional drop).
- **1 unresolved gap requiring adjudication before Step 2**: `test-nightly.yml`'s stress job
  (`INTEGRATION_STRESS_N=500`, `-race`, 60-minute timeout) has no successor anywhere in
  `nightly.yml`. This is the same shape of loss the brief warns about for
  `invariants.yml`'s migration-monotonicity gate — except unlike that gate (confirmed present,
  see ledger row 6), this one is confirmed **absent**.
- **1 script-name mismatch worth fixing before Step 2's `git rm`**:
  `release-readiness.yml` invokes `scripts/phase3-release-readiness.ps1`, but the plan's Step
  2 script-deletion list (and Task 12's Files section) names `scripts/release-readiness.ps1`
  — a file with a different name. If a file called exactly `scripts/release-readiness.ps1`
  does not exist, that `git rm` will simply fail loudly (safe); if it exists as a stale
  duplicate alongside `phase3-release-readiness.ps1`, both should be checked.

### Migration-monotonicity gate — explicitly confirmed

The brief calls this out as "the one nearly lost before." Confirmed present: registry ID
`migration-gapless` (`tools/verify/registry.go:364-373`) wraps
`scripts/check-migration-gapless.sh`, runs the identical gapless-sequence + no-historical-edit
logic that lived inline in `invariants.yml`'s `migration-gapless` job, and its `CIJob` field
already reads `ci.yml:verify`. `ci.yml:verify`'s own `go run ./tools/verify --require-infra
--profile=changed` step (line 86) executes it. This control is not lost.

### Recommendation

Do not proceed to Step 2 (`git rm`) until an operator adjudicates:
1. `test-nightly.yml`'s stress suite — give it a home in `nightly.yml` or record an evidenced
   reason it dies (spec §4.5 style), the same bar every other deletion in this ledger met.
2. The `scripts/release-readiness.ps1` vs. `scripts/phase3-release-readiness.ps1` naming
   mismatch, so Step 2's `git rm` targets the file that actually exists.
3. Whether spec §4.1's 27+2/11 counts should be corrected in the spec itself (the controls
   are not lost, but the accounting that was supposed to prove that is wrong), so future
   readers of §4.1 don't re-derive a false "everything's accounted for" belief from a table
   that undercounts by 14 registry IDs and overcounts non-registry by 4.

Everything else in this ledger — every job, every step, in all 18 files — has a verified
successor or a verified, evidenced reason to die with none.
