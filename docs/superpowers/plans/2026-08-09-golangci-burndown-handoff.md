# Handoff: golangci-lint backlog burn-down (fresh session, from zero)

**Date:** 2026-08-09. **Written by:** the CI-restructure session, immediately after PR #97 merged to main (`d1626a6f`).
**Read this file first; it is self-contained.**

## Operator ruling (verbatim, binding)

> "Remover ony-new issuses, ou seja, não deve nem existir essa cnfiguração nos lints, CI e etc."

Meaning: `only-new-issues` must **not exist anywhere** — not in ci.yml, not in any lint config. Lint is whole-tree blocking. Companion doctrine (earlier, verbatim): "Lint should capture it no bypass, fallback only to show green I want to catch the erros" — never delete/skip/downgrade a check to reach green; fix the findings.

Operator sequencing decision (verbatim): "Merge #97, depois branch burn-down, mas em uma outra sessão do 0" — #97 is merged; this session is that "outra sessão".

## The task

1. Branch off `main` (e.g. `ci/golangci-burndown`).
2. In `.github/workflows/ci.yml`, job `lint-go`: **delete** `only-new-issues: true` and delete the TRANSITIONAL comment block above the golangci-lint step (it names this burn-down as the milestone that deletes it). The step becomes plain whole-tree blocking golangci-lint.
3. Burn down the whole-tree backlog: ~1078 findings at configured scope (`--timeout=5m ./apps/api/... ./internal/... ./tools/...`, config `.golangci.yml`, golangci-lint v2.11 via action v9.3.0). Spec §11.1 of `docs/superpowers/plans/2026-08-08-ci-phases-1-5.md` describes the backlog (1173 whole-tree / ~1078 configured scope).
4. Merge via PR **only when whole-tree lint is green**. Mechanical constraint: `lint-go` is in `required`'s needs (`scripts/required-gate.jq` demands literal "success" from all 4 jobs) — so once step 2 lands on the branch, that PR is red until the backlog is zero. That is intended ("first red PR is the goal"). Nothing else can merge through a red lint-go only if it introduces new issues — main keeps only-new-issues until THIS branch merges, so other PRs are unaffected meanwhile.
5. Never loosen `.golangci.yml` to reach green (that is bypass). Never `--no-verify`. `bypass_actors` stays `[]`.

Run locally with the same scope/config to iterate: `golangci-lint run --timeout=5m ./apps/api/... ./internal/... ./tools/...` (v2.11).

## Constraints & context

- CI shape (post-#97): `.github/workflows/ci.yml` = 4 jobs (verify, test-integration, security, lint-go) + `required` aggregator; sole required context `required`, strict, ruleset 20560142. Runbook: `docs/runbooks/ci-required-gate-and-hardening.md`.
- Known lint-red file already identified by bots: `tools/verify/main_test.go` (`errcheck` unchecked `f.Close`; `gocritic` exitAfterDefer).
- Memory rules: sonnet implementers, haiku mechanical, never fable workers; ≤15 concurrent; commit without asking, never push without permission (pushing the burn-down branch to open its PR needs operator OK or is implied by the ruling — ask once).
- Suite discipline: mechanical lint fixes still compile-check with `go vet -tags integration ./...` before commit (integration-tag compile gap memory).

## Accepted-open items adjudicated to this backlog era (from PR #97 bot threads, resolved-to-merge 2026-08-09)

These were bot review threads on #97, resolved to unblock merge (thread-resolution rule), dispositions recorded here — they are real follow-ups, not dismissed:

- **P1 (codex):** `tools/verify/main.go` — a PR touching only `tools/verify/registry.go` can edit a check's Argv/paths without that check running. Consider: registry.go in every check's effective paths, or force `--profile=changed` to select all when registry.go changes.
- **P2 (codex) ×2:** codegen-drift-backend/frontend path scoping misses `go.mod`/`go.sum`/lockfile generator upgrades — add those to the checks' Paths.
- **coderabbit Major:** `nightly.yml` checkouts before artifact uploads should set `persist-credentials: false`.
- **coderabbit Major:** `nightly.yml` backend-ready wait loop exits 0 after 30 failed health checks — must `exit 1`.
- **coderabbit Major:** `tools/verify/main_test.go` lint failures (see above — part of the burn-down proper).
- **coderabbit Minor:** stale wording in `docs/superpowers/reports/2026-08-08-gosec-govulncheck-measurement.md`; MD040 fence language in the runbook; `changed_test.go` asserts a clean working tree (flaky locally).
- Full pre-merge accepted-open list (e2e/axe nightly-only, .golangci/.gitleaks loosening ungated, Test-RealCodeChange fixture holes, no push:main trigger, docx-renderer duplicate compute, D13 ratchet deviation, ME-15→ROADMAP 4.7): `.superpowers/sdd/progress.md` in the repo root's `.superpowers/` (git-ignored scratch — may be gone; the deletion ledger `docs/superpowers/reports/2026-08-08-workflow-deletion-ledger.md` is the committed authority).

## Definition of done

- `only-new-issues` absent from the entire repo (grep proves it).
- `golangci-lint run` (configured scope) exits 0 whole-tree.
- PR merged with `required` green for the right reason.
- Findings fixed at root cause; any allowlist/nolint addition needs an in-code justification comment and operator visibility.

## Closure status (2026-08-09, session 2)

Burn-down executed in 3 waves + orchestrator passes; branch `ci/golangci-burndown`:

- `aa939136` — ci.yml: only-new-issues + TRANSITIONAL block deleted; lint-go whole-tree blocking.
- `09f3a9d9` — wave 1 (mechanical: revive/errcheck/errorlint/staticcheck/gocritic/exhaustive), 247 files, 891 findings.
- `eda1ca26` — wave 2 (contextcheck/gocognit/gocyclo, 155 findings) + api-lint allowlist re-true + full-suite green.
- `801114b8` — wave 3 (gosec/nilerr/bodyclose, 24 findings incl. 3 in out-of-scope apps/jobs+worker).

**Definition-of-done evidence:**
- only-new-issues: zero live-config references (grep whole repo — remaining matches are historical analysis/spec/report docs describing the pre-removal state; ledger doc documents the transitional label as history).
- `golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 ./apps/api/... ./internal/... ./tools/...` → **0 issues, exit 0** (uncapped; the default caps HIDE findings — any future gate run must uncap). `./apps/...` (jobs/worker, outside CI scope) also 0.
- `go build ./...`, `go vet -tags integration ./...`, `go test ./...` green (full suite re-run at wave-2 gate; wave-3 touched packages re-run green; final full suite run at closure).
- nolint additions on the branch (all with in-code justification): 4 revive (OpenAPI operationId-pinned ServerInterface method names), 18 gosec (12 G706 slog-JSONHandler FPs, 4 G118 cancel-returned-as-stop FPs, 1 G117 redacted-before-marshal, 1 G705 cached self-produced JSON replay), 2 nilerr (best-effort WalkDir, shutdown-mid-batch).

**Deviations / operator-visibility items:**
1. `submit_service_test.go` structural text-scan test widened from the old monolithic function's source span to whole-file scan (wave-2 decomposition made the old span meaningless); the authz-before-UPDATE ordering assertion is preserved.
2. api-lint allowlists re-trued after decomposition: seed-chokepoint line pins remapped (same 13 sites, monotonic shift, no new seeds); tripwire allowlist +2 extracted documents-repo helpers (`writeCreateDocumentSnapshotAndPlaceholders`, `insertCommittedRevision`) — callers `CreateDocumentTx`/`CommitUpload` hold `authz.Require(CapDocumentEdit)` in the same tx (verified).
3. Latent shutdown defect FIXED in all 3 binaries' mains: os.Exit skipped deferred cleanups (otel shutdown, signal stop, deps cleanup, rate-limit store close) on failure paths — mains now `os.Exit(runMain())` with defers running on every path.
4. A pre-existing `//nolint:gocyclo` in release_coordinator was removed by genuine decomposition (net nolint delta on gocyclo: −1).
5. **CI lint scope gap (pre-existing, not changed by this branch):** lint-go covers `./apps/api/... ./internal/... ./tools/...` only — `apps/jobs` and `apps/worker` are unlinted in CI. Their findings were fixed anyway on this branch (now 0); adding them to the scope is a one-line args change when the operator wants it.
6. The non-lint accepted-open items (nightly.yml persist-credentials + exit-1 health loop, verify path-scoping P1/P2) remain open — separate follow-ups, not this branch's scope.

**Remaining to done:** push branch + open PR (operator OK required), `required` green in CI, merge.
