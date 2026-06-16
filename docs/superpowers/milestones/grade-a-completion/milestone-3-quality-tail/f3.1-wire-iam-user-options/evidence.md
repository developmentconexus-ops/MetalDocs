# Feature F3.1 — Evidence

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.1-wire-iam-user-options`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

- Created `apps/api/internal/wiring/iam_user_options.go` — new `DocumentsIAMUserOptions` adapter struct implementing `documents.application.IAMUserOptionsReader` by wrapping a narrow `authUserLister` interface (single method: `ListUsers`). Filters `IsActive == true`, maps to `UserOption{UserID, DisplayName}`, sorts by `strings.ToLower(DisplayName)` ASC / `UserID` ASC tie-break, returns non-nil empty slice on no results, propagates error unchanged.
- Created `apps/api/internal/wiring/iam_user_options_test.go` — table-driven unit tests (5 subtests) covering all spec Validation Gate rows 1–5. TDD: failing test committed first, then implementation made it green.
- Modified `apps/api/cmd/metaldocs-api/main.go:429` — added one line to `docDeps` literal: `IAMUserOptions: wiring.NewDocumentsIAMUserOptions(authService)`. Fixes mission §5 E1.

Commits:
- `d3bca7ae test(f3.1): failing tests for wiring.DocumentsIAMUserOptions adapter`
- `5588a1f6 test(f3.1): strengthen DisplayName assertion in TestDocumentsIAMUserOptions`
- `90992099 feat(f3.1): wiring.DocumentsIAMUserOptions adapter for documents IAMUserOptionsReader port`
- `e91d1daa feat(f3.1): wire DocumentsIAMUserOptions into documents composition root (mission §5 E1)`

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first | `go test ./apps/api/internal/wiring/ -run TestDocumentsIAMUserOptions -v` on commit `d3bca7ae` | `undefined: NewDocumentsIAMUserOptions` build error — red phase confirmed | fixture |
| TDD — green after implementation | Same command on commit `90992099` | All 5 subtests PASS (see unit test output below) | fixture |
| Static (build) | `go build ./apps/api/cmd/metaldocs-api/` | `BUILD OK` — clean exit, no output | — |
| Targeted test (wiring package regression) | `go test ./apps/api/internal/wiring/...` | PASS — all wiring tests green including pre-existing `documents_adapters_test.go` | fixture |
| Runtime proof — user-type placeholder returns real IAM users | `POST /api/v1/auth/login` → `GET /api/v1/documents/$docID/placeholder-options/$pid` | **REAL_PROVIDER** — 4 active users, sorted ASC by display_name (see runtime output below) | **real** |

### Unit test output (rows 1–5)

```
=== RUN   TestDocumentsIAMUserOptions
=== CONT  TestDocumentsIAMUserOptions/filters_inactive_—_Bob_dropped
=== CONT  TestDocumentsIAMUserOptions/sorts_case-insensitive_ASC,_tie-break_by_UserID_ASC
=== CONT  TestDocumentsIAMUserOptions/empty_result_returns_non-nil_empty_slice
=== CONT  TestDocumentsIAMUserOptions/propagates_underlying_error_and_returns_nil_slice
=== CONT  TestDocumentsIAMUserOptions/forwards_tenantID_verbatim_(tenant_isolation)
--- PASS: TestDocumentsIAMUserOptions (0.00s)
    --- PASS: TestDocumentsIAMUserOptions/filters_inactive_—_Bob_dropped (0.00s)
    --- PASS: TestDocumentsIAMUserOptions/sorts_case-insensitive_ASC,_tie-break_by_UserID_ASC (0.00s)
    --- PASS: TestDocumentsIAMUserOptions/empty_result_returns_non-nil_empty_slice (0.00s)
    --- PASS: TestDocumentsIAMUserOptions/propagates_underlying_error_and_returns_nil_slice (0.00s)
    --- PASS: TestDocumentsIAMUserOptions/forwards_tenantID_verbatim_(tenant_isolation) (0.00s)
PASS
ok  metaldocs/apps/api/internal/wiring
```

### Runtime output (row 6) — labeled `real`

Provider: live dev API + seeded admin tenant. No user-type placeholder existed in existing documents; smoke document inserted directly via SQL to exercise the handler path (identical code path to production).

Document: `f3110000-0000-0000-0000-000000000001` · Placeholder: `phuser1` (type: `user`, label: `Responsible`)

```json
{
  "options": [
    { "user_id": "admin",         "display_name": "Administrator" },
    { "user_id": "approver",      "display_name": "Approver Dev"  },
    { "user_id": "approver-test", "display_name": "Approver Test" },
    { "user_id": "author-test",   "display_name": "Author Test"   }
  ]
}
```

Non-empty: yes (4 active users). Sorted ASC by display_name (case-insensitive): Administrator < Approver Dev < Approver Test < Author Test. No inactive users present.

### Wire grep (row 7)

```
apps/api/cmd/metaldocs-api/main.go:429:		IAMUserOptions:               wiring.NewDocumentsIAMUserOptions(authService),
```

`authService` is the existing `*authapp.Service` declared at line 180 — not a new variable.

### Regression + sentinel checks (row 8)

`go test ./...` — **PASS** — 72 packages, all green (all cached; no F3.1-introduced failures).

M1 H-D sentinel (`grep -RInE "map\[string\]any\s*\{"` in delivery dirs, excluding `_test.go`):
- 40 matches — **all pre-existing** before F3.1. F3.1 added **zero** new instances. These are M1 technical-debt legacy patterns predating this feature; the sentinel confirms F3.1 introduced no regression.

M2 `NewTextHandler` sentinel (`grep -RIn "NewTextHandler" internal/modules/jobs/`):
- 0 matches — **PASS**

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| 1. Adapter filters `!IsActive` — deactivated users excluded | yes | `TestDocumentsIAMUserOptions/filters_inactive_—_Bob_dropped` PASS |
| 2. Adapter sorts by `strings.ToLower(DisplayName)` ASC, tie-break `UserID` ASC | yes | `TestDocumentsIAMUserOptions/sorts_case-insensitive_ASC,_tie-break_by_UserID_ASC` PASS |
| 3. Empty result returns non-nil empty slice | yes | `TestDocumentsIAMUserOptions/empty_result_returns_non-nil_empty_slice` PASS |
| 4. Underlying error propagates unchanged | yes | `TestDocumentsIAMUserOptions/propagates_underlying_error_and_returns_nil_slice` PASS |
| 5. Tenant isolation — tenantID forwarded verbatim | yes | `TestDocumentsIAMUserOptions/forwards_tenantID_verbatim_(tenant_isolation)` PASS |
| 6. E2E: placeholder-options returns real user list (labeled `real`) | yes | Runtime output above — 4 active users, sorted, non-empty — **real** provider |
| 7. Wiring proof — `IAMUserOptions` set in `docDeps` with `authService` | yes | Wire grep: `main.go:429` |
| 8. Whole-repo regression green; M1/M2 sentinels pass | yes | `go test ./...` PASS; M2 sentinel 0; M1 sentinel 40 pre-existing (F3.1 added 0) |
| 9. Authz-scope check | N/A | F3.1 adds no parameter and removes none. Recorded N/A. |

## Review disposition

- Spec-compliance review: **APPROVED** — all 10 adapter spec requirements met; all 6 wiring requirements met. No violations.
- Code-quality review (per-task): **APPROVED** — one nit (`sort.SliceStable` vs `sort.Slice`; both correct since tie-break by unique UserID makes stability irrelevant); one false-positive about UUID fixture reproducibility (wantUserIDs uses same variable references — internally stable). DisplayName assertion strengthened from non-empty check to exact match per review finding. No blocking issues.
- Final holistic code review (full F3.1 diff, base `22a80208` → head `d6a3d416`): **Ready to merge: Yes.** No Critical or Important issues. Two Minor/advisory improvements applied: (1) comment added in `ListUserOptions` clarifying that `auth.Service.ListUsers` tenant-filters by role-map but does NOT filter `IsActive` — that is the adapter's responsibility; (2) type guard moved from inside test loop to package scope (`var _ docapp.IAMUserOptionsReader = (*DocumentsIAMUserOptions)(nil)`) for idiomatic Go placement. Both applied and re-verified green.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| M1 H-D legacy `map[string]any` (40 matches in delivery dirs) | Pre-existing before F3.1; M1 technical debt; F3.1 added zero new instances | M1 scope — owned by M3+ cleanup roadmap |
| Runtime smoke via SQL-inserted fixture (no API-created doc with user placeholder in seed) | Seed data gap; handler code path is identical to production; unit tests cover contract exhaustively | Seed improvement tracked separately; not a F3.1 correctness gap |
