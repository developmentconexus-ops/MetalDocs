# Pass 11 — Frontend Architecture Map (`frontend/apps/web`)

**Date:** 2026-08-09
**Baseline:** `main@418070bf`
**Status:** reproduced-current (every claim below re-derived from the working tree at this commit; no claim taken on faith)

---

## 1. Feature inventory (`src/features/`)

14 directories under `src/features/`. Size = non-test `.ts`/`.tsx` files and LOC (tests excluded to isolate production surface).

| Feature | Files | LOC | Subdirs present | Purpose |
|---|---:|---:|---|---|
| `documents` | 136 | 15,460 | adapters, api, canvas†, components, hooks, lib, pages, queries, runtime†, state | Document lifecycle: library, wizard, workspace/editor, distribution, comments. Largest feature by a wide margin — includes two fully dead subtrees (†, see §7). |
| `approval` | 82 | 9,919 | api, components, hooks, lib, pages, queries | Signoff routing, inbox, route admin, delegations, SLA. |
| `iam` | 82 | 7,674 | components, mutations, pages, presenters, queries, tabs | Users/roles/groups/process-areas admin, audit events. |
| `templates` | 69 | 7,605 | adapters, api, components, hooks, lib, pages, queries, state, tokens | Template authoring/editor, wizard, review canvas. |
| `shared` | 30 | 3,573 | components, controlled-artifact | Cross-feature view-model layer (document/template artifact shell) — lives under `features/shared`, not `src/shared`. |
| `tokens` | 18 | 738 | api, components, pages, queries | Placeholder-token catalog admin. |
| `notifications` | 17 | 975 | api, components, lib, queries | Notification bell/list. |
| `taxonomy` | 21 | 2,165 | api, components, pages, queries | Process-area/profile taxonomy admin. |
| `dashboard` | 11 | 635 | api, lib, pages, queries | Home dashboard stats/widgets. |
| `auth` | 8 | 633 | api, pages | Login. |
| `shell` | 7 | 620 | components, pages | App chrome (nav, layout). |
| `controlled-documents` | 7 | 317 | api, queries, `__tests__` | Controlled-document profile/preview surface. |
| `feature-flags` | 2 | 56 | `__tests__` | Feature-flag read hook. |
| `password-change` | 2 | 13 | pages | Password-change page. |

Total: 401 non-test source files, 47,551 LOC in `src/` overall (incl. the 8,362-line generated `lib/api-types/index.d.ts`).

---

## 2. Cross-feature imports

The repo's own gate (`eslint.config.mjs`, F1.4 feature-boundary guard) enumerates a **shrink-only allowlist of 19 `(from, to)` pairs**. I re-derived the actual edges by grepping relative-import paths (the repo has no `@/features` alias, confirmed in the eslint config comment, so relative-path grep is exhaustive):

| Edge | Files | In allowlist? |
|---|---:|---|
| documents → shared | 15 | not applicable — `shared` is exempt (see below) |
| templates → shared | 14 | exempt |
| documents → taxonomy | 12 | yes |
| documents → controlled-documents | 10 | yes |
| documents → approval | 9 | yes |
| dashboard → documents | 4 | yes |
| approval → taxonomy | 4 | yes |
| templates → taxonomy | 4 | yes |
| templates → tokens | 4 | yes |
| documents → templates | 5 | yes |
| controlled-documents → documents | 3 | yes |
| dashboard → approval | 3 | yes |
| shell → auth | 3 | yes |
| approval → controlled-documents | 2 | yes |
| approval → shared | 2 | exempt |
| approval → documents | 1 | yes |
| auth → shared | 1 | exempt |
| shell → notifications | 1 | yes |
| shell → shared | 1 | exempt |
| taxonomy → templates | 1 | yes |
| tokens → templates | 1 | yes |

**Finding — stale allowlist entries.** Three allowlisted pairs have **zero live call sites**: `documents → iam`, `templates → iam`, `tokens → iam`. The allowlist comment says "entries are only ever removed... never added" but these three have not been pruned as the edges were refactored away. Not a runtime defect — a governance-hygiene gap in a self-described shrink-only list. Maps to **#91/A2** (quality ratchet not actually ratcheting).

**Finding — `shared` is a feature-shaped exception, undocumented as such.** `features/shared` sits physically inside `src/features/`, but the ESLint `FEATURE_NAMES` list (13 entries) omits it, so it is treated as unrestricted cross-cutting code — any feature may import it freely, and it may import any feature (`approval`, `documents`, `templates` types are visible to it via adapters). The eslint config's own comment describes the exempt bucket as "`src/shared`, `src/lib`, `src/store`," etc. — none of which is the actual path. The exemption is correct in effect (this is a legitimate view-model boundary, see §3) but the doc comment names the wrong directory, which will mislead the next person editing the guard.

No unhealthy "deep reach into a foreign feature's guts" edges were found beyond what's already allowlisted — every cross-feature edge above targets a feature's `api/`, `queries/`, or `types` surface, not private page/component internals.

---

## 3. Public feature surfaces (barrels vs. deep reach)

Only **3 of 14** features export a barrel (`index.ts`): `taxonomy`, `templates`, `tokens`. The other 11 (including the two largest, `documents` and `approval`) have no barrel — every cross-feature import necessarily reaches into a subpath.

More importantly, **even the 3 barrel-having features are not consumed through their barrel** by cross-feature callers:
- `documents → taxonomy` imports `../taxonomy/queries/useProfilesQuery`, `../taxonomy/api/taxonomy`, `../taxonomy/types` directly — never `../taxonomy` (the barrel).
- `templates → tokens` imports `../tokens/api/tokensTypes`, `../tokens/queries/useTokensQuery`, `../tokens/useTokenCatalog` directly.
- `documents → templates` is the one exception that *does* import the bare `../../templates` barrel in 2 files.

**Verdict:** the barrel pattern exists in 3 features but is dead weight — consumers reach into internals regardless of whether a barrel is present. There is no enforced "public surface" concept for features; the F1.4 ESLint guard restricts *which* feature may be imported, not *what part* of it. This is a real architectural gap relative to a proper module-boundary system (compare the backend's application-service/published-interface discipline in `internal/modules/`) but is scoped to this pass as an observation, not a fix.

---

## 4. Shared/platform layout

```
src/
  app/        AppRouter.tsx, RootProviders.tsx (QueryClientProvider, staleTime 30s/retry 1), bootstrap.tsx
  components/ WorkspaceDataState.tsx (dead, see §7), WorkspaceViewFrame.tsx, ui/ (25 files: Dialog, Stepper, SelectMenu, StatusPill, TabBar, Avatar, etc.)
  lib/        api/ (client.ts, errors, problem+json, authBus), api-types/ (generated), format/, hooks/, iam/, inbox/, labels/, observability/, types/, queryKeys.ts
  routing/    workspaceRoutes.ts (dead, see §7)
  store/      auth.store.ts, ui.store.ts (2 Zustand stores total)
  styles/     tokens.css + global styles.css
```

**Inversion check (shared → feature imports):** clean, with one exception. `lib/inbox/sortByDue.ts` imports `InboxItem` from `../../features/approval/api/approvalTypes` and is consumed only by `features/approval/pages/InboxPage.tsx`. This is feature-specific code misplaced under `lib/`, not genuine shared code — a single, small instance of the inversion the F1.4 guard doesn't catch (it only polices feature→feature, not lib→feature). `components/`, `store/`, `routing/` have zero feature imports. `app/AppRouter.tsx` importing features is the composition root wiring pages to routes — expected, not an inversion.

---

## 5. API client architecture

**Generated types:** `src/lib/api-types/index.d.ts`, 8,362 lines, produced by `openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts` (`gen:api` script in `package.json`). Imported directly by 43 files.

**Transport tiering** (`lib/api/client.ts`, self-documented in-file as "TRANSPORT HIERARCHY (FE-13)"):
1. Generated `api` (openapi-fetch client typed off `api-types`) — canonical, compile-time drift detection. **62 files** use it.
2. `apiFetch` — sanctioned escape hatch for ETag/non-contracted routes.
3. `request`/`requestRaw`/`requestBlob` — explicitly labeled "legacy thin wrappers... Shrink-only: do not add new callers." **1 file** (`features/approval/api/approvalApi.ts`, one `requestRaw` call at line 64) still uses it.

This is a textbook example of CLAUDE.md's "labelled transitional local maximum" — the legacy tier is named, bounded, and nearly drained (1 call site left). No action needed; noted as healthy, not a defect.

**Hand-written type files (9 total, non-test):**

| File | LOC | Imports generated types? | Own declarations | Verdict |
|---|---:|---|---:|---|
| `features/documents/runtime/schemaRuntimeTypes.ts` | 335 | no | 25 | Dead (part of the runtime/canvas dead cluster, §7) — not a live duplication |
| `features/shared/controlled-artifact/types.ts` | 443 | no | 22 | Legitimate view-model boundary, explicitly documented as "kind-agnostic... adapters map API shapes into these types" — not a DTO duplication |
| `features/taxonomy/types.ts` | 121 | yes (1) | 11 | Mixed — mostly derives from generated types |
| `features/approval/api/approvalTypes.ts` | 79 | yes (2) | 15 | Mixed |
| `features/templates/placeholder-types.ts` | 62 | no | 7 | UI-only placeholder-editor concepts, not wire DTOs |
| `features/documents/canvas/templateTypes.ts` | 12 | no | 3 | Dead (canvas cluster) |
| `features/controlled-documents/types.ts` | 15 | yes (1) | 2 | Mostly derived |
| `features/iam/types.ts` | 5 | no | 1 | Trivial |
| `features/tokens/api/tokensTypes.ts` | 9 | yes (1) | 4 | Mostly derived |

**No hand-written DTO file duplicates a live generated wire shape wholesale.** The two largest hand-rolled type files are either dead code or a deliberate, documented adapter-boundary pattern (view-model, not wire DTO). This part of the original claim frame ("hand-written client DTOs duplicating generated types") does not reproduce as stated — the actual gap is elsewhere (dead code volume, see §7), not DTO duplication. Relevant to **#90/A3** (generated-client adoption): adoption is in fact strong (62 canonical vs. 1 legacy call site; 43 files touch generated types directly).

**Legacy wrappers:** only `lib/legacyStatus.ts` matches a "legacy" naming pattern outside the client tiering already covered.

---

## 6. TanStack Query discipline

**Query keys:** centralized in `lib/queryKeys.ts` (`QK` namespace object, one entry per domain/sub-resource), with an explicit file-header rule: *"All useQuery / invalidateQueries calls must import from here - never inline string arrays."* 74 files import `QK`. Only **one violation** found: `features/iam/tabs/PeopleDetailDrawer.tsx:173` inlines `queryKey: ["iam", "admin", "users"]` instead of using `QK`. Otherwise fully honored — this is a well-enforced convention despite having no lint rule backing it (enforcement is discipline/review only).

**Invalidation pattern:** 45 `invalidateQueries` call sites, dominant pattern is domain-broad invalidation (`QK.approval.all`, `QK.documents.all`) rather than surgical per-key invalidation — a consistent, simplicity-over-precision choice across mutations, not an inconsistency.

**QueryClient config:** single instance in `app/RootProviders.tsx`, `staleTime: 30_000`, `retry: 1` — one config, no per-feature divergence found.

**Local vs. server state:** clean separation. Only 2 global Zustand stores exist (`store/auth.store.ts` — identity/session/login-form; `store/ui.store.ts` — global message/error strings), neither caches server data. Feature-local multi-step state (document wizard, template wizard) uses `useReducer` (`features/documents/state/wizard.reducer.ts`, `features/templates/state/templateWizard.reducer.ts`), not Zustand. **No Zustand/TanStack Query overlap found** — the claim frame's implicit worry does not reproduce.

---

## 7. Dead code (reproduced count exceeds the claim)

Method: for every non-test `.ts`/`.tsx` file, grep the whole `src/` tree for any reference to its basename from another file; then manually re-check every "isolated" hit for transitive dead clusters (files only referenced by other already-dead files). Cross-checked against `git grep` for full import-path strings, not just basenames.

**Cluster A — `features/documents/canvas/` + `features/documents/runtime/`: 19 files, fully disconnected from the live app.**
Nothing outside this pair of directories imports anything from either of them; `canvas/` imports `runtime/`, so the two form one closed, unreachable subtree.

- `canvas/`: `DocumentCanvas.tsx`, `DocumentCanvas.module.css`, `TemplateNodeRenderer.tsx`, `slotBindings.ts`, `slotValues.ts`, `templateAdapters.ts`, `templateTypes.ts`, `rich/metaldocsRich.ts`, `slots/FieldSlot.tsx`, `slots/RichSlot.tsx` (10 files)
- `runtime/`: `DynamicEditor.tsx`, `DynamicEditor.module.css`, `schemaRuntimeAdapters.ts`, `schemaRuntimeTypes.ts`, `RichField.module.css`, `fields/RepeatField.tsx`, `fields/RichField.tsx`, `fields/ScalarField.tsx`, `fields/TableField.tsx` (9 files)

**Cluster B — standalone dead files (no cross-reference to each other): 10 files.**

| File | Notes |
|---|---|
| `components/WorkspaceDataState.tsx` | zero importers anywhere |
| `features/approval/queries/useDelegationsMutations.ts` | zero importers; its internal `invalidateQueries` calls are dead code calling dead code |
| `features/documents/components/CheckpointsDialog.tsx` | zero importers |
| `features/documents/components/styles/CheckpointsDialog.module.css` | co-located with the above |
| `features/documents/components/Pagination.tsx` | zero importers |
| `features/documents/components/Pagination.module.css` | co-located |
| `features/documents/components/PDFCell.tsx` | zero importers |
| `features/iam/queries/useOverviewQuery.ts` | zero importers |
| `features/shared/documentDisplay.ts` | zero importers |
| `routing/workspaceRoutes.ts` | zero importers; pure path-parsing helpers, no dynamic/file-based routing magic that would make this reachable implicitly |

**Total reproduced: 29 dead files** (19 + 10), versus the claimed 13. The claim significantly **undercounts** — it correctly identified the canvas subtree as abandoned but missed that `documents/runtime/` is entirely a dependency of that same dead subtree (not independently alive), and missed all 10 Cluster-B files.

**Unused exports ("48" claim):** NOT independently verified. `frontend/apps/web/node_modules` is not installed in this worktree (0 entries), and the repo has no `knip`/`ts-prune`/`depcheck` script wired in `package.json` to reproduce an export-level count without a fresh install. The file-level dead-code count above (29) is the reproducible metric for this pass; the exports claim is left open rather than asserted or fabricated.

Maps to **#91/A2** (no ratcheted whole-repo quality baseline — dead code of this volume is exactly the kind of regression a ratchet would catch and this repo currently has none for the frontend).

---

## 8. Component hotspots (files > 400 lines)

Recount, non-test, non-generated `.ts`/`.tsx` files only:

| File | LOC |
|---|---:|
| `features/documents/pages/DocumentWorkspacePage.tsx` | 738 |
| `features/approval/components/sidebar/DecisionFooter.tsx` | 506 |
| `features/templates/pages/TemplateEditorPage.tsx` | 495 |
| `features/shared/controlled-artifact/types.ts` | 443 |
| `features/approval/pages/route-admin/StageCard.tsx` | 404 |

**Reproduced: 5 files exceed 400 lines, max 738** — the claim ("4 exceed, max 738") is off by one file; the max line count matches exactly. The delta is `features/shared/controlled-artifact/types.ts` (443 LOC), which is a type-declaration file, not a component — if the original claim scoped "component hotspots" to `.tsx` component files only, excluding one `.ts` types file, the count of 4 *components* is reproducible; the count of 4 *files* over 400 lines is not (it's 5).

Per the task framing: **LOC is a hotspot signal only, not decomposition proof.** `DocumentWorkspacePage.tsx` and `TemplateEditorPage.tsx` are page-assembly components (top-of-page composition, naturally larger); `DecisionFooter.tsx` and `StageCard.tsx` are approval-domain components with dense conditional-state UI. None of these should be recommended for a LOC-driven split — line count alone says nothing about whether their internal responsibilities are already well-separated.

---

## 9. Lint/type enforcement state

**ESLint** (`eslint.config.mjs`, root-level, applies repo-wide including `frontend/apps/web`): confirmed exactly as claimed.
- `@typescript-eslint` and `eslint-plugin-react-hooks` are *registered* (so pre-existing inline `eslint-disable` directives resolve) but **every rule from both plugins is OFF** — the file's own comment says so explicitly: *"Turning those rules on is a separate, deliberate future decision."*
- The only rules actually enforced are `no-restricted-imports` instances: the Eigenpal ACL boundary (ADR 0046) and the F1.4 feature-boundary guard (§2 above). Confirmed: **import-boundary rules only**, nothing else.

**tsconfig** (`frontend/apps/web/tsconfig.json`): `"strict": true` only. No `noUncheckedIndexedAccess`, no `noImplicitOverride`, no `exactOptionalPropertyTypes`, no other opt-in strictness flags. Confirmed: **strict-only, no additional hardening flags**.

**CSS token gate** (`scripts/check-css-token-discipline.sh`, FE-18): checks all `*.module.css` under `frontend/apps/web/src` for raw hex literals, exempting 27 grandfathered files on a shrink-only allowlist.

Ran it against this baseline:
```
$ bash scripts/check-css-token-discipline.sh
css-tokens: clean (144 module.css files checked, 27 grandfathered)
EXIT=0
```

**This does NOT reproduce.** The claim states "currently red, 3 unallowlisted hex" — at `main@418070bf` the gate is **green** (exit 0, zero violations). Either the claim is stale (fixed since it was filed and not yet retired from whatever tracked it) or it describes a different baseline/branch. Reporting as **claim not reproduced — gate is currently passing**, not carrying it forward as a live defect.

---

## 10. Claimed vs. reproduced — summary table

| # | Claim | Reproduced? | Actual finding |
|---|---|---|---|
| 1 | 13 dead `src/` files incl. abandoned `documents/canvas/` subtree | **Undercount** | 29 dead files: `canvas/`(10) + `runtime/`(9, entirely a dependency of the dead canvas cluster, not independently live) + 10 standalone (incl. 2 co-located `.module.css`) |
| 2 | 48 unused exports | **Not verified** | No `knip`/`ts-prune`/`depcheck` tooling wired; `node_modules` not installed in this worktree; left open rather than asserted |
| 3 | 4 files exceed 400 lines, max 738 | **Off by one** | 5 files exceed 400 (max 738 confirmed exact); the 5th (`shared/controlled-artifact/types.ts`, 443 LOC) is a types file, not a `.tsx` component — count of 4 holds if scoped to components only |
| 4 | @typescript-eslint + react-hooks registered but none enabled; import-boundary rules only | **Confirmed exact** | Verbatim match to `eslint.config.mjs` comments and rule config |
| 5 | tsconfig strict only | **Confirmed exact** | `"strict": true`, no other strictness flags |
| 6 | CSS token gate currently red, 3 unallowlisted hex | **Not reproduced** | Gate is green at this baseline: 144 files checked, 27 grandfathered, 0 violations, exit 0 |
| 7 | Hand-written client DTOs duplicating generated types | **Not reproduced as framed** | No live file duplicates a generated wire DTO wholesale; the two largest hand-rolled type files are either dead code (`schemaRuntimeTypes.ts`, in Cluster A) or a documented, deliberate view-model adapter boundary (`shared/controlled-artifact/types.ts`) |
| 8 | Legacy API wrappers | **Confirmed, and nearly drained** | 3-tier transport is explicitly labeled/bounded (FE-13); only 1 live call site (`approvalApi.ts:64`) still uses the legacy tier vs. 62 files on the canonical generated client |

**New findings not in the original claim frame** (surfaced by this pass):
- 3 stale entries in the F1.4 cross-feature-import allowlist (`documents/templates/tokens → iam`) with zero live call sites — the shrink-only list hasn't shrunk.
- `features/shared` is exempted from the feature-boundary guard but the guard's own comment misnames its actual path (says `src/shared`, is really `src/features/shared`).
- One shared→feature import inversion: `lib/inbox/sortByDue.ts` imports feature-specific types from `features/approval` and is only consumed by that feature — misplaced, not general-purpose.
- One query-key discipline violation: `features/iam/tabs/PeopleDetailDrawer.tsx:173` inlines a query key instead of using the centralized `QK` object.
- Zero Zustand/TanStack Query overlap — local vs. server state discipline is clean (2 minimal global stores, feature-local `useReducer` for multi-step wizards).
- 11 of 14 features have no barrel export at all, and the 3 that do (`taxonomy`, `templates`, `tokens`) are bypassed by cross-feature callers anyway — "public feature surface" is not an enforced concept in this codebase; only *which* feature may be imported is policed (F1.4), not *what part* of it.

---

## 11. Map to program findings

- **#91/A2** (no ratcheted whole-repo quality baseline): dead-code volume (29 files, likely more once export-level tooling is run), the stale allowlist entries, and the CSS-gate claim/reality gap are all symptoms of the same root — nothing in CI currently ratchets these frontend metrics forward. Acceptance property should be "count only ever shrinks, verified by a wired tool," not a specific number.
- **#90/A3** (API contract stops before runtime behaviour / generated-client adoption): adoption is actually strong on the frontend side (62 canonical vs. 1 legacy call site, 43 files on generated types directly) — this pass found no frontend-side gap here worth adding to #90/A3's scope.
- **Frontend-local debt** (not filed against either): the barrel/public-surface gap (§3), the `shared` naming/scoping drift in `eslint.config.mjs` (§2), the `lib/inbox` inversion (§4), and the `PeopleDetailDrawer` query-key violation (§6) are small, contained items with no existing home — worth a standalone frontend-hygiene entry rather than folding into #91/A2's dead-code scope.
