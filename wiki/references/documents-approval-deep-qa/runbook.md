# Documents + Approval Deep QA Runbook

Date: 2026-05-20
Status: active
Scope: modern `documents + approval` deep QA execution for canonical `/documents/:id`, approval lifecycle, OCC, authz, and worker-owned scheduled publish

## 1. Purpose

This runbook is the operational entrypoint for deep-QA sessions on the modern `documents + approval` flow.

It exists to remove session-to-session improvisation by centralizing:

- current runtime topology
- startup truth
- surface ownership
- evidence collection recipes
- worker and async validation
- stop rules
- session close-out rules

## 2. Current Runtime Topology

- frontend owns browser rendering at `http://localhost:4173`
- API owns synchronous document and approval HTTP behavior plus schedule enqueue
- `metaldocs-worker` owns outbox and PDF work only
- `metaldocs-jobs` owns scheduled publish cutover execution
- canonical UI truth for governed detail remains `/documents/:id`

## 3. Canonical Startup Paths

- Use `.\scripts\start-api.ps1` as the canonical local startup path
- Default script truth starts API, jobs, and worker unless explicitly disabled
- Use `.\scripts\start-jobs.ps1` only when jobs runtime must be controlled manually
- Run `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-system-runnable.ps1 -TargetRoute <route>` before screen work
- Use `-TargetRoute /api/v1/documents/{id}` or another scenario-specific route before a deep-QA pass
- If document approval depends on DOCX or PDF fanout, confirm the worker-side dependency chain before attributing failures to product logic

## 4. Surface-to-Owner Map

- `/documents/:id`: canonical product surface for governed revision truth
- `GET /api/v1/controlled-documents/{id}/active-document`: technical active-sibling and publish-context lookup
- `POST /api/v1/documents/{id}/finalize`: governed transition into approval flow
- `POST /api/v1/documents/{id}/publish`: publish-now transition
- `POST /api/v1/documents/{id}/schedule-publish`: schedule persistence plus enqueue
- `metaldocs-jobs`: delayed scheduled publish execution
- `metaldocs-worker`: outbox consumption, DOCX/PDF fanout, and PDF completion evidence
- frontend detail screen and backing API proof must agree before a scenario is marked proved

## 5. Evidence Collection Recipes

- Canonical runtime UX: browser proof plus backing API response
- Route truth: runtime API proof preferred; contract test allowed when setup cost is disproportionate
- OCC: focused automated proof allowed by default, but the exact invariant must be named
- Fault injection: explicitly label evidence as injected, not live-runtime
- For `/documents/:id`, capture the visible state, CTA availability, and the backing document status that explains that state
- For `active-document` lookups, capture both the controlled-document ID used and the returned active or published sibling context
- For finalize, publish, and schedule transitions, collect request shape, status code, and the post-action read that proves persisted state
- When async work is involved, record enqueue evidence and execution evidence as separate artifacts

## 6. Worker and Async Validation

- prove enqueue separately from execution
- if schedule request succeeds but no cutover happens, inspect jobs host before API
- do not collapse enqueue success into worker success
- treat `schedule-publish` as a two-phase proof: transactional persistence first, temporal cutover second
- for scheduled publish, confirm the pre-cutover status on `/documents/:id`, then confirm the post-cutover state after `metaldocs-jobs` runs
- for PDF or outbox behavior, inspect `metaldocs-worker` evidence independently from publish-state evidence
- if the API returns success but async side effects do not materialize, classify the first failing boundary instead of filing a generic product bug

## 7. Fault-Injection Map

- authz denial on finalize, publish, or schedule-publish: validate capability and ownership boundaries without mutating happy-path fixtures
- OCC mismatch on finalize or publish: use focused automated proof when concurrent runtime reproduction is too expensive
- missing or stale active-sibling context: validate through `GET /api/v1/controlled-documents/{id}/active-document` before assuming canonical-screen drift
- schedule enqueue without cutover: inject or force time-based scenarios only when jobs-host evidence is captured separately
- snapshot-corruption or invalid-freeze scenarios: prefer injected contract or integration proof until a safe runtime fixture exists
- worker fanout failures: label as worker-owned fault injection when docgen/PDF dependencies are intentionally degraded

## 8. Known Tooling Gaps

- some deep-QA scenarios still depend on dedicated fixtures that may not exist as reusable runtime targets
- authz and snapshot-corruption coverage can require injected proof when safe live fixtures are unavailable
- scheduled publish validation is sensitive to jobs-host runtime health and effective-time control, so evidence must name the exact runtime boundary used
- async failures can look like product regressions unless API, jobs, and worker evidence are split cleanly

## 9. Stop Rules and Classification

- stop on route ownership contradiction
- stop on runtime/contract/wiki divergence that affects the active scenario boundary
- stop when the missing capability is cross-boundary or prerequisite-grade
- stop when startup truth is unclear enough that the current evidence cannot be trusted
- classify every blocked scenario as one of:
- runtime prerequisite
- shared contract prerequisite
- module-local implementation
- screen-local implementation
- wiki-memory drift
- workflow/tooling gap
- defer

## 10. Session Close-Out Checklist

- update matrix row state
- attach evidence links
- update runbook if execution truth changed
- update module wiki if code truth changed
- record exact blockers instead of informal follow-ups
