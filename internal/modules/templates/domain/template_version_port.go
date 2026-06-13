package domain

import "context"

// TemplateVersionPort is the read-only capability that other modules use to
// verify template version state without depending on taxonomy infrastructure.
type TemplateVersionPort interface {
	// IsPublished reports whether the given template version is published and
	// returns the doc_type_code of the owning template.
	IsPublished(ctx context.Context, versionID string) (bool, string, error)
}
