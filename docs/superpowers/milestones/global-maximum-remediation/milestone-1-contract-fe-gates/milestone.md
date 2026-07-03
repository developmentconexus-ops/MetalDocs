# Milestone 1 — Contract & frontend governance gates

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M1)
> **Status:** Spec approved
> **Authored:** 2026-07-03 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M1 is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-
> milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document and against
> the binding `validation-contract.md` (D4). Drift between implementation and that contract is
> **HS-7**.

## Objective

Make contract truth and frontend module boundaries **machine-enforced from clean state** so the two
defect classes that shipped incidents in the last month become structurally impossible, not
discipline-dependent:

- the **nullable-not-required** wire-shape bug (commit 9f86828b — a field that can be `null` but is
  absent from `required`, so the generated client types it optional and a present-but-null value
  drifts silently), and
- **contract↔runtime↔frontend drift** and **cross-feature frontend coupling**, both currently caught
  only by convention or an advisory (non-CI) script.

The bar this milestone advances: the review's **Dimension 3 (Contract-first API)** and the **FE
boundary** item of the cross-cutting finding move from **DEBT** toward **CONFIRMED** — every gate
below is green from a clean tree **and** has a recorded negative proof (it fails when it should).
This is the P1 "structural drift-proofing" slice (review §Priorities P1 items 3 and 4).

Source findings: review §Cross-cutting items 3, 4, 5; §Dimension 3; §Priorities P1.3 / P1.4.

## Discovered runtime truth (recorded before implementation — HS-6 surface, not silent expansion)

Investigation while authoring this spec (2026-07-03) established ground truth that **refines the
mission's per-feature sizing**. Recorded here so the expansion is visible to the validator and the
operator, not absorbed silently:

1. **F1.3 is script-wide, not "4 templates items."** `check-module-contract-sync.ps1` was written
   against the **pre-AD-1** world (absolute `/api/v1/...` OpenAPI path keys) and an older mount
   layout. The spec has since migrated to **relative path keys** (`  /templates:`, `  /documents:` —
   confirmed openapi.yaml:1105, 2439; enforced by the `PATH-BASE-PREFIX` api-lint rule) and modules
   mount via the **generated boundary** (`templatesapi.HandlerWithOptions`). Result: **every**
   configured module (templates, documents, controlleddocuments, taxonomy, approval) reports DRIFT
   under the stale patterns — the "4 templates DRIFT items" the mission cites are one instance of a
   systemic staleness. The global-maximum fix is to **reconcile the checker to current runtime/spec
   truth for every module it will gate, then promote to blocking** — a checker that stays advisory-
   stale, or is promoted while red on a clean tree, satisfies neither the mission's intent (finding
   3: "wire the check into CI") nor the literal acceptance ("green on clean tree"). Runtime truth
   beats docs (CLAUDE.md). Scope detail and the approval carve-out are in `validation-contract.md §F1.3`.

2. **F1.2 struct burn-down is ~0, not 133.** The redocly comment claims 133 suppressed errors; with
   `struct: error` today the spec produces **exactly 1** error (empty `components.parameters` node,
   openapi.yaml:4290). The burn-down is: fix that 1 node, enable `struct` blocking, record zero (or
   one documented) residual suppression. `operation-summary` / `security-defined` remain scoped
   separately (see contract §F1.2).

3. **F1.4 has no FE ESLint boundary regime today.** The only ESLint config is the root
   `eslint.config.mjs` (eigenpal ACL guard, ADR 0046). Cross-feature relative imports are widespread
   (~50+ genuine cross-feature edges across `src/features/*`), so existing violations are grandfathered
   via a **shrink-only allowlist** (the established repo idiom: css-token-discipline, test-discipline),
   and the gate blocks **new** cross-feature imports. Two real `Omit<>` overrides of generated types
   remain (`features/templates/api/templates.ts:36,44`, `TemplateDTO`/`VersionDTO`); M0 shrank but did
   not delete them. F1.4 removes them.

None of the above crosses a feature's boundary into redesign (no HS-2): each is the assigned feature
executed against runtime truth. The expansion is reported at the HS-1 operator gate.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1.1 | `f1.1-oasdiff-gate` | A breaking-change CI gate (oasdiff) on PRs touching `api/openapi/v1/openapi.yaml`: compares the PR spec against the base-branch spec and fails on any OpenAPI breaking change. Wired in `.github/workflows`. | Gate **fails** a synthetic breaking change (e.g. remove a required response field / delete a path); **passes** a backward-compatible change and a clean tree. Both proven with captured command output. |
| F1.2 | `f1.2-shape-lint` | A blocking **nullable⇒required** shape rule (`SHAPE-NULLABLE-NOT-REQUIRED`) added to `scripts/api-lint` (the proven Go blocking-lint framework), erroring on any schema property that is nullable but absent from its schema's `required` list. Plus re-enable redocly `struct` (fix the 1 live error) with a recorded suppression burn-down (count + owner + trigger). | Lint **fails** a fixture schema with a nullable-not-required field; **current spec passes** the new rule (zero live violations, or each fixed); `redocly lint` with `struct: error` is clean; suppression count recorded. |
| F1.3 | `f1.3-contract-sync-ci` | Reconcile `check-module-contract-sync.ps1` to current runtime/spec truth (relative path keys, generated-boundary mount files) for the gated modules {templates, documents, controlleddocuments, taxonomy}; add a CI wrapper that runs it across those modules and **promote to blocking**. Approval deferred to M9 (§F9.5 module-promotion) with a recorded trigger. | CI **red** on injected drift (a spec path with no runtime owner, or a wrapper/type mismatch); **green** on the clean tree; **zero live DRIFT** across the gated modules. Reconciliation preserves drift-detection power (proven by the injected-drift negative). |
| F1.4 | `f1.4-fe-boundaries` | ESLint feature-boundary rule for `frontend/apps/web/src/features/*` (no cross-feature imports) with an explicit **shrink-only allowlist** grandfathering current edges; remove the two remaining `Omit<>` overrides of generated API types in `features/templates/api/templates.ts`. | ESLint **red** on a synthetic **new** cross-feature import (one not in the allowlist); **green** on the clean tree (allowlisted edges pass); **zero** `Omit<>`-style overrides of generated types remain in `features/*/api/*.ts`; allowlist count recorded with owner/trigger. |

For each feature, "what to validate" is objectively checkable — a gate run that fails on the negative
fixture and passes on the clean tree, with captured command output as evidence (positive + negative
proof per D4). This is a **build-time milestone**: no live API drive is required (mission §7 M1 gate).

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M1:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and each feature's
   consumer contract in `spec.md` is honored. Each gate is demonstrated **failing-then-passing from
   clean state** (negative fixture red → clean tree green), with real captured output.
2. **Validation-contract conformance (D4)** — implementation is checked against `validation-contract.md`
   section-by-section; any divergence is HS-7 (fix code to contract, or re-open the contract with
   operator approval — never silently adjust).
3. **Workflow-class QA** — backend-api contract-gate class + CI-workflow wiring. No screen/runtime
   drive (build-time milestone). QA re-runs the deterministic gates (`go test ./scripts/api-lint/...`,
   `go run ./scripts/api-lint`, redocly lint, the contract-sync wrapper, eslint) from a clean tree.
4. **Regression** — M0 (VersionRef contract) still passes its gate: openapi lint clean, templates pin
   tests green, `go build ./...` clean. No new gate regresses M0's shipped shapes.
5. **Root-cause check** — the 9f86828b nullable-not-required class is **structurally** closed by the
   F1.2 shape rule (not merely the one historical field patched); the drift-detection power of the
   reconciled F1.3 checker is proven live (injected drift → red), not gutted to pass.
6. **No unplanned scope** — anything implemented beyond this list is recorded with rationale. The
   F1.3 approval carve-out and the F1.4 allowlist are recorded bounded defers with triggers.

## Dependencies & constraints

- **Depends on:** M0 (VersionRef contract) passed and committed — F1.2/F1.3 operate on the post-M0
  spec + generated types; F1.4 removes the `Omit<>` overrides M0 shrank.
- **Contract-first (non-negotiable):** any change to the wire contract goes through `api/openapi/v1/openapi.yaml`
  + regeneration (`go generate ./...`, `pnpm run gen:api`); **zero hand-edits to generated files**.
  F1.2's struct fix touches only the spec, then regenerates.
- **CI-truth:** every gate this milestone adds is **blocking** by construction (matches the repo's
  Principle-5 posture — api-lint main.go: "the model is bound by CI, not by discipline"). No
  reported-only/continue-on-error tier.
- **Build-time only:** no live API drive; no migrations; no DB. Targeted commands only — the full
  integration suite is **not** run (20+ min box constraint, mission §10).
- **Known env risk:** `frontend/apps/web` pnpm tree has junction drift breaking vitest/vite (memory
  `fe-node-modules-junction-drift`). ESLint does not need vite; if `pnpm run lint` is blocked locally,
  the F1.4 gate is still demonstrated (npx eslint against the config + a synthetic probe file) and the
  block is recorded as a bounded defer with the complete-pnpm-install trigger.
- **Model policy:** sonnet implement/review; haiku mechanical sweeps; never fable workers; ≤15
  concurrent. Subagent-driven implementation (`superpowers:subagent-driven-development`); the main
  session orchestrates, reviews between features, and commits.
- **Commit after verified work** (standing auth); **never push**; **never commit `docs/release/`**;
  plans dir is gitignored (never force-add).

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M2 and no push without approval. |
| HS-2 | A gate fix implies redesign outside its boundary — e.g. F1.3 reconciliation would require re-mounting a module or changing the approval boundary (that is M9 F9.5). Stop; report; do not patch across the boundary. The approval carve-out exists precisely to avoid this. |
| HS-3 | A prerequisite fails from clean state — `go build ./...` red, `go generate` drift, or M0 gate regressed. Repair the prerequisite first; rerun; resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery mid-milestone beyond the recorded F1.3/F1.4 expansions. Stop; surface; replan. |
| HS-7 (mission) | Implementation deviates from `validation-contract.md` — fix code to the contract, or re-open the contract WITH operator approval; never silently adjust the contract to match the code. |
