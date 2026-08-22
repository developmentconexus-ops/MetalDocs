# T9 — Golden Flows & Validation Baseline

> **TEMPORARY T9 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T9 derives the smallest falsifiable composed-system validation contract capable of disproving the accepted T1→T8 architecture.

T9 does **not** implement Product code, create test-only Product semantics, add operations, choose T10 transition mechanics, create the T11 implementation graph, or perform the T12 readiness attack.

Fixed input:

```text
T1 → T8-H                 CLOSED / OPERATOR-RATIFIED / INTEGRATED
application operations     78
orphaned operations         0
invented operations         0
operation 79                ABSENT
Product implementation      BLOCKED
legacy implementation       ABSENT
```

A T9 failure may reopen only the smallest owning T1→T8 authority when the failure demonstrates a real contradiction. A failing proof is never repaired by inventing a T9-only endpoint, state machine, permission, semantic owner, persistence truth or runtime capability.

## 2. Global validation law

A claim is T9-proved only when all three are true:

```text
1. the real production subject/boundary is executed or mechanically inspected;
2. the expected invariant is observed on the positive path;
3. a causal negative/fault case demonstrates that the same proof fails when the invariant is violated.
```

Not sufficient by itself:

```text
mock/fake success for an external dependency
fixture text that merely contains the expected answer
a probe that validates only its own fixture schema
a green test that bypasses the production transaction/AuthZ/wire/runtime path
presence of a file, dependency or config key without executing its property
```

Legitimate negative controls attack the actual subject, for example:

```text
forged/replayed/tampered OIDC callback
verified provider subject with no current ProviderSubjectBinding
stale real ETag
same Idempotency-Key + changed semantic request
required Audit append failure inside the real local transaction
revoked/disabled current User or grant
concurrent distinct Document creates competing for one code
corrupt stored bytes against a real semantic descriptor
stopped renderer/scanner dependency
incompatible database schema
missing required secret / impossible content envelope
River redelivery/terminal failure
backup-pin/GC race
restore from a real backup recovery point
SIGTERM during accepted work
```

Tool identity is not authority. T8-G's current proof-backed candidates (for example Schemathesis, Testcontainers-Go, `govulncheck`, OSV-Scanner) may realize a proof lane only when they execute the required subject and falsifier.

## 3. Evidence classes — exactly 6

| Class | Evidence | Cannot be replaced by |
|---|---|---|
| E1 — structural/contract | real package graph, canonical OpenAPI, generated projections, schema/config validators | prose or self-referential string checks |
| E2 — database integration | real PostgreSQL target primitives + real transaction/constraint behavior; River when claimed | repository/mock implementation of PostgreSQL semantics |
| E3 — application HTTP | real composed application path through transport → application → owners/mechanisms | direct owner calls presented as API proof |
| E4 — browser | real browser against the actual SPA/application origin | component mocks for cookie, redirect, routing or Fetch behavior |
| E5 — dependency | real selected OIDC/content-store/renderer/scanner mechanism or provider-equivalent proof lane | fake clients when claiming dependency behavior |
| E6 — failure/recovery | controlled outage, corruption, redelivery, shutdown, backup/restore or resource-envelope attack | scripted fake returning the expected error |

No class is mandatory when the protected property does not depend on it. Proportional evidence is required; ceremonial end-to-end duplication is not.

## 4. Golden Flow set — exactly 6

The six flows are a composition basis, not an operation-by-operation acceptance suite. Full 78-operation closure is split deliberately:

```text
static census/schema/generated-projection closure
+
runtime execution of the accepted T8-E wire-conformance fixture classes
+
representative composed Golden Flows
```

### GF1 — Identity, session, access and revocation

**Purpose:** prove the complete browser trust chain from the OIDC boundary through current MetalDocs Authorization without moving Authorization into OIDC or React.

Success lane:

```text
external OIDC authentication
→ callback/exchange validates the external protocol result
→ verified issuer+subject resolves one current ProviderSubjectBinding
→ exact enabled User is resolved
→ HttpOnly ApplicationSession issued
→ synchronizer CSRF material obtained through authenticated session
→ Product read succeeds only after current User/Role/scope/domain checks
→ access administration changes current eligibility/grant
→ next protected request reflects the revocation
→ historical actor/Audit attribution remains intact
```

Required falsifiers:

```text
forged/replayed/tampered callback — including invalid state/nonce/code/issuer — cannot create an ApplicationSession
verified provider subject with no current ProviderSubjectBinding cannot obtain a session or auto-create a User
provider role/group claim alone cannot grant MetalDocs Permission
missing/wrong CSRF cannot reach unsafe semantic effect
revoked/disabled current access cannot remain usable through stale browser/server cache
expired ApplicationSession fails authentication
session cookie replay after endSession fails authentication
ProviderSubjectBinding replacement terminates all existing ApplicationSessions for that User
OIDC network outage does not destroy an already-valid ApplicationSession
non-disclosable resource stays non-disclosable
```

Minimum evidence: E2 + E3 + E4; E5 for the selected OIDC profile.

### GF2 — Governance configuration to atomic Document creation

**Purpose:** prove that current administration/configuration can feed one real Document creation without duplicate truth, stale configuration acceptance or partial commit.

Success lane:

```text
current Area / access / DocumentType-governance configuration
→ whole-replacement governance/config mutation under its accepted ETag when the flow exercises configuration
→ creation options/numbering preview
→ createDocument with one logical Idempotency-Key
→ final number allocation + Document current truth + required Audit commit atomically
→ retry of the same completed logical command returns the accepted replay semantics
```

Required falsifiers:

```text
stale configuration ETag cannot silently overwrite newer governance/config truth
concurrent retry of the same logical create cannot create two Documents or two semantic effects
same key + changed semantic request is rejected
two distinct concurrent logical creates in one numbering scope cannot commit the same Document code
required Audit append failure rolls back the business mutation
retired/ineligible referenced configuration cannot be accepted from stale preview
number preview never reserves or becomes final authority
```

Minimum evidence: E2 + E3.

### GF3 — Revision authoring, upload admission and concurrency

**Purpose:** prove browser authoring through exact content, upload capability and DRAFT concurrency boundaries.

Success lane:

```text
createDocumentRevision
→ load Document Work / DRAFT ETag
→ allocate bounded upload capability
→ browser uploads intended bytes to ManagedContentStore
→ complete/admit exact descriptor
→ update DRAFT under exact If-Match
→ read back canonical DRAFT/source truth
→ createSubmission with same logical-command key + semantic DRAFT precondition
```

Required falsifiers:

```text
stale DRAFT ETag always yields the DRAFT-specific precondition failure; no silent merge/LWW
expired upload allocation is never revived/reused
provider PUT success alone is not READY/WorkingContent/Submission
wrong size/hash/format cannot become accepted exact content
same createSubmission key with materially changed semantic input cannot replay
```

Minimum evidence: E2 + E3 + E4 + E5 for the selected content-store profile.

### GF4 — Governance to Release / OfficialRendition

**Purpose:** prove the most important multi-owner lifecycle composition from immutable Submission through governance to official/effective truth.

Success lane:

```text
Submission
→ GovernanceAttempt + immutable Step configuration/candidate snapshot
→ actor-relevant governance work
→ feedback/Step Decision under current Authorization
→ required owner mutation + required Audit + any actually-required durable effect intent commit coherently
→ when durable work exists, River redelivery-safe processing
→ conditional rendition path
→ canonical Release/EFFECTIVE truth
→ Document Official + History + official exact-byte read
```

Two mandatory realization variants:

```text
submitted PDF + required PDF
  → reuse exact admitted bytes
  → no renderer intent/copy/transformation

submitted DOCX + required PDF
  → private renderer
  → candidate bytes re-admitted/verified
  → eligibility reloaded before OfficialRendition finalization
  → selected renderer/font/config profile passes representative MetalDocs DOCX fidelity corpus
```

Required falsifiers:

```text
GroupMembership drift after Step activation cannot rewrite the frozen candidate snapshot
current Authorization denial cannot be bypassed by being in the snapshot
required Audit/actually-required intent failure cannot leave a committed semantic transition
PDF→PDF path cannot emit renderer intent merely for symmetry
River duplicate/redelivery cannot create duplicate Release/Rendition semantic truth
renderer failure cannot convert acceptance into false success
representative DOCX corpus fidelity failure blocks production eligibility for that renderer profile
provider byte corruption cannot produce a complete successful exact-byte response
malicious/unscanned required governed bytes cannot cross the immutable governed boundary
clean scan evidence must bind the exact bytes that are later admitted at the governed boundary
```

Minimum evidence: E2 + E3 + E5 + E6; E4 for browser governance/official lenses.

### GF5 — Official discovery, obsolescence and disclosure-safe routing

**Purpose:** prove that current official/discovery views remain honest while work/obsolescence context is disclosed only when authorized.

Success lane:

```text
Library/Search official discovery
→ DocumentOfficialView
→ open_revision present only for the unique disclosable current DRAFT/SUBMITTED Revision
→ active_obsolescence_request_id present only for the unique disclosable ACTIVE request
→ create obsolescence request
→ either withdraw through the admitted path
   OR complete the accepted governance/system-owned obsolescence progression and observe resulting official/history truth
→ History remains semantic history, not a current-resource resolver
→ Audit remains evidence, not current business truth
```

Required falsifiers:

```text
absence of open_revision/request id cannot prove non-existence to an unauthorized caller
official lens never renders DRAFT WorkingContent as official truth
Search cannot acquire a second materialized/external current-truth authority
History/Audit cannot be used to bypass current disclosure/AuthZ
withdrawn/obsolete progression cannot leave a stale active request pointer as Product truth
operation 79 cannot appear to repair a navigation/read-symmetry gap
```

Minimum evidence: E2 + E3 + E4.

### GF6 — Runtime failure isolation, shutdown and recovery

**Purpose:** prove that mechanisms fail according to their accepted blast radius and recovery cannot resurrect unsafe state.

Success/fault matrix:

```text
bad config / missing required secret
  → serve fails closed

accepted content envelope > configured renderer/scanner/body/workspace capability
  → deployment/startup/conformance proof fails before production readiness

incompatible schema
  → serve fails

PostgreSQL unavailable
  → /readyz fails while process remains live

renderer unavailable
  → unrelated Product serving survives; required rendition work fails/retries truthfully

MalwareInspector unavailable/signature policy unsatisfied
  → governed admissions requiring CLEAN blocked; unrelated serving survives

OIDC unavailable
  → new login/provider-dependent work may fail; already-valid ApplicationSessions survive their own validity

SIGTERM
  → unready first; bounded drain; interruption relies only on accepted rollback/idempotency/redelivery semantics

River terminal required work
  → visible, inspectable, redrivable; queue state never becomes Product truth

backup/restore
  → DB + required exact-content manifest form a complete recovery point
  → backup pin / GC exclusion prevents selected required content from disappearing during capture
  → restored sessions are invalidated before serving
  → privacy/security reconciliation completes before readiness
```

Additional falsifiers:

```text
secrets never appear in logs/metrics/browser/health
renderer outbound network denied
scanner signature updater has only approved signature-source egress
resource profile remains bounded for accepted content size × configured concurrency
surviving runtime component census contains only named consumers
```

Minimum evidence: E2 + E5 + E6.

## 5. Cross-cutting validation matrix — exactly 10 properties

### V1 — Product/wire census, generated projections and runtime wire conformance

V1 has two distinct proof lanes and both are mandatory.

**Static/structural lane — E1:**

```text
application operations   = 78
orphaned                  = 0
invented                  = 0
operation 79              = absent
OpenAPI paths/operationIds/schemas/headers/Problems match accepted wire law
generated Go/TypeScript projections compile and cannot widen closed wire vocabularies
T8-E-FR read-symmetry precision is represented only through the wire SSOT; no parallel executable authority exists
```

A deliberately added/removed/renamed operation or incompatible schema must make the real contract proof fail.

**Runtime lane — E3, with E2/E4/E5 only where the fixture's protected property requires them:**

Execute the accepted T8-E §9.4 wire-conformance fixture classes against the real composed HTTP path `transport → application → owners/mechanisms`; direct owner calls do not close this lane.

The runtime suite must exercise the accepted classes, including at least:

```text
strict JSON/query decoding, including duplicate/unknown members
body on bodyless operation rejection
405 exact Allow behavior
unsupported media/content-coding and 65,536-byte raw JSON ceiling
PROFILE_REPLACE If-Match / If-None-Match matrix
cursor tamper/filter-replay rejection + current AuthZ recheck across pages
complete/non-truncated creation-options projections
Content-Digest/body SHA-256 coherence
Range rejection
corrupt exact bytes cannot emit a successful complete response
ReplaySnapshot <= 2048 and closed/stable replay shape
same-PDF RequireOfficialRendition emits no duplicate bytes/job/renderer intent
operation-specific Problem/header/status behavior from the canonical 78-row ledger
```

Each fixture class requires a positive case and at least one causal negative that makes the same production path fail when its invariant is violated. Property-attack tooling may help generate cases but never replaces the real composed path.

### V2 — Closed-world architecture dependency graph

Must classify the live first-party target packages bidirectionally and reject every unallowed edge class, including transport→owner, application→application, owner→owner, application→owner-private, platform→owner, foreign SQL and hidden second Authorization evaluator.

Negative fixtures mutate the real import/SQL subject and run the same production architecture verifier.

### V3 — Authentication, Authorization, disclosure and CSRF

Must prove the canonical ALLOW/default-DENY authority is singular, current access is revocable, domain predicates fail closed, provider claims never become Product authorization, OIDC callback results cannot bypass binding/session admission, disclosure is respected, and unsafe browser requests require the accepted CSRF mechanism.

### V4 — Transaction + required Audit atomicity

The transition-class enumeration source is the **closed T3 §15 Required same-local-commit Audit census**, reconciled by later accepted T8-E precision. T9 must cover every required census class, using representative instances inside a class rather than sampling away entire classes.

For each class:

```text
semantic mutation + every required Audit event + required durable intent (when any)
commit together or none commit
```

Inject failure after mutation but before required append/intent completion and prove no partial semantic commit survives. Offboarding multi-event teardown must remain reconstructible under the same transaction.

### V5 — Idempotency and replay

Must prove current-user + operationId + UUID scoping, normalized semantic fingerprinting, live AuthZ/disclosure before replay disclosure, exact completed ReplaySnapshot behavior, changed-request same-key rejection, no duplicate semantic mutation/Audit and 24h semantic expiry.

### V6 — Concurrency / ETag / narrow serialization

Must prove stale whole-replacement semantics, stale DRAFT-specific semantics, protected current eligibility during correctness-critical transitions, one-open/effective uniqueness, concurrent Document-code uniqueness across distinct commands, and accepted READ COMMITTED + narrow serialization/CAS behavior under real concurrent transactions.

### V7 — Exact content, malware and rendition integrity

Must prove semantic ExactContentDescriptor authority over provider identity, exact size/hash/format verification, clean-gate requirements, scan evidence binding to the exact admitted bytes, same-PDF zero-transformation path, DOCX conditional-render path, representative DOCX fidelity corpus eligibility, immutable accepted rendition bytes and failure-before-success-headers for corruption.

### V8 — Durable work / River semantics

Must prove at-least-once duplicate/redelivery safety for **each activated named durable-effect class in the accepted T5 census**, including managed-content GC when active; current-state revalidation before effect/finalization; terminal visibility/redrive; and separation between River mechanism state and Product truth.

For managed-content GC specifically, a stale cleanup intent must be unable to delete current WorkingContent, immutable governed content, current claim-protected content or backup-protected content after eligibility/reference/claim state changes.

### V9 — Runtime readiness, failure isolation, resource and observability laws

Must prove accepted `/livez`/`/readyz` distinction, dependency-scoped blast radii, bounded time/memory/scratch/concurrency, no secret/PII leakage, and evidence that material failure paths emit attributable logs/metrics/traces without creating Product authority.

### V10 — Backup/restore and privacy/security readiness

Must prove a restorable recovery point contains the required PostgreSQL + exact-content set and that restore cannot serve until sessions/security/privacy reconciliation is complete.

A causal backup-capture race must demonstrate that backup pin / GC exclusion protects the selected required content until capture completes. A restore that resurrects sessions, leaves required exact content unverifiable or bypasses privacy/security reconciliation is a hard failure.

## 6. Forward-obligation disposition in T9

T9 does not reopen the 51 forward obligations wholesale.

Stage-relevant consumption:

```text
PRESERVE
  AUTH-02  external OIDC/Keycloak evidence remains replaceable mechanism evidence
  AUTH-06  no cross-system atomic IdP transaction claim
  ASY-04   one PostgreSQL-backed durable-job mechanism / River reference
  DB-01..07, DB-10  persistence/transaction properties are validation inputs
  SEC-01   maintenance/operator/system surfaces never receive implicit serving access

REOPEN
  CNT-03 EditorSession remains NOT required unless T9 produces concrete falsifying UX/integration evidence
  AUD-06 retention is outside current Golden Flow correctness unless a real requirement appears
  MIG-10 migration target families remain outside T9

DEFERRED
  remain counterexamples only; no T9 fixture, component or endpoint may activate them by convenience
```

Any T9 evidence that materially changes one of these dispositions returns to the smallest owner before the T9 baseline can close.

## 7. Coverage law

Golden Flow coverage is property-based, not line-count theater.

T9 closure requires:

```text
6/6 Golden Flows have executable future proof plans with positive + causal negative cases
10/10 cross-cutting properties have an identified production subject and falsifier
78/78 application operations remain in canonical wire census with no invented operation
T8-E §9.4 runtime wire-conformance classes execute against the real composed HTTP path
all T3 §15 required same-local-commit Audit classes are represented in V4 proof
all claim-relevant external mechanisms have a real-dependency proof lane
no MATERIAL accepted invariant is closed by mocks-only evidence
no unresolved contradiction is hidden as a test exemption/baseline waiver
```

T9 does **not** require every one of the 78 operations to appear in an end-to-end Golden Flow. Runtime operation/wire conformance plus representative composed flows is the smaller and stronger architecture proof.

## 8. Reopen / stop law

A T9 failure is classified before correction:

```text
implementation/proof defect
  → future implementation/T9 verifier fixes; architecture remains closed

accepted mechanism choice falsified but semantic architecture intact
  → smallest owning T8 technical authority reopens

wire/application consumer missing or extra
  → smallest Product/T6/T8-E owner reopens; never add operation 79 silently

semantic ownership/lifecycle/AuthZ contradiction
  → smallest T1→T7 owner reopens

transition-only problem
  → record for T10; do not solve in T9
```

Preference, test convenience, framework fashion, old implementation shape and hypothetical future scale are not reopen triggers.

## 9. Fable Round 1 adjudication

Independent Evidence PR #155 reviewed the exact operator-approved candidate `2d5d127e95821eac355296e0a7f09c93aef6cef3` and returned:

```text
VERDICT = NOT CONVERGED
MATERIAL findings = 2
Round 2 justified = YES
```

Lead adjudication:

```text
F1 MATERIAL  ACCEPT
  V1 now owns runtime execution of the T8-E §9.4 wire-conformance fixture classes
  against the real composed E3 path, not only static census/schema/generation proof.

F2 MATERIAL  ACCEPT
  GF1 now attacks forged/replayed/tampered OIDC callback semantics and verified-but-unbound
  provider subjects before ApplicationSession issuance.

F3 MINOR     ACCEPT
  GF1 explicitly covers expiry, endSession replay and binding-replacement session revocation.

F4 MINOR     ACCEPT
  V8/V10 explicitly name managed-content GC revalidation and backup-pin/GC capture race.

F5 MINOR     ACCEPT
  V4 binds its class enumeration to the closed T3 §15 same-local-commit Audit census.

F6 MINOR     ACCEPT
  GF2/V6 explicitly attack distinct concurrent Document creates competing for one code.

F7 NOTE      TRACEABILITY ONLY
  V1 explicitly records T8-E-FR read-symmetry as wire-SSOT-only executable meaning.

F8 NOTE      NO CHANGE
  rate limiting remains outside T9 because no accepted T1→T8 mechanism/invariant requires a limiter.
```

No T1→T8 authority reopens. Golden Flows remain 6, properties remain 10, evidence classes remain 6, application operations remain 78 and operation 79 remains absent.

## 10. Candidate exit criteria

The corrected T9 architecture candidate is ready for bounded Round 2 independent confirmation only when:

```text
Golden Flow set remains exactly 6 unless a genuinely uncovered accepted property is demonstrated
cross-cutting property set remains exactly 10 unless a gap is demonstrated
F1/F2 corrections are mechanically present without inventing a new authority or stage
no flow invents Product meaning
no proof depends on legacy implementation
no T10/T11/T12 work is embedded
implementation remains blocked
required repository CI is green on the exact corrected candidate HEAD
```

Round 2 is a bounded confirmation of the Round-1 corrections, not a fresh unconstrained redesign. Reviewer Evidence remains non-authoritative and must remain isolated from the candidate/main tree.
