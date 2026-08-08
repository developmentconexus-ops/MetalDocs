# Runbook — CI required-gate aggregator and the phase3-hardening-gate job

**Scripts:** `scripts/required-gate.jq`, `scripts/check-required-gate.sh`,
`scripts/phase3-hardening-gate.ps1`
**Owner:** on-call (whoever is triaging a red PR or a red
`phase3-hardening-gate` run)
**Context:** CI restructure phases 1-5 split `required`'s pass/fail logic out
of inline YAML into `scripts/required-gate.jq`, added a bash selftest for it
(`scripts/check-required-gate.sh`), and removed the `contract-baseline` step
from `scripts/phase3-hardening-gate.ps1` because the suite it ran
(`go test ./tests/contract`) was deleted in `dc0572f6` — it asserted against a
`workflow` module the repo no longer has. This runbook covers both changes:
what each script does, and what to do when either goes red.

## scripts/required-gate.jq — the branch-protection aggregator

### What it does

`ci.yml`'s `required` job is the one check GitHub branch protection actually
requires. It depends on `[verify, test-integration, security, lint-go]` and
runs:

```
jq -e -f scripts/required-gate.jq <<<"$NEEDS_JSON"
```

against `toJSON(needs)`. The `.jq` file is the single source of the pass
condition — `required`'s YAML step reads it directly, so there is nothing to
keep in sync by hand. The rule: the key set must be exactly those four job
names, and every one of them must report `"result": "success"`. No
allowance for `"skipped"` — `verify` and `test-integration` are wired so
that a skip is only reachable if something upstream failed, so treating a
skip as green would let a partially-run pipeline pass a merge gate.

### When it goes red

- **`required` fails, and the four upstream jobs' own logs show a real
  failure or skip.** That is the gate working as intended — read the failing
  job's log, not this script. Nothing to do here.
- **`required` fails but every upstream job's log looks fine.** Compare
  `ci.yml`'s `required.needs:` list against the key array in
  `scripts/required-gate.jq`. If a job was renamed or added to one and not
  the other, they'll disagree — `go run ./tools/verify --audit` catches this
  automatically as an **A5** finding (it parses both files and diffs the
  sets), so run that first before comparing by hand.
- **The `.jq` expression itself seems wrong** (e.g. it accepts a result set
  it shouldn't, or rejects one it shouldn't): run
  `scripts/check-required-gate.sh` against the fixtures in
  `scripts/testdata/required-gate/*.json`. Each fixture is named
  `pass-*.json` or `fail-*.json` and the script asserts the `.jq` file's
  verdict matches the filename. Add a new fixture for the case you found
  before changing the expression, so the fix is pinned.

### How to change the required-job set

Editing which jobs gate a merge means editing three things together, in the
same commit:
1. `ci.yml`'s `required.needs:` list.
2. The key array in `scripts/required-gate.jq`.
3. A fixture in `scripts/testdata/required-gate/` covering the new set (at
   minimum, a `pass-*` fixture with the new full key set).

`go run ./tools/verify --audit` (A5) fails closed if 1 and 2 drift, so a
change to only one of them is caught before merge rather than silently
changing what "required" means.

## scripts/phase3-hardening-gate.ps1 — the PR hardening gate

### What it does

Runs on every PR to `main` (`.github/workflows/phase3-hardening-gate.yml`).
Three steps, each gating the next:
1. `go test ./...`
2. `scripts/check-module-boundaries.ps1`
3. `scripts/security-baseline.ps1` (with `-SkipGovulncheck` unless the
   caller overrides it), then reads the newest
   `non_git/security/security_baseline_*.json` and requires
   `status == "approved"`.

Any step's non-zero exit throws, the `catch` block marks the evidence
`status = "rejected"`, and the script re-throws so the job goes red. The
`finally` block always writes the evidence JSON to
`non_git/hardening/phase3_hardening_gate_<timestamp>.json` regardless of
outcome, so a rejected run still leaves a receipt naming which step failed
and (for the security step) which evidence file it read.

A `contract-baseline` step used to run between steps 1 and 2
(`go test ./tests/contract`, asserting the OpenAPI spec covers the runtime
endpoints). It was removed 2026-08-08: the suite imported a `workflow`
module that no longer exists in the 15-module layout, so it could not
compile, let alone assert anything. Its claim is now proved a different way,
at boot rather than in CI: `assertSurface`
(`apps/api/cmd/metaldocs-api/surface.go`) is boot-fatal and records
per-publisher coverage, so a route with no OpenAPI entry fails
`metaldocs-api` startup instead of a CI-only test. See
`docs/superpowers/specs/2026-08-07-ci-restructure-design.md` §11.3 R-1 for
the design rationale.

### When it goes red

- **`go test ./...` step failed.** Read `non_git/hardening/phase3_hardening_gate_*.json`
  for the `steps.go_test.exit_code`; the raw `go test` output is in the
  workflow log above the PowerShell error, same as any other Go test
  failure. Fix the failing test/package; this gate does not narrow by
  package the way `tools/verify` profiles do.
- **`check-module-boundaries` step failed.** `steps.module_boundaries.passed`
  stays `false` in the evidence file. Run
  `./scripts/check-module-boundaries.ps1` locally for the same output the CI
  job saw; it reports which module imported another module's internals.
- **`security-baseline` step failed, or its evidence's `status` isn't
  `"approved"`.** The evidence file's `steps.security_baseline.evidence_file`
  points at the specific `non_git/security/security_baseline_*.json` that was
  read — open that file for the actual finding (govulncheck/gosec output),
  not this wrapper.
- **The job fails with "Nao foi encontrado arquivo de evidencia de security
  baseline."** `security-baseline.ps1` did not write a
  `non_git/security/security_baseline_*.json` file at all — that means it
  exited before reaching its own evidence-write step. Check its log directly;
  this wrapper only looks for the file, it does not run the security tooling
  itself.

### Local reproduction

```powershell
.\scripts\phase3-hardening-gate.ps1
```

Requires a Go toolchain on PATH (the script auto-adds
`C:\Program Files\Go\bin` on Windows if `go` isn't already there) and, unless
`-SkipGovulncheck` is passed, `gosec` and `govulncheck` installed
(`go install ...@latest`, same as the workflow's "Install security tools"
step). Evidence lands under `non_git/hardening/`, gitignored, same as the
workflow's artifact.
