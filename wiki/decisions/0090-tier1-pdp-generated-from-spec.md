# ADR 0090 — Tier-1's data source moves from a hand-typed table to the generated HTTP surface (amends ADR 0022)

> **Status:** Accepted and implemented 2026-08-06. Amends [`0022`](0022-authz-capability-coherence.md); supersedes nothing.
> **Scope:** Tier-1 route→capability resolution's DATA SOURCE only. Tier-2 (`authz.Require`), the DB tripwire, and the capability vocabulary are untouched.
> **Key files:**
> - `apps/api/cmd/metaldocs-api/httpsurface_gen.go:17` — generated `httpSurface` map (147 route entries), keyed by mux pattern
> - `apps/api/cmd/metaldocs-api/permissions.go:17-35` — `newPermissionResolver`, the sole tier-1 authority
> - `apps/api/cmd/metaldocs-api/surface.go:25-30` — `assertSurface`, the boot assertion `newPermissionResolver` depends on for its unreachable-branch guarantee
> - `internal/modules/iam/domain/model.go:73,166` — `Capability` type + `validCapabilities` (unchanged)
> - `internal/modules/iam/authz/authz.go:100` — `Require(ctx, tx, capability, areaCode)` (unchanged)
> - `apps/api/cmd/metaldocs-api/chain.go:25` — middleware chain (byte-identical; no new link)

## Context

ADR 0022 established capability-based authorization as a coherent, CI-bound two-tier model: tier-1 rejects at the HTTP edge, tier-2 enforces the real area inside the transaction, and a DB trigger tripwire is the last line. That shape is sound and is not being revisited here.

What ADR 0022 did not question was tier-1's **data source**: `apps/api/cmd/metaldocs-api/permissions.go`'s `routeRules` — a 286-line, hand-typed, ordered slice of method/path-pattern/capability/visibility rows, resolved by a first-match-wins linear scan (`resolveRoutePermission`). It was one of five hand-synced enumerations of route truth the http-surface-protocol program's opening problem statement named (alongside per-module `RegisterRoutes` methods, the OpenAPI spec's own route list, generated codegen tags, and the CI H-D response-typing scope list). Each enumeration could drift from the others independently, and `routeRules` in particular could drift from the OpenAPI spec that is supposed to be contract truth (ADR 0012) — a route could carry one capability in the spec's `x-authz-*` annotations and a different one, or none, in `routeRules`, and nothing would notice until a live request disagreed with a reviewer's expectation.

The http-surface-protocol program (Tasks 1–18 of this branch) replaced the five-enumeration mess with one generated artifact: `cmd/gen-http-surface` reads the OpenAPI spec's `x-authz-capability`/`x-authz-visibility` extensions and emits `apps/api/cmd/metaldocs-api/httpsurface_gen.go`, a `map[string]surfaceRule` keyed by the exact mux pattern each operation mounts under. Task 18 then deleted `routeRules` and its walker outright (commit `cc35b5f8`) once every operation had a generated entry and the boot assertion (ADR 0091) proved mounted routes and declared routes coincide.

## Decision

**Tier-1 becomes a pattern lookup into the generated table.** `newPermissionResolver` (`permissions.go:17-35`) no longer scans an ordered rule list; it does `rule, ok := httpSurface[pattern]` (`permissions.go:27`) against the pattern the stdlib `http.ServeMux` itself matched (`mux.Handler(r)`, `permissions.go:19`). No match on a matched pattern is `VisibilityUnresolved` — not a fallback guess — because ADR 0091's boot assertion (`assertSurface`, four checks) makes that branch unreachable at boot: a declared operation with no mount, or a mount with no declared entry, is a startup fatal, not a runtime possibility. `newPublicPathChecker` and `newPasswordChangeAllowedChecker` derive from the same table (`permissions.go:37-42, 51-63`), so there is one authoritative source for all three questions tier-1 answers, not three hand-maintained ones.

**What did NOT change, stated plainly because this is an amendment:**
- The capability vocabulary — `Capability` type and `validCapabilities` (`internal/modules/iam/domain/model.go:73,166`) — is untouched. Every value `httpSurface` carries is the same typed const tier-2 and the DB tripwire already spoke.
- The two-tier shape is untouched: tier-1 still rejects cheaply at the edge; tier-2 still enforces the real area inside the transaction.
- Tier-2's in-tx `authz.Require(ctx, tx, capability, areaCode)` (`internal/modules/iam/authz/authz.go:100`) is untouched — same signature, same call sites, same area-scope binding guard (ADR 0022 Phase 7).
- The DB tripwire (`enforce_capability_asserted()`, ADR 0022 §Item 7) is untouched.
- The middleware chain is untouched — `chain.go:25` is byte-identical before and after this program; tier-1 resolution happens inside the existing `iam_authz` link, not a new one.

**What changed is exactly one thing: where tier-1 gets `(Capability, Visibility)` from.** Before: a hand-typed Go literal, reviewed by eye against the spec. After: generated from the spec's own `x-authz-*` annotations, so the spec and the enforcement it describes cannot independently drift — a change to who may call an operation is a spec edit, and `go generate` is what turns it into code, the same discipline ADR 0012 already requires for every other part of the contract.

## Relationship to the unresolved grant-model finding

This ADR does not touch, and does not resolve, the separate defect `TestNoDeclaredOperationIsUnreachable` reports: tier-1's `CapabilityService.CanDo` and tier-2's `authz.Require` read **disjoint grant tables** (`iam_user_roles`/groups vs. `user_process_areas`), so a capability can be declared reachable by this ADR's generated table yet be grantable to no one through the tier-1 path. That is a grant-assignment problem, one level below where this ADR operates (this ADR is about which capability a route *declares*, not which principals *hold* it). It is recorded, with the operator's ratified remediation decisions, in `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` (commit `8cdb66ac`), and is deliberately being fixed in its own program rather than folded into this one — see that document's Scope note. `TestNoDeclaredOperationIsUnreachable` stays red as the evidence that justifies it.

## Consequences

- **Wins.** One fewer hand-synced enumeration (five → four, pending the grant-model program closing a different one). A capability change is a spec edit + `go generate`, not a spec edit *and* a separate hand-typed table edit that can silently miss the sync. `newPermissionResolver` is 19 lines instead of a 286-line table plus a matching walker; `permissions.go` fell from ~290 lines to 63.
- **The unreachable-fallback branch changes meaning.** Under `routeRules`, no match meant "nobody wrote a rule for this path" — a plausible, if bad, steady state. Under the generated table, no match on a matched mux pattern means the boot assertion (ADR 0091) failed to run or was bypassed — a wiring bug, not a tier to guess at (`permissions.go:29-30`).
- **Costs.** The generator (`cmd/gen-http-surface`) is new machinery with its own correctness surface (parsing the spec's `x-authz-*` extensions correctly); errors there are now a codegen bug class instead of a hand-typed-table bug class. Mitigated by Task 17's row-by-row human review of all 147 annotations (`docs/superpowers/analysis/2026-08-05-annotation-review.md`) — the one mitigation the http-surface-protocol program's own Definition of Done names as load-bearing, because the conformance suite (ADR 0091 §Conformance) cannot catch a transcription error: it derives its expected capability from the same annotation the implementation derives its enforced capability from.
- **No differential/parity gate exists between the old and new tables, deliberately.** `routeRules` was never an oracle, only one hand-typed attempt to express the same decision the spec now expresses directly; comparing two non-authoritative artifacts yields differences forever and truth never (http-surface-protocol program, Definition of Done §"Deliberately Not Proven").

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Keep `routeRules`, add a CI diff-check against the spec | Rejected | Still two hand-synced artifacts, now with a lint bolted on; the lint itself becomes a fifth enumeration (the http-surface-protocol system-impact analysis flagged exactly this shape as a prior review's own fourth hand-synced enumeration of route truth, later removed). |
| Fold tier-1 resolution into an interprocedural call-graph lint over handler bodies | Rejected | Already rejected for the analogous tier-2 `authz-call-present` rule (ADR 0022 amendment 2026-06-08, ADR 0023) — the tx-layer area is DB-derived, not request-supplied, and the equivalent handler-body shape does not exist for capability either. |
| Generate tier-1 from a second, IAM-owned annotation file instead of the OpenAPI spec | Rejected | Reintroduces a second source of truth for the same fact the spec already declares (ADR 0012: the spec is the single public HTTP contract). |

## References
- ADR [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) — the two-tier model and capability vocabulary this ADR amends
- ADR [`0091-http-surface-protocol.md`](0091-http-surface-protocol.md) — the platform framework (`SurfacePublisher`, the boot assertion) this ADR's generated table depends on for its unreachable-branch guarantee
- ADR [`0012-contract-first-api.md`](0012-contract-first-api.md) — spec-as-source-of-truth, which this ADR extends to authorization metadata
- `docs/superpowers/analysis/2026-08-05-http-surface-protocol-system-impact.md:160-161` — "ADR required" ruling that named this ADR and ADR 0091 as two decisions, not one
- `docs/superpowers/analysis/2026-08-05-annotation-review.md` — the 147-row human review mitigating the "no parity gate" residual risk
- `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` — the disjoint-grant-table finding this ADR does not resolve
