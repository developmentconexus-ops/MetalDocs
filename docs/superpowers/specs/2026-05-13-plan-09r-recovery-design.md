# Plan 9R Recovery Design: Transactional + Idempotency Completion

> **Date:** 2026-05-13
> **Status:** Approved design
> **Scope:** Recovery plan for incomplete Plan 9 transactional, idempotency, optimistic-lock, and template workflow-alignment work.
> **Out of scope:** Audit T-004 hash-chain tamper evidence, cursor pagination, Plan 10 rename/cleanup work, broad route migrations, and speculative new workflow behavior.

## 1. Why Plan 9R exists

Plan 9 is marked done in `wiki/backlog/roadmap.md`, but implementation audit found several Plan 9 requirements still open in runtime code, schema, or capability wiring. Plan 9R is a focused recovery pass that completes only the already-intended Plan 9 scope.

Audit T-004 hash-chain tamper evidence was listed as absorbed into Plan 9, but it is a separate design problem with schema hashing, concurrency, backfill/baseline, and integrity-validation concerns. Plan 9R explicitly defers that work to a dedicated follow-up plan.

## 2. Scope

Plan 9R closes these gaps:

- Auth T-004/R-004: make `CreateUser` atomic across `auth_identities`, `iam_users`, and `iam_user_roles`.
- Documents T-006/R-006: wire finalize idempotency using `internal/platform/idempotency` and `metaldocs.idempotency_keys`.
- Documents T-009/R-009: add a forward migration fixing `document_placeholder_values.revision_id` to reference `document_revisions(id)`.
- Templates v2 T-007/R-007: wrap `CreateTemplate`, `Approve`, and `PublishTemplateVersion` multi-step writes in a single transaction.
- Templates v2 T-009/R-009: require/read `Idempotency-Key` on POST create/mutate routes and replay cached successful responses.
- Templates v2 T-010/R-010: enforce `ExpectedLockVersion` on draft saves and map stale writes to HTTP 412.
- Templates v2 workflow alignment remainder: add typed `CapTemplateReview`, seed `template.review`, and remove remaining reviewer/approver conflation where runtime still permits it.
- Taxonomy T-007/R-007: wrap family deactivate read/check/update in one transaction, lock the family row, and tenant-scope `HasActiveProfiles`.
- Taxonomy R-011 conflict mapping: map `23505` to HTTP 409 for family, profile, and area write paths.

Plan 9R does not close:

- Audit T-004/R-004 hash-chain tamper evidence.
- Cursor pagination work.
- Plan 10 legacy purge, route renames, or module directory renames.
- Unrelated tech-debt rows discovered while touching these files.

## 3. Design Principles

Keep the fix simple and local:

- Prefer passing `*sql.Tx` down existing repository methods over new frameworks.
- Reuse `internal/platform/idempotency`; do not add a second idempotency system.
- Preserve generated `ServerInterfaceWrapper` route wiring for migrated modules.
- Add only the OpenAPI parameters needed for changed public contracts.
- Keep commits scoped by workstream.
- Write focused failing tests before implementation where current behavior is wrong.

Stop if route ownership, generated method signatures, or path parameter names conflict between runtime and OpenAPI.

## 4. Workstream A: Auth Atomic CreateUser

`auth/application/service.go:CreateUser` must wrap identity creation and IAM role replacement in one outer transaction. The auth repository and role admin repository need transaction-aware methods so both write sets share the same `*sql.Tx`.

The transaction should cover:

- Unique identity check.
- Insert into `metaldocs.auth_identities`.
- IAM user upsert.
- IAM role delete/insert.
- `authz.Require` for `user.manage` where the IAM role write requires the tier-2 assertion.

Success criteria:

- Forced failure in role assignment rolls back `auth_identities`.
- Existing successful create behavior remains unchanged.
- No second independent transaction is opened inside the atomic path.

## 5. Workstream B: Documents Finalize Idempotency + FK

Finalize must require and read `Idempotency-Key`, use `internal/platform/idempotency`, and store/replay responses through `metaldocs.idempotency_keys`.

The replay behavior should match the platform contract:

- First request with key executes finalize/submit and stores the response.
- Same key and same body replays the cached successful response with `Idempotent-Replay: true`.
- Same key and different body returns the platform idempotency conflict.

The placeholder FK fix must be a new forward migration. It should not edit historical migration `0152`. The migration must move `document_placeholder_values.revision_id` from `documents(id)` to `document_revisions(id)`.

Success criteria:

- Duplicate finalize POST with same key returns replayed success rather than a new 409.
- Missing key returns the expected client error.
- FK validation targets real document revisions.

## 6. Workstream C: Templates v2 Transactions, Idempotency, OCC, Review Capability

Transaction recovery:

- `CreateTemplate` inserts template, first version, approval config, and audit in one transaction.
- `Approve` updates version/template, obsoletes previous published versions, and writes audit in one transaction.
- `PublishTemplateVersion` updates version/template, obsoletes previous published versions, creates the next draft, and writes audit in one transaction.

Optimistic locking:

- Draft save must use `expected_lock_version` in the repository update predicate.
- Stale writes must return a domain conflict that HTTP maps to `412 Precondition Failed`.
- Successful writes must advance the stored lock version.

Idempotency:

- Require/read `Idempotency-Key` on POST `/api/v2/templates`, `/publish`, `/submit`, `/review`, and `/approve`.
- Use the shared platform idempotency store with 24h TTL.
- Replay returns the same cached success response and does not create duplicate audit rows or repeat state transitions.

Workflow alignment remainder:

- Add `CapTemplateReview Capability = "template.review"` to IAM domain model.
- Seed `template.review` for approver and system_admin roles in a migration.
- Ensure review route checks the typed capability path and review service checks reviewer role plus segregation of duties.
- Approve must no longer act as the reviewer step when a reviewer exists.

Success criteria:

- Create/publish/approve forced mid-flow failures roll back all writes in that flow.
- Stale `expected_lock_version` returns 412.
- Replaying POST routes returns cached success without duplicate state mutation.
- `template.review` exists in code and seed data.

## 7. Workstream D: Taxonomy Deactivate + Conflict Mapping

Family deactivate must execute `GetByCode`, `HasActiveProfiles`, and `Update` inside one transaction. The family row must be locked with `SELECT ... FOR UPDATE` before the active-profile check.

`HasActiveProfiles` must be tenant-scoped. The transaction-aware variant should accept tenant ID explicitly or derive it from the context consistently with existing taxonomy repository style.

Conflict mapping must be completed across:

- `writeFamilyError`
- `writeProfileError`
- `writeAreaError`

Each `23505` unique violation should return HTTP 409 with a resource-specific conflict code, not 500.

Success criteria:

- Deactivate cannot pass the active-profile check based on another tenant's profiles.
- Concurrent deactivate/profile-creation race is closed by transaction and row lock.
- Family/profile/area unique violations map to 409.

## 8. API and Codegen Expectations

Any changed public route contract must update `api/openapi/v1/openapi.yaml` first, then regenerate affected module API packages.

Expected contract changes:

- Add required `Idempotency-Key` headers for documents finalize and templates POST mutate routes in Plan 9R scope.
- Add 412 response for template draft save if not already represented.
- Keep path names and generated method signatures aligned with runtime routes.

Verification gates:

- `npx @redocly/cli lint api/openapi/v1/openapi.yaml`
- `$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/templates_v2/api/...`
- `$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/documents/api/...`
- `go build ./...`
- Targeted Go tests by workstream.
- `go test ./...` before completion.
- Frontend API type generation if OpenAPI changes require it.

## 9. Documentation and Roadmap Updates

Plan 9R should update:

- `wiki/backlog/roadmap.md`: mark original Plan 9 as partially recovered or add Plan 9R status with explicit audit T-004 deferral.
- Affected module tech-debt/backlog rows: mark only rows actually closed.
- `wiki/README.md`: refresh index lines if module docs changed.

Audit T-004 must remain open and point to a future dedicated tamper-evidence plan.

## 10. Final Review Standard

After all workstreams land, run one grouped implementation audit. The audit should check:

- Spec compliance against this design.
- Plan compliance by workstream.
- Transaction boundaries.
- Idempotency replay behavior.
- Stale lock conflict behavior.
- Schema migration correctness.
- Roadmap/wiki truthfulness.

Plan 9R is complete only when the grouped review finds no critical or major gaps tied to this design.
