# R10-B1 Corrected Target — Independent Fable Bounded Delta Review

> **Status:** INDEPENDENT BOUNDED DELTA REVIEW OF RECORD — evidence, **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** Claude Fable 5 — same independent reviewer as the R10-B1 review of record; bounded delta scope per adjudication §13
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Corrected candidate:** `docs/superpowers/analysis/2026-08-17-r10-b1-fable-adjudication-corrected-target.md` @ `92cba574`
> **Prior review of record:** `docs/superpowers/analysis/2026-08-17-r10-b1-independent-fable-review.md` @ `b38f598b`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** the seven delta questions in adjudication §13 only. No broad substrate re-review was warranted: the corrections introduce no new material contradiction (§4 below).
> **Implementation gate:** CLOSED — this review authorizes nothing.

---

# 1. Verdict

```text
VERDICT: APPROVE R10-B1 CORRECTED TARGET
```

```text
BLOCKER = 0
MAJOR   = 0
LOW     = 2 (non-blocking clarity observations; no adjudication required before promotion)
```

All six prior findings are CLOSED by the corrected text. No correction created a
new semantic owner, a material contradiction, or R9.5/R10-A reopen evidence.

---

# 2. Prior-finding disposition

```text
F1. CLOSED — corrected §3 + §4 + §5 — isolation-law scope now bound to semantic
    persistence class; durable intent explicitly DURABLE MECHANISM with a
    bounded claim surface and an R10-D no-content-bypass constraint.
F2. CLOSED — corrected §2.4 + §11.2 — closed positive allowed set
    (RESTRICT / NO ACTION, DELETE and UPDATE); CASCADE / SET NULL / SET DEFAULT
    forbidden by name; obligation extended to "delete or mutate".
F3. CLOSED — corrected §6.2 — non-serving maintenance principal named as the
    sanctioned home with minimum-bypass + unreachable-from-request-path +
    BYPASSRLS≠Authorization constraints; mechanism deferred to R10-F/ops.
F4. CLOSED — corrected §7 — exactly two lawful discovery shapes; the third
    path (failing-open semantic-table isolation for a scheduler) declared
    illegal target architecture.
F5. CLOSED — corrected §9 — Artifact→B3/B5, Audit→B6, Interchange→B6,
    Notifications/Search/async mechanism persistence→R10-D; all R10-A
    supporting owners have explicit homes.
F6. CLOSED — corrected §2.1 + §10 — namespace stays `metaldocs`; target-vs-
    legacy manifest/choreography routed to R10-F; namespace declared
    non-provenance during transition.
```

---

# 3. The seven delta questions

## 3.1 Is F1 truly closed by the orthogonal model, without a hidden new semantic owner?

**YES.** The root cause I identified — one law scoping its strongest control by
a predicate ("product") that another law's taxonomy should own — is removed
structurally, not patched: authority, durability and mutability are now
independent dimensions (§3), and the isolation law quantifies over the semantic
class dimension explicitly (§4.1–4.5). Checks performed:

- **No hidden owner.** The five classes are properties of persisted state, not
  owners. `DURABLE MECHANISM`/`EPHEMERAL MECHANISM` map onto R10-A's commodity
  mechanism list (outbox/queue/leases/DLQ); `ATTRIBUTED SUPPORT` and
  `REBUILDABLE PROJECTION` reuse R10-A's own vocabulary for
  Notifications/Search verbatim. Every `SEMANTIC AUTHORITY` fact still resolves
  to exactly one R10-A owner. The 8+3 set is unchanged.
- **No unbound scope terms remain.** Every isolation/role/outcome clause I
  could find now quantifies over defined vocabulary: §4.1/§4.2 ("tenant-owned
  SEMANTIC AUTHORITY / ATTRIBUTED SUPPORT"), §6.1 ("tenant semantic/support
  tables"), §12 ("protected semantic/support state"). The prior ambiguity is
  not re-expressible under the corrected text: a durable-intent table can no
  longer be honestly read as fail-closed-RLS-mandatory, because §4.3 assigns
  mechanisms an explicit per-block posture instead of inheritance.
- **The example classification table (§3) is coherent.** AuditEvent as
  SEMANTIC AUTHORITY/APPEND-ONLY is consistent with Audit being a supporting
  *semantic* owner (R10-A), and pulls Audit tables under the §4.1 fail-closed
  stack — which is correct and is operationally survivable because §7 gives
  the integrity-scan class its two lawful shapes.
- **The §4.3 deferral is bounded, not open.** A later block choosing a
  mechanism posture is constrained twice: the claim surface may expose only
  enumerated routing/mechanism facts, and §5 requires R10-D to preserve both
  claimability and tenant-content isolation if payload grows. Falsifiable.

## 3.2 Does DURABLE MECHANISM keep async intent claimable without global tenant-content access?

**YES — and the shape is mechanism-proven.** The §5 flow (claim routing fact →
obtain `tenant_id` + opaque target → re-enter ordinary tenant-scoped
transaction → load content under normal isolation/AuthZ) is exactly the
claim-then-seed pattern the current system already runs (per-message tenant
seeding after a cross-tenant claim; ADR 0027 M3 amendment, ADR 0054), minus the
NULL-permissive hatch on semantic tables that made it dangerous. The one real
delta from current practice is load-bearing and correct: today's
`outbox_events.payload` carries business content into the mechanism table
(`wiki/database/tables/outbox_events.md`); the corrected law confines the
globally claimable surface to routing metadata and pushes content loading
behind the tenant-scoped re-entry. A worker that needs content reads it from
semantic tables under seeded context via the opaque reference. Operationally
complete: same-commit insert works (tenant ctx present at write), global claim
works (mechanism posture, routing-only), content access works (re-entry).
No global tenant-content visibility is required at any step, and §5's closing
constraint pins R10-D to that property.

## 3.3 Is F2 closed by the positive closed FK-action set?

**YES.** §2.4 states the allowed set positively (`RESTRICT`/`NO ACTION`, both
`ON DELETE` and `ON UPDATE`), enumerates the forbidden remainder
(`CASCADE`/`SET NULL`/`SET DEFAULT`), and states the invariant in
delete-**or-mutate** form; proof obligation §11.2 matches. The prohibition-list
fail-open defect is gone — an unenumerated future referential action is now
illegal by default rather than legal by omission. Within-owner cascade remains
correctly non-default and bounded to strictly subordinate children.

## 3.4 Do F3/F4 corrections preserve fail-closed serving + operable jobs + possible maintenance simultaneously?

**YES — all three properties hold at once, each with a distinct mechanism:**

- **Serving fail-closed:** §4.1/§4.2 (missing context = fail closed on
  semantic/support state) + §6.1 (serving roles NOSUPERUSER / NOBYPASSRLS /
  non-owner). No serving path can widen visibility.
- **Jobs operable:** §7's two shapes cover the whole due-work surface —
  per-tenant iteration for scan-style work (watchdogs, integrity scans,
  review/retention due), tenant-written routing intent for event-driven work.
  Both proven viable in the prior review's operational analysis; nothing in
  the corrected text narrows them.
- **Maintenance possible:** §6.2 sanctions a distinct non-serving maintenance
  principal with minimum bypass for true cross-tenant migration/backfill/
  restore, with the constraints that matter (never the serving role, never
  reachable from the request path, BYPASSRLS never implies product
  Authorization), and routes the concrete credential/workflow to R10-F/ops.

The three properties are attached to three disjoint surfaces (serving runtime,
background discovery, platform maintenance), so none is satisfied by weakening
another. The prior review's operational-viability condition is discharged.

## 3.5 Does F5's assignment cover all supporting owners without pulling R10-C/D/E/F forward?

**YES.** Census against the promoted R10-A ownership set:

| R10-A owner/support | Corrected home | Premature mechanism pull? |
|---|---|---|
| Artifact (supporting owner) | B3 relational core (first consumer) + B5 second-consumer/no-orphan closure | No — physical storage/relocation stays R10-C |
| Audit (supporting owner) | B6 relational state | No — durability/execution mechanics stay R10-D where applicable |
| Interchange (supporting owner) | B6 batch/plan/outcome state | No — connector/external mechanics stay R10-C/D |
| Notifications (attributed support) | R10-D persistence details | No — correctly pushed out of R10-B |
| Search (projection) | R10-D persistence details | No — same |
| async mechanism tables | R10-D | No — B1 fixed only the law they must obey |

Nothing from R10-C/D/E/F is decided in R10-B by this map; the flow is
outbound (Notifications/Search/async pushed out), not inbound. One sequencing
tension examined and accepted: B2–B5 must declare Audit-append obligations
before B6 closes the Audit table shape — legal because B1 already fixes the
append *law* (same-commit, §5/§8) and R10-A fixes the seam; B6 closes shape
and the cross-owner matrix together, where global coherence is checked anyway.

## 3.6 Is F6 correctly deferred to R10-F without weakening the canonical namespace?

**YES.** §2.1 and §10 keep `metaldocs` as the final target namespace — the
correct Global-Maximum choice — while making the transition hazard explicit:
namespace is never provenance during cutover, R10-F must carry an explicit
target-vs-legacy manifest/choreography, and namespace occupancy is declared
non-reopen-evidence. This is a deferral with a named successor and a named
obligation, exactly the legal shape.

## 3.7 Did any correction create new authority, a material contradiction, or reopen pressure?

**NO.** Checks performed:

- **New authority:** none. The taxonomy is vocabulary; the maintenance
  principal is an ops trust surface (mechanism); the claim surface is R10-A
  commodity machinery. No twelfth semantic owner, no second authority for an
  existing meaning.
- **Contradiction against R10-A:** none found. Ownership of every fact family
  is untouched; composition still owns no durable meaning; acyclic/seam rules
  unaffected. Notifications *ownership* stays with `support/notifications`
  (R10-A); only its table-design *work package* moves to R10-D.
- **Contradiction against the frozen ledger:** none found. Same-commit audit
  (ledger §5), outbox premise, PlatformOperator/SystemPrincipal
  no-implicit-access (ledger §6) are all preserved or strengthened; no frozen
  invariant is weakened by any correction.
- **Internal contradiction between corrections:** none found. §4 (isolation by
  class), §5 (intent law), §6 (roles), §7 (discovery) compose without overlap
  or gap on the surfaces I attacked in the prior review.
- **Proof obligations §11:** all sixteen are falsifiable at design level, and
  the four implementation proof slots (fail-closed negative RLS proof under a
  serving-class role, FK-action census, same-commit rollback proof, role
  posture proof) are concrete enough to fail. Obligation 5's "both dimensions
  declared" is mechanically checkable at each B-block closure.
- **Reopen sets:** R9.5 = EMPTY, R10-A = EMPTY. Confirmed.

---

# 4. New findings in the corrected material

No BLOCKER. No MAJOR. Two LOW clarity observations, neither of which blocks
promotion nor requires a further review round:

## L1 — LOW — §9 B4 lists "Rendition/Release relational state" without naming its owner

R10-A assigns Rendition/Release/effectivity semantics to Controlled
Information. The corrected B4 line groups their relational-state design with
Approval/Distribution — a sensible *work-package* sequencing (Release binds
approval outcomes to effectivity), but a reader could misread the map as
ownership. One parenthetical — "(Controlled Information-owned)" — in the B4
entry, or an explicit "B-blocks are design work packages, not ownership
reassignment" sentence above the map, removes the misreading. Promotion may
apply this as an editorial touch; no decision changes.

## L2 — LOW — §4.2's narrower-representation escape clause should eventually be exercised against a named example

ATTRIBUTED SUPPORT gets the full isolation stack "unless a later owning block
proves a narrower representation that preserves the same isolation claim". The
clause is correctly bounded (the isolation *claim* is preserved, only the
mechanism may narrow), but it is the one place in the corrected text where a
later block can vary the stack. When R10-D designs Notifications persistence,
its closure evidence should explicitly exercise or decline this clause so the
escape hatch never becomes an unexamined default. This is a note for the R10-D
proof record, not a B1 text change.

---

# 5. Convergence

```text
prior findings: 6/6 CLOSED
new findings:   0 BLOCKER / 0 MAJOR / 2 LOW (editorial / successor proof note)
altitude:       mechanical; no structural concern recurred
stop condition: met — target properties structurally solved, remaining items
                are below the material threshold
```

Per the review protocol's stop conditions, the loop should stop: findings
decreased from 2 MAJOR + 4 LOW to 0 MAJOR + 2 LOW, and the remaining items are
editorial or future-proof-record notes. A further adversarial round would
manufacture scope, not surface risk.

---

# 6. Promotion recommendation

R10-B1 corrected target is fit for operator adjudication and promotion into
`wiki/architecture/r10-technical-architecture.md` as promoted R10-B1 law.

Recommended at promotion time (non-blocking):

1. optionally apply L1's one-line ownership annotation to the §9 map;
2. carry the §11 implementation proof slots into the promoted text unchanged;
3. record L2 as an R10-D closure-evidence expectation.

R10-B2 remains blocked until promotion; implementation remains blocked.
Findings here are evidence; the operator adjudicates under the Method.
