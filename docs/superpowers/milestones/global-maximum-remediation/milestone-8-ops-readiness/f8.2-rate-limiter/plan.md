# Feature F8.2 — Distributed rate limiter

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.2-rate-limiter`
> **Status:** Planning

## Source

- Milestone spec row F8.2 (`../milestone.md`) + validation-contract §2 (binding).

## Plan

**Task A — store abstraction + Redis backend + guard (implementer subagent, sonnet, TDD)**
Files: `internal/platform/ratelimit/*` (new store.go / store_memory.go / store_redis.go or similar),
`go.mod`/`go.sum` (go-redis v9 + redis_rate v10), config additions, `apps/api/cmd/metaldocs-api/main.go`
(construction wiring only), stale comment fix `config.go:12`.
1. Failing test first: `TestCrossReplicaSharedBudget` (2 middleware instances, 1 Redis, quota Q,
   combined admits ≤ Q+burst; contrast subtest: 2 memory instances admit ~2Q). Real Redis via
   `REDIS_ADDR`-style env with skip-if-unavailable + miniredis fallback labeled.
2. Failing guard test: multi-replica=true + store=memory → constructor error.
3. Implement: extract counting decision behind an interface; memory impl = existing logic verbatim;
   redis impl = redis_rate GCRA mapped to the per-route per-window quotas; env selection + guard.
4. All existing ratelimit tests green; `go build ./...` green.

**Task B — ADR 0071 (implementer subagent, sonnet)**
`wiki/decisions/0071-distributed-rate-limiting-shared-store.md`: defect (per-replica ×N), decision
(shared-store GCRA via redis_rate), memory-default rationale, guard, algorithm delta
(fixed-window→GCRA burst semantics), named triggers (first multi-node deploy → redis mandatory;
K8s adoption → revisit exposure/topology). Status ≤3 lines. Wiki index entry.

**Task C — compose wiring + 2-replica live drive (main session)**
- compose api service: `METALDOCS_RATELIMIT_STORE=redis`, addr → compose redis.
- Scale api ×2 (or duplicate service on 8081/8082 behind the gateway if scale conflicts with published
  port), hammer login route past quota, capture combined admits + 429 body. Contract §2.4.

Order: A → B → C. Spec + quality reviews after A (code) and B (doc); C is evidence.

## Execution notes

(filled during execution)
