package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"

	"gopkg.in/yaml.v3"
)

// routePath normalizes a spec path key to the runtime route path. The spec mixes
// conventions: some keys carry the /api/v1 base (servers prefix) inline, others
// are base-relative. Runtime routeRules use the full /api/v1 path, so ensure the
// base is present exactly once.
func routePath(specPath string) string {
	if strings.HasPrefix(specPath, "/api/v1") {
		return specPath
	}
	return "/api/v1" + specPath
}

// specOp is one OpenAPI operation with its route coordinates and x-authz markers.
type specOp struct {
	opID, method, path string
	hasArea, hasSkip   bool
	skipReason         string
}

// loadSpecOps parses the OpenAPI doc into a flat list of operations carrying the
// HTTP method + path (so each can be matched against the runtime route table)
// and the x-authz markers.
func loadSpecOps(t *testing.T) []specOp {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate spec")
	}
	// apps/api/cmd/metaldocs-api/ -> repo root is four levels up.
	specPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..",
		"api", "openapi", "v1", "openapi.yaml")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec %s: %v", specPath, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	root := doc.Content[0]
	paths := yamlMapGet(root, "paths")
	if paths == nil {
		t.Fatal("spec missing paths")
	}

	out := []specOp{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathKey := paths.Content[i].Value
		pathVal := paths.Content[i+1]
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			method := pathVal.Content[j].Value
			op := pathVal.Content[j+1]
			opID := yamlScalar(yamlMapGet(op, "operationId"))
			if opID == "" {
				continue
			}
			out = append(out, specOp{
				opID:       opID,
				method:     strings.ToUpper(method),
				path:       pathKey,
				hasArea:    yamlMapGet(op, "x-authz-area") != nil,
				hasSkip:    yamlMapGet(op, "x-authz-skip-area") != nil,
				skipReason: yamlScalar(yamlMapGet(op, "x-authz-skip-reason")),
			})
		}
	}
	return out
}

// TestAreaEnforcedOpsAnnotated is self-maintaining (ADR 0022 Phase 7): instead of
// a hand-curated op list, it DERIVES the obligation from the source of truth —
// the set of capabilities classified IsAreaGrade. For every area-grade capability
// it requires that the spec carries at least one route-matched operation
// declaring an area posture (x-authz-area for a request-supplied area, or a
// justified x-authz-skip-area for a DB/payload-derived area). Consequence: when a
// capability is flipped to area-grade (or a new one is added) and its tier-2
// enforcement gap is closed, the build fails until its HTTP surface is annotated
// — closing the gap forces the declaration.
//
// The obligation is cap-level, not op-level, on purpose: a capability's area
// posture is a property of the capability, and not every operation that carries
// an area-grade TIER-1 route cap performs a TIER-2 area check (e.g. a session
// heartbeat refresh). Annotating every such op "area enforced" would assert a
// posture that doesn't exist; the AST guard (authz-area-scope-binding) is the
// per-call-site binding that statically forbids an area-grade cap from being
// enforced with the literal "tenant". The server base path is /api/v1.
func TestAreaEnforcedOpsAnnotated(t *testing.T) {
	t.Parallel()

	// Collect, per area-grade capability, whether it has any documented HTTP
	// surface (a route-matched spec op) and whether any such op declares an area
	// posture. Flag any annotated-but-unjustified skip.
	hasSurface := map[iamdomain.Capability]bool{}
	annotatedCap := map[iamdomain.Capability]bool{}
	for _, o := range loadSpecOps(t) {
		cap, _, ok := resolveRoutePermission(o.method, routePath(o.path))
		if !ok || !iamdomain.IsAreaGrade(cap) {
			continue
		}
		hasSurface[cap] = true
		if o.hasSkip && o.skipReason == "" {
			t.Errorf("op %q has x-authz-skip-area but no x-authz-skip-reason — a skip must be justified (ADR 0022).", o.opID)
		}
		if o.hasArea || o.hasSkip {
			annotatedCap[cap] = true
		}
	}

	// Every area-grade capability that HAS a documented HTTP surface must declare
	// an area posture on it. Caps with no spec surface (e.g. document.create — its
	// POST /documents route is a raw, undocumented handler) cannot be annotated in
	// the spec; those are bound by the authz-area-scope-binding AST guard + tests,
	// which is the stronger per-call-site enforcement.
	areaGradeSeen := 0
	for _, cap := range iamdomain.AllCapabilities() {
		if !iamdomain.IsAreaGrade(cap) {
			continue
		}
		areaGradeSeen++
		if hasSurface[cap] && !annotatedCap[cap] {
			t.Errorf("area-grade capability %q has a documented HTTP surface but no operation declaring x-authz-area or a justified x-authz-skip-area (ADR 0022 Phase 7: closing a tier-2 area-enforcement gap forces the spec declaration).", cap)
		}
	}
	if areaGradeSeen == 0 {
		t.Fatal("no area-grade capabilities found — IsAreaGrade wiring is broken")
	}
}

// TestMembershipManageIsAreaGradeAndAnnotated binds the typed capabilityScopes
// classification to the spec for the one IAM area surface that carries a
// request-supplied area: membership.manage MUST be area-grade in the registry,
// and its grant/revoke ops MUST be annotated. This ties the Go classification to
// the OpenAPI declaration so the two cannot drift apart.
func TestMembershipManageIsAreaGradeAndAnnotated(t *testing.T) {
	t.Parallel()

	if !iamdomain.IsAreaGrade(iamdomain.CapMembershipManage) {
		t.Fatalf("membership.manage must be classified area-grade in capabilityScopes (ADR 0022)")
	}
	byID := map[string]specOp{}
	for _, o := range loadSpecOps(t) {
		byID[o.opID] = o
	}
	for _, opID := range []string{"grantAreaMembership", "revokeAreaMembership"} {
		if !byID[opID].hasArea {
			t.Errorf("membership op %q must carry x-authz-area (the request areaCode source) — it is the request-area-bearing area-grade surface.", opID)
		}
	}
}

// yamlMapGet / yamlScalar are local yaml.Node helpers (the spec_rules helpers
// live in the api-lint main package, not importable here).
func yamlMapGet(n *yaml.Node, key string) *yaml.Node {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

func yamlScalar(n *yaml.Node) string {
	if n == nil {
		return ""
	}
	return n.Value
}
