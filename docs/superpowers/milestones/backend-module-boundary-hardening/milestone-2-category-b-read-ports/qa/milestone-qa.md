# Milestone 2 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-21 (HS-4 re-dispatch, after the F2.5 fix-feature remediation) · **HEAD:** `78f44f78` · **Verdict:** see C7.
> Re-validation after a prior FAIL on C4 (broken-and-undisclosed in-repo cilint test). This run re-verifies
> C4 rigorously and re-confirms C1/C2/C3/C5/C6 still hold at HEAD.

**Inputs loaded:** milestone spec; F2.1–F2.5 `spec.md`/`plan.md`/`evidence.md`; program `README.md`;
governing `mission.md` (referenced via README); aggregate diff `git diff 4ac99bed..78f44f78` (M1 close → HEAD).
All present and readable — no fail-fast.

**Environment:** `GOCACHE=$PWD/.gocache`; integration `METALDOCS_DATABASE_URL=postgres://metaldocs:metaldocs@localhost:5434/metaldocs?sslmode=disable`, `-tags integration`, test PG live on :5434. All re-runs from clean state (`-count=1` / `go clean -testcache`).

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 cd-read-port (B1,B5,B6) | ✅ — `CDFieldReader` in `controlleddocuments/domain`; port shape (`ProfileCode`, `ProcessAreaCode` + `found`) read from the 3 call sites; `db.DB` executor → tx-aware for B5/B6 in-tx | ✅ | ✅ — no view/migration/snapshot; `user_process_areas` untouched; no parallel reader | spec.md + evidence.md; ports verified in owner domain |
| F2.2 documents-read-port (B2,B3,B4) | ✅ — `ActiveInstanceReader`/`ActiveInstanceView` in `documents/domain`; CD maps 1:1; owner typed status constants, no bare literals | ✅ | ✅ — off-tx projection; only the 3 reads moved | spec.md + evidence.md |
| F2.3 taxonomy-read-port (B7,B8) | ✅ — `AreaCatalogReader` in `taxonomy/domain`; `(name,found)` + existence bool; tx-aware for in-tx B7 | ✅ | ✅ — one narrow port, non-recording (HS-PRE-1) | spec.md + evidence.md |
| F2.4 templates-read-port (N1) | ✅ — **extended** existing ADR-0030 `TemplateVersionPort` (no parallel reader); consumer resolves version id on own table then delegates | ✅ | ✅ — tenant-scope divergence documented as strict tightening | spec.md + evidence.md |
| F2.5 guard-test-realign (HS-4 fix) | ✅ — consumer = cilint suite; fixture realigned to a **live** ledger entry | ✅ | ✅ — test-only, no production/analyzer/ledger change | spec.md + plan.md + evidence.md |

- All 5 features have `spec.md`/`plan.md`/`evidence.md`; plans are execution-shaped (tasks, files, ordering),
  not re-spec; evidence acceptance tables map row-for-row to spec Validation Gates; deviations carry rationale
  (F2.2 adapter placed in `documents/repository` not `documents/infrastructure`; F2.4 tenant-scope note).
- Interview records populated on all five (Q&A tables for F2.1/F2.2/F2.3/F2.4; explicit justification for F2.5).
- **Minor observation (non-blocking):** F2.3 and F2.4 `spec.md` carry no explicit per-feature
  `Approved before code: <date>/<operator>` header line (F2.1/F2.2 do, citing the M2 spec-gate). Mitigated:
  `milestone.md` records the **milestone-level** M2 spec-gate operator approval (leandrotca, 2026-06-20)
  authored *before any feature began*, and both specs are substantive (interview + consumer-contract +
  concrete gate), not thin. Recorded as a documentation gap, not a C1 fail.

## C2 — Gates re-run, isolated

| Feature | Command re-run (clean, real PG :5434) | Real output | Pass? |
|---------|----------------------------------------|-------------|-------|
| F2.1 ports | `go test -tags integration -count=1 -run 'TestCDFieldReader_ProfileCode_ParityWithRawSQL\|...ProcessAreaCode...' .../controlleddocuments/infrastructure/` | `ok ... 3.303s` | ✅ |
| F2.1 resolvers | `...-run TestLoadDocumentAreaCode_ParityPrePostPort .../documents/application/` + `TestLoadInstanceAreaCode_ParityPrePostPort .../approval/application/` | both `ok` | ✅ |
| F2.2 | `...-run TestActiveInstanceReader_ParityWithRawGetActiveInstance .../controlleddocuments/infrastructure/` | `ok`; verbose: 4/4 subtests PASS (active_draft_only/published_only/under_review_with_in_progress_approval/none) | ✅ |
| F2.3 | `...-run TestAreaCatalogReader_Area .../taxonomy/infrastructure/` | `ok 3.148s` | ✅ |
| F2.4 | `...-run TestLoadFillInSchema_ParityWithRawJoin .../documents/application/` | `ok`; verbose: 3/3 subtests PASS (present_schema/absent_document/null_schema) | ✅ |
| F2.5 (HS-4) | `go clean -testcache && go test -count=1 -v -run TestHGCrossModule ./tools/cilint/...` | 8/8 PASS incl. `TestHGCrossModule_Negative_PendingBaseline`; `ok ... 1.432s` | ✅ |

- Subtests confirmed to genuinely execute (verbose `=== RUN` lines, real PG seeding) — named tests are not zero-matched.
- All parity proofs are **real-provider (PG)**, explicitly distinguished from unit fakes in evidence.

## C3 — Senior review of the aggregate milestone diff

Reviewed `git diff 4ac99bed..78f44f78` as one unit (57 files, +1837/−227 excl. docs).

- **Findings:** none material. The 4 ports are interfaces in the **owner's** `domain` package
  (`controlleddocuments/domain/cd_field_reader_port.go`, `documents/domain/active_instance_port.go`,
  `taxonomy/domain/area_catalog_reader_port.go`, `templates/domain/template_version_port.go`); adapters in
  owner `infrastructure`/`repository`; all wired at the composition root (`apps/api/.../main.go` +
  `apps/jobs/.../main.go`). No import cycle (build clean). No duplicated logic, no contract defined two ways.
- **No split-brain:** each fact now has exactly one home — the owner's port. The cd-area-read column-shape
  change was handled by a single shared `docAreaRows` driver type, not N per-file copies.
- **No dead code / no superseded approach left behind:** all 9 raw reads deleted (grep below).
- **No feature broke another:** all touched-module unit suites green; integration build/vet clean.
- **Staff-engineer bar met?** ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api: persistence + module-boundaries / DDD ownership) | pass | Producer=owner, consumer imports interface only; tx-aware ports for in-tx B5/B6/B7; HS-PRE-1 honored — no authz GUC/recording write in any adapter (`git grep` of `set_config\|authz\|record` in the 4 adapters → only confirming doc-comments). Parity tests use the canonical integration harness on real PG. |
| **Broken-and-undisclosed in-repo test (prior C4 FAIL)** | **pass** | `go test ./tools/cilint/...` GREEN from clean state (`go clean -testcache`); `TestHGCrossModule_Negative_PendingBaseline` realigned (F2.5) to `{controlleddocuments/infrastructure/repository.go, user_process_areas}` — **confirmed present in `hgPendingRemediation`** today (C1+C2 row, `hgcrossmodule.go:102`). `user_process_areas` owner = `iam` (`hgcrossmodule.go:52`), reader = `controlleddocuments` → genuinely cross-module, suppressed **only** by the ledger → true suppression proof, not a test passing for an unrelated reason. F2.5 diff = test file only (no production/analyzer/ledger change); guard binary `go run ./tools/cilint ./...` exit 0. |
| Scan for OTHER undisclosed broken tests in M2 scope | pass | `go vet -tags integration` across all touched modules + cilint → exit 0 (compile-clean); all touched-module unit suites green. The cilint test was the **only** previously-green in-repo test M2 broke. |
| Regression vs prior milestones | all still pass | M0 ADR-0039: no ADR files in M2 diff. M1 typed constants: `resolution.go` not in M2 diff. Ledger: only the 9 B/N1 entries removed; C1/C2/C3 (M3) + C4a–e (M4) untouched. `go run ./tools/cilint ./...` exit 0 on full tree. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Category-B class in `hgPendingRemediation` | 9 sites (B1–B8 + N1) | **0 sites** | Ledger diff removes exactly the 9 entries; only M3 (C) + M4 (C4) rows remain. |
| 0 raw foreign base-table reads in consumer pkgs | 9 raw reads | **0** | `git grep -nE '(FROM\|JOIN)...'` for `controlled_documents` under `documents/`, `documents\|document_revisions\|approval_instances` under `controlleddocuments/`, `document_process_areas` under `documents/`+`iam/`, `templates_template_version` under `documents/` → all **CLEAN** (non-test). Each replaced by a call to the **owner's** published port — root cause (consumer encoding owner storage shape), not a consumer-local query copy or duplicate adapter (ADR-0030 no-parallel-reader honored; F2.4 extended the existing port). |
| `go build ./...` | green | green | exit 0. |

- **Could it be built better?** No material rework needed for M2. Forward-discipline note (already documented
  in F2.5): when M3 ports C1+C2 it must realign `TestHGCrossModule_Negative_PendingBaseline` again — the same
  drain-vs-stale-fixture coupling that caused the original C4 FAIL. The durable lesson — *run
  `go test ./tools/cilint/...` whenever an `hgPendingRemediation` entry is drained* — should be carried as an
  M3 checklist item. Recorded as next-milestone input; does not affect this verdict.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (per-feature parity mapped in C2)*
- [ ] Fixture/mock passed off as real-provider proof — *clean (parity is real PG; unit fakes labeled as such)*
- [ ] Consumer contract guessed rather than read from the consumer — *clean (ports read from call sites; interviews populated)*
- [ ] Split-brain (one fact, two sources of truth) — *clean (owner is single home; no parallel reader)*
- [ ] Self-judged close / validator edited or fixed code — *clean (validator judged only; wrote this file only)*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean (no migrations/views/schema; all files in declared scope; deviations have rationale)*
- [ ] Symptom-patch (bar moved by masking, root cause intact) — *clean (raw reads deleted; F2.5 realigned to a live ledger entry, not weakened/deleted)*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- The prior C4 FAIL (broken-and-undisclosed `TestHGCrossModule_Negative_PendingBaseline`) is remediated by
  F2.5 exactly as named: test-only realign to a still-pending live ledger entry; suite GREEN from clean state;
  guard binary exit 0; no production/analyzer/ledger drift; no other undisclosed in-repo test broken by M2.
  C1/C2/C3/C5/C6 re-confirmed at HEAD `78f44f78`. Both dimensions pass: code-wise (owner-published ports,
  no split-brain, no dead code, no guessed contract) and function-wise (real-PG parity across all 9 sites,
  Category-B class drained to 0).
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only the main session, on this PASS
