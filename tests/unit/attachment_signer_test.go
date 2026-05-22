package unit

import (
	"testing"

	"metaldocs/internal/platform/security"
)

func TestNewAttachmentSignerPanicsOnShortSecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when constructing AttachmentSigner with 31-byte secret")
		}
	}()
	_ = security.NewAttachmentSigner("0123456789012345678901234567890")
}

func TestNewAttachmentSignerPanicsOnEmptySecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when constructing AttachmentSigner with empty secret")
		}
	}()
	_ = security.NewAttachmentSigner("")
}

func TestNewAttachmentSignerAcceptsThirtyTwoByteSecret(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("did not expect panic for 32-byte secret: %v", r)
		}
	}()
	s := security.NewAttachmentSigner("01234567890123456789012345678901")
	if s == nil {
		t.Fatalf("expected non-nil signer")
	}
}
