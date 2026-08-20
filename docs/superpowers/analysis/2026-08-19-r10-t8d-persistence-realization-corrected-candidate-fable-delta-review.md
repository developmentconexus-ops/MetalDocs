# R10-T8D — Persistence Realization — Corrected Candidate Fable Delta Review

```text
BOUNDED ROUND-2 DELTA REVIEW EVIDENCE
NON-AUTHORITATIVE
NOT TARGET AUTHORITY
NOT OPERATOR RATIFICATION
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Review class:** Standard Fable review, bounded Round 2 (`developmentconexus-ops/conexus-methodology/README.md`)
> **Reviewed artifact:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md`
> **Effective input:** original candidate + adjudicated corrected overlay; overlay controls on conflict
> **Stage:** T8-D ACTIVE — bounded corrected-delta gate
> **Implementation:** BLOCKED

This artifact is reviewer evidence. Findings are evidence until final Lead adjudication and explicit operator ratification. This review creates no authority and changes no upstream decision.

---

## 1. Revalidated remote HEAD

The handoff-supplied HEAD was **not** trusted as input. Revalidated independently at review start:

```text
git fetch origin --prune
  From https://github.com/developmentconexus-ops/MetalDocs
     9f830850..825fb643  docs/a8-authz-approval-redesign-ledger -> origin/...

git ls-remote origin docs/a8-authz-approval-redesign-ledger
  825fb643224f7f9e621dfb058417a2baaaf79b75  refs/heads/docs/a8-authz-approval-redesign-ledger

gh pr view 131 --json number,state,headRefName,headRefOid,baseRefName,isDraft,mergeable
  number       131
  state        OPEN
  isDraft      true
  headRefName  docs/a8-authz-approval-redesign-ledger
  headRefOid   825fb643224f7f9e621dfb058417a2baaaf79b75
  baseRefName  main
  mergeable    MERGEABLE
```

```text
REVIEWED REMOTE HEAD = 825fb643224f7f9e621dfb058417a2baaaf79b75
```

The independently revalidated branch HEAD and PR head OID **agree with each other and with the handoff claim**. Review executed in a detached worktree at that exact commit.

Publication delta from the Round-1 reviewed HEAD (`9f830850`) is four commits, all documentation routing/staging:

```text
1ad8de73  materialize adjudicated T8-D corrected candidate
eb9390f2  route T8-D to bounded round 2
fae85d46  route T8-D to bounded Fable round 2
825fb643  hand off T8-D bounded round 2
```

No product code, schema, migration, grant, contract or upstream durable authority changed between Round 1 and this review.

---

## 2. Authority reconstruction

Reconstructed from repository files in the repository-mandated order. This prompt, chat context, Lead reasoning, the candidate's own summaries and the Round-1 reviewer's conclusions were treated as **non-authority**.

```text
AGENTS.md                                                       routing/bootstrap only
docs/engineering/standards/root-cause-global-maximum-method.md  Method v1.0.0 ACCEPTED
                                                                verified byte-identical to
                                                                conexus-methodology/METHOD.md
conexus-methodology/README.md                                   Standard Fable review workflow
wiki/references/current-agent-handoff.md                        T8-D ACTIVE / bounded Round 2 next
wiki/architecture/r10-technical-architecture.md                 sole stage/status/next-action router
wiki/architecture/r10-t2-governance-effectivity-transactions.md candidate-snapshot law
wiki/architecture/r10-t3-authorization-audit-enforcement.md     §10/§11 eligibility serialization,
                                                                §16 bounded Audit facts
wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
                                                                T4-E/F/G/H/K admission + GC
wiki/architecture/r10-t6-canonical-api-frontend-journeys.md     complete mutation-route census
wiki/architecture/r10-t8c-internal-communication-contracts.md   §15 AdmissionClaims, §18 Idempotency,
                                                                §20 GC choreography, §21 in/out of Scope
docs/superpowers/analysis/...-t8d-...-bootstrap.md              staging router
docs/superpowers/analysis/...-t8d-...-global-maximum-candidate.md        original candidate
docs/superpowers/analysis/...-t8d-...-independent-fable-review.md        Round-1 evidence
docs/superpowers/analysis/...-t8d-...-adjudicated-corrected-candidate.md ROUND-2 TARGET
db/grants/0000_identity_roles.sql                               current DB trust evidence
db/grants/0001_role_grants.sql                                  current grant-surface evidence
internal/platform/jobs/river/client.go                          current River client evidence
internal/platform/bootstrap/jobs.go                             current River migration ownership
GOMODCACHE river@v0.37.1 + riverdriver/riverdatabasesql@v0.37.1 exact pinned dependency source
```

Current code, schema and grants were used **only** as evidence for concrete reuse/feasibility claims, never as target authority.

---

## 3. Primary verdict

```text
APPROVE CORRECTED T8-D DELTA WITH MATERIAL FIXES
```

Both Round-1 blockers are closed by the selected corrections. The Global Maximum class survives unchanged and no alternative class was found. Seven MAJOR findings are bounded, in-class and closable by the Lead without redesign; none is a new BLOCKER and none reopens an upstream stage.

---

## 4. Finding counts

```text
BLOCKER   0
MAJOR     7
LOW       6

Round-1 BLOCKER-1 closed             YES
Round-1 BLOCKER-2 closed             YES
Global Maximum class                 CONFIRMED
surviving material contradictions    3   (MJ-2, MJ-5, MJ-7)
upstream reopen                      NO
stage trespass                       NO
another review round required        NO
final Lead adjudication may proceed  YES
```

---

## 5. Round-1 blocker closure

### BLOCKER-1 — CLOSED

The corrected selection resolves the D26 + D30 + D04 unsatisfiability without weakening the DB trust boundary. Detail in §6.

### BLOCKER-2 — CLOSED

The universal attach-side `FOR SHARE` law closes the constructible lost-governed-bytes race on **every** semantic reference path. The path census was verified independently against the schema rather than against the prose — it is exhaustive. Detail in §7.

Two residues sit adjacent to the closed blockers and are reported as MAJOR rather than as blocker reopenings:

```text
MJ-1  the M2 malware correction is under-closed by a mutable digest (adjacent to M2, not B2)
MJ-7  GC-side proof lock mode is unstated and §18/§20 acquisition order is in tension
```

---

## 6. R2-1 — River / PostgreSQL 16 / DB privileges

Verified against the **exact pinned** dependency in `GOMODCACHE`, not against current upstream documentation.

### 6.1 Can River v0.37.1 run correctly without its self-reindexer?

**YES — confirmed from source.**

```text
river@v0.37.1/client.go:288-291
  ReindexerIndexNames []string
  "customizes which indexes River periodically reindexes.
   If nil, River uses [ReindexerIndexNamesDefault]. If non-nil, the provided ..."

river@v0.37.1/client.go:409-412
  reindexerIndexNames := ReindexerIndexNamesDefault()
  if c.ReindexerIndexNames != nil {
      reindexerIndexNames = make([]string, len(c.ReindexerIndexNames))
      copy(reindexerIndexNames, c.ReindexerIndexNames)
  }
  -> a NON-NIL EMPTY slice is honoured verbatim; it is not coerced back to the default

river@v0.37.1/internal/maintenance/reindexer.go:39-42
  // IndexNames is the exact list of indexes to reindex on each run. It must
  // be non-nil. An empty slice disables reindex work.

reindexer.go:141-152
  the run loop calls exec.IndexesExist(IndexNames: ...) and reindexes each returned name
  -> with an empty list the loop issues NO REINDEX at all
```

So a **public, documented, version-pinned** configuration field disables the behavior. Deferring the exact wiring to T8-G is therefore safe: T8-D is not deferring to a mechanism that might not exist.

The un-disabled failure mode is also confirmed non-fatal: `reindexer.go` logs `Error reindexing` and `continue`s. River does not crash; it accrues a permanently failing nightly job — exactly the outcome Round 1 predicted.

### 6.2 Is any required runtime River operation impossible under non-owner grants?

**NO.** Every other River runtime path is DML:

```text
job_cleaner.go            DELETE river_job
job_rescuer.go            UPDATE river_job
job_scheduler.go          UPDATE river_job
queue_cleaner.go          DELETE river_queue
periodic_job_enqueuer.go  INSERT river_job
leader election           INSERT/UPDATE/DELETE river_leader
InsertTx                  INSERT river_job
NotifyMany                pg_notify()  — EXECUTE granted to PUBLIC by default
```

`riverdriver/riverdatabasesql@v0.37.1/river_database_sql_driver.go:90-91` reports `SupportsListener() == false`, so the database/sql driver never issues `LISTEN`; it polls. No privileged path there either.

The only DDL lives in `riverdriver/riverdatabasesql@v0.37.1/migration/main/*.up.sql`, which runs through the owner-capable migrator path.

### 6.3 Does any hidden River maintenance path still require ownership?

**NO, with one grant-surface omission (LOW).** Objects created by the pinned migrations:

```text
CREATE TABLE   river_migration / river_job / river_queue / river_leader
CREATE TYPE    river_job_state AS ENUM(...)        <- 002_initial_schema.up.sql:1
CREATE INDEX   river_job_* (several)
CREATE TRIGGER river_notify + function river_job_notify (002) — both DROPPED again in 004
no CREATE SCHEMA anywhere
```

The absence of `CREATE SCHEMA` independently confirms Round 1: the `river` schema must be provisioned by MetalDocs, which the corrected §3.2 now states. Confirmed.

The corrected §3.1 enumerates the runtime grant surface as *"USAGE/DML/sequence/function privileges"*. It omits **TYPE**: `river.river_job_state` is an ENUM the runtime must use. PostgreSQL grants type `USAGE` to `PUBLIC` by default, so this is harmless today — but a stage that freezes a *closed* DB-object catalog should name the object class rather than rely on an unstated default. → **L-3**.

### 6.4 Is disabling reindex the smallest solution vs weakening the trust boundary?

**YES — and the repository has already paid for the alternative.**

Round-1 option (a) (serving role owns `river.*`) collides with a hard constraint the estate documents in its own words:

```text
db/grants/0000_identity_roles.sql:29-35
  "HARD CONSTRAINT: the serving identity (metaldocs_runtime) must never own a
   table. Postgres RLS does not apply to the table owner unless FORCE ROW
   LEVEL SECURITY is set, so ownership on any non-FORCE table would be a
   silent full RLS bypass ... This file only ever grants metaldocs_runtime DML
   privileges ... it never makes it an owner of anything."
```

A PostgreSQL table owner also holds implicit `DROP`/`ALTER`/`TRUNCATE` on its objects. Making serving an owner of *anything* converts a single structural rule ("serving owns nothing") into a policed exception list. Disabling one optional background maintenance service is strictly smaller and touches no first-party object. **The Lead selected the correct option.**

### 6.5 Correctly deferred to T8-G?

**YES.** T8-D decides *that* self-REINDEX is off (a privilege/correctness decision) and records the PG16 `MAINTAIN`-absence consequence; T8-G wires the pinned field. That is the right seam.

One Method gap: the reopen trigger — *"If later operating evidence proves River index maintenance is materially needed"* — names no observable. Method §3: *"A control that cannot be shown to fire is not proven."* A reopen trigger with no detectable signal cannot fire. → **L-4**.

### 6.6 Current-state corroboration

The corrected posture is a **preserved proven property**, not new invention:

```text
internal/platform/bootstrap/jobs.go:99-116
  MigrateRiverSchema ... "Owned exclusively by the metaldocs-dbprovision one-shot binary
  (A6.1 re-cut, issue #88 ...)" — app binaries never migrate River
internal/platform/config/jobs.go:25   METALDOCS_JOBS_RIVER_SCHEMA already exists
internal/platform/jobs/river/client.go:53-62
  river.Config{...} sets Queues/Schema/retention but NOT ReindexerIndexNames
```

That last line is the live instance of BLOCKER-1: today's client runs the default reindexer as the non-owner `metaldocs_runtime`. Reported as evidence that the defect class is real and currently reachable — **not** as a T10 transition instruction.

```text
R2-1 VERDICT   BLOCKER-1 CLOSED. River/PG16 privilege model COHERENT.
               Findings: L-3, L-4.
```

---

## 7. R2-2 — ManagedContent attach vs GC

### 7.1 Is the attach path census exhaustive?

The overlay's prose list was not accepted as the census. Every semantic `managed_content_id` column in the candidate schema:

```text
candidate line 745   controlled_docs.working_contents.managed_content_id     NOT NULL FK
candidate line 778   controlled_docs.submissions.managed_content_id          NOT NULL FK
candidate line 995   controlled_docs.official_renditions.managed_content_id  NOT NULL FK
candidate line 1105  platform.admission_claims.managed_content_id            mechanism, not semantic
candidate line 1129  platform.managed_content_backup_pins.managed_content_id mechanism, not semantic
```

`controlled_docs.document_origins` (candidate lines 655-668) carries `source_sha256/size/format` **only** — no handle — so template provenance is not an attach path.

Write paths that create or replace those three semantic columns:

```text
working_contents INSERT      Document create               §27
working_contents INSERT      Next Revision                 §27
working_contents UPDATE      DRAFT PATCH source replace    §27
submissions INSERT           SUBMIT                        §27
official_renditions INSERT   OfficialRendition finalization §27
```

The corrected §4 list — *Document / next-Revision seed, DRAFT source replacement, Submission admission, OfficialRendition admission* — **covers all five, exactly**. The trailing *"other T4 semantic-reference creation already ratified for Launch"* is a harmless catch-all; nothing at Launch is left for it to catch.

### 7.2 Can any attach-vs-GC race still delete governed bytes?

Attempted construction. **No surviving race.**

```text
attach holds FOR SHARE through commit (PostgreSQL row locks are held to end of transaction)
GC phase 1 and phase 2 both take FOR UPDATE
FOR SHARE x FOR UPDATE = conflicting  ->  total order on the same technical row

GC first:      attach's own locked re-read observes GC_PENDING -> fail closed
attach first:  GC blocks; on acquiring, its mandatory repeated semantic-reference proof
               sees the committed reference -> abort deletion
```

Secondary checks:

```text
multi-handle attach transactions?    none exist — every attach path writes exactly one handle
replacement path (DRAFT PATCH)?      FOR SHARE on the NEW handle; the released old handle
                                     becoming GC-eligible is the correct outcome
FK KEY SHARE timing hole (Round-1)?  closed — the lock is now taken at proof time, not at INSERT
concurrent descriptor UPDATE?        a plain UPDATE takes FOR NO KEY UPDATE, which conflicts with
                                     FOR SHARE — attach also blocks provider-locator migration
                                     for its (short) duration. Desirable.
backup-pin INSERT?                   not a semantic reference; a pin can only matter for content
                                     already semantically referenced, in which case GC aborts on
                                     the reference proof. No governed loss.
```

### 7.3 AdmissionClaim non-universality

Keeping the claim non-universal is correct: T4-F ties the binding to *"knowing a handle UUID never authorizes attaching it"*, which is a threat only for externally allocated handles. Making it universal to close a lock-ordering race would have been mechanism inflation. The Lead chose the smaller instrument.

However the overlay's own new census row and the unchanged D21/D22 disagree about **when** a claim exists — see **MJ-5** (§12.3).

```text
R2-2 VERDICT   BLOCKER-2 CLOSED. Attach census independently verified EXHAUSTIVE.
               Residues: MJ-5 (claim reserve timing), MJ-7 (GC-side lock mode).
```

---

## 8. R2-3 — Governance frozen-candidate structural barrier

```text
governance_decisions(step_id, actor_user_id)
  FK -> governance_step_candidates(step_id, user_id)     PK(step_id,user_id) — valid FK target
```

| Attack | Result |
|---|---|
| NAMED_USER steps | Overlay §5 upgrades the original candidate's hedged *"the same table **may** hold the resolved candidate for NAMED_USER Steps"* (candidate line 909) into *"remains the exact frozen active-candidate authority for NAMED_USER and GROUP steps"*. The hedge is removed. **Closed** — but the activation write is now mandatory and appears in no census row → **L-1**. |
| GROUP steps | Activation resolves enabled members, inserts frozen candidates, deletes the live dependency; the FK binds the decider to that snapshot. Matches T2 §6 line 174 verbatim. **Closed** |
| Empty GROUP snapshot | Zero candidate rows ⇒ no `GovernanceDecision` satisfies the FK ⇒ structurally undecidable. Matches T2's ratified bounded escape (withdraw → fix route → resubmit). **Closed structurally**, not by application check. |
| Later GroupMembership drift | Candidates are insert-only post-activation and the FK points at them, so drift cannot retroactively enfranchise or disenfranchise a decider. Matches T2 lines 167/500. **Closed** |
| Does the FK replace Authorization? | **No.** The FK proves *"was in this Step's activation snapshot"*. T3 `Authorization` still proves *"may act now"*, and SoD still applies. The overlay states the separation explicitly. Two different facts, two mechanisms, no duplicate authority. |
| Deletion / cascade coupling | `governance_decisions` and `governance_step_candidates` are both insert-only; candidates FK `org.users(id)` and Users are never hard-deleted (offboarding disables; lawful erasure deletes `user_profiles`, not `org.users`). No cascade hazard. |

This remains, as Round 1 said, a net complexity reduction: one FK replaces an application invariant on the highest-value governance path.

```text
R2-3 VERDICT   M1 CLOSED. Finding: L-1.
```

---

## 9. R2-4 — Immutable malware evidence

### 9.1 What the correction gets right

```text
mechanism-only?        YES — 6 columns, no state machine, no owner, no business aggregate
insert-only?           YES — SELECT + INSERT; no UPDATE/DELETE/TRUNCATE
one terminal verdict?  YES — UNIQUE(managed_content_id, digest) makes a contradictory second
                       verdict for the same exact bytes UNCOMMITTABLE
exact 32-byte digest?  YES — CHECK(octet_length(digest)=32)
security subsystem?    NO
```

Materially stronger than mutable columns: a `MALICIOUS` verdict on exact bytes becomes permanent and cannot be overwritten by the same role that consumes it.

### 9.2 False-positive recovery vs T4 create-once — coherent

`MALICIOUS` on `(H, D)` makes that handle non-admissible forever; recovery is a **new handle**. T4-H already makes a new handle the normal path (*"DRAFT replacement → new handle"*). The recovery story therefore uses create-once rather than fighting it. Recovery from a genuine scanner false positive still requires a scanner-side fix — which is correct: the database must not be where security verdicts are argued with.

### 9.3 Can CLEAN still be forged through another mutable path? — **YES**

This is the one place the accepted correction does not deliver the invariant it claims.

The admission gate is stated as:

```text
inspection.verdict = 'CLEAN'
AND inspection.digest = semantic exact SHA-256
```

The right-hand side is `platform.managed_content.sha256` (or the copy propagated from it into `working_contents.sha256`). Both are **runtime-UPDATE-able**:

```text
platform.managed_content is explicitly mechanism state with a lifecycle (D20), NOT in D31's
immutable-history class list. Runtime must hold UPDATE on it for OPEN->READY, ready_at,
gc_pending_at and provider-locator migration. There is no column-level split, and a
write-once column cannot be expressed by a column-level GRANT.
```

Constructible bypass using only privileges the corrected design grants:

```text
UPDATE platform.managed_content SET sha256 = <digest of the bytes to admit> WHERE id = H;
INSERT INTO platform.malware_inspections (managed_content_id, digest, verdict, inspected_at)
       VALUES (H, <that same digest>, 'CLEAN', now());       -- INSERT is granted
-- no UNIQUE(managed_content_id, digest) conflict, because the digest is new
-- gate: verdict CLEAN AND inspection.digest = managed_content.sha256  ->  PASSES
```

This is Round-1 M2's own threat model (*"a defect or compromise on any managed_content UPDATE path"*) surviving in the half of the comparison the correction did not move. The correction closed the verdict and left the referent open, so net protection is bounded by the mutability of the descriptor.

In bounded-Round-2 scope because the corrected delta **asserts a closure it does not achieve**, not because the underlying column is a Round-1-unchanged decision.

**Smallest sufficient fix — same shape as the accepted M2 fix, no new authority:** make the READY exact-descriptor facts (`sha256`, `size_bytes`, `content_format`) insert-only, written exactly once at OPEN→READY — e.g. `platform.managed_content_descriptors(managed_content_id PK, sha256, size_bytes, content_format, derived_at)` with `SELECT + INSERT` only — leaving `platform.managed_content` holding only the mutable mechanism lifecycle columns. It keys identically to `platform.malware_inspections`, adds no lifecycle, and makes D31's privilege-enforced-immutability claim true for the whole admission proof rather than half of it. → **MJ-1**.

```text
R2-4 VERDICT   M2 PARTIALLY CLOSED — mechanism-only and immutable, but the proof referent
               remains mutable by the consuming role. Finding: MJ-1.
```

---

## 10. R2-5 — ProviderSubjectBinding uniqueness

```text
UNIQUE(issuer, subject) WHERE replaced_at IS NULL
UNIQUE(user_id)         WHERE replaced_at IS NULL
no provider_subject_identities table
```

| Attack | Result |
|---|---|
| Current double binding | Impossible. Two current rows for one `(issuer,subject)` violate the first partial unique; two current rows for one User violate the second. **Closed** |
| Admin mis-binding recovery | An A → S mis-bind is corrected by replacing A's binding (`replaced_at` set), after which B → S is admissible. M3's permanent-lockout sequence no longer terminates in a dead end. **Closed** |
| Historical record truthful | The table already carries `bound_at` **and** `replaced_at` (candidate lines 356-368), so every historical row is a closed interval `[bound_at, replaced_at)` and the current row is `[bound_at, ∞)`. The overlay's claim that history can record *"A was bound during [t1,t2); B became current at t3"* is structurally supported. **Closed** |
| Legitimate revert (IdP rollback) | Admissible: a previously used subject may become current again for its original User. **Closed** |
| Resurrection of a replaced row | The only mutable field is `replaced_at`. Setting it back to NULL while another current row exists is blocked by the partial unique index itself. Where no current row exists, resurrection *is* the legitimate revert. The constraint is self-protecting — a genuine structural strength worth recording. |
| Ambiguity requiring a stronger constraint | **No.** T3 correlates Audit by `actor_user_id`, never by subject (T3 §16 bounded facts carry no subject). Reassignment invalidates sessions and bumps version per the retained replacement law. No correlation is lost and no ambiguity arises that global uniqueness would have prevented. |

Rejecting the second table was correct: it would have created a permanent record of the *first* association — the same unrecoverable state under a new name, at the cost of a new persistent concept.

```text
R2-5 VERDICT   M3 CLOSED with the Lead's narrowing. No findings.
```

---

## 11. R2-6 — User lock ordering

Corrected rule: collect → dedupe → sort UUID ascending → acquire once at the strongest required mode.

| Attack | Result |
|---|---|
| Two admins offboarding each other | A and B are one sorted set in both transactions; both acquire `min(A,B)` first at update strength. One blocks, no cycle. **M4's constructible 40P01 is eliminated.** |
| Eligibility replacement | Same lock class, same rule. **Closed** |
| Responsible-owner assignment | §27 takes actor + target User before `Document FOR UPDATE` — class 2 before class 5, consistent with §18. **Closed** |
| Direct RoleAssignment | actor + target User in class 2. **Closed** |
| GroupMembership add | actor + target User in class 2. **Closed** |
| Session issuance | single User. **Closed** |
| Same User as actor and target | explicitly "acquire once at strongest required mode". **Closed** |
| Cycle with Document locks | No transaction in the entire census locks **two** `controlled_docs.documents` rows. Template-seeded create locks the *source* Document and INSERTs the new one (nothing to lock). So no intra-class Document cycle exists, and UUID ordering would cover one if it appeared. **Closed** |
| Cycle with DocumentType locks | DocumentType (class 4) is always acquired before Document (class 5) in SUBMIT and obsolescence; configuration replacement takes DocumentType alone. **Closed** |
| Cycle between attach and GC | **NOT closed — MJ-7 below.** |

### MJ-7 — GC-side proof lock mode is unstated, and §18 contradicts §20

§18 places `platform.managed_content` at class **7**, after `Document` at class **5**. But GC's own sequence (§20 / D22) acquires `managed_content FOR UPDATE` **first** and only then performs its ControlledDocs semantic-reference proofs — class 7 before class 5/6.

That inversion is safe **if and only if** GC's reference proofs are non-locking reads. Under READ COMMITTED they can be: the `FOR SHARE` / `FOR UPDATE` conflict on the single technical root provides the serialization, so GC needs no lock on the referencing rows at all.

The corrected candidate never says this. And B2 has just established the idiom *"prove under `FOR SHARE`"* on exactly these rows, which makes the wrong instinct the natural one. An implementer who applies the new idiom to GC's proof produces:

```text
attach:  Document FOR UPDATE (5)        ->  managed_content FOR SHARE (7)
GC:      managed_content FOR UPDATE (7) ->  Document/child FOR SHARE (5)
         = architecture-induced 40P01 cycle on the highest-value path
```

Same defect class as M4 — an ordering the design forbids in one place and invites in another — at the stage that freezes lock behavior.

**Smallest fix:** one sentence in §18/§20 — *"GC's ControlledDocs / AdmissionClaim / backup-pin proofs are non-locking reads; the `managed_content` row lock is the sole serialization root for GC, and GC acquires no lock in classes 1-6."* → **MJ-7**.

```text
R2-6 VERDICT   M4 CLOSED for the User lock class. Finding: MJ-7.
```

---

## 12. R2-7 — Transaction census completeness

### 12.1 Added rows verified

```text
Area create            present  — Idempotency-Key per T6 §18, Organization insert + Audit + Replay
Group create           present  — same shape
RoleAssignment revoke  present  — natural DELETE idempotency, no durable key. Correct: T6 lists
                                  DELETE /api/v1/role-assignments/{assignment_id} with no key
Company replacement    present  — VersionToken CAS + protected actor + Audit
Draft upload allocate  present  — OPEN handle, explicitly "no governed semantic attachment yet"
Draft upload complete  present  — see 12.2
Session logout/revoke  present  — DELETE current ApplicationSession, natural idempotency
```

### 12.2 Upload completion — does it carry everything required?

```text
descriptor derivation  YES  "derive exact descriptor + actual format"                    (T4-E)
format validation      YES  "structural validation"                                      (T4-E)
malware evidence       YES  "insert immutable malware inspection evidence when an
                            inspection reaches a terminal verdict"                       (T4-G)
OPEN->READY semantics  YES  "local mechanism transaction serializes ManagedContent state
                            and persists READY exact facts"
create-once            NOT NAMED in the census row. The original candidate places it at
                       "provider primitive/conformance proof plus target write path"
                       (candidate line 1089). The one T6 route where create-once is actually
                       proven is this one, and the new row omits it.        -> folded into MJ-5
lock mode              NOT NAMED. The row says "serializes ManagedContent state" without a
                       primitive. An OPEN handle is not GC-eligible (GC phase 1 requires READY),
                       so this is not a correctness hole — but this is the stage that freezes
                       lock mapping.                                        -> folded into MJ-5
```

### 12.3 MJ-5 — AdmissionClaim reserve timing is now self-contradictory

The new census row places the claim at allocation:

```text
overlay §9:  "Draft upload allocate / OPEN ... create ManagedContent OPEN + AdmissionClaim/binding
              as required"
```

The unchanged original text places Reserve after READY:

```text
candidate D21:  "Reserve = INSERT after proving handle READY"
candidate D22:  "New AdmissionClaim reservation only succeeds while handle is READY, so no new
                 in-flight attachment can begin after GC_PENDING commits"   <- a load-bearing GC
                 argument that depends on the READY precondition
```

Upstream is unambiguous and agrees with the **new** row, not with D21:

```text
T4-F (lines 226-231):  "managed-content allocation -> server binds handle to intended
                        operation/root through an opaque unforgeable binding/claim"
T4-F (line 235):       "A live admission claim/binding reserved for an in-flight attachment
                        protects that READY handle from GC eligibility ..."
T4 header (line 5):    the 2026-08-18 bounded amendment is titled, in part,
                       "admission-claim GC liveness"
```

Consequence of the D21 reading: the handle is `READY` — hence GC-eligible — **before** any claim exists. GC phase 1 can legitimately mark a just-completed, not-yet-attached upload `GC_PENDING`. The failure is fail-closed rather than data-losing, but T4-F's ratified protection is simply not realized.

The overlay's ledger (§17) records D21 as *"CORRECTED: not universal; attach FOR SHARE is universal"* — it does **not** record the timing reversal. A promoted authority built from this ledger would carry the wrong precondition and D22's stale justification.

**Smallest fix:** correct D21 to *"Reserve = INSERT at allocation, in or before the transaction that commits `OPEN`"*; restate D22's GC argument as *"a live claim blocks phase-1 eligibility regardless of state"*; and add create-once plus the OPEN→READY lock primitive to the upload-complete row. → **MJ-5**.

### 12.4 Still-missing mutation families

The census was re-derived from T6's own route list rather than from the overlay's additions list.

**MJ-6 — one T6 mutation route is still absent:**

```text
T6 line 1331:  PUT /api/v1/areas/{area_id}/lifecycle
```

Distinct from `PUT /api/v1/areas/{area_id}`, which the census covers as *"Area/Group replacement"* (and which the candidate explicitly limits: *"Area code is not mutable"*). Area lifecycle is a **state** transition standing in a direct serialization relationship with an operation already in the census:

```text
§27 Document create:  "protect/revalidate active Area"
```

If Area retirement does not take a matching lock, a Document can be created against an Area retired concurrently — a constructible interleaving whose resolution T8-D has not decided. By the Lead's own accepted M5 standard (*"an operation absent from §27 is an undecided decision, not an implementation detail"*), this is a MAJOR-class omission, and it is a miss **inside** the accepted M5 correction. → **MJ-6**.

**L-1 — governance candidate materialization has no census row.** The M1 FK makes candidate rows mandatory for every decided Step, including NAMED_USER. §27 SUBMIT (which activates the first Step) says *"freeze route/steps"* and *"create live GROUP dependency rows for unactivated GROUP steps"* — never *"insert frozen candidates for the activated Step"*. Same for the ACCEPT-with-more-steps branch of `Governance decision`, where only the GROUP resolution is described. One line each.

**L-2 — `platform.managed_content_backup_pins` has no write path anywhere.** The table exists, GC phases 1 and 2 both prove against it, and T4-L makes pin correctness a restore-integrity property — but no transaction in the candidate creates, releases or expires a pin. T8-D should either place the pin write family in the census or state explicitly that pin lifecycle is deferred to T8-G, so that "undecided" is not the residual state.

Everything else reconciles. A route-by-route check against T6 lines 1309-1402 maps the remaining 31 mutation routes and the 3 non-route mutation families (session issuance, OfficialRendition finalization, GC) to a census row. `PUT /api/v1/users/{user_id}/provider-binding` is decided in the candidate's §7 rather than §27 — decided, so not a gap.

```text
R2-7 VERDICT   M5 SUBSTANTIALLY CLOSED but still incomplete.
               Findings: MJ-5, MJ-6, L-1, L-2.
```

---

## 13. R2-8 — GovernanceAttempt uniqueness

```text
UNIQUE(submission_id)
UNIQUE(obsolescence_request_id)
```

The overlay's PostgreSQL reasoning is correct: default `NULLS DISTINCT` means obsolescence attempts (all with `submission_id IS NULL`) do not collide, and vice versa, so no partial predicate is needed. The existing XOR CHECK guarantees exactly one non-null subject.

| Attack | Result |
|---|---|
| RETURN_FOR_CHANGES | attempt → RETURNED (terminal); Revision SUBMITTED→DRAFT, `current_submission_id=NULL`. Resubmission INSERTs a **new** immutable Submission → new subject → new attempt. Unique holds. |
| Withdraw | attempt → WITHDRAWN; same resubmission path. Holds. |
| Cancel | attempt → CANCELLED; Revision → CANCELLED. Holds. |
| Resubmit | Cannot reuse a Submission — every SUBMIT inserts an immutable Submission row. Holds. |
| NoHumanApproval SUBMIT | Creates no attempt at all; the unique is vacuous, not violated. |
| Obsolescence withdrawal → new request | New `ObsolescenceRequest` → new subject → new attempt. Holds. |
| Two attempts racing on one Submission | Blocked by the unique index rather than by an application check — which is the point. |

Consistent with T2 §7-§8: an attempt is terminal on RETURN/WITHDRAW/CANCEL and retry requires a new Submission, so exactly one attempt per governed subject exists for all time.

```text
R2-8 VERDICT   M6 CLOSED. No findings.
```

---

## 14. R2-9 — Lead rejection of Round-1 M7

Evaluated independently, assuming neither the reviewer nor the Lead was correct.

### 14.1 The Round-1 redundancy argument is technically wrong on its two sharpest examples

Round 1 claimed `authz.role_assignments.company_id` is *"redundant with `area_id NULL = Company scope`"* and that `audit.events.company_id` is *"redundant with the same `area_id NULL` convention"*.

Reading the actual column semantics:

```text
candidate lines 534-537   role_assignments:
                            area_id NULL     = Company scope
                            area_id non-null = Area scope
candidate lines 1315-1318 audit.events:
                            area_id NULL     = Company historical visibility attribution
                            area_id non-null = exact historical Area attribution
```

`area_id IS NULL` encodes the **scope kind**. It does not encode the **scope identity**. For an Area-scoped row the Company is recoverable through `areas.company_id`; for a **Company-scoped row `area_id` is NULL and `company_id` is the only carrier of which Company the grant or the event belongs to.** The two columns hold different facts, so there is no duplicate current truth. Constancy under a singleton is not duplication — a constant fact still needs exactly one home.

The same test applied to the remaining columns: `org.users`, `org.areas`, `org.groups`, `controlled_docs.document_types` and `controlled_docs.documents` carry `company_id` because their ratified uniqueness rules are stated as *within Company* (T2/T6), and `org.group_memberships` carries it to make the same-Company law structural. None is a second copy of a fact stored elsewhere.

**The Lead's REJECT is sustained on the merits, and on stronger grounds than the Lead stated.** No `company_id` subtraction is warranted; there is no *"only SOME columns are redundant"* subset to name.

### 14.2 The internal-consistency charge was never actually answered — and has a good answer

Round 1's real charge was not only redundancy: it was that `CHECK(singleton_key = 1)` plus full propagation is *"building the substrate while forbidding its use"*, and it offered (a) drop the columns or (b) drop the CHECK. The Lead took neither and re-argued semantics instead, so the structural objection stands formally unanswered in the corrected staging.

It has an answer, and it is the stronger reading of the current design:

> With **no RLS, no tenant GUC and no isolation substrate**, a second `org.companies` row would be catastrophic — every `company_id` predicate would become load-bearing for isolation on a system that has no isolation enforcement. `CHECK(singleton_key = 1)` is therefore not a constraint that forbids the seam's use; it is the **fail-closed interlock that makes shipping without an isolation substrate safe**, and the thing a future pooled-tenancy reopen must delete *deliberately and together with* the substrate it then owes.

Recording that turns two decisions that look contradictory into one coherent pair, and it is exactly the reopen discipline D07 already implies.

### 14.3 Residual: some corrected controls cannot fire

§16.5 states the same-Company composite FKs *"MUST structurally prove `membership.company_id = group.company_id = user.company_id`"*. While the singleton CHECK stands, that equality **cannot fail and cannot be falsified by any admissible insert**. Method §3: *"A control counts only when its firing can be demonstrated or credibly falsified."*

The constraint is not wrong and should be kept — it is the seam. But it should be recorded as **structural preparation whose firing becomes demonstrable only after a pooled-tenancy reopen**, rather than listed among T8-D's current proofs, so §33's proof obligations do not claim a control that cannot fire. → **L-5** (together with §14.2's interlock rationale).

```text
R2-9 VERDICT   M7 Lead REJECTION SUSTAINED — and Round-1's redundancy reasoning is refuted,
               not merely overruled. No company_id subtraction. Finding: L-5.
```

---

## 15. R2-10 — Go ↔ DDL closed-vocabulary parity

```text
static owner/product vocabulary set == corresponding DDL-accepted vocabulary set
any addition/removal/mismatch -> verification FAIL
```

| Attack | Result |
|---|---|
| Is parity/generation enough? | Yes, and it is demonstrably firable: add a value on either side and the check fails. That satisfies Method §3 where the original §33 had no control at all. |
| Does any listed vocabulary not need a DB CHECK? | No — every listed class is genuinely closed and at least one (`role_code`) is load-bearing for authorization per D06. Open vocabularies that correctly carry **no** CHECK (`operation_code`, `resource_kind`, `system_actor_code`) are properly excluded. |
| Does it reintroduce a generic enumeration framework? | **No.** It is a verification obligation, not a runtime mechanism — no registry, no reflection, no shared enum abstraction, no PostgreSQL ENUM. The rejected alternative (PG ENUM) is still rejected for the right reason: same hand-sync coupling minus the type. |
| Is T11 correctly left free? | Freezing the obligation and leaving generation-vs-inspection open is the right split. But a **blocking verification control** sits closer to T9's falsifiable Validation Baseline than to T11's execution graph; §12 routes it to T11 only. Minor routing imprecision → **L-6**. |

```text
R2-10 VERDICT   M8 CLOSED. Finding: L-6 (routing only).
```

---

## 16. R2-11 — Idempotency corrections

### 16.1 Retention order — correct

```text
BEGIN
DELETE expired platform.idempotency_replays
DELETE matching platform.idempotency_keys
COMMIT
```

Verified against the retained FK design: `replays.key_id → keys.id` is immediate, so deleting the child first violates nothing; the reverse `keys.id → replays.key_id` is `DEFERRABLE INITIALLY DEFERRED` and is satisfied at COMMIT once the key is gone. Deleting the key first fires the immediate FK; deleting the replay alone fails at COMMIT. Round-1's PostgreSQL 16.14 evidence is consistent with the retained design. Keeping the asymmetry — rather than deferring both FKs for order-independence — is right: the asymmetry is what makes an orphaned key structurally uncommittable.

### 16.2 Exact-replay completion invariant — preserved

Unchanged by the HMAC correction. `successful idempotent semantic fact commit ⇔ completed ReplaySnapshot commit` still rests on the deferred completion FK, and D25's winner/loser realization is untouched.

### 16.3 Full semantic command equality — preserved

Rejecting Round-1's option (b) (excluding erasable fields) was correct: it would have made materially different `POST /users` commands compare **equal** under one Idempotency-Key, converting a privacy improvement into a correctness defect. Option (a) was the right pick.

### 16.4 Offline dictionary attack — closed

`HMAC-SHA-256(server-held key, canonical identity ‖ canonical validated command)` with *"Do not persist the HMAC key in the product database"* defeats offline brute force against a database-only compromise, which is the threat M10 actually described. `fingerprint_key_version` is a version ordinal and is PII-free. `ReplaySnapshot` remains independently PII-free by construction per T8-C §18.4. All confirmed.

### 16.5 MJ-4 — the key-version dimension has no home in the equality contract

T8-C §18.1 fixes the signature:

```text
BeginIn(scope, actor, operationID, key, semanticFingerprint)
```

The **application derives** the fingerprint and passes it in; the **platform compares** opaque bytes (T8-C §18.2). The HMAC correction adds a version dimension this contract cannot carry:

```text
t0   key row created; fingerprint stored under version v1
t1   fingerprint key rotates to v2
t2   the SAME client retries the SAME operation with the SAME Idempotency-Key
     application derives with the CURRENT version v2  (it cannot know the stored version —
     BeginIn is the first call, so there is nothing to read first)
     platform compares HMAC_v2(cmd) against stored HMAC_v1(cmd)  ->  MISMATCH
     T8-C law: "same key + different fingerprint -> conflict; no business mutation"
     ->  an honest exact retry is refused as a semantic conflict
```

The corrected compatibility law — *"key material needed to compare fingerprints for non-expired idempotency keys must remain available until those keys expire"* — guarantees only that the key still **exists**. It does not say which version **derivation** uses, and under T8-C's contract shape the deriver cannot select the stored one. The law is stated one layer away from where the defect lives, and the failure lands on precisely the property idempotency exists to protect.

**Smallest fix, entirely inside T8-D, with no T8-C change** — freeze the derivation rule as well as the availability rule:

```text
exactly one fingerprint key version is active for derivation at any time;
rotation DRAINS: a new version becomes the derivation version only after every idempotency key
issued under the prior version has expired or been retired;
the prior version's material is retained until that drain completes.
```

`fingerprint_key_version` then stays meaningful — it proves which key produced a stored row and lets the drain be verified — and no honest retry can straddle two versions. Rotation mechanics and secret provisioning stay in T8-G; only the equality law is T8-D's. → **MJ-4**.

### 16.6 Boundary and compatibility

```text
T8-D/T8-G boundary        CORRECT as split (persisted columns + equality law here; provisioning
                          and rotation mechanics there) — but incomplete until MJ-4 closes the
                          derivation half of the equality law
compatibility requirement INSUFFICIENT as written — see MJ-4
fingerprint_key_version   no declared type/CHECK where every comparable bounded column in the
                          candidate carries one                                     -> L-6
```

```text
R2-11 VERDICT   M9 CLOSED. M10 privacy CLOSED, equality-under-rotation OPEN.
                Findings: MJ-4, L-6.
```

---

## 17. R2-12 — DB trust classes

### 17.1 Classes — correct and evidence-backed

The four classes match a model the repository already proved and documented:

```text
db/grants/0000_identity_roles.sql:13-35
  1. bootstrap superuser  — creates roles/extensions, provisioning only
  2. metaldocs_owner      — NOSUPERUSER NOBYPASSRLS NOLOGIN, owns schema, runs all DDL,
                            "only ever reached via SET ROLE metaldocs_owner from the bootstrap
                             superuser session inside metaldocs-dbprovision"
  3. metaldocs_runtime    — DML only; "the serving identity must never own a table"
  +  metaldocs_ci         — NOSUPERUSER NOBYPASSRLS LOGIN non-owner test role
```

```text
unnecessary role proliferation?    NO — four classes, each with a named risk reduction, and the
                                   overlay explicitly forbids more without one
missing trust class?               NO — provisioning, DDL ownership, serving and proof are all
                                   covered; no read-only/reporting identity is required at Launch
can serving inherit owner powers?  NO — "never receives owner membership / never SET ROLEs into
                                   metaldocs_owner" closes the exact BLOCKER-1 workaround
provisioning vs T8-G topology?     CORRECT — classes and privilege relationships are T8-D;
                                   credential/process placement is deferred
```

One precision gap: the overlay says runtime never `SET ROLE`s into owner, but never says **which** class may. Since owner is NOLOGIN, DDL is unreachable unless exactly one class holds that right. → **L-3** (with the missing TYPE class).

### 17.2 MJ-3 — the proof role's serving-equivalence is asserted, never proven, and the named role currently diverges by design

§15 requires the immutable-history proofs to run as `metaldocs_ci` *"or an equivalent non-owner serving-equivalent role with the same intended grants"*, including:

```text
runtime cannot UPDATE/DELETE/TRUNCATE AuditEvent
```

The only role in the estate called `metaldocs_ci` is **deliberately granted exactly those privileges**:

```text
db/grants/0001_role_grants.sql:141-147
  "Grant surface mirrors metaldocs_ci's DML posture ... but, unlike metaldocs_ci, also inherits
   the audit_events and outbox_events hardening below ... metaldocs_ci is a test role that
   deliberately keeps full audit_events DML for audit-chain test assertions."

db/grants/0001_role_grants.sql:205-215
  REVOKE UPDATE, DELETE, TRUNCATE ON TABLE metaldocs.audit_events FROM metaldocs_runtime
  -- metaldocs_ci is deliberately NOT revoked
```

So the role T8-D names as the proof role is, in the current estate, the one role for which the flagship proof is **false by design**. Run as written, the proof either fails or — worse, if someone "fixes" it by relaxing the assertion — produces a green that means nothing.

The deeper issue is structural, not naming: **M11 replaced one unproven assumption (that the proof connects as a non-owner) with another (that the proof role's grants equal the serving role's).** The repository has already demonstrated that these two grant sets diverge, deliberately, for a real reason. An unverified equivalence assumption sitting under every immutable-history proof is the same false-green class M11 exists to eliminate.

**Smallest fix, in-class with the M8 obligation already accepted:** require a **blocking grant-set equality proof** — compare the effective privileges of the proof role and the serving role across the closed DB-object catalog and fail verification on any difference — and either name a T8-D proof role that is not the existing test role, or state that the existing `metaldocs_ci` divergence must be removed as part of the target. → **MJ-3**.

```text
R2-12 VERDICT   M11 classes CLOSED; the proof mechanism is NOT yet sound.
                Findings: MJ-3, L-3.
```

---

## 18. R2-13 — Accepted LOW corrections

| Item | Verdict |
|---|---|
| **16.1** ReplaySnapshot 64 KiB removed; exact max deferred to T8-E | **CORRECT.** T8-D keeps the *obligation* (bounded + PII-free) and defers only the *value* to the stage that owns the success-representation census — a declared DEFER SAFELY with a named later owner, not an abdication, and not a T8-E trespass in either direction. The separately proof-backed Audit 64 KiB bound (candidate line 1310) correctly stays. |
| **16.2** RoleAssignment partial uniques | **CORRECT.** `WHERE user_id IS NOT NULL` / `WHERE group_id IS NOT NULL` removes the L2 collision; `NULLS NOT DISTINCT` then applies only to `area_id`, which is exactly the Company-scope duplicate case that must be caught. Minimal and right. |
| **16.3** SUBMITTED ⇔ current_submission_id CHECK | **CORRECT.** `(state = 'SUBMITTED') = (current_submission_id IS NOT NULL)` is a valid PostgreSQL boolean CHECK and cannot evaluate to NULL here (`state` is NOT NULL and `IS NOT NULL` never yields NULL). The composite FK proving same-Revision ownership remains separate and still needed. |
| **16.4** Revision/Release coherence proof obligation | **CORRECT.** Three named, falsifiable obligations replace an unnamed residual, with no semantic lifecycle trigger — consistent with D32 and with the enforcement ladder. |
| **16.5** GroupMembership same-Company composite FKs | **CORRECT as structure, over-claimed as proof** — see §14.3 / **L-5**. |
| **16.6** One-open Revision derivation cited | **CORRECT.** Derived from T2 §8 / T6 §5/§40 singular current-open Revision semantics; `CANCELLED` correctly excluded from the predicate. Citation gap closed; no T2 reopen. |
| **16.7** Protected actor narrowed | **NOT CLOSED — MJ-2 below.** |
| **16.8** `audit.events.resource_id` stays NOT NULL | **CORRECT, and the LOW was rightly rejected.** Every *semantic* event in the census has a natural resource; the L8 examples (restore-readiness session invalidation) are operational reconciliation, and T3 §12 already forbids inferring semantic Audit from mechanism activity. Manufacturing a resource UUID to satisfy a column would have been the worse outcome. |
| **16.9** Restore technical-state convergence | **CORRECT.** Each restored mechanism class converges through its own existing rule (River revalidate/resume, GC phase-2 re-proof, claim expiry, idempotency expiry + completion invariant) with no new restore table or recovery journal. Disclosure gap closed at zero structural cost. |
| **16.10** Explicit SQL / no-ORM as reopenable baseline | **CORRECT.** Reframing D38/D39 as a Launch baseline with a named reopen condition removes the T11 over-reach Round 1 flagged, without weakening owner-private query ownership. |

### MJ-2 — the protected-actor narrowing de-freezes a decision the stage exists to freeze

§16.7 deletes D33's blanket rule and replaces it with:

```text
"actor User FOR SHARE is required exactly for the T3/T8-C operation census whose correctness
 requires current enabled-User eligibility to serialize with offboarding/disable."
"The corrected transaction matrix names the requirement per operation."
```

Two problems, both structural:

**1. The artifact it points at does not exist.** No corrected transaction matrix is materialized anywhere in the staging chain. The effective corrected candidate is the original §27, whose rows say `protected actor` on essentially every operation — i.e. they still express the blanket rule the overlay just deleted. §8 has the same dangling reference (*"The corrected transaction matrix must not separately prescribe a contradictory actor-before-target order"*), though there the overlay's own rule resolves the conflict; in §16.7 nothing resolves it.

**2. The census it defers to is explicitly open-ended.** T3 §11 reads:

```text
T3 line 458:      "Governed/security-changing operations whose correctness depends on an enabled
                   User must serialize with offboarding on current User eligibility."
T3 lines 473-479: "This applies AT LEAST to:
                     new Session issuance
                     GroupMembership add
                     new direct User RoleAssignment
                     governance ACCEPT / RETURN_FOR_CHANGES
                     Submission / withdraw / cancel / obsolescence user mutations"
```

*"at least"* is an open lower bound T3 deliberately left for a later stage to close. Operations that plainly meet T3's own predicate but are **absent** from the list include Document create, next Revision, DRAFT PATCH, responsible-owner replacement, UserProfile replacement/erasure, Company replacement and DocumentType configuration replacement. Under D33's blanket rule all of these were decided. After the narrowing, none of them is.

The trade is also worth naming on the merits. `FOR SHARE` is self-compatible, so the blanket rule contends **only** with offboarding — near-zero cost — and it covers, by construction, *every path capable of reaching the protected state*, which is Method §3's enforcement rule stated exactly. The narrowing replaces that with a maintained enumeration: the same hand-synced-enumeration meta-defect class M8 was raised to guard against, now applied to a security serialization property with no parity control.

**Two admissible fixes; the first is smaller and stronger:**

```text
(a) restore D33's blanket rule and record the justification Round-1 L7 asked for —
    "it covers all paths capable of reaching the protected state at self-compatible lock cost"
(b) keep the narrowing and MATERIALIZE the closed per-operation list inside T8-D,
    closing T3 §11's "at least" explicitly
```

Leaving §16.7 as written means a T11 Writer must re-derive a security serialization census from an open-ended upstream list — which the program's own binding law forbids: *"No Writer task may contain a material architecture decision that should have been decided before execution."* → **MJ-2**.

```text
R2-13 VERDICT   9 of 10 accepted LOW corrections VERIFIED CORRECT.
                Findings: MJ-2, L-5.
```

---

## 19. Consolidated finding register

### MAJOR

```text
MJ-1  M2 under-closed: platform.managed_content.sha256 — the referent the CLEAN proof is
      compared against — remains runtime-UPDATE-able, so CLEAN is still forgeable by the
      consuming role. Fix: insert-only READY exact-descriptor facts, same shape as the
      accepted malware relation.                                                      §9.3

MJ-2  §16.7 de-freezes the protected-actor decision: it defers to a "corrected transaction
      matrix" that does not exist and to a T3 §11 census that is explicitly open-ended
      ("applies at least to"). Fix: restore the blanket rule with its justification, or
      materialize the closed per-operation list in T8-D.                              §18

MJ-3  M11's proof role is unsound: metaldocs_ci deliberately retains audit_events DML today,
      and the proof role's serving-equivalence is asserted, never proven. Fix: blocking
      grant-set equality proof between proof role and serving role.                   §17.2

MJ-4  M10's HMAC introduces a key-version dimension into an equality contract with no version
      parameter; after rotation an honest exact retry derives under the wrong version and is
      refused as a semantic conflict. Fix: one active derivation version; rotation drains. §16.5

MJ-5  AdmissionClaim reserve timing is self-contradictory — the new census row places the claim
      at allocation (matching T4-F), unchanged D21/D22 place Reserve after READY, and the
      ledger records neither. Under D21's reading T4-F's GC-liveness protection is not
      realized. Fix: correct D21/D22 and add create-once + lock primitive to the
      upload-complete row.                                                            §12.3

MJ-6  Transaction census still omits PUT /api/v1/areas/{area_id}/lifecycle, whose serialization
      against Document create's "protect/revalidate active Area" is undecided.
      Fix: one census row.                                                            §12.4

MJ-7  GC's semantic-reference proof lock mode is unstated and §18's class order (managed_content
      last) contradicts §20's GC sequence (managed_content first). B2 has just established the
      "prove under FOR SHARE" idiom on exactly these rows, making the cycle-producing instinct
      the natural one. Fix: state that GC's proofs are non-locking reads and that GC acquires
      no lock in classes 1-6.                                                         §11
```

### LOW

```text
L-1  governance_step_candidates materialization (including NAMED_USER) appears in no census row,
     though the M1 FK now makes it mandatory for every decided Step.                  §12.4

L-2  platform.managed_content_backup_pins has no write path in the census; GC proves against it
     and T4-L makes pin correctness a restore-integrity property. Decide, or explicitly defer
     to T8-G.                                                                         §12.4

L-3  River grant class list omits TYPE (river.river_job_state ENUM) and relies on the unstated
     PUBLIC default; and no class is named as the one that may SET ROLE metaldocs_owner.
                                                                                §6.3 / §17.1

L-4  The reindex-OFF reopen trigger names no observable, so it cannot fire. Name the signal
     (river_job index bloat / fetch latency) or route the observability obligation to T8-G. §6.5

L-5  §16.5's same-Company composite FKs cannot fire while CHECK(singleton_key = 1) stands.
     Record them as structural preparation rather than current proofs — and record the singleton
     CHECK as the fail-closed interlock that makes "no isolation substrate" safe, which is the
     answer Round-1's internal-consistency charge never received.              §14.2-14.3

L-6  Routing/precision: the vocabulary-parity control is a falsifiable verification control and
     sits closer to T9's Validation Baseline than to T11; and fingerprint_key_version carries no
     declared type/CHECK where every comparable bounded column does.           §15 / §16.6
```

---

## 20. Global Maximum verdict

```text
GLOBAL MAXIMUM CLASS   CONFIRMED — unchanged
```

Re-tested against the corrected delta rather than by re-litigating Round-1's alternatives.

```text
does any correction move the design toward a different class?   NO
  every correction is subtractive, declarative or privilege-based:
    one insert-only mechanism relation (malware evidence)
    one composite FK (governance decider)
    two partial uniques narrowed to current truth (provider subject)
    two plain uniques added (governance attempt subject)
    one lock-mode statement (attach FOR SHARE)
    one lock-ordering rule (User class)
    one keyed digest (fingerprint)
    one deletion order (retention)
    one privilege/ownership statement (river.*)
    one verification obligation (vocabulary parity)
  no framework, no registry, no generic abstraction, no new authority, no new owner

does any correction create a worse architecture?                NO
  the two places that could have — a second subject-identity table (M3) and universal
  AdmissionClaim consumption (B2) — were both correctly declined in favour of the smaller
  instrument

did a better Global Maximum appear under the corrected delta?   NO
  the strongest candidate tested was "move all mechanism-state mutation behind an insert-only
  event relation" — i.e. generalising the malware fix. Rejected: it converts a bounded
  mechanism table into an event-sourced substrate, which the class explicitly excludes and
  which no current consumer requires. MJ-1's targeted fix achieves the same protection for
  the one class that needs it.
```

---

## 21. Cross-cutting answers

```text
1.  Are BOTH Round-1 blockers actually closed?      YES — BLOCKER-1 and BLOCKER-2 both CLOSED
                                                    (B2's attach census independently verified
                                                    exhaustive against the schema)
2.  Any new BLOCKER?                                NO
3.  Any surviving material contradiction?           YES — 3
                                                      MJ-2  §16.7 vs a non-existent matrix and
                                                            an open-ended T3 census
                                                      MJ-5  new census row vs D21/D22
                                                      MJ-7  §18 class order vs §20 GC sequence
4.  Global Maximum class still CONFIRMED?           YES
5.  Any T8-C reopen?                                NO — MJ-4's fix is a T8-D-local derivation
                                                    law; BeginIn's signature is untouched
6.  Any T8-B reopen?                                NO
7.  Any T1→T7 reopen?                               NO — MJ-5 realizes T4-F rather than reopening
                                                    it; MJ-2 closes a lower bound T3 §11
                                                    explicitly left open
8.  Any T8-E trespass?                              NO — §16.1 defers the exact bound correctly
9.  Any T8-F trespass?                              NO
10. Any T8-G trespass?                              NO — River wiring, secret provisioning and
                                                    deployment topology all correctly deferred
11. Any T10 trespass?                               NO — current-estate River/grant facts are
                                                    cited as evidence only, never as transition
                                                    instructions
12. Any T11 leakage?                                NO material leakage; one LOW (L-6 routing).
                                                    D38/D39 reframed as reopenable removes the
                                                    Round-1 over-reach
13. Is another review round materially required?    NO — all seven MAJORs are bounded, in-class
                                                    and closable without redesign. A third full
                                                    round would be ceremony. Re-review is
                                                    warranted only if the Lead's chosen
                                                    realization of MJ-1 or MJ-2 introduces new
                                                    structure rather than the minimal fix
14. May final Lead adjudication proceed?            YES
```

---

## 22. Reviewer method disclosure

```text
authority reconstructed independently from repository files in the mandated order;
  this prompt, chat context, Lead reasoning and Round-1 conclusions treated as non-authority
remote HEAD and PR head OID revalidated before reading anything; handoff HEAD not trusted
Method mirror verified byte-identical to conexus-methodology/METHOD.md v1.0.0
review executed in a detached worktree at 825fb643; no user working tree was modified
River behaviour verified from the EXACT pinned module source in GOMODCACHE
  (river@v0.37.1, riverdriver/riverdatabasesql@v0.37.1), not from current upstream docs
attach-path census derived from the candidate SCHEMA (managed_content_id columns), not from
  the overlay's prose list
transaction census re-derived from T6's own mutation-route list (T6 lines 1309-1402), not from
  the overlay's additions list
current code, schema and grants read strictly as evidence for feasibility/reuse claims
no database was created, modified, read or probed; no service was started; PostgreSQL version
  behaviour was taken from Round-1's recorded 16.14 evidence and from documented version facts,
  not re-run against any MetalDocs or shared database
bounded scope honoured: the 26 Round-1 decisions accepted unchanged were reopened only where a
  corrected item asserts a closure it does not achieve (MJ-1) or makes a previously inert
  property load-bearing (MJ-7)
```

---

**End of bounded Round-2 delta review. Reviewer evidence only — not T8-D authority.**
