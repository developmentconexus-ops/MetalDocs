package application

import "context"

// NewServiceForTest returns a minimal *Service with only profileTemplates wired.
// Used by service_cd_test.go to exercise resolveTemplateVersionID without a full
// dependency graph.
func NewServiceForTest(profileTemplates ProfileDefaultTemplateReader) *Service {
	return &Service{profileTemplates: profileTemplates}
}

// ExportedResolveTemplateVersionID exposes the unexported resolveTemplateVersionID
// method for white-box testing from the application_test package.
func (s *Service) ExportedResolveTemplateVersionID(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	return s.resolveTemplateVersionID(ctx, tenantID, profileCode, templateVersionID)
}
