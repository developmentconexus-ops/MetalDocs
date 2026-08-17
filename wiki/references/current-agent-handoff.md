# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN / R10-A CLOSED / R10-B NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical-architecture authority
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

The old R3–R9.5 ledger remains binding for frozen semantics. Its historical `R10 = NEXT` routing text is superseded by the current program/stage authorities; do not edit frozen product semantics merely to make that historical status line current.

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
R10-B   = NEXT / DESIGN ONLY
R10-C   = NOT STARTED
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

R10-A promotion is authoritative in:

`wiki/architecture/r10-technical-architecture.md`

---

## R10-A promoted outcome

The target ownership set is now fixed for V1 as:

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

`jobs`, outbox/queue/workers, storage/render/connector providers, RLS, HTTP/OpenAPI/codegen, cache, rate limiting, observability and crypto primitives are mechanisms, not semantic owners.

Key R10-A promoted rulings:

- Tenant/Area/User/Group/GroupMembership and Tenant lifecycle belong to Organization.
- Tenant settings/configuration are part of the Organization/Tenant fact family; no standalone `TenantSettings` owner is introduced.
- Tenant DEK/key-custody lifecycle facts belong to Organization; crypto/KEK mechanics remain platform.
- Authorization owns grants/evaluation/composition contract shape; each domain owns its relationship predicates.
- Documents/ControlledDocuments/Templates converge into one Controlled Information lifecycle authority.
- Document owner/responsibility meaning belongs to Controlled Information, but R10-A deliberately does not choose its participant type/cardinality/physical representation.
- Tenant Dictionary and System Value Catalog are internal Controlled Information fact classes; no standalone Dictionary owner.
- No V1 `RetentionPolicy` entity: DocumentType/EvidenceType store frozen retention-rule values directly; Records Governance owns rule meaning/bindings/holds/disposition.
- Evidence/Dossier belong to Documentary Context.
- Rendition/release/effectivity semantics belong to Controlled Information; render providers are mechanisms.
- Artifact is the exact-byte/physical-content supporting owner shared by Controlled Information and Documentary Context.
- Notifications owns delivery/inbox/read state only; Search is rebuildable projection only.
- Interchange owns transfer process truth, never imported target business truth.
- Composition coordinates concrete cross-owner use cases but owns no durable meaning.

The package classification, dependency/seam rules, legacy delete/split map, surface classification, closure evidence and reopen triggers live in the R10 stage authority.

---

## R10-A review/closure record

R10-A underwent:

1. independent adversarial ownership review;
2. Method adjudication and correction;
3. cold delta/global coherence review;
4. final completeness correction;
5. independent mechanical frozen-fact/permission-catalog sweep;
6. final operator adjudication.

Final closure evidence found:

```text
BLOCKER                    = 0
remaining topology defect = 0
duplicate owners           = 0
invented fact families     = 0
RetentionPolicy entity     = ABSENT / PASS
R9.5 reopen set            = EMPTY
```

Review artifacts remain evidence, not authority. Do not re-litigate R10-A for package naming preference, current-schema convenience, provider capability or hypothetical futures.

---

## Exact next step — R10-B

Start **R10-B — Transactional Domain State & DB Invariants** from the promoted R10-A ownership topology.

Do not make one microdecision at a time. First perform a whole-block decomposition of the material state/invariant decisions that belong to R10-B, then close them in coherent clusters.

At minimum R10-B must derive:

```text
semantic owner → table/aggregate ownership
identity/key/FK/reference rules across owners
Document/Revision/WorkingContent/Submission persistent state
Document owner/responsibility representation only to the minimum frozen semantics
one-open-Revision and exactly-one-EFFECTIVE DB enforcement
WorkingContent working_version OCC
Submission coherent-state atomicity + immutability
Approval policy/instance/participant-snapshot/decision constraints
SoD invariant backstops
Evidence/Dossier lifecycle and relation constraints
RetentionBinding/LegalHold/Disposition state constraints
Artifact no-confirmed-orphan structural backstop
Tenant deletion/erasure/tombstone/key-custody durable state
local cross-owner transaction boundaries
same-commit Audit append seam
outbox-intent insertion points required by atomic business changes
historical-import representation that cannot fabricate native Approval/Release facts
```

Boundary with later blocks:

```text
R10-C → physical Artifact/storage/relocation/restore mechanics
R10-D → durable async execution, retries, projections, external effect truth
R10-E → final API contracts + frontend journeys + canonical access surfaces
R10-F → Historical Migration/cutover/final deletion map
```

R10-B may prepare a seam needed by those blocks but must not pull their mechanism design forward without necessity.

Use the DevelopmentConexus Engineering Method v1.0.0. Current runtime/schema/tests may be inspected as evidence for a specific decision, never as target entitlement.

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

**CLOSED.** No product implementation starts in R10-B. Product implementation begins only after the integrated R10 technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.
