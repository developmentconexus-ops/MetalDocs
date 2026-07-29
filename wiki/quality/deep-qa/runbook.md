# Documents + Approval Deep QA Runbook

Date: 2026-05-20
Status: active
Canonical home: `wiki/quality/deep-qa/runbook.md`
Compatibility path: `wiki/references/documents-approval-deep-qa/runbook.md`
Scope: modern `documents + approval` deep QA execution for canonical `/documents/:id`, approval lifecycle, OCC, authz, and coordinator-owned release

> **Amended 2026-07-28 ([ADR 0085](../../decisions/0085-release-coordinator-approval-driven-publication.md) Stage B).** `POST /documents/{id}/publish`, `/schedule-publish` and `/supersede` no longer exist, and neither does the capability `document.publish` or the River kind `scheduled_publish_cutover`. Publication is not a request any more: the release coordinator reacts to durable readiness facts (approval fact × final-DOCX and final-PDF artifact facts × effective-date gate × supersession head) through the `release_evaluate` River job. Every publish/schedule step below has been restated in those terms — a QA session that still tries to *call* a publish route is proving a route that was deleted, and will read the resulting 404 as a product bug.

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
- API owns synchronous document and approval HTTP behavior plus the fact writes (approval fact, artifact facts) that enqueue release evaluation
- `metaldocs-worker` owns outbox and PDF work only
- `metaldocs-jobs` owns release evaluation (`release_evaluate`) — the only executor that publishes
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
- the terminal approving signoff (`RecordSignoff` → `ReleaseFactRecorder.RecordTerminalApproval`, `internal/modules/approval/application/decision_service.go:592`; the eQMS review verdict path does the same at `review_verdict_service.go:320`): the last human act before release — it writes the approval fact the coordinator keys on. There is no publish request to make after it.
- `metaldocs-jobs`: executes the `release_evaluate` job — the only actor that moves a document to `published` (system principal `system:release-coordinator`)
- `metaldocs-worker`: outbox consumption, DOCX/PDF fanout, and PDF completion evidence
- frontend detail screen and backing API proof must agree before a scenario is marked proved

## 5. Evidence Collection Recipes

- Canonical runtime UX: browser proof plus backing API response
- Route truth: runtime API proof preferred; contract test allowed when setup cost is disproportionate
- OCC: focused automated proof allowed by default, but the exact invariant must be named
- Fault injection: explicitly label evidence as injected, not live-runtime
- For `/documents/:id`, capture the visible state, CTA availability, and the backing document status that explains that state
- For `active-document` lookups, capture both the controlled-document ID used and the returned active or published sibling context
- For finalize, submit and signoff transitions, collect request shape, status code, and the post-action read that proves persisted state. For the release itself there is no request to capture: the evidence is the `release_generations` fact row before, and `documents.status` + `released_at` after.
- When async work is involved, record enqueue evidence and execution evidence as separate artifacts

## 6. Worker and Async Validation

- prove enqueue separately from execution
- if the approval fact lands but the document never publishes, inspect the jobs host before the API
- do not collapse enqueue success into worker success
- treat release as a two-phase proof: the fact write (approval fact, then each artifact fact) is transactional and synchronous with the request that caused it; the release itself is the coordinator's later reaction to those facts
- for a release, confirm the pre-release status on `/documents/:id` **plus** the `release_generations` hold reason that explains it (`materializing`, `awaiting_effective_date`, `supersede_conflict`), then confirm the post-release state after `metaldocs-jobs` runs `release_evaluate`. A document sitting at `approved` with `hold_reason = 'materializing'` is a correct intermediate state, not a stuck publish.
- for PDF or outbox behavior, inspect `metaldocs-worker` evidence independently from publish-state evidence
- if the API returns success but async side effects do not materialize, classify the first failing boundary instead of filing a generic product bug

## 7. Fault-Injection Map

- authz denial on finalize, submit, or signoff: validate capability and ownership boundaries without mutating happy-path fixtures. There is no publish authz case left to inject — `document.publish` is retired and the coordinator releases as the system principal; the cross-document supersede check moved to submit time (`document.supersede` on the **target's** area).
- OCC mismatch on finalize or signoff: use focused automated proof when concurrent runtime reproduction is too expensive
- missing or stale active-sibling context: validate through `GET /api/v1/controlled-documents/{id}/active-document` before assuming canonical-screen drift
- facts complete but no release: inject or force time-based scenarios only when jobs-host evidence is captured separately. Distinguish the three holds before filing anything — a missing artifact fact (`materializing`) and a future `planned_effective_from` (`awaiting_effective_date`) are the coordinator working, not failing.
- snapshot-corruption or invalid-freeze scenarios: prefer injected contract or integration proof until a safe runtime fixture exists
- worker fanout failures: label as worker-owned fault injection when docgen/PDF dependencies are intentionally degraded

## 8. Known Tooling Gaps

- some deep-QA scenarios still depend on dedicated fixtures that may not exist as reusable runtime targets
- authz and snapshot-corruption coverage can require injected proof when safe live fixtures are unavailable
- release validation is sensitive to jobs-host runtime health and effective-time control, so evidence must name the exact runtime boundary used
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
