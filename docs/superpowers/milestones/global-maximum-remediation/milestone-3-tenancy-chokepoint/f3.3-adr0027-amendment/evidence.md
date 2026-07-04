# F3.3 evidence — ADR 0027 amendment + live tenancy wiki (docs-only)

> Contract: `../validation-contract.md` §3 + §4. **No code/behavior change** — documents F3.1/F3.2 as
> shipped (committed `de94758a`, `bad74d86`). Authored via subagent from the two `evidence.md` files;
> wiki-curator pass; main session reviewed diffs + anchors and commits.

## What shipped (docs only)

**T1 — ADR 0027 dated amendment.** `wiki/decisions/0027-rls-adoption-sequencing.md` gained
`## Amendment 2026-07-03 (M3 tenancy chokepoint)` **appended after References** — the original
Context/Decision/Consequences body (Wave-Z record) is byte-untouched (no history rewrite). All 5 contract
§3.1 points, each mapped:

| §3.1 point | Amendment location |
|---|---|
| 1. NULL-permissive deliberate & load-bearing (must not be made fail-closed) | "1. The NULL-permissive design is deliberate and load-bearing" |
| 2. Pre-M3 sync↔async asymmetry (API seeded / worker+jobs seeded nothing) | "2. The pre-M3 sync/async asymmetry" |
| 3. How M3 closes it — (a) chokepoint autoseed F3.1, (b) SeedTxTenant F3.2 completing ADR 0054 rule 2, (c) 2 blocking lints + negative RLS proof | "3. How M3 closes it" (a/b/c) |
| 4. Residual sanctioned GUC-unset surface (outbox claim, cross-tenant scans, `idempotency_keys`/`job_leases`) | "4. Residual sanctioned GUC-unset surface" |
| 5. Cross-ref ADR 0054 + M3 folder | "5. Cross-references" |

Plus the full **§4 per-binary posture table** (metaldocs-api / -worker / -jobs / janitors) embedded, and the
**"policy byte-identical, only GUC-seeding coverage changed"** invariant stated explicitly.

**T2 — Live tenancy wiki pages.**
- `wiki/architecture/tenant-context.md`: corrected the stale "row-level Postgres isolation via GUC/RLS
  (tracked per-module as tech debt)" out-of-scope line → now points at amended ADR 0027 (RLS is live, not
  deferred); §4 middleware block shows `platformtenant.WithActorID` (F3.1) beside `WithTenantID` + a note
  that the `TxRunner` chokepoint auto-seeds so handlers/repos no longer hand-seed; new
  `## 9. RLS backstop & GUC seeding (M3)` summary (autoseed / SeedTxTenant / 2 lints / NULL-permissive
  hatch, cross-linked ADR 0027 + 0054 + M3 folder); Cross-links renumbered §8→§10; `Last verified` → 2026-07-03.
- `wiki/concepts/authz-tiers.md`: one surgical "Common pitfalls" note — request paths auto-seed via the
  chokepoint (M3); `SeedTxIdentity`/`SeedTxTenant` remain for system/async paths; stamp → 2026-07-03.

**T3 — wiki-curator pass.** Verified all new-section file:line anchors resolve against current source
(`middleware.go:106-107`, `runner.go:63`+`:94`, `context.go:81`); **corrected one** — the amendment's
reference to the pre-M3 `SeedTxIdentity` pattern `context.go:60` → **`context.go:48`** (real defn line; the
`:60` in the ADR's *original* body is pre-amendment, left untouched as history). Refreshed authz-tiers stamp.
All internal cross-links (0054, 0007, 0022, M3 folder, evidence files, relative ADR paths) confirmed valid.
`wiki/README.md` needed no change (these pages surface via folder READMEs, no new page). No TBD/TODO left.

## Gates (§3.3 exit criteria + named proofs)

| Gate | Result |
|---|---|
| PG-1 — amendment present, dated, all 5 points, 0054 + M3 cross-refs resolve | ✓ (table above; curator confirmed links) |
| PG-2 — `grep -rniE "async .*no .*backstop\|~?85 sites\|only on controlled_documents\|RLS only on\|no RLS backstop"` on the 3 live pages | **0 survivors** (exit 1) |
| PG-3 — wiki-curator verdict | clean (anchors resolve, 1 corrected; stamps refreshed) |
| PG-4 — `git diff --stat` for the feature | **only `wiki/**`** (3 files, +136/−7); no `.go`/`.sql`/migration touched |

## Scope discipline
- **No history rewrite:** amendment appended; ADR original body + all dated `_artifacts/`, `roadmap.md`,
  `wave-*.md`, `reviews/` records left untouched (one stale ref in a frozen `_artifacts/stage1` doc found
  and **deliberately left** — historical artifact, out of scope).
- **No code diff:** verified — the modified tree is exactly the 3 wiki files.

## Contract conformance (§3.3)
Dated amendment, 5 points ✓ · per-binary §4 table carried into the ADR ✓ · live tenancy docs match runtime
truth (no stale claim) ✓ · 0027 ↔ 0054 ↔ M3 cross-refs resolve ✓ · wiki-curator clean ✓ · no code-behavior
change ✓.
