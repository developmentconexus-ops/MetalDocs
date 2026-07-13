package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// ErrFreezeBlockedByUnresolvedComments is returned by executeFreeze when the
// instance still has comments created during its review window (spec.md
// §2.2 "Comment-resolution scope"). Distinct from
// ErrApprovalBlockedByUnresolvedComments (the now-removed final-approve gate,
// W10): this sentinel is freeze's own gate, the sole remaining place this
// concern is enforced.
var ErrFreezeBlockedByUnresolvedComments = errors.New("approval: unresolved instance comments block freeze")

// freezeRepo is the narrow persistence port executeFreeze needs — a subset of
// infrastructure.ApprovalRepository. Declaring it here (rather than depending
// on the full interface) keeps the freeze executor's actual dependency
// surface honest and independently fakeable in unit tests.
type freezeRepo interface {
	HasUnresolvedInstanceComments(ctx context.Context, tx db.Tx, tenantID, documentID string, since time.Time) (bool, error)
	PinFrozenHash(ctx context.Context, tx db.Tx, tenantID, instanceID, hash string) (bool, error)
}

// executeFreeze implements the freeze boundary (design spec.md §2.2, F5).
// Called inside the caller's stage-transition tx, either from
// ReviewVerdictService.RecordVerdict (last review-kind stage completing into
// an approval-kind stage or into instance-approved) or from
// SubmitService.SubmitRevisionForReview (approval-only routes, at submit
// time). Idempotent: re-entry on an already-frozen instance is a no-op.
//
// Concurrency: the caller already holds the instance row FOR UPDATE (via
// LoadInstance) before this runs, so PinFrozenHash's CAS is defense-in-depth,
// not the actual race-resolution mechanism — a concurrent stage-transition
// attempt blocks on the row lock and then fails its own OCC check
// (ExpectedRevisionVersion) on re-read, surfacing as the existing
// infrastructure.ErrStaleRevision (409). No new concurrency primitive is
// introduced by this function.
func executeFreeze(ctx context.Context, tx db.Tx, repo freezeRepo, tenantID string, instance *domain.Instance) error {
	if instance.FrozenContentHash != nil {
		// Already frozen — idempotent no-op, never an error.
		return nil
	}

	blocked, err := repo.HasUnresolvedInstanceComments(ctx, tx, tenantID, instance.DocumentID, instance.SubmittedAt)
	if err != nil {
		return fmt.Errorf("executeFreeze: check unresolved instance comments: %w", err)
	}
	if blocked {
		return ErrFreezeBlockedByUnresolvedComments
	}

	hash := instance.ContentHashAtSubmit
	won, err := repo.PinFrozenHash(ctx, tx, tenantID, instance.ID, hash)
	if err != nil {
		return fmt.Errorf("executeFreeze: pin frozen hash: %w", err)
	}
	if won {
		instance.FrozenContentHash = &hash
	} else {
		// Lost the CAS to a concurrent freeze that beat us to it (defensive —
		// the row lock makes this practically unreachable, see doc comment
		// above). Some freeze happened; treat as success, still reflect the
		// locally-known hash in-memory since it is the same canonical value
		// the winner would have pinned.
		instance.FrozenContentHash = &hash
	}
	return nil
}
