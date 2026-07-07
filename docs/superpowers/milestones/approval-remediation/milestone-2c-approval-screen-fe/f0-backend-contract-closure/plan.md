# F0 — Plan

Seeded from master plan `docs/superpowers/plans/2026-07-07-m2c-approval-screen-fe.md` §F0,
amended by the HS-2 decision in `spec.md` (freeze markup gate removed, not wired).

## Tasks

1. **[TDD] Wire enum** — `contracts/instance_read.go`: add `InstanceStatusChangesRequested` +
   `IsValidInstanceStatus` helper. Test `TestInstanceStatusWireEnumComplete` (all 5 statuses).
2. **[TDD] Route stage_kind** — `contracts/route.go`: add `StageKind` type + field on
   `StageRequest`/`StageResponse`/`ListStageItem`, validate `review|approval|""` in `validateStages`.
   `route_admin_handler.go`: map `StageRequest.StageKind → domain.Stage.Kind` (empty stays zero,
   defaulted to approval at persistence); populate `ListStageItem.StageKind` via `mapStageKind`.
   Test `TestStageRequestStageKindValidation`. (Persistence + read layers already support Kind.)
3. **openapi** — `ApprovalInstanceByDocumentResponse.status` enum += `changes_requested`;
   `StageRequest` + `StageSummary` += `stage_kind` enum `[review, approval]`.
4. **HS-2 resolution** — delete `application/markup_gate.go` + `markup_gate_test.go` (misdirected;
   no references). Register server-authoritative suggestion gate as bounded defer (README).
5. **Regen** — `go generate` in `approval/api` (oapi-codegen); `npm run gen:api` for FE api-types.
6. **Verify** — `go build ./...`; `go test ./internal/modules/documents/approval/...`; grep-zero
   markup util; FE api-types contain `stage_kind` + `changes_requested`.

## Files touched

- `internal/modules/documents/approval/http/contracts/instance_read.go`
- `internal/modules/documents/approval/http/contracts/route.go`
- `internal/modules/documents/approval/http/contracts/f0_contract_test.go` (new)
- `internal/modules/documents/approval/http/route_admin_handler.go`
- `api/openapi/v1/openapi.yaml`
- `internal/modules/documents/approval/api/api.gen.go` (regen)
- `frontend/apps/web/src/lib/api-types/index.d.ts` (regen)
- Deleted: `application/markup_gate.go`, `application/markup_gate_test.go`
