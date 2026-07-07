package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/pagination"
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
	repo   infrastructure.ApprovalRepository
	cdRead controlleddocumentsdomain.CDFieldReader
}

func newReadService(repo infrastructure.ApprovalRepository, cdRead controlleddocumentsdomain.CDFieldReader) *ReadService {
	if cdRead == nil {
		cdRead = controlleddocumentsdomain.NoopCDFieldReader{}
	}
	return &ReadService{repo: repo, cdRead: cdRead}
}

// Read methods intentionally use default (read-write) transactions because
// repository stage loads may use SELECT ... FOR UPDATE on approval rows.

// LoadInstance loads a single approval instance by ID for the given tenant.
func (s *ReadService) LoadInstance(ctx context.Context, runner db.TxRunner, tenantID, instanceID string) (*domain.Instance, error) {
	var inst *domain.Instance
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		_, found, err := loadInstanceAreaCode(ctx, tx, s.cdRead, tenantID, instanceID)
		if err != nil {
			return fmt.Errorf("read load instance: load area: %w", err)
		}
		if !found {
			return infrastructure.ErrNoActiveInstance
		}

		ctx := authz.WithCapCache(ctx)
		// CapDocumentView is tenant-grade (iam/domain/capability_scope.go:51); pass the
		// "tenant" sentinel so the area filter is intentionally OFF — mirrors the
		// canonical documents/application/view_service.go:71. CapApprovalOversee
		// (M2b F3) is an explicit alternative — oversight is its own capability,
		// never a role check (ADR 0022).
		if viewErr := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); viewErr != nil {
			if overseeErr := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); overseeErr != nil {
				return viewErr
			}
		}

		loaded, err := s.repo.LoadInstance(ctx, tx, tenantID, instanceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return err
		}
		if loaded == nil {
			return infrastructure.ErrNoActiveInstance
		}

		inst = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// LoadActiveInstanceByDocument finds the current active approval instance for a document.
func (s *ReadService) LoadActiveInstanceByDocument(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error) {
	var inst *domain.Instance
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		// CapDocumentView is tenant-grade (iam/domain/capability_scope.go:51); pass the
		// "tenant" sentinel so the area filter is intentionally OFF — mirrors the
		// canonical documents/application/view_service.go:71. A missing instance is
		// surfaced as ErrNoActiveInstance by the repo lookup below. CapApprovalOversee
		// (M2b F3) is an explicit alternative — oversight is its own capability, never
		// a role check (ADR 0022).
		if viewErr := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); viewErr != nil {
			if overseeErr := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); overseeErr != nil {
				return viewErr
			}
		}

		loaded, err := s.repo.LoadActiveInstanceByDocument(ctx, tx, tenantID, documentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return err
		}
		if loaded == nil {
			return infrastructure.ErrNoActiveInstance
		}

		inst = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// LoadActiveInstanceByDocumentForMutation finds the current active approval
// instance for a document without enforcing read-capability checks.
// Mutation services enforce their own capability gates (e.g. signoff/cancel).
func (s *ReadService) LoadActiveInstanceByDocumentForMutation(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error) {
	var inst *domain.Instance
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		loaded, err := s.repo.LoadActiveInstanceByDocument(ctx, tx, tenantID, documentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return infrastructure.ErrNoActiveInstance
			}
			return err
		}
		if loaded == nil {
			return infrastructure.ErrNoActiveInstance
		}

		inst = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// ListPendingForActor lists inbox items pending actor action.
func (s *ReadService) ListPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID string, areaCode string, limit, offset int) ([]domain.Instance, error) {
	if limit <= 0 {
		limit = 25
	}

	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return nil, fmt.Errorf("list pending: marshal actor: %w", err)
	}

	var out []domain.Instance
	err = runner.Do(ctx, func(tx *sql.Tx) error {
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
			return fmt.Errorf("list pending: query: %w", err)
		}

		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return fmt.Errorf("list pending: scan id: %w", err)
			}
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list pending: rows: %w", err)
		}

		// Batch-load all instances in a single query set (REQ-DATA-2 / F-10).
		loaded, err := s.repo.LoadInstancesByIDs(ctx, tx, tenantID, ids)
		if err != nil {
			return fmt.Errorf("list pending: batch load instances: %w", err)
		}

		out = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListInboxItems returns inbox view rows for the given tenant + actor.
// Single JOIN against documents and a signoff-count subquery so the UI can
// render document titles and quorum progress without N+1 lookups.
//
// Deprecated: callers that also need the total pending count should use
// ListInboxItemsWithTotal instead, which computes both in one query/tx and
// therefore cannot observe the snapshot drift a signoff committed between two
// independent queries can cause (T-005). Retained for existing callers/tests
// that only need the page of items.
func (s *ReadService) ListInboxItems(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]InboxView, error) {
	items, _, err := s.listInboxItems(ctx, runner, tenantID, actorID, areaCode, limit, offset, false)
	return items, err
}

// ListInboxItemsWithTotal returns the inbox page and the total pending count
// for the given tenant + actor in a single query executed inside one
// transaction (T-005 fix). Using COUNT(*) OVER() instead of a second
// LIMIT/OFFSET-free COUNT query eliminates the snapshot-drift window where a
// signoff committed between two independent queries could make
// total < len(items) or vice versa. limit is clamped via the shared platform
// pagination bounds (internal/platform/pagination.ClampLimit); offset < 0 is
// treated as 0.
func (s *ReadService) ListInboxItemsWithTotal(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int) ([]InboxView, int, error) {
	return s.listInboxItems(ctx, runner, tenantID, actorID, areaCode, limit, offset, true)
}

func (s *ReadService) listInboxItems(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, limit, offset int, withTotal bool) ([]InboxView, int, error) {
	limit = pagination.ClampLimit(limit)
	if offset < 0 {
		offset = 0
	}

	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return nil, 0, fmt.Errorf("list inbox: marshal actor: %w", err)
	}

	var items []InboxView
	var total int
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		// COUNT(*) OVER() piggybacks the total-matching-rows count onto the same
		// result set as the page, so the page and the total are always read from
		// the same MVCC snapshot inside this transaction — no second round-trip
		// that a concurrent signoff could land between (T-005).
		totalSelect := "0 AS total_count"
		if withTotal {
			totalSelect = "COUNT(*) OVER() AS total_count"
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
				), 0) AS signed,
				`+totalSelect+`
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
			ORDER BY ai.submitted_at DESC, ai.id DESC
			LIMIT $4 OFFSET $5`,
			tenantID, actorJSON, areaCode, limit, offset,
		)
		if err != nil {
			return fmt.Errorf("list inbox: query: %w", err)
		}

		for rows.Next() {
			var v InboxView
			var signed, required, rowTotal int
			if err := rows.Scan(
				&v.InstanceID, &v.DocumentID, &v.ControlledDocumentID, &v.DocumentTitle,
				&v.AreaCode, &v.SubmittedBy, &v.SubmittedAt,
				&v.StageLabel, &required, &signed, &rowTotal,
			); err != nil {
				rows.Close()
				return fmt.Errorf("list inbox: scan: %w", err)
			}
			v.QuorumProgress = fmt.Sprintf("%d/%d", signed, required)
			items = append(items, v)
			total = rowTotal
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list inbox: rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	if withTotal && len(items) == 0 {
		// COUNT(*) OVER() only appears on returned rows; an empty page (e.g. an
		// offset past the end, or genuinely zero pending items) must fall back to
		// a real count so the caller doesn't report total=0 for an out-of-range
		// page. This is a second query (like the pre-fix behavior) but only for
		// the empty-page edge case — the common (non-empty) case reads both the
		// page and the total from one snapshot, closing the T-005 drift window.
		count, err := s.CountPendingForActor(ctx, runner, tenantID, actorID, areaCode)
		if err != nil {
			return nil, 0, err
		}
		total = count
	}
	return items, total, nil
}

// CountPendingForActor returns the total number of pending approval instances
// for the given tenant + actor (no LIMIT/OFFSET) so the UI can paginate.
func (s *ReadService) CountPendingForActor(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string) (int, error) {
	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return 0, fmt.Errorf("count pending: marshal actor: %w", err)
	}

	var total int
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `
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
			return fmt.Errorf("count pending: query: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

// loadInstanceAreaCode resolves an approval instance's area, preferring the active
// stage snapshot, then the document snapshot, then the controlled-document area.
// found reports whether the instance row exists; areaCode is "" when the row
// exists with no area anywhere in the chain. Like the document-keyed
// docapp.LoadDocumentAreaCode (ADR 0022 Phase 11 F7) it bakes in NO empty-area
// default — the (tenant-grade) callers COALESCE "" -> "tenant" at the call site so
// the area-filter-OFF decision is explicit, not hidden in the resolver.
func loadInstanceAreaCode(ctx context.Context, tx *sql.Tx, cdRead controlleddocumentsdomain.CDFieldReader, tenantID, instanceID string) (areaCode string, found bool, err error) {
	if cdRead == nil {
		cdRead = controlleddocumentsdomain.NoopCDFieldReader{}
	}
	// Own-module read only: the approval/documents tables. The controlled_documents
	// JOIN was deleted (M2/F2.1 B6, ADR-0039 D3(b)); its area is resolved through
	// the CDFieldReader port below, preserving the original COALESCE precedence
	// (active-stage snapshot, then document snapshot, then CD area, then ""). The
	// snapshots are read as NULLable so an empty-string snapshot still WINS the
	// COALESCE exactly as the SQL did — only a NULL snapshot falls through.
	var (
		stageSnap sql.NullString
		docSnap   sql.NullString
		cdID      sql.NullString
	)
	err = tx.QueryRowContext(ctx, `
		SELECT asi.area_code_snapshot,
		       d.process_area_code_snapshot,
		       d.controlled_document_id
		  FROM approval_instances ai
		  JOIN documents d
		    ON d.id = ai.document_id
		   AND d.tenant_id = ai.tenant_id
		  LEFT JOIN approval_stage_instances asi
		    ON asi.approval_instance_id = ai.id
		   AND asi.status = 'active'
		 WHERE ai.id = $1
		   AND ai.tenant_id = $2
		 LIMIT 1`,
		instanceID, tenantID,
	).Scan(&stageSnap, &docSnap, &cdID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load instance area code: %w", err)
	}
	switch {
	case stageSnap.Valid:
		return stageSnap.String, true, nil
	case docSnap.Valid:
		return docSnap.String, true, nil
	case cdID.Valid:
		// tx-aware foreign read-port (in-tx, non-recording SELECT — HS-PRE-1 safe).
		// The original LEFT JOIN keyed only on cd.id (a globally-unique PK); the
		// port additionally scopes by tenant_id, which is parity-preserving because
		// documents.controlled_document_id always references a same-tenant CD.
		area, _, perr := cdRead.ProcessAreaCode(ctx, tx, tenantID, cdID.String)
		if perr != nil {
			return "", false, fmt.Errorf("load instance area code: cd area port: %w", perr)
		}
		return area, true, nil
	default:
		return "", true, nil
	}
}
