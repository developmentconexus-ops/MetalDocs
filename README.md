# MetalDocs

MetalDocs is a governed operational-information platform for organizations that need controlled, versioned, reviewable and auditable procedures, policies, instructions, forms and records.

## Current project state — 2026-08-14

MetalDocs is in a **Cohesive Platform Redesign** before the next large product implementation wave.

The redesign is deliberately design-first: product/domain semantics, invariants, authorization, approval, Controlled Information, supporting concerns and migration strategy are being closed **before** code/schema/API implementation resumes.

**No product implementation is currently authorized by the redesign.**

## Read first

For any non-trivial work:

1. [`AGENTS.md`](AGENTS.md)
2. [`wiki/standards/root-cause-global-maximum-method.md`](wiki/standards/root-cause-global-maximum-method.md)
3. **[`wiki/architecture/cohesive-platform-redesign.md`](wiki/architecture/cohesive-platform-redesign.md)**
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. [`wiki/references/current-agent-handoff.md`](wiki/references/current-agent-handoff.md)

The canonical wiki landing page is [`wiki/index.md`](wiki/index.md).

## Approved target direction so far

### Organization / Authorization

- Tenant, Area, User, flat Group and GroupMembership.
- Built-in V1 roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`.
- User/Group RoleAssignments scoped to Tenant or Area.
- Permission-based authorization; no tenant-owner bypass.
- Current AuthN may remain for V1; external IdP/Keycloak is future-triggered.

### Approval V1

- Versioned sequential governed-information ApprovalPolicy.
- Ordered human steps.
- Named user / group / role-in-area participant rules.
- ANY / ALL completion.
- `accept` / `return_for_changes` decisions.
- Audited reassignment and optional step reauthentication.
- Not a generic BPM engine.

### Controlled Information

- `documents`, `controlleddocuments` and `templates` do **not** survive as three target bounded contexts.
- Target core is `Document` + `DocumentRevision`.
- Separate `ControlledDocument` target concept is being retired.
- `DocumentProfile` is converging toward `DocumentType`.
- Area belongs to Organization.
- Template is a designation/role of an exact governed DocumentRevision, not a parallel lifecycle/version counter.
- Changing template layout/placeholders/schema/constraints/visibility/resolver semantics creates a new DocumentRevision.
- Derived documents remain bound to the exact source template revision/hash used to seed them.
- Approval, freeze and official Rendition must bind the exact revision/hash reviewed by the human.
- Release/effectivity remains downstream of human approval.

## Product scope still being designed

Before implementation the program still must close:

- DocumentType / Family / GovernanceClass / TemplateDesignation;
- Document + Revision lifecycle and immutable submission evidence;
- numbering / NumberSeries;
- periodic review / reason-for-change;
- renditions/rendering/reconstruction evidence;
- release/effectivity/supersession;
- distribution/read/acknowledgement;
- tokens/computed-value snapshot semantics;
- audit/evidence;
- notifications/search;
- tenant lifecycle/security/external IdP trigger;
- final Permission Catalog + role bundles;
- bounded contexts, data model, table/transaction ownership, events, APIs, frontend journeys;
- explicit delete/move/rename/rewrite map from current code;
- final ADR/spec set and implementation plan.

## Documentation reset

The previous accumulation of `docs/superpowers` roadmaps, milestones, plans, reports, specs and analyses was removed from the live tree on 2026-08-14. Git history is the archive.

Core wiki module pages affected by the redesign are explicitly marked LEGACY/current-state. Do not restore historical planning artifacts or treat current module existence as target architecture.

## Runtime stack

The repository still contains the current running modular-monolith implementation while the redesign is being specified. Use current code/schema/OpenAPI to answer **what runs today**; use the active redesign authority to answer **what should exist after the reset**.

Stable engineering practices such as contract-first OpenAPI, RFC 9457 errors, multi-tenant isolation/RLS defense-in-depth, transactional outbox, auditable writes and DB-enforced invariants remain in force unless explicitly re-adjudicated.

## Exact next design step

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

No code yet.
