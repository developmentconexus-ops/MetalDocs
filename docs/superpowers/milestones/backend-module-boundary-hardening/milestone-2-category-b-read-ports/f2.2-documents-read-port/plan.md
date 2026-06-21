# Feature F2.2 — Plan — documents active-instance read-port

> Input: `spec.md` (approved). Engine: subagent-driven-development run inline (TDD).

## Plan

### Files touched
- **New** `internal/modules/documents/domain/active_instance_port.go` — `ActiveInstanceView` struct +
  `ActiveInstanceReader` interface (+ `NoopActiveInstanceReader` for default wiring/tests).
- **New** `internal/modules/documents/infrastructure/active_instance_reader.go` — `ActiveInstanceReaderPG`
  adapter holding `*sql.DB`; `NewActiveInstanceReaderPG(db)`. Runs the identical FULL OUTER JOIN
  projection + derived approval lookup, status literals → owner typed constants as params.
- **New** `internal/modules/controlleddocuments/infrastructure/active_instance_parity_integration_test.go`
  — `TestActiveInstanceReader_ParityWithRawGetActiveInstance` (raw-SQL baseline vs port across
  active-only / published-only / under-review / none).
- **Edit** `controlleddocuments/infrastructure/repository.go` — struct gains `activeInstance
  documentsdomain.ActiveInstanceReader`; `NewPostgresControlledDocumentRepository(db, activeInstance)`;
  `GetActiveInstance` body becomes port-call + 1:1 map; delete the two raw queries.
- **Edit** `controlleddocuments/module.go` — construct `NewActiveInstanceReaderPG(deps.DB)`, pass to repo.
- **Edit** the 3 CD repo test call sites — pass `infrastructure.NewActiveInstanceReaderPG(db)` (real,
  for integration tests) or `documentsdomain.NoopActiveInstanceReader{}` (unit).
- **Edit** `tools/cilint/internal/analyzers/hgcrossmodule.go` — remove B2/B3/B4 entries.

### Ordering (parity-before-delete, D6)
1. Add `ActiveInstanceView` + `ActiveInstanceReader` (+ Noop) to `documents/domain`.
2. Implement `ActiveInstanceReaderPG` adapter (identical projection, typed-constant params).
3. Write parity test: seed CD + documents in each state; assert raw `GetActiveInstance`-shape SQL ==
   `port.ActiveInstanceForControlledDocument`. Run RED→GREEN (compile against new wiring).
4. Wire the port into CD repo + module + composition roots; map view→`ActiveDocumentInstance`.
5. **Parity green** → delete the raw `documents`/`document_revisions`/`approval_instances` reads from
   `GetActiveInstance`.
6. Remove B2/B3/B4 from `hgPendingRemediation`; `cilint` exit 0.
7. `go build ./...`; targeted + module test runs; `git grep` proofs; `evidence.md`.

### Test strategy
- Real PG (:5434) parity test is the binding acceptance. Cases: active-only (no published),
  published-only (no active), under_review (active + in-progress approval instance present), none
  (neither → nil,nil). Content-hash fallback exercised by an active row with NULL
  `content_hash_at_submit` but a `document_revisions` row.
- Unit CD repo tests that don't exercise `GetActiveInstance` get `NoopActiveInstanceReader{}`.

### Import-cycle / boundary check
- `documents/domain` imports only stdlib/platform — must NOT import controlleddocuments. Verify `go build`.
- CD `infrastructure` → `documents/domain` (consumer→owner interface) is the only new edge.

### Risks
- **Projection parity (largest port).** Mitigated by preserving the exact SQL (casts, ORDER BY, LIMIT,
  COALESCE subquery) and only swapping status literals for typed-constant params.
- **Status-param equivalence.** `status IN ($3..$7)` with the 5 active constants must equal the literal
  IN-list; published `= $8`, approval `= $n`. Parity test's under_review + active cases lock this.
