# MetalDocs Architecture Audit — Reproduced Inventory Evidence

**Date:** 2026-08-09  
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273`  
**Status:** checked-in mechanical evidence + direct current-source validation; not re-executed in this ChatGPT terminal.

## 1. Provenance

The repository already contains a prior architecture discovery artifact at `docs/superpowers/analysis/inventory/layering.md`. It records commands and computed results from a local repository pass, including `go list -json ./internal/...`, Tarjan SCC analysis, fan-in/fan-out calculation, symmetric module-edge analysis, grep-based SQL checks and direct source inspection.

The current assistant terminal cannot resolve `github.com`, so it cannot clone the repository and pretend to freshly rerun `go list`. Direct GitHub source/index inspection is available and has been used to validate representative seams and current topology at the pinned baseline.

## 2. Graph facts already mechanically measured

The existing inventory reports:

- **136 first-party packages** in `./internal/...`;
- **0 multi-node package-level SCCs** under Tarjan analysis;
- **7 module-level reciprocal relationships** after collapsing package edges to module identity;
- **0 intra-module layer inversions** where a module's own `domain` imports its own `infrastructure`/`delivery`, or its own `application` imports its own `delivery`;
- `iam/domain` fan-in = **32** importers;
- `iam/authz` fan-in = **26** importers;
- `platform/db` fan-in = **32** importers;
- `platform/httprouter` fan-in = **21** importers.

Critical distinction:

```text
package graph acyclic  !=  module graph acyclic
```

The Go package structure avoids compiler-level cycles while reciprocal semantic dependencies still exist after collapsing to bounded-context/module identity.

## 3. Actual current module topology

At `main@418070bf...`, `internal/modules/` contains **15** directories:

`approval`, `audit`, `auth`, `controlleddocuments`, `distribution`, `documents`, `iam`, `jobs`, `notifications`, `render`, `search`, `security`, `taxonomy`, `templates`, `tokens`.

Older wiki artifacts that report 11/12 modules, describe nested `documents/approval`, or say top-level Approval is absent are stale/historical on those facts.

## 4. Persistence leakage through domain contracts

The inventory records two related classes:

- direct `database/sql` types in domain port signatures (`*sql.Tx`, `sql.NullTime`);
- domain packages importing `internal/platform/db` transaction plumbing.

Scale: **9 of 15** module domain packages import `database/sql` and/or `platform/db`.

Current policy interpretation after inspecting `internal/platform/db/runner.go` and `tx.go`:

- `TxRunner` deliberately exposes the live transaction in an application callback as a bounded concession while owning begin/commit/rollback;
- this does not grant business `domain` packages permission to accumulate persistence-driver vocabulary;
- application/infrastructure transaction seams can use the shared abstraction transitionally where atomicity requires it;
- new domain persistence-type leakage is frozen pending A4/A5 migration.

## 5. Producer-owned contracts at module seams

The inventory confirms both directions of the same design problem.

Examples:

- Documents application constructs/uses ControlledDocuments domain types/functions;
- Documents exposes Templates domain values in application signatures;
- Approval application accepts Documents application/domain types;
- Security infrastructure is typed by reader interfaces declared in IAM domain;
- IAM infrastructure consumes a reader interface declared in Taxonomy domain.

### Positive counter-example

`documents/application` declares its own `DictionaryValueReader`. `apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go` adapts Tokens to that consumer-owned port and translates `tokensdomain.ErrNotFound` at the composition root.

Desired shape:

```text
consumer owns required capability contract
        ↑
composition adapter
        ↑
producer implementation
```

Root-cause owner: **#93 / A4**.

## 6. Foreign error identity as an undeclared contract

Measured inventory:

- **62** `errors.Is(err, <foreign>domain.Err...)` call sites;
- across approval, auth, controlleddocuments, documents, iam and templates.

The important property is not the count: producer sentinel identity is functioning as a cross-context API without an explicit seam contract.

Owners:

- seam design: **#93 / A4**;
- HTTP/runtime error translation: **#90 / A3**.

## 7. Cross-module SQL / data ownership

The inventory reproduces at least **17+ foreign-table reads**:

- Approval -> Documents: multiple `documents` / `document_comments` reads in `postgres_approval_repository.go`;
- Documents -> Approval: reads of `approval_instances`, `approval_signoffs`, `release_generations` across Documents infrastructure/application files;
- Approval -> ControlledDocuments: direct `controlled_documents` join.

This proves that Go import rules alone cannot establish data ownership.

Owner: **#93 / A4**, coordinated with #92/A5; ADR 0093/A9 will absorb seams that legitimately become intra-context after Controlled Information consolidation.

## 8. Platform -> module direction

The inventory distinguishes legitimate composition from misplaced module-specific platform code.

- `internal/composition/tenantdata/registry` imports 12/15 module infrastructure packages to assemble the registry: composition-shaped, not automatically a domain defect.
- platform packages `bootstrap`, `authn`, `docgenv2`, `tripwire`, `worker` carry **11** module-specific edges outside that registry.

Target REQ-TOP-2 says platform is domain-free. A4 owns the migration of module-specific concerns out of generic platform locations.

## 9. Architecture properties that are currently healthy

Preserve:

1. zero package-level import cycles;
2. zero measured same-module layer inversions of the `domain -> infrastructure/delivery` / `application -> delivery` classes;
3. no literal SQL execution in domain packages;
4. explicit composition-root wiring with no reflection DI framework;
5. the consumer-owned DictionaryValueReader adapter pattern;
6. explicit SQL, parameter binding and DB constraints rather than an ORM rewrite.

## 10. Recent-current-state correction for #87/#91

The original issue evidence predates merged #97/#99.

At baseline `418070bf...`:

- `.github/workflows/` contains 5 files, not the original 20-workflow topology;
- `tools/verify/` exists as a substantial Go verifier/registry with tests;
- PR #99 removed `only-new-issues` and reports whole-tree golangci burn-down to zero for its configured scope.

These are completed improvements and must not be reimplemented. #87 and #91 remain governed by their **acceptance properties**, with their original counts treated as filing-time evidence.

## 11. What remains to make the architecture graph itself a durable product

1. promote module graph derivation into the trusted verifier (#87), not a one-off audit script;
2. emit machine-readable module adjacency/SCC evidence from current Go packages;
3. derive SQL ownership violations against one authoritative ownership catalog;
4. classify/rachet foreign sentinel/type coupling mechanically;
5. enforce platform domain-freedom while distinguishing composition roots;
6. keep new domain persistence leakage red while A4/A5 burn down existing debt;
7. keep wiki module topology/maturity synchronized from code truth after each owning program lands.

## 12. Remediation ownership

| Finding family | Owner |
|---|---|
| module cycles / seam direction / producer-owned contracts | #93 / A4 |
| foreign SQL ownership | #93 / A4 + #92 / A5 |
| foreign error identity | #93 / A4 + #90 / A3 |
| domain persistence-type leakage | #93 / A4 + #92 / A5 |
| platform -> module direction | #93 / A4 |
| verifier reachability / negative fixtures | #87 / A1 |
| Controlled Information decomposition | #94 / A9 |
| access-semantics dependency | #89 / A8 |

No new root-cause issue is justified by this evidence.
