# Feature F0.3 — Plan

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Folder:** `f0.3-tenant-grade-view`
> **Status:** Planning

## Source

- Milestone spec row F0.3: *implement* — `CapDocumentView` no longer narrows a tenant-grade view to area-grade when an area code is present (`documents/approval/application/read_service.go:68`). *validate* — tenant-role-only viewer can read a document carrying a real area code.
- Spec (this folder, `spec.md`): scope is **both** sibling call sites (`:68` LoadInstance and `:114` LoadActiveInstanceByDocument); root-cause family fix; align to declared `ScopeTenant` for `CapDocumentView`.
- Governing references: mission §5 B3, §7 F0.3; ADR 0022 Phase 2 (`iam/domain/capability_scope.go:51`); canonical sibling `internal/modules/documents/application/view_service.go:71`.

## Plan

### Files touched

1. **`internal/modules/documents/approval/application/read_service.go`** — two call-site alignments (~12 lines net delta).
   - `LoadInstance` (lines 45-90): drop `if areaCode == "" { areaCode = "tenant" }`; rewrite `authz.Require(... areaCode)` → `authz.Require(... "tenant")`; replace the stale "COALESCE / Phase 11 F7" comment with one-line "tenant-grade read; area filter intentionally OFF (capability_scope.go: CapDocumentView → ScopeTenant)".
   - `LoadActiveInstanceByDocument` (lines 93-136): delete the `docapp.LoadDocumentAreaCode` call + `areaCode`/`found` locals + coalesce. Rewrite `authz.Require(... areaCode)` → `authz.Require(... "tenant")` with the same one-line comment. The `docapp` package import is still needed only if other consumers remain in this file — verify; remove the import if it becomes unused.
   - `loadInstanceAreaCode` helper (lines 360-384): **unchanged**. Still used by LoadInstance for its `found` existence probe.

2. **`internal/modules/documents/approval/application/read_service_test.go`** — one existing case retargeted.
   - `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` (lines 174-217): currently expects `Require` to be called with area `qa` after `LoadDocumentAreaCode` returns `qa`. Post-fix, `LoadDocumentAreaCode` is gone from this method and `Require` is called with `"tenant"`. Update mock expectations: drop the `SELECT COALESCE(d.process_area_code_snapshot ...` mock; keep the `EXISTS(... iam_user_roles ...)` (system_admin bypass) and `EXISTS(... role_capabilities ...)` (grant) mocks; keep the `ErrCapDenied` assertion (security property unchanged — same deny, new envelope `AreaCode:"tenant"`). Add an assertion that `denied.AreaCode == "tenant"` so the test now pins the contract.

3. **`internal/modules/documents/approval/application/read_service_tenant_grade_view_integration_test.go`** — **new**, `//go:build integration`, real-Postgres proof. Six tests (per spec Validation Gate):
   - `TestLoadInstance_TenantGradeViewer_DocWithAreaCode_Granted` (RED pre-fix, GREEN post-fix).
   - `TestLoadActiveInstanceByDocument_TenantGradeViewer_DocWithAreaCode_Granted` (RED pre-fix, GREEN post-fix).
   - `TestLoadInstance_NoViewGrant_Denied` (asserts `ErrCapDenied{AreaCode:"tenant"}`).
   - `TestLoadActiveInstanceByDocument_NoViewGrant_Denied` (same).
   - `TestLoadInstance_SystemAdmin_Granted`.
   - `TestLoadActiveInstanceByDocument_SystemAdmin_Granted`.

### Test strategy (TDD — failing tests first)

**Fixture pattern** mirrors F0.2's `manual_code_create_integration_test.go` (`seedActorWithGrant`) and the approval factories in `tests/integration/testdb/factory.go` (`NewApprovalInstance` → seeds `documents` → seeds `controlled_documents` → mints process area). Helper inside the new file:

```go
func seedDocWithAreaAndActor(t *testing.T, db *sql.DB, role string) (tenantID, actorID, areaCode, docID, instanceID string) {
    tenant := testdb.NewTenant(t, db)
    user   := testdb.NewUser(t, db, testdb.WithTenant(tenant.ID))
    inst   := testdb.NewApprovalInstance(t, db,
        testdb.WithTenant(tenant.ID), testdb.WithStatus("in_progress"))
    // grant viewer in a DIFFERENT area than the doc's area (tenant-grade should still grant).
    if role != "" {
        otherArea := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenant.ID)).ProcessAreaCode
        testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
            _, err := tx.ExecContext(ctx,
                `INSERT INTO metaldocs.user_process_areas
                   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
                 VALUES ($1, $2::uuid, $3, $4, now(), NULL, $1)`,
                user.ID, tenant.ID, otherArea, role)
            return err
        })
    }
    // resolve doc area for return (the doc carries an area via process_area_code_snapshot).
    var area string
    _ = db.QueryRowContext(ctx,
        `SELECT COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')
           FROM documents d JOIN controlled_documents cd ON cd.id = d.controlled_document_id
          WHERE d.id = $1`, inst.DocumentID).Scan(&area)
    return tenant.ID, user.ID, area, inst.DocumentID, inst.ID
}
```

Granted role: **`viewer`** (curated `role_capabilities` row: `viewer → document.view` per `db/reference-data/0001_product_reference_data.sql:84`). The grant is seeded in an **other** area than the document's — this is the load-bearing assertion: tenant-grade `document.view` must grant regardless of where in the tenant the role is held.

Service wiring per test: real `ApprovalRepository` from `internal/modules/documents/approval/repository/postgres` (cf. existing approval integration tests for the constructor), `platformdb.NewTxRunner(db)`. Context carries `tenant.WithTenantID` + `iamdomain.WithAuthContext` (same idiom as `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` lines 201-202).

### RED proof (before edit)

```
go test -tags integration -run 'TestLoad(Instance|ActiveInstanceByDocument)_TenantGradeViewer_DocWithAreaCode_Granted' \
  ./internal/modules/documents/approval/application/ -v
```

Expected output on pre-fix code: both FAIL with `authz: capability "document.view" denied for actor "…" in area "<real area>"` — because the resolved area is `<real area>` and the viewer's grant is in a *different* area, so the `Require` `upa.area_code = $2` predicate matches nothing. This is exactly mission §5 B3's symptom.

### Implementation (make it green)

`read_service.go` LoadInstance (45-90) — final shape of the authz block:

```go
areaCode, found, err := loadInstanceAreaCode(ctx, tx, tenantID, instanceID)
if err != nil { return fmt.Errorf("read load instance: load area: %w", err) }
if !found { return repository.ErrNoActiveInstance }
_ = areaCode // tenant-grade read; resolver retained for its existence probe.

ctx := authz.WithCapCache(ctx)
// CapDocumentView is tenant-grade (iam/domain/capability_scope.go:51); pass the
// "tenant" sentinel so the area filter is intentionally OFF — mirrors the
// canonical documents/application/view_service.go:71.
if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
    return err
}
```

`read_service.go` LoadActiveInstanceByDocument (93-136) — final shape of the authz block:

```go
actorID := iamdomain.UserIDFromContext(ctx)
if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
    return fmt.Errorf("read load instance by document: %w", err)
}

ctx := authz.WithCapCache(ctx)
// CapDocumentView is tenant-grade (iam/domain/capability_scope.go:51); pass the
// "tenant" sentinel so the area filter is intentionally OFF — mirrors the
// canonical documents/application/view_service.go:71. Missing document/instance
// returns ErrNoActiveInstance via the repo lookup below.
if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
    return err
}

loaded, err := s.repo.LoadActiveInstanceByDocument(ctx, tx, tenantID, documentID)
// … unchanged from here on
```

If `docapp` import becomes unused after the deletion, drop it.

### Ordering

1. Author the new integration test file + helper.
2. Run `go test -tags integration -run 'TenantGradeViewer_DocWithAreaCode_Granted' ./internal/modules/documents/approval/application/` → confirm both **FAIL** with the expected envelope; deny/system_admin tests will already PASS in part (deny path still denies, sysadmin still bypasses) — record what was red and what was already green.
3. Apply the two-call-site alignment in `read_service.go`.
4. Update `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` mocks.
5. Re-run integration → all 6 new tests GREEN.
6. `go test ./internal/modules/documents/approval/application/...` (unit, no tag) → GREEN.
7. `go test -tags integration -count=1 ./internal/modules/documents/approval/...` → GREEN (modulo pre-existing `TestSequenceAllocatorNextAndIncrement_Concurrent` defer carried from F0.2).
8. `go build ./...` and `go test ./...` → GREEN.

## Execution notes

- Two-line semantic change at two sites + one test mock retarget + one new integration file. Too small to fan out to `subagent-driven-development`; implement inline under self-review discipline.
- Industry-standard root-cause discipline: the fix targets the **family** (both divergent call sites), aligns to the **declared** capability scope, and uses the **canonical sibling's idiom** byte-for-byte. No shared-API redesign (HS-2 respected). No symptom patch (no per-caller hack, no scope re-declaration, no schema change).
- Real-DB proof is the gate; if the `testdb` harness is unavailable in this environment, that is recorded honestly in `evidence.md` (the test exists and is correct; the real-provider run is the acceptance gate).
