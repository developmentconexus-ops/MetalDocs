package idempotency_test

// H11 regression tests: actor_user_id non-empty, response_body BYTEA + 64 KiB cap.

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

const (
	h11Tenant = "00000000-0000-4000-8000-0000000000bb"
	h11Actor  = "actor-h11"
)

// TestBeginReplay_EmptyActorReturnsError proves the Go-layer guard rejects
// an empty actorID before any DB round-trip.
func TestBeginReplay_EmptyActorReturnsError(t *testing.T) {
	s := idempotency.New(nil, "POST /h11/{id}") // nil db — guard fires before tx open
	_, _, err := s.BeginReplay(context.Background(), h11Tenant, "", uniqueKey("empty-actor"), "hash-a")
	if err == nil {
		t.Fatal("expected error for empty actorID, got nil")
	}
	if !strings.Contains(err.Error(), "actorID") {
		t.Fatalf("error should mention actorID, got: %v", err)
	}
}

// TestCompleteReplay_OversizedBodyReturnsError proves that a body larger than
// maxBodyBytes (64 KiB) is rejected before the DB UPDATE is attempted.
func TestCompleteReplay_OversizedBodyReturnsError(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("oversized")

	handle, _, err := s.BeginReplay(ctx, h11Tenant, h11Actor, key, "hash-a")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}

	big := bytes.Repeat([]byte("x"), 64*1024+1)
	if err := s.CompleteReplay(handle, 200, big); err == nil {
		t.Fatal("expected error for oversized body, got nil")
	}

	// Slot should be released (handle is now closed with rollback). A retry
	// must be able to claim the key again.
	handle2, _, err := s.BeginReplay(ctx, h11Tenant, h11Actor, key, "hash-a")
	if err != nil {
		t.Fatalf("retry after oversized rejection: %v", err)
	}
	_ = s.FailReplay(handle2, nil)
}

// TestRecordReplay_OversizedBodyReturnsError proves the legacy path also
// enforces the cap.
func TestRecordReplay_OversizedBodyReturnsError(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()

	big := bytes.Repeat([]byte("z"), 64*1024+1)
	err := s.RecordReplay(ctx, h11Tenant, h11Actor, uniqueKey("rec-oversized"), "hash-a", 200, big)
	if err == nil {
		t.Fatal("expected error for oversized body in RecordReplay, got nil")
	}
}

// TestCompleteReplay_NonJSONBodyRoundTrips proves that BYTEA stores arbitrary
// bytes faithfully — no JSON coercion that would fail on binary or plain-text
// response bodies.
func TestCompleteReplay_NonJSONBodyRoundTrips(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("nonjson")

	// Non-JSON body: plain-text with bytes that would fail JSON parsing.
	body := []byte("OK\r\nstatus: accepted\x00binary\xff")

	handle, _, err := s.BeginReplay(ctx, h11Tenant, h11Actor, key, "hash-bin")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if err := s.CompleteReplay(handle, 200, body); err != nil {
		t.Fatalf("CompleteReplay with non-JSON body: %v", err)
	}

	_, replay, err := s.BeginReplay(ctx, h11Tenant, h11Actor, key, "hash-bin")
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
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /h11/{id}")
	ctx := context.Background()
	key := uniqueKey("at-cap")

	body := bytes.Repeat([]byte("a"), 64*1024)

	handle, _, err := s.BeginReplay(ctx, h11Tenant, h11Actor, key, "hash-cap")
	if err != nil {
		t.Fatalf("BeginReplay: %v", err)
	}
	if err := s.CompleteReplay(handle, 201, body); err != nil {
		t.Fatalf("CompleteReplay at exact cap: %v", err)
	}
}
