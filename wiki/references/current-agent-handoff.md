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
- `DocumentProfile` is replaced by tenant-scoped `DocumentType` with immutable code, ACTIVE/INACTIVE lifecycle and no independent versioning;
- a Document's type is immutable after creation in V1;
- `DocumentFamily` becomes optional classification-only `DocumentTypeCategory`; no inherited policies/hierarchy in V1;
- `GovernanceClass {controlado, simples, livre}` is deleted; each authority owns explicit configuration instead;
- Approval configuration explicitly distinguishes `NoHumanApproval` from `UsePolicy(...)`;
- Area belongs to Organization, not taxonomy;
- template is a role of a governed Document, not a parallel lifecycle;
- TemplateUse is M:N between template Documents and DocumentTypes, with at most one default per type; default is UX only;
- blank creation remains allowed in V1; no `template_required` rule without a real requirement;
- creation resolves the template Document's current effective revision and permanently pins source document + exact revision + content hash;
- newer template revisions apply only to future creations; existing documents never rebind;
- changing template placeholder/schema/layout/resolver semantics means a new ordinary DocumentRevision;
- official human/audit revision labels are `REV001`, `REV002`, `REV003`, ... — never user-facing `v7`; technical row/schema/policy versions remain separate namespaces;
- freeze/approval/rendition must always bind the exact revision/hash the human reviewed;
- Release Coordinator/effectivity remains downstream from human approval.

## Important product evidence

A real browser QA run proved the old content model was structurally contradictory: the user edited and the approver reviewed one revision, while freeze rendered the blank template snapshot. The final signed PDF was blank and the signed hash did not represent the reviewed content. The redesign must make that state impossible, not patch one renderer call.

## Whole-product coverage still required

Before code we still have to close, explicitly:

- **NEXT:** Document + DocumentRevision lifecycle, REV allocation and immutable submission evidence;
- numbering/NumberSeries;
- TemplateSpec exact revision payload and source-provenance placement;
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

Continue **R4 — Document + DocumentRevision lifecycle + immutable submission evidence**:

```text
Document lifecycle
+ DocumentRevision lifecycle
+ REVxxx allocation semantics
+ mutable working content vs immutable submission identity
+ whether SubmissionSnapshot is a first-class concept
+ return-for-changes/resubmission
+ effective/superseded/obsolete relationships
+ template-source provenance placement
```

Every transition must have exactly one authority and a clear immutable fact/evidence proving what content the transition applies to. Do not implement.

## Documentation reset

`docs/superpowers` was intentionally collapsed on 2026-08-14 to only the active redesign staging material. Old plans/specs/milestones/reports/analyses remain in Git history and are not forward authority.

Core wiki module pages and old roadmap/backlog surfaces are marked LEGACY/HISTORICAL and redirected to the cohesive redesign authority so a fresh agent cannot accidentally continue the old architecture.
