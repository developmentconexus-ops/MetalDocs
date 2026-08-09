# PASS 10 — API / Runtime Contract Architecture

**Date:** 2026-08-09
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273`
**Status:** reproduced-current (fresh local measurement in an isolated worktree)

Evidence-class labels: `reproduced-current`, `historical`, `stale` (see PASS 1 §1).

Root-cause tags used below (per audit convention): **#90/A3** default/general
architecture drift; **#89/A8** identity-specific; **#87/A1** enforcement
reachability (code exists but nothing calls it).

## 1. OpenAPI SSOT and codegen layout

- Single spec-of-record: `api/openapi/v1/openapi.yaml`, plus a separate
  `api/openapi/internal-e2e.yaml` for the test-seed surface. CLAUDE.md's
  "Contract-first" invariant: routes change only by editing this spec.
- **16 generated packages**, one per surface tag: 13 under
  `internal/modules/<module>/api/api.gen.go` (approval, audit, auth,
  controlleddocuments, distribution, documents, iam, notifications, search,
  security, taxonomy, templates, tokens) + 3 under `internal/platform/`
  (`featureflags/api`, `observability/api`, `observability/healthapi` — the
  last two were deliberately split into separate tags/publishers, per
  `observability/health.go`'s own comment, to keep "one tag, one
  publisher").
- Each generated package has its own `cfg.yaml` + a `go:generate` directive
  (confirmed, e.g. `internal/modules/approval/api/gen.go:3`):
  `go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../../../api/openapi/v1/openapi.yaml`.
  Tool: `oapi-codegen v2.7.0`, embedded-spec + strict-server + std-http-server
  generation modes (confirmed by generated-file header + `GetSwagger`
  presence in every `api.gen.go`).
- Embedded `swaggerSpec` (base64-gzip literal inside each `api.gen.go`) is
  regenerated wholesale on any spec edit — per user's standing memory
  ("OpenAPI embedded-spec regen churn... full regen canonical, partial =
  forbidden drift"), this is expected, not a bug.

## 2. Runtime request validation: confirmed unreachable

Claim under test: "GetSwagger only, unenforced constraints."

- `GetSwagger()` is **defined 16 times** — once per generated package
  (exact count, `grep -c "func GetSwagger"` across the 16 `api.gen.go`
  files = 16).
- `GetSwagger()` is **called zero times** outside its own defining file —
  confirmed by `grep -rn "GetSwagger()" internal apps | grep -v api.gen.go`
  returning no results. No router, no middleware, no handler ever invokes
  it.
- `openapi3filter`/`kin-openapi` (the schema-validation packages
  `oapi-codegen`'s generated code depends on for `GetSwagger`'s loader) are
  imported **only inside the 16 generated files themselves** — confirmed
  zero references to `openapi3filter.` anywhere else in `internal`/`apps`.
  No `openapi3filter.ValidateRequest` or equivalent call exists anywhere in
  the runtime path.

**Confirmed: the OpenAPI spec's request/response schema constraints (body
shape, field types, required-ness, enum values, format strings) are never
enforced by kin-openapi/openapi3filter at ingress.** Whatever validation
happens at runtime is either (a) `oapi-codegen`'s generated Go types
(compile-time shape only — a field either binds or the JSON decode fails,
but constraints like `minLength`, `pattern`, `enum` are not re-checked),
or (b) hand-written validation inside each handler (see `formval`,
`strictjson`, and the per-handler checks in §4/§5 below).

The specific "123 unenforced constraints" figure was **not independently
reproduced in this pass** — verifying it would require parsing the spec's
schema tree and counting `minLength`/`maxLength`/`pattern`/`enum`/`format`/
`required` nodes not mirrored by equivalent Go-side checks, which is a
distinct counting exercise from confirming validation is unreachable. The
*architecture* claim (validation code exists, is never invoked) is
confirmed with certainty; the *count* is unverified — report separately if
that number is load-bearing for a decision.

**Root cause: #87/A1 (enforcement reachability)** — the validation
machinery is present in the dependency graph (oapi-codegen pulls in
kin-openapi transitively) but was never wired into the request path. This
is not a missing-feature gap; it is a built-but-unwired gap, the defining
shape of an #87/A1 finding.

## 3. Hand-typed anonymous request structs in handlers

Claim: 11. **Reproduced exactly: 11.**

| # | File:line |
|---|---|
| 1 | `internal/modules/documents/delivery/http/handler.go:606` |
| 2 | `internal/modules/documents/delivery/http/handler.go:746` |
| 3 | `internal/modules/documents/delivery/http/handler.go:770` |
| 4 | `internal/modules/documents/delivery/http/handler.go:796` |
| 5 | `internal/modules/documents/delivery/http/handler.go:820` |
| 6 | `internal/modules/documents/delivery/http/handler.go:863` |
| 7 | `internal/modules/documents/delivery/http/handler.go:1032` |
| 8 | `internal/modules/documents/delivery/http/handler.go:1139` |
| 9 | `internal/modules/documents/delivery/http/handler.go:1177` |
| 10 | `internal/modules/templates/delivery/http/routes_autosave.go:68` |
| 11 | `internal/modules/templates/delivery/http/routes_schema.go:32` |

All 11 are `var req struct { ... }` blocks decoded via `readStrictJSON`/
`readJSON` instead of using an `oapi-codegen`-generated request-body type
from the corresponding module's `api.gen.go`. 9 of 11 sit in one file
(`documents/delivery/http/handler.go`), suggesting this module's HTTP
layer under-adopted the generated-type contract more than the others, not
an evenly-spread pattern.

**Root cause: #90/A3** (default/general drift — contract-first is the
rule, but nothing prevents a handler from reaching for an ad hoc struct
instead of the generated one; no lint currently forbids `var req struct`
in a `delivery/http` package).

## 4. Actor extraction: two `UserIDFromContext` helpers, opposite fail semantics

Two independently-defined helpers with the identical name, opposite
contracts:

- **`internal/modules/iam/domain/context.go`** — `UserIDFromContext(ctx) string`
  (fail-**open**: returns `""` silently if the context has no actor; no
  error, no bool).
- **`internal/platform/authn/context.go`** — `UserIDFromContext(ctx) (string, bool)`
  (fail-**closed**: `ok=false` if absent/blank, forcing the caller to
  branch). Its doc comment explicitly names the fail-open sibling and
  explains why this wrapper exists.

**Live call-site counts (fresh grep, excluding the two helpers' own doc
comments and internal self-reference):**

| Helper | Claim | Reproduced | Delta |
|---|---|---|---|
| `iamdomain.UserIDFromContext` (fail-open) | 17 live sites | **17** | exact match |
| `authn.UserIDFromContext` (fail-closed) | ~14/36 (bool discarded / branched) | **36** total call sites | count matches; the 14-vs-36 split (how many discard the `ok` bool vs branch on it) was not re-verified line-by-line in this pass — flag as unconfirmed sub-split, architecture-level 36 total is solid |
| `tenant.ActorFromContext` | 0 callers | **1 caller** — `internal/platform/db/runner.go:118` | **correction**: claim overstated; 1 real caller exists, but it is an internal infra chokepoint (the tx-runner GUC-seed path), not a business-logic call site — practically the claim's *intent* ("nothing in application/handler code uses this path") still holds, the literal "zero" does not |

The 17 fail-open sites span 8 modules/binaries: `approval` (4: `handler.go`
×2, `inbox_handler.go`, `submit_handler.go`), `documents` (2:
`fillin_handler.go`, `handler.go`), `iam` (2: `middleware.go`,
`presence/middleware.go`), `search` (1), `taxonomy` (1), `templates` (2),
`tokens` (3), plus `apps/api/cmd/metaldocs-api/main.go` (2). Each of these
is a live path where an absent/malformed actor silently becomes `""`
rather than a 401/500 — the actor then flows into authz checks, audit
writes, or SQL predicates as an empty string rather than failing the
request outright.

**Root cause: #89/A8 (identity)** — this is the canonical shape of an
identity-specific defect class: two competing "get the current actor"
primitives with inverted failure semantics and no lint/convention
enforcing which one a new handler should reach for. `authn`'s fail-closed
wrapper is clearly the intended target contract (its own comment names
the fail-open sibling as the thing to avoid), but 17 call sites across 8
modules never migrated.

## 5. Problem writer landscape (brief — see PASS 06-08 for full detail)

- Single canonical writer: `problem.Write(w, p *Problem)` at
  `internal/platform/problem/code.go:103`, RFC 9457 `application/problem+json`.
  35 call sites across `internal`/`apps` invoke `problem.Write`/`problem.New`
  as the terminal error path.
- From the ingress/contract angle specifically: the OpenAPI spec's
  `4xx`/`5xx` response schemas describe the problem+json envelope shape,
  but — consistent with §2 — nothing validates a handler's actual
  `problem.Write` call against the spec's declared response schema for
  that operation; the contract and the runtime writer are not
  cross-checked, they merely happen to agree by convention.
- Full inventory of writer variants/deviations is PASS 06-08's scope; not
  re-litigated here.

## 6. Pagination contracts

Claim: 4 limit/default policies, 2 envelope shapes. **Reproduced: 2
envelope shapes confirmed exactly; policy count is closer to 5 distinct
implementations than a clean 4** — reported honestly below rather than
forced to match.

### 6.1 Envelope shapes (2, confirmed)

| Shape | Spec schema | Example |
|---|---|---|
| Cursor envelope `{items, page:{next_cursor, has_more}}` | `CursorPage` (`api/openapi/v1/openapi.yaml:4953`, 7 `$ref` sites incl. `NotificationsListResponse:5043`) | notifications list, audit list |
| `data`/`meta` envelope `{data:{...:[]}, meta:{limit, offset[, total]}}` | `ListTemplatesResponse` (`:6405`), `ListTemplateAuditResponse` (`:6587`), `DocumentListResponse` (`:5459`, adds `total`) | templates, template audit, documents |

### 6.2 Limit/default policies (5 distinct implementations found)

| # | Location | Policy |
|---|---|---|
| 1 | `internal/platform/pagination/cursor.go` | canonical keyset cursor: `DefaultLimit=20`, `MaxLimit=100`, `ClampLimit()` — silently clamps, never rejects |
| 2 | `internal/modules/documents/delivery/http/handler.go:305-415` (two call sites) | reject-based: `limit<1` or `limit>100` → 400 problem+json |
| 3 | `internal/modules/templates/infrastructure/postgres.go:145-165` + `delivery/http/routes_query.go:30-50` | hybrid: `clampTemplatesLimit()` internally, but the HTTP layer separately rejects `limit>200` with 400; default 50, `templatesMaxLimit=200` |
| 4 | `internal/modules/auth/infrastructure/postgres/sessions_admin.go:15-35` | silent clamp: `limit<=0→50; limit>200→200` (no rejection path at all) |
| 5 | `internal/modules/iam/application/people_service.go:730-745` | silent clamp: `limit<=0\|\|limit>100 → limit=50` (different bounds than #4, same silent-clamp style) |

Two more call sites enforce **reject-only** bounds with no clamp fallback
and their own local defaults, structurally distinct from all 5 above:
`internal/modules/iam/delivery/http/sessions_handler.go:165-182`
(`defaultSessionsPageLimit`, reject `<1||>100`) and
`internal/modules/controlleddocuments/delivery/http/routes.go:790-800`
(`applyListLimit()`, reject `<1||>100`).

**Assessment:** the "4 policies" claim under-counts what is actually a
7-call-site spread of at least 5 meaningfully distinct limit/bound
strategies (canonical-clamp, reject-100, templates' clamp+200-reject
hybrid, two differently-bounded silent clamps, two differently-scoped
reject-onlys) sharing no single source of truth beyond
`platform/pagination` — and even that canonical package is bypassed by 6
of the ~8 call sites found. Whether the true count is "4" or "5+" depends
on how finely two silent-clamp variants with different numeric bounds are
counted as "the same policy" — reported as a range rather than forcing a
false-precision match.

**Root cause: #90/A3** — `platform/pagination` exists as the intended
canonical primitive (per its own doc comment: "shared by list endpoints...
so all list endpoints share one base64 dialect"), but adoption is
partial; each module independently reinvented bounds rather than calling
the shared `ClampLimit()`.

## 7. Idempotency: 4 independent implementations, compared

Claim: 4. **Reproduced: 4**, all keyed on the same `Idempotency-Key`
header name (contract-consistent at the wire level even though the
storage/replay mechanics diverge):

| # | Mechanism | Location | Header | Key storage | Replay semantics |
|---|---|---|---|---|---|
| 1 | Platform middleware (canonical) | `internal/platform/idempotency/middleware.go` (`Require()`), `postgres_store.go` | `Idempotency-Key` (validated UUID via `ValidateKey`) | Generic `Store` keyed on `(route, key, request-hash)` | Full two-phase claim/replay: `serveWithIdempotency` + `runClaimedHandler`; `responseRecorder` captures and replays byte-identical response; `WithStreamingOptOut` guards streaming incompatibility. Used by exactly 5 `approval` routes registered in `idempotentRoutes` (`internal/modules/approval/http/router.go`). |
| 2 | Approval signoff-family bespoke store | `internal/modules/approval/application/signoff_idemp.go` (`SignoffReplay`, `SignoffReplayCommitter`), `infrastructure/idempotency/postgres_signoff_idemp_store.go` (`PostgresSignoffIdempStore`) | `Idempotency-Key` (same header, self-managed instead of the platform middleware) | Bespoke Postgres table scoped to signoff outcomes; explicitly guards against a replayed key colliding with an unrelated stage's row | Same-key → same outcome AND same identifier; a failed attempt is NOT cached as a terminal replay (only successful outcomes replay), so a retry after failure re-attempts rather than replaying the failure. |
| 3 | Approval route-admin bespoke store, reused for fast-forward | `internal/modules/approval/application/route_admin_idemp.go` (`RouteAdminReplay`, `RouteAdminReplayCommitter`), `infrastructure/idempotency/postgres_route_admin_idemp_store.go` | `Idempotency-Key` | Same concrete `PostgresSignoffIdempStore`-family store as #2 — confirmed via `internal/modules/approval/http/handler.go`'s compile-time assertion `var _ fastForwardIdempStore = (*approvalidempinfra.PostgresSignoffIdempStore)(nil)`; `BeginFastForwardReplay` (`postgres_signoff_idemp_store.go:96`) is a second method on the *same* concrete store type, not a third store | Payload-hash computed over `(Idempotency-Key + profile + name + stages)`; explicit doc comment warns against reusing one key across differently-shaped route-admin calls (would "silently replay across them instead of surfacing" a conflict) |
| 4 | Submit DB-constraint backstop | `internal/modules/approval/application/submit_service.go:165-215` | N/A — not header-driven; enforced by a `UNIQUE(document_id, idempotency_key)` DB constraint | The constraint itself is the store | `ErrDuplicateSubmission`; explicitly documented as "a real DB-enforced replay backstop behind the HTTP idempotency middleware (F-D4)" — i.e., this is deliberately a second line of defense layered under mechanism #1 for submit specifically, not a peer replacement |

`approval/http/router.go`'s own doc comment names the design intent
explicitly: 5 routes use the platform middleware (#1); "five other
Idempotency-Key-bearing routes are deliberately excluded" because they are
self-managed by #2/#3/#4 instead. **This split is documented and
intentional, not an oversight** — but it still means 4 different replay
contracts exist for one header name, each with its own storage schema and
partial-failure semantics, entirely inside one module (`approval`); no
other module has needed a bespoke idempotency store.

**Root cause: #90/A3** (the platform middleware is the canonical
mechanism; `approval`'s 3 self-managed variants exist because its routes'
semantics — stage-signoff identity preservation, route-admin payload
shape checks, submit's DB backstop — needed replay behavior the generic
`Store` doesn't provide, which is a legitimate case for a module-local
extension, but the audit flags that 3 separate bespoke mechanisms
exist where 1 might have sufficed with a richer platform contract).

## 8. Concurrency / ETag / If-Match

`If-Match`/`ETag` usage is **scoped exclusively to the `approval`
module** — confirmed by grep, 19 files reference `If-Match`/`ETag`, every
one under `internal/modules/approval/{api,application,http}/` or
`internal/platform/problem/codes.go` (the shared 412-precondition-failed
problem code, generic and reused by any module that wants it — currently
only `approval` does). No other module (`documents`, `templates`,
`controlleddocuments`, etc.) uses `ETag`/`If-Match` at all; their
optimistic-concurrency needs, where present, are handled via a
body-embedded `expected_lock_version`/similar field instead of the
HTTP-standard header mechanism.

**Assessment:** this is an inconsistency, not a defect in `approval`
itself — `approval` chose the HTTP-native OCC mechanism, every other
module with a version-conflict concern chose a body-field mechanism, and
nothing in the contract layer (spec or platform) mandates one over the
other. **Root cause: #90/A3.**

## 9. Frontend consumption: generated client vs hand-written types

Claim: 18 hand vs 4 generated files. **Not cleanly reproducible as
stated** — the codebase has fewer genuinely generated files than claimed;
reported with full evidence below rather than forced to match.

**Generated files — confirmed exactly 2, not 4:**

| File | Generator | Evidence |
|---|---|---|
| `frontend/apps/web/src/lib/api-types/index.d.ts` (8,362 lines) | `openapi-typescript` v7.13.0, via `package.json`'s `"gen:api"` script against `api/openapi/v1/openapi.yaml` | Header: "This file was auto-generated by openapi-typescript. Do not make direct changes to the file." |
| `frontend/apps/web/src/lib/api/error-codes.generated.json` | `go run ./cmd/problem-codes-dump` from the backend's `problem.Register` catalog | Header: `"$comment": "Generated by go run ./cmd/problem-codes-dump from the problem.Register catalog. Do not edit by hand."` |

No `package.json` script or codegen config for 2 more generated files was
found anywhere under `frontend/apps/web` (searched for `*.generated.*`,
`*.gen.ts*`, and any `orval`/`codegen`-named config — none exist beyond
the two above). **The "4 generated" figure is not reproduced; either it
is stale/overstated, or it refers to a category (e.g. per-domain type
subsets, or files that merely *consume* generated types) not literally
"files a generator writes."**

**Hand-written files — 24-27 depending on inclusion rule**, all under
`features/*/api/*.ts` + `lib/api/*.ts` (excluding tests and the 2
generated files above):

- 26 files match `*.ts` under an `*api*` path (excl. tests, excl. the 2
  generated files).
- Of those, **13 hand-written files consume the generated `api-types`
  contract directly** (`import type { paths }`/`operations` from
  `../api-types`, or transitively via `lib/api/client.ts`'s
  `createClient<paths>` — confirmed at `lib/api/client.ts:147`) —
  spanning `approval`, `auth`, `controlled-documents`, `dashboard/audit`,
  `documents/{distribution,documents,library}`, `notifications`,
  `taxonomy`, `templates/{catalog,templates}`, `tokens/tokensTypes`, plus
  the client itself.
- **13 hand-written files do NOT import `api-types` at all** — these are
  infra/error-handling/cache glue rather than response-shape definitions:
  `approval/api/{etagCache,index,mutationClient}.ts`,
  `dashboard/api/dashboard.ts`, `documents/api/exports.ts`,
  `tokens/api/tokens.ts`, `lib/api/{authBus,errorMessages,errors,index,problem,resolveQueryError}.ts`,
  `lib/observability/apiTrace.ts`.

**Assessment:** the architecturally meaningful fact — reproducible with
certainty — is that the frontend has exactly **one** generated
request/response type surface (`api-types/index.d.ts`) which roughly half
of the hand-written API layer actually imports; the other half is
error/cache/observability plumbing that has no reason to import response
types at all. Whether a hand-written file that *does* import the
generated `paths`/`operations` types but then locally re-declares or
narrows a subset (a pattern present in files like `approvalTypes.ts`,
`tokensTypes.ts`, `controlledDocuments/types.ts`, `taxonomy/types.ts` —
confirmed by filename convention, not deep-read in this pass) should
count as "hand-written duplicating the contract" or "generated-type
consumer" was not resolved line-by-line for all 26 files; that finer
grained recount is a bounded follow-up, not completed here.

**Root cause: #90/A3** if the pattern is "some files re-derive
hand-written types instead of importing the generated ones" (ordinary
adoption drift); **not** an #89/A8 (identity) issue — this section is
purely a contract-type-consumption question, unrelated to actor
extraction.

## 10. Claimed-vs-reproduced summary (see final message for the combined table across both parts)

| Claim | Status |
|---|---|
| 4 platform packages import modules | confirmed exact |
| GetSwagger-only / unenforced validation | confirmed (architecture); "123" count unverified |
| 11 anonymous request structs | confirmed exact, file:line listed |
| 17 fail-open `UserIDFromContext` sites | confirmed exact |
| `tenant.ActorFromContext` zero callers | corrected — 1 caller, infra-internal |
| 4 pagination policies / 2 envelopes | envelopes confirmed exact; policy count is 5+ (under-stated) |
| 4 idempotency implementations | confirmed exact, compared |
| ETag/If-Match scoped to approval only | confirmed |
| 18 hand vs 4 generated frontend files | not reproduced as stated — 2 generated confirmed, hand-written count 24-27 depending on inclusion rule |
