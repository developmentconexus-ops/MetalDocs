//go:build integration
// +build integration

// Package approval_test — SLICE 8a (unit 3.2) resolution-parity regression
// guard for TemplateSubmitService.PreviewRoute against a real Postgres.
//
// The reviewer's Major: prove the read-only approval-preview resolves the SAME
// active route, keyed the SAME way (ROUTE.subject_key = the template's
// doc_type_code since ADR 0086, NOT the template id and NOT the template
// VERSION id), and surfaces the SAME stages+selectors — including a
// submit_choice selector's role + area_code — that an ACTUAL submit for the
// same subject resolves. The independent oracle is a real end-to-end
// SubmitTemplateVersionForReview: preview.RouteID must equal the resulting
// approval instance's RouteID, and preview's stages/selectors must deep-equal
// the route the submit path loaded. If preview and submit ever diverge on
// route resolution or selector shape, this test fails.
//
// Reuses seedProcessArea / grantActiveArea from
// resolve_eligible_actors_for_selectors_integration_test.go (same package).
package approval_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// ---------------------------------------------------------------------------
// Test-local template-version port fakes.
//
// This external package must never import templates/infrastructure, so the
// approval-owned ports (application.TemplateVersionReader, the submit-lock +
// completion writers) are satisfied here with raw in-tx SQL — the same
// technique the approval/application package's own integration fakes use.
// ---------------------------------------------------------------------------

type previewTemplateVersionReader struct{}

func (previewTemplateVersionReader) LoadTemplateVersionStatus(ctx context.Context, tx db.Tx, tenantID, versionID string) (string, bool, error) {
	var status sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT status FROM templates_template_version WHERE id = $1 AND tenant_id = $2`,
		versionID, tenantID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !status.Valid) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return status.String, true, nil
}

func (previewTemplateVersionReader) LoadTemplateVersionContentHash(ctx context.Context, tx db.Tx, tenantID, versionID string) (string, bool, error) {
	var hash sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT content_hash FROM templates_template_version WHERE id = $1 AND tenant_id = $2`,
		versionID, tenantID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && (!hash.Valid || hash.String == "")) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return hash.String, true, nil
}

// LoadTemplateDocTypeCode mirrors the production adapter: post-ADR-0086 the
// template ROUTE is keyed by the template's doc_type_code (= profile code),
// so this is what both preview and submit resolve the route by. Fails closed.
func (previewTemplateVersionReader) LoadTemplateDocTypeCode(ctx context.Context, tx db.Tx, tenantID, templateID string) (string, error) {
	var code sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT doc_type_code FROM templates_template WHERE id = $1 AND tenant_id = $2`,
		templateID, tenantID,
	).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("preview parity fake: template not found")
	}
	if err != nil {
		return "", err
	}
	if !code.Valid || code.String == "" {
		return "", errors.New("preview parity fake: template has no doc_type_code")
	}
	return code.String, nil
}

func (previewTemplateVersionReader) LoadTemplateInboxMeta(ctx context.Context, tx db.Tx, tenantID string, versionIDs []string) (map[string]domain.TemplateInboxMeta, error) {
	out := make(map[string]domain.TemplateInboxMeta)
	for _, versionID := range versionIDs {
		var templateID, name string
		err := tx.QueryRowContext(ctx,
			`SELECT t.id, t.name
			   FROM templates_template_version tv
			   JOIN templates_template t
			     ON t.id = tv.template_id
			    AND t.tenant_id = tv.tenant_id
			  WHERE tv.tenant_id = $1
			    AND tv.id = $2`,
			tenantID, versionID,
		).Scan(&templateID, &name)
		if errors.Is(err, sql.ErrNoRows) {
			continue // absent / cross-tenant rows are omitted, mirroring the port contract
		}
		if err != nil {
			return nil, err
		}
		out[versionID] = domain.TemplateInboxMeta{TemplateID: templateID, Title: name}
	}
	return out, nil
}

// previewTemplateVersionWriter satisfies BOTH the submit-lock port
// (TemplateVersionSubmitWriter, MarkTemplateVersionUnderReview) and the
// completion port (TemplateCompletionWriter) so a single value wires through
// Services.WithTemplateCompletionWriter (which type-asserts the submit-lock
// leg). Only the submit-lock UPDATE is exercised here.
type previewTemplateVersionWriter struct{}

func (previewTemplateVersionWriter) MarkTemplateVersionUnderReview(ctx context.Context, tx db.Tx, tenantID, versionID string) error {
	// Mirrors the production adapter's CAS (content_hash guard in the WHERE)
	// plus its in-tx disambiguation on 0 rows — the port contract.
	res, err := tx.ExecContext(ctx,
		`UPDATE templates_template_version
		   SET status = 'under_review', submitted_at = now(), lock_version = lock_version + 1
		 WHERE id = $1 AND tenant_id = $2 AND status = 'draft'
		   AND content_hash IS NOT NULL AND content_hash <> ''`,
		versionID, tenantID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		var status, contentHash sql.NullString
		if err := tx.QueryRowContext(ctx,
			`SELECT status, content_hash FROM templates_template_version WHERE id = $1 AND tenant_id = $2`,
			versionID, tenantID,
		).Scan(&status, &contentHash); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("preview parity fake: version not draft")
			}
			return err
		}
		if status.String == "draft" && (!contentHash.Valid || contentHash.String == "") {
			return domain.ErrTemplateVersionNoContent
		}
		return errors.New("preview parity fake: version not draft")
	}
	return nil
}

func (previewTemplateVersionWriter) MarkTemplateVersionApproved(context.Context, db.Tx, string, string, string) error {
	return nil
}
func (previewTemplateVersionWriter) MarkTemplateVersionRejected(context.Context, db.Tx, string, string) error {
	return nil
}

// seedPreviewTemplateVersion inserts a templates_template + draft
// templates_template_version pair for the tenant/actor, returning both ids.
// docTypeCode must be a real profile code (ADR 0086:
// templates_template_doc_type_code_required_check forbids ”) and is the key
// the template route is looked up by.
func seedPreviewTemplateVersion(t *testing.T, dbc *sql.DB, tenantID, actorID, docTypeCode string) (templateID, versionID string) {
	t.Helper()
	ctx := context.Background()
	key := "preview-parity-" + uuid.NewString()
	docxKey := "templates/preview-parity/" + key + "/body.docx"
	contentHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed template version: begin: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		t.Fatalf("seed template version: identity: %v", err)
	}
	testdb.SetCapsOnTx(t, tx, `[{"cap":"template.create"}]`)
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template
		   (id, tenant_id, doc_type_code, key, name, description, latest_version, created_by, created_at)
		 VALUES (gen_random_uuid(), $1::uuid, $2, $3, 'Preview Parity Template', '', 1, $4::text, now())
		 RETURNING id::text`,
		tenantID, docTypeCode, key, actorID,
	).Scan(&templateID); err != nil {
		t.Fatalf("seed template version: insert template: %v", err)
	}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template_version
		   (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
		    metadata_schema, placeholder_schema, author_id)
		 VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 1, 'draft', $3, $4, '{}'::jsonb, '[]'::jsonb, $5::text)
		 RETURNING id::text`,
		tenantID, templateID, docxKey, contentHash, actorID,
	).Scan(&versionID); err != nil {
		t.Fatalf("seed template version: insert version: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seed template version: commit: %v", err)
	}
	return templateID, versionID
}

// previewParityStages is a two-stage route whose stage 2 is a submit_choice
// selector carrying role + area_code — the exact shape a chosen_actors picker
// inspects. Stage 1 (role_in_fixed_area at the area-blind "tenant" sentinel)
// contributes a pool from the route alone; stage 2's pool comes solely from
// the caller's validated chosen_actors.
func previewParityStages() []domain.Stage {
	return []domain.Stage{
		{
			Order:              1,
			Name:               "quality",
			RequiredCapability: "document.signoff",
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorRoleInFixedArea, Role: "approver", AreaCode: "tenant"},
			},
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
		},
		{
			Order:              2,
			Name:               "approval",
			RequiredCapability: "document.signoff",
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorSubmitChoice, Role: "author", AreaCode: "quality"},
			},
			Quorum:             domain.QuorumAny1Of,
			OnEligibilityDrift: domain.DriftReduceQuorum,
		},
	}
}

// TestPreviewRoute_TemplateSubject_ResolutionParity_RealDB is the guard.
func TestPreviewRoute_TemplateSubject_ResolutionParity_RealDB(t *testing.T) {
	ctx := context.Background()
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedPreviewTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	services := application.NewServices(repo, application.NewSQLEmitter(), application.RealClock{}, nil).
		WithTemplateVersionReader(previewTemplateVersionReader{}).
		WithTemplateCompletionWriter(previewTemplateVersionWriter{})
	runner := db.NewTxRunner(dbc)

	// --- Seed the active TEMPLATE route with a submit_choice stage 2 directly
	// through RouteAdmin.Create: post-ADR-0086 it persists
	// (subject_kind='template', subject_key=profile_code), so no raw-UPDATE
	// subject flip is needed (or permitted by
	// approval_routes_template_subject_key_check). ---
	rctx := platformtenant.WithActorID(platformtenant.WithTenantID(ctx, tnt.ID), actor.ID)
	createOut, err := services.RouteAdmin.Create(rctx, runner, application.CreateRouteInput{
		TenantID:    tnt.ID,
		ProfileCode: tax.ProfileCode,
		SubjectKind: string(domain.SubjectKindTemplate),
		Name:        "Preview Parity Route",
		ActorUserID: actor.ID,
		Stages:      previewParityStages(),
	})
	if err != nil {
		t.Fatalf("seed route: Create: %v", err)
	}

	// Eligibility pools: stage 1 approver (area-blind "tenant" matches any real
	// area), stage 2 submit_choice chosen author in area "quality".
	seedProcessArea(t, dbc, tnt.ID, "quality", "Quality")
	reviewer := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	chosen := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantActiveArea(t, dbc, tnt.ID, reviewer.ID, "quality", "approver")
	grantActiveArea(t, dbc, tnt.ID, chosen.ID, "quality", "author")

	// --- (A) The preview under test. ---
	preview, err := services.TemplateSubmit.PreviewRoute(rctx, runner, tnt.ID, templateID)
	if err != nil {
		t.Fatalf("PreviewRoute: %v", err)
	}
	if preview.RouteID != createOut.RouteID {
		t.Fatalf("preview RouteID = %q, want %q", preview.RouteID, createOut.RouteID)
	}

	// The submit_choice selector must survive with role + area_code so a picker
	// can enumerate the constraint pool.
	if len(preview.Stages) != 2 {
		t.Fatalf("preview stages = %d, want 2", len(preview.Stages))
	}
	sc := preview.Stages[1].Selectors[0]
	if sc.Kind != domain.SelectorSubmitChoice || sc.Role != "author" || sc.AreaCode != "quality" {
		t.Fatalf("stage 2 selector = %+v, want submit_choice{role author, area quality}", sc)
	}

	// --- (B) Independent oracle: a real end-to-end submit for the SAME subject.
	// Its resolved instance.RouteID and the route it loaded are the ground
	// truth preview must match. ---
	res, err := services.TemplateSubmit.SubmitTemplateVersionForReview(rctx, runner, application.TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
		ChosenActors: []application.StageChosenActors{
			{StageOrder: 2, UserIDs: []string{chosen.ID}},
		},
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}

	var submitRoute domain.Route
	var instRouteID string
	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		inst, e := repo.LoadInstance(ctx, tx, tnt.ID, res.InstanceID)
		if e != nil {
			return e
		}
		instRouteID = inst.RouteID
		submitRoute, e = repo.LoadRoute(ctx, tx, tnt.ID, inst.RouteID)
		return e
	}); err != nil {
		t.Fatalf("load submitted instance + route: %v", err)
	}

	// Parity 1: preview and the real submit resolved the SAME route id from the
	// SAME subject (the template's doc_type_code, not the template id and not
	// the version id).
	if preview.RouteID != instRouteID {
		t.Fatalf("preview RouteID %q != submit instance RouteID %q — resolution parity broken", preview.RouteID, instRouteID)
	}

	// Parity 2: preview's stages/selectors deep-equal the route the submit path
	// actually loaded.
	if len(preview.Stages) != len(submitRoute.Stages) {
		t.Fatalf("preview has %d stages, submit route has %d", len(preview.Stages), len(submitRoute.Stages))
	}
	for i, ps := range preview.Stages {
		os := submitRoute.Stages[i]
		if ps.StageOrder != os.Order || ps.Label != os.Name {
			t.Fatalf("stage %d = {order %d, label %q}, want {order %d, label %q}", i, ps.StageOrder, ps.Label, os.Order, os.Name)
		}
		if len(ps.Selectors) != len(os.Selectors) {
			t.Fatalf("stage %d selector count = %d, want %d", i, len(ps.Selectors), len(os.Selectors))
		}
		for j, psel := range ps.Selectors {
			if psel != os.Selectors[j] {
				t.Fatalf("stage %d selector %d = %+v, want %+v", i, j, psel, os.Selectors[j])
			}
		}
	}
}
