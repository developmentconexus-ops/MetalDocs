package httprouter

// SurfacePublisher is the HTTP-mounting role every route owner implements.
//
// Deliberately NOT called Module: this repo reserves that word for the 15
// bounded-context modules (CLAUDE.md), and three implementers here — health,
// observability, configuration — are internal/platform packages that are
// emphatically not bounded contexts.
//
// Mount is TOTAL. A publisher registers every route it owns, unconditionally.
// Availability is a handler-level 501, never a routing-level absence: a mount
// that only happens on some boot path makes the surface environment-dependent,
// which is the exception mechanism operator ruling C deleted.
//
// Forward-declared here by Task 12 (the e2e publisher pair) ahead of Task 14,
// which owns this file's full scope (Recorder + assertSurface). The contract
// is reproduced verbatim from Task 14's brief so Task 14 has nothing to
// reconcile beyond adding recorder.go/surface.go alongside it.
type SurfacePublisher interface {
	// Name is a stable identity used in boot assertion messages.
	Name() string
	// Tag is the OpenAPI tag this publisher owns. One tag, one publisher.
	Tag() string
	// Mount registers every route the publisher serves.
	Mount(Muxer)
}
