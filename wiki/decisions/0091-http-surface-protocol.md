# ADR 0091 — The HTTP surface protocol: SurfacePublisher, the generated surface table, and Mount-is-total

> **Status:** Accepted and implemented 2026-08-06.
> **Scope:** How every module's HTTP delivery seam mounts routes and how the composition root proves the mounted surface matches the declared one. Not the PDP — see [`0090`](0090-tier1-pdp-generated-from-spec.md) for tier-1's data-source change.
> **Key files:**
> - `internal/platform/httprouter/publisher.go:19-26` — `SurfacePublisher` interface
> - `internal/platform/httprouter/recorder.go` — `Recorder`, the per-publisher `Muxer` wrapper feeding the boot assertion
> - `apps/api/cmd/metaldocs-api/surface.go:25-30` — `assertSurface`, the four boot checks
> - `apps/api/cmd/metaldocs-api/main.go:840-884` — publisher construction, mount loop, and the `assertSurface` call
> - `apps/api/cmd/metaldocs-api/httpsurface_gen.go`, `httpsurface_e2e_gen.go` — generated surface tables (`cmd/gen-http-surface`)
> - `apps/api/cmd/metaldocs-api/surface_conformance_test.go` — the per-operation conformance suite

## Context

Before this program, module HTTP delivery seams had no single shape. The system-impact analysis for this work (`docs/superpowers/analysis/2026-08-05-http-surface-protocol-system-impact.md`) found route truth scattered across five independently-drifting enumerations: hand-typed `routeRules` (tier-1, closed by ADR 0090), per-module `RegisterRoutes` methods with no common contract, the OpenAPI spec's own declared operations, generated codegen tags, and an ad-hoc response-typing scope list. A module could register a route that the spec never declared, or declare an operation in the spec that no module ever mounted, and nothing would notice until a request or a manual audit found it. `wiki/backend/http-kernel.md`'s own "Legacy and open flags" register (pre-this-program) named the resulting public-path list a "drifted copy" of tier-1's classification — the exact symptom of no single mounted-surface authority existing.

The operator ruling on 2026-08-05 (system-impact analysis §10, AS-2) chose the whole-surface option: build the protocol once, over every module's delivery seam including the five families that were assumed off-spec but turned out already to be in the spec (auth, search, security, feature-flags, health) — no two-regime state, no permanently-off-spec carve-out.

## Decision

**Every route owner implements `SurfacePublisher`** (`internal/platform/httprouter/publisher.go:19-26`):

```go
type SurfacePublisher interface {
    Name() string   // stable identity used in boot-assertion messages
    Tag() string     // the OpenAPI tag this publisher owns — one tag, one publisher
    Mount(Muxer)      // registers every route the publisher serves
}
```

Sixteen publishers exist, one per OpenAPI tag (`main.go:840-857`: auth, health, observability, featureFlags, audit, search, security, taxonomy, tokens, controlledDocuments, iam, documents, templates, approval, distribution, notifications) — a seventeenth (`e2ePublisher`) is appended only under `METALDOCS_E2E=1`. `RegisterRoutes` is gone; `permissions.go`'s comment on the deletion (`permissions.go:14-15`) and this ADR both record it: **`Mount` replaces `RegisterRoutes`, everywhere, no partial migration.**

**Mount is total.** A publisher registers every route it owns, unconditionally, on every boot (`publisher.go:10-13`). Availability is a handler-level 501, never a routing-level absence — a route that is only sometimes mounted makes the surface environment-dependent, which defeats the boot assertion below (it can only prove a claim that holds on every boot, not one that holds only when some optional dependency happened to be present).

**The generated surface table is the single declared truth.** `cmd/gen-http-surface` reads the OpenAPI spec's `x-authz-capability`/`x-authz-visibility` extensions and emits `apps/api/cmd/metaldocs-api/httpsurface_gen.go` — a `map[string]surfaceRule` (`httpsurface_gen.go:10-17`) keyed by the exact mux pattern (147 entries at time of writing), plus `httpsurface_e2e_gen.go` for the optional E2E-only spec. ADR 0090 covers what tier-1 does with this table; this ADR covers how it comes to be trustworthy at all.

**The boot assertion is the proof, not a convention.** `assertSurface` (`surface.go:25-30`) runs once at startup against `mounted` (per-publisher recorded patterns, from one `httprouter.Recorder` per publisher — `main.go:868-873`) and `surface` (the generated table). Four checks, all evaluated and aggregated into one error so a failure is fixable in one restart (`surface.go:23-24,31`):

1. **Tag coverage** — exactly one publisher claims each expected tag (`surface.go:33-54`).
2. **Mounted ⊆ declared** — every pattern any publisher actually registered has a surface rule (`surface.go:64-71`).
3. **Declared ⊆ mounted** — every surface rule's pattern was actually registered by some publisher (`surface.go:73-80`).
4. **Ownership** — a publisher may only mount patterns whose surface tag equals its own claimed tag (`surface.go:82-100` onward). This is deliberately evaluated **per-publisher against that publisher's own recorded patterns**, not against the global union, because a global-set check would let every publisher claim a distinct tag while silently mounting each other's operations and checks 1–3 would all still pass (`surface.go:14-21`).

A failure here is `slog.Error("http surface", "err", err)` followed by `os.Exit(1)` (`main.go:880-884`) — a fatal, not a logged warning.

**What was actually verified, stated precisely.** The program's Definition of Done asked for a real boot on both the ordinary path and `METALDOCS_E2E=1`. That is **not** what happened. Every attempt hit port 8081 held by an unrelated `wslrelay.exe`, so the assertion was exercised through `TestRealPublishersSatisfyTheAssertion` — the real publisher list against the real generated table, in-process — and the negative case (commenting out one publisher produces exactly that fatal) was verified the same way in Task 15 step 9. The ledger records the fallback as a fallback (`.superpowers/sdd/progress.md:219-221`, `:353-357`; commit `cc35b5f8`).

This is written out rather than smoothed over because a governance record that reads stronger than its evidence is the failure mode this repo catalogs as Class 12. Concretely unproven: `deps.Cleanup()` on the assertion-failure path under a real connection pool. The in-process test constructs no pool, so a leaked connection or a cleanup panic on that path would not have been observed. Anyone who boots the API on a free port closes this gap for free — do it and amend this paragraph.

**The conformance suite proves enforcement, not just wiring.** `surface_conformance_test.go` walks every declared operation and asserts its capability is genuinely enforced at tier-1 — the per-operation completion of the "boot proves wiring" guarantee above with "and the wiring means what it says". `TestNoDeclaredOperationIsUnreachable` is one case in that suite; see below.

## The one deliberately red test

`TestNoDeclaredOperationIsUnreachable` is **red by design** and is expected to stay red until a separate program lands. It asserts that every permission-guarded operation's declared capability is grantable to at least one assignable role — and finds several that are not, because tier-1's `CanDo` and tier-2's `authz.Require` read disjoint grant tables (`iam_user_roles`/groups vs. `user_process_areas`; full finding in `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md`, commit `8cdb66ac`). The boot assertion this ADR records proves **mounted ⊆ declared** — that the HTTP surface is exactly what the spec says it is. It says nothing about who is authorized to call what it mounts; that is the grant model's job, and the grant model is what the red test is reporting on. The two findings are independent: this ADR's four checks all pass; the fifth, separate question the conformance suite also asks — reachability through the grant tables — does not, and is not silenced to make the checklist look clean.

## No third ADR for the middleware chain

The plan for this program considered a third ADR for a `route_resolve` middleware link. §6 of the design dropped that link before implementation, so `chain.go:25` (`apiChain`, the ordered `[]chainLink` literal) is byte-identical before and after this program, and the locked chain-order constraint (CLAUDE.md, `chain_test.go`'s REQ-MW-7) holds without any change requiring a decision record.

## Consequences

- **Wins.** One shape for all 15 modules' delivery seams (plus 2 platform packages: health, observability — deliberately not called "modules", `publisher.go:5-8`, since CLAUDE.md reserves that word for the 15 bounded contexts). A missing or misrouted publisher is a boot fatal on the very next start, not a runtime discovery. The generated table (ADR 0090) and the boot assertion are two independent proofs of the same claim from different directions — codegen proves declared-from-spec is internally consistent; `assertSurface` proves mounted matches declared at runtime — so a bug in one is unlikely to be masked by the other.
- **Costs.** Sixteen (seventeen with E2E) publishers is more moving parts than one composition-root file, and `SurfacePublisher.Mount`'s total-mount rule pushes conditional availability down into handler-level 501s across every module rather than letting the composition root skip a mount — a larger, more mechanical diff, chosen once (the operator's whole-surface ruling) rather than reopened per module.
- **What this ADR does not claim.** It does not claim the mounted surface is correct in the sense of "does the right thing" — only that it is exactly the declared surface, mounted exactly once, by exactly the publisher that owns its tag. Business correctness inside each handler is each module's own test suite's job.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Keep per-module `RegisterRoutes` with no common interface, add a CI script that diffs mounted routes against the spec | Rejected | A CI diff script over runtime output is itself a fourth or fifth enumeration — the same shape the http-surface-protocol program exists to delete, just moved into CI instead of `main.go`. |
| A single "route registry" package that every module writes into eagerly at init time (`init()` side effects) | Rejected | Hides the mount order and ownership behind package-init side effects, which is harder to reason about than an explicit `Mount(Muxer)` call in the composition root and breaks the "one recorder per publisher" mechanism check 4 depends on. |
| Make the boot assertion advisory (log, don't exit) so a partial boot is still reachable in a degraded mode | Rejected | Directly contradicts Mount-is-total's own rationale: availability is meant to be a handler-level concern, not a routing-level one; an advisory assertion reintroduces the environment-dependent surface this ADR exists to remove. |
| Migrate only the codegen-covered modules now, leave `auth`/`search`/`security`/`featureFlags`/`health` as a permanently-off-spec carve-out | Rejected — operator ruling A, 2026-08-05 | See Context; the assumed off-spec cost was wrong (all five were already in the spec), so the larger-scope option was also the cheaper one once measured correctly. |

## References
- ADR [`0090-tier1-pdp-generated-from-spec.md`](0090-tier1-pdp-generated-from-spec.md) — the companion ADR: what tier-1 does with the table this ADR's boot assertion proves is trustworthy. Different owner (PDP vs. delivery-seam wiring), different reviewers, different lifetime — why this is two ADRs, not one with two sections.
- `docs/superpowers/analysis/2026-08-05-http-surface-protocol-system-impact.md:160-161` — the "ADR required, and they are not the same decision" ruling
- `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` — the disjoint-grant-table finding `TestNoDeclaredOperationIsUnreachable` reports, out of scope for this ADR
- `wiki/decisions/0012-contract-first-api.md` — spec-as-source-of-truth, the contract this program's generator reads from
- CLAUDE.md — "Async = transactional outbox" / module-boundary rules that reserve "module" for the 15 bounded contexts, motivating `publisher.go`'s naming note
