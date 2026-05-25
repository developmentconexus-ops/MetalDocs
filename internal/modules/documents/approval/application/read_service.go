package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// InboxView is the read-model projection for the inbox UI.
type InboxView struct {
	InstanceID           string
	DocumentID           string
	ControlledDocumentID string
	DocumentTitle        string
	AreaCode             string
	SubmittedBy          string
	SubmittedAt          time.Time
	StageLabel           string
	QuorumProgress       string // e.g. "1/2"
}

// ReadService exposes read-only operations for approval HTTP handlers.
type ReadService struct {
	repo repository.ApprovalRepository
}

func newReadService(repo repository.ApprovalRepository) *ReadService {
	return &ReadService{repo: repo}
}

// LoadInstance loads a single approval instance by ID for the given tenant.
func (s *ReadService) LoadInstance(ctx context.Context, db *sql.DB, tenantID, actorID, instanceID string) (*domain.Instance, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("read load instance: begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
		return nil, fmt.Errorf("read load instance: %w", err)
	}

	areaCode, err := loadInstanceAreaCode(ctx, tx, tenantID, instanceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoActiveInstance
		}
		return nil, fmt.Errorf("read load instance: load area: %w", err)
	}

	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), areaCode); err != nil {
		return nil, err
	}

	inst, err := s.repo.LoadInstance(ctx, tx, tenantID, instanceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoActiveInstance
		}
		return nil, err
	}
	if inst == nil {
		return nil, repository.ErrNoActiveInstance
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("read load instance: commit tx: %w", err)
	}
	return inst, nil
}

// LoadActiveInstanceByDocument finds the current active approval instance for a document.
func (s *ReadService) LoadActiveInstanceByDocument(ctx context.Context, db *sql.DB, tenantID, documentID string) (*domain.Instance, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("read load instance by document: begin tx: %w", err)
	}
	defer tx.Rollback()

	inst, err := s.repo.LoadActiveInstanceByDocument(ctx, tx, tenantID, documentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoActiveInstance
		}
		return nil, err
	}
	if inst == nil {
		return nil, repository.ErrNoActiveInstance
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("read load instance by document: commit tx: %w", err)
	}
	return inst, nil
}

// ListPendingForActor lists inbox items pending actor action.
func (s *ReadService) ListPendingForActor(ctx context.Context, db *sql.DB, tenantID, actorID string, areaCode string, limit, offset int) ([]domain.Instance, error) {
	if limit <= 0 {
		limit = 25
	}

	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return nil, fmt.Errorf("list pending: marshal actor: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list pending: begin tx: %w", err)
	}
	defer tx.Rollback()

	const q = `
		SELECT DISTINCT ai.id
		FROM approval_instances ai
		JOIN approval_stage_instances asi ON asi.approval_instance_id = ai.id
		WHERE ai.tenant_id = $1
		  AND ai.status = 'in_progress'
		  AND asi.status = 'active'
		  AND asi.eligible_actor_ids @> $2::jsonb
		  AND ($3 = '' OR asi.area_code_snapshot = $3)
		ORDER BY ai.id
		LIMIT $4 OFFSET $5`

	rows, err := tx.QueryContext(ctx, q, tenantID, string(actorJSON), areaCode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list pending: query: %w", err)
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("list pending: scan id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list pending: rows: %w", err)
	}

	out := make([]domain.Instance, 0, len(ids))
	for _, id := range ids {
		inst, err := s.repo.LoadInstance(ctx, tx, tenantID, id)
		if err != nil {
			return nil, fmt.Errorf("list pending: load instance %s: %w", id, err)
		}
		out = append(out, *inst)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("list pending: commit: %w", err)
	}
	return out, nil
}

// ListInboxItems returns inbox view rows for the given tenant + actor.
// Single JOIN against documents and a signoff-count subquery so the UI can
// render document titles and quorum progress without N+1 lookups.
func (s *ReadService) ListInboxItems(ctx context.Context, db *sql.DB, tenantID, actorID, areaCode string, limit, offset int) ([]InboxView, error) {
	if limit <= 0 {
		limit = 25
	}

	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return nil, fmt.Errorf("list inbox: marshal actor: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list inbox: begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
		return nil, fmt.Errorf("list inbox: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			ai.id,
			ai.document_id,
			COALESCE(d.controlled_document_id::text, '') AS controlled_document_id,
			COALESCE(d.name, '') AS doc_title,
			COALESCE(asi.area_code_snapshot, '') AS area_code,
			ai.submitted_by,
			ai.submitted_at,
			COALESCE(asi.name_snapshot, '') AS stage_label,
			COALESCE(
				CASE asi.quorum_snapshot
					WHEN 'all_of'  THEN COALESCE(jsonb_array_length(asi.eligible_actor_ids), 0)
					WHEN 'm_of_n'  THEN COALESCE(asi.quorum_m_snapshot, 1)
					ELSE 1
				END, 1) AS required,
			COALESCE((
				SELECT count(*)
				FROM approval_signoffs s
				WHERE s.approval_instance_id = ai.id
				  AND s.stage_instance_id = asi.id
				  AND s.actor_tenant_id = ai.tenant_id
				  AND s.decision = 'approve'
			), 0) AS signed
		FROM approval_instances ai
		JOIN approval_stage_instances asi
		  ON asi.approval_instance_id = ai.id
		 AND asi.status = 'active'
		LEFT JOIN documents d
		  ON d.id = ai.document_id AND d.tenant_id = ai.tenant_id
		WHERE ai.tenant_id = $1::uuid
		  AND ai.status = 'in_progress'
		  AND asi.eligible_actor_ids @> $2::jsonb
		  AND ($3 = '' OR asi.area_code_snapshot = $3)
		ORDER BY ai.submitted_at DESC
		LIMIT $4 OFFSET $5`,
		tenantID, actorJSON, areaCode, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list inbox: query: %w", err)
	}

	var items []InboxView
	for rows.Next() {
		var v InboxView
		var signed, required int
		if err := rows.Scan(
			&v.InstanceID, &v.DocumentID, &v.ControlledDocumentID, &v.DocumentTitle,
			&v.AreaCode, &v.SubmittedBy, &v.SubmittedAt,
			&v.StageLabel, &required, &signed,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("list inbox: scan: %w", err)
		}
		v.QuorumProgress = fmt.Sprintf("%d/%d", signed, required)
		items = append(items, v)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list inbox: rows: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("list inbox: commit: %w", err)
	}
	return items, nil
}

// CountPendingForActor returns the total number of pending approval instances
// for the given tenant + actor (no LIMIT/OFFSET) so the UI can paginate.
func (s *ReadService) CountPendingForActor(ctx context.Context, db *sql.DB, tenantID, actorID, areaCode string) (int, error) {
	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return 0, fmt.Errorf("count pending: marshal actor: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("count pending: begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := setAuthzGUC(ctx, tx, tenantID, actorID); err != nil {
		return 0, fmt.Errorf("count pending: %w", err)
	}

	var total int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT ai.id)
		FROM approval_instances ai
		JOIN approval_stage_instances asi
		  ON asi.approval_instance_id = ai.id
		 AND asi.status = 'active'
		WHERE ai.tenant_id = $1::uuid
		  AND ai.status = 'in_progress'
		  AND asi.eligible_actor_ids @> $2::jsonb
		  AND ($3 = '' OR asi.area_code_snapshot = $3)`,
		tenantID, actorJSON, areaCode,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count pending: query: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("count pending: commit: %w", err)
	}
	return total, nil
}

func loadInstanceAreaCode(ctx context.Context, tx *sql.Tx, tenantID, instanceID string) (string, error) {
	var areaCode string
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(asi.area_code_snapshot, d.process_area_code_snapshot, cd.process_area_code, 'tenant')
		  FROM approval_instances ai
		  JOIN documents d
		    ON d.id = ai.document_id
		   AND d.tenant_id = ai.tenant_id
		  LEFT JOIN controlled_documents cd
		    ON cd.id = d.controlled_document_id
		  LEFT JOIN approval_stage_instances asi
		    ON asi.approval_instance_id = ai.id
		   AND asi.status = 'active'
		 WHERE ai.id = $1
		   AND ai.tenant_id = $2
		 LIMIT 1`,
		instanceID, tenantID,
	).Scan(&areaCode)
	if err != nil {
		return "", err
	}
	if areaCode == "" {
		return "tenant", nil
	}
	return areaCode, nil
}
