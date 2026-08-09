# MetalDocs Architecture Audit — Reproduced Inventory Evidence

**Date:** 2026-08-09  
**Status:** connector-retrieved evidence from an existing mechanically-produced inventory; not re-executed in this ChatGPT runtime.

## 1. Provenance

The repository already contains a prior architecture discovery artifact at:

- `docs/superpowers/analysis/inventory/layering.md`

That artifact records commands and computed results from a local repository pass, including `go list -json ./internal/...`, Tarjan SCC analysis, fan-in/fan-out calculation, symmetric module-edge analysis, grep-based SQL checks, and direct source inspection.

This file promotes the parts relevant to the current architecture-audit program into one traceable snapshot. It does **not** claim that this ChatGPT session re-ran those local commands. The terminal available in this session could not resolve `github.com`, so local clone/re-execution remains a workstation verification step.

## 2. Graph facts already mechanically measured

The existing inventory reports:

- **136 first-party packages** in the `./internal/...` Go graph;
- **0 multi-node package-level SCCs** under Tarjan analysis — expected because Go forbids import cycles;
- **7 module-level cycles** after collapsing package edges to module identity / testing symmetric module edges;
- **0 intra-module layer inversions** where a module's own `domain` imports its own `infrastructure`/`delivery`, or its own `application` imports its own `delivery`;
- `iam/domain` fan-in = **32 importers**;
- `iam/authz` fan-in = **26 importers**;
- `platform/db` fan-in = **32 importers**;
- `platform/httprouter` fan-in = **21 importers**.

The critical distinction is therefore:

```text
package graph acyclic  !=  module graph acyclic
```

The current Go package structure successfully routes around compiler-level import cycles while still allowing reciprocal semantic dependencies between bounded-context candidates.

## 3. Persistence leakage through domain contracts

The inventory records two related classes:

### LAYERING-01

Direct `database/sql` types leak into domain port signatures, including `*sql.Tx` and `sql.NullTime`.

Sampled evidence includes:

- `internal/modules/approval/domain/release_hold_port.go`
- `internal/modules/approval/domain/sla_port.go`
- `internal/modules/auth/domain/session_admin.go`
- `internal/modules/documents/domain/review_due_port.go`
- `internal/modules/documents/domain/review_surface_port.go`
- `internal/modules/security/domain/tenant_crypto.go`
- `internal/modules/tokens/domain/port.go`

Scale recorded: **7 files, 5 modules, ~20 signatures**.

### LAYERING-02

Domain packages also import `internal/platform/db` transaction plumbing in port signatures.

Scale recorded: **9 of 15 module domain packages** import either `database/sql` and/or `platform/db`.

This is not being split into nine defects. It is a repo-wide seam convention that belongs to the architecture/persistence programs (#93 / A4 with #92 / A5).

## 4. Producer-owned contracts at module seams

The existing inventory confirms both directions of the same design problem.

### Consumer reaches into producer types/functions

Examples recorded:

- `documents/application` constructs `controlleddocumentsdomain.TemplateVersionCandidate` and calls `controlleddocumentsdomain.Resolve(...)`;
- `documents/application` exposes `templatesdomain.Placeholder` in a method signature;
- `approval/application` accepts `docapp.ApproverContext`;
- `approval/application` calls `docapp.LoadDocumentAreaCode(...)`.

### Consumer is typed by producer-declared interfaces

Examples recorded:

- `security/infrastructure/postgres` consumes multiple interfaces declared by `iam/domain`;
- `iam/infrastructure/postgres` consumes `taxonomydomain.AreaCatalogReader` declared by `taxonomy/domain`.

### Positive counter-example

`documents/application` declares its own `DictionaryValueReader`, and the composition root supplies an adapter in `apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go`.

This is the pattern the remediation program should generalize:

```text
consumer owns required capability contract
        ↓
composition root owns adapter
        ↓
producer implementation stays private
```

Root-cause owner: **#93 / A4**.

## 5. Foreign error identity as an undeclared contract

The inventory reproduces:

- **62** `errors.Is(err, <foreign>domain.Err...)` call sites;
- spread across **6 modules**: approval, auth, controlleddocuments, documents, iam, templates.

The important architectural property is not the raw count. It is that sentinel identity is functioning as a cross-context API without an explicit contract boundary.

Root-cause ownership:

- seam design: **#93 / A4**;
- HTTP/runtime error translation conventions: **#90 / A3**.

## 6. Cross-module SQL / data ownership

The existing inventory gives concrete evidence for at least **17+ foreign-table reads** across three module directions.

### Approval -> Documents

`internal/modules/approval/infrastructure/postgres_approval_repository.go` directly reads `documents` / `document_comments` through multiple `FROM` / `JOIN` sites.

Recorded scale: **9+ sites** in this direction.

### Documents -> Approval

Documents code directly reads tables owned by Approval, including:

- `approval_instances`
- `approval_signoffs`
- `release_generations`

Recorded evidence spans:

- `internal/modules/documents/infrastructure/active_instance_reader.go`
- `internal/modules/documents/infrastructure/repository.go`
- `internal/modules/documents/infrastructure/resolver_readers.go`
- `internal/modules/documents/application/context_builder.go`

Recorded scale: **7 sites**.

### Approval -> Controlled Documents

`internal/modules/approval/application/read_service.go` directly joins `controlled_documents`.

Recorded scale: **1 site**.

This confirms that Go-import guards alone cannot prove bounded-context data ownership.

Root-cause owner: **#93 / A4**, coordinated with #92 / A5 for persistence mechanisms.

## 7. Platform -> module inversion

The prior inventory distinguishes composition-shaped wiring from generic platform concerns.

### Composition registry

`internal/composition/tenantdata/registry` imports **12 of 15 modules' infrastructure packages** to register tenant-data ports. This package is composition-root-shaped and should be evaluated as composition, not blindly counted as a platform violation.

### Module-specific dependencies housed under `internal/platform`

The inventory records five platform packages with module-specific imports outside that registry:

- `bootstrap`
- `authn`
- `docgenv2`
- `tripwire`
- `worker`

Recorded scale: **11 module edges** across those five packages.

Examples include:

- `platform/bootstrap` -> IAM/Auth/Audit domain/infrastructure packages;
- `platform/authn` -> Auth application + IAM domain;
- `platform/docgenv2` -> Documents application/domain;
- `platform/tripwire` -> IAM domain;
- `platform/worker` -> IAM authz.

This is direct evidence against the target rule that platform remain domain-free, while also showing why composition wiring needs its own classification (`W`) rather than being mixed into `P`.

Root-cause owner: **#93 / A4**.

## 8. Architecture properties that are currently healthy

The audit must preserve positive evidence and avoid turning every unusual shape into a defect.

The existing inventory records:

1. **0 package-level import cycles**;
2. **0 intra-module layer inversions** at package-import granularity;
3. domain packages contain **no literal SQL execution** — raw SQL is outside domain;
4. explicit composition-root hand-wiring is legible and no DI framework is hiding ownership;
5. the `DictionaryValueReader` adapter demonstrates a correct consumer-owned seam already exists in production code.

These should become regression-preservation properties during remediation.

## 9. What remains mechanically unresolved

Even with the existing inventory, the current audit still needs a workstation pass for exact reproducibility and artifact generation:

1. record the exact current `main` SHA and Go version;
2. re-run the 136-package graph against that SHA;
3. emit the **exact seven module-cycle pairs/SCC membership** as a generated artifact rather than relying on prose;
4. generate a machine-readable module adjacency matrix with per-edge source paths;
5. reconcile the 15-module inventory with older topology docs that list fewer modules;
6. derive foreign SQL edges from a single machine-readable ownership catalog;
7. classify all 62 foreign-sentinel call sites by layer (delivery/application/etc.);
8. distinguish legitimate shared vocabulary from accidental producer-owned contracts;
9. verify whether `*sql.Tx` / `platform/db.Tx` in domain ports has an explicit ADR or is an unlabeled convention;
10. feed any future guard into #87's single verifier with a proven negative fixture.

## 10. Remediation ownership — no new root cause discovered yet

Current evidence still fits the existing program:

| Finding family | Existing owner |
|---|---|
| module cycles / seam direction / producer-owned contracts | #93 / A4 |
| foreign SQL ownership | #93 / A4 + #92 / A5 |
| foreign error identity | #93 / A4 + #90 / A3 |
| domain persistence-type leakage | #93 / A4 + #92 / A5 |
| platform -> module direction | #93 / A4 |
| verifier reachability / negative fixtures | #87 / A1 |
| Controlled Information decomposition | #94 / A9 |
| access-semantics dependency | #89 / A8 |

**Decision:** do not create another remediation issue from this evidence. The audit umbrella remains #100, and #87–#95 remain the implementation owners.
