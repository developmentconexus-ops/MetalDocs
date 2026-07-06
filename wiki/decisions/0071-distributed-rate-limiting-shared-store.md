# ADR 0071 — Distributed rate limiting via a pluggable counting-store abstraction (memory | Redis GCRA)

> **Status:** Accepted (2026-07-05)
> **Module(s):** `internal/platform/ratelimit` (counting-store abstraction, memory + Redis backends) · `metaldocs-api` (composition root, startup guard)
> **REQ IDs:** `backend-target-architecture.md` REQ-ASYNC-* (fail-open availability posture, precedent: ADR 0067), REQ-DB-* (fail-fast invariant guards at startup)
> **Supersedes / amends:** none. New platform framework; no prior ADR governed rate-limiting scale-out.

## Context

The 2026-07-03 final architecture review (§9) named a pre-customer blocker: `internal/platform/ratelimit`
implements a fixed-window limiter keyed in a per-process `sync.Map`. `metaldocs-api` is stateless and
horizontally replicated by design, but the limiter's counters are not shared across replicas — running
N replicas silently multiplies every configured limit by N, with no error, warning, or signal that the
control has been diluted. This is a correctness gap in a security-adjacent control, not a performance
issue. Redis 7 is already present in the docker-compose stack for other purposes, so a shared backing
store is available without adding new infrastructure.

## Decision

1. **A counting-store abstraction replaces the direct `sync.Map` implementation.** `ratelimit` exposes
   a store interface that the limiter drives; call sites are unchanged.
2. **Two backends.** `memory` (today's `sync.Map` fixed-window counter, default, correct only for a
   single replica — dev/test posture) and `redis` (GCRA leaky-bucket via `go-redis/v9` +
   `go-redis/redis_rate/v10`, sharing one budget across all replicas via one Redis instance/cluster).
3. **Explicit env selection, no autodetection.** `METALDOCS_RATELIMIT_STORE=memory|redis` selects the
   backend; `METALDOCS_RATELIMIT_REDIS_ADDR` configures the Redis target when `redis` is selected.
4. **Fail-fast startup guard against silent scale-out dilution.** If `METALDOCS_MULTI_REPLICA=true` and
   the store is `memory`, the process refuses to boot. A misconfigured multi-replica deployment must
   fail loudly at startup, not silently under-enforce its limits in production.
5. **Fail-open on Redis runtime error.** A Redis error/timeout at request time logs a warning and lets
   the request through rather than rejecting it or crashing the handler. Availability wins over strict
   enforcement for this control: a Redis outage must not take the API down.
   - **Login-limiter tradeoff (named, compensated).** The pre-auth login limiter is a brute-force
     control, and fail-open means a Redis outage disables per-IP login throttling. This is acceptable
     because brute-force is independently bounded by **account lockout** (`METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS`
     / `METALDOCS_AUTH_LOGIN_LOCK_MINUTES`), which is DB-backed and unaffected by Redis availability — so
     the compensating control holds when the rate limiter is degraded. The rate limiter shaves attempt
     *velocity*; lockout enforces the *ceiling*.

## Consequences

- **Positive.** Rate limits hold their configured value regardless of replica count once `redis` is
  selected; the startup guard converts an invisible scale-out defect into an immediate boot failure;
  single-replica dev/test keeps today's zero-dependency `memory` path with no behavior change.
- **Algorithm delta (accepted, not hidden).** `memory` is fixed-window; `redis` is GCRA leaky-bucket.
  Burst-admission semantics differ slightly at window boundaries between the two backends — this is an
  accepted divergence, not a defect, since only one backend is ever active per deployment. The 429
  wire shape (RFC 9457 `problem+json`) and route/key derivation are unchanged across both backends.
- **Costs / risks.** Redis becomes a soft runtime dependency for multi-replica deployments (mitigated
  by fail-open); one more backend to operate and monitor.
- **Named triggers (bounded follow-ups).**
  - First multi-node/production deployment → `redis` store becomes mandatory; the startup guard is the
    enforcement mechanism, not a documentation reminder.
  - Kubernetes adoption → revisit the deployment target and scrape/limits topology together with the
    M8 deploy doc; not addressed by this ADR.

## Alternatives considered

- **Per-replica quota division** (divide the configured limit by N replicas) — rejected: wrong under
  autoscaling, where N changes at runtime and every replica would need live reconfiguration.
- **Sticky sessions** (route a client to the same replica so its local counter is authoritative) —
  rejected: breaks the stateless-replica invariant the whole request lifecycle depends on.
- **Postgres-based counting** (a shared counter table) — rejected: puts hot-path rate-limit writes on
  the primary transactional database for every request, competing with real transactional load.
