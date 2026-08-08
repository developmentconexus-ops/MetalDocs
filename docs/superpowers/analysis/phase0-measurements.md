# Phase 0 measurements

Inputs the CI restructure (Plan B, Phases 1-5) cannot be written without.
Measured 2026-08-08 on branch ci/a1-verify-single-entry-point at 01715dac.

## 1. golangci-lint whole-tree

`.github/workflows/golangci-lint.yml:24` sets `only-new-issues: true`, so CI has
never reported the tree's absolute lint state. Measured with the CI-pinned
version at both the CI's scoped path list and the true whole tree, so the
restructure can see what exists outside the workflow's current scope too.

**Version:** golangci-lint v2.11.0 (CI pins `v2.11`; `v2.11.0` is a published
tag, confirmed via `go list -m -versions github.com/golangci/golangci-lint/v2`
under `-mod=mod`).

### As-configured: CI's scoped paths

**Command:**
```
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.0 run \
  --timeout=10m --output.text.path=stdout \
  ./apps/api/... ./internal/... ./tools/...
```
(This matches `golangci-lint.yml:28` exactly, `--timeout` aside — see Timeout
note below.)

**Total findings:** 214
**Wall-clock:** 11s (warm module cache and warm Go build cache — see Timeout
note). Exit code 1 (issues found, not a tool error).

**Per linter:**

| Linter | Count |
|---|---|
| gocognit | 50 |
| revive | 50 |
| errcheck | 33 |
| gosec | 22 |
| gocyclo | 16 |
| contextcheck | 12 |
| errorlint | 8 |
| exhaustive | 7 |
| gocritic | 6 |
| staticcheck | 5 |
| nilerr | 3 |
| bodyclose | 2 |
| **Total** | **214** |

`.golangci.yml` enables 15 linters (errcheck, govet, staticcheck, nilerr,
errorlint, gosec, gocritic, revive, gocyclo, gocognit, sqlclosecheck,
rowserrcheck, bodyclose, exhaustive, contextcheck). Three reported zero
findings tree-wide: **govet, sqlclosecheck, rowserrcheck** — clean, not
unrun (see "no failed linters" below).

**Per package/directory** (top `internal/modules/*`, `internal/platform/*`,
`apps/api/*`, `tools/*` by finding count):

| Package | Count |
|---|---|
| internal/modules/approval | 37 |
| internal/modules/iam | 26 |
| internal/modules/documents | 23 |
| internal/modules/render | 13 |
| internal/modules/controlleddocuments | 13 |
| internal/modules/auth | 13 |
| internal/modules/templates | 11 |
| tools/cilint | 8 |
| internal/modules/security | 6 |
| internal/modules/jobs | 6 |
| internal/platform/render | 5 |
| (22 more packages, 1-4 findings each) | 33 |

Findings concentrate: the top 3 packages (approval, iam, documents) account
for 86 of 214 (40%); the top 7 account for 136 (64%). The long tail (22
packages with 1-4 findings) is genuinely thin, not hiding a second
concentration.

### Whole tree: `./...`

The scoped run only covers three subtrees. `apps/worker`, `apps/jobs`,
`cmd/`, `scripts/`, and `tests/` sit outside it — the restructure needs to
know what is out there before deciding the workflow's scope is final.

**Command:**
```
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.0 run \
  --timeout=10m --output.text.path=stdout ./...
```

**Total findings:** 237
**Wall-clock:** 7s (warm caches, same session). Exit code 1.

**Per linter:**

| Linter | Count |
|---|---|
| gocognit | 50 |
| revive | 50 |
| errcheck | 37 |
| gosec | 26 |
| gocyclo | 19 |
| contextcheck | 12 |
| errorlint | 10 |
| exhaustive | 10 |
| gocritic | 9 |
| staticcheck | 8 |
| nilerr | 4 |
| bodyclose | 2 |
| **Total** | **237** |

**Delta vs scoped (237 - 214 = 23), entirely outside `apps/api/`, `internal/`,
`tools/`:**

| Location | Count |
|---|---|
| scripts/api-lint/ | 7 |
| tests/unit/iam_memberships/ | 3 |
| cmd/problem-codes-dump/ | 3 |
| scripts/release-backfill/ | 2 |
| apps/worker/cmd/metaldocs-worker/ | 2 |
| scripts/req-trace/ | 4 |
| cmd/gen-tripwire/ | 1 |
| apps/jobs/cmd/metaldocs-jobs/ | 1 |

None of this is large on its own, but it is real: the workflow's current
scope (`apps/api/`, `internal/`, `tools/`) silently excludes `apps/worker`
and `apps/jobs` — two of the four production binaries — plus `cmd/`,
`scripts/`, and `tests/`. That is a scope gap worth a Plan B line item
independent of the lint-count question.

### No failed linters, no timeouts

Both runs exited cleanly (`exit status 1`, meaning "issues found," not a
tool crash) and produced golangci-lint's own summary block, which lists
every linter that ran. No linter is missing from the summary, no `panic:`,
no `level=error` diagnostic, no `deadline exceeded`. A build failure
anywhere in `./...` would have made every type-aware linter (staticcheck,
gocritic, gosec, etc.) refuse to run for that package and there is no sign
of that here — both scopes compile clean under `golangci-lint`'s own build.

### Timeout note (5m budget)

The workflow pins `--timeout=5m`; I ran with `--timeout=10m` as a
measurement safety margin, per the brief's own command form. That must not
be read as "it fit in 5 minutes because I gave it 10." What actually
happened:

- The **reported wall-clock** above (11s scoped, 7s whole-tree) was measured
  with a **warm** Go module cache and warm build cache: two earlier
  backgrounded attempts in this same session had already downloaded
  `golangci-lint`'s dependency closet (`go: downloading
  github.com/sonatard/noctx`, `github.com/ryanrolds/sqlclosecheck`, etc.)
  before being interrupted, and a stale `/tmp/golangci-lint.lock` from one
  of those orphaned runs had to be cleared before the runs recorded here
  could start.
- I do not have a clean cold-cache wall-clock number. The CI runner's
  `actions/setup-go` cache behavior differs from this box's, so a true
  cold-CI number cannot be extrapolated precisely from this measurement.
- What I can say plainly: a warm run finishes in single-digit seconds,
  nowhere near 5m. Even allowing generously for cold module download and a
  cold build cache (which the CI runner mitigates via `cache: true` on
  `setup-go`), there is no evidence in this measurement that a whole-tree or
  scoped run would approach the 5m budget. But this is inference from a
  warm number, not a directly observed cold one — flagged rather than
  asserted as fact.

## Overlap with the standalone `staticcheck` registry check

`tools/verify/registry.go:109` runs staticcheck standalone; `.golangci.yml`
also enables `staticcheck` among its 15 linters. Scopes differ: standalone
`./...` vs golangci's three subtrees (or `./...` in the whole-tree run above).

**Command (verbatim from registry.go:116):**
```
go run honnef.co/go/tools/cmd/staticcheck@2025.1.1 ./...
```

**Standalone findings:** 0. Exit code 0. Wall-clock 5s (warm cache).

**golangci's staticcheck slice:** 5 (scoped run) / 8 (whole-tree run) — all
`QF1001` (De Morgan's law simplification, 3 occurrences) and `ST1005` (error
strings should not be capitalized, 2 occurrences) in the scoped run; the
whole-tree run adds 3 more `QF1012` (2 more findings overall are elsewhere
in the same files, not new packages — see per-linter table above).

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
does not report at all. I did not isolate whether that QF gap is a config
difference or a version difference between golangci-lint v2.11.0's bundled
staticcheck analyzer and the standalone `2025.1.1` binary; either way, the
two checks are running under different effective configurations today, not
just different path scopes.

### Bearing on spec §4.6

**The two stacks are not redundant as configured, and folding one into the
other is not a same-day change.** The standalone staticcheck run
(config-aware, `./...`) found zero issues — it is already green and has
presumably been kept green by the existing registry gate. golangci-lint's
staticcheck slice (5-8 issues) only exists because golangci's integration
silently ignores `staticcheck.conf` and runs a different effective check
set. Deleting the standalone check and relying solely on golangci's
staticcheck linter would either (a) silently start enforcing `ST1005`
against generated oapi-codegen code — the exact churn `staticcheck.conf`
was written to avoid — or (b) require reproducing `staticcheck.conf`'s
exclusion list inside `.golangci.yml`'s `linters.settings.staticcheck.checks`,
which is a real but small config-migration task, not a decision that falls
out of the numbers alone.

Recommendation: **keep both for now, but record this as a named gap**,
not a footnote — `.golangci.yml`'s staticcheck linter and the standalone
`tools/verify` staticcheck check are running two different effective
configurations against overlapping code today, and nobody has verified
which one the team actually wants. Phase 5 should either (a) explicitly
port `staticcheck.conf`'s exclusions into `.golangci.yml` and retire the
standalone check, or (b) explicitly configure golangci's staticcheck
linter to match the standalone one's exclusions so the two stop disagreeing
silently. Either is fine; "leave it as an unexamined dual stack" is not.

### Bearing on "can `lint-go` become a blocking gate"

At 214 issues on the CI's own scope (237 whole-tree), **a naive
"add `only-new-issues: false` and block on green" is not achievable without
real remediation work first.** But the shape of the backlog matters more
than the raw count for judging how much work that is:

- 100 of 214 (47%) are exactly two linters — `gocognit` (50) and `revive`
  (50) — both style/complexity linters where findings are typically
  mechanical (rename, extract-function, add doc comment) rather than
  logic bugs. These are individually cheap but numerous.
- `errcheck` (33) and `gosec` (22) require case-by-case judgment (is the
  unchecked error safe to ignore, is the gosec finding a true positive) —
  slower to clear per finding, but a fixed, enumerable list.
- Concentration in `approval`, `iam`, and `documents` (40% of scoped
  findings) means the backlog is not evenly smeared across the whole
  codebase — a remediation pass could plausibly go module-by-module and
  show visible progress quickly, rather than needing a single big-bang PR.
- Zero linter failures/timeouts and single-digit-second runtimes both mean
  the check itself is cheap to run repeatedly during remediation — the
  backlog is a content problem, not a tooling problem.

**Read: the backlog is real but tractable, not a hard blocker.** 214-237
findings across 15 linters, concentrated in 3-4 modules, with no
correctness-analyzer (govet, sqlclosecheck, rowserrcheck all already clean)
findings at all, reads as "a few focused remediation passes," not "this
gate can never go green." A blocking `lint-go` gate is plausible after
scoped remediation (starting with gocognit/revive as the cheapest 47%,
then errcheck/gosec), but it is not achievable by just flipping
`only-new-issues` off today — that would break every open PR on day one.
This is the number spec §4.6 and the wider Phase 5 blocking-gate decision
needed; it was previously unmeasured.
