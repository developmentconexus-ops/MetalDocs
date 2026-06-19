package infrastructure

import (
	"strings"
	"testing"

	templatesdomain "metaldocs/internal/modules/templates/domain"
)

// TestIsPublishedComparesAgainstDomainConstant pins the wire value of the
// published template-version state. IsPublished compares the scanned SQL status
// column against this constant; if the constant's wire string ever drifts from
// "published" the comparison would silently desync. Regression guard, not a
// behavioral red-green (the F5.1 swap is behavior-preserving).
func TestIsPublishedComparesAgainstDomainConstant(t *testing.T) {
	if got := string(templatesdomain.VersionStatusPublished); got != "published" {
		t.Fatalf("VersionStatusPublished wire value = %q, want %q", got, "published")
	}
}

func TestNewTemplateVersionReader_PanicsWhenDBNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = NewTemplateVersionReader(nil)
}

func TestTemplateVersionQueryUsesDocumentTypeCode(t *testing.T) {
	if !strings.Contains(templateVersionQuery, "t.doc_type_code") {
		t.Fatalf("templateVersionQuery should select t.doc_type_code, got: %s", templateVersionQuery)
	}
	if !strings.Contains(templateVersionQuery, "t.tenant_id = $2::uuid") {
		t.Fatalf("templateVersionQuery should scope by tenant_id, got: %s", templateVersionQuery)
	}
	if strings.Contains(templateVersionQuery, "t.profile_code") {
		t.Fatalf("templateVersionQuery should not select t.profile_code, got: %s", templateVersionQuery)
	}
}
