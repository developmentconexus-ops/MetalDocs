package httpdelivery_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	auditapi "metaldocs/internal/modules/audit/api"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/domain"
)

// strictDecode decodes body into v rejecting any wire key absent from v's generated
// type. It is the F7.1 parity guard: the typed-response swap must keep the wire
// byte-identical, so any accidental field drift turns these tests red.
func strictDecode(t *testing.T, body []byte, v any) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("strict-decode into %T failed (unexpected wire key or shape): %v\nbody: %s", v, err, body)
	}
}

// TestAuditHandler_ListEventsTypedShape locks the GET /audit/events 200 body to the
// generated auditapi.ListAuditEventsResponse (items + CursorPage envelope), including
// second-precision occurred_at parity and the allowlisted payload map round-trip.
func TestAuditHandler_ListEventsTypedShape(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	handler := httpdelivery.NewHandler(fakeAuditQuerier{
		items: []domain.Event{{
			ID: "evt-a", OccurredAt: now, ActorID: "actor-a", Action: "audit.test",
			ResourceType: "document", ResourceID: "doc-a", PayloadJSON: `{"k":"v"}`,
			TraceID: "trace-a", TenantID: "tenant-a",
		}},
		hasMore: true,
	})
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events", "tenant-a"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp auditapi.ListAuditEventsResponse
	strictDecode(t, rec.Body.Bytes(), &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(resp.Items))
	}
	it := resp.Items[0]
	if it.Id != "evt-a" || it.ActorId != "actor-a" || it.Action != "audit.test" {
		t.Fatalf("item fields wrong: %+v", it)
	}
	if !it.OccurredAt.Equal(now) {
		t.Fatalf("occurred_at = %s, want %s (second-precision parity)", it.OccurredAt, now)
	}
	if it.Payload["k"] != "v" {
		t.Fatalf("payload not round-tripped: %+v", it.Payload)
	}
	if !resp.Page.HasMore {
		t.Fatalf("page.has_more = false, want true")
	}
	if resp.Page.NextCursor == nil || *resp.Page.NextCursor == "" {
		t.Fatalf("page.next_cursor empty, want a cursor")
	}
}

// TestAuditHandler_ExportPOSTTypedShape locks the POST /audit/events/export 202 body
// to the generated auditapi.AuditExportResponse.
func TestAuditHandler_ExportPOSTTypedShape(t *testing.T) {
	t.Parallel()

	tenantID := "aaaaaaaa-0000-0000-0000-000000000001"
	svc, _ := newExportService(t, []domain.Event{{
		ID: "evt-1", OccurredAt: time.Now().UTC().Truncate(time.Second), ActorID: "actor-test",
		Action: "doc.create", ResourceType: "document", ResourceID: "d1", PayloadJSON: `{}`, TenantID: tenantID,
	}})
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authedRequest(t, http.MethodPost, "/api/v1/audit/events/export", tenantID, "actor-test", `{"format":"csv","filter":{}}`))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", rec.Code, rec.Body.String())
	}
	var resp auditapi.AuditExportResponse
	strictDecode(t, rec.Body.Bytes(), &resp)
	if resp.ExportId == "" || resp.Status == "" || resp.SignedUrl == "" {
		t.Fatalf("export response missing fields: %+v", resp)
	}
}

// TestAuditHandler_ExportStatusTypedShape locks the GET /audit/events/export/{id}
// 200 body to the generated auditapi.AuditExportStatusResponse — the site of the
// confirmed Major (previously an untyped map[string]any).
func TestAuditHandler_ExportStatusTypedShape(t *testing.T) {
	t.Parallel()

	tenantID := "aaaaaaaa-0000-0000-0000-000000000001"
	svc, _ := newExportService(t, []domain.Event{{
		ID: "evt-1", OccurredAt: time.Now().UTC().Truncate(time.Second), ActorID: "actor-test",
		Action: "doc.create", ResourceType: "document", ResourceID: "d1", PayloadJSON: `{}`, TenantID: tenantID,
	}})
	handler := httpdelivery.NewHandler(svc).WithExporter(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, authedRequest(t, http.MethodPost, "/api/v1/audit/events/export", tenantID, "actor-test", `{"format":"csv","filter":{}}`))
	if postRec.Code != http.StatusAccepted {
		t.Fatalf("export create status = %d: %s", postRec.Code, postRec.Body.String())
	}
	var created auditapi.AuditExportResponse
	strictDecode(t, postRec.Body.Bytes(), &created)

	statusRec := httptest.NewRecorder()
	mux.ServeHTTP(statusRec, authedRequest(t, http.MethodGet, "/api/v1/audit/events/export/"+created.ExportId, tenantID, "actor-test", ""))
	if statusRec.Code != http.StatusOK {
		t.Fatalf("export status = %d: %s", statusRec.Code, statusRec.Body.String())
	}
	var status auditapi.AuditExportStatusResponse
	strictDecode(t, statusRec.Body.Bytes(), &status)
	if status.ExportId != created.ExportId {
		t.Fatalf("status export_id = %q, want %q", status.ExportId, created.ExportId)
	}
	if status.Status == "" || status.SignedUrl == "" {
		t.Fatalf("status missing fields: %+v", status)
	}
}
