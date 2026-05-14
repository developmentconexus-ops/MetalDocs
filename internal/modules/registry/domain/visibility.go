package domain

import (
	"errors"
	"slices"
	"strings"
)

type VisibilityScope string

const (
	VisibilityScopeCompany    VisibilityScope = "company"
	VisibilityScopeRestricted VisibilityScope = "restricted"
)

type Visibility struct {
	Scope     VisibilityScope `json:"scope"`
	AreaCodes []string        `json:"areaCodes"`
	UserIDs   []string        `json:"userIds"`
}

var (
	ErrVisibilityScopeInvalid = errors.New("visibility scope is invalid")
)

func NewVisibility(scope string, areaCodes, userIDs []string, defaultAreaCode string) (Visibility, error) {
	normalizedScope := VisibilityScope(strings.TrimSpace(strings.ToLower(scope)))
	if normalizedScope == "" {
		normalizedScope = VisibilityScopeRestricted
	}
	if normalizedScope != VisibilityScopeCompany && normalizedScope != VisibilityScopeRestricted {
		return Visibility{}, ErrVisibilityScopeInvalid
	}

	if normalizedScope == VisibilityScopeCompany {
		return Visibility{Scope: VisibilityScopeCompany, AreaCodes: []string{}, UserIDs: []string{}}, nil
	}

	normAreas := uniqueNonEmpty(areaCodes)
	if len(normAreas) == 0 {
		fallback := strings.TrimSpace(defaultAreaCode)
		if fallback != "" {
			normAreas = []string{fallback}
		}
	}

	return Visibility{
		Scope:     VisibilityScopeRestricted,
		AreaCodes: normAreas,
		UserIDs:   uniqueNonEmpty(userIDs),
	}, nil
}

func uniqueNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if v == "" || slices.Contains(out, v) {
			continue
		}
		out = append(out, v)
	}
	return out
}
