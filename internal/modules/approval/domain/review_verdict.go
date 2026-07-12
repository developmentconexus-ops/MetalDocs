package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// Verdict represents a review-stage runtime verdict.
type Verdict string

// Verdict values a ReviewVerdict may carry.
const (
	// VerdictReady is the review-stage "approve equivalent": counted via
	// EvaluateQuorum exactly like a signoff approval.
	VerdictReady Verdict = "ready"
	// VerdictRequestChanges is the review-stage "reject equivalent": no
	// quorum needed, transitions the instance to InstanceChangesRequested
	// and the document back to draft. Requires a non-empty Comment.
	VerdictRequestChanges Verdict = "request_changes"
)

// ErrVerdictWrongStageKind is returned by NewVerdict when the target stage
// kind is neither StageKindReview nor StageKindApproval.
var ErrVerdictWrongStageKind = errors.New("review_verdict: stage kind must be 'review' or 'approval'")

// ErrVerdictCommentRequired is returned by NewVerdict when Verdict is
// VerdictRequestChanges and Comment is empty.
var ErrVerdictCommentRequired = errors.New("review_verdict: comment is required for request_changes")

// ErrVerdictReadyOnApprovalStage is returned by NewVerdict when a `ready`
// verdict targets an approval-kind stage. Approval stages accept only
// `request_changes` (R3/G2: an approver may return a document without
// signing, but `ready` would bypass the e-signature ceremony).
var ErrVerdictReadyOnApprovalStage = errors.New("review_verdict: 'ready' verdict is not allowed on an approval-kind stage")

// ReviewVerdict is an immutable value object recorded against a stage of the
// approval route (review- or approval-kind). Mirrors Signoff's shape
// (unexported fields, getters, no setters).
type ReviewVerdict struct {
	id                       string
	approvalInstanceID       string
	stageInstanceID          string
	actorUserID              string
	actorTenantID            string
	verdict                  Verdict
	comment                  string
	verdictAt                time.Time
	actorDisplayNameSnapshot string
	onBehalfOfUserID         string // F9/ADR 0077: delegator's user_id, "" when acting as self
}

// ID returns the verdict's identifier.
func (v *ReviewVerdict) ID() string { return v.id }

// ApprovalInstanceID returns the parent approval instance's identifier.
func (v *ReviewVerdict) ApprovalInstanceID() string { return v.approvalInstanceID }

// StageInstanceID returns the stage instance this verdict was recorded against.
func (v *ReviewVerdict) StageInstanceID() string { return v.stageInstanceID }

// ActorUserID returns the recording user's identifier.
func (v *ReviewVerdict) ActorUserID() string { return v.actorUserID }

// ActorTenantID returns the recording user's tenant identifier.
func (v *ReviewVerdict) ActorTenantID() string { return v.actorTenantID }

// Verdict returns the ready/request_changes value.
func (v *ReviewVerdict) Verdict() Verdict { return v.verdict }

// Comment returns the reviewer's comment (required for request_changes).
func (v *ReviewVerdict) Comment() string { return v.comment }

// VerdictAt returns the timestamp the verdict was recorded.
func (v *ReviewVerdict) VerdictAt() time.Time { return v.verdictAt }

// ActorDisplayNameSnapshot returns the actor's display name as captured at
// verdict time.
func (v *ReviewVerdict) ActorDisplayNameSnapshot() string { return v.actorDisplayNameSnapshot }

// OnBehalfOf returns the delegator's user_id when this verdict was recorded
// via an active delegation (F9/ADR 0077); "" when the actor acted as
// themselves.
func (v *ReviewVerdict) OnBehalfOf() string { return v.onBehalfOfUserID }

// VerdictParams holds constructor inputs for NewVerdict.
type VerdictParams struct {
	ID                       string
	ApprovalInstanceID       string
	StageInstanceID          string
	StageKind                StageKind
	ActorUserID              string
	ActorTenantID            string
	Verdict                  Verdict
	Comment                  string
	VerdictAt                time.Time
	ActorDisplayNameSnapshot string
	OnBehalfOfUserID         string // F9/ADR 0077: delegator's user_id, "" when acting as self
}

// NewVerdict constructs an immutable ReviewVerdict value object. Enforces:
// review-kind stages accept both ready and request_changes; approval-kind
// stages accept only request_changes (R3/G2) and reject ready via
// ErrVerdictReadyOnApprovalStage; any other stage kind is rejected via
// ErrVerdictWrongStageKind. request_changes requires a non-empty Comment;
// Verdict must be one of the two defined values.
func NewVerdict(p VerdictParams) (*ReviewVerdict, error) {
	if p.ID == "" {
		return nil, errors.New("review_verdict: ID is required")
	}
	if p.ApprovalInstanceID == "" {
		return nil, errors.New("review_verdict: ApprovalInstanceID is required")
	}
	if p.StageInstanceID == "" {
		return nil, errors.New("review_verdict: StageInstanceID is required")
	}
	if p.ActorUserID == "" {
		return nil, errors.New("review_verdict: ActorUserID is required")
	}
	if p.ActorTenantID == "" {
		return nil, errors.New("review_verdict: ActorTenantID is required")
	}
	if p.VerdictAt.IsZero() {
		return nil, errors.New("review_verdict: VerdictAt must not be zero")
	}
	switch p.StageKind {
	case StageKindReview:
		// review stages accept both ready and request_changes
	case StageKindApproval:
		if p.Verdict == VerdictReady {
			return nil, ErrVerdictReadyOnApprovalStage
		}
		// approval stages accept only request_changes (R3/G2)
	default:
		return nil, ErrVerdictWrongStageKind
	}
	switch p.Verdict {
	case VerdictReady:
	case VerdictRequestChanges:
		if p.Comment == "" {
			return nil, ErrVerdictCommentRequired
		}
	default:
		return nil, errors.New("review_verdict: Verdict must be 'ready' or 'request_changes'")
	}

	return &ReviewVerdict{
		id:                       p.ID,
		approvalInstanceID:       p.ApprovalInstanceID,
		stageInstanceID:          p.StageInstanceID,
		actorUserID:              p.ActorUserID,
		actorTenantID:            p.ActorTenantID,
		verdict:                  p.Verdict,
		comment:                  p.Comment,
		verdictAt:                p.VerdictAt,
		actorDisplayNameSnapshot: p.ActorDisplayNameSnapshot,
		onBehalfOfUserID:         p.OnBehalfOfUserID,
	}, nil
}

// MarshalJSON exposes ReviewVerdict for API responses.
func (v *ReviewVerdict) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"id":                          v.id,
		"approval_instance_id":        v.approvalInstanceID,
		"stage_instance_id":           v.stageInstanceID,
		"actor_user_id":               v.actorUserID,
		"actor_tenant_id":             v.actorTenantID,
		"verdict":                     v.verdict,
		"comment":                     v.comment,
		"verdict_at":                  v.verdictAt,
		"actor_display_name_snapshot": v.actorDisplayNameSnapshot,
		"on_behalf_of_user_id":        v.onBehalfOfUserID,
	})
}
