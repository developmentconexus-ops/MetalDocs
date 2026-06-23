# Eigenpal Anti-Corruption Layer — Design Spec

> **Date:** 2026-06-23
> **Branch context:** discovered on `eigenpal-review-cockpit` during the eigenpal 1.9 migration
> **Status:** Design approved (pending written-spec review) — implementation plan to follow

## 1. Problem

`@eigenpal/*` is a third-party DOCX editor/engine vendor. MetalDocs is supposed to consume it
through a single Anti-Corruption Layer (ACL) so the rest of the system never couples to the
vendor. Today that boundary leaks. A review of the current integration found:

| # | Severity | Problem |
|---|---|---|
| P1 | 🔴 Major | **Phantom dependency.** No `package.json` in the repo declares `@eigenpal/docx-editor-core`. Three production files import it (`editor-ui/src/types.ts`, `editor-ui/src/index.ts`, `apps/docx-renderer/src/render/fanout.ts`). It only resolves transitively via `@eigenpal/docx-editor-react@1.9.0 → @eigenpal/docx-editor-core@^1.9.0`. If react drops core, or strict hoisting flips, every import breaks. |
| P2 | 🔴 Major | **Two package managers.** Root = npm workspaces (`editor-ui`, `docx-renderer`); `frontend/apps/web` = pnpm. Each resolves the bare `@eigenpal/*` specifier against its own tree. Root cause of the documented "junction drift" instability and duplicate installs. |
| P3 | 🔴 Major | **`any` escape hatches.** `eigenpal-template-mode.ts:29` builds SDT props as `any`; `:79` uses fragile structural casts into eigenpal node shapes. No type safety against the vendor's actual types. |
| P4 | 🟠 Arch | **The ACL is a pass-through, not an ACL.** `editor-ui` re-exports *raw* eigenpal types (`BlockContent`, `Paragraph`, `Table`, `Comment`, `EditorPlugin`) unchanged. MetalDocs callers depend on eigenpal's exact type shapes. A vendor type change ripples repo-wide. |
| P5 | 🟠 Arch | **Deep-subpath coupling.** `core/types/document`, `core/types/content`, `core/headless` reach into eigenpal's internal module layout. |
| P6 | 🟡 Legacy | **Broken spike.** `frontend/apps/web/src/editor-adapters/__spike__/eigenpal-placeholder-spike.test.ts` fails to load under 1.9 (`Unknown file extension ".css"`), runs 0 tests, fails the suite. Exploratory P0.1 scaffolding (`7f0341cd`). |
| P7 | 🟡 Legacy | **`runToPlaceholder` has no live coverage** — only the broken spike "covered" the reverse mapping. |
| P8 | 🟡 Drift | Stale wiki: `03-deps.md` claims `docx-renderer` declares core (it declares react only); tech-debt T-006 claims `types.ts:25` declares `onLockLost` (it does not). |
| P9 | 🟡 Legacy | Dead/mislabeled `editor-ui` surface: `createOutlinePlugin` exported but never registered (T-004); `mergefieldPlugin.ts` exports a data fn, not a plugin (T-005). **Out of scope here** — editor-ui hygiene, not the eigenpal seam; tracked in the tech-debt register. |
| P10 | 🟡 Gov | No ADR for the ACL/wrapper-only rule (T-008) or the `templatePlugin` mode-gating rule (T-007). The boundary we enforce has no decision record. |

## 2. The test that defines "done"

> **"Eigenpal ships a breaking 2.0, or we swap to vendor X. How many files change?"**

**Today:** every file importing an eigenpal type or calling an eigenpal function — scattered across
`editor-ui`, `docx-renderer`, and feature code.

**Grade-A target:** only files inside **one adapter package**. Everything else speaks *MetalDocs*
vocabulary and never knew eigenpal existed. A CI lint rule makes regression impossible.

## 3. Architecture

### 3.1 One system-wide adapter package — `@metaldocs/eigenpal-adapter`

A new leaf package. It is the **only** place in the repo allowed to depend on `@eigenpal/*`.
Both runtimes consume it: the browser editor (`editor-ui`) and the server render sidecar
(`docx-renderer`).

**Two doors (subpath export conditions)** so the server never pulls React into a backend bundle:

| Entry | Contents | Consumers | React? |
|---|---|---|---|
| `.` (framework-free) | Domain types, placeholder codec, headless template processing (`processTemplate`) | `docx-renderer` (server) + browser | No |
| `./react` | Editor mount component + capability assembly | `editor-ui` (browser) | Yes |

The server door (`.`) is Node-safe — no DOM/React assumptions. A Google-senior default:
never make the backend pay for frontend dependencies.

### 3.2 Capabilities, not plugins

`editor-ui` declares **what it wants** in MetalDocs terms — it does **not** know eigenpal
implements placeholders as a plugin:

```
mountEditor({ placeholders: true, outline: true, comments: {...} })
```

The adapter owns **how**: for eigenpal it assembles `templatePlugin` + the sidebar plugin
behind the scenes; for a vendor where placeholders are native, it flips native flags. The
capability vocabulary (`placeholders`, `outline`, `comments`) is MetalDocs'. The realization
is the adapter's private business. No eigenpal plugin type (`EditorPlugin`, `ReactSidebarItem`)
ever leaks to a caller.

### 3.3 Domain types are ours; the adapter maps at the seam

MetalDocs-owned types replace the raw re-exports. No raw eigenpal type crosses the boundary —
not even re-exported.

| MetalDocs type | Maps eigenpal | Replaces |
|---|---|---|
| `TemplatePlaceholder` | inline/block SDT nodes | `PlaceholderRun` in `eigenpal-template-mode.ts` |
| `DocumentBlock` / `DocumentParagraph` / `DocumentTable` | `BlockContent` / `Paragraph` / `Table` | raw re-exports in `editor-ui/types.ts` |
| `EditorComment` | `Comment` | raw `Comment` re-export |

### 3.4 Ports (MetalDocs interfaces the app codes against)

- **`EditorMount`** — the React editor component + `EditorHandle` (`getDocumentBuffer`, `saveNow`,
  `getPageCount`, `focus`). Today's `MetalDocsEditorRef`, now vendor-neutral. (`./react` door.)
- **`TemplateProcessor`** — `processTemplate(buffer, values) → RenderResult`. Wraps
  `processTemplateDetailed`. The server's **only** eigenpal touchpoint. (`.` door.)
- **`PlaceholderCodec`** — `toNode(TemplatePlaceholder)` / `fromNode(node) → TemplatePlaceholder | null`.
  Today's `placeholderToRun` / `runToPlaceholder`, but **typed against eigenpal internally,
  MetalDocs types at the surface** — kills the `any` (P3). (`.` door.)

### 3.5 Why the server sidecar stays (and stays on eigenpal)

`docx-renderer` is the right solution greenfield, not by inertia. eigenpal is JS-only;
server-side render must have a JS runtime. Embedding JS in Go (goja/v8go) can't run a
prosemirror/jszip-scale lib; reimplementing DOCX fill in pure Go would create **two independent
DOCX engines that must agree byte-for-byte forever** — and the forensic *reconstruction* feature
hash-compares re-renders, so divergence breaks the audit trail. Using eigenpal on **both** the
editor and the render side guarantees WYSIWYG fidelity **by construction**. The vendor coupling
on the render path is the feature, not debt.

Under this design the sidecar's eigenpal seam shrinks to one call: `docx-renderer/render/fanout.ts`
imports `@metaldocs/eigenpal-adapter` (`.` door) instead of `@eigenpal/docx-editor-core/headless`.
The 5 OOXML sub-block renderers are MetalDocs composition logic (vendor-neutral) — they stay in
`docx-renderer`, not the adapter.

### 3.6 What changes for everyone else

- `editor-ui` stops being the ACL → becomes a **consumer** of `./react`. Keeps its React wrapper,
  autosave, sidebar model. Loses all raw eigenpal imports + type re-exports; builds plugins via
  adapter capability assembly.
- `docx-renderer/render/fanout.ts` calls `TemplateProcessor` instead of `processTemplateDetailed`.
- The eigenpal-2.0 / vendor-swap diff = this one package's `.` and `./react` internals. Nothing else.

## 4. Package-manager consolidation — pnpm, repo-wide

A single `pnpm-workspace.yaml` at the repo root covering `packages/*`, `apps/*`,
`frontend/apps/*`, and `@metaldocs/eigenpal-adapter`. One lockfile.

**Why pnpm specifically:** pnpm's strict, non-flat `node_modules` makes importing an undeclared
transitive dependency **fail at install time** — it structurally bans P1's bug class. npm
workspaces' hoisting is what *allowed* P1; staying on npm means fixing P1 by hand with nothing
stopping its return. Half the repo (`frontend/apps/web`) is already pnpm, so we converge toward
the stricter manager. The content-addressed store also kills the duplicate eigenpal installs and
the documented junction-drift instability.

**Migration:** convert root `package.json` `workspaces` → `pnpm-workspace.yaml`; delete
`package-lock.json`; one clean `pnpm install`; `npm -ws run` scripts (`build:docx-v2`, etc.) →
`pnpm -r`. Go backend untouched. This is the highest-risk workstream — it also finally completes
the half-finished frontend pnpm install (root cause of junction drift).

## 5. Enforcement — vendor ban as lint

Repo-wide ESLint `no-restricted-imports` (or `eslint-plugin-import` boundaries) banning
`@eigenpal/*` in every package **except** `@metaldocs/eigenpal-adapter`. Plus: only the adapter's
`package.json` lists `@eigenpal/*`. The package manager + the linter together make the boundary
self-enforcing — rule #3 ("vendor name appears in exactly one place") cannot regress silently.

## 6. Legacy purge (P6, P7, P8)

- Delete `frontend/apps/web/src/editor-adapters/__spike__/` (broken, zero live coverage).
- Recover `runToPlaceholder` coverage with a framework-free round-trip unit test
  (`fromNode(toNode(p))` — no DOCX bytes, no prosemirror, no CSS). The byte-serialization the
  spike exercised is eigenpal's own concern, not MetalDocs logic.
- Fix wiki drift: `03-deps.md` renderer claim; tech-debt T-006 stale `onLockLost` row.

## 7. Governance (P10)

- ADR: **Eigenpal Anti-Corruption Layer** — the one-adapter, capability-port, vendor-ban rule
  (closes T-008).
- ADR: **`templatePlugin` mode-gating** — closes T-007.
- Refresh `wiki/modules/editor-ui-eigenpal*` and `03-deps.md` to the new topology.

## 8. Phases (ordered by risk, low → high)

1. **Phase 1 — Legacy purge** (§6). Low risk, immediate. Removes the failing suite, restores coverage.
2. **Phase 2 — Governance** (§7). Low risk. ADRs + wiki, can land alongside.
3. **Phase 3 — The adapter** (§3). The bulk. New package, domain types, ports, capability assembly,
   migrate `editor-ui` + `docx-renderer` consumers, kill `any`. Land the lint ban here.
4. **Phase 4 — pnpm consolidation** (§4). Highest risk, done last so the adapter's deps are declared
   honestly before the workspace is rebuilt. One clean install validates the whole topology.

Rationale for order: purge and govern cheaply first; build the adapter so all deps are explicit;
*then* flip the package manager so the single clean `pnpm install` is the final proof that no
phantom deps remain and the boundary holds.

## 9. Testing strategy

- **Adapter `.`**: unit tests for `PlaceholderCodec` round-trip and `TemplateProcessor` over a
  fixture DOCX (server-safe, no React/CSS).
- **Adapter `./react`**: mount + capability-assembly tests (jsdom), mirroring existing
  `editor-ui` test mocks.
- **`editor-ui`**: existing suites pass unchanged in behavior; imports now point at the adapter.
- **`docx-renderer`**: existing `fanout` + sub-block suites pass; the eigenpal call is now via the
  adapter.
- **Lint gate**: a CI check that `@eigenpal/*` appears only inside the adapter package.
- **Install gate**: `pnpm install --frozen-lockfile` succeeds with no phantom-dep resolution.

## 10. Non-goals (YAGNI)

- No full MetalDocs *plugin framework* — capabilities are config, not a plugin abstraction.
- No `worker_threads` pool for `processTemplateDetailed` — real throughput concern, **separate**
  perf workstream, not this refactor.
- No second vendor adapter now — the design *enables* one; we don't build one speculatively.
- Go backend, messaging/outbox, Gotenberg PDF path: untouched.

## 11. Risks

- **pnpm migration** (Phase 4) is the real risk: lockfile churn, the existing junction-drift state, CI
  scripts. Mitigation: do it last, against an already-honest dependency graph; validate with a
  clean `--frozen-lockfile` install and the full test matrix.
- **Capability-assembly parity**: the adapter must reproduce today's exact plugin wiring
  (`templatePlugin` gating, sidebar model) — covered by porting existing `editor-ui` tests.
- **Server bundle**: the `.`/`./react` split must actually keep React out of `docx-renderer`'s
  esbuild bundle — verify bundle contents post-migration.
