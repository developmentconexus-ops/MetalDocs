// Package httprouter defines the minimal http.ServeMux surface delivery
// handlers depend on for route registration.
package httprouter

import "net/http"

// Muxer is the *http.ServeMux surface delivery handlers rely on: plain
// Handle/HandleFunc registration (legacy handlers) and the HandleFunc +
// http.Handler surface oapi-codegen's generated per-module ServeMux
// interface expects (StdHTTPServerOptions.BaseRouter). *http.ServeMux
// satisfies this today with zero changes; a recording wrapper used by
// TestRouteCoverage (apps/api/cmd/metaldocs-api/permissions_test.go)
// satisfies it too, letting the test observe every pattern the real server
// mounts instead of relying on a hand-maintained fixture list.
type Muxer interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
