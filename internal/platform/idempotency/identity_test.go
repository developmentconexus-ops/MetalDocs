package idempotency

import (
	"context"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// TestTenantActorFromContext_MissingTenant_ReturnsError is the RED-then-GREEN
// proof for #90/A3.5: before this change, taxonomy's (and templates'/
// approval's idempotentHandler's) hand-copied extraction discarded
// tenant.FromContext's error and let "" flow forward. TenantActorFromContext
// is the shared replacement; this test proves it fails closed the same way
// authn.RequireUserID already does for a missing actor.
//
// RED proof: reverting TenantActorFromContext to `tenantID, _ :=
// tenant.FromContext(ctx)` (the exact pre-fix taxonomy/templates shape) makes
// this test fail, because err would be nil and tenantID would be "" instead
// of erroring.
func TestTenantActorFromContext_MissingTenant_ReturnsError(t *testing.T) {
	ctx := iamdomain.WithAuthContext(context.Background(), "actor-1", nil)
	// Deliberately do NOT call tenant.WithTenantID — this is the "authenticated
	// actor, absent tenant claim" shape the A3.3-deferred defect exercised.

	gotTenant, gotActor, err := tenantActorFromContext(ctx)

	if err == nil {
		t.Fatal("expected an error for a missing tenant claim, got nil")
	}
	if !errors.Is(err, tenant.ErrTenantMissing) {
		t.Fatalf("expected tenant.ErrTenantMissing, got: %v", err)
	}
	if gotTenant != "" || gotActor != "" {
		t.Fatalf("expected both return values empty on error, got tenant=%q actor=%q", gotTenant, gotActor)
	}
}

// TestTenantActorFromContext_MissingActor_ReturnsError proves the actor half
// still fails closed too (unchanged from the pre-existing authn.RequireUserID
// contract) — the shared resolver must not regress the half that already
// worked while fixing the half that didn't.
func TestTenantActorFromContext_MissingActor_ReturnsError(t *testing.T) {
	ctx := tenant.WithTenantID(context.Background(), "11111111-1111-4111-8111-111111111111")
	// Deliberately do NOT call iamdomain.WithAuthContext.

	gotTenant, gotActor, err := tenantActorFromContext(ctx)

	if err == nil {
		t.Fatal("expected an error for a missing actor claim, got nil")
	}
	if gotTenant != "" || gotActor != "" {
		t.Fatalf("expected both return values empty on error, got tenant=%q actor=%q", gotTenant, gotActor)
	}
}

// TestTenantActorFromContext_BothPresent_ReturnsBoth is the control: proves
// the happy path is unaffected by the guard added for the two failure cases
// above.
func TestTenantActorFromContext_BothPresent_ReturnsBoth(t *testing.T) {
	const wantTenant = "22222222-2222-4222-8222-222222222222"
	const wantActor = "actor-2"
	ctx := tenant.WithTenantID(context.Background(), wantTenant)
	ctx = iamdomain.WithAuthContext(ctx, wantActor, nil)

	gotTenant, gotActor, err := tenantActorFromContext(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTenant != wantTenant || gotActor != wantActor {
		t.Fatalf("got tenant=%q actor=%q, want tenant=%q actor=%q", gotTenant, gotActor, wantTenant, wantActor)
	}
}
