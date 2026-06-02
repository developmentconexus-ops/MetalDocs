package httpdelivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/audit/application"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/audit/infrastructure/memory"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

func authedRequest(t *testing.T, method, target, tenantID, actorID string, body string) *http.Request {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	var req *http.Request
	if reader == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, reader)
	}
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, actorID, []iamdomain.Role{iamdomain.RoleSystemAdmin})
	return req.WithContext(ctx)
}

func newExportService(t *testing.T, events []domain.Event) (*application.Service, *memory.ExportJobRepository) {
	t.Helper()
	writer := memory.NewWriter()
	for _, e := range events {
		if err := writer.Record(context.Background(), e); err != nil {
			t.Fatalf("record %s: %v", e.ID, err)
		}
	}
	exports := memory.NewExportJobRepository()
	svc := application.NewService(writer).WithExports(writer, exports, writer, func(j domain.ExportJob) string {
		return "/api/v1/audit/events/export/" + j.ID + "/download?token=" + j.DownloadToken
	})
	return svc, exports
}

func TestAuditHandler_MultiAxisFilter(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []domain.Event{
		{ID: "evt-1", OccurredAt: now, ActorID: "alice", Action: "auth.login", ResourceType: "auth_session", ResourceID: "s1", PayloadJSON: `{"x":1}`, TenantID: "tenant-a"},
		{ID: "evt-2", OccurredAt: now.Add(time.Second), ActorID: "bob", Action: "auth.logout", ResourceType: "auth_session", ResourceID: "s2", PayloadJSON: `{"x":2}`, TenantID: "tenant-a"},
		{ID: "evt-3", OccurredAt: now.Add(2 * time.Second), ActorID: "alice", Action: "doc.create", ResourceType: "document", ResourceID: "d1", PayloadJSON: `{"x":3}`, TenantID: "tenant-a"},
	}
	svc, _ := newExportService(t, events)
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := authedRequest(t, http.MethodGet, "/api/v1/audit/events?actorId=alice&action=auth.*", "tenant-a", "actor-test", "")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "evt-1" {
		t.Fatalf("expected only evt-1, got %+v", body.Items)
	}
}

func TestAuditHandler_CursorPagination(t *testing.T) {
	t.Parallel()

	base := time.Now().UTC().Truncate(time.Millisecond)
	events := make([]domain.Event, 0, 5)
	for i := 0; i < 5; i++ {
		events = append(events, domain.Event{
			ID:           "evt-" + string(rune('a'+i)),
			OccurredAt:   base.Add(time.Duration(i) * time.Second),
			ActorID:      "alice",
			Action:       "doc.create",
			ResourceType: "document",
			ResourceID:   "d" + string(rune('1'+i)),
			PayloadJSON:  `{}`,
			TenantID:     "tenant-a",
		})
	}
	svc, _ := newExportService(t, events)
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, authedRequest(t, http.MethodGet, "/api/v1/audit/events?limit=2", "tenant-a", "actor-test", ""))
	var page1 struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		NextCursor string `json:"nextCursor"`
		HasMore    bool   `json:"hasMore"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &page1); err != nil {
		t.Fatalf("decode page1: %v", err)
	}
	if len(page1.Items) != 2 || !page1.HasMore || page1.NextCursor == "" {
		t.Fatalf("unexpected page1: %+v", page1)
	}
	if page1.Items[0].ID != "evt-e" || page1.Items[1].ID != "evt-d" {
		t.Fatalf("page1 wrong order: %+v", page1.Items)
	}

	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, authedRequest(t, http.MethodGet, "/api/v1/audit/events?limit=2&cursor="+page1.NextCursor, "tenant-a", "actor-test", ""))
	var page2 struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &page2); err != nil {
		t.Fatalf("decode page2: %v", err)
	}
	if len(page2.Items) != 2 || page2.Items[0].ID != "evt-c" || page2.Items[1].ID != "evt-b" {
		t.Fatalf("page2 wrong: %+v", page2.Items)
	}
}

func TestAuditHandler_ExportCSVThenDownloadIsTenantScoped(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []domain.Event{
		{ID: "evt-1", OccurredAt: now, ActorID: "alice", Action: "doc.create", ResourceType: "document", ResourceID: "d1", PayloadJSON: `{"k":"v"}`, TenantID: "tenant-a"},
		{ID: "evt-2", OccurredAt: now.Add(time.Second), ActorID: "bob", Action: "doc.create", ResourceType: "document", ResourceID: "d2", PayloadJSON: `{"k":"v2"}`, TenantID: "tenant-a"},
		{ID: "evt-x", OccurredAt: now, ActorID: "carol", Action: "doc.create", ResourceType: "document", ResourceID: "dx", PayloadJSON: `{}`, TenantID: "tenant-b"},
	}
	svc, _ := newExportService(t, events)
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authedRequest(t, http.MethodPost, "/api/v1/audit/events/export", "tenant-a", "actor-test", `{"format":"csv","filter":{}}`))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("export status = %d (%s)", rec.Code, rec.Body.String())
	}
	var export struct {
		ExportID  string `json:"exportId"`
		Status    string `json:"status"`
		SignedURL string `json:"signedUrl"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if export.Status != "ready" || export.SignedURL == "" {
		t.Fatalf("expected ready+signed, got %+v", export)
	}

	dlReq := httptest.NewRequest(http.MethodGet, export.SignedURL, nil)
	dlRec := httptest.NewRecorder()
	mux.ServeHTTP(dlRec, dlReq)
	if dlRec.Code != http.StatusOK {
		t.Fatalf("download status = %d (%s)", dlRec.Code, dlRec.Body.String())
	}
	if ct := dlRec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("content-type = %q", ct)
	}
	bodyStr := dlRec.Body.String()
	if strings.Count(bodyStr, "\n") < 3 {
		t.Fatalf("csv shorter than expected (need header + 2 rows): %q", bodyStr)
	}
	if strings.Contains(bodyStr, "tenant-b") || strings.Contains(bodyStr, "carol") || strings.Contains(bodyStr, "evt-x") {
		t.Fatalf("csv leaked tenant-b row: %q", bodyStr)
	}
}

func TestAuditHandler_ExportJobActorScoped(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []domain.Event{
		{ID: "evt-1", OccurredAt: now, ActorID: "actor-test", Action: "doc.create", ResourceType: "document", ResourceID: "d1", PayloadJSON: `{}`, TenantID: "tenant-a"},
	}
	svc, exports := newExportService(t, events)
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	other := domain.ExportJob{
		ID: "exp-other", TenantID: "tenant-a", ActorID: "someone-else",
		Format: domain.ExportFormatCSV, FilterJSON: `{}`, Status: domain.ExportStatusReady,
		DownloadToken: "tok-other", ExpiresAt: now.Add(time.Hour), CreatedAt: now,
	}
	if err := exports.Save(context.Background(), other); err != nil {
		t.Fatalf("seed: %v", err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authedRequest(t, http.MethodGet, "/api/v1/audit/events/export/exp-other", "tenant-a", "actor-test", ""))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-actor lookup, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestAuditHandler_ExportRejectsBadFormat(t *testing.T) {
	t.Parallel()

	svc, _ := newExportService(t, nil)
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authedRequest(t, http.MethodPost, "/api/v1/audit/events/export", "tenant-a", "actor-test", `{"format":"xml","filter":{}}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", rec.Code, rec.Body.String())
	}
}
