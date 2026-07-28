package idempotency_test

import (
	"errors"
	"testing"

	"metaldocs/internal/platform/idempotency"
)

// TestValidateKey pins the single Idempotency-Key wire rule (F-QA4-6). Both the
// Require middleware and the bespoke-replay handlers (approval signoff /
// fast-forward / review-verdict / route-admin / submit / delegation) route
// through this function, so its behaviour IS the API-wide contract.
func TestValidateKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		key  string
		want error
	}{
		{name: "empty", key: "", want: idempotency.ErrKeyRequired},
		{name: "free string", key: "idem-1", want: idempotency.ErrKeyInvalid},
		{name: "uuid missing dashes", key: "11111111111141118111111111111111", want: idempotency.ErrKeyInvalid},
		{name: "uuid with surrounding space", key: " 11111111-1111-4111-8111-111111111111 ", want: idempotency.ErrKeyInvalid},
		{name: "uuid too long", key: "11111111-1111-4111-8111-1111111111111", want: idempotency.ErrKeyInvalid},
		{name: "lowercase uuid", key: "11111111-1111-4111-8111-111111111111", want: nil},
		{name: "uppercase uuid", key: "AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE", want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := idempotency.ValidateKey(tc.key)
			if !errors.Is(err, tc.want) {
				t.Fatalf("ValidateKey(%q) = %v, want %v", tc.key, err, tc.want)
			}
			// IsValidKey must stay consistent with ValidateKey — no second dialect.
			if got, want := idempotency.IsValidKey(tc.key), tc.want == nil; got != want {
				t.Fatalf("IsValidKey(%q) = %v, want %v", tc.key, got, want)
			}
		})
	}
}
