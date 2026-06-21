# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-21  ·  **Verdict:** see C7 — **PASS**.
> Milestone: Category C — published active-membership view (`metaldocs.v_active_user_areas`) + 3 consumers.
> Commits under judgment: F3.1 `fe181f34`, F3.2 `764a9d08`, F3.3 `c1b654d6` (branch `main`, local only).
> Environment: PG :5434 **OPEN** — all integration parity gates re-run live (no HS-3 not-run steps).

## Inputs loaded (all present, all readable)

milestone spec; F3.1/F3.2/F3.3 `spec.md`+`plan.md`+`evidence.md`; program README; mission governing spec
(via README + ADR-0039/0037 refs); migration `db/migrations/0242_iam_v_active_user_areas_view.sql`; the
aggregate diff `fe181f34~1..c1b654d6`. No input missing.

## C1 — Spec & plan conformance (per feature)

Each feature has `spec.md` (approval line filled: 2026-06-21 / leandrotca), a populated Interview record
(consumer-derived, not guessed — fail-closed honored), an execution-shaped `plan.md` (files/order/test
strategy, not a re-spec), and an `evidence.md` acceptance table matching the spec Validation Gate row-for-row.
Consumer contract honored: the view shape (`tenant_id, user_id, area_code, role`; `effective_to IS NULL`)
was **read from the three consumer call sites** — `role` exists because C3 (`ResolveEligibleActors`) filters
by it. Producer matches consumer, not the reverse. Non-goals respected (no base-table / passthrough-view
change; no temporal columns exposed; no consumer touched in F3.1; CD-owned grant legs untouched).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F3.1 (view producer) | ✅ view = consumer-derived union (`role` for C3); `effective_to IS NULL` only | ✅ all 5 gate rows | ✅ no consumer repoint; base table + passthrough untouched | `0242_*.sql`; `TestActiveUserAreasView_ParityWithBaseActiveNow` GREEN (re-run) |
| F3.2 (CD C1+C2) | ✅ both legs read `metaldocs.v_active_user_areas`; CD-owned legs unchanged | ✅ all 7 gate rows | ✅ set-based EXISTS preserved; signatures/scan untouched | `TestCanRead_/TestList_ViewParityWithRaw` GREEN (re-run) |
| F3.3 (approval C3, H-PRE-1) | ✅ in-tx SELECT reads view, `role` filter kept, no temporal predicate | ✅ all 8 gate rows | ✅ signature/`db.Tx`/scan/never-nil contract untouched | `TestResolveEligibleActors_ViewParityWithRaw` GREEN (re-run) |

## C2 — Gates re-run, isolated (validator-run from clean state, `-count=1`, not trusted from transcript)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F3.1 | `go test -tags integration -count=1 -run ActiveUserAreasView ./internal/modules/iam/infrastructure/postgres/` | `ok …/iam/infrastructure/postgres 3.792s` | ✅ |
| F3.2 | `go test -tags integration -count=1 -run 'TestCanRead_ViewParityWithRaw\|TestList_ViewParityWithRaw' ./…/controlleddocuments/infrastructure/` | `--- PASS: TestCanRead_ViewParityWithRaw` / `--- PASS: TestList_ViewParityWithRaw`; `ok 3.427s` | ✅ |
| F3.3 | `go test -tags integration -count=1 -run TestResolveEligibleActors_ViewParityWithRaw ./…/approval/repository/` | `--- PASS: TestResolveEligibleActors_ViewParityWithRaw`; `ok 3.391s` | ✅ |
| build | `go build ./...` | exit 0 | ✅ |
| guard | `go run ./tools/cilint ./...` | exit 0 | ✅ |
| cilint suite | `go test -count=1 ./tools/cilint/...` | `ok …/internal/analyzers 1.973s` | ✅ |

**Parity tests are genuine, not vacuous.** Each carries a **verbatim inline copy of the deleted raw SQL** as
its baseline (`rawCanRead`/`rawListIDs` with `FROM user_process_areas … effective_to IS NULL`; `rawEligible`
with the full interval form `effective_from <= now() AND (effective_to IS NULL OR effective_to > now())`),
seeds a **revoked-membership** discriminator (past `effective_to`, `revoked_by` set), and asserts set-equality
repo(view) == raw(baseline). F3.3 runs on a real `*sql.Tx` (`db.BeginTx`) with revoked/wrong-role/wrong-area/
wrong-tenant discriminators — empirically locking the ADR-0037 Model-A set-equality claim.

## C3 — Senior review of the aggregate milestone diff

Scope is exactly the 8 declared files (`git diff --stat fe181f34~1..c1b654d6`): 1 migration, 2 consumer
production edits, 3 parity tests, 2 cilint (ledger + fixture). Findings:

- **No split-brain.** The active-now predicate now lives in exactly one place — `v_active_user_areas`
  (`WHERE effective_to IS NULL`). Base table `public.user_process_areas` and the 1:1 passthrough
  `metaldocs.user_process_areas` are untouched. The three consumers re-derive no temporal predicate (the only
  `effective_to`/`effective_from` strings left in the two consumer production files are explanatory comments).
- **No N+1.** Both CD legs keep the membership `EXISTS` subquery (now over the view); C3 stays one set-based
  SELECT. No per-row Go membership loop introduced.
- **No H-PRE-1 violation (C3).** The `ResolveEligibleActors` diff is a single relation-token swap + dropped
  temporal predicate inside the *same* `tx.QueryContext` on the caller's `db.Tx` — no authz-recording read,
  no lock, no extra round-trip, no tx-structure change. View is `SELECT`-only (no `security_invoker`).
- **No dead code, no one-feature-breaks-another.** Migration is forward-only (0242 = next after 0241, unique),
  transactional (BEGIN/COMMIT), idempotent (`ON CONFLICT DO NOTHING`), with a `schema_migrations` row and an
  ADR-cited `COMMENT ON VIEW`.
- Staff-engineer bar: **met.** ✅

## C4 — Workflow-class QA (persistence/migration + module-boundaries + authz-invariant) + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Persistence/migration checklist | pass | forward-only, transactional, idempotent, `schema_migrations` row, view DDL encodes exactly `effective_to IS NULL`; bootstrap applies clean (F3.1 testdb row + view parity GREEN) |
| Module-boundaries (DDD ownership) | pass | iam owns + publishes the view (migration in `db/migrations/`); CD + approval read the published contract, name no iam base table |
| Authz-invariant (H-PRE-1 / ADR-0022) for C3 | pass | in-tx non-recording SELECT preserved; one-token swap |
| Test-framework discipline (ADR-0034) | pass | parity tests use canonical `testdb` fixture framework, `-tags integration`, PG :5434 |
| Regression vs M0/M1/M2 | all still pass | ADR-0039 intact; cilint guard exit 0 + suite green; M2's 9 B/N1 ledger entries remain drained (ledger comments B1–B8/N1 ported; only the 5 C4/search rows remain live); no C4 row touched |
| CD integration package | pass-with-defer | only `TestSequenceAllocatorNextAndIncrement_Concurrent` FAILs (`metaldocs.document_profiles` absent — raw-DSN defer); references no `List`/`CanRead`/membership (grep 0); file untouched by M3 |
| Approval integration package | pass-with-defer | only `TestPostgresLimiter_Live` (`public.auth_failure_counters` absent) + `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` (`metaldocs.tenants` absent) FAIL — raw-DSN defers; reference no `ResolveEligibleActors`/view (grep 0); files untouched by M3 |

**Pre-existing-defer independently confirmed:** the three failing tests (a) reference none of the M3-touched
methods, and (b) `git diff --name-only fe181f34~1..c1b654d6` shows their files were not modified by any M3
commit. They fail because they connect directly to the raw base DSN without the testdb bootstrap schema —
orthogonal to the membership-read seam. Not M3 regressions.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Category-C `user_process_areas` reads in `hgPendingRemediation` | 2 entries (C1+C2, C3) | **0** | `hgcrossmodule.go` ledger now holds only the 5 C4/search rows; CD + approval entries drained with ported notes |
| Raw `user_process_areas` production reads in consumers | 3 (CD List/CanRead, approval) | **0** | `git grep -n user_process_areas -- …controlleddocuments/ …documents/approval/ ':!*_test.go'` → CD: none; approval: one **comment** (`route.go:150`). Production reads gone. |
| C3 Model-B interval leak (`effective_to > now()`) | present | retired | replaced by `effective_to IS NULL` view; set-equality parity-proven incl. revoked row |

Root cause fixed, not symptom-patched: each consumer was repointed to the **iam-published view** (the single
home for the active-membership predicate), **not** to a consumer-local copy of the query or a parallel
passthrough view. The C3 predicate change is justified-and-parity-proven (set-equality under Model A on a real
`*sql.Tx`), not a silent visibility change.

- Could it be built better? No material improvement for M3. The published-view mechanism (ADR-0039 D3a) is the
  correct construction for set-based in-SQL membership predicates. (C4/search is the deliberately-deferred
  redesign — M4 — and is untouched here.)

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (per-feature named tests re-run + mapped)*
- [ ] Fixture/mock passed off as real-provider proof — *clean (all parity GREEN on real PG :5434; verbatim raw baselines, real `*sql.Tx`)*
- [ ] Consumer contract guessed rather than read from the consumer — *clean (view shape derived from the 3 call sites; `role` for C3)*
- [ ] Split-brain (one fact, two sources of truth) — *clean (active-now predicate lives only in the view)*
- [ ] Self-judged close / validator edited or fixed code — *clean (validator only judged; wrote only this file)*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean (exactly the 8 declared files)*
- [ ] Symptom-patch (bar moved by masking, root cause intact) — *clean (consumers read the published view; root cause closed)*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. Code-wise: senior-level, contract-clean, no split-brain, no dead code, no guessed
  contract, H-PRE-1 preserved. Function-wise/QA: all three named parity gates re-run GREEN from clean state on
  real PG :5434 with verbatim raw baselines + revoked-membership discriminators proving zero authz/visibility
  drift; Category-C class re-measured to 0 ledger entries and 0 production reads; build + guard + cilint suite
  green; M0–M2 regression intact; the three integration failures are independently-confirmed pre-existing
  raw-DSN defers unrelated to the membership-read change.
- Handed back to the main session to flip status (README M3 → passed) and present the **HS-1 operator gate**.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS, by the main session
