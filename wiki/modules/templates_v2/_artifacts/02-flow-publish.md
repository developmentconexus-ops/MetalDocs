# Operation Trace: publishTemplateVersionV2

Operation `publishTemplateVersionV2` (HTTP `POST /api/v2/templates/{id}/versions/{n}/publish`) in module `internal/modules/templates_v2`.

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `publishTemplateVersionV2` | `api/openapi/v1/openapi.yaml:2912` |
| Generated server stub | `ServerInterface.PublishTemplateVersionV2` | `internal/modules/templates_v2/api/api.gen.go:582` |
| Handler | `Handler.PublishTemplateVersionV2` | `internal/modules/templates_v2/delivery/http/routes_generated.go:185` |

## 2. Call chain

1. `internal/modules/templates_v2/delivery/http/handler.go:46` route registration binds publish path.
   -> calls: `internal/modules/templates_v2/api/api.gen.go:764` `ServerInterfaceWrapper.PublishTemplateVersionV2`
2. `internal/modules/templates_v2/api/api.gen.go:764` wrapper parses path params.
   -> calls: `internal/modules/templates_v2/api/api.gen.go:1419` `strictHandler.PublishTemplateVersionV2`
3. `internal/modules/templates_v2/api/api.gen.go:1419` strict middleware decodes request body.
   -> calls: `internal/modules/templates_v2/api/api.gen.go:1433` `sh.ssi.PublishTemplateVersionV2(...)`
4. `internal/modules/templates_v2/delivery/http/routes_generated.go:185` handler edge logic.
   -> authz call: `h.authz(r, tenantID, "*", "template.approve")` (`routes_generated.go:188`)
   -> calls: `internal/modules/templates_v2/application/lifecycle.go:265` `Service.PublishTemplateVersion`
5. `internal/modules/templates_v2/application/lifecycle.go:265` publish flow.
   -> `repo.GetTemplate` (`:266`)
   -> `repo.GetVersion` (`:270`)
   -> guard `version.Status == draft` (`:274-276`)
   -> `repo.ObsoletePreviousPublished` (`:283`)
   -> `repo.UpdateTemplate` (`:287`)
   -> `repo.UpdateVersion` (`:290`)
   -> `repo.AppendAudit(Action=AuditPublished)` (`:293-300`)
   -> `Service.CreateNextVersion` (`:305`, defined `application/create.go:115`)
6. `internal/modules/templates_v2/application/create.go:115` next draft creation.
   -> `repo.CreateVersion` (`create.go:148`)
   -> `repo.UpdateTemplate` (`create.go:153`)
7. `internal/modules/templates_v2/repository/postgres.go:269` `ObsoletePreviousPublished` SQL UPDATE.
8. `internal/modules/templates_v2/repository/postgres.go:229` `UpdateVersion` SQL UPDATE.
9. `internal/modules/templates_v2/repository/postgres.go:311` `AppendAudit` SQL INSERT.

**Transaction boundary:** No `BeginTx` / `sql.Tx` / `pgx.Tx` in templates_v2 handler/service/repo publish path. Repo methods execute `r.db.ExecContext(...)` independently (`postgres.go:253`, `:274`, `:324`).

**Idempotency:** No idempotency store/key on publish path. Replay against same version fails state guard (`lifecycle.go:274-276`) -> mapped to `409 invalid_state_transition` (`delivery/http/errors.go:20-21`).

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `TemplateVersion` | `draft` | `published` | `PublishTemplateVersion` | edge check `template.approve` (`routes_generated.go:188`) |
| `TemplateVersion` (previous published rows) | `published` | `obsolete` | `ObsoletePreviousPublished` | none in repo method |
| `Template` | previous/null `published_version_id` | current version id | `UpdateTemplate` in publish flow | edge check `template.approve` |
| `TemplateVersion` (new row) | n/a | `draft` | `CreateNextVersion` side-effect | edge check `template.approve` |

Precursor lifecycle ops (1-line each):
- Submit: `SubmitForReviewCmd` (`application/lifecycle.go:9`) `draft -> in_review`.
- Review: `ReviewCmd` (`application/lifecycle.go:66`) `in_review -> approved` (accept) or `in_review -> draft` (reject).
- Approve: `ApproveCmd` (`application/lifecycle.go:150`) formal sign-off path, including publish on accept.
- Publish (this trace): `PublishTemplateVersionCmd` (`application/lifecycle.go:253`) requires `draft`, sets `published`.

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/templates_v2/repository/postgres.go:269` | UPDATE | `templates_v2_template_version` | none |
| `internal/modules/templates_v2/repository/postgres.go:229` | UPDATE | `templates_v2_template_version` | none |
| `internal/modules/templates_v2/repository/postgres.go:151` | UPDATE | `templates_v2_template` | none |
| `internal/modules/templates_v2/repository/postgres.go:311` | INSERT | `templates_v2_audit_log` | none |
| `internal/modules/templates_v2/repository/postgres.go:118` | INSERT | `templates_v2_template_version` | none |

Tripwire pairing (`authz.Require(...)` before mutating SQL on same tx):
- `authz.Require` in templates_v2 path: not found.
- Mutating SQL present.
- Pairing status: `VIOLATION`.

## 5. Response shape

- 2xx schema ref: inline object schema at `api/openapi/v1/openapi.yaml:2932` (no `#/components/schemas/...` ref).
- Declared op error responses in OpenAPI: none listed for this op; `responses` shows only `200` (`openapi.yaml:2926-2932`).
- Handler error envelope is legacy `{"error":{"code","message"}}` via `writeErr`/`writeMappedErr` (`delivery/http/handler.go:95-101`, `:108-113`), not RFC 9457.
- Problem `type` URI: none in handler path.

## 6. Explicit findings

**Authz:**
- `authz == nil` defaults to allow-all function in handler constructor (`delivery/http/handler.go:25-27`).
- Publish route checks `template.approve` (`routes_generated.go:188`), not `template.publish`.
- `template.publish` exists in seed migration `0165` for `system_admin` (`migrations/0165_role_capabilities_reseed.sql:38`).
- **GAP:** `template.publish` capability is defined in seed data but the publish route enforces `template.approve` instead. If `authz` is nil (allow-all), enforcement is fully bypassed regardless.

**Transactional boundary:**
- Publish + ObsoletePreviousPublished run as separate `ExecContext` calls with no wrapping transaction.
- **Race window:** two concurrent publish calls for the same template can each pass the draft guard before either commits, resulting in two rows with `status='published'` simultaneously.

**Audit:**
- `AuditPublished` appended at `lifecycle.go:298`.
- `AuditObsoleted` constant exists (`domain/audit.go:15`) but is **not** appended for the ObsoletePreviousPublished side-effect.

**Idempotency:**
- Replay does not no-op; it fails the draft guard and returns a 409 conflict mapping.

**ISO segregation (SoD):**
- `ApproveCmd` calls `domain.CheckSegregation("approver", ...)` (`lifecycle.go:181`).
- `PublishTemplateVersion` has **no** segregation check; the author can publish their own version.

**Error envelope:**
- Legacy `{"error":{"code","message"}}` envelope. Not RFC 9457.

---
Last verified: 2026-05-10
