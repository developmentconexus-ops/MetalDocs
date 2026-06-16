# F1.3 — createTemplate declared-fields-only — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development (recommended)
> or superpowers:executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove undeclared top-level `id` / `version_id` from `createTemplate` 201 response and
wire typed `TemplateDTO` / `VersionDTO` via a new `toAPITemplateDTO` mapper, making the response a
strict subset of the OpenAPI `CreateTemplateResponse` schema (A3/H-D fix).

**Architecture:** Add `toAPITemplateDTO(t *domain.Template) (templatesapi.TemplateDTO, error)` to
`routes_mapping.go` (mirrors `toAPIVersionDTO`). In `CreateTemplate` (`routes_generated.go`), call
both mappers pre-flight, assemble `templatesapi.CreateTemplateResponse`, and `writeJSON` with the
typed struct — the map literal with `"id"` / `"version_id"` is replaced entirely.

**Tech Stack:** Go `net/http`, oapi-codegen generated types (`templatesapi`), `go test`, `github.com/google/uuid`.

---

### Task 1: Write the failing test (RED)

**Files:**
- Modify: `internal/modules/templates/delivery/http/routes_typed_response_test.go`

> The test uses a valid UUID tenant so `toAPITemplateDTO` can parse `TenantId`. `"tenant-a"` used
> by `withHeaders` is not a valid UUID, so this test sets up its own context.

- [ ] **Step 1: Add imports to `routes_typed_response_test.go`** — the file currently does not import
  `tenant` or `iamdomain`. Add them:

In the `import` block of `routes_typed_response_test.go`, change:

```go
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/templates/domain"
)
```

to:

```go
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)
```

- [ ] **Step 2: Add `TestCreateTemplate_TypedResponseShape` at the end of `routes_typed_response_test.go`:**

```go
// TestCreateTemplate_TypedResponseShape — V1/V2 (F1.3/A3).
// POST /api/v1/templates (createTemplate) must return 201 + {data:{template:TemplateDTO,version:VersionDTO}}
// with NO undeclared top-level "id" or "version_id" (H-D fix).
// Uses a UUID-format tenant so toAPITemplateDTO can parse TenantId (the mapper calls uuid.Parse).
func TestCreateTemplate_TypedResponseShape(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	const tenantUUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	body := createBody("shape-f13-key")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Idempotency-Key", "33333333-3333-4333-8333-333333333333")
	ctx := tenant.WithTenantID(req.Context(), tenantUUID)
	ctx = iamdomain.WithAuthContext(ctx, "user-shape", []iamdomain.Role{})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw: %v (body=%s)", err, rr.Body.String())
	}

	// V1: undeclared top-level fields must be absent.
	if _, ok := raw["id"]; ok {
		t.Error("top-level 'id' must not be present (A3/H-D: field not declared in CreateTemplateResponse)")
	}
	if _, ok := raw["version_id"]; ok {
		t.Error("top-level 'version_id' must not be present (A3/H-D: field not declared in CreateTemplateResponse)")
	}

	// V2: declared envelope structure must be present.
	dataRaw, ok := raw["data"]
	if !ok {
		t.Fatal("missing top-level 'data' key (declared by CreateTemplateResponse)")
	}

	var dataObj struct {
		Template json.RawMessage `json:"template"`
		Version  json.RawMessage `json:"version"`
	}
	if err := json.Unmarshal(dataRaw, &dataObj); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if dataObj.Template == nil {
		t.Fatal("data.template must be present")
	}
	if dataObj.Version == nil {
		t.Fatal("data.version must be present")
	}

	// Smoke fields: data.template must carry id and tenant_id (TemplateDTO required).
	var tpl struct {
		Id       string `json:"id"`
		TenantId string `json:"tenant_id"`
		Key      string `json:"key"`
	}
	if err := json.Unmarshal(dataObj.Template, &tpl); err != nil {
		t.Fatalf("decode data.template: %v", err)
	}
	if tpl.Id == "" {
		t.Error("data.template.id must be non-empty (TemplateDTO required field)")
	}
	if tpl.TenantId != tenantUUID {
		t.Errorf("data.template.tenant_id = %q; want %q", tpl.TenantId, tenantUUID)
	}
	if tpl.Key != "shape-f13-key" {
		t.Errorf("data.template.key = %q; want %q", tpl.Key, "shape-f13-key")
	}

	// Smoke fields: data.version must carry version_number (VersionDTO required).
	var ver struct {
		VersionNumber int `json:"version_number"`
	}
	if err := json.Unmarshal(dataObj.Version, &ver); err != nil {
		t.Fatalf("decode data.version: %v", err)
	}
	if ver.VersionNumber != 1 {
		t.Errorf("data.version.version_number = %d; want 1", ver.VersionNumber)
	}
}
```

- [ ] **Step 3: Run to confirm RED:**

```
go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate_TypedResponseShape -v
```

Expected: **FAIL** — `top-level 'id' must not be present` (the old `map[string]any` literal includes `"id"` and `"version_id"`).

---

### Task 2: Add `toAPITemplateDTO` mapper

**Files:**
- Modify: `internal/modules/templates/delivery/http/routes_mapping.go`

- [ ] **Step 1: Add `toAPITemplateDTO` at the end of `routes_mapping.go`** (after `placeholdersToSlice`):

```go
// toAPITemplateDTO maps a domain.Template to the OpenAPI-generated TemplateDTO type.
// F1.3 / ADR 0035 — flat typed wire shape. Mirrors the toAPIVersionDTO pattern.
// SystemOwned is not present in TemplateDTO (not a public API field).
func toAPITemplateDTO(t *domain.Template) (templatesapi.TemplateDTO, error) {
	if t == nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("toAPITemplateDTO: nil template")
	}
	id, err := uuid.Parse(t.ID)
	if err != nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("template id %q: %w", t.ID, err)
	}
	tenantID, err := uuid.Parse(t.TenantID)
	if err != nil {
		return templatesapi.TemplateDTO{}, fmt.Errorf("template tenant_id %q: %w", t.TenantID, err)
	}

	latestRevNum := int32(t.LatestRevisionNumber)
	dto := templatesapi.TemplateDTO{
		Id:                   id,
		TenantId:             tenantID,
		Key:                  t.Key,
		Name:                 t.Name,
		LatestVersion:        t.LatestVersion,
		LatestRevisionNumber: latestRevNum,
		CreatedBy:            t.CreatedBy,
		CreatedAt:            t.CreatedAt.UTC(),
	}

	if t.Description != "" {
		dto.Description = &t.Description
	}
	if t.DocTypeCode != "" {
		dto.DocTypeCode = &t.DocTypeCode
	}
	if t.PublishedVersionID != nil {
		pvID, err := uuid.Parse(*t.PublishedVersionID)
		if err != nil {
			return templatesapi.TemplateDTO{}, fmt.Errorf("template published_version_id %q: %w", *t.PublishedVersionID, err)
		}
		dto.PublishedVersionId = &pvID
	}
	dto.PublishedVersionNumber = t.PublishedVersionNumber
	if t.CurrentRevisionNumber != nil {
		n := int32(*t.CurrentRevisionNumber)
		dto.CurrentRevisionNumber = &n
	}
	if t.ArchivedAt != nil {
		u := t.ArchivedAt.UTC()
		dto.ArchivedAt = &u
	}
	return dto, nil
}
```

- [ ] **Step 2: Build to verify types compile:**

```
go build ./internal/modules/templates/delivery/http/...
```

Expected: exit 0. (The `templatesapi.TemplateDTO` struct fields must match what we assign.)

---

### Task 3: Hoist mappers + write typed response in `CreateTemplate`

**Files:**
- Modify: `internal/modules/templates/delivery/http/routes_generated.go`

- [ ] **Step 1: Replace the `writeJSON` block in `CreateTemplate`.**

Find this block (lines 64–71):

```go
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         res.Template.ID,
		"version_id": res.Version.ID,
		"data": map[string]any{
			"template": toTemplateResponse(res.Template),
			"version":  toVersionResponse(res.Version),
		},
	})
```

Replace with:

```go
	tplDTO, err := toAPITemplateDTO(res.Template)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	vDTO, err := toAPIVersionDTO(res.Version)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, codeTplInternalError, "internal server error")
		return
	}
	var resp templatesapi.CreateTemplateResponse
	resp.Data.Template = tplDTO
	resp.Data.Version = vDTO
	writeJSON(w, http.StatusCreated, resp)
```

- [ ] **Step 2: Add `templatesapi` import if not present.** Check the import block at the top of
  `routes_generated.go` — `templatesapi "metaldocs/internal/modules/templates/api"` is already
  there (line 9). No change needed.

- [ ] **Step 3: Build:**

```
go build ./internal/modules/templates/delivery/http/...
```

Expected: exit 0.

---

### Task 4: GREEN + update existing test

**Files:**
- Modify: `internal/modules/templates/delivery/http/routes_create_test.go`

- [ ] **Step 1: Run the new test to confirm GREEN:**

```
go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate_TypedResponseShape -v
```

Expected: **PASS**.

- [ ] **Step 2: Run existing `TestCreateTemplate_Happy` to confirm it still passes:**

```
go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate_Happy -v
```

Expected: **PASS**. (The test decodes `data.template` as `map[string]any` — still works since
`TemplateDTO` marshals to the same JSON keys. `data.template["id"]` is still present.)

- [ ] **Step 3: Add top-level absence assertions to `TestCreateTemplate_Happy`.** In
  `routes_create_test.go`, find `TestCreateTemplate_Happy` and the `json.Unmarshal` into `out`.
  After the `Unmarshal` but before the `gotTenant` check (around line 418), add:

```go
	// F1.3: top-level undeclared fields must be absent.
	var rawTop map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &rawTop); err != nil {
		t.Fatalf("decode raw top-level: %v", err)
	}
	if _, ok := rawTop["id"]; ok {
		t.Error("top-level 'id' must not be present after F1.3 (A3/H-D)")
	}
	if _, ok := rawTop["version_id"]; ok {
		t.Error("top-level 'version_id' must not be present after F1.3 (A3/H-D)")
	}
```

- [ ] **Step 4: Run all createTemplate tests:**

```
go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate -v
```

Expected: all PASS — `TestCreateTemplate_Happy`, `TestCreateTemplate_TypedResponseShape`,
`TestCreateTemplate_RejectUnknownField`, `TestCreateTemplate_KeyConflict`.

---

### Task 5: H-D structural proof + full suite

**Files:** none (verification only)

- [ ] **Step 1: H-D grep — confirm no undeclared fields in `writeJSON` for `createTemplate`:**

```
grep -n '"id"\|"version_id"' internal/modules/templates/delivery/http/routes_generated.go
```

Expected: **no output** (zero matches on the CreateTemplate writeJSON call). Any match means an
undeclared field survived — fix before continuing.

- [ ] **Step 2: Full build + vet:**

```
go build ./...
go vet ./internal/modules/templates/...
```

Expected: both exit 0, no output.

- [ ] **Step 3: Full test suite:**

```
go test ./...
```

Expected: 0 FAIL lines. (All prior tests still pass; `toVersionResponse` / `toTemplateResponse`
callers in lifecycle/query/schema routes are unchanged — scope guard respected.)

- [ ] **Step 4: Commit:**

```bash
git add internal/modules/templates/delivery/http/routes_mapping.go \
        internal/modules/templates/delivery/http/routes_generated.go \
        internal/modules/templates/delivery/http/routes_typed_response_test.go \
        internal/modules/templates/delivery/http/routes_create_test.go \
        docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.3-declared-fields-only/spec.md \
        docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.3-declared-fields-only/plan.md
git commit -m "$(cat <<'EOF'
fix(templates): createTemplate drops undeclared id/version_id fields (F1.3/A3)

Replace map[string]any literal (which leaked undeclared top-level id and
version_id) with typed templatesapi.CreateTemplateResponse. Add
toAPITemplateDTO mapper (mirrors toAPIVersionDTO / ADR 0035). TDD:
TestCreateTemplate_TypedResponseShape RED->GREEN; TestCreateTemplate_Happy
updated with top-level absence assertions.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Self-Review

**Spec coverage:**
- V1 (no `id`/`version_id` top-level) → Task 1 (test) + Task 3 (handler)
- V2 (typed shape) → Task 1 (test) + Tasks 2+3 (mapper + handler)
- V3 (`TestCreateTemplate_Happy` still passes + absence assertion) → Task 4
- V4 (H-D grep clean) → Task 5 Step 1
- V5 (build + vet + suite) → Task 5 Steps 2+3

**Placeholder scan:** None. All steps have exact code + expected output.

**Type consistency:**
- `toAPITemplateDTO` returns `(templatesapi.TemplateDTO, error)` — matches both the call site
  (`tplDTO, err := toAPITemplateDTO(res.Template)`) and the assignment
  (`resp.Data.Template = tplDTO`).
- `toAPIVersionDTO` already exists — signature unchanged.
- `resp.Data.Template` field is `templatesapi.TemplateDTO`; `resp.Data.Version` is `templatesapi.VersionDTO` — both match the `CreateTemplateResponse` anonymous struct in `api.gen.go:77-79`.
- `domain.Template.LatestRevisionNumber` is `int`; cast to `int32` for `TemplateDTO.LatestRevisionNumber int32`. ✓
- `domain.Template.CurrentRevisionNumber` is `*int`; converted to `*int32`. ✓
- `domain.Template.PublishedVersionID` is `*string`; parsed to `*openapi_types.UUID`. ✓
- `uuid` package already imported in `routes_mapping.go` (used by `toAPIVersionDTO`). ✓
- `fmt` already imported in `routes_mapping.go`. ✓
- `templatesapi` already imported in `routes_generated.go`. ✓

**Scope guard (HS-6):** `toVersionResponse` and `toTemplateResponse` not removed — each has callers
in `routes_lifecycle.go`, `routes_query.go`, `routes_schema.go` (F1.4 territory). Only `routes_generated.go:64-71`
changes in the handler layer.
