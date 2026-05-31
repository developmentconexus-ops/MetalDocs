package domain

import (
	"fmt"
	"strings"
)

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

var validRoles = map[Role]struct{}{
	RoleApprover:    {},
	RoleAuthor:      {},
	RoleEditor:      {},
	RoleSystemAdmin: {},
	RoleViewer:      {},
}

func IsValidRole(role Role) bool {
	_, ok := validRoles[role]
	return ok
}

func ParseRole(raw string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(raw)))
	if role == "" || !IsValidRole(role) {
		return "", ErrInvalidRole
	}
	return role, nil
}

type Capability string

const (
	CapDocumentView      Capability = "document.view"
	CapDocumentCreate    Capability = "document.create"
	CapDocumentEdit      Capability = "document.edit"
	CapDocumentSubmit    Capability = "document.submit"
	CapDocumentSignoff   Capability = "document.signoff"
	CapDocumentSupersede Capability = "doc.supersede"
	CapWorkflowReview    Capability = "workflow.review"
	CapWorkflowApprove   Capability = "workflow.approve"

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

var validCapabilities = map[Capability]struct{}{
	CapDocumentView:                {},
	CapDocumentCreate:              {},
	CapDocumentEdit:                {},
	CapDocumentSubmit:              {},
	CapDocumentSignoff:             {},
	CapDocumentSupersede:           {},
	CapWorkflowReview:              {},
	CapWorkflowApprove:             {},
	CapTemplateView:                {},
	CapTemplateCreate:              {},
	CapTemplateEdit:                {},
	CapTemplateSubmit:              {},
	CapTemplateReview:              {},
	CapTemplateApprove:             {},
	CapTemplatePublish:             {},
	CapControlledDocumentCreate:    {},
	CapControlledDocumentObsolete:  {},
	CapControlledDocumentSupersede: {},
	CapTaxonomyManage:              {},
	CapMembershipManage:            {},
	CapRouteManage:                 {},
	CapUserManage:                  {},
	CapAuditRead:                   {},
}

func IsValidCapability(cap Capability) bool {
	_, ok := validCapabilities[cap]
	return ok
}

// AllCapabilities returns every known capability. Used by callers that need
// to express a "holds everything" set (e.g. system_admin bypass).
func AllCapabilities() []Capability {
	out := make([]Capability, 0, len(validCapabilities))
	for cap := range validCapabilities {
		out = append(out, cap)
	}
	return out
}

func MustCapability(raw string) Capability {
	capability := Capability(raw)
	if !IsValidCapability(capability) {
		panic(fmt.Sprintf("invalid capability: %q", raw))
	}
	return capability
}
