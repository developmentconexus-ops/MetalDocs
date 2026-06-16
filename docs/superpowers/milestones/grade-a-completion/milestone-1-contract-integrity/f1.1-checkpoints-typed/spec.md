# Feature F1.1 — Spec

> **Milestone:** 1 — Contract / API integrity  ·  **Folder:** `f1.1-checkpoints-typed`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / leandrotca.work — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Is the contract ambiguous? | **No.** Mission §5 A1 names the producer site (`internal/modules/documents/delivery/http/handler.go:881`) and the consumer side is locked: generated server type `internal/modules/documents/api/api.gen.go:166` (`DocumentCheckpoint`, snake_case JSON tags), OpenAPI schema `DocumentCheckpoint` at `frontend/apps/web/src/lib/api-types/index.d.ts:2264` and the FE consumer at `frontend/apps/web/src/features/documents/api/documents.ts:16,92-100` (`Checkpoint = components['schemas']['DocumentCheckpoint']`). The shape is already declared; the handler is the only side disagreeing. |
| 2 | Status codes in scope? | **No — keep both as spec declares.** Generated typed responses use `ListDocumentCheckpoints200JSONResponse` (`api.gen.go:2686`) and `CreateDocumentCheckpoint201JSONResponse` (`api.gen.go:2773`). The 201 on create is already spec-conformant; A4/A5 (201→200 corrections) are templates routes and live in F1.2. |
| 3 | `restoreCheckpoint` in scope? | **No.** Mission §5 A1 is bounded to handler.go:881 (list) and the adjacent create at :954, both raw-domain emits. `restoreCheckpoint` (handler.go:977) is a `map[string]any` emit — part of the A6 "pervasive map[string]any" class → covered by F1.4, not F1.1. Pulling it in here would drift scope (HS-6). |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the FE `Checkpoint` type at `frontend/apps/web/src/features/documents/api/documents.ts:16`
  (`type Checkpoint = components['schemas']['DocumentCheckpoint']`) consumed by `listCheckpoints` /
  `createCheckpoint` (lines 92–100) and rendered by `CheckpointsPanel` /
  `frontend/apps/web/src/features/documents/components/CheckpointsPanel.tsx`. The FE relies on the
  schema being exactly the snake_case `DocumentCheckpoint` codegen shape; today it receives PascalCase
  keys (`ID`, `DocumentID`, `RevisionID`, `VersionNum`, `Label`, `CreatedAt`, `CreatedBy`) because
  `internal/modules/documents/domain/model.go:108` `Checkpoint` carries **no JSON tags**, so the handler
  marshals Go field names.
- **Contract:** both endpoints emit the generated `DocumentCheckpoint` schema:
  - `GET  /api/v1/documents/{id}/checkpoints` → `200` body **`[]DocumentCheckpoint`** (empty array when
    none — never `null`).
  - `POST /api/v1/documents/{id}/checkpoints` → `201` body **`DocumentCheckpoint`** (single object).
  Each `DocumentCheckpoint` is exactly `{ id, document_id, revision_id, version_num, label, created_at,
  created_by }` — uuid strings for id/document_id/revision_id, RFC3339 for created_at, `version_num` an
  integer, `label`/`created_by` strings. Wire JSON keys are snake_case; **no PascalCase keys, no extra
  keys, no missing keys**.
- **Source of truth:** the generated server types `api.DocumentCheckpoint` /
  `ListDocumentCheckpoints200JSONResponse` / `CreateDocumentCheckpoint201JSONResponse` in
  `internal/modules/documents/api/api.gen.go` and the OpenAPI schema they're generated from.

## What this feature implements

At the handler/contract surface (`internal/modules/documents/delivery/http/handler.go`):

1. Replace the `httpresponse.WriteJSON(w, http.StatusOK, items)` raw-domain emit in `listCheckpoints`
   (line 881) with construction of `api.ListDocumentCheckpoints200JSONResponse` (a typed slice of
   `api.DocumentCheckpoint`) from the `[]domain.Checkpoint` result, then write it via the generated
   `Visit…Response` path **or** the equivalent typed `httpresponse.WriteJSON` payload — whichever
   matches the established handler pattern in the same package (chosen in `plan.md`).
2. Replace the `httpresponse.WriteJSON(w, http.StatusCreated, cp)` raw-domain emit in `createCheckpoint`
   (line 954) with construction of `api.CreateDocumentCheckpoint201JSONResponse` (`= api.DocumentCheckpoint`)
   from the `*domain.Checkpoint` result, written the same way.
3. Introduce a small in-package mapper `toAPICheckpoint(domain.Checkpoint) api.DocumentCheckpoint` and
   `toAPICheckpoints([]domain.Checkpoint) []api.DocumentCheckpoint` to keep handler bodies thin and the
   domain type unchanged. (No JSON tags added to `domain.Checkpoint` — domain stays transport-agnostic.)

No other behavior changes: authz, error mapping, status codes, route paths, and request decoding are
untouched.

## Non-goals (mandatory)

- **No** change to `domain.Checkpoint` (no JSON tags, no field rename) — the mapping lives at the
  handler boundary.
- **No** OpenAPI shape change, no `api.gen.go` regeneration, no FE codegen regen in this feature
  (the schema already matches; codegen regen is F1.4's milestone-close action).
- **No** touch on `restoreCheckpoint` (handler.go:977) — that is F1.4 (A6 map[string]any class).
- **No** touch on any other handler or any other route — list + create checkpoint **only**.
- **No** change to status codes (200 list / 201 create) — those already match the spec.
- **No** FE-side shim, adapter, or translator — the fix lives server-side at the contract surface.
- **No** schema/migration, service-layer, or repository change.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `listCheckpoints` response body is a JSON array whose elements have **exactly** the keys `{id, document_id, revision_id, version_num, label, created_at, created_by}` — no PascalCase, no extras, no missing | `TestListCheckpoints_TypedResponseShape` (new, handler-level using `httptest` + a fake `svc`) asserts each emitted object's key set equals the snake_case set and rejects PascalCase keys | **fixture** (handler test with stubbed service — shape is the surface under test) |
| `createCheckpoint` response body is a single JSON object with the same snake_case key set; status is `201` | `TestCreateCheckpoint_TypedResponseShape` (new, same harness) asserts status `201`, the object's keys, and the `version_num` is an integer | **fixture** |
| Empty list returns `[]` not `null` | Same test, sub-case `Empty` | **fixture** |
| H-D grep (mission report §6 commands) on the two sites returns **0** | `rg "WriteJSON\([^)]*Status\w+,\s*(items|cp)\)" internal/modules/documents/delivery/http/handler.go` returns no match in lines 867–955 | command output saved to evidence |
| No regression — package + whole-repo build | `go build ./...` clean; `go test ./internal/modules/documents/...` green; `go test ./...` green | mixed |

> TDD: write `TestListCheckpoints_TypedResponseShape` first asserting the snake_case key set (will
> **fail** against current code — current wire keys are PascalCase), then add the mapper + typed
> emit to make it green. Same for the create test.

## ADR needed?

- [x] No durable decision — skip. This conforms the handler to the *existing* generated contract
  (`api.DocumentCheckpoint`) and the *existing* OpenAPI schema. No new architectural choice.
