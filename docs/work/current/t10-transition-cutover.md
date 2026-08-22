# T10 — Transition / Cutover

> **TEMPORARY T10 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T10 derives the smallest truthful current→target transition and rollback-barrier contract from accepted T1→T9 authority.

T10 does **not** implement Product code, create migration tooling, add compatibility layers, define the T11 implementation graph, perform T12 readiness attack, add operations, or reopen accepted Product/T1→T9 authority by preference.

Fixed opening state:

```text
opening main                          fc7030e98021bdb55fa806df68821cf19ed1a40c
T1 → T9                               CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10                                   OPEN / ACTIVE candidate
T11 → T12                             NOT OPEN
Product implementation                BLOCKED
legacy implementation in live tree    ABSENT
application operations                78
operation 79                          ABSENT
```

## 2. Binding source truth

T7 already closed the historical-business-migration question:

```text
pre-R10 MetalDocs business history    NONE
required pre-R10 business corpus      NONE
historical business migration         NOT REQUIRED
DEV/test state preservation           REJECTED
```

Therefore T10 is a **technical activation/recovery contract**, not a business-history migration program.

Any real external DB/content/IdP/deploy estate discovered later remains **unclassified technical estate** until B0 proves whether it contains authority. DEV/test is the expected baseline from T7, but that expectation never grants disposal permission.

If concrete evidence instead proves that pre-R10 authoritative business history or a required pre-R10 corpus exists, T10 stops and routes a bounded reopen to T7/the smallest implicated owner before any migration design proceeds.

## 3. Selected approach

Three transition shapes were considered:

```text
A  one-way greenfield activation with explicit barriers   SELECTED
B  blue/green dual business authority                     REJECTED
C  migration/compatibility bridge                         REJECTED
```

B is rejected because no accepted continuity requirement justifies two systems capable of committing Product truth.

C is rejected because it would recreate import, schema-translation, legacy fallback or compatibility machinery that T7 found unnecessary for Launch.

Selected law:

```text
ONE-WAY GREENFIELD R10 ACTIVATION
+
PRIVATE PREPARATION
+
PROOF + VERIFIED CLEAN SEAL BEFORE AUTHORITATIVE BOOTSTRAP
+
FIRST POST-SEAL AUTHORITATIVE PRODUCT MUTATION = POINT OF NO RETURN
+
AUTHORITATIVE RECOVERY POINT BEFORE NORMAL SERVING
+
SERVING ACTIVATION ONLY AFTER LEGACY/DEV-TEST SERVING IS FENCED
+
ONE BUSINESS AUTHORITY AT A TIME
+
FAIL-CLOSED / R10 RECOVERY AFTER AUTHORITY BEGINS
```

Explicitly absent:

```text
historical business import
generic ETL/import framework
dual write
dual Product authority
legacy read fallback
schema/API translation compatibility layer
shadow business mutations
old/new business reconciliation
Product activation marker/table/endpoint
rollback by restoring disposable DEV/test state as Product authority
```

## 4. Five monotonic barriers

### B0 — Source truth classified

Before target preparation is treated as a cutover plan, prove:

```text
pre-R10 authoritative business history = NONE
required pre-R10 business corpus        = NONE
legacy compatibility consumer           = NONE
```

Inventory any surviving external technical estate that could affect cutover:

```text
PostgreSQL databases/schemas
managed-content buckets/objects
OIDC clients/realms/configuration
deployed application/runtime instances
DNS/ingress/origin configuration
previously published user-reachable origins/endpoints
secrets/config stores
backup/recovery artifacts
```

An item may be marked disposable only after proving it is non-authoritative DEV/test state. Absence is a valid inventory result.

**Barrier failure:** evidence of real business truth or a compatibility consumer blocks T10 and routes a bounded reopen.

### B1 — Target privately prepared

A future target realization may be provisioned privately with only the accepted T8 runtime components and mechanisms, including when actually required:

```text
one PostgreSQL Product-state database
accepted schema
one modular-monolith application runtime
River workers in-process
one active ManagedContentStore
external OIDC mechanism
private MalwareInspector
conditional private renderer
verified ephemeral exact-byte spool
accepted config/secrets/observability/network controls
```

B1 does not authorize Product implementation now; it defines the later transition requirement.

No DEV/test Product rows, governed bytes, Audit history, Releases, approvals or sessions are imported merely to populate the target.

### B2 — Target proven and clean-sealed while still non-authoritative

Before the first authoritative R10 Product mutation, the **exact deployed production candidate/profile** must have sufficient evidence that the target realization is fit to become authority. A drifted "production-equivalent" twin does not satisfy this barrier for properties that bind to the actual candidate/profile.

Required evidence includes the accepted claim-relevant gates from T8/T9, such as:

```text
schema/runtime compatibility
startup/readiness/config/secret fail-closed behavior
selected real-dependency proof lanes
security/network boundaries
backup/restore capability and isolated restore-drill path
T9 proof coverage against the real candidate/profile subject/boundary
no hidden legacy-compatibility dependency
```

T10 does not duplicate T9 proof definitions. It makes their successful realization a transition barrier.

B2 proof may legitimately create disposable Product-shaped fixture state through real production paths. Therefore B2 closes only after a **verified clean seal**:

```text
all proof-fixture Product truth is removed/reset
all disposable proof sessions are invalidated
all proof-fixture Audit/history is absent from the authority baseline
all proof-bound semantic content/River intents are absent or proven non-authoritative/reclaimable
exact production artifact/config/profile identity is recorded
mechanical clean-baseline verification succeeds
proof mutation paths are disabled/fenced before authoritative bootstrap begins
```

The clean seal is an **operations/provenance evidence artifact only**. It is not Product state, not a Product table, not an application operation, not a Permission and not semantic authority.

Any pre-B3 reset/rebuild invalidates the clean seal and all B2 evidence that depended on the reset resource. Those affected proofs must be re-established against the resulting exact candidate/profile before a new clean seal. Evidence that is structurally independent of the reset may survive only when its exact artifact/config/dependency subject is unchanged.

After a valid clean seal, any unexpected Product mutation before the intended bootstrap is an authority-boundary incident: destructive reset is no longer assumed safe until that mutation is classified.

A target that is merely deployed or merely tested is not authority-ready.

### B3 — R10 business authority begins

B3 is the **point of no return**. It occurs on the first committed authoritative R10 Product mutation **after a valid B2 clean seal**, through the accepted explicit non-serving bootstrap/administrative concern.

Examples include any committed Product truth such as:

```text
Company/User/ProviderSubjectBinding bootstrap truth
Organization/access configuration
DocumentType/governance configuration
Document/Revision/Submission state
Audit-required semantic mutation
Release/obsolescence state
```

The first post-seal bootstrap Product commit itself is the authority edge. Its canonical Product facts and required Audit evidence, where T3 requires Audit, are the durable system evidence that authority has begun. T10 creates no synthetic activation marker/table/endpoint.

B3 is not delayed until public traffic or the first controlled Document. Any authoritative Product mutation starts native R10 business history.

After B3:

```text
destructive reset to pre-R10/DEV/test authority = FORBIDDEN
discarding committed R10 business history       = FORBIDDEN
restoring disposable pre-R10 state as truth     = FORBIDDEN
```

From B3 onward, incidents are R10 recovery events even if public serving has not started yet.

### B4 — Canonical serving authority activated

Only after B3 has established the required authoritative R10 baseline may canonical production ingress/origin expose normal R10 serving.

B4 additionally requires all of the following:

```text
target remains ready against the accepted production profile
at least one complete authoritative R10 recovery point covers the current B3 baseline
that recovery point passes complete-set/manifest/ExactContentDescriptor integrity checks
all disposable DEV/test serving estates found by B0 are stopped or fenced from ordinary user requests
no previously published user-reachable DEV/test origin can still accept Product mutations
canonical OIDC/origin/ingress configuration points normal serving only at R10
```

The restore **capability** and isolated restore drill are proved under B2/T8-G. B4 does not require a fresh full restore drill of every newly captured recovery point; it requires a complete authoritative recovery point to exist before ordinary business traffic is exposed.

Activation law:

```text
one canonical production ingress/origin
→ one R10 serving system
→ no ordinary fallback or stale-path mutation against DEV/test estate
```

DNS changes alone are not fencing: stale resolver caches, direct old origins or bookmarked endpoints must be unable to accept ordinary Product mutations.

Parallel private deployment is allowed before B4; parallel **business authority** is never allowed.

After B4, traffic may be stopped or failed closed during an incident, but it may not be redirected to disposable DEV/test Product authority.

## 5. Rollback / recovery law

### Before B3 — reversible technical preparation window

Destructive technical reset is permitted only while a valid B2 clean-seal sequence has not been followed by any Product mutation whose authority status is unresolved.

Permitted pre-B3 response may include:

```text
remove/redirect non-authoritative test traffic
invalidate disposable target sessions
reset/destroy the non-authoritative target Product database
reclaim target content proven non-authoritative
correct deployment/configuration
repeat affected proof
re-establish the B2 clean seal
```

Any reset invalidates the affected B2 evidence and clean seal as defined in B2.

### After B3 — forward recovery only

Permitted response classes are:

```text
fail closed / stop normal serving
recover from a coherent R10 recovery point
restore R10 and satisfy all restore readiness barriers
apply a proven compatible R10 correction/deployment
redrive accepted durable work under canonical Product truth
```

A failure in the B3→B4 window is already an R10 recovery event. Recovery/correction must preserve the B3 authority baseline; after recovery, B4 preconditions are re-established before normal serving.

Deployment rollback is permitted only when the older deployment is compatible with already-committed R10 data/wire/runtime truth. A binary/config rollback that cannot safely consume current R10 state is forbidden.

A deployment rollback is never permission to roll business truth backward to DEV/test state.

If the canonical R10 authority **and every coherent authoritative recovery point** are lost, the system remains fail-closed. No automated re-bootstrap, reset to disposable state or silent invention of replacement business truth is permitted. That condition is catastrophic authority loss and requires an explicit operator/business recovery decision plus the smallest implicated architecture reopen before a new authority baseline can be established.

## 6. Data/content and bootstrap transition law

Because the historical-business-import branch is empty:

```text
business rows imported from pre-R10 estate   0
business Audit history reconstructed          0
pre-R10 governed content imported             0
legacy sessions preserved                     0
```

Required authoritative bootstrap data may be created only through the already-accepted **non-serving operations concern**: T3 states that bootstrap/recovery is explicit non-serving operations work and never an ordinary RBAC bypass. T10 does not invent an ordinary Product endpoint, maintenance bypass, synthetic Product state or operation 79 for bootstrap.

The T8-D `bootstrap/provisioner` database trust class is **provisioning/DDL-only** and must not be reinterpreted as semantic Product-bootstrap authority.

Exact runtime realization of semantic bootstrap belongs to later implementation planning. It must fit an accepted non-serving operations surface without widening Product semantics. If T11 implementation evidence proves the accepted T8-G runtime shells/private surfaces cannot realize bootstrap truthfully, that is a bounded T8-G reopen before implementation — never silent shell expansion or a serving-wire workaround.

Provider/object identity never becomes Product identity during technical setup.

## 7. Identity transition law

The accepted OIDC provider remains an external mechanism behind the Authentication anti-corruption seam.

T10 permits provider configuration needed for R10 login but creates no cross-system atomic cutover transaction.

```text
provider configuration
≠ Product business commit
```

Provider roles/groups/claims never become Product Authorization truth.

Pre-R10 application sessions are never imported. R10 sessions exist only against R10 authoritative Product identity/session truth.

## 8. Durable work and content boundary

There is no pre-R10 durable Product work to carry forward by default.

R10 durable work begins only from accepted R10 Product commits/intents. River state never becomes migration source or Product truth.

Any managed content created before B3 but not bound to authoritative Product truth is disposable technical state and remains subject to the B2 clean seal plus accepted content-integrity/GC laws.

After B3, governed exact content and required durable effects are R10 authority/recovery inputs. The authoritative recovery point required before B4 must cover the complete recovery set defined by T8-G, including required exact content and coherent transaction-coupled River intents.

## 9. Cleanup / technical deletion map

Cleanup is evidence-driven and monotonic.

A surviving external technical resource may be removed only after proving:

```text
not current R10 authority
not required for pre-B3 technical reversal
not required for R10 recovery/audit/provenance
contains no business truth requiring bounded reopen
```

Any business truth discovered in a resource believed disposable — whether written before or after B4 — stops cleanup and routes a bounded reopen/adjudication. Temporal labels never make business truth disposable.

Candidate cleanup classes:

```text
disposable DEV/test database/schema
disposable DEV/test managed-content objects
obsolete DEV/test deployment/runtime
obsolete test OIDC client/configuration
obsolete ingress/DNS/config/secrets
```

T10 does not require any class to exist and does not add a ceremonial waiting/observation barrier. B4 serving fencing and the evidence-driven cleanup predicates are the required safety boundaries.

No legacy compatibility code is preserved merely to make cleanup easier.

## 10. Forward-obligation disposition

T10 consumes the migration obligations as follows:

```text
MIG-05 PRESERVE
  plan/dry-run/reconciliation/idempotency remain future evidence if a real migration appears;
  not activated for Launch because historical business migration is absent.

MIG-06 PRESERVE
  CURRENT_STATE/FULL_HISTORY remain evidence-only migration modes;
  no Launch mode is selected because there is no source business corpus.

MIG-10 REOPEN
  NOT TRIGGERED;
  there are no imported target families to define.
```

All deferred import/export/repository/portability platform ideas remain DEFERRED and cannot be activated by transition convenience.

## 11. T10 falsifiers

The transition contract is false if any of these becomes true:

```text
a real authoritative business corpus exists but T10 discards it
B2 proof for a state-dependent property is reused after its subject was reset/rebuilt
proof-fixture Product truth survives the B2 clean seal into the authority baseline
any Product mutation occurs after the clean seal and destructive reset proceeds without classifying it
B3 authoritative mutation occurs before the required B2 proof + clean-seal barrier closes
two systems can commit Product truth concurrently
B4 normal serving begins before an authoritative recovery point covers the B3 baseline
any disposable DEV/test serving estate remains able to accept ordinary Product mutations after B4
normal serving begins before the required authoritative R10 baseline exists
an ordinary request can fall back or stale-route to DEV/test authority
first R10 Product mutation commits and the plan still permits destructive reset to pre-R10 truth
a post-B3 rollback deploy cannot safely consume current R10 state
pre-R10 sessions or DEV/test Audit/history are silently promoted into R10 business truth
cleanup deletes a resource containing any business truth merely because the truth is post-cutover
catastrophic loss silently re-bootstraps or promotes disposable state
operation 79 or new Product semantics are introduced solely to support cutover
```

Any failure routes to the smallest owning authority rather than being patched with compatibility infrastructure by default.

## 12. Fable Round 1 adjudication

Independent Evidence PR #159 reviewed the exact operator-approved candidate `0b90f26690b2b2bbf627f0c72283ff14c0ce9b84` and returned:

```text
VERDICT = NOT CONVERGED
MATERIAL findings = 3
Round 2 justified = YES
```

Lead adjudication:

```text
F1 MATERIAL  ACCEPT / SOLUTION REFINED
  add B2 verified-clean seal and a mechanically observable authority edge;
  reject a Product activation marker/table/endpoint;
  B3 remains the first post-seal Product bootstrap commit.

F2 MATERIAL  ACCEPT / SOLUTION REFINED
  require a complete authoritative R10 recovery point before B4;
  total loss of canonical authority + all recovery points is catastrophic authority loss,
  not ordinary re-bootstrap.

F3 MATERIAL  ACCEPT
  fence all user-reachable disposable serving estate before/at B4;
  remove the erroneous pre-R10 temporal qualifier from cleanup safety.

F4 MINOR     ACCEPT
  bind B2 evidence to the exact production candidate/profile and re-arm affected proof after resets.

F5 MINOR     PARTIAL ACCEPT
  cite T3 non-serving bootstrap concern;
  explicitly reject T8-D bootstrap/provisioner as semantic bootstrap (it is DDL/provisioning only);
  require bounded T8-G reopen if no accepted runtime surface can realize semantic bootstrap.

F6 NOTE      ALIGN WORDING
  external estate is unclassified until B0 proof; DEV/test expectation is not disposal permission.

F7 NOTE      NO CHANGE
  migration forward-obligation dispositions already conform.
```

No Product/T1→T9 authority is reopened by these candidate corrections. A possible future T8-G reopen is only a **conditional implementation falsifier** if T11 proves semantic bootstrap cannot fit an accepted non-serving runtime surface.

## 13. Handoff to later stages

T10 defines transition semantics and barriers only.

T11 must later decompose implementation work so that implementation, bootstrap, deployment and cutover can satisfy B0→B4 without weakening them.

T12 must later attack implementation readiness, including whether the actual implementation has credible evidence for:

```text
source-estate classification
exact production-candidate proof before authority begins
B2 fixture cleanup + clean-seal evidence
single-authority non-serving bootstrap
first-post-seal authoritative Product mutation discipline
a complete authoritative recovery point before serving
legacy/DEV-test serving fencing
post-B3 recovery compatibility
cleanup safety
```

T10 itself does not execute production cutover.

## 14. Candidate exit criteria

The corrected T10 candidate is ready for bounded independent confirmation only when:

```text
historical-business-import branch remains empty unless concrete evidence reopens T7
B0→B4 remain exactly five explicit monotonic barriers
B2 proof binds to the exact production candidate/profile
B2 clean seal prevents proof-fixture promotion into authority
B3 is mechanically observable as the first post-seal authoritative Product mutation without new Product state
pre-B3 technical reversal and post-B3 R10 recovery are unambiguously different
at least one complete authoritative R10 recovery point covers the B3 baseline before B4
B4 fences every inventoried user-reachable disposable serving path
cleanup cannot delete any discovered business truth based on its timestamp
no compatibility/dual-write/import platform is introduced without a named consumer
MIG-05/MIG-06 remain evidence-only and MIG-10 remains untriggered
78-operation census remains unchanged and operation 79 absent
T11/T12 remain unopened
Product implementation remains BLOCKED
required repository CI is green on the exact corrected candidate HEAD
```

Round 2, if opened, is a bounded confirmation of F1–F3 closure plus regression of the fixed five-barrier/78-operation envelope. It is not a fresh unconstrained redesign.

Independent reviewer Evidence must remain isolated and non-authoritative under Repository Standard v1.
