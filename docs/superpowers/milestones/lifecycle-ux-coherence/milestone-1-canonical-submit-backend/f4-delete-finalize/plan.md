# Feature F4 — Plan

> **Milestone:** 1 — canonical-submit-backend · **Folder:** `f4-delete-finalize`
> **Status:** Done

## Source
- Milestone row: *Delete the finalize wrapper chain from Go; retain the domain sentinels (now
  submit-owned); one submit entrypoint.*
- Governing-spec reference: ADR 0073 §1.

## Plan

1. **Remove the OpenAPI `/finalize` path** (dirty tree — already done in the absorbed pre-work), leave
   a deprecation note in the `/submit` description; regen `oapi-codegen` (absorbed).
2. **Delete the Go wrapper surface**: `finalizeDocument` handler, its route registration, the
   finalize-only contract types, and the live `GetFinalizePrereqs` repository method (its query was
   already moved into the F1 resolver).
3. **Retain** the four domain sentinels — they are now the submit path's error contract (F1/F3).
4. **Retain** provenance comments citing `repository.go:1801-1817` — documentation, not code.
5. **Do not touch** the unrelated `FreezeFinalizer` / render `MarkFailed(finalize=true)` / legacy-state
   rejection — different concepts that happen to share the substring.

### Files touched
- `api/openapi/v1/openapi.yaml` (absorbed — path removed, note added)
- generated `api.gen.go` families (absorbed regen)
- approval `http` + `infrastructure` (wrapper handler/route/method deletion)

### Test strategy
- **Grep gate** — `func.*finalize|GetFinalizePrereqs|/finalize|finalizeDocument` under approval →
  comments only, zero live symbols; OpenAPI has no `/finalize` path.
- `go build ./...` + `go test ./...` green (nothing references deleted symbols).

### Ordering
Confirm F1 moved the query → delete wrapper → grep gate → build/test.

## Execution notes
Built inline (spike), retro-formalized. Deletion is safe because F1 already re-homed the only real
logic the wrapper carried (the prereq reads); everything else was a thin pass-through.
