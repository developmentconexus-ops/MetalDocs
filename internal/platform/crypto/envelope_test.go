package crypto_test

// envelope_test.go — M7 F7.3 Task B TDD (RED-first): pure unit tests for the
// AES-256-GCM envelope primitives in internal/platform/crypto. No DB, no
// network — this package is a leaf platform primitive.
//
//   go test ./internal/platform/crypto/...

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"

	"metaldocs/internal/platform/crypto"
)

func mustKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read key: %v", err)
	}
	return key
}

func TestSeal_Open_RoundTrip(t *testing.T) {
	key := mustKey(t)
	plaintext := []byte("tenant secret payload")

	sealed, err := crypto.Seal(key, plaintext)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	if len(sealed) == 0 {
		t.Fatalf("Seal() returned empty ciphertext")
	}

	opened, err := crypto.Open(key, sealed)
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("Open() = %q, want %q", opened, plaintext)
	}
}

func TestSeal_ProducesRandomNoncePerCall(t *testing.T) {
	key := mustKey(t)
	plaintext := []byte("same plaintext every time")

	first, err := crypto.Seal(key, plaintext)
	if err != nil {
		t.Fatalf("Seal() first call error = %v", err)
	}
	second, err := crypto.Seal(key, plaintext)
	if err != nil {
		t.Fatalf("Seal() second call error = %v", err)
	}
	if bytes.Equal(first, second) {
		t.Fatalf("Seal() produced identical ciphertext across two calls — nonce is not random")
	}
}

func TestOpen_TamperedCiphertextFails(t *testing.T) {
	key := mustKey(t)
	sealed, err := crypto.Seal(key, []byte("do not tamper with me"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 0xFF // flip last ciphertext byte

	if _, err := crypto.Open(key, tampered); err == nil {
		t.Fatalf("Open() on tampered ciphertext = nil error, want failure")
	}
}

func TestOpen_WrongKeyFails(t *testing.T) {
	key := mustKey(t)
	wrongKey := mustKey(t)

	sealed, err := crypto.Seal(key, []byte("secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	if _, err := crypto.Open(wrongKey, sealed); err == nil {
		t.Fatalf("Open() with wrong key = nil error, want failure")
	}
}

func TestSeal_RejectsWrongKeyLength(t *testing.T) {
	if _, err := crypto.Seal([]byte("too-short"), []byte("x")); err == nil {
		t.Fatalf("Seal() with a non-32-byte key = nil error, want failure")
	}
}

func TestOpen_RejectsSealedShorterThanNonce(t *testing.T) {
	key := mustKey(t)
	if _, err := crypto.Open(key, []byte("short")); err == nil {
		t.Fatalf("Open() on input shorter than the nonce = nil error, want failure")
	}
}

func TestWrapDEK_UnwrapDEK_RoundTrip(t *testing.T) {
	kek := mustKey(t)
	dek := mustKey(t)

	wrapped, err := crypto.WrapDEK(kek, dek)
	if err != nil {
		t.Fatalf("WrapDEK() error = %v, want nil", err)
	}
	if bytes.Equal(wrapped, dek) {
		t.Fatalf("WrapDEK() returned the DEK unmodified — not actually wrapped")
	}

	unwrapped, err := crypto.UnwrapDEK(kek, wrapped)
	if err != nil {
		t.Fatalf("UnwrapDEK() error = %v, want nil", err)
	}
	if !bytes.Equal(unwrapped, dek) {
		t.Fatalf("UnwrapDEK() = %x, want %x", unwrapped, dek)
	}
}

func TestUnwrapDEK_WrongKEKFails(t *testing.T) {
	kek := mustKey(t)
	wrongKEK := mustKey(t)
	dek := mustKey(t)

	wrapped, err := crypto.WrapDEK(kek, dek)
	if err != nil {
		t.Fatalf("WrapDEK() error = %v", err)
	}
	if _, err := crypto.UnwrapDEK(wrongKEK, wrapped); err == nil {
		t.Fatalf("UnwrapDEK() with wrong KEK = nil error, want failure")
	}
}

func TestEnvelope_EncodeDecode_RoundTrip(t *testing.T) {
	sealed := []byte{0x01, 0x02, 0x03, 0x04}

	encoded := crypto.EncodeEnvelope(sealed)
	if !strings.Contains(encoded, `"enc":"aesgcm.v1"`) {
		t.Fatalf("EncodeEnvelope() = %q, want it to carry enc=aesgcm.v1", encoded)
	}

	decoded, err := crypto.DecodeEnvelope(encoded)
	if err != nil {
		t.Fatalf("DecodeEnvelope() error = %v, want nil", err)
	}
	if !bytes.Equal(decoded, sealed) {
		t.Fatalf("DecodeEnvelope() = %x, want %x", decoded, sealed)
	}
}

func TestDecodeEnvelope_RejectsUnknownEnc(t *testing.T) {
	if _, err := crypto.DecodeEnvelope(`{"enc":"rot13.v1","data":"AAAA"}`); err == nil {
		t.Fatalf("DecodeEnvelope() with unknown enc = nil error, want failure")
	}
}

func TestDecodeEnvelope_RejectsMalformedJSON(t *testing.T) {
	if _, err := crypto.DecodeEnvelope(`not json at all`); err == nil {
		t.Fatalf("DecodeEnvelope() with malformed JSON = nil error, want failure")
	}
}

func TestIsEnvelope_DetectsEnvelopeVsPlaintext(t *testing.T) {
	sealed := []byte{0xAA, 0xBB}
	envelope := crypto.EncodeEnvelope(sealed)

	if !crypto.IsEnvelope(envelope) {
		t.Fatalf("IsEnvelope(%q) = false, want true", envelope)
	}
	if crypto.IsEnvelope("plain old audit payload json {}") {
		t.Fatalf("IsEnvelope() on plaintext payload = true, want false")
	}
	if crypto.IsEnvelope(`{"some":"other json"}`) {
		t.Fatalf("IsEnvelope() on unrelated JSON = true, want false")
	}
	if crypto.IsEnvelope("") {
		t.Fatalf("IsEnvelope(\"\") = true, want false")
	}
}
