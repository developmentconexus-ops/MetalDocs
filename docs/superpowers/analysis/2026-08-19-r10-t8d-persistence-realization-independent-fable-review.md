# R10-T8D — Persistence Realization — Independent Fable Review

```text
INDEPENDENT REVIEW EVIDENCE
NON-AUTHORITATIVE
NOT TARGET AUTHORITY
NOT OPERATOR RATIFICATION
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Review class:** Standard Fable independent review (`developmentconexus-ops/conexus-methodology/README.md`)
> **Reviewed artifact:** `docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`
> **Stage:** T8-D ACTIVE — independent challenge gate
> **Implementation:** BLOCKED

This artifact is reviewer evidence. Findings are evidence until Lead adjudication and explicit operator ratification. This review creates no authority and changes no upstream decision.

---

## 1. Revalidated reviewed remote HEAD

Revalidated independently at review start; the handoff-supplied HEAD was **not** trusted as input.

```text
git fetch --all --prune
  From https://github.com/developmentconexus-ops/MetalDocs
     02e886cb..a8880d75  docs/a8-authz-approval-redesign-ledger -> origin/...

git ls-remote origin docs/a8-authz-approval-redesign-ledger
  a8880d75cc53f94411a9771ea6004944d7c9760d   refs/heads/docs/a8-authz-approval-redesign-ledger

gh pr view 131 --json headRefOid,state,baseRefName,mergeable
  headRefOid   a8880d75cc53f94411a9771ea6004944d7c9760d
  headRefName  docs/a8-authz-approval-redesign-ledger
  baseRefName  main
  state        OPEN
  mergeable    MERGEABLE
```

```text
REVIEWED REMOTE HEAD = a8880d75cc53f94411a9771ea6004944d7c9760d
```

The independently revalidated HEAD **matches** the handoff claim. The reviewed candidate file at that HEAD declares its own revalidated baseline as `b1d6e36b`; the candidate has not been modified by the four subsequent routing commits.

---

## 2. Authority reconstruction summary

Reconstructed from repository truth in the repository-mandated order, not from this prompt, chat history, the candidate's own summaries or prior Lead reasoning.

```text
AGENTS.md                                                    read — routing/bootstrap only
docs/engineering/standards/root-cause-global-maximum-method.md  read — Method v1.0.0 ACCEPTED
wiki/references/current-agent-handoff.md                      read — T8-D ACTIVE / independent review next
wiki/architecture/r10-technical-architecture.md               read — sole stage/status router
wiki/architecture/launch-v1-product-contract.md               located; consumed via T1→T6 derivations
wiki/architecture/r10-t1-semantic-state-invariants.md         read — owners, lifecycle laws, absent set, seams
wiki/architecture/r10-t2-governance-effectivity-transactions.md read — tx law, OCC, route, Release
wiki/architecture/r10-t3-authorization-audit-enforcement.md   read — roles/scopes, Group deletion, offboarding, Audit
wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md read — malware gate, create-once, restore
wiki/architecture/r10-t5-durable-async-search-external-effects.md read — durable intent, GC, Search OFF
wiki/architecture/r10-t6-canonical-api-frontend-journeys.md   read — §4/§6/§10/§18/§19/§24/§29
wiki/architecture/r10-t7-...                                  consumed: no historical business migration
wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md read — five-proof reuse gate
wiki/architecture/r10-t8b-backend-module-package-topology.md  consumed via router §5
wiki/architecture/r10-t8c-internal-communication-contracts.md read — §3, §8, §9, §10, §18, §20, §21
Decision Registry + amendments through T8-C                   scanned
wiki/architecture/r10-post-t6-implementation-readiness-program.md read — T8-D/T8-E scope split
wiki/architecture/r10-technical-realization-reconciliation-baseline.md read — ownership routing table
```

Current code/schema/migrations were used **as evidence only**, to falsify or confirm T8-A reuse claims:

```text
db/baseline/0001_current_schema.sql
db/grants/0000_identity_roles.sql
internal/platform/idempotency/postgres_store.go
tools/cilint/internal/analyzers/table_ownership.go + table-ownership.json
go.mod / go.sum (River pin)
```

External load-bearing behavior was verified against **primary evidence**, not documentation summaries:

```text
PostgreSQL 16.14 (isolated throwaway container, no project database touched)
River v0.37.1 pinned module source in GOMODCACHE
```

T7's ruling was honoured throughout: no finding in this review derives from compatibility with current DEV/TEST business data.

---

## 3. Primary verdict

```text
APPROVE T8-D GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
```

The selected class is correct and survives adversarial challenge. I could not construct a materially stronger Global Maximum with lower total complexity, and I actively attempted five alternative classes (§6). The candidate nevertheless contains **two blocking realization holes** — one that can lose governed bytes, one that makes two ratified candidate decisions mutually unsatisfiable against the pinned River dependency on the pinned PostgreSQL floor — plus eleven material gaps.

None of the findings require a different class. All are correctable inside T8-D.

---

## 4. Finding counts

```text
BLOCKER   2
MAJOR    11
LOW      10
```

---

## 5. Global Maximum class — confirmed?

```text
CONFIRMED — YES
```

Confirmed after attempting to defeat it. The subtractive pass produced two genuine subtraction candidates (M7 `company_id` propagation, L7 protected-actor widening) and one genuine **addition** that reduces total complexity (M1, a composite FK replacing application-enforced logic). The class itself is not the problem.

---

## 6. Alternative classes tested and defeated

Per Method §3, the class was attacked rather than accepted by default.

### 6.1 Single product schema instead of owner-namespaced schemas — REJECTED

Not on aesthetics. With schema-qualified names, "does owner X emit SQL against owner Y's tables" becomes a **schema-prefix match on the SQL literal inside the owner package**. Under one schema it stays what the repository already runs: a hand-maintained table→owner map (`tools/cilint/internal/analyzers/table-ownership.json`, 56 live base tables, whose own `_doc` records that incompleteness previously produced 19 unguarded tables and invisible foreign writes).

Owner namespacing therefore **removes** an instance of this repository's known hand-synced-enumeration meta-defect. The candidate under-claims this; it is the strongest argument for D02 and should be recorded as such. Migration/grant cost is real but one-time and bounded by D30/D31, which the candidate already pays.

The candidate correctly guards the false implication: *"PostgreSQL schemas are namespaces inside one database, not microservice boundaries."* No schema is speculative — all seven map to a ratified owner or a named current mechanism. `authz.*` holding one table is thin but is a ratified semantic owner with its own grant surface.

### 6.2 Database/service per owner — REJECTED

Contradicts T2 §2 (one local ACID transaction per business transition). Correctly rejected by the candidate.

### 6.3 Release-chain-only effectivity, no `Revision.state` — REJECTED, with proof

This is the required direct challenge (prompt §7). The candidate rejects release-chain-only. I tested whether the rejection is preference or proof.

A release chain *can* carry a structural at-most-one barrier: `UNIQUE(revision_id)` + `UNIQUE(submission_id)` + partial `UNIQUE(document_id) WHERE predecessor_revision_id IS NULL` + `UNIQUE(predecessor_revision_id) WHERE NOT NULL` yields a linear chain with exactly one tip. So the naive objection ("chain-only has no barrier") is false.

It fails for a different, decisive reason: **chain tip ≠ EFFECTIVE**. T2 §10 obsolescence makes the tip non-effective with no successor Release, and T2 §8 cancellation removes candidates without a Release. Under chain-only, "at most one EFFECTIVE Revision per Document" — T1 §2's hardest stated invariant and T2 §9's *"No successful externally observable state may contain two EFFECTIVE Revisions"* — becomes a three-table derivation with **no single declarative barrier**, forcing either SERIALIZABLE, a semantic trigger, or an exclusion constraint over a computed expression. All three contradict the candidate's own ratified posture (READ COMMITTED, D32 zero-trigger baseline).

With stored `Revision.state`, the same invariant is one partial unique index.

```text
release-chain-only = materially WEAKER
candidate rejection = PROVEN, not preference
```

`Revision.state` is technically derivable and therefore carries a coherence obligation (see L4), but it is not duplicate *authority*: Release remains the sole effectivity-establishing fact, and T2 §7–§10 ratify explicit Revision transitions as semantic law. This is exactly T2's lifecycle state plus T2's immutable effectivity evidence, not two authorities for one meaning.

### 6.4 `Document.current_status` / `current_revision_id` pointer — REJECTED

T6 §6 is explicit: *"There is no persisted `Document.current_status` or equivalent second current-state authority."* The candidate's refusal is upstream-mandated, not discretionary.

### 6.5 Single-row "completed-only" idempotency table — REJECTED, empirically

See §15. The alternative requires a commit-time "payload must be present" assertion on one table. Verified on PostgreSQL 16.14:

```text
CREATE TABLE ... CONSTRAINT payload_present CHECK (payload IS NOT NULL) DEFERRABLE INITIALLY DEFERRED;
ERROR:  CHECK constraints cannot be marked DEFERRABLE
```

PostgreSQL accepts `DEFERRABLE` only on UNIQUE, PRIMARY KEY, EXCLUDE and REFERENCES. The single-table variant is therefore expressible only via a constraint trigger — which the candidate explicitly rejected as more hidden — or by committing an `IN_PROGRESS` row, which T8-C §18.3 and T6 §18 forbid. **The two-table paired design is not overengineering; it is the only declarative realization available on the ratified feature floor.**

### 6.6 Generic ORM/repository framework — REJECTED

No named current consumer. Would obscure the per-owner lock/CAS semantics that T8-D exists to freeze.

---

## 7. Disposition — T8D-D01 → D40

```text
D01  one PostgreSQL database                         ACCEPT
D02  owner/mechanism schemas                         ACCEPT WITH CORRECTION   (B1)
D03  fully-qualified first-party SQL                 ACCEPT
D04  PostgreSQL-16 feature floor                     ACCEPT WITH CORRECTION   (B1)
D05  closed bidirectional ownership catalog          ACCEPT WITH CORRECTION   (B1)

D06  persist RoleAssignment only                     ACCEPT WITH CORRECTION   (L2)
D07  no Launch RLS/tenant substrate                  ACCEPT
D08  no permission/status/Search cache               ACCEPT

D09  explicit BIGINT VersionToken                    ACCEPT
D10  WorkingContent generation OCC                   ACCEPT
D11  Revision.state canonical lifecycle              ACCEPT
D12  Release fact + EFFECTIVE barrier                ACCEPT WITH CORRECTION   (L4)
D13  one open Revision                               ACCEPT WITH CORRECTION   (L6)
D14  bounded current_submission_id                   ACCEPT WITH CORRECTION   (L3)

D15  closed relational governance model              ACCEPT
D16  one ACTIVE Step                                 ACCEPT
D17  live GROUP dependency vs candidate snapshot     ACCEPT
D18  no generic workflow persistence                 ACCEPT

D19  semantic exact descriptors                      ACCEPT
D20  ManagedContent mechanism state only             ACCEPT
D21  row-existence AdmissionClaim                    ACCEPT WITH CORRECTION   (B2)
D22  two-phase GC with repeated proof                ACCEPT WITH CORRECTION   (B2)

D23  paired Key + Replay                             ACCEPT
D24  deferred completion FK                          ACCEPT WITH CORRECTION   (M9)
D25  no durable IN_PROGRESS/FAILED                   ACCEPT

D26  River under river.*                             ACCEPT WITH CORRECTION   (B1)
D27  no first-party raw River SQL                    ACCEPT

D28  identity/existence-only cross-owner FK          ACCEPT
D29  no cross-owner semantic cascades                ACCEPT
D30  serving runtime role != DDL owner               ACCEPT WITH CORRECTION   (M11, B1)
D31  grants enforce append-only classes              ACCEPT WITH CORRECTION   (M2, M11)
D32  zero semantic lifecycle triggers baseline       ACCEPT

D33  protected actor FOR SHARE                       ACCEPT WITH CORRECTION   (L7)
D34  User offboarding/eligibility root               ACCEPT WITH CORRECTION   (M4)
D35  Document FOR UPDATE lifecycle root              ACCEPT
D36  DocumentType config serialization               ACCEPT
D37  deterministic lock ordering                     ACCEPT WITH CORRECTION   (M4)

D38  explicit owner-private database/sql SQL         ACCEPT
D39  reject generic ORM/repository                   ACCEPT
D40  normal views only; materialized Search OFF      ACCEPT
```

```text
ACCEPT                    26
ACCEPT WITH CORRECTION    14
REJECT                     0
```

No decision is rejected. Every correction is bounded and lands inside T8-D.

---

## 8. BLOCKER findings

### BLOCKER-1 — `river.*` ownership is unspecified, and D26 + D30 + D04 are mutually unsatisfiable against River v0.37.1's default behavior

**Where:** D02, D04, D05, D26, D30.

**Verified from the exact pinned dependency** (`go.mod`: `github.com/riverqueue/river v0.37.1`, source in `GOMODCACHE`), not from latest-version documentation:

```text
client.go:325-330   Config.Schema EXISTS in v0.37.1
                    "All table references in database queries will use this value as a prefix.
                     Defaults to empty, which causes the client to look for tables using the
                     setting of Postgres search_path."
client.go:518-521   schema name validated (length + [A-Za-z_][A-Za-z0-9_]*)
CHANGELOG.md:292    Config.Schema added precisely so schema selection does not depend on search_path
rivermigrate/river_migrate.go:71-75,167,631-734
                    Migrator Config.Schema EXISTS and is threaded into every migration operation
riverdriver/riverdatabasesql@v0.37.1:48   func New(dbPool *sql.DB) *Driver
riverdriver/riverdatabasesql@v0.37.1:94   func (d *Driver) UnwrapExecutor(tx *sql.Tx) riverdriver.ExecutorTx
client.go:1765      func (c *Client[TTx]) InsertTx(ctx, tx TTx, args, opts)
```

So the candidate's three positive River claims are **CONFIRMED for v0.37.1**: custom schema exists, migrations honour it, and `InsertTx` accepts exactly the concrete `*sql.Tx` that T8-C §3.4 `SQLTx(scope)` yields. Rehoming to `river.*` also *strengthens* D03, because `Config.Schema` is the documented way to stop depending on `search_path`.

**What the candidate missed — three unstated couplings:**

1. **River runs a `REINDEX` maintenance service by default.** `internal/maintenance/reindexer.go:71` — *"Reindexer periodically executes a REINDEX command on the important job indexes"*; `client.go:284-286` — *"ReindexerSchedule ... If nil, the reindexer will run at midnight UTC every day."*

2. **PostgreSQL 16 has no grantable `MAINTAIN` privilege** (added in PostgreSQL 17). Verified on PostgreSQL 16.14:
   ```text
   GRANT MAINTAIN ON public.probe TO PUBLIC;
   ERROR:  unrecognized privilege type "maintain"
   ```

3. **A non-owner role cannot REINDEX.** Verified on PostgreSQL 16.14 with the candidate's own role split (`ddl_owner` owns `river.*`, `serving_runtime` has USAGE + full DML):
   ```text
   SET ROLE serving_runtime;
   REINDEX INDEX CONCURRENTLY river.river_job_state_idx;
   ERROR:  must be owner of index river_job_state_idx
   REINDEX INDEX river.river_job_state_idx;
   ERROR:  must be owner of index river_job_state_idx

   GRANT ddl_owner TO serving_runtime;  SET ROLE serving_runtime;
   REINDEX INDEX river.river_job_state_idx;   -- REINDEX (succeeds)
   ```
   The membership workaround defeats D30 entirely: `SET ROLE` then grants the serving process the DDL owner's full powers.

**Additionally:** River's migrations contain no `CREATE SCHEMA` (only its test fixtures do). The `river` schema must be provisioned by MetalDocs DDL. The candidate never states who creates `river`, who owns it, or what the serving role is granted on it — while D05 makes *"live target object absent from catalog = FAIL"* a binding verification law and D27 forbids first-party SQL there. As written, D05 cannot be satisfied for `river.*`.

**Why BLOCKER rather than MAJOR:** this is not an omission of detail. Two ratified candidate decisions (D26, D30) plus the ratified feature floor (D04) plus the pinned dependency's default behavior cannot all hold simultaneously. T8-D is the stage that freezes the privilege model and the ownership catalog; shipping the contradiction forward means either a permanently failing nightly maintenance job or a silent D30 breach discovered during implementation.

**Smallest sufficient corrections (Lead to choose one):**

```text
(a) serving runtime role OWNS river.*    — river.* is third-party mechanism state, carries no
                                           first-party immutable history, and D31's protections
                                           do not apply to it; preserves D30's actual intent
                                           (protect first-party semantic/immutable objects)
(b) River Reindexer disabled             — ReindexerIndexNames: []string{}
                                           (reindexer.go:40-42: "An empty slice disables reindex work")
```

Either way T8-D must additionally state: who creates the `river` schema, who owns it, what the serving role is granted on it, how it enters the D05 catalog as a third-party class, and that D04's floor is stated with its one material consequence (no `MAINTAIN` before PostgreSQL 17).

---

### BLOCKER-2 — attach-side ManagedContent lock mode is unspecified, admitting a constructible lost-governed-bytes race

**Where:** D21, D22, D37 step 7.

D22 correctly serializes both GC phases on `platform.managed_content FOR UPDATE`. The semantic attach paths, however, are specified only as *"prove exact handle READY"* (§27 SUBMIT / Document create / Next Revision / DRAFT PATCH / OfficialRendition finalization) and D37 step 7 names *"ManagedContent technical row/claim where participating"* **without a lock mode**.

**Constructible sequence** (READ COMMITTED, per-command snapshots, no attach-side lock):

```text
t0  attach tx BEGIN; SELECT state FROM platform.managed_content WHERE id=H  -> READY
t1  GC phase 1 commits: H -> GC_PENDING
t2  GC phase 2: H FOR UPDATE (uncontended — attach tx holds no lock on H)
      still GC_PENDING                                     ok
      full ControlledDocs semantic-reference proof          none (attach not yet inserted)
      live-claim proof                                      none (this path took no claim)
      backup-pin proof                                      none
    phase 2 COMMITS
t3  attach tx INSERTs the semantic row referencing H; FK satisfied (row still exists); COMMIT
t4  provider DeleteReclaimable(H)

result: committed governed reference to deleted bytes
```

The AdmissionClaim closes this **only if claim consumption is universal and in-transaction** — because a claim is either still visible to phase 2, or its `DELETE` is bundled into the attach commit alongside the semantic row, so phase 2 sees one or the other. The candidate does not make it universal: Document create says *"consume AdmissionClaim **if applicable**"*, and D21 describes the claim lifecycle without a rule binding every `managed_content_id` write to a claim consumption.

This defeats T4's strongest promise (*"Safe failure remains leaked storage, never deleted governed truth"*) and T4-M restore readiness, which will then fail closed on a reference whose bytes are gone.

Note the FK gives no protection: the attach's FK-induced `FOR KEY SHARE` is acquired at `t3`, after phase 2 has already committed.

**Verified fix on PostgreSQL 16.14** — making the attach-side READY proof take a row-level share lock closes the window structurally:

```text
session A (attach):  BEGIN; SELECT state FROM mc WHERE id=1 FOR SHARE;  -- holds 3s
session B (GC ph.2): SELECT state FROM mc WHERE id=1 FOR UPDATE;
                     Time: 2081.472 ms      <- blocked until attach committed
```

Phase 2 then either waits and re-proves (seeing the committed semantic row → abort) or wins first, in which case the attach's own re-read under the lock observes `GC_PENDING` and fails closed.

**Correction:** state the attach-side lock mode explicitly — every semantic transaction that writes a `managed_content_id` takes `SELECT ... FROM platform.managed_content WHERE id = $1 FOR SHARE` as part of its READY proof — **and/or** make in-transaction `AdmissionClaim` consumption structurally universal for every `managed_content_id` write. D37 step 7 must carry the mode, not just the class.

---

## 9. MAJOR findings

### M1 — missing structural barrier: the decider is not constrained to the frozen candidate snapshot

T2 §6 is explicit: a Group Step is ANY-one, *"one currently authorized User **from the activation snapshot** may make the Step decision"*, and *"later Group membership drift does not rewrite the active Step candidate set"*. The empty-snapshot case must remain undecidable — the bounded escape is withdraw → fix route → resubmit.

The candidate leaves this entirely to owner SQL. `controlled_docs.governance_decisions` has `step_id` and `actor_user_id` with **no relationship to `governance_step_candidates`**.

A constraint is already available at zero structural cost. `governance_step_candidates` has `PRIMARY KEY(step_id, user_id)`, and the candidate already states the table holds the resolved NAMED_USER candidate too (*"giving one exact active-candidate authority"*). Therefore:

```text
governance_decisions (step_id, actor_user_id)
  FOREIGN KEY -> governance_step_candidates (step_id, user_id)
```

makes "the decider was a frozen candidate of exactly this Step" structural, and makes "empty snapshot ⇒ no decision is committable" structural as a consequence rather than as a hoped-for application check. This is a **net reduction** in total complexity: one FK replaces an application invariant on the highest-value governance path.

Per D32's own enforcement ladder (constraints → grants → SQL), a constraint that is available and sufficient must be preferred. This is the single strongest improvement I found.

### M2 — malware verdict is the one security-critical proof left mutable, and mutable by the path that consumes it

T4-G: *"Untrusted external bytes cannot become immutable governed MetalDocs content without successful malware inspection of **those exact immutable bytes**."*

The candidate stores `malware_verdict`, `malware_digest`, `malware_checked_at` as **nullable, mutable columns** on `platform.managed_content` — a table the serving runtime must hold `UPDATE` on for `OPEN→READY→GC_PENDING`, `ready_at`, `gc_pending_at` and provider-locator migration. Column-level grants cannot separate the writer from the reader here, because the malware inspection path and the SUBMIT gate run under the same serving role.

Consequences:

```text
the gate reads   malware_verdict='CLEAN' AND malware_digest = <semantic sha256>
the same role can write both columns
=> a defect or compromise on any managed_content UPDATE path can flip MALICIOUS -> CLEAN
=> the admission-time proof is destroyed by any later overwrite; no history survives
```

Every other integrity-bearing class in this candidate is protected by the design's own `PRIVILEGE-ENFORCED IMMUTABLE HISTORY` element. This one is not, and it is the class where fail-open is worst.

**Correction:** an insert-only proof relation, e.g. `platform.content_malware_verdicts(managed_content_id, digest, verdict, checked_at)` with `SELECT, INSERT` only. Cost: one small mechanism table, no new authority, no new owner, no lifecycle. Benefit: a `MALICIOUS` verdict becomes permanent; a contradictory later row is visible rather than silent; rescan or provider migration cannot invalidate the admission-time proof. This is strictly stronger at essentially equal complexity and is required to make D31 coherent.

### M3 — global `UNIQUE(issuer, subject)` makes an ordinary admin mis-binding permanently unrecoverable

`authn.provider_subject_bindings` declares `UNIQUE(issuer, subject)` over **all** rows including replaced history, plus a partial unique current binding per User.

Failure sequence from a plausible operator error:

```text
admin mis-binds User A to subject S_B (which belongs to person B)
correction: A's binding replaced -> A's row for S_B remains with replaced_at set
now bind User B to S_B  -> BLOCKED by global UNIQUE(issuer,subject), permanently
```

There is no binding-delete API in T6 §29 and history rows are immutable apart from `replaced_at`. Person B can never authenticate. The same rigidity blocks any legitimate revert to a previously used subject (IdP migration rolled back).

The property actually being protected — a subject never silently moves to a different User — does not require uniqueness over history. Audit correlates by `actor_user_id`, never by subject, so no historical correlation is lost by scoping the current-holder rule to current rows.

**Correction (either is sufficient):**

```text
(a) partial UNIQUE(issuer,subject) WHERE replaced_at IS NULL
    + insert-only authn.provider_subject_identities(issuer, subject, user_id) PRIMARY KEY(issuer,subject)
    bindings FK to it  -> same-User re-binding allowed, cross-User reassignment structurally impossible
(b) keep global uniqueness and explicitly ratify that historical subject reuse is forbidden,
    naming the non-serving operations repair path for the mis-binding case
```

Option (a) is strictly more expressive at one small table; option (b) is acceptable only if the operator accepts the unrecoverable state as a deliberate ruling.

### M4 — §27 offboarding lock order contradicts D37 and yields a constructible deadlock cycle

D37 requires ascending stable-UUID ordering within a lock class, and places actor and target User locks together in class 2. The §27 matrix instead prescribes a fixed order:

```text
Offboarding:              protected current admin actor -> target User FOR UPDATE
User eligibility replace: protected actor -> target User update lock
```

Two admins offboarding or disabling each other concurrently:

```text
tx1: FOR SHARE(A)  then wants FOR UPDATE(B)
tx2: FOR SHARE(B)  then wants FOR UPDATE(A)
FOR UPDATE conflicts with FOR SHARE  ->  40P01 cycle
```

This is **architecture-induced and avoidable**, not an acceptable retryable 40P01: D37's own rule eliminates it if actor and target User rows are treated as one ordered set acquired in UUID order at their required strengths. The matrix text as written instructs the opposite.

**Correction:** state in D37 and in §27 that when a transaction needs both the protected actor row and a target User row, both are acquired in ascending UUID order, each at its required mode.

### M5 — transaction/lock matrix is incomplete against T6 §29 and T6 §18

T8-D freezes the transaction/serialization/lock mapping, so an operation absent from §27 is an undecided decision, not an implementation detail. Missing:

```text
POST /api/v1/areas                          durable Idempotency-Key per T6 §18 — absent
POST /api/v1/groups                         durable Idempotency-Key per T6 §18 — absent
DELETE /api/v1/role-assignments/{id}        access revocation — absent entirely
PUT /api/v1/company                         Company VersionToken CAS (T6 §24) — absent
POST /api/v1/revisions/{id}/draft/uploads
POST .../uploads/{upload_id}/complete       the OPEN->READY create-once/malware admission path — absent
DELETE /api/v1/session                      logout — absent (trivial, listed for completeness)
```

The upload/complete omission is the material one: it is the transaction that establishes `READY`, the create-once property and the malware verdict, and it is where BLOCKER-2 and M2 both land.

### M6 — `governance_attempts` lacks per-subject uniqueness

Nothing structurally prevents two attempts for one governed subject. Per T2 §7–§8 an attempt is terminal on RETURN/WITHDRAW/CANCEL and a retry requires a **new** Submission, so exactly one attempt per subject exists for all time:

```text
UNIQUE(submission_id)              -- missing
UNIQUE(obsolescence_request_id)    -- missing
```

Both are partial-free plain uniques on nullable columns (NULLs distinct by default), so they cost nothing and compose with the existing XOR check. The candidate structurally enforces one ACTIVE Step (D16) but leaves the enclosing attempt unbounded.

### M7 — hard Company singleton and full `company_id` propagation cannot both be the smallest sustainable solution

`org.companies.singleton_key SMALLINT NOT NULL UNIQUE CHECK(singleton_key = 1)` structurally forbids a second Company. Yet a `NOT NULL company_id` FK is carried on `org.users`, `org.areas`, `org.groups`, `org.group_memberships`, `authz.role_assignments`, `controlled_docs.document_types`, `controlled_docs.documents` and `audit.events` — nine columns whose value is provably constant while the CHECK stands.

Under the singleton, `UNIQUE(company_id, code)` is exactly `UNIQUE(code)`, `role_assignments.company_id` is redundant with `area_id NULL = Company scope`, and `audit.events.company_id` is redundant with the same `area_id NULL` convention the candidate itself defines for visibility.

This collides with two binding laws the candidate accepts:

```text
candidate class:  - DUPLICATE CURRENT TRUTH
T1 §6:            "Defer the capability; preserve the evolution seam.
                   Prepare the seam, not the dormant implementation."
                   pooled tenancy -> stable Company identity + reopenable substrate
```

T1's seam is *stable Company identity*. Replicating the tenancy column skeleton across nine tables under a DDL constraint that forbids the second tenant is building the substrate while forbidding its use.

**Correction — pick one and state it:**

```text
(a) Company is genuinely singleton at Launch -> drop the redundant company_id columns;
    uniqueness scopes collapse to their natural keys
(b) company_id is the ratified pooled-tenancy seam -> drop CHECK(singleton_key = 1) and enforce
    the Launch single-Company fact through bootstrap/configuration, not a DDL constraint that
    a later reopen must delete
```

This is not fatal to the class; it is an internal-consistency ruling T8-D owes.

### M8 — closed-vocabulary `TEXT + CHECK` enumerations reintroduce the hand-synced-enumeration defect class with no parity guard

Roughly twelve closed vocabularies are realized as `TEXT + CHECK (... IN (...))`: `revisions.state`, `areas.state`, `governance_attempts.state`, `governance_attempt_steps.state`, `governance_decisions.outcome`, `obsolescence_requests.state`, `managed_content.state`, `managed_content.trust_class`, `malware_verdict`, `numbering_scope`, `governance_mode`, `representation_mode`, `selector_kind`, `actor_kind`, plus `role_assignments.role_code` *"in static Launch Role vocabulary"*.

The candidate rejects PostgreSQL enum types *"because they add DDL coupling without improving current semantic authority"*. But `TEXT + CHECK` has **the same** DDL coupling to the Go authority, minus the type. Either way the Go vocabulary and the DDL predicate are two copies of one truth, hand-synced.

This is a named MetalDocs meta-defect class (hand-synced enumerations), and D06 makes it load-bearing: `role_code` correctness is an authorization property. §33's proof obligations contain **no** Go↔DDL vocabulary parity control.

**Correction:** add a named parity/generation control to §33 — either the CHECK predicates are generated from the Go vocabulary authority, or a blocking parity check compares them — so the control can be shown to fire (Method §3: *"A control that cannot be shown to fire is not proven"*).

### M9 — idempotency retention has a mandatory delete order the candidate does not state

Verified on PostgreSQL 16.14 against the candidate's exact constraint pair:

```text
-- replay first, then key: OK
BEGIN; DELETE FROM idempotency_replays ...; DELETE FROM idempotency_keys ...; COMMIT;   -> COMMIT

-- key first, then replay: FAILS
BEGIN; DELETE FROM idempotency_keys ...;
ERROR:  update or delete on table "idempotency_keys" violates foreign key constraint
        "replay_key_fk" on table "idempotency_replays"

-- replay alone (orphaning a key): correctly FAILS at COMMIT via the deferred constraint
ERROR:  update or delete on table "idempotency_replays" violates foreign key constraint
        "key_completion_fk" on table "idempotency_keys"
```

The forward FK `replays.key_id → keys.id` is **NOT DEFERRABLE**, so it fires at end-of-statement. D24's *"Retention cleanup removes key+replay together inside one technical transaction"* is therefore insufficient: same-transaction is necessary but not sufficient, the order is mandatory.

**Correction:** state the mandatory order (replay before key) **or** declare the forward FK `DEFERRABLE INITIALLY DEFERRED` as well, making order irrelevant. The third result above is a genuine strength worth recording: a key cannot be orphaned by deleting its replay.

### M10 — the 32-byte semantic fingerprint can outlive a lawful UserProfile erasure

`platform.idempotency_keys.semantic_fingerprint` is a plain 32-byte digest derived from *"canonical validated semantic command fields"*. For `POST /api/v1/users` those fields include the erasable UserProfile data (`display_name`, `email`) that T6 §4 places in the create command.

A raw SHA-256 over a low-entropy value such as an email address is trivially brute-forceable, so the stored digest remains reasonably-linkable derived personal data. The candidate's defense — *"It is not a copy of raw HTTP bytes or erasable profile data"* — is true of the representation and false of the information content.

T8-C §18.4 binds the **ReplaySnapshot** to PII-free-by-construction; it says nothing about the fingerprint. T6 §19 does state *"Replay storage must not become an unintended retention root for erasable UserProfile PII"*, which the fingerprint arguably violates.

Mitigating facts: `expires_at` bounds the window and retention removes the pair. So this is bounded exposure, not indefinite.

**Correction (either):**

```text
(a) keyed digest — HMAC-SHA-256 under a server-held key, making offline brute force infeasible
(b) exclude erasable UserProfile fields from the fingerprint input, fingerprinting only
    non-erasable semantic command identity
```

Collision risk is negligible and creates no authority problem: 32 bytes of SHA-256 output, and the fingerprint only discriminates within an already client-chosen `(actor, operation, key)` scope.

### M11 — D30 selects two DB trust classes where current evidence proves more, and D31's property is false-green unless the proof names its connecting role

`db/grants/0000_identity_roles.sql` shows the estate already operates four identity classes:

```text
bootstrap superuser (provisioning: CREATE ROLE, ownership transfer)
metaldocs_owner     NOSUPERUSER NOBYPASSRLS NOLOGIN   — owns schema, runs DDL
metaldocs_runtime   NOSUPERUSER NOBYPASSRLS LOGIN     — serving
metaldocs_ci        NOSUPERUSER NOBYPASSRLS LOGIN     — non-owner DML test role
```

D30 names two. The candidate's T8-A row claims `runtime DB identity != DDL owner → PRESERVE`, preserving one half of a lesson the repository paid for in full. Two consequences:

1. **Provisioning identity is unclassified.** Someone must `CREATE SCHEMA authn/org/authz/controlled_docs/audit/platform/river` and assign ownership before any migration runs. D05's bidirectional catalog has no class for it, and BLOCKER-1 lands exactly here.

2. **D31 is unprovable from the wrong connection.** A PostgreSQL table owner holds implicit full privileges; `REVOKE UPDATE` against the owner does not restrict the owner. So *"Audit rows cannot be UPDATE/DELETE/TRUNCATE by runtime"* (§33) is **false-green if the proof suite connects as the owner or a superuser** — precisely the failure mode `metaldocs_ci` was created to eliminate.

**Correction:** classify the provisioning identity explicitly, and require every §33 grant-enforcement proof obligation to name the connecting role as a non-owner serving-equivalent role. Do not add further role proliferation beyond these named risk reductions.

---

## 10. LOW findings

```text
L1  ReplaySnapshot 64 KiB cap is undisclosed legacy carryover.
    db/baseline/0001_current_schema.sql:1351 caps idempotency_keys.response_body at 65536 —
    the raw-HTTP-body artifact D23/D25 explicitly reject. The audit 64 KiB bound IS disclosed
    as PRESERVE PROPERTY; this one is not, and no T8-A proof is applied to the constant.
    Answers prompt H: accidental carryover, harmless (fail-closed bound), but it should be
    either declared as a reused property or re-derived from the ReplaySnapshot content census.

L2  authz.role_assignments duplicate-prevention indexes must be PARTIAL. As written,
    UNIQUE NULLS NOT DISTINCT(user_id, role_code, area_id) would treat every Group-subject row
    as user_id NULL and collide them. Requires WHERE user_id IS NOT NULL / WHERE group_id IS NOT NULL.

L3  "SUBMITTED iff current_submission_id is present" is stated as law with no named constraint.
    Expressible: CHECK ((state = 'SUBMITTED') = (current_submission_id IS NOT NULL)).

L4  Revision.state is stored and also derivable, so state/Release coherence is a real residual.
    Not declaratively closable without a trigger (correctly avoided per D32), so §33 must carry
    the proof obligation: no Revision reaches EFFECTIVE/SUPERSEDED/OBSOLETE without a Release row.

L5  org.group_memberships same-Company enforcement is left as "through target composite constraints
    where practical" — an undecided constraint in the stage that freezes constraints. Composite FKs
    against unique (company_id, id) keys are available; decide or state why not.

L6  D13's one-open-Revision partial uniqueness is upstream-IMPLIED, not explicitly ratified.
    T2 §8 "the current open Revision", T6 §5/§40 "exact current open Revision DRAFT" are singular,
    and T6 §10 defers to "no T2 conflict". The constraint is correct — CANCELLED is excluded from
    the predicate, and a second open cycle is clearly unintended — but cite the derivation rather
    than assert it. This is a citation gap, NOT a T2 reopen.

L7  D33 widens T3 §11's protected-actor census from its named list to every authenticated semantic
    mutation. T3 permits it ("applies at least to"), and FOR SHARE is self-compatible so contention
    is only against offboarding. But no named consumer requires the widening; justify or narrow.

L8  audit.events.resource_id UUID NOT NULL forces a resource UUID on every event. Deployment-scope
    events (e.g. restore-readiness session invalidation, T4-M) have no natural resource identity.

L9  Restore readiness enumerates only ApplicationSessions. River pending jobs, GC_PENDING handles and
    AdmissionClaims are also restored mechanism state. Each self-heals structurally — renditions by
    UNIQUE(submission_id, required_format), GC by phase-2 re-proof, claims by expires_at — so this is
    a disclosure gap, not a defect.

L10 D38/D39 (explicit SQL, no ORM) edges toward T11 implementation choice. Acceptable as framed
    ("not required by T8-D", reopen on named benefit); mark as baseline-with-reopen rather than frozen.
```

---

## 11. Persistent-state census completeness verdict

```text
COMPLETE, with one classification correction (M2) and one subtraction question (M7)
```

Attacked every PERSIST / STATIC / DERIVED / DEFER classification against T1 §1/§5/§6, T3, T4, T5 and T6 §29.

**Confirmed correct:**

```text
Role / Permission / Role→Permission = static code authority
  T1 §1: "Role/Permission semantics are product-owned, not customer-defined platform data."
  T3 §3-§5 fixes the vocabulary and bundles; T6 §29 exposes GET /roles as read-only.
  No customization surface exists -> no persistent catalog is required anywhere. CONFIRMED.

RoleAssignment = persisted current grant truth              CONFIRMED (T1 §1, T3 §6)
no RLS / tenant substrate                                    CONFIRMED (T1 §6 reopenable substrate)
no permission cache / materialized ACL                       CONFIRMED (T3 evaluates per request)
no persisted Document.current_status                         CONFIRMED — T6 §6 forbids it explicitly
no materialized Search                                       CONFIRMED (T5 §7-§10, T6 §6 all OFF)
no generic Workflow                                          CONFIRMED (T1 §1 bounded attempt, T2 §6)
no generic outbox                                            CONFIRMED (T5 §5 named intents only)
no notifications                                             CONFIRMED (T5 §12 no Launch consumer)
no Artifact / TemplateVersion semantic family                CONFIRMED (T1 §5 absent set)
```

**No upstream fact was found unpersisted.** I checked every T1 §1 family and every T6 §29 operation against the table set; each persistent fact has exactly one home.

**One persisted fact whose protection class is wrong:** the malware verdict (M2) — persisted correctly as mechanism state, but classified as mutable when its consumer is a security gate.

**One persisted fact with no current consumer:** `company_id` propagation under the singleton CHECK (M7).

Notable positive: the census correctly resists the two easiest mistakes — no `Document.current_status` (T6-forbidden) and no persistent Role catalog (T1/T3-unnecessary).

---

## 12. Schema / table / constraint completeness verdict

```text
SUBSTANTIALLY COMPLETE — three missing structural constraints, one under-specified index family
```

```text
missing  governance_decisions (step_id, actor_user_id) -> governance_step_candidates   M1
missing  governance_attempts UNIQUE(submission_id) / UNIQUE(obsolescence_request_id)   M6
missing  CHECK ((state='SUBMITTED') = (current_submission_id IS NOT NULL))             L3
partial  authz.role_assignments duplicate indexes need WHERE predicates                L2
open     org.group_memberships same-Company composite FKs "where practical"            L5
```

**Verified sound** (traced constraint-by-constraint):

```text
revisions UNIQUE(document_id, ordinal) + no serving DELETE  => ordinal non-reuse (T2 §1)   OK
documents UNIQUE(company_id, code)     + no serving DELETE  => code never rebinds (T2 §4)  OK
revisions (id, current_submission_id) -> submissions(revision_id, id)
  MATCH SIMPLE default: NULL current_submission_id satisfies the FK trivially — correct     OK
submissions UNIQUE(revision_id, id) exists to support that composite FK                     OK
releases UNIQUE(revision_id) + UNIQUE(submission_id)
  + partial UNIQUE(document_id) WHERE predecessor IS NULL
  + UNIQUE(predecessor_revision_id)                          => linear release chain        OK
document_number_counters UNIQUE NULLS NOT DISTINCT(document_type_id, area_id)
  NULLS NOT DISTINCT is PostgreSQL 15+; correct on the PG16 floor                           OK
revisions/submissions circular FK is order-satisfiable in the SUBMIT sequence (insert
  submission, then update revision) — no deferral required                                  OK
```

`platform.managed_content` referenced by `working_contents`, `submissions`, `official_renditions` and `document_origins` gives an **unclaimed structural backstop**: no managed_content row can be dropped while any semantic row references it. The candidate under-claims this; it should be recorded, because it is a second line behind the GC proofs (though it does **not** close BLOCKER-2, which is about provider bytes, not the row).

---

## 13. Transaction / lock matrix completeness verdict

```text
INCOMPLETE — six operations absent (M5), one internal contradiction (M4), one unspecified mode (BLOCKER-2)
```

Traced all 26 §27 entries plus the T6 §29 census. Deadlock analysis:

```text
create User / offboarding / eligibility        M4 cycle — real, avoidable via D37
GroupMembership add                            actor FOR SHARE + target FOR SHARE — compatible, no cycle
RoleAssignment create                          same — no cycle
Group delete                                   FK RESTRICT backstops both interleavings — fail-closed
Document create / create from Template         one Document row locked (target not yet inserted) — no cycle
next Revision / DRAFT update / SUBMIT          single Document root — no cycle
ACCEPT / RETURN / withdraw / cancel            single Document root — no cycle
owner replacement / Template role change       actor+target FOR SHARE then Document FOR UPDATE — no cycle
governance-config / eligible-template replace  DocumentType class 4 only, no Document lock — no cycle
obsolescence initiate / withdraw / complete    single Document root — no cycle
OfficialRendition completion                   single Document root — no cycle
GC phase 1 / phase 2                           managed_content root; see BLOCKER-2 for the attach side
idempotency contention                         verified non-deadlocking (§15); claim is lock class 1,
                                               acquired before any Document lock — correct ordering
```

Only M4 produces an architecture-induced cycle, and it is eliminated by the candidate's own D37 rule.

### `ProtectedSecuritySubjectIn → User FOR SHARE` — challenged and CONFIRMED

The prompt asks whether `FOR SHARE` is the correct minimum. Verified empirically on PostgreSQL 16.14:

```text
holder: SELECT ... FROM users_probe WHERE id=1 FOR SHARE          (held 3s)
writer: UPDATE users_probe SET enabled=false, eligibility_version=2 WHERE id=1
        Time: 2210.067 ms        <- BLOCKED for the holder's lifetime

holder: SELECT ... FOR KEY SHARE                                  (held 3s)
writer: UPDATE users_probe SET enabled=false WHERE id=1
        Time: 0.596 ms           <- NOT blocked
```

Offboarding is a non-key `UPDATE`, which takes `FOR NO KEY UPDATE`. That conflicts with `FOR SHARE` but **not** with `FOR KEY SHARE`.

```text
FOR SHARE          correct minimum — blocks offboarding, self-compatible so independent
                   protected actions do not serialize against each other
FOR KEY SHARE      INSUFFICIENT — would not serialize eligibility at all (0.596 ms, proven)
FOR NO KEY UPDATE  unnecessarily exclusive — would serialize concurrent protected actions
                   against each other for no invariant gain
FOR UPDATE         strictly worse for the same reason
```

The second measurement also proves that FK-induced key-share locks give **zero** eligibility serialization, which is why the candidate is right to say FK internal locking is not cross-owner semantic communication.

```text
D33 lock-mode selection = CONFIRMED CORRECT with primary evidence
```

---

## 14. Cross-owner FK verdict

```text
CORRECT — the identity-only rule is stable and Writer-usable
```

Traced every proposed cross-owner FK against T8-B/T8-C ownership.

```text
no proposed FK moves semantic authority into the database
  every one references a stable identity/existence row, never a state column;
  no FK reads enabled, authorized, EFFECTIVE, eligible or deletable
owner mutations satisfy every FK without foreign SQL
  each referencing owner writes only its own tables; the referenced side is read via
  T8-C application-routed facts, not via a join
D29 no cross-owner cascades — correct; RESTRICT/NO ACTION keeps deletion decisions in owners
```

**Group deletion satisfies T3 §8 exactly.** T3 §8 names four live dependencies; the candidate produces exactly four structural blockers and no fifth:

```text
current GroupMembership                  org.group_memberships.group_id -> org.groups(id)
current Group RoleAssignment             authz.role_assignments.group_id -> org.groups(id)
current GovernanceRoute GROUP selector   document_type_governance_steps.group_id -> org.groups(id)
unactivated live GROUP Step              governance_group_dependencies.group_id -> org.groups(id)
```

And T3 §8's release condition — *"Completed historical attempts therefore do not keep Group alive solely because the Group once participated"* — is satisfied by `group_id_snapshot` carrying **no** FK. Exact match, no over- and no under-constraint.

**UserProfile erasure and User stability remain correct.** Every historical actor reference (`submissions.submitter_user_id`, `governance_decisions.actor_user_id`, `audit.events.actor_user_id`, `governance_step_candidates.user_id`) targets `org.users`, which is non-deletable; erasure drops only the `org.user_profiles` row. Display names are composed outside owner SQL, so erasure degrades presentation without breaking history.

**"Identity-only FK" is a stable rule for Writers,** because the candidate supplies the falsifiable negative form — *"Database FK never decides: User enabled? actor authorized? Document current EFFECTIVE? Template currently eligible? Group deletion semantically allowed?"* That is a checkable test, not a slogan.

**One consistency observation (not a defect):** Audit carries FKs to `org.users`/`org.companies`/`org.areas`, so audit history pins those identities — the opposite of the deliberate no-FK treatment of `group_id_snapshot`. This is coherent only because Users are non-deletable and Areas retire rather than delete (`state IN ('ACTIVE','RETIRED')`), while Groups genuinely delete. Worth stating explicitly so a Writer does not read the two patterns as contradictory. Note `audit.events.resource_id` correctly carries **no** FK, which is what keeps Group deletion possible after Group-related audit events exist.

---

## 15. Idempotency Key↔Replay verdict — with PostgreSQL evidence

```text
DESIGN CONFIRMED VALID AND OPERABLE ON POSTGRESQL 16 — one stated correction (M9), one privacy correction (M10)
```

Executed against **PostgreSQL 16.14** in an isolated throwaway container, using the candidate's exact declared shapes. No project database was touched. Round-1 T8-C savepoint/retry assumptions were not carried in; every claim below rests on primary evidence.

**A. Can the cyclic Key↔Replay FK schema be created cleanly? — YES**

```text
CREATE TABLE platform.idempotency_keys (... UNIQUE(actor_user_id, operation_id, key) ...);   CREATE TABLE
CREATE TABLE platform.idempotency_replays (key_id UUID PRIMARY KEY, ...);                    CREATE TABLE
ALTER TABLE idempotency_replays ADD CONSTRAINT replay_key_fk FK(key_id) -> keys(id);         ALTER TABLE
ALTER TABLE idempotency_keys    ADD CONSTRAINT key_completion_fk FK(id) -> replays(key_id)
                                    DEFERRABLE INITIALLY DEFERRED;                           ALTER TABLE
```

The cycle is creatable via the standard create-then-ALTER sequence. Both referenced columns are primary keys, so both FKs are well-formed.

**B. Does the winner path satisfy the deferred FK? — YES**

```text
BEGIN;
INSERT INTO idempotency_keys ... ON CONFLICT (actor_user_id, operation_id, key) DO NOTHING;   INSERT 0 1
  <semantic work>
INSERT INTO idempotency_replays VALUES (<key id>, 1, '\xCAFE', now());                        INSERT 0 1
COMMIT;                                                                                       COMMIT
```

**C. Does forgetting `CompleteIn` fail COMMIT and roll back the business transaction? — YES**

```text
BEGIN;
INSERT INTO idempotency_keys VALUES ('2222...', ...);        INSERT 0 1
COMMIT;
ERROR:  insert or update on table "idempotency_keys" violates foreign key constraint "key_completion_fk"
DETAIL: Key (id)=(2222...) is not present in table "idempotency_replays".

post-check: only the test-B key remains  ->  the whole transaction rolled back
```

D24's central claim — *"successful idempotent semantic fact commit ⇔ completed ReplaySnapshot commit"* — is **structurally enforced**, and the enforcement was observed firing.

**D. Can retention cleanup remove the pair cleanly? — YES, in one order only**

See M9. Replay-then-key commits; key-then-replay fails on the non-deferrable forward FK; deleting the replay alone correctly fails at COMMIT, so a key cannot be orphaned.

**E. Does the immediate scoped unique remain usable by `ON CONFLICT DO NOTHING` while the reverse FK is deferred? — YES**

```text
BEGIN;
INSERT ... ON CONFLICT (actor_user_id, operation_id, key) DO NOTHING;   INSERT 0 0
SELECT 'E_TX_ALIVE';                                                     E_TX_ALIVE
COMMIT;                                                                  COMMIT
```

The arbiter index is the immediate, non-deferrable scoped unique, exactly as the candidate specifies. The deferred FK does not interfere with speculative insertion.

**F. Loser behavior under ratified READ COMMITTED — SATISFIES T8-C §18.3**

Winner commits (two concurrent sessions, real contention):

```text
W: BEGIN; INSERT key K9 ON CONFLICT DO NOTHING  -> INSERT 0 1; sleep 3; INSERT replay; COMMIT
L: BEGIN; INSERT key K9 ON CONFLICT DO NOTHING  -> blocked ~2s, then INSERT 0 0
   -- SAME transaction, next command:
   SELECT k.id, encode(r.payload,'hex'), fingerprint_match
     FROM idempotency_keys k JOIN idempotency_replays r ON r.key_id = k.id ...
   -> 9999...9991 | cafe | t
   COMMIT   ->  "L: COMMITTED WITHOUT POISON"
```

The loser blocked (serialized), took no error, and its **next command in the same transaction** saw the winner's committed key *and* completed replay — READ COMMITTED per-command snapshots deliver exactly the behavior T8-C requires. **No Scope poisoning; no 40001.**

Winner rolls back:

```text
W: BEGIN; INSERT key K10 -> INSERT 0 1; sleep 3; ROLLBACK
L: BEGIN; INSERT key K10 ON CONFLICT DO NOTHING -> INSERT 0 1   <- contender BECAME OWNER
   INSERT replay; COMMIT
```

Same key, different fingerprint:

```text
INSERT ... ON CONFLICT DO NOTHING  -> INSERT 0 0
SELECT fingerprint_equal ...       -> f
ROLLBACK  ->  422 path, zero business mutation, scope alive
```

```text
winner commits  -> loser sees completed replay in a subsequent command   CONFIRMED
winner rolls back -> contender becomes owner                             CONFIRMED
expected race does not poison Scope                                      CONFIRMED
same key + different fingerprint -> conflict, no mutation                CONFIRMED
```

**G. 32-byte semantic fingerprint — security/PII and collision** — see M10. Collision: no authority problem. PII: correction required.

**H. 64 KiB ReplaySnapshot cap** — see L1. Accidental legacy carryover; harmless but undisclosed under T8-A.

**I. Is a single-row completed-only alternative materially smaller? — NO, it is not expressible**

```text
CHECK (payload IS NOT NULL) DEFERRABLE INITIALLY DEFERRED
ERROR:  CHECK constraints cannot be marked DEFERRABLE
```

PostgreSQL defers only UNIQUE, PRIMARY KEY, EXCLUDE and REFERENCES. A single table can therefore express the commit-time completion invariant only through a constraint trigger (rejected by the candidate as more hidden, and contrary to D32) or by committing an `IN_PROGRESS` row (forbidden by T6 §18 and T8-C §18.3). Inserting the row only at completion loses the claim entirely, so both contenders would execute full business work and the loser's unique violation would abort its Scope — a direct T8-C §18.3 violation.

```text
the paired two-table design is the MINIMUM declarative realization on the ratified feature floor
```

---

## 16. River v0.37.1 exact-version verdict

```text
MECHANISM COMPATIBILITY CONFIRMED — PERSISTENCE/PRIVILEGE BOUNDARY INCOMPLETE (BLOCKER-1)
```

Verified against the pinned module source in `GOMODCACHE`, not latest-version documentation.

```text
Config.Schema present and reliable in exactly v0.37.1                    CONFIRMED
  client.go:325-330, validated at client.go:518-521
  CHANGELOG.md:292 — added specifically so schema selection does not rely on search_path
riverdatabasesql accepts the concrete tx selected in T8-C                CONFIRMED
  riverdatabasesql@v0.37.1:48  New(dbPool *sql.DB) *Driver
  riverdatabasesql@v0.37.1:94  UnwrapExecutor(tx *sql.Tx)
  client.go:1765               InsertTx(ctx, tx TTx, args, opts)  -> TTx = *sql.Tx
  exact match for T8-C §3.4 SQLTx(scope) (*sql.Tx, error)
migrations and runtime can use the same custom schema                    CONFIRMED
  rivermigrate Config.Schema (river_migrate.go:71-75) threaded at :167 and :631-734
does River require search_path unexpectedly?                             NO, when Schema is set
  maintenance services default to search_path ONLY when Schema is empty
rehoming to river.* is lower complexity than the default schema          YES
  it removes River from the first-party namespace, so D05's catalog and D27's
  "no raw first-party River SQL" become schema-prefix checks rather than a name blocklist
```

**Unstated couplings the candidate must resolve** — the nightly `REINDEX` default, the absence of `MAINTAIN` on PostgreSQL 16, non-owner REINDEX rejection, missing `CREATE SCHEMA` in River migrations, and the backup/restore coupling (River job state is inside the product database and is restored with it; renditions self-heal via `UNIQUE(submission_id, required_format)`, but this is undisclosed). Full evidence in BLOCKER-1 and L9.

---

## 17. Immutable-history / grant verdict

```text
SOUND IN PRINCIPLE — one class left unprotected (M2), one proof-role gap (M11)
```

```text
INSERT+SELECT-only genuinely protects Audit/Submission/Decision/Release rows        YES
  no UPDATE/DELETE/TRUNCATE privilege = no runtime rewrite path for those tables
column-level UPDATE grants can sustainably protect immutable snapshot columns       YES, with care
  PostgreSQL column-level UPDATE privileges are real and stable; the
  governance_attempt_steps split (selector snapshot frozen, state/activated_at bounded)
  is a correct and maintainable use
required serving operations impossible under the proposed grants                    NONE FOUND
  traced every §27 mutation against the D31 grant list; each immutable table is
  INSERT-only in every flow, and no flow needs to update one
indirect runtime rewrite of immutable history                                        NONE FOUND
  no cross-owner cascade (D29), no trigger (D32), no owner-level privilege (D30)
```

**Two gaps:**

1. **M2** — the malware verdict is the one integrity-bearing fact stored as mutable columns on a table the serving role must be able to update, so it falls outside this protection entirely.
2. **M11** — the property is **false-green if proven from an owner or superuser connection**, because a PostgreSQL table owner's privileges are implicit and unaffected by `REVOKE`. §33 must name the connecting role.

---

## 18. Effectivity model verdict

```text
CORRECT — Revision.state + Release is exactly T2, not duplicate authority; the rejected
release-chain-only model is materially WEAKER, proven rather than preferred
```

```text
duplicate effectivity authority?                 NO
  Revision.state    = T2 §7-§10 explicit lifecycle transitions (ratified semantic law)
  Release           = T2 §9 immutable fact binding the exact winning Submission + predecessor
  the forbidden duplicate is Document.current_status, which T6 §6 bans and the candidate refuses

partial UNIQUE(document_id) WHERE state='EFFECTIVE' provides the final barrier?   YES
  realizes T1 §2 "at most one EFFECTIVE Revision per Document" and
  T2 §9 "No successful externally observable state may contain two EFFECTIVE Revisions"

replacement under Document FOR UPDATE remains atomic?                             YES
  predecessor EFFECTIVE->SUPERSEDED and successor SUBMITTED->EFFECTIVE in one transaction under
  the Document lifecycle root; the partial unique is the declarative last line if the lock is missed

current_submission_id = bounded current state, not a latest pointer?              YES
  exists only while state='SUBMITTED'; cleared by RETURN/withdraw/release/cancel; bound to the
  same Revision by a composite FK; historical Submissions stay immutable and are never reachable
  through it. Correction L3 makes the biconditional a declared CHECK.

one open DRAFT/SUBMITTED Revision required by upstream?                            IMPLIED (L6)

NoHumanApproval + SourceOnly same-tx Submission + Release satisfies all constraints?  YES
  T2 §9 explicitly permits same-transaction Release. Traced: INSERT submission -> UPDATE revision
  (SUBMITTED, current_submission_id) -> INSERT release -> UPDATE revision (EFFECTIVE,
  current_submission_id=NULL). The EFFECTIVE partial unique sees one row; the open-revision
  partial unique is vacated by the same UPDATE; the composite FK is satisfied throughout because
  current_submission_id is set and cleared within the transaction; MATCH SIMPLE makes the cleared
  state trivially valid. No deferral needed.
```

---

## 19. Governance / GROUP-deletion verdict

```text
CORRECT AND EXACTLY T3 §8 — one missing structural barrier (M1), one missing uniqueness (M6)
```

Immutable governance history checked property by property:

```text
no old Submission rewrite on resubmit      YES — submissions is INSERT-only; resubmit inserts a new row
one active Step                            YES — partial UNIQUE(attempt_id) WHERE state='ACTIVE' (D16)
selector snapshot immutability             YES — column grants exclude selector columns from UPDATE
Group candidate snapshot immutability      YES — governance_step_candidates insert-only after activation
empty GROUP snapshot stays empty           STRUCTURALLY INCOMPLETE — see M1
cancellation terminates without fake verdict  YES — attempts.state='CANCELLED' is a distinct value and
                                           no governance_decisions row is fabricated (T1 §3 compliance)
RETURN vs withdraw remain distinct         YES — RETURNED (a decision outcome) vs WITHDRAWN (an
                                           immutable submission_withdrawals row); T2 §8 preserved
immutable tables protected by DB grants    YES — see §17
bounded mutable columns permit no rewrite  YES — state/activated_at/ended_at only
```

**GROUP deletion, attacked aggressively:**

```text
current membership dependency              org.group_memberships FK           -> blocks
current Group RoleAssignment               authz.role_assignments.group_id FK -> blocks
current GovernanceRoute GROUP selector     document_type_governance_steps FK  -> blocks
unactivated live GROUP Step                governance_group_dependencies FK   -> blocks
activated/completed history                group_id_snapshot has NO FK        -> does NOT block
```

Exactly the four T3 §8 blockers, and exactly the T3 §8 release condition. No fifth blocker, no missing blocker.

**Concurrent race analysis:**

```text
delete vs unactivated Step insert
  SUBMIT uncommitted -> DELETE blocks on the group row, then fails the RI check -> fail closed
  DELETE committed first -> SUBMIT's dependency INSERT hits 23503 -> fail closed
delete vs activation
  activation deletes the dependency and freezes candidates in ONE transaction; until it commits
  the dependency row is visible, so DELETE fails; after it commits the Group is legitimately free
  and the frozen candidates reference org.users, which is non-deletable      SAFE
delete vs withdraw / cancel
  both remove unresolved GROUP dependency rows in their own transaction; whichever commits first,
  the other observes a consistent dependency set                             SAFE
activation snapshot vs concurrent membership change
  the resolver's read is snapshot-semantics at activation instant; T2 §6 requires exactly a
  snapshot, and "later Group membership drift does not rewrite the active Step candidate set"
  is satisfied by the insert-only candidate rows                             CORRECT
```

The design is right. What it lacks is M1: nothing structurally stops a decision by a non-candidate, which is also what would make the empty-snapshot rule self-enforcing.

---

## 20. ManagedContent / AdmissionClaim / GC verdict

```text
ARCHITECTURE CORRECT — one reachable lost-bytes race (BLOCKER-2), one proof-class gap (M2)
```

```text
ManagedContent remains mechanism only                         YES — no owner_type/owner_id, no
                                                              Document/Revision/Submission field
semantic rows own descriptor truth                            YES — sha256/size/format on
                                                              WorkingContent, Submission,
                                                              OfficialRendition, DocumentOrigin (T4-A)
provider locator never semantic identity                      YES — explicitly, and provider migration
                                                              may change it without touching descriptors
CLEAN digest belongs to exact immutable bytes                 YES in the gate; NOT protected — see M2
row-existence AdmissionClaim rollback safety                  YES — ConsumeIn is a DELETE inside the
                                                              semantic transaction, so rollback restores
                                                              the live claim (verified by construction)
claim reservation vs GC_PENDING race                          SAFE — a claim reserved on a GC_PENDING
                                                              handle is seen by phase 2, which aborts;
                                                              safe failure is leaked storage
claim consume vs semantic attachment atomicity                SAFE where a claim is used — phase 2 sees
                                                              either the claim row or the committed
                                                              semantic row, never neither
full semantic proof before GC_PENDING and before delete       SPECIFIED — and correctly placed inside
                                                              the same FOR UPDATE transaction
platform performs no ControlledDocs semantic SQL              YES — proofs are routed through the
                                                              ControlledDocs contract per T8-C §20
```

**Attach-vs-delete race that loses governed bytes: FOUND.** Reachable only for an attachment path that does not consume a claim, which the candidate's *"consume AdmissionClaim if applicable"* leaves open. Full construction and the verified fix are in BLOCKER-2.

---

## 21. Malware proof persistence verdict

```text
INSUFFICIENT AS THE SMALLEST *COMPLETE* PROOF — see M2
```

```text
smallest sufficient current proof?          NO — sufficient for the gate, insufficient as proof:
                                            the same role that consumes it can rewrite it
creates a business scan lifecycle?          NO — correctly avoided; the columns are mechanism state
                                            and no scan workflow/owner is introduced
rescan / provider migration invalidation?   YES — a rescan overwrite destroys the admission-time
                                            proof; create-once keeps the BYTES stable but not the
                                            VERDICT RECORD
digest correlation structurally sufficient? YES — malware_digest CHECK length 32 matched against the
                                            semantic sha256 correctly binds the verdict to exact bytes
separate immutable proof row materially stronger?   YES — one insert-only mechanism table with
                                            SELECT+INSERT grants; no new authority, no new owner,
                                            no lifecycle; makes MALICIOUS permanent and any
                                            contradiction visible rather than silent
```

---

## 22. T8-A selective-reuse verdict

```text
DISPOSITIONS SOUND — two claims need correction (M11 scope, L1 disclosure)
```

Revalidated independently against current code rather than trusting the candidate's table. Every PRESERVE/REFINE/REHOME claim tested against the five T8-A proofs (named R10 consumer / no legacy semantic authority / dependency direction fits / proof asserts target property / simpler than rewrite).

```text
auth_identities            DELETE/REWRITE   CORRECT — password/lockout/profile coupling has no
                                            T6 consumer (OIDC-only); fails proofs 1 and 2
auth_sessions              REWRITE          CORRECT — only the durable-current-session+expiry
                                            property survives; T4-M restore invalidation preserved
iam_*                      REHOME/REWRITE   CORRECT — legacy module ownership fails proof 3
role_capabilities          DELETE           CORRECT — T1 §1/T3 make Role→Permission static code
                                            authority; a grant table has no consumer
tenant/RLS/GUC             DELETE           CORRECT — T1 §6 makes pooled tenancy a reopen seam;
                                            see M7 for the residual company_id question
controlled_documents       REWRITE          CORRECT
technical document_revisions DELETE/REWRITE CORRECT — T1 §2 separates Revision from WorkingContent;
                                            the autosave table is not business history
approval_*                 REWRITE          CORRECT — bounded governance relations, not a generic engine
taxonomy / template tables DELETE/fold      CORRECT — T1 §5 absent set
audit_events / hash-chain  REWRITE + DELETE CORRECT — T1 §1 explicitly rejects a Launch global hash chain
audit append-only grants   PRESERVE PROPERTY EVIDENCE-BACKED — db/reference-data/0001:90 records the
                                            durable REVOKE UPDATE/DELETE/TRUNCATE on audit_events;
                                            passes all five proofs.  Proof-role gap: M11
audit 64 KiB bounded facts REFINE/PRESERVE  EVIDENCE-BACKED — db/baseline:986
                                            audit_events_payload_size_cap CHECK <= 65536
current idempotency shape  REWRITE          CORRECT and evidence-backed — the current table carries
                                            status IN ('in_flight','completed','failed') plus
                                            response_status/response_body (db/baseline:1345-1353),
                                            exactly the durable IN_PROGRESS/raw-HTTP state D25 rejects
unique-key concurrency     PRESERVE/REFINE  EVIDENCE-BACKED — postgres_store.go:124 already uses
                                            INSERT ... ON CONFLICT (...) DO NOTHING; the REFINE is
                                            route_template -> canonical operation_id.  See L1 for the
                                            undisclosed 64 KiB carryover on the replay side
River                      PRESERVE/REHOME  CORRECT on mechanism (§16); boundary incomplete (BLOCKER-1)
outboxes / notifications   DELETE           CORRECT — T5 §5 named intents only; T5 §12 no consumer
runtime != DDL identity    PRESERVE         PARTIALLY CORRECT — preserves two of the four classes the
                                            estate actually runs (db/grants/0000_identity_roles.sql).
                                            See M11
table-ownership analyzer   PRESERVE/REWRITE CORRECT and materially improved — the existing
                                            completeness-in-both-directions property
                                            (tools/cilint/.../table_ownership_completeness_test.go)
                                            is preserved while schema qualification makes the
                                            owner mapping derivable instead of hand-maintained.
                                            Catalog gap for river.*: BLOCKER-1
```

**No disposition preserves legacy authority by accident.** The one place where legacy shape leaked without disclosure is the 64 KiB replay cap (L1) — a constant, not an authority.

---

## 23. YAGNI / future-leakage verdict

```text
CLEAN except M7
```

Searched for dormant persistence against T1 §5/§6 and T6 §31:

```text
pooled tenancy        company_id propagation under a singleton CHECK   -> M7 (only real hit)
Distribution          absent
Periodic Review       absent
Dossier / Evidence    absent
Records               absent
notifications         absent
fresh-auth/eSignature absent
generic Change Control absent
body/OCR/vector search absent
materialized Search   absent — no table, no view, no refresh job
generic retention     absent — only bounded mechanism expires_at on claims and idempotency keys
generic Artifact      absent
generic Workflow      absent — D18 explicitly, and governance_attempts uses a real XOR rather
                      than a polymorphic subject_type/subject_id pair
```

The candidate is unusually disciplined here. `governance_attempts`' nullable-XOR instead of a polymorphic pair is exactly "prepare the seam, not the dormant implementation" done correctly, and it is the pattern I would have flagged had it gone the other way.

---

## 24. Zero-semantic-trigger baseline verdict

```text
SUSTAINABLE — CONFIRMED
```

I looked specifically for an accepted invariant that requires a trigger because owner SQL + row locks + constraints cannot structurally close a real race. **None found.**

The two invariants lacking full declarative closure are:

```text
decider in the frozen candidate snapshot     -> closable by an FK (M1), NOT a trigger
Revision.state <-> Release coherence         -> not declaratively closable, but it is a
                                                write-path discipline under Document FOR UPDATE,
                                                not a race; closed by a named proof (L4)
```

Adding a trigger for either would be adding a mechanism because it is possible, which D32's own ladder forbids.

---

## 25. Query ownership / no-foreign-SQL verdict

```text
CORRECT — and the Library/Search claim survives the hardest test
```

```text
Authorization queries only authz.role_assignments while Organization supplies membership
  CORRECT — T3 §2 partition preserved; static Role->Permission expansion stays in Go; no org.* join
ControlledDocs does not join org.* for display or eligibility
  CORRECT — and, critically, NOT broken by T6 §6
Audit filters its own historical attribution
  CORRECT — audit.events carries company_id/area_id, so the visibility predicate is owner-local
  and applied BEFORE cursor pagination, exactly as required
application read composition
  CORRECT — display-name enrichment of an already-fixed page is composition, not a hidden join
GC semantic reference proof through the ControlledDocs contract
  CORRECT — platform never emits ControlledDocs SQL
no cross-owner view becomes hidden shared truth
  CORRECT — D40 permits only owner-private non-materialized views
```

**Library/Search attacked specifically.** If T6 required free-text or ranking over an Organization-owned name, ControlledDocs SQL could not produce a correctly paginated ordered page without a forbidden join, and post-filtering a page is explicitly disallowed. T6 §6 closes this:

```text
q = code + current EFFECTIVE title                      <- no Organization data in free text
ranking: exact code -> code prefix -> title prefix
         -> title contains -> code + stable id          <- no Organization data in ranking
typed filters: Document Type, Area, responsible owner   <- filtered by IDENTITY, resolvable
                                                           against documents.responsible_user_id
                                                           and documents.area_id
```

Every canonical Library predicate and ordering term is owner-local. Only display names require composition. **No break.** This is a real load-bearing property of the design that the candidate states but does not defend; it should be recorded, because it is what makes "materialized Search OFF" survivable.

---

## 26. Genuinely missing operations / tables / invariants

```text
MISSING INVARIANT   decider in frozen candidate snapshot            M1  (FK available)
MISSING INVARIANT   one GovernanceAttempt per governed subject      M6  (2 uniques)
MISSING INVARIANT   SUBMITTED <-> current_submission_id biconditional L3 (CHECK)
MISSING PROOF       no EFFECTIVE/SUPERSEDED/OBSOLETE without Release L4
MISSING PROOF       Go <-> DDL closed-vocabulary parity              M8
MISSING PROOF       grant proofs must name a non-owner connection    M11
MISSING TABLE       insert-only malware verdict proof                M2  (recommended)
MISSING TABLE       provider-subject identity row                    M3  (option (a))
MISSING OPERATION   POST /areas, POST /groups                        M5
MISSING OPERATION   DELETE /role-assignments/{id}                    M5
MISSING OPERATION   PUT /company                                     M5
MISSING OPERATION   draft upload initiate + complete (OPEN->READY)   M5  (most material)
MISSING SPEC        attach-side managed_content lock mode            BLOCKER-2
MISSING SPEC        river schema creation / ownership / grants       BLOCKER-1
MISSING SPEC        idempotency retention delete order               M9
MISSING CLASS       provisioning/bootstrap DB identity               M11
```

---

## 27. Upstream reopen verdicts

```text
T8-C reopen required?      NO
T8-B reopen required?      NO
T1 -> T7 reopen required?  NO
```

Nothing found contradicts a ratified upstream decision.

```text
T8-C  §3 database/sql substrate confirmed compatible with River v0.37.1 InsertTx (primary evidence)
      §18.3 concurrent same-key law empirically satisfied by the candidate's realization
      §18.4 ReplaySnapshot law respected; M10 concerns the FINGERPRINT, which T8-C does not
            govern — it is a T8-D-local correction, not a T8-C defect
      §8 GROUP resolver, §9 cross-owner facts, §20 GC choreography all realized without
            foreign SQL or new authority
T8-B  owner roots, one public surface, platform-as-mechanism all preserved;
      schema namespacing reinforces rather than alters the topology
T1-T7 D33 widens T3 §11's protected-actor census, which T3 permits ("applies at least to"),
      so it is a strengthening inside the ratified boundary, not a reopen (L7 asks for
      justification, not for T3 to be reopened)
L6 is a CITATION gap in the candidate, not a T2 defect
```

BLOCKER-1 is a conflict between two **candidate** decisions and a pinned dependency, not a conflict with upstream authority. It is correctable entirely inside T8-D.

---

## 28. Stage-trespass verdicts

```text
T8-E trespass?   NO
T8-F trespass?   NO
T8-G trespass?   NO
T10 trespass?    NO
T11 leakage?     NO (one boundary observation, L10)
```

```text
exact HTTP JSON schemas / status / headers   NOT decided — explicitly deferred, e.g.
                                             "Exact wire maximum length remains T8-E because T6
                                             explicitly leaves that bound there" and
                                             "Exact wire errors remain T8-E"
frontend realization                         absent
process / deployment topology                absent
current -> target migration / cutover order  explicitly deferred: "Concrete migration/cutover/
                                             deletion order remains T10"
implementation tranche decomposition         absent; "Exact query text is T11-local"
```

§34's non-decision list matches `r10-post-t6-implementation-readiness-program.md` §T8-D/§T8-E exactly.

**One boundary observation (L10):** D38/D39 (explicit SQL, reject ORM baseline) sits close to T11. It is acceptable as framed — the candidate hedges it as a baseline with a named reopen rather than a frozen prohibition — because the persistence-framework axis is named in the T8-D class itself and determines whether per-owner lock/CAS semantics stay visible, which is squarely T8-D's subject. Recommend marking it explicitly as baseline-with-reopen.

Note the corrections requested by this review are all T8-D-native: constraints, lock modes, grants, ownership, and proof obligations. None requires opening T8-E, T8-G, T10 or T11.

---

## 29. Another Fable round?

```text
NOT MATERIALLY REQUIRED
```

The class is confirmed, no upstream reopen is implicated, and the two BLOCKERs and eleven MAJORs are bounded corrections with named, verified fixes — several already validated empirically in this review (attach-side `FOR SHARE`, retention delete order, River ownership options).

A **bounded delta review** of the corrected candidate is appropriate, confined to:

```text
BLOCKER-1  river.* ownership / grants / schema provisioning / D04 MAINTAIN consequence
BLOCKER-2  attach-side managed_content lock mode and/or universal claim consumption
M1         governance_decisions -> governance_step_candidates composite FK
M2         insert-only malware verdict proof relation
M3         provider-subject uniqueness resolution
M4         actor+target User lock ordering
M5         six missing transaction-matrix operations
M6-M11     as listed
```

A full re-review would re-litigate 26 ACCEPT decisions that survived independent challenge with primary evidence.

---

## 30. May Lead adjudication proceed?

```text
YES
```

The candidate is coherent, complete enough to adjudicate, and its class is confirmed. Lead adjudication may proceed on this evidence, subject to:

```text
BLOCKER-1 and BLOCKER-2 must be resolved before T8-D promotion — not deferred to T10/T11
each of M1-M11 must be dispositioned (accept / correct / ratify as deliberate)
promotion still requires explicit operator ratification
T8-E remains NOT OPEN
implementation remains BLOCKED
```

Per Method §3, these findings are **evidence, not requirement authority**. Where a finding proposes something that would create new authority rather than correct a defect against existing authority — specifically M2's proof relation, M3 option (a)'s identity table, and M7's subtraction — the Lead must route it back through decision, not admit it as a correction.

---

## 31. Reviewer method disclosure

```text
authority reconstructed independently from repository files, in the mandated order
prompt / chat history / candidate summaries / prior Lead reasoning treated as non-authority
current code, schema and grants used strictly as evidence for reuse claims
PostgreSQL behavior verified on 16.14 in an isolated throwaway container;
  no MetalDocs or third-party project database was created, modified or read
River behavior verified from the exact pinned module source in GOMODCACHE, not from
  latest-version documentation
Round-1 T8-C savepoint/retry assumptions were NOT carried into the idempotency analysis
no compatibility with current DEV/TEST business data was required of any finding (T7)
```

Empirical results reproduced in this artifact are verbatim tool output.

---

**End of independent Fable review. Reviewer evidence only — not T8-D authority.**
