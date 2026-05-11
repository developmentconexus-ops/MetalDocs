## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | n/a - internal port | n/a |
| Generated server stub | n/a - internal port | n/a |
| Handler | `AdminHandler.handleUnlockUser` (trigger context only; not audit module entry) | `internal/modules/iam/delivery/http/admin_handler.go:423` |

## 2. Call chain

1. `internal/modules/iam/delivery/http/admin_handler.go:423` `(*AdminHandler).handleUnlockUser` - IAM unlock handler triggers audit after successful unlock.
   -> calls: `internal/modules/iam/delivery/http/admin_handler.go:432` `(*AdminHandler).recordAudit`
2. `internal/modules/iam/delivery/http/admin_handler.go:449` `(*AdminHandler).recordAudit` - marshals payload and emits audit event through audit writer port.
   -> calls: `internal/modules/iam/delivery/http/admin_handler.go:457` `h.audit.Record` (`auditdomain.Writer`)
3. `internal/modules/audit/domain/port.go:25` `Writer.Record(ctx, event) error` - audit module write port contract.
   -> calls: `internal/modules/audit/infrastructure/postgres/writer.go:20` `(*postgres.Writer).Record`
4. `internal/modules/audit/infrastructure/postgres/writer.go:20` `(*Writer).Record` - sink implementation inserts event row.
   -> calls: `internal/modules/audit/infrastructure/postgres/writer.go:27` `w.db.ExecContext`
5. `internal/modules/audit/infrastructure/postgres/writer.go:27` `(*sql.DB).ExecContext` - executes `INSERT INTO metaldocs.audit_events`.

Transaction boundary: no `BeginTx`/`Commit`/`Rollback` appears in the cited unlock/audit path; calls pass `r.Context()` directly into `h.audit.Record` and then `w.db.ExecContext` (`internal/modules/iam/delivery/http/admin_handler.go:457`, `internal/modules/audit/infrastructure/postgres/writer.go:27`).

Authz calls: none in cited files for this write path (`internal/modules/iam/delivery/http/admin_handler.go:423-466`, `internal/modules/audit/infrastructure/postgres/writer.go:20-42`).

Idempotency interactions: none in cited files for this write path (`internal/modules/iam/delivery/http/admin_handler.go:423-466`, `internal/modules/audit/infrastructure/postgres/writer.go:20-42`).

## 3. State changes

none - append-only sink (`internal/modules/audit/infrastructure/postgres/writer.go:22-25`).

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/audit/infrastructure/postgres/writer.go:22` | INSERT | `metaldocs.audit_events` | none |

Tripwire pairing (`authz.Require` before mutating SQL on same tx): N/A - no authz call and no explicit transaction boundary in this path (`internal/modules/iam/delivery/http/admin_handler.go:423-466`, `internal/modules/audit/infrastructure/postgres/writer.go:20-42`).

## 5. Response shape

n/a - internal port, returns error only (`internal/modules/audit/domain/port.go:26`).

## 6. Cross-references

- Idempotency: no (`internal/modules/iam/delivery/http/admin_handler.go:423-466`, `internal/modules/audit/infrastructure/postgres/writer.go:20-42`)
- Pagination: no (write path only; `Writer.Record` is single-event write) (`internal/modules/audit/domain/port.go:25-27`)
- Audit log emission: self (this is the sink) (`internal/modules/audit/infrastructure/postgres/writer.go:20-27`)

### (a) Transaction context

`Record` is called in the request flow using `r.Context()` with no surrounding explicit transaction in the shown caller path.

Quoted lines:

- `internal/modules/iam/delivery/http/admin_handler.go:428`
  `if err := h.authService.UnlockUser(r.Context(), userID); err != nil {`
- `internal/modules/iam/delivery/http/admin_handler.go:432`
  `h.recordAudit(r, userID, "auth.user.unlocked", map[string]any{})`
- `internal/modules/iam/delivery/http/admin_handler.go:457`
  `_ = h.audit.Record(r.Context(), auditdomain.Event{`

### (b) Postgres writer tx usage

Yes. `writer.go` uses `w.db.ExecContext` directly, with no transaction object in `Record`.

Citation: `internal/modules/audit/infrastructure/postgres/writer.go:27`.

### (c) Error handling in caller

The error return from `h.audit.Record` is ignored.

Quoted line:

- `internal/modules/iam/delivery/http/admin_handler.go:457`
  `_ = h.audit.Record(r.Context(), auditdomain.Event{`
