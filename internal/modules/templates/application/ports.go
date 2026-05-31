package application

import (
	"context"
	"database/sql"
	"time"

	"metaldocs/internal/modules/templates/domain"
)

type Repository interface {
	CreateTemplate(ctx context.Context, t *domain.Template) error
	CreateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error
	GetTemplate(ctx context.Context, tenantID, id string) (*domain.Template, error)
	GetTemplateByKey(ctx context.Context, tenantID, key string) (*domain.Template, error)
	ListTemplates(ctx context.Context, f ListFilter) ([]*domain.Template, error)
	UpdateTemplate(ctx context.Context, t *domain.Template) error
	UpdateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error

	CreateVersion(ctx context.Context, v *domain.TemplateVersion) error
	CreateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error
	GetVersion(ctx context.Context, tenantID, templateID string, n int) (*domain.TemplateVersion, error)
	GetVersionByID(ctx context.Context, tenantID, id string) (*domain.TemplateVersion, error)
	UpdateVersion(ctx context.Context, tenantID string, v *domain.TemplateVersion) error
	UpdateVersionTx(ctx context.Context, tx *sql.Tx, tenantID string, v *domain.TemplateVersion) error
	UpdateVersionDraftCAS(ctx context.Context, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error
	UpdateVersionDraftCASTx(ctx context.Context, tx *sql.Tx, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error
	UpdateVersionSchemaCAS(ctx context.Context, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error
	UpdateVersionSchemaCASTx(ctx context.Context, tx *sql.Tx, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error
	ObsoletePreviousPublished(ctx context.Context, templateID, keepVersionID string) error
	ObsoletePreviousPublishedTx(ctx context.Context, tx *sql.Tx, templateID, keepVersionID string) error

	GetApprovalConfig(ctx context.Context, tenantID, templateID string) (*domain.ApprovalConfig, error)
	UpsertApprovalConfig(ctx context.Context, c *domain.ApprovalConfig) error
	UpsertApprovalConfigTx(ctx context.Context, tx *sql.Tx, c *domain.ApprovalConfig) error

	AppendAudit(ctx context.Context, e *domain.AuditEvent) error
	AppendAuditTx(ctx context.Context, tx *sql.Tx, e *domain.AuditEvent) error
	ListAudit(ctx context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error)
}

type Presigner interface {
	PresignPUT(ctx context.Context, key string, expires time.Duration) (url string, err error)
	PresignGET(ctx context.Context, key string, expires time.Duration) (url string, err error)
	HeadContentHash(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type Clock interface{ Now() time.Time }
type UUIDGen interface{ New() string }
type ResolverRegistryReader interface{ Known() map[string]int }

type ListFilter struct {
	TenantID    string
	DocTypeCode *string
	Status      *domain.VersionStatus
	Limit       int
	Offset      int
}
