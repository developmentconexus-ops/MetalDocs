---
name: metaldocs-screen-integration-audit
description: Use when a MetalDocs designed screen or screen backlog must be checked against real backend capabilities, API contracts, TanStack Query wiring, generated types, module wiki truth, and missing/deferred product behavior before implementation.
---

# MetalDocs Screen Integration Audit

Use this before real screen finalization when a `design-source/<slug>/` screen may contain mock-era widgets, legacy frontend wiring, missing API behavior, or deferred backlog items.

This skill does not implement code. It produces the integration boundary for the next implementation session.

## Clarification Gate (No Assumptions)

If any item cannot be classified from evidence, stop and ask the user a focused question before continuing.

Ask when any of these are unclear:
- product semantics (what the UI should mean or display)
- backend capability existence or ownership
- response shape/field meaning
- role/status mapping for a visible design element
- whether a gap is local to this screen or shared across modules

Never guess or infer a default for ambiguous items.

## Required Skills

Load only what the audit touches:
- `metaldocs-frontend` for screen files, routes, components, and feature placement.
- `metaldocs-tanstack-query` for API wrappers, query hooks, generated frontend types, query keys, cache, invalidation, or mutations.
- `metaldocs-backend-api` for HTTP routes, OpenAPI, generated backend APIs, handlers, route ownership, or API behavior.
- `runtime-contract-prereq` if startup/auth/route truth or runtime/spec/generated/frontend-wrapper alignment is failing.

Use `metaldocs-screen-implementation` only after this audit says the screen is ready for implementation.

## Inputs

For screen `<slug>`, read the smallest relevant set:
1. `wiki/backlog/<slug>.md`
2. `frontend/apps/web/design-source/<slug>/NOTES.md`
3. current frontend screen/component/API/query files for the slug
4. related module wiki, tech-debt, and refactor backlog
5. runtime route, OpenAPI path, generated backend API, generated frontend types, and feature wrapper for touched modules
6. `wiki/references/ai-operating-system.md` for classifications and gates

Do not bulk-read unrelated modules or artifacts.

## Classify Each Item

Create one row for every meaningful visible widget, action, state, and open backlog item:

| Classification | Meaning | Next action |
|---|---|---|
| `implemented and aligned` | real behavior exists and current wiring matches runtime/contract/wiki truth | eligible for screen implementation |
| `implemented but legacy-wired` | behavior exists, but frontend wrapper/query/codegen usage is old or inconsistent | repair locally if screen-bounded |
| `screen-local integration fix` | missing wiring is owned by this screen/module and does not change shared contracts | include in screen PR |
| `shared contract prerequisite` | runtime/OpenAPI/generated/frontend-wrapper mismatch affects shared consumers | stop and use `runtime-contract-prereq` |
| `missing backend capability` | design/backlog expects behavior with no real endpoint/service/contract | split prerequisite or defer |
| `defer` | product semantics or backend capability are not ready for this screen finalization | document and leave out |

## Report Format

Save the formal report in `wiki/backlog/<slug>.md` under a dated `Integration Audit` section. Add or update a short summary in `design-source/<slug>/NOTES.md`.

Use this table:

```markdown
| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Template list | design + backlog | `GET /api/v1/templates` exists | wrapper present; envelope check needed | screen-local integration fix | include in Plan 12.1 |
```

Then include:
- `Ready for implementation`: items that can proceed now.
- `Prerequisites`: shared/runtime/backend work that must happen first.
- `Deferred`: items intentionally excluded from the real implementation.
- `Verification needed next`: exact commands or gates to run before coding.

## Stop Rules

Stop before implementation if:
- startup, auth/session, target route, or generated contract truth is failing
- a design item requires a backend capability that does not exist
- an API mismatch affects more than the current screen/module
- the screen backlog and module wiki contradict route ownership, status enums, or capability existence
- classification depends on an assumption instead of evidence

If the issue is local, record it as a local integration fix. If it is shared, route it to prerequisite work.

## Completion Checklist

- Every design/backlog item has a classification.
- No mock-only widget is marked implementable.
- Every missing capability is either prerequisite or defer.
- Existing legacy frontend API/query wiring is identified.
- Formal report is in `wiki/backlog/<slug>.md`.
- Design-local summary is in `design-source/<slug>/NOTES.md`.
- Next workflow is named: screen implementation, runtime prerequisite, backend/API prerequisite, or defer.
