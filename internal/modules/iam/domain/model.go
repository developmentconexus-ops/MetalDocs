package domain

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleReviewer    Role = "reviewer"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

type Capability string

const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"
)
