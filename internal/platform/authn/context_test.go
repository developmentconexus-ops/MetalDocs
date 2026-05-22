package authn

import (
	"context"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestUserIDFromContext_EmptyContextReportsAbsent(t *testing.T) {
	got, ok := UserIDFromContext(context.Background())
	if ok {
		t.Fatalf("ok = true, want false for empty context")
	}
	if got != "" {
		t.Fatalf("id = %q, want \"\"", got)
	}
}

func TestUserIDFromContext_NilContextReportsAbsent(t *testing.T) {
	got, ok := UserIDFromContext(nil)
	if ok {
		t.Fatalf("ok = true, want false for nil context")
	}
	if got != "" {
		t.Fatalf("id = %q, want \"\"", got)
	}
}

func TestUserIDFromContext_PopulatedContextReportsPresent(t *testing.T) {
	ctx := iamdomain.WithAuthContext(context.Background(), "actor-7", nil)
	got, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatalf("ok = false, want true for populated context")
	}
	if got != "actor-7" {
		t.Fatalf("id = %q, want \"actor-7\"", got)
	}
}

func TestUserIDFromContext_WhitespaceOnlyTreatedAsAbsent(t *testing.T) {
	ctx := iamdomain.WithAuthContext(context.Background(), "   ", nil)
	got, ok := UserIDFromContext(ctx)
	if ok {
		t.Fatalf("ok = true for whitespace-only id, want false")
	}
	if got != "" {
		t.Fatalf("id = %q, want trimmed empty string", got)
	}
}

func TestUserIDFromContext_TrimsSurroundingWhitespace(t *testing.T) {
	ctx := iamdomain.WithAuthContext(context.Background(), "  actor-9  ", nil)
	got, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatalf("ok = false, want true after trim")
	}
	if got != "actor-9" {
		t.Fatalf("id = %q, want \"actor-9\"", got)
	}
}
