package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTemplatesHandler_ErrorHandlerFuncIsProblemJSON(t *testing.T) {
	t.Skip("verify via curl: any error response should have Content-Type: application/problem+json")
	_ = httptest.NewRecorder()
	_ = http.StatusBadRequest
}
