# Current Agent Handoff

> **Last verified:** 2026-08-14
> **Status:** ACTIVE — Cohesive Platform Redesign
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Read this first

A fresh session MUST read:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. this file

Do not resume any old roadmap, milestone, spec, plan, deleted `docs/superpowers` artifact, or PR #113 implementation by inertia.

## Where we are

MetalDocs is being redesigned as a cohesive platform before the next implementation wave. The redesign expanded from AuthZ → Approval → Controlled Information after proving that the current module split contains overlapping authorities and contradictory content/lifecycle models.

Locked so far:

- current AuthN can remain for V1; Keycloak/external IdP is deferred behind a future enterprise-identity seam;
- Organization owns Tenant/Area/User/Group; Groups are flat and receive ordinary RoleAssignments;
- built-in roles are `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`;
- scoped RBAC + Groups is enough for V1; no OpenFGA/SpiceDB now;
- Approval V1 is a versioned sequential governed-document approval engine, not BPM;
- Approval steps use named user / group / role-in-area participant rules and ANY/ALL completion;
- human outcomes are `accept` / `return_for_changes`; edited content creates a new approval attempt;
- `documents` + `controlleddocuments` + `templates` do not survive as three target bounded contexts;
- target Controlled Information core is `Document` + `DocumentRevision`;
- template is a designation/role of an exact governed revision, not a parallel lifecycle; changing placeholder/schema/layout/resolver semantics means a new DocumentRevision;
- derived documents remain bound to the exact template revision/hash used to seed them;
- freeze/approval/rendition must always bind the exact revision/hash the human reviewed;
- `DocumentProfile` is converging toward `DocumentType`;
- Area moves out of taxonomy into Organization;
- Release Coordinator/effectivity remains downstream from human approval.

## Important product evidence

A real browser QA run proved the old content model was structurally contradictory: the user edited and the approver reviewed one revision, while freeze rendered the blank template snapshot. The final signed PDF was blank and the signed hash did not represent the reviewed content. The redesign must make that state impossible, not patch one renderer call.

## Whole-product coverage still required

Before code we still have to close, explicitly:

- DocumentType / Family / GovernanceClass / TemplateDesignation;
- Document + Revision lifecycle and submission snapshots;
- numbering/NumberSeries;
- periodic review/reason-for-change;
- renditions/rendering/reconstruction evidence;
- release/effectivity/supersession;
- distribution/read/acknowledgement;
- token/computed-value snapshot semantics;
- audit/evidence boundary;
- notifications and search projections;
- tenant lifecycle/security/external IdP trigger;
- final Permission catalog + role bundles;
- bounded contexts, table/transaction ownership, events, data model, APIs, frontend journeys;
- delete/migrate/rename map from current implementation;
- final ADR/spec set and implementation plan.

## Exact next step

Continue the design discussion with:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

For each, determine whether it has independent business meaning or is only historical encoding of another authority. Do not implement.

## Documentation reset

`docs/superpowers` was intentionally collapsed on 2026-08-14 to only the active redesign staging material. Old plans/specs/milestones/reports/analyses remain in Git history and are not forward authority.

Core wiki module pages and old roadmap/backlog surfaces are being marked LEGACY/HISTORICAL and redirected to the cohesive redesign authority so a fresh agent cannot accidentally continue the old architecture.
