package application

import "fmt"

// templateDocxKey builds the canonical tenant-scoped object key for a template
// version's docx. The tenants/{tenantID}/ prefix satisfies the VerifiedStore
// write-path guard.
func templateDocxKey(tenantID, templateID string, versionNumber int) string {
	return fmt.Sprintf("tenants/%s/templates/%s/versions/%d.docx", tenantID, templateID, versionNumber)
}

// TemplateSchemaKey builds the canonical tenant-scoped object key for a template
// version's JSON schema. Exported so the HTTP handler can compute the key before
// calling PresignTemplateUpload (the handler owns tenantID at call time).
func TemplateSchemaKey(tenantID, templateID string, versionNumber int) string {
	return fmt.Sprintf("tenants/%s/templates/%s/versions/%d.schema.json", tenantID, templateID, versionNumber)
}
