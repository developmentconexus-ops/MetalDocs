# Audit — Phase 6.75 self-review

Composer: main agent (Opus 4.7). Date: 2026-05-10. Re-read of `wiki/modules/audit.md` + `audit-tech-debt.md` + `backlog/audit-refactor.md` against the five `_artifacts/` files.

## Checklist

1. **Severity rubric application.** Re-checked every Critical/Major row against `templates/tech-debt-register.md` triggers and the user-supplied audit rubric.
   - T-001 (Critical) — authn/authz bypass trigger fires. Correct.
   - T-004 (Critical) — user rubric "audit-trail tampering path = Critical". Correct.
   - T-005 (Major, not Critical) — user rubric: "missing audit on regulated write = Major in audit register IF consumer-side, Critical IF audit itself drops events." The drop is consumer-side (callers discard error). Major is correct; consumer-side Critical rows live in auth T-002, iam T-005, documents T-005 — cross-linked in §11.
   - T-002 (Major), T-003 (Major), T-007 (Major) — contract violation / regulated governance / latent multi-tenant. Correct.
   - All Minors (T-006, T-008..T-012) — latent/docs/missing-ADR. Correct.

2. **Mermaid box ↔ prose.** Three diagrams (§3 Context, §5.1 Container, §6.1/6.2/6.3 sequences).
   - §3 boxes: `audit`, `iam`, `documents`, `documents/export`, `auth`, `Postgres`, `Admin`. All named in prose (§1 + §3.2 + §11). `auth` appears with "MISSING (auth T-002)" — explained in §11 Top-3 cross-link.
   - §5.1 boxes: `HTTP Handler`, `Application Service`, `Domain Port`, `Postgres adapter`, `Memory adapter`, `metaldocs.audit_events`. All in §5.2 surface table + §4 strategy.
   - §6 sequence participants: `Consumer Module`, `auditdomain.Writer`, `postgres.Writer`, `metaldocs.audit_events`, `Handler.handleEvents`, `Service.ListEvents`, `iam.AdminHandler.handleAdminOverview`, `auditdomain.Reader`, `Admin`. All anchored in §5.2 or §8.
   - No stray boxes.

3. **Top-3 in §11.** Ordered by severity then blast-radius: T-001 (Critical, broad — every reachable network actor), T-004 (Critical, full-history — tampering invisible), T-005 (Major, regulated-paths — silent drops). Order respects rubric, not authorship.

4. **Cross-link existence.** Verified `ls`-style:
   - `wiki/decisions/0007-two-tier-authz.md` ✓
   - `wiki/decisions/0012-contract-first-api.md` ✓
   - `wiki/decisions/0011-cd-atomic-create.md` ✓
   - `wiki/concepts/iso-segregation.md` ✓
   - `wiki/concepts/error-ux.md` ✓
   - `wiki/concepts/authz-tiers.md` ✓
   - `wiki/modules/auth-tech-debt.md` ✓
   - `wiki/modules/iam-tech-debt.md` ✓
   - `wiki/modules/documents-tech-debt.md` ✓
   - `wiki/backlog/audit-refactor.md` ✓ (new this session)
   - `wiki/modules/audit-tech-debt.md` ✓ (new this session)

5. **Key Files freshness.** Sampled 3 anchors:
   - `internal/modules/audit/domain/port.go:8-31` — `Event` starts at line 8, ends line 31 (last interface). ✓
   - `internal/modules/audit/application/service.go:18` — `ListEvents` declared at line 18. ✓
   - `migrations/0005_grant_workflow_audit_privileges.sql:2` — INSERT grant on line 2. ✓
   - Plus: `apps/api/cmd/metaldocs-api/main.go:193` (`auditHandler.RegisterRoutes(mux)`) confirmed by Phase 2 artifact §2c.
   - Plus: `apps/api/cmd/metaldocs-api/main.go:445-479` (`documentsV2AuditAdapter`) confirmed by Phase 3 §2c.

6. **Backlog ↔ debt linkage.** R-001..R-012 each carry a `T-NNN` debt_id matching the register's T-001..T-012. One-to-one. Verified by tally_check.sh (PASS).

7. **Industry citations.** Every §5 (industry) row anchored to an index ID: IP-001, IP-004, IP-006, IP-008. Two un-cited concerns (tamper-evidence, retention) are explicitly called out in `_artifacts/05-industry.md` as "Patterns deliberately NOT cited" — neither appears as an "industry standard says…" sentence in the final doc; they appear only as recorded facts in the tech-debt register.

8. **Subagent purity.** Re-skimmed all 5 artifacts for "should / recommend / professional / industry-standard":
   - `02-flow-record.md` — none.
   - `02-flow-list.md` — none. (Wording is factual; mentions "limit clamp" without prescription.)
   - `02-flow-admin-read.md` — none.
   - `03-deps.md` — none. (One "NOTE:" annotation about same `*Writer` serving both interfaces; factual.)
   - `04-persistence.md` — none. (One footnote "side-effect sinks are out of the tripwire model" — factual classification, not a prescription.)
   - No re-dispatch needed.

## Tally check

`bash .claude/skills/metaldocs-module-doc/scripts/tally_check.sh audit` → `[tally] PASS` (severities 2/4/6 match register; backlog debt_id column clean).

## Verdict

No fixes required. Doc, register, and backlog are coherent and traceable. Phase 7 (wiki-curator) clear to proceed.
