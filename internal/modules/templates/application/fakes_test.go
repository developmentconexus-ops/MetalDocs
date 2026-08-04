package application_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/objectstore"
)

// newPermissiveMockDB returns a *sql.DB backed by sqlmock configured to accept
// any SQL without strict matching.  Use it when a test exercises business logic
// that requires a DB (Begin/authz-GUC/Commit) but does not need to assert on
// the exact SQL emitted.  The mock is seeded with enough Begin/Exec/Query/Commit
// expectations (out-of-order) to cover all templates service operations;
// leftover expectations are silently dropped when the DB is closed at test end.
func newPermissiveMockDB(t *testing.T) *sql.DB {
	t.Helper()
	anyMatcher := sqlmock.QueryMatcherFunc(func(_, _ string) error { return nil })
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(anyMatcher))
	if err != nil {
		t.Fatalf("newPermissiveMockDB: sqlmock.New: %v", err)
	}
	mock.MatchExpectationsInOrder(false)
	t.Cleanup(func() { _ = mockDB.Close() })

	// Seed more expectations than any single operation needs.
	// Out-of-order matching ensures any call sequence is satisfied.
	for i := 0; i < 5; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	for i := 0; i < 30; i++ {
		mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	// Queries must return at least one column with a value that scans into the
	// expected types.  The authz sequence reads:
	//   1. actor_id GUC  (string)
	//   2. tenant_id GUC (string)
	//   3. system_admin EXISTS (bool → true so the check passes without cap query)
	//   4. asserted_caps GUC  (string)
	for i := 0; i < 5; i++ {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("user-a"))
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow("tenant-a"))
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(""))
	}
	return mockDB
}

type fakeRepo struct {
	templates      map[string]*domain.Template
	versions       map[string]*domain.TemplateVersion
	audit          []*domain.AuditEvent
	receivedFilter application.ListFilter

	ignoreTenantOnGetTemplate bool
	lockVersions              map[string]int
	getTemplateByKeyErr       error

	// ADR 0088 blank-materialization source overrides.
	systemBlankVersion *domain.TemplateVersion
	systemBlankErr     error

	// UpdateTemplateTxCalls records the LatestVersion on each UpdateTemplateTx
	// call.  Tests that assert F-T5 (single template write per tx) inspect this.
	UpdateTemplateTxCalls []int
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

// readOf wraps a stored write-aggregate into the read model the Repository
// port now returns (ADR 0065). The Latest ref is derived from the aggregate's
// scalar version pointer so promoted-field access stays truthful.
func readOf(t *domain.Template) *domain.TemplateRead {
	return &domain.TemplateRead{
		Template: *t,
		Latest: domain.VersionRef{
			Number: t.LatestVersion,
		},
	}
}

func (r *fakeRepo) GetTemplate(_ context.Context, tenantID, id string) (*domain.TemplateRead, error) {
	t, ok := r.templates[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if !r.ignoreTenantOnGetTemplate && t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return readOf(t), nil
}

func (r *fakeRepo) GetTemplateByKey(_ context.Context, tenantID, key string) (*domain.TemplateRead, error) {
	if r.getTemplateByKeyErr != nil {
		return nil, r.getTemplateByKeyErr
	}
	for _, t := range r.templates {
		if t.TenantID == tenantID && t.Key == key {
			return readOf(t), nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) ListTemplates(_ context.Context, f application.ListFilter) ([]*domain.TemplateRead, error) {
	r.receivedFilter = f
	out := make([]*domain.TemplateRead, 0, len(r.templates))
	for _, t := range r.templates {
		if f.TenantID != "" && t.TenantID != f.TenantID {
			continue
		}
		if f.PublishedOnly && t.PublishedVersionID == nil {
			continue
		}
		out = append(out, readOf(t))
	}
	if len(out) == 0 {
		return nil, domain.ErrNotFound
	}
	return out, nil
}

func (r *fakeRepo) UpdateTemplate(_ context.Context, t *domain.Template) error {
	if _, ok := r.templates[t.ID]; !ok {
		return domain.ErrNotFound
	}
	clone := *t
	r.templates[t.ID] = &clone
	return nil
}

func (r *fakeRepo) CreateVersion(_ context.Context, v *domain.TemplateVersion) error {
	if _, ok := r.lockVersions[v.ID]; !ok {
		r.lockVersions[v.ID] = v.LockVersion
	}
	clone := *v
	r.versions[v.ID] = &clone
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
			clone := *v
			return &clone, nil
		}
	}
	return nil, domain.ErrNotFound
}

// fakeSystemBlankHash stands in for the reference-data-pinned sha256 of
// deploy/assets/system-blank.docx (64 hex chars, the only shape the ADR 0088
// materialization path accepts).
const fakeSystemBlankHash = "5cdae1bb25103bbc121cdc696ed11eb09aa22041940f199164ebc302f6923d2e"

// GetSystemBlankVersion mirrors the reference-data row by default so every
// pre-existing CreateTemplate test keeps working unchanged. Tests that exercise
// the fail-closed arms set systemBlankErr or systemBlankVersion explicitly.
func (r *fakeRepo) GetSystemBlankVersion(_ context.Context) (*domain.TemplateVersion, error) {
	if r.systemBlankErr != nil {
		return nil, r.systemBlankErr
	}
	if r.systemBlankVersion != nil {
		clone := *r.systemBlankVersion
		return &clone, nil
	}
	return &domain.TemplateVersion{
		ID:             "00000000-0000-0000-0000-000000000102",
		TenantID:       "ffffffff-ffff-ffff-ffff-ffffffffffff",
		TemplateID:     "00000000-0000-0000-0000-000000000101",
		VersionNumber:  1,
		Status:         domain.VersionStatusPublished,
		DocxStorageKey: "system/templates/blank.docx",
		ContentHash:    fakeSystemBlankHash,
	}, nil
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
	clone := *v
	return &clone, nil
}

func (r *fakeRepo) UpdateVersion(_ context.Context, _ string, v *domain.TemplateVersion) error {
	if _, ok := r.versions[v.ID]; !ok {
		return domain.ErrNotFound
	}
	current := r.lockVersions[v.ID]
	if current != v.LockVersion {
		return domain.ErrStaleLockVersion
	}
	clone := *v
	clone.LockVersion++
	r.lockVersions[v.ID] = clone.LockVersion
	r.versions[v.ID] = &clone
	v.LockVersion = clone.LockVersion
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

func (r *fakeRepo) CreateTemplateTx(_ context.Context, _ db.Tx, t *domain.Template) error {
	return r.CreateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateTemplateTx(_ context.Context, _ db.Tx, t *domain.Template) error {
	r.UpdateTemplateTxCalls = append(r.UpdateTemplateTxCalls, t.LatestVersion)
	return r.UpdateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateVersionTx(_ context.Context, _ db.Tx, tenantID string, v *domain.TemplateVersion) error {
	return r.UpdateVersion(context.Background(), tenantID, v)
}

func (r *fakeRepo) ObsoletePreviousPublished(_ context.Context, _ /*tenantID*/ string, templateID, keepVersionID string) error {
	found := false
	for _, v := range r.versions {
		if v.TemplateID != templateID {
			continue
		}
		found = true
		if v.Status == domain.VersionStatusPublished && v.ID != keepVersionID {
			now := time.Now().UTC()
			v.ObsoletedAt = &now
		}
	}
	if !found {
		return domain.ErrNotFound
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
	matched := make([]*domain.AuditEvent, 0, len(r.audit))
	for _, e := range r.audit {
		if e.TenantID == tenantID && e.TemplateID == templateID {
			matched = append(matched, e)
		}
	}
	if len(matched) == 0 {
		return nil, domain.ErrNotFound
	}
	if offset >= len(matched) {
		return nil, domain.ErrNotFound
	}
	end := len(matched)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return matched[offset:end], nil
}

type fakePresigner struct {
	PutKeys    []string
	CopyPairs  [][2]string
	getCalls   []string
	confirmErr error
	copyErr    error
	// existsResult is a pointer so the zero-value fakePresigner{} (used
	// pervasively across this package's other tests) defaults Exists to
	// true — those tests build versions with a non-empty DocxStorageKey and
	// expect a presigned URL, matching pre-Exists-gate behavior. Tests that
	// exercise the "object not yet uploaded" path set it explicitly.
	existsResult *bool
	existsErr    error
}

func (f *fakePresigner) PresignPut(_ context.Context, _ string, key string, _ time.Duration) (string, error) {
	f.PutKeys = append(f.PutKeys, key)
	return "https://example/put", nil
}

func (f *fakePresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	f.getCalls = append(f.getCalls, key)
	return "https://example/get/" + key, nil
}

func (f *fakePresigner) Exists(_ context.Context, _ string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	if f.existsResult != nil {
		return *f.existsResult, nil
	}
	return true, nil
}

func (f *fakePresigner) Confirm(_ context.Context, _, key, expected string) (objectstore.VerifiedPointer, error) {
	if f.confirmErr != nil {
		return objectstore.VerifiedPointer{}, f.confirmErr
	}
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: expected, SizeBytes: 1}, nil
}

func (f *fakePresigner) Copy(_ context.Context, _ string, srcKey, dstKey string) error {
	if f.copyErr != nil {
		return f.copyErr
	}
	f.CopyPairs = append(f.CopyPairs, [2]string{srcKey, dstKey})
	return nil
}

func (f *fakePresigner) Delete(_ context.Context, _ string) error { return nil }

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

type fakeResolverRegistry struct {
	known map[string]int
}

func (r fakeResolverRegistry) Known() map[string]int {
	return r.known
}

type serviceOption func(*serviceOptions)

type serviceOptions struct {
	resolvers application.ResolverRegistryReader
}

func WithKnownResolvers(keys ...string) serviceOption {
	return func(opts *serviceOptions) {
		known := make(map[string]int, len(keys))
		for i, key := range keys {
			known[key] = i + 1
		}
		opts.resolvers = fakeResolverRegistry{known: known}
	}
}

func newService(repo *fakeRepo, opts ...serviceOption) *application.Service {
	options := serviceOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	if options.resolvers != nil {
		return application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{}, options.resolvers)
	}
	return application.New(repo, &fakePresigner{}, fakeClock{}, &fakeUUID{})
}

func updateCmdWithComputed(phID, resolverKey string) application.UpdateSchemasCmd {
	return application.UpdateSchemasCmd{
		TenantID:      "tenant-a",
		TemplateID:    "tpl-1",
		VersionNumber: 1,
		MetadataSchema: domain.MetadataSchema{
			RetentionDays: 1,
		},
		PlaceholderSchema: []domain.Placeholder{
			{
				ID:          phID,
				Type:        domain.PHComputed,
				Computed:    true,
				ResolverKey: &resolverKey,
			},
		},
	}
}
