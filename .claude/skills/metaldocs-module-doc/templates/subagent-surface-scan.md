# Subagent prompt — Phase 1: Surface scan

You are a research-only Codex subagent. Output FACTS only — no recommendations, no "should", no "consider".

## Task

For the Go module at `<MODULE_PATH>`, produce an artifact at `<ARTIFACT_PATH>` containing four sections.

### 1. File tree

`tree`-style listing of the module (depth ≤ 3, group by subpackage). Each row: relative path + 1-line purpose inferred from the file's leading doc comment or top-level types. If no doc comment, write `(undocumented)`.

### 2. Public surface

Use Go AST. For every exported symbol (rune uppercase) at top-level, table row:

| File:line | Kind | Name | Signature / receiver | Doc comment first line |
|---|---|---|---|---|
| `path:LL` | func / type / const / var / iface | `Name` | `func (r Recv) Name(args) (rets)` | first line, or `(undocumented)` |

Include methods on exported receiver types. Skip generated files (`*.gen.go`, anything matching `//go:generate` output).

### 3. HTTP operations

If the module wires routes (look for `chi`, `mux`, `httprouter`, or generated `RegisterHandlers` calls):

| Method | Path | Handler symbol | Source file:line |
|---|---|---|---|
| GET | `/api/v1/...` | `Handler.ListX` | `path:LL` |

If routes come from oapi-codegen, list ops by operationID, even if the path stitching is in generated code.

### 4. Migration list

If the module has its own SQL migrations (under `internal/modules/<m>/migrations/` or referenced from a registry file), list them:

| Filename | Verb | Tables touched |
|---|---|---|
| `0001_init.sql` | CREATE TABLE | `templates` |

If migrations live elsewhere, write `migrations: external — see persistence artifact`.

## Constraints

- Read-only. Do NOT edit any source file.
- Do NOT speculate. If a fact is unclear, write `(unclear: <reason>)` instead of guessing.
- Do NOT propose changes. No "should", "recommend", "consider", "professional", "industry-standard".
- Keep the artifact under 400 lines. If the module is larger, prioritise depth in §2 and §3; truncate §1 with a footnote.

## Output

Write a single file to `<ARTIFACT_PATH>`. After write, print the path and a one-line summary: total exported symbols · total HTTP operations · total migrations.

Model: `--model gpt-5.3-codex`.
