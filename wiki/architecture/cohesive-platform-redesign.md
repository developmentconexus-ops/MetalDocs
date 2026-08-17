# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN / GCR-REFINED; R10-A CLOSED / GCR-REFINED; R10-B1 CLOSED; R10-B2 NEXT / NO PRODUCT IMPLEMENTATION AUTHORIZED**
> **Established:** 2026-08-14
> **R9.5 freeze ratified:** 2026-08-17
> **R10-A promotion ratified:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Repository baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** [`../engineering/standards/root-cause-global-maximum-method.md`](../../docs/engineering/standards/root-cause-global-maximum-method.md)
> **Frozen R3–R9.5 product/domain ledger:** [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md)
> **Active R10 technical authority:** [`r10-technical-architecture.md`](r10-technical-architecture.md)

## 1. Purpose

MetalDocs is being redesigned as one coherent product before the next large implementation wave.

Authentication, IAM, areas, approval routes, Documents, Controlled Documents, Templates, taxonomy, rendering and release evolved incrementally and created overlapping authority. Current code remains valuable evidence about real requirements, failures and operational constraints, but **current implementation shape is not admissible as proof that the target should keep the same nouns, modules, providers or boundaries**.

The target is the smallest professional architecture that:

- represents the real controlled-information domain correctly;
- gives every business fact one authority;
- preserves multi-tenancy, auditability, immutable evidence and fail-closed authorization;
- deletes duplicated lifecycle/policy implementations;
- uses mature commodity mechanisms where they reduce total complexity without becoming business authority;
- avoids speculative BPM, ReBAC, policy languages, storage providers, security platforms or generic infrastructure without a real consumer;
- preserves only extension seams justified by evidenced future requirements;
- is specified end-to-end before product implementation begins.

## 2. Fresh-session reading order

Any new session working on product/domain/technical architecture MUST start with `AGENTS.md`. The current route is:

1. [`../../AGENTS.md`](../../AGENTS.md)
2. [`../../docs/engineering/standards/root-cause-global-maximum-method.md`](../../docs/engineering/standards/root-cause-global-maximum-method.md)
3. [`../references/current-agent-handoff.md`](../references/current-agent-handoff.md) — current status / exact next step
4. **this file** — program authority / scope / global coherence
5. [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md) — frozen R3–R9.5 product/domain decisions, including promoted GCR refinements
6. [`r10-technical-architecture.md`](r10-technical-architecture.md) — active R10 technical-stage authority
7. review artifacts only when auditing a promoted gate

Do not start an old roadmap unit, milestone, migration, implementation PR or historical plan by inertia.

## 3. Authority during the redesign

For **target design** questions, authority is:

1. operator-approved decisions in the owning active authority:
   - R3–R9.5 product/domain semantics → frozen redesign ledger;
   - R10 technical decisions → `wiki/architecture/r10-technical-architecture.md`;
2. this page for program scope, authority routing and global coherence;
3. canonical cross-cutting standards;
4. final ADRs/specs explicitly retained or promoted by this program;
5. runtime/schema/OpenAPI/module docs as evidence only;
6. historical plans/specs/ADRs as evidence only.

For **what runs today**, runtime/code/database and OpenAPI remain authoritative.

Review artifacts are evidence, not parallel target authority. They may justify a promoted decision but do not override it after adjudication.

Global Coherence Review evidence chain:

- `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-fable-review-request.md`;
- `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-independent-fable-review.md`;
- `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-adjudicated-corrected-target.md`;
- `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-corrected-target-fable-delta-review.md`.

Final GCR result:

```text
APPROVE GCR ADJUDICATED CORRECTED TARGET
BLOCKER = 0
MAJOR   = 0
prior findings closed = 11/11
new material contradiction = NONE
fifth material local maximum = NONE
```

## 4. Frozen north star — GCR refined

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and are bound to MetalDocs organizational identity through a stable provider subject identity. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

```text
Authentication
provider subject binding / app Session / assurance
     │
     ▼
Organization ───────────── Authorization
Tenant                    Roles / Permissions
Areas                     Role Assignments
Users                     User + Group principals
Groups                    Tenant / Area scopes
     │                         │
     └──────────────┬──────────┘
                    ▼
          Controlled Information
          Document / Revision
          DocumentType
          Template-as-revision-role
          Artifact / WorkingContent
          Numbering / lifecycle
                    │
                    ▼
                Approval
         exact immutable Submission
         versioned sequential policy
         human steps + participants
         decisions + evidence
                    │
                    ▼
           Domain Governance
       freeze / SoD / lifecycle rules
                    │
                    ▼
           Release Coordinator
                    │
                    ▼
             Effective Revision
```

This diagram is product/domain-level. `Domain Governance` and `Release Coordinator` are not separate bounded contexts.

Supporting concerns consume those authorities instead of redefining them:

```text
Evidence / Dossier
Retention / Legal Hold / Disposition
Import / Historical Migration / Export
Audit
Rendering / Renditions
Periodic Review
Distribution / Read-Acknowledge
Notifications
Search / Projections
Token Dictionary / Computed Values
Tenant Lifecycle / Security
Async orchestration / outbox / jobs
External Repository Connectors
```

## 5. Frozen principal decisions — GCR-refined mirrors

### Authentication

- AuthN and product AuthZ stay separate.
- **Keycloak is the V1 Authentication provider.** It owns credential storage/policy, provider account activation/lockout, password recovery, MFA/passkeys, upstream OIDC/SAML/LDAP/AD federation and provider authentication journeys/session.
- MetalDocs Authentication owns provider-subject binding, opaque MetalDocs application Session, application-session lifecycle/revocation, authentication-assurance/fresh-auth facts and the provider anti-corruption contract.
- Stable provider identity is based on `issuer + subject`; email/username/display name are not technical identity.
- Provider roles/groups/organizations/permissions and arbitrary provider claims are not canonical MetalDocs Authorization inputs; no provider-role mapping or claim-to-permission bridge exists V1.
- Keycloak Organizations, if used for upstream-IdP routing, remain an AuthN projection of MetalDocs Tenant state, never product tenancy authority.
- Candidate provider topology is one realm per environment/application trust domain, not realm-per-Tenant, unless later material tenant-policy evidence invalidates that assumption.
- Keycloak/provider persistence is separate authority. No MetalDocs invariant may depend on an atomic transaction across the MetalDocs product-state DB and provider-owned persistence.

### Organization + Authorization

- `Area` belongs to Organization, not document taxonomy.
- V1 organization: Tenant, Area, User, Group, GroupMembership.
- Groups are flat and may receive ordinary RoleAssignments.
- Built-in roles: `tenant_owner`, `area_manager`, `author`, `approver`, `viewer`.
- Roles are bundles; checks use semantic Permissions.
- RoleAssignment subject is User or Group; scope is typed Tenant or Area.
- Grants compose additively; default deny; no explicit deny engine V1.
- `tenant_owner` is never a bypass.
- OpenFGA/SpiceDB are not required for V1.
- The frozen R9 + R9.5 catalogs contain 29 base + 16 bounded whole-product permissions; the exact catalog/bundles live in the frozen ledger.
- Organization owns no mandatory Tenant DEK/key-custody fact family V1. Such a family re-enters only if a named target data family later proves application-layer tenant encryption is necessary.

### Approval V1

- Specialized governed-information approval, not generic BPM.
- Versioned `ApprovalPolicy` with ordered sequential `ApprovalStep`s.
- Initial actor rules: named user, group, role-in-area.
- Completion: ANY or ALL only.
- Participants resolve on Step activation and are snapshotted; current authorization is rechecked when acting.
- Human outcomes: `accept`, `return_for_changes`.
- `return_for_changes` terminates the attempt; edited content is resubmitted as a new immutable Submission/ApprovalInstance as required.
- `withdraw`, `cancel`, `reassign` remain separate operations.
- Reauthentication may be required by a Step and consumes Authentication-owned assurance/fresh-auth facts, not local password challenges.
- Strict SoD: creator/submitter cannot accept own Submission; same user cannot accept two Steps of one ApprovalInstance; reassignment remains qualified and SoD-valid.
- No BPMN, generic branching, CEL, M-of-N or generic delegation/escalation engine V1.

### Controlled Information

- `documents`, `controlleddocuments`, `templates` do not survive as three target bounded contexts.
- Target core is `Document` + `DocumentRevision` inside Controlled Information.
- Separate `ControlledDocument` target object is retired.
- `DocumentProfile` converges into `DocumentType`; GovernanceClass is deleted.
- Template has no independent lifecycle/version counter; it is a designation/role of an exact governed DocumentRevision.
- Derived documents pin the exact effective source template Revision/hash used at creation.
- `Document` is stable identity; REV labels are `REV001`, `REV002`, ...; at most one EFFECTIVE + one open Revision V1.
- `RevisionSubmission` is the immutable attempt identity.
- Approval/Rendition/Release bind the same exact Submission/digest.
- Exactly one effective Revision per Document is a core invariant.

### WorkingContent / authoring

- `WorkingContent` is format-agnostic persisted DRAFT authority, independent of editor/provider.
- DRAFT uses one monotonic `working_version`/OCC across every governed mutation.
- Authorized DRAFT edit/upload/replacement is allowed; MetalDocs does not track arbitrary offline-file ancestry or require long checkout.
- Replacement is whole-WorkingContent and dispositions representation-dependent structured state/provenance in the same OCC step.
- Submission freezes one coherent accepted WorkingContent state; SUBMITTED rejects mutation.
- EigenPal is a DOCX provider/adapter, never Document or WorkingContent identity.
- Realtime coauthoring remains trigger-based/deferred.

### Content / storage / representation

- `Artifact` is immutable exact-byte technical identity with canonical SHA-256; provider URL/key/version never becomes business identity.
- Exactly one primary Artifact per DocumentRevision/Evidence V1.
- One active Managed Artifact Store/deployment V1.
- The first-class architecture is the **`ManagedArtifactStore` port + provider conformance contract**, not a provider-name list.
- Local is the first-class dev/test profile; AWS S3 is the reference production profile. Any additional/self-hosted provider is selected only for a real deployment requirement and must pass conformance.
- MinIO OSS has no product entitlement. A frozen MinIO image or another compatible endpoint may remain temporarily only as a dev/CI mechanism with a deletion/replacement condition owned by R10-C.
- Provider relocation copies exact bytes, verifies canonical SHA-256, then cuts over without creating new Artifact/REV/Submission.
- External repositories use explicit `IMPORT_COPY` / `PUBLISH_COPY`; no silent synchronization.
- Universal mandatory PDF is retired.
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`; at most one required derived rendition V1.
- Unsupported/unproven inline review representation falls back to a supported inspection path for the exact Submission; preview/viewer never becomes authority.
- Production confidentiality baseline remains encrypted transport + provider encryption at rest. V1 has no mandatory application-layer Tenant DEK.

### Dossier / Evidence

- Dossier = small stable documentary context, never ERP/PLM/custom-object authority.
- DossierType remains small; no custom forms/fields/workflow/ACL/hierarchy/completeness engine V1.
- Dossier↔Document is M:N over stable Document identity; links never grant access.
- CAPTURED Evidence has exactly one immutable primary Dossier and reuses its scope.
- Cross-scope queries/projections/exports reapply canonical AuthZ; contextual links are never transitive grants.

### Retention / Legal Hold / Erasure

- No generic Record declaration entity.
- CAPTURED Evidence and first-submitted DocumentRevision become retention subjects automatically.
- Frozen retention rules are explicit values and are snapshotted into RetentionBinding; expiry = disposition eligibility, never automatic deletion.
- Current EFFECTIVE Revision is never disposition-eligible.
- Physical disposal requires authorized disposition, no active hold and verified removal before DispositionRecord completion.
- LegalHold scopes: Evidence, stable Document, Dossier.
- Active Document/Dossier holds materialize current and newly entering retention subjects while within their live scope; unlink/lifecycle cannot release already-held subjects.
- Hold covers confirmed governed records, not never-submitted DRAFT/ESI; eDiscovery is future-triggered.
- Tenant erasure remains blocked while retention/hold obligations survive.
- V1 erasure uses verified substantive-row/blob deletion, an allowed PII-minimized/non-PII audit/platform skeleton, erasure/tombstone facts and restore reconciliation. There is no mandatory Tenant DEK/crypto-shred step.
- B6 must prove the immutable Audit skeleton is PII-minimized/non-PII. If it proves a real immutable Target Data family must remain stored yet become unintelligible, the DEK decision reopens before crypto-erasure machinery is added.
- R10-A does not introduce a standalone `RetentionPolicy` entity.

### Import / Historical Migration / Export

- Ordinary import follows normal lifecycle/target permissions.
- Historical Migration is privileged and never fabricates native approval/effectivity/actor history.
- Unknown source truth stays unknown; no fake Revision without exact bytes.
- Migration is batch/plan-based with true dry-run, deterministic outcomes and reconciliation; atomicity is per semantic unit.
- Backup, Tenant Portability Export, Governed Subject Export and external `PUBLISH_COPY` are distinct contracts.
- Portability/governed exports use provider-independent manifests with canonical hashes and no secrets/runtime internals.
- Export completeness must be explicit and authorization-safe; a contract claiming completeness fails closed rather than silently omitting required unauthorized subjects.

### Launch attestation / content safety

- V1 claims authenticated application approval, not ICP-Brasil/qualified-signature semantics.
- ApprovalDecision preserves exact Submission/digest + actor/Step/policy/server-time/AuthN assurance evidence.
- Approved source bytes are never stamped/mutated; human-readable manifestations are derived.
- Launch content safety includes supported-format allowlist, size/type/structural coherence and a **production-required malware-inspection gate before untrusted bytes become `CONFIRMED Artifact`**.
- Scanner unavailable/incomplete/malicious result means no confirmation. Dev/test profiles may explicitly disable inspection; a disabled profile must not be able to identify itself as production.
- No tenant-facing scanning-policy authority exists.
- Quarantine/CDR/rescan/intelligence/sandbox security platforms remain future-triggered.

## 6. Whole-product freeze status after GCR

R9.5 remains complete and frozen, with only the promoted GCR bounded refinements:

```text
R9.5-1 Content Model                         LOCKED
R9.5-2 Storage / Repository Strategy         LOCKED / GCR-REFINED
R9.5-3 Authoring / EigenPal                  LOCKED (R9.5-8 refinement)
R9.5-4 Dossier / Context                     LOCKED
R9.5-5 Retention / Records / Legal Hold      LOCKED (R9.5-8 refinement)
R9.5-6 Import / Migration / Export           LOCKED
R9.5-7 Launch Attestation / Content Safety   LOCKED / GCR-REFINED
R9.5-8 Whole-Product Adversarial Freeze      CLOSED / APPROVED
R9.5                                         FROZEN / GCR-REFINED
reopen set                                   EMPTY
```

Do not reopen frozen decisions for preference or hypothetical futures. GCR itself found no fifth material local maximum.

## 7. R10-A promoted ownership topology after GCR

R10-A remains **CLOSED / APPROVED** with exactly 8 business bounded contexts + 3 supporting semantic owners. Detailed authority lives in [`r10-technical-architecture.md`](r10-technical-architecture.md).

### Business bounded contexts — exactly 8

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

### Supporting semantic owners — exactly 3

```text
Artifact
Audit
Interchange
```

### Attributed support / projection classification

```text
Notifications → internal/support/notifications
Search        → internal/projections/search
```

Key ownership rulings:

- Authentication owns provider subject binding, MetalDocs app Session, app-session lifecycle/revocation, authentication assurance/fresh-auth and provider anti-corruption semantics; credential mechanisms belong to Keycloak/provider.
- Organization owns Tenant/Area/User/Group/GroupMembership, Tenant settings/configuration, Tenant lifecycle and erasure tombstones. It owns no mandatory V1 tenant key-custody fact family.
- Authorization owns grants/evaluation/composition contract shape; domains own relationship predicate meaning.
- Controlled Information owns the one Document/Revision/WorkingContent/Submission/Template/Rendition/Release lifecycle authority.
- Document owner/responsibility meaning belongs to Controlled Information; R10-A does not prematurely fix participant type/cardinality/representation.
- Tenant Dictionary + System Value Catalog live as distinct internal Controlled Information fact classes; no standalone Dictionary owner.
- Approval owns policy/instance/participants/decisions/attestation/SoD, never Document effectivity.
- Documentary Context owns Evidence/Dossier/ExternalReference/context relationships.
- Records Governance owns retention-rule meaning, bindings, holds and disposition; no standalone V1 `RetentionPolicy` entity.
- Artifact owns exact-byte/physical-content truth, not business lifecycle.
- Audit owns the transversal timeline, not domain state.
- Interchange owns transfer process truth, never imported target-object truth.
- Composition coordinates concrete cross-owner use cases but owns no durable semantic fact.
- Notifications owns attributed delivery/read state only; Search remains rebuildable projection only.

## 8. Current module disposition

Current module docs are current-state evidence, not target architecture. R10-A fixes the semantic disposition:

- `approval` → Approval V1;
- `audit` → Audit supporting semantic owner;
- `auth` → Authentication provider-subject binding + MetalDocs app-session/assurance semantics; local credential machinery is migration/deletion evidence only;
- `controlleddocuments` → delete target BC; responsibilities → Controlled Information;
- `distribution` → Distribution;
- `documents` → delete legacy BC; responsibilities → Controlled Information;
- `iam` → split Organization + Authorization;
- `jobs` → no BC; owner-attributed work + composition/platform async;
- `notifications` → `support/notifications`;
- `render` → dismantle; Rendition/value semantics → Controlled Information, providers → infrastructure;
- `search` → `projections/search`;
- `security` → delete BC; product-facing AuthN facts → Authentication, commodity security → platform; legacy Tenant-DEK/KEK machinery has no V1 target entitlement;
- `taxonomy` → dismantle: Area → Organization; DocumentType/classification → Controlled Information; GovernanceClass deleted;
- `templates` → delete parallel lifecycle; template role/designation/use → Controlled Information;
- `tokens` → delete standalone owner; Tenant Dictionary/System Value Catalog → Controlled Information.

Exact table/API/frontend cutover mechanics remain later R10 work.

## 9. Build-vs-buy posture after GCR

Use commodity mechanisms when they reduce total complexity without acquiring product/domain authority.

Current rulings:

- **Keycloak/OIDC: V1 Authentication provider.** Provider roles/groups/organizations never become canonical MetalDocs Organization/AuthZ authority.
- Managed Artifact Store: port + conformance is the first-class target; Local dev/test and AWS S3 reference production profile; no frozen self-hosted provider without a real consumer.
- OpenFGA/SpiceDB: defer until arbitrary relationship-sharing/hierarchy graph justifies it.
- Camunda/Flowable/BPMN: not for Approval V1.
- Temporal: not an Approval prerequisite; reconsider in R10-D only if repeated long-running durable workflow/timer/retry/compensation machinery becomes a real defect class.
- CEL/expression language: defer until typed configuration cannot represent a real policy requirement.
- External ECM/JCR kernel: not the MetalDocs domain kernel.
- SharePoint Embedded/M365: future enterprise content profile, not universal storage baseline.
- PKI/qualified e-signature, eDiscovery and realtime coauthoring remain trigger-based.

## 10. Documentation lifecycle

`wiki/` holds durable maintained product/repository truth; `docs/` holds active staging/working evidence unless an owner explicitly says otherwise. Git history is the archive.

Authority is delegated by stage:

- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` remains the frozen binding detailed R3–R9.5 product/domain record, including promoted GCR bounded refinements;
- `wiki/architecture/r10-technical-architecture.md` is the active durable R10 technical-stage authority;
- review packets remain evidence of how decisions were challenged and adjudicated; they are not parallel authority.

## 11. Implementation gate

**Closed.** No product implementation yet.

R10-B2 through R10-F must complete persistent state, constraints, transactions/events, physical integrity, async/external effects, API/frontend journeys, migration/delete map, durable target specs/ADRs and required adversarial/operator gates. Only then may an implementation plan be authored from the accepted target.

## 12. Exact next step — R10-B2 Authentication / Organization / Authorization State

R10-A and R10-B1 are closed. Start **R10-B2** in design-only mode from the promoted GCR-refined topology and substrate.

B2 must derive at minimum:

```text
provider subject binding representation and tenant-scoped uniqueness/cardinality
Authentication ↔ Organization User binding integrity
opaque MetalDocs application Session and provider-disable/live-session posture
fresh-auth / authentication assurance representation
structural provider anti-corruption proof; no provider role/group/org/permission consumption
provider provisioning/binding idempotency + reconciliation choreography

Tenant / Area / User / Group / GroupMembership persistent state
Tenant settings/configuration persistence
Tenant lifecycle ACTIVE/SUSPENDED/ERASED
TenantDeletionRequest / TenantErasureRecord / tombstone + restore-reconciliation state

Permission / Role / RoleAssignment representation
User|Group subject and Tenant|Area typed scope
grant/revocation evidence
canonical grant-evaluation read surface
same-Tenant FK + RLS application under B1
semantic persistence + mutation-law classification
membership/grant/lifecycle transaction boundaries
required same-commit Audit/durable-intent insertion points
```

B2 does not design local credentials/MFA/password/lockout tables, Keycloak role mappings, Tenant DEK/KEK infrastructure or cross-provider distributed transactions.

Later failure classes remain deferred unless a minimal seam is needed:

```text
R10-C → Artifact/storage/relocation/restore + ManagedArtifactStore conformance + malware gate/profile integrity
R10-D → durable async/projections/retry/external-effect execution
R10-E → final API/frontend journeys, including provider-hosted authentication journeys
R10-F → Historical Migration/cutover/final deletion map
```

Technology/topology choices remain subordinate to ownership and invariants. Current runtime/code/schema/OpenAPI are evidence, never target entitlement.
