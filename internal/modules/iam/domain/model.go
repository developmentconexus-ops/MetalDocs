package domain

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

type Capability string

const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapDocumentSubmit  Capability = "document.submit"
	CapDocumentSignoff Capability = "document.signoff"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"

	CapTemplateView    Capability = "template.view"
	CapTemplateCreate  Capability = "template.create"
	CapTemplateEdit    Capability = "template.edit"
	CapTemplateSubmit  Capability = "template.submit"
	CapTemplateReview  Capability = "template.review"
	CapTemplateApprove Capability = "template.approve"
	CapTemplatePublish Capability = "template.publish"

	CapControlledDocumentCreate    Capability = "controlled_documents.create"
	CapControlledDocumentObsolete  Capability = "controlled_documents.obsolete"
	CapControlledDocumentSupersede Capability = "controlled_documents.supersede"
	CapTaxonomyManage              Capability = "taxonomy.manage"
	CapMembershipManage            Capability = "membership.manage"
	CapRouteManage                 Capability = "route.manage"
	CapUserManage                  Capability = "user.manage"
	CapAuditRead                   Capability = "audit.read"
)
