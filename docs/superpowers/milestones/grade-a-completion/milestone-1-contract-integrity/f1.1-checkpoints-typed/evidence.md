# Feature F1.1 — Evidence

> **Spec:** `./spec.md` (approved 2026-06-15 / leandrotca.work)
> **Plan:** `./plan.md`
> **Status:** Closed — 2026-06-15
> **Closed against the spec's Validation Gate; commands + real output below.**

## What landed

| File | Change |
|------|--------|
| `internal/modules/documents/delivery/http/handler.go` | Added `toAPICheckpoint` / `toAPICheckpoints` (handler-boundary mappers `domain.Checkpoint` → `documentsapi.DocumentCheckpoint`); rewrote `listCheckpoints` (`:881`) and `createCheckpoint` (`:954`) to emit the generated typed response via `httpresponse.WriteJSON`; added `github.com/google/uuid` import for the mapper. Defensive 500 on malformed-UUID (should never trigger — stored UUIDs are validated at write). |
| `internal/modules/documents/delivery/http/handler_test.go` | Added `listCheckpointsItems []domain.Checkpoint` + `createCheckpointResult *domain.Checkpoint` fields on `fakeSvc`; existing canned-return fallbacks preserved → existing tests stay green. |
| `internal/modules/documents/delivery/http/handler_checkpoints_test.go` | **NEW** — `TestListCheckpoints_TypedResponseShape{Items,Empty}`, `TestCreateCheckpoint_TypedResponseShape`; shared `assertCheckpointShape` asserts exact snake_case key set, rejects PascalCase, asserts `version_num` is integer; `Empty` asserts `[]` not `null`. |

`domain.Checkpoint`, `api.gen.go`, OpenAPI YAML, FE codegen — untouched. `restoreCheckpoint` untouched (F1.4 turf).

## TDD red → green

**Red (pre-implementation; mappers absent, handler still emitted raw `domain.Checkpoint`):**

```
$ go test ./internal/modules/documents/delivery/http/ \
    -run "TestListCheckpoints_TypedResponseShape|TestCreateCheckpoint_TypedResponseShape" -v
=== RUN   TestListCheckpoints_TypedResponseShape
=== RUN   TestListCheckpoints_TypedResponseShape/Items
    handler_checkpoints_test.go:91: unexpected key "RevisionID" in checkpoint response (extra/PascalCase): keys=[DocumentID RevisionID VersionNum Label CreatedAt CreatedBy ID]
=== RUN   TestListCheckpoints_TypedResponseShape/Empty
--- FAIL: TestListCheckpoints_TypedResponseShape (0.00s)
    --- FAIL: TestListCheckpoints_TypedResponseShape/Items (0.00s)
    --- PASS: TestListCheckpoints_TypedResponseShape/Empty (0.00s)
=== RUN   TestCreateCheckpoint_TypedResponseShape
    handler_checkpoints_test.go:146: unexpected key "CreatedAt" in checkpoint response (extra/PascalCase): keys=[CreatedBy ID DocumentID RevisionID VersionNum Label CreatedAt]
--- FAIL: TestCreateCheckpoint_TypedResponseShape (0.00s)
FAIL
FAIL	metaldocs/internal/modules/documents/delivery/http	2.760s
FAIL
```

**Green (after `toAPICheckpoint(s)` + typed emits):**

```
$ go test ./internal/modules/documents/delivery/http/ \
    -run "TestListCheckpoints_TypedResponseShape|TestCreateCheckpoint_TypedResponseShape" -v
=== RUN   TestListCheckpoints_TypedResponseShape
=== RUN   TestListCheckpoints_TypedResponseShape/Items
=== RUN   TestListCheckpoints_TypedResponseShape/Empty
--- PASS: TestListCheckpoints_TypedResponseShape (0.00s)
    --- PASS: TestListCheckpoints_TypedResponseShape/Items (0.00s)
    --- PASS: TestListCheckpoints_TypedResponseShape/Empty (0.00s)
=== RUN   TestCreateCheckpoint_TypedResponseShape
--- PASS: TestCreateCheckpoint_TypedResponseShape (0.00s)
PASS
ok  	metaldocs/internal/modules/documents/delivery/http	2.813s
```

## Validation Gate (per `spec.md`)

| Acceptance criterion | Proof | Result |
|----------------------|-------|--------|
| `listCheckpoints` body is `[]DocumentCheckpoint`; each element has exactly `{id, document_id, revision_id, version_num, label, created_at, created_by}`; no PascalCase, no extras, no missing | `TestListCheckpoints_TypedResponseShape/Items` (fixture: handler + fake svc) | PASS |
| `createCheckpoint` body is a single `DocumentCheckpoint` object with the same snake_case key set; status `201` | `TestCreateCheckpoint_TypedResponseShape` (fixture) | PASS |
| Empty list returns `[]` not `null` | `TestListCheckpoints_TypedResponseShape/Empty` (fixture) | PASS |
| H-D grep on the two sites returns 0 | `rg "WriteJSON\([^)]*Status\w+,\s*(items\|cp)\)" internal/modules/documents/delivery/http/handler.go` → no matches | PASS (command output below) |
| Package green (incl. existing `TestCreateCheckpoint_EmptyLabel_Returns400`) | `go test ./internal/modules/documents/delivery/http/` → `ok` | PASS |
| Documents module regression | `go test ./internal/modules/documents/...` → all `ok` | PASS |
| Whole-repo build clean | `go build ./...` → no output (exit 0) | PASS |
| Whole-repo tests green | `go test ./...` → no `FAIL` lines | PASS |

## H-D grep proof

```
$ rg "WriteJSON\([^)]*Status\w+,\s*(items|cp)\)" internal/modules/documents/delivery/http/handler.go
(no matches)
```

→ neither `items` nor `cp` is emitted raw anywhere in the handler file. 0 H-D sites for the F1.1 scope.

## Real vs fixture

- **Fixture:** all shape proof is at the handler layer (httptest + fake svc). The contract surface
  under test **is** the handler — the wire bytes the FE receives — so a fixture proof of the wire
  shape is the right grain for this defect. Repository/DB code is not touched by F1.1.
- **No real-DB integration test required** for F1.1 (unlike F0.1 where the SQL predicate itself was
  the fix). Marshaling correctness is fully observable at the handler boundary.

## Review / QA disposition

- **Self-review (code):** mapper isolated, no domain change, defensive UUID-parse error path mirrors
  `toDocumentSummary` style; status codes unchanged (200/201 already spec-conformant); no FE-side
  shim, no codegen edit, no OpenAPI shape change.
- **Workflow-class QA lens:** `backend-api-qa-checklist` contract-truth — route truth-table for the
  two checkpoint endpoints is now reconciled across runtime / spec / codegen / FE codegen (all
  agree on snake_case `DocumentCheckpoint`); no hand-edit of generated wiring.
- **Skill routing:** backend HTTP/contract → `metaldocs-backend-api` (followed).

## Bounded defers

None for F1.1.

- `restoreCheckpoint` (handler.go:977) still emits a `map[string]any` — **deferred to F1.4** (A6
  class), per `spec.md` non-goals and F1.1 interview Q3. Trigger: F1.4 implementation start.

## Hard-stop watch

- HS-2 / HS-3 / HS-6 — none tripped. No OpenAPI shape change, no prereq failure, no scope drift.

## Commit

To be created after operator close-out review (CLAUDE.md §5.0 — commit without asking, but milestone
workflow per-feature flow records evidence first, then commits).
