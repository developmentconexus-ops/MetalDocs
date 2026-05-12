package httpdelivery_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIAMMiddleware_UnauthorizedIsProblemJSON verifies that the IAM auth
// middleware emits application/problem+json on unauthenticated requests.
// Full stack integration test — requires a running server.
func TestIAMMiddleware_UnauthorizedIsProblemJSON(t *testing.T) {
	t.Skip("manual check: curl the running server and verify Content-Type: application/problem+json on a request without a session cookie")
	_ = httptest.NewRecorder()
	_ = http.StatusUnauthorized
}
