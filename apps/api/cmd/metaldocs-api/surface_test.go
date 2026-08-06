package main

import (
	"net/http"
	"strings"
	"testing"

	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	"metaldocs/internal/platform/httprouter"
)

type fakePublisher struct{ name, tag string }

func (f fakePublisher) Name() string           { return f.name }
func (f fakePublisher) Tag() string            { return f.tag }
func (f fakePublisher) Mount(httprouter.Muxer) {}

func rule(tag string) surfaceRule {
	return surfaceRule{visibility: iamdelivery.VisibilityPermissionGuarded, tag: tag}
}

// Check 1 — tag coverage. A publisher constructed but never listed leaves its
// tag unclaimed.
func TestAssertSurface_TagWithNoPublisher(t *testing.T) {
	err := assertSurface(
		map[string][]string{"docs": {"GET /api/v1/d"}},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents"), "GET /api/v1/t": rule("templates")},
		[]string{"documents", "templates"},
		[]httprouter.SurfacePublisher{fakePublisher{"docs", "documents"}},
	)
	if err == nil || !strings.Contains(err.Error(), "templates") {
		t.Fatalf("want the unclaimed tag named, got %v", err)
	}
}

// Check 1 — the other half: two publishers claiming one tag.
func TestAssertSurface_TwoPublishersOneTag(t *testing.T) {
	err := assertSurface(
		map[string][]string{"a": {"GET /api/v1/d"}, "b": {}},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents")},
		[]string{"documents"},
		[]httprouter.SurfacePublisher{fakePublisher{"a", "documents"}, fakePublisher{"b", "documents"}},
	)
	if err == nil || !strings.Contains(err.Error(), "documents") {
		t.Fatalf("want the duplicated tag named, got %v", err)
	}
}

// Check 2 — mounted ⊆ declared. A route mounted with no policy.
func TestAssertSurface_MountedWithNoDeclaration(t *testing.T) {
	err := assertSurface(
		map[string][]string{"docs": {"GET /api/v1/d", "GET /api/v1/undeclared"}},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents")},
		[]string{"documents"},
		[]httprouter.SurfacePublisher{fakePublisher{"docs", "documents"}},
	)
	if err == nil || !strings.Contains(err.Error(), "GET /api/v1/undeclared") {
		t.Fatalf("want the undeclared pattern named, got %v", err)
	}
}

// Check 3 — declared ⊆ mounted. A spec operation whose handler was never wired.
func TestAssertSurface_DeclaredButNotMounted(t *testing.T) {
	err := assertSurface(
		map[string][]string{"docs": {"GET /api/v1/d"}},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents"), "GET /api/v1/orphan": rule("documents")},
		[]string{"documents"},
		[]httprouter.SurfacePublisher{fakePublisher{"docs", "documents"}},
	)
	if err == nil || !strings.Contains(err.Error(), "GET /api/v1/orphan") {
		t.Fatalf("want the unmounted declaration named, got %v", err)
	}
}

// Check 4 — ownership. NOT redundant with check 1: recording one global
// pattern set would let every publisher claim a distinct tag while mounting
// each other's operations, and checks 1-3 would all pass.
func TestAssertSurface_PublisherMountsAnotherTagsRoute(t *testing.T) {
	err := assertSurface(
		map[string][]string{
			"docs": {"GET /api/v1/d", "GET /api/v1/t"}, // mounts a templates route
			"tpl":  {},
		},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents"), "GET /api/v1/t": rule("templates")},
		[]string{"documents", "templates"},
		[]httprouter.SurfacePublisher{fakePublisher{"docs", "documents"}, fakePublisher{"tpl", "templates"}},
	)
	if err == nil || !strings.Contains(err.Error(), "GET /api/v1/t") {
		t.Fatalf("want the cross-mounted pattern named, got %v", err)
	}
}

func TestAssertSurface_HappyPath(t *testing.T) {
	err := assertSurface(
		map[string][]string{"docs": {"GET /api/v1/d"}, "tpl": {"GET /api/v1/t"}},
		map[string]surfaceRule{"GET /api/v1/d": rule("documents"), "GET /api/v1/t": rule("templates")},
		[]string{"documents", "templates"},
		[]httprouter.SurfacePublisher{fakePublisher{"docs", "documents"}, fakePublisher{"tpl", "templates"}},
	)
	if err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestMergedSurfacePanicsOnCollision(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic on a colliding key; rule 6 should have caught it at build time")
		}
	}()
	mergedSurface(
		map[string]surfaceRule{"GET /x": rule("a")},
		map[string]surfaceRule{"GET /x": rule("b")},
	)
}

// testPublisherDeps builds a real publisherDeps instance for the boot
// assertion test below, reusing buildTestRouteHandlers (permissions_test.go)
// — the same nil-service construction TestGeneratedKeysMatchCodegenPatterns
// and mountRoutesWithNilPresence already use. Registration hands method
// *values* to the mux and never invokes them, so nil inner services are safe
// at mount time.
func testPublisherDeps(t *testing.T) publisherDeps {
	t.Helper()
	return buildTestRouteHandlers()
}

// TestRealPublishersSatisfyTheAssertion is the boot assertion run in a test,
// against the real publisher list built from real (nil-service) handlers. It
// is the same call main() makes: one httprouter.Recorder per publisher,
// fed to assertSurface against the generated httpSurface/specTags tables.
func TestRealPublishersSatisfyTheAssertion(t *testing.T) {
	mux := http.NewServeMux()
	pubs := buildPublishers(testPublisherDeps(t)) // nil services are safe at mount time

	mounted := map[string][]string{}
	for _, p := range pubs {
		rec := httprouter.NewRecorder(mux)
		p.Mount(rec)
		mounted[p.Name()] = rec.Patterns()
	}
	if err := assertSurface(mounted, httpSurface, specTags, pubs); err != nil {
		t.Fatalf("the real surface does not satisfy its own declaration:\n%v", err)
	}
}
