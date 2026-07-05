// Package crypto is a pure, DB-free platform primitive: AES-256-GCM envelope
// sealing (Seal/Open), DEK wrapping under a KEK (WrapDEK/UnwrapDEK), and the
// JSON envelope wire codec (EncodeEnvelope/DecodeEnvelope/IsEnvelope) used by
// the security module's TenantCrypto port (M7 F7.3, ADR 0070 decision 4).
//
// This package has no knowledge of tenants, Postgres, or any MetalDocs
// domain concept — it only knows about bytes and keys. Callers (the security
// module) own key lifecycle, tenant scoping, and persistence.
//
// Key material must NEVER be logged or printed by any caller of this
// package.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// KeySize is the required length (bytes) for every key this package accepts
// — AES-256 keys and DEKs/KEKs alike.
const KeySize = 32

// envelopeEnc is the only "enc" discriminator this package's codec currently
// recognises. Future algorithm migrations would add a new discriminator
// value rather than repurpose this one (ADR 0070 non-goal: no KEK rotation /
// re-wrap tooling in F7.3).
const envelopeEnc = "aesgcm.v1"

// envelopeWire is the on-the-wire JSON shape: {"enc":"aesgcm.v1","data":"<base64>"}.
type envelopeWire struct {
	Enc  string `json:"enc"`
	Data string `json:"data"`
}

// Seal encrypts plaintext under key using AES-256-GCM with a freshly
// generated random nonce, prefixed to the returned ciphertext
// (nonce || ciphertext || tag). key must be exactly KeySize (32) bytes.
func Seal(key, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return sealed, nil
}

// Open decrypts sealed (as produced by Seal) under key, verifying the GCM
// authentication tag. Any tampering with the ciphertext, or use of the wrong
// key, fails closed with a non-nil error and no partial plaintext.
func Open(key, sealed []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(sealed) < nonceSize {
		return nil, fmt.Errorf("crypto: sealed input shorter than nonce size (%d)", nonceSize)
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: open failed (tampered ciphertext or wrong key): %w", err)
	}
	return plaintext, nil
}

// WrapDEK encrypts dek under kek, returning the sealed envelope bytes
// (Seal's nonce-prefixed output). Both keys must be KeySize (32) bytes.
func WrapDEK(kek, dek []byte) ([]byte, error) {
	wrapped, err := Seal(kek, dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap DEK: %w", err)
	}
	return wrapped, nil
}

// UnwrapDEK decrypts a DEK previously produced by WrapDEK. Fails closed on a
// wrong KEK or tampered wrapped bytes, exactly like Open.
func UnwrapDEK(kek, wrapped []byte) ([]byte, error) {
	dek, err := Open(kek, wrapped)
	if err != nil {
		return nil, fmt.Errorf("crypto: unwrap DEK: %w", err)
	}
	return dek, nil
}

// EncodeEnvelope wraps sealed ciphertext bytes into the canonical envelope
// JSON string: {"enc":"aesgcm.v1","data":"<base64>"}.
func EncodeEnvelope(sealed []byte) string {
	w := envelopeWire{Enc: envelopeEnc, Data: base64.StdEncoding.EncodeToString(sealed)}
	// json.Marshal on this fixed, always-valid struct shape cannot fail.
	b, _ := json.Marshal(w)
	return string(b)
}

// DecodeEnvelope parses payload as envelope JSON and returns the decoded
// sealed ciphertext bytes. Returns an error if payload is not valid envelope
// JSON or carries an unrecognised "enc" discriminator.
func DecodeEnvelope(payload string) ([]byte, error) {
	var w envelopeWire
	if err := json.Unmarshal([]byte(payload), &w); err != nil {
		return nil, fmt.Errorf("crypto: decode envelope: not valid JSON: %w", err)
	}
	if w.Enc != envelopeEnc {
		return nil, fmt.Errorf("crypto: decode envelope: unrecognised enc %q, want %q", w.Enc, envelopeEnc)
	}
	sealed, err := base64.StdEncoding.DecodeString(w.Data)
	if err != nil {
		return nil, fmt.Errorf("crypto: decode envelope: invalid base64 data: %w", err)
	}
	return sealed, nil
}

// IsEnvelope reports whether payload is well-formed envelope JSON carrying
// the recognised "enc" discriminator — the detector callers (e.g. the audit
// reader) use to distinguish encrypted payloads from legacy plaintext rows
// without attempting a decode-then-fallback.
func IsEnvelope(payload string) bool {
	var w envelopeWire
	if err := json.Unmarshal([]byte(payload), &w); err != nil {
		return false
	}
	return w.Enc == envelopeEnc
}

// newGCM constructs an AES-256-GCM AEAD from key, validating key length.
func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("crypto: key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new GCM: %w", err)
	}
	return gcm, nil
}
