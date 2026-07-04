# F4.1 evidence — unified state machine + `rejected` removal

> **Contract:** `../validation-contract.md` §1. Closed across 3 tasks + 1 review-fix.
> Commits: A `3581769b`, B `52188636`, C `51950c26`, review-fix `93f40451`.

## What shipped (per §1 layer)

| §1 requirement | Implementation | Proof |
|---|---|---|
| Single exhaustive transition fn in `documents/domain` (mirrors templates `CanTransition`, typed error, total) | `domain.CanTransitionDocumentStatus(cur,next) error` (`state.go`), replacing the dead `CanTransitionDocument` | `go build ./...` green; dead fn + `model_test.go` gone |
| Legal set == DB-trigger arc set (§1.2 table, 11 arcs post-`rejected`) | 11 legal arcs, exact mirror of `enforce_document_transition` | `TestCanTransitionDocumentStatus` enumerates all 8×8 pairs, 11 legal — **green** |
| App↔DB parity pinned by test (§1.4) | `state_parity_test.go` pins the fn legal-set to the literal post-0272 trigger transcription | `TestCanTransitionDocumentStatus_DBTriggerParity` — **green** |
| Every §0.3 approval service routes through the fn; scattered guard census = 0 | submit/decision/publish/scheduler/supersede/obsolete/cancel wired through `CanTransitionDocumentStatus`; OCC `WHERE status=… AND revision_version=$N` retained | `grep -E "Status *!= *(DocStatus…|\")"` in approval/application non-test = **0** residual lifecycle guards; only `instance.Status != InstanceApproved` remain (approval-INSTANCE FSM, allowlisted per §1.3) |
| `rejected` removed — domain enum | `DocStatusRejected` deleted from `model.go` | build green; no doc-status refs remain |
| `rejected` removed — reader | dropped from `activeInstanceStatuses` (`active_instance_reader.go`), now `{draft,under_review,approved,scheduled}`; SQL placeholders renumbered | reader compiles; parity test updated |
| `rejected` removed — DB CHECK + trigger | migration `0272_documents_remove_rejected.sql`: row pre-check (RAISE if any `status='rejected'`), CHECK tightened to 8 values (`archived` kept), `enforce_document_transition` rewritten dropping `under_review→rejected` + `rejected→draft` arcs (cancel-GUC + all else verbatim) | migration mirrors 0265 idiom; negative proof retargeted (below) |
| `rejected` removed — OpenAPI + regen | `DocumentSummary` status enum: `rejected` removed (openapi.yaml); BE `api.gen.go` + FE `api-types/index.d.ts` regenerated | **zero hand-edits** to generated (enum const + Valid() case removed; swaggerSpec base64 changed wholesale = legit regen); openapi lint green |
| `rejected` removed — FE | `parseDocumentStatus.ts`, `StatusPill.tsx`(+css), `documentStatusPresentation.ts`, `libraryStatus.ts`, `LibrarySidebar.tsx`(+css), `DocumentEditorPage.tsx`, `deriveDashboardStats.ts`(+tests) | `tsc` + vitest green |
| `rejected` removed — tests | `service_review_roundtrip_integration_test.go` Scenario 4 retargeted to a **negative proof**: raw `UPDATE … SET status='rejected' WHERE status='under_review'` now expected to FAIL with check_violation | suite compiles; removed-arc drive asserts the trigger rejects it |

## Two-"rejected" disambiguation (verified kept — NOT removed)

`rejected` as a **document status** is dead → removed. `rejected` as an **approval DECISION**
(`approval_state` / `StateRejected`, live arc under_review→rejected→draft) is a **different state machine**
→ retained. Confirmed the boundary held:
- `approval/api/api.gen.go` `ApprovalStageActorResponseStatusRejected` + `signoff_handler.go:128 return "rejected"` — approval-decision, correctly kept.
- FE `documentWorkflow.ts` `ACTIVE_SIBLING_STATES` keeps `'rejected'` — traced `useDocumentArtifact.ts:199` to `activeDocument.approval_state` (the live decision field), **not** `documents.status`. Correctly kept; not a Task C miss.

## Review disposition

- **Finding (main-session review):** `state.go` header comment claimed "rejected still exists as a
  DocumentStatus constant" — runtime-false after Task B removed the constant. **Fixed** (`93f40451`,
  comment-only) to state it was removed and the fn+trigger dropped its arcs together.
- Guard census re-run after fix: **0** scattered lifecycle guards. Build + domain unit + parity tests **green**.

## Commands (real output)

```
$ go build ./...                                                         → BUILD OK
$ go test ./internal/modules/documents/domain/ -run 'CanTransitionDocumentStatus|DBTriggerParity' -count=1
  ok  metaldocs/internal/modules/documents/domain  1.152s
$ grep -rnE "Status *!= *(DocStatus|documentsdomain\.DocStatus|\")" internal/modules/documents/approval/application/ | grep -v _test.go
  (empty — 0 residual lifecycle guards)
```

## Bounded defers

| Defer | Rationale | Trigger |
|---|---|---|
| `active_instance_parity_test.go` / roundtrip negative-proof **live run** (`-tags integration`) | 20-min box; not run locally (contract §5/§6) | Run on real CI/test-DB before program close-out; drive authored + committed |
| `archived` status-value cleanup | ADR 0010 retains it deliberately; out of M4 scope (contract §6) | M9 governance-hygiene if desired |
