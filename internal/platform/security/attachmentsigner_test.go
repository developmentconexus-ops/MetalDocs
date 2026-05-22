package security

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

const testSignerSecret = "0123456789abcdef0123456789abcdef"

func TestSignedURL_IsExpired(t *testing.T) {
	expiry := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	su := SignedURL{URL: "x", ExpiresAt: expiry}

	if su.IsExpired(expiry.Add(-1)) {
		t.Fatalf("expected not expired 1ns before expiry")
	}
	if !su.IsExpired(expiry) {
		t.Fatalf("expected expired at expiry instant")
	}
	if !su.IsExpired(expiry.Add(1)) {
		t.Fatalf("expected expired 1ns after expiry")
	}
}

func TestAttachmentSigner_VerifyBoundary(t *testing.T) {
	expiry := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	attachmentID := "att-123"
	expiryStr := expiry.UTC().Format(time.RFC3339)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"1ns before expiry valid", expiry.Add(-1), true},
		{"at expiry invalid", expiry, false},
		{"1ns after expiry invalid", expiry.Add(1), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			signer := NewAttachmentSigner(testSignerSecret).WithClock(func() time.Time { return tc.now })
			sig := signer.Sign(attachmentID, expiry)
			got := signer.Verify(attachmentID, expiryStr, sig)
			if got != tc.want {
				t.Fatalf("Verify=%v want %v (now=%v expiry=%v)", got, tc.want, tc.now, expiry)
			}
		})
	}
}

func TestAttachmentSigner_VerifyBadSignature(t *testing.T) {
	expiry := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	signer := NewAttachmentSigner(testSignerSecret).WithClock(func() time.Time { return expiry.Add(-time.Hour) })

	if signer.Verify("att-123", expiry.UTC().Format(time.RFC3339), "deadbeef") {
		t.Fatalf("expected bogus signature to fail")
	}
	if signer.Verify("att-123", "not-a-timestamp", "deadbeef") {
		t.Fatalf("expected unparseable timestamp to fail")
	}
}

func TestAttachmentSigner_BuildDownloadURL(t *testing.T) {
	expiry := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	signer := NewAttachmentSigner(testSignerSecret).WithClock(func() time.Time { return expiry.Add(-time.Hour) })

	su := signer.BuildDownloadURL("/api/v1/attachments/att-123/download", "att-123", expiry)

	if !su.ExpiresAt.Equal(expiry.UTC()) {
		t.Fatalf("ExpiresAt=%v want %v", su.ExpiresAt, expiry.UTC())
	}
	if !strings.HasPrefix(su.URL, "/api/v1/attachments/att-123/download?") {
		t.Fatalf("URL missing base path: %s", su.URL)
	}

	parsed, err := url.Parse(su.URL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	q := parsed.Query()
	if got := q.Get("expiresAt"); got != expiry.UTC().Format(time.RFC3339) {
		t.Fatalf("expiresAt=%q want %q", got, expiry.UTC().Format(time.RFC3339))
	}
	sig := q.Get("signature")
	if sig == "" {
		t.Fatalf("signature query param missing")
	}

	if !signer.Verify("att-123", q.Get("expiresAt"), sig) {
		t.Fatalf("expected freshly built URL to verify")
	}

	if su.IsExpired(expiry.Add(-time.Hour)) {
		t.Fatalf("SignedURL should not be expired before its expiry")
	}
	if !su.IsExpired(expiry) {
		t.Fatalf("SignedURL should be expired at expiry instant")
	}
}

func TestNewAttachmentSigner_DefaultClock(t *testing.T) {
	signer := NewAttachmentSigner(testSignerSecret)
	if signer.now == nil {
		t.Fatalf("expected default clock to be set")
	}
}
