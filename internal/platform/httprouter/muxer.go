// Package httprouter defines the minimal http.ServeMux surface delivery
// handlers depend on for route registration.
package httprouter

import "net/http"

// Muxer is the *http.ServeMux surface delivery handlers rely on: plain
// Handle/HandleFunc registration and the HandleFunc + http.Handler surface
// oapi-codegen's generated per-module ServeMux interface expects
// (StdHTTPServerOptions.BaseRouter). *http.ServeMux satisfies it directly.
//
// The other implementer is Recorder, and it runs in PRODUCTION startup: the
// composition root mounts every SurfacePublisher through one Recorder each and
// feeds what they registered to assertSurface, which refuses to boot on any
// mismatch with the spec-generated surface table. The recorder is the boot-time
// surface recorder, not a test fixture.
type Muxer interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
