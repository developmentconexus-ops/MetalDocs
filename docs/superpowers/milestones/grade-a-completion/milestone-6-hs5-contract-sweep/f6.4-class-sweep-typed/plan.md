# F6.4 Plan

## Steps executed

1. Read all 5 target files + domain types (`securitydomain.MfaCoverage`, `Lockout`, `Signal`) + taxonomy handler interfaces + templatesapi structs.
2. Confirmed:
   - `Signal.Evidence` is `map[string]any` in domain — struct field must mirror it.
   - `MfaCoverage.MfaEnabledPct` is `float32` (not float64).
   - `areaService.List` returns `[]domain.ProcessArea` (value slice, not pointer slice).
   - `familyService.List` returns `[]domain.DocumentFamily`.
   - `templatesapi.PlaceholderCatalogEntry` fields: `Key`, `Label`, `Description`.
   - `templatesapi.TemplateVersionEnvelope.Data.Version VersionDTO` — already imported in handler.go.
3. Added typed struct block at top of security handler.go (after package/imports).
4. Replaced 6 `map[string]any` sites in security handler.go.
5. Added `listAreasResponse` to routes_areas.go; replaced 1 site.
6. Added `listFamiliesResponse` to routes_families.go; replaced 1 site.
7. Added `templatesapi` import to routes_catalog.go; replaced 1 site with conversion loop + typed response.
8. Added `templatesapi` import to routes_schema.go; replaced 1 site with `TemplateVersionEnvelope`.
9. `go build ./...` — clean.
10. Module tests — all pass.
11. Full regression `go test ./...` — all pass, zero FAIL.
