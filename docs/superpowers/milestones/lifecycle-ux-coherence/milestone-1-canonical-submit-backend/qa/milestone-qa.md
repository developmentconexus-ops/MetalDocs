# Milestone 1 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-06  ·  **Verdict:** see C7 — **PASS**.
> Aggregate diff under test: single commit `f26bb38b` (baseline `da5b45b6`; range `da5b45b6..f26bb38b`).
> HEAD == `f26bb38b`; working-tree changes are all FE (M2 scope) + user PRP-doc deletions +
> `scratch_qa/` fixtures — the Go backend under judgment is identical to the committed diff.

## Inputs loaded (none missing)

milestone.md · README.md · governing spec (`2026-07-06-lifecycle-ux-coherence-design.md`) ·
ADR 0073 · all 15 feature artifacts (F1–F5 spec/plan/evidence) · aggregate diff (`git show f26bb38b`) ·
live-QA transcript (`scratch_qa/M1-live-qa.md`).

## C1 — Spec & plan conformance (per feature)

Every feature has spec.md + plan.md + evidence.md. Every spec.md carries `Approved: 2026-07-06` with a
populated **interview waiver** ("none needed — contract fixed by ratified ADR 0073"). ADR 0073 is
Accepted 2026-07-06 and is the pinned contract, so the spike-then-formalize path satisfies C1.1
(approved-before-code) and C1.2 (interview populated-or-justified) on artifacts — the contract was
**read from a ratified ADR, not guessed**. Each plan.md is execution-shaped (task order, files touched,
test strategy). Deviations carry written rationale.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F1 in-tx-resolution | ✅ author submits no route/hash → server resolves in-tx before CAS (`submit_service.go:104-119, 342-366`); explicit values short-circuit (prior semantics) | ✅ 6/6 integration cases exist + compile+vet under `//go:build integration`; live 201 zero-gov + live 409 no-route | ✅ no FE, no idemp/ETag change | evidence.md + `scratch_qa/M1-live-qa.md` |
| F2 submit-contract | ✅ `SubmitRequest.Validate()` optional route/hash (empty OK, present-format-checked), `RevisionTitle` threaded | ✅ `TestSubmitRequestValidate` real matrix: omitted→nil, `not-a-uuid`→err, bad hash→err; `TestSubmitHandler/validate_fails`→400 | ✅ no spec re-edit; conforms to on-disk OpenAPI | contracts_test.go:59-76 |
| F3 error-mapping | ✅ four sentinel arms present w/ correct status+typed code (`errors.go:153-172`: title→422, not-draft→409, profile→400, route-missing→409) | ⚠️ met **functionally** (live 409 route-missing; F1 svc-layer 422 title) but the claimed "http unit table" over `MapErrorToResponse` does **not** exist for the 4 arms — see C1 note | evidence.md + errors.go:153-172 |
| F4 delete-finalize | ✅ one submit entrypoint; wrapper Go surface + `GetFinalizePrereqs` gone; sentinels retained & used | ✅ grep gate = 3 hits all comments, 0 live symbols; no `/finalize:` path (only `/submit` deprecation note @3394); 0 finalize refs in generated code | ✅ FreezeFinalizer / render `finalize=true` / legacy-state guard untouched | grep-gate re-run below |
| F5 idempotency-map | ✅ Go idempotent set == spec `Idempotency-Key`-declaring set (25/25, 0 orphans); mark-reviewed present; finding-17 absent both sides | ✅ parity re-verified by validator (see C2); defer bounded+triggered+owned, HS-1-flagged | ✅ no store consolidation (#22 defer); no spec header added | spec.md + C2 parity |

**C1 note (non-blocking):** F3's evidence claims a "http unit table over `MapErrorToResponse`: each
sentinel → status + non-`internal.unknown` code." No such assertions exist —
`errors_test.go`/`TestMapErrorToResponse` has no subtest for the four F3 sentinels; only
`ErrRevisionTitleRequired` is asserted, and at the **service** layer
(`submit_service_test.go:536,992`), not the HTTP mapping layer. The mapping **code** is present and
correct by inspection (statuses 422/409/400/409, all typed dot-notation codes), and route-missing→409
is **live-proven** (transcript row 2) while title→422 is service-proven. So the contract is honored and
the objective works end-to-end; the defect is **overstated test coverage in the F3 evidence table**, not
a functional or contract miss. Recorded as a retrospective tightening item (C5), not a C1/C6 FAIL:
`ErrDocumentNotDraft`→409 and `ErrProfileNotConfigured`→400 currently rest on code-inspection only.

## C2 — Gates re-run, isolated (validator-run, clean state, `-count=1`)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| all | `go build ./...` | `BUILD_EXIT=0` | ✅ |
| all | `go test -count=1 ./internal/modules/documents/approval/... ./internal/modules/documents/... ./apps/api/cmd/metaldocs-api/...` | every pkg `ok` (approval/application, http, contracts, infrastructure, jobs, documents, delivery/http, metaldocs-api) | ✅ |
| all | `go test -count=1 ./...` (C4 regression) | 96 pkgs `ok`, **0 FAIL lines** | ✅ |
| F2 | `go test -v -run TestSubmitHandler …/http` | `TestSubmitHandler/validate_fails` PASS (+7 subtests) | ✅ |
| F2 | `go test -v -run TestSubmitRequestValidate …/contracts` | PASS (matrix: omitted→nil, `not-a-uuid`→err, bad-hash→err) | ✅ |
| F3 | `go test -v -run TestMapErrorToResponse …/http` | PASS — but **no** subtest for the 4 submit sentinels (see C1 note) | ⚠️ pass, coverage-gap noted |
| F4 | `grep -rniE 'func.*finalize\|GetFinalizePrereqs\|/finalize\|finalizeDocument' internal/modules/documents/approval` | 3 hits, all comments (`submit_defaults.go:13`, `submit_service.go:375`, `postgres_approval_repository.go:1222`); 0 live symbols | ✅ |
| F4 | `grep -i finalize api/openapi/v1/openapi.yaml` | 1 hit — deprecation note inside `/submit` description (line 3394); no `/finalize:` path key (only `/submit:` @3389) | ✅ |
| F4 | `grep finalize …/api.gen.go (both)` | 0 hits (clean regeneration) | ✅ |
| F5 | spec `Idempotency-Key` decls (yaml-parsed) vs all Go idemp registrations | **25 spec decls == 25 Go** (approval map 8 + templates map 5 + controlleddocuments bespoke 4 + taxonomy bespoke 3 + approval bespoke stores 5). mark-reviewed `POST /documents/{id}/review` in `router.go:36` ✅. `/templates/{id}/archive` + `/approval-config` absent from **both** spec and Go ✅ | ✅ |
| F1 | `go vet -tags integration …/approval/application/...` | `VET_EXIT=0` — 6 real-DB cases (`TestSubmit{ResolvesActiveRoute,BindsHeadContentHash,Rev0NoGovernanceData,Rev1RequiresReason,ExplicitRouteAndHash,ReplayDuplicate}_RealDB`) compile under `//go:build integration` | ✅ compiles |

**F1 integration + live docker QA — operator-attested, NOT re-run by validator.** DB not stood up on
this Windows/Docker host (evidence: pg17 wedges at initdb; testdb bootstrap `0284_ci_rls_role.sql` needs
a connected superuser `metaldocs_app`). Per the milestone's own C2 fallback, recorded as
**operator-attested real-DB 6/6 on pg16** citing `scratch_qa/M1-live-qa.md` + F1 evidence table, and
labeled honestly as not-re-run-by-me. Validator independently confirmed the suite **compiles+vets**
under the integration tag and the six case names match the evidence one-for-one. The live docker QA
(REV0 zero-gov → 201, clean 409 sentinels, no 500s) is likewise operator-attested via the transcript.

## C3 — Senior review of the aggregate milestone diff

Reviewed `f26bb38b` as one unit.

- **In-tx resolution** (`submit_service.go:104-119` route, `342-366` content-hash) runs inside the tx
  **before** the atomic CAS `UPDATE … WHERE status='draft' AND revision_version=$3` (line 235) — the
  wrapper-era off-tx TOCTOU is genuinely closed, not shimmed.
- **ISP port** `SubmitDefaultsResolver` (`submit_defaults.go`) keeps `ApprovalRepository` untouched;
  wired by type-assertion in `services.go:77` with **fail-closed nil handling** (nil resolver →
  `ErrApprovalRouteMissing`, never `LoadRoute("")`).
- **No split-brain:** the finalize route query was *moved* (not copied) into the resolver; the wrapper
  and its `GetFinalizePrereqs` are deleted; provenance comments cite the origin line. Idempotency has one
  source of truth per module and the platform set == spec set (25/25).
- **Cross-module read** uses `CDFieldReader.ProfileCode` only (ADR 0072); the `*sql.Tx` shares GUC/RLS
  scope. Caller's form-data map is copied, never mutated. App checks mirror DB CHECK/trigger (documented
  as friendly-first-line).
- **No dead code, no feature broke another** — full `go test ./...` green (96 ok). Generated code carries
  zero finalize references.
- **Staff-engineer bar met? ✅** (with the F3 test-coverage tightening noted in C1/C5).

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api: route/contract/error discipline) | pass | contract-first (spec = route truth, no hand-route reinvention); RFC 9457 problem+json for all submit sentinels (live: no 500s); capabilities-not-roles unchanged; tenant-scoped in-tx reads; HS-PRE-1 non-recording SELECTs |
| test-discipline (testdb factory for DB integration) | pass | F1 uses `//go:build integration` real-DB suite (not unit fakes) for DB-shaped resolution — correct class |
| Regression vs prior milestones | all still pass | M1 is the **first** milestone of this program; the prior GMR program was terminal-accepted before this baseline. `go test ./...` = 96 ok, 0 FAIL — no regression against the pre-M1 baseline |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| ADR 0073 realized in Go | wrapper alive; `/submit` needs author-unsuppliable route+hash; off-tx TOCTOU | one `/submit` entrypoint owns in-tx resolution; wrapper deleted | grep-gate 0 live symbols; `submit_service.go:104-119,342-366` resolve **before** CAS (235); live 201 zero-gov |
| v0 OCC first-submit bug (findings 1,5) | no entrypoint accepted true `v0` state | `parseIfMatch` accepts `v0`; live 201 from fresh draft | transcript rows 1→2→4 (stale binary 400 → rebuilt → 201) |
| submit sentinels never 500 (finding 3) | `ErrRevisionTitleRequired` unmapped → 500 | 4 arms mapped to 422/409/400/409 typed | errors.go:153-172; live 409 (no 500s) |
| idempotency map complete (16,17) | mark-reviewed unmapped claim; #17 unmapped | 25/25 parity; mark-reviewed present; #17 correct contract-first defer | validator parity re-count |

- **Could it be built better?** The F3 mapping is **correct in production code and live-proven for the
  primary path**, but two of its four arms (`ErrDocumentNotDraft`→409, `ErrProfileNotConfigured`→400)
  are backed by **code-inspection only** — no unit assertion at `MapErrorToResponse`, contrary to the F3
  evidence claim. Recommended next-step (not blocking): add the four-row `MapErrorToResponse` table the
  F3 evidence describes, so the submit error contract is regression-locked at the HTTP boundary. This is
  a **test-hardening input**, not an unsoundness — the mapping code and end-to-end behavior are correct.
  Root cause is genuinely fixed, not symptom-patched.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean; each feature's acceptance is mapped to a validator-run gate or an honestly-labeled operator attestation.*
- [ ] Fixture/mock passed off as real-provider proof — *clean; integration is honestly labeled operator-attested-not-re-run; live QA is real docker.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean; contract pinned by ratified ADR 0073, not guessed.*
- [ ] Split-brain (one fact, two sources of truth) — *clean; finalize query moved not copied; idempotency 25/25 single-source.*
- [ ] Self-judged close / validator edited or fixed code — *clean; validator judged only, wrote one file, changed no source.*
- [ ] Scope drift — *clean; F5 #17 defer is pre-authorized by milestone.md F5 rule; FE left as M2; no extra work.*
- [ ] Symptom-patch — *clean; TOCTOU closed in-tx, v0 accepted at parser, wrapper deleted.*

(All unchecked = clean. The F3 evidence-vs-test-coverage gap is a documented C1/C5 tightening note; it is
**not** a forbidden-list hit — the mapping is real production code, live-proven on the primary path, with
no false real-provider claim.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-level, contract-clean, no split-brain, no dead code, no
  guessed contract) and **function-wise** (live docker 201 zero-gov submit + clean typed 409/422
  sentinels; operator-attested 6/6 real-DB integration; 96/96 pkgs green; F4 grep-clean; F5 25/25
  parity). Both operator-known deviations are correctly bounded: the spike-then-formalize path satisfies
  C1 because the contract was pinned by ratified ADR 0073 (not guessed), and the F5 finding-17 defer is
  the milestone's own pre-agreed rule executing (map iff spec declares the key), bounded + triggered +
  owned + HS-1-flagged.
- **One tracked improvement (non-blocking):** F3 evidence overstates HTTP-layer test coverage — add the
  four-row `MapErrorToResponse` unit table it claims. Carry as a hardening item, not a fix-feature.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — surface the F5 finding-17 defer + the F3 test-hardening note.
> - Status flipped in `README.md`: no — only the main session, only on this PASS.
