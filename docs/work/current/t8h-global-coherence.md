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

This is temporary branch-only work. It must not enter `main`. Durable meaning lives in the owning authorities; mutable program status lives only in `docs/roadmap.md`.

## 1. Review objective

T8-H asks whether T8-A→T8-G form one executable technical realization without:

```text
duplicate or missing authority
incompatible assumptions
removed required seams
repeated correctness machinery
speculative capability
accidental complexity created by the architecture itself
```

The review follows DevelopmentConexus Engineering Method v1.0.0. Accepted stages reopen only on material evidence, never preference.

Bounded primary pack:

```text
T8-A  docs/architecture/technical-baseline.md
T8-B  docs/architecture/backend.md
T8-C  docs/architecture/interfaces.md
T8-D  docs/architecture/persistence.md
T8-E  docs/architecture/wire-contract.md
T8-F  docs/architecture/frontend.md
T8-G  docs/architecture/runtime.md
```

Additional files were opened only for concrete authority/routing seams.

## 2. Opening revalidation

```text
main                              0b4ef6ef891b01f907804cff4bd3c0022aebad80
expected SHA                       MATCH
open PRs                           0
T8-H branch                        arch/t8h-global-coherence
Draft PR                           #148
closeout PR #147                   MERGED
PR #147 head                       ed0ff1c2883d92b65f6502cfa90798abb1cf8ed3
CI #1079                           SUCCESS
CI #1080                           SUCCESS
CI #1080 job required              SUCCESS
```

Remote historical/review branches are housekeeping/provenance only.

## 3. Whole-T8 attack result

| Seam | Result |
|---|---|
| T8-A → T8-B clean-slate derivation | PASS |
| T8-B → T8-C package/consumer fit | H3 found; operator-approved bounded precision assembled |
| T8-C → T8-D transaction/Audit/AuthZ/idempotency/GC/jobs | PASS |
| T8-D → T8-E persistence/concurrency → wire | PASS |
| T8-E → T8-F fresh-navigation/read symmetry | H2 found; operator-approved authority consolidation assembled |
| T8-F → T8-G concrete runtime consumers | PASS |
| Whole-T8 mutable status authority | H1 found; operator-approved authority cleanup assembled |
| Authentication / CSRF | PASS |
| Authorization single ALLOW/default-DENY authority | PASS |
| same-commit Audit / transaction | PASS |
| River durable intent | PASS |
| idempotency / replay | PASS |
| exact content / exact-byte delivery | PASS |
| OfficialRendition | PASS |
| PostgreSQL Search baseline | PASS |
| application operation census | PASS — 78; operation 79 absent |

## 4. H1 — mutable status authority duplication

### Finding

Durable T8-F/T8-G authority and ratification records retained progression snapshots such as `INTEGRATION PENDING`, `T8-G NOT OPEN` and `T8-H NOT OPEN`, while repository law makes `docs/roadmap.md` the sole mutable stage/status authority.

### Operator adjudication

```text
APPROVED
```

### Assembled correction

```text
preserve immutable ratification/review facts
remove mutable downstream-stage/integration snapshots
remove INTEGRATION PENDING from current decision dispositions
point current integration/stage/implementation/next-action state to docs/roadmap.md
```

Affected durable surfaces:

```text
docs/architecture/frontend.md
docs/architecture/runtime.md
docs/decisions/t8f-ratification.md
docs/decisions/t8g-ratification.md
docs/decisions/index.md
```

No Product or T8 semantic meaning changed.

## 5. H2 — executable wire authority split

### Finding

The effective operation-47 `DocumentOfficialView` required `frontend-read-symmetry.md` to supersede the schema declared by `wire-contract.md`, leaving two live executable wire sources.

### Operator adjudication

```text
APPROVED
```

### Assembled correction

`docs/architecture/wire-contract.md` now contains:

```text
OpenRevisionRoutingReference
DocumentOfficialView.open_revision?
DocumentOfficialView.active_obsolescence_request_id?
exact existence + disclosure presence laws
absence-is-not-nonexistence law
follow-up Authorization/lifecycle recheck law
```

`docs/decisions/frontend-read-symmetry.md` now preserves decision rationale/provenance only and explicitly does not supersede executable schema. `docs/index.md` routes executable-wire work to `wire-contract.md` alone.

Wire effect remains:

```text
operations added       0
operations removed     0
application census     78
operation 79           absent
Permissions added      0
Problems added         0
persistence pointers   0
```

## 6. H3 — missing accepted maintenance leaf

### Finding

T8-C ratified T5-J managed-content GC choreography at `internal/application/maintenance`, but T8-B's frozen package projection omitted that already-accepted leaf.

### Operator adjudication

```text
APPROVED
```

### Assembled correction

T8-B now projects:

```text
internal/application/maintenance
  current consumer = T5-J managed-content GC choreography
  class = existing non-semantic application class
```

It adds no semantic owner, public owner surface, dependency class, runtime shell or Product operation.

## 7. Challenged seams with no material finding

### Imported `wiki/...` strings

Documentation governance classifies the remaining imported `wiki/...` authority strings as non-navigational provenance. No cleanup is justified.

### River / `database/sql`

Current River upstream support includes `riverdatabasesql`; its transaction form uses `*sql.Tx` and transactional insertion can share the caller transaction. This fits the accepted T8-C/D transaction seam and T8-G in-process worker topology.

### Security / content / state

No second Authorization evaluator, Product lifecycle truth, exact-content authority, durable frontend entity store, Redis dependency, external Search authority, generic outbox/event bus or provider-key Product identity survived the attack.

## 8. Correction proof contract

The corrected candidate must prove:

```text
P1  roadmap is the sole mutable stage/status/next-action authority
P2  affected T8-F/T8-G durable records contain no stale integration/downstream-stage state
P3  effective DocumentOfficialView schema + disclosure laws live in wire-contract.md
P4  frontend-read-symmetry.md is provenance, not second executable schema authority
P5  application/maintenance is present in T8-B with no new dependency class edge
P6  application census remains exactly 78 and operation 79 absent
P7  required candidate CI passes
```

Independent Fable challenge follows only after the exact corrected candidate satisfies this proof contract.

## 9. CI quality audit — separate from T8 architecture findings

The operator requested a proportionality review because required CI has produced repeated red runs.

Current evidence:

```text
CI #1081
  failed only because git diff --check rejected three Markdown trailing-space hard breaks
  corrected mechanically
  same gate CI #1082 then passed

review PR #145
  structurally valid Fable Evidence PR
  CI #1064 failure by workflow design

review PR #146
  structurally valid converged Fable Evidence PR
  CI #1068 failure by workflow design
```

The current review-branch path validates isolation and then deliberately executes `exit 1`. Repository Standard v1 requires review isolation to be proven; it does not require a correct Evidence PR to be red.

Provisional classification:

```text
KEEP / BLOCK
  architecture-only tracked-path envelope while implementation is blocked
  required bootstrap surfaces + bootstrap budget
  docs/work absence from merge candidate/main
  router/reachability proof
  required provenance/archive refs

REDESIGN
  valid review branch should PASS isolation check, not fail by design
  forward-obligation proof should compare document rows to its own summary,
    not duplicate 21/3/27/51 as workflow authority

DOWNGRADE FROM REQUIRED BLOCKER
  generic Markdown trailing-whitespace failure; it is formatting hygiene,
  and in Markdown two trailing spaces can be intentional syntax

ADD SMALL OPERATIONAL IMPROVEMENT
  cancel superseded CI runs for the same PR
```

No CI workflow correction has been applied yet. It is a separate operator decision and must not be smuggled into H1–H3.

## 10. Current T8-H decision

```text
T8-H outcome            NOT YET RATIFIED
H1                       APPROVED / CORRECTION ASSEMBLED
H2                       APPROVED / CORRECTION ASSEMBLED
H3                       APPROVED / CORRECTION ASSEMBLED
Product reopen           NO
T1→T7 reopen             NO
T8-A reopen              NO
application operations   78
operation 79             ABSENT
T9                       NOT OPEN
implementation           BLOCKED
```

Next: prove P1–P7 on the exact candidate, resolve the bounded CI-quality decision, then perform fresh independent challenge.
