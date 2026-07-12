package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// Ensure the postgres implementation satisfies the full interface at compile time.
var _ ApprovalRepository = (*postgresApprovalRepository)(nil)

type postgresApprovalRepository struct {
	db          *sql.DB
	displayName iamdomain.UserDisplayNameReader
}

// ErrInvalidScheduledSupersedeTarget is returned when a schedule-publish
// request names a supersede target that does not exist or is not published.
var ErrInvalidScheduledSupersedeTarget = errors.New("approval: invalid scheduled supersede target")

// NewPostgresApprovalRepository constructs a production Postgres-backed ApprovalRepository.
// displayName is a required collaborator — use iamdomain.NoopUserDisplayNameReader{} for
// paths that do not need display-name resolution (e.g. scheduled-publish jobs that never
// call LoadActorDisplayName). Returns the ApprovalRepository interface so the concrete
// type is not exposed to callers.
func NewPostgresApprovalRepository(db *sql.DB, displayName iamdomain.UserDisplayNameReader) ApprovalRepository {
	return &postgresApprovalRepository{db: db, displayName: displayName}
}

// InsertInstance writes a new approval_instances row within the caller's transaction.
func (r *postgresApprovalRepository) InsertInstance(ctx context.Context, tx db.Tx, inst domain.Instance) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO approval_instances
		  (id, tenant_id, document_id, route_id, route_version_snapshot,
		   status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		   frozen_content_hash, cancel_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		inst.ID,
		inst.TenantID,
		inst.DocumentID,
		inst.RouteID,
		inst.RouteVersionSnapshot,
		string(inst.Status),
		inst.SubmittedBy,
		inst.SubmittedAt,
		inst.ContentHashAtSubmit,
		inst.IdempotencyKey,
		inst.FrozenContentHash,
		inst.CancelReason,
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	return nil
}

// InsertStageInstances bulk-inserts all stage instances for an approval in one round-trip.
func (r *postgresApprovalRepository) InsertStageInstances(ctx context.Context, tx db.Tx, stages []domain.StageInstance) error {
	if len(stages) == 0 {
		return nil
	}

	// Build multi-row VALUES clause.
	const colCount = 16
	placeholders := make([]string, 0, len(stages))
	args := make([]any, 0, len(stages)*colCount)

	for i, s := range stages {
		base := i * colCount
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,%s)",
			base+1, base+2, base+3, base+4, base+5,
			base+6, base+7, base+8, base+9, base+10,
			base+11, base+12, base+13, base+14, base+15,
			// due_at (F8, spec.md §4/W4): computed here, not passed as a bind
			// param, so the DB's own clock is authoritative for the SLA start
			// point — mirrors UpdateStageStatus's activation CASE, which uses
			// the identical now()+interval expression for stages that become
			// active later via AdvanceStage rather than at initial insert.
			// Only the stage inserted as already-active (stage 0, submit path)
			// needs a due_at computed at INSERT time; pending stages have no
			// opened_at yet, so due_at stays NULL until their own activation
			// UPDATE runs.
			// $base+16::int casts: DueInDaysSnapshot is only ever referenced
			// inside this CASE expression, never bound to a typed column, so
			// when a stage has no SLA configured (*int is nil -> untyped
			// NULL), Postgres cannot infer the param's type and the whole
			// multi-row INSERT fails with "could not determine data type of
			// parameter" (SQLSTATE 42P08). The explicit ::int annotation
			// gives Postgres a type for the NULL case; (n || ' days')::interval
			// semantics are unaffected since int concatenated with text still
			// yields the same text before the interval cast.
			fmt.Sprintf("CASE WHEN $%d = 'active' AND $%d::int IS NOT NULL THEN now() + ($%d::int || ' days')::interval ELSE NULL END", base+13, base+16, base+16),
		))

		eligibleJSON, err := json.Marshal(s.EligibleActorIDs)
		if err != nil {
			return fmt.Errorf("marshal eligible_actor_ids for stage %s: %w", s.ID, err)
		}

		// stage_kind is NOT NULL DEFAULT 'approval'; an empty domain zero-value
		// must not clobber the DB default with an empty string.
		kind := s.Kind
		if kind == "" {
			kind = domain.StageKindApproval
		}

		args = append(args,
			s.ID,
			s.ApprovalInstanceID,
			s.StageOrder,
			s.NameSnapshot,
			s.RequiredRoleSnapshot,
			s.RequiredCapabilitySnapshot,
			s.AreaCodeSnapshot,
			string(s.QuorumSnapshot),
			s.QuorumMSnapshot,
			string(s.OnEligibilityDriftSnapshot),
			eligibleJSON,
			s.EffectiveDenominator,
			string(s.Status),
			string(kind),
			s.DueInDaysSnapshot,
			s.DueInDaysSnapshot,
		)
	}

	query := `INSERT INTO approval_stage_instances
		(id, approval_instance_id, stage_order, name_snapshot,
		 required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		 quorum_snapshot, quorum_m_snapshot, on_eligibility_drift_snapshot,
		 eligible_actor_ids, effective_denominator, status, stage_kind,
		 due_in_days_snapshot, due_at)
		VALUES ` + strings.Join(placeholders, ",")

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	return nil
}

// InsertSignoff inserts a signoff with ON CONFLICT DO NOTHING.
// If the row already exists it calls LoadSignoffByActor to compare fields.
// Matching fields → WasReplay=true. Mismatching fields → ErrActorAlreadySigned.
func (r *postgresApprovalRepository) InsertSignoff(ctx context.Context, tx db.Tx, s domain.Signoff) (SignoffInsertResult, error) {
	payload := s.SignaturePayload()
	if payload == nil {
		payload = json.RawMessage("{}")
	}

	// actor_display_name_snapshot is a DB-enforced invariant (NOT NULL + CHECK
	// (<> ''), migration 0294 / ADR 0079). Bind it UNCONDITIONALLY — an empty
	// snapshot fails closed at the DB (eQMS audit-truth), never silently NULL.
	actorDisplayNameSnapshot := s.ActorDisplayNameSnapshot()

	var onBehalfOf sql.NullString
	if v := s.OnBehalfOf(); v != "" {
		onBehalfOf = sql.NullString{String: v, Valid: true}
	}

	var returnedID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO approval_signoffs
		  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
		   decision, comment, signed_at, signature_method, signature_payload, content_hash,
		   actor_display_name_snapshot, signature_meaning, on_behalf_of_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (stage_instance_id, actor_user_id) DO NOTHING
		RETURNING id`,
		s.ID(),
		s.ApprovalInstanceID(),
		s.StageInstanceID(),
		s.ActorUserID(),
		s.ActorTenantID(),
		string(s.Decision()),
		s.Comment(),
		s.SignedAt(),
		s.SignatureMethod(),
		payload,
		s.ContentHash(),
		actorDisplayNameSnapshot,
		s.SignatureMeaning(),
		onBehalfOf,
	).Scan(&returnedID)

	if err == nil {
		// Fresh insert — RETURNING produced a row.
		return SignoffInsertResult{ID: returnedID, WasReplay: false}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// A real DB error (constraint check, PgError, etc.)
		return SignoffInsertResult{}, MapPgError(err, MapHints{})
	}

	// ON CONFLICT fired — RETURNING was empty. Load the existing signoff.
	existing, loadErr := r.loadSignoffByStageActor(ctx, tx, s.ActorTenantID(), s.StageInstanceID(), s.ActorUserID())
	if loadErr != nil {
		return SignoffInsertResult{}, fmt.Errorf("load existing signoff for replay check: %w", loadErr)
	}

	// Replay: same stage, same decision, same content_hash.
	if existing.StageInstanceID() == s.StageInstanceID() &&
		existing.Decision() == s.Decision() &&
		existing.ContentHash() == s.ContentHash() {
		return SignoffInsertResult{ID: existing.ID(), WasReplay: true}, nil
	}

	// Different fields — actor already signed with different parameters.
	return SignoffInsertResult{}, ErrActorAlreadySigned
}

// LoadSignoffByActor loads a signoff by (tenantID, instanceID, actorUserID).
// Returns nil, ErrActorAlreadySigned if not found (caller decides on semantics).
func (r *postgresApprovalRepository) LoadSignoffByActor(ctx context.Context, tx db.Tx, tenantID, instanceID, actorUserID string) (*domain.Signoff, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id, s.actor_user_id,
		       s.actor_tenant_id, s.decision, coalesce(s.comment,''), s.signed_at,
		       s.signature_method, s.signature_payload, s.content_hash, s.signature_meaning,
		       coalesce(s.on_behalf_of_user_id,'')
		FROM approval_signoffs s
		JOIN approval_instances i ON i.id = s.approval_instance_id
		WHERE s.approval_instance_id = $1
		  AND s.actor_user_id = $2
		  AND i.tenant_id = $3`,
		instanceID, actorUserID, tenantID,
	)
	return scanSignoff(row)
}

func (r *postgresApprovalRepository) loadSignoffByStageActor(ctx context.Context, tx db.Tx, tenantID, stageInstanceID, actorUserID string) (*domain.Signoff, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id, s.actor_user_id,
		       s.actor_tenant_id, s.decision, coalesce(s.comment,''), s.signed_at,
		       s.signature_method, s.signature_payload, s.content_hash, s.signature_meaning,
		       coalesce(s.on_behalf_of_user_id,'')
		FROM approval_signoffs s
		JOIN approval_instances i ON i.id = s.approval_instance_id
		WHERE s.stage_instance_id = $1
		  AND s.actor_user_id = $2
		  AND i.tenant_id = $3::uuid`,
		stageInstanceID, actorUserID, tenantID,
	)
	return scanSignoff(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSignoff(row rowScanner) (*domain.Signoff, error) {
	var (
		id, instanceID, stageID, actorUserID, actorTenantID string
		decision, comment, signatureMethod, contentHash     string
		signatureMeaning                                    string
		onBehalfOf                                           string
		signedAt                                            time.Time
		sigPayload                                          []byte
	)
	err := row.Scan(&id, &instanceID, &stageID, &actorUserID, &actorTenantID,
		&decision, &comment, &signedAt, &signatureMethod, &sigPayload, &contentHash, &signatureMeaning,
		&onBehalfOf)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return domain.NewSignoff(domain.SignoffParams{
		ID:                 id,
		ApprovalInstanceID: instanceID,
		StageInstanceID:    stageID,
		ActorUserID:        actorUserID,
		ActorTenantID:      actorTenantID,
		Decision:           domain.Decision(decision),
		Comment:            comment,
		OnBehalfOfUserID:   onBehalfOf,
		SignedAt:           signedAt,
		SignatureMethod:    signatureMethod,
		SignaturePayload:   json.RawMessage(sigPayload),
		ContentHash:        contentHash,
		SignatureMeaning:   signatureMeaning,
	})
}

// LoadInstance loads an approval instance and its stage instances by ID.
// Returns ErrNoActiveInstance if not found or tenant mismatch.
func (r *postgresApprovalRepository) LoadInstance(ctx context.Context, tx db.Tx, tenantID, id string) (*domain.Instance, error) {
	var inst domain.Instance
	var completedAt sql.NullTime
	var frozenContentHash, cancelReason sql.NullString

	err := tx.QueryRowContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key,
		       ai.frozen_content_hash, ai.cancel_reason
		FROM approval_instances ai
		JOIN documents d
		  ON d.id = ai.document_id
		 AND d.tenant_id = ai.tenant_id
		WHERE ai.id = $1 AND ai.tenant_id = $2`,
		id, tenantID,
	).Scan(
		&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.RouteID, &inst.RouteVersionSnapshot,
		&inst.RevisionVersion,
		&inst.Status, &inst.SubmittedBy, &inst.SubmittedAt, &completedAt,
		&inst.ContentHashAtSubmit, &inst.IdempotencyKey,
		&frozenContentHash, &cancelReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoActiveInstance
	}
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}
	if frozenContentHash.Valid {
		inst.FrozenContentHash = &frozenContentHash.String
	}
	if cancelReason.Valid {
		inst.CancelReason = &cancelReason.String
	}

	stages, err := r.loadStageInstances(ctx, tx, tenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	inst.Stages = stages
	return &inst, nil
}

// LoadActiveInstanceByDocument loads the single in_progress instance for a document.
// Returns ErrNoActiveInstance when none exists or tenant doesn't match.
func (r *postgresApprovalRepository) LoadActiveInstanceByDocument(ctx context.Context, tx db.Tx, tenantID, docID string) (*domain.Instance, error) {
	var inst domain.Instance
	var completedAt sql.NullTime
	var frozenContentHash, cancelReason sql.NullString

	err := tx.QueryRowContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key,
		       ai.frozen_content_hash, ai.cancel_reason
		FROM approval_instances ai
		JOIN documents d
		  ON d.id = ai.document_id
		 AND d.tenant_id = ai.tenant_id
		WHERE ai.document_id = $1
		  AND ai.tenant_id = $2
		  AND ai.status IN ('in_progress', 'approved')
			ORDER BY ai.submitted_at DESC, ai.id DESC
			LIMIT 1`,
		docID, tenantID,
	).Scan(
		&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.RouteID, &inst.RouteVersionSnapshot,
		&inst.RevisionVersion,
		&inst.Status, &inst.SubmittedBy, &inst.SubmittedAt, &completedAt,
		&inst.ContentHashAtSubmit, &inst.IdempotencyKey,
		&frozenContentHash, &cancelReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoActiveInstance
	}
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}
	if frozenContentHash.Valid {
		inst.FrozenContentHash = &frozenContentHash.String
	}
	if cancelReason.Valid {
		inst.CancelReason = &cancelReason.String
	}

	stages, err := r.loadStageInstances(ctx, tx, tenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	inst.Stages = stages
	return &inst, nil
}

// LoadInstanceByDocumentForView is LoadActiveInstanceByDocument's view-only
// sibling (see the interface doc comment in approval_repository.go): identical
// in every respect except the status filter also admits changes_requested.
func (r *postgresApprovalRepository) LoadInstanceByDocumentForView(ctx context.Context, tx db.Tx, tenantID, docID string) (*domain.Instance, error) {
	var inst domain.Instance
	var completedAt sql.NullTime
	var frozenContentHash, cancelReason sql.NullString

	err := tx.QueryRowContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key,
		       ai.frozen_content_hash, ai.cancel_reason
		FROM approval_instances ai
		JOIN documents d
		  ON d.id = ai.document_id
		 AND d.tenant_id = ai.tenant_id
		WHERE ai.document_id = $1
		  AND ai.tenant_id = $2
		  AND ai.status IN ('in_progress', 'approved', 'changes_requested')
			ORDER BY ai.submitted_at DESC, ai.id DESC
			LIMIT 1`,
		docID, tenantID,
	).Scan(
		&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.RouteID, &inst.RouteVersionSnapshot,
		&inst.RevisionVersion,
		&inst.Status, &inst.SubmittedBy, &inst.SubmittedAt, &completedAt,
		&inst.ContentHashAtSubmit, &inst.IdempotencyKey,
		&frozenContentHash, &cancelReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoActiveInstance
	}
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}
	if frozenContentHash.Valid {
		inst.FrozenContentHash = &frozenContentHash.String
	}
	if cancelReason.Valid {
		inst.CancelReason = &cancelReason.String
	}

	stages, err := r.loadStageInstances(ctx, tx, tenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	inst.Stages = stages
	return &inst, nil
}

func (r *postgresApprovalRepository) ValidateScheduledSupersedeTarget(ctx context.Context, tx db.Tx, tenantID, documentID, supersededDocumentID string) error {
	var valid bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM documents candidate
			  JOIN documents target
			    ON target.id = $2
			   AND target.tenant_id = $1
			 WHERE candidate.id = $3
			   AND candidate.tenant_id = $1
			   AND candidate.status = 'published'
			   AND candidate.controlled_document_id = target.controlled_document_id
		)`,
		tenantID, documentID, supersededDocumentID,
	).Scan(&valid)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	if !valid {
		return ErrInvalidScheduledSupersedeTarget
	}
	return nil
}

func (r *postgresApprovalRepository) LoadCurrentPublishedHeadForDocument(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error) {
	var publishedDocumentID string
	err := tx.QueryRowContext(ctx, `
		SELECT current_head.id
		  FROM documents target
		  JOIN documents current_head
		    ON current_head.tenant_id = target.tenant_id
		   AND current_head.controlled_document_id = target.controlled_document_id
		   AND current_head.status = 'published'
		 WHERE target.tenant_id = $1
		   AND target.id = $2
		 ORDER BY current_head.revision_number DESC
		 LIMIT 1
		 FOR UPDATE OF current_head`,
		tenantID, documentID,
	).Scan(&publishedDocumentID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", MapPgError(err, MapHints{})
	}
	return publishedDocumentID, nil
}

func (r *postgresApprovalRepository) LoadCurrentPublishedHead(ctx context.Context, tx db.Tx, tenantID, controlledDocumentID string) (string, error) {
	var documentID string
	err := tx.QueryRowContext(ctx, `
		SELECT id
		  FROM documents
		 WHERE tenant_id = $1
		   AND controlled_document_id = $2
		   AND status = 'published'
		 ORDER BY revision_number DESC
		 LIMIT 1
		 FOR UPDATE`,
		tenantID, controlledDocumentID,
	).Scan(&documentID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", MapPgError(err, MapHints{})
	}
	return documentID, nil
}

func (r *postgresApprovalRepository) GetDocumentRevisionVersion(ctx context.Context, tx db.Tx, documentID, tenantID string) (int, error) {
	var revisionVersion int
	err := tx.QueryRowContext(ctx, `
		SELECT revision_version
		  FROM documents
		 WHERE id = $1
		   AND tenant_id = $2::uuid
		 FOR UPDATE`,
		documentID, tenantID,
	).Scan(&revisionVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrStaleRevision
	}
	if err != nil {
		return 0, MapPgError(err, MapHints{})
	}
	return revisionVersion, nil
}

// LoadGovernedRevisionNumber reads documents.revision_number inside the
// caller's transaction (T8b). Unlike GetDocumentRevisionVersion this is a
// plain read (no FOR UPDATE) — the submit path only needs the current
// governed value to gate the reason/title normalizers and bind the content
// hash; it does not CAS on revision_number.
func (r *postgresApprovalRepository) LoadGovernedRevisionNumber(ctx context.Context, tx db.Tx, tenantID, documentID string) (int, error) {
	var revisionNumber int64
	err := tx.QueryRowContext(ctx, `
		SELECT revision_number
		  FROM documents
		 WHERE id = $1
		   AND tenant_id = $2::uuid`,
		documentID, tenantID,
	).Scan(&revisionNumber)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrStaleRevision
	}
	if err != nil {
		return 0, MapPgError(err, MapHints{})
	}
	return int(revisionNumber), nil
}

// ListRoutes loads tenant-scoped route configuration and stages in one joined query.
func (r *postgresApprovalRepository) ListRoutes(ctx context.Context, tenantID string) ([]Route, error) {
	rows, err := r.db.QueryContext(ctx, listRoutesQuery, tenantID)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()
	return scanRouteListRows(rows)
}

// ListRoutesTx is the transaction-scoped variant used by the route admin
// service so the tenant GUC set on the transaction governs row visibility
// (single-tx authz + read; no TOCTOU).
func (r *postgresApprovalRepository) ListRoutesTx(ctx context.Context, tx db.Tx, tenantID string) ([]Route, error) {
	rows, err := tx.QueryContext(ctx, listRoutesQuery, tenantID)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()
	return scanRouteListRows(rows)
}

// LoadActorDisplayName returns the approver's display name via the iam-owned
// UserDisplayNameReader port (M4/F4.1). The read runs on the pool — never inside
// the caller's signoff advisory-lock transaction (H-PRE-1 unchanged). Returns ""
// when the user is absent (best-effort snapshot, same semantics as before).
func (r *postgresApprovalRepository) LoadActorDisplayName(ctx context.Context, tenantID, userID string) (string, error) {
	return r.displayName.DisplayName(ctx, tenantID, userID)
}

const listRoutesQuery = `
		SELECT r.id, r.name, r.tenant_id::text, r.profile_code, r.active, r.version, r.created_at, r.created_at AS updated_at,
		       s.stage_order, s.name, s.required_role, s.required_capability, s.area_code, s.quorum, s.quorum_m, s.on_eligibility_drift,
		       s.stage_kind, s.due_in_days,
		       (SELECT COUNT(*) FROM approval_routes WHERE tenant_id = $1::uuid) AS total_count
		  FROM approval_routes r
		  JOIN approval_route_stages s
		    ON s.route_id = r.id
		 WHERE r.tenant_id = $1::uuid
		 ORDER BY r.created_at DESC, s.stage_order ASC`

func scanRouteListRows(rows *sql.Rows) ([]Route, error) {

	routeMap := make(map[string]*Route)
	var routeOrder []string
	for rows.Next() {
		var (
			routeID, routeName, routeTenantID, profileCode string
			active                                         bool
			version                                        int
			createdAt, updatedAt                           time.Time
			stage                                          RouteStage
			stageName, stageRole, stageCapability          sql.NullString
			stageArea, stageQuorum, stageDrift             sql.NullString
			stageQuorumM                                   sql.NullInt64
			stageKind                                      sql.NullString
			stageDueInDays                                 sql.NullInt64
			totalCount                                     int64
		)
		if err := rows.Scan(
			&routeID, &routeName, &routeTenantID, &profileCode, &active, &version, &createdAt, &updatedAt,
			&stage.Order, &stageName, &stageRole, &stageCapability, &stageArea, &stageQuorum, &stageQuorumM, &stageDrift,
			&stageKind, &stageDueInDays,
			&totalCount,
		); err != nil {
			return nil, fmt.Errorf("scan approval route list row: %w", err)
		}

		route := routeMap[routeID]
		if route == nil {
			route = &Route{
				ID:          routeID,
				Name:        routeName,
				TenantID:    routeTenantID,
				ProfileCode: profileCode,
				Active:      active,
				Version:     version,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				Stages:      []RouteStage{},
				Total:       int(totalCount),
			}
			routeMap[routeID] = route
			routeOrder = append(routeOrder, routeID)
		}

		if stageName.Valid {
			stage.Name = stageName.String
		}
		if stageRole.Valid {
			stage.RequiredRole = stageRole.String
		}
		if stageCapability.Valid {
			stage.RequiredCapability = stageCapability.String
		}
		if stageArea.Valid {
			stage.AreaCode = stageArea.String
		}
		if stageQuorum.Valid {
			stage.Quorum = stageQuorum.String
		}
		if stageQuorumM.Valid {
			m := int(stageQuorumM.Int64)
			stage.QuorumM = &m
		}
		if stageDrift.Valid {
			stage.DriftPolicy = stageDrift.String
		}
		if stageKind.Valid {
			stage.Kind = stageKind.String
		}
		if stageDueInDays.Valid {
			v := int(stageDueInDays.Int64)
			stage.DueInDays = &v
		}
		route.Stages = append(route.Stages, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approval route list rows: %w", err)
	}

	routes := make([]Route, 0, len(routeOrder))
	for _, routeID := range routeOrder {
		routes = append(routes, *routeMap[routeID])
	}
	return routes, nil
}

func (r *postgresApprovalRepository) MarkSuperseded(ctx context.Context, tx db.Tx, tenantID, documentID string) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET status           = 'superseded',
		       revision_version = revision_version + 1
		 WHERE id        = $1
		   AND tenant_id = $2
		   AND status    = 'published'`,
		documentID, tenantID,
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrScheduledSupersedeConflict
	}
	return nil
}

// loadStageInstances loads all stage instances for a given approval instance, ordered by stage_order.
func (r *postgresApprovalRepository) loadStageInstances(ctx context.Context, tx db.Tx, tenantID, instanceID string) ([]domain.StageInstance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_order, name_snapshot,
		       required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		       quorum_snapshot, quorum_m_snapshot,
		       on_eligibility_drift_snapshot,
		       eligible_actor_ids, effective_denominator,
		       status, opened_at, completed_at, skip_reason,
		       stage_kind, due_at, due_in_days_snapshot
		FROM approval_stage_instances
		WHERE approval_instance_id = $1
		  AND EXISTS (
		      SELECT 1
		        FROM approval_instances ai
		       WHERE ai.id = $1
		         AND ai.tenant_id = $2::uuid
		  )
		-- FOR UPDATE: prevents re-submit from re-snapshotting eligible_actor_ids during signoff (J1).
		ORDER BY stage_order ASC
		FOR UPDATE`,
		instanceID, tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	var stages []domain.StageInstance
	for rows.Next() {
		var s domain.StageInstance
		var quorumMSnapshot sql.NullInt32
		var effectiveDenominator sql.NullInt32
		var openedAt, completedAt, dueAt sql.NullTime
		var eligibleJSON []byte
		var skipReason sql.NullString
		var kindStr string
		var dueInDaysSnapshot sql.NullInt32

		err := rows.Scan(
			&s.ID, &s.ApprovalInstanceID, &s.StageOrder, &s.NameSnapshot,
			&s.RequiredRoleSnapshot, &s.RequiredCapabilitySnapshot, &s.AreaCodeSnapshot,
			&s.QuorumSnapshot, &quorumMSnapshot,
			&s.OnEligibilityDriftSnapshot,
			&eligibleJSON, &effectiveDenominator,
			&s.Status, &openedAt, &completedAt, &skipReason,
			&kindStr, &dueAt, &dueInDaysSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("scan stage instance for approval instance %s: %w", instanceID, err)
		}

		if quorumMSnapshot.Valid {
			v := int(quorumMSnapshot.Int32)
			s.QuorumMSnapshot = &v
		}
		if effectiveDenominator.Valid {
			v := int(effectiveDenominator.Int32)
			s.EffectiveDenominator = &v
		}
		if openedAt.Valid {
			s.OpenedAt = &openedAt.Time
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		if skipReason.Valid {
			s.SkipReason = skipReason.String
		}
		s.Kind = domain.StageKind(kindStr)
		if dueAt.Valid {
			s.DueAt = &dueAt.Time
		}
		if dueInDaysSnapshot.Valid {
			v := int(dueInDaysSnapshot.Int32)
			s.DueInDaysSnapshot = &v
		}

		if len(eligibleJSON) > 0 {
			if err := json.Unmarshal(eligibleJSON, &s.EligibleActorIDs); err != nil {
				return nil, fmt.Errorf("unmarshal eligible_actor_ids for stage %s: %w", s.ID, err)
			}
		}

		stages = append(stages, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stage instances for approval instance %s: %w", instanceID, err)
	}

	// Load signoffs for the instance and attach to stages.
	byStage, err := r.loadSignoffsForInstance(ctx, tx, tenantID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("load signoffs for approval instance %s: %w", instanceID, err)
	}
	for i := range stages {
		if sigs, ok := byStage[stages[i].ID]; ok {
			stages[i].Signoffs = sigs
		}
	}
	return stages, nil
}

// loadSignoffsForInstance fetches all signoffs for an approval instance,
// keyed by stage_instance_id.
func (r *postgresApprovalRepository) loadSignoffsForInstance(ctx context.Context, tx db.Tx, tenantID, instanceID string) (map[string][]*domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id, actor_user_id,
		       actor_tenant_id, decision, coalesce(comment,''), signed_at,
		       signature_method, signature_payload, content_hash,
		       actor_display_name_snapshot, signature_meaning
		FROM approval_signoffs
		WHERE approval_instance_id = $1
		  AND EXISTS (
		      SELECT 1
		        FROM approval_instances ai
		       WHERE ai.id = $1
		         AND ai.tenant_id = $2::uuid
		  )
		ORDER BY signed_at ASC`,
		instanceID, tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	out := map[string][]*domain.Signoff{}
	for rows.Next() {
		var (
			id, instID, stageID, actorUserID, actorTenantID string
			decision, comment, signatureMethod, contentHash string
			displayName                                     string
			signatureMeaning                                string
			signedAt                                        time.Time
			sigPayload                                      []byte
		)
		if err := rows.Scan(&id, &instID, &stageID, &actorUserID, &actorTenantID,
			&decision, &comment, &signedAt, &signatureMethod, &sigPayload,
			&contentHash, &displayName, &signatureMeaning); err != nil {
			return nil, fmt.Errorf("scan signoff row for approval instance %s: %w", instanceID, err)
		}
		sig, err := domain.NewSignoff(domain.SignoffParams{
			ID:                       id,
			ApprovalInstanceID:       instID,
			StageInstanceID:          stageID,
			ActorUserID:              actorUserID,
			ActorTenantID:            actorTenantID,
			Decision:                 domain.Decision(decision),
			Comment:                  comment,
			SignedAt:                 signedAt,
			SignatureMethod:          signatureMethod,
			SignaturePayload:         sigPayload,
			ContentHash:              contentHash,
			ActorDisplayNameSnapshot: displayName,
			SignatureMeaning:         signatureMeaning,
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s: %w", id, err)
		}
		out[stageID] = append(out[stageID], sig)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signoff rows for approval instance %s: %w", instanceID, err)
	}
	return out, nil
}

// LoadInstancesByIDs batch-loads approval instances by their IDs.
// Uses a single query for headers, one for all stage instances, and one for all
// signoffs — 3 round trips instead of 3N. Order matches ids; missing IDs are
// silently omitted (tenant mismatch or not found).
func (r *postgresApprovalRepository) LoadInstancesByIDs(ctx context.Context, tx db.Tx, tenantID string, ids []string) ([]domain.Instance, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Dedupe ids preserving first-occurrence order so duplicate inputs cannot
	// produce duplicate rows in the result.
	seen := make(map[string]struct{}, len(ids))
	deduped := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			deduped = append(deduped, id)
		}
	}
	ids = deduped

	// ── 1. Load instance headers ──────────────────────────────────────────────
	headerRows, err := tx.QueryContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key,
		       ai.frozen_content_hash, ai.cancel_reason
		FROM approval_instances ai
		JOIN documents d
		  ON d.id = ai.document_id
		 AND d.tenant_id = ai.tenant_id
		WHERE ai.id = ANY($1)
		  AND ai.tenant_id = $2`,
		pq.Array(ids), tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer headerRows.Close()

	// Maintain insertion order from ids slice (deduped above).
	byID := make(map[string]*domain.Instance, len(ids))
	for headerRows.Next() {
		var inst domain.Instance
		var completedAt sql.NullTime
		var frozenContentHash, cancelReason sql.NullString
		if err := headerRows.Scan(
			&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.RouteID, &inst.RouteVersionSnapshot,
			&inst.RevisionVersion,
			&inst.Status, &inst.SubmittedBy, &inst.SubmittedAt, &completedAt,
			&inst.ContentHashAtSubmit, &inst.IdempotencyKey,
			&frozenContentHash, &cancelReason,
		); err != nil {
			return nil, fmt.Errorf("scan approval instance header: %w", err)
		}
		if completedAt.Valid {
			inst.CompletedAt = &completedAt.Time
		}
		if frozenContentHash.Valid {
			inst.FrozenContentHash = &frozenContentHash.String
		}
		if cancelReason.Valid {
			inst.CancelReason = &cancelReason.String
		}
		cp := inst
		byID[inst.ID] = &cp
	}
	if err := headerRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approval instance headers: %w", err)
	}
	if len(byID) == 0 {
		return nil, nil
	}

	// ── 2. Load all stage instances for the found set ─────────────────────────
	knownIDs := make([]string, 0, len(byID))
	for id := range byID {
		knownIDs = append(knownIDs, id)
	}
	stageRows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_order, name_snapshot,
		       required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		       quorum_snapshot, quorum_m_snapshot,
		       on_eligibility_drift_snapshot,
		       eligible_actor_ids, effective_denominator,
		       status, opened_at, completed_at, skip_reason,
		       stage_kind, due_at, due_in_days_snapshot
		FROM approval_stage_instances
		WHERE approval_instance_id = ANY($1)
		  AND EXISTS (
		      SELECT 1
		        FROM approval_instances ai
		       WHERE ai.id = approval_stage_instances.approval_instance_id
		         AND ai.tenant_id = $2::uuid
		  )
		ORDER BY approval_instance_id, stage_order ASC
		FOR UPDATE`,
		pq.Array(knownIDs), tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer stageRows.Close()

	stagesByInst := make(map[string][]domain.StageInstance, len(byID))
	for stageRows.Next() {
		var s domain.StageInstance
		var quorumMSnapshot sql.NullInt32
		var effectiveDenominator sql.NullInt32
		var openedAt, completedAt, dueAt sql.NullTime
		var eligibleJSON []byte
		var skipReason sql.NullString
		var kindStr string
		var dueInDaysSnapshot sql.NullInt32
		if err := stageRows.Scan(
			&s.ID, &s.ApprovalInstanceID, &s.StageOrder, &s.NameSnapshot,
			&s.RequiredRoleSnapshot, &s.RequiredCapabilitySnapshot, &s.AreaCodeSnapshot,
			&s.QuorumSnapshot, &quorumMSnapshot,
			&s.OnEligibilityDriftSnapshot,
			&eligibleJSON, &effectiveDenominator,
			&s.Status, &openedAt, &completedAt, &skipReason,
			&kindStr, &dueAt, &dueInDaysSnapshot,
		); err != nil {
			return nil, fmt.Errorf("scan stage instance in batch load: %w", err)
		}
		if quorumMSnapshot.Valid {
			v := int(quorumMSnapshot.Int32)
			s.QuorumMSnapshot = &v
		}
		if effectiveDenominator.Valid {
			v := int(effectiveDenominator.Int32)
			s.EffectiveDenominator = &v
		}
		if openedAt.Valid {
			s.OpenedAt = &openedAt.Time
		}
		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		if skipReason.Valid {
			s.SkipReason = skipReason.String
		}
		s.Kind = domain.StageKind(kindStr)
		if dueAt.Valid {
			s.DueAt = &dueAt.Time
		}
		if dueInDaysSnapshot.Valid {
			v := int(dueInDaysSnapshot.Int32)
			s.DueInDaysSnapshot = &v
		}
		if len(eligibleJSON) > 0 {
			if err := json.Unmarshal(eligibleJSON, &s.EligibleActorIDs); err != nil {
				return nil, fmt.Errorf("unmarshal eligible_actor_ids for stage %s: %w", s.ID, err)
			}
		}
		stagesByInst[s.ApprovalInstanceID] = append(stagesByInst[s.ApprovalInstanceID], s)
	}
	if err := stageRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stage instances in batch load: %w", err)
	}

	// ── 3. Load all signoffs for the found set ────────────────────────────────
	signoffRows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id, actor_user_id,
		       actor_tenant_id, decision, coalesce(comment,''), signed_at,
		       signature_method, signature_payload, content_hash,
		       actor_display_name_snapshot, signature_meaning
		FROM approval_signoffs
		WHERE approval_instance_id = ANY($1)
		  AND EXISTS (
		      SELECT 1
		        FROM approval_instances ai
		       WHERE ai.id = approval_signoffs.approval_instance_id
		         AND ai.tenant_id = $2::uuid
		  )
		ORDER BY signed_at ASC`,
		pq.Array(knownIDs), tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer signoffRows.Close()

	// signoffsByStage: stageInstanceID → []*Signoff
	signoffsByStage := make(map[string][]*domain.Signoff)
	for signoffRows.Next() {
		var (
			id, instID, stageID, actorUserID, actorTenantID string
			decision, comment, signatureMethod, contentHash string
			displayName                                     string
			signatureMeaning                                string
			signedAt                                        time.Time
			sigPayload                                      []byte
		)
		if err := signoffRows.Scan(&id, &instID, &stageID, &actorUserID, &actorTenantID,
			&decision, &comment, &signedAt, &signatureMethod, &sigPayload,
			&contentHash, &displayName, &signatureMeaning); err != nil {
			return nil, fmt.Errorf("scan signoff row in batch load: %w", err)
		}
		sig, err := domain.NewSignoff(domain.SignoffParams{
			ID:                       id,
			ApprovalInstanceID:       instID,
			StageInstanceID:          stageID,
			ActorUserID:              actorUserID,
			ActorTenantID:            actorTenantID,
			Decision:                 domain.Decision(decision),
			Comment:                  comment,
			SignedAt:                 signedAt,
			SignatureMethod:          signatureMethod,
			SignaturePayload:         sigPayload,
			ContentHash:              contentHash,
			ActorDisplayNameSnapshot: displayName,
			SignatureMeaning:         signatureMeaning,
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s in batch load: %w", id, err)
		}
		signoffsByStage[stageID] = append(signoffsByStage[stageID], sig)
	}
	if err := signoffRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signoff rows in batch load: %w", err)
	}

	// ── 4. Assemble in input order ────────────────────────────────────────────
	result := make([]domain.Instance, 0, len(ids))
	for _, id := range ids {
		inst, ok := byID[id]
		if !ok {
			continue
		}
		stages := stagesByInst[id]
		for i := range stages {
			if sigs, ok := signoffsByStage[stages[i].ID]; ok {
				stages[i].Signoffs = sigs
			}
		}
		inst.Stages = stages
		result = append(result, *inst)
	}
	return result, nil
}

// UpdateStageStatus applies an OCC (optimistic concurrency control) UPDATE.
// Checks RowsAffected == 0 — which means expectedOldStatus was not the current value — and returns ErrStageNotActive.
func (r *postgresApprovalRepository) UpdateStageStatus(ctx context.Context, tx db.Tx, tenantID, stageID string, newStatus, expectedOldStatus domain.StageStatus) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE approval_stage_instances asi
		SET status = $1,
		    opened_at    = CASE WHEN $1 = 'active'    THEN now() ELSE asi.opened_at    END,
		    completed_at = CASE WHEN $1 IN ('completed','skipped','rejected_here') THEN now() ELSE asi.completed_at END,
		    -- due_at (F8, spec.md §4/W4): computed from this row's own
		    -- due_in_days_snapshot at the moment it activates, using the DB's
		    -- own now() so the SLA clock's start point never depends on app
		    -- clock skew. Mirrors InsertStageInstances' identical CASE for
		    -- the stage that starts already-active at submit time. A stage
		    -- with no due_in_days_snapshot (nil — no SLA configured) stays
		    -- NULL, never defaulted to a substitute value.
		    due_at = CASE WHEN $1 = 'active' AND asi.due_in_days_snapshot IS NOT NULL
		                  THEN now() + (asi.due_in_days_snapshot || ' days')::interval
		                  ELSE asi.due_at END
		FROM approval_instances ai
		WHERE asi.id = $2
		  AND asi.status = $3
		  AND asi.approval_instance_id = ai.id
		  AND ai.tenant_id = $4`,
		string(newStatus), stageID, string(expectedOldStatus), tenantID,
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("approval: update stage status rows affected: %w", err)
	}
	if n == 0 {
		return ErrStageNotActive
	}
	return nil
}

// UpdateInstanceStatus applies an OCC UPDATE on approval_instances.
// Checks RowsAffected == 0 → ErrInstanceCompleted (stale read or already terminal).
func (r *postgresApprovalRepository) UpdateInstanceStatus(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE approval_instances
		SET status = $1, completed_at = $2
		WHERE id = $3
		  AND tenant_id = $4
		  AND status = $5`,
		string(newStatus), completedAt, instID, tenantID, string(expectedStatus),
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("approval: update instance status rows affected: %w", err)
	}
	if n == 0 {
		return ErrInstanceCompleted
	}
	return nil
}

// UpdateInstanceStatusWithReason is UpdateInstanceStatus plus a cancel_reason
// write (F4): shared by CancelInstance and the request_changes verdict path.
func (r *postgresApprovalRepository) UpdateInstanceStatusWithReason(ctx context.Context, tx db.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time, reason string) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE approval_instances
		SET status = $1, completed_at = $2, cancel_reason = $3
		WHERE id = $4
		  AND tenant_id = $5
		  AND status = $6`,
		string(newStatus), completedAt, reason, instID, tenantID, string(expectedStatus),
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("approval: update instance status with reason rows affected: %w", err)
	}
	if n == 0 {
		return ErrInstanceCompleted
	}
	return nil
}

// InsertVerdict inserts a review-stage runtime verdict with ON CONFLICT DO
// NOTHING (F4), mirroring InsertSignoff exactly. If the row already exists it
// loads the existing verdict to compare fields: matching → WasReplay=true,
// mismatching → ErrActorAlreadySigned (reused sentinel — same SoD-class
// conflict, "actor already recorded a verdict on this stage").
func (r *postgresApprovalRepository) InsertVerdict(ctx context.Context, tx db.Tx, v domain.ReviewVerdict) (VerdictInsertResult, error) {
	// actor_display_name_snapshot is a DB-enforced invariant (NOT NULL + CHECK
	// (<> ''), migration 0294 / ADR 0079). Bind it UNCONDITIONALLY — an empty
	// snapshot now fails closed at the DB (the eQMS audit-truth guarantee) rather
	// than being silently written as NULL. No read fallback exists to mask it.
	actorDisplayNameSnapshot := v.ActorDisplayNameSnapshot()

	var onBehalfOf sql.NullString
	if val := v.OnBehalfOf(); val != "" {
		onBehalfOf = sql.NullString{String: val, Valid: true}
	}

	var returnedID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO approval_review_verdicts
		  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
		   verdict, comment, verdict_at, actor_display_name_snapshot, on_behalf_of_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (stage_instance_id, actor_user_id) DO NOTHING
		RETURNING id`,
		v.ID(),
		v.ApprovalInstanceID(),
		v.StageInstanceID(),
		v.ActorUserID(),
		v.ActorTenantID(),
		string(v.Verdict()),
		v.Comment(),
		v.VerdictAt(),
		actorDisplayNameSnapshot,
		onBehalfOf,
	).Scan(&returnedID)

	if err == nil {
		return VerdictInsertResult{ID: returnedID, WasReplay: false}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return VerdictInsertResult{}, MapPgError(err, MapHints{})
	}

	existing, loadErr := r.loadVerdictByStageActor(ctx, tx, v.ActorTenantID(), v.StageInstanceID(), v.ActorUserID())
	if loadErr != nil {
		return VerdictInsertResult{}, fmt.Errorf("load existing verdict for replay check: %w", loadErr)
	}
	if existing.StageInstanceID() == v.StageInstanceID() &&
		existing.Verdict() == v.Verdict() &&
		existing.Comment() == v.Comment() {
		return VerdictInsertResult{ID: existing.ID(), WasReplay: true}, nil
	}
	return VerdictInsertResult{}, ErrActorAlreadySigned
}

func (r *postgresApprovalRepository) loadVerdictByStageActor(ctx context.Context, tx db.Tx, tenantID, stageInstanceID, actorUserID string) (*domain.ReviewVerdict, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT v.id, v.approval_instance_id, v.stage_instance_id, v.actor_user_id,
		       v.actor_tenant_id, v.verdict, coalesce(v.comment,''), v.verdict_at,
		       v.actor_display_name_snapshot, coalesce(v.on_behalf_of_user_id,'')
		FROM approval_review_verdicts v
		JOIN approval_instances i ON i.id = v.approval_instance_id
		WHERE v.stage_instance_id = $1
		  AND v.actor_user_id = $2
		  AND i.tenant_id = $3::uuid`,
		stageInstanceID, actorUserID, tenantID,
	)
	return scanVerdict(row)
}

func scanVerdict(row rowScanner) (*domain.ReviewVerdict, error) {
	var (
		id, instanceID, stageID, actorUserID, actorTenantID string
		verdict, comment, displayName                       string
		onBehalfOf                                          string
		verdictAt                                           time.Time
	)
	err := row.Scan(&id, &instanceID, &stageID, &actorUserID, &actorTenantID,
		&verdict, &comment, &verdictAt, &displayName, &onBehalfOf)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	// StageKind is not persisted on the verdict row (it's a property of the
	// stage, not the vote); NewVerdict's stage-kind gate was already satisfied
	// when this row was inserted, so re-validating here would need a snapshot
	// value we don't store. Pass StageKindReview — the only kind this table's
	// rows can legally carry (enforced by the service before insert).
	return domain.NewVerdict(domain.VerdictParams{
		ID:                       id,
		ApprovalInstanceID:       instanceID,
		StageInstanceID:          stageID,
		StageKind:                domain.StageKindReview,
		ActorUserID:              actorUserID,
		ActorTenantID:            actorTenantID,
		Verdict:                  domain.Verdict(verdict),
		Comment:                  comment,
		VerdictAt:                verdictAt,
		ActorDisplayNameSnapshot: displayName,
		OnBehalfOfUserID:         onBehalfOf,
	})
}

// LoadStageVerdicts fetches all verdicts for a single stage instance, used to
// evaluate quorum for the `ready` path exactly like LoadStageSignoffs.
func (r *postgresApprovalRepository) LoadStageVerdicts(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.ReviewVerdict, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT v.id, v.approval_instance_id, v.stage_instance_id,
		       v.actor_user_id, v.actor_tenant_id, v.verdict,
		       coalesce(v.comment,''), v.verdict_at, v.actor_display_name_snapshot,
		       coalesce(v.on_behalf_of_user_id,'')
		  FROM approval_review_verdicts v
		  JOIN approval_instances ai
		    ON ai.id = v.approval_instance_id
		 WHERE v.stage_instance_id = $1
		   AND ai.tenant_id = $2::uuid
		   AND v.actor_tenant_id = ai.tenant_id
		 ORDER BY v.verdict_at ASC`,
		stageInstanceID, tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	return scanVerdicts(rows)
}

// LoadInstanceVerdicts fetches ALL verdicts for an approval instance across
// every stage, ordered chronologically (verdict_at ASC), for the by-document
// view's verdict-history projection (F2d.2, ADR 0079). Unlike LoadStageVerdicts
// (single-stage, quorum counting), this spans the whole instance. The actor
// display name is selected DIRECTLY — no coalesce — because migration 0294 made
// actor_display_name_snapshot NOT NULL + non-empty; the name is the immutable
// value cast at the moment of the verdict (no read fallback, no live lookup).
func (r *postgresApprovalRepository) LoadInstanceVerdicts(ctx context.Context, tx db.Tx, tenantID, instanceID string) ([]domain.ReviewVerdict, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT v.id, v.approval_instance_id, v.stage_instance_id,
		       v.actor_user_id, v.actor_tenant_id, v.verdict,
		       coalesce(v.comment,''), v.verdict_at, v.actor_display_name_snapshot,
		       coalesce(v.on_behalf_of_user_id,'')
		  FROM approval_review_verdicts v
		  JOIN approval_instances ai
		    ON ai.id = v.approval_instance_id
		 WHERE v.approval_instance_id = $1
		   AND ai.tenant_id = $2::uuid
		   AND v.actor_tenant_id = ai.tenant_id
		 ORDER BY v.verdict_at ASC`,
		instanceID, tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	return scanVerdicts(rows)
}

// scanVerdicts materializes a verdict row set shared by LoadStageVerdicts and
// LoadInstanceVerdicts (identical column projection and order).
func scanVerdicts(rows *sql.Rows) ([]domain.ReviewVerdict, error) {
	var verdicts []domain.ReviewVerdict
	for rows.Next() {
		var (
			id, instanceID, stageID, actorUserID, actorTenantID string
			verdict, comment, displayName                       string
			onBehalfOf                                          string
			verdictAt                                           time.Time
		)
		if err := rows.Scan(&id, &instanceID, &stageID, &actorUserID, &actorTenantID,
			&verdict, &comment, &verdictAt, &displayName, &onBehalfOf); err != nil {
			return nil, err
		}
		v, err := domain.NewVerdict(domain.VerdictParams{
			ID:                       id,
			ApprovalInstanceID:       instanceID,
			StageInstanceID:          stageID,
			StageKind:                domain.StageKindReview,
			ActorUserID:              actorUserID,
			ActorTenantID:            actorTenantID,
			Verdict:                  domain.Verdict(verdict),
			Comment:                  comment,
			VerdictAt:                verdictAt,
			ActorDisplayNameSnapshot: displayName,
			OnBehalfOfUserID:         onBehalfOf,
		})
		if err != nil {
			return nil, fmt.Errorf("scan verdict %s: %w", id, err)
		}
		verdicts = append(verdicts, *v)
	}
	return verdicts, rows.Err()
}

// ── H-5.1 relocated read helpers ─────────────────────────────────────────────

// LoadPriorSignoffs fetches all signoffs for the instance EXCEPT the active stage,
// used for SoD checking (actor must not have signed in any prior stage).
func (r *postgresApprovalRepository) LoadPriorSignoffs(ctx context.Context, tx db.Tx, tenantID, instanceID, activeStageID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id,
		       actor_user_id, actor_tenant_id, decision,
		       comment, signed_at, signature_method, signature_payload, content_hash,
		       signature_meaning, coalesce(on_behalf_of_user_id,'')
		FROM approval_signoffs
		WHERE approval_instance_id = $1
		  AND stage_instance_id != $2
		  AND actor_tenant_id = $3
		ORDER BY signed_at ASC`,
		instanceID, activeStageID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSignoffsRows(rows)
}

// LoadStageSignoffs fetches all signoffs for a single stage instance.
func (r *postgresApprovalRepository) LoadStageSignoffs(ctx context.Context, tx db.Tx, tenantID, stageInstanceID string) ([]domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id,
		       s.actor_user_id, s.actor_tenant_id, s.decision,
		       s.comment, s.signed_at, s.signature_method, s.signature_payload, s.content_hash,
		       s.signature_meaning, coalesce(s.on_behalf_of_user_id,'')
		  FROM approval_signoffs s
		  JOIN approval_stage_instances asi
		    ON asi.id = s.stage_instance_id
		  JOIN approval_instances ai
		    ON ai.id = asi.approval_instance_id
		   AND ai.id = s.approval_instance_id
		 WHERE s.stage_instance_id = $1
		   AND ai.tenant_id = $2::uuid
		   AND s.actor_tenant_id = ai.tenant_id
		 ORDER BY s.signed_at ASC`,
		stageInstanceID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSignoffsRows(rows)
}

// scanSignoffsRows reads *sql.Rows into a domain.Signoff slice.
func scanSignoffsRows(rows *sql.Rows) ([]domain.Signoff, error) {
	var signoffs []domain.Signoff
	for rows.Next() {
		var (
			id                 string
			approvalInstanceID string
			stageInstanceID    string
			actorUserID        string
			actorTenantID      string
			decision           string
			comment            string
			signedAt           time.Time
			signatureMethod    string
			signaturePayload   []byte
			contentHash        string
			signatureMeaning   string
			onBehalfOf         string
		)
		if err := rows.Scan(
			&id, &approvalInstanceID, &stageInstanceID,
			&actorUserID, &actorTenantID, &decision,
			&comment, &signedAt, &signatureMethod, &signaturePayload, &contentHash,
			&signatureMeaning, &onBehalfOf,
		); err != nil {
			return nil, err
		}
		s, err := domain.NewSignoff(domain.SignoffParams{
			ID:                 id,
			ApprovalInstanceID: approvalInstanceID,
			StageInstanceID:    stageInstanceID,
			ActorUserID:        actorUserID,
			ActorTenantID:      actorTenantID,
			Decision:           domain.Decision(decision),
			Comment:            comment,
			SignedAt:           signedAt,
			SignatureMethod:    signatureMethod,
			SignaturePayload:   json.RawMessage(signaturePayload),
			ContentHash:        contentHash,
			SignatureMeaning:   signatureMeaning,
			OnBehalfOfUserID:   onBehalfOf,
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s: %w", id, err)
		}
		signoffs = append(signoffs, *s)
	}
	return signoffs, rows.Err()
}

// HasUnresolvedComments returns true when the document has one or more
// unresolved comments.
func (r *postgresApprovalRepository) HasUnresolvedComments(ctx context.Context, tx db.Tx, tenantID, documentID string) (bool, error) {
	var unresolvedCount int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL`,
		tenantID, documentID,
	).Scan(&unresolvedCount)
	if err != nil {
		return false, err
	}
	return unresolvedCount > 0, nil
}

// HasUnresolvedInstanceComments is HasUnresolvedComments scoped to comments
// created at or after `since` (F5, spec.md §2.2 "Comment-resolution scope").
func (r *postgresApprovalRepository) HasUnresolvedInstanceComments(ctx context.Context, tx db.Tx, tenantID, documentID string, since time.Time) (bool, error) {
	var unresolvedCount int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM document_comments
		 WHERE tenant_id = $1
		   AND document_id = $2
		   AND resolved_at IS NULL
		   AND created_at >= $3`,
		tenantID, documentID, since,
	).Scan(&unresolvedCount)
	if err != nil {
		return false, err
	}
	return unresolvedCount > 0, nil
}

// PinFrozenHash CAS-writes the frozen content hash the first time an instance
// reaches the freeze boundary (F5). Only succeeds while the column is NULL;
// won=false (0 rows affected) means the instance was already frozen —
// idempotent no-op, never an error.
func (r *postgresApprovalRepository) PinFrozenHash(ctx context.Context, tx db.Tx, tenantID, instanceID, hash string) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		UPDATE approval_instances
		   SET frozen_content_hash = $1
		 WHERE id = $2
		   AND tenant_id = $3
		   AND frozen_content_hash IS NULL`,
		hash, instanceID, tenantID,
	)
	if err != nil {
		return false, MapPgError(err, MapHints{})
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("approval: pin frozen hash rows affected: %w", err)
	}
	return n == 1, nil
}

// LoadFrozenContentHash reads ONLY approval_instances.frozen_content_hash by
// instance id (F6, spec §11 no-fallback principle). This deliberately
// replaces the former LoadActiveDocumentContentHash, which COALESCEd
// documents.content_hash_at_submit against the latest document_revisions
// hash — a silent-fallback that could validate a signature against floating
// head content instead of the frozen pin the signer actually reviewed. There
// is no fallback here: a NULL frozen_content_hash or a missing instance row
// both return ErrNoActiveContentHash and the caller must fail closed.
func (r *postgresApprovalRepository) LoadFrozenContentHash(ctx context.Context, tx db.Tx, tenantID, instanceID string) (string, error) {
	var hash sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT frozen_content_hash
		  FROM approval_instances
		 WHERE id = $1
		   AND tenant_id = $2`,
		instanceID, tenantID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoActiveContentHash
	}
	if err != nil {
		return "", err
	}
	if !hash.Valid {
		return "", ErrNoActiveContentHash
	}
	return hash.String, nil
}

// ResolveEligibleActors returns the user_ids of all users who hold required_role
// in area_code for the given tenant as of now. Returns empty slice (never nil).
//
// Reads iam's PUBLISHED active-membership view metaldocs.v_active_user_areas
// (M3/F3.3, ADR-0039 D3a) — the view encodes active-now (effective_to IS NULL,
// ADR 0037 D1), so this consumer states no temporal predicate. Under ADR-0037
// Model A this is the identical set the prior interval form selected (effective_from
// is never future; effective_to > now() is empty), retiring the Model-B interval
// leak (ADR 0037 D2). H-PRE-1 preserved: still a plain, non-recording in-tx SELECT.
func (r *postgresApprovalRepository) ResolveEligibleActors(ctx context.Context, tx db.Tx, tenantID, areaCode, requiredRole string) ([]string, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT user_id
		   FROM metaldocs.v_active_user_areas
		  WHERE tenant_id = $1::uuid
		    AND area_code = $2
		    AND role      = $3`,
		tenantID, areaCode, requiredRole,
	)
	if err != nil {
		return []string{}, fmt.Errorf("resolveEligibleActors: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return []string{}, fmt.Errorf("resolveEligibleActors: scan: %w", err)
		}
		ids = append(ids, uid)
	}
	if ids == nil {
		ids = []string{}
	}
	if err := rows.Err(); err != nil {
		return []string{}, fmt.Errorf("resolveEligibleActors: rows: %w", err)
	}
	return ids, nil
}

// LoadControlledDocumentID returns documents.controlled_document_id for
// (tenantID, documentID) inside the caller's transaction. ok is false when the
// document row is absent or the controlled-document link is NULL. Plain
// non-recording SELECT (HS-PRE-1); part of the ADR 0073 SubmitDefaultsResolver.
func (r *postgresApprovalRepository) LoadControlledDocumentID(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, bool, error) {
	var cdID sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT controlled_document_id::text
		  FROM documents
		 WHERE id = $1 AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&cdID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !cdID.Valid || cdID.String == "" {
		return "", false, nil
	}
	return cdID.String, true, nil
}

// LoadActiveRouteIDByProfile returns the id of the single active approval route
// for (tenantID, profileCode), newest route version first — the in-tx form of
// the wrapper's GetFinalizePrereqs route query (repository.go:1801-1817).
// Returns ErrNoActiveApprovalRoute when none is active. Column is `active`
// (migration 0146). Plain non-recording SELECT (HS-PRE-1).
func (r *postgresApprovalRepository) LoadActiveRouteIDByProfile(ctx context.Context, tx db.Tx, tenantID, profileCode string) (string, error) {
	var routeID string
	err := tx.QueryRowContext(ctx, `
		SELECT id
		  FROM approval_routes
		 WHERE tenant_id    = $1
		   AND profile_code = $2
		   AND active       = true
		 ORDER BY version DESC
		 LIMIT 1`,
		tenantID, profileCode,
	).Scan(&routeID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoActiveApprovalRoute
	}
	if err != nil {
		return "", err
	}
	return routeID, nil
}

// LoadHeadContentHash returns the most recent autosaved revision's content_hash
// for the document, or "" when none exists — the in-tx form of the wrapper's
// head-hash query (repository.go:1819-1828), COALESCE / no-rows tolerant. Plain
// non-recording SELECT (HS-PRE-1).
func (r *postgresApprovalRepository) LoadHeadContentHash(ctx context.Context, tx db.Tx, tenantID, documentID string) (string, error) {
	var contentHash string
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(dr.content_hash, '')
		  FROM document_revisions dr
		  JOIN documents d ON d.id = dr.document_id
		 WHERE dr.document_id = $1
		   AND d.tenant_id    = $2
		 ORDER BY dr.created_at DESC, dr.id DESC
		 LIMIT 1`,
		documentID, tenantID,
	).Scan(&contentHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return contentHash, nil
}

// LoadRoute fetches an approval route and its stages from the database within
// the caller's transaction.
func (r *postgresApprovalRepository) LoadRoute(ctx context.Context, tx db.Tx, tenantID, routeID string) (domain.Route, error) {
	var route domain.Route
	err := tx.QueryRowContext(ctx, `
		SELECT id, tenant_id, profile_code, version
		FROM approval_routes
		WHERE id = $1 AND tenant_id = $2 AND active = TRUE`,
		routeID, tenantID,
	).Scan(&route.ID, &route.TenantID, &route.ProfileCode, &route.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Route{}, fmt.Errorf("route %s not found for tenant %s", routeID, tenantID)
		}
		return domain.Route{}, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT ars.stage_order, ars.name, ars.required_role, ars.required_capability,
		       ars.area_code, ars.quorum, ars.quorum_m, ars.on_eligibility_drift,
		       ars.stage_kind, ars.due_in_days
		  FROM approval_route_stages ars
		  JOIN approval_routes ar
		    ON ar.id = ars.route_id
		   AND ar.tenant_id = $2
		 WHERE ars.route_id = $1
		 ORDER BY ars.stage_order ASC`,
		routeID, tenantID,
	)
	if err != nil {
		return domain.Route{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var stage domain.Stage
		var quorumM sql.NullInt32
		var dueInDays sql.NullInt32
		var kindStr string
		if err := rows.Scan(
			&stage.Order, &stage.Name, &stage.RequiredRole, &stage.RequiredCapability,
			&stage.AreaCode, &stage.Quorum, &quorumM, &stage.OnEligibilityDrift,
			&kindStr, &dueInDays,
		); err != nil {
			return domain.Route{}, err
		}
		if quorumM.Valid {
			v := int(quorumM.Int32)
			stage.QuorumM = &v
		}
		stage.Kind = domain.StageKind(kindStr)
		if dueInDays.Valid {
			v := int(dueInDays.Int32)
			stage.DueInDays = &v
		}
		route.Stages = append(route.Stages, stage)
	}
	if err := rows.Err(); err != nil {
		return domain.Route{}, err
	}

	return route, nil
}

// ── F9 delegation (ADR 0077) ─────────────────────────────────────────────────

// InsertDelegation persists a new approval_delegations row. Plain insert — no
// idempotency-replay concept; overlapping active grants for the same
// delegator are explicitly allowed (union semantics, spec.md Interview #7).
func (r *postgresApprovalRepository) InsertDelegation(ctx context.Context, tx db.Tx, d domain.Delegation) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO approval_delegations
		  (id, tenant_id, delegator_id, delegate_id, starts_at, ends_at, reason, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.ID(), d.TenantID(), d.DelegatorID(), d.DelegateID(),
		d.StartsAt(), d.EndsAt(), d.Reason(), d.CreatedBy(), d.CreatedAt(),
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	return nil
}

// DeleteDelegation hard-deletes an approval_delegations row IFF tenant_id
// matches and the caller owns it (is the delegator) or holds oversee. The
// ownership check is expressed in the WHERE clause itself (same OCC-via-WHERE
// idiom this package already uses, e.g. CancelInstance's revision-version
// guard) so the delete is atomic — no separate load-then-check race window.
func (r *postgresApprovalRepository) DeleteDelegation(ctx context.Context, tx db.Tx, tenantID, delegationID, callerID string, callerIsOversee bool) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		DELETE FROM approval_delegations
		 WHERE id = $1
		   AND tenant_id = $2
		   AND (delegator_id = $3 OR $4::boolean)`,
		delegationID, tenantID, callerID, callerIsOversee,
	)
	if err != nil {
		return false, MapPgError(err, MapHints{})
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// LoadActiveDelegationsFor returns every approval_delegations row where
// delegate_id matches and the window covers asOf. Always called in-tx at
// actual use-time (F9/ADR 0077 §5) — no caching, no off-tx read.
func (r *postgresApprovalRepository) LoadActiveDelegationsFor(ctx context.Context, tx db.Tx, tenantID, delegateID string, asOf time.Time) ([]domain.Delegation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, tenant_id, delegator_id, delegate_id, starts_at, ends_at, reason, created_by, created_at
		  FROM approval_delegations
		 WHERE tenant_id = $1
		   AND delegate_id = $2
		   AND starts_at <= $3
		   AND ends_at > $3`,
		tenantID, delegateID, asOf,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	var delegations []domain.Delegation
	for rows.Next() {
		var (
			id, tid, delegatorID, delegateIDCol, reason, createdBy string
			startsAt, endsAt, createdAt                            time.Time
		)
		if err := rows.Scan(&id, &tid, &delegatorID, &delegateIDCol, &startsAt, &endsAt, &reason, &createdBy, &createdAt); err != nil {
			return nil, err
		}
		d, err := domain.NewDelegation(domain.DelegationParams{
			ID: id, TenantID: tid, DelegatorID: delegatorID, DelegateID: delegateIDCol,
			StartsAt: startsAt, EndsAt: endsAt, Reason: reason, CreatedBy: createdBy, CreatedAt: createdAt,
		})
		if err != nil {
			return nil, fmt.Errorf("scan delegation %s: %w", id, err)
		}
		delegations = append(delegations, *d)
	}
	return delegations, rows.Err()
}
