package httpdelivery_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "metaldocs/internal/modules/audit/application"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/audit/infrastructure/memory"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

func TestAuditHandler_InvalidLimitIsProblemJSON(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events?limit=invalid", "tenant-a")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", got)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected code VALIDATION_ERROR, got %q", body.Code)
	}
}

func TestAuditHandler_RequiresAuthenticatedContext(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuditHandler_ListEventsIsTenantScoped(t *testing.T) {
	t.Parallel()

	writer := memory.NewWriter()
	now := time.Now().UTC()
	if err := writer.Record(t.Context(), domain.Event{
		ID:           "evt-a",
		OccurredAt:   now,
		ActorID:      "actor-a",
		Action:       "audit.test",
		ResourceType: "document",
		ResourceID:   "doc-a",
		PayloadJSON:  `{"tenant":"a"}`,
		TraceID:      "trace-a",
		TenantID:     "tenant-a",
	}); err != nil {
		t.Fatalf("record tenant-a event: %v", err)
	}
	if err := writer.Record(t.Context(), domain.Event{
		ID:           "evt-b",
		OccurredAt:   now.Add(time.Second),
		ActorID:      "actor-b",
		Action:       "audit.test",
		ResourceType: "document",
		ResourceID:   "doc-b",
		PayloadJSON:  `{"tenant":"b"}`,
		TraceID:      "trace-b",
		TenantID:     "tenant-b",
	}); err != nil {
		t.Fatalf("record tenant-b event: %v", err)
	}

	service := application.NewService(writer)
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events", "tenant-a")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(body.Items))
	}
	if body.Items[0].ID != "evt-a" {
		t.Fatalf("expected tenant-a event, got %q", body.Items[0].ID)
	}
}

func authenticatedAuditRequest(method, target, tenantID string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	return req.WithContext(ctx)
}
