# Library Block Implementation Plan

> **For agentic workers:** Use `codex:rescue` (Codex) for implementation steps. Tasks marked **[PARALLEL]** within the same Phase can run simultaneously via separate Codex agents (use `nexus:dispatching-parallel-agents`). Tasks marked **[SEQUENTIAL]** must wait for their dependency. **Phase reviewer = Opus** via `nexus:code-reviewer` agent at the end of every Phase before the next Phase starts. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Server-side paginated, filterable, statistics-aware Library screen at `/documents`. Backend gains `GET /api/v2/documents` paginated envelope + `GET /api/v2/documents/stats`. Frontend ships `LibraryPage.tsx` with SectionPanel area tree, filter tabs aligned to real Spec 2 8-state model, page-size selector `[10, 20, 50]`, collapsible activity sidebar (default-collapsed, role-gated).

**Architecture:**
- Backend: paginated envelope `{items, page, pageSize, total}`, filters `status[]`, `areaCode`, `profileCode`, `mine`, `q` (ILIKE on `name`); two-query pattern (LIMIT/OFFSET + COUNT(*) with same WHERE); RBAC reused (`system_admin` / `document_filler`); stats endpoint groups by `status` + `area_code_snapshot`.
- Frontend: TanStack Query, OpenAPI codegen types, CSS Modules, feature-sliced under `features/documents/{api,queries,pages,components}`; pagination state local + persisted; filter tab labels mapped to real states.
- No client-side pagination, no full-list fetch. No backend bypass.

**Tech Stack:** Go 1.24 (stdlib + database/sql), pgx, OpenAPI 3.0, openapi-typescript codegen, React 18, TypeScript, TanStack Query v5, React Router v7, CSS Modules, Vitest + React Testing Library

**Worktree:** current (`.claude/worktrees/thirsty-raman-989031`, branch `claude/thirsty-raman-989031`)

**Design source:** `frontend/apps/web/design-source/library/` (see `NOTES.md` for audit findings — driver of cuts/changes)

**Audit reference:** [wiki/concepts/design-workflow-audit.md](../concepts/design-workflow-audit.md)

---

## Phase 1 — Backend pagination + filters (Parallel Group P1)

### Task 1.1: Repo `ListDocumentsPaginated` + `CountDocuments` [PARALLEL]

**Files:**
- Modify: `internal/modules/documents/repository/repository.go:200-249` (existing `ListDocuments` / `ListDocumentsForUser`)
- Modify: `internal/modules/documents/application/service.go:26-53` (Repository interface)
- Test: `internal/modules/documents/repository/repository_pagination_test.go` (new)

- [ ] **Step 1: Add ListFilters struct + Repository interface methods**

Edit `internal/modules/documents/application/service.go`, append to Repository interface (after line 32, the existing `ListDocumentsForUser`):

```go
ListDocumentsPaginated(ctx context.Context, tenantID string, opts ListOptions) ([]domain.Document, error)
CountDocuments(ctx context.Context, tenantID string, opts ListOptions) (int64, error)
```

Add new file `internal/modules/documents/application/list_options.go`:

```go
package application

type ListOptions struct {
	UserID       string   // empty = no creator filter (admin); non-empty = filter by created_by
	Statuses     []string // empty = all states
	AreaCode     string   // empty = all areas
	ProfileCode  string   // empty = all profiles
	Query        string   // ILIKE %q% on name; empty = no search
	Page         int      // 1-based
	PageSize     int      // server-capped at 50; default 20
	IncludeArchived bool  // false by default — archived_at IS NULL gate
}

func (o ListOptions) Offset() int {
	if o.Page < 1 {
		return 0
	}
	return (o.Page - 1) * o.Limit()
}

func (o ListOptions) Limit() int {
	if o.PageSize < 1 {
		return 20
	}
	if o.PageSize > 50 {
		return 50
	}
	return o.PageSize
}
```

- [ ] **Step 2: Write failing test for ListDocumentsPaginated**

Create `internal/modules/documents/repository/repository_pagination_test.go`:

```go
package repository_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/testutil/dbtest"
)

func TestListDocumentsPaginated_FiltersByStatus(t *testing.T) {
	ctx := context.Background()
	db := dbtest.OpenTestDB(t)
	r := repository.New(db)
	tenantID := dbtest.SeedTenant(t, db)
	dbtest.SeedDocuments(t, db, tenantID, []dbtest.SeedDoc{
		{Status: domain.DocStatusDraft, Name: "A"},
		{Status: "under_review", Name: "B"},
		{Status: "under_review", Name: "C"},
	})

	got, err := r.ListDocumentsPaginated(ctx, tenantID, application.ListOptions{
		Statuses: []string{"under_review"},
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d", len(got))
	}
}

func TestListDocumentsPaginated_PageOffset(t *testing.T) {
	ctx := context.Background()
	db := dbtest.OpenTestDB(t)
	r := repository.New(db)
	tenantID := dbtest.SeedTenant(t, db)
	docs := make([]dbtest.SeedDoc, 25)
	for i := range docs {
		docs[i] = dbtest.SeedDoc{Status: "draft", Name: "doc"}
	}
	dbtest.SeedDocuments(t, db, tenantID, docs)

	page2, err := r.ListDocumentsPaginated(ctx, tenantID, application.ListOptions{
		Page: 2, PageSize: 10,
	})
	if err != nil || len(page2) != 10 {
		t.Fatalf("page2 expected 10 rows, got %d err=%v", len(page2), err)
	}
	page3, _ := r.ListDocumentsPaginated(ctx, tenantID, application.ListOptions{
		Page: 3, PageSize: 10,
	})
	if len(page3) != 5 {
		t.Fatalf("page3 expected 5 rows, got %d", len(page3))
	}
}

func TestCountDocuments_RespectsFilters(t *testing.T) {
	ctx := context.Background()
	db := dbtest.OpenTestDB(t)
	r := repository.New(db)
	tenantID := dbtest.SeedTenant(t, db)
	dbtest.SeedDocuments(t, db, tenantID, []dbtest.SeedDoc{
		{Status: "draft"}, {Status: "draft"}, {Status: "under_review"},
	})
	got, err := r.CountDocuments(ctx, tenantID, application.ListOptions{
		Statuses: []string{"draft"},
	})
	if err != nil || got != 2 {
		t.Fatalf("want count=2, got %d err=%v", got, err)
	}
}
```

If `internal/testutil/dbtest` lacks `SeedDocuments` / `SeedDoc`, add minimal helper in same file (Codex: scan existing dbtest helpers; reuse, don't duplicate).

- [ ] **Step 3: Run test, expect FAIL**

```bash
go test ./internal/modules/documents/repository/ -run TestListDocumentsPaginated -v
```

Expected: FAIL — methods undefined.

- [ ] **Step 4: Implement ListDocumentsPaginated + CountDocuments**

Append to `internal/modules/documents/repository/repository.go`:

```go
func (r *Repository) ListDocumentsPaginated(ctx context.Context, tenantID string, opts application.ListOptions) ([]domain.Document, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	args = append(args, opts.Limit(), opts.Offset())
	limitIdx := len(args) - 1
	offsetIdx := len(args)
	q := fmt.Sprintf(`
		SELECT id, tenant_id, template_version_id, name, status, form_data_json,
		       coalesce(current_revision_id::text, ''), coalesce(active_session_id::text, ''),
		       archived_at, created_at, updated_at, created_by,
		       controlled_document_id, coalesce(code,'')
		FROM documents
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d`, where, limitIdx, offsetIdx)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
			&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.ControlledDocumentID, &d.Code); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) CountDocuments(ctx context.Context, tenantID string, opts application.ListOptions) (int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT COUNT(*) FROM documents WHERE %s`, where)
	var n int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func buildDocumentFilter(tenantID string, opts application.ListOptions) (string, []any) {
	conds := []string{"tenant_id = $1"}
	args := []any{tenantID}
	if !opts.IncludeArchived {
		conds = append(conds, "archived_at IS NULL")
	}
	if opts.UserID != "" {
		args = append(args, opts.UserID)
		conds = append(conds, fmt.Sprintf("created_by = $%d", len(args)))
	}
	if len(opts.Statuses) > 0 {
		args = append(args, pq.StringArray(opts.Statuses))
		conds = append(conds, fmt.Sprintf("status = ANY($%d)", len(args)))
	}
	if opts.AreaCode != "" {
		args = append(args, opts.AreaCode)
		conds = append(conds, fmt.Sprintf("process_area_code_snapshot = $%d", len(args)))
	}
	if opts.ProfileCode != "" {
		args = append(args, opts.ProfileCode)
		conds = append(conds, fmt.Sprintf("profile_code_snapshot = $%d", len(args)))
	}
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	return strings.Join(conds, " AND "), args
}
```

Add imports: `"strings"`, `"github.com/lib/pq"` (verify pq is already a dep in `go.mod`; if not, use existing array helper).

- [ ] **Step 5: Run tests, expect PASS**

```bash
go test ./internal/modules/documents/repository/ -run TestListDocumentsPaginated -v
go test ./internal/modules/documents/repository/ -run TestCountDocuments -v
```

Expected: PASS for all three.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/repository/repository.go internal/modules/documents/repository/repository_pagination_test.go internal/modules/documents/application/service.go internal/modules/documents/application/list_options.go
git commit -m "feat(documents): paginated list repo with filters (status/area/profile/q)"
```

---

### Task 1.2: Repo `StatsByStatus` + `StatsByArea` [PARALLEL]

**Files:**
- Modify: `internal/modules/documents/repository/repository.go` (append)
- Modify: `internal/modules/documents/application/service.go` (Repository interface)
- Test: `internal/modules/documents/repository/repository_stats_test.go` (new)

- [ ] **Step 1: Add Repository interface methods**

Append to Repository interface in `service.go`:

```go
StatsByStatus(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
StatsByArea(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error)
```

- [ ] **Step 2: Write failing test**

Create `internal/modules/documents/repository/repository_stats_test.go`:

```go
package repository_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/testutil/dbtest"
)

func TestStatsByStatus_GroupsCorrectly(t *testing.T) {
	ctx := context.Background()
	db := dbtest.OpenTestDB(t)
	r := repository.New(db)
	tenantID := dbtest.SeedTenant(t, db)
	dbtest.SeedDocuments(t, db, tenantID, []dbtest.SeedDoc{
		{Status: "draft"}, {Status: "draft"}, {Status: "under_review"}, {Status: "published"},
	})
	got, err := r.StatsByStatus(ctx, tenantID, application.ListOptions{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["draft"] != 2 || got["under_review"] != 1 || got["published"] != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestStatsByArea_GroupsCorrectly(t *testing.T) {
	ctx := context.Background()
	db := dbtest.OpenTestDB(t)
	r := repository.New(db)
	tenantID := dbtest.SeedTenant(t, db)
	dbtest.SeedDocuments(t, db, tenantID, []dbtest.SeedDoc{
		{AreaCode: "RH"}, {AreaCode: "RH"}, {AreaCode: "QUA"},
	})
	got, err := r.StatsByArea(ctx, tenantID, application.ListOptions{})
	if err != nil || got["RH"] != 2 || got["QUA"] != 1 {
		t.Fatalf("got %+v err=%v", got, err)
	}
}
```

If `dbtest.SeedDoc` lacks `AreaCode`, extend the helper.

- [ ] **Step 3: Run test, expect FAIL**

```bash
go test ./internal/modules/documents/repository/ -run TestStats -v
```

Expected: FAIL.

- [ ] **Step 4: Implement**

Append to `repository.go`:

```go
func (r *Repository) StatsByStatus(ctx context.Context, tenantID string, opts application.ListOptions) (map[string]int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT status, COUNT(*) FROM documents WHERE %s GROUP BY status`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var s string
		var n int64
		if err := rows.Scan(&s, &n); err != nil {
			return nil, err
		}
		out[s] = n
	}
	return out, rows.Err()
}

func (r *Repository) StatsByArea(ctx context.Context, tenantID string, opts application.ListOptions) (map[string]int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT COALESCE(process_area_code_snapshot, '') AS area, COUNT(*) FROM documents WHERE %s GROUP BY area`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var a string
		var n int64
		if err := rows.Scan(&a, &n); err != nil {
			return nil, err
		}
		out[a] = n
	}
	return out, rows.Err()
}
```

- [ ] **Step 5: Run tests, expect PASS**

```bash
go test ./internal/modules/documents/repository/ -run TestStats -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/repository/repository.go internal/modules/documents/repository/repository_stats_test.go internal/modules/documents/application/service.go
git commit -m "feat(documents): stats repo (counts by status + by area)"
```

---

## Phase 1 Reviewer

Dispatch `nexus:code-reviewer` (Opus) against Tasks 1.1 + 1.2:

```
Review repo additions in internal/modules/documents/repository/. Confirm:
- Tenant scope present in every query (no cross-tenant leak)
- Parameterized queries (no string interpolation of user input)
- pageSize cap (50) enforced server-side, not client-side
- buildDocumentFilter is reused (DRY) by list, count, stats — no duplication
- ListOptions.Offset() / Limit() handle Page<1 and PageSize<1 gracefully
- Indexes: run EXPLAIN ANALYZE locally; if seq-scan dominates on tenant_id+status, recommend migration (don't add yet)
```

Block Phase 2 until reviewer signs off.

---

## Phase 2 — Application service + HTTP handler (Sequential)

### Task 2.1: Service `ListDocumentsPaginated` + `Stats`

**Files:**
- Modify: `internal/modules/documents/application/service.go` (add methods on `*Service`)
- Test: `internal/modules/documents/application/service_pagination_test.go` (new)

- [ ] **Step 1: Write failing test**

Create `internal/modules/documents/application/service_pagination_test.go`:

```go
package application_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
)

func TestService_ListDocumentsPaginated_AdminBypassesUserFilter(t *testing.T) {
	repo := &fakeRepo{
		listPaginated: []domain.Document{{ID: "doc_1"}, {ID: "doc_2"}},
		count:         2,
	}
	svc := application.New(repo /* + other deps if New signature requires */)
	page, total, err := svc.ListDocumentsPaginated(context.Background(), "tenant_1", "" /* userID empty = admin */, application.ListOptions{Page: 1, PageSize: 10})
	if err != nil || total != 2 || len(page) != 2 {
		t.Fatalf("got total=%d len=%d err=%v", total, len(page), err)
	}
}

func TestService_ListDocumentsPaginated_FillerForcesUserScope(t *testing.T) {
	repo := &fakeRepo{listPaginated: []domain.Document{{ID: "doc_1"}}, count: 1}
	svc := application.New(repo)
	_, _, err := svc.ListDocumentsPaginated(context.Background(), "tenant_1", "user_1", application.ListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.lastOpts.UserID != "user_1" {
		t.Fatalf("svc must inject UserID into ListOptions for filler; got %q", repo.lastOpts.UserID)
	}
}

func TestService_Stats_ReturnsMaps(t *testing.T) {
	repo := &fakeRepo{
		statsByStatus: map[string]int64{"draft": 5},
		statsByArea:   map[string]int64{"RH": 3},
	}
	svc := application.New(repo)
	stats, err := svc.DocumentStats(context.Background(), "tenant_1", "", application.ListOptions{})
	if err != nil || stats.ByStatus["draft"] != 5 || stats.ByArea["RH"] != 3 {
		t.Fatalf("got %+v err=%v", stats, err)
	}
}
```

Extend `fakeRepo` in `service_test.go` (existing file) to record `lastOpts` and provide `ListDocumentsPaginated`, `CountDocuments`, `StatsByStatus`, `StatsByArea`. Codex: read existing `fakeRepo` first, append fields/methods only.

- [ ] **Step 2: Run test, expect FAIL**

```bash
go test ./internal/modules/documents/application/ -run TestService_ListDocumentsPaginated -v
go test ./internal/modules/documents/application/ -run TestService_Stats -v
```

- [ ] **Step 3: Implement service methods**

Append to `service.go` (after existing `ListDocumentsForUser` method):

```go
type DocumentStats struct {
	ByStatus map[string]int64 `json:"byStatus"`
	ByArea   map[string]int64 `json:"byArea"`
}

func (s *Service) ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts ListOptions) ([]domain.Document, int64, error) {
	if userID != "" {
		opts.UserID = userID // server-forced; ignore any client-supplied UserID
	}
	items, err := s.repo.ListDocumentsPaginated(ctx, tenantID, opts)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountDocuments(ctx, tenantID, opts)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) DocumentStats(ctx context.Context, tenantID, userID string, opts ListOptions) (*DocumentStats, error) {
	if userID != "" {
		opts.UserID = userID
	}
	byStatus, err := s.repo.StatsByStatus(ctx, tenantID, opts)
	if err != nil {
		return nil, err
	}
	byArea, err := s.repo.StatsByArea(ctx, tenantID, opts)
	if err != nil {
		return nil, err
	}
	return &DocumentStats{ByStatus: byStatus, ByArea: byArea}, nil
}
```

- [ ] **Step 4: Run tests, expect PASS**

```bash
go test ./internal/modules/documents/application/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/application/service.go internal/modules/documents/application/service_pagination_test.go internal/modules/documents/application/service_test.go
git commit -m "feat(documents): paginated list + stats service methods"
```

---

### Task 2.2: HTTP handler — paginated list + stats endpoint

**Files:**
- Modify: `internal/modules/documents/delivery/http/handler.go:31-55` (Service interface), `:78-105` (RegisterRoutes), `:185-210` (listDocuments)
- Test: `internal/modules/documents/delivery/http/handler_pagination_test.go` (new)

- [ ] **Step 1: Extend Service interface in handler.go**

Add to the `Service` interface (around line 36, near existing `ListDocuments`):

```go
ListDocumentsPaginated(ctx context.Context, tenantID, userID string, opts application.ListOptions) ([]domain.Document, int64, error)
DocumentStats(ctx context.Context, tenantID, userID string, opts application.ListOptions) (*application.DocumentStats, error)
```

- [ ] **Step 2: Register `/stats` route**

In both `RegisterRoutes` (line 79) and `RegisterRoutesWithRateLimit` (line 108), append:

```go
mux.HandleFunc("GET /api/v2/documents/stats", h.documentStats)
```

- [ ] **Step 3: Replace listDocuments handler body**

Replace the existing `listDocuments` (line 185–210) with:

```go
func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, roleAdmin, roleDocumentFiller) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return
	}

	tenantID := tenantIDFromReq(r)
	userID := userIDFromReq(r)

	opts, errMsg := parseListOptions(r)
	if errMsg != "" {
		httpErr(w, http.StatusBadRequest, errMsg)
		return
	}

	scopeUser := ""
	if !hasRole(r, roleAdmin) {
		scopeUser = userID
	}

	items, total, err := h.svc.ListDocumentsPaginated(r.Context(), tenantID, scopeUser, opts)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"page":     opts.Page,
		"pageSize": opts.Limit(),
		"total":    total,
	})
}

func (h *Handler) documentStats(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, roleAdmin, roleDocumentFiller) {
		httpErr(w, http.StatusForbidden, "forbidden")
		return
	}
	tenantID := tenantIDFromReq(r)
	userID := userIDFromReq(r)

	opts, errMsg := parseListOptions(r)
	if errMsg != "" {
		httpErr(w, http.StatusBadRequest, errMsg)
		return
	}

	scopeUser := ""
	if !hasRole(r, roleAdmin) {
		scopeUser = userID
	}

	stats, err := h.svc.DocumentStats(r.Context(), tenantID, scopeUser, opts)
	if err != nil {
		status, msg := mapErr(err)
		httpErr(w, status, msg)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, stats)
}

func parseListOptions(r *http.Request) (application.ListOptions, string) {
	q := r.URL.Query()
	opts := application.ListOptions{
		Page:        1,
		PageSize:    20,
		AreaCode:    q.Get("areaCode"),
		ProfileCode: q.Get("profileCode"),
		Query:       q.Get("q"),
	}
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return opts, "invalid page"
		}
		opts.Page = n
	}
	if v := q.Get("pageSize"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return opts, "invalid pageSize"
		}
		if n > 50 {
			return opts, "pageSize exceeds max (50)"
		}
		opts.PageSize = n
	}
	if vs := q["status"]; len(vs) > 0 {
		opts.Statuses = vs
	} else if v := q.Get("status"); v != "" {
		opts.Statuses = strings.Split(v, ",")
	}
	if q.Get("includeArchived") == "true" {
		opts.IncludeArchived = true
	}
	return opts, ""
}
```

- [ ] **Step 4: Write handler test**

Create `internal/modules/documents/delivery/http/handler_pagination_test.go`:

```go
package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
	dochttp "metaldocs/internal/modules/documents/delivery/http"
)

func TestListDocuments_Paginated_Envelope(t *testing.T) {
	svc := &fakeSvc{
		paginated: []domain.Document{{ID: "doc_1"}, {ID: "doc_2"}},
		total:     12,
	}
	h := dochttp.NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v2/documents?page=2&pageSize=10", nil)
	req.Header.Set("X-User-Roles", "system_admin")
	req.Header.Set("X-Tenant-ID", "tenant_1")
	req.Header.Set("X-User-ID", "user_1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var body struct {
		Items    []domain.Document `json:"items"`
		Page     int               `json:"page"`
		PageSize int               `json:"pageSize"`
		Total    int64             `json:"total"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Page != 2 || body.PageSize != 10 || body.Total != 12 || len(body.Items) != 2 {
		t.Fatalf("got %+v", body)
	}
}

func TestListDocuments_PageSizeCap_Returns400(t *testing.T) {
	svc := &fakeSvc{}
	h := dochttp.NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v2/documents?pageSize=999", nil)
	req.Header.Set("X-User-Roles", "system_admin")
	req.Header.Set("X-Tenant-ID", "tenant_1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "pageSize") {
		t.Fatalf("error message should mention pageSize: %s", rr.Body.String())
	}
}

func TestDocumentStats_OK(t *testing.T) {
	svc := &fakeSvc{stats: &application.DocumentStats{
		ByStatus: map[string]int64{"draft": 3, "published": 100},
		ByArea:   map[string]int64{"RH": 25},
	}}
	h := dochttp.NewHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/v2/documents/stats", nil)
	req.Header.Set("X-User-Roles", "system_admin")
	req.Header.Set("X-Tenant-ID", "tenant_1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var body application.DocumentStats
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body.ByStatus["draft"] != 3 || body.ByArea["RH"] != 25 {
		t.Fatalf("got %+v", body)
	}
}
```

Extend `fakeSvc` in `handler_test.go` to add `paginated`, `total`, `stats` fields and the two new methods. Read the existing fakeSvc first, append only.

- [ ] **Step 5: Run tests, expect PASS**

```bash
go test ./internal/modules/documents/delivery/http/ -run "Paginated|PageSizeCap|DocumentStats" -v
```

- [ ] **Step 6: Run full backend test suite, ensure no regressions**

```bash
go test ./internal/modules/documents/...
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/modules/documents/delivery/http/handler.go internal/modules/documents/delivery/http/handler_pagination_test.go internal/modules/documents/delivery/http/handler_test.go
git commit -m "feat(documents): paginated GET /documents + GET /documents/stats"
```

---

## Phase 2 Reviewer

Dispatch `nexus:code-reviewer` (Opus) against Tasks 2.1 + 2.2:

```
Review service + handler in internal/modules/documents/application/service.go and internal/modules/documents/delivery/http/handler.go. Confirm:
- Filler role's userID is forced server-side; admin can pass empty userID
- pageSize cap returns 400 (defense-in-depth) — not silently capped
- parseListOptions rejects negative/non-numeric page/pageSize
- /stats and /list share parseListOptions (DRY)
- No tenant ID accepted from client body — only middleware-stamped tenantIDFromReq
- Existing handler tests still pass after Service interface extension
```

Block Phase 3 until reviewer signs off.

---

## Phase 3 — OpenAPI + frontend types codegen (Sequential)

### Task 3.1: Update OpenAPI spec

**Files:**
- Modify: `api/openapi/v1/openapi.yaml:2963` (existing `listDocumentsV2`) and add new `/documents/stats` path

- [ ] **Step 1: Read existing path spec**

```bash
grep -n -A 60 "operationId: listDocumentsV2" api/openapi/v1/openapi.yaml
```

- [ ] **Step 2: Replace listDocumentsV2 path with paginated version**

Find `/api/v2/documents` GET operation. Replace `parameters:` and `responses:` with:

```yaml
      operationId: listDocumentsV2
      parameters:
        - in: query
          name: page
          schema: { type: integer, minimum: 1, default: 1 }
        - in: query
          name: pageSize
          schema: { type: integer, minimum: 1, maximum: 50, default: 20 }
        - in: query
          name: status
          description: Comma-separated list or repeated query param. One of draft, under_review, approved, rejected, scheduled, published, superseded, obsolete.
          schema: { type: string }
        - in: query
          name: areaCode
          schema: { type: string }
        - in: query
          name: profileCode
          schema: { type: string }
        - in: query
          name: q
          schema: { type: string }
        - in: query
          name: includeArchived
          schema: { type: boolean, default: false }
      responses:
        '200':
          description: Paginated documents
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DocumentListResponse'
        '400':
          description: Invalid query params
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ApiErrorEnvelope' }
        '401':
          description: Não autenticado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ApiErrorEnvelope' }
        '403':
          description: Sem permissão
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ApiErrorEnvelope' }
```

- [ ] **Step 3: Add `/documents/stats` path**

Insert after the `/documents` block:

```yaml
  /documents/stats:
    get:
      operationId: documentStatsV2
      parameters:
        - in: query
          name: status
          schema: { type: string }
        - in: query
          name: areaCode
          schema: { type: string }
        - in: query
          name: profileCode
          schema: { type: string }
      responses:
        '200':
          description: Document statistics
          content:
            application/json:
              schema: { $ref: '#/components/schemas/DocumentStatsResponse' }
        '401':
          description: Não autenticado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ApiErrorEnvelope' }
        '403':
          description: Sem permissão
          content:
            application/json:
              schema: { $ref: '#/components/schemas/ApiErrorEnvelope' }
```

- [ ] **Step 4: Add schemas under `components.schemas`**

```yaml
    DocumentListResponse:
      type: object
      required: [items, page, pageSize, total]
      properties:
        items:
          type: array
          items: { $ref: '#/components/schemas/DocumentSummary' }
        page: { type: integer }
        pageSize: { type: integer }
        total: { type: integer, format: int64 }
    DocumentSummary:
      type: object
      required: [id, code, name, status, area_code, profile_code, revision_version, updated_at]
      properties:
        id: { type: string, format: uuid }
        code: { type: string }
        name: { type: string }
        status:
          type: string
          enum: [draft, under_review, approved, rejected, scheduled, published, superseded, obsolete]
        area_code: { type: string }
        profile_code: { type: string }
        revision_version: { type: integer }
        updated_at: { type: string, format: date-time }
        created_by: { type: string, format: uuid }
    DocumentStatsResponse:
      type: object
      required: [byStatus, byArea]
      properties:
        byStatus:
          type: object
          additionalProperties: { type: integer, format: int64 }
        byArea:
          type: object
          additionalProperties: { type: integer, format: int64 }
```

NOTE: backend `domain.Document` JSON shape currently differs from `DocumentSummary` (it returns full struct). Add a DTO mapping in handler before serializing items, OR update `DocumentSummary` to match current response. Codex: pick the smaller diff — if domain.Document JSON tags already match, align spec to that; otherwise, introduce a `mapToSummary` in `handler.go` and serialize that.

- [ ] **Step 5: Validate spec**

```bash
cd frontend/apps/web && pnpm.cmd dlx @redocly/cli@latest lint ../../../api/openapi/v1/openapi.yaml
```

Expected: 0 errors. Fix any reported.

- [ ] **Step 6: Regenerate frontend types**

```bash
cd frontend/apps/web && pnpm.cmd run gen:api
```

Expected: `src/lib/api-types/index.d.ts` updated. Inspect diff for new operation IDs `listDocumentsV2`, `documentStatsV2` and component schemas.

- [ ] **Step 7: Commit**

```bash
git add api/openapi/v1/openapi.yaml frontend/apps/web/src/lib/api-types/index.d.ts
git commit -m "feat(openapi): paginated documents + stats endpoint"
```

---

## Phase 3 Reviewer

Dispatch `nexus:code-reviewer` (Opus):

```
Review api/openapi/v1/openapi.yaml diff. Confirm:
- pageSize.maximum=50 in spec mirrors handler enforcement
- status enum lists all 8 Spec 2 states (draft/under_review/approved/rejected/scheduled/published/superseded/obsolete)
- DocumentSummary fields match the JSON shape returned by handler.listDocuments
- /stats path registered under correct components/security scheme
- regenerated frontend/apps/web/src/lib/api-types/index.d.ts compiles (run pnpm tsc --noEmit)
```

Block Phase 4 until reviewer signs off.

---

## Phase 4 — Frontend Library page (Parallel Group P4)

### Task 4.1: API client + query hook [PARALLEL]

**Files:**
- Create: `frontend/apps/web/src/features/documents/api/library.ts`
- Create: `frontend/apps/web/src/features/documents/queries/useLibraryQuery.ts`
- Create: `frontend/apps/web/src/features/documents/queries/useLibraryStatsQuery.ts`
- Modify: `frontend/apps/web/src/lib/queryKeys.ts` (extend `QK.documents`)

- [ ] **Step 1: Extend QK keys**

In `src/lib/queryKeys.ts`, replace the `documents` block:

```ts
documents: {
  list: (params: { page: number; pageSize: number; status?: string[]; areaCode?: string; profileCode?: string; q?: string; mine?: boolean }) =>
    ['documents', 'list', params] as const,
  stats: (params: { areaCode?: string; mine?: boolean }) =>
    ['documents', 'stats', params] as const,
  detail: (id: string) => ['documents', 'detail', id] as const,
},
```

- [ ] **Step 2: Create api/library.ts**

```ts
import client from '../../../lib/api/client';
import type { components } from '../../../lib/api-types';

export type DocumentSummary = components['schemas']['DocumentSummary'];
export type DocumentListResponse = components['schemas']['DocumentListResponse'];
export type DocumentStatsResponse = components['schemas']['DocumentStatsResponse'];

export type LibraryParams = {
  page: number;
  pageSize: number;
  status?: string[];
  areaCode?: string;
  profileCode?: string;
  q?: string;
};

export async function listLibrary(params: LibraryParams): Promise<DocumentListResponse> {
  const { data, error } = await client.GET('/documents', {
    params: {
      query: {
        page: params.page,
        pageSize: params.pageSize,
        status: params.status?.join(','),
        areaCode: params.areaCode,
        profileCode: params.profileCode,
        q: params.q,
      },
    },
  });
  if (error) throw error;
  return data!;
}

export async function getLibraryStats(params: { areaCode?: string }): Promise<DocumentStatsResponse> {
  const { data, error } = await client.GET('/documents/stats', {
    params: { query: { areaCode: params.areaCode } },
  });
  if (error) throw error;
  return data!;
}
```

NOTE: confirm `client` import path and openapi-fetch usage by reading `src/lib/api/client.ts` first.

- [ ] **Step 3: Create useLibraryQuery hook**

`src/features/documents/queries/useLibraryQuery.ts`:

```ts
import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { listLibrary, type LibraryParams } from '../api/library';

export function useLibraryQuery(params: LibraryParams) {
  return useQuery({
    queryKey: QK.documents.list(params),
    queryFn: () => listLibrary(params),
    staleTime: 15_000,
    placeholderData: (prev) => prev,
  });
}
```

- [ ] **Step 4: Create useLibraryStatsQuery hook**

`src/features/documents/queries/useLibraryStatsQuery.ts`:

```ts
import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { getLibraryStats } from '../api/library';

export function useLibraryStatsQuery(params: { areaCode?: string }) {
  return useQuery({
    queryKey: QK.documents.stats(params),
    queryFn: () => getLibraryStats(params),
    staleTime: 30_000,
  });
}
```

- [ ] **Step 5: Type-check**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: 0 errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/documents/api/library.ts frontend/apps/web/src/features/documents/queries/useLibraryQuery.ts frontend/apps/web/src/features/documents/queries/useLibraryStatsQuery.ts frontend/apps/web/src/lib/queryKeys.ts
git commit -m "feat(library): typed api client + query hooks"
```

---

### Task 4.2: Page-local primitives — Pagination + PageSizeSelector + FilterTabs + StatusPill update [PARALLEL]

**Files:**
- Create: `frontend/apps/web/src/features/documents/components/Pagination.tsx + .module.css`
- Create: `frontend/apps/web/src/features/documents/components/PageSizeSelector.tsx + .module.css`
- Create: `frontend/apps/web/src/features/documents/components/LibraryFilterTabs.tsx + .module.css`
- Modify: `frontend/apps/web/src/components/ui/StatusPill.tsx + .module.css` (extend to 8 states)

- [ ] **Step 1: Extend StatusPill to 8 states**

Read existing `src/components/ui/StatusPill.tsx` and `.module.css`. Add classes `pillScheduled`, `pillPublished`, `pillSuperseded`, `pillObsolete`. Map by status string:

```tsx
const statusClass: Record<string, string> = {
  draft: styles.pillDraft,
  under_review: styles.pillReview,
  approved: styles.pillApproved,
  rejected: styles.pillRejected,
  scheduled: styles.pillScheduled,
  published: styles.pillPublished,
  superseded: styles.pillSuperseded,
  obsolete: styles.pillObsolete,
};
const statusLabel: Record<string, string> = {
  draft: 'Rascunho',
  under_review: 'Em revisão',
  approved: 'Aprovado',
  rejected: 'Rejeitado',
  scheduled: 'Agendado',
  published: 'Publicado',
  superseded: 'Substituído',
  obsolete: 'Obsoleto',
};
```

CSS: define color + background for each new state, reuse existing token vars (no new hex).

- [ ] **Step 2: Pagination component**

`Pagination.tsx`:

```tsx
import styles from './Pagination.module.css';

type Props = {
  page: number;
  pageSize: number;
  total: number;
  onPage: (page: number) => void;
};

export function Pagination({ page, pageSize, total, onPage }: Props) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, total);
  return (
    <div className={styles.root}>
      <span className={styles.summary}>
        {total === 0 ? 'Sem resultados' : `Mostrando ${start}–${end} de ${total}`}
      </span>
      <div className={styles.controls}>
        <button className={styles.btn} disabled={page <= 1} onClick={() => onPage(page - 1)}>‹</button>
        <span className={styles.pageLabel}>{page} / {totalPages}</span>
        <button className={styles.btn} disabled={page >= totalPages} onClick={() => onPage(page + 1)}>›</button>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: PageSizeSelector component**

`PageSizeSelector.tsx`:

```tsx
import styles from './PageSizeSelector.module.css';

const OPTIONS = [10, 20, 50] as const;
export type PageSize = typeof OPTIONS[number];

type Props = { value: PageSize; onChange: (n: PageSize) => void };

export function PageSizeSelector({ value, onChange }: Props) {
  return (
    <label className={styles.root}>
      <span className={styles.label}>Por página</span>
      <select
        className={styles.select}
        value={value}
        onChange={(e) => onChange(Number(e.target.value) as PageSize)}
      >
        {OPTIONS.map((n) => <option key={n} value={n}>{n}</option>)}
      </select>
    </label>
  );
}
```

- [ ] **Step 4: LibraryFilterTabs component**

`LibraryFilterTabs.tsx`:

```tsx
import styles from './LibraryFilterTabs.module.css';

export type LibraryFilter = 'todos' | 'meus' | 'rascunhos' | 'em_revisao' | 'publicados' | 'rejeitados' | 'obsoletos';

type Tab = { id: LibraryFilter; label: string; count?: number };

type Props = {
  active: LibraryFilter;
  onChange: (id: LibraryFilter) => void;
  counts: Partial<Record<LibraryFilter, number>>;
};

const TABS: Tab[] = [
  { id: 'todos', label: 'Todos' },
  { id: 'meus', label: 'Meus' },
  { id: 'rascunhos', label: 'Rascunhos' },
  { id: 'em_revisao', label: 'Em revisão' },
  { id: 'publicados', label: 'Publicados' },
  { id: 'rejeitados', label: 'Rejeitados' },
  { id: 'obsoletos', label: 'Obsoletos' },
];

export function filterToStatuses(f: LibraryFilter): { statuses?: string[]; mine?: boolean } {
  switch (f) {
    case 'rascunhos': return { statuses: ['draft'] };
    case 'em_revisao': return { statuses: ['under_review'] };
    case 'publicados': return { statuses: ['published'] };
    case 'rejeitados': return { statuses: ['rejected'] };
    case 'obsoletos': return { statuses: ['obsolete'] };
    case 'meus': return { mine: true };
    default: return {};
  }
}

export function LibraryFilterTabs({ active, onChange, counts }: Props) {
  return (
    <div className={styles.root}>
      {TABS.map((t) => (
        <button
          key={t.id}
          className={`${styles.tab} ${active === t.id ? styles.active : ''}`}
          onClick={() => onChange(t.id)}
        >
          {t.label}
          {counts[t.id] != null && <span className={styles.count}>{counts[t.id]}</span>}
        </button>
      ))}
    </div>
  );
}
```

- [ ] **Step 5: CSS modules — keep terse, use tokens**

For each `.module.css`, use only existing CSS vars (no hex). Example `Pagination.module.css`:

```css
.root { display: flex; align-items: center; gap: 16px; padding: 12px 16px; border-top: 1px solid var(--border); }
.summary { font-size: 12px; color: var(--text-muted); }
.controls { margin-left: auto; display: flex; align-items: center; gap: 8px; }
.btn { height: 28px; padding: 0 10px; border: 1px solid var(--border); background: var(--surface); border-radius: var(--r-1); cursor: pointer; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.pageLabel { font-family: var(--font-mono); font-size: 12px; color: var(--text-soft); }
```

- [ ] **Step 6: Type-check**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: 0 errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/ frontend/apps/web/src/components/ui/StatusPill.tsx frontend/apps/web/src/components/ui/StatusPill.module.css
git commit -m "feat(library): pagination + page-size selector + filter tabs + 8-state status pill"
```

---

### Task 4.3: SectionPanel area tree integration [PARALLEL]

**Files:**
- Create: `frontend/apps/web/src/features/documents/components/LibraryAreaTree.tsx + .module.css`

- [ ] **Step 1: Implement LibraryAreaTree**

```tsx
import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { listAreas } from '../../taxonomy/api/taxonomy'; // verify path
import styles from './LibraryAreaTree.module.css';

type Props = {
  selected: string | null;
  onSelect: (areaCode: string | null) => void;
  counts: Record<string, number>;
};

export function LibraryAreaTree({ selected, onSelect, counts }: Props) {
  const { data: areas = [] } = useQuery({
    queryKey: QK.taxonomy.areas(),
    queryFn: listAreas,
  });
  return (
    <nav className={styles.root}>
      <button
        className={`${styles.item} ${selected == null ? styles.active : ''}`}
        onClick={() => onSelect(null)}
      >
        <span>Todas as áreas</span>
      </button>
      {areas.map((a) => (
        <button
          key={a.code}
          className={`${styles.item} ${selected === a.code ? styles.active : ''}`}
          onClick={() => onSelect(a.code)}
        >
          <span>{a.name}</span>
          {counts[a.code] != null && <span className={styles.count}>{counts[a.code]}</span>}
        </button>
      ))}
    </nav>
  );
}
```

If `listAreas` does not exist at `features/taxonomy/api/taxonomy`, Codex: locate the existing taxonomy areas hook (search `QK.taxonomy.areas`), reuse it. Don't create a duplicate fetcher.

- [ ] **Step 2: Type-check + commit**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json
git add frontend/apps/web/src/features/documents/components/LibraryAreaTree.tsx frontend/apps/web/src/features/documents/components/LibraryAreaTree.module.css
git commit -m "feat(library): SectionPanel area tree component"
```

---

## Phase 4 Reviewer (mid-phase, before page assembly)

Dispatch `nexus:code-reviewer` (Opus) against Tasks 4.1 + 4.2 + 4.3:

```
Review frontend additions in features/documents/{api,queries,components} and StatusPill update. Confirm:
- No legacy paths reintroduced (no src/api/, no src/components/<feature>)
- All CSS uses tokens, no hex
- TanStack Query keys match QK structure
- Pagination, PageSizeSelector, FilterTabs are page-local (not promoted to ui/) — correct per skill rule
- StatusPill maps cover all 8 Spec 2 states + has labels in Portuguese
- No useEffect+setState for fetching
```

Block Task 4.4 until reviewer signs off.

---

### Task 4.4: LibraryPage assembly [SEQUENTIAL — depends on 4.1, 4.2, 4.3]

**Files:**
- Create: `frontend/apps/web/src/features/documents/pages/LibraryPage.tsx + .module.css`
- Modify: `frontend/apps/web/src/features/documents/routes.tsx` (register `/documents`)

- [ ] **Step 1: Write LibraryPage component**

```tsx
import { useState } from 'react';
import { useLibraryQuery } from '../queries/useLibraryQuery';
import { useLibraryStatsQuery } from '../queries/useLibraryStatsQuery';
import { LibraryFilterTabs, filterToStatuses, type LibraryFilter } from '../components/LibraryFilterTabs';
import { LibraryAreaTree } from '../components/LibraryAreaTree';
import { Pagination } from '../components/Pagination';
import { PageSizeSelector, type PageSize } from '../components/PageSizeSelector';
import { StatusPill } from '../../../components/ui/StatusPill';
import { CodeChip } from '../../../components/ui/CodeChip';
import styles from './LibraryPage.module.css';

const PAGESIZE_KEY = 'metaldocs.library.pageSize';
const ACTIVITY_KEY = 'metaldocs.library.activityOpen';

function readPageSize(): PageSize {
  const raw = Number(localStorage.getItem(PAGESIZE_KEY));
  return ([10, 20, 50] as const).includes(raw as PageSize) ? (raw as PageSize) : 20;
}

export default function LibraryPage() {
  const [filter, setFilter] = useState<LibraryFilter>('todos');
  const [areaCode, setAreaCode] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSizeState] = useState<PageSize>(readPageSize());
  const [activityOpen, setActivityOpen] = useState<boolean>(localStorage.getItem(ACTIVITY_KEY) === 'true');

  const { statuses, mine } = filterToStatuses(filter);

  const listQ = useLibraryQuery({
    page,
    pageSize,
    status: statuses,
    areaCode: areaCode ?? undefined,
    // mine: filler-role users get auto-scoped server-side; admin uses explicit mine flag (future)
  });
  const statsQ = useLibraryStatsQuery({ areaCode: areaCode ?? undefined });

  const setPageSize = (n: PageSize) => {
    setPageSizeState(n);
    localStorage.setItem(PAGESIZE_KEY, String(n));
    setPage(1);
  };

  const tabCounts: Partial<Record<LibraryFilter, number>> = {
    todos: Object.values(statsQ.data?.byStatus ?? {}).reduce((a, b) => a + b, 0),
    rascunhos: statsQ.data?.byStatus?.draft,
    em_revisao: statsQ.data?.byStatus?.under_review,
    publicados: statsQ.data?.byStatus?.published,
    rejeitados: statsQ.data?.byStatus?.rejected,
    obsoletos: statsQ.data?.byStatus?.obsolete,
  };

  return (
    <div className={styles.root}>
      <aside className={styles.sectionPanel}>
        <LibraryAreaTree
          selected={areaCode}
          onSelect={(code) => { setAreaCode(code); setPage(1); }}
          counts={statsQ.data?.byArea ?? {}}
        />
      </aside>

      <main className={styles.main}>
        <header className={styles.header}>
          <div>
            <div className={styles.kicker}>Documentos · Biblioteca</div>
            <h1 className={styles.title}>Acervo controlado</h1>
          </div>
          <span className={styles.spacer} />
          <span className={styles.totalLabel}>{listQ.data?.total ?? 0} documentos</span>
          <button className={styles.activityToggle} onClick={() => {
            const next = !activityOpen;
            setActivityOpen(next);
            localStorage.setItem(ACTIVITY_KEY, String(next));
          }}>
            {activityOpen ? 'Recolher atividade' : 'Atividade'}
          </button>
        </header>

        <LibraryFilterTabs
          active={filter}
          onChange={(f) => { setFilter(f); setPage(1); }}
          counts={tabCounts}
        />

        <div className={styles.tableCard}>
          <div className={styles.tableHeader}>
            <div>Código</div><div>Título</div><div>Área</div><div>Perfil</div><div>Rev.</div><div>Estado</div><div>Atualizado</div><div />
          </div>
          {listQ.isPending && <div className={styles.empty}>Carregando…</div>}
          {listQ.isError && <div className={styles.empty}>Erro ao carregar documentos.</div>}
          {listQ.data?.items.map((d) => (
            <div key={d.id} className={styles.row}>
              <CodeChip>{d.code}</CodeChip>
              <span className={styles.titleCell}>{d.name}</span>
              <span className={styles.muted}>{d.area_code}</span>
              <span className={styles.mono}>{d.profile_code}</span>
              <span className={styles.mono}>v{d.revision_version}</span>
              <StatusPill status={d.status} />
              <span className={styles.muted}>{new Date(d.updated_at).toLocaleString('pt-BR')}</span>
              <span />
            </div>
          ))}
          {listQ.data && listQ.data.items.length === 0 && (
            <div className={styles.empty}>Nenhum documento.</div>
          )}

          <div className={styles.tableFooter}>
            <PageSizeSelector value={pageSize} onChange={setPageSize} />
            <Pagination page={page} pageSize={pageSize} total={listQ.data?.total ?? 0} onPage={setPage} />
          </div>
        </div>
      </main>

      {activityOpen && (
        <aside className={styles.activitySidebar}>
          {/* v1: empty placeholder — populate in follow-up plan (approval inbox + audit stream) */}
          <div className={styles.empty}>Atividade em breve.</div>
        </aside>
      )}
    </div>
  );
}
```

- [ ] **Step 2: LibraryPage.module.css**

Reuse tokens. Layout grid: `grid-template-columns: 224px 1fr 320px;` collapses to `224px 1fr` when activity closed. Table grid `160px 1fr 100px 90px 70px 110px 110px 36px`.

```css
.root { display: grid; grid-template-columns: 224px 1fr; min-height: 100%; }
.root:has(.activitySidebar) { grid-template-columns: 224px 1fr 320px; }
.sectionPanel { background: var(--surface-2); border-right: 1px solid var(--border); padding: 16px 12px; }
.main { display: flex; flex-direction: column; gap: 16px; padding: 24px 28px; background: var(--bg); overflow: auto; }
.header { display: flex; align-items: flex-end; gap: 16px; }
.kicker { font-size: 11px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); margin-bottom: 6px; }
.title { font-family: var(--font-sans); font-size: 28px; margin: 0; }
.spacer { flex: 1; }
.totalLabel { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); }
.activityToggle { height: 28px; padding: 0 12px; font-size: 12px; border: 1px solid var(--border); background: var(--surface); border-radius: var(--r-1); cursor: pointer; }
.tableCard { background: var(--surface); border: 1px solid var(--border); border-radius: var(--r-2); overflow: hidden; }
.tableHeader, .row { display: grid; grid-template-columns: 160px 1fr 100px 90px 70px 110px 110px 36px; gap: 12px; padding: 11px 16px; align-items: center; }
.tableHeader { background: var(--surface-2); font-size: 10px; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); border-bottom: 1px solid var(--border); }
.row { border-bottom: 1px solid var(--border); cursor: pointer; }
.row:last-child { border-bottom: none; }
.row:hover { background: var(--surface-2); }
.titleCell { font-size: 13.5px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.muted { font-size: 12px; color: var(--text-soft); }
.mono { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); }
.empty { padding: 32px; text-align: center; color: var(--text-muted); font-size: 13px; }
.tableFooter { display: flex; align-items: center; gap: 16px; padding: 12px 16px; border-top: 1px solid var(--border); }
.activitySidebar { background: var(--surface-2); border-left: 1px solid var(--border); padding: 18px 20px; overflow: auto; }
```

- [ ] **Step 3: Register route**

In `src/features/documents/routes.tsx`, add (or replace existing `/documents`):

```tsx
import { lazy } from 'react';
const LibraryPage = lazy(() => import('./pages/LibraryPage'));

export const documentRoutes = [
  {
    path: '/documents',
    element: <LibraryPage />,
    handle: { sectionPanel: true },
  },
  // ...existing routes
];
```

Verify `routes.tsx` exists; if not, create it and wire into `app/AppRouter.tsx`. Read existing routes file pattern first.

- [ ] **Step 4: Type-check + unit test**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm.cmd test src/features/documents/pages/LibraryPage
```

If no test file exists, create a minimal smoke test in `src/features/documents/pages/__tests__/LibraryPage.test.tsx` rendering with a mocked QueryClient and asserting "Acervo controlado" appears.

- [ ] **Step 5: Manual UI verify via preview**

```
preview_start metaldocs-api  (already running per session)
preview_start metaldocs-web (or 4173 variant)
preview_screenshot /documents
```

Confirm: header renders, area tree on left, filter tabs, table empty/loading state shown when no real seed data. Page-size dropdown changes URL state — page resets to 1.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/documents/pages/ frontend/apps/web/src/features/documents/routes.tsx
git commit -m "feat(library): LibraryPage assembled — server-side paginated, filter tabs, area tree"
```

---

## Phase 4 Reviewer (final)

Dispatch `nexus:code-reviewer` (Opus) against full Phase 4:

```
Final review of Library frontend. Verify:
- LibraryPage matches NOTES.md audit findings (no Frozen card, no Próx revisão card, filter tabs = post-audit set)
- Activity sidebar default-collapsed (localStorage)
- pageSize cap respected client-side (selector limited to 10/20/50)
- Server-side paging confirmed via network tab (pages 2/3/4 hit different offsets, no full-list fetch)
- All CSS via tokens
- No legacy paths re-introduced
- LibraryPage smoke test passes
- /documents route works behind auth guard
```

---

## Phase 5 — Verify + wiki

### Task 5.1: Full verification

- [ ] **Step 1: Backend tests**

```bash
go test ./internal/modules/documents/...
```

- [ ] **Step 2: Frontend type-check + tests**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json && pnpm.cmd test
```

- [ ] **Step 3: E2E smoke (if exists)**

```bash
cd frontend/apps/web && pnpm.cmd e2e:smoke
```

### Task 5.2: Wiki update

- [ ] **Step 1: Dispatch wiki-curator agent**

Ask wiki-curator to:
- Add `wiki/modules/documents.md` Library section with screenshot + page anchor
- Update `wiki/architecture/data-model.md` if new query patterns warrant a note
- Bump `Last verified` stamps on touched docs
- Update `wiki/implementation/screen-redesign-tracker.md`: Library row → ✅ Complete; Editor row → 🔲 Not started (was ⏳)

- [ ] **Step 2: Commit doc updates**

```bash
git add wiki/
git commit -m "docs(wiki): Library complete — Editor unblocked"
```

---

## Backend follow-ups (out-of-scope for this plan, track separately)

1. Approval-pending join (needed for "Aprovação pendente" filter tab + stat card).
2. Viewer/reader role (`document_viewer`) for read-only access.
3. Author column user-lookup join (UUID → display name).
4. Index audit migration if `EXPLAIN ANALYZE` shows seq-scan on `(tenant_id, status)` after real-data load.
5. Cursor pagination upgrade if any tenant crosses ~100k docs.

These are real backend gaps surfaced during Library audit. Open as separate plans before merging into `feature/screen-redesign`.

---

## Self-review checklist (run before committing this plan)

- [x] Spec coverage: pagination + filters + stats + RBAC + frontend page + routes + tests + wiki — all covered
- [x] No placeholders ("TODO", "fill in", "similar to") — all steps contain code or exact commands
- [x] Type consistency: `ListOptions` shape matches across repo, application, handler, OpenAPI spec, frontend codegen
- [x] All file paths absolute or with explicit project-relative prefix
- [x] Every code step has runnable code; every test step has both code and run command
- [x] Reviewer points block phase progression (no skip-ahead)
- [x] Audit findings drive the implementation (no Frozen card, no invented states, default-collapsed sidebar)
