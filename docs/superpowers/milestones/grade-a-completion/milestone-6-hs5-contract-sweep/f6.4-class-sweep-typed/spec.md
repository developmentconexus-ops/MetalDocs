# F6.4 Spec — Class Sweep: Security / Taxonomy / Templates Catalog & Schema Typed Responses

## Goal
Eliminate all ad-hoc `map[string]any` response construction from five handler files, replacing with typed Go structs that guarantee JSON shape at compile time.

## Scope — 5 files
1. `internal/modules/security/delivery/http/handler.go`
2. `internal/modules/taxonomy/delivery/http/routes_areas.go`
3. `internal/modules/taxonomy/delivery/http/routes_families.go`
4. `internal/modules/templates/delivery/http/routes_catalog.go`
5. `internal/modules/templates/delivery/http/routes_schema.go`

## Contract
- Zero `writeJSON(w, ..., map[string]any{...})` call-sites in scope files after change.
- Typed struct fields that happen to carry `map[string]any` values (e.g. `signalItem.Evidence` mirroring `domain.Signal.Evidence`) are acceptable — the domain model owns that type.
- `go build ./...` clean.
- `go test -count=1 ./internal/modules/security/... ./internal/modules/taxonomy/... ./internal/modules/templates/...` all pass.
- Full `go test -count=1 ./...` regression clean.

## Typed structs introduced

### security/delivery/http (local package `httpdelivery`)
- `mfaCoverageByRoleItem` — mirrors `securitydomain.MfaRoleSlice`
- `mfaCoverageResponse` — top-level MFA coverage envelope
- `lockoutItem` — single lockout row (time pointers serialized as RFC3339 strings)
- `lockoutsResponse{Items []lockoutItem}`
- `signalItem` — single signal row; `Evidence map[string]any` retained to mirror domain
- `signalsResponse{Items []signalItem}`

### taxonomy/delivery/http
- `listAreasResponse{Items []domain.ProcessArea}` in routes_areas.go
- `listFamiliesResponse{Items []domain.DocumentFamily}` in routes_families.go

### templates/delivery/http
- routes_catalog.go: uses `templatesapi.PlaceholderCatalogResponse{Items []PlaceholderCatalogEntry}`
- routes_schema.go: uses `templatesapi.TemplateVersionEnvelope` (added F6.1)
