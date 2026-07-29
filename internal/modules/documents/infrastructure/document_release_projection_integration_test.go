//go:build integration
// +build integration

package infrastructure_test

// ADR 0085 Stage C (W3) — read-side integration proof for the readiness-hold
// projection on the document detail read model (contract-first:
// api/openapi/v1/openapi.yaml DocumentDetailResponse.release ->
// DocumentReleaseProjection, required+nullable per ADR 0035).
//
// Proves, against the real GetDocument query (the SINGLE detail read; the
// projection is a LEFT JOIN LATERAL inside it, never a second round-trip):
//
//   1. a document with NO release generation -> Release nil (no synthesized
//      hold — pre-approval drafts and legacy non-backfilled rows);
//   2. a HELD generation -> State=hold, hold_reason/hold_detail/
//      last_evaluated_at surfaced, released_at nil, and
//      planned_effective_from sourced from documents.planned_effective_from
//      (the PLAN half of ADR 0085's planned/actual split), NOT from the
//      generation row;
//   3. a RELEASED generation -> State=released + released_at (state is derived
//      server-side from released_at IS NOT NULL, never stored);
//   4. latest-generation-wins: with two generations for one document, the one
//      with the higher generation_seq is projected;
//   5. tenant isolation: a generation row carrying another tenant's tenant_id
//      is invisible to the owning tenant's detail read (the lateral's explicit
//      tenant_id predicate — RLS-truth discipline, since the tenant_isolation
//      policy has a NULL-GUC bypass branch).

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

// insertReleaseGeneration writes ONE public.release_generations row for
// (tenantID, documentID) and returns its id. release_generations carries no
// tripwire (migration 0310: "written only by the approval kernel's own
// in-transaction fact producers and by the coordinator job"), so a plain INSERT
// is the fixture path — no cap assertion needed.
//
// A released row must satisfy ck_release_generations_released_implies_facts
// (both fact timestamps present) and ck_release_generations_artifact_fact_full_set
// (artifact fact implies BOTH final artifact keys), so releasedAt != nil also
// stamps the two facts and both s3 keys.
func insertReleaseGeneration(
	t *testing.T,
	db *sql.DB,
	tenantID, documentID string,
	holdReason, holdDetail *string,
	releasedAt, lastEvaluatedAt *time.Time,
) string {
	t.Helper()
	ctx := context.Background()

	var approvalFactAt, artifactFactAt *time.Time
	var docxKey, pdfKey *string
	if releasedAt != nil {
		factAt := releasedAt.Add(-time.Hour)
		approvalFactAt = &factAt
		artifactFactAt = &factAt
		d := "tenant/" + tenantID + "/final.docx"
		p := "tenant/" + tenantID + "/final.pdf"
		docxKey = &d
		pdfKey = &p
	}

	var id string
	err := db.QueryRowContext(ctx, `
		INSERT INTO public.release_generations
		    (tenant_id, document_id, approval_instance_id, revision_id,
		     revision_version, frozen_content_hash,
		     approval_fact_at, artifact_fact_at, final_docx_s3_key, final_pdf_s3_key,
		     hold_reason, hold_detail, released_at, last_evaluated_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid,
		        1, $5,
		        $6, $7, $8, $9,
		        $10, $11, $12, $13)
		RETURNING id::text`,
		tenantID, documentID, uuid.NewString(), uuid.NewString(),
		strings.Repeat("a", 64),
		approvalFactAt, artifactFactAt, docxKey, pdfKey,
		holdReason, holdDetail, releasedAt, lastEvaluatedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertReleaseGeneration: %v", err)
	}
	return id
}

func newReleaseProjectionRepo(db *sql.DB) *infrastructure.Repository {
	return infrastructure.New(db,
		iamdomain.NoopUserDisplayNameReader{},
		controlleddocumentsdomain.NoopCDFieldReader{},
		taxonomydomain.NoopAreaCatalogReader{})
}

// TestGetDocument_ReleaseProjection_NoGeneration proves a document with no
// release generation projects NOTHING — no-fallback, never a synthesized hold.
func TestGetDocument_ReleaseProjection_NoGeneration(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := newReleaseProjectionRepo(db)

	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))

	got, err := repo.GetDocument(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.Release != nil {
		t.Fatalf("Release = %+v, want nil for a document with no release generation", got.Release)
	}
}

// TestGetDocument_ReleaseProjection_Held proves a generation with released_at
// NULL projects state=hold with the recorded reason/detail, and that
// planned_effective_from comes from the DOCUMENT row.
func TestGetDocument_ReleaseProjection_Held(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := newReleaseProjectionRepo(db)

	now := time.Now().UTC().Truncate(time.Second)
	plannedEffectiveFrom := now.Add(14 * 24 * time.Hour)
	lastEvaluatedAt := now.Add(-3 * time.Minute)

	doc := testdb.NewDocument(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithStatus("approved"),
		testdb.WithPlannedEffectiveFrom(plannedEffectiveFrom),
	)

	holdReason := "materializing"
	holdDetail := "final PDF still rendering"
	genID := insertReleaseGeneration(t, db, tnt.ID, doc.ID, &holdReason, &holdDetail, nil, &lastEvaluatedAt)

	got, err := repo.GetDocument(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	rel := got.Release
	if rel == nil {
		t.Fatalf("Release = nil, want the held generation projection")
	}
	if rel.GenerationID != genID {
		t.Errorf("GenerationID = %q, want %q", rel.GenerationID, genID)
	}
	if rel.State != domain.ReleaseStateHold {
		t.Errorf("State = %q, want %q", rel.State, domain.ReleaseStateHold)
	}
	if rel.HoldReason == nil || *rel.HoldReason != holdReason {
		t.Errorf("HoldReason = %v, want %q", rel.HoldReason, holdReason)
	}
	if rel.HoldDetail == nil || *rel.HoldDetail != holdDetail {
		t.Errorf("HoldDetail = %v, want %q", rel.HoldDetail, holdDetail)
	}
	if rel.ReleasedAt != nil {
		t.Errorf("ReleasedAt = %v, want nil while on hold", rel.ReleasedAt)
	}
	if rel.LastEvaluatedAt == nil || !rel.LastEvaluatedAt.Equal(lastEvaluatedAt) {
		t.Errorf("LastEvaluatedAt = %v, want %v", rel.LastEvaluatedAt, lastEvaluatedAt)
	}
	if rel.PlannedEffectiveFrom == nil || !rel.PlannedEffectiveFrom.Equal(plannedEffectiveFrom) {
		t.Errorf("PlannedEffectiveFrom = %v, want %v (sourced from documents.planned_effective_from)",
			rel.PlannedEffectiveFrom, plannedEffectiveFrom)
	}
}

// TestGetDocument_ReleaseProjection_Released proves released_at IS NOT NULL
// derives state=released and surfaces the release timestamp.
func TestGetDocument_ReleaseProjection_Released(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := newReleaseProjectionRepo(db)

	now := time.Now().UTC().Truncate(time.Second)
	releasedAt := now.Add(-2 * time.Hour)
	lastEvaluatedAt := releasedAt

	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("draft"))
	// The winning release tx clears hold_reason/hold_detail; mirror that here.
	genID := insertReleaseGeneration(t, db, tnt.ID, doc.ID, nil, nil, &releasedAt, &lastEvaluatedAt)

	got, err := repo.GetDocument(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	rel := got.Release
	if rel == nil {
		t.Fatalf("Release = nil, want the released generation projection")
	}
	if rel.GenerationID != genID {
		t.Errorf("GenerationID = %q, want %q", rel.GenerationID, genID)
	}
	if rel.State != domain.ReleaseStateReleased {
		t.Errorf("State = %q, want %q", rel.State, domain.ReleaseStateReleased)
	}
	if rel.ReleasedAt == nil || !rel.ReleasedAt.Equal(releasedAt) {
		t.Errorf("ReleasedAt = %v, want %v", rel.ReleasedAt, releasedAt)
	}
	if rel.HoldReason != nil || rel.HoldDetail != nil {
		t.Errorf("HoldReason/HoldDetail = %v/%v, want nil/nil on a released generation", rel.HoldReason, rel.HoldDetail)
	}
	// planned_effective_from was never set on this document -> nil ("effective
	// on release"), NOT a fallback to the actual release timestamp.
	if rel.PlannedEffectiveFrom != nil {
		t.Errorf("PlannedEffectiveFrom = %v, want nil (no plan date on the document row)", rel.PlannedEffectiveFrom)
	}
}

// TestGetDocument_ReleaseProjection_LatestGenerationWins proves the lateral's
// ORDER BY generation_seq DESC LIMIT 1: with an older RELEASED generation and a
// newer HELD one (the re-approval-after-release shape), the read model projects
// the NEWER one.
func TestGetDocument_ReleaseProjection_LatestGenerationWins(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	repo := newReleaseProjectionRepo(db)

	now := time.Now().UTC().Truncate(time.Second)
	oldReleasedAt := now.Add(-48 * time.Hour)
	newEvaluatedAt := now.Add(-1 * time.Minute)

	doc := testdb.NewDocument(t, db, testdb.WithTenant(tnt.ID), testdb.WithStatus("under_review"))

	oldGenID := insertReleaseGeneration(t, db, tnt.ID, doc.ID, nil, nil, &oldReleasedAt, &oldReleasedAt)
	holdReason := "awaiting_approval_fact"
	newGenID := insertReleaseGeneration(t, db, tnt.ID, doc.ID, &holdReason, nil, nil, &newEvaluatedAt)

	got, err := repo.GetDocument(ctx, tnt.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	rel := got.Release
	if rel == nil {
		t.Fatalf("Release = nil, want the latest generation projection")
	}
	if rel.GenerationID == oldGenID {
		t.Fatalf("Release projected the OLDER generation %q; latest (higher generation_seq) must win", oldGenID)
	}
	if rel.GenerationID != newGenID {
		t.Fatalf("GenerationID = %q, want the latest generation %q", rel.GenerationID, newGenID)
	}
	if rel.State != domain.ReleaseStateHold {
		t.Errorf("State = %q, want %q (latest generation is still held)", rel.State, domain.ReleaseStateHold)
	}
	if rel.HoldReason == nil || *rel.HoldReason != holdReason {
		t.Errorf("HoldReason = %v, want %q", rel.HoldReason, holdReason)
	}
}

// TestGetDocument_ReleaseProjection_TenantIsolation proves the lateral's
// explicit tenant_id predicate: a generation row that names ANOTHER tenant but
// this document's id is invisible to the owning tenant's detail read. This is
// the RLS-truth belt-and-suspenders — the tenant_isolation policy on
// public.release_generations carries a NULL-GUC bypass branch, so the read must
// not depend on the GUC alone.
func TestGetDocument_ReleaseProjection_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	repo := newReleaseProjectionRepo(db)

	doc := testdb.NewDocument(t, db, testdb.WithTenant(tntA.ID), testdb.WithStatus("draft"))

	// Foreign generation: tenant B's row pointing at tenant A's document id.
	holdReason := "supersede_conflict"
	foreignGenID := insertReleaseGeneration(t, db, tntB.ID, doc.ID, &holdReason, nil, nil, nil)

	got, err := repo.GetDocument(ctx, tntA.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.Release != nil {
		t.Fatalf("Release = %+v (generation %q belongs to tenant %s), want nil — cross-tenant generation must be invisible",
			got.Release, foreignGenID, tntB.ID)
	}

	// Control: once tenant A owns a generation of its own, it IS projected —
	// proving the nil above is isolation, not a broken join.
	ownGenID := insertReleaseGeneration(t, db, tntA.ID, doc.ID, &holdReason, nil, nil, nil)
	got, err = repo.GetDocument(ctx, tntA.ID, doc.ID)
	if err != nil {
		t.Fatalf("GetDocument (control): %v", err)
	}
	if got.Release == nil || got.Release.GenerationID != ownGenID {
		t.Fatalf("Release = %+v, want the tenant's own generation %q", got.Release, ownGenID)
	}
}
