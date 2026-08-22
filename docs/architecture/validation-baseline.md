---
id: golden-flows-validation-baseline
kind: authority
owner: architecture
summary: Operator-ratified T9 authority for the smallest falsifiable composed-system Golden Flows and validation proof classes derived from accepted T1→T8 authority.
---

# T9 — Golden Flows & Validation Baseline

> **Ratification:** OPERATOR-RATIFIED on 2026-08-21 after required CI and bounded independent Fable convergence. Current program stage, integration status, implementation permission and exact next action remain exclusively in `../roadmap.md`.

T9 freezes the smallest falsifiable composed-system validation contract derived from the accepted T1→T8 architecture. It defines **what future proof must attack and what evidence can close each claim**. It does not claim that a not-yet-implemented runtime has already executed those proofs.

T9 does not implement Product code, create test-only Product semantics, add operations, choose T10 transition mechanics, create the T11 implementation graph, or perform the T12 readiness attack.

## 1. Fixed envelope

```text
T1 → T8-H                 CLOSED / OPERATOR-RATIFIED / INTEGRATED
Golden Flows              6
cross-cutting properties  10
evidence classes           6
application operations     78
orphaned operations         0
invented operations         0
operation 79                ABSENT
new Permission              NONE
new semantic owner          NONE
Product implementation      BLOCKED
legacy implementation       ABSENT
```

A validation failure may reopen only the smallest owning accepted authority when it demonstrates a real contradiction. A failing proof is never repaired by inventing a T9-only endpoint, state machine, Permission, semantic owner, persistence truth or runtime capability.

## 2. Global validation law

A future implementation claim protected by this baseline is closed only when all three are true:

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
presence of a file, dependency or config key without executing its protected property
```

Representative causal negatives include:

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

Tool identity is not authority. Schemathesis-, Testcontainers-, vulnerability-analysis- or other tooling may realize a proof lane only when it executes the required subject and falsifier.

## 3. Evidence classes — exactly 6

| Class | Evidence | Cannot be replaced by |
|---|---|---|
| E1 — structural/contract | real package graph, canonical OpenAPI, generated projections, schema/config validators | prose or self-referential string checks |
| E2 — database integration | real PostgreSQL target primitives + real transaction/constraint behavior; River when claimed | repository/mock implementations of PostgreSQL semantics |
| E3 — application HTTP | real composed path through transport → application → owners/mechanisms | direct owner calls presented as API proof |
| E4 — browser | real browser against the actual SPA/application origin | component mocks for cookie, redirect, routing or Fetch behavior |
| E5 — dependency | real selected OIDC/content-store/renderer/scanner mechanism or provider-equivalent proof lane | fake clients when claiming dependency behavior |
| E6 — failure/recovery | controlled outage, corruption, redelivery, shutdown, backup/restore or resource-envelope attack | scripted fake returning the expected error |

No class is mandatory when the protected property does not depend on it. Proportional evidence is required; ceremonial end-to-end duplication is not.

## 4. Golden Flow basis — exactly 6

The six flows are a composition basis, not an operation-by-operation acceptance suite. Full application closure is deliberately split into:

```text
static census/schema/generated-projection closure
+
future runtime execution of the accepted T8-E wire-conformance fixture classes
+
representative composed Golden Flows
```

### GF1 — Identity, session, access and revocation

Success composition:

```text
external OIDC authentication
→ callback/exchange validates protocol result
→ verified issuer+subject resolves one current ProviderSubjectBinding
→ exact enabled User resolves
→ HttpOnly ApplicationSession issues
→ synchronizer CSRF material is obtained through the authenticated session
→ Product access succeeds only after current User/Role/scope/domain checks
→ access administration changes current eligibility/grant
→ next protected request reflects revocation
→ historical actor/Audit attribution remains intact
```

Required falsifiers include:

```text
forged/replayed/tampered callback, including invalid state/nonce/code/issuer, cannot create an ApplicationSession
verified provider subject without a current ProviderSubjectBinding cannot obtain a session or auto-create a User
provider role/group claim alone cannot grant MetalDocs Permission
missing/wrong CSRF cannot reach unsafe semantic effect
revoked/disabled current access cannot remain usable through stale browser/server cache
expired ApplicationSession fails authentication
session cookie replay after endSession fails authentication
ProviderSubjectBinding replacement terminates all existing ApplicationSessions for that User
OIDC network outage does not destroy an otherwise-valid existing ApplicationSession
non-disclosable resource stays non-disclosable
```

Minimum evidence: E2 + E3 + E4; E5 for the selected OIDC profile.

### GF2 — Governance configuration to atomic Document creation

Success composition:

```text
current Area / access / DocumentType-governance configuration
→ accepted whole-replacement configuration mutation under its ETag when exercised
→ creation options / numbering preview
→ createDocument with one logical Idempotency-Key
→ final number allocation + Document current truth + required Audit commit atomically
→ retry of the same completed logical command returns accepted replay semantics
```

Required falsifiers include:

```text
stale configuration ETag cannot silently overwrite newer governance/config truth
concurrent retry of the same logical create cannot create duplicate semantic effects
same key + changed semantic request is rejected
two distinct concurrent logical creates in one numbering scope cannot commit the same Document code
required Audit append failure rolls back the business mutation
retired/ineligible referenced configuration cannot be accepted from a stale preview
number preview never reserves or becomes final authority
```

Minimum evidence: E2 + E3.

### GF3 — Revision authoring, upload admission and concurrency

Success composition:

```text
createDocumentRevision
→ load Document Work / DRAFT ETag
→ allocate bounded upload capability
→ browser uploads intended bytes to ManagedContentStore
→ complete/admit exact descriptor
→ update DRAFT under exact If-Match
→ read back canonical DRAFT/source truth
→ createSubmission with logical-command key + semantic DRAFT precondition
```

Required falsifiers include:

```text
stale DRAFT ETag always yields the DRAFT-specific precondition failure; no silent merge/LWW
expired upload allocation is never revived/reused
provider PUT success alone is not READY/WorkingContent/Submission
wrong size/hash/format cannot become accepted exact content
same createSubmission key with materially changed semantic input cannot replay
```

Minimum evidence: E2 + E3 + E4 + E5 for the selected content-store profile.

### GF4 — Governance to Release / OfficialRendition

Success composition:

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

Mandatory realization variants:

```text
submitted PDF + required PDF
  → reuse exact admitted bytes
  → no renderer intent/copy/transformation

submitted DOCX + required PDF
  → private renderer
  → candidate bytes re-admitted/verified
  → eligibility reloaded before OfficialRendition finalization
  → selected renderer/font/config profile passes a representative MetalDocs DOCX fidelity corpus
```

Required falsifiers include frozen-candidate drift, current AuthZ denial, same-transaction Audit/intent failure, duplicate River delivery, renderer failure, fidelity failure, provider byte corruption, malware-gate failure and exact scan-evidence binding.

Minimum evidence: E2 + E3 + E5 + E6; E4 for browser governance/official lenses.

### GF5 — Official discovery, obsolescence and disclosure-safe routing

Success composition:

```text
Library/Search official discovery
→ DocumentOfficialView
→ disclosure-safe current open-Revision and active-obsolescence references
→ admitted obsolescence create/withdraw or completion path
→ resulting official/history truth
→ History remains semantic history, not a current-resource resolver
→ Audit remains evidence, not current business truth
```

Required falsifiers include unauthorized absence inference, DRAFT-as-official leakage, a second Search truth authority, History/Audit disclosure bypass, stale active-obsolescence routing and any attempt to create operation 79 to repair navigation/read symmetry.

Minimum evidence: E2 + E3 + E4.

### GF6 — Runtime failure isolation, shutdown and recovery

The proof plan must cover at least:

```text
bad config / missing required secret
  → serve fails closed

accepted content envelope exceeds configured capability
  → readiness/conformance eligibility fails before production promotion

incompatible schema
  → serve fails

PostgreSQL unavailable
  → /readyz fails while process remains live

renderer unavailable
  → unrelated Product serving survives; required rendition work fails/retries truthfully

MalwareInspector unavailable/signature policy unsatisfied
  → governed admissions requiring CLEAN block; unrelated serving survives

OIDC unavailable
  → new login/provider-dependent work may fail; otherwise-valid existing sessions survive their own validity

SIGTERM
  → unready first; bounded drain; interruption relies only on accepted rollback/idempotency/redelivery semantics

River terminal required work
  → visible, inspectable, redrivable; queue state never becomes Product truth

backup/restore
  → DB + required exact-content manifest form a complete recovery point
  → backup pin / GC exclusion protects selected required content through capture
  → restored sessions invalidate before serving
  → privacy/security reconciliation completes before readiness
```

Additional falsifiers cover secret leakage, renderer egress, scanner-updater egress, accepted resource bounds and component-census consumer justification.

Minimum evidence: E2 + E5 + E6.

## 5. Cross-cutting validation matrix — exactly 10

### V1 — Product/wire census, generated projections and runtime wire conformance

Two proof lanes are mandatory.

**Static/structural — E1:**

```text
application operations   = 78
orphaned                  = 0
invented                  = 0
operation 79              = absent
OpenAPI paths/operationIds/schemas/headers/Problems match accepted wire law
generated Go/TypeScript projections compile and cannot widen closed wire vocabularies
T8-E-FR read-symmetry precision exists only through the wire SSOT
```

**Runtime — E3, plus only claim-relevant E2/E4/E5:**

The future executable proof plan must target the accepted T8-E §9.4 runtime-conformance fixture classes on the real composed HTTP path `transport → application → owners/mechanisms`; direct owner calls do not close this lane.

The fixture source remains T8-E §9.4, including strict decoding/envelope behavior, exact method/status/header contracts, conditional matrices, cursor integrity/AuthZ rechecks, complete projections, exact-byte integrity, replay bounds and operation-specific wire behavior. Each class needs a positive case plus a causal negative on the same production path.

### V2 — Closed-world architecture dependency graph

Classify live first-party target packages bidirectionally and reject every unallowed edge class, including transport→owner, application→application, owner→owner, application→owner-private, platform→owner, foreign SQL and a hidden second Authorization evaluator. Negative fixtures mutate the real import/SQL subject and run the same architecture verifier.

### V3 — Authentication, Authorization, disclosure and CSRF

Prove one canonical ALLOW/default-DENY authority, current revocability, fail-closed domain predicates, no provider-claim Product authorization, fail-closed OIDC binding/session admission, disclosure safety and accepted CSRF enforcement.

### V4 — Transaction + required Audit atomicity

The class enumeration source is the closed T3 §15 same-local-commit Audit census, reconciled by later accepted T8-E precision. Every required census class must be represented; representative instances may be used within a class, but entire required classes may not be sampled away.

For each class:

```text
semantic mutation + every required Audit event + required durable intent when any
commit together or none commit
```

Failure injection after mutation but before required append/intent completion must leave no partial semantic commit. Offboarding multi-event teardown remains reconstructible.

### V5 — Idempotency and replay

Prove current-user + operationId + UUID scoping, normalized semantic fingerprinting, live AuthZ/disclosure before replay disclosure, exact completed ReplaySnapshot behavior, changed-request same-key rejection, no duplicate semantic mutation/Audit and 24-hour semantic expiry.

### V6 — Concurrency / ETag / narrow serialization

Prove stale whole-replacement semantics, stale DRAFT-specific semantics, protected current eligibility during correctness-critical transitions, one-open/effective uniqueness, concurrent Document-code uniqueness across distinct commands and the accepted READ COMMITTED + narrow serialization/CAS posture under real concurrent transactions.

### V7 — Exact content, malware and rendition integrity

Prove ExactContentDescriptor authority over provider identity, exact size/hash/format verification, clean-gate requirements, scan evidence binding to exact admitted bytes, same-PDF zero-transformation, DOCX conditional rendering, representative fidelity-corpus eligibility, immutable accepted rendition bytes and failure-before-success-headers for corruption.

### V8 — Durable work / River semantics

Prove at-least-once duplicate/redelivery safety for each activated named durable-effect class in the accepted T5 census, including managed-content GC when active; current-state revalidation before effect/finalization; terminal visibility/redrive; and separation between River mechanism state and Product truth.

A stale GC intent must be unable to delete current WorkingContent, immutable governed content, current claim-protected content or backup-protected content after eligibility/reference/claim state changes.

### V9 — Runtime readiness, failure isolation, resource and observability laws

Prove the accepted `/livez`/`/readyz` distinction, dependency-scoped blast radii, bounded time/memory/scratch/concurrency, no secret/PII leakage, and attributable logs/metrics/traces on material failure paths without creating Product authority.

### V10 — Backup/restore and privacy/security readiness

Prove a restorable recovery point contains the required PostgreSQL + exact-content set and that restore cannot serve until sessions/security/privacy reconciliation is complete. A causal backup-capture race must demonstrate backup pin / GC exclusion. Session resurrection, unverifiable required content or bypassed privacy/security reconciliation is a hard failure.

## 6. Forward obligations consumed by T9

T9 does not reopen the forward-obligation set wholesale.

```text
PRESERVE
  AUTH-02   external OIDC/Keycloak evidence remains replaceable mechanism evidence
  AUTH-06   no cross-system atomic IdP transaction claim
  ASY-04    one PostgreSQL-backed durable-job mechanism / River reference
  DB-01..07, DB-10   persistence/transaction properties remain validation inputs
  SEC-01    maintenance/operator/system surfaces receive no implicit serving access

REOPEN
  CNT-03    EditorSession remains unnecessary unless concrete evidence falsifies that posture
  AUD-06    retention remains outside current Golden Flow correctness absent a real requirement
  MIG-10    migration target families remain outside T9

DEFERRED
  remain counterexamples only; no validation fixture/component/endpoint may activate them by convenience
```

## 7. Coverage and future-proof closure law

T9 ratification closes the **baseline design**, not runtime proof execution. Later implementation/readiness work that claims compliance with this baseline must provide executable proof plans and then evidence satisfying:

```text
6/6 Golden Flows have positive + causal negative proof lanes
10/10 cross-cutting properties have an identified production subject and falsifier
78/78 application operations remain in the canonical wire census with no invented operation
T8-E §9.4 runtime wire-conformance classes have an executable proof plan targeting the real composed HTTP path
all T3 §15 required same-local-commit Audit classes are represented
all claim-relevant external mechanisms have a real-dependency proof lane
no MATERIAL accepted invariant is closed by mocks-only evidence
no unresolved contradiction is hidden as a test exemption/baseline waiver
```

T9 does not require all 78 operations to appear in an end-to-end Golden Flow. Runtime operation/wire conformance plus representative composed flows is the smaller and stronger proof shape.

## 8. Reopen / stop law

A future validation failure is classified before correction:

```text
implementation/proof defect
  → implementation/proof work fixes it; architecture remains closed

accepted mechanism choice falsified but semantic architecture intact
  → smallest owning T8 technical authority reopens

wire/application consumer missing or extra
  → smallest Product/T6/T8-E owner reopens; never add operation 79 silently

semantic ownership/lifecycle/AuthZ contradiction
  → smallest T1→T7 owner reopens

transition-only problem
  → route to T10; do not solve it inside this baseline
```

Preference, test convenience, framework fashion, old implementation shape and hypothetical future scale are not reopen triggers.

## 9. Preserved absences

This baseline creates no requirement for:

```text
operation 79
new Product endpoint
new Permission
new semantic owner
test-only Product state
Redis
BFF
realtime
external Search
generic event bus/workflow platform
legacy implementation restoration
```

Current progression and implementation permission remain solely in `../roadmap.md`.
