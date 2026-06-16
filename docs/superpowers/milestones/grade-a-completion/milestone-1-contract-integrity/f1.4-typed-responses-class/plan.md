# F1.4 — typed-responses class — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development (recommended)
> or superpowers:executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Drive H-D class to zero — replace all `map[string]any` / raw domain emissions in documents, taxonomy, and audit delivery handlers with typed OpenAPI-generated structs.

**Architecture:** Inline struct constructions in `documents/delivery/http/handler.go` (6 sites); a new `routes_mapping.go` + mapper in `taxonomy/delivery/http/`; a single `w.Header().Set` in `audit/delivery/http/handler.go`. No schema changes, no FE codegen regen — all typed structs already exist in each module's `api.gen.go`.

**Tech Stack:** Go `net/http`, oapi-codegen generated types (`documentsapi`, `taxonomyapi`, `auditapi`), `github.com/google/uuid`, `go test`.

---

### Task 1: RED tests — documents handler typed responses

**Files:**
- Create: `internal/modules/documents/delivery/http/handler_typed_response_test.go`

Write a test file with typed-response shape assertions for the 6 map[string]any sites. Each test:
1. Builds a minimal `Handler` with fake service (or uses existing test helpers)
2. Calls the endpoint
3. Asserts the top-level JSON keys match the declared OpenAPI type (no extra keys)

Check existing test helpers in `internal/modules/documents/delivery/http/` to understand what test infrastructure is available (mux, fake services, auth context) before writing.

- [ ] **Step 1: Read the existing test files** to understand helper patterns:

Read `internal/modules/documents/delivery/http/handler_test.go` (first 100 lines) to understand `newTestHandler` or equivalent setup pattern.

- [ ] **Step 2: Write `TestPresignDocumentAutosave_TypedResponseShape`:**

```go
// asserts response keys = {expires_at, pending_upload_id, upload_url} — no extras
var raw map[string]json.RawMessage
json.Unmarshal(body, &raw)
assertKeys(t, raw, []string{"expires_at", "pending_upload_id", "upload_url"})
```

- [ ] **Step 3: Write `TestCommitDocumentAutosave_TypedResponseShape`:**

```go
// asserts response has revision_id, revision_num; no map[string]any envelope
assertKeys(t, raw, []string{"revision_id", "revision_num"}) // required; optional keys allowed
```

- [ ] **Step 4: Write `TestRestoreDocumentCheckpoint_TypedResponseShape`:**

```go
// asserts response has new_revision_id, new_revision_num, idempotent
assertKeyPresent(t, raw, "new_revision_id")
assertKeyPresent(t, raw, "new_revision_num")
assertKeyPresent(t, raw, "idempotent")
```

- [ ] **Step 5: Write `TestListDocumentRevisionHistory_TypedResponseShape`:**

```go
// asserts response has "items" (no bare map envelope with extra keys)
assertKeyPresent(t, raw, "items")
```

- [ ] **Step 6: Run to confirm RED** (all 4 tests fail because handlers still emit map[string]any):

```
go test ./internal/modules/documents/delivery/http/... -run "TypedResponseShape" -v
```

Expected: all FAIL.

---

### Task 2: RED tests — taxonomy profiles + audit Allow header

**Files:**
- Create: `internal/modules/taxonomy/delivery/http/routes_profiles_typed_test.go`
- Create: `internal/modules/audit/delivery/http/handler_allow_test.go`

- [ ] **Step 1: Read existing taxonomy test helpers:**

Read `internal/modules/taxonomy/delivery/http/` directory listing and read an existing test file (e.g., `routes_profiles_test.go` if it exists, or any other) to understand setup.

- [ ] **Step 2: Write `TestListProfiles_TypedResponseShape`:**

```go
// asserts response has "items" key (from ListDocumentProfilesResponse)
// and items[0] has "code", "family_code", "name" keys
```

- [ ] **Step 3: Write `TestGetProfile_TypedResponseShape`:**

```go
// asserts response has "code", "name", "family_code" keys (DocumentProfileItem fields)
// and does NOT have "tenant_id", "owner_user_id", "editable_by_role" (domain-only fields)
```

- [ ] **Step 4: Write `TestHandleEvents_405_Allow`** in `audit/delivery/http/handler_allow_test.go`:

```go
// POST /audit/events (or whichever path) → 405
// assert response.Header.Get("Allow") == "GET"
```

Read `internal/modules/audit/delivery/http/handler_test.go` first to understand the test setup.

- [ ] **Step 5: Run to confirm RED:**

```
go test ./internal/modules/taxonomy/delivery/http/... -run "TypedResponseShape" -v
go test ./internal/modules/audit/delivery/http/... -run "Allow" -v
```

Expected: all FAIL.

---

### Task 3: Fix documents handler — 6 map[string]any sites + stats

**Files:**
- Modify: `internal/modules/documents/delivery/http/handler.go`

Add `documentsapi "metaldocs/internal/modules/documents/api"` to imports if not already present. Also ensure `"github.com/google/uuid"` is imported.

- [ ] **Step 1: Fix `getDocumentSession` readonly branch (line ~694):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
    "mode":       "readonly",
    "held_by":    sess.UserID,
    "held_until": sess.ExpiresAt,
})
```

With:
```go
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentSessionReadonlyResponse{
    Mode:      documentsapi.DocumentSessionReadonlyResponseMode("readonly"),
    HeldBy:    sess.UserID,
    HeldUntil: sess.ExpiresAt,
})
```

- [ ] **Step 2: Fix `getDocumentSession` writer branch (line ~701):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusCreated, map[string]any{
    "mode":                 "writer",
    "session_id":           sess.ID,
    "expires_at":           sess.ExpiresAt,
    "last_ack_revision_id": sess.LastAcknowledgedRevisionID,
})
```

With:
```go
sessID, err := uuid.Parse(sess.ID)
if err != nil {
    httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
    return
}
lastAckID, err := uuid.Parse(sess.LastAcknowledgedRevisionID)
if err != nil {
    httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
    return
}
httpresponse.WriteJSON(w, http.StatusCreated, documentsapi.DocumentSessionWriterResponse{
    Mode:              documentsapi.DocumentSessionWriterResponseMode("writer"),
    SessionId:         sessID,
    ExpiresAt:         sess.ExpiresAt,
    LastAckRevisionId: lastAckID,
})
```

- [ ] **Step 3: Fix `presignDocumentAutosave` (line ~814):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
    "upload_url":        res.UploadURL,
    "pending_upload_id": res.PendingUploadID,
    "expires_at":        res.ExpiresAt,
})
```

With:
```go
pendingID, err := uuid.Parse(res.PendingUploadID)
if err != nil {
    httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
    return
}
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentAutosavePresignResponse{
    UploadUrl:       res.UploadURL,
    PendingUploadId: pendingID,
    ExpiresAt:       res.ExpiresAt,
})
```

- [ ] **Step 4: Fix `commitDocumentAutosave` (line ~855):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
    "revision_id":       res.RevisionID,
    "revision_num":      res.RevisionNum,
    "idempotent_replay": res.AlreadyConsumed,
    "file_size_bytes":   res.FileSizeBytes,
    "page_count":        res.PageCount,
})
```

With:
```go
revID, err := uuid.Parse(res.RevisionID)
if err != nil {
    httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
    return
}
idempotentReplay := res.AlreadyConsumed
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.CommitDocumentAutosave200JSONResponse{
    RevisionId:       revID,
    RevisionNum:      res.RevisionNum,
    IdempotentReplay: &idempotentReplay,
    FileSizeBytes:    &res.FileSizeBytes,
    PageCount:        res.PageCount,
})
```

Note: verify exact field types of `res` (application result) at implementation time. If `FileSizeBytes` or `PageCount` are non-pointer in the result, take address to match the `*int64`/`*int` in the response struct.

- [ ] **Step 5: Fix `listRevisionHistory` (line ~903):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
    "items": toRevisionHistoryResponse(items),
})
```

With:
```go
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentRevisionHistoryResponse{
    Items: toRevisionHistoryResponse(items),
})
```

Note: `toRevisionHistoryResponse` already returns `[]documentsapi.DocumentRevisionHistoryItem` (verify at implementation time).

- [ ] **Step 6: Fix `restoreDocumentCheckpoint` (line ~1023):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
    "new_revision_id":               res.NewRevisionID,
    "new_revision_num":              res.NewRevisionNum,
    "source_checkpoint_version_num": versionNum,
    "idempotent":                    res.Idempotent,
})
```

With:
```go
newRevID, err := uuid.Parse(res.NewRevisionID)
if err != nil {
    httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
    return
}
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.RestoreDocumentCheckpoint200JSONResponse{
    NewRevisionId:              newRevID,
    NewRevisionNum:             res.NewRevisionNum,
    SourceCheckpointVersionNum: &versionNum,
    Idempotent:                 res.Idempotent,
})
```

Note: `versionNum` is the checkpoint version from the path parameter. If `res.Idempotent` is named differently on the result struct, adjust.

- [ ] **Step 7: Fix raw stats (line ~319):**

Read `application.DocumentStats` struct definition. Then replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, stats)
```

With:
```go
httpresponse.WriteJSON(w, http.StatusOK, documentsapi.DocumentStatsResponse{
    ByArea:   stats.ByArea,
    ByStatus: stats.ByStatus,
})
```

Verify field names on `stats` match. `DocumentStatsResponse.ByArea = map[string]int64`, `ByStatus = map[string]int64`.

- [ ] **Step 8: Build:**

```
go build ./internal/modules/documents/delivery/http/...
```

Expected: exit 0.

---

### Task 4: GREEN — documents tests pass

- [ ] **Step 1: Run typed-response shape tests:**

```
go test ./internal/modules/documents/delivery/http/... -run "TypedResponseShape" -v
```

Expected: all PASS.

- [ ] **Step 2: Run full documents test suite:**

```
go test ./internal/modules/documents/delivery/http/...
```

Expected: 0 FAIL.

---

### Task 5: Fix taxonomy profiles — mapper + handler wiring

**Files:**
- Create: `internal/modules/taxonomy/delivery/http/routes_mapping.go`
- Modify: `internal/modules/taxonomy/delivery/http/routes_profiles.go`

- [ ] **Step 1: Create `routes_mapping.go`:**

```go
package http

import (
    taxonomyapi "metaldocs/internal/modules/taxonomy/api"
    "metaldocs/internal/modules/taxonomy/domain"
)

func toDocumentProfileItem(p *domain.DocumentProfile) taxonomyapi.DocumentProfileItem {
    if p == nil {
        return taxonomyapi.DocumentProfileItem{}
    }
    dto := taxonomyapi.DocumentProfileItem{
        Code:               string(p.Code),
        FamilyCode:         string(p.FamilyCode),
        Name:               p.Name,
        Description:        p.Description,
        ReviewIntervalDays: p.ReviewIntervalDays,
        // Missing domain fields — zero-fill (D-1: domain data gap, not F1.4 scope)
        ActiveSchemaVersion: 0,
        ApprovalRequired:    false,
        RetentionDays:       0,
        ValidityDays:        0,
        WorkflowProfile:     "",
    }
    if p.Alias != "" {
        dto.Alias = &p.Alias
    }
    return dto
}
```

Note: verify `domain.ProfileCode` and `domain.FamilyCode` are string types or have String() method.

- [ ] **Step 2: Fix `listProfiles` (line ~67) in `routes_profiles.go`:**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
```

With:
```go
dtos := make([]taxonomyapi.DocumentProfileItem, len(items))
for i, p := range items {
    dtos[i] = toDocumentProfileItem(p)
}
httpresponse.WriteJSON(w, http.StatusOK, taxonomyapi.ListDocumentProfilesResponse{Items: dtos})
```

Note: `items` is `[]*domain.DocumentProfile` or `[]domain.DocumentProfile` — adjust pointer/value accordingly.

- [ ] **Step 3: Fix `createProfile` (line ~111):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusCreated, profile)
```

With:
```go
httpresponse.WriteJSON(w, http.StatusCreated, toDocumentProfileItem(profile))
```

- [ ] **Step 4: Fix `getProfile` (line ~126):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, profile)
```

With:
```go
httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile))
```

- [ ] **Step 5: Fix `updateProfile` (line ~169):**

Replace:
```go
httpresponse.WriteJSON(w, http.StatusOK, profile)
```

With:
```go
httpresponse.WriteJSON(w, http.StatusOK, toDocumentProfileItem(profile))
```

- [ ] **Step 6: Add `taxonomyapi` import to `routes_profiles.go`** if not present:

```go
taxonomyapi "metaldocs/internal/modules/taxonomy/api"
```

- [ ] **Step 7: Build:**

```
go build ./internal/modules/taxonomy/delivery/http/...
```

Expected: exit 0.

---

### Task 6: Fix audit Allow header

**Files:**
- Modify: `internal/modules/audit/delivery/http/handler.go`

- [ ] **Step 1: Find the exact 405 block** (around line 81):

Read `internal/modules/audit/delivery/http/handler.go` lines 75–95. Confirm the 405 path.

- [ ] **Step 2: Add Allow header before the problem write:**

Before the `writeProblem(w, problem.New(http.StatusMethodNotAllowed, ...))` call, add:
```go
w.Header().Set("Allow", "GET")
```

- [ ] **Step 3: Build:**

```
go build ./internal/modules/audit/delivery/http/...
```

Expected: exit 0.

---

### Task 7: GREEN — taxonomy + audit tests pass + full suite

- [ ] **Step 1: Run taxonomy typed-response tests:**

```
go test ./internal/modules/taxonomy/delivery/http/... -run "TypedResponseShape" -v
```

Expected: all PASS.

- [ ] **Step 2: Run taxonomy full suite:**

```
go test ./internal/modules/taxonomy/delivery/http/...
```

Expected: 0 FAIL.

- [ ] **Step 3: Run audit Allow test:**

```
go test ./internal/modules/audit/delivery/http/... -run "Allow" -v
```

Expected: PASS.

- [ ] **Step 4: H-D grep (V1):**

```
grep -rn "map\[string\]any" internal/modules/documents/delivery/http/ internal/modules/taxonomy/delivery/http/ internal/modules/audit/delivery/http/
```

Expected: **0 matches**.

- [ ] **Step 5: Full build + vet:**

```
go build ./...
go vet ./internal/modules/documents/... ./internal/modules/taxonomy/... ./internal/modules/audit/...
```

Expected: both exit 0.

- [ ] **Step 6: Full suite:**

```
go test ./...
```

Expected: 0 FAIL.

- [ ] **Step 7: Commit:**

```bash
git add \
  internal/modules/documents/delivery/http/handler.go \
  internal/modules/documents/delivery/http/handler_typed_response_test.go \
  internal/modules/taxonomy/delivery/http/routes_mapping.go \
  internal/modules/taxonomy/delivery/http/routes_profiles.go \
  internal/modules/taxonomy/delivery/http/routes_profiles_typed_test.go \
  internal/modules/audit/delivery/http/handler.go \
  internal/modules/audit/delivery/http/handler_allow_test.go \
  docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.4-typed-responses-class/spec.md \
  docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.4-typed-responses-class/plan.md
git commit -m "$(cat <<'EOF'
fix(handlers): H-D class to zero — typed responses in documents, taxonomy, audit (F1.4/A6-A8)

Replace map[string]any literals with typed generated structs across 6 document
handler sites, 4 taxonomy profile sites, and raw stats. Add RFC 7231 Allow header
to audit 405. H-D grep returns 0. TDD: RED before, GREEN after.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Self-Review

**Spec coverage:**
- V1 (H-D grep 0) → Tasks 3+5+6 (handler fixes) + Task 7 Step 4
- V2 (documents suite) → Task 4
- V3 (taxonomy suite) → Task 7 Step 2
- V4 (audit suite) → Task 7 Step 3
- V5 (build + vet) → Task 7 Step 5
- V6 (full suite) → Task 7 Step 6
- V7 (audit Allow test) → Tasks 2+7
- V8 (taxonomy shape tests) → Tasks 2+7

**Placeholder scan:** None. Steps 3.5 and 3.6 note "verify at implementation time" for `toRevisionHistoryResponse` return type and commit result field names — these are bounded verification notes, not implementation gaps. The mapper in Task 5 shows all fields explicitly.

**Type consistency:**
- `documentsapi.DocumentSessionWriterResponse.SessionId` is `openapi_types.UUID` (= `uuid.UUID`) — uuid.Parse required ✓
- `documentsapi.CommitDocumentAutosave200JSONResponse.RevisionId` is `openapi_types.UUID` — uuid.Parse required ✓
- `documentsapi.RestoreDocumentCheckpoint200JSONResponse.NewRevisionId` is `openapi_types.UUID` — uuid.Parse required ✓
- `taxonomyapi.DocumentProfileItem` all fields: string/int/bool — no UUID parsing needed ✓
- `documentsapi.DocumentStatsResponse.ByArea / ByStatus` are `map[string]int64` — direct from `application.DocumentStats` ✓

**Scope guard:** `map[string]string` at lines 671 and 1045 are NOT fixed in F1.4 (not map[string]any, don't fail H-D grep). D-2 bounded defer recorded in spec.md.
