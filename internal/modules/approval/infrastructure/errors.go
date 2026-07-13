package infrastructure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrStaleRevision is returned when an OCC UPDATE affects zero rows because the
	// caller's expected revision_version no longer matches the stored value.
	ErrStaleRevision = errors.New("approval: stale revision — concurrent modification detected")
	// ErrNoActiveInstance is returned when no in-progress approval instance exists for the document.
	ErrNoActiveInstance = errors.New("approval: no active approval instance for document")
	// ErrDuplicateSubmission is returned when a submission reuses an idempotency key already recorded.
	ErrDuplicateSubmission = errors.New("approval: duplicate submission idempotency key")
	// ErrActorAlreadySigned is returned when the actor already recorded a signoff on this instance.
	ErrActorAlreadySigned = errors.New("approval: actor already signed this instance")
	// ErrCrossTenantSignoff is returned when the acting user's tenant does not match the instance's tenant.
	ErrCrossTenantSignoff = errors.New("approval: cross-tenant signoff rejected")
	// ErrInstanceCompleted is returned when a write is attempted against an instance already in a terminal status.
	ErrInstanceCompleted = errors.New("approval: instance is already terminal")
	// ErrStageNotActive is returned when a stage write is attempted against a stage not in the expected status.
	ErrStageNotActive = errors.New("approval: stage is not in expected status")
	// ErrFKViolation maps Postgres SQLSTATE 23503 (foreign_key_violation).
	ErrFKViolation = errors.New("approval: foreign key violation")
	// ErrCheckViolation maps Postgres SQLSTATE 23514 (check_violation).
	ErrCheckViolation = errors.New("approval: check constraint violation")
	// ErrInsufficientPrivilege maps Postgres SQLSTATE 42501, indicating the tx-local
	// GUC identity context required by RLS was not seeded.
	ErrInsufficientPrivilege = errors.New("approval: insufficient privilege — GUC context missing")
	// ErrUnknownDB is the fallback for any Postgres error MapPgError does not recognize; wraps the original.
	ErrUnknownDB = errors.New("approval: unknown database error")
	// ErrRouteInUse is returned when a route mutation is blocked by the
	// enforce_route_immutable() trigger because instances still reference the route.
	ErrRouteInUse = errors.New("approval: route is referenced by one or more instances and cannot be modified")
	// ErrDuplicateRouteProfile is returned when a route already exists for the tenant+profile_code combination.
	ErrDuplicateRouteProfile = errors.New("approval: a route already exists for this tenant+profile combination")
	// ErrNoActiveContentHash is returned by LoadFrozenContentHash when the
	// approval instance is missing or its frozen_content_hash is NULL (F6,
	// spec §11 no-fallback principle: never substitutes a head-revision hash
	// or any other value). The application layer maps this to
	// ErrContentHashMismatch.
	ErrNoActiveContentHash = errors.New("approval: no active document content hash")
	// ErrNoActiveApprovalRoute is returned by LoadActiveRouteIDByProfile when no
	// active approval route exists for the (tenant, profile). The submit service
	// maps this to documents/domain.ErrApprovalRouteMissing (ADR 0073 in-tx
	// route resolution).
	ErrNoActiveApprovalRoute = errors.New("approval: no active approval route for profile")
	// ErrInstanceNotVisible is returned by ReadService.LoadInstance /
	// LoadActiveInstanceByDocument (F8, spec.md §6.3) when the actor holds
	// SOME view-adjacent capability (passed the tier-2 authz.Require gate)
	// but is not in the instance's visibility set — not the submitting
	// author, not a current-or-past stage pool member (eligible_actor_ids on
	// any stage), and does not hold CapApprovalOversee (tenant-wide) or
	// CapDocumentEdit (area-grade, checked against the instance's own
	// resolved area — never a "tenant" skip-the-area-filter sentinel).
	// Distinct from authz.ErrCapDenied (zero capability at all,
	// mapped to 403): this is "authenticated, has a view-shaped capability,
	// but this specific in-flight instance is outside their eligibility
	// boundary" — mapped to 404 ("cross-boundary = not-found", spec.md §6.3),
	// never 403, so an ineligible viewer cannot distinguish "instance exists
	// but I can't see it" from "instance doesn't exist".
	ErrInstanceNotVisible = errors.New("approval: instance not visible to this actor")
)

// MapHints carries constraint-name hints for SQLSTATE 23505 disambiguation.
type MapHints struct {
	UniqueConstraint string // e.g. "ux_approval_instances_active_document_id"
}

// MapPgError translates *pgconn.PgError to domain errors.
// Falls back to ErrUnknownDB (wrapping original) for unrecognized codes.
func MapPgError(err error, hints MapHints) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505": // unique_violation
		switch pgErr.ConstraintName {
		case "ux_approval_instances_active",
			"ux_approval_instances_active_document_id",
			"approval_instances_document_id_idempotency_key_key":
			return ErrDuplicateSubmission
		case "approval_signoffs_stage_instance_id_actor_user_id_key":
			return ErrActorAlreadySigned
		case "approval_routes_tenant_profile_key", // pre-F2 single-active-route constraint (dropped by migration 0287, kept here for old-row/rollback safety)
			"approval_routes_active_profile_uq": // F2/ADR 0074 partial-unique active-per-profile constraint (migration 0287)
			return ErrDuplicateRouteProfile
		default:
			if hints.UniqueConstraint != "" && pgErr.ConstraintName == hints.UniqueConstraint {
				return ErrDuplicateSubmission
			}
			return fmt.Errorf("%w: unique constraint %q", ErrUnknownDB, pgErr.ConstraintName)
		}
	case "23503": // foreign_key_violation
		return ErrFKViolation
	case "23514": // check_violation
		if pgErr.Message != "" {
			return fmt.Errorf("%w: %s", ErrCheckViolation, pgErr.Message)
		}
		return ErrCheckViolation
	case "42501": // insufficient_privilege
		return ErrInsufficientPrivilege
	case "P0001":
		// enforce_route_immutable() sets message prefix "ErrRouteInUse:" as a stable
		// contract token (migration 0145). String check is intentional; if the trigger
		// message text ever changes, update both here and the migration.
		if strings.HasPrefix(pgErr.Message, "ErrRouteInUse") {
			return ErrRouteInUse
		}
		return fmt.Errorf("%w: %s (SQLSTATE %s)", ErrUnknownDB, pgErr.Message, pgErr.Code)
	default:
		return fmt.Errorf("%w: %s (SQLSTATE %s)", ErrUnknownDB, pgErr.Message, pgErr.Code)
	}
}
