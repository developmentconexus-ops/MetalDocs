# IAM module doc — sync log

One line per `metaldocs-module-doc-sync` run. Append-only.

## 2026-05-11 · Plan 6a — close T-005

- **Context:** Plan 6a (commit f27529e8) · emit recordAudit in handleUserRoleUpsert + handleCreateUser
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-005 · evidence: handleUserRoleUpsert now calls h.recordAudit after writeJSON; handleCreateUser emits auth.user.created event
- **R-NNN updated:** R-005 → merged · commit f27529e8
- **§11 counts after:** Critical=2 Major=5 Minor=5 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam-tech-debt.md · wiki/backlog/iam-refactor.md
- **Structural changes noted:** none for iam (behavioral change in existing handlers only)

## 2026-05-11 · Plan 4 — capability namespace collapse + IAM dual-surface consolidation

- **Context:** Plan 4 tasks 1-9 completed: deleted capabilities.go, role_capabilities.go, authorization.go, startup.go, area_membership/; renamed authz.ErrCapabilityDenied→ErrCapDenied; extended model.go to 18 typed Capability consts; migration 0186 reseeded doc.*→document.*
- **Anchors moved:** 8+ (startup.go:15 DELETED; area_membership/area_membership.go DELETED; authorization.go DELETED; role_capabilities.go DELETED; authz/authz.go ErrCapabilityDenied→ErrCapDenied)
- **Symbols renamed:** 1 (authz.ErrCapabilityDenied→authz.ErrCapDenied — all occurrences in doc trio updated)
- **T-NNN closed:** T-001, T-002, T-003, T-009, T-012 · evidence: referenced files deleted; typed Capability namespace unified in model.go
- **R-NNN updated:** R-001→merged, R-002→merged, R-003→merged, R-009→merged, R-012→merged · PR: Plan 4 (2026-05-11, commits 3a227642/8da32dbf/a66a8d62/ec7d151a/0cd2e75d)
- **§11 counts after:** Critical=1 Major=3 Minor=3
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam.md · wiki/modules/iam-tech-debt.md · wiki/backlog/iam-refactor.md

- 2026-05-11 · Plan 3 (session-bound tenant resolution, post-merge sweep). Patched anchors shifted by ~3 lines in `admin_handler.go` and `middleware.go` / `routes_memberships.go` (file growth from `tenant.FromContext` migration). Files: `wiki/modules/iam.md` (§2 + §6.4 envelope anchors :129→:132, :137→:150); `wiki/modules/iam-tech-debt.md` (T-005 :316→:319/:457→:454; T-006 :129→:132/:137→:150; Last verified bump); `wiki/backlog/iam-refactor.md` (Last verified bump); `wiki/README.md` (iam-tech-debt + iam-refactor index stamps). T-NNN affected: T-005, T-006 (anchors only — severity unchanged, debt not resolved). R-NNN affected: none. Escalation: no — verified no Plan 3 ADR exists in `wiki/decisions/` (flagged to caller).

## 2026-05-13 - Plan 10 IAM memberships route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (/api/v2/iam/area-memberships -> /api/v1/iam/area-memberships)
- **Mode:** structural refresh
- **Anchors moved:** memberships route canonicalized to /api/v1
- **Public surface:** no semantic change
- **Routes/API:** IAM route references/artifacts updated to v1
- **Runtime flows:** unchanged behavior
- **Persistence:** none
- **Dependencies:** permission resolver path mapping aligned
- **T-NNN touched:** IAM T-010 note remains docs-only and deferred ADR linkage
- **R-NNN touched:** IAM backlog wording canonicalized
- **Counts after:** Critical=2 Major=5 Minor=5; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam.md; wiki/modules/iam-tech-debt.md; wiki/backlog/iam-refactor.md; wiki/modules/iam/_artifacts/*
