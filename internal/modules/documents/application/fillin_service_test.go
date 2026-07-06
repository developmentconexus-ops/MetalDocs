package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	v2domain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	templatesdomain "metaldocs/internal/modules/templates/domain"
)

type fakeSchemaReader struct {
	placeholders []templatesdomain.Placeholder
	err          error
}

func (f fakeSchemaReader) LoadPlaceholderSchema(_ context.Context, _, _ string) ([]templatesdomain.Placeholder, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.placeholders, nil
}

type fakeFillInWriter struct {
	upserts []infrastructure.PlaceholderValue
	err     error
}

func (f *fakeFillInWriter) UpsertValue(_ context.Context, v infrastructure.PlaceholderValue, _ ...infrastructure.DBTX) error {
	if f.err != nil {
		return f.err
	}
	f.upserts = append(f.upserts, v)
	return nil
}

// UpsertAuthorValue satisfies the expanded FillInWriter interface; delegates to
// UpsertValue semantics for test purposes (no governing rows in existing tests).
func (f *fakeFillInWriter) UpsertAuthorValue(_ context.Context, v infrastructure.PlaceholderValue, _ ...infrastructure.DBTX) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.upserts = append(f.upserts, v)
	return 1, nil
}

// CurrentSource satisfies the expanded FillInWriter interface; returns no
// existing row so the guard passes (existing tests don't seed governed rows).
func (f *fakeFillInWriter) CurrentSource(_ context.Context, _, _, _ string) (string, bool, error) {
	return "", false, nil
}

// newPermissiveFillInDB returns a sqlmock *sql.DB that satisfies the
// requireDocEditDraft authz sequence (BeginTx → SeedTxIdentity GUC EXECs →
// LoadDocumentAreaCode 2-arg query → authz.Require GUC reads + EXISTS).
// Mirrors the established newPermissiveMockDB pattern from
// controlleddocuments/delivery/http/routes_contract_test.go.
func newPermissiveFillInDB(t *testing.T) *sql.DB {
	t.Helper()
	// Substring matcher: an empty expected pattern is a wildcard; otherwise the
	// actual SQL must contain the (lowercased) expected token. This lets the
	// ported 2-column LoadDocumentAreaCode read (SELECT process_area_code_snapshot,
	// controlled_document_id FROM documents ...) be routed to a 2-column result,
	// distinct from the 1-column authz EXISTS read that also takes two args.
	matcher := sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		if strings.TrimSpace(expected) == "" {
			return nil
		}
		if strings.Contains(strings.ToLower(actual), strings.ToLower(expected)) {
			return nil
		}
		return fmt.Errorf("actual SQL %q does not contain %q", actual, expected)
	})
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatalf("newPermissiveFillInDB: sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	t.Cleanup(func() { _ = mockDB.Close() })
	for i := 0; i < 20; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	for i := 0; i < 100; i++ {
		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	// Ported B5 area read (declared first so it wins routing for "from documents").
	for i := 0; i < 20; i++ {
		mock.ExpectQuery("from documents").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"process_area_code_snapshot", "controlled_document_id"}).AddRow("QA", nil))
	}
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("exists").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("user-1"))
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("tenant-1"))
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	}
	return mockDB
}

func TestFillInService_ValueRejectedIfFailsRegex(t *testing.T) {
	re := "^[A-Z]{3}$"
	schema := []templatesdomain.Placeholder{{ID: "p1", Type: templatesdomain.PHText, Regex: &re}}
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, &fakeFillInWriter{})
	err := svc.SetPlaceholderValue(context.Background(), "tenant", "actor", "rev", "p1", "abc")
	if !errors.Is(err, v2domain.ErrValidationFailed) {
		t.Fatalf("got %v", err)
	}
}

func TestFillInService_ValueAcceptedIfMatches(t *testing.T) {
	re := "^[A-Z]{3}$"
	schema := []templatesdomain.Placeholder{{ID: "p1", Type: templatesdomain.PHText, Regex: &re}}
	writer := &fakeFillInWriter{}
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, writer)
	if err := svc.SetPlaceholderValue(context.Background(), "tenant", "actor", "rev", "p1", "ABC"); err != nil {
		t.Fatal(err)
	}
	if len(writer.upserts) != 1 || *writer.upserts[0].ValueText != "ABC" {
		t.Fatalf("bad upsert: %+v", writer.upserts)
	}
}

func TestFillInService_SetPlaceholderValue_ValidationMatrix(t *testing.T) {
	maxLen := 3
	minN := 10.0
	maxN := 20.0
	minD := "2026-04-01"
	maxD := "2026-04-30"

	schema := []templatesdomain.Placeholder{
		{ID: "req", Type: templatesdomain.PHText, Required: true},
		{ID: "len", Type: templatesdomain.PHText, MaxLength: &maxLen},
		{ID: "num", Type: templatesdomain.PHNumber, MinNumber: &minN, MaxNumber: &maxN},
		{ID: "date", Type: templatesdomain.PHDate, MinDate: &minD, MaxDate: &maxD},
		{ID: "sel", Type: templatesdomain.PHSelect, Options: []string{"A", "B"}},
	}

	cases := []struct {
		name          string
		placeholderID string
		value         string
		wantErr       bool
	}{
		{name: "required-empty", placeholderID: "req", value: "", wantErr: true},
		{name: "max-length", placeholderID: "len", value: "abcd", wantErr: true},
		{name: "number-too-low", placeholderID: "num", value: "9", wantErr: true},
		{name: "number-too-high", placeholderID: "num", value: "21", wantErr: true},
		{name: "number-ok", placeholderID: "num", value: "11"},
		{name: "date-before-min", placeholderID: "date", value: "2026-03-31", wantErr: true},
		{name: "date-after-max", placeholderID: "date", value: "2026-05-01", wantErr: true},
		{name: "date-ok", placeholderID: "date", value: "2026-04-20"},
		{name: "select-invalid", placeholderID: "sel", value: "X", wantErr: true},
		{name: "select-ok", placeholderID: "sel", value: "A"},
		{name: "unknown", placeholderID: "missing", value: "A", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writer := &fakeFillInWriter{}
			svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, writer)
			err := svc.SetPlaceholderValue(context.Background(), "tenant", "actor", "rev", tc.placeholderID, tc.value)
			if tc.wantErr {
				if !errors.Is(err, v2domain.ErrValidationFailed) {
					t.Fatalf("expected ErrValidationFailed, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(writer.upserts) != 1 {
				t.Fatalf("expected one upsert, got %d", len(writer.upserts))
			}
		})
	}
}

// --- IAM user placeholder validation ---

type fakeIAMOptionsReader struct {
	opts []UserOption
	err  error
}

func (f *fakeIAMOptionsReader) ListUserOptions(_ context.Context, _ string) ([]UserOption, error) {
	return f.opts, f.err
}

// TestFillInService_UserPlaceholder_NoIAMReader_FailClosed asserts the
// fail-closed posture: a PHUser placeholder is REJECTED (not silently accepted)
// when WithIAMReader was never called. Mirrors the nil-area → deny behaviour of
// requireDocEditDraft (fillin_authz.go).
func TestFillInService_UserPlaceholder_NoIAMReader_FailClosed(t *testing.T) {
	schema := []templatesdomain.Placeholder{{ID: "p-user", Type: templatesdomain.PHUser}}
	// No WithIAMReader — iam field stays nil.
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, &fakeFillInWriter{})

	err := svc.SetPlaceholderValue(context.Background(), "t", "actor", "r", "p-user", "any-value")
	if !errors.Is(err, v2domain.ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed (fail-closed), got %v", err)
	}
}

func TestFillInService_UserPlaceholder_KnownUser_Accepted(t *testing.T) {
	schema := []templatesdomain.Placeholder{{ID: "p-user", Type: templatesdomain.PHUser}}
	iam := &fakeIAMOptionsReader{opts: []UserOption{
		{UserID: "u1", DisplayName: "Alice"},
	}}
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, &fakeFillInWriter{}).
		WithIAMReader(iam)

	if err := svc.SetPlaceholderValue(context.Background(), "t", "actor", "r", "p-user", "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFillInService_UserPlaceholder_UnknownUser_Returns422(t *testing.T) {
	schema := []templatesdomain.Placeholder{{ID: "p-user", Type: templatesdomain.PHUser}}
	iam := &fakeIAMOptionsReader{opts: []UserOption{
		{UserID: "u1", DisplayName: "Alice"},
	}}
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{placeholders: schema}, &fakeFillInWriter{}).
		WithIAMReader(iam)

	err := svc.SetPlaceholderValue(context.Background(), "t", "actor", "r", "p-user", "unknown-uid")
	if !errors.Is(err, v2domain.ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
}

func TestFillInService_GetPlaceholderValues_RequiresReader(t *testing.T) {
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{}, &fakeFillInWriter{})

	_, err := svc.GetPlaceholderValues(context.Background(), "tenant", "doc")
	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestFillInService_GetFillInSchema_RequiresSchemaReader(t *testing.T) {
	svc := NewFillInService(newTxRunner(newPermissiveFillInDB(t)), fakeSchemaReader{}, &fakeFillInWriter{})

	_, err := svc.GetFillInSchema(context.Background(), "tenant", "doc")
	if err == nil {
		t.Fatal("expected configuration error")
	}
}
