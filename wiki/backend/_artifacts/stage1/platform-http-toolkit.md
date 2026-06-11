# Stage-1 Audit Artifact — Platform HTTP Toolkit

> Area: `internal/platform/{httpresponse,problem,pagination,idempotency,requesttrace,useragent,httpclient,formval}`
> Produced: 2026-06-10 | Model: claude-sonnet-4-6 | Stage: 1 (truth-map only)

---

## 1. Identity & Purpose

The **platform HTTP toolkit** is the shared infrastructure layer that enforces the MetalDocs API design system contract across all HTTP-facing modules. It contains eight cohesive packages: `problem` implements the RFC 9457 Problem Details envelope with a typed error-code catalog and a catalog-guard test that prevents ad-hoc string codes from leaking into production; `httpresponse` wraps that envelope plus `encoding/json` into three zero-ceremony helpers used by every HTTP handler; `pagination` provides the single opaque keyset-cursor codec (base64 URL-safe, sort-value + id tuple) shared by all list endpoints; `idempotency` provides a Postgres-backed two-phase Require middleware and `Store` that serializes concurrent replay via `SELECT FOR UPDATE`; `requesttrace` injects and normalises a request trace-ID into `context.Context` with header-injection poisoning defence; `useragent` produces human-readable device labels from raw User-Agent strings for the Sessions & Security admin tab; `httpclient` vends a pre-tuned `*http.Client` for internal service fanout (Gotenberg, eigenpal); and `formval` wraps `santhosh-tekuri/jsonschema/v6` for server-side JSON Schema validation of document fill-in data. Together these packages constitute the API design system's runtime enforcement surface (wiki/architecture/api-design-system.md).

---

## 2. File Inventory

### `internal/platform/httpresponse/`
| File | Role |
|---|---|
| `response.go` | Three exported helpers: `WriteJSON`, `WriteError` (delegates to `problem.Write`), `ReadJSON` |
| `response_test.go` | Unit test verifying `WriteError` emits `application/problem+json` with correct status |

### `internal/platform/problem/`
| File | Role |
|---|---|
| `problem.go` | `Problem` struct (RFC 9457), `FieldError`, fluent builder (`WithDetail`, `WithInstance`, `WithFieldError`, `WithType`), `FromValidation`, `Write`, `Error` |
| `codes.go` | Canonical `Code` type + 40 exported constants covering HTTP-level, auth, domain, and field-level codes |
| `problem_test.go` | Unit tests: marshal/unmarshal, fluent chain, `FromValidation`, `Write`, `Error` interface, panic on invalid status |
| `codes_catalog_guard_test.go` | AST-based regression guard (`TestNoAdHocStringCodes`, `TestCanonicalCodeSelectorsExist`) — scans guarded packages at test time to enforce catalog discipline |

### `internal/platform/pagination/`
| File | Role |
|---|---|
| `cursor.go` | `EncodeCursor`, `DecodeCursor`, `ClampLimit`, `ErrInvalidCursor`, `DefaultLimit=20`, `MaxLimit=100` |
| `cursor_test.go` | Round-trip, URL-safety (B2 guard), blank=first-page, malformed, `ClampLimit` boundary tests |

### `internal/platform/idempotency/`
| File | Role |
|---|---|
| `middleware.go` | `Require` middleware factory (header validation, body hash, `BeginReplay`/`CompleteReplay`/`FailReplay` lifecycle, `responseRecorder`, `Flush` fail-closed); `IsValidKey`; `RequestHash`; `WithStreamingOptOut` |
| `postgres_store.go` | `Store`, `ReplayHandle`, `Replay`; `BeginReplay` (INSERT ON CONFLICT + SELECT FOR UPDATE two-phase protocol); `CompleteReplay`; `FailReplay`; `ErrConflict`; `maxBodyBytes=65536` |
| `middleware_test.go` | Integration tests: missing header, invalid UUID, first-call record, conflict 422, different path 422 |
| `middleware_concurrency_test.go` | Concurrency tests: same-key executes once (C3 fix), panic releases slot, non-2xx releases slot |
| `middleware_streaming_test.go` | Streaming tests: `Flush` without opt-out panics with directive, opt-out bypasses, `false` matcher still enforces |
| `two_phase_test.go` | Store unit tests: round-trip, conflict, fail-releases-slot, double-complete fails, invalid status rejected, concurrent same-hash serialises, concurrent diff-hash gives ErrConflict |
| `h11_schema_test.go` | H11 regression tests: empty actorID guard, oversized body (>64KiB) rejected, non-JSON body round-trips |

### `internal/platform/requesttrace/`
| File | Role |
|---|---|
| `context.go` | `WithTraceID`, `FromContext`, `Resolve` (generates UUID when absent), `Normalize` (printable ASCII, ≤128 chars, anti-header-injection) |
| `context_test.go` | Tests: rejects `\r\n` poison, reuses context trace-ID, generates UUID when absent |

### `internal/platform/useragent/`
| File | Role |
|---|---|
| `parse.go` | `Label(ua string) string` — best-effort browser+OS label; `Unknown` constant |
| `parse_test.go` | 10 table-driven cases covering Edge, Opera, Firefox, Chrome, Mobile Safari, iOS, curl, OS-only |

### `internal/platform/httpclient/`
| File | Role |
|---|---|
| `internal_client.go` | `NewInternalClient() *http.Client` — pre-tuned transport: dial 5s, TLS 5s, header 10s, idle 90s, HTTP/2 forced, MaxConnsPerHost 50 |
| `internal_client_test.go` | Verifies all transport field values |

### `internal/platform/formval/`
| File | Role |
|---|---|
| `gojsonschema.go` | `Gojsonschema` struct implementing `Validate(schemaJSON, formData) (bool, []string, error)` via `santhosh-tekuri/jsonschema/v6`; `flattenValidationErrors` recursively flattens `ValidationError.Causes` |

---

## 3. Public Surface

### Exported Types

| Package | Symbol | Kind | Consumed by |
|---|---|---|---|
| `problem` | `Problem` | struct | all HTTP handlers, `httpresponse` |
| `problem` | `Code` | type (`string`) | all handlers, `formval` output wiring |
| `problem` | `FieldError` | struct | documents, templates handlers |
| `problem` | `New`, `FromValidation`, `Write` | funcs | direct or via `httpresponse.WriteError` |
| `problem` | 40 `Code*` / `FieldCode*` constants | consts | every HTTP module |
| `httpresponse` | `WriteJSON`, `WriteError`, `ReadJSON` | funcs | 10+ handler files |
| `pagination` | `EncodeCursor`, `DecodeCursor` | funcs | documents, controlleddocuments, audit handlers + repos |
| `pagination` | `ClampLimit` | func | documents repo, audit handler |
| `pagination` | `ErrInvalidCursor`, `DefaultLimit`, `MaxLimit` | vars/consts | handlers, repos |
| `idempotency` | `Store`, `ReplayHandle`, `Replay` | structs | approval infra stores, controlleddocuments handler, templates handler, documents handler |
| `idempotency` | `New` | func | all idempotency consumers above |
| `idempotency` | `Require`, `WithStreamingOptOut` | funcs | controlleddocuments handler, templates handler |
| `idempotency` | `IsValidKey`, `RequestHash` | funcs | documents handler (manual idempotency, see §10) |
| `idempotency` | `ErrConflict` | var | approval errors.go, documents handler |
| `requesttrace` | `WithTraceID`, `FromContext`, `Resolve`, `Normalize` | funcs | `observability.http` middleware, `main.go`, auth handler |
| `useragent` | `Label`, `Unknown` | func/const | `iam/delivery/http/sessions_handler.go` only |
| `httpclient` | `NewInternalClient` | func | `main.go` (fanout client), worker `main.go` |
| `formval` | `NewGojsonschema`, `(*Gojsonschema).Validate` | func/method | `main.go` → `documents.Dependencies.FormVal` |

No HTTP routes are registered by any of these packages directly. All are pure libraries consumed by higher-level modules.

---

## 4. Logic Flows

### Flow 1 — Idempotency: First Request (Winner Path)

1. `Require` middleware called; `cfg.streamingOptOut` returns false — proceed.  
   `middleware.go:87`
2. Extract `Idempotency-Key` header; validate non-empty and UUID shape via `IsValidKey`.  
   `middleware.go:91–98`
3. Extract `tenantID, actorID` from context via caller-supplied `actorFromCtx`.  
   `middleware.go:101`
4. Wrap `r.Body` in `http.MaxBytesReader(1 MiB)`; call `RequestHash` which reads body, rewinds with `NopCloser`, and returns SHA-256(`method\npath?query\nbody`).  
   `middleware.go:103–113`, `middleware.go:28–45`
5. Call `store.BeginReplay(ctx, tenantID, actorID, key, hash)`.  
   `postgres_store.go:74`
6. Open a `sql.Tx`; execute `INSERT INTO metaldocs.idempotency_keys … ON CONFLICT DO NOTHING RETURNING tenant_id`.  
   `postgres_store.go:86–95`
7. `RETURNING` row is scanned → caller wins; return `(handle, nil, nil)`.  
   `postgres_store.go:96–99`
8. Back in middleware: `handle` non-nil, `replay` nil → create `responseRecorder` wrapping `w`, serve `next.ServeHTTP(rec, r)`.  
   `middleware.go:147–157`
9. On handler return: `rec.status >= 200 && < 300` → call `store.CompleteReplay(handle, rec.status, rec.body.Bytes())`.  
   `middleware.go:159–169`
10. `CompleteReplay` validates status in `[100,599]`, body ≤ 64 KiB, then `UPDATE … SET status='completed'` gated on `AND status='in_flight'`; commits tx.  
    `postgres_store.go:206–249`
11. `released = true` — deferred `FailReplay` is now a no-op.  
    `middleware.go:169`

### Flow 2 — Idempotency: Replay Hit (Loser Path)

1. Steps 1–6 identical to Flow 1. `INSERT … ON CONFLICT DO NOTHING` returns no rows.  
   `postgres_store.go:100`
2. Loser executes `SELECT … FOR UPDATE` — blocks until winner's tx commits.  
   `postgres_store.go:114–122`
3. `storedHash == payloadHash` — no conflict. `status = 'completed'` branch: read `response_status`, `response_body`.  
   `postgres_store.go:134–151`
4. Return `(nil, &Replay{…}, nil)`.  
5. Middleware: `replay != nil` → write `Content-Type: application/json` + `Idempotent-Replay: true` header + status + body directly to `w`; handler never called.  
   `middleware.go:127–132`

### Flow 3 — RFC 9457 Error Response Assembly

1. Handler (or middleware) decides an error should be returned; calls `httpresponse.WriteError(w, status, problem.CodeNotFound, "not found")`.  
   `response.go:16–18`
2. `WriteError` calls `problem.New(status, code, message)` — panics if `status ∉ [100,599]`.  
   `problem.go:31–40`
3. `problem.Write(w, p)` marshals `Problem` to JSON, sets `Content-Type: application/problem+json`, writes status, writes body.  
   `problem.go:77–87`
4. The catalog guard `TestNoAdHocStringCodes` (running at `go test`) rejects any raw string literal in the `code` argument position inside guarded packages, enforcing that only `problem.Code*` constants are used.  
   `codes_catalog_guard_test.go:142–178`

### Flow 4 — Keyset Cursor Encode / Decode

1. After fetching `limit+1` rows, handler (or repo) takes the last item's sort key (e.g. `UpdatedAt.UTC().Format(time.RFC3339Nano)`) and its `ID`.  
   `documents/delivery/http/handler.go:211`
2. Calls `pagination.EncodeCursor(sortValue, id)` → `base64.RawURLEncoding(sortValue + "|" + id)`.  
   `cursor.go:47–49`
3. Returned `cursor` string is embedded in the JSON response as `next_cursor`.
4. On the next request, handler calls `pagination.DecodeCursor(cursor)` → trims whitespace; blank → `("", "", nil)` (first page); else base64-decode, split on `|`, validate both parts non-empty.  
   `cursor.go:54–68`
5. `ErrInvalidCursor` returned to caller on decode failure; handlers map it to `problem.CodeInvalidCursor` + HTTP 400.  
   `documents/delivery/http/handler.go:199`, `controlleddocuments/delivery/http/routes.go:45`

### Flow 5 — Request Trace-ID Injection

1. `HTTPObservability.Wrap` (outer middleware) reads `X-Trace-Id` header; calls `requesttrace.Normalize(raw)` — rejects header if empty, >128 chars, or contains non-printable/non-ASCII bytes.  
   `observability/http.go:61–65`, `requesttrace/context.go:41–52`
2. If accepted: `requesttrace.WithTraceID(ctx, traceID)` stores it in context via unexported key.  
   `context.go:12–19`
3. If absent or rejected: `requesttrace.Resolve(ctx)` generates a fresh UUID.  
   `context.go:34–39`
4. Inner handlers and `main.go::traceIDFromContext` retrieve it via `requesttrace.Resolve(ctx)` (reuses if present, generates if absent).  
   `main.go:832`
5. `traceID` is emitted on every `slog` log line as `trace_id`.  
   `observability/http.go:95`

---

## 5. Dependencies

### Outbound (imports)

| Package | External / Internal | Why |
|---|---|---|
| `problem` | `encoding/json`, `fmt`, `net/http` | Marshal problem struct, write HTTP response |
| `httpresponse` | `encoding/json`, `net/http`; `metaldocs/internal/platform/problem` | Delegating to `problem.Write` for errors, JSON encode/decode |
| `pagination` | `encoding/base64`, `errors`, `strings` (stdlib only) | Pure keyset cursor codec, no external deps |
| `idempotency` | `bytes`, `context`, `crypto/sha256`, `database/sql`, `encoding/hex`, `errors`, `fmt`, `io`, `log/slog`, `net/http`, `regexp`; `metaldocs/internal/platform/problem` | Two-phase Postgres store, request hash, RFC 9457 errors |
| `requesttrace` | `context`, `strings`; `github.com/google/uuid` | UUID generation for missing trace-IDs |
| `useragent` | `strings` (stdlib only) | Pure string matching |
| `httpclient` | `net`, `net/http`, `time` (stdlib only) | Transport configuration |
| `formval` | `bytes`, `encoding/json`, `fmt`; `github.com/santhosh-tekuri/jsonschema/v6` | JSON Schema validation |

### Inbound (verified by grep)

| Package | Inbound consumers |
|---|---|
| `httpresponse` | `modules/audit`, `modules/auth`, `modules/controlleddocuments`, `modules/documents`, `modules/iam`, `modules/search`, `modules/security`, `modules/taxonomy`, `modules/templates` (10 files) |
| `problem` | All of the above + `platform/idempotency`, `platform/ratelimit`, `platform/security` (32 files total) |
| `pagination` | `modules/audit/delivery/http`, `modules/audit/infrastructure/postgres`, `modules/controlleddocuments/delivery/http`, `modules/controlleddocuments/infrastructure`, `modules/documents/delivery/http`, `modules/documents/repository` (7 files) |
| `idempotency` | `modules/controlleddocuments/delivery/http`, `modules/documents/delivery/http`, `modules/documents/approval/infrastructure`, `modules/documents/approval/application`, `modules/documents/approval/http`, `modules/templates/delivery/http` (14 files) |
| `requesttrace` | `platform/observability/http`, `modules/auth/delivery/http`, `apps/api/cmd/metaldocs-api/main.go` (4 files) |
| `useragent` | `modules/iam/delivery/http/sessions_handler.go` only (1 file) |
| `httpclient` | `apps/api/cmd/metaldocs-api/main.go`, `apps/worker/cmd/metaldocs-worker/main.go` (2 files) |
| `formval` | `apps/api/cmd/metaldocs-api/main.go` only — injected as `documents.Dependencies.FormVal` (1 file) |

---

## 6. Persistence

### `idempotency`
Table: `metaldocs.idempotency_keys`  
Columns (inferred from SQL in `postgres_store.go`):  
- `tenant_id TEXT`, `actor_user_id TEXT`, `route_template TEXT`, `key TEXT` — composite PK  
- `payload_hash TEXT`, `status TEXT` (`in_flight | completed | failed`), `expires_at TIMESTAMPTZ`  
- `response_status INTEGER NULL`, `response_body BYTEA NULL`  

Query patterns:
- INSERT with `ON CONFLICT DO NOTHING RETURNING` — winner detection. `postgres_store.go:86–95`
- `SELECT … FOR UPDATE` — blocks concurrent losers until winner's tx resolves. `postgres_store.go:114–122`
- `UPDATE … SET status='completed'` gated on `status='in_flight'`. `postgres_store.go:224–236`
- `DELETE … WHERE status='in_flight'` (FailReplay) and `WHERE status='failed'` (retry after fail). `postgres_store.go:265–277`, `183–195`

Migration reference: `maxBodyBytes = 64 * 1024` matches a CHECK constraint referenced as "migration 0204" in inline comment (`postgres_store.go:17`).  
Janitor job (`idempotency_janitor`) sweeps expired rows; registered in `main.go:537–541` when `ENABLE_JOB_IDEMPOTENCY_JANITOR != false`.

All other packages are stateless.

---

## 7. Config & Environment

| Package | Config / Env | Where parsed |
|---|---|---|
| `httpclient` | None — all transport values are compile-time constants | `internal_client.go:13–29` |
| `formval` | None | `gojsonschema.go` |
| `pagination` | None — `DefaultLimit=20`, `MaxLimit=100` are compile-time constants | `cursor.go:21–24` |
| `problem` | None | `problem.go`, `codes.go` |
| `httpresponse` | None | `response.go` |
| `requesttrace` | None | `context.go` |
| `useragent` | None | `parse.go` |
| `idempotency` | `maxIdempotencyRequestBodyBytes = 1 MiB` (compile-time); `maxBodyBytes = 64 KiB` (compile-time); TTL = `24 hours` (SQL literal in `postgres_store.go:91,229`) | `middleware.go:19`, `postgres_store.go:19` |

No environment variables are read by any of these eight packages.

---

## 8. Concurrency & Async

### `idempotency`
- `BeginReplay` opens a `*sql.Tx` per call; the transaction's row-level lock (`SELECT FOR UPDATE`) serialises concurrent goroutines racing for the same key. `postgres_store.go:78–200`
- `FailReplay` is always called from `defer` in `Require` middleware, so it executes on the serving goroutine after the handler returns or panics. `middleware.go:137–145`
- `responseRecorder.body` (`bytes.Buffer`) is written from the single serving goroutine; no synchronisation needed.
- `Flush()` on the recorder panics rather than silently buffering — fail-closed, no async race possible. `middleware.go:206–210`
- `BeginReplay` recurses (tail-call style) on two transient conditions: winner rolled back (row vanished after loser's `FOR UPDATE` read), and expired `in_flight` orphan after DELETE+commit. Maximum recursion depth bounded by real concurrency load; not a goroutine spawn. `postgres_store.go:127`, `173`, `195`

### All others
No goroutines, channels, or async primitives in `httpresponse`, `problem`, `pagination`, `requesttrace`, `useragent`, `httpclient`, or `formval`.

---

## 9. Error Handling & Observability

### Error patterns

- **`problem`**: every error response is an RFC 9457 `application/problem+json` body. `problem.New` panics on `status ∉ [100,599]` — a hard invariant, never a recoverable error. `problem.go:31–34`
- **`httpresponse`**: `WriteError` silently drops the marshal/write error (`_ = problem.Write(...)`). `response.go:17`
- **`idempotency`**: `writeErrJSON` calls `problem.Write`; on failure it logs a warning via `slog.Warn`. `middleware.go:216–219`. Successful replays set `Idempotent-Replay: true` header so clients can detect them. `middleware.go:128`. Persistence failures on `CompleteReplay` (response already written) are logged as `slog.ErrorContext` with message "retry may duplicate". `middleware.go:163–166`.
- **`pagination`**: `ErrInvalidCursor` is a sentinel error. Handlers map it to `problem.CodeInvalidCursor` + HTTP 400.
- **`formval`**: `flattenValidationErrors` recurses over `ValidationError.Causes`; returns `[]string` to caller. No logging inside the package.
- **`requesttrace`**: `Normalize` silently rejects bad trace-IDs and returns `("", false)`; callers fall through to UUID generation.

### Observability

- `idempotency` emits `slog.ErrorContext` on `BeginReplay` failure, handler panic, and `CompleteReplay` persistence failure; `slog.Warn` on `problem.Write` failure. All entries include `key` and `tenant` fields.
- `problem`, `httpresponse`, `pagination`, `useragent`, `httpclient`, `formval`, `requesttrace` emit no logs directly.
- No metrics, tracing spans, or Prometheus counters in any of these packages.

---

## 10. Legacy / Duplication / Smell Flags

- **Ad-hoc string codes in `idempotency.Require`**: `writeErrJSON` is called with raw string literals `"IDEMPOTENCY_KEY_REQUIRED"`, `"IDEMPOTENCY_KEY_INVALID"`, `"REQUEST_BODY_TOO_LARGE"`, `"BAD_REQUEST"`, `"IDEMPOTENCY_KEY_CONFLICT"`, `"INTERNAL"` (`middleware.go:93,97,108,111,117,123`). These strings are coerced to `problem.Code` via `problem.Code(code)` cast and bypass the catalog guard (the guard covers `internal/modules/documents/…` and `internal/modules/templates/…` only, not `internal/platform/idempotency`). Most map to canonical constants in `codes.go` (`CodeIdempotencyKeyRequired`, `CodeIdempotencyKeyReused`), but `"IDEMPOTENCY_KEY_INVALID"`, `"BAD_REQUEST"`, and `"INTERNAL"` have no catalog entry. WHY SUSPECT: inconsistency with the catalog discipline enforced everywhere else; `"INTERNAL"` is especially fragile as it diverges from `CodeInternalError = "INTERNAL_ERROR"`. The inline comment on `writeErrJSON` (line 212–215) acknowledges this is a Phase D sweep miss.
- **`idempotency.Require` not used on `documents/delivery/http` finalize route**: documents handler wires idempotency manually for the finalize path (`handler.go:445–475`) — validates key, hashes body, and calls `store.BeginReplay` / `CompleteReplay` / `FailReplay` inline rather than via the `Require` middleware. WHAT: bespoke inline idempotency wiring. WHERE: `internal/modules/documents/delivery/http/handler.go:440–495`. WHY SUSPECT: creates two divergent idempotency implementations (middleware vs. manual), making it harder to change the protocol and easy to miss the FailReplay defer. All other routes use `Require` or the approval infra stores.
- **`idempotency` guard scope excludes the approval HTTP package**: the catalog guard (`codes_catalog_guard_test.go:31–35`) guards `documents/delivery/http` and `templates/delivery/http` but not `documents/approval/http`. The approval package defines its own dot-notation code taxonomy (e.g. `"idempotency.key_required"`, `"internal.unknown"`) as `problem.Code` constants (`errors.go:27–55`). This is a deliberate documented exclusion ("dotted-taxonomy packages … are intentionally excluded", `codes_catalog_guard_test.go:32–34`), but it means the guard does not prevent the approval package from ever using `problem.CodeForbiddenCapability` vs. `"conflict.stale_revision"` inconsistently. WHAT: guard scope gap. WHERE: `internal/platform/problem/codes_catalog_guard_test.go:31–35`. WHY SUSPECT: the exclusion comment is accurate but the divergent code taxonomy is not documented in the API design system wiki.
- **`pagination` not adopted by templates list endpoint**: the templates audit listing uses offset-based pagination (`limit`/`offset` query params, `routes_query.go:211–223`) with a `TODO(pagination)` comment flagging the migration work. WHAT: non-uniform list endpoint behaviour. WHERE: `internal/modules/templates/delivery/http/routes_query.go:222`. WHY SUSPECT: all other list endpoints (`documents`, `controlled-documents`, `audit`) use keyset cursor; templates is the only remaining consumer of offset-based pagination. RF candidate.
- **`formval` has only one consumer and no interface abstraction at the platform level**: `NewGojsonschema` returns the concrete `*Gojsonschema` type. The documents module wires it via the `documents.Dependencies.FormVal` field, which is typed as an unexported interface in `documents/module.go`. WHAT: concrete type exposed from platform. WHERE: `internal/platform/formval/gojsonschema.go:12–14`. WHY SUSPECT: the `*Gojsonschema` is stateless and recompiles the schema on every call (new `Compiler` instance per `Validate` invocation at `gojsonschema.go:17`), which wastes allocation for static schemas. Single-use but no caching option.
- **`useragent` has only one consumer**: `Label` is imported only from `iam/delivery/http/sessions_handler.go`. WHAT: narrow utility package. WHERE: `internal/platform/useragent/parse.go`. WHY SUSPECT: not a flag per se, but the package's scope is narrow; if the Sessions & Security tab is the only use case, the package may be better colocated with the IAM module. Not a correctness issue.
- **`idempotency` TODO: missing tenant_id FK**: inline TODO in `postgres_store.go:54` acknowledges that `tenant_id` lacks a FK constraint in the `idempotency_keys` schema; also `postgres_store.go:87` notes column type tightening is pending. WHAT: schema enforcement gap. WHERE: `internal/platform/idempotency/postgres_store.go:54,87`. WHY SUSPECT: application-layer isolation is in place, but schema-level FK protection against orphaned tenant references is absent.
- **`idempotency` TTL is hardcoded as a SQL string literal**: the 24-hour TTL appears as the SQL string `'24 hours'` in two places (`postgres_store.go:91` and `postgres_store.go:229`). WHAT: magic literal duplication. WHERE: `internal/platform/idempotency/postgres_store.go:91,229`. WHY SUSPECT: if the TTL needs changing it requires two edits and there is no named constant.
- **`httpresponse.WriteError` silently discards write error**: `_ = problem.Write(...)` discards the error returned from `problem.Write`. WHERE: `internal/platform/httpresponse/response.go:17`. WHY SUSPECT: a failed response write (client disconnect, broken pipe) is silently dropped with no logging. Callers cannot detect partial writes. Contrast with `idempotency.writeErrJSON` which at least logs a warning.
- **`normalizeRoute` in `platform/observability/http.go` is a hardcoded prefix-switch**: the route-normalisation function that groups metrics by route template uses manual `strings.HasPrefix`/`strings.HasSuffix` checks for a fixed list of known path shapes (`observability/http.go:178–209`). WHAT: out-of-band route registry. WHERE: `internal/platform/observability/http.go:178–209`. WHY SUSPECT: this is not strictly in scope for the eight packages under audit, but `requesttrace` feeds into this same file; the normalisation does not use `net/http`'s `ServeMux` pattern, so adding a new parameterised route requires an update here or metrics will report raw paths instead of grouped routes.

---

## 11. Wiki Drift

No existing wiki document covers this area. Section 11 per task spec: **No existing doc**.

---

## 12. Open Questions

- **[runtime-unverified]** The idempotency janitor job (`idempotency_janitor`) is registered when `ENABLE_JOB_IDEMPOTENCY_JANITOR` env var is not `false`. It is not a test subject of this audit. Its actual sweep query and scheduling frequency are in `internal/modules/jobs/idempotency_janitor/` (not read). Whether expired `in_flight` rows that survive past TTL without the janitor cause operational issues is unverified without a running instance.
- **[runtime-unverified]** The `formval.Gojsonschema.Validate` creates a new `jsonschema.Compiler` on every invocation (`gojsonschema.go:17`). Whether this causes measurable allocation pressure under concurrent load is unverified without profiling data.
- **[runtime-unverified]** The `idempotency.BeginReplay` recursive call path (loser after winner rollback, or expired orphan reclaim) has no depth limit. In theory, a high-contention scenario where winners repeatedly roll back could cause deep recursion. Whether this is observable in practice requires load testing.
- The `codes_catalog_guard_test.go` guards only two directories and `pdf_webhook_handler.go`. The `documents/approval/http` package uses a distinct dot-notation code taxonomy. It is unclear whether the two vocabularies (`SNAKE_UPPER` catalog vs. `dot.notation` approval codes) represent an intentional contract boundary or a migration defer. The guard file comment says "intentionally excluded" but no ADR is referenced.
- The `maxIdempotencyRequestBodyBytes` cap (1 MiB in middleware) and `maxBodyBytes` cap (64 KiB in store) are independent limits. A request body between 64 KiB and 1 MiB will be hashed, accepted by `BeginReplay`, the handler will run, and `CompleteReplay` will reject the body — causing silent idempotency loss for that request with the error logged but the 2xx response already delivered to the client. This design choice is documented in the `maxBodyBytes` comment (`postgres_store.go:17–22`) but the asymmetry between the two limits may surprise future maintainers.
