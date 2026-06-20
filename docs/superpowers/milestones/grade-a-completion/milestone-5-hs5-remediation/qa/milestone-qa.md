# Milestone 5 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-19 (RE-DISPATCH after HS-4 fix feature F5.8)  ·  **Verdict:** see C7 — **PASS**.
> This supersedes the prior FAIL verdict (C6 scope-drift on `61389120`). The C6 finding is now
> reconciled by F5.8 (disclosed + ratified + reclassified with independently re-run git evidence).

## Inputs loaded

- Milestone spec `milestone.md` (amended: F5.8 row, ratified observability-test exception, semconv
  reclassification note, HS-4 record).
- All 8 features' `spec.md` / `plan.md` / `evidence.md` (F5.6 has no separate `plan.md` — inline
  execution plan in `spec.md` Validation Gate; permitted under C1 "equivalent inline output present").
- Aggregate diff: commits `61389120`, `d294b4ea`, `219070b4`, `8874eaf5`, `bd903bb4` on `main`
  (parent baseline `3d71b3e6`).

## C1 — Spec & plan conformance (per feature)

All 8 features carry an approved `spec.md` (approval date + operator) before code, an evidence
acceptance table mapping to re-runnable commands, and respect declared non-goals.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F5.1 templates-literal | ✅ (`VersionStatusPublished` enum) | ✅ | ✅ | H-G literal grep 0 |
| F5.2 auth-usertenant-port | ✅ (port + Noop, off-tx) | ✅ | ✅ | H-G boundary grep 0; port files added |
| F5.3 routes-generated-typed | ✅ (strict-server DTO) | ✅ | ✅ | H-D grep 0 |
| F5.4 templates-routes-typed | ✅ (toAPITemplateDTO) | ✅ | ✅ | no map[string]any on routes |
| F5.5 iam-admin-typed | ✅ (generated admin type) | ✅ | ✅ | commit d294b4ea |
| F5.6 authz-effective-to | ✅ (REFUTED — ADR 0037) | ✅ (false-positive, evidenced) | ✅ | ADR 0037 Accepted; soft-delete model |
| F5.7 role-admin-tenant-id | ✅ (tenant_id in upsert) | ✅ | ✅ | role_admin_repository.go:66,69,118 |
| F5.8 otel-scope-reconcile | ✅ (doc reconcile, no code) | ✅ | ✅ | git evidence re-run below (C6) |

Note (non-failing): F5.6 has no standalone `plan.md`; its `spec.md` carries an execution-shaped
plan (ADR + code anchors + wiki updates enumerated in the Validation Gate) and a filled operator
approval line. C1 permits equivalent inline output. Not a fail.

## C2 — Gates re-run, isolated (validator-run, clean state)

| Feature / gate | Command re-run | Real output | Pass? |
|----------------|----------------|-------------|-------|
| Build | `go build ./...` | exit 0, no output | ✅ |
| Full suite | `go test -count=1 ./...` | FAIL count = 0 | ✅ |
| F5.1 H-G literal | `grep -rn '"published"' …/templates/infrastructure/ \| grep -v _test` | 0 matches (exit 1) | ✅ |
| F5.2 H-G boundary | `grep -rn "FROM metaldocs\.iam_user_roles" …/internal/modules/ \| grep -v iam/ \| grep -v _test` | 0 matches (exit 1) | ✅ |
| F5.3 H-D | `grep -rn "map\[string\]any" …/routes_generated.go` | 0 matches (exit 1) | ✅ |
| F5.6/F5.7 | `go test -count=1 ./…/iam/authz/… ./…/iam/infrastructure/postgres/…` | both `ok` | ✅ |
| F5.8 ratified otel tests | `go test -count=1 ./…/controlleddocuments/application/… ./…/documents/approval/application/…` | both `ok` | ✅ |

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff reviewed as one unit (8 features, parent `3d71b3e6`).

- F5.1–F5.5 are surgical literal/typed-response swaps + one new off-tx port (F5.2); no duplicated
  logic, no split-brain, no dead code introduced. The superseded `map[string]any` helpers
  (`toTemplateResponse`/`toVersionResponse`, orphan `timePtrRFC3339`) were **deleted**, not left
  dangling.
- F5.6 refutation does not change SQL; it adds Go anchor comments + ADR 0037. It *prevents* a
  split-truth (read-path predicate diverging from the unique partial index) — the right call.
- F5.7 adds `tenant_id` to both iam_users upserts (lines 66/69 and 118) incl. `EXCLUDED.tenant_id`
  on conflict to repair pre-existing sentinel rows — root-cause fix, not symptom patch.
- The one aggregate-only smell (the 24,170-line `61389120` payload) is now fully accounted for by
  F5.8 — see C6.
- Findings: none beyond the F5.6 missing-`plan.md` note (non-failing).
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api / contract) | pass | H-G=0, H-D=0, no map[string]any on public routes; build clean |
| Regression vs prior milestones (M0–M4) | all still pass | `go test -count=1 ./...` whole-repo = 0 FAIL; M0/F0.1 effective_from and M4/F4.1 literal fixes intact |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| H-G sites | 2 | 0 | boundary + literal greps both 0 (validator-run) |
| H-D sites | 2 | 0 | routes_generated.go map[string]any grep = 0 |
| Confirmed Majors | 4 | 0 | #3/#4 typed (F5.4/F5.5); #1 refuted w/ evidence (ADR 0037); #2 fixed w/ tenant_id (F5.7) |
| module-boundaries / contract-api | B+ | indicatively A− | port boundary + typed contracts in place |

- Could it be built better? The F2.3-lineage otel tests would ideally live under an M2 test path;
  F5.8 records this as a documentation tidy, not a defer. F5.6's canonical-predicate legibility
  could later be a SQL view/function (ADR 0037 D1) if re-flagged. Neither is unsound now. No FAIL.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean (per-feature mapped in C1/C2)
- [ ] Fixture/mock passed off as real-provider proof — clean (live SQL tests SKIP without DB; not claimed as live)
- [ ] Consumer contract guessed rather than read from the consumer — clean
- [ ] Split-brain (one fact, two sources of truth) — clean (F5.6 explicitly prevents one)
- [ ] Self-judged close / validator edited or fixed code — clean (validator wrote only this file)
- [ ] Scope drift (work beyond the spec, no rationale) — **RECONCILED by F5.8** (see below)
- [ ] Symptom-patch (bar moved by masking) — clean (F5.7 root-cause; F5.6 refuted not patched)

### C6 scope-drift reconciliation — independently re-run (NOT taken from the feature's word)

The prior FAIL was the 24,170-line out-of-scope payload in `61389120` against the M5
"do not touch observability" constraint. F5.8 splits it into two parts; I re-ran every cited command:

1. **Semconv vendor (~23.8k lines) = build-prerequisite repair, NOT new scope.** Verified:
   - `git show 3d71b3e6:vendor/modules.txt | grep semconv/v1.24.0` → declared at lines 316–317
     (both v1.24.0 and v1.30.0) at the **parent** commit.
   - `git ls-tree 3d71b3e6 -- vendor/.../semconv/v1.24.0` and `…/v1.30.0` → **empty** (files
     missing at parent → `go build -mod=vendor` broken).
   - `git show 61389120 --stat -- go.mod go.sum vendor/modules.txt` → **empty** (no manifest change).
   All three hold → the files were required-but-missing transitive vendor content (otelsql/otel,
   M2/F2.3, ADR 0036). Materializing them is HS-3-class prerequisite hygiene, correctly reclassified.

2. **Two `*_otel_test.go` files = additive, disclosed + operator-ratified.** Verified:
   - `git show 61389120 --name-only --diff-filter=A | grep otel_test.go` → exactly the two files at
     `internal/modules/controlleddocuments/application/service_otel_test.go` and
     `internal/modules/documents/approval/application/decision_otel_test.go` (the prior `iam/approval`
     labels were a `--stat` path-elision misread; corrected).
   - Both are `package application_test`, asserting **existing** F2.3 spans (`cd.create`,
     `signoff.record`) via `tracetest.SpanRecorder` — test-only, **no production observability code**.
   - `milestone.md` now carries the ratified, disclosed exception (Out-of-scope §, HS-4 record) and
     the semconv reclassification note. Operator ratified keeping them (2026-06-19).

An operator-ratified, disclosed deviation recorded in the milestone spec is a legitimate
disposition, not undisclosed drift. **C6 scope-drift is satisfied.** No forbidden-list hit remains.

## C7 — Verdict

- **VERDICT: PASS**
- All checks clean: C1 (8 features conform), C2 (build 0 / tests 0 FAIL / all greps 0, validator-run),
  C3 (staff-bar met, deletions clean), C4 (M0–M4 regression green), C5 (H-G=0, H-D=0, 0 Majors,
  root-cause fixes), C6 (scope-drift reconciled via F5.8 with independently re-run git evidence).
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — present before re-audit re-run.
> - Status flipped in program README: no (only on PASS — now permitted; main session's action).
