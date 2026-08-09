# ADR 0095 — Search stays live-query today; a derived index is a named, triggered future

- **Status:** Accepted — transitional (SHOULD, with a named promotion trigger)
- **Date:** 2026-08-08
- **Scope:** Splits the former REQ-SEARCH-1 (`wiki/architecture/backend-target-architecture.md`, pre-amendment) into two REQ IDs and rules on the second. REQ-SEARCH-1 (authz independence) stays MUST, already true, already tested — untouched by this ADR beyond the split. This ADR rules on the new REQ-SEARCH-2 (derived, rebuildable index + tested reindex procedure): defers it, names the global-maximum structure it defers, and names the concrete trigger that promotes it back to MUST.
- **Depends on:** none. Amends no prior ADR — no ADR previously ruled on search index architecture.

---

## Context

`wiki/architecture/backend-target-architecture.md:225` (pre-amendment) read:

> **REQ-SEARCH-1** Search indexes are derived and rebuildable; a full reindex procedure exists and is tested. Search is never consulted for authz decisions. (MUST)

This is two unrelated claims yoked into one REQ, with different truth values:

**Claim A — "search is never consulted for authz decisions" — true and tested today.**
`internal/modules/search/infrastructure/v2documents/reader.go` builds one SQL query that joins the visibility predicate directly into the `WHERE` clause (owner / area-grant / user-grant / company-scope, per `reader.go`'s query shape) — the same grant tables the primary document read path uses, not a search-owned decision. `internal/modules/search/application/service.go`'s `SearchDocuments` forwards `ActorUserID`/`TenantID` straight to the reader and makes no allow/deny call of its own (`TestSearchDocumentsForwardsActorAndTenantToReader`). Nothing downstream asks "does search say this is visible?" to decide access; search asks the same question access-control asks, independently, in SQL, every time.

**Claim B — "search indexes are derived and rebuildable" — not true.**
- `grep -rn -i 'reindex\|tsvector' --include=*.go internal/modules/search/` → no matches.
- `reader.go` queries the live `documents` table (via the `v2documents` views) with `LOWER(...) LIKE '%' || $2 || '%' ESCAPE '\'` (`reader.go:75`), escaped through `internal/platform/sqlescape.LikeEscape`. This is a live substring scan over the base tables at read time, not a derived structure. There is nothing to rebuild, and therefore nothing a "reindex procedure" could target.
- Directory listing of `internal/modules/search/infrastructure/v2documents/` confirms no `reindex.go`, `index_builder.go`, backfill job, or staleness-tracking column exists.

Claim B was never built. It is not a regression — nothing in the module's history shows a derived index that was later removed — it describes a design this codebase chose not to build, at a scale where it has not needed to.

## Decision

**Split the REQ.** REQ-SEARCH-1 keeps only Claim A (MUST, already satisfied — see the traceability report for its test citations). The new **REQ-SEARCH-2** carries Claim B, reclassified **SHOULD**, transitional per CLAUDE.md's Global Maximum rule.

### Why SHOULD, not "amend MUST to match reality," and not "build it now"

Downgrading Claim B to match the code without naming what it costs would be exactly the silent softening CLAUDE.md forbids. This ADR instead does what the rule requires: name the design being deferred, name why the current design is adequate *now*, and name a concrete, observable trigger that makes the deferral end.

**Current design and why it is adequate at present scale.** `ILIKE`-over-live-table search on `documents` (a bounded, per-tenant row set, already filtered by the same visibility join used everywhere else) is a correct, simple, always-fresh implementation: there is no index to go stale, no backfill job to fail silently, no reindex procedure to forget to run after a schema change. Its cost is a sequential-ish scan proportional to the visible row count per query; at the row counts this system operates at today, that cost is not a measured problem (no reported timeout, no logged slow-query alert on the search path as of this ADR's date).

**The global-maximum structure this defers.** A derived, rebuildable full-text index — a `tsvector` generated column (or equivalent projection) backed by a GIN index, populated by a **tested full-reindex procedure** that can rebuild the index from the base tables from scratch on demand (the literal text of Claim B). This is the correct end state for search at scale: sub-linear lookup cost independent of table size, with staleness bounded by trigger-maintained generated columns rather than a live scan.

**The promotion trigger — observable, not "when we get to it".** REQ-SEARCH-2 is promoted back to MUST, and the derived-index work becomes required, the first time **either** of these is observed:

1. **Latency signal.** The search list endpoint's REQ-OBS-2-mandated per-route duration histogram shows p99 latency exceeding **800ms sustained for any 5-minute window** in a production or production-like environment. REQ-OBS-2 already mandates this metric exists (`RED metrics per route ... duration histogram`) — this trigger adds no new instrumentation, only a threshold read off metrics this system already commits to having.
2. **Scale signal.** Any single tenant's `documents` row count exceeds **250,000 rows**. This is a real, checkable number (`SELECT count(*) FROM documents WHERE tenant_id = ...`), not a vague "big" — chosen because it is the point past which an unindexed `ILIKE` scan's linear cost growth stops being masked by per-tenant row-set size and starts being visible in the latency signal above; recording it explicitly means the promotion does not have to wait for the latency signal to actually fire before someone can see it coming.

Whichever fires first ends the deferral. This is a fact anyone can check against a dashboard or a `count(*)`, not a judgment call.

## Consequences

- REQ-SEARCH-1 (authz independence) is unambiguously satisfied and stays MUST — no change in behavior, only in which REQ ID owns the claim.
- REQ-SEARCH-2 is honestly uncovered-by-design, not uncovered-by-oversight: it is SHOULD, not MUST, so `scripts/req-trace`'s uncovered-MUST gate does not fail on it, and its evidence pointer (in the traceability report) states plainly that it is not yet built.
- The next engineer who hits a real search latency complaint has a pre-written go/no-go check instead of having to invent one under pressure.
- No code, schema, or migration change. This ADR is a classification and deferral decision only.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Build the `tsvector`/GIN index and reindex procedure now | Rejected — not this task | Real implementation cost (index maintenance triggers or generated columns, backfill migration, staleness handling, a tested rebuild path) with no measured need yet; building it speculatively is exactly the kind of unbounded gold-plating CLAUDE.md's scoping discipline warns against. If the operator judges the cost worth paying ahead of the trigger, that is a product-roadmap call, not something this REQ-ruling task enacts. |
| Silently reclassify Claim B to MUST-satisfied via a `doc-annotation` (e.g. "MUST — satisfied by live ILIKE query") | Rejected | This is the annotation this task's constraints explicitly forbid: it would make a real, load-bearing divergence (no rebuildable index exists) undetectable by the gate while reading as satisfied. |
| Leave REQ-SEARCH-1 as one combined REQ and accept it stays permanently uncovered MUST | Rejected | Conflates a true, tested claim with a false, untested one under one ID — a reviewer citing REQ-SEARCH-1 as evidence could not tell which half they were vouching for. Splitting is not a downgrade of rigor; it is the only way to keep the true claim's MUST status honest while being honest about the false one. |

## References
- `internal/modules/search/infrastructure/v2documents/reader.go` — the live `ILIKE` query and its visibility join.
- `internal/modules/search/infrastructure/v2documents/reader_visibility_integration_test.go` (`TestListDocuments_EnforcesUnifiedVisibility`) — cited evidence for REQ-SEARCH-1.
- `internal/modules/search/infrastructure/v2documents/reader_contract_parity_integration_test.go` — proves the reader's rewrite changed "no authz/visibility/ordering" (its own header comment), cited evidence for REQ-SEARCH-1.
- `internal/modules/search/application/service_test.go` (`TestSearchDocumentsForwardsActorAndTenantToReader`) — proves the application layer makes no independent authz decision, cited evidence for REQ-SEARCH-1.
- `wiki/architecture/backend-target-architecture.md` §7 REQ-OBS-2 — the existing per-route duration-histogram requirement this ADR's latency trigger reads off, without adding new instrumentation.
- `docs/superpowers/analysis/req-disposition-2026-08-07.md` §"REQ-SEARCH-1" — the disposition analysis this ADR resolves.
