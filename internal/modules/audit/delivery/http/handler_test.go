package httpdelivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	handler.Mount(mux)

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

	if body.Code != "request.invalid" {
		t.Fatalf("expected code VALIDATION_ERROR, got %q", body.Code)
	}
}

func TestAuditHandler_RequiresAuthenticatedContext(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.Mount(mux)

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
	handler.Mount(mux)

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

func TestAuditHandler_InvalidStoredPayloadFailsHonestly(t *testing.T) {
	t.Parallel()

	handler := httpdelivery.NewHandler(fakeAuditQuerier{
		items: []domain.Event{{
			ID:           "evt-bad",
			OccurredAt:   time.Now().UTC(),
			ActorID:      "actor-a",
			Action:       "audit.test",
			ResourceType: "document",
			ResourceID:   "doc-a",
			PayloadJSON:  "{bad json",
			TraceID:      "trace-a",
			TenantID:     "tenant-a",
		}},
	})
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events", "tenant-a")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "internal.unknown" {
		t.Fatalf("expected code INTERNAL_ERROR, got %q", body.Code)
	}
}

// TestAuditHandler_ExactPageOmitsNextCursor is the B1 handler-boundary guard:
// when the reader reports hasMore=false (exact-multiple last page), the page
// envelope must carry has_more=false and a null next_cursor.
func TestAuditHandler_ExactPageOmitsNextCursor(t *testing.T) {
	t.Parallel()

	handler := httpdelivery.NewHandler(fakeAuditQuerier{
		items: []domain.Event{{
			ID:           "evt-a",
			OccurredAt:   time.Now().UTC(),
			ActorID:      "actor-a",
			Action:       "audit.test",
			ResourceType: "document",
			ResourceID:   "doc-a",
			PayloadJSON:  `{}`,
			TenantID:     "tenant-a",
		}},
		hasMore: false,
	})
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events", "tenant-a")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Page struct {
			NextCursor *string `json:"next_cursor"`
			HasMore    bool    `json:"has_more"`
		} `json:"page"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Page.HasMore {
		t.Fatalf("has_more = true on exact page, want false")
	}
	if body.Page.NextCursor != nil {
		t.Fatalf("next_cursor = %q on exact page, want null", *body.Page.NextCursor)
	}
}

// TestAuditHandler_HasMoreEmitsNextCursor is the B1 handler-boundary guard for
// the next-page case: hasMore=true yields has_more=true and a non-empty,
// URL-safe next_cursor.
func TestAuditHandler_HasMoreEmitsNextCursor(t *testing.T) {
	t.Parallel()

	handler := httpdelivery.NewHandler(fakeAuditQuerier{
		items: []domain.Event{{
			ID:           "evt-a",
			OccurredAt:   time.Now().UTC(),
			ActorID:      "actor-a",
			Action:       "audit.test",
			ResourceType: "document",
			ResourceID:   "doc-a",
			PayloadJSON:  `{}`,
			TenantID:     "tenant-a",
		}},
		hasMore: true,
	})
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events", "tenant-a")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Page struct {
			NextCursor *string `json:"next_cursor"`
			HasMore    bool    `json:"has_more"`
		} `json:"page"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Page.HasMore {
		t.Fatalf("has_more = false with probe row, want true")
	}
	if body.Page.NextCursor == nil || *body.Page.NextCursor == "" {
		t.Fatalf("next_cursor empty with has_more=true, want a cursor")
	}
	if strings.ContainsAny(*body.Page.NextCursor, "+/=") {
		t.Fatalf("next_cursor %q is not URL-safe (has +, / or =)", *body.Page.NextCursor)
	}
}

// TestAuditHandler_LimitClampedToMax is the B3 regression guard: a limit above
// the design-system max (100) is clamped to 100 before reaching the reader.
func TestAuditHandler_LimitClampedToMax(t *testing.T) {
	t.Parallel()

	capture := &queryCaptureQuerier{}
	handler := httpdelivery.NewHandler(capture)
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events?limit=500", "tenant-a")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if capture.query.Limit != 100 {
		t.Fatalf("clamped limit = %d, want 100", capture.query.Limit)
	}
}

type fakeAuditQuerier struct {
	items   []domain.Event
	hasMore bool
	err     error
}

func (f fakeAuditQuerier) ListEvents(context.Context, domain.ListEventsQuery) ([]domain.Event, bool, error) {
	return f.items, f.hasMore, f.err
}

// queryCaptureQuerier records the query it receives so a test can assert the
// handler's parse/clamp behavior.
type queryCaptureQuerier struct {
	query domain.ListEventsQuery
}

func (q *queryCaptureQuerier) ListEvents(_ context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error) {
	q.query = query
	return nil, false, nil
}

func authenticatedAuditRequest(method, target, tenantID string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	return req.WithContext(ctx)
}
