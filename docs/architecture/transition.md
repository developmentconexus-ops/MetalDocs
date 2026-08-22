---
id: transition-cutover
kind: authority
owner: architecture
summary: Ratified MetalDocs historical-migration truth plus the one-way R10 transition, authority, recovery and cutover barriers.
---

# MetalDocs transition & cutover

This authority combines the ratified T7 historical-migration truth with the ratified T10 technical transition/cutover contract. Current stage/status/implementation permission remains exclusively in `../roadmap.md`.

## 1. Binding source truth — T7

Launch source facts are:

```text
pre-R10 MetalDocs business history    NONE
required pre-R10 business corpus      NONE
historical business migration         NOT REQUIRED
DEV/test state preservation           REJECTED
```

Therefore Launch does not include historical business-document import, historical approval/Release reconstruction, actor/timestamp backfill, a generic importer/ETL framework, generic repository connector, or speculative `CURRENT_STATE` / `FULL_HISTORY` migration modes.

No DEV/test Document, Revision, Submission, governance decision, Release, Audit history, session or content becomes business history merely because it exists.

T1's imported/native provenance seam remains available for a future concrete import requirement; it is a seam, not dormant implementation.

T7 reopens only if concrete evidence appears, such as:

```text
a named pre-R10 business corpus required in MetalDocs
a contractual/regulatory preservation requirement for pre-R10 history
a production MetalDocs dataset created before cutover that becomes business-authoritative
a named merger/onboarding/import requirement requiring historical provenance
```

A hypothetical future import is not a trigger.

## 2. T10 purpose and boundary

T10 defines the smallest truthful technical transition from a non-authoritative R10 target to the single authoritative production MetalDocs.

T10 does not implement Product code, create migration tooling, add compatibility layers, define the T11 implementation graph, perform the T12 readiness attack, add application operations, or reopen T1→T9 by preference.

Fixed non-regression envelope:

```text
application operations                78
operation 79                          ABSENT
historical business migration         ABSENT
business authority                    singular
new Permission                        NONE
new semantic owner                    NONE
new Product state                     NONE
new runtime capability                NONE from T10
Product implementation                BLOCKED until roadmap authorization
```

## 3. Selected transition law

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
SERVING ACTIVATION ONLY AFTER DEV/TEST SERVING IS FENCED
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

Blue/green **deployment overlap** may exist privately where an implementation mechanism requires it; blue/green **business authority** is rejected.

## 4. Five monotonic barriers

### B0 — Source truth classified

Before treating target preparation as a cutover plan, prove:

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

An item may be marked disposable only after proving it non-authoritative. DEV/test is the expected classification from T7, never a disposal shortcut. Absence is a valid inventory result.

Evidence of real business truth or a real compatibility consumer stops T10 execution and routes a bounded reopen to the smallest owning authority.

### B1 — Target privately prepared

The future target may be provisioned privately using only already-accepted T8 mechanisms actually required by the target, including:

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

B1 does not itself authorize Product implementation.

No DEV/test Product rows, governed bytes, Audit history, Releases, approvals or sessions are imported merely to populate the target.

### B2 — Exact production candidate proven and clean-sealed while non-authoritative

Before the first authoritative R10 Product mutation, the **exact deployed production candidate/profile** must satisfy the accepted claim-relevant T8/T9 proof obligations. A drifted "production-equivalent" twin cannot substitute for candidate-bound properties.

Required evidence includes, where applicable:

```text
schema/runtime compatibility
startup/readiness/config/secret fail-closed behavior
real selected-dependency proof lanes
security/network boundaries
backup/restore capability and isolated restore-drill path
T9 proof against the real candidate/profile subject/boundary
no hidden legacy-compatibility dependency
```

B2 proof may legitimately create disposable Product-shaped fixtures through real production paths. B2 therefore closes only after a **verified clean seal** proving:

```text
all proof-fixture Product truth removed/reset
all disposable proof sessions invalidated
proof-fixture Audit/history absent from the authority baseline
proof-bound semantic content/River intents absent or proven non-authoritative/reclaimable
exact production artifact/config/profile identity recorded
proof mutation paths disabled/fenced
mechanical clean-baseline verification succeeds at seal completion after fencing
```

The clean seal is operations/provenance evidence only. It is not Product state, a Product table, an application operation, a Permission, a semantic owner or an activation marker.

Any pre-B3 reset/rebuild invalidates the clean seal and every B2 proof whose exact subject changed. Affected proof must be re-established before a new seal. Structurally independent evidence may survive only when its exact subject is unchanged.

After a valid seal, any unexpected Product mutation is an authority-boundary incident. Destructive reset is forbidden until that mutation is classified; an authoritative committed mutation means B3 has occurred.

### B3 — R10 business authority begins

B3 is the **point of no return**: the first committed authoritative R10 Product mutation after a valid B2 clean seal, through the accepted explicit non-serving bootstrap/administrative concern.

Representative Product truth that can cross B3 includes:

```text
Company/User/ProviderSubjectBinding bootstrap truth
Organization/access configuration
DocumentType/governance configuration
Document/Revision/Submission state
Audit-required semantic mutation
Release/obsolescence state
```

The first post-seal Product commit itself is the authority edge. Its canonical Product facts plus required Audit evidence, where T3 requires Audit, are the durable system evidence that R10 business history has begun.

No synthetic Product activation marker/table/endpoint exists.

After B3:

```text
destructive reset to pre-R10/DEV/test authority = FORBIDDEN
discarding committed R10 business history       = FORBIDDEN
restoring disposable pre-R10 state as truth     = FORBIDDEN
```

A B3→B4 incident is already an R10 recovery event even though ordinary serving has not started.

### B4 — Canonical serving authority activated

Ordinary production serving may begin only when all are true:

```text
B3 authoritative baseline exists
target remains ready against the accepted production profile
at least one complete authoritative R10 recovery point covers the current B3 baseline
that recovery point passes complete-set/manifest/ExactContentDescriptor integrity checks
all disposable DEV/test serving estates found by B0 are stopped or fenced from ordinary user requests
no previously published user-reachable DEV/test origin can accept Product mutations
canonical OIDC/origin/ingress configuration points normal serving only at R10
```

Restore capability and the repeatable isolated restore drill are B2/T8-G proof. B4 requires an actual complete authoritative recovery point before business traffic; it does not demand a fresh full restore drill for every newly captured point.

Activation law:

```text
one canonical production ingress/origin
→ one R10 serving system
→ no ordinary fallback or stale-path mutation against DEV/test estate
```

DNS change alone is not fencing. Stale resolver caches, direct old origins and bookmarked endpoints must be unable to accept ordinary Product mutations.

Parallel private deployment is allowed before B4; parallel business authority is never allowed.

## 5. Rollback and recovery law

### Before B3 — reversible technical preparation

Destructive technical reset is permitted only while no post-seal Product mutation has crossed or ambiguously approached the authority boundary.

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

If an authoritative Product mutation has committed, B3 has occurred and this section no longer applies.

### After B3 — forward recovery only

Permitted response classes are:

```text
fail closed / stop normal serving
recover from a coherent R10 recovery point
restore R10 and satisfy all restore readiness barriers
apply a proven compatible R10 correction/deployment
redrive accepted durable work under canonical Product truth
```

Recovery must preserve the B3 authority baseline and re-establish B4 preconditions before normal serving.

Deployment rollback is allowed only when the older deployment can safely consume the already-committed R10 data/wire/runtime truth. Deployment rollback never means business-truth rollback.

If canonical R10 authority **and every coherent authoritative recovery point** are lost, the system remains fail-closed. No automated re-bootstrap, reset to disposable state or silent invention of replacement business truth is allowed. This is **catastrophic authority loss** and requires an explicit operator/business recovery decision plus the smallest implicated architecture reopen before a new authority baseline can be established.

## 6. Bootstrap, identity, content and durable-work law

Because historical business import is empty:

```text
business rows imported from pre-R10 estate   0
business Audit history reconstructed         0
pre-R10 governed content imported            0
legacy sessions preserved                    0
```

Authoritative bootstrap uses the already-accepted T3 **non-serving operations concern**. It never becomes an ordinary Product endpoint, RBAC bypass, synthetic Product state or operation 79.

The T8-D `bootstrap/provisioner` database trust class remains provisioning/DDL-only and is not semantic Product-bootstrap authority.

Exact runtime realization of semantic bootstrap belongs to later implementation planning. If T11 evidence proves the accepted T8-G runtime shells/private surfaces cannot realize bootstrap truthfully, T8-G must be boundedly reopened before implementation; no silent shell expansion or serving-wire workaround is allowed.

OIDC provider configuration is an external mechanism change, not a Product business commit. Provider roles/groups/claims never become Product Authorization truth. Pre-R10 application sessions are never imported.

River state never becomes migration source or Product truth. Any managed content created before B3 and not bound to authoritative Product truth remains non-authoritative and is subject to the B2 clean seal plus accepted content-integrity/GC laws.

After B3, governed exact content and required durable effects are R10 recovery inputs. The authoritative recovery point required before B4 covers the complete T8-G recovery set, including required exact content and coherent transaction-coupled River intents.

## 7. Cleanup / technical deletion law

A surviving technical resource may be removed only after proving:

```text
not current R10 authority
not required for pre-B3 technical reversal
not required for R10 recovery/audit/provenance
contains no business truth requiring bounded reopen
```

Any business truth discovered in a supposedly disposable resource — whether written before or after B4 — stops cleanup and routes bounded adjudication/reopen. Temporal labels never make business truth disposable.

Candidate cleanup classes may include disposable DEV/test databases, content objects, deployments, OIDC configuration, ingress/DNS configuration and secrets, but T10 does not require any such class to exist.

No ceremonial waiting barrier is added. B4 fencing plus evidence-driven deletion predicates are the safety boundary.

## 8. Forward-obligation disposition

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

## 9. Material falsifiers

T10 is false if any of these becomes true:

```text
a real authoritative business corpus exists but T10 discards it
state-dependent B2 proof is reused after its exact subject was reset/rebuilt
proof-fixture Product truth survives the B2 clean seal
post-seal Product mutation occurs and destructive reset proceeds without classification
B3 occurs before B2 proof + clean seal closes
two systems can commit Product truth concurrently
B4 begins before an authoritative recovery point covers the current B3 baseline
a disposable DEV/test serving estate can still accept ordinary Product mutations after B4
ordinary traffic can fallback/stale-route to DEV/test authority
a post-B3 deployment rollback cannot safely consume current R10 truth
pre-R10 sessions or DEV/test Audit/history are promoted into R10 business truth
cleanup deletes any discovered business truth because of its timestamp
catastrophic authority loss silently re-bootstraps or promotes disposable state
operation 79 or new Product semantics are introduced solely to support cutover
```

A material failure routes to the smallest owning authority; compatibility infrastructure is never the automatic repair.

## 10. Ratified result and later-stage handoff

Ratified T10 shape:

```text
barriers                             exactly 5 / B0→B4
historical business migration       absent
business authority                  singular
application operations              78
operation 79                        absent
new Permission                      none
new semantic owner                  none
new Product state                   none
new runtime capability from T10     none
```

T11 may later decompose implementation, bootstrap, deployment and cutover work only in a way that preserves B0→B4. T12 may later attack whether the concrete implementation can actually satisfy source-estate classification, exact-candidate proof, clean sealing, single-authority bootstrap, authoritative recovery before serving, serving fencing, post-B3 recovery and cleanup safety.

T10 itself does not execute production cutover and does not authorize Product implementation.
