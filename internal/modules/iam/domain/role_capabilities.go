package domain

const RoleCapabilitiesVersion = 2

var RoleCapabilities = map[Role][]Capability{
	RoleSystemAdmin: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
		CapTemplatePublish,
		CapWorkflowApprove,
		CapRegistryCreate,
	},
	RoleAuthor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
	},
	RoleViewer: {
		CapDocumentView,
		CapTemplateView,
	},
	RoleEditor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
	},
	RoleReviewer: {
		CapDocumentView,
		CapDocumentEdit,
		CapWorkflowReview,
		CapTemplateView,
	},
	RoleApprover: {
		CapDocumentView,
		CapWorkflowApprove,
		CapTemplateView,
		CapTemplatePublish,
	},
}
