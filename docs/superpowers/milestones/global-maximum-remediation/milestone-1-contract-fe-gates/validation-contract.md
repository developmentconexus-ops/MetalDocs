# Milestone 1 — Validation Contract (D4, authored BEFORE implementation)

> **Program:** global-maximum-remediation · **Milestone:** 1 — Contract & frontend governance gates
> **Authored:** 2026-07-03 — **before any implementation task began** (mission.md D4).
> **Binding:** This document is the acceptance contract. Implementation is compared against it;
> divergence is **HS-7** (fix the code to the contract, or re-open the contract with operator
> approval — never silently rationalize the contract to match the code). The `milestone-validator`
> reads this file to judge C1/C2/C5 compliance.
> **Derived from:** mission.md §7 M1; the 2026-07-03 final architecture review §Cross-cutting (items
> 3,4,5), §Dimension 3, §Priorities P1.3/P1.4; and the runtime-truth investigation recorded in
> `milestone.md §Discovered runtime truth`.

Every gate in this milestone must satisfy the universal **D4 proof shape**:

- **POSITIVE proof** — from a clean tree the gate runs and **passes** (exit 0 / green), with captured
  command output in the feature's `evidence.md`.
- **NEGATIVE proof** — a temporary fixture / injected change makes the gate **fail** (non-zero exit /
  red), captured verbatim, then reverted. A gate that cannot be made to fail is not a gate.
- **CI wiring point** — the exact workflow file + job that runs the gate blocking on the relevant PR
  path filter.

No gate is "reported-only" — all are blocking by construction (repo Principle-5 posture).

---

## F1.1 — oasdiff breaking-change gate

### Expected behavior
A CI job diffs the PR's `api/openapi/v1/openapi.yaml` against the **base branch** version using
`oasdiff` (breaking-changes mode) and fails on any breaking change.

- **Tool:** `oasdiff` (github.com/oasdiff/oasdiff). Not present on the box → the CI job installs it
  (Go install or the published action). Local proof may `go install github.com/oasdiff/oasdiff@latest`
  or use the container; whatever is used is recorded.
- **Command shape:** `oasdiff breaking <base-spec> <revision-spec> --fail-on ERR` (exact flags recorded
  in evidence; `--fail-on ERR` or `changelog --fail-on` is acceptable so long as a breaking change
  yields non-zero exit).
- **Base resolution in CI:** the job checks out the PR base ref (e.g. `git show origin/main:api/openapi/v1/openapi.yaml`
  or an `actions/checkout` of the base) as the "from" spec, and the PR head as the "to" spec.

### POSITIVE proof
`oasdiff breaking <spec> <spec>` (spec against itself) → **"No breaking changes"** / exit 0. And a
**backward-compatible** synthetic change (e.g. add a new **optional** response field, or add a new
path) → exit 0.

### NEGATIVE proof (the fixture)
A synthetic **breaking** change to a throwaway copy of the spec, at least one of:
- delete an existing path or operation, **or**
- remove a property from a response schema's `required` list → make it absent, **or**
- narrow an enum / change a field type.

Running the gate `from = original spec`, `to = broken spec` → **non-zero exit**, output naming the
breaking change. Captured verbatim, fixture discarded (never committed).

### CI wiring point
New workflow `.github/workflows/openapi-breaking.yml` (or a new job in `api-contract.yml`), triggered
`on: pull_request: paths: ['api/openapi/v1/openapi.yaml', '.github/workflows/openapi-breaking.yml']`.
Blocking (no `continue-on-error`).

### Exit criteria
1. Workflow exists, path-filtered to the spec, blocking.
2. NEGATIVE: breaking fixture → job red (captured).
3. POSITIVE: identical/compatible spec → job green (captured).
4. On a real breaking change the PR author gets a named failure (documented in the workflow comment).

---

## F1.2 — nullable⇒required shape lint + redocly `struct`

### Part A — `SHAPE-NULLABLE-NOT-REQUIRED` api-lint rule
A new rule in `scripts/api-lint` (added to `RunSpecRules` in `spec_rules.go`, emitted as a
`Violation{Rule: "SHAPE-NULLABLE-NOT-REQUIRED"}`, blocking like every other api-lint rule per
`main.go`).

**Rule semantics (the 9f86828b guarantee):** for every object schema under `components.schemas`
(and inline body/response schemas that declare `properties`), a property that is **nullable** but
**not listed in that schema's `required` array** is a violation. "Nullable" means either:
- OpenAPI 3.0 style: `nullable: true` on the property schema, **or**
- OpenAPI 3.1 style: `type: ["<t>", "null"]` (a type array containing `"null"`).

Rationale: a nullable-but-optional field lets a code generator emit `field?: T | null`, so a
present-and-`null` value is indistinguishable from an absent key on the consumer — the exact wizard
bug (9f86828b). Requiring nullable fields to be in `required` forces the key to always be present
(value may be `null`), which is the AIP-134 field-behavior contract the review names.

- **Message shape:** `schema <Name> property "<prop>" is nullable but not in required (present-and-null drifts to optional)`.
- **Line anchor:** the property key node's line.
- **Scope note:** the rule walks declared `properties` maps; it does not flag free-form
  `additionalProperties`/`examples`. `allOf`/`$ref` composed `required` is honored if reasonably
  resolvable; where composition makes `required` non-local, the rule errs toward **not** false-
  positiving (documented in the rule comment) — a missed exotic case is a bounded defer, a false
  positive on the live spec is not acceptable (must be zero on the clean tree).

### AMENDMENT — response scope (operator-approved 2026-07-03, HS-7 cleared)

**Discovered runtime truth (during F1.2 impl):** the rule as first specified ("every object-schema
property") fires **60** genuine violations on the live spec — reality worse than the pre-code
assumption of a clean spec (same pattern as the struct "1 not 133" and F1.3 "all modules drift"
findings). Classification (runtime handlers read, agent `af455dac`): **50** are response-DTO fields
(present-and-null server-emitted → the real 9f86828b drift); **10** are request-body fields
(`createTemplate.reviewer_role`, `commitDocumentAutosave` req `page_count`,
`DocumentCommentCreateRequest.parent_library_id`, `TaxonomyProfile/AreaUpsertRequest.*` ×5,
`Create/UpdateTokenDictionaryEntryRequest.description`) — all confirmed full-replace/create with **no
null-vs-absent distinction** (no genuine PATCH clear-semantics).

**Operator fork #1:** reconcile the 60 → **full burn-down in M1** (not a grandfather allowlist).

**Operator fork #2:** the 10 request fields cannot be fixed non-breakingly — *both* remove-`nullable`
(`request-property-became-not-nullable`) *and* add-`required` (`request-property-became-required`) are
oasdiff-**breaking**, and F1.1's own new gate would red the M1 PR. Root cause: the invariant is a
**response** (generated-consumer-shape) concern; request bodies legitimately use optional+nullable for
upsert/PATCH. **Resolution (operator-approved): scope the rule to response-reachable schemas.** The
rule now exempts any schema reachable **only** from `requestBody` (transitive `$ref` closure) and any
inline schema lexically under a `requestBody:` ancestor; response-reachable schemas (incl. shared) are
still checked. No protection lost — the 9f86828b bug class is response-side.

**Burn-down applied:** the **50** response fields added to their schemas' `required:` arrays via
`openapi.yaml` + regen (non-breaking; oasdiff response-required-add = 0 errors). The **10** request
fields left as-is (nullable+optional), now exempt by scope. Live spec → **0 violations**; oasdiff
base→head → **exit 0 (no breaking)**.

**New finding surfaced (HS-6, NOT fixed in F1.2 — bounded defer):** `PUT /templates/{id}/approval-config`
(`upsertTemplateApprovalConfig`) has **no `requestBody` block** in the spec, yet the handler
(`routes_lifecycle.go:192-239`) decodes `reviewer_role`/`approver_role` from the body — an
undocumented-request-contract defect. Out of F1.2's nullable boundary (fix = add requestBody + regen →
changes the generated handler interface = delivery work). Owner: Leandro; trigger: contract-truth
hygiene micro-feature or fold to M9 governance. Reported at HS-1.

**Unit test:** `scripts/api-lint/spec_rules_test.go` (or a new `shape_rules_test.go`) with:
- a fixture schema `{nullable: true}` prop absent from `required` → **1 violation** (asserted rule
  name + message substring);
- the same prop **added to `required`** → **0 violations**;
- a non-nullable optional prop → **0 violations** (nullable is the trigger, not optionality per se).

### Part B — redocly `struct` re-enable + burn-down
- The live spec currently produces **exactly 1** `struct` error: `components.parameters` is an empty
  (`null`) node at `api/openapi/v1/openapi.yaml:4290`. Fix = remove the empty `parameters:` key (or
  populate it) via the spec, then regenerate if the byte output changes.
- Set `struct: error` in `redocly.yaml` (remove the `struct: off` line).
- **Burn-down record** (in `evidence.md`): starting suppressed count for `struct` = **1** (measured);
  ending = **0**. `operation-summary` and `security-defined` remain `off` **with a recorded owner +
  trigger** (they are separate hygiene rules, out of the 9f86828b bug class this milestone targets;
  keeping them off is not new debt introduced here — it is pre-existing, now documented). Owner:
  Leandro; trigger: pre-v1 spec-hygiene pass or first external API consumer.

### POSITIVE proof
- `go test ./scripts/api-lint/... -count=1` → PASS (incl. the new rule's tests).
- `go run ./scripts/api-lint -only SHAPE-NULLABLE-NOT-REQUIRED api/openapi/v1/openapi.yaml .` →
  `0 violation(s)`, exit 0 (the live spec is clean under the new rule — or every live violation was
  fixed via the spec and this is recorded).
- `redocly lint api/openapi/v1/openapi.yaml` with `struct: error` → **"Your API description is valid"**,
  exit 0.

### NEGATIVE proof (the fixtures)
- **Rule:** feed the rule a fixture spec (unit test **or** a temp file) with a `nullable: true`
  property omitted from `required` → the rule reports ≥1 `SHAPE-NULLABLE-NOT-REQUIRED` violation,
  non-zero exit. Captured.
- **struct:** temporarily re-introduce the empty `parameters:` node (or any struct violation) → redocly
  `struct: error` → red. Captured, reverted.

### CI wiring point
The new rule runs inside the **existing** `api-design-system-lint` job of `.github/workflows/api-contract.yml`
(`go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .`) — no new job needed; the rule is
picked up automatically because it is registered in `RunSpecRules`. The `struct` re-enable is enforced
by the existing `openapi-lint` job (`redocly lint`) in the same workflow (reads `redocly.yaml`).

### Exit criteria
1. New rule registered + unit-tested (pass-and-fail cases **incl. the response-scope case**: a
   nullable-not-required property is 0 violations under `requestBody`, 1 under `responses`).
2. Live spec clean under the rule (0 violations — via the 50 response add-required burn-down; the 10
   request fields exempt by response-scope) and clean under `struct: error`.
3. Both negatives proven red.
4. Burn-down recorded (nullable-not-required 60→0 = 50 response add-required + 10 request exempt;
   struct 1→0; operation-summary/security-defined owner+trigger).
5. oasdiff base→head over the burn-down = **exit 0, no breaking** (the burn-down must not break the API).
6. Missing-requestBody defer (upsertTemplateApprovalConfig) recorded with owner+trigger.

---

## F1.3 — contract-sync promoted to blocking CI (reconciled to runtime truth)

### The reconciliation (why it is not a symptom-patch)
`check-module-contract-sync.ps1` is stale against **runtime truth**, not wrong-about-a-real-drift:
- **OpenApiPatterns / FrontendTypesPatterns** hardcode absolute `/api/v1/<x>:` keys; the spec + the
  generated `index.d.ts` now use **relative** keys (`/<x>:`) since the AD-1 base-path migration
  (enforced by the `PATH-BASE-PREFIX` api-lint rule). → update these patterns to the relative form.
- **RuntimePatterns / RuntimeFile** for `documents` point at `module.go` and expect
  `documentsapi.HandlerWithOptions` there; the generated-boundary mount lives in the delivery/http
  file. → point the runtime-owner file(s) at the actual mount site; expect the actual mount token.
- **`OpenApiForbiddenPatterns`** for documents forbids `"\n  /documents:"` — which now matches the
  **correct** relative key, inverting the check. → drop/repair this obsolete forbid so it does not
  fire on the canonical relative key.
- **`Test-FrontendGeneratedTypeUsage`** regex excludes `paths[`/`components[` derived aliases but not
  `operations[` — so legitimate `type X = operations['...']` aliases are false-flagged as handwritten
  drift. → extend the allow-regex to recognize `operations[` and derived-from-generated aliases
  (`= <GeneratedAlias>[`, `NonNullable<... >`), so the check flags only genuinely-handwritten shapes.

**Symptom-patch test (C6):** after reconciliation the checker must **still catch real drift**. The
NEGATIVE proof below injects genuine drift and requires red. Loosening a pattern such that injected
drift passes is a FAIL. Reconciliation = align to current correct truth **while preserving detection
power**; it is not "edit the checker until it's quiet."

### Scope of the blocking gate
Gated modules (each reconciled to **zero live DRIFT**): **templates, documents, controlleddocuments,
taxonomy**. A CI wrapper iterates these and fails if any exits non-zero.

**Approval carve-out (bounded defer, recorded):** `approval` (`UsesGeneratedBoundary=$false`) reports
genuine ownership questions (`OpenAPI paths without runtime owners`, wrapper method gaps) that are
entangled with the **approval-module-promotion decision the mission assigns to M9 (§F9.5)** — the
review's Dimension 1 "documents/approval is structurally a 15th module" finding. Reconciling or
re-mounting approval here would cross into that boundary (**HS-2**). Approval is therefore **excluded
from the F1.3 blocking set**, with trigger: *M9 F9.5 (approval promotion / boundary decision) folds
approval into the contract-sync gate.* Owner: Leandro. This is recorded, not silent.

### POSITIVE proof
For each of {templates, documents, controlleddocuments, taxonomy}:
`./scripts/check-module-contract-sync.ps1 -Module <m>` → **no `[DRIFT]` lines**, exit 0. And the CI
wrapper (`scripts/check-contract-sync-all.ps1` or equivalent) over the four → exit 0 on the clean tree.

### NEGATIVE proof (the fixture)
Inject **genuine** drift on a throwaway change and show the wrapper red, e.g. one of:
- add a path key to `openapi.yaml` under a gated family with no runtime owner, **or**
- rename a wrapper function the checker asserts (e.g. remove `finalizeDocument` from `documents.ts`), **or**
- add a handwritten `interface Foo {}` (non-generated) to a gated feature's `api/*.ts`.

Wrapper → non-zero exit, output naming the drift + module. Captured; fixture reverted.

### CI wiring point
New job in `.github/workflows/api-contract.yml` (or `module-boundaries.yml`) named e.g.
`contract-sync`, running the wrapper on `windows-latest` **or** a `pwsh`-shelled step on
`ubuntu-latest` (PowerShell Core is available on GitHub ubuntu runners — `shell: pwsh`). Path filter:
`api/openapi/v1/openapi.yaml`, `internal/modules/**`, `frontend/apps/web/src/features/**`,
`frontend/apps/web/src/lib/api-types/**`, `scripts/check-module-contract-sync.ps1`,
`scripts/check-contract-sync-all.ps1`. Blocking.

### Exit criteria
1. Reconciled checker: zero live DRIFT for the four gated modules (captured per module).
2. CI wrapper + workflow job exist, path-filtered, blocking, `shell: pwsh`.
3. NEGATIVE: injected genuine drift → wrapper red (captured).
4. POSITIVE: clean tree → wrapper green (captured).
5. Approval carve-out recorded with trigger + owner. No other module silently dropped.
6. Detection power preserved (the negative proves the reconciled checker still catches real drift).

---

## F1.4 — ESLint feature-boundary rule + remove `Omit<>` overrides

### Part A — feature-boundary ESLint rule
A rule that forbids a file under `frontend/apps/web/src/features/<A>/**` from importing
`frontend/apps/web/src/features/<B>/**` (A ≠ B), i.e. no cross-feature imports. Shared code
(`src/shared`, `src/lib`, `src/store`, `src/queries`, cross-cutting) is **not** a feature and is
allowed. Implementation must be **zero-new-dependency** (extend the root `eslint.config.mjs` flat
config using built-in `no-restricted-imports` zones; do **not** add an eslint plugin — the FE pnpm
tree has junction drift and a `--frozen-lockfile` CI, so a new dep is a risk/HS-3 magnet). If a
zero-dep construction proves infeasible, that is HS-7 → surface, do not silently add a dep.

**Grandfathering (shrink-only allowlist):** existing cross-feature edges (~50+, widespread per the
review) are **not** fixed in this milestone (out of appetite; churn risk). They are grandfathered via
an **explicit, enumerable, shrink-only allowlist** — the established repo idiom (css-token-discipline,
test-discipline). The allowlist is either (a) inline `// eslint-disable-next-line` markers at each
existing edge (grep-countable) or (b) a data table in the config of allowed `(fromFeature →
toFeature)` pairs — decided in F1.4 `spec.md`; either way it is **finite, listed, and can only
shrink**. Count recorded in `evidence.md` with owner (Leandro) + trigger (*incremental de-coupling;
allowlist entries removed as cross-feature deps are refactored to shared modules; never added to*).

### Part B — remove `Omit<>` overrides
Delete the two hand-written `Omit<>` re-typings of generated types in
`features/templates/api/templates.ts`:
- `export type TemplateDTO = Omit<GeneratedTemplateDTO, ...> & {...}` (line ~36)
- `export type VersionDTO = Omit<GeneratedVersionDTO, ...> & {...}` (line ~44)

Replace each with the **generated type directly** (`export type TemplateDTO = components['schemas']['TemplateDTO']`
and likewise `VersionDTO`), **iff** the generated types already carry the correct nullability (post-M0
they should — M0's contract made `published_version` present-and-nullable and the nullable scalars
correct). **If** removing an `Omit<>` reveals a genuine generated-type nullability gap, that gap is a
**spec** defect → fix via `openapi.yaml` + regen (contract-first), **not** by keeping the `Omit<>`.
Keeping an `Omit<>` "because the generated type is wrong" is exactly the drift vector the review
names — so the resolution is always spec-side. If the spec fix is larger than M1's boundary, that is
HS-2 → surface. `WirePlaceholder` / `TemplateSchemas` (genuinely local view types, not overrides of a
generated schema) may remain; they are not `Omit<>`-overrides of generated types.

**Post-condition:** `grep -rnE "Omit<\s*(Generated|components\[)" src/features/**/api/*.ts` → **zero
hits**. `tsc --noEmit` clean.

### POSITIVE proof
- ESLint over `src/features` on the clean tree (allowlisted edges present) → **0 errors** from the
  boundary rule (`pnpm run lint`, or `npx eslint` if the pnpm tree blocks the script — recorded).
- `tsc --noEmit` (or `pnpm run build`'s tsc step) → clean after the `Omit<>` removal.
- `grep` for `Omit<>` overrides in `features/*/api/*.ts` → zero hits (captured).

### NEGATIVE proof (the fixture)
Add a throwaway import of another feature into a feature file **not** in the allowlist, e.g. in a temp
`src/features/documents/__boundary_probe.ts`: `import { X } from '../tokens/api/tokensTypes'` (a
from→to edge with no allowlist entry) → ESLint reports the boundary error (`no-restricted-imports`),
non-zero exit. Captured; probe deleted.

### CI wiring point
The **existing** `eslint` job in `.github/workflows/lint.yml` (`pnpm run lint`, reads the root
`eslint.config.mjs`). The new boundary config lives in `eslint.config.mjs`; the workflow already
path-filters `**/*.ts(x)` + `eslint.config.mjs`, so the gate is wired by editing the config — no new
workflow. Blocking (the job already fails the build on lint error).

### Exit criteria
1. Boundary rule in `eslint.config.mjs`, zero new deps.
2. NEGATIVE: synthetic new cross-feature import → eslint red (captured).
3. POSITIVE: clean tree (allowlisted edges) → eslint green (captured; or documented pnpm-tree defer
   with the gate still demonstrated via npx).
4. Zero `Omit<>` overrides of generated types in `features/*/api/*.ts`; `tsc --noEmit` clean.
5. Allowlist enumerated, count recorded, owner + shrink-only trigger stated.

---

## Cross-cutting definition of done for M1 (all must hold)

1. All four gates satisfy the D4 proof shape: POSITIVE (clean → green) **and** NEGATIVE (fixture →
   red), each captured verbatim in the feature's `evidence.md`.
2. Every gate is **blocking** and wired at the CI point named above; path filters correct.
3. `go build ./...` clean; `go test ./scripts/api-lint/... -count=1` green. Generated files are in
   sync with the spec: after the committed regeneration, a fresh `go generate ./...` + `pnpm run
   gen:api` produce **no further diff** (contract-first, zero hand-edits — the F1.2 burn-down's
   generated churn is the committed regeneration of the 50 response fields, not a hand-edit).
4. M0 regression intact: `redocly lint` clean, templates pin tests green, no shipped M0 shape changed.
5. Recorded bounded defers, each with a written trigger + owner: (a) F1.3 approval carve-out → M9
   F9.5; (b) F1.4 cross-feature allowlist (shrink-only); (c) F1.2 operation-summary/security-defined
   still off; (d) any local pnpm-tree block on `pnpm run lint`/vitest (gate still demonstrated);
   (e) F1.2 `upsertTemplateApprovalConfig` missing-requestBody contract defect → contract-hygiene
   micro-feature / M9. No defer beyond these without surfacing (HS-6).
6. All work committed locally (standing auth); **never pushed**; `docs/release/` never committed;
   plans dir never force-added.
7. `milestone-validator` verdict **PASS** written to `qa/milestone-qa.md`; then HS-1 operator gate —
   summary reported (including the F1.3 runtime-truth expansion), **no M2 until operator approval in a
   fresh session**.
