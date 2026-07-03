# Feature F1.4 — ESLint feature-boundary rule + remove `Omit<>` overrides — Evidence

> **Status:** CLOSED. Zero-dep boundary gate live + proven red/green; both `Omit<>` overrides removed
> (0 hits); tsc baseline-clean. **Contract:** `../validation-contract.md §F1.4`. **Implemented by:**
> subagent (sonnet); **reviewed + verified by:** main session (independent red/green re-run).

## Part A — feature-boundary ESLint rule (zero new deps)

Extended the root `eslint.config.mjs` with one `no-restricted-imports` config block per feature dir
(13 features), forbidding imports from every OTHER feature except allowlisted edges. Built on the
stock rule already used for the eigenpal ACL — **no eslint plugin added**.

- **Regex over glob (design decision):** glob `group` patterns can't match relative specifiers
  (`../tokens/...`) — minimatch `**` doesn't cross a leading `../`, and depth varies per file. Used
  `no-restricted-imports` `patterns[].regex` (built-in, ESLint ≥9.3; repo on 10.5.0 → still zero-dep),
  anchored `^((\.\./)+|\./)(features/)?<name>(/|$)` so `documents` never false-matches
  `controlled-documents`. Verified both directions.
- **Flat-config merge hazard fixed:** a later block's `rules['no-restricted-imports']` replaces (not
  merges) the base eigenpal block for files matched by both, so each per-feature block re-injects the
  eigenpal ACL patterns. Verified the eigenpal guard still fires inside `features/**`.
- **No `@` alias** in this repo's vite/tsconfig (imports are 100% relative); alias regex forms
  included defensively so a future alias can't silently reopen the gap.

### Shrink-only allowlist (grandfathering)
Enumerated 2026-07-03 via a full scan of `features/**/*.{ts,tsx}` relative imports resolving into a
different feature dir: **19 distinct (from→to) pairs / 112 import statements**. Encoded as the finite
`ALLOWLIST` array in `eslint.config.mjs`. A NEW (non-allowlisted) cross-feature edge errors; every
existing edge passes. **Owner:** Leandro. **Trigger:** incremental de-coupling — entries removed as
edges are refactored to shared modules, never added.

Pairs: approval→{controlled-documents,documents,taxonomy}; controlled-documents→documents;
dashboard→{approval,documents}; documents→{approval,controlled-documents,iam,taxonomy,templates};
shell→{auth,notifications}; taxonomy→templates; templates→{iam,taxonomy,tokens}; tokens→{iam,templates}.

## Part B — remove `Omit<>` overrides

`features/templates/api/templates.ts`: `TemplateDTO` and `VersionDTO` were `Omit<Generated…, …> & {…}`
hand-retypings (the drift vector the review named). Post-F1.2 the generated types carry correct
required+nullable shape, so both are now `= components['schemas']['TemplateDTO'|'VersionDTO']`
directly. The dead `GeneratedTemplateDTO`/`GeneratedVersionDTO` aliases removed.

One genuine narrowing surfaced (NOT a nullability drift, NOT an `Omit<>` re-typing): the generated
`VersionDTO.placeholder_schema` is `{[key:string]:unknown}[] | null` (spec can't statically type the
placeholder shape). `deriveTemplateSchemas` (templates.ts:258) needs the precise `WirePlaceholder`
element shape for its `.map`. Resolved with a local `as WirePlaceholder[] | null` cast **at that one
call site** — structurally the same blessed local-view-type pattern as `WirePlaceholder`/
`TemplateSchemas` (which the contract explicitly permits). No spec/ADR change; no HS-2/HS-7.

## Validation Gate — proof (verified by main session)

| Criterion | Proof command | Result |
|-----------|---------------|--------|
| ESLint green on clean tree | `pnpm run lint` (root; `eslint .`) | **exit 0**, 0 boundary errors (19 allowlisted edges pass) |
| ESLint red on new cross-feature import | temp `features/documents/__boundary_probe.ts` importing `../../tokens/api/tokensTypes` (documents→tokens not allowlisted) | **exit 1**: `no-restricted-imports` `F1.4: cross-feature import into "documents" from sibling feature "tokens"…` (probe deleted; re-ran → exit 0) |
| Zero `Omit<>` overrides | `grep -rnE "Omit<\s*(Generated\|components\[)" frontend/apps/web/src/features --include="*.ts"` | **0 hits** |
| Types still compile | `npx tsc --noEmit -p .` (frontend/apps/web) | 5 errors, **all pre-existing baseline** (verified via git stash/pop — identical 5 in unrelated test files; 0 attributable to this change) |
| Allowlist recorded | 19 pairs / 112 imports; owner Leandro; shrink-only trigger | in `eslint.config.mjs` + above |

## CI wiring point (no new workflow)

The existing `eslint` job in `.github/workflows/lint.yml` (`pnpm run lint`, reads root
`eslint.config.mjs`); path filter already covers `**/*.ts(x)` + `eslint.config.mjs`. Blocking.

## Bounded defers

- The ~50+ (19-pair/112-stmt) pre-existing cross-feature edges are grandfathered (shrink-only), not
  refactored — out of M1 appetite. Trigger: incremental de-coupling.
- No pnpm-junction-drift block hit this run (`pnpm run lint` + `tsc` ran clean).

## Review disposition

Zero-dep regex construction correct + disambiguates similar feature names; eigenpal guard preserved;
gate proven red/green by independent re-run; `Omit<>` fully removed with the one call-site narrowing
justified. Fixed a stale "25"→"19" allowlist-count comment. **APPROVED** — committed.
