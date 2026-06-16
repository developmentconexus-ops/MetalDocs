# Feature F3.1 — wire-iam-user-options

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Folder:** `f3.1-wire-iam-user-options`
> **Status:** Planning

This is the feature's **execution plan** — the "how" that `milestone.md` deliberately
left out. Produced by `superpowers:writing-plans`.

## Source

- Milestone spec row: **F3.1 — wire-iam-user-options.** *What:* wire existing `IAMUserOptions` dep at composition root `apps/api/cmd/metaldocs-api/main.go` so `internal/modules/documents/module.go:96` `newPlaceholderOptionsIAMAdapter` is fed a non-nil reader backed by IAM's user-options query. *Validate:* integration test placeholder-options user-type returns non-empty list; empty-tenant case returns empty without nil-deref; wire grep shows the line; `go test ./...` green.
- Governing-spec reference: mission §5 finding **E1** (`cmd/metaldocs-api/main.go:413` — path drifted; real anchor `apps/api/cmd/metaldocs-api/main.go:414`).
- Feature spec (binding consumer contract + Validation Gate): `./spec.md` (approved 2026-06-16).

## Plan

# F3.1 wire-iam-user-options Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix mission §5 E1 — the documents-module placeholder catalog returns real IAM user options for the `user`-type placeholder picker instead of a silent empty list.

**Architecture:** New adapter `wiring.DocumentsIAMUserOptions` in `apps/api/internal/wiring/` (Composition Root / Ports & Adapters sibling of `wiring.NewMfaCoveragePctReader`, `NewControlledDocumentDuplicator`, `NewProfileDefaults`) implements the consumer-defined port `documents.application.IAMUserOptionsReader` by wrapping the existing `auth.Service.ListUsers(ctx, tenantID)`. Adapter responsibilities: filter `IsActive`, map to `UserOption{UserID, DisplayName}`, sort by lower-case DisplayName ASC tie-break UserID ASC, return non-nil empty slice on no results. Composition root gains a single line in the `docDeps` literal.

**Tech Stack:** Go 1.25, `metaldocs/internal/modules/auth` (existing tenant-scoped `ListUsers`), `metaldocs/internal/modules/documents/application` (consumer port). Standard library only: `context`, `sort`, `strings`. No new dependencies.

---

### Task 1: Create failing unit tests for the new adapter

**Files:**
- Create: `apps/api/internal/wiring/iam_user_options_test.go`

Test scope = spec.md Validation Gate rows 1–5 (active-filter, sort, non-nil-empty, error-propagate, tenant-isolation). Follows CLAUDE.md §4 framework: **table-driven Go test + in-memory fake `authListUsersFunc`**; **UUID-shaped** `UserID` fixtures via `uuid.New().String()`; deterministic seeded `DisplayName`s. No sloppy strings like `"user_1"`.

- [ ] **Step 1: Write the failing test file**

```go
package wiring

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	authdomain "metaldocs/internal/modules/auth/domain"
	docapp "metaldocs/internal/modules/documents/application"
)

// authListUsersFunc is the narrow function-typed seam the adapter consumes.
// In production it is satisfied by *auth.Service.ListUsers. The test stands a
// trivial fake against it; no sloppy strings, UUIDs for UserID per CLAUDE.md §4.
type authListUsersFunc func(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)

func (f authListUsersFunc) ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error) {
	return f(ctx, tenantID)
}

func newManagedUser(displayName string, isActive bool) authdomain.ManagedUser {
	return authdomain.ManagedUser{
		UserID:      uuid.New().String(),
		DisplayName: displayName,
		IsActive:    isActive,
	}
}

func TestDocumentsIAMUserOptions(t *testing.T) {
	t.Parallel()

	tenantA := uuid.New().String()

	zoeActive := newManagedUser("Zoe", true)
	aliceActive := newManagedUser("alice", true) // lower-case to exercise case-insensitive sort
	bobInactive := newManagedUser("Bob", false)
	mikeActive := newManagedUser("Mike", true)
	mikeTwinActive := authdomain.ManagedUser{
		// Same DisplayName as mikeActive to exercise UserID tie-break.
		// Force UserID lexicographically smaller than mikeActive.UserID.
		UserID:      "00000000-0000-0000-0000-000000000001",
		DisplayName: "Mike",
		IsActive:    true,
	}

	sentinelErr := errors.New("auth boom")

	cases := []struct {
		name        string
		tenantID    string
		fake        authListUsersFunc
		wantUserIDs []string // order-significant; UserID is unique
		wantErr     error
		assertNotNil bool // empty result must still be a non-nil slice
	}{
		{
			name:     "filters inactive — Bob dropped",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return []authdomain.ManagedUser{zoeActive, bobInactive, aliceActive}, nil
			},
			wantUserIDs: []string{aliceActive.UserID, zoeActive.UserID}, // case-insensitive: alice < Zoe
		},
		{
			name:     "sorts case-insensitive ASC, tie-break by UserID ASC",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return []authdomain.ManagedUser{mikeActive, aliceActive, mikeTwinActive}, nil
			},
			// alice < Mike; among the two "Mike"s, the lex-smaller UUID wins.
			wantUserIDs: []string{aliceActive.UserID, mikeTwinActive.UserID, mikeActive.UserID},
		},
		{
			name:     "empty result returns non-nil empty slice",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return nil, nil
			},
			wantUserIDs:  []string{},
			assertNotNil: true,
		},
		{
			name:     "propagates underlying error and returns nil slice",
			tenantID: tenantA,
			fake: func(_ context.Context, _ string) ([]authdomain.ManagedUser, error) {
				return nil, sentinelErr
			},
			wantErr: sentinelErr,
		},
		{
			name:     "forwards tenantID verbatim (tenant isolation)",
			tenantID: tenantA,
			fake: func(_ context.Context, gotTenant string) ([]authdomain.ManagedUser, error) {
				if gotTenant != tenantA {
					t.Fatalf("adapter did not forward tenantID: got %q want %q", gotTenant, tenantA)
				}
				return []authdomain.ManagedUser{aliceActive}, nil
			},
			wantUserIDs: []string{aliceActive.UserID},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewDocumentsIAMUserOptions(tc.fake)

			got, err := adapter.ListUserOptions(context.Background(), tc.tenantID)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err: got %v want %v", err, tc.wantErr)
				}
				if got != nil {
					t.Fatalf("on error want nil slice, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if tc.assertNotNil && got == nil {
				t.Fatalf("empty result must be non-nil []UserOption{}, got nil")
			}

			if len(got) != len(tc.wantUserIDs) {
				t.Fatalf("len: got %d (%+v) want %d (%v)", len(got), got, len(tc.wantUserIDs), tc.wantUserIDs)
			}
			for i, want := range tc.wantUserIDs {
				if got[i].UserID != want {
					t.Errorf("idx %d UserID: got %q want %q", i, got[i].UserID, want)
				}
				if got[i].DisplayName == "" {
					t.Errorf("idx %d DisplayName empty", i)
				}
			}

			// Type guard: result must satisfy the consumer port.
			var _ docapp.IAMUserOptionsReader = adapter
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run from repo root:
```
go test ./apps/api/internal/wiring/ -run TestDocumentsIAMUserOptions -v
```

Expected: build error `undefined: NewDocumentsIAMUserOptions` (and `undefined: docapp.IAMUserOptionsReader` only if the import wasn't resolved — it will be).

- [ ] **Step 3: Commit failing test**

```
git add apps/api/internal/wiring/iam_user_options_test.go
git commit -m "test(f3.1): failing tests for wiring.DocumentsIAMUserOptions adapter"
```

---

### Task 2: Implement the adapter to green

**Files:**
- Create: `apps/api/internal/wiring/iam_user_options.go`

Adapter pattern matches `apps/api/internal/wiring/iam_adapters.go:13-32` (narrow consumer-typed interface seam, value-receiver methods, nil-source returns empty/zero). Use a function-typed seam (`authListUsersFunc` from the test) so the production caller passes `authService.ListUsers` directly without an interface assertion.

- [ ] **Step 1: Write the adapter file**

```go
package wiring

import (
	"context"
	"sort"
	"strings"

	authdomain "metaldocs/internal/modules/auth/domain"
	docapp "metaldocs/internal/modules/documents/application"
)

// authUserLister is the narrow function-typed seam the documents-side adapter
// consumes. *auth.Service.ListUsers satisfies it directly — pass the method
// value. Keeping the seam this thin (a single function, not a multi-method
// interface) avoids hauling the entire auth.Service surface into the documents
// composition path and keeps the unit test fake trivial.
type authUserLister interface {
	ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)
}

// DocumentsIAMUserOptions adapts an auth user-lister to
// documents.application.IAMUserOptionsReader. It is the production adapter for
// the consumer-defined port at internal/modules/documents/application/iam_user_options.go.
//
// Semantics (binding — see f3.1 spec.md):
//   - Filters auth.ManagedUser.IsActive == true (deactivated users dropped).
//   - Maps to UserOption{UserID, DisplayName}.
//   - Sorts by strings.ToLower(DisplayName) ASC; tie-break by UserID ASC.
//   - Returns a non-nil empty slice when no users qualify.
//   - Propagates the underlying error unchanged on failure.
type DocumentsIAMUserOptions struct {
	lister authUserLister
}

// NewDocumentsIAMUserOptions returns the production
// documents.application.IAMUserOptionsReader. Panics if lister is nil — the
// composition root is the only caller and must wire a real source.
func NewDocumentsIAMUserOptions(lister authUserLister) *DocumentsIAMUserOptions {
	if lister == nil {
		panic("wiring: NewDocumentsIAMUserOptions: lister is nil")
	}
	return &DocumentsIAMUserOptions{lister: lister}
}

// ListUserOptions implements docapp.IAMUserOptionsReader.
func (a *DocumentsIAMUserOptions) ListUserOptions(ctx context.Context, tenantID string) ([]docapp.UserOption, error) {
	users, err := a.lister.ListUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]docapp.UserOption, 0, len(users))
	for _, u := range users {
		if !u.IsActive {
			continue
		}
		out = append(out, docapp.UserOption{
			UserID:      u.UserID,
			DisplayName: u.DisplayName,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		li := strings.ToLower(out[i].DisplayName)
		lj := strings.ToLower(out[j].DisplayName)
		if li != lj {
			return li < lj
		}
		return out[i].UserID < out[j].UserID
	})
	return out, nil
}
```

- [ ] **Step 2: Run unit tests to verify they pass**

```
go test ./apps/api/internal/wiring/ -run TestDocumentsIAMUserOptions -v
```

Expected: PASS — all 5 subtests green.

- [ ] **Step 3: Run the wiring package full suite (regression check on existing wiring tests)**

```
go test ./apps/api/internal/wiring/...
```

Expected: PASS (existing `documents_adapters_test.go` etc. unaffected).

- [ ] **Step 4: Commit the adapter**

```
git add apps/api/internal/wiring/iam_user_options.go
git commit -m "feat(f3.1): wiring.DocumentsIAMUserOptions adapter for documents IAMUserOptionsReader port"
```

---

### Task 3: Wire the adapter into the composition root

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go:414` (`docDeps` literal) — add ONE line.

The `docDeps` literal currently ends with `DisplayNameReader: displayNameRepo,` at line 428. Add `IAMUserOptions: wiring.NewDocumentsIAMUserOptions(authService),` adjacent. `authService` is already in scope from line 180. Surgical change — no reordering, no extra wiring, no log line.

- [ ] **Step 1: Add the wire line**

In `apps/api/cmd/metaldocs-api/main.go`, locate the `docDeps := documents.Dependencies{...}` block at line 414. Add this single field — adjacent to `DisplayNameReader`:

```go
		IAMUserOptions:               wiring.NewDocumentsIAMUserOptions(authService),
```

Field alignment matches existing tab stops in the literal (Go fmt will normalize).

- [ ] **Step 2: Build to verify the wire compiles**

```
go build ./apps/api/cmd/metaldocs-api/
```

Expected: clean exit (no output).

- [ ] **Step 3: Run focused regression suites**

```
go test ./internal/modules/documents/... ./internal/modules/auth/... ./apps/api/internal/wiring/...
```

Expected: PASS.

- [ ] **Step 4: Wire-line grep proof**

```
grep -RIn "IAMUserOptions" apps/api/cmd/metaldocs-api/
```

Expected: at least one match showing `IAMUserOptions:` paired with `wiring.NewDocumentsIAMUserOptions(authService)`. Capture the output for `evidence.md` (Validation Gate row 7).

- [ ] **Step 5: Commit the wiring**

```
git add apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(f3.1): wire DocumentsIAMUserOptions into documents composition root (mission §5 E1)"
```

---

### Task 4: Whole-repo regression + M0/M1/M2 sentinel re-check

**Files:** none modified. Validation Gate row 8.

- [ ] **Step 1: Whole-repo test run**

```
go test ./...
```

Expected: PASS. If any pre-existing flaky test fails, re-run that single package; only F3.1-introduced failures are blockers.

- [ ] **Step 2: M1 H-D sentinel re-grep (validates contract-class stayed at 0)**

Use the report §6 H-D class grep set. From repo root:

```
grep -RInE "map\[string\]any\s*\{" internal/modules/documents/delivery/ internal/modules/templates/delivery/ internal/modules/taxonomy/delivery/ internal/modules/audit/delivery/ 2>&1 | grep -v _test.go
```

Expected: 0 lines (M1 baseline). Capture output for `evidence.md`. If non-zero, HS-3 — investigate before continuing.

- [ ] **Step 3: M2 composition-root sentinel re-grep**

```
grep -RIn "NewTextHandler" internal/modules/jobs/
```

Expected: 0 lines (M2 baseline).

- [ ] **Step 4: Commit regression evidence — none (no source change)**

No commit. Evidence captured to `evidence.md` in Task 6.

---

### Task 5: Runtime smoke (real provider — Validation Gate row 6)

**Files:** none modified. Captures runtime artifact for `evidence.md`.

This is the **labeled real** proof per mission §8. Requires the dev API + seeded tenant.

- [ ] **Step 1: Start the API on :8081**

```
.\scripts\start-api.ps1
```

(PowerShell — canonical entrypoint per CLAUDE.md §1. Do NOT use bash/source.) Wait until the script reports `listening on :8081`.

- [ ] **Step 2: Login as admin and capture the session cookie**

In PowerShell:

```powershell
$resp = Invoke-WebRequest -Uri http://localhost:8081/api/v1/auth/login `
  -Method POST -ContentType 'application/json' `
  -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}' `
  -SessionVariable sess
$resp.StatusCode  # expect 200
```

- [ ] **Step 3: Pick a seeded document + user-type placeholder pid**

In the dev tenant, list documents and choose one whose latest revision has at least one `user`-type placeholder. If the seed catalog already exposes a known doc, use it; otherwise:

```powershell
Invoke-RestMethod -Uri http://localhost:8081/api/v1/documents -WebSession $sess | Select-Object -First 1
```

Note `$docID` and pick a placeholder `pid` from its template schema. Record both in `evidence.md`.

- [ ] **Step 4: Call placeholder-options for the user-type placeholder**

```powershell
$opts = Invoke-RestMethod `
  -Uri "http://localhost:8081/api/v1/documents/$docID/placeholder-options/$pid" `
  -WebSession $sess
$opts | ConvertTo-Json -Depth 4
```

Expected: JSON body — non-empty array of `{user_id, display_name}` objects; sorted by display_name ASC (case-insensitive); no entry with `is_active=false` source.

- [ ] **Step 5: Negative case — empty-tenant equivalent**

If the dev seed includes a tenant with zero active users, repeat with that tenant context (impersonation / session). If not feasible, document the gap in `evidence.md` and rely on Task 1 Step 1's unit-test row for empty-result coverage; the unit test IS the contract for the nil-vs-empty semantics.

- [ ] **Step 6: Stop the API**

```powershell
Get-Process metaldocs-api -ErrorAction SilentlyContinue | Stop-Process
```

- [ ] **Step 7: Paste captured JSON + commands into `evidence.md`** (Task 6) **labeled `real` per mission §8.**

No commit yet — evidence file is created in Task 6.

---

### Task 6: Write `evidence.md` (close-out proof)

**Files:**
- Create: `docs/superpowers/milestones/grade-a-completion/milestone-3-quality-tail/f3.1-wire-iam-user-options/evidence.md`

Follows the milestone skill `feature-evidence.md` template. One row per Validation Gate acceptance criterion (1–9 in `spec.md`).

- [ ] **Step 1: Copy the template + fill rows**

Use `.claude/skills/milestone/templates/feature-evidence.md` as the shape. Required rows mapped to spec acceptance:

1. **Row 1–5 (unit-test fixture):** paste `go test ./apps/api/internal/wiring/ -run TestDocumentsIAMUserOptions -v` real output (PASS).
2. **Row 6 (runtime real):** paste Task 5 commands + redacted JSON body. **Label `real`** explicitly.
3. **Row 7 (wire grep):** paste Task 3 Step 4 output.
4. **Row 8 (regression):** paste `go test ./...` summary + the two sentinel greps (H-D, NewTextHandler) returning 0.
5. **Row 9 (authz-scope check):** `N/A — F3.1 adds/removes no parameter.`
6. **TDD proof:** Task 1 commit SHA shows failing-test commit precedes Task 2 implementation commit.
7. **Review/QA disposition:** filled after `code-review` runs (Task 7).
8. **Bounded defers:** none expected. If any surface (e.g. empty-tenant runtime gap), record with trigger + owner.

- [ ] **Step 2: Commit evidence (partial — review row may be empty pending Task 7)**

```
git add docs/superpowers/milestones/grade-a-completion/milestone-3-quality-tail/f3.1-wire-iam-user-options/evidence.md
git commit -m "docs(f3.1): evidence — unit + runtime real proof; regression sentinels green"
```

---

### Task 7: Code review + fix-by-family

**Files:** any flagged by review.

Per CLAUDE.md §4 default close-out loop step 3.

- [ ] **Step 1: Run code-review on the F3.1 diff**

```
/code-review
```

Reviewer scope: the 3 commits since branching for F3.1 (failing test, adapter, wiring). Reviewer must flag — root-cause-vs-symptom, test-framework conformance (UUIDs, no sloppy strings, table-driven, in-memory fake), surgical-change rule, port-vs-adapter ownership.

- [ ] **Step 2: Classify findings by root-cause family + fix**

Per CLAUDE.md §4 default close-out loop step 5: fix by family, not scattered symptom patching. Common F3.1 families likely empty (small feature). If any surface (e.g. sort-stability concern, panic-vs-error choice), address at the right seam.

- [ ] **Step 3: Re-run unit + targeted regression**

```
go test ./apps/api/internal/wiring/ ./internal/modules/documents/... ./internal/modules/auth/...
```

Expected: PASS.

- [ ] **Step 4: Update `evidence.md` review/QA row with findings + disposition**

```
git add docs/superpowers/milestones/grade-a-completion/milestone-3-quality-tail/f3.1-wire-iam-user-options/evidence.md <fix files if any>
git commit -m "fix(f3.1): address code-review findings — <one-line family summary>"
```

If no fixes required: amend the evidence commit with `Review: clean, no findings.` (per CLAUDE.md §5.0 standing commit auth, no push).

---

## Files touched (close-out summary)

| Action | Path | Why |
|--------|------|-----|
| Create | `apps/api/internal/wiring/iam_user_options.go` | Adapter satisfying `documents.application.IAMUserOptionsReader` |
| Create | `apps/api/internal/wiring/iam_user_options_test.go` | Table-driven unit tests (canonical framework per CLAUDE.md §4) |
| Modify | `apps/api/cmd/metaldocs-api/main.go:~429` | One-line `docDeps.IAMUserOptions = wiring.NewDocumentsIAMUserOptions(authService)` |
| Create | `docs/superpowers/milestones/grade-a-completion/milestone-3-quality-tail/f3.1-wire-iam-user-options/evidence.md` | Close-out proof per milestone skill |

No other file should appear in the F3.1 diff. Anything else = scope drift (HS-6).

## Self-review

**1. Spec coverage:** all 9 Validation Gate rows mapped to a task — rows 1–5 (Task 1 + Task 2), row 6 (Task 5), row 7 (Task 3 Step 4 + Task 6 Row 7), row 8 (Task 4 + Task 6 Row 8), row 9 (Task 6 Row 9 — recorded N/A). Interview Q&A captured in `spec.md`. ADR check recorded as not-needed.

**2. Placeholder scan:** no "TBD"/"TODO"/"implement later"; every code step has the full code body; every command is exact.

**3. Type consistency:** `NewDocumentsIAMUserOptions(lister authUserLister) *DocumentsIAMUserOptions` is the same name + signature in test (Task 1), implementation (Task 2), wiring (Task 3). `authUserLister` interface with single method `ListUsers(ctx, tenantID) ([]authdomain.ManagedUser, error)` is consistent. `docapp.UserOption{UserID, DisplayName}` matches `internal/modules/documents/application/iam_user_options.go:5`.

**4. Test-framework conformance (CLAUDE.md §4):** Task 1 is table-driven + in-memory fake (function-typed `authListUsersFunc` satisfies the production interface); UUIDs via `uuid.New().String()` for `UserID` and `tenantA`; deterministic seeded `DisplayName`s; no sloppy strings (`"user_1"`, `"tenant-a"`). Conforms to test-framework hard gate.

**5. Surgical change (CLAUDE.md §5.3):** every changed line in Task 3 traces directly to mission §5 E1; no drive-by edits in `main.go` or `module.go`; the nil-safe branch at `module.go:144` stays as defense-in-depth (not removed).

## Execution notes

(Empty — filled by `superpowers:subagent-driven-development` as it runs.)
