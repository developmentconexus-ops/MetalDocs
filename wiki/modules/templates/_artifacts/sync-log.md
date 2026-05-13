# Sync log — templates

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-11 · Plan 6a — close T-013

- **Context:** Plan 6a (commit 71a2dc53) · AppendAudit now calls canonical auditdomain.Writer instead of inserting to local templates_audit_log
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-013 · evidence: Repository.AppendAudit body replaced — calls r.audit.Record(ctx, auditdomain.Event{...}); WithAudit setter added; wired in main.go
- **R-NNN updated:** R-013 → merged · commit 71a2dc53
- **§11 counts after:** Critical=4 Major=6 Minor=4 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates-tech-debt.md · wiki/backlog/templates-refactor.md
- **Structural changes noted (sweep needed):** WithAudit setter new on Repository; auditdomain OUT-edge added to templates/repository — §5 Key Files + §8 cross-deps not yet updated

## 2026-05-13 - Plan 10 route canonicalization + templates rename sweep

- **Context:** uncommitted Plan 10 implementation diff (module rename, API v1 sweep, final permission remediation)
- **Mode:** structural refresh
- **Anchors moved:** internal/modules/templates -> internal/modules/templates; /api/v2/templates* -> /api/v1/templates*
- **Public surface:** updated route prefix and capability mapping references
- **Routes/API:** templates and signed endpoints documented under /api/v1
- **Runtime flows:** unchanged behavior; canonical path updates only
- **Persistence:** none
- **Dependencies:** composition root + permission resolver references updated
- **T-NNN touched:** T-012/T-010 closure evidence aligned for Plan 10 sweep
- **R-NNN touched:** R-100/R-101 alignment notes refreshed
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/modules/templates/_artifacts/*
