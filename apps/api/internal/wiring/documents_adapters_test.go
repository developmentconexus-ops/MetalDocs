package wiring

import (
	"context"
	"reflect"
	"testing"

	controlleddocumentsapp "metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// fakeDuplicatorSvc implements controlledDocumentDuplicatorService for unit tests.
type fakeDuplicatorSvc struct {
	source      *controlleddocumentsapp.ControlledDocument
	capturedCmd controlleddocumentsapp.CreateControlledDocumentCmd
}

func (f *fakeDuplicatorSvc) Get(_ context.Context, _, _ string) (*controlleddocumentsapp.ControlledDocument, error) {
	return f.source, nil
}

func (f *fakeDuplicatorSvc) Create(_ context.Context, cmd controlleddocumentsapp.CreateControlledDocumentCmd) (*controlleddocumentsapp.CreateResult, error) {
	f.capturedCmd = cmd
	return &controlleddocumentsapp.CreateResult{
		DocumentRef: &controlleddocumentsdomain.DocumentRef{
			ID:         "doc-1",
			RevisionID: "rev-1",
			SessionID:  "sess-1",
		},
	}, nil
}

func TestDuplicateControlledDocument_PreservesVisibility(t *testing.T) {
	fake := &fakeDuplicatorSvc{
		source: &controlleddocumentsapp.ControlledDocument{
			ProfileCode:     "QMS",
			ProcessAreaCode: "PROD",
			Title:           "Source Doc",
			OwnerUserID:     "owner-1",
			Visibility: controlleddocumentsdomain.Visibility{
				Scope:     controlleddocumentsdomain.VisibilityScope("restricted"),
				AreaCodes: []string{"A"},
				UserIDs:   []string{"u1", "u2"},
			},
		},
	}

	adapter := &controlledDocumentDuplicatorAdapter{svc: fake}

	_, err := adapter.DuplicateControlledDocument(context.Background(), docapp.DuplicateControlledDocumentInput{
		TenantID:             "t1",
		ControlledDocumentID: "cd-1",
		ActorUserID:          "actor",
		DocumentName:         "Dup",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := fake.capturedCmd
	if cmd.VisibilityScope != "restricted" {
		t.Errorf("VisibilityScope: got %q, want %q", cmd.VisibilityScope, "restricted")
	}
	if !reflect.DeepEqual(cmd.VisibilityAreaCodes, []string{"A"}) {
		t.Errorf("VisibilityAreaCodes: got %v, want %v", cmd.VisibilityAreaCodes, []string{"A"})
	}
	if !reflect.DeepEqual(cmd.VisibilityUserIDs, []string{"u1", "u2"}) {
		t.Errorf("VisibilityUserIDs: got %v, want %v", cmd.VisibilityUserIDs, []string{"u1", "u2"})
	}
}

type fakeProfileRepo struct {
	profile *taxonomydomain.DocumentProfile
}

func (f *fakeProfileRepo) GetByCode(_ context.Context, _ string, _ taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error) {
	return f.profile, nil
}

type fakeTplState struct {
	status     *string
	docType    string
	gotTenant  string
	gotVersion string
}

func (f *fakeTplState) GetTemplateVersionState(_ context.Context, tenantID, versionID string) (*string, string, error) {
	f.gotTenant = tenantID
	f.gotVersion = versionID
	return f.status, f.docType, nil
}

func strptr(s string) *string { return &s }

func TestProfileDefaults_ReadsRealStatusFromPort(t *testing.T) {
	versionID := "tv-obsolete-1"
	obsolete := "obsolete"
	tpl := &fakeTplState{status: &obsolete, docType: "po"}
	adapter := NewProfileDefaults(
		&fakeProfileRepo{profile: &taxonomydomain.DocumentProfile{DefaultTemplateVersionID: &versionID}},
		tpl,
	)

	id, status, err := adapter.GetDefaultTemplateVersionID(context.Background(), "tenant-1", "po")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == nil || *id != versionID {
		t.Fatalf("templateVersionID: got %v, want %q", id, versionID)
	}
	// The real status must flow through — NOT the old hardcoded "published".
	if status == nil || *status != "obsolete" {
		t.Fatalf("status: got %v, want %q (real status from port, not hardcoded)", status, "obsolete")
	}
	if tpl.gotTenant != "tenant-1" || tpl.gotVersion != versionID {
		t.Fatalf("port called with (%q,%q), want (tenant-1,%q)", tpl.gotTenant, tpl.gotVersion, versionID)
	}
}

func TestProfileDefaults_NoDefaultTemplate_ReturnsNil(t *testing.T) {
	adapter := NewProfileDefaults(
		&fakeProfileRepo{profile: &taxonomydomain.DocumentProfile{DefaultTemplateVersionID: nil}},
		&fakeTplState{status: strptr("published"), docType: "po"},
	)
	id, status, err := adapter.GetDefaultTemplateVersionID(context.Background(), "tenant-1", "po")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != nil || status != nil {
		t.Fatalf("no-default: got (%v,%v), want (nil,nil)", id, status)
	}
}
