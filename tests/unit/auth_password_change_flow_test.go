package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	httpdelivery "metaldocs/internal/modules/auth/delivery/http"
	authmemory "metaldocs/internal/modules/auth/infrastructure/memory"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

func TestPasswordChangeRevokesSessionAndClearsMustChangePassword(t *testing.T) {
	repo := authmemory.NewRepository()
	cfg := authapp.Config{
		SessionCookieName:      "metaldocs_session",
		SessionTTL:             time.Hour,
		SessionSecret:          "0123456789abcdef0123456789abcdef",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      5 * time.Minute,
		AllowDevTenantFallback: true,
	}
	svc, err := authapp.NewService(repo, repo, repo, noopLoginCtxPort{}, cfg)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	if err := svc.CreateUser(context.Background(), "flow-user", "flow.user", "flow.user@test.local", "Flow User", "abc12345", tenant.DevTenantID, []iamdomain.Role{iamdomain.RoleViewer}, "test"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	authHandler := httpdelivery.NewHandler(svc)
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	handler := httpdelivery.NewMiddleware(svc, cfg, true).Wrap(mux)

	loginResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/login", `{"identifier":"flow.user","password":"abc12345"}`, nil)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected initial login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	loginPayload := decodeMap(t, loginResp.Body.String())
	userPayload := loginPayload["user"].(map[string]any)
	if userPayload["must_change_password"] != true {
		t.Fatalf("expected must_change_password=true on first login, got %#v", userPayload["must_change_password"])
	}

	sessionCookie := findCookie(t, loginResp.Result().Cookies(), cfg.SessionCookieName)
	changeResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/change-password", `{"new_password":"abc12346"}`, sessionCookie)
	if changeResp.Code != http.StatusOK {
		t.Fatalf("expected change password 200, got %d body=%s", changeResp.Code, changeResp.Body.String())
	}

	changePayload := decodeMap(t, changeResp.Body.String())
	rotatedUser := changePayload["user"].(map[string]any)
	if rotatedUser["must_change_password"] != false {
		t.Fatalf("expected must_change_password=false after password change, got %#v", rotatedUser["must_change_password"])
	}

	// A3 (commit 371e2fcea): self-service password change revokes ALL of the
	// user's sessions, including the current one (CWE-613). The cookie that was
	// valid before the change must now be rejected; the client re-authenticates
	// with the new credential.
	meResp := performJSONRequest(t, handler, http.MethodGet, "/api/v1/auth/me", "", sessionCookie)
	if meResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected existing session to be revoked after password change, got %d body=%s", meResp.Code, meResp.Body.String())
	}

	oldLoginResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/login", `{"identifier":"flow.user","password":"abc12345"}`, nil)
	if oldLoginResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password login to fail with 401, got %d body=%s", oldLoginResp.Code, oldLoginResp.Body.String())
	}

	newLoginResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/login", `{"identifier":"flow.user","password":"abc12346"}`, nil)
	if newLoginResp.Code != http.StatusOK {
		t.Fatalf("expected new password login 200, got %d body=%s", newLoginResp.Code, newLoginResp.Body.String())
	}
	newLoginPayload := decodeMap(t, newLoginResp.Body.String())
	newUserPayload := newLoginPayload["user"].(map[string]any)
	if newUserPayload["must_change_password"] != false {
		t.Fatalf("expected must_change_password=false after relogin, got %#v", newUserPayload["must_change_password"])
	}
}

// F0.4: handleChangePassword must emit an expired session cookie on success,
// mirroring handleLogout. Server-side revocation already happens in the service
// tx (CWE-613). This test pins the client-visible cookie clear so the browser
// drops the now-revoked cookie instead of carrying a dead value.
func TestPasswordChangeEmitsExpiredCookie(t *testing.T) {
	repo := authmemory.NewRepository()
	cfg := authapp.Config{
		SessionCookieName:      "metaldocs_session",
		SessionTTL:             time.Hour,
		SessionSecret:          "0123456789abcdef0123456789abcdef",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      5 * time.Minute,
		AllowDevTenantFallback: true,
	}
	svc, err := authapp.NewService(repo, repo, repo, noopLoginCtxPort{}, cfg)
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	if err := svc.CreateUser(context.Background(), "cookie-user", "cookie.user", "cookie.user@test.local", "Cookie User", "abc12345", tenant.DevTenantID, []iamdomain.Role{iamdomain.RoleViewer}, "test"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	authHandler := httpdelivery.NewHandler(svc)
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	handler := httpdelivery.NewMiddleware(svc, cfg, true).Wrap(mux)

	loginResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/login", `{"identifier":"cookie.user","password":"abc12345"}`, nil)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected initial login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	sessionCookie := findCookie(t, loginResp.Result().Cookies(), cfg.SessionCookieName)

	changeResp := performJSONRequest(t, handler, http.MethodPost, "/api/v1/auth/change-password", `{"new_password":"abc12346"}`, sessionCookie)
	if changeResp.Code != http.StatusOK {
		t.Fatalf("expected change password 200, got %d body=%s", changeResp.Code, changeResp.Body.String())
	}

	expired := findCookie(t, changeResp.Result().Cookies(), cfg.SessionCookieName)
	if expired.MaxAge >= 0 {
		t.Fatalf("expected expired session cookie (MaxAge<0), got MaxAge=%d", expired.MaxAge)
	}
	if expired.Value != "" {
		t.Fatalf("expected empty cookie value, got %q", expired.Value)
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %s not found", name)
	return nil
}

func decodeMap(t *testing.T, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode json: %v body=%s", err, body)
	}
	return payload
}
