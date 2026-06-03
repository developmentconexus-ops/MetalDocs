package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"

	"gopkg.in/yaml.v3"
)

// areaEnforcedOps are the OpenAPI operations whose tier-2 area enforcement ADR
// 0022 declared and annotated: the documents/approval lifecycle (Phase 2) and
// the IAM membership grant/revoke (Phase 3). Each MUST carry x-authz-area or a
// justified x-authz-skip-area so the area posture of an area-grade write is an
// explicit, reviewed declaration — never an accident of which string a call site
// passed to authz.Require.
//
// Excluded by design (ADR 0022 Phase 2 "runtime gap" + Phase 4 notes): the
// document.create / document.edit / controlled_documents.* write ops are
// classified area-grade in capabilityScopes but enforced via the DB-derived
// tx-layer tripwire (area resolved from the row, not the request) and still pass
// the literal "tenant" today. Annotating/realigning them is later-phase work,
// not Phase 5; locking them here would assert a posture that does not yet exist.
var areaEnforcedOps = []string{
	"grantAreaMembership",
	"revokeAreaMembership",
	"submitDocumentForApproval",
	"recordApprovalStageSignoff",
	"recordDocumentSignoff",
	"publishDocument",
	"scheduleDocumentPublish",
	"supersedeDocument",
	"obsoleteDocument",
}

type opAuthz struct {
	found      bool
	hasArea    bool
	hasSkip    bool
	skipReason string
}

// loadSpecAuthz parses the OpenAPI doc and returns operationId -> x-authz markers.
func loadSpecAuthz(t *testing.T) map[string]opAuthz {
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

	out := map[string]opAuthz{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathVal := paths.Content[i+1]
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			op := pathVal.Content[j+1]
			opID := yamlScalar(yamlMapGet(op, "operationId"))
			if opID == "" {
				continue
			}
			skip := yamlMapGet(op, "x-authz-skip-area")
			out[opID] = opAuthz{
				found:      true,
				hasArea:    yamlMapGet(op, "x-authz-area") != nil,
				hasSkip:    skip != nil,
				skipReason: yamlScalar(yamlMapGet(op, "x-authz-skip-reason")),
			}
		}
	}
	return out
}

// TestAreaEnforcedOpsAnnotated locks ADR 0022's area-enforcement declarations:
// every declared area-enforced op carries x-authz-area or a justified
// x-authz-skip-area. Removing an annotation (silently defaulting an area-grade
// write to tenant-wide) fails the build.
func TestAreaEnforcedOpsAnnotated(t *testing.T) {
	t.Parallel()

	spec := loadSpecAuthz(t)
	for _, opID := range areaEnforcedOps {
		a, ok := spec[opID]
		if !ok || !a.found {
			t.Errorf("area-enforced op %q not found in OpenAPI spec — was it renamed? Update areaEnforcedOps / ADR 0022.", opID)
			continue
		}
		if !a.hasArea && !a.hasSkip {
			t.Errorf("area-enforced op %q carries neither x-authz-area nor x-authz-skip-area — an area-grade write must declare its area posture (ADR 0022 Phase 2/3).", opID)
			continue
		}
		if a.hasSkip && a.skipReason == "" {
			t.Errorf("area-enforced op %q has x-authz-skip-area but no x-authz-skip-reason — a skip must be justified.", opID)
		}
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
	spec := loadSpecAuthz(t)
	for _, opID := range []string{"grantAreaMembership", "revokeAreaMembership"} {
		a := spec[opID]
		if !a.hasArea {
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
