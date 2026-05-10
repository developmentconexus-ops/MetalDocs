## A
- **Pinned generator in this repo:** `github.com/oapi-codegen/oapi-codegen/v2 v2.7.0` (`go.mod:12`).
- **Current module configs generate both std-http and strict layers** (`generate.std-http-server: true`, `generate.strict-server: true`) in:
  - `internal/modules/registry/api/cfg.yaml`
  - `internal/modules/templates_v2/api/cfg.yaml`
  - `internal/modules/documents/api/cfg.yaml`
- **Per-operation dispatch templates (std-http + strict) in oapi-codegen v2.7.0:**
  - `.../pkg/codegen/templates/stdhttp/std-http-middleware.tmpl`
    - Emits `func (siw *ServerInterfaceWrapper) <OperationId>(...)` and calls `siw.Handler.<OperationId>(...)` per op.
  - `.../pkg/codegen/templates/strict/strict-http.tmpl`
    - Emits `func (sh *strictHandler) <OperationId>(...)` per op, decodes body, then calls `sh.ssi.<OperationId>(ctx, requestObj)`.
  - `.../pkg/codegen/templates/stdhttp/std-http-handler.tmpl`
    - Registers per-op routes to wrapper methods (`wrapper.<HandlerName>`). Not the decode site, but part of per-op dispatch wiring.
- **Evidence of template composition:** `.../pkg/codegen/operations.go`:
  - `GenerateStdHTTPServer` uses `std-http-interface.tmpl`, `std-http-middleware.tmpl`, `std-http-handler.tmpl`.
  - `GenerateStrictServer` (for std-http/chi/gorilla) uses `strict-interface.tmpl`, `strict-http.tmpl`.

## B
- **Yes, a user-template override can access OpenAPI extensions for an operation** via `OperationDefinition.Spec`:
  - `OperationDefinition` includes `Spec *openapi3.Operation` (`.../pkg/codegen/operations.go`, struct definition).
  - `opDef.Spec = op` when operation definitions are built (`.../pkg/codegen/operations.go`, opDef construction).
  - `openapi3.Operation` has `Extensions map[string]any` (`.../go/pkg/mod/github.com/getkin/kin-openapi@v0.135.0/openapi3/operation.go`).
- **Template syntax (Go `text/template`)**:
  - Access extension by key: `{{ $xa := index .Spec.Extensions "x-authz-area" }}`
  - Guard for presence: `{{ if $xa }} ... {{ end }}`
- **Security scopes are already exposed as first-class template data**:
  - `OperationDefinition.SecurityDefinitions []SecurityDefinition` with fields `ProviderName`, `Scopes` (`.../pkg/codegen/operations.go`).
  - Populated from OpenAPI `security` (`DescribeSecurityDefinition`, same file).
- **Important caveat:** `Extensions` values are `any`; template-only extraction of nested structures can be awkward/fragile (type assertions are limited in templates).

## C
- **Pseudocode snippet (strict std-http override of `strict/strict-http.tmpl`)** showing two-tier flow at required points.

```gotemplate
{{range .}}
{{$opid := .OperationId}}
func (sh *strictHandler) {{.OperationId}}(w http.ResponseWriter, r *http.Request{{genParamArgs .PathParams}}{{if .RequiresParamObject}}, params {{.OperationId}}Params{{end}}) {
    var request {{$opid | ucFirst}}RequestObject

    // ... existing path/params wiring ...

    // Tier 1: capability check at handler entry (before body decode)
    {{if .SecurityDefinitions}}
    {{/* assumes convention: bearerAuth with exactly one capability scope */}}
    {{range .SecurityDefinitions}}
    {{if eq .ProviderName "bearerAuth"}}
    if err := sh.capabilityService.Require(r.Context(), "{{index .Scopes 0}}", userIDFromContext(r.Context())); err != nil {
        sh.options.ResponseErrorHandlerFunc(w, r, err)
        return
    }
    {{end}}
    {{end}}
    {{else}}
    {{/* strict mode policy: missing security declaration -> fail */}}
    sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("authz not declared for {{.OperationId}}"))
    return
    {{end}}

    // ... existing request body decode into request.<...>Body ...

    // Tier 2: area check after decode, if x-authz-area exists
    {{$xa := index .Spec.Extensions "x-authz-area"}}
    {{if $xa}}
    {{/* parse area codes from decoded request body according to extension contract */}}
    areaCodes, cap := extractAreaCodesAndCapabilityFromExtensionAndRequest($xa, request)
    if err := authz.Require(r.Context(), cap, areaCodes); err != nil {
        sh.options.ResponseErrorHandlerFunc(w, r, err)
        return
    }
    {{end}}

    // Delegate to business handler
    handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
        return sh.ssi.{{.OperationId}}(ctx, request.({{$opid | ucFirst}}RequestObject))
    }
    // ... existing strict middlewares + response emission ...
}
{{end}}
```

- **UNCERTAIN — insufficient evidence** that pure template logic alone can robustly parse arbitrary nested `x-authz-area` objects without helper funcs; practical implementation may need generated helper Go code or constrained extension schema.

## D
- **Template ABI stability across upgrades:**
  - Not guaranteed as a formal stable ABI surface. Evidence:
    - README states backward compatibility is aimed for, but breaking behavior fixes can happen.
    - Active issue to clarify SemVer/breaking-change coverage (`https://github.com/oapi-codegen/oapi-codegen/issues` listing includes `#2168 docs: explicitly note what is/is not covered under SemVer / breaking change guarantees`).
  - Risk: medium.
- **Copy-paste maintenance debt (approach #1):**
  - High likelihood. Overriding `strict/strict-http.tmpl` typically means vendoring a large upstream template and carrying diffs.
  - Any upstream template changes (bugfixes, response/body decoding tweaks) must be manually merged.
- **Approach #2 (embedded-spec + post-process wrappers):**
  - Cleaner decoupling: keeps upstream templates untouched; authz generation logic lives in project-owned code.
  - Better upgrade posture: regenerate wrappers from spec metadata without rebasing large template files.
  - Tradeoff: additional generator program + wrapper integration complexity.
- **Known upstream warnings/discussions relevant to override risk:**
  - Custom templates are supported, but filenames must match exactly (`README`, user-templates section).
  - Remote template sourcing is flagged as reproducibility risk and planned to be disabled-by-default in future (`issue #1564`: https://github.com/oapi-codegen/oapi-codegen/issues/1564).
  - No direct upstream statement found saying “do not override templates”; risk conclusion is based on maintenance mechanics + SemVer caveats above.

## E
- **Recommendation: approach #2 (embedded-spec + post-process wrapper generation)** for the long-lived Track D requirement.
- **Why (simplest reliable over lifecycle):**
  - Avoids maintaining a forked copy of `strict/strict-http.tmpl`.
  - Preserves clean upgrades of `oapi-codegen`.
  - Keeps authz policy generation explicit, testable, and isolated from template internals.
  - Still gives static guarantee if CI enforces wrapper generation drift + spec lint (`security` and `x-authz-area` presence).
- **When approach #1 is acceptable:**
  - Short-term spike/prototype, pinned generator version, and willingness to rebase template diffs on upgrades.
- **Fallback (hand-written `authz.Require` calls):**
  - Lowest tooling effort now, but loses strong compile-time/spec-driven guarantee and invites drift; use only if timeline blocks generator work.

### Evidence links
- Local repo docs:
  - `docs/superpowers/specs/2026-05-10-api-design-system.md` (section 4.5)
  - `docs/superpowers/plans/2026-05-10-api-design-system-plan-1-primitives.md` (Task 7, Step 7.1)
  - `wiki/architecture/api-contract.md`
  - `wiki/decisions/0007-two-tier-authz.md` (exists)
- oapi-codegen local module cache (v2.7.0):
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/oapi-codegen/oapi-codegen/v2@v2.7.0/pkg/codegen/operations.go`
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/oapi-codegen/oapi-codegen/v2@v2.7.0/pkg/codegen/templates/stdhttp/std-http-middleware.tmpl`
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/oapi-codegen/oapi-codegen/v2@v2.7.0/pkg/codegen/templates/stdhttp/std-http-handler.tmpl`
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/oapi-codegen/oapi-codegen/v2@v2.7.0/pkg/codegen/templates/strict/strict-http.tmpl`
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/oapi-codegen/oapi-codegen/v2@v2.7.0/README.md`
- kin-openapi local module cache:
  - `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/go/pkg/mod/github.com/getkin/kin-openapi@v0.135.0/openapi3/operation.go`
- Upstream issue/discussion URLs:
  - https://github.com/oapi-codegen/oapi-codegen/issues/1564
  - https://github.com/oapi-codegen/oapi-codegen/issues
