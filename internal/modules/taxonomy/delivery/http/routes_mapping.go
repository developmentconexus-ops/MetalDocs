package http

import (
	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/domain"
)

// toDocumentProfileItem maps a profile to its wire DTO. Both readiness flags
// are supplied by the caller (from profileService.RouteReadySubjects) rather
// than read here: they are approval-owned state, not profile fields, and the
// list route resolves them for the whole page in one read per subject kind.
func toDocumentProfileItem(p *domain.DocumentProfile, hasActiveRoute, hasActiveTemplateRoute bool) taxonomyapi.DocumentProfileItem {
	if p == nil {
		return taxonomyapi.DocumentProfileItem{}
	}
	dto := taxonomyapi.DocumentProfileItem{
		Code:                   string(p.Code),
		FamilyCode:             string(p.FamilyCode),
		Name:                   p.Name,
		Description:            p.Description,
		ReviewIntervalDays:     p.ReviewIntervalDays,
		GovernanceClass:        taxonomyapi.DocumentProfileItemGovernanceClass(p.GovernanceClass),
		HasActiveRoute:         hasActiveRoute,
		HasActiveTemplateRoute: hasActiveTemplateRoute,
		// Fields not yet in domain.DocumentProfile — zero-fill (D-1 bounded defer)
		ActiveSchemaVersion: 0,
		ApprovalRequired:    false,
		RetentionDays:       0,
		ValidityDays:        0,
		WorkflowProfile:     "",
	}
	if p.Alias != "" {
		dto.Alias = &p.Alias
	}
	return dto
}
