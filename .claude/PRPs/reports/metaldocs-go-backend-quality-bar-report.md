# Implementation Report: MetalDocs Go Backend Quality Bar

## Summary

Published the MetalDocs Go backend quality bar under `wiki/standards/golang/`, added a golangci-lint v2 config, added a Go Lint workflow, and indexed the new standards from `wiki/README.md`.

## Assessment vs Reality

| Metric | Predicted (Plan) | Actual |
|---|---|---|
| Complexity | Large | Large |
| Confidence | High after evidence read | High |
| Files Changed | 13 new + 1 updated | 13 new + 1 updated |

## Tasks Completed

| # | Task | Status | Notes |
|---|---|---|---|
| 1 | Read evidence files | done | Review tracker, #2a, #1, code patterns, ECC Go rules, and workflow read |
| 2 | Create README | done | Anchor table covers all topic docs |
| 3 | Typed boundaries | done | Includes `Role` and `problem.Code` patterns |
| 4 | Errors and logging | done | Documents `problem.Write` and audit JSON exceptions |
| 5 | Security boundaries | done | C1/C5/H7/H4 citations included |
| 6 | Idempotency and concurrency | done | Three `BeginReplay` outcomes documented |
| 7 | Persistence | done | pgx, parameterization, rows discipline, no new sqlmock |
| 8 | HTTP handlers | done | Handler anatomy and middleware order documented |
| 9 | Testing | done | No new sqlmock rule and race requirement documented |
| 10 | Package layout | done | Import direction law and constructor invariant pattern documented |
| 11 | Refactor playbook | done | Review/fix/baseline workflow documented |
| 12 | `.golangci.yml` | done | v2 config verified with golangci-lint v2.11.0 |
| 13 | Go Lint workflow | done | Uses action v9 and gates new issues from `origin/main` |

## Validation Results

| Level | Status | Notes |
|---|---|---|
| Static Analysis | pass | `golangci-lint config verify` passed via `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.0` |
| Unit Tests | N/A | Documentation/config only |
| Build | pass with scope | `go build ./apps/api/... ./internal/... ./tools/...` passed |
| Integration | N/A | No runtime code or server behavior changed |
| Edge Cases | pass | YAML parsed; front matter present; new-from-rev lint returned 0 issues |

## Files Changed

| File | Action | Lines |
|---|---|---|
| `wiki/standards/golang/README.md` | created | quality bar index |
| `wiki/standards/golang/typed-boundaries.md` | created | typed boundary rules |
| `wiki/standards/golang/errors-and-logging.md` | created | errors/logging rules |
| `wiki/standards/golang/security-boundaries.md` | created | security boundary rules |
| `wiki/standards/golang/idempotency-and-concurrency.md` | created | idempotency rules |
| `wiki/standards/golang/persistence.md` | created | persistence rules |
| `wiki/standards/golang/http-handlers.md` | created | handler rules |
| `wiki/standards/golang/testing.md` | created | testing rules |
| `wiki/standards/golang/package-layout.md` | created | package layout rules |
| `wiki/standards/golang/refactor-playbook.md` | created | refactor process |
| `.golangci.yml` | created | lint config |
| `.github/workflows/golangci-lint.yml` | created | CI workflow |
| `wiki/README.md` | updated | standards index entry |

## Deviations from Plan

- Used current documented golangci-lint action guidance: `golangci/golangci-lint-action@v9` with `version: v2.11`, instead of the plan's older `@v6` / `v1.64.8` snippet. Reason: current golangci-lint v2 docs and action docs align on v2 binaries for `version: "2"` config.
- Workflow runs `--new-from-rev=origin/main ./apps/api/... ./internal/... ./tools/...`. Reason: full lint over the backend currently reports 186 existing baseline issues, and `./...` also includes the existing `non_git` scratch package that does not build.
- Did not modify `.github/workflows/invariants.yml`. Reason: the task table explicitly says to create `.github/workflows/golangci-lint.yml` and keep invariants clean.

## Issues Encountered

- Local `golangci-lint` was not installed. Resolved by running the pinned v2 CLI through `go run`.
- Windows temp cleanup failed once while bootstrapping golangci-lint. Resolved by setting `GOTMPDIR` to a repo-local `.tmp/go-build`.
- `go build ./...` fails on existing `non_git` duplicate `main` functions. Verified the real backend scope with `go build ./apps/api/... ./internal/... ./tools/...`.

## Tests Written

| Test File | Tests | Coverage |
|---|---|---|
| N/A | 0 | Documentation/config only |

## Next Steps

- [ ] Code review via `/code-review`
- [ ] Create PR via `/prp-pr`
