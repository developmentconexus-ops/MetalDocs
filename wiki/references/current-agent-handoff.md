# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN / R10-A CLOSED / R10-B1 CLOSED / R10-B2 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical-architecture authority, including promoted R10-A and R10-B1
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

The old R3–R9.5 ledger remains binding for frozen semantics. Historical stage-routing text inside frozen artifacts is superseded by the current program/stage authorities; do not edit frozen product semantics merely to make historical status lines current.

---

## Current checkpoint

```text
R3–R9   = LOCKED

R9.5-1  = LOCKED
R9.5-2  = LOCKED
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED
R9.5-5  = LOCKED (refined by R9.5-8)
R9.5-6  = LOCKED
R9.5-7  = LOCKED
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN
reopen set = EMPTY

R10-A   = CLOSED / APPROVED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED
R10-B2  = NEXT / DESIGN ONLY
R10-B3  = NOT STARTED
R10-B4  = NOT STARTED
R10-B5  = NOT STARTED
R10-B6  = NOT STARTED
R10-C   = NOT STARTED
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Promoted R10 authority:

`wiki/architecture/r10-technical-architecture.md`

---

## R10-A promoted outcome

The V1 ownership set remains fixed:

### Business bounded contexts — 8

```text
Authentication
Organization
Authorization
Controlled Information
Approval
Documentary Context
Records Governance
Distribution
```

### Supporting semantic owners — 3

```text
Artifact
Audit
Interchange
```

### Attributed support / projections

```text
Notifications → internal/support/notifications
Search        → internal/projections/search
```

R10-A reopen set remains empty. Package naming preference, current-schema convenience, provider capability and hypothetical futures are not reopen evidence.

---

## R10-B1 promoted outcome

R10-B1 — **Relational Substrate, Tenancy & Reference Law** — is now CLOSED / APPROVED.

Key promoted laws:

```text
one PostgreSQL DB                         = YES
canonical target product schema           = metaldocs
schema-per-BC / schema-per-Tenant          = NO

tenant-owned identity                     = PRIMARY KEY (tenant_id, id)
technical id                               = UUID
business/provider/external id as PK        = NO
same-Tenant existence reference            = composite FK
cross-owner FK                             = authority-neutral
cross-owner FK actions                     = RESTRICT / NO ACTION only
cross-owner CASCADE/SET NULL/SET DEFAULT   = FORBIDDEN
universal polymorphic business registry    = FORBIDDEN

business timestamp                         = TIMESTAMPTZ
canonical SHA-256                           = BYTEA + 32-byte constraint
frozen vocabulary default                  = TEXT + CHECK
real unknown/absence                       = NULL

semantic persistence classes:
  SEMANTIC AUTHORITY
  ATTRIBUTED SUPPORT
  DURABLE MECHANISM
  EPHEMERAL MECHANISM
  REBUILDABLE PROJECTION

mutation law is a separate dimension:
  MUTABLE
  IMMUTABLE / APPEND-ONLY
  TERMINAL / TOMBSTONED
  REBUILDABLE
  or explicit constrained state machine

RLS on tenant semantic authority           = fail-closed / ENABLE + FORCE
RLS canonical Authorization                = NO
ordinary serving DB role                   = non-owner / NOSUPERUSER / NOBYPASSRLS
serving system content access              = explicit Tenant context
true bulk maintenance                      = separate non-serving trust surface, later R10-F/ops

default transaction isolation              = READ COMMITTED
cross-owner frozen atomicity               = one local PostgreSQL transaction
mandatory Audit append                     = same commit when frozen-required
mandatory durable async intent             = same commit when required
async execution/retry/DLQ                  = R10-D
```

Only two background-discovery shapes are lawful under fail-closed semantic/support isolation:

1. enumerate Tenants, then query/work tenant-by-tenant;
2. consume a tenant-written platform routing/due intent exposing routing metadata only, then re-enter tenant-scoped execution.

A scheduler/job does not get a third path that fails open tenant semantic/support RLS.

R10-B1 independent closure evidence ended at:

```text
prior findings closed = 6/6
BLOCKER = 0
MAJOR   = 0
LOW     = 2 non-blocking, both dispositioned at promotion
R9.5 reopen = EMPTY
R10-A reopen = EMPTY
```

The two LOWs are closed as follows:

- B4 is explicitly a design work package; Rendition/Release/effectivity remain **Controlled Information-owned**.
- R10-D must explicitly exercise or decline the narrower-representation clause for Notifications persistence while preserving the same tenant-isolation claim.

Review artifacts remain evidence, not authority.

---

## R10-B design-package map

```text
B2 → Authentication + Organization + Authorization relational state
B3 → Artifact relational core + Controlled Information + WorkingContent + Submission
B4 → Approval + Controlled Information-owned Rendition/Release/effectivity relational state + Distribution
B5 → Documentary Context + Records Governance + Artifact second-consumer/no-confirmed-orphan closure
B6 → Audit relational state + Interchange batch/plan/outcome state + cross-owner transaction matrix + imported-history/global DB coherence

R10-D → Notifications attributed-support persistence + Search projection persistence + async mechanism persistence/execution
```

This sequencing does not move R10-A ownership.

---

## Exact next step — R10-B2

Open **R10-B2 — Authentication / Organization / Authorization State** in design-only mode.

Do not begin by copying current IAM/auth/security tables. First perform an independent B2 intake/decomposition under the DevelopmentConexus Engineering Method and B1 substrate laws.

At minimum B2 must decide:

```text
Authentication credential/session identity representation
Authentication ↔ Organization User binding
Tenant / Area / User / Group / GroupMembership persistent state
Tenant settings/configuration persistence without new authority
Tenant lifecycle ACTIVE/SUSPENDED/ERASED durable representation
TenantDeletionRequest / TenantErasureRecord / tombstone state
tenant key-custody lifecycle facts without moving crypto mechanism into Organization
Permission / Role / RoleAssignment representation
User|Group subject representation
Tenant|Area typed scope representation
grant/revocation evidence
canonical grant-evaluation read surface for later owners
same-Tenant FK + RLS application under B1
semantic persistence + mutation-law classification for every B2 family
transaction boundaries for membership/grant/lifecycle mutations
required same-commit Audit/durable-intent insertion points
```

B2 must preserve:

- Authentication ≠ Organization ≠ Authorization;
- no `tenant_owner` bypass;
- flat groups only V1;
- exactly five frozen tenant roles;
- RoleAssignment subject = User|Group and scope = Tenant|Area;
- additive/default-deny grants;
- domain relationship predicate meaning remains outside Authorization;
- PlatformOperator/SystemPrincipal remain outside tenant RBAC with no implicit tenant-content authority;
- no Keycloak, OpenFGA/SpiceDB, nested groups, generic ACL/ReBAC graph, deny engine or speculative enterprise identity machinery without a real trigger.

Current auth/IAM/security tables, schema, code and runtime are evidence only. No schema/code implementation is authorized.

---

## Explicitly deferred from launch

These remain future triggers, not hidden V1 TODOs:

```text
malware scanning / ClamAV / quarantine / periodic rescans
ArtifactSecurityAssessment / CDR / advanced content security
ICP-Brasil / PKI / DocuSign / Adobe Sign / RFC3161 / TSA / HSM
cryptographically signed export packages
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery / ESI preservation
realtime coauthoring / WOPI-style collaboration
true indivisible multi-file ArtifactPackage without a real triggering format
Keycloak/external IdP without enterprise identity trigger
OpenFGA/SpiceDB without arbitrary relationship-sharing requirement
BPMN/Camunda/Flowable/Temporal as Approval prerequisites
```

## Implementation gate

**CLOSED.** No product implementation starts in R10-B2. Product implementation begins only after the integrated R10 technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.
