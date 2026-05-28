package httpdelivery

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

type stubRoleAdminRepository struct{}

func (stubRoleAdminRepository) HasAnyRole(context.Context, iamdomain.Role, string) (bool, error) {
	return false, nil
}

func (stubRoleAdminRepository) UpsertUserAndAssignRole(context.Context, string, string, string, iamdomain.Role, string) error {
	return nil
}

func (stubRoleAdminRepository) ReplaceUserRoles(context.Context, string, string, string, iamdomain.Role, string) error {
	return nil
}

type stubUserAdminService struct {
	listUsersTenantID       string
	listOnlineUsersTenantID string
}

func (s *stubUserAdminService) ListUsers(context.Context, string) ([]authdomain.ManagedUser, error) {
	return nil, nil
}

func (s *stubUserAdminService) ListOnlineUsers(_ context.Context, tenantID string, _ time.Time) ([]authdomain.OnlineUser, error) {
	s.listOnlineUsersTenantID = tenantID
	return nil, nil
}

func (s *stubUserAdminService) CreateUser(context.Context, string, string, string, string, string, string, []iamdomain.Role, string) error {
	return nil
}

func (s *stubUserAdminService) UpdateUser(context.Context, authdomain.UpdateUserParams, string) error {
	return nil
}

func (s *stubUserAdminService) AdminResetPassword(context.Context, string, string) error {
	return nil
}

func (s *stubUserAdminService) UnlockUser(context.Context, string) error {
	return nil
}

func TestHandleReplaceUserRoles_RejectsMultipleRoles(t *testing.T) {
	handler := NewAdminHandler(iamapp.NewAdminService(stubRoleAdminRepository{}, nil), nil)

	body := bytes.NewBufferString(`{"displayName":"Alice","roles":["editor","viewer"]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/iam/users/alice/roles", body)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	handler.handleReplaceUserRoles(rec, req, "alice")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte("Exactly one role is required")) {
		t.Fatalf("response body %q does not mention single-role requirement", got)
	}
}

func TestHandleCreateUser_RejectsMultipleRoles(t *testing.T) {
	authSvc := &stubUserAdminService{}
	handler := NewAdminHandler(iamapp.NewAdminService(stubRoleAdminRepository{}, nil), authSvc)

	payload, err := json.Marshal(CreateUserRequest{
		Username:    "alice",
		DisplayName: "Alice",
		Password:    "Password123!",
		Roles:       []string{"editor", "viewer"},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/iam/users", bytes.NewReader(payload))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	handler.handleCreateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte("Exactly one role is required")) {
		t.Fatalf("response body %q does not mention single-role requirement", got)
	}
}

func TestHandleAdminOverview_PassesTenantIDToOnlineUsers(t *testing.T) {
	authSvc := &stubUserAdminService{}
	handler := NewAdminHandler(nil, authSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/admin/overview", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	handler.handleAdminOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if authSvc.listOnlineUsersTenantID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("ListOnlineUsers tenant = %q", authSvc.listOnlineUsersTenantID)
	}
}
