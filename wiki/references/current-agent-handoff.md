# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + GCR-REFINED + SINGLE-COMPANY-REFINED / R10-A CLOSED + SINGLE-COMPANY-REFINED / R10-B1 CLOSED + SINGLE-COMPANY-RESTRUCTURED / R10-B2 CLOSED + APPROVED + INTEGRATED / R10-B3 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority / global coherence
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority including promoted GCR + single-company refinements
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical authority including promoted B1 and integrated B2
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

---

## Current checkpoint

```text
R3–R9   = LOCKED

R9.5-1  = LOCKED / SINGLE-COMPANY-REFINED where tenancy wording changed
R9.5-2  = LOCKED / GCR-REFINED / SINGLE-COMPANY-REFINED
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED / SINGLE-COMPANY-REFINED where uniqueness wording changed
R9.5-5  = LOCKED (refined by R9.5-8; single-company privacy re-anchor applied)
R9.5-6  = LOCKED / SINGLE-COMPANY-REFINED (Tenant Portability deferred)
R9.5-7  = LOCKED / GCR-REFINED
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED
reopen set = EMPTY

GCR                         = CLOSED / APPROVED
Single-Company Rebaseline   = CLOSED / APPROVED

R10-A    = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-B    = IN PROGRESS / DESIGN ONLY
R10-B1   = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2   = CLOSED / APPROVED / INTEGRATED
R10-B2-1 = CLOSED / APPROVED
R10-B2-2 = CLOSED / APPROVED / INTEGRATED
R10-B2-3 = CLOSED / APPROVED / INTEGRATED
R10-B2-4 = CLOSED / APPROVED / INTEGRATED
R10-B3   = NEXT / DESIGN ONLY
R10-B4   = NOT STARTED
R10-B5   = NOT STARTED
R10-B6   = NOT STARTED
R10-C    = NOT STARTED
R10-D    = NOT STARTED
R10-E    = NOT STARTED
R10-F    = NOT STARTED

implementation = BLOCKED
```

Promoted target authority:

- R3–R9.5 product/domain semantics: `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
- R10 technical architecture including integrated B2: `wiki/architecture/r10-technical-architecture.md`
- program/global-coherence mirror: `wiki/architecture/cohesive-platform-redesign.md`

---

## Single-Company Deployment / Tenancy Rebaseline — promoted outcome

V1 deployment invariant:

> **One MetalDocs deployment serves exactly one company. The same product codebase, build artifacts and migrations are reused for every deployment; customer-specific forks are forbidden. Shared/pooled multi-customer tenancy is deferred until measured evidence selects it.**

`Tenant` is the single company/organization root of a deployment, whole-company Authorization scope target and deployment↔DB identity anchor. It is not a V1 partition dimension. Exactly one Tenant row exists per product DB; `Tenant.id` is immutable. Deployment config pins `expected_tenant_id`; missing/multiple/mismatch fails closed.

B1 target substrate:

```text
one PostgreSQL product DB / schema metaldocs
id UUID PRIMARY KEY
ordinary typed FKs
cross-owner RESTRICT / NO ACTION only
no universal tenant_id/company_id/deployment_id partition column
no composite tenant PK/FK
no Tenant/Area/role/Permission RLS
no Tenant GUC/context/customer routing
serving DB role non-owner + NOSUPERUSER
READ COMMITTED
required Audit + durable intent in same local commit
provider DB atomicity dependency forbidden
```

Tenant customer lifecycle/deletion/tombstones and Tenant Portability Export are deferred. User/data-subject privacy, Retention, LegalHold, Disposition, Backup/Restore remain.

---

## R10-B2 Integrated Authentication / Organization / Authorization — promoted outcome

B2 was closed as one batch around already-promoted B2-1 rather than through separate B2-2/B2-3/B2-4 microgates.

Evidence chain:

```text
candidate                  b814f672
independent full review    34a567fd  APPROVE WITH MATERIAL FIXES
corrected target           2908a884
bounded delta review       507075a8  APPROVE
BLOCKER                    0
MAJOR                      0
final LOW                  2 non-blocking notes
prior material findings    CLOSED
exact 5×43 bundles         MATCH
A1 access administration   APPROVE tenant-owner-only
deadlock under lock law    NONE
new material contradiction NONE
B2-1 reopen                NO
reopen outside B2          NO
broad review required      NO
```

### Authentication

Exactly two semantic persistent families remain:

```text
ProviderSubjectBinding
ApplicationSession
```

Stable provider identity is `(issuer,subject)`. Binding mapping fields are immutable; acceptance is reversible `disabled_at`; total provider-subject uniqueness and one-enabled-binding/User are structural. Subject selection must be causally tied to exact provisioning/correlation intent or explicit trusted-human correlation; email/username/display name never select identity.

ApplicationSession is opaque/local/digest-only, finite, terminally revocable and carries no Tenant/AuthZ/provider-token snapshot. Fresh-auth Session fields are evidence only; consumer policy chooses one-shot or explicitly bounded freshness. Provider effects never join MetalDocs DB atomicity.

### Organization

```text
Tenant(id, display_name)
Area(id, code UNIQUE immutable, name, disabled_at?)
User(id, disabled_at?)
UserProfile(user_id PK/FK, display_name, email?)
Group(id, name UNIQUE)
GroupMembership(user_id,group_id PK)
```

Tenant structural at-most-one = constant-expression unique index; readiness supplies at-least-one and `expected_tenant_id` check. Area retirement is reversible, preserves existing refs/grants and blocks new use. User is minimal stable identity; human-readable PII lives in erasable UserProfile. No User.home_area, provider mirror or username/email identity. Groups are flat/company-wide; hard deletion fails while any live typed reference exists. Membership is current truth only; Audit owns transitions.

### Authorization

Permissions/Roles are static product catalogs, not DB-editable configuration. Exact current bundles are pinned in the R10 authority:

```text
viewer       3
author      15
approver     4
area_manager 25
tenant_owner 43
```

Only `tenant_owner` carries `access.manage`; V1 GroupMembership/RoleAssignment administration is therefore TenantScope tenant-owner-only. `area_manager` is operational, not RBAC administration.

One persisted family:

```text
RoleAssignment
  id UUID PK
  user_id XOR group_id
  role_code
  tenant_scope_id XOR area_scope_id
```

DB CHECKs enforce role vocabulary and legal role↔scope matrix:

```text
tenant_owner → TenantScope
area_manager → AreaScope
author/approver/viewer → TenantScope | AreaScope
```

Four partial unique indexes reject duplicate current grants. Current row = current grant; delete = revoke; Audit owns transition evidence. No effective-permission tables, Session AuthZ snapshot, provider mapping, RLS policy engine, custom-role platform, deny engine or generic ReBAC graph.

Canonical evaluation = live direct/group grants → static bundle → scope → domain relationship → governance → default-deny result.

### Offboarding / privacy / concurrency

Offboarding atomically disables User, revokes local Sessions, deletes memberships/direct grants, writes required Audit and durable provider-disable intent. Binding correlation remains. Re-enable restores only eligibility; old access/Sessions never silently return.

B2 lock law under READ COMMITTED:

```text
User
→ Binding(s) ordered
→ Area
→ ordered Session / Membership / RoleAssignment child sets
```

Lifecycle mutation uses `FOR UPDATE`; eligibility/acceptance reader uses `FOR SHARE`. Group deletion is isolated Group→memberships→group grants and does not later acquire User/Binding/Area. Delta review found no wait-for cycle.

Offboarding is not privacy erasure. UserProfile/Auth state can be erased when lawful while governed history retains stable User UUID and a PII-minimized/non-PII Audit skeleton. Binding erasure surrenders structural no-recorrelation for the erased subject.

### Successor obligations

- B3–B5/R10-E permission checks declare Tenant-wide vs Area-targeted.
- B4 owns typed live Group references for ApprovalPolicy/Distribution, retired-Area policy rejection, concrete-User snapshots and bounded fresh-auth policy.
- B5 applies Group RESTRICT only if it proves a real Group reference.
- B6 finalizes Audit privacy/forensic fields and cross-owner same-commit matrix.
- R10-C proves restore non-resurrection.
- R10-D executes provider durable intents.
- R10-E chooses Session TTL/journeys/scope-use and optional last-admin UX.
- R10-F specifies bootstrap/lockout recovery, legacy IAM cutover and static-catalog↔CHECK parity gate.

Implementation remains blocked.

---

## Exact next step — R10-B3 Controlled Information + Artifact relational core

Open **R10-B3** in batch/design-only mode. Do not reopen B2 unless material evidence satisfies its reopen contract.

First deliverable = one integrated intake/decomposition and candidate relational system covering:

```text
Artifact core
DocumentType / optional category / numbering
Document + DocumentRevision
WorkingContent + working_version OCC
immutable RevisionSubmission + digest/source Artifact relation
template Document role + TemplateUse + TemplateSpec + origin provenance
EditorialComment where material
PeriodicReview policy/record + responsible owner
creation / draft mutation / submission-freeze transaction boundaries
uniqueness / immutable identity / exactly-one-open/effective constraints
```

Separate clearly:

```text
Approval/Rendition/Release/Distribution → B4
Evidence/Dossier/Records + Artifact closure → B5
Audit/Interchange/cross-owner matrix → B6
physical storage/malware/restore → R10-C
async/projections/provider effects → R10-D
API/frontend → R10-E
historical cutover/deletion → R10-F
```

Use one serious adversarial review of the B3 batch after the integrated candidate; do not return to microdecision review ceremony without a truly independent failure class.

Current implementation is evidence only. Implementation remains **BLOCKED**.