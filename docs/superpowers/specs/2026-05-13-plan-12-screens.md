# Plan 12 Screen Finalization x7

> Date: 2026-05-13
> Mode: execution design, no implementation
> Scope: Plan 12 coordination for seven MetalDocs design-source screens

## Decision

Plan 12 is open. `wiki/backlog/roadmap.md` currently marks Plan 12 as done, but that is a roadmap bug, not implementation evidence. The execution must begin by correcting that status before screen work starts.

This plan uses one prep/spec PR, seven screen PRs, and one finalization PR. Each screen remains isolated: one screen per PR, no bundled implementation, no shared "while here" refactors.

## Prerequisites

Plans 7, 8, 9, and 11 are prerequisites and must stay confirmed before screen work starts:

- Plan 7: RFC 9457 stable error envelope.
- Plan 8: OpenAPI/codegen contract availability.
- Plan 9: idempotency on POST create flows.
- Plan 11: editor chrome stable for editor screens.

If any prerequisite evidence conflicts with the roadmap, stop and report the exact conflicting file paths.

## Required Sources

Global sources, in order:

1. `wiki/backlog/roadmap.md`
2. `wiki/README.md`
3. `CLAUDE.md`
4. `wiki/concepts/design-workflow-audit.md`
5. `wiki/architecture/frontend-structure.md`
6. `wiki/architecture/api-contract.md`

Screen backlog sources:

- `wiki/backlog/templates.md`
- `wiki/backlog/caixa-aprovacao.md`
- `wiki/backlog/novo-documento.md`
- `wiki/backlog/novo-template-wizard.md`
- `wiki/backlog/template-editor.md`
- `wiki/backlog/documento-publicado.md`
- `wiki/backlog/library-screen.md`

For any backlog row that cites `_artifacts/XX.md section Y`, read that exact section before implementation. No inferred context may replace the cited artifact.

## Screen Order

| Order | Screen | Design source | Backlog | Initial backlog count |
|---:|---|---|---|---:|
| 1 | Templates list | `frontend/apps/web/design-source/templates/` | `wiki/backlog/templates.md` | 5 |
| 2 | Caixa de Aprovacao | `frontend/apps/web/design-source/caixa-aprovacao/` | `wiki/backlog/caixa-aprovacao.md` | 8 |
| 3 | Novo Documento | `frontend/apps/web/design-source/novo-documento/` | `wiki/backlog/novo-documento.md` | 4 open items plus 2 follow-ups |
| 4 | Novo Template Wizard | `frontend/apps/web/design-source/novo-template-*` | `wiki/backlog/novo-template-wizard.md` | 11 open items plus 1 resolved item |
| 5 | Template Editor | `frontend/apps/web/design-source/template-editor/` | `wiki/backlog/template-editor.md` | 7 |
| 6 | Documento Publicado | `frontend/apps/web/design-source/documento-publicado/` | `wiki/backlog/documento-publicado.md` | 12 |
| 7 | Library | `frontend/apps/web/design-source/library/` or documented actual slug | `wiki/backlog/library-screen.md` | 7 |

The Library design-source slug must be verified before work starts. The current design-source directory listing contains `biblioteca`, while the roadmap and backlog refer to `library`. If no `library` artifact exists and no documented mapping to `biblioteca` exists, stop and report the missing artifact as a blocker.

## Order Rationale

Start with `templates` because it has the smallest backlog and mostly low-risk list-screen deltas. Then handle `caixa-aprovacao`, where several items are local wiring or review cleanup but endpoint gaps must be preserved honestly. Next, run the two wizard screens while Plan 9 idempotency context is fresh. Then handle `template-editor`, which depends on Plan 11 editor chrome stability. `documento-publicado` comes later because it has many backend-dependent gaps. `library` is last because its backlog includes known mocked panels and possible slug drift, making it the highest risk for unsupported behavior.

## Per-Screen Workflow

Each screen PR must run `metaldocs-screen-implementation` with the mandatory `design-workflow-audit` gate:

1. Confirm design-source artifact exists.
2. Read screen `NOTES.md`, design HTML, screenshot, backlog file, and cited artifacts.
3. Create or update the screen `IMPLEMENTATION.md` worksheet.
4. Phase 0 Audit: classify every visible widget as Keep, Cut, or Defer.
5. Record Keep/Cut/Defer in `NOTES.md` and `artifacts/phase0-audit.md` before TSX/CSS changes.
6. Phase 1 Map: placement, state, query, route, and backend contract decisions.
7. Implement only Keep items.
8. Preserve Cut/Defer items in the backlog with exact rationale and backend dependency.
9. Run frontend verification, dev smoke, screenshot capture, and frontend-screen-reviewer.
10. Dispatch wiki-curator or run the documented module-doc sync required by changed module docs.

No screen may progress past Phase 0 without evidence artifacts. Visual parity is not self-approved; the user remains the visual approver.

## Stop Conditions

Stop and report exact blockers when any of these occur:

- Required backend endpoint, response field, request field, status enum, or RBAC role contract does not exist.
- API contract ambiguity would force guessing.
- Design artifact is missing or its slug cannot be mapped by documented evidence.
- Design conflicts with real product state, persona, or RBAC rules and cannot be resolved using existing docs.
- A UI element requires mocked data to look correct.
- Two screens would need to land in the same PR.
- The audit phase or frontend-screen-reviewer would be skipped.

## No-Fallback Rules

- No fake stat cards, placeholder backend states, mocked counts, or hardcoded user-facing values.
- No new backend behavior inside screen PRs.
- If a design element depends on a missing backend contract, classify it as Defer.
- If a local fix exposes unsupported behavior, classify it as Cut or Defer instead.
- Do not add compatibility shims or legacy frontend paths.

## PR Structure

Prep/spec PR:

- Add this spec.
- Correct `wiki/backlog/roadmap.md` so Plan 12 is open/not started and notes the erroneous done stamp.
- Preserve done evidence for Plans 7, 8, 9, and 11.

Screen PRs:

- One PR per screen in the order above.
- Each PR updates only the target screen, its design-source notes/artifacts, its backlog file, and directly affected module docs.
- Each PR includes verification output, smoke notes, screenshots, frontend-screen-reviewer result, and wiki update evidence.

Finalization PR:

- Update `wiki/backlog/roadmap.md` Plan 12 to done only after all seven screen PRs land.
- Ensure all seven backlog files reflect final state.
- Run one grouped final review across the full Plan 12 result.

## Risk Notes

- Roadmap drift already exists: Plan 12 is marked done before implementation. The first execution step must correct this.
- `library` vs `biblioteca` slug drift may block the Library screen until documented.
- Several screen backlogs name missing backend endpoints or fields; those items are likely Defer, not Keep.
- Existing screen code may already contain mocks from earlier iterations. Plan 12 must remove or defer unsupported UI rather than polishing fake behavior.
- Approval and document detail surfaces have partial or raw API contract coverage. Any new contract dependency must stop instead of guessing.
- Wizard screens may tempt client-side previews or local staging behavior. Those are only allowed when backed by existing product rules and honest UI states.

## Success Criteria

- The roadmap accurately shows Plan 12 as open during execution and done only after all seven PRs land.
- Each screen has Phase 0 Keep/Cut/Defer evidence before code changes.
- All Keep items are implemented and verified.
- All Cut/Defer items remain documented with exact rationale.
- No mock data path, fake backend state, or unsupported affordance is introduced.
- Tests and manual smoke pass for each changed screen.
- Frontend-screen-reviewer has no unresolved Critical or Major findings per screen PR.
- Module/backlog/wiki docs are updated from the concrete code changes.
