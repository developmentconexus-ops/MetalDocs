# Sync log â€” templates

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-15 - Plan 12.4 novo-template-wizard create-path sync

- **Context:** uncommitted Plan 12.4 implementation diff for `/templates/new` plus runtime prerequisite repair in `internal/modules/templates/application/create.go`
- **Mode:** structural refresh
- **Anchors moved:** 1; `CreateTemplate` authz line moved after transaction GUC setup
- **Public surface:** no exported API change; added unexported `setAuthzGUC` helper
- **Routes/API:** no public route or OpenAPI shape change; documented `POST /api/v1/templates` idempotency wrapper and verified create response
- **Runtime flows:** added CreateTemplate runtime flow and Plan 12.4 smoke evidence
- **Persistence:** transaction-local `metaldocs.tenant_id` / `metaldocs.actor_id` GUC setup documented for create path
- **Dependencies:** none
- **T-NNN touched:** T-001 evidence updated for create-path GUC setup; T-009 text corrected to partial wrapper support with replay audit still open
- **R-NNN touched:** R-009 wording updated to replay-audit follow-up
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/modules/templates/_artifacts/sync-log.md
## 2026-05-14 - Plan 12.1 templates screen reality-first sync

- **Context:** commits `12188f98..eea76b14` (plan docs + templates screen implementation + backlog/design notes sync)
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** none (frontend wiring only; no templates backend route/contract delta)
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** R-009 wording updated to replay-audit follow-up
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates/_artifacts/sync-log.md
## 2026-05-11 Â· Plan 6a â€” close T-013

- **Context:** Plan 6a (commit 71a2dc53) Â· AppendAudit now calls canonical auditdomain.Writer instead of inserting to local templates_audit_log
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-013 Â· evidence: Repository.AppendAudit body replaced â€” calls r.audit.Record(ctx, auditdomain.Event{...}); WithAudit setter added; wired in main.go
- **R-NNN updated:** R-013 â†’ merged Â· commit 71a2dc53
- **Â§11 counts after:** Critical=4 Major=6 Minor=4 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates-tech-debt.md Â· wiki/backlog/templates-refactor.md
- **Structural changes noted (sweep needed):** WithAudit setter new on Repository; auditdomain OUT-edge added to templates/repository â€” Â§5 Key Files + Â§8 cross-deps not yet updated

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
