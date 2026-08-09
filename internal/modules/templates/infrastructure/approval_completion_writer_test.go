package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// recordingTx captures the UPDATE the completion writer issues without needing
// a real database. Only ExecContext is exercised by MarkTemplateVersionApproved.
type recordingTx struct {
	query string
	args  []any
	calls int
}

func (tx *recordingTx) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	tx.calls++
	tx.query = query
	tx.args = args
	return oneRowResult{}, nil
}

func (tx *recordingTx) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("unused")
}

func (tx *recordingTx) QueryRowContext(context.Context, string, ...any) *sql.Row {
	panic("unused")
}

type oneRowResult struct{}

func (oneRowResult) LastInsertId() (int64, error) { return 0, nil }
func (oneRowResult) RowsAffected() (int64, error) { return 1, nil }

var _ db.Tx = (*recordingTx)(nil)

// TestMarkTemplateVersionUnderReview_CASCarriesContentHashPrecondition pins the
// TOCTOU fix: the committed-content precondition must live IN the CAS WHERE
// clause, not only in the caller's earlier read. Without it a concurrent edit
// that clears content_hash between check and CAS flips the row to under_review
// and trips chk_template_version_content_hash_non_draft as a raw 23514/500.
func TestMarkTemplateVersionUnderReview_CASCarriesContentHashPrecondition(t *testing.T) {
	tx := &recordingTx{}
	w := NewApprovalCompletionWriter()

	if err := w.MarkTemplateVersionUnderReview(context.Background(), tx, "tenant-1", "version-1"); err != nil {
		t.Fatalf("MarkTemplateVersionUnderReview: %v", err)
	}
	if tx.calls != 1 {
		t.Fatalf("ExecContext calls = %d, want 1", tx.calls)
	}
	for _, want := range []string{"status       = 'draft'", "content_hash IS NOT NULL", "content_hash <> ''"} {
		if !strings.Contains(tx.query, want) {
			t.Fatalf("CAS is missing %q — transition and precondition must be atomic: %s", want, tx.query)
		}
	}
}

// TestMarkTemplateVersionUnderReview_ZeroRowCASDisambiguates is the other half
// of the fix: with two preconditions in the CAS, "no rows" no longer identifies
// which one failed, so the writer re-reads in the SAME tx. A still-draft row
// with an empty hash (the lost race) must surface approval's
// ErrTemplateVersionNoContent — never a constraint error, never a generic
// stale-status conflict; any other cause keeps ErrTemplateVersionNotDraft.
func TestMarkTemplateVersionUnderReview_ZeroRowCASDisambiguates(t *testing.T) {
	cases := map[string]struct {
		status      string
		contentHash string
		wantErr     error
	}{
		"draft_empty_hash_is_no_content":   {status: "draft", contentHash: "", wantErr: approvaldomain.ErrTemplateVersionNoContent},
		"non_draft_stays_stale_conflict":   {status: "under_review", contentHash: strings.Repeat("a", 64), wantErr: ErrTemplateVersionNotDraft},
		"approved_stays_stale_conflict":    {status: "approved", contentHash: strings.Repeat("b", 64), wantErr: ErrTemplateVersionNotDraft},
		"draft_with_hash_stays_stale_conf": {status: "draft", contentHash: strings.Repeat("c", 64), wantErr: ErrTemplateVersionNotDraft},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer func() { _ = mockDB.Close() }()

			mock.ExpectBegin()
			mock.ExpectExec("UPDATE templates_template_version").
				WithArgs("version-1", "tenant-1").
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery("SELECT status, content_hash").
				WithArgs("version-1", "tenant-1").
				WillReturnRows(sqlmock.NewRows([]string{"status", "content_hash"}).AddRow(tc.status, tc.contentHash))

			tx, err := mockDB.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}

			gotErr := NewApprovalCompletionWriter().MarkTemplateVersionUnderReview(context.Background(), tx, "tenant-1", "version-1")
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("err = %v, want %v", gotErr, tc.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations (the disambiguating re-read must run in the same tx): %v", err)
			}
		})
	}
}

// TestMarkTemplateVersionUnderReview_ZeroRowCASAbsentRow keeps the pre-existing
// not-found shape: no row for this tenant+id is a stale/conflict error, not a
// content error.
func TestMarkTemplateVersionUnderReview_ZeroRowCASAbsentRow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = mockDB.Close() }()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE templates_template_version").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT status, content_hash").
		WillReturnRows(sqlmock.NewRows([]string{"status", "content_hash"}))

	tx, err := mockDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	gotErr := NewApprovalCompletionWriter().MarkTemplateVersionUnderReview(context.Background(), tx, "tenant-1", "version-1")
	if !errors.Is(gotErr, ErrTemplateVersionNotDraft) {
		t.Fatalf("err = %v, want ErrTemplateVersionNotDraft", gotErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestMarkTemplateVersionApproved_StampsApproverID is the F-E4-4 pin: the
// deciding actor's user id must be bound into the UPDATE's approver_id, so
// approved_at never lands without attribution.
func TestMarkTemplateVersionApproved_StampsApproverID(t *testing.T) {
	tx := &recordingTx{}
	w := NewApprovalCompletionWriter()

	if err := w.MarkTemplateVersionApproved(context.Background(), tx, "tenant-1", "version-1", "user-approver"); err != nil {
		t.Fatalf("MarkTemplateVersionApproved: %v", err)
	}
	if tx.calls != 1 {
		t.Fatalf("ExecContext calls = %d, want 1", tx.calls)
	}
	if !strings.Contains(tx.query, "approver_id = $3") {
		t.Fatalf("UPDATE does not set approver_id: %s", tx.query)
	}
	if len(tx.args) != 3 || tx.args[2] != "user-approver" {
		t.Fatalf("args = %v, want the deciding actor id bound as $3", tx.args)
	}
}

// TestMarkTemplateVersionApproved_EmptyApproverFailsClosed pins the
// no-fallback rule: an absent deciding actor is a wiring fault to surface, not
// a NULL or placeholder to persist. The UPDATE must never be issued.
func TestMarkTemplateVersionApproved_EmptyApproverFailsClosed(t *testing.T) {
	for name, approver := range map[string]string{"empty": "", "whitespace": "   "} {
		t.Run(name, func(t *testing.T) {
			tx := &recordingTx{}
			w := NewApprovalCompletionWriter()

			err := w.MarkTemplateVersionApproved(context.Background(), tx, "tenant-1", "version-1", approver)
			if !errors.Is(err, ErrApproverUserIDRequired) {
				t.Fatalf("err = %v, want ErrApproverUserIDRequired", err)
			}
			if tx.calls != 0 {
				t.Fatalf("ExecContext calls = %d, want 0 (fail closed before the UPDATE)", tx.calls)
			}
		})
	}
}
