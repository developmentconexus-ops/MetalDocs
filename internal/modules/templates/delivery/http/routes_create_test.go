package http_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	tmplhttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeRepo struct {
	templates       map[string]*domain.Template
	versions        map[string]*domain.TemplateVersion
	audit           []*domain.AuditEvent
	approvalConfigs map[string]*domain.ApprovalConfig
	lockVersions    map[string]int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		templates:       map[string]*domain.Template{},
		versions:        map[string]*domain.TemplateVersion{},
		audit:           []*domain.AuditEvent{},
		approvalConfigs: map[string]*domain.ApprovalConfig{},
		lockVersions:    map[string]int{},
	}
}

func (r *fakeRepo) CreateTemplate(_ context.Context, t *domain.Template) error {
	r.templates[t.ID] = t
	return nil
}

func (r *fakeRepo) GetTemplate(_ context.Context, tenantID, id string) (*domain.Template, error) {
	t, ok := r.templates[id]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) GetTemplateByKey(_ context.Context, tenantID, key string) (*domain.Template, error) {
	for _, t := range r.templates {
		if t.TenantID == tenantID && t.Key == key {
			return t, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) ListTemplates(_ context.Context, f application.ListFilter) ([]*domain.Template, error) {
	out := make([]*domain.Template, 0, len(r.templates))
	for _, t := range r.templates {
		if t.TenantID == f.TenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *fakeRepo) UpdateTemplate(_ context.Context, t *domain.Template) error {
	if _, ok := r.templates[t.ID]; !ok {
		return domain.ErrNotFound
	}
	r.templates[t.ID] = t
	return nil
}

func (r *fakeRepo) CreateVersion(_ context.Context, v *domain.TemplateVersion) error {
	if _, ok := r.lockVersions[v.ID]; !ok {
		r.lockVersions[v.ID] = 0
	}
	r.versions[v.ID] = v
	return nil
}

func (r *fakeRepo) CreateVersionTx(_ context.Context, _ *sql.Tx, v *domain.TemplateVersion) error {
	return r.CreateVersion(context.Background(), v)
}

func (r *fakeRepo) GetVersion(_ context.Context, templateID string, n int) (*domain.TemplateVersion, error) {
	for _, v := range r.versions {
		if v.TemplateID == templateID && v.VersionNumber == n {
			return v, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) GetVersionByID(_ context.Context, id string) (*domain.TemplateVersion, error) {
	v, ok := r.versions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (r *fakeRepo) CreateTemplateTx(_ context.Context, _ *sql.Tx, t *domain.Template) error {
	return r.CreateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateTemplateTx(_ context.Context, _ *sql.Tx, t *domain.Template) error {
	return r.UpdateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateVersionTx(_ context.Context, _ *sql.Tx, v *domain.TemplateVersion) error {
	return r.UpdateVersion(context.Background(), v)
}

func (r *fakeRepo) UpdateVersion(_ context.Context, v *domain.TemplateVersion) error {
	if _, ok := r.versions[v.ID]; !ok {
		return domain.ErrNotFound
	}
	r.versions[v.ID] = v
	return nil
}

func (r *fakeRepo) UpdateVersionDraftCAS(_ context.Context, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	v, ok := r.versions[versionID]
	if !ok {
		return domain.ErrNotFound
	}
	current := r.lockVersions[versionID]
	if current != expectedLockVersion {
		return domain.ErrStaleLockVersion
	}
	v.DocxStorageKey = docxStorageKey
	v.ContentHash = docxContentHash
	r.lockVersions[versionID] = current + 1
	return nil
}

func (r *fakeRepo) ObsoletePreviousPublished(_ context.Context, templateID, keepVersionID string) error {
	for _, v := range r.versions {
		if v.TemplateID == templateID && v.Status == domain.VersionStatusPublished && v.ID != keepVersionID {
			now := time.Now().UTC()
			v.ObsoletedAt = &now
			v.Status = domain.VersionStatusObsolete
		}
	}
	return nil
}

func (r *fakeRepo) ObsoletePreviousPublishedTx(_ context.Context, _ *sql.Tx, templateID, keepVersionID string) error {
	return r.ObsoletePreviousPublished(context.Background(), templateID, keepVersionID)
}

func (r *fakeRepo) GetApprovalConfig(_ context.Context, templateID string) (*domain.ApprovalConfig, error) {
	cfg, ok := r.approvalConfigs[templateID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cfg, nil
}

func (r *fakeRepo) UpsertApprovalConfig(_ context.Context, c *domain.ApprovalConfig) error {
	r.approvalConfigs[c.TemplateID] = c
	return nil
}

func (r *fakeRepo) UpsertApprovalConfigTx(_ context.Context, _ *sql.Tx, c *domain.ApprovalConfig) error {
	return r.UpsertApprovalConfig(context.Background(), c)
}

func (r *fakeRepo) AppendAudit(_ context.Context, e *domain.AuditEvent) error {
	r.audit = append(r.audit, e)
	return nil
}

func (r *fakeRepo) AppendAuditTx(_ context.Context, _ *sql.Tx, e *domain.AuditEvent) error {
	return r.AppendAudit(context.Background(), e)
}

func (r *fakeRepo) ListAudit(_ context.Context, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	_ = limit
	_ = offset
	out := make([]*domain.AuditEvent, 0, len(r.audit))
	for _, e := range r.audit {
		if e.TemplateID == templateID {
			out = append(out, e)
		}
	}
	return out, nil
}

type fakeClock struct{}

func (fakeClock) Now() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

type fakeUUID struct {
	counter int
}

func (u *fakeUUID) New() string {
	u.counter++
	return fmt.Sprintf("id_%d", u.counter)
}

type fakePresigner struct{}

func (fakePresigner) PresignPUT(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://presigned/put/" + key, nil
}

func (fakePresigner) PresignGET(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://presigned/get/" + key, nil
}

func (fakePresigner) HeadContentHash(_ context.Context, _ string) (string, error) {
	return "hash_abc", nil
}

func (fakePresigner) Delete(_ context.Context, _ string) error { return nil }

func newMux(t *testing.T, authz tmplhttp.AuthzFunc, repo *fakeRepo) *http.ServeMux {
	t.Helper()
	svc := application.New(repo, fakePresigner{}, fakeClock{}, &fakeUUID{})
	h := tmplhttp.New(svc, authz)
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func createBody(key string) []byte {
	req := map[string]any{
		"key":         key,
		"name":        "Contract Template",
		"description": "Default contract",
	}
	raw, _ := json.Marshal(req)
	return raw
}

func withHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Idempotency-Key", "11111111-1111-1111-1111-111111111111")
	*req = *req.WithContext(tenant.WithTenantID(req.Context(), "tenant-a"))
	*req = *req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-a", []iamdomain.Role{}))
}

func TestCreateTemplate_Happy(t *testing.T) {
	repo := newFakeRepo()

	var gotTenant, gotArea, gotAction string
	authz := func(_ *http.Request, tenantID, area, action string) error {
		gotTenant = tenantID
		gotArea = area
		gotAction = action
		return nil
	}

	mux := newMux(t, authz, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(createBody("contract-default")))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	if gotTenant != "tenant-a" || gotArea != "*" || gotAction != "template.create" {
		t.Fatalf("unexpected authz call: tenant=%q area=%q action=%q", gotTenant, gotArea, gotAction)
	}

	var out struct {
		Data struct {
			Template struct {
				ID string `json:"id"`
			} `json:"template"`
			Version struct {
				VersionNumber int `json:"version_number"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Data.Template.ID == "" {
		t.Fatal("expected template.id to be present")
	}
	if out.Data.Version.VersionNumber != 1 {
		t.Fatalf("expected version.version_number=1, got %d", out.Data.Version.VersionNumber)
	}
}

func TestCreateTemplate_RejectUnknownField(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewBufferString(`{"key":"contract-default","name":"Contract Template","visibility":"weird"}`))
	withHeaders(req)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "invalid_body" {
		t.Fatalf("expected error.code=invalid_body, got %q", out.Code)
	}
}

func TestCreateTemplate_KeyConflict(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	first := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(createBody("contract-default")))
	withHeaders(first)
	firstRR := httptest.NewRecorder()
	mux.ServeHTTP(firstRR, first)
	if firstRR.Code != http.StatusCreated {
		t.Fatalf("expected first request 201, got %d body=%s", firstRR.Code, firstRR.Body.String())
	}

	second := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(createBody("contract-default")))
	withHeaders(second)
	secondRR := httptest.NewRecorder()
	mux.ServeHTTP(secondRR, second)
	if secondRR.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", secondRR.Code, secondRR.Body.String())
	}

	var out struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(secondRR.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if out.Code != "key_conflict" {
		t.Fatalf("expected error.code=key_conflict, got %q", out.Code)
	}
}

func TestCreateNextVersion_SystemOwnedTemplateImmutable(t *testing.T) {
	repo := newFakeRepo()
	templateID := "00000000-0000-0000-0000-000000000101"
	repo.templates[templateID] = &domain.Template{
		ID:          templateID,
		TenantID:    "tenant-a",
		SystemOwned: true,
	}
	repo.versions["ver-1"] = &domain.TemplateVersion{
		ID:            "ver-1",
		TemplateID:    templateID,
		VersionNumber: 1,
		Status:        domain.VersionStatusPublished,
	}
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/versions", nil)
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
	if out.Code != "SYSTEM_TEMPLATE_IMMUTABLE" {
		t.Fatalf("expected error.code=SYSTEM_TEMPLATE_IMMUTABLE, got %q", out.Code)
	}
}
