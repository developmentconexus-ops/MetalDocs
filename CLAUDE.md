# MetalDocs Agent Instructions

## Role
Act as a careful MetalDocs maintainer. Preserve runtime truth, contract truth,
wiki truth, and verification truth. Make small, verified changes; stop on
architecture contradictions instead of patching around them.

## Always-On Rules
- Never read, print, commit, or expose `.env` secrets.
- Use PowerShell scripts for local startup; do not use bash or `source .env`.
- Keep changes scoped to the request. Do not refactor adjacent code or revert user work.
- Runtime truth beats docs. When runtime, contract, generated code, and wiki disagree, classify the mismatch and stop if it is outside the current boundary.
- Evidence before closure: report commands, outcomes, QA/review disposition, and bounded defers before saying done.
- Commits are allowed after verified work; never push without explicit permission.

## Root Cause / Global Maximum

Canonical method: `docs/engineering/root-cause-global-maximum-method.md`.

Before non-trivial work, identify the root cause and target invariant before choosing a patch. Do not optimize inside a known workaround or local maximum. MetalDocs-specific invariants below remain binding constraints on every candidate solution.

## Commands
- Start API: `.\scripts\start-api.ps1`
- Rebuild/start API: `.\scripts\start-api.ps1 -Build`
- System runnable check: `.\scripts\check-system-runnable.ps1`
- Go build: `go build ./...`
- Go tests: `go test ./...`
- Frontend tests: `make test`
- Docx workspace build/test/typecheck: `npm run build:docx-v2`, `npm run test:docx-v2`, `npm run typecheck:docx-v2`

## System Facts (hold these before planning anything)
MetalDocs is a **modular monolith**, 4 binaries: `metaldocs-api` (sync + authz, stateless; also joins River leader election so it can enqueue the periodic maintenance jobs — `stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `document-review-surfacer`, `approval-sla-surfacer`, `release-hold-reconciler`, plus the outbox-retention purge — though only `metaldocs-jobs` subscribes the `maintenance` queue and actually executes them, per ADR 0067's dual-define pattern; `apps/api/cmd/metaldocs-api/main.go:645-672`), `metaldocs-worker` (async outbox consumers), `metaldocs-jobs` (hosts + executes the River periodic jobs from `internal/modules/jobs/maintenance/periodic.go`, plus scheduled-publish + notifications fanout), `docx-renderer` (internal only). The stuck-instance watchdog is alert-only, not auto-canceling (ADR 0068); the old Postgres-lease scheduler and its `lease-reaper` are retired (M5).

**15 bounded-context modules** under `internal/modules/`: approval · audit · auth · controlleddocuments · distribution · documents · iam · jobs · notifications · render · search · security · taxonomy · templates · tokens. Cross-module access goes through a module's application service or published Go interface — **never** another module's repository, SQL, or domain internals. (`approval` was promoted from the nested `documents/approval` exception to a first-class top-level module — ADR 0082, the 15th module, supersedes ADR 0072's nested ruling.)

Non-negotiable invariants (violating these is a defect, not a design choice):
- **AuthZ = capabilities, never roles.** Two-tier PDP: tier-1 route→capability (middleware), tier-2 capability×area in-tx (`authz.Require`), DB tripwire last line. Never reason as "admin/author/editor can X" — reason in capabilities. Governed by ADR 0022.
- **Fixed request lifecycle.** Middleware chain is `panic_recovery → otel → http_obs → cors → origin_protection → pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit → contract_validation → method_not_allowed` (`apps/api/cmd/metaldocs-api/chain.go:25`); new routes don't reinvent auth/validation/errors. Idempotency is **not** a chain link — it is enforced per-handler/per-service where needed (e.g. `internal/modules/documents/approval/application/signoff_idemp.go`, `internal/platform/idempotency/postgres_store.go`). Errors are RFC 9457 `problem+json`.
- **Contract-first.** Routes change ONLY by editing `api/openapi` + `oapi-codegen`. The spec is route truth.
- **Multi-tenant pooled.** Every tenant table has `tenant_id`; tx-local GUCs only; tenant-namespaced blob keys; cross-tenant URL → 404.
- **Async = transactional outbox.** State-write + external side effect never share a tx with a network call; consumers idempotent.
- **DB enforces invariants** (triggers/constraints); app checks are the friendly first line.

Governing target spec (source of truth when this list drifts): `wiki/architecture/backend-target-architecture.md` (REQ IDs; reviews cite them).

**Orientation rule:** before planning any new feature or improvement, state (a) which module(s) own it, (b) which invariants above it must satisfy, and (c) read the owning `wiki/modules/<name>.md`. Plan against the whole system, not only the code immediately around the change. For new modules/features, use `.claude/skills/developing-new-work/SKILL.md` before brainstorming; it emits a system-impact analysis plus Green/Yellow/Red verdict.

## Context Map
| Task | Read |
|---|---|
| General orientation | `wiki/index.md`, then `wiki/architecture/system-map.md` |
| Canonical engineering method | `docs/engineering/root-cause-global-maximum-method.md` |
| Local startup/runtime | `wiki/references/local-dev-startup.md` |
| Backend/API route or contract | `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md` |
| Frontend under `frontend/apps/web` | `wiki/architecture/frontend-structure.md` |
| Query/API client work | `wiki/architecture/frontend-structure.md` query/API sections plus generated API types |
| Database/migration/bootstrap | `wiki/database/index.md` and relevant database docs |
| QA/close-out | `wiki/quality/qa-operating-system.md` and relevant `wiki/quality/*-checklist.md` |
| Test framework discipline | `wiki/quality/test-discipline.md`, ADR `wiki/decisions/0034-integration-test-fixture-framework.md` |
| Starting any new module or feature | `.claude/skills/developing-new-work/SKILL.md` |
| Reviewing a design/plan/diff | `.claude/skills/adversarial-review/SKILL.md` |
| Code relationship / impact tracing | `.claude/skills/gitnexus/SKILL.md` when needed |
| What AI gets wrong in design work | `docs/engineering/defect-class-catalog.md` Part II + Appendix C |
| Program/milestone work | HARNESS-CORE + `docs/HARNESS-PROFILE.md` + `docs/superpowers/ROADMAP.md` |
| Docs governance/wiki sync | `wiki/standards/documentation-governance.md` |

## Workflow
- Load only the docs needed for the task boundary.
- Prefer wiki domain indexes over global file dumps.
- Use Context7 for current library/framework/API docs.
- For prerequisite failures in startup, auth/session, target route, or contract/generated alignment, stop local feature work and repair the prerequisite first.
