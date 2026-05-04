package resolvers

import (
	"bytes"
	"context"
	"testing"
)

func TestControlledByAreaResolver_Resolve(t *testing.T) {
	r := ControlledByAreaResolver{}
	in := ResolveInput{
		TenantID:         "tenant-a",
		AreaCodeSnapshot: "ENG",
	}

	v1, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}

	if v1.Value != "ENG" {
		t.Fatalf("expected area code ENG, got %#v", v1.Value)
	}
	if !bytes.Equal(v1.InputsHash, v2.InputsHash) {
		t.Fatal("expected stable hash across repeated resolves")
	}
}

func TestControlledByAreaResolver_ReturnsAreaName(t *testing.T) {
	r := ControlledByAreaResolver{}
	in := ResolveInput{
		TenantID:         "t1",
		AreaCodeSnapshot: "QA-01",
		AreaNameSnapshot: "Quality Assurance — Area 1",
	}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := out.Value.(string); got != "Quality Assurance — Area 1" {
		t.Fatalf("want area name, got %q", got)
	}
	if out.ResolverVer != 2 {
		t.Fatalf("want version 2, got %d", out.ResolverVer)
	}
}

func TestControlledByAreaResolver_FallsBackToCode(t *testing.T) {
	r := ControlledByAreaResolver{}
	in := ResolveInput{TenantID: "t1", AreaCodeSnapshot: "QA-01"}
	out, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got := out.Value.(string); got != "QA-01" {
		t.Fatalf("want fallback to code, got %q", got)
	}
}
