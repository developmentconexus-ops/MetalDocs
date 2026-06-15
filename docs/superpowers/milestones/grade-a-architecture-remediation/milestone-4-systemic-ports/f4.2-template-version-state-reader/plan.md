# Feature F4.2 — TemplateVersionStateReader (extend templates port)

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.2-template-version-state-reader`
> **Status:** Planning

This is the feature's **execution plan** — the "how". Contract + acceptance live in `spec.md`.

## Source

- Milestone spec row (F4.2, HS-6-amended): extend the existing templates `TemplateVersionPort` with a
  raw-state read; delete CD's `PostgresTemplateVersionChecker` and wire the templates reader; replace
  the `documents_adapters.go:113` `status := "published"` hardcode. Validate: 0 `templates_*` SQL under
  `controlleddocuments/`; 0 `status := "published"` in `wiring/`; consumer contract preserved; reads
  live + off-tx.
- Governing-spec reference: §5.2 / §6 H-G class (reach-without-a-port + hardcoded-domain-state); design
  D4/Approach-3 (reads live, no snapshot).

## Plan

Ordering is TDD-first per slice; each slice builds green before the next.

### Slice 1 — Extend the templates port (producer)
1. **RED:** add `TestTemplateVersionReader_GetTemplateVersionState_Live` (`-tags integration`) in
   `internal/modules/templates/infrastructure/template_version_reader_test.go` (or a new
   `_integration_test.go`): seed a template + version (published), assert
   `GetTemplateVersionState(ctx, tenantID, versionID)` → `(*"published", "<doc_type_code>", nil)`;
   NULL-status version → `(nil, docType, nil)`; absent id → `(nil, "", nil)`; other-tenant id →
   `(nil, "", nil)`. Compiles-but-fails (method absent).
2. **GREEN:** add method to `templates/domain.TemplateVersionPort`:
   `GetTemplateVersionState(ctx context.Context, tenantID, versionID string) (*string, string, error)`.
   Implement on `templates/infrastructure.TemplateVersionReader` reusing `templateVersionQuery`
   (explicit `tenantID` arg, not `tenant.FromContext`); NullString → `*string` (nil when invalid),
   not-found → `(nil, "", nil)`. Keep `IsPublished` untouched.
3. Verify taxonomy still builds (it depends only on `IsPublished`).

### Slice 2 — Delete CD checker, wire templates reader (consumer 1)
1. **RED:** ensure CD override tests (`service_test.go` published/draft/obsolete/profile-mismatch cases)
   still target `application.TemplateVersionChecker`; they use `fakeTemplateVersionChecker` so they stay
   green by construction — the behavior contract is the guard. Add a thin compile-time assertion that
   `*templatesinfra.TemplateVersionReader` satisfies `controlleddocuments/application.TemplateVersionChecker`
   (a `var _ application.TemplateVersionChecker = (*templatesinfra.TemplateVersionReader)(nil)` in CD
   module or a test) — RED until the method exists (already added in Slice 1) and wiring compiles.
2. **GREEN:** delete `PostgresTemplateVersionChecker` (struct + `NewPostgresTemplateVersionChecker` +
   `GetTemplateVersionState`) from `controlleddocuments/infrastructure/repository.go`. In
   `controlleddocuments/module.go`, import `templatesinfra` and set
   `tplCheck := templatesinfra.NewTemplateVersionReader(deps.DB)`.
3. Confirm `service.go:209/308` unchanged; build CD + wiring.

### Slice 3 — Replace the hardcode (consumer 2)
1. **RED:** unit test on `profileDefaultsAdapter.GetDefaultTemplateVersionID` with an injected fake
   port returning status `"obsolete"` for the profile's default version → assert returned status is
   `"obsolete"` (fails today: hardcode returns `"published"`).
2. **GREEN:** give `profileDefaultsAdapter` a port field (templates state reader interface, defined
   consumer-side in `wiring` or reuse `docapp` contract); `NewProfileDefaults` takes the reader;
   `GetDefaultTemplateVersionID` calls `GetTemplateVersionState(ctx, tenantID, *defaultVersionID)` for
   the real status instead of `status := "published"`. Update the `NewProfileDefaults` call site in
   `main.go` to pass the templates reader.
3. Preserve `(nil, nil, nil)` when the profile has no default template (unchanged early return).

### Slice 4 — Class proof + gates
1. `grep -rn "templates_template" internal/modules/controlleddocuments/` → 0 SQL (comments ok).
2. `grep -rn 'status := "published"' apps/api/internal/wiring/` → 0.
3. `go build ./...`; `go vet` (plain + `-tags integration`) on templates / controlleddocuments / wiring.
4. Run: templates + taxonomy + CD application suites; the new integration test (live PG via
   `start-api.ps1` DB or `METALDOCS_DATABASE_URL`).
5. H-PRE-1 runtime: exercise a CD-create override path; confirm via `pg_locks` the status read is not
   inside the create lock (call sites unmoved; reader uses its own pool conn). Reuse the M3 F3.2
   evidence pattern.
6. backend-api-qa-checklist + workflow-async-qa-checklist.

## Files touched (expected)

- `internal/modules/templates/domain/template_version_port.go` — add method to interface.
- `internal/modules/templates/infrastructure/template_version_reader.go` — add impl method.
- `internal/modules/templates/infrastructure/template_version_reader_test.go` (+ integration) — new test.
- `internal/modules/controlleddocuments/infrastructure/repository.go` — **delete** checker.
- `internal/modules/controlleddocuments/module.go` — wire templates reader as `tplCheck`.
- `apps/api/internal/wiring/documents_adapters.go` — inject port, drop hardcode.
- `apps/api/cmd/metaldocs-api/main.go` — pass templates reader into `NewProfileDefaults` (and confirm
  the single `NewTemplateVersionReader` instance / construct as needed for CD too).
- adapter unit test (wiring) — new/updated.

## Execution notes

- Watch wiring: `main.go:675` already constructs `templatesinfra.NewTemplateVersionReader(deps.SQLDB)`
  for taxonomy's `TplChecker`. Prefer one shared reader instance threaded to taxonomy + CD + the
  profile-defaults adapter rather than three constructions — but CD's `module.New` currently builds its
  own deps from `deps.DB`; decide minimal-churn wiring at implementation (either pass the reader into CD
  `Dependencies` or construct inside `module.New`). Keep it surgical.
- `IsPublished` explicit-tenant migration deliberately deferred (non-goal).
