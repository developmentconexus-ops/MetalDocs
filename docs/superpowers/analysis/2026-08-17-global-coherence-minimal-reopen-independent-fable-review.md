# MetalDocs Global Coherence Review — Minimal Reopen / Independent Fable Review

> **Status:** INDEPENDENT COLD REVIEW — **EVIDENCE ONLY, NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed HEAD:** `7a9125e9b0a4415e5d96bb1097936e1d8979ffba` (authority baseline `9f2f0a4c` + review packet)
> **Review packet:** `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-fable-review-request.md`
> **Method:** DevelopmentConexus Engineering Method v1.0.0 (local mirror)
> **Reviewer posture:** cold session; repository truth only; no conversation memory used as authority
> **Implementation gate:** CLOSED — this review authorizes nothing.

Findings below are evidence. The primary architect/operator adjudicates them under the Method before any authority changes.

---

# 0. Review basis

Read order executed: `AGENTS.md` → Method mirror → `wiki/references/current-agent-handoff.md` → `wiki/architecture/cohesive-platform-redesign.md` → frozen R3–R9.5 ledger → `wiki/architecture/r10-technical-architecture.md` → review packet → claim-specific code/schema evidence → primary external sources.

## 0.1 External evidence verification (all verified 2026-08-17)

| Claim | Result |
|---|---|
| `minio/minio` archived | **VERIFIED** — archived by owner 2026-04-25, read-only; README: "THIS REPOSITORY IS NO LONGER MAINTAINED"; community edition source-only, no maintained binaries; users directed to AIStor Free/Enterprise |
| Keycloak commodity surface | **VERIFIED** — current docs (v26.x server admin guide) document OIDC auth code flow, identity brokering (OIDC/SAML), LDAP/AD federation, OTP/WebAuthn/passkeys, session management, ACR/LoA step-up, Organizations with org-specific IdP routing |
| OIDC Core §5.7 | **VERIFIED** — `iss` + `sub` together are "the only Claims that an RP can rely upon as a stable identifier for the End-User"; email/username carry no stability/uniqueness guarantee |
| OAuth Browser-Based Apps BFF | **VERIFIED** — draft-ietf-oauth-browser-based-apps-26 (2025-12-04, active); BFF "strongly recommended for business applications, sensitive applications, and applications that handle personal data"; tokens server-side, cookie session to browser |
| NIST cryptographic erase | **VERIFIED** — SP 800-88r2: sanitizing "one or more keys providing confidentiality protections for the encrypted target data". Destroying a key that protects no Target Data is not cryptographic erase |
| OWASP File Upload Cheat Sheet | **VERIFIED with nuance** — AV/sandbox scanning is recommended as defense-in-depth **"if available"**, not as a mandatory core control; core controls are extension/type/signature/size/naming/storage validation |

The OWASP nuance matters for GCR-R3 honesty and is handled in §5.

## 0.2 Claim-specific implementation evidence (current state only, never target entitlement)

| Fact | Anchor |
|---|---|
| Local credential stack is substantial but incomplete: Argon2id/bcrypt hashing (~306 LOC), lockout counters, admin-only password reset, opaque HMAC-stored cookie session; **no self-service reset, no invite/activation flow, MFA is a reporting-only stub** | `internal/platform/passwordhash/passwordhash.go`, `internal/modules/auth/application/service.go:316-1101`, `internal/modules/auth/infrastructure/postgres/repository.go:274-400`, `db/baseline/0001_current_schema.sql:1044,1067` |
| Password policy is length-only (default 8) | `internal/modules/auth/application/service.go:1162`, `internal/platform/authn/config.go:75` |
| Fresh-auth/reauth exists for approval e-signature | `internal/modules/approval/infrastructure/signature/password_reauth.go` |
| Tenant DEK encrypts **exactly one** data family: `metaldocs.audit_events.payload` (AES-256-GCM envelope) | `internal/modules/audit/infrastructure/postgres/writer.go:211-260`, `internal/platform/crypto/envelope.go`, `db/baseline/0001_current_schema.sql:1519-1526` |
| DEK machinery is **fail-open by configuration**: KEK env unset ⇒ crypto disabled ⇒ audit payload stored plaintext and tenant-erasure crypto-shred is a silently skipped no-op | `internal/platform/config/tenant_crypto.go:31-44`, `writer.go:227-231`, `internal/modules/iam/application/tenant_lifecycle_service.go:633-654` |
| Object-store bytes are **not** DEK-encrypted | claim-2 sweep, no ciphertext columns beyond `tenant_keys.wrapped_dek` |
| MinIO is the **only implemented** storage provider; `local`/`memory` are dead config enums with no adapter; dev/CI compose depends on MinIO server; `minio-go/v7` is a direct go.mod dependency | `internal/platform/objectstore/verified_store.go`, `internal/platform/config/attachments.go:11-15,58-61`, `internal/platform/bootstrap/api.go:96-107`, `go.mod:26`, `deploy/compose/docker-compose.yml` |
| Upload pipeline is presign → confirm → pointer with **hash + 25 MiB size check only**; no MIME/extension/magic-byte validation, no malware inspection anywhere | `verified_store.go:113-146`, `internal/modules/documents/application/service.go:670-770` |

---

# 1. Verdict

```text
APPROVE GCR MINIMAL REOPEN SET WITH MATERIAL FIXES
```

All four reopens survive adversarial falsification in direction and scope. None survives unmodified: each needs a bounded material fix recorded below before promotion. Both clarifications are accepted (C2 with one addition). No fifth material local maximum was found at the authority level; three near-misses are recorded and bounded into existing reopens/blocks.

---

# 2. Findings

## BLOCKER

None.

## MAJOR

| ID | Against | Finding |
|---|---|---|
| M1 | GCR-R1 / R10-A | Adopting Keycloak V1 **requires a bounded R10-A amendment**, not only an R9/R9.5 build-vs-buy reopen. The promoted Authentication owned-fact list (`credential/identity binding, activation, opaque session, lockout/revocation, fresh-auth`) is partially falsified: credential storage, password policy, activation and lockout become provider facts. Authentication must be re-scoped to: provider subject binding (`iss`+`sub` → Organization.User), application authenticated-context session, trusted assurance/fresh-auth facts, provider anti-corruption contract. Topology (8 BCs + 3 supporting owners) is unchanged. Packet Q9 is answered: yes, bounded R10-A amendment; no topology reopen. |
| M2 | GCR-R1 / C2 | Keycloak persistence is a separate authority (C2), therefore **no MetalDocs invariant may claim atomicity across the product DB and the provider DB**. User provisioning/binding is necessarily a non-atomic cross-system choreography. B2 must own the binding lifecycle explicitly: Keycloak-user-exists/no-binding, binding-exists/Keycloak-user-gone, duplicate-binding rejection (`UNIQUE(iss, sub)`), and reconciliation posture. Without this, the first implementation wave will invent it ad hoc. |
| M3 | GCR-R1 | The "Keycloak roles/groups/Organizations are never MetalDocs authority" rule must be **structural, not disciplinary**: B2's design must give MetalDocs Authorization no representation into which provider roles/groups/org claims could be consumed (no provider-role mapping table, no claim-to-permission bridge). The only facts crossing the anti-corruption contract are `iss`, `sub`, and enumerated authentication-assurance claims (`auth_time`, ACR/AMR). This is the enforcement layer for packet Q8 and must be a named B2 proof obligation. |
| M4 | GCR-R4 | The packet's evidence claim "no concrete V1 Target Data family identified" is **overstated against current-state evidence**: exactly one real family exists — `audit_events.payload` envelope-encrypted under the Tenant DEK, built to reconcile append-only/tamper-evident Audit with retention-aware tenant erasure (`writer.go:211-260`). STOP/SPLIT remains the correct outcome, but the prove-or-remove obligation must **explicitly adjudicate the immutable-audit-PII family (B6) and backup/restore implications (R10-C/R10-F)** before the REMOVE arm can be taken; otherwise removal silently reintroduces the audit⊥erasure conflict later. Additionally: the current mechanism is fail-open (KEK unset ⇒ plaintext + silently skipped crypto-shred), so whichever arm B2/B6 takes must be fail-closed under B1 law — a configuration-dependent erasure claim is exactly the false-security semantics the candidate warns about. If REMOVE is taken, the cascade must be executed completely: Organization "tenant key-custody lifecycle facts" family, the B2 scope line, and the platform crypto/KEK mechanism entry are all removed with it (an abstraction retained only because another abstraction existed). |
| M5 | GCR-R3 | As worded, the invariant is satisfiable by a vacuous profile ("required malware inspection for the deployment profile" = none ⇒ gate always passes). The invariant must be **secure by default at the deployment/platform level**: the production deployment profile requires inspection unless the operator explicitly opts out at deployment configuration; dev/test profiles may default off. Requirement authority is deployment configuration, **not** tenant-facing configuration — per-tenant scanning policy would create a tenant-owned security-authority surface V1 has no consumer for. |
| M6 | GCR-R2 | Restate the durable first-class surface: it is the **ManagedArtifactStore port + conformance contract**, not any provider name. AWS S3 = reference production profile (viability evidence); Local = first-class dev/test provider. This removes the residual coupling-class defect (naming providers as entitlements) that produced the MinIO exposure. Consequence for R10-C: the conformance suite needs a concrete execution mechanism (which S3-compatible endpoint exercises the S3 adapter in dev/CI) — note that today `local`/`memory` adapters do not exist (dead enums) and dev/CI run against a MinIO server, so a transition dependency on a frozen MinIO image (or a deliberate alternative) must be recorded as **bounded test/dev mechanism, never product entitlement**. |

## LOW

| ID | Against | Finding |
|---|---|---|
| L1 | GCR-R2 | `minio-go/v7` (direct dependency) is from the same abandoned-OSS vendor family as the server. R10-C should select the S3 client library deliberately (e.g., AWS SDK for Go v2) instead of inheriting it. |
| L2 | GCR-R3 | Pipeline order: structural/format validation itself parses untrusted bytes before the scanner runs. Scanner placement relative to parsing, and validator hardening posture, are R10-C decisions; sandboxing validators remains correctly deferred. |
| L3 | GCR-R1 | The buy case includes a frontend deletion not yet stated: Keycloak-hosted (themable) login/recovery/MFA-enrollment journeys replace custom MetalDocs credential screens. R10-E should record this deletion rather than rebuilding login UX against Keycloak APIs. |
| L4 | GCR-C2 | C2 wording should add one sentence: no MetalDocs invariant may depend on cross-database atomicity between the product-state database and any provider-owned database. (This is the C2 face of M2.) |
| L5 | evidence only | Current default storage config (`local`) silently wires no object store, and KEK-unset silently disables crypto — two fail-open configuration defaults. Out of GCR scope (implementation gate closed); recorded as evidence that B1's fail-closed law must reach configuration wiring at implementation time. |

---

# 3. GCR-R1 — Authentication build-vs-buy — **APPROVE (RESTRUCTURE NOW) with M1, M2, M3, L3**

## 3.1 Method chain

**Evidence.** Operator requirement recorded in the packet: start with a dedicated IdP; prior operational Keycloak experience. External: Keycloak commodity surface verified (§0.1); OIDC `iss+sub` stability verified; BFF verified as the IETF-recommended browser architecture for exactly this application class. Implementation: the local stack is ~900 LOC of credential-specific code **and still lacks** self-service reset, invite/activation, and MFA (§0.2) — reaching credible V1 parity means building more local credential machinery, not keeping what exists.

**Known / Inferred / Unknown / Deferred.** Known: operator requirement; Keycloak capability; local-stack gaps. Inferred: total lifecycle complexity of running Keycloak (JVM service + own DB + upgrade cadence) is bounded by operator's demonstrated operational competence. Unknown: none material at design level. Deferred: HA topology, realm import automation, sizing (packet §3.7, correctly).

**Root cause.** Build-vs-buy was decided when "local credentials are simple enough" was locally true; the requirement changed (operator mandates IdP) and the local path's real cost includes the unbuilt table-stakes features. Local maximum confirmed.

**Target invariant.** As stated in packet §3.4 — provider-replaceable authentication; `Keycloak identity ≠ MetalDocs organizational identity ≠ Authorization authority`. Sound and unchanged by this review.

**Credible alternatives compared.**
1. *Keep/complete local auth* — fails the changed operator requirement; requires building reset/invite/MFA/policy machinery Keycloak already owns; highest future cost for federation/SSO.
2. *Zitadel / Ory / Authentik* — no material superiority under real constraints: self-hostable, mature, standards-complete, and — decisive under total-lifecycle-complexity comparison — the operator has operational competence with Keycloak, not with these. Choosing an unfamiliar IdP to avoid "popularity bias" would itself be preference, not evidence.
3. *Hosted IdP (Auth0/Cognito/Entra)* — external service dependency for self-hostable deployments, data residency and cost coupling; conflicts with deployment reality without any offsetting authority benefit.

**Structural Inversion.** If the legacy system had used Keycloak from day one, no evidence in this review would justify migrating to local credentials. The conclusion follows from constraints, not from current implementation shape. Passes.

**Local vs Global Maximum.** Keycloak-as-provider is the global maximum for AuthN mechanism under a stated IdP requirement; local auth was a local maximum inside "no enterprise trigger yet" — and the trigger has now fired as an operator requirement change, which is legitimate reopen evidence under the Method (changed requirement, not preference).

**Essential vs accidental.** Essential complexity retained by MetalDocs: binding, app session, assurance facts, ACL contract. Accidental complexity deleted: password storage/policy/reset/lockout/MFA machinery + credential frontend journeys (L3). Accidental complexity added: one operated service + realm configuration. Net deletion is real because the deleted surface includes unbuilt-but-required features.

**YAGNI / future cost.** No SCIM/JIT provisioning, no HA/ops design now (packet §3.7 fences hold). MinIO's fate is the cautionary tale for single-vendor OSS; Keycloak differs materially: CNCF project, multi-vendor ecosystem, Red Hat commercial productization. Residual dependency risk is recorded as a reopen trigger, not a blocker.

**Authority/mechanism boundary.** Keycloak = mechanism/provider for credential authentication. Authority for organizational identity, membership, grants stays in MetalDocs. M3 makes the boundary structural.

**Proof strategy.** B2: binding uniqueness (`UNIQUE(iss,sub)` per binding), binding↔User integrity under B1 law, assurance-fact representation for `requires_reauthentication` steps (ACR/`auth_time` consumption), negative proof that no provider role/group claim has any consumable representation (M3), reconciliation choreography proof (M2). Viability probe of the auth-code+BFF flow against a disposable Keycloak realm is legitimate design-time evidence and requires no product implementation.

**Strongest adversarial counterexample.** Smallest on-prem deployment gains a JVM service + DB it did not need for pure local login. Accepted: the operator owning deployment reality mandates the IdP; the counterexample is a footprint cost, not an authority/correctness defect. Recorded as reopen trigger if a deployment class emerges where the footprint is prohibitive.

**Decision.** RESTRUCTURE NOW — Keycloak as V1 authentication provider, with M1–M3 fixes at adjudication.

**Reopen triggers.** Keycloak licensing/maintenance change of MinIO class; a deployment class where the footprint is prohibitive; a real per-tenant authentication-policy divergence that single-realm + Organizations cannot express (see Q6/Q7).

## 3.2 Packet questions 1–9

1. **Total complexity:** materially reduced, not moved — the deleted surface includes credential features not yet built (reset/invite/MFA) that V1 parity would force; added surface is deployment/ops of one mature service the operator already knows.
2. **Stability:** yes — CNCF, Red Hat-backed, LTS-grade cadence; pin major version; risk recorded as reopen trigger.
3. **`iss+sub`:** correct and verified as the only spec-guaranteed stable RP identifier; email/username as identity keys are spec-violating. Binding table keyed on (issuer, subject).
4. **Opaque app Session:** justified, not duplicate. The BFF architecture itself requires an application session distinct from the IdP SSO session; the app session is where MetalDocs enforces its own timeouts, revocation on suspension, and assurance consumption. `session.manage` survives against app sessions.
5. **BFF/cookie:** correct — IETF-recommended for precisely this class (business app handling sensitive data); token-in-browser alternatives are strictly worse for this product.
6. **Realm per environment:** sound. Realm-per-tenant would bind provider realm lifecycle to Tenant lifecycle (provider acquiring provisioning authority) and explodes operational surface.
7. **Tenant-specific IdP:** Keycloak Organizations + org-specific IdP routing handles per-tenant upstream IdPs inside one realm as a projection of MetalDocs Tenant. Reopen only if a tenant requires genuinely divergent authentication policy that conditional flows/Organizations cannot express.
8. **Second-authority pressure:** the risk is real and is answered structurally (M3), not by policy discipline.
9. **R10-A reopen:** yes — bounded amendment of the Authentication owned-fact family and provider placement (M1). No topology change.

---

# 4. GCR-R2 — Managed Artifact Store provider entitlement — **APPROVE with M6, L1**

## 4.1 Method chain (condensed)

**Evidence.** `minio/minio` archived/unmaintained — verified primary external change (§0.1). Implementation: MinIO is the only implemented provider and the dev/CI store (§0.2).

**Root cause.** Correctly identified: a replaceable mechanism was named as first-class product entitlement. The residual defect the candidate does not fully close: it still frames the outcome as a provider list. M6 restates the durable first-class surface as the port + conformance contract.

**Target invariant.** Frozen storage semantics (hash identity, one active store, copy+verify+cutover) — unchanged and unchallenged by the external event. The archived server invalidates a provider name, not the architecture.

**Alternatives.** (1) Replace MinIO with another named self-hosted store now — rejected: repeats the coupling class with no deployment consumer (Garage/Ceph/AIStor selection without evidence). (2) Keep MinIO OSS as entitlement — rejected: unmaintained source-only distribution cannot be a first-class *target* commitment for governed content. (3) Candidate (port + conformance seam, S3 as reference profile) — correct.

**Structural Inversion.** If R9.5 had named Garage and Garage were archived, the same removal would follow — the root cause generalizes, which is why M6 removes the provider-name coupling class rather than swapping names.

**YAGNI.** Selecting a self-hosted provider now is speculative capability; the conformance seam is the prepared seam. Deferral is safe **because** the seam (port + suite) is real — which makes the R10-C conformance-execution mechanism (M6) load-bearing, not optional.

**Strongest counterexample (Q10 gap claim).** "A self-hosted production customer exists at V1 and cannot use AWS S3." No such named consumer exists in any authority document. If one appears, the reopen trigger fires and provider selection happens through the conformance suite — that is the designed path, not a gap. The genuine near-gap is dev/CI (MinIO server today); M6 bounds it as frozen test mechanism.

**Decision.** RESTRUCTURE TARGET PROVIDER SET as candidate states, with M6 restatement.

**Reopen triggers.** Named self-hosted production consumer; S3-API divergence that breaks conformance portability; test-double unavailability.

## 4.2 Packet questions 10–13

10. **Gap:** no production gap now (no named self-hosted consumer); real transition note is dev/CI dependence on MinIO server — bounded as mechanism (M6).
11. **AWS S3 as reference profile:** reasonable; provider URL/key/version stays out of business identity per frozen law; reference profile ≠ authority.
12. **Select now vs seam:** seam is genuinely sufficient; selecting now would be the same defect class the reopen removes.
13. **Conformance suite minimum:** presign/confirm round-trip with hash verification over the port; over-size and hash-mismatch rejection with object cleanup; tenant-prefix isolation; no-overwrite of existing keys; copy+verify+cutover relocation semantics; restore byte-integrity (hash re-verification). R10-C owns the concrete contract; these follow from frozen invariants plus `verified_store.go` behavior worth preserving.

---

# 5. GCR-R3 — Bounded malware inspection before confirmation — **APPROVE (BOUNDED RESTRUCTURE) with M5, L2**

## 5.1 Method chain (condensed)

**Evidence.** OWASP verified **with nuance**: AV scanning is defense-in-depth "if available", not a mandated core control (§0.1). The honest driver is therefore not OWASP authority but product exposure: MetalDocs parses, renders, previews and **distributes** third-party uploaded bytes to other users under a governed-trust label, and current implementation confirms bytes with hash+size checks only — not even the frozen R9.5-7 format-coherence validation exists yet (§0.2). A governed document-control system is a high-credibility malware distribution channel precisely because recipients trust released content.

**Root cause.** Correctly identified: R9.5-7 bundled a small ingress gate with a large security platform and rejected the bundle. YAGNI removed a safety property along with genuine accidental complexity — the exact failure mode the Method's YAGNI clause forbids ("MUST NOT remove a known invariant, safety property…").

**Target invariant.** As stated, with M5: untrusted bytes → STAGED → validation → **required-by-deployment-profile inspection (production default: required)** → CONFIRMED; unavailable/incomplete inspection ⇒ non-confirmation, visible failure. Fail-closed matches B1's isolation posture and avoids the L5 class (fail-open by configuration).

**Alternatives.** (1) Status quo (defer everything) — leaves distribution-amplified ingress risk with zero inspection; rejected. (2) Full security platform (quarantine/CDR/rescans) — rejected by candidate itself; correct. (3) Candidate bounded gate — one inspection step at the existing staging→confirmation seam, scanner unselected. Correct global maximum: the seam already exists (`Confirm` at `verified_store.go:113-146` is the natural insertion point), so essential safety is bought at near-zero structural cost.

**Structural Inversion.** If V1 had shipped with scanning and the proposal were to remove it to simplify, removal would be rejected as deleting a safety property. Asymmetry confirms the property is essential, not preferential.

**Q14 (proportional?).** Yes, with M5. Bounded to one gate at one seam; no new lifecycle, no new owner, no scanner selection, no quarantine aggregate. What keeps it proportional is exactly what the candidate excludes.

**Q15 (fail-closed operability).** Acceptable: failure is visible (upload/confirm fails), bounded (staged bytes remain staged; authoring metadata unaffected), and honest (silent-weaken is worse than visible outage for a governed-content product). Scanner availability posture (local engine vs sidecar, retry) is R10-C.

**Q16 (no quarantine lifecycle?).** Yes: non-confirmed staged bytes already have a lifecycle (staging expiry/GC). Rejected bytes need no new aggregate — rejection is an operation outcome plus audit evidence, not a state machine.

**Q17 (ownership split).** Correct: Artifact owns staging/validation/confirmation facts (R10-A already says so); inspection is a platform mechanism invoked at confirmation; the requirement level is deployment configuration (M5). No new semantic owner.

**Decision.** BOUNDED RESTRUCTURE as candidate states, with M5 wording at adjudication.

**Reopen triggers.** Real quarantine/rescan/CDR customer requirement; evidence the gate cannot be operated fail-closed in a real deployment class.

---

# 6. GCR-R4 — Tenant DEK / cryptographic erase — **APPROVE (STOP / SPLIT PREREQUISITE) with M4**

## 6.1 Method chain (condensed)

**Evidence.** NIST verified: cryptographic erase is only meaningful against keys that actually protect Target Data (§0.1). Authority texts assert DEK lifecycle facts (R8 §6, R9.5-2, R10-A Organization) without naming the protected Target Data — the packet's evidence-gap claim is structurally correct. However (M4), current implementation contains exactly one real family: audit-PII envelope encryption (`audit_events.payload`), created to reconcile append-only Audit with tenant erasure — and simultaneously demonstrates the false-security failure mode the candidate warns about (KEK unset ⇒ plaintext + silently skipped shred, so the "crypto-shred" step of erasure can be a no-op while the erasure record still completes).

**Root cause.** A key-destruction step was frozen into the erasure sequence (R8) without a frozen statement of what the key protects. Mechanism was promoted ahead of its Target Data — inverse of Method order.

**Target invariant (of the split).** Either (a) a named Target Data family exists whose confidentiality depends on the Tenant DEK, with envelope/boundary/backup/restore semantics proven and fail-closed enforcement, or (b) the mandatory DEK lifecycle, the crypto-shred erasure claim, Organization's key-custody fact family, the B2 key-custody scope line and the platform crypto/KEK mechanism entry are all removed together. No third state (key destruction as ceremony) is acceptable.

**The decisive case B2/B6 must adjudicate (M4).** The immutable-audit-PII family: ledger §6 requires erasure to preserve a "non-PII audit/platform skeleton" while Audit is append-only/tamper-evident. If B6 puts person-identifying data inside immutable audit rows, crypto-shred of a per-tenant key is the standard resolution of that conflict — a real Target Data family. If B6 instead keeps immutable rows PII-free (opaque UUID actor references whose linking records live in erasable Organization state), the conflict dissolves and no DEK is needed. Both arms are coherent; the choice belongs to B2/B6 + R10-C (backup implications: ledger §6 already prescribes tombstone reconciliation on restore, so backups alone do not force a DEK). The candidate's proof conditions 1–6 are exactly right; M4 adds only that this named family must be explicitly dispositioned before the REMOVE arm may be taken.

**Q18.** Today: audit event payloads (when KEK configured). In the accepted target: unnamed — that is the defect.
**Q19.** If no family survives adjudication: yes — mandatory DEK is accidental complexity plus false-security semantics, and the removal cascade of M4 applies.
**Q20.** No — the frozen retention-aware erasure invariant (verified deletion, tombstones, restore reconciliation) does not depend on key destruction; it is preserved by the candidate explicitly.
**Q21.** If justified: single per-tenant DEK wrapped by a platform KEK (current envelope shape is adequate evidence of the minimal model), Organization owns lifecycle facts, platform owns wrap/unwrap, **fail-closed** (a deployment claiming crypto-erase must refuse to run erasure with crypto disabled).
**Q22.** Distinct requirements: truthful retention-aware erasure is frozen and real; cryptographic erase is only required where intelligible ciphertext must survive in immutable/backup surfaces — i.e., exactly the audit-PII/backup question B2/B6/R10-C must answer.

**Decision.** STOP / SPLIT PREREQUISITE as candidate states, with M4 sharpening.

**Reopen triggers.** As candidate; plus: if REMOVE is taken, a later requirement to encrypt tenant content at the application layer re-enters through the retained seam, not by resurrecting the mandatory lifecycle.

---

# 7. Clarifications

## GCR-C1 — **ACCEPT**

The refinement preserves the North Star's actual claim (MetalDocs = system of record for product/organizational identity, governance, revision, evidence, documentary context) while removing a misreading ("identity" ⊇ credential storage) that would contradict GCR-R1. Binding language matches OIDC-verified `iss+sub` stability. MetalDocs remains authoritative for **who exists in the organization** (User provisioning/membership); the provider is authoritative only for **how that person authenticates**.

## GCR-C2 — **ACCEPT with L4**

"One MetalDocs product-state PostgreSQL database" is the correct reading of B1 — B1 §9.1 governs product state, and provider-owned persistence (Keycloak's DB) was never inside that law. Same-server operational co-location does not merge persistence authority. Add L4: no MetalDocs invariant may depend on cross-database atomicity between product-state and provider-owned databases (the C2 face of M2). With L4, this is a wording clarification, not a B1 reopen.

---

# 8. Fifth material local maximum — **NONE FOUND at authority level**

Whole-platform sweep (packet Q23–Q30) across R3–R10-B1. Candidates examined and rejected:

| Candidate | Disposition |
|---|---|
| Rendering (DOCX→PDF) rebuild | Already classified as replaceable provider mechanism (R10-A §2.3); provider selection is R10-C/D. No authority defect. |
| Search engine rebuild | Rebuildable projection, external engine defer-confirmed with trigger. Sound. |
| Async/workflow rebuild (vs Temporal) | Defer with explicit R10-D trigger already recorded. Sound. |
| Audit tamper-evidence chain (vs ledger DB/QLDB-class) | Small DB-internal mechanism; an external ledger product would add a dependency with authority gravity for zero invariant gain. Keep. |
| Authorization engine rebuild (vs OpenFGA/SpiceDB) | Defer-confirmed; V1 model (5 roles, additive grants, typed scopes) is differentiated domain semantics, not commodity RBAC. Keep. |
| EigenPal editor as vendored provider | Provider-around-WorkingContent with anti-corruption seam, frozen by R9.5-8; no new evidence. Keep. |
| Interchange as wrapper (Q26) | Owns real process truth (dry-run, deterministic outcomes, reconciliation); not a provider wrapper. Keep. |
| Post-R1 Authentication BC as wrapper (Q26) | Retains real semantics (binding, app session, assurance); small ≠ wrapper; separateness is what keeps the provider replaceable. Keep. |
| Abstraction existing only because another exists (Q27) | Found one: key-custody lifecycle facts + platform KEK machinery exist only because the DEK exists — already captured inside R4's removal cascade (M4), not a fifth reopen. |
| Deletable-today concept (Q30) | R9.5-8's subtractive pass holds; no concept found whose deletion removes complexity without weakening a distinct property. Nearest candidates (EditorSession, WorkingSnapshot, System Value Catalog) each protect a distinct property (staleness lease, recovery, bounded resolution). |
| YAGNI-removed essential property beyond malware (Q28) | None found. Trusted-timestamp (TSA), signed exports, eDiscovery remain honestly deferred with claims scoped accordingly (V1 claims authenticated application approval, not qualified signature). |

Three near-misses recorded, all bounded into existing reopens/blocks rather than new reopens:

1. **Conformance-execution mechanism** (dev/CI S3 endpoint after MinIO demotion) — folded into M6 / R10-C.
2. **Frontend credential-journey deletion** under Keycloak — folded into L3 / R10-E.
3. **S3 client library selection** (`minio-go` inheritance) — folded into L1 / R10-C.

**What would have falsified "none":** a mature commodity product owning ≥ the equivalent of one bounded context's essential semantics (as Keycloak does for credentials) with a changed requirement or external event making local ownership a dead end. Systematically checked each of the 8+3 owners against that test; only Authentication matched — and it is already GCR-R1.

# 8.1 Confirmed-decision falsification pass (packet §8 table)

Each CONFIRM row was attacked for a counterexample under current evidence; none produced material reopen evidence. Notable checks: modular monolith (no independent-deploy/trust boundary in evidence — Keycloak is a provider beside the monolith, not a service split of it); B1 composite tenant law (Keycloak adds no tenant-owned table to product state; no counterexample); outbox law (unchanged by provider adoption); specialized Approval (no branching/BPM requirement surfaced); WORM-as-enforcement (unchanged). The packet's own reopen-only-on columns remain the falsification criteria of record.

---

# 9. Resulting reopen sets

## R9.5 reopen set (bounded deltas, at adjudication)

```text
ledger §1   Authentication paragraph        → Keycloak V1 provider + binding model (GCR-R1)
ledger §7   North Star wording              → C1 refinement
ledger §9   R9.5-2 provider set             → MinIO OSS demotion; port+conformance restatement (GCR-R2 + M6)
ledger §14  R9.5-7 content safety           → bounded inspection gate, secure-by-default profile (GCR-R3 + M5)
ledger §6   R8 erasure DEK/crypto-shred     → CONDITIONALLY amended pending R4 resolution (B2/B6/R10-C);
                                              not edited now
```

Everything else in R9.5 remains FROZEN; reopen set otherwise EMPTY.

## R10-A reopen set (bounded amendment; topology 8+3 unchanged)

```text
Authentication owned-fact family    → provider subject binding / app session / assurance facts / ACL contract
                                      (credential storage, password policy, activation, lockout → provider)
Organization key-custody fact line  → annotated conditional on R4 resolution
commodity mechanisms list           → add IdP provider client/adapter; crypto/KEK entry conditional on R4
provider placement table            → Keycloak adapter → Authentication infrastructure
```

## R10-B1 reopen set

```text
STRUCTURAL REOPEN = NONE
wording amendment §9.1 → "one MetalDocs product-state PostgreSQL database" (C2)
                        + no cross-database atomicity clause (L4)
```

All B1 laws (composite tenant identity, FK action law, RLS, transaction law, background discovery) stand unchanged.

---

# 10. R10-B2 resumption and promotion conditions

**R10-B2 may resume after adjudication** — and only after it, because B2's own scope changes:

- ADD: provider subject-binding representation (`UNIQUE(iss, sub)`), assurance-fact representation for fresh-auth/step-up consumption, binding↔User lifecycle + reconciliation choreography (M2), structural no-provider-role-consumption proof (M3);
- REMOVE: local credential-table design (password hash/policy/lockout/activation state) from B2 scope;
- CONDITION: key-custody fact representation proceeds only under R4's PROVE arm; under REMOVE it is deleted from scope with the M4 cascade;
- RETAIN: all other B2 fences (no OpenFGA/SpiceDB, no nested groups, flat groups, five roles, RLS tenant-isolation-only, PlatformOperator outside tenant RBAC).

**Exact promotion conditions for this GCR:**

1. Operator ratifies R1–R4 dispositions and C1/C2 under the Method, including the M1–M6 material fixes.
2. Ledger §1/§7/§9/§14 amended as bounded deltas (no rewrite of frozen sections beyond the named paragraphs); §6 marked conditionally pending R4 resolution.
3. R10-A amended per §9 above, with amendment provenance recorded in `r10-technical-architecture.md`.
4. B1 §9.1 wording amended per C2 + L4.
5. `wiki/references/current-agent-handoff.md` B2 scope updated per §10 above; the "no Keycloak without a real trigger" fence replaced by the ratified Keycloak V1 decision (the trigger fired: changed operator requirement).
6. R4 proof obligations assigned explicitly: B2 (key-custody facts, if retained), B6 (immutable-audit-PII posture), R10-C (backup/restore implications), with the REMOVE-cascade named.
7. R3 invariant text lands with M5 (deployment-level requirement authority, production secure-by-default); scanner selection and failure posture remain R10-C.
8. R2 lands with M6 (port + conformance as first-class surface; AWS S3 reference profile; dev/CI S3 endpoint recorded as bounded mechanism); L1 noted for R10-C.

---

# 11. Review-round necessity

**Another full adversarial round is NOT necessary.** The four reopens were coherently scoped, externally evidenced, and survived falsification with bounded fixes only; no fix changes direction or creates a new authority. A **bounded delta review of the amended authority texts** (ledger deltas, R10-A amendment, B1 wording, handoff B2 scope) is sufficient before R10-B2 opens. R4's resolution inside B2/B6/R10-C is already covered by those stages' own independent review gates.

---

# 12. Required reporting summary

```text
VERDICT: APPROVE GCR MINIMAL REOPEN SET WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 6  (M1 R10-A amendment required; M2 cross-DB provisioning choreography;
              M3 structural ACL; M4 audit-PII family + fail-closed + removal cascade;
              M5 secure-by-default inspection profile; M6 port-as-entitlement + CI mechanism)
LOW     = 5  (L1 SDK choice; L2 validator/scanner order; L3 frontend journey deletion;
              L4 cross-DB atomicity clause; L5 fail-open config evidence)

R1 = APPROVE (RESTRUCTURE NOW)            + M1 M2 M3 L3
R2 = APPROVE (RESTRUCTURE PROVIDER SET)   + M6 L1
R3 = APPROVE (BOUNDED RESTRUCTURE)        + M5 L2
R4 = APPROVE (STOP / SPLIT PREREQUISITE)  + M4
C1 = ACCEPT
C2 = ACCEPT                               + L4

FIFTH MATERIAL LOCAL MAXIMUM = NONE (three bounded near-misses recorded in §8)

R9.5 reopen   = {ledger §1, §7, §9/R9.5-2, §14/R9.5-7} + conditional {§6 DEK clause}
R10-A reopen  = bounded Authentication/Organization fact-family amendment; topology unchanged
R10-B1 reopen = none structural; §9.1 wording per C2 + L4

R10-B2 = MAY RESUME after adjudication with amended scope (§10)
NEXT REVIEW = bounded delta check of amended authority texts; no full round required
```
