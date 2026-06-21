package domain

import "context"

// TemplateVersionPort is the read-only capability that other modules use to
// read template version state without reaching into templates' owned tables.
type TemplateVersionPort interface {
	// IsPublished reports whether the given template version is published and
	// returns the doc_type_code of the owning template. Tenant is taken from ctx.
	IsPublished(ctx context.Context, versionID string) (bool, string, error)

	// GetTemplateVersionState returns the raw status and owning-template
	// doc_type_code for a version, scoped to tenantID. status is nil when the
	// version is not found or its status column is NULL; not-found is
	// (nil, "", nil). This is the primitive consumers needing the raw status
	// (e.g. controlled-documents override resolution) depend on; IsPublished is
	// the derived published-or-not predicate.
	GetTemplateVersionState(ctx context.Context, tenantID, versionID string) (*string, string, error)

	// PlaceholderSchema returns templates_template_version.placeholder_schema as
	// raw JSON for versionID, scoped to tenantID via the owning template. Returns
	// (nil, nil) when the version is absent or the column is NULL — the caller
	// treats nil as "no schema". Lets documents read fill-in schema (M2/F2.4;
	// ADR-0039 D3(b)) without joining templates_template_version to its own table.
	PlaceholderSchema(ctx context.Context, tenantID, versionID string) ([]byte, error)
}
