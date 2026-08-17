# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + GCR-REFINED / R10-A CLOSED + GCR-REFINED / R10-B1 CLOSED / R10-B2 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority / global coherence
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority, refined only by the promoted 2026-08-17 GCR deltas recorded in-place
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical authority, including promoted R10-A, R10-B1 and the promoted GCR amendments
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

---

## Current checkpoint

```text
R3–R9   = LOCKED

R9.5-1  = LOCKED
R9.5-2  = LOCKED (GCR-refined provider entitlement / no mandatory DEK)
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED
R9.5-5  = LOCKED (refined by R9.5-8)
R9.5-6  = LOCKED
R9.5-7  = LOCKED (GCR-refined production malware-inspection gate)
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN / GCR-REFINED
reopen set = EMPTY

GCR     = CLOSED / APPROVED

R10-A   = CLOSED / APPROVED / GCR-REFINED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED / GCR-CLARIFIED
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

Promoted target authority:

- R3–R9.5 product/domain semantics: `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
- R10 technical architecture: `wiki/architecture/r10-technical-architecture.md`
- current program/global-coherence mirrors: `wiki/architecture/cohesive-platform-redesign.md`

---

## Global Coherence Review — promoted outcome

The 2026-08-17 GCR closed after:

1. minimal-reopen candidate;
2. independent cold whole-platform review;
3. operator adjudication / corrected target;
4. independent bounded delta review.

Final delta result:

```text
VERDICT = APPROVE GCR ADJUDICATED CORRECTED TARGET
BLOCKER = 0
MAJOR   = 0
prior findings closed = 11/11
new material contradiction = NONE
fifth material local maximum = NONE
```

The GCR changed only the bounded decisions below; all other frozen/promoted decisions remain in force.

### GCR-R1 — Authentication provider

```text
Keycloak = V1 Authentication provider
```

Keycloak/provider owns credential mechanisms: credential storage, password policy, provider account activation/lockout, password recovery, MFA/passkeys, upstream OIDC/SAML/LDAP/AD federation and provider authentication journeys/session.

MetalDocs Authentication owns only product-facing authentication semantics:

```text
provider subject binding
opaque MetalDocs application Session
application-session lifecycle/revocation
authentication-assurance / fresh-auth facts
provider anti-corruption contract
```

The provider boundary is structural. MetalDocs Authorization has no canonical representation for provider roles, groups, organizations, permissions or arbitrary claim-to-permission mappings. Authentication may expose only stable subject identity and enumerated assurance facts such as `issuer`, `subject`, `authenticated_at`, `auth_time`, `acr?`, `amr?`.

Keycloak Organizations, groups and roles are never MetalDocs Organization/Authorization authority. A provider Organization may later act only as AuthN routing/federation projection of a MetalDocs Tenant.

Candidate provider topology remains one Keycloak realm per environment/application trust domain, not realm-per-Tenant; B2 may refine only on material provider/tenant-policy evidence.

### GCR-R2 — Managed Artifact Store

The first-class storage surface is:

```text
ManagedArtifactStore port
+ provider conformance contract
```

Profiles:

```text
Local  = first-class dev/test
AWS S3 = reference production profile
other provider = selected only for a real deployment requirement and must pass conformance
```

MinIO OSS has no product entitlement. A frozen MinIO image or other compatible endpoint may exist temporarily only as dev/CI mechanism until R10-C lands a deliberate conformance environment; that dependency must have a deletion/replacement condition.

### GCR-R3 — production malware inspection

Production invariant:

> Untrusted bytes cannot become `CONFIRMED Artifact` without a successful malware-inspection result.

```text
untrusted bytes
→ STAGED
→ bounded size / ContentFormat / structural-integrity validation
→ malware inspection
→ semantic-owner confirmation validation
→ CONFIRMED
```

Scanner unavailable/incomplete/malicious result means no confirmation. Explicit dev/test profiles may disable inspection; production cannot silently opt out. R10-C must prove the deployment-profile declaration is single-sourced so an inspection-disabled deployment cannot present itself as production.

No V1 quarantine aggregate, CDR platform, periodic-rescan platform, malware-intelligence platform, custom sandbox cluster or `ArtifactSecurityAssessment` domain is introduced.

### GCR-R4 — Tenant DEK removed from V1

Mandatory V1 concepts removed:

```text
Tenant DEK lifecycle
Organization tenant key-custody fact family
mandatory crypto-shred step
mandatory platform KEK / wrap-unwrap machinery
mandatory envelope encryption solely for tenant erasure
```

Tenant erasure remains retention-aware and preserves only an allowed PII-minimized/non-PII audit/platform skeleton. B6 must classify the surviving immutable Audit skeleton field-by-field. Human-readable/user enrichment must be separately erasable or projection/read enrichment.

If B6 proves a real immutable Target Data family must remain stored while becoming unintelligible after lawful erasure, the DEK/key-custody decision reopens with R10-C/R10-F backup/restore proof. No cryptographic-erasure claim may exist without a named Target Data family and fail-closed enforcement.

### GCR clarifications

North Star:

> MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and are bound to MetalDocs organizational identity through a stable provider subject identity.

Database topology:

```text
one MetalDocs product-state PostgreSQL database
canonical target product-state schema = metaldocs
```

Provider-owned products retain separate persistence authority. **No MetalDocs invariant may depend on atomicity across the MetalDocs product-state database and a provider-owned database.** Physical co-location on one PostgreSQL server/cluster does not merge authority.

---

## R10-A promoted outcome after GCR refinement

The V1 ownership set remains exactly:

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

GCR did not alter this topology. It only refined Authentication facts/provider placement and removed the unsupported mandatory Organization key-custody fact family.

---

## R10-B1 promoted outcome after GCR clarification

R10-B1 remains CLOSED / APPROVED. Structural laws are unchanged.

Key substrate laws:

```text
one MetalDocs product-state PostgreSQL DB    = YES
canonical target product schema              = metaldocs
schema-per-BC / schema-per-Tenant             = NO
provider-owned DB atomicity dependency         = FORBIDDEN

tenant-owned identity                         = PRIMARY KEY (tenant_id, id)
technical id                                   = UUID
business/provider/external id as PK            = NO
same-Tenant existence reference                = composite FK
cross-owner FK                                 = authority-neutral
cross-owner FK actions                         = RESTRICT / NO ACTION only
cross-owner CASCADE/SET NULL/SET DEFAULT       = FORBIDDEN
universal polymorphic business registry        = FORBIDDEN

business timestamp                             = TIMESTAMPTZ
canonical SHA-256                              = BYTEA + 32-byte constraint
frozen vocabulary default                      = TEXT + CHECK
real unknown/absence                           = NULL

semantic persistence classes:
  SEMANTIC AUTHORITY
  ATTRIBUTED SUPPORT
  DURABLE MECHANISM
  EPHEMERAL MECHANISM
  REBUILDABLE PROJECTION

mutation law is separate:
  MUTABLE
  IMMUTABLE / APPEND-ONLY
  TERMINAL / TOMBSTONED
  REBUILDABLE
  or explicit constrained state machine

RLS on tenant semantic authority               = fail-closed / ENABLE + FORCE
RLS canonical Authorization                    = NO
ordinary serving DB role                       = non-owner / NOSUPERUSER / NOBYPASSRLS
serving system content access                  = explicit Tenant context
true bulk maintenance                          = separate non-serving trust surface

default transaction isolation                  = READ COMMITTED
cross-owner frozen atomicity                   = one local MetalDocs PostgreSQL transaction
mandatory Audit append                         = same commit when frozen-required
mandatory durable async intent                 = same commit when required
async execution/retry/DLQ                      = R10-D
```

No XA/2PC/distributed transaction with Keycloak or any provider DB.

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

Do not copy current auth/IAM/security tables. Derive target persistent state from the promoted semantics and B1 substrate.

At minimum B2 must decide, line-by-line:

```text
Authentication provider-subject binding representation
  - stable issuer + subject boundary
  - explicit tenant dimension / uniqueness law
  - one User ↔ one/many provider-subject cardinality decision
Authentication ↔ Organization User integrity without collapsing AuthN into Organization
MetalDocs opaque application Session representation/lifecycle
fresh-auth / authentication-assurance representation
provider anti-corruption contract with structural no-provider-role/group/org/permission consumption proof
provider-binding/provisioning lifecycle + idempotent reconciliation choreography
  - User exists / provider subject absent
  - provider subject exists / binding absent
  - binding exists / provider subject removed or disabled
  - duplicate subject binding attempt
  - provider unavailable
  - retry after uncertain provider response
provider-side disable vs already-live MetalDocs Session posture

Tenant / Area / User / Group / GroupMembership persistent state
Tenant settings/configuration persistence without inventing a new authority
Tenant lifecycle ACTIVE/SUSPENDED/ERASED durable representation
TenantDeletionRequest / TenantErasureRecord / tombstone + restore-reconciliation state

Permission / Role / RoleAssignment representation
User|Group subject representation
Tenant|Area typed scope representation
grant/revocation evidence
canonical grant-evaluation read surface for later owners

same-Tenant FK + RLS application under B1
semantic persistence class + mutation-law classification for every B2 family
transaction boundaries for membership/grant/lifecycle mutations
required same-commit Audit/durable-intent insertion points
```

B2 explicitly does **not** design:

```text
password hash / password-policy tables
credential activation/lockout/MFA/passkey persistence
provider roles/groups/orgs as product authority
claim-to-permission mappings
Tenant DEK / key-custody / KEK / wrap-unwrap infrastructure
XA / 2PC across provider and product databases
```

B2 must preserve:

- Authentication ≠ Organization ≠ Authorization;
- exactly five frozen tenant roles;
- flat Groups only V1;
- RoleAssignment subject = User|Group and scope = Tenant|Area;
- additive/default-deny grants; no `tenant_owner` bypass;
- domain relationship predicate meaning remains outside Authorization;
- RLS remains Tenant isolation only;
- PlatformOperator/SystemPrincipal remain outside tenant RBAC with no implicit tenant-content authority;
- Keycloak is V1 AuthN provider but never canonical Organization/AuthZ authority;
- no OpenFGA/SpiceDB, generic ACL/ReBAC graph, deny engine or nested groups without a real trigger.

### Successor proof obligations already routed

- **B6:** prove the immutable Audit skeleton can remain PII-minimized/non-PII after tenant erasure; if not, reopen GCR-R4 before inventing crypto-erasure machinery.
- **R10-C:** ManagedArtifactStore conformance suite; deliberate S3 client; scanner/parser ordering; production malware-inspection fail-closed gate; deployment-profile declaration integrity; staged-byte cleanup; temporary dev/CI S3 endpoint deletion/replacement condition.
- **R10-D:** provider provisioning/external-effect retry mechanics after B2 fixes semantic lifecycle; no distributed transaction.
- **R10-E:** use provider-hosted/themed login/recovery/MFA journeys; do not rebuild credential UI through Keycloak admin APIs.
- **R10-F:** migration/cutover deletes legacy credential/DEK machinery only from accepted target mappings; provider DB remains separate authority.

---

## Explicitly deferred from launch

These remain future triggers, not hidden V1 TODOs:

```text
Quarantine aggregate / periodic malware rescans / CDR / advanced active-content security
ArtifactSecurityAssessment domain
ICP-Brasil / PKI / DocuSign / Adobe Sign / RFC3161 / TSA / HSM
cryptographically signed export packages
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery / ESI preservation
realtime coauthoring / WOPI-style collaboration
true indivisible multi-file ArtifactPackage without a real triggering format
OpenFGA/SpiceDB without arbitrary relationship-sharing requirement
BPMN/Camunda/Flowable as Approval prerequisites
Temporal until R10-D proves repeated long-running durable-workflow/timer/retry/compensation machinery
mandatory tenant application-layer encryption / cryptographic erasure without a named Target Data family
self-hosted production object-store provider until a real deployment requires one
```

## Implementation gate

**CLOSED.** No product implementation starts in R10-B2. Product implementation begins only after the integrated R10 technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.
