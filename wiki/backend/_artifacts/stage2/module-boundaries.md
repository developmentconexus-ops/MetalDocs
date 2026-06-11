# Stage 2 Evaluation — Module Boundaries & Platform Layering

> **Theme:** Module boundaries & platform layering
> **Findings:** F-06, F-05, D-04, F-15
> **Author:** Stage-2 evaluation agent (2026-06-11)
> **Standards judged against:** Cockburn Hexagonal Architecture (alistair.cockburn.us/hexagonal-architecture), DDD Bounded Context (Evans 2003; Fowler bliki/BoundedContext), Go `internal/` visibility spec (go.dev/doc/modules/layout), Go proverb "a little copying is better than a little dependency" (go-proverbs.github.io), REQ-TOP-1, REQ-TOP-2

---

## How to read this document

Each section: confirms the finding is current (file:line re-verification), states the standard, delivers a verdict with rationale, names the smallest correct fix, assigns effort + blast-radius, states whether an ADR is needed, and includes an explicit over-engineering check.

Verdict vocabulary: **KEEP** (already meets bar), **SIMPLIFY** (reduce/collapse), **REFACTOR** (restructure to meet standard), **DELETE** (remove), **DEFER** (real but trigger-dependent).

---

## F-06 — Cross-Module SQL and Infrastructure Boundary Violations

### Current state (re-confirmed)

Nine distinct boundary violations are live in the codebase. Re-verified anchors:

| Violation | File:line | Cross-boundary direction |
|---|---|---|
| `auth/infrastructure/postgres` writes to `metaldocs.iam_users` (last_login_ip, user_agent) | `internal/modules/auth/infrastructure/postgres/repository.go:211-236` (RecordLastLoginContext const UPDATE iam_users) | auth infra → IAM table |
| `controlleddocuments/delivery/http` issues 2+ raw SQL queries directly against `h.db` inside the handler | `internal/modules/controlleddocuments/delivery/http/routes.go:281-317` (inline SELECT EXISTS on controlled_documents) | delivery layer → DB |
| `taxonomy/infrastructure.TemplateVersionChecker` queries `templates_template_version` and `templates_template` directly | `internal/modules/taxonomy/infrastructure/template_version_checker.go:30-47` (templateVersionQuery) | taxonomy infra → templates tables |
| `iam/delivery/http/sessions_handler.go` imports `auth/infrastructure/postgres` types directly | `internal/modules/iam/delivery/http/sessions_handler.go:19` (`authpg "metaldocs/internal/modules/auth/infrastructure/postgres"`) | IAM delivery → auth infra package |
| `documents/delivery/http/handler.go` issues 4 raw db.QueryRowContext inside the finalizeDocument handler | `internal/modules/documents/delivery/http/handler.go:512-536` (inline SELECT on documents and controlled_documents) | delivery layer → DB |
| `platform/observability` imports `modules/auth/domain` | `internal/platform/observability/http.go:15` (authdomain import) | platform → module domain (REQ-TOP-2) |

The remaining three from the register (security module JOIN, standalone `PostgresControlledDocumentRepository` construction in `main.go:357`, CanRead SQL duplication) follow the same pattern and are confirmed by grep. None have been resolved since the Stage-1 audit.

### Standard

**Cockburn Hexagonal Architecture (2005):** the application has a primary (driving) side and a secondary (driven) side. A module's internal state (its SQL tables, its repository) is behind a port; other modules must adapt through that port — they do not reach through it into the hexagon's interior. Direct table references from a foreign module are an adapter-bypass: the boundary protection is void.

**DDD Bounded Context (Evans 2003; Fowler bliki/BoundedContext):** model elements inside one bounded context are not referenced directly from another context. Cross-context collaboration uses the published interface of the upstream context; the downstream context is isolated from upstream internals. Any schema change inside the upstream context that touches the raw table must then propagate by scanning all downstream raw-SQL callers — the signature of a broken boundary.

**REQ-TOP-1:** "Cross-module access goes through a module's application service or published Go interface — never another module's repository, SQL, or domain internals." (MUST).

**REQ-TOP-2:** "A platform package importing a module is a layering violation." (MUST). The `platform/observability` → `auth/domain` import violates this independently of REQ-TOP-1.

### Verdict: REFACTOR (P1)

The violations are real, confirmed, and non-trivial. However severity is **medium**, not high or critical, because:

1. The codebase is a modular monolith — no network boundary exists, so a cross-module SQL read does not introduce latency or distributed-consistency risk today.
2. All affected tables are in the same Postgres instance. The pragmatic risk is **schema-change coupling** and **authz bypass** (delivery-layer SQL skips the module's service-layer authz checks), not data corruption.
3. Some violations are genuinely more severe than others and must be prioritized differently.

**Prioritised sub-findings:**

**F-06a — `platform/observability` imports `auth/domain` (REQ-TOP-2 MUST, P0).**
This is a hard layering inversion: a platform package depends on a module domain package. The fix is small: replace `authdomain.CurrentUserFromContext` with a callback/interface injected at construction time — identical to the pattern already used correctly in `platform/ratelimit.Middleware` (which accepts `userExtractor func(*http.Request) string`). Effort: S. No ADR needed. This is the smallest and most important fix in the family.

**F-06b — Delivery-layer raw SQL in `controlleddocuments` and `documents` handlers (REQ-H-1 MUST, P1).**
`REQ-H-1` states: "Handlers do: decode → validate at boundary → call one application service → map result/error." Inline `h.db.QueryRowContext` in a handler is a structural violation. The fix is extracting the queries into the module's own repository/application service. This is a medium-effort refactor but entirely self-contained — blast radius is the affected module only.

**F-06c — `auth` repo writes to `iam_users` (P1).**
`RecordLastLoginContext` is genuinely cross-module (auth infra mutating IAM table). The correct fix is to extract a narrow `LoginContextRecorder` port interface in IAM, implement it in IAM's infra, and call it from auth's service via the port. Alternatively — and simpler — the write can be moved to IAM's application layer, invoked via a published service call from auth. The latter is the minimum-change path. Effort: M.

**F-06d — `iam/delivery` imports `auth/infrastructure/postgres` (V-01, P1).**
The sessions handler uses `authpg.SessionAdminQuery` and `authpg.SessionListItem` types defined in auth's infrastructure package. The fix is to promote these types to `auth/domain` or define a matching interface/type in IAM's domain, breaking the direct infra import. Effort: S-M (type migration, not logic change).

**F-06e — `taxonomy.TemplateVersionChecker` queries templates tables directly (P2).**
This is a pure boundary violation but the query is read-only and the coupling is narrow (one method, one query). The correct fix is a `TemplateVersionPort` interface on the templates module. Because this is a read-only cross-module lookup with no authz consequence, the risk is maintenance-only. Effort: S. Can be sequenced after F-06a/b/c/d.

### Smallest correct fix per sub-finding

| Sub | Fix | Effort |
|---|---|---|
| F-06a | Extract user-ID resolution in `httpObs` to a `func(*http.Request) string` callback injected at construction; remove `authdomain` import | S |
| F-06b | Move inline SELECT calls from delivery handlers into repository methods; handlers call service | M per module |
| F-06c | Introduce `iam.LoginContextPort` interface; move write logic to IAM infra; auth calls port | M |
| F-06d | Promote `SessionAdminQuery`/`SessionListItem` to `auth/domain` or define mirror in `iam/domain` | S |
| F-06e | Introduce `templates.VersionCheckerPort` in templates/domain; taxonomy calls port | S |

### Effort / blast radius

- F-06a: S effort, contained (one platform package)
- F-06b: M effort, module-level (controlleddocuments, documents delivery layers)
- F-06c: M effort, cross-module (auth + iam — but bounded to one call path)
- F-06d: S effort, module-level (iam delivery layer)
- F-06e: S effort, cross-module (taxonomy + templates, one query)

### ADR needed?

No new ADR required. REQ-TOP-1 and REQ-TOP-2 are already normative. The `auth`↔`iam` circular coupling is tracked in ADR 0007 (auth/IAM separation) — the fixes here advance that ADR's intent without requiring a new decision.

### Over-engineering check

Do NOT introduce message-passing, event buses, or gRPC-style inter-module stubs for these boundaries. This is a modular monolith; the correct abstraction is a Go interface defined in the upstream module's domain or application package, implemented in that module's infrastructure, injected at composition root. No extra indirection layers beyond that. F-06e in particular could be over-engineered — the query is 2 lines; a thin port interface is sufficient and the correct answer.

---

## F-05 + D-04 — Duplicate Rate-Limiters; platform/security vs platform/ratelimit Same Concern

### Current state (re-confirmed)

Two independent rate-limiter implementations coexist:

| Package | File | Algorithm | Domain imports | Production status |
|---|---|---|---|---|
| `platform/security` | `internal/platform/security/ratelimit.go:1-204` | Fixed-window, global, per-identity | `auth/domain` (line 12), `iam/domain` (line 13) | Active — wired at `main.go:598` via `rateLimiter.Wrap(mux)` |
| `platform/ratelimit` | `internal/platform/ratelimit/middleware.go:1-~270` | Token-bucket, per-route, callback-based | None (userExtractor callback) | Inactive — `main.go:501` calls `RegisterRoutes(mux)` not `RegisterRoutesWithRateLimit`; `documents/module.go:118-119` passes `nil` |

Secondary confirmed defect: `security.RateLimiter.requestIdentity` (lines 181-192) checks `authdomain.CurrentUserFromContext` then `iamdomain.UserIDFromContext` as a fallback; both are written simultaneously by auth middleware, so the fallback is dead code (confirmed: `platform/authn.UserIDFromContext` already provides the canonical single accessor).

D-04 (from cross-area duplication register) describes the same root cause: two packages implement the same concern at different maturity levels with different dependency models and no documented relationship between them.

### Standard

**Go `internal/` visibility and package cohesion (go.dev/doc/modules/layout):** each package should have a single, well-defined responsibility. Two packages implementing the same concern at different maturity levels without a declared supersession relationship violates the single-responsibility principle at the package level and is a maintenance hazard.

**REQ-TOP-2 (MUST):** `platform/security/ratelimit.go` imports `modules/auth/domain` and `modules/iam/domain` — a hard layering violation. The fix to REQ-TOP-2 and the consolidation of the two limiters are the same operation.

**Go proverb "a little copying is better than a little dependency" (Rob Pike, Gopherfest SV 2015; go-proverbs.github.io):** this proverb argues that a small amount of copied code is acceptable to avoid a dependency. It does NOT argue for maintaining two independently-growing implementations of the same concern indefinitely. The proverb is satisfied when the duplication is truly incidental (a few trivial helper lines). It is not satisfied here: the two limiters implement the same HTTP middleware concern, share the same `ClientIP` call, and will diverge in behavior if either is modified independently.

### Verdict: REFACTOR (P1) — delete `platform/security.RateLimiter`; activate `platform/ratelimit.Middleware`

The correct decision is:

1. **Delete** `security.RateLimiter` (the older, domain-importing, fixed-window implementation). It is the inferior implementation on every axis: it imports domain packages (REQ-TOP-2 violation), it uses a fixed window (weaker against burst than token-bucket), its identity resolution has a dead fallback, and it sits at the wrong position in the middleware chain for login protection (it wraps the entire mux after auth, so it cannot implement the pre-auth IP-keyed login rate limit required by REQ-MW-5).

2. **Activate** `platform/ratelimit.Middleware` for per-route limiting, and add a pre-auth IP-keyed limiter for the login path as part of the F-01 middleware chain fix. `platform/ratelimit.Middleware` already uses the correct inversion-of-control pattern (`userExtractor` callback) and is fully tested.

3. The identity extraction fix needed for `platform/observability` (F-06a) is the same pattern: inject a callback rather than importing domain packages directly.

**Rate-limit activation detail:** activating `platform/ratelimit` requires two changes: (a) call `RegisterRoutesWithRateLimit` instead of `RegisterRoutes` in `main.go:501` and (b) construct a `ratelimit.Middleware` with a configured quota and pass it. The `DefaultConfig` already ships sensible defaults (60 presign, 30 commit, 20 export). This is a two-file change in main.go and documents/module.go.

### Smallest correct fix

1. Remove `platform/security/ratelimit.go` (the `RateLimiter` type and its `NewRateLimiter`, `Wrap`, `allow`, `requestIdentity`, `sweepExpiredLocked`, `shouldSkipRateLimit`, `strconvSecondsCeil` functions and `windowCounter` struct). Remove the `authdomain` and `iamdomain` imports from `platform/security`.
2. In `main.go`: replace `security.NewRateLimiter(...)` and `rateLimiter.Wrap(mux)` with `ratelimit.New(ctx, ratelimit.DefaultConfig())` wired at the correct chain position (see F-01 for chain rewiring scope).
3. Activate per-route limiting: change `main.go:501` to call `docMod.RegisterRoutesWithRateLimit(mux, rlMiddleware)`.
4. Remove the `RateLimitConfig` fields from `internal/platform/config/ratelimit.go` that were specific to the old fixed-window limiter (`WindowSeconds`, `MaxRequests`) once no caller references them.

### Effort / blast radius

Effort: **M** (touches main.go wiring, deletes one 204-line file, activates 2-line change in documents/module.go, removes config fields). Blast radius: **module** (changes the security platform package, main.go composition root, and documents module wiring — all well-understood surfaces with no hidden consumers beyond those named).

### ADR needed?

No. REQ-TOP-2 is already normative. The per-route limiter was always intended to supersede the global one (per `wiki/architecture/rate-limiting.md §2.2`). The consolidation is execution of existing intent, not a new decision. Document the deletion in the PR message referencing RF-2 and RF-9.

### Over-engineering check

Do NOT introduce a unified "rate limit abstraction" interface over both implementations before consolidating. Do NOT redesign the quota model or add Redis backing before this baseline is activated. The right move is the smallest one: delete the inferior implementation, activate the existing good one. Redis-backed distributed rate limiting (needed for multi-replica correctness) is a separate, later concern — defer it until replica count > 1 is confirmed in production.

---

## F-15 — Duplicate Private Helpers Across Platform Packages

### Current state (re-confirmed)

Two pairs of duplicated private helper functions exist in the platform layer:

| Helper | Location A | Location B |
|---|---|---|
| `parseBoolEnv` | `internal/platform/config/attachments.go:108-114` | `internal/platform/authn/config.go:213-220` |
| `splitCSV` | `internal/platform/config/cors.go:63-76` | `internal/platform/authn/config.go:222-231` |

Both are re-confirmed by direct file read. The implementations are **not identical** in a strict sense:

- `parseBoolEnv` in `attachments.go`: `strings.EqualFold(raw, "true") || raw == "1"` (2 accepted truthy values)
- `parseBoolEnv` in `authn/config.go`: `raw == "1" || raw == "true" || raw == "yes" || raw == "on"` (4 accepted truthy values — more permissive)
- `splitCSV` in `cors.go`: trims spaces, filters empty strings — identical logic to `authn/config.go:222-231`

The `parseBoolEnv` divergence is the actual drift risk the register flagged: the two functions now accept different sets of truthy env-var values. A developer setting `METALDOCS_SOME_FLAG=yes` would see different results depending on which package reads the flag.

### Standard

**Go `internal/` visibility (go.dev/doc/modules/layout):** code that is reusable within the module but not exported publicly belongs in a shared `internal/` package. Private helpers duplicated across sibling packages under `internal/platform/` are the exact use case for an `internal/platform/config/` shared utility — the `platform/config` package already exists and already owns the canonical config-loading concern.

**Go proverb "a little copying is better than a little dependency" (Rob Pike, Gopherfest SV 2015):** this proverb applies when the dependency is a third-party or cross-module package and the copied code is truly trivial. Here both callers are `internal/platform/*` — same module, no external dependency is introduced by consolidation. The proverb does not protect cross-package duplication within the same internal module tree. More importantly, the implementations have already diverged (`parseBoolEnv` truthy-value set), which is the exact failure mode the proverb acknowledges: copying works only if the copies never need to change together.

**REQ-TOP-3 (adjacent):** "every platform package either has production consumers or does not exist." While not a strict REQ-TOP-3 violation (both packages have consumers), the register correctly identifies this as a hygiene item adjacent to RF-7.

### Verdict: SIMPLIFY (P3)

The finding is **real but low-severity**, correctly classified. The `splitCSV` duplication is incidental (identical implementations, no drift). The `parseBoolEnv` divergence is the sole concrete risk: one package accepts `yes`/`on`, the other does not. No production bug is confirmed, but the inconsistency will trip an operator who sets a truthy value other than `true` or `1`.

The fix is simple: export `ParseBoolEnv(name string, defaultValue bool) bool` from `internal/platform/config` (or a new `internal/platform/config/env.go` file) using the **more permissive** semantics (`true`, `1`, `yes`, `on`) since that is the standard POSIX/shell convention for boolean env vars, and update the two callers in `attachments.go` and `authn/config.go`.

Do NOT export `splitCSV` — it is truly trivial (5 lines) and the copies are identical. The Go proverb applies here: the copies are not worth a shared function because there is zero drift risk and the function is too small to have meaningful independent variation.

### Smallest correct fix

1. Add `func ParseBoolEnv(name string, defaultValue bool) bool` to `internal/platform/config/env.go` (new file, ~8 lines) using the 4-value truthy set (`1`, `true`, `yes`, `on` case-insensitive).
2. Replace the `parseBoolEnv` call in `internal/platform/config/attachments.go:94-95` with `config.ParseBoolEnv(...)`. (Self-import: move the function within the same package, or export from a subfile — both are valid since `attachments.go` is already in `package config`.)
3. Replace the `parseBoolEnv` calls in `internal/platform/authn/config.go:106,139,141` with `config.ParseBoolEnv(...)`.
4. Delete the private `parseBoolEnv` from both files.
5. Leave `splitCSV` duplicated — the proverb applies; both copies are 5-line identical functions.

**Note:** `attachments.go` is already in `package config`, so `parseBoolEnv` there is a private function in the same package. The cleanest fix is simply to keep it in `package config` and delete the one in `authn/config.go`, then have `authn/config.go` call `config.parseBoolEnv` — but since that's unexported, it must be exported as `ParseBoolEnv`. One exported function, one file (`env.go`), three call sites updated.

### Effort / blast radius

Effort: **S** (one new 8-line file, three call-site changes, no logic change). Blast radius: **contained** (changes are within `platform/config` and `platform/authn` — no domain module or binary is affected).

### ADR needed?

No. This is a hygiene fix within the platform layer.

### Over-engineering check

Do NOT create a `platform/envutil` package, a `platform/config/loader` abstraction, or any generic config-loading framework. One exported function in the existing `platform/config` package is the entire correct solution.

---

## Summary Table

| Finding | Verdict | Priority | Effort | Blast radius | ADR needed |
|---|---|---|---|---|---|
| F-06a: `platform/observability` imports `auth/domain` | REFACTOR | P0-prerequisite | S | contained | No |
| F-06b: Delivery-layer raw SQL in CD + documents handlers | REFACTOR | P1 | M | module | No |
| F-06c: `auth` repo writes to `iam_users` | REFACTOR | P1 | M | cross-module | No |
| F-06d: `iam/delivery` imports `auth/infrastructure/postgres` | REFACTOR | P1 | S | module | No |
| F-06e: `taxonomy.TemplateVersionChecker` queries templates tables | REFACTOR | P2 | S | cross-module | No |
| F-05 + D-04: Delete `security.RateLimiter`; activate `platform/ratelimit` | REFACTOR | P1 | M | module | No |
| F-15: `parseBoolEnv` drift | SIMPLIFY | P3 | S | contained | No |
| F-15: `splitCSV` duplication | KEEP | P3 | — | — | No |

---

## External citations

- Cockburn, A. (2005). "The Hexagonal (Ports & Adapters) Architecture." https://alistair.cockburn.us/hexagonal-architecture/
- Evans, E. (2003). *Domain-Driven Design*. Addison-Wesley. Fowler summary: https://martinfowler.com/bliki/BoundedContext.html
- Pike, R. (2015). Go Proverbs, Gopherfest SV. https://go-proverbs.github.io/
- Go module layout documentation: https://go.dev/doc/modules/layout
- golang-standards/project-layout `internal/` convention: https://github.com/golang-standards/project-layout
