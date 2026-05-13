# Phase 6.75 — Self-Review (iam)

Run after Phase 6.5 `tally_check.sh` PASS. Catches what mechanical grep cannot: severity judgment, mermaid box drift, rubric application.

**Reviewer:** main agent (Opus 4.7), same session
**Date:** 2026-05-11
**Composed doc state at review:** Plan 4 structural sweep — deleted AuthorizationService, area_membership/ pkg, RoleCapabilities map; renamed authz.ErrCapDenied; collapsed capability namespaces. Tally: 2/5/5 (all items including closed); missing-ADR: 12/12.

## Checklist (8 items)

### 1. Severity rubric application — every Critical / Major row re-checked

| Row | Severity | Trigger fired | Verdict |
|---|---|---|---|
| T-001 dual capability namespaces | Critical (closed) | "contract violation downstream depends on" — resolved in Plan 4; typed `Capability` namespace is now single source. | Correct (closed) |
| T-005 missing audit on role upsert | Critical | "regulated audit-trail gap: a mutation on an ISO 9001 / QMS path is not written to the audit sink" | Correct |
| T-002 two area-membership write surfaces | Major (closed) | "duplicated write surfaces with different semantics" — resolved in Plan 4; `area_membership/` wrapper deleted. | Correct (closed) |
| T-003 `AuthorizationService` unused-but-exported | Major (closed) | "duplicated write surfaces" (third authz surface) — resolved in Plan 4; deleted. | Correct (closed) |
| T-004 IAM mutations have no tier-2 or tripwire | Major | "defense-in-depth single-point gap" — tier-1 only on regulated path | Correct. Borderline Critical but tier-1 is in place; hold at Major. Escalate if a tier-1 bypass like J2 recurs. |
| T-006 IAM error envelope not RFC 9457 | Major | "documented contract not yet followed by this module" | Correct |
| T-007 governance logger wired with `nil` | Major | "governance / observability sink wired to `nil` on a regulated path" | Correct |

Minor rows: T-008 (latent cache gap), T-009 (collision — closed Plan 4), T-010 (bidirectional dep, non-circular), T-011 (missing-ADR for enforced rule), T-012 (in-process map — closed Plan 4). All match Minor triggers. No changes.

### 2. Mermaid box ↔ prose

| Diagram | Boxes | Each named in prose? |
|---|---|---|
| §3 Context | `actor`, `iam`, `docs`, `auth`, `audit`, `platform`, `db` | yes — §3.1 / §3.2 / §8.4 reference each |
| §5.1 Container | `mw`, `adminH`, `memH`, `adminSvc`, `memSvc`, `capSvc`, `authz`, `roleProv`, `db` | yes — §5.2 surface table covers every container. `authzSvc` and `memberPkg` removed in Plan 4 sweep. |
| §6.1 / §6.2 / §6.3 sequence | Standard layers | yes |

One residual concern: §3 collapses `documents/approval/templates` into a single `System` box. When approval gets its own module doc, refactor this into two boxes (one for documents+approval, one for templates). Not a current debt — log as `_artifacts/00-context.md` note for next module sweep.

### 3. Top-3 in §11 ordered by severity-then-blast-radius

T-001/T-002 closed. Current order: T-005 (Critical — regulated audit gap), T-004 (Major — IAM mutations single-layer defense, broad blast-radius), T-006 (Major — RFC 9457 gap, all HTTP error surfaces). Correct by severity-then-blast-radius.

### 4. Cross-link existence

Verified earlier in review session: `wiki/concepts/authz-tiers.md`, `wiki/concepts/iso-segregation.md`, `wiki/modules/approval.md`, `wiki/modules/documents.md`, `wiki/modules/templates.md`, `wiki/decisions/0007-two-tier-authz.md`, `wiki/decisions/0012-contract-first-api.md`, `wiki/architecture/api-design-system.md`, `wiki/architecture/api-contract.md` all exist. `wiki/modules/iam-rbac.md` correctly removed (R-014).

### 5. Key Files freshness

Sample-verified 2026-05-11 (post-Plan 4):
- `capability_service.go:31` → `func (s *CapabilityService) CanDo(...)` ✓
- `authz/authz.go:44` → `func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string)` ✓
- `authz/context.go:13` → `var ErrActorContextMissing = errors.New(...)` ✓

3/3 pass. Anchors trustworthy.

### 6. Backlog ↔ debt linkage

12 T-NNN rows in register → 12 R-NNN backlog rows + R-013/R-014 (maint:doc-cleanup). Tally PASS. Closed rows (T-001/T-002/T-003/T-009/T-012) marked `merged` in backlog with Plan 4 commit refs. ✓

### 7. Industry citations

`_artifacts/05-industry.md` admits IP-001, IP-004, IP-005, IP-006, IP-008. All present in `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. Zero new patterns introduced. Three additional patterns (IP-002, IP-003, IP-007) explicitly rated not-applicable with one-line rationale. Clean.

### 8. Subagent purity

Spot-check on subagent artifacts for "should / recommend / professional / industry-standard":
- `01-surface.md` — factual symbol table, no prescriptions. Clean.
- `02-flow-*.md` (3 files) — sequence + SQL + tripwire-pairing rows. Clean.
- `03-deps.md` — import tables only. Clean.
- `04-persistence.md` — table/trigger/migration tables. Clean.
- `05-industry.md` uses "full / partial / gap" — factual ratings against the index, not prescriptions. Clean.

No subagent leak.

## Verdict

PASS. Doc ready for Phase 7 (wiki-curator).

## Notes for next module sweep

- §3 Context diagram pattern: when grouping ≥3 distinct consumer modules into one box, log it for the day the grouped modules each get their own doc.
- `tally_check.sh` does not yet validate the "Top 3 ordering" rule (severity, then blast-radius). Possible enhancement: enforce that the first numbered item in the Top 3 list has severity Critical if any Critical row exists in the register.
- T-004 borderline call (Major vs Critical) should be re-audited if a tier-1 bypass bug recurs in the IAM surface.
