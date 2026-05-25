package resolvers

import (
	"bytes"
	"context"
	"testing"
	"time"
)

type stubRevReaderZero struct{}

func (stubRevReaderZero) GetRevisionNumber(_ context.Context, _ TenantID, _ RevisionID) (int64, error) {
	return 0, nil
}
func (stubRevReaderZero) GetEffectiveFrom(_ context.Context, _ TenantID, _ RevisionID) (time.Time, error) {
	return time.Time{}, nil
}
func (stubRevReaderZero) GetAuthor(_ context.Context, _ TenantID, _ RevisionID) (AuthorInfo, error) {
	return AuthorInfo{}, nil
}

// NULL effective_from is valid (H5: commit 939bf24a). Resolver returns empty
// string value with no error — callers treat empty string as "not set".
func TestEffectiveDateResolver_NullReturnsEmptyValue(t *testing.T) {
	r := EffectiveDateResolver{}
	in := ResolveInput{
		TenantID: "t1", RevisionID: "r1",
		RevisionReader: stubRevReaderZero{},
	}
	v, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatalf("want nil error for zero effective_from, got %v", err)
	}
	if v.Value != "" {
		t.Fatalf("want empty Value for zero effective_from, got %q", v.Value)
	}
}

func TestEffectiveDateResolver_Resolve(t *testing.T) {
	r := EffectiveDateResolver{}
	in := ResolveInput{
		TenantID:   "tenant-a",
		RevisionID: "rev-1",
		RevisionReader: fakeRevisionReader{
			effectiveFrom: time.Date(2026, time.April, 2, 12, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60)),
		},
	}

	v1, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := r.Resolve(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}

	if v1.Value != "2026-04-02" {
		t.Fatalf("expected effective date 2026-04-02, got %#v", v1.Value)
	}
	if !bytes.Equal(v1.InputsHash, v2.InputsHash) {
		t.Fatal("expected stable hash across repeated resolves")
	}
}
