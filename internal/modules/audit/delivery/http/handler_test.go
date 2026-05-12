package httpdelivery_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	application "metaldocs/internal/modules/audit/application"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/infrastructure/memory"
)

func TestAuditHandler_InvalidLimitIsProblemJSON(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events?limit=invalid", nil)
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
