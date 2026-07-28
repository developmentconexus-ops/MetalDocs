package http

import (
	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/domain"
)

// toDocumentProfileItem maps a profile to its wire DTO. hasActiveRoute is
// supplied by the caller (from profileService.RouteReadySubjects) rather than
// read here: it is approval-owned state, not a profile field, and the list route
// resolves it for the whole page in one read.
func toDocumentProfileItem(p *domain.DocumentProfile, hasActiveRoute bool) taxonomyapi.DocumentProfileItem {
	if p == nil {
		return taxonomyapi.DocumentProfileItem{}
	}
	dto := taxonomyapi.DocumentProfileItem{
		Code:               string(p.Code),
		FamilyCode:         string(p.FamilyCode),
		Name:               p.Name,
		Description:        p.Description,
		ReviewIntervalDays: p.ReviewIntervalDays,
		GovernanceClass:    taxonomyapi.DocumentProfileItemGovernanceClass(p.GovernanceClass),
		HasActiveRoute:     hasActiveRoute,
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
