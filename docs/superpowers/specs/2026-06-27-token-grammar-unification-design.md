# Token Grammar Unification — SP-1 Entry Gate

**Date:** 2026-06-27
**Parent:** `2026-06-27-template-tokens-north-star.md`
**Status:** Approved design. This is an **entry gate for SP-1** — land before the tenant token dictionary, not after.
**Scope:** `@metaldocs/shared-tokens` (grammar core), `@metaldocs/editor-ui` (detection consumer), `apps/docx-renderer` (parity test), spec/wiki (boundary invariant). **No** Go change, **no** new capability, **no** tenant dictionary, **no** render-pipeline change.

---

## 1. Problem & intent

MetalDocs has two **independent** `{token}` parsers that are safe today only because the
computed catalog is closed and snake_case:

1. **Detection** — `@metaldocs/editor-ui` `getUsedTokens()` hand-rolls `/\{([A-Za-z0-9_]+)\}/g`
   across the body and every header/footer band (a *deliberate* vendor-independent text-parse;
   the vendor's native `getTemplatePluginTags` is body-only and cannot see HF bands —
   `wiki/modules/editor-ui-eigenpal.md` §8.4, Task 7–9, 2026-06-27).
2. **Freeze** — `apps/docx-renderer` `fanout()` → `@metaldocs/eigenpal-adapter` →
   the vendor's docxtemplater (`@eigenpal/docx-editor-core/headless`), default `{`/`}`
   delimiters, tag body `[^{}]+`.

Because they are different parsers, the detector's accept-set can diverge from what freeze
acts on. The divergence is **already real**: an author who types `{a.b}` gets a token freeze
acts on (docxtemplater nested-path → blanked / `UnreplacedVar`) but the `[A-Za-z0-9_]+`
detector silently drops — invisible to the author, surfaced only as a silent unreplaced var
at freeze. When SP-1's tenant dictionary lands (authors define arbitrary keys), this
silent under-reporting widens.

**Intent:** make detection and freeze provably agree, from one grammar core, and write the
boundary that keeps future increments (SP-1/SP-2) from re-forking the grammar — especially
in Go.

## 2. Binding constraints (non-negotiable)

Inherits north-star §2. Adds, as the durable root-cause fix:

- **Token grammar is Node-owned.** Token parsing lives in exactly two places: the
  **adapter detection** (`@metaldocs/editor-ui`, browser) and the **vendor freeze**
  (docxtemplater, in `apps/docx-renderer`). Both consume the **single grammar core** in
  `@metaldocs/shared-tokens`. **Go never parses tokens** and never grows a `{...}` regex.
- **SP-2 tag-validation is membership, not parsing.** When SP-2 validates "template tokens ⊆
  registry", it does so over the **detector's output** (`getUsedTokens()` /
  `detectTokens()`) plus the fanout `unreplacedVars` backstop — never a second server-side
  token parser. If SP-2 needs server-side extraction (template saved without a client pass),
  it calls the Node renderer's grammar core, not a Go reimplementation.
- **Detect broad, validate strict.** Detection extracts every `{[^{}]+}` (minus docxtemplater
  control prefixes) so nothing freeze acts on is ever invisible. *Validity* (is this a
  legal key) is a separate classification step over the canonical charset.
- **Canonical charset = snake_case.** `IDENT_RE = /^[A-Za-z_][A-Za-z0-9_]*$/`
  (`shared-tokens/src/grammar.ts`, already present). Dotted / hyphenated keys are **invalid
  keys** — they sidestep docxtemplater's nested-property and arithmetic semantics. This
  governs **key validity and SP-1 dictionary-key creation**, NOT detection breadth.
- **One scanner.** The text scan and the docx-run scan share a single core. The entry gate
  must *reduce* the parser count, never add one.

## 3. Current state (runtime truth, verified 2026-06-27)

- `shared-tokens/src/grammar.ts` — `IDENT_RE`, `RESERVED_IDENTS`, `MAX_SECTION_DEPTH`,
  `isValidIdent`, `isReservedIdent`. Live (exported, used by `parser.ts` + tests).
- `shared-tokens/src/parser.ts` — `parseDocxTokens(buf)`: `TOKEN_RE = /\{([#^/])?([^{}]+)\}/g`
  over joined run text, classifies var/section/inverted/closing, validates idents.
  **Imported only by its own tests** — dead scaffolding for SP-2, not wired into any runtime.
  Same for `diff.ts`, `ooxml.ts`.
- `editor-ui/src/MetalDocsEditor.tsx` `getUsedTokens()` — hand-rolled
  `/\{([A-Za-z0-9_]+)\}/g` over `textBetween` for body + every `getHfPmViews()` band.
  Already depends on `@metaldocs/shared-tokens` (`package.json`) but does not use it.
- `apps/docx-renderer/src/render/fanout.ts` — real vendor freeze. The adapter
  (`eigenpal-adapter/src/index.ts`) maps **both** `replacedVariables` and
  `unreplacedVariables`, but `fanout()`'s `FanoutResult` (`fanout.ts:17-21,65-69`)
  **returns only `unreplacedVars` and discards `replacedVariables`** — so the parity oracle
  in §4.3 is not buildable from `fanout()` as it stands today (see §4.3 fix).
- `packages/shared-tokens/package.json` — source-only package (`main`/`types` →
  `./src/index.ts`), **no `sideEffects` field, no `exports` map**; barrel `index.ts`
  re-exports `parseDocxTokens` from `parser.ts`, which top-imports `jszip` +
  `fast-xml-parser`. So importing `detectTokens` via the barrel risks pulling JSZip into the
  browser bundle (see §4.1 fix).
- `packages/editor-ui/test/public-surface.test.ts` — ACL guard greps `src/index.ts` +
  `src/types.ts` for `/@eigenpal/` only. A `shared-tokens`-owned `DetectedToken` on the ref
  is therefore safe; a vendor type would not be.
- **No Go token grammar exists** (`internal/modules` has no `{...}` token regex).

## 4. Design

### 4.1 Grammar core (`@metaldocs/shared-tokens`)

Add to `grammar.ts` (pure, zero runtime deps — no JSZip / fast-xml-parser).

**Browser-bundle safety (resolves the JSZip-via-barrel risk):** `package.json` must add
**both** (a) `"sideEffects": false`, and (b) an `exports` map with a `./grammar` subpath:

```json
"exports": {
  ".":        { "import": "./src/index.ts", "types": "./src/index.ts" },
  "./grammar": { "import": "./src/grammar.ts", "types": "./src/grammar.ts" }
},
"sideEffects": false
```

editor-ui imports the detector from the subpath — `import { detectTokens } from
'@metaldocs/shared-tokens/grammar'` — so the JSZip-importing barrel is never in the
browser graph regardless of tree-shaking. The plan verifies the Vite path-alias
(`frontend/apps/web/vite.config.ts`, `tsconfig.json`) resolves the subpath; if the
source-only alias cannot honor `exports` subpaths, fall back to a direct module import of
`grammar.ts` (still barrel-free). Either way the barrel is not the import path.

```ts
export type TokenKind = 'var' | 'section' | 'inverted' | 'closing' | 'partial';
export interface RawTag { raw: string; kind: TokenKind; inner: string; start: number; end: number; }
export interface DetectedToken { name: string; kind: TokenKind; valid: boolean; reserved: boolean; }

/** The single scan core: every `{...}` tag in a flat string, control-prefix classified. */
export function scanText(text: string): RawTag[];

/** Detection contract: broad extraction + strict classification, var-only by default. */
export function detectTokens(text: string): DetectedToken[];
```

- `scanText` owns the one `TOKEN_RE`. Control prefixes `#` `^` `/` `>` → `section` /
  `inverted` / `closing` / `partial`; otherwise `var`. `inner` is trimmed.
- `detectTokens` = `scanText` → keep `kind === 'var'` → annotate `valid`
  (`isValidIdent`) and `reserved` (`isReservedIdent`). **Returns invalid/reserved tokens
  too** (with `valid:false`) — that is what kills the silent under-report.
- **Trim once, in the core.** `scanText` trims `inner` (matching `parser.ts`'s current
  `inner.trim()` under `trimValues:false`) so the browser detector and the docx parser treat
  `{ name }` identically. The core is the only place trimming happens.
- Refactor `parser.ts` `scanTokens` to consume `scanText` (run-span mapping stays in
  `parser.ts`; the regex + kind logic moves to the core). **Behavior-preserving caveat:**
  current `scanTokens` classifies only `#^/` prefixes, so `{>x}` parses as a `var` with ident
  `">x"` → `malformed_token`. `scanText` adds `>`→`partial`, which would reclassify `{>x}`
  out of the var path and could drop that error. `parseDocxTokens` MUST keep emitting the
  same diagnostic for unsupported/partial tags — the plan adds a `{>x}` regression assertion
  to lock the error path. Token *output* is otherwise unchanged (split-runs, section-depth,
  trim all stay in `parser.ts`), confirmed against the existing `parser.*.test.ts` suite.

### 4.2 Detection consumer (`@metaldocs/editor-ui`)

`getUsedTokens()` (the `MetalDocsEditorRef` contract — unchanged signature) now:

- builds the per-band text exactly as today (body + `getHfPmViews()`),
- calls `detectTokens(text)` per band, unions first-seen,
- returns valid var **names** (`valid && !reserved`) — preserving the existing page
  contract (`TemplateEditorPage` `partitionTokens` still gets `string[]`).

**Ref contract change (settled — no open options):** add `getDetectedTokens():
DetectedToken[]` to `MetalDocsEditorRef` in `editor-ui/src/types.ts`, with `DetectedToken`
**imported/re-exported from `@metaldocs/shared-tokens`** (never a vendor type). `getUsedTokens()`
becomes a thin filter over `getDetectedTokens()` (`valid && !reserved`, var names) — existing
`string[]` contract unchanged. Because `types.ts`/`index.ts` name only `@metaldocs/shared-tokens`,
the `public-surface.test.ts` ACL guard (greps `/@eigenpal/`) stays green. The page consumes
`getDetectedTokens()` to render the invalid-token warning.

### 4.3 Parity test (`apps/docx-renderer`)

The only site where the real vendor engine runs server-side.

**Prerequisite (resolves B1 — the oracle needs `replacedVars`):** `fanout()` currently
discards `replacedVariables`. The plan widens `FanoutResult` to
`{ buffer, contentHash, unreplacedVars, replacedVars }` and surfaces
`result.replacedVariables` at `fanout.ts:65-69`. This is a **return-shape widening only — no
change to substitution logic or the frozen bytes** (contentHash is computed over `buffer`,
untouched). Existing fanout tests stay green (additive field). Alternative considered and
rejected: calling the adapter directly in the test — rejected because the parity gate should
exercise the same `fanout()` path production freezes through.

Build a golden `.docx` (existing `buildTemplateDocx` pattern) whose body contains a probe
set:

```
{a_b} {ABC} {x1} {_y}        ← canonical-valid
{a.b} {a-b} {1n} {a b}       ← canonical-invalid but freeze may act on them
```

Provide values for **all** probes (including space/dotted shapes — keys like `"a b"`,
`"a.b"` in the value map). Run the real `fanout()`. Assert the contract over the widened
`FanoutResult`:

1. **No silent loss:** `detectTokens(probeText)` detected-set (var kind, any validity) ⊇
   freeze's touched-set = `replacedVars ∪ unreplacedVars`. Anything freeze acts on is
   detected.
2. **Validity ⟺ substitution:** for each probe, `isValidIdent` valid ⟺ the probe is in
   `replacedVars` (substituted), invalid ⟺ in `unreplacedVars` or not a flat tag.
   (Documents and pins the vendor's actual charset behavior; if a vendor upgrade shifts it,
   this test fails loudly — intended alarm.)

The probe set deliberately includes `{a b}` (internal space): `[^{}]+`-valid for the vendor
and `scanText`, `isValidIdent`-invalid. The test first asserts the engine **emits it in
`unreplacedVars` rather than throwing** — if the vendor hard-errors on such a tag, that is a
finding to record (the spec's `invalid ⟺ unreplacedVars` arm assumes graceful non-substitution),
not a test to silently weaken.

If the probe reveals the vendor accepts a shape our canonical grammar rejects (or vice
versa), that is an architecture signal to resolve in the spec, not to paper over in the test.

### 4.4 Spec & wiki (the boundary)

- This document records the §2 boundary invariants as binding for SP-1/SP-2.
- `wiki/modules/editor-ui-eigenpal.md` §8.4 — update the detection subsection to point at the
  shared `detectTokens` core and the detect-broad/validate-strict model.
- `wiki/concepts/token-syntax.md` — note the snake_case key-validity rule and the
  Node-owned-grammar boundary.
- Relocate the snake_case key-validation requirement into SP-1's spec as the
  dictionary-key creation constraint (`isValidIdent` + `isReservedIdent` + computed-catalog
  collision check).

## 5. Testing

- **shared-tokens unit:** `scanText` (control-prefix classification, trims, split braces,
  empty/`{}`), `detectTokens` (valid / invalid-ident / reserved annotation, var-only).
- **shared-tokens regression:** existing `parseDocxTokens` tests stay green after the
  `scanText` refactor; **add a `{>x}` assertion** locking the unsupported/partial-tag
  diagnostic so the prefix reclassification (§4.1) cannot silently drop it.
- **docx-renderer regression:** existing `fanout.*.test.ts` stay green after the
  `FanoutResult` widening (additive `replacedVars`).
- **editor-ui unit** (`MetalDocsEditor.tokens.test.tsx`, extend): `getUsedTokens()` returns
  valid var names across body + HF bands; `getDetectedTokens()` surfaces invalid/reserved.
- **docx-renderer parity** (new): §4.3 contract.
- All via the canonical frontend/node test framework (no bespoke harness).

## 6. Out of scope (deferred to their SP)

- SP-1 tenant dictionary domain / storage / capabilities / CRUD (consumes the relocated
  snake_case key-validation rule).
- SP-2 freeze injection + tag-validation (consumes `detectTokens` / `unreplacedVars`; wires
  the currently-dead `parser.ts`/`diff.ts` instead of re-forking).
- Wiring `parser.ts`/`diff.ts`/`ooxml.ts` into a runtime path.
- Any Go change.

**In scope, narrowly:** a `FanoutResult` return-shape widening (additive `replacedVars`,
§4.3) — this is the *only* render-package touch, and it changes no substitution logic or
frozen bytes.

## 7. Risks & trade-offs

- **Parity test pins current vendor behavior.** A vendor upgrade that changes the charset
  breaks the test — intended; it is the drift alarm.
- **Browser bundle weight.** `package.json` today has neither `sideEffects` nor an `exports`
  map, and the barrel re-exports the JSZip-importing `parser.ts`. Mitigation is **not**
  "trust tree-shaking" — it is the `./grammar` subpath import + `sideEffects:false` from §4.1,
  so the barrel is never in the browser graph. Plan verifies the Vite alias honors the
  subpath.
- **Snake_case-only forever** rejects dotted/hyphenated keys at the source — deliberate,
  documented; sidesteps docxtemplater nested-property semantics.
- **Page contract unchanged.** `getUsedTokens(): string[]` stays; the new `getDetectedTokens`
  is additive, so SP-0's `TemplateEditorPage` wiring is not disturbed.
