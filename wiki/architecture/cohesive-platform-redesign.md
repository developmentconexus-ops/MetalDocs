# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED; R10-A CLOSED / SINGLE-COMPANY-REFINED; R10-B1 CLOSED / SINGLE-COMPANY-RESTRUCTURED; R10-B2 IN PROGRESS; R10-B2-1 CLOSED / APPROVED; R10-B2-2 NEXT / NO PRODUCT IMPLEMENTATION AUTHORIZED**
> **Established:** 2026-08-14
> **R9.5 freeze ratified:** 2026-08-17
> **R10-A promotion ratified:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Single-Company Deployment / Tenancy Rebaseline ratified:** 2026-08-17
> **R10-B2-1 promotion ratified:** 2026-08-17
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

## 3. North star / deployment model

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and are bound to MetalDocs organizational identity through a stable provider subject identity. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

V1 deployment invariant:

> **One company per deployment. The same codebase, build artifacts and migrations are reused for every deployment; customer-specific forks are forbidden. Shared/pooled customer tenancy is deferred until measured evidence selects it.**

`Tenant` remains the singleton company root and whole-company Authorization scope target; it is not a database partition.

## 4. Principal decisions — refined mirrors

### Authentication

Keycloak is V1 AuthN provider. MetalDocs Authentication owns provider-subject binding, opaque app Session, Session revocation/lifecycle, assurance/fresh-auth and anti-corruption contract. Stable provider identity = `issuer+subject`; provider roles/groups/organizations/permissions/arbitrary claims never become canonical AuthZ. Keycloak Organizations/company switching are not V1 requirements. No cross-provider DB atomicity.

R10-B2-1 now fixes the concrete V1 Authentication state:

```text
ProviderSubjectBinding
  id UUID
  user_id UUID
  issuer
  subject
  created_at
  disabled_at?
  UNIQUE(issuer,subject)
  UNIQUE(user_id) WHERE disabled_at IS NULL

ApplicationSession
  id UUID
  subject_binding_id UUID
  credential_digest
  created_at
  expires_at
  revoked_at?
  latest_reauthenticated_at?
  latest_provider_auth_time?
  latest_acr?
  latest_amr?
```

Binding mapping fields are immutable; `disabled_at` is reversible acceptance state. At most one enabled binding/User V1. One retained `(issuer,subject)` cannot be handed to another MetalDocs User without a material reopen. Binding creation requires intent-bound/provider-proven or explicit trusted-human correlation; email/username/display name never select identity.

ApplicationSession is a local high-entropy opaque bearer with only a one-way verifier persisted, finite absolute expiry, server-side revocation and no Tenant/AuthZ/provider-token snapshot. Provider-only disable has bounded staleness no longer than the remaining local Session lifetime absent earlier reconciliation; MetalDocs User offboarding revokes local access independently of Keycloak availability.

Fresh-auth Session fields are evidence inputs only. `requires_reauthentication` is never satisfied by bare non-NULL state: a consumer must use one-shot operation evidence or an explicitly bounded freshness rule owned by that consumer's policy authority. Forced reauth must prove the same `(issuer,subject)` and cannot revive a revoked/expired Session.

### Organization + Authorization

V1 Organization = exactly one Tenant root + Area + User + Group + GroupMembership. `Tenant.id` immutable; editable identity/settings mutable; exactly one root structurally per product DB. No universal `tenant_id`/company/deployment partition column by reflex.

Five roles remain: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`. RoleAssignment = User|Group over `TenantScope|AreaScope`; TenantScope means whole company. R9+R9.5 catalog = **27 base + 16 = 43 permissions**; `tenant.export` and `tenant.deletion.request` are removed while their owning features are deferred. No Tenant/Area/role/Permission RLS as canonical AuthZ.

### Controlled Information / Approval / Context / Records / Distribution

All frozen Document/Revision/WorkingContent/Submission/Template/Rendition/Release, specialized Approval/SoD, Dossier/Evidence, Records Governance and Distribution semantics remain. Former tenant-qualified uniqueness is re-derived to actual deployment/semantic scope.

### Storage / Artifact

Artifact exact-byte identity/hash remains provider-independent. ManagedArtifactStore port+conformance is first-class; Local dev/test and AWS S3 reference production. Tenant/company key prefix is not an isolation invariant; keys remain opaque/immutable/no-overwrite. Production malware inspection before confirming untrusted bytes remains mandatory/fail-closed.

### Retention / privacy

Tenant customer lifecycle/deletion is not V1 product state; deployment decommission is operations. User/data-subject privacy remains: offboarding/session revocation, erasable human-readable enrichment, PII-minimized/non-PII immutable Audit skeleton and restore non-resurrection proof. Retention/LegalHold/Disposition remain binding. No generic privacy workflow implied.

Authentication Binding/Session rows are not governed-retention subjects. Lawful erasure may remove them after dependent Sessions are handled; erasing a Binding row also surrenders its structural no-recorrelation guarantee for that provider subject, so any later correlation is a new trusted decision.

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

Refinement effects: Organization owns singleton Tenant/settings + Area/User/Group/GroupMembership + User lifecycle/offboarding, no customer lifecycle/tombstone family; Audit owns surviving privacy-safe evidence skeleton; Interchange has no Tenant Portability process V1; customer routing is not a jobs/platform requirement.

## 7. Deployment/build-vs-buy posture

One company per deployment, same product artifacts, no customer forks. Keycloak V1, Organizations/company switching not required. ManagedArtifactStore port+conformance. OpenFGA/SpiceDB, generic BPM, PKI/eDiscovery/realtime coauthoring remain trigger-based.

Shared/pooled tenancy re-enters only on measured evidence: unsustainable stamp economics, operations-capacity failure despite automation, genuine cross-company product capability, self-service provisioning becoming a proven blocker, or contractual/compliance requirement. A second customer alone triggers an economics review, not automatic pooling.

## 8. Implementation gate

**Closed.** R10-B2-2 through R10-F must close before implementation specification/plan/code.

## 9. Exact next step — R10-B2-2 Organization

R10-B2 is **IN PROGRESS**. R10-B2-1 is **CLOSED / APPROVED**. Start **R10-B2-2 — Organization singleton root / people / groups** from the promoted B2-1 Authentication contract and B1 single-company substrate.

At minimum B2-2 must decide:

```text
Tenant singleton root
  exact persistent representation
  structural exactly-one enforcement
  immutable Tenant.id
  editable company identity/settings
  expected_tenant_id startup/readiness consumer surface

Area
  identity/stable fields
  real lifecycle if any
  deployment-wide uniqueness

User
  technical identity
  authentication-eligibility/offboarding semantics
  erasable profile/enrichment boundary
  Area relationship only if semantically required
  no provider-state mirror

Group
  flat-group identity
  deployment-wide uniqueness
  real lifecycle if any

GroupMembership
  User↔Group relationship
  mutation/evidence semantics
  no nested groups

B2-1 integration
  ProviderSubjectBinding → User integrity
  Session issuance consumes current User eligibility + accepted Binding
  offboarding revokes ApplicationSessions race-safely
```

B2-2 must not design Authorization grants/RoleAssignment (B2-3), final cross-owner transaction/locking matrix (B2-4), Keycloak retry/provisioning mechanics (R10-D), frontend/admin journeys (R10-E), or resurrect Tenant customer lifecycle/partitioning/RLS.

Keep successor notes visible for B2-4/B6: binding re-enable participates in the same acceptance/Session-issuance serialization discipline; lawful Binding erasure surrenders the retained-row no-recorrelation guarantee; same-commit Audit/durable-intent points remain to close in B2-4.
