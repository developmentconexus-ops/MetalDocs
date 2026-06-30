package application

import "fmt"

// TenantTemplatePrefix is the object-store namespace under which every template
// object (docx + schema) for a tenant lives — the tenants/{tenantID}/templates/
// root that satisfies the VerifiedStore write-path guard. It is the SINGLE SOURCE
// OF TRUTH for that root: templateDocxKey / templateSchemaKey build on top of it,
// so a reconciliation sweeper that enumerates this prefix is guaranteed to scan
// exactly the namespace the writers use (the root cannot drift between them).
func TenantTemplatePrefix(tenantID string) string {
	return fmt.Sprintf("tenants/%s/templates/", tenantID)
}

// templateDocxKey builds the canonical tenant-scoped object key for a template
// version's docx, rooted at TenantTemplatePrefix.
func templateDocxKey(tenantID, templateID string, versionNumber int) string {
	return fmt.Sprintf("%s%s/versions/%d.docx", TenantTemplatePrefix(tenantID), templateID, versionNumber)
}

func templateSchemaKey(tenantID, templateID string, versionNumber int) string {
	return fmt.Sprintf("%s%s/versions/%d.schema.json", TenantTemplatePrefix(tenantID), templateID, versionNumber)
}

// TemplateVersionDocxKey exposes the canonical docx-object key for a template
// version. The stored docx_storage_key column normally carries this value; the
// exported builder lets tests and reconciliation tooling derive the same key
// through the single source of truth instead of duplicating the format string.
func TemplateVersionDocxKey(tenantID, templateID string, versionNumber int) string {
	return templateDocxKey(tenantID, templateID, versionNumber)
}

// TemplateVersionSchemaKey exposes the DERIVED schema-object key for a template
// version. The schema key is never stored in a DB column (unlike docx_storage_key);
// it is derived from the version coordinates by PresignTemplateSchemaUpload. The
// orphan-GC sweeper MUST treat every version's derived schema key as referenced,
// so it consumes this single-sourced builder rather than re-deriving the format.
func TemplateVersionSchemaKey(tenantID, templateID string, versionNumber int) string {
	return templateSchemaKey(tenantID, templateID, versionNumber)
}
