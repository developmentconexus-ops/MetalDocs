//go:build integration

// Unit 4.2 (template-inbox) — real-DB proof that ListWorklist is
// subject-generic: a template kernel instance (subject_kind='template')
// surfaces in the worklist alongside document instances, resolving its
// display metadata {templateId, name} through the approval-owned
// TemplateVersionReader.LoadTemplateInboxMeta port (never a raw join into
// templates_template*), while the existing document path (LEFT JOIN
// documents) stays byte-for-byte unchanged and a cross-tenant template
// version id never leaks its title/ref into another tenant's row.
package application

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// fakeTemplateInboxMetaReader is a test-local double for
// TemplateVersionReader.LoadTemplateInboxMeta, duplicating (rather than
// importing) the production templates/infrastructure adapter's tenant-scoped
// set-based join — this test package (approval/application) must never
// import templates infrastructure. Only LoadTemplateInboxMeta is implemented;
// the other two TemplateVersionReader methods are unused by ListWorklist.
type fakeTemplateInboxMetaReader struct{}

func (fakeTemplateInboxMetaReader) LoadTemplateVersionStatus(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	return "", false, nil
}

func (fakeTemplateInboxMetaReader) LoadTemplateVersionContentHash(ctx context.Context, tx db.Tx, tenantID, templateVersionID string) (string, bool, error) {
	return "", false, nil
}

func (fakeTemplateInboxMetaReader) LoadTemplateDocTypeCode(ctx context.Context, tx db.Tx, tenantID, templateID string) (string, error) {
	return "", nil
}

func (fakeTemplateInboxMetaReader) LoadTemplateInboxMeta(ctx context.Context, tx db.Tx, tenantID string, versionIDs []string) (map[string]domain.TemplateInboxMeta, error) {
	out := make(map[string]domain.TemplateInboxMeta)
	if len(versionIDs) == 0 {
		return out, nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT tv.id, t.id, t.name
		  FROM templates_template_version tv
		  JOIN templates_template t
		    ON t.id = tv.template_id
		   AND t.tenant_id = tv.tenant_id
		 WHERE tv.tenant_id = $1
		   AND tv.id = ANY($2)`,
		tenantID, pq.Array(versionIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var versionID, templateID, name string
		if err := rows.Scan(&versionID, &templateID, &name); err != nil {
			return nil, err
		}
		out[versionID] = domain.TemplateInboxMeta{TemplateID: templateID, Title: name}
	}
	return out, rows.Err()
}

// seedActiveStageWithEligibleActor seeds an in-progress instance's active
// stage with eligible_actor_ids containing exactly actorID, so ListWorklist's
// direct eligibility-pool predicate (not the Oversee bypass) is what surfaces
// the row.
func seedActiveStageWithEligibleActor(t *testing.T, database *sql.DB, instanceID, actorID string) {
	t.Helper()
	eligible := `["` + actorID + `"]`
	testdb.SeedWithCaps(t, database, `[{"cap":"document.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `
			INSERT INTO public.approval_stage_instances
			  (approval_instance_id, stage_order, name_snapshot, required_role_snapshot,
			   area_code_snapshot, quorum_snapshot, on_eligibility_drift_snapshot,
			   eligible_actor_ids, status, stage_kind, required_capability_snapshot)
			VALUES ($1::uuid, 1, 'Stage 1', 'reviewer', 'quality', 'any_1_of', 'keep_snapshot',
			        $2::jsonb, 'active', 'approval', 'document.review')`,
			instanceID, eligible,
		)
		return err
	})
}

// TestListWorklist_DocumentInstance_StillReturnsTitleAndControlledDocID_RealDB
// pins that Unit 4.2 does not change the document worklist path: the LEFT
// JOIN documents-sourced title/ref/CD-id stay exactly as before.
func TestListWorklist_DocumentInstance_StillReturnsTitleAndControlledDocID_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, dbc)
	actor := testdb.NewUser(t, dbc, testdb.WithTenant(ten.ID))

	cd := testdb.NewControlledDoc(t, dbc, testdb.WithTenant(ten.ID), testdb.WithCode("POP-QUA-0148"))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(ten.ID), testdb.WithControlledDoc(cd), testdb.WithName("Quality Manual"), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, dbc, testdb.WithDocument(doc), testdb.WithStatus("in_progress"))
	seedActiveStageWithEligibleActor(t, dbc, inst.ID, actor.ID)

	svc := newReadServiceForIntegration(t, dbc)
	runner := db.NewTxRunner(dbc)

	items, _, err := svc.ListWorklist(ctxWithIdentity(ten.ID, actor.ID), runner, ten.ID, actor.ID, "", InboxFilter{}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist: %v", err)
	}

	var found *InboxView
	for i := range items {
		if items[i].InstanceID == inst.ID {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("ListWorklist items = %#v, want to include document instance %q", items, inst.ID)
	}
	if found.SubjectKind != "document" {
		t.Errorf("SubjectKind = %q, want %q", found.SubjectKind, "document")
	}
	if found.SubjectTitle != "Quality Manual" {
		t.Errorf("SubjectTitle = %q, want %q", found.SubjectTitle, "Quality Manual")
	}
	if found.SubjectRef != doc.ID {
		t.Errorf("SubjectRef = %q, want documentId %q", found.SubjectRef, doc.ID)
	}
	if found.ControlledDocumentID != doc.ControlledDocumentID {
		t.Errorf("ControlledDocumentID = %q, want %q", found.ControlledDocumentID, doc.ControlledDocumentID)
	}
	// F-QA4-8: the worklist must carry the CANONICAL human code
	// (controlled_documents.code) so the inbox stops rendering the uuid.
	if found.ControlledDocumentCode != cd.Code {
		t.Errorf("ControlledDocumentCode = %q, want %q", found.ControlledDocumentCode, cd.Code)
	}
}

// TestListWorklist_TemplateInstance_AppearsWithSubjectTitleAndRef_RealDB is
// the RED->GREEN target: a template kernel instance (subject_kind='template')
// appears in ListWorklist for an eligible actor, resolving its title/ref
// through TemplateVersionReader.LoadTemplateInboxMeta.
func TestListWorklist_TemplateInstance_AppearsWithSubjectTitleAndRef_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	ten := testdb.NewTenant(t, dbc)
	admin := testdb.NewUser(t, dbc, testdb.WithTenant(ten.ID), testdb.WithRole("system_admin"))

	tax := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(ten.ID), testdb.WithGovernanceClass("controlado"))
	templateID, versionID := seedTemplateVersion(t, dbc, ten.ID, admin.ID, tax.ProfileCode, "draft")
	seedTemplateRoute(t, dbc, ten.ID, admin.ID, tax.ProfileCode)

	reviewer := testdb.NewUser(t, dbc, testdb.WithTenant(ten.ID))
	manager := testdb.NewUser(t, dbc, testdb.WithTenant(ten.ID))
	grantAreaBlindRole(t, dbc, ten.ID, reviewer.ID, "approver")
	grantAreaBlindRole(t, dbc, ten.ID, manager.ID, "author")

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, nil)
	submitSvc := NewTemplateSubmitService(repo, NewSQLEmitter(), RealClock{}, fakeTemplateVersionReader{})
	submitSvc.versionWriter = fakeTemplateVersionSubmitWriter{}
	runner := db.NewTxRunner(dbc)

	res, err := submitSvc.SubmitTemplateVersionForReview(newTemplateSubmitCtx(ten.ID, admin.ID), runner, TemplateSubmitRequest{
		TenantID:          ten.ID,
		TemplateID:        templateID,
		TemplateVersionID: versionID,
		SubmittedBy:       admin.ID,
		IdempotencyKey:    "idem-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("SubmitTemplateVersionForReview: %v", err)
	}

	readSvc := newReadService(repo, testdb.NewCDFieldReader(t), fakeTemplateInboxMetaReader{})

	items, _, err := readSvc.ListWorklist(ctxWithIdentity(ten.ID, reviewer.ID), runner, ten.ID, reviewer.ID, "", InboxFilter{}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist: %v", err)
	}

	var found *InboxView
	for i := range items {
		if items[i].InstanceID == res.InstanceID {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("ListWorklist items = %#v, want to include template instance %q", items, res.InstanceID)
	}
	if found.SubjectKind != "template" {
		t.Errorf("SubjectKind = %q, want %q", found.SubjectKind, "template")
	}
	if found.SubjectTitle != "Test Template" {
		t.Errorf("SubjectTitle = %q, want %q (seedTemplateVersion's fixed template name)", found.SubjectTitle, "Test Template")
	}
	if found.SubjectRef != templateID {
		t.Errorf("SubjectRef = %q, want templateId %q (not the version id %q)", found.SubjectRef, templateID, versionID)
	}
	if found.ControlledDocumentID != "" {
		t.Errorf("ControlledDocumentID = %q, want empty for a template instance", found.ControlledDocumentID)
	}
	// F-QA4-8: a template subject has no controlled document, hence no code —
	// empty here, explicit null on the wire (never a substitute value).
	if found.ControlledDocumentCode != "" {
		t.Errorf("ControlledDocumentCode = %q, want empty for a template instance", found.ControlledDocumentCode)
	}
}

// TestListWorklist_TemplateInstance_CrossTenantVersion_NoLeak_RealDB proves
// the reader's tenant scoping: an approval_instances row whose subject_key
// happens to reference ANOTHER tenant's template version never resolves that
// other tenant's title/ref into the row — no leak, no fallback (the row's
// SubjectTitle/SubjectRef simply stay empty).
func TestListWorklist_TemplateInstance_CrossTenantVersion_NoLeak_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	tenA := testdb.NewTenant(t, dbc)
	tenB := testdb.NewTenant(t, dbc)
	actorA := testdb.NewUser(t, dbc, testdb.WithTenant(tenA.ID))
	adminB := testdb.NewUser(t, dbc, testdb.WithTenant(tenB.ID), testdb.WithRole("system_admin"))

	// A real template version that exists ONLY under tenant B.
	taxB := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tenB.ID), testdb.WithGovernanceClass("controlado"))
	_, versionB := seedTemplateVersion(t, dbc, tenB.ID, adminB.ID, taxB.ProfileCode, "draft")

	// A tenant-A instance whose subject_key is (illegitimately, for this test)
	// tenant B's version id — seeded by creating a normal document instance
	// under tenant A, then flipping it to a template subject pointing at the
	// cross-tenant version id via raw UPDATE (the INSTANCE-side subject is not
	// constrained by ADR 0086, which re-keys ROUTES only; the INSERT-time
	// tripwire fires only on INSERT, so the flip needs no capability assertion).
	docA := testdb.NewDocument(t, dbc, testdb.WithTenant(tenA.ID), testdb.WithStatus("approved"))
	instA := testdb.NewApprovalInstance(t, dbc, testdb.WithDocument(docA), testdb.WithStatus("in_progress"))
	if _, err := dbc.ExecContext(context.Background(), `
		UPDATE public.approval_instances
		   SET subject_kind = 'template', subject_key = $2, document_id = NULL
		 WHERE id = $1::uuid`,
		instA.ID, versionB,
	); err != nil {
		t.Fatalf("flip instance to cross-tenant template subject: %v", err)
	}
	seedActiveStageWithEligibleActor(t, dbc, instA.ID, actorA.ID)

	readSvc := newReadService(approvalrepo.NewPostgresApprovalRepository(dbc, nil), testdb.NewCDFieldReader(t), fakeTemplateInboxMetaReader{})
	runner := db.NewTxRunner(dbc)

	items, _, err := readSvc.ListWorklist(ctxWithIdentity(tenA.ID, actorA.ID), runner, tenA.ID, actorA.ID, "", InboxFilter{}, 25, 0)
	if err != nil {
		t.Fatalf("ListWorklist: %v", err)
	}

	var found *InboxView
	for i := range items {
		if items[i].InstanceID == instA.ID {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("ListWorklist items = %#v, want to include instance %q", items, instA.ID)
	}
	if found.SubjectKind != "template" {
		t.Errorf("SubjectKind = %q, want %q", found.SubjectKind, "template")
	}
	if found.SubjectTitle != "" {
		t.Errorf("SubjectTitle = %q, want empty (no leak of tenant B's template name into tenant A)", found.SubjectTitle)
	}
	if found.SubjectRef != "" {
		t.Errorf("SubjectRef = %q, want empty (no leak of tenant B's template id into tenant A)", found.SubjectRef)
	}
	if found.ControlledDocumentID != "" {
		t.Errorf("ControlledDocumentID = %q, want empty (document_id was nulled on the flip)", found.ControlledDocumentID)
	}
}
