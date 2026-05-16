# Current Agent Handoff

> **Last verified:** 2026-05-16
> **Scope:** Recovery context for a fresh or reinstalled Codex session.
> **Out of scope:** Full implementation details for each screen. Follow the linked plan/spec/backlog files for that.

## Why this exists

Codex Desktop became unstable and the user planned to reinstall it. This page is the durable landing document for the next session so the agent does not lose the current roadmap state, the Plan 12 screen-finalization intent, or the recovered information from an inaccessible templates review thread.

Start a fresh session by reading this file after `AGENTS.md`, `CLAUDE.md`, `wiki/README.md`, and `wiki/references/ai-operating-system.md`.

## Workspace state at handoff

- Repository: `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\Documents\MetalDocs`
- Current branch observed: `codex/release-v2-name-polish`
- Latest observed commit: `7b79f48b docs: design template wizard stabilization`
- Unrelated untracked file observed: `repair-codex-chat-render.ps1`
- Do not delete, stage, or commit `repair-codex-chat-render.ps1` unless the user explicitly asks.

## Current roadmap position

Plan 12 is the screen finalization plan. It converts the mock-era designed screens into real product screens backed by runtime capability, OpenAPI/generated types, TanStack Query, and documented backend behavior.

The user-reported Plan 12 status is:

| Screen plan | Screen | User-reported status |
|---|---|---|
| Plan 12.1 | `templates` list | done |
| Plan 12.2 | `caixa-aprovacao` | done |
| Plan 12.3 | `novo-documento` | done |
| Plan 12.4 | `novo-template-wizard` | functionally working, but see recovered review checkpoint below |
| Plan 12.5 | `template-editor` | next target |
| Plan 12.6 | `documento-publicado` | remaining |
| Plan 12.7 | `library` / biblioteca | remaining |

Important: the next agent must not treat "visual done" as "product done." The original mock screens intentionally used mock data for desired future behavior. Plan 12 is the real-integration pass: keep the visual direction, but wire only behavior that exists or is explicitly implemented.

## Plan 12 operating rule

Before implementing any screen:

1. Use `metaldocs-frontend`.
2. Use `metaldocs-screen-integration-audit` when the screen/backlog may contain mock-era widgets, legacy wrappers, missing backend capability, or deferred behavior.
3. Use `metaldocs-screen-implementation` before writing TSX/CSS for a designed screen.
4. Use `metaldocs-tanstack-query` when touching frontend API wrappers, generated API types, query hooks, query keys, cache invalidation, or server state.
5. Use `runtime-contract-prereq` if startup, auth/session, target route, OpenAPI/generated types, or frontend wrapper truth is not trustworthy.
6. Use `verification-before-completion` before claiming fixed, done, green, or passing.

Classify every mismatch before touching code:

| Classification | Meaning |
|---|---|
| `runtime prerequisite` | Startup/auth/route does not run reliably enough for screen work. |
| `shared contract prerequisite` | OpenAPI/runtime/generated/frontend wrappers disagree. |
| `module-local implementation` | Backend/API/module fix belongs to the owning module and is needed now. |
| `screen-local implementation` | UI/state/query fix belongs to the screen PR. |
| `wiki-memory drift` | Code truth changed and wiki/backlog/module docs need sync. |
| `workflow/tooling gap` | A skill, script, or guide failed to prevent confusion. |
| `defer` | Desired product behavior lacks backend/runtime support and must not be faked. |

Default mismatch rule: continue only if the mismatch is local to the current task boundary. Otherwise stop and surface the prerequisite first.

## Recovered inaccessible thread

The user could not open this thread:

- Thread id: `019e3160-26da-78e1-b81c-19017a5f2b5d`
- URI: `codex://threads/019e3160-26da-78e1-b81c-19017a5f2b5d`
- Local backup found: `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\.codex\repair-backups\single-session-render-20260516-132734\rollout-2026-05-16T12-20-45-019e3160-26da-78e1-b81c-19017a5f2b5d.jsonl`

Recovered summary:

- The vanished thread started because a templates-v2/templates review session disappeared.
- It identified the active branch as `codex/release-v2-name-polish`.
- It reconstructed a review from the branch diff because no saved review artifact existed.
- It found the branch was not ready to merge.
- The branch mixed wizard completion, release-facing V2 naming cleanup, OpenAPI/generated truth, DB migration correctness, frontend wrappers, screenshots, and wiki drift.
- A subagent review also found critical contract drift around `POST /api/v1/templates`.
- The user chose a split sequence: finish the template wizard professionally first, then handle broader release cleanup separately.
- The session wrote and committed `docs/superpowers/specs/2026-05-16-template-wizard-stabilization-design.md`.

Recovered review blockers to verify before declaring Plan 12.4 closed:

| Finding | Action from recovered design |
|---|---|
| `POST /api/v1/templates` OpenAPI partial, bundled OpenAPI, runtime, and frontend wrapper disagree | Fix contract truth, regenerate backend/frontend types, align wrapper |
| Wizard slug/stepper can bypass required slug validation | Fix as screen-local wizard behavior |
| Active template wizard/catalog paths still used raw `fetch` | Move touched active paths to canonical API client behavior |
| Upgrade migration can leave template DB tripwire disabled after table rename | Fix migration/runtime DB truth |
| Templates wiki still referenced old placeholder catalog path | Sync wiki memory |
| Screenshot/artifact files appeared under nested `frontend/apps/web/frontend/apps/web/...` | Clean artifact placement |
| Global V2 inventory still had many hits | Defer to release cleanup unless blocking the wizard |

The committed design spec is the authoritative recovery artifact:

- `docs/superpowers/specs/2026-05-16-template-wizard-stabilization-design.md`

## Plan 12.5 next target

Plan 12.5 is the `template-editor` screen. Do not start broad implementation until the previous section is resolved or consciously accepted by the user as a separate follow-up.

Recommended first reads for Plan 12.5:

- `docs/superpowers/specs/2026-05-13-plan-12-screens.md`
- `wiki/backlog/template-editor.md`
- `wiki/modules/templates.md`
- `wiki/modules/templates-tech-debt.md`
- `wiki/modules/editor-chrome.md`
- `wiki/modules/editor-ui-eigenpal.md`
- `wiki/architecture/frontend-structure.md`
- `wiki/architecture/api-contract.md`
- `wiki/concepts/design-workflow-audit.md`
- `frontend/apps/web/design-source/template-editor/NOTES.md` if present

Recommended gates before screen work:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
```

Template-editor specific focus:

- Preserve the current visual direction unless the user asks for redesign.
- Do not add mocked comments, fake version history, fake status workflows, or fake backend metrics.
- Use real generated contract types and canonical API wrappers.
- Verify editor-chrome and editor-ui-eigenpal behavior before changing editor integration.
- If a backlog item requires a backend endpoint that does not exist, mark it `defer` in the screen notes/backlog instead of mocking it.

## Documents naming context

The user stated they are changing references from `documents_v2` to `documents` and that functionality is working as intended. Treat this as naming consistency unless direct verification shows runtime or contract drift. Do not launch a broad rename sweep from Plan 12.5 unless the user explicitly asks.

## How to proceed in a fresh session

1. Read `AGENTS.md`, `CLAUDE.md`, `wiki/README.md`, `wiki/references/ai-operating-system.md`, and this file.
2. Run `git status --short` and identify current branch.
3. Inspect whether `docs/superpowers/specs/2026-05-16-template-wizard-stabilization-design.md` has been implemented after commit `7b79f48b`.
4. If not implemented, ask the user whether to complete that stabilization before Plan 12.5 or keep it as a separate tracked follow-up.
5. If proceeding to Plan 12.5, create or read the Plan 12.5 execution plan and start with `metaldocs-screen-integration-audit`.
6. After implementation changes, run the relevant verification gates and sync wiki memory with `metaldocs-module-doc-sync` for touched documented modules.

