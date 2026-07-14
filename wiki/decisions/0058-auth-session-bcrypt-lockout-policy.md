# ADR 0058 — Auth session, bcrypt, and account-lockout policy

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records the existing, enforced credential-mechanics policy for the `auth` module: session-cookie format, session TTL + sliding idle timeout, bcrypt work factor, and account-lockout thresholds. Closes tech-debt T-010 (`wiki/modules/auth-tech-debt.md`). Does not cover session tenancy binding (ADR 0055) or the tier-1/tier-2 authz split (ADR 0007).
- **Depends on:** ADR 0007 (two-tier authz — tier split, not credential mechanics).

---

## Context

`internal/modules/auth/application/service.go` and `internal/platform/authn/config.go` enforce a specific set of credential-mechanics parameters, backed by code and tests, but no ADR previously recorded the choice. ADR 0007 covers the tier-1/tier-2 authorization split, not these values.

### Verified runtime facts

- **bcrypt cost 12.** `internal/modules/auth/application/service.go:33` — `bcryptCost = 12`, used both for real password hashing (`service.go:945`, `bcrypt.GenerateFromPassword(password, bcryptCost)`) and for a constant-time dummy hash on unknown-user login attempts (`service.go:153`) to avoid a timing oracle that would distinguish "user not found" from "wrong password."
- **Session TTL: 12-hour absolute default, env-overridable.** `internal/platform/authn/config.go:58-65` — `sessionTTLHours := 12`, overridden by `METALDOCS_AUTH_SESSION_TTL_HOURS` (validated `>= 1`). Applied at mint time: `service.go:322`, `ExpiresAt: now.Add(s.cfg.SessionTTL)`.
- **Sliding idle timeout: 30-minute default, env-overridable, independent of the absolute TTL.** `internal/platform/authn/config.go:67-77` — `sessionIdleMinutes := 30` ("Defaults to 30 minutes (ISO 27001 path / backend standardization parameters)"), overridden by `METALDOCS_AUTH_SESSION_IDLE_MINUTES` (`0` explicitly disables idle expiry; the 12h absolute TTL still applies in that case). Enforced per-request: `service.go:394` — `if s.cfg.SessionIdleTimeout > 0 && now.Sub(session.LastSeenAt) > s.cfg.SessionIdleTimeout { ... }` expires the session independently of the absolute `ExpiresAt`.
- **Session-cookie/token format.** `internal/modules/auth/application/service.go:117-126` (referenced by T-010's original surface citation) — opaque bearer token `<base64url(rand32)>.<base64url(HMAC-SHA256(secret, token))>`; only `SHA-256(token)` is persisted server-side as `session_id`, so a DB read alone cannot forge or replay a session token.
- **Account lockout: 5 failed attempts default, 15-minute lock default, both env-overridable with floors.** `internal/platform/authn/config.go:88-104` — `maxFailedAttempts := 5` (env override validated `>= 3`), `lockMinutes := 15` (env override validated `>= 1`). Enforced at login: `service.go:267`, `if state.LockedUntil != nil && state.LockedUntil.After(s.now().UTC()) { ... }`, gated by a per-identity lock while checking (`service.go:249`, comment: "Verify credentials while holding a per-identity lock so the lockout check ... [is race-free]").
- **Password minimum length: 8 chars default, env-overridable with an 8-char floor.** `internal/platform/authn/config.go:79-86` — `passwordMinLength := 8`, env override validated `>= 8` (the floor cannot be lowered via env, only raised).

## Decision

**The following credential-mechanics values are the binding policy, expressed as defaults with a validated env-override floor/ceiling; the code in `internal/platform/authn/config.go` is the single source of truth for the live value in any environment:**

1. **Password hashing:** bcrypt, cost factor 12, fixed in code (not env-configurable) — `bcryptCost` at `service.go:33`. A constant-time dummy hash is computed on every login path regardless of whether the user exists, to prevent user-enumeration via timing.
2. **Session absolute TTL:** 12 hours by default; `METALDOCS_AUTH_SESSION_TTL_HOURS` may raise or lower it (floor `1`).
3. **Session sliding idle timeout:** 30 minutes by default; `METALDOCS_AUTH_SESSION_IDLE_MINUTES` may change it, `0` explicitly disables idle expiry (absolute TTL still applies). This is the backend-standardization-parameters value referenced in the code comment.
4. **Session token format:** opaque random 32-byte token, HMAC-SHA256-signed with a server secret, stored server-side only as its SHA-256 digest (`session_id`). Never a JWT or other self-describing/decodable token.
5. **Account lockout:** 5 consecutive failed login attempts by default (env floor `3`) locks the account for 15 minutes by default (env floor `1` minute). Lockout check + failure recording happen under a per-identity lock to prevent a race that would let concurrent attempts bypass the threshold.
6. **Password minimum length:** 8 characters, and the env override cannot lower this floor.

Any change to these defaults or floors is a policy change and MUST be proposed as an amendment to this ADR (or a superseding ADR), not a silent config-default edit.

## Consequences

- T-010 (`wiki/modules/auth-tech-debt.md`) is closed by this ADR — the policy is now a decision record, not just enforced code.
- Reviewers checking auth changes for policy drift should diff against the six rules above, not against the current `config.go` defaults alone (env overrides are legitimate; changing the *default* or the *floor* is not, without amending this ADR).
- No migration, schema change, or code change is required by this ADR — it documents and binds existing, verified runtime behavior.

## References

- `internal/platform/authn/config.go:58-104` — all six default values + env-override validation.
- `internal/modules/auth/application/service.go:33,117-126,153,249,267,322,394,945` — bcrypt cost, token format, dummy-hash timing defense, lockout check, TTL application, idle-timeout enforcement, real hashing call site.
- `wiki/modules/auth-tech-debt.md` T-010 — tech-debt row closed by this ADR.
- ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md) — tier split this ADR does not restate.
- ADR [`0055-global-auth-identities.md`](0055-global-auth-identities.md) — session tenancy binding, a distinct concern from the credential mechanics recorded here.
