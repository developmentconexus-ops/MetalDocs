package wiring

import (
	"context"
	"reflect"
	"testing"

	controlleddocumentsapp "metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docapp "metaldocs/internal/modules/documents/application"
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
