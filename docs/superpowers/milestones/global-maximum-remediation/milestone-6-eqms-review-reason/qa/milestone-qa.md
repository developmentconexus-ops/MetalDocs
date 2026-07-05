# Milestone 6 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding, HS-7) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-05  ·  **HEAD:** `fc3057d4`  ·  **Base (M6 start):** `93cd6114`  ·  **Verdict:** see C7 — **FAIL**.
> Environment: `DATABASE_URL`/`METALDOCS_DATABASE_URL` **unset**; `.env` sourcing forbidden (CLAUDE.md).
> Integration proofs therefore ran as `t.Skip` — recorded honestly as **authored-not-executed**, not counted green.

## Inputs loaded (none missing)

milestone.md · validation-contract.md · f6.2 spec/plan/evidence · f6.3 spec/plan/evidence · F6.1 gate
(`docs/superpowers/analysis/2026-07-04-m6-eqms-review-reason-system-impact.md`, Yellow, committed 93cd6114) ·
program README · governing mission.md · aggregate diff `git diff 93cd6114..fc3057d4`.

---

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F6.1 gate | ✅ | ✅ | ✅ | Gate committed 93cd6114, verdict 🟡 Yellow, HS-8 did not fire; 9 constraints carried into spec + contract |
| F6.2 | ❌ **(§4 surfacer diverges silently)** | ⚠️ partial | ✅ | spec approved 2026-07-04/Leandro; interview 7-row; **but shipped surfacer contradicts contract §4.2/§4.3 with no HS-7 erratum — see finding 1** |
| F6.3 | ✅ | ✅ (HTTP legs) / authored-not-executed (DB legs) | ✅ | spec approved 2026-07-04/Leandro; interview 4-row; audit-trail claim honestly labeled — see C2 |

Artifact hygiene (all features): `spec.md` approval line filled (date+operator); interview records populated;
`plan.md` execution-shaped (task list, files, test strategy, ordering); `evidence.md` acceptance table present.
**No missing artifacts.** The C1 failure is substantive, not structural: F6.2's shipped surfacer does **not**
honor the binding §4 consumer contract, and the divergence was **not** recorded as an HS-7 erratum
(contrast M3/M4/M5, each of which recorded a loud dated erratum in the README hard-stops table when it
diverged from its D4 contract).

## C2 — Gates re-run, isolated (from clean state, HEAD fc3057d4)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| Build | `go build ./...` | exit 0, no output | ✅ |
| Vet | `go vet ./...` | exit 0 | ✅ |
| Registry size 35 + classify | `go test -count=1 -run 'TestCapabilityRegistrySize\|TestEveryCapabilityClassified' ./internal/modules/iam/domain/...` | both PASS; `model_test.go:96 const want = 35`; `CapDocumentReview="document.review"` ScopeTenant | ✅ |
| Tripwire golden + arm parity | `go test -count=1 ./internal/platform/tripwire/...` | `ok` 1.126s | ✅ |
| Read-side + approval contract + application | `go test -count=1 ./internal/modules/documents/delivery/... .../approval/http/contracts/... .../approval/application/...` | all `ok` | ✅ |
| Named pins | `TestDocumentSummaryAndDetail_ReviewFieldsWireContract`, `TestSubmitRequestReasonField`, `TestSubmitReasonRequiredRev1{,_DerivedFromDocumentRow…}`, `TestSubmitReasonOptionalAtRev0`, `TestSubmitPersistsReason*`, `TestSubmitReasonCategoryInvalidRejected` | all `--- PASS` | ✅ |
| api-lint strict | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | 2 violations: `SEED-CHOKEPOINT-ALLOWLIST-STALE cancel_service.go:76`, `ASYNC-TENANT-SEED fanout_worker.go:98` | ⚠️ pre-existing (see below) |
| **Integration** (`-tags integration`) | `TestDocumentReviewCheckConstraints`, `TestMarkReviewed_*`, `TestIntegration_Surfacer_*`, `TestSubmit*_RealDB`, `TestListDueForReview_*` | **`--- SKIP` — "DATABASE_URL/METALDOCS_DATABASE_URL not set"** | ⛔ authored-not-executed |

**Pre-existing-violation verification (not taken on trust).** Ran api-lint **at base `93cd6114`** in a
detached worktree: the **same 2 violations** appear identically. Source `cancel_service.go` /
`fanout_worker.go` are **untouched** by the M6 range (`git diff 93cd6114..HEAD` empty for both). The only
M6 edit to `seed-chokepoint-allowlist.txt` is a mechanical line-number re-sync of unrelated
`repository.go` offsets (F6.2 T6 shifted them); the stale `cancel_service.go:76` line is **byte-identical**
at base and head. The `scripts/api-lint` package self-test (`TestAsyncTenantSeed_CleanTreeGreen`,
`TestExitCode_FailsOnAnyViolationZeroOnCleanSpec`) is **red at base too** — confirmed in-worktree. → **not
an M6 regression**; correctly M3-deferred with triggers in f6.2 evidence.

**Integration tests are authored (real testdb-factory files), not fabricated** — verified the functions
exist: `mark_reviewed_service_integration_test.go` (5 fns incl. the tripwire-negative
`…NoCapAssertedIsRejected`/`…DocumentReviewArm` P0001 proof), `document_review_columns_integration_test.go`,
`document_review_surfacer/job_integration_test.go` (3 fns), `submit_service_reason_integration_test.go`
(2 fns), `review_due_reader_integration_test.go` (3 fns). But none executed here (no DB). Acceptance
criteria resting **only** on authored-not-executed proof: DB CHECK rejection
(`ck_documents_effective_window`/`ck_documents_review_due_sane`/`ck_documents_reason_category`), OCC
conflict, mark-reviewed tenant isolation, tripwire-negative P0001, **surfacer idempotency + tenant
behavior**, read-port filtering, reason-persist-on-row, reason-on-`governance_events`-payload.

## C3 — Senior review of the aggregate milestone diff (`93cd6114..fc3057d4`, 65 files, +4518/−435)

- **Finding 1 (blocking — HS-7 silent contract divergence).** `internal/modules/jobs/document_review_surfacer/job.go`
  runs the surfacer tick **unseeded** under `authz.BypassSystem` and its write-port
  `ReviewSurfaceWriterPG.MarkSurfaced` (`review_surface_writer.go`) issues a **cross-tenant all-tenant
  UPDATE** on `public.documents` relying on the RLS NULL-GUC "all tenants" branch. The binding contract
  §4.2 mandates *"the job **seeds per-run identity** (`SeedTxIdentity`/`SeedTxTenant`, M3 backstop) before
  the tenant-scoped read"* and §4.3 mandates a **cross-tenant isolation** proof
  (*"tenant-B doc **not surfaced under tenant-A identity**"*). The shipped test
  `TestIntegration_Surfacer_CrossTenant_MarksDueUntouchedNotDue:130` asserts the **inverse**:
  *"cross-tenant sweep must cover both tenants"*. §4 is a **named load-bearing clause**
  (contract header line 12). No HS-7 erratum was recorded in the contract, the README hard-stops table,
  or the feature docs. Note the M5 watchdog precedent the code cites actually **does the opposite** for its
  tenant-scoped write: `stuck_instance_watchdog/job.go:174` explicitly `SeedTxTenant(inst.TenantID)` before
  the `governance_events` INSERT, *"no tenant-scoped async write runs unseeded outside the §2.4 allowlist"*.
  The surfacer's un-seeded tenant-scoped write is precisely the pattern the M3 async-RLS backstop invariant
  and the `ASYNC-TENANT-SEED` lint exist to forbid. This is a substantive divergence from a binding clause,
  silently taken.

- **Finding 2 (blocking — unwired side effect / dead-end).** The surfacer's entire output —
  `review_surfaced_at` (migration `0276`) written by `MarkSurfaced` — has **no consumer** beyond the
  surfacer's own idempotency predicate. F6.2 spec interview #7 promised *"the FE review-due **list filter**
  reads them"*, but the shipped `review_due=true` filter (`repository.go:485`) recomputes
  `review_due_at <= now()` and **never reads `review_surfaced_at`**; no DTO, no read path, no FE surface
  exposes it (grep across `delivery/`, `domain/model.go`, FE `documents/` = none). The hourly River job
  therefore produces a write-only column. The surfacing deliverable is functionally inert to every consumer.

- **Finding 3 (non-blocking — triple-authored predicate / split-brain risk).** The "due for review"
  business rule (`status='published' AND review_due_at<=now AND effective window`) is hand-copied into
  **three** SQL statements: `MarkSurfaced`, `ListDueForReview` (read-port), and the `ReviewDue` list filter.
  Consistent today and each site comments that it mirrors the others, but there is no single source of
  truth for the predicate. Retrospective input, not a fail on its own.

- **Clean aspects:** mark-reviewed OCC CAS (`revision_version` guard, 0-rows→`ErrMarkReviewedStaleRevision`)
  is correct; `authz.Require(document.review,"tenant")` in-tx; DB-CHECK friendly first-line mirrors present;
  `EventTypeDocumentReviewed` emitted in-tx; the 656-line approval `api.gen.go` churn is genuine
  oapi-codegen path-sort reorder (spot-checked). The T8b root-fix (`06b1929f`) — `LoadGovernedRevisionNumber`
  deriving `revision_number` in-tx so the REV≥1 gate actually fires on the HTTP path — is a real root fix,
  not a symptom patch.

- **Staff-engineer bar met?** ❌ — findings 1 and 2 mean the headline F6.2 surfacer is neither built to its
  binding contract nor consumed.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Contract/authz/DB-invariant/async subsets | pass-with-defers on the deterministic gates | build/vet/registry/tripwire/pins green; DB-invariant + async idempotency proofs authored-not-executed (no DB) |
| Regression vs prior milestones (M0–M5) | **all still pass** | M2 registry+tripwire green with `document.review` added; api-lint self-test red **identically at base** (pre-existing, not M6); iam domain green; build/vet green |
| Separation-of-powers self-check | ✅ | main session committed the `7b3f0f82` fix; validator only judged. `7b3f0f82` real: all 3 handlers now `r.PathValue("id")`, router idempotency key matches mux pattern |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| effective/expiry reuse (no dup column family) | `effective_to` unwritten | ✅ reused, no `published_at`/`effective_date` added | migration 0274; §2 honored |
| reason-for-change structured (not `revision_title`) | free-text only | ✅ distinct `reason_for_change`(+category); grep census clean | T8 + T8b root fix; audit payload carries it (authored-not-executed on DB) |
| capability via M2-generated arm | — | ✅ arm regenerated (0275), drift+golden green, registry 35 | tripwire tests green |
| **periodic review surfaced (River, no hand-rolled scheduler)** | missing | ❌ **surfacer built but (a) diverges from §4 seed/isolation contract silently, (b) side effect unconsumed** | findings 1–2 |
| mark-reviewed via M4 unified fn | — | ⚠️ intentionally NOT routed through `CanTransitionDocumentStatus` (published→published is not a transition; F6.2 interview #4 refined this pre-code) — defensible, but contract §4.4/milestone-obj-#3 wording "routed through M4" is unmet-as-written | acceptable refinement, noted |

- **Could it be built better?** Yes and it must be, to close this milestone: (1) either seed the surfacer
  per-tenant per the contract OR record an operator-approved HS-7 erratum ratifying the all-tenant sweep
  as the M5-precedent design (and flip the test from "sweep both" to the contract's isolation intent, or
  re-state the acceptance); (2) wire `review_surfaced_at` to an actual consumer (a `surfaced`-based read
  filter / DTO field) or delete the marker + migration 0276 if due-ness is authoritatively recomputed —
  do not ship a write-only side effect; (3) extract the triple-authored "due" predicate to one source.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence
- [ ] Fixture/mock passed off as real-provider proof — *(clean: evidence honestly splits live-HTTP vs authored-not-executed; F6.3 governance_events "no HTTP read surface" claim verified true)*
- [ ] Consumer contract guessed rather than read from the consumer
- [x] **Split-brain risk (finding 3)** — one "due" fact authored in three SQL sites (non-blocking, but flagged)
- [ ] Self-judged close / validator edited or fixed code — *(clean)*
- [x] **Scope/contract drift without recorded rationale (finding 1)** — surfacer diverges from binding §4.2/§4.3 with **no HS-7 erratum**; the mission's central D4/HS-7 discipline (implementation compared section-by-section, divergence recorded or ratified, never silent) was not followed for §4.
- [ ] Symptom-patch — *(clean: T8b is a genuine root fix)*

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed checks:** **C1** (F6.2 consumer contract §4 not honored), **C3** (staff bar not met — findings 1 & 2),
  **C6** (contract/scope drift without recorded HS-7 rationale; split-brain risk).
- **Root cause:** the F6.2 River surfacer — a headline eQMS deliverable — (1) **silently diverges** from the
  binding validation-contract §4.2 (per-run tenant seed / M3 backstop) and §4.3 (cross-tenant **isolation**
  proof), shipping an unseeded all-tenant sweep and a test asserting the **inverse** of the contract, with
  **no HS-7 erratum**; and (2) writes a `review_surfaced_at` marker that **no consumer reads**, leaving the
  spec-promised "list filter reads surfaced docs" unfulfilled and the surfacing side effect inert.
- **Minimum fix feature to open:** **`f6.4-surfacer-contract-and-consumer`** — it must, in one lifecycle:
  1. Resolve the §4 divergence **per HS-7**: either (a) seed the surfacer per-run/per-tenant so the M3
     FORCE-RLS backstop is live on the `MarkSurfaced` write and add a genuine cross-tenant **isolation**
     proof matching §4.3; **or** (b) obtain operator approval to re-open contract §4.2/§4.3, record a loud
     dated erratum ratifying the all-tenant-sweep design (with the M5-watchdog-write asymmetry addressed),
     and re-state the surfacer test intent so it no longer asserts the inverse of a live contract clause.
  2. **Wire `review_surfaced_at` to a real consumer** (a surfaced-based read filter/DTO/FE surface) or, if
     due-ness is authoritatively recomputed from `review_due_at`, remove the write-only marker + migration
     0276 — no unconsumed side effect ships.
  3. Extract the "due for review" predicate to a single source (close finding 3).
  4. **Execute the authored-not-executed integration suite on a DB-capable box** and attach real output for
     the surfacer idempotency/tenant proofs (and the F6.2/F6.3 CHECK, OCC, tripwire-negative, reason-persist,
     reason-on-audit rows) — this milestone's DB-invariant and async acceptance currently rests only on
     authored-not-executed proof.
- Milestone stays **active**. Main session does **not** advance to M7 and does **not** flip README status.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): N/A — verdict is FAIL; no HS-1 presented.
> - Status flipped in `README.md`: **no** (FAIL).
> - Next: HS-4 — open `f6.4-surfacer-contract-and-consumer`, run its lifecycle, re-dispatch this validator.
