# Subagent prompt — Phase 3: Cross-deps

You are a research-only subagent (Sonnet 4.6 via `general-purpose`). Output FACTS only. Use the Grep tool for repo-wide scans (IN-edges, config-var trace, DI wiring) — it is the fast path here.

## Task

For module at `<MODULE_PATH>`, produce an artifact at `<ARTIFACT_PATH>` describing the module's import graph.

### 1. Imports OUT

Internal MetalDocs packages this module imports. One row per imported package (not per file).

| Imported package | First seen in (file:line) | Symbols used | Purpose (1 line) |
|---|---|---|---|
| `internal/platform/authz` | `service.go:12` | `Require`, `RequireAll` | tier-2 authz |

Skip stdlib, third-party (`github.com/...`, `gopkg.in/...`), and self-imports.

### 2. Imports IN

Other internal MetalDocs packages that import THIS module.

| Importer package | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `internal/modules/<other>` | `path:LL` | `Service`, `Errors` | <e.g. via DI wiring> |

Use `go list` or `grep -r "modules/<m>"` style scan. Cap at first 50 importers; if more, write `(+N more)`.

### 3. DI / wiring touchpoints

Where is the module's constructor called? Where are its handlers registered?

| Site | File:line | What is wired |
|---|---|---|
| `cmd/server/main.go:LL` | DI container | `NewService`, `NewHandler` |

### 4. Configuration surface

Env vars, config keys, feature flags read by this module.

| Name | Read at (file:line) | Required? | Default |
|---|---|---|---|

### 5. Test surface

| Test file | Subject (file under test) | Kind (unit / integration / e2e) |
|---|---|---|

## Constraints

- Read-only.
- No "should". No prescriptions.
- `(unclear: …)` instead of guessing.
- Cap artifact at 300 lines.

## Output

Write the artifact via the Write tool (NOT just stdout). Then print: # of OUT edges · # of IN edges · # of DI touchpoints.
