package infrastructure

import (
	"strings"
	"testing"
)

func TestNewTemplateVersionChecker_PanicsWhenDBNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = NewTemplateVersionChecker(nil)
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
