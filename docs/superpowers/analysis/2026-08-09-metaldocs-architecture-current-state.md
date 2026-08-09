# MetalDocs Architecture — Current-State Evidence Map

**Date:** 2026-08-09  
**Baseline `main`:** `418070bf38a9f358f9131bcc36b7a6bcbc069273`  
**Status:** current-state evidence snapshot; descriptive, not a target-architecture decision.

## 0. Evidence provenance

This audit is pinned to `main` commit `418070bf38a9f358f9131bcc36b7a6bcbc069273`, which includes merged PR #99 (`only-new-issues` removed; golangci whole-tree burn-down).

Evidence classes:

- **runtime/repository direct:** current files and directory topology read from GitHub at the baseline SHA;
- **mechanically reproduced inventory:** checked-in `docs/superpowers/analysis/inventory/*` outputs whose commands include `go list`, grep, fan-in/fan-out and Tarjan SCC analysis;
- **issue/ADR finding:** prior measured evidence that remains useful but must be checked for staleness against the baseline before execution;
- **historical wiki artifact:** useful for evolution/history only when it conflicts with current runtime topology.

The current session cannot clone GitHub into its local terminal because that environment cannot resolve `github.com`; therefore no claim is labeled as a fresh local rerun unless it is already represented by checked-in mechanical evidence. Static source inspection and GitHub-index evidence remain available directly.

## 1. Repository/runtime shape

MetalDocs currently operates as a mixed-language monorepo centered on a Go modular monolith. The documented runtime shape contains:

- `metaldocs-api` — synchronous data/control plane;
- `metaldocs-worker` — asynchronous outbox/worker plane;
- `metaldocs-jobs` — scheduled/River jobs plane;
- `apps/docx-renderer` — separate Node.js rendering service;
- Postgres, Redis and MinIO as state/supporting infrastructure.

The Go backend uses a single module (`module metaldocs`) with business code under `internal/modules/`, technical platform code under `internal/platform/`, and composition-specific wiring also present under executable roots and `internal/composition/`.

## 2. Actual current module inventory

At the pinned baseline, `internal/modules/` contains **15** directories:

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

Older topology documents that report 11/12 modules are stale on this fact. The existence of a current module directory is implementation evidence only; it does not prove that the directory is the correct long-term bounded context.

## 3. Current coupling model — hotspot view

The measured August evidence supports this qualitative hotspot view:

```mermaid
flowchart LR
    DOC[documents]
    CD[controlleddocuments]
    TMP[templates]
    APR[approval]
    IAM[iam]
    SEC[security]
    TAX[taxonomy]
    PLAT[internal/platform]

    DOC <--> APR
    CD <--> APR
    TMP <--> DOC
    APR <--> TAX
    PLAT --> IAM
    PLAT --> DOC
    PLAT --> APR
    SEC --> IAM
```

This diagram is intentionally not presented as the complete adjacency matrix. The checked-in mechanical layering inventory reports:

- 136 Go packages analyzed;
- zero multi-node package-level SCCs;
- seven reciprocal relationships after collapsing package paths to module identity;
- zero same-module `domain -> infrastructure/delivery` and `application -> delivery` import inversions;
- multiple cross-module producer-owned seams, foreign-table SQL reads, foreign sentinel contracts and platform→module edges.

## 4. Coupling dimensions that must be kept separate

### 4.1 Go-import coupling (`G`)

The current boundary checker catches imports into forbidden implementation layers but permits cross-module imports into selected published layers/packages. It does not reject reciprocal legal edges and therefore is not a module cycle detector.

### 4.2 SQL/data coupling (`S`)

The layering inventory reproduces 17+ cross-context table reads across Approval/Documents/ControlledDocuments. These bypass Go package boundaries and are invisible to import-only guards.

### 4.3 Error identity coupling (`E`)

The layering inventory reports 62 `errors.Is` call sites depending on another module's raw domain sentinel across six modules. Sentinel identity therefore behaves as an undeclared integration contract.

### 4.4 Foreign type / persistence leakage (`T`)

Nine of fifteen domain packages import `database/sql` and/or `internal/platform/db` in port signatures. This does not mean SQL is executed in domain — the same inventory found no literal domain SQL — but persistence vocabulary leaks into the domain API surface.

### 4.5 Platform inversion (`P`)

The target architecture requires `internal/platform` to be domain-free. The measured inventory identifies module-specific imports from platform packages including `bootstrap`, `authn`, `docgenv2`, `tripwire` and `worker`, separate from legitimate composition wiring.

### 4.6 Composition-only wiring (`W`)

Dependencies located exclusively in executable/composition roots are not automatically architectural coupling defects. Composition is expected to know implementations so that business modules do not have to.

## 5. Current positive exemplar: consumer-owned port

`documents/application` declares `DictionaryValueReader`, a Documents-owned capability requiring only `Lookup(...)`.

The API composition root implements an adapter backed by Tokens and translates `tokensdomain.ErrNotFound` into `(found=false)` at the boundary. Documents therefore imports no Tokens domain type/error.

This is the local exemplar for A4:

```text
Documents application (consumer-owned port)
            ^
            |
composition adapter
            |
Tokens producer
```

## 6. Current counter-exemplar: producer-owned seams

Current `approval/application/decision_service.go` directly imports:

- `controlleddocuments/domain`;
- `documents/application`;
- `documents/domain`;
- `iam/domain`.

`PinInvoker` currently carries `docapp.ApproverContext`, and other Approval seams consume producer-declared readers/types. The finding is not that any cross-module collaboration is forbidden; it is that the contract is shaped by producer internals rather than by the consumer's minimum capability.

## 7. Controlled Information boundary

ADR 0093 and issue #94 supersede the assumption that `documents`, `templates`, and `controlleddocuments` are three peer bounded contexts.

Current implementation topology:

```text
documents
controlleddocuments
templates
approval
```

Target domain ruling already made:

```text
Controlled Information
├── ControlledDocument
├── DocumentRevision
├── NumberSeries
└── TemplateUsePolicy

Approval & Evidence
└── subject-generic approval lifecycle/evidence
```

The audit maps the current implementation into this target without using current tables/directories as evidence for preserving the split.

## 8. Current architecture guard blind spots

### `scripts/check-module-boundaries.ps1`

Proves:

- a cross-module import does not target prohibited implementation layers subject to its configured published surface.

Does not prove:

- the collapsed module graph is acyclic;
- the dependency is semantically owned by the consumer;
- a producer-declared interface is a healthy seam;
- foreign SQL/table access is absent;
- foreign error/sentinel identity is absent;
- platform is domain-free;
- a published `application`/`domain` type is a minimal stable contract.

This is a valid guard for one property, not an architecture-completeness proof.

## 9. Transaction/persistence finding

`internal/platform/db.TxRunner` explicitly states that it owns begin/commit/rollback and exposes live `*sql.Tx` in the callback as a bounded concession required by current authz/catalogue behavior.

That concession does **not** justify spreading driver/persistence types into business-domain APIs. Current policy direction captured in the rulebook is:

- application/infrastructure transaction seams may use the shared transaction abstraction where atomicity requires it;
- new `domain` APIs do not gain `*sql.Tx`, `sql.Null*`, `sql.Row`, `sql.Result`, or platform DB types without an explicit ruling;
- A5 owns migration toward one transaction lifecycle mechanism and typed query machinery.

The persistence inventory also records 82 direct `BeginTx` sites across 25 files and 242 hand-maintained scan sites, which remain A5 evidence at this baseline unless changed by later merges.

## 10. Existing root-cause programs

| Program | Root cause | Architecture-audit relation |
|---|---|---|
| #87 / A1 | verifier is not one trusted product | architecture guards must join one reproducible verifier with negative fixtures |
| #91 / A2 | no ratcheted whole-repo quality baseline | architecture metrics should ratchet regressions, not trigger count-driven rewrites |
| #90 / A3 | API contract stops before runtime behaviour | owns error/validation/generated-contract runtime convergence |
| #93 / A4 | module seams expose implementations, not capabilities | primary owner of cross-module contracts, module cycles, SQL ownership and platform inversions |
| #92 / A5 | persistence correctness maintained by visual agreement | owns transaction/query/scan/driver mechanics |
| #88 / A6 | security properties are configuration-contingent | owns fail-closed runtime/deployment security preconditions |
| #95 / A7 | async operations have no closed feedback loop | owns worker/jobs observability and trace/liveness closure |
| #89 / A8 | authorization grant model is dual-source | must settle access semantics before migrations depending on them |
| #94 / A9 | Controlled Information implemented as three peer contexts | owns the ruled domain consolidation |

## 11. Staleness correction: A1/A2 evidence after PR #97/#99

The original issue bodies are findings at filing time, not immutable current-state facts.

At baseline `418070bf...`:

- workflow topology has already been collapsed substantially (current `.github/workflows/` contains `ci.yml`, `docx-renderer.yml`, `nightly.yml`, `release.yml`, `smoke.yml`, not the earlier 20-workflow layout);
- `tools/verify/` now exists as a significant Go verifier/registry with tests;
- PR #99 removed `only-new-issues` and reports a whole-tree golangci burn-down to zero for its configured scope.

Therefore those specific A1/A2 symptoms must not be re-implemented. The remaining work should be evaluated against #87/#91 **acceptance properties**, not the stale counts in their original bodies.

## 12. Preliminary sequencing

```mermaid
flowchart TD
    A1[#87 A1 verifier spine]
    A3[#90 A3 contract/runtime]
    A8[#89 A8 authz single source]
    A4[#93 A4 module seams]
    A5[#92 A5 persistence]
    A9[#94 A9 controlled information]
    A7[#95 A7 async feedback]
    A6[#88 A6 fail-closed security]
    A2[#91 A2 ratcheted quality]

    A1 --> A3
    A1 --> A8
    A1 --> A6
    A1 --> A2
    A8 --> A4
    A3 --> A4
    A4 --> A5
    A8 --> A9
    A4 --> A9
    A3 --> A7
```

This is dependency guidance derived from owning issue/ADR constraints; it is not a mega-PR instruction.

## 13. Findings already supported strongly enough not to rediscover

### F-AUD-01 — Module cycles are an architecture property missing from the legacy checker

**Evidence:** mechanically reproduced module-pair analysis reports seven reciprocal relationships while package SCC count is zero.

**Owner:** #93 / A4.

**Acceptance property:** future verifier derives the Go package graph, collapses paths to module identity, and rejects prohibited reciprocal/SCC relationships mechanically.

### F-AUD-02 — Import boundaries alone cannot enforce data ownership

**Evidence:** 17+ foreign-table SQL reads reproduced in the layering lane.

**Owner:** #93 / A4 (with A9 absorbing seams that become intra-context after ADR 0093 execution).

**Acceptance property:** a single machine-readable table-ownership catalog plus SQL identifier analysis makes foreign-table coupling visible at verification time.

### F-AUD-03 — Raw foreign sentinels currently form undeclared integration APIs

**Evidence:** 62 cross-module `errors.Is` sites in the measured layering inventory.

**Owner:** #93 / A4, coordinated with #90 / A3 for HTTP translation.

**Acceptance property:** consumer-owned stable result/error contracts at module seams; producer raw sentinels are translated at adapters/boundaries.

### F-AUD-04 — Platform/domain direction is not consistently inward

**Evidence:** platform→module edges reproduced in the layering lane; target REQ-TOP-2 forbids them.

**Owner:** #93 / A4.

**Acceptance property:** platform remains generic/domain-free; module-specific wiring moves to explicit composition/module locations.

### F-AUD-05 — Wiki topology/maturity contains stale claims

Evidence includes old dependency artifacts that still describe nested `documents/approval` and deny the existence of top-level Approval, plus architecture docs carrying older module counts/maturity labels.

**Class:** wiki-memory drift.

**Disposition:** update as part of audit/module-doc sync; do not use stale wiki as target evidence.

## 14. Issue creation policy

Do not create a new root-cause issue merely because another call site is found. Create one only when:

- the finding has a distinct root cause not owned by #87-#95;
- the required target property conflicts with an existing program;
- the finding blocks sequencing but has no owner;
- or the acceptance mechanism materially exceeds the owning issue's charter.

Implementation sub-slices may be planned under the owning issue without pretending each symptom is a new architectural cause.

## 15. Companion rulebook

Engineering rules for module ownership, consumer-owned ports, data ownership, layering, platform/composition, transaction spine, API/errors and architecture verification live in:

`docs/superpowers/analysis/2026-08-09-metaldocs-architecture-engineering-rulebook.md`
