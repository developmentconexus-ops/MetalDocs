package application_test

// snapshot_wire_test.go — unit test asserting SnapshotService.ResolveTemplate
// is called inside CreateDocument when wired via NewServiceWithSnapshot, and
// that the resolved snapshot is set on the document before repo.CreateDocument.

import (
	"context"
	"encoding/json"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
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

func TestCreateDocument_SnapshotPopulated(t *testing.T) {
	const tenantID = tenant.DevTenantID

	innerRepo := &fakeRepo{createDocIDs: [3]string{"doc-snap-1", "rev-snap-1", "sess-snap-1"}}
	repo := &captureRepo{fakeRepo: innerRepo}

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd-snap-1",
		TenantID:        tenantID,
		ProfileCode:     "PROC",
		ProcessAreaCode: "AREA-01",
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	reader := &trackingSnapshotReader{
		snap: domain.TemplateSnapshot{
			PlaceholderSchemaJSON: []byte(`{"placeholders":[]}`),
			CompositionJSON:       []byte(`{}`),
			BodyDocxS3Key:         "s3://t/k",
		},
	}
	// writer is nil — ResolveTemplate does not use it
	snapSvc := application.NewSnapshotService(reader, nil)

	svc := application.NewServiceWithSnapshot(
		repo,
		&fakePresigner{hashReturn: "h_init"},
		fakeTplReader{},
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeControlledDocumentReader{cd: cd},
		&fakeAuthzChecker{},
		&fakeProfileDefaultTemplateReader{id: strptr("tv-snap-1"), status: strptr("published")},
		snapSvc,
	)

	_, err := svc.CreateDocument(context.Background(), application.CreateDocumentInput{
		TenantID:             tenantID,
		ActorUserID:          "user-1",
		ControlledDocumentID: "cd-snap-1",
		TemplateVersionID:    "tv-snap-1",
		Name:                 "Test Doc",
		FormData:             json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	if !reader.called {
		t.Fatal("expected SnapshotService.ResolveTemplate to call LoadForSnapshot")
	}
	if repo.createdDoc == nil {
		t.Fatal("captureRepo did not capture document")
	}
	if string(repo.createdDoc.TemplateSnapshot.PlaceholderSchemaJSON) != string(reader.snap.PlaceholderSchemaJSON) {
		t.Fatalf("doc.TemplateSnapshot not populated: got %q", repo.createdDoc.TemplateSnapshot.PlaceholderSchemaJSON)
	}
}
