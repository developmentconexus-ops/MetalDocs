# T10 — Transition / Cutover

> **TEMPORARY T10 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T10 derives the smallest truthful current→target transition and rollback-barrier contract from accepted T1→T9 authority.

T10 does **not** implement Product code, create migration tooling, add compatibility layers, define the T11 implementation graph, perform T12 readiness attack, add operations, or reopen accepted Product/T1→T9 authority by preference.

Fixed opening state:

```text
opening main                         fc7030e98021bdb55fa806df68821cf19ed1a40c
T1 → T9                              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10                                  OPEN / ACTIVE candidate
T11 → T12                            NOT OPEN
Product implementation                BLOCKED
legacy implementation in live tree    ABSENT
application operations                78
operation 79                          ABSENT
```

## 2. Binding source truth

T7 already closed the historical-business-migration question:

```text
current/pre-R10 MetalDocs business history      NONE
required pre-R10 business corpus                 NONE
historical business migration at Launch          NOT REQUIRED
DEV/test state preservation                      REJECTED
```

Therefore T10 is a **technical activation/recovery contract**, not a business-history migration program.

Any real external DB/content/IdP/deploy estate discovered later is classified first as technical DEV/test estate. It gains no business-authority status merely because it exists.

If concrete evidence instead proves that pre-R10 authoritative business history or a required pre-R10 corpus exists, T10 stops and routes a bounded reopen to T7/the smallest implicated owner before any migration design proceeds.

## 3. Options considered

### A — One-way greenfield activation with explicit barriers — SELECTED

```text
prepare target privately
→ prove target
→ activate one canonical serving authority
→ first authoritative R10 Product mutation
→ destructive rollback becomes forbidden
```

This is the smallest model consistent with T7 and the clean-slate T8 baseline.

### B — Blue/green dual business authority — REJECTED

Rejected because no business-history continuity consumer requires dual serving or dual mutation authority. It would introduce synchronization/divergence failure classes without protecting an accepted requirement.

### C — Migration/compatibility bridge — REJECTED

Rejected because it would recreate import, schema-translation, legacy fallback or compatibility machinery that T7 explicitly found unnecessary for Launch.

## 4. Selected transition law

```text
ONE-WAY GREENFIELD R10 ACTIVATION
+
PRIVATE PREPARATION
+
PROOF BEFORE NORMAL SERVING
+
ONE BUSINESS AUTHORITY AT A TIME
+
EXPLICIT FIRST-AUTHORITATIVE-MUTATION BARRIER
+
FAIL-CLOSED / R10 RECOVERY AFTER THAT BARRIER
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

## 5. Transition barriers

T10 freezes five monotonic barriers.

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

A future target realization may be provisioned privately with the accepted T8 runtime components and mechanisms, but it is not yet normal production authority.

Target preparation may include only already-accepted consumers/mechanisms, for example:

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

B1 does not authorize Product implementation now; it defines what later implementation must prepare.

No DEV/test Product rows, governed bytes, Audit history, Releases, approvals or sessions are imported merely to make the target look populated.

### B2 — Target proven before normal serving

Normal serving cannot begin until target evidence satisfies all accepted claim-relevant gates, including:

```text
schema/runtime compatibility
startup/readiness/config/secret fail-closed behavior
selected real-dependency proof lanes
security/network boundaries
backup/restore readiness
T9 validation obligations against the real target subject/boundary when implementation exists
no hidden legacy-compatibility dependency
```

T10 does not duplicate T9 proof definitions. It makes successful proof a cutover barrier.

A target that is only "deployed" is not cutover-ready.

### B3 — Canonical serving authority activated

After B2, production ingress/origin may switch to the R10 target.

Activation law:

```text
one canonical production ingress/origin
→ one R10 serving system
→ no ordinary fallback to a prior DEV/test implementation
```

Parallel private deployment is allowed before activation; parallel **business authority** is not.

At B3, traffic rollback remains possible only while B4 has not occurred and no authoritative Product mutation has committed.

### B4 — Business authority begun

B4 occurs on the first committed authoritative R10 Product mutation.

Examples include any committed Product truth such as:

```text
Organization/access configuration
DocumentType/governance configuration
Document/Revision/Submission state
Audit-required semantic mutation
Release/obsolescence state
```

B4 is not delayed until the first controlled Document exists. Any authoritative Product mutation starts native R10 business history.

After B4:

```text
destructive rollback to pre-R10/DEV/test authority = FORBIDDEN
resetting/discarding committed R10 business history = FORBIDDEN
restoring disposable pre-R10 state as Product truth = FORBIDDEN
```

Incidents after B4 are R10 recovery events, not cutover rollback.

## 6. Rollback law

### Before B4 — reversible activation window

If B0→B3 were satisfied but no authoritative Product mutation has committed, recovery may include:

```text
remove/redirect production traffic
invalidate target sessions
reset or destroy the target Product database
reclaim target managed content proven non-authoritative
correct deployment/configuration
repeat proof and activation
```

This is safe only because no R10 business history exists yet.

### After B4 — forward recovery only

Permitted response classes are:

```text
fail closed / stop normal serving
recover from a coherent R10 recovery point
restore R10 and satisfy all restore readiness barriers
apply a proven compatible R10 correction/deployment
redrive accepted durable work under canonical Product truth
```

The system must never "rollback" by making disposable DEV/test state authoritative again.

Deployment rollback is permitted only when it is compatible with already-committed R10 data/wire/runtime truth. A binary rollback that cannot safely read/write the current R10 state is forbidden.

## 7. Data/content transition law

Because the historical-business-import branch is empty:

```text
business rows imported from pre-R10 estate   0
business Audit history reconstructed          0
pre-R10 governed content imported             0
legacy sessions preserved                     0
```

Required target bootstrap data may be created only through explicit accepted bootstrap/administrative mechanisms defined by later implementation planning; T10 does not invent ordinary Product endpoints or maintenance bypasses for bootstrap.

Provider/object identity never becomes Product identity during technical setup.

## 8. Identity transition law

The accepted OIDC provider remains an external mechanism behind the Authentication anti-corruption seam.

T10 permits target-side provider configuration needed for R10 login, but does not create a cross-system atomic cutover transaction.

```text
provider/config change
≠ Product business commit
```

Normal serving cannot rely on provider roles/groups/claims as Product Authorization truth.

Pre-R10 application sessions are never imported. Any target sessions are R10 sessions and must obey current restore/session invalidation laws.

## 9. Durable work and content boundary

At activation there may be no pre-R10 durable Product work to carry forward.

R10 durable work begins only from accepted R10 Product commits/intents. River state never becomes a migration source or Product truth.

Any target managed content created before B4 but not bound to authoritative Product truth is disposable technical state and must be reclaimable under the accepted content-integrity/GC laws.

After B4, governed exact content is R10 authoritative state and follows R10 backup/restore/recovery laws.

## 10. Cleanup / legacy technical deletion map

Cleanup is evidence-driven and monotonic.

A surviving external technical resource may be removed only after proving:

```text
not current R10 authority
not required for rollback before B4
not required for recovery/audit/provenance
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

T10 does not require any of these classes to exist.

No legacy compatibility code is preserved merely to make cleanup easier.

## 11. Forward-obligation disposition

T10 consumes the stage-relevant migration obligations as follows.

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

All deferred import/export/repository/portability platform ideas remain DEFERRED and cannot be activated by T10 convenience.

## 12. T10 falsifiers

The transition contract is false if any of these becomes true:

```text
a real pre-R10 authoritative business corpus exists but T10 discards it
normal serving begins before required target proof gates close
two systems can commit Product truth concurrently
an ordinary request can fall back to DEV/test authority
first R10 Product mutation commits and the plan still permits destructive reset to pre-R10 truth
a post-B4 rollback deploy cannot safely consume current R10 state
pre-R10 sessions or DEV/test Audit/history are silently promoted into R10 business truth
cleanup can delete a resource without proving it non-authoritative/non-required
operation 79 or new Product semantics are introduced solely to support cutover
```

Any such failure routes to the smallest owning authority rather than being patched with compatibility infrastructure by default.

## 13. Verification / handoff to later stages

T10 defines transition semantics and barriers only.

T11 must later decompose implementation work so that the transition can be executed without weakening B0→B4.

T12 must later attack implementation readiness, including whether the actual implementation has credible evidence for:

```text
source-estate classification
target proof before traffic
single-authority activation
first-authoritative-mutation detection/operational discipline
post-B4 recovery compatibility
cleanup safety
```

T10 itself does not execute production cutover.

## 14. Candidate exit criteria

The T10 candidate is ready for independent challenge only when:

```text
historical-business-import branch remains empty unless concrete evidence reopens T7
B0→B4 are explicit and monotonic
pre-B4 rollback and post-B4 recovery are unambiguously different
single Product authority is preserved throughout
no compatibility/dual-write/import platform is introduced without a named consumer
MIG-05/MIG-06 remain evidence-only and MIG-10 remains untriggered
78-operation census remains unchanged and operation 79 absent
T11/T12 remain unopened
Product implementation remains BLOCKED
required repository CI is green on the exact candidate HEAD
```

Independent reviewer Evidence, if authorized later, must remain isolated and non-authoritative under Repository Standard v1.
