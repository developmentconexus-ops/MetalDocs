package application

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"
)

// buildTestDocx builds a minimal in-memory zip with a single word/document.xml
// entry whose body is the caller-supplied fragment, wrapped in a minimal valid
// OOXML document envelope. Real archive/zip construction (no external
// fixtures), per plan.md's TDD requirement.
func buildTestDocx(t *testing.T, bodyXML string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body>` + bodyXML + `</w:body></w:document>`
	if _, err := w.Write([]byte(doc)); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func buildDocxMissingDocumentXML(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/styles.xml")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := w.Write([]byte(`<w:styles/>`)); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestScanForUnresolvedMarkup_Clean(t *testing.T) {
	docx := buildTestDocx(t, `<w:p><w:r><w:t>Hello, clean document.</w:t></w:r></w:p>`)
	if err := ScanForUnresolvedMarkup(docx); err != nil {
		t.Fatalf("expected nil error for clean docx, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_TrackedInsertion(t *testing.T) {
	docx := buildTestDocx(t, `<w:ins w:id="1" w:author="reviewer"><w:r><w:t>added</w:t></w:r></w:ins>`)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrUnresolvedTrackedChanges) {
		t.Fatalf("expected ErrUnresolvedTrackedChanges for w:ins, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_TrackedDeletion(t *testing.T) {
	docx := buildTestDocx(t, `<w:del w:id="2" w:author="reviewer"><w:r><w:delText>removed</w:delText></w:r></w:del>`)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrUnresolvedTrackedChanges) {
		t.Fatalf("expected ErrUnresolvedTrackedChanges for w:del, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_ParagraphPropertyChange(t *testing.T) {
	docx := buildTestDocx(t, `<w:p><w:pPr><w:pPrChange w:id="3"/></w:pPr></w:p>`)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrUnresolvedTrackedChanges) {
		t.Fatalf("expected ErrUnresolvedTrackedChanges for w:pPrChange, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_CommentRangeStart(t *testing.T) {
	docx := buildTestDocx(t, `<w:commentRangeStart w:id="1"/><w:p><w:r><w:t>text</w:t></w:r></w:p>`)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrUnstrippedComments) {
		t.Fatalf("expected ErrUnstrippedComments for w:commentRangeStart, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_CommentReference(t *testing.T) {
	docx := buildTestDocx(t, `<w:p><w:r><w:commentReference w:id="1"/></w:r></w:p>`)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrUnstrippedComments) {
		t.Fatalf("expected ErrUnstrippedComments for w:commentReference, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_MissingDocumentXML(t *testing.T) {
	docx := buildDocxMissingDocumentXML(t)
	err := ScanForUnresolvedMarkup(docx)
	if !errors.Is(err, ErrInvalidDocxBuffer) {
		t.Fatalf("expected ErrInvalidDocxBuffer for missing word/document.xml, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_CorruptBuffer(t *testing.T) {
	err := ScanForUnresolvedMarkup([]byte("not a zip file at all"))
	if !errors.Is(err, ErrInvalidDocxBuffer) {
		t.Fatalf("expected ErrInvalidDocxBuffer for corrupt buffer, got %v", err)
	}
}

func TestScanForUnresolvedMarkup_EmptyBuffer(t *testing.T) {
	err := ScanForUnresolvedMarkup(nil)
	if !errors.Is(err, ErrInvalidDocxBuffer) {
		t.Fatalf("expected ErrInvalidDocxBuffer for empty buffer, got %v", err)
	}
}
