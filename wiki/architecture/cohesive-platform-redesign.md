# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED; R10-A CLOSED / SINGLE-COMPANY-REFINED; R10-B1 CLOSED / SINGLE-COMPANY-RESTRUCTURED; R10-B2 CLOSED / APPROVED / INTEGRATED; R10-B3 NEXT / DESIGN ONLY / NO PRODUCT IMPLEMENTATION AUTHORIZED**
> **Established:** 2026-08-14
> **R9.5 freeze ratified:** 2026-08-17
> **R10-A promotion ratified:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Single-Company Deployment / Tenancy Rebaseline ratified:** 2026-08-17
> **R10-B2-1 promotion ratified:** 2026-08-17
> **R10-B2 integrated promotion ratified:** 2026-08-17
> **Repository baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** [`../engineering/standards/root-cause-global-maximum-method.md`](../../docs/engineering/standards/root-cause-global-maximum-method.md)
> **Frozen R3–R9.5 product/domain ledger:** [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md)
> **Active R10 technical authority:** [`r10-technical-architecture.md`](r10-technical-architecture.md)

## 1. Purpose

MetalDocs is being redesigned as one coherent product before the next large implementation wave. Current implementation is evidence, never automatic target entitlement.

The target is the smallest professional architecture that models controlled information correctly, gives each fact one authority, preserves auditability/least privilege, serves one company per V1 deployment, uses commodity mechanisms without surrendering business authority, avoids speculative platforms/pooled tenancy, and is specified end-to-end before implementation.

## 2. Authority / evidence

Fresh sessions follow `AGENTS.md` → Method → current handoff → this page → frozen ledger → active R10 authority. Review artifacts are evidence only.

Single-company evidence chain:

```text
candidate                cba89d9d
independent review       1acd5128  APPROVE WITH MATERIAL FIXES
corrected target          31a57e5b
bounded delta review      c87751f3  APPROVE
BLOCKER                   0
MAJOR                     0
prior findings closed     9/9
new material contradiction NONE
```

R10-B2-1 evidence chain:

```text
candidate                9cba3acd
independent review       361f6c8b  APPROVE WITH MATERIAL FIXES
corrected target          ee0a0ce0
bounded delta review      6593c471  APPROVE
BLOCKER                   0
MAJOR                     0
prior findings closed     8/8
new material contradiction NONE
new concurrency counterexample NONE
```

Integrated R10-B2 evidence chain:

```text
candidate                  b814f672
independent full review    34a567fd  APPROVE WITH MATERIAL FIXES
corrected target           2908a884
bounded delta review       507075a8  APPROVE
BLOCKER                    0
MAJOR                      0
final LOW                  2 non-blocking notes
exact 5×43 bundle matrix   MATCH
A1 access administration   APPROVE tenant-owner-only
reviewed deadlock cycle    NONE
new material contradiction NONE
B2-1 reopen                NO
reopen outside B2          NO
broad review required      NO
```

## 3. North star / deployment model

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and are bound to MetalDocs organizational identity through a stable provider subject identity. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

V1 deployment invariant:

> **One company per deployment. The same codebase, build artifacts and migrations are reused for every deployment; customer-specific forks are forbidden. Shared/pooled customer tenancy is deferred until measured evidence selects it.**

`Tenant` remains the singleton company root and whole-company Authorization scope target; it is not a database partition.

## 4. Principal decisions — refined mirrors

### Authentication

Keycloak is V1 AuthN provider. MetalDocs Authentication owns provider-subject binding, opaque app Session, Session revocation/lifecycle, assurance/fresh-auth and anti-corruption contract. Stable provider identity = `issuer+subject`; provider roles/groups/organizations/permissions/arbitrary claims never become canonical AuthZ. Keycloak Organizations/company switching are not V1 requirements. No cross-provider DB atomicity.

Concrete V1 Authentication state remains exactly `ProviderSubjectBinding` + `ApplicationSession`. Binding uses immutable `(issuer,subject)→User` mapping fields plus reversible acceptance; total `(issuer,subject)` uniqueness and one-enabled-binding/User are structural. Session is opaque/local/digest-only/finite/terminally revocable and contains no AuthZ/provider-token snapshot. Fresh-auth fields are evidence only; consumer policy must use one-shot or explicitly bounded freshness. Provider effects are post-commit R10-D choreography.

### Organization + Authorization — integrated R10-B2 promoted

Organization persistent state is exactly:

```text
Tenant(id, display_name)
Area(id, immutable unique code, name, disabled_at?)
User(id, disabled_at?)
UserProfile(user_id, display_name, email?)
Group(id, unique name)
GroupMembership(user_id, group_id) current pair
```

Tenant at-most-one is a constant-expression unique index; startup/readiness proves at-least-one and matching `expected_tenant_id`. Area retirement is reversible, preserves existing references/grants and blocks new references. User identity is minimal; PII enrichment is separately erasable. Groups are flat/company-wide and hard-delete only when no live typed reference exists.

Authorization uses static product Role/Permission catalogs and one persisted `RoleAssignment` family. Five roles remain: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`. Exact current role-bundle counts are **3 / 15 / 4 / 25 / 43** respectively; `tenant_owner` owns all 43. `access.manage` exists only in `tenant_owner`, so V1 RoleAssignment/GroupMembership administration is tenant-owner/TenantScope-only. `area_manager` is operational, never RBAC admin.

RoleAssignment is typed `User|Group × Role × TenantScope|AreaScope`, with subject XOR, scope XOR, role vocabulary CHECK, role↔scope DB CHECK and four partial uniqueness backstops. `tenant_owner→TenantScope`; `area_manager→AreaScope`; author/approver/viewer may use either. Current row = current grant; delete = revoke; Audit owns transition evidence. No roles/permissions tables, effective-permission store, Session AuthZ snapshot, provider role bridge, RLS policy engine, custom role platform, deny engine or ReBAC graph.

Canonical AuthZ remains live additive/default-deny:

```text
current direct/group grants
→ static Role→Permission bundle
→ scope
→ domain relationship predicate
→ domain governance constraints
→ ALLOW / DENY
```

User offboarding atomically disables User, revokes local Sessions, deletes current memberships/direct grants, writes required Audit and durable provider intent. Binding correlation remains. Re-enable restores identity eligibility only; prior access/Sessions never resurrect automatically.

B2 concurrency under READ COMMITTED uses deterministic lock classes `User → Binding(s) → Area → ordered child sets`, with lifecycle mutators `FOR UPDATE` and eligibility/acceptance readers `FOR SHARE`; Group deletion is an isolated Group→memberships→group-grants order. The independent delta found no wait-for cycle under this law.

### Controlled Information / Approval / Context / Records / Distribution

All frozen Document/Revision/WorkingContent/Submission/Template/Rendition/Release, specialized Approval/SoD, Dossier/Evidence, Records Governance and Distribution semantics remain. Former tenant-qualified uniqueness is re-derived to actual deployment/semantic scope.

### Storage / Artifact

Artifact exact-byte identity/hash remains provider-independent. ManagedArtifactStore port+conformance is first-class; Local dev/test and AWS S3 reference production. Tenant/company key prefix is not an isolation invariant; keys remain opaque/immutable/no-overwrite. Production malware inspection before confirming untrusted bytes remains mandatory/fail-closed.

### Retention / privacy

Tenant customer lifecycle/deletion is not V1 product state; deployment decommission is operations. User/data-subject privacy remains: offboarding/session revocation, erasable human-readable enrichment, PII-minimized/non-PII immutable Audit skeleton and restore non-resurrection proof. Retention/LegalHold/Disposition remain binding. No generic privacy workflow implied.

Authentication Binding/Session rows are not governed-retention subjects. Lawful erasure may remove them after dependents are handled; erasing a Binding row also surrenders its structural no-recorrelation guarantee, so later correlation is a new trusted decision.

### Migration / export

V1 keeps Backup/Restore, Historical Migration, Governed Subject Export, IMPORT_COPY/PUBLISH_COPY and authorization-safe completeness. Tenant Portability Export is deferred; equivalent stamp movement uses Backup/Restore absent a real product-exit contract.

## 5. R9.5 status

```text
R9.5-1 LOCKED / SINGLE-COMPANY-REFINED where scope wording changed
R9.5-2 LOCKED / GCR + SINGLE-COMPANY-REFINED
R9.5-3 LOCKED (R9.5-8 refinement)
R9.5-4 LOCKED / SINGLE-COMPANY-REFINED where uniqueness wording changed
R9.5-5 LOCKED (R9.5-8 + privacy re-anchor)
R9.5-6 LOCKED / SINGLE-COMPANY-REFINED
R9.5-7 LOCKED / GCR-REFINED
R9.5-8 CLOSED / APPROVED
R9.5   FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED
reopen set EMPTY
```

## 6. R10-A ownership after refinement

Business BCs remain exactly 8: Authentication, Organization, Authorization, Controlled Information, Approval, Documentary Context, Records Governance, Distribution.

Supporting semantic owners remain exactly 3: Artifact, Audit, Interchange. Notifications = attributed support; Search = rebuildable projection.

Refinement effects: Organization owns singleton Tenant/settings + Area/User/UserProfile/Group/GroupMembership + User lifecycle/offboarding; Authorization owns static role/permission semantics, current RoleAssignment state and canonical evaluation; Audit owns surviving privacy-safe evidence skeleton; Interchange has no Tenant Portability process V1; customer routing is not a jobs/platform requirement.

## 7. Deployment/build-vs-buy posture

One company per deployment, same product artifacts, no customer forks. Keycloak V1, Organizations/company switching not required. ManagedArtifactStore port+conformance. OpenFGA/SpiceDB, generic BPM, PKI/eDiscovery/realtime coauthoring remain trigger-based.

Shared/pooled tenancy re-enters only on measured evidence: unsustainable stamp economics, operations-capacity failure despite automation, genuine cross-company product capability, self-service provisioning becoming a proven blocker, or contractual/compliance requirement. A second customer alone triggers an economics review, not automatic pooling.

## 8. Implementation gate

**Closed.** R10-B3 through R10-F must close before implementation specification/plan/code.

## 9. Exact next step — R10-B3 Controlled Information + Artifact relational core

R10-B2 is **CLOSED / APPROVED / INTEGRATED**. Start **R10-B3 — Controlled Information + Artifact relational core** from the promoted B1 substrate and B2 identity/access laws.

B3 uses batch mode: first perform one integrated intake/decomposition and candidate relational system, covering Artifact core, DocumentType/configuration, Document/Revision, WorkingContent/OCC, immutable RevisionSubmission, template role/use/spec/provenance, numbering, Editorial/Periodic Review state and the same-commit constraints that bind them. Separate explicitly what belongs to B4/B5/B6/R10-C/D/E/F.

Do not reopen B2 for implementation convenience. Current documents/controlleddocuments/templates/render schema/code are current-state evidence only. Product implementation remains **BLOCKED**.