# Phase 6.75 — Self-review (auth)

**Date:** 2026-05-10
**Composer:** main agent

## Checklist

1. **Severity rubric application.** Re-rated each Critical/Major against `templates/tech-debt-register.md`:
   - T-001 (LegacyHeader bypass) → Critical via "authn/authz bypass" trigger. Correct.
   - T-002 (audit-trail gap on identity mutations) → Critical via "regulated audit-trail gap" — identity is QMS-regulated. Correct.
   - T-003 (RFC 9457 drift) → Major via "documented contract not followed with measurable consumer impact". Not Critical (no bypass, no data loss). Correct.
   - T-004 (CreateUser two-tx) → Major. Reviewed for Critical "data-loss path" — orphan rows are recoverable manually and no caller observed today. Held at Major; rationale in row's `Observation`.
   - T-005 (no IP rate limit) → Major via "defense-in-depth gap". Not Critical because per-account lockout is live. Correct.

2. **Mermaid box ↔ prose.** Walked §3 C4Context boxes (auth_module, iam_module, documents_module, postgres_db, browser_client, ops_console) — every one is named in §1, §3 prose, §5, or §8. Walked §5.1 C4Container boxes (delivery_http, application, domain, infra_pg, infra_mem) — all referenced in §5.2 building-block table. No stray boxes.

3. **Top-3 in §11.** Re-checked §11 Top-3: T-001 (Critical, blast = every authenticated route), T-002 (Critical, blast = every regulated mutation), T-005 (Major, blast = login surface). Correct ordering: severity first, then blast-radius. T-002 is broader-blast than T-001 but both Critical → tied; current order acceptable.

4. **Cross-link existence.** Verified each wiki link target exists:
   - `wiki/decisions/0007-two-tier-authz.md` → exists.
   - `wiki/concepts/authz-tiers.md` → exists.
   - `wiki/modules/iam.md` → exists.
   - `wiki/modules/documents.md` → exists.
   - `wiki/backlog/auth-refactor.md` → exists (just authored).
   - `wiki/modules/iam-tech-debt.md` (T-006, T-010 sibling refs) → exists.

5. **Key Files freshness.** Sampled three `path:LL` anchors against current code:
   - `internal/modules/auth/delivery/http/middleware.go:58` → `LegacyHeaderEnabled` branch present. ✓
   - `internal/modules/auth/application/service.go:431` → `bcrypt.GenerateFromPassword` call. ✓
   - `internal/modules/auth/infrastructure/postgres/repository.go:80` → `TouchSession` UPDATE. ✓

6. **Backlog ↔ debt linkage.** Each T-NNN (T-001..T-012) has a matching R-NNN row in `backlog/auth-refactor.md`. tally_check.sh confirms. No orphaned debt rows.

7. **Industry citations.** Three §5 citations: IP-001 (RFC 9457), IP-004 (NIST SP 800-95 §4.3), IP-008 (Crunchy Data tenant_id). All three trace back to existing rows in `references/industry-patterns-index.md`. No fresh additions in this pass.

8. **Subagent purity.** Re-skimmed 02-flow-login.md, 02-flow-resolve-session.md, 02-flow-create-user.md, 03-deps.md, 04-persistence.md. No occurrences of "should", "recommend", "professional", or "industry-standard" in any artifact. Subagent research-only rule held.

## Outcome

All 8 items resolved. No fixes required after self-review beyond the missing-ADR count adjustment caught by tally_check (8 → 9, applied pre-self-review). Ready for Phase 7.
