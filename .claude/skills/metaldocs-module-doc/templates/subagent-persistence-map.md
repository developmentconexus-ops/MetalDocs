# Subagent prompt — Phase 4: Persistence map

You are a research-only Codex subagent. Output FACTS only.

## Task

For module at `<MODULE_PATH>`, produce an artifact at `<ARTIFACT_PATH>` mapping its Postgres persistence surface.

### 1. Tables owned

Tables created by this module's migrations (or the global migration set, scoped by table-name prefix / ownership comment).

| Table | Created in (migration filename) | Notes |
|---|---|---|

For each table, sub-table:

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|

### 2. Tables read or written but NOT owned

Foreign tables this module touches.

| Table | Owner module | Read / Write | Operations using it |
|---|---|---|---|

### 3. Triggers, GUCs, functions

List Postgres triggers, functions, and GUCs (`SET LOCAL ...`) the module installs or relies on.

| Object | Kind (trigger / function / GUC) | File:line | Purpose |
|---|---|---|---|
| `metaldocs.asserted_caps` | GUC | `migrations/00NN_authz_tripwire.sql:LL` | tripwire enforcer |

### 4. Indexes

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|

### 5. Tripwire pairing audit

For every repo method that mutates a table owned by this module:

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|

If a row shows `Authz.Require called? = NO` and verb ∈ {INSERT, UPDATE, DELETE}: flag with `VIOLATION` in the row.

### 6. Migration history

Chronological list of migrations affecting this module.

| Order | Filename | Verb summary | Date (from filename or commit) |
|---|---|---|---|

## Constraints

- Read-only.
- No "should". No prescriptions.
- `(unclear: …)` instead of guessing.
- Cap artifact at 400 lines.

## Output

Write the artifact and print: # tables owned · # tripwire violations (if any) · # migrations.

Model: `--model gpt-5.3-codex`.
