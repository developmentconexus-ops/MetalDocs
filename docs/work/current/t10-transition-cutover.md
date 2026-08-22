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

Any real external DB/content/IdP/deploy estate discovered later is classified first as technical DEV/test estate. It gains no business-authority status merely because it exists.

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
PROOF BEFORE AUTHORITATIVE BOOTSTRAP
+
FIRST-AUTHORITATIVE-MUTATION = POINT OF NO RETURN
+
SERVING ACTIVATION ONLY AFTER AUTHORITATIVE BASELINE EXISTS
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
secrets/config stores
backup/recovery artifacts
```

An item may be marked disposable only after proving it is non-authoritative DEV/test state. Absence is a valid inventory result.

**Barrier failure:** evidence of real pre-R10 business truth or a compatibility consumer blocks T10 and routes a bounded reopen.

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

### B2 — Target proven while still non-authoritative

Before the first authoritative R10 Product mutation, the production candidate/profile must have sufficient evidence that the target realization is fit to become authority.

Required evidence includes the accepted claim-relevant gates from T8/T9, such as:

```text
schema/runtime compatibility
startup/readiness/config/secret fail-closed behavior
selected real-dependency proof lanes
security/network boundaries
backup/restore capability
T9 proof coverage on production-equivalent real mechanisms/boundaries
no hidden legacy-compatibility dependency
```

T10 does not duplicate T9 proof definitions. It makes their successful realization a transition barrier.

Proof fixtures used before authority begins are disposable and cannot be promoted as business history. The authoritative production Product database/content baseline is clean of DEV/test business state before B3.

A target that is merely deployed is not authority-ready.

### B3 — R10 business authority begins

B3 is the **point of no return**. It occurs on the first committed authoritative R10 Product mutation, including non-serving bootstrap/administrative creation of the minimum Product truth required for Launch.

Examples include any committed Product truth such as:

```text
Company/User/ProviderSubjectBinding bootstrap truth
Organization/access configuration
DocumentType/governance configuration
Document/Revision/Submission state
Audit-required semantic mutation
Release/obsolescence state
```

B3 is not delayed until public traffic or the first controlled Document. Any authoritative Product mutation starts native R10 business history.

After B3:

```text
destructive reset to pre-R10/DEV/test authority = FORBIDDEN
discarding committed R10 business history       = FORBIDDEN
restoring disposable pre-R10 state as truth     = FORBIDDEN
```

From B3 onward, incidents are R10 recovery events even if public serving has not started yet.

### B4 — Canonical serving authority activated

Only after B3 has established the required authoritative R10 baseline and the target remains ready may canonical production ingress/origin expose normal R10 serving.

Activation law:

```text
one canonical production ingress/origin
→ one R10 serving system
→ no ordinary fallback to a prior DEV/test implementation
```

Parallel private deployment is allowed before B4; parallel **business authority** is never allowed.

After B4, traffic may be stopped or failed closed during an incident, but it may not be redirected to disposable DEV/test Product authority.

## 5. Rollback / recovery law

### Before B3 — reversible technical preparation window

While no authoritative R10 Product mutation has committed, recovery may include:

```text
remove/redirect non-authoritative test traffic
invalidate disposable target sessions
reset/destroy the non-authoritative target Product database
reclaim target content proven non-authoritative
correct deployment/configuration
repeat proof
```

This is safe only because R10 business history has not begun.

### After B3 — forward recovery only

Permitted response classes are:

```text
fail closed / stop normal serving
recover from a coherent R10 recovery point
restore R10 and satisfy all restore readiness barriers
apply a proven compatible R10 correction/deployment
redrive accepted durable work under canonical Product truth
```

Deployment rollback is permitted only when the older deployment is compatible with already-committed R10 data/wire/runtime truth. A binary rollback that cannot safely consume current R10 state is forbidden.

A deployment rollback is never permission to roll business truth backward to DEV/test state.

## 6. Data/content transition law

Because the historical-business-import branch is empty:

```text
business rows imported from pre-R10 estate   0
business Audit history reconstructed          0
pre-R10 governed content imported             0
legacy sessions preserved                     0
```

Required authoritative bootstrap data may be created only through explicit accepted non-serving bootstrap/administrative mechanisms defined by later implementation planning. T10 does not invent an ordinary Product endpoint, maintenance bypass, or synthetic Product state for bootstrap.

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

Any managed content created before B3 but not bound to authoritative Product truth is disposable technical state and remains subject to accepted content-integrity/GC laws.

After B3, governed exact content and required durable effects follow R10 backup/restore/recovery laws.

## 9. Cleanup / technical deletion map

Cleanup is evidence-driven and monotonic.

A surviving external technical resource may be removed only after proving:

```text
not current R10 authority
not required for pre-B3 technical reversal
not required for R10 recovery/audit/provenance
contains no pre-R10 business truth requiring bounded reopen
```

Candidate cleanup classes:

```text
disposable DEV/test database/schema
disposable DEV/test managed-content objects
obsolete DEV/test deployment/runtime
obsolete test OIDC client/configuration
obsolete ingress/DNS/config/secrets
```

T10 does not require any class to exist.

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
a real pre-R10 authoritative business corpus exists but T10 discards it
B3 authoritative mutation occurs before the target has passed the required B2 proof barrier
two systems can commit Product truth concurrently
normal serving begins before the required authoritative R10 baseline exists
an ordinary request can fall back to DEV/test authority
first R10 Product mutation commits and the plan still permits destructive reset to pre-R10 truth
a post-B3 rollback deploy cannot safely consume current R10 state
pre-R10 sessions or DEV/test Audit/history are silently promoted into R10 business truth
cleanup can delete a resource without proving it non-authoritative/non-required
operation 79 or new Product semantics are introduced solely to support cutover
```

Any failure routes to the smallest owning authority rather than being patched with compatibility infrastructure by default.

## 12. Handoff to later stages

T10 defines transition semantics and barriers only.

T11 must later decompose implementation work so that implementation, bootstrap, deployment and cutover can satisfy B0→B4 without weakening them.

T12 must later attack implementation readiness, including whether the actual implementation has credible evidence for:

```text
source-estate classification
production-equivalent target proof before authority begins
single-authority bootstrap
first-authoritative-mutation operational discipline
serving activation after authoritative baseline
post-B3 recovery compatibility
cleanup safety
```

T10 itself does not execute production cutover.

## 13. Candidate exit criteria

The T10 candidate is ready for independent challenge only when:

```text
historical-business-import branch remains empty unless concrete evidence reopens T7
B0→B4 are explicit and monotonic
B3 is unambiguously the first authoritative R10 Product mutation / point of no return
pre-B3 technical reversal and post-B3 R10 recovery are unambiguously different
B4 serving activation cannot create a second Product authority or fallback path
no compatibility/dual-write/import platform is introduced without a named consumer
MIG-05/MIG-06 remain evidence-only and MIG-10 remains untriggered
78-operation census remains unchanged and operation 79 absent
T11/T12 remain unopened
Product implementation remains BLOCKED
required repository CI is green on the exact candidate HEAD
```

Independent reviewer Evidence, if authorized later, must remain isolated and non-authoritative under Repository Standard v1.
