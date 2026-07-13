//go:build integration
// +build integration

// Package approval_test — unit 3.2 slice 5 (M4 ActorSelector, B5 submit
// chosen_actors) real-Postgres coverage for
// postgresApprovalRepository.ValidateSubmitChoiceActors
// (docs/superpowers/specs/2026-07-13-unit-3.2-actor-selector-design.md §5).
//
// The service-level branching (missing entry, empty user_ids, wrong-stage
// target, union) is fully unit-covered with a fake repo in
// internal/modules/approval/application/submit_choice_test.go. The ONLY new
// DB code slice 5 adds is the constraint-pool membership query inside
// ValidateSubmitChoiceActors — that is what this real-DB suite exercises:
//
//  1. Accept: every chosen user active in role x area_code returns the
//     deduped, ascending-sorted validated ids.
//  2. Reject (wrong role): a chosen user active in the area but under a
//     different role fails closed with ErrSubmitChoiceConstraintViolated.
//  3. Reject (non-member): a chosen user with no grant at all fails closed.
//  4. Dedup: duplicate chosen ids collapse to a single entry.
//
// Reuses seedProcessArea / grantActiveArea / sortedIDs from
// resolve_eligible_actors_for_selectors_integration_test.go (same package).
package approval_test

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// TestValidateSubmitChoiceActors_AcceptsMembersOfConstraintPool proves that a
// set of chosen users, all active in the submit_choice selector's
// role x area_code, is accepted and returned deduped + ascending-sorted.
func TestValidateSubmitChoiceActors_AcceptsMembersOfConstraintPool(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const area = "quality"
	seedProcessArea(t, database, tenantID, area, "Quality")

	chosen1 := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	chosen2 := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	grantActiveArea(t, database, tenantID, chosen1, area, "approver")
	grantActiveArea(t, database, tenantID, chosen2, area, "approver")

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: area}
	got, err := repo.ValidateSubmitChoiceActors(ctx, tx, tenantID, sel, []string{chosen2, chosen1})
	if err != nil {
		t.Fatalf("ValidateSubmitChoiceActors: %v", err)
	}

	want := sortedIDs([]string{chosen1, chosen2})
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// TestValidateSubmitChoiceActors_RejectsWrongRole proves a user active in the
// area but under a DIFFERENT role than the selector's is not a constraint
// member — fail-closed with ErrSubmitChoiceConstraintViolated.
func TestValidateSubmitChoiceActors_RejectsWrongRole(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const area = "quality"
	seedProcessArea(t, database, tenantID, area, "Quality")

	inPool := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	wrongRole := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	grantActiveArea(t, database, tenantID, inPool, area, "approver")
	// Active in the same area, but as editor — outside an approver constraint.
	grantActiveArea(t, database, tenantID, wrongRole, area, "editor")

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: area}
	_, err = repo.ValidateSubmitChoiceActors(ctx, tx, tenantID, sel, []string{inPool, wrongRole})
	if !errors.Is(err, domain.ErrSubmitChoiceConstraintViolated) {
		t.Fatalf("error = %v, want ErrSubmitChoiceConstraintViolated", err)
	}
}

// TestValidateSubmitChoiceActors_RejectsNonMember proves a chosen user with no
// grant at all in the constraint role x area fails closed.
func TestValidateSubmitChoiceActors_RejectsNonMember(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const area = "quality"
	seedProcessArea(t, database, tenantID, area, "Quality")

	inPool := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	grantActiveArea(t, database, tenantID, inPool, area, "approver")
	// nonMember has a user row but no grant in (approver, quality).
	nonMember := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: area}
	_, err = repo.ValidateSubmitChoiceActors(ctx, tx, tenantID, sel, []string{inPool, nonMember})
	if !errors.Is(err, domain.ErrSubmitChoiceConstraintViolated) {
		t.Fatalf("error = %v, want ErrSubmitChoiceConstraintViolated", err)
	}
}

// TestValidateSubmitChoiceActors_DedupsDuplicateChoice proves a duplicate
// chosen id collapses to a single validated entry.
func TestValidateSubmitChoiceActors_DedupsDuplicateChoice(t *testing.T) {
	ctx := context.Background()
	database, _ := testdb.Open(t)
	repo := infrastructure.NewPostgresApprovalRepository(database, iamdomain.NoopUserDisplayNameReader{})

	tenantID := testdb.NewTenant(t, database).ID
	const area = "quality"
	seedProcessArea(t, database, tenantID, area, "Quality")

	chosen := testdb.NewUser(t, database, testdb.WithTenant(tenantID)).ID
	grantActiveArea(t, database, tenantID, chosen, area, "approver")

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: area}
	got, err := repo.ValidateSubmitChoiceActors(ctx, tx, tenantID, sel, []string{chosen, chosen})
	if err != nil {
		t.Fatalf("ValidateSubmitChoiceActors: %v", err)
	}
	if len(got) != 1 || got[0] != chosen {
		t.Fatalf("got %v, want single [%s]", got, chosen)
	}
}
