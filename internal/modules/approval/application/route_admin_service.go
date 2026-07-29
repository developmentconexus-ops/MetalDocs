package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// routeProfileFKConstraint is the FK on approval_routes (tenant_id, profile_code)
// → metaldocs.document_profiles. A violation of this specific constraint means
// the profile is not registered for the tenant: report it as a validation error.
// Any other FK violation must fall through to a 500 with diagnostic context.
const routeProfileFKConstraint = "approval_routes_document_profile_fk"

// RouteAdminService manages approval route configuration changes.
type RouteAdminService struct {
	repo         infrastructure.ApprovalRepository
	emitter      EventEmitter
	clock        Clock
	idempStore   RouteAdminIdempStore
	policyReader ProfilePolicyReader // G1: resolves the profile route-signature policy off-tx; nil ⇒ friendly check skipped (DB trigger is the last line)
}

// WithPolicyReader returns a copy of the service wired with the G1
// profile-policy reader. Passing nil leaves the friendly route-shape check
// disabled (the DB deferrable trigger remains the authoritative last line).
func (s *RouteAdminService) WithPolicyReader(reader ProfilePolicyReader) *RouteAdminService {
	if s == nil {
		return nil
	}
	cp := *s
	cp.policyReader = reader
	return &cp
}

// WithIdempStore returns a copy of the service wired with an idempotency store.
// Passing nil leaves replay disabled (useful for unit tests that exercise the
// service without the Postgres store).
func (s *RouteAdminService) WithIdempStore(store RouteAdminIdempStore) *RouteAdminService {
	if s == nil {
		return nil
	}
	cp := *s
	cp.idempStore = store
	return &cp
}

// ErrRouteNotFound is returned when a route does not exist for the tenant.
var ErrRouteNotFound = errors.New("route_admin: route not found")

// ErrRouteAlreadyInactive is returned when the route is already inactive.
var ErrRouteAlreadyInactive = errors.New("route_admin: route already inactive")

// ErrRouteDeactivateReasonRequired is returned when DeactivateRouteInput.Reason
// is empty or whitespace-only.
var ErrRouteDeactivateReasonRequired = errors.New("route_admin: deactivate reason must not be empty")

// ErrRouteProfileUnknown is returned when the route's profile_code has no
// matching document profile for the tenant (FK violation on create). Surfaced
// as an actionable 4xx instead of an opaque 500.
var ErrRouteProfileUnknown = errors.New("route_admin: profile_code not registered for tenant")

// ErrDocumentSubjectKeyMismatch is returned when a document-subject route's
// effective SubjectKey diverges from its ProfileCode. Per ratified rail R1,
// profile_code is the backward-compat ALIAS for (document, profile_code) —
// LoadActiveRouteIDByProfile delegates to LoadActiveRouteIDBySubject(tenantID,
// "document", profileCode), so a document route whose subject_key differs
// from profile_code would be unfindable by the alias lookup. That divergence
// must be impossible, not merely discouraged.
var ErrDocumentSubjectKeyMismatch = errors.New("approval: document route subject_key must equal profile_code")

// ErrTemplateSubjectKeyMismatch is the template-subject twin of
// ErrDocumentSubjectKeyMismatch (ADR 0086). A template ROUTE is keyed by the
// profile it governs, so subject_key must equal profile_code — the DB enforces
// it via approval_routes_template_subject_key_check (migration 0315) and the
// submit/preview paths resolve the route by the template's doc_type_code, so a
// divergent key would make the route unfindable.
var ErrTemplateSubjectKeyMismatch = errors.New("approval: template route subject_key must equal profile_code")

// CreateRouteInput carries all inputs for Create.
type CreateRouteInput struct {
	TenantID    string
	ProfileCode string
	Name        string
	ActorUserID string
	// SubjectKind and SubjectKey generalize what the route governs (M3 kernel
	// extraction, ADR 0082 / P2.S3). An empty SubjectKind means document.
	// Post-ADR-0086 BOTH kinds are profile-keyed: an empty SubjectKey defaults
	// to ProfileCode, and a supplied SubjectKey that diverges from ProfileCode
	// is rejected (ErrDocumentSubjectKeyMismatch / ErrTemplateSubjectKeyMismatch).
	SubjectKind    string
	SubjectKey     string
	IdempotencyKey string
	Stages         []domain.Stage
}

// CreateRouteResult is returned on successful route creation. Beyond RouteID
// it carries the route's PERSISTED projection (F-E4-2): the create response
// used to serialize name:"", version:0, active:false, stages:null while the DB
// row was fully populated, because the handler had nothing but an id to build
// RouteResponse from. Every field below is read back from the committed row —
// never echoed from the request — so create and the read side cannot drift.
type CreateRouteResult struct {
	RouteID string
	// ProfileCode is non-nil for every route post-ADR-0086 (both subject kinds
	// are profile-keyed). The pointer is kept because the column is still
	// NULL-able at the DB level: a nil here is a data defect surfaced honestly
	// as JSON null, never collapsed to the "" sentinel.
	ProfileCode *string
	Name        string
	Version     int
	Active      bool
	// InUse reports whether any approval instance already references the route
	// (routes are immutable once used — F2/W1).
	InUse     bool
	CreatedAt time.Time
	Stages    []domain.Stage
}

// UpdateRouteInput carries all inputs for Update.
type UpdateRouteInput struct {
	TenantID        string
	RouteID         string
	Name            string
	ActorUserID     string
	IdempotencyKey  string
	ExpectedVersion int
	Stages          []domain.Stage
}

// UpdateRouteResult is returned on successful route update.
type UpdateRouteResult struct {
	RouteID    string
	NewVersion int
}

// DeactivateRouteInput carries all inputs for Deactivate.
type DeactivateRouteInput struct {
	TenantID        string
	RouteID         string
	ActorUserID     string
	IdempotencyKey  string
	Reason          string
	ExpectedVersion int
}

// DeactivateRouteResult is returned on successful route deactivation.
type DeactivateRouteResult struct {
	RouteID string
}

// ListRoutesResult is returned by List.
type ListRoutesResult struct {
	Routes []infrastructure.Route
}

// Create creates a new approval route and all route stages.
func (s *RouteAdminService) Create(ctx context.Context, runner db.TxRunner, in CreateRouteInput) (CreateRouteResult, error) {
	const op = "create"

	// Resolve the effective subject up front (a pure function of client input,
	// H-PRE-1-safe — no tx, no I/O). It feeds BOTH the idempotency hash below
	// (QR-A finding A: profile_code alone is "" for every template route, so
	// two creates sharing an Idempotency-Key + name + stages but DIFFERENT
	// subject_key used to hash identically and silently replay) and createTx.
	// A resolution failure (e.g. ErrDocumentSubjectKeyMismatch) is a
	// deterministic function of the request body — retried identically it
	// fails identically — so it is reported directly without ever occupying
	// an idempotency slot.
	subject, err := resolveCreateRouteSubject(in)
	if err != nil {
		return CreateRouteResult{}, wrapRouteAdminErr(op, err)
	}

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeCreateRoutePayloadHash(in.ProfileCode, in.Name, in.Stages, subject)
		var err error
		committer, replay, err = s.idempStore.BeginCreateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return CreateRouteResult{}, err
		}
		if replay != nil {
			// F-E4-2: a replay must return the SAME body the original create
			// returned, so it re-reads the persisted projection rather than
			// echoing a bare route id (the replay envelope stores only the id).
			return s.loadRouteProjection(ctx, runner, in.TenantID, replay.RouteID)
		}
	}

	result, err := s.createTx(ctx, runner, in, subject)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return CreateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		if cerr := committer.Complete(result.RouteID, nil); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

// resolvePolicy returns the profile's route-signature policy via the G1 reader,
// resolved OFF any write tx (H-PRE-1: the reader records CapTaxonomyView in
// taxonomy's own short tx, so it must never run inside a lock-holding approval
// tx). A nil reader yields "" (friendly check skipped; the DB deferrable trigger
// stays authoritative). A read error propagates so a genuine lookup failure
// surfaces rather than being silently downgraded (no-fallback principle).
func (s *RouteAdminService) resolvePolicy(ctx context.Context, tenantID, profileCode string) (taxonomydomain.RoutePolicy, error) {
	if s.policyReader == nil {
		return "", nil
	}
	return s.policyReader.RoutePolicy(ctx, tenantID, profileCode)
}

// resolveCreateRouteSubject builds the route's Subject from CreateRouteInput
// (M3 kernel extraction, ADR 0082 / P2.S3). When both SubjectKind and
// SubjectKey are absent, this reproduces the pre-P2.S3 default exactly:
// (document, ProfileCode). An explicit SubjectKind is used as given; an
// absent SubjectKey defaults to ProfileCode for BOTH kinds (ADR 0086), so an
// empty ProfileCode yields an empty key and fails Subject.Validate rather
// than persisting an unkeyed route.
//
// Rail R1 invariant: profile_code is the backward-compat ALIAS for
// (document, profile_code). So once defaulting is done, a document-kind
// subject whose key diverges from ProfileCode is rejected with
// ErrDocumentSubjectKeyMismatch — an explicit subject_kind=document plus a
// divergent subject_key would otherwise be silently accepted and become
// unfindable via the profile-code alias lookup.
//
// ADR 0086 puts template routes under the identical rule: a template ROUTE is
// keyed by the profile (doc_type_code) it governs, so its key defaults to and
// must equal ProfileCode too (DB truth:
// approval_routes_template_subject_key_check, migration 0315). The mismatch
// sentinel is subject-specific only in its message; both kinds reject the same
// divergence for the same reason.
func resolveCreateRouteSubject(in CreateRouteInput) (domain.Subject, error) {
	kind := domain.SubjectKind(in.SubjectKind)
	if kind == "" {
		kind = domain.SubjectKindDocument
	}
	key := in.SubjectKey
	if key == "" {
		key = in.ProfileCode
	}
	subject := domain.Subject{Kind: kind, Key: key}
	// Kind validity is decided FIRST: an unknown subject_kind is a malformed
	// request (ErrInvalidSubjectKind), not a key mismatch, and must not be
	// reported as one just because its key also happens to diverge.
	if err := subject.Validate(); err != nil {
		return domain.Subject{}, err
	}
	if key != in.ProfileCode {
		if kind == domain.SubjectKindTemplate {
			return domain.Subject{}, ErrTemplateSubjectKeyMismatch
		}
		return domain.Subject{}, ErrDocumentSubjectKeyMismatch
	}
	return subject, nil
}

// createTx takes the already-resolved subject (Create resolves it once, up
// front, so both the idempotency hash and this call see the identical
// value — see resolveCreateRouteSubject's callsite in Create).
func (s *RouteAdminService) createTx(ctx context.Context, runner db.TxRunner, in CreateRouteInput, subject domain.Subject) (CreateRouteResult, error) {
	var result CreateRouteResult

	if err := subject.Validate(); err != nil {
		return result, err
	}

	// G1: resolve the per-profile route-signature policy OFF-TX (H-PRE-1) so the
	// friendly route-shape check runs before the write tx opens. Every route —
	// document and template alike — is keyed by a profile (ADR 0086), so the
	// per-profile signature policy applies to both kinds; there is no skip.
	policy, err := s.resolvePolicy(ctx, in.TenantID, in.ProfileCode)
	if err != nil {
		return result, err
	}
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		route := domain.Route{
			TenantID:    in.TenantID,
			ProfileCode: in.ProfileCode,
			Subject:     subject,
			Version:     1,
			Stages:      in.Stages,
		}
		if err := route.Validate(policy); err != nil {
			return err
		}

		var routeID string
		err = tx.QueryRowContext(ctx, `
			INSERT INTO approval_routes
				(tenant_id, profile_code, name, version, created_by, active, subject_kind, subject_key)
			VALUES ($1, $2, $3, 1, $4, TRUE, $5, $6)
			RETURNING id`,
			in.TenantID, in.ProfileCode, in.Name, in.ActorUserID,
			string(route.Subject.Kind), route.Subject.Key,
		).Scan(&routeID)
		if err != nil {
			mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
			// Only a 23503 violation of the profile FK means the profile is not
			// registered for the tenant — map that to a validation error. Any other
			// FK violation falls through to a wrapped error so WriteError logs it at 500.
			if errors.Is(mapped, infrastructure.ErrFKViolation) {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.ConstraintName == routeProfileFKConstraint {
					return ErrRouteProfileUnknown
				}
			}
			return fmt.Errorf("insert route: %w", mapped)
		}

		if err := insertRouteStages(ctx, tx, in.TenantID, routeID, in.Stages); err != nil {
			return err
		}

		// profile_code is always present now (ADR 0086): every route, both
		// kinds, is keyed by a profile. subject_kind/subject_key carry the
		// canonical resolved values, not the raw possibly-defaulted input.
		payload, err := json.Marshal(map[string]any{
			"route_id":      routeID,
			"profile_code":  in.ProfileCode,
			"subject_kind":  string(subject.Kind),
			"subject_key":   subject.Key,
			"stage_count":   len(in.Stages),
			"initial_state": "active",
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigCreated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   routeID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		// F-E4-2: build the response from the PERSISTED row + stages, read back
		// inside this same tx right after the writes above. Reading (rather
		// than echoing the request) is what makes create's body identical in
		// shape and content to the read side, and picks up DB-applied values
		// the request never carried (created_at, the stage_kind NOT NULL
		// DEFAULT, version).
		result, err = scanRouteProjectionTx(ctx, tx, in.TenantID, routeID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return CreateRouteResult{}, err
	}
	return result, nil
}

// loadRouteProjection reads a route's persisted projection in its own short
// transaction (F-E4-2 replay path). route.manage is asserted here exactly as
// every other RouteAdminService entry point does — a replayed create is still
// a route-admin read of tenant configuration, not an unguarded lookup.
func (s *RouteAdminService) loadRouteProjection(ctx context.Context, runner db.TxRunner, tenantID, routeID string) (CreateRouteResult, error) {
	var result CreateRouteResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}
		var err error
		result, err = scanRouteProjectionTx(ctx, tx, tenantID, routeID)
		return err
	})
	if err != nil {
		return CreateRouteResult{}, wrapRouteAdminErr("create", err)
	}
	return result, nil
}

// scanRouteProjectionTx reads the committed approval_routes row plus its
// persisted stages (with selectors) inside the caller's tx. profile_code is
// scanned as a nullable column because the column is still NULL-able at the DB
// level; post-ADR-0086 every route carries one, so a NULL here is a data defect
// and stays NULL on the wire instead of collapsing to the "" sentinel. in_use
// is computed from approval_instances, tenant-scoped on both sides of the
// correlation.
func scanRouteProjectionTx(ctx context.Context, tx *sql.Tx, tenantID, routeID string) (CreateRouteResult, error) {
	var (
		out         CreateRouteResult
		profileCode sql.NullString
	)
	err := tx.QueryRowContext(ctx, `
		SELECT r.id, r.name, r.version, r.active, r.created_at, r.profile_code,
		       EXISTS (
		         SELECT 1 FROM approval_instances ai
		          WHERE ai.route_id = r.id
		            AND ai.tenant_id = r.tenant_id
		       ) AS in_use
		  FROM approval_routes r
		 WHERE r.id = $1
		   AND r.tenant_id = $2::uuid`,
		routeID, tenantID,
	).Scan(&out.RouteID, &out.Name, &out.Version, &out.Active, &out.CreatedAt, &profileCode, &out.InUse)
	if errors.Is(err, sql.ErrNoRows) {
		return CreateRouteResult{}, ErrRouteNotFound
	}
	if err != nil {
		return CreateRouteResult{}, fmt.Errorf("load route projection: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}
	if profileCode.Valid {
		code := profileCode.String
		out.ProfileCode = &code
	}

	stages, err := loadRouteStagesTx(ctx, tx, routeID)
	if err != nil {
		return CreateRouteResult{}, err
	}
	out.Stages = stages
	return out, nil
}

// Update updates route metadata and replaces all route stages atomically.
func (s *RouteAdminService) Update(ctx context.Context, runner db.TxRunner, in UpdateRouteInput) (UpdateRouteResult, error) {
	const op = "update"

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeUpdateRoutePayloadHash(in.RouteID, in.ExpectedVersion, in.Name, in.Stages)
		var err error
		committer, replay, err = s.idempStore.BeginUpdateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return UpdateRouteResult{}, err
		}
		if replay != nil {
			res := UpdateRouteResult{RouteID: replay.RouteID}
			if replay.NewVersion != nil {
				res.NewVersion = *replay.NewVersion
			}
			return res, nil
		}
	}

	result, err := s.updateTx(ctx, runner, in)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return UpdateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		nv := result.NewVersion
		if cerr := committer.Complete(result.RouteID, &nv); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

// resolveUpdatePolicy resolves the route's profile route-signature policy OFF-TX
// for an update (H-PRE-1). UpdateRouteInput carries no profile_code, so it first
// reads the route's (immutable) profile_code in a short non-recording tx, then
// resolves the policy via the G1 reader. The profile_code read is scoped to
// ACTIVE routes to mirror the DB direction-A trigger, which only enforces on
// active routes: an inactive/absent route yields "" (friendly check skipped),
// keeping the friendly and authoritative layers identical in scope while the
// write tx below surfaces any real not-found/stale error. Reading profile_code
// off-tx is TOCTOU-safe because a route's profile_code is immutable across its
// life (supersede preserves it). An active route whose profile_code is empty is
// a post-ADR-0086 data defect and fails closed here — see below.
func (s *RouteAdminService) resolveUpdatePolicy(ctx context.Context, runner db.TxRunner, tenantID, routeID string) (taxonomydomain.RoutePolicy, error) {
	if s.policyReader == nil {
		return "", nil
	}
	var (
		profileCode string
		found       bool
	)
	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		code, ok, err := loadActiveRouteProfileCode(ctx, tx, tenantID, routeID)
		profileCode, found = code, ok
		return err
	}); err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}
	// Post-ADR-0086 every route — both subject kinds — carries a profile_code,
	// so an active route with an empty one is a data defect (0315's CHECKs make
	// it unreachable). Fail closed rather than silently skipping the
	// route-signature policy, which would let a mis-shaped route through.
	if profileCode == "" {
		return "", fmt.Errorf("route_admin: active route %s has no profile_code (ADR 0086 requires one for every subject kind)", routeID)
	}
	return s.policyReader.RoutePolicy(ctx, tenantID, profileCode)
}

func (s *RouteAdminService) updateTx(ctx context.Context, runner db.TxRunner, in UpdateRouteInput) (UpdateRouteResult, error) {
	var result UpdateRouteResult
	// G1: resolve the route's profile policy OFF-TX (H-PRE-1) before the write tx.
	policy, err := s.resolveUpdatePolicy(ctx, runner, in.TenantID, in.RouteID)
	if err != nil {
		return result, err
	}
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		locked, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID)
		if err != nil {
			return err
		}

		currentStages, err := loadRouteStagesTx(ctx, tx, in.RouteID)
		if err != nil {
			return err
		}
		stagesChanged := !stagesEqual(currentStages, in.Stages)

		route := domain.Route{
			ID:       in.RouteID,
			TenantID: in.TenantID,
			Stages:   in.Stages,
		}
		if err := route.Validate(policy); err != nil {
			return err
		}

		if in.ExpectedVersion != 0 && locked.Version != in.ExpectedVersion {
			return infrastructure.ErrStaleRevision
		}

		newRouteID, newVersion, err := s.updateInPlaceOrSupersede(ctx, tx, in, locked, stagesChanged)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"route_id":    newRouteID,
			"new_version": newVersion,
			"stage_count": len(in.Stages),
			"superseded":  newRouteID != in.RouteID,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigUpdated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   newRouteID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		result = UpdateRouteResult{RouteID: newRouteID, NewVersion: newVersion}
		return nil
	})
	if err != nil {
		return UpdateRouteResult{}, err
	}
	return result, nil
}

// updateInPlaceOrSupersede tries the cheap in-place mutation first (matches
// pre-F2 behavior for a route never referenced by any instance). If the
// enforce_route_immutable trigger (migration 0287) rejects it with
// ErrRouteInUse, it falls back to the versioned-supersede path: the current
// row is retired (active=false, superseded_at=now() — the ONLY columns the
// trigger still allows to change on an in-use row) and a brand-new row is
// inserted carrying the incremented version and the new definition. Both
// branches run inside the caller's single transaction.
func (s *RouteAdminService) updateInPlaceOrSupersede(ctx context.Context, tx *sql.Tx, in UpdateRouteInput, locked lockedRouteState, stagesChanged bool) (routeID string, newVersion int, err error) {
	// The in-place UPDATE below is speculative: enforce_route_immutable() may
	// reject it with ErrRouteInUse (P0001), which — per Postgres transaction
	// semantics — aborts the entire enclosing transaction, not just the failed
	// statement. Without a SAVEPOINT here, every subsequent statement in this
	// tx (including the supersede fallback three lines below) fails closed
	// with SQLSTATE 25P02 ("current transaction is aborted"), masking the real
	// ErrRouteInUse behind an opaque 500. F10 live-QA caught this: a fresh
	// in-use route's first-ever versioned update 500'd instead of superseding.
	const savepoint = "route_update_attempt"
	if _, spErr := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); spErr != nil {
		return "", 0, fmt.Errorf("savepoint before speculative update: %w", spErr)
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE approval_routes
		   SET name = $1,
		       version = version + 1
		 WHERE id = $2
		   AND tenant_id = $3
		RETURNING version`,
		in.Name, in.RouteID, in.TenantID,
	).Scan(&newVersion)
	if err == nil {
		if _, relErr := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint); relErr != nil {
			return "", 0, fmt.Errorf("release savepoint after in-place update: %w", relErr)
		}
		if stagesChanged {
			if _, err := tx.ExecContext(ctx, `DELETE FROM approval_route_stages WHERE route_id = $1`, in.RouteID); err != nil {
				return "", 0, fmt.Errorf("delete stages: %w", err)
			}
			if err := insertRouteStages(ctx, tx, in.TenantID, in.RouteID, in.Stages); err != nil {
				return "", 0, err
			}
		}
		return in.RouteID, newVersion, nil
	}

	mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
	if !errors.Is(mapped, infrastructure.ErrRouteInUse) {
		return "", 0, fmt.Errorf("update route: %w", mapped)
	}

	// Roll back to the savepoint to un-abort the transaction before issuing
	// any further statements — the failed in-place UPDATE must not poison the
	// supersede fallback that follows.
	if _, rbErr := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint); rbErr != nil {
		return "", 0, fmt.Errorf("rollback to savepoint after ErrRouteInUse: %w", rbErr)
	}

	// In use: retire the old row (trigger permits an active/superseded_at-only
	// UPDATE on an in-use row), then insert the new version as a new row. Order
	// matters — the partial unique index on (tenant_id, profile_code) WHERE
	// active would reject the new row's INSERT if the old row were still
	// active at that point.
	if _, err := tx.ExecContext(ctx, `
		UPDATE approval_routes
		   SET active = FALSE,
		       superseded_at = $1
		 WHERE id = $2
		   AND tenant_id = $3`,
		s.clock.Now(), in.RouteID, in.TenantID,
	); err != nil {
		return "", 0, fmt.Errorf("supersede route: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}

	newVersion = locked.Version + 1
	err = tx.QueryRowContext(ctx, `
		INSERT INTO approval_routes
			(tenant_id, profile_code, name, version, created_by, active, subject_kind, subject_key)
		SELECT tenant_id, profile_code, $1, $2, $3, TRUE, subject_kind, subject_key
		  FROM approval_routes
		 WHERE id = $4
		RETURNING id`,
		in.Name, newVersion, in.ActorUserID, in.RouteID,
	).Scan(&routeID)
	if err != nil {
		return "", 0, fmt.Errorf("insert superseding route version: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}

	if err := insertRouteStages(ctx, tx, in.TenantID, routeID, in.Stages); err != nil {
		return "", 0, err
	}

	return routeID, newVersion, nil
}

// Deactivate marks a route inactive.
func (s *RouteAdminService) Deactivate(ctx context.Context, runner db.TxRunner, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	const op = "deactivate"

	reason := strings.TrimSpace(in.Reason)
	in.Reason = reason

	var (
		committer RouteAdminReplayCommitter
		replay    *RouteAdminReplay
	)
	if s.idempStore != nil && in.IdempotencyKey != "" {
		hash := computeDeactivateRoutePayloadHash(in.RouteID, in.ExpectedVersion, in.Reason)
		var err error
		committer, replay, err = s.idempStore.BeginDeactivateReplay(ctx, in.TenantID, in.ActorUserID, in.IdempotencyKey, hash)
		if err != nil {
			return DeactivateRouteResult{}, err
		}
		if replay != nil {
			return DeactivateRouteResult{RouteID: replay.RouteID}, nil
		}
	}

	if reason == "" {
		if committer != nil {
			if ferr := committer.Fail(ErrRouteDeactivateReasonRequired); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", ErrRouteDeactivateReasonRequired,
					"fail_err", ferr,
				)
			}
		}
		return DeactivateRouteResult{}, ErrRouteDeactivateReasonRequired
	}

	result, err := s.deactivateTx(ctx, runner, in)
	if err != nil {
		if committer != nil {
			if ferr := committer.Fail(err); ferr != nil {
				slog.ErrorContext(ctx, "route_admin: committer.Fail failed; idempotency slot may be orphaned",
					"op", op,
					"key", in.IdempotencyKey,
					"primary_err", err,
					"fail_err", ferr,
				)
			}
		}
		return DeactivateRouteResult{}, wrapRouteAdminErr(op, err)
	}
	if committer != nil {
		if cerr := committer.Complete(result.RouteID, nil); cerr != nil {
			return result, wrapRouteAdminErr(op, fmt.Errorf("complete replay: %w", cerr))
		}
	}
	return result, nil
}

func (s *RouteAdminService) deactivateTx(ctx context.Context, runner db.TxRunner, in DeactivateRouteInput) (DeactivateRouteResult, error) {
	var result DeactivateRouteResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		lockedRoute, err := lockRouteForUpdate(ctx, tx, in.TenantID, in.RouteID)
		if err != nil {
			return err
		}
		if !lockedRoute.Active {
			if in.ExpectedVersion > 0 && lockedRoute.Version != in.ExpectedVersion {
				return infrastructure.ErrStaleRevision
			}
			return ErrRouteAlreadyInactive
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE approval_routes
			   SET active = FALSE,
			       version = version + 1
			 WHERE id = $1
			   AND tenant_id = $2
			   AND active = TRUE
			   AND ($3 = 0 OR version = $3)`,
			in.RouteID, in.TenantID, in.ExpectedVersion,
		)
		if err != nil {
			mapped := infrastructure.MapPgError(err, infrastructure.MapHints{})
			if errors.Is(mapped, infrastructure.ErrRouteInUse) {
				return mapped
			}
			return fmt.Errorf("deactivate route: %w", mapped)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("deactivate route rows affected: %w", err)
		}
		if rows == 0 {
			return infrastructure.ErrStaleRevision
		}

		payload, err := json.Marshal(map[string]any{
			"route_id": in.RouteID,
			"active":   false,
			"reason":   in.Reason,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		if err := s.emitter.Emit(ctx, tx, GovernanceEvent{
			TenantID:     in.TenantID,
			EventType:    EventTypeRouteConfigDeactivated,
			ActorUserID:  in.ActorUserID,
			ResourceType: "approval_route",
			ResourceID:   in.RouteID,
			PayloadJSON:  payload,
			OccurredAt:   s.clock.Now(),
		}); err != nil {
			return fmt.Errorf("emit event: %w", err)
		}

		result = DeactivateRouteResult{RouteID: in.RouteID}
		return nil
	})
	if err != nil {
		return DeactivateRouteResult{}, err
	}
	return result, nil
}

// List loads tenant-scoped routes inside a single transaction that owns the
// tenant/actor GUCs and the route.admin authz check. This closes the TOCTOU
// gap where the prior handler ran authz in one tx and the list query in another.
func (s *RouteAdminService) List(ctx context.Context, runner db.TxRunner, tenantID, actorID string) (ListRoutesResult, error) {
	const op = "list"

	var result ListRoutesResult
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)

		if err := authz.Require(ctx, tx, string(iamdomain.CapRouteManage), "tenant"); err != nil {
			return err
		}

		routes, err := s.repo.ListRoutesTx(ctx, tx, tenantID)
		if err != nil {
			return err
		}

		result = ListRoutesResult{Routes: routes}
		return nil
	})
	if err != nil {
		return ListRoutesResult{}, wrapRouteAdminErr(op, err)
	}
	return result, nil
}

// wrapRouteAdminErr standardises the error envelope. Sentinels and mapped
// repository errors pass through unchanged so handler-level errors.Is checks
// keep working; everything else is annotated with the op name.
func wrapRouteAdminErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if isPassThroughRouteAdminErr(err) {
		return err
	}
	return fmt.Errorf("route_admin: %s: %w", op, err)
}

func isPassThroughRouteAdminErr(err error) bool {
	switch {
	case errors.Is(err, ErrRouteNotFound),
		errors.Is(err, ErrRouteAlreadyInactive),
		errors.Is(err, ErrRouteDeactivateReasonRequired),
		errors.Is(err, ErrRouteProfileUnknown),
		errors.Is(err, infrastructure.ErrStaleRevision),
		errors.Is(err, infrastructure.ErrRouteInUse),
		errors.Is(err, infrastructure.ErrDuplicateRouteProfile):
		return true
	}
	var capDenied authz.ErrCapDenied
	return errors.As(err, &capDenied)
}

type lockedRouteState struct {
	ID      string
	Version int
	Active  bool
}

// loadActiveRouteProfileCode reads the immutable profile_code of an ACTIVE route
// with a plain non-recording SELECT (safe in any tx; no authz recording, so it
// never trips H-PRE-1). Returns found=false when the route is inactive or absent
// so the G1 friendly check can be skipped in lockstep with the DB trigger's
// active-only scope. The column is still NULL-able at the DB level, so a NULL
// is reported as found=true with code="" — the caller treats that as a data
// defect (ADR 0086 / migration 0315 require a profile_code on every route),
// distinct from "route not eligible" (absent/inactive).
func loadActiveRouteProfileCode(ctx context.Context, tx *sql.Tx, tenantID, routeID string) (string, bool, error) {
	var code sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT profile_code
		  FROM approval_routes
		 WHERE id = $1
		   AND tenant_id = $2
		   AND active = TRUE`,
		routeID, tenantID,
	).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("route_admin: load route profile_code: %w", err)
	}
	return code.String, true, nil
}

func lockRouteForUpdate(ctx context.Context, tx *sql.Tx, tenantID, routeID string) (lockedRouteState, error) {
	var state lockedRouteState
	err := tx.QueryRowContext(ctx, `
		SELECT id, version, active
		  FROM approval_routes
		 WHERE id = $1
		   AND tenant_id = $2
		 FOR UPDATE`,
		routeID, tenantID,
	).Scan(&state.ID, &state.Version, &state.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lockedRouteState{}, ErrRouteNotFound
		}
		return lockedRouteState{}, fmt.Errorf("route_admin: lock route: %w", err)
	}
	return state, nil
}

// insertRouteStages persists the approval_route_stages rows, then
// insertRouteStageSelectors persists the ActorSelector rows that govern each
// stage's eligible-actor pool (unit 3.2 M4, B2 DB expand / 6b contract step).
// Every stage's Selectors is written — Selectors is the sole source of truth
// post-slice-6b; the flat RequiredRole/AreaCode → selector synthesis happens
// at the HTTP boundary (route_admin_handler.go) before a Stage ever reaches
// this service, so Selectors is always non-empty here.
func insertRouteStages(ctx context.Context, tx *sql.Tx, tenantID, routeID string, stages []domain.Stage) error {
	if len(stages) == 0 {
		return nil
	}
	const colsPerRow = 9
	placeholders := make([]string, 0, len(stages))
	args := make([]any, 0, len(stages)*colsPerRow)
	for i, st := range stages {
		base := i*colsPerRow + 1
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8,
		))
		// stage_kind is NOT NULL DEFAULT 'approval'; an empty domain zero-value
		// must not clobber the DB default with an empty string (F1 migration 0286).
		kind := st.Kind
		if kind == "" {
			kind = domain.StageKindApproval
		}
		args = append(args,
			routeID,
			st.Order,
			st.Name,
			st.RequiredCapability,
			st.Quorum,
			st.QuorumM,
			st.OnEligibilityDrift,
			string(kind),
			st.DueInDays,
		)
	}
	query := `
		INSERT INTO approval_route_stages
			(route_id, stage_order, name, required_capability, quorum, quorum_m, on_eligibility_drift, stage_kind, due_in_days)
		VALUES ` + strings.Join(placeholders, ",")
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("route_admin: insert stages: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}

	if err := insertRouteStageSelectors(ctx, tx, tenantID, routeID, stages); err != nil {
		return err
	}
	return nil
}

// insertRouteStageSelectors writes the approval_route_stage_selectors rows
// for every stage's Selectors. It correlates each selector row to its parent
// approval_route_stages row entirely in SQL — a single INSERT ... SELECT ...
// JOIN keyed on (route_id, stage_order) — rather than a round-trip to fetch
// the just-inserted stage ids, since the batched stage INSERT above has no
// RETURNING clause.
func insertRouteStageSelectors(ctx context.Context, tx *sql.Tx, tenantID, routeID string, stages []domain.Stage) error {
	type selectorRow struct {
		stageOrder int
		order      int
		kind       string
		userID     any
		role       any
		areaCode   any
	}

	var rows []selectorRow
	for _, st := range stages {
		for i, sel := range st.Selectors {
			rows = append(rows, selectorRow{
				stageOrder: st.Order,
				order:      i + 1,
				kind:       string(sel.Kind),
				userID:     nullableString(sel.UserID),
				role:       nullableString(sel.Role),
				areaCode:   nullableString(sel.AreaCode),
			})
		}
	}
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 6
	placeholders := make([]string, 0, len(rows))
	args := make([]any, 0, len(rows)*colsPerRow+2)
	args = append(args, tenantID)
	for i, rr := range rows {
		base := i*colsPerRow + 2
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d::int,$%d::int,$%d::text,$%d::text,$%d::text,$%d::text)",
			base, base+1, base+2, base+3, base+4, base+5,
		))
		args = append(args, rr.stageOrder, rr.order, rr.kind, rr.userID, rr.role, rr.areaCode)
	}
	routeIDPlaceholder := len(args) + 1
	args = append(args, routeID)

	query := fmt.Sprintf(`
		INSERT INTO approval_route_stage_selectors
			(tenant_id, route_stage_id, selector_order, kind, user_id, role, area_code)
		SELECT $1::uuid, s.id, v.selector_order, v.kind, v.user_id, v.role, v.area_code
		  FROM approval_route_stages s
		  JOIN (VALUES %s) AS v(stage_order, selector_order, kind, user_id, role, area_code)
		    ON s.stage_order = v.stage_order
		 WHERE s.route_id = $%d::uuid`,
		strings.Join(placeholders, ","), routeIDPlaceholder,
	)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("route_admin: insert stage selectors: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}
	return nil
}

func loadRouteStagesTx(ctx context.Context, tx *sql.Tx, routeID string) ([]domain.Stage, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, stage_order, name, required_capability, quorum, quorum_m, on_eligibility_drift, stage_kind, due_in_days
		  FROM approval_route_stages
		 WHERE route_id = $1
		 ORDER BY stage_order ASC`,
		routeID,
	)
	if err != nil {
		return nil, fmt.Errorf("route_admin: load stages: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
	}
	defer rows.Close()

	var out []domain.Stage
	var stageIDs []string
	for rows.Next() {
		var (
			st        domain.Stage
			stageID   string
			quorumM   sql.NullInt64
			kindStr   string
			dueInDays sql.NullInt64
		)
		if err := rows.Scan(&stageID, &st.Order, &st.Name, &st.RequiredCapability, &st.Quorum, &quorumM, &st.OnEligibilityDrift, &kindStr, &dueInDays); err != nil {
			return nil, fmt.Errorf("route_admin: scan stage: %w", err)
		}
		if quorumM.Valid {
			m := int(quorumM.Int64)
			st.QuorumM = &m
		}
		st.Kind = domain.StageKind(kindStr)
		if dueInDays.Valid {
			d := int(dueInDays.Int64)
			st.DueInDays = &d
		}
		out = append(out, st)
		stageIDs = append(stageIDs, stageID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("route_admin: iterate stages: %w", err)
	}

	// Selectors is the sole source of truth (slice 6b); load them here too so
	// stagesEqual's DeepEqual comparison against in.Stages (whose Selectors is
	// always populated by the HTTP boundary synthesis) is meaningful rather
	// than spuriously "changed" on every update.
	if len(stageIDs) > 0 {
		selRows, err := tx.QueryContext(ctx, `
			SELECT route_stage_id, kind, user_id, role, area_code
			  FROM approval_route_stage_selectors
			 WHERE route_stage_id = ANY($1)
			 ORDER BY route_stage_id, selector_order ASC`,
			pq.Array(stageIDs),
		)
		if err != nil {
			return nil, fmt.Errorf("route_admin: load stage selectors: %w", infrastructure.MapPgError(err, infrastructure.MapHints{}))
		}
		defer selRows.Close()

		selectorsByStageID := make(map[string][]domain.ActorSelector, len(stageIDs))
		for selRows.Next() {
			var stageID, kind string
			var userID, role, areaCode sql.NullString
			if err := selRows.Scan(&stageID, &kind, &userID, &role, &areaCode); err != nil {
				return nil, fmt.Errorf("route_admin: scan stage selector: %w", err)
			}
			sel := domain.ActorSelector{Kind: domain.SelectorKind(kind)}
			if userID.Valid {
				sel.UserID = userID.String
			}
			if role.Valid {
				sel.Role = role.String
			}
			if areaCode.Valid {
				sel.AreaCode = areaCode.String
			}
			selectorsByStageID[stageID] = append(selectorsByStageID[stageID], sel)
		}
		if err := selRows.Err(); err != nil {
			return nil, fmt.Errorf("route_admin: iterate stage selectors: %w", err)
		}
		for i, stageID := range stageIDs {
			out[i].Selectors = selectorsByStageID[stageID]
		}
	}

	return out, nil
}

// stagesEqual compares stage sets for the update-diff decision. Kind is
// normalized to its DB default (StageKindApproval) before comparing so a
// caller that omits Kind (its Go zero value, "") is never spuriously
// considered "changed" against a stage loaded back from the DB, which always
// reads a concrete value thanks to the NOT NULL DEFAULT (migration 0286). No
// other field changed shape, so DeepEqual remains the comparison for those.
func stagesEqual(a, b []domain.Stage) bool {
	norm := func(in []domain.Stage) []domain.Stage {
		out := make([]domain.Stage, len(in))
		copy(out, in)
		for i := range out {
			if out[i].Kind == "" {
				out[i].Kind = domain.StageKindApproval
			}
		}
		return out
	}
	return reflect.DeepEqual(norm(a), norm(b))
}
