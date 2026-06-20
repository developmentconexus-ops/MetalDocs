package httpdelivery

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	authdomain "metaldocs/internal/modules/auth/domain"
)

// jsonKeys marshals v and returns its top-level JSON object keys, sorted. It is the
// F7.2 wire-contract probe: auth is pre-codegen, so the hand-rolled response structs
// must serialize to exactly the OpenAPI-declared key set.
func jsonKeys(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// TestAuthLoginResponse_WireContract locks authLoginResponse to the OpenAPI
// AuthLoginResponse schema (required [user, expires_at]).
func TestAuthLoginResponse_WireContract(t *testing.T) {
	resp := authLoginResponse{
		User:      authdomain.CurrentUser{UserID: "u-1", TenantID: "t-1", Username: "alice"},
		ExpiresAt: "2026-06-20T12:00:00Z",
	}
	if got, want := jsonKeys(t, resp), "expires_at,user"; got != want {
		t.Fatalf("login response keys = %q, want %q (OpenAPI AuthLoginResponse)", got, want)
	}

	raw, _ := json.Marshal(resp)
	var probe struct {
		User struct {
			UserID string `json:"user_id"`
		} `json:"user"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if probe.User.UserID != "u-1" {
		t.Fatalf("user.user_id = %q, want u-1", probe.User.UserID)
	}
	if probe.ExpiresAt != "2026-06-20T12:00:00Z" {
		t.Fatalf("expires_at = %q, want preserved RFC3339 string", probe.ExpiresAt)
	}
}

// TestChangePasswordResponse_WireContract locks changePasswordResponse to the OpenAPI
// ChangePasswordResponse schema (required [changed, user]).
func TestChangePasswordResponse_WireContract(t *testing.T) {
	resp := changePasswordResponse{
		Changed: true,
		User:    authdomain.CurrentUser{UserID: "u-1"},
	}
	if got, want := jsonKeys(t, resp), "changed,user"; got != want {
		t.Fatalf("change-password response keys = %q, want %q (OpenAPI ChangePasswordResponse)", got, want)
	}

	raw, _ := json.Marshal(resp)
	var probe struct {
		Changed bool `json:"changed"`
		User    struct {
			UserID string `json:"user_id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !probe.Changed || probe.User.UserID != "u-1" {
		t.Fatalf("change-password values wrong: %+v", probe)
	}
}
