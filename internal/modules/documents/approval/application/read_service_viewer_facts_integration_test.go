//go:build integration
// +build integration

package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/documents/approval/domain"
	approvalrepo "metaldocs/internal/modules/documents/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// F2d.1 (ADR 0078) — real-DB proof of the 7 viewer-facts scenarios in
// spec.md's Validation Gate. Each test drives
// LoadInstanceByDocumentForViewWithViewer against a real Postgres testdb and
// asserts the returned domain.ViewerFacts, proving the read-path composition
// on the SAME write-path primitives (ResolveEligibleIdentity + CheckSoD) that
// domain/viewer_test.go proves in isolation.

// viewerFixture holds the seeded tenant/document/instance/repo needed to
// drive LoadInstanceByDocumentForViewWithViewer.
type viewerFixture struct {
	tenantID   string
	docID      string
	instanceID string
	authorID   string
	repo       approvalrepo.ApprovalRepository
	db         *sql.DB
}

// newViewerFixtureBase seeds tenant + author + document (status approved) +
// approval instance (in_progress, submitted_by=author), WITHOUT any stage —
// callers add the active stage (and any signoffs/delegations) once every
// actor identity the stage needs to reference has been created.
func newViewerFixtureBase(t *testing.T) viewerFixture {
	t.Helper()
	dbc, _ := testdb.Open(t)

	ten := testdb.NewTenant(t, dbc)
	author := testdb.NewUser(t, dbc, testdb.WithTenant(ten.ID))
	doc := testdb.NewDocument(t, dbc, testdb.WithTenant(ten.ID), testdb.WithOwner(author.ID), testdb.WithStatus("approved"))
	inst := testdb.NewApprovalInstance(t, dbc,
		testdb.WithDocument(doc), testdb.WithOwner(author.ID), testdb.WithStatus("in_progress"),
	)

	repo := approvalrepo.NewPostgresApprovalRepository(dbc, iamdomain.NoopUserDisplayNameReader{})

	return viewerFixture{
		tenantID:   ten.ID,
		docID:      doc.ID,
		instanceID: inst.ID,
		authorID:   author.ID,
		repo:       repo,
		db:         dbc,
	}
}

// seedActiveStage inserts the ONE active stage instance carrying
// eligibleActorIDs, returning the stage's ID.
func seedActiveStage(t *testing.T, fx viewerFixture, eligibleActorIDs []string) string {
	t.Helper()
	stageID := uuid.NewString()
	stage := domain.StageInstance{
		ID:                         stageID,
		ApprovalInstanceID:         fx.instanceID,
		StageOrder:                 1,
		NameSnapshot:               "Qualidade",
		RequiredRoleSnapshot:       "reviewer",
		RequiredCapabilitySnapshot: "doc.signoff",
		QuorumSnapshot:             domain.QuorumAllOf,
		OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
		Kind:                       domain.StageKindApproval,
		EligibleActorIDs:           eligibleActorIDs,
		Status:                     domain.StageActive,
	}

	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	if err := fx.repo.InsertStageInstances(context.Background(), tx, []domain.StageInstance{stage}); err != nil {
		t.Fatalf("InsertStageInstances: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return stageID
}

// grantDocumentView grants the tenant-grade "viewer" role (document.view
// capability) to userID somewhere in the tenant — required to clear this
// package's tier-1 CapDocumentView gate before requireInstanceVisible even
// runs.
func grantDocumentView(t *testing.T, dbc *sql.DB, tenantID, userID string) {
	t.Helper()
	area := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tenantID)).ProcessAreaCode
	grantRoleInArea(t, dbc, tenantID, userID, area, "viewer")
}

// grantApprovalOversee grants the tenant-wide CapApprovalOversee capability
// (also clears the tier-1 gate and satisfies requireInstanceVisible's
// non-pool-member fallback) — used for the observer scenario, which must be
// able to SEE the instance without being an eligible actor on any stage.
func grantApprovalOversee(t *testing.T, dbc *sql.DB, tenantID, userID string) {
	t.Helper()
	area := testdb.NewTaxonomy(t, dbc, testdb.WithTenant(tenantID)).ProcessAreaCode
	grantRoleInArea(t, dbc, tenantID, userID, area, "qms_admin")
}

func seedSignoff(t *testing.T, fx viewerFixture, stageID, actorID string) {
	t.Helper()
	sig, err := domain.NewSignoff(domain.SignoffParams{
		ID:                 uuid.NewString(),
		ApprovalInstanceID: fx.instanceID,
		StageInstanceID:    stageID,
		ActorUserID:        actorID,
		ActorTenantID:      fx.tenantID,
		Decision:                 domain.DecisionApprove,
		SignedAt:                 time.Now().UTC(),
		ContentHash:              "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1",
		ActorDisplayNameSnapshot: "Signed Actor", // migration 0294: snapshot is NOT NULL / non-empty
	})
	if err != nil {
		t.Fatalf("NewSignoff: %v", err)
	}
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	// approval_signoffs carries a DB write-tripwire requiring the
	// document.signoff capability be asserted tx-locally (mirrors what the
	// production authz layer does at the signoff write site). Assert it here so
	// the seed insert satisfies the tripwire.
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.signoff"}]`)
	if _, err := fx.repo.InsertSignoff(context.Background(), tx, *sig); err != nil {
		t.Fatalf("InsertSignoff: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func seedDelegation(t *testing.T, fx viewerFixture, delegatorID, delegateID string) {
	t.Helper()
	now := time.Now().UTC()
	d, err := domain.NewDelegation(domain.DelegationParams{
		ID:          uuid.NewString(),
		TenantID:    fx.tenantID,
		DelegatorID: delegatorID,
		DelegateID:  delegateID,
		StartsAt:    now.Add(-time.Hour),
		EndsAt:      now.Add(time.Hour),
		Reason:      "F2d.1 viewer facts integration test",
		CreatedBy:   delegatorID,
		CreatedAt:   now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("NewDelegation: %v", err)
	}
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	if err := fx.repo.InsertDelegation(context.Background(), tx, *d); err != nil {
		t.Fatalf("InsertDelegation: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func loadViewerFacts(t *testing.T, fx viewerFixture, viewerID string) (*domain.Instance, domain.ViewerFacts) {
	t.Helper()
	svc := newReadServiceForIntegration(t, fx.db)
	runner := platformdb.NewTxRunner(fx.db)
	inst, vf, _, err := svc.LoadInstanceByDocumentForViewWithViewer(ctxWithIdentity(fx.tenantID, viewerID), runner, fx.tenantID, fx.docID)
	if err != nil {
		t.Fatalf("LoadInstanceByDocumentForViewWithViewer: %v", err)
	}
	return inst, vf
}

// TestViewerBlock_Author — the submitting author: is_author=true,
// eligible=false by the author-rule (CheckSoD), even when the author is NOT
// in the stage's eligible_actor_ids.
func TestViewerBlock_Author(t *testing.T) {
	fx := newViewerFixtureBase(t)
	approver := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	seedActiveStage(t, fx, []string{approver.ID})
	grantDocumentView(t, fx.db, fx.tenantID, fx.authorID)

	_, vf := loadViewerFacts(t, fx, fx.authorID)

	if !vf.IsAuthor {
		t.Error("IsAuthor = false, want true")
	}
	if vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = true, want false (author-rule)")
	}
	if vf.HasSignedActiveStage {
		t.Error("HasSignedActiveStage = true, want false")
	}
	if vf.ViaDelegationFromUserID != "" {
		t.Errorf("ViaDelegationFromUserID = %q, want empty", vf.ViaDelegationFromUserID)
	}
}

// TestViewerBlock_SnapshotActor — a viewer directly in eligible_actor_ids:
// eligible=true, via_delegation_from = "" (not via delegation).
func TestViewerBlock_SnapshotActor(t *testing.T) {
	fx := newViewerFixtureBase(t)
	approver := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	seedActiveStage(t, fx, []string{approver.ID})
	grantDocumentView(t, fx.db, fx.tenantID, approver.ID)

	_, vf := loadViewerFacts(t, fx, approver.ID)

	if vf.IsAuthor {
		t.Error("IsAuthor = true, want false")
	}
	if !vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = false, want true")
	}
	if vf.HasSignedActiveStage {
		t.Error("HasSignedActiveStage = true, want false")
	}
	if vf.ViaDelegationFromUserID != "" {
		t.Errorf("ViaDelegationFromUserID = %q, want empty (direct snapshot membership)", vf.ViaDelegationFromUserID)
	}
}

// TestViewerBlock_Delegate — a delegate acting via an active delegation from
// a snapshot-eligible delegator: eligible=true, via_delegation_from =
// delegator.
func TestViewerBlock_Delegate(t *testing.T) {
	fx := newViewerFixtureBase(t)
	delegator := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	delegate := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	seedActiveStage(t, fx, []string{delegator.ID})
	seedDelegation(t, fx, delegator.ID, delegate.ID)
	grantDocumentView(t, fx.db, fx.tenantID, delegate.ID)

	_, vf := loadViewerFacts(t, fx, delegate.ID)

	if !vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = false, want true")
	}
	if vf.ViaDelegationFromUserID != delegator.ID {
		t.Errorf("ViaDelegationFromUserID = %q, want delegator %q", vf.ViaDelegationFromUserID, delegator.ID)
	}
}

// TestViewerBlock_AlreadySigned — an eligible actor who already recorded a
// signoff on the active stage: eligible=true, has_signed=true.
func TestViewerBlock_AlreadySigned(t *testing.T) {
	fx := newViewerFixtureBase(t)
	approver := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stageID := seedActiveStage(t, fx, []string{approver.ID})
	seedSignoff(t, fx, stageID, approver.ID)
	grantDocumentView(t, fx.db, fx.tenantID, approver.ID)

	_, vf := loadViewerFacts(t, fx, approver.ID)

	if !vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = false, want true")
	}
	if !vf.HasSignedActiveStage {
		t.Error("HasSignedActiveStage = false, want true")
	}
}

// TestViewerBlock_Observer — not the author, not in eligible_actor_ids, no
// delegation: all facts false/empty. Visibility is granted via
// CapApprovalOversee so the read succeeds (visibility != eligibility).
func TestViewerBlock_Observer(t *testing.T) {
	fx := newViewerFixtureBase(t)
	approver := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	observer := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	seedActiveStage(t, fx, []string{approver.ID})
	grantApprovalOversee(t, fx.db, fx.tenantID, observer.ID)

	_, vf := loadViewerFacts(t, fx, observer.ID)

	if vf.IsAuthor {
		t.Error("IsAuthor = true, want false")
	}
	if vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = true, want false")
	}
	if vf.HasSignedActiveStage {
		t.Error("HasSignedActiveStage = true, want false")
	}
	if vf.ViaDelegationFromUserID != "" {
		t.Errorf("ViaDelegationFromUserID = %q, want empty", vf.ViaDelegationFromUserID)
	}
}

// TestViewerBlock_AuthorAlsoApprover — the author is ALSO in the active
// stage's eligible_actor_ids snapshot: still excluded by the author-rule
// (CheckSoD), eligible=false.
func TestViewerBlock_AuthorAlsoApprover(t *testing.T) {
	fx := newViewerFixtureBase(t)
	seedActiveStage(t, fx, []string{fx.authorID})
	grantDocumentView(t, fx.db, fx.tenantID, fx.authorID)

	_, vf := loadViewerFacts(t, fx, fx.authorID)

	if !vf.IsAuthor {
		t.Error("IsAuthor = false, want true")
	}
	if vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = true, want false (author-rule excludes even snapshot membership)")
	}
}

// TestViewerBlock_NoActiveStage — a terminal instance (no active stage): all
// facts false/empty, is_author still reflects reality.
func TestViewerBlock_NoActiveStage(t *testing.T) {
	fx := newViewerFixtureBase(t)
	// No seedActiveStage call: the instance has zero stages, so Active()
	// returns nil — mirrors a published/cancelled terminal instance.
	grantDocumentView(t, fx.db, fx.tenantID, fx.authorID)

	_, vf := loadViewerFacts(t, fx, fx.authorID)

	if !vf.IsAuthor {
		t.Error("IsAuthor = false, want true")
	}
	if vf.EligibleForActiveStage {
		t.Error("EligibleForActiveStage = true, want false (no active stage)")
	}
	if vf.HasSignedActiveStage {
		t.Error("HasSignedActiveStage = true, want false")
	}
	if vf.ViaDelegationFromUserID != "" {
		t.Errorf("ViaDelegationFromUserID = %q, want empty", vf.ViaDelegationFromUserID)
	}
}
