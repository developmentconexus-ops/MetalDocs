package idempotency_test

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/testsupport/pgtest"
)

func TestStore_RecordAndReplay_Hit(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v2/test/{id}")
	ctx := context.Background()
	body := []byte(`{"ok":true}`)
	err := s.RecordReplay(ctx, "tenant-1", "actor-1", "key-1", "hash-a", 201, body)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	replay, err := s.CheckReplay(ctx, "tenant-1", "actor-1", "key-1", "hash-a")
	if err != nil || replay == nil {
		t.Fatalf("replay miss: %v", err)
	}
	if replay.Status != 201 || string(replay.Body) != string(body) {
		t.Fatalf("replay payload mismatch: %+v", replay)
	}
}

func TestStore_CheckReplay_Conflict(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v2/test/{id}")
	ctx := context.Background()
	_ = s.RecordReplay(ctx, "tenant-1", "actor-1", "key-2", "hash-a", 201, []byte(`{}`))
	_, err := s.CheckReplay(ctx, "tenant-1", "actor-1", "key-2", "hash-b")
	if !errors.Is(err, idempotency.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestStore_CheckReplay_Miss(t *testing.T) {
	db := pgtest.OpenAndMigrate(t)
	s := idempotency.New(db, "POST /api/v2/test/{id}")
	replay, err := s.CheckReplay(context.Background(), "tenant-x", "actor-x", "missing", "h")
	if err != nil || replay != nil {
		t.Fatalf("expected nil replay, got %+v err=%v", replay, err)
	}
}
