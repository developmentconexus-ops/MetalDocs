package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewIdentity_ZeroValueSafeDefaults(t *testing.T) {
	identity := NewIdentity(UserID(" user-1 "), " user ")
	if identity.UserID != "user-1" {
		t.Fatalf("UserID = %q, want %q", identity.UserID, "user-1")
	}
	if identity.Username != "user" {
		t.Fatalf("Username = %q, want %q", identity.Username, "user")
	}
	if !identity.IsActive {
		t.Fatal("IsActive = false, want true")
	}
	if identity.PasswordHash != "" {
		t.Fatalf("PasswordHash = %q, want empty", identity.PasswordHash)
	}
}

func TestAuthenticatedSession_RedactsRawToken(t *testing.T) {
	session := AuthenticatedSession{
		RawToken:  "secret-token",
		ExpiresAt: time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC),
	}
	if got := session.String(); strings.Contains(got, "secret-token") {
		t.Fatalf("String() leaked token: %s", got)
	}
	raw, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if strings.Contains(string(raw), "secret-token") {
		t.Fatalf("MarshalJSON leaked token: %s", raw)
	}
	if !strings.Contains(string(raw), "***") {
		t.Fatalf("MarshalJSON did not include redaction marker: %s", raw)
	}
}
