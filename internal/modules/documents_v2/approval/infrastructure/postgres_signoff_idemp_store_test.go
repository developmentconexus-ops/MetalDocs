package infrastructure_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents_v2/approval/infrastructure"
)

func TestPostgresSignoffIdempStore_NilDB(t *testing.T) {
	store := infrastructure.NewPostgresSignoffIdempStore(nil)

	_, _, err := store.CheckReplay(context.Background(), "tenant", "actor", "key")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}

	err = store.RecordReplay(context.Background(), "tenant", "actor", "key", "approved")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}
}
