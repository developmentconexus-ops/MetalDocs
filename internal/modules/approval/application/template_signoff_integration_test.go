//go:build integration
// +build integration

package application

// P3.S2b-3b-iii-b (M3 kernel extraction) — real-DB proof of the template
// version LIFECYCLE under kernel governance: the signoff-completion leg of
// DecisionService.recordSignoffInTx driving a (subject_kind='template')
// instance to a terminal outcome, plus the submit-lock that closes the
// concurrent-edit hole.
//
// The subject-generic quorum engine (EvaluateQuorum/ApplyEligibilityDrift) is
// reused verbatim; only the terminal side effects differ by subject kind:
//   - cap asserted at signoff: template.approve (not document.signoff)
//   - content hash: read off templates_template_version.content_hash via the
//     approval-owned TemplateVersionReader port (no freeze pin, no client echo)
//   - completion: MarkTemplateVersionApproved / MarkTemplateVersionRejected via
//     the approval-owned TemplateCompletionWriter port (no documents UPDATE, no
//     async-freeze Pin, no lifecycle enqueue, no PDF dispatch)
//
// Every test drives the REAL templates-infrastructure adapters
// (templatesinfra.ApprovalVersionReader / ApprovalCompletionWriter) end-to-end
// so the module seam itself is exercised, not a local double. This is a
// _test.go file, so importing templates/infrastructure and templates/application
// is boundary-clean (check-module-boundaries.ps1 excludes test files; precedent:
// documents/application/fillin_schema_parity_integration_test.go).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	templatesapp "metaldocs/internal/modules/templates/application"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// templateApprovalFixture is the shared, submitted-and-locked starting state
// for the signoff tests: a template version flipped draft -> under_review by
// the real submit-lock, an in_progress instance with stage 1 active + stage 2
// pending, and two eligible system_admin signers (distinct from the submitter,
// so SoD is satisfiable).
type templateApprovalFixture struct {
	dbc        *sql.DB
	tenantID   string
	templateID string
	versionID  string
	instanceID string
	stage1ID   string
	stage2ID   string
	reviewerID string // stage 1 approver-role signer
	managerID  string // stage 2 author-role signer
	services   *Services
	repo       approvalrepo.ApprovalRepository
	runner     db.TxRunner
}

// setupTemplateApproval seeds a draft template version + active template route,
// wires a Services with the REAL templates-infra reader + completion writer,
// and runs the real submit path (which locks the version under_review). It
// returns the fixture positioned exactly where the first signoff can happen.
func setupTemplateApproval(t *testing.T) templateApprovalFixture {
	t.Helper()
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, "draft")
	seedTemplateRoute(t, dbc, tnt.ID, actor.ID, templateID)

	// Signers: system_admin (so authz.Require(template.approve, "tenant")
	// short-circuits and appends the asserted cap the 0300 signoff tripwire
	// arm demands) AND holding the eligibility role their stage requires.
	// Distinct from the submitter (actor) so domain.CheckSoD is satisfiable.
	reviewer := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	manager := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	grantAreaBlindRole(t, dbc, tnt.ID, reviewer.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, manager.ID, "author")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iampostgres.NewUserDisplayNameRepository(dbc))
	services := NewServices(repo, NewSQLEmitter(), RealClock{}, nil).
		WithTemplateVersionReader(templatesinfra.NewApprovalVersionReader()).
		WithTemplateCompletionWriter(templatesinfra.NewApprovalCompletionWriter())
	runner := db.NewTxRunner(dbc)

	res, err := services.TemplateSubmit.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("setup: SubmitTemplateVersionForReview: %v", err)
	}

	// Submit-lock precondition for the whole slice: the version is now
	// under_review, which is what makes the completion port's under_review CAS
	// satisfiable (and what the edit-lock test below proves closes edits).
	if got := templateVersionStatus(t, dbc, tnt.ID, versionID); got != "under_review" {
		t.Fatalf("setup: version status = %q, want under_review (submit-lock)", got)
	}

	ctx := context.Background()
	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("setup: begin read tx: %v", err)
	}
	defer tx.Rollback()
	inst, err := repo.LoadInstance(ctx, tx, tnt.ID, res.InstanceID)
	if err != nil {
		t.Fatalf("setup: LoadInstance: %v", err)
	}
	if len(inst.Stages) != 2 {
		t.Fatalf("setup: len(Stages) = %d, want 2", len(inst.Stages))
	}

	return templateApprovalFixture{
		dbc:        dbc,
		tenantID:   tnt.ID,
		templateID: templateID,
		versionID:  versionID,
		instanceID: res.InstanceID,
		stage1ID:   inst.Stages[0].ID,
		stage2ID:   inst.Stages[1].ID,
		reviewerID: reviewer.ID,
		managerID:  manager.ID,
		services:   services,
		repo:       repo,
		runner:     runner,
	}
}

func templateVersionStatus(t *testing.T, dbc *sql.DB, tenantID, versionID string) string {
	t.Helper()
	var status string
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status FROM templates_template_version WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		versionID, tenantID,
	).Scan(&status); err != nil {
		t.Fatalf("query version status: %v", err)
	}
	return status
}

// templateVersionApproverID reads the attribution columns the terminal
// completion port writes (F-E4-4). approver_id is TEXT and nullable, so a
// NULL is reported as "" — which is exactly the live-QA defect state.
func templateVersionApproverID(t *testing.T, dbc *sql.DB, tenantID, versionID string) string {
	t.Helper()
	var approver sql.NullString
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT approver_id FROM templates_template_version WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		versionID, tenantID,
	).Scan(&approver); err != nil {
		t.Fatalf("query version approver_id: %v", err)
	}
	if !approver.Valid {
		return ""
	}
	return approver.String
}

func instanceStatusOf(t *testing.T, dbc *sql.DB, instanceID string) string {
	t.Helper()
	var status string
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status FROM public.approval_instances WHERE id = $1::uuid`, instanceID,
	).Scan(&status); err != nil {
		t.Fatalf("query instance status: %v", err)
	}
	return status
}

func templateSignoffRowCount(t *testing.T, dbc *sql.DB, instanceID string) int {
	t.Helper()
	var n int
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.approval_signoffs WHERE approval_instance_id = $1::uuid`, instanceID,
	).Scan(&n); err != nil {
		t.Fatalf("count signoffs: %v", err)
	}
	return n
}

// TestTemplateSignoff_Approve_CompletesVersion_RealDB is the primary
// GREEN target: both stages approved drives the instance to approved AND the
// completion port transitions templates_template_version under_review ->
// approved, all through the real templates-infra adapter. Two signoff rows
// land — proving the 0300 tripwire's template arm accepted template.approve.
func TestTemplateSignoff_Approve_CompletesVersion_RealDB(t *testing.T) {
	fx := setupTemplateApproval(t)

	// Stage 1 (any_1_of, approver): reviewer approves -> stage completes,
	// stage 2 activates, instance still in_progress.
	r1, err := fx.services.Decision.RecordSignoff(newTemplateSubmitCtx(fx.tenantID, fx.reviewerID), fx.runner, SignoffRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.stage1ID,
		ActorUserID:      fx.reviewerID,
		Decision:         domain.DecisionApprove,
		Comment:          "stage 1 ok",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
	})
	if err != nil {
		t.Fatalf("stage 1 RecordSignoff: %v", err)
	}
	if !r1.StageCompleted {
		t.Errorf("stage 1: StageCompleted = false, want true")
	}
	if r1.InstanceApproved {
		t.Errorf("stage 1: InstanceApproved = true, want false (stage 2 still pending)")
	}
	if got := instanceStatusOf(t, fx.dbc, fx.instanceID); got != string(domain.InstanceInProgress) {
		t.Errorf("after stage 1: instance status = %q, want in_progress", got)
	}
	if got := templateVersionStatus(t, fx.dbc, fx.tenantID, fx.versionID); got != "under_review" {
		t.Errorf("after stage 1: version status = %q, want under_review (not yet terminal)", got)
	}

	// Stage 2 (all_of, author, sole eligible = manager): manager approves ->
	// instance approved -> completion port fires.
	r2, err := fx.services.Decision.RecordSignoff(newTemplateSubmitCtx(fx.tenantID, fx.managerID), fx.runner, SignoffRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.stage2ID,
		ActorUserID:      fx.managerID,
		Decision:         domain.DecisionApprove,
		Comment:          "stage 2 ok",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
	})
	if err != nil {
		t.Fatalf("stage 2 RecordSignoff: %v", err)
	}
	if !r2.InstanceApproved {
		t.Errorf("stage 2: InstanceApproved = false, want true")
	}

	if got := instanceStatusOf(t, fx.dbc, fx.instanceID); got != string(domain.InstanceApproved) {
		t.Errorf("final: instance status = %q, want approved", got)
	}
	// The load-bearing assertion: the completion port drove the template
	// version terminal, in the same tx as the final signoff.
	if got := templateVersionStatus(t, fx.dbc, fx.tenantID, fx.versionID); got != "approved" {
		t.Errorf("final: version status = %q, want approved (completion port)", got)
	}
	if got := templateSignoffRowCount(t, fx.dbc, fx.instanceID); got != 2 {
		t.Errorf("signoff rows = %d, want 2 (0300 template arm accepted template.approve twice)", got)
	}
	// F-E4-4: approved_at without attribution was the live-QA defect. The
	// deciding actor of the terminal signoff (manager, stage 2) must be
	// stamped onto approver_id in the very same tx.
	if got := templateVersionApproverID(t, fx.dbc, fx.tenantID, fx.versionID); got != fx.managerID {
		t.Errorf("final: approver_id = %q, want %q (deciding actor of the terminal signoff)", got, fx.managerID)
	}
}

// TestTemplateSignoff_Reject_ReturnsVersionToDraft_RealDB proves the reject
// leg: a reject on the any_1_of stage 1 collapses the instance to rejected and
// the completion port returns templates_template_version under_review -> draft
// (so the author can edit and resubmit) — with no documents-table transition
// and no cancel GUC (both document-only).
func TestTemplateSignoff_Reject_ReturnsVersionToDraft_RealDB(t *testing.T) {
	fx := setupTemplateApproval(t)

	r, err := fx.services.Decision.RecordSignoff(newTemplateSubmitCtx(fx.tenantID, fx.reviewerID), fx.runner, SignoffRequest{
		TenantID:         fx.tenantID,
		InstanceID:       fx.instanceID,
		StageInstanceID:  fx.stage1ID,
		ActorUserID:      fx.reviewerID,
		Decision:         domain.DecisionReject,
		Comment:          "needs work",
		SignatureMethod:  "password",
		SignaturePayload: map[string]any{"password_token": "unused-in-test-bypass"},
	})
	if err != nil {
		t.Fatalf("reject RecordSignoff: %v", err)
	}
	if !r.InstanceRejected {
		t.Errorf("InstanceRejected = false, want true")
	}

	if got := instanceStatusOf(t, fx.dbc, fx.instanceID); got != string(domain.InstanceRejected) {
		t.Errorf("instance status = %q, want rejected", got)
	}
	if got := templateVersionStatus(t, fx.dbc, fx.tenantID, fx.versionID); got != "draft" {
		t.Errorf("version status = %q, want draft (completion port reject -> draft)", got)
	}
}

// TestTemplateSubmit_LocksEditsClosed_RealDB proves the concurrent-edit hole
// (#24 gap folded into this slice) is closed end-to-end: after submit the
// version is under_review, so a templates draft-gated endpoint
// (Service.PresignTemplateUpload, which rejects any non-draft status) fails
// with domain.ErrInvalidStateTransition. This exercises the real templates
// module's own gate against the state the real submit-lock produced.
func TestTemplateSubmit_LocksEditsClosed_RealDB(t *testing.T) {
	fx := setupTemplateApproval(t)

	svc := templatesapp.New(templatesinfra.New(fx.dbc), nil, nil, nil)
	_, err := svc.PresignTemplateUpload(context.Background(), templatesapp.PresignTemplateUploadCmd{
		TenantID:      fx.tenantID,
		ActorUserID:   fx.reviewerID,
		TemplateID:    fx.templateID,
		VersionNumber: 1,
	})
	if !errors.Is(err, templatesdomain.ErrInvalidStateTransition) {
		t.Fatalf("PresignTemplateUpload err = %v, want ErrInvalidStateTransition (edit-lock closed)", err)
	}
}

// TestTemplateSignoff_DocumentSignoffCap_TripwireRejects_RealDB is the
// negative control that proves the 0300 tripwire's parent-discriminated arm is
// LIVE (not merely rendered): a raw INSERT into approval_signoffs against a
// template-parent instance while asserting ONLY document.signoff is rejected by
// the DB with SQLSTATE P0001. This is the security property ADR 0083 guards —
// a document.signoff holder must NOT be able to sign a template subject.
func TestTemplateSignoff_DocumentSignoffCap_TripwireRejects_RealDB(t *testing.T) {
	fx := setupTemplateApproval(t)

	ctx := context.Background()
	tx, err := fx.dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, fx.tenantID, fx.reviewerID); err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	// Assert the WRONG capability for a template subject.
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, `[{"cap":"document.signoff"}]`,
	); err != nil {
		t.Fatalf("assert caps: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.approval_signoffs
		  (id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id,
		   decision, comment, signed_at, signature_method, signature_payload, content_hash)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, $3, $4::uuid,
		   'approve', '', now(), 'password', '{}'::jsonb, $5)`,
		fx.instanceID, fx.stage1ID, fx.reviewerID, fx.tenantID,
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"[:64],
	)
	if err == nil {
		t.Fatal("raw signoff INSERT with document.signoff on a template parent succeeded; want P0001 tripwire rejection")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "P0001" {
		t.Fatalf("err = %v, want SQLSTATE P0001 (parent-discriminated tripwire)", err)
	}
}
