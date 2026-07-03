# Feature F1.1 — oasdiff breaking-change gate — Spec

> **Milestone:** 1 — Contract & frontend governance gates  ·  **Folder:** `f1.1-oasdiff-gate`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (operator) — contract fully specified in
> `../validation-contract.md §F1.1`; this feature implements it verbatim.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | Contract is fully derived from mission.md §7 M1 (F1.1 row) + `validation-contract.md §F1.1`, which fix the tool (oasdiff), the command shape, the base-vs-PR comparison, the CI wiring point, and the exact positive/negative proofs. Nothing was guessed; the consumer (CI + PR authors) contract is explicit. |
| 2 | oasdiff vs redocly changelog? | oasdiff (mission-named "oasdiff or equivalent"); it is the industry-standard OpenAPI breaking-change detector and gives a clean non-zero exit on breaking diffs. |
| 3 | Base spec source in CI? | The PR base ref's `api/openapi/v1/openapi.yaml` (via `actions/checkout` of the base or `git show <base>:<path>`), diffed against PR head. |

## Consumer contract (FIRST)

- **Consumer(s):** GitHub Actions CI on `pull_request`; the PR author (who must see a named failure on
  a breaking spec change); the mission's terminal gate inventory (needs a recorded negative proof).
- **Contract:** a **blocking** CI job that, on any PR modifying `api/openapi/v1/openapi.yaml`, runs
  `oasdiff` in breaking-change mode comparing base→head and exits **non-zero** iff there is an
  OpenAPI-level breaking change; exits **zero** on compatible/no change.
- **Source of truth for the contract:** `validation-contract.md §F1.1`; existing CI conventions in
  `.github/workflows/api-contract.yml` (checkout/setup-go patterns, path filters, blocking posture).

## What this feature implements

A new blocking workflow `.github/workflows/openapi-breaking.yml` (or a new job in `api-contract.yml`)
that installs oasdiff, resolves the base-branch spec, runs `oasdiff breaking <base> <head>` with a
fail-on-breaking flag, and fails the check on any breaking change. Path-filtered to the spec file +
the workflow itself.

## Non-goals (mandatory)

- No changes to `api/openapi/v1/openapi.yaml` content (F1.2 owns the one struct fix).
- No non-breaking-change linting (that's redocly / api-lint — other features).
- No auto-approval or auto-labeling of breaking changes; the gate only reports pass/fail.
- No changelog publishing / release automation.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Gate green on identical/compatible spec | `oasdiff breaking <spec> <spec>` → "No breaking changes", exit 0 | real |
| Gate red on a synthetic breaking change | `oasdiff breaking <orig> <broken>` (broken = a deleted path or a required field removed) → non-zero, names the break | fixture (temp broken spec, discarded) |
| Wired blocking in CI, correct path filter | workflow YAML present; `on: pull_request: paths:[api/openapi/v1/openapi.yaml, .github/workflows/openapi-breaking.yml]`; no `continue-on-error` | real |
| Workflow syntactically valid | YAML parses; job/step structure matches repo conventions | real |

> TDD note: for a CI-workflow feature, "test first" = author the NEGATIVE fixture and prove the gate
> command fails on it **before** wiring the workflow, then prove the workflow passes on the clean
> spec. Both outputs captured in `evidence.md`.

## ADR needed?

- [x] No durable architecture decision — this is enforcement tooling implementing an existing mission
  decision (contract-first governance). Skip.
