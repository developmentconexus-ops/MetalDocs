package application

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	"metaldocs/internal/platform/db"
)

// ---------------------------------------------------------------------------
// submit_choice resolution unit tests (M4, unit 3.2, slice 5). No DB — these
// exercise the pure branching helpers shared by SubmitService and
// TemplateSubmitService (submit_service.go: stagesWithSubmitChoice,
// validateChosenActorsTargets, findChosenActors, resolveSubmitChoiceEligibleIDs,
// unionUserIDs).
// ---------------------------------------------------------------------------

func submitChoiceRoute() []domain.Stage {
	return []domain.Stage{
		{
			Order: 1,
			Kind:  domain.StageKindReview,
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorRoleInFixedArea, Role: "editor", AreaCode: "QA"},
			},
		},
		{
			Order: 2,
			Kind:  domain.StageKindApproval,
			Selectors: []domain.ActorSelector{
				{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: "QA"},
			},
		},
	}
}

func TestStagesWithSubmitChoice(t *testing.T) {
	stages := submitChoiceRoute()
	got := stagesWithSubmitChoice(stages)
	if len(got) != 1 {
		t.Fatalf("stagesWithSubmitChoice: got %d entries, want 1", len(got))
	}
	sel, ok := got[2]
	if !ok {
		t.Fatal("stagesWithSubmitChoice: expected an entry for stage order 2")
	}
	if sel.Kind != domain.SelectorSubmitChoice || sel.Role != "approver" || sel.AreaCode != "QA" {
		t.Fatalf("stagesWithSubmitChoice: got selector %+v, want submit_choice/approver/QA", sel)
	}
	if _, ok := got[1]; ok {
		t.Fatal("stagesWithSubmitChoice: stage 1 (role_in_fixed_area only) must not be present")
	}
}

func TestValidateChosenActorsTargets(t *testing.T) {
	submitChoiceStages := stagesWithSubmitChoice(submitChoiceRoute())

	t.Run("no chosen actors is legal", func(t *testing.T) {
		if err := validateChosenActorsTargets(nil, submitChoiceStages); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("chosen actors targeting the submit_choice stage is legal", func(t *testing.T) {
		chosen := []StageChosenActors{{StageOrder: 2, UserIDs: []string{"user-1"}}}
		if err := validateChosenActorsTargets(chosen, submitChoiceStages); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("chosen actors targeting a non-submit_choice stage fails closed", func(t *testing.T) {
		// Stage 1 has only a role_in_fixed_area selector — no submit_choice.
		chosen := []StageChosenActors{{StageOrder: 1, UserIDs: []string{"user-1"}}}
		err := validateChosenActorsTargets(chosen, submitChoiceStages)
		if !errors.Is(err, domain.ErrSubmitChoiceConstraintViolated) {
			t.Fatalf("error = %v, want ErrSubmitChoiceConstraintViolated", err)
		}
	})

	t.Run("chosen actors targeting a stage_order that does not exist fails closed", func(t *testing.T) {
		chosen := []StageChosenActors{{StageOrder: 99, UserIDs: []string{"user-1"}}}
		err := validateChosenActorsTargets(chosen, submitChoiceStages)
		if !errors.Is(err, domain.ErrSubmitChoiceConstraintViolated) {
			t.Fatalf("error = %v, want ErrSubmitChoiceConstraintViolated", err)
		}
	})
}

func TestFindChosenActors(t *testing.T) {
	chosen := []StageChosenActors{
		{StageOrder: 1, UserIDs: []string{"a"}},
		{StageOrder: 2, UserIDs: []string{"b", "c"}},
	}

	got, ok := findChosenActors(chosen, 2)
	if !ok {
		t.Fatal("expected an entry for stage order 2")
	}
	if len(got.UserIDs) != 2 || got.UserIDs[0] != "b" || got.UserIDs[1] != "c" {
		t.Fatalf("got %+v, want UserIDs [b c]", got)
	}

	if _, ok := findChosenActors(chosen, 3); ok {
		t.Fatal("expected no entry for stage order 3")
	}
}

func TestUnionUserIDs(t *testing.T) {
	got := unionUserIDs([]string{"b", "a"}, []string{"a", "c"}, nil)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if empty := unionUserIDs(); empty == nil || len(empty) != 0 {
		t.Fatalf("unionUserIDs() with no pools = %v, want empty non-nil slice", empty)
	}
}

// fakeSubmitChoiceRepo is a minimal ApprovalRepository stub exercising only
// ValidateSubmitChoiceActors, for resolveSubmitChoiceEligibleIDs tests. Embeds
// the interface (nil) so every other method panics if called — matches the
// fakeSubmitRepo pattern in submit_service_test.go.
type fakeSubmitChoiceRepo struct {
	infrastructure.ApprovalRepository
	validated []string
	err       error
}

func (r *fakeSubmitChoiceRepo) ValidateSubmitChoiceActors(_ context.Context, _ db.Tx, _ string, _ domain.ActorSelector, _ []string) ([]string, error) {
	return r.validated, r.err
}

func TestResolveSubmitChoiceEligibleIDs_MissingEntry(t *testing.T) {
	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: "QA"}
	_, err := resolveSubmitChoiceEligibleIDs(context.Background(), nil, &fakeSubmitChoiceRepo{}, "tenant-1", sel, nil, 2)
	if !errors.Is(err, domain.ErrSubmitChoiceRequired) {
		t.Fatalf("error = %v, want ErrSubmitChoiceRequired", err)
	}
}

func TestResolveSubmitChoiceEligibleIDs_EmptyUserIDsEntry(t *testing.T) {
	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: "QA"}
	chosen := []StageChosenActors{{StageOrder: 2, UserIDs: nil}}
	_, err := resolveSubmitChoiceEligibleIDs(context.Background(), nil, &fakeSubmitChoiceRepo{}, "tenant-1", sel, chosen, 2)
	if !errors.Is(err, domain.ErrSubmitChoiceRequired) {
		t.Fatalf("error = %v, want ErrSubmitChoiceRequired", err)
	}
}

func TestResolveSubmitChoiceEligibleIDs_ConstraintViolation(t *testing.T) {
	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: "QA"}
	chosen := []StageChosenActors{{StageOrder: 2, UserIDs: []string{"user-not-in-pool"}}}
	repo := &fakeSubmitChoiceRepo{err: domain.ErrSubmitChoiceConstraintViolated}
	_, err := resolveSubmitChoiceEligibleIDs(context.Background(), nil, repo, "tenant-1", sel, chosen, 2)
	if !errors.Is(err, domain.ErrSubmitChoiceConstraintViolated) {
		t.Fatalf("error = %v, want ErrSubmitChoiceConstraintViolated", err)
	}
}

func TestResolveSubmitChoiceEligibleIDs_Success(t *testing.T) {
	sel := domain.ActorSelector{Kind: domain.SelectorSubmitChoice, Role: "approver", AreaCode: "QA"}
	chosen := []StageChosenActors{{StageOrder: 2, UserIDs: []string{"user-1", "user-2"}}}
	repo := &fakeSubmitChoiceRepo{validated: []string{"user-1", "user-2"}}
	got, err := resolveSubmitChoiceEligibleIDs(context.Background(), nil, repo, "tenant-1", sel, chosen, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "user-1" || got[1] != "user-2" {
		t.Fatalf("got %v, want [user-1 user-2]", got)
	}
}
