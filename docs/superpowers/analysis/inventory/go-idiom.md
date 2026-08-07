# Lane: go-idiom

Scope: Go source under `internal/` and `apps/` only. Excludes `.gocache-build/`, `vendor/`, generated
`*.gen.go` (16 files, 37,383 lines) unless noted. Non-test, non-gen production code: 516 files,
80,179 lines. Test code: 465 files, 86,753 lines.

## Findings

| ID | Class | Finding | Evidence | Scale |
|----|-------|---------|----------|-------|
| GOI-01 | gap | `.golangci.yml` has no `formatters:` block at all (v2 splits gofmt/goimports/gofumpt into a separate stanza from `linters:`) — nothing enforces import grouping or canonical formatting through the lint gate. | `.golangci.yml:1-64` (no `formatters` key); `grep -n "formatters\|gofmt\|goimports\|gofumpt" .golangci.yml` → 0 hits | 1 config file, whole repo |
| GOI-02 | gap | CI lint runs `only-new-issues: true`, so any violation already in `main` before a PR touches that file is permanently invisible — the config can tighten but pre-existing violations never surface without a manual full-repo run. | `.github/workflows/golangci-lint.yml:22-27` | 1 workflow, unbounded backlog |
| GOI-03 | gap | High-value linters a serious Go repo runs are absent from `linters.enable`: `unused` (dead code), `ineffassign` (ineffectual assignments), `unparam` (unused params), `unconvert`, `goconst` (repeated literals), `noctx` (context-less HTTP calls), `nolintlint` (lints the nolint directives themselves), `wrapcheck` (error-wrap discipline — directly relevant to GOI-05 below), `misspell`, `prealloc`. Only 15 linters enabled total. | `.golangci.yml:8-23` enable list vs standard golangci-lint v2 catalog | 10 named linters off |
| GOI-04 | idiom | `panic()` used outside `main`/init in production code, including in `application` and `domain` layers reachable from live request paths (constructor guard panics, e.g. `audit.WithExports`, `auth.NewService`, plus enum-validation panics in `iam/domain`). | `internal/modules/iam/domain/model.go:231`, `internal/modules/iam/domain/catalog.go:159`, `internal/modules/audit/application/service.go:79,88`, `internal/modules/auth/application/service.go:168`; `grep -rln "panic(" internal apps --include=*.go \| grep -v _test.go` → 49 files, 110 call sites total | 49 files, 110 sites |
| GOI-05 | idiom | `fmt.Errorf` used without `%w` (error context lost, breaks `errors.Is/As` chain) at meaningful scale alongside heavy correct usage — not systemic but sizeable. | `grep -rn "fmt.Errorf(" internal apps --include=*.go \| grep -v _test.go` → 2047 total, 1608 with `%w`, 439 without (206 files contain at least one bare Errorf) | 439 bare-wrap call sites / 206 files |
| GOI-06 | idiom | Non-idiomatic capitalized/punctuated error strings, which `revive`'s `error-strings` rule (enabled in config) is supposed to catch — violations exist in shipped code, implying either the rule isn't actually firing in CI (see GOI-02: pre-existing, invisible to diff-scoped lint) or these predate the rule. | `internal/modules/approval/domain/sod.go:7,9` (`"SoD: ..."`), `internal/modules/iam/delivery/http/admin_handler.go:420,423` (`"Invalid roles"`, `"Exactly one role is required"`) | 6 `errors.New(` sites, 56 `fmt.Errorf(` sites starting capitalized |
| GOI-07 | idiom | God package: `approval/application` is 30 files / 8,755 lines / 583 exported symbols — 3–8x every sibling module's application layer. | `internal/modules/approval/application/*.go` (non-test) vs `documents/application` (17 files/2,439 lines/297 exported), `iam/application` (11/3,210/182), `templates/application` (12/1,449/153), `controlleddocuments/application` (2/1,099/115) | 30 files, 8,755 lines, 583 exported symbols in one package |
| GOI-08 | idiom | Constructors returning an interface type instead of a concrete struct (violates accept-interfaces/return-structs) in non-generated code — small in count but a real instance, not zero. | `internal/modules/approval/infrastructure/postgres_approval_repository.go:39` — `func NewPostgresApprovalRepository(...) ApprovalRepository` | 1 site outside `.gen.go` (33 total incl. generated `NewStrictHandler`/`NewStrictHandlerWithOptions` in every module's `api.gen.go`, which is oapi-codegen boilerplate, not hand-written) |
| GOI-09 | idiom | Goroutines spawned with `go func()` without a visible lifecycle primitive (no `sync.WaitGroup`, no `errgroup`, no returned cancel/done channel) at several call sites — cannot confirm leak by static read alone, flagged as inspection-worthy. | `internal/modules/documents/jobs/orphan_pending_sweeper.go:17`, `internal/modules/documents/jobs/session_sweeper.go:17`, `internal/modules/iam/application/cached_role_provider.go:61`, `internal/modules/templates/jobs/orphan_object_sweeper.go:61`, `apps/api/cmd/metaldocs-api/main.go:1028,1033,1295` | 10 `go func(` sites total; 7 have no visible WaitGroup/errgroup |
| GOI-10 | gap | Generics (Go 1.18+ type parameters) essentially unused: exactly one real site in ~80k LOC of production code, a private slice-conversion helper. Not a defect per se (Go doesn't require generics), but notable given the repo's age and the volume of hand-written `[]any` conversion/mapping helpers that generics would collapse. | `internal/modules/documents/delivery/http/fillin_handler.go:219` — `func toAnySlice[T any](items []T) []any` | 1 site |

## The five heaviest, with detail

**GOI-07 (approval/application god package).** One package carries 583 exported symbols across 30
files — roughly 3x the next-largest module's application layer (documents, 297) and 8x the smallest
comparable one (controlleddocuments, 115/2 files). It costs onboarding time (no sub-boundary inside
the package to orient by) and blocks safe parallel work (30 files in one `application` namespace means
near-certain merge contention). This is architecturally consistent with — but distinct evidence from —
the already-known M3 approval-kernel-extraction effort in the memory log; the file/symbol counts here
are the quantified version of that intuition.

**GOI-01 + GOI-02 (lint gate has a formatting hole and a permanent blind spot).** No `formatters:`
stanza means gofmt/goimports canonicalization is not enforced by the one lint job the repo runs, and
`only-new-issues: true` means anything already merged before a rule existed (or before a file was last
touched) never gets flagged again. Together these mean the `.golangci.yml` "quality bar" document
(its own header claims to be one) is only a bar for *new* lines, and even for new lines it doesn't
touch formatting. This directly explains why GOI-06 violations (revive error-strings should catch
them) exist in shipped code without a CI failure.

**GOI-03 (linter set is narrow for the enforcement burden this repo carries).** 15 linters enabled,
`default: none`. Missing `unused`, `unparam`, `ineffassign`, `wrapcheck`, `noctx`, `goconst`,
`nolintlint` are not exotic — they are standard in any `default: none` golangci-lint v2 config for a
production Go service. `wrapcheck` specifically would have caught GOI-05 mechanically instead of by
grep. `nolintlint` would keep the 10 existing `//nolint` directives (GOI list below) honest — right now
nothing checks that a `//nolint:X` still needs to exist or has a required explanation.

**GOI-04 (panic used as control-flow/guard in request-reachable code).** 110 panic sites across 49
files is not itself alarming (many are legitimate: init-time validation, "invalid enum value" that
truly cannot happen post-validation, constructor precondition guards that fire once at wiring time).
But some sites are in `application`-layer constructors and domain enum stringification that are, in
principle, reachable from data that crossed a network boundary earlier (e.g. `iam/domain/model.go:231`
panics on an "invalid capability" string). Whether each site is wiring-time-only or data-path-reachable
needs per-site judgment, not a blanket rule — flagged for the synthesis step to triage, not to fix here.

**GOI-05 (439 non-wrapped `fmt.Errorf` sites, 206 files).** The dominant pattern (1,608 of 2,047 sites,
79%) *does* wrap with `%w`, so this is not systemic sloppiness — but the remaining 21% means
`errors.Is`/`errors.As` chains silently break at unpredictable points, and nothing catches a
regression from wrapped to bare (no `wrapcheck`, see GOI-03). The inconsistency itself — not the
raw count — is the finding: a reviewer cannot assume `fmt.Errorf` in this repo preserves the chain.

## What is actually fine

- **No `util`/`helpers`/`common`/`shared` catch-all package** anywhere under `internal/` or `apps/`
  (`find internal apps -type d -iname "util*" -o -iname "helpers*" -o -iname "common*" -o -iname "shared*"`
  → 0 hits). Package naming discipline holds.
- **`go.mod` hygiene is clean**: Go 1.25.0 declared (toolchain 1.26.4 available), 28 direct deps, 76
  indirect, **zero `replace` directives**. `vendor/` exists with a real `modules.txt` (genuine `go mod
  vendor` output, not a stray checkout) but `GOFLAGS=-mod=mod` is set in `scripts/start-api.ps1:348`
  and `scripts/start-jobs.ps1:35`, so vendor mode is not actually the load path for local dev — worth
  a note for synthesis (is `vendor/` load-bearing or dead weight?) but not a go-idiom defect by itself.
- **Interface placement leans consumer-side**: 37 interface declarations live in `application`
  packages (the consumer) vs 17 in `infrastructure` (the producer) and 34 in `domain` — the
  accept-interfaces convention is the dominant pattern, not the exception. Only 1 real (non-generated)
  constructor returns an interface (GOI-08) — the rest of the 33-count is oapi-codegen boilerplate.
- **No `context.Context` stored in a struct field** anywhere (`grep -rnE "^\s+[a-zA-Z]+\s+context\.Context\s*$"` →
  0 hits) — a genuinely common Go anti-pattern that this codebase does not have.
- **All outbound HTTP requests use `NewRequestWithContext`** (0 uses of bare `http.NewRequest`, 4 of
  `NewRequestWithContext`) — small surface but clean on this axis.
- **`context.Background()` in non-test code is confined to true root call sites**: 4 `main.go`
  entrypoints, `migrate.go`, `requesttrace/context.go`, and 2 uses in `iam/presence` that start a
  detached heartbeat goroutine explicitly decoupled from the request context (a legitimate pattern for
  a long-lived connection handler, not a lost-context bug) — 16 total sites, none look like an
  accidental drop of an available request context.
- **`rows.Close`/`rows.Err` discipline reads consistent** on inspection: 113 `.Query(` call sites,
  103 `defer ...Close()`, 84 explicit `rows.Err()` checks — the gap between 113 and 103 is plausibly
  `QueryRow` (no `Close` needed) rather than leaks; `sqlclosecheck`/`rowserrcheck` are both enabled in
  `.golangci.yml`, which is the correct mechanical guard for this class regardless.
- **Generics non-use (GOI-10) is not itself a defect** — flagged as gap/observation only because the
  volume of hand-rolled `[]any` conversion code is large enough that it's worth the synthesis step
  asking "would generics collapse N of these," not because absence of generics is wrong on its own.

## Unverified / needs judgment

- Whether the 110 `panic()` sites (GOI-04) are actually reachable from untrusted/network-derived data,
  or are all wiring-time/enum-exhaustiveness guards that are safe by construction — needs per-site
  triage, not a blanket verdict.
- Whether `vendor/` (real, `go mod vendor`-produced) is load-bearing for any build/CI path, given local
  dev scripts force `-mod=mod` — could not verify from `.golangci.yml`/scripts alone; needs a check of
  `.github/workflows/*.yml` build steps and `Dockerfile`s.
- Whether GOI-06's revive violations mean the `error-strings` rule silently never fires (a config bug)
  or whether these lines simply predate the rule and never got PR-touched since (GOI-02's blind spot) —
  could not distinguish without git blame per line, out of lane budget.
- Could not run `golangci-lint` itself (not installed in this environment) to get an authoritative
  current-violation count — all lint-adjacent findings here are from direct grep against the patterns
  each linter would catch, not from an actual lint run.

## Commands run

```
go version
cat go.mod
cat .golangci.yml
find internal apps -name "*.go" -not -path "*/.gocache-build/*" -not -path "*/vendor/*" -not -name "*_test.go" -not -name "*.gen.go" | wc -l   # 516
... | xargs wc -l (via while-read loop, xargs wc -l batching was unreliable)                                                                   # 80179
find ... -name "*_test.go" ...                                                                                                                  # 465 files, 86753 lines
find ... -name "*.gen.go" ...                                                                                                                    # 16 files, 37383 lines
grep -rn "fmt.Errorf(" internal apps --include=*.go | grep -v _test.go | wc -l          # 2047
grep -rn "fmt.Errorf(.*%w" internal apps --include=*.go | grep -v _test.go | wc -l      # 1608
grep -rn "errors.Is(" / "errors.As(" internal apps --include=*.go | grep -v _test.go    # 445 / 125
grep -rn "= errors.New(" internal apps --include=*.go | grep -v _test.go | wc -l        # 294
grep -rln "panic(" internal apps --include=*.go | grep -v _test.go                      # 49 files
grep -rn "panic(" internal apps --include=*.go | grep -v _test.go | wc -l               # 110
grep -rln "context.Background()" internal apps --include=*.go | grep -v _test.go        # 8 files, 16 sites
grep -rn "^type [A-Za-z_]* interface" internal apps --include=*.go | grep -v _test.go | grep -v .gen.go | wc -l   # 248
grep -rln "^type [A-Za-z_]* interface" ... | sed -E 's#.*/(application|domain|infrastructure|delivery|api|jobs)/.*#\1#' | sort | uniq -c
grep -rnE "^func New[A-Za-z0-9_]*\(.*\) [A-Za-z]+(Repository|Port|Interface|Provider)( |$|,)" ... | grep -v .gen.go   # 1 real site
grep -rn "\[T " internal/modules/documents/delivery/http/fillin_handler.go                                          # generics site check
grep -rn "//nolint" internal apps --include=*.go                                        # 10 sites
grep -rln "sync.WaitGroup" / "errgroup" internal apps --include=*.go | grep -v _test.go
grep -rn "sync.Mutex\|sync.RWMutex" internal apps --include=*.go | grep -v _test.go | wc -l   # 19
grep -rn "^\s*go func(" internal apps --include=*.go | grep -v _test.go                 # 10 sites
grep -rn "\.Query(" / "defer.*\.Close()" / "rows.Err()" internal apps --include=*.go | grep -v _test.go   # 113 / 103 / 84
grep -rn "defer.*Rollback" / "\.Rollback(" internal apps --include=*.go | grep -v _test.go   # 56 / 104
grep -rn "http.NewRequest(" / "http.NewRequestWithContext(" internal apps --include=*.go | grep -v _test.go   # 0 / 4
grep -rnE "^\s+[a-zA-Z]+\s+context\.Context\s*$" internal apps --include=*.go | grep -v _test.go   # 0
grep -n "formatters\|gofmt\|goimports\|gofumpt" .golangci.yml                            # 0 hits
cat .github/workflows/golangci-lint.yml
for p in approval/application documents/application iam/application templates/application controlleddocuments/application; do files/lines counts; done
find internal apps -type d -iname "util*" -o -iname "helpers*" -o -iname "common*" -o -iname "shared*"   # 0 hits
```
