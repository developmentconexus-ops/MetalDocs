# oapi-codegen How-To

> **Last verified:** 2026-05-08
> **Scope:** Operational guide for running oapi-codegen v2 in the MetalDocs backend: regenerate, vendor-mode gotcha, add a new module, include-tags filter.
> **Out of scope:** Overall contract-first architecture (`architecture/api-contract.md`), ADR rationale (`decisions/0012-contract-first-api.md`).
> **Key files:**
> - `internal/modules/controlleddocuments/api/cfg.yaml:1` - canonical controlled-documents config example
> - `internal/modules/controlleddocuments/api/gen.go:1` - `//go:generate` invocation
> - `go.mod:12` - `github.com/oapi-codegen/oapi-codegen/v2 v2.7.0`

---

## Regenerate an existing module

```bash
# From repo root. Use GOFLAGS=-mod=mod because the project has a vendor/ directory;
# without it, go generate refuses to resolve the oapi-codegen binary dependency.
GOFLAGS=-mod=mod go generate ./internal/modules/controlleddocuments/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...

# Or regenerate all modules at once (what CI does - CI sets its own go cache, no vendor issue):
go generate ./...
```

Commit the resulting `api.gen.go`. CI (`api-contract.yml`) will fail the PR if the committed file is stale.

---

## Vendor-mode note

The project uses `vendor/` (`go.mod` lists `github.com/oapi-codegen/oapi-codegen/v2`). In vendor mode, `go run github.com/oapi-codegen/...` inside `//go:generate` invocations requires the module to be resolvable. Set `GOFLAGS=-mod=mod` to bypass the vendor restriction for the generator binary itself. The generated output (Go source) is committed to the repo so production builds never invoke the generator.

---

## Add a new module

1. Define canonical ownership before editing:
   - canonical module owner
   - canonical public path namespace
   - canonical operationId set
   - canonical OpenAPI tag used by `include-tags`

2. Author operations in `api/openapi/v1/openapi.yaml` with the canonical `tags: [<module>]` value.

3. Create `internal/modules/<x>/api/cfg.yaml`:

   ```yaml
   package: <x>api
   generate:
     models: true
     std-http-server: true
     strict-server: true
     embedded-spec: true
   output: api.gen.go
   output-options:
     include-tags:
       - <module-tag>
   ```

4. Create `internal/modules/<x>/api/gen.go`:

   ```go
   package <x>api

   //go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../../../api/openapi/v1/openapi.yaml
   ```

5. Run:

   ```bash
   GOFLAGS=-mod=mod go generate ./internal/modules/<x>/api/...
   ```

6. Implement `ServerInterface` on the handler struct. Wire via generated boundary (`ServerInterfaceWrapper` or `HandlerWithOptions`) so generated public modules do not expose raw public mux ownership (see `internal/modules/controlleddocuments/delivery/http/handler.go:72` for the canonical pattern).

7. Verify freeze-law alignment for touched endpoints before commit:
   - runtime route owner/path
   - OpenAPI namespace/path/operation/tag
   - generated backend package/interface + mounted generated boundary
   - generated frontend wrappers/types for generated routes
   - module wiki status used for planning

8. Commit `api.gen.go` along with the handler changes.

---

## include-tags filter

Each module's `cfg.yaml` uses `include-tags` to scope the generated file to only the operations tagged for that module. Without this, every module would regenerate the entire spec. Operations must have a matching `tags:` value in the spec for them to appear in the generated output.

Example: `include-tags: [controlled-documents]` causes only operations with `tags: [controlled-documents]` in `openapi.yaml` to be included in `controlleddocumentsapi/api.gen.go`.

---

## See also

- `wiki/architecture/api-contract.md` - full architecture overview, runtime enforcement gaps, CI drift guard, freeze-law truth
- `wiki/decisions/0012-contract-first-api.md` - ADR and root-cause analysis
- `wiki/backlog/contract-first-followups.md` - deferred modules and migration template
