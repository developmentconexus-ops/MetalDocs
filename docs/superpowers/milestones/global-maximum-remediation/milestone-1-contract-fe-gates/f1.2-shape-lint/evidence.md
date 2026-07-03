# Feature F1.2 — nullable⇒required shape lint + redocly `struct` — Evidence

> **Status:** CLOSED. Rule shipped blocking; live spec 0 violations via a full 60→0 burn-down;
> struct re-enabled; both negatives proven; oasdiff confirms the burn-down is non-breaking.
> **Contract:** `../validation-contract.md §F1.2` (incl. the operator-approved response-scope
> AMENDMENT). **Implemented by:** subagents (sonnet); **reviewed + verified by:** main session.

## Summary of what shipped

- **New api-lint rule `SHAPE-NULLABLE-NOT-REQUIRED`** (`scripts/api-lint/shape_rules.go`, wired in
  `spec_rules.go` `RunSpecRules`, blocking by construction). Flags a nullable property (3.0
  `nullable:true` / 3.1 `type:[...,"null"]`) absent from its schema's `required`, on
  **response-reachable schemas only** (see scope below). Message: `schema <Name> property "<prop>" is
  nullable but not in required (present-and-null drifts to optional)`.
- **60→0 burn-down** of live-spec violations: **50** response-DTO fields added to their `required:`
  arrays via `api/openapi/v1/openapi.yaml` + regen; **10** request-body fields left nullable+optional
  and exempted by response-scope.
- **redocly `struct` re-enabled**: removed the empty `components.parameters` node; `struct: off →
  error` in `redocly.yaml`.

## Discovered runtime truth + operator decisions (HS-7 / HS-6)

The pre-code contract assumed the live spec was clean under the rule. It was not — **60 genuine
violations**. Two operator forks, both recorded in `../validation-contract.md §F1.2 AMENDMENT`:
1. **Full burn-down in M1** (not a grandfather allowlist).
2. **Scope the rule to response schemas** — because the 10 request-body violations cannot be fixed
   non-breakingly (both remove-`nullable` and add-`required` are oasdiff-breaking) and the invariant
   is fundamentally a response/generated-consumer concern. Request-field semantics were classified
   against the runtime handlers (agent `af455dac`): all 10 are full-replace/create with no
   null-vs-absent distinction, so exempting them loses no real capability.

**New defect surfaced, NOT fixed here (bounded defer → HS-1):** `PUT /templates/{id}/approval-config`
(`upsertTemplateApprovalConfig`) has **no `requestBody`** in the spec though the handler reads
`reviewer_role`/`approver_role` from the body. Out of the nullable boundary; owner Leandro; trigger:
contract-hygiene micro-feature or M9.

## Rule scope (response-reachable only)

`reachableSchemas` computes the transitive `$ref` closure from every `paths.*.*.requestBody` (`reqReach`)
and every `paths.*.*.responses` (`respReach`); `requestOnly = reqReach \ respReach`. The walk skips
(a) any `components.schemas.<Name>` in `requestOnly` and (b) inline schemas lexically under a
`requestBody:` ancestor. Response-reachable schemas (incl. any shared) stay checked — response
integrity wins (documented in the rule comment; no such shared schema exists in the live spec today).

## Validation Gate — proof (verified by main session on a clean build)

| Criterion | Proof command | Result |
|-----------|---------------|--------|
| Rule flags nullable-not-required (response) | unit `TestShapeNullableNotRequired/nullable_not_required_bites` | RED→GREEN, 1 violation w/ correct message |
| Rule clears when required-added | `.../nullable_and_required_clean` | 0 violations |
| Non-nullable optional not flagged | `.../optional_not_nullable_clean` | 0 violations |
| **Response-scope** (NEW) | `.../request_only_exempt_response_still_checked` — same nullable-not-required prop: 0 under requestBody (inline + `*Request` component), 1 under responses | PASS |
| Live spec clean under rule | `go run ./scripts/api-lint -only SHAPE-NULLABLE-NOT-REQUIRED api/openapi/v1/openapi.yaml .` | **`0 violation(s)`, exit 0** |
| api-lint suite green | `go test ./scripts/api-lint/... -count=1` | **`ok metaldocs/scripts/api-lint`** PASS |
| struct enabled + spec valid | `redocly lint api/openapi/v1/openapi.yaml` (`struct: error`) | **"Your API description is valid" exit 0** |
| struct NEGATIVE | re-add empty `parameters:` → redocly | **red**: `Expected type NamedParameters (object) but got null … struct rule` exit 1 (reverted) |
| Burn-down non-breaking | `oasdiff breaking <HEAD spec> <head spec> --fail-on ERR` | **`No breaking changes to report`, exit 0** (the earlier 15 `request-property-became-not-nullable` errors gone after scope+revert) |
| Build clean | `go build ./...` | exit 0 |
| Generated in-sync | `go generate ./...` + `pnpm run gen:api` regenerated; committed | 7 `*.gen.go` + `index.d.ts` reflect the 50 response fields; fresh regen = no further diff |

### Burn-down record
- `SHAPE-NULLABLE-NOT-REQUIRED`: **60 → 0** (50 response add-required, 10 request exempt by scope).
- `struct`: **1 → 0** (empty `components.parameters` removed; `struct: error` active).
- `operation-summary`, `security-defined`: remain **off** — pre-existing, out of the 9f86828b bug
  class. Owner: Leandro; trigger: pre-v1 spec-hygiene pass / first external API consumer.

## CI wiring point (no new workflow)

- Rule: auto-picked-up by the existing `api-design-system-lint` job in
  `.github/workflows/api-contract.yml` (`go run ./scripts/api-lint/ -strict …`) — blocking because
  every api-lint rule is blocking.
- struct: enforced by the existing `openapi-lint` (redocly) job in the same workflow via `redocly.yaml`.

## Files changed

- `api/openapi/v1/openapi.yaml` — 50 response `required` additions; empty `parameters:` removed.
- `scripts/api-lint/shape_rules.go` (new), `shape_rules_test.go` (new), `spec_rules.go` (rule wired).
- `scripts/api-lint/testdata/` — 4 new fixtures (incl. `shape_request_scope_good.openapi.yaml`) +
  `Page.required` fix in 3 shared fixtures (the shared `Page` schema declared a nullable
  `next_cursor` with no `required`, which the wired rule correctly flags; the fix mirrors the real
  spec's `CursorPage`).
- `redocly.yaml` — `struct: error`.
- Generated (regenerated, not hand-edited): 7 `internal/modules/*/api/api.gen.go`,
  `frontend/apps/web/src/lib/api-types/index.d.ts`.

## Bounded defers

- `operation-summary` / `security-defined` redocly rules still off (owner/trigger above).
- `upsertTemplateApprovalConfig` missing-requestBody contract defect (owner/trigger above).

## Review disposition

Rule TDD'd (red→green incl. the response-scope case); live spec 0; struct 1→0 with negative;
**oasdiff exit 0 proves the burn-down is non-breaking** — the decisive check that the response-scope
resolution was correct. Diff inspected: 50 additions land in the correct schemas, no stray edits.
**APPROVED** — committed.
