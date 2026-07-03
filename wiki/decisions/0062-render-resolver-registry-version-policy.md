# ADR 0062 — Resolver registry version field: recorded per-resolution, no compatibility/deprecation policy defined yet

- **Status:** Accepted (records current shape; declines to invent an unneeded policy — see Decision)
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records what the render module's resolver `Version()` field actually does today (a per-resolver integer stamped onto every resolved value) and what it does not do (no compatibility check, no deprecation mechanism, no registry-level version). Closes backlog row R-003 (`wiki/backlog/render-fanout-refactor.md`) and its linked tech-debt T-003.
- **Depends on:** ADR 0050 (computed-token catalog single source — the catalog this registry realizes).

---

## Context

Backlog row R-003 asked to "record resolver registry/version compatibility policy in ADR." Before writing a policy, the actual mechanism needs to be verified — a policy ADR that describes a compatibility/deprecation scheme not backed by code would itself be a documentation-drift risk.

### Verified runtime facts

- **`Registry` is a flat, unversioned key→resolver map.** `internal/modules/render/resolvers/registry.go` — `Registry.items map[string]ComputedResolver`, `Register(cr ComputedResolver)` keys by `cr.Key()` (line 19, last write wins on a duplicate key — no collision error), `Get(key string)` does a plain map lookup (lines 22-27). There is no registry-level schema version, no registration-time compatibility check, and no rejection of re-registering an existing key with a different `Version()`.
- **`Version()` is a per-resolver integer, carried into every resolved value, with no consumer-side comparison logic.** `internal/modules/render/resolvers/resolver.go:37-41` — the `ComputedResolver` interface requires `Key() string`, `Version() int`, `Resolve(...) (ResolvedValue, error)`. `ResolvedValue.ResolverVer int` (line 32) is populated per-resolution (presumably from the resolver's own `Version()` at the call site — the field exists to be *recorded*, e.g. into a persisted/audited resolution record, not to gate anything at resolve time). `Registry.Known() map[string]int` (`registry.go:29-37`) exposes the current `key → Version()` map, apparently for introspection/diagnostics, not for a compatibility gate — nothing in the `resolvers` package reads `Known()`'s output to reject or warn on a version mismatch.
- **No deprecation, no multi-version coexistence.** A given `Key()` can only have one live implementation in the registry at a time (last `Register()` call wins per key — `registry.go:16-20`). There is no shape for "resolver X version 1 is deprecated, version 2 is now canonical, both resolve for existing data" — a version bump today just means whoever wrote the new resolver bumped the constant; no other code reacts to that.
- **Consumer of the catalog:** ADR 0050 establishes `render/domain.ComputedCatalog()` as the single source of truth for the 8 computed tokens; this registry is the resolution-time implementation registry keyed by the same resolver keys the catalog declares, not a second source of truth.

## Decision

**Record the mechanism as-is: `Version()` is a per-resolver, self-reported integer that is stamped onto each `ResolvedValue` for downstream recording/audit purposes. It is not consulted anywhere to gate compatibility, and the registry does not currently support or need multiple live versions of the same resolver key.**

This ADR does **not** invent a compatibility/deprecation policy that doesn't exist in code — doing so would document an aspirational scheme, not runtime truth. Instead it records the binding rule that keeps the current, simpler shape defensible:

1. **One live implementation per resolver key.** `Registry.Register` intentionally allows overwrite-by-key (no duplicate-registration error) because the expected pattern is "the composition root registers exactly the resolvers that exist today," not "multiple competing versions coexist at runtime." If a resolver's output shape changes in a way that is not backward-compatible for already-persisted `ResolvedValue` rows, that is a **data-migration concern for the consumer of persisted resolutions**, not something the registry itself is responsible for gating.
2. **`Version()` is provenance metadata, not a compatibility gate.** Its purpose is so a persisted/audited `ResolvedValue` can later be traced to "which resolver logic version produced this value" (useful for the eventual archaeology of "why does this old freeze show a different computed value than a new one would"). Bumping a resolver's `Version()` constant when its `Resolve` logic changes is expected practice; consumers reading historical `ResolverVer` values should treat it as an informational stamp, not assume the registry enforces anything about it.
3. **If a real multi-version coexistence need emerges** (e.g. a resolver's output format changes and both old and new consumers must be served simultaneously), that is a new architectural decision — a successor ADR — not a silent extension of `Registry`. The current flat map is deliberately not designed for that; retrofitting it is a bigger change than adding a field.
4. **`Registry.Known()` remains a diagnostics/introspection surface** (e.g. for a startup log line or a debug endpoint listing active resolver versions), not a runtime dependency of the resolution path.

## Consequences

- R-003 (`wiki/backlog/render-fanout-refactor.md`) and its linked tech-debt T-003 are closed by this ADR.
- The next person touching `resolvers/registry.go` has a decision record confirming the flat, overwrite-by-key, no-compatibility-check shape is intentional, not an oversight — they should not add speculative multi-version machinery without a new ADR justified by an actual requirement.
- No migration, schema change, or code change is required by this ADR — it documents and binds existing, verified runtime behavior.

## References

- `internal/modules/render/resolvers/registry.go` — `Registry` type, `Register`/`Get`/`Known`.
- `internal/modules/render/resolvers/resolver.go:28-41` — `ResolvedValue.ResolverVer`, `ComputedResolver` interface.
- `wiki/backlog/render-fanout-refactor.md` R-003 — backlog row closed by this ADR.
- ADR [`0050-computed-token-catalog-single-source.md`](0050-computed-token-catalog-single-source.md) — the catalog this registry implements resolvers for.
