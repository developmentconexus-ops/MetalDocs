package application

import (
	"context"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/domain"
)

// TestResolveSubjectAreaCode_Template asserts a template subject resolves to
// the 'tenant' sentinel WITHOUT issuing any documents-table area query — the
// conn's areaCode is deliberately set to a non-"tenant" value so a leaked
// document-area query would be caught by the mismatched return value.
func TestResolveSubjectAreaCode_Template(t *testing.T) {
	db, conn := newSubmitTestDBWithConn(t)
	conn.areaCode = "QA" // would fail the assertion below if the helper mistakenly queried it

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := resolveSubjectAreaCode(context.Background(), tx, nil, conn.tenantID, domain.Subject{Kind: domain.SubjectKindTemplate, Key: "tmpl-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tenant" {
		t.Fatalf("expected 'tenant' sentinel for template subject, got %q", got)
	}
}

// TestResolveSubjectAreaCode_Document asserts a document subject delegates to
// docapp.LoadDocumentAreaCode and returns the derived area, using the shared
// docAreaRows fake — proving zero document-path behavior change.
func TestResolveSubjectAreaCode_Document(t *testing.T) {
	db, conn := newSubmitTestDBWithConn(t)
	conn.areaCode = "QA"

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	got, err := resolveSubjectAreaCode(context.Background(), tx, nil, conn.tenantID, domain.NewDocumentSubject("doc-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "QA" {
		t.Fatalf("expected derived area 'QA', got %q", got)
	}
}

// TestResolveSubjectAreaCode_UnknownKind asserts an unrecognized subject kind
// fails closed with an error rather than silently returning an area string.
func TestResolveSubjectAreaCode_UnknownKind(t *testing.T) {
	db, conn := newSubmitTestDBWithConn(t)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	_, err = resolveSubjectAreaCode(context.Background(), tx, nil, conn.tenantID, domain.Subject{Kind: "bogus", Key: "x"})
	if err == nil {
		t.Fatal("expected error for unknown subject kind, got nil")
	}
	if !strings.Contains(err.Error(), "unknown subject kind") {
		t.Fatalf("expected 'unknown subject kind' error, got: %v", err)
	}
}
