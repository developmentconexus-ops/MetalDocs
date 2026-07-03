# Feature F1.4 — ESLint feature-boundary rule + remove `Omit<>` overrides — Spec

> **Milestone:** 1 — Contract & frontend governance gates  ·  **Folder:** `f1.4-fe-boundaries`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (operator) — contract in `../validation-contract.md §F1.4`.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | Contract derived from mission.md §7 M1 (F1.4) + `validation-contract.md §F1.4`, grounded in the root `eslint.config.mjs` (read at author time — narrow eigenpal-ACL flat config) and the enumerated `Omit<>` sites (`features/templates/api/templates.ts:36,44`). |
| 2 | New ESLint plugin (eslint-plugin-boundaries) or zero-dep? | Zero-dep. The FE pnpm tree has junction drift + CI runs `--frozen-lockfile`; a new dep is an HS-3 magnet. Use built-in `no-restricted-imports` zones in the existing flat config. |
| 3 | Fix all ~50+ existing cross-feature imports, or grandfather? | Grandfather via an explicit **shrink-only allowlist** (repo idiom: css-token-discipline, test-discipline). Fixing 50+ edges is out of appetite + churn risk. Gate blocks NEW cross-feature imports. |
| 4 | Why remove the `Omit<>` overrides? | They are the drift vector the review names (templates.ts hand-typing over generated types let the wizard bug hide). Post-M0 the generated types carry correct nullability, so the overrides are now redundant. |

## Consumer contract (FIRST)

- **Consumer(s):** the `eslint` CI job in `.github/workflows/lint.yml` (`pnpm run lint`, root
  `eslint.config.mjs`); every FE feature author (a NEW cross-feature import must fail their build);
  the generated-types contract (templates.ts must consume `components['schemas'][...]` directly, no
  hand `Omit<>`).
- **Contract:**
  1. A file in `src/features/<A>/**` importing `src/features/<B>/**` (A≠B), when that edge is **not**
     in the shrink-only allowlist, is an ESLint **error**. Shared/lib/store/queries are not features
     and are allowed.
  2. `features/*/api/*.ts` contains **zero** `Omit<Generated…>` / `Omit<components[…]>` overrides of
     generated types.
- **Source of truth:** `validation-contract.md §F1.4`; `eslint.config.mjs`; `frontend/apps/web/src/features/*`.

## What this feature implements

- Extend `eslint.config.mjs` with a feature-boundary regime (zero new deps): per-feature
  `no-restricted-imports` zones forbidding imports from sibling feature dirs, with an explicit
  shrink-only allowlist grandfathering current edges. Enumerate current cross-feature edges to seed
  the allowlist; record the count.
- Remove the two `Omit<>` overrides in `features/templates/api/templates.ts` (`TemplateDTO`,
  `VersionDTO`), replacing with direct generated-type re-exports. If removal exposes a genuine
  generated-type nullability gap, fix it **spec-side** (openapi.yaml + regen); do NOT retain the
  `Omit<>`. If the spec fix exceeds M1's boundary → HS-2, surface.

## Non-goals (mandatory)

- NOT refactoring the ~50+ existing cross-feature imports (grandfathered).
- NOT adding an ESLint plugin dependency.
- NOT turning on the eigenpal/other rules currently off (ADR 0046 scope unchanged).
- NOT touching `WirePlaceholder` / `TemplateSchemas` (genuine local view types, not generated-type overrides).
- NOT changing any generated file by hand.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| ESLint red on new cross-feature import | temp `src/features/documents/__boundary_probe.ts` importing `../tokens/...` (no allowlist entry) → `npx eslint` / `pnpm run lint` non-zero, boundary error | fixture (deleted) |
| ESLint green on clean tree | `pnpm run lint` (or `npx eslint src/features`) → 0 boundary errors (allowlisted edges pass) | real |
| Zero Omit<> overrides | `grep -rnE "Omit<\s*(Generated|components\[)" frontend/apps/web/src/features/**/api/*.ts` → 0 hits | real |
| Types still compile | `pnpm exec tsc --noEmit` (or build tsc step) → clean | real |
| Allowlist recorded | count + owner (Leandro) + shrink-only trigger in evidence.md | real |

> Known env risk: if `pnpm run lint` is blocked by the pnpm junction drift, demonstrate the gate via
> `npx eslint --config eslint.config.mjs <probe>` and record the block as a bounded defer (trigger:
> complete pnpm install). The gate must still be shown red-on-probe / green-on-clean.

## ADR needed?

- [x] No durable architecture decision — enforces existing FE hard-rule #8 (no cross-feature imports)
  and the contract-first generated-types rule. Skip.
