package contracts

import (
	"fmt"
	"regexp"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// QuorumKind is the wire representation of a stage's quorum policy.
type QuorumKind string

// QuorumKind values.
const (
	QuorumKindAny1Of QuorumKind = "any_1_of"
	QuorumKindAllOf  QuorumKind = "all_of"
	QuorumKindMOfN   QuorumKind = "m_of_n"
)

// DriftPolicyKind is the wire representation of a stage's eligibility-drift policy.
type DriftPolicyKind string

// DriftPolicyKind values.
const (
	DriftPolicyKindReduceQuorum DriftPolicyKind = "reduce_quorum"
	DriftPolicyKindFailStage    DriftPolicyKind = "fail_stage"
	DriftPolicyKindKeepSnapshot DriftPolicyKind = "keep_snapshot"
)

var (
	routeCodePattern          = regexp.MustCompile(`^[a-z0-9_-]+$`)
	requiredCapabilityPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z0-9._]*[a-z0-9]$`)
)

// StageKind is the wire representation of a stage's kind (review vs approval).
type StageKind string

// StageKind values.
const (
	StageKindReview   StageKind = "review"
	StageKindApproval StageKind = "approval"
)

// SelectorKind is the wire representation of an ActorSelector's discriminator
// (M4 ActorSelector, unit 3.2). Mirrors domain.SelectorKind exactly.
type SelectorKind string

// SelectorKind values.
const (
	SelectorKindNamedUser          SelectorKind = "named_user"
	SelectorKindRoleInFixedArea    SelectorKind = "role_in_fixed_area"
	SelectorKindRoleInDocumentArea SelectorKind = "role_in_document_area"
	SelectorKindSubmitChoice       SelectorKind = "submit_choice"
)

// ActorSelector is the wire representation of one way a Stage may resolve its
// eligible actor pool (M4 ActorSelector, unit 3.2). Kind discriminates which
// of UserID/Role/AreaCode are meaningful; an empty field mirrors an absent
// value, matching domain.ActorSelector.
type ActorSelector struct {
	Kind     SelectorKind `json:"kind"`
	UserID   string       `json:"user_id,omitempty"`
	Role     string       `json:"role,omitempty"`
	AreaCode string       `json:"area_code,omitempty"`
}

// Validate mirrors domain.ActorSelector.Validate's exact field
// presence/absence contract for s.Kind:
//
//   - named_user: UserID set, Role and AreaCode absent.
//   - role_in_fixed_area: Role and AreaCode set, UserID absent.
//   - role_in_document_area: Role set, UserID and AreaCode absent.
//   - submit_choice: Role and AreaCode set, UserID absent.
//
// An unrecognized Kind or a recognized Kind with the wrong fields
// present/absent returns a descriptive error.
func (s ActorSelector) Validate() error {
	switch s.Kind {
	case SelectorKindNamedUser:
		if s.UserID != "" && s.Role == "" && s.AreaCode == "" {
			return nil
		}
	case SelectorKindRoleInFixedArea:
		if s.Role != "" && s.AreaCode != "" && s.UserID == "" {
			return nil
		}
	case SelectorKindRoleInDocumentArea:
		if s.Role != "" && s.AreaCode == "" && s.UserID == "" {
			return nil
		}
	case SelectorKindSubmitChoice:
		if s.Role != "" && s.AreaCode != "" && s.UserID == "" {
			return nil
		}
	default:
		return fmt.Errorf("kind must be one of: named_user, role_in_fixed_area, role_in_document_area, submit_choice")
	}
	return fmt.Errorf("fields invalid for kind %q", s.Kind)
}

// StageRequest is the wire representation of one stage in a create/update route request.
type StageRequest struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
	// StageKind (F0, spec.md §4/W2): review or approval. Optional on the wire;
	// empty defaults to approval at the persistence layer (route_admin_service,
	// migration 0286 DEFAULT 'approval') so pre-existing approval-only callers
	// are unchanged. A review-kind route is only creatable once this field can
	// be supplied.
	StageKind StageKind `json:"stage_kind,omitempty"`
	// Selectors (M4 ActorSelector, unit 3.2). Optional during the flat-column
	// coexistence window; when empty, RequiredRole/AreaCode still govern via
	// the domain EffectiveSelectors bridge.
	Selectors []ActorSelector `json:"selectors,omitempty"`
}

// CreateRouteRequest is the decoded body for the create-route endpoint.
type CreateRouteRequest struct {
	ProfileCode string         `json:"profile_code"`
	Name        string         `json:"name"`
	Stages      []StageRequest `json:"stages"`
	// SubjectKind and SubjectKey generalize what the route governs (M3 kernel
	// extraction, ADR 0082 / P2.S3). Both optional on the wire; profile_code
	// stays required as the backward-compat alias for the document case.
	// Omitting both preserves the legacy (document, profile_code) subject
	// byte-equal to pre-P2.S3 behavior — the application service applies that
	// default, not this contract layer.
	SubjectKind    string `json:"subject_kind,omitempty"`
	SubjectKey     string `json:"subject_key,omitempty"`
	IdempotencyKey string
}

// Validate enforces ProfileCode and Name are non-empty, ProfileCode matches the
// route-code pattern, SubjectKind (if supplied) is a known enum value, and
// Stages passes validateStages.
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
	switch r.SubjectKind {
	case "", "document", "template":
	default:
		return wrapValidation(fmt.Errorf("subject_kind must be one of: document, template"))
	}
	return wrapValidation(validateStages(r.Stages))
}

// UpdateRouteRequest is the decoded body for the update-route endpoint.
type UpdateRouteRequest struct {
	Name           string         `json:"name"`
	Stages         []StageRequest `json:"stages"`
	IdempotencyKey string
}

// Validate enforces Name is non-empty and Stages passes validateStages.
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
		// stage_kind is optional; empty defaults to approval downstream. When
		// supplied it must be one of the two canonical kinds (no free text).
		switch stage.StageKind {
		case "", StageKindReview, StageKindApproval:
		default:
			return fmt.Errorf("stages[%d].stage_kind must be one of: review, approval", i)
		}
		// Selectors are optional during the flat-column coexistence window
		// (M4, unit 3.2): an empty slice is not an error here — the
		// EffectiveSelectors bridge + domain Route.Validate cover the ≥1
		// selector rule downstream. When present, each selector must be
		// internally consistent for its kind.
		for j, sel := range stage.Selectors {
			if err := sel.Validate(); err != nil {
				return fmt.Errorf("stages[%d].selectors[%d]: %w", i, j, err)
			}
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

// validateRequiredCapability binds the stage's required_capability to the IAM
// capability registry — the same registry-binding ADR 0022 applied to
// required_role (validateAreaRole above). Without it the field was free text: an
// admin could configure a phantom/typo'd capability ("doc.signof") that
// validates and stores but matches nothing. NOTE: required_capability is
// currently advisory config — sign-off eligibility is resolved on required_role
// + area, and the enforced sign-off capability is the fixed CapDocumentSignoff
// (decision_service.go). Per-stage capability enforcement is not wired (would be
// a new ADR per ADR 0022 §scope). Binding here keeps the field honest until then.
func validateRequiredCapability(field, value string) error {
	if len(value) > 64 {
		return fmt.Errorf("%s must be at most 64 characters", field)
	}
	if !requiredCapabilityPattern.MatchString(value) {
		return fmt.Errorf("%s must match <namespace>.<action> using [a-z0-9_.]", field)
	}
	if !iamdomain.IsValidCapability(iamdomain.Capability(value)) {
		return fmt.Errorf("%s %q is not a registered capability", field, value)
	}
	return nil
}

// RouteResponse is the response body for a successful create/update/get route.
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

// StageResponse is the wire representation of one stage within a RouteResponse.
type StageResponse struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
	StageKind          StageKind       `json:"stage_kind"`
	// Selectors (M4 ActorSelector, unit 3.2). Optional during the flat-column
	// coexistence window.
	Selectors []ActorSelector `json:"selectors,omitempty"`
}

// ListStageItem is the wire representation of one stage within a ListRouteItem.
type ListStageItem struct {
	Order              int             `json:"order"`
	Name               string          `json:"name"`
	RequiredRole       string          `json:"required_role"`
	RequiredCapability string          `json:"required_capability"`
	AreaCode           string          `json:"area_code"`
	Quorum             QuorumKind      `json:"quorum"`
	QuorumM            *int            `json:"quorum_m,omitempty"`
	DriftPolicy        DriftPolicyKind `json:"drift_policy"`
	StageKind          StageKind       `json:"stage_kind"`
	// Selectors (M4 ActorSelector, unit 3.2). Optional during the flat-column
	// coexistence window.
	Selectors []ActorSelector `json:"selectors,omitempty"`
}

// ListRouteItem is one row of the list-routes response.
type ListRouteItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	TenantID    string `json:"tenant_id"`
	ProfileCode string `json:"profile_code"`
	// SubjectKind and SubjectKey generalize what this route governs (M3
	// kernel extraction, ADR 0082 / P2.S3). document/profile_code for every
	// route created via the legacy document path.
	SubjectKind string          `json:"subject_kind,omitempty"`
	SubjectKey  string          `json:"subject_key,omitempty"`
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

// DeactivateRouteRequest is the decoded body for the deactivate-route endpoint.
type DeactivateRouteRequest struct {
	Reason string `json:"reason"`
}

// Validate enforces Reason is non-empty after trimming whitespace.
func (r DeactivateRouteRequest) Validate() error {
	if err := validateRequired("reason", strings.TrimSpace(r.Reason)); err != nil {
		return wrapValidation(err)
	}
	return nil
}
