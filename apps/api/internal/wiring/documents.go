package wiring

import (
	"context"

	docsapp "metaldocs/internal/modules/documents/application"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// capabilityServiceAdapter bridges *iamapp.CapabilityService (takes string capability)
// to docsapp.CapabilityChecker (takes iamdomain.Capability). The types are both
// string-backed but Go's type system requires an explicit conversion.
type capabilityServiceAdapter struct {
	svc *iamapp.CapabilityService
}

func (a capabilityServiceAdapter) CanDo(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error {
	return a.svc.CanDo(ctx, userID, tenantID, string(cap))
}

// DocumentsDeps bundles the upstream services the documents module consumes.
// Add fields here when documents.Dependencies grows.
type DocumentsDeps struct {
	Repo               docsapp.Repository
	Docgen             docsapp.DocgenRenderer
	Presigner          docsapp.Presigner
	Templates          docsapp.TemplateReader
	FormValidator      docsapp.FormValidator
	Audit              docsapp.Audit
	Registry           docsapp.RegistryReader
	RegistryDuplicator docsapp.RegistryDuplicator
	ProfileTemplates   docsapp.ProfileDefaultTemplateReader
	Caps               *iamapp.CapabilityService
}

// WireDocuments builds the documents module Dependencies struct, binding the
// CapabilityChecker port to the production CapabilityService.
func WireDocuments(d DocumentsDeps) docsapp.CapabilityChecker {
	return capabilityServiceAdapter{svc: d.Caps}
}
