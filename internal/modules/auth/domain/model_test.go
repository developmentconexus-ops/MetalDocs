package domain

import "testing"

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
