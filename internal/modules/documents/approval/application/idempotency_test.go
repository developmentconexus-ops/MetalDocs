package application

import (
	"testing"
	"time"
)

func TestIdempotencyDoubleClickSameSecond(t *testing.T) {
	// Fixed time anchored at the start of a second — immune to second-boundary race.
	fixedBase := time.Date(2026, 1, 15, 10, 30, 42, 0, time.UTC)
	base := IdempotencyInput{
		ActorUserID: "u1", DocumentID: "d1", StageInstanceID: "s1",
		Decision: "approve", Timestamp: fixedBase,
	}
	// Sub-second difference → same bucket → same key.
	base2 := base
	base2.Timestamp = base.Timestamp.Add(500 * time.Millisecond)

	k1, err := ComputeIdempotencyKey(base)
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(base): %v", err)
	}
	k2, err := ComputeIdempotencyKey(base2)
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(base2): %v", err)
	}
	if k1 != k2 {
		t.Errorf("same-second timestamps should produce same key; got %s and %s", k1, k2)
	}
}

func TestIdempotencyDifferentSecond(t *testing.T) {
	base := IdempotencyInput{
		ActorUserID: "u1", DocumentID: "d1", StageInstanceID: "s1",
		Decision: "approve", Timestamp: time.Now(),
	}
	base2 := base
	base2.Timestamp = base.Timestamp.Add(2 * time.Second)

	k1, err := ComputeIdempotencyKey(base)
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(base): %v", err)
	}
	k2, err := ComputeIdempotencyKey(base2)
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(base2): %v", err)
	}
	if k1 == k2 {
		t.Error("different seconds should produce different keys")
	}
}

func TestIdempotencyDifferentActor(t *testing.T) {
	ts := time.Now()
	k1, err := ComputeIdempotencyKey(IdempotencyInput{ActorUserID: "u1", DocumentID: "d1", StageInstanceID: "s1", Decision: "approve", Timestamp: ts})
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(k1): %v", err)
	}
	k2, err := ComputeIdempotencyKey(IdempotencyInput{ActorUserID: "u2", DocumentID: "d1", StageInstanceID: "s1", Decision: "approve", Timestamp: ts})
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey(k2): %v", err)
	}
	if k1 == k2 {
		t.Error("different actors should produce different keys")
	}
}

func TestIdempotencyOutputFormat(t *testing.T) {
	k, err := ComputeIdempotencyKey(IdempotencyInput{ActorUserID: "u", DocumentID: "d", StageInstanceID: "s", Decision: "approve", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("ComputeIdempotencyKey: %v", err)
	}
	if len(k) != 64 {
		t.Errorf("key length = %d; want 64", len(k))
	}
}
