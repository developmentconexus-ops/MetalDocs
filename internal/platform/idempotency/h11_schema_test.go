//go:build integration

package idempotency_test

// H11 regression tests: actor_user_id non-empty, response_body BYTEA + 64 KiB cap.

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/tests/integration/testdb"
)

const h11Actor = "actor-h11"

// TestBeginReplay_EmptyActorReturnsError proves the Go-layer guard rejects
// an empty actorID before any DB round-trip.
func TestBeginReplay_EmptyActorReturnsError(t *testing.T) {
	s := idempotency.New(nil, "POST /h11/{id}") // nil db — guard fires before tx open
	_, _, err := s.BeginReplay(context.Background(), "00000000-0000-4000-8000-0000000000bb", "", uniqueKey("empty-actor"), "hash-a")
	if err == nil {
		t.Fatal("expected error for empty actorID, got nil")
	}
	if !strings.Contains(err.Error(), "actorID") {
		t.Fatalf("error should mention actorID, got: %v", err)
	}
}

// TestBeginReplay_EmptyTenantReturnsError proves the Go-layer guard rejects
// an empty tenantID before any DB round-trip, symmetric with
// TestBeginReplay_EmptyActorReturnsError above (#90/A3.5). This is the
// persistence-boundary half of the fix for the A3.3-deferred defect
// (PR #108 review thread, commit 66cfb664): even a future caller that bypasses
// the middleware's identity boundary cannot smuggle an empty tenant past this
// store.
func TestBeginReplay_EmptyTenantReturnsError(t *testing.T) {
	s := idempotency.New(nil, "POST /h11/{id}") // nil db — guard fires before tx open
	_, _, err := s.BeginReplay(context.Background(), "", "actor-h11", uniqueKey("empty-tenant"), "hash-a")
	if err == nil {
		t.Fatal("expected error for empty tenantID, got nil")
	}
	if !strings.Contains(err.Error(), "tenantID") {
		t.Fatalf("error should mention tenantID, got: %v", err)
	}
}

// TestBeginReplay_WhitespaceTenantReturnsError proves whitespace-only tenant
// input cannot create the same shared collision slot as an empty tenant.
func TestBeginReplay_WhitespaceTenantReturnsError(t *testing.T) {
	s := idempotency.New(nil, "POST /h11/{id}") // nil db — guard fires before tx open
	_, _, err := s.BeginReplay(context.Background(), strings.Repeat(" ", 3), h11Actor, uniqueKey("whitespace-tenant"), "hash-a")
	if err == nil {
		t.Fatal("expected error for whitespace-only tenantID, got nil")
	}
	if !strings.Contains(err.Error(), "tenantID") {
		t.Fatalf("error should mention tenantID, got: %v", err)
	}
}

// TestCompleteReplay_OversizedBodyReturnsError proves that a body larger than
// maxBodyBytes (64 KiB) is rejected before the DB UPDATE is attempted.
func TestCompleteReplay_OversizedBodyReturnsError(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("oversized")

	handle, _, err := s.BeginReplay(ctx, tenant.ID, h11Actor, key, "hash-a")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}

	big := bytes.Repeat([]byte("x"), 64*1024+1)
	if err := s.CompleteReplay(handle, 200, big); err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}

	// Slot should be released (handle is now closed with rollback). A retry
	// must be able to claim the key again.
	handle2, _, err := s.BeginReplay(ctx, tenant.ID, h11Actor, key, "hash-a")
	if err != nil {
		t.Fatalf("retry after oversized rejection: %v", err)
	}
	_ = s.FailReplay(handle2, nil)
}

// TestCompleteReplay_NonJSONBodyRoundTrips proves that BYTEA stores arbitrary
// bytes faithfully — no JSON coercion that would fail on binary or plain-text
// response bodies.
func TestCompleteReplay_NonJSONBodyRoundTrips(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("nonjson")

	// Non-JSON body: plain-text with bytes that would fail JSON parsing.
	body := []byte("OK\r\nstatus: accepted\x00binary\xff")

	handle, _, err := s.BeginReplay(ctx, tenant.ID, h11Actor, key, "hash-bin")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if err := s.CompleteReplay(handle, 200, body); err != nil {
		t.Fatalf("CompleteReplay with non-JSON body: %v", err)
	}

	_, replay, err := s.BeginReplay(ctx, tenant.ID, h11Actor, key, "hash-bin")
	if err != nil {
		t.Fatalf("second BeginReplay: %v", err)
	}
	if replay == nil {
		t.Fatal("expected replay hit, got nil")
	}
	if !bytes.Equal(replay.Body, body) {
		t.Fatalf("body mismatch:\n got:  %q\n want: %q", replay.Body, body)
	}
}

// TestCompleteReplay_ExactlyAtCapSucceeds proves the boundary: a body of
// exactly maxBodyBytes (64 KiB) is accepted.
func TestCompleteReplay_ExactlyAtCapSucceeds(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("at-cap")

	body := bytes.Repeat([]byte("a"), 64*1024)

	handle, _, err := s.BeginReplay(ctx, tenant.ID, h11Actor, key, "hash-cap")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if err := s.CompleteReplay(handle, 201, body); err != nil {
		t.Fatalf("CompleteReplay at exact cap: %v", err)
	}
}
