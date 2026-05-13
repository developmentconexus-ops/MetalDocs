# Plan 9R Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the unfinished Plan 9 transactional, idempotency, optimistic-lock, and template workflow-alignment work without pulling in audit hash-chain or Plan 10 cleanup scope.

**Architecture:** Keep changes local to the affected modules. Use transaction-aware repository methods that accept `*sql.Tx`, reuse `internal/platform/idempotency` for HTTP replay, and keep OpenAPI/generated wrapper contracts aligned before changing handler behavior.

**Tech Stack:** Go, `database/sql`, PostgreSQL migrations, `net/http`, oapi-codegen v2, OpenAPI 3.0.3, existing MetalDocs RFC 9457/problem and idempotency helpers.

---

## File Structure

Contract/API files:
- Modify: `api/openapi/v1/openapi.yaml` - add `Idempotency-Key` header parameters and 412 response.
- Regenerate: `internal/modules/templates_v2/api/api.gen.go` - generated templates contract.
- Regenerate: `internal/modules/documents/api/api.gen.go` - generated documents contract, even though documents remains partially raw.

Auth files:
- Modify: `internal/modules/auth/application/service.go` - begin one outer tx for `CreateUser`.
- Modify: `internal/modules/auth/domain/port.go` or equivalent auth port file if present; otherwise update the local repository interface in `service.go`.
- Modify: `internal/modules/auth/infrastructure/postgres/repository.go` - add `CreateUserTx`.
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` - add `ReplaceUserRolesTx`.
- Test: `internal/modules/auth/application/service_test.go` and/or `internal/modules/auth/infrastructure/postgres/repository_test.go`.

Documents files:
- Modify: `internal/modules/documents/delivery/http/handler.go` - wire finalize idempotency.
- Modify: `internal/modules/documents/delivery/http/handler_test.go` - add finalize idempotency tests.
- Add: `migrations/0191_document_placeholder_values_revision_fk.sql` - forward FK correction.
- Test: `internal/modules/documents/repository/fillin_repository_integration_test.go` if the existing test harness can validate FK behavior.

Templates files:
- Modify: `internal/modules/templates_v2/application/ports.go` - add tx-aware repository methods needed by create/lifecycle.
- Modify: `internal/modules/templates_v2/application/create.go` - single tx for template creation.
- Modify: `internal/modules/templates_v2/application/lifecycle.go` - single tx for submit/review/approve/publish paths as required.
- Modify: `internal/modules/templates_v2/application/autosave.go` - pass expected lock version to repository.
- Modify: `internal/modules/templates_v2/domain/version.go` or `errors.go` - add stale lock domain error if absent.
- Modify: `internal/modules/templates_v2/repository/postgres.go` - add tx-aware writes, OCC update, and lock version increment.
- Modify: `internal/modules/templates_v2/delivery/http/routes_generated.go` and `routes_lifecycle.go` - require idempotency keys and map stale lock to 412.
- Modify: `internal/modules/templates_v2/delivery/http/errors.go` or equivalent error mapper - add 412 mapping.
- Modify: `internal/modules/iam/domain/model.go` - add `CapTemplateReview`.
- Add: `migrations/0192_template_review_capability.sql` - seed `template.review`.
- Test: `internal/modules/templates_v2/application/*_test.go`, `internal/modules/templates_v2/delivery/http/*_test.go`, `internal/modules/templates_v2/repository/*_test.go`.

Taxonomy files:
- Modify: `internal/modules/taxonomy/domain/port.go` - add tx-aware family methods.
- Modify: `internal/modules/taxonomy/application/family_service.go` - single tx deactivate flow.
- Modify: `internal/modules/taxonomy/infrastructure/family_repository.go` - row lock and tenant-scoped active-profile check.
- Modify: `internal/modules/taxonomy/delivery/http/routes_families.go`, `routes_profiles.go`, `routes_areas.go` - finish 23505 to 409 mapping.
- Test: `internal/modules/taxonomy/application/family_service_test.go`, `internal/modules/taxonomy/infrastructure/family_repository_test.go`, `internal/modules/taxonomy/delivery/http/*_test.go`.

Docs/verification files:
- Modify: `wiki/backlog/roadmap.md` - add Plan 9R status and keep audit T-004 open.
- Modify: affected backlog/module docs only for rows actually closed.
- Modify: `wiki/README.md` if module doc summaries change.

---

## Parallel Execution Model

Task 1 is shared contract prep and should land first. After Task 1, Tasks 2, 3, 4, and 5 can run in parallel with disjoint ownership:

- Worker A owns auth files only.
- Worker B owns documents files and migration `0191`.
- Worker C owns templates_v2, IAM capability, migration `0192`, and generated templates API.
- Worker D owns taxonomy files only.

Task 6 runs after all workstreams merge.

---

### Task 1: Contract Prep and Route Truth

**Files:**
- Modify: `api/openapi/v1/openapi.yaml`
- Regenerate: `internal/modules/templates_v2/api/api.gen.go`
- Regenerate: `internal/modules/documents/api/api.gen.go`

- [ ] **Step 1: Build the Plan 9R route truth table**

Record this table in the implementation notes or PR body before editing the spec:

```markdown
| Module | Method | Runtime path | Runtime handler | Spec path | OperationId | Generated method | Plan 9R change |
|---|---|---|---|---|---|---|---|
| documents | POST | /api/v2/documents/{id}/finalize | Handler.finalizeDocument | /api/v2/documents/{id}/finalize | finalizeDocumentV2 | FinalizeDocumentV2 | Add required Idempotency-Key |
| templates_v2 | POST | /api/v2/templates | Handler.CreateTemplateV2 | /api/v2/templates | createTemplateV2 | CreateTemplateV2 | Add required Idempotency-Key |
| templates_v2 | POST | /api/v2/templates/{id}/versions/{n}/publish | Handler.PublishTemplateVersionV2 | same | publishTemplateVersionV2 | PublishTemplateVersionV2 | Add required Idempotency-Key |
| templates_v2 | POST | /api/v2/templates/{id}/versions/{n}/submit | Handler.SubmitTemplateVersionV2 | same | submitTemplateVersionV2 | SubmitTemplateVersionV2 | Add required Idempotency-Key |
| templates_v2 | POST | /api/v2/templates/{id}/versions/{n}/review | Handler.ReviewTemplateVersionV2 | same | reviewTemplateVersionV2 | ReviewTemplateVersionV2 | Add required Idempotency-Key |
| templates_v2 | POST | /api/v2/templates/{id}/versions/{n}/approve | Handler.ApproveTemplateVersionV2 | same | approveTemplateVersionV2 | ApproveTemplateVersionV2 | Add required Idempotency-Key |
| templates_v2 | PUT | /api/v2/templates/{id}/versions/{n}/draft | Handler.SaveTemplateDraftV2 | same | saveTemplateDraftV2 | SaveTemplateDraftV2 | Add 412 response |
```

- [ ] **Step 2: Add OpenAPI header parameters**

In `api/openapi/v1/openapi.yaml`, add this parameter block to each Plan 9R POST operation listed above:

```yaml
        - name: Idempotency-Key
          in: header
          required: true
          schema: { type: string, format: uuid }
```

For `saveTemplateDraftV2`, add a `412` response:

```yaml
        '412':
          description: stale expected_lock_version
```

- [ ] **Step 3: Run OpenAPI lint**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
```

Expected: no new lint errors from the changed operations.

- [ ] **Step 4: Regenerate affected API packages**

Run:

```powershell
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/templates_v2/api/...
go generate ./internal/modules/documents/api/...
```

Expected: generated files update only from the OpenAPI changes.

- [ ] **Step 5: Inspect generated signatures**

Run:

```powershell
rg -n "CreateTemplateV2|PublishTemplateVersionV2|SubmitTemplateVersionV2|ReviewTemplateVersionV2|ApproveTemplateVersionV2|SaveTemplateDraftV2|FinalizeDocumentV2" internal/modules/templates_v2/api/api.gen.go internal/modules/documents/api/api.gen.go
```

Expected: method names remain unchanged.

- [ ] **Step 6: Commit contract prep**

Run:

```powershell
git add api/openapi/v1/openapi.yaml internal/modules/templates_v2/api/api.gen.go internal/modules/documents/api/api.gen.go
git commit -m "api: add Plan 9R idempotency and OCC contract"
```

---

### Task 2: Auth Atomic CreateUser

**Files:**
- Modify: `internal/modules/auth/application/service.go`
- Modify: `internal/modules/auth/infrastructure/postgres/repository.go`
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`
- Test: `internal/modules/auth/application/service_test.go`

- [ ] **Step 1: Write failing service test**

Add a test that uses fakes where `CreateUserTx` succeeds and `ReplaceUserRolesTx` fails, then asserts the service returns the role error and the fake tx rolled back.

```go
func TestService_CreateUser_RollsBackIdentityWhenRoleAssignmentFails(t *testing.T) {
	db := newAuthTxFakeDB(t)
	repo := &fakeAuthRepoTx{db: db}
	roleAdmin := &fakeRoleAdminTx{err: errors.New("insert iam role: forced")}
	svc := New(repo, nil, roleAdmin, nil)

	err := svc.CreateUser(context.Background(), "u1", "alice", "a@example.com", "Alice", "Password123!", tenant.DevTenantID, []iamdomain.Role{iamdomain.RoleAuthor}, "admin")

	if err == nil || !strings.Contains(err.Error(), "forced") {
		t.Fatalf("CreateUser error = %v, want forced role error", err)
	}
	if !db.rolledBack {
		t.Fatal("expected outer transaction rollback")
	}
	if db.committed {
		t.Fatal("did not expect commit after role assignment failure")
	}
}
```

- [ ] **Step 2: Run the failing auth test**

Run:

```powershell
go test ./internal/modules/auth/application -run TestService_CreateUser_RollsBackIdentityWhenRoleAssignmentFails -count=1
```

Expected: fail because the transaction-aware path does not exist yet.

- [ ] **Step 3: Add transaction-aware interfaces**

Add narrow optional interfaces in `service.go`:

```go
type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type createUserTxRepository interface {
	CreateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.CreateUserParams) error
}

type replaceUserRolesTxRepository interface {
	ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, roles []iamdomain.Role, assignedBy string) error
}
```

Use these only when both repositories support the tx path.

- [ ] **Step 4: Implement repository tx methods**

In `auth/infrastructure/postgres/repository.go`, extract the body of `CreateUser` into:

```go
func (r *Repository) CreateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.CreateUserParams) error {
	if err := ensureUniqueIdentity(ctx, tx, params.UserID, params.Username, params.Email); err != nil {
		return err
	}
	const insertIdentity = `
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, NULL, 0, NULL, NOW(), NOW())
`
	if _, err := tx.ExecContext(ctx, insertIdentity, params.UserID, params.Username, params.Email, params.DisplayName, params.IsActive, params.PasswordHash, params.PasswordAlgo, params.MustChangePassword); err != nil {
		return fmt.Errorf("insert auth identity: %w", err)
	}
	return nil
}
```

Keep `CreateUser` as a wrapper that begins, calls `CreateUserTx`, and commits.

- [ ] **Step 5: Implement IAM role tx method**

In `role_admin_repository.go`, extract `ReplaceUserRoles` into:

```go
func (r *RoleAdminRepository) ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, roles []iamdomain.Role, assignedBy string) error {
	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("require iam user.manage authorization: %w", err)
	}
	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`, tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}
	var lastRole string
	for _, role := range roles {
		if code := strings.TrimSpace(string(role)); code != "" {
			lastRole = code
		}
	}
	if lastRole == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, lastRole, assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}
	return nil
}
```

Keep `ReplaceUserRoles` as a wrapper.

- [ ] **Step 6: Use one outer transaction in service**

In `CreateUser`, when repo and role admin support tx methods, do:

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("begin create user tx: %w", err)
}
defer func() { _ = tx.Rollback() }()
if err := repoTx.CreateUserTx(ctx, tx, params); err != nil {
	return err
}
if err := roleTx.ReplaceUserRolesTx(ctx, tx, userID, displayName, tenantID, roles, createdBy); err != nil {
	return err
}
return tx.Commit()
```

If the current service does not own `*sql.DB`, add a narrow constructor/field only for the postgres path instead of changing all callers.

- [ ] **Step 7: Verify auth**

Run:

```powershell
go test ./internal/modules/auth/... ./internal/modules/iam/infrastructure/postgres/... -count=1
```

Expected: pass.

- [ ] **Step 8: Commit auth workstream**

Run:

```powershell
git add internal/modules/auth internal/modules/iam/infrastructure/postgres
git commit -m "fix(auth): make managed user creation atomic"
```

---

### Task 3: Documents Finalize Idempotency and Placeholder FK

**Files:**
- Modify: `internal/modules/documents/delivery/http/handler.go`
- Modify: `internal/modules/documents/delivery/http/handler_test.go`
- Add: `migrations/0191_document_placeholder_values_revision_fk.sql`

- [ ] **Step 1: Write failing finalize idempotency tests**

Add tests in `handler_test.go`:

```go
func TestFinalizeDocument_RequiresIdempotencyKey(t *testing.T) {
	h := newFinalizeHandlerWithSubmit(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/documents/doc-1/finalize", nil)
	req.SetPathValue("id", "doc-1")
	req = req.WithContext(testTenantAndActorContext(req.Context()))
	rr := httptest.NewRecorder()

	h.finalizeDocument(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestFinalizeDocument_IdempotencyReplayReturnsCachedCreated(t *testing.T) {
	h := newFinalizeHandlerWithSubmit(t)
	key := "11111111-1111-4111-8111-111111111111"

	first := performFinalize(t, h, "doc-1", key)
	second := performFinalize(t, h, "doc-1", key)

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("statuses = %d/%d, want 201/201", first.Code, second.Code)
	}
	if second.Header().Get("Idempotent-Replay") != "true" {
		t.Fatalf("missing replay header")
	}
}
```

- [ ] **Step 2: Run failing documents handler tests**

Run:

```powershell
go test ./internal/modules/documents/delivery/http -run "TestFinalizeDocument_.*Idempotency" -count=1
```

Expected: fail because finalize does not require/read the header.

- [ ] **Step 3: Wire idempotency helper around finalize**

Add a helper in `handler.go`:

```go
func (h *Handler) finalizeIdempotencyStore() *idempotency.Store {
	if h.db == nil {
		return nil
	}
	return idempotency.New(h.db, "POST /api/v2/documents/{id}/finalize")
}
```

At the top of the submit-enabled finalize path, before the draft-state query, check:

```go
idempKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
if idempKey == "" {
	httpresponse.WriteError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header required")
	return
}
store := h.finalizeIdempotencyStore()
payloadHash := idempotency.HashPayload([]byte(docID))
if replay, err := store.CheckReplay(r.Context(), tenantID, actorID, idempKey, payloadHash); err != nil {
	if errors.Is(err, idempotency.ErrConflict) {
		httpresponse.WriteError(w, http.StatusUnprocessableEntity, "IDEMPOTENCY_KEY_REUSED", "Idempotency-Key reused with different payload")
		return
	}
	httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "idempotency check failed")
	return
} else if replay != nil {
	w.Header().Set("Idempotent-Replay", "true")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(replay.Status)
	_, _ = w.Write(replay.Body)
	return
}
```

After building the success response bytes, call:

```go
body, _ := json.Marshal(map[string]string{"instanceId": result.InstanceID})
_ = store.RecordReplay(r.Context(), tenantID, actorID, idempKey, payloadHash, http.StatusCreated, body)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_, _ = w.Write(body)
```

If `HashPayload` is unexported, add a small exported helper in `internal/platform/idempotency` with a unit test, or use the middleware directly by wrapping the route in `RegisterRoutes`.

- [ ] **Step 4: Add placeholder FK migration**

Create `migrations/0191_document_placeholder_values_revision_fk.sql`:

```sql
-- Plan 9R: document_placeholder_values.revision_id points to document_revisions(id).

ALTER TABLE document_placeholder_values
  DROP CONSTRAINT IF EXISTS document_placeholder_values_revision_id_fkey;

ALTER TABLE document_placeholder_values
  ADD CONSTRAINT document_placeholder_values_revision_id_fkey
  FOREIGN KEY (revision_id) REFERENCES document_revisions(id) ON DELETE CASCADE;
```

- [ ] **Step 5: Verify documents**

Run:

```powershell
go test ./internal/platform/idempotency ./internal/modules/documents/delivery/http ./internal/modules/documents/repository -count=1
```

Expected: pass.

- [ ] **Step 6: Commit documents workstream**

Run:

```powershell
git add internal/platform/idempotency internal/modules/documents/delivery/http migrations/0191_document_placeholder_values_revision_fk.sql
git commit -m "fix(documents): make finalize idempotent"
```

---

### Task 4: Templates v2 Transactions, Idempotency, OCC, Review Capability

**Files:**
- Modify: `internal/modules/templates_v2/application/ports.go`
- Modify: `internal/modules/templates_v2/application/create.go`
- Modify: `internal/modules/templates_v2/application/lifecycle.go`
- Modify: `internal/modules/templates_v2/application/autosave.go`
- Modify: `internal/modules/templates_v2/domain/version.go`
- Modify: `internal/modules/templates_v2/repository/postgres.go`
- Modify: `internal/modules/templates_v2/delivery/http/routes_generated.go`
- Modify: `internal/modules/templates_v2/delivery/http/routes_lifecycle.go`
- Modify: `internal/modules/iam/domain/model.go`
- Add: `migrations/0192_template_review_capability.sql`

- [ ] **Step 1: Write failing OCC test**

Add or update repository/service test:

```go
func TestSaveTemplateDraft_StaleLockVersionReturnsConflict(t *testing.T) {
	repo := newFakeRepoWithVersion(&domain.TemplateVersion{
		ID: "v1", TemplateID: "tpl1", VersionNumber: 1, Status: domain.VersionStatusDraft, LockVersion: 2,
	})
	svc := New(repo, fakePresigner{}, fixedClock{}, fixedUUID{})

	err := svc.SaveTemplateDraft(context.Background(), SaveTemplateDraftCmd{
		TenantID: "tenant-1", ActorUserID: "author-1", TemplateID: "tpl1", VersionNumber: 1,
		ExpectedLockVersion: 1, DocxStorageKey: "docx", SchemaStorageKey: "schema",
		DocxContentHash: "hash-a", SchemaContentHash: "hash-b",
	})

	if !errors.Is(err, domain.ErrStaleLockVersion) {
		t.Fatalf("err = %v, want ErrStaleLockVersion", err)
	}
}
```

- [ ] **Step 2: Write failing tx atomicity tests**

Add tests for create and publish:

```go
func TestCreateTemplate_RollsBackWhenVersionInsertFails(t *testing.T) {
	db := newTemplatesTxFakeDB(t)
	repo := &fakeRepo{createVersionErr: errors.New("forced version failure")}
	svc := New(repo, fakePresigner{}, fixedClock{}, fixedUUID{}).WithDB(db)

	_, err := svc.CreateTemplate(context.Background(), CreateTemplateCmd{
		TenantID: "tenant-1", ActorUserID: "author-1", Key: "tpl", Name: "Template",
		Visibility: domain.VisibilityPublic, ApproverRole: "approver",
	})

	if err == nil || !db.rolledBack {
		t.Fatalf("expected rollback on version insert failure, err=%v", err)
	}
}
```

- [ ] **Step 3: Write failing HTTP idempotency tests**

For each POST route group, add one representative test first for create and publish:

```go
func TestCreateTemplateV2_RequiresIdempotencyKey(t *testing.T) {
	h := newTemplateHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/templates", strings.NewReader(`{"key":"k","name":"Name"}`))
	req = req.WithContext(testTenantActorContext(req.Context()))
	rr := httptest.NewRecorder()

	h.CreateTemplateV2(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", rr.Code)
	}
}
```

- [ ] **Step 4: Run failing templates tests**

Run:

```powershell
go test ./internal/modules/templates_v2/application ./internal/modules/templates_v2/delivery/http -run "StaleLock|RollsBack|Idempotency" -count=1
```

Expected: fail.

- [ ] **Step 5: Add domain lock error and field**

If `TemplateVersion` lacks `LockVersion`, add:

```go
var ErrStaleLockVersion = errors.New("template version lock version is stale")

type TemplateVersion struct {
	// existing fields
	LockVersion int
}
```

Update scan functions in `repository/postgres.go` to read `lock_version`.

- [ ] **Step 6: Add tx-aware repository methods**

Extend `Repository` with:

```go
CreateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error
UpsertApprovalConfigTx(ctx context.Context, tx *sql.Tx, c *domain.ApprovalConfig) error
ObsoletePreviousPublishedTx(ctx context.Context, tx *sql.Tx, templateID, keepVersionID string) error
AppendAuditTx(ctx context.Context, tx *sql.Tx, e *domain.AuditEvent) error
```

Implement each by using `tx.ExecContext` and preserving the existing non-tx wrapper behavior.

- [ ] **Step 7: Implement OCC update**

Add a repository method:

```go
func (r *Repository) UpdateVersionWithExpectedLock(ctx context.Context, v *domain.TemplateVersion, expected int) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}
	const q = `
UPDATE templates_v2_template_version
SET status = $2, docx_storage_key = $3, content_hash = $4,
    metadata_schema = $5, placeholder_schema = $6,
    lock_version = lock_version + 1
WHERE id = $1 AND lock_version = $7`
	res, err := r.db.ExecContext(ctx, q, v.ID, string(v.Status), v.DocxStorageKey, v.ContentHash, metadataJSON, placeholderJSON, expected)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrStaleLockVersion
	}
	v.LockVersion = expected + 1
	return nil
}
```

Use the full column set from the existing `UpdateVersion` so lifecycle fields are not lost.

- [ ] **Step 8: Use one tx for create**

In `CreateTemplate`, replace the current split flow with:

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
	return nil, fmt.Errorf("templates create: begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()
if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateCreate), "tenant"); err != nil {
	return nil, fmt.Errorf("templates create: authz: %w", err)
}
if err := s.repo.CreateTemplateTx(ctx, tx, template); err != nil {
	return nil, err
}
if err := s.repo.CreateVersionTx(ctx, tx, version); err != nil {
	return nil, err
}
if err := s.repo.UpsertApprovalConfigTx(ctx, tx, &domain.ApprovalConfig{TemplateID: template.ID, ApproverRole: cmd.ApproverRole, ReviewerRole: cmd.ReviewerRole}); err != nil {
	return nil, err
}
if err := s.repo.AppendAuditTx(ctx, tx, createdAuditEvent); err != nil {
	return nil, err
}
if err := tx.Commit(); err != nil {
	return nil, err
}
```

- [ ] **Step 9: Use one tx for approve and publish**

Move `ObsoletePreviousPublished` and `CreateNextVersion` internals into the same tx. Do not call non-tx `CreateNextVersion` from `PublishTemplateVersion`; instead create a private helper:

```go
func (s *Service) createNextVersionInTx(ctx context.Context, tx *sql.Tx, template *domain.Template, source *domain.TemplateVersion, cmd CreateVersionCmd) (*domain.TemplateVersion, error)
```

Use `CreateVersionTx`, `UpdateTemplateTx`, `UpdateVersionTx`, `ObsoletePreviousPublishedTx`, and `AppendAuditTx` before commit.

- [ ] **Step 10: Add typed review capability and seed migration**

In `internal/modules/iam/domain/model.go`:

```go
CapTemplateReview Capability = "template.review"
```

Create `migrations/0192_template_review_capability.sql`:

```sql
INSERT INTO metaldocs.capabilities (code, description)
VALUES ('template.review', 'Review template versions')
ON CONFLICT (code) DO NOTHING;

INSERT INTO metaldocs.role_capabilities (role_code, capability)
VALUES
  ('approver', 'template.review'),
  ('system_admin', 'template.review')
ON CONFLICT DO NOTHING;
```

- [ ] **Step 11: Require idempotency in templates HTTP**

Add a handler helper:

```go
func requireIdempotencyKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header required")
		return "", false
	}
	return key, true
}
```

Use the shared store per route template. Replay before invoking the service, record successful 2xx responses after service success.

- [ ] **Step 12: Map stale lock to 412**

In the templates error mapper:

```go
case errors.Is(err, domain.ErrStaleLockVersion):
	writeErr(w, http.StatusPreconditionFailed, "CONCURRENT_MODIFICATION", "template version changed; refresh and retry")
```

- [ ] **Step 13: Verify templates**

Run:

```powershell
go test ./internal/modules/templates_v2/... ./internal/modules/iam/domain -count=1
```

Expected: pass.

- [ ] **Step 14: Commit templates workstream**

Run:

```powershell
git add internal/modules/templates_v2 internal/modules/iam/domain/model.go migrations/0192_template_review_capability.sql
git commit -m "fix(templates): complete Plan 9R transaction and idempotency hardening"
```

---

### Task 5: Taxonomy Deactivate and Conflict Mapping

**Files:**
- Modify: `internal/modules/taxonomy/domain/port.go`
- Modify: `internal/modules/taxonomy/application/family_service.go`
- Modify: `internal/modules/taxonomy/infrastructure/family_repository.go`
- Modify: `internal/modules/taxonomy/delivery/http/routes_families.go`
- Modify: `internal/modules/taxonomy/delivery/http/routes_profiles.go`
- Modify: `internal/modules/taxonomy/delivery/http/routes_areas.go`

- [ ] **Step 1: Write failing tenant predicate test**

In repository tests, assert the active-profile query includes tenant scope:

```go
func TestFamilyRepository_HasActiveProfilesTx_UsesTenantPredicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tx, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1::uuid AND family_code = $2 AND archived_at IS NULL
)`)).
		WithArgs("tenant-1", "policy").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	repo := NewFamilyRepository(db)
	_, err = repo.HasActiveProfilesTx(context.Background(), tx, "tenant-1", "policy")
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Write failing 23505 mapping tests**

Add tests for profile and area handlers:

```go
func TestWriteProfileError_UniqueViolationReturns409(t *testing.T) {
	h := &Handler{}
	rr := httptest.NewRecorder()
	h.writeProfileError(rr, &pgconn.PgError{Code: "23505", ConstraintName: "document_profiles_tenant_id_code_key"})
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409", rr.Code)
	}
}
```

Repeat for `writeAreaError`.

- [ ] **Step 3: Run failing taxonomy tests**

Run:

```powershell
go test ./internal/modules/taxonomy/... -run "HasActiveProfilesTx|UniqueViolation" -count=1
```

Expected: fail.

- [ ] **Step 4: Add tx-aware family repository methods**

In `domain/port.go`:

```go
type FamilyRepository interface {
	GetByCode(ctx context.Context, code string) (*DocumentFamily, error)
	GetByCodeForUpdateTx(ctx context.Context, tx *sql.Tx, tenantID, code string) (*DocumentFamily, error)
	List(ctx context.Context, includeInactive bool) ([]DocumentFamily, error)
	Create(ctx context.Context, f *DocumentFamily) error
	Update(ctx context.Context, f *DocumentFamily) error
	UpdateTx(ctx context.Context, tx *sql.Tx, f *DocumentFamily) error
	HasActiveProfiles(ctx context.Context, familyCode string) (bool, error)
	HasActiveProfilesTx(ctx context.Context, tx *sql.Tx, tenantID, familyCode string) (bool, error)
}
```

- [ ] **Step 5: Implement row lock and tenant predicate**

In `family_repository.go`:

```go
func (r *FamilyRepository) GetByCodeForUpdateTx(ctx context.Context, tx *sql.Tx, tenantID, code string) (*domain.DocumentFamily, error) {
	const q = `
SELECT code, name, description, is_active
FROM metaldocs.document_families
WHERE code = $1
FOR UPDATE`
	return scanFamily(tx.QueryRowContext(ctx, q, code))
}

func (r *FamilyRepository) HasActiveProfilesTx(ctx context.Context, tx *sql.Tx, tenantID, familyCode string) (bool, error) {
	const q = `
SELECT EXISTS(
  SELECT 1 FROM metaldocs.document_profiles
  WHERE tenant_id = $1::uuid AND family_code = $2 AND archived_at IS NULL
)`
	var exists bool
	err := tx.QueryRowContext(ctx, q, tenantID, familyCode).Scan(&exists)
	return exists, err
}
```

If `document_families` is intentionally global and has no tenant column, keep the family row lookup global but tenant-scope only the profile check.

- [ ] **Step 6: Use one tx in FamilyService.Deactivate**

In `family_service.go`:

```go
tenantID, err := tenant.FromContext(ctx)
if err != nil {
	return err
}
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("taxonomy deactivate family: begin tx: %w", err)
}
defer func() { _ = tx.Rollback() }()
f, err := s.families.GetByCodeForUpdateTx(ctx, tx, tenantID, code)
if err != nil {
	return err
}
hasProfiles, err := s.families.HasActiveProfilesTx(ctx, tx, tenantID, code)
if err != nil {
	return err
}
if hasProfiles {
	return domain.ErrFamilyHasProfiles
}
if err := f.Deactivate(); err != nil {
	return err
}
if err := s.families.UpdateTx(ctx, tx, f); err != nil {
	return err
}
if err := tx.Commit(); err != nil {
	return err
}
```

Add `db *sql.DB` to `FamilyService` only if it does not already have a tx-capable dependency.

- [ ] **Step 7: Complete 23505 mappings**

In profile and area error mappers, add:

```go
case errors.As(err, &pgErr) && pgErr.Code == "23505":
	httpresponse.WriteError(w, http.StatusConflict, "RESOURCE_CONFLICT", "resource already exists")
```

Use `writeError` instead of `httpresponse.WriteError` in files that already use `writeError`.

- [ ] **Step 8: Verify taxonomy**

Run:

```powershell
go test ./internal/modules/taxonomy/... -count=1
```

Expected: pass.

- [ ] **Step 9: Commit taxonomy workstream**

Run:

```powershell
git add internal/modules/taxonomy
git commit -m "fix(taxonomy): make family deactivate transactional"
```

---

### Task 6: Wiki, Full Verification, and Final Audit Prep

**Files:**
- Modify: `wiki/backlog/roadmap.md`
- Modify: affected `wiki/backlog/*-refactor.md`
- Modify: affected `wiki/modules/*-tech-debt.md`
- Modify: `wiki/README.md` if summaries change.

- [ ] **Step 1: Update roadmap truthfully**

In `wiki/backlog/roadmap.md`, add a Plan 9R note:

```markdown
## Plan 9R · Transactional + idempotency recovery

- **Goal:** Complete the non-audit Plan 9 gaps verified by implementation audit.
- **Closes:** auth T-004/R-004; documents T-006/R-006, T-009/R-009; templates_v2 T-007/R-007, T-009/R-009, T-010/R-010; taxonomy T-007/R-007, R-011 conflict mapping; templates_v2 workflow-alignment remainder.
- **Explicit deferral:** audit T-004/R-004 hash-chain tamper evidence remains open for a dedicated follow-up plan.
- **Status:** done YYYY-MM-DD after verification.
```

Use the actual completion date and commit list after implementation.

- [ ] **Step 2: Update affected tech-debt rows**

For each closed row, change status text to include:

```markdown
**Resolution (2026-05-13, Plan 9R):** Closed by transaction/idempotency recovery. Verification: targeted Go tests plus `go test ./...`.
```

Do not mark audit T-004 closed.

- [ ] **Step 3: Run codegen drift check**

Run:

```powershell
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/templates_v2/api/...
go generate ./internal/modules/documents/api/...
git diff --exit-code -- internal/modules/templates_v2/api/api.gen.go internal/modules/documents/api/api.gen.go
```

Expected: no diff after committed generated files.

- [ ] **Step 4: Run backend verification**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
go build ./...
go test ./...
```

Expected: all pass.

- [ ] **Step 5: Regenerate frontend API types if OpenAPI drift reaches frontend**

Run:

```powershell
Push-Location frontend/apps/web
pnpm gen:api
npx tsc --noEmit
Pop-Location
```

Expected: generated type diff is committed if the header/response changes alter frontend types.

- [ ] **Step 6: Commit docs and verification fixes**

Run:

```powershell
git add wiki api frontend/apps/web/src/lib/api-types
git commit -m "docs: mark Plan 9R recovery complete"
```

- [ ] **Step 7: Run grouped implementation audit**

Use the `implementation-audit` skill against:

```text
Spec: docs/superpowers/specs/2026-05-13-plan-09r-recovery-design.md
Plan: docs/superpowers/plans/2026-05-13-plan-09r-recovery.md
Diff: all Plan 9R commits
Tests: OpenAPI lint, go build ./..., go test ./..., frontend API generation if run
```

Expected: PASS or PASS_WITH_ISSUES with no critical/major Plan 9R gaps before calling the recovery complete.

---

## Self-Review Checklist

- [x] Spec coverage: every Plan 9R design requirement maps to Tasks 1-6.
- [x] Placeholder scan: no banned placeholder instructions remain.
- [x] Type consistency: transaction-aware method names are consistent across service, repository, and tests.
- [x] Scope check: audit hash-chain, cursor pagination, Plan 10 cleanup, and broad route migration are excluded.
- [x] Parallel safety: auth, documents, templates, and taxonomy write sets are disjoint after Task 1.
