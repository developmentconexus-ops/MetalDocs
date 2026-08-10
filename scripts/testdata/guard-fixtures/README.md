# Guard negative fixtures

Deliberately invalid input. Every file here is wrong on purpose.

A guard that has never been observed to fail is not a control. Unit-testing a
guard's helpers does not close that gap either: CI runs a *command*, and the
property that matters is that the command exits non-zero on bad input — a
helper can be perfectly tested while the command still returns 0.

Each subdirectory is named for a `tools/verify` check ID. `verify
--guard-fixtures` copies the tree into a throwaway git repo and runs that
check's own argv against it, requiring a non-zero exit and (when `Want` is
declared) the specific message under test. The harness lives in
`tools/verify/fixtures.go`; the declarations live next to each check in
`tools/verify/registry.go`.

The positive half of the property is not fixtured here on purpose. Every one
of these checks runs against this repository on every PR, and this repository
is valid input — a synthetic good-case fixture would prove strictly less than
what already runs.

## Conventions

- **Every source file carries a trailing `.txt`**, stripped when copied into
  the sandbox. This is load-bearing. `tools/cilint`'s walker does not skip
  `testdata/`, `scripts/check-gofmt.sh` scans `git ls-files '*.go'`, and
  `scripts/check-test-discipline.sh` reads tracked sources — a fixture named
  `*.go` would be scanned by the very guards it exists to break, turning this
  repository's own CI red on files whose entire purpose is to be wrong.
  `scripts/testdata/test-discipline/` already uses this convention.
- **Flat layout** (files directly under the fixture dir) is the default: one
  commit, one bad tree.
- **Layered layout** — a `base/` and a `head/` subdirectory — exists for
  diff-shaped guards. `base/` is committed first and `refs/remotes/origin/main`
  is pointed at it, then `head/` is copied over and committed, so the guard
  sees a real before and after.
- Fixtures are inert: nothing outside the harness reads them, and no build,
  lint or test in this repository compiles them.
