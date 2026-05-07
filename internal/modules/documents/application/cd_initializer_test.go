package application_test

import (
	"testing"

	"metaldocs/internal/modules/documents/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
)

// TestCDDocumentInitializer_Interface is a compile-time check that
// *CDDocumentInitializer satisfies the registry DocumentInitializer port.
// Full integration coverage (real DB tx) lives in Phase 6's
// integration_test.go for the registry atomic-create flow.
func TestCDDocumentInitializer_Interface(t *testing.T) {
	var _ registrydomain.DocumentInitializer = (*application.CDDocumentInitializer)(nil)
}
