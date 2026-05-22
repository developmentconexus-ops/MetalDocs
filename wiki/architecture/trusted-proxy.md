# Architecture: Trusted-Proxy CIDR Allowlist

> **Last verified:** 2026-05-21 (commit `def24e4a`)
> **Scope:** how MetalDocs decides whether to honor `X-Forwarded-For` / `X-Forwarded-Proto` from an upstream hop; one CIDR allowlist drives every header-trust decision in `internal/platform/security` and the auth audit-log path.
> **Out of scope:** TLS termination, load-balancer health-check gating, per-route rate-limit quotas (see `internal/platform/ratelimit/`).
> **Key files:**
> - `internal/platform/config/trusted_proxy.go`     `LoadTrustedProxyCIDRs` / `ParseTrustedProxyCIDRs`
> - `internal/platform/security/proxy.go`     `IsTrustedRemote` / `ClientIP` helpers
> - `internal/platform/security/origin_protection.go`     `(p *OriginProtection) sameOrigin` consumer
> - `internal/platform/security/ratelimit.go`     `(r *RateLimiter) requestIdentity` consumer
> - `internal/modules/auth/application/service.go`     `(s *Service) remoteIP` consumer (audit log)
> - `internal/platform/authn/config.go`     loads CIDRs once at boot and propagates them to `authapp.Config`
> - `apps/api/cmd/metaldocs-api/main.go`     wires CIDRs into `OriginProtectionConfig` at startup
> - `tests/unit/trusted_proxy_test.go`     helper invariants (CIDR match, IPv6, multi-hop, fail-closed)
> - `tests/unit/origin_protection_test.go`     scheme-spoof regression guard (C1)
> - `tests/unit/rate_limit_middleware_test.go`     anon-bucket-collapse regression guard (H4)

---

## 1. Problem this solves

`X-Forwarded-*` headers can be set by any client that can reach the server. Without a hop-trust check:

- A directly-reachable attacker can send `X-Forwarded-Proto: https` + `Origin: https://allowed-host` and bypass the origin allowlist on cookied mutators (CSRF gate collapse — finding **C1**).
- A reverse-proxied deployment that keys rate-limit buckets on `r.RemoteAddr` ends up with every anon request sharing the proxy's IP, so a single bucket caps the entire tenant (finding **H4**).

The fix is the standard pattern: only honor forwarded headers when the immediate upstream peer (`r.RemoteAddr`) sits inside an operator-configured CIDR allowlist. Default empty = no header trust = fail-closed.

---

## 2. Configuration

### 2.1 Env var

```
METALDOCS_TRUSTED_PROXY_CIDRS=10.0.0.0/8,fd00::/8,127.0.0.1/32
```

- Comma-separated; whitespace tolerated.
- Both IPv4 and IPv6 prefixes accepted (`netip.ParsePrefix`).
- Malformed entries fail loud at boot — `LoadTrustedProxyCIDRs()` returns an error that propagates through `authn.LoadRuntimeConfig` and aborts startup.
- Empty / unset = `nil` slice = every remote treated as untrusted (no header is ever honored).

### 2.2 Loader

`config.LoadTrustedProxyCIDRs()` is called exactly once per boot from `internal/platform/authn/config.go`. The result is stored on `authapp.Config.TrustedProxyCIDRs` and copied into:

- `OriginProtectionConfig.TrustedProxyCIDRs` (wired in `apps/api/cmd/metaldocs-api/main.go`).
- `RateLimitConfig.TrustedProxyCIDRs` (already populated upstream of `NewRateLimiter`).
- `authapp.Config.TrustedProxyCIDRs` (consumed by `Service.remoteIP` when emitting audit-log entries).

This is the single source of truth — adding a new consumer means reading the slice off the loaded config, not re-parsing the env var.

---

## 3. Helpers (`internal/platform/security/proxy.go`)

### 3.1 `IsTrustedRemote(r *http.Request, cidrs []netip.Prefix) bool`

Returns `true` iff `r.RemoteAddr` parses to a `netip.Addr` contained by at least one prefix in `cidrs`. Returns `false` on:

- `nil` / empty `cidrs` (fail-closed).
- Malformed `RemoteAddr` (`net.SplitHostPort` failure or unparseable host).
- Any IP outside every configured prefix.

IPv4-mapped IPv6 addresses are normalized via `Unmap()` before prefix membership testing so `::ffff:10.0.0.1` matches `10.0.0.0/8`.

### 3.2 `ClientIP(r *http.Request, cidrs []netip.Prefix) netip.Addr`

- When `IsTrustedRemote` is `true`: returns the leftmost valid `X-Forwarded-For` token (the original client). Multi-hop chains (`client, proxy-1, proxy-N`) keep the leftmost entry.
- Otherwise: returns the parsed `r.RemoteAddr`.
- Returns the zero `netip.Addr` (`!IsValid()`) when nothing usable can be parsed; callers must check `IsValid()` before stringifying.

The helper never panics and never reads `X-Forwarded-For` from an untrusted source.

---

## 4. Consumers

### 4.1 CSRF origin gate — `OriginProtection.sameOrigin`

Scheme defaults to `https` when `r.TLS != nil`, else `http`. `X-Forwarded-Proto` is consulted **only** when `IsTrustedRemote` returns `true`. Host is taken from `r.Host` and normalized via `normalizeOrigin`.

Regression guard: `TestOriginProtectionBlocksSpoofedXForwardedProtoFromUntrustedSource` — a 203.0.113.7 attacker sending `X-Forwarded-Proto: https` against an `Origin: https://internal-host:8080` cookied POST must be rejected with 403 even when no trusted CIDR matches it.

### 4.2 Rate-limit identity — `RateLimiter.requestIdentity`

Order of preference:

1. Authenticated user via `authdomain.CurrentUserFromContext` (preferred — survives proxy changes).
2. `iamdomain.UserIDFromContext` fallback (legacy IAM-only paths).
3. `ClientIP(req, r.trustedProxyCIDRs)` — leftmost trusted XFF or `RemoteAddr`.
4. `"ip:unknown"` last-ditch sentinel so the bucket is bounded to a single key per process.

Regression guard: `TestRateLimiterIsolatesAnonByXFFWhenBehindTrustedProxy` proves two distinct upstream clients behind one trusted proxy do not collapse into a shared bucket; companion `TestRateLimiterIgnoresXFFFromUntrustedSource` proves a directly-reachable client cannot forge `X-Forwarded-For` to dodge the limit.

### 4.3 Auth audit log — `Service.remoteIP`

Truncated to 128 bytes after resolution, then persisted in `auth_login_events` / `auth_sessions` audit rows. Without the helper, behind-proxy logins all showed the LB IP; after the fix the audit trail reflects the real client.

---

## 5. Operator runbook

1. Identify every hop between the public client and the API process. List the immediate-upstream IPs (the box that opens the TCP connection to the API).
2. Set `METALDOCS_TRUSTED_PROXY_CIDRS` to the union of those IPs as CIDRs. Prefer narrow `/32` / `/128` over broad `/8` ranges; only widen when the upstream is operator-owned and not user-reachable.
3. Restart the API. The loader rejects malformed entries at boot — no silent partial trust.
4. Verify behind the proxy: a `GET /api/v1/auth/me` from two distinct upstream clients (different `X-Forwarded-For` values, same proxy) must be rate-limited independently. From a non-proxied path, a forged `X-Forwarded-For` must be ignored.
5. Empty / unset is safe — the API runs fail-closed and refuses to honor any forwarded header. The cost is reduced fidelity in rate-limit bucketing and audit logs, not a security loss.

---

## 6. Adding a new consumer

If a future module needs the real client IP or the real request scheme:

1. Add a `TrustedProxyCIDRs []netip.Prefix` field on the module's `Config` struct.
2. Populate it in `internal/platform/authn/config.go` (or wherever the module config is assembled) — do **not** re-parse `METALDOCS_TRUSTED_PROXY_CIDRS` from env.
3. Call `security.IsTrustedRemote` / `security.ClientIP` — never write a bespoke `X-Forwarded-*` parser.
4. Add a regression test in `tests/unit/` that proves an untrusted source cannot use the new header path to escalate.

The contract is intentionally narrow: one env var, one parser, one helper, one allowlist. Drift here means CSRF-gate or rate-limit fail-open bugs reappear.
