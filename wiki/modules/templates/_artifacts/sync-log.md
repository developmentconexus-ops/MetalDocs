# Sync log â€” templates

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-17 - active v2 reference memory sync

- **Context:** post-merge scan after templates route/name polish and wizard DOCX import completion; active backlog/roadmap text still referenced `/api/v2/templates` for future work.
- **Mode:** lite patch.
- **Affected-surface scan:** `rg` across active wiki/backlog/architecture/module docs; historical migrations, release inventory, runbooks, and test artifacts intentionally left unchanged.
- **Routes/API:** no runtime route change; active future template endpoints canonicalized to `/api/v1/templates/*`.
- **Runtime flows:** none.
- **Persistence:** none.
- **Debt/backlog:** template editor and novo-template-wizard backlog route text aligned with production v1; roadmap Plan 9/10 examples aligned to v1.
- **Tally gate:** preflight PASS before edits; final tally recorded in session output.
- **Patched files:** `wiki/backlog/template-editor.md`; `wiki/backlog/novo-template-wizard.md`; `wiki/backlog/roadmap.md`; `wiki/modules/templates/_artifacts/sync-log.md`.

## 2026-05-17 - Template wizard DOCX import + permission simplification sync

- **Context:** uncommitted implementation diff for `/templates/new`, template create/list OpenAPI/domain/API cleanup, frontend generated types/wrapper changes, and browser validation of blank + imported DOCX Eigenpal editor handoff.
- **Mode:** structural refresh.
- **Affected modules:** templates only. Editor UI was already updated by the DOCX runtime repair and was not further changed in this sync.
- **Affected-surface scan:** frontend wizard state/components/API wrapper, backend OpenAPI/generated API/create/list/domain/repository mappings, persistence compatibility columns, wizard backlog, list-flow artifact, public-surface artifact, and persistence artifact.
- **Routes/API:** `TemplateDTO` and create/list behavior no longer expose or accept creator-scoped `visibility`, `areas`, or `specific_areas`; create responses still return `data.template` + `data.version`.
- **Runtime flows:** `/templates/new` is four steps; DOCX start stores the selected file until submit, creates the template/version, uploads through `/autosave/presign`, commits through `/autosave/commit`, then opens Eigenpal on `/templates/{id}/versions/{n}`. Blank start opens Eigenpal with a blank draft.
- **Persistence:** existing `templates_template.areas`, `visibility`, and `specific_areas` columns remain inert compatibility fields; repository inserts fixed empty/public values while runtime/API selection ignores them.
- **Debt/backlog:** no new T/R debt rows opened. Wizard backlog marked DOCX upload and editor handoff resolved, and former permissions/visibility API rows removed by product decision.
- **Verification:** targeted frontend template tests, frontend build typecheck, editor-ui tests/typecheck, backend `go generate`, backend templates tests, and browser runtime validation completed. Redocly lint was blocked by npm `ECOMPROMISED`; broad frontend suite still has unrelated pre-existing failures.
- **Tally gate:** PASS with Git Bash `PATH="/usr/bin:/bin:$PATH"`; severity count 4/6/4 matched, missing-ADR register count 11 (script could not parse a stated doc count and still exited 0).
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/backlog/novo-template-wizard.md; wiki/modules/templates/_artifacts/01-surface.md; wiki/modules/templates/_artifacts/02-flow-list.md; wiki/modules/templates/_artifacts/04-persistence.md; wiki/modules/templates/_artifacts/sync-log.md

## 2026-05-17 - DOCX import runtime repair sync

- **Context:** uncommitted fix for Eigenpal `.docx` import failing under Docker local runtime
- **Mode:** structural refresh
- **Affected surfaces:** templates autosave/import service path, repository draft CAS helper, Docker MinIO local endpoint/CORS bootstrap
- **Runtime facts:** browser could not fetch `http://minio:9000` signed URLs; compose now signs with `host.docker.internal:9000` and starts `minio-init` to create the attachments bucket. MinIO server CORS allows local Vite origins.
- **Authz/tripwire:** `SaveTemplateDraft` and `CommitAutosave` now run version writes and audit rows inside `template.edit` authz transactions when `s.db != nil`, including transaction-local tenant/actor GUC setup.
- **T-NNN touched:** T-001 evidence extended for autosave/import; T-010 partially closed for generated `SaveTemplateDraft` CAS enforcement while legacy `/autosave/commit` remains hash-gated.
- **R-NNN touched:** R-001 note extended; R-010 marked merged (partial).
- **Verification:** `go test ./internal/modules/templates/application -count=1 -timeout=60s`; `go test ./internal/modules/templates/delivery/http -count=1 -timeout=60s`; Docker API rebuilt/restarted healthy; runtime smoke showed MinIO PUT 200 and `/autosave/commit` 200.
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/modules/templates/_artifacts/sync-log.md

## 2026-05-16 - Plan 12.4 template wizard stabilization sync

- **Context:** uncommitted Plan 12.4 stabilization diff for `/templates/new`, template create contract, generated API surfaces, placeholder catalog wrapper, migration tripwire refresh, and local startup fallback
- **Mode:** structural refresh
- **Anchors moved:** generated API shape for `POST /api/v1/templates`; placeholder catalog canonical route documented
- **Public surface:** `POST /api/v1/templates` response documented as `data.template` + `data.version`; frontend wrapper types derive from generated OpenAPI types
- **Routes/API:** bundled and partial OpenAPI include the mounted template route set; `api.gen.go` and frontend API types regenerated; placeholder catalog now has generated `PlaceholderCatalogResponse`
- **Runtime flows:** runtime smoke verified authenticated `POST /api/v1/templates` with `Idempotency-Key` and `doc_type_code: "POP"` returns HTTP 201 and editor redirect path
- **Persistence:** migration 0203 refreshes the authz tripwire function for renamed template tables
- **Dependencies:** `scripts/start-api.ps1` now falls back to `go run` when Windows denies execution of the repo-local built `.exe`
- **T-NNN touched:** T-006 closed for route/spec/generated coverage and catalog response typing; T-009 narrowed to replay-audit debt after create-path header coverage
- **R-NNN touched:** R-006 marked merged for contract coverage; R-009 remains open for same-key replay audit
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/backlog/novo-template-wizard.md; wiki/modules/templates/_artifacts/sync-log.md

## 2026-05-15 - Plan 12.4 novo-template-wizard create-path sync

- **Context:** uncommitted Plan 12.4 implementation diff for `/templates/new` plus runtime prerequisite repair in `internal/modules/templates/application/create.go`
- **Mode:** structural refresh
- **Anchors moved:** 1; `CreateTemplate` authz line moved after transaction GUC setup
- **Public surface:** no exported API change; added unexported `setAuthzGUC` helper
- **Routes/API:** no public route or OpenAPI shape change; documented `POST /api/v1/templates` idempotency wrapper and verified create response
- **Runtime flows:** added CreateTemplate runtime flow and Plan 12.4 smoke evidence
- **Persistence:** transaction-local `metaldocs.tenant_id` / `metaldocs.actor_id` GUC setup documented for create path
- **Dependencies:** none
- **T-NNN touched:** T-001 evidence updated for create-path GUC setup; T-009 text corrected to partial wrapper support with replay audit still open
- **R-NNN touched:** R-009 wording updated to replay-audit follow-up
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/modules/templates/_artifacts/sync-log.md
## 2026-05-14 - Plan 12.1 templates screen reality-first sync

- **Context:** commits `12188f98..eea76b14` (plan docs + templates screen implementation + backlog/design notes sync)
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** none (frontend wiring only; no templates backend route/contract delta)
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** R-009 wording updated to replay-audit follow-up
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates/_artifacts/sync-log.md
## 2026-05-11 Â· Plan 6a â€” close T-013

- **Context:** Plan 6a (commit 71a2dc53) Â· AppendAudit now calls canonical auditdomain.Writer instead of inserting to local templates_audit_log
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-013 Â· evidence: Repository.AppendAudit body replaced â€” calls r.audit.Record(ctx, auditdomain.Event{...}); WithAudit setter added; wired in main.go
- **R-NNN updated:** R-013 â†’ merged Â· commit 71a2dc53
- **Â§11 counts after:** Critical=4 Major=6 Minor=4 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates-tech-debt.md Â· wiki/backlog/templates-refactor.md
- **Structural changes noted (sweep needed):** WithAudit setter new on Repository; auditdomain OUT-edge added to templates/repository â€” Â§5 Key Files + Â§8 cross-deps not yet updated

## 2026-05-13 - Plan 10 route canonicalization + templates rename sweep

- **Context:** uncommitted Plan 10 implementation diff (module rename, API v1 sweep, final permission remediation)
- **Mode:** structural refresh
- **Anchors moved:** internal/modules/templates -> internal/modules/templates; /api/v2/templates* -> /api/v1/templates*
- **Public surface:** updated route prefix and capability mapping references
- **Routes/API:** templates and signed endpoints documented under /api/v1
- **Runtime flows:** unchanged behavior; canonical path updates only
- **Persistence:** none
- **Dependencies:** composition root + permission resolver references updated
- **T-NNN touched:** T-012/T-010 closure evidence aligned for Plan 10 sweep
- **R-NNN touched:** R-100/R-101 alignment notes refreshed
- **Counts after:** Critical=4 Major=6 Minor=4; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/templates.md; wiki/modules/templates-tech-debt.md; wiki/backlog/templates-refactor.md; wiki/modules/templates/_artifacts/*
