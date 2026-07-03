package domain

import (
	"errors"
	"slices"
	"strings"
)

// VisibilityScope controls who can read a controlled document beyond its
// owner: the whole company, or a restricted set of areas/users.
type VisibilityScope string

// VisibilityScope enum values.
const (
	VisibilityScopeCompany    VisibilityScope = "company"
	VisibilityScopeRestricted VisibilityScope = "restricted"
)

// Visibility is the read-access scope attached to a ControlledDocument.
// AreaCodes/UserIDs are only meaningful when Scope is
// VisibilityScopeRestricted; NewVisibility normalizes both to empty slices
// for VisibilityScopeCompany.
type Visibility struct {
	Scope     VisibilityScope `json:"scope"`
	AreaCodes []string        `json:"area_codes"`
	UserIDs   []string        `json:"user_ids"`
}

// ErrVisibilityScopeInvalid is returned by NewVisibility when scope is
// neither "company" nor "restricted" (case-insensitive).
var (
	ErrVisibilityScopeInvalid = errors.New("visibility scope is invalid")
)

// NewVisibility normalizes scope (defaulting empty to restricted) and
// dedupes/trims areaCodes and userIDs. For a restricted scope with no
// explicit area codes, it falls back to defaultAreaCode (the creating
// controlled document's own area) so restricted-by-default documents are
// never visible to zero areas.
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
