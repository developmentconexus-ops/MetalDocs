# R10-T8C — Internal Communication Contracts — Final Lead Adjudication

```text
FINAL LEAD ADJUDICATION
NON-AUTHORITATIVE STAGING
OPERATOR-RATIFICATION INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Round-2 reviewed HEAD:** `6c1d1929cf975a8aa72f6ebcfcd976bc7abdcf27`  
> **Round-2 publication HEAD before this adjudication:** `02e886cb41182e10228ef52e78ccf6126eeb9d18`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Stage:** T8-C ACTIVE  
> **Implementation:** BLOCKED

This artifact is the final Lead adjudication of the bounded Round-2 Fable delta review. It is the final non-authoritative operator-ratification input for T8-C. It does not promote T8-C, open T8-D or authorize implementation.

Read this artifact as a bounded amendment over:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md`

The corrected candidate remains the full T8-C contract package; this file records the final Round-2 precision closures that must be folded into durable authority if the operator ratifies T8-C.

---

## 1. Round-2 evidence result

Round-2 Fable primary verdict:

```text
APPROVE CORRECTED T8-C DELTA WITH MATERIAL FIXES

BLOCKER   0
MAJOR     5
LOW       5

SURVIVING MATERIAL CONTRADICTION  0
GLOBAL MAXIMUM CLASS              CONFIRMED
T8-B REOPEN                       NO
T1→T7 REOPEN                      NO
T8-D TRESPASS                     NO
T8-E TRESPASS                     NO
OPERATION-CENSUS DELTA            COMPLETE
ANOTHER FABLE ROUND               NO
FINAL LEAD ADJUDICATION           MAY PROCEED
```

Round 2 independently upheld both materially contested Lead positions:

```text
Round-1 B5 blocker        NOT SUSTAINED
PII-free replay selection UPHELD
```

Reviewer evidence is not authority. The dispositions below are the Lead's final technical adjudication against repository authority, primary tool/database evidence and the Method.

---

## 2. Final Global Maximum

The selected model remains unchanged:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL

SEMANTIC OWNERS
  concrete producer-owned public APIs
  one public package path per owner

TECHNICAL DEPENDENCIES
  narrow consumer-owned ports only for real consumers
  provider SDK types stay inside mechanisms

APPLICATION
  stateless choreography
  sole semantic inbound orchestration class
  cross-owner facts are gathered/mapped, never re-owned

TRANSACTION
  one caller-owned local Scope
  application owns begin/commit/rollback lifecycle
  database/sql-family substrate
  no *sql.Tx / pgx.Tx in semantic-owner signatures

AUDIT
  owner authors intrinsic evidence meaning
  application mechanically maps/routes
  Audit appends in the same Scope before commit

AUTHORIZATION
  Organization supplies current subject/scope facts
  resource owner supplies relationship/state/governance predicates
  Authorization alone returns ALLOW/default-DENY

DURABLE EFFECTS
  named current intent ports only
  River remains mechanism
  no EventBus/generic outbox

IDEMPOTENCY
  application-owned semantic fingerprint + self-contained ReplaySnapshot
  platform-owned opaque claim/replay mechanism
  PII-free durable snapshot by construction

READS
  Authorization scope prefilter
  owner canonical filtered/paginated truth
  exact current decision where required
  application composition

NO
  shared/contracts
  common/models
  generic UnitOfWork
  generic ServiceLocator
  generic Repository framework
  generic policy language
  generic DomainEvent bus
```

No Round-2 finding changes this class.

---

## 3. MAJOR-1 — idempotency isolation dependency

### Lead disposition

```text
ACCEPT FINDING WITH AUTHORITY CORRECTION
```

The reviewer correctly identified that the non-aborting `ON CONFLICT DO NOTHING` winner/loser path relies on PostgreSQL READ COMMITTED semantics.

However, this is **not a new isolation decision for T8-C**. T2 already ratified the Launch posture:

```text
PostgreSQL READ COMMITTED
+ narrow explicit lifecycle serialization
+ OCC/CAS
+ later structural constraints where required
```

Therefore final T8-C law is:

```text
D19 inherits the ratified T2 READ COMMITTED posture.

Under that posture, the idempotency mechanism must realize:
  concurrent same scoped key+same fingerprint -> serialize
  winner commit -> loser obtains completed replay without poisoning Scope
  winner rollback -> waiting contender may become claim owner
  different fingerprint -> conflict / no business mutation

T8-C does not prescribe savepoint, SQL statement or lock syntax.
T8-D must realize the law under the already-ratified T2 isolation posture.
```

If implementation evidence later proves this cannot be realized sustainably under the ratified T2 posture, reopen the implicated T2/T8-C decision deliberately. A T8-D-only election of Repeatable Read/Serializable that introduces mandatory whole-transaction retry is not a silent implementation choice because it would contradict the accepted T2 posture and D19 behavior.

Round-1 B5 remains rejected as a blocker.

---

## 4. MAJOR-2 — Scope sealing and native SQL binding

### Lead disposition

```text
ACCEPT
```

The unexported `isScope()` marker blocks from-scratch implementations outside `platform/txscope`, but Go embedding can promote the embedded interface's method set. The corrected candidate's stronger wording is therefore narrowed.

Final law:

```text
Scope has an unexported txscope-owned marker.
This prevents external first-party types from implementing Scope from scratch.
It is not claimed to make embedding impossible.

No first-party package outside platform/txscope may embed txscope.Scope.
That rule is mechanically enforced in tools/cilint with a RED/negative fixture.
```

The native binding contract is refined conceptually to:

```go
func SQLTx(scope Scope) (*sql.Tx, error)
```

Laws:

```text
only a live Scope created by the target txscope Runner is accepted
nil / foreign / embedded-wrapper / otherwise unrecognized Scope -> explicit fail-closed error
no panic-based success contract
application may not call SQLTx
semantic owners may not call SQLTx
only explicitly catalogued platform mechanisms whose external API requires *sql.Tx may call it
current named consumer = platform River adapter
```

The call-site allowlist and forbidden embedding are architecture rules with negative fixtures. This removes the current ad-hoc distributed `tx.(*sql.Tx)` downcast pattern without pretending Go provides package-set/friend visibility that it does not.

The `database/sql` family selection remains intact.

---

## 5. MAJOR-3 — GC second semantic re-proof

### Lead disposition

```text
ACCEPT
```

T4-K and T5-J require a second fail-closed proof immediately before physical deletion.

Final T5-J choreography law:

```text
PHASE 1 — semantic/mechanism eligibility transaction
  claim bounded technical candidate
  -> prove no current WorkingContent reference
  -> prove no immutable Submission/Rendition/imported governed reference
  -> prove no live admission claim/binding
  -> prove no backup pin/exclusion
  -> mark technical GC_PENDING
  -> commit

PHASE 2 — immediately before provider DeleteReclaimable
  re-read/re-prove GC_PENDING
  -> re-prove no current WorkingContent reference
  -> re-prove no immutable Submission/Rendition/imported governed reference
  -> re-prove no live admission claim/binding
  -> re-prove no backup pin/exclusion
  -> only then call provider DeleteReclaimable outside semantic transaction
  -> finalize technical mechanism state
```

The second Controlled Documents canonical-reference proof is mandatory. Safe failure remains leaked storage, never lost governed truth.

No schema, lock statement, lease table or GC_PENDING persistence realization is frozen here; those remain T8-D/T8-G concerns as already staged.

---

## 6. MAJOR-4 — T5-J application host

### Lead disposition

```text
ACCEPT ROUND-2 DETERMINATION
T8-B REOPEN = NO
```

The corrected candidate may not leave the T5-J host to a Writer.

Final host selection:

```text
internal/application/maintenance
```

Meaning:

```text
non-semantic maintenance application leaf
hosts T5-J managed-content GC reconciliation choreography
is not a product route/workspace
is not a semantic owner
is not a storage/retention domain
```

Why no T8-B reopen is required:

```text
application is already a ratified architecture class
transport/jobs -> application is already ratified
application -> semantic-owner-public is already ratified
application -> platform/txscope is already ratified
mechanism implementation is injected through a consumer-owned port, so no new application -> platform import edge is needed
no new architecture class is introduced
no owner surface changes
no dependency direction changes
no T8-B reopen trigger is met
```

T8-B's named T6-derived leaf set is treated as the inbound product-use-case sufficiency floor, not as a prohibition on a named non-semantic maintenance leaf needed by an already-ratified T5 mechanism family.

A durable T8-C promotion must make this refinement explicit so a future Writer does not route GC through `admin`, directly from `transport/jobs`, or through a new mechanism authority.

---

## 7. MAJOR-5 — one replay reconstruction rule

### Lead disposition

```text
ACCEPT
SELECT SNAPSHOT-ONLY RECONSTRUCTION
```

The final replay law is intentionally one rule:

```text
A completed durable replay response is reconstructed deterministically from the stored operation-local ReplaySnapshot version alone.

Replay MUST NOT query current mutable state to reconstruct the historical success response.
Replay MUST NOT depend on later canonical resource existence to reconstruct the historical success response.
```

Reason:

```text
resources/grants/configuration may later change or disappear
current-state re-projection could produce a different historical response
T6 requires a completed replay result sufficient for exact status/body replay
self-contained bounded snapshots eliminate another temporal dependency
```

PII law remains:

```text
ReplaySnapshot must be PII-free with respect to erasable UserProfile data by construction.
No baseline replay purge/redaction subsystem exists.
```

Free-form governance/cancellation/obsolescence/feedback text is excluded from baseline ReplaySnapshot for **snapshot minimality and duplicate-retention minimization**, not because all such text is erasable UserProfile PII.

Therefore T8-E must design replay-required POST success representations so exact status/body reconstruction does not require echoing excluded free-form text. T6 does not currently require those POST success bodies to echo that text.

If a future promoted requirement genuinely requires a replay-required success body to contain excluded free-form content or erasable PII, reopen the replay representation decision and select the smallest explicit representation/erasure mechanism then.

This preserves the PII-free Global Maximum without introducing current-state re-projection or a purge subsystem.

---

## 8. LOW-1 — database/sql selection rationale

### Lead disposition

```text
ACCEPT PRECISION
```

River is not an independent reason to prefer `database/sql`; River v0.37.1 provides multiple driver families.

Final rationale is:

```text
Go database/sql is a mature standard transaction/query primitive
current repo already proves the narrow database/sql executor property
no named Launch consumer requires a pgx-native-only capability
one transaction substrate is simpler than maintaining parallel Row/Rows abstractions
pgx may still participate beneath database/sql through a compatible driver realization
```

River/riverdatabasesql is evidence that the selected substrate serves the current durable-intent consumer, not the independent reason for selecting the substrate.

No D04 reopen.

---

## 9. LOW-2 — ManagedContent create-once law

### Lead disposition

```text
ACCEPT
```

`PresignCreate` means:

```text
create once / no overwrite
```

not merely "produce an upload URL".

Every production provider implementation must prove that property using an appropriate provider primitive. Provider-specific headers/keys remain outside the consumer contract.

---

## 10. LOW-3 — provider-directory callback semantics

### Lead disposition

```text
ACCEPT
```

The consumer-owned provider directory seam remains primitive/import-free and bounded.

Conceptual contract precision:

```go
SearchSubjects(
    ctx context.Context,
    query string,
    emit func(ref string, displayHints []string) error,
) error
```

Laws:

```text
results are bounded for the current selection journey
emit is synchronous enumeration, not a durable/streaming protocol
non-nil emit error aborts enumeration and is propagated by SearchSubjects
ref is opaque provider-subject reference
bounded displayHints are presentation hints only
provider roles/groups/permissions/claims do not cross the seam
```

No shared Provider DTO package is introduced.

---

## 11. LOW-4 — AuthorizedScopes proof obligation

### Lead disposition

```text
ACCEPT
```

Add an explicit falsifiable proof obligation:

```text
AuthorizedScopes is a prefilter over current grant/scope truth only.
It never substitutes for required exact-resource Decide/DecideMany evaluation.
```

Later proof must include negative cases showing a resource inside an authorized scope but failing its owner-authored domain predicate is still DENIED and cannot be served/mutated merely because the scope appeared in `AuthorizedScopes`.

No second Authorization evaluator or cached authority is introduced.

---

## 12. LOW-5 — VersionToken no-op result

### Lead disposition

```text
ACCEPT
```

For an exact already-current idempotent replacement that becomes an owner no-op:

```text
no duplicate semantic Audit evidence
no version advance merely for the repeat
return the current authoritative VersionToken
```

T8-E later encodes that current token as the response ETag according to the wire contract.

---

## 13. Final Round-2 delta disposition

```text
R2-1   txscope / River / database/sql      CLOSED
R2-2   Audit read                          CLOSED
R2-3   AuthorizedScopes                    CLOSED
R2-4   ManagedContent claims + GC          CLOSED
R2-5   eligibility serialization           CLOSED
R2-6   VersionToken                        CLOSED
R2-7   ProviderClient                      CLOSED
R2-8   malware exact-byte correlation      CLOSED
R2-9   idempotency concurrency             CLOSED
R2-10  PII-free ReplaySnapshot             CLOSED
R2-11  OfficialRendition content read      CLOSED
R2-12  operation-census delta              CLOSED

BLOCKER                                0
SURVIVING MATERIAL CONTRADICTION      0
THIRD FABLE ROUND                     NOT REQUIRED
```

---

## 14. Final decision additions / refinements

The corrected T8C-D01→D31 set remains selected, as amended by these final refinements:

```text
D32 SELECT D19 explicitly inherits the T2 READ COMMITTED isolation posture;
    no new T8-C isolation decision is created.

D33 SELECT Scope sealing as defense-in-depth:
    from-scratch external implementation blocked by unexported marker;
    embedding outside platform/txscope mechanically forbidden;
    SQLTx returns explicit error for any unrecognized/non-target Scope.

D34 SELECT application/maintenance as the non-semantic T5-J GC choreography host;
    no T8-B reopen/new class/new edge/new owner.

D35 SELECT two-phase GC proof with full semantic/live-reference/claim/backup
    re-proof immediately before provider deletion.

D36 SELECT self-contained snapshot-only replay reconstruction;
    no replay-time current-state/canonical re-projection.

D37 SELECT free-form replay exclusion by snapshot minimality;
    T8-E replay-required success responses must remain reconstructible without it.

D38 SELECT database/sql on standard-primitive + single-substrate + no current
    pgx-native consumer evidence; River compatibility is supporting evidence only.

D39 SELECT ManagedContent PresignCreate = create-once/no-overwrite property.

D40 SELECT bounded synchronous provider-directory callback semantics with
    propagated callback failure and bounded display hints.

D41 SELECT explicit proof that AuthorizedScopes never grants exact resource action.

D42 SELECT owner no-op replacement returns current VersionToken with no version/audit fabrication.
```

---

## 15. Stage-boundary final check

```text
T8-B reopen required?                    NO
T1→T7 reopen required?                   NO
T8-D trespass?                           NO
T8-E trespass?                           NO
new semantic owner?                      NO
new architecture class?                  NO
new generic framework?                   NO
future/Launch+ speculative contract?     NO
```

T8-C remains an internal communication-contract stage. Persistence SQL/locks/schema remain T8-D. Exact HTTP/OpenAPI/header/status/body encoding remains T8-E. Frontend/runtime/transition/execution remain later stages.

---

## 16. Verification / proof carried forward

T8-C durable promotion, if ratified, must carry at least these proof obligations forward:

```text
closed-world package/edge classification
application cannot execute Scope SQL
owner/application cannot call SQLTx
non-txscope packages cannot embed txscope.Scope
foreign/unrecognized Scope -> SQLTx fail closed
required Audit failure rolls back business mutation
Audit historical visibility filters before pagination
AuthorizedScopes cannot substitute for exact Decide/DecideMany
protected eligibility read serializes with offboarding
VersionToken stale update -> zero mutation
exact no-op replacement returns current token / no duplicate Audit
GROUP snapshot membership frozen, empty stays empty
ManagedContent create-once/no-overwrite
AdmissionClaim rollback/consume/release safety
GC performs both initial and immediate pre-delete semantic reference proofs
malware CLEAN digest matches exact admitted bytes
required OfficialRendition enqueue shares semantic transaction
idempotency concurrent loser never poisons Scope under ratified T2 READ COMMITTED posture
same-key different fingerprint never performs business mutation
completed replay is self-contained, exact, live-authorized and PII-free
OfficialRendition read never exposes provider identity
```

Exact implementation proofs remain T9/T11 as routed; T8-C records the falsifiable obligations.

---

## 17. Final gate

```text
T8-B  CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-C  ACTIVE
      Global Maximum class confirmed by independent Round 1
      Round-1 Lead adjudication complete
      corrected candidate reviewed by bounded Round 2
      Round-2 BLOCKER = 0
      Round-2 surviving material contradiction = 0
      final Lead adjudication = COMPLETE AT STAGING LEVEL
      operator ratification = NEXT

T8-D  NOT OPEN
T8-E  NOT OPEN
T8-F  NOT OPEN
T8-G  NOT OPEN
T8-H  NOT OPEN
T9→T12 NOT OPEN
implementation BLOCKED
```

No third Fable round is justified by current evidence.

### Exact next action

```text
operator reviews final T8-C package
→ explicit ratification if accepted
→ only after ratification:
     promote one consolidated durable T8-C authority into wiki/
     reconcile Decision Registry with a T8-C amendment
     update router/handoff/PR to T8-C CLOSED / T8-D ACTIVE
     clean/tombstone superseded T8-C staging as repository tooling allows
→ implementation remains BLOCKED
```

Do not promote T8-C or open T8-D without explicit operator ratification.

---

**End of final non-authoritative T8-C Lead adjudication.**