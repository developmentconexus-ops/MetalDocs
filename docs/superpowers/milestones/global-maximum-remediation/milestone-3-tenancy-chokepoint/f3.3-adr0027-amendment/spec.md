# F3.3 — ADR 0027 + wiki amendment (document the NULL-permissive design + async posture)

> **Milestone:** M3 · **Binding contract:** `../validation-contract.md` §3 + §4 (per-binary table).
> Contract governs on conflict (HS-7). **Approval:** approved 2026-07-03. **No code-behavior change.**

## Consumer contract (who consumes what)

Consumer: a future maintainer/reviewer reading the tenancy decision record. Today the NULL-permissive RLS
design lives only in a migration comment; the pre-M3 sync↔async asymmetry is undocumented. The consumer
needs the durable ADR + wiki to state the design, the closed gap, and the residual sanctioned surface — so
no one re-derives it from source or "fixes" the NULL-permissive escape hatch as a bug.

**Required end-state:**
1. Dated amendment appended (history NOT rewritten) to `wiki/decisions/0027-rls-adoption-sequencing.md`
   with all **5 points** (contract §3.1): NULL-permissive is deliberate & load-bearing; the pre-M3
   sync-seeds / async-seeds-nothing asymmetry; how M3 closes it (chokepoint autoseed + `SeedTxTenant` +
   two lints + negative proof, completing ADR 0054 rule 2); the residual sanctioned GUC-unset surface;
   cross-refs to ADR 0054 + the M3 folder.
2. Tenancy wiki pages updated so NO stale claim survives: no "async has no RLS backstop", no "seed
   manually at ~85 sites", no "RLS only on controlled_documents + audit_events". Reflect chokepoint
   autoseed, async tenant-seed, the two lints, per-binary posture (§4 table).
3. `wiki-curator` pass clean (stamps refreshed, file:line anchors resolve).

## Non-goals (mandatory)
- **No code change, no behavior change** — docs only.
- **No history rewrite** — append a dated amendment; do not edit the original ADR decision body.
- **No RLS/policy claim invention** — every statement traces to F3.1/F3.2 shipped behavior + contract §4.
- **No new ADR** — this amends 0027; F3.1/F3.2 needed no new ADR (0054 already sanctions the async seam).

## Validation gate
- **PG-1:** ADR 0027 amendment present, dated, all 5 points, 0054 + M3 cross-refs resolve.
- **PG-2:** grep the wiki for the stale claims above ⇒ 0 surviving; per-binary posture matches §4.
- **PG-3:** `wiki-curator` verdict clean (no broken anchors, stamps refreshed on touched docs).
- **PG-4:** no `.go`/SQL/migration file changed by this feature (docs-only diff).

## Named proof commands
- `grep -rniE "async .*no .*backstop|~?85 sites|only on controlled_documents" wiki` → 0
- `wiki-curator` agent pass (clean)
- `git diff --stat` for the feature → only `wiki/**` touched

## Interview record

| Q | A |
|---|---|
| Amend or new ADR? | Amend 0027 (it owns RLS adoption sequencing); dated amendment, no history rewrite. |
| Which wiki pages? | `wiki/decisions/0027-*`, tenancy/architecture pages describing seeding/RLS, `wiki/concepts/authz-tiers.md` if it mentions seeding — curator finds the rest. |
| Any behavior change? | None — F3.3 is documentation of F3.1/F3.2 shipped truth. |
| Who validates wiki truth? | `wiki-curator` agent (stamps + anchors); contract §4 per-binary table is the truth to match. |
