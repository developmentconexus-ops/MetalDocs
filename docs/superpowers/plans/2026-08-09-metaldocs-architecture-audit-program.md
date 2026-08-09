# MetalDocs Architecture Audit Program Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a reproducible, evidence-first current-state architecture map and remediation traceability package without changing product/runtime behavior.

**Architecture:** The audit is split into a descriptive Layer A (current-state graph/data/contracts/gates) and a prescriptive Layer B (root-cause ownership and sequencing). Machine-derived evidence is preferred over hand-maintained inventories. Existing issues #87–#95 are reused unless a materially distinct root cause is proven.

**Tech Stack:** Go package graph, repository SQL/source scanning, PowerShell/Go verifier tooling where appropriate, Mermaid/Markdown documentation, GitHub issues/PRs.

## Global Constraints

- Analysis/documentation only in this program; no product/runtime refactor.
- Do not reopen ADR 0092 or ADR 0093 without a new material finding.
- Current implementation topology may describe current state but may not prove target decomposition.
- Composition-root wiring must not be counted as domain coupling by default.
- New issues require a root cause not already subsumed by #87–#95.
- Any future architecture guard must join the single verifier program from #87 and include a negative fixture.

---

### Task 1: Freeze the evidence baseline

**Files:**
- Read: `AGENTS.md`
- Read: `wiki/backend/repo-topology.md`
- Read: `wiki/architecture/backend-target-architecture.md`
- Read: `scripts/check-module-boundaries.ps1`
- Read: `wiki/decisions/0092-*`
- Read: `wiki/decisions/0093-*`
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Consumes: current runtime/wiki truth and issue evidence #87–#95.
- Produces: a dated evidence header containing the exact `main` commit SHA used by all later measurements.

- [ ] **Step 1:** Record `git rev-parse HEAD`, `go version`, and repository dirty status before measurements.
- [ ] **Step 2:** Confirm the current module list from `internal/modules/` and platform package list from `internal/platform/`.
- [ ] **Step 3:** Re-read #87–#95 and mark each quoted count as either `reproduced`, `stale`, or `not-yet-reproduced`.
- [ ] **Step 4:** Commit only the evidence-header update.

### Task 2: Build the exact Go package and module graph

**Files:**
- Create or extend verifier tooling only under the repository's existing verification/tooling convention after checking #87.
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Consumes: `go list -deps -json ./...` or an equivalent Go-native package graph source.
- Produces: directed package edges, collapsed module edges, SCC/cycle list, and per-edge source evidence.

- [ ] **Step 1:** Extract all first-party Go package imports under `metaldocs/internal/**`.
- [ ] **Step 2:** Collapse paths under `internal/modules/<name>/...` to module identity while preserving package-level evidence.
- [ ] **Step 3:** Exclude same-module edges and classify composition-root wiring separately as `W`.
- [ ] **Step 4:** Compute strongly connected components on the directed module graph.
- [ ] **Step 5:** Verify whether the 7-cycle count from #93 still holds; record exact SCC members rather than a hand-written cycle list.
- [ ] **Step 6:** Compare SCC results with `scripts/check-module-boundaries.ps1` and document every false-negative class.
- [ ] **Step 7:** Commit graph evidence and the generated/reproducible command, not a manually curated kill list.

### Task 3: Map database ownership and cross-module SQL edges

**Files:**
- Read: database ownership/wiki/baseline sources under `wiki/database/` and `db/`.
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Consumes: table ownership declarations and SQL identifiers used in module repositories/application code.
- Produces: `module -> foreign table owner` edges classified `S` with source paths.

- [ ] **Step 1:** Identify the closest existing machine-readable source of schema/table ownership; do not create a second hand-synced ownership list if one already exists.
- [ ] **Step 2:** Enumerate SQL-issuing production Go files by module.
- [ ] **Step 3:** Extract referenced owned tables and flag reads/writes whose owner differs from the caller module.
- [ ] **Step 4:** Reproduce or correct the `17+` foreign-table count from #93.
- [ ] **Step 5:** Distinguish legitimate shared/platform tables from bounded-context tables explicitly.
- [ ] **Step 6:** Record the minimal future enforcement property: ownership as data plus a verifier over SQL usage.

### Task 4: Map foreign type, sentinel, and transaction leakage

**Files:**
- Scan: `internal/modules/**/*.go`
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Produces: `T` and `E` edge inventories and domain-port persistence-leak inventory.

- [ ] **Step 1:** Enumerate `errors.Is` / `errors.As` references whose target error originates in another module.
- [ ] **Step 2:** Reproduce or correct the 62 foreign-sentinel sites from #93 and group them by producer/consumer pair.
- [ ] **Step 3:** Enumerate cross-module parameters/returns using foreign domain structs/types where they serve as seam contracts.
- [ ] **Step 4:** Enumerate domain/application interfaces exposing `*sql.Tx`, `sql.Null*`, DB/platform transaction types, or concrete repository types.
- [ ] **Step 5:** Classify each seam as consumer-owned port, producer-owned port, concrete implementation dependency, or ambiguous.
- [ ] **Step 6:** Record only root-cause families; do not open per-call-site issues.

### Task 5: Prove platform layering direction

**Files:**
- Scan: `internal/platform/**/*.go`
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Produces: exact `platform -> module` edge list with owning platform package and target module.

- [ ] **Step 1:** Enumerate all imports from `internal/platform/**` into `internal/modules/**`.
- [ ] **Step 2:** Reproduce or correct the #93 counts and separate `tenantdata` from other inversions.
- [ ] **Step 3:** Determine whether any listed path is actually a misplaced composition root rather than reusable platform code.
- [ ] **Step 4:** Record the future verifier rule corresponding to REQ-TOP-2.

### Task 6: Audit architecture guards by semantic property

**Files:**
- Read: `scripts/check-module-boundaries.ps1`
- Read: `scripts/api-lint/**`
- Read: `tools/cilint/**`
- Read: relevant `.github/workflows/**`
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Produces: guard matrix `guard -> property proven -> blind spots -> CI/verifier reachability`.

- [ ] **Step 1:** Inventory architecture-related scripts/tools and identify whether #87's verifier problem currently reaches them.
- [ ] **Step 2:** For each guard, write one sentence naming the semantic property it proves.
- [ ] **Step 3:** Write one sentence naming what it does not prove.
- [ ] **Step 4:** Confirm whether each blocking guard has a negative fixture that demonstrates failure.
- [ ] **Step 5:** Flag syntactic-proxy guards that can stay green while the semantic defect reappears, following ME-12.

### Task 7: Reconcile wiki maturity claims with current evidence

**Files:**
- Read/Modify as appropriate: `wiki/architecture/backend-blueprint.md`
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Produces: explicit distinction among `target`, `historically completed milestone`, and `current residual debt`.

- [ ] **Step 1:** Review every `✅ industry-grade` backend concern against open August findings.
- [ ] **Step 2:** Mark stale claims as wiki-memory drift rather than silently rewriting history.
- [ ] **Step 3:** Propose the smallest documentation correction that preserves historical milestone evidence while accurately exposing current residual debt.
- [ ] **Step 4:** Do not create a separate implementation issue unless the drift itself blocks execution semantics.

### Task 8: Produce the remediation traceability matrix

**Files:**
- Modify: `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md`

**Interfaces:**
- Consumes: Tasks 2–7 evidence.
- Produces: one row per material root-cause family.

- [ ] **Step 1:** For every finding, map `symptom -> root cause -> owner -> existing issue -> sequencing -> target property -> acceptance mechanism`.
- [ ] **Step 2:** Mark findings subsumed by #87–#95 explicitly; do not create duplicate issues.
- [ ] **Step 3:** Apply the issue-creation policy from the design spec to any uncovered finding.
- [ ] **Step 4:** If a new issue is justified, write it at root-cause level with sized evidence and a semantic acceptance property.

### Task 9: Final audit review and checkpoint

**Files:**
- Review: all three audit artifacts on this branch.

**Interfaces:**
- Produces: one reviewable draft PR with no runtime/product code changes.

- [ ] **Step 1:** Scan audit docs for `TBD`, `TODO`, unqualified counts, and claims without evidence source.
- [ ] **Step 2:** Verify every structural recommendation passes the ME-13 inversion test.
- [ ] **Step 3:** Verify no finding uses the current implementation as proof of the desired target shape.
- [ ] **Step 4:** Verify the PR changes documentation/audit tooling only and contains no feature/refactor implementation.
- [ ] **Step 5:** Present the audit PR as the operator checkpoint before any remediation implementation plan is authorized.