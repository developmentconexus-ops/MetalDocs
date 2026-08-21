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
T1 → T8-F                             CLOSED / OPERATOR-RATIFIED / INTEGRATED
T8-G RUNTIME / PROCESS / DEPLOYMENT   NEXT / NOT STARTED
T8-H → T12                            NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T8-F — Frontend Realization is **CLOSED / OPERATOR-RATIFIED / INTEGRATED** as of 2026-08-21.

PR #139 was squash-merged into `main` as:

```text
711b8526ebacea9e15034459f85fc37707a32ab4
```

The merge commit carries tree:

```text
ea128d0900667846f33ef188552624886d7885b0
```

That tree is exactly the tree of the final authorized T8-F HEAD `ac3a4b23655e9529e73a6208d41f1804b27855f4`. The same HEAD passed required CI #1055 while Draft and required CI #1056 after the PR was marked ready. The squash commit therefore integrates the exact ratified and verified tree; no Product/runtime delta was introduced by merge.

Ratified durable authority:

```text
docs/architecture/frontend.md
+ docs/decisions/frontend-read-symmetry.md   // bounded T8-E-FR precision discovered by T8-F
+ docs/decisions/t8f-ratification.md         // explicit operator ratification record
```

The integrated frontend result remains:

```text
accepted application operations      78 / 78 covered
orphaned operations                  0
invented application operations      0
operation 79                         absent
stable SPA route meanings            exact accepted T6 route set
frontend semantic owner added        none
server-state authority               TanStack Query / server truth
parallel global server store         absent
manual parallel DTO/API authority    absent
frontend Authorization engine        absent
interactive DOCX runtime             one adapter boundary
```

### Independent challenge and adjudication

Fable Round 1 Evidence PR #140 challenged candidate `a32ba8b58f5574336f825f46bd552dd96246de7f` and returned F1–F6. Lead adjudication accepted all six and the operator approved the bounded correction package.

The only upstream precision was T8-E-FR:

```text
DocumentOfficialView
  + open_revision?: { revision:RevisionIdentity, state:OpenRevisionState }
  + active_obsolescence_request_id?: Uuid
```

These are disclosure-safe derived routing references, not persisted pointers. They add no Product capability, Permission, lifecycle state, T8-C contract class or application operation. T8-D already guarantees uniqueness of the current DRAFT/SUBMITTED Revision and ACTIVE ObsolescenceRequest per Document.

T8-F additionally closed:

```text
navigation presence is not permission-filtering authority
permission.csrf_failed -> session/CSRF re-bootstrap before safe same-command retry
state.upload_expired -> preserve local bytes; fresh allocation/upload/complete; no capability revival
Audit -> inspection/paging only; no inferred filter
History/My Work -> never current-resource identity resolvers
```

Bounded Fable Round 2 Evidence PR #141 reviewed exact corrected candidate `e54986904063c982315129635191ebade8f9b9ed` and returned:

```text
CONVERGED
MATERIAL findings = 0
Round 3 = NOT JUSTIFIED
```

The operator explicitly ratified T8-F on 2026-08-21. Both Evidence PRs are closed and unmerged.

## Ratified T8-E baseline

T8-E remains **CLOSED / OPERATOR-RATIFIED / INTEGRATED** with the bounded T8-E-FR member precision above.

Application census remains exactly 78 operations. Operation 79 remains a material Product/T6 reopen.

T8-E proof remains:

```text
ledger rows                         78 / exact 1..78
operationId                         78 unique
method + path                       78 unique
Idempotency-Key creations           exact 10
ETag read / mutation domains        13 / 13
exact-byte resources                exact 4
Audit operation codes               37 unique
Problem namespace                   https://errors.metaldocs.io/{code}
ShortText / LongText                256 / 4096
attention_required                  absent
PROFILE_REPLACE If-Match+absent     412 precondition.resource_changed
rows 3 / 45 validation.failed       absent
row 42 validation.failed            present
forward obligations                 21 / 3 / 27 = 51
```

T8-E-FR changes none of those counts.

## Exact next action

```text
explicit operator authorization to open T8-G
→ start fresh from current `main` and revalidate its exact SHA at execution time
→ read AGENTS.md → docs/index.md → docs/roadmap.md → only the bounded T8-G authority pack routed from there
→ derive the smallest Runtime / Process / Deployment contract from the accepted T8-A→T8-F consumers
→ do not open T8-H and do not implement Product code
```

Candidate/review branch cleanup is non-authoritative housekeeping and does not open or block T8-G; remove absorbed branches when tooling permits and provenance is no longer needed.

Do not reopen completed T1→T8-F or the 78-operation Product/T6 census by preference. New material evidence reopens only the authority it actually implicates.

## Remaining architecture program

| Stage | Owns | Opens / exits |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED; T8-E-FR precision ratified with T8-F |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, query/state behavior, read-model consumption, editor/viewer boundaries | CLOSED / OPERATOR-RATIFIED / INTEGRATED |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River workers, renderer/provider boundary, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, recovery/runtime profiles | NEXT / NOT STARTED; opens only on explicit operator authorization from current `main` |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | NOT OPEN; opens after T8-G closes |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes | NOT OPEN; opens after Whole T8 coherence |
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

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, or hypothetical future capability are not reopen triggers.
