# Milestone 2 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-22 (RE-VALIDATION under HS-4 after fix feature F2.4)  ·  **Verdict:** see C7.
> The validator judged and wrote this file only; it edited no source and flipped no status.
> The prior **FAIL** verdict (commit `3e6cc227`) is preserved in git history and superseded by this file.

## Re-validation context (HS-4)

The prior run returned **FAIL** for exactly one blocker: a non-deterministic CI-gating test,
`scripts/api-lint` `TestSeedRegistryParity_RegistryNotSeeded`, which indexed `caps[0]` of the
map-randomized `AllCapabilities()` and flaked (~1/30 full-suite runs) when the random first cap was
the F2.1c-introduced deferred cap `CapDistributionRead` (the checker exempts deferred caps → `want 1,
got 0`). The named minimum fix feature **`f2.4-deferredcap-parity-test-determinism`** (commit
`3e6cc227`) is now closed. C1/C3/C5 PASSed previously, unaffected by a test-only fix — carried forward
with spot-confirmation that nothing regressed.

## Inputs loaded (all present, all readable)

- Milestone spec `../milestone.md` (incl. the 2026-06-22 operator correction: F2.2 authz is tier-1 middleware).
- **Six** feature folders, each with `spec.md` + `plan.md` + `evidence.md`: F2.1a, F2.1b, F2.1c, F2.2,
  F2.3, **F2.4** (new). All confirmed present this run.
- Program `README.md` (M0+M1 passed; M2 in-progress) + governing `mission.md`.
- F2.4 diff: `git show --stat 3e6cc227` — touches **only** `scripts/api-lint/registry_rules_test.go`
  (17 lines, +sort import / sort+skip-deferred selection / `t.Fatal` guard) plus the F2.4 feature docs
  and the carried-in prior validator verdict. **No** product / checker / registry / contract / migration
  / authz change.
- Aggregate M2 diff still `f357fb15^..HEAD`; sacred scopes re-confirmed empty (see C4).

Environment: live Postgres reachable via the running `metaldocs-postgres` container (healthy, 2h up,
localhost:5433). DSN formed from the **container's own runtime env** (`docker inspect` — not the repo
`.env`, which remains unread per CLAUDE.md). The distribution integration suite was therefore re-run
**live, not skipped** (`testdb.Open(t)` factory, 69.36s real work).

## C1 — Spec & plan conformance (per feature)

Carried forward from the prior PASS for F2.1a/b/c, F2.2, F2.3 (test-only fix did not touch them);
**F2.4 added and judged fresh.** Each feature has `spec.md` (filled `Approved before code:` +
populated Interview/justification), an execution-shaped `plan.md`, and an `evidence.md` whose
acceptance maps row-for-row to the spec Validation Gate.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1a | ✅ view shape == spec (5 cols, 3-leg UNION, DISTINCT precedence) | ✅ | ✅ (`v_cd_grantee`/0243/search untouched) | live integration PASS (C2 prior) |
| F2.1b | ✅ `(tenant_id, area_code, area_name)` 1:1 | ✅ | ✅ | live integration PASS (prior) |
| F2.1c | ✅ denominator-only OpenAPI; cap deferred; types generated | ✅ | ✅ | api-lint=0; FE types denominator-only |
| F2.2 | ✅ projects two views + iam port; **tier-1 guard** per 2026-06-22 correction | ✅ | ✅ | integration live-PASS this run incl. `fail_closed` |
| F2.3 | ✅ consumes generated types; no hand-rolled shapes | ✅ | ✅ | FE 16/16; both FE reviewers APPROVE on record |
| **F2.4** | ✅ test-only; deferredCaps consumed (not redefined); guard not weakened | ✅ G1–G5 | ✅ (no product/contract/migration/authz/registry change) | `git show --stat 3e6cc227` = test+docs only; gates re-run in C2 |

F2.4 spec/plan/evidence approved pre-code (2026-06-22), opened explicitly under HS-4 by the prior
verdict. The fix **does not weaken the guard**: `registry_rules_test.go:158` still asserts exactly
`countRule == 1`, `:161` still asserts the message names the omitted cap, and the omitted cap is now
guaranteed non-deterministic-free (sorted, first non-deferred, with a `t.Fatal` backstop if none) — so
the test still bites on real seed/registry drift. **C1 PASS.**

## C2 — Gates re-run, isolated (validator-run, this session)

| Feature / gate | Command re-run (validator) | Real output | Pass? |
|---------|----------------|-------------|-------|
| **F2.4 G1 — parity determinism** | `go test ./scripts/api-lint/... -run TestSeedRegistryParity -count=100` | `ok metaldocs/scripts/api-lint 4.727s` | ✅ |
| **F2.4 G2 — package stable** | `go test ./scripts/api-lint/... -count=50` | `ok metaldocs/scripts/api-lint 225.797s` | ✅ |
| **F2.4 G1 — forced-uncached** | `go test ./scripts/api-lint/... -count=1 -run TestSeedRegistryParity` ×6 separate invocations | all `exit=0` (map re-randomized each fresh process) | ✅ |
| **F2.4 G3 — full suite green** | `go test ./...` (multiple runs; grep `FAIL|panic`) | **zero FAIL/panic lines**; `api-lint` `ok`/cached, exit 0 every run | ✅ |
| F2.4 G5 — diff scope | `git show --stat 3e6cc227` | only `registry_rules_test.go` (+ docs) | ✅ |
| F2.2 integration (incl. fail_closed G5) | `go test -tags=integration -run TestDistributionCoverage ./internal/modules/distribution/infrastructure/... -v` (live PG) | `--- RUN TestDistributionCoverage/fail_closed`; `--- PASS: TestDistributionCoverage (69.36s)`; `ok … 72.702s` | ✅ |
| api-lint -strict (G7) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)`, exit 0 | ✅ |
| cilint / hgcrossmodule (G8) | `go run ./tools/cilint ./internal/modules/distribution/...` | 0 findings, exit 0 | ✅ |
| FE distribution (V6) | `npx vitest run useDistribution DocumentDistributionPage` | **16/16 passed**, exit 0 | ✅ |
| FE typecheck (V7) | `npx tsc --noEmit` (web) | exit 0 | ✅ |
| build / vet | `go build ./...`; `go vet ./...` | exit 0 / exit 0 | ✅ |

Combined this run: **>150 executions** of the previously-flaky `TestSeedRegistryParity_RegistryNotSeeded`
(100 in-process + 50-run package suite + 6 forced-uncached fresh processes + full-suite runs) with
**zero** failures. The historical ~1/30 flake would have surfaced. **The prior C2 blocker is cleared.**
**C2 PASS.**

## C3 — Senior review of the aggregate milestone diff

Carried forward from the prior PASS (the substance diff `f357fb15^..HEAD` was staff-engineer clean:
single source of truth in the `0245` view, repo projects not re-derives, module-boundary clean, no dead
code, no feature breaking another). The F2.4 increment is a **17-line test-only** change that *removes*
a fragility rather than adding code: it sorts the cap slice and picks the first non-deferred cap with a
`t.Fatal` guard. No product, contract, or split-brain surface touched. The one aggregate-only defect the
prior run flagged (test non-determinism) is now resolved at its root — the test no longer depends on map
iteration order. Staff-engineer bar met on both product **and** test-suite determinism. **C3 PASS.**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| BE canonical (`backend-api-qa-checklist` + CI guards + api-lint=0 + integration) | **pass** | api-lint -strict=0; cilint/hgcrossmodule=0 on distribution; build/vet=0; distribution integration live-PASS incl. fail_closed; **`scripts/api-lint` guard now deterministic** (was the sole flaky guard) |
| FE canonical (`screen-definition-of-done` D2) | pass | distribution 16/16; tsc=0; both FE reviewers APPROVE on record (prior); V1 grep=0 |
| Regression vs M0 | pass | M0 dirs not in diff |
| Regression vs M1 | pass | `MOCK_STATS`/`MOCK_ACTIVITY` grep = 0 (re-run this session) |
| Publish path / `v_cd_grantee` / 0243 / search untouched | pass | `git diff f357fb15^..HEAD` on all four scopes = empty (re-confirmed this session) |
| `go build` / `go vet` | pass | exit 0 |
| `go test ./...` | **deterministically green** | zero FAIL/panic across this session's full-suite runs; the prior ~3% flake eliminated |

**C4 PASS** — the BE workflow-class regression that previously failed (a non-deterministic CI guard) is
resolved; all guards green deterministically.

## C5 — Quality-bar re-measure + retrospective

Carried forward (unchanged by a test-only fix) and spot-re-measured:

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Screen-DoD D2: no illustrative on Distribuição denominator | violated | **met** | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" …/DocumentDistributionPage.tsx` = 0; `MOCK_DISTRIBUTION` in `features` = 0 (re-run) |
| Denominator live (real producer) | none | **met** | distribution integration live-PASS (69.36s real PG); FE 16/16 typed to generated contract |
| Numerator honesty (no fabrication) | n/a | **met** | no `role`/read/ack field in generated `Distribution*` types (api-lint=0) |
| **CI-guard determinism (prior blocker)** | flaky ~1/30 | **met — root cause fixed** | F2.4 makes the fixture cap selection deterministic (sort + skip `deferredCaps`); the *checker* was always correct (rightly exempts deferred caps); >150 clean re-runs this session |

Root cause fixed, not symptom-patched: the determinism fix targets the *test's fixture selection* (the
actual defect), leaves `AllCapabilities()` ordering / the checker / `deferredCaps` contents untouched
(non-goal honored), and the guard's assertions are preserved so it still detects real drift.

- **Could it be built better?** No outstanding blocker. The F2.3 V2/V3 runtime-screenshot gap remains
  the operator-spot-checkable bounded item already judged **acceptable** below — not a blocker.

## C6 — Forbidden-list

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean* (each G/V mapped to a validator-run command).
- [ ] Fixture/mock passed off as real-provider proof — *clean* (distribution proven on **live PG**, 69.36s real work, not skipped).
- [ ] Consumer contract guessed — *clean* (generated types; F2.4 consumes `deferredCaps`, does not redefine).
- [ ] Split-brain — *clean* (union defined once in `0245`).
- [ ] Self-judged close / validator edited code — *clean* (validator wrote only this verdict file; F2.4 was implemented by the fix-feature session, judged here independently).
- [ ] Scope drift — *clean* (F2.4 is test-only, opened under HS-4, `git show --stat` confirms only the test file + docs).
- [ ] **Symptom-patch / suite-green-as-pass — NO LONGER TRIGGERED.** The prior run's qualified trigger
  (non-deterministic CI guard) is resolved at its **root**: the test's map-order dependence is removed,
  not masked; the guard still fires exactly 1 violation on real drift and names the cap. >150 clean
  re-runs this session, including forced-uncached fresh processes that re-randomize the map.

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- **All checks pass:** C1 (six features each with conforming spec/plan/evidence; F2.4 consumer-contract
  clean and guard not weakened), C2 (every named gate re-run by the validator from clean state — the
  prior flaky parity test now deterministic across >150 executions; distribution integration live-PASS
  incl. fail_closed), C3 (aggregate diff staff-engineer clean; F2.4 removes a fragility), C4
  (workflow-class QA + M0/M1 regression green; sacred scopes empty; full suite deterministically green),
  C5 (quality bars met at root, not symptom-patched), C6 (forbidden-list clean; prior symptom-patch
  trigger resolved).
- **The sole prior blocker is cleared.** `f2.4-deferredcap-parity-test-determinism` (commit `3e6cc227`)
  is a test-only fix touching exactly `scripts/api-lint/registry_rules_test.go`; it makes
  `TestSeedRegistryParity_RegistryNotSeeded` deterministic by sorting the cap slice and selecting the
  first non-deferred cap with a `t.Fatal` backstop, leaving product/checker/registry/contract/migration/
  authz untouched. The guard still bites on real drift.
- Handed back to the **main session** to flip M2 status (`README.md`) on this PASS and present the
  **HS-1 operator gate**.

### On the F2.3 V2/V3 runtime-screenshot bounded gap (explicit call, unchanged)

**Acceptable for milestone close — not a blocker.** The producer is proven live end-to-end (distribution
integration against real PG, validator-re-run this session), the FE consumes the generated contract
types (tsc=0), the denominator logic is covered by 16/16 typed tests, and both FE reviewers APPROVE on
record. The missing artifact is a screenshot, not a capability; the operator HS-1 review can spot-check
at runtime. This gap does **not** affect the PASS.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — now unblocked by this PASS.
> - Status flipped in `README.md`: pending main session (only on this PASS).
