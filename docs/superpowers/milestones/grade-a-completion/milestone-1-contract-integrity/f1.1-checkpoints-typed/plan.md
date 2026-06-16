# Feature F1.1 — Plan

> **Spec:** `./spec.md` (approved 2026-06-15)
> **Scope:** map `domain.Checkpoint` → `documentsapi.DocumentCheckpoint` at the handler boundary for
> list + create. No domain change, no codegen change, no other route.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/documents/delivery/http/handler.go` | Add `toAPICheckpoint` / `toAPICheckpoints` helpers; rewrite `listCheckpoints` (`:867`) and `createCheckpoint` (`:928`) to emit the typed response. |
| `internal/modules/documents/delivery/http/handler_checkpoints_test.go` | **NEW** — handler-level shape tests (`TestListCheckpoints_TypedResponseShape`, `TestCreateCheckpoint_TypedResponseShape`, empty-list sub-case). |

No other file changes. `domain.Checkpoint`, `api.gen.go`, OpenAPI YAML, FE codegen — all untouched.

## Steps (in order)

1. **TDD red — write `handler_checkpoints_test.go`** with two top-level tests:
   - `TestListCheckpoints_TypedResponseShape`
     - Sub-case `Items`: `fakeSvc.listCheckpoints` returns a non-trivial `[]domain.Checkpoint` (filled
       UUIDs, label, `VersionNum`, `CreatedAt`, `CreatedBy`). Hit `GET /api/v1/documents/{id}/checkpoints`.
       Assert status `200`; decode body into a `[]map[string]json.RawMessage`; assert each element's key
       set **equals** `{id, document_id, revision_id, version_num, label, created_at, created_by}`; reject
       any PascalCase key (`ID`, `DocumentID`, `RevisionID`, `VersionNum`, `Label`, `CreatedAt`,
       `CreatedBy`); assert `version_num` decodes to an integer.
     - Sub-case `Empty`: `fakeSvc.listCheckpoints` returns `[]domain.Checkpoint{}`; body equals `[]`
       (raw bytes), not `null`.
   - `TestCreateCheckpoint_TypedResponseShape`
     - Stub returns a populated `*domain.Checkpoint`. POST `{"label":"v1"}`. Assert status `201`; body is
       a single JSON object with the snake_case key set above; reject PascalCase keys.

   The existing `fakeSvc.ListCheckpoints` / `CreateCheckpoint` (test file:155-161) return only `ID` +
   `VersionNum`; **extend** them via new fields on `fakeSvc` (e.g. `listCheckpointsItems`,
   `createCheckpointResult`) so existing tests stay green. (Pattern matches `revisionHistory` field on
   the same fake.) Existing 400-on-empty-label test stays untouched.

   Run `go test ./internal/modules/documents/delivery/http/ -run TestListCheckpoints_TypedResponseShape` →
   expect **FAIL** (current wire keys are PascalCase). Capture failure output for evidence.

2. **TDD green — add helpers + rewrite emits** in `handler.go`:

   ```go
   func toAPICheckpoint(cp domain.Checkpoint) documentsapi.DocumentCheckpoint {
       return documentsapi.DocumentCheckpoint{
           Id:         openapi_types.UUID(uuid.MustParse(cp.ID)),         // see note
           DocumentId: openapi_types.UUID(uuid.MustParse(cp.DocumentID)),
           RevisionId: openapi_types.UUID(uuid.MustParse(cp.RevisionID)),
           VersionNum: cp.VersionNum,
           Label:      cp.Label,
           CreatedAt:  cp.CreatedAt,
           CreatedBy:  cp.CreatedBy,
       }
   }

   func toAPICheckpoints(cps []domain.Checkpoint) []documentsapi.DocumentCheckpoint {
       out := make([]documentsapi.DocumentCheckpoint, 0, len(cps))
       for _, c := range cps { out = append(out, toAPICheckpoint(c)) }
       return out
   }
   ```

   Replace handler.go:881 with:
   ```go
   httpresponse.WriteJSON(w, http.StatusOK, toAPICheckpoints(items))
   ```
   Replace handler.go:954 with:
   ```go
   httpresponse.WriteJSON(w, http.StatusCreated, toAPICheckpoint(*cp))
   ```

   **UUID-parse note:** `documentsapi.DocumentCheckpoint.Id/DocumentId/RevisionId` are
   `openapi_types.UUID`. The repository stores them as strings. The precedent
   `toDocumentSummary` (handler.go:408) parses string→UUID with explicit error mapping; F1.1 mirrors
   that — replace `MustParse` with a non-panicking variant that returns an error, and the handler
   propagates a 500 via the existing `mapErr` path if a stored UUID is malformed (defensive — should
   never trigger in practice). Final form decided when writing the code; the test cares only about wire
   shape, not error mapping.

3. **TDD green verify** — rerun the two new tests, the existing
   `TestCreateCheckpoint_EmptyLabel_Returns400`, and the whole package: all green.

4. **H-D grep proof** — run the §6-style grep from `spec.md`:
   ```
   rg "WriteJSON\([^)]*Status\w+,\s*(items|cp)\)" internal/modules/documents/delivery/http/handler.go
   ```
   Expect: **0 matches** in the checkpoint range. Save output to `evidence.md`.

5. **Broader regression** — `go build ./...`; `go test ./internal/modules/documents/...`; `go test ./...`.

6. **Evidence record** — fill `evidence.md` with: commands run, real output, TDD red-then-green proof,
   grep output, fixture/real labels, any defers (none expected).

## Risk / hard-stop watch

- **HS-2:** if the typed mapper forces an OpenAPI shape change (e.g. a field rename for non-UUID
  reasons) → stop, surface, do not symptom-patch.
- **HS-3:** if `go build ./...` fails before tests start (unrelated drift) → `runtime-contract-prereq`.
- **HS-6:** if writing the test reveals a checkpoint contract gap F5.1 missed (e.g. extra/missing
  field on the OpenAPI schema vs domain) → stop, surface.

## Out of scope (re-stated from spec)

`restoreCheckpoint`, `domain.Checkpoint` JSON tags, `api.gen.go` regen, FE codegen regen, status-code
changes, FE-side adapters, schema/migration, service/repository.
