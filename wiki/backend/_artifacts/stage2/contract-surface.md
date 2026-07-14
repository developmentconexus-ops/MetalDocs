# Stage-2 Evaluation — Contract Surface

> **Theme:** API contract surface
> **Findings evaluated:** F-03 (parallel spec2/api/v2 surface), F-13 (schema coverage gaps and camelCase drift), F-09 (idempotency inline reimplementation and raw codes)
> **Evaluator:** Stage-2 sub-agent, 2026-06-11
> **Standards applied:** OpenAPI Specification 3.x, RFC 9457 (Problem Details for HTTP APIs), Stripe API design guide (idempotency), Google AIP-193/AIP-004 (API versioning and errors), OWASP API Security Top 10 2023 (API9:2023 Improper Inventory Management)
> **Normative requirements mapped:** RF-4, RF-5, RF-10, REQ-API-2, REQ-API-5, REQ-H-2
> **Context:** The api-contract-hardening program ran Phases A–F through 2026-06-08 and closed with a 4-dimension re-audit at 0 CRITICAL/0 HIGH. Evaluations here acknowledge that context — they do not re-litigate what the program has already closed. Verdicts are grounded in what the code **currently** shows.

---

## How to read this document

For each finding:
1. **Current state** — re-confirmed against the code anchors in the register (files/lines checked 2026-06-11).
2. **Applicable standard** — the precise spec section, RFC paragraph, or published guide that defines the bar.
3. **Verdict + rationale** — one of KEEP / SIMPLIFY / REFACTOR / DELETE / DEFER with evidence-based argument.
4. **Smallest correct fix** — minimum scope that actually reaches the professional bar; any larger proposal is explicitly called out as over-engineering.
5. **Effort / blast-radius** — implementation size and change surface.
6. **ADR needed?** — yes/no with reason.
7. **Over-engineering check** — is there a simpler path than the obvious fix?

---

## F-03 — Parallel Contract Surface (spec2.yaml / internal/api/v2)

### Current state (re-confirmed 2026-06-11)

**spec2.yaml** (`api/openapi/spec2.yaml`):
- 1 061 lines; 13 approval state-machine routes with paths rooted at `/api/v1` (`servers.url: /api/v1` confirmed at line 30).
- Error schema is `{error:{code,message,details},request_id}` at lines 596–606 — a bespoke envelope that predates RFC 9457.
- No global or per-operation `security:` declarations.
- No `//go:generate` or `cfg.yaml` references it. api-lint does not run against it. It is linted by nothing.
- The commit that last touched it is described as "purge v2 routes" (e1944bc4a) — the file was not deleted, only the routes it referenced were removed from whatever consumed it at that time.
- The 13 routes it declares overlap in HTTP path with the live v1 `approval` tag operations that are fully served and code-generated from `openapi.yaml`.

**internal/api/v2/types_gen.go** (`internal/api/v2/types_gen.go`):
- 60 lines; hand-maintained (no `//go:generate` directive at line 1); `_gen.go` suffix is a misnomer that causes api-lint's pagination-codec check (`code_rules.go:97`) and any other `*.gen.go`-skipping tool to silently exempt this hand-maintained file from checks.
- Defines `apiv2.APIError{Code, Message, Details, TraceID}` — a flat shape that differs structurally from the canonical `Problem{title, status, code, detail, instance, errors}`.
- Consumed only by three contract test files (`routes_memberships_contract_test.go:109–114`, `routes_profiles_contract_test.go`, `routes_contract_test.go`). Tests decode via tolerant `json.Unmarshal`; the shape mismatch is invisible to passing tests.
- Also defines `ProfileResponse`, `AreaResponse`, `ControlledDocumentResponse`, `MembershipResponse` — all with snake_case `json:` tags. These shapes overlap with generated types in the iamapi/controlleddocumentsapi packages but are structurally separate.

**Capability-catalog CI gate** (`ops/CAPABILITY_CATALOG.sha256`, `.github/workflows/invariants.yml:89–118`):
- `CAPABILITY_CATALOG.sha256` contains the literal string `placeholder-hash-update-after-catalog-created` (confirmed).
- The CI workflow resolves path `sql/seeds/capabilities_v2.sql` which does not exist in the repository. The `[ ! -f "$CATALOG" ]` guard exits 0 with "Catalog file not found — skip". The gate is therefore a no-op in every CI run.

### Applicable standard

**REQ-API-2** (backend-target-architecture.md §5): "One API surface. The spec2.yaml / internal/api/v2 parallel surface is converged into v1 or formally fenced with an ADR and a sunset plan. No third option. (MUST — refactor item RF-4.)"

**OpenAPI Specification 3.0.3 §4.1** (https://spec.openapis.org/oas/v3.0.3#openapi-document): an OpenAPI document "describes" the API it governs; a published spec file implies the operations it declares are served. Having a spec file describing 13 operations that are already documented elsewhere with an incompatible error schema creates a dual-spec state that violates the "single source of truth" principle RF-4 requires.

**OWASP API Security Top 10 2023, API9:2023 — Improper Inventory Management** (https://owasp.org/API-Security/editions/2023/en/0xa9-improper-assets-management/): "Undocumented APIs or APIs created with older versions of an API definition … can bypass security controls." A spec file that defines real API paths but is not linted, not referenced by codegen, and not covered by authz checks is an inventory gap regardless of whether the routes themselves are served.

**RFC 9457 §3** (https://www.rfc-editor.org/rfc/rfc9457#section-3): defines the `application/problem+json` error type. The spec2.yaml `ErrorResponse{error{code,message,details},request_id}` envelope is not `application/problem+json` and not conformant. The v1 contract is now fully RFC 9457 (Phase D outcome); spec2.yaml contradicts it.

For `internal/api/v2/types_gen.go`, the `_gen.go` misnomer is a **Go toolchain convention issue**: `go generate`, `gomod`, and the Go doc tool treat `_gen.go`-named files as machine-generated (https://go.dev/blog/generate — "by convention, generated files have a `_generated.go` or `_gen.go` suffix"). Tools that skip generated files will skip this hand-maintained package, silently defeating any rule that scans non-generated code.

### Verdict: REFACTOR → DELETE both artifacts; fix the CI gate

**Rationale:**

1. **spec2.yaml** has no active consumers (no codegen, no linting). Its 13 routes duplicate live v1 operations with an incompatible error schema and no security declarations. It exists as a documentation artefact of a migration that was never completed. Retaining it violates REQ-API-2 explicitly. The "fence with ADR + sunset" option requires affirmative justification (a consumer that needs the v2 shape); no such consumer exists. The only correct action is deletion.

2. **internal/api/v2/types_gen.go** is consumed by exactly three contract test files, all of which decode into `apiv2.APIError` via tolerant `json.Unmarshal`. Those tests pass while testing nothing about the actual Problem contract, because `json.Unmarshal` fills in zero values for missing fields. The tests should be migrated to assert against the `problem.Problem` type directly (or against the generated `iamapi` response types). Once migrated, `internal/api/v2/` has zero importers and can be deleted. The `apiv2.ProfileResponse`, `AreaResponse`, `ControlledDocumentResponse`, and `MembershipResponse` types have equivalents in the generated packages — they are not needed as a separate package.

3. **CI gate** (`capability-catalog-hash`): the gate is currently silently non-functional. The seed file it references (`sql/seeds/capabilities_v2.sql`) does not exist; the gate exits 0. Two options: (a) delete the gate entirely since `api-lint`'s `checkSeedRegistryParity` rule already enforces `db/reference-data/0001_product_reference_data.sql` ↔ domain capability registry parity (confirmed: `registry_rules.go`), making this gate a redundant and broken second check; (b) fix the path to `db/reference-data/0001_product_reference_data.sql` and regenerate the SHA256 pin. Option (a) is the simpler, non-redundant fix — the api-lint rule already covers the invariant; this gate adds no independent protection.

**Anti-over-engineering check:** The register implies a "converge or fence" choice. Fencing with an ADR would require maintaining a separate linting and codegen pipeline for spec2.yaml — more work and more surface to drift. Convergence is already done (v1 covers all 13 routes with RFC 9457 errors). DELETE is the simplest correct outcome; it does not require designing anything new.

### Smallest correct fix

1. Delete `api/openapi/spec2.yaml`.
2. Migrate the three contract test files (`routes_memberships_contract_test.go:109–114`, `routes_profiles_contract_test.go`, `routes_contract_test.go`) to assert against `problem.Problem` or the generated iamapi response types. Then delete `internal/api/v2/`.
3. Delete the `capability-catalog-hash` CI job from `invariants.yml` (redundant with api-lint's `checkSeedRegistryParity`), or fix the path + regenerate the pin. Either closes the non-functional gate.

Steps 1 and 3 are independent; step 2 requires migrating 3 test files first.

### Effort / blast-radius

| Step | Effort | Blast radius |
|---|---|---|
| Delete spec2.yaml | S (1 file, git rm) | contained — no codegen or linter references it |
| Migrate 3 contract tests + delete api/v2/ | S (3 test files, ~25 lines each) | contained — test-only |
| Fix/delete the CI gate | S (1 workflow file) | contained — CI only |

**Total: S, contained.**

### ADR needed?

No. REQ-API-2 already mandates "one API surface" and the "fence" option requires a justification (a live consumer in a different shape) that does not exist. Deleting a dead artefact does not warrant an ADR. The program's existing ADR 0025 (error-envelope-rfc9457) makes the intent clear.

---

## F-13 — Missing or Split OpenAPI Schema Coverage

### Current state (re-confirmed 2026-06-11)

**SearchDocumentItem schema gap** (`api/openapi/v1/openapi.yaml:4785–4820`):
- The spec declares `SearchDocumentItem` with 13 fields: `document_id`, `title`, `document_type`, `document_profile`, `document_family`, `document_sequence`, `document_code`, `process_area`, `owner_id`, `department`, `status`, `effective_at`, `expiry_at`, `created_at`.
- The Go handler's `SearchDocumentResponse` struct (`internal/modules/search/delivery/http/handler.go:24–43`) emits four additional fields: `subject` (line 33), `business_unit` (line 35), `classification` (line 37), `tags` (line 39).
- However, `internal/modules/search/infrastructure/v2documents/reader.go` (Phase B) documents at lines 22–27 that `subject`, `business_unit`, `classification`, `tags` exist on the legacy `metaldocs.documents` schema which was decommissioned — the v2 reader never populates them. These four fields are always zero-valued (`""` or `nil`) on every response. The schema divergence is between a spec that omits them and a response struct that declares them — but at runtime the fields emit as empty/omitted because they are never populated by the query.

**`businessUnit` camelCase query parameter** (`internal/modules/search/delivery/http/handler.go:99`):
- The handler reads `r.URL.Query().Get("businessUnit")` at line 99. This is camelCase; all other query parameters in the handler and the spec use snake_case.
- The `businessUnit` parameter is not declared in the spec at all (search route spec at lines 807–868 declares `q`, `owner_id`, `document_type`, `document_profile`, `document_family`, `process_area`, `department`, `status`, `expiry_before`, `expiry_after`, `limit` — no `businessUnit` or `business_unit`).
- Even if the parameter were declared, it would be passed to the reader where it becomes a no-op (same decommissioned column reason as above).

**Partial files camelCase drift** (`api/openapi/v1/partials/`):
- Three files: `controlled-documents.yaml` (306 lines), `documents.yaml` (247 lines), `templates.yaml` (108 lines).
- All confirmed dead from pipeline perspective (not referenced by any `cfg.yaml`, `//go:generate` directive, or `openapi.yaml` `$ref`). Canonical spec codegen reads `openapi.yaml` directly.
- `controlled-documents.yaml:207–229` uses camelCase schema properties (`tenantId`, `profileCode`, `ownerUserId`, `areaCodes`, `userIds`). The canonical spec uses snake_case (`area_codes`, `user_ids` at lines 5022–5028).
- The PATH-BASE-PREFIX and CASING-DRIFT lint rules do not run against partial files.
- These files are not consumed and have zero impact on the served contract. The only risk is if a future developer incorrectly references them.

**`POST /iam/users` undeclared deprecation** (`api/openapi/v1/openapi.yaml:213–224`):
- The operation `createManagedUser` has a Portuguese-language `summary` that reads "Cria usuario interno com credencial local inicial (legacy; prefer POST /iam/users/invite)" — human-readable deprecation signal only.
- No `deprecated: true` field at the operation level. `deprecated: true` is a first-class OpenAPI 3.x field.
- The `iamapi` codegen package (`api.gen.go`) has no machine-readable deprecation annotation on the generated method.

**iamapi codegen bundles three domain concerns** (`internal/modules/iam/api/cfg.yaml:9–11`):
- `include-tags: [iam, audit, security]` produces a single 5 691-line `api.gen.go`. Audit and security are distinct domain concerns; their types, `ServerInterface` methods, and embedded spec slice are co-generated into the IAM package.

### Applicable standard

**OpenAPI Specification 3.0.3 §4.8.20 — Schema Object**: declares that a schema documents "the data type" of a value transmitted in or from an API. Fields emitted on the wire that are absent from the schema create an undocumented contract (consumers cannot rely on them; code generators omit them from client types). This is the "schema drift" class that ENVELOPE-DRIFT and the codegen-drift CI gate guard against.

**OpenAPI Specification 3.0.3 §4.8.11.6 — Operation Object, `deprecated`**: "Declares this operation to be deprecated. Consumers SHOULD refrain from usage of the declared operation." The field is a boolean, not documentation prose. SDK generators, API portals, and internal tooling (e.g. openapi-typescript) surface this field to developers.

**OWASP API Security Top 10 2023, API9:2023**: undocumented parameters accepted by handlers but absent from the spec are inventory gaps. A parameter that is accepted, forwarded to a domain query, and silently no-ops is both an inventory gap and a user-experience defect (a consumer who passes `businessUnit` receives no error and no results filtered by it).

**Google AIP-004** (https://google.aip.dev/004): recommends one API version surface; splitting generation concerns by domain enables independent evolution. Bundling three domains into one generated file is not forbidden by OpenAPI or Go tooling, but it couples regeneration of audit and security contracts to every IAM change.

**REQ-API-4** (backend-target-architecture.md §5): "Conventions are uniform: snake_case params." A camelCase `businessUnit` query parameter violates this requirement.

### Verdict (per sub-finding)

**F-13a — SearchDocumentItem four missing fields (`subject`, `business_unit`, `classification`, `tags`):** SIMPLIFY

The fields are declared in the Go struct but never populated. The correct fix is to remove them from the `SearchDocumentResponse` Go struct entirely (they are dead fields for the same reason the Phase B fix removed the reader query for those columns). Once removed from the struct, the spec already matches reality — no spec change needed. Alternatively, if the business domain is expected to add those fields to `public.documents` in a future phase, add them to the spec with `nullable: true` so the contract leads the implementation. The former (remove dead struct fields) is the minimal fix now; the latter is a product decision.

**Over-engineering check:** Do NOT add the four fields to the spec while they are always zero-valued. Adding dead fields to a public contract is misleading. Remove them from the Go struct; spec is already correct.

**F-13b — `businessUnit` camelCase undocumented query parameter:** REFACTOR

Rename the query param read in the handler from `"businessUnit"` to `"business_unit"` and add it to the spec (even if the filter is a no-op at the SQL layer, the parameter is accepted by the handler). The alternative is to remove the parameter from the handler since it silently no-ops — which is likely the cleaner option until the data model supports it. Removing it from the handler and documenting the absence is preferable to declaring an accepted-but-ineffective parameter in the spec.

**Over-engineering check:** This is a 1-line handler change. Do not introduce a query-param normalization layer for it.

**F-13c — Partial files camelCase drift:** DELETE

The three partial files are confirmed dead from the pipeline perspective. They have no codegen consumers, no spec `$ref`s, and no linting coverage. They represent a schema representation of the API as it existed before the contract-hardening program, with casing and path conventions that directly contradict the current spec. Retaining them is an OWASP API9 inventory risk (future developers may mistake them for authoritative) and a maintenance hazard. Delete all three.

**Over-engineering check:** Do NOT add partial-file linting to cover them instead of deleting them. They are dead artefacts.

**F-13d — `POST /iam/users` undeclared deprecation:** SIMPLIFY

Add `deprecated: true` to the `createManagedUser` operation in `openapi.yaml`. This is a 1-line spec change. Regenerate all six `api.gen.go` packages (oapi-codegen reflects `deprecated: true` in generated code as a Go comment `// Deprecated:`). Regenerate FE types (`npm run gen:api`). This satisfies the OpenAPI 3.0.3 §4.8.11.6 requirement for machine-readable deprecation signaling.

**Over-engineering check:** Do not version the endpoint or create a migration plan; this is a human-intent signal already present in the summary text, being promoted to a machine-readable field.

**F-13e — iamapi codegen bundles three domain concerns:** DEFER

Splitting `internal/modules/iam/api/cfg.yaml` into three separate configs (`iam`, `audit`, `security`) would produce three separate generated packages, three `api.gen.go` files, and require updating 14 import sites across delivery layers. The current bundled file compiles, passes CI, and has no correctness defect. The blast-radius is medium (six files to create, multiple delivery files to re-import). The justification for splitting — independent evolution of audit/security contract — is valid but not urgent. Trigger: when the audit or security tag acquires enough new operations or schema types that the 5 691-line generated file becomes a merge-conflict bottleneck, or when the modules are reorganized.

**Over-engineering check:** Splitting now for cleanliness when there is no active conflict or correctness bug is over-engineering. DEFER.

### Summary verdict table

| Sub-finding | Verdict | Effort | Blast-radius |
|---|---|---|---|
| F-13a: four dead search struct fields | SIMPLIFY (remove from Go struct) | S | contained (search module only) |
| F-13b: `businessUnit` camelCase param | REFACTOR (remove or rename + spec) | S | contained (search handler + spec) |
| F-13c: partial files dead with camelCase | DELETE | S | contained (3 files, git rm) |
| F-13d: `POST /iam/users` no `deprecated: true` | SIMPLIFY (add 1-line spec field) | S | contained (spec + regen) |
| F-13e: iamapi three-domain bundle | DEFER | M | module (6+ files, delivery rewires) |

### ADR needed?

No for F-13a–d (all are small conformance fixes to a documented standard). No for F-13e (the deferral trigger is explicit — no ambiguity that requires an ADR-level decision to defer).

---

## F-09 — Idempotency Inline Re-Implementation (Diverges from Middleware)

### Current state (re-confirmed 2026-06-11)

**Finalize handler inline idempotency** (`internal/modules/documents/delivery/http/handler.go:440–499`):
- Confirmed present and substantively unchanged from the register description. The handler reads the `Idempotency-Key` header at line 440, validates it at 445 via `idempotency.IsValidKey`, hashes the body at 450 via `idempotency.RequestHash`, manually calls `idempStore.BeginReplay` at 474, manually defers `FailReplay` at 494, and manually calls `CompleteReplay` later in the handler body.
- This is a correct, fully-formed inline implementation — it calls `IsValidKey`, `RequestHash`, `BeginReplay`, `CompleteReplay`, and `FailReplay` in the right order, uses `problem.Code*` constants throughout (CodeIdempotencyKeyRequired:442, CodeValidationError:447, CodeIdempotencyKeyReused:476, CodeInternalError:481/463).
- The inline implementation diverges from `idempotency.Require` in two observable ways: (1) it interleaves idempotency setup with handler business logic rather than sitting outside it as a middleware; (2) it uses `log.Printf` on two error paths (lines 480, 495) rather than `slog`.

**Idempotency middleware raw string codes** (`internal/platform/idempotency/middleware.go:93,97,108,111,117,123`):
- Confirmed present. The six raw string literal codes are:
  - Line 93: `"IDEMPOTENCY_KEY_REQUIRED"` — this IS a catalog constant (`CodeIdempotencyKeyRequired = "IDEMPOTENCY_KEY_REQUIRED"` in `codes.go:30`), so it matches; but it is still a raw string cast, not the constant.
  - Line 97: `"IDEMPOTENCY_KEY_INVALID"` — NOT in `codes.go`. No catalog entry.
  - Line 108: `"REQUEST_BODY_TOO_LARGE"` — NOT in `codes.go`. No catalog entry.
  - Line 111: `"BAD_REQUEST"` — NOT in `codes.go`. No catalog entry.
  - Line 117: `"IDEMPOTENCY_KEY_CONFLICT"` — NOT in `codes.go`. No catalog entry (the catalog has `CodeIdempotencyKeyReused = "IDEMPOTENCY_KEY_REUSED"` at line 25 — a different name and value).
  - Line 123: `"INTERNAL"` — NOT in `codes.go`; diverges from `CodeInternalError = "INTERNAL_ERROR"` (line 28).
- The Phase F hardening fix (`d85c465b5`) migrated `writeErrJSON` to route through `problem.Write` (confirmed: `middleware.go:216–219` now calls `problem.Write`). The codes are still raw strings cast to `problem.Code(code)` — the fix addressed the HTTP body format (now `application/problem+json`) but not the code vocabulary.
- The catalog guard test (`codes_catalog_guard_test.go:31–35`) explicitly lists `guardedPackages` as only `documents/delivery/http` and `templates/delivery/http` — `internal/platform/idempotency/` is not guarded. So these out-of-catalog strings pass `go test`.

**Additional idempotency defects** (from register, confirmed by code read):
- TTL string `'24 hours'` duplicated at `postgres_store.go:91,229` — a named Go constant would prevent the two-edit requirement.
- `idempotency_keys.tenant_id` has no FK constraint (inline TODO at `postgres_store.go:54,87`).

### Applicable standard

**REQ-API-5** (backend-target-architecture.md §5): "Mutating endpoints that clients may retry accept an idempotency key and replay the original result. Approval/freeze-class operations are in scope by default. (MUST)"

**RF-10** (backend-target-architecture.md §10): "Idempotency coverage audit: which mutating routes accept keys vs should."

**REQ-H-2** (backend-target-architecture.md §2.2): "Every error path returns RFC 9457 problem+json with a code from the closed vocabulary. No bare http.Error, no ad hoc JSON. (MUST)"

**Stripe idempotency design guide** (https://stripe.com/docs/api/idempotent_requests, widely referenced in the industry): idempotency key validation should return a documented, client-actionable error code. Using `"BAD_REQUEST"` or `"INTERNAL"` (codes not in the public catalog) prevents clients from differentiating idempotency errors from general request errors. The Stripe guide specifically recommends a typed, stable error code for each idempotency failure mode (missing key, duplicate key with different payload, etc.).

**RFC 9457 §5 — Extension Members** (https://www.rfc-editor.org/rfc/rfc9457#section-5): the `code` extension member in MetalDocs's `Problem` type is a custom extension. Its purpose is to give clients a stable, machine-readable code. Emitting `"INTERNAL"` instead of `"INTERNAL_ERROR"` for the same class of error from two different code sites creates an inconsistent vocabulary that violates the spirit of a "closed vocabulary" (REQ-H-2).

### Verdict (per sub-finding)

**F-09a — Finalize handler inline idempotency:** KEEP with bounded annotation

The inline implementation is functionally correct. It calls all five protocol methods (`IsValidKey`, `RequestHash`, `BeginReplay`, `CompleteReplay`, `FailReplay`) in the right order and defers `FailReplay`. The handler's business logic is structurally incompatible with the `Require` middleware pattern: the handler needs to (a) hash the body before decoding it (done correctly), then (b) authorize based on the decoded document, then (c) complete/fail the idempotency slot. The `Require` middleware cannot interleave with that sequence — it wraps the entire handler, which means the slot is completed or failed after the full handler returns. Migrating to `Require` would require restructuring the authorization and business logic to separate the idempotency lifecycle from handler logic, which is a medium-effort refactor with no observable change to callers.

The register's concern was about the risk of missing `FailReplay` on new code paths. That concern is partially addressed: the deferred `FailReplay` at line 492–498 correctly releases the slot on any return path after `idempHandle` is set. The `log.Printf` on error paths (lines 480, 495) should migrate to `slog` (F-02 family), but that is not an idempotency correctness issue.

**Over-engineering check:** Migrating to `Require` middleware for this handler would require significant restructuring to split the handler into a pre-auth idempotency check and a post-auth execution, or passing idempotency state through context — both are more complex than the current inline approach. The inline implementation IS the right pattern when business logic must interleave with the idempotency lifecycle. KEEP it; add a code comment referencing this analysis so the decision is not re-litigated.

**F-09b — Raw string codes in the idempotency middleware:** REFACTOR

This is a small, clear REQ-H-2 violation. Five of the six codes (`"IDEMPOTENCY_KEY_INVALID"`, `"REQUEST_BODY_TOO_LARGE"`, `"IDEMPOTENCY_KEY_CONFLICT"`, `"BAD_REQUEST"`, `"INTERNAL"`) are not in `codes.go`. The correct fix is:

1. Add four new catalog constants to `codes.go`: `CodeIdempotencyKeyInvalid`, `CodeRequestBodyTooLarge`, `CodeIdempotencyKeyConflict`, and optionally a clarifying note that `CodeValidationError` covers `"BAD_REQUEST"`-class errors.
2. Replace the raw strings at lines 93, 97, 108, 111, 117, 123 with the catalog constants.
3. Specifically: `"INTERNAL"` at line 123 should become `problem.CodeInternalError` (existing constant, value `"INTERNAL_ERROR"`) — this is a bug: the middleware currently emits a different error code (`"INTERNAL"`) than every other internal error path in the codebase (`"INTERNAL_ERROR"`).
4. Expand `guardedPackages` in `codes_catalog_guard_test.go` to include `internal/platform/idempotency/` so these call sites are covered by the AST guard going forward.

**Note on `"IDEMPOTENCY_KEY_CONFLICT"` vs `CodeIdempotencyKeyReused`:** The register notes divergence but the difference matters semantically. The middleware's `"IDEMPOTENCY_KEY_CONFLICT"` fires on same-key/different-payload — which is exactly what `CodeIdempotencyKeyReused = "IDEMPOTENCY_KEY_REUSED"` is meant to represent. They should be the same code. The finalize handler already uses `CodeIdempotencyKeyReused` correctly at line 476. The fix is to use `CodeIdempotencyKeyReused` in the middleware too, not introduce a new constant.

**Over-engineering check:** Do NOT redesign the middleware's error-reporting surface. The fix is: add 3 constants to `codes.go`, replace 6 raw strings with constants, expand the guard's package list. Total change is approximately 15 lines.

**F-09c — Idempotency TTL duplicated SQL string:** SIMPLIFY

Define a single `const idempotencyTTL = "'24 hours'"` (or use a `time.Duration` parameter) in `postgres_store.go` and reference it at both lines 91 and 229. This is a routine DRY fix — no design change.

**F-09d — Missing FK on `idempotency_keys.tenant_id`:** DEFER

The inline TODO acknowledges this. Application-layer tenant isolation is in place (`tenantID` is always passed as a query parameter). Adding the FK requires a migration and testing that no orphaned rows exist. Trigger: when the idempotency table gains an operational review or when the broader tenant-isolation hardening program (F-12) runs its migration pass. This is not urgent and should not block the F-09b fix.

### Summary verdict table

| Sub-finding | Verdict | Effort | Blast-radius |
|---|---|---|---|
| F-09a: finalize handler inline idempotency | KEEP (add decision comment) | S | contained |
| F-09b: raw string codes in middleware | REFACTOR (add 3 constants, replace 6 strings, expand guard) | S | contained (platform/idempotency + codes.go + guard test) |
| F-09c: TTL string duplicated | SIMPLIFY (1 constant) | S | contained |
| F-09d: missing FK | DEFER (trigger: F-12 tenant-isolation program) | M | module (migration required) |

### ADR needed?

No. F-09a's KEEP decision is a rationale that belongs in a code comment, not an ADR — it is an implementation-detail judgment, not an architecture decision. F-09b/c are straightforward conformance fixes. F-09d's deferral trigger is explicit.

---

## Cross-finding observations

**D-03 (bare 405 responses in search and security modules)** is documented in the register as a cross-area finding. The search handler at `internal/modules/search/delivery/http/handler.go:54–56` returns a bare `w.WriteHeader(http.StatusMethodNotAllowed)` with no RFC 9457 body — confirmed present in the current code read. This violates REQ-H-2. It is not one of the assigned findings (F-03/F-13/F-09), but it is co-located with the F-13 search schema gaps and should be fixed in the same search-module touch (1-line fix: `httpresponse.WriteError(w, 405, problem.CodeValidationError, "method not allowed")`). Flagged here for the executing engineer; it does not change the verdict on F-13.

**Idempotency coverage scope (RF-10):** the register asks which mutating routes should accept idempotency keys. The finalize handler (F-09a) does so correctly. The approval signoff and route-admin handlers also have idempotency stores (`postgres_signoff_idemp_store.go`, `postgres_route_admin_idemp_store.go` — flagged under F-04 as structural duplicates). Whether additional mutating routes (document state transitions, controlled-document mutations) should accept idempotency keys is a REQ-API-5 compliance question for RF-10 and is outside the scope of this finding's evaluation.

---

## Priority and sequencing recommendation

| Finding | Verdict | Priority | Sequencing note |
|---|---|---|---|
| F-03: DELETE spec2.yaml + api/v2/ + fix CI gate | DELETE / REFACTOR | P1 | Independent; no sequencing constraint |
| F-09b: raw codes in idempotency middleware | REFACTOR | P1 | Independent; unblocks adding middleware to guard scope |
| F-13b: `businessUnit` param removal/rename | REFACTOR | P1 | Search module touch; group with F-13a and D-03 |
| F-13a: remove dead search struct fields | SIMPLIFY | P1 | Same touch as F-13b |
| F-13c: delete partial files | DELETE | P2 | Independent; zero-risk git rm |
| F-13d: `deprecated: true` on createManagedUser | SIMPLIFY | P2 | Trivial spec change; group with any spec-regen pass |
| F-09a: finalize handler | KEEP | — | Add decision comment only |
| F-09c: TTL constant | SIMPLIFY | P3 | Low urgency; group with any postgres_store.go touch |
| F-09d: idempotency_keys FK | DEFER | — | Trigger: F-12 tenant-isolation program |
| F-13e: iamapi three-domain bundle | DEFER | — | Trigger: merge-conflict bottleneck or module reorganization |

---

## Sources

- `wiki/backend/legacy-register.md` (F-03, F-13, F-09)
- `wiki/backend/contract-surface.md`
- `wiki/backend/platform/http-toolkit.md`
- `wiki/backlog/api-contract-hardening.md` (Phases A–F, closing re-audit)
- `wiki/architecture/backend-target-architecture.md` (REQ-API-2, REQ-API-5, REQ-H-2, RF-4, RF-5, RF-10)
- Code reads (2026-06-11): `api/openapi/spec2.yaml`, `internal/api/v2/types_gen.go`, `internal/platform/idempotency/middleware.go`, `internal/platform/problem/codes.go`, `internal/modules/documents/delivery/http/handler.go:435–540`, `internal/modules/search/delivery/http/handler.go`, `api/openapi/v1/openapi.yaml:4785–4820, 802–880`, `api/openapi/v1/partials/controlled-documents.yaml:200–229`, `ops/CAPABILITY_CATALOG.sha256`, `.github/workflows/invariants.yml:89–118`
- External standards: OpenAPI 3.0.3 specification (spec.openapis.org), RFC 9457 (rfc-editor.org), OWASP API Security Top 10 2023 (owasp.org), Stripe idempotency guide (stripe.com/docs), Google AIP-004 (google.aip.dev)
