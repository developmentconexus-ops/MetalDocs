# Tech Debt Register — editor-ui-eigenpal

> Companion to `wiki/modules/editor-ui-eigenpal.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/editor-ui-eigenpal-refactor.md`.

**Last verified:** 2026-07-03

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Triggers used here:

- **Critical** — schema/version drift the boot check should catch but does not; or data-loss / supply-chain path where fresh installs break.
- **Major** — defense-in-depth gap; documented contract not followed; duplicated write surfaces; cross-module dependency that blocks another module's refactor; false-pass test risk on a load-bearing branch.
- **Minor** — latent symbol, doc/code drift on a non-load-bearing path, missing standalone ADR for a rule already enforced by code + tests.

## Items

### T-001 · Vendored eigenpal tarball absent from `main` — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** previously `packages/editor-ui/package.json`, `apps/docgen-v2/package.json`, `frontend/apps/web/package.json` — each referenced `file:../../[…]/third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- **Resolution (2026-05-11):** Tarball restored at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` from git history. **2026-06-14:** tarball relocated to `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. **2026-06-23:** vendored tarball fully retired; `@eigenpal/docx-editor-react@1.9.0` now installed from npm registry. Tarball path `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` deleted; `third_party/eigenpal/NOTICE` present. All `package.json` `file:` refs replaced with npm registry refs.
- **Evidence:** `@eigenpal/docx-editor-react@1.9.0` in `package.json` dependencies; `third_party/eigenpal/NOTICE` present.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-001` (closed)
- **Linked ADR:** `wiki/decisions/0001-eigenpal-adoption.md`

### T-002 · TemplateEditorPage bypasses the `MetalDocsEditor` wrapper — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — previously imported `DocxEditor` directly from `@eigenpal/docx-editor-react`.
- **Resolution (2026-05-11, commit `60fa5473`):** `TemplateEditorPage` migrated to `MetalDocsEditor`. Direct `@eigenpal/docx-editor-react` imports removed; `useRef<MetalDocsEditorRef>` now used. Repo-wide grep `@eigenpal/docx-editor-react` in `frontend/apps/web/src` returns zero outside type-only positions. Anti-Corruption Layer now holds for both consumer pages.
- **Evidence:** `_artifacts/03-deps.md` IN-edges table (updated).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-002` (closed)
- **Linked ADR:** missing-ADR (see T-008 — the wrapper-only rule still lacks its own ADR; the package rename from `@eigenpal/docx-js-editor` to `@eigenpal/docx-editor-react` 2026-06-23 is covered by the ADR 0001 amendment)

### T-003 · `templatePlugin.wiring.test.tsx` asserts pre-gating contract — **RESOLVED 2026-05-11**
- **Severity:** major → **resolved**
- **Surface:** `packages/editor-ui/test/templatePlugin.wiring.test.tsx`
- **Resolution (2026-05-11, commit `ce6d809a`):** Test rewritten to 5 correct assertions gated on `template-draft` mode, aligned with the 2026-05-06 plugin-gating refactor. `document-edit` path asserts `data-plugins='0'` for `templatePlugin`. No stale contract survives.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md` "Stale wiring spec" subsection (now stale-free).
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-003` (closed)
- **Linked ADR:** missing-ADR (see T-007 for the gating rule itself)

### T-007 · No ADR for `templatePlugin` mode-gating rule
- **Severity:** minor
- **Surface:** `packages/editor-ui/src/MetalDocsEditor.tsx:55-56`; rationale lives only in source comments and `wiki/modules/editor-ui-eigenpal.md` "Plugin registration § templatePlugin mode gating".
- **Observation:** The rule "do not re-add `templatePlugin` unconditionally to document-edit; use CSS to hide chips instead" is enforced in code and prose. No ADR captures the decision or its rationale.
- **Evidence:** `_artifacts/02-flow-plugin-registration.md`; missing entry in `wiki/decisions/` index.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-007`
- **Linked ADR:** ADR 0047 — decision recorded; implementation in Phase 3

### T-008 · No ADR for Anti-Corruption Layer / wrapper-only consumption rule — CLOSED (ADR 0046; ratified 2026-07-02 by ADR 0064)
- **Severity:** minor (closed)
- **Surface:** `packages/editor-ui/` as a whole; rule implied by ADR 0001 § Consequences ("All editor-related code consolidates in `packages/editor-ui/`"), `wiki/references/eigenpal-controlled-package.md` § "What belongs in MetalDocs docs".
- **Observation (original):** No ADR explicitly mandates that all `@eigenpal/docx-editor-react` access in `frontend/apps/web` goes through `@metaldocs/editor-ui`. T-002 (TemplateEditorPage bypass) was a consequence of this gap — now resolved; the rule still lacks a formal decision record.
- **Resolution:** ADR 0046 (Status: Accepted, 2026-06-26) already fully answers this — its own Consequences section states it "closes tech-debt T-008." Enforcement is live via `eslint.config.mjs:16-25` (`no-restricted-imports` banning `@eigenpal/*` outside the two ACL walls) plus `packages/editor-ui/test/public-surface.test.ts`. ADR 0064 (2026-07-02) verified this and formally ratifies the closure — no second/competing ADR was written; this row's heading is updated to CLOSED to match its own "Linked ADR" field, which already named ADR 0046.
- **Evidence:** `_artifacts/03-deps.md` direct-eigenpal IN-edges table; `eslint.config.mjs:16-25,62-63`; `wiki/decisions/0064-eigenpal-wrapper-only-already-decided.md`.
- **Linked backlog row:** `backlog/editor-ui-eigenpal-refactor.md#R-008` (can be closed)
- **Linked ADR:** `wiki/decisions/0046-eigenpal-anti-corruption-layer.md` (primary decision); `wiki/decisions/0064-eigenpal-wrapper-only-already-decided.md` (closure ratification)

### T-009 · Worker retry not wired on `*RenderError.Retryable()` — **RESOLVED 2026-06-26, test-hardened 2026-07-02 (APP-02)**
- **Severity:** major → **resolved**
- **Surface:** `internal/platform/worker/service.go` `markFailure`; error originates in `internal/modules/render/fanout/client.go` (`*RenderError`, `Retryable() bool`) and flows through `internal/modules/documents/application/freeze_service.go` (`Materialize`, wraps `fmt.Errorf("materialize: fanout: %w", err)`) and `internal/platform/worker/materialize_job_runner.go` (`Handle`, wraps `fmt.Errorf("materialize job runner: %w", err)`) into the worker's outbox failure path.
- **Resolution (commit `9aab29c5`, 2026-06-26, "harden(eigenpal): audit remediation across ACL render + worker paths"):** `markFailure` now does a structural `errors.As(handleErr, &interface{ Retryable() bool })` match — unwrapping through both `%w` layers — and forces `attempt = MaxAttempts` when the matched error reports `Retryable() == false`, so a permanent defect (e.g. `template_parse`, 4xx) dead-letters on the very first observed failure instead of consuming the full retry budget. Transient/unclassified errors (5xx, network errors, or anything not implementing the interface) keep the existing exponential-backoff path unchanged. The match is structural (not `errors.As(&*fanout.RenderError)`) specifically so `internal/platform/worker` stays decoupled from the `render` module — no import inversion.
- **2026-07-02 (APP-02) gap found and closed:** the fix logic existed but had zero unit-test coverage proving the branch — `service_test.go` only covered the unrelated `errUnsupportedEventType` DLQ path. `internal/modules/render/fanout/reconstruction.go`'s `ReconstructService.Reconstruct` still propagates the raw (unwrapped) `*RenderError` — that path is a manual/forensic re-render invoked outside the outbox retry loop (no worker consumer calls it), so `Retryable()` classification does not apply there; confirmed no consumer retries it blindly.
- **Evidence:** `internal/platform/worker/service_test.go` — `TestWorkerService_NonRetryableRenderError_DeadLettersOnFirstAttempt` (permanent error → `DeadLetteredAt` set, `NextAttemptAt` nil, on `AttemptCount: 1`) and `TestWorkerService_RetryableRenderError_SchedulesBackoffLikeBefore` (transient error → `NextAttemptAt` set, `DeadLetteredAt` nil, unchanged backoff). Both drive `markFailure` with a `%w`-wrapped fake classified error mirroring the real two-layer wrap chain. `go build ./...`, `go test ./internal/platform/worker/... ./internal/modules/render/... ./internal/modules/documents/application/...`, `go vet -tags integration ./...` all green.
- **Linked backlog row:** (not filed — resolved before filing)
- **Linked ADR:** none (structural-interface pattern is a targeted platform/module decoupling technique, not a standalone architectural decision)

### T-010 · Renderer OTel exporter absent (REQ-OBS-3 / RF-1 deferred) — **RESOLVED 2026-07-03 (TST-05)**
- **Severity:** minor → **resolved**
- **Surface:** `apps/docx-renderer/src/render/fanout.ts`, `apps/docx-renderer/src/routes/fanout.ts`, `internal/modules/render/fanout/client.go`.
- **Observation (as found):** the `ProcessTemplateOptions.traceparent` field existed in `packages/eigenpal-adapter/src/index.ts` but was dead — no caller ever set it (`apps/docx-renderer/src/render/fanout.ts`'s `fanout()` never passed `opts.traceparent` to `processTemplate`). The Go fanout client (`internal/modules/render/fanout/client.go`, `Client.Fanout`) never injected a `traceparent` header on the outbound request, and `internal/platform/httpclient/internal_client.go`'s `NewInternalClient` had no `otelhttp` transport. No `@opentelemetry/*` package existed anywhere in the monorepo. The worker→renderer hop (`FreezeService.Materialize` → `fanout.Client.Fanout`, invoked from the outbox consumer inside the `metaldocs-api` process) was confirmed to carry a real `ctx` with an active span when one exists, but nothing propagated it.
- **Resolution:** Renderer-side OTel installed following the exact env conventions of the Go `SetupOTel` (`internal/platform/observability/otel.go`): `OTEL_TRACES_EXPORTER` / `OTEL_EXPORTER_OTLP_ENDPOINT` unset ⇒ clean no-op (global OTel API no-op tracer, zero SDK install, no log spam); `OTEL_TRACES_EXPORTER=none` short-circuits even with an endpoint set; `OTEL_TRACES_EXPORTER=console` writes spans to stdout; otherwise OTLP/HTTP export. Service name `docx-renderer`, overridable via `OTEL_SERVICE_NAME` (same override convention as Go). New `apps/docx-renderer/src/observability/otel.ts` (`setupOTel`, `getTracer`) and `apps/docx-renderer/src/observability/tracing.ts` (`extractTraceContext` — W3C propagator extraction from inbound headers; `withRenderSpan` — wraps the render call in a `docx_renderer.render_fanout` span, parented to the extracted context, root span when absent, records exceptions + sets Error status on throw). Wired into `apps/docx-renderer/src/routes/fanout.ts:73-108` (extract before the render call, span carries `tenant_id` only — no document content, no secrets) and `apps/docx-renderer/src/index.ts` (SDK installed before `buildApp()`; `SIGTERM`/`SIGINT` handlers flush via `otel.shutdown()` before exit). On the Go side, `internal/modules/render/fanout/client.go`'s `Client.Fanout` now calls `otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))` right before `c.http.Do` — a no-op when OTel is not installed (global propagator stays the OTel no-op), and injects a real W3C `traceparent` when a span is active in `ctx`, closing the one hop REQ-OBS-3 flagged as unpropagated.
- **Dependencies added** (`apps/docx-renderer/package.json`): `@opentelemetry/api@1.9.1`, `@opentelemetry/core@2.9.0`, `@opentelemetry/exporter-trace-otlp-http@0.220.0`, `@opentelemetry/resources@2.9.0`, `@opentelemetry/sdk-trace-base@2.9.0`, `@opentelemetry/sdk-trace-node@2.9.0`, `@opentelemetry/semantic-conventions@1.41.1`. `build.mjs`'s existing "externalize all real node_modules deps" esbuild plugin requires no change — these load natively at runtime like `fastify`/`minio`.
- **Tests:** `apps/docx-renderer/src/observability/__tests__/otel.test.ts` (6 tests — inert when unconfigured, inert on `OTEL_TRACES_EXPORTER=none` even with endpoint set, installs on `console`, installs on endpoint-only, never throws on an unreachable OTLP endpoint at setup time, global no-op tracer usable pre-install); `apps/docx-renderer/src/observability/__tests__/tracing.test.ts` (5 tests — real upstream span's traceparent parents the render span, missing header produces a root span without throwing, malformed header does not throw, `tenant_id` attribute set only when provided, exception recorded + Error status + rethrow). `internal/modules/render/fanout/client_test.go` — two new cases: `TestClient_Fanout_InjectsTraceparent_WhenSpanActive` (real SDK `TracerProvider` + `tracetest.SpanRecorder`, asserts the outbound `traceparent` header carries the active trace ID) and `TestClient_Fanout_NoTraceparent_WhenPropagatorUnset` (empty composite propagator ⇒ no header added).
- **Evidence:** `apps/docx-renderer/src/observability/otel.ts`, `apps/docx-renderer/src/observability/tracing.ts`, `apps/docx-renderer/src/routes/fanout.ts:8,73`, `apps/docx-renderer/src/index.ts:9,26-41`, `internal/modules/render/fanout/client.go:12-13,80-85`, `internal/modules/render/fanout/client_test.go`. Verified: `npm run build:docx-v2`, `npm run typecheck:docx-v2`, `npm run test:docx-v2` (19 test files / 53 tests in `docx-renderer` alone, 121 files / 746 tests monorepo-wide) all green; `go build ./...`, `go vet ./...`, `go test -count=1 ./internal/modules/render/... ./internal/platform/observability/...` all green.
- **Linked backlog row:** none was filed for T-010 (RF-1); no `wiki/backlog/editor-ui-eigenpal-refactor.md` row to close.
- **Linked ADR:** none (implementation against the existing REQ-OBS-3 target; mirrors the already-accepted Go `SetupOTel` pattern, no new architectural decision).

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 0 / 9 (all exports cited in §5.2 of `editor-ui-eigenpal.md`)
- Operations missing C4 placement: 0 / 0 (no HTTP)
- Cross-deps missing in §5/§8: 0 / 5
- State transitions missing in §6: 0 / 0 (no state machine)
- Decisions without ADR link: 0 / 5 (T-007 → ADR 0047, T-008 → ADR 0046; T-004, T-005, T-006 removed as stale 2026-06-23; T-002 and T-003 resolved)
