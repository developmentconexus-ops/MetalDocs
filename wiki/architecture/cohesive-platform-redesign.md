# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN / R10 NEXT / NO PRODUCT IMPLEMENTATION AUTHORIZED**
> **Established:** 2026-08-14
> **R9.5 freeze ratified:** 2026-08-17
> **Repository baseline inspected:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** [`../engineering/standards/root-cause-global-maximum-method.md`](../../docs/engineering/standards/root-cause-global-maximum-method.md)
> **Detailed active ledger:** [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md)

## 1. Purpose

MetalDocs is being redesigned as one coherent product before the next large implementation wave.

Authentication, IAM, areas, approval routes, Documents, Controlled Documents, Templates, taxonomy, rendering and release evolved incrementally and created overlapping authority. Current code remains valuable evidence about real requirements, failures and operational constraints, but **current implementation shape is not admissible as proof that the target should keep the same nouns, modules or boundaries**.

The target is the smallest professional architecture that:

- represents the real controlled-information domain correctly;
- gives every business fact one authority;
- preserves multi-tenancy, auditability, immutable evidence and fail-closed authorization;
- deletes duplicated lifecycle/policy implementations;
- avoids speculative BPM, ReBAC, policy languages or external identity infrastructure;
- preserves only extension seams justified by evidenced future requirements;
- is specified end-to-end before product implementation begins.

## 2. Fresh-session reading order

Any new session working on product/domain/technical architecture MUST start with `AGENTS.md`. The current route is:

1. [`../../AGENTS.md`](../../AGENTS.md)
2. [`../../docs/engineering/standards/root-cause-global-maximum-method.md`](../../docs/engineering/standards/root-cause-global-maximum-method.md)
3. [`../references/current-agent-handoff.md`](../references/current-agent-handoff.md) — current status / exact next step
4. **this file** — program authority / scope
5. [`../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`](../../docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md) — binding detailed decisions
6. R9.5-8 review artifacts only when auditing the freeze.

Do not start an old roadmap unit, milestone, migration, implementation PR or historical plan by inertia.

## 3. Authority during the redesign

For **target design** questions, authority is:

1. operator-approved decisions in the active ledger;
2. this page for program state/scope;
3. canonical cross-cutting standards;
4. final ADRs/specs explicitly retained or promoted by this program;
5. runtime/schema/OpenAPI/module docs as evidence only;
6. historical plans/specs/ADRs as evidence only.

For **what runs today**, runtime/code/database and OpenAPI remain authoritative.

Review artifacts are evidence, not parallel target authority. R9.5-8 review evidence lives at:

- `docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md`;
- `docs/superpowers/analysis/2026-08-17-r9.5-8-independent-adversarial-challenge.md`.

The independent challenge verdict `APPROVE / FREEZE R9.5` was operator-ratified and promoted into the active ledger on 2026-08-17.

## 4. Frozen north star

```text
Authentication
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

## 5. Frozen principal decisions

### Authentication

- AuthN and product AuthZ stay separate.
- Current MetalDocs authentication/session approach is acceptable for V1 behind the AuthN boundary.
- Keycloak/external IdP remains future-triggered.

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
- The frozen R9 + R9.5 catalogs contain 29 base + 16 bounded whole-product permissions; the exact catalog/bundles live in the active ledger.

### Approval V1

- Specialized governed-information approval, not generic BPM.
- Versioned `ApprovalPolicy` with ordered sequential `ApprovalStep`s.
- Initial actor rules: named user, group, role-in-area.
- Completion: ANY or ALL only.
- Participants resolve on Step activation and are snapshotted; current authorization is rechecked when acting.
- Human outcomes: `accept`, `return_for_changes`.
- `return_for_changes` terminates the attempt; edited content is resubmitted as a new immutable Submission/ApprovalInstance as required.
- `withdraw`, `cancel`, `reassign` remain separate operations.
- Reauthentication may be required by a Step.
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
- One Managed Artifact Store/deployment V1; first-class Local/MinIO/AWS S3 adapters.
- Provider relocation copies exact bytes, verifies canonical SHA-256, then cuts over without creating new Artifact/REV/Submission.
- External repositories use explicit `IMPORT_COPY` / `PUBLISH_COPY`; no silent synchronization.
- Universal mandatory PDF is retired.
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`; at most one required derived rendition V1.
- Unsupported/unproven inline review representation falls back to a supported inspection path for the exact Submission; preview/viewer never becomes authority.

### Dossier / Evidence

- Dossier = small stable documentary context, never ERP/PLM/custom-object authority.
- DossierType remains small; no custom forms/fields/workflow/ACL/hierarchy/completeness engine V1.
- Dossier↔Document is M:N over stable Document identity; links never grant access.
- CAPTURED Evidence has exactly one immutable primary Dossier and reuses its scope.
- Cross-scope queries/projections/exports reapply canonical AuthZ; contextual links are never transitive grants.

### Retention / Legal Hold / Erasure

- No generic Record declaration entity.
- CAPTURED Evidence and first-submitted DocumentRevision become retention subjects automatically.
- Policies are explicit and snapshotted; expiry = disposition eligibility, never automatic deletion.
- Current EFFECTIVE Revision is never disposition-eligible.
- Physical disposal requires authorized disposition, no active hold and verified removal before DispositionRecord completion.
- LegalHold scopes: Evidence, stable Document, Dossier.
- Active Document/Dossier holds materialize current and newly entering retention subjects while within their live scope; unlink/lifecycle cannot release already-held subjects.
- Hold covers confirmed governed records, not never-submitted DRAFT/ESI; eDiscovery is future-triggered.
- Tenant erasure remains blocked while retention/hold obligations survive; required decryption capability remains until preserved content may lawfully be destroyed.

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
- Launch content safety = supported-format allowlist, size/type coherence and non-execution/download-safe behavior.
- Malware/quarantine/CDR, PKI/TSA/HSM, signed export packages, macro-enabled Office, custom renderer sandbox and eDiscovery remain explicit future triggers.

## 6. Whole-product freeze status

R9.5 is complete and frozen:

```text
R9.5-1 Content Model                         LOCKED
R9.5-2 Storage / Repository Strategy         LOCKED
R9.5-3 Authoring / EigenPal                  LOCKED (R9.5-8 refinement)
R9.5-4 Dossier / Context                     LOCKED
R9.5-5 Retention / Records / Legal Hold      LOCKED (R9.5-8 refinement)
R9.5-6 Import / Migration / Export           LOCKED
R9.5-7 Launch Attestation / Content Safety   LOCKED
R9.5-8 Whole-Product Adversarial Freeze      CLOSED / APPROVED
R9.5                                         FROZEN
```

The independent review attacked all 15 mandatory end-to-end cases, the 16-permission delta and the final subtractive/YAGNI pass. Reopen set after disposition: **EMPTY**.

Do not reopen frozen decisions for preference or hypothetical futures. Reopen only on material evidence under the DevelopmentConexus Engineering Method and only the minimal implicated decision set.

## 7. Current module disposition

Current module docs are current-state evidence, not target architecture.

High-level target disposition remains:

- `approval` → Approval V1;
- `auth` → retain V1 behind AuthN boundary;
- `iam` → conceptually split Organization + Authorization;
- `taxonomy` → dismantle: Area → Organization; legitimate type/classification meaning → Controlled Information;
- `controlleddocuments` → absorb legitimate identity/numbering responsibilities into Controlled Information;
- `documents` → Controlled Information core after responsibility cleanup;
- `templates` → retire parallel lifecycle; template becomes revision role/designation;
- `jobs` → orchestration/composition, not bounded context;
- `audit`, `render`, `search`, `notifications`, `distribution`, `tokens`, `security` → supporting concerns/mechanisms, retained only where they own distinct semantics.

R10 owns the final bounded-context/package/table ownership and deletion/rename map.

## 8. Build-vs-buy posture

Do not choose infrastructure before the frozen domain proves the requirement.

Current rulings:

- Keycloak/OIDC: defer until enterprise identity requirement fires.
- OpenFGA/SpiceDB: defer until arbitrary relationship-sharing graph justifies it.
- Camunda/Flowable/BPMN: not for Approval V1.
- Temporal: not an Approval prerequisite.
- CEL/expression language: defer until typed configuration cannot represent a real policy requirement.
- External ECM/JCR kernel: not the MetalDocs domain kernel.

External libraries/frameworks may be evaluated in R10+ only for precise mechanisms after ownership/semantics are fixed.

## 9. Documentation lifecycle

`wiki/` holds durable maintained product/repository truth; `docs/` holds active staging/working evidence unless an owner explicitly says otherwise. Git history is the archive.

The active ledger remains the binding detailed redesign decision record while R10 is open. R9.5 review packets remain evidence of how the freeze was challenged and dispositioned; they do not become parallel authority.

## 10. Implementation gate

**Closed.** No product implementation yet.

R10 and subsequent technical-design work must first complete bounded contexts/ownership, data model/constraints, transactions/events, API/frontend journeys, migration/delete map, durable target specs/ADRs and the required adversarial/operator gates. Only then may an implementation plan be authored from the accepted target.

## 11. Exact next step — R10 Technical Architecture

Open **R10** in design-only mode.

Start with an integrated technical decomposition before microdecisions. R10 must derive, not redefine, the frozen R3–R9.5 model and disposition at minimum:

```text
bounded contexts / module owners / dependency DAG
filesystem/package ownership + legacy deletion/rename map
target data model + table ownership
DB invariant constraints
transaction boundaries
durable events / outbox / async ownership
Artifact staging/storage/relocation/restore mechanics
WorkingContent OCC + coherent Submission atomicity
Release idempotency / exactly-one-EFFECTIVE enforcement
Retention / LegalHold / disposition mechanics
tenant erasure / restore reconciliation
canonical AuthZ on query/search/export surfaces
Historical Migration transaction/idempotency contracts
external publish/job effect truth
API contracts
frontend journeys
final migration/delete map
```

Technology/topology choices are subordinate to those ownership and invariant decisions. Current runtime/code/schema/OpenAPI are evidence, never target entitlement.