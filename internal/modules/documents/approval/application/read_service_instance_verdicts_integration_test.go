//go:build integration
// +build integration

package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/documents/approval/domain"
	platformdb "metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

// F2d.2 (ADR 0079) — real-DB proof of the verdict-history projection and the
// DB-enforced actor_display_name_snapshot invariant (migration 0294). Drives
// LoadInstanceByDocumentForViewWithViewer (3rd return = []ReviewVerdict) and
// the raw InsertVerdict/InsertSignoff write path against a real Postgres testdb.

// seedReviewStage inserts one review-kind stage; returns its ID.
func seedReviewStage(t *testing.T, fx viewerFixture, order int, eligibleActorIDs []string) string {
	t.Helper()
	stageID := uuid.NewString()
	stage := domain.StageInstance{
		ID:                         stageID,
		ApprovalInstanceID:         fx.instanceID,
		StageOrder:                 order,
		NameSnapshot:               "Revisão",
		RequiredRoleSnapshot:       "reviewer",
		RequiredCapabilitySnapshot: "doc.signoff",
		QuorumSnapshot:             domain.QuorumAllOf,
		OnEligibilityDriftSnapshot: domain.DriftKeepSnapshot,
		Kind:                       domain.StageKindReview,
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

// seedVerdict inserts one verdict via the repo write path (unconditional
// snapshot bind, migration 0294). Fails the test on insert error.
func seedVerdict(t *testing.T, fx viewerFixture, stageID, actorID, name string, verdict domain.Verdict, comment string, at time.Time) {
	t.Helper()
	v, err := domain.NewVerdict(domain.VerdictParams{
		ID:                       uuid.NewString(),
		ApprovalInstanceID:       fx.instanceID,
		StageInstanceID:          stageID,
		StageKind:                domain.StageKindReview,
		ActorUserID:              actorID,
		ActorTenantID:            fx.tenantID,
		Verdict:                  verdict,
		Comment:                  comment,
		VerdictAt:                at,
		ActorDisplayNameSnapshot: name,
	})
	if err != nil {
		t.Fatalf("NewVerdict: %v", err)
	}
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	if _, err := fx.repo.InsertVerdict(context.Background(), tx, *v); err != nil {
		t.Fatalf("InsertVerdict: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func loadVerdicts(t *testing.T, fx viewerFixture, viewerID string) []domain.ReviewVerdict {
	t.Helper()
	svc := newReadServiceForIntegration(t, fx.db)
	runner := platformdb.NewTxRunner(fx.db)
	_, _, verdicts, err := svc.LoadInstanceByDocumentForViewWithViewer(ctxWithIdentity(fx.tenantID, viewerID), runner, fx.tenantID, fx.docID)
	if err != nil {
		t.Fatalf("LoadInstanceByDocumentForViewWithViewer: %v", err)
	}
	return verdicts
}

// TestInstanceVerdicts_Ready — one ready verdict → single record with the
// snapshot name, verdict=ready, reason=comment.
func TestInstanceVerdicts_Ready(t *testing.T) {
	fx := newViewerFixtureBase(t)
	reviewer := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stageID := seedReviewStage(t, fx, 1, []string{reviewer.ID})
	seedVerdict(t, fx, stageID, reviewer.ID, "Ana Reviewer", domain.VerdictReady, "looks good", time.Now().UTC())
	grantDocumentView(t, fx.db, fx.tenantID, reviewer.ID)

	vs := loadVerdicts(t, fx, reviewer.ID)
	if len(vs) != 1 {
		t.Fatalf("verdicts len = %d, want 1", len(vs))
	}
	got := &vs[0]
	if got.Verdict() != domain.VerdictReady {
		t.Errorf("Verdict = %q, want ready", got.Verdict())
	}
	if got.ActorDisplayNameSnapshot() != "Ana Reviewer" {
		t.Errorf("display name = %q, want %q", got.ActorDisplayNameSnapshot(), "Ana Reviewer")
	}
	if got.Comment() != "looks good" {
		t.Errorf("comment = %q, want %q", got.Comment(), "looks good")
	}
}

// TestInstanceVerdicts_RequestChanges — request_changes verdict carries its
// mandatory reason.
func TestInstanceVerdicts_RequestChanges(t *testing.T) {
	fx := newViewerFixtureBase(t)
	reviewer := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stageID := seedReviewStage(t, fx, 1, []string{reviewer.ID})
	seedVerdict(t, fx, stageID, reviewer.ID, "Bob Reviewer", domain.VerdictRequestChanges, "fix section 3", time.Now().UTC())
	grantDocumentView(t, fx.db, fx.tenantID, reviewer.ID)

	vs := loadVerdicts(t, fx, reviewer.ID)
	if len(vs) != 1 {
		t.Fatalf("verdicts len = %d, want 1", len(vs))
	}
	if vs[0].Verdict() != domain.VerdictRequestChanges {
		t.Errorf("Verdict = %q, want request_changes", vs[0].Verdict())
	}
	if vs[0].Comment() != "fix section 3" {
		t.Errorf("comment = %q, want %q", vs[0].Comment(), "fix section 3")
	}
}

// TestInstanceVerdicts_MultiStageOrdered — verdicts on two stages return
// chronological (verdict_at asc), spanning both stages.
func TestInstanceVerdicts_MultiStageOrdered(t *testing.T) {
	fx := newViewerFixtureBase(t)
	r1 := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	r2 := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stage1 := seedReviewStage(t, fx, 1, []string{r1.ID})
	stage2 := seedReviewStage(t, fx, 2, []string{r2.ID})
	base := time.Now().UTC().Add(-time.Hour)
	// Insert out of chronological order to prove ORDER BY, not insert order.
	seedVerdict(t, fx, stage2, r2.ID, "Second", domain.VerdictReady, "", base.Add(30*time.Minute))
	seedVerdict(t, fx, stage1, r1.ID, "First", domain.VerdictReady, "", base)
	grantDocumentView(t, fx.db, fx.tenantID, r1.ID)

	vs := loadVerdicts(t, fx, r1.ID)
	if len(vs) != 2 {
		t.Fatalf("verdicts len = %d, want 2", len(vs))
	}
	if vs[0].ActorDisplayNameSnapshot() != "First" || vs[1].ActorDisplayNameSnapshot() != "Second" {
		t.Errorf("order = [%q, %q], want [First, Second]", vs[0].ActorDisplayNameSnapshot(), vs[1].ActorDisplayNameSnapshot())
	}
	if vs[0].StageInstanceID() == vs[1].StageInstanceID() {
		t.Error("expected verdicts spanning two distinct stages")
	}
}

// TestInstanceVerdicts_None — no verdicts → empty (non-nil handled at handler).
func TestInstanceVerdicts_None(t *testing.T) {
	fx := newViewerFixtureBase(t)
	reviewer := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	seedReviewStage(t, fx, 1, []string{reviewer.ID})
	grantDocumentView(t, fx.db, fx.tenantID, reviewer.ID)

	vs := loadVerdicts(t, fx, reviewer.ID)
	if len(vs) != 0 {
		t.Fatalf("verdicts len = %d, want 0", len(vs))
	}
}

// TestInsertVerdict_EmptySnapshot_Rejected — the DB CHECK/NOT NULL invariant
// (migration 0294) rejects an empty actor_display_name_snapshot at insert.
func TestInsertVerdict_EmptySnapshot_Rejected(t *testing.T) {
	fx := newViewerFixtureBase(t)
	reviewer := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stageID := seedReviewStage(t, fx, 1, []string{reviewer.ID})
	v, err := domain.NewVerdict(domain.VerdictParams{
		ID:                       uuid.NewString(),
		ApprovalInstanceID:       fx.instanceID,
		StageInstanceID:          stageID,
		StageKind:                domain.StageKindReview,
		ActorUserID:              reviewer.ID,
		ActorTenantID:            fx.tenantID,
		Verdict:                  domain.VerdictReady,
		VerdictAt:                time.Now().UTC(),
		ActorDisplayNameSnapshot: "", // empty — must be rejected by the DB
	})
	if err != nil {
		t.Fatalf("NewVerdict: %v", err)
	}
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	_, err = fx.repo.InsertVerdict(context.Background(), tx, *v)
	if err == nil {
		t.Fatal("InsertVerdict with empty snapshot succeeded, want DB rejection")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") &&
		!strings.Contains(strings.ToLower(err.Error()), "not null") &&
		!strings.Contains(strings.ToLower(err.Error()), "check") {
		t.Errorf("error %q does not look like a DB constraint violation", err.Error())
	}
}

// TestInsertSignoff_EmptySnapshot_Rejected — same invariant on approval_signoffs.
func TestInsertSignoff_EmptySnapshot_Rejected(t *testing.T) {
	fx := newViewerFixtureBase(t)
	approver := testdb.NewUser(t, fx.db, testdb.WithTenant(fx.tenantID))
	stageID := seedActiveStage(t, fx, []string{approver.ID})
	sig, err := domain.NewSignoff(domain.SignoffParams{
		ID:                 uuid.NewString(),
		ApprovalInstanceID: fx.instanceID,
		StageInstanceID:    stageID,
		ActorUserID:        approver.ID,
		ActorTenantID:      fx.tenantID,
		Decision:           domain.DecisionApprove,
		SignedAt:           time.Now().UTC(),
		ContentHash:        strings.Repeat("a", 64),
		// ActorDisplayNameSnapshot omitted → empty, must be rejected.
	})
	if err != nil {
		t.Fatalf("NewSignoff: %v", err)
	}
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.signoff"}]`)
	_, err = fx.repo.InsertSignoff(context.Background(), tx, *sig)
	if err == nil {
		t.Fatal("InsertSignoff with empty snapshot succeeded, want DB rejection")
	}
}
