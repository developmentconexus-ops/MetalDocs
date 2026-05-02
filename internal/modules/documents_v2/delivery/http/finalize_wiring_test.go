package http_test

import (
	"context"
	"database/sql"
	"testing"

	deliveryhttp "metaldocs/internal/modules/documents_v2/delivery/http"
	approvalapp "metaldocs/internal/modules/documents_v2/approval/application"
)

type stubApprovalSubmitter struct {
	called bool
}

func (s *stubApprovalSubmitter) SubmitRevisionForReview(_ context.Context, _ *sql.DB, _ approvalapp.SubmitRequest) (approvalapp.SubmitResult, error) {
	s.called = true
	return approvalapp.SubmitResult{InstanceID: "inst-test"}, nil
}

func TestNewHandlerWithSubmit_Compiles(t *testing.T) {
	stub := &stubApprovalSubmitter{}
	h := deliveryhttp.NewHandlerWithSubmit(nil, nil, stub)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if stub.called {
		t.Fatal("submit should not be called during construction")
	}
}
