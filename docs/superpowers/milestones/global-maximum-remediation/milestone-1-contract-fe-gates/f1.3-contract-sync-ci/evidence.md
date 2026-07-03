# Feature F1.3 — contract-sync promoted to blocking CI (reconciled) — Evidence

> **Status:** CLOSED. Checker reconciled to runtime truth → zero live DRIFT across the 4 gated
> modules; wrapper + blocking CI job added; detection power proven preserved (C6). **Contract:**
> `../validation-contract.md §F1.3`. **Implemented by:** subagent (sonnet); **reviewed + verified by:**
> main session (independent wrapper positive + real exit-code negative).

## Reconciliation — stale points fixed (aligned to runtime truth, detection preserved)

| # | Stale point | Runtime truth found | Fix |
|---|-------------|---------------------|-----|
| 1 | templates OpenApi/FrontendTypes patterns (absolute `/api/v1/templates:`) | spec key `/templates:` (openapi.yaml:1105); `index.d.ts:634` `"/templates":` (relative, post-AD-1) | → relative form (all 3 templates paths) |
| 2 | templates RuntimePatterns `generated.ListTemplates` (dead token) | actual mount `templatesapi.HandlerWithOptions` (delivery/http/handler.go:149) | → real mount token |
| 3 | documents RuntimeFile `module.go` (no mount there) | mount `documentsapi.HandlerWithOptions` (delivery/http/handler.go:265) | → repoint RuntimeFile; drop never-present tokens |
| 4 | documents/controlleddocuments/taxonomy OpenApi+FrontendTypes absolute keys | live spec + index.d.ts relative | → relative (verified per line) |
| 5 | documents OpenApiForbiddenPatterns `` "`n  /documents:" `` matches the CORRECT relative key (inverted → permanent drift) | `PATH-BASE-PREFIX` api-lint rule already rejects any `/api/v1`-prefixed key → the legacy hazard can't recur | → retired the forbid (`@()`) |
| 6 | taxonomy FrontendWrapperPatterns `const BASE="/api/v1/taxonomy"` (gone) | `taxonomy.ts` uses `api.GET("/taxonomy/profiles",…)` (openapi-fetch typed client) | → the 3 actual `api.GET(...)` patterns |
| 7 | Test-FrontendGeneratedTypeUsage regex false-flags `operations[` / `NonNullable<…>` / `Alias['prop']` | 8 legit generated-derived aliases in templates/documents/controlledDocuments `.ts` (incl. F1.4's new `components['schemas']['TemplateDTO'|'VersionDTO']` + `VersionDTO['status']`) | → extend allow-regex to recognize them |

**C6 (no symptom-patch):** every change aligns to current correct truth; none silences a real drift.
Proven by the NEGATIVE below (injected drift still caught). Where zero-DRIFT would otherwise require
blinding a real check, it was NOT done — see the surfaced item.

### Architecture contradiction surfaced (not silently patched)
`templates.ts` has `interface TemplateSchemas` / `interface WirePlaceholder` — genuinely local
wire-adapter view types that F1.4's own contract (`validation-contract.md` §F1.4) explicitly blesses
as permanent non-generated types (the generated DTO types `placeholder_schema` as untyped JSON, so a
local view type is structurally required). Outside F1.3's named stale-point list; would otherwise
permanently block zero-DRIFT with no honest regex fix. Resolved via a **finite, named, shrink-only
`FrontendTypeAllowlist`** (idiom: css-token-discipline / test-discipline) listing exactly those two,
owner+trigger commented — **not** a regex loosening, so detection power for everything else is intact
(the negative still catches an injected `interface` with no allowlist entry). Owner: Leandro; trigger:
if the spec ever types placeholders precisely, remove the entries.

## Scope + carve-out

Gated (reconciled to zero DRIFT): **templates, documents, controlleddocuments, taxonomy**.
**`approval` excluded** (`UsesGeneratedBoundary=$false`) — ownership questions entangled with the M9
F9.5 approval-promotion decision (HS-2 if reconciled here). Its config left untouched, un-gated,
recorded in the CI job comment. Owner: Leandro; trigger: M9 F9.5.

## Validation Gate — proof (verified by main session)

| Criterion | Proof | Result |
|-----------|-------|--------|
| Zero DRIFT ×4 | `pwsh -File ./scripts/check-module-contract-sync.ps1 -Module <m>` for each | all exit 0, no `[DRIFT]` |
| Wrapper green clean | `pwsh -File ./scripts/check-contract-sync-all.ps1` | **exit 0**, "zero DRIFT across templates, documents, controlleddocuments, taxonomy" |
| Wrapper red on injected drift | appended `export interface F13…Probe{}` to `taxonomy/api/taxonomy.ts` | **exit 1**, `[DRIFT] taxonomy … handwritten frontend types: interface F13…`; `RESULT: contract-sync FAILED - drifted module(s): taxonomy`; reverted (git clean) |
| Wrapper exit code (CI-critical) | real `$?` on drift vs clean | **1 on drift, 0 on clean** (CI will block) |
| CI job blocking + filtered | `contract-sync` job in `api-contract.yml` | `runs-on: ubuntu-latest`, `shell: pwsh`, `actions/checkout@v4`, **no `continue-on-error`**; path filter covers openapi, internal/modules/**, features/**, api-types/**, both scripts |
| Detection power preserved | the negative proves the reconciled checker still catches real drift | ✓ |

## CI wiring point

New `contract-sync` job in `.github/workflows/api-contract.yml`. `ubuntu-latest` + `shell: pwsh`
(PowerShell Core preinstalled on GitHub ubuntu runners; the checker is pure Get-Content/regex/HashSet,
no Windows-only cmdlets — matches every other job's runner, no Windows-runner cost). Blocking.

## Files changed

- `scripts/check-module-contract-sync.ps1` — config/pattern reconciliation + shrink-only
  `FrontendTypeAllowlist` + FE-type allow-regex extension (logic not weakened).
- `scripts/check-contract-sync-all.ps1` (new) — wrapper over the 4 gated modules; exit 1 if any drift.
- `.github/workflows/api-contract.yml` — new blocking `contract-sync` job + path-filter additions.

## Bounded defers

- `approval` reconciliation/gating → M9 F9.5 (owner Leandro).
- `FrontendTypeAllowlist` {TemplateSchemas, WirePlaceholder} shrink-only (owner Leandro; trigger:
  precise placeholder spec typing).

## Review disposition

7 stale points each justified against runtime truth (cited); zero live DRIFT ×4; wrapper exit-code
proven 1/0; negative proves detection survived; carve-out recorded. The `FrontendTypeAllowlist` is the
correct named-shrink-only idiom, not a detection weakening. **APPROVED** — committed.
