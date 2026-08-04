//go:build integration
// +build integration

package application

// P3.S2b-3b-ii (M3 kernel extraction) — real-DB proof of the thin,
// subject-generic template submit path. TemplateSubmitService is a SUBSET of
// SubmitService.SubmitRevisionForReview: it reuses the same kernel
// stage-seeding primitives (ResolveEligibleActors -> StageInstance ->
// InsertStageInstances) against a (subject_kind='template') route/instance
// pair, and omits every document-only concern (content hash, freeze,
// revision-title, documents status transition).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
)

// fakeTemplateVersionReader is a test-local double for the approval-owned
// TemplateVersionReader port, backed by a raw in-tx read against
// templates_template_version. It intentionally duplicates (rather than
// imports) the production templates/infrastructure adapter's query — this
// test package (approval/application) must never import templates
// infrastructure, so the fixture reproduces the same minimal SQL locally.
type fakeTemplateVersionReader struct{}

func (fakeTemplateVersionReader) LoadTemplateVersionStatus(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	var status sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT status FROM templates_template_version
		 WHERE id = $1 AND tenant_id = $2`,
		templateVersionID, tenantID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !status.Valid {
		return "", false, nil
	}
	return status.String, true, nil
}

func (fakeTemplateVersionReader) LoadTemplateVersionContentHash(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	var hash sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT content_hash FROM templates_template_version
		 WHERE id = $1 AND tenant_id = $2`,
		templateVersionID, tenantID,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !hash.Valid || hash.String == "" {
		return "", false, nil
	}
	return hash.String, true, nil
}

// LoadTemplateDocTypeCode (ADR 0086) mirrors the production adapter: the
// template's doc_type_code IS the template ROUTE subject key, so submit
// resolves the route by profile, not by template instance. Fails closed on a
// missing/blank code exactly like the real reader.
func (fakeTemplateVersionReader) LoadTemplateDocTypeCode(ctx context.Context, tx db.Tx, tenantID, templateID string) (string, error) {
	var code sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT doc_type_code FROM templates_template
		 WHERE id = $1 AND tenant_id = $2`,
		templateID, tenantID,
	).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("fake reader: template not found")
	}
	if err != nil {
		return "", err
	}
	if !code.Valid || code.String == "" {
		return "", errors.New("fake reader: template has no doc_type_code")
	}
	return code.String, nil
}

// LoadTemplateInboxMeta (Unit 4.2) mirrors the production adapter's batch
// resolve of {templateId, name} per version id, scoped to tenantID, read inside
// tx. Per-id loop (rather than ANY($)) to avoid a pq dependency in this test
// package; a version id with no matching row is simply absent from the map.
func (fakeTemplateVersionReader) LoadTemplateInboxMeta(ctx context.Context, tx db.Tx, tenantID string, versionIDs []string) (map[string]domain.TemplateInboxMeta, error) {
	out := make(map[string]domain.TemplateInboxMeta)
	for _, versionID := range versionIDs {
		var templateID, name string
		err := tx.QueryRowContext(ctx, `
			SELECT t.id, t.name
			  FROM templates_template_version tv
			  JOIN templates_template t
			    ON t.id = tv.template_id
			   AND t.tenant_id = tv.tenant_id
			 WHERE tv.tenant_id = $1
			   AND tv.id = $2`,
			tenantID, versionID,
		).Scan(&templateID, &name)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out[versionID] = domain.TemplateInboxMeta{TemplateID: templateID, Title: name}
	}
	return out, nil
}

// fakeTemplateVersionSubmitWriter is a test-local double for the approval-owned
// TemplateVersionSubmitWriter port (M3 P3.S2b-3b-iii-b submit-lock). Like the
// reader fake it duplicates (rather than imports) the production
// templates/infrastructure adapter's UPDATE — this test package must never
// import templates infrastructure. Mirrors ApprovalCompletionWriter's
// draft -> under_review CAS.
type fakeTemplateVersionSubmitWriter struct{}

func (fakeTemplateVersionSubmitWriter) MarkTemplateVersionUnderReview(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE templates_template_version
		   SET status = 'under_review', submitted_at = now(), lock_version = lock_version + 1
		 WHERE id = $1 AND tenant_id = $2 AND status = 'draft'
		   AND content_hash IS NOT NULL AND content_hash <> ''`,
		templateVersionID, tenantID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Same in-tx disambiguation the production adapter performs: a 0-row
		// CAS on a still-draft row with no committed hash is the content case,
		// not a stale-status conflict.
		var status, contentHash sql.NullString
		if err := tx.QueryRowContext(ctx, `
			SELECT status, content_hash FROM templates_template_version
			 WHERE id = $1 AND tenant_id = $2`,
			templateVersionID, tenantID,
		).Scan(&status, &contentHash); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("fake submit-lock: version not draft")
			}
			return err
		}
		if status.String == "draft" && (!contentHash.Valid || contentHash.String == "") {
			return domain.ErrTemplateVersionNoContent
		}
		return errors.New("fake submit-lock: version not draft")
	}
	return nil
}

// seedTemplateVersion inserts a templates_template + templates_template_version
// pair, mirroring testdb.InsertDraftDocument's canonical-family fixture but
// parameterized on status (draft vs published) since that unexported helper
// always seeds 'published'. content_hash is populated regardless of status —
// chk_template_version_content_hash_non_draft only requires it for non-draft
// rows, but a fixed 64-hex value is harmless for draft too.
// docTypeCode must be a real profile code: post-ADR-0086
// templates_template_doc_type_code_required_check rejects '' outright, and it
// is what the template route is keyed by.
func seedTemplateVersion(t *testing.T, dbc *sql.DB, tenantID, actorID, docTypeCode, status string) (templateID, versionID string) {
	t.Helper()
	ctx := context.Background()
	key := "tsvc-tmpl-" + uuid.NewString()
	docxKey := "templates/test-seed/" + key + "/body.docx"
	contentHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"[:64]

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedTemplateVersion: begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		t.Fatalf("seedTemplateVersion: seed identity: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, `[{"cap":"template.create"}]`,
	); err != nil {
		t.Fatalf("seedTemplateVersion: assert caps: %v", err)
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template
		   (id, tenant_id, doc_type_code, key, name, description, latest_version, created_by, created_at)
		 VALUES (gen_random_uuid(), $1::uuid, $2, $3, 'Test Template', '', 1, $4::text, now())
		 RETURNING id::text`,
		tenantID, docTypeCode, key, actorID,
	).Scan(&templateID); err != nil {
		t.Fatalf("seedTemplateVersion: insert template: %v", err)
	}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template_version
		   (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
		    metadata_schema, placeholder_schema, author_id)
		 VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 1, $3, $4, $5, '{}'::jsonb, '[]'::jsonb, $6::text)
		 RETURNING id::text`,
		tenantID, templateID, status, docxKey, contentHash, actorID,
	).Scan(&versionID); err != nil {
		t.Fatalf("seedTemplateVersion: insert template version: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedTemplateVersion: commit: %v", err)
	}
	return templateID, versionID
}

// seedTemplateVersionNoContent mirrors seedTemplateVersion but leaves
// content_hash empty — the exact state the legacy docx-upload-url direct-PUT
// path produces (only autosave/commit sets a hash). This is legal for a draft
// row (chk_template_version_content_hash_non_draft demands the 64-char hash
// only for non-draft statuses), which is precisely why the constraint used to
// fire at submit time, mid-tx, as a raw 23514 (F-E4-1).
func seedTemplateVersionNoContent(t *testing.T, dbc *sql.DB, tenantID, actorID, docTypeCode string) (templateID, versionID string) {
	t.Helper()
	ctx := context.Background()
	key := "tsvc-tmpl-" + uuid.NewString()
	docxKey := "templates/test-seed/" + key + "/body.docx"

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedTemplateVersionNoContent: begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		t.Fatalf("seedTemplateVersionNoContent: seed identity: %v", err)
	}
	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, `[{"cap":"template.create"}]`,
	); err != nil {
		t.Fatalf("seedTemplateVersionNoContent: assert caps: %v", err)
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template
		   (id, tenant_id, doc_type_code, key, name, description, latest_version, created_by, created_at)
		 VALUES (gen_random_uuid(), $1::uuid, $2, $3, 'Test Template', '', 1, $4::text, now())
		 RETURNING id::text`,
		tenantID, docTypeCode, key, actorID,
	).Scan(&templateID); err != nil {
		t.Fatalf("seedTemplateVersionNoContent: insert template: %v", err)
	}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO public.templates_template_version
		   (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
		    metadata_schema, placeholder_schema, author_id)
		 VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 1, 'draft', $3, '', '{}'::jsonb, '[]'::jsonb, $4::text)
		 RETURNING id::text`,
		tenantID, templateID, docxKey, actorID,
	).Scan(&versionID); err != nil {
		t.Fatalf("seedTemplateVersionNoContent: insert template version: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedTemplateVersionNoContent: commit: %v", err)
	}
	return templateID, versionID
}

// seedTemplateRoute creates an ACTIVE template approval route with two stages
// (AreaCode="tenant", the area-blind sentinel — P3.S2b-3b-i) through the real
// RouteAdminService.Create entry point. Post-ADR-0086 no raw-UPDATE subject
// flip is needed (nor legal): Create persists
// (subject_kind='template', subject_key=profile_code) directly, which is what
// approval_routes_template_subject_key_check now demands.
// templateRouteStages mirrors validRouteStages()'s shape (two stages, each a
// single role_in_fixed_area selector at AreaCode="tenant" area-blind sentinel —
// P3.S2b-3b-i) but uses role values from the real user_process_areas_role_check
// enum ("approver", "author") rather than validRouteStages()'s
// route-admin-unit-test-only role labels ("qa_reviewer"/"qa_manager"), which are
// never persisted against a real DB in that mocked-DB test file and so are not
// constrained to the enum. This real-DB integration test must grant roles the
// DB will accept. (Selectors are the sole actor-pool source of truth since
// migration 0305 dropped the flat required_role/area_code columns.)
func templateRouteStages() []domain.Stage {
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
				{Kind: domain.SelectorRoleInFixedArea, Role: "author", AreaCode: "tenant"},
			},
			Quorum:             domain.QuorumAllOf,
			OnEligibilityDrift: domain.DriftFailStage,
		},
	}
}

func seedTemplateRoute(t *testing.T, dbc *sql.DB, tenantID, actorID, profileCode string) approvalrepo.Route {
	t.Helper()
	ctx := context.Background()

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := &RouteAdminService{repo: repo, emitter: NewSQLEmitter(), clock: RealClock{}}
	runner := db.NewTxRunner(dbc)

	rctx := platformtenant.WithTenantID(context.Background(), tenantID)
	rctx = platformtenant.WithActorID(rctx, actorID)

	out, err := svc.Create(rctx, runner, CreateRouteInput{
		TenantID:    tenantID,
		ProfileCode: profileCode,
		SubjectKind: string(domain.SubjectKindTemplate),
		Name:        "Template Route",
		ActorUserID: actorID,
		Stages:      templateRouteStages(),
	})
	if err != nil {
		t.Fatalf("seed template route: Create: %v", err)
	}

	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed template route: begin: %v", err)
	}
	defer tx.Rollback()
	route, err := repo.LoadRoute(ctx, tx, tenantID, out.RouteID)
	if err != nil {
		t.Fatalf("seed template route: load: %v", err)
	}
	return approvalrepo.Route{ID: route.ID, Version: route.Version}
}

// grantAreaBlindRole grants userID the given role in an arbitrary real area
// (area-blind ResolveEligibleActors treats AreaCode="tenant" as "role held
// anywhere for this tenant" — P3.S2b-3b-i).
// seedFixedArea creates a real process area row (document_process_areas),
// required by user_process_areas' FK before any role grant into that area —
// mirrors infrastructure/eligible_actors_view_parity_integration_test.go's
// helper of the same name, duplicated here (not imported) to keep this
// application-package test self-contained.
func seedFixedArea(t *testing.T, dbc *sql.DB, tenantID, code, name string) {
	t.Helper()
	testdb.SeedWithCaps(t, dbc, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1, $2::uuid, $3)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			code, tenantID, name,
		)
		return err
	})
}

func grantAreaBlindRole(t *testing.T, dbc *sql.DB, tenantID, userID, role string) {
	t.Helper()
	area := "quality"
	seedFixedArea(t, dbc, tenantID, area, "Quality")
	testdb.SeedWithCaps(t, dbc, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1, $2::uuid, $3, $4, now() - interval '1 hour')
			 ON CONFLICT DO NOTHING`,
			userID, tenantID, area, role,
		)
		return err
	})
}

func newTemplateSubmitCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	return platformtenant.WithActorID(ctx, actorID)
}

// TestSubmitTemplateVersionForReview_Success_RealDB is the RED->GREEN target:
// submit creates an in_progress instance with Subject=(template, versionID),
// route resolved by the template's doc_type_code (ADR 0086 — not by template
// id and not by version id), N stage instances
// seeded with the first active, and an approval_submitted governance event
// emitted.
func TestSubmitTemplateVersionForReview_Success_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "draft")
	route := seedTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	// Eligibility pools for templateRouteStages(): stage 1 role "approver",
	// stage 2 role "author".
	reviewer := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	manager := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID))
	grantAreaBlindRole(t, dbc, tnt.ID, reviewer.ID, "approver")
	grantAreaBlindRole(t, dbc, tnt.ID, manager.ID, "author")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{})
	svc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	res, err := svc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}
	if res.InstanceID == "" {
		t.Fatal("SubmitTemplateVersionForReview: empty InstanceID")
	}

	ctx := context.Background()
	tx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin read tx: %v", err)
	}
	defer tx.Rollback()

	inst, err := repo.LoadInstance(ctx, tx, tnt.ID, res.InstanceID)
	if err != nil {
		t.Fatalf("LoadInstance: %v", err)
	}
	if inst.Subject.Kind != domain.SubjectKindTemplate || inst.Subject.Key != versionID {
		t.Fatalf("Subject = %+v, want {template %s}", inst.Subject, versionID)
	}
	if inst.RouteID != route.ID {
		t.Fatalf("RouteID = %q, want %q (resolved by doc_type_code)", inst.RouteID, route.ID)
	}
	if inst.Status != domain.InstanceInProgress {
		t.Fatalf("Status = %q, want in_progress", inst.Status)
	}
	if inst.DocumentID != "" {
		t.Fatalf("DocumentID = %q, want empty for a template instance", inst.DocumentID)
	}
	if len(inst.Stages) != 2 {
		t.Fatalf("len(Stages) = %d, want 2", len(inst.Stages))
	}
	if inst.Stages[0].Status != domain.StageActive {
		t.Fatalf("Stages[0].Status = %q, want active", inst.Stages[0].Status)
	}
	if inst.Stages[1].Status != domain.StagePending {
		t.Fatalf("Stages[1].Status = %q, want pending", inst.Stages[1].Status)
	}

	// Governance event: approval_submitted, resource_type=template,
	// resource_id=versionID.
	var eventType, resourceType, resourceID string
	if err := dbc.QueryRowContext(ctx,
		`SELECT event_type, resource_type, resource_id
		   FROM governance_events
		  WHERE tenant_id = $1::uuid AND resource_id = $2
		  ORDER BY created_at DESC LIMIT 1`,
		tnt.ID, versionID,
	).Scan(&eventType, &resourceType, &resourceID); err != nil {
		t.Fatalf("query governance event: %v", err)
	}
	if eventType != "approval_submitted" || resourceType != "template" || resourceID != versionID {
		t.Fatalf("event = (%q,%q,%q), want (approval_submitted,template,%q)", eventType, resourceType, resourceID, versionID)
	}

	// Submit-lock (M3 P3.S2b-3b-iii-b): the version's own status column must
	// now be under_review, so the templates module's draft-gated edit/upload
	// endpoints reject further author edits for the duration of the approval.
	var versionStatus string
	if err := dbc.QueryRowContext(ctx,
		`SELECT status FROM templates_template_version WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		versionID, tnt.ID,
	).Scan(&versionStatus); err != nil {
		t.Fatalf("query version status: %v", err)
	}
	if versionStatus != "under_review" {
		t.Fatalf("version status = %q, want under_review (submit-lock)", versionStatus)
	}
}

// TestSubmitTemplateVersionForReview_NonDraft_Rejected_RealDB proves the
// precondition guard: a published (non-draft) template version is rejected
// with ErrTemplateVersionNotDraft and no instance row is written.
func TestSubmitTemplateVersionForReview_NonDraft_Rejected_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersion(t, dbc, tnt.ID, actor.ID, tax.ProfileCode, "published")
	seedTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{})
	svc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	_, err := svc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if !errors.Is(err, ErrTemplateVersionNotDraft) {
		t.Fatalf("err = %v, want ErrTemplateVersionNotDraft", err)
	}

	var count int
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM public.approval_instances WHERE tenant_id = $1::uuid AND subject_kind = 'template' AND subject_key = $2`,
		tnt.ID, versionID,
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("instance count = %d, want 0", count)
	}
}

// TestSubmitTemplateVersionForReview_NoContentHash_Rejected_RealDB is the
// F-E4-1 regression pin against the real DB and the REAL templates-infra
// reader/writer adapters: a draft version with an empty content_hash is
// rejected with ErrTemplateVersionNoContent, the version stays draft (the
// submit-lock never ran), and no approval instance is created. Before the fix
// the submit reached MarkTemplateVersionUnderReview, tripped
// chk_template_version_content_hash_non_draft, and surfaced as a raw 500
// carrying the constraint name.
func TestSubmitTemplateVersionForReview_NoContentHash_Rejected_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tnt := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tnt.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersionNoContent(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)
	seedTemplateRoute(t, dbc, tnt.ID, actor.ID, tax.ProfileCode)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	svc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{})
	svc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	_, err := svc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(tnt.ID, actor.ID), runner, TemplateSubmitRequest{
		TenantID:          tnt.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       actor.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if !errors.Is(err, ErrTemplateVersionNoContent) {
		t.Fatalf("err = %v, want ErrTemplateVersionNoContent", err)
	}

	var status string
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT status FROM templates_template_version WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		versionID, tnt.ID,
	).Scan(&status); err != nil {
		t.Fatalf("query version status: %v", err)
	}
	if status != "draft" {
		t.Fatalf("version status = %q, want draft (submit-lock must never have run)", status)
	}

	var count int
	if err := dbc.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM public.approval_instances WHERE tenant_id = $1::uuid AND subject_kind = 'template' AND subject_key = $2`,
		tnt.ID, versionID,
	).Scan(&count); err != nil {
		t.Fatalf("count check: %v", err)
	}
	if count != 0 {
		t.Fatalf("instance count = %d, want 0", count)
	}
}
