package application

import (
	"context"
	"time"

	"metaldocs/internal/modules/templates/domain"
)

// GetTemplate fetches a template by ID, scoped to tenantID. A template
// belonging to a different tenant is treated as not found (cross-tenant
// lookups never leak existence).
func (s *Service) GetTemplate(ctx context.Context, tenantID, id string) (*domain.TemplateRead, error) {
	t, err := s.repo.GetTemplate(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

// GetVersion fetches template version n, scoped to tenantID, after
// confirming the parent template exists and belongs to the tenant.
func (s *Service) GetVersion(ctx context.Context, tenantID, templateID string, n int) (*domain.TemplateVersion, error) {
	if _, err := s.GetTemplate(ctx, tenantID, templateID); err != nil {
		return nil, err
	}
	return s.repo.GetVersion(ctx, tenantID, templateID, n)
}

// ListTemplates returns templates matching the given filter.
func (s *Service) ListTemplates(ctx context.Context, f ListFilter) ([]*domain.TemplateRead, error) {
	return s.repo.ListTemplates(ctx, f)
}

// ListAudit returns the paginated audit trail for a template, after
// confirming the template exists and belongs to the tenant.
func (s *Service) ListAudit(ctx context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	if _, err := s.GetTemplate(ctx, tenantID, templateID); err != nil {
		return nil, err
	}
	return s.repo.ListAudit(ctx, tenantID, templateID, limit, offset)
}

const docxDownloadTTL = 15 * time.Minute

// GetDocxURLCmd identifies the template version to generate a docx download
// URL for.
type GetDocxURLCmd struct {
	TenantID, TemplateID string
	VersionNumber        int
}

// GetDocxURL returns a time-limited presigned GET URL for a template
// version's docx object. Returns ErrUploadMissing if the version has no
// docx object recorded yet.
func (s *Service) GetDocxURL(ctx context.Context, cmd GetDocxURLCmd) (string, error) {
	if _, err := s.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return "", err
	}
	v, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return "", err
	}
	if v.DocxStorageKey == "" {
		return "", domain.ErrUploadMissing
	}
	return s.presign.PresignGet(ctx, v.DocxStorageKey, docxDownloadTTL)
}
