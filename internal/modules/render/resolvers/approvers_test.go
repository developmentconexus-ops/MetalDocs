package resolvers

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type fakeWorkflowReader struct {
	approvers           []ApproverInfo
	approversByInstance map[string][]ApproverInfo
	finalApprovalDate   time.Time
	err                 error
}

func (f fakeWorkflowReader) GetApprovers(_ context.Context, _ TenantID, _ RevisionID, instanceID ApprovalInstanceID) ([]ApproverInfo, error) {
	if f.approversByInstance != nil {
		return f.approversByInstance[string(instanceID)], f.err
	}
	return f.approvers, f.err
}

func (f fakeWorkflowReader) GetFinalApprovalDate(ctx context.Context, tenantID TenantID, revisionID RevisionID) (time.Time, error) {
	return f.finalApprovalDate, f.err
}

func TestApproversResolver_Resolve(t *testing.T) {
	r := ApproversResolver{}
	in := ResolveInput{
		TenantID:   "tenant-a",
		RevisionID: "rev-1",
		WorkflowReader: fakeWorkflowReader{
			approvers: []ApproverInfo{
				{
					UserID:      "u-1",
					DisplayName: "Jane",
					SignedAt:    time.Date(2026, time.April, 21, 14, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	v1, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}

	if v1.Value != "Jane" {
		t.Fatalf("Value = %q, want %q", v1.Value, "Jane")
	}
	if !bytes.Equal(v1.InputsHash, v2.InputsHash) {
		t.Fatal("expected stable hash across repeated resolves")
	}
}

func TestApproversResolver_NoApprovers_ReturnsPortuguesePending(t *testing.T) {
	r := ApproversResolver{}
	in := ResolveInput{
		TenantID:       "t1",
		RevisionID:     "rev1",
		WorkflowReader: fakeWorkflowReader{approvers: nil},
	}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve err = %v", err)
	}
	if out.Value != "[aguardando aprovação]" {
		t.Fatalf("Value = %q, want %q", out.Value, "[aguardando aprovação]")
	}
}

func TestApproversResolver_FiltersByInstanceID(t *testing.T) {
	want := []ApproverInfo{{UserID: "u-1", DisplayName: "Jane", SignedAt: time.Now()}}
	wr := fakeWorkflowReader{
		approversByInstance: map[string][]ApproverInfo{
			"inst-current": want,
			"inst-stale":   {{UserID: "u-2", DisplayName: "Stale"}},
		},
	}
	r := ApproversResolver{}
	in := ResolveInput{
		TenantID:           "t1",
		RevisionID:         "r1",
		ApprovalInstanceID: "inst-current",
		WorkflowReader:     wr,
	}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	valStr := fmt.Sprintf("%v", out.Value)
	if !strings.Contains(valStr, "Jane") {
		t.Fatalf("want Jane in output, got %q", valStr)
	}
	if strings.Contains(valStr, "Stale") {
		t.Fatalf("stale approver must not appear in output, got %q", valStr)
	}
}
