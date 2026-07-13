package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamauthz "metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	tmplhttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/templates/domain"
	platformdb "metaldocs/internal/platform/db"
)

// Fakes for the approval-kernel seams (M3 P3.S2b-4, R2a). Mirror
// approval/http's own fakeDecisionService pattern (signoff_handler_test.go):
// the thin templates handler depends on narrow interfaces, not the concrete
// approval/application service structs, so these behavior tests never touch
// a real database or the password-reauth signature verifier.

type fakeApprovalSubmit struct {
	gotReq approvalapp.TemplateSubmitRequest
	res    approvalapp.TemplateSubmitResult
	err    error
	calls  int
}

func (f *fakeApprovalSubmit) SubmitTemplateVersionForReview(_ context.Context, _ platformdb.TxRunner, req approvalapp.TemplateSubmitRequest) (approvalapp.TemplateSubmitResult, error) {
	f.calls++
	f.gotReq = req
	return f.res, f.err
}

// PreviewRoute is a no-op stub only to satisfy the approvalSubmitService
// interface — this fake drives the submit path tests, which never call
// PreviewRoute. PreviewRoute's own behavior is covered in
// internal/modules/approval/application/route_preview_test.go and the
// resolution-parity integration test under tests/integration/approval/.
func (f *fakeApprovalSubmit) PreviewRoute(_ context.Context, _ platformdb.TxRunner, _, _ string) (approvalapp.RoutePreview, error) {
	return approvalapp.RoutePreview{}, nil
}

type fakeApprovalDecision struct {
	gotReq approvalapp.SignoffRequest
	res    approvalapp.SignoffResult
	err    error
	calls  int
}

func (f *fakeApprovalDecision) RecordSignoff(_ context.Context, _ platformdb.TxRunner, req approvalapp.SignoffRequest) (approvalapp.SignoffResult, error) {
	f.calls++
	f.gotReq = req
	return f.res, f.err
}

type fakeApprovalRead struct {
	inst *approvaldomain.Instance
	err  error
}

func (f *fakeApprovalRead) LoadActiveInstanceBySubjectForMutation(_ context.Context, _ platformdb.TxRunner, _, _, _ string) (*approvaldomain.Instance, error) {
	return f.inst, f.err
}

// stubTxRunner satisfies platformdb.TxRunner without ever touching a real
// *sql.DB — the fakes above ignore the tx argument entirely.
type stubTxRunner struct{}

func (stubTxRunner) Do(_ context.Context, fn func(tx *sql.Tx) error) error { return fn(nil) }
func (stubTxRunner) DoReadOnly(_ context.Context, fn func(tx *sql.Tx) error) error {
	return fn(nil)
}

// newMuxWithApprovalKernelOnly builds a templates Handler wired the same way
// newMux/newMuxWithPresigner do (routes_create_test.go), plus
// WithApprovalKernel — the seam the two new kernel routes need. Kept
// separate from newMux so every other templates HTTP test is unaffected.
func newMuxWithApprovalKernelOnly(t *testing.T, authz tmplhttp.AuthzFunc, repo *fakeRepo, submit *fakeApprovalSubmit, decision *fakeApprovalDecision, read *fakeApprovalRead) *http.ServeMux {
	t.Helper()
	mockDB := newPermissiveMockDB(t)
	svc := application.New(repo, fakePresigner{}, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(mockDB))
	h := tmplhttp.New(svc, authz, mockDB).WithApprovalKernel(submit, decision, read, stubTxRunner{})
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

const (
	testTemplateID = "11111111-1111-1111-1111-111111111111"
	testVersionID  = "22222222-2222-4222-8222-222222222222"
	testTenantID   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func seedDraftTemplateAndVersion(repo *fakeRepo) {
	repo.templates[testTemplateID] = &domain.Template{ID: testTemplateID, TenantID: testTenantID}
	repo.versions[testVersionID] = &domain.TemplateVersion{
		ID:            testVersionID,
		TemplateID:    testTemplateID,
		VersionNumber: 1,
		Status:        domain.VersionStatusDraft,
		ContentHash:   "deadbeef",
		AuthorID:      "author-1",
	}
}

func TestSubmitTemplateVersionForApproval_Happy(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)

	instanceID := uuid.New().String()
	submit := &fakeApprovalSubmit{res: approvalapp.TemplateSubmitResult{InstanceID: instanceID}}
	decision := &fakeApprovalDecision{}
	read := &fakeApprovalRead{}

	var gotAction string
	authz := func(_ *http.Request, _, _, action string) error {
		gotAction = action
		return nil
	}
	mux := newMuxWithApprovalKernelOnly(t, authz, repo, submit, decision, read)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/submit-for-approval", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if gotAction != "template.submit" {
		t.Fatalf("authz action = %q, want template.submit", gotAction)
	}
	if submit.calls != 1 {
		t.Fatalf("SubmitTemplateVersionForReview calls = %d, want 1", submit.calls)
	}
	if submit.gotReq.TemplateID != testTemplateID {
		t.Errorf("req.TemplateID = %q, want %q", submit.gotReq.TemplateID, testTemplateID)
	}
	if submit.gotReq.TemplateVersionID != testVersionID {
		t.Errorf("req.TemplateVersionID = %q, want %q", submit.gotReq.TemplateVersionID, testVersionID)
	}
	if submit.gotReq.IdempotencyKey == "" {
		t.Error("req.IdempotencyKey must not be empty")
	}

	var out struct {
		Data struct {
			InstanceID    string `json:"instance_id"`
			VersionStatus string `json:"version_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.InstanceID != instanceID {
		t.Errorf("data.instance_id = %q, want %q", out.Data.InstanceID, instanceID)
	}
	if out.Data.VersionStatus != "under_review" {
		t.Errorf("data.version_status = %q, want under_review", out.Data.VersionStatus)
	}
}

func TestSubmitTemplateVersionForApproval_WrongState(t *testing.T) {
	repo := newFakeRepo()
	repo.templates[testTemplateID] = &domain.Template{ID: testTemplateID, TenantID: testTenantID}
	repo.versions[testVersionID] = &domain.TemplateVersion{
		ID:            testVersionID,
		TemplateID:    testTemplateID,
		VersionNumber: 1,
		Status:        domain.VersionStatusUnderReview, // not draft
	}
	submit := &fakeApprovalSubmit{err: approvalapp.ErrTemplateVersionNotDraft}
	mux := newMuxWithApprovalKernelOnly(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo, submit, &fakeApprovalDecision{}, &fakeApprovalRead{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/submit-for-approval", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "CONFLICT_ERROR" {
		t.Fatalf("expected error.code=CONFLICT_ERROR, got %q", out.Code)
	}
}

func TestSubmitTemplateVersionForApproval_NoActiveRoute(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)
	submit := &fakeApprovalSubmit{err: approvalapp.ErrNoActiveApprovalRoute}
	mux := newMuxWithApprovalKernelOnly(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo, submit, &fakeApprovalDecision{}, &fakeApprovalRead{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/submit-for-approval", nil)
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "APPROVAL_ROUTE_MISSING" {
		t.Fatalf("expected error.code=APPROVAL_ROUTE_MISSING, got %q", out.Code)
	}
}

func stageInstance(id string) approvaldomain.StageInstance {
	return approvaldomain.StageInstance{ID: id, Status: approvaldomain.StageActive}
}

func TestSignoffTemplateVersion_Happy(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)
	repo.versions[testVersionID].Status = domain.VersionStatusUnderReview

	inst := &approvaldomain.Instance{
		ID:     uuid.New().String(),
		Status: approvaldomain.InstanceInProgress,
		Subject: approvaldomain.Subject{
			Kind: approvaldomain.SubjectKindTemplate,
			Key:  testVersionID,
		},
		Stages: []approvaldomain.StageInstance{stageInstance(uuid.New().String())},
	}
	read := &fakeApprovalRead{inst: inst}
	decision := &fakeApprovalDecision{res: approvalapp.SignoffResult{StageCompleted: true}}

	var gotAction string
	authz := func(_ *http.Request, _, _, action string) error {
		gotAction = action
		return nil
	}
	mux := newMuxWithApprovalKernelOnly(t, authz, repo, &fakeApprovalSubmit{}, decision, read)

	body, _ := json.Marshal(map[string]any{"decision": "approve", "password": "correct-horse"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/signoff", bytes.NewReader(body))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if gotAction != "template.approve" {
		t.Fatalf("authz action = %q, want template.approve", gotAction)
	}
	if decision.calls != 1 {
		t.Fatalf("RecordSignoff calls = %d, want 1", decision.calls)
	}
	if decision.gotReq.InstanceID != inst.ID {
		t.Errorf("req.InstanceID = %q, want %q", decision.gotReq.InstanceID, inst.ID)
	}
	if decision.gotReq.StageInstanceID != inst.Stages[0].ID {
		t.Errorf("req.StageInstanceID = %q, want %q", decision.gotReq.StageInstanceID, inst.Stages[0].ID)
	}
	if decision.gotReq.Decision != approvaldomain.DecisionApprove {
		t.Errorf("req.Decision = %q, want approve", decision.gotReq.Decision)
	}
	if decision.gotReq.SignaturePayload["password_token"] != "correct-horse" {
		t.Errorf("req.SignaturePayload[password_token] = %v, want correct-horse", decision.gotReq.SignaturePayload["password_token"])
	}

	var out struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Outcome != "stage_completed" {
		t.Errorf("outcome = %q, want stage_completed", out.Outcome)
	}
}

func TestSignoffTemplateVersion_CapabilityDenied(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)

	authz := func(_ *http.Request, tenantID, _, action string) error {
		if action == string(iamdomain.CapTemplateApprove) {
			return iamauthz.ErrCapDenied{Capability: action, AreaCode: "tenant", ActorID: "user-a"}
		}
		return nil
	}
	mux := newMuxWithApprovalKernelOnly(t, authz, repo, &fakeApprovalSubmit{}, &fakeApprovalDecision{}, &fakeApprovalRead{})

	body, _ := json.Marshal(map[string]any{"decision": "approve", "password": "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/signoff", bytes.NewReader(body))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSignoffTemplateVersion_NoActiveInstance(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)
	read := &fakeApprovalRead{err: approvalapp.ErrNoActiveInstance}
	mux := newMuxWithApprovalKernelOnly(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo, &fakeApprovalSubmit{}, &fakeApprovalDecision{}, read)

	body, _ := json.Marshal(map[string]any{"decision": "approve", "password": "x"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+testTemplateID+"/versions/1/signoff", bytes.NewReader(body))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// TestKernelRoutes_ExistAndRoundTrip is the contract-surface guard (M3
// P3.S2b-4): both new routes are registered by the generated ServerInterface
// wiring and reach the handler (not a 404 from the mux), completing round
// trips through the fakes above.
func TestKernelRoutes_ExistAndRoundTrip(t *testing.T) {
	repo := newFakeRepo()
	seedDraftTemplateAndVersion(repo)

	inst := &approvaldomain.Instance{
		ID:      uuid.New().String(),
		Status:  approvaldomain.InstanceInProgress,
		Subject: approvaldomain.Subject{Kind: approvaldomain.SubjectKindTemplate, Key: testVersionID},
		Stages:  []approvaldomain.StageInstance{stageInstance(uuid.New().String())},
	}
	submit := &fakeApprovalSubmit{res: approvalapp.TemplateSubmitResult{InstanceID: uuid.New().String()}}
	decision := &fakeApprovalDecision{res: approvalapp.SignoffResult{StageCompleted: true}}
	read := &fakeApprovalRead{inst: inst}
	mux := newMuxWithApprovalKernelOnly(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo, submit, decision, read)

	tests := []struct {
		name string
		path string
		body []byte
	}{
		{name: "submit-for-approval", path: "/api/v1/templates/" + testTemplateID + "/versions/1/submit-for-approval"},
		{name: "signoff", path: "/api/v1/templates/" + testTemplateID + "/versions/1/signoff", body: jsonBody(t, map[string]any{"decision": "approve", "password": "x"})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Reader
			if tt.body != nil {
				body = bytes.NewReader(tt.body)
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(http.MethodPost, tt.path, body)
			withHeaders(req)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200 (route exists and round-trips), got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}
