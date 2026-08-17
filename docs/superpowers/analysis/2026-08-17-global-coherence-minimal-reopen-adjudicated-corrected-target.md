# MetalDocs Global Coherence Review — Adjudicated Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — PENDING BOUNDED DELTA CHECK — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline before GCR:** `9f2f0a4ca2e390e67a2351cfd6ccaa578f5d690d`
> **GCR candidate:** `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-fable-review-request.md` @ `7a9125e9`
> **Independent review:** `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-independent-fable-review.md` @ `ccf97578`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this artifact records operator adjudication and the corrected target only. It does not amend R9.5, R10-A, R10-B1, current-agent-handoff, code, schema, OpenAPI, frontend, deployment, or implementation.

---

# 1. Independent review result

```text
VERDICT: APPROVE GCR MINIMAL REOPEN SET WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 6
LOW     = 5
FIFTH MATERIAL LOCAL MAXIMUM = NONE
```

Adjudication principle:

> Findings are evidence, not authority. Apply only corrections that remove the identified defect class while preserving the smallest sustainable architecture. Do not turn a review fix into a new framework, product requirement or duplicate authority.

---

# 2. Operator adjudication of findings

| Finding | Decision | Corrected target |
|---|---|---|
| M1 — Keycloak also changes promoted R10-A Authentication facts | **ACCEPT** | Bounded R10-A amendment: Authentication retains provider subject binding, MetalDocs application Session, authentication-assurance/fresh-auth facts and provider anti-corruption contract. Credential storage/policy/activation/lockout moves to the IdP provider. 8+3 topology unchanged. |
| M2 — provider DB cannot share product-DB atomicity | **ACCEPT** | No MetalDocs invariant may depend on atomic commit across the MetalDocs product-state DB and a provider-owned DB. B2 must model provider-binding/provisioning as idempotent choreography with reconciliation. |
| M3 — Keycloak claims must not become a second AuthZ authority | **ACCEPT / STRENGTHEN** | The Authentication anti-corruption contract exposes only stable provider subject + enumerated assurance facts. No provider role/group/org/permission claims have a representation consumable by MetalDocs Authorization. |
| M4 — current DEK has a real legacy consumer (`audit_events.payload`) | **ACCEPT / ROOT-CAUSE RESTRUCTURE** | Current audit-payload crypto is evidence, not target entitlement. V1 target removes mandatory Tenant DEK/KEK/crypto-shred by designing immutable Audit as a non-PII/minimized skeleton whose human enrichment resolves through separately erasable state. If B6 proves an immutable Target Data family must remain stored yet become unintelligible, this decision reopens. |
| M5 — malware gate could be vacuous by profile | **ACCEPT CAUSE / REJECT PRODUCTION OPT-OUT** | Production requires malware inspection before untrusted bytes become CONFIRMED. Unavailable/incomplete inspection fails closed. Explicit dev/test profiles may disable inspection. No tenant-facing or ordinary production opt-out exists. |
| M6 — first-class storage surface is the port, not a provider list | **ACCEPT** | `ManagedArtifactStore` port + conformance contract is the durable first-class surface. Local is first-class dev/test. AWS S3 is a reference production profile. No self-hosted provider is frozen without a real consumer. |
| L1 — `minio-go/v7` inheritance | **ACCEPT / ROUTE R10-C** | Select S3 client library deliberately; current dependency has no target entitlement. |
| L2 — scanner vs parser/validator order | **ACCEPT / ROUTE R10-C** | R10-C decides safe validation/inspection ordering and validator hardening without introducing a sandbox platform absent evidence. |
| L3 — delete custom credential journeys under Keycloak | **ACCEPT / ROUTE R10-E** | MetalDocs does not rebuild password-reset/MFA-enrollment/recovery UI via provider admin APIs. Use provider-hosted/themed authentication journeys where appropriate. |
| L4 — C2 cross-DB atomicity wording | **ACCEPT** | Add the explicit no-cross-provider-DB-atomicity law. |
| L5 — fail-open current configuration | **ACCEPT AS IMPLEMENTATION EVIDENCE** | Later implementation proof must show production configuration cannot silently disable a required security/correctness property. No current implementation behavior is promoted. |

No finding creates a fifth bounded context or new semantic owner.

---

# 3. Corrected GCR-R1 — Keycloak V1 Authentication provider

```text
OUTCOME = RESTRUCTURE NOW
```

## 3.1 Authority split

Keycloak/provider owns Authentication mechanisms:

```text
credential storage
password policy
credential activation / provider account enablement
provider lockout / brute-force protection
password recovery
MFA / passkeys
upstream OIDC / SAML / LDAP / AD federation
provider authentication session
provider-hosted authentication journeys
```

MetalDocs Authentication owns product-facing authentication semantics:

```text
ProviderSubjectBinding
MetalDocs opaque application Session
application-session revocation/lifecycle
authentication assurance facts
fresh-auth / reauthentication evidence
provider anti-corruption contract
```

Organization continues to own:

```text
Tenant
Area
User
Group
GroupMembership
```

Authorization continues to own:

```text
Permission
Role
RoleAssignment
canonical grant evaluation
```

No Keycloak entity becomes product/domain authority.

## 3.2 Stable provider identity

The target binding uses provider identity:

```text
issuer + subject
```

Email, username and display name are attributes, not stable technical identity.

Conceptual target:

```text
AuthenticationSubjectBinding
  tenant_id
  id
  organization_user_id
  issuer
  subject
  created_at
```

B2 must decide the exact cardinality and uniqueness rules, including whether one Organization User may bind more than one provider subject in V1/future, while guaranteeing duplicate `(issuer, subject)` binding rejection.

## 3.3 Structural anti-corruption contract

The published Authentication provider result may expose only enumerated facts such as:

```text
issuer
subject
authenticated_at
auth_time
acr? / assurance level?
amr? / authentication methods?
```

It must not expose a generic claims map to Authorization or domain owners.

Forbidden as canonical inputs to MetalDocs Authorization:

```text
provider roles
realm roles
client roles
provider groups
provider organizations
provider permissions
arbitrary claim-to-permission mappings
```

There is no provider-role mapping table and no claim-to-MetalDocs-permission bridge V1.

## 3.4 Application session remains distinct

```text
Keycloak/IdP session != MetalDocs application Session
```

The MetalDocs Session remains opaque to the browser and carries application authentication context, not Authorization state.

It may carry references/assurance such as:

```text
session identity
tenant/application context
organization_user_id
provider subject reference
authenticated_at
fresh-auth/assurance state
expiry / revocation
```

It does not snapshot canonical roles, permissions, groups or Area grants.

## 3.5 Provider persistence and non-atomic choreography

Keycloak/provider persistence is outside the MetalDocs product-state DB authority.

New law:

> No MetalDocs invariant may require an atomic transaction across the MetalDocs product-state database and any provider-owned database.

B2/R10-D must design idempotent choreography and reconciliation for at least:

```text
Organization User exists / provider subject absent
provider subject exists / MetalDocs binding absent
binding exists / provider subject removed or disabled
duplicate issuer+subject attempted
provider temporarily unavailable
retry after uncertain provider response
```

No XA/2PC/distributed transaction is introduced.

## 3.6 Topology

Candidate provider topology:

```text
one Keycloak realm per environment/application trust domain
not one realm per MetalDocs Tenant
```

Tenant-specific upstream IdP routing may later use Keycloak Organizations or equivalent provider mechanisms only as an Authentication routing/federation projection of MetalDocs Tenant state. Provider Organizations remain non-authoritative for product tenancy.

Exact Keycloak HA/sizing/config automation is implementation/operations work, not GCR authority.

---

# 4. Corrected GCR-R2 — Managed Artifact Store port + conformance

```text
OUTCOME = RESTRUCTURE PROVIDER ENTITLEMENT
```

Durable architecture:

```text
ManagedArtifactStore port
+ provider conformance contract
```

Profiles:

```text
Local     = first-class dev/test provider
AWS S3    = reference production provider profile
other provider = selected only from a real deployment requirement and must pass conformance
```

MinIO OSS has no frozen V1 product entitlement.

A frozen MinIO image or another compatible endpoint may temporarily remain a dev/CI execution mechanism while R10-C lands a deliberate conformance environment. That transitional dependency is not product authority and must have a deletion/replacement condition.

Minimum R10-C conformance properties include:

```text
put/presign-confirm round-trip
exact-byte SHA-256 verification
over-size rejection
hash-mismatch rejection + cleanup
tenant namespace isolation
no overwrite of immutable existing object keys
copy + verify + cutover relocation semantics
restore byte-integrity verification
safe delete semantics
```

R10-C selects the concrete S3 client library deliberately; the current `minio-go/v7` dependency is evidence only.

---

# 5. Corrected GCR-R3 — secure-by-default malware inspection gate

```text
OUTCOME = BOUNDED RESTRUCTURE
```

Target production flow:

```text
untrusted bytes
→ STAGED
→ bounded size / ContentFormat / structural-integrity validation
→ malware inspection
→ semantic-owner attachment/confirmation validation
→ CONFIRMED Artifact
```

Production invariant:

> Untrusted bytes cannot become `CONFIRMED Artifact` without a successful malware-inspection result.

Failure posture:

```text
scanner unavailable      → remain non-confirmed / operation fails visibly
inspection incomplete    → remain non-confirmed
malware detected         → never confirm
```

Explicit dev/test profiles may disable inspection. Ordinary production profiles cannot silently opt out.

No new semantic owner or business lifecycle is introduced. V1 does not create:

```text
Quarantine aggregate
CDR platform
periodic rescanning platform
malware intelligence platform
custom sandbox cluster
ArtifactSecurityAssessment domain
macro-enabled Office support
```

R10-C owns scanner selection, scanner availability/retry mechanics, ordering relative to parsers/validators, safe staged-byte cleanup, and proof that the confirmation gate cannot be bypassed.

---

# 6. Corrected GCR-R4 — remove mandatory Tenant DEK from V1

```text
OUTCOME = RESTRUCTURE / REMOVE V1
```

## 6.1 Root-cause decision

The current implementation encrypts one legacy family (`audit_events.payload`) with a Tenant DEK. That proves the mechanism can exist; it does not prove the target must preserve an immutable PII payload and therefore carry a crypto-shred subsystem.

The frozen target already requires post-erasure preservation of only an allowed **non-PII audit/platform skeleton**.

Therefore V1 target chooses the smaller structure:

```text
immutable AuditEvent skeleton = PII-minimized / non-PII durable authority
human-readable/user enrichment = resolved through separately erasable Organization or other owned state
```

Tenant erasure removes the linking/substantive state while the permitted immutable audit skeleton may remain.

## 6.2 V1 deletions

Remove as mandatory V1 concepts:

```text
Tenant DEK lifecycle
Organization tenant key-custody fact family
mandatory crypto-shred step in Tenant erasure
mandatory platform KEK integration
mandatory wrap/unwrap machinery
mandatory envelope-encryption subsystem solely for tenant erasure
```

The Tenant erasure invariant remains:

```text
evaluate retention / active holds
→ block while lawfully retained state requires preservation
→ revoke/suspend access/session capability as required
→ erase eligible substantive rows/blobs
→ preserve only allowed non-PII audit/platform skeleton
→ record authoritative erasure/tombstone facts
→ Tenant ERASED
```

Backup/restore still must reapply tombstones and reconcile retention/hold state before service resumes.

## 6.3 B6 proof obligation

B6 must prove the target Audit schema does not require immutable, non-erasable PII payloads.

At minimum classify:

```text
which AuditEvent fields are durable non-PII skeleton
which actor/resource identifiers are opaque references
which human-readable enrichment is separately erasable or projection-only
which event facts must remain after Tenant erasure
```

If B6 demonstrates a real immutable Target Data family that must remain physically stored while becoming unintelligible after lawful erasure, R4 reopens and the minimum envelope/key-custody design is reconsidered with R10-C/R10-F backup/restore proof.

No cryptographic-erasure claim may exist without a named Target Data family and fail-closed enforcement.

---

# 7. Corrected clarification C1 — North Star identity meaning

Replace the ambiguous interpretation of “MetalDocs is the system of record for identity” with:

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and are bound to MetalDocs organizational identity through a stable provider subject identity.**

MetalDocs answers who the User is in the product/organization. The provider answers how the external subject authenticated.

---

# 8. Corrected clarification C2 — product-state database scope

R10-B1 remains structurally unchanged.

Wording clarification:

```text
one MetalDocs product-state PostgreSQL database
canonical target product-state schema = metaldocs
```

Provider-owned products retain separate persistence authority, migrations, credentials and lifecycle.

New invariant:

> **No MetalDocs invariant may depend on cross-database atomicity between the MetalDocs product-state database and any provider-owned database.**

Physical co-location on one PostgreSQL server/cluster does not merge logical persistence authority.

---

# 9. Confirmed architecture / no further material reopen found

The independent cold review found no fifth material local maximum. These decisions remain confirmed absent new material evidence:

```text
modular monolith
8 business bounded contexts + 3 supporting semantic owners
one MetalDocs product-state relational substrate
composite tenant identity / same-tenant FKs / fail-closed RLS
specialized Approval rather than generic BPM
WorkingContent + OCC + immutable RevisionSubmission
Artifact identity separate from storage provider
Dossier/Evidence bounded documentary context
Records Governance / retention / hold / disposition semantics
Audit as semantic evidence owner
same-commit Audit law
outbox/durable-intent law
Search as rebuildable projection
Notifications as attributed support
Historical Migration / Interchange boundary
OpenFGA/SpiceDB deferred
Camunda/Flowable/BPM deferred
Temporal deferred with explicit R10-D trigger
Elasticsearch/OpenSearch deferred until query/scale evidence
SharePoint Embedded deferred as enterprise content profile
realtime collaboration deferred
PKI/qualified e-signature deferred
eDiscovery/ESI deferred
```

Near-misses are successor decisions only:

```text
dev/CI S3 conformance endpoint → R10-C
S3 client library              → R10-C
credential-journey deletion    → R10-E
```

---

# 10. Resulting bounded authority amendments — candidate only

If the bounded delta review approves this corrected target, promotion should change only the implicated authority text.

## R9.5 bounded deltas

```text
§1 Authentication
  local credential V1 → Keycloak/provider V1 + stable subject binding

§7 North Star
  refine identity to product/organizational identity

§9 / R9.5-2 Storage
  first-class surface = ManagedArtifactStore port + conformance
  remove MinIO OSS nominal entitlement
  Local dev/test + AWS S3 reference profile
  remove mandatory Tenant DEK statement

§14 / R9.5-7 Content Safety
  production secure-by-default malware inspection before Artifact confirmation

§6 Tenant lifecycle / Platform Security
  remove mandatory Tenant DEK destruction / crypto-shred step
  preserve retention-aware verified deletion + non-PII skeleton + tombstone/restore reconciliation
```

Everything else remains frozen.

## R10-A bounded amendment

```text
Authentication owned facts:
  provider subject binding
  opaque MetalDocs application Session
  application revocation/lifecycle
  assurance/fresh-auth facts
  provider anti-corruption contract

REMOVE from Authentication ownership:
  credential storage/policy/activation/lockout
  (provider-owned mechanisms)

Organization:
  REMOVE mandatory tenant key-custody lifecycle fact family V1

Commodity/platform mechanisms:
  Keycloak/IdP adapter belongs Authentication infrastructure
  REMOVE mandatory KEK/wrap-unwrap V1 mechanism
```

Topology remains exactly 8+3.

## R10-B1 wording-only amendment

```text
one MetalDocs product-state PostgreSQL database
+ no invariant may rely on atomicity with provider-owned databases
```

All other B1 laws remain unchanged.

---

# 11. R10-B2 scope after promotion — candidate

After authority promotion, B2 may resume with:

```text
B2-1 Authentication
  provider subject binding
  app Session
  assurance/fresh-auth
  provider-binding lifecycle + reconciliation
  structural no-provider-role-consumption proof

B2-2 Organization
  Tenant / Area / User / Group / GroupMembership

B2-3 Authorization
  Permission / Role / RoleAssignment

B2-4 Tenant lifecycle
  ACTIVE / SUSPENDED / ERASED
  TenantDeletionRequest
  TenantErasureRecord
  tombstone/reconciliation state
  NO mandatory Tenant DEK V1

B2-5 coherence
  user provisioning/offboarding
  Keycloak/provider choreography
  membership/grant lifecycle
  Tenant suspension/deletion interactions
  required Audit/durable-intent boundaries
```

B2 must not design password/MFA/credential storage tables or tenant KEK/DEK infrastructure.

---

# 12. Bounded delta review questions

The next independent reviewer must inspect **only the corrected delta**, not restart the whole-platform review by default.

Required questions:

1. Does M1 now correctly amend Authentication ownership without changing the 8+3 topology?
2. Is the Keycloak/provider anti-corruption contract structural enough to prevent provider roles/groups/orgs from becoming canonical AuthZ?
3. Does the no-cross-database-atomicity law make provider provisioning/binding operable rather than underspecified?
4. Is keeping the opaque MetalDocs application Session still coherent after the corrected provider boundary?
5. Does M6 now remove provider-name entitlement rather than merely replacing MinIO with another provider?
6. Is the transitional dev/CI S3 endpoint correctly mechanism-only and safely deferred to R10-C?
7. Is production malware inspection truly secure-by-default and non-vacuous, while remaining bounded enough to avoid quarantine/CDR/security-platform scope?
8. Does removing mandatory Tenant DEK V1 preserve the real tenant-erasure invariant?
9. Is the non-PII Audit skeleton approach coherent with append-only/tamper-evident Audit, retention and post-erasure evidence?
10. Does R4 retain an explicit reopen trigger if B6 proves a real immutable Target Data family requiring crypto-erasure?
11. Do C1/C2 eliminate ambiguity without creating new authority?
12. Did any correction create a new material contradiction or fifth local maximum?
13. Can R10-B2 resume after promotion with the scope in §11?

Required verdict:

```text
APPROVE GCR ADJUDICATED CORRECTED TARGET
APPROVE GCR ADJUDICATED CORRECTED TARGET WITH MATERIAL FIXES
DO NOT APPROVE GCR ADJUDICATED CORRECTED TARGET
```

A further broad GCR is unnecessary unless the delta itself creates material new evidence.

---

# 13. Current gate

```text
authority baseline = unchanged
R9.5              = still FROZEN authority pending promotion
R10-A             = still promoted authority pending bounded amendment
R10-B1            = still promoted authority pending wording clarification
R10-B2            = PAUSED pending bounded delta review + promotion
implementation    = BLOCKED
```

No product implementation, schema migration, API/frontend change, provider deployment or authority promotion is authorized by this artifact.