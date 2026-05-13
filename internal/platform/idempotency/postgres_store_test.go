package idempotency_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

func compactJSON(b []byte) []byte {
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return b
	}
	return buf.Bytes()
}

const (
	testTenantA = "00000000-0000-4000-8000-000000000001"
	testTenantB = "00000000-0000-4000-8000-000000000002"
)

func TestStore_RecordAndReplay_Hit(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v1/test/{id}")
	ctx := context.Background()
	body := []byte(`{"ok":true}`)
	err := s.RecordReplay(ctx, testTenantA, "actor-1", "key-1", "hash-a", 201, body)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	replay, err := s.CheckReplay(ctx, testTenantA, "actor-1", "key-1", "hash-a")
	if err != nil || replay == nil {
		t.Fatalf("replay miss: %v", err)
	}
	if replay.Status != 201 || !bytes.Equal(compactJSON(replay.Body), compactJSON(body)) {
		t.Fatalf("replay payload mismatch: got %s want %s", replay.Body, body)
	}
}

func TestStore_CheckReplay_Conflict(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v1/test/{id}")
	ctx := context.Background()
	_ = s.RecordReplay(ctx, testTenantA, "actor-1", "key-2", "hash-a", 201, []byte(`{}`))
	_, err := s.CheckReplay(ctx, testTenantA, "actor-1", "key-2", "hash-b")
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestStore_CheckReplay_Miss(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v1/test/{id}")
	replay, err := s.CheckReplay(context.Background(), testTenantB, "actor-x", "missing", "h")
	if err != nil || replay != nil {
		t.Fatalf("expected nil replay, got %+v err=%v", replay, err)
	}
}
