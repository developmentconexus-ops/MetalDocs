# Cohesive Platform Redesign — Active Architecture Authority

> **Status:** Active design authority — **R9.5 FROZEN; R10-A/B1/B2 PROMOTED; R10-B3/B4/B5/B6 ACCEPTED FOR R10 INTEGRATION / NON-FINAL; R10-B INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL; R10-C NEXT; NO PRODUCT IMPLEMENTATION AUTHORIZED**  
> **Established:** 2026-08-14  
> **R9.5 freeze ratified:** 2026-08-17  
> **R10-A/B1/B2 promoted:** 2026-08-17  
> **R10-B3 integration acceptance:** 2026-08-17 — non-final / not independently ratified  
> **R10-B4 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **R10-B5 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **R10-B6 integration acceptance:** 2026-08-18 — non-final / not independently ratified  
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Frozen product/domain ledger:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`  
> **Promoted R10 technical authority through B2:** `wiki/architecture/r10-technical-architecture.md`

Accepted current R10 integration inputs:

- `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md`
- `docs/superpowers/analysis/2026-08-18-r10-b5-documentary-context-records-governance-artifact-closure-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b5-integration-acceptance.md`
- `docs/superpowers/analysis/2026-08-18-r10-b6-audit-interchange-cross-owner-atomicity-integrated-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-b6-integration-acceptance.md`

---

## 1. Purpose / north star

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential/upstream identity-provider truth may be provider-owned; physical storage, authoring/editor technology, viewers and upstream repositories are replaceable providers/connectors around the MetalDocs kernel.**

Target posture:

- smallest professional architecture preserving real invariants;
- one canonical authority per business/system fact;
- one company per V1 deployment; common code/build/migrations; no customer forks;
- commodity mechanism may be externalized without taking semantic authority;
- no speculative ECM/BPM/ReBAC/low-code/object/records/integration/event-sourcing platform;
- current implementation is evidence, never automatic target entitlement.

Fresh sessions follow `AGENTS.md` → Method → current handoff → this page → frozen ledger → promoted R10 authority → accepted non-final B3–B6 working inputs.

R3–R9.5 remains frozen historical/product-domain authority except where an explicitly recorded bounded reopen/refinement is operator-approved for current R10 integration. B3–B6 remain non-final and challengeable only by material later-stage counterexample.

B6 included the internal whole-R10-B coherence/adversarial challenge required to close the relational/transactional design block for continued integration. This does **not** replace Whole-R10 Global Coherence Review + cold independent review before final ratification.

---

## 2. Deployment / ownership / access posture

V1 deployment invariant:

> **One company per deployment. `Tenant` is the singleton company root and TenantScope target, not a database partition.**

Promoted B1 substrate:

```text
one PostgreSQL product DB / schema metaldocs
UUID PKs
ordinary typed FKs
cross-owner RESTRICT / NO ACTION
no universal tenant/company/deployment partition column
no Tenant/Area/role/Permission RLS policy engine
serving DB role non-owner + NOSUPERUSER
READ COMMITTED by default
same-local-commit frozen cross-owner invariants
no provider DB/object-store atomicity dependency
```

Business bounded contexts remain exactly 8:

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

Supporting semantic owners remain exactly 3:

```text
Artifact
Audit
Interchange
```

Notifications are attributed support; Search is rebuildable projection; composition/platform/workers/providers own mechanisms only.

Authentication uses Keycloak as V1 AuthN provider while MetalDocs owns provider-subject binding, local Session and bounded assurance consumption. Provider roles/groups/orgs/permissions never become canonical AuthZ.

Authorization remains live additive/default-deny from static Role/Permission catalogs + current RoleAssignment + scope + owning-domain relationship/governance predicate. `tenant_owner` is a grant bundle, never bypass.

---

## 3. R10-B3 accepted working target

```text
Document
→ business DocumentRevision
→ WorkingContent + monotonic working_version OCC
→ immutable RevisionSubmission + deterministic governed manifest/digest
→ exact provider-neutral Artifact
```

Key laws:

- Document identity ≠ Revision ≠ autosave ≠ exact bytes;
- WorkingContent is sole mutable DRAFT authority;
- SUBMIT freezes exactly one accepted generation and consumes the OCC generation;
- same-REV return/resubmit never mutates old Submission;
- Artifact is exact-byte identity, not provider location and not global content-hash business identity;
- Template reuses Document lifecycle; no parallel TemplateVersion aggregate;
- one-open/one-EFFECTIVE have structural backstops;
- downstream native governance binds exact immutable Submission.

### B5-approved bounded B3 refinements

1. `DocumentOrigin` does not retain source Submission/Artifact forever through a strong FK; it preserves permanent source Revision identity plus exact source provenance snapshots.
2. `DocumentRevision.cancelled_at` / `obsoleted_at` are canonical native lifecycle instants; native supersession time remains B4 ReleaseRecord authority and is not duplicated.
3. `RevisionDictionarySnapshot` is separated from permanent Revision identity so lawful disposition may remove retained payload without deleting/reusing the revision ordinal or mutating immutable JSON in place.

### B6-approved imported-history refinements

4. `RevisionOrdinalReservation` preserves a historically real ordinal when insufficient exact bytes prevent creation of a truthful DocumentRevision.
5. `RevisionImportedContent` is CI-owned immutable exact imported Revision content and never a fake native RevisionSubmission.
6. `DocumentRevision.history_kind = NATIVE|IMPORTED`; imported governance/effectivity/actor proof is target-owner imported governance state, not synthetic ApprovalDecision/ReleaseRecord/native User action.
7. Template origin supports `NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT` source kind through immutable source identity/digest/hash snapshots.
8. native terminal timestamps remain native-only; already-terminal imported history may keep native timestamp NULL while imported governance preserves truthful historical timestamp/unknown.
9. current Tenant Dictionary is never resolved merely to fabricate imported historical dictionary state.

---

## 4. R10-B4 accepted working target

```text
RevisionSubmission
→ immutable SubmissionApprovalRequirement / ReleasePlan snapshots
→ one sequential Approval/Governance Step model
→ detached SubmissionFeedback
→ immutable Rendition when produced
→ automatic system-owned Release
→ concrete DistributionObligation snapshot in winning Release transaction
→ explicit immutable AcknowledgementRecord
```

Key laws:

- one Step semantic type; historical `review|approval` discriminator is removed in current R10 working target;
- Step carries label, ordered position, `NamedUser|Group|RoleInArea`, `ANY|ALL`, optional fresh-auth and due date;
- active participants may inspect exact Submission in-product, including exact PDF rendition when available, and may comment/annotate/suggest;
- feedback never mutates Submission; applying returned suggestion is later B3 OCC mutation;
- strict SoD has live/domain + structural backstops;
- required fresh-auth is Authentication-owned input and immutable Approval-owned consumption evidence;
- return-for-changes requires reason and returns same Revision to DRAFT without resetting B3 generation;
- Approval requirement/PolicyVersion + representation requirement are snapshotted per Submission;
- viewer capability is independent from official representation policy;
- Release is automatic/system-owned and sole native effectivity transition;
- Release atomically establishes EFFECTIVE/SUPERSEDED and concrete Distribution obligations;
- Distribution never grants access; explicit acknowledgement is the only obligation completion signal.

### Explicit bounded R9.5 Approval refinement

Frozen historical ledger contains `Step.purpose = review | approval`. Operator-approved R10 overlay removes it because external evidence + Structural Inversion Test showed it encoded legacy editor/UI ceremony rather than an invariant. Collaboration and fresh-auth remain orthogonal capabilities; `PeriodicReview` remains the distinct CI review concept.

---

## 5. R10-B5 accepted working target

### Documentary Context

- Dossier is stable documentary context only; never physical folder/content owner/access grant/retention authority/ERP-PLM master.
- DossierType remains small with explicit eligible DocumentTypes/EvidenceTypes; no custom fields/forms/workflow/ACL/completeness engine.
- Dossier↔Document is M:N over stable Document identity, copies no content and changes no lifecycle/Area/AuthZ.
- Dossier scope = exactly one TenantScope|AreaScope; stable type/key/scope; title mutable; archive reversible navigation only.
- External source identity/provenance remains explicit; no heuristic source takeover/merge.

### Evidence

- lifecycle = `DRAFT → CAPTURED → VOIDED`; VOIDED means invalid MetalDocs capture only;
- CAPTURE freezes immutable payload/metadata + exactly one primary Artifact and creates RetentionBinding;
- every CAPTURED Evidence has exactly one immutable primary Dossier and may have secondary Dossier context;
- Evidence reuses primary-Dossier scope and does not use REV/Approval/Release by default;
- canonical name freezes at CAPTURE; user filename remains provenance;
- current `{SEQ}` candidate uses one monotonic EvidenceType series; Dossier-local reset is a reopen trigger.

### B6-approved imported Evidence refinement

- `Evidence.history_kind = NATIVE|IMPORTED`;
- historical Evidence uses immutable target-owner imported capture/provenance state rather than fake native captured-by/captured-at;
- imported Evidence receives RetentionBinding inside the migration semantic unit; unknown historical anchor never silently becomes disposition-eligible.

### Records Governance

- no generic Record aggregate/declaration operation;
- first native DocumentRevision Submission / Evidence CAPTURE automatically create immutable typed RetentionBinding; imported target subjects create Binding in their migration semantic unit;
- policy = explicit `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`;
- anchor derives from canonical owner lifecycle/imported-history facts; Audit never becomes retention-clock authority;
- RetentionExtension only lengthens and cannot be added after DispositionFence;
- LegalHold scopes V1 = Evidence | stable Document | Dossier and materialize exact RetentionBindings, including newly entering live scope;
- Hold activation is fail-closed/all-or-nothing for current preservable scope when a fence/inconsistent subject is encountered;
- expiry only means eligibility; never automatic delete;
- DispositionFence is semantic irreversible authorization barrier; worker/retry state belongs R10-D;
- DispositionRecord means verified physical + semantic completion, not merely requested deletion;
- business lifecycle and records disposition are orthogonal axes.

### Artifact closure

- every confirmed Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence;
- multiple references within that root are allowed; cross-root Artifact-row reuse is rejected;
- identical bytes may have separate Artifact semantic rows across roots; physical dedupe remains mechanism freedom;
- no Artifact retention policy, generic owner registry or ref-count authority;
- semantic Artifact survival/deletion is derived from typed reachability.

### Retained DocumentRevision unit

The retained Revision unit includes subordinate immutable governance history: B3 Submission/dictionary/provenance/PeriodicReview state and B4 Approval/fresh-auth/feedback/Rendition/Release/Distribution/Acknowledgement evidence. Independent authorities such as User, Group, DocumentType and ApprovalPolicy are not recursively swallowed into the retention unit merely because referenced.

---

## 6. R10-B6 accepted working target

### Audit

```text
business/system owner mutation
→ required durable async intent(s) when needed
→ AuditChainHead lock
→ required AuditEvent append(s)
→ one COMMIT / ROLLBACK
```

- AuditEvent is transversal forensic/timeline evidence, never canonical business state;
- required governed mutations cannot report success without same-local-commit Audit;
- one deployment-wide append chain reuses RFC 8785 JCS + SHA-256;
- ordinary serving role has no direct mutation right on AuditEvent/AuditChainHead and uses only a narrow Audit-owned DB append primitive within the caller transaction;
- `AuditChainHead` is always the final contended semantic lock; no domain lock follows it;
- facts schemas are bounded product-owned contracts, not arbitrary JSON dumps;
- immutable skeleton contains stable UUIDs/codes/times/digests/outcomes only; names/emails/profile/JWT/secrets/request bodies/free-form reasons remain outside indefinite Audit;
- User UUID may survive as a PII-minimized pseudonymous identity skeleton while UserProfile remains separately erasable/read-resolved;
- Audit skeleton retention V1 = `Indefinite`; finite pruning/checkpoint/external crypto anchoring are reopen triggers;
- Audit is not outbox, event sourcing, telemetry, Approval/effectivity/AuthZ/retention authority.

### Interchange

Distinct contracts remain:

```text
Historical Migration
Governed Subject Export
External Repository IMPORT_COPY
External Repository PUBLISH_COPY
```

Backup/Restore remains separate; Tenant Portability Export remains deferred.

Historical Migration:

- never fabricates native Submission/Approval/Release/User facts;
- unknown source history remains unknown;
- source actors/timestamps/labels remain imported provenance;
- plan + true dry-run/apply + deterministic outcomes + reconciliation;
- APPLY atomic per explicit semantic import unit; partial batch success allowed/reconciled;
- new historical Document identity commits with at least one exact imported Revision/content unit or rolls back;
- long-lived idempotency uses stable SHA-256 source-identity digest instead of requiring raw human-readable source IDs.

Governed Subject Export:

- `COMPLETE` has explicit Document/Evidence/Dossier closure semantics and fails closed when required linked content is unauthorized;
- short `REPEATABLE READ` snapshot creates immutable provider-independent manifest facts; package assembly runs after commit;
- manifest uses exact object/relationship/provenance/file/hash facts + JCS/SHA-256;
- generated package is temporary delivery output, not automatically Artifact/Evidence/retention subject;
- V1 has no masquerading partial-export contract and no mandatory signing/BagIt/PKI.

External Repository:

- `InterchangeConnection` is stable logical repository identity; credentials/endpoints are provider/deployment mechanism;
- PUBLISH_COPY pins exact source identity/hash/format, and success exists only through immutable external receipt after provider confirmation;
- IMPORT_COPY uses ordinary target-owner lifecycle/permissions and commits target semantic creation + immutable receipt together;
- stable provider object identity is preferred over transient URL/path/display name.

### Cross-owner atomicity / lock law

- composition opens one local PostgreSQL transaction and calls transactionally composable owner seams;
- owners do not import each other's repositories or hide nested commits;
- provider/network/object-store effects never join the DB transaction;
- required durable effect intents are inserted in the same transaction and execute later in R10-D;
- default isolation remains READ COMMITTED; Governed Subject Export snapshot is the bounded REPEATABLE READ exception;
- lock acquisition follows one compatible forward partial order, with deterministic ordering inside equivalent sets;
- before implementation the whole B1→B6 admitted-write lock graph must be mechanically demonstrated acyclic.

---

## 7. R10-B integrated design-block closure

With operator acceptance of B6:

```text
R10-B1 = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2 = CLOSED / APPROVED / INTEGRATED
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B6 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL
```

Internal B1↔B6 coherence/adversarial review is closed sufficiently for continued R10 integration. Whole-R10 integration + Global Coherence Review + cold independent review still occur after R10-C–F and before final R10 ratification.

---

## 8. Storage / privacy / migration posture carried forward

- ManagedArtifactStore provider-neutral boundary remains first-class; production malware inspection remains mandatory/fail-closed before untrusted bytes become confirmed Artifact;
- provider WORM/ObjectLock/Purview may enforce physical retention but never becomes MetalDocs Records-Governance authority;
- no generic privacy workflow or mandatory crypto-erasure platform without named immutable Target Data;
- B6 accepted PII-minimized Audit skeleton posture; R10-C must prove physical restore/non-resurrection behavior where applicable;
- Historical Migration/Governed Export/IMPORT_COPY/PUBLISH_COPY business meaning is closed non-final; R10-F later owns legacy cutover/execution/deletion mapping.

---

## 9. Implementation gate

**CLOSED.** B3–B6 are accepted only for continued R10 integration, not final ratification.

Before implementation:

```text
R10-C
→ R10-D
→ R10-E
→ R10-F
→ Whole-R10 integration
→ Global Coherence Review
→ cold independent review
→ operator adjudication
→ final R10 ratification
→ implementation specification/plan
→ code
```

---

## 10. Exact next step — R10-C Artifact / Records Physical Integrity

Open **R10-C** as one integrated research-heavy design batch from promoted B1/B2 + accepted non-final B3–B6.

R10-C owns physical integrity and provider-mechanism semantics. It must not move Artifact/Controlled Information/Records Governance authority into storage-provider configuration.

At minimum jointly close:

```text
ManagedArtifactStore
  provider-neutral physical-location/reference representation
  conformance contract
  Local dev/test + reference production adapter posture

staging / confirmation
  temporary staging identity/lifecycle
  production malware inspection / fail-closed gate
  hash/size/format verification before semantic confirmation
  no confirmed orphan Artifact

physical integrity
  byte verification on read where required
  relocation/copy verification
  restore verification
  provider version/object identity treatment
  relocation never changes Artifact business identity or Submission digest

Records physical enforcement
  Object Lock/WORM interaction as enforcement only
  LegalHold/Retention/Disposition remain Records Governance authority
  physical deletion after DispositionFence
  provider-neutral deletion verification before DispositionRecord completion

recovery / cleanup
  failed staging/render/export physical cleanup
  orphan mechanism-state reconciliation
  backup/restore integrity
  privacy non-resurrection constraints where restore can reintroduce erased human-readable data
```

Explicitly challenge:

```text
B5 one-semantic-retention-root Artifact law vs physical dedupe
DispositionFence vs ObjectLock/provider deletion races
storage relocation vs exact Submission/Rendition identity
malware result vs semantic Artifact confirmation
backup/restore vs legal disposition/privacy erasure
provider outage/retry without business-truth rewrite
```

Route later work correctly:

```text
async retry/lease/jobs/reconciliation execution → R10-D
API/frontend/viewer/download journeys           → R10-E
historical cutover/legacy deletion/bootstrap    → R10-F
```

Implementation remains **BLOCKED**.