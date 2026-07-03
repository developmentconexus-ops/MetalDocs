package application

import (
	"context"
	"time"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/objectstore"
)

// Repository defines the persistence port required by the application
// layer: CRUD and CAS-guarded updates for templates and template versions,
// approval configuration, and audit trail access. Implementations must
// provide both a plain and a Tx-suffixed variant of each mutating method so
// callers can compose multi-step writes inside a single caller-owned
// transaction (via db.Tx) when a use case requires atomicity across several
// repository calls.
type Repository interface {
	CreateTemplate(ctx context.Context, t *domain.Template) error
	CreateTemplateTx(ctx context.Context, tx db.Tx, t *domain.Template) error
	GetTemplate(ctx context.Context, tenantID, id string) (*domain.TemplateRead, error)
	GetTemplateByKey(ctx context.Context, tenantID, key string) (*domain.TemplateRead, error)
	ListTemplates(ctx context.Context, f ListFilter) ([]*domain.TemplateRead, error)
	UpdateTemplate(ctx context.Context, t *domain.Template) error
	UpdateTemplateTx(ctx context.Context, tx db.Tx, t *domain.Template) error

	CreateVersion(ctx context.Context, v *domain.TemplateVersion) error
	CreateVersionTx(ctx context.Context, tx db.Tx, v *domain.TemplateVersion) error
	GetVersion(ctx context.Context, tenantID, templateID string, n int) (*domain.TemplateVersion, error)
	GetVersionByID(ctx context.Context, tenantID, id string) (*domain.TemplateVersion, error)
	UpdateVersion(ctx context.Context, tenantID string, v *domain.TemplateVersion) error
	UpdateVersionTx(ctx context.Context, tx db.Tx, tenantID string, v *domain.TemplateVersion) error
	UpdateVersionSchemaCAS(ctx context.Context, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error
	UpdateVersionSchemaCASTx(ctx context.Context, tx db.Tx, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error
	ObsoletePreviousPublished(ctx context.Context, tenantID, templateID, keepVersionID string) error
	ObsoletePreviousPublishedTx(ctx context.Context, tx db.Tx, tenantID, templateID, keepVersionID string) error

	GetApprovalConfig(ctx context.Context, tenantID, templateID string) (*domain.ApprovalConfig, error)
	UpsertApprovalConfig(ctx context.Context, c *domain.ApprovalConfig) error
	UpsertApprovalConfigTx(ctx context.Context, tx db.Tx, c *domain.ApprovalConfig) error

	AppendAudit(ctx context.Context, e *domain.AuditEvent) error
	AppendAuditTx(ctx context.Context, tx db.Tx, e *domain.AuditEvent) error
	ListAudit(ctx context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error)
}

// Presigner defines the object-store port used to issue presigned
// upload/download URLs for template docx and schema objects, confirm
// completed uploads against an expected content hash, copy objects when
// spawning a new draft version, and delete objects.
type Presigner interface {
	PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Confirm(ctx context.Context, tenantID, key, expectedHash string) (objectstore.VerifiedPointer, error)
	Copy(ctx context.Context, tenantID, srcKey, dstKey string) error
	Delete(ctx context.Context, key string) error
}

// Clock abstracts the current time so the application layer can be tested
// deterministically.
type Clock interface{ Now() time.Time }

// UUIDGen abstracts identifier generation so the application layer can be
// tested deterministically.
type UUIDGen interface{ New() string }

// ResolverRegistryReader exposes the set of known native/computed
// placeholder resolver keys, used to validate placeholder schemas without
// the application layer depending on the render module's concrete registry.
type ResolverRegistryReader interface{ Known() map[string]int }

// ListFilter narrows a ListTemplates query by tenant, document type, and
// status, with pagination via Limit/Offset.
type ListFilter struct {
	TenantID    string
	DocTypeCode *string
	Status      *domain.VersionStatus
	// PublishedOnly, when true, restricts the result to templates that
	// currently have a published version (published_version_id IS NOT
	// NULL) — the selectable-to-clone set for the new-document wizard.
	// False (default) keeps the management view: templates in every
	// lifecycle status.
	PublishedOnly bool
	Limit         int
	Offset        int
}
