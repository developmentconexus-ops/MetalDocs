package http

import (
	taxonomyapi "metaldocs/internal/modules/taxonomy/api"
	"metaldocs/internal/modules/taxonomy/domain"
)

func toDocumentProfileItem(p *domain.DocumentProfile) taxonomyapi.DocumentProfileItem {
	if p == nil {
		return taxonomyapi.DocumentProfileItem{}
	}
	dto := taxonomyapi.DocumentProfileItem{
		Code:               string(p.Code),
		FamilyCode:         string(p.FamilyCode),
		Name:               p.Name,
		Description:        p.Description,
		ReviewIntervalDays: p.ReviewIntervalDays,
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
