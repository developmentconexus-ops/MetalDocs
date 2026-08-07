# Lane: cicd

## 1. Workflow inventory

| Workflow | Trigger | Asserts | Blocking? | Runtime (inferred) | Pre/post-merge |
|---|---|---|---|---|---|
| `ci.yml` (docx-renderer CI) | PR paths: docx-renderer/templates/featureflags | pnpm typecheck/test/build:docx-v2; `go build ./...`; scoped `go test` (platform config/featureflags/servicebus/ratelimit/objectstore + templates + docx_v2) | Blocking (PR check, no continue-on-error) | ~3-6 min (node+go jobs) | Pre-merge |
| `lint.yml` | PR paths `**/*.ts,tsx,module.css`, eslint config, lockfile; push main | `pnpm run lint` (ESLint incl. Eigenpal ACL boundary rule, ADR 0046); `check-css-token-discipline.sh` (FE-18 hex-literal ban) | Blocking | ~2-4 min | Pre-merge |
| `golangci-lint.yml` | PR (all); push main | golangci-lint v2.11, `only-new-issues: true`, scoped to `./apps/api/... ./internal/... ./tools/...` | Blocking | ~2-5 min | Pre-merge |
| `module-boundaries.yml` | PR to main | `check-module-boundaries.ps1` (layer-only, no cycle detection — already known); `check-test-discipline.sh` | Blocking | <1 min | Pre-merge |
| `phase3-hardening-gate.yml` | PR to main | `phase3-hardening-gate.ps1`: `go test ./...`, module-boundaries, `contract-baseline.ps1`, `security-baseline.ps1` (gosec/govulncheck, `-SkipGovulncheck` default true) | Blocking (job fails on any non-zero) | ~5-10 min (installs gosec+govulncheck from `@latest` each run) | Pre-merge |
| `release-readiness.yml` | `workflow_dispatch` only | `phase3-release-readiness.ps1`: governance-check + hardening-gate chained | Blocking within its own run, but never runs automatically | ~8-15 min | Neither — manual, not on any PR/push path |
| `supply-chain.yml` | push main/tags; weekly cron Mon 04:00 UTC; PR on `go.mod/go.sum/package.json/pnpm-lock.yaml` | SBOM (Syft, tag-only); CVE scan (Grype, `fail-build: true`, severity-cutoff high); dependabot PR labeling | CVE scan blocking on the PR path it triggers on; SBOM only on tags (never gates a PR) | ~2-5 min | Mixed: CVE scan is pre-merge (PR-triggered) for lockfile changes only; SBOM is post-tag |
| `secret-scan.yml` | every push; every PR | gitleaks v8.24.3 full-history scan (`--exit-code 1`), allowlist in `.gitleaks.toml` (32 lines) | Blocking | ~1-3 min (full history `fetch-depth: 0`, grows with repo size) | Pre-merge |
| `test-smoke.yml` | PR to main/develop | `go build ./...`; `go test -count=1 -timeout 600s ./...` (unit only, 12 min cap); 4 named integration tests only (`TestTriggerBypass\|TestMembership\|TestSchemaLockdown\|TestLegacy\|TestE2E`) against a fresh `postgres:16` service, 2 min cap | Blocking | ~5-12 min | Pre-merge |
| `openapi-breaking.yml` | PR touching `api/openapi/v1/openapi.yaml` | oasdiff (installed `@latest`) breaking-change diff base-vs-head, `--fail-on ERR` | Blocking, explicitly no continue-on-error (documented in-file) | ~1-2 min | Pre-merge |
| `req-traceability.yml` | PR paths (target-arch doc, req-trace-map, tests); push main | `go test ./scripts/req-trace/...`; `go run ./scripts/req-trace` — every MUST REQ ID needs resolvable evidence, committed traceability report must equal fresh regen | Blocking | ~1-3 min | Pre-merge (+ push main) |
| `governance-check.yml` | PR to main | `check-governance.ps1`: 3 diff-based rules (contract change needs OpenAPI update; domain change needs `tests/` update; infra/scripts change needs runbook update); `check-adr-status.sh` (ADR status block ≤3 lines/400 chars); conditional `docx-v2-isolation` job (CK5 path guard on docx-v2-titled PRs) | Blocking | <1 min | Pre-merge |
| `smoke.yml` | `workflow_dispatch` only | `ops/smoke/healthz.sh` against an operator-supplied `base_url` | N/A — never auto-runs; workflow comment states no hosted env exists yet | ~1 min | Neither |
| `fe-ci.yml` | PR paths (frontend/apps/web, packages, lockfiles); push main same paths | `pnpm run typecheck` (tsc --noEmit); `pnpm run test` (vitest) for `@metaldocs/web` | Blocking | ~3-6 min | Pre-merge |
| `invariants.yml` | PR (all); push main | `cilint` custom Go linters (SARIF upload advisory, then a real blocking re-run); migration gapless/no-historical-edit check; `go vet ./...`; staticcheck 2025.1.1 | Blocking (cilint fail step + staticcheck + go vet all hard) | ~2-5 min | Pre-merge |
| `api-contract.yml` | PR paths (openapi, api.gen.go, api-types, lockfiles, `internal/modules/**`, wiki authz docs, etc. — very broad path list); push main | backend codegen drift (`go generate` + diff); frontend codegen drift (`pnpm gen:api` + diff); Redocly lint ×2 specs (`@redocly/cli@latest`); `PATH-BASE-PREFIX` api-lint rule ×2; full api-lint guard suite `-strict` (registry-binding, dialect bans, envelope/authz/pagination-drift — "0 on a clean tree, no deferral tier" per in-file comment); problem-code freshness (`cmd/problem-codes-dump -check`); `check-contract-sync-all.ps1` over 4 of 5 generated-boundary modules (`approval` excluded) | All 7 jobs blocking, no continue-on-error | ~10-20 min total across jobs (parallel) | Pre-merge |
| `e2e-coverage-gate.yml` | PR paths (e2e, approval FE, documents/jobs modules, `api/**`); push main | `coverage-map-check`: every COVERAGE.md invariant row must have ≥1 spec ID (real, blocking) + a PR-body checkbox grep that only fires when COVERAGE.md itself is touched (self-documented advisory, see file header); `axe-baseline-check`: baseline entries need `approved_by`/`reason`, and zero `critical`-impact entries allowed; `e2e-smoke`: Playwright approval-flow run + `axe-diff.mjs` (the axe-diff step is `|| true` — failures are swallowed, only visible via uploaded artifact) | Mixed: coverage-map "all invariants mapped" and axe-baseline checks are real gates; the PR-checkbox step and the axe-diff step are advisory/soft-fail by construction | ~8-15 min (Playwright install + browser run) | Pre-merge |
| `perf.yml` | PR paths (`internal/modules/approval/application/**`, `api/handlers/approval/**`); push main; manual w/ full-suite toggle | k6 benchmarks (submit/signoff/publish reduced on PR; + scheduler_tick full on push/manual) against a live `go run` backend | No pass/fail threshold visible in the workflow (k6 scripts not read in this pass — thresholds, if any, live inside the `.js` files) — treat as **unverified whether it can fail the build** | ~5-10 min reduced / longer full | Mixed (reduced set pre-merge, full set post-merge/manual) |
| `test-full.yml` | push main only | `go test -tags integration -count=1 -race -timeout 900s ./tests/... ./internal/... ./apps/...` — the entire integration suite, with `-race` | Blocking for the push (but a push to main is already merged) | up to 20 min cap | **Post-merge only** |
| `test-nightly.yml` | cron `0 2 * * *`; manual | Same full integration tree at `INTEGRATION_STRESS_N=500`, `-race`, 3600s cap; opens a GitHub issue on failure | Blocking within the run, but detached from any PR/push | up to 60 min cap | Post-merge, scheduled only |

Note on `perf.yml`: I did not open `tools/perfbench/*.js` to confirm k6 threshold config — flagged as `unverified` below rather than asserted.

## 2. Coverage matrix

| Defect class | Gate | Pre/post-merge | Notes |
|---|---|---|---|
| Contract drift (OpenAPI ↔ generated code) | `api-contract.yml` (`backend-codegen-drift`, `frontend-codegen-drift`) | Pre-merge | Full regen + `git diff --exit-code` |
| Codegen drift (other generated artifacts) | `api-contract.yml` (`problem-codes-freshness`) | Pre-merge | `cmd/problem-codes-dump -check` |
| OpenAPI breaking changes | `openapi-breaking.yml` | Pre-merge | oasdiff base-vs-head |
| Module boundaries (layer only, no cycles) | `module-boundaries.yml` | Pre-merge | Already known: no cycle detection |
| DB schema/migration monotonicity | `invariants.yml` (`migration-gapless`) | Pre-merge | Gapless sequence + no historical edits. Directory is currently empty post-fold, so the gapless branch is a no-op until the next migration is filed |
| DB schema baseline (bootstrap/dictionary/equivalence checks) | **No workflow** | — | `check-db-bootstrap.ps1`, `check-db-dictionary-coverage.ps1`, `check-baseline-equivalence.ps1` all exist, none referenced by any `.yml` (matches what was already known) |
| AuthZ/tripwire parity, capability registry binding | `api-contract.yml` (`api-design-system-lint` — api-lint `-strict`) | Pre-merge | Per in-file comment, "0 on a clean tree… no deferral tier" |
| Contract-sync (module-level generated boundary) | `api-contract.yml` (`contract-sync`) | Pre-merge | Only 4 of 5 generated-boundary modules; `approval` explicitly excluded pending M9/F9.5 |
| Secrets | `secret-scan.yml` | Both (runs on every push and PR) | gitleaks full-history scan |
| Vulnerabilities (deps) | `supply-chain.yml` (`cve-scan`) | Pre-merge only when `go.mod/go.sum/package.json/pnpm-lock.yaml` changed; also push/tag/weekly | Grype, fail-build true, high severity cutoff |
| Vulnerabilities (Go-specific, govulncheck) | `phase3-hardening-gate.yml` (via `security-baseline.ps1`) | Pre-merge | `-SkipGovulncheck` **defaults to `$true`** in the gate itself — so govulncheck is opted out by default even where wired; only `release-readiness.yml` (manual dispatch) exposes the toggle |
| Static analysis / lint | `golangci-lint.yml`, `invariants.yml` (staticcheck, go vet, cilint) | Pre-merge | `only-new-issues: true` on golangci-lint — pre-existing findings in touched files are not surfaced |
| Tests — unit (Go) | `test-smoke.yml` | Pre-merge | Full `go test ./...`, no `-race`, no `-tags integration` |
| Tests — integration (Go) | `test-smoke.yml` (4 named tests only); `test-full.yml` (entire suite) | 4 tests pre-merge; **everything else post-merge only** | See §6 |
| Tests — nightly stress | `test-nightly.yml` | Post-merge, scheduled | N=500, `-race` |
| Tests — frontend unit | `fe-ci.yml` | Pre-merge | vitest + tsc |
| Tests — E2E | `e2e-coverage-gate.yml` (`e2e-smoke`) | Pre-merge, but path-gated (only fires on e2e/approval/documents/jobs/`api/**` paths) | Playwright approval flows only — no general E2E coverage |
| Frontend type/lint | `fe-ci.yml`, `lint.yml` | Pre-merge | tsc, ESLint, CSS token discipline |
| Frontend accessibility | `e2e-coverage-gate.yml` (`axe-baseline-check`, `axe-diff`) | Pre-merge, path-gated | axe-diff step is `|| true` (soft) |
| Perf/benchmarks | `perf.yml` | Reduced set pre-merge (path-gated to approval module only), full set post-merge/manual | Pass/fail threshold **unverified** (k6 script internals not read) |
| Docs/ADR governance | `governance-check.yml` (`check-adr-status.sh`), `req-traceability.yml` | Pre-merge | ADR status-block budget; REQ traceability |
| Backup/restore integrity | **No workflow** | — | `run-backup-restore-gate.ps1` (a real, well-built gate: backup → validate → restore → row-count smoke against `documents/document_versions/iam_users/iam_user_roles/audit_events/outbox_events`) exists and is never invoked by any workflow |
| eslint Eigenpal ACL boundary | `lint.yml` | Pre-merge | — |
| CK5/docx-v2 isolation | `governance-check.yml` (`docx-v2-isolation`) | Pre-merge, conditional on PR title/branch name | Fragile trigger — see §3 |

## 3. Gaps

**Defect classes with no gate at all:**
- DB bootstrap/dictionary-coverage/baseline-equivalence checks — scripts exist, zero wiring (already known, confirmed complete: all three).
- Backup/restore integrity — `run-backup-restore-gate.ps1` and its three dependents (`backup-postgres.ps1`, `validate-backup.ps1`, `restore-postgres.ps1`) are a complete, evidence-emitting gate with zero CI wiring. This is new relative to the three already-known scripts.
- System-runnable smoke (`check-system-runnable.ps1`) — no workflow reference.
- Wiki tally/doc-drift (`wiki-tally-check.ps1`) — no workflow reference.
- Eigenpal selector pin (`check-eigenpal-selector-pin.sh`) — no workflow reference, despite `lint.yml`'s header explicitly citing ADR 0046 Eigenpal ACL enforcement (that workflow enforces the *import-boundary* half via ESLint, not the *selector-pin* half via this script).
- release-v2 naming check (`check-release-v2-names.ps1`) — no workflow reference.

**Complete cross-reference: scripts in `scripts/` (44 `.ps1`/`.sh` files, excluding `api-lint/` and `req-trace/` subpackages and SQL/testdata) referenced by zero `.github/workflows/*.yml`:**
```
check-db-bootstrap.ps1            (already known)
check-db-dictionary-coverage.ps1  (already known)
check-baseline-equivalence.ps1    (already known)
run-backup-restore-gate.ps1
validate-backup.ps1
backup-postgres.ps1
restore-postgres.ps1
check-system-runnable.ps1
wiki-tally-check.ps1
check-eigenpal-selector-pin.sh
check-release-v2-names.ps1
classify-test-failure.sh
test-integration.ps1
test.ps1
e2e-smoke.ps1
e2e-seed.ps1
dev-db-reset.ps1
dev-migrate.ps1
dev-bootstrap.ps1
dev-bootstrap-baseline.ps1
dev-api.ps1
dev-api-perf.ps1
dev-local.ps1
dev-docx-renderer.ps1
start-api.ps1
start-worker.ps1
start-jobs.ps1
seed-system-blank-template.ps1
export-schema-baseline.ps1
openapi-lint-local.ps1
tidy.ps1
```
Of these, `dev-*`, `start-*`, `test.ps1`, `test-integration.ps1`, `openapi-lint-local.ps1`, `dev-db-reset.ps1`, `dev-migrate.ps1`, `tidy.ps1` are local-orchestration/dev-loop scripts and are not gates by intent — their absence from CI is expected, not a gap. The ones that read as verification/gate scripts by name (`check-*`, `validate-*`, `run-*-gate.ps1`, `classify-test-failure.sh`, `e2e-smoke.ps1`, `e2e-seed.ps1`) and are still unwired total **13**: the 3 already-known plus `run-backup-restore-gate.ps1`, `validate-backup.ps1`, `backup-postgres.ps1`, `restore-postgres.ps1`, `check-system-runnable.ps1`, `wiki-tally-check.ps1`, `check-eigenpal-selector-pin.sh`, `check-release-v2-names.ps1`, `classify-test-failure.sh`, `e2e-smoke.ps1`, `e2e-seed.ps1` (the last two duplicate what `e2e-coverage-gate.yml`'s `e2e-smoke` job does inline via raw `pnpm exec playwright` rather than calling these scripts — a second, unused implementation of the same concern).

Scripts confirmed **wired** (do not re-flag): `contract-baseline.ps1` and `security-baseline.ps1` (called from inside `phase3-hardening-gate.ps1`, itself called from `phase3-hardening-gate.yml` and `release-readiness.ps1`); `check-module-contract-sync.ps1` (called from `check-contract-sync-all.ps1`, called from `api-contract.yml`); `check-governance.ps1`, `check-adr-status.sh`, `check-module-boundaries.ps1`, `check-test-discipline.sh`, `check-css-token-discipline.sh`, `axe-diff.mjs`, `scripts/req-trace/*`, `scripts/api-lint/*`.

**`release-readiness.yml` is itself effectively unwired as a gate**: `workflow_dispatch` only, no `schedule`, no `pull_request`, no `push`. It chains governance + hardening + evidence upload but nothing triggers it automatically — it is a runbook step, not a gate.

**`docx-v2-isolation` job's trigger is a string match** on PR title containing `'docx-v2'` or head-ref starting with `feat/docx-v2-` (`governance-check.yml:33`) — a PR that touches the guarded CK5 paths without using that naming convention skips the check entirely. This is a gate that can be silently bypassed by naming, not by content.

## 4. Reproducibility

- **Unpinned tool installs (`@latest`)**: `gosec@latest`, `govulncheck@latest` (×2 workflows: `phase3-hardening-gate.yml:23-24`, `release-readiness.yml:35-36`); `oasdiff@latest` (`openapi-breaking.yml:44`); `@redocly/cli@latest` (×3 sites: `api-contract.yml:62,69`, and locally in `openapi-lint-local.ps1:14` — the local mirror is unpinned the same way). A gosec/govulncheck/oasdiff/redocly release landing between two CI runs on identical source can flip a PR from green to red with no code change.
- **Docker image tags not pinned to digest**: `postgres:16` (floating minor/patch) used in `test-smoke.yml`, `test-full.yml`, `test-nightly.yml`; `ghcr.io/gitleaks/gitleaks:v8.24.3` is the one properly pinned exact-version image in the set.
- **GitHub Actions pinned to major-version tags, not commit SHAs**: every `uses:` across all 20 workflows is `owner/action@vN` (e.g. `actions/checkout@v4`, `golangci-lint-action@v9`) — standard practice but not SHA-pinned, so a compromised or force-moved tag on any of ~12 distinct third-party actions is trusted silently. `golangci-lint` itself is version-pinned inside the action (`version: v2.11`) — the one place a tool version is explicitly locked.
- **Caching**: present and consistent — every Go job uses `actions/setup-go@v5` with `cache: true`; every Node/pnpm job uses `actions/setup-node@v4` with `cache: pnpm`. No caching gap found in the read workflows.
- **Nondeterministic steps**: `secret-scan.yml` does a full-history gitleaks scan (`fetch-depth: 0`) — runtime grows with repo history, not source-of-truth nondeterminism but an unbounded-runtime risk. `req-traceability.yml` also uses `fetch-depth: 0` for its commit-hash verification. `perf.yml`'s k6 runs against a live backend with no visible statistical-significance handling in the workflow file itself (thresholds live in the `.js` files, unverified).

## 5. Local↔CI parity

- **No single verify entry point.** Root `Makefile` (4 targets: `up`, `down`, `logs`, `test`) only wraps `docker compose` and frontend vitest (`cd frontend/apps/web && pnpm exec vitest run`). It does not run Go tests, lint, api-lint, contract-sync, or any of the governance/hardening gates.
- `package.json` (root) exposes 4 scripts: `build:docx-v2`, `test:docx-v2`, `typecheck:docx-v2`, `lint` (ESLint) — covers `ci.yml` and half of `lint.yml`, nothing else.
- The closest thing to a composite local gate is `scripts/phase3-hardening-gate.ps1` (go test + module-boundaries + contract-baseline + security-baseline) and `scripts/phase3-release-readiness.ps1` (governance + hardening) — both runnable locally via PowerShell, and both mirror what `phase3-hardening-gate.yml`/`release-readiness.yml` run in CI. This is real parity for that slice.
- Everything else — `golangci-lint.yml`, `invariants.yml` (cilint/staticcheck/go vet/migration-gapless), `api-contract.yml` (7 jobs), `openapi-breaking.yml`, `req-traceability.yml`, `e2e-coverage-gate.yml`, `fe-ci.yml`, `secret-scan.yml`, `supply-chain.yml` — has **no single local script wrapping it**. A developer/agent reproduces each by hand-running the underlying tool (`go run ./scripts/api-lint/ -strict ...`, `go run ./scripts/req-trace`, `npx @redocly/cli lint ...`, `go run ./tools/cilint ./...`, etc.), which is possible but requires already knowing the exact 20-workflow surface — there is no `make verify` / `make ci` that fans out to all of it.
- **The gap between "green locally" and "green in CI" is therefore the whole surface except**: docx-v2 (ci.yml has direct npm/go equivalents), frontend lint+CSS discipline (direct commands), and the phase3 hardening/release chain (has local PS1 mirrors). Everything path-gated, integration-tagged, or living only inside a workflow YAML step (api-lint, req-trace, cilint, openapi lint, codegen-drift diff, contract-sync, ADR-status, migration-gapless, axe-diff, oasdiff) has to be independently discovered and invoked — there is no manifest of "these are the N checks CI runs, run them all."

## 6. Feedback latency — what only fails after merge

- **`test-full.yml`** (`push: branches: [main]`) is the only place the entire `-tags integration` suite runs: `go test -tags integration -count=1 -race -timeout 900s ./tests/... ./internal/... ./apps/...`. `test-smoke.yml` (PR-time) runs only 4 named integration tests (`TestTriggerBypass|TestMembership|TestSchemaLockdown|TestLegacy|TestE2E`) scoped to `./tests/integration/scenarios/`, capped at 2 minutes.
- Quantified: `grep -rl "go:build integration"` finds **359** Go files carrying the integration build tag, against **1119** total `_test.go` files repo-wide (32%). `grep -rE "^func Test" tests/` finds **309** test functions under `tests/`. The PR-time smoke gate names 5 test-name patterns; the rest of those 309+ tests, and all coverage inside `internal/...` and `apps/...` compiled only under `-tags integration`, run for the first time **after the change is already on `main`**.
- `-race` is likewise post-merge only for the integration tree — a PR can merge with a race condition in integration-tagged code that only surfaces on `test-full.yml`, after merge, on `main`.
- `test-nightly.yml` (N=500 stress, `-race`, 60 min cap) is scheduled/manual only — further detached from any specific change.
- `perf.yml`'s full k6 suite (`scheduler_tick` included) runs on `push: main` / manual only; the PR-time perf gate is a "reduced" set and is also path-gated to `internal/modules/approval/application/**` and `api/handlers/approval/**` only — a performance regression anywhere else in the codebase has no pre-merge signal at all.
- `supply-chain.yml`'s SBOM job (`sbom`) only runs `if: startsWith(github.ref, 'refs/tags/')` — tag-time, i.e. release-time, not even push-to-main time.
- For a repo where an AI agent authors changes, this means: **roughly a third of the Go test surface, all `-race` integration coverage, most perf coverage, and SBOM generation give their first signal only after the change is unrevertable-by-review** (already merged). The agent's PR-time signal is unit tests + 5 named integration tests + all the static/contract/lint gates in §1 — real coverage, but structurally blind to whatever the other 300+ integration tests and the full-suite race detector would have caught.

## 7. Branch protection / required checks

- No `CODEOWNERS` file anywhere under `.github/` (confirmed: `.github/` contains only `PULL_REQUEST_TEMPLATE.md` and the 20 `workflows/*.yml`; no `CODEOWNERS`, no `dependabot.yml`, no `settings.yml`).
- `gh api repos/leandrotcawork/MetalDocs/branches/main/protection` returns **HTTP 403**: `"Upgrade to GitHub Pro or make this repository public to enable this feature."` — branch protection is inaccessible via API on the current plan/visibility, and (per that same message) may not even be an available feature on this repo tier. Which of the 20 workflows are marked "required" for merge is therefore **unverifiable from the repo** — there is no settings file, no CODEOWNERS, and no API access to confirm it. This is itself worth flagging: the entire blocking/advisory posture asserted in §1 is inferred from workflow trigger + exit-code behavior, not from GitHub's actual merge-gating configuration, which could disagree with every "blocking" label above.

## 8. Can an implementation change modify the gate that judges it?

**Yes, in the same commit surface, with no separation of powers found anywhere in this lane.**
- All 20 workflow YAMLs live under `.github/workflows/` in the same repo, same branch, no separate protected path.
- The hand-maintained allowlists that soften gates live beside product code and are editable by the same PR they'd otherwise block: `.gitleaks.toml` (32 lines, secret-scan allowlist), `scripts/api-lint/tripwire-allowlist.txt` and `scripts/api-lint/sole-rls-read-allowlist.txt` (already known: 5 files/221 lines total under `scripts/api-lint/`), `frontend/apps/web/e2e/axe-baseline.json` (accessibility baseline — `e2e-coverage-gate.yml` only checks it has `approved_by`/`reason` fields and no `critical` entries, not that the *fields* were reviewed by anyone other than the PR author), and the ADR status-block format checked by `check-adr-status.sh`.
- `governance-check.yml`'s own rules are pure diff pattern-matches (`check-governance.ps1`) that a PR could satisfy by adding an unrelated one-line file under `tests/`, `docs/runbooks/`, etc. — the check verifies presence of a path, not review of content, and both the check and a satisfying-but-hollow file can land in the same PR.
- No workflow in this set restricts who/what can edit `.github/workflows/**` itself (no separate approval gate on workflow-file changes visible in-repo — that would have to come from branch protection settings, which §7 established are unverifiable here).
- One partial counterexample: `phase3-hardening-gate.ps1` and `phase3-release-readiness.ps1` write timestamped JSON evidence files to `non_git/hardening/`, `non_git/release/`, etc., and `release-readiness.yml` uploads them as workflow artifacts — that produces an audit trail of what actually ran, which is a real (if soft) check on silently gaming the gate, but it doesn't prevent editing the gate's own logic in the same PR the change under review is in.

## The five heaviest, with detail

1. **A third of the test suite, and all `-race` integration coverage, is post-merge-only** (§6). 359 of 1119 `_test.go` files are integration-tagged; PR-time exercises only 5 named tests from that set. For an AI-authored-change repo this is the single biggest structural gap — the agent's fastest, cheapest feedback loop (the PR check) is blind to the majority of the correctness surface CI actually has, and the failure only surfaces once the change is already merged to `main`.

2. **A real backup/restore integrity gate exists, fully built, and is wired to nothing** (§3). `run-backup-restore-gate.ps1` → `backup-postgres.ps1` → `validate-backup.ps1` → `restore-postgres.ps1`, with a row-count smoke test across 6 core tables and JSON evidence output, is a complete disaster-recovery verification chain that no workflow ever calls. This blocks confidence in "can we actually restore from backup" as an ongoing, continuously-verified property rather than a one-time manual exercise.

3. **13 check-shaped scripts are unwired**, 10 beyond the 3 already known (§3): `check-system-runnable.ps1`, `wiki-tally-check.ps1`, `check-eigenpal-selector-pin.sh`, `check-release-v2-names.ps1`, `classify-test-failure.sh`, `e2e-smoke.ps1`, `e2e-seed.ps1`, plus the 3 backup-chain scripts. Each looks, by name and content, like it was built to be a gate; none is invoked by CI. This is the "gate that exists but doesn't fire" pattern at scale, not an isolated oversight.

4. **Security-tool versions and MFA-adjacent gates are unpinned or opted out by default** (§4, §2). `gosec`, `govulncheck`, `oasdiff`, and `@redocly/cli` all install via `@latest` in CI, so a PR can flip from green to red with zero source change when any of those 4 tools ships a release. Compounding it, `phase3-hardening-gate.ps1` defaults `-SkipGovulncheck` to `$true`, so even where the security gate *is* wired (`phase3-hardening-gate.yml`, every PR to main), govulncheck does not actually run unless the caller explicitly opts in — only the manual `release-readiness.yml` dispatch exposes the toggle.

5. **No verified separation between the gates and the code they judge** (§8). All 20 workflows, all allowlists (gitleaks, api-lint's two allowlist files, axe baseline), and all product code share one commit surface with no CODEOWNERS and unverifiable branch protection (§7, 403 on the GitHub API). A PR can simultaneously loosen the gate and satisfy it, and nothing in this repo's visible configuration prevents that — it may be prevented by GitHub branch-protection settings this lane could not read, which is itself the finding: the safety property is not verifiable from the repository.

## What is actually fine

- `api-contract.yml`'s 7-job design is genuinely thorough and internally self-aware — comments explicitly document what each job does and does not cover (e.g. the `problem-codes-freshness` job explaining why nothing else would notice staleness; the `api-design-system-lint` job noting there is no deferral tier). This is the best-documented workflow in the set and should not be touched casually.
- Caching is uniformly correct across all 20 workflows — every Go job caches via `setup-go@v5`, every Node job via `setup-node@v4` + pnpm. No workflow in this set needs a caching fix.
- `secret-scan.yml`'s full-history gitleaks scan with a documented, triaged allowlist (comment cites "Wave Z item Z-26 (F-18 round-2); zero real secrets found in history") is a properly reasoned, complete implementation — not a stub.
- `openapi-breaking.yml`'s design rationale comment (why push:main is intentionally excluded, why PR-only is sufficient given "every change to main goes through a PR") is a correct and explicit argument, not an unexamined gap.
- `e2e-coverage-gate.yml` is honest about its own limits: the file's opening comment states outright that the coverage-map checkbox check is advisory, not a real coverage guarantee, and explains exactly what would need to change to make it enforcing. This is the opposite of a hidden gap — it is a documented, intentional deferral.
- `req-traceability.yml`'s anti-rot design (committed report must match a fresh regeneration, not just "some report exists") is a correctly-built drift guard.
- Action-version pinning to major-version tags (not `@latest`) is consistent across all third-party GitHub Actions — the `@latest` reproducibility problem in this lane is confined to Go-installed/npx-installed CLI tools, not to the Actions themselves.

## Unverified / needs judgment

- Does `perf.yml`'s k6 suite (`tools/perfbench/*.js`) have actual pass/fail thresholds, or does the job only fail on backend-startup/script-crash? Not read in this pass — the workflow YAML shows no visible threshold gate.
- Is `release-readiness.yml` intended to ever be automated (e.g. on release-branch push), or is manual-dispatch-only its permanent, deliberate design? The workflow itself gives no indication either way.
- Are any of the 20 workflows marked "required" in GitHub's branch-protection UI? Completely unverifiable from this repo (§7) — the blocking/advisory classification in §1 is inferred from trigger + exit-code shape only.
- Is `.gitleaks.toml`'s allowlist reviewed on a cadence, or only grown reactively? The comment in `secret-scan.yml` describes one triage pass (Wave Z / F-18) but nothing enforces the allowlist doesn't silently accumulate.
- `docx-v2-isolation`'s title/branch-name trigger (§3) — is there a complementary path-based check elsewhere that would catch a mis-named PR touching CK5 paths? Not found in this lane's search, but worth a second look before calling it a confirmed gap.

## Commands run

```
Glob .github/workflows/*.yml                              → 20 files
Glob scripts/**/*                                          → 132 files (100 shown; enumerated further via targeted Glob)
Read: all 20 workflow YAMLs; Makefile; package.json;
      mechanical-enforcement-register.md;
      phase3-hardening-gate.ps1; phase3-release-readiness.ps1;
      check-governance.ps1; check-contract-sync-all.ps1;
      test.ps1; test-integration.ps1; run-backup-restore-gate.ps1;
      openapi-lint-local.ps1
grep -rn "check-release-v2-names|check-system-runnable|wiki-tally-check|check-eigenpal-selector-pin|test-integration\.ps1|scripts/test\.ps1|e2e-smoke\.ps1|e2e-seed\.ps1|run-backup-restore-gate|validate-backup|restore-postgres|backup-postgres|classify-test-failure|verify-triggers\.sql|replay-materialize-pdf-deadletters|dev-bootstrap-baseline|export-schema-baseline|openapi-lint-local|docx-v2-seed-minio|docx-v2-verify-migrations|seed-system-blank-template|dev-db-reset|dev-migrate\.ps1|tidy\.ps1" .github/
  → 0 matches (confirms unwired set)
grep -rn <same pattern, narrower> scripts/       → 12 files (self- and cross-references within scripts/)
grep -n "@latest" .github/workflows/             → 7 hits across 4 workflows
grep -n "uses: [a-zA-Z]" .github/workflows/      → full action-pin inventory (all @vN tags, no SHA pins)
gh api repos/leandrotcawork/MetalDocs/branches/main/protection → 403 (Pro/public required)
ls .github/ (find -maxdepth 2 -type f)           → no CODEOWNERS
grep -rl "go:build integration" --include=*.go . | wc -l   → 359
find . -name "*_test.go" | wc -l                            → 1119
grep -rE "^func Test" tests/ | wc -l                         → 309
wc -l .gitleaks.toml                                          → 32
```
