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
// version's docx object.
//
// ADR 0088: the old rationale here — "blank templates assign a storage key at
// create time and lazily provision the object on first autosave confirm" — is
// FALSE as of this ADR. Every version materializes its object PRE-TX at
// creation, so the editor never opens onto a missing object and a fresh blank
// template resolves a URL immediately.
//
// The Exists gate is KEPT and still returns ErrUploadMissing: it now guards
// only genuine store-side loss (an object deleted or never restored out of
// band, or a pre-0317 legacy row) — the sentinel's honest residual meaning. It
// is a read-path integrity check that fails closed, not a lifecycle gate.
func (s *Service) GetDocxURL(ctx context.Context, cmd GetDocxURLCmd) (string, error) {
	if _, err := s.GetTemplate(ctx, cmd.TenantID, cmd.TemplateID); err != nil {
		return "", err
	}
	v, err := s.repo.GetVersion(ctx, cmd.TenantID, cmd.TemplateID, cmd.VersionNumber)
	if err != nil {
		return "", err
	}
	// The `DocxStorageKey == ""` branch that used to sit here is DELETED
	// (legacy-extermination): docx_storage_key is NOT NULL in the schema and
	// every writer sets it from templateDocxKey, so the branch was dead even
	// before this ADR — and an empty key now cannot survive creation at all,
	// since materialization copies to that very key. Exists() below is the
	// real, reachable guard; an empty key would fail it anyway rather than
	// silently presigning garbage.
	exists, err := s.presign.Exists(ctx, v.DocxStorageKey)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", domain.ErrUploadMissing
	}
	return s.presign.PresignGet(ctx, v.DocxStorageKey, docxDownloadTTL)
}
