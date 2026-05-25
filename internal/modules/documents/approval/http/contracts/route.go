package contracts

import "fmt"

const slugMaxLen = 64

type QuorumKind string

const (
	QuorumKindAny1Of QuorumKind = "any_1_of"
	QuorumKindAllOf  QuorumKind = "all_of"
	QuorumKindMOfN   QuorumKind = "m_of_n"
)

type DriftPolicyKind string

const (
	DriftPolicyReduceQuorum DriftPolicyKind = "reduce_quorum"
	DriftPolicyFailStage    DriftPolicyKind = "fail_stage"
	DriftPolicyKeepSnapshot DriftPolicyKind = "keep_snapshot"
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
	if err := validateSlug("profile_code", r.ProfileCode, slugMaxLen); err != nil {
		return err
	}
	if err := validateRequired("name", r.Name); err != nil {
		return err
	}
	return validateStages(r.Stages)
}

type UpdateRouteRequest struct {
	Name           string         `json:"name"`
	Stages         []StageRequest `json:"stages"`
	IdempotencyKey string
}

func (r UpdateRouteRequest) Validate() error {
	if err := validateRequired("name", r.Name); err != nil {
		return err
	}
	return validateStages(r.Stages)
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
		if err := validateSlug(fmt.Sprintf("stages[%d].required_role", i), stage.RequiredRole, slugMaxLen); err != nil {
			return err
		}
		if err := validateRequired(fmt.Sprintf("stages[%d].required_capability", i), stage.RequiredCapability); err != nil {
			return err
		}
		if err := validateSlug(fmt.Sprintf("stages[%d].area_code", i), stage.AreaCode, slugMaxLen); err != nil {
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
		case DriftPolicyReduceQuorum, DriftPolicyFailStage, DriftPolicyKeepSnapshot:
		default:
			return fmt.Errorf("stages[%d].drift_policy must be one of: reduce_quorum, fail_stage, keep_snapshot", i)
		}
	}
	return nil
}

type RouteResponse struct {
	RouteID     string          `json:"route_id"`
	ProfileCode string          `json:"profile_code"`
	Name        string          `json:"name"`
	Version     int             `json:"version"`
	NewVersion  int             `json:"new_version,omitempty"`
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
	Label              string          `json:"label"`
	Members            []string        `json:"members"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	QuorumKind         QuorumKind      `json:"quorum_kind"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
}

type ListRouteItem struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TenantID    string          `json:"tenant_id"`
	ProfileCode string          `json:"profile_code"`
	Active      bool            `json:"active"`
	Stages      []ListStageItem `json:"stages"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}
