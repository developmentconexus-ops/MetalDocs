package application_test

import (
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
)

// TestCDDocumentInitializer_Interface is a compile-time check that
// *CDDocumentInitializer satisfies the controlled-document DocumentInitializer port.
// Full integration coverage (real DB tx) lives in Phase 6's
// integration_test.go for the controlled-document atomic-create flow.
func TestCDDocumentInitializer_Interface(t *testing.T) {
	var _ controlleddocumentsdomain.DocumentInitializer = (*application.CDDocumentInitializer)(nil)
}
