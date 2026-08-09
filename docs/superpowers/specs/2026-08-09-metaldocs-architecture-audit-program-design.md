# MetalDocs Architecture Audit Program — Design

**Status:** proposed
**Date:** 2026-08-09
**Scope:** analysis and documentation only; no product/runtime refactor is authorized by this document.

## 1. Purpose

Create a reproducible, evidence-first map of the MetalDocs architecture as it exists now, then trace each material finding to its owning root-cause program. The audit must separate factual current-state evidence from target-architecture decisions so that the current implementation cannot be used as evidence that its own structure is correct.

The immediate operator goal is to arrive at implementation time with the architecture already mapped, findings classified, existing issues reused, missing work isolated, and sequencing explicit.

## 2. Non-goals

- Do not refactor production code.
- Do not reopen ADR 0092 or ADR 0093 without a new material finding.
- Do not create one issue per observed edge, duplicate, or cycle.
- Do not treat LOC, file count, import count, or cycle count alone as a decomposition decision.
- Do not convert the modular monolith into microservices.
- Do not replace explicit composition-root wiring with a DI container.

## 3. Two-layer audit model

The audit has two deliberately separate layers.

### Layer A — Current-State Architecture Map

This layer is descriptive only. It records what exists and how it is coupled.

Required views:

1. Runtime topology and composition roots.
2. Go package dependency graph.
3. Module-to-module dependency graph.
4. Strongly connected components / cycle inventory.
5. Database ownership graph.
6. Cross-module SQL dependency graph.
7. Foreign domain type and sentinel/error coupling.
8. Platform-to-module dependency inversions.
9. Public contracts and port ownership.
10. Transaction ownership and cross-context transaction leakage.
11. Frontend-to-backend contract coupling.
12. Existing architecture/mechanical gates and their blind spots.
13. Runtime/contract/wiki/execution truth drift.

Every edge must be classified by mechanism rather than collapsed into a generic dependency:

- `G` — Go import.
- `S` — SQL/table knowledge.
- `T` — foreign domain/type coupling.
- `E` — foreign sentinel/error identity coupling.
- `C` — explicit stable contract/port.
- `P` — platform inversion.
- `W` — composition-root wiring only.

### Layer B — Architecture Remediation Map

This layer is prescriptive only after Layer A evidence is recorded.

For each material finding, record:

- observed symptom;
- evidence path/count;
- root-cause family;
- owning bounded context or platform concern;
- existing issue/ADR that subsumes it;
- sequencing dependency;
- target property;
- acceptance mechanism that makes regression mechanically visible or unrepresentable.

New GitHub issues are created only when a material finding is not already subsumed by an existing program.

## 4. Evidence rules

1. Runtime truth outranks wiki claims about what currently exists.
2. Machine-derived counts outrank hand-written inventories when both exist.
3. Current schema/module/import topology is admissible for describing current state, but inadmissible as proof that the topology is the correct target decomposition.
4. A structural judgment must pass the inversion test from ME-13: state which conclusion would survive if the current implementation had the opposite structure.
5. A guard is credited only for the semantic property it actually proves. Syntactic proxies must list their blind spots.
6. A finding is not closed by prose. Closure requires runtime assertion, compile/static enforcement, generated single-source derivation, or an explicitly bounded residual where stronger enforcement is impossible.

## 5. Initial evidence baseline

The audit starts from evidence already established by the August architecture program, especially issues #87–#95.

### Module seams

Issue #93 records:

- 17+ cross-module/foreign-table SQL reads;
- 62 `errors.Is(err, <foreign>domain.Err...)` sites across six modules;
- 9 of 15 domain packages exposing `database/sql` or platform DB concepts in port signatures;
- 20 platform→module import edges across six platform packages, plus additional edges outside `tenantdata`;
- 7 known module cycles that the current checker does not reject;
- 12 hand-written `tenant_data_port.go` skeletons;
- `approval/application` as a major coupling hotspot.

This establishes that REQ-TOP-1/2 are target properties, not yet universally satisfied runtime properties.

### Controlled Information decomposition

Issue #94 plus ADR 0093 rules that `documents`, `templates`, and `controlleddocuments` are not to be defended as three peer bounded contexts merely because the current implementation has three modules/tables. The target domain is Controlled Information with explicit aggregate ownership, while Approval & Evidence remains subject-generic and separate.

### API/runtime contract

Issue #90 records parallel error writers, runtime validation gaps, duplicate identity helpers, request-shape drift, pagination dialects, and generated-contract bypasses. The audit treats these as contract-boundary findings, not isolated handler cleanup.

### Verification system

Issue #87 establishes that architecture/quality claims cannot rely on dormant scripts or CI jobs that do not exercise them. Any new architecture guard created by the remediation program must enter the future single verification product and include a negative fixture.

## 6. Known blind spot in the current module-boundary checker

`scripts/check-module-boundaries.ps1` validates whether a cross-module import targets an allowed layer (`domain`, `application`, `api`) or explicit published package. It does not build the directed module graph or reject reciprocal allowed imports.

Therefore both edges below can be individually legal to the script while forming a cycle:

```text
documents -> approval/application
approval  -> documents/domain
```

The audit must distinguish:

- layer-visibility violation;
- module-cycle violation;
- semantic ownership violation;
- SQL/data-ownership violation.

One checker cannot be assumed to prove all four.

## 7. Output artifacts

The program produces:

1. this design spec;
2. `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md` — evidence snapshot and remediation traceability table;
3. `docs/superpowers/plans/2026-08-09-metaldocs-architecture-audit-program.md` — executable analysis plan for the workstation session;
4. one umbrella GitHub issue linking #87–#95 and any genuinely uncovered findings;
5. one draft PR as the review checkpoint for the audit artifacts.

## 8. Success criteria

The audit is ready for implementation planning only when:

- every business module and platform package appears in the dependency inventory;
- all module SCCs/cycles are enumerated mechanically;
- cross-module SQL/table ownership is inventoried from a machine-readable ownership catalog or equivalent generated scan;
- foreign error/type coupling is sized and attributed;
- platform→module inversions are enumerated;
- each material finding maps to an existing issue or a newly justified issue;
- each remediation item names a semantic acceptance property, not merely a file edit;
- no new product implementation is mixed into the audit PR.

## 9. Review checkpoint

This branch and its draft PR are the only checkpoint for this audit design. Approval of the audit artifacts authorizes later implementation planning, not implementation itself.