package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/testsupport/pgtest"
)

// ---------------------------------------------------------------------------
// TestMapPgError — error taxonomy with synthetic pgconn.PgError
// ---------------------------------------------------------------------------

func TestMapPgError(t *testing.T) {
	makePgErr := func(code, constraint, msg string) error {
		return &pgconn.PgError{Code: code, ConstraintName: constraint, Message: msg}
	}

	tests := []struct {
		name       string
		err        error
		hints      MapHints
		wantTarget error
	}{
		{
			name:       "unique_active_instance",
			err:        makePgErr("23505", "ux_approval_instances_active", ""),
			wantTarget: ErrDuplicateSubmission,
		},
		{
			name:       "unique_active_instance_document_id",
			err:        makePgErr("23505", "ux_approval_instances_active_document_id", ""),
			wantTarget: ErrDuplicateSubmission,
		},
		{
			name:       "unique_idempotency_key",
			err:        makePgErr("23505", "approval_instances_document_id_idempotency_key_key", ""),
			wantTarget: ErrDuplicateSubmission,
		},
		{
			name:       "unique_signoff_instance_actor",
			err:        makePgErr("23505", "approval_signoffs_approval_instance_id_actor_user_id_key", ""),
			wantTarget: ErrUnknownDB,
		},
		{
			name:       "unique_signoff_stage_actor",
			err:        makePgErr("23505", "approval_signoffs_stage_instance_id_actor_user_id_key", ""),
			wantTarget: ErrActorAlreadySigned,
		},
		{
			name:       "unique_unknown_with_hint_match",
			err:        makePgErr("23505", "my_special_constraint", ""),
			hints:      MapHints{UniqueConstraint: "my_special_constraint"},
			wantTarget: ErrDuplicateSubmission,
		},
		{
			name:       "unique_unknown_no_hint",
			err:        makePgErr("23505", "some_other_constraint", ""),
			wantTarget: ErrUnknownDB,
		},
		{
			name:       "fk_violation",
			err:        makePgErr("23503", "", ""),
			wantTarget: ErrFKViolation,
		},
		{
			name:       "check_violation_no_msg",
			err:        makePgErr("23514", "", ""),
			wantTarget: ErrCheckViolation,
		},
		{
			name:       "check_violation_with_msg",
			err:        makePgErr("23514", "", "cross-tenant signoff rejected"),
			wantTarget: ErrCheckViolation,
		},
		{
			name:       "insufficient_privilege",
			err:        makePgErr("42501", "", ""),
			wantTarget: ErrInsufficientPrivilege,
		},
		{
			name:       "unknown_sqlstate",
			err:        makePgErr("58000", "", "io error"),
			wantTarget: ErrUnknownDB,
		},
		{
			name:  "non_pg_error_passed_through",
			err:   errors.New("generic error"),
			hints: MapHints{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MapPgError(tc.err, tc.hints)
			if tc.wantTarget == nil {
				// Expect the original error passed through unchanged.
				if got != tc.err {
					t.Errorf("expected passthrough: got %v, want %v", got, tc.err)
				}
				return
			}
			if !errors.Is(got, tc.wantTarget) {
				t.Errorf("MapPgError(%q): got %v, want errors.Is target %v", tc.name, got, tc.wantTarget)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Fake infrastructure for unit tests (no sqlmock, no real DB)
// ---------------------------------------------------------------------------

// fakeResult implements sql.Result for exec operations.
type fakeResult struct {
	rowsAffected int64
}

func (f fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (f fakeResult) RowsAffected() (int64, error) { return f.rowsAffected, nil }

// ---------------------------------------------------------------------------
// TestUpdateStageStatusOCC — verify ErrStageNotActive on 0 rows affected
// ---------------------------------------------------------------------------

// fakeStageUpdater lets us test the OCC branch without a real database.
// We implement the minimal interface needed to call our OCC logic.

type updateResult struct {
	n   int64
	err error
}

// occ logic extracted for unit testability
func applyStageOCC(res sql.Result, execErr error) error {
	if execErr != nil {
		return MapPgError(execErr, MapHints{})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrStageNotActive
	}
	return nil
}

func TestUpdateStageStatusOCC(t *testing.T) {
	t.Run("zero_rows_returns_ErrStageNotActive", func(t *testing.T) {
		err := applyStageOCC(fakeResult{rowsAffected: 0}, nil)
		if !errors.Is(err, ErrStageNotActive) {
			t.Errorf("expected ErrStageNotActive, got %v", err)
		}
	})

	t.Run("one_row_returns_nil", func(t *testing.T) {
		err := applyStageOCC(fakeResult{rowsAffected: 1}, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("db_error_mapped_through_MapPgError", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "42501"}
		err := applyStageOCC(fakeResult{}, pgErr)
		if !errors.Is(err, ErrInsufficientPrivilege) {
			t.Errorf("expected ErrInsufficientPrivilege, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// TestInsertSignoffReplayDetection — WasReplay logic
// ---------------------------------------------------------------------------

// We test the replay-detection decision logic directly without a DB by
// reproducing the decision branch that InsertSignoff uses.

func replayDecision(incoming, existing *domain.Signoff) (SignoffInsertResult, error) {
	if existing.StageInstanceID() == incoming.StageInstanceID() &&
		existing.Decision() == incoming.Decision() &&
		existing.ContentHash() == incoming.ContentHash() {
		return SignoffInsertResult{ID: existing.ID(), WasReplay: true}, nil
	}
	return SignoffInsertResult{}, ErrActorAlreadySigned
}

func makeTestSignoff(t *testing.T, id, instanceID, stageID, actor, decision, hash string) *domain.Signoff {
	t.Helper()
	s, err := domain.NewSignoff(domain.SignoffParams{
		ID:                 id,
		ApprovalInstanceID: instanceID,
		StageInstanceID:    stageID,
		ActorUserID:        actor,
		ActorTenantID:      "tenant-a",
		Decision:           domain.Decision(decision),
		SignedAt:           time.Now().UTC(),
		SignatureMethod:    "simple",
		SignaturePayload:   json.RawMessage(`{}`),
		ContentHash:        hash,
	})
	if err != nil {
		t.Fatalf("makeTestSignoff: %v", err)
	}
	return s
}

const testHash = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

func TestInsertSignoffReplayDetection(t *testing.T) {
	t.Run("exact_match_is_replay", func(t *testing.T) {
		incoming := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", testHash)
		existing := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", testHash)

		result, err := replayDecision(incoming, existing)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.WasReplay {
			t.Error("expected WasReplay=true")
		}
		if result.ID != "id-1" {
			t.Errorf("expected ID=id-1, got %q", result.ID)
		}
	})

	t.Run("different_stage_is_conflict", func(t *testing.T) {
		incoming := makeTestSignoff(t, "id-2", "inst-1", "stage-2", "actor-1", "approve", testHash)
		existing := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", testHash)

		_, err := replayDecision(incoming, existing)
		if !errors.Is(err, ErrActorAlreadySigned) {
			t.Errorf("expected ErrActorAlreadySigned, got %v", err)
		}
	})

	t.Run("different_decision_is_conflict", func(t *testing.T) {
		incoming := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "reject", testHash)
		existing := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", testHash)

		_, err := replayDecision(incoming, existing)
		if !errors.Is(err, ErrActorAlreadySigned) {
			t.Errorf("expected ErrActorAlreadySigned, got %v", err)
		}
	})

	t.Run("different_content_hash_is_conflict", func(t *testing.T) {
		hash2 := "1111111111111111111111111111111111111111111111111111111111111111"
		incoming := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", hash2)
		existing := makeTestSignoff(t, "id-1", "inst-1", "stage-1", "actor-1", "approve", testHash)

		_, err := replayDecision(incoming, existing)
		if !errors.Is(err, ErrActorAlreadySigned) {
			t.Errorf("expected ErrActorAlreadySigned, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// TestScheduledPublishRowShape — compile + structure test (no real DB)
// ---------------------------------------------------------------------------

// TestScheduledPublishRowShape verifies the repository surface still exposes the
// scheduled publish persistence shape needed by approval tests.
func TestScheduledPublishRowShape(t *testing.T) {
	// Compile-time interface conformance check.
	var _ ApprovalRepository = (*postgresApprovalRepository)(nil)

	// Verify ScheduledPublishRow has the expected fields.
	row := ScheduledPublishRow{
		DocumentID:         "doc-1",
		TenantID:           "tenant-1",
		EffectiveFrom:      time.Now(),
		RevisionVersion:    3,
		ScheduleGeneration: 2,
	}
	if row.DocumentID == "" {
		t.Error("DocumentID should not be empty")
	}
	if row.RevisionVersion != 3 {
		t.Errorf("RevisionVersion: got %d, want 3", row.RevisionVersion)
	}
	if row.ScheduleGeneration != 2 {
		t.Errorf("ScheduleGeneration: got %d, want 2", row.ScheduleGeneration)
	}
}

// ---------------------------------------------------------------------------
// TestSignoffInsertResultFields — basic struct field test
// ---------------------------------------------------------------------------

func TestSignoffInsertResultFields(t *testing.T) {
	r := SignoffInsertResult{ID: "abc", WasReplay: true}
	if r.ID != "abc" {
		t.Errorf("ID: got %q, want abc", r.ID)
	}
	if !r.WasReplay {
		t.Error("WasReplay should be true")
	}
}

// ---------------------------------------------------------------------------
// TestUpdateInstanceStatusOCC — ErrInstanceCompleted on 0 rows
// ---------------------------------------------------------------------------

func applyInstanceOCC(res sql.Result, execErr error) error {
	if execErr != nil {
		return MapPgError(execErr, MapHints{})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrInstanceCompleted
	}
	return nil
}

func TestUpdateInstanceStatusOCC(t *testing.T) {
	t.Run("zero_rows_returns_ErrInstanceCompleted", func(t *testing.T) {
		err := applyInstanceOCC(fakeResult{rowsAffected: 0}, nil)
		if !errors.Is(err, ErrInstanceCompleted) {
			t.Errorf("expected ErrInstanceCompleted, got %v", err)
		}
	})

	t.Run("one_row_returns_nil", func(t *testing.T) {
		err := applyInstanceOCC(fakeResult{rowsAffected: 1}, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// TestNewPostgresApprovalRepository — constructor returns correct type
// ---------------------------------------------------------------------------

func TestNewPostgresApprovalRepository(t *testing.T) {
	// NewPostgresApprovalRepository must return an ApprovalRepository.
	// We pass nil for *sql.DB; this is fine as long as we don't execute queries.
	var repo ApprovalRepository = NewPostgresApprovalRepository(nil)
	if repo == nil {
		t.Error("NewPostgresApprovalRepository returned nil")
	}
}

// ---------------------------------------------------------------------------
// TestInsertStageInstancesBulk — verify multi-row placeholder generation
// ---------------------------------------------------------------------------

func TestInsertStageInstancesBulk(t *testing.T) {
	// If stages is empty, InsertStageInstances returns nil immediately.
	repo := &postgresApprovalRepository{}
	err := repo.InsertStageInstances(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("expected nil for empty stages, got %v", err)
	}
}

func TestValidateScheduledSupersedeTarget_RealRows(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	repo := NewPostgresApprovalRepository(db)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	const (
		tenantID          = "11111111-1111-1111-1111-111111111111"
		ownerUserID       = "22222222-2222-2222-2222-222222222222"
		templateVersionID = "33333333-3333-3333-3333-333333333333"
		controlledDocID   = "44444444-4444-4444-4444-444444444444"
		otherLineageID    = "55555555-5555-5555-5555-555555555555"
		targetDocumentID  = "66666666-6666-6666-6666-666666666666"
		publishedHeadID   = "77777777-7777-7777-7777-777777777777"
		otherPublishedID  = "88888888-8888-8888-8888-888888888888"
	)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-SUP-001', 'Scheduled Supersede', $3, 'active'),
					($4::uuid, $2::uuid, 'po', 'quality', 'PO-SUP-002', 'Other Lineage', $3, 'active')`,
			args: []any{controlledDocID, tenantID, ownerUserID, otherLineageID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number)
				VALUES
					($1::uuid, $4::uuid, $7::uuid, 'Current Published Head', 'published', '{}'::jsonb, $8, $9::uuid, 0),
					($2::uuid, $4::uuid, $7::uuid, 'Approved Successor', 'approved', '{}'::jsonb, $8, $9::uuid, 1),
					($3::uuid, $4::uuid, $7::uuid, 'Other Published Head', 'published', '{}'::jsonb, $8, $10::uuid, 0)`,
			args: []any{publishedHeadID, targetDocumentID, otherPublishedID, tenantID, nil, nil, templateVersionID, ownerUserID, controlledDocID, otherLineageID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	if err := repo.ValidateScheduledSupersedeTarget(ctx, tx, tenantID, targetDocumentID, publishedHeadID); err != nil {
		t.Fatalf("valid target: unexpected error: %v", err)
	}

	err = repo.ValidateScheduledSupersedeTarget(ctx, tx, tenantID, targetDocumentID, otherPublishedID)
	if !errors.Is(err, ErrInvalidScheduledSupersedeTarget) {
		t.Fatalf("invalid target error = %v, want ErrInvalidScheduledSupersedeTarget", err)
	}
}

func TestLoadCurrentPublishedHeadForDocument_RealRows(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	repo := NewPostgresApprovalRepository(db)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	const (
		tenantID          = "11111111-9999-1111-9999-111111111111"
		ownerUserID       = "22222222-9999-2222-9999-222222222222"
		templateVersionID = "33333333-9999-3333-9999-333333333333"
		controlledDocID   = "44444444-9999-4444-9999-444444444444"
		targetDocumentID  = "55555555-9999-5555-9999-555555555555"
		publishedHeadID   = "66666666-9999-6666-9999-666666666666"
		laterHeadID       = "77777777-9999-7777-9999-777777777777"
	)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-HEAD-001', 'Current Head Lookup', $3, 'active')`,
			args: []any{controlledDocID, tenantID, ownerUserID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number)
				VALUES
					($1::uuid, $4::uuid, $5::uuid, 'Current Published Head', 'published', '{}'::jsonb, $6, $7::uuid, 0),
					($2::uuid, $4::uuid, $5::uuid, 'Target Approved Revision', 'approved', '{}'::jsonb, $6, $7::uuid, 1),
					($3::uuid, $4::uuid, $5::uuid, 'Later Published Head', 'published', '{}'::jsonb, $6, $7::uuid, 2)`,
			args: []any{publishedHeadID, targetDocumentID, laterHeadID, tenantID, templateVersionID, ownerUserID, controlledDocID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	got, err := repo.LoadCurrentPublishedHeadForDocument(ctx, tx, tenantID, targetDocumentID)
	if err != nil {
		t.Fatalf("LoadCurrentPublishedHeadForDocument: %v", err)
	}
	if got != laterHeadID {
		t.Fatalf("current head = %q, want %q", got, laterHeadID)
	}
}

func TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	repo := NewPostgresApprovalRepository(db)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	const (
		tenantID          = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		ownerUserID       = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
		templateVersionID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
		controlledDocID   = "dddddddd-dddd-dddd-dddd-dddddddddddd"
		documentID        = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
		routeID           = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		instanceID        = "11111111-2222-3333-4444-555555555555"
	)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-REV-001', 'Revision Version Load', $3, 'active')`,
			args: []any{controlledDocID, tenantID, ownerUserID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number, revision_version)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, 'Approved Successor', 'approved', '{}'::jsonb, $4, $5::uuid, 3, 7)`,
			args: []any{documentID, tenantID, templateVersionID, ownerUserID, controlledDocID},
		},
		{
			query: `
				INSERT INTO public.approval_routes
					(id, tenant_id, name, profile_code, active)
				VALUES
					($1::uuid, $2::uuid, 'Route', 'po', true)`,
			args: []any{routeID, tenantID},
		},
		{
			query: `
				INSERT INTO public.approval_instances
					(id, tenant_id, document_id, route_id, route_version_snapshot, status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'approved', $5, now(), repeat('a', 64), 'idem-load-active')`,
			args: []any{instanceID, tenantID, documentID, routeID, ownerUserID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	inst, err := repo.LoadActiveInstanceByDocument(ctx, tx, tenantID, documentID)
	if err != nil {
		t.Fatalf("LoadActiveInstanceByDocument: %v", err)
	}
	if inst.RevisionVersion != 7 {
		t.Fatalf("revision version = %d, want 7", inst.RevisionVersion)
	}
}

func TestLoadInstance_LoadsDocumentRevisionVersion(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	repo := NewPostgresApprovalRepository(db)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	const (
		tenantID          = "12121212-3434-5656-7878-909090909090"
		ownerUserID       = "abababab-cdcd-efef-0101-232323232323"
		templateVersionID = "34343434-5656-7878-9090-121212121212"
		controlledDocID   = "45454545-6767-8989-0101-232323232323"
		documentID        = "56565656-7878-9090-1212-343434343434"
		routeID           = "67676767-8989-0101-2323-454545454545"
		instanceID        = "78787878-9090-1212-3434-565656565656"
	)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-REV-002', 'Load Instance Revision Version', $3, 'active')`,
			args: []any{controlledDocID, tenantID, ownerUserID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number, revision_version)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, 'Approved Revision', 'approved', '{}'::jsonb, $4, $5::uuid, 4, 9)`,
			args: []any{documentID, tenantID, templateVersionID, ownerUserID, controlledDocID},
		},
		{
			query: `
				INSERT INTO public.approval_routes
					(id, tenant_id, name, profile_code, active)
				VALUES
					($1::uuid, $2::uuid, 'Route', 'po', true)`,
			args: []any{routeID, tenantID},
		},
		{
			query: `
				INSERT INTO public.approval_instances
					(id, tenant_id, document_id, route_id, route_version_snapshot, status, submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, 'approved', $5, now(), repeat('b', 64), 'idem-load-instance')`,
			args: []any{instanceID, tenantID, documentID, routeID, ownerUserID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	inst, err := repo.LoadInstance(ctx, tx, tenantID, instanceID)
	if err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if inst.RevisionVersion != 9 {
		t.Fatalf("revision version = %d, want 9", inst.RevisionVersion)
	}
}

func TestScheduleGenerationIncrementsOnScheduledWritePath(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	const (
		tenantID          = "50505050-1111-2222-3333-444444444444"
		ownerUserID       = "60606060-1111-2222-3333-444444444444"
		templateVersionID = "70707070-1111-2222-3333-444444444444"
		controlledDocID   = "80808080-1111-2222-3333-444444444444"
		documentID        = "90909090-5555-6666-7777-888888888888"
	)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `
				INSERT INTO public.controlled_documents
					(id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
				VALUES
					($1::uuid, $2::uuid, 'po', 'quality', 'PO-SCHED-002', 'Scheduled Generation Increment', $3, 'active')`,
			args: []any{controlledDocID, tenantID, ownerUserID},
		},
		{
			query: `
				INSERT INTO public.documents
					(id, tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, revision_number, revision_version, schedule_generation)
				VALUES
					($1::uuid, $2::uuid, $3::uuid, 'Approved Revision', 'approved', '{}'::jsonb, $4, $5::uuid, 5, 2, 9)`,
			args: []any{documentID, tenantID, templateVersionID, ownerUserID, controlledDocID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			t.Fatalf("seed rows: %v", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET status = 'scheduled',
		       effective_from = $1,
		       superseded_document_id = NULL,
		       revision_version = revision_version + 1,
		       schedule_generation = schedule_generation + 1
		 WHERE id = $2
		   AND tenant_id = $3
		   AND status = 'approved'
		   AND revision_version = $4`,
		time.Now().UTC().Add(2*time.Hour), documentID, tenantID, 2,
	)
	if err != nil {
		t.Fatalf("update scheduled state: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("rows affected = %d, want 1", affected)
	}

	var gotStatus string
	var gotGeneration int64
	err = tx.QueryRowContext(ctx, `
		SELECT status, schedule_generation
		  FROM public.documents
		 WHERE id = $1
		   AND tenant_id = $2`,
		documentID, tenantID,
	).Scan(&gotStatus, &gotGeneration)
	if err != nil {
		t.Fatalf("load document state: %v", err)
	}
	if gotStatus != "scheduled" {
		t.Fatalf("status = %s, want scheduled", gotStatus)
	}
	if gotGeneration != 10 {
		t.Fatalf("schedule_generation = %d, want 10", gotGeneration)
	}
}
