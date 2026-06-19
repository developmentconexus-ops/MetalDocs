# Milestone 5 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-19  ·  **Verdict:** see C7. **VERDICT: FAIL (C6 scope drift).**

## Inputs loaded

- Milestone spec `../milestone.md` — OK
- F5.1–F5.7 `spec.md` / `plan.md` / `evidence.md` — all read (see C1 note re F5.6 plan.md)
- Program `../../README.md` — OK
- Governing spec `mission.md §8` + re-audit `wiki/backend/_artifacts/architecture-re-audit-2026-06-16.md` — OK
- Aggregate diff: `3d71b3e6..HEAD` = commits 61389120 (F5.1–F5.4), d294b4ea (F5.5), 219070b4 (F5.6), 8874eaf5 (F5.7) — OK

## C1 — Spec & plan conformance (per feature)

| Feature | spec approved before code | plan.md execution-shaped | evidence↔gate match | Consumer contract honored | Non-goals respected | Result |
|---------|---------------------------|--------------------------|---------------------|---------------------------|---------------------|--------|
| F5.1 templates-literal | ✅ Approved 2026-06-16 | ✅ | ✅ | ✅ intra-module domain constant | ✅ | ✅ |
| F5.2 auth-usertenant-port | ✅ Approved 2026-06-16 | ✅ | ✅ | ✅ iam-owned read port (mirrors TenantUserReader) | ✅ | ✅ |
| F5.3 routes-generated-typed | ✅ Approved 2026-06-16 | ✅ | ✅ | ✅ generated 200 types | ✅ | ✅ |
| F5.4 templates-routes-typed | ✅ Approved 2026-06-16 | ✅ | ✅ | ✅ strict-server DTOs | ✅ envelopes carved out w/ rationale | ✅ |
| F5.5 iam-admin-typed | ✅ Approved 2026-06-19 | ✅ | ✅ | ✅ iamapi.AdminOverviewResponse | ✅ roles routes deferred w/ rationale | ✅ |
| F5.6 authz-effective-to | ✅ DECIDED 2026-06-19 (Option A, operator-approved) | ⚠ no `plan.md` file (inline in spec) | ✅ | ✅ refutation evidence-backed | ✅ | ✅ (see note) |
| F5.7 role-admin-tenant-id | ✅ Approved 2026-06-19 | ✅ | ✅ | ✅ caller tenant_id threaded | ✅ | ✅ |

**F5.6 plan.md note:** F5.6 has no separate `plan.md`. Its `spec.md` is a decision record (Option A —
refute & document, doc-only) carrying the execution shape inline (files-touched table, validation gate,
disposition). Per C1's "binds on artifacts, not skills — equivalent inline output present = PASS", the
requirement is met. Recorded as a deviation with rationale; **not** a C1 fail.

**C1: PASS.** Every feature has an approval line with date+operator; interview records populated;
acceptance rows map to re-runnable commands.

## C2 — Gates re-run, isolated (validator-run, clean state)

| Gate / Feature | Command re-run | Real output | Pass? |
|----------------|----------------|-------------|-------|
| H-G grep #1 (iam_user_roles) | `grep -rn "FROM metaldocs\.iam_user_roles" … internal/modules/ \| grep -v iam/ \| grep -v _test` | 0 matches (exit 1) | ✅ |
| H-G grep #2 (published literal) | `grep -rn '"published"' … templates/infrastructure/ \| grep -v _test` | 0 matches (exit 1) | ✅ |
| H-D grep | `grep -n "map\[string\]any" …/routes_generated.go` | 0 matches (exit 1) | ✅ |
| F5.4 helpers removed | `grep -rn "toTemplateResponse\|toVersionResponse" …templates/` | 0 matches (exit 1) | ✅ |
| F5.7 tenant-less upsert | `grep "iam_users (user_id, display_name, is_active" internal/ \| grep -v _test` | only `e2e_seed.go:509` which DOES carry `tenant_id` → not tenant-less | ✅ |
| F5.1 | `go test -run TestIsPublishedComparesAgainstDomainConstant ./…/infrastructure/` | PASS | ✅ |
| F5.2 (unit) | `go test -run UserTenant ./…/iam/infrastructure/postgres/` | `TestUserTenantsQueryParity` PASS, `TestNoopUserTenantReaderReturnsEmptyNonNil` PASS | ✅ |
| F5.2 (live) | DB env unset → `go vet -tags integration …/postgres/` | VET OK (compiles); live run deferred, honestly labeled, no fixture-as-live | ✅ (deferred) |
| F5.3 | `go test -run 'Presign…\|Publish…DeclaredKeysOnly' ./…/delivery/http/` | both PASS | ✅ |
| F5.5 | `go test -run 'HandleAdminOverview_DecodesIntoGeneratedContract\|_DropsUsersField' …` | both PASS | ✅ |
| F5.6 | sqlmock exact-match `Test*PassesTenantID`, `Test*DeleteThenInsert` (authz SQL verbatim) | PASS → SQL byte-identical confirmed | ✅ |
| F5.7 | `go test -run 'PassesTenantID\|DeleteThenInsert_PersistsSingleRole' …/postgres/` | both PASS; test asserts 3-arg INSERT (`"alice","Alice",testTenant`) → red-on-2-arg real | ✅ |
| Module suites | `go test -count=1 ./templates/... ./iam/... ./auth/... ./controlleddocuments/... ./search/...` | all `ok` | ✅ |
| Build | `go build ./...` | exit 0 | ✅ |

**C2: PASS.** All milestone-defined gates re-run clean from validator's state.

## C3 — Senior review of the aggregate milestone diff

Reviewed `3d71b3e6..HEAD` as one unit.

**In-scope work is clean and senior-level:**
- F5.1 literal→constant: correct single source of truth.
- F5.2 port extraction mirrors the existing `TenantUserReader` (ADR 0031), off-tx (H-PRE-1 respected),
  query parity pinned; no import cycle; nil→Noop default sound.
- F5.3/F5.4/F5.5 typed-response lifts: helpers deleted, no orphan callers, strict-server structs are
  the single contract source; publish-200 undeclared-field leak closed at the type.
- F5.6 refutation: source changes are **comment-only** (verified — no SQL line added/removed in
  authz.go / user_area_repository.go / reader.go); sqlmock exact-match tests pass → SQL byte-identical;
  ADR 0037 present; refutation correctly grounded in schema partial-unique index + no-end-date Grant +
  read-only DTO. Legitimate evidence-backed disposition, not a dodge.
- F5.7 real fix: `tenant_id`/`$3::uuid`/`EXCLUDED.tenant_id` added to both upserts; TDD red→green real.
- No split-brain, no duplication, no dead code, no feature breaking another **within the 7 features**.

**Finding — out-of-scope payload bundled into commit 61389120 (the "F5.1+F5.2+F5.3+F5.4" commit):**
that single commit also adds **24,170 insertions** that are NOT any of the 7 M5 features and are **not
mentioned** in the commit message or any M5 artifact:
- `vendor/go.opentelemetry.io/otel/semconv/v1.24.0/**` and `**/v1.30.0/**` (~23.8k lines vendored;
  v1.30.0 was absent at base `3d71b3e6`).
- `internal/modules/controlleddocuments/application/service_otel_test.go` (+179)
- `internal/modules/documents/approval/application/decision_otel_test.go` (+167)

These are **OTel span / composition-observability** artifacts. The M5 spec is explicit: *"In scope:
exactly the 7 features below — no wider refactor"* and *"Do not touch composition/observability (already
A−)."* This payload directly touches the dimension the milestone forbade, carries **no recorded
rationale anywhere**, and was folded silently into a feature commit. It compiles and the new tests pass
(no functional breakage), but its presence is unexplained scope.

- **Staff-engineer bar met for the 7 features? ✅. For the milestone diff as delivered? ❌** — a staff
  engineer would reject the commit for bundling 24k lines of unrelated, unmentioned, spec-forbidden
  observability vendor+tests into a remediation feature commit.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api / contract) | pass | typed responses on cited public routes; H-D/H-G grep gates 0 |
| Whole-repo regression `go test -count=1 ./...` | **0 FAIL** (exit 0) | re-run from clean state by validator |
| M0–M4 sentinels | clean | targeted module suites + full suite green; no prior-milestone gate regressed |

**C4: PASS** (function/regression dimension). The OTel test additions, though out of scope, do not
regress any suite.

## C5 — Quality-bar re-measure + retrospective

| Bar (mission §8 / re-audit) | Before (2026-06-16) | After (this run) | Root-cause-fixed evidence |
|-----------------------------|---------------------|------------------|---------------------------|
| H-G | 2 | **0** | both grep gates exit 1 (F5.1 constant, F5.2 port) |
| H-D | 2 | **0** | grep on routes_generated.go exit 1 (F5.3) |
| Confirmed Majors | 4 | **0** | #1 refuted w/ evidence (ADR 0037, SQL byte-identical); #2 fixed (F5.7 TDD); #3 fixed (F5.4 helpers deleted); #4 fixed (F5.5 typed) |
| module-boundaries / contract-api ≥ A− | B+ / B+ | indicatively met by the above closures | *re-grade is the main-session re-audit's job at HS-1, not the validator's* |

Root causes are fixed, not symptom-patched: F5.2 moves the boundary violation behind a port (not a
suppression); F5.7 threads the real tenant (not a default tweak); F5.6 refutes rather than applying a
dead `OR` clause that would have regressed the index (the *correct* anti-symptom-patch call).

**Retrospective / could it be built better:** F5.6 carries no `plan.md` (inline-acceptable but
inconsistent with the other six). The OTel drift (C3) should have been a separate observability
commit/feature or omitted entirely. Otherwise the construction is sound.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean (per-feature acceptance mapped in C2)
- [ ] Fixture/mock passed off as real-provider proof — clean (F5.2 live test honestly labeled deferred, not faked)
- [ ] Consumer contract guessed rather than read from the consumer — clean
- [ ] Split-brain (one fact, two sources of truth) — clean (F5.6 explicitly preserves single active-predicate truth)
- [ ] Self-judged close / validator edited or fixed code — clean (validator only judged + wrote this file)
- [x] **Scope drift (work beyond the spec, no rationale)** — **HIT.** Commit 61389120 bundles 24,170
      insertions of OTel semconv vendor + two `*_otel_test.go` files — composition/observability work the
      M5 spec explicitly forbade ("Do not touch composition/observability"), with no rationale in the
      commit message or any M5 artifact.
- [ ] Symptom-patch (bar moved by masking) — clean

**C6: FAIL (scope-drift hit).**

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed check:** **C6** (scope drift) — corroborated by **C3** (aggregate-diff senior bar not met as
  delivered). All technical gates (C1, C2, C4, C5) and the 7 in-scope features pass; the milestone fails
  the gate solely because 24,170 lines of unrelated, spec-forbidden OTel observability vendor+tests were
  committed under the M5 feature commit with no recorded rationale.
- **Minimum fix feature to open:** **`f5.8-otel-scope-reconcile`** — reconcile the out-of-scope payload
  in commit 61389120 by one of:
  1. **Remove** the OTel semconv vendor (`vendor/go.opentelemetry.io/otel/semconv/v1.24.0`, `v1.30.0`)
     and the two `*_otel_test.go` files from the M5 line if they are not required by the 7 features
     (confirm `go build ./...` + `go mod verify` still clean without them); **or**
  2. **Document & ratify** the inclusion: record an operator-approved rationale (these are an M2
     observability follow-on / Minor #23 closure pulled forward) in an M5 artifact + amend the milestone
     spec's scope, and split/annotate the commit so the observability change is not silently inside a
     remediation feature commit.
  Then re-dispatch the validator. Milestone stays **active**; main session does **not** advance, does
  **not** flip status, does **not** run the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): **blocked** pending C6 clearance (HS-4 fires: open `f5.8`, re-dispatch).
> - Status flipped in `README.md`: **no** (FAIL).
