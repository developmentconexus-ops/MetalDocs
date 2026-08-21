---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE                       CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                      MERGED / OPERATOR-RATIFIED
REPOSITORY STANDARD V1 ALIGNMENT      MERGED
PRODUCT / OWNERSHIP                   OPERATOR-APPROVED
T1 → T8-G                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T8-H WHOLE-T8 GLOBAL COHERENCE        OPEN / ACTIVE / CORRECTED
T9 → T12                              NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-H — Whole-T8 Global Coherence Review is **OPEN / ACTIVE** as of 2026-08-21 after explicit operator authorization and fresh repository revalidation.

Opening base:

```text
main                               0b4ef6ef891b01f907804cff4bd3c0022aebad80
open PRs at opening                0
active candidate branch            arch/t8h-global-coherence
active Draft PR                    #148
Product implementation             BLOCKED
T9                                 NOT OPEN
```

The opening `main` commit is the T8-G closeout squash from PR #147. Its candidate HEAD `ed0ff1c2883d92b65f6502cfa90798abb1cf8ed3` passed required CI **#1079** and **#1080**; CI #1080 job `required` completed successfully.

## T8-H coherence corrections

The Whole-T8 adversarial pass found three material seams. The operator approved all three bounded corrections:

```text
H1  mutable stage/integration snapshots duplicated outside docs/roadmap.md
    -> remove mutable snapshots; retain immutable ratification evidence

H2  effective operation-47 DocumentOfficialView split across two live wire authorities
    -> fold the already-ratified read-symmetry precision into wire-contract.md;
       retain frontend-read-symmetry.md as routed provenance only

H3  accepted application/maintenance leaf omitted from T8-B topology projection
    -> reflect the existing T5-J GC consumer inside the existing application class
```

The corrections create no new Product capability, semantic owner, Permission, persistence authority, dependency class, runtime component or application operation.

Correction proof:

```text
P1  roadmap remains sole mutable stage/status/next-action authority       PASS
P2  affected T8-F/T8-G records contain no stale progression snapshot      PASS
P3  effective DocumentOfficialView schema/laws live in wire-contract.md   PASS
P4  frontend-read-symmetry.md is routed provenance, not schema SSOT       PASS
P5  application/maintenance is present with no new dependency class       PASS
P6  application census remains exactly 78; operation 79 absent            PASS
P7  required CI #1095                                                      PASS
```

CI #1093 materially helped the proof: after H2 removed the precision record from the executable-wire route, repository reachability correctly rejected it as an unrouted durable document. The correction now routes the provenance record through the decision register while keeping executable-wire routing singular.

## CI quality correction

A bounded proportionality audit found that the `required` gate mixed material repository protections with avoidable red noise. The operator approved the smallest correction, now applied in `.github/workflows/ci.yml`:

```text
KEEP BLOCKING
  architecture-only tracked-path envelope while implementation is blocked
  bootstrap surfaces + <=20 KiB budget
  sole-roadmap leakage guard
  durable-document routing/reachability
  docs/work exclusion from merge candidates/main
  required authority presence
  archive/provenance reachability

CORRECT
  valid review/* Evidence PR -> PASS when Draft + isolated + candidate-based
  review/* marked ready or violating isolation -> FAIL
  forward-obligation counts -> derive from document declarations, not workflow constants
  Markdown whitespace -> warning only, not architecture failure
  superseded runs for the same PR -> cancel automatically
```

The workflow correction itself passed required CI **#1096**. The `required` context name and repository protection binding remain unchanged.

Active branch-only review ledger:

```text
docs/work/current/t8h-global-coherence.md
```

The ledger is temporary work, not authority, and must not enter `main`.

## Integrated T8 baseline

T8-A→T8-G remain operator-ratified. T8-E/T8-F/T8-G remain integrated. Whole-T8 closure continues to preserve:

```text
accepted application operations      78
orphaned operations                  0
invented application operations      0
operation 79                         absent
Idempotency-Key creations            exact 10
ETag read / mutation domains         13 / 13
exact-byte resources                 exact 4
stable SPA route meanings            exact accepted T6 route set
frontend semantic owner added        none
frontend Authorization engine        absent
parallel global server store         absent
one modular-monolith application runtime
one PostgreSQL product-state database
River workers in-process
one active ManagedContentStore
private conditional renderer + MalwareInspector
verified ephemeral exact-byte spool
Redis / BFF / realtime / external Search / generic event bus absent
Product implementation               BLOCKED
```

The bounded T8-E-FR read-symmetry meaning remains unchanged; T8-H only consolidates its executable representation into the T8-E wire SSOT.

## Exact next action

```text
run required CI on the exact current T8-H candidate
→ create isolated review/t8h-fable from that exact candidate
→ review branch adds only docs/work/current/ai-dialog.md
→ fresh independent Fable challenge attacks H1-H3 corrections + Whole-T8 regression + CI proportionality
→ adjudicate any material reviewer finding against current authority
→ if converged, prepare T8-H ratification/closure candidate
→ do not begin T9 and do not implement Product code
```

Candidate/review branch cleanup is non-authoritative housekeeping and does not open or block T8-H.

Do not reopen completed T1→T8-G or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR meaning retained and executable representation consolidated in wire SSOT |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/scanner/provider boundaries, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | OPEN / ACTIVE; H1-H3 + CI proportionality corrected; independent challenge next |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | NOT OPEN; opens only after Whole-T8 coherence closes |
| T10 — Transition / Cutover | Real current→target transition and rollback barriers | NOT OPEN; opens after T9 baseline |
| T11 — Implementation Program & Execution Graph | Bounded work graph and proof obligations | NOT OPEN; opens after T1→T10 accepted |
| T12 — Adversarial Implementation-Readiness | Independent implementation-readiness attack | NOT OPEN; opens after T11 |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED
T9  CLOSED / OPERATOR-RATIFIED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or generic infrastructure fashion are not reopen triggers.
