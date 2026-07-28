package idempotency_test

import (
	"context"
	"encoding/json"
	"testing"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/infrastructure/idempotency"
)

func TestPostgresSignoffIdempStore_NilDB(t *testing.T) {
	store := idempotency.NewPostgresSignoffIdempStore(nil)

	_, _, err := store.BeginDocumentReplay(context.Background(), "tenant", "actor", "key", "hash")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}

	_, _, err = store.BeginStageReplay(context.Background(), "tenant", "actor", "key", "hash")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}

	_, _, err = store.BeginFastForwardReplay(context.Background(), "tenant", "actor", "key", "hash")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}
}

// TestSignoffReplayEnvelope_CarriesSignoffID pins the persisted replay body
// shape (F-QA4-7): SignoffReplayHandle.Complete marshals this struct into
// idempotency_keys.response_body, and beginReplay unmarshals it back, so the
// signoff id only survives an Idempotency-Key retry if it is on the wire form.
func TestSignoffReplayEnvelope_CarriesSignoffID(t *testing.T) {
	body, err := json.Marshal(application.SignoffReplay{Outcome: "approved", SignoffID: "sig-1"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got, want := string(body), `{"outcome":"approved","signoff_id":"sig-1"}`; got != want {
		t.Fatalf("persisted body = %s, want %s", got, want)
	}

	var round application.SignoffReplay
	if err := json.Unmarshal(body, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if round.SignoffID != "sig-1" || round.Outcome != "approved" {
		t.Fatalf("round-tripped envelope = %+v", round)
	}

	// Backward compatibility: envelopes persisted before F-QA4-7 carry only the
	// outcome. They must still decode (empty id = pre-fix behaviour), never error.
	var legacy application.SignoffReplay
	if err := json.Unmarshal([]byte(`{"outcome":"approved"}`), &legacy); err != nil {
		t.Fatalf("legacy envelope must still decode: %v", err)
	}
	if legacy.SignoffID != "" || legacy.Outcome != "approved" {
		t.Fatalf("legacy envelope = %+v, want empty signoff_id", legacy)
	}
}
