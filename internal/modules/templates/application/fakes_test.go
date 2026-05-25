package application_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
)

type fakeRepo struct {
	templates       map[string]*domain.Template
	versions        map[string]*domain.TemplateVersion
	audit           []*domain.AuditEvent
	approvalConfigs map[string]*domain.ApprovalConfig
	receivedFilter  application.ListFilter

	ignoreTenantOnGetTemplate bool
	lockVersions              map[string]int
	failCreateVersion         bool
	getTemplateByKeyErr       error
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
	if !ok {
		return nil, domain.ErrNotFound
	}
	if !r.ignoreTenantOnGetTemplate && t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) GetTemplateByKey(_ context.Context, tenantID, key string) (*domain.Template, error) {
	if r.getTemplateByKeyErr != nil {
		return nil, r.getTemplateByKeyErr
	}
	for _, t := range r.templates {
		if t.TenantID == tenantID && t.Key == key {
			return t, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeRepo) ListTemplates(_ context.Context, f application.ListFilter) ([]*domain.Template, error) {
	r.receivedFilter = f
	out := make([]*domain.Template, 0, len(r.templates))
	for _, t := range r.templates {
		if f.TenantID != "" && t.TenantID != f.TenantID {
			continue
		}
		out = append(out, t)
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
	r.templates[t.ID] = t
	return nil
}

func (r *fakeRepo) CreateVersion(_ context.Context, v *domain.TemplateVersion) error {
	if r.failCreateVersion {
		return errors.New("forced create version failure")
	}
	if _, ok := r.lockVersions[v.ID]; !ok {
		r.lockVersions[v.ID] = 0
	}
	r.versions[v.ID] = v
	return nil
}

func (r *fakeRepo) CreateVersionTx(_ context.Context, _ *sql.Tx, v *domain.TemplateVersion) error {
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

func (r *fakeRepo) UpdateVersion(_ context.Context, _ string, v *domain.TemplateVersion) error {
	if _, ok := r.versions[v.ID]; !ok {
		return domain.ErrNotFound
	}
	r.versions[v.ID] = v
	return nil
}

func (r *fakeRepo) UpdateVersionDraftCAS(_ context.Context, _ string, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
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

func (r *fakeRepo) UpdateVersionDraftCASTx(_ context.Context, _ *sql.Tx, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	return r.UpdateVersionDraftCAS(context.Background(), tenantID, versionID, expectedLockVersion, docxStorageKey, docxContentHash)
}

func (r *fakeRepo) CreateTemplateTx(_ context.Context, _ *sql.Tx, t *domain.Template) error {
	return r.CreateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateTemplateTx(_ context.Context, _ *sql.Tx, t *domain.Template) error {
	return r.UpdateTemplate(context.Background(), t)
}

func (r *fakeRepo) UpdateVersionTx(_ context.Context, _ *sql.Tx, tenantID string, v *domain.TemplateVersion) error {
	return r.UpdateVersion(context.Background(), tenantID, v)
}

func (r *fakeRepo) ObsoletePreviousPublished(_ context.Context, templateID, keepVersionID string) error {
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

func (r *fakeRepo) ObsoletePreviousPublishedTx(_ context.Context, _ *sql.Tx, templateID, keepVersionID string) error {
	return r.ObsoletePreviousPublished(context.Background(), templateID, keepVersionID)
}

func (r *fakeRepo) GetApprovalConfig(_ context.Context, tenantID, templateID string) (*domain.ApprovalConfig, error) {
	t, ok := r.templates[templateID]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	c, ok := r.approvalConfigs[templateID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c, nil
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
	HeadResult   string
	HeadErr      error
	DeleteErr    error
	DeleteCalled int
	PutKeys      []string
}

func (p *fakePresigner) PresignPUT(_ context.Context, key string, _ time.Duration) (string, error) {
	p.PutKeys = append(p.PutKeys, key)
	return "https://presigned/put/" + key, nil
}

func (p *fakePresigner) PresignGET(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://presigned/get/" + key, nil
}

func (p *fakePresigner) HeadContentHash(_ context.Context, _ string) (string, error) {
	if p.HeadResult == "" {
		return "hash_abc", p.HeadErr
	}
	return p.HeadResult, p.HeadErr
}

func (p *fakePresigner) Delete(_ context.Context, _ string) error {
	p.DeleteCalled++
	return p.DeleteErr
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
