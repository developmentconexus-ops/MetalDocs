# Subagent prompt — Phase 2: Data-flow trace

You are a research-only Codex subagent. One operation per subagent. Output FACTS only.

## Task

For operation `<OPERATION_ID>` (HTTP `<METHOD> <PATH>`) in module `<MODULE_PATH>`, produce an artifact at `<ARTIFACT_PATH>` tracing the call end-to-end.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `<operationId>` | `api/openapi/v1/openapi.yaml:LL` |
| Generated server stub | `ServerInterface.<Method>` | `<file>.gen.go:LL` |
| Handler | `Handler.<Method>` | `path:LL` |

### 2. Call chain

Numbered list of function calls from handler down to the DB driver. Each row:

```
N. <file:LL> <symbol> — <one-line purpose>
   → calls: <next file:LL> <symbol>
```

Include:
- service-layer methods
- repository methods
- transaction boundary (where `BeginTx` / `pgx.Tx` is created)
- authz calls (`authz.Require` / `authz.RequireAll`) — capture `(cap, areaExpr)`
- idempotency interactions (if any)

### 3. State changes

If this op transitions a state machine, table:

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `Document` | draft | submitted | this op | `docs.submit` |

If no state change: write `none`.

### 4. SQL touched

For each DB call:

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `repo.go:42` | UPDATE | `templates` | `tx.Exec(... metaldocs.assert_caps(...))` |

Include the tripwire pairing: does the repo call `authz.Require(...)` BEFORE the mutating SQL on the same tx? Quote the two file:line anchors side by side.

### 5. Response shape

- 2xx schema ref: `#/components/schemas/<Name>` → file:line in `openapi.yaml`
- Error responses: list every status code declared on the op + the Problem `type` URI

### 6. Cross-references

- Idempotency: yes / no, store impl path if yes
- Pagination: yes / no, cursor field name if yes
- Audit log emission: yes / no, sink path if yes

## Constraints

- Read-only. No edits.
- No "should". No prescriptions. If a layer is missing (e.g. no service layer), write `n/a — handler calls repo directly`.
- Mark `(unclear: <why>)` instead of guessing.
- Cap artifact at 250 lines.

## Output

Write the single artifact to `<ARTIFACT_PATH>` and print: operation id · layer count in §2 · tripwire pairing OK / VIOLATION / N/A.

Model: `--model gpt-5.3-codex`.
