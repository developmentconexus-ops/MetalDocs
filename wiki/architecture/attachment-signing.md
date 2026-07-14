# Attachment Signing

Last verified: 2026-05-22

## Purpose

Time-bounded HMAC-SHA256 signatures for attachment download URLs. Prevents
unbounded shareable links and ensures only callers with the signing secret can
mint valid URLs.

## Key Length Invariant

`METALDOCS_ATTACHMENTS_SIGNING_SECRET` MUST be at least **32 bytes**.

- NIST SP 800-107 / FIPS 198-1: HMAC keys SHOULD be at least the hash output
  size (32 bytes for SHA-256) to preserve the full security strength.
- Enforced at two layers (defense in depth):
  - `config.LoadAttachmentsConfig` — returns an error on startup if the env
    var is shorter than 32 bytes.
    See [config/attachments.go:56-58](../../internal/platform/config/attachments.go).
  - `security.NewAttachmentSigner` — **panics** if constructed with a secret
    shorter than 32 bytes. The signer is load-bearing for download
    authorization; a soft return path would let mis-wired tests or future
    sub-services silently weaken the invariant.
    See [security/attachmentsigner.go](../../internal/platform/security/attachmentsigner.go).

## Generating a Secret

```bash
# 64-char hex string = 32 bytes of entropy (well over the minimum)
openssl rand -hex 32

# base64-encoded 32 bytes → 44-char string (also fine)
openssl rand -base64 32
```

## Key Files

| File | Role |
|------|------|
| [internal/platform/config/attachments.go](../../internal/platform/config/attachments.go) | Env validation, 32-byte check |
| [internal/platform/security/attachmentsigner.go](../../internal/platform/security/attachmentsigner.go) | Signer — HMAC sign/verify/URL builder |
| [tests/unit/attachments_config_test.go](../../tests/unit/attachments_config_test.go) | Config length tests |
| [tests/unit/attachment_signer_test.go](../../tests/unit/attachment_signer_test.go) | Signer panic tests |
