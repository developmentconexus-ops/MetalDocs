---
id: t8h-global-coherence-review
kind: work
owner: architecture
summary: Branch-only T8-H adversarial review ledger for cross-stage realization coherence.
---

# T8-H — Whole-T8 Global Coherence Review

> **Status:** OPEN / ACTIVE / NOT RATIFIED
> **Opening base:** `main @ 0b4ef6ef891b01f907804cff4bd3c0022aebad80`
> **Implementation:** BLOCKED
> **Scope:** T8-A→T8-G coherence only; T9 is not open.

This is temporary branch-only work under the repository documentation law. It records the active T8-H review and must not enter `main`. Durable corrections, if operator-approved, belong in the owning authorities; final stage/status remains exclusively in `docs/roadmap.md`.

## 1. Review objective and law

T8-H asks one question:

> Do the operator-ratified T8-A→T8-G decisions form one executable technical realization without duplicate or missing authority, incompatible assumptions, removed required seams, or accidental complexity created by the architecture itself?

The review follows DevelopmentConexus Engineering Method v1.0.0. It does not reopen an accepted decision because another shape is preferable. A finding is material only when current evidence touches an invariant, authority/boundary, public contract, persistence meaning, security/trust property, concurrency/recovery property, or user-observable behavior.

The bounded review pack is:

```text
T8-A  docs/architecture/technical-baseline.md
T8-B  docs/architecture/backend.md
T8-C  docs/architecture/interfaces.md
T8-D  docs/architecture/persistence.md
T8-E  docs/architecture/wire-contract.md
T8-F  docs/architecture/frontend.md
T8-G  docs/architecture/runtime.md
```

Additional authority was opened only for concrete seams:

```text
docs/decisions/frontend-read-symmetry.md
  reason: effective DocumentOfficialView differs from the T8-E body

docs/decisions/t8f-ratification.md
+ docs/decisions/t8g-ratification.md
+ docs/decisions/index.md
+ docs/development/documentation.md
  reason: test whether stage/integration state is duplicated outside roadmap
```

Historical implementation, closed review PRs and old `wiki/...` authority paths were not used as target authority.

## 2. Opening revalidation

Observed at T8-H opening:

```text
main                              0b4ef6ef891b01f907804cff4bd3c0022aebad80
main expected SHA                  MATCH
open PRs                           0
T8-H pre-existing branch           none
T8-H branch                        arch/t8h-global-coherence
closeout PR #147                   MERGED
PR #147 head                       ed0ff1c2883d92b65f6502cfa90798abb1cf8ed3
CI #1079                           SUCCESS
CI #1080                           SUCCESS
CI #1080 aggregate job `required`  SUCCESS
```

Remote historical/candidate/review branches remain housekeeping only. Their existence does not open, close or block T8-H.

## 3. Cross-stage attack matrix

| Seam | Attack | Current result |
|---|---|---|
| T8-A → T8-B | Did legacy shape regain authority through package topology? | PASS — owner-first modular monolith was independently rederived. |
| T8-B → T8-C | Can every accepted internal consumer fit the closed dependency classes without a hidden owner/edge? | MATERIAL PRECISION — `application/maintenance` is accepted by T8-C but absent from T8-B's frozen package projection. |
| T8-C → T8-D | Can transaction, Audit, AuthZ, idempotency, GC and durable intent contracts be persisted without foreign SQL or duplicate truth? | PASS. |
| T8-D → T8-E | Can persisted/concurrency truth realize ETags, CSRF, replay and exact-byte wire laws? | PASS. |
| T8-E → T8-F | Can a fresh browser navigation resolve every accepted stable lens from the exact wire? | MATERIAL AUTHORITY SPLIT — effective operation-47 schema is defined by T8-E plus a separate superseding precision. |
| T8-F → T8-G | Does runtime serve exactly the admitted frontend consumers without speculative infrastructure? | PASS. |
| Whole T8 status authority | Is mutable stage/integration state owned only by roadmap? | MATERIAL DUPLICATION — stale mutable status remains in durable T8-F/T8-G and decision records. |
| AuthN / CSRF | Is session bootstrap reconstructible and same-origin runtime-compatible? | PASS. |
| AuthZ | Is there exactly one final ALLOW/default-DENY authority? | PASS. |
| Audit / transaction | Can required evidence append before one local commit? | PASS. |
| Durable jobs | Can River intent participate atomically without a second business-state authority? | PASS; external mechanism fit checked. |
| Idempotency | Are key scope, fingerprint, replay snapshot, 24h expiry and frontend retry semantics compatible? | PASS. |
| Exact content | Do semantic descriptor, storage mechanism, wire proof, frontend consumption and runtime spool agree? | PASS. |
| Official rendition | Do same-PDF reuse and conditional DOCX→PDF behavior agree? | PASS. |
| Search | Is Launch Search still canonical PostgreSQL read behavior with no dormant external engine? | PASS. |
| Operation census | Did any seam require a new application operation? | PASS — 78 remains closed; operation 79 remains absent. |

## 4. Material finding H1 — duplicated mutable program status

### Evidence

Repository documentation governance states that only `docs/roadmap.md` owns mutable stage/gate/implementation status and exact next action.

Current durable documents nevertheless retain mutable progression snapshots that are now false, including:

```text
docs/architecture/frontend.md
  integration through PR #139 remains pending
  T8-G NOT OPEN / remains unopened

docs/architecture/runtime.md
  integration through PR #144 remains pending
  T8-H NOT OPEN

docs/decisions/t8f-ratification.md
  T8-G NOT OPEN / integration remains separate

docs/decisions/t8g-ratification.md
  T8-H NOT OPEN / integration remains separate

docs/decisions/index.md
  T8-F ... INTEGRATION PENDING
  T8-G ... INTEGRATION PENDING
```

These statements do not override roadmap, but their survival creates the exact parallel mutable-status authority that repository governance forbids and gives fresh actors contradictory current-state evidence.

### Root cause

Durable architecture/ratification artifacts captured future program-state snapshots instead of preserving only immutable ratification facts and delegating all mutable stage state to roadmap.

### Smallest correction candidate

```text
preserve immutable ratification/review facts
remove stale mutable downstream-stage/integration snapshots from affected T8-F/T8-G authority and ratification records
remove `INTEGRATION PENDING` from current decision dispositions
replace mutable progression prose with one stable pointer to docs/roadmap.md where needed
```

No Product/T1→T8-G semantic reopen is required.

## 5. Material finding H2 — executable wire authority is split

### Evidence

`docs/architecture/wire-contract.md` declares itself the exact 78-operation application-wire authority. However the effective `DocumentOfficialView` member set is not present there. `docs/decisions/frontend-read-symmetry.md` explicitly supersedes that member set by adding:

```text
open_revision?: OpenRevisionRoutingReference
active_obsolescence_request_id?: Uuid
```

The index compensates by routing fresh actors to both files, and T8-F correctly consumes the precision. Therefore the runtime semantics are not missing, but the exact executable contract has two simultaneous sources.

### Root cause

A valid bounded T8-E precision discovered during T8-F was recorded as a durable superseding patch instead of being folded back into the exact wire SSOT after approval.

### Smallest correction candidate

```text
fold the already-ratified DocumentOfficialView precision and presence/disclosure laws into docs/architecture/wire-contract.md
retain docs/decisions/frontend-read-symmetry.md as immutable ratification/provenance of the bounded correction, not a live schema override
route executable-wire consumers to one exact wire authority
preserve 78 operations, all operationIds, Problems, headers and operation 79 absence
```

This consolidates already-approved meaning; it does not create a new operation, Permission, owner or persistence pointer.

## 6. Material finding H3 — T8-B package projection omits accepted maintenance leaf

### Evidence

T8-B says it freezes backend repository/package topology and projects the accepted `internal/application/*` leaves. T8-C later ratified T5-J GC choreography at:

```text
internal/application/maintenance
```

T8-C explicitly keeps this inside the existing non-semantic `application` class and states that no T8-B reopen is required. The leaf adds no semantic owner, public surface or dependency direction, but the T8-B frozen projection does not show it.

### Root cause

A later accepted concrete consumer refined the contents of an already-admitted application class without reflecting that precision back into the topology projection.

### Smallest correction candidate

```text
add application/maintenance to the T8-B target projection
state its current consumer = T5-J managed-content GC choreography
change no dependency-class law, semantic owner, runtime shell or Product operation
```

This is a bounded coherence precision, not a topology redesign.

## 7. Challenged seams with no material finding

### Historical `wiki/...` header strings

T8-C/T8-D still carry imported `wiki/...` upstream-authority strings. `docs/development/documentation.md` explicitly classifies these as non-navigational provenance. They are therefore not a current authority conflict and are not a T8-H cleanup target.

### River / `database/sql`

T8-C selects the `database/sql` transaction family and exposes the concrete transaction only to catalogued platform mechanisms. T8-D requires durable River intent to share the semantic transaction. T8-G runs River workers in-process.

Current upstream River documentation confirms a maintained `riverdatabasesql` driver for Go `database/sql`; River's client transaction type is `*sql.Tx` for that driver, and transactional insert APIs keep job insertion atomic with the caller transaction. No hidden substrate contradiction was found.

Reference evidence:

```text
https://pkg.go.dev/github.com/riverqueue/river/riverdriver/riverdatabasesql
https://pkg.go.dev/github.com/riverqueue/river
```

### Security / content / state

No second Authorization evaluator, Product lifecycle truth, exact-content authority, durable frontend entity store, Redis dependency, external Search authority, generic outbox/event bus or provider-key Product identity was found across T8-A→T8-G.

## 8. Proposed proof after operator adjudication

If H1–H3 are approved, the bounded correction must prove:

```text
P1  roadmap remains the only mutable stage/status/next-action authority
P2  no affected durable T8-F/T8-G document asserts stale integration/downstream-stage state
P3  exact DocumentOfficialView effective schema exists in wire-contract.md
P4  frontend-read-symmetry.md no longer acts as a second executable schema SSOT
P5  application/maintenance appears in the T8-B target topology with no new class edge
P6  accepted application census remains exactly 78 and operation 79 absent
P7  required CI passes
```

An independent Fable challenge is justified after the material coherence corrections are assembled, because H1/H2 touch authority placement and H3 touches a frozen topology seam. Reviewer findings remain Evidence, not authority.

## 9. Current T8-H decision

```text
T8-H outcome        NOT YET RATIFIED
material findings  H1, H2, H3
Product reopen      NO
T1→T7 reopen        NO
T8-A reopen         NO
operation 79        ABSENT
T9                   NOT OPEN
implementation      BLOCKED
```

T8-H remains active. The next action is operator adjudication of H1–H3 before changing the already-ratified owning authorities.
