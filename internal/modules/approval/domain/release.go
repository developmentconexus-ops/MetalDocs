package domain

import "time"

// ReleaseHoldReason is the readiness-hold vocabulary from ADR 0085
// §"Observability, attribution, reconciliation". A hold is a VISIBLE,
// queryable state with an explicit reason — never an error, never a state that
// requires human action to leave (except supersede_conflict/plan_invalid,
// which are alert-only per ADR 0068).
//
// Mirrored by the ck_release_generations_hold_reason CHECK in migration 0310;
// the two are kept in sync by ReleaseHoldReasons() + the migration-parity test
// in release_test.go, which parses the CHECK out of the migration file.
type ReleaseHoldReason string

// ReleaseHoldReason values.
const (
	// HoldAwaitingApprovalFact — the generation exists but the approval fact
	// has not been recorded. Structurally impossible on the happy path (the
	// fact producer writes it in the terminal-approval tx), so seeing it means
	// a defect the reconciliation sweep must surface.
	HoldAwaitingApprovalFact ReleaseHoldReason = "awaiting_approval_fact"
	// HoldMaterializing — approved, but the full final artifact set
	// (DOCX + PDF) for this generation does not exist yet.
	HoldMaterializing ReleaseHoldReason = "materializing"
	// HoldAwaitingEffectiveDate — everything is ready but the plan's
	// planned_effective_from is still in the future; a timer is armed.
	HoldAwaitingEffectiveDate ReleaseHoldReason = "awaiting_effective_date"
	// HoldSupersedeConflict — a supersession target moved out from under the
	// plan. The whole release tx rolled back; no partial release.
	HoldSupersedeConflict ReleaseHoldReason = "supersede_conflict"
	// HoldPlanInvalid — the plan cannot be satisfied against ACTUAL release
	// time (e.g. effective_to already elapsed). Never a mid-tx constraint
	// violation.
	HoldPlanInvalid ReleaseHoldReason = "plan_invalid"
	// HoldFailed — the release transaction failed for a non-classified reason.
	HoldFailed ReleaseHoldReason = "failed"
)

// ReleaseHoldReasons returns every legal hold reason, in the ADR's order. The
// single Go-side source for the DB CHECK parity test.
func ReleaseHoldReasons() []ReleaseHoldReason {
	return []ReleaseHoldReason{
		HoldAwaitingApprovalFact,
		HoldMaterializing,
		HoldAwaitingEffectiveDate,
		HoldSupersedeConflict,
		HoldPlanInvalid,
		HoldFailed,
	}
}

// ReleaseGenerationKey is the durable identity every release fact, timer,
// coordinator evaluation and emitted event keys on (ADR 0085 §"Release
// generation: the shared identity").
//
// SubjectKind is retained from the ADR 0082 kernel purely to make the
// document-only scope explicit and machine-checkable: the coordinator refuses
// any non-document subject, and template subjects never produce release facts.
type ReleaseGenerationKey struct {
	TenantID           string
	SubjectKind        SubjectKind
	DocumentID         string
	ApprovalInstanceID string
	RevisionID         string
	RevisionVersion    int
	FrozenContentHash  string
}

// ReleaseAction is the outcome of evaluating the release predicate.
type ReleaseAction string

// ReleaseAction values.
const (
	// ReleaseActionNoop — nothing to do: already released, no longer the
	// freeze head, or the document left the releasable status set. A
	// concurrent evaluation won, or the generation was superseded.
	ReleaseActionNoop ReleaseAction = "noop"
	// ReleaseActionRelease — the full predicate holds: run the release tx.
	ReleaseActionRelease ReleaseAction = "release"
	// ReleaseActionSchedule — everything holds except the effective-date gate:
	// transition approved→scheduled (if not already scheduled) and arm the
	// timer at the planned date.
	ReleaseActionSchedule ReleaseAction = "schedule"
	// ReleaseActionHold — record the hold reason and stop.
	ReleaseActionHold ReleaseAction = "hold"
)

// ReleaseInputs is the fully-loaded, in-transaction snapshot the release
// predicate is evaluated against. Every field is read under the generation
// row's lock, so the predicate itself is a pure function — which is what makes
// it unit-testable without a database and what keeps "why did this not
// release" answerable from one place.
type ReleaseInputs struct {
	// Released is true once a winning release tx stamped released_at.
	Released bool
	// IsFreezeHead is false when a NEWER generation exists for the same
	// document (the document was returned to draft and re-approved).
	IsFreezeHead bool
	// ApprovalFact / ArtifactFact are the two mandatory durable facts.
	ApprovalFact bool
	ArtifactFact bool
	// DocumentStatus is the document's current status, read in-tx.
	DocumentStatus string
	// PlannedEffectiveFrom is the immutable plan date (nil = "effective on
	// release").
	PlannedEffectiveFrom *time.Time
	// EffectiveTo is the plan's expiry, validated against ACTUAL release time.
	EffectiveTo *time.Time
}

// ReleaseDecision is what EvaluateRelease returns.
type ReleaseDecision struct {
	Action ReleaseAction
	// Hold is set only when Action is ReleaseActionHold or
	// ReleaseActionSchedule (schedule carries HoldAwaitingEffectiveDate as its
	// projection reason — the document IS in a readiness hold, it just has a
	// timer attached).
	Hold ReleaseHoldReason
}

// releasableDocumentStatuses is the source-CAS status set from ADR 0085
// §"The release transaction": `WHERE status IN ('approved','scheduled')`.
func releasableDocumentStatus(status string) bool {
	return status == string(DocStatusApprovedForRelease) ||
		status == string(DocStatusScheduledForRelease)
}

// ReleaseSourceStatus mirrors the two documents-module statuses the
// coordinator may transition FROM. They are declared here (as their own type)
// rather than imported from the documents module so the approval kernel's pure
// predicate has no cross-module dependency; the values are pinned to the
// documents domain by the parity test in release_test.go.
type ReleaseSourceStatus string

// ReleaseSourceStatus values.
const (
	DocStatusApprovedForRelease  ReleaseSourceStatus = "approved"
	DocStatusScheduledForRelease ReleaseSourceStatus = "scheduled"
)

// EvaluateRelease is the ADR 0085 release predicate, evaluated as a pure
// function over an in-transaction snapshot.
//
// The conjuncts, in the ADR's order:
//
//  1. Approval fact for G exists AND G is still D's freeze head.
//  2. Artifact fact for G exists (full DOCX+PDF set).
//  3. Effective-date gate: planned_effective_from absent or <= now.
//  4. Supersession-head check — NOT evaluated here: it requires locking the
//     target rows, so it is enforced inside the release transaction and
//     surfaces as HoldSupersedeConflict on rollback.
//
// Ordering note: the head check and the document-status check come FIRST and
// produce a no-op (not a hold), because a superseded generation is not
// "waiting" for anything — recording a hold reason on it would pollute the
// projection with permanently-stuck rows.
func EvaluateRelease(in ReleaseInputs, now time.Time) ReleaseDecision {
	if in.Released {
		return ReleaseDecision{Action: ReleaseActionNoop}
	}
	if !in.IsFreezeHead {
		return ReleaseDecision{Action: ReleaseActionNoop}
	}
	if !releasableDocumentStatus(in.DocumentStatus) {
		return ReleaseDecision{Action: ReleaseActionNoop}
	}
	if !in.ApprovalFact {
		return ReleaseDecision{Action: ReleaseActionHold, Hold: HoldAwaitingApprovalFact}
	}
	if !in.ArtifactFact {
		return ReleaseDecision{Action: ReleaseActionHold, Hold: HoldMaterializing}
	}
	// Delayed-release validation (ADR 0085): a plan whose effective_to is
	// already <= actual time FAILS THE PREDICATE — it must never reach the DB
	// as a ck_documents_effective_window violation mid-tx.
	if in.EffectiveTo != nil && !in.EffectiveTo.After(now) {
		return ReleaseDecision{Action: ReleaseActionHold, Hold: HoldPlanInvalid}
	}
	if in.PlannedEffectiveFrom != nil && in.PlannedEffectiveFrom.After(now) {
		return ReleaseDecision{Action: ReleaseActionSchedule, Hold: HoldAwaitingEffectiveDate}
	}
	return ReleaseDecision{Action: ReleaseActionRelease}
}

// ResolveActualReviewDueAt implements ADR 0085's review-cycle rule against
// ACTUAL release time (ADR 0069 semantics preserved):
//
//   - an explicitly planned review_due_at that is >= actual release time is
//     KEPT verbatim (the human's plan wins);
//   - a planned review_due_at that is BEFORE actual release time is
//     RECOMPUTED from actual release time using the profile's review interval;
//   - no planned date and a positive profile interval derives one from actual
//     release time;
//   - no planned date and no usable interval leaves it unset (nil) — a
//     document with no review cycle, exactly as before ADR 0085.
//
// Returning a value rather than mutating keeps this testable without a DB and
// keeps ck_documents_review_due_sane (review_due_at >= effective_from,
// migration 0274) satisfiable by construction.
func ResolveActualReviewDueAt(planned *time.Time, actualEffectiveFrom time.Time, reviewIntervalDays int) *time.Time {
	if planned != nil && !planned.Before(actualEffectiveFrom) {
		kept := *planned
		return &kept
	}
	if reviewIntervalDays <= 0 {
		if planned != nil {
			// A stale planned date with no interval to recompute from cannot be
			// kept (it would violate ck_documents_review_due_sane); the honest
			// answer is "no review cycle", surfaced by the surfacer's absence
			// of a due date rather than by a constraint violation.
			return nil
		}
		return nil
	}
	derived := actualEffectiveFrom.AddDate(0, 0, reviewIntervalDays)
	return &derived
}
