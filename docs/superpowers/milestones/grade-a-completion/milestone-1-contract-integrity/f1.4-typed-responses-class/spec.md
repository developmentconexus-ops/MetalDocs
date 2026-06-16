# Feature F1.4 — typed-responses class — Spec

> **Milestone:** 1 (Contract / API integrity) · **Program:** grade-a-completion
> **Folder:** `f1.4-typed-responses-class`
> **Closes:** A6, A7, A8 — H-D class across documents / taxonomy / audit modules.
> **H-D class zero gate:** `grep -rn "map\[string\]any" internal/modules/*/delivery/http/` returns 0 at close.

## Interview record (contract read from source)

| Q | A | Source |
|---|---|--------|
| What does the A6 documents handler emit? | 6 `map[string]any` literals across `documents/delivery/http/handler.go` (lines 694, 701, 814, 855, 903, 1023) instead of typed generated structs. Also: line 319 emits raw `application.DocumentStats` (A8). | `handler.go` grepped |
| What typed response types exist in the documents module? | `DocumentSessionReadonlyResponse` (line 268), `DocumentSessionWriterResponse` (line 278), `DocumentAutosavePresignResponse` (line 158), `CommitDocumentAutosave200JSONResponse` (line 2410), `DocumentRevisionHistoryResponse` (line 263), `RestoreDocumentCheckpoint200JSONResponse` (line 2876), `DocumentStatsResponse` (line 289). All in `internal/modules/documents/api/api.gen.go`. | `api.gen.go` read |
| Are there UUID fields requiring uuid.Parse? | Yes: `PendingUploadId openapi_types.UUID` (presign), `RevisionId openapi_types.UUID` (commit), `NewRevisionId openapi_types.UUID` (restore). Need `uuid.Parse`. | `api.gen.go` struct defs |
| What does the A7 taxonomy handler emit? | 4 handlers emit raw `*domain.DocumentProfile` or `map[string]any{"items": []domain.DocumentProfile}` — `routes_profiles.go:67, 111, 126, 169`. | `routes_profiles.go` read |
| What typed type exists for taxonomy profiles? | `taxonomyapi.DocumentProfileItem` (line 36) with: `ActiveSchemaVersion int`, `Alias *string`, `ApprovalRequired bool`, `Code string`, `Description string`, `FamilyCode string`, `Name string`, `RetentionDays int`, `ReviewIntervalDays int`, `ValidityDays int`, `WorkflowProfile string`. And `ListDocumentProfilesResponse` (line 65) with `Items []DocumentProfileItem`. | `taxonomy/api/api.gen.go` |
| Are all DocumentProfileItem fields in domain.DocumentProfile? | **No.** `ActiveSchemaVersion`, `ApprovalRequired`, `RetentionDays`, `ValidityDays`, `WorkflowProfile` are **not in** `domain.DocumentProfile`. Domain has: `Code`, `FamilyCode`, `Name`, `Description`, `Alias`, `ReviewIntervalDays`, `DefaultTemplateVersionID`, `OwnerUserID`, `EditableByRole`, `ArchivedAt`, `CreatedAt`. | `domain/profile.go` read |
| Resolution for missing domain fields? | Zero-fill (`0`, `false`, `""`) in mapper — closes H-D class (raw domain no longer emitted). The semantic data gap (missing business values) is a bounded defer (D-1), not blocking M1. This is NOT HS-2: no shared API redesign, no DB schema change required by F1.4. | Decision 2026-06-15 |
| What does the A8 audit site emit? | `audit/delivery/http/handler.go:81` — returns HTTP 405 Method Not Allowed without the required RFC 7231 `Allow` header. | `handler.go:81` |
| Is there an export PDF URL response type (handler.go:1045)? | `handler.go:1045` emits `map[string]string{"url": url}`. This is `map[string]string`, NOT `map[string]any` — NOT caught by the H-D grep. Out of F1.4 scope (no blocking gate impact). Note as bounded defer D-2. | `handler.go:1045` |
| Any FE codegen regen needed? | Only if OpenAPI schema changes. F1.4 uses EXISTING generated types from existing OpenAPI schemas — no schema addition, no regen needed. | milestone §10 |

## What this implements

### Documents module — 6 map[string]any sites → typed structs (A6)

In `internal/modules/documents/delivery/http/handler.go`, replace each `map[string]any` literal with its typed equivalent from `internal/modules/documents/api/api.gen.go`. Use inline struct construction; no mapper function (each site is 1-3 field assignments). UUID fields need `uuid.Parse`.

| Handler:Line | Current | Typed replacement | Notes |
|---|---|---|---|
| `getDocumentSession:694` | `map[string]any{mode:"readonly", held_by, held_until}` | `documentsapi.DocumentSessionReadonlyResponse{HeldBy, HeldUntil, Mode}` | Mode: `documentsapi.DocumentSessionReadonlyResponseMode("readonly")` |
| `getDocumentSession:701` | `map[string]any{mode:"writer", session_id, expires_at, last_ack_revision_id}` | `documentsapi.DocumentSessionWriterResponse{SessionId, ExpiresAt, LastAckRevisionId, Mode}` | SessionId, LastAckRevisionId: `uuid.Parse`. Mode: `documentsapi.DocumentSessionWriterResponseMode("writer")` |
| `presignDocumentAutosave:814` | `map[string]any{upload_url, pending_upload_id, expires_at}` | `documentsapi.DocumentAutosavePresignResponse{UploadUrl, PendingUploadId, ExpiresAt}` | PendingUploadId: `uuid.Parse`. 500 on parse error. |
| `commitDocumentAutosave:855` | `map[string]any{revision_id, revision_num, idempotent_replay, file_size_bytes, page_count}` | `documentsapi.CommitDocumentAutosave200JSONResponse{RevisionId, RevisionNum, IdempotentReplay, FileSizeBytes, PageCount}` | RevisionId: `uuid.Parse`. Optional fields: `*bool`, `*int64`, `*int`. 500 on parse error. |
| `listRevisionHistory:903` | `map[string]any{"items": toRevisionHistoryResponse(items)}` | `documentsapi.DocumentRevisionHistoryResponse{Items: items}` | `items` is already `[]documentsapi.DocumentRevisionHistoryItem` from `toRevisionHistoryResponse` — just assign. |
| `restoreDocumentCheckpoint:1023` | `map[string]any{new_revision_id, new_revision_num, source_checkpoint_version_num, idempotent}` | `documentsapi.RestoreDocumentCheckpoint200JSONResponse{NewRevisionId, NewRevisionNum, SourceCheckpointVersionNum, Idempotent}` | NewRevisionId: `uuid.Parse`. SourceCheckpointVersionNum: `*int`. 500 on parse error. |

### Documents module — raw domain stats → typed (A8)

Handler `handler.go:319`: `writeJSON(w, http.StatusOK, stats)` where `stats` is `application.DocumentStats`.

`application.DocumentStats` fields → `documentsapi.DocumentStatsResponse`:
- `stats.ByArea map[string]int64` → `DocumentStatsResponse.ByArea`
- `stats.ByStatus map[string]int64` → `DocumentStatsResponse.ByStatus`

Replace with inline construction: `documentsapi.DocumentStatsResponse{ByArea: stats.ByArea, ByStatus: stats.ByStatus}`. Verify field names at implementation time — if `application.DocumentStats` has the same fields, direct assignment; if different names, map explicitly.

### Taxonomy module — raw domain.DocumentProfile → DocumentProfileItem (A7)

Add a private mapper `toDocumentProfileItem(p *domain.DocumentProfile) taxonomyapi.DocumentProfileItem` in a new file `internal/modules/taxonomy/delivery/http/routes_mapping.go` (mirrors templates pattern). Map all available fields; zero-fill missing required fields (D-1 bounded defer):

```
p.Code          → dto.Code           (string)
p.FamilyCode    → dto.FamilyCode     (string cast)
p.Name          → dto.Name
p.Description   → dto.Description
p.Alias         → dto.Alias          (*string — non-empty → &p.Alias, else nil)
p.ReviewIntervalDays → dto.ReviewIntervalDays
// Missing — zero-fill (D-1):
ActiveSchemaVersion  = 0
ApprovalRequired     = false
RetentionDays        = 0
ValidityDays         = 0
WorkflowProfile      = ""
```

Replace all 4 handler call sites:
- `routes_profiles.go:67` (`listProfiles`): `map[string]any{"items": items}` → `taxonomyapi.ListDocumentProfilesResponse{Items: dtos}` where `dtos` is `items` mapped through `toDocumentProfileItem`
- `routes_profiles.go:111` (`createProfile`): `profile` → `toDocumentProfileItem(profile)`
- `routes_profiles.go:126` (`getProfile`): `profile` → `toDocumentProfileItem(profile)`
- `routes_profiles.go:169` (`updateProfile`): `profile` → `toDocumentProfileItem(profile)`

### Audit module — 405 Allow header (A8)

In `internal/modules/audit/delivery/http/handler.go:81`, before writing the 405 problem response, add:
```go
w.Header().Set("Allow", "GET")
```

## Non-goals

- No OpenAPI schema changes — all typed responses already declared in `api.gen.go`.
- No FE codegen regen — existing generated types, no schema additions.
- No fix to `handler.go:671` (`map[string]string{document_id, ...}`) or `handler.go:1045` (`map[string]string{"url": url}`) — `map[string]string` not caught by H-D grep; D-2 bounded defer.
- No domain/DB schema change for taxonomy profile missing fields — D-1 bounded defer.
- No touch to templates or documents-approval modules (separate milestones).
- No retire of existing mappers outside changed call sites.

## Validation Gate

| # | Criterion | Named proof | Real vs fixture |
|---|-----------|-------------|-----------------|
| V1 | H-D grep returns 0: `grep -rn "map\[string\]any" internal/modules/*/delivery/http/ --include="*.go"` | Command output | structural |
| V2 | Documents `go test ./internal/modules/documents/delivery/http/...` green | Suite pass | real (logic) |
| V3 | Taxonomy `go test ./internal/modules/taxonomy/delivery/http/...` green | Suite pass | real (logic) |
| V4 | Audit `go test ./internal/modules/audit/delivery/http/...` green | Suite pass | real (logic) |
| V5 | `go build ./...` and `go vet ./internal/modules/...` exit 0 | Build output | real |
| V6 | `go test ./...` 0 FAIL | Full suite output | real |
| V7 | `audit/handler_test.go` new test: 405 response includes `Allow: GET` header | `TestHandleEvents_405_Allow` PASS | real (logic) |
| V8 | `taxonomy/routes_profiles_test.go` new tests: list/get/create/update handlers return `DocumentProfileItem` fields (Code, FamilyCode, Name) | Shape assertions PASS | real (logic) |

## ADR needed?

No. Pattern follows ADR 0035 (typed responses at handler boundary). Taxonomy mapper follows F1.3's `toAPITemplateDTO` pattern. Audit Allow header follows RFC 7231 — no new durable decision.

## Approval

Consumer contract read from `api.gen.go` (documents, taxonomy, audit modules) 2026-06-15. No implementation before this spec.
