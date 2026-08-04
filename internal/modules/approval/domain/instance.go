package domain

import (
	"errors"
	"time"
)

var (
	// ErrNoActiveStage is returned by stage-transition methods when the instance has no stage in StageActive status.
	ErrNoActiveStage = errors.New("no active stage in instance")
	// ErrRevisionRegression is returned by BumpRevisionVersion when next is lower than the current RevisionVersion.
	ErrRevisionRegression = errors.New("revision_version cannot decrease")
	// ErrInstanceTerminal is returned by Cancel when the instance Status is already approved, rejected, or cancelled.
	ErrInstanceTerminal = errors.New("instance is already in a terminal state")
	// ErrAutoApproveHasStages is returned by AutoApprove when the instance
	// carries stages. Auto-approval is the consequence of a ZERO-stage route
	// (ADR 0087); an instance with stages must be approved through its stages,
	// never short-circuited.
	ErrAutoApproveHasStages = errors.New("auto-approve requires a stageless instance")
)

// InstanceStatus represents the top-level lifecycle of an approval instance.
type InstanceStatus string

// InstanceStatus values. InstanceInProgress and InstanceChangesRequested are
// the only non-terminal statuses; the other three are terminal and reject
// further stage transitions.
const (
	InstanceInProgress InstanceStatus = "in_progress"
	InstanceApproved   InstanceStatus = "approved"
	InstanceRejected   InstanceStatus = "rejected"
	InstanceCancelled  InstanceStatus = "cancelled"
	// InstanceChangesRequested (F4, design spec §2) is reached when a
	// review-kind stage records a request_changes verdict: the author must
	// revise the document (which reverts to draft) and resubmit. Non-terminal
	// — Cancel remains callable from this status, same as InstanceInProgress.
	InstanceChangesRequested InstanceStatus = "changes_requested"
)

// StageStatus represents per-stage lifecycle.
type StageStatus string

// StageStatus values. Exactly one stage in an instance may be StageActive at a time.
const (
	StagePending      StageStatus = "pending"
	StageActive       StageStatus = "active"
	StageCompleted    StageStatus = "completed"
	StageSkipped      StageStatus = "skipped"
	StageRejectedHere StageStatus = "rejected_here"
)

// StageInstance holds runtime state for one approval stage.
type StageInstance struct {
	ID                         string
	ApprovalInstanceID         string
	StageOrder                 int
	NameSnapshot               string
	RequiredRoleSnapshot       string
	RequiredCapabilitySnapshot string
	AreaCodeSnapshot           string
	QuorumSnapshot             QuorumPolicy
	QuorumMSnapshot            *int
	OnEligibilityDriftSnapshot DriftPolicy
	Kind                       StageKind
	EligibleActorIDs           []string
	EffectiveDenominator       *int
	Status                     StageStatus
	OpenedAt                   *time.Time
	CompletedAt                *time.Time
	SkipReason                 string
	DueAt                      *time.Time
	// DueInDaysSnapshot (F8, spec.md §4/W4) snapshots the route stage's
	// due_in_days config at submit time, mirroring every other *Snapshot
	// field on this struct. The repository's activation UPDATE reads this
	// column (not approval_route_stages) to compute DueAt, so a due date
	// stays pinned to the config the instance started with even if the
	// route is later re-versioned (F2) before this stage activates.
	DueInDaysSnapshot *int
	// SelectorsSnapshot (M4, unit 3.2 slice 6 / Option C′) freezes the stage's
	// actor selectors onto the instance at submit, with submit_choice already
	// materialized into concrete named_user selectors. It is the SOLE source the
	// drift path (decision/review) re-resolves eligibility from — pool kinds
	// (role_in_fixed_area, role_in_document_area) re-resolve live; named_user
	// (exempt) ids are unioned unconditionally. Mirrors the self-contained-freeze
	// intent of the other *Snapshot fields (see DueInDaysSnapshot).
	SelectorsSnapshot []ActorSelector
	Signoffs          []*Signoff
}

// Instance is the approval instance aggregate.
type Instance struct {
	ID         string
	TenantID   string
	DocumentID string
	// Subject generalizes what this instance governs (M3 kernel extraction,
	// ADR 0082 / P2.S2). For every instance constructed by the current
	// document code path, Subject == {SubjectKindDocument, DocumentID} by
	// construction — DocumentID remains the canonical field for existing
	// consumers, Subject is the field the repository persists to
	// subject_kind/subject_key.
	Subject              Subject
	RouteID              string
	RouteVersionSnapshot int
	Status               InstanceStatus
	SubmittedBy          string
	SubmittedAt          time.Time
	CompletedAt          *time.Time
	ContentHashAtSubmit  string
	IdempotencyKey       string
	RevisionVersion      int
	FrozenContentHash    *string
	CancelReason         *string
	Stages               []StageInstance
}

// Active returns the current active StageInstance or nil.
func (inst *Instance) Active() *StageInstance {
	for i := range inst.Stages {
		if inst.Stages[i].Status == StageActive {
			return &inst.Stages[i]
		}
	}
	return nil
}

// AdvanceStage moves the active stage to completed and activates the next pending stage.
// When the last stage completes, Status=InstanceApproved.
func (inst *Instance) AdvanceStage() error {
	activeIdx := -1
	for i, s := range inst.Stages {
		if s.Status == StageActive {
			activeIdx = i
			break
		}
	}
	if activeIdx == -1 {
		return ErrNoActiveStage
	}

	now := time.Now().UTC()
	inst.Stages[activeIdx].Status = StageCompleted
	inst.Stages[activeIdx].CompletedAt = &now

	// Activate next pending stage.
	for i := activeIdx + 1; i < len(inst.Stages); i++ {
		if inst.Stages[i].Status == StagePending {
			inst.Stages[i].Status = StageActive
			inst.Stages[i].OpenedAt = &now
			return nil
		}
	}

	// No more pending — instance approved.
	inst.Status = InstanceApproved
	inst.CompletedAt = &now
	return nil
}

// AutoApprove completes a STAGELESS instance at creation time (ADR 0087): a
// route bound to a livre profile carries zero stages, so there is nothing to
// satisfy and the instance is approved in the very transaction that created it.
//
// It is deliberately NOT a special case inside AdvanceStage — AdvanceStage's
// contract is "the active stage completed", and a stageless instance never has
// one. Keeping the two apart is what stops "no stages" from being silently
// readable as "all stages done" on an instance that merely lost its stages.
//
// Fails closed on a staged instance (ErrAutoApproveHasStages) and on an
// instance that already left in_progress (ErrInstanceTerminal).
func (inst *Instance) AutoApprove(now time.Time) error {
	if len(inst.Stages) > 0 {
		return ErrAutoApproveHasStages
	}
	if inst.Status != InstanceInProgress {
		return ErrInstanceTerminal
	}
	inst.Status = InstanceApproved
	inst.CompletedAt = &now
	return nil
}

// RejectHere marks the active stage as rejected_here and sets instance Status=InstanceRejected.
func (inst *Instance) RejectHere(reason string) error {
	activeIdx := -1
	for i, s := range inst.Stages {
		if s.Status == StageActive {
			activeIdx = i
			break
		}
	}
	if activeIdx == -1 {
		return ErrNoActiveStage
	}

	now := time.Now().UTC()
	inst.Stages[activeIdx].Status = StageRejectedHere
	inst.Stages[activeIdx].SkipReason = reason
	inst.Stages[activeIdx].CompletedAt = &now
	inst.Status = InstanceRejected
	inst.CompletedAt = &now
	return nil
}

// BumpRevisionVersion enforces monotonic revision_version — mirrors DB trigger.
func (inst *Instance) BumpRevisionVersion(next int) error {
	if next < inst.RevisionVersion {
		return ErrRevisionRegression
	}
	inst.RevisionVersion = next
	return nil
}

// Cancel sets Status=InstanceCancelled and stores reason on CancelReason for
// the caller to persist to approval_instances.cancel_reason. Errors if
// already terminal.
func (inst *Instance) Cancel(reason string) error {
	switch inst.Status {
	case InstanceApproved, InstanceRejected, InstanceCancelled:
		return ErrInstanceTerminal
	}
	now := time.Now().UTC()
	inst.Status = InstanceCancelled
	inst.CompletedAt = &now
	inst.CancelReason = &reason
	return nil
}
