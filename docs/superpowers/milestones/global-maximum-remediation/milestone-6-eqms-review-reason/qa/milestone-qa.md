# Milestone 6 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding, HS-7) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run 2:** 2026-07-05 · **HEAD:** `b8abea1d` · **Base (M6 start):** `93cd6114` · **Verdict:** see C7 — **PASS**.
> Environment: `DATABASE_URL`/`METALDOCS_DATABASE_URL` **unset** in this validator shell (`.env` sourcing forbidden, CLAUDE.md). Deterministic gates re-run here from clean; the `-tags integration` suite is verified **structurally** (tests exist as real testdb-factory integration tests, code-under-test matches §4, evidence table internally consistent) — the main session executed them on real Postgres under F6.4 gate#7.

---

## Run 1 → Run 2 delta (FAIL → PASS audit trail)

**Run 1 (`fc3057d4`, 2026-07-05) returned FAIL** on C1/C3/C6 with 2 blocking + 1 non-blocking finding on the F6.2 River surfacer, and directed HS-4 fix-feature `f6.4-surfacer-contract-and-consumer` with 4 fix-requirements. **Run 2 confirms all four are cleared:**

| Run-1 fix-requirement | Run-2 disposition | Evidence (verified in source, this run) |
|---|---|---|
| **1. Resolve §4 divergence** — surfacer shipped an unseeded all-tenant sweep + a test asserting the **inverse** of §4.3 isolation, no HS-7 erratum | **CLEARED — conformed to contract (no re-open)** | `job.go run()` is now cross-tenant **read** (`ListTenantsWithDueReviews`, bypass, unseeded — safe read) → **per-tenant** loop: `BypassSystem`+`SeedTxTenant(tenantID)` → `ListDueForReview`+`MarkSurfaced`, both carrying explicit `tenant_id = $N::uuid`. No unseeded tenant-scoped write survives. The false "mirrors watchdog sweep" comment is gone. Isolation test now split: `..._FullTick_IteratesAllTenants` (end-state, explicitly NOT the isolation proof) + `..._Writer_TenantSeed_DoesNotSurfaceOtherTenant` (real §4.3: seed A only → A surfaced, **B untouched**). Asserts the contract, not its inverse. |
| **2. Wire `review_surfaced_at` to a real consumer** | **CLEARED — worklist consumer** | `repository.go:485` `ReviewDue` branch now reads `review_surfaced_at IS NOT NULL AND review_surfaced_at >= review_due_at` (was a recompute of `review_due_at <= now()`). Exposed on `DocumentSummary`+`DocumentDetailResponse` DTOs (openapi + regen BE `api.gen.go` + FE `index.d.ts`), scanned in `GetDocument`/`ListDocumentsPaginated`. mark-reviewed advancing `review_due_at` auto-expels. |
| **3. Single-source the triple-authored due predicate** | **CLEARED (reader/writer layer)** | `const dueCorePredicate` (`review_due_reader.go:40`) referenced by `ListDueForReview`, `ListTenantsWithDueReviews`, and `MarkSurfaced`. The list-filter site restates the effective-window fragment (documented: `dueCorePredicate` is a fixed-`$1` string that cannot splice into the arg-numbered filter builder) but derives from the same definition — an acceptable, documented derivation, not a silent second source. |
| **4. Execute the authored-not-executed integration suite on real Postgres** | **CLEARED (main session, F6.4 gate#7)** | F6.4 evidence table lists every M6 DB proof `--- PASS` with real timings (surfacer isolation/idempotency/re-surface, ListDueForReview isolation/filter, ReviewDueFilter_ReadsSurfaced, CHECK constraints, mark-reviewed OCC/isolation/authz, tripwire-negative P0001, reason-persist, reason-on-audit, schedule-publish). Validator confirmed structural coherence (tests exist, are testdb-factory integration tests, compile under `-tags integration`, code matches §4). |

The mission's HS-7 discipline was honored: the divergence was **fixed to the contract** (the feasible path), not silently ratified — no contract re-open, no erratum needed.

---

## Inputs loaded (none missing)

milestone.md · validation-contract.md (§0–§8) · f6.2 spec/plan/evidence · f6.3 evidence · f6.4 spec/plan/evidence · F6.1 gate (`docs/superpowers/analysis/2026-07-04-m6-eqms-review-reason-system-impact.md`, Yellow, 93cd6114) · program README/mission · aggregate diff `93cd6114..b8abea1d` (+ F6.4 sub-range `fc3057d4..b8abea1d`, 29 files, +1226/−344).

---

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F6.1 gate | ✅ | ✅ | ✅ | Gate committed 93cd6114, 🟡 Yellow, HS-8 did not fire; 9 constraints carried into spec + contract |
| F6.2 | ✅ **(via F6.4 conform)** | ✅ | ✅ | §4 surfacer now conforms (per-tenant seed + explicit predicate + real isolation proof); evidence rows flipped from authored-not-executed → executed-real-DB PASS |
| F6.3 | ✅ | ✅ | ✅ | reason-for-change structured (not `revision_title`); reason-persist + reason-on-`governance_events` executed real DB PASS |
| F6.4 (HS-4 fix) | ✅ | ✅ | ✅ | spec approved 2026-07-05/Leandro, 3-row interview populated; plan execution-shaped (Task A/B/C + ordering + defers); evidence acceptance table (gates 1–8) all met |

Artifact hygiene: every feature has `spec.md` (approval line date+operator), populated interview record, execution-shaped `plan.md`, and an `evidence.md` acceptance table matching the spec Validation Gate. **No missing artifacts.** The run-1 substantive C1 failure (F6.2's §4 consumer contract not honored) is resolved: the producer (`MarkSurfaced`/`ListDueForReview`) now matches the binding §4.2/§4.3 consumer contract, verified by reading the surfacer job + the two ports + the isolation tests.

## C2 — Gates re-run, isolated (from clean state, HEAD b8abea1d)

| Gate | Command re-run (this validator) | Real output | Pass? |
|------|----------------------------------|-------------|-------|
| Build | `go build ./...` | exit 0, no output | ✅ |
| Vet | `go vet ./...` | exit 0 | ✅ |
| Vet (integration tag) | `go vet -tags integration ./…/document_review_surfacer/… ./…/repository/… ./…/approval/application/… ./tests/integration/documents/…` | exit 0 — authored integration tests **compile** (real code, not stubs) | ✅ |
| Registry 35 + classify | `go test -count=1 -run 'TestCapabilityRegistrySize\|TestEveryCapabilityClassified' ./internal/modules/iam/domain/...` | `ok … 1.128s`; `model_test.go:96 const want = 35` | ✅ |
| Tripwire golden + arm parity | `go test -count=1 ./internal/platform/tripwire/...` | `ok … 1.099s` | ✅ |
| DTO wire-contract pin | `go test -count=1 -run TestDocumentSummaryAndDetail_ReviewFieldsWireContract ./internal/modules/documents/delivery/http/...` | `ok … 3.438s` (includes new `review_surfaced_at`) | ✅ |
| documents + jobs unit suites (regression) | `go test -count=1 ./internal/modules/documents/... ./internal/modules/jobs/...` | exit 0 (no failures) | ✅ |
| api-lint strict | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | **exactly 2** violations: `SEED-CHOKEPOINT-ALLOWLIST-STALE cancel_service.go:76`, `ASYNC-TENANT-SEED fanout_worker.go:98` | ⚠️ pre-existing (below) |
| Integration (`-tags integration`) | not executed here (no DSN); **structurally verified** | tests exist, testdb-factory, compile, code matches §4; main session ran them (F6.4 gate#7) — every row `--- PASS` | ✅ (structural) |

**Pre-existing-violation verification (not on trust).** `git diff 93cd6114..HEAD` for `cancel_service.go` and `fanout_worker.go` is **empty** — both untouched by M6. The only M6 edit to `seed-chokepoint-allowlist.txt` in the F6.4 range is a mechanical **+18 line-number resync** of the same `repository.go` `raw BeginTx` entries (offsets shifted by F6.4's code additions) — **no new suppression, no new tenant-scoped write masked**. Zero F6.4-introduced violations. Correctly M3-deferred with triggers.

## C3 — Senior review of the aggregate milestone diff (`93cd6114..b8abea1d`)

Run-1 findings re-checked against source at HEAD:

- **Finding 1 (was blocking — §4 silent divergence): RESOLVED.** `job.go run()` no longer does an unseeded all-tenant tenant-scoped write. It reads tenants cross-tenant (safe, unseeded), then writes **per-tenant under `SeedTxTenant`**, and both `ListDueForReview`+`MarkSurfaced` carry an explicit `tenant_id = $N::uuid` predicate (correct-by-construction isolation; RLS is backstop). The isolation test asserts §4.3 (seed A → B untouched), not its inverse. No HS-7 erratum was needed because the code was conformed to the contract.
- **Finding 2 (was blocking — unwired marker): RESOLVED.** `review_surfaced_at` now has a real consumer — the `review_due=true` worklist filter reads it, and it is exposed on both document DTOs (contract-first, regen). Not a write-only column.
- **Finding 3 (was non-blocking — triple-authored predicate): RESOLVED at the reader/writer layer.** One `dueCorePredicate` const shared by the 3 SQL sites; the filter-builder derivation is documented and single-definition, not a divergent source.
- **F6.4 diff quality:** scoped and coherent — surfacer conform (job.go + ports), worklist filter + DTO (repository/handler/model/openapi/api.gen.go/FE), predicate single-source, tests, docs. `api.gen.go` churn is genuine oapi-codegen regen from the DTO addition (contract-first). No new capability, route, or migration family (matches non-goals). Real-DB fixes bundled in `bf9eadaf` (NULLABLE `code` scan, TEXT `resource_id` cast, bare-ctx identity seed) are test-setup/scan corrections diagnosed to root cause — **no production code touched to make a test pass**.
- **Working tree:** no uncommitted source changes (the run-1 `typed_response_test.go` modification is committed within the F6.4 range).

**Staff-engineer bar met?** ✅ — the headline F6.2 surfacer is now both built to its binding contract and consumed.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Contract/authz/DB-invariant/async subsets | pass | Deterministic gates green here; DB-invariant + async idempotency proofs executed on real Postgres (F6.4 gate#7), structurally verified this run |
| Regression vs prior milestones (M0–M5) | **all still pass** | M2 registry 35 + tripwire green with `document.review`; documents/jobs unit suites green; api-lint 2 violations identical & untouched at base; build/vet green (incl. `-tags integration` compile) |
| Separation-of-powers self-check | ✅ | Main session committed the F6.4 fix (`7d398d92`…`b8abea1d`); this validator only judged + wrote this file. No source edited. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before (run 1) | After (run 2) | Root-cause-fixed evidence |
|-------------|----------------|---------------|---------------------------|
| effective/expiry reuse (no dup column family) | ✅ | ✅ | migration 0274; §2 honored |
| reason-for-change structured (not `revision_title`) | ✅ | ✅ | reason-persist + reason-on-audit PASS (real DB) |
| capability via M2-generated arm | ✅ | ✅ | registry 35, tripwire drift+golden green; P0001 negative PASS |
| **periodic review surfaced (River, no hand-rolled scheduler)** | ❌ (§4 divergence + inert side effect) | ✅ | per-tenant seed + explicit predicate + real §4.3 isolation proof; marker consumed by worklist filter |
| mark-reviewed via M4 unified fn | ⚠️ intentionally NOT routed through `CanTransitionDocumentStatus` | ⚠️ unchanged | published→published is not a status transition; F6.2 interview #4 refined this pre-code — defensible, documented refinement (not a silent drift). Milestone-obj-#3 "routed through M4" wording is met-in-spirit (friendly first-line DB-CHECK mirror + OCC CAS); recorded as retrospective input, not a fail. |

- **Could it be built better?** The ASYNC-TENANT-SEED lint has a documented **port-indirection blind spot** (it can't see a tenant-scoped write behind a cross-module port — exactly the gap the shipped surfacer exploited). F6.4 correctly records this as a **bounded defer with a trigger** (lint-hardening / next tenancy sweep) — the human §4.2 backstop caught it this time; machine-checking it is the durable fix. This is next-milestone input, not an M6 blocker.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *(clean: every acceptance row maps to a named test/proof)*
- [ ] Fixture/mock passed off as real-provider proof — *(clean: evidence splits live-HTTP vs testdb integration honestly; validator did not run integration here and says so)*
- [ ] Consumer contract guessed rather than read from the consumer — *(clean: §4 consumer contract now honored, verified in source)*
- [ ] Split-brain (one fact, two sources of truth) — *(clean: `dueCorePredicate` single-sourced; filter derivation documented)*
- [ ] Self-judged close / validator edited or fixed code — *(clean: main session built F6.4; validator only judged)*
- [ ] Scope drift (work beyond the spec, no rationale) — *(clean: F6.4 within its non-goals; no new cap/route/migration family)*
- [ ] Symptom-patch (bar moved by masking) — *(clean: surfacer conformed to contract at root; real-DB fixes root-caused)*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- All run-1 fix-requirements (1–4) are cleared: the F6.2 River surfacer conforms to the binding validation-contract §4.2/§4.3 (per-tenant `SeedTxTenant` + explicit `tenant_id` predicate + a real cross-tenant isolation proof that asserts the contract, not its inverse); `review_surfaced_at` is consumed by the `review_due=true` worklist filter and exposed on both document DTOs; the due-core predicate is single-sourced; and the previously authored-not-executed integration suite was executed on real Postgres (F6.4 gate#7) with every M6 DB proof `--- PASS`. Deterministic gates re-run by this validator from clean state are green (build, vet, registry 35, tripwire, DTO pin, unit suites); api-lint shows exactly the 2 pre-existing, untouched, M3-deferred violations with zero F6.4-introduced. No forbidden-list hit. Both dimensions (code-wise and function-wise) pass.
- Milestone → handed back to the main session to flip status and present the **HS-1 operator gate**.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — present to operator on this PASS.
> - Status flipped in `README.md` / milestone.md: only on PASS + after HS-1 (main session).
> - Never push (mission §2, §10).
