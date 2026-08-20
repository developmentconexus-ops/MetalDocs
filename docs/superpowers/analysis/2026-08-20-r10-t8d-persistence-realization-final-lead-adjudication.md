# R10-T8D — Persistence Realization — Final Lead Adjudication

```text
FINAL LEAD ADJUDICATION
NON-AUTHORITATIVE STAGING
OPERATOR-RATIFICATION INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Round-2 review HEAD:** `e63104fefd6566986c8bd6ae947e9d2ba12c5350`  
> **Original candidate:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`  
> **Round-1 review:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md`  
> **Adjudicated corrected overlay:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md`  
> **Round-2 delta review:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-corrected-candidate-fable-delta-review.md`  
> **Stage:** T8-D ACTIVE  
> **Implementation:** BLOCKED

This artifact is the final Lead confrontation/adjudication after bounded Fable Round 2. It is not durable T8-D authority. It exists only to freeze the final staging delta for explicit operator ratification.

For ratification, the effective final T8-D staging candidate is:

```text
original Global Maximum candidate
+
adjudicated corrected candidate overlay
+
this final Lead adjudication overlay
```

Where later staging conflicts with earlier staging, the later adjudication controls. Durable promotion, if explicitly authorized by the operator, must consolidate the effective result into one maintained T8-D authority rather than preserve this layered staging form as target documentation.

---

## 1. Round-2 evidence accepted

The bounded Round-2 review revalidated the exact corrected-candidate HEAD and published only its authorized review artifact. Its primary result was:

```text
APPROVE CORRECTED T8-D DELTA WITH MATERIAL FIXES

BLOCKER   0
MAJOR     7
LOW       6

Round-1 BLOCKER-1 closed          YES
Round-1 BLOCKER-2 closed          YES
Global Maximum class              CONFIRMED
upstream reopen                   NO
stage trespass                    NO
third review round required       NO
final Lead adjudication           MAY PROCEED
```

The Lead independently accepts the review's central result:

```text
both Round-1 blockers are closed
the Global Maximum class remains selected
no T1→T7 / T8-B / T8-C reopen is required
no T8-E/F/G/T10/T11 boundary needs to be crossed
all remaining findings are bounded T8-D persistence/constraint/lock/proof corrections
```

No third Fable round is selected. The Round-2 reviewer explicitly found that another full round would be ceremony unless the Lead chose a materially different realization for MJ-1 or MJ-2. The Lead chooses the smallest realizations proposed within the already-confirmed class.

---

## 2. Global Maximum — FINAL STAGING SELECTION

Unchanged:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+
DECLARATIVE CORRECTNESS
+
PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+
READ COMMITTED NARROW SERIALIZATION
+
EXPLICIT CAS
+
IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+
TRANSACTIONAL KEY↔REPLAY COMPLETION
+
THIRD-PARTY RIVER SCHEMA ISOLATION
+
PROOF-BACKED SELECTIVE LEGACY PROPERTY REUSE
-
LEGACY PHYSICAL SHAPE INHERITANCE
-
GENERIC PERSISTENCE FRAMEWORKS
-
DUPLICATE CURRENT TRUTH
```

The selected class now has zero surviving blocker and zero unresolved architecture-class contradiction.

---

## 3. MJ-1 — immutable READY descriptor proof — ACCEPT

Round 1 correctly moved malware verdict evidence out of mutable `platform.managed_content`. Round 2 correctly observed that leaving the exact descriptor on a generally UPDATE-able mechanism row still left half of the admission proof mutable.

### Final selection

`platform.managed_content` keeps only mutable mechanism lifecycle/location state required by current consumers, conceptually:

```text
id
state = OPEN | READY | GC_PENDING
provider_locator
trust_class
created_at
ready_at NULL
gc_pending_at NULL
```

The exact READY descriptor moves to an insert-only technical relation:

```text
platform.managed_content_descriptors

managed_content_id UUID PRIMARY KEY
sha256             BYTEA NOT NULL CHECK octet_length(sha256) = 32
size_bytes         BIGINT NOT NULL CHECK size_bytes >= 0
content_format     closed Launch format vocabulary NOT NULL
derived_at         TIMESTAMPTZ NOT NULL
```

Properties:

```text
one descriptor row per ManagedContent handle
SELECT + INSERT only for serving runtime
NO UPDATE / DELETE / TRUNCATE by serving runtime
descriptor is mechanism evidence of the exact create-once bytes
semantic records still copy and own their own ExactContentDescriptor truth
provider locator migration cannot rewrite descriptor facts
```

This does **not** create a second semantic descriptor authority. T4 already requires the ManagedContent mechanism to prove the exact bytes being admitted while separately requiring the semantic Submission / OfficialRendition / WorkingContent fact to own its descriptor truth. The mechanism descriptor is the immutable proof source used to establish that semantic truth.

### OPEN → READY law

Provider/network reading, hashing, structural-format validation and optional scanner work occur outside the local technical state transaction.

The local mechanism transition then:

```text
ManagedContent row FOR UPDATE
→ require current state OPEN
→ require live admission claim where the preparation path carries one
→ INSERT managed_content_descriptors exactly once
→ INSERT terminal malware_inspections evidence if a terminal inspection result exists
→ state OPEN → READY
→ ready_at = trusted time
→ commit
```

`READY` does not itself mean malware-clean. Immutable governed admission still requires applicable CLEAN evidence matching the immutable descriptor digest.

### Final malware proof relation

Retain the Round-1 correction:

```text
platform.malware_inspections

managed_content_id
sha256 digest CHECK 32 bytes
verdict = CLEAN | MALICIOUS
inspected_at

UNIQUE(managed_content_id, digest)
SELECT + INSERT only
```

For untrusted bytes:

```text
immutable governed admission requires
managed_content_descriptors.sha256
=
malware_inspections.digest
AND verdict = CLEAN
```

A contradictory second terminal verdict for the same handle+digest is not silently overwritable. A false-positive recovery uses a new create-once handle after scanner correction; it does not rewrite historical evidence.

### Disposition

```text
MJ-1 = CLOSED
```

---

## 4. MJ-2 — protected actor serialization — ACCEPT; restore the blanket rule

The narrowing introduced after Round 1 is withdrawn.

### Final law

For Launch:

> **Every authenticated user-initiated semantic product mutation obtains the actor's protected current User row through Organization `ProtectedSecuritySubjectIn`, realized with `FOR SHARE`, and holds that protection until transaction completion.**

Why this is selected:

```text
T3 §11 states an open lower bound: "applies at least to"
FOR SHARE is self-compatible among ordinary protected actions
its meaningful contention is exactly with offboarding/eligibility UPDATE
blanket coverage avoids another hand-maintained security-operation census
the rule covers all paths capable of reaching the protected semantic state
```

SYSTEM/background transitions without an authenticated human actor do not manufacture a User lock.

Access-removing technical cleanup outside a live authenticated semantic command (for example restore-time session purge) is not reclassified as a user-initiated semantic mutation merely to trigger this rule.

This restores the original D33 class with the missing justification supplied; it does not reopen T3.

### User lock ordering remains

When actor and one or more target Users participate:

```text
collect all required User ids
deduplicate
sort by stable UUID ascending
acquire each once at strongest required mode
```

Actor protection uses `FOR SHARE`; target offboarding/eligibility mutation uses update-strength locking. Same User acting as actor+target is acquired once at the stronger mode.

### Disposition

```text
MJ-2 = CLOSED
```

---

## 5. MJ-3 — serving-equivalent DB proof role — ACCEPT WITH TARGET RENAMING

The target does **not** inherit the legacy meaning of the current `metaldocs_ci` role. Current `metaldocs_ci` deliberately carries grants that differ from runtime and therefore is evidence of the failure class, not the target proof role.

### Final trust classes

Exactly four target trust classes:

```text
1. bootstrap / provisioner
   - provisioning-only identity/session
   - creates roles/schemas/extensions as required
   - the ONLY class permitted to SET ROLE metaldocs_owner
   - never serves product traffic

2. metaldocs_owner
   - NOLOGIN
   - owns first-party target schemas/tables and river.* DDL objects
   - migration/DDL authority only

3. metaldocs_runtime
   - LOGIN serving identity class
   - no table/schema ownership
   - no owner-role membership
   - only target serving privileges

4. metaldocs_verifier
   - non-owner proof/CI identity class
   - no owner-role membership
   - effective object privileges MUST equal metaldocs_runtime across the closed target DB-object catalog
```

A fifth trust class requires a named future risk reduction.

### Blocking grant parity proof

T8-D freezes the obligation:

```text
for every target schema/table/sequence/type/function/object class in the closed DB-object catalog:

effective privileges(metaldocs_verifier)
==
effective privileges(metaldocs_runtime)

mismatch => verification FAIL
```

Security tests asserting that serving cannot UPDATE/DELETE/TRUNCATE immutable classes must execute as `metaldocs_verifier`, not as superuser or owner.

The current legacy role named `metaldocs_ci` does not satisfy the target equality law and receives no name/shape survival entitlement from T8-A. Concrete current→target grant transition remains T10.

### Disposition

```text
MJ-3 = CLOSED
```

---

## 6. MJ-4 — HMAC fingerprint key rotation equality — ACCEPT

Retain:

```text
semantic_fingerprint = HMAC-SHA-256(
    server-held fingerprint key,
    canonical operation identity + complete validated semantic command
)
```

Persist:

```text
semantic_fingerprint BYTEA exact 32 bytes
fingerprint_key_version INTEGER NOT NULL CHECK fingerprint_key_version > 0
```

The HMAC key is never persisted in the product database.

### Final derivation / rotation law

T8-C fixes application-side fingerprint derivation before `BeginIn`, so an honest replay must never straddle two derivation keys while an old idempotency key is still valid.

Therefore:

```text
exactly one fingerprint-key version is ACTIVE FOR DERIVATION at a time

new version may become active for derivation ONLY AFTER
all persisted idempotency keys produced under the prior active version
have expired or been safely retired under the normal retention law

prior key material remains available until that drain completes
```

There is no dual-derivation trial, no multi-fingerprint BeginIn contract and no T8-C reopen.

T8-G owns secret provisioning/storage and operational rotation mechanics. T8-D owns this equality/compatibility constraint because otherwise identical retries could become false fingerprint conflicts.

### Disposition

```text
MJ-4 = CLOSED
```

---

## 7. MJ-5 — AdmissionClaim timing — ACCEPT; align with T4-F

The original staging statement "Reserve after proving READY" is superseded.

### Final claim lifecycle

For claim-bound preparation paths, especially browser upload:

```text
allocation
→ create ManagedContent OPEN
→ create opaque AdmissionClaim/binding in or before the same local mechanism commit
→ return handle/upload target only after the binding exists

provider writes exact create-once bytes

complete
→ require same live claim
→ derive exact descriptor / structural format
→ OPEN → READY

semantic attach
→ require same live claim
→ ManagedContent FOR SHARE + READY/descriptor revalidation
→ create/replace semantic managed_content_id reference
→ ConsumeIn = DELETE claim inside same semantic Scope
→ commit
```

A live claim therefore spans OPEN and READY until consume/release/expiry. It is **not** a READY-only reservation.

Explicit release/expiry makes an abandoned prepared handle reclaimable. Rollback of `ConsumeIn` restores the claim because the claim DELETE participates in the semantic transaction.

AdmissionClaim remains bounded, not universal as a generic owner registry. The universal attach-vs-GC safety law remains the ManagedContent `FOR SHARE` lock on every semantic reference write.

### GC statement corrected

GC eligibility proof treats **any live AdmissionClaim** as a blocker independent of the content's mechanism state. Phase-1 candidate selection itself still operates only on a reclaimable state class selected by the GC algorithm; a live claim makes the candidate ineligible.

### Upload-complete census precision

Draft upload complete now explicitly freezes:

```text
provider create-once/no-overwrite proof
same live allocation claim
ManagedContent FOR UPDATE
state OPEN
server-derived exact descriptor
structural ContentFormat validation
INSERT immutable managed_content_descriptors
INSERT terminal malware evidence if produced
state OPEN → READY
Audit: none unless upstream semantic authority requires one (mechanism transition only)
```

### Disposition

```text
MJ-5 = CLOSED
```

---

## 8. MJ-6 — Area lifecycle transaction missing — ACCEPT

Add the distinct Area lifecycle mutation family required by T6.

### Final Area locking relation

Document creation that depends on Area eligibility performs:

```text
Area row FOR SHARE
→ prove current ACTIVE
→ hold through semantic Document-create commit
```

Area lifecycle replacement/retirement performs:

```text
protected actor
→ Area row FOR UPDATE
→ expected owner VersionToken/CAS where the owning replacement contract applies
→ validate requested lifecycle transition
→ material change increments Area version
→ required Audit
→ commit
```

Consequences:

```text
create linearizes first  -> the Document creation is valid; Area may retire afterward
retire linearizes first  -> create later observes RETIRED and fails closed
no global table lock
```

Exact HTTP precondition/status mapping remains T8-E.

### Disposition

```text
MJ-6 = CLOSED
```

---

## 9. MJ-7 — GC proof lock mode / lock-order contradiction — ACCEPT

Round 1 correctly selected `platform.managed_content FOR UPDATE` as the GC serialization root. Round 2 correctly identifies that GC must not then acquire lower-class semantic locks during its proof phase.

### Final GC locking law

Both GC phase 1 and phase 2:

```text
ManagedContent row FOR UPDATE
→ hold as the SOLE GC serialization lock
→ perform ControlledDocs semantic-reference proofs as NON-LOCKING current reads
→ perform AdmissionClaim proof as NON-LOCKING current read
→ perform backup-pin proof as NON-LOCKING current read
```

GC acquires **no row lock in semantic/claim/pin classes after taking ManagedContent FOR UPDATE**.

Why those reads are stable enough under READ COMMITTED:

```text
every new semantic managed_content reference must first obtain ManagedContent FOR SHARE
→ blocked while GC owns FOR UPDATE

every backup-pin acquisition that creates new protection must first obtain ManagedContent FOR SHARE
→ blocked while GC owns FOR UPDATE

claim-bound upload allocation creates its claim while content is OPEN, before READY/GC candidacy;
semantic claim consumption is coupled to an attach that already holds ManagedContent FOR SHARE
```

Thus after GC acquires the root lock, no new protective semantic reference/pin can commit before GC re-proves and completes the local phase. GC does not need to lock those referencing rows to obtain the required safety property.

### Phase 2 remains asymmetric

Immediately before provider delete:

```text
FOR UPDATE ManagedContent
→ require still GC_PENDING
→ repeat complete semantic-reference proof
→ repeat live-claim proof
→ repeat backup-pin proof
→ commit local proof transaction
→ provider DeleteReclaimable outside semantic tx
```

Safe failure remains leaked storage, never deleted governed truth.

### Disposition

```text
MJ-7 = CLOSED
```

---

## 10. LOW L-1 — governance candidate materialization — ACCEPT

Candidate materialization is now an explicit transaction responsibility.

When a Step activates in the same Scope:

```text
NAMED_USER
→ insert exactly the configured User as governance_step_candidates(step_id,user_id)

GROUP
→ resolve current enabled members in same Scope
→ insert all returned Users as governance_step_candidates
→ empty result inserts zero rows and stays empty

→ delete any live GROUP dependency for the activated Step
→ Step becomes ACTIVE
```

This occurs for:

```text
initial Step activation during SUBMIT
next Step activation after an ACCEPT when more Steps remain
```

The `governance_decisions(step_id,actor_user_id)` FK therefore always targets the one frozen candidate authority for both selector kinds.

```text
L-1 = CLOSED
```

---

## 11. LOW L-2 — backup-pin write family — ACCEPT

`platform.managed_content_backup_pins` remains a technical mechanism relation because T4 requires backup protection to participate in reclaimability proof.

T8-D freezes the persistence/locking contract; T8-G owns when/how backup workflows invoke it.

### Acquire protection

```text
ManagedContent FOR SHARE
→ require handle still exists and is not already irreversibly deleted
→ INSERT backup pin
→ commit
```

The `FOR SHARE` requirement makes pin acquisition serialize against GC's `FOR UPDATE` root.

### Release / expiry

```text
DELETE pin on successful backup-lifecycle release or bounded expiry
```

Deleting protection cannot create a governed-byte loss by itself; GC still must perform both full proof phases before provider deletion.

Exact backup scheduling/provider snapshot mechanics remain T8-G/T10 as appropriate.

```text
L-2 = CLOSED
```

---

## 12. LOW L-3 — River object classes + SET ROLE — ACCEPT

The third-party `river.*` catalog explicitly includes object classes created by the pinned River migrations, including:

```text
tables
indexes
sequences where present
types, including river.river_job_state
functions/triggers where present for the pinned migration state
migration metadata objects
```

The target grants explicit schema/type/object privileges required by River runtime rather than relying on an undocumented assumption that PUBLIC defaults will remain the security contract.

Trust rule:

```text
bootstrap/provisioner = only class permitted to SET ROLE metaldocs_owner
metaldocs_runtime      = never
metaldocs_verifier     = never
```

```text
L-3 = CLOSED
```

---

## 13. LOW L-4 — River REINDEX reopen signal — ACCEPT

T8-D keeps self-REINDEX OFF for Launch.

The reopen is not vague. T8-G must make at least the following condition classes observable enough to support a decision:

```text
river_job index bloat / index-size growth relative to live rows
River job-fetch / availability-query latency degradation
repeated River maintenance/index-related errors if any remain
```

Exact metrics, thresholds and alert wiring belong T8-G/T9. A material measured degradation attributable to missing index maintenance is the reopen trigger for evaluating owner-run maintenance, a PostgreSQL floor change or another bounded mechanism.

```text
L-4 = CLOSED
```

---

## 14. LOW L-5 — Company singleton interlock / same-Company seam — ACCEPT PRECISION

Round-2 review sustains the Lead's rejection of Company/company_id subtraction.

Final statement:

```text
Company is a current Launch semantic scope/identity.
company_id columns carry current scope identity where the ratified semantics require it.
area_id NULL encodes Company-SCOPE KIND, not Company identity.
```

The singleton constraint has a deliberate fail-closed purpose:

> **Because Launch intentionally has no pooled-tenancy isolation substrate, `org.companies.singleton_key = 1` prevents a second Company from becoming live before an isolation-model reopen has supplied the missing security/runtime substrate.**

The propagated `company_id` relationships preserve current semantic ownership and the future reopen seam; they do not activate pooled tenancy.

Same-Company composite FKs remain structural preparation and integrity documentation. While the singleton interlock stands, their cross-Company mismatch branch is not claimed as a currently demonstrable proof obligation. On a future pooled-tenancy reopen, those constraints become active structural backstops and must be revalidated with the new isolation model.

```text
L-5 = CLOSED
```

---

## 15. LOW L-6 — validation routing + key-version type — ACCEPT

Closed-vocabulary parity remains a T8-D structural requirement:

```text
static Go/product vocabulary set
==
corresponding DDL CHECK accepted set
```

T8-D freezes the equality obligation. T9 owns the falsifiable Validation Baseline / exact control execution that proves it; T11 does not own the decision.

Implementation may generate DDL predicates from the code authority or inspect both sets, but may not create a second runtime vocabulary registry.

Idempotency key version type is now frozen as:

```text
fingerprint_key_version INTEGER NOT NULL CHECK fingerprint_key_version > 0
```

```text
L-6 = CLOSED
```

---

## 16. Final transaction/lock precision additions

The final transaction census must include, in addition to the previously corrected matrix:

```text
Area lifecycle mutation
initial governance candidate materialization
next governance candidate materialization
backup-pin acquisition/release
Draft upload allocate OPEN + AdmissionClaim
Draft upload complete OPEN→READY
```

Global lock rule remains:

```text
Idempotency claim where required
→ all User locks in sorted stable UUID order
→ Organization mutable roots (including Area)
→ DocumentType
→ Document
→ owner-local semantic child/current rows
→ ManagedContent attach lock where semantic reference participates
→ Audit / River inserts
```

Special GC exception is explicit and safe:

```text
GC does NOT follow the semantic lock chain.
GC begins at ManagedContent FOR UPDATE and performs only non-locking proofs afterward.
This is allowed precisely because every new protective reference/pin is serialized on the same ManagedContent row before it can commit.
```

No semantic operation may invert this relation by taking ManagedContent and then acquiring a Document/User/DocumentType lock. Semantic attach paths acquire their semantic roots first and ManagedContent `FOR SHARE` last; GC takes only the ManagedContent root.

---

## 17. Final proof obligations added by adjudication

Before implementation-readiness can later be claimed, the proof baseline must be able to falsify at least:

```text
READY ManagedContent cannot exist through the target write path without exactly one immutable descriptor row
serving runtime cannot UPDATE/DELETE descriptor or malware evidence rows
malware CLEAN admission must match the immutable descriptor digest

all authenticated user-initiated semantic mutations take actor FOR SHARE
actor+target User sets acquire deterministic UUID ordering

runtime and verifier effective DB grants are exactly equal over the closed target object catalog
neither runtime nor verifier can SET ROLE owner
only provisioner can reach owner role

same Idempotency-Key exact retry cannot straddle fingerprint derivation versions
new HMAC derivation version cannot activate until old-version live keys drain

AdmissionClaim is created at claim-bound allocation, not after READY
claim remains live through OPEN/READY until consume/release/expiry

Area retirement and Document create serialize on the Area row

GC holds ManagedContent FOR UPDATE and takes no lower-class locks during proof
semantic attach and backup-pin acquisition both require ManagedContent FOR SHARE

NAMED_USER and GROUP Step activation both materialize governance_step_candidates
empty GROUP snapshot still makes a Decision structurally impossible

Go/static closed vocabularies exactly equal DDL CHECK vocabularies
```

These are proof obligations, not implementation authorization.

---

## 18. Final T8-D staging decision ledger delta

Original/adjudicated decisions remain unless explicitly superseded above. Final delta:

```text
F01  SELECT immutable platform.managed_content_descriptors; remove exact descriptor mutability
     from general ManagedContent lifecycle state.

F02  RESTORE D33 blanket protected-actor FOR SHARE for every authenticated user-initiated
     semantic product mutation; retain deterministic strongest-mode User lock ordering.

F03  REPLACE target proof-role name/meaning with metaldocs_verifier and require blocking
     effective-grant equality with metaldocs_runtime across the closed DB-object catalog.

F04  SELECT one active HMAC fingerprint derivation version and drain-before-rotation law;
     fingerprint_key_version INTEGER > 0.

F05  CORRECT AdmissionClaim Reserve timing to allocation/OPEN for claim-bound preparation paths;
     claim spans OPEN→READY→consume/release/expiry.

F06  ADD Area lifecycle transaction/serialization row; Area create eligibility read uses FOR SHARE,
     Area lifecycle mutation uses FOR UPDATE.

F07  FREEZE GC proof reads as non-locking after ManagedContent FOR UPDATE; GC acquires no semantic
     row locks after its technical root.

F08  FREEZE governance candidate materialization for NAMED_USER and GROUP on every Step activation.

F09  FREEZE backup-pin acquire/release persistence family; acquire takes ManagedContent FOR SHARE.

F10  EXPAND river.* third-party object catalog/grants to TYPE and other pinned migration object
     classes; provisioner alone may SET ROLE metaldocs_owner.

F11  ADD observable River index-maintenance reopen condition classes; thresholds remain T8-G/T9.

F12  CLARIFY Company singleton as a fail-closed no-isolation interlock; keep semantic company_id;
     same-Company FKs are seam/backstop, not a currently firing cross-Company proof under singleton.

F13  ROUTE closed-vocabulary parity execution to T9 validation baseline; no generic enum framework.
```

---

## 19. Final Lead verdict

```text
ROUND-1 BLOCKERS                          CLOSED
ROUND-2 BLOCKERS                          0
ROUND-2 MAJORS                            7 / 7 ADJUDICATED + CLOSED IN FINAL STAGING
ROUND-2 LOWS                              6 / 6 ADJUDICATED + CLOSED IN FINAL STAGING
SURVIVING MATERIAL CONTRADICTION          0
GLOBAL MAXIMUM CLASS                      CONFIRMED
T1→T7 REOPEN                              NO
T8-B REOPEN                               NO
T8-C REOPEN                               NO
T8-E/F/G/T10/T11 TRESPASS                NO
THIRD FABLE ROUND                         NOT MATERIAL / NOT REQUIRED
IMPLEMENTATION                            BLOCKED
```

### Ratification gate

The Lead recommendation is:

```text
T8-D FINAL EFFECTIVE STAGING CANDIDATE
= READY FOR EXPLICIT OPERATOR RATIFICATION
```

Explicit operator ratification is still required before any of the following:

```text
promote one consolidated durable T8-D authority into wiki/
append Decision Registry with T8-D amendment
retire/tombstone live T8-D staging artifacts
mark T8-D CLOSED / OPERATOR-RATIFIED / PROMOTED
open T8-E Executable Wire Contract
```

No code, target migration, OpenAPI or implementation work is authorized by this adjudication.

---

**End of final non-authoritative T8-D Lead adjudication.**