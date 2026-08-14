# MetalDocs Agent Instructions

## ACTIVE DESIGN RESET — HARD STOP

MetalDocs is currently in the **Cohesive Platform Redesign**. Before any product/domain architecture work, read:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. `wiki/references/current-agent-handoff.md`

**No product code, schema, migration, OpenAPI or frontend implementation is authorized by the redesign yet.**

Historical `docs/superpowers` roadmaps/milestones/specs/plans/reports were intentionally removed from the live tree on 2026-08-14. Do not restore or execute them from Git history. PR #113 is historical/superseded implementation work and must not be resumed by inertia.

## Role

Act as a careful MetalDocs maintainer and design collaborator. Preserve runtime truth when describing the current system, but do not mistake current runtime/module/schema shape for target domain authority.

During the redesign:

- current code/schema/OpenAPI answer **what runs today**;
- `wiki/architecture/cohesive-platform-redesign.md` + the active ledger answer **what the target is becoming**;
- legacy/current-state wiki pages are evidence only.

## Always-On Rules

- Never read, print, commit, or expose `.env` secrets.
- Use PowerShell scripts for local startup; do not use bash or `source .env` for operator startup.
- Keep changes scoped to the request; do not revert user work.
- Stop on architecture contradictions instead of patching around them.
- Evidence before closure: commands/outcomes/review/QA/bounded defers.
- Commits are allowed after verified work; never push without explicit permission.

## Root Cause / Global Maximum

Canonical method: `wiki/standards/root-cause-global-maximum-method.md`.

For non-trivial work, identify root cause, target invariant, authority/boundary, local maximum, global maximum, enforcement and proof before implementation.

Global Maximum means the smallest sustainable structure that preserves correctness and removes accidental complexity. It does **not** mean maximum abstraction or maximum infrastructure.

## Current redesign facts

The following target directions are operator-approved and are summarized here only for orientation; the active ledger contains exact wording and open items.

### Authentication / Organization / Authorization

- AuthN and product AuthZ are separate.
- Current MetalDocs AuthN may remain for V1; external IdP/Keycloak is future-triggered.
- Organization owns Tenant, Area, User, Group, GroupMembership.
- Groups are flat principals and receive ordinary RoleAssignments.
- Built-in V1 roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`.
- Authorization is permission-based scoped RBAC; no tenant-owner bypass.
- OpenFGA/SpiceDB are not V1 dependencies.

### Approval V1

- Specialized governed-document approval, not generic BPM.
- Versioned sequential ApprovalPolicy with ordered ApprovalSteps.
- Initial participants: named user, group, role-in-area.
- Completion: ANY / ALL.
- Human outcomes: `accept` / `return_for_changes`.
- Edited content creates a new approval attempt.
- Audited reassignment is V1; generic delegation/BPMN/CEL/M-of-N are deferred.

### Controlled Information

- `documents`, `controlleddocuments`, `templates` are not three target bounded contexts.
- Target core: `Document` + `DocumentRevision`.
- Separate `ControlledDocument` target object is being retired.
- `DocumentProfile` is converging toward `DocumentType`.
- Area moves to Organization.
- Template is a role/designation of an exact governed DocumentRevision, not a parallel lifecycle/version counter.
- Changing template layout/placeholders/constraints/visibility/resolver semantics requires a new DocumentRevision.
- Derived documents remain bound to the exact template revision/hash used to seed them.
- Freeze, Approval evidence and official Rendition must bind the exact submitted/reviewed Revision/hash.
- Release/effectivity remains downstream of human approval.

## Current repository facts vs target facts

The repository currently still contains 15 module directories. **Do not call all 15 target bounded contexts.** Several are explicitly legacy boundaries under redesign (`approval`, `iam`, `taxonomy`, `documents`, `controlleddocuments`, `templates`, `jobs`).

Healthy supporting responsibilities such as audit evidence, render/rendition infrastructure, search projections, notifications, distribution read models, tokens and security should be revalidated rather than rewritten automatically.

## Stable engineering constraints

Unless the redesign explicitly changes them, continue to respect:

- contract-first OpenAPI/generated types;
- RFC 9457 problem details;
- pooled multi-tenant isolation and RLS defense-in-depth;
- transactional outbox / idempotent async consumers;
- database constraints for invariants where appropriate;
- no foreign-context persistence mutation in the final architecture;
- audit/evidence preservation.

## Commands

- Start API: `.\scripts\start-api.ps1`
- Rebuild/start API: `.\scripts\start-api.ps1 -Build`
- System runnable check: `.\scripts\check-system-runnable.ps1`
- Go build: `go build ./...`
- Go tests: `go test ./...`
- Frontend tests: `make test`
- Docx workspace: `npm run build:docx-v2`, `npm run test:docx-v2`, `npm run typecheck:docx-v2`

These commands are runtime/verification references only. The current redesign is not at implementation stage.

## Context Map

| Task | Read |
|---|---|
| Cohesive product/domain redesign | `AGENTS.md` → Global-Maximum method → `wiki/architecture/cohesive-platform-redesign.md` → active ledger → handoff |
| General wiki orientation | `wiki/index.md` |
| Backend/API contract evidence | `wiki/architecture/backend-api-structure.md`, `api-contract.md`, `api-design-system.md` |
| Frontend evidence | `wiki/architecture/frontend-structure.md` |
| Current DB/migrations | `wiki/database/index.md` + current schema/migrations |
| Current module implementation evidence | `wiki/modules/index.md` then owning LEGACY/current-state page/code |
| QA/close-out | `wiki/quality/qa-operating-system.md` + relevant checklist |
| Test discipline | `wiki/quality/test-discipline.md`, ADR 0034 |
| Adversarial design review | `.claude/skills/adversarial-review/SKILL.md` |
| Code relationship/impact tracing | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Docs governance | `wiki/standards/documentation-governance.md` |

## Workflow while redesign gate is closed

1. Read active redesign authority.
2. State the domain/product question.
3. Inspect current code only as evidence.
4. Compare mature products/standards/libraries where useful.
5. Apply Root-Cause / Global-Maximum + YAGNI.
6. Present alternatives and recommend the smallest correct target.
7. Record operator-approved decisions in the active ledger.
8. Continue to the next dependency.
9. **Do not implement product code.**

## Exact next design step

Continue with:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

Only after that is approved proceed to Document/Revision lifecycle design.
