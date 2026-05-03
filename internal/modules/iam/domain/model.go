package domain

type Role string

const (
	RoleAdmin       Role = "admin" // keep — removed in Task 8
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author" // new
	RoleEditor      Role = "editor"
	RoleReviewer    Role = "reviewer"     // keep — removed in Task 8
	RoleSystemAdmin Role = "system_admin" // new
	RoleViewer      Role = "viewer"
)

type Permission string

const (
	PermDocumentCreate            Permission = "document:create"
	PermDocumentEdit              Permission = "document:edit"
	PermDocumentRead              Permission = "document:read"
	PermDocumentUploadAttachment  Permission = "document:upload_attachment"
	PermDocumentManagePermissions Permission = "document:manage_permissions"
	PermVersionRead               Permission = "version:read"
	PermWorkflowTransition        Permission = "workflow:transition"
	PermSearchRead                Permission = "search:read"
	PermIAMManageRoles            Permission = "iam:manage_roles"
	PermTemplateView              Permission = "template:view"
	PermTemplateEdit              Permission = "template:edit"
	PermTemplatePublish           Permission = "template:publish"
	PermTemplateExport            Permission = "template:export"
)

type Capability string

const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"
)
