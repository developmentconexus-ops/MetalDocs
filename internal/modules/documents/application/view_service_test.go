package application

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
)

// fakePDFOutboxStateReader is a spy PDFOutboxStateReader (view_service.go:27-29)
// used to prove GetViewURL only consults the reader on the "no pdf key yet"
// branch, and to drive the failed/pending outcomes independently of any real
// outbox_events row.
type fakePDFOutboxStateReader struct {
	calls int
	state string
	err   error
}

func (f *fakePDFOutboxStateReader) ReadState(_ context.Context, _, _ string) (string, error) {
	f.calls++
	return f.state, f.err
}

type fakeViewPresigner struct {
	calls int
	url   string
	err   error
}

func (f *fakeViewPresigner) AssertedPresignGet(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.url, nil
}

// viewAuthzCtx builds a context carrying the tenant+actor identity the
// TxRunner chokepoint (internal/platform/db/runner.go, M3 F3.1) auto-seeds
// from ctx on every Do call, matching the identity expectViewAuthz's
// mocked set_config exec expects.
func viewAuthzCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	return platformtenant.WithActorID(ctx, actorID)
}

// expectViewIdentitySeed wires the sqlmock expectation for the TxRunner
// chokepoint auto-seed (internal/platform/db/runner.go seedTxIdentityFromContext),
// which runs immediately after BeginTx and before any application-level query.
func expectViewIdentitySeed(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectExec(regexp.QuoteMeta(`
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`)).
		WithArgs(tenantID, actorID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// expectViewAuthz wires the sqlmock expectations for the full authz.Require
// tier-2 check GetViewURL performs for CapDocumentView ("document.view") in
// the tenant-grade scope, taking the system_admin bypass branch (recordBypass
// is a no-op with no bypass-audit sink configured in these unit tests).
func expectViewAuthz(mock sqlmock.Sqlmock, actorID, tenantID string) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(actorID))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(tenantID))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
)`)).
		WithArgs(actorID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps', $1, true)")).
		WithArgs(`[{"area":"tenant","cap":"document.view"}]`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestViewService_GetViewURL_ReadyWhenPDFKeyPresent(t *testing.T) {
	tenantID := "tenant-a"
	actorID := "user-a"
	docID := "doc-1"

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })

	presigner := &fakeViewPresigner{url: "https://signed.example/doc-1.pdf"}
	outbox := &fakePDFOutboxStateReader{state: "failed"} // must NOT be consulted
	svc := NewViewService(db.NewTxRunner(mockDB), presigner, outbox)

	mock.ExpectBegin()
	expectViewIdentitySeed(mock, actorID, tenantID)
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT status, final_pdf_s3_key
			  FROM documents
			 WHERE tenant_id=$1::uuid AND id=$2::uuid`)).
		WithArgs(tenantID, docID).
		WillReturnRows(sqlmock.NewRows([]string{"status", "final_pdf_s3_key"}).
			AddRow("approved", "tenants/tenant-a/docs/doc-1/final.pdf"))
	expectViewAuthz(mock, actorID, tenantID)
	mock.ExpectCommit()

	result, err := svc.GetViewURL(viewAuthzCtx(tenantID, actorID), tenantID, actorID, docID)
	if err != nil {
		t.Fatalf("GetViewURL returned error: %v", err)
	}
	if result.PDFStatus != "ready" {
		t.Fatalf("PDFStatus = %q, want %q", result.PDFStatus, "ready")
	}
	if result.SignedURL != presigner.url {
		t.Fatalf("SignedURL = %q, want %q", result.SignedURL, presigner.url)
	}
	if presigner.calls != 1 {
		t.Fatalf("presigner calls = %d, want 1", presigner.calls)
	}
	if outbox.calls != 0 {
		t.Fatalf("outbox reader must not be consulted when a pdf key is present, got %d call(s)", outbox.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestViewService_GetViewURL_FailedWhenOutboxReaderReportsFailed(t *testing.T) {
	tenantID := "tenant-a"
	actorID := "user-a"
	docID := "doc-1"

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })

	presigner := &fakeViewPresigner{}
	outbox := &fakePDFOutboxStateReader{state: "failed"}
	svc := NewViewService(db.NewTxRunner(mockDB), presigner, outbox)

	mock.ExpectBegin()
	expectViewIdentitySeed(mock, actorID, tenantID)
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT status, final_pdf_s3_key
			  FROM documents
			 WHERE tenant_id=$1::uuid AND id=$2::uuid`)).
		WithArgs(tenantID, docID).
		WillReturnRows(sqlmock.NewRows([]string{"status", "final_pdf_s3_key"}).
			AddRow("approved", nil))
	expectViewAuthz(mock, actorID, tenantID)
	mock.ExpectCommit()

	result, err := svc.GetViewURL(viewAuthzCtx(tenantID, actorID), tenantID, actorID, docID)
	if err != nil {
		t.Fatalf("GetViewURL returned error: %v", err)
	}
	if result.PDFStatus != "failed" {
		t.Fatalf("PDFStatus = %q, want %q", result.PDFStatus, "failed")
	}
	if result.SignedURL != "" {
		t.Fatalf("SignedURL = %q, want empty", result.SignedURL)
	}
	if presigner.calls != 0 {
		t.Fatalf("presigner must not be called when no pdf key is present, got %d call(s)", presigner.calls)
	}
	if outbox.calls != 1 {
		t.Fatalf("outbox reader calls = %d, want 1", outbox.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestViewService_GetViewURL_PendingWhenOutboxReaderReportsNoState(t *testing.T) {
	tenantID := "tenant-a"
	actorID := "user-a"
	docID := "doc-1"

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })

	presigner := &fakeViewPresigner{}
	outbox := &fakePDFOutboxStateReader{state: ""}
	svc := NewViewService(db.NewTxRunner(mockDB), presigner, outbox)

	mock.ExpectBegin()
	expectViewIdentitySeed(mock, actorID, tenantID)
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT status, final_pdf_s3_key
			  FROM documents
			 WHERE tenant_id=$1::uuid AND id=$2::uuid`)).
		WithArgs(tenantID, docID).
		WillReturnRows(sqlmock.NewRows([]string{"status", "final_pdf_s3_key"}).
			AddRow("approved", nil))
	expectViewAuthz(mock, actorID, tenantID)
	mock.ExpectCommit()

	result, err := svc.GetViewURL(viewAuthzCtx(tenantID, actorID), tenantID, actorID, docID)
	if err != nil {
		t.Fatalf("GetViewURL returned error: %v", err)
	}
	if result.PDFStatus != "pending" {
		t.Fatalf("PDFStatus = %q, want %q", result.PDFStatus, "pending")
	}
	if outbox.calls != 1 {
		t.Fatalf("outbox reader calls = %d, want 1", outbox.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
