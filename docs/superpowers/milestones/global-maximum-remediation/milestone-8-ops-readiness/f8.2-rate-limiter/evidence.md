# Feature F8.2 — Evidence

> **Milestone:** 8 — Ops Readiness  ·  **Feature:** `f8.2-rate-limiter`  ·  **Closed:** 2026-07-06
> **Contract:** `spec.md` (distributed shared-store rate limiter; consumer = middleware chain
> `pre_auth_login_rate_limit` + `rate_limit`, N replicas one shared budget).

## What was implemented

- **Counting-store abstraction** in `internal/platform/ratelimit`: `Store` interface with two backends
  — `memory_store.go` (default, preserves the existing `sync.Map` fixed-window semantics) and
  `redis_store.go` (GCRA leaky-bucket via `redis_rate/v10` on `go-redis/v9`, single-round-trip Lua,
  one budget shared across replicas). `store_config.go` selects the backend by explicit env
  (`METALDOCS_RATELIMIT_STORE=memory|redis`, default `memory`;
  `METALDOCS_RATELIMIT_REDIS_ADDR` [+ optional password]) and **fails boot fast** when
  `METALDOCS_MULTI_REPLICA=true` with a memory store.
- **Call sites unchanged** — `middleware.go` `New(ctx,cfg)` / `GlobalEnvelopeWrap` / login wrap keep
  the same signatures; only construction in `main.go` wires the selected store. Same route keying,
  same quotas, same 429 RFC 9457 problem+json + headers.
- **Compose wiring** — api service sets `METALDOCS_RATELIMIT_STORE=redis` + `REDIS_ADDR=redis:6379`
  + `MULTI_REPLICA=true` (committed `1bbb59be`); Redis 7 already in the stack (no new infra).
- **ADR 0071** records the defect, the GCRA-vs-fixed-window delta, the memory-default rationale, the
  guard, and the named scale-out/K8s revisit trigger.
- Producer matches consumer contract: chain links call the unchanged `ratelimit.Middleware` API;
  shared budget is enforced at the store, transparent to the chain.
- Commits: `1bbb59be feat(m8): distributed rate limiting (shared store) + Prometheus /metrics`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — cross-replica failing→green | `go test ./internal/platform/ratelimit/ -run TestCrossReplicaSharedBudget -v` | `redis shared budget: 10/30 admitted (quota 10, tolerance +1)`; contrast `memory: 20/30 admitted — double the intended quota of 10 (defect pinned)` | **real** (real Redis at 127.0.0.1:6379) |
| Guard — multi-replica+memory refuses boot | `go test -run TestStoreConfig_MultiReplicaGuard -v` | 7/7 subtests PASS: `multi-replica + explicit memory → error`, `+ default (empty) → error`, `+ redis w/addr → ok`, `redis without addr → error`, `unknown backend → error`, single-replica paths ok | real |
| Existing semantics preserved | `go test ./internal/platform/ratelimit/...` | `ok metaldocs/internal/platform/ratelimit 2.427s` (full suite incl. burst/per-user/eviction/race) | real |
| Static build | `go build ./...` | exit 0 | — |
| Runtime — 2-replica live shared budget (§2.4) | Bring up `api` + `api-b` (both `STORE=redis`, `REDIS_ADDR=redis:6379`, `MULTI_REPLICA=true`), drive 30 `POST /api/v1/auth/login` alternating replicas by DNS from one curl container (identical source IP → identical IP-keyed login key) | **admitted=10, limited(429)=20** — combined admits exactly the login quota (~10/min) across both replicas, NOT 2×10. Requests 1–10 → 401 (bad creds = admitted pre-auth), remainder → 429 | **real** (containers on `compose_default`, real Redis) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Cross-replica: 2 instances + 1 Redis admit ≤ Q(+burst) | yes | redis subtest 10/30 (Q=10, +1 tol) |
| Contrast pin: 2 in-memory admit up to 2×Q | yes | memory subtest 20/30 (defect made visible) |
| Guard: multi-replica intent + memory refuses boot | yes | TestStoreConfig_MultiReplicaGuard 7/7 |
| Existing semantics preserved | yes | full ratelimit suite green |
| Compose wiring + 2-replica live drive ≤ Q(+burst) | yes | live drive admitted=10 limited=20 |
| ADR 0071 Accepted, cited by commits, ≤3-line status | yes | `wiki/decisions/0071-*.md`, cited by `1bbb59be` |

## Review disposition

- Spec-compliance review: PASS — no per-route env quota overrides, no quota changes, no chain-order /
  429-shape change, worker/jobs untouched (they mount no HTTP limiter); stale `config.go:12`
  env-override comment corrected, not implemented (rabbit hole avoided).
- Code-quality review: PASS — Store abstraction keeps call sites unchanged; fail-open on Redis error
  (availability > strict limiting); GCRA burst delta documented in ADR + test tolerance, not hidden.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Fleet-wide per-route env quota tuning | Rabbit hole per spec; current quotas correct for v1 | Revisit at first multi-node prod load-tune (ADR 0071 scale-out trigger) |
