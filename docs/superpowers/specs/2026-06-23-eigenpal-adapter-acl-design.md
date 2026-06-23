# Eigenpal Anti-Corruption Layer — Design Spec

> **Date:** 2026-06-23
> **Version:** v2 (evidence-grounded; supersedes v1 committed `7e2041f9`)
> **Branch context:** discovered on `eigenpal-review-cockpit` during the eigenpal 1.9 migration
> **Status:** Design under written-spec review — every claim below is anchored to a verified
> `file:line` fact (see §1 Evidence base). Implementation plan to follow after sign-off.

---

## 0. Why v2

v1 was directionally right (one adapter, vendor ban, pnpm) but made three claims the research
could not support, and the implementation plan would have drifted on them:

1. **It proposed full OOXML domain types** (`DocumentBlock`/`DocumentParagraph`/`DocumentTable`).
   The evidence shows those eigenpal types are deeply recursive and **never manipulated by live
   code** — mapping them is a large YAGNI violation. v2: the document model stays **opaque**
   (bytes cross the seam, not nodes).
2. **It proposed a `PlaceholderCodec` port** wrapping `placeholderToRun`/`runToPlaceholder`.
   The evidence shows the entire file holding those functions is **dead in production** (test/
   spike-only callers). v2: **delete it**, do not build a port around it. (Reviewer C2.)
3. **It treated "two package managers" as the whole problem.** The real state is npm root + an
   empty committed pnpm stub + 3 disjoint pnpm islands + a React version skew + a malformed
   workspace file + 2 CI workflows already pointing at nonexistent lockfiles. v2 sequences the
   migration around those concrete landmines. (Reviewer C3.)

v2 also folds in what v1 omitted entirely: an **error-translation taxonomy** (reviewer C1) and an
**observability** position (REQ-OBS-3 / RF-1).

---

## 1. Evidence base

Every architectural claim in this spec traces to one of these verified facts. Line numbers are as
of 2026-06-23 on `eigenpal-review-cockpit`.

### 1.1 Phantom dependency (confirmed)
- No `package.json` in the repo declares `@eigenpal/docx-editor-core` as a direct dependency.
  Only `@eigenpal/docx-editor-react` is declared (in `packages/editor-ui/package.json` and
  `apps/docx-renderer/package.json`).
- `@eigenpal/docx-editor-react@1.9.0` lists `@eigenpal/docx-editor-core: ^1.9.0` in its own deps —
  so `core` resolves **transitively only**.
- Direct `core` imports in production code:
  - `packages/editor-ui/src/types.ts:2` — `@eigenpal/docx-editor-core/types/content` (`Comment`)
  - `packages/editor-ui/src/types.ts:6` — `@eigenpal/docx-editor-core/types/document` (`BlockContent`, `Paragraph`, `Table`)
  - `apps/docx-renderer/src/render/fanout.ts:2` — `@eigenpal/docx-editor-core/headless` (`processTemplateDetailed`)

### 1.2 The `any` masks a real defect (new — strongest evidence for the refactor)
- `frontend/apps/web/src/editor-adapters/eigenpal-template-mode.ts:29` — `const props: any = { … }`.
- Line 23 of that object sets `sdtType: "dropdown"`.
- eigenpal's `SdtType` union (`@eigenpal/docx-editor-core` dist, `content-B8ScSBzC.d.ts:742`) is
  `'richText' | 'plainText' | 'dropDownList' | 'comboBox' | 'date' | 'picture' | 'checkbox' | …` —
  **there is no `"dropdown"` member**. The canonical value is `"dropDownList"`.
- Conclusion: the `any` is not merely "untyped" — it **suppresses a compile error that would catch
  wrong output**. Replacing `any` with the real type (`SdtProperties`, same dist file `:768`)
  surfaces the bug. This is the concrete proof that the escape hatch is harmful, not cosmetic.

### 1.3 The placeholder-codec file is dead in production (new — reframes the whole codec port)
- `eigenpal-template-mode.ts` exports `placeholderToRun`, `runToPlaceholder`, `wrapFrozenContent`.
- Repo-wide search for importers: the **only** callers are
  `frontend/apps/web/src/editor-adapters/eigenpal-template-mode.test.ts`,
  two sibling test files, and `…/__spike__/eigenpal-placeholder-spike.test.ts`.
- **Zero** feature/runtime modules import any of the three functions.
- Live placeholder authoring at runtime goes through eigenpal's `templatePlugin` assembled inside
  `MetalDocsEditor` (`packages/editor-ui/src/MetalDocsEditor.tsx:72-76`), **not** through these
  hand-written mappers.
- The branch is `eigenpal-review-cockpit` (active WIP) — so "dead today" must be confirmed against
  intent before deletion (see §6 decision D-1).

### 1.4 Document-model types are deep recursive OOXML (new — reframes domain types)
- `Table` → `rows: TableRow[]` → `cells: TableCell[]` → `content: (Paragraph | Table)[]` (mutually
  recursive). `Paragraph.content: ParagraphContent[]` is an **18-member union**.
- No live module reads or constructs `Paragraph`/`Table`/`BlockContent` as structured objects; the
  re-exports at `editor-ui/src/types.ts:6` exist only to feed the dead codec file (§1.3).
- Conclusion: faithfully mapping these to MetalDocs types is expensive and buys nothing. Keep them
  off the boundary entirely.

### 1.5 `processTemplateDetailed` error + result contract (reviewer C1 evidence)
- Return shape (dist `headless` typings): `{ buffer: ArrayBuffer; replacedVariables: string[];
  unreplacedVariables: string[]; warnings: TemplateError[] }`.
- `TemplateError` is an **interface, not a class**:
  `{ message: string; variable?: string; type: 'parse' | 'render' | 'undefined' | 'unknown'; originalError?: unknown }`.
  → Translation must narrow on `.type` / `.message`, **never** `instanceof TemplateError`.
- Real error **classes** exist for the editor-handle path (e.g. `ContentControlNotFoundError`,
  `ContentControlLockedError`) — those *can* use `instanceof`.
- `nullGetter: 'empty'` is the option in use (`fanout.ts:60`); unresolved vars surface in
  `unreplacedVariables`, not as throws.

### 1.6 The render error path is currently unclassified (reviewer C1 evidence)
- `apps/docx-renderer/src/render/fanout.ts:54-61` calls `processTemplateDetailed` with **no
  try/catch** — any throw propagates to Fastify's default handler → HTTP 500, opaque body.
- `apps/docx-renderer/src/routes/fanout.ts` only catches Zod validation (→ 400); engine failures
  fall through.
- Go caller `internal/.../docxrenderer/client.go:58-61` collapses any non-200 to
  `fmt.Errorf("fanout status %d: %s", …)` — **no error classification, no retry signal**.
- Conclusion: the failure path is the half of the ACL that doesn't exist yet. Completing it is C1.

### 1.7 Build / package-manager ground truth (reviewer C3 evidence)
- Root `package.json` → `workspaces: ["packages/*", "apps/docx-renderer"]` (**npm**). Tracked
  `package-lock.json` is ~258 KB and real.
- Root `pnpm-lock.yaml` is a committed **9-line empty stub**; there is **no** root
  `pnpm-workspace.yaml`.
- `frontend/apps/web` is **not** in the root workspaces; it has its own tracked
  `pnpm-lock.yaml` (React **18.3.1**).
- Untracked, on this branch: `packages/editor-ui/pnpm-lock.yaml` (React **18.2.0**) and
  `packages/editor-ui/pnpm-workspace.yaml` that is **malformed** (contains literal
  `allowBuilds: esbuild: set this to true or false`).
- → **One npm root + three disjoint pnpm islands + a React 18.2.0/18.3.1 skew.**
- All `@metaldocs/*` packages are consumed as **raw `.ts` source** via Vite `resolve.alias` +
  `tsconfig` paths; there is **no `dist`** and `editor-ui`'s `build` script is `tsc --noEmit`.
- `apps/docx-renderer/build.mjs` (esbuild) **externalizes every bare specifier except
  `@metaldocs/*`**, which it inlines; it emits **no metafile** (so no bundle-size gate exists).
- CI: `.github/workflows/ci.yml` uses `npm ci --include-workspace-root`; the e2e workflow uses
  `pnpm`. Two workflows reference lockfiles that **do not exist** at the stated paths
  (`frontend/apps/web/package-lock.json`; `frontend/pnpm-lock.yaml`).

### 1.7a No JS linter exists (decides §9 enforcement)
- Repo-wide there is **no** ESLint / Biome / oxlint dependency, config, or `lint` script (verified
  at root and in `frontend/apps/web`). JS quality is `tsc --noEmit` + `vitest run` only.
- → The vendor-import ban cannot be an ESLint rule without introducing a new toolchain. Decision
  (operator-confirmed 2026-06-23): enforce via a **vitest guard test** + structural single-declarer
  + pnpm-strict, matching the existing toolchain (§9).

### 1.8 Capability-assembly ground truth (for §3.2)
- `MetalDocsEditor.tsx:61-62` maps MetalDocs `EditorMode` → eigenpal mode flags.
- `:72-76` assembles the plugin array (`templatePlugin`, sidebar plugin via
  `plugins/sidebarModelBridge.ts`, `externalPlugins`).
- `MetalDocsEditorRef.focus()` is a **no-op** today (`:33`); `getDocumentBuffer()` ≡ `saveNow()`.
- The `plugins/` dir contains **only** `sidebarModelBridge.ts` + `sidebarModelData.ts`.
  There is **no** `OutlinePlugin` and **no** `mergefieldPlugin.ts` → v1's P9 (and tech-debt
  T-004/T-005) describe code that does not exist. **Dropped from this spec as stale.**

### 1.9 Observability requirement (for §8)
- `wiki` REQ-OBS-3 (MUST): a W3C `traceparent` must propagate edge → api → outbox → worker →
  docx-renderer.
- Today the worker→renderer hop uses a plain HTTP transport (`docxrenderer/internal_client.go`,
  no `otelhttp`), and the renderer has **no OTel instrumentation at all**. Tracked as open item
  RF-1. → The renderer is the one hop that breaks the trace.

### 1.10 House conventions (for §3, §7)
- Grade-A: REQ-TOP-1 (published-interface discipline), REQ-TOP-2 (platform layer domain-free).
- Error envelope at the **public API edge**: RFC 9457 problem+json, single writer
  `internal/platform/problem` (ADR 0025, REQ-H-2 MUST). The renderer is an **internal** sidecar,
  so it owes a *structured, classified* error contract — not necessarily problem+json.
- Port house style: interface in the domain package, implementation in infra, wired at the
  composition root, `Noop` null-object for tests, hand-written `fake*`/`stub*` doubles. The narrow
  `FanoutClient` port already exists (`fanout/reconstruction.go:17`) and is the pattern to mirror.
- Next ADR number: **0046**. Closed status vocabulary + the `0029`/`0038` template
  (Status / Last verified / Scope / Context / Decision / Consequences / Alternatives / Related).

### 1.11 Comment flow ground truth (for §3.4)
- `features/documents/hooks/editor/useDocumentComments.ts:2` imports raw eigenpal `Comment` from
  `@metaldocs/editor-ui`; `partitionByMarkers` (`:11-18`) **branches on `comment.id`** → Class A,
  not pass-through.
- `features/documents/queries/useDocumentCommentsQuery.ts` already holds the mappers:
  `rowToLibraryComment(DocumentComment) → Comment` builds the editor shape and **synthesizes
  `initials` via `toInitials`** and `date` from `created_at` (Class B UI fields);
  `libraryCommentToPayloadContent(Comment) → content` is the inverse for persistence.
- `DocumentComment` (API type, `features/documents/api/documents.ts:21`) carries **both** a server
  `id` and a `library_comment_id`; the editor reconciles on `library_comment_id` (the dual-id
  mismatch). Mutations build create/patch payloads from `comment.id` (= `library_comment_id`),
  `parentId`, `author`, `content`, `done`.
- The eigenpal `Comment` shape is `{ id, parentId, author, initials, date, content, done }`;
  `content` is ProseMirror JSON (the opaque body).
- Separate display path: `approval/pages/SignoffDetailPage.tsx:121-125` flattens
  `DocumentCommentResponse.content` via `approval/lib/commentPlainText.ts` — operates on the API
  type, not the editor `Comment`; outside the adapter boundary.

---

## 2. The test that defines "done"

> **"Eigenpal ships a breaking 2.0, or we swap to vendor X. How many files change?"**

**Today:** every file importing an eigenpal type or calling an eigenpal function — `editor-ui`
(`types.ts:2,6`), `docx-renderer` (`fanout.ts:2`), and the dead template-mode file.

**Grade-A target:** only files inside **one adapter package**. Everything else speaks *MetalDocs*
vocabulary — capabilities, an opaque document buffer, a narrow comment DTO, and a classified
`RenderError`. A CI guard test + a strict package manager make regression impossible (§9).

---

## 3. Architecture

### 3.0 The boundary decision rule (how we consume *any* vendor)

The ACL is not "wrap everything." Wrapping a vendor concept that is purely the vendor's
adds a mapping to maintain and buys nothing; leaking a MetalDocs concept couples our logic
to a shape we don't control. So every value that touches the seam is classified once, and
the classification decides its treatment:

| Class | Definition | Treatment | Examples in this system |
|---|---|---|---|
| **A — MetalDocs domain concept** | Identity, persistence, reconciliation, or business rules MetalDocs owns and reasons about | **Crosses as a MetalDocs type.** Vendor field shapes never reach our code. | comment `{id, parentId, author, done}`; editor `mode`; render `RenderError` kind |
| **B — Vendor internal realization** | How the vendor renders/wires a capability | **Stays inside the adapter.** Never crosses. | `templatePlugin`/sidebar plugin wiring; comment `initials`, thread rendering, date formatting; eigenpal `EditorPlugin` |
| **C — Opaque payload** | A blob we move but never inspect or model | **Crosses as an opaque value** (`ArrayBuffer` / JSON), not a structured type | DOCX document bytes; comment `body` (ProseMirror JSON) |

Test for A-vs-pass-through: *does any MetalDocs code read or branch on this value's fields?*
If yes → Class A, give it a MetalDocs type. If the app only hands the value straight back to
the vendor untouched → it's an opaque handle (Class C), don't model it. This is the rule that
makes the document model opaque (§3.3, all-C), comments a slim DTO with an opaque body and
contained vendor UI (§3.4, A+B+C), placeholders a capability (§3.2, B), and errors a taxonomy
(§3.6, A). Applying one rule consistently is the guarantee that the boundary is right, not taste.

### 3.1 One system-wide adapter package — `@metaldocs/eigenpal-adapter`

A new leaf package under `packages/`. It is the **only** place in the repo allowed to depend on
`@eigenpal/*`. Both runtimes consume it: the browser editor (`editor-ui`) and the server render
sidecar (`docx-renderer`).

**Two entrypoints (not an `exports`-map build).** The repo consumes `@metaldocs/*` as raw `.ts`
via Vite alias + tsconfig paths (§1.7) — there is no `dist` and no `exports`-condition resolution
in play. Introducing a built package with `exports` conditions would make the adapter the *only*
package that breaks house convention. Instead the two doors are **two source entry files**, and
which one you import *is* the door:

| Entry file | Contents | Consumers | React? |
|---|---|---|---|
| `src/index.ts` (framework-free) | `TemplateProcessor`, `RenderError` taxonomy, narrow domain DTOs | `docx-renderer` (server, esbuild) + browser | **No** |
| `src/react.tsx` | `EditorMount` component + capability assembly | `editor-ui` (browser, Vite) | Yes |

- Server resolution: `docx-renderer`'s esbuild already inlines `@metaldocs/*` and resolves package
  `main`; `main` → `src/index.ts`. The server **never imports `/react`**, so React never enters
  the backend bundle.
- Browser resolution: add two Vite/tsconfig aliases — `@metaldocs/eigenpal-adapter` → `src/index.ts`
  and `@metaldocs/eigenpal-adapter/react` → `src/react.tsx`.
- **Framework-free guarantee** (the load-bearing invariant): `src/index.ts` must not import from
  `src/react.tsx`, `react`, or `@eigenpal/docx-editor-react`. Enforced by a guard test **and** a
  server bundle assertion (§8). An `exports` map can be added later as hardening if/when the
  package is ever published — explicitly out of scope now.

### 3.2 Capabilities, not plugins

`editor-ui` declares **what it wants** in MetalDocs terms; it does **not** know eigenpal implements
placeholders/sidebar as plugins (§1.8):

```
mountEditor({ mode, placeholders: true, outline: true, comments: {...} })
```

The adapter owns **how**: for eigenpal it reproduces today's `MetalDocsEditor.tsx:72-76` assembly
(`templatePlugin` + `sidebarModelBridge` + `externalPlugins`) and the `:61-62` mode mapping behind
the capability surface. No eigenpal plugin type (`EditorPlugin`, `ReactSidebarItem`) crosses to a
caller. For a vendor where placeholders are native, the same capability flips a native flag — the
caller code is identical. The capability vocabulary is MetalDocs'; the realization is private.

### 3.3 The document model stays opaque (corrected from v1)

No `DocumentBlock`/`DocumentParagraph`/`DocumentTable` domain types. Per §1.4 those eigenpal types
are deep recursive OOXML that no live code manipulates. What actually crosses the boundary:

| Crosses the seam | Form | Rationale |
|---|---|---|
| Document content | `ArrayBuffer` (DOCX bytes), opaque | Live code only ever moves whole buffers (`onAutoSave`, `getDocumentBuffer`, `fanout`). Never structured nodes. |
| Comments | `EditorComment` DTO (flat metadata) | 5 callbacks currently leak raw `Comment` (§3.4). |
| Render outcome | `RenderResult` + `RenderError` | Replaces raw `processTemplateDetailed` result/throw. |

The raw `Paragraph`/`Table`/`BlockContent` re-exports at `editor-ui/src/types.ts:6` are **removed**
(they only fed dead code).

### 3.4 Comment boundary — slim `EditorComment` (applies the §3.0 rule)

Comments are the case where ACL dogma can produce a *worse* design, so they are classified by the
§3.0 rule rather than reflex. The evidence (grounded, not assumed — §1.11):

- The feature layer is **not pass-through**: it runs MetalDocs logic on comments — persistence
  (`useDocumentCommentsQuery` mutations build API payloads from `id`/`parentId`/`author`/`content`/
  `done`) and reconciliation (`useDocumentComments.ts` `partitionByMarkers` branches on `.id`).
  Per §3.0 that logic is **Class A** → it must speak a MetalDocs type, not vendor field shapes.
- Neither existing type is the right boundary verbatim:
  - `DocumentComment` (our API/persisted model) is server-shaped and carries a **dual-id mismatch** —
    a server-row `id` *and* a `library_comment_id`; the editor reconciles on the latter. Using it
    raw would push that ambiguity through the seam.
  - eigenpal `Comment` carries **Class B vendor UI fields** (`initials`, formatted `date`) that exist
    only to render the in-canvas thread — not MetalDocs concepts.
- A facade alias (`type EditorComment = eigenpal Comment`) is rejected: it re-couples every consumer
  to the vendor's exact field names (a 2.0 `done`→`resolved` rename breaks the app).

**Decision:**
- The boundary type is a **slim, vendor-neutral `EditorComment`**: `{ id, parentId?, author, body, done }`.
  `id` is the `library_comment_id` (the value the app actually reconciles on). The five callbacks at
  `types.ts:18-23` change from raw `Comment` to `EditorComment`.
- **`body` is opaque (Class C)** — the ProseMirror JSON already persisted by the backend and already
  flattened for display by `approval/lib/commentPlainText.ts`. No new body structure, no recursive
  OOXML mapping (§3.3).
- **Class B stays in the adapter.** The existing mappers — `rowToLibraryComment`,
  `libraryCommentToPayloadContent`, `toInitials` (today in `useDocumentCommentsQuery`) — move into the
  adapter, which alone produces eigenpal `Comment` (with `initials`/`date`) to feed the vendor's
  thread UI. eigenpal keeps everything it owns; the feature layer stops importing eigenpal `Comment`.
- **Layering after the move:** adapter owns `EditorComment ↔ eigenpal Comment`; the feature/api layer
  owns `EditorComment ↔ DocumentComment`/API payloads. The vendor type is fully contained; the
  persisted type stays in the data layer; `EditorComment` is the neutral currency between them.
- **Counterfactual (honesty check):** were the feature pure pass-through (hand the array straight back
  to the editor, never reading a field), the right answer would be an opaque handle and *no* DTO. The
  slim DTO is justified by real Class-A logic, not by ACL reflex.

### 3.5 Ports (MetalDocs interfaces the app codes against)

- **`EditorMount`** (`src/react.tsx`) — the React editor component + `EditorHandle`
  (`getDocumentBuffer`, `saveNow`, `getPageCount`, `focus`). Today's `MetalDocsEditorRef`, now
  vendor-neutral.
- **`TemplateProcessor`** (`src/index.ts`) — `processTemplate(buffer, values, opts) → RenderResult`.
  Wraps `processTemplateDetailed`. The server's **only** eigenpal touchpoint. Owns the
  error-translation taxonomy (§3.6). Mirrors the existing Go `FanoutClient` port style (§1.10).
- ~~`PlaceholderCodec`~~ — **removed.** It would wrap dead code (§1.3). Live placeholder authoring
  is eigenpal's `templatePlugin`, already behind the `placeholders` capability (§3.2).

### 3.6 Error translation taxonomy (reviewer C1 — new section)

The adapter completes the failure path the seam currently drops (§1.6).

- **In the adapter (`src/index.ts`):** `processTemplate` wraps `processTemplateDetailed` in
  try/catch and translates to a MetalDocs `RenderError`:
  ```
  RenderError = {
    kind: 'template_parse' | 'template_render' | 'undefined_variable' | 'unknown',
    message: string,
    variable?: string,
    cause?: unknown,
  }
  ```
  - Engine throws are **narrowed on the `TemplateError` shape's `.type`** (`'parse'|'render'|
    'undefined'|'unknown'`) — **not** `instanceof` (§1.5).
  - `warnings` / `unreplacedVariables` from a *successful* call are surfaced on `RenderResult`
    (the renderer already reads `unreplacedVariables` at `fanout.ts:68`) — not turned into errors,
    preserving today's `nullGetter: 'empty'` behavior.
- **In `docx-renderer`:** the fanout route maps `RenderError.kind` → a **stable, classified JSON
  error body** + HTTP status (4xx for `template_parse`/`undefined_variable`, 5xx for `unknown`).
  The renderer is internal, so this is a structured contract, **not** problem+json (problem+json is
  the *public-edge* writer per ADR 0025, §1.10).
- **In Go (`client.go`):** classify by status instead of the current opaque
  `fmt.Errorf("fanout status %d")` (§1.6), so the worker can distinguish a permanent template
  defect (don't retry) from a transient engine failure (retry). The exact Go-side change is scoped
  in the implementation plan; the **contract** (classified body) is owned here.

### 3.7 Why the server sidecar stays (and stays on eigenpal)

`docx-renderer` is the right greenfield solution, not inertia. eigenpal is JS-only; server-side
render needs a JS runtime. Embedding JS in Go (goja/v8go) can't run a prosemirror/jszip-scale lib;
reimplementing DOCX fill in pure Go would create **two independent DOCX engines that must agree
forever** — and the forensic *reconstruction* feature hash-compares re-renders (`fanout.ts:63-64`),
so divergence breaks the audit trail. Using eigenpal on **both** editor and render side keeps the
two paths on one engine.

> Precision (reviewer): shared-engine **reduces the probability** of WYSIWYG divergence and the
> reconstruction hash-compare **detects** any residual divergence — it does not make fidelity
> *impossible*. The conclusion (keep the sidecar, keep it on eigenpal) holds; the v1 phrase
> "fidelity by construction" was overstated and is dropped.

Under this design the sidecar's eigenpal seam shrinks to one swap: `fanout.ts:2` imports
`@metaldocs/eigenpal-adapter` (`src/index.ts`) instead of `@eigenpal/docx-editor-core/headless`,
and calls `TemplateProcessor`. The 5 OOXML sub-block renderers are MetalDocs composition logic
(vendor-neutral) and stay in `docx-renderer`.

### 3.8 What changes for everyone else

- `editor-ui` stops being the ACL → becomes a **consumer** of `…/react`. Keeps its React wrapper,
  autosave, sidebar model. Loses all raw eigenpal imports + the `types.ts:6` re-exports + the
  `Comment` leak; builds plugins via adapter capability assembly.
- `docx-renderer/render/fanout.ts` calls `TemplateProcessor` and handles `RenderError`.
- The eigenpal-2.0 / vendor-swap diff = this one package's `src/index.ts` + `src/react.tsx`.

---

## 4. Observability position (REQ-OBS-3 / RF-1)

The worker→renderer hop is the one place a `traceparent` dies today (§1.9). Full remediation is a
TS-OTel rollout in the renderer (no OTel exists there) — too large to bundle into the ACL refactor
and tracked separately as RF-1.

**Decision:** make the seam **trace-ready** now; **bounded-defer** the exporter wiring to RF-1.

- In scope (Phase 3): `processTemplate` accepts an optional propagation context, emits a span
  around the engine call when a tracer is present, and logs `RenderError` with structured fields
  (`kind`, `variable`). The renderer route reads incoming `traceparent`.
- Out of scope (→ RF-1): standing up the OTel SDK/exporter in the renderer and adding `otelhttp`
  to `internal_client.go`. Explicitly cited so the deferral is visible, not silent.

---

## 5. Package-manager consolidation — pnpm, repo-wide (reviewer C3)

**Target:** one root `pnpm-workspace.yaml` covering the JS dirs (`packages/*`,
`apps/docx-renderer`, `frontend/apps/*`, and the new adapter); one lockfile; Go untouched.

**Why pnpm specifically:** pnpm's strict, non-flat `node_modules` makes importing an undeclared
transitive dependency **fail at install** — it structurally bans P1's bug class. npm hoisting is
what *allowed* P1. `frontend/apps/web` is already pnpm, so we converge on the stricter manager.

**The real starting state is messier than "two PMs"** (§1.7) — the migration must clear these
concrete landmines *in order*, each its own verifiable step:

1. **Reconcile the React skew first** (18.2.0 island vs 18.3.1) → pick one version repo-wide.
   A single store with two React copies is its own class of bug.
2. **Delete the cruft:** the malformed `packages/editor-ui/pnpm-workspace.yaml`, the untracked
   `packages/editor-ui/pnpm-lock.yaml`, and the empty committed root `pnpm-lock.yaml` stub.
3. **Author the real root `pnpm-workspace.yaml`**; convert root `package.json` `workspaces`.
4. **Fix the 2 broken CI lockfile references** (§1.7) in the same change — otherwise CI stays red.
5. **One clean `pnpm install`**; convert `npm -ws run` scripts (`build:docx-v2`, etc.) to `pnpm -r`.
6. **Go/no-go checkpoint:** `pnpm install --frozen-lockfile` is green, the full test matrix passes,
   and the renderer bundle assertion (§8) holds. Only then delete `package-lock.json`.

This is the **highest-risk** workstream and is sequenced **last** (§7).

---

## 6. Legacy purge

- **D-1 — `eigenpal-template-mode.ts` + its 3 test files + `__spike__/`.** Dead in production
  (§1.3) and the spike is broken (`Unknown file extension ".css"`, runs 0 tests). **Recommended:
  delete all of it.** Per reviewer C2, do **not** write a new test pinning `runToPlaceholder` —
  that pins dead code. *Gate:* because the branch is active WIP (`eigenpal-review-cockpit`),
  confirm with the owner that none of the three functions is pending-wiring before deleting; git
  history preserves them either way. (This also resolves the original task that opened this
  session — the offending eigenpal import disappears with the file.)
- **D-2 — wiki drift:** fix `editor-ui-eigenpal/_artifacts/03-deps.md` (renderer declares
  `react`, not `core`) and the tech-debt T-006 stale `onLockLost` row. Remove the stale
  T-004/T-005/P9 entries (the `OutlinePlugin`/`mergefieldPlugin` they describe do not exist, §1.8).

## 7. Governance

- **ADR 0046 — Eigenpal Anti-Corruption Layer:** the one-adapter, two-entrypoint, capabilities-
  not-plugins, opaque-document-model, classified-`RenderError`, vendor-ban rule. Relates to /
  amends ADR 0001 (eigenpal adoption). Closes tech-debt T-008. Use the `0029`/`0038` template;
  Status `Proposed` → `Accepted (in execution)` on landing.
- **ADR 0047 — `templatePlugin` mode-gating:** records the `:61-62`/`:72-76` mode→plugin rule.
  Closes T-007.
- Refresh `wiki/modules/editor-ui-eigenpal*` + `03-deps.md` to the new topology.

## 8. Phases (ordered by risk, low → high)

1. **Phase 1 — Legacy purge** (§6). Lowest risk; shrinks the surface before anything is built.
   Removes the failing suite and the dead file (after the D-1 gate).
2. **Phase 2 — Governance** (§7). Low risk; ADRs 0046/0047 + wiki, lands alongside.
3. **Phase 3 — The adapter** (§3, §4). The bulk:
   - New package; two source entrypoints (§3.1).
   - `TemplateProcessor` + `RenderError` taxonomy (§3.6); migrate `fanout.ts` + classify in
     `client.go`.
   - `EditorMount` capability assembly (§3.2); migrate `editor-ui`; resolve the comment-DTO
     evidence gate (§3.4).
   - Kill the `any` — typing it surfaces and fixes the `"dropdown"`→`"dropDownList"` bug (§1.2).
   - Trace-ready seam (§4).
   - Land the vendor-ban guard test (§9) + the server bundle assertion.
4. **Phase 4 — pnpm consolidation** (§5). Highest risk, last, against an already-honest dependency
   graph; the 6-step landmine sequence with a go/no-go checkpoint before deleting `package-lock.json`.

Rationale: purge and govern cheaply; build the adapter so every dep is declared honestly; *then*
flip the package manager so one clean `--frozen-lockfile` install is the final proof.

## 9. Enforcement & testing

- **Vendor-ban guard (no linter exists — see §1.7a):** enforcement is defense-in-depth using the
  repo's existing `tsc`+`vitest` toolchain, **not** ESLint:
  1. **Guard test** — a vitest test that walks the source tree and **fails** if any `@eigenpal/*`
     import appears outside `packages/eigenpal-adapter/`.
  2. **Structural** — only the adapter's `package.json` declares `@eigenpal/*`.
  3. **pnpm-strict** — Phase 4's strict `node_modules` fails an undeclared import at install.
  (Adopting ESLint import-boundaries repo-wide is a separate toolchain decision, deliberately not
  bundled into this refactor.)
- **Framework-free assertion:** a vitest/CI check that `src/index.ts`'s import graph (and the
  emitted server bundle metafile, below) contains no `react` / `@eigenpal/docx-editor-react` (§3.1).
- **Adapter `src/index.ts`:** unit tests for `TemplateProcessor` over a fixture DOCX (server-safe,
  no React/CSS) **including** the `RenderError` translation cases per `.type` (§3.6).
- **Adapter `src/react.tsx`:** mount + capability-assembly tests (jsdom), mirroring existing
  `editor-ui` mocks; a hand-written `fake`/`stub` `TemplateProcessor` for consumer tests
  (mirroring the Go `fakeFanoutClient` style, §1.10).
- **`editor-ui` / `docx-renderer`:** existing suites pass unchanged in behavior; imports now point
  at the adapter.
- **Install gate:** `pnpm install --frozen-lockfile` green with no phantom-dep resolution.
- **Bundle-size gate (Major):** `docx-renderer/build.mjs` emits an esbuild **metafile** so a CI
  check can assert React never enters the server bundle and flag size regressions (today there is
  no metafile, §1.7).

## 10. Non-goals (YAGNI)

- No full OOXML domain model (§0/§1.4) — the document is opaque bytes.
- No `PlaceholderCodec` port (§3.5) — it wrapped dead code.
- No published-package `exports` map now (§3.1) — two source entries suffice; revisit only on publish.
- No full renderer OTel rollout (§4) — RF-1.
- No `worker_threads` pool for `processTemplateDetailed` — separate perf workstream.
- No second vendor adapter — the design *enables* one; we don't build it speculatively.
- Go backend, messaging/outbox, Gotenberg PDF path: untouched.

## 11. Risks

- **pnpm migration (Phase 4)** — the real risk, now with named landmines: React skew, malformed
  workspace file, empty lock stub, 2 broken CI refs (§1.7/§5). Mitigation: the 6-step sequence +
  go/no-go checkpoint; done last against an honest graph.
- **Comment-DTO migration (§3.4)** — relocating `rowToLibraryComment` into the adapter and
  switching the five callbacks to `EditorComment` touches the live review-cockpit comment flow.
  Mitigation: the row↔vendor mapper moves verbatim (behavior-preserving); `body` stays the existing
  ProseMirror JSON; existing comment tests port onto the new boundary before cutover.
- **Capability-assembly parity (§3.2)** — the adapter must reproduce `:61-62`/`:72-76` exactly.
  Mitigation: port the existing `editor-ui` tests onto the adapter.
- **Error-contract reach (§3.6)** — the classified body touches three layers (adapter, renderer
  route, Go client). Mitigation: the *contract* is owned here; the Go change is scoped, small, and
  test-covered (classify-by-status).
- **Server bundle (§3.1/§9)** — the framework-free guarantee is an invariant, not a hope; the
  metafile assertion is what enforces it.
