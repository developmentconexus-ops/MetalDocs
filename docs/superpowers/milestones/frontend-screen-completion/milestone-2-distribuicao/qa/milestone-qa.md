# Milestone 2 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-22  ·  **Verdict:** see C7.
> The validator judged and wrote this file only; it edited no source and flipped no status.

## Inputs loaded (all present, all readable)

- Milestone spec `../milestone.md` (incl. the **2026-06-22 operator correction**: F2.2 authz is tier-1 middleware, not the in-tx write-pattern).
- All five features' `spec.md` + `plan.md` + `evidence.md` (F2.1a, F2.1b, F2.1c, F2.2, F2.3).
- Program `README.md` (M0+M1 passed; M2 in-progress) + governing `mission.md`.
- Aggregate M2 diff: `git diff --stat f357fb15^..HEAD` — 62 files, +6268/−984; first M2 source commit `f357fb15`, HEAD `b27bf2c6`.

Environment: live Postgres reachable via the running `metaldocs-postgres` container (localhost:5433). DSN formed from the **container's own runtime env** (`docker inspect` — not the repo `.env`, which remains unread per CLAUDE.md). All three integration suites were therefore **run live, not skipped**.

## C1 — Spec & plan conformance (per feature)

Each feature has `spec.md` (with a filled `Approved before code:` line + a populated Interview record), an execution-shaped `plan.md`, and an `evidence.md` whose acceptance table maps row-for-row to the spec Validation Gate. Consumer-contract-first throughout: distribution reads only the two published views + the ADR-0029 iam port; the FE consumes generated types only.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1a | ✅ — view shape == spec contract block (5 cols, 3-leg UNION, DISTINCT precedence) | ✅ | ✅ (`v_cd_grantee`/0243/search untouched; no Go) | spec approved 2026-06-21; integration suite live-PASS (see C2) |
| F2.1b | ✅ — `(tenant_id, area_code, area_name)` 1:1 projection | ✅ | ✅ (no extra cols; no base-table change; no Go) | spec approved 2026-06-21; integration suite live-PASS |
| F2.1c | ✅ — denominator-only OpenAPI; cap deferred; FE+Go types generated | ✅ | ✅ (no handler/SQL/migration/numerator/`role`/pre-grant) | cap anchors ×4 verified; api-lint=0; FE types denominator-only |
| F2.2 | ✅ — projects the two views + iam port; **tier-1 guard per the 2026-06-22 correction** | ✅ | ✅ (no migration/publish/search/`v_cd_grantee`/role-grant/FE) | G1–G11 live-PASS; no in-tx `authz.Require` in module |
| F2.3 | ✅ — consumes generated types; no hand-rolled shapes | ✅ | ✅ (no backend/regen/primitive/M0-M1 touch) | V1–V10; both FE reviewers APPROVE on record |

All approval lines filled with date+operator; all interview records populated. F2.2's spec approval correctly records the **tier-1 correction** ("tier-1 middleware read-guard … correcting the milestone.md write-pattern mandate") — F2.2 is **not** failed for following it. **C1 PASS.**

## C2 — Gates re-run, isolated (validator-run, live, from clean state)

| Feature / gate | Command re-run (validator) | Real output | Pass? |
|---------|----------------|-------------|-------|
| F2.1a integration | `go test -tags=integration -run TestObligatedReaders ./internal/modules/controlleddocuments/infrastructure/...` (live PG 5433) | `ok … 97.278s` | ✅ |
| F2.1b integration | `go test -tags=integration -run TestProcessAreaName ./internal/modules/taxonomy/infrastructure/...` (live PG) | `ok … 116.339s` | ✅ |
| F2.2 integration (G1–G6) | `go test -tags=integration -run TestDistributionCoverage ./internal/modules/distribution/... -v` (live PG) | **9/9 PASS** incl. `fail_closed` (G5) + `recipients_pagination_null_bucket`; `ok … 96.526s` | ✅ |
| api-lint -strict (G7) | `./scripts/api-lint/api-lint.exe -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)`, exit 0 | ✅ |
| hgcrossmodule (G8) | `go run ./tools/cilint/... ./internal/modules/distribution/... .../controlleddocuments/... .../taxonomy/...` | exit 0 | ✅ |
| permissions table (G6) | `go test ./apps/api/cmd/metaldocs-api/... -run 'TestPermissions|TestEveryCapSeededOrDeferred' -count=1` | `ok … 7.159s` | ✅ |
| iam cap registration | `go test ./internal/modules/iam/domain/... -count=1` | `ok … 1.006s` | ✅ |
| FE distribution tests (V6) | `npx vitest run useDistribution.test.tsx DocumentDistributionPage.test.tsx` | **16/16 PASS** | ✅ |
| FE typecheck (V7) | `npx tsc --noEmit` (web) | exit 0 | ✅ |
| build / vet | `go build ./...`; `go vet ./...` | exit 0 / exit 0 | ✅ |
| migration ledger / shapes | live `information_schema` + `schema_migrations` | 0245→1, 0246→1; views 5 cols / 3 cols | ✅ |
| **full `go test ./...`** | `go test ./...` (run twice) | **run 1: FAIL** `TestSeedRegistryParity_RegistryNotSeeded` (`want 1, got 0`); **run 2: exit 0** | ⚠️ **FLAKY — see C6** |

The FE full vitest baseline (4 files / 36 tests: `InboxPage`, `DocumentEditorPage` ×2, `templates.create`) was confirmed **identical and pre-existing**, none import F2.3 modules; distribution targets are green. That baseline is **not** held against M2.

**C2 finding (non-deterministic CI guard):** `go test ./...` is **not deterministically green**. `scripts/api-lint` test `TestSeedRegistryParity_RegistryNotSeeded` failed on the first full-suite run and passed on the second + on 40 isolated runs. Root cause established (see C6): `iamdomain.AllCapabilities()` iterates a **map** (`model.go:173-179`, randomized order); the test picks `caps[0]` as the "missing-from-seed" cap and asserts exactly 1 parity violation — but the checker legitimately **exempts deferred caps** (`registry_rules.go:700`). When the random `caps[0]` lands on `CapDistributionRead` (the **sole** deferred cap, introduced by F2.1c), zero violations fire → spurious FAIL. Probability ≈ 1/30 per full-suite run.

## C3 — Senior review of the aggregate milestone diff

Reviewed `f357fb15^..HEAD` as one unit.

- **Single source of truth (no split-brain):** the obligated-set union + DISTINCT precedence is defined in **exactly one place** — the `0245` view. `coverage_repository.go` *projects* it (`FROM metaldocs.v_cd_obligated_readers`, lines 51/108/129/153/265) and **does not re-derive** from base tables (grep for `controlled_document_*_grants`/`UNION` in the repo = none). The `user_grant`/`area_grant`/`company_scope` literals in the repo are the fail-closed `validateSource` guard + DTO mapping — consuming the contract, not re-encoding it.
- **Module-boundary clean:** distribution reads only `metaldocs.v_*` views + the iam `DisplayNames` port; `hgcrossmodule` = 0 across all touched modules.
- **No dead code / no superseded approach:** the FE deletes `MOCK_DISTRIBUTION` + `distributionMeta.ts` at root (not flag-hidden); numerator components route to one shared `TrackingPendingNote`.
- **No feature broke another:** M0/M1 surfaces untouched (dashboard/home/app dirs empty in the diff; `MOCK_STATS`/`MOCK_ACTIVITY` greps still 0).
- Findings: the C2/C6 test-determinism regression is the **one** aggregate-only defect; it is a test-design fragility newly *exposed* (not authored) by F2.1c.
- Staff-engineer bar met on the **product** diff? ✅. On the **test suite's determinism**? ❌ (one flaky CI-gating test).

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| BE canonical (`backend-api-qa-checklist` + 6 CI guards + api-lint=0 + integration) | **pass-with-one-flaky-guard** | api-lint=0, hgcrossmodule=0, build/vet=0, integration 9/9 live; **but** the `scripts/api-lint` package guard is non-deterministic (C2/C6) |
| FE canonical (`screen-definition-of-done` D2: both reviewers APPROVE + live denominator) | pass | `frontend-screen-reviewer` + `frontend-code-reviewer` both APPROVE on record (after one REQUEST-CHANGES→fix cycle); V1 grep=0 |
| Regression vs M0 | pass | single index route / no dead stubs untouched; M0 FE dirs not in diff |
| Regression vs M1 | pass | `MOCK_STATS`/`MOCK_ACTIVITY` greps = 0; dashboard surfaces untouched |
| Publish path / `v_cd_grantee` / 0243 / search untouched | pass | all four `git diff f357fb15^..HEAD` scopes empty |
| `go build` / `go vet` | pass | exit 0 |
| `go test ./...` | **flaky** | see C2/C6 — passes ~97% of full-suite runs, fails ~3% on `TestSeedRegistryParity_RegistryNotSeeded` |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Screen-DoD D2: no illustrative / `MOCK_`/`em breve` on Distribuição denominator | violated (whole screen illustrative) | **met** | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" …/DocumentDistributionPage.tsx` = 0; `MOCK_DISTRIBUTION` in `features/documents` = 0; `distributionMeta.ts` deleted at root |
| Denominator live (real producer) | none | **met** | F2.2 integration 9/9 live PASS; views serve real rows; FE 16/16 typed to generated contract |
| Numerator honesty (no fabrication) | n/a | **met** | `TrackingPendingNote.tsx` carries zero numbers; errored total renders `"—"`, never a fabricated 0; no `role`/read/ack field anywhere in the generated `Distribution*` types |

Root cause fixed, not symptom-patched: the scaffolding is **deleted at root**, the denominator is served by a real Grade-A endpoint, and the numerator gap is disclosed rather than faked.

- **Could it be built better?** The `TestSeedRegistryParity_RegistryNotSeeded` flake (C6) should be made deterministic — pick the "missing" cap from a **non-deferred** capability (filter `deferredCaps` before `caps[0]`, or sort + skip deferred). This is the minimum fix and becomes the fix-feature below. Secondary (defer, not blocking): the F2.3 V2/V3 runtime-with-live-API screenshot — judged acceptable (see C7).

## C6 — Forbidden-list

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean* (per-feature G/V criteria each mapped to a validator-run command).
- [ ] Fixture/mock passed off as real-provider proof — *clean* (F2.2/F2.1a/F2.1b proven on **live PG**, not fixtures; FE fixture tests labeled as such).
- [ ] Consumer contract guessed — *clean* (contract read from `DocumentDistributionPage` consumer; generated types only).
- [ ] Split-brain — *clean* (union defined once in `0245`; repo projects, never re-derives).
- [ ] Self-judged close / validator edited code — *clean* (validator wrote only this file).
- [ ] Scope drift — *clean* (only the 5 features; no numerator producer, no publish-path change, no action layer, no `v_cd_grantee` mod; aggregate diff confined to distribution + iam-cap + the two views + FE wire + ADRs).
- [x] **Symptom-patch / suite-green-as-pass — TRIGGERED (qualified):** the milestone gate (validation-def #2) requires "all 6 CI guards green" and the full `go test ./...` green. The `scripts/api-lint` guard is **non-deterministic**: F2.1c introduced the program's **first-ever** `deferredCap` (`deferredCaps` was empty `{}` at `f357fb15^`), which newly exposes a latent test-design bug — `TestSeedRegistryParity_RegistryNotSeeded` picks `caps[0]` from a randomized map without excluding deferred caps and asserts 1 violation, so ~1/30 of full-suite runs it fails `want 1, got 0`. The F2.1c/F2.2 evidence reported `go test ./...` exit 0 truthfully **but on a lucky branch** — the suite is not deterministically green. The **product code is correct** (the parity checker rightly exempts deferred caps); the defect is in a CI-gating **test's determinism**, introduced by an M2 change.

This is recorded as a real finding, not waved through. A flaky CI-gating guard violates the milestone's own "6 CI guards green" requirement and the C2 "flaky is not green" rule.

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed check(s):** C2 (named full-suite gate fails on isolated/repeat re-run — flaky, not green), C4 (BE workflow-class regression: a CI guard is non-deterministic), C6 (the milestone's "6 CI guards green / `go test ./...` green" bar is not met deterministically).
- **Scope of the failure is narrow and the milestone's substance is sound:** C1, C3, C5 all PASS; every per-feature acceptance (F2.1a/b/c, F2.2 G1–G11 live, F2.3 V1–V10) is met with real, validator-re-run, live-PG evidence; the product diff is staff-engineer clean with no split-brain, no scope drift, no fabrication, and the corrected tier-1 authz is honored. The **sole** blocker is one M2-introduced flaky CI-gating test.
- **Minimum fix feature to open:** **`f2.4-deferredcap-parity-test-determinism`** — make `TestSeedRegistryParity_RegistryNotSeeded` (and any sibling that indexes `AllCapabilities()`) deterministic by selecting the "missing-from-seed" capability from the **non-deferred** set (filter `deferredCaps` before choosing, or sort the slice and skip deferred caps) so the `~1/30` map-iteration flake cannot fire. Acceptance: `go test ./scripts/api-lint/... -count=200` green; `go test ./...` green across ≥10 consecutive full-suite runs; no product/contract/migration/authz change. (Pure test-determinism fix; no source-behavior change.)
- Milestone stays **active**; the main session does **not** advance. Per HS-4: open the fix feature, run its lifecycle, re-dispatch the validator.

### On the F2.3 V2/V3 runtime-screenshot bounded gap (explicit call)

**Acceptable for milestone close — not a blocker.** The producer is proven live end-to-end (F2.2 9/9 against real PG, validator-re-run), the FE consumes the **generated** contract types (tsc=0), the denominator logic is covered by 16/16 typed-fixture tests, and both FE reviewers APPROVE on record. The missing artifact is a screenshot, not a missing capability; the operator HS-1 review can spot-check at runtime. This gap does **not** contribute to the FAIL.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — blocked by this FAIL.
> - Status flipped in `README.md`: no (only on PASS).
