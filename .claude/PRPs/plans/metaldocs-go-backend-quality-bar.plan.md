# Plan: MetalDocs Go Backend Production-Grade Quality Bar (v1)

## Summary
Publish a written, enforced Go quality bar for MetalDocs — 9 topic docs under `wiki/standards/golang/`, a curated `.golangci.yml`, and a CI lint gate wired into the existing `invariants.yml` workflow. Every rule is grounded in the 4+11 Critical+High findings already landed in module #1 + #2a.

## User Story
As Leandro (solo dev), I want a published quality bar + enforced lint config, so that the next 3 module reviews each cost <$15, produce zero repeat Critical findings, and every finding cites a bar section by anchor.

## Problem → Solution
No written standard → agent reviews re-invent rules each time, fixes regress, ISO audit has no evidence of code-quality controls → `wiki/standards/golang/` v1 + `.golangci.yml` + CI gate

## Metadata
- **Complexity**: Large
- **Source PRD**: `.claude/PRPs/prds/metaldocs-go-backend-quality-bar.prd.md`
- **PRD Phase**: single (no phase table in PRD)
- **Estimated Files**: 13 new + 1 updated (invariants.yml)

---

## UX Design

N/A — internal change. No user-facing UI surface.

---

## Mandatory Reading

Files that MUST be read before implementing:

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `wiki/reviews/2026-05-21-go-backend-review.md` | all | Evidence tracker — anchor citations required in bar docs |
| P0 | `wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md` | all | C1/C2/C3/C4/C5/H7/H10/H11 findings — bar content source |
| P0 | `wiki/reviews/2026-05-21-go-backend-review/cmd-metaldocs-api.md` | all | Module #1 findings — feeds package-layout + http-handlers |
| P0 | `.github/workflows/invariants.yml` | all | CI file to patch — add golangci-lint job |
| P1 | `internal/platform/idempotency/postgres_store.go` | all | BeginReplay/CompleteReplay/FailReplay canonical impl |
| P1 | `internal/platform/authn/context.go` | all | UserIDFromContext fail-closed pattern |
| P1 | `internal/platform/config/trusted_proxy.go` | all | CIDR trusted-proxy pattern |
| P1 | `internal/platform/problem/codes.go` | all | RFC 9457 typed error codes |
| P1 | `internal/modules/iam/domain/model.go` | all | iamdomain.Role typed enum |
| P1 | `internal/modules/audit/delivery/http/handler.go` | all | Canonical handler pattern |
| P2 | `~/.claude/rules/ecc/golang/coding-style.md` | all | What bar extends |
| P2 | `~/.claude/rules/ecc/golang/security.md` | all | What bar extends |
| P2 | `~/.claude/rules/ecc/golang/testing.md` | all | What bar extends |
| P2 | `~/.claude/rules/ecc/golang/patterns.md` | all | What bar extends |
| P2 | `go.mod` | all | Confirm Go version + dependencies |

---

## Patterns to Mirror

Code patterns discovered in the codebase. Follow these exactly.

### ERROR_WRAPPING
```go
// SOURCE: internal/platform/idempotency/postgres_store.go:71,92
return fmt.Errorf("idempotency: begin tx: %w", err)
return fmt.Errorf("idempotency: insert in_flight: %w", err)
// Pattern: "<subsystem>: <operation>: %w" — subsystem prefix mandatory
```

### TYPED_ENUM
```go
// SOURCE: internal/modules/iam/domain/model.go:3-10
type Role string
const (
    RoleApprover   Role = "approver"
    RoleAuthor     Role = "author"
    RoleEditor     Role = "editor"
    RoleSystemAdmin Role = "system_admin"
    RoleViewer     Role = "viewer"
)
```

### TYPED_ERROR_CODE
```go
// SOURCE: internal/platform/problem/codes.go:7,11-30
type Code string
const (
    CodeValidationError  Code = "VALIDATION_ERROR"
    CodeUnauthenticated  Code = "UNAUTHENTICATED"
    // ...
)
```

### CONTEXT_EXTRACTOR_FAIL_CLOSED
```go
// SOURCE: internal/platform/authn/context.go:10-24
func UserIDFromContext(ctx context.Context) (string, bool) {
    // Returns ("", false) on any absence — never returns zero-value silently
}
// Pattern: always (value, bool) — caller MUST check bool before using value
```

### TWO_PHASE_IDEMPOTENCY
```go
// SOURCE: internal/platform/idempotency/postgres_store.go:52-88
// BeginReplay returns one of three outcomes:
//   (handle, nil, nil)     = this caller is the winner — must call CompleteReplay/FailReplay
//   (nil, replay, nil)     = cache hit — return cached response
//   (nil, nil, ErrConflict) = in-flight conflict — return 409
handle, replay, err := store.BeginReplay(ctx, tenantID, actorID, key, payloadHash)
```

### TRUSTED_PROXY_CIDR
```go
// SOURCE: internal/platform/config/trusted_proxy.go:10-37
func LoadTrustedProxyCIDRs() ([]netip.Prefix, error) // reads METALDOCS_TRUSTED_PROXY_CIDRS
func ParseTrustedProxyCIDRs(raw string) ([]netip.Prefix, error)
// Empty env → nil slice = no upstream trust (fail-closed)
```

### SLOG_STRUCTURED
```go
// SOURCE: internal/modules/jobs/audit_integrity_validator/job.go:21,37
slog.ErrorContext(ctx, "audit_integrity_validator: validation failed", "error", err)
slog.InfoContext(ctx, "audit_integrity_validator: tick complete", "validated", n)
// Pattern: subsystem-prefix message, structured KV pairs, ErrorContext/InfoContext when ctx available
```

### HTTP_HANDLER
```go
// SOURCE: internal/modules/audit/delivery/http/handler.go
// 1. Check method → 405
// 2. Parse + trim + validate query params → problem.Write 400
// 3. Call service → problem.Write 500 on error
// 4. Map domain to response struct → JSON encode
```

### PROBLEM_WRITE
```go
// SOURCE: internal/platform/problem/codes.go + handler pattern
_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit value"))
_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit events"))
// Return after every problem.Write — no fallthrough
```

### TEST_TABLE_DRIVEN
```go
// SOURCE: internal/modules/auth/delivery/http/middleware_test.go:19-41
cases := []struct {
    method string
    path   string
    public bool
}{
    {http.MethodGet, "/api/v1/health/live", true},
    // ...
}
for _, tc := range cases {
    got := fn(tc.method, tc.path)
    if got != tc.public {
        t.Errorf("fn(%q, %q) = %v, want %v", tc.method, tc.path, got, tc.public)
    }
}
```

### PACKAGE_LAYER_ORDER
```
// SOURCE: directory survey of internal/modules/*/
// Dependency direction (imports allowed →):
// delivery/http → application/service → store → domain
// platform packages may be imported by any layer above domain
// No reverse imports (domain must not import application/service)
```

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `wiki/standards/golang/README.md` | CREATE | Index + anchor table for all bar sections |
| `wiki/standards/golang/typed-boundaries.md` | CREATE | string→typed migration patterns |
| `wiki/standards/golang/errors-and-logging.md` | CREATE | Error wrapping + slog conventions |
| `wiki/standards/golang/security-boundaries.md` | CREATE | Fail-closed authn, trusted-proxy, CORS, RFC 9457 |
| `wiki/standards/golang/idempotency-and-concurrency.md` | CREATE | Two-phase write, retry-safe handlers |
| `wiki/standards/golang/persistence.md` | CREATE | pgx, parameterized queries, tx discipline |
| `wiki/standards/golang/http-handlers.md` | CREATE | Context, validation, middleware ordering |
| `wiki/standards/golang/testing.md` | CREATE | Table-driven, t.Helper, no mock DB |
| `wiki/standards/golang/package-layout.md` | CREATE | Module structure + import direction |
| `wiki/standards/golang/refactor-playbook.md` | CREATE | Step-by-step legacy → bar sequence |
| `.golangci.yml` | CREATE | Curated linter config for MetalDocs |
| `.github/workflows/golangci-lint.yml` | CREATE | New workflow — lint gate on PR + main |

## NOT Building

- TypeScript/React/frontend rules (out of scope)
- Database migration policy (lives in `wiki/database/`)
- OpenTelemetry/Prometheus instrumentation (not wired)
- OpenAPI contract refactors (separate concern)
- E-signature module section (not built yet)
- `gofumpt` adoption (Q1 from PRD — defer to v2)
- Generated code lint (`api.gen.go` excluded in `.golangci.yml`)
- Pre-commit hook (v2 stretch)
- `make audit` target (v2 stretch)
- Auto-applied `goimports` PostToolUse hook (v2 stretch)
- Continuing #2a remaining Highs or #2b/#2c reviews (cursor stays at #2b until bar lands)

---

## Step-by-Step Tasks

### Task 1: Read all evidence files
- **ACTION**: Read mandatory files listed in Mandatory Reading table
- **IMPLEMENT**: Read `wiki/reviews/2026-05-21-go-backend-review.md`, `platform-2a-security.md`, `cmd-metaldocs-api.md` in full. Extract finding IDs (C1-C5, H1-H11, M-series) and their commit SHAs.
- **MIRROR**: N/A — research only
- **IMPORTS**: N/A
- **GOTCHA**: The review files contain the exact finding IDs and commit SHAs needed for bar evidence citations. Do not invent IDs.
- **VALIDATE**: You have a mapping: finding ID → description → commit SHA → which bar doc it feeds

### Task 2: Create `wiki/standards/golang/` directory + README
- **ACTION**: Create `wiki/standards/golang/README.md`
- **IMPLEMENT**: Front-matter with `extends` + `evidence` + `enforced_by` per PRD §15. Body: one-paragraph purpose, then an anchor table with columns: Section | Bar Doc | Failure Mode Prevented | Finding ID | Commit SHA | Lint Rule | Extends Rule. Cover all 9 topic docs. Add a "How to use this bar" section for agents: "Cite `wiki/standards/golang/<doc>.md#<anchor>` in every Critical/High finding."
- **MIRROR**: Wiki README pattern from existing `wiki/README.md`
- **IMPORTS**: N/A
- **GOTCHA**: `Extends:` front-matter paths must be exactly `~/.claude/rules/ecc/golang/<file>.md` — these are the ECC rule files
- **VALIDATE**: Every row in anchor table has a non-empty Lint Rule entry or "manual-review" annotation

### Task 3: Create `typed-boundaries.md`
- **ACTION**: Write `wiki/standards/golang/typed-boundaries.md`
- **IMPLEMENT**:
  - Front-matter: extends `coding-style.md`, evidence M-series findings from #2a, enforced_by `exhaustive` + `revive`
  - Sections:
    1. **The rule**: `string` is banned for IDs, roles, error codes, idempotency keys at any package boundary — use distinct named types
    2. **Proven types**: `iamdomain.Role` (source: `internal/modules/iam/domain/model.go:3-10`), `problem.Code` (source: `internal/platform/problem/codes.go:7`). Show type definitions verbatim.
    3. **Migration pattern**: `type TenantID string` / `type UserID string` — where to define (in domain package), how to migrate (wrap, not replace all at once)
    4. **Anti-patterns**: `func DoSomething(role string)` → `func DoSomething(role iamdomain.Role)`. No bare `string` parameters named `role`, `tenantID`, `userID`, `errorCode`, `idempotencyKey`
    5. **When typed ID is not needed**: internal helper with no cross-boundary exposure
- **MIRROR**: TYPED_ENUM, TYPED_ERROR_CODE patterns
- **IMPORTS**: N/A
- **GOTCHA**: `TenantID` and `UserID` are currently `string` in OpenAPI-generated types — document that the generated types are exempt (excluded from lint), but the hand-written domain + service + store layer must use typed IDs
- **VALIDATE**: Every anti-pattern has a compliant alternative shown inline

### Task 4: Create `errors-and-logging.md`
- **ACTION**: Write `wiki/standards/golang/errors-and-logging.md`
- **IMPLEMENT**:
  - Front-matter: extends `coding-style.md`, evidence H-series silent-failure findings + C3 idempotency error-swallow, enforced_by `errcheck` + `errorlint` + `nilerr`
  - Sections:
    1. **Error wrapping rule**: always `fmt.Errorf("subsystem: operation: %w", err)`. Show ERROR_WRAPPING pattern. Forbidden: `fmt.Errorf("...", err)` without `%w` for wrapped errors.
    2. **Never swallow**: `_ = err` is banned except for: `_ = problem.Write(w, ...)` (write errors non-actionable in handler), `json.Unmarshal` in audit payload (explicitly documented exception). Every other `_` assignment that discards an error must have a comment justifying it.
    3. **errors.Is / errors.As discipline**: always prefer `errors.Is(err, ErrFoo)` over string matching. Sentinel errors defined at package level with `var Err... = errors.New(...)`.
    4. **slog conventions**: 
       - Use `slog.ErrorContext(ctx, ...)` / `slog.InfoContext(ctx, ...)` when ctx available
       - Message: `"<subsystem>: <what happened>"` — no interpolation in message string
       - Fields: structured KV pairs only — `"error", err`, `"tenant_id", tenantID`, `"resource_id", id`
       - Forbidden: `log.Printf`, `fmt.Println`, `log.Fatal` outside of `main`/`cmd`
    5. **Log-and-return anti-pattern**: if you log an error and return it, the caller logs it again → double-logging. Rule: log OR return, not both. Exception: package boundaries where the original error loses context.
- **MIRROR**: ERROR_WRAPPING, SLOG_STRUCTURED patterns
- **IMPORTS**: N/A
- **GOTCHA**: `_ = problem.Write(w, ...)` is the documented exception for discarded write errors — ensure this is called out explicitly so reviewers don't flag it as a violation
- **VALIDATE**: Each rule has a DO / DO NOT code block pair

### Task 5: Create `security-boundaries.md`
- **ACTION**: Write `wiki/standards/golang/security-boundaries.md`
- **IMPLEMENT**:
  - Front-matter: extends `security.md`, evidence C1/C2/C5/H7, commits `def24e4a`/`2f8f6dcc`/`d2242313`, enforced_by `gosec` + `contextcheck`
  - Sections:
    1. **Fail-closed authn — UserIDFromContext**: show CONTEXT_EXTRACTOR_FAIL_CLOSED pattern. Rule: any function that needs actor identity calls `authn.UserIDFromContext(ctx)` and checks the bool. If `!ok` → return 401 via `problem.Write`. Forbidden: ignoring the bool, using the zero-value string as a valid actor.
    2. **Trusted-proxy CIDR**: show TRUSTED_PROXY_CIDR pattern. Rule: `X-Forwarded-For` / `X-Real-IP` trusted only if request originates from a CIDR in `METALDOCS_TRUSTED_PROXY_CIDRS`. Empty env = no upstream trust. Never call `r.RemoteAddr` substitution without CIDR check.
    3. **RFC 9457 problem envelope**: show PROBLEM_WRITE pattern. Rule: all error responses use `problem.Write`. No raw `http.Error`, no naked JSON `{"error": "..."}`, no status-only responses for 4xx/5xx.
    4. **CORS reject-disallowed**: cite H7. Rule: CORS middleware rejects origins not in the allowlist — no wildcard `*` in production. Link to the landed fix commit.
    5. **Header trust rules**: document which headers are trusted when and under what conditions (`X-Forwarded-Proto`, `X-Forwarded-For`). Show the middleware order: trusted-proxy check must happen before any header-based routing/rate-limiting.
    6. **Authn callsite audit checklist**: before any new handler, verify: (a) auth middleware in the chain, (b) UserIDFromContext bool checked, (c) 401 returned not panicked.
- **MIRROR**: CONTEXT_EXTRACTOR_FAIL_CLOSED, TRUSTED_PROXY_CIDR, PROBLEM_WRITE patterns
- **IMPORTS**: N/A
- **GOTCHA**: The commit SHAs in the PRD (`def24e4a`, `2f8f6dcc`, `d2242313`) must appear verbatim as evidence citations — do not paraphrase
- **VALIDATE**: Each section has a "Finding: [ID] → Fixed: [commit SHA]" line

### Task 6: Create `idempotency-and-concurrency.md`
- **ACTION**: Write `wiki/standards/golang/idempotency-and-concurrency.md`
- **IMPLEMENT**:
  - Front-matter: extends `patterns.md`, evidence C3/C4/H11, commit `07312d58`, enforced_by manual-review (no lint rule catches this class)
  - Sections:
    1. **Two-phase write pattern**: show TWO_PHASE_IDEMPOTENCY pattern verbatim. Explain three outcomes. Rule: no single-phase replay storage — `BeginReplay` must be called before any state change; `CompleteReplay`/`FailReplay` must be called in all code paths (defer-FailReplay pattern).
    2. **ON CONFLICT DO NOTHING RETURNING**: explain why the atomic insert-or-return avoids the check-then-act race. Show the SQL pattern from `postgres_store.go`.
    3. **Retry-safe handler semantics**: a handler is retry-safe if calling it N times with the same idempotency key produces the same visible outcome. Rule: mutating handlers (POST/PUT/PATCH/DELETE) must be either (a) idempotency-key gated, or (b) naturally idempotent (e.g. DELETE 404 = success).
    4. **H11 Go-layer guards**: actor non-empty check before DB round-trip. Response body size cap (64 KiB). Show the guard code from `postgres_store.go:65-67`.
    5. **Replay-race fix**: explain the replay collision scenario, cite C4, show how two-phase resolves it.
    6. **Out-of-scope (v1)**: distributed lock patterns, saga orchestration — document as deferred.
- **MIRROR**: TWO_PHASE_IDEMPOTENCY pattern
- **IMPORTS**: N/A
- **GOTCHA**: The `defer-FailReplay` pattern is critical — if `CompleteReplay` is called but a panic occurs before it, the slot stays in-flight forever. Document the canonical `defer store.FailReplay(...)` with a sentinel variable pattern.
- **VALIDATE**: The three BeginReplay outcomes are listed with caller obligations for each

### Task 7: Create `persistence.md`
- **ACTION**: Write `wiki/standards/golang/persistence.md`
- **IMPLEMENT**:
  - Front-matter: extends `patterns.md`, evidence open H-series persistence findings from #2a, enforced_by `sqlclosecheck` + `rowserrcheck` + `bodyclose`
  - Sections:
    1. **pgx v5 as the driver**: `github.com/jackc/pgx/v5` is the canonical driver. `lib/pq` is legacy and being phased out — new code uses pgx only.
    2. **Parameterized queries only**: `db.QueryContext(ctx, "SELECT ... WHERE id = $1", id)` — no string concatenation or `fmt.Sprintf` in SQL. Show compliant + anti-pattern.
    3. **RowsAffected discipline**: for UPDATE/DELETE, check `RowsAffected() == 0` and return a domain error — don't silently succeed on no-op mutations.
    4. **Rows.Close discipline**: `defer rows.Close()` immediately after `QueryContext`. `rows.Err()` must be checked after loop.
    5. **Transaction boundaries**: transactions opened in store layer, not service layer. Service calls store methods that optionally accept a `pgx.Tx` or use the store's default pool. Commits and rollbacks in the same function that calls `Begin`.
    6. **Connection pool hygiene**: never hold a conn across HTTP round-trips. Context must flow through every DB call.
- **MIRROR**: ERROR_WRAPPING pattern (for wrapping DB errors)
- **IMPORTS**: N/A
- **GOTCHA**: `DATA-DOG/go-sqlmock` exists in `go.mod` — document that it's legacy from before the "no mock DB" rule. New tests must use a real Postgres container (see `testing.md`). Do NOT write new tests with sqlmock.
- **VALIDATE**: Each rule has a compliant code snippet

### Task 8: Create `http-handlers.md`
- **ACTION**: Write `wiki/standards/golang/http-handlers.md`
- **IMPLEMENT**:
  - Front-matter: extends `patterns.md` + `security.md`, evidence module #1 Critical findings + H1/H4/H7, enforced_by `contextcheck` + `gocyclo`
  - Sections:
    1. **Handler anatomy**: show HTTP_HANDLER pattern (method check → parse+trim+validate → service call → map → encode). Each step mandatory.
    2. **Context discipline**: `r.Context()` flows to every DB/service call. No `context.Background()` in handlers. Context carries deadline from HTTP server config.
    3. **Request validation at boundary**: trim whitespace (`strings.TrimSpace`) on all string query params and body fields before use. Validate ranges before passing to service layer.
    4. **No panic in handlers**: recovery middleware catches panics, but handlers must not rely on it. Explicit error returns only.
    5. **Middleware ordering** (canonical stack, top-to-bottom):
       ```
       Recovery
       Trusted-proxy CIDR extraction
       Rate limiting (uses trusted IP)
       Idempotency key extraction
       CORS
       Authn (sets UserID in context)
       Authz (checks role)
       Handler
       ```
    6. **problem.Write for all error responses**: see `security-boundaries.md` for RFC 9457 envelope rule.
    7. **HTTP method routing**: use `http.ServeMux` pattern with method check in handler body (not path-level routing per method) — consistent with existing codebase pattern.
- **MIRROR**: HTTP_HANDLER, PROBLEM_WRITE patterns
- **IMPORTS**: N/A
- **GOTCHA**: The middleware ordering is security-critical — rate limiting must use the trusted IP (after trusted-proxy extraction), not the raw RemoteAddr. Authn must come before authz. Document this explicitly.
- **VALIDATE**: Middleware stack is documented as an ordered numbered list

### Task 9: Create `testing.md`
- **ACTION**: Write `wiki/standards/golang/testing.md`
- **IMPLEMENT**:
  - Front-matter: extends `~/.claude/rules/ecc/golang/testing.md`, evidence no specific finding — this is the MetalDocs-specific extension of the ECC rule
  - Sections:
    1. **No mock DB rule**: `DATA-DOG/go-sqlmock` is BANNED for new tests. Integration tests hit real Postgres via testcontainers or the CI-provided postgres service. Reason: prior incident class where mock/real divergence masked broken behavior (cite the "no mock DB" memory: `feedback_audit_ai_designs.md` context).
    2. **Table-driven tests**: show TEST_TABLE_DRIVEN pattern. Use anonymous struct slice with named fields. `t.Errorf` includes all inputs. `t.Helper()` in shared assertion helpers.
    3. **Test naming**: `Test<FunctionName>_<Scenario>` e.g. `TestBeginReplay_EmptyActorReturnsError`.
    4. **testdata/ fixtures**: SQL fixtures for integration tests live in `testdata/` inside the package. JSON/YAML test inputs go there too.
    5. **No shared mutable state**: test cases must not depend on execution order. No package-level vars mutated by tests.
    6. **Race detection**: always `go test -race ./...` in CI. No data races tolerated.
    7. **testsupport package**: shared test helpers live in `internal/testsupport/`. No test helpers in `internal/platform/` or module packages.
    8. **Context_test.go pattern**: for platform packages with context extractors, test: nil context, empty context, populated context, whitespace, trim behavior — as in `internal/platform/authn/context_test.go`.
- **MIRROR**: TEST_TABLE_DRIVEN pattern
- **IMPORTS**: N/A
- **GOTCHA**: `go-sqlmock` exists in go.mod but must not be used in new code. Explicitly call this out so agents don't reach for it.
- **VALIDATE**: "No mock DB" rule includes the enforcement mechanism (CI real Postgres service)

### Task 10: Create `package-layout.md`
- **ACTION**: Write `wiki/standards/golang/package-layout.md`
- **IMPLEMENT**:
  - Front-matter: extends `patterns.md`, evidence module #1 Criticals (wiring/lifecycle issues), enforced_by `depguard` or manual-review
  - Sections:
    1. **Module template**: show canonical directory structure:
       ```
       internal/modules/<name>/
         domain/         # types, interfaces, sentinel errors
         application/    # service, use-case orchestration
         store/          # DB access (pgx)
         delivery/http/  # handler, routes, request/response types
       ```
    2. **Import direction (law)**: `delivery → application → store → domain`. No reverse imports. Domain imports nothing from the module. Show what each layer may import from `internal/platform/`.
    3. **Platform packages**: `internal/platform/` packages are shared infrastructure — `authn`, `problem`, `idempotency`, `ratelimit`, `security`, `tenant`, `httpresponse`, `config`, `db`, `observability`. Do not put module-specific logic in platform.
    4. **Constructor invariant pattern** (from H10): NewXxx constructors validate all mandatory fields and return `(T, error)`. No valid zero-value structs for types with required fields. Show the `NewConfig` pattern from `internal/platform/ratelimit/`.
    5. **cmd entrypoints**: `apps/api/cmd/metaldocs-api` is the only HTTP server entrypoint. Wiring lives in `apps/api/internal/wiring/`. No business logic in `cmd`.
    6. **No import cycles**: if two packages need each other, extract shared types to a third package (typically the domain package).
    7. **Generated code**: `internal/api/v2/types_gen.go` and `api.gen.go` are excluded from lint via `.golangci.yml` `exclude-files`. Never edit generated files by hand.
- **MIRROR**: PACKAGE_LAYER_ORDER pattern
- **IMPORTS**: N/A
- **GOTCHA**: `apps/api/internal/wiring/` imports across module boundaries intentionally (that's its job). Dependency-direction rule applies to module-internal layers only, not to the wiring package.
- **VALIDATE**: Import direction rule is stated as a one-sentence law with direction arrows

### Task 11: Create `refactor-playbook.md`
- **ACTION**: Write `wiki/standards/golang/refactor-playbook.md`
- **IMPLEMENT**:
  - Front-matter: no lint enforced_by (this is process doc)
  - Sections:
    1. **Overview**: the playbook describes how to bring a legacy module from zero to bar-compliant in phases. Works for any `internal/modules/<name>` or `internal/platform/<name>` package.
    2. **Phase 0 — Scope**: identify the module (package path), run ECC `ecc:go-review` agent with `wiki/standards/golang/README.md` as the rubric reference. Capture findings in `wiki/reviews/<date>-go-backend-review/<module-slug>.md`.
    3. **Phase 1 — Critical fixes**: land all Critical findings first. Each fix gets its own commit. Commit message: `fix(<module>): <finding-ID> <one-line description>`. Update findings doc with commit SHA.
    4. **Phase 2 — High fixes**: land High findings. Prioritize: security-boundary Highs before API-shape Highs. Same commit discipline.
    5. **Phase 3 — Lint baseline**: run `golangci-lint run ./path/to/module/...`. Document surviving issues in `wiki/reviews/<date>-go-backend-review/<module-slug>-lint-baseline.md`. Gate future PRs with `--new-from-rev`.
    6. **Phase 4 — Medium/Low (incremental)**: Medium findings addressed in subsequent PRs tagged with the module. Low findings addressed opportunistically when touching the file.
    7. **Phase 5 — Bar update**: if the fix introduces a new pattern not yet in `wiki/standards/golang/`, update the relevant bar doc same PR.
    8. **Split-and-conquer rule**: if a module exceeds 500 LoC of changes, split into sub-modules (e.g., #2a → security platform packages split from full platform review). Reference the #2a precedent.
    9. **LoC bucket guidance**: <200 LoC = single PR; 200-500 LoC = 2-3 PRs; >500 LoC = split-and-conquer
    10. **Tracker update**: after each phase, update `wiki/reviews/2026-05-21-go-backend-review.md` cursor row.
- **MIRROR**: N/A — process doc
- **IMPORTS**: N/A
- **GOTCHA**: Playbook must be prescriptive enough that an ECC agent can follow it cold. Include exact commit message format, exact file paths, exact agent invocation (`ecc:go-review`).
- **VALIDATE**: A hypothetical agent reading only this doc knows exactly what to do in each phase

### Task 12: Create `.golangci.yml`
- **ACTION**: Create `.golangci.yml` in repo root
- **IMPLEMENT**: Use golangci-lint v2 config format (project on Go 1.25 — v2 is appropriate per PRD Q2). Include:

```yaml
# golangci-lint v2 config — MetalDocs Go Backend Quality Bar v1
# See wiki/standards/golang/README.md for rationale

version: "2"

linters:
  enable:
    # Core correctness
    - errcheck          # unchecked errors
    - govet             # go vet checks
    - staticcheck       # SA + ST checks (already in CI, keep aligned)
    - nilerr            # return nil where err != nil
    - errorlint         # misuse of errors.Is/As/New
    # Security
    - gosec             # OWASP-class security issues
    # Code quality
    - gocritic          # opinionated style/correctness
    - revive            # fast, configurable linter
    - gocyclo           # cyclomatic complexity
    - gocognit          # cognitive complexity
    # DB / IO discipline
    - sqlclosecheck     # rows.Close not called
    - rowserrcheck      # rows.Err not checked
    - bodyclose         # http response body not closed
    # Exhaustiveness
    - exhaustive        # enum switch exhaustiveness

linters-settings:
  gocyclo:
    min-complexity: 15   # warning v1, gate v2
  gocognit:
    min-complexity: 20   # warning v1, gate v2
  errcheck:
    # Allow _ = problem.Write(...) — write errors are non-actionable in HTTP handlers
    exclude-functions:
      - (github.com/your-org/metaldocs/internal/platform/problem).Write
  gosec:
    excludes:
      - G104  # errors unhandled — covered by errcheck with our exception list
  revive:
    rules:
      - name: exported
      - name: var-naming
      - name: error-return
      - name: error-strings   # error strings must not be capitalized or end with punctuation
  exhaustive:
    default-signifies-exhaustive: true

issues:
  exclude-files:
    # Generated code — never edit by hand
    - ".*\\.gen\\.go$"
    - "internal/api/v2/types_gen\\.go"
  exclude-rules:
    # Test files: relax some rules
    - path: "_test\\.go$"
      linters:
        - gosec
        - gocyclo
        - gocognit

run:
  timeout: 5m
  # Exclude vendor if present
  skip-dirs:
    - vendor
```

- **MIRROR**: N/A — config file
- **IMPORTS**: N/A
- **GOTCHA 1**: The `errcheck.exclude-functions` path must match the actual module path in `go.mod`. Read `go.mod` to get the module name before writing this field.
- **GOTCHA 2**: `staticcheck` is already run via the `dominikh/staticcheck-action` in `invariants.yml`. Including it in golangci-lint too is fine (golangci-lint wraps it), but ensure the versions don't conflict. Keep `staticcheck` in golangci-lint for local `golangci-lint run` convenience.
- **GOTCHA 3**: `gocyclo` and `gocognit` are warnings in v1 — do NOT set them as hard failures. Gate on v2.
- **VALIDATE**: Run `golangci-lint run --help` mentally — `version: "2"` is the v2 config key. Linter names are lowercase, matching golangci-lint v2 canonical names.

### Task 13: Create `.github/workflows/golangci-lint.yml`
- **ACTION**: Create new workflow file (do NOT modify invariants.yml — keep it clean)
- **IMPLEMENT**:

```yaml
name: Go Lint

on:
  pull_request:
  push:
    branches: [main]

jobs:
  golangci-lint:
    name: golangci-lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.64.8   # last stable before v2 action; update with project cadence
          args: --timeout=5m
```

- **MIRROR**: `invariants.yml` job structure (checkout → setup-go → lint)
- **IMPORTS**: N/A
- **GOTCHA**: Use `golangci/golangci-lint-action@v6` — it supports both v1 and v2 config formats. Pin the version. Do NOT use `latest` tag.
- **VALIDATE**: Workflow triggers match `invariants.yml` (pull_request + push main)

---

## Testing Strategy

### Unit Tests (for the bar itself)
No runtime code is produced — the deliverables are documentation files + config. Testing is validation-based.

### Validation Checklist for Bar Docs

| Check | How | Expected |
|---|---|---|
| All finding IDs cited | grep `C[1-5]\|H[1-9]\|H10\|H11` across all bar docs | ≥ 15 citations across 9 docs |
| Commit SHAs cited | grep `def24e4a\|2f8f6dcc\|d2242313\|07312d58\|4a6a9e8b` | ≥ 1 per security/idempotency doc |
| Extends front-matter present | head every bar doc | All 9 docs have `extends:` field |
| enforced_by present | head every bar doc | All 9 docs have `enforced_by:` or `manual-review` |
| No invented patterns | Every code snippet traceable to Patterns to Mirror section or read source files | — |

### Edge Cases Checklist
- [ ] Generated code exclusion in `.golangci.yml` correctly escapes regex
- [ ] `errcheck` exclude-functions uses correct module path from `go.mod`
- [ ] Workflow file uses correct `golangci-lint-action` version tag
- [ ] `gocyclo`/`gocognit` NOT set as hard failures in v1

---

## Validation Commands

### Static Analysis
```bash
# Verify golangci-lint config is valid
golangci-lint config verify

# Run lint on the Go codebase (expect baseline issues, not zero)
golangci-lint run ./...
```
EXPECT: No "invalid config" error. Some existing issues reported (that's the baseline).

### Lint Smoke Test
```bash
# Verify the workflow file is valid YAML
python -c "import yaml; yaml.safe_load(open('.github/workflows/golangci-lint.yml'))"
```
EXPECT: No error

### Build Check
```bash
go build ./...
```
EXPECT: Zero errors (lint changes don't touch Go code)

### Documentation Completeness
```bash
# Verify all 9 topic docs + README + refactor-playbook exist
ls wiki/standards/golang/
```
EXPECT: `README.md typed-boundaries.md errors-and-logging.md security-boundaries.md idempotency-and-concurrency.md persistence.md http-handlers.md testing.md package-layout.md refactor-playbook.md`

### Manual Validation
- [ ] Open `wiki/standards/golang/README.md` — anchor table has all 9 docs with non-empty Lint Rule column
- [ ] Open `security-boundaries.md` — C1/C2/C5 finding IDs visible with commit SHAs
- [ ] Open `idempotency-and-concurrency.md` — three BeginReplay outcomes listed with caller obligations
- [ ] Open `.golangci.yml` — `errcheck.exclude-functions` has correct module path
- [ ] Open `.github/workflows/golangci-lint.yml` — triggers on pull_request + push main

---

## Acceptance Criteria
- [ ] All 10 files under `wiki/standards/golang/` exist
- [ ] `.golangci.yml` exists in repo root and passes `golangci-lint config verify`
- [ ] `.github/workflows/golangci-lint.yml` exists and is valid YAML
- [ ] Every bar doc has `extends:`, `evidence:`, `enforced_by:` front-matter
- [ ] Every Critical finding (C1-C5) from #2a is cited in at least one bar doc with commit SHA
- [ ] H10 + H11 are cited in bar docs
- [ ] `go build ./...` still passes (no Go code changed)
- [ ] README anchor table covers all 9 docs with lint rule mapping

## Completion Checklist
- [ ] Code follows discovered patterns (bar docs use real code snippets from codebase)
- [ ] Error handling matches codebase style (error wrapping examples use subsystem prefix)
- [ ] Logging follows codebase conventions (slog examples use ErrorContext/InfoContext)
- [ ] No hardcoded values (golangci-lint version pinned but documented)
- [ ] Documentation updated — `wiki/README.md` should reference the new `wiki/standards/` directory
- [ ] No unnecessary scope additions (no OTel, no e-signature section, no pre-commit hook)
- [ ] Self-contained — no questions needed during implementation

## Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `go.mod` module path not matching `errcheck.exclude-functions` | Medium | Lint false-positives on `problem.Write` | Read go.mod Task 12, confirm module path before writing |
| `golangci-lint-action@v6` + v2 config format mismatch | Low | CI fails on lint action | v6 action supports v2 config; test with `golangci-lint config verify` |
| Finding IDs from review files differ from PRD §14 | Low | Incorrect citations in bar docs | Read actual review files in Task 1 before writing any citations |
| `wiki/standards/golang/` directory drift from wiki/README.md index | Medium | wiki-curator gap | After all files created, update `wiki/README.md` index |

## Notes

- **Go module path**: read from `go.mod` first line — needed for `errcheck.exclude-functions`. Likely `github.com/something/metaldocs`.
- **PRD open questions resolved**:
  - Q1 (`gofumpt`): defer to v2 — use `gofmt` only in v1
  - Q2 (golangci-lint v2 config): use v2 format (project on Go 1.25)
  - Q3 (`gocyclo`/`gocognit`): warnings v1, hard gate v2 — reflected in config
  - Q4 (CI location): new workflow file, not modifying invariants.yml
  - Q5 (generated code): excluded via `exclude-files` regex in `.golangci.yml`
- **wiki/README.md update**: after creating `wiki/standards/golang/`, add an entry to `wiki/README.md` under a new "Standards" section pointing to `wiki/standards/golang/README.md`.
- **Branch**: implement on a fresh branch off `main` — not `codex/backend-invariant-slice`. The PRD explicitly states "fresh branch (not extending `codex/backend-invariant-slice`)".
