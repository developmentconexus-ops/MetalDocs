package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
)

type postgresApprovalRepository struct {
	db *sql.DB
}

var ErrInvalidScheduledSupersedeTarget = errors.New("approval: invalid scheduled supersede target")

// NewPostgresApprovalRepository constructs a production Postgres-backed ApprovalRepository.
func NewPostgresApprovalRepository(db *sql.DB) ApprovalRepository {
	return &postgresApprovalRepository{db: db}
}

// execer is the common interface satisfied by both *sql.Tx and *sql.DB.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// InsertInstance writes a new approval_instances row within the caller's transaction.
func (r *postgresApprovalRepository) InsertInstance(ctx context.Context, tx *sql.Tx, inst domain.Instance) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO approval_instances
		  (id, tenant_id, document_id, route_id, route_version_snapshot,
		   status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
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
	)
	if err != nil {
		return MapPgError(err, MapHints{})
	}
	return nil
}

// InsertStageInstances bulk-inserts all stage instances for an approval in one round-trip.
func (r *postgresApprovalRepository) InsertStageInstances(ctx context.Context, tx *sql.Tx, stages []domain.StageInstance) error {
	if len(stages) == 0 {
		return nil
	}

	// Build multi-row VALUES clause.
	const colCount = 13
	placeholders := make([]string, 0, len(stages))
	args := make([]any, 0, len(stages)*colCount)

	for i, s := range stages {
		base := i * colCount
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5,
			base+6, base+7, base+8, base+9, base+10,
			base+11, base+12, base+13,
		))

		eligibleJSON, err := json.Marshal(s.EligibleActorIDs)
		if err != nil {
			return fmt.Errorf("marshal eligible_actor_ids for stage %s: %w", s.ID, err)
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
		)
	}

	query := `INSERT INTO approval_stage_instances
		(id, approval_instance_id, stage_order, name_snapshot,
		 required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		 quorum_snapshot, quorum_m_snapshot, on_eligibility_drift_snapshot,
		 eligible_actor_ids, effective_denominator, status)
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
func (r *postgresApprovalRepository) InsertSignoff(ctx context.Context, tx *sql.Tx, s domain.Signoff) (SignoffInsertResult, error) {
	payload := s.SignaturePayload()
	if payload == nil {
		payload = json.RawMessage("{}")
	}

	var actorDisplayNameSnapshot sql.NullString
	if v := s.ActorDisplayNameSnapshot(); v != "" {
		actorDisplayNameSnapshot = sql.NullString{String: v, Valid: true}
	}

	var returnedID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO approval_signoffs
		  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
		   decision, comment, signed_at, signature_method, signature_payload, content_hash,
		   actor_display_name_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
func (r *postgresApprovalRepository) LoadSignoffByActor(ctx context.Context, tx *sql.Tx, tenantID, instanceID, actorUserID string) (*domain.Signoff, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id, s.actor_user_id,
		       s.actor_tenant_id, s.decision, coalesce(s.comment,''), s.signed_at,
		       s.signature_method, s.signature_payload, s.content_hash
		FROM approval_signoffs s
		JOIN approval_instances i ON i.id = s.approval_instance_id
		WHERE s.approval_instance_id = $1
		  AND s.actor_user_id = $2
		  AND i.tenant_id = $3`,
		instanceID, actorUserID, tenantID,
	)
	return scanSignoff(row)
}

func (r *postgresApprovalRepository) loadSignoffByStageActor(ctx context.Context, tx *sql.Tx, tenantID, stageInstanceID, actorUserID string) (*domain.Signoff, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT s.id, s.approval_instance_id, s.stage_instance_id, s.actor_user_id,
		       s.actor_tenant_id, s.decision, coalesce(s.comment,''), s.signed_at,
		       s.signature_method, s.signature_payload, s.content_hash
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
		signedAt                                            time.Time
		sigPayload                                          []byte
	)
	err := row.Scan(&id, &instanceID, &stageID, &actorUserID, &actorTenantID,
		&decision, &comment, &signedAt, &signatureMethod, &sigPayload, &contentHash)
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
		SignedAt:           signedAt,
		SignatureMethod:    signatureMethod,
		SignaturePayload:   json.RawMessage(sigPayload),
		ContentHash:        contentHash,
	})
}

// LoadInstance loads an approval instance and its stage instances by ID.
// Returns ErrNoActiveInstance if not found or tenant mismatch.
func (r *postgresApprovalRepository) LoadInstance(ctx context.Context, tx *sql.Tx, tenantID, id string) (*domain.Instance, error) {
	var inst domain.Instance
	var completedAt sql.NullTime

	err := tx.QueryRowContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key
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

	stages, err := r.loadStageInstances(ctx, tx, tenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	inst.Stages = stages
	return &inst, nil
}

// LoadActiveInstanceByDocument loads the single in_progress instance for a document.
// Returns ErrNoActiveInstance when none exists or tenant doesn't match.
func (r *postgresApprovalRepository) LoadActiveInstanceByDocument(ctx context.Context, tx *sql.Tx, tenantID, docID string) (*domain.Instance, error) {
	var inst domain.Instance
	var completedAt sql.NullTime

	err := tx.QueryRowContext(ctx, `
		SELECT ai.id, ai.tenant_id, ai.document_id, ai.route_id, ai.route_version_snapshot,
		       d.revision_version,
		       ai.status, ai.submitted_by, ai.submitted_at, ai.completed_at,
		       ai.content_hash_at_submit, ai.idempotency_key
		FROM approval_instances ai
		JOIN documents d
		  ON d.id = ai.document_id
		 AND d.tenant_id = ai.tenant_id
		WHERE ai.document_id = $1
		  AND ai.tenant_id = $2
		  AND ai.status IN ('in_progress', 'approved')
			ORDER BY ai.submitted_at DESC
			LIMIT 1`,
		docID, tenantID,
	).Scan(
		&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.RouteID, &inst.RouteVersionSnapshot,
		&inst.RevisionVersion,
		&inst.Status, &inst.SubmittedBy, &inst.SubmittedAt, &completedAt,
		&inst.ContentHashAtSubmit, &inst.IdempotencyKey,
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

	stages, err := r.loadStageInstances(ctx, tx, tenantID, inst.ID)
	if err != nil {
		return nil, err
	}
	inst.Stages = stages
	return &inst, nil
}

func (r *postgresApprovalRepository) ValidateScheduledSupersedeTarget(ctx context.Context, tx *sql.Tx, tenantID, documentID, supersededDocumentID string) error {
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

func (r *postgresApprovalRepository) LoadCurrentPublishedHeadForDocument(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) {
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

func (r *postgresApprovalRepository) LoadCurrentPublishedHead(ctx context.Context, tx *sql.Tx, tenantID, controlledDocumentID string) (string, error) {
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

func (r *postgresApprovalRepository) GetDocumentRevisionVersion(ctx context.Context, tx *sql.Tx, documentID, tenantID string) (int, error) {
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

// ListRoutes loads tenant-scoped route configuration and stages in one joined query.
func (r *postgresApprovalRepository) ListRoutes(ctx context.Context, tenantID string) ([]Route, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.tenant_id::text, r.profile_code, r.active, r.created_at, r.created_at AS updated_at,
		       s.stage_order, s.name, s.required_role, s.required_capability, s.area_code, s.quorum, s.on_eligibility_drift
		  FROM approval_routes r
		  JOIN approval_route_stages s
		    ON s.route_id = r.id
		 WHERE r.tenant_id = $1::uuid
		 ORDER BY r.created_at DESC, s.stage_order ASC`,
		tenantID,
	)
	if err != nil {
		return nil, MapPgError(err, MapHints{})
	}
	defer rows.Close()

	routeMap := make(map[string]*Route)
	var routeOrder []string
	for rows.Next() {
		var (
			routeID, routeName, routeTenantID, profileCode string
			active                                         bool
			createdAt, updatedAt                           time.Time
			stage                                          RouteStage
			stageName, stageRole, stageCapability          sql.NullString
			stageArea, stageQuorum, stageDrift             sql.NullString
		)
		if err := rows.Scan(
			&routeID, &routeName, &routeTenantID, &profileCode, &active, &createdAt, &updatedAt,
			&stage.Order, &stageName, &stageRole, &stageCapability, &stageArea, &stageQuorum, &stageDrift,
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
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				Stages:      []RouteStage{},
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
		if stageDrift.Valid {
			stage.DriftPolicy = stageDrift.String
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

func (r *postgresApprovalRepository) MarkSuperseded(ctx context.Context, tx *sql.Tx, tenantID, documentID string) error {
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
func (r *postgresApprovalRepository) loadStageInstances(ctx context.Context, tx *sql.Tx, tenantID, instanceID string) ([]domain.StageInstance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_order, name_snapshot,
		       required_role_snapshot, required_capability_snapshot, area_code_snapshot,
		       quorum_snapshot, quorum_m_snapshot,
		       on_eligibility_drift_snapshot,
		       eligible_actor_ids, effective_denominator,
		       status, opened_at, completed_at, skip_reason
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
		var openedAt, completedAt sql.NullTime
		var eligibleJSON []byte
		var skipReason sql.NullString

		err := rows.Scan(
			&s.ID, &s.ApprovalInstanceID, &s.StageOrder, &s.NameSnapshot,
			&s.RequiredRoleSnapshot, &s.RequiredCapabilitySnapshot, &s.AreaCodeSnapshot,
			&s.QuorumSnapshot, &quorumMSnapshot,
			&s.OnEligibilityDriftSnapshot,
			&eligibleJSON, &effectiveDenominator,
			&s.Status, &openedAt, &completedAt, &skipReason,
		)
		if err != nil {
			return nil, err
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

		if len(eligibleJSON) > 0 {
			if err := json.Unmarshal(eligibleJSON, &s.EligibleActorIDs); err != nil {
				return nil, fmt.Errorf("unmarshal eligible_actor_ids for stage %s: %w", s.ID, err)
			}
		}

		stages = append(stages, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load signoffs for the instance and attach to stages.
	byStage, err := r.loadSignoffsForInstance(ctx, tx, tenantID, instanceID)
	if err != nil {
		return nil, err
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
func (r *postgresApprovalRepository) loadSignoffsForInstance(ctx context.Context, tx *sql.Tx, tenantID, instanceID string) (map[string][]*domain.Signoff, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, approval_instance_id, stage_instance_id, actor_user_id,
		       actor_tenant_id, decision, coalesce(comment,''), signed_at,
		       signature_method, signature_payload, content_hash,
		       coalesce(actor_display_name_snapshot,'')
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
			signedAt                                        time.Time
			sigPayload                                      []byte
		)
		if err := rows.Scan(&id, &instID, &stageID, &actorUserID, &actorTenantID,
			&decision, &comment, &signedAt, &signatureMethod, &sigPayload,
			&contentHash, &displayName); err != nil {
			return nil, err
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
		})
		if err != nil {
			return nil, fmt.Errorf("scan signoff %s: %w", id, err)
		}
		out[stageID] = append(out[stageID], sig)
	}
	return out, rows.Err()
}

// UpdateStageStatus applies an OCC (optimistic concurrency control) UPDATE.
// Checks RowsAffected == 0 — which means expectedOldStatus was not the current value — and returns ErrStageNotActive.
func (r *postgresApprovalRepository) UpdateStageStatus(ctx context.Context, tx *sql.Tx, tenantID, stageID string, newStatus, expectedOldStatus domain.StageStatus) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE approval_stage_instances asi
		SET status = $1,
		    opened_at    = CASE WHEN $1 = 'active'    THEN now() ELSE asi.opened_at    END,
		    completed_at = CASE WHEN $1 IN ('completed','skipped','rejected_here') THEN now() ELSE asi.completed_at END
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
func (r *postgresApprovalRepository) UpdateInstanceStatus(ctx context.Context, tx *sql.Tx, tenantID, instID string, newStatus domain.InstanceStatus, expectedStatus domain.InstanceStatus, completedAt *time.Time) error {
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
