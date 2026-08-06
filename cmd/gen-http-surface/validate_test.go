package main

import (
	"strings"
	"testing"
)

func reg() map[string]string {
	return map[string]string{"metrics.view": "CapMetricsView", "document.view": "CapDocumentView"}
}

func docWith(ops ...Operation) Document {
	return Document{Path: "test.yaml", ServerBase: "/api/v1", Operations: ops}
}

func errText(errs []error) string {
	parts := make([]string, 0, len(errs))
	for _, e := range errs {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, "\n")
}

func TestRule1_NoMarker(t *testing.T) {
	errs := Validate([]Document{docWith(Operation{
		ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents",
	})}, reg())
	if !strings.Contains(errText(errs), "getThing") {
		t.Fatalf("want getThing named in error, got: %s", errText(errs))
	}
}

func TestRule2_TwoMarkers(t *testing.T) {
	errs := Validate([]Document{docWith(Operation{
		ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents",
		Capability: "document.view", Public: true,
	})}, reg())
	if !strings.Contains(errText(errs), "getThing") {
		t.Fatalf("want getThing named, got: %s", errText(errs))
	}
}

func TestRule3_UnknownCapability(t *testing.T) {
	errs := Validate([]Document{docWith(Operation{
		ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents",
		Capability: "document.teleport",
	})}, reg())
	if !strings.Contains(errText(errs), "document.teleport") {
		t.Fatalf("want the unknown capability named, got: %s", errText(errs))
	}
}

func TestRule4_BlankNoneReason(t *testing.T) {
	errs := Validate([]Document{docWith(Operation{
		ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents",
		CapabilityNone: "   ",
	})}, reg())
	if !strings.Contains(errText(errs), "getThing") {
		t.Fatalf("want getThing named, got: %s", errText(errs))
	}
}

func TestRule5_TwoTags(t *testing.T) {
	errs := Validate([]Document{docWith(Operation{
		ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents,templates",
		Capability: "document.view",
	})}, reg())
	if !strings.Contains(errText(errs), "getThing") {
		t.Fatalf("want getThing named, got: %s", errText(errs))
	}
}

// Rule 6 is the only cross-document rule, which is why the generator takes
// both documents in ONE invocation: a rule comparing two documents cannot be
// checked by a run that has only ever seen one.
func TestRule6_CrossDocumentCollision(t *testing.T) {
	public := Document{Path: "openapi.yaml", ServerBase: "/api/v1", Operations: []Operation{
		{ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents", Capability: "document.view"},
	}}
	e2e := Document{Path: "internal-e2e.yaml", ServerBase: "", Operations: []Operation{
		{ID: "seedThing", Method: "GET", Path: "/api/v1/thing", Tag: "internal-e2e", Public: true},
	}}
	errs := Validate([]Document{public, e2e}, reg())
	if !strings.Contains(errText(errs), "GET /api/v1/thing") {
		t.Fatalf("want the colliding key named, got: %s", errText(errs))
	}
}

// The server base is read FROM the document, never hard-coded: internal-e2e
// declares an empty base and mounts at the server root.
func TestServerBaseMustPrefixOwnPaths(t *testing.T) {
	// The real malformation: a server base that is neither absolute nor empty.
	// A concatenation check is tautologically true for any base, so it can
	// never catch this; the rule instead validates the declared base's shape
	// directly.
	bad := Document{Path: "bad.yaml", ServerBase: "api/v2", Operations: []Operation{
		{ID: "getThing", Method: "GET", Path: "/thing", Tag: "documents", Capability: "document.view"},
	}}
	errs := Validate([]Document{bad}, reg())
	if len(errs) == 0 {
		t.Fatal("want a base-prefix error")
	}
}

func TestKeyFormula(t *testing.T) {
	pub := Document{ServerBase: "/api/v1"}
	if got := Key(pub, Operation{Method: "GET", Path: "/metrics"}); got != "GET /api/v1/metrics" {
		t.Fatalf("public key = %q", got)
	}
	e2e := Document{ServerBase: ""}
	if got := Key(e2e, Operation{Method: "POST", Path: "/internal/test/seed"}); got != "POST /internal/test/seed" {
		t.Fatalf("e2e key = %q", got)
	}
}
