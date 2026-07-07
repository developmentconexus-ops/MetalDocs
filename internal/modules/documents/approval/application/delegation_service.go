package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// ErrDelegationNotFoundOrNotOwned is returned by RevokeDelegation when no
// matching delegation row exists for the given id/tenant, OR the row exists
// but the caller is neither the delegator nor an oversee-holder. Deliberately
// a single, ambiguous error for both cases (mirrors F8's
// ErrInstanceNotVisible precedent) — distinguishing "doesn't exist" from
// "exists but isn't yours" would leak information about another actor's
// delegation to an unauthorized caller.
var ErrDelegationNotFoundOrNotOwned = errors.New("delegation: not found or not owned by caller")

// DelegationService manages approval delegation grants and revocations
// (F9/ADR 0077). Delegation widens the INPUT to domain.CheckEligibility /
// domain.CheckSoD (via domain.ResolveEligibleIdentity, consumed by
// DecisionService and ReviewVerdictService) — it never grants a capability
// and never bypasses authz.Require.
type DelegationService struct {
	repo    infrastructure.ApprovalRepository
	emitter EventEmitter
	clock   Clock
}

// newDelegationService constructs a DelegationService.
func newDelegationService(repo infrastructure.ApprovalRepository, emitter EventEmitter, clock Clock) *DelegationService {
	return &DelegationService{repo: repo, emitter: emitter, clock: clock}
}

// CreateDelegation grants delegator's eligibility to delegate for the given
// window. App-checks-first (domain.NewDelegation validates self-delegation
// and window before any DB write); the DB CHECK constraints on
// approval_delegations are the tripwire-last-line backstop for the same two
// rules. delegatorID is always the calling actor's own identity — never
// client-supplied — enforced by the HTTP handler deriving it from the
// authenticated session before this is called.
func (s *DelegationService) CreateDelegation(ctx context.Context, runner db.TxRunner, tenantID, delegatorID, delegateID, reason string, startsAt, endsAt time.Time) (domain.Delegation, error) {
	now := s.clock.Now()
	d, err := domain.NewDelegation(domain.DelegationParams{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		DelegatorID: delegatorID,
		DelegateID:  delegateID,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Reason:      reason,
		CreatedBy:   delegatorID,
		CreatedAt:   now,
	})
	if err != nil {
		return domain.Delegation{}, err
	}

	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		if err := s.repo.InsertDelegation(ctx, tx, *d); err != nil {
			return fmt.Errorf("create delegation: insert: %w", err)
		}

		payload, err := json.Marshal(map[string]any{
			"delegator_id": delegatorID,
			"delegate_id":  delegateID,
			"starts_at":    startsAt,
			"ends_at":      endsAt,
			"reason":       reason,
		})
		if err != nil {
			return fmt.Errorf("create delegation: marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     tenantID,
			EventType:    EventTypeDelegationGranted,
			ActorUserID:  delegatorID,
			ResourceType: "approval_delegation",
			ResourceID:   d.ID(),
			Reason:       reason,
			PayloadJSON:  payload,
			OccurredAt:   now,
		}); err != nil {
			return fmt.Errorf("create delegation: emit event: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.Delegation{}, err
	}
	return *d, nil
}

// RevokeDelegation hard-deletes a delegation (F9/ADR 0077 §5: revocation is
// real-time — the next LoadActiveDelegationsFor call, in-tx at use-time,
// cannot see a deleted row; there is no soft revoked_at flag to go stale).
// Only the delegator or a CapApprovalOversee holder may revoke.
func (s *DelegationService) RevokeDelegation(ctx context.Context, runner db.TxRunner, tenantID, delegationID, callerID string) error {
	return runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		// Oversee is a tenant-grade capability; a caller who holds it may
		// revoke any delegation in the tenant (mirrors the module's existing
		// oversee-widens-scope pattern from F8's worklist). A caller who does
		// not hold it may only revoke their own grant, enforced atomically by
		// DeleteDelegation's WHERE clause.
		callerIsOversee := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant") == nil

		deleted, delegatorID, err := deleteDelegationAndGetDelegator(ctx, tx, s.repo, tenantID, delegationID, callerID, callerIsOversee)
		if err != nil {
			return fmt.Errorf("revoke delegation: %w", err)
		}
		if !deleted {
			return ErrDelegationNotFoundOrNotOwned
		}

		payload, err := json.Marshal(map[string]any{
			"delegation_id": delegationID,
			"delegator_id":  delegatorID,
		})
		if err != nil {
			return fmt.Errorf("revoke delegation: marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     tenantID,
			EventType:    EventTypeDelegationRevoked,
			ActorUserID:  callerID,
			ResourceType: "approval_delegation",
			ResourceID:   delegationID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("revoke delegation: emit event: %w", err)
		}
		return nil
	})
}

// deleteDelegationAndGetDelegator wraps repo.DeleteDelegation. The repository
// method itself does not return the delegator id (it only reports whether a
// row was deleted), so for the governance-event payload we accept callerID as
// a reasonable best-effort delegator hint when the caller IS the delegator
// (the common case); when revocation happens via oversee, delegatorID may
// differ from callerID and is best-effort empty in that path — the
// delegation_id in the payload remains the authoritative audit key.
func deleteDelegationAndGetDelegator(ctx context.Context, tx *sql.Tx, repo infrastructure.ApprovalRepository, tenantID, delegationID, callerID string, callerIsOversee bool) (bool, string, error) {
	deleted, err := repo.DeleteDelegation(ctx, tx, tenantID, delegationID, callerID, callerIsOversee)
	if err != nil {
		return false, "", err
	}
	return deleted, callerID, nil
}
