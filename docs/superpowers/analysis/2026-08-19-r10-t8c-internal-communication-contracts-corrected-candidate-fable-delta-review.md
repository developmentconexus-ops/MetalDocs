# R10-T8C — Internal Communication Contracts — Corrected Candidate — Bounded Round-2 Fable Delta Review

```text
INDEPENDENT REVIEW EVIDENCE
NON-AUTHORITATIVE
BOUNDED ROUND-2 DELTA REVIEW
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
NOT OPERATOR RATIFICATION
```

> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Workflow:** canonical Standard Fable review workflow, `developmentconexus-ops/conexus-methodology/README.md`
> **Review target:** `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md`
> **Reviewer role:** independent adversarial challenger; output is evidence, never authority

---

## 1. Revalidated reviewed remote HEAD

The expected HEAD supplied in the task prompt was **not** trusted. Remote state was re-established first.

```text
$ git fetch --all --prune
From https://github.com/developmentconexus-ops/MetalDocs
   ffb6125a..6c1d1929  docs/a8-authz-approval-redesign-ledger -> origin/...

$ git rev-parse origin/docs/a8-authz-approval-redesign-ledger
6c1d1929cf975a8aa72f6ebcfcd976bc7abdcf27

$ gh pr view 131 --json headRefName,headRefOid,state,baseRefName
{"baseRefName":"main",
 "headRefName":"docs/a8-authz-approval-redesign-ledger",
 "headRefOid":"6c1d1929cf975a8aa72f6ebcfcd976bc7abdcf27",
 "number":131,"state":"OPEN"}
```

```text
REVIEWED REMOTE HEAD = 6c1d1929cf975a8aa72f6ebcfcd976bc7abdcf27
PR #131 = OPEN, base main
```

Independently revalidated HEAD matches the prompt's expected value. The match was established by fetch, not assumed.

Review executed in an isolated detached worktree at that exact commit. Pre-correction Round-1 HEAD was `ffb6125a`; the corrected candidate landed in `497643b6`, routing in `12d32694` / `686111f0`, handoff in `6c1d1929`.

---

## 2. Confirmation this is a bounded Round 2

Confirmed. This review does **not** re-derive the confirmed model class.

```text
SCOPE       material correction delta R2-1 -> R2-12 only
NOT SCOPE   AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
NOT SCOPE   T8C-D01/D02/D03/D05/D09/D10/D11/D16/D18/D21/D23/D24 as unchanged decisions
```

The unchanged decisions were re-read for **consistency with the delta**, not re-adjudicated. Per the routing constraint, a broad Round-2 re-derivation would have been justified only by a new material contradiction invalidating the confirmed class. **No such contradiction was found.**

### Authority reconstructed independently before reading the delta

Reconstructed bottom-up from the repository at the reviewed HEAD, in the order `AGENTS.md` mandates, before opening the corrected candidate:

```text
developmentconexus-ops/conexus-methodology/README.md          Standard Fable review workflow (fetched)
developmentconexus-ops/conexus-methodology/METHOD.md          v1.0.0 ACCEPTED (fetched)
AGENTS.md                                                     repo bootstrap/router
wiki/references/current-agent-handoff.md                      checkpoint + T8-C route
wiki/architecture/r10-technical-architecture.md               sole stage/status/next-action router
wiki/architecture/r10-t8b-backend-module-package-topology.md  CLOSED/RATIFIED topology law
wiki/architecture/r10-t1-semantic-state-invariants.md          User vs erasable UserProfile
wiki/architecture/r10-t3-authorization-audit-enforcement.md    S2 S9 S11 S14 S16
wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md   T4-D T4-F T4-G T4-H T4-K
wiki/architecture/r10-t5-durable-async-search-external-effects.md     T5-A T5-B T5-J T5-M
wiki/architecture/r10-t6-canonical-api-frontend-journeys.md    S4 S8 S12 S18 S19 S20 S24 S26 S27 S29
```

Round-1 review, Lead adjudication and the task prompt were treated as **claims to be tested**, not as authority.

---

## 3. Delta verdict

```text
APPROVE CORRECTED T8-C DELTA WITH MATERIAL FIXES
```

The corrected delta is a real improvement over Round 1 and closes every operation-census omission Round 1 identified. Every material fix below is closeable **inside T8-C by text**. None changes the selected model class, none reopens T8-B, none reopens T1 -> T7, and none is a stage trespass.

Two of the three most contested Lead positions — **rejecting B5 as a blocker** and **selecting PII-free replay by construction** — are **upheld** on independent primary evidence, not conceded. Both are upheld with an added precision obligation, and in the B5 case the evidence that upholds the Lead is also what exposes the one unnamed constraint that makes the rejection safe.

---

## 4. Finding counts

```text
BLOCKER   0
MAJOR     5
LOW       5
```

```text
SURVIVING MATERIAL CONTRADICTION   0
```

No finding requires the corrected candidate to be rebuilt, and no finding invalidates the confirmed Global Maximum class.

---

## 5. MAJOR findings

### MAJOR-1 — The frozen concurrent-idempotency law is realizable only under READ COMMITTED, and T8-C never says so

**Where:** corrected candidate §14 / §14.1, D19.

The frozen law is:

```text
concurrent same scoped key+fingerprint requests serialize
winner commit -> loser can return completed replay without leaving caller Scope unusable
```

and §14 defers the SQL realization to T8-D with the binding constraint *"A realization that aborts the transaction on an expected same-key race does not satisfy T8-C."*

The behavior the Lead relies on is real, but PostgreSQL's own documentation attaches an isolation-level qualifier that the candidate quotes around and then drops. Primary source, `postgresql.org/docs/current/transaction-iso.html`:

> "INSERT with an ON CONFLICT DO NOTHING clause may have insertion not proceed for a row due to the outcome of another transaction whose effects are not visible to the INSERT snapshot. **Again, this is only the case in Read Committed mode.**"

Under Repeatable Read, same page:

> "…then the repeatable read transaction will be rolled back with the message `ERROR: could not serialize access due to concurrent update`"
> "When an application receives this error message, it should abort the current transaction and **retry the whole transaction from the beginning**."

That retry-the-whole-transaction contract is precisely the application-level retry semantic Round-1 B5 asked T8-C to freeze and the Lead rejected. So the Lead's rejection is correct **only on the condition that T8-D binds this path to READ COMMITTED** (or to some other realization proven to satisfy the law). T8-C freezes an outcome whose realizability depends on a choice it leaves entirely unfrozen *and unmentioned*.

This is not a stylistic gap. The candidate itself establishes the opposite precedent in §3.1: where it knew the realization space was constrained, it named the constraint (`database/sql` family) rather than saying "T8-D must satisfy the contract." The same standard applied here yields one missing line. T8-B §8.1 delegates "isolation/locks/serialization/PostgreSQL mapping" to T8-D wholesale, so nothing upstream supplies the constraint either.

Under the program's binding execution law — *no Writer task may contain a material architecture decision that should have been decided before execution* — an unnamed isolation dependency on which a frozen correctness law silently rests is exactly the class of decision that must not reach a Writer.

**Fix (T8-C text, one clause):** state that the idempotency claim path is frozen at an isolation level at which the concurrent-outcome law holds, name READ COMMITTED as the qualifying baseline, and record that any T8-D election of Repeatable Read / Serializable for that path reopens D19 because it reintroduces the mandatory-retry contract.

**Note:** this finding *strengthens* the B5 rejection rather than reviving B5. See §7.

---

### MAJOR-2 — The `isScope()` seal is not absolute in Go; embedding defeats it, and `SQLTx` has no fail-closed law

**Where:** corrected candidate §3.2 / §3.3, D04.

§3.2 claims: *"The unexported marker means arbitrary first-party packages cannot implement alternative Scope realizations by convenience."* §3.3 claims the platform-only native binding *"removes the current runtime downcast failure class."*

Both claims are partly true and materially overstated. Tested empirically at the reviewed HEAD with a minimal reproduction (`go vet`, repo Go toolchain):

```go
// package txscope
type Scope interface { Exec(q string) error; isScope() }

// package evil — attempt 1: implement from scratch
type FromScratch struct{}
func (FromScratch) Exec(string) error { return nil }
var _ txscope.Scope = FromScratch{}

// attempt 2: embed the sealed interface
type Embedded struct{ txscope.Scope }
var _ txscope.Scope = Embedded{}
```

Result:

```text
attempt 1  vet.exe: evil\neg.go:6:23: cannot use FromScratch{} (value of struct type
           FromScratch) as txscope.Scope value in variable declaration:
           FromScratch does not implement txscope.Scope (missing method isScope)

attempt 2  compiles clean, exit 0
```

The seal blocks a from-scratch realization. It does **not** block struct embedding, which promotes the unexported method into the outer type's method set from any package. Embedding is the *most* convenient escape, not an exotic one.

Two consequences:

1. §3.2's enforcement claim must be narrowed to "cannot implement from scratch," and the embedding case must become an explicit mechanical rejection class. T8-B §10 requires *"negative fixture for each mechanical rejection class"* — so this is a named, falsifiable obligation, not a style note.
2. §3.3 freezes `func SQLTx(scope Scope) *sql.Tx` with **no error return and no defined behavior for a non-target Scope**. Handed an embedded wrapper (or one whose embedded field is nil), the binding must do something; the contract says nothing. That is the runtime failure class §3.3 claims to have removed — narrowed, but not eliminated.

The underlying direction is still correct, and the improvement over the current code is real. Verified current downcast sites at the reviewed HEAD:

```text
internal/modules/approval/jobs/approval_notification_enqueuer.go:33   sqlTx, ok := tx.(*sql.Tx)
internal/modules/approval/jobs/lifecycle_event_enqueuer.go:43         sqlTx, ok := tx.(*sql.Tx)
internal/modules/approval/jobs/release_evaluate_job.go:132            sqlTx, ok := tx.(*sql.Tx)
internal/modules/audit/infrastructure/postgres/writer.go:218          if sqlTx, ok := tx.(*sql.Tx); ok
internal/modules/documents/infrastructure/repository.go:64            sqltx, ok := tx.(*sql.Tx)
```

Two of these sit inside semantic-owner packages (`audit`, `documents`). The target removes those entirely — owners receive `Scope` and call its methods. The remaining need is one platform mechanism. That is a genuine and large reduction.

**Fix (T8-C text):** narrow the §3.2 sealing claim to from-scratch realization; add "no first-party package outside `platform/txscope` may embed `txscope.Scope`" to the cilint proof set with its negative fixture; give `SQLTx` a defined fail-closed outcome for a Scope not created by the target Runner.

---

### MAJOR-3 — The GC coordination law drops the semantic reference re-proof that T4-K and T5-J both require immediately before delete

**Where:** corrected candidate §11.1 / §11.3.

T4-K (ratified) is explicit about the second phase:

> "Immediately before provider delete, execution must re-read/re-prove `GC_PENDING` and absence of **semantic/live references**, live admission claims/bindings and backup protection."

T5-J (ratified) enumerates the same second phase step by step:

```text
claim bounded GC_PENDING set
-> immediately re-prove no current WorkingContent reference
-> immediately re-prove no immutable governed/imported reference
-> re-prove no live admission claim/binding
-> re-prove no backup exclusion/pin
-> provider DeleteReclaimable outside semantic tx
```

The candidate's mechanism-side list in §11.1 carries only:

```text
re-prove GC_PENDING immediately before provider delete
```

and §11.3's coordination law compresses the second phase to an unqualified *"technical GC_PENDING / immediate recheck"*, with the ControlledDocs canonical reference check appearing only once, in the **first** phase.

A Writer reading §11.1 as the mechanism contract and §11.3 as the choreography can legally implement: check references once, mark `GC_PENDING`, later re-read only `GC_PENDING`, delete. That drops a ratified protection from the exact window it was written to cover — the gap between marking and physical deletion. The failure mode is the one T4-K exists to prevent, and its consequence runs against the one asymmetry T5-J calls out as non-negotiable: safe failure is leaked storage, **never** lost governed truth.

**Fix (T8-C text):** make the second ControlledDocs canonical-reference check explicit in both §11.1 and §11.3, matching T5-J's ordering.

---

### MAJOR-4 — The T5-J application host is deferred on a false premise; T8-C must name it (but this is **not** a T8-B reopen)

**Where:** corrected candidate §11.3, final paragraph.

The candidate states:

> "The **exact existing T8-B application leaf** hosting this non-semantic maintenance choreography is not redefined here… If later runtime/package proof shows no existing T8-B application surface can legally host the T5-J invocation without violating the frozen dependency graph, that concrete contradiction reopens only the implicated T8-B leaf decision."

The premise is wrong, and the prompt correctly forbids deferring this. Determined directly.

**A host is structurally mandatory.** T8-B §9.1's allowed-edge list is *"exhaustive at the class/direction level"*. Under it:

```text
transport -> semantic owner            FORBIDDEN  (S5, S9.2)
transport -> platform                  only platform/observability (S9.1)
application -> semantic-owner-public   ALLOWED
application -> platform/txscope        ALLOWED
```

T4-K requires the eligibility proof and the `GC_PENDING` mark to be one transaction spanning mechanism state **and** a ControlledDocs semantic reference check. Only the `application` class can legally hold `semantic-owner-public` + `platform/txscope` together. A job adapter cannot do it directly. So an application leaf must host T5-J. Confirmed.

**No existing leaf fits.** T8-B §2 freezes nine leaves — `session`, `library`, `mywork`, `documentofficial`, `documentwork`, `governancecase`, `history`, `audit`, `admin` — and §4 states they are *"derived from ratified T6 use-case/lens meaning."* T5-J is not a T6 use case: T6 §29's closed census contains no GC route, and T5-A places GC as periodic reconciliation. `admin` is the nearest, and it is a T6 administration lens, not a storage-reclamation surface; hosting T5-J there conflates a product-facing admin lens with a non-semantic maintenance mechanism, which is the conflation T8-B §7 exists to prevent.

**Determination: option A — the topology already admits a legal path. No T8-B reopen.**

Evidence for A over B:

```text
a new internal/application/<maintenance> package maps to the existing `application`
target class; S9 closed-world requires a class, and this needs no new class

every edge it needs is already affirmatively allowed:
  -> semantic-owner-public (ControlledDocs reference check)
  -> platform/txscope      (Scope)
  -> consumer-owned GC mechanism port, bound by composition — no import edge at all

S3.1 gates a *second public package path for one semantic owner* as material
architecture; it gates nothing for application-leaf granularity

S4 anticipates omission directly: "An omitted leaf never creates a
transport -> owner exception" — the remedy for a missing leaf is a leaf,
not a bypass and not a topology reopen

S12's nearest reopen trigger is "a ratified T6 use case cannot pass through an
application leaf without material harm". T5-J is not a T6 use case, so the
trigger does not fire
```

The leaf set is a T6-derived **sufficiency floor for inbound use-case coverage**, not an exhaustive closure of the `application` class.

What *is* wrong is the deferral itself. Because no **existing** leaf fits, the candidate's sentence presupposes a fact that is false, and its escape clause routes a decision to "later runtime/package proof" — that is, to a Writer. The host is load-bearing (both alternative doors are forbidden), so its absence is a material placement decision reaching execution, which the binding execution law forbids.

**Fix (T8-C text):** name the hosting application leaf for the T5-J choreography, and record that it is a non-semantic maintenance leaf outside the T6 lens set — not a new semantic owner, not a product route. No upstream stage changes.

---

### MAJOR-5 — §15.3 states two different reconstruction rules, and they diverge on the two reason-bearing idempotent POSTs

**Where:** corrected candidate §15.3, D20 / D27.

The law reads:

> "T8-E must design the exact success representation for replay-required POSTs so exact status/body replay can be deterministically reconstructed **from the PII-free snapshot** **without querying current mutable PII state**."

Clause 1 is snapshot-only. Clause 2 permits querying anything that is not current mutable PII. These are different rules, and they diverge materially rather than academically.

Two of T6 §18's ten Idempotency-Key POSTs carry free-form reason text as their semantic payload:

```text
POST /api/v1/governance-attempts/{attempt_id}/feedback
POST /api/v1/documents/{document_id}/obsolescence-requests
```

and §15.3 forbids *"free-form governance/cancellation/obsolescence reason text"* in the durable snapshot. So:

```text
reading 1 (snapshot-only)   T8-E must design 201 bodies that never echo the
                            submitted reason text — a real, unnamed constraint
reading 2 (queries allowed) T8-E re-projects the reason from canonical
                            ControlledDocs state at replay time
```

Both are implementable; they are different designs, and T8-E cannot pick correctly from this text.

Separately, the **stated rationale for excluding reason text does not hold**. §15.2 justifies the exclusion by erasure/retention: *"storing erasable PII creates a concrete replay retention root."* But governance and obsolescence reason text is immutable governance evidence — it is retained permanently in the canonical ControlledDocs record and, per T3 §16, in Audit — while replay retention is bounded (T6 §18: *"Replay retention is bounded and exceeds ordinary browser/network retry windows"*). Excluding it from a bounded-retention snapshot removes **no** retention root. The exclusion may still be right on snapshot-minimality grounds, but it is currently carried by a rationale that does not support it, and that mis-attribution is what leaves the reconstruction rule ambiguous.

This finding does **not** touch erasable `UserProfile` data, where the erasure rationale holds exactly and the exclusion is correct.

**Fix (T8-C text):** state one reconstruction rule; and re-base the free-form-reason exclusion on snapshot minimality rather than erasure, or permit non-erasable immutable reason facts in the snapshot.

---

## 6. LOW findings

### LOW-1 — One of the three stated reasons for selecting `database/sql` is not independent evidence

§3.1 lists as a reason: *"River v0.37.1 + riverdatabasesql has a named current consumer requiring `*sql.Tx`."*

Verified from the pinned module source, not from docs:

```text
go.mod
  github.com/riverqueue/river v0.37.1
  github.com/riverqueue/river/riverdriver/riverdatabasesql v0.37.1

river@v0.37.1/client.go:1765
  func (c *Client[TTx]) InsertTx(ctx, tx TTx, args JobArgs, opts *InsertOpts) (...)

riverdatabasesql@v0.37.1/river_database_sql_driver.go:94
  func (d *Driver) UnwrapExecutor(tx *sql.Tx) riverdriver.ExecutorTx      -> TTx = *sql.Tx

riverpgxv5@v0.37.1/river_pgx_v5_driver.go:103
  func (d *Driver) UnwrapExecutor(tx pgx.Tx) riverdriver.ExecutorTx       -> TTx = pgx.Tx
```

River's client is **generic over the driver's transaction type**. River v0.37.1 ships both drivers, and the repo already carries `jackc/pgx/v5 v5.9.2` as a direct dependency. River therefore does not require `database/sql`; it adapts to whichever substrate is chosen. Reason 3 restates the consequence of the selection as a justification for it — under METHOD.md's Structural Inversion Test it flips with the premise and is not load-bearing.

**The decision itself survives** on the two remaining reasons plus a point the candidate does not make: no named Launch consumer requires a pgx-native capability (T5-B leaves materialized Search OFF and T4 routes content through the object store, so no `COPY` / `LISTEN` / native-type consumer exists), and pgx remains usable *as* a `database/sql` driver, so the constraint on T8-D is genuinely small. Recommend restating the reasons; do not reopen D04.

### LOW-2 — `ManagedContentStore` omits T4-D's "create-once / no overwrite" law

T4-D's minimum contract lists `create-once / no overwrite` alongside the operations. §10.1 freezes the operation set (correctly deferring `OpenRange` as unactivated) but attaches no create-once law to `PresignCreate`. T4-H binds a Writer regardless, so this is precision, not a hole — but T8-C is the stage that freezes the port contract.

### LOW-3 — `SearchSubjects` callback has undefined error semantics, and `displayHint` is singular against T6 §4

`emit func(ref string, displayHint string) error` does not say what a non-nil return means — stop iteration cleanly, or abort the search. T6 §4 specifies *"opaque provider_subject_ref + bounded display hints"* (plural); one string may not carry what the selection UI needs. For a bounded result set the callback is also heavier than needed, though the import-free alternatives are materially uglier and the callback is defensible. Record that `emit` must not later acquire streaming semantics.

### LOW-4 — No named falsifiable proof that `AuthorizedScopes` never substitutes for `Decide`

T8-B §10 requires target architecture laws to be falsifiable and says single-source Authorization proofs *"become mechanically exact once T8-C/T8-D provide their contracts."* §5 correctly forbids domain-predicate evaluation inside scope enumeration and §18 requires `Decide` / `DecideMany` for exact candidate actions — but T8-C names no proof for it. That is the one path by which a legitimate prefilter degrades into a de-facto second evaluator, which T8-B §9.2 forbids by name.

### LOW-5 — VersionToken no-op repeat does not say which token is returned

§7 restates T6 §24's *"Exact already-current repeats may be owner no-op and must not fabricate duplicate Audit evidence"* but does not say whether the no-op returns the current or the expected `VersionToken`. Trivial to a human; it is nonetheless the value the client caches next.

---

## 7. Explicit resolution of the B5 Lead / Fable disagreement

```text
LEAD REJECTION OF B5 AS A REQUIRED BLOCKER — UPHELD
ROUND-1 FABLE B5 — NOT SUSTAINED AS A BLOCKER
CONDITION — MAJOR-1 must be closed
```

Resolved from PostgreSQL primary documentation and repository evidence, agreeing with neither party by default.

**Round-1 B5's premise is false as stated.** B5 assumed the losing concurrent request *must* hit a unique violation, aborting its caller transaction, and concluded T8-C must freeze a savepoint or application-retry contract. That holds for a plain `INSERT`. It does not hold for `INSERT … ON CONFLICT DO NOTHING`, which `postgresql.org/docs/current/sql-insert.html` describes as the alternative action: *"ON CONFLICT DO NOTHING simply avoids inserting a row as its alternative action."*

**Each of the prompt's required sub-questions, answered from primary evidence:**

```text
Q  Can T8-D satisfy the frozen behavior without a T8-C savepoint/retry contract?
A  YES — at READ COMMITTED. See MAJOR-1 for the qualifier.

Q  Does ON CONFLICT-based realization avoid the assumed transaction-abort path?
A  YES. DO NOTHING raises no error, so the loser's Scope stays usable.

Q  After DO NOTHING due to a just-committed concurrent row, can a subsequent
   statement in the same READ COMMITTED transaction obtain the completed replay?
A  YES, directly documented:
   "Because Read Committed mode starts each command with a new snapshot that
    includes all transactions committed up to that instant, subsequent commands
    in the same transaction will see the effects of the committed concurrent
    transaction in any case."

Q  If the winner rolls back, can the contender become the insert winner?
A  YES. The documented behavior is outcome-dependent, not existence-dependent:
   "INSERT with an ON CONFLICT DO NOTHING clause may have insertion not proceed
    for a row due to the OUTCOME of another transaction whose effects are not
    visible to the INSERT snapshot."
   Insertion not proceeding is conditioned on the other transaction's outcome;
   on abort, the waiter's insertion proceeds and it becomes claim owner.

Q  Is any application-visible retry semantic still necessary?
A  NO at READ COMMITTED. YES at Repeatable Read / Serializable, where the docs
   require "retry the whole transaction from the beginning" — hence MAJOR-1.

Q  Are there other concurrency cases where the Scope becomes unusable?
A  Yes, but all are excluded by the frozen law itself: plain INSERT (unique
   violation), and RR/SERIALIZABLE (serialization failure). S14's constraint
   "A realization that aborts the transaction on an expected same-key race does
   not satisfy T8-C" correctly forecloses both. MAJOR-1 asks only that the
   isolation half be named rather than left implicit.

Q  Does semantic fingerprint mismatch remain distinguishable safely?
A  YES. After DO NOTHING the row is readable in a subsequent command; the stored
   fingerprint is compared and mismatch yields conflict with no mutation and no
   abort. T6 S18's "same key + different fingerprint -> 422" is satisfiable.
```

**Repository evidence independently corroborates realizability.** The current implementation already runs exactly this pattern:

```text
internal/platform/idempotency/postgres_store.go:124
  ON CONFLICT (tenant_id, actor_user_id, route_template, key) DO NOTHING

postgres_store.go:117-134  err == nil -> won; sql.ErrNoRows -> lost, no error, no abort
postgres_store.go:138-141  resolveExistingSlot "blocks on the winner's row-level lock
                           (FOR UPDATE) until the winner commits or rolls back,
                           then reacts to the row's terminal state"
```

The loser path is already non-aborting and already reads the winner's terminal state. This is running evidence that the frozen outcome is realizable without a savepoint or application-retry contract — while the current durable `in_flight` status remains correctly rejected by the target (§22: "expected races must not abort target Scope"; D19: "no public/durable IN_PROGRESS business state").

**Net:** the Lead was right to reject B5 as a required blocker, and right that the SQL realization belongs to T8-D. Round-1 Fable was right that something must be named — but it named the wrong thing. The load-bearing omission is not savepoint-versus-retry; it is the **isolation level** on which the non-aborting path depends. That is MAJOR-1, and it is a one-clause fix.

---

## 8. Explicit verdict on the PII-free replay decision

```text
SELECT B — PII-FREE REPLAY SNAPSHOT BY CONSTRUCTION   UPHELD
REJECT baseline replay-purge/redaction subsystem      UPHELD
CONDITION — MAJOR-5 must be closed
```

**The decision is in-boundary.** T8-B §7.6 assigns it explicitly: *"replay persistence must be erasure-safe and not become a PII retention root… Exact replay contract/schema/retention and PII-free-vs-purge realization remain T8-C/T8-D."* T6 §19 states the requirement without mandating a mechanism: *"Replay storage must not become an unintended retention root for erasable UserProfile PII; response persistence/redaction must remain compatible with T3/T4 privacy/restore laws."* T8-C making this selection is the assigned work, not a trespass.

**Each of the prompt's required attack points:**

```text
Q  Does this really satisfy T6's exact status/body replay?
A  YES, conditional on T8-E. T6 S18 requires "completed replay result sufficient
   for exact status/body replay"; T6 does not freeze response bodies — T6 S29
   freezes paths/operations only — so T8-E retains the freedom the law needs.

Q  Does POST /users force erasable profile data into its success result anywhere
   in current authority?
A  NO. T6 S4 requires the transaction to CREATE a UserProfile
   (BEGIN -> User -> required UserProfile -> ProviderSubjectBinding -> Audit COMMIT)
   but nowhere requires the 201 body to CONTAIN it. The creating client already
   holds the profile it submitted. Confirmed against T6 S4, S26 and S29.

Q  Can T8-E legally choose a success representation of only stable non-erasable
   result facts without reopening T6?
A  YES. T6 S26 fixes read models by name (UserReference etc.) and list envelopes,
   not POST success bodies.

Q  Is ReplaySnapshot becoming a hidden wire DTO?
A  NO — but only under reading 1 of S15.3. This is precisely the ambiguity in
   MAJOR-5: a snapshot-only rule pushes T8-E's body shape toward the snapshot's
   shape and needs the boundary stated deliberately.

Q  Can exact replay survive deployment/version changes within replay retention?
A  YES. S15.3 requires a "versioned operation-local snapshot", and T6 S18 bounds
   retention. Versioning plus bounded retention is the correct minimum.

Q  Is re-projection of current state accidentally required?
A  Only for the two reason-bearing POSTs, and only under reading 2 — MAJOR-5.
   For all other replay-required POSTs, stable ids plus completion-time result
   facts suffice with no re-projection.

Q  After UserProfile erasure, can the original replay still be returned exactly?
A  YES, and this is the decision's strongest property. DELETE /users/{id}/profile
   is a real Launch operation (T6 S29). Under option A the erasure would have to
   reach into replay storage; under option B there is nothing to reach.
   T1 already separates stable User identity from erasable UserProfile, so
   PII-free-by-construction follows the ratified semantic split rather than
   fighting it.

Q  Are any other required Idempotency-Key POSTs inherently PII-bearing?
A  NO. Walked all ten in T6 S18. Eight carry only stable semantic ids and
   configuration values. The two reason-bearing ones carry free-form text, which
   is governance evidence, not erasable UserProfile PII — see MAJOR-5.

Q  Is a purge/redaction mechanism actually unavoidable?
A  NO. No current Product or T6 requirement forces erasable PII into an exact
   durable replay response. I looked specifically for one and did not find it.

Q  Is the selected shape the Global Maximum under YAGNI?
A  YES. Option A buys a cross-owner purge subsystem coupling platform/idempotency
   to Organization erasure, with new partial-failure and restore-reconciliation
   modes against T4-N — to protect data that need never be stored. Option B
   removes the defect class structurally rather than guarding it, which is the
   Method's stated preference.
```

No current Product / T6 requirement forces erasable PII into an exact durable replay response. The selection stands.

---

## 9. Explicit verdict on the GC application-path / possible T8-B leaf contradiction

```text
DETERMINATION           A — the current T8-B topology already admits a legal
                        cohesive application path
T8-B REOPEN             NO
T8-C ACTION REQUIRED    YES — name the hosting leaf; do not defer (MAJOR-4)
```

Full reasoning in MAJOR-4. Summary of the decision actually made rather than deferred:

```text
an application leaf is structurally MANDATORY
  transport -> owner and transport -> platform (beyond observability) are both
  forbidden by T8-B S5/S9.1, and T4-K requires mechanism state and a
  ControlledDocs reference check inside one transaction

no EXISTING leaf fits
  the nine frozen leaves are T6 use-case/lens derived; T5-J is periodic
  reconciliation with no T6 route (T6 S29, T5-A/T5-B)

but the CLASS already permits one
  a new application-class package needs no new class and no new edge; every
  edge it requires is already affirmatively allowed, and the consumer-owned
  GC port is bound by composition, creating no import edge at all

and T8-B ANTICIPATES this
  S4:   "An omitted leaf never creates a transport -> owner exception"
  S3.1  gates only a second public path for a semantic OWNER
  S12's nearest trigger is T6-use-case-specific and does not fire

therefore the leaf set is a T6-derived sufficiency floor for inbound use-case
coverage, not an exhaustive closure of the application class
```

No T8-B leaf decision is implicated, so nothing upstream reopens. What must change is one sentence of T8-C.

---

## 10. Disposition R2-1 -> R2-12

```text
R2-1   txscope / River / database-sql substrate     CLOSED WITH PRECISION   MAJOR-2, LOW-1
R2-2   Audit read contract                          CLOSED / ACCEPT
R2-3   Authorization AuthorizedScopes               CLOSED WITH PRECISION   LOW-4
R2-4   ManagedContent claims + GC                   CLOSED WITH PRECISION   MAJOR-3, MAJOR-4
R2-5   eligibility serialization                    CLOSED / ACCEPT
R2-6   owner VersionToken / If-Match boundary       CLOSED WITH PRECISION   LOW-5
R2-7   ProviderClient correction                    CLOSED WITH PRECISION   LOW-3
R2-8   malware exact-byte correlation               CLOSED / ACCEPT
R2-9   B5 idempotency concurrency                   CLOSED WITH PRECISION   MAJOR-1
R2-10  PII-free replay snapshot                     CLOSED WITH PRECISION   MAJOR-5
R2-11  OfficialRendition content read               CLOSED / ACCEPT
R2-12  operation-census delta closure               CLOSED WITH PRECISION   LOW-2

SURVIVING MATERIAL CONTRADICTION                    NONE
```

### R2-1 — CLOSED WITH PRECISION

Substrate selection is justified and correctly framed as a deliberate constraint rather than an accident. It does **not** improperly overconstrain T8-D: T8-D keeps driver, pool, statement and schema choice, and pgx remains usable beneath `database/sql`. Preserving a pgx-native option would **not** reduce total complexity — it would require custom Row/Rows abstractions to keep both live, with no named Launch consumer (LOW-1). Sealing works against from-scratch realization but not embedding (MAJOR-2). `platform/postgres` can supply the concrete transaction without an illegal dependency: `platform -> platform` is allowed as a technical DAG, and composition can construct the Runner, so no cycle is forced and no T8-D detail leaks into T8-C. Platform-only `SQLTx` is mechanically enforceable — `application` and owners may both import `platform/txscope`, so an import rule is insufficient and the candidate correctly specifies a **call-site allowlist** in `tools/cilint`, which the T8-B closed-world model hosts. Application retains no unsafe SQL capability beyond what static enforcement covers, and static call-site rejection is the strongest reasonable mechanism available — Go cannot express "may hold this value but not call its methods". No materially stronger contract at lower total complexity was found.

### R2-2 — CLOSED / ACCEPT

Preserves T3 §14 exactly, including the `audit.read @ Company -> Company events + all Area-attributed events` / `@ Area X -> only events historically attributed to Area X` split and the "current relocation never rewrites visibility" law. Audit remains the sole authority for historical attribution; Authorization remains the sole authority for current grants; application only maps between the two — it evaluates nothing. Filter-before-pagination is stated as a law (*"Audit applies historical attribution filter before pagination"*, *"application never post-filters a paginated Audit page"*), which is the correct placement. The Audit-owned opaque `PageCursor` is sufficient and steals no T8-E wire syntax; `EventPage{Events, NextCursor, HasMore}` maps onto T6 §26's envelope without freezing it, and no limit defaults are frozen. Company-wide and multi-Area visibility are representable without duplicated semantics: `ReadVisibility` and `AuthorizedScopeSet` are near-identical in shape but distinct in meaning (historical-attribution input vs current-grant output), and merging them would require either a shared contracts package or an owner -> owner import — both forbidden. The duplication is essential, not accidental. No required Audit read/filter operation is missing: T6 §29 exposes only `GET /api/v1/audit/events`, T6 §26 forbids a generic filter/sort DSL and binds filter semantics to the cursor, so the candidate's refusal to invent filters absent a ratified consumer is correct.

### R2-3 — CLOSED WITH PRECISION

`AuthorizedScopes` is genuinely the same authority, not a second evaluator: it lives inside Authorization, reuses the same direct/group RoleAssignment resolution, the same static Role -> Permission bundles and the same scope semantics, and answers a different question ("where is this actor granted this permission?") rather than re-deriving ALLOW. It is a projection of a prefix of T3 §2's equation, deliberately stopping before the Controlled Documents predicates. Sufficiency verified against all three named consumers: T6 §8's creation options map exactly onto `CompanyWide -> all current ACTIVE eligible Areas` / `Area-scoped -> only matching current ACTIVE Areas`; audit visibility maps via R2-2; Library prefilter maps via §18. It solves filter-before-pagination coherently by moving the scope restriction ahead of the owner's canonical paginated query. It correctly omits domain predicates, which stay owner-authored — T6 §8's Template / DocumentType / D4-owner conditions run through `Decide` / `DecideMany`, and a second `AuthorizedScopes(document.owner.manage)` call covers the owner-candidate gate without any application-side role reasoning. `DecideMany` correspondence is now explicit (`len(decisions) == len(checks)`, index-exact, mismatch = internal error). Precision gap only: no named falsifiable proof that the prefilter never substitutes for the decision (LOW-4).

### R2-4 — CLOSED WITH PRECISION

T4-D is satisfied — `Allocate` / `PresignCreate` / `Stat` / `OpenExact` / `CopyToNewHandle` / `DeleteReclaimable` covers the minimum contract, with `OpenRange` correctly deferred as unactivated per T4-D "optional" and T6 §20 "activated only on real viewer/file-size evidence" (a correct YAGNI call, not an omission); create-once is unstated (LOW-2). T4-F is satisfied: the claim binds one handle to one in-flight authorized attachment, is opaque/unforgeable, protects against GC eligibility while live, and introduces **no** `owner_type/owner_id` registry — verified against T4-F's explicit prohibition. Claim consumption is correctly coupled to semantic attachment (`ConsumeIn` atomic in the caller Scope), and rollback stays safe in the right direction: no committed semantic attachment without its consumption. Bounded expiry is mechanism-only, with duration and persistence deferred — it creates no retention or business authority. `DeleteReclaimable` remains mechanism-only and no governed-content delete path appears, per T4-K. No ManagedContent/Artifact semantic owner is resurrected, directly or indirectly — T8-B §3.2's prohibited peer list is respected. Schema, lease and `GC_PENDING` realization are correctly left to T8-D; periodic scheduling correctly left to T8-G. Two gaps: the pre-delete semantic re-proof (MAJOR-3) and the deferred host (MAJOR-4).

### R2-5 — CLOSED / ACCEPT

Sufficient to express T3 §11. The guarantee — *"returns current User eligibility + GroupMembership facts AND serializes that User's eligibility against concurrent offboarding/eligibility-disable until caller Scope completes"* — states the required outcome without naming a mechanism, so it freezes no lock: row lock, advisory lock and predicate locking all remain open to T8-D, matching T8-B §8.1's delegation. The `SecuritySubjectIn` / `ProtectedSecuritySubjectIn` naming is precise enough that a Writer cannot silently pick the weaker form, because the protected obligation is attached to named operations rather than left to judgment. All five T3 §11 operations are covered, in both actor and target directions: Session issuance, governance ACCEPT / RETURN_FOR_CHANGES and Submission / withdraw / cancel / obsolescence use the protected **actor** read; GroupMembership add is same-owner and serializes the **target** internally without a public contract — correctly, since exposing a cross-owner contract for an intra-owner concern would be unnecessary surface; and new direct User RoleAssignment gets protected **target** eligibility through `RoleAssignmentTargetFactsIn`. Responsible-owner target eligibility is protected per the D4 amendment. The T3 §11 outcome table (action-first commits under valid pre-offboarding state; offboarding-first fails closed) is satisfiable under this contract.

### R2-6 — CLOSED WITH PRECISION

Concurrency meaning stays in the owner: the owner issues the token, the owner compares it, and the owner raises the stale/precondition error. HTTP syntax stays wholly in T8-E — `VersionToken` is declared semantically opaque outside the owner, and quoting/format/header encoding are explicitly listed as T8-E's. T8-C does **not** over-specify persistence representation: generation and storage are left to T8-D, so row version, `xmin`, monotonic counter and hash all remain open. One opaque token concept is sufficient across the T6 §24 set — verified item-by-item against §24's list, which matches the candidate's eleven exactly. DRAFT is correctly excluded: T6 §12 already ratifies `ETag: "draft-<generation>"` over WorkingContent generation as its own OCC token, and inventing a second token there would create the duplicate authority the Method forbids. Only LOW-5 remains.

### R2-7 — CLOSED WITH PRECISION

Platform retains raw protocol handling as T8-B §7.3 requires: discovery, token exchange, JWKS and protocol verification stay in `platform/identityprovider`, and only verified `issuer` / `subject` cross the seam. Authentication retains `ProviderSubjectBinding` and `ApplicationSession` meaning. Platform can satisfy the interface **without importing Authentication** — verified: every parameter and result is stdlib or primitive (`context.Context`, `string`, `func(string,string) error`), and Go's structural interface satisfaction requires no import of the declaring package. Issuer + subject are sufficient for Launch identity truth: T6 §4's flow resolves the subject first and then carries profile data in the POST body, so no provider enrichment is needed at binding time. Provider roles/groups/permissions cannot leak — there is no channel for them, satisfying T8-B §7.3 and T3 §2's prohibition on provider-role mapping. No raw claim is required by any current named consumer; the deferral clause for a future assurance consumer is a seam, not dormant capability. Correctly does **not** create a shared DTO/contracts package. Callback ergonomics only (LOW-3).

### R2-8 — CLOSED / ACCEPT

Structurally binds CLEAN to the exact inspected bytes, which is precisely T4-G's invariant: *"Untrusted external bytes cannot become immutable governed MetalDocs content without successful malware inspection of those exact immutable bytes."* Because the digest is computed by the scanner over the stream it actually read, a verdict cannot be transplanted onto different bytes — the correlation is structural rather than procedural, which is the stronger form. A raw `[32]byte` SHA-256 is the smallest sufficient correlation: T4-A's `ExactContentDescriptor` already carries `sha256`, so the comparison reuses an existing fact and introduces no new identity. No redundant second hash or second read is forced — the scanner hashes the stream it is already consuming. The scanner remains mechanism-only, owning no Document/Submission meaning and no scan lifecycle, and trust-class selection correctly stays with Controlled Documents as T4-G's "server-selected" classes require.

### R2-9 — CLOSED WITH PRECISION

See §7. Lead rejection upheld on primary evidence; MAJOR-1 is the one required fix.

### R2-10 — CLOSED WITH PRECISION

See §8. Selection upheld; MAJOR-5 is the one required fix.

### R2-11 — CLOSED / ACCEPT

Covers T6's distinct rendition resource `GET /api/v1/official-renditions/{rendition_id}/content` (T6 §29), which Round-1 M6 correctly flagged as uncontracted. Current-versus-historical authorization is correct against T3 §9: effective read requires `document.read_effective` **and** target Revision = current EFFECTIVE, while historical read requires `document.read_history` over authorized immutable history — so an EFFECTIVE rendition cannot be reached through the historical path, nor a superseded one through the effective path. Governance continues to use exact Submission source, matching T6 §20's *"Governance -> exact Submission source primary decision content"*, and the rendition is explicitly not treated as Release source. The provider handle stays hidden, satisfying T6 §20's *"never exposes provider bucket/key/version or managed_content_id as product identity."* The flow ordering is right — identity, then owner scope/access facts, then Authorization, then bytes — so authorization precedes disclosure. No read contract is missing: all four T6 §20 byte resources are covered, three by the unchanged source-read contracts and the fourth by this addition.

### R2-12 — CLOSED WITH PRECISION

Every family Round 1 recorded as missing or unfrozen is now closed. Walked against Round-1 §16's list, T6 §29's census, T6 §24's If-Match list, T4-D's minimum contract and T5-J's activated family:

```text
Audit list/read                        CLOSED  S4.2   Audit.ListEvents + ReadVisibility
document-creation scope enumeration    CLOSED  S5     Authorization.AuthorizedScopes
OfficialRendition content read         CLOSED  S17    OfficialRenditionContent
admission claim lifecycle              CLOSED  S10.2  Reserve/ProveLive/ConsumeIn/Release
DeleteReclaimable                      CLOSED  S10.1  ManagedContentStore
T5-J GC contract family                CLOSED  S11    with MAJOR-3, MAJOR-4
If-Match owner token                   CLOSED  S7     11 resources = T6 S24 exactly
eligibility serialization              CLOSED  S6     ProtectedSecuritySubjectIn
provider issuer+subject split          CLOSED  S8     verified primitives
malware byte correlation               CLOSED  S12    returned digest
replay PII/erasure                     CLOSED  S15    with MAJOR-5
RoleAssignment POST idempotency        CLOSED  S19 L1 BeginIn/CompleteIn + snapshot
DecideMany ordering                    CLOSED  S5     index-exact correspondence
responsible-owner read                 CLOSED  S19 L3 + VersionToken
template-role read                     CLOSED  S19 L3 + VersionToken
obsolescence completion host           CLOSED  S19 L4 no new command invented
provider-binding replacement wording   CLOSED  S19 L5 no unratified disable API
```

```text
OPERATION-CENSUS DELTA:  COMPLETE
```

No Writer must invent a material internal contract for any of these. LOW-2 is the only census-adjacent precision item, and T4-H binds regardless.

---

## 11. Genuinely NEW material findings introduced by the correction

Two, both surfaced only by the corrected text and neither reachable from Round 1:

```text
NEW-1   The corrected B5 resolution silently depends on an unfrozen isolation
        level. Round 1 could not have found this: the corrected candidate is
        what introduces ON CONFLICT-based realizability as the reason for
        rejecting the savepoint/retry contract. The PostgreSQL qualifier
        "only the case in Read Committed mode" is what makes the rejection
        conditional.                                                   -> MAJOR-1

NEW-2   The corrected txscope seal is newly claimed as an enforcement property.
        Round 1's B1 was about River incompatibility, not sealing. The seal
        is new text, and its stated strength does not survive an empirical
        Go test.                                                       -> MAJOR-2
```

MAJOR-3, MAJOR-4 and MAJOR-5 are precision defects **within** newly added corrected text rather than newly created risks: each is a place where the correction closed a Round-1 omission but left the closing text under-determined.

Notably, the corrections introduced **no** new authority, no new owner, no new framework and no new speculative capability. That is the outcome an adjudicated correction should have, and it is worth recording positively.

---

## 12. Stage-boundary test

```text
T8-B reopen required?                    NO
T1 -> T7 reopen required?                NO
T8-D trespass?                           NO
T8-E trespass?                           NO
new semantic owner?                      NO
new generic framework?                   NO
future / Launch+ speculative contract?   NO
```

**T8-B reopen — NO.** The only candidate was the T5-J host (MAJOR-4), determined as option A: the topology already admits a legal path within the existing `application` class, requiring no new class, no new edge and no gate T8-B defines. None of T8-B §12's reopen triggers fires. Every delta contract respects the closed-world edge law: owner -> owner absent, transport -> owner absent, application -> platform limited to `txscope` / `idempotency` / `observability`, all other mechanism access through consumer-owned seams bound by composition.

**T1 -> T7 reopen — NO.** Every delta contract was checked against its upstream anchor and each preserves it: T3 §2, §9, §11, §14; T4-D, T4-F, T4-G, T4-K; T5-A, T5-B, T5-J; T6 §4, §8, §12, §18, §19, §20, §24, §26, §29; T1's User / UserProfile split. No delta requires an upstream semantic change. Where the delta narrows upstream freedom — for example requiring T8-E to design PII-free-reconstructible success bodies — it does so within space T6 left open, not by contradicting a ratified law.

**T8-D trespass — NO.** The `database/sql` family selection is a declared substrate **constraint**, not a persistence design, and T8-B §8.1 already assigns "isolation/locks/serialization/PostgreSQL mapping" to T8-D while leaving the participation contract to T8-C. Schema, indexes, lock clauses, statements, expiry storage, `GC_PENDING` tables and pool configuration all remain T8-D's, and §24's split is accurate. Worth stating plainly: MAJOR-1 asks T8-C to add **more** constraint on T8-D, not less — and that constraint is required precisely because T8-C froze a behavior whose realizability is not self-evident.

**T8-E trespass — NO.** `VersionToken` and `PageCursor` are declared opaque; no ETag quoting, header grammar, JSON shape, status code or `Idempotency-Key` wire syntax is frozen; T6 §26's list envelope is mirrored structurally without being fixed, and limit defaults are untouched. The obligations placed on T8-E (design replay-required success representations that are reconstructible; map `VersionToken` to strong ETag) are contract requirements, which is what a communication-contract stage is for. MAJOR-5's ambiguity is a defect in how one such obligation is worded, not an encroachment on T8-E's design space.

**No new semantic owner.** The five Launch homes are unchanged. No Artifact, ManagedContent, Retention, Records, Approval or Search owner appears directly or indirectly; ManagedContent and its claim lifecycle stay mechanism-only.

**No new generic framework.** No EventBus, generic outbox, UnitOfWork, ServiceLocator, generic Repository family, generic policy language or shared contracts package. `AuthorizedScopes` is one bounded query, not a policy DSL.

**No speculative capability.** `OpenRange` deferred until a real viewer/file-size consumer; no raw claim bag; no replay purge subsystem; no standalone provider-binding disable API; no obsolescence completion command. Each deferral names its activation trigger — seam prepared, capability not built.

---

## 13. Required verdict fields

```text
revalidated reviewed remote HEAD    6c1d1929cf975a8aa72f6ebcfcd976bc7abdcf27
bounded Round 2                     CONFIRMED
delta verdict                       APPROVE CORRECTED T8-C DELTA WITH MATERIAL FIXES

BLOCKER                             0
MAJOR                               5
LOW                                 5

B5 resolution                       LEAD REJECTION UPHELD; Round-1 B5 not sustained
                                    as a blocker; conditional on MAJOR-1
PII-free replay verdict             SELECTION UPHELD; conditional on MAJOR-5
GC / T8-B leaf verdict              OPTION A — legal application path exists;
                                    NO T8-B reopen; T8-C must name the host (MAJOR-4)

Global Maximum class confirmed      YES
T8-B reopen                         NO
upstream T1 -> T7 reopen            NO
T8-D trespass                       NO
T8-E trespass                       NO
operation-census delta complete     YES
another Fable round required        NO
final Lead adjudication may proceed YES
```

### Why another Fable round is not materially required

All five MAJOR findings are closeable by bounded text edits inside T8-C, and none is a design question:

```text
MAJOR-1  add one isolation-level clause to D19
MAJOR-2  narrow one sealing claim; add one cilint rule + negative fixture;
         define SQLTx fail-closed behavior
MAJOR-3  restore the semantic re-proof step in S11.1/S11.3
MAJOR-4  name the T5-J host leaf instead of deferring
MAJOR-5  state one reconstruction rule; re-base one exclusion's rationale
```

None reopens a decision, changes an owner, moves a boundary or alters the model class. Method §4's stop condition is met: evidence is sufficient for the claims, root cause and invariants are clear, credible alternatives were compared for every contested selection, complexity / authority / proof were checked, the strongest objections were raised and answered, and no material contradiction remains. A third round would re-confirm an already twice-confirmed model.

### Reviewer summary

The corrected candidate is a materially better artifact than the Round-1 candidate. It closes every operation-census omission Round 1 found, and it does so without adding an owner, a framework or a speculative capability — the failure mode most available to a correction pass. The two contested Lead positions survive independent challenge on primary evidence rather than deference.

The residual weakness is consistent and has one shape: **where the correction closed a gap, it sometimes stated the outcome without stating the condition the outcome depends on.** MAJOR-1 (isolation level), MAJOR-2 (what sealing actually prevents), MAJOR-3 (which re-proof), MAJOR-4 (which host) and MAJOR-5 (which reconstruction rule) are five instances of one pattern, not five unrelated defects. Recording the pattern is more useful than recording the instances: a corrected candidate written under adjudication pressure tends to assert the corrected conclusion and under-specify the premise that makes it true — which is exactly the class the program's binding execution law exists to catch before a Writer inherits it.

Reviewer output is evidence, never authority. Lead confrontation and explicit operator ratification remain required before any durable T8-C promotion.

---

```text
END OF BOUNDED ROUND-2 FABLE DELTA REVIEW
NON-AUTHORITATIVE REVIEW EVIDENCE
```
