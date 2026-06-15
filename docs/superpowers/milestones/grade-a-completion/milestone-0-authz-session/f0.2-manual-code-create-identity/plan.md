# Feature F0.2 — manual-code create identity

> **Milestone:** 0  ·  **Folder:** `f0.2-manual-code-create-identity`
> **Status:** Planning

## Source

- Milestone spec row: *implement* — manual-code CD-create branch seeds tx identity so non-admin
  creates pass PEP/PDP. *validate* — non-admin manual-code create succeeds; admin path still works.
- Governing-spec reference: mission §5 B2, §7 F0.2; ADR 0022 Phase 7.

## Plan

**Files touched**
1. `internal/modules/controlleddocuments/application/service.go` — refactor the manual-code branch
   (lines 173-276) per spec §What this feature implements.
2. `internal/modules/controlleddocuments/application/manual_code_create_integration_test.go` —
   **new**, `//go:build integration`, real Postgres.

**Test strategy (TDD — RED first)**

Real-DB integration test, mirroring `tests/integration/testdb` factory patterns + `service_test.go`
construction of `NewControlledDocumentService`. Shared seeding helper:

```go
func seedActorWithCap(t, db, role, cap) (tenantID, actorUserID, area, profileCode string)
```
- `testdb.NewTenant` → tenantID.
- `testdb.NewTaxonomy(WithTenant)` → real (profileCode, processAreaCode); area=tax.ProcessAreaCode.
- `testdb.NewUser(WithTenant)` → actorUserID.
- Insert `user_process_areas` (effective_from=now()) via `SeedWithCaps('[{"cap":"membership.manage"}]', ...)` — role grants the cap via curated `role_capabilities` (e.g. `author` → `controlled_documents.create`).

System-admin variant: `testdb.SeedSystemAdmin(t, db, tenantID, userID, "Sys Admin")`.

**Tests**

1. `TestCreate_ManualCode_NonAdminWithCap_Succeeds` — RED on current code (`ErrActorContextMissing`),
   GREEN after refactor. Seeds `author` (curated holds `controlled_documents.create`); calls service
   `Create` with `ManualCode=&"MANUAL-001"`, `ManualCodeReason=&"reason"`, `ProcessAreaCode=area`.
   Assert: err nil; result.ControlledDocument.Code == "MANUAL-001".
2. `TestCreate_ManualCode_NonAdminWithoutCap_Denied` — seeds a user with NO membership; expect
   `authz.ErrCapDenied`.
3. `TestCreate_ManualCode_SystemAdmin_Succeeds` — seeds `iam_user_roles role_code=system_admin`;
   expect success regardless of `user_process_areas`.

Service construction will need:
- `runner` = a real `platformdb.TxRunner` over the test `*sql.DB`.
- `docs` = `controlleddocumentsinfrastructure.NewPostgresControlledDocumentRepository(db)`.
- `profiles`, `areas` = real postgres repos against the taxonomy seeded by factory.
- `seq` = real sequence allocator (not used by manual path).
- `now` = `time.Now` or fixed `func() time.Time`.
- `tplCheck` = a tiny fake returning a valid published template state when needed (not exercised in
  the no-override base case; the simplest manual create doesn't use override).
- `docInit` = `nil` is OK for the no-override path (auto path requires it; manual path's only
  `docInit` use is `ensureTemplateArtifact` which returns `ErrTemplateArtifactInvariantUnconfigured`
  when `docInit==nil`).

If wiring the real preflight stack is too heavy for these three tests, **prefer the smallest viable
fake for `ensureTemplateArtifact`** rather than dragging in the whole template/storage stack — but
preserve the real DB for `CodeExists`/`CreateTx`/`Require`. Decision recorded in evidence.

**Implementation (make it green)**

Replace lines 270-276:
```go
if err := s.docs.Create(ctx, doc); err != nil {
    return nil, fmt.Errorf("...")
}
```
with:
```go
if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
    if err := authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID); err != nil {
        return fmt.Errorf("controlled_documents: set authz context: %w", err)
    }
    if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), cmd.ProcessAreaCode); err != nil {
        return fmt.Errorf("controlled_documents: authz check manual-code create: %w", err)
    }
    return s.docs.CreateTx(ctx, tx, doc)
}); err != nil {
    return nil, fmt.Errorf("controlled_documents: create controlled document (manual): %w", err)
}
```
Delete the "DELIBERATELY-PRESERVED asymmetry" comment block (lines 174-178). All preflight reads stay
exactly where they are (OFF-tx).

**Ordering**
1. Write 3 integration tests; run with `-tags integration`. Expect: NonAdminWithCap **FAILS** with
   `ErrActorContextMissing` (or wrapped); NonAdminWithoutCap **likely also fails the same way**
   (current code can't even reach the deny — flag in evidence if so); SystemAdmin path's outcome
   under current code is also broken by the same `MustActorID` failure → expect it fails too.
2. Apply the refactor.
3. Re-run integration → all 3 green.
4. `go test ./internal/modules/controlleddocuments/...` (unit) → green (no regression in existing
   `service_test.go` — there is one manual-code unit test, `TestCreate_ManualCode_Succeeds_NoOverride`
   or similar — find it and verify it still works against the fake repo, or update its fake repo if
   `Create` → `CreateTx` switch breaks the fake).
5. `go test ./...` whole-repo green.

## Execution notes

- Single-purpose surgical refactor — direct implementation under review discipline; no SDD fan-out.
- Will likely need to update at least one fake in `controlleddocuments/application/service_test.go`
  to expect `CreateTx` instead of `Create` for the manual path. That is harness alignment, not a
  symptom patch.
