# Lane: testing

## Findings

| ID | Class | Finding | Evidence | Scale |
|----|-------|---------|----------|-------|
| TEST-01 | gap | PR gate (`test-smoke.yml`) runs only 5 name-filtered integration test functions out of 553 repo-wide; the other 548 (43,691 LOC across 342 `//go:build integration` files) run only on `push` to `main` (`test-full.yml`), i.e. post-merge | `.github/workflows/test-smoke.yml:47` (`-run "TestTriggerBypass\|TestMembership\|TestSchemaLockdown\|TestLegacy\|TestE2E"` over `tests/integration/scenarios/` only); `.github/workflows/test-full.yml:3-4,29-33` (`on: push: branches:[main]`, `go test -tags integration ./tests/... ./internal/... ./apps/...`) | 548/553 funcs (99%), 342 files |
| TEST-02 | gap | 5 of 13 `check-*.sh`/`check-*.ps1` CI guard scripts have zero test files proving they fail on a bad input — only `scripts/api-lint` (Go, 12 `_test.go` + 20 negative `testdata/*.yaml`) and `tools/cilint` (Go, 9 `_test.go`, Positive+Negative pairs) are self-tested | `find` over `scripts/*.sh`/`*.ps1` matching `check-` returned 13 files; none has a sibling `_test.go`/`_test.ps1` or `bats`/Pester harness (verified via targeted `find`) | 13 shell/PS guard scripts, 0 self-tests |
| TEST-03 | drift | `wiki/quality/test-discipline.md`'s allowlist table (5 files, R3+R4 only) is stale against `scripts/check-test-discipline.sh`, which also carries an undocumented `R2_ALLOWLIST` (2 more files: `templates/infrastructure/tenant_id_rls_integration_test.go`, `tests/integration/iam/migration_0170_test.go`) added in a later commit (`c1b37817`) | `wiki/quality/test-discipline.md:106-119` vs `scripts/check-test-discipline.sh:57-68` | 2 undocumented allowlist entries, 7 unique files total vs 5 documented |
| TEST-04 | gap | `distribution` module (6 prod files, 1,555 LOC) has exactly 1 test file, and it is integration-only (`coverage_repository_integration_test.go`); `domain/types.go`, `delivery/http/handler.go`, `delivery/http/routes.go` have zero unit coverage | `internal/modules/distribution/**` file listing | 1 test file / 6 prod files, ratio 0.32 |
| TEST-05 | gap | Backend e2e directory is a placeholder: `tests/e2e/` contains only `.gitkeep`, no Go e2e tests exist despite the dir's presence in the repo tree | `tests/e2e/.gitkeep` (sole entry) | 0 files |
| TEST-06 | idiom | 30 `t.Skip(` call sites across the Go test suite; not triaged in this pass whether these are environment-gated (e.g. missing `INTEGRESQL_URL`) or genuinely disabled tests | `grep -rn "t.Skip(" --include="*_test.go" internal apps tests` | 30 sites |
| TEST-07 | gap | `iam` module — the largest single module by production LOC (15,704) and origin of tier-1/tier-2 authz plumbing — has a test-LOC ratio of 0.44, second-lowest among modules with real coverage (only `distribution` and `tokens` are lower) | module LOC sweep (see Census below) | ratio 0.44, 15,704 prod LOC |

## The five heaviest, with detail

**TEST-01 — the PR gate does not run the integration suite.** `test-smoke.yml` filters to 5 of 553 integration test functions via a hardcoded `-run` regex over a single directory; the full 43,691-LOC/342-file integration suite (R1-R4 discipline rules, tripwire firing, RLS isolation, outbox behavior) only executes on `push` to `main` in `test-full.yml`. An author can merge a PR having seen zero integration-test evidence beyond `go build ./...` and `go test ./...` (no tags — compiles nothing under `//go:build integration`). This is the highest-leverage finding in the lane: it means ADR 0034's entire framework, and everything layered on it (R1-R4, tripwire proofs, RLS proofs), is a post-merge safety net, not a pre-merge gate. Any regression surfaces after the merge, not before.

**TEST-02 — most of the guard layer is unproven.** `scripts/api-lint` and `tools/cilint` (Go, compiled) both carry rigorous positive/negative test suites — e.g. `tools/cilint/internal/analyzers/hgcrossmodule_test.go` has explicit `_Positive_`/`_Negative_` pairs, and `scripts/api-lint/testdata/*_bad.openapi.yaml` fixtures exist for casing, envelope, pagination rules. But the 13 shell/PowerShell `check-*.sh`/`.ps1` scripts — including `check-module-boundaries.ps1` (the boundary-cycle checker the shared brief already flags as blind to cycles), `check-test-discipline.sh` itself, `check-adr-status.sh`, `check-governance.ps1` — have no test proving any of them actually fails on a crafted bad input. A guard that has never been shown to fire is a guard whose failure mode is unknown.

**TEST-03 — the allowlist doc is drifted from the enforcement script.** ADR 0034 states the allowlist "can only shrink" and requires operator approval to grow; `wiki/quality/test-discipline.md` documents 5 files. The live script has grown to include an `R2_ALLOWLIST` (2 more files, added in commit `c1b37817`) that the wiki page never mentions. Whether that growth had operator approval is unverified from the file alone — the governance mechanism (wiki table as the audit trail) has silently stopped tracking what the script enforces.

**TEST-04/TEST-05 — thin or absent coverage on real surfaces.** `distribution` ships an HTTP handler and domain types with no test touching them directly (only an integration test against its repository). The backend e2e directory is vestigial — whatever e2e assurance exists for the backend is not exercised as Go e2e tests; frontend Playwright specs (`frontend/apps/web/e2e/`, `frontend/apps/web/tests/e2e/`, 19+12 spec files) are the only e2e layer that actually runs.

**TEST-06 — 30 `t.Skip` sites, untriaged.** Not disqualifying on its own (many are likely `INTEGRESQL_URL` absence guards per ADR 0034's stated design), but the count was not broken down by cause in this pass; a remediation program should distinguish "skips because infra is absent" from "skips because the test is broken and was silenced."

## What is actually fine

- **`tools/cilint` analyzer test discipline** — every analyzer (`hgcrossmodule`, `nodualmode`, `noresponsemap`, `nosqltxindomain`, `legacyvocab`, `outboxpair`, `platformboundary`, `postcommitaudit`) has explicit `_Positive_*`/`_Negative_*` test functions, several also covering allow-directive escape hatches and out-of-scope-file guards. This is a genuine proof-on-bad-input guard layer — do not touch.
- **`scripts/api-lint`** — 12 Go test files plus 20 `testdata/*.yaml` fixtures with clearly-named bad cases (`casing_bad.openapi.yaml`, `missing_cursor.openapi.yaml`, `missing_security.openapi.yaml`, `envelope_unresolved_ref.openapi.yaml`). Same standard as cilint; fine as-is.
- **Tier-2 authz core (`internal/modules/iam/authz/authz.go` `Require`)** — 11 test functions in `authz_test.go` covering grant/deny, cache scoping per (tx, actor, tenant), system-admin bypass with audit-emit and fail-closed-without-audit, and a `TestDenyByDefault_EveryCapabilityDeniesUnprivilegedActor` sweep. This is well-exercised.
- **DB tripwire actually proven to fire** — `internal/modules/iam/integration_test.go:123-124` and `:322-323` assert `P0001` (`ErrCapabilityNotAsserted`) and a route-immutability trigger fire on bad input against a live DB, not just that generated SQL matches a registry.
- **`internal/platform/worker`** — both `materialize_job_runner` and `pdf_job_runner` have unit + integration test pairs; outbox consumer logic is not a blind spot.
- **`invariants.yml`** runs `tools/cilint` and `go vet` on both `pull_request` and `push` — the compiled-Go guard layer, unlike the integration test suite, genuinely is pre-merge.

## Unverified / needs judgment

- Was the `R2_ALLOWLIST` growth (TEST-03) operator-approved per ADR 0034's stated rule? Not determinable from the file; would need the PR/commit `c1b37817`'s review trail.
- Of the 30 `t.Skip` sites, how many are `INTEGRESQL_URL`-absence guards (by design) vs disabled/broken tests silenced without a tracking issue?
- Is `go test ./...` (untagged, in `test-smoke.yml:18`) actually exercising meaningful assertions for the 15 modules, or mostly compiling packages whose real logic lives behind the integration tag? Not quantified — would need a per-package tag-vs-untagged assertion count.
- `time.Now()` appears in 105 test files — not broken down between legitimate use (fixture timestamps) and literal `time.Now()`-dependent assertions that could flake near midnight/DST boundaries; needs a manual pass, not a grep count.

## Commands run

```
grep -rL "go:build integration" --include="*_test.go" internal apps | grep -v "/tests/integration/"
grep -rl "^//go:build integration" --include="*_test.go" .
find internal apps tests -name "*_test.go" | xargs cat | wc -l
find internal apps -name "*.go" ! -name "*_test.go" | xargs cat | wc -l
find internal apps tests -name "*_test.go" | xargs grep -l "^//go:build integration"  # -> 180 files, 43691 LOC
find frontend -name "*.test.ts*" -o -name "*.spec.ts*" | grep -v node_modules | wc -l
for m in <15 modules>; do find internal/modules/$m -name "*.go" ! -name "*_test.go" | xargs cat | wc -l; ... done
git log --oneline --follow -- scripts/check-test-discipline.sh
git log -p --follow -- scripts/check-test-discipline.sh | grep -E "^\+.*ALLOWLIST=\(|^\+  \"|^-  \""
find tools/cilint -name "*_test.go"; grep -n "func Test" tools/cilint/internal/analyzers/*_test.go
find scripts/api-lint -type f
grep -rli "tripwire" --include="*_test.go" internal | wc -l
grep -rln "P0001" --include="*_test.go" internal
grep -n "^on:" .github/workflows/{ci,test-full,test-smoke,test-nightly,invariants}.yml
grep -rE "^func Test.*(TriggerBypass|Membership|SchemaLockdown|Legacy|E2E)" tests/integration/scenarios/*.go | wc -l
cat /tmp/int_files.txt | xargs grep -c "^func Test" | awk -F: '{sum+=$2} END{print sum}'
grep -rn "t.Skip(\|time.Sleep(\|NumGoroutine" --include="*_test.go" internal apps tests
```
