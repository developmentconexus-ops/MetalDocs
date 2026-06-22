# Feature F3.1 — Evidence (notifications-contract)

> **Status:** CLOSED
> **Closed:** 2026-06-22
> **ADR:** [ADR-0043](../../../../../wiki/decisions/0043-notifications-module-and-lifecycle-bundle.md)

## Acceptance gates (from spec.md Validation Gate)

| Criterion | Command / proof | Result |
|-----------|-----------------|--------|
| `notifications` paths parse clean | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| Generated Go contract present + compiles | `go build ./...` exit 0; `api.gen.go` has `Notification`, `NotificationsListResponse`, `UnreadCountResponse`, `ListNotificationsParams`, `GetNotificationsUnreadCountParams`, `MarkNotificationReadParams` | exit 0 |
| Generated FE type present + field equality | `npm run gen:api` (from `frontend/apps/web`); verified `Notification` at `index.d.ts:2192` vs `NotificationItem` at `lib/types/index.ts:178` — 10 fields match snake_case→camelCase | PASS |
| `event_type` open string; `status` 3-enum | `api.gen.go:93` `EventType string`; `api.gen.go:33-35` `NotificationStatus PENDING\|SENT\|READ` | PASS |
| Cap registered across 4 files + deferred | `go test ./scripts/api-lint/... -run TestSeedRegistryParity` → ok; `go test ./internal/modules/iam/domain/...` → ok; `go test ./apps/api/cmd/metaldocs-api/ -run TestEveryCapSeededOrDeferred` → ok | all green |
| Cap classification guard passes | `go test ./apps/api/cmd/metaldocs-api/ -run TestEveryCapabilityClassified,TestAreaGradeCapabilitySet,TestCapabilityRegistrySize` → ok (size bumped 30→31 + ADR-0043 annotation) | green |
| ADR present + index updated | `wiki/decisions/0043-notifications-module-and-lifecycle-bundle.md` created; `wiki/decisions/index.md` row + Last verified stamp updated | PASS |
| FE typecheck | `npx tsc --noEmit` from `frontend/apps/web` | exit 0 |

## TDD proof

F3.1 is a contract+codegen feature (no runtime handler logic). Guard-level TDD:
- `api-lint -strict` fails before schema exists (OpenAPI parse error on unknown `$ref`); passes at `0 violation(s)` after schema is authored.
- `TestSeedRegistryParity` / `TestEveryCapSeededOrDeferred` fail if cap registered but not in `deferredCaps`; pass once deferred.
- `TestCapabilityRegistrySize` fails at 31 != 30 until count bumped with ADR annotation.
- No fabricated runtime proof — the interface compiles standalone; the handler is F3.2.

## Files changed

| File | Change |
|------|--------|
| `api/openapi/v1/openapi.yaml` | + `notifications` tag; + 3 paths (`/notifications`, `/notifications/unread-count`, `/notifications/{id}/read`); + schemas `Notification`, `NotificationsListResponse`, `UnreadCountResponse` |
| `internal/modules/notifications/api/cfg.yaml` | NEW |
| `internal/modules/notifications/api/gen.go` | NEW |
| `internal/modules/notifications/api/api.gen.go` | GENERATED (32 KB) |
| `internal/modules/iam/domain/model.go` | + `CapNotificationRead` const + `validCapabilities` entry |
| `internal/modules/iam/domain/catalog.go` | + pt-BR description row |
| `internal/modules/iam/domain/capability_scope.go` | + `ScopeTenant` entry + self-scope comment |
| `internal/modules/iam/domain/model_test.go` | size guard 30→31 + ADR-0043 annotation |
| `scripts/api-lint/registry_rules.go` | + `CapNotificationRead` in `deferredCaps` + comment |
| `apps/api/cmd/metaldocs-api/permissions_test.go` | + `CapNotificationRead` in `TestEveryCapSeededOrDeferred` deferred map |
| `frontend/apps/web/src/lib/api-types/index.d.ts` | REGENERATED — carries `Notification`, `NotificationsListResponse`, `UnreadCountResponse` |
| `wiki/decisions/0043-notifications-module-and-lifecycle-bundle.md` | NEW ADR |
| `wiki/decisions/index.md` | + ADR-0043 row + Last verified bump |
| `docs/.../f3.1-notifications-contract/spec.md` | NEW |
| `docs/.../f3.1-notifications-contract/plan.md` | NEW |
| `docs/.../f3.1-notifications-contract/evidence.md` | this file |

## Bounded defers

None. All spec non-goals were respected: no handler, no migration, no FE wire, no projector, no emit-site edit. The two additive published views (`v_approval_instance_submitter`, `v_document_cd_mapping`) are F3.2/F3.3 work — tracked in ADR-0043 §6.

## Review / QA disposition

Spec approved before code (2026-06-22, operator HS-1 start gate). ADR-0043 durable decisions recorded. Separation of powers: milestone-validator judges at milestone close; feature code review on aggregate diff.
