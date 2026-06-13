# Wave Z — Sealed-Scope Backend Close-Out — Implementation Plan

> **For agentic workers:** Execution model is defined by the SPEC §4 (`docs/superpowers/specs/2026-06-12-wave-z-sealed-closeout-design.md`) — ultracode workflows, **sonnet** implement/review agents, **haiku** mechanical checks, **NEVER fable workers**, **max 15 concurrent agents**. Steps use checkbox (`- [ ]`) syntax. One commit per task (Z-ID in message). The orchestrator session reads: `CLAUDE.md` → `wiki/references/current-agent-handoff.md` → the spec → THIS PLAN.

**Goal:** Empty the sealed manifest Z-1..Z-33 so the DONE gate G1–G11 (spec §2) passes and the MetalDocs backend is closed forever (feature work only afterward).

**Architecture:** Verification-gated phases P0–P6. P1 fans out independent small fixes grouped by module ownership (no two agents share a file). P2 lands the two runtime-verified big items (OTel, full RLS). P3 serializes codegen-touching items. P4 runs the three structural M-items. P5 is docs/governance (ADR audit fans out). P6 runs the DONE gate. G1–G4 (build/vet/test/api-lint/cilint) rerun between phases; red gate stops the next phase.

**Tech stack:** Go 1.22+, PostgreSQL 16 (RLS, pg_trgm), OpenTelemetry Go SDK (`otelhttp`, `autoexport`), River, oapi-codegen, gitleaks v8.24.3+, sqlmock, Docker Compose.

**The anti-circle rule applies to every task:** discovery that isn't a Wave-Z-caused regression → one line in `wiki/backend/post-v1-backlog.md`, move on. No task grows.

---

## Phase P0 — Preflight (orchestrator inline, no fan-out)

### Task P0: Baseline + scaffolding

**Files:**
- Create: `wiki/backend/post-v1-backlog.md`

- [ ] **Step 1: Baseline gates.** Run, in order, and record outputs:
```
go build ./...                                                  # expect exit 0
go vet ./...                                                    # expect exit 0
go test -p 2 ./...                                              # expect exit 0, 0 FAIL
go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .  # expect "0 violation(s)"
go run ./tools/cilint/... ./...                                 # expect exit 0
```
If ANY is red, STOP — the tree drifted since Wave F; report to user.

- [ ] **Step 2: Docker stack up.** `docker ps` must show postgres+redis+minio+gotenberg healthy and docx-renderer Up. Start API per script-truth when runtime proofs needed: `.\scripts\start-api.ps1 -Build` (NOT with `METALDOCS_E2E=1` except where a task says so).

- [ ] **Step 3: Refresh GitNexus** (background, don't block): `node .gitnexus/run.cjs analyze`.

- [ ] **Step 4: Create the backlog file** `wiki/backend/post-v1-backlog.md`:
```markdown
# Post-v1 Backlog (Wave Z anti-circle parking lot)

> **Last verified:** <date>
> **Scope:** Pre-existing findings discovered during Wave Z that are NOT Wave-Z-caused regressions. Parked here by the anti-circle rule (spec §1). None block backend DONE. Triage post-v1.

| # | Found during | One-line description | Pointer |
|---|---|---|---|
```

- [ ] **Step 5: Re-run the RLS census** (Z-2 input; spec list may have drifted):
```
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -t -c "SELECT table_schema||'.'||table_name FROM information_schema.columns WHERE column_name='tenant_id' AND table_schema IN ('public','metaldocs') ORDER BY 1;"
```
Save the list; subtract `public.controlled_documents` + `metaldocs.audit_events` (already covered by 0234).

- [ ] **Step 6: Commit scaffolding.** `git add wiki/backend/post-v1-backlog.md && git commit -m "docs(wave-z): P0 scaffolding — post-v1 backlog parking lot"`
  (NOTE all session: `.gitnexus/` breaks `git add -A` — always stage explicit paths.)

---

## Phase P1 — Parallel small code items (fan-out ≤15; groups below NEVER share files)

Dispatch one sonnet agent per GROUP (8 groups). Each agent: read the task, read the anchor files, implement, run the listed verify commands, commit with the listed message. A haiku agent then re-runs each group's verify command and confirms the commit exists.

### Task Z-4 [group: audit]: Audit writer hard-required

**Files:**
- Modify: `internal/modules/audit/application/service.go` (WithExports :58; ExportEvents :179-196)
- Modify: `internal/modules/audit/application/service_test.go` (whatever constructs WithExports with nil writer)

- [ ] **Step 1:** In `WithExports`, panic on nil deps (pattern = taxonomy module.go fail-loud):
```go
func (s *Service) WithExports(counter domain.Counter, repo domain.ExportJobRepository, writer domain.Writer, urlBuilder SignedURLBuilder) *Service {
	if counter == nil || repo == nil || writer == nil || urlBuilder == nil {
		panic("audit.WithExports: all export dependencies are required")
	}
	...
}
```
- [ ] **Step 2:** In `ExportEvents`, delete the `if s.writer != nil { //cilint:allow-dualmode` guard + its comment block (:178-182). Replace with a plain unconditional emit. Keep the Wave-F `slog.Warn` on Record error. If `ExportEvents` is reachable when exports were never wired (check: grep the handler registration — `grep -rn "ExportEvents\|exportAuditEvents" internal/modules/audit/delivery/ apps/api/`), add at the top of `ExportEvents`: `if s.writer == nil { return nil, ErrExportsDisabled }` with `ErrExportsDisabled = errors.New("audit: export pipeline not configured")` — that is a feature-gate precondition (nodualmode-exempt shape `if x == nil { return }`), not a dual path.
- [ ] **Step 3:** Fix/extend tests: any test calling WithExports with nils now must pass fakes; add one test asserting the panic.
- [ ] **Step 4: Verify:** `go test ./internal/modules/audit/... && go run ./tools/cilint/... ./internal/modules/audit/...` → green, AND `git grep -c "allow-dualmode" -- internal/modules/audit/` → 0.
- [ ] **Step 5: Commit:** `fix(wave-z): Z-4 audit export writer hard-required — allow-dualmode retired (T-012)`

### Task Z-6 [group: iam-app]: Membership governance in-tx

**Files:**
- Modify: `internal/modules/iam/application/area_membership_service.go` (:48-160 — Grant/Revoke + `MembershipGovernanceLogger`)
- Modify: `apps/api/cmd/metaldocs-api/main.go` (the `membershipGovernanceLogger` adapter from commit `ab949015`)
- Modify: iam repo implementing `UserAreaWriteRepository` (find: `grep -rn "UserAreaWriteRepository" internal/modules/iam/`)
- Test: `internal/modules/iam/application/area_membership_service_test.go`

- [ ] **Step 1:** Read the Wave 2.2 pattern first: `git show 6cc9595 --stat` and one in-tx call site (`internal/modules/taxonomy/application/family_service.go`, `LogTx`).
- [ ] **Step 2:** Extend the port: `MembershipGovernanceLogger` gains `LogTx(ctx context.Context, tx db.Tx, action string, m AreaMembership) error`. The main.go adapter implements it by delegating to `auditWriter.RecordTx` (the audit writer already has RecordTx since Wave 2.2).
- [ ] **Step 3:** Move the two best-effort post-commit `s.logger.Log(...)` calls (grant :136, revoke sibling) INSIDE the repository transaction: the repo's Grant/Revoke must accept a callback or expose tx — follow whichever shape the repo already uses for tx (read it; Wave 2.2 used `LogTx(tx)` inside the mutation tx before Commit). Delete the two `slog.WarnContext(... best-effort ...)` branches — failure now rolls back the mutation (that IS the T-007 fix).
- [ ] **Step 4:** Test: rollback atomicity — sqlmock: governance insert error → expect Grant returns error AND ExpectRollback met. Pattern: copy the Wave 2.2 atomicity test (`git grep -l "rollback" internal/modules/taxonomy/application/*_test.go`).
- [ ] **Step 5: Verify:** `go test ./internal/modules/iam/... ./apps/api/... && go run ./tools/cilint/... ./...` green.
- [ ] **Step 6: Commit:** `fix(wave-z): Z-6 membership governance in-tx via LogTx — closes T-007 (REQ-ASYNC-1)`

### Task Z-7 [group: taxonomy]: TemplateVersionPort

**Files:**
- Create: `internal/modules/templates/domain/template_version_port.go`
- Create: `internal/modules/templates/infrastructure/template_version_reader.go` (move SQL verbatim from taxonomy)
- Modify: `internal/modules/taxonomy/infrastructure/template_version_checker.go` (delete; or thin adapter consuming the port)
- Modify: `apps/api/cmd/metaldocs-api/main.go` (wiring) + `internal/modules/controlleddocuments/module.go` if it constructs the checker (it does: `tplCheck`)

- [ ] **Step 1:** Read `internal/modules/taxonomy/infrastructure/template_version_checker.go:30-48`. Port interface mirrors its single method exactly:
```go
// templates/domain/template_version_port.go
package domain
type TemplateVersionPort interface {
	VersionExists(ctx context.Context, tenantID, versionID string) (bool, error) // match the real signature found in step 1
}
```
- [ ] **Step 2:** Move the SQL into `templates/infrastructure/template_version_reader.go` implementing the port (SQL byte-identical — this is a move, not a rewrite).
- [ ] **Step 3:** Consumers (taxonomy/CD wiring) accept the port; delete the taxonomy-side checker file. Wire the templates impl at composition root.
- [ ] **Step 4: Verify:** `git grep -n "templates_template\|template_versions" -- internal/modules/taxonomy/` → 0 hits; `go test ./internal/modules/taxonomy/... ./internal/modules/templates/... ./internal/modules/controlleddocuments/...` green.
- [ ] **Step 5: Commit:** `refactor(wave-z): Z-7 taxonomy reads template versions via templates/domain port (F-06e, REQ-TOP-1)`

### Task Z-8 + Z-9 [group: security+wiring]: security read port · CD repo wiring

**Files (Z-8):**
- Read: `internal/modules/security/infrastructure/postgres/repository.go:84-100`
- Either create `internal/modules/iam/domain/` narrow port + impl (LoginContextPort pattern, commit `07f914e9`) OR modify `wiki/modules/security.md` with an accepted-boundary note.

- [ ] **Step 1 (Z-8):** Read the JOIN. Decision rule: if the JOIN exists to RESOLVE tenant membership (the tenancy mechanism itself per ADR 0027 / T-008 by-design), write the accepted-boundary note in the security module doc (cite ADR 0027 + F-06 register entry) and DO NOT refactor. If it reads other iam_users columns (display fields etc.), extract a port exactly like `LoginContextPort` (`git show 07f914e9 --stat` is the template). Document the choice taken in the commit message.
- [ ] **Step 2 (Z-9):** `grep -n "PostgresControlledDocumentRepository" apps/api/cmd/metaldocs-api/main.go` (around :357 historically). Expose the repo on the CD module's `Dependencies`/return struct (`internal/modules/controlleddocuments/module.go`) and let main.go take it from the module instead of constructing a second instance.
- [ ] **Step 3: Verify:** `go build ./... && go test ./internal/modules/security/... ./internal/modules/controlleddocuments/... ./apps/api/...` green; `grep -c "infrastructure.NewPostgresControlledDocumentRepository" apps/api/cmd/metaldocs-api/main.go` → 0.
- [ ] **Step 4: Commit:** `refactor(wave-z): Z-8 security boundary disposition + Z-9 CD repo exposed via module (F-06 residuals)`

### Task Z-12 + Z-13 [group: config+documents-domain]: ParseBoolEnv · DocumentStatus enum

**Files:**
- Modify: `internal/platform/config/attachments.go:108` (parseBoolEnv → exported), `internal/platform/authn/config.go:213` (use it)
- Modify: `internal/modules/documents/domain/model.go:8-13` (+6 constants), documents-module callsites

- [ ] **Step 1 (Z-12):** Export `func ParseBoolEnv(name string, def bool) bool` in `platform/config` (keep the richer 4-value semantics of the two copies — read both first, take the superset: "1,true,yes,on" / "0,false,no,off", else default). Delete the authn private copy; authn imports `platform/config`. KEEP `splitCSV` copies untouched.
- [ ] **Step 2 (Z-13):** Add to `documents/domain/model.go`:
```go
const (
	StatusApproved   DocumentStatus = "approved"
	StatusPublished  DocumentStatus = "published"
	StatusSuperseded DocumentStatus = "superseded"
	StatusObsolete   DocumentStatus = "obsolete"
	StatusScheduled  DocumentStatus = "scheduled"
	StatusRejected   DocumentStatus = "rejected"
)
```
(match the existing 3 constants' naming style exactly — read the file first). Then convert documents-module raw literals: `git grep -n '"approved"\|"published"\|"superseded"\|"obsolete"\|"scheduled"\|"rejected"' -- internal/modules/documents/ | grep -v _test` — convert each Go callsite that means a DocumentStatus (NOT SQL string literals inside queries — leave SQL alone).
- [ ] **Step 3: Verify:** `go build ./... && go test ./internal/platform/config/... ./internal/platform/authn/... ./internal/modules/documents/...` green.
- [ ] **Step 4: Commit:** `refactor(wave-z): Z-12 ParseBoolEnv consolidation (F-15) + Z-13 DocumentStatus enum completion (D-07)`

### Task Z-14 + Z-15 [group: cd-http+idempotency]: typed codes + guard · TTL const

**Files:**
- Modify: `internal/modules/controlleddocuments/delivery/http/handler.go:54,106,109`
- Modify: `internal/platform/problem/codes.go` (only if a const is missing)
- Modify: `internal/platform/problem/codes_catalog_guard_test.go` (guardedPackages)
- Modify: `internal/platform/httpresponse/` WriteError signature IF it takes string (check first)
- Modify: `internal/platform/idempotency/postgres_store.go:91,229`

- [ ] **Step 1 (Z-14):** Check `httpresponse.WriteError` signature. If `code string` → widen to `code problem.Code` (Wave 1.4 precedent `ed7890597`); update ALL callers (`git grep -n "httpresponse.WriteError(" -- '*.go'` — mechanical sweep, every literal becomes its catalog const; any literal with NO catalog const gets one added to codes.go + FE regen via `go run ./scripts/dump-error-codes.go` — check the FE coverage test path in that script's header).
- [ ] **Step 2:** Enroll `internal/modules/controlleddocuments/delivery/http` in `guardedPackages` in codes_catalog_guard_test.go.
- [ ] **Step 3 (Z-15):** In postgres_store.go, both `'24 hours'` SQL literals → one `const idempotencyTTL = "24 hours"` interpolated (or `interval` param) — read both queries, keep behavior identical.
- [ ] **Step 4: Verify:** `go test ./internal/platform/problem/... ./internal/platform/idempotency/... ./internal/modules/controlleddocuments/... ./tests/unit/...` green; FE coverage test if codes were added (`cd frontend/apps/web && npm test -- --run error` or per the catalog script's documented check).
- [ ] **Step 5: Commit:** `fix(wave-z): Z-14 CD handler typed problem codes + guard enrollment, Z-15 idempotency TTL const (F-09 family)`

### Task Z-21 [group: slog-sweep]: log.Printf → slog (56 sites / 12 files + objectstore)

**Files (all from the F-02 census):**
`internal/modules/auth/delivery/http/handler.go`, `.../middleware.go`, `internal/modules/documents/approval/http/doc_approval_handler.go`, `.../signoff_handler.go`, `internal/modules/documents/delivery/http/handler.go`, `internal/modules/documents/jobs/orphan_pending_sweeper.go`, `.../session_sweeper.go`, `internal/modules/iam/application/dev_role_provider.go`, `internal/modules/iam/delivery/http/admin_handler.go`, `.../people_handler.go`, `.../routes_memberships.go`, `internal/modules/templates/application/lifecycle.go`, `internal/platform/objectstore/document_presigner.go`

- [ ] **Step 1:** Mechanical rule per site: `log.Printf("msg %v", err)` → `slog.Warn("msg", "err", err)` (Error for error-paths returning 5xx, Warn for degraded paths, Info for lifecycle notices — judge from the message text). Message string keeps its identifying prefix; format-verb args become structured keys (err, id, user_id…). NO new context plumbing — use the package-level slog (these handlers run under the chain's trace-aware slog handler already).
- [ ] **Step 2:** Remove now-unused `"log"` imports.
- [ ] **Step 3: Verify:** `git grep -n "log\.Printf" -- internal/ apps/ ':!*_test.go'` → **0** ; `go build ./... && go test ./internal/modules/auth/... ./internal/modules/documents/... ./internal/modules/iam/... ./internal/modules/templates/... ./internal/platform/objectstore/...` green.
- [ ] **Step 4: Commit:** `refactor(wave-z): Z-21 log.Printf -> slog across 13 files (F-02, REQ-OBS-1)`

### Task Z-23 + Z-25 + Z-17 + Z-18 [group: small-ops]: readiness errgroup · subject_code drop · partials · deprecated flag

- [ ] **Step 1 (Z-23):** `internal/platform/observability/runtime.go:286-323` — sequential dependency checks → `golang.org/x/errgroup` (already in go.mod? `grep errgroup go.mod`; if absent use plain goroutines+WaitGroup) with one shared `context.WithTimeout(ctx, 5*time.Second)`; collect each check's result into the same response struct. Keep response shape byte-identical. Test: existing runtime tests + add one asserting two slow fake checks complete within single budget.
- [ ] **Step 2 (Z-25):** Verify zero refs: `git grep -n "subject_code" -- internal/ apps/ db/ ':!db/migrations/0236*'` and `docker exec metaldocs-postgres psql ... -c "SELECT count(*) FROM information_schema.columns WHERE column_name='subject_code'"`. Migration `db/migrations/0238_drop_documents_subject_code.sql`:
```sql
-- 0238: drop orphan metaldocs.documents.subject_code (+index) — FK was CASCADE-dropped by 0236 (CD T-010).
DROP INDEX IF EXISTS metaldocs.idx_documents_subject_code;
ALTER TABLE metaldocs.documents DROP COLUMN IF EXISTS subject_code;
```
(verify real index name first: `\di metaldocs.*subject*`). Apply live, confirm ledgered in `public.schema_migrations`, retire dictionary column row.
- [ ] **Step 3 (Z-17):** Prove partials dead: `git grep -n "partials" -- '*.yaml' '*.yml' '*.go' '*.json' '*.ps1' scripts/ .github/ api/`. If only self-references → `git rm api/openapi/v1/partials/*.yaml`; update `wiki/architecture/backend-api-structure.md` if it mentions partials. If consumed → snake_case the camelCase props instead and record which consumer.
- [ ] **Step 4 (Z-18):** `api/openapi/v1/openapi.yaml` op `createManagedUser` (~:215): add `deprecated: true` under the operation. Regen per Task Z-16's codegen procedure (or fold into Z-16's regen if same session — coordinate via orchestrator; THIS task may leave regen to Z-16 and only edit the spec, noting it).
- [ ] **Step 5: Verify:** api-lint 0 · `go build ./...` · runtime `/readyz` smoke after Z-23 (`Invoke-WebRequest http://localhost:8081/readyz` — find the actual readiness path first: `git grep -n "readyz\|readiness" apps/api/`) · migration ledgered.
- [ ] **Step 6: Commit** (one per sub-item): `fix(wave-z): Z-23 concurrent readiness checks (F-16C)` · `chore(wave-z): Z-25 drop orphan subject_code (0238)` · `chore(wave-z): Z-17 partials disposition (F-13c)` · `docs(wave-z): Z-18 deprecated:true on createManagedUser (F-13d)`

### Task Z-11 [group: bootstrap]: MinIO clients 3→2

**Files:**
- Modify: `internal/platform/bootstrap/api.go:85-103,158-178` (buildMinioClients + miniostore wiring)
- Modify: the `miniostore.NewStore` constructor (find: `git grep -n "func NewStore" -- internal/ | grep -i minio`)

- [ ] **Step 1:** Change `miniostore.NewStore` to accept `*minio.Client` (drop its internal dial); bootstrap passes the existing internal client from `buildMinioClients`.
- [ ] **Step 2: Verify:** `go build ./... && go test ./internal/platform/...` green. Runtime: restart API, login, one presign smoke (autosave presign route or template presign — any 200 with a signed URL).
- [ ] **Step 3: Commit:** `refactor(wave-z): Z-11 single MinIO byte-IO client reused from bootstrap (D-02)`

---

## Phase P2 — The big two (sequential, orchestrator supervises, runtime-proofed)

### Task Z-2 + Z-3: Full-table RLS (migration 0237)

**Files:**
- Create: `db/migrations/0237_rls_all_tenant_tables.sql`
- Modify: `wiki/decisions/0027-rls-adoption-sequencing.md` (status → executed-in-full)
- Modify: `wiki/database/` dictionary pages happen in Z-33 sweep

- [ ] **Step 1:** Use the P0 census (≈28 tables). Read `db/migrations/0234_rls_controlled_documents_audit_events.sql` fully — 0237 extends its exact pattern.
- [ ] **Step 2:** Generate DDL with ONE uniform NULL-permissive policy form (column cast to text handles both uuid and text tenant_id):
```sql
-- For EACH table from the census (example: metaldocs.iam_users):
ALTER TABLE metaldocs.iam_users ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.iam_users FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.iam_users;
CREATE POLICY tenant_isolation ON metaldocs.iam_users
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id::text = NULLIF(current_setting('metaldocs.tenant_id', true), '')
  );
```
Generator query to emit all blocks (run, paste output into the migration, review by eye):
```sql
SELECT format(E'ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY;\nALTER TABLE %I.%I FORCE ROW LEVEL SECURITY;\nDROP POLICY IF EXISTS tenant_isolation ON %I.%I;\nCREATE POLICY tenant_isolation ON %I.%I USING (NULLIF(current_setting(''metaldocs.tenant_id'', true), '''') IS NULL OR tenant_id::text = NULLIF(current_setting(''metaldocs.tenant_id'', true), ''''));\n', table_schema,table_name,table_schema,table_name,table_schema,table_name,table_schema,table_name)
FROM information_schema.columns
WHERE column_name='tenant_id' AND table_schema IN ('public','metaldocs')
  AND NOT (table_schema='public' AND table_name='controlled_documents')
  AND NOT (table_schema='metaldocs' AND table_name='audit_events')
ORDER BY 1;
```
Header comment: cite ADR 0027 (sequencing now EXECUTED IN FULL, operator override of D-3 per Wave Z spec), F-12 tail, RF-6, REQ-TEN-1; repeat 0234's superuser/NOSUPERUSER caveat verbatim.
- [ ] **Step 3 (Z-3):** Append to the same migration the idempotency FK (verify column/table first: `\d metaldocs.idempotency_keys`):
```sql
ALTER TABLE metaldocs.idempotency_keys
  ADD CONSTRAINT fk_idempotency_keys_tenant
  FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE;
```
(Confirm the tenants PK table name via `\dt metaldocs.*tenant*` — if rows exist with orphan tenant_ids, the ALTER fails → triage: fix data or use NOT VALID + VALIDATE.)
- [ ] **Step 4:** Apply live (API restart applies migrations — `.\scripts\start-api.ps1`), confirm `0237` in `public.schema_migrations`.
- [ ] **Step 5: Full test suite** `go test -p 2 ./...` — RLS must not break any test (NULL-permissive + dev superuser ⇒ identical behavior expected). Any failure = real GUC-dependence found → fix the test or the query path (that IS in scope: Wave-Z-caused regression).
- [ ] **Step 6: Runtime proof (NOSUPERUSER probe, Wave-F method):** create `rls_z_probe NOSUPERUSER NOBYPASSRLS`, GRANT SELECT on `metaldocs.iam_users` + `public.documents`, seed/identify rows for 2 tenants, prove GUC-unset→all / GUC=A→only-A / GUC=B→only-B on BOTH tables, then REVOKE + DROP role.
- [ ] **Step 7: System-path proof:** worker + jobs containers still process (re-run the Wave-F worker relay probe: insert `pdf_dispatch_outbox` row → dispatched + `outbox_events` relay; delete probes).
- [ ] **Step 8:** Amend ADR 0027: status `Accepted (executed in full by Wave Z, 2026-06-13)`, one-paragraph note.
- [ ] **Step 9: Commit:** `feat(wave-z): Z-2/Z-3 RLS on all 28 tenant tables + idempotency tenant FK (F-12 tail, RF-6, REQ-TEN-1, ADR 0027 executed)`

### Task Z-1: Minimal OTel

**Files:**
- Create: `internal/platform/observability/otel.go` (+ `otel_test.go`)
- Modify: `apps/api/cmd/metaldocs-api/chain.go` + `chain_test.go` (new link)
- Modify: `apps/api/cmd/metaldocs-api/main.go` (Setup call + shutdown hook)
- Modify: `internal/platform/observability/http.go` (`normalizeRoute` → `r.Pattern`; trace-id bridge)
- Modify: `go.mod` (otel deps)

- [ ] **Step 1 (MANDATORY consultation):** Look up CURRENT otel-go API before writing any code — context7/web: `go.opentelemetry.io/otel`, `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`, `go.opentelemetry.io/contrib/exporters/autoexport` latest stable versions + the current `autoexport.NewSpanExporter` signature + `sdktrace.NewTracerProvider` idiom. Do not write from memory.
- [ ] **Step 2:** `internal/platform/observability/otel.go` — env-gated setup, INERT by default:
```go
// SetupOTel installs a real tracer provider ONLY when the operator configured
// an exporter (OTEL_TRACES_EXPORTER or OTEL_EXPORTER_OTLP_ENDPOINT set).
// Unconfigured ⇒ no SDK install, global tracer stays no-op, zero overhead —
// REQUIRED behavior per Wave Z spec Z-1 (single-host default deployment).
func SetupOTel(ctx context.Context) (shutdown func(context.Context) error, enabled bool, err error) {
	if os.Getenv("OTEL_TRACES_EXPORTER") == "" && os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return func(context.Context) error { return nil }, false, nil
	}
	exp, err := autoexport.NewSpanExporter(ctx)
	if err != nil { return nil, false, err }
	res, _ := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("metaldocs-api")))
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp.Shutdown, true, nil
}
```
(adjust names to the API found in Step 1).
- [ ] **Step 3:** Chain: add link `{"otel", otelWrap}` directly after `panic_recovery` (recovery must still be outermost so a panicked span doesn't kill the process). `otelWrap` = `func(next http.Handler) http.Handler { return otelhttp.NewHandler(next, "http", otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return r.Method + " " + routePattern(r) })) }` — pass nil when OTel disabled (buildChain skips nil links already). Update `apiChain` signature + BOTH canon lists in `chain_test.go`.
- [ ] **Step 4 (bridge):** in `observability/http.go`, where the trace id is resolved for logs/metrics: if `trace.SpanContextFromContext(r.Context()).IsValid()`, use its TraceID hex as `trace_id` (falls back to existing requesttrace when OTel off). Keep `X-Trace-Id` inbound acceptance + response header behavior unchanged.
- [ ] **Step 5 (normalizeRoute):** replace the 5-pattern switch with `r.Pattern` (Go 1.22 ServeMux sets it); fall back to the old normalize for empty Pattern. Cardinality fix per F-17 secondary defect.
- [ ] **Step 6 (outbox propagation):** producer writes the ACTIVE trace id into `outbox_events.trace_id` (it already writes requesttrace's id — with Step 4's bridge this becomes the OTel trace id automatically; verify by reading `platform/messaging/outbox/postgres/publisher.go` and the trace-id source it uses; adjust to the bridged resolver if it reads the old one directly).
- [ ] **Step 7: Tests:** `otel_test.go`: (a) no env → enabled=false, shutdown no-ops; (b) `OTEL_TRACES_EXPORTER=console` → enabled=true. Chain test updated for the new link. `go test ./internal/platform/observability/... ./apps/api/...`.
- [ ] **Step 8: Runtime proof:** (a) `.\scripts\start-api.ps1 -Build` WITHOUT otel env → boot log identical, smoke 200, NO otel noise; (b) restart with `OTEL_TRACES_EXPORTER=console` → make 1 authed request → span JSON visible in `logs/api.log` with `http.route`-style pattern + W3C `traceparent` accepted when sent. Record both outputs.
- [ ] **Step 9: Commit:** `feat(wave-z): Z-1 minimal OTel — otelhttp + W3C traceparent + autoexport (env-gated, inert by default) (F-17, RF-1, REQ-OBS-1/2/3)`

---

## Phase P3 — Contract/codegen (SERIALIZED — regen conflicts; one agent at a time; metaldocs-backend-api skill governs)

### Task Z-16: roles enum regen

- [ ] **Step 1:** Locate the enum in spec: `grep -n "CreateManagedUserRequest" api/openapi/v1/openapi.yaml` → add `signer`, `area_admin`, `qms_admin` to the roles enum (canonical 8-role list: `git grep -n "qms_admin\|area_admin" internal/modules/iam/domain/model.go` confirms names).
- [ ] **Step 2:** Find the codegen entrypoint: `git grep -ln "oapi-codegen" -- scripts/ Makefile '*.ps1' '*.yaml'` → run it for the iam package; then FE types regen (`npm run gen:api` in `frontend/apps/web` — verify script name in its package.json).
- [ ] **Step 3: Verify:** codegen idempotent (re-run → empty diff), `go build ./...`, api-lint 0, FE `tsc` clean (`npm run typecheck` or equivalent).
- [ ] **Step 4: Commit:** `fix(wave-z): Z-16 CreateManagedUserRequest roles enum completes 8-role set + regen (F-11 residual)`

### Task Z-19: iamapi codegen split

- [ ] **Step 1:** Read `internal/modules/iam/api/cfg.yaml` (+ how `api.gen.go` is produced — include filters/tags). Create `internal/modules/audit/api/cfg.yaml` + `internal/modules/security/api/cfg.yaml` with tag/path filters selecting only their ops (`audit`-tagged, `security`-tagged — confirm tags exist in spec; if filtering is by path globs use those).
- [ ] **Step 2:** Generate three packages; move/keep package names (`iamapi`→ slimmer, new `auditapi`, `securityapi`); update imports (`git grep -ln "iam/api\"" internal/ apps/`) so audit/security delivery code imports its own package.
- [ ] **Step 3: Verify:** `go build ./...` green, regen idempotent, generated-only diff except import lines, api-lint 0, full module tests green.
- [ ] **Step 4: Commit:** `refactor(wave-z): Z-19 split iam/audit/security codegen packages (F-13e)`
- [ ] **Hard-stop note:** if filters can't cleanly partition shared schemas → STOP this task, park in post-v1-backlog with the blocking shape, continue wave (spec §4 allows).

### Task Z-24: CD trigram index

- [ ] **Step 1:** `docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -c "SELECT * FROM pg_available_extensions WHERE name='pg_trgm';"` — must be available (stock postgres:16 image has it).
- [ ] **Step 2:** Migration `db/migrations/0239_cd_trgm_search.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_controlled_documents_title_trgm
  ON public.controlled_documents USING GIN (title gin_trgm_ops);
```
(verify the actual searched column(s) at `internal/modules/controlleddocuments/infrastructure/repository.go:128` TODO(phase11) — index exactly what the ILIKE touches; multiple columns → one GIN each).
- [ ] **Step 3:** The ILIKE query itself stays (trigram GIN accelerates `ILIKE '%x%'` natively); remove the TODO comment, reference the index.
- [ ] **Step 4: Verify:** migration ledgered; `EXPLAIN` on the search SQL shows Bitmap Index Scan (seed 3 CD rows first if table empty, then clean up); CD tests green.
- [ ] **Step 5: Commit:** `perf(wave-z): Z-24 pg_trgm GIN index for CD search ILIKE (F-20b)`

### Task Z-26: gitleaks full-history

**Files:** `.github/workflows/secret-scan.yml`, `.gitleaks.toml`

- [ ] **Step 1:** Local full-history run to surface the triage set:
  `docker run --rm -v "${PWD}:/repo" ghcr.io/gitleaks/gitleaks:v8.24.3 detect --source /repo --redact --exit-code 1` (git mode = default without `--no-git`). Capture the findings list (~15 expected: test fixtures/dummy tokens).
- [ ] **Step 2:** Triage EVERY finding: each gets either a `[allowlist]` regex/path entry WITH a rationale comment, or — if anything real surfaces — STOP, report to user (secret in history = user decision, D-4b context). Remember v8.24.3 ignores the plural `[[allowlists]]` form — stay singular; if rule-scoped allowlisting is needed, bump the pinned image to the latest gitleaks (check release notes via web first) and re-verify the existing allowlist still passes.
- [ ] **Step 3:** Workflow: add `fetch-depth: 0` to checkout, drop `--no-git`, update the explanatory comment.
- [ ] **Step 4: Verify:** local full-history run with final config → exit 0; `python -c "import yaml,sys; yaml.safe_load(open('.github/workflows/secret-scan.yml'))"` parses.
- [ ] **Step 5: Commit:** `feat(wave-z): Z-26 gitleaks full-history scanning + triaged allowlist (1.11, F-18 round-2)`

---

## Phase P4 — Structural M-items (3 focused sonnet agents, parallel — disjoint files)

### Task Z-5: freeze_service in-tx-only

**Files:** `internal/modules/documents/application/freeze_service.go` (:178, :314, :379) + its tests + ADR `wiki/decisions/0015-async-freeze-pin-materialize.md`

- [ ] **Step 1:** Read all three `if tx != nil { //cilint:allow-dualmode` blocks + their callers (`git grep -n "\.Freeze(\|\.Pin(\|\.Materialize(" -- internal/ apps/ ':!*_test.go'`).
- [ ] **Step 2:** Collapse: each method requires tx (`if tx == nil { return fmt.Errorf("freeze_service: tx required (ADR 0015 amended by Wave Z)") }`); the else-branch (autocommit path) is deleted. Callers currently passing nil must open the tx themselves — wire each one (expected few; Wave 2.13 already noted both branches write identically apart from enlistment).
- [ ] **Step 3:** Delete the 3 `//cilint:allow-dualmode` directives. Tests updated to always pass tx (sqlmock ExpectBegin where needed).
- [ ] **Step 4:** Amend ADR 0015: `Status: Accepted (amended 2026-06-13 — optional-tx-enlistment retired by Wave Z Z-5; tx is mandatory)` + 3-line note.
- [ ] **Step 5: Verify:** `git grep -c "allow-dualmode" -- internal/` → counts ONLY non-freeze ones remaining (after Z-4+Z-5: target 0); `go test ./internal/modules/documents/... && go run ./tools/cilint/... ./...` green; live Pin/Materialize probe (Wave-F worker relay probe again).
- [ ] **Step 6: Commit:** `refactor(wave-z): Z-5 freeze service tx-mandatory — ADR 0015 amended, allow-dualmode retired (T-013)`

### Task Z-10: outbox generics + dead loop + idemp-store dedup

**Files:**
- Create: `internal/modules/render/fanout/staging_outbox.go` (generic repo+worker)
- Modify→delete-or-thin: `pdf_outbox_repository.go`, `materialize_outbox_repository.go`, `pdf_outbox_worker.go`, `materialize_outbox_worker.go` (+tests)
- Modify: `apps/api/cmd/metaldocs-api/main.go:462-486` (startOutboxWorker dead restart loop)
- Modify: `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go` + `postgres_route_admin_idemp_store.go` (shared envelope helper)

- [ ] **Step 1:** Generic core (rows are structurally identical — OutboxRow == MaterializeOutboxRow):
```go
type StagingOutboxRepository struct{ db *sql.DB; table string } // table fully-qualified, validated against an allowlist {"metaldocs.pdf_dispatch_outbox","metaldocs.materialize_dispatch_outbox"}
```
All six methods move in with `fmt.Sprintf` on the validated table name (sqlescape if the repo has an identifier helper — check `internal/platform/sqlescape`). `PDFOutboxRepository`/`MaterializeOutboxRepository` become thin aliases or are replaced at the call sites — prefer REPLACING call sites and deleting the old types (grep consumers; Wave-F nil-tx contract + tests carry over verbatim).
- [ ] **Step 2:** One generic worker parameterized by `(repo, eventType messaging.EventType, idempPrefix string)` — diff the two workers first (`git diff --no-index internal/modules/render/fanout/pdf_outbox_worker.go internal/modules/render/fanout/materialize_outbox_worker.go`) — parameterize exactly the differing tokens.
- [ ] **Step 3:** Delete the dead restart loop in `startOutboxWorker` (main.go:462-486) — `Run()` never returns non-nil (F-04 evidence); replace with a single invocation + comment.
- [ ] **Step 4:** Approval idemp stores: extract the duplicated ReplayHandle/JSON-envelope helpers into one shared private file in the same package; both stores call it (95% duplicate per register).
- [ ] **Step 5: Verify:** `go test ./internal/modules/render/... ./internal/modules/documents/approval/... ./apps/api/...` green (Wave-F NilTxRejected + Idempotent tests must still pass against the generic) · live worker relay probe (insert pdf_dispatch row → dispatched + outbox_events; cleanup).
- [ ] **Step 6: Commit:** `refactor(wave-z): Z-10 generic staging outbox repo/worker, dead restart loop deleted, idemp-store envelope dedup (F-04)`

### Task Z-22: WS presence drain

**Files:** locate hub: `git grep -ln "websocket\|Upgrader\|presence" -- internal/modules/ internal/platform/ | grep -iv test` (presence/collaboration module) + `apps/api/cmd/metaldocs-api/main.go` (server shutdown sequence)

- [ ] **Step 1:** Read the hub + the existing `server.Shutdown` call in main.go (graceful path from F-16/1.2).
- [ ] **Step 2:** Hub tracks live conns (mutex map add-on-upgrade/remove-on-close — if a registry already exists, reuse). Add `hub.CloseAll(ctx)`: send close frame (`websocket.CloseGoingAway`) to each, then close.
- [ ] **Step 3:** `server.RegisterOnShutdown(func(){ hub.CloseAll(shutdownCtx) })` in main.go.
- [ ] **Step 4: Test:** unit — register fake conn, call CloseAll, assert close frame + removal. `go test ./...` for the touched module.
- [ ] **Step 5: Commit:** `feat(wave-z): Z-22 WS presence drain on shutdown (F-16B, RF-9, REQ-REL-2)`

---

## Phase P5 — Docs + governance (fan-out ≤15: 5 ADR agents + 1 each Z-28/29/30/31/32)

### Task Z-27: ADR lifecycle audit (5 sonnet agents × ~5 ADRs each + 1 index agent)

**Files:** all of `wiki/decisions/` (25 numbered + `2026-06-03-audit-events-cursor-shape.md` + `index.md` + `README.md`)

- [ ] **Step 1 (each ADR agent):** For each assigned ADR: read it; grep the code/wiki for its key claims; set the canonical header line (FIRST blockquote line):
  `> **Status:** Accepted` | `Accepted (amended YYYY-MM-DD — <what>)` | `Superseded by ADR-XXXX` | `Deprecated (<why>)` | `Historical (<context>)`.
  Where the body describes dead reality, add ONE line directly under the status: `> **Current reality (2026-06):** <what changed, pointer>`. NEVER rewrite the body (ADRs are immutable records).
  Known dispositions to apply: `0003` = stub → write the 10-line real record from `wiki/concepts/placeholders.md` + ADR 0008 context, or mark `Historical (stub; superseded by 0008's fixed catalog)` — whichever the evidence supports. `0012`/`0013` = add missing status lines (both Accepted; verify claims). `0015` = amended by Z-5 (P4 already wrote it — verify, don't duplicate). `0022` = Z-28 flips it (skip here). `0027` = Z-2 amended it (verify).
- [ ] **Step 2 (index agent, AFTER the 5):** Decide `2026-06-03-audit-events-cursor-shape.md`: it's CLOSED — rename to next free number (`0028-audit-events-cursor-shape.md`, update inbound links: `git grep -ln "2026-06-03-audit-events"`) with `Status: Accepted (closed 2026-06-08)`. Rebuild `index.md`:
```markdown
| # | Title | Status | Superseded by | Current relevance |
|---|-------|--------|---------------|-------------------|
```
one row per ADR. Update `README.md` with the status vocabulary + the rule "new ADRs MUST carry a Status header from this vocabulary".
- [ ] **Step 3: Verify:** `for f in wiki/decisions/00*.md; do grep -L "^> \*\*Status:\*\*" $f; done` → empty output; index row count == file count.
- [ ] **Step 4: Commit:** `docs(wave-z): Z-27 ADR lifecycle audit — canonical statuses, stub resolved, stray renumbered, index rebuilt`

### Task Z-28: ADR 0022 Phase 6 (wiki-curator dispatch)

- [ ] **Step 1:** Dispatch the `wiki-curator` agent (sonnet) with change context: "ADR 0022 phases 1–13 complete + Waves 1–2 authz changes (typed caps, no-rawstring-tier1-authz lint, scope binding, RolesByUserIDs batch, UserActiveInTenant) — sync `wiki/concepts/authz-tiers.md`, iam/auth module docs' authz sections, capability-catalog references; bump stamps."
- [ ] **Step 2:** Mark Phase 6 complete in ADR 0022's phase table; flip its status: `Accepted (fully executed 2026-06-13)`.
- [ ] **Step 3: Verify:** curator report lists files+stamps; ADR 0022 table has no open phase.
- [ ] **Step 4: Commit:** `docs(wave-z): Z-28 ADR 0022 Phase 6 wiki sync — program fully executed`

### Task Z-29: Cache invalidation contract (RF-3 + D-05)

**Files:** `internal/modules/iam/application/cached_role_provider.go` (+ wherever the Redis capability cache lives: `git grep -ln "redis" -- internal/modules/iam/ internal/platform/ | grep -iv test`)

- [ ] **Step 1 (verify invalidation paths FIRST — this is the real RF-3 risk):** `git grep -n "InvalidateUserTenant\|Invalidate" -- internal/modules/iam/` — then trace every role-mutating path (role grant/revoke, membership grant/revoke, user deactivate: grep their service methods) and confirm each calls the invalidator. **A mutation path that does NOT invalidate = fix in-wave** (add the call after commit; test it).
- [ ] **Step 2:** Write the contract doc block atop `CachedRoleProvider`:
```go
// CacheContract (REQ-CACHE-1, RF-3):
//   Source of truth: postgres RoleProvider. Cache-aside, per (user,tenant) key.
//   TTL: 30s default (ctor param). Staleness bound: max(TTL) — a grant/revoke
//   is visible no later than TTL even if invalidation is missed.
//   Invalidation: InvalidateUserTenant called by <list every caller found in step 1>.
//   Failure mode: cache errors fall through to source; never serves wrong-tenant data
//   (key embeds tenantID). Eviction: ticker sweep every TTL.
```
(fill the caller list with the REAL sites found). Mirror a matching block on the Redis capability cache.
- [ ] **Step 3: Live probe:** grant a role via API → immediately read the user's roles → revoke → read again; both reads reflect the change within the contract bound. Record requests+outputs.
- [ ] **Step 4:** Blueprint C4 → ✅ (cite this task) — leave the actual edit for Z-33 if conflict risk, else do it here.
- [ ] **Step 5: Commit:** `docs(wave-z): Z-29 cache contracts + invalidation path verification (RF-3, D-05, REQ-CACHE-1)`

### Task Z-30 + Z-31: flag lifecycle + messaging fence

- [ ] **Step 1 (Z-30):** Create `wiki/standards/feature-flag-lifecycle.md` (~30 lines): naming `ff_<area>_<purpose>`, every flag declares owner + ramp plan + cleanup date at introduction, dead-flag removal on next touch, current inventory (read `internal/platform/featureflags/handler.go` — list the real flags, incl. `MDDM_NATIVE_EXPORT_ROLLOUT_PCT`).
- [ ] **Step 2 (Z-31):** In `wiki/architecture/backend-blueprint.md` D8 + the messaging section of `wiki/backend/index.md`: document `platform/servicebus` as what it IS — the synchronous Gotenberg PDF HTTP adapter, NOT a broker; `platform/messaging` = outbox + noop only; any future broker = new ADR. Flip D7+D8 → ✅.
- [ ] **Step 3: Commit:** `docs(wave-z): Z-30 flag lifecycle standard + Z-31 messaging/servicebus fence (RF-7, RF-8)`

### Task Z-32: v1 release re-baseline runbook

**Files:** Create `wiki/runbooks/v1-release-rebaseline.md`

- [ ] **Step 1:** Write the runbook — exact commands, operator-executable cold:
```markdown
# v1 Release Re-baseline Runbook (D-4b — closes F-18 permanently)
1. Preconditions: Wave Z DONE gate green; working tree clean; final build smoke on :8081.
2. mkdir ../metaldocs-v1 && cd ../metaldocs-v1 && git init -b main
3. robocopy/cp the WORKING TREE of the old repo (exclude: .git, .gitnexus, bin/, logs/, node_modules — honor .gitignore: `git -C ../MetalDocs ls-files | <copy listed files>` is the precise method)
4. Verify NO secret: docker run gitleaks detect --no-git over the new tree → exit 0
5. git add -A && git commit -m "MetalDocs v1.0.0" && git tag v1.0.0
6. Create NEW GitHub repo (do NOT reuse the old one); push main + tag.
7. Old repo: archive on GitHub (Settings → Archive). Never force-push/delete — just freeze.
8. Re-point CI secrets/env on the new repo; clone fresh for all future work.
9. Post-check: new repo `git log` = 1 commit; gitleaks full-history on new repo = exit 0.
```
(flesh each step with the real commands; step 3 must use the `git ls-files` copy method verbatim so gitignored junk can't leak in).
- [ ] **Step 2: Commit:** `docs(wave-z): Z-32 v1 re-baseline runbook (D-4b, F-18 closure)`

---

## Phase P6 — DONE gate (orchestrator inline + ONE review workflow)

### Task Z-33 + G1–G11

- [ ] **Step 1:** Rerun G1–G4 fresh (build, vet, full `-p 2` suite, api-lint, cilint) — record outputs.
- [ ] **Step 2 (G4 extra):** `git grep -c "cilint:allow-dualmode" -- internal/` → **0**.
- [ ] **Step 3 (G9 runtime smoke):** restart stack clean (`.\scripts\start-api.ps1 -Build`, NO E2E env): login 200 · authed GET 200 · RLS NOSUPERUSER probe on iam_users + documents · OTel inert boot + console-exporter span proof · taxonomy family create → in-tx audit row → cleanup · panic probe (needs one E2E-enabled restart OR skip with the Wave-F evidence cited — prefer rerun) · 13-burst login → 429.
- [ ] **Step 4 (G10):** ONE review workflow (≤10 sonnet agents): slices = P1 groups + P2 + P3 + P4 diffs, prompt restricted verbatim to: *"Defects INTRODUCED by commits <Z-range> only. Pre-existing issues: report in one line for the post-v1 backlog, do NOT block."* Fix Wave-Z regressions by family (own commits); park the rest.
- [ ] **Step 5 (G5–G8, Z-33 docs):**
  - `legacy-register.md`: every Wave-F defer note → resolution note citing the Z-commit; Excluded-forever items (F-09a, F-20f, splitCSV, F-18-at-release) get a final `**Closed (Wave Z): KEEP/at-release**` line. Header: register CLOSED.
  - `backend-blueprint.md`: scoreboard ALL-✅ (B2 cite Z-28+Z-29; A3 cite Z-20/backlog closure; D2 cite Z-1; C4 cite Z-29; D7/D8 cite Z-30/31) + one at-release line for F-18. REQ table: REQ-OBS + REQ-CACHE flip to MET.
  - `wiki/backend/roadmap.md`: append frozen `## Wave Z addendum` table (Z-ID → commit → one-line evidence); Status header notes Wave Z complete.
  - `current-agent-handoff.md`: Wave Z closing block (gate outputs, commit list, backlog pointer, runbook pointer).
  - `post-v1-backlog.md`: final stamp.
- [ ] **Step 6 (Z-20 fold-in):** blueprint A3 row cites `wiki/backlog/api-contract-hardening.md` Phase-F closure (done 2026-06-08, 0 CRITICAL/0 HIGH, api-lint 0/0) — doc-only.
- [ ] **Step 7: Final commit:** `docs(wave-z): Z-33 DONE gate green — register closed, blueprint all-green, backend frozen for feature work` — then STOP. **Do NOT merge. Present the gate evidence to the user for sign-off + Sunday release via Z-32 runbook.**

---

## Orchestration cheat-sheet (for the fresh session)

| Phase | Mode | Agents | Max concurrent |
|---|---|---|---|
| P0 | inline | 0 | — |
| P1 | workflow fan-out | 8 sonnet groups + 4 haiku verifiers | ≤12 |
| P2 | inline-supervised, sequential | 1 sonnet helper each | ≤2 |
| P3 | serialized | 1 sonnet at a time | 1 |
| P4 | workflow fan-out | 3 sonnet | 3 |
| P5 | workflow fan-out | 5 ADR + 4 doc + curator | ≤10 |
| P6 | inline + 1 review workflow | ≤10 sonnet | ≤10 |

Rules repeated because they are binding: **sonnet/haiku only, never fable workers · ≤15 concurrent · G1–G4 between phases · one commit per task · stage explicit paths (no `git add -A`) · anti-circle rule on every discovery · hard-stop = park-and-continue, not patch.**
