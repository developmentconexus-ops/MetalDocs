# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front, HS-6-amended spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-15  ·  **Verdict:** see C7 — **PASS**.
> The M3 code changes were judged **uncommitted in the working tree** (operator commits at the HS-1
> gate). All re-runs below executed against that working tree.

## Inputs loaded (none missing → did not FAIL-fast)

- Milestone spec `../milestone.md` (incl. the top-of-file HS-6 scope-reconciliation note + per-feature
  acceptance table). ✅
- Every feature's evidence: F3.1, F3.2 (verify-already-done evidence rows); F3.3, F3.4, F3.5
  (spec+plan+evidence). F3.6 STRUCK (HS-6). ✅
- Program `README.md` (status table: M0/M1/M2 passed; M3 in-progress; HS-1/HS-6 history). ✅
- Governing spec `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
  (referenced via the milestone's §5.2/§6 quotes + the F3.4 HS-6 reconciliation). ✅
- Aggregate milestone diff: `git status` + `git diff` over the 5 source/test files + README. ✅

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F3.1 (verify-already-done) | ✅ n/a (no code) | ✅ | ✅ | All 7 named symbols absent from the **real source tree** (`internal/`, `cmd/`) — see C2. `coverage_boost_test.go` present and live. |
| F3.2 (verify-already-done) | ✅ n/a (no code) | ✅ | ✅ | Deadlock read (`ensureTemplateArtifact`/`ResolveTemplateVersionID`) pre-flighted OFF-TX before `s.runner.Do` (`controlleddocuments/application/service.go:278–292`); residual in-tx reads `GetTemplateVersionState`/`CodeExists` are plain non-authz SELECTs (`infrastructure/repository.go:702,72`). Port refactor explicitly deferred to M4 F4.2. |
| F3.3 | ✅ producer→generated `AtomicCreateResponse`, read from OpenAPI + FE first | ✅ G1–G6 | ✅ | Handler emits `controlleddocumentsapi.AtomicCreateResponse` via reused `controlledDocumentResponse`/`documentRefResponse` mappers; only wire delta is `null`→omitted on 3 absent optionals (drift-fix onto the declared `,omitempty`). No OpenAPI/FE regen. |
| F3.4 | ✅ `(items,total,hasMore)` external shape preserved; repo→service boundary only | ✅ G1–G6 | ✅ | CTE single-query; `COUNT(*) OVER()` over base filter **inside** the CTE, cursor predicate in the **outer** query (the exact bug the spec warns against is avoided — confirmed in the diff). `CountDocuments` retained as a standalone capability; only its list-path call removed. |
| F3.5 | ✅ structured WARN via existing in-module `slog` convention; behavior-preserving | ✅ G1–G5 | ✅ | Single `DeleteObject` swallow site (`documents/application/service.go:535`) now guarded; WARN carries `storage_key`/`document_id`/`err`; still returns `ErrContentHashMismatch`, delete best-effort. 3-site→1-site reconciliation resolved under the row's documented "or documented" clause. |
| ~~F3.6~~ | — | n/a (STRUCK) | ✅ | Correctly struck, not silently skipped: both auth marshallers are live security-redaction guards; deleting one would risk leaking secrets. Recorded at the HS-6 operator gate. |

No spec/evidence row missing. No deviation lacks a written rationale. **C1 pass.**

## C2 — Gates re-run, isolated (validator-run, clean state — not trusted from transcripts)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| whole-tree | `go build ./...` | exit 0, no output | ✅ |
| F3.1 | `grep -rnE '<7 symbols>' --include='*.go' internal cmd` | `NO MATCHES in internal/ cmd/` (note: a stale `.claude/worktrees/agent-…` checkout still contains them — that is **not** the production tree; the real source tree is clean) | ✅ |
| F3.2 | read `service.go:255–295`, `infrastructure/repository.go:702,72` | pre-flight comment + off-tx ordering confirmed; both residual reads are plain `SELECT` on `c.db`/`r.db`, no `authz.Require`/`GetByCode` | ✅ |
| F3.3 G1 | `go test ./internal/modules/controlleddocuments/delivery/http/ -run TestAtomicCreate_UsesGeneratedResponse -count=1` | `--- PASS … ok …controlleddocuments/delivery/http 2.349s` | ✅ |
| F3.3 G2 | `grep -n 'map\[string\]any' routes.go` | only `:89`,`:224` (request-side `formData` decode); the `:123` 201-response map is gone | ✅ |
| F3.3 G3 | `go test ./internal/modules/controlleddocuments/... -count=1` | all `ok` (application/delivery-http/domain/infrastructure) | ✅ |
| F3.4 G1/G3/G4 | `go test ./internal/modules/documents/repository/ -run TestListDocumentsPaginated -count=1` | 4 subtests PASS, `ok …documents/repository`; mock asserts `COUNT\(\*\) OVER\(\)` on page 1 and `(updated_at, id) < ($2::timestamptz, $3)` on page 2; `total1==25` and `total2==25` (page-independent grand total) | ✅ |
| F3.4 G2 | `grep -n CountDocuments service.go` | only the interface decl (`:36`); no call in the list service path | ✅ |
| F3.5 G1/G2 | `go test ./internal/modules/documents/application/ -run TestCommitAutosave -count=1` | `TestCommitAutosave_LogsCleanupFailureOnHashMismatch --- PASS`; asserts `level=WARN` + `storage_key` value + `document_id=doc_1` + `err`, `ErrContentHashMismatch` returned, `deleteCalls==1` | ✅ |
| F3.5 G3 | `grep -nE '\.DeleteObject\(' service.go` | `535: if err := s.presigner.DeleteObject(...); err != nil {` — only call, now guarded | ✅ |
| all G4/G5 | `go test ./internal/modules/documents/... -count=1`; `go build ./...` | all packages `ok`; build clean | ✅ |
| F3.3/F3.4/F3.5 lint | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` | `Your API description is valid. 🎉` (no OpenAPI edit → no contract drift) | ✅ |

Every named gate re-ran green from clean state. **C2 pass.**

## C3 — Senior review of the aggregate milestone diff

Diff scope = exactly the declared 5 files (3 source + their tests) + README + the M3 doc tree; nothing
adjacent touched (surgical, per CLAUDE.md §5.3). Reviewed as one unit:

- **F3.3 routes.go** — raw `map[string]any` replaced by the generated `AtomicCreateResponse` built from
  the **already-existing** Get/List mappers; mapper-error path mirrors the sibling Get handler. No new
  mapper, no duplication, no second contract source. Clean.
- **F3.4 repository.go** — CTE is correctly structured: `COUNT(*) OVER() AS total_count` lives inside
  the CTE over the base `WHERE %s` (no cursor), and `cursorClause` is interpolated into the **outer**
  `FROM filtered%s`. Arg numbering preserved (`len(args)-1`,`len(args)` for the cursor pair; `limit+1`
  last). 19-col scan aligned. Empty-result `total=0` edge commented. This is the spec-correct
  construction, not a band-aid. `service.go` drops the second round-trip and takes `total` from the one
  call — a genuine TOCTOU root-cause fix (single MVCC snapshot), not symptom masking.
- **F3.5 service.go** — one swallow → `if err != nil { slog.WarnContext(...) }`; behavior identical
  (`ErrContentHashMismatch` still returned, delete still best-effort). Adds observability without
  hiding the failure. Uses the in-module `slog` convention (no new `Service` field).
- **No split-brain:** F3.3 keeps OpenAPI/generated type as the single response contract source; F3.4
  keeps one total source (the windowed CTE) and removes the divergent second `CountDocuments` query
  from the list path; F3.5 adds no new fact. **No dead code** introduced (retained `CountDocuments` is
  a still-tested public capability, explicitly recorded). **No feature breaks another.**

Staff-engineer bar met? ✅ **C3 pass.**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| `backend-api-qa-checklist` (all features) | pass | Build clean; contracts lint-clean; no shape drift; error path mirrors canonical Get (F3.3). |
| `workflow-async-qa-checklist` (F3.2, CD-create lock-bearing path) | pass | H-PRE-1 advisory-lock invariant intact: deadlock read hoisted off-tx, no Tx-variant authz read slipped inside the lock; residual in-tx reads non-authz. |
| Regression vs prior milestones (M0/M1/M2) | all still pass | `go test ./...` → **exit code 0**, 0 `FAIL`/`panic` lines, 87 `ok` packages. The M1 full-HTTP `seed→finalize→signoff` E2E area (`documents/approval/http`, `tests/`) all green — F3.4/F3.5 touch the documents/commit area without regressing it. |

**C4 pass** (clean, with the bounded defers in C5 carried — not blockers).

## C5 — Quality-bar re-measure + retrospective

Milestone bar: **code-quality + persistence dimensions ≥ A−**, root causes **fixed not symptom-patched**.

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Dead surface (7 orphans) | flagged in 06-13 audit | gone | Already removed in Wave 2.11/2.12; re-proven zero-caller against the real `internal/`+`cmd/` tree (C2). Root cause = surface deleted, not hidden. |
| Tx-hazard (H-PRE-1 deadlock) | flagged | closed | Authz-recording read hoisted OFF-TX (`service.go:278–292`); residual in-tx reads proven non-authz. H-PRE-1 invariant intact — no Tx-variant read inside the lock. |
| Raw-map response (F3.3) | producer emitted domain types as `map[string]any` → `null` on absent optionals | producer emits generated `AtomicCreateResponse` → optionals omitted | Producer aligned to the already-declared `,omitempty` contract the FE already types; root cause (untyped producer) removed, not papered over. |
| Pagination TOCTOU (F3.4) | two statements (page query + separate count), no shared snapshot | one CTE statement, page rows + grand total share one MVCC snapshot | Real single-snapshot fix; the separate-count race window is eliminated, not narrowed. CTE windows the count pre-cursor (correct keyset semantics). |
| Swallowed delete error (F3.5) | `_ = DeleteObject(...)` (silent) | structured WARN on failure | Observability added; failure surfaced, not masked; behavior unchanged (best-effort preserved). |

Both **code-quality** and **persistence** dimensions re-measure **≥ A−**. Root causes fixed, not
symptom-patched.

**Bounded defers judged (acceptable, not FAIL):**
- F3.4 — *no real-DB integration test* proves the CTE cursor-vs-window placement; sqlmock returns
  author-supplied `total_count`. **Judged acceptable:** the placement is correct in source (verified in
  the C3 diff review), guarded at the shape level by the `COUNT\(\*\) OVER\(\)` + outer-cursor mock
  assertions and the two-stage review; the gap is a unit-tier *regression guard*, not a present defect.
  Honestly labeled (sqlmock = fixture) with a written trigger (Postgres documents harness / M4). This
  is the kind of bounded defer the gate permits — it does not move the bar by masking.
- F3.3 — `go generate` rewrites the embedded `swaggerSpec` gzip blob with **zero Go type change**;
  reverted to keep the feature surgical; recorded as a codegen-freshness micro-task. Benign.
- F3.5 — non-atomic test counter (single-goroutine test). Benign.

Could it be built better? The F3.4 real-DB integration test is the one materially-better construction —
correctly routed to a future integration-test milestone, not silently dropped. No rework required now.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean*: every
      feature's acceptance is mapped to a specific re-run command + real output in C2.
- [ ] Fixture/mock passed off as real-provider proof — *clean*: sqlmock (F3.4), `fakePresigner`/buffer
      `slog` (F3.5), `httptest`/spy (F3.3) all **explicitly labeled fixture**; the F3.4 real-DB gap is
      declared, not hidden.
- [ ] Consumer contract guessed — *clean*: F3.3 read from OpenAPI + FE `controlledDocuments.ts`; F3.4
      read from the handler/service boundary; F3.5 read from the in-module `slog` convention.
- [ ] Split-brain — *clean*: one response-contract source (F3.3), one total source (F3.4), no new fact (F3.5).
- [ ] Self-judged close / validator edited code — *clean*: validator judged and wrote only this file;
      no source/spec/evidence edited; status not flipped (left to the main session on PASS).
- [ ] Scope drift — *clean*: diff = exactly the declared files; F3.6 struck with rationale; deferred
      items (F4.2 port, codegen-freshness, integration test) recorded with triggers.
- [ ] Symptom-patch — *clean*: F3.4 is a single-snapshot root-cause fix; F3.5 surfaces (does not mask)
      the failure; F3.3 aligns the producer to the declared contract.

All unchecked = clean. **No C6 hit.**

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass independently: **code-wise** (senior-clean aggregate diff, no split-brain/dead
  code/guessed contract) and **function-wise/QA** (every per-feature acceptance re-run green from clean
  state; full-tree regression `go test ./...` exit 0 / 87 `ok`; lint clean; H-PRE-1 invariant intact).
  The verify-already-done rows (F3.1, F3.2) are independently re-proven against the real source tree;
  F3.6 is correctly struck. Bounded defers are honestly labeled and acceptable.
- Handed back to the main session to flip status (M3 → passed in `README.md`) and present the HS-1
  operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — present this PASS to the operator.
> - Status flipped in `README.md`: no — only the main session, only on this PASS.
> - Carry-forward defers to honor: F4.2 `TemplateVersionStateReader` port (M4); F3.4 real-DB
>   integration test (integration-test milestone); CD `api.gen.go` codegen-freshness micro-task;
>   stale wiki/GitNexus refs to deleted symbols (wiki-curator).
