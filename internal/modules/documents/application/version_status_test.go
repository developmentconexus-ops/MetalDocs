package application_test

import (
	"testing"

	templatesdomain "metaldocs/internal/modules/templates/domain"
)

// TestVersionStatusPublished_WireValue guards the F4.1 fix: the constant
// templatesdomain.VersionStatusPublished must resolve to the wire string
// "published". If templates ever renames the state value, this catches the
// divergence immediately.
func TestVersionStatusPublished_WireValue(t *testing.T) {
	if got := string(templatesdomain.VersionStatusPublished); got != "published" {
		t.Fatalf("templatesdomain.VersionStatusPublished wire value = %q, want %q", got, "published")
	}
}
