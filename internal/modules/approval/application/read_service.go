package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
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
	// StageKind (F8, spec.md §4/W4) is the pending/active stage's kind
	// (review or approval), sourced from the stage's own snapshot column —
	// never re-derived from route config, so an in-flight instance's worklist
	// row stays pinned to what the stage started with (mirrors every other
	// *_snapshot field, e.g. NameSnapshot/AreaCodeSnapshot).
	StageKind domain.StageKind
	// DueAt (F8, spec.md §4/W4) is the stage's SLA due date, NULL when no SLA
	// is configured for the stage (no-fallback principle, spec §11) — never
	// substituted with a computed/default value.
	DueAt *time.Time
}

// InboxFilter (F8, spec.md §4/W4, §6.3 P2/P3/P8) narrows the worklist beyond
// the base tenant+actor+area scope already on ListInboxItemsWithTotal.
// Zero value = no additional filtering (every field optional).
type InboxFilter struct {
	// StageKind, when non-empty, restricts to stages of this kind only.
	StageKind domain.StageKind
	// DueBefore, when non-nil, restricts to stages whose due_at is non-null
	// AND <= this timestamp. A stage with a NULL due_at (no SLA configured)
	// never matches a DueBefore filter — no-fallback principle, spec §11.
	DueBefore *time.Time
	// Oversee, when true, lists ALL in-progress instances in the tenant (not
	// just the actor's own eligible-pool items) and REQUIRES the caller
	// already hold CapApprovalOversee — enforced by the caller (ListWorklist)
	// via authz.Require before this filter is applied, never trusted as a
	// client-asserted bypass.
	Oversee bool
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
		areaCode, found, err := loadInstanceAreaCode(ctx, tx, s.cdRead, tenantID, instanceID)
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

		if err := s.requireInstanceVisible(ctx, tx, tenantID, loaded, areaCode); err != nil {
			return err
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

		areaCode, found, err := loadInstanceAreaCode(ctx, tx, s.cdRead, tenantID, loaded.ID)
		if err != nil {
			return fmt.Errorf("read load active instance by document: load area: %w", err)
		}
		if !found {
			return infrastructure.ErrNoActiveInstance
		}

		if err := s.requireInstanceVisible(ctx, tx, tenantID, loaded, areaCode); err != nil {
			return err
		}

		inst = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// LoadInstanceByDocumentForView is LoadActiveInstanceByDocument's view-only
// sibling: it calls the repo's LoadInstanceByDocumentForView (whose status
// filter also admits changes_requested) instead of LoadActiveInstanceByDocument,
// but otherwise enforces the identical view-authz gate (CapDocumentView with
// the CapApprovalOversee fallback, area-scoped visibility). This is what lets
// the FE author-facing GET /documents/{id}/approval-instance read back a
// non-terminal changes_requested instance after a request_changes verdict.
// Publish and mutation flows must keep using LoadActiveInstanceByDocument /
// LoadActiveInstanceByDocumentForMutation, not this method.
func (s *ReadService) LoadInstanceByDocumentForView(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, error) {
	var inst *domain.Instance
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		loaded, err := s.loadInstanceByDocumentForViewTx(ctx, tx, tenantID, documentID)
		if err != nil {
			return err
		}
		inst = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// LoadInstanceByDocumentForViewWithViewer is LoadInstanceByDocumentForView's
// viewer-facts sibling (F2d.1, ADR 0078): same load + visibility gate, inside
// the SAME view tx it additionally loads the caller's active delegations
// (plain SELECT, H-PRE-1 safe) and computes domain.ViewerFacts via the
// SHARED write-path eligibility primitives (domain.ViewerEligibility, which
// composes ResolveEligibleIdentity + CheckSoD) — never a second membership or
// SoD rule.
func (s *ReadService) LoadInstanceByDocumentForViewWithViewer(ctx context.Context, runner db.TxRunner, tenantID, documentID string) (*domain.Instance, domain.ViewerFacts, []domain.ReviewVerdict, error) {
	var inst *domain.Instance
	var vf domain.ViewerFacts
	var verdicts []domain.ReviewVerdict
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		loaded, err := s.loadInstanceByDocumentForViewTx(ctx, tx, tenantID, documentID)
		if err != nil {
			return err
		}

		viewerID, err := authz.MustActorID(ctx, tx)
		if err != nil {
			return err
		}

		delegations, err := s.repo.LoadActiveDelegationsFor(ctx, tx, tenantID, viewerID, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("read load instance by document with viewer: load delegations: %w", err)
		}

		// Verdict history (F2d.2, ADR 0079): all stages, chronological. Plain
		// SELECT in the same view tx (H-PRE-1 safe — no lock held on this path).
		loadedVerdicts, err := s.repo.LoadInstanceVerdicts(ctx, tx, tenantID, loaded.ID)
		if err != nil {
			return fmt.Errorf("read load instance by document with viewer: load verdicts: %w", err)
		}

		inst = loaded
		vf = domain.ViewerEligibility(viewerID, loaded.SubmittedBy, loaded.Active(), delegations, loaded.Stages)
		verdicts = loadedVerdicts
		return nil
	})
	if err != nil {
		return nil, domain.ViewerFacts{}, nil, err
	}
	return inst, vf, verdicts, nil
}

// loadInstanceByDocumentForViewTx is the shared body of
// LoadInstanceByDocumentForView and LoadInstanceByDocumentForViewWithViewer:
// load-by-document, resolve area, enforce the view-authz gate (CapDocumentView
// with the CapApprovalOversee fallback), then requireInstanceVisible. Callers
// run this inside their own runner.Do so a viewer-facts caller can extend the
// SAME tx with additional in-tx reads (e.g. delegations) after this returns.
func (s *ReadService) loadInstanceByDocumentForViewTx(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (*domain.Instance, error) {
	ctx = authz.WithCapCache(ctx)
	// CapDocumentView is tenant-grade (iam/domain/capability_scope.go:51); pass the
	// "tenant" sentinel so the area filter is intentionally OFF — mirrors the
	// canonical documents/application/view_service.go:71. A missing instance is
	// surfaced as ErrNoActiveInstance by the repo lookup below. CapApprovalOversee
	// (M2b F3) is an explicit alternative — oversight is its own capability, never
	// a role check (ADR 0022).
	if viewErr := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); viewErr != nil {
		if overseeErr := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); overseeErr != nil {
			return nil, viewErr
		}
	}

	loaded, err := s.repo.LoadInstanceByDocumentForView(ctx, tx, tenantID, documentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, infrastructure.ErrNoActiveInstance
		}
		return nil, err
	}
	if loaded == nil {
		return nil, infrastructure.ErrNoActiveInstance
	}

	areaCode, found, err := loadInstanceAreaCode(ctx, tx, s.cdRead, tenantID, loaded.ID)
	if err != nil {
		return nil, fmt.Errorf("read load instance by document for view: load area: %w", err)
	}
	if !found {
		return nil, infrastructure.ErrNoActiveInstance
	}

	if err := s.requireInstanceVisible(ctx, tx, tenantID, loaded, areaCode); err != nil {
		return nil, err
	}

	return loaded, nil
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

// ListWorklist (F8, spec.md §4/W4, §6.3 P2/P3/P8) is the filtered worklist
// listing: adds stage_kind / due_before / scope=oversee on top of the base
// ListInboxItemsWithTotal contract, in its own method (rather than widening
// ListInboxItemsWithTotal's positional signature) to avoid breaking the 5
// existing call sites/fakes pinned to that signature. Like the other read
// paths, filtering is enforced at THIS query/repository layer — never
// client-side — so a caller cannot see oversight-scope rows without actually
// holding CapApprovalOversee, and a stage without an SLA never satisfies a
// DueBefore filter (no-fallback principle, spec §11).
func (s *ReadService) ListWorklist(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, filter InboxFilter, limit, offset int) ([]InboxView, int, error) {
	limit = pagination.ClampLimit(limit)
	if offset < 0 {
		offset = 0
	}

	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return nil, 0, fmt.Errorf("list worklist: marshal actor: %w", err)
	}

	var items []InboxView
	var total int
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		if filter.Oversee {
			// scope=oversee REQUIRES CapApprovalOversee — never trusted from the
			// request alone (P2/P3/P8, spec.md §6.3). Denial surfaces as the
			// standard authz.ErrCapDenied -> 403 (tier-2 gate, same shape as every
			// other capability check in this package).
			if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); err != nil {
				return err
			}
		}

		// eligibilityPredicate is switched off entirely for scope=oversee (list
		// every in-progress instance in the tenant), rather than passed a
		// wildcard actor value, so the SQL shape itself documents the two
		// distinct listing modes. $2::jsonb IS NOT NULL is a tautology (actorJSON
		// is always a valid non-nil marshaled array) kept ONLY so $2 still
		// appears in the query text with an inferable type in the oversee
		// branch — Postgres's parameter-type inference (SQLSTATE 42P18) fails
		// if a placeholder is bound but never referenced anywhere in the SQL,
		// which happened when this branch was the bare literal TRUE. F10
		// live-QA caught this: scope=oversee 500'd on every call against real
		// Postgres (never caught by sqlmock-based unit tests, which don't
		// model Postgres's parameter-type inference at all).
		eligibilityPredicate := "asi.eligible_actor_ids @> $2::jsonb"
		if filter.Oversee {
			eligibilityPredicate = "($2::jsonb IS NOT NULL)"
		}

		stageKindPredicate := "$6 = '' OR asi.stage_kind = $6"
		stageKindArg := string(filter.StageKind)

		dueBeforePredicate := "$7::timestamptz IS NULL OR (asi.due_at IS NOT NULL AND asi.due_at <= $7::timestamptz)"
		var dueBeforeArg *time.Time
		if filter.DueBefore != nil {
			v := filter.DueBefore.UTC()
			dueBeforeArg = &v
		}

		totalSelect := "COUNT(*) OVER() AS total_count"

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
				asi.stage_kind,
				asi.due_at,
				`+totalSelect+`
			FROM approval_instances ai
			JOIN approval_stage_instances asi
			  ON asi.approval_instance_id = ai.id
			 AND asi.status = 'active'
			LEFT JOIN documents d
			  ON d.id = ai.document_id AND d.tenant_id = ai.tenant_id
			WHERE ai.tenant_id = $1::uuid
			  AND ai.status = 'in_progress'
			  AND (`+eligibilityPredicate+`)
			  AND ($3 = '' OR asi.area_code_snapshot = $3)
			  AND (`+stageKindPredicate+`)
			  AND (`+dueBeforePredicate+`)
			ORDER BY ai.submitted_at DESC, ai.id DESC
			LIMIT $4 OFFSET $5`,
			tenantID, actorJSON, areaCode, limit, offset, stageKindArg, dueBeforeArg,
		)
		if err != nil {
			return fmt.Errorf("list worklist: query: %w", err)
		}

		for rows.Next() {
			var v InboxView
			var signed, required, rowTotal int
			var stageKind string
			var dueAt sql.NullTime
			if err := rows.Scan(
				&v.InstanceID, &v.DocumentID, &v.ControlledDocumentID, &v.DocumentTitle,
				&v.AreaCode, &v.SubmittedBy, &v.SubmittedAt,
				&v.StageLabel, &required, &signed, &stageKind, &dueAt, &rowTotal,
			); err != nil {
				rows.Close()
				return fmt.Errorf("list worklist: scan: %w", err)
			}
			v.QuorumProgress = fmt.Sprintf("%d/%d", signed, required)
			v.StageKind = domain.StageKind(stageKind)
			if dueAt.Valid {
				t := dueAt.Time
				v.DueAt = &t
			}
			items = append(items, v)
			total = rowTotal
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list worklist: rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		// Empty page (offset past end, or genuinely zero matches): COUNT(*)
		// OVER() only appears on returned rows, mirroring the
		// ListInboxItemsWithTotal empty-page fallback. A second, filter-aware
		// count query — CountPendingForActor's WHERE shape does not know about
		// stage_kind/due_before/oversee, so it cannot be reused here.
		count, err := s.countWorklist(ctx, runner, tenantID, actorID, areaCode, filter)
		if err != nil {
			return nil, 0, err
		}
		total = count
	}
	return items, total, nil
}

// countWorklist mirrors ListWorklist's WHERE clause (minus LIMIT/OFFSET) for
// the empty-page total fallback. Kept in lockstep with ListWorklist's
// predicates deliberately — any new filter added to one must be added to
// both.
func (s *ReadService) countWorklist(ctx context.Context, runner db.TxRunner, tenantID, actorID, areaCode string, filter InboxFilter) (int, error) {
	actorJSON, err := json.Marshal([]string{actorID})
	if err != nil {
		return 0, fmt.Errorf("count worklist: marshal actor: %w", err)
	}

	var total int
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		ctx := authz.WithCapCache(ctx)
		if filter.Oversee {
			if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); err != nil {
				return err
			}
		}

		// See ListWorklist's identical comment: $2 must stay referenced (with
		// an inferable type) in both branches or Postgres's parameter-type
		// inference fails closed with SQLSTATE 42P18 on the oversee branch.
		eligibilityPredicate := "asi.eligible_actor_ids @> $2::jsonb"
		if filter.Oversee {
			eligibilityPredicate = "($2::jsonb IS NOT NULL)"
		}
		stageKindArg := string(filter.StageKind)
		var dueBeforeArg *time.Time
		if filter.DueBefore != nil {
			v := filter.DueBefore.UTC()
			dueBeforeArg = &v
		}

		err := tx.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT ai.id)
			FROM approval_instances ai
			JOIN approval_stage_instances asi
			  ON asi.approval_instance_id = ai.id
			 AND asi.status = 'active'
			WHERE ai.tenant_id = $1::uuid
			  AND ai.status = 'in_progress'
			  AND (`+eligibilityPredicate+`)
			  AND ($3 = '' OR asi.area_code_snapshot = $3)
			  AND ($4 = '' OR asi.stage_kind = $4)
			  AND ($5::timestamptz IS NULL OR (asi.due_at IS NOT NULL AND asi.due_at <= $5::timestamptz))`,
			tenantID, actorJSON, areaCode, stageKindArg, dueBeforeArg,
		).Scan(&total)
		if err != nil {
			return fmt.Errorf("count worklist: query: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
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

// requireInstanceVisible enforces the F8 visibility-gating model (spec.md
// §6.3) on a single already-loaded in-flight instance: a bare tenant-grade
// CapDocumentView holder (or CapApprovalOversee holder without a more
// specific relationship) is NOT automatically entitled to see every in-flight
// instance in the tenant — visibility is scoped to {the submitting author,
// any current-or-past stage pool member (eligible_actor_ids on ANY stage,
// not just the active one — a completed stage's reviewers keep visibility
// into the instance they acted on) OR an active DELEGATE of such a member,
// a CapApprovalOversee holder, or a CapDocumentEdit holder}. Returns
// infrastructure.ErrInstanceNotVisible (mapped to 404, "cross-boundary =
// not-found") when none of these hold.
//
// Delegation-awareness (F2d.1, ADR 0078 completing ADR 0075 × ADR 0077): the
// pool-membership branch composes on the SAME single eligibility primitive the
// write path uses — domain.CheckEligibility for the direct fast path, then
// domain.ResolveEligibleIdentity over the actor's active delegations on the
// fallback. Read-visibility is therefore a projection of act-eligibility and
// cannot diverge from it (the split-brain M2d exists to kill): a delegate who
// can sign a stage (decision_service.go / review_verdict_service.go, delegation
// -aware) can now also LOAD the instance, instead of a prior 404. The delegation
// SELECT runs only after the direct check misses, so the common non-delegated
// read pays no extra round-trip (mirrors ADR 0077's fallback-query-only
// discipline on the write path).
//
// This tightens the pre-F8 behavior documented in
// read_service_tenant_grade_view_integration_test.go (a bare tenant-grade
// viewer could load ANY in-flight instance) — the ratified spec §6.3
// explicitly frames that prior breadth as the exact remediation gap M2b
// exists to close, not a contradiction to preserve.
//
// Deliberately query/predicate-layer enforcement (reads the actor id off the
// tx's own seeded GUC via authz.MustActorID, checks eligible_actor_ids
// already loaded onto the domain aggregate, and re-runs authz.Require for
// the two capability alternatives) — NEVER a client-side filter. A caller
// cannot bypass this by requesting the same instance_id twice or by any
// client-controlled parameter.
//
// areaCode is the instance's OWN resolved area (from loadInstanceAreaCode),
// NOT the "tenant" sentinel: CapApprovalOversee is tenant-grade
// (capability_scope.go) so it intentionally passes "tenant" to skip the area
// filter, but CapDocumentEdit is area-grade — passing "tenant" for it would
// wrongly grant edit-anywhere-in-tenant semantics. An empty areaCode (no area
// resolvable anywhere in the chain) fails closed for CapDocumentEdit for
// non-system actors, mirroring documents/application/fillin_authz.go's
// requireDocEditDraft.
func (s *ReadService) requireInstanceVisible(ctx context.Context, tx *sql.Tx, tenantID string, inst *domain.Instance, areaCode string) error {
	actorID, err := authz.MustActorID(ctx, tx)
	if err != nil {
		return err
	}

	if inst.SubmittedBy == actorID {
		return nil
	}

	// Pool membership across ALL stages (current-or-past — a completed stage's
	// reviewers keep visibility, per ADR 0075). Evaluated through the SAME
	// eligibility primitive the write path uses so visibility cannot diverge
	// from act-eligibility.
	var pool []string
	for i := range inst.Stages {
		pool = append(pool, inst.Stages[i].EligibleActorIDs...)
	}
	// Direct fast path: no delegation load for the common non-delegated read.
	if domain.CheckEligibility(actorID, pool) == nil {
		return nil
	}
	// Delegation fallback (ADR 0077): an active delegate of any pool member is
	// visible exactly as they are eligible to act. In-tx plain SELECT, H-PRE-1
	// safe; runs only after the direct check misses.
	delegations, err := s.repo.LoadActiveDelegationsFor(ctx, tx, tenantID, actorID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("require instance visible: load delegations: %w", err)
	}
	if _, err := domain.ResolveEligibleIdentity(actorID, pool, delegations); err == nil {
		return nil
	}

	// Not the author, not a (possibly delegated) pool member on any stage: fall
	// back to the two oversight/edit capabilities. Either satisfies visibility;
	// system_admin already short-circuits inside authz.Require itself.
	if err := authz.Require(ctx, tx, string(iamdomain.CapApprovalOversee), "tenant"); err == nil {
		return nil
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), areaCode); err == nil {
		return nil
	}

	return infrastructure.ErrInstanceNotVisible
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
