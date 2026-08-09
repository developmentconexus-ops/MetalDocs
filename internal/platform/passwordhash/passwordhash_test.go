package passwordhash_test

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"metaldocs/internal/platform/passwordhash"
)

// TestHashArgon2id_RoundTrip proves REQ-AUTHN-1's memory-hard-KDF clause:
// passwords are hashed with Argon2id (not a fast/non-memory-hard function),
// the PHC string carries the params it was minted with, and a matching
// plaintext verifies while a wrong one is rejected.
func TestHashArgon2id_RoundTrip(t *testing.T) {
	encoded, err := passwordhash.HashArgon2id([]byte("CorrectHorseBatteryStaple!"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=") {
		t.Fatalf("encoded hash = %q, want $argon2id$v=... PHC format", encoded)
	}

	ok, err := passwordhash.VerifyArgon2id(encoded, []byte("CorrectHorseBatteryStaple!"))
	if err != nil {
		t.Fatalf("VerifyArgon2id(correct): %v", err)
	}
	if !ok {
		t.Fatal("VerifyArgon2id(correct) = false, want true")
	}

	ok, err = passwordhash.VerifyArgon2id(encoded, []byte("WrongPassword!"))
	if err != nil {
		t.Fatalf("VerifyArgon2id(wrong): %v", err)
	}
	if ok {
		t.Fatal("VerifyArgon2id(wrong) = true, want false")
	}
}

func TestHashArgon2id_ParamsSurviveRoundTrip(t *testing.T) {
	encoded, err := passwordhash.HashArgon2id([]byte("pw"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	summary, err := passwordhash.Argon2ParamsSummary(encoded)
	if err != nil {
		t.Fatalf("Argon2ParamsSummary: %v", err)
	}
	// REQ-AUTHN-1: memory-hard KDF params (m=KiB, t=iterations, p=parallelism)
	// must be present so a future parameter change never orphans this hash.
	if !strings.Contains(summary, "m=") || !strings.Contains(summary, "t=") || !strings.Contains(summary, "p=") {
		t.Fatalf("params summary = %q, want m=/t=/p= present", summary)
	}
}

func TestVerify_DispatchesOnAlgo(t *testing.T) {
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	argonHash, err := passwordhash.HashArgon2id([]byte("secret"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}

	ok, err := passwordhash.Verify(passwordhash.AlgoBcrypt, bcryptHash, []byte("secret"))
	if err != nil || !ok {
		t.Fatalf("Verify(bcrypt, correct) = (%v, %v), want (true, nil)", ok, err)
	}
	ok, err = passwordhash.Verify(passwordhash.AlgoBcrypt, bcryptHash, []byte("wrong"))
	if err != nil || ok {
		t.Fatalf("Verify(bcrypt, wrong) = (%v, %v), want (false, nil)", ok, err)
	}

	ok, err = passwordhash.Verify(passwordhash.AlgoArgon2id, []byte(argonHash), []byte("secret"))
	if err != nil || !ok {
		t.Fatalf("Verify(argon2id, correct) = (%v, %v), want (true, nil)", ok, err)
	}
	ok, err = passwordhash.Verify(passwordhash.AlgoArgon2id, []byte(argonHash), []byte("wrong"))
	if err != nil || ok {
		t.Fatalf("Verify(argon2id, wrong) = (%v, %v), want (false, nil)", ok, err)
	}
}

// TestVerify_FailsClosedOnUnrecognizedAlgorithm proves the no-fallback
// principle: an unrecognized password_algo value must never match, and must
// never silently try a different comparison.
func TestVerify_FailsClosedOnUnrecognizedAlgorithm(t *testing.T) {
	ok, err := passwordhash.Verify("scrypt", []byte("whatever-bytes"), []byte("secret"))
	if ok {
		t.Fatal("Verify(unrecognized algo) = true, want false (fail closed)")
	}
	if !errors.Is(err, passwordhash.ErrUnrecognizedAlgorithm) {
		t.Fatalf("err = %v, want ErrUnrecognizedAlgorithm", err)
	}
}

func TestVerify_FailsClosedOnMalformedArgon2idHash(t *testing.T) {
	ok, err := passwordhash.Verify(passwordhash.AlgoArgon2id, []byte("not-a-phc-string"), []byte("secret"))
	if ok {
		t.Fatal("Verify(malformed argon2id hash) = true, want false")
	}
	if !errors.Is(err, passwordhash.ErrMalformedHash) {
		t.Fatalf("err = %v, want ErrMalformedHash", err)
	}
}

// TestVerify_FailsClosedOnOutOfRangeArgon2idParams proves the bounds check in
// parseArgon2id: a stored PHC string with an out-of-range m/t/p must be
// rejected with ErrMalformedHash rather than reaching argon2.IDKey. The
// negative-m case is the concrete defect this guards - Sscanf's %d parses a
// negative m with no error, and casting it straight to uint32 wraps to a
// ~4 TiB allocation request, which crashes the whole process (an
// unrecoverable Go runtime OOM, not a request-scoped panic).
func TestVerify_FailsClosedOnOutOfRangeArgon2idParams(t *testing.T) {
	tests := []struct {
		name   string
		params string
	}{
		{name: "negative memory", params: "m=-1,t=1,p=1"},
		{name: "oversized memory", params: "m=99999999999,t=1,p=1"},
		{name: "zero time", params: "m=19456,t=0,p=1"},
		{name: "negative time", params: "m=19456,t=-1,p=1"},
		{name: "zero parallelism", params: "m=19456,t=1,p=0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := "$argon2id$v=19$" + tt.params + "$c29tZXNhbHQ$c29tZWhhc2g"
			ok, err := passwordhash.Verify(passwordhash.AlgoArgon2id, []byte(encoded), []byte("secret"))
			if ok {
				t.Fatalf("Verify(%q) = true, want false (fail closed)", tt.params)
			}
			if !errors.Is(err, passwordhash.ErrMalformedHash) {
				t.Fatalf("Verify(%q) err = %v, want ErrMalformedHash", tt.params, err)
			}
		})
	}
}
