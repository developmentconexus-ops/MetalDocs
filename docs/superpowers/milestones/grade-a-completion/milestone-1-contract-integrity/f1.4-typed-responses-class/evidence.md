# F1.4 evidence — typed-responses class

> **Feature:** F1.4 — typed-responses-class (Milestone 1 / Contract integrity)
> **Closed:** 2026-06-15
> **Scope:** A6, A7, A8 — H-D class across documents / taxonomy / audit modules.

---

## 1. Changes

**Backend**

- `internal/modules/documents/delivery/http/handler.go`
  - Line 319: stats → typed `documentsapi.DocumentStatsResponse{ByArea, ByStatus}`.
  - Line 697–701: readonly session → typed `documentsapi.DocumentSessionReadonlyResponse{Mode: Readonly, HeldBy, HeldUntil}`.
  - Line 704–719: writer session → typed `documentsapi.DocumentSessionWriterResponse{Mode: Writer, SessionId, ExpiresAt, LastAckRevisionId}` (uuid.Parse on `sess.ID` and `sess.LastAcknowledgedRevisionID`; 500 on parse error).
  - Line 825–833: presign autosave → typed `documentsapi.DocumentAutosavePresignResponse{UploadUrl, PendingUploadId, ExpiresAt}` (uuid.Parse on `PendingUploadID`).
  - Line 873–892: commit autosave → typed `documentsapi.CommitDocumentAutosave200JSONResponse{RevisionId, RevisionNum, IdempotentReplay, FileSizeBytes, PageCount, PageCountSource}` (uuid.Parse on `res.RevisionID`).
  - Line 933–940: revision history → typed `documentsapi.DocumentRevisionHistoryResponse{Items}` via new `toAPIRevisionHistoryItems` mapper.
  - Line 1003–1020: new helper `toAPIRevisionHistoryItems([]domain.RevisionHistoryItem) ([]documentsapi.DocumentRevisionHistoryItem, error)`.
  - Line 1033–1043: restore checkpoint → typed `documentsapi.RestoreDocumentCheckpoint200JSONResponse{NewRevisionId, NewRevisionNum, SourceCheckpointVersionNum, Idempotent}` (uuid.Parse on `NewRevisionID`).

- `internal/modules/taxonomy/delivery/http/routes_mapping.go` (NEW)
  - `toDocumentProfileItem(*domain.DocumentProfile) taxonomyapi.DocumentProfileItem` — maps available domain fields; zero-fills the 5 fields missing from `domain.DocumentProfile` (D-1).

- `internal/modules/taxonomy/delivery/http/routes_profiles.go`
  - Line 68–72 (listProfiles): `map[string]any{"items": items}` → `taxonomyapi.ListDocumentProfilesResponse{Items: dtos}`.
  - Line 116 (createProfile): `profile` → `toDocumentProfileItem(profile)`.
  - Line 131 (getProfile): `profile` → `toDocumentProfileItem(profile)`.
  - Line 174 (updateProfile): `profile` → `toDocumentProfileItem(profile)`.
  - Line 208 (setDefaultTemplate, drive-by): `map[string]any{}` → `taxonomyapi.SetTaxonomyProfileDefaultTemplate200Response{}`.

- `internal/modules/audit/delivery/http/handler.go`
  - Line 82: added `w.Header().Set("Allow", "GET")` before 405 writeProblem (RFC 7231).

**Tests (RED → GREEN)**

- `internal/modules/taxonomy/delivery/http/routes_profiles_typed_test.go` (NEW)
  - `TestListProfiles_DropsDomainFields` — asserts `tenant_id`, `owner_user_id`, `editable_by_role` absent.
  - `TestGetProfile_DropsDomainFields` — same for single GET.
- `internal/modules/audit/delivery/http/handler_allow_test.go` (NEW)
  - `TestHandleEvents_405_Allow` — asserts non-empty `Allow` header in 405 response.

**Fixture migration (drive-by repair under new CLAUDE.md §4 hard gate)**

The typed contract change forced sloppy non-UUID test fixtures into compliance:
- `internal/modules/documents/delivery/http/handler_test.go`
  - `acquireSession` fake: `"sess_1"` → `"aaaaaaaa-aaaa-4aaa-8aaa-000000000001"`; added `LastAcknowledgedRevisionID` UUID.
  - `AcquireSession_Happy` fixture: same migration.
  - `commitResult` fakes: `"rev_2"` → `"bbbbbbbb-bbbb-4bbb-8bbb-000000000001"` (3 sites).
  - `revisionHistory` fake `DocumentID`: `"doc_1"` → `"11111111-1111-4111-8111-111111111111"`.
  - `PresignAutosave` fake: `PendingUploadID: "pending_1"` → UUID.
- `internal/modules/documents/module_wrapper_test.go`
  - `moduleTestService.PresignAutosave`: `"pending_1"` → UUID.

Pre-existing inconsistency (production paths always carry UUIDs from the DB). F1.4 = natural trigger to fix. No production code touched by the fixture changes.

**Docs**

- `f1.4-typed-responses-class/spec.md` (existing — interview record + contract read)
- `f1.4-typed-responses-class/plan.md` (existing)
- `f1.4-typed-responses-class/evidence.md` (this file)
- `CLAUDE.md` §4 — added "Test framework hard gate" subsection (triggered by F1.4 fixture-drift discovery)
- Memory: `test-framework-hard-gate.md` added

---

## 2. Validation gates — outcomes

| Gate | Result | Evidence |
|------|--------|----------|
| V1 — H-D grep returns 0 across the 7 spec'd sites | **GREEN (scoped)** | `grep map\[string\]any` on `documents/handler.go` 7 spec sites, `taxonomy/routes_profiles.go` 4 spec sites + drive-by site 208, `audit/handler.go` Allow path → all clear. Remaining hits in `audit/handler.go` (lines 120/127/216/268/404) are OUT OF F1.4 SCOPE (spec interview restricted A8 to the Allow header; audit other endpoints = D-3). One hit at `documents:615` is a service-input map (`ContentFormData`), not a response emission — not H-D class. One hit at `audit:51` is the `Payload` struct field — input type, not response — not H-D class. |
| V2 — Documents `go test ./internal/modules/documents/delivery/http/...` green | **GREEN** | `ok  metaldocs/internal/modules/documents/delivery/http 3.015s` |
| V3 — Taxonomy `go test ./internal/modules/taxonomy/delivery/http/...` green | **GREEN** | `ok  metaldocs/internal/modules/taxonomy/delivery/http 2.686s` |
| V4 — Audit `go test ./internal/modules/audit/delivery/http/...` green | **GREEN** | `ok  metaldocs/internal/modules/audit/delivery/http 1.840s` |
| V5 — `go build ./...` exit 0 | **GREEN** | (exit 0) |
| V6 — `go test ./...` 0 FAIL | **GREEN** | Full suite all `ok` — including drive-by repair of `internal/modules/documents.TestRegisterRoutesWithRateLimit_WrapperForwardingStillWorks`. |
| V7 — `TestHandleEvents_405_Allow` PASS | **GREEN** | Audit suite includes new test PASS. |
| V8 — `TestListProfiles_DropsDomainFields` + `TestGetProfile_DropsDomainFields` PASS | **GREEN** | Taxonomy suite includes both new tests PASS. |

---

## 3. TDD proof

**RED (pre-impl):**
- `TestListProfiles_DropsDomainFields` FAIL — `tenant_id`, `owner_user_id`, `editable_by_role` present in response (raw domain emission).
- `TestGetProfile_DropsDomainFields` FAIL — same.
- `TestHandleEvents_405_Allow` FAIL — `Allow` header empty in 405.

**GREEN (post-impl):** all 3 new tests PASS. No documents RED test added — `map[string]any` and the typed struct emit identical JSON shape for matching keys; the structural V1 grep is the contract enforcement (no test can distinguish). Documents-side proof = compile-time + existing handler-test coverage GREEN after fixture migration.

---

## 4. Bounded defers

- **D-1 — Taxonomy `DocumentProfileItem` semantic gap.** 5 OpenAPI-required fields (`ActiveSchemaVersion`, `ApprovalRequired`, `RetentionDays`, `ValidityDays`, `WorkflowProfile`) are not in `domain.DocumentProfile`. Mapper zero-fills them. Closes H-D class (raw domain no longer emitted). Domain/DB schema extension is a separate concern — does NOT trigger HS-2 (no shared API redesign required by F1.4). Trigger to retire: when taxonomy domain model is extended (separate ticket).
- **D-2 — `map[string]string` sites not caught by H-D grep.** `documents/handler.go:671` (download URL) and `:1045` (export PDF URL) emit `map[string]string{"url": ...}`. Out of H-D grep coverage. Same structural fix pattern when addressed.
- **D-3 — Audit module other endpoints.** `audit/handler.go` lines 120, 127, 216, 268, 404 still emit `map[string]any` response bodies. Out of F1.4 spec scope (interview limited A8 to the 405 Allow header). Separate feature to typed-ify audit list/ingest/aggregate endpoints. Not blocking M1 (these are not on the H-D class M1 closes; F1.4 closed the 3 named sites the interview surfaced).
- **D-4 — Drive-by repair: setDefaultTemplate empty response (taxonomy:208).** Fixed in this PR per CLAUDE.md §4 drive-by policy. No defer.
- **D-5 — Test framework formalization.** HTTP handler-test framework is not yet formalized (CLAUDE.md §4 notes "formalization pending — ADR when scaffolded"). F1.4 migrated the touched fixtures (sess_1, rev_2, pending_1, doc_1) to UUIDs as drive-by. Adjacent untouched HTTP handler tests remain on the legacy ad-hoc `fakeSvc` pattern until their own feature touches them. Same surgical rule.

---

## 5. Run logs

```
$ go test ./internal/modules/documents/delivery/http/... ./internal/modules/taxonomy/delivery/http/... ./internal/modules/audit/delivery/http/...
ok  metaldocs/internal/modules/documents/delivery/http  3.015s
ok  metaldocs/internal/modules/taxonomy/delivery/http   2.686s
ok  metaldocs/internal/modules/audit/delivery/http      1.840s

$ go build ./...
(exit 0)

$ go test ./...
(all packages ok — 0 FAIL)
```

---

## 6. Closure

F1.4 closed at the V1–V8 bar declared in spec.md. A6, A7, A8 defects resolved at the spec-scoped boundary. Drive-by repair on taxonomy:208. Test fixture migration to UUID identity in 2 files (driven by typed contract enforcement — root cause = non-framework fixtures, fix = framework adoption, NOT contract weakening). New CLAUDE.md §4 test-framework hard gate codifies the policy going forward.
