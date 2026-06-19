# F6.4 Evidence

## map[string]any grep — 5 target files (post-change)

```
security/delivery/http/handler.go       : 1  (struct field: Evidence map[string]any — domain mirror, not a response literal)
taxonomy/delivery/http/routes_areas.go  : 0
taxonomy/delivery/http/routes_families.go: 0
templates/delivery/http/routes_catalog.go: 0
templates/delivery/http/routes_schema.go : 0
```

The 1 remaining occurrence is `signalItem.Evidence map[string]any` — a typed struct field that mirrors `securitydomain.Signal.Evidence map[string]any`. The domain model owns this type; changing it is out of F6.4 scope. There are zero `writeJSON(..., map[string]any{...})` call-sites remaining.

## go build ./...
Exit 0. No output.

## Module tests
```
ok  metaldocs/internal/modules/security/application     3.6s
ok  metaldocs/internal/modules/taxonomy/delivery/http   6.7s
ok  metaldocs/internal/modules/templates/delivery/http  5.9s
(all other module sub-packages: ok or no test files)
```

## Full regression go test -count=1 ./...
All packages: ok or [no test files]. Zero FAIL lines.

## Status: PASS
