# Milestone 2b — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-07  ·  **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file only; it did not edit source, fix findings, or flip status.
> The **main session flips status only on this PASS** (then the HS-1 operator gate).

## Inputs loaded

- Milestone spec `milestone.md` (incl. the 7 locked runtime-truth corrections) — read.
- All 10 features' `spec.md` / `plan.md` / `evidence.md` — read.
- Program `README.md`, governing spec (`2026-07-07-approval-remediation-design.md`, ratified
  e4a0717a/046f0633/68a0b3b8), plan, and Yellow system-impact analysis (locked constraints) — read.
- Aggregate milestone diff `b0f3c81d^..HEAD` (9 commits, 133 files, +13295/-662) — reviewed.

No input missing or unreadable — validation proceeded (did not fail-fast).

## C1 — Spec & plan conformance (per feature)

All 10 features have `spec.md` + `plan.md` + `evidence.md`. Approval-before-code present for every
feature (F1–F3 "Approved before code: 2026-07-07 — operator via ratified governing spec"; F4–F9
"Status: Approved for implementation"; F10 references the ratified spec). Governing spec ratified
before any feature commit. Interviews populated; consumer contracts read (not guessed) — spot-verified
F6 (`LoadFrozenContentHash` matches the signoff/publish call sites), F7 (`signature_meaning` DTO
matches the live hand-maintained `contracts.SignoffRecord`, not the dead generated type), F8
(`requireInstanceVisible` area threaded from the instance's own resolved area).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F1 stage-kind schema | ✅ | ✅ (live: `TestStageKindSchemaExpand_*` PASS 234s/69s, real DB) | ✅ | expand-only 0286; DB CHECK live-verified |
| F2 route versioning | ✅ | ✅ (live: 4× `TestRouteVersioning_*` PASS, real DB) | ✅ | 0287 partial-unique + immutability tripwire live-verified |
| F3 caps review/oversee | ✅ | ✅ (registry +2 → 40; lints green; prefix-fallback grep-zero) | ✅ | tier-1 explicit rows; ADR 0075 |
| F4 review verdicts | ✅ | ✅ code+unit / integration SKIP (creds) | ✅ | cancel_reason overload fix confirmed in code (see C2) |
| F5 freeze boundary | ✅ | ✅ code+unit / integration SKIP; markup gate unwired (bounded, ADR 0076 §4) | ✅ | comment-gate + hash-pin + CAS wired; ADR 0076 |
| F6 no-fallback chain | ✅ | ✅ unit + live-HTTP 412 fail-closed | ✅ | two-branch (no COALESCE) confirmed both call sites |
| F7 sig-meaning + SoD | ✅ | ✅ (reject-meaning defect fix confirmed in code + unit) | ✅ | single `CheckSoD` predicate, both sites |
| F8 SLA/visibility/worklist | ✅ | ✅ (visibility at SQL layer; oversee 403/200 live) | ✅ | area-scope bypass fix confirmed in code |
| F9 delegation | ✅ | ✅ code+unit / integration SKIP | ✅ | 0293 tenant_id + RLS + no-self/window CHECKs; ADR 0077 |
| F10 ADRs/wiki/live-QA | ✅ | ✅ (4 ADRs indexed; 3 live bugs fixed+verified) | ✅ | see C4/C5 |

**C1 finding (non-blocking):** F10 §1 names ADR 0075 as `0075-approval-review-oversee-capabilities.md`;
the actual file is `0075-approval-oversee-visibility.md` (present, Accepted, indexed at index.md:75).
Cosmetic evidence-text misnomer — the ADR itself is correct and complete.

## C2 — Gates re-run, isolated (validator, clean state)

| Gate | Command | Real output | Pass? |
|------|---------|-------------|-------|
| Build | `go build ./...` | exit 0, no output | ✅ |
| Integration build | `go build -tags integration ./...` | exit 0 | ✅ |
| Full suite | `go test ./...` | 96 `ok`, **0 FAIL** | ✅ |
| Module re-run | `go test -count=1 ./internal/modules/documents/approval/... ./internal/modules/iam/domain/... ./internal/modules/jobs/... ./scripts/api-lint/...` | all `ok` (13 pkgs) | ✅ |
| api-lint strict | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` | ✅ |
| Registry size | `go test ./internal/modules/iam/domain/... -run TestCapabilityRegistrySize -v` | PASS, `want=40` (test comment: "M2b F3: +2 approval.review/approval.oversee") | ✅ |
| Cap scopes | grep `capability_scope.go` | `CapApprovalReview: ScopeArea`, `CapApprovalOversee: ScopeTenant` (matches ADR 0075) | ✅ |
| Grep-zero SkipStage | `grep -rn 'SkipStage\|ErrCannotSkipLastStage' --include=*.go .` | 0 hits | ✅ |
| Grep-zero prefix fallback | `permissions.go` `/approval/` scan | 0 catch-all rows; only explicit method+prefix+suffix+cap rows + deletion comments | ✅ |
| Grep-zero forbidden COALESCE | F6 call sites | `LoadFrozenContentHash` = bare `SELECT frozen_content_hash` (no COALESCE); `active_instance_reader.go` = explicit two-step branch; only remaining COALESCE is single-column null-tolerance in `LoadHeadContentHash` (the sanctioned draft/no-instance branch, correction #2) | ✅ |

**High-risk fixes independently re-verified present in current code (not just described):**
- **F4 cancel_reason overload:** `UpdateInstanceStatusWithReason` is called ONLY from
  `cancel_service.go:104`; the `request_changes` path (`review_verdict_service.go:320-321`) uses plain
  `UpdateInstanceStatus` with an inline comment documenting the fix. ✅
- **F7 signature-meaning defect:** `decision_service.go:327-330` derives `signatureMeaning` from
  `req.Decision` (`reject → "rejection"`), passed to `NewSignoff` — no silent "approval" default. ✅
- **F8 area-scope visibility bypass:** `requireInstanceVisible` (read_service.go:695) checks
  `authz.Require(CapDocumentEdit, areaCode)` with the instance's real resolved area (via
  `loadInstanceAreaCode`), not the `"tenant"` filter-skip sentinel; enforcement is inside the tx at
  the SQL/repo layer — worklist `eligibilityPredicate` is a literal WHERE fragment, no client-side
  filter path. ✅
- **F10 3 live bugs:** Bug#1 constraint name `approval_routes_active_profile_uq` (errors.go:91); Bug#2
  `SAVEPOINT route_update_attempt` + `ROLLBACK TO SAVEPOINT` (route_admin_service.go:401); Bug#3
  `($2::jsonb IS NOT NULL)` tautology replacing bare `"TRUE"` (read_service.go:440,573). All ✅.

**Integration tests:** the M2b `//go:build integration` suites (`TestReviewVerdict_*`, `TestFreeze_*`,
`TestRouteVersioning_*`, SLA, SoD-trigger) compile clean and **SKIP** on my re-run — no
`DATABASE_URL`/`METALDOCS_DATABASE_URL` in the validator environment, obtainable only by reading
`.env` (forbidden by the standing rule that binds the validator too). Local Postgres is up on :5433
but the DSN is credential-gated. This is the identical, legitimate wall every feature documented — a
SKIP (compile-clean), never a FAIL. F1/F2 evidence shows credible live runs (real-DB durations
234s/158s) for the two schema-critical migrations, and F10 live-HTTP QA covered the
route/visibility/no-fallback/filter paths. No gate is flaky or environment-coupled to FAIL.

## C3 — Senior review of the aggregate milestone diff

- **Scope:** every touched path is inside the declared boundary — `documents/approval`, `iam/domain`,
  `jobs` (approval_sla_surfacer sibling + maintenance registration), `apps/api`/`apps/jobs` wiring,
  migrations, `api/openapi` + regen, api-lint config/allowlist, tests, wiki/docs. **No scope drift.**
- **Split-brain:** none found. F6 no-fallback is a single source of truth (frozen pin), the forbidden
  polymorphic COALESCE is deleted at both call sites and replaced by explicit status-scoped branches.
  F7 collapses the dual SoD sites onto one `CheckSoD` predicate + one shared `enforce_approval_sod()`
  trigger. Delegation (0293) has one tenant/RLS-scoped table with DB CHECKs (`ends_at>starts_at`,
  `delegator_id<>delegate_id`) mirrored by the app layer.
- **Contract-first:** api-lint strict `0 violations`; new routes landed via openapi + oapi-codegen.
- **DB-enforced invariants:** stage_kind CHECK (0286, live-verified reject), route immutability P0001
  tripwire (0287, live-verified), delegation CHECKs (0293), SoD trigger widened symmetrically (0290).
- **Findings (non-blocking):**
  1. **F6 dead parity test:** F6 (commit 1458ee7d) added parity sub-tests to
     `tests/integration/controlleddocuments/active_instance_parity_test.go`, which was **already
     uncompilable before M2b** (stale `metaldocs/internal/modules/documents/repository` import from
     the ADR-0073 `repository→infrastructure` rename — confirmed present at `b0f3c81d^`). Those
     sub-tests are therefore dead-on-arrival. F6's *substantive* no-fallback proof lives in compilable,
     passing unit tests (`TestRecordSignoff_NullFrozenHash_FailsClosed`,
     `TestPublish/SchedulePublish_NullFrozenHash_FailsClosed`) plus the live-HTTP 412, so this is a
     coverage-thinness nit, not a proof gap. Recorded for the F10-cleanup / ADR-0073 rename-debt task.
  2. **Migration numbering gaps** 0289/0291 are absent (actual set: 0286,0287,0288,0290,0292,0293).
     Each file self-numbers coherently; lexical apply order is intact; no duplicate/conflicting ledger
     rows. Cosmetic. (F10 §4's "0286-0293, 33/33" narrative overstates — 0289/0291 don't exist — but
     the applied set is internally consistent.)
- **Staff-engineer bar met?** ✅ — the diff is senior-level: root-cause structural fixes (versioned
  routes, freeze+no-fallback chain, capability tier coherence), honest defer discipline, no guessed
  contracts, no dead production code (the two nits are test-side/cosmetic).

## C4 — Workflow-class QA (backend-api) + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Contract↔handler parity | pass | api-lint strict 0 violations |
| AuthZ tier-1/tier-2 coherence | pass | explicit tier-1 rows per verb; `authz.Require` at every mutating site; tripwire lints green |
| Multi-tenant isolation (new tables) | pass | `approval_review_verdicts` (0288) + `approval_delegations` (0293) both `tenant_id NOT NULL` + RLS ENABLE/FORCE + tenant_isolation |
| Async/idempotency (surfacer) | pass | approval_sla_surfacer is a genuine sibling (grep-zero import of document_review_surfacer), alert-only ADR 0068, dual-define ADR 0067 |
| DB-invariant tripwires | pass | new CHECKs/triggers present; F1/F2 live-verified |
| Regression vs prior programs (GMR, module-boundary-hardening) | **all still pass** | `go test ./...` 96 ok / 0 FAIL |
| Pre-existing integration-compile breakage | pre-existing, NOT a regression | `tests/integration/{controlleddocuments,documents,templates}` `[setup failed]` on stale `*/repository` imports and `scenarios` `TestReflect_RepositoryNoBeginTx` stale walk path — **all four confirmed red at `b0f3c81d^` (pre-M2b)**, ADR-0073 rename debt inherited, not caused by M2b |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| W1 route permanent-freeze | route frozen after first use (contradicts ADR 0018) | versioned-immutable rows; in-use PUT supersedes, never destructive-updates | live-verified F2/F10 Bug#2: PUT on in-use route → v2 created, v1 `active=false`, P0001 on definition-column update — **structural, not a trigger-condition patch** |
| W2 floating hash pin | signoff hash COALESCEd to head revision (no production writer) | freeze pins `frozen_content_hash`; signoff echoes it; publish verifies it; NULL → fail-closed 412 | live-verified F6 412 on NULL pin; two-branch reads (no COALESCE) — **structural freeze+no-fallback chain, not another COALESCE branch** |
| AuthZ capabilities-not-roles coherence | generic `/approval/` tier-1 prefix fallback; tenant-sentinel reads | explicit tier-1 per verb; `approval.review`/`approval.oversee` caps; eligibility∪oversee∪edit visibility (404 across boundary) | grep-zero prefix fallback; registry +2; oversee 403/200 live-verified |

- Root cause fixed, not symptom-patched — confirmed for W1 and W2 (the two named substrate defects).
- **Could it be built better?** (a) `stage_kind` belongs on the route-create HTTP contract so review
  stages are creatable via API (currently domain/DB-only) — drives the F10 reachability gap; a real
  contract-first feature for M2c or a follow-up. (b) The ADR-0073 `*/repository` rename left 4
  integration packages uncompilable tree-wide — a dedicated rename-debt sweep would restore that
  regression surface (independent of M2b). (c) `approval_review_verdicts` has no tripwire arm (arm
  registry only checks registered tables) — bounded-deferred by F4. None make the current
  construction unsound; all are recorded as next-milestone / cleanup input.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clean**
  (each feature's acceptance row is mapped to a named test/grep/response, not a bare suite-green).
- [ ] Fixture/mock passed off as real-provider proof — **clean** (evidence explicitly distinguishes
  fixture/unit from live-DB and live-HTTP; integration SKIPs are labeled as deferred, not passed).
- [ ] Consumer contract guessed rather than read — **clean** (F6/F7/F8 contracts verified against real
  consumer sites).
- [ ] Split-brain — **clean**.
- [ ] Self-judged close / validator edited or fixed code — **clean** (validator only judged + wrote
  this file; no source edited).
- [ ] Scope drift — **clean** (all diff in-boundary).
- [ ] Symptom-patch — **clean** (W1/W2 structurally fixed).

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**

Both dimensions pass. **Code-wise:** senior-level, contract-clean, no split-brain, no dead production
code, no guessed contracts; the three history-flagged fixes (F4 cancel_reason overload, F7 reject-
meaning mislabel, F8 area-scope visibility bypass) and F10's three live-QA bug fixes are all
independently confirmed present in the current tree. **Function-wise/QA:** W1 (versioned routes) and
W2 (freeze + no-fallback hash chain) are live-verified as structural root-cause fixes; F1/F2's
schema-critical DB invariants have credible real-DB runs; F10 live-HTTP QA exercised the
route/visibility/no-fallback/worklist paths and caught+fixed three genuine production bugs. The
deferred live-DB integration runs are a consistent, legitimate, non-symptom-patch pattern — each has a
named rerun trigger + owner, they SKIP (compile-clean) rather than FAIL, and the same `.env` credential
wall binds the validator, so they cannot be forced open here. The one real coverage hole — the
review-verdict→freeze→signoff lifecycle has no live end-to-end proof because `stage_kind` is not yet on
the route-create HTTP contract — is honestly disclosed, appropriately bounded (the behavior is proven
at the unit + DB-CHECK + partial-live layers; the blocker is a pre-existing contract gap, not a
regression), and does not by itself unseat the milestone. Two non-blocking findings are recorded (F6
dead parity sub-tests on a pre-broken ADR-0073 file; migration numbering gaps 0289/0291) for the
next cleanup pass.

- **Fix features to open:** none required to clear the gate. Recorded as bounded follow-ups (not
  milestone-blocking): `f-followup-stage-kind-route-contract` (expose `stage_kind` on
  `StageRequest`, contract-first, unblocks the full review-kind live walkthrough — natural M2c
  dependency) and `f-followup-adr0073-integration-rename-debt` (repair the 4 tree-wide
  `*/repository` integration-compile failures + the F6 dead parity sub-tests).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — validator PASS triggers the operator review; no M2c start and no
>   push without explicit approval.
> - Status flipped in `README.md`: no — main session flips to `validated (pending HS-1)` on this PASS.
