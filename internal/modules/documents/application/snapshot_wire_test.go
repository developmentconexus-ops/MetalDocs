package application_test

// snapshot_wire_test.go — unit test asserting SnapshotService.ResolveTemplate
// is called inside cloneIntoTx when wired via NewServiceWithSnapshot, and
// that the resolved snapshot is set on the document before repo.CreateDocumentTx.

import (
	"context"
	"database/sql"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/tenant"
)

// trackingSnapshotReader records LoadForSnapshot calls.
type trackingSnapshotReader struct {
	snap   domain.TemplateSnapshot
	called bool
}

func (r *trackingSnapshotReader) LoadForSnapshot(_ context.Context, _, _ string) (domain.TemplateSnapshot, error) {
	r.called = true
	return r.snap, nil
}

// captureTxRepo wraps fakeRepo and captures the document passed to CreateDocumentTx.
type captureTxRepo struct {
	*fakeRepo
	createdDoc  *domain.Document
	initialHash string
}

func (r *captureTxRepo) CreateDocumentTx(ctx context.Context, _ db.Tx, d *domain.Document, initialContentHash, _ string, _ []templatesdomain.Placeholder) (string, string, string, error) {
	r.createdDoc = d
	r.initialHash = initialContentHash
	return r.fakeRepo.CreateDocumentTx(ctx, (*sql.Tx)(nil), d, initialContentHash, "", nil)
}

func TestCloneIntoTx_SnapshotPopulated(t *testing.T) {
	const tenantID = tenant.DevTenantID

	innerRepo := &fakeRepo{createDocIDs: [3]string{"doc-snap-1", "rev-snap-1", "sess-snap-1"}}
	repo := &captureTxRepo{fakeRepo: innerRepo}

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd-snap-1",
		TenantID:        tenantID,
		ProfileCode:     "PROC",
		ProcessAreaCode: "AREA-01",
		Code:            "PROC-001",
		OwnerUserID:     "user-1",
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	reader := &trackingSnapshotReader{
		snap: domain.TemplateSnapshot{
			PlaceholderSchemaJSON: []byte(`{"placeholders":[]}`),
			CompositionJSON:       []byte(`{}`),
			BodyDocxS3Key:         "s3://t/k",
		},
	}
	snapSvc := application.NewSnapshotService(reader)

	svc := application.NewServiceWithSnapshot(
		repo,
		&fakePresigner{},
		fakeTplReader{},
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeProfileDefaultTemplateReader{id: strptr("tv-snap-1"), status: strptr("published")},
		snapSvc,
	)
	initializer := application.NewCDDocumentInitializer(svc)

	req, err := controlleddocumentsdomain.NewCloneTemplateRequest(nil, "Test Doc", nil)
	if err != nil {
		t.Fatalf("NewCloneTemplateRequest: %v", err)
	}

	_, err = initializer.CloneTemplate(context.Background(), (*sql.Tx)(nil), cd, req)
	if err != nil {
		t.Fatalf("CloneTemplate: %v", err)
	}
	if !reader.called {
		t.Fatal("expected SnapshotService.ResolveTemplate to call LoadForSnapshot")
	}
	if repo.createdDoc == nil {
		t.Fatal("captureTxRepo did not capture document")
	}
	if string(repo.createdDoc.TemplateSnapshot.PlaceholderSchemaJSON) != string(reader.snap.PlaceholderSchemaJSON) {
		t.Fatalf("doc.TemplateSnapshot not populated: got %q", repo.createdDoc.TemplateSnapshot.PlaceholderSchemaJSON)
	}
}
