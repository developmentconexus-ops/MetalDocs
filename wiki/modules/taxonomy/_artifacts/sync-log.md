# Sync log — taxonomy

**Last verified:** 2026-06-11 (Stage-1 backend audit drift patch)

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-06-10 — Stage-1 backend audit drift patch

- **Context:** Stage-1 mapper found 8 mismatches between wiki text and current code.
- **Mode:** lite patch (surgical corrections only; no restructuring).
- **Affected files:** `wiki/modules/taxonomy.md`; `wiki/modules/taxonomy-tech-debt.md`; `wiki/modules/taxonomy/_artifacts/sync-log.md`.
- **Corrections applied:**
  1. Key files line: `FamilyService` — updated struct anchor `:13-19`; corrected description to reflect `govLogger` field + `NewFamilyService(families, govLogger)` (T-004 closed).
  2. Key files line: `ProfileService` — corrected Create `:70` and Update `:96` both call `s.govLogger.Log` (T-005 closed).
  3. Key files line: `AreaService` — corrected Create `:59` and Update `:98` both call `s.govLogger.Log` (T-005 closed).
  4. Key files line + §8.2 + T-007: `HasActiveProfiles` — replaced with `HasActiveProfilesTx` which takes `tenantID`; `WHERE tenant_id=$1 AND family_code=$2` present; T-007 TOCTOU resolved (single tx + FOR UPDATE). Anchor corrected to `infrastructure/family_repository.go:218-240` (method definition line 218; query lines 230-233). Prior stale anchor `153-178` pointed to the non-transactional `HasActiveProfiles` variant (lines 156-178), not to `HasActiveProfilesTx`.
  5. §6.3 sequence diagram + §8.7 + T-007: `deactivateFamily` — redrawn to show `BeginTx`/`GetByCodeForUpdate`/`HasActiveProfilesTx`/`UpdateTx`/Commit; removed "NO tx · NO row lock" annotations.
  6. §2 + §4 + §5.2 + §9 + T-009: removed "No OpenAPI spec / raw ServeMux" claims; added `internal/modules/taxonomy/api/` (api.gen.go, cfg.yaml, gen.go), `handler.go:42-51` HandlerWithOptions call, `routes_generated.go:10` ServerInterface assertion; T-009 marked closed.
  7. §3.2 + §5.2 + failure-modes: `registry/module.go:31` corrected to `internal/modules/controlleddocuments/module.go:37`.
  8. Key files + §8.9: `main.go:197-201,225,508-524` corrected to `main.go:314-315,358,412,908-924`.
  9. §1.2 + §4 + §8.8 + §10: migration `0122` trigger anchor corrected from `0122:33-39` to `0122:35-49`. Actual `reject_code_update()` function definition starts at line 35; `DROP TRIGGER / CREATE TRIGGER trg_document_profiles_code_immutable` block is lines 46-49. Prior anchor `33-39` fell inside the unique-index comment and start of a different block.
  10. §1.2 + §4 + §8.8: migration `0123` trigger anchor corrected from `0123:33-37` to `0123:61-75`. Actual `reject_code_update()` function definition starts at line 61; `DROP TRIGGER / CREATE TRIGGER trg_process_areas_code_immutable` block is lines 72-75. Prior anchor `33-37` fell inside the FK-constraint `DO $$` block.
- **T-NNN status changes:** T-004 confirmed closed; T-005 confirmed closed; T-007 marked resolved; T-009 marked closed.
- **§11 severity counts:** no count change (all items were already closed in the register text; only Key Files and behavioral prose were stale).
- **Tally gate:** drift patch only — no implementation change; preflight not applicable.

## 2026-05-17 - active v2 reference memory sync

- **Context:** post-merge scan found active taxonomy module memory still naming the documents consumer as `documents_v2` after the production module/API naming moved to `documents` and `/api/v1`.
- **Mode:** lite patch.
- **Affected-surface scan:** taxonomy module doc + tech-debt register only; historical migration/drop facts left unchanged.
- **Routes/API:** no taxonomy route change.
- **Runtime flows:** downstream consumer wording corrected from `documents_v2` to `documents`.
- **Persistence:** none.
- **Debt/backlog:** no T/R rows opened or closed.
- **Tally gate:** preflight PASS before edits; final tally recorded in session output.
- **Patched files:** `wiki/modules/taxonomy.md`; `wiki/modules/taxonomy-tech-debt.md`; `wiki/modules/taxonomy/_artifacts/sync-log.md`.

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

## 2026-05-13 - Plan 10 taxonomy route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (/api/v2/taxonomy* to /api/v1/taxonomy*)
- **Mode:** structural refresh
- **Anchors moved:** taxonomy endpoints canonicalized to /api/v1
- **Public surface:** no semantic capability model change
- **Routes/API:** route truth/artifacts refreshed to v1 paths
- **Runtime flows:** unchanged behavior
- **Persistence:** none
- **Dependencies:** permission resolver checks aligned to v1 paths
- **T-NNN touched:** T-013/T-015 evidence continuity maintained
- **R-NNN touched:** R-013/R-015 plan linkage maintained
- **Counts after:** Critical=5 Major=5 Minor=6; missing-ADR=14
- **Tally gate:** PASS
- **Patched files:** wiki/modules/taxonomy.md; wiki/modules/taxonomy-tech-debt.md; wiki/backlog/taxonomy-refactor.md; wiki/modules/taxonomy/_artifacts/*
