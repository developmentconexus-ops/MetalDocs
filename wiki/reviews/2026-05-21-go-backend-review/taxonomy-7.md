# Module #7 Review — `internal/modules/taxonomy`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:database-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer
**Severity totals:** 5 Critical / 14 High / 13 Medium / 8 Low
**Files reviewed (generated excluded):**
- `domain/area.go`, `domain/family.go`, `domain/profile.go`, `domain/port.go`
- `application/area_service.go`, `application/family_service.go`, `application/profile_service.go`
- `application/governance_logger.go`, `application/audit_governance_adapter.go`
- `delivery/http/handler.go`, `delivery/http/routes_areas.go`, `delivery/http/routes_families.go`, `delivery/http/routes_profiles.go`
- `infrastructure/repository.go`, `infrastructure/family_repository.go`
- `infrastructure/authz_guc.go`, `infrastructure/template_version_checker.go`
- `module.go`

---

## Critical

### C1 — Read-path HTTP handlers missing `authz.Require` gate → any authenticated user reads all tenant taxonomy

Six read handlers have no capability check:
- `delivery/http/routes_areas.go:23` — `listAreas`
- `delivery/http/routes_areas.go:76` — `getArea`
- `delivery/http/routes_profiles.go:49` — `listProfiles`
- `delivery/http/routes_profiles.go:114` — `getProfile`
- `delivery/http/routes_families.go:20` — `listFamilies`
- `delivery/http/routes_families.go:62` — `getFamily`

All write paths (`createArea`, `updateArea`, `createProfile`, `createFamily`, etc.) correctly call `authz.Require`. Read paths have no gate — any authenticated user, regardless of role, can enumerate the full taxonomy (areas, profiles, families) for the tenant.

**Recommend:** add `authz.Require(ctx, capability.TaxonomyManage)` (or a dedicated `TaxonomyRead` capability) to each read handler before the service call. Pattern: write handlers in the same files.

**Fix branch:** `fix/taxonomy-7-authz-reads-c1`

---

### C2 — Repository read methods missing `setAuthzGUC` → RLS context unset for all SELECT queries

All three repository read methods execute SQL with no GUC context:
- `infrastructure/repository.go:22` — `ProfileRepository.GetByCode`
- `infrastructure/repository.go:57` — `ProfileRepository.List`
- `infrastructure/repository.go:203` — `AreaRepository.GetByCode`
- `infrastructure/repository.go:236` — `AreaRepository.List`
- `infrastructure/family_repository.go:29` — `FamilyRepository.GetByCode`
- `infrastructure/family_repository.go:48` — `FamilyRepository.List`

Write paths (`Create`, `Update`) all correctly open a transaction and call `setAuthzGUC` + `authz.Require` before any SQL. Read paths omit both. If RLS SELECT policies depend on `app.tenant_id`/`app.user_id` GUCs (which `authz_guc.go` sets), the policies silently evaluate against empty values, meaning cross-tenant rows may be returned.

**Recommend:** wrap each read method in a transaction, call `setAuthzGUC(ctx, tx)` + `authz.Require(...)`, then execute the SELECT. Pattern: `ProfileRepository.Create` lines 105–117.

**Fix branch:** `fix/taxonomy-7-authz-reads-c1` (same branch — HTTP + repo layer are one fix cluster)

---

### C3 — `infrastructure/template_version_checker.go:13` — no `tenant_id` predicate → cross-tenant IDOR

```sql
SELECT v.is_published, t.doc_type_code
FROM templates_template_version v
JOIN templates_template t ON t.id = v.template_id
WHERE v.id = $1
```

No `tenant_id` filter on either table. Any caller who knows or guesses a template version UUID from another tenant can confirm its existence, published status, and `doc_type_code`.

**Recommend:** add `AND t.tenant_id = $2` with `tenantID` from context. Update `IsPublished` signature to accept `tenantID`. Verify `templates_template` has a `tenant_id` column; add via migration if missing.

**Fix branch:** `fix/taxonomy-7-tenant-isolation-c3`

---

### C4 — `infrastructure/family_repository.go:48` — `FamilyService.Update` lost-update race: read outside transaction

`FamilyService.Update` calls `repo.GetByCode` (non-transactional), modifies fields, then calls `repo.Update` (separate transaction). Between the two calls, a concurrent `Update` on the same family code overwrites both changes silently. The `Deactivate` path correctly uses `GetByCodeForUpdate` inside a transaction, but `Update` does not.

**Recommend:** use `GetByCodeForUpdate` inside a transaction for `FamilyService.Update`, matching `Deactivate`. Or add an optimistic-concurrency version column checked in the UPDATE WHERE clause.

**Fix branch:** `fix/taxonomy-7-toctou-c4-c5`

---

### C5 — `application/area_service.go:67` + `application/profile_service.go:77,115` — read-then-write races without `FOR UPDATE`

Three methods read a domain object then write it back without an enclosing transaction or `FOR UPDATE` lock:
- `AreaService.SetParent`: GetByCode → GetByCode(parent) → ListAncestors (cycle check) → Update. Concurrent `SetParent` calls can both pass the cycle check and create a real cycle in the hierarchy.
- `ProfileService.SetDefaultTemplate`: GetByCode → IsPublished → Update. Concurrent Archive between the read and write accepts a template on an archived profile.
- `ProfileService.Archive`: GetByCode → Update. Concurrent Archive/Update overwrites state.

**Recommend:** wrap each flow in a transaction with `SELECT ... FOR UPDATE` on the primary row before any validation. Pattern: `FamilyService.Deactivate` uses `GetByCodeForUpdate` + transaction.

**Fix branch:** `fix/taxonomy-7-toctou-c4-c5`

---

## High

### H1 — All `govLogger.Log` errors discarded across all services (widespread audit trail silencing)

Every governance event call in the module discards the `Log` error with `_ =`:
- `area_service.go:41` (Create), `:58` (Update)
- `family_service.go:37` (Create), `:62` (Update), `:110` (Deactivate)
- `profile_service.go:50` (Create), `:68` (Update)

`Archive` and `SetParent` in area, and `SetDefaultTemplate` in profile, correctly return the log error. The inconsistency means create/update operations can succeed while their governance records are silently dropped.

**Recommend:** return `govLogger.Log` errors from `Create` and `Update` in all three services, consistent with `Archive`/`SetParent`/`SetDefaultTemplate`. The audit writer persists to DB — a transient failure is not a safe no-op.

---

### H2 — All `json.Marshal` errors discarded across governance event sites

Same sites as H1 plus `profile_service.go:102` (SetDefaultTemplate payload):
- `area_service.go:38` (Create), `:55` (Update)
- `family_service.go:37` (Create), `:61` (Update), `:111` (Deactivate)
- `profile_service.go:46` (Create), `:64` (Update), `:102` (SetDefaultTemplate)

On marshal failure the event is emitted with a `nil` payload — the governance record is silently corrupted.

**Recommend:** propagate the marshal error (fail the operation) or use a `platform/events.MarshalPayload` helper that returns an error. Uniform fix across all three service files.

---

### H3 — `infrastructure/template_version_checker.go:25` — nil `db` silently returns `(false, "", nil)` → misconfiguration invisible

The `c.db == nil` guard returns a success nil error with published=false. `ProfileService.SetDefaultTemplate` interprets this as "version not published" and returns `ErrTemplateNotPublished` to the caller with no indication the checker was never configured. Production misconfiguration produces no alarm.

**Recommend:** remove the nil guard; enforce `db != nil` in `NewTemplateVersionChecker` with an explicit error or panic. A nil DB is a programming error, not an expected runtime state.

---

### H4 — `delivery/http/handler.go:43` — `NewHandler` takes concrete service types, defeating local interfaces

`NewHandler` accepts `*application.ProfileService`, `*application.AreaService`, `*application.FamilyService` directly. The same file defines `profileService`, `areaService`, `familyService` interfaces. Using concrete types makes the handler untestable without full service graphs.

**Recommend:** change constructor parameters to the three local interface types.

---

### H5 — `domain/area.go:8` + `domain/family.go:8` + `domain/profile.go:8` — exported domain structs, no constructors

All three aggregates have fully exported fields with no constructor. All invariants (non-empty Code, positive ReviewIntervalDays, valid TenantID) are comments, not enforcement. Callers (including test helpers) construct zero-value structs and assign fields directly, bypassing all validation.

**Recommend:** add `NewProcessArea(...)`, `NewDocumentFamily(...)`, `NewDocumentProfile(...)` constructors that validate required fields. Unexport mutable fields; expose via methods.

---

### H6 — `domain/port.go:24` — `GovernanceEvent` exported with no constructor + bare string `EventType`/`ResourceType`

`EventType` and `ResourceType` are untyped strings. Any string (including typos) reaches the audit log. `TenantID` and `ActorUserID` are routinely omitted (H7 below), leaving ungroupable audit records.

**Recommend:** define `type EventType string` and `type ResourceType string` with package-level constants. Add `NewGovernanceEvent(...)` constructor that requires non-empty `EventType`, `ResourceType`, `ResourceID`.

---

### H7 — `application/area_service.go:41,58` + `application/profile_service.go:50,68` — governance events omit `TenantID` and `ActorUserID`

`Create` and `Update` governance events in both services omit `TenantID` and `ActorUserID`. `Archive`, `SetParent`, and `SetDefaultTemplate` populate both fields. Audit records for creates and updates cannot be attributed to a tenant or actor.

**Recommend:** extract `tenantID` and `actorID` from context at the top of each service method and populate all `GovernanceEvent` fields consistently.

---

### H8 — `application/family_service.go:31` — `FamilyService.Create` mutates caller-supplied pointer as side effect

`Create` sets `f.IsActive = true` on the caller's `*DocumentFamily` pointer. The caller's struct is modified invisibly. This violates immutability conventions and causes aliasing bugs in tests.

**Recommend:** construct a new `DocumentFamily` value internally (via domain constructor) rather than mutating the argument.

---

### H9 — `delivery/http/routes_areas.go:154` + `routes_profiles.go:241` — raw `pgErr.Message` returned to caller

Postgres check-constraint violations (code `23514`) forward the raw `pgErr.Message` string in the HTTP response. This leaks internal constraint names and column names to clients.

**Recommend:** map `23514` to a generic validation message ("value violates domain constraint") rather than forwarding the driver message.

---

### H10 — `application/area_service.go:33` + `application/profile_service.go:41` — `Create` accepts caller-built struct with no validation

`AreaService.Create` and `ProfileService.Create` do not validate `Code`, `TenantID`, `Name`, `ReviewIntervalDays` before persisting. A zero-value struct passes through to the DB.

**Recommend:** validate required fields at service layer entry (or enforce via domain constructor — see H5).

---

### H11 — `application/profile_service.go:59` — `Update` overwrites immutable fields without fetching existing record

`ProfileService.Update` persists the caller's entire `DocumentProfile`, including `Code` and `CreatedAt`, without reading the current record. The business rule `ErrProfileCodeImmutable` is checked only if a code mismatch is caught, but the service never fetches the stored code to compare. A caller can overwrite `Code` silently.

**Recommend:** fetch the existing record, apply only mutable fields (Name, ReviewIntervalDays, EditableByRole, OwnerUserID), persist the merged result. Pattern: `FamilyService.Update`.

---

### H12 — `application/family_service.go:88` — `tenant.FromContext` called after transaction starts

In `Deactivate`, `tenantID` is resolved from context after `BeginTx`. If context is missing the tenant, the transaction is left open until the deferred rollback fires — but this ordering is fragile and context checks should precede any DB work.

**Recommend:** resolve `tenantID` (and `actorID`) before `BeginTx`.

---

### H13 — `application/profile_service.go:27` — nil `govLogger` guard inconsistent across service constructors

`NewProfileService` panics on nil `govLogger`; `NewAreaService` and `NewFamilyService` do not guard. A nil logger injected into area or family services fails at call time with a nil-pointer dereference at the first governance event.

**Recommend:** add nil guard to `NewAreaService` and `NewFamilyService` matching `NewProfileService`.

---

### H14 — `infrastructure/repository.go:188,356` + `family_repository.go:120,202` — `RowsAffected()` error discarded

Four update methods discard the error from `result.RowsAffected()`:
- `repository.go:188` — ProfileRepository.Update
- `repository.go:356` — AreaRepository.Update
- `family_repository.go:120` — FamilyRepository.Update
- `family_repository.go:202` — FamilyRepository.UpdateTx

A driver error produces zero `rowsAffected`, firing a spurious `ErrNotFound`. The real error is silently dropped.

**Recommend:** `n, err := result.RowsAffected(); if err != nil { return fmt.Errorf("update rows affected: %w", err) }`.

---

## Medium

### M1 — `infrastructure/repository.go:57,236` — `List` queries have no `LIMIT`

`ProfileRepository.List` and `AreaRepository.List` select all rows. Large tenants stream unbounded result sets into memory.

**Recommend:** add keyset pagination or a hard cap (`LIMIT 1000`) with a documentation note. Pass `maxRows` from service layer.

---

### M2 — `infrastructure/repository.go:363` — recursive CTE in `ListAncestors` has no depth guard

The recursive term has no `LIMIT` or depth counter. A corrupt parent_code chain (pre-existing data issue) causes infinite CTE expansion until the DB exhausts memory or stack.

**Recommend:** add `LIMIT 50` (or known max tree depth) after the final SELECT.

---

### M3 — `domain/port.go:34` — `FamilyTx` interface forces a type assertion in every transactional repository method

Infrastructure type-asserts `tx.(familyTx)` at every Tx-suffixed call site. Any mock implementation panics at the assertion. The domain port leaks an infrastructure concern.

**Recommend:** pass `*sql.Tx` directly (narrower signature) or pass tenant/actor via context so the transactional variant does not need the raw tx through the domain port.

---

### M4 — `domain/port.go:34-50` — `FamilyRepository` interface bloat (plain + Tx variants of same operations)

`GetByCode`/`GetByCodeForUpdate`, `Update`/`UpdateTx`, `HasActiveProfiles`/`HasActiveProfilesTx` all coexist in one interface. This leaks transactional concerns into the domain port and makes mocking expensive.

**Recommend:** split into `FamilyReader` and `FamilyWriter`/`FamilyTxWriter`.

---

### M5 — `application/governance_logger.go` + `application/audit_governance_adapter.go` — dual logger with no deprecation plan

Both `DBGovernanceLogger` and `AuditGovernanceAdapter` exist. `module.go` picks between them at wire-up. If `AuditGovernanceAdapter` is canonical, `DBGovernanceLogger` should be marked deprecated and scheduled for removal.

**Recommend:** add `// Deprecated: use AuditGovernanceAdapter` comment to `DBGovernanceLogger` and open a ticket to remove it.

---

### M6 — `domain/family.go:13` — `IsActive` is a public `bool` field

Any package can set `f.IsActive = false` without going through `Deactivate()`, bypassing the `ErrFamilyHasProfiles` guard. The other aggregates expose `IsActive()` as a method.

**Recommend:** unexport the field; expose via `IsActive() bool` method.

---

### M7 — `infrastructure/authz_guc.go:19` — missing-actor error is an untyped string; no sentinel for `errors.Is`

`fmt.Errorf("taxonomy: actor context missing")` cannot be matched by callers with `errors.Is`.

**Recommend:** `var ErrActorMissing = errors.New("taxonomy: actor context missing")` and return it directly.

---

### M8 — `application/area_service.go:50` — `AreaService.Update` allows overwriting immutable fields

Same issue as H11 for profiles: `Update` persists the caller's full struct without fetching the existing record. `Code` and `TenantID` can be silently overwritten.

**Recommend:** fetch existing record, apply mutable fields only, persist merged result.

---

### M9 — `delivery/http/routes_areas.go:43` + `routes_profiles.go:70` — `OwnerUserID`/`DefaultApproverRole`/`EditableByRole` accepted without existence check

Caller-supplied user/role IDs are stored without verifying they exist in the IAM module for the current tenant. Results in dangling references.

**Recommend:** validate against IAM, or explicitly handle the FK violation (`23503`) with a descriptive error response.

---

### M10 — `delivery/http/routes_profiles.go:199` — manual `{}` response instead of shared `writeJSON` helper

`setDefaultTemplate` success path writes `{}` manually, diverging from all other handlers.

**Recommend:** `httpresponse.WriteJSON(w, http.StatusOK, struct{}{})` or `w.WriteHeader(http.StatusNoContent)`.

---

### M11 — `infrastructure/family_repository.go:146` — `GetByCodeForUpdate` has no authz guard

The `FOR UPDATE` SELECT runs without `setAuthzGUC` or `authz.Require`. If used outside the `UpdateTx` flow, the lock is held without an authorization check.

**Recommend:** move the authz check before `GetByCodeForUpdate`, or guard inside the method.

---

### M12 — `application/area_service.go:98` — `SetParent` governance payload is empty `{}`

The event carries no information about what changed (old vs new parent code). Audit records for parent changes are unactionable.

**Recommend:** include `{"old_parent": oldCode, "new_parent": newParentCode}` in the payload.

---

### M13 — `module.go:27` — `govLogger` falls back to `DBGovernanceLogger` silently in production

If `AuditWriter` is nil, the module wires `DBGovernanceLogger` with no warning. A production misconfiguration produces no alarm.

**Recommend:** log a prominent warning (`slog.Warn`) when the fallback path is taken, or fail startup if `AuditWriter` is required.

---

## Low

### L1 — `domain/area.go:16` — `DefaultApproverRole *string` should be `*RoleID`

Same primitive-obsession as profile's `EditableByRole`. Role identifiers across domain types should share a named type.

---

### L2 — `domain/port.go:44` — `HasActiveProfiles` takes `tenantID` but `GetByCode` does not — inconsistent scoping convention

Document the invariant (families are platform-global) or align the signatures.

---

### L3 — `delivery/http/routes_areas.go` — uses string literals `"AREA_NOT_FOUND"` while `routes_profiles.go` uses typed `codeTax*` constants

**Recommend:** apply typed constants in `routes_areas.go` for typo safety.

---

### L4 — `delivery/http/routes_areas.go:143` — `writeAreaError` default case has no server-side log

Unrecognised errors reach the default branch silently as HTTP 500 with no log entry. `writeFamilyError` and `writeProfileError` have the same gap.

**Recommend:** add `slog.Error("taxonomy error", "err", err)` in each default case.

---

### L5 — `application/family_service.go:34` — nil `govLogger` guard duplicated across three methods instead of constructor

Nil check at call time rather than construction time; nil is not a valid value.

---

### L6 — `infrastructure/template_version_checker.go:32` — `sql.ErrNoRows` returns `(false, "", nil)`; caller cannot distinguish "version not found" from "version exists but not published"

**Recommend:** return a distinct `ErrTemplateVersionNotFound` sentinel.

---

### L7 — `application/audit_governance_adapter.go:26` — `json.Marshal` error discarded on fallback empty payload

Low risk (static map cannot fail), but inconsistent with the intended pattern.

---

### L8 — `delivery/http/routes_profiles.go:23-31` — `var writeJSON = ...` package-level mutable alias instead of direct call

Recommend: call `httpresponse.WriteJSON` / `httpresponse.WriteError` directly throughout.

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/taxonomy-7-authz-reads-c1` | C1 (6 read handlers) + C2 (repo read authz) | 1st (highest exposure) |
| `fix/taxonomy-7-tenant-isolation-c3` | C3 template_version_checker no tenant_id | 2nd |
| `fix/taxonomy-7-toctou-c4-c5` | C4 FamilyUpdate lost-update + C5 SetParent/SetDefaultTemplate/Archive races | 3rd |
| `fix/taxonomy-7-audit-trail-h1-h2` | H1 govLogger.Log discards + H2 json.Marshal discards | 4th |
| `fix/taxonomy-7-domain-constructors-h5-h6` | H5 domain structs no constructors + H6 GovernanceEvent type safety | 5th (ripples through callers) |
