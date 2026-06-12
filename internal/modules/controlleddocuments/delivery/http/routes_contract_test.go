package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	"metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// newPermissiveMockDB returns a sqlmock *sql.DB with permissive expectations for
// the authz.Require GUC sequence.  Used in tests that exercise handler routing
// when the service now unconditionally uses a DB transaction for authz.
//
// Query pool uses arg-count discrimination (WithoutArgs / 2-arg AnyArg) so that
// 0-arg GUC reads and 2-arg EXISTS checks each consume from their own ordered
// sub-pool, matching the templates delivery/http pattern.
func newPermissiveMockDB(t *testing.T) *sql.DB {
	t.Helper()
	anyMatcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { return nil })
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(anyMatcher))
	if err != nil {
		t.Fatalf("newPermissiveMockDB: sqlmock.New: %v", err)
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
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("").
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

type fakeRegistryDocs struct{}

func (f fakeRegistryDocs) GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error) {
	return nil, controlleddocumentsdomain.ErrCDNotFound
}

func (f fakeRegistryDocs) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*controlleddocumentsdomain.ControlledDocument, error) {
	return nil, nil
}

func (f fakeRegistryDocs) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error) {
	return false, nil
}

func (f fakeRegistryDocs) List(ctx context.Context, tenantID string, filter controlleddocumentsdomain.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, bool, error) {
	return nil, false, nil
}
func (f fakeRegistryDocs) CanRead(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (bool, error) {
	return true, nil
}

func (f fakeRegistryDocs) GetActiveInstance(ctx context.Context, tenantID, controlledDocumentID string) (*controlleddocumentsdomain.ActiveDocumentInstance, error) {
	return nil, nil
}

func (f fakeRegistryDocs) Create(ctx context.Context, doc *controlleddocumentsdomain.ControlledDocument) error {
	return nil
}

func (f fakeRegistryDocs) CreateTx(ctx context.Context, tx controlleddocumentsdomain.DBTX, doc *controlleddocumentsdomain.ControlledDocument) error {
	return nil
}

func (f fakeRegistryDocs) UpdateStatus(ctx context.Context, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error {
	return nil
}

func (f fakeRegistryDocs) UpdateStatusTx(_ context.Context, _ controlleddocumentsdomain.DBTX, _, _ string, _ controlleddocumentsdomain.CDStatus, _ time.Time) error {
	return nil
}

type fakeSequenceAllocator struct{}

func (f fakeSequenceAllocator) NextAndIncrement(ctx context.Context, tx controlleddocumentsdomain.DBTX, tenantID, profileCode, areaCode string) (int, error) {
	return 1, nil
}

func (f fakeSequenceAllocator) Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	return 1, nil
}

func (f fakeSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error {
	return nil
}

type fakeTemplateChecker struct{}

func (f fakeTemplateChecker) GetTemplateVersionState(ctx context.Context, tenantID, templateVersionID string) (*string, string, error) {
	return nil, "", nil
}

type fakeProfileReader struct{}

func (f fakeProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	return &taxonomydomain.DocumentProfile{Code: taxonomydomain.ProfileCode(code), TenantID: tenantID}, nil
}

type fakeAreaReader struct{}

func (f fakeAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	return &taxonomydomain.ProcessArea{Code: taxonomydomain.AreaCode(code), TenantID: tenantID}, nil
}

type fakeGovernanceLogger struct{}

func (f fakeGovernanceLogger) Log(ctx context.Context, e taxonomydomain.GovernanceEvent) error {
	return nil
}

func (f fakeGovernanceLogger) LogTx(_ context.Context, _ db.Tx, _ taxonomydomain.GovernanceEvent) error {
	return nil
}

type spyControlledDocumentService struct {
	gotCreate          application.CreateControlledDocumentCmd
	gotListFilter      application.CDFilter
	gotListTenantID    string
	listResult         []controlleddocumentsdomain.ControlledDocument
	gotRevision        application.CreateRevisionCmd
	revisionResult     *controlleddocumentsdomain.DocumentRef
	revisionErr        error
	gotGetID           string
	getResult          *controlleddocumentsdomain.ControlledDocument
	getErr             error
	gotObsoleteID      string
	gotSupersedeID     string
	gotPeekProfile     string
	gotPeekArea        string
	peekResult         int
	activeInstResult   *controlleddocumentsdomain.ActiveDocumentInstance
	activeInstErr      error
}

func (s *spyControlledDocumentService) Create(ctx context.Context, cmd application.CreateControlledDocumentCmd) (*application.CreateResult, error) {
	s.gotCreate = cmd
	return &application.CreateResult{
		ControlledDocument: &controlleddocumentsdomain.ControlledDocument{ID: "cd-1"},
		DocumentRef:        &controlleddocumentsdomain.DocumentRef{ID: "doc-1", ContentHash: "hash-1"},
	}, nil
}

func (s *spyControlledDocumentService) List(ctx context.Context, tenantID string, filter application.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, bool, error) {
	s.gotListTenantID = tenantID
	s.gotListFilter = filter
	return s.listResult, false, nil
}

func (s *spyControlledDocumentService) CreateRevision(ctx context.Context, cmd application.CreateRevisionCmd) (*controlleddocumentsdomain.DocumentRef, error) {
	s.gotRevision = cmd
	if s.revisionErr != nil {
		return nil, s.revisionErr
	}
	if s.revisionResult != nil {
		return s.revisionResult, nil
	}
	return &controlleddocumentsdomain.DocumentRef{
		ID:          "11111111-1111-1111-1111-111111111111",
		ContentHash: "hash-1",
	}, nil
}

func (s *spyControlledDocumentService) Get(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error) {
	s.gotGetID = id
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.getResult != nil {
		return s.getResult, nil
	}
	return nil, controlleddocumentsdomain.ErrCDNotFound
}

func (s *spyControlledDocumentService) GetActiveInstance(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ActiveDocumentInstance, error) {
	return s.activeInstResult, s.activeInstErr
}

func (s *spyControlledDocumentService) Obsolete(ctx context.Context, tenantID, id string) error {
	s.gotObsoleteID = id
	return nil
}

func (s *spyControlledDocumentService) Supersede(ctx context.Context, tenantID, id string) error {
	s.gotSupersedeID = id
	return nil
}

func (s *spyControlledDocumentService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	s.gotPeekProfile = profileCode
	s.gotPeekArea = areaCode
	if s.peekResult != 0 {
		return s.peekResult, nil
	}
	return 1, nil
}

// helpers

func newTestHandler(db *sql.DB) *Handler {
	svc := application.NewControlledDocumentService(
		db,
		fakeRegistryDocs{},
		fakeSequenceAllocator{},
		fakeTemplateChecker{},
		fakeProfileReader{},
		fakeAreaReader{},
		fakeGovernanceLogger{},
		nil,
	)
	return NewHandler(svc, db)
}

// newSpyHandler builds a Handler backed by a configurable spy service.
// Use for tests that need to control GetActiveInstance (and other) responses.
func newSpyHandler(spy *spyControlledDocumentService) *Handler {
	return &Handler{svc: spy}
}

func newAuthedRequest(t *testing.T, method, url, tenantID string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	ctx := tenant.WithTenantID(req.Context(), tenantID)
	ctx = iamdomain.WithAuthContext(ctx, "user-1", []iamdomain.Role{"author"})
	req = req.WithContext(ctx)
	// extract {id} path value from URL pattern /api/v1/controlled-documents/{id}/active-document
	// httptest doesn't set path values automatically; set manually
	// URL format: /api/v1/controlled-documents/<id>/active-document
	// Parse id from URL
	const prefix = "/api/v1/controlled-documents/"
	const suffix = "/active-document"
	if len(url) > len(prefix)+len(suffix) {
		id := url[len(prefix) : len(url)-len(suffix)]
		req.SetPathValue("id", id)
	}
	return req
}

func sampleControlledDocument() controlleddocumentsdomain.ControlledDocument {
	departmentCode := "DP"
	sequenceNum := 1
	overrideTemplateVersionID := "66666666-6666-6666-6666-666666666666"
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	return controlleddocumentsdomain.ControlledDocument{
		ID:                        "77777777-7777-7777-7777-777777777777",
		TenantID:                  "88888888-8888-8888-8888-888888888888",
		ProfileCode:               "DC",
		ProcessAreaCode:           "RH",
		DepartmentCode:            &departmentCode,
		Code:                      "DC-RH-001",
		SequenceNum:               &sequenceNum,
		Title:                     "Policy",
		OwnerUserID:               "user-1",
		OverrideTemplateVersionID: &overrideTemplateVersionID,
		Visibility: controlleddocumentsdomain.Visibility{
			Scope:     controlleddocumentsdomain.VisibilityScopeRestricted,
			AreaCodes: []string{"RH"},
			UserIDs:   []string{"user-2"},
		},
		Status:    controlleddocumentsdomain.CDStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// existing tests

func TestRegistryHandler_ErrorEnvelopeContract(t *testing.T) {
	svc := application.NewControlledDocumentService(
		nil,
		fakeRegistryDocs{},
		fakeSequenceAllocator{},
		fakeTemplateChecker{},
		fakeProfileReader{},
		fakeAreaReader{},
		fakeGovernanceLogger{},
		nil,
	)
	handler := NewHandler(svc, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/99999999-9999-9999-9999-999999999999", nil)
	ctx := tenant.WithTenantID(req.Context(), "test-tenant")
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var apiErr problem.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal api error: %v body=%s", err, rec.Body.String())
	}
	if apiErr.Code == "" {
		t.Fatalf("expected non-empty code in API error: %s", rec.Body.String())
	}
}

func TestWriteDomainError_TemplateMismatchIs422(t *testing.T) {
	handler := &Handler{}
	rec := httptest.NewRecorder()

	handler.writeDomainError(rec, controlleddocumentsdomain.ErrTemplateProfileMismatch)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != "template_invalid" {
		t.Fatalf("code = %q, want %q", body.Code, "template_invalid")
	}
}

func TestWriteDomainError_TemplateArtifactMissingIs409(t *testing.T) {
	handler := &Handler{}
	rec := httptest.NewRecorder()

	handler.writeDomainError(rec, application.ErrTemplateArtifactMissing)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != "template.artifact_missing" {
		t.Fatalf("code = %q, want %q", body.Code, "template.artifact_missing")
	}
}

func TestWriteDomainError_ActorMissingIs401(t *testing.T) {
	handler := &Handler{}
	rec := httptest.NewRecorder()

	handler.writeDomainError(rec, application.ErrActorMissing)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != "UNAUTHENTICATED" {
		t.Fatalf("code = %q, want UNAUTHENTICATED", body.Code)
	}
}

func TestAtomicCreate_MissingAuthContext_Returns401NotFullTenant(t *testing.T) {
	spy := &spyControlledDocumentService{}
	handler := &Handler{svc: spy}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(`{
		"document_name":"Policy v1",
		"visibility":{"scope":"company","area_codes":[],"user_ids":[]},
		"profile_code":"DC",
		"process_area_code":"RH",
		"title":"Policy",
		"owner_user_id":"user-1"
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if spy.gotCreate.ActorUserID != "" {
		t.Fatalf("service Create must not be invoked when actor missing; got ActorUserID=%q", spy.gotCreate.ActorUserID)
	}
}

func TestWriteDomainError_TemplateArtifactInvariantUnconfiguredIs500(t *testing.T) {
	handler := &Handler{}
	rec := httptest.NewRecorder()

	handler.writeDomainError(rec, application.ErrTemplateArtifactInvariantUnconfigured)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != "template.artifact_invariant_unconfigured" {
		t.Fatalf("code = %q, want %q", body.Code, "template.artifact_invariant_unconfigured")
	}
}

func TestAtomicCreate_MissingDocumentName_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(`{
		"visibility":{"scope":"restricted","area_codes":["RH"],"user_ids":[]},
		"profile_code":"DC",
		"process_area_code":"RH",
		"title":"Policy",
		"owner_user_id":"user-1"
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "document_name") {
		t.Fatalf("body %q does not mention document_name", rec.Body.String())
	}
}

func TestAtomicCreate_UnknownField_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(`{
		"document_name":"Policy v1",
		"visibility":{"scope":"restricted","area_codes":["RH"],"user_ids":[]},
		"profile_code":"DC",
		"process_area_code":"RH",
		"title":"Policy",
		"owner_user_id":"user-1",
		"evilField":true
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "evilField") {
		t.Fatalf("body %q does not mention evilField", rec.Body.String())
	}
}

func TestAtomicCreate_BodyTooLarge_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	oversized := `{"document_name":"` + strings.Repeat("x", int(maxControlledDocumentsJSONBodyBytes)) + `","visibility":{"scope":"company","area_codes":[],"user_ids":[]},"profile_code":"DC","process_area_code":"RH","title":"Policy","owner_user_id":"user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(oversized))
	ctx := tenant.WithTenantID(req.Context(), "test-tenant")
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestAtomicCreate_ForwardsGeneratedOnlyFields(t *testing.T) {
	spy := &spyControlledDocumentService{}
	handler := &Handler{svc: spy}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", strings.NewReader(`{
		"document_name":"Policy v1",
		"visibility":{"scope":"company","area_codes":[],"user_ids":[]},
		"profile_code":"DC",
		"process_area_code":"RH",
		"title":"Policy",
		"owner_user_id":"user-1",
		"template_version_id":"11111111-1111-1111-1111-111111111111",
		"form_data":{"summary":"hello","count":2}
	}`))
	ctx := tenant.WithTenantID(req.Context(), "test-tenant")
	ctx = iamdomain.WithAuthContext(ctx, "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.AtomicCreateControlledDocument(rec, req, controlleddocumentsapi.AtomicCreateControlledDocumentParams{})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotCreate.DocumentName != "Policy v1" {
		t.Fatalf("DocumentName = %q, want Policy v1", spy.gotCreate.DocumentName)
	}
	if spy.gotCreate.TemplateVersionID == nil || *spy.gotCreate.TemplateVersionID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("TemplateVersionID = %v, want 11111111-1111-1111-1111-111111111111", spy.gotCreate.TemplateVersionID)
	}
	if spy.gotCreate.FormData["summary"] != "hello" {
		t.Fatalf("FormData[summary] = %v, want hello", spy.gotCreate.FormData["summary"])
	}
	if spy.gotCreate.VisibilityScope != "company" {
		t.Fatalf("VisibilityScope = %q, want company", spy.gotCreate.VisibilityScope)
	}
}

func TestListControlledDocuments_UsesGeneratedParams(t *testing.T) {
	spy := &spyControlledDocumentService{
		listResult: []controlleddocumentsdomain.ControlledDocument{sampleControlledDocument()},
	}
	handler := &Handler{svc: spy}
	status := controlleddocumentsapi.ListControlledDocumentsParamsStatusActive
	limit := 10
	cursor := "Y3Vyc29y"
	profileCode := "DC"
	processAreaCode := "RH"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	rec := httptest.NewRecorder()

	handler.ListControlledDocuments(rec, req, controlleddocumentsapi.ListControlledDocumentsParams{
		ProfileCode:     &profileCode,
		ProcessAreaCode: &processAreaCode,
		Status:          &status,
		Limit:           &limit,
		Cursor:          &cursor,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotListFilter.ProfileCode == nil || *spy.gotListFilter.ProfileCode != profileCode {
		t.Fatalf("ProfileCode = %v, want %s", spy.gotListFilter.ProfileCode, profileCode)
	}
	if spy.gotListFilter.ProcessAreaCode == nil || *spy.gotListFilter.ProcessAreaCode != processAreaCode {
		t.Fatalf("ProcessAreaCode = %v, want %s", spy.gotListFilter.ProcessAreaCode, processAreaCode)
	}
	if spy.gotListFilter.Status == nil || *spy.gotListFilter.Status != controlleddocumentsdomain.CDStatusActive {
		t.Fatalf("Status = %v, want active", spy.gotListFilter.Status)
	}
	if spy.gotListFilter.Limit != limit || spy.gotListFilter.Cursor != cursor {
		t.Fatalf("limit/cursor = %d/%q, want %d/%q", spy.gotListFilter.Limit, spy.gotListFilter.Cursor, limit, cursor)
	}
	if !strings.Contains(rec.Body.String(), `"items"`) {
		t.Fatalf("body %q does not contain generated items response", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"user_ids":["user-2"]`) {
		t.Fatalf("body %q does not contain persisted visibility user grants", rec.Body.String())
	}
}

func TestListControlledDocuments_InvalidStatus_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents", nil)
	rec := httptest.NewRecorder()
	status := controlleddocumentsapi.ListControlledDocumentsParamsStatus("retired")

	handler.ListControlledDocuments(rec, req, controlleddocumentsapi.ListControlledDocumentsParams{Status: &status})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestListControlledDocuments_ZeroTenantContext_Returns500(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), uuid.Nil.String()))
	rec := httptest.NewRecorder()

	handler.ListControlledDocuments(rec, req, controlleddocumentsapi.ListControlledDocumentsParams{})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body = %s", rec.Code, rec.Body.String())
	}
}

func TestGetControlledDocument_UsesGeneratedResponse(t *testing.T) {
	doc := sampleControlledDocument()
	spy := &spyControlledDocumentService{getResult: &doc}
	handler := &Handler{svc: spy}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/"+doc.ID, nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.GetControlledDocument(rec, req, uuid.MustParse(doc.ID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotGetID != doc.ID {
		t.Fatalf("got id %q, want %q", spy.gotGetID, doc.ID)
	}
	var body controlleddocumentsapi.ControlledDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Id.String() != doc.ID {
		t.Fatalf("id = %s, want %s", body.Id, doc.ID)
	}
}

func TestGetControlledDocument_InvalidPathUUID_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestPreviewControlledDocumentCode_UsesGeneratedParamsAndResponse(t *testing.T) {
	spy := &spyControlledDocumentService{peekResult: 7}
	handler := &Handler{svc: spy}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/preview-code", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.PreviewControlledDocumentCode(rec, req, controlleddocumentsapi.PreviewControlledDocumentCodeParams{
		ProfileCode: " dc ",
		AreaCode:    " rh ",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotPeekProfile != "dc" || spy.gotPeekArea != "rh" {
		t.Fatalf("peek params = %q/%q, want dc/rh", spy.gotPeekProfile, spy.gotPeekArea)
	}
	var body controlleddocumentsapi.PreviewCodeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.NextSeq != 7 || body.Code != "DC-RH-007" {
		t.Fatalf("preview = %+v, want nextSeq 7 and code DC-RH-007", body)
	}
}

func TestCreateControlledDocumentRevision_UsesGeneratedBody(t *testing.T) {
	spy := &spyControlledDocumentService{}
	handler := &Handler{svc: spy}
	cdID := "22222222-2222-2222-2222-222222222222"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents/"+cdID+"/revisions", strings.NewReader(`{
		"name":"Revision 2",
		"template_version_id":"33333333-3333-3333-3333-333333333333",
		"form_data":{"field":"value"}
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.CreateControlledDocumentRevision(rec, req, uuid.MustParse(cdID), controlleddocumentsapi.CreateControlledDocumentRevisionParams{})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotRevision.CDID != cdID || spy.gotRevision.Name != "Revision 2" {
		t.Fatalf("revision cmd = %+v, want cd id and name", spy.gotRevision)
	}
	if spy.gotRevision.TemplateVersionID == nil || *spy.gotRevision.TemplateVersionID != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("TemplateVersionID = %v, want generated UUID string", spy.gotRevision.TemplateVersionID)
	}
	if spy.gotRevision.FormData["field"] != "value" {
		t.Fatalf("FormData[field] = %v, want value", spy.gotRevision.FormData["field"])
	}
}

func TestCreateControlledDocumentRevision_UnknownField_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	cdID := "22222222-2222-2222-2222-222222222222"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents/"+cdID+"/revisions", strings.NewReader(`{
		"name":"Revision 2",
		"evilField":true
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.CreateControlledDocumentRevision(rec, req, uuid.MustParse(cdID), controlleddocumentsapi.CreateControlledDocumentRevisionParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "evilField") {
		t.Fatalf("body %q does not mention evilField", rec.Body.String())
	}
}

func TestCreateControlledDocumentRevision_BodyTooLarge_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	cdID := "22222222-2222-2222-2222-222222222222"
	oversized := `{"name":"` + strings.Repeat("x", int(maxControlledDocumentsJSONBodyBytes)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents/"+cdID+"/revisions", strings.NewReader(oversized))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.CreateControlledDocumentRevision(rec, req, uuid.MustParse(cdID), controlleddocumentsapi.CreateControlledDocumentRevisionParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateControlledDocumentRevision_MissingName_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	cdID := "22222222-2222-2222-2222-222222222222"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents/"+cdID+"/revisions", strings.NewReader(`{
		"form_data":{"field":"value"}
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.CreateControlledDocumentRevision(rec, req, uuid.MustParse(cdID), controlleddocumentsapi.CreateControlledDocumentRevisionParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "name") {
		t.Fatalf("body %q does not mention name", rec.Body.String())
	}
}

func TestCreateControlledDocumentRevision_ActiveSiblingConflict_Returns409(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{revisionErr: controlleddocumentsdomain.ErrActiveRevisionExists}}
	cdID := "22222222-2222-2222-2222-222222222222"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents/"+cdID+"/revisions", strings.NewReader(`{
		"name":"Revision 2"
	}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.CreateControlledDocumentRevision(rec, req, uuid.MustParse(cdID), controlleddocumentsapi.CreateControlledDocumentRevisionParams{})

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "ACTIVE_REVISION_ALREADY_EXISTS") {
		t.Fatalf("body %q does not mention ACTIVE_REVISION_ALREADY_EXISTS", rec.Body.String())
	}
}

func TestObsoleteControlledDocument_UsesGeneratedPathParam(t *testing.T) {
	spy := &spyControlledDocumentService{}
	handler := &Handler{svc: spy}
	cdID := "99999999-9999-9999-9999-999999999999"
	req := httptest.NewRequest(http.MethodPut, "/api/v1/controlled-documents/"+cdID+"/obsolete", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.ObsoleteControlledDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotObsoleteID != cdID {
		t.Fatalf("obsolete id = %q, want %q", spy.gotObsoleteID, cdID)
	}
}

func TestObsoleteControlledDocument_InvalidPathUUID_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/controlled-documents/not-a-uuid/obsolete", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestSupersedeControlledDocument_UsesGeneratedPathParam(t *testing.T) {
	spy := &spyControlledDocumentService{}
	handler := &Handler{svc: spy}
	cdID := "aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa"
	req := httptest.NewRequest(http.MethodPut, "/api/v1/controlled-documents/"+cdID+"/supersede", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()

	handler.SupersedeControlledDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", rec.Code, rec.Body.String())
	}
	if spy.gotSupersedeID != cdID {
		t.Fatalf("supersede id = %q, want %q", spy.gotSupersedeID, cdID)
	}
}

func TestSupersedeControlledDocument_InvalidPathUUID_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/controlled-documents/not-a-uuid/supersede", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestActiveDocumentResponse_IncludesApprovalInstanceID(t *testing.T) {
	approvalInstanceID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	docID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	approvalState := controlleddocumentsapi.UnderReview
	contentHash := "hash-1"
	revVersion := 2
	resp := controlleddocumentsapi.ActiveDocumentResponse{
		DocumentId:         &docID,
		ApprovalState:      &approvalState,
		ContentHash:        &contentHash,
		RevisionVersion:    &revVersion,
		ApprovalInstanceId: &approvalInstanceID,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got["approval_instance_id"] != approvalInstanceID.String() {
		t.Fatalf("approval_instance_id = %v, want %s", got["approval_instance_id"], approvalInstanceID)
	}
}

// E10 contract tests — use spyControlledDocumentService so the handler's
// boundary with the service is tested without touching the database layer.

func ptr[T any](v T) *T { return &v }

// TestActiveDocument_OnlyPublished_Returns200_WithPublishedID: controlled document has only a
// published revision (no draft/under_review/etc.) — getActiveDocument must return 200 with
// publishedDocumentId set and documentId absent.
func TestActiveDocument_OnlyPublished_Returns200_WithPublishedID(t *testing.T) {
	publishedDocID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	spy := &spyControlledDocumentService{
		activeInstResult: &controlleddocumentsdomain.ActiveDocumentInstance{
			PublishedDocumentID: ptr(publishedDocID),
		},
	}
	handler := newSpyHandler(spy)

	tenantID := "tenant-1"
	cdID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["published_document_id"] != publishedDocID {
		t.Errorf("published_document_id = %v, want %s", body["published_document_id"], publishedDocID)
	}
	if _, ok := body["document_id"]; ok {
		t.Errorf("document_id should be absent (omitempty), got %v", body["document_id"])
	}
	if _, ok := body["approval_state"]; ok {
		t.Errorf("approval_state should be absent when no active document exists, got %v", body["approval_state"])
	}
}

// TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth: active draft + published revision
// both exist — both documentId and publishedDocumentId must be present.
func TestActiveDocument_BothActiveAndPublished_Returns200_WithBoth(t *testing.T) {
	activeDocID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	publishedDocID := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	contentHash := "abc123"
	revVersion := 3
	approvalState := "draft"

	spy := &spyControlledDocumentService{
		activeInstResult: &controlleddocumentsdomain.ActiveDocumentInstance{
			DocumentID:          ptr(activeDocID),
			ContentHash:         ptr(contentHash),
			RevisionVersion:     ptr(revVersion),
			ApprovalState:       ptr(approvalState),
			PublishedDocumentID: ptr(publishedDocID),
		},
	}
	handler := newSpyHandler(spy)

	tenantID := "tenant-2"
	cdID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["document_id"] != activeDocID {
		t.Errorf("document_id = %v, want %s", body["document_id"], activeDocID)
	}
	if body["published_document_id"] != publishedDocID {
		t.Errorf("published_document_id = %v, want %s", body["published_document_id"], publishedDocID)
	}
	if _, ok := body["approval_instance_id"]; ok {
		t.Errorf("approval_instance_id should be absent when no in-progress instance exists, got %v", body["approval_instance_id"])
	}
}

func TestActiveDocument_ScheduledActive_ReturnsScheduledState(t *testing.T) {
	activeDocID := "66666666-7777-8888-9999-000000000000"
	publishedDocID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	contentHash := "abc123"
	revVersion := 4
	approvalState := "scheduled"

	spy := &spyControlledDocumentService{
		activeInstResult: &controlleddocumentsdomain.ActiveDocumentInstance{
			DocumentID:          ptr(activeDocID),
			ContentHash:         ptr(contentHash),
			RevisionVersion:     ptr(revVersion),
			ApprovalState:       ptr(approvalState),
			PublishedDocumentID: ptr(publishedDocID),
		},
	}
	handler := newSpyHandler(spy)

	tenantID := "tenant-2"
	cdID := "11111111-2222-3333-4444-555555555555"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["approval_state"] != approvalState {
		t.Fatalf("approval_state = %v, want %s", body["approval_state"], approvalState)
	}
	if _, ok := body["approval_instance_id"]; ok {
		t.Errorf("approval_instance_id should be absent when no in-progress instance exists, got %v", body["approval_instance_id"])
	}
}

func TestActiveDocument_UnderReview_ReturnsApprovalInstanceID(t *testing.T) {
	activeDocID := "aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb"
	publishedDocID := "cccccccc-1111-2222-3333-dddddddddddd"
	approvalInstanceID := "eeeeeeee-1111-2222-3333-ffffffffffff"
	contentHash := "hash-under-review"
	revVersion := 5
	approvalState := "under_review"

	spy := &spyControlledDocumentService{
		activeInstResult: &controlleddocumentsdomain.ActiveDocumentInstance{
			DocumentID:          ptr(activeDocID),
			ContentHash:         ptr(contentHash),
			RevisionVersion:     ptr(revVersion),
			ApprovalState:       ptr(approvalState),
			PublishedDocumentID: ptr(publishedDocID),
			ApprovalInstanceID:  ptr(approvalInstanceID),
		},
	}
	handler := newSpyHandler(spy)

	tenantID := "tenant-4"
	cdID := "12345678-1111-2222-3333-444444444444"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["approval_state"] != approvalState {
		t.Fatalf("approval_state = %v, want %s", body["approval_state"], approvalState)
	}
	if body["approval_instance_id"] != approvalInstanceID {
		t.Fatalf("approval_instance_id = %v, want %s", body["approval_instance_id"], approvalInstanceID)
	}
}

func TestActiveDocument_ServiceError_Returns500(t *testing.T) {
	spy := &spyControlledDocumentService{
		activeInstErr: errors.New("repository error"),
	}
	handler := newSpyHandler(spy)

	tenantID := "tenant-5"
	cdID := "87654321-1111-2222-3333-444444444444"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["code"] != "INTERNAL_ERROR" {
		t.Fatalf("code = %v, want INTERNAL_ERROR", body["code"])
	}
}

// TestPreviewCode tests below use snake_case params which are query params (not JSON body)
// so they don't need migration from the femap

// TestPostControlledDocuments_MissingIdempotencyKey_400: POST to atomic-create
// endpoint without Idempotency-Key header must return 400 with code IDEMPOTENCY_KEY_REQUIRED.
func TestPostControlledDocuments_MissingIdempotencyKey_400(t *testing.T) {
	handler := newTestHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/controlled-documents", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body["code"] != "IDEMPOTENCY_KEY_REQUIRED" {
		t.Fatalf("code = %v, want IDEMPOTENCY_KEY_REQUIRED", body["code"])
	}
}

// TestGetPreviewCode_200: GET /api/v1/controlled-documents/preview-code with valid
// query params returns 200 with profile_code, area_code, next_seq, and code fields.
func TestGetPreviewCode_200(t *testing.T) {
	mockDB := newPermissiveMockDB(t)
	handler := newTestHandler(mockDB)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/preview-code?profile_code=DC&area_code=RH", nil)
	ctx := tenant.WithTenantID(req.Context(), "tenant-1")
	ctx = iamdomain.WithAuthContext(ctx, "user-1", []iamdomain.Role{"author"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	for _, field := range []string{"profile_code", "area_code", "next_seq", "code"} {
		if _, ok := body[field]; !ok {
			t.Errorf("missing field %q in response: %s", field, rec.Body.String())
		}
	}
}

// TestGetPreviewCode_MissingParams_400: GET /api/v1/controlled-documents/preview-code
// without required query params returns 400.
func TestGetPreviewCode_MissingParams_400(t *testing.T) {
	handler := newTestHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/preview-code?profile_code=DC", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

// TestActiveDocument_NoneExist_Returns404: no active doc and no published revision —
// must return 404.
func TestActiveDocument_NoneExist_Returns404(t *testing.T) {
	// spy returns nil, nil → handler maps to 404.
	spy := &spyControlledDocumentService{activeInstResult: nil}
	handler := newSpyHandler(spy)

	tenantID := "tenant-3"
	cdID := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/"+cdID+"/active-document", tenantID)
	rec := httptest.NewRecorder()
	handler.GetActiveDocument(rec, req, uuid.MustParse(cdID))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404; body=%s", rec.Code, rec.Body.String())
	}
}

func TestActiveDocument_InvalidPathUUID_Returns400(t *testing.T) {
	handler := &Handler{svc: &spyControlledDocumentService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/controlled-documents/not-a-uuid/active-document", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

// TestGetActiveDocument_ServiceReturnsErrNoActiveInstance_WiresNO_ACTIVE_INSTANCE
// asserts that ErrNoActiveInstance (returned by the service for the denied/absent
// paths) is mapped to 404 NO_ACTIVE_INSTANCE — not CONTROLLED_DOCUMENT_NOT_FOUND.
func TestGetActiveDocument_ServiceReturnsErrNoActiveInstance_WiresNO_ACTIVE_INSTANCE(t *testing.T) {
	spy := &spyControlledDocumentService{
		activeInstErr: controlleddocumentsdomain.ErrNoActiveInstance,
	}
	handler := &Handler{svc: spy}
	req := newAuthedRequest(t, http.MethodGet,
		"/api/v1/controlled-documents/11111111-1111-1111-1111-111111111111/active-document",
		"88888888-8888-8888-8888-888888888888",
	)
	rec := httptest.NewRecorder()

	handler.GetActiveDocument(rec, req, openapi_types.UUID{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404; body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	if body.Code != "NO_ACTIVE_INSTANCE" {
		t.Fatalf("code=%q, want NO_ACTIVE_INSTANCE", body.Code)
	}
}
