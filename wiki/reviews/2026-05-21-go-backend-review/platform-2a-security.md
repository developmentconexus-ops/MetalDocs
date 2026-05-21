# Module #2a — Platform Security Boundary

**Initiative:** [`2026-05-21-go-backend-review`](../2026-05-21-go-backend-review.md)
**Scope:** `internal/platform/{authn,security,idempotency,ratelimit,tenant,problem,httpresponse}`
**Reviewers:** `ecc:go-reviewer`, `ecc:security-reviewer`, `ecc:silent-failure-hunter`, `ecc:type-design-analyzer`, `ecc:database-reviewer`
**Date:** 2026-05-21
**Status:** Findings consolidated (append-only). Fix tracking in tracker row.

Each finding is attributed to the agent(s) that surfaced it: `[g]` go-reviewer, `[s]` security-reviewer, `[sf]` silent-failure-hunter, `[t]` type-design-analyzer, `[db]` database-reviewer.

A finding flagged by multiple agents indicates higher signal; severity normalized to the highest of the lenses.

---

## Critical

### C1 — `X-Forwarded-Proto` trusted unconditionally → CSRF origin check bypass `[s,g,sf]`

**File:** [internal/platform/security/origin_protection.go:100](../../../internal/platform/security/origin_protection.go) (`sameOrigin`)

`sameOrigin` overwrites the request scheme with the raw `X-Forwarded-Proto` header value before allowlist comparison. Any directly reachable client can set `X-Forwarded-Proto: https` and craft an `Origin: https://allowed-host` that matches the canonical form — bypassing CSRF protection on every mutating method that carries a session cookie.

**Recommend:** Add a `TrustedProxyCIDRs` field on `OriginProtectionConfig`; only read `X-Forwarded-Proto` when `r.RemoteAddr` parses to an IP inside that list. Default empty (no header trust) so misconfiguration fails closed. Same TrustedProxy concept also needed by `security/ratelimit.go:96` (`requestIdentity` fallback to `r.RemoteAddr` — see H4). Reuse `auth/application/service.go:542 remoteIP` resolution logic.

---

### C2 — `security.RateLimiter.byIdentity` map grows unbounded → in-process memory DoS `[s,g,sf,t]`

**File:** [internal/platform/security/ratelimit.go:28-37,70-88](../../../internal/platform/security/ratelimit.go)

`byIdentity map[string]windowCounter` only resets `windowStart` on the same key; entries are never deleted. Every unique IP (or user ID) seen since process start consumes a permanent slot. A burst of distinct source IPs (rotating-proxy DoS, broad scanner sweep) inflates the map until OOM. No size cap, no TTL sweep.

The parallel `ratelimit.Middleware.limiters sync.Map` at [internal/platform/ratelimit/middleware.go:36](../../../internal/platform/ratelimit/middleware.go) has the identical structural defect for per-user/per-route `*rate.Limiter` entries.

**Recommend:** Inside the existing locked section in `allow()`, sweep entries whose `windowStart + window < now` before the lock releases. For the sync.Map limiter, add a `time.Ticker`-driven background sweeper started by the constructor; record last-access time on each entry and evict after `2 × window`. Bound either implementation with a hard max-entries LRU as defense-in-depth.

---

### C3 — Idempotency store has no locking → concurrent same-key requests both execute the handler `[db]`

**File:** [internal/platform/idempotency/postgres_store.go:42-52](../../../internal/platform/idempotency/postgres_store.go) (`CheckReplay`), [internal/platform/idempotency/middleware.go:48-93](../../../internal/platform/idempotency/middleware.go) (`Require`)

`CheckReplay` is a plain `SELECT`. Two simultaneous requests with the same Idempotency-Key both see a miss, both execute the handler, both commit side effects. `ON CONFLICT DO UPDATE` in `RecordReplay` only prevents duplicate row storage — not duplicate execution. The store is a cache, not a guard. For document finalization or scheduled-supersede paths, this is a data-integrity violation: the client gets a 2xx and assumes single execution.

**Recommend:** Two-phase the operation inside a transaction:
1. `INSERT ... status='in_flight' ... ON CONFLICT DO NOTHING RETURNING tag` — if `RETURNING` yields a row, this thread owns the key; proceed to handler.
2. If the insert hit conflict, `SELECT FOR UPDATE` the existing row (loser waits), then `CheckReplay` for the persisted response.

Status column already supports this — see C4. After handler completes, `UPDATE ... SET status='completed', response_status=..., response_body=...` instead of UPSERT.

---

### C4 — Idempotency schema defines `in_flight` / `failed` states but Go code never writes them `[db]`

**File:** [migrations/0147_idempotency_keys.sql:14,20](../../../migrations/0147_idempotency_keys.sql), [internal/platform/idempotency/postgres_store.go](../../../internal/platform/idempotency/postgres_store.go) (whole file)

The DDL's `CHECK (status IN ('in_flight','completed','failed'))` and the comment ("janitor sweep") show the original design intended a two-phase write. The Go code only ever writes `'completed'`. The states that would have implemented C3's locking strategy are dead schema.

**Recommend:** Implement the missing states as the fix for C3 (insert `in_flight` first, transition to `completed`/`failed`). If the design has been abandoned, strip `'in_flight'` and `'failed'` from the `CHECK` constraint and update the janitor comment — divergence between schema and code is itself a hazard.

---

### C5 — `authn.UserIDFromContext` returns `""` silently on missing key → cross-tenant list exposure `[sf]`

**File:** [internal/platform/authn/context.go:11-13](../../../internal/platform/authn/context.go); downstream callsite [internal/modules/controlleddocuments/application/service.go:375-379](../../../internal/modules/controlleddocuments/application/service.go)

`UserIDFromContext` returns a bare `string`. The controlled-documents `List` service treats empty actor as "no visibility filter" and the repository skips the per-user `WHERE` clause entirely — returning every document in the tenant when the context is missing the IAM key (middleware misordering, new route without IAM middleware attached, future refactor regression). The `Get` path fails closed by accident (empty string matches no rows); `List` fails open.

**Recommend:** Change signature to `UserIDFromContext(ctx) (string, bool)` (matches `iamdomain.UserIDFromContext` and `authdomain.CurrentUserFromContext`). At the only `List` callsite, `if !ok { return nil, errInternal("actor user id missing") }`. The mutation path at `service.go:474-477` already enforces this pattern — apply the same discipline to read paths. This is the most surgical fix; long-term, the `authn.RolesFromContext` re-export should also be deleted (see M6) and callers should consume `iamdomain` directly.

---

## High

### H1 — Idempotency `RecordReplay` write error silently swallowed → broken idempotency on transient DB error `[g,sf,db]`

**File:** [internal/platform/idempotency/middleware.go:89](../../../internal/platform/idempotency/middleware.go)

`_ = store.RecordReplay(...)`. Response already shipped; a transient DB error leaves no row. Next retry with the same key sees a miss and re-executes the handler — duplicate side effect with the client believing idempotency held.

**Recommend:** `if err := store.RecordReplay(...); err != nil { slog.ErrorContext(r.Context(), "idempotency: record failed — retry may duplicate", "key", key, "tenant", tenantID, "err", err) }`. Same logger pattern as `apps/api/cmd/metaldocs-api/main.go` shutdown path. Surfacing this is also a prerequisite for any alerting that wants to catch silent idempotency collapse.

---

### H2 — `ratelimit.Middleware.Limit` bypasses quota when `userExtractor` returns `""` → unauth callers get unlimited quota on misordered routes `[sf,g]`

**File:** [internal/platform/ratelimit/middleware.go:29-33](../../../internal/platform/ratelimit/middleware.go)

Empty user → `next.ServeHTTP`, no rate limit, no log. The inline comment ("IAM middleware should have rejected already") is convention, not enforcement. Any future public route given a rate limiter without considering middleware order silently disables the limiter for unauthenticated traffic.

**Recommend:** Fall back to IP-keyed limiting using the trusted-proxy resolution from C1 instead of bypassing. If the route is intentionally anonymous-allowed, IP is the right key; if not, the IAM middleware will already block — but the failure mode must be fail-closed, not fail-quiet. Add `slog.DebugContext` so the bypass condition is observable in tests.

---

### H3 — `responseRecorder` does not implement `http.Flusher` → streaming endpoints silently buffer or panic `[g]`

**File:** [internal/platform/idempotency/middleware.go:101-116](../../../internal/platform/idempotency/middleware.go)

The wrapper exposes only `WriteHeader` and `Write`. Any wrapped handler that does `w.(http.Flusher).Flush()` fails its type assertion (panic with `http.ResponseWriter` typed nil receiver) or silently buffers (when written through to a non-stdlib writer). Idempotency-protected endpoints cannot be streaming-safe.

**Recommend:** Either implement `Flush()` (delegating to the wrapped `ResponseWriter` if it is a Flusher) and document that flushes are still captured into the body buffer — or refuse to wrap streaming handlers with an explicit `Content-Type: text/event-stream` opt-out at middleware entry. Make the contract explicit; today it is implicit.

---

### H4 — `security.RateLimiter` falls back to `r.RemoteAddr` for unauthenticated identity → behind a proxy, every anon caller shares one bucket `[s]`

**File:** [internal/platform/security/ratelimit.go:96-107](../../../internal/platform/security/ratelimit.go) (`requestIdentity`)

`r.RemoteAddr` is the upstream proxy IP in any deployment with a reverse proxy or load balancer. All pre-login traffic collapses into a single bucket and either rate-limits everyone together or (more likely) blows past the threshold immediately for legitimate users. This is the same trust-the-proxy problem as C1 — fix together.

**Recommend:** Share the trusted-proxy CIDR config from C1; on a trusted upstream, extract the leftmost `X-Forwarded-For` value (re-use `auth/application/service.go:542` `remoteIP`). When the upstream is untrusted, keep `RemoteAddr`. Document the chosen mode in `wiki/architecture/` or a dedicated security note.

---

### H5 — Attachment signer secret has no minimum length → HMAC-SHA256 keyed with sub-output entropy `[s,sf]`

**File:** [internal/platform/config/attachments.go:52-55](../../../internal/platform/config/attachments.go), [internal/platform/security/attachmentsigner.go:17-18](../../../internal/platform/security/attachmentsigner.go)

`LoadAttachmentsConfig` rejects empty secret but not short secret. NIST SP 800-107 / FIPS 198-1 require HMAC-SHA256 keys >= 32 bytes. A 4-byte secret silently passes today, producing brute-forceable signed URLs if any leak into logs.

**Recommend:** In `LoadAttachmentsConfig`, `if len(secret) < 32 { return cfg, fmt.Errorf("METALDOCS_ATTACHMENTS_SIGNING_SECRET must be at least 32 bytes") }`. Add the same guard inside `NewAttachmentSigner` (`panic` on short input) so the invariant holds even when the signer is constructed outside the config loader (tests, future sub-service).

---

### H6 — `AttachmentSigner.Sign` returns bare `string`; `time.Now` not injected `[g,t]`

**File:** [internal/platform/security/attachmentsigner.go:21-25,32](../../../internal/platform/security/attachmentsigner.go)

Two coupled defects:
1. `Sign` returns `string` and `BuildDownloadURL` returns `string`. No `SignedURL` value type couples `(attachmentID, expiresAt, signature)`; argument-position swaps between two string params at `Verify` callsites are silent.
2. `Verify` hard-codes `time.Now().UTC()`; expiry boundary cannot be deterministically tested.

**Recommend:** Define `type SignedURL struct { URL string; ExpiresAt time.Time }` returned by `BuildDownloadURL`, with `(s SignedURL) IsExpired(now time.Time) bool`. Add `now func() time.Time` field on `AttachmentSigner` defaulted to `time.Now`. Pattern already present in `security.RateLimiter` at [ratelimit.go:26](../../../internal/platform/security/ratelimit.go) — reuse it.

---

### H7 — CORS middleware lets non-preflight cross-origin requests with disallowed Origin reach handler `[s,g,t]`

**File:** [internal/platform/security/cors.go:57-64](../../../internal/platform/security/cors.go)

Preflight rejects with 403 (correct). Non-preflight with disallowed `Origin` falls through to `next.ServeHTTP` with no CORS headers set. Browsers suppress the response, but non-browser clients (curl, scripts, server-to-server) receive and act on it — and side effects on the server were already committed. CORS is not the only gate (`OriginProtection` covers cookied mutators), but a publicly mutable endpoint with no session would bypass both layers.

**Recommend:** Reject non-preflight requests with `Origin` set and not allowlisted: return 403 with `problem.Write(problem.CodeForbidden, "cross-origin request blocked")`. Also normalize allowlist + incoming `Origin` to lowercase (RFC 6454 — see L4) so case-mismatched config doesn't silently mis-match. Reuse `normalizeOrigin` from `origin_protection.go:124` instead of re-implementing.

---

### H8 — `METALDOCS_AUTH_ENABLED=false` disables session enforcement with no `APP_ENV` guard `[s]`

**File:** [internal/platform/authn/config.go:14-20,41](../../../internal/platform/authn/config.go)

`Enabled()` is a plain boolean env read. A misconfigured production deployment silently runs with authentication disabled. `CookieSecure` defaults correctly based on `APP_ENV` at [config.go:114](../../../internal/platform/authn/config.go) — the same defense should apply here.

**Recommend:** In `LoadRuntimeConfig`, `if !cfg.Enabled() && strings.ToLower(os.Getenv("APP_ENV")) != "local" { return cfg, errors.New("METALDOCS_AUTH_ENABLED=false only permitted when APP_ENV=local") }`. Fail-fast at startup so the failure is visible from logs, not from incident retro.

---

### H9 — `problem.Problem` / `FieldError` accept zero-value construction; `Code` typed as plain `string` `[t]`

**File:** [internal/platform/problem/problem.go:12-27](../../../internal/platform/problem/problem.go), [internal/platform/problem/codes.go:6-32](../../../internal/platform/problem/codes.go)

All fields public; `Problem{Status: 0}` is constructible and would cause `w.WriteHeader(0)` (undefined). `Code` is `string` — the catalog in `codes.go` is advisory; nothing prevents a handler from writing `problem.New(400, "validation_error", ...)` with wrong casing or fabricated codes.

**Recommend:** `type Code string` in `codes.go`; change catalog constants to that type; change `Problem.Code` and `FieldError.Code` to `Code`. Make `Problem` struct fields unexported and keep `New`/`WithDetail`/`WithFieldError` builders as the only construction path. `httpresponse.WriteError` signature also updates to take `problem.Code`. Existing callers using catalog constants compile unchanged; magic-string callers fail to compile (intended).

---

### H10 — `ratelimit.Config.Quotas` is a public mutable map with no validation; zero value panics `[t]`

**File:** [internal/platform/ratelimit/config.go:17-19](../../../internal/platform/ratelimit/config.go), [internal/platform/ratelimit/middleware.go:36](../../../internal/platform/ratelimit/middleware.go)

Any importer can mutate `cfg.Quotas[RouteExportPDF] = 0`. Middleware then computes `rate.Every(time.Minute / time.Duration(0))` -> integer divide by zero -> panic on first request. `DefaultConfig()` returns a value but the map header is shared.

**Recommend:** Unexport `quotas`; constructor `NewConfig(q map[RouteKey]int) (Config, error)` validates every value >= 1; accessor `Config.QuotaFor(k RouteKey) (int, bool)`. Defense-in-depth: in `middleware.go:36`, guard `if quota <= 0 { next.ServeHTTP(w, r); slog.Error("ratelimit: invalid quota", ...); return }` so a misconfigured route degrades to no-limit rather than crashing the process.

---

### H11 — Idempotency `actor_user_id TEXT` (should be UUID + FK) and unbounded `response_body JSONB` `[db]`

**File:** [migrations/0147_idempotency_keys.sql:8,13](../../../migrations/0147_idempotency_keys.sql)

- `actor_user_id TEXT NOT NULL` accepts any string including `""`. `tenant_id` is correctly `UUID` with FK; the actor column is not — silent invalid writes possible.
- `response_body JSONB` has no size limit. A multi-MB rendered-PDF response stored per call inflates the table.
- `JSONB` also wrong type for `[]byte` from `r.body.Bytes()` — `BYTEA` is correct; valid-JSON coercion fails on non-JSON 2xx responses.

**Recommend:** Migration: `ALTER COLUMN actor_user_id TYPE UUID USING actor_user_id::uuid` (after audit of existing data); add FK to the principal table; change `response_body` to `BYTEA`; add `CHECK (octet_length(response_body) <= 65536)`. Document the 64 KiB cap in middleware so callers know their stored response is truncated/refused above that size.

---

## Medium

### M1 — Idempotency middleware reads request body with no `MaxBytesReader` `[g,s,sf]`

**File:** [internal/platform/idempotency/middleware.go:24-41](../../../internal/platform/idempotency/middleware.go) (`RequestHash`)

`io.ReadAll(r.Body)` unbounded. Multi-GB POST exhausts memory before reaching the handler. `httpresponse.ReadJSON` at [response.go:20-22](../../../internal/platform/httpresponse/response.go) has the same gap.

**Recommend:** Wrap with `http.MaxBytesReader(w, r.Body, maxBodyBytes)` before `io.ReadAll`. Cap should match the largest expected non-attachment payload (1-4 MiB likely sufficient; attachments use signed-URL paths). Apply the same cap to `httpresponse.ReadJSON` so every handler using the helper inherits the protection.

---

### M2 — Idempotency `RecordReplay` UPSERT overwrites stored `payload_hash` on conflict `[g,s,db]`

**File:** [internal/platform/idempotency/postgres_store.go:67-82](../../../internal/platform/idempotency/postgres_store.go)

`ON CONFLICT DO UPDATE SET payload_hash = EXCLUDED.payload_hash`. After C3's locking is in place this race shrinks, but the UPDATE still allows a successful completed entry's authoritative hash to be replaced by a later concurrent first-request — `CheckReplay` then accepts a different request body under the same key without `ErrConflict` ever firing.

**Recommend:** Once the `in_flight` two-phase write from C3/C4 is implemented, the second writer hits an existing `in_flight` row and waits — the UPDATE shouldn't run at all. As a belt: `ON CONFLICT (...) DO UPDATE SET ... WHERE idempotency_keys.status <> 'completed'` so completed rows become immutable.

---

### M3 — `tenant.WithTenantID` accepts empty string; invariant enforced only at read `[t]`

**File:** [internal/platform/tenant/context.go:18-20](../../../internal/platform/tenant/context.go)

`WithTenantID(ctx, "")` produces a context where `ctx.Value(ctxKey{}) == ""` (present, not missing). `FromContext` returns `ErrTenantMissing` for empty — correct — but the asymmetry between Setter (accepts anything) and Getter (validates) means a caller bug stores a sentinel-looking populated context.

**Recommend:** In `WithTenantID`, `if strings.TrimSpace(tenantID) == "" { panic("tenant: WithTenantID called with empty tenantID") }`. Auth middleware is the only production caller; a panic is a load-bearing bug signal. `DevTenantID` passes fine.

---

### M4 — `authn.DevRoleMap` reads `os.Getenv` per request `[g]`

**File:** [internal/platform/authn/config.go:127](../../../internal/platform/authn/config.go)

Function name implies cached config but reads + parses env on every call. Allocation hot-spot in dev mode and a subtle TOCTOU on env change.

**Recommend:** Parse and cache during `LoadRuntimeConfig`; expose the map (or a `RoleMap()` accessor) on the loaded config struct. Move env access to one boot-time site.

---

### M5 — `authn.Config` (`authapp.Config`) exposes secret/password fields directly `[t]`

**File:** [internal/modules/auth/application/service.go:27-47](../../../internal/modules/auth/application/service.go) (definition), [internal/platform/authn/config.go:100-117](../../../internal/platform/authn/config.go) (construction)

`SessionSecret string` and `BootstrapAdminPassword string` are public fields. Any importer can construct `Config{}` literal with blank secret — validation lives only in `LoadRuntimeConfig`. No "validated" type distinguishes a sane config from a zero value.

**Recommend:** Either unexport sensitive fields and add accessor methods, or have `auth/application.New(cfg Config) (*Service, error)` re-validate `SessionSecret != ""` (when enabled) and other invariants at the boundary closest to use. Boundary re-validation is the smaller change and still closes the gap.

---

### M6 — `authn.RolesFromContext` re-exports `iamdomain.RolesFromContext` while erasing `[]Role -> []string` `[t]`

**File:** [internal/platform/authn/context.go:16-29](../../../internal/platform/authn/context.go)

The wrapper lowercases/trims and then returns `[]string`, losing the `iamdomain.Role` enum. Downstream consumers can no longer match against `iamdomain.RoleSystemAdmin` without a cast — a stringly-typed boundary re-introduced inside the codebase. `ratelimit/middleware.go:100` already proves the direct call works.

**Recommend:** Delete `authn.RolesFromContext`. Callers should call `iamdomain.RolesFromContext` directly. Same for `UserIDFromContext` once C5 is fixed.

---

### M7 — Origin-protection error response uses hard-coded `"trace-local"` trace ID `[g]`

**File:** [internal/platform/security/origin_protection.go:137](../../../internal/platform/security/origin_protection.go)

The `requestTraceID` helper at [security/ratelimit.go:110-115](../../../internal/platform/security/ratelimit.go) already exists. Origin-blocked requests should carry the real trace ID so incident investigation can correlate upstream logs.

**Recommend:** Extract `requestTraceID` into a small `internal/platform/security/trace.go` helper (or move to `httpresponse`), and call it from both writers. Two-line refactor.

---

### M8 — `LegacyHeaderEnabled` flag undocumented in config `[s]`

**File:** [internal/platform/authn/config.go:107](../../../internal/platform/authn/config.go)

An env-toggleable boolean affecting authentication behavior with no doc comment on what mechanism it enables. Operationally a privilege-escalation hazard if enabled to debug something else.

**Recommend:** Doc-comment the field and its callers; if it enables a deprecated credential-passing path, mark `// Deprecated:` and tie its disablement to a removal date. If no live callers exist, delete.

---

### M9 — `idempotency.Replay.Status int` not range-validated -> `WriteHeader(0)` possible on corrupt row `[t]`

**File:** [internal/platform/idempotency/postgres_store.go:13-16](../../../internal/platform/idempotency/postgres_store.go), middleware write at [middleware.go:80](../../../internal/platform/idempotency/middleware.go)

A row with `response_status = 0` (corrupt data, manual edit, future bug) causes an invalid HTTP write. No range check at load or write.

**Recommend:** In `CheckReplay`, `if status < 100 || status > 599 { return nil, fmt.Errorf("idempotency: corrupt status %d", status) }`. Add the same `CHECK (response_status BETWEEN 100 AND 599)` in the migration — both ends of the boundary covered.

---

### M10 — Janitor sweeps `'completed'` only; `'in_flight'` / `'failed'` rows leak `[db]`

**File:** [internal/modules/jobs/idempotency_janitor/job.go:22-27](../../../internal/modules/jobs/idempotency_janitor/job.go)

After C3/C4 lands, `in_flight` rows from crashed handlers must be reaped. Without it, the table grows; with C3's `FOR UPDATE` semantics, a wedged row could block legitimate retries indefinitely.

**Recommend:** Drop the `status = 'completed'` filter from the `WHERE` clause once C3 is implemented — keep only `WHERE expires_at < now()`. Add a separate query for `in_flight` rows where `expires_at < now() - interval '5 min'` with explicit `slog.Warn` (these are orphans from crashes; visibility matters).

---

### M11 — No FK from `idempotency_keys.tenant_id` to `tenants` `[db]`

**File:** [migrations/0147_idempotency_keys.sql:7](../../../migrations/0147_idempotency_keys.sql)

Every other multi-tenant table FKs back to the tenants anchor. Deleted-tenant idempotency rows orphan indefinitely.

**Recommend:** `ADD CONSTRAINT idempotency_keys_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE` (verify path to the tenants table — schema name from `wiki/database/`).

---

## Low

### L1 — `httpresponse.ReadJSON` and `WriteError` divergent shape vs the rest of the codebase `[g,s,sf,t]`

**Files:** [internal/platform/httpresponse/response.go:16-22](../../../internal/platform/httpresponse/response.go), [internal/platform/security/ratelimit.go:117-128](../../../internal/platform/security/ratelimit.go) (`writeAPIError`), [internal/platform/idempotency/middleware.go:119-123](../../../internal/platform/idempotency/middleware.go) (`writeErrJSON`)

Three different inline JSON encoder paths emit three different error shapes:
- `problem.Write` -> RFC 9457 `{"title","status","code","detail"}`
- `security.writeAPIError` -> `{"error":{"code","message","details","trace_id"}}`
- `idempotency.writeErrJSON` -> `{"code","message"}`

API clients must parse all three. Each also discards `json.Encode` error after `WriteHeader` — silent truncation on future non-marshallable additions.

**Recommend:** Unify on `problem.Write` everywhere; delete `security.writeAPIError` and `idempotency.writeErrJSON`. Add `slog.Error` if the encode fails so future regressions are visible. `security` package can import `problem` (no cycle). Tracks the same root cause as H9.

---

### L2 — `problem.Write` return ignored at every callsite `[g,sf,t]`

**File:** [internal/platform/problem/problem.go:73-82](../../../internal/platform/problem/problem.go)

Function returns `error` but `WriteHeader` already fired before the marshal failure could be reported. Every callsite does `_ = problem.Write(...)`. Misleading signature.

**Recommend:** Change to `func Write(w http.ResponseWriter, p *Problem)` and `slog.ErrorContext` internally on marshal failure (currently unreachable for built-in `Problem`, but a latent trap for future fields).

---

### L3 — `ReadJSON` does not check `Content-Type` `[g]`

**File:** [internal/platform/httpresponse/response.go:20-22](../../../internal/platform/httpresponse/response.go)

No `Content-Type: application/json` assertion before decode. Mildly relaxes the contract; non-issue today.

**Recommend:** Optional. If added, return 415 for non-JSON content types. Low priority next to M1's size cap.

---

### L4 — CORS allowlist match is case-sensitive on the host portion `[s]`

**File:** [internal/platform/security/cors.go:95-101](../../../internal/platform/security/cors.go); origin already lowercases scheme/host at [origin_protection.go:124-133](../../../internal/platform/security/origin_protection.go) (`normalizeOrigin`) — not used here.

`METALDOCS_CORS_ALLOWED_ORIGINS=https://App.Example.Com` silently fails to match the browser-sent `https://app.example.com`.

**Recommend:** Apply `normalizeOrigin` to both the allowlist entries (in `NewCORS`) and the incoming `Origin` header before map lookup. Folds together with H7 cleanup.

---

### L5 — Trace ID header reflected verbatim into JSON error body `[s]`

**File:** [internal/platform/security/ratelimit.go:110-115](../../../internal/platform/security/ratelimit.go) (`requestTraceID`)

Caller-supplied `X-Trace-Id` is trimmed and echoed. A 4 KiB or control-character-laden value lands in every rate-limit error body and downstream log ingestion.

**Recommend:** Validate as `[A-Za-z0-9-]{1,128}` before reflecting; otherwise generate a server-side trace ID. Consolidate with M7's helper extraction.

---

### L6 — `mac.Write` and `bytes.Buffer.Write` errors discarded with bare `_, _` `[sf]`

**Files:** [internal/platform/security/attachmentsigner.go:23](../../../internal/platform/security/attachmentsigner.go), [internal/platform/idempotency/middleware.go:115](../../../internal/platform/idempotency/middleware.go)

Both writers structurally never error today. The bare discard hides the audit intent — future replacements (HSM-backed HMAC, size-limited buffer) silently inherit the swallow.

**Recommend:** Add inline comment: `// hash.Hash.Write never errors; cast suppresses linter`. Cheap, prevents future regressions during refactors.

---

### L7 — `authn.RolesFromContext` drops whitespace-only roles silently `[sf]`

**File:** [internal/platform/authn/context.go:16-29](../../../internal/platform/authn/context.go)

`if strings.TrimSpace(string(role)) == "" && string(role) != ""` is never logged. A corrupt role surfaces as a missing permission, indistinguishable from a legitimate empty role set.

**Recommend:** Folded by M6 (delete this wrapper). If retained, `slog.Warn` on non-empty input that normalizes to empty.

---

### L8 — `tenant.DevTenantID` is untyped `string` `[t]`

**File:** [internal/platform/tenant/const.go:4](../../../internal/platform/tenant/const.go)

Mixes with any other UUID-shaped string. Aligns with the larger typed-ID pattern recommendation.

**Recommend:** `type TenantID string`; `const DevTenantID TenantID = "..."`. Threads through `WithTenantID` / `FromContext`. Low ROI for internal-only code but consistent with `iamdomain.Role` / `Capability` precedent.

---

## Out of Scope (Deferred to Other Modules)

These were surfaced by `[s]` but belong to module reviews that haven't started:
- **Internal error strings leaked through `err.Error()` to clients** in `internal/modules/documents/delivery/http/handler.go:607,181,215`, `internal/modules/controlleddocuments/delivery/http/routes.go:28,60,181`, `internal/modules/controlleddocuments/delivery/http/handler.go:101`, `internal/modules/auth/delivery/http/handler.go:150`. Track under modules #3 (auth), #5 (documents), #6 (controlleddocuments).
- **`tenantIDFromContext` in controlleddocuments handler falls back to `DevTenantID`** at `internal/modules/controlleddocuments/delivery/http/handler.go:61-66` — violates the `tenant.FromContext` contract. Track under module #6.

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 5     |
| High     | 11    |
| Medium   | 11    |
| Low      | 8     |
| **Total** | **35** |

**Cross-cutting themes:**

1. **Trusted-proxy gap** (C1 + H4): both rate limiter and origin protection trust upstream headers unconditionally. One CIDR-driven config fix closes both.
2. **Unbounded in-memory state** (C2 + the `ratelimit.Middleware` parallel): two limiter implementations, both unbounded. Pick one canonical, add eviction.
3. **Two-phase idempotency never implemented** (C3 + C4 + M10): schema designed for it, code doesn't use it, janitor doesn't sweep it. Single coordinated fix unblocks all three.
4. **Stringly-typed boundaries everywhere** (C5, H9, H11, M3, M5, M6, L8): platform layer accepts raw `string` for tenant ID, user ID, error code, role, signed URL component. The IAM domain (`Role`, `Capability`) shows the pattern works — extend it down.
5. **Three divergent JSON error shapes** (L1): merge on `problem.Write`.

## Verification of Module #1 Fixes

Module #1 critical fixes from the cmd-metaldocs-api review are landed (commits `6eb31ec7`, `66fe1ee3`). No regressions surfaced by #2a agents on those areas.

## Next

Cursor advances to module #2b (`platform/{db,migrate,bootstrap,objectstore,storage,messaging,servicebus,jobs,worker}`).
