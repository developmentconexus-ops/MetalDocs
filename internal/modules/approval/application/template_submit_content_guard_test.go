package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// F-E4-1 — submitting a template version that carries no committed
// content_hash used to flip the row to under_review and only then hit
// templates_template_version's chk_template_version_content_hash_non_draft
// CHECK, aborting the tx with a raw 23514 that surfaced as a 500 with the
// constraint name in the problem title. These tests pin the friendly first
// line: the guard rejects the submit with ErrTemplateVersionNoContent BEFORE
// MarkTemplateVersionUnderReview is ever attempted.

// stubTemplateVersionReader is a fully in-memory TemplateVersionReader double
// (no SQL) so the guard's ordering can be asserted without a real DB.
type stubTemplateVersionReader struct {
	status      string
	statusOK    bool
	statusErr   error
	contentHash string
	hashOK      bool
	hashErr     error

	hashReads int
}

func (r *stubTemplateVersionReader) LoadTemplateVersionStatus(context.Context, db.Tx, string, string) (string, bool, error) {
	return r.status, r.statusOK, r.statusErr
}

func (r *stubTemplateVersionReader) LoadTemplateVersionContentHash(context.Context, db.Tx, string, string) (string, bool, error) {
	r.hashReads++
	return r.contentHash, r.hashOK, r.hashErr
}

func (r *stubTemplateVersionReader) LoadTemplateInboxMeta(context.Context, db.Tx, string, []string) (map[string]domain.TemplateInboxMeta, error) {
	return nil, nil
}

// LoadTemplateDocTypeCode returns the ADR 0086 template ROUTE subject key. The
// content guard runs before any route lookup, so the value is inert here.
func (r *stubTemplateVersionReader) LoadTemplateDocTypeCode(context.Context, db.Tx, string, string) (string, error) {
	return "po", nil
}

// recordingSubmitWriter records whether the submit-lock was reached.
type recordingSubmitWriter struct {
	calls int
	err   error
}

func (w *recordingSubmitWriter) MarkTemplateVersionUnderReview(context.Context, db.Tx, string, string) error {
	w.calls++
	return w.err
}

func newContentGuardService(t *testing.T, reader TemplateVersionReader, writer TemplateVersionSubmitWriter) (*TemplateSubmitService, db.TxRunner) {
	t.Helper()
	svc := NewTemplateSubmitService(&fakeSubmitRepo{}, &MemoryEmitter{}, fixedClock{t: time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)}, reader, nil)
	svc.versionWriter = writer
	return svc, newTxRunner(newSubmitTestDB(t, true))
}

func contentGuardRequest() TemplateSubmitRequest {
	return TemplateSubmitRequest{
		TenantID:          tenant.DevTenantID,
		TemplateID:        "tmpl-1",
		TemplateVersionID: "tmpl-version-1",
		SubmittedBy:       "user-1",
		IdempotencyKey:    "44444444-4444-4444-4444-444444444444",
	}
}

// TestSubmitTemplateVersionForReview_NoContentHash_RejectedBeforeSubmitLock is
// the F-E4-1 regression pin.
func TestSubmitTemplateVersionForReview_NoContentHash_RejectedBeforeSubmitLock(t *testing.T) {
	reader := &stubTemplateVersionReader{status: "draft", statusOK: true, hashOK: false}
	writer := &recordingSubmitWriter{}
	svc, runner := newContentGuardService(t, reader, writer)

	_, err := svc.SubmitTemplateVersionForReview(authzCtx(tenant.DevTenantID, "user-1"), runner, contentGuardRequest())

	if !errors.Is(err, ErrTemplateVersionNoContent) {
		t.Fatalf("err = %v, want ErrTemplateVersionNoContent", err)
	}
	if reader.hashReads != 1 {
		t.Fatalf("LoadTemplateVersionContentHash called %d times, want 1 (same tx as the status read, no TOCTOU re-read)", reader.hashReads)
	}
	if writer.calls != 0 {
		t.Fatalf("MarkTemplateVersionUnderReview called %d times, want 0 — the guard must fail closed BEFORE the under_review transition", writer.calls)
	}
}

// TestSubmitTemplateVersionForReview_ContentHashPresent_ReachesSubmitLock is
// the positive control: a version WITH a committed hash is not blocked by the
// new guard and does reach MarkTemplateVersionUnderReview. The writer returns
// a sentinel so the assertion does not depend on any downstream step.
func TestSubmitTemplateVersionForReview_ContentHashPresent_ReachesSubmitLock(t *testing.T) {
	sentinel := errors.New("submit-lock reached")
	reader := &stubTemplateVersionReader{
		status:      "draft",
		statusOK:    true,
		contentHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		hashOK:      true,
	}
	writer := &recordingSubmitWriter{err: sentinel}
	svc, runner := newContentGuardService(t, reader, writer)

	_, err := svc.SubmitTemplateVersionForReview(authzCtx(tenant.DevTenantID, "user-1"), runner, contentGuardRequest())

	if errors.Is(err, ErrTemplateVersionNoContent) {
		t.Fatalf("a version with a committed content hash must not be rejected by the F-E4-1 guard, got %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want the submit-lock sentinel (proving the guard let the submit through)", err)
	}
	if writer.calls != 1 {
		t.Fatalf("MarkTemplateVersionUnderReview called %d times, want 1", writer.calls)
	}
}
