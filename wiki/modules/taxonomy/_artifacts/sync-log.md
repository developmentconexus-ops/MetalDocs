# Sync log — taxonomy

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-11 · Plan 6a — close T-004 T-005 T-010

- **Context:** Plan 6a (commits 115cb635 + 20bf2067 + 71a2dc53) · FamilyService gets govLogger; Profile/Area emit; governance_events collapsed onto canonical audit_events
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-004 (commit 115cb635), T-005 (commit 20bf2067), T-010 (commit 71a2dc53) · evidence: FamilyService.govLogger wired; profile/area Create+Update emit; AuditGovernanceAdapter routes all events to auditdomain.Writer
- **R-NNN updated:** R-004 → merged, R-005 → merged, R-010 → merged · commits per row
- **§11 counts after:** Critical=5 Major=5 Minor=6 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/taxonomy-tech-debt.md · wiki/backlog/taxonomy-refactor.md
- **Structural changes noted (sweep needed):** AuditGovernanceAdapter new type in taxonomy/application; AuditWriter field added to Dependencies; govLogger field added to FamilyService — §5 Key Files not yet updated
