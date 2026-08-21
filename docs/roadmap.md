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
T8-H WHOLE-T8 GLOBAL COHERENCE        OPEN / ACTIVE / ROUND-1 MATERIAL CORRECTED
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

The Whole-T8 Lead pass found three material seams. The operator approved all three bounded corrections:

```text
H1  mutable stage/integration snapshots duplicated outside docs/roadmap.md
    -> remove mutable snapshots; retain immutable ratification evidence

H2  effective operation-47 DocumentOfficialView split across two live wire authorities
    -> fold the already-ratified read-symmetry precision into wire-contract.md;
       retain frontend-read-symmetry.md as routed provenance only

H3  accepted application/maintenance leaf omitted from T8-B topology projection
    -> reflect the existing T5-J GC consumer inside the existing application class
```

H2 and H3 survived independent challenge. H1's first correction was real but incomplete: the independent Round-1 review found older durable Product/T8 authorities still carrying mutable progression snapshots.

No correction creates a new Product capability, semantic owner, Permission, persistence authority, dependency class, runtime component or application operation.

### Initial proof and falsification

The initial local proof established the intended H1-H3 properties and required CI passed, including a useful routing failure in CI #1093. Fable Round 1 then materially falsified **H1 completeness**, not the sole-roadmap law itself.

```text
initial H2 wire SSOT proof                         PASS
initial H3 maintenance-leaf proof                  PASS
application census = 78 / operation 79 absent      PASS
initial required candidate CI                      PASS
H1 complete durable-status sweep                    FALSIFIED BY FABLE F1
```

The failure was accepted as a T8-H/H1 scope correction only; no Product/T1→T8 semantic authority reopened.

## Fable Round 1 — PR #149

Independent review was isolated from exact candidate:

```text
reviewed candidate HEAD       96c99fa6edfed5b015c5f93dbec9a40d806b8c95
candidate required CI         #1098 SUCCESS
Evidence PR                   #149 / review/t8h-fable
final review HEAD             4990a8b1c0e3123a135989577d0f13bda2175932
review required CI            #1100 SUCCESS
review delta                  docs/work/current/ai-dialog.md only
verdict                       NOT CONVERGED
MATERIAL findings             1
```

Lead adjudication:

```text
F1 MATERIAL  ACCEPTED
  H1 sweep was incomplete.
  Remove live current-stage/next-action/implementation routing from the implicated
  Whole-Product/T8-A→T8-D durable authorities and preserve only immutable
  ratification/handoff facts; current program state belongs to roadmap.

F2 MINOR     ACCEPTED
  git diff --check remains non-blocking for whitespace, but its
  "leftover conflict marker" diagnostic remains a blocking correctness failure.

F3 MINOR     ACKNOWLEDGED / NO BROAD REGEX
  Recurrence risk is real, but a blanket durable-doc text grep would conflate
  historical ratification/provenance with mutable state and recreate low-signal CI.
  Documentation governance + explicit sole-roadmap routing remain the rule;
  bounded Round 2 must attack this adjudication.
```

PR #149 is **CLOSED / UNMERGED**. Its evidence remains tied to the exact reviewed candidate and is never rebased into authority.

## Round-1 correction package

F1 correction is deliberately prose/status-only:

```text
docs/product/alignment.md
  remove stale T1 ACTIVE / Current gate / staging-packet / NEXT state
  retain immutable A1-A10 + 4+1 + T1→T7 ratification facts
  route mutable program state to roadmap

docs/architecture/technical-baseline.md
  freeze T8-A closure + historical T8-B handoff
  remove live "next allowed stage" / blocked-program assertions

docs/architecture/backend.md
  freeze T8-B closure + historical T8-C handoff
  remove live "next open stage" / implementation assertion

docs/architecture/interfaces.md
  freeze T8-C closure + historical T8-D handoff
  express T8-D/T8-E sections as authority partition, not current progression

docs/architecture/persistence.md
  remove the direct stale T8-E ACTIVE / T8-F→T8-H NOT OPEN contradiction
  freeze T8-D closure + historical T8-E handoff
```

Existing imported header/provenance strings not acting as live routing remain governed by `docs/development/documentation.md` and are not recursively normalized by preference.

F2 correction is equally bounded:

```text
Markdown trailing whitespace        warning only
leftover merge-conflict marker      required failure
```

The correction uses Git's own `git diff --check` diagnostic rather than inventing a second conflict-marker parser.

## CI proportionality baseline

The earlier operator-approved proportionality correction remains intact:

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

Workflow correction commit `bb8164f1f8cc784268f5f1f7515614c4703a37af` passed required CI **#1096**. Valid review isolation then passed on #1099/#1100. F2 only restores the non-whitespace correctness class accidentally carried by `git diff --check`.

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
prove the exact Round-1-corrected T8-H candidate with required CI
→ if green, open a fresh bounded Round-2 Evidence branch from that exact HEAD
→ Round 2 attacks F1 closure, F2 conflict-marker behavior and the F3 no-broad-regex adjudication
→ verify no Product/semantic/census regression; application operations remain exactly 78
→ adjudicate any surviving material finding against the smallest owning authority
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
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, runtime and accepted upstream authorities as one system | OPEN / ACTIVE; Round-1 F1/F2 corrected; exact CI + bounded Round 2 pending |
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
