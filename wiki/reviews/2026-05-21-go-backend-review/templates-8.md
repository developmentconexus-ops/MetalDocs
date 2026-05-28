# Module #8 Review — `internal/modules/templates`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:database-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer
**Severity totals:** 9 Critical / 16 High / 13 Medium / 7 Low
**Files reviewed (generated excluded):**
- `domain/{template,version,approval,schemas,audit,errors}.go`
- `application/{service,lifecycle,create,autosave,queries,schema,approval_config,visibility_graph,ports,authz_guc}.go`
- `delivery/http/{handler,routes_lifecycle,routes_create,routes_catalog,routes_query,routes_schema,routes_autosave,errors}.go`
- `repository/{postgres,mappers}.go`

---

## Critical

### C1 — `delivery/http/routes_lifecycle.go:225-248` — `X-Actor-Roles` header fallback → role bypass

`actorRolesFromReq` falls back to the untrusted `X-Actor-Roles` HTTP header when the IAM context is empty. Any caller who can reach a lifecycle handler before the IAM middleware sets context roles (misconfigured route, test harness leak) can self-assert arbitrary roles and bypass the reviewer/approver checks in `Review()` and `Approve()`.

**Recommend:** remove the header fallback entirely. If IAM middleware hasn't populated roles, return an empty slice and let `containsRole` deny. Gate service-to-service calls via a signed service token, not an arbitrary header.

**Fix branch:** `fix/templates-8-actor-roles-c1`

---

### C2 — `application/lifecycle.go:58-60` — `SubmitForReview` calls `authz.Require` without prior `setAuthzGUC`

`SubmitForReview` opens a transaction, then immediately calls `authz.Require` without calling `setAuthzGUC(ctx, tx, ...)` first. The GUC variables (`metaldocs.tenant_id`, `metaldocs.actor_id`) are never set, so RLS/authz policies evaluate against empty or stale context. `CreateTemplate` and `SaveTemplateDraft` both correctly call `setAuthzGUC` before `authz.Require`.

**Recommend:** add `setAuthzGUC(ctx, tx, cmd.TenantID, cmd.ActorUserID)` immediately after `BeginTx` in `SubmitForReview`. Audit all lifecycle methods — `Review`, `Approve`, `ArchiveTemplate`, `PublishTemplateVersion` all have the same gap (see M1).

**Fix branch:** `fix/templates-8-authz-guc-c2-c3`

---

### C3 — `application/lifecycle.go:498-513` — `updateVersionWithAuthz` calls `authz.Require` without `setAuthzGUC`

`updateVersionWithAuthz` (used by `Review` and `Approve` reject paths) calls `authz.Require` without setting GUC context. The helper currently accepts no `tenantID`/`actorID` parameters, so it cannot call `setAuthzGUC` even if corrected.

**Recommend:** extend `updateVersionWithAuthz` signature with `tenantID, actorID string`, call `setAuthzGUC(ctx, tx, tenantID, actorID)` before `authz.Require`. Update all callers.

**Fix branch:** `fix/templates-8-authz-guc-c2-c3`

---

### C4 — `repository/postgres.go:236` — `GetVersionByID` no `tenant_id` predicate → cross-tenant IDOR

`GetVersionByID` fetches any version row by UUID with no tenant guard:
```sql
SELECT ... FROM templates_template_version WHERE id = $1
```
`create.go:159` acknowledges this with a post-fetch equality check — but the data is already returned from the DB before the check fires. Any actor who learns or guesses a version UUID retrieves another tenant's full version record.

**Recommend:** add `AND template_id IN (SELECT id FROM templates_template WHERE tenant_id = $2)`, threading `tenantID` through the signature. Or remove the method in favour of `GetVersion` once it also carries `tenantID` (see C5).

**Fix branch:** `fix/templates-8-tenant-isolation-c4-c5-c6-c7`

---

### C5 — `repository/postgres.go:217` — `GetVersion` no `tenant_id` predicate

`GetVersion` queries by `template_id` + `version_number` only. Called throughout `lifecycle.go` (lines 26, 107, 194, 337) without `tenantID` in the SQL predicate. Tenant isolation is implied by the prior `GetTemplate` call — convention, not enforcement. Any future code path that skips the pre-check gains cross-tenant read access.

**Recommend:** add `tenant_id` parameter to `GetVersion` and enforce it in SQL. Update all 4+ call sites in `lifecycle.go`.

**Fix branch:** `fix/templates-8-tenant-isolation-c4-c5-c6-c7`

---

### C6 — `repository/postgres.go:447` — `GetApprovalConfig` no `tenant_id` predicate → cross-tenant config read

```sql
SELECT ... FROM templates_approval_config WHERE template_id = $1
```

No tenant scope. An actor who supplies a cross-tenant `template_id` reads another tenant's approval role configuration.

**Recommend:** add `AND template_id IN (SELECT id FROM templates_template WHERE tenant_id = $2)`, pass `tenantID` from service.

**Fix branch:** `fix/templates-8-tenant-isolation-c4-c5-c6-c7`

---

### C7 — `repository/postgres.go:518` — `ListAudit` no `tenant_id` predicate → cross-tenant audit log read

```sql
SELECT ... FROM templates_audit_log WHERE template_id = $1 ...
```

No tenant guard. An actor who supplies a cross-tenant `template_id` (bypassing the service-layer `GetTemplate` guard via a race or a future call site) receives that tenant's full audit trail.

**Recommend:** add `AND tenant_id = $4` (column already selected in projection), thread `tenantID` from service.

**Fix branch:** `fix/templates-8-tenant-isolation-c4-c5-c6-c7`

---

### C8 — `repository/postgres.go:495` — `json.Marshal(entry.Details)` error discarded → corrupted audit record

```go
payload, _ := json.Marshal(entry.Details)
```

On marshal failure, `payload` is `nil`. The audit record is written with an empty payload — silent governance data loss.

**Recommend:** `payload, err := json.Marshal(entry.Details); if err != nil { return fmt.Errorf("templates audit: marshal details: %w", err) }`.

**Fix branch:** `fix/templates-8-audit-integrity-c8-c9`

---

### C9 — `repository/postgres.go:514` — `AppendAuditTx` ignores `*sql.Tx` → ghost audit entries on rollback

`AppendAuditTx` accepts a `tx *sql.Tx` parameter (named `_`) and delegates to `AppendAudit`, which uses the pool connection. The audit write is not part of the surrounding transaction. If the outer transaction rolls back, the audit record is already persisted — creating a ghost audit entry for a failed operation. Conversely, if the audit pool write fails, the business transaction has already committed.

**Recommend:** thread `tx` through to the audit writer (requires the writer to accept a `*sql.Tx` executor), or use an outbox table inside the same transaction and flush asynchronously on commit.

**Fix branch:** `fix/templates-8-audit-integrity-c8-c9`

---

## High

### H1 — `delivery/http/handler.go:27-29` — `authz` defaults to no-op when `nil` passed to `New`

If `New(svc, nil)` is called, `h.authz` is set to `func(...) error { return nil }`. All authorization is silently skipped. A future wiring mistake in production silently passes every auth check.

**Recommend:** panic or return an error in `New` if `authz` is nil. Never accept nil as a valid authorizer.

---

### H2 — `delivery/http/routes_catalog.go:23-25` — `listPlaceholderCatalog` has no authz check

`GET /api/v1/templates/placeholder-catalog` is registered with no capability check. Any authenticated user can enumerate system-wide placeholder capabilities.

**Recommend:** add `h.authz(r, tenantID, "*", "template.view")` consistent with other read endpoints, or explicitly document and gate as internal-only.

---

### H3 — `application/autosave.go:33-57` — `PresignTemplateUpload` accepts caller-supplied `StorageKey` → writable presign to arbitrary path

`cmd.StorageKey` is passed directly to `PresignPUT` without any path validation. An authenticated user can supply an arbitrary S3 key (including keys belonging to other tenants or system objects) and receive a writable presigned URL for that path.

**Recommend:** always derive the storage key server-side from the template/version ID (as `PresignAutosave` does). Reject any caller-supplied key that doesn't match the expected prefix or remove the parameter entirely.

---

### H4 — `application/lifecycle.go:329-438` — non-tx path in `PublishTemplateVersion` creates next draft without authz and non-atomically

The `else` branch (when `s.db == nil`) calls `CreateNextVersion` after `UpdateTemplate` and `UpdateVersion` have committed separately. If `CreateNextVersion` fails, the template is published but has no next draft. Additionally, `CreateNextVersion` has no `authz.Require` gate of its own.

**Recommend:** enforce `s.db != nil` at startup for write operations (panic/fatal if nil). Eliminate the non-tx path for lifecycle mutations.

---

### H5 — `application/create.go:30-32` — key-uniqueness pre-check masks real errors with `ErrKeyConflict`

```go
if _, err := s.repo.GetTemplateByKey(...); !errors.Is(err, domain.ErrNotFound) {
    return nil, domain.ErrKeyConflict
}
```

A transient DB error (network blip, timeout) satisfies `!errors.Is(err, ErrNotFound)` and returns `ErrKeyConflict` to the caller — a misleading 409 instead of 500.

**Recommend:** `if err == nil { return nil, domain.ErrKeyConflict }; if !errors.Is(err, domain.ErrNotFound) { return nil, fmt.Errorf(...) }`. Better yet: remove the pre-check and rely solely on the `ON CONFLICT` / `23505` path in `CreateTemplateTx`.

---

### H6 — `application/create.go:183` — `CreateNextVersion` two writes outside transaction → inconsistent state on partial failure

`CreateVersion` + `UpdateTemplate` + `AppendAudit` are three separate non-atomic operations. Failure between them leaves an orphaned version row with stale `latest_version` on the template.

**Recommend:** wrap all three in a single transaction. Require `s.db != nil` for this operation.

---

### H7 — `repository/postgres.go:401` — `updateVersionDraftCAS` no `tenant_id` → cross-tenant draft overwrite

The CAS UPDATE filters only by `id` and `lock_version`. A compromised actor who knows a version UUID can overwrite another tenant's draft storage key and content hash.

**Recommend:** add `AND template_id IN (SELECT id FROM templates_template WHERE tenant_id = $N)`.

---

### H8 — `repository/postgres.go:290,381` — `UpdateVersion` / `UpdateVersionTx` no `tenant_id` predicate

Updates by version ID alone. If a version UUID is resolved without a prior tenant-scoped `GetTemplate` guard, cross-tenant mutation is possible.

**Recommend:** add tenant scope to UPDATE predicate, consistent with C5 fix.

---

### H9 — `repository/postgres.go:156,290,340,381,413` — `RowsAffected()` errors discarded at 5 sites

`UpdateTemplate`, `UpdateVersion`, `UpdateTemplateTx`, `UpdateVersionTx`, `updateVersionDraftCAS` all use `n, _ := res.RowsAffected()`. A driver error produces `n=0`, triggering spurious `ErrNotFound` and hiding the real error.

**Recommend:** `n, err := res.RowsAffected(); if err != nil { return fmt.Errorf("update rows affected: %w", err) }` at each site.

---

### H10 — `application/autosave.go:172` — `presign.Delete` error silently discarded on hash mismatch cleanup

When content-hash mismatch is detected, `s.presign.Delete(ctx, version.DocxStorageKey)` is called but its error is ignored with `_`. A failed deletion leaves an orphaned storage object with no signal.

**Recommend:** log at warn level: `if delErr := s.presign.Delete(...); delErr != nil { slog.Warn("autosave: cleanup failed", "err", delErr) }`.

---

### H11 — `application/lifecycle.go:71-84,136-146,295-308` — `AppendAudit` called post-commit → state committed without guaranteed audit

In `SubmitForReview`, `Review`, and the rejection branch of `Approve`, `AppendAudit` is called after the business transaction commits. If `AppendAudit` fails, the caller receives an error but the state change is already persisted. Caller may retry, causing duplicate state transitions.

**Recommend:** use `AppendAuditTx` inside the transaction (after fixing C9), or accept best-effort and suppress the error with a structured log rather than returning it to trigger retries.

---

### H12 — `domain/version.go:18` — `TemplateVersion` exported, no constructor; ISO segregation invariants bypassable

All approval-flow invariants (`ContentHash` non-empty before submission, valid `PendingApproverRole`, non-negative `LockVersion`) are enforced in service methods, not at the type boundary. Any code constructing a `TemplateVersion` literal bypasses them.

**Recommend:** add `NewDraftVersion(...)` constructor, expose `Transition(next VersionStatus) error` method that validates and assigns in one step.

---

### H13 — `domain/approval.go:17` — `CheckSegregation` accepts bare string `role` → typo-silent default branch

`"Reviewer"` (wrong case) hits the `default: return ErrForbiddenRole` silently — indistinguishable from a legitimate segregation violation.

**Recommend:** `type SegregationRole string` with constants `SegregationRoleReviewer`/`SegregationRoleApprover`. Change function signature accordingly.

---

### H14 — `application/ports.go:11` — 28-method `Repository` interface; `*sql.Tx` leaks into application port

No single consumer uses all 28 methods. The 14 paired `Foo`/`FooTx` variants leak `*sql.Tx` (a concrete infrastructure type) into the application layer, violating dependency inversion.

**Recommend:** split into focused interfaces (`TemplateReader`, `VersionWriter`, `AuditAppender`). Use a `Transactor`/`UnitOfWork` abstraction to hide `*sql.Tx` from the application port.

---

### H15 — `application/service.go:5` — `WithDB` post-construction mutator + 6-site nil-db branching

`WithDB` modifies `s.db` after construction and is not goroutine-safe. The `if s.db != nil` branch is duplicated across 6+ lifecycle methods, each with its own `BeginTx`/`defer Rollback`/`Commit` boilerplate, with subtle authz-check asymmetries between branches.

**Recommend:** make `db *sql.DB` a required constructor parameter. Extract `func (s *Service) withTx(ctx context.Context, fn func(*sql.Tx) error) error` helper. Eliminate the nil-db branching.

---

### H16 — `application/lifecycle.go:71-84` — `AppendAudit` post-commit error returned causes caller retry of already-committed state

See H11. Distinct from H11 in that the error return triggers HTTP 500, and a client retry resubmits a template already in `PendingReview` state, receiving `ErrInvalidTransition`. Silent data-integrity gap for clients that don't handle this error code.

---

## Medium

### M1 — `application/authz_guc.go` — `setAuthzGUC` not called on `Review`, `Approve`, `ArchiveTemplate`, `PublishTemplateVersion`

Broader pattern underlying C2/C3: only `CreateTemplate`, `SaveTemplateDraft`, `CommitAutosave` call `setAuthzGUC`. All other transactional lifecycle methods run authz with empty GUC context.

**Recommend:** call `setAuthzGUC` at the start of every `BeginTx` block, before any `authz.Require` call. Treat this as a checklist item for every new write operation.

---

### M2 — `application/schema.go:20-21` — `UpdateSchemas` reads version without verifying tenant via `GetTemplate`

`UpdateSchemas` calls `GetVersion` (no tenant in SQL per C5) without a prior `GetTemplate` guard. An actor who knows a `version_number` but not the owning tenant can mutate schemas across tenant boundaries.

**Recommend:** add `s.repo.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID)` before `GetVersion`, matching every other mutating service method.

---

### M3 — `application/approval_config.go:52-54` — `UpsertApprovalConfig` writes config then appends audit non-atomically

Config write and audit append are separate operations. Also, `UpsertApprovalConfig` uses its own `containsRole` check instead of `authz.Require`, leaving the DB-level authz tripwire unarmed.

**Recommend:** wrap in transaction; call `setAuthzGUC` + `authz.Require(CapTemplateAdmin)`; use `AppendAuditTx` after C9 fix.

---

### M4 — `application/lifecycle.go:52-69` — `SubmitForReview` reads state before opening transaction; no re-validation inside tx

`GetTemplate`, `GetVersion`, `GetApprovalConfig` are called pre-transaction. A concurrent mutation between these reads and the `UpdateVersionTx` call is not detected.

**Recommend:** re-validate `version.Status` inside the transaction with `SELECT ... FOR UPDATE`.

---

### M5 — `application/create.go:140-205` — `CreateNextVersion` performs 3 separate non-atomic writes

`CreateVersion` → `UpdateTemplate` → `AppendAudit` are three separate operations with no transaction. (Distinct from H6: H6 covers the published-next-draft inconsistency; this covers the audit gap separately.)

**Recommend:** wrap in transaction (requires `s.db` non-nil enforcement — see H15).

---

### M6 — `repository/postgres.go:519` — `ListAudit` uses OFFSET pagination → degrades on busy audit tables

OFFSET-based pagination scans and discards prior rows. On templates with many audit events, this degrades over time.

**Recommend:** cursor pagination via `WHERE occurred_at < $last AND id < $last_id`.

---

### M7 — `repository/postgres.go:99` — `ListTemplates` does not cap `Limit`

`Limit: 0` or a very large value reaches the DB unchecked.

**Recommend:** enforce `max(1, min(limit, 200))` in the repository or service layer.

---

### M8 — `delivery/http/routes_lifecycle.go:235-248` — `actorRolesFromReq` described as "test/service-to-service" but reachable in production

Even if the header fallback is not removed (per C1 fix), there is no environment gate. The comment acknowledges the risk but provides no enforcement.

---

### M9 — `application/service.go:14` — nil `resolvers` survives construction → runtime panic

`resolvers` is variadic, silently nil if omitted. `DetectVisibilityCycle` calls `resolvers.Known()` → nil-pointer panic at request time.

**Recommend:** make `resolvers` a required non-variadic parameter; panic in `New()` if nil with a descriptive message; or use a no-op default registry.

---

### M10 — `domain/schemas.go:3` — `MetadataSchema.RetentionDays` has no lower-bound invariant

`RetentionDays: 0` or negative is accepted silently — nonsensical business value.

**Recommend:** validate `RetentionDays >= 0` in constructor or `Validate() error`.

---

### M11 — `domain/schemas.go:22` — `VisibilityCondition.Op` is free-form string

The set of valid operators is closed but nothing prevents invalid values from being stored.

**Recommend:** `type VisibilityOp string` with named constants + `Valid() bool`.

---

### M12 — `application/lifecycle.go:329` — next-draft construction duplicated across `s.db != nil` / `else` branches

Two code paths build the next draft with different logic (inline struct literal vs `CreateNextVersion`), divergence-prone.

**Recommend:** extract `buildNextDraftVersion(...)` pure factory, call from both branches.

---

### M13 — `delivery/http/errors.go:66` — default `MapErr` branch surfaces raw `err.Error()` to client

Internal driver errors, SQL messages, or wrapped stack details can reach the HTTP response body.

**Recommend:** log underlying error server-side; return only generic `"internal_error"` code to client.

---

## Low

### L1 — `domain/errors.go:10-18` — schema-validation error sentinels missing `"templates: "` namespace prefix

`ErrPlaceholderIDEmpty`, `ErrDuplicatePlaceholderID`, et al. lack the prefix used by all other sentinels in the package.

**Recommend:** add `"templates: "` prefix for consistent grep-ability.

---

### L2 — `domain/template.go:9` — `ID`, `TenantID`, `CreatedBy`, `DocTypeCode`, `Key` are bare `string`

Silently swappable in function calls. At minimum `type TemplateID string` + `type TenantID string`.

---

### L3 — `repository/mappers.go:104-109` — `marshalAuditDetails` defined but never called

Dead code. `AppendAudit` calls `json.Marshal` directly.

**Recommend:** remove the helper or use it (and propagate its error).

---

### L4 — `domain/approval.go:3` — `ApprovalConfig` no constructor; `ApproverRole` can be empty

**Recommend:** `NewApprovalConfig(templateID, approverRole string, reviewerRole *string) (ApprovalConfig, error)`.

---

### L5 — `domain/audit.go:21` — `AuditEvent` no constructor; `OccurredAt` can be zero

12 construction sites scatter audit event assembly.

**Recommend:** `NewAuditEvent(tenantID, templateID, actorID string, versionID *string, action AuditAction, details map[string]any, now time.Time) AuditEvent`.

---

### L6 — `application/visibility_graph.go:27` — `cycle` slice reuses `stack` backing array

`append(stack, id)` may reuse backing array if capacity allows, producing fragile path reporting (single-goroutine DFS, no race, but fragile).

**Recommend:** `cycle := make([]string, len(stack)+1); copy(cycle, stack); cycle[len(stack)] = id`.

---

### L7 — `application/service.go:22-29` — `WithDB` post-construction mutator not goroutine-safe

If `WithDB` is called after the service is shared across goroutines, the write to `s.db` races with concurrent readers.

**Recommend:** accept `db` in `New` constructor directly (see H15).

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/templates-8-actor-roles-c1` | C1 X-Actor-Roles header role bypass | 1st (highest exposure) |
| `fix/templates-8-authz-guc-c2-c3` | C2 SubmitForReview + C3 updateVersionWithAuthz missing setAuthzGUC | 2nd |
| `fix/templates-8-tenant-isolation-c4-c5-c6-c7` | C4 GetVersionByID + C5 GetVersion + C6 GetApprovalConfig + C7 ListAudit no tenant_id | 3rd |
| `fix/templates-8-audit-integrity-c8-c9` | C8 json.Marshal discarded + C9 AppendAuditTx ignores tx | 4th |
| `fix/templates-8-presign-h3` | H3 PresignTemplateUpload arbitrary storage key | 5th |
| `fix/templates-8-rows-affected-h9` | H9 RowsAffected errors at 5 sites | 6th |
