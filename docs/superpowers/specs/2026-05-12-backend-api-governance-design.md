# Backend/API Governance Harness Design

> **Last verified:** 2026-05-12
> **Scope:** Step 0 governance layer before resuming Plan 8: backend/API wiki rules, required agent pointers, project skill, and Plan 8 restructure gate.
> **Out of scope:** Implementing Plan 8 module migrations, changing OpenAPI routes, regenerating code, or refactoring handlers.
> **Key files:**
> - `wiki/architecture/api-contract.md` - current contract-first architecture and migration status
> - `wiki/architecture/api-design-system.md` - API conventions for errors, pagination, idempotency, authz, naming
> - `wiki/README.md` - wiki front door and drift policy
> - `CLAUDE.md` - current agent-facing project instructions
> - `AGENTS.md` - Codex-facing project instructions
> - `.claude/skills/metaldocs-frontend/SKILL.md` - existing frontend skill pattern to mirror

---

## Problem

MetalDocs has the right contract-first direction, but backend modules do not yet share one enforced HTTP/API workflow. Registry is mostly canonical, templates_v2 is partially migrated, documents is split across several handlers with spec drift, and approval/taxonomy/audit still expose raw routes. Plan 8 exposed the risk: an implementation plan can be directionally correct while assuming handler ownership, route names, or generated method signatures that do not match the actual code.

This is a professional SaaS codebase. The fix is not to patch around mismatches. The fix is to create a durable governance harness so every future backend/API change starts from the same rules, same route truth process, and same verification gates.

## Decision

Adopt **Option A: wiki as source of truth, skill as workflow, agent files as pointers**.

The governance stack has three layers:

| Layer | Responsibility |
|---|---|
| Wiki architecture doc | Canonical backend/API rules and target patterns. Humans and agents read this as the source of truth. |
| Project skill | Repeatable workflow for backend/API work: orient, build route truth table, compare spec/codegen/wiki, choose canonical pattern, verify. |
| `CLAUDE.md` / `AGENTS.md` | Short pointers only. They tell agents which skill and wiki docs to use without duplicating the rules. |

## Deliverables

### 1. Backend/API governance wiki doc

Create `wiki/architecture/backend-api-structure.md`.

It defines:

- Module HTTP ownership rules.
- OpenAPI-first route authoring.
- Per-module `internal/modules/<module>/api/{cfg.yaml,gen.go,api.gen.go}` layout.
- Canonical generated wrapper wiring using `ServerInterfaceWrapper`.
- Route truth table requirement before migrations.
- Rules for legacy raw routes.
- Rules for path parameter naming, operationId naming, and tag ownership.
- Rules for avoiding duplicate product surfaces such as competing document-create routes.
- Required checks before changing API contracts.
- Relationship to `api-contract.md` and `api-design-system.md`.

`api-contract.md` remains the operational OpenAPI/codegen guide. `api-design-system.md` remains the API behavior contract. The new doc owns backend/API structure and module migration discipline.

### 2. Backend/API skill

Create `.claude/skills/metaldocs-backend-api/SKILL.md`.

Trigger on:

- Backend HTTP route work.
- OpenAPI or oapi-codegen work.
- Changes under `api/openapi/v1/openapi.yaml`.
- Changes under `internal/modules/*/api/`.
- Changes under `internal/modules/*/delivery/http/` or `internal/modules/documents/approval/http/`.
- API behavior work involving RFC 9457, idempotency, authz, tenant context, pagination, or frontend generated types.

The skill workflow:

1. Read `wiki/README.md`, then `wiki/architecture/backend-api-structure.md`, `api-contract.md`, and `api-design-system.md`.
2. Identify affected module docs and tech-debt registers.
3. Build a route truth table from actual runtime registrations before changing routes.
4. Compare runtime routes to OpenAPI paths, operationIds, generated `ServerInterface`, wiki module docs, and tech-debt records.
5. Classify each mismatch as product decision, spec drift, handler drift, generated-code drift, or legacy debt.
6. Choose the canonical module pattern before implementation.
7. Implement only after the route truth and target pattern are clear.
8. Verify with OpenAPI lint, `GOFLAGS=-mod=mod go generate`, generated interface inspection, `go build`, targeted tests, and frontend codegen when relevant.
9. Update wiki docs and stamps after code changes.

### 3. Codex skill pointer

Create a Codex-discoverable skill or pointer under `.agents/skills/metaldocs-backend-api/SKILL.md`.

The Codex version can either duplicate the short `.claude` skill body or point to it plus the wiki docs. The preferred first version is a self-contained short skill to avoid tool-specific path assumptions.

### 4. Agent instruction pointers

Update `CLAUDE.md` and `AGENTS.md` with short backend/API guidance:

- For backend/API/module HTTP work, use `metaldocs-backend-api`.
- Read `wiki/architecture/backend-api-structure.md`, `wiki/architecture/api-contract.md`, and `wiki/architecture/api-design-system.md`.
- Do not duplicate the rules in the instruction files.

### 5. Plan 8 restructure

Before resuming Plan 8 implementation, rewrite the plan around a route truth table and target-state decisions.

The restructured Plan 8 must include:

- Actual runtime route table by module.
- Spec path and operationId table.
- Generated `ServerInterface` method names and signatures after each generate.
- Explicit product decisions for ambiguous surfaces, especially document creation and audit versioning.
- STOP rules for route ownership, generated signature mismatch, missing private handler, and path parameter mismatch.
- One canonical migration pattern per module.

## Professional Target Pattern

The long-term target is:

```text
OpenAPI path + operationId
  -> per-module generated api package
  -> handler implements generated ServerInterface
  -> RegisterRoutes wires only generated ServerInterfaceWrapper
  -> frontend types generated from the same spec
```

Raw `mux.HandleFunc` routes are allowed only for modules not yet migrated. They must be documented as migration debt and must not be mixed into a module after that module is declared fully migrated.

When a module has multiple internal handlers, the team must make an explicit architectural choice:

- Consolidate route ownership into one HTTP handler when that improves clarity without moving business logic.
- Use a named module-level contract handler when multiple sub-handlers represent stable internal boundaries.

The second option is acceptable only when it is a deliberate module boundary, documented in the module wiki, and tested through generated interface satisfaction. It is not a fallback for avoiding cleanup.

## Plan 8 Implications

Plan 8 should not continue as currently written.

Known corrections needed before implementation:

- `documents` must resolve `POST /api/v2/documents` versus registry-owned atomic create.
- `documents` must align checkpoint restore path parameter naming (`version` versus `versionNum`).
- `documents` must account for export routes being owned by `ExportHandler`.
- `documents` generated method signatures include path/query params and cannot use no-param delegation examples.
- `templates_v2` must complete all 20 routes in spec before it can become wrapper-only.
- `approval` and `taxonomy` need codegen bootstrap from real route truth.
- `audit` must decide whether it remains `/api/v1/audit/events` or gets a v2 route; this is a product/API decision.

## Success Criteria

Step 0 is complete when:

- `wiki/architecture/backend-api-structure.md` exists and is indexed from `wiki/README.md`.
- `.claude/skills/metaldocs-backend-api/SKILL.md` exists.
- `.agents/skills/metaldocs-backend-api/SKILL.md` exists or an explicit Codex-compatible pointer exists.
- `CLAUDE.md` and `AGENTS.md` point to the backend/API skill and wiki docs.
- Plan 8 is updated or superseded so implementation begins from route truth, not assumptions.
- No production code or OpenAPI contract changes are made as part of Step 0.

## Risks

| Risk | Mitigation |
|---|---|
| Rules duplicate across docs and drift | Keep rules in wiki; keep agent files as pointers. |
| Skill becomes too long and agents skip it | Keep `SKILL.md` workflow-focused; move heavy details to wiki. |
| Plan 8 loses momentum | Treat this as Step 0, then resume Plan 8 with a stronger plan. |
| Adapter pattern becomes a shortcut | Require explicit module wiki justification and generated-interface verification. |
| Existing dirty worktree gets mixed into Step 0 | Commit only governance docs and skill/pointer files for Step 0. |

## Verification

For Step 0 documentation and skill files:

```powershell
go build ./...
```

Expected: exits 0.

For the skill content:

- Check frontmatter has `name` and `description`.
- Check `CLAUDE.md` and `AGENTS.md` contain pointers, not duplicated rule bodies.
- Check wiki links point to existing files.
- Check the placeholder scan is clean before committing.
