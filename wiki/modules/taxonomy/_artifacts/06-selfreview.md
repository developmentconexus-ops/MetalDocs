# Phase 6.75 — Self-review (taxonomy)

Run against the composed `wiki/modules/taxonomy.md`, `taxonomy-tech-debt.md`, `backlog/taxonomy-refactor.md` on 2026-05-11.

1. **Severity rubric application.**
   - T-001 (tenant header trust) — Critical. Rubric trigger: "multi-tenant data leak". ✓
   - T-002 (global families, no ADR) — Critical. Trigger: "multi-tenant data leak + cross-tenant write surface on regulated catalog". ✓
   - T-003 (PATCH bypass) — Critical. Trigger: "authn/authz bypass". ✓
   - T-004, T-005 (no governance emit on regulated mutations) — Critical. Trigger: "regulated audit-trail gap". ✓ Compare audit T-001/T-004 rubric calls.
   - T-006 (single-tier defense) — Major. Trigger: "defense-in-depth gap". Not escalated to Critical because the tier-1 layer exists and is functional for routes other than the PATCH bypass; the PATCH bypass is its own row (T-003). ✓
   - T-007 (TOCTOU + cross-tenant SELECT) — Major. Latent (no observed exploit; bug visible in code). ✓ Considered Critical given cross-tenant SELECT — left Major because the read is on an EXISTS aggregate (no row content leaks).
   - T-008, T-009, T-010 — Major. Trigger: "documented contract not followed with measurable consumer impact". ✓
   - T-011..T-016 — Minor. Latent / docs / missing-ADR. ✓
   - No rating downgraded against rubric.

2. **Mermaid box ↔ prose.** Three diagrams (Context, Container, three Sequence). All boxes named in prose:
   - Context: taxonomy, registry, documents_v2, documents, templates, pg, admin — all referenced in §1, §3.1, §3.2, §8.9.
   - Container: http, svc, domain, repo, db, tplv2 — all referenced in §5.1 prose + §5.2 table.
   - Sequences: every participant resolves to a file in Key Files.

3. **Top-3 in §11.** Ordered by severity, then blast radius:
   1. T-001 (any tenant cross-write) — widest blast.
   2. T-002 (families global, every tenant) — also wide blast, but predicated on `taxonomy.manage` cap.
   3. T-003 (PATCH bypass) — Critical but narrowest (only families/{code} PATCH route).
   Order defensible.

4. **Cross-link existence.** Verified each path:
   - `wiki/decisions/0007-two-tier-authz.md` ✓
   - `wiki/decisions/0010-soft-archive-via-timestamp.md` ✓
   - `wiki/decisions/0012-contract-first-api.md` ✓
   - `wiki/concepts/authz-tiers.md` ✓
   - `wiki/concepts/controlled-documents.md` ✓
   - `wiki/concepts/iso-segregation.md` ✓
   - `wiki/concepts/error-ux.md` ✓
   - `wiki/modules/controlled-documents.md` ✓
   - `wiki/modules/documents.md` ✓
   - `wiki/modules/templates.md` ✓
   - `wiki/modules/audit.md` ✓

5. **Key Files freshness.** Spot-checked:
   - `internal/modules/taxonomy/delivery/http/handler.go:51-68` — matches grep `HandleFunc` output. ✓
   - `apps/api/cmd/metaldocs-api/permissions.go:158-180` — matches the dispatcher switch. ✓
   - `internal/modules/taxonomy/delivery/http/routes_profiles.go:197-203` — matches `tenantIDFromRequest` per Phase 2 artifact. ✓

6. **Backlog ↔ debt linkage.** Every T-001..T-016 has a matching R-NNN in the backlog. R-017 is a `maint:docs-link` row with no T- origin (allowed per `maint:<kind>` enum). Tally check 6.5 PASS confirms.

7. **Industry citations.** §5 of `_artifacts/05-industry.md` cites IP-001, IP-004, IP-005 (via N/A), IP-006, IP-008 — all from `references/industry-patterns-index.md`. No new patterns added.

8. **Subagent purity.** Re-skimmed `_artifacts/02-flow-*.md`, `03-deps.md`, `04-persistence.md`:
   - "should" / "recommend" / "professional" / "industry-standard" — grep returned zero matches in 02-* and 04-*.
   - `03-deps.md` (Sonnet output) — clean of advisory prose; observations only.
   - `02-flow-deactivate-family.md` contains the phrase "**Race window (TOCTOU)**" as a fact descriptor — not a recommendation. Acceptable.

## Action items from self-review
None. No fixes triggered. Tally re-run not needed.

