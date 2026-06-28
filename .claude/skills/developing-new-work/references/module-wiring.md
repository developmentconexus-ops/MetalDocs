# Module wiring — birth a new module (ordered)

**Last verified:** 2026-06-28

A new bounded-context module under `internal/modules/<name>/` is wired in this order. Exemplars:
**taxonomy** (smallest complete module, has a `module.go`), **templates** (no `module.go`). The
composition root is `apps/api/cmd/metaldocs-api/main.go`.

1. **Folders** — `{api, application, domain, delivery/http, infrastructure}` under
   `internal/modules/<name>/`.

2. **Domain** — entities + `port.go` interfaces (the provider ports this module publishes). Pure
   domain; no SQL, no HTTP.

3. **Application** — the application service + `ports.go` (the consumer ports this module needs from
   others). Orchestrates; owns the tx boundary via `TxRunner`.

4. **Infrastructure** — the repository. **Touches only this module's own tables.** Sets the authz GUC
   and calls `authz.Require` in-tx. Never reads another module's tables.

5. **Delivery** — `Handler` + `RegisterRoutes(mux)`. Thin: decode → call application service → write
   response via `problem`/`httpresponse`.

6. **api codegen** — `api/cfg.yaml` (`include-tags` for this module) + `gen.go`; regenerate from the
   spec.

7. **OpenAPI** — add the module's `tags:` entry in `api/openapi/v1/openapi.yaml` and tag **every** route
   with it. Contract-first: the route exists in the spec before it exists in Go.

8. **`module.go` (optional)** — `New(Dependencies)` constructor; **panic on nil deps** (fail fast at
   composition, not at first request). Follow taxonomy's shape.

9. **Composition root** — wire the module in `apps/api/cmd/metaldocs-api/main.go` (+ the worker and
   jobs binaries if it has async consumers or recurring janitors).

10. **Migration** — `db/migrations/0NNN_*.sql`: tables with `tenant_id`, plus the DB constraints/
    triggers that enforce the module's invariants.

11. **Docs** — `wiki/modules/<name>.md` (12-section structure) + `<name>-tech-debt.md` +
    an entry in `wiki/modules/index.md`. See `docs-adr-governance.md`.

**Boundary rule throughout:** cross-module access is via published interfaces only (invariant 6). If
the new module needs data from an existing one, depend on that module's port — add a method to its
`port.go` if needed — never query its tables. If a consumer (e.g. render) must not depend on this new
module, that direction constraint is a locked design constraint, record it in §1 of the artifact.
