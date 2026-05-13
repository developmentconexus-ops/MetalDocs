# Phase 0 — Context · templates

> Date: 2026-05-10
> Composer: main agent (Opus 4.7)

## Module identity

- **Module name on disk:** `internal/modules/templates/`
- **Wiki doc filename target:** `wiki/modules/templates.md` (matches on-disk; rename to `templates.md` happens in same commit as `templates/ → templates/` code rename — NOT this commit)
- **Status:** active production module; Plan 2 (commits ae1229e8..c84215f7) flipped `/api/v1/` → `/api/v1/` and started API design system spec (RFC 9457 envelope, cursor pagination). Code rename `templates/ → templates/` pending.
- **Predecessor stub:** `wiki/modules/templates.md` (kebab dash) — heavy on frontend wizard, light on backend; will be retired in same commit (replaced by `templates.md` underscore, mirrors on-disk dir).

## Pre-flight reads (done)

| Doc | Key takeaway |
|---|---|
| `wiki/README.md` | Modules indexed; existing stub `modules/templates.md` covers frontend wizard but lacks Arc42 + tech-debt register |
| `wiki/modules/templates.md` | Frontend-heavy stub; backend section is `## API surface — TBD`. This doc will replace it (rename to underscore) |
| `wiki/concepts/placeholders.md` | Fixed 7-token catalog enforced by `templates/application/validate_placeholders.go`; non-catalog names rejected at schema-save |
| `wiki/concepts/token-syntax.md` | `{name}` single-brace eigenpal-native; legacy `{{uuid}}` removed |
| `wiki/decisions/0001-eigenpal-adoption.md` | DOCX editor choice; templates own DOCX content for downstream document instantiation |
| `wiki/modules/documents.md` (cross-ref) | Documents are downstream consumers — every document instantiates from a template via `placeholder_schema_snapshot`; templates is upstream |
| `wiki/modules/approval.md` (cross-ref) | Approval module SoD probing needs `TemplateAuthorChecker` interface (iam T-003) — templates carry author identity that approval consumes |
| `wiki/modules/iam.md` + migration 0165 | IAM `template.*` capability namespace seeded: `template.view/create/edit/submit/approve/publish` |
| `wiki/architecture/api-design-system.md` | Plan 2 mid-rollout: RFC 9457 Problem envelope, cursor pagination, Stripe-model idempotency, two-tier authz |

## Module dir snapshot (top-level files)

```
internal/modules/templates/
├── api/                       # oapi-codegen output (api.gen.go, cfg.yaml, gen.go)
├── application/               # service layer + ports + validators
│   ├── approval_config.go
│   ├── autosave.go
│   ├── create.go
│   ├── lifecycle.go           # state-transition operations
│   ├── ports.go
│   ├── queries.go
│   ├── schema.go              # schema CRUD + ValidatePlaceholders
│   ├── service.go
│   ├── visibility_graph.go
│   └── *_test.go
├── delivery/http/             # HTTP routing + handlers
│   ├── handler.go
│   ├── routes_autosave.go
│   ├── routes_catalog.go
│   ├── routes_create.go
│   ├── routes_generated.go
│   ├── routes_lifecycle.go
│   ├── routes_query.go
│   ├── routes_schema.go
│   ├── errors.go
│   └── *_test.go
├── domain/                    # entities + invariants
│   ├── approval.go
│   ├── audit.go
│   ├── errors.go
│   ├── schemas.go
│   ├── template.go
│   ├── version.go
│   └── *_test.go
└── repository/
    ├── mappers.go
    ├── postgres.go
    └── postgres_integration_test.go
```

## Plan 2 path-rename note (record actual on-disk state today)

- API path: was `/api/v1/templates`, Plan 2 flipped to `/api/v1/templates` (commits ae1229e8..c84215f7).
- Code dir: still `internal/modules/templates/`. Rename to `internal/modules/templates/` is pending.
- Wiki filename: this doc lands as `templates.md`. Rename to `templates.md` happens in same commit as code rename — flagged in tech-debt as `maint:migration-cleanup`.

## Cross-deps to flag in Phase 6

- **OUT-edges (expected, confirm in Phase 3):** iam (`CapabilityChecker`, `authz.Require`), audit (writer), platform/httpclient, taxonomy (profiles), eigenpal (DOCX bytes pass-through).
- **IN-edges (expected):** documents (snapshot template at instantiation), approval (TemplateAuthorChecker contract for SoD), search (template index).
- **iam capability namespace:** `template.view/create/edit/submit/approve/publish` seeded in migration 0165 — applies dual-namespace debt T-001 mirror (same shape as documents, auth, approval).

## Phase 2 op picks (preliminary, confirm after Phase 1)

1. **Read** — list templates (`routes_query.go` → `Service.List` / `Repository.List`)
2. **Write** — update template schema (`routes_schema.go` → `Service.UpdateSchema` → `ValidatePlaceholders` → `Repository.SaveSchema`). Critical because placeholder catalog enforcement is the security boundary.
3. **State-transition** — submit → approve → publish (`routes_lifecycle.go` → `Service.Submit/Approve/Publish` → `Version.Transition` invariants → audit).

If Phase 1 reveals a different shape, repick and record.

## Severity rubric for templates (apply in Phase 6)

- **Critical:** template-injection or placeholder-escape (catalog bypass at schema-save or runtime); tenant leak via shared/global template; regulated audit-trail gap on publish (publish is the "approved-content" enforcement boundary — failure = loss of regulatory traceability).
- **Major:** RFC 9457 envelope absent on templates routes; OpenAPI/handler drift; concurrent autosave race; lifecycle idempotency gap.
- **Minor:** doc-cleanup; dep-bump; test-only; migration-cleanup (templates/ → templates/ rename); docs-link.

## Open questions for the user

None blocking — proceed.

## Skips recorded

None. All 8 phases will run.
