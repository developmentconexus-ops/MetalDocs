# MetalDocs Global Coherence Review — Minimal Reopen / Independent Fable Review Request

> **Status:** CANDIDATE / INDEPENDENT REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline HEAD:** `9f2f0a4ca2e390e67a2351cfd6ccaa578f5d690d`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** whole-platform Global Coherence Review after R10-B1 promotion and before continuing R10-B2
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this artifact is review evidence only. It does not amend R9.5, R10-A, R10-B1, open implementation, or authorize any schema/code/API/frontend change.

---

## 0. Cold reviewer bootstrap

Reconstruct state from the repository only. Do not use prior conversation memory as authority.

Start at `AGENTS.md` and follow its complete read order / authority chain. At minimum read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. this review packet
8. current code/schema/runtime only as claim-specific evidence
9. primary external sources only where a material technology/external-change claim must be verified

Current implementation is evidence, never target entitlement.

Apply the Method directly:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Apply Structural Inversion aggressively: if the legacy system had used the opposite provider, module layout, credential store, storage system or crypto mechanism, would the candidate conclusion still follow from the real product constraints?

A finding is evidence, not requirement authority. Do not create a new bounded context, platform, compliance requirement, framework or provider merely because it would be convenient.

---

# 1. Verified stage baseline

At authority baseline `9f2f0a4c`:

```text
R3–R9   = LOCKED
R9.5    = FROZEN

R10-A   = CLOSED / APPROVED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED
R10-B2  = NEXT / DESIGN ONLY
R10-B3..B6 = NOT STARTED
R10-C..F   = NOT STARTED

implementation = BLOCKED
```

R10-A topology and R10-B1 relational substrate are authoritative unless this review produces material evidence against a specific promoted assumption.

The purpose of this GCR is **not** to redesign the platform again. It is to determine whether recent evidence exposes bounded local maxima before B2–B6 make them expensive to unwind.

---

# 2. Global invariant under review

The platform should satisfy this architecture law:

> **MetalDocs owns differentiated product/business authority and history; mature commodity mechanisms should be reused when doing so reduces total complexity, but an external mechanism/provider must never become a second authority for MetalDocs domain meaning.**

This specializes, but does not replace, the existing Method rule `mechanism != authority`.

The reviewer must attack both directions:

- **under-buying / accidental rebuild:** MetalDocs implements commodity infrastructure that a mature standard/provider can own more safely and cheaply;
- **over-buying / authority leakage:** a framework/provider is introduced for a problem smaller than the framework or starts owning MetalDocs business semantics.

---

# 3. Candidate GCR-R1 — Authentication build-vs-buy

## 3.1 Frozen/current premise being challenged

R9/R9.5 currently freezes local V1 credentials, activation, session, lockout/revocation and fresh-auth under Authentication, with Keycloak/external IdP deferred until enterprise identity/MFA/federation triggers.

R10-A promotes Authentication as the semantic owner of credential/session identity facts and retains external IdP as a provider seam.

The implementation also contains local password hashes, password algorithms, lockout counters and password-reset flows. Those are current-state evidence, not target authority.

## 3.2 New material evidence

The operator now explicitly requires starting with a dedicated IdP and has prior operational experience running Keycloak successfully. This is a changed real requirement/constraint, not a hypothetical future.

Current Keycloak official documentation shows that it already provides the exact commodity surface the local-auth target would otherwise have to build and maintain, including:

- OIDC Authorization Code Flow for browser-based apps;
- identity brokering / upstream OIDC and SAML;
- LDAP/AD federation;
- MFA / OTP / WebAuthn/passkeys;
- session management;
- step-up authentication / ACR/LoA flows;
- Organizations and organization-specific identity-provider routing.

Primary source: `https://www.keycloak.org/docs/latest/server_admin/`.

OpenID Connect Core states that `iss + sub` is the stable end-user identifier an RP can rely on; email/username are not stable identity keys.

Primary source: `https://openid.net/specs/openid-connect-core-1_0-18.html`, §5.7.

The OAuth Browser-Based Apps work describes the BFF pattern in which the backend is the confidential OAuth client, keeps OAuth tokens server-side and exposes a cookie-based application session to the browser.

Primary source: `https://datatracker.ietf.org/doc/draft-ietf-oauth-browser-based-apps/26/`.

## 3.3 Root cause

The earlier build-vs-buy decision treated credential authentication as part of the product kernel because it was simple enough to implement locally. That is a local maximum if the product can instead depend on a mature standards-based IdP while retaining all business/organizational identity and authorization authority inside MetalDocs.

## 3.4 Target invariant

```text
Authentication mechanism may be replaced
without changing:
  Tenant
  User
  Area
  Group
  Role
  RoleAssignment
  Document access semantics
  Approval participant semantics
  governed history
```

And:

```text
Keycloak credential/IdP identity
!= MetalDocs organizational identity
!= MetalDocs Authorization authority
```

## 3.5 Candidate outcome

```text
GCR-R1 = RESTRUCTURE NOW

V1 authentication provider = Keycloak

Keycloak owns:
  credentials/password policy
  MFA/passkeys
  password recovery
  upstream OIDC/SAML/LDAP/AD authentication
  IdP authentication session
  authentication flows

MetalDocs Authentication owns:
  provider subject binding
  application authenticated-context session
  trusted authentication assurance/fresh-auth facts
  provider anti-corruption contract

MetalDocs Organization owns:
  Tenant/User/Area/Group/GroupMembership

MetalDocs Authorization owns:
  Permission/Role/RoleAssignment/canonical evaluation
```

Candidate stable binding:

```text
issuer + subject → MetalDocs Authentication binding → Organization.User
```

Email, username and display name remain claims/profile data, never provider-independent identity.

## 3.6 Candidate topology

```text
Browser
  ↓
MetalDocs BFF / application backend
  ↓ Authorization Code
Keycloak
  ↓ callback
MetalDocs Authentication
  ↓ stable iss/sub binding
Organization.User
  ↓
opaque MetalDocs application Session
```

Candidate deployment posture:

```text
one Keycloak realm per environment / application trust domain
NOT one realm per MetalDocs Tenant
```

Keycloak Organizations, if later used, are an AuthN routing/federation projection of a MetalDocs Tenant. They are never Tenant/Group/Role/Authorization authority.

Keycloak roles/groups/Authorization Services are not canonical MetalDocs authorization sources in V1.

## 3.7 What GCR-R1 must NOT pull forward

Do not decide final login endpoints/DTOs/screens here; R10-E owns final access/frontend journeys.

Do not design Keycloak HA topology, realm import automation, production sizing or operations beyond seams required to validate viability; implementation/ops planning comes later.

Do not create SCIM/JIT enterprise provisioning unless a real consumer requires it.

---

# 4. Candidate GCR-R2 — Managed Artifact Store provider entitlement

## 4.1 Frozen premise being challenged

R9.5 currently freezes first-class Managed Artifact Store adapters as:

```text
Local(dev/test)
MinIO
AWS S3
```

with other S3-compatible products behind conformance validation.

## 4.2 New external evidence

The official `minio/minio` repository is archived/read-only and its README now states that the repository is no longer maintained, directs users to AIStor editions and describes the community distribution as source-only.

Primary evidence:

- GitHub repository state: `https://github.com/minio/minio` — archived
- official README: `https://github.com/minio/minio/blob/master/README.md`

This is an external change under the Method reopen rule. It does not invalidate S3-compatible object storage as an architecture.

## 4.3 Root cause

The frozen target named a replaceable mechanism as a first-class product entitlement. Replacing MinIO with another named self-hosted object store now would repeat the same coupling without a deployment consumer.

## 4.4 Candidate outcome

```text
GCR-R2 = RESTRUCTURE TARGET PROVIDER SET

Local
  = first-class dev/test provider

AWS S3
  = reference production ManagedArtifactStore profile

MinIO OSS
  = remove nominal first-class target entitlement

self-hosted S3-compatible storage
  = select only when a real deployment requires it
  = must pass the ManagedArtifactStore conformance suite
```

Do not select Garage, Ceph, AIStor or another self-hosted provider merely to fill a slot.

The frozen domain/storage semantics remain:

```text
Artifact identity != provider identity
canonical hash = MetalDocs SHA-256
provider URL/key/version/ETag != business identity
one active Managed Artifact Store per deployment V1
copy + verify + cutover for relocation
no permanent dual-write V1
```

R10-C owns the concrete conformance contract and physical integrity proof.

---

# 5. Candidate GCR-R3 — basic malware inspection before Artifact confirmation

## 5.1 Frozen premise being challenged

R9.5-7 deliberately rejected a broad V1 security platform and deferred malware scanning/quarantine/CDR/rescanning together. The accepted V1 retained allowlisted formats, size/coherence validation and a `staging → validation → confirmation` seam.

## 5.2 Evidence

OWASP's File Upload Cheat Sheet recommends defense-in-depth for uploaded files: extension/type/signature/size validation, server-controlled naming/storage and malware inspection via antivirus/sandbox when available. It also highlights malicious files that target parsers/processors or client-side active content.

Primary source: `https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html`.

MetalDocs accepts business documents/evidence and later may parse, render, preview, distribute or export them. That makes pre-confirmation inspection a concrete ingress risk rather than a speculative compliance platform.

## 5.3 Root cause

The prior YAGNI decision grouped a small ingress safety control with a much larger security platform. Rejecting the whole bundle removed more than accidental complexity.

## 5.4 Candidate outcome

```text
GCR-R3 = BOUNDED RESTRUCTURE

untrusted bytes
→ STAGED
→ size / ContentFormat / structural validation
→ malware inspection
→ semantic attachment validation
→ CONFIRMED Artifact
```

Invariant:

> Bytes originating from an untrusted upload/import path do not become a CONFIRMED governed Artifact until the required malware inspection for the deployment profile has produced an acceptable result.

If inspection is unavailable/incomplete, the content remains non-confirmed; the operation fails visibly rather than silently weakening the gate.

Explicitly still out of V1 unless separately triggered:

```text
quarantine business aggregate/product
CDR platform
custom sandbox cluster
periodic rescans
malware intelligence platform
advanced ArtifactSecurityAssessment model
macro-enabled Office support
```

Do not select ClamAV or any scanner in this GCR. R10-C compares concrete scanning mechanisms and failure posture.

---

# 6. Candidate GCR-R4 — Tenant DEK / cryptographic-erasure prerequisite

## 6.1 Frozen premise being challenged

R8/R9.5 currently includes:

```text
Organization owns tenant key-custody lifecycle facts
terminal erasure destroys no-longer-needed Tenant DEK
DEK needed for retained intelligible content survives
```

R9.5 storage simultaneously states that the Tenant DEK does **not** encrypt every Artifact V1 and that production baseline uses provider encryption at rest.

R10-A therefore promotes Organization ownership of key-custody lifecycle facts while platform owns crypto/KEK primitives.

## 6.2 Evidence gap

Current target authority and current-state evidence reviewed so far do not identify a concrete V1 `Target Data` family whose confidentiality/integrity depends on a Tenant DEK in a way that makes destroying that key a meaningful cryptographic erase of that data.

NIST defines cryptographic erase as sanitizing the key(s) that actually provide confidentiality protection to encrypted **Target Data**, making recovery of the decrypted target data infeasible.

Primary source: `https://csrc.nist.gov/glossary/term/cryptographic_erase` and NIST SP 800-88 Rev. 2.

AWS KMS documentation similarly states that deleting a symmetric encryption key makes the remaining ciphertexts encrypted under that key unrecoverable.

Primary source: `https://docs.aws.amazon.com/kms/latest/developerguide/deleting-keys.html`.

Destroying a tenant-associated key that does not protect the substantive target data would be security ceremony and could create a false erasure claim.

## 6.3 Candidate outcome

```text
GCR-R4 = STOP / SPLIT PREREQUISITE — PROVE OR REMOVE
```

Before retaining mandatory Tenant DEK/key-custody lifecycle as a V1 target invariant, B2/R10-C must prove:

1. the exact Target Data encrypted under the Tenant DEK or a hierarchy rooted in it;
2. the encryption boundary and envelope/key relationship;
3. restore/recovery behavior and backup implications;
4. which retained data requires the key to survive;
5. which ciphertext becomes unrecoverable when the key is destroyed;
6. why verified logical/physical deletion plus provider encryption/storage controls is insufficient for the actual V1 requirement.

If no concrete V1 target data satisfies those conditions:

```text
REMOVE mandatory Tenant DEK lifecycle from V1
REMOVE cryptographic-shred claim from Tenant erasure
retain a future key-custody seam only when a real tenant-encrypted data profile appears
```

Do not replace the DEK abstraction with a generic KMS/Vault framework unless the data-protection requirement exists.

The frozen retention-aware erasure invariant itself remains: blocked data must not be destroyed while retention/hold requires preservation; eligible substantive product state must be erased and the terminal erasure record/tombstone must remain truthful.

---

# 7. Candidate clarifications

## GCR-C1 — refine the North Star meaning of identity

Current wording says MetalDocs is the system of record for identity, governance, revision, evidence and documentary context.

Candidate refinement:

> **MetalDocs is the system of record for product/organizational identity, governance, revision, evidence and documentary context. Authentication credential and upstream identity-provider truth may be owned by a dedicated Authentication provider and is bound into MetalDocs organizational identity through stable provider subject identity.**

This is intended to preserve the original North Star while preventing `identity` from being misread as an obligation to store passwords/passkeys locally.

## GCR-C2 — clarify R10-B1 database topology

B1's `one PostgreSQL database` is product-state topology, not a mandate that every provider product persist in the same logical database/schema.

Candidate wording:

```text
one MetalDocs product-state PostgreSQL database
```

Provider-owned products such as Keycloak retain separate:

```text
persistence ownership
migrations
credentials
schema/database lifecycle
```

A small deployment may share a PostgreSQL server/cluster operationally; that does not merge logical persistence authority.

---

# 8. Candidate decisions explicitly confirmed by this GCR

The reviewer must try to falsify these confirmations, but must not reopen them for preference or provider convenience.

| Area | Candidate disposition | Reopen only on |
|---|---|---|
| modular monolith / local ACID product DB | **CONFIRM** | real independent-deploy/scale/trust boundary or invariant impossible locally |
| R10-A 8 business BCs + 3 supporting semantic owners | **CONFIRM** | a real fact family cannot be coherently owned or duplicate/missing authority is proven |
| R10-B1 composite tenant identity/FK/RLS law | **CONFIRM** | material counterexample against a protected invariant |
| Organization and Authorization remain MetalDocs-owned | **CONFIRM** | real requirement proves provider must own meaning, not merely mechanism |
| 5 V1 roles + RoleAssignment Tenant/Area scope | **CONFIRM** | real per-object sharing/hierarchy/custom-role requirement |
| OpenFGA/SpiceDB | **DEFER CONFIRMED** | arbitrary resource sharing/hierarchy/reverse relationship graph becomes real |
| Controlled Information Document/Revision/WorkingContent/Submission kernel | **CONFIRM** | identity/history invariant counterexample |
| WorkingContent OCC + immutable Submission | **CONFIRM** | concurrency/product requirement that cannot be expressed by current seam |
| EigenPal/editor as provider around WorkingContent | **CONFIRM** | provider cannot preserve frozen content/review invariants |
| realtime collaboration | **DEFER CONFIRMED** | concrete concurrent coauthoring requirement |
| specialized Approval, no generic BPM | **CONFIRM** | real branching/ad-hoc/business-designed process requirement |
| Camunda/Flowable/BPMN | **DEFER CONFIRMED** | Approval/workflow semantics materially exceed the specialized model |
| outbox/durable-intent law | **CONFIRM** | external-effect correctness cannot be preserved by same-commit intent seam |
| Temporal | **DEFER with explicit R10-D trigger** | repeated durable timer/retry/resume/compensation state machines exceed simpler machinery |
| Search as rebuildable projection | **CONFIRM** | real query/scale/ranking evidence requires external engine |
| Notifications as attributed delivery/read support | **CONFIRM** | real independent business lifecycle emerges |
| Artifact identity/hash provider separation | **CONFIRM** | provider semantics become unavoidable identity invariant, not convenience |
| provider WORM/Object Lock as enforcement, not Records authority | **CONFIRM** | product gives record authority to provider explicitly |
| Documentary Context / Evidence / Dossier small model | **CONFIRM** | real ERP/PLM/object-model requirement invalidates scope |
| Records/Retention/Hold/Disposition ownership | **CONFIRM** | material retention/hold counterexample |
| Historical Migration without fabricated native history | **CONFIRM** | trustworthy portability contract proves a different native-preservation path |
| SharePoint Embedded future profile | **DEFER CONFIRMED** | named Microsoft-enterprise consumer requires it |
| PKI/qualified e-signature | **DEFER CONFIRMED** | legal/customer signature requirement |
| eDiscovery/ESI preservation | **DEFER CONFIRMED** | concrete legal discovery requirement |

---

# 9. Mandatory cold adversarial questions

The reviewer must answer all material questions and may add better counterexamples.

## Authentication / Keycloak

1. Does Keycloak materially reduce total complexity, or merely move it into deployment/adapter complexity?
2. Is Keycloak a stable enough V1 dependency under the real MetalDocs deployment constraints?
3. Is `iss + sub` the correct provider-independent binding identity?
4. Is an opaque MetalDocs application Session still justified after Keycloak, or is it duplicate session machinery?
5. Is BFF/cookie session the right browser architecture or is a simpler safe flow superior?
6. Is one realm per environment/application trust domain sound for MetalDocs multi-tenancy?
7. Does any real Tenant-specific IdP requirement force realm-per-Tenant, or can realm Organizations/identity routing remain a projection?
8. Can Keycloak Organizations/groups/roles be kept strictly mechanism/projection, or will operational pressure create a second Organization/AuthZ authority?
9. Does adopting Keycloak require reopening R10-A ownership, or only the R9/R9.5 build-vs-buy decision?

## Artifact storage

10. Does removing MinIO OSS as a named provider create a production/self-hosted gap that must be filled now?
11. Is AWS S3 a reasonable reference production profile without becoming product authority?
12. Should a self-hosted provider be selected now, or is a conformance seam genuinely sufficient?
13. What minimum behavior must the S3-compatible conformance suite prove to preserve Artifact invariants?

## Malware inspection

14. Is pre-confirmation malware inspection a real V1 safety requirement for MetalDocs, or an overreaction relative to the actual supported formats/processors?
15. Does fail-closed inspection make upload/authoring operationally fragile without enough safety gain?
16. Can the requirement be implemented without inventing a quarantine/security-domain lifecycle?
17. Is `STAGED → validation → scan → CONFIRMED` the correct ownership/mechanism split?

## Tenant DEK / crypto erase

18. What exact V1 Target Data is protected by the Tenant DEK today or in the accepted target?
19. If no such Target Data exists, is mandatory Tenant DEK pure accidental complexity / false-security semantics?
20. Does removing mandatory Tenant DEK weaken a real tenant-erasure property that cannot be achieved by verified logical/physical deletion?
21. If a DEK is actually justified, what is the smallest concrete envelope-encryption/key-custody model and which owner/mechanism split preserves it?
22. Is there a real requirement for cryptographic erase, as distinct from truthful erasure with retention-aware verified deletion?

## Whole-platform challenge

23. Is there a **fifth material local maximum** anywhere in R3–R10-B1?
24. Is MetalDocs implementing another mature commodity capability that should be bought/reused instead?
25. Is any current/proposed external dependency acquiring product/business authority it should not have?
26. Is any internal bounded context merely a wrapper around a provider and therefore not a real semantic owner?
27. Is any confirmed framework/abstraction present only because another abstraction exists?
28. Did YAGNI remove any essential invariant/safety property beyond the malware case?
29. Did this GCR add any mechanism/platform whose long-term complexity exceeds the defect class it removes?
30. Which target concept, if deleted today, would reduce complexity without weakening a distinct material property?

---

# 10. Global Maximum / YAGNI proof requirements

For each R1–R4 candidate, provide:

```text
Evidence
Known / Inferred / Unknown / Deferred
Root Cause
Target Invariant
2–3 credible alternatives
Structural Inversion result
Local vs Global Maximum
Essential vs Accidental Complexity
YAGNI / Future Cost
Authority / Mechanism boundary
Proof strategy
Strongest adversarial counterexample
Decision
Reopen triggers
```

Do not accept a provider because it is popular. Do not reject a provider because it is external. Compare **total lifecycle complexity and authority correctness**.

For confirmed decisions, identify what evidence would actually have changed the verdict. If nothing could falsify the conclusion, the review is not adversarial enough.

---

# 11. Required reviewer verdict

Return exactly one top-level verdict:

```text
APPROVE GCR MINIMAL REOPEN SET
APPROVE GCR MINIMAL REOPEN SET WITH MATERIAL FIXES
DO NOT APPROVE GCR MINIMAL REOPEN SET
```

Then report:

1. BLOCKER / MAJOR / LOW findings;
2. R1–R4 disposition individually;
3. C1/C2 disposition;
4. any newly discovered fifth material local maximum;
5. R9.5 reopen set after review;
6. R10-A reopen set after review;
7. R10-B1 reopen set after review;
8. whether R10-B2 may safely resume after adjudication;
9. exact promotion conditions;
10. whether another full adversarial round is necessary or a bounded delta check is sufficient.

---

# 12. Review write authorization

The independent reviewer is authorized **only** to create, commit and push a new review artifact:

`docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-independent-fable-review.md`

Do not alter:

- R9.5 frozen authority;
- `wiki/architecture/r10-technical-architecture.md`;
- `wiki/references/current-agent-handoff.md`;
- this candidate packet;
- product code/schema/OpenAPI/frontend;
- R10-B2+ design authority;
- implementation.

Findings are evidence. The primary architect/operator adjudicates them under the DevelopmentConexus Engineering Method before any authority changes.