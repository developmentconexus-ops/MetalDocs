package domain

import "testing"

func TestValuesHash_OrderIndependent(t *testing.T) {
	a := map[string]any{"p1": "x", "p2": "y"}
	b := map[string]any{"p2": "y", "p1": "x"}
	hashA, err := ComputeValuesHash(a)
	if err != nil {
		t.Fatalf("ComputeValuesHash(a): %v", err)
	}
	hashB, err := ComputeValuesHash(b)
	if err != nil {
		t.Fatalf("ComputeValuesHash(b): %v", err)
	}
	if hashA != hashB {
		t.Fatal("hash must be order-independent")
	}
}

func TestValuesHash_ChangesOnValueChange(t *testing.T) {
	hashX, err := ComputeValuesHash(map[string]any{"p1": "x"})
	if err != nil {
		t.Fatalf("ComputeValuesHash(x): %v", err)
	}
	hashY, err := ComputeValuesHash(map[string]any{"p1": "y"})
	if err != nil {
		t.Fatalf("ComputeValuesHash(y): %v", err)
	}
	if hashX == hashY {
		t.Fatal("hash must differ on value change")
	}
}

func TestValuesHash_ReturnsMarshalError(t *testing.T) {
	if _, err := ComputeValuesHash(map[string]any{"bad": func() {}}); err == nil {
		t.Fatal("expected marshal error")
	}
}
