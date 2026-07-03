# Feature F1.2 — nullable⇒required shape lint + redocly `struct` — Spec

> **Milestone:** 1 — Contract & frontend governance gates  ·  **Folder:** `f1.2-shape-lint`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (operator) — contract fully specified in
> `../validation-contract.md §F1.2`.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | Contract derived from mission.md §7 M1 (F1.2) + `validation-contract.md §F1.2`, grounded in the api-lint framework read at author time (`scripts/api-lint/spec_rules.go`, `main.go`). The 9f86828b bug class and the AIP-134 remedy are named in the review. Nothing guessed. |
| 2 | Custom redocly rule or api-lint Go rule? | api-lint Go rule. It is the proven blocking-lint framework (`RunSpecRules`, `Violation{}`), already CI-wired via the `api-design-system-lint` job; a new rule is picked up automatically and is blocking by construction. A bespoke redocly plugin would fork enforcement into a second, weaker mechanism. |
| 3 | struct burn-down size? | Measured: `struct: error` yields exactly **1** live error today (empty `components.parameters` null node, openapi.yaml:4290), not the 133 the stale comment claims. Fix the node, enable struct, record 1→0. |
| 4 | operation-summary / security-defined? | Out of the 9f86828b bug class; pre-existing `off`. Keep off but record owner (Leandro) + trigger (pre-v1 hygiene / first external consumer). Not new debt introduced here. |

## Consumer contract (FIRST)

- **Consumer(s):** the api-lint CI job (`api-design-system-lint` in `api-contract.yml`); the redocly
  `openapi-lint` job; every future spec author (the rule must fail their PR if they add a nullable-
  not-required field — the 9f86828b class); the generated TS/Go clients (which must never again
  receive an optional-yet-nullable field that drifts).
- **Contract:**
  1. `RunSpecRules` emits `Violation{Rule:"SHAPE-NULLABLE-NOT-REQUIRED", File, Line, Message}` for
     every declared object-schema property that is nullable (3.0 `nullable:true` OR 3.1 `type:[...,"null"]`)
     and absent from that schema's `required` array. Blocking (non-zero exit) like all api-lint rules.
  2. `redocly.yaml` sets `struct: error`; `redocly lint` is clean on the live spec.
- **Source of truth for the contract:** `validation-contract.md §F1.2`; `scripts/api-lint/spec_rules.go`
  (rule registration + `mapGet`/`derefSchema` helpers to reuse); `redocly.yaml`.

## What this feature implements

- A new `SHAPE-NULLABLE-NOT-REQUIRED` rule in `scripts/api-lint` (walk `components.schemas` + inline
  `properties` maps; honor `required`; reuse existing yaml-node helpers), with unit tests (pass + fail
  cases) in `scripts/api-lint`.
- Fix the 1 live `struct` error via `api/openapi/v1/openapi.yaml` (remove/populate the empty
  `parameters:` node), regenerate if byte output changes (`go generate ./...`, `pnpm run gen:api`),
  and set `struct: error` in `redocly.yaml`.

## Non-goals (mandatory)

- No `operation-summary` / `security-defined` re-enable (recorded defer).
- No sweep of the whole spec for other shape rules (AIP-fuzzing is post-v1, mission §2 Non-Goals).
- No changes to any module's Go handlers or FE consumers.
- No hand-edits to generated files (`*.api.gen.go`, `index.d.ts`) — only via regen.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Rule flags nullable-not-required | api-lint unit test: fixture prop `{nullable:true}` absent from `required` → 1 `SHAPE-NULLABLE-NOT-REQUIRED` violation | fixture (unit test) |
| Rule clears when required-added | same fixture, prop added to `required` → 0 violations | fixture (unit test) |
| Live spec clean under rule | `go run ./scripts/api-lint -only SHAPE-NULLABLE-NOT-REQUIRED api/openapi/v1/openapi.yaml .` → `0 violation(s)` exit 0 | real |
| api-lint suite green | `go test ./scripts/api-lint/... -count=1` → PASS | real |
| struct fixed + enabled | `redocly lint api/openapi/v1/openapi.yaml` (struct:error) → "valid", exit 0 | real |
| struct negative | re-introduce empty `parameters:` → redocly red | fixture (reverted) |
| No codegen drift | `go generate ./...` + `pnpm run gen:api` → `git diff` empty | real |

> TDD: write the failing unit test for the rule first (red), implement the rule (green), then fix the
> struct node.

## ADR needed?

- [x] No durable architecture decision — enforcement tooling for an existing contract-first decision.
  Skip. (The rule comment cites AIP-134 / the 9f86828b incident inline.)
