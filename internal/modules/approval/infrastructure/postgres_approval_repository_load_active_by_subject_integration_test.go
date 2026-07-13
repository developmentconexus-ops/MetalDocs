//go:build integration
// +build integration

package infrastructure

// P3.S2b-4 review-fix: real-DB coverage for LoadActiveInstanceBySubject
// (added in commit 9918848a, #19 template HTTP entry points). The reviewer
// flagged this method as a fail-closed integrity READ path — tenant-scoped,
// subject-generic (subject_kind/subject_key), LEFT JOIN on documents so a
// template subject's NULL document_id is not dropped — with ZERO real-DB
// coverage: only reached indirectly through fakes in the HTTP tests. This
// file closes that gap with two cases:
//
//  1. Happy path: a template-subject instance (NULL document_id) round-trips
//     through the method, proving the LEFT JOIN guard actually works against
//     Postgres, not just in a fake.
//  2. Cross-tenant isolation: the same subject_key under a different tenant
//     is invisible — proving the tenant_id predicate is load-bearing, not
//     just present in the SQL text.

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// TestLoadActiveInstanceBySubject_TemplateSubject_RoundTrips_RealDB proves the
// happy path: an active (in_progress) template-subject instance — NULL
// document_id, subject_kind='template' — round-trips through
// LoadActiveInstanceBySubject, including its active stage. Before this test,
// nothing exercised the method's LEFT JOIN against a real NULL-document_id
// row; an INNER JOIN regression here would have silently returned
// ErrNoActiveInstance instead of the instance.
func TestLoadActiveInstanceBySubject_TemplateSubject_RoundTrips_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	cd := testdb.NewControlledDoc(t, dbc)
	doc := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cd), testdb.WithStatus("approved"))
	route := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(doc.TenantID))

	templateKey := "tmpl-" + uuid.NewString()
	instID := uuid.NewString()

	// Raw insert of a legal template row (NULL document_id, active status).
	// insertInstanceTx (shared helper, defined in the sibling
	// postgres_approval_repository_subject_integration_test.go) seeds
	// tenant/actor identity plus the document.submit/template.submit caps the
	// approval_instances INSERT tripwire requires.
	tx := insertInstanceTx(t, dbc, doc.TenantID, doc.Owner)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_instances
		   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
		    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		    subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, NULL, $3::uuid, 1, 'in_progress', $4, now(), repeat('a', 64), $5, 'template', $6)`,
		instID, doc.TenantID, route.ID, doc.Owner, "idem-"+uuid.NewString(), templateKey,
	); err != nil {
		tx.Rollback()
		t.Fatalf("raw insert template instance: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	stageTx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin stage tx: %v", err)
	}
	stage := domain.StageInstance{
		ID:                         uuid.NewString(),
		ApprovalInstanceID:         instID,
		StageOrder:                 1,
		NameSnapshot:               "Review",
		RequiredRoleSnapshot:       "reviewer",
		RequiredCapabilitySnapshot: "template.review",
		AreaCodeSnapshot:           "eng",
		QuorumSnapshot:             domain.QuorumAllOf,
		QuorumMSnapshot:            nil,
		OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
		Kind:                       domain.StageKindApproval,
		EligibleActorIDs:           []string{doc.Owner},
		EffectiveDenominator:       nil,
		Status:                     domain.StageActive,
		DueInDaysSnapshot:          nil,
	}
	if err := repo.InsertStageInstances(ctx, stageTx, []domain.StageInstance{stage}); err != nil {
		stageTx.Rollback()
		t.Fatalf("InsertStageInstances: %v", err)
	}
	if err := stageTx.Commit(); err != nil {
		t.Fatalf("commit stage tx: %v", err)
	}

	readTx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin read tx: %v", err)
	}
	defer readTx.Rollback()

	got, err := repo.LoadActiveInstanceBySubject(ctx, readTx, doc.TenantID, string(domain.SubjectKindTemplate), templateKey)
	if err != nil {
		t.Fatalf("LoadActiveInstanceBySubject: %v", err)
	}
	if got.ID != instID {
		t.Fatalf("LoadActiveInstanceBySubject ID = %q, want %q", got.ID, instID)
	}
	if got.Subject.Kind != domain.SubjectKindTemplate || got.Subject.Key != templateKey {
		t.Fatalf("LoadActiveInstanceBySubject Subject = %+v, want {template %s}", got.Subject, templateKey)
	}
	if got.DocumentID != "" {
		t.Fatalf("LoadActiveInstanceBySubject DocumentID = %q, want empty (NULL document_id must not be substituted)", got.DocumentID)
	}
	if len(got.Stages) != 1 {
		t.Fatalf("LoadActiveInstanceBySubject Stages len = %d, want 1", len(got.Stages))
	}
	if got.Stages[0].Status != domain.StageActive {
		t.Fatalf("LoadActiveInstanceBySubject Stages[0].Status = %q, want %q", got.Stages[0].Status, domain.StageActive)
	}
}

// TestLoadActiveInstanceBySubject_CrossTenant_NotFound_RealDB proves the
// tenant_id predicate is load-bearing: an active instance under tenant B's
// subject_key must be invisible to a query scoped to tenant A, even though
// the subject_kind/subject_key match exactly. This is the fail-closed
// integrity guarantee the reviewer flagged as unproven — a broken or dropped
// tenant predicate would leak the row across tenants instead of returning
// ErrNoActiveInstance.
func TestLoadActiveInstanceBySubject_CrossTenant_NotFound_RealDB(t *testing.T) {
	dbc, _ := testdb.Open(t)
	repo := NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})
	ctx := context.Background()

	// Tenant A: an unrelated document/route, used only to seed a distinct
	// tenant_id to query against.
	cdA := testdb.NewControlledDoc(t, dbc)
	docA := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cdA), testdb.WithStatus("approved"))

	// Tenant B: owns the actual template-subject instance.
	cdB := testdb.NewControlledDoc(t, dbc)
	docB := testdb.NewDocument(t, dbc, testdb.WithControlledDoc(cdB), testdb.WithStatus("approved"))
	if docA.TenantID == docB.TenantID {
		t.Fatal("test setup broken: expected distinct tenants for cross-tenant isolation check")
	}
	routeB := testdb.NewApprovalRoute(t, dbc, testdb.WithTenant(docB.TenantID))

	templateKey := "tmpl-" + uuid.NewString()
	instID := uuid.NewString()

	tx := insertInstanceTx(t, dbc, docB.TenantID, docB.Owner)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO public.approval_instances
		   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
		    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
		    subject_kind, subject_key)
		 VALUES ($1::uuid, $2::uuid, NULL, $3::uuid, 1, 'in_progress', $4, now(), repeat('a', 64), $5, 'template', $6)`,
		instID, docB.TenantID, routeB.ID, docB.Owner, "idem-"+uuid.NewString(), templateKey,
	); err != nil {
		tx.Rollback()
		t.Fatalf("raw insert template instance (tenant B): %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	readTx, err := dbc.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin read tx: %v", err)
	}
	defer readTx.Rollback()

	// Query as tenant A with tenant B's exact subject_key: must be a clean
	// not-found, never tenant B's row.
	got, err := repo.LoadActiveInstanceBySubject(ctx, readTx, docA.TenantID, string(domain.SubjectKindTemplate), templateKey)
	if !errors.Is(err, ErrNoActiveInstance) {
		t.Fatalf("LoadActiveInstanceBySubject(tenant A, tenant B's subject_key) err = %v, want ErrNoActiveInstance", err)
	}
	if got != nil {
		t.Fatalf("LoadActiveInstanceBySubject(tenant A, tenant B's subject_key) = %+v, want nil — tenant isolation breach", got)
	}

	// Sanity: the same subject_key under the OWNING tenant (B) is found —
	// proves the not-found above is tenant isolation, not a broken fixture.
	got, err = repo.LoadActiveInstanceBySubject(ctx, readTx, docB.TenantID, string(domain.SubjectKindTemplate), templateKey)
	if err != nil {
		t.Fatalf("LoadActiveInstanceBySubject(tenant B, own subject_key): %v", err)
	}
	if got == nil || got.ID != instID {
		t.Fatalf("LoadActiveInstanceBySubject(tenant B, own subject_key) = %+v, want instance %s", got, instID)
	}
}
