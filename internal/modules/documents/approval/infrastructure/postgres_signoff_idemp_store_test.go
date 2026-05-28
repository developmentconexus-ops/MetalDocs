package infrastructure_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/approval/infrastructure"
)

func TestPostgresSignoffIdempStore_NilDB(t *testing.T) {
	store := infrastructure.NewPostgresSignoffIdempStore(nil)

	_, _, err := store.BeginDocumentReplay(context.Background(), "tenant", "actor", "key", "hash")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}

	_, _, err = store.BeginStageReplay(context.Background(), "tenant", "actor", "key", "hash")
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}
}
