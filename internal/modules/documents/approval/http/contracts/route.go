package contracts

import (
	"fmt"
	"regexp"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

type QuorumKind string

const (
	QuorumKindAny1Of QuorumKind = "any_1_of"
	QuorumKindAllOf  QuorumKind = "all_of"
	QuorumKindMOfN   QuorumKind = "m_of_n"
)

type DriftPolicyKind string

const (
	DriftPolicyKindReduceQuorum DriftPolicyKind = "reduce_quorum"
	DriftPolicyKindFailStage    DriftPolicyKind = "fail_stage"
	DriftPolicyKindKeepSnapshot DriftPolicyKind = "keep_snapshot"
)

var (
	routeCodePattern          = regexp.MustCompile(`^[a-z0-9_-]+$`)
	requiredCapabilityPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z0-9._]*[a-z0-9]$`)
)

type StageRequest struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
}

type CreateRouteRequest struct {
	ProfileCode    string         `json:"profile_code"`
	Name           string         `json:"name"`
	Stages         []StageRequest `json:"stages"`
	IdempotencyKey string
}

func (r CreateRouteRequest) Validate() error {
	if err := validateRequired("profile_code", r.ProfileCode); err != nil {
		return wrapValidation(err)
	}
	if err := validateRouteCode("profile_code", r.ProfileCode); err != nil {
		return wrapValidation(err)
	}
	if err := validateRequired("name", r.Name); err != nil {
		return wrapValidation(err)
	}
	return wrapValidation(validateStages(r.Stages))
}

type UpdateRouteRequest struct {
	Name           string         `json:"name"`
	Stages         []StageRequest `json:"stages"`
	IdempotencyKey string
}

func (r UpdateRouteRequest) Validate() error {
	if err := validateRequired("name", r.Name); err != nil {
		return wrapValidation(err)
	}
	return wrapValidation(validateStages(r.Stages))
}

func validateStages(stages []StageRequest) error {
	if len(stages) == 0 {
		return fmt.Errorf("stages must contain at least one stage")
	}
	seenNames := make(map[string]struct{}, len(stages))
	for i, stage := range stages {
		expectedOrder := i + 1
		if stage.Order != expectedOrder {
			return fmt.Errorf("stages[%d].order must be %d", i, expectedOrder)
		}
		if err := validateRequired(fmt.Sprintf("stages[%d].name", i), stage.Name); err != nil {
			return err
		}
		if _, exists := seenNames[stage.Name]; exists {
			return fmt.Errorf("stages[%d].name duplicates an earlier stage", i)
		}
		seenNames[stage.Name] = struct{}{}
		if err := validateRequired(fmt.Sprintf("stages[%d].required_role", i), stage.RequiredRole); err != nil {
			return err
		}
		if err := validateRouteCode(fmt.Sprintf("stages[%d].required_role", i), stage.RequiredRole); err != nil {
			return err
		}
		if err := validateAreaRole(fmt.Sprintf("stages[%d].required_role", i), stage.RequiredRole); err != nil {
			return err
		}
		if err := validateRequired(fmt.Sprintf("stages[%d].required_capability", i), stage.RequiredCapability); err != nil {
			return err
		}
		if err := validateRequiredCapability(fmt.Sprintf("stages[%d].required_capability", i), stage.RequiredCapability); err != nil {
			return err
		}
		if err := validateRequired(fmt.Sprintf("stages[%d].area_code", i), stage.AreaCode); err != nil {
			return err
		}
		if err := validateRouteCode(fmt.Sprintf("stages[%d].area_code", i), stage.AreaCode); err != nil {
			return err
		}
		switch stage.Quorum {
		case QuorumKindAny1Of, QuorumKindAllOf:
			if stage.QuorumM != nil {
				return fmt.Errorf("stages[%d].quorum_m must be omitted unless quorum is m_of_n", i)
			}
		case QuorumKindMOfN:
			if stage.QuorumM == nil {
				return fmt.Errorf("stages[%d].quorum_m is required when quorum is m_of_n", i)
			}
			if *stage.QuorumM < 1 {
				return fmt.Errorf("stages[%d].quorum_m must be >= 1", i)
			}
		default:
			return fmt.Errorf("stages[%d].quorum must be one of: any_1_of, all_of, m_of_n", i)
		}
		switch stage.DriftPolicy {
		case DriftPolicyKindReduceQuorum, DriftPolicyKindFailStage, DriftPolicyKindKeepSnapshot:
		default:
			return fmt.Errorf("stages[%d].drift_policy must be one of: reduce_quorum, fail_stage, keep_snapshot", i)
		}
	}
	return nil
}

func validateRouteCode(field, value string) error {
	if len(value) > 64 {
		return fmt.Errorf("%s must be at most 64 characters", field)
	}
	if !routeCodePattern.MatchString(value) {
		return fmt.Errorf("%s must match [a-z0-9_-]+", field)
	}
	return nil
}

// validateAreaRole binds a stage's required_role to the IAM role registry: it
// must be a canonical AREA role (a value a user can actually hold in
// user_process_areas, against which approval eligibility is resolved). Without
// this, required_role was free text — an admin could configure a stage requiring
// a role no user can ever hold (a phantom or a typo), producing a silently
// unsatisfiable stage (empty eligible pool → blocked approvals). The value has
// already passed the lowercase [a-z0-9_-]+ format check (ADR 0022 — role strings
// bound to the registry, not free text).
func validateAreaRole(field, value string) error {
	if iamdomain.IsAreaRole(iamdomain.Role(value)) {
		return nil
	}
	allowed := make([]string, 0, len(iamdomain.AreaRoles()))
	for _, r := range iamdomain.AreaRoles() {
		allowed = append(allowed, string(r))
	}
	return fmt.Errorf("%s %q is not a canonical area role (allowed: %s)", field, value, strings.Join(allowed, ", "))
}

func validateRequiredCapability(field, value string) error {
	if len(value) > 64 {
		return fmt.Errorf("%s must be at most 64 characters", field)
	}
	if !requiredCapabilityPattern.MatchString(value) {
		return fmt.Errorf("%s must match <namespace>.<action> using [a-z0-9_.]", field)
	}
	return nil
}

type RouteResponse struct {
	RouteID     string          `json:"route_id"`
	ProfileCode string          `json:"profile_code"`
	Name        string          `json:"name"`
	Version     int             `json:"version"`
	NewVersion  *int            `json:"new_version,omitempty"`
	Active      bool            `json:"active"`
	InUse       bool            `json:"in_use"`
	Stages      []StageResponse `json:"stages"`
	CreatedAt   string          `json:"created_at"`
}

type StageResponse struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
}

type ListStageItem struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
}

type ListRouteItem struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TenantID    string          `json:"tenant_id"`
	ProfileCode string          `json:"profile_code"`
	Active      bool            `json:"active"`
	Version     int             `json:"version"`
	Stages      []ListStageItem `json:"stages"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// ListRoutesResponse mirrors the OpenAPI `ListRoutesResponse` schema introduced
// in PR-1 (`api/openapi/v1/openapi.yaml` 5690-5698). The handler emits this
// typed envelope instead of an ad-hoc `map[string]any`.
type ListRoutesResponse struct {
	Routes []ListRouteItem `json:"routes"`
	Total  int             `json:"total"`
}

type DeactivateRouteRequest struct {
	Reason string `json:"reason"`
}

func (r DeactivateRouteRequest) Validate() error {
	if err := validateRequired("reason", strings.TrimSpace(r.Reason)); err != nil {
		return wrapValidation(err)
	}
	return nil
}
