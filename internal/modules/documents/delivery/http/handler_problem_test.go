package http_test

import (
	"testing"
)

func TestDocumentsHandler_NotFoundIsProblemJSON(t *testing.T) {
	t.Skip("verify via curl after implementation: Content-Type: application/problem+json")
}

func TestDocumentsHandler_MapErrIsProblemJSON(t *testing.T) {
	t.Skip("see integration test below")
}
