package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SignedURL couples a signed attachment download URL with its expiry, so
// callers cannot accidentally swap the (attachmentID, expiresAt, signature)
// arguments at Verify time. Use SignedURL.IsExpired with an injectable clock
// for deterministic expiry checks.
type SignedURL struct {
	URL       string
	ExpiresAt time.Time
}

// IsExpired reports whether the signed URL has reached or passed its expiry
// relative to the supplied clock. Boundary semantics match Verify: the URL
// becomes invalid AT the expiry instant (now == expiresAt is expired).
func (s SignedURL) IsExpired(now time.Time) bool {
	return !now.UTC().Before(s.ExpiresAt.UTC())
}

type AttachmentSigner struct {
	secret []byte
	now    func() time.Time
}

// MinAttachmentSecretLength is the minimum HMAC-SHA256 key length per
// NIST SP 800-107 / FIPS 198-1: keys MUST be at least the hash output
// size (32 bytes for SHA-256) to preserve the full security strength.
const MinAttachmentSecretLength = 32

func NewAttachmentSigner(secret string) *AttachmentSigner {
	if len(secret) < MinAttachmentSecretLength {
		panic(fmt.Sprintf("security: attachment signing secret must be at least %d bytes, got %d", MinAttachmentSecretLength, len(secret)))
	}
	return &AttachmentSigner{secret: []byte(secret), now: time.Now}
}

// WithClock injects a deterministic time source. Intended for tests.
func (s *AttachmentSigner) WithClock(now func() time.Time) *AttachmentSigner {
	if now != nil {
		s.now = now
	}
	return s
}

func (s *AttachmentSigner) Sign(attachmentID string, expiresAt time.Time) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(attachmentID + "|" + expiresAt.UTC().Format(time.RFC3339)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *AttachmentSigner) Verify(attachmentID, expiresAtRFC3339, signature string) bool {
	expiresAt, err := time.Parse(time.RFC3339, strings.TrimSpace(expiresAtRFC3339))
	if err != nil {
		return false
	}
	if !s.now().UTC().Before(expiresAt.UTC()) {
		return false
	}
	expected := s.Sign(attachmentID, expiresAt)
	return hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(strings.TrimSpace(signature))))
}

func (s *AttachmentSigner) BuildDownloadURL(basePath, attachmentID string, expiresAt time.Time) SignedURL {
	values := url.Values{}
	values.Set("expiresAt", expiresAt.UTC().Format(time.RFC3339))
	values.Set("signature", s.Sign(attachmentID, expiresAt))
	return SignedURL{
		URL:       fmt.Sprintf("%s?%s", strings.TrimSpace(basePath), values.Encode()),
		ExpiresAt: expiresAt.UTC(),
	}
}
