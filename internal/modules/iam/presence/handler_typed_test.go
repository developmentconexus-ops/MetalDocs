package presence

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamapi "metaldocs/internal/modules/iam/api"
	"metaldocs/internal/platform/tenant"
)

// TestHandler_Snapshot_TypedBody is a parity-lock characterization test for
// F8.1: the snapshot route emits the generated iamapi.PresenceSnapshotResponse
// typed body (not a map[string]any literal). It strict-decodes the 200 body
// into the generated type with DisallowUnknownFields and asserts every item's
// fields round-trip and that `status` is always present (the generated
// OnlinePresenceItem.Status is *omitempty — a nil pointer would silently drop
// it). Wire-equivalent, not byte-identical: per-item key order follows the
// struct, which is irrelevant per RFC 8259 + the OpenAPI contract (M7 F7.1).
func TestHandler_Snapshot_TypedBody(t *testing.T) {
	repo := newFakeRepo()
	now := time.Now().UTC()
	repo.SetSnapshot("tenant-a", []Item{
		{UserID: "u1", Username: "alice", DisplayName: "Alice", LastSeenAt: now.Add(-5 * time.Second)},
		{UserID: "u2", Username: "bob", DisplayName: "Bob", LastSeenAt: now.Add(-2 * time.Minute)},
	})

	hub := NewHub(repo, nil)
	h := NewHandler(hub, repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/iam/presence/snapshot", nil).
		WithContext(tenant.WithTenantID(context.Background(), "tenant-a"))
	rec := httptest.NewRecorder()
	h.handleSnapshot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200; body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: got %q want application/json", ct)
	}

	// Strict-decode into the generated typed response: any unknown/dropped
	// field fails here, pinning the typed body to the contract.
	dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
	dec.DisallowUnknownFields()
	var got iamapi.PresenceSnapshotResponse
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("strict-decode into PresenceSnapshotResponse: %v; body=%s", err, rec.Body.String())
	}

	if len(got.Items) != 2 {
		t.Fatalf("items: got %d want 2; body=%s", len(got.Items), rec.Body.String())
	}
	// Snapshot is ordered most-recent-first (fakeRepo sorts by LastSeenAt desc).
	first := got.Items[0]
	if first.UserId != "u1" || first.Username != "alice" || first.DisplayName != "Alice" {
		t.Fatalf("item[0] fields: got %+v", first)
	}
	if !first.LastSeenAt.Equal(now.Add(-5 * time.Second)) {
		t.Fatalf("item[0] last_seen_at: got %v want %v", first.LastSeenAt, now.Add(-5*time.Second))
	}
	// status is *omitempty in the generated model — it MUST be present (online).
	if first.Status == nil {
		t.Fatalf("item[0] status: nil pointer dropped the field from the wire")
	}
	if string(*first.Status) != string(StatusOnline) {
		t.Fatalf("item[0] status: got %q want %q", *first.Status, StatusOnline)
	}
	if got.Items[1].Status == nil || string(*got.Items[1].Status) != string(StatusIdle) {
		t.Fatalf("item[1] status: got %v want idle", got.Items[1].Status)
	}
}
