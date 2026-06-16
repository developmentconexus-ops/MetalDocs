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

## Commands
- Start API: `.\scripts\start-api.ps1`
- Rebuild/start API: `.\scripts\start-api.ps1 -Build`
- System runnable check: `.\scripts\check-system-runnable.ps1`
- Go build: `go build ./...`
- Go tests: `go test ./...`
- Frontend tests: `make test`
- Docx workspace build/test/typecheck: `npm run build:docx-v2`, `npm run test:docx-v2`, `npm run typecheck:docx-v2`

## Context Map
| Task | Read |
|---|---|
| General orientation | `wiki/index.md`, then `wiki/architecture/system-map.md` |
| Local startup/runtime | `wiki/references/local-dev-startup.md` |
| Backend/API route or contract | `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md` |
| Frontend under `frontend/apps/web` | `wiki/architecture/frontend-structure.md` |
| Query/API client work | `wiki/architecture/frontend-structure.md` query/API sections plus generated API types |
| Database/migration/bootstrap | `wiki/database/index.md` and relevant database docs |
| QA/close-out | `wiki/quality/qa-operating-system.md` and relevant `wiki/quality/*-checklist.md` |
| Test framework discipline | `wiki/quality/test-discipline.md`, ADR `wiki/decisions/0034-integration-test-fixture-framework.md` |
| Program/milestone work | `.claude/skills/mission/SKILL.md` or `.claude/skills/milestone/SKILL.md` |
| GitNexus-heavy exploration | `.claude/skills/gitnexus/*/SKILL.md` |
| Docs governance/wiki sync | `wiki/standards/documentation-governance.md`, `.claude/agents/wiki-curator.md` |

## Workflow
- Load only the docs needed for the task boundary.
- Prefer the wiki domain indexes over global file dumps.
- Use GitNexus only for high-risk, unfamiliar, cross-module, rename, refactor, or blast-radius work.
- Use Context7 for current library/framework/API docs.
- For prerequisite failures in startup, auth/session, target route, or contract/generated alignment, stop local feature work and repair the prerequisite first.