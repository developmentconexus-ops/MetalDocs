# F4.1 plan

Executed via subagents (subagent-driven-development); main session reviews + commits each task.

## Task A — unified transition function + wiring (Go only, no `rejected` removal yet)
Files: `internal/modules/documents/domain/state.go` (replace), new `state_test.go`, delete dead
`CanTransitionDocument` + its `model_test.go` cases; `internal/modules/documents/approval/application/*`
(submit/decision/publish/scheduler/supersede/obsolete/cancel services — route through the fn, drop
scattered status-equality guards, keep OCC `WHERE` + instance-status guards).
- Function = contract §1.2 table (INCLUDING `approved→draft`, `scheduled→draft`, `under_review→draft`;
  EXCLUDING every `rejected` and `archived` arc). Typed `ErrInvalidStateTransition` (reuse existing).
- Coverage test: every ordered `(cur,next)` pair over the 8 non-rejected values → expected per §1.2.
- Gate: `go build ./...`; `go test ./internal/modules/documents/...` targeted green; guard census = 0.
- Parity strict-equality test deferred to Task B (lands with the trigger arc removal).

## Task B — `rejected` removal: DB + Go + parity (depends on A)
Files: new migration `db/migrations/NNNN_documents_remove_rejected.sql` (tighten
`documents_status_check` to 8 values with a DO-block row precheck like `0265`; rewrite
`enforce_document_transition` dropping `under_review→rejected` + `rejected→draft`); baseline sync if
required; `domain/model.go` (drop `DocStatusRejected`); `approval/repository/active_instance_reader.go`
(+ parity test); delete/retarget the raw-SQL rejected roundtrip integration test.
- Add app↔DB parity test (contract §1.4): fn legal set == trigger arcs; negative proof
  `under_review→rejected` → `check_violation`.
- Gate: migration applies clean (precheck aborts if any `status='rejected'` row); `go build ./...`;
  targeted go tests + parity test green.

## Task C — `rejected` removal: contract + FE (depends on B enum shape)
Files: `api/openapi/**` document-status enum (drop `rejected`) + regen BE (`api.gen.go` ×3) + FE
(`api-types`); FE `parseDocumentStatus`/`documentStatusPresentation`/`StatusPill`/`documentWorkflow`/
`libraryStatus`/`approvalWorkflow` + fixtures.
- Gate: `oapi-codegen` + FE typegen clean, zero hand-edits to generated files; openapi lint green;
  `tsc --noEmit` + targeted vitest green; no `rejected` branch remains.

## Ordering
A → B → C. A is solo (everything depends on the fn). B and C are sequential (C needs B's enum truth).
