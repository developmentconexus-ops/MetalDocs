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

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/templates/application"
	tmplhttp "metaldocs/internal/modules/templates/delivery/http"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/tenant"
)

// newPermissiveMockDB creates a *sql.DB backed by sqlmock that accepts any SQL
// and returns suitable values for the two SQL layers that fire on every
// write-path in the templates HTTP handler:
//
//  1. Idempotency middleware (routes wrapped in h.idempotent):
//     BeginTx → QueryRowContext(INSERT…RETURNING, 5 args) → [handler runs] →
//     Exec(UPDATE idempotency_keys) → Commit
//
//  2. authz.Require inside the service:
//     BeginTx → 2×Exec(setAuthzGUC) →
//     QueryRowContext(actor_id, 0 args) → QueryRowContext(tenant_id, 0 args) →
//     QueryRowContext(system_admin EXISTS, 2 args) →
//     QueryRowContext(asserted_caps, 0 args) → Exec(set asserted_caps) → Commit
//
// With MatchExpectationsInOrder(false) sqlmock picks the first unfulfilled
// expectation whose arg count matches via attemptArgMatch.  We exploit this to
// separate the three query shapes so each pool is consumed independently:
//
//   - 5-arg pool → idempotency INSERT…RETURNING (returns a non-empty string so
//     the middleware treats the request as the "winner")
//   - 2-arg pool → system_admin EXISTS (returns bool true)
//   - 0-arg pool (WithoutArgs) → GUC reads: actor_id, tenant_id, asserted_caps
//     (returned in cycles of three; true/true/"" so every 3rd is empty)
//
// Used exclusively in delivery/http tests that verify routing and marshalling,
// not authz or idempotency SQL fidelity.
func newPermissiveMockDB(t *testing.T) *sql.DB {
	t.Helper()
	anyMatcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { return nil })
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(anyMatcher))
	if err != nil {
		t.Fatalf("newPermissiveMockDB: sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	t.Cleanup(func() { _ = mockDB.Close() })

	// Begin / Commit / Exec pools (untyped; ordered=false picks first available).
	for i := 0; i < 20; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	for i := 0; i < 100; i++ {
		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	}

	// 5-arg query pool: idempotency INSERT…RETURNING.
	// Returns a non-empty string so the middleware sees itself as the "winner".
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
	}

	// 2-arg query pool: system_admin EXISTS check → bool true.
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}

	// 0-arg query pool: actor_id, tenant_id, asserted_caps GUC reads.
	// Each Require call consumes three 0-arg queries in this order:
	//   1. actor_id  → "user-a"  (non-empty, passes MustActorID check)
	//   2. tenant_id → "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" (non-empty, passes MustTenantID check)
	//   3. asserted_caps → "" (empty, skips json.Unmarshal in loadAssertedCaps)
	// 10 cycles covers up to 10 Require calls per test.
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("user-a"))
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"))
		mock.ExpectQuery("").WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	}
	return mockDB
}

type fakeRepo struct {
	templates    map[string]*domain.Template
	versions     map[string]*domain.TemplateVersion
	audit        []*domain.AuditEvent
	lockVersions map[string]int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		templates:    map[string]*domain.Template{},
		versions:     map[string]*domain.TemplateVersion{},
		audit:        []*domain.AuditEvent{},
		lockVersions: map[string]int{},
	}
}

func (r *fakeRepo) CreateTemplate(_ context.Context, t *domain.Template) error {
	r.templates[t.ID] = t
	return nil
}

// syntheticVersionUUID derives a deterministic, valid UUID for a template's
// version ref when no real *domain.TemplateVersion row exists in r.versions
// to source an id from. Deterministic (not random) so assertions across a
// single test stay stable; namespaced by template id + role so latest and
// published never collide.
func syntheticVersionUUID(templateID, role string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(templateID+"|"+role)).String()
}

// readOf wraps a write-aggregate *domain.Template into the read model
// returned by the repository post-ADR-0065. It resolves Latest (and
// Published, when PublishedVersionID is set) against r.versions when a
// matching version row exists, so wire fields like latest_version.id carry
// real, self-consistent data; otherwise it falls back to a deterministic
// synthesized UUID so toAPIVersionRef's uuid.Parse never fails.
func (r *fakeRepo) readOf(t *domain.Template) *domain.TemplateRead {
	out := &domain.TemplateRead{Template: *t}

	out.Latest = domain.VersionRef{
		ID:     syntheticVersionUUID(t.ID, "latest"),
		Number: t.LatestVersion,
	}
	for _, v := range r.versions {
		if v.TemplateID == t.ID && v.VersionNumber == t.LatestVersion {
			out.Latest = domain.VersionRef{
				ID:             v.ID,
				Number:         v.VersionNumber,
				RevisionNumber: v.RevisionNumber,
				Status:         v.Status,
			}
			break
		}
	}

	if t.PublishedVersionID != nil {
		ref := domain.VersionRef{
			ID: *t.PublishedVersionID,
		}
		if v, ok := r.versions[*t.PublishedVersionID]; ok {
			ref.Number = v.VersionNumber
			ref.RevisionNumber = v.RevisionNumber
			ref.Status = v.Status
		}
		if ref.ID == "" {
			ref.ID = syntheticVersionUUID(t.ID, "published")
		}
		out.Published = &ref
	}

	return out
}

func (r *fakeRepo) GetTemplate(_ context.Context, tenantID, id string) (*domain.TemplateRead, error) {
	t, ok := r.templates[id]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return r.readOf(t), nil
}

func (r *fakeRepo) GetTemplateByKey(_ context.Context, tenantID, key string) (*domain.TemplateRead, error) {
	for _, t := range r.templates {
		if t.TenantID == tenantID && t.Key == key {
			return r.readOf(t), nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) ListTemplates(_ context.Context, f application.ListFilter) ([]*domain.TemplateRead, error) {
	out := make([]*domain.TemplateRead, 0, len(r.templates))
	for _, t := range r.templates {
		if t.TenantID != f.TenantID {
			continue
		}
		if f.PublishedOnly && t.PublishedVersionID == nil {
			continue
		}
		out = append(out, r.readOf(t))
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

func (r *fakeRepo) CreateVersionTx(_ context.Context, _ db.Tx, v *domain.TemplateVersion) error {
	return r.CreateVersion(context.Background(), v)
}

func (r *fakeRepo) GetVersion(_ context.Context, tenantID, templateID string, n int) (*domain.TemplateVersion, error) {
	t, ok := r.templates[templateID]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	for _, v := range r.versions {
		if v.TemplateID == templateID && v.VersionNumber == n {
			return v, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) GetVersionByID(_ context.Context, tenantID, id string) (*domain.TemplateVersion, error) {
	v, ok := r.versions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	t, ok := r.templates[v.TemplateID]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return v, nil
}

func (r *fakeRepo) CreateTemplateTx(_ context.Context, _ db.Tx, t *domain.Template) error {
	return r.CreateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateTemplateTx(_ context.Context, _ db.Tx, t *domain.Template) error {
	return r.UpdateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateVersionTx(_ context.Context, _ db.Tx, tenantID string, v *domain.TemplateVersion) error {
	return r.UpdateVersion(context.Background(), tenantID, v)
}

func (r *fakeRepo) UpdateVersion(_ context.Context, _ string, v *domain.TemplateVersion) error {
	if _, ok := r.versions[v.ID]; !ok {
		return domain.ErrNotFound
	}
	r.versions[v.ID] = v
	return nil
}

func (r *fakeRepo) UpdateVersionSchemaCAS(_ context.Context, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error {
	stored, ok := r.versions[v.ID]
	if !ok {
		return domain.ErrNotFound
	}
	t, ok := r.templates[stored.TemplateID]
	if !ok || t.TenantID != tenantID {
		return domain.ErrNotFound
	}
	current := r.lockVersions[v.ID]
	if current != expectedLockVersion {
		return domain.ErrStaleLockVersion
	}
	stored.MetadataSchema = v.MetadataSchema
	stored.PlaceholderSchema = v.PlaceholderSchema
	r.lockVersions[v.ID] = current + 1
	v.LockVersion = current + 1
	return nil
}

func (r *fakeRepo) UpdateVersionSchemaCASTx(_ context.Context, _ db.Tx, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error {
	return r.UpdateVersionSchemaCAS(context.Background(), tenantID, v, expectedLockVersion)
}

func (r *fakeRepo) ObsoletePreviousPublished(_ context.Context, _ string, templateID, keepVersionID string) error {
	for _, v := range r.versions {
		if v.TemplateID == templateID && v.Status == domain.VersionStatusPublished && v.ID != keepVersionID {
			now := time.Now().UTC()
			v.ObsoletedAt = &now
			v.Status = domain.VersionStatusObsolete
		}
	}
	return nil
}

func (r *fakeRepo) ObsoletePreviousPublishedTx(_ context.Context, _ db.Tx, tenantID, templateID, keepVersionID string) error {
	return r.ObsoletePreviousPublished(context.Background(), tenantID, templateID, keepVersionID)
}

func (r *fakeRepo) AppendAudit(_ context.Context, e *domain.AuditEvent) error {
	r.audit = append(r.audit, e)
	return nil
}

func (r *fakeRepo) AppendAuditTx(_ context.Context, _ db.Tx, e *domain.AuditEvent) error {
	return r.AppendAudit(context.Background(), e)
}

func (r *fakeRepo) ListAudit(_ context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	_ = limit
	_ = offset
	out := make([]*domain.AuditEvent, 0, len(r.audit))
	for _, e := range r.audit {
		if e.TenantID == tenantID && e.TemplateID == templateID {
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
	// Deterministic, valid UUIDv4 shape — required by F1.2 mappers (toAPIVersionDTO,
	// toAPITemplateDTO) which uuid.Parse the field. Counter populates the last 12 hex chars.
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", u.counter)
}

type fakePresigner struct{}

func (fakePresigner) PresignPut(_ context.Context, _ string, key string, _ time.Duration) (string, error) {
	return "https://presigned/put/" + key, nil
}

func (fakePresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://presigned/get/" + key, nil
}

func (fakePresigner) Confirm(_ context.Context, _, key, expected string) (objectstore.VerifiedPointer, error) {
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: expected, SizeBytes: 1}, nil
}

func (fakePresigner) Copy(_ context.Context, _, _, _ string) error { return nil }

func (fakePresigner) Delete(_ context.Context, _ string) error { return nil }

func newMux(t *testing.T, authz tmplhttp.AuthzFunc, repo *fakeRepo) *http.ServeMux {
	t.Helper()
	return newMuxWithPresigner(t, authz, repo, fakePresigner{})
}

func newMuxWithPresigner(t *testing.T, authz tmplhttp.AuthzFunc, repo *fakeRepo, p application.Presigner) *http.ServeMux {
	t.Helper()
	db := newPermissiveMockDB(t)
	svc := application.New(repo, p, fakeClock{}, &fakeUUID{}).WithRunner(newTxRunner(db))
	h := tmplhttp.New(svc, authz, db)
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

// mismatchPresigner always returns ErrHashMismatch from Confirm.
type mismatchPresigner struct{ fakePresigner }

func (mismatchPresigner) Confirm(_ context.Context, _, _ string, _ string) (objectstore.VerifiedPointer, error) {
	return objectstore.VerifiedPointer{}, objectstore.ErrHashMismatch
}

func TestNew_PanicsWithoutAuthz(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected New to panic when authz is nil")
		}
	}()
	_ = tmplhttp.New(application.New(newFakeRepo(), fakePresigner{}, fakeClock{}, &fakeUUID{}), nil, nil)
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
	ctx := tenant.WithTenantID(req.Context(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	ctx = iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	*req = *req.WithContext(ctx)
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

	if gotTenant != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" || gotArea != "*" || gotAction != "template.create" {
		t.Fatalf("unexpected authz call: tenant=%q area=%q action=%q", gotTenant, gotArea, gotAction)
	}

	var out struct {
		Data struct {
			Template map[string]any `json:"template"`
			Version  struct {
				VersionNumber int `json:"version_number"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	// F1.3: top-level undeclared fields must be absent.
	var rawTop map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &rawTop); err != nil {
		t.Fatalf("decode raw top-level: %v", err)
	}
	if _, ok := rawTop["id"]; ok {
		t.Error("top-level 'id' must not be present after F1.3 (A3/H-D)")
	}
	if _, ok := rawTop["version_id"]; ok {
		t.Error("top-level 'version_id' must not be present after F1.3 (A3/H-D)")
	}
	if out.Data.Template["id"] == "" {
		t.Fatal("expected template.id to be present")
	}
	if _, ok := out.Data.Template["visibility"]; ok {
		t.Fatal("expected template response to omit visibility")
	}
	if _, ok := out.Data.Template["areas"]; ok {
		t.Fatal("expected template response to omit areas")
	}
	if _, ok := out.Data.Template["specific_areas"]; ok {
		t.Fatal("expected template response to omit specific_areas")
	}
	if out.Data.Version.VersionNumber != 1 {
		t.Fatalf("expected version.version_number=1, got %d", out.Data.Version.VersionNumber)
	}
	// TemplateDTO uses omitempty — nil published_version_number is omitted, not null.
	if pvnField, ok := out.Data.Template["published_version_number"]; ok && pvnField != nil {
		t.Fatalf("expected published_version_number absent or null on freshly created template, got %v", pvnField)
	}
}

func TestCreateTemplate_RejectUnknownField(t *testing.T) {
	repo := newFakeRepo()
	mux := newMux(t, func(_ *http.Request, _, _, _ string) error { return nil }, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewBufferString(`{"key":"contract-default","name":"Contract Template","visibility":"public"}`))
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
	if out.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected error.code=VALIDATION_ERROR, got %q", out.Code)
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
	if out.Code != "ALREADY_EXISTS" {
		t.Fatalf("expected error.code=ALREADY_EXISTS, got %q", out.Code)
	}
}

func TestCreateNextVersion_SystemOwnedTemplateImmutable(t *testing.T) {
	repo := newFakeRepo()
	templateID := "00000000-0000-0000-0000-000000000101"
	repo.templates[templateID] = &domain.Template{
		ID:          templateID,
		TenantID:    "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
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
