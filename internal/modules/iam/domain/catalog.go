package domain

import "sort"

// Catalog metadata for the IAM Admin Center "Roles & Capabilities" tab.
// Labels are pt-BR per project i18n convention. Categories partition the
// catalogue into operator-friendly groupings.

// RoleCategory partitions roles by scope: tenant-wide vs area-scoped.
type RoleCategory string

const (
	RoleCategoryTenant RoleCategory = "tenant"
	RoleCategoryArea   RoleCategory = "area"
)

// RoleDescriptor describes a canonical role for the Admin Center catalogue.
type RoleDescriptor struct {
	Code        Role
	Label       string
	Description string
	Category    RoleCategory
}

// canonicalRoles is the source of truth for the 8 canonical roles surfaced
// to the Admin Center. Order matches the operator-facing display order
// (tenant-wide first, then area-scoped).
var canonicalRoles = []RoleDescriptor{
	{
		Code:        RoleSystemAdmin,
		Label:       "System Admin",
		Description: "Autoridade administrativa total no tenant; ignora checagens de capacidade de tier-1 e tier-2.",
		Category:    RoleCategoryTenant,
	},
	{
		Code:        RoleApprover,
		Label:       "Aprovador",
		Description: "Aprova documentos controlados no fluxo de workflow.",
		Category:    RoleCategoryTenant,
	},
	{
		Code:        RoleAuthor,
		Label:       "Autor",
		Description: "Cria e elabora documentos.",
		Category:    RoleCategoryTenant,
	},
	{
		Code:        RoleEditor,
		Label:       "Editor",
		Description: "Edita documentos de sua autoria ou de sua área.",
		Category:    RoleCategoryTenant,
	},
	{
		Code:        RoleViewer,
		Label:       "Visualizador",
		Description: "Acesso somente leitura a documentos publicados.",
		Category:    RoleCategoryTenant,
	},
	{
		Code:        RoleSigner,
		Label:       "Assinante",
		Description: "Assina documentos dentro da área de processo concedida.",
		Category:    RoleCategoryArea,
	},
	{
		Code:        RoleAreaAdmin,
		Label:       "Administrador de Área",
		Description: "Gerencia usuários e políticas dentro da área de processo concedida.",
		Category:    RoleCategoryArea,
	},
	{
		Code:        RoleQmsAdmin,
		Label:       "Administrador QMS",
		Description: "Administrador de conformidade ISO 9001; leitura cross-area de auditoria e governança.",
		Category:    RoleCategoryArea,
	},
}

// CanonicalRoles returns a copy of the 8-role catalogue used by the Admin
// Center.
func CanonicalRoles() []RoleDescriptor {
	out := make([]RoleDescriptor, len(canonicalRoles))
	copy(out, canonicalRoles)
	return out
}

// CapabilityDescriptor describes a single Capability for the Admin Center
// catalogue. Category is derived from the code prefix at catalogue build time.
type CapabilityDescriptor struct {
	Code        Capability
	Description string
	Category    string
}

// RoleCapabilityLink is a single (role, capability) pair from the global
// role↔capability matrix.
type RoleCapabilityLink struct {
	Role       Role
	Capability Capability
}

// capabilityDescriptions maps every Capability const to its pt-BR description.
// CapabilityCatalog() panics if a const is added without a description here.
var capabilityDescriptions = map[Capability]string{
	CapDocumentView:                "Visualizar documentos",
	CapDocumentCreate:              "Criar documentos",
	CapDocumentEdit:                "Editar documentos",
	CapDocumentSubmit:              "Submeter documento para revisão",
	CapDocumentSignoff:             "Aprovar/recusar documento",
	CapDocumentSupersede:           "Tornar documento obsoleto/sucessor",
	CapWorkflowReview:              "Revisar workflow de aprovação",
	CapWorkflowApprove:             "Aprovar workflow",
	CapTemplateView:                "Visualizar templates",
	CapTemplateCreate:              "Criar templates",
	CapTemplateEdit:                "Editar templates",
	CapTemplateSubmit:              "Submeter template para revisão",
	CapTemplateReview:              "Revisar template",
	CapTemplateApprove:             "Aprovar template",
	CapTemplatePublish:             "Publicar template",
	CapControlledDocumentCreate:    "Criar documento controlado",
	CapControlledDocumentObsolete:  "Tornar documento controlado obsoleto",
	CapControlledDocumentSupersede: "Substituir documento controlado",
	CapTaxonomyView:                "Visualizar taxonomia",
	CapTaxonomyManage:              "Gerenciar taxonomia (famílias, perfis, áreas)",
	CapMembershipView:              "Visualizar memberships",
	CapMembershipManage:            "Gerenciar memberships de área",
	CapRouteManage:                 "Gerenciar rotas de aprovação",
	CapUserView:                    "Visualizar usuários",
	CapUserManage:                  "Gerenciar usuários",
	CapMetricsView:                 "Visualizar métricas e KPIs",
	CapAuditRead:                   "Visualizar trilha de auditoria",
	CapSessionManage:               "Gerenciar sessões (force-logout)",
}

// CapabilityCatalog returns one CapabilityDescriptor per Capability const,
// sorted by code for deterministic output. Category is derived from the
// prefix before the first '.' (e.g. "document.view" -> "document"); codes
// without a prefix fall back to "general". Panics if a Capability const lacks
// a description.
func CapabilityCatalog() []CapabilityDescriptor {
	out := make([]CapabilityDescriptor, 0, len(validCapabilities))
	for cap := range validCapabilities {
		desc, ok := capabilityDescriptions[cap]
		if !ok {
			panic("iam: capability missing description: " + string(cap))
		}
		out = append(out, CapabilityDescriptor{
			Code:        cap,
			Description: desc,
			Category:    capabilityCategory(cap),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func capabilityCategory(cap Capability) string {
	s := string(cap)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return s[:i]
		}
	}
	return "general"
}
