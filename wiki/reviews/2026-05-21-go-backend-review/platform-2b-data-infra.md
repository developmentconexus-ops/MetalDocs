# Platform #2b — Data + Infra Findings

**Module group:** `internal/platform/{db,migrate,bootstrap,objectstore,storage,messaging,servicebus,jobs,worker}`
**Review date:** 2026-05-22
**Reviewers (parallel ECC lenses):** `ecc:go-reviewer` [g], `ecc:security-reviewer` [s], `ecc:silent-failure-hunter` [sf], `ecc:type-design-analyzer` [t], `ecc:database-reviewer` [db]
**Build truth:** `go build ./internal/platform/{db,migrate,bootstrap,objectstore,storage,messaging,servicebus,jobs,worker}/...` — green. `go vet ./...` — green.
**LoC (hand-written, excl. tests/gen):** db 31, migrate 88, bootstrap 333, objectstore 368, storage 127, messaging 288, servicebus 159, jobs 45, worker 200 → 1639 total. 9 sub-packages (> 5 threshold; no single pkg exceeds 800 LoC, single-session pass acceptable).
**Tracker:** [`../2026-05-21-go-backend-review.md`](../2026-05-21-go-backend-review.md) row #2b.
**Rubric:** [`../../../standards/golang/README.md`](../../../standards/golang/README.md).
**Attribution legend:** `[g]` go, `[s]` security, `[sf]` silent-failure, `[t]` type-design, `[db]` database. Multi-lens overlap = strongest signal.

---

## Severity Counts

| Severity | Count |
|----------|-------|
| Critical | 10 |
| High     | 24 |
| Medium   | 16 |
| Low      | 8 |

---

## Critical

### C1 [g,s,sf,db] `migrate.loadApplied` swallows every query error as "no migrations applied"

**File:** `internal/platform/migrate/migrate.go:74-78`

```go
rows, err := db.QueryContext(ctx, `SELECT version FROM public.schema_migrations`)
if err != nil {
    // Table may not exist on a brand-new DB; treat as empty.
    return out, nil
}
```

Any error — auth denial, connection failure, network timeout, ctx cancel — is silently treated as "no migrations applied yet." `Apply` then attempts to re-run every migration. With duplicate-prefix files (C6, C7) and migrations missing `BEGIN/COMMIT` (C8, C9), this produces partial-apply and silent skip in the same boot cycle.

**Recommend:** Detect only Postgres `42P01` (`undefined_table`) via `pgconn.PgError.Code` or `pgerrcode.UndefinedTable`. Wrap and propagate every other error.

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
    return out, nil
}
return nil, fmt.Errorf("migrate: load applied: %w", err)
```

**Cites:** [`errors-and-logging.md#never-swallow`](../../../standards/golang/errors-and-logging.md#never-swallow), [`errors-and-logging.md#errorsis--errorsas-discipline`](../../../standards/golang/errors-and-logging.md#errorsis--errorsas-discipline).

---

### C2 [g,s] `storage/local` — no containment check after `filepath.Join`, arbitrary host read/write/delete via `..` in `storageKey`

**File:** `internal/platform/storage/local/store.go:20,31,40`

```go
target := filepath.Join(s.rootPath, filepath.FromSlash(storageKey))
```

`filepath.Join` invokes `Clean` which collapses `..` segments — but the result is never re-checked to confirm containment under `rootPath`. A `storageKey` like `../../../etc/cron.d/evil` resolves outside the root and is opened/written/deleted with no error. `storageKey` flows from outbox payloads + DB rows; a poisoned event or compromised tenant write to a key field escalates to arbitrary FS access on the host.

**Recommend:** Containment guard after `Join` on every method (`Get`, `Put`, `Delete`). Also reject null-byte keys.

```go
target := filepath.Join(s.rootPath, filepath.FromSlash(storageKey))
absRoot, _ := filepath.Abs(s.rootPath)
absTarget, _ := filepath.Abs(target)
if !strings.HasPrefix(absTarget+string(filepath.Separator), absRoot+string(filepath.Separator)) {
    return fmt.Errorf("local store: key %q escapes root", storageKey)
}
if strings.ContainsRune(storageKey, 0) {
    return fmt.Errorf("local store: key contains null byte")
}
```

**Cites:** [`security-boundaries.md#fail-closed-authn-userIdfromcontext`](../../../standards/golang/security-boundaries.md#fail-closed-authn-useridfromcontext) (fail-closed principle). **Bar gap — anchor needed:** add `security-boundaries.md#local-store-path-containment` for the explicit storage rule.

---

### C3 [s] SSRF — `docgen_v2` + `gotenberg` `baseURL` accepted without scheme/host validation

**Files:** `internal/platform/config/docgen_v2.go:20-32`, `internal/platform/servicebus/docgen_v2_client.go:19-24`, `internal/platform/servicebus/docgen_v2_pdf.go:42`, `internal/platform/servicebus/docgen_v2_validate.go:17`, `internal/platform/bootstrap/api.go:185` (gotenberg health)

```go
apiURL := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_URL"))
// no scheme/host/allowlist validation
```

`METALDOCS_DOCGEN_V2_URL` and `METALDOCS_GOTENBERG_URL` flow straight into `c.baseURL+"/convert/pdf"` requests. Misconfiguration or env tampering → POSTs to cloud metadata (`http://169.254.169.254/...`), file URIs, internal services, or RFC-1918 hosts. Docgen-v2 also receives the service token, so an attacker-pointed URL exfiltrates it on the first request.

**Recommend:** Validate URL on load. Require `https` scheme + non-empty host. In production builds reject RFC-1918/loopback hosts.

```go
u, err := url.Parse(apiURL)
if err != nil || u.Scheme != "https" || u.Host == "" {
    return DocgenV2Config{}, fmt.Errorf("METALDOCS_DOCGEN_V2_URL must be a valid https URL")
}
```

**Cites:** [`security-boundaries.md#fail-closed-authn-userIdfromcontext`](../../../standards/golang/security-boundaries.md#fail-closed-authn-useridfromcontext). **Bar gap — anchor needed:** `security-boundaries.md#ssrf-url-validation`.

---

### C4 [s] Empty `METALDOCS_DOCGEN_V2_SERVICE_TOKEN` silently accepted — unauthenticated service-to-service calls

**Files:** `internal/platform/config/docgen_v2.go:24`, `internal/platform/servicebus/docgen_v2_pdf.go:47`, `internal/platform/servicebus/docgen_v2_validate.go:22`

```go
token := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN"))
// no non-empty check
```

When `URL` is set but `SERVICE_TOKEN` is unset/empty, every outbound request carries `X-Service-Token: ""`. Fail-open service auth on a regulatory pipeline (controlled documents).

**Recommend:** Require non-empty token whenever the feature is enabled.

```go
if apiURL != "" && token == "" {
    return DocgenV2Config{}, fmt.Errorf("METALDOCS_DOCGEN_V2_SERVICE_TOKEN required when docgen-v2 is enabled")
}
```

**Cites:** [`security-boundaries.md#fail-closed-authn-userIdfromcontext`](../../../standards/golang/security-boundaries.md#fail-closed-authn-useridfromcontext).

---

### C5 [db] Migration runner — no advisory lock; concurrent runners double-apply

**File:** `internal/platform/migrate/migrate.go:24-69`

No `pg_advisory_lock` between `loadApplied` and the file-apply loop. Two API/worker pods starting concurrently both observe the same `applied` snapshot and both run the same migration. Non-idempotent migrations (`0044`, `0063`, `0064`, and any pre-`IF NOT EXISTS` DDL) corrupt schema or data.

**Recommend:** Wrap `Apply` in a session-level advisory lock.

```go
const migrateLockKey int64 = 0x4D44_4D49_4752_8000 // any stable constant
if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrateLockKey); err != nil {
    return fmt.Errorf("migrate: acquire advisory lock: %w", err)
}
defer db.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, migrateLockKey)
```

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### C6 [db] Duplicate migration prefix `0042` — one file permanently skipped

**Files:** `migrations/0042_init_document_departments.sql`, `migrations/0042_add_profile_content_schema.sql`

`migrate.go:18` keys on the 4-digit prefix. Whichever file sorts first inserts version `0042`; the second is silently skipped forever. Different deployments diverge by filename sort order.

**Recommend:** Renumber one (e.g., `0042a`/`0042b` or assign the next unused slot). Add a CI lint asserting prefix uniqueness across `migrations/*.sql`.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries). **Bar gap — anchor needed:** `persistence.md#migration-version-uniqueness`.

---

### C7 [db] Duplicate migration prefix `0130` — one file permanently skipped

**Files:** `migrations/0130_iam_users_tenant_deactivated.sql`, `migrations/0130_documents_drop_old_template_version_fk.sql`

Same root cause as C6. The IAM file adds `tenant_id` + unique indexes; the documents file drops an FK. Neither is idempotent if missed.

**Recommend:** Renumber immediately. Lint enforces prefix uniqueness going forward.

**Cites:** same as C6.

---

### C8 [db] `migrations/0176_pdf_dispatch_outbox.sql` missing `BEGIN/COMMIT` — DDL + ledger insert not atomic

**File:** `migrations/0176_pdf_dispatch_outbox.sql:1-24`

```sql
CREATE TABLE IF NOT EXISTS metaldocs.pdf_dispatch_outbox (...);
CREATE INDEX IF NOT EXISTS ix_pdf_dispatch_outbox_pending ...;
INSERT INTO public.schema_migrations (version, description) VALUES ('0176', ...) ON CONFLICT (version) DO NOTHING;
```

Runner uses `db.ExecContext` with no Go-layer tx (`migrate.go:63`). If the ledger insert fails (e.g., `schema_migrations` permission glitch), DDL is already committed and the runner re-runs the file on every startup.

**Recommend:** Wrap in `BEGIN;` ... `COMMIT;`. Better: enforce at runner level — pre-flight check that file starts with `BEGIN` and ends with `COMMIT`, or open an explicit Go-layer tx per file.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### C9 [db] `migrations/0111_docx_v2_exports.sql` — unqualified FK reference + missing `BEGIN/COMMIT`

**File:** `migrations/0111_docx_v2_exports.sql:1,6`

```sql
CREATE TABLE IF NOT EXISTS document_exports (
    ...
    document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
```

No `public.` / `metaldocs.` schema prefix on `document_exports` or `documents(id)`. Resolution depends on session `search_path` and on whether `documents` survives migration `0113`'s table drop. Also missing `BEGIN/COMMIT`.

**Recommend:** Add explicit schema prefixes on every table reference; wrap in tx; confirm the FK target survives `0113`.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### C10 [t] `messaging.Event` — raw `string` IDs at every module boundary

**File:** `internal/platform/messaging/events.go:7-18`

```go
type Event struct {
    EventID        string
    EventType      string
    AggregateID    string
    IdempotencyKey string
    TraceID        string
    Payload        map[string]any
    ...
}
```

`EventID`, `AggregateID`, `IdempotencyKey` are bare `string` at the public boundary. Any call site can transpose them silently. `EventType` is a bare-string discriminant driving the `switch` in `worker/service.go:49` — no exhaustiveness check; a new event type is a grep-and-pray change. This pattern propagates through `Consumer.MarkPublished([]string)` (H14) and `PDFPersister.WritePDF(tenant, docID, s3Key string)` (H18).

**Recommend:** Newtype the IDs and the discriminant in the `messaging` package; update interface signatures together.

```go
type EventID string
type AggregateID string
type IdempotencyKey string
type EventType string

const EventTypePDFConvert EventType = "docgen_v2_pdf"
```

**Cites:** [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule), [`typed-boundaries.md#anti-patterns`](../../../standards/golang/typed-boundaries.md#anti-patterns).

---

## High

### H1 [g,s,sf] `docgen_v2_validate.go` — `io.ReadAll` unbounded + error discarded; `json.Marshal` error discarded

**File:** `internal/platform/servicebus/docgen_v2_validate.go:16,28`

```go
body, _ := json.Marshal(map[string]string{"docx_key": docxKey, "schema_key": schemaKey})
...
raw, _ := io.ReadAll(resp.Body)
```

Sibling `docgen_v2_pdf.go:55` already uses `io.LimitReader(resp.Body, 64*1024)`. Validation path was missed. Unbounded read on a misbehaving service → OOM. Discarded `ReadAll` err → caller sees `(false, []byte{}, nil)` indistinguishable from clean 200-empty response.

**Recommend:** Mirror the PDF path: `LimitReader` + propagated errors with subsystem prefix.

```go
body, err := json.Marshal(...)
if err != nil { return false, nil, fmt.Errorf("docgen-v2: validate: marshal: %w", err) }
raw, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
if err != nil { return false, nil, fmt.Errorf("docgen-v2: validate: read body: %w", err) }
```

**Cites:** [`errors-and-logging.md#never-swallow`](../../../standards/golang/errors-and-logging.md#never-swallow), [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### H2 [g] `gotenberg.ConvertHTMLToPDF` + `ConvertDocxToPDF` — unbounded success-body read

**File:** `internal/platform/render/gotenberg/client.go:73,112`

Error-body reads are bounded to 4 KiB (lines 69, 108); success PDF body is not. Slow/malicious Gotenberg → OOM in synchronous HTTP request path.

**Recommend:** Cap success body at 100 MB (one trip-wire byte above to detect overflow).

```go
const maxPDFBytes = 100 * 1024 * 1024
pdfBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxPDFBytes+1))
if err != nil { return nil, fmt.Errorf("gotenberg: read pdf: %w", err) }
if int64(len(pdfBytes)) > maxPDFBytes {
    return nil, fmt.Errorf("gotenberg: pdf exceeds %d-byte limit", maxPDFBytes)
}
```

**Cites:** [`errors-and-logging.md#log-or-return`](../../../standards/golang/errors-and-logging.md#log-or-return). **Bar gap — anchor needed:** `persistence.md#http-response-size-limit` or new `errors-and-logging.md#http-response-size-limit`.

---

### H3 [g,s,db] `Consumer.MarkPublished` — `fmt.Sprintf` SQL pattern + unbounded `IN` batch

**File:** `internal/platform/messaging/outbox/postgres/consumer.go:111-132`

```go
for idx, eventID := range eventIDs {
    placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
    args = append(args, strings.TrimSpace(eventID))
}
q := fmt.Sprintf(`UPDATE metaldocs.outbox_events SET ... WHERE event_id IN (%s)`, strings.Join(placeholders, ", "))
```

Values are parameterized (no injection) but the **query string** is built with `fmt.Sprintf` — the pattern the bar prohibits and one indistinguishable from forbidden form. Batch size is also unbounded (current callers cap at 25 but contract does not); approach Postgres's 65535-parameter limit on a future bulk caller.

**Recommend:** Switch to `= ANY($1::uuid[])` with `pq.Array`/`pgtype.Array`. Add explicit batch ceiling defense-in-depth.

```go
const maxBatch = 500
if len(eventIDs) > maxBatch {
    return fmt.Errorf("outbox: mark published: batch %d > %d", len(eventIDs), maxBatch)
}
const q = `UPDATE metaldocs.outbox_events SET published_at = $2, next_attempt_at = NULL, last_error = NULL WHERE event_id = ANY($1)`
```

Side fix: `MarkFailed` error at line 137 has no subsystem prefix — wrap.

**Cites:** [`persistence.md#parameterized-queries-only`](../../../standards/golang/persistence.md#parameterized-queries-only), [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### H4 [g] `worker.RunOnce` — returns DB error naked from inside loop, abandons remaining batch

**File:** `internal/platform/worker/service.go:68-70,104`

```go
if err := s.consumer.MarkPublished(ctx, []string{event.EventID}); err != nil {
    return err
}
...
if err := s.consumer.MarkFailed(ctx, failure); err != nil {
    return false, err
}
```

A transient write to `MarkPublished` aborts the entire `RunOnce` loop. The `ClaimUnpublished` tx already committed the lease; remaining events sit in-flight until lease expiry — full polling cycle blocked behind one flaky DB write.

**Recommend:** Log + `continue` for per-event DB-write failures; do not return up to the supervisor mid-batch.

```go
if err := s.consumer.MarkPublished(ctx, []string{event.EventID}); err != nil {
    slog.ErrorContext(ctx, "worker: mark published failed; will retry next poll",
        "event_id", event.EventID, "error", err)
    continue
}
```

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule), [`idempotency-and-concurrency.md#retry-safe-handler-semantics`](../../../standards/golang/idempotency-and-concurrency.md#retry-safe-handler-semantics).

---

### H5 [g] `worker.backoffDuration` — integer overflow on `1 << (attempt-1)` for large attempt counts

**File:** `internal/platform/worker/service.go:121`

```go
multiplier := 1 << (attempt - 1)
delaySeconds := baseSeconds * multiplier
```

`event.AttemptCount` from DB is not range-checked. Anomalous attempt count → multiplier overflow → negative/zero `delaySeconds` → `nextAttempt` in the past → tight retry loop.

**Recommend:** Clamp the shift operand.

```go
const maxShift = 30
shift := attempt - 1
if shift > maxShift { shift = maxShift }
multiplier := 1 << shift
```

**Cites:** [`idempotency-and-concurrency.md#retry-safe-handler-semantics`](../../../standards/golang/idempotency-and-concurrency.md#retry-safe-handler-semantics). **Bar gap — anchor needed:** `idempotency-and-concurrency.md#backoff-overflow-guard`.

---

### H6 [g] `idempotency.RequestHash` middleware — `io.ReadAll(r.Body)` with no size limit

**File:** `internal/platform/idempotency/middleware.go:26`

Idempotency middleware buffers the entire body to hash it. If upstream `http.MaxBytesReader` is not applied, a multi-GB upload starves memory. The middleware contract does not document the precondition.

**Recommend:** Apply `http.MaxBytesReader` at the chain root for idempotency-guarded routes AND document the contract on `RequestHash`. Alternative: take a `maxBytes int64` parameter.

**Cites:** [`http-handlers.md#request-validation-at-boundary`](../../../standards/golang/http-handlers.md#request-validation-at-boundary). **Bar gap — anchor needed:** `http-handlers.md#request-body-size-limit`.

---

### H7 [g] `bootstrap.bucketEnsurer` — dead interface; `EnsureBucket` never called

**File:** `internal/platform/bootstrap/api.go:56-57`, `internal/platform/storage/minio/store.go` (EnsureBucket impl)

Interface declared, never invoked. MinIO bucket existence is assumed at runtime; first attachment surfaces an opaque MinIO error instead of failing fast at startup.

**Recommend:** Call `miniostore.EnsureBucket(ctx)` during `BuildAPIDependencies` MinIO branch; remove the dead interface or wire it to the call.

```go
if err := store.EnsureBucket(ctx); err != nil {
    _ = closeDB(db)
    return APIDependencies{}, fmt.Errorf("bootstrap: ensure minio bucket: %w", err)
}
```

**Cites:** [`package-layout.md#constructor-invariant-pattern`](../../../standards/golang/package-layout.md#constructor-invariant-pattern).

---

### H8 [s,sf] `worker/service.go` — `log.Printf` throughout (5 sites); slog convention bypass + uncontrolled error string leakage

**File:** `internal/platform/worker/service.go:54,72,76,95,100`

```go
log.Printf("worker_event event_id=%s ... error=%q", event.EventID, ..., failure.LastError)
```

Stdlib `log` package, unstructured, no `ctx` propagation, no field-level redaction surface. `error=%q` dumps post-truncate error strings (512B) including remote service bodies (see H11).

**Recommend:** Replace all 5 sites with `slog.InfoContext`/`slog.ErrorContext` using structured fields. Pass `ctx` into `markFailure`.

**Cites:** [`errors-and-logging.md#slog-conventions`](../../../standards/golang/errors-and-logging.md#slog-conventions).

---

### H9 [s,sf] `objectstore/document_presigner.go:80` — `log.Printf` on cleanup failure

**File:** `internal/platform/objectstore/document_presigner.go:80`

```go
log.Printf("objectstore: adopt tmp cleanup failed for key=%s: %v", tmpKey, err)
```

Same as H8. `tmpKey` carries tenant + document path; structured slog field is required so log filtering can scrub.

**Recommend:** `slog.WarnContext(ctx, "objectstore: adopt tmp cleanup failed", "key", tmpKey, "error", err)`. Drop the `log` import.

**Cites:** [`errors-and-logging.md#slog-conventions`](../../../standards/golang/errors-and-logging.md#slog-conventions).

---

### H10 [s] Presigners accept caller-supplied `storageKey` with no tenant-prefix enforcement

**Files:** `internal/platform/objectstore/presign.go:61-67`, `internal/platform/objectstore/templates_presigner.go:39-45`, `internal/platform/objectstore/document_presigner.go:53-62`

```go
func (p *TemplatePresigner) PresignObjectGET(ctx context.Context, storageKey string) (string, error) {
    u, err := p.signingClient.PresignedGetObject(ctx, p.bucket, storageKey, p.ttl, nil)
```

Key prefix tenant scope depends entirely on upstream callers. If a tenant's DB row stores `tenants/OTHER_TENANT/...` as `StorageKey` (write path elsewhere), the presigner signs a valid URL to the foreign object.

**Recommend:** Add a `tenantID` parameter to every presign method and validate `strings.HasPrefix(storageKey, "tenants/"+tenantID+"/")` before signing. Breaking-but-essential interface change.

**Cites:** [`security-boundaries.md#fail-closed-authn-userIdfromcontext`](../../../standards/golang/security-boundaries.md#fail-closed-authn-useridfromcontext). **Bar gap — anchor needed:** `security-boundaries.md#object-key-tenant-scoping`.

---

### H11 [s] `docgen_v2_pdf.go:61` — remote service error body echoed verbatim into error message

**File:** `internal/platform/servicebus/docgen_v2_pdf.go:61`

```go
return zero, fmt.Errorf("docgen-v2 convert pdf: unexpected status %d body=%s", resp.StatusCode, string(respBody))
```

`respBody` flows through worker → `truncateError(512)` → `metaldocs.outbox_events.last_error` row in DB. Anything docgen-v2 leaks (internal paths, keys, stack traces) is persisted.

**Recommend:** Log raw body at debug; return sanitized error.

```go
slog.WarnContext(ctx, "docgen-v2: non-200", "status", resp.StatusCode, "body", string(respBody))
return zero, fmt.Errorf("docgen-v2 convert pdf: unexpected status %d", resp.StatusCode)
```

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### H12 [s] `worker/pdf_job_runner.go` — payload `tenant_id`/`revision_id` used in object key construction without UUID validation

**File:** `internal/platform/worker/pdf_job_runner.go:34-44`

```go
tenantID, _ := event.Payload["tenant_id"].(string)
revisionID, _ := event.Payload["revision_id"].(string)
docxKey = fmt.Sprintf("tenants/%s/revisions/%s/frozen.docx", tenantID, revisionID)
```

Payload is `jsonb` with no ingest schema check. Poison payload with `tenant_id="../admin"` → traversal in storage key. Mitigated upstream by trusted producers but defense-in-depth required.

**Recommend:** UUID-validate both fields before key construction.

```go
if _, err := uuid.Parse(tenantID); err != nil {
    return fmt.Errorf("pdf job: invalid tenant_id in payload")
}
```

**Cites:** [`http-handlers.md#request-validation-at-boundary`](../../../standards/golang/http-handlers.md#request-validation-at-boundary), [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule).

---

### H13 [s] `bootstrap/api.go:117-141` — `default:` case of `repoMode` switch fail-opens to memory/dev mode with seeded dev credentials

**File:** `internal/platform/bootstrap/api.go:117-141`

```go
switch repoMode {
case ...:
default:
    roles := authn.DevRoleMap()
    authRepo := authmemory.NewRepository()
    ...
}
```

Unknown/typo'd `METALDOCS_REPOSITORY_MODE` silently activates memory mode + pre-seeded dev users. No `APP_ENV` guard.

**Recommend:** Replace `default:` with allowlist `case config.RepositoryMemory:` + final `default: return ..., fmt.Errorf("unsupported repo mode: %q", repoMode)`. Refuse memory mode when `APP_ENV in {production, staging}`.

**Cites:** [`security-boundaries.md#fail-closed-authn-userIdfromcontext`](../../../standards/golang/security-boundaries.md#fail-closed-authn-useridfromcontext).

---

### H14 [t] `messaging.Consumer.MarkPublished([]string)` — raw IDs at interface (resolved together with C10)

**File:** `internal/platform/messaging/consumer.go:17`

Bare `[]string` event-ID slice. After C10's `EventID` newtype, this becomes `[]EventID`.

**Recommend:** Land in C10's PR.

**Cites:** [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule).

---

### H15 [t] `messaging.Event.Payload map[string]any` erases per-event-type invariants

**File:** `internal/platform/messaging/events.go:17`

`worker/pdf_job_runner.go:34-43` immediately type-asserts raw strings back out. Missing-field errors are unnamed. Schema lives only in producer/consumer agreement.

**Recommend:** Typed payload structs per event type; unmarshal at dispatch.

```go
type PDFConvertPayload struct {
    TenantID       string `json:"tenant_id"`
    RevisionID     string `json:"revision_id"`
    FinalDocxS3Key string `json:"final_docx_s3_key,omitempty"`
}
```

**Cites:** [`typed-boundaries.md#anti-patterns`](../../../standards/golang/typed-boundaries.md#anti-patterns).

---

### H16 [t] `objectstore.PresignContext` — exported mutable fields after constructor validation

**File:** `internal/platform/objectstore/presign.go:16-29`

`NewPresignContext` validates `MaxSizeBytes > 0` and `TTL > 0` then returns a `*PresignContext` whose fields are publicly mutable. Pointer-shared mutation reopens the invariant.

**Recommend:** Unexport fields, expose accessors, return value not pointer.

```go
type PresignContext struct { maxSizeBytes int64; ttl time.Duration }
func (p PresignContext) MaxSizeBytes() int64 { return p.maxSizeBytes }
func (p PresignContext) TTL() time.Duration  { return p.ttl }
```

Also: appears unused in production paths (only `presign_test.go`). Verify and delete if dead.

**Cites:** [`package-layout.md#constructor-invariant-pattern`](../../../standards/golang/package-layout.md#constructor-invariant-pattern).

---

### H17 [t] `TemplateDocxKey` / `TemplateSchemaKey` / `PresignRevisionPUT` — bare positional `string` tenant/template IDs

**Files:** `internal/platform/objectstore/template_keys.go:5-11`, `internal/platform/objectstore/presign.go:43,52`, `internal/platform/objectstore/document_presigner.go:41`

```go
func TemplateDocxKey(tenantID, templateID string, versionNum int) string
```

Two adjacent `string` params with distinct roles → silent positional swap produces valid-looking wrong key.

**Recommend:** Introduce domain newtypes (`tenant.ID`, `template.ID`) and `type ObjectKey string` return. Intermediate: at least make the **return** a distinct type so callers cannot re-feed it as a tenant ID.

**Cites:** [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule).

---

### H18 [t,sf] `DocumentPresigner` — nil-client checks repeated in every method body; `HeadObject`/`SizeObject` silently return `(false,nil)`/`(0,nil)` on nil client

**Files:** `internal/platform/objectstore/document_presigner.go:41-147`, `internal/platform/objectstore/document_presigner_export.go:10-12,25-27`

`HeadObject` returns "object does not exist" on misconfiguration → callers re-upload silently. `SizeObject` returns zero → export records claim 0 bytes.

**Recommend:** Reject nil client at construction; remove all per-method nil checks; align both export methods with the rest of the type (return error).

```go
func NewDocumentPresigner(client, signingClient *minio.Client, ...) (*DocumentPresigner, error) {
    if client == nil { return nil, errors.New("objectstore: minio client required") }
    ...
}
```

**Cites:** [`package-layout.md#constructor-invariant-pattern`](../../../standards/golang/package-layout.md#constructor-invariant-pattern), [`errors-and-logging.md#never-swallow`](../../../standards/golang/errors-and-logging.md#never-swallow).

---

### H19 [t] `worker.PDFPersister.WritePDF(tenant, docID, s3Key string, ...)` — three consecutive bare string IDs

**File:** `internal/platform/worker/pdf_job_runner.go:17-18`

Three same-typed positional params with distinct semantic roles.

**Recommend:** Typed IDs once domain owners exist; until then a local `type StorageKey string` separates at least the third position.

**Cites:** [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule).

---

### H20 [db] `idx_outbox_claimable` partial index missing `next_attempt_at` predicate — retry-loaded scans degrade

**File:** `migrations/0020_extend_outbox_retry_and_dlq.sql:10-13`, query at `internal/platform/messaging/outbox/postgres/consumer.go:43`

```sql
CREATE INDEX idx_outbox_claimable
ON metaldocs.outbox_events (occurred_at ASC)
WHERE published_at IS NULL AND dead_lettered_at IS NULL;
```

Query filters `next_attempt_at IS NULL OR next_attempt_at <= NOW()`. Under retry pressure the predicate evaluates as heap filter — near-full pending-set scan.

**Recommend:** Add `next_attempt_at` as a leading or composite column. Safest portable form:

```sql
CREATE INDEX idx_outbox_claimable
ON metaldocs.outbox_events (next_attempt_at NULLS FIRST, occurred_at ASC)
WHERE published_at IS NULL AND dead_lettered_at IS NULL;
```

**Cites:** [`persistence.md#parameterized-queries-only`](../../../standards/golang/persistence.md#parameterized-queries-only) (index hygiene adjacent). **Bar gap — anchor needed:** `persistence.md#index-covering-query-predicates`.

---

### H21 [db] Dual pool — API and Jobs each open `MaxOpenConns(25)`; co-resident binary opens 50

**Files:** `internal/platform/bootstrap/api.go:73`, `internal/platform/bootstrap/jobs.go:31`

Two independent `pgdb.Open` calls; no shared `*sql.DB`. In dev (single binary) → 50 connections immediately.

**Recommend:** Share a single `*sql.DB` when API + Jobs co-reside, or externalize `MaxOpenConns` and lower it.

**Cites:** [`persistence.md#connection-pool-hygiene`](../../../standards/golang/persistence.md#connection-pool-hygiene).

---

### H22 [db] `migrations/0042_init_document_departments.sql` — no `BEGIN/COMMIT`, no ledger insert

**File:** `migrations/0042_init_document_departments.sql`

`CREATE TABLE` + `CREATE INDEX` without tx wrap; if index creation fails, table is committed standalone. Compounded by C6.

**Recommend:** Add `BEGIN;` / `COMMIT;` and ledger insert. Same fix family as C8/C9.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### H23 [db] `idempotency.BeginReplay` — unbounded recursion (potential stack overflow on race / failed-row loop)

**File:** `internal/platform/idempotency/postgres_store.go:114-185`

Three recursive self-calls (lines 116, 162, 185) handle "row vanished after commit", "expired in-flight reclaim", "failed-row retry". Persistent state in any branch → unbounded depth.

**Recommend:** Convert to iterative loop with explicit retry counter (e.g., 3).

```go
const maxReplayRetries = 3
for attempt := 0; attempt < maxReplayRetries; attempt++ {
    // ... existing branches use continue instead of recursion
}
return nil, nil, fmt.Errorf("idempotency: begin replay: exceeded %d retries", maxReplayRetries)
```

**Cites:** [`idempotency-and-concurrency.md#retry-safe-handler-semantics`](../../../standards/golang/idempotency-and-concurrency.md#retry-safe-handler-semantics).

---

### H24 [sf] `idempotency.CompleteReplay` — `res.RowsAffected()` error silently discarded

**File:** `internal/platform/idempotency/postgres_store.go:231`

```go
n, _ := res.RowsAffected()
```

Driver-level `RowsAffected` error masked as `n=0`, triggering `n != 1` branch with misleading error message.

**Recommend:** Capture both errors per [`persistence.md#rowsaffected-discipline`](../../../standards/golang/persistence.md#rowsaffected-discipline).

```go
n, raErr := res.RowsAffected()
if raErr != nil { ... return fmt.Errorf("idempotency: complete: rows affected: %w", raErr) }
```

**Cites:** [`persistence.md#rowsaffected-discipline`](../../../standards/golang/persistence.md#rowsaffected-discipline).

---

## Medium

### M1 [s,db] `config/postgres.go:30` — default `PGSSLMODE=disable` ships unencrypted DB in any deployment that omits the env var

**Recommend:** Default to `require`. Explicit override for local dev.

**Cites:** [`persistence.md#connection-pool-hygiene`](../../../standards/golang/persistence.md#connection-pool-hygiene). **Bar gap:** `persistence.md#tls-required`.

---

### M2 [g] `docgenv2.TemplateReader.GetPublishedVersion` returns naked `sql.ErrNoRows`

**File:** `internal/platform/docgenv2/template_reader.go:37`

**Recommend:** Wrap with tenant + version context; preserve `errors.Is` discrimination.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M3 [g] `docgenv2.TemplatesSnapshotReader` returns naked non-`ErrNoRows` error

**File:** `internal/platform/docgenv2/templates_snapshot_reader.go:42-43`

**Recommend:** Wrap with subsystem prefix.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M4 [g] `migrate.go:82-84` — `rows.Scan` error returned naked

**Recommend:** `fmt.Errorf("migrate: scan schema_migrations version: %w", err)`.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M5 [g] `observability/runtime.go:169-190` — readiness probe error strings have no subsystem prefix

**Recommend:** Prefix per subsystem before stuffing into the response map.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M6 [g] `servicebus/docgen_v2_client.go:37-41`, `docgen_v2_pdf.go:43,49` — bare `NewRequest` / `Do` errors

**Recommend:** Wrap with `"docgen-v2: <op>: ..."`.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M7 [g] `objectstore/presign.go:45-66` — three presign methods return naked MinIO errors

**Recommend:** Wrap with op + key context.

**Cites:** [`errors-and-logging.md#error-wrapping-rule`](../../../standards/golang/errors-and-logging.md#error-wrapping-rule).

---

### M8 [t] `servicebus.NewDocgenV2Client(baseURL, token string, ...)` — transposable adjacent `string` constructor args

**Recommend:** `type ServiceToken string` newtype.

**Cites:** [`typed-boundaries.md#the-rule`](../../../standards/golang/typed-boundaries.md#the-rule).

---

### M9 [t] `servicebus.PDFRenderOpts.PaperSize string` — unenumerated discriminant

**File:** `internal/platform/servicebus/docgen_v2_pdf.go:19-23`

**Recommend:** Named type + constants (`PaperSizeA4`, `PaperSizeLetter`).

**Cites:** [`typed-boundaries.md#anti-patterns`](../../../standards/golang/typed-boundaries.md#anti-patterns).

---

### M10 [t] `jobs/river.ClientBundle.Driver` — public concrete `*riverdatabasesql.Driver` field leaks infrastructure type to callers

**File:** `internal/platform/jobs/river/client.go:17-19`

**Recommend:** Unexport `driver`; expose only what consumers need.

**Cites:** [`package-layout.md#platform-packages`](../../../standards/golang/package-layout.md#platform-packages).

---

### M11 [t] `bootstrap.APIDependencies` — MinIO trio (`MinioClient`, `MinioPublicClient`, `MinioBucket`) as disconnected nilable/empty fields permits invalid combinations

**File:** `internal/platform/bootstrap/api.go:33-54`

**Recommend:** Group into a `MinIODeps` sub-struct with its own constructor invariant.

**Cites:** [`package-layout.md#constructor-invariant-pattern`](../../../standards/golang/package-layout.md#constructor-invariant-pattern).

---

### M12 [t] `config.AttachmentsConfig.Provider string` — bare-string storage discriminant

**File:** `internal/platform/config/attachments.go:16-28`

**Recommend:** `type StorageProvider string` with constants; exhaustive lint.

**Cites:** [`typed-boundaries.md#anti-patterns`](../../../standards/golang/typed-boundaries.md#anti-patterns).

---

### M13 [db] `migrations/0007_init_outbox_events.sql` — `event_id`/`idempotency_key` are `TEXT` PK/unique with no `CHECK char_length(...)` bound

**Recommend:** `CHECK (char_length(event_id) <= 256)`, `CHECK (char_length(idempotency_key) <= 512)`.

**Cites:** [`persistence.md#parameterized-queries-only`](../../../standards/golang/persistence.md#parameterized-queries-only) (column hygiene adjacent).

---

### M14 [db] `migrations/0148_job_leases.sql:120` — `heartbeat_lease` hardcodes 5-minute extension; ignores caller TTL

**Recommend:** Accept `_ttl interval` param; use it in the `UPDATE`.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### M15 [db] `migrations/0176` `ix_pdf_dispatch_outbox_pending` — partial index with `status IN (...)` predicate not reliably usable by planner

**Recommend:** Two partial indexes (one per status value) OR composite `(status, next_retry_at)` without partial predicate.

**Cites:** [`persistence.md#parameterized-queries-only`](../../../standards/golang/persistence.md#parameterized-queries-only) (index hygiene).

---

### M16 [db] `migrations/0148_job_leases.sql:16-19` — creates `idempotency_keys` index inside the job-leases migration; ordering hazard if `0147` not yet applied

**Recommend:** Move into `0147_idempotency_keys.sql` or dedicated `0147b`.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

## Low

### L1 [g] `migrate.go:63` — relies on each migration file containing its own `BEGIN/COMMIT`; not runner-enforced

**Recommend:** Pre-flight lint enforcing `BEGIN` prefix and `COMMIT` suffix on every file.

**Cites:** [`persistence.md#transaction-boundaries`](../../../standards/golang/persistence.md#transaction-boundaries).

---

### L2 [g] `bootstrap/api.go:94` — silently discards `closeDB(db)` cleanup error without comment

**Recommend:** Add `// best-effort cleanup; original error returned below` comment.

**Cites:** [`errors-and-logging.md#never-swallow`](../../../standards/golang/errors-and-logging.md#never-swallow).

---

### L3 [g] `objectstore/document_presigner.go:157` — `isNoSuchKeyErr` falls back to `strings.Contains(err.Error(), "NoSuchKey")`

**Recommend:** Drop string fallback; rely on `errors.As(&minio.ErrorResponse{})` + `*url.Error`.

**Cites:** [`errors-and-logging.md#errorsis--errorsas-discipline`](../../../standards/golang/errors-and-logging.md#errorsis--errorsas-discipline).

---

### L4 [g] `config/worker.go:55-59` — `MaxAttempts` lower-bounded only; no upper bound enables H5 overflow path

**Recommend:** Reject `MaxAttempts > 100`.

**Cites:** [`package-layout.md#constructor-invariant-pattern`](../../../standards/golang/package-layout.md#constructor-invariant-pattern).

---

### L5 [s,db] `db/postgres/connect.go:17-21` — `MaxOpenConns(25)`, `MaxIdleConns(25)` hardcoded; no override path

**Recommend:** Externalize through `config.PostgresConfig`.

**Cites:** [`persistence.md#connection-pool-hygiene`](../../../standards/golang/persistence.md#connection-pool-hygiene).

---

### L6 [s] `worker/service.go:32-38` — `batchSize` accepts any positive int with no upper bound inside `RunOnce`

**Recommend:** Defense-in-depth ceiling (e.g., 1000).

**Cites:** [`http-handlers.md#request-validation-at-boundary`](../../../standards/golang/http-handlers.md#request-validation-at-boundary).

---

### L7 [t] `bootstrap.APIDependencies.Cleanup func()` — no error surface; flush/drain failures swallowed silently

**File:** `internal/platform/bootstrap/api.go:53`

**Recommend:** `type CleanupFn func(ctx context.Context) error`.

**Cites:** [`errors-and-logging.md#never-swallow`](../../../standards/golang/errors-and-logging.md#never-swallow).

---

### L8 [db] `storage/minio/store.go:22` — second MinIO client construction path without `Region`; differs from `bootstrap/api.go:148` which hardcodes `us-east-1`

**Recommend:** Extract shared `newMinioClient(cfg)` helper.

**Cites:** [`package-layout.md#platform-packages`](../../../standards/golang/package-layout.md#platform-packages).

---

## Bar Gaps (new anchors required)

These Critical/High findings cite an existing principle but lack a precise anchor. Each gap should be backfilled in the same PR that lands the corresponding fix (per plan §0 goal 2).

| Gap | Doc | New anchor proposed | Driven by |
|-----|-----|---------------------|-----------|
| Local file-store path containment | `security-boundaries.md` | `#local-store-path-containment` | C2 |
| Outbound URL SSRF validation | `security-boundaries.md` | `#ssrf-url-validation` | C3 |
| Object key tenant-scoping at presign boundary | `security-boundaries.md` | `#object-key-tenant-scoping` | H10 |
| HTTP response body size limit | `errors-and-logging.md` or `persistence.md` | `#http-response-size-limit` | H2 |
| Backoff shift overflow guard | `idempotency-and-concurrency.md` | `#backoff-overflow-guard` | H5 |
| Request body size limit on idempotency middleware | `http-handlers.md` | `#request-body-size-limit` | H6 |
| Migration version uniqueness | `persistence.md` | `#migration-version-uniqueness` | C6, C7 |
| Index covering query predicates | `persistence.md` | `#index-covering-query-predicates` | H20 |
| Default TLS required on DB | `persistence.md` | `#tls-required` | M1 |

---

## Cross-Lens Overlap Notes

Multi-lens hits = highest signal. Top overlaps:

- **C1** (migrate.loadApplied swallow) — 4 lenses [g,s,sf,db].
- **H1** (docgen_v2_validate unbounded read + json.Marshal _) — 3 lenses [g,s,sf].
- **H3** (MarkPublished fmt.Sprintf + unbounded IN) — 3 lenses [g,s,db].
- **H8** (worker log.Printf) — 2 lenses [s,sf].
- **H9** (cleanup log.Printf) — 2 lenses [s,sf].
- **H18** (DocumentPresigner nil checks + silent HeadObject/SizeObject) — 2 lenses [t,sf].

---

## Module Wiki Drift (G0 note)

No `wiki/modules/platform-2b.md` or equivalent module wiki exists for the #2b sub-group. Module wiki coverage gap — not blocking review per plan §3 G0. Dispatch `wiki-curator` after #2b fixes land to author the trio.

---

## Critical Handoff (G3)

Plan §6 G3 requires: either land fix in same session OR create backlog row with owner + ETA. **Decision pending in cursor update** — see `memory/project_go_backend_review.md`.

Recommended grouping for Critical fix branches (per plan §10 cadence):

| Fix branch | Criticals bundled | Estimated effort |
|-----------|-------------------|------------------|
| `fix/migrate-2b-c1-c5-c8-c9` | C1, C5, C6, C7, C8, C9 (migration runner + duplicates + tx wrap) | M |
| `fix/storage-2b-c2` | C2 (path containment) | S |
| `fix/docgen-2b-c3-c4` | C3, C4 (SSRF + token) | S |
| `fix/messaging-2b-c10` | C10 (typed Event boundary — cascades to H14, H15, H19) | M-L |

C6 and C7 (duplicate migration prefixes) are silent-data-bomb fixes — recommend landing first, as a standalone surgical change, before any other migration touches the tree.
