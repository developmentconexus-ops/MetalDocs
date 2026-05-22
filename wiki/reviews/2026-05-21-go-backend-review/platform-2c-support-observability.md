# Module 2c — `internal/platform/{config,observability,cache,featureflags,formval,httpclient,pagination,docgenv2,render}`

**Reviewed:** 2026-05-22
**Reviewers (ECC, Sonnet 4.6):** go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer
**Scope notes:**
- `internal/platform/cache/` — empty placeholder, no `.go` files; no findings.
- `internal/platform/config/docgen_v2.go` — already filed under 2b-C3 (SSRF) and 2b-C4 (empty token). Skipped here to avoid duplication.
- Tests reviewed alongside production code.

## Severity counts (deduped)

| Critical | High | Medium | Low |
|----------|------|--------|-----|
| 1        | 14   | 25     | 18  |

---

## Critical

### C1 — `internal/platform/pagination/cursor.go:37-43` — Cursor lacks HMAC; anchor tampering bypasses keyset bounds
**Lenses:** go-reviewer (High → escalated), security-reviewer (Low), type-design-analyzer (High)
**Problem:** `Encode` wraps cursor JSON in base64 without MAC. `Anchor` (`map[string]any`) carries keyset-pagination column values. A client decodes, mutates anchor, re-encodes — `Decode` + `ValidateMatch` only check version, sort, and filter-hash. Anchor itself is fully attacker-controlled. If repository layer doesn't re-validate tenant scope at the SQL bind site, this leaks cross-tenant rows at page boundaries.
**Why Critical:** MetalDocs is a multi-tenant regulatory document system. Tenant-leak via tampered cursor is data-loss-grade.
**Recommend:** HMAC-SHA256 sign the encoded cursor with a server-side secret (reuse `AttachmentsConfig.DownloadSecret` or new `METALDOCS_CURSOR_SIGNING_SECRET`); verify in `Decode`. Defense-in-depth: every repository keyset query re-checks `tenant_id` at bind time regardless of anchor source.
**Fix branch reserved:** `fix/cursor-2c-c1`

---

## High

### H1 — `internal/platform/observability/http.go:73-79` — Torn atomic/mutex reads in `routeMetrics`
**Lens:** go-reviewer
**Problem:** `requests`/`errors`/`durationMs` written via `atomic.AddUint64`; `samples`/`cursor` under `m.mu`. `snapshot()` reads counters atomically without `m.mu`, so the avg derived from `durationMs/requests` can use values from different moments. No happens-before relative to `record()`.
**Recommend:** Pick one regime. Easiest: protect all fields with `m.mu` (sample window is 200 — lock contention negligible). Matches mutex-per-struct pattern elsewhere.

### H2 — `internal/platform/observability/runtime.go:169-190` — Shared `statsCtx` starves serial queries
**Lenses:** silent-failure-hunter (High), go-reviewer (Medium), database-reviewer (Medium)
**Problem:** Three sequential `QueryRowContext` calls share a single 3 s deadline. If query #1 burns 2.9 s, #2 and #3 inherit ~100 ms and silently cancel — auth filled, sessions/outbox zeroed, but zeros are returned alongside `errors` map (consumers reading `metrics["auth"]["users"]["active"]` see legit zeros).
**Recommend:** Independent `context.WithTimeout(ctx, 2*time.Second)` per query OR concurrent with `errgroup`. Write metric sub-maps only on query success; omit / mark `"unavailable"` on error.

### H3 — `internal/platform/observability/runtime.go:214-227` — Zero-valued metric maps written before error check
**Lens:** silent-failure-hunter
**Problem:** `metrics["auth"]`, `metrics["worker"]`, `metrics["sessions"]`, `metrics["outbox"]` populated with zeros before `if authErr != nil`. Consumers reading without checking `metrics["errors"]` get fake zeros as real data.
**Recommend:** Wrap each block: only assign sub-map on successful scan; otherwise emit `nil` / sentinel.

### H4 — `internal/platform/observability/runtime.go:128` — Wrong context passed to `applyDependencyChecks`
**Lens:** silent-failure-hunter
**Problem:** `PostgresRuntimeStatusProvider.Ready` derives `readyCtx` (3 s) but passes raw `ctx` to `applyDependencyChecks`. Dependency-check sub-timeouts derive from `ctx`, bypassing the readiness deadline.
**Recommend:** Pass `readyCtx` so all readiness work shares the same bound.

### H5 — `internal/platform/observability/runtime.go:51-54` — `StaticRuntimeStatusProvider.Ready` ignores dependency-check status pointers
**Lens:** go-reviewer
**Problem:** `StaticRuntimeStatusProvider.Ready` passes `nil` for `status *string` / `code *int` to `applyDependencyChecks`. Failed dependency check entry shows `"down"` but the top-level response still returns `200 ready`. Postgres provider does this correctly.
**Recommend:** Mirror the Postgres provider — pass live pointers so degraded checks propagate.

### H6 — `internal/platform/render/gotenberg/client.go:73,112` — Unbounded `io.ReadAll` on success body
**Lens:** security-reviewer
**Problem:** Success path `io.ReadAll(resp.Body)` has no size cap. Compromised/misconfigured Gotenberg can return arbitrary body → OOM in API process, all tenants affected. Error path already uses `LimitReader`; inconsistency is the smell.
**Recommend:** Wrap with `io.LimitReader(resp.Body, maxPDFBytes)` (config-driven, e.g. 50 MB) on both success paths; return distinct error if cap hit.

### H7 — `internal/platform/pagination/cursor.go:14` — `SortField.Direction string` lacks compile-time guard
**Lens:** type-design-analyzer
**Problem:** Direction validated only inside `Decode`. Anyone constructing `SortField{Direction: "lol"}` directly (e.g. tests, manual repo calls) bypasses validation.
**Recommend:** `type SortDirection string`; `const (SortAsc SortDirection = "asc"; SortDesc SortDirection = "desc")`; field becomes `Direction SortDirection`.

### H8 — `internal/platform/pagination/cursor.go:14-17` — `SortField.Field` un-allowlisted; SQL-injection surface
**Lens:** database-reviewer
**Problem:** `Field string` decoded from client cursor with no allowlist. If any SQL builder concatenates `ORDER BY ` + field, injection. Even safe builders depend on convention.
**Recommend:** Allowlist regex in `Decode` (`^[a-z][a-z0-9_]*$`) and tie to per-endpoint allowed-column set passed into `ValidateMatch`.

### H9 — `internal/platform/pagination/cursor.go:19,22` — `Cursor.Anchor map[string]any` open bag
**Lens:** type-design-analyzer
**Problem:** Anchor values must match column types. Open `any` lets callers stuff slices, nested maps, nil; JSON re-decode gives `float64` where DB expects bigint, `string` where DB expects `timestamptz` with TZ suffix.
**Recommend:** `type Anchor map[string]json.RawMessage` (forces typed unmarshal per field) OR generic `Cursor[A any]` with typed anchor structs per entity.

### H10 — `internal/platform/config/repository.go:10-14` — `RepositoryMode()` returns raw `string`
**Lens:** type-design-analyzer
**Problem:** Untyped constants ("memory" / "postgres") forces every consumer to re-switch on raw strings. Renames/typos silent.
**Recommend:** `type RepositoryMode string` + typed `const` block + `Valid() bool`. Rename loader to `LoadRepositoryConfig() (RepositoryMode, error)` for symmetry with sibling loaders.

### H11 — `internal/platform/config/attachments.go:17` — `Provider string` primitive obsession
**Lens:** type-design-analyzer
**Problem:** Storage provider re-validated downstream via string compare. Mismatched strings compile.
**Recommend:** `type StorageProvider string` + typed constants. Same treatment for `AppEnv string`.

### H12 — `internal/platform/config/gotenberg.go:8-18` — Raw URL accepted; SSRF on misconfig
**Lenses:** security-reviewer (Medium → escalated by overlap with H6), type-design-analyzer (High)
**Problem:** `LoadGotenbergConfig` stores raw env value without scheme/host validation. Operator injection or misconfig → `http://169.254.169.254` or `file://` → SSRF on every render. Same shape for `LoadDocgenConfig`'s `APIURL`.
**Recommend:** `url.ParseRequestURI`; reject empty host; allowlist `http`/`https`; store `*url.URL` or opaque `type ServiceURL string` with constructor. Apply to both Gotenberg and docgen configs.

### H13 — `internal/platform/observability/runtime.go:10` — `RuntimeStatusProvider` returns `map[string]any` everywhere
**Lens:** type-design-analyzer
**Problem:** Payload shape (`status`, `checks`, `repositoryMode`, ...) is contract-by-string-key. No compile-time guarantee; `.(type)` assertions to act on fields.
**Recommend:** Typed `LiveResponse`, `ReadyResponse`, `RuntimeMetrics` structs returned by the interface; `map[string]any` reserved for the JSON serialization layer.

### H14 — `internal/platform/docgenv2/template_reader.go:25` — 3-positional-string return swappable
**Lens:** type-design-analyzer
**Problem:** `GetPublishedVersion` returns `(docxKey, schemaKey, schemaJSON string, err error)`. Both keys typed `string`; caller swap → silent miscompile-passing.
**Recommend:** `type StorageKey string`, `type SchemaJSON string`; return struct `TemplateVersionAssets{ DocxKey StorageKey; SchemaKey StorageKey; SchemaJSON SchemaJSON }`.

### H15 — `internal/platform/docgenv2/templates_snapshot_reader.go:28-37` — Tenant-fallback asymmetry vs `TemplatesTemplateReader`
**Lens:** database-reviewer
**Problem:** `LoadForSnapshot` is tenant-strict (no system-template OR). `TemplatesTemplateReader.GetPublishedVersion` admits system tenant. Same documents reachable via one path, `sql.ErrNoRows` via the other. `FanoutTemplateReader` doesn't compose `TemplatesSnapshotReader`, so gap stays open.
**Recommend:** Either add `OR tpl.tenant_id = systemTemplateTenantID` to `LoadForSnapshot`, or explicitly document/test that system templates are never snapshot-loaded.

---

## Medium

| ID | Location | Problem | Recommend |
|----|----------|---------|-----------|
| M1 | `observability/http.go:152-169` | Optimistic RLock→Lock double-check is correct but undocumented; future contributor risk | Add comment naming the pattern |
| M2 | `observability/http.go:172-201` | Duplicate `HasPrefix("/api/v1/documents/")` (line 186 unreachable); dead suffix branch | Consolidate prefix+segment split into one pass |
| M3 | `observability/http.go:238-250` | `statusWriter.Write` reads `"Status"` response header that Go never sets — dead code | Drop the header-inspection branch; `status == 0 → 200` default is enough |
| M4 | `observability/http.go:86-101` | `r.URL.Path` logged raw — log-injection via newline/control chars in path segments; undermines audit trail | Use `r.URL.EscapedPath()` or strip `\r\n\t` via `strings.Map` |
| M5 | `observability/runtime.go:62-88,138-230` | `/health/ready` (public) exposes `repositoryMode`/`storageProvider`/`authEnabled` — internal-topology recon | Drop those fields from the public response; keep only on auth-gated metrics endpoint |
| M6 | `observability/runtime.go:243-283` | `applyDependencyChecks` uses `*string`/`*int` out-params | Return tuple `(checks []map[string]any, status string, code int)`; caller reassigns |
| M7 | `observability/runtime.go:26` | `PostgresRuntimeStatusProvider` embeds `*StaticRuntimeStatusProvider`; promoted methods/state leak risk | Compose explicitly; add `_ RuntimeStatusProvider = (*PostgresRuntimeStatusProvider)(nil)` compile-time assertion |
| M8 | `observability/runtime.go:15` | `DependencyCheckResult.Meta map[string]any` open bag (all real uses string) | `map[string]string` |
| M9 | `observability/http.go:34` | `metricItem` exported name + fields, but only used internally | Rename `metricSnapshot` (unexported) |
| M10 | `config/docgen.go:25-28` | Bad timeout silently → default 10s | `LoadDocgenConfig() (DocgenConfig, error)`; propagate `strconv.Atoi` error (matches sibling loaders) |
| M11 | `config/docgen.go (whole)` | Only `Load*Config` returning bare value; breaks package convention | Same as M10 |
| M12 | `config/feature_flags.go:22-29` | Bad rollout % silently → 0% — operator sees no signal | `(FeatureFlagsConfig, error)` propagating parse error |
| M13 | `config/worker.go:69-72` | "invalid RETRY_MAX_SECONDS" hides inter-constraint with RETRY_BASE | Include `must be >= RetryBaseSeconds (%d)` in error |
| M14 | `config/attachments.go:18` | `AppEnv string` repeated as primitive obsession across packages | `type AppEnv string` + validation |
| M15 | `config/postgres.go:11` | `DSN string` public-mutable post-load | Unexport + accessor |
| M16 | `config/jobs.go:14` | `Queues map[string]river.QueueConfig` mutable public map | Unexport; `QueueFor(name)`, `QueueNames()` accessors |
| M17 | `config/gotenberg.go:6` / `docgen.go:11` | Public fields allow post-load mutation (`Enabled=true` w/ empty `URL`) | Unexport + `Validate() error` |
| M18 | `config/postgres.go:29-30` | `sslmode=disable` default — credentials plaintext in transit | Default `require`; document `disable` is local-only |
| M19 | `docgenv2/templates_reader.go:33-36` | `GetPublishedVersion` returns bare `err` (no wrap context) | `fmt.Errorf("templates template reader: get published version %s: %w", id, err)` |
| M20 | `docgenv2/templates_reader.go:50-58` | `FanoutTemplateReader` secondary-fail error not wrapped | `fmt.Errorf("fanout template reader (secondary): %w", err)` |
| M21 | `docgenv2/template_reader.go:41-42` | nil client + non-empty schemaKey silently skips schema | Explicit error when `schemaKey != "" && t.client == nil` |
| M22 | `docgenv2/templates_reader.go:46` | `FanoutTemplateReader` holds concrete `*TemplateReader` / `*TemplatesTemplateReader` — untestable | Extract `TemplateVersionReader` interface; fanout holds two of those |
| M23 | `render/gotenberg/client.go:18-25` | No `CheckRedirect`; default follows ≤10 redirects → SSRF on misconfigured base URL | `CheckRedirect: func(...) error { return http.ErrUseLastResponse }` |
| M24 | `render/gotenberg/client.go:13-25` | `NewClient` accepts blank/schemeless baseURL — fails at first call; no test-double interface | Validate URL in constructor → `NewClient(baseURL) (*Client, error)`; declare `PDFConverter` interface at consumer |
| M25 | `formval/gojsonschema.go:11-29` | Zero-field struct method receiver disguising a function; schema param `string` vs `json.RawMessage`; bare `err` on non-ValidationError path | Define `Validator` interface; accept `json.RawMessage`; wrap non-ValidationError with `fmt.Errorf("jsonschema validate: %w", err)` |
| M26 | `pagination/cursor.go:103-123` | `+=` concat in loop — O(n²) (style violation per project Go rules) | `strings.Builder` |
| M27 | `pagination/cursor.go:19` | No constructor; zero value partially valid; direct literal bypass | `NewCursor(sort, anchor, filterHash)` setting `V: CursorVersion` |

---

## Low

| ID | Location | Note |
|----|----------|------|
| L1 | `observability/health.go:43` | `_ = json.Encode(...)` discards error after headers committed; log `Warn` |
| L2 | `observability/http.go:119` / `featureflags/handler.go:38` | Same `_ =` Encode pattern |
| L3 | `observability/http.go:84-88` | `X-Trace-Id` header logged uncapped — log inflation; cap at 128 chars |
| L4 | `observability/http.go:18` | Mixed atomic/mutex regime in one struct, undocumented; split into two embedded structs for clarity (cf. H1) |
| L5 | `render/gotenberg/client.go:69,108` | `respBody, _ := io.ReadAll(...)` on error path drops read error | `if readErr != nil { return nil, fmt.Errorf("gotenberg: status %d (body unreadable: %w)", resp.StatusCode, readErr) }` |
| L6 | `render/gotenberg/client_test.go:82` | Test handler discards body-read error → silent test passes | `if err != nil { t.Fatalf(...) }` |
| L7 | `docgenv2/templates_snapshot_reader_test.go:12` | Only constructor exercised; `LoadForSnapshot` untested | Add nil-db error case + sqlmock happy path |
| L8 | `httpclient/internal_client.go` (whole) | No `CheckRedirect`; default follows up to 10 redirects | Set `http.ErrUseLastResponse` (mirrors M23) |
| L9 | `httpclient/internal_client.go` | Returns bare `*http.Client` — no type wrapper; no functional-options for timeout | `type InternalClient struct { *http.Client }` or options pattern |
| L10 | `attachments.go:53` | `DownloadSecret` public, mutable post-load | Add `Validate()` re-check; document immutable-after-load |
| L11 | `worker.go:10` | Public fields allow post-load violation of cross-field invariant | Unexport + `Validate() error` for re-check on copy |
| L12 | `jobs.go:14` | Queue name `"temporal"` hard-coded magic string | `const QueueNameTemporal = "temporal"` |
| L13 | `docgenv2/template_reader.go:13` | `systemTemplateTenantID = "fff...ff"` bare string constant | `type TenantID string` if domain layer already has it |
| L14 | `formval/gojsonschema.go` (whole) | Schema source path lacks defense-in-depth ref-loader guard | Add ref-blocking `loader.Loader` even if schemas are system-owned today |
| L15 | `pagination/cursor.go:138` | `AppendIDTieBreaker` emits unqualified `id` — breaks in JOINs | Document contract that callers must qualify when JOINing |
| L16 | `pagination/cursor.go:19-24` | Anchor types ambiguous on JSON decode (`float64` vs UUID `string` vs TZ format) | Document anchor value types OR `map[string]json.RawMessage` (cf. H9) |
| L17 | `cmd/seed-test-document/main.go:25-28` | Hardcoded DSN + MinIO creds as Go const | Move to env / `.env` excluded from VCS; add build-tag guard so it can't ship to production |
| L18 | `docgenv2` reader queries | No `deleted_at IS NULL` filter (if templates use soft-delete) | Confirm schema; add filter if so |

---

## Notes for the cursor / next module

- `cache/` package empty (placeholder `.gitkeep` only). No findings, no follow-up — confirm whether the package is dead and either populate or remove.
- Two new fix-branch reservations needed: `fix/cursor-2c-c1` (C1) plus consider bundling H1-H5 observability fixes into `fix/observability-2c-h1-h5`.
- Pagination findings cluster (C1 + H7-H9 + M26-M27 + L15-L16) covers most of `cursor.go` — single coordinated rewrite would close 7 findings.
- Observability findings cluster (H1-H5 + M1-M9) covers most of `observability/` — second coordinated rewrite.
- Config findings cluster (H10-H12 + M10-M18 + L10-L13) reflects a missing `Validate() error` + typed-constants convention across the package; one templated refactor closes most of them.
