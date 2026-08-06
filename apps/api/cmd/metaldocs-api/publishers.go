package main

import (
	"metaldocs/internal/platform/httprouter"
)

// publisherDeps collects every httprouter.SurfacePublisher the composition
// root constructs. It is the input to buildPublishers — NOT the thing the
// boot assertion audits. The audited artifact is the LIST buildPublishers
// returns; this struct exists only so main() (and testPublisherDeps in
// surface_test.go) has named fields to assign into, mirroring the shape
// router.go's routeHandlers used to have.
type publisherDeps struct {
	auth                httprouter.SurfacePublisher
	health              httprouter.SurfacePublisher
	observability       httprouter.SurfacePublisher
	featureFlags        httprouter.SurfacePublisher
	audit               httprouter.SurfacePublisher
	search              httprouter.SurfacePublisher
	security            httprouter.SurfacePublisher
	taxonomy            httprouter.SurfacePublisher
	tokens              httprouter.SurfacePublisher
	controlledDocuments httprouter.SurfacePublisher
	iam                 httprouter.SurfacePublisher
	documents           httprouter.SurfacePublisher
	templates           httprouter.SurfacePublisher
	approval            httprouter.SurfacePublisher
	distribution        httprouter.SurfacePublisher
	notifications       httprouter.SurfacePublisher
}

// buildPublishers is the composition root's route inventory. It is a LIST,
// not a struct: a forgotten publisher is a missing list element and §5's
// tag-coverage assertion (assertSurface, check 1) fires at boot. There is no
// field to leave unset — which is what closes main.go's former keyed-literal
// hole (routeHandlers{...}) structurally, not by convention.
//
// 16 publishers, one per OpenAPI tag (specTags), matching Task 15a's actual
// owner set after it split health/observability into two publishers and
// folded presence into iam. No module imports iamdomain to declare its own
// capability: capability is IAM's published concept. Publishers declare
// routes; the root declares policy (expectedTags/surface in main.go, derived
// only from useE2E — never from this list).
func buildPublishers(d publisherDeps) []httprouter.SurfacePublisher {
	return []httprouter.SurfacePublisher{
		d.auth, d.health, d.observability, d.featureFlags, d.audit, d.search,
		d.security, d.taxonomy, d.tokens, d.controlledDocuments, d.iam,
		d.documents, d.templates, d.approval, d.distribution, d.notifications,
	}
}
