# Lane: frontend

Scope: `frontend/apps/web` (React + TypeScript SPA), audited against its own stated rules in
`wiki/architecture/frontend-structure.md` and against general TS/React practice.

## Findings

| ID | Class | Finding | Evidence | Scale |
|----|-------|---------|----------|-------|
| FE-01 | gap | ESLint config enforces exactly one thing (feature-boundary + eigenpal-ACL import restrictions); `@typescript-eslint` and `eslint-plugin-react-hooks` are registered but **all their rules are off** — no `no-unused-vars`, no `exhaustive-deps`, no `no-floating-promises`, nothing. | `eslint.config.mjs:8-11` (comment states this explicitly), `:180-186` | 1 config file, repo-wide; 0 quality rules active |
| FE-02 | hazard | `check-css-token-discipline.sh` (the only design-token gate) is **currently red**: 3 unlisted raw-hex violations in a file not on its own allowlist. | `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.module.css:214,240,241`; ran `bash scripts/check-css-token-discipline.sh` | 3 violations, 1 file |
| FE-03 | idiom | God-component rule (`wiki/architecture/frontend-structure.md` §15: ">400 lines. Split.") is violated by 4 files. | `DocumentWorkspacePage.tsx` 738 LOC, `DecisionFooter.tsx` 506, `TemplateEditorPage.tsx` 495, `StageCard.tsx` 405 (counted via `wc -l`) | 4 files, up to 738 lines |
| FE-04 | drift | Structure doc mandates every feature owns `index.ts` "barrel: only public API of the feature" (`frontend-structure.md:57`). 11 of 14 features have no `index.ts` at all. | checked `src/features/{approval,auth,controlled-documents,dashboard,documents,feature-flags,iam,notifications,password-change,shared,shell}/index.ts` — none exist; only `taxonomy`, `templates`, `tokens` have one | 11/14 features |
| FE-05 | layering | Cross-feature imports exist and are governed only by a hand-maintained, shrink-only ESLint allowlist (`ALLOWLIST`, 19 from→to pairs), not by structure. Confirmed live: 66 non-test cross-feature relative import statements. | `eslint.config.mjs:84-104`; `grep -rnE "from ['\"](\.\./)+(approval\|auth\|...)/`" over `src/features` | 19 allowlisted edges, 66+ live import statements |
| FE-06 | gap | Design-token discipline check covers **only raw hex literals** in CSS Modules. It does not check `rgb()/rgba()` literals or raw px sizing/spacing values, both of which are common. | `scripts/check-css-token-discipline.sh:76` (`grep -nE '#[0-9a-fA-F]{3,8}'` is the entire check) | rgba/rgb: 82 occurrences in 18 files; raw `px` literals: 2005 occurrences vs 935 `var(--sp-*)` uses across 144 `*.module.css` files |
| FE-07 | duplication | `formatDateTime`/`formatDate` reimplemented locally instead of importing the canonical `lib/formatDate.ts`. | `src/features/documents/components/wizard/steps/StepConfirm.tsx:210`, `src/features/documents/components/wizard/steps/StepTemplate.tsx:192` vs canonical `src/lib/formatDate.ts:14` | 2 duplicate implementations |
| FE-08 | idiom | `tsconfig.json` sets `strict: true` only. No `noUncheckedIndexedAccess`, `noImplicitOverride`, `noFallthroughCasesInSwitch`, `exactOptionalPropertyTypes` — all high-value in a strict TS repo of this size, and cheap because ESLint enforces nothing else (FE-01). | `frontend/apps/web/tsconfig.json:2-24` | repo-wide compiler config |
| FE-09 | gap | Test suite could not be proven green under this audit: parallel run (`vitest run`) crashed with `Error: Worker exited unexpectedly` (tinypool); single-forked rerun did not finish inside 100s and was still emitting DOM dumps from a Testing-Library assertion mismatch in a route-admin dialog test. | ran `npx vitest run --reporter=dot`; ran `npx vitest run --pool=forks --poolOptions.forks.singleFork=true` (both under `frontend/apps/web`) | 2 run attempts, neither reached a summary |
| FE-10 | gap | knip finds 70 unused files under `design-source/` (committed screen exports, expected per policy) plus **13 unused files under `src/`** (real dead code) and 48 unused named/default exports. | ran `npx knip --no-progress` | 13 unreferenced `src/` files (e.g. `src/features/documents/canvas/*` — 8 files, `src/components/WorkspaceDataState.tsx`, `src/features/auth/routes.tsx`, `src/routing/workspaceRoutes.ts`); 48 unused exports |
| FE-11 | idiom | 21 declared `package.json` dependencies (tiptap × 10, uppy × 5, docx, dompurify, fast-xml-parser, react-icons, react-pdf) are unused per knip; `react-pdf`/`react-icons` being flagged is surprising for an app with a PDF viewer and icon system — worth a second look, not taken at face value here. | `npx knip --no-progress` "Unused dependencies" section | 21 deps flagged |

## The five heaviest, with detail

**FE-01 — the lint config enforces nothing but import boundaries.** The comment in the config file says it outright: `@typescript-eslint` and `react-hooks` plugins are wired up only so pre-existing `eslint-disable` comments don't error on an unknown rule, and every actual quality rule is off. This means `react-hooks/exhaustive-deps`, unused-variable detection, floating-promise detection, and every other TS-idiom guard a serious 42.9k-LOC TS codebase would run are silently absent — anything that isn't a type error or a cross-feature import currently ships unchallenged. It's why FE-08 (weak tsconfig) and drift like FE-07 (duplicated formatters) survive: nothing was ever positioned to catch them.

**FE-05 — cross-feature imports are governed by a hand-authored allowlist, not architecture.** The structure doc's hard rule #8 forbids cross-feature imports outright, but the live code has 66+ of them across 19 allowlisted from→to pairs, and the guard that lets them through is a literal array of exceptions someone enumerated by hand in July. It is explicitly labeled shrink-only and dated, which is the right shape for a transitional local maximum — but it means the "no cross-feature imports" rule, as currently enforced, is really "no *new* cross-feature imports beyond this specific list," and nothing distinguishes those two claims when reading the wiki alone.

**FE-09 — the test suite's actual health is unverified, and that itself is informative.** One run mode (default multi-worker) crashes a worker process outright; a slower single-forked mode didn't finish in 100 seconds and was mid-dump on a DOM-snapshot assertion mismatch in a route-admin dialog test when time ran out. Neither run produced a pass/fail summary. This blocks any confident claim about "N tests, M passing" for this lane, and it corroborates the known FE node_modules junction-drift issue's *shape* (flaky/mode-dependent vitest execution) even though this audit could not pin the exact cause.

**FE-06/FE-02 — the token-discipline gate exists, is narrow, and is currently failing.** The only automated design-token check greps for `#hex` literals in CSS Modules and nothing else — not `rgb()/rgba()`, not raw pixel sizing (2005 raw-px occurrences vs. 935 uses of the spacing token `var(--sp-*)`, a roughly 2:1 ratio of hardcoded to tokenized sizing). And even within its narrow scope, the gate is red right now: `DocumentWorkspacePage.module.css` — which is also the 738-line largest file in the codebase (FE-03) — has three raw-hex literals that are on nobody's allowlist.

**FE-03/FE-04 — the two structural rules with the cleanest enforcement gap.** The 400-line god-component ceiling and the "every feature has a barrel `index.ts`" rule are both explicit, both easy to check mechanically, and both currently violated at real scale (4 files over the line; 11 of 14 features missing the barrel). Neither has any CI teeth — they exist only as prose in `frontend-structure.md`.

## What is actually fine

- **No import cycles.** `npx madge --circular` over `src` (709 files) returns clean.
- **Zero cross-feature imports outside the governed allowlist** — the ESLint boundary rule genuinely fires (it's the one rule that's on) and the allowlist is explicitly shrink-only and attributed, which is the right shape for a labeled transitional exception.
- **Query-key discipline is strong**: 103 call sites use `QK.*` centralized constants vs. 1 inline `queryKey: [...]` array (`src/features/iam/tabs/PeopleDetailDrawer.tsx:173`).
- **Zustand discipline holds exactly as documented**: only `store/auth.store.ts` and `store/ui.store.ts` call `create(`; no feature-scoped stores exist.
- **`any`/non-null-assertion usage is genuinely rare**: 6 `: any` sites (all `catch (e: any)` or one localStorage cast), 0 postfix non-null assertions (`x!.y`) in production code, only 2 `@ts-expect-error` (both for an untyped `uuid` import, with a stated reason).
- **`tsc --noEmit` is clean** — the codebase currently type-checks with no errors under `strict: true`.
- **API-type aliasing is mostly done right where it matters**: `approvalTypes.ts` and most of `documents.ts`'s exported types are direct aliases of `components['schemas'][...]` rather than hand-written shapes; the apparent "hand-written Response type" files mostly turned out, on inspection, to be either generated-type aliases or legitimate app-internal domain models (e.g. `schemaRuntimeTypes.ts` describes the dynamic document-editor's runtime schema, not an API contract).
- **`fetch()` usage outside `lib/api/` is narrow and already documented as sanctioned**: 7 sites, all presigned blob-storage uploads/downloads (S3/MinIO, different origin) or the ETag mutation client, which `frontend-structure.md §7` names explicitly as transport tier 2/3.
- **`role-vocabulary.ts`** — per the shared brief, already a known-good pattern; not re-reported here.

## Unverified / needs judgment

- Does the vitest crash (`Worker exited unexpectedly`) reproduce the known "FE node_modules junction drift breaks vitest" memory entry, or is it a distinct, newer failure (e.g. Node 26 + tinypool compatibility, or a genuine snapshot/assertion bug in the route-admin dialog test)? Could not isolate the root cause within this lane's budget.
- Is the knip "unused dependency" list for `react-pdf`/`react-icons` a true positive (config-driven dynamic import knip can't trace) or a real removal candidate? Worth a targeted check before acting on it.
- Are the 13 `src/`-rooted knip-flagged files (`documents/canvas/*`, `documents/runtime/*` field components, `WorkspaceDataState.tsx`, `auth/routes.tsx`, `routing/workspaceRoutes.ts`) genuinely dead, or reachable via a dynamic/lazy path knip's static analysis misses (e.g. `React.lazy` string-built paths)? Several are full sub-trees (a whole `canvas/` folder), which reads more like an abandoned parallel implementation than incidental dead exports — worth a closer look by someone who knows the document-editor history.
- FE-05's cross-feature allowlist is dated "2026-07-03" in its own comment; is it still accurate, or has it silently grown stale relative to the current import graph (i.e., are all 66 live cross-feature imports actually covered by the 19 pairs, or would a fresh `eslint` run surface new violations)? Not run in this audit — recommend running `pnpm lint` (or equivalent) directly rather than inferring from grep.

## Commands run

```
find src -type f | wc -l
find src -type f \( -name "*.ts" -o -name "*.tsx" \) -exec cat {} + | wc -l
find src -type f \( -name "*.ts" -o -name "*.tsx" \) ! -name "*.test.*" -exec wc -l {} + | sort -rn | head -25
find src -type f -name "*.tsx" ! -name "*.test.*" -exec wc -l {} + | awk '$1>400 {print}'
grep -rE "from ['\"](\.\./){3,}" src --include="*.ts" --include="*.tsx" | wc -l
grep -rnE "from ['\"](\.\./)+(approval|auth|controlled-documents|dashboard|documents|feature-flags|iam|notifications|password-change|shell|taxonomy|templates|tokens)/" src/features --include="*.ts" --include="*.tsx"
grep -rn "fetch(" src --include="*.ts" --include="*.tsx" | grep -v "src/lib/api" | grep -v "apiFetch"
grep -rn ": any\b\|<any>\|as any\b" src --include="*.ts" --include="*.tsx"
grep -rn "@ts-ignore\|@ts-expect-error" src --include="*.ts" --include="*.tsx"
grep -rn "eslint-disable" src --include="*.ts" --include="*.tsx"
npx --yes madge --circular --extensions ts,tsx src
npx --yes knip --no-progress
npx vitest run --reporter=dot
npx vitest run --reporter=basic --pool=forks --poolOptions.forks.singleFork=true
npx tsc --noEmit -p tsconfig.json
bash scripts/check-css-token-discipline.sh   (run from repo root)
grep -roE "[0-9]+px" src --include="*.module.css" | wc -l
grep -roE "var\(--sp-" src --include="*.module.css" | wc -l
grep -rn "QK\." src/features --include="*.ts" --include="*.tsx" | wc -l
```
