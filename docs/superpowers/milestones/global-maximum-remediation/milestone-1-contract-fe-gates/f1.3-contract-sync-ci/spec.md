# Feature F1.3 — contract-sync promoted to blocking CI (reconciled) — Spec

> **Milestone:** 1 — Contract & frontend governance gates  ·  **Folder:** `f1.3-contract-sync-ci`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (operator) — contract in `../validation-contract.md §F1.3`.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | Contract derived from mission.md §7 M1 (F1.3) + `validation-contract.md §F1.3`, grounded in a full read of `scripts/check-module-contract-sync.ps1` and live runs of it across templates/documents/controlleddocuments/taxonomy/approval at author time. |
| 2 | "4 templates DRIFT" only, or wider? | Wider — recorded runtime truth: the checker is stale against the AD-1 relative-path migration + generated-boundary mounts, so ALL modules drift. Global-max fix = reconcile every gated module + promote. Surfaced in milestone.md §Discovered runtime truth (HS-6, not silent). |
| 3 | Is reconciliation a symptom-patch (C6)? | No — the checker's drift-detection power is preserved and PROVEN by the injected-drift NEGATIVE. Reconciliation aligns stale patterns to current correct runtime/spec truth; it does not silence real drift. |
| 4 | Why exclude approval? | approval (`UsesGeneratedBoundary=$false`) has genuine ownership questions entangled with the M9 F9.5 approval-promotion/boundary decision. Reconciling it here crosses that boundary (HS-2). Excluded with a recorded trigger → M9. |

## Consumer contract (FIRST)

- **Consumer(s):** GitHub Actions CI (a new blocking `contract-sync` job); every contract author (a
  spec path with no runtime owner, or a wrapper/type mismatch, must fail their PR); the mission's gate
  inventory (needs recorded negative proof).
- **Contract:**
  1. `check-module-contract-sync.ps1 -Module <m>` exits 0 with no `[DRIFT]` lines for each gated
     module {templates, documents, controlleddocuments, taxonomy} on a clean tree.
  2. A CI wrapper runs the checker across the four and fails (non-zero) if any drifts.
  3. Injecting genuine drift (unowned spec path / renamed wrapper fn / handwritten interface) → wrapper red.
- **Source of truth:** `validation-contract.md §F1.3`; `scripts/check-module-contract-sync.ps1`;
  the live spec (`api/openapi/v1/openapi.yaml`, relative keys) + generated `index.d.ts` + module
  handler mount files.

## What this feature implements

Reconcile `check-module-contract-sync.ps1` config/checks to current runtime truth for the four gated
modules (see validation-contract §F1.3 for the exact stale points: absolute→relative path patterns;
runtime-owner file + mount-token corrections; the inverted `OpenApiForbiddenPatterns`; the
`Test-FrontendGeneratedTypeUsage` regex missing `operations[`/derived aliases). Add a wrapper
`scripts/check-contract-sync-all.ps1` iterating the four modules. Add a blocking `contract-sync` CI
job (`shell: pwsh`) in `.github/workflows/api-contract.yml` (or module-boundaries.yml). Record the
approval carve-out as a bounded defer → M9 F9.5.

## Non-goals (mandatory)

- NOT reconciling / re-mounting / promoting `approval` (M9 F9.5; HS-2 if attempted here).
- NOT changing any module's runtime handlers, the spec, or generated files — only the checker + a
  wrapper + a workflow. (The FE type-drift check aligns to whatever templates.ts is AFTER F1.4; this
  feature runs after F1.4.)
- NOT weakening a check such that injected drift passes (C6 symptom-patch — forbidden).
- NOT adding new module configs beyond the four gated (+ existing approval config left intact but
  un-gated).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Zero DRIFT, templates | `./scripts/check-module-contract-sync.ps1 -Module templates` → no `[DRIFT]`, exit 0 | real |
| Zero DRIFT, documents | `… -Module documents` → no `[DRIFT]`, exit 0 | real |
| Zero DRIFT, controlleddocuments | `… -Module controlleddocuments` → no `[DRIFT]`, exit 0 | real |
| Zero DRIFT, taxonomy | `… -Module taxonomy` → no `[DRIFT]`, exit 0 | real |
| Wrapper green clean tree | `./scripts/check-contract-sync-all.ps1` → exit 0 | real |
| Wrapper red on injected drift | add unowned spec path / rename a wrapper fn / add handwritten interface → wrapper non-zero, names drift+module | fixture (reverted) |
| CI job wired blocking | `contract-sync` job in workflow, `shell: pwsh`, correct path filter, no continue-on-error | real |
| Detection power preserved | the injected-drift negative proves the reconciled checker still catches real drift | fixture |

## ADR needed?

- [x] No durable architecture decision — reconciles + CI-wires an existing advisory checker to
  runtime truth. The approval carve-out references the M9 F9.5 decision (not made here). Skip.
