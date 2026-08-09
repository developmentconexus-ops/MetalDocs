# ADR 0094 — Session tokens are opaque server-side handles; RFC 8725 does not apply

- **Status:** Accepted
- **Date:** 2026-08-08
- **Scope:** Ratifies opaque server-side session tokens as MetalDocs's sole session mechanism and explicitly rejects JWTs for sessions. Rewrites REQ-AUTHN-3 in `wiki/architecture/backend-target-architecture.md` to state the invariants this design actually holds, in place of a requirement written against a technology never adopted. Does not change any other credential-mechanics value.
- **Amends:** ADR [`0058`](0058-auth-session-bcrypt-lockout-policy.md) decision item 4 (session token format), which already records the *what* — this ADR records the *why*, against RFC 8725 specifically, and is the citation target for REQ-AUTHN-3.
- **Depends on:** ADR [`0007`](0007-two-tier-authz.md) (tier split, unaffected).

---

## Context

`wiki/architecture/backend-target-architecture.md:123` (REQ-AUTHN-3, pre-amendment) read:

> Token handling follows RFC 8725 (alg pinning, no `none`, audience/issuer checks, short TTL). (MUST)

RFC 8725 is the JWT Best Current Practices document. Verified facts:

- `grep -rln 'jwt' --include=*.go internal/ apps/ | grep -v vendor` returns no matches. No file under `internal/` or `apps/` (excluding vendor) contains the string `jwt`.
- `go.mod` carries no JWT library dependency.
- The session/auth flow is entirely opaque server-side state, at `internal/modules/auth/application/service.go:1207-1238`:
  - `newSessionToken()` (`service.go:1207-1216`) generates 32 bytes from `crypto/rand`, base64url-encodes them (`token`), computes `sig := s.signToken(token)` (HMAC-SHA256 over `token`, keyed by `s.cfg.SessionSecret`), and returns `cookieValue = token + "." + sig` to the client alongside `hashToken(token)` — a raw SHA-256 hex digest — as the value persisted server-side (`session_id`).
  - `tokenHashFromCookieValue()` (`service.go:1218-1227`) splits the cookie on `.`, rejects malformed shapes, and compares the presented signature against `s.signToken(parts[0])` using `hmac.Equal` (`service.go:1223`) — not `==` — before resolving the session.
  - `hashToken()` (`service.go:1235-1238`) is the sole function that computes the server-side lookup key; it is `sha256.Sum256`, one-way.

No token encoding, claim structure, algorithm negotiation, or self-contained authorization data exists anywhere in this path. The REQ as written describes controls (alg pinning, rejecting `none`, audience/issuer claim checks) for a class of vulnerability — JWT algorithm confusion and claim forgery — that has no surface here, because there is no JWT.

## Decision

**Opaque, server-side-resolved, HMAC-signed session tokens are the session mechanism, permanently. JWTs are rejected for session tokens.** Adopting a JWT to satisfy REQ-AUTHN-3's letter would be a downgrade: a self-contained token cannot be invalidated before its stated expiry without an additional revocation-list side-channel, which is exactly the durable weakness RFC 8725 exists to manage around. This system already has the stronger property — server-side session state that is authoritative and instantly revocable — by construction, not by policy discipline layered on top of a self-contained token.

### RFC 8725 clause-by-clause disposition

| RFC 8725 concern | What this design does instead |
|---|---|
| **§3.1 alg pinning / reject `none`** | Not applicable — there is no algorithm negotiation for an attacker to redirect. One hard-coded construction, `hmac.New(sha256.New, secret)` (`service.go:1230`). This is not "pinned" in the RFC 8725 sense (a fixed choice among several the verifier still has to check); it is **unrepresentable** — the verifier has no code path that accepts anything but this one construction. Unrepresentable is the stronger property: pinning can be misconfigured or bypassed by a future edit that adds a second accepted algorithm; unrepresentable requires deleting `tokenHashFromCookieValue` itself to regress. |
| **§3.4/§3.11 audience/issuer checks** | Not applicable — the token carries no claims at all, self-asserted or otherwise. `token` is 32 random bytes; it encodes no identity, tenant, role, or expiry. Every fact the RFC's audience/issuer checks exist to validate — "does this token belong to this service, this principal, this tenant?" — is instead answered by a database read: `tokenHashFromCookieValue` resolves to a `session_id`, and `ResolveSession` looks up the session row that hash points to. A token that resolves to no row, an expired row, or a revoked row is rejected (`authdomain.ErrSessionNotFound` / `ErrSessionExpired` / `ErrSessionRevoked`, `service.go:483-503`). Nothing the bearer asserts is trusted; everything is looked up. |
| **§4.1 short TTL** | `SessionTTL`, already configured per ADR 0058 item 2 (12h absolute default, env-floor 1h) plus a sliding idle timeout (ADR 0058 item 3, 30 min default). Both are enforced server-side at `ResolveSession` time, not by trusting a client-presented `exp` claim. |
| **What JWT structurally cannot do, and this design can** | **Server-side revocation before expiry.** A signed JWT with a live signature is valid until its `exp` claim says otherwise — revoking it early requires a denylist side-channel that reintroduces exactly the server-side state a JWT was chosen to avoid, at which point the JWT is not actually stateless. This system revokes by design: `Logout` (`service.go:520-531`) flips `RevokedAt` on the one session; `RevokeSessionsByUserID`/`RevokeSessionsByUserIDTx` (`service.go:193-194`) mass-revoke on password change (`ChangePasswordForUser`), admin reset, and account deactivation (`UpdateUser`) — CWE-613 closed by construction, not by an added revocation list. This is the reason for the choice, not an incidental side benefit. |

### What is NOT claimed

This ADR does not claim RFC 8725 is bad guidance — it is the correct document for anyone shipping JWTs. It claims RFC 8725 is the wrong yardstick for a system that does not ship JWTs, and that measuring this design against it (as the pre-amendment REQ-AUTHN-3 did) can only ever produce a false negative or an incentive to add JWT machinery this system does not need.

## REQ-AUTHN-3 rewrite

The REQ line in `wiki/architecture/backend-target-architecture.md` is rewritten to state the invariants this design must actually hold, derived from the table above. ID and MUST class are unchanged.

## Consequences

- REQ-AUTHN-3 becomes measurable against this codebase's real design instead of a design it does not have.
- Any future proposal to introduce JWTs anywhere in the session path is a decision that supersedes this ADR, not a routine implementation choice — it must be raised as its own ADR, because it reopens a question this one closes.
- No code, schema, or migration change. This ADR ratifies existing, tested behavior.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Implement JWT alg-pinning/audience/issuer checks to satisfy REQ-AUTHN-3 literally | Rejected | Would mean building a JWT subsystem with no product need for it, purely to satisfy a requirement's letter — and the resulting design would be weaker (no cheap server-side revocation) than what exists today. |
| Delete REQ-AUTHN-3 rather than rewrite it | Rejected | The underlying concern (non-forgeable, short-lived, revocable session credentials) is real and must stay a MUST; deleting the REQ would remove the gate's ability to notice a future regression (e.g., someone weakening `hmac.Equal` to `==`, or dropping revocation on deactivation). |
| Reclassify REQ-AUTHN-3 as SHOULD until a "real" token-hardening review happens | Rejected | The properties in the rewritten REQ are already true and already tested (see traceability report); downgrading a satisfied MUST to SHOULD to sidestep writing the ADR would be exactly the kind of unlabelled, unjustified softening CLAUDE.md's Global Maximum rule forbids. |

## References

- ADR [`0058-auth-session-bcrypt-lockout-policy.md`](0058-auth-session-bcrypt-lockout-policy.md) — decision item 4, the pre-existing record of the token *format*; this ADR adds the RFC 8725 disposition and is what REQ-AUTHN-3 now cites.
- `internal/modules/auth/application/service.go:1207-1238` — `newSessionToken`, `tokenHashFromCookieValue`, `signToken`, `hashToken`.
- `internal/modules/auth/application/service.go:483-531` — `ResolveSession` (lookup-not-trust) and `Logout` (single-session revocation).
- `internal/modules/auth/application/service_test.go` — `TestLogout_EmptyAndMalformedTokenReturnError`, `TestResolveSession_IdleTimeout`, `TestChangePasswordForUser_RevokesSessions`, `TestUpdateUser_DeactivateRevokesSessions`, `TestNewService_RejectsShortSessionSecret` — behavioral proof of the properties this ADR records, cited from REQ-AUTHN-3.
- `internal/modules/auth/application/service_session_opacity_test.go` — new tests added by this change proving token opacity, CSPRNG entropy, and hashed-at-rest storage (properties nothing previously exercised directly).
- `internal/modules/auth/domain/model_test.go` — `TestAuthenticatedSession_RedactsRawToken`.
- `docs/superpowers/analysis/req-disposition-2026-08-07.md` §"REQ-AUTHN-3" — the disposition analysis this ADR resolves.
