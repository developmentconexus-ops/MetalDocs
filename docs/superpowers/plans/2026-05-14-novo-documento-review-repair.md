# Novo Documento Review Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Repair PR 1 visibility and PR 2 blank-template review findings so the implementation is rollout-safe, contract-truthful, authz-aligned, and honest in the wizard UI.

**Architecture:** Keep the approved architecture: registry owns controlled-document visibility, templates owns the immutable system blank template, and the wizard submits only real backend-supported behavior. Repair the implementation by making migrations safe for existing data, making read models reflect persisted grants, adding explicit system-template immutability semantics, and exposing real area grant controls in Step 2.

**Tech Stack:** Go, Postgres migrations, OpenAPI-generated Go/TypeScript surfaces already present, React, TanStack Query, Vitest, MetalDocs wiki/skill workflow.

---

## Required Skills and Gates

Use these skills during execution:

- `metaldocs-backend-api` for registry/templates API or handler changes.
- `metaldocs-frontend` for `frontend/apps/web/` changes.
- `metaldocs-tanstack-query` if any query key or server-state wiring changes are touched.
- `runtime-contract-prereq` if any startup, route, migration, OpenAPI, generated type, or wrapper gate fails.
- `verification-before-completion` before claiming completion.

Before edits, run:

```powershell
git status --short --branch
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

Expected: status is understood; registry/templates contract gates pass; runtime target route is reachable. If any gate fails, stop feature repair and switch to `runtime-contract-prereq`.

## Parallel Agent Strategy

After preflight, these implementation tasks are parallel-safe because their write sets are disjoint:

- Task 1: Registry visibility rollout/read-model repair.
- Task 2: Templates blank-template authz/immutable-error repair.
- Task 3: Wizard area grant UI repair.

Do not run two workers against the same task. Each worker prompt must include:

```text
You are not alone in this codebase. Do not revert or overwrite edits outside your owned files. Adapt to existing changes.
```

Final verification and wiki sync are sequential.

## File Map

### Registry Repair

Modify:

- `migrations/0198_controlled_document_visibility.sql` - data-safe default for existing controlled documents.
- `internal/modules/registry/infrastructure/repository.go` - batch/single grant loading and conservative actor filtering support.
- `internal/modules/registry/application/service.go` - only inject actor filter when actor ID is non-empty.
- `internal/modules/registry/application/service_test.go` - service-level actor-filter regression test if fake repo supports filter capture.
- `internal/modules/registry/delivery/http/routes_contract_test.go` - response/grant mapping contract test.

Create:

- `internal/modules/registry/infrastructure/repository_test.go` - repository tests for grant round-trip if SQL-level tests are clearer than service fakes.

### Templates Repair

Modify:

- `internal/modules/templates/domain/template.go` - add `ErrSystemTemplateImmutable`.
- `internal/modules/templates/application/create.go` - return immutable error for system-owned create-next-version.
- `internal/modules/templates/application/lifecycle.go` - return immutable error for protected lifecycle mutations.
- `internal/modules/templates/application/approval_config.go` - return immutable error for protected approval-config mutation if currently guarded as archived.
- `internal/modules/templates/delivery/http/errors.go` - map immutable error to `409 SYSTEM_TEMPLATE_IMMUTABLE`.
- `internal/modules/templates/delivery/http/errors_test.go` - error mapping test.
- `internal/modules/templates/delivery/http/routes_query.go` - add `template.view` authz to blank endpoint.
- `internal/modules/templates/delivery/http/routes_create_test.go` - assert immutable next-version conflict code.
- `internal/modules/templates/delivery/http/routes_lifecycle_test.go` - assert immutable lifecycle conflict code.

### Frontend Wizard Repair

Create:

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/AreaVisibilitySubcontrols.tsx` - real selected-area controls.

Modify:

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/index.tsx` - render area grant controls for `visibility === 'area'`.
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/StepAreaCodeVisibility.module.css` - make area controls active and keep people/external deferred.
- `frontend/apps/web/src/features/documents/state/wizard.reducer.ts` - preserve locked document area when setting visibility areas.
- `frontend/apps/web/src/features/documents/state/__tests__/wizard.reducer.test.ts` - area grant reducer tests.
- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx` - pass selected area codes and toggle dispatch to Step 2.
- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx` - payload includes added area grant.

### Wiki/Docs Sync

Modify:

- `wiki/modules/registry.md`
- `wiki/modules/templates.md`
- `wiki/modules/novo-documento-wizard.md`
- `wiki/backlog/novo-documento.md`
- `frontend/apps/web/design-source/novo-documento/NOTES.md`

---

## Task 1: Registry Visibility Rollout and Read-Model Repair `[parallel-safe after preflight]`

**Owned files:**

- `migrations/0198_controlled_document_visibility.sql`
- `internal/modules/registry/infrastructure/repository.go`
- `internal/modules/registry/infrastructure/repository_test.go`
- `internal/modules/registry/application/service.go`
- `internal/modules/registry/application/service_test.go`
- `internal/modules/registry/delivery/http/routes_contract_test.go`

- [ ] **Step 1: Fix migration default for rollout safety**

In `migrations/0198_controlled_document_visibility.sql`, change:

```sql
ALTER TABLE controlled_documents
  ADD COLUMN IF NOT EXISTS visibility_scope TEXT NOT NULL DEFAULT 'restricted',
```

to:

```sql
ALTER TABLE controlled_documents
  ADD COLUMN IF NOT EXISTS visibility_scope TEXT NOT NULL DEFAULT 'company',
```

Expected behavior: existing controlled documents without grant rows remain visible after migration.

- [ ] **Step 2: Add a failing repository test for persisted grant round-trip**

Create `internal/modules/registry/infrastructure/repository_test.go` if it does not exist.

Add this test. It is self-contained and uses `sqlmock`; no database fixture is required:

```go
package infrastructure

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresControlledDocumentRepository_GetByIDLoadsVisibilityGrants(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresControlledDocumentRepository(db)
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND id = $2`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "profile_code", "process_area_code", "department_code",
			"code", "sequence_num", "title", "owner_user_id", "override_template_version_id",
			"visibility_scope", "status", "created_at", "updated_at",
		}).AddRow("cd-1", "tenant-1", "POP", "QA", nil, "POP-QA-001", 1, "Procedure", "owner-1", "", "restricted", "active", now, now))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY area_code`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{"area_code"}).AddRow("QA").AddRow("RH"))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY user_id`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-2"))

	doc, err := repo.GetByID(context.Background(), "tenant-1", "cd-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if doc.Visibility.Scope != "restricted" {
		t.Fatalf("scope = %q, want restricted", doc.Visibility.Scope)
	}
	if got := doc.Visibility.AreaCodes; len(got) != 2 || got[0] != "QA" || got[1] != "RH" {
		t.Fatalf("area grants = %#v, want [QA RH]", got)
	}
	if got := doc.Visibility.UserIDs; len(got) != 1 || got[0] != "user-2" {
		t.Fatalf("user grants = %#v, want [user-2]", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
```

- [ ] **Step 3: Run the new repository test to verify it fails**

Run:

```powershell
go test ./internal/modules/registry/infrastructure -run GetByIDLoadsVisibilityGrants -count=1
```

Expected: fail because `GetByID` currently calls `scanControlledDocument` and does not query grant tables.

- [ ] **Step 4: Add grant loading helpers**

In `internal/modules/registry/infrastructure/repository.go`, add helpers near `scanControlledDocument`:

```go
func (r *PostgresControlledDocumentRepository) loadVisibilityGrants(ctx context.Context, tenantID, controlledDocumentID string) (registrydomain.Visibility, error) {
	areas, err := r.loadAreaGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return registrydomain.Visibility{}, err
	}
	users, err := r.loadUserGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return registrydomain.Visibility{}, err
	}
	return registrydomain.NewVisibility(string(registrydomain.VisibilityScopeRestricted), areas, users, "")
}

func (r *PostgresControlledDocumentRepository) loadAreaGrants(ctx context.Context, tenantID, controlledDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY area_code`, tenantID, controlledDocumentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var areaCode string
		if err := rows.Scan(&areaCode); err != nil {
			return nil, err
		}
		out = append(out, areaCode)
	}
	return out, rows.Err()
}

func (r *PostgresControlledDocumentRepository) loadUserGrants(ctx context.Context, tenantID, controlledDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY user_id`, tenantID, controlledDocumentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		out = append(out, userID)
	}
	return out, rows.Err()
}
```

- [ ] **Step 5: Use stored grants in detail reads**

In `GetByID`, replace:

```go
return scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, id))
```

with:

```go
doc, err := scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, id))
if err != nil {
	return nil, err
}
if doc.Visibility.Scope == registrydomain.VisibilityScopeRestricted {
	vis, err := r.loadVisibilityGrants(ctx, tenantID, doc.ID)
	if err != nil {
		return nil, err
	}
	doc.Visibility = vis
}
return doc, nil
```

Apply the same pattern to `GetByCode`.

- [ ] **Step 6: Add a batch grant loader for list reads**

In `repository.go`, add:

```go
func (r *PostgresControlledDocumentRepository) hydrateVisibilityGrants(ctx context.Context, tenantID string, docs []registrydomain.ControlledDocument) error {
	ids := make([]string, 0, len(docs))
	indexByID := make(map[string]int, len(docs))
	for i := range docs {
		if docs[i].Visibility.Scope != registrydomain.VisibilityScopeRestricted {
			continue
		}
		ids = append(ids, docs[i].ID)
		indexByID[docs[i].ID] = i
		docs[i].Visibility.AreaCodes = []string{}
		docs[i].Visibility.UserIDs = []string{}
	}
	if len(ids) == 0 {
		return nil
	}

	areaRows, err := r.db.QueryContext(ctx, `
SELECT controlled_document_id::text, area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = ANY($2)
ORDER BY controlled_document_id, area_code`, tenantID, pgtype.FlatArray[string](ids))
	if err != nil {
		return err
	}
	defer areaRows.Close()
	for areaRows.Next() {
		var docID, areaCode string
		if err := areaRows.Scan(&docID, &areaCode); err != nil {
			return err
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.AreaCodes = append(docs[idx].Visibility.AreaCodes, areaCode)
		}
	}
	if err := areaRows.Err(); err != nil {
		return err
	}

	userRows, err := r.db.QueryContext(ctx, `
SELECT controlled_document_id::text, user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = ANY($2)
ORDER BY controlled_document_id, user_id`, tenantID, pgtype.FlatArray[string](ids))
	if err != nil {
		return err
	}
	defer userRows.Close()
	for userRows.Next() {
		var docID, userID string
		if err := userRows.Scan(&docID, &userID); err != nil {
			return err
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.UserIDs = append(docs[idx].Visibility.UserIDs, userID)
		}
	}
	return userRows.Err()
}
```

- [ ] **Step 7: Call the batch grant loader in `List`**

In `List`, after row iteration and before return, replace:

```go
if err := rows.Err(); err != nil {
	return nil, err
}
return out, nil
```

with:

```go
if err := rows.Err(); err != nil {
	return nil, err
}
if err := r.hydrateVisibilityGrants(ctx, tenantID, out); err != nil {
	return nil, err
}
return out, nil
```

- [ ] **Step 8: Keep scan focused on row columns**

In `scanControlledDocument`, keep:

```go
vis, err := registrydomain.NewVisibility(visibilityScope, nil, nil, doc.ProcessAreaCode)
```

but change it to avoid inventing area grants for persisted restricted rows:

```go
vis, err := registrydomain.NewVisibility(visibilityScope, nil, nil, "")
```

Then ensure `NewVisibility` can represent restricted-with-empty-grants for an intermediate scan by returning restricted with empty slices when `defaultAreaCode == ""`.

In `internal/modules/registry/domain/visibility.go`, replace the restricted return section with:

```go
normAreas := uniqueNonEmpty(areaCodes)
if len(normAreas) == 0 {
	fallback := strings.TrimSpace(defaultAreaCode)
	if fallback != "" {
		normAreas = []string{fallback}
	}
}

return Visibility{
	Scope:     VisibilityScopeRestricted,
	AreaCodes: normAreas,
	UserIDs:   uniqueNonEmpty(userIDs),
}, nil
```

This code already exists; do not add a validation that rejects empty restricted grants in row scanning.

- [ ] **Step 9: Fix service actor filter injection**

In `internal/modules/registry/application/service.go`, change:

```go
actorUserID := authn.UserIDFromContext(ctx)
filter.ActorUserID = &actorUserID
```

to:

```go
if actorUserID := authn.UserIDFromContext(ctx); strings.TrimSpace(actorUserID) != "" {
	filter.ActorUserID = &actorUserID
}
```

Confirm `strings` is already imported in the file. It is already used by create validation.

- [ ] **Step 10: Add service test for empty actor behavior**

In `internal/modules/registry/application/service_test.go`, extend `fakeControlledDocumentRepository`:

```go
lastListFilter registrydomain.CDFilter
```

Update fake `List`:

```go
func (f *fakeControlledDocumentRepository) List(_ context.Context, _ string, filter registrydomain.CDFilter) ([]registrydomain.ControlledDocument, error) {
	f.lastListFilter = filter
	return nil, nil
}
```

Add test:

```go
func TestList_DoesNotInjectEmptyActorFilter(t *testing.T) {
	repo := newFakeControlledDocumentRepository()
	svc := NewRegistryService(nil, repo, &fakeSequenceAllocator{}, &fakeTemplateVersionChecker{}, &fakeProfileReader{}, &fakeAreaReader{}, &fakeGovernanceLogger{}, nil)

	_, err := svc.List(context.Background(), "tenant-a", registrydomain.CDFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastListFilter.ActorUserID != nil {
		t.Fatalf("ActorUserID = %v, want nil", *repo.lastListFilter.ActorUserID)
	}
}
```

- [ ] **Step 11: Run registry tests**

Run:

```powershell
go test ./internal/modules/registry/domain ./internal/modules/registry/application ./internal/modules/registry/infrastructure ./internal/modules/registry/delivery/http -count=1
```

Expected: all listed packages pass.

- [ ] **Step 12: Commit registry repair**

Run:

```powershell
git add migrations/0198_controlled_document_visibility.sql internal/modules/registry
git commit -m "fix(registry): preserve and return controlled document visibility"
```

---

## Task 2: Templates Authz and Immutable Error Repair `[parallel-safe after preflight]`

**Owned files:**

- `internal/modules/templates/domain/template.go`
- `internal/modules/templates/application/create.go`
- `internal/modules/templates/application/lifecycle.go`
- `internal/modules/templates/application/approval_config.go`
- `internal/modules/templates/delivery/http/errors.go`
- `internal/modules/templates/delivery/http/errors_test.go`
- `internal/modules/templates/delivery/http/routes_query.go`
- `internal/modules/templates/delivery/http/routes_create_test.go`
- `internal/modules/templates/delivery/http/routes_lifecycle_test.go`

- [ ] **Step 1: Add immutable domain error**

In `internal/modules/templates/domain/template.go`, add:

```go
ErrSystemTemplateImmutable = errors.New("templates_v2: system_template_immutable")
```

inside the existing `var` block.

- [ ] **Step 2: Update error mapping test first**

In `internal/modules/templates/delivery/http/errors_test.go`, add a case to the existing table:

```go
{name: "system template immutable", err: domain.ErrSystemTemplateImmutable, wantStatus: http.StatusConflict, wantCode: "SYSTEM_TEMPLATE_IMMUTABLE"},
```

- [ ] **Step 3: Run error mapping test to verify failure**

Run:

```powershell
go test ./internal/modules/templates/delivery/http -run TestMapErr -count=1
```

Expected: fail because `MapErr` does not map `ErrSystemTemplateImmutable`.

- [ ] **Step 4: Map immutable error to explicit conflict code**

In `internal/modules/templates/delivery/http/errors.go`, add before `ErrArchived`:

```go
case errors.Is(err, domain.ErrSystemTemplateImmutable):
	return http.StatusConflict, "SYSTEM_TEMPLATE_IMMUTABLE"
```

- [ ] **Step 5: Replace archived error in system-owned guards**

In `internal/modules/templates/application/create.go`, replace system-owned guard returns:

```go
return nil, domain.ErrArchived
```

with:

```go
return nil, domain.ErrSystemTemplateImmutable
```

Do the same in `internal/modules/templates/application/lifecycle.go`.

In `internal/modules/templates/application/approval_config.go`, replace the system-owned guard return with:

```go
return nil, domain.ErrSystemTemplateImmutable
```

Keep true archived-template checks returning `domain.ErrArchived`.

- [ ] **Step 6: Add blank endpoint authz**

In `internal/modules/templates/delivery/http/routes_query.go`, change:

```go
if _, err := tenantIDFromReq(r); err != nil {
	writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
	return
}
```

to:

```go
tenantID, err := tenantIDFromReq(r)
if err != nil {
	writeErr(w, http.StatusInternalServerError, "internal_error", "internal server error")
	return
}
if err := h.authz(r, tenantID, "*", "template.view"); err != nil {
	writeMappedErr(w, err)
	return
}
```

Leave the sentinel read as:

```go
tpl, err := h.svc.GetTemplate(r.Context(), systemBlankTemplateTenantID, systemBlankTemplateID)
```

because the system template is intentionally cross-tenant and immutable.

- [ ] **Step 7: Update HTTP immutable tests**

In `routes_create_test.go` and `routes_lifecycle_test.go`, update assertions that currently expect `archived` for system-owned templates to expect:

```go
if rec.Code != http.StatusConflict {
	t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
}
if !strings.Contains(rec.Body.String(), "SYSTEM_TEMPLATE_IMMUTABLE") {
	t.Fatalf("body = %s, want SYSTEM_TEMPLATE_IMMUTABLE", rec.Body.String())
}
```

Ensure `strings` is imported if the file does not already import it.

- [ ] **Step 8: Run templates tests**

Run:

```powershell
go test ./internal/modules/templates/... -count=1
```

Expected: all templates packages pass.

- [ ] **Step 9: Commit templates repair**

Run:

```powershell
git add internal/modules/templates
git commit -m "fix(templates): enforce system blank template boundaries"
```

---

## Task 3: Wizard Area Grant UI Repair `[parallel-safe after preflight]`

**Owned files:**

- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/AreaVisibilitySubcontrols.tsx`
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/index.tsx`
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/StepAreaCodeVisibility.module.css`
- `frontend/apps/web/src/features/documents/state/wizard.reducer.ts`
- `frontend/apps/web/src/features/documents/state/__tests__/wizard.reducer.test.ts`
- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx`
- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx`

- [ ] **Step 1: Add reducer tests for locked own area and additional area grants**

In `frontend/apps/web/src/features/documents/state/__tests__/wizard.reducer.test.ts`, add under `area visibility defaults`:

```ts
it('keeps selected document area when setting additional visibility areas', () => {
  const seeded = {
    ...INITIAL_STATE,
    areaCode: 'QA',
    visibility: 'area' as const,
    visibilityAreaCodes: ['QA'],
  };
  const next = wizardReducer(seeded, {
    type: 'setVisibilityAreas',
    codes: ['RH'],
  });
  expect(next.visibilityAreaCodes).toEqual(['QA', 'RH']);
});

it('clears restricted grants when switching to company visibility', () => {
  const seeded = {
    ...INITIAL_STATE,
    areaCode: 'QA',
    visibility: 'area' as const,
    visibilityAreaCodes: ['QA', 'RH'],
    invitees: [{ id: 'user-1', label: 'User 1' }],
  };
  const next = wizardReducer(seeded, { type: 'setVisibility', key: 'company' });
  expect(next.visibilityAreaCodes).toEqual([]);
  expect(next.invitees).toEqual([]);
});
```

- [ ] **Step 2: Run reducer test to verify failure**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd test src/features/documents/state/__tests__/wizard.reducer.test.ts
```

Expected: fail because `setVisibilityAreas` currently replaces the locked document area and company does not clear invitees/area grants.

- [ ] **Step 3: Update reducer behavior**

In `wizard.reducer.ts`, update `setVisibility`:

```ts
case 'setVisibility':
  if (action.key === 'company') {
    return {
      ...state,
      visibility: action.key,
      visibilityAreaCodes: [],
      invitees: [],
      external: { passwordRequired: false, watermark: false, expiresInDays: null },
    };
  }
  return {
    ...state,
    visibility: action.key,
    visibilityAreaCodes:
      action.key === 'area'
        ? state.areaCode
          ? [state.areaCode]
          : []
        : state.visibilityAreaCodes,
  };
```

Update `setVisibilityAreas`:

```ts
case 'setVisibilityAreas': {
  const unique = Array.from(new Set(action.codes.filter(Boolean)));
  const withOwnArea =
    state.areaCode && !unique.includes(state.areaCode) ? [state.areaCode, ...unique] : unique;
  return { ...state, visibilityAreaCodes: withOwnArea };
}
```

- [ ] **Step 4: Create active area subcontrols component**

Create `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility/AreaVisibilitySubcontrols.tsx`:

```tsx
import type { ProcessArea } from '../../../../../taxonomy/types';
import styles from './StepAreaCodeVisibility.module.css';

export type AreaVisibilitySubcontrolsProps = {
  areas: ProcessArea[];
  documentAreaCode: string;
  selectedCodes: string[];
  onSetAreaCodes: (codes: string[]) => void;
};

export function AreaVisibilitySubcontrols({
  areas,
  documentAreaCode,
  selectedCodes,
  onSetAreaCodes,
}: AreaVisibilitySubcontrolsProps): JSX.Element {
  function toggle(code: string) {
    if (code === documentAreaCode) return;
    const next = selectedCodes.includes(code)
      ? selectedCodes.filter((current) => current !== code)
      : [...selectedCodes, code];
    onSetAreaCodes(next);
  }

  return (
    <div className={`card ${styles.areaSubcontrolsCard}`} role="group" aria-label="Areas visiveis">
      <div className="kicker">Areas visiveis</div>
      <div className={styles.subcontrolsRow}>
        {areas.map((area) => {
          const checked = selectedCodes.includes(area.code);
          const locked = area.code === documentAreaCode;
          return (
            <label key={area.code} className={styles.areaChoice}>
              <input
                type="checkbox"
                checked={checked}
                disabled={locked}
                onChange={() => toggle(area.code)}
              />
              <span>
                {area.code} - {area.name}
                {locked ? ' (area do documento)' : ''}
              </span>
            </label>
          );
        })}
      </div>
    </div>
  );
}
```

The display text uses ASCII `Areas` to match current file encoding in this component family.

- [ ] **Step 5: Add CSS for active area controls**

In `StepAreaCodeVisibility.module.css`, add:

```css
.areaSubcontrolsCard {
  margin-top: calc(-1 * var(--sp-3));
  margin-bottom: var(--sp-6);
}

.areaChoice {
  display: inline-flex;
  align-items: center;
  gap: var(--sp-2);
  min-height: 32px;
  padding: var(--sp-2) var(--sp-3);
  border: 1px solid var(--border);
  border-radius: var(--r-2);
  background: var(--surface);
  font-size: 13px;
}

.areaChoice input {
  margin: 0;
}
```

Leave `.subcontrolsCard` disabled/opaque for people and external only.

- [ ] **Step 6: Wire Step 2 props**

In `StepAreaCodeVisibility/index.tsx`, import:

```ts
import { AreaVisibilitySubcontrols } from './AreaVisibilitySubcontrols';
```

Add props:

```ts
visibilityAreaCodes: string[];
onSetVisibilityAreas: (codes: string[]) => void;
```

Destructure them from `props`.

After the visibility grid and before people controls, render:

```tsx
{visibility === 'area' ? (
  <AreaVisibilitySubcontrols
    areas={areas}
    documentAreaCode={areaCode}
    selectedCodes={visibilityAreaCodes}
    onSetAreaCodes={onSetVisibilityAreas}
  />
) : null}
```

- [ ] **Step 7: Pass Step 2 area grant state from page**

In `NewDocumentWizardPage.tsx`, add props to `StepAreaCodeVisibility`:

```tsx
visibilityAreaCodes={state.visibilityAreaCodes}
onSetVisibilityAreas={(codes) => dispatch({ type: 'setVisibilityAreas', codes })}
```

- [ ] **Step 8: Add page test for additional area grant payload**

In `NewDocumentWizardPage.test.tsx`, add or update the submit payload test to simulate selected area grants through reducer/page behavior. If the Step 2 component is mocked, test `buildVisibilityPayload` directly:

```ts
it('builds restricted visibility payload with additional selected areas', () => {
  const payload = buildVisibilityPayload({
    ...INITIAL_STATE,
    visibility: 'area',
    areaCode: 'QA',
    visibilityAreaCodes: ['QA', 'RH'],
  });

  expect(payload).toEqual({
    scope: 'restricted',
    areaCodes: ['QA', 'RH'],
    userIds: [],
  });
});
```

If `INITIAL_STATE` is not imported in this test file, import it from:

```ts
import { INITIAL_STATE } from '../state/wizard.reducer';
```

- [ ] **Step 9: Run frontend targeted tests and typecheck**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd test src/features/documents/state/__tests__/wizard.reducer.test.ts src/features/documents/pages/NewDocumentWizardPage.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: targeted tests and typecheck pass.

- [ ] **Step 10: Commit frontend repair**

Run:

```powershell
git add frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility frontend/apps/web/src/features/documents/state frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx
git commit -m "fix(web): align visibility area grants with wizard contract"
```

---

## Task 4: Wiki and Design-Source Sync `[seq after Tasks 1-3]`

**Owned files:**

- `wiki/modules/registry.md`
- `wiki/modules/templates.md`
- `wiki/modules/novo-documento-wizard.md`
- `wiki/backlog/novo-documento.md`
- `frontend/apps/web/design-source/novo-documento/NOTES.md`

- [ ] **Step 1: Update registry wiki truth**

In `wiki/modules/registry.md`, update the introduction/solution strategy or add a short subsection with:

```markdown
### Controlled-document visibility

Controlled-document visibility is stored on registry-owned `controlled_documents` rows plus grant tables. Existing controlled documents remain `company` visible after migration. New restricted documents persist area/user grants in the same atomic create transaction, and list/detail reads enforce visibility server-side before pagination while returning the persisted grant summary.
```

- [ ] **Step 2: Update templates wiki truth**

In `wiki/modules/templates.md`, add:

```markdown
### System blank template

The novo-documento blank option is backed by a system-owned immutable template/version. Normal template list queries hide system-owned templates. Normal mutation routes reject system-owned templates with `SYSTEM_TEMPLATE_IMMUTABLE`, and the dedicated blank-template read endpoint still requires `template.view`.
```

- [ ] **Step 3: Update novo-documento wizard wiki/backlog**

In `wiki/modules/novo-documento-wizard.md` and `wiki/backlog/novo-documento.md`, record:

```markdown
- Visibility persistence and backend read enforcement are implemented for company and selected-area visibility.
- The Step 2 area visibility UI submits real selected area grants.
- Specific-people visibility remains deferred pending an author-safe IAM user search endpoint.
- External sharing remains deferred.
- Blank document creation is backed by the real immutable system blank template version.
```

- [ ] **Step 4: Update design-source notes**

In `frontend/apps/web/design-source/novo-documento/NOTES.md`, add the same implementation truth:

```markdown
Review repair 2026-05-14:
- Existing controlled documents remain company-visible after migration.
- Restricted visibility returns persisted area/user grants from backend reads.
- Step 2 selected-area controls are real; own document area is locked selected.
- Specific people and external sharing remain deferred.
- Blank template selection uses `GET /api/v1/templates/system/blank`.
```

- [ ] **Step 5: Commit docs sync**

Run:

```powershell
git add wiki/modules/registry.md wiki/modules/templates.md wiki/modules/novo-documento-wizard.md wiki/backlog/novo-documento.md frontend/apps/web/design-source/novo-documento/NOTES.md
git commit -m "docs: sync novo documento review repair truth"
```

---

## Task 5: Final Verification and Review Handoff `[seq final]`

**Owned files:** none unless verification exposes a bug in task-owned files.

- [ ] **Step 1: Run backend focused verification**

Run:

```powershell
go test ./internal/modules/registry/... ./internal/modules/templates/... -count=1
```

Expected: all registry and templates packages pass.

- [ ] **Step 2: Run frontend focused verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd test src/features/documents/state/__tests__/wizard.reducer.test.ts src/features/documents/pages/NewDocumentWizardPage.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: targeted wizard tests and typecheck pass.

- [ ] **Step 3: Run contract/runtime gates**

Run from repo root:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

Expected: all pass.

- [ ] **Step 4: Run full backend verification**

Run:

```powershell
go test ./... -count=1
```

Expected: pass.

- [ ] **Step 5: Run frontend full tests and record known unrelated failures**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd test
```

Expected: either pass or fail only in pre-existing unrelated suites. Record exact failing suites. The prior known unrelated failures were in approval route admin, API client tests, PDF polling/editor tests, and DocumentsHubView edit-button tests.

- [ ] **Step 6: Run final status/scope check**

Run:

```powershell
git status --short
git diff --stat HEAD~5..HEAD
```

Expected: changed scope is limited to registry visibility repair, templates blank-template repair, wizard area controls, docs/wiki sync, and generated/contract files already present from the parent feature.

- [ ] **Step 7: Request final code review**

Use `superpowers:requesting-code-review` with this description:

```text
Repair pass for novo-documento PR1/PR2 review findings: data-safe visibility migration, persisted grant read model, explicit system-template immutability/authz, wizard area grant controls, and wiki truth sync.
```

Ask reviewer to focus on:

```text
Migration safety, visibility grant round-trip correctness, read filtering before pagination, immutable system-template error semantics, blank endpoint authz, and wizard area grant payload honesty.
```

Expected: reviewer reports no Critical or Important findings before merge handoff.
