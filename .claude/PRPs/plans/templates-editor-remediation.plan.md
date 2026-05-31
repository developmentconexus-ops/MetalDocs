# Plan — Template Editor remediation (post qa/templates-editor)

> Source: [QA-REPORT-templates-editor.md](../../../frontend/apps/web/QA-REPORT-templates-editor.md) (2026-05-31).
> Grounding: [CLAUDE.md](../../../CLAUDE.md), [wiki/quality/qa-operating-system.md](../../../wiki/quality/qa-operating-system.md), [wiki/architecture/api-design-system.md](../../../wiki/architecture/api-design-system.md), [wiki/concepts/authz-tiers.md](../../../wiki/concepts/authz-tiers.md), [wiki/modules/templates.md](../../../wiki/modules/templates.md) §6 lifecycle / §8.6 concurrency, [wiki/modules/templates-tech-debt.md](../../../wiki/modules/templates-tech-debt.md) T-004 / T-009 / T-010, [wiki/decisions/0012-contract-first-api.md](../../../wiki/decisions/0012-contract-first-api.md), [wiki/decisions/0007-two-tier-authz.md](../../../wiki/decisions/0007-two-tier-authz.md), [wiki/backlog/templates-refactor.md](../../../wiki/backlog/templates-refactor.md) R-004 / R-009 / R-010.

The fix that landed on `qa/templates-editor` (commit `c9b70616`) closed the immediate CRITICAL — FE lifecycle POSTs now send `Idempotency-Key` — but five graded findings remain. This plan describes how each one should be closed, in dependency order, with one branch per workstream, contract test required, and the boundary the change owns.

---

## Workstream order (dependency-aware)

| # | Branch | Boundary | Severity | Hard-stop? | Depends on |
|---|---|---|---|---|---|
| 1 | `fix/templates-publish-role-binding`           | backend / templates module | HIGH | yes | — |
| 2 | `feat/templates-approve-next-draft-response`   | OpenAPI + backend + FE     | HIGH | yes (closed 2026-05-31) | none, but ships first user-visible win |
| 3 | `fix/templates-schema-occ-lock`                | OpenAPI + backend + FE     | HIGH | partial | (2) lands the response-shape pattern reused here |
| 4 | `chore/fe-capability-gates-templates`          | FE only                    | MEDIUM | no | requires `useMe()` capability/role surface (verify exists) |
| 5 | `chore/fe-error-message-coverage`              | FE only                    | MEDIUM | no | independent |

Each branch:
- One owning module per [CLAUDE.md §3 Skill Routing](../../../CLAUDE.md).
- Contract test first ([development-workflow.md](../../../C:/Users/leandro.theodoro/.claude/rules/common/development-workflow.md) TDD), then implementation, then review, then close-out checklist per [qa-operating-system.md](../../../wiki/quality/qa-operating-system.md).
- One PR back to `main` with: scope statement, contract delta (when applicable), evidence block, wiki `Last verified` bump on every doc the change references.

---

## Workstream 1 — `fix/templates-publish-role-binding`

> Closes residual half of [templates-tech-debt T-004](../../../wiki/modules/templates-tech-debt.md). Brings the direct `POST /publish` path to the same two-tier authz parity as `Approve`.

**Skill route.** [`metaldocs-backend-api`](../../../.agents/skills/metaldocs-backend-api/SKILL.md). Not the editor screen — owning boundary is the templates Service.

**Problem.** `Service.PublishTemplateVersion` enforces capability `template.publish` + SoD + `content_hash` gate but does **not** check that the actor's roles satisfy `version.PendingApproverRole`. `Service.Approve` does. Two parallel publish paths, one secure, one not. ISO 9001 §7.5 identity-traceability gap.

**Architecture.** Two-tier authz ([0007-two-tier-authz.md](../../../wiki/decisions/0007-two-tier-authz.md)): Tier 1 = capability (RBAC) — already enforced. Tier 2 = role-binding (workflow) — must be added. Both tiers must pass for the action to commit. Postgres tripwire on `templates_template_version` already catches a missing capability assertion; role-binding is application-layer.

### Plan

1. **Domain helper** — `internal/modules/templates/domain/version.go`: add
   ```go
   // RoleBindingFor returns the role required to approve/publish this version,
   // or "" if no binding is required.
   func (v *TemplateVersion) RoleBindingFor(transition VersionStatus) string {
       switch transition {
       case VersionStatusInReview, VersionStatusApproved:
           if v.PendingReviewerRole != nil { return *v.PendingReviewerRole }
       case VersionStatusPublished:
           if v.PendingApproverRole != nil { return *v.PendingApproverRole }
       }
       return ""
   }
   ```
   Reused by Approve and Publish — DRY ([common/coding-style.md](../../../C:/Users/leandro.theodoro/.claude/rules/common/coding-style.md)).

2. **Service** — `internal/modules/templates/application/lifecycle.go` `PublishTemplateVersion`: after the existing `authz.Require(CapTemplatePublish)` call and *inside the same tx*, add
   ```go
   if role := version.RoleBindingFor(domain.VersionStatusPublished); role != "" &&
       !containsRole(cmd.ActorRoles, role) {
       return nil, domain.ErrForbiddenRole
   }
   ```
   Use the existing `containsRole` helper already used by `Approve`.

3. **Refactor `Approve`** to call the same `RoleBindingFor` helper. One source of truth.

4. **RFC 9457 mapping** — `internal/platform/problem/codes.go`: confirm `forbidden_role` is registered with a stable `code` + `title` ("Papel obrigatório não atendido"). `Handler.MapErr` already routes `domain.ErrForbiddenRole` → 403 — verify on the publish path.

5. **Audit** — `lifecycle.go`: emit `template.publish.forbidden_role` audit event (denied attempt). Sink is canonical `metaldocs.audit_events` per [audit.md](../../../wiki/modules/audit.md) (T-013 closed).

### TDD — write tests first

`internal/modules/templates/application/lifecycle_publish_role_test.go`:
- Table-driven, `-race`:
  - actor has `template.publish` AND `pending_approver_role` → 200, status flips published.
  - actor has `template.publish` but role mismatch → `ErrForbiddenRole` (403 `forbidden_role`), no state change, audit row written.
  - actor lacks `template.publish` → `authz.ErrCapability` (403, canonical authz error — current behavior preserved).
  - `pending_approver_role` nil (no binding required) → publish succeeds with capability alone.

Contract test `internal/modules/templates/delivery/http/routes_contract_test.go`:
- `POST /publish` with valid Idempotency-Key but wrong role → 403 + RFC 9457 `code: "forbidden_role"`.

Both run under `go test -race ./internal/modules/templates/...`.

### Close-out (per [qa-operating-system.md](../../../wiki/quality/qa-operating-system.md))

- Gate 1: `go vet ./... && go test -race ./internal/modules/templates/...`
- Gate 2: `ecc:go-reviewer` on the diff.
- Gate 3: backend QA via curl matrix in `wiki/quality/backend-api-qa-checklist.md`.
- Gate 5: re-run `scripts/check-module-contract-sync.ps1 -Module templates`.
- Wiki bump: `wiki/modules/templates.md` + close T-004 in `templates-tech-debt.md` + move R-004 to closed in `backlog/templates-refactor.md`.

---

## Workstream 2 — `feat/templates-approve-next-draft-response`

> Closes the user-visible gap "approve succeeds but where's the new draft?". Contract change. [0012-contract-first-api.md](../../../wiki/decisions/0012-contract-first-api.md) — OpenAPI first.

**Skill route.** [`metaldocs-backend-api`](../../../.agents/skills/metaldocs-backend-api/SKILL.md) then [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md) for the FE consumer.

**Problem.** `Service.PublishTemplateVersion` returns `PublishTemplateVersionResult{Published, NextDraft}` but `Service.Approve` (the no-reviewer path the editor calls) returns only the current `VersionDTO`. The next-draft id/number is created in-tx but discarded before the HTTP response. FE has no way to navigate the user from `vN` (now published) to `vN+1` (new draft), forcing a list-page bounce.

**Architecture pattern.** Both endpoints terminate the lifecycle with the same side effect (publish + spawn next draft). Their response shapes must mirror. This is a Stripe-grade API hygiene rule: lifecycle transitions return both the *current* resource and any *spawned* resources in one round-trip.

### Plan

1. **OpenAPI partial** — `api/openapi/v1/partials/templates.yaml`:
   ```yaml
   ApproveTemplateVersionResponse:
     type: object
     required: [version]
     properties:
       version: { $ref: '#/components/schemas/VersionDTO' }
       next_draft:
         description: |
           When the Approve transition terminates as publish (no-reviewer path)
           or the approver explicitly publishes, a fresh draft v(n+1) is spawned
           in the same transaction. Returned so the caller can navigate without
           a list refetch.
         type: object
         nullable: true
         properties:
           id: { type: string, format: uuid }
           version_number: { type: integer }
   ```

2. **Regen** — `internal/modules/templates/api/api.gen.go` via oapi-codegen.

3. **Service** — `internal/modules/templates/application/lifecycle.go`: `Approve` already conditionally calls the publish path internally; thread the spawned `NextDraft` into a new `ApproveResult{Version, NextDraft}` return type. Backward-compatible at the HTTP boundary because the field is `nullable`.

4. **Handler** — translate `ApproveResult` to `ApproveTemplateVersionResponse`.

5. **FE types** — `corepack pnpm gen:api` → `frontend/apps/web/src/lib/api-types/`.

6. **FE consumer** — `frontend/apps/web/src/features/templates/api/templates.ts` `approveVersion` returns `{ version: VersionDTO, nextDraft: { id, versionNumber } | null }`. `TemplateEditorPage.handleApprove` (or `VersionActionPanel` callback) reads `nextDraft` and calls `navigate('/templates/:id/versions/:n')` to the new draft. TanStack invalidation per [metaldocs-tanstack-query](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md): bust template list, template detail, version list, version(prev) → drives a fresh fetch when the user later returns.

7. **Toast** — show success "`Publicado. Nova versão de rascunho criada.`" so the user understands what just happened (per [common/coding-style.md](../../../C:/Users/leandro.theodoro/.claude/rules/common/coding-style.md) error handling: explicit at boundaries).

### TDD

- Backend contract test: Approve as the bound approver on a no-reviewer template → response carries `next_draft.id` matching the row Postgres just inserted; `next_draft.version_number == version.version_number + 1`.
- Backend table test for the reviewer path (`pending_reviewer_role != null`, status flips `in_review → approved` without publish): `next_draft` is `null`.
- FE vitest: mock the new response shape and assert `navigate` is called with the correct path.

### Close-out

- `scripts/check-module-contract-sync.ps1 -Module templates`
- FE: `corepack pnpm tsc --noEmit -p tsconfig.build.json` + `vitest run src/features/templates`
- Preview drive: submit → approve → land on the new draft URL; refresh re-entry on the new draft honest.
- Wiki bump: `templates.md` §6.3 lifecycle diagram; close hard-stop note in this plan.

### Status (2026-05-31) — CLOSED

- OpenAPI: `ApproveTemplateVersionResponse` added (bundled + partial) with nullable `next_draft`.
- Backend: `Service.Approve` returns `*ApproveResult{Version, NextDraft}`; spawns v(n+1) in same tx for both with-reviewer and no-reviewer accept paths; handler emits `data.next_draft = { id, version_number } | null`.
- Backend tests: `TestApprove_Accept_WithReviewer`, `TestApprove_Accept_NoReviewer`, `TestApprove_Reject` updated to assert NextDraft populated/nil; `TestApprove_Accept_Happy` and new `TestApprove_Reject_NextDraftNull` HTTP contract tests. `go test ./internal/modules/templates/...` ✓ (race requires cgo; gcc unavailable in this Windows env).
- FE: `approveVersion` returns `{ version, nextDraft }`; `VersionActionPanel` forwards nextDraft + "Publicado. Nova versão de rascunho criada." toast; `TemplateEditorPage` calls `onNavigateToVersion(templateId, nextDraft.versionNumber)`. `pnpm tsc --noEmit -p tsconfig.build.json` ✓. `vitest run src/features/templates` 37 pass / 5 pre-existing skipped (new `VersionActionPanel.nextDraft.test.tsx` ✓).
- Contract sync: `check-module-contract-sync.ps1 -Module templates` — only pre-existing `frontend handwritten type drift` (TemplateDTO/VersionDTO/… wrappers already on main) bumped by `NextDraftRef` + `ApproveVersionResult`; not a regression.
- Deferred: preview drive (no preview running; this branch ships next; verify on QA pass).

---

## Workstream 3 — `fix/templates-schema-occ-lock`

> Closes the residual half of [T-010](../../../wiki/modules/templates-tech-debt.md). Brings placeholder-schema saves to the same lock_version OCC discipline as draft DOCX saves. No more silent last-write-wins between two tabs.

**Skill route.** [`metaldocs-backend-api`](../../../.agents/skills/metaldocs-backend-api/SKILL.md) + [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md).

**Problem.** FE `putTemplateSchemas` hardcodes `expected_content_hash: ''` ([api/templates.ts:403](../../../frontend/apps/web/src/features/templates/api/templates.ts#L403)) — backend accepts empty as "skip CAS", silently. Two tabs editing placeholders concurrently: last writer wins, no error, no audit trail of the lost edit.

**Architecture.** `SaveTemplateDraft` already uses `UpdateVersionDraftCAS` with `expected_lock_version` ([templates.md §8.6](../../../wiki/modules/templates.md)). Schema PUT should follow the same pattern. CAS token is `lock_version`, not `content_hash`, for schema — schema is a structured field, not a blob.

### Plan

1. **OpenAPI partial** — request body for `PUT /api/v1/templates/{id}/versions/{n}/schema`:
   ```yaml
   PutSchemaRequest:
     type: object
     required: [placeholder_schema, metadata_schema, expected_lock_version]
     properties:
       placeholder_schema: { ... }
       metadata_schema: { ... }
       expected_lock_version: { type: integer, minimum: 0 }
   ```
   Remove `expected_content_hash` from the schema-PUT contract (legacy). Bump major version note in the partial header.

2. **Backend** — `internal/modules/templates/application/schema.go`: `UpdateSchema` takes `expectedLockVersion`, calls `repo.UpdateVersionSchemaCASTx(tenant, version, expectedLockVersion)`. Returns `ErrStaleLockVersion` on CAS miss. Repo method mirrors `UpdateVersionDraftCAS`.

3. **Problem mapping** — `ErrStaleLockVersion` already maps to RFC 9457 `code: "stale_lock_version"` 409 ([codes.go](../../../internal/platform/problem/codes.go)). Verify on this path.

4. **Regen** + **FE types**.

5. **FE** — `useTemplateSchemas`:
   - Hold `lockVersion` alongside `schemas` in state.
   - `save(s)` sends `expected_lock_version: lockVersion`.
   - On 409 `stale_lock_version`: surface alert "`Outro editor alterou os placeholders. Recarregue para ver as alterações.`" + offer `refetch()` (don't auto-overwrite — let the user decide).
   - On success, bump local `lockVersion` from the response (backend should return the new version DTO).

### TDD

- Backend table tests: stale lock → 409; correct lock → 200 + lock bumped; missing field → 400.
- Backend contract test: schema PUT with `expected_lock_version` mismatch → RFC 9457 `code: "stale_lock_version"`.
- FE vitest: two simultaneous saves; second one mocked to 409 → user sees the alert; no silent overwrite.

### Close-out

- Wiki: bump `templates.md` §8.6 concurrency table; flip T-010 to CLOSED with this branch listed.

---

## Workstream 4 — `chore/fe-capability-gates-templates`

> Pre-disable action buttons by capability + role-binding so the user never clicks an action they can't complete. Defense-in-depth — backend remains the sole enforcer ([web/security.md](../../../C:/Users/leandro.theodoro/.claude/rules/web/security.md), [authz-tiers.md](../../../wiki/concepts/authz-tiers.md)).

**Skill route.** [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md) + [`metaldocs-tanstack-query`](../../../.agents/skills/metaldocs-tanstack-query/SKILL.md).

**Problem.** `TemplateEditorPage` and `VersionActionPanel` gate buttons by `version.status` only. A `system_admin` whose role doesn't satisfy `pending_approver_role` sees Publish, clicks it, and learns of the denial via a raw error code. Acceptable for security; UX bug.

**Architecture.** Server is the truth. FE pre-gate is a hint, not a guarantee. The hint comes from a lightweight `useMe()` capability/role surface that is cache-friendly and tenant-scoped.

### Plan

1. **Verify** an authz surface already exists: grep `useAuthSession`, `useMe`, `/auth/me`, `/api/v1/auth/capabilities`. If only `roles` is returned, extend `/auth/me` to also return `capabilities[]` (cap codes from `iamdomain`). If capabilities aren't already exposed, that's a tiny backend addition under the same skill route.

2. **Helper** — `frontend/apps/web/src/features/templates/lib/canActOnVersion.ts`:
   ```ts
   export type VersionActionGate = { allowed: boolean; reason?: string };
   export function canSubmit(v: VersionDTO, me: Me): VersionActionGate { ... }
   export function canApprove(v: VersionDTO, me: Me): VersionActionGate { ... }
   export function canReview(v: VersionDTO, me: Me): VersionActionGate { ... }
   ```
   - `allowed` = status valid + capability present + (role binding satisfied | binding null).
   - `reason` = Portuguese explanation for tooltip ("Você não possui o papel necessário ({role})." / "Sua sessão não inclui a capacidade {cap}.").

3. **Consumers** — `TemplateEditorPage` `isDraft` button + `VersionActionPanel` Approve/Reject/Publish: pass `disabled={!gate.allowed}` and `title={gate.reason}`. Don't hide the buttons — disabled+tooltip teaches the user what they need.

4. **Vitest** — gate function unit tests cover the capability × role × status matrix (table-driven). Plus a component test asserting tooltip surfaces when disabled.

### Close-out

- Preview drive matrix: `admin` / `reviewer` / `approver` / `author` against draft / in_review / approved. Confirm correct buttons are enabled for each.

---

## Workstream 5 — `chore/fe-error-message-coverage`

> Map every templates-domain backend code to a Portuguese user-facing string. No more raw `templates: forbidden_role` leaking into alerts.

**Skill route.** [`metaldocs-frontend`](../../../.agents/skills/metaldocs-frontend/SKILL.md).

**Problem.** `resolveErrorMessage(code, fallback)` in `frontend/apps/web/src/lib/api/errorMessages.ts` has gaps. Unknown codes fall through to the backend message (often a developer string) or the generic `"Erro interno"`. Once Workstream 1 adds `forbidden_role` on Publish and Workstream 3 surfaces `stale_lock_version`, the gap grows.

### Plan

1. **Inventory** — grep `internal/platform/problem/codes.go` and per-module `domain/errors.go` for every code. Generate a canonical list.

2. **Map** — extend `errorMessages.ts` with one Portuguese entry per code. Group by module. Examples:
   - `idempotency_key_required` → "Identificador de operação ausente. Recarregue e tente novamente."
   - `forbidden_role` → "Você não possui o papel necessário para esta ação."
   - `stale_lock_version` → "Outro usuário alterou este recurso. Recarregue para continuar."
   - `idempotency_in_flight` → "Operação anterior ainda em andamento. Aguarde a confirmação."
   - `content_hash_mismatch` → "O conteúdo do arquivo mudou desde o carregamento. Reenvie o documento."

3. **Fallback** — generic fallback message changes from `"Erro interno"` to:
   ```
   "Não foi possível concluir a ação. Código: {code}"
   ```
   Gives Support an actionable identifier without scaring the user. Add a one-liner Sentry breadcrumb when the fallback fires so unmapped codes are visible in observability.

4. **Lint** — Vitest test that imports the codes file (or a generated JSON of codes) and asserts every code has a Portuguese entry. Fails CI if a new code lands without a mapping.

### Close-out

- Preview spot-check on a few flows that previously surfaced raw codes.

---

## Cross-cutting principles applied (so this aligns with the wiki)

- **Contract-first.** Every backend-surface change (1, 2, 3) starts in OpenAPI partials, then `go generate` + `pnpm gen:api`. No hand-rolled types ([0012-contract-first-api.md](../../../wiki/decisions/0012-contract-first-api.md)).
- **Two-tier authz everywhere.** Capability and role-binding gates both run, in this order, inside the same tx. Postgres tripwire backs the capability tier ([0007-two-tier-authz.md](../../../wiki/decisions/0007-two-tier-authz.md)).
- **RFC 9457 problem+json.** All error responses already on this contract ([api-design-system.md](../../../wiki/architecture/api-design-system.md)). FE message mapping (Workstream 5) is the last mile.
- **OCC via `lock_version`.** Schema PUT joins the existing draft-save discipline; eigenpal legacy autosave commit remains hash-gated (per T-010 residual scope note).
- **Idempotency-Key on every mutation.** Already enforced by backend middleware; Workstream context ensures every FE call honors it (already shipped this branch).
- **TDD first.** Contract test → service test → handler test → FE test, all before the implementation per [development-workflow.md](../../../C:/Users/leandro.theodoro/.claude/rules/common/development-workflow.md).
- **Evidence rule.** No close-out without runtime + persisted/API proof per [qa-operating-system.md](../../../wiki/quality/qa-operating-system.md). The QA report on this branch is the template.
- **No symptom patching.** Each workstream fixes at the owning boundary. The button-disable FE patch (4) is acceptable *because* the backend gate (1) is fixed in the same plan — UX defense-in-depth on top of a correct enforcer, not in place of one.

---

## Out of scope (intentionally deferred)

- `editable_zones` column purge ([T-012](../../../wiki/modules/templates-tech-debt.md)) — DDL hygiene, no behavior impact.
- Eigenpal legacy `/autosave/commit` lock_version retrofit ([T-010 residual note](../../../wiki/modules/templates-tech-debt.md)) — touches the third-party-compatible import contract; separate plan.
- `ListTemplates` pagination ([T-011](../../../wiki/modules/templates-tech-debt.md)) — minor; not editor-screen.
- Exported-symbol Go docs ([T-014](../../../wiki/modules/templates-tech-debt.md)) — hygiene sweep, separate.

---

## Definition of done for the whole plan

- Five branches merged into `main`. Each closed via the QA close-out checklist with evidence.
- `wiki/modules/templates.md` `Last verified` reflects the latest merge.
- `wiki/modules/templates-tech-debt.md`: T-004 closed; T-009 closed (FE half); T-010 residual list updated.
- `wiki/backlog/templates-refactor.md`: R-004 / R-009 / R-010 moved to "closed".
- A follow-up QA pass on `qa/templates-editor` re-validates the editor end-to-end on a fresh draft → submit → approve → publish → new draft, with full Preview evidence.
