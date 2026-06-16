# Milestone 1 — contract-integrity — Close-Gate Verdict

**Date:** 2026-06-15
**Validator:** milestone-validator (independent subagent)
**Milestone spec:** docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/milestone.md
**Verdict:** PASS

---

## Inputs loaded

| Input | Path | Status |
|-------|------|--------|
| Milestone spec | `milestone-1-contract-integrity/milestone.md` | READ |
| Governing program spec | `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` | INDEX READ (program-level M0..M5) |
| Mission (program governing) | `docs/superpowers/milestones/grade-a-completion/mission.md` | READ |
| F5.1 re-audit report (defines §6 H-D grep commands + DRIFT-1..4 sites) | `wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md` | READ (lines 141–170 — H-D class re-measurement) |
| Binding checklist | `.claude/skills/milestone/references/milestone-end-validation.md` | READ |
| F1.1 trio | `f1.1-checkpoints-typed/{spec,plan,evidence}.md` | READ |
| F1.2 trio | `f1.2-status-and-body-conformance/{spec,plan,evidence}.md` | READ |
| F1.3 trio | `f1.3-declared-fields-only/{spec,plan,evidence}.md` | READ |
| F1.4 trio | `f1.4-typed-responses-class/{spec,plan,evidence}.md` | READ |
| Aggregate milestone diff | `git log 81ce1815..HEAD` (six commits, milestone-start = M0 close at `81ce1815`) | READ |

All 12 feature artifacts present; no missing-artifact FAIL.

---

## C1 — Per-feature spec/plan conformance

### F1.1 — checkpoints-typed → APPROVED

- spec.md `Approved before code: 2026-06-15 / leandrotca.work` — populated, before commit `fb5de59d`.
- Interview record populated (Q1–Q3) with explicit consumer-contract cites (`api.gen.go:166`, `index.d.ts:2264`, FE `documents.ts:16`).
- Consumer contract = generated `DocumentCheckpoint` snake_case shape; producer (`handler.go:881,954`) now matches via `toAPICheckpoint(s)` mapper.
- plan.md is execution-shaped: file list, ordered TDD steps (red→green→grep→broader), code sketches.
- Validation Gate (5 rows) maps 1:1 to evidence.md table — all PASS.
- Non-goals respected: `restoreCheckpoint` left for F1.4; no codegen/OpenAPI/FE shim.

### F1.2 — status-and-body-conformance → APPROVED

- spec.md `Approved: 2026-06-15 / leandrotca.work` (operator confirmed A4 HS-6 deviation).
- Interview record populated (Q1–Q7) including documented operator pivot mid-flight (Q7 P2-light expansion).
- Consumer contracts cited per site (OpenAPI + FE adapter file:line for A2/A4/A5/P2-A/P2-B).
- ADR 0035 created (`wiki/decisions/0035-flat-typed-responses-and-presign-status.md`) per "ADR needed? Yes".
- Validation Gate V1–V10 map 1:1 to evidence table — all GREEN; HS-6 deviation (A4 keeps 201) recorded both in interview Q2 and milestone.md.
- P2-light scope expansion was operator-approved and reflected in milestone.md F1.2/F1.4 row swap — no silent scope drift.

### F1.3 — declared-fields-only → APPROVED

- spec.md "Approval" footer: design presented to operator 2026-06-15, approved ("Proceed").
- Interview record populated (8 Qs, all source-cited).
- Consumer contract: typed `CreateTemplateResponse` (`api.gen.go:75-81`), declared envelope kept.
- Validation Gate V1–V5 GREEN per evidence; TDD red→green proof recorded.
- D-3 (toVersionResponse/toTemplateResponse retirement) explicitly bounded as F1.4 territory per HS-6.

### F1.4 — typed-responses-class → APPROVED

- spec.md "Approval" footer: consumer contracts read from `api.gen.go` (documents/taxonomy/audit) 2026-06-15.
- Interview record populated (10 Qs) — included explicit D-1 zero-fill decision rationale (not HS-2) and D-2 scope-exclusion (map[string]string out of grep coverage).
- Validation Gate V1–V8 GREEN per evidence. V1 "scoped" wording is honest — F1.4 closes the 7 spec'd sites + drive-by site 208; out-of-scope hits are explicitly enumerated as D-2 (map[string]string) and D-3 (audit other endpoints).
- Drive-by repair (taxonomy `setDefaultTemplate` site 208) recorded as D-4, consistent with CLAUDE.md §5.3 surgical rule.

**C1 verdict: APPROVED for all four features.** No missing trio; all approvals dated pre-commit; consumer contracts read from source (not guessed); plans execution-shaped (not re-stated specs); Validation Gates map row-for-row to evidence.

---

## C2 — Gates re-run, isolated (clean state, no cache)

```
$ go test -count=1 ./internal/modules/documents/delivery/http/... ./internal/modules/taxonomy/delivery/http/... ./internal/modules/audit/delivery/http/...
ok      metaldocs/internal/modules/documents/delivery/http      2.954s
ok      metaldocs/internal/modules/taxonomy/delivery/http       2.969s
ok      metaldocs/internal/modules/audit/delivery/http          1.905s
```

```
$ go test -count=1 -run "TestListCheckpoints_TypedResponseShape|TestCreateCheckpoint_TypedResponseShape|TestRenameDocument_TypedResponseShape|TestCreateNextVersion_TypedResponseShape|TestPresignAutosave_TypedResponseShape|TestCommitAutosave_TypedResponseShape|TestGetTemplateVersion_TypedResponseShape|TestCreateTemplate_TypedResponseShape|TestCreateTemplate_Happy|TestHandleEvents_405_Allow|TestListProfiles_DropsDomainFields|TestGetProfile_DropsDomainFields" ./internal/modules/{documents,templates,taxonomy,audit}/delivery/http/... -v
--- PASS: TestListCheckpoints_TypedResponseShape (Items + Empty)
--- PASS: TestCreateCheckpoint_TypedResponseShape
--- PASS: TestRenameDocument_TypedResponseShape
--- PASS: TestCreateTemplate_Happy
--- PASS: TestCreateNextVersion_TypedResponseShape
--- PASS: TestPresignAutosave_TypedResponseShape
--- PASS: TestCommitAutosave_TypedResponseShape
--- PASS: TestGetTemplateVersion_TypedResponseShape
--- PASS: TestCreateTemplate_TypedResponseShape
--- PASS: TestListProfiles_DropsDomainFields
--- PASS: TestGetProfile_DropsDomainFields
--- PASS: TestHandleEvents_405_Allow
PASS (12/12 named acceptance tests across F1.1–F1.4)
```

### F5.1 §6 H-D drift sites — re-verification of root-cause closure

| Site | Status | Evidence |
|------|--------|----------|
| DRIFT-1 `templates/routes_generated.go:64` (`CreateTemplate`) | RESOLVED | Now emits typed `templatesapi.CreateTemplateResponse{Data:{Template,Version}}` via `toAPITemplateDTO`/`toAPIVersionDTO`. Source-verified lines 64–77. |
| DRIFT-2 `templates/routes_autosave.go:42` (`presignAutosave`) | RESOLVED | Status 200 + typed `templatesapi.TemplatePresignAutosaveResponse`. Source-verified line 42. |
| DRIFT-3 `templates/routes_create.go:36` (`createNextVersion`) | RESOLVED | Status 201 (HS-6 deviation, ADR 0035 §D2 governs) + typed `VersionDTO`. Source-verified line 36–41. |
| DRIFT-4 `taxonomy/routes_profiles.go:67,111,126,169` | RESOLVED | All 4 sites + drive-by 208 emit typed `DocumentProfileItem` / `ListDocumentProfilesResponse` via `toDocumentProfileItem`. |

Additional sites the milestone spec named (beyond F5.1's 4-count):
- A1 `documents/handler.go:881,954` (checkpoints) — typed (F1.1).
- A2 `documents/handler.go:521` (renameDocument) — 200 empty body (F1.2).
- A6 6 sites `documents/handler.go` lines 694/701/814/855/903/1023 — typed (F1.4).
- A8 stats `documents/handler.go:319` — typed `DocumentStatsResponse` (F1.4).
- A8 audit 405 `Allow: GET` header — added (F1.4).

**C2 verdict: PASS.** All declared acceptance tests pass on isolated clean re-run; all F5.1 §6 DRIFT sites + every site listed in the milestone Features table are resolved at source.

---

## C3 — Senior review of aggregate milestone diff

Commit list (milestone start at `81ce1815`, six commits to HEAD `7381b9c4`):

```
fb5de59d fix(documents-api): checkpoints emit typed DocumentCheckpoint (F1.1, A1)
9b5d22f0 feat(contracts): F1.2 status-and-body conformance — flat typed responses + canonical status (ADR 0035)
2be8e585 fix(templates): createTemplate drops undeclared id/version_id fields (F1.3/A3)
5111474a docs(f1.3): evidence.md — V1-V5 GREEN, bounded defers D-1..D-3
43da5f57 docs(f1.4): spec.md + plan.md — typed-responses class (A6/A7/A8)
7381b9c4 feat(contracts): F1.4 typed-responses class — H-D zero across documents/taxonomy/audit
```

Observations:

- **Consistent pattern across modules.** Every typed-response migration follows the same shape: domain → API mapper at handler boundary, `uuid.Parse` for UUID fields, 500 on parse error, no domain type change. F1.1 set the template (`toAPICheckpoint`), F1.2 added `toAPIVersionDTO`, F1.3 mirrored with `toAPITemplateDTO`, F1.4 mirrored with `toDocumentProfileItem`. No split-brain: one mapper per output type, one source of truth per shape.
- **OpenAPI authority preserved.** Schema changes (F1.2: `TemplatePresignAutosaveResponse`, `VersionDTO.placeholder_schema` array fix, four operation amendments) are all driven by ADR 0035 and regenerated, not hand-edited. Backend `api.gen.go` regen + FE `index.d.ts` regen diff aligned.
- **No leaked `map[string]any` regression on the spec'd sites.** Source-verified each declared site emits typed structs.
- **Legacy mappers retained intentionally.** `toVersionResponse`/`toTemplateResponse` remain in `routes_create.go` because callers in `routes_lifecycle.go`, `routes_query.go`, `routes_schema.go` still use the `{data:{...}}` envelope. F1.3 spec interview Q6/Q7 explicitly enumerated these callers and scoped retirement out (D-3). No dead code (each mapper has ≥3 callers).
- **No scope drift.** P2-light expansion of F1.2 (Q7) was operator-approved with paired F1.4 row shrink in milestone.md, recorded same edit. Drive-by repair (taxonomy setDefaultTemplate site 208) is a 1-line typed-empty-response touch, surgical under CLAUDE.md §5.3.
- **Test-fixture migration to UUID-shaped identifiers** triggered by typed contract enforcement (`uuid.Parse` in mappers). Repaired in same commits per CLAUDE.md §4 new test-framework hard gate. No production code changed by fixture repair; documented honestly as drive-by, not scope creep.
- **HS-6 deviation (A4 keeps 201 instead of mission §5 "201→200")** recorded transparently in F1.2 spec Q2, milestone.md F1.2 row, and ADR 0035 §D2 Negative consequences. Operator pre-approved.

Staff-engineer bar: this diff would pass review. The work is consistent, the contract surface is honored, no split-brain, no dead code, no scope drift.

**C3 verdict: PASS.**

---

## C4 — Workflow-class QA (backend-api-qa-checklist) + regression

Applied to the 4 changed delivery/http packages (documents, templates, taxonomy, audit):

| Checklist line | Result | Evidence |
|----------------|--------|----------|
| Route truth-table reconciled across runtime / spec / codegen / wiki | PASS | OpenAPI amended (F1.2), codegen regenerated (backend `api.gen.go` + FE `index.d.ts`), wiki/architecture/api-contract.md stamp bumped per ADR 0035. |
| Regen order honored (truth-table → spec → codegen) | PASS | F1.2 evidence §3 + ADR 0035 record order; no hand-edits to `api.gen.go`. |
| No hand-edits to generated wiring | PASS | `git diff fb5de59d..HEAD -- "**/api.gen.go"` shows only regenerated content (matching the OpenAPI delta). |
| OpenAPI shape unchanged except where a feature explicitly amends | PASS | Only the 4 F1.2 operation amendments + 2 new schemas, all declared in spec and recorded in ADR 0035. F1.3 dropped undeclared fields (subset correction, not shape change). F1.4 used existing schemas. |
| Typed responses at handler boundary (no FE shims) | PASS | FE adapter edits in F1.2 are decode-side only (drop `body.data.…` indirection); no shape translation. |
| Error mapping unchanged | PASS | All features added 500 on UUID-parse defensive errors only; existing `writeMappedErr` / `mapErr` paths untouched. |
| Authz/idempotency unchanged | PASS | No authz, route-mount, or idempotency change in any commit. |
| Status codes spec-conformant | PASS | A2 200 empty; A4 201 (HS-6 ADR-governed); A5/P2-A/P2-B 200; F1.1 200/201 unchanged. |
| Regression against M0 (authz/session corpus) | PASS | `go test -count=1 ./...` 0 FAIL; M0 packages `iam/*`, `auth/*`, `iam/authz` all `ok`. |
| Whole-repo build clean | PASS | `go build ./...` exit 0. |

**C4 verdict: PASS.** M0 regression confirmed clean (no prior-milestone gate flipped).

---

## C5 — Regression against all prior milestones

```
$ go build ./...
(exit 0)

$ go test -count=1 ./...
ok  metaldocs/apps/api/cmd/metaldocs-api          1.451s
ok  metaldocs/apps/api/internal/wiring            6.983s
ok  metaldocs/apps/worker/cmd/metaldocs-worker    0.640s
... (80 packages, all ok)
ok  metaldocs/tests/docx_v2                       1.378s

$ grep -c "FAIL" full-run.log
0
```

M0 (HS-1 approved per memory) — authz/session corpus packages all `ok`. No regression introduced by M1.

**C5 verdict: PASS.**

---

## C6 — Quality-bar re-measure (root-cause vs symptom)

### Quality bar declared by milestone.md

> "**Bar:** for each defect, a regression/contract test fails before and passes after the fix;
> the report §6 grep commands return **0** for the H-D class at milestone close; FE codegen
> regen is clean; no public route emits raw domain or `map[string]any`. The fix lands at the
> contract surface (handler returns the generated type), not via downstream FE shims or
> stringly typed translators."

### Re-measurement against F5.1 §6 method

The F5.1 §6 H-D class re-measurement enumerated **4** drift sites (DRIFT-1..4) using the cited grep commands as discovery tools, then classified each surfaced site as drift-vs-not-drift against the OpenAPI declaration. The method counted drift, not raw grep hits. All 4 DRIFT sites are now resolved at root cause (typed structs at handler boundary, OpenAPI-amended where the schema disagreed — F1.2 P2-light expansion):

| DRIFT | Root-cause fix | Symptom-patch check |
|-------|---------------|--------------------|
| DRIFT-1 (createTemplate id/version_id leak) | typed `CreateTemplateResponse` via mappers (F1.3) | not silenced, not deleted handler — actual contract emitted |
| DRIFT-2 (presignAutosave 201 envelope) | 200 + typed `TemplatePresignAutosaveResponse` (F1.2) | OpenAPI amended in same change; FE adapter rewired same-commit |
| DRIFT-3 (createTemplateVersion 201 envelope) | 201 + typed `VersionDTO` (F1.2; HS-6 ADR-governed) | ADR 0035 records the durable decision |
| DRIFT-4 (taxonomy profiles raw domain) | typed `DocumentProfileItem` via `toDocumentProfileItem` mapper (F1.4); zero-fill for 5 OpenAPI-required fields missing from domain | D-1 bounded defer for domain extension (not HS-2 — no API redesign) |

### Strict-literal grep interpretation (broader bar)

The milestone Objective also reads: "no public route emits raw `domain.*` or `map[string]any`". A strict-literal raw grep of `map[string]any` across `internal/modules/*/delivery/http/*.go` still produces hits in non-M1 sites (audit list/ingest/aggregate, templates lifecycle/query/schema/catalog, search, security, taxonomy areas/families, iam admin/memberships/observability, controlleddocuments, documents fillin/placeholder_options/pdf_webhook/view).

**Resolution:** the milestone Features table scopes M1 to specific sites: F1.1 = checkpoints (A1); F1.2 = 5 templates+documents sites; F1.3 = createTemplate (A3); F1.4 = "documents module raw-domain / `map[string]any` sweep at `documents/delivery/http/handler.go:816(+5)` (A6) and `:317` (A8); taxonomy at `routes_profiles.go:67,111,126,169` (A7); `Allow` header on `audit/delivery/http/handler.go:81` 405 (A8)." F1.4 evidence honestly labels V1 as "GREEN (scoped)" and explicitly enumerates the out-of-scope remaining hits as D-3 (audit other endpoints) — those endpoints were never part of the M1 spec sweep.

The mission §5 H-D class enumeration A1–A8 maps exactly to the four features' scope. Sites like `iam observability`, `security`, `controlleddocuments`, `search`, `audit list/ingest/aggregate`, `templates lifecycle/query/schema/catalog` are not in A1–A8 and were not scoped into M1. They are scoped to later milestones / future features per the mission's M2..M5 plan and the F1.4 D-3 defer.

This is the **F5.1-method bar** the milestone was authored against (per `milestone.md` lines 33–34, 52 — "mission report §6 grep commands return 0"). The §6 grep commands were the discovery commands; §6's class re-measurement step produced the **count = 4** that the bar is measured against. That count is now 0.

The strict-literal interpretation is the program-level Grade-A bar (M5 terminal acceptance), not the M1 close bar. M1 closes its named scope; remaining `map[string]any` hits are downstream milestone work.

### Bounded defers audit

| Feature | Defer | Named | Bounded | Trigger documented |
|---------|-------|-------|---------|-------------------|
| F1.1 | (none formal — `restoreCheckpoint` rolled into F1.4) | ✓ | ✓ | F1.4 implementation start |
| F1.2 | D-1 vitest harness broken (env) | ✓ | ✓ | later FE/devx task |
| F1.2 | D-2 FE adapter unit tests | ✓ | ✓ | alongside D-1 |
| F1.2 | D-3 `toVersionResponse` retirement | ✓ | ✓ | F1.3 territory per ADR 0035 ledger |
| F1.2 | D-4 `placeholder_schema` legacy FE type | ✓ | ✓ | focused FE follow-up |
| F1.3 | D-1 zero-timestamp test fixture | ✓ | ✓ | test-quality follow-up |
| F1.3 | D-2 declared-key set coverage | ✓ | ✓ | F1.4 or dedicated shape-test hardening |
| F1.3 | D-3 toVersionResponse/toTemplateResponse retirement | ✓ | ✓ | F1.4 territory (HS-6) |
| F1.4 | D-1 taxonomy `DocumentProfileItem` semantic gap (5 fields zero-filled) | ✓ | ✓ | when taxonomy domain model is extended |
| F1.4 | D-2 `map[string]string` sites | ✓ | ✓ | grep-coverage extension follow-up |
| F1.4 | D-3 audit module other endpoints | ✓ | ✓ | separate feature to typed-ify audit list/ingest/aggregate |
| F1.4 | D-4 drive-by `setDefaultTemplate` (resolved this PR) | ✓ | ✓ | resolved in-flight |
| F1.4 | D-5 test framework formalization | ✓ | ✓ | ADR when scaffolded |

All defers are (a) named, (b) bounded by file/feature, (c) trigger-documented. None block M1 close.

### Retrospective ("could it be built better")

- **D-1 taxonomy semantic gap (5 zero-filled fields)** is the most material concern: the H-D class is closed (no raw domain on the wire), but the typed response now lies — `ActiveSchemaVersion`, `ApprovalRequired`, `RetentionDays`, `ValidityDays`, `WorkflowProfile` are always `0/false/""` because they don't exist on the domain. A more thorough construction would extend the domain model + DB schema first (separate ticket per F1.4 interview Q7) and only then close the H-D site. The F1.4 spec explicitly chose to close H-D class first and defer the semantic gap (with operator agreement per interview Q7), justified as "not HS-2 — no shared API redesign". This is defensible and recorded — does not FAIL M1 on its own.
- **Strict-literal grep gap** (the remaining `map[string]any` in out-of-scope modules) is acknowledged correctly as scoped-out. A better construction would have authored M1 scope to enumerate every remaining site, but the F5.1 §6 method was the agreed-upon discovery instrument, and only A1–A8 were skeptic-confirmed at audit time. New sites discovered now are downstream-milestone scope.

**C6 verdict: PASS.** Root-cause closed at the F5.1 §6 measurement method; symptom-patch checks clean (no deleted handler, no silenced grep, no muted test). Defers properly bounded.

---

## C7 — Forbidden-list scan

| Forbidden pattern | Result | Evidence |
|-------------------|--------|----------|
| `--no-verify` / skipped hooks | CLEAN | `git log --pretty=full 81ce1815..HEAD` — no skipped-hook markers, no `-c commit.gpgsign=false`, all commits signed with standard author. |
| `t.Skip` / `t.SkipNow` added | CLEAN | `git diff 81ce1815..HEAD` shows no `+t.Skip` additions in any Go file. |
| Build-tag `// +build ignore` / `//go:build ignore` added | CLEAN | No build-ignore additions in milestone diff. |
| Silenced lints `//nolint` without justification | CLEAN | No `//nolint` additions in milestone diff. |
| Symptom patches (catch error → return success without acting) | CLEAN | All 500-on-UUID-parse paths return the error via `writeErr` / `writeMappedErr` — no swallow-and-200 patterns introduced. |
| Suite-green-as-pass without per-feature mapping | CLEAN | Each feature's evidence maps Validation Gate row-for-row to named tests; aggregate suite-green is supplementary, not the primary acceptance signal. |
| Fixture-as-real proof | CLEAN | Each evidence file explicitly labels Real vs fixture (F1.1 §"Real vs fixture", F1.2 V column, F1.3 V column, F1.4 V column). Handler-level fixtures used appropriately for wire-shape proofs; build/test/grep results labeled real. |
| Guessed contract | CLEAN | Every feature spec cites the consumer source-of-truth file:line (`api.gen.go`, FE adapter, OpenAPI YAML). No spec reads "designed from memory". |
| Split-brain (one fact two sources of truth) | CLEAN | Single mapper per output type (`toAPICheckpoint`, `toAPIVersionDTO`, `toAPITemplateDTO`, `toDocumentProfileItem`); OpenAPI single source for schema; ADR 0035 documents the durable design decisions. |
| Self-judged close | CLEAN | This verdict is written by an independent subagent (validator), not the implementing session. The implementing session has not yet flipped milestone status (per workflow: validator PASS → main session flips). |
| Scope drift | CLEAN | P2-light F1.2 expansion was operator-approved and reflected in milestone.md row swap; drive-by F1.4 fix surgical (1 line, CLAUDE.md §5.3 compliant); no work delivered outside the four-feature scope. |

**C7 verdict: CLEAN — no forbidden-list hits.**

---

## Verdict reasoning

Milestone 1 — contract-integrity passes both the code-wise dimension (senior-grade, contract-clean, no split-brain, no dead code, no guessed contracts) and the function-wise/QA dimension (every Validation Gate per feature is GREEN on isolated re-run; the F5.1 §6 H-D class re-measurement count goes from 4 → 0; full-repo regression including M0 authz/session corpus is clean; FE codegen regen is clean and consumer-facing). The HS-6 deviation (A4 keeps 201) is operator-approved and ADR-governed. The strict-literal raw-grep gap (`map[string]any` still present in out-of-scope modules like audit list/ingest/aggregate, templates lifecycle/query/schema/catalog, security, iam admin/observability, controlleddocuments) is correctly scoped out of M1 — A1–A8 from mission §5 are the M1 scope, F5.1 §6 produced the H-D drift count = 4 as the binding measurement, and the remaining hits are recorded as F1.4 D-3 / future-milestone work. The five-field zero-fill in `DocumentProfileItem` (F1.4 D-1) closes the H-D wire-shape defect but defers the semantic data gap to a domain-extension ticket; this is defensible per F1.4 interview Q7 and operator-aligned.

Bounded defers (13 total across the 4 features) are all named, bounded by file/feature, and have documented triggers. None block M1 close. Forbidden-list scan is clean across all six commits in the milestone diff (`81ce1815..7381b9c4`).

The milestone-validator gate is independent from the implementing session per the binding skill contract. This verdict only judges; it does not edit source, does not flip milestone/program status, and does not fix findings — those are main-session actions on PASS.

**VERDICT: PASS** — main session may flip M1 status to `closed` and present the HS-1 operator gate.

## On FAIL: named fix feature

(N/A — verdict is PASS.)
