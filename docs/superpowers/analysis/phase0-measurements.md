# Phase 0 measurements

Inputs the CI restructure (Plan B, Phases 1-5) cannot be written without.
Measured 2026-08-08 on branch ci/a1-verify-single-entry-point at 01715dac.

**Correction (this revision):** the original measurement in this document
(214 scoped / 237 whole-tree) was wrong. golangci-lint v2's defaults are
`max-issues-per-linter: 50` and `max-same-issues: 3`; `.golangci.yml`
overrides neither. Every number below the caps looked like a real count and
every number sitting at exactly 50 looked like a coincidence. It wasn't —
`gocognit` and `revive` were both silently truncated at the 50-per-linter
ceiling, and the true totals are roughly 5x higher. This revision re-runs
both scopes with `--max-issues-per-linter=0 --max-same-issues=0` and
replaces every number that came from a capped run. See "Output caps" below
before trusting any golangci-lint count in this repo again.

## 1. golangci-lint whole-tree

`.github/workflows/golangci-lint.yml:24` sets `only-new-issues: true`, so CI has
never reported the tree's absolute lint state. Measured with the CI-pinned
version at both the CI's scoped path list and the true whole tree, so the
restructure can see what exists outside the workflow's current scope too.

**Version:** golangci-lint v2.11.0 (CI pins `v2.11`; `v2.11.0` is a published
tag, confirmed via `go list -m -versions github.com/golangci/golangci-lint/v2`
under `-mod=mod`).

### Output caps: golangci-lint v2 defaults truncate, and this repo doesn't override them

golangci-lint v2's built-in defaults are:

- `output.show-stats` aside, the **issue-selection** defaults are
  `max-issues-per-linter: 50` and `max-same-issues: 3` (see `golangci-lint
  run --help` / the v2 docs for `--max-issues-per-linter` and
  `--max-same-issues`).
- `.golangci.yml` in this repo sets neither key. Both caps are active on
  every CI run and on the original Task 7 measurement.

A linter that would report more than 50 issues stops at 50 and prints no
indication that it was cut off — no `truncated`, no `+N more`, nothing. A
linter whose issues cluster into more than 3 near-duplicates (`max-same-issues`,
which golangci-lint applies per rule-text/pattern) drops the extras the same
way. The original measurement's `revive: 50` and `gocognit: 50` were the
`max-issues-per-linter` ceiling, not the true count, and `errcheck: 33`
(scoped) / `37` (whole-tree) were `max-same-issues`-shaped undercounts, not a
true measurement that happened to look plausible. **Every count in this
document is now measured with both caps disabled** (`--max-issues-per-linter=0
--max-same-issues=0`); do not remove those flags in a future re-run of this
comparison without re-deriving the caveat above.

### As-configured: CI's scoped paths

**Command:**
```
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.0 run \
  --timeout=10m --max-issues-per-linter=0 --max-same-issues=0 \
  --output.text.path=stdout \
  ./apps/api/... ./internal/... ./tools/...
```
(Matches `golangci-lint.yml:28` path list, plus the two cap-disabling flags
and `--timeout` — see Timeout note below.)

**Total findings:** **1078** (previously reported: 214 — a 5.0x undercount)
**Wall-clock:** 7.0s (warm module cache and warm Go build cache — see Timeout
note). Exit code 1 (issues found, not a tool error).

**Per linter:**

| Linter | Count (true) | Previously reported (capped) |
|---|---:|---:|
| revive | **582** | 50 |
| errcheck | **267** | 33 |
| gocognit | **67** | 50 |
| contextcheck | **51** | 12 |
| gosec | **40** | 22 |
| errorlint | **31** | 8 |
| gocyclo | **16** | 16 |
| exhaustive | **7** | 7 |
| staticcheck | **6** | 5 |
| gocritic | **6** | 6 |
| nilerr | **3** | 3 |
| bodyclose | **2** | 2 |
| **Total** | **1078** | **214** |

`gocyclo`, `exhaustive`, `gocritic`, `nilerr`, and `bodyclose` did not move —
they were already under both caps, so the previously reported numbers for
those five linters stand. `revive`, `gocognit`, `errcheck`, `contextcheck`,
`gosec`, `errorlint`, and `staticcheck` were all undercounted; `revive` and
`gocognit` were undercounted the most severely because they hit the harder
`max-issues-per-linter: 50` ceiling directly, while `errcheck`/`contextcheck`/
`gosec`/`errorlint`/`staticcheck` lost issues to `max-same-issues: 3` on
repeated same-shaped findings (e.g. many `errcheck` hits are the same
unchecked-return pattern in different call sites, which is exactly what
`max-same-issues` collapses).

`.golangci.yml` enables 15 linters (errcheck, govet, staticcheck, nilerr,
errorlint, gosec, gocritic, revive, gocyclo, gocognit, sqlclosecheck,
rowserrcheck, bodyclose, exhaustive, contextcheck). Three reported zero
findings tree-wide: **govet, sqlclosecheck, rowserrcheck** — clean, not
unrun (see "no failed linters" below). This did not change under uncapping;
zero is not truncatable.

**Per package/directory** (uncapped scoped run; recomputed from scratch —
the previous package table was built from capped per-linter output and its
rankings do not survive uncapping):

| Package | Count |
|---|---:|
| internal/modules/documents | 291 |
| internal/modules/approval | 117 |
| internal/modules/iam | 94 |
| internal/modules/render | 76 |
| internal/modules/templates | 53 |
| internal/modules/taxonomy | 43 |
| internal/platform/messaging | 29 |
| internal/modules/security | 27 |
| internal/modules/search | 25 |
| internal/modules/auth | 24 |
| internal/modules/controlleddocuments | 24 |
| internal/modules/audit | 21 |
| internal/modules/jobs | 20 |
| (28 more packages, 1-19 findings each) | 234 |

**The ranking inverted.** The capped run's top package was `approval` (37 of
214, 17%); uncapped, `approval` is second and `documents` — invisible near
the top of the old table — is first with 291 of 1078 (27%), almost
certainly because `documents` carries a large share of the tree's
`errcheck`/`revive` volume that the caps were hiding. Top 3 (documents,
approval, iam) = 502/1078 (47%); top 7 = 703/1078 (65%). The concentration
*shape* (a handful of modules carry most of the backlog) still holds, but
the previous document named the wrong modules as the concentration and
would have sent remediation at the wrong targets first.

### Whole tree: `./...`

The scoped run only covers three subtrees. `apps/worker`, `apps/jobs`,
`cmd/`, `scripts/`, and `tests/` sit outside it — the restructure needs to
know what is out there before deciding the workflow's scope is final.

**Command:**
```
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.0 run \
  --timeout=10m --max-issues-per-linter=0 --max-same-issues=0 \
  --output.text.path=stdout ./...
```

**Total findings:** **1173** (previously reported: 237 — a 4.9x undercount)
**Wall-clock:** 5.5s (warm caches, same session). Exit code 1.

**Per linter:**

| Linter | Count (true) | Previously reported (capped) |
|---|---:|---:|
| revive | **584** | 50 |
| errcheck | **283** | 37 |
| gocognit | **93** | 50 |
| gosec | **67** | 26 |
| contextcheck | **55** | 12 |
| errorlint | **36** | 10 |
| gocyclo | **19** | 19 |
| exhaustive | **12** | 10 |
| staticcheck | **9** | 8 |
| gocritic | **9** | 9 |
| nilerr | **4** | 4 |
| bodyclose | **2** | 2 |
| **Total** | **1173** | **237** |

**Delta vs scoped (1173 - 1078 = 95), entirely outside `apps/api/`, `internal/`,
`tools/`:**

| Location | Count |
|---|---:|
| scripts/api-lint/ | 43 |
| scripts/req-trace/ | 9 |
| scripts/release-backfill/ | 5 |
| tests/unit/iam_memberships/ | 15 |
| tests/unit/iam_people/ | 3 |
| tests/unit/auth_login_policy_test.go | 2 |
| cmd/gen-http-surface/ | 5 |
| cmd/problem-codes-dump/ | 4 |
| cmd/gen-tripwire/ | 1 |
| apps/worker/cmd/metaldocs-worker/ | 5 |
| apps/jobs/cmd/metaldocs-jobs/ | 3 |

Delta by linter: errcheck +16, gosec +27, gocognit +26, errorlint +5,
exhaustive +5, contextcheck +4, gocyclo +3, gocritic +3, staticcheck +3,
revive +2, nilerr +1, bodyclose +0 (=95). This delta table also moved under
uncapping — the previous version (built from capped counts) understated
`scripts/api-lint` in particular, which was itself hitting `max-same-issues`
in the scoped run's shared linters.

### Headline finding: two of the four shipped production binaries have zero golangci-lint coverage in CI

`.github/workflows/golangci-lint.yml:27` lints exactly
`./apps/api/... ./internal/... ./tools/...`. `apps/worker` (5 findings
above: contextcheck, gocognit, gocritic, gosec, nilerr — one each) and
`apps/jobs` (3 findings: contextcheck, gocritic, gosec — one each) are not
in that path list and never have been. **`metaldocs-worker` and
`metaldocs-jobs` — two of MetalDocs' four production binaries — ship with
zero golangci-lint coverage in CI today**, independent of any question
about lint-count severity or output caps. This is a scope gap in the
workflow's path list, not a finding-count problem, and it should be a named
line item for the CI restructure regardless of what Phase 5 decides about
blocking gates. (The finding counts in those two binaries are small right
now, which is exactly why closing the gap is cheap — the size of the gap
only grows the longer it's uncovered.)

### No failed linters, no timeouts

Both runs exited cleanly (`exit status 1`, meaning "issues found," not a
tool crash) and produced golangci-lint's own summary block, which lists
every linter that ran. No linter is missing from the summary, no `panic:`,
no `level=error` diagnostic, no `deadline exceeded`. A build failure
anywhere in `./...` would have made every type-aware linter (staticcheck,
gocritic, gosec, etc.) refuse to run for that package and there is no sign
of that here — both scopes compile clean under `golangci-lint`'s own build.

### Timeout note (5m budget)

The workflow pins `--timeout=5m`; this measurement ran with `--timeout=10m`
as a safety margin, plus the two cap-disabling flags. That must not be read
as "it fit in 5 minutes because I gave it 10." What actually happened:

- The **reported wall-clock** above (7.0s scoped, 5.5s whole-tree) was
  measured with a **warm** Go module cache and warm build cache (this
  session's earlier runs already populated both).
- No genuinely cold-cache wall-clock number exists for this repo. The CI
  runner's `actions/setup-go` cache behavior differs from this box's, so a
  true cold-CI number cannot be extrapolated precisely from this
  measurement.
- What can be said plainly: a warm run with caps disabled finishes in
  single-digit seconds — disabling the caps did not meaningfully change the
  runtime, because the caps only affect how many issues golangci-lint
  *reports*, not how much analysis it runs (every linter still walks the
  whole scope; it just stops appending to the output list past the cap).
  Even allowing generously for cold module download and a cold build cache
  (which the CI runner mitigates via `cache: true` on `setup-go`), there is
  no evidence in this measurement that a whole-tree or scoped run would
  approach the 5m budget. But this is inference from a warm number, not a
  directly observed cold one — flagged rather than asserted as fact.

## Overlap with the standalone `staticcheck` registry check

`tools/verify/registry.go:109` runs staticcheck standalone; `.golangci.yml`
also enables `staticcheck` among its 15 linters. Scopes differ: standalone
`./...` vs golangci's three subtrees (or `./...` in the whole-tree run above).

**Command (verbatim from registry.go:116):**
```
go run honnef.co/go/tools/cmd/staticcheck@2025.1.1 ./...
```

**Standalone findings:** 0. Exit code 0. Wall-clock 5s (warm cache).

**golangci's staticcheck slice (uncapped):** 6 (scoped run) / 9 (whole-tree
run) — up from the previously reported 5/8. The staticcheck slice barely
moved under uncapping (it was never near either cap), so the root-cause
analysis below is confirmed, not revised. Breakdown:

- Scoped (6): `QF1001` x4 (De Morgan's law), `ST1005` x2 (capitalized error
  strings).
- Whole-tree (9): `QF1001` x4, `QF1012` x3 (new — only visible outside the
  scoped subtrees), `ST1005` x2.

**Overlap: none observed.** The repo carries `staticcheck.conf` at the root:

```
checks = ["all", "-ST1000", "-ST1003", "-ST1016", "-ST1020", "-ST1021", "-ST1022", "-ST1005"]
```

That config explicitly disables `ST1005` (with a comment explaining that
oapi-codegen's generated code capitalizes error strings and churning it for
a style check isn't worth it) — so the standalone check, honoring this repo
config file, would never flag the two `ST1005` findings that golangci-lint's
staticcheck linter reports. `.golangci.yml` has no `linters.settings.staticcheck`
block, so golangci's staticcheck integration runs with its own defaults
rather than reading `staticcheck.conf`, and it also surfaces `QF1001`/`QF1012`
(quickfix-category checks) that the standalone run with this repo's config
does not report at all. **Still open, not resolved by this correction pass:**
whether the `QF1001`/`QF1012` gap is a config difference or a
golangci-bundled-vs-standalone-binary version difference between
golangci-lint v2.11.0's bundled staticcheck analyzer and the standalone
`2025.1.1` binary. Either way, the two checks are running under different
effective configurations today, not just different path scopes.

### Bearing on spec §4.6

**The two stacks are not redundant as configured, and folding one into the
other is not a same-day change.** The standalone staticcheck run
(config-aware, `./...`) found zero issues — it is already green and has
presumably been kept green by the existing registry gate. golangci-lint's
staticcheck slice (6-9 issues, uncapped) only exists because golangci's
integration silently ignores `staticcheck.conf` and runs a different
effective check set. Deleting the standalone check and relying solely on
golangci's staticcheck linter would either (a) silently start enforcing
`ST1005` against generated oapi-codegen code — the exact churn
`staticcheck.conf` was written to avoid — or (b) require reproducing
`staticcheck.conf`'s exclusion list inside `.golangci.yml`'s
`linters.settings.staticcheck.checks`, which is a real but small
config-migration task, not a decision that falls out of the numbers alone.

Recommendation: **keep both for now, but record this as a named gap**,
not a footnote — `.golangci.yml`'s staticcheck linter and the standalone
`tools/verify` staticcheck check are running two different effective
configurations against overlapping code today, and nobody has verified
which one the team actually wants. Phase 5 should either (a) explicitly
port `staticcheck.conf`'s exclusions into `.golangci.yml` and retire the
standalone check, or (b) explicitly configure golangci's staticcheck
linter to match the standalone one's exclusions so the two stop disagreeing
silently. Either is fine; "leave it as an unexamined dual stack" is not.

### Bearing on "can `lint-go` become a blocking gate" (revised)

The original version of this section concluded the backlog was tractable in
"a few focused remediation passes," reasoning from 214 findings of which
47% was "two mechanical linters." **That reasoning does not survive
uncapping and the conclusion is withdrawn, not softened.** At the true
count:

- **1078 findings on CI's own scope (1173 whole-tree) is not "a real but
  tractable backlog."** It is dominated by two linters that together are
  849 of 1078 (79%): `revive` (582, 54%) and `errcheck` (267, 25%). Neither
  is uniformly mechanical at this volume — `revive`'s 582 findings span
  whatever rule set `.golangci.yml` enables for it (not just naming/doc
  comments), and 267 `errcheck` findings each require a real judgment call
  (is this unchecked error safe to drop, or a live bug) rather than a
  scripted fix.
- **Concentration exists but does not make the count small.** Documents,
  approval, and iam carry 47% of the scoped backlog (502/1078), which means
  a module-by-module remediation order is possible, but the modules
  themselves are large (documents alone: 291 findings) — "concentrated" is
  not a synonym for "quick."
- **Zero correctness-analyzer findings still holds** (`govet`,
  `sqlclosecheck`, `rowserrcheck` are all clean at both scopes, uncapped),
  which is genuinely good news: none of this backlog is a `govet`-class
  bug. But it is a statement about severity mix, not about size — 1078
  style/complexity/error-handling findings still block a green gate.
- **Runtime is still cheap** (uncapped runs finish in single-digit seconds
  at both scopes), so re-running the check during a remediation effort
  costs nothing. That was true before and remains true; it just does not
  make the backlog itself smaller.

**Revised read: golangci-lint cannot become a blocking whole-tree (or even
CI-scoped) gate in this phase.** Flipping `only-new-issues: false` today
would fail essentially every PR against a 1078-1173-finding backlog
dominated by two linters that both require non-mechanical judgment at scale
(revive's rule breadth, errcheck's case-by-case triage). A blocking gate is
only plausible after a real, likely multi-pass remediation program — most
plausibly linter-by-linter (clear `revive` and `errcheck` first, since they
are 79% of the scoped total) or module-by-module (documents, approval, iam
first, since they are 47%) — not as a same-phase flip. Phase 5 should treat
"make `lint-go` blocking" as its own scoped follow-on work with its own
remediation budget, not a checkbox this Phase 0 measurement clears.

## Recommendation: stop letting golangci-lint hide its own backlog

`.golangci.yml` should set `max-issues-per-linter: 0` and `max-same-issues: 0`
so that every future run — CI or local — reports the true count instead of
silently truncating at 50/3. This is a recommendation for a later phase to
act on, not a change made here: **Task 7 measures the tree's lint state; it
does not modify `.golangci.yml`, and this document does not touch it.**
Whoever owns the Phase 5 (or earlier) config change should also decide
whether `only-new-issues: true` stays on for PR-diff ergonomics once the
backlog is being worked down — that is a separate, related decision from
disabling the output caps, and this document takes no position on it beyond
flagging that the caps themselves should go regardless of what
`only-new-issues` does.

## 2. Why `full` is red

Measured 2026-08-08 after Tasks 1–5 (`test-discipline`, `db-dictionary`, and
`cilint` already green, so they do not appear below as findings — but they
are still in the table for completeness).

**Command:** `go run ./tools/verify --profile=full`
**Wall clock:** the aggregate command itself did not complete — see
"Methodology" below. Summed solo runtimes for the 29 checks: **≈471s**
serial (excludes `go-test-integration`, which could not execute its
registry-defined command at all — see its row). CI's own historical `full`
wall clock is **1176s** (~19.6 min, cited in this plan's spec); that number
is credible given `go-test-integration`'s no-race diagnostic alone took
439s (7m19s) locally, and `-race` plus CI's own resource profile would push
that further.
**Result:** exit 1 (real failures + a check that could not execute its
registry-defined command at all).

### Methodology: the aggregate run itself is not trustworthy on this box

`go run ./tools/verify --profile=full` was started in the background and
produced 16 of 29 check lines, then stopped producing output entirely — no
further checks completed, no `EXITCODE` line was ever written, and no `go`,
`gcc`, `node`, or `pnpm` process was left running under it. The process had
died without printing a final report. Per this task's own escalation rule,
the aggregate run was abandoned after confirming (via `TaskStop`, which
reported "no task found" for both the run and its poller) that the process
was gone, and the remaining 13 checks — plus the 6 that had already reported
FAIL in the aggregate's partial output — were re-run individually instead.

That per-check rerun surfaced a second, independent finding: **5 of the 6
checks the aggregate run reported as FAIL are clean when run alone**
(`gofmt`, `adr-status`, `test-discipline`, `css-token-discipline`,
`eigenpal-selector-pin` — all exit 0 standalone, with real output proving a
real pass, e.g. `test-discipline: clean (176 integration test files
checked)`, `OK: 1148 Go files are gofmt-clean`). All five reported FAIL in
the aggregate run in 0.1–0.3s, far faster than their real standalone runtime
(0.3s–56s) — too fast to have actually executed their check logic. The
registry runs checks at `parallelism=6` by default
(`tools/verify/main.go`'s `run()`), and something about 6 concurrent `bash`/
`git`/subprocess invocations on this Windows/git-bash box produces spurious
near-instant failures for several of them. This was not investigated
further (out of scope: this task inventories, it does not fix `tools/
verify`), but it means **`--profile=full`'s own aggregate/parallel mode
cannot be trusted at face value on this machine** — every FAIL it reports
must be confirmed by a solo rerun before being counted as real. Only
`req-trace` (0.5s aggregate, 0.37s solo, exit 1 both times, identical
output) was consistent between the two run modes. This is a distinct
finding from the SKIP/false-green problem the task brief names, but it is
the same shape: an aggregate "FAIL" that never actually observed the code
under test is a false negative the way a "SKIP reported as pass" is a false
positive. CI runs on Linux, not this Windows box, so this specific
parallelism artifact may not reproduce there — but it is a reason to treat
any *local* `--profile=full` red as unverified until each failing line is
re-run alone.

### Real failures — code is wrong

| Check | Failure | Owner module |
|---|---|---|
| `go-test-unit` | `TestHGCrossModule_Negative_SubpackageSameModule` (tools/cilint/internal/analyzers) fails: the test asserts that an intra-module `documents/approval ⊂ documents` read must NOT be flagged, but the analyzer now flags it (1 finding). This is very likely a fallout of ADR 0082 (approval promoted from a nested exception under `documents` to its own top-level module, landed on this branch): the negative-test fixture still encodes the pre-ADR-0082 subpackage relationship, so it is asserting a containment that is no longer architecturally true. Whole run also failed with this as the only Go-side failure. | tools/cilint (test fixture), approval (ADR 0082 fallout) |
| `go-test-integration` | See its own row below — could not execute its registry-defined command at all (missing cgo), so no real-failure verdict is available from the exact check. The diagnostic no-race substitute (not the registry command — see that row) surfaced 14 failing packages; of those, `internal/platform/idempotency`'s `TestMiddleware_Conflict_Returns422` / `TestMiddleware_SameKeyDifferentResourcePath_Returns422` (idempotency conflict returns HTTP 409, test wants 422) and `apps/api/cmd/metaldocs-api`'s `TestNoDeclaredOperationIsUnreachable` (a self-reporting test that intentionally fails while `metaldocs.role_capabilities` seeds `area_admin`/`qms_admin`/`signer` on capabilities no assignable role can hold — its own message calls this "pre-existing") read as real code-level findings, not environment noise. The remaining ~11 diagnostic failures cluster at suspiciously uniform ~60s durations (60.05s–61.49s) across otherwise-unrelated packages (`tests/integration/{audit,controlleddocuments,documents,iam,scenarios,templates,tenantdata}`, `internal/modules/{approval/infrastructure,iam/authz,templates/infrastructure}`) — consistent with resource/connection-pool contention from running many DB-heavy packages at Go's default test parallelism against one shared local Postgres container, not 11 independent code defects. **Not confirmed either way** — isolating each would require re-running it alone against the DB, which this task's time budget did not cover. Flagged, not asserted. | approval, iam, templates, tenantdata, platform/idempotency, platform/bootstrap (mixed; see next section for the bootstrap ones, which are environment not code) |

### Infrastructure skips — the false-green class

`tools/verify/main.go:341`'s `missingInfra` returns SKIP when a `Needs:`
precondition is unmet, and `report` exits 0 on a skip. **None fired in this
measurement** — `METALDOCS_DATABASE_URL` was set and the container
(`metaldocs-postgres`) was healthy and accepting connections before every
run, this repo is not a shallow clone (`needsGitDepth` satisfied), `docker`
is on PATH (`needsDocker` satisfied, though no check in `full` currently
declares it), and network access was available for `staticcheck`'s
`needsNetwork`. So the zero-skip result here is not "the control was never
exercised" — it is "the control's trigger condition (missing infra) did not
occur this run," which is exactly why this section is empty and Step 4
below is the sharper finding instead.

| Check | Unmet `Needs:` | What it would have measured |
|---|---|---|
| *(none — all `Needs:` preconditions were satisfied this run)* | — | — |

**This is not the same as "the false-green risk is closed."** It only means
this particular run had the infrastructure present. The mechanism
(`missingInfra` → SKIP → exit 0) is unconditional: any future run — a CI
runner without `METALDOCS_DATABASE_URL` wired, a laptop with Postgres
stopped, a shallow-clone CI checkout for `req-trace` — reproduces exactly
the false-green the task brief describes, with zero code change. See
"Bearing on Plan B" below for the sharper version of this risk that this
run *did* surface: a check whose `Needs:` are satisfied can still fail to
exercise the suite, for a reason the registry does not declare at all.

### Environment failures — local only

| Check | Cause |
|---|---|
| `go-test-integration` | `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` — exit 2, 0.14s, before compiling a single package. This machine has `CGO_ENABLED=0` (Go's own default when no C toolchain is detected) and no `gcc`/`cc` binary anywhere on `PATH` (`where gcc.exe` / `where cc.exe` both report nothing found; `/mingw64/bin` carries only `libgcc_s_seh-1.dll`, a runtime DLL, not a compiler). **This is not the `missingInfra`/SKIP path** — the registry's only declared `Needs:` for this check is `postgres`, which was satisfied, so `missingInfra` returned `""` and the check was attempted for real. It failed anyway, immediately, on a dependency the registry does not know to check for. This is the sharpest version of the task brief's "the required job can be green over zero tests" concern turned inside out: here it is *neither* green *nor* a declared skip — it is a silent-if-you're-not-reading-closely exit 2 that a CI log would show as "FAIL go-test-integration" with a one-line cause easy to skim past as "some flaky test," when the real story is "the suite never got past `go build`." A CI Linux runner with `build-essential`/`gcc` installed would not hit this; this box does not have that installed. Recorded as an environment gap, not a code defect — but also as evidence that `Needs:` in the registry is incomplete for this check (cgo/gcc is a real, undeclared precondition). |
| `gofmt`, `adr-status`, `test-discipline`, `css-token-discipline`, `eigenpal-selector-pin` | Reported FAIL only inside the aggregate `--profile=full` run, in 0.1–0.3s each — far below their real standalone runtime (0.3s–56s). Solo reruns of the identical registry command (same argv, same working directory) all exit 0 with real, substantive pass output. Classified as an artifact of this Windows/git-bash box's behavior under the registry's default `parallelism=6`, not a code defect — see "Methodology" above. |

### Bearing on Plan B

**Can `test-integration` (`go-test-integration`) be a blocking job? Not yet,
and not for the reason the plan expected.** The concern going in was the
`missingInfra`/SKIP false-green path — that a required integration job could
report green over zero tests because `METALDOCS_DATABASE_URL` was unset.
This run had that variable set and Postgres healthy, so that specific path
did not fire. But the check still could not produce a real result: it died
on a missing C toolchain, a dependency the registry's `Needs:` list does not
mention at all. That means **the skip mechanism is not the only way this
check can go dark** — an operator who wires up Postgres and calls it done
has fixed exactly one of at least two ways `go-test-integration` can fail to
exercise the suite. Before this can block a PR:

1. **The registry's `Needs:` for `go-test-integration` must be corrected**
   to also declare (or the CI job's environment must guarantee) a working
   cgo toolchain, or the check must drop `-race` and accept weaker coverage,
   or `missingInfra` must grow a new precondition class for it. Silently
   discovering this a second time, in CI, the way it was discovered here, is
   the same failure mode the task brief opened with — just for a different
   missing dependency.
2. **The one real, confirmed code-level failure this measurement surfaced**
   (`go-test-unit`'s `TestHGCrossModule_Negative_SubpackageSameModule`) must
   be fixed or the fixture updated for ADR 0082's approval-module promotion
   before Phase 2 adds `ci.yml`, or every PR touching anything under
   `internal/` inherits a red `go-test-unit` from day one.
3. **The idempotency status-code mismatch and the pre-existing
   unreachable-capability finding** (both surfaced only via the no-race
   diagnostic substitute, not the registry's own `-race` command) need a
   disposition — fix or explicit acceptance — before they can sit behind a
   blocking gate; right now they are undispositioned reds with no `req-
   trace`-style writeup.
4. **The ~11 uniform-duration diagnostic failures need isolation, not
   dismissal.** They read like test-parallelism/DB-connection contention,
   not 11 independent bugs, but that is a hypothesis, not a confirmed
   classification — Phase 2 should not blindly retry-until-green a job that
   might be surfacing a real, intermittent concurrency defect.

**If the skip list is non-empty, that is the direct justification for
`--require-infra`:** in this run, the skip list is empty because the
infrastructure happened to be present — but the mechanism that would have
produced a false green (missing `METALDOCS_DATABASE_URL`) is still live in
`tools/verify/main.go:341` for every check tagged `Needs: postgres`
(currently `go-test-integration` alone) and every check tagged `Needs:
git-history` (`req-trace`) or `Needs: network` (`staticcheck`). A
`--require-infra` flag that turns "SKIP" into "FAIL: infra required and
absent" would convert exactly **1 check** (`go-test-integration`) from a
silent pass-over-nothing into an honest failure in the one scenario this
task exists to prevent (CI without `METALDOCS_DATABASE_URL` wired), plus
harden `req-trace` and `staticcheck` against the same failure mode under a
shallow clone or offline runner respectively — 3 checks total that currently
have a "green without running" escape hatch.

**Net verdict:** `full` cannot move behind the `required` gate as-is.
Staging it there today would stage one confirmed real failure
(`go-test-unit`), one check that cannot currently produce a verdict at all
on an undeclared missing dependency (`go-test-integration`), and an
unknown number (0–11+) of unconfirmed real failures under the same red —
exactly the "staging a red job just moves the redness later" problem the
plan opened with. Phase 2 needs items 1–4 above closed (or explicitly
deferred with a written disposition, `req-trace`-style) first.
