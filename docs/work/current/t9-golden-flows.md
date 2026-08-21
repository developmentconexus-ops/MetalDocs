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
stale real ETag
same Idempotency-Key + changed semantic request
required Audit append failure inside the real local transaction
revoked/disabled current User or grant
corrupt stored bytes against a real semantic descriptor
stopped renderer/scanner dependency
incompatible database schema
missing required secret / impossible content envelope
River redelivery/terminal failure
restore from a real backup recovery point
SIGTERM during accepted work
```

Tool identity is not authority. T8-G's current proof-backed candidates (for example Schemathesis, Testcontainers-Go, `govulncheck`, OSV-Scanner) may realize a proof lane only when they execute the required subject and falsifier.

## 3. Evidence classes

T9 uses six evidence classes. A flow/property selects only the classes required by its claim.

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

The six flows are a composition basis, not an operation-by-operation acceptance suite. Full 78-operation surface closure is proved separately by V1.

### GF1 — Identity, session, access and revocation

**Purpose:** prove the browser authentication/access chain without moving Authorization into OIDC or React.

Success lane:

```text
external OIDC authentication
→ ProviderSubjectBinding resolves exact User
→ HttpOnly ApplicationSession issued
→ synchronizer CSRF material obtained through authenticated session
→ Product read succeeds only after current User/Role/scope/domain checks
→ access administration changes current eligibility/grant
→ next protected request reflects the revocation
→ historical actor/Audit attribution remains intact
```

Required falsifiers:

```text
provider role/group claim alone cannot grant MetalDocs Permission
missing/wrong CSRF cannot reach unsafe semantic effect
revoked/disabled current access cannot remain usable through stale browser/server cache
OIDC network outage does not destroy an already-valid ApplicationSession
non-disclosable resource stays non-disclosable
```

Minimum evidence: E2 + E3 + E4; E5 for the selected OIDC profile.

### GF2 — Governance configuration to atomic Document creation

**Purpose:** prove that current administration/configuration can feed one real Document creation without duplicate truth or partial commit.

Success lane:

```text
current Area / access / DocumentType-governance configuration
→ creation options/numbering preview
→ createDocument with one logical Idempotency-Key
→ final number allocation + Document current truth + required Audit commit atomically
→ retry of the same completed logical command returns the accepted replay semantics
```

Required falsifiers:

```text
concurrent duplicate logical create cannot create two Documents or two accepted semantic effects
same key + changed semantic request is rejected
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
→ required owner mutation + Audit + durable rendition intent commit coherently
→ River redelivery-safe processing
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
```

Required falsifiers:

```text
GroupMembership drift after Step activation cannot rewrite the frozen candidate snapshot
current Authorization denial cannot be bypassed by being in the snapshot
required Audit/intent failure cannot leave a committed semantic transition
River duplicate/redelivery cannot create duplicate Release/Rendition semantic truth
renderer failure cannot convert acceptance into false success
provider byte corruption cannot produce a complete successful exact-byte response
malicious/unscanned required governed bytes cannot cross the immutable governed boundary
```

Minimum evidence: E2 + E3 + E5 + E6; E4 for the browser governance/official lenses.

### GF5 — Official discovery, obsolescence and disclosure-safe routing

**Purpose:** prove that current official/discovery views remain honest while work/obsolescence context is disclosed only when authorized.

Success lane:

```text
Library/Search official discovery
→ DocumentOfficialView
→ open_revision present only for the unique disclosable current DRAFT/SUBMITTED Revision
→ active_obsolescence_request_id present only for the unique disclosable ACTIVE request
→ create/inspect/withdraw obsolescence through admitted operations
→ History remains semantic history, not a current-resource resolver
→ Audit remains evidence, not current business truth
```

Required falsifiers:

```text
absence of open_revision/request id cannot prove non-existence to an unauthorized caller
official lens never renders DRAFT WorkingContent as official truth
Search cannot acquire a second materialized/external current-truth authority
History/Audit cannot be used to bypass current disclosure/AuthZ
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

### V1 — Product/wire census and generated projection coherence

Must prove against the canonical source:

```text
application operations   = 78
orphaned                  = 0
invented                  = 0
operation 79              = absent
OpenAPI paths/operationIds/schemas/headers/Problems match accepted wire law
generated Go/TypeScript projections compile and cannot widen closed wire vocabularies
```

A deliberately added/removed/renamed operation or incompatible schema must make the real contract proof fail.

### V2 — Closed-world architecture dependency graph

Must classify the live first-party target packages bidirectionally and reject every unallowed edge class, including transport→owner, application→application, owner→owner, application→owner-private, platform→owner, foreign SQL and hidden second Authorization evaluator.

Negative fixtures mutate the real import/SQL subject and run the same production architecture verifier.

### V3 — Authentication, Authorization, disclosure and CSRF

Must prove the canonical ALLOW/default-DENY authority is singular, current access is revocable, domain predicates fail closed, provider claims never become Product authorization, disclosure is respected, and unsafe browser requests require the accepted CSRF mechanism.

### V4 — Transaction + required Audit atomicity

For every representative transition class that requires same-local-commit evidence:

```text
semantic mutation + required Audit + required durable intent (when any)
commit together or none commit
```

Inject failure after mutation but before required append/intent completion and prove no partial semantic commit survives.

### V5 — Idempotency and replay

Must prove current-user + operationId + UUID scoping, normalized semantic fingerprinting, live AuthZ/disclosure before replay disclosure, exact completed ReplaySnapshot behavior, changed-request same-key rejection, no duplicate semantic mutation/Audit and 24h semantic expiry.

### V6 — Concurrency / ETag / narrow serialization

Must prove stale whole-replacement semantics, stale DRAFT-specific semantics, protected current eligibility during correctness-critical transitions, one-open/effective uniqueness and accepted READ COMMITTED + narrow serialization/CAS behavior under real concurrent transactions.

### V7 — Exact content, malware and rendition integrity

Must prove semantic ExactContentDescriptor authority over provider identity, exact size/hash/format verification, clean-gate requirements, same-PDF zero-transformation path, DOCX conditional-render path, immutable accepted rendition bytes and failure-before-success-headers for corruption.

### V8 — Durable work / River semantics

Must prove at-least-once duplicate/redelivery safety for each activated named durable-effect class, current-state revalidation before effect/finalization, terminal visibility/redrive and separation between River mechanism state and Product truth.

### V9 — Runtime readiness, failure isolation, resource and observability laws

Must prove accepted `/livez`/`/readyz` distinction, dependency-scoped blast radii, bounded time/memory/scratch/concurrency, no secret/PII leakage, and evidence that material failure paths emit attributable logs/metrics/traces without creating Product authority.

### V10 — Backup/restore and privacy/security readiness

Must prove a restorable recovery point contains the required PostgreSQL + exact-content set and that restore cannot serve until sessions/security/privacy reconciliation is complete. A restore that resurrects sessions or leaves required exact content unverifiable is a hard failure.

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
all claim-relevant external mechanisms have a real-dependency proof lane
no MATERIAL accepted invariant is closed by mocks-only evidence
no unresolved contradiction is hidden as a test exemption/baseline waiver
```

T9 does **not** require every one of the 78 operations to appear in an end-to-end Golden Flow. Operation-level contract/census conformance plus representative composed flows is the smaller and stronger architecture proof.

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

## 9. Candidate exit criteria

The T9 architecture candidate is ready for independent challenge only when:

```text
Golden Flow set remains exactly 6 unless a seventh protects a genuinely uncovered accepted property
cross-cutting property set remains exactly 10 unless a gap is demonstrated
no flow invents Product meaning
no proof depends on legacy implementation
no T10/T11/T12 work is embedded
implementation remains blocked
required repository CI is green on the exact candidate HEAD
```

After operator acceptance of the Lead candidate, the next proportional gate is an isolated independent Fable challenge against the exact candidate. Reviewer Evidence does not become authority.
