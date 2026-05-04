package application

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/render/resolvers"
)

type stubRevisionReader struct{}

func (stubRevisionReader) GetRevisionNumber(_ context.Context, _, _ string) (int64, error) {
	return 0, nil
}
func (stubRevisionReader) GetEffectiveFrom(_ context.Context, _, _ string) (time.Time, error) {
	return time.Time{}, nil
}
func (stubRevisionReader) GetAuthor(_ context.Context, _, _ string) (resolvers.AuthorInfo, error) {
	return resolvers.AuthorInfo{}, nil
}

type stubWorkflowReader struct{}

func (stubWorkflowReader) GetApprovers(_ context.Context, _, _, _ string) ([]resolvers.ApproverInfo, error) {
	return nil, nil
}
func (stubWorkflowReader) GetFinalApprovalDate(_ context.Context, _, _ string) (time.Time, error) {
	return time.Time{}, nil
}

type stubRegistryReader struct{}

func (stubRegistryReader) GetControlledDocument(_ context.Context, _, _ string) (resolvers.ControlledDocumentInfo, error) {
	return resolvers.ControlledDocumentInfo{}, nil
}

type stubDocumentReader struct{}

func (stubDocumentReader) GetDocumentTitle(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func TestBuild_WiresDocumentReaderAndControlledDocumentID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	tenantID := "00000000-0000-0000-0000-000000000001"
	revisionID := "00000000-0000-0000-0000-000000000002"
	cdID := "00000000-0000-0000-0000-000000000003"
	areaCode := "QMS"

	mock.ExpectQuery(`SELECT coalesce\(process_area_code_snapshot, ?''\), coalesce\(controlled_document_id::text, ?''\) FROM documents WHERE`).
		WithArgs(tenantID, revisionID).
		WillReturnRows(sqlmock.NewRows([]string{"area_code", "controlled_document_id"}).AddRow(areaCode, cdID))

	mock.ExpectQuery(`SELECT id FROM approval_instances`).
		WithArgs(tenantID, revisionID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	docReader := stubDocumentReader{}
	b := NewDocumentContextBuilder(db, stubRevisionReader{}, stubWorkflowReader{}, stubRegistryReader{}, docReader)

	in, err := b.Build(context.Background(), tenantID, revisionID, ApproverContext{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if in.DocumentReader == nil {
		t.Fatalf("DocumentReader was not set on ResolveInput")
	}
	if in.ControlledDocumentID != cdID {
		t.Fatalf("ControlledDocumentID = %q, want %q", in.ControlledDocumentID, cdID)
	}
	if in.AreaCodeSnapshot != areaCode {
		t.Fatalf("AreaCodeSnapshot = %q, want %q", in.AreaCodeSnapshot, areaCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
