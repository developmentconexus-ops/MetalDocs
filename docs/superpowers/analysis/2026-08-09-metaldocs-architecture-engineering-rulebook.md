# MetalDocs Architecture Engineering Rulebook and Remediation Map

**Date:** 2026-08-09  
**Status:** architecture-audit working rulebook; no runtime implementation is authorized by this document.  
**Branch:** `docs/architecture-audit-current-state`  
**Umbrella:** #100  
**Draft checkpoint:** #101

## 1. Purpose

This document turns the current architecture audit into an explicit engineering model for MetalDocs. It answers four questions:

1. What properties are already healthy and must not be damaged?
2. What structural properties are currently violated or ambiguous?
3. What rules should govern new/refactored code?
4. Which existing remediation program owns each class of defect?

It is deliberately not a second roadmap. The implementation programs remain issues #87-#95 and their ratified ADRs/specs.

## 2. Authority hierarchy

Use the repository's existing truth hierarchy:

1. runtime truth — code/schema/process behavior;
2. contract truth — OpenAPI/generated contract surfaces;
3. wiki/ADR truth — intended ownership, architectural decisions, debt;
4. execution truth — verifier, scripts, CI gates, evidence.

When these disagree, record the mismatch. Do not make code conform to stale prose merely because the prose is older.

## 3. Actual current module inventory

`internal/modules/` currently contains 15 module directories:

- `approval`
- `audit`
- `auth`
- `controlleddocuments`
- `distribution`
- `documents`
- `iam`
- `jobs`
- `notifications`
- `render`
- `search`
- `security`
- `taxonomy`
- `templates`
- `tokens`

Older topology/blueprint documents that say 11/12 modules are stale on this fact. This is wiki-memory drift, not evidence that the newer directories are or are not correct bounded contexts.

## 4. Architecture style that remains valid

MetalDocs remains a **modular monolith**. Do not convert these findings into a microservice program.

Preserve:

- explicit composition roots;
- package-level acyclicity (Go compiler + measured graph: zero multi-node package SCCs);
- zero same-module `domain -> infrastructure/delivery` and `application -> delivery` import inversions found in the measured package graph;
- contract-first HTTP/OpenAPI direction;
- transactional outbox rather than network calls inside write transactions;
- explicit SQL and DB constraints rather than an ORM rewrite;
- parameter binding and tenant predicate discipline;
- hand wiring in `main.go` rather than introducing reflection DI solely to reduce wiring lines.

## 5. Fundamental distinction: package cycle vs module cycle

Go package cycles are impossible and the measured graph contains none.

That does **not** mean the architecture is acyclic at the bounded-context/module level.

Example shape:

```text
module A/application -> module B/domain
module B/application -> module A/domain
```

Both package edges compile, but after collapsing packages to module identity the graph contains:

```text
A <-> B
```

The inventory reproduced seven such module-level reciprocal relationships. Therefore:

- package acyclicity is a compiler property;
- module acyclicity is an architecture property;
- `check-module-boundaries.ps1` currently proves only a subset of import visibility rules, not module acyclicity.

Future architecture verification must compute the module graph from the Go package graph and test SCCs/reciprocal edges mechanically.

## 6. Module ownership rules

### R-MOD-1 — A module is a business ownership boundary, not a folder category

A directory under `internal/modules/` is a current implementation fact. It earns bounded-context status only if it owns a coherent business language, invariants, state, and lifecycle.

ADR 0093 already rules that current `documents`, `controlleddocuments`, and `templates` topology is not proof of three peer contexts.

### R-MOD-2 — One authoritative owner per business state

Every mutable business fact/table belongs to one context. Other contexts must not read or mutate its schema directly merely because all tables share Postgres.

Current counter-evidence:

- Approval directly reads Documents tables;
- Documents directly reads Approval tables;
- Approval directly reads Controlled Documents data.

This is an ownership violation even when there is no Go import violation.

### R-MOD-3 — Cross-module collaboration is capability-shaped

A consumer should depend on the **smallest capability it needs**, not on the producer's internal model.

Preferred shape:

```go
// documents/application — owned by the consumer

type DictionaryValueReader interface {
    Lookup(ctx context.Context, tenantID, name string) (value string, found bool, err error)
}
```

Composition root:

```text
tokens producer -> adapter -> documents-owned port
```

This pattern already exists in `dictionary_reader_adapter.go` and is the local exemplar.

### R-MOD-4 — Ports belong to the consumer by default

Do not solve coupling by moving a producer-defined interface into the producer's `domain` package and making every consumer import it.

Bad default:

```text
security -> iam/domain.UserReader
```

Preferred:

```text
security/application.UserDirectory   (consumer-owned)
               ^
               |
composition adapter
               |
             iam
```

Exceptions require a genuinely shared stable vocabulary and an explicit architectural reason, not convenience.

### R-MOD-5 — Do not pass foreign aggregate/domain types across seams when stable values suffice

A module should not require another context's entity, aggregate, or large application DTO simply to perform one collaboration.

Prefer explicit seam values such as:

```text
SubjectRef
UserRef
AreaRef
RevisionRef
ApprovalReceipt
ReleasedArtifactRef
```

with only the fields the consuming use case requires.

Current examples to remove under A4 include Approval depending on Documents application/domain types and Documents depending on ControlledDocuments/Templates domain types.

### R-MOD-6 — Raw foreign sentinels are not integration contracts

`errors.Is(err, othermoduledomain.ErrX)` across module boundaries makes the producer's raw error identity part of the consumer contract.

At module seams, translate producer failures in an adapter/application boundary into a stable consumer-owned result/error vocabulary.

HTTP translation is a separate concern: domain/application errors map once into the canonical RFC 9457 problem surface under A3.

## 7. Data ownership rules

### R-DATA-1 — No foreign-table SQL across bounded contexts

A repository/application component may query tables owned by its own context. It must not depend directly on another context's table/column shape.

Reason: a Go import checker cannot see SQL strings, and a column/schema change otherwise creates hidden blast radius.

### R-DATA-2 — Reads do not get a blanket exception

"It is only a SELECT" is still coupling to the foreign schema. Cross-context reads use:

- a consumer-owned read port;
- a producer application/query service;
- a deliberate read model/projection whose ownership is explicit;
- an integration event/materialized projection where latency/scale justifies it.

### R-DATA-3 — Cross-context atomicity requires an explicit ownership decision

Do not expose another module's repository or raw transaction merely to preserve a multi-module transaction.

When one business invariant truly requires atomic updates across currently separate contexts, treat that as evidence that either:

- the state belongs to one context;
- an application-level orchestrator owns the unit of work with explicit ports;
- or the interaction should become asynchronous/outbox-based.

Do not hide an unresolved boundary under shared SQL.

## 8. Layering rules inside a module

Target conceptual direction:

```text
delivery/http -> application -> domain
                    |
                    v
             consumer-owned ports
                    ^
                    |
              infrastructure
```

### R-LAYER-1 — Domain contains business language and rules only

`domain` must not execute SQL, know HTTP, Redis, MinIO, OpenAPI, logging frameworks, or another context's infrastructure.

The audit already found no literal SQL executed in domain packages; preserve this.

### R-LAYER-2 — Domain should not expose persistence-driver types

The inventory found 9/15 domain packages importing `database/sql` and/or `internal/platform/db` in port signatures.

Target rule:

- business domain values remain persistence-neutral;
- transaction/executor mechanics belong to application/infrastructure seams, not business entity/value-object APIs;
- existing `db.Tx` use in **application-owned ports** can be a pragmatic transitional mechanism when same-transaction behavior is required;
- new domain types/ports must not introduce `*sql.Tx`, `sql.Null*`, `sql.Row`, `sql.Result`, `db.Tx`, or repository-specific types without an explicit ADR/ruling.

This respects the existing `TxRunner` bounded concession while stopping persistence vocabulary from spreading further into domain code.

### R-LAYER-3 — Application orchestrates use cases; infrastructure supplies adapters

Application may coordinate domain logic, authorization, transactions, clocks, outbox publication, and consumer-owned ports.

Infrastructure owns Postgres/Redis/MinIO/HTTP-client implementations.

### R-LAYER-4 — Delivery does decode/validate/call/map

Handlers should:

1. decode generated/contract input;
2. apply request-boundary validation;
3. call one application use case/service;
4. map the result into the contract.

Business rules, SQL, authz policy derivation, or per-handler error dialects do not belong in delivery.

## 9. Platform vs composition rules

### R-PLAT-1 — `internal/platform` is domain-free

This is already REQ-TOP-2.

A platform package must be reusable technical machinery: DB, telemetry, HTTP primitives, security middleware, object storage, configuration, idempotency, etc.

It must not import `internal/modules/**`.

The audit found module-specific imports from platform packages such as `bootstrap`, `authn`, `docgenv2`, `tripwire`, and `worker`. These are owned by #93/A4.

### R-PLAT-2 — Composition is allowed to know implementations

Composition roots are the correct place to import multiple modules and bind adapters.

`internal/composition/**` or executable roots may import module infrastructure/application surfaces for wiring. That is not a domain dependency.

Keep composition separate from platform so the architecture verifier can distinguish `W` (wiring) from real module coupling.

## 10. Transaction/persistence spine

The current platform contains `TxRunner`, explicitly intended to own begin/commit/rollback, but the persistence audit found 82 direct `BeginTx` sites across 25 files and parallel transaction abstractions.

### R-TX-1 — One transaction lifecycle mechanism

New application code must not introduce a fresh manual begin/defer-rollback/commit pattern when the shared transaction runner can express the use case.

A5 owns migration of existing sites.

### R-TX-2 — Do not make domain ports own transaction lifecycle

A domain interface should not define `BeginTx`, `Commit`, or `Rollback` behavior. Unit-of-work lifecycle is application/infrastructure mechanics.

### R-TX-3 — Keep SQL explicit but make correctness mechanical

The target is not an ORM. The A5 `sqlc` spike is appropriate because it can make query parameters/results and column scans compile-checked while preserving visible SQL.

Adopt incrementally by repository/hotspot rather than flag-day conversion.

## 11. API/error contract rules

### R-API-1 — OpenAPI remains route/shape source of truth

Do not hand-create a parallel request/response type when a generated contract type exists.

### R-API-2 — Runtime validation must enforce contract constraints

The audit found OpenAPI constraints that exist only as documentation. A3 owns ingress validation so the runtime and contract cannot disagree silently.

### R-ERR-1 — One external error envelope

All HTTP failures use the canonical RFC 9457 problem vocabulary/writer.

No local `writeProblem` variants, bare `http.Error`, or ad-hoc JSON error envelope for normal API routes.

### R-ERR-2 — Internal module errors are not automatically public API errors

Translate:

```text
producer internal failure
  -> consumer/application error/result
  -> HTTP problem code at delivery boundary
```

Do not make HTTP status codes part of domain errors.

## 12. Cross-cutting concerns

Cross-cutting technical capabilities belong in platform/shared infrastructure only when they remain domain-free:

- logging;
- tracing;
- metrics;
- request correlation;
- idempotency machinery;
- rate limiting;
- object-store clients;
- generic DB execution;
- generic HTTP clients;
- config parsing/validation.

Audit/evidence that is itself a business/regulatory concept remains a domain module even if many modules use it.

## 13. Architecture tests / mechanical enforcement

Architecture is not closed because a wiki says the right thing.

Future verifier properties:

### V-ARCH-1 — package/module graph

Build the actual Go import graph; collapse `internal/modules/<module>/**` to module identity; fail on prohibited SCCs/reciprocal edges.

### V-ARCH-2 — forbidden implementation imports

Retain/replace the useful property from `check-module-boundaries.ps1`: cross-module code cannot import another module's infrastructure/repository/delivery implementation.

### V-ARCH-3 — platform domain freedom

Fail if `internal/platform/**` imports `internal/modules/**`.

Composition paths are classified separately, not allowlisted as platform exceptions.

### V-ARCH-4 — SQL ownership

Use one machine-readable ownership catalog. Scan/parse SQL identifiers and fail when a module references a foreign-owned business table outside an explicitly designed projection/composition mechanism.

Do not maintain a manual "known foreign SQL files" kill list.

### V-ARCH-5 — foreign sentinel/type contract

Detect cross-module dependencies on raw domain sentinels and disallowed producer concrete types. Ratchet transitional debt to zero rather than accepting new sites.

### V-ARCH-6 — domain purity

Block new `domain` imports of HTTP, module infrastructure, DB driver types, object store, generated API packages, or other forbidden technical surfaces.

### V-ARCH-7 — every guard has a negative fixture

Per #87, an architecture guard that has never proven it fails on bad input is not a trusted control.

The same verifier manifest must run locally, for agents, and in CI.

## 14. Current-state hotspots and owners

| Hotspot | Evidence class | Owner |
|---|---|---|
| 7 module-level cycles | Go/module graph | #93 / A4 |
| 17+ cross-context table reads | SQL ownership | #93 / A4; A9 absorbs Controlled Information internal seams after ADR 0093 |
| 62 foreign sentinel checks | error contract | #93 / A4 + #90 / A3 at HTTP translation |
| producer-owned types/interfaces at seams | contract ownership | #93 / A4 |
| 9/15 domain packages leak SQL/platform DB types | layering/transaction vocabulary | #93 / A4 + #92 / A5 |
| platform imports modules | layering | #93 / A4 |
| 82 direct BeginTx sites | transaction mechanics | #92 / A5 |
| 242 hand-maintained scan sites | persistence correctness | #92 / A5 |
| request/error/validation dialects below OpenAPI | boundary contract | #90 / A3 |
| dual-source authorization | domain/security semantics | #89 / A8 |
| documents/templates/controlleddocuments peer-context split | domain decomposition | #94 / A9, ruled by ADR 0093 |
| async worker/jobs blind spots | operability | #95 / A7 |
| inert/non-reproducible verification | governance/tooling | #87 / A1 |
| quality rules not ratcheted | maintainability | #91 / A2 |
| fail-closed deployment/security assumptions | security | #88 / A6 |

## 15. Execution dependency rule

Do not execute #87-#95 as nine unrelated parallel refactors.

Safe program shape from the current issue/ADR constraints:

```text
A1 verifier spine
  |-- A3 API/runtime contract
  |-- A8 authz unification
  |-- A6 fail-closed security
  `-- A2 quality ratchet

A3 + A8 -> A4 module seam migration
A4 -> A5 persistence migration where seam work touches repositories
A8 + A4 (+ persistence spine where needed) -> A9 Controlled Information consolidation
A3 -> A7 async trace/operability contract where propagation conventions are shared
```

This is dependency guidance, not authorization to merge every axis into one mega-PR.

## 16. Claude Code issue execution contract

When Claude Code takes an issue (#87-#95), it should:

1. read `AGENTS.md` and the issue's owning ADR/spec;
2. identify the exact root-cause property, not only the first symptom;
3. list affected architecture rules from this document;
4. build a measured current-state baseline for that issue;
5. implement the smallest coherent slice that makes one target property mechanically stronger;
6. add/extend a verifier or negative fixture where the issue's acceptance requires mechanical enforcement;
7. run targeted and broad regression appropriate to the crossed boundaries;
8. update wiki/current-state evidence only after code truth changes;
9. never claim the issue closed because counts decreased if the defect remains representable;
10. stop if implementation uncovers a target contradiction with an ADR or with another active root-cause program.

## 17. Definition of architecture improvement

A refactor counts as architectural improvement when at least one of these becomes true:

- an invalid dependency is no longer representable;
- ownership becomes singular and machine-checkable;
- a consumer can change independently of a producer's internals;
- a business invariant gets one authoritative implementation/owner;
- a contract is generated or verified instead of hand-synchronized;
- a security/operability property becomes fail-closed and observable;
- a guard is proven to fire before handoff.

Moving files, renaming packages, adding interfaces on the producer side, or reducing a grep count without changing these properties is not sufficient.

## 18. Non-goals

- microservices;
- DI framework adoption;
- ORM adoption;
- arbitrary package splitting based on LOC alone;
- replacing explicit SQL solely for aesthetic reasons;
- rewriting all errors at once without the A3/A4 contract model;
- treating every duplicated name as a defect across bounded contexts;
- using current topology as evidence that current topology is correct.
