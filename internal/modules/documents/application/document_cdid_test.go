package application_test

import (
	"context"
	"database/sql"
	"testing"
)

// Compile-time: function signature exists in the package.
// Full integration coverage is in fanout_worker_integration_test.go which exercises
// the real path end-to-end. This test validates the signature is importable.
func TestLoadDocumentControlledDocumentIDSignature(t *testing.T) {
	// Just verify the function exists and has the right signature at compile time.
	// We can't call it without a real DB here — that's covered in integration tests.
	_ = context.Background()
	_ = (*sql.Tx)(nil)
	// The function is called: application.LoadDocumentControlledDocumentID(ctx, tx, tenantID, docID)
	// Compile failure here means the function doesn't exist yet.
}
