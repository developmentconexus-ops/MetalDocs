# Feature F8.2 — Spec (distributed rate limiter)

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.2-rate-limiter`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-05 — operator delegated approval via the M8 `/goal` brief
> ("prefer the real fix"); binding shape `../validation-contract.md` §2 (committed `218e2d12`).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Real fix or ADR-only? | Operator brief: "Prefer the real fix" → shared-store limiter, Redis-backed. Redis 7 already in compose (no new infra service). |
| 2 | (verified) Current limiter shape? | `internal/platform/ratelimit`: fixed-window counts in `sync.Map` (`middleware.go:39`), quotas hardcoded `config.go:78–91`; two mounts — pre-auth login (IP-keyed) + global envelope (`main.go:428/438`, `chain.go:32/36`). |
| 3 | (verified) Config surface? | Only `METALDOCS_TRUSTED_PROXY_CIDRS` wired; `config.go:12` env-override comment is stale/aspirational — correct the comment, do NOT implement per-route env quotas (milestone rabbit hole). |
| 4 | Library choice? | `github.com/redis/go-redis/v9` + `github.com/go-redis/redis_rate/v10` (GCRA leaky-bucket, single-round-trip Lua, battle-tested). Recorded in ADR 0071. |
| 5 | Behavior delta memory→redis? | Fixed window (memory) vs GCRA (redis) admit slightly different burst shapes. Contract §2.3 requires the test state the algorithm's documented burst tolerance; 429 shape/headers identical. |

## Consumer contract (FIRST)

- **Consumers:** the middleware chain links `pre_auth_login_rate_limit` (`chain.go:32`) and
  `rate_limit` (`chain.go:36`) — they call the existing `ratelimit.Middleware` API
  (`New(ctx, cfg)`, `GlobalEnvelopeWrap`, the login wrap) and must keep working **unchanged at the
  call sites** except for store wiring at construction (`main.go`); N horizontally-scaled api
  replicas consuming one shared budget.
- **Contract:** same route keying (`<route>:user:<id>` / `<route>:ip:<addr>`); same quotas; same 429
  RFC 9457 problem+json + headers; store backend selected by explicit env
  (`METALDOCS_RATELIMIT_STORE=memory|redis`, default `memory`; `METALDOCS_RATELIMIT_REDIS_ADDR`
  [+ optional password env] when redis); boot fails fast when
  `METALDOCS_MULTI_REPLICA=true` and store=memory.
- **Source of truth:** `internal/platform/ratelimit` existing tests + `chain_test.go` (REQ-MW-7 order)
  + validation-contract §2.

## What this feature implements

A counting-store abstraction inside `internal/platform/ratelimit` with the existing in-memory backend
(default, semantics preserved) and a Redis GCRA backend (`redis_rate`) sharing one budget across
replicas; env-based backend selection + fail-fast multi-replica guard; compose api service wired to
redis store; stale config comment corrected; ADR 0071 (defect, decision, algorithm delta,
memory-default rationale, guard, named scale-out/K8s revisit trigger; status ≤3 lines).

## Non-goals (mandatory)

- No per-route env quota overrides; no quota value changes; no chain-order or 429-shape changes.
- No Redis usage outside ratelimit (no cache/session creep); worker/jobs binaries untouched (they
  mount no HTTP rate limiter).
- No replica auto-detection (explicit env flag only).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Cross-replica: 2 limiter instances + 1 Redis admit ≤ Q(+documented burst) combined | new integration test in `internal/platform/ratelimit` (name fixed at impl, e.g. `TestCrossReplicaSharedBudget`) against real local Redis (compose) — fixture (miniredis) only if real unavailable, labeled | real preferred |
| Contrast pin: 2 in-memory instances admit up to 2×Q (defect made visible) | same test file, contrast subtest | real |
| Guard: multi-replica intent + memory store refuses boot | unit test on config/constructor error path | real |
| Existing semantics preserved | existing ratelimit test suite green (`go test ./internal/platform/ratelimit/...`) | real |
| Compose wiring | api service env sets redis store; 2-replica live drive (contract §2.4) admits ≤ Q(+burst) | real |
| ADR 0071 Accepted | file exists, cited by commits, status ≤3 lines | real |

## ADR needed?

- [x] Durable decision → **ADR 0071** `wiki/decisions/0071-distributed-rate-limiting-shared-store.md`.
