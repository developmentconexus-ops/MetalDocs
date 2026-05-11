# Refactor Plan Prompts

> Each section = copy-paste prompt for a fresh session.
> Always read `wiki/backlog/roadmap.md` first in every session — it is the source of truth for status + scope.
> Execution order: 3 → 4 (‖ 11) → 5 (‖ 6) → 7 (‖ 8) → 9 → 10 → 12 (‖ 13)

## Execution model (applies to every plan)

| Step | Who |
|------|-----|
| Writing implementation spec | Sonnet — `nexus:writing-plans` |
| Coding — implementing Workstreams | **Codex** — `codex:rescue` skill |
| Writing + running tests | **Sonnet or Haiku** — after Codex returns |
| Commits | **Sonnet or Haiku** — review diff, then commit |
| Final review | **Opus** only — once per Plan (all PRs grouped), not per PR |
| Roadmap status update | **Sonnet** |

**Rule:** Codex implements. Sonnet/Haiku own tests + commits. Opus reviews once per Plan after all workstreams are committed — not after each PR. Token efficiency: one Opus session reviews the whole plan diff together via `nexus:requesting-code-review`.

---

## Plan 4 · Capability namespace collapse + IAM dual-surface consolidation

```
Mode: implementation. Plan 4 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 4. Note anchor decision: typed iamdomain.Capability wins.
2. wiki/README.md
3. CLAUDE.md
4. wiki/modules/iam-tech-debt.md (T-001, T-002, T-003, T-009, T-012)
5. wiki/backlog/iam-refactor.md (R-001..R-003, R-009, R-012)
6. wiki/modules/documents-tech-debt.md (T-008) + wiki/backlog/documents-refactor.md (R-008)
7. wiki/modules/iam.md §5 (surfaces + DI)
8. wiki/concepts/authz-tiers.md
9. wiki/decisions/0007-two-tier-authz.md

Goal: one capability namespace, one area-membership write surface, no dead authz surface.

Workstream A — Capability namespace collapse (T-001 / R-001):
- Canonical: typed Capability consts in internal/modules/iam/domain/model.go:16
  (CapDocumentView, CapDocumentCreate, CapDocumentEdit, CapWorkflowReview, CapWorkflowApprove).
- Delete: internal/modules/iam/domain/capabilities.go (the 16 string consts doc.view / doc.create / …).
- DB seed: write migration <next>_role_capabilities_typed.sql that repopulates
  metaldocs.role_capabilities using the typed names (e.g. "document.view" not "doc.view").
- Consumer fanout: update every call site that passed a string cap literal to use the typed const:
  * internal/modules/documents/application/fillin_authz.go:9
  * internal/modules/documents/approval/application/submit_service.go:85 ("doc.submit" string)
  * any other grep hits for "doc.", "template.", "registry.", "taxonomy.", "membership.", "route.", "user." string cap literals.
- Also fix: ErrCapabilityDenied dual definition (T-009 / R-009):
  * internal/modules/iam/application/capability_service.go:10 (sentinel)
  * internal/modules/iam/authz/authz.go:11 (typed struct)
  * Pick one shape, consolidate, fix consumer at internal/modules/documents/delivery/http/handler.go:17.
- Also fix: RoleCapabilities in-process map (T-012 / R-012) — decide: keep in-process map and keep
  it in sync with DB types, OR remove map and drive CanDo fully from DB. Recommend removing the map
  if AuthorizationService (see Workstream C) is also deleted; document the decision as an ADR stub.

Workstream B — Area-membership dual-surface consolidation (T-002 / R-002):
- Two surfaces: internal/modules/iam/area_membership/area_membership.go (SECURITY DEFINER funcs)
  vs internal/modules/iam/application/area_membership_service.go (direct DML).
- Pick one. Recommend: keep the application-service + repo path (DML, no SECURITY DEFINER),
  because the DEFINER path emits governance events through SQL whereas the service path
  calls govLogger — making the service path composable with Plan 6 audit work.
  If you pick DEFINER path, state reason clearly in a code comment.
- Delete or rename the losing surface. Update DI in apps/api/cmd/metaldocs-api/main.go.

Workstream C — AuthorizationService: wire or delete (T-003 / R-003):
- internal/modules/iam/application/authorization.go:42 is not wired in main.go.
- Decision: if typed Capability consolidation (Workstream A) makes AuthorizationService
  the clean home for resource-aware SoD checks, wire it. Otherwise delete it.
- Recommend: delete — Plan 5 will wire tier-2 authz.Require per-module, which covers
  the resource-ctx use case without a third surface. Document deletion rationale in a
  // removed: reason comment on the call site and a TODO-ADR stub.

Process:
1. Use nexus:writing-plans to author implementation plan at
   docs/superpowers/specs/2026-05-NN-plan-04-capability-namespace.md.
2. Confirm plan before executing.
3. After implementing, run existing test suite (go test ./...). Fix any broken tests.
   Do NOT write new tests for code you are deleting.
4. Start local API (.\scripts\start-api.ps1 -Build), login, exercise one
   capability-gated route (e.g. POST /api/v1/iam/users/{id}/roles) to confirm
   tier-1 middleware still gates correctly with renamed caps.
5. Dispatch wiki-curator after each PR.
6. On completion: update wiki/backlog/roadmap.md Plan 4 status → done, link PRs,
   mark closed rows in iam-tech-debt.md + documents-tech-debt.md + backlogs.

/simplify rules:
- Do NOT expand scope to add tier-2 authz.Require (that is Plan 5).
- Do NOT refactor unrelated IAM handlers. Touch only what Workstreams A/B/C require.
- Surgical changes only. Every changed line traces to a Workstream task.
- Missing-ADR stubs: one-liner // ADR-TODO: <topic> in code is fine. Full ADR docs go in Plan 13.

Push back if I try to:
- Bundle tier-2 authz.Require additions (Plan 5).
- "Clean up" unrelated IAM handler code while you are in the file.
- Keep both capability namespaces "for backwards compat" — one namespace is the deliverable.
- Start Plan 5 before Plan 4 is committed and green.
```

---

## Plan 11 · Editor frontend stabilization (parallel to Plan 4)

```
Mode: implementation. Plan 11 of the MetalDocs refactor roadmap (runs parallel to Plan 4).

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 11.
2. wiki/README.md
3. CLAUDE.md (note /simplify and frontend skill rules)
4. wiki/modules/editor-ui-eigenpal-tech-debt.md (T-002, T-003, T-004, T-005, T-006, T-007, T-008)
5. wiki/backlog/editor-ui-eigenpal-refactor.md (R-002..R-010)
6. wiki/modules/editor-chrome-tech-debt.md (T-001..T-009)
7. wiki/backlog/editor-chrome-refactor.md (R-001..R-009)
8. wiki/modules/editor-ui-eigenpal.md (ACL contract, plugin gating, autosave debounce)
9. wiki/modules/editor-chrome.md (slot API, AutosaveStatus states)
10. wiki/decisions/0001-eigenpal-adoption.md

Prerequisite: Plan 3 done (tarball restored). Confirm before executing Workstream A.

Goal: ACL wrapper enforced on all eigenpal consumers. Editor-chrome tested, token-driven, a11y-correct.

Workstream A — TemplateEditorPage → MetalDocsEditor adapter (T-002 / R-002):
- frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:4 currently imports
  DocxEditor directly from @eigenpal/docx-js-editor/react. Must import MetalDocsEditor from
  @metaldocs/editor-ui instead. Ref type changes to MetalDocsEditorRef.
- After migration: verify autosave debounce (1500ms), inFlightRef guard, mode="template-draft"
  plugin gating all apply correctly.
- Do NOT add new props or behaviors beyond what DocumentEditorPage already uses via the adapter.

Workstream B — Stale wiring test (T-003 / R-003):
- packages/editor-ui/test/templatePlugin.wiring.test.tsx:29 currently expects templatePlugin
  included on mode="document-edit". Current code gates it to mode="template-draft" only.
- Rewrite test to assert: mode="template-draft" → plugin present; mode="document-edit" → plugin absent.
- Run: cd packages/editor-ui && npm test (or pnpm test) to confirm green.

Workstream C — Editor-chrome R-001..R-009:
- R-001: Widen AutosaveStatus exported union from 4 to 7 states matching
  documents/hooks/v2/useDocumentAutosave.ts AutosaveStatus (add dirty, stale, session_lost).
  Update DocumentEditorPage adapter at pages/DocumentEditorPage.tsx:184 — remove ternary collapse.
- R-002: Add role="status" aria-live="polite" to AutosaveStatus wrapper span.
- R-003: No code change — add a CI snapshot test or comment guard that names the
  eigenpal version (0.2.0) so any version bump triggers a review. Minimal.
- R-004: Write RTL tests for EditorChrome: slot rendering (left/center/right/alert), 
  autosave state branches, null-slot collapse. Co-locate in editor-chrome/ as EditorChrome.test.tsx.
- R-005: Replace magic px values in EditorChrome.module.css with CSS tokens where tokens exist
  (--sp-*, --r-*). Do NOT invent new tokens; skip items where no token maps.
- R-006..R-009: Minor cleanup — typed style re-export, pointer-events JSDoc, slot truthy-collapse
  note in JSDoc. One-liners each.

Process:
1. Use nexus:writing-plans to author docs/superpowers/specs/2026-05-NN-plan-11-editor-frontend.md.
2. Confirm plan, then execute using metaldocs-frontend skill per CLAUDE.md for any frontend work.
3. Start dev server (frontend), exercise TemplateEditorPage and DocumentEditorPage manually:
   confirm autosave indicator states + plugin gating + slot rendering look correct.
4. Run full test suite: cd packages/editor-ui && npm test; cd frontend && npm test (vitest).
5. Dispatch wiki-curator after each PR.
6. On completion: update wiki/backlog/roadmap.md Plan 11 status → done, link PRs, mark closed
   rows in both tech-debt registers + backlogs.

/simplify rules:
- Do NOT add new eigenpal features. ACL enforcement + chrome polish only.
- Do NOT redesign the EditorChrome API (slot shape stays). R-001 widens a type, not an API.
- R-004 tests: RTL only, no E2E Playwright for this round.
- Skip R-005 token items where no existing token maps cleanly — leave a // TODO:token comment.

Push back if I try to:
- Redesign EditorChrome slot API.
- Add eigenpal 0.3.x features.
- Merge Plan 11 scope with Plan 12 screen implementation work.
```

---

## Plan 5 · Tier-2 `authz.Require` + Postgres tripwire on regulated tables

```
Mode: implementation. Plan 5 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 5. Confirm Plans 3 + 4 are done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/concepts/authz-tiers.md — two-tier model, GUC setup, pitfalls.
5. wiki/decisions/0007-two-tier-authz.md
6. wiki/modules/iam-tech-debt.md (T-004) + backlog R-004
7. wiki/modules/documents-tech-debt.md (T-003) + backlog R-003
8. wiki/modules/registry-tech-debt.md (T-001, T-004) + backlog R-001, R-004
9. wiki/modules/taxonomy-tech-debt.md (T-003, T-006, T-013) + backlog R-003, R-006, R-013
10. wiki/modules/templates_v2-tech-debt.md (T-001, T-002, T-004) + backlog R-001, R-002, R-004
11. Read migrations/0142b_role_capabilities_v2_enforce.sql — this is the approval-module tripwire
    pattern. Replicate it for new tables.

Prerequisite: Plan 4 done (canonical capability namespace in place — needed for tripwire GUC cap names).

Goal: every regulated mutating table protected by tier-1 cap middleware + tier-2 authz.Require +
enforce_capability_asserted Postgres trigger. Matches the approval-module pattern.

Tables to cover (currently unprotected):
- IAM: iam_user_roles, user_process_areas, iam_users (T-004)
- documents: public.documents (T-003)
- registry: controlled_documents, cd_sequence_counters (T-001, T-004)
- taxonomy: document_profiles, document_process_areas, document_families (T-003, T-006)
- templates_v2: templates_v2_template, templates_v2_template_version (T-001, T-004)

Per-module tasks:

IAM (T-004 / R-004):
- Add authz.Require(ctx, tx, CapAdminManage, "") call before mutations in
  infrastructure/postgres/role_admin_repository.go:33,72 and user_area_repository.go:51,75,90.
- Extend enforce_capability_asserted trigger to iam_user_roles + user_process_areas tables via new migration.

documents (T-003 / R-003):
- Add authz.Require to repository/repository.go:73,216,428,1071,1082 before each mutation.
- Extend trigger to public.documents table.

registry (T-001, T-004 / R-001, R-004):
- Add authz.Require to infrastructure/repository.go:133,137,184,208,239.
- Wire capability check in Obsolete + Supersede handlers (routes.go:328,337): determine
  which cap applies (registry.obsolete / registry.supersede — seed in migration if absent).
- Extend trigger to controlled_documents + cd_sequence_counters.

taxonomy (T-003, T-006, T-013 / R-003, R-006, R-013):
- Fix PATCH families gap: add MethodPatch to capability dispatcher at
  apps/api/cmd/metaldocs-api/permissions.go:174.
- Add authz.Require in taxonomy service call sites (profile, area, family mutations).
- Add DB trigger to document_profiles + document_process_areas + document_families.
- Add DB-level code-immutability trigger on document_families.code (mirrors 0122/0123 pattern for profiles/areas).

templates_v2 (T-001, T-002, T-004 / R-001, R-002, R-004):
- Wire real AuthzFunc in apps/api/cmd/metaldocs-api/main.go:329 (currently passes nil).
- Add authz.Require inside service methods for create/submit/approve/publish.
- Fix PublishTemplateVersion (lifecycle.go:265): add domain.CheckSegregation + content_hash gate.
- For T-002 (GetVersionByID cross-tenant): add tenant_id predicate to both repository getters or
  validate tenant binding at service level before the call.
- Extend trigger to templates_v2_template + templates_v2_template_version tables.

Migration strategy: one migration per module group or one consolidated migration for all new
trigger attachments — choose based on review atomicity. Keep each migration reversible.

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-05-tripwire.md.
2. Confirm plan.
3. Run go test ./... before + after each PR. Tests must stay green.
4. Start API (.\scripts\start-api.ps1 -Build). Login. Exercise: create document, submit for review,
   attempt an unauthorized mutation (expect 403). Confirm tripwire fires on direct-DB access attempt.
5. Dispatch wiki-curator. Update roadmap.md.

/simplify rules:
- Do NOT refactor service or handler code beyond adding authz.Require + fixing T-001/T-002/T-004.
- Migrations: one trigger template, replicate. Do not over-generalize.
- If a cap name doesn't exist yet (e.g. registry.obsolete), add it to the seed migration here —
  do NOT create a whole new capability design.

Push back if I try to:
- Start before Plans 3 + 4 are confirmed done.
- Add RFC 9457 error envelope (Plan 7).
- Refactor repository signatures beyond adding tx param where needed for authz.Require.
```

---

## Plan 6 · Audit-trail completeness sweep + audit-module hardening

```
Mode: implementation. Plan 6 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 6. Confirm Plans 3 + 4 done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/modules/audit.md + wiki/modules/audit-tech-debt.md (all 12 items)
5. wiki/backlog/audit-refactor.md
6. wiki/modules/auth-tech-debt.md (T-002) + backlog R-002
7. wiki/modules/iam-tech-debt.md (T-005) + backlog R-005
8. wiki/modules/registry-tech-debt.md (T-002, T-008) + backlog R-002, R-008
9. wiki/modules/taxonomy-tech-debt.md (T-004, T-005, T-010) + backlog R-004, R-005, R-010
10. wiki/modules/documents-tech-debt.md (T-005) + backlog R-005
11. wiki/modules/templates_v2-tech-debt.md (T-013) + backlog R-013

Prerequisite: Plans 3 + 4 done (tenant resolution + cap namespace).

Goal: one canonical sink (metaldocs.audit_events). Every regulated mutation emits to it.
Audit module hardened: gated read, tenant_id, tamper-evidence, retention.

Split rule: if tamper-evidence (hash chain) scope balloons (own migration + validation job +
ops runbook), split as Plan 6b and ship Plan 6a (emission completeness + sink consolidation +
tenant_id + retention policy + gated read) first. Decide after drafting the spec.

Workstream A — Audit module hardening (T-001, T-003, T-004, T-005, T-007, T-010 / R-001..R-010):
- T-001: Gate GET /api/v1/audit/events — add capability rule in permissions.go.
- T-003: Add retention strategy — simplest option: range partitioning by created_at or a
  scheduled DELETE with configurable retention_days env var. Pick whichever is smaller code.
- T-004 (tamper-evidence): Add prev_hash + row_hash columns via migration. Compute on write
  in internal/modules/audit/infrastructure/postgres/writer.go:20. Add read-path hash-chain
  validator (internal function, not HTTP). If this is > 1 day of work, defer to Plan 6b.
- T-005: Replace fire-and-forget _ = h.audit.Record(...) at all call sites with error propagation
  (log + metric counter at minimum; propagate 500 on Write failure only on critical mutations).
- T-007: Add tenant_id column to metaldocs.audit_events via migration. Populate from context.

Workstream B — Emission gaps: every regulated mutation now emits:
- auth: service.go:117 (Login), :279 (Logout? — check), :305 (CreateUser), password-change.
  Wire auditdomain.Writer into auth Service constructor if not already wired.
- iam: admin_handler.go:316 (handleUserRoleUpsert) — call recordAudit (pattern matches :457).
- registry: service.go:293 (changeStatus / Obsolete + Supersede) — add govLogger.Log calls.
- taxonomy: family_service.go:11 — add govLogger field + call on Create/Update/Deactivate.
  profile_service.go:41,55 — add Log calls on Create + Update (Archive already emits).
  area_service.go — same for Create + Update.
- documents: service.go:575 RenameDocument — move audit write inside the SQL UPDATE transaction
  (wrap both in explicit tx).

Workstream C — Sink consolidation:
- templates_v2 T-013: Switch repository/postgres.go:318 AppendAudit from templates_v2_audit_log
  to auditdomain.Writer port. Wire the writer in main.go. Drop templates_v2_audit_log writes
  (keep table for now — drop in Plan 10 cleanup).
- taxonomy T-010 + registry T-008: Switch governance_events-writing GovernanceLogger to emit
  via auditdomain.Writer instead. Registry module.go:31 must stop importing taxonomyapp logger —
  wire its own audit writer directly.

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-06-audit.md. Include 6a/6b split
   decision in the spec.
2. Confirm plan + split decision before executing.
3. go test ./... before + after each PR.
4. Start API. Login as admin. Perform role upsert, document rename, obsolete a registry entry.
   Query GET /api/v1/audit/events (now gated — use admin session). Confirm rows present.
   Attempt GET without auth — expect 401/403.
5. Dispatch wiki-curator. Update roadmap.md.

/simplify rules:
- Retention: pick the smaller implementation (config-driven DELETE job or partitioning) — do not
  over-design. A simple DELETE WHERE created_at < now() - interval '$DAYS days' job is fine.
- Tamper-evidence: SHA-256 hash chain only. No asymmetric signing this round — that is infra not in place.
- Do NOT add OpenTelemetry / Prometheus metrics unless already wired. Log + structured error is enough.
- Do NOT migrate to RFC 9457 envelope (Plan 7).

Push back if I try to:
- Add distributed tracing or metrics pipelines.
- Implement WORM storage or external audit mirror (not in roadmap).
- Start Plan 7 work (envelope) while Plan 6 is in flight.
```

---

## Plan 7 · RFC 9457 envelope rollout

```
Mode: implementation. Plan 7 of the MetalDocs refactor roadmap (can run parallel to Plan 8).

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 7. Confirm Plans 3–6 done (or 6a at minimum).
2. wiki/README.md
3. CLAUDE.md
4. wiki/architecture/api-design-system.md — RFC 9457 contract + the internal/platform/problem package.
5. wiki/concepts/error-ux.md — frontend ApiError / resolveErrorMessage / auth-bus.
6. Per-module tech-debt rows: iam T-006, documents T-001, auth T-003, approval T-001+T-003,
   audit T-002, templates_v2 T-005, registry T-003, taxonomy T-008.
7. Per-module backlog rows: R-006 (iam), R-001 (documents), R-003 (auth), R-001+R-003 (approval),
   R-002 (audit), R-005 (templates_v2), R-003 (registry), R-008 (taxonomy).
8. Grep internal/platform/problem/ to confirm the helper exists before starting.

Goal: application/problem+json on every non-2xx response. Frontend ApiError parser updated.

Per-module changes (one PR each):
- iam: replace middleware.go:129 writeAPIError + routes_memberships.go:137 writeMembershipAPIError
  with httpresponse.WriteProblem (or platform/problem helper).
- documents: replace handler.go:958 mapErr + :1013 httpErr.
- auth: replace handler.go:166 + middleware.go:65,76,79,83 error writers.
- approval: replace errors.go:147 WriteError. Fix T-003: replace looksLikeValidationError substring
  classifier with typed domain error matching → correct 409 state.instance_completed.
- audit: replace handler.go:48,60,97 error writers.
- templates_v2: replace handler.go:95 writeErr + errors.go:10 MapErr.
- registry: replace httpresponse/response.go:14 usages in routes.go. Add missing 422 template_invalid
  branch in writeDomainError.
- taxonomy: replace writeFamilyError + writeProfileError + area error helpers.
- frontend (one PR): update frontend/apps/web/src/lib/api/ ApiError parser to read
  type/title/status/detail/instance from application/problem+json. Update approval's ApprovalError
  (features/approval/api/mutationClient.ts:9). Keep backward-compat parsing if needed during rollout
  (Content-Type sniff: if problem+json use new parser; else legacy).

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-07-rfc9457.md.
2. Confirm plan.
3. Per module: go test ./... stays green. No handler logic changes — only error-response shape.
4. Start API. Trigger a 400, 403, 404, 422, 500 on each migrated module. Confirm
   Content-Type: application/problem+json + body has type/title/status/detail fields.
5. Frontend: after backend PRs merged, run frontend dev server. Confirm error toasts
   still display correct messages (resolveErrorMessage parses new shape).
6. Dispatch wiki-curator. Update roadmap.md.

/simplify rules:
- Do NOT change business logic in any handler. Shape of error response only.
- Do NOT write new error types beyond what problem+json requires.
- Backward-compat parsing in frontend: one Content-Type sniff, not a full version-negotiation system.
- Do NOT add new error codes beyond fixing approval T-003 (state.instance_completed).

Push back if I try to:
- Change handler business logic while "in the file fixing envelopes."
- Add new API error codes not listed in Closes rows.
- Start Plan 8 (OpenAPI) before Plan 7 is green — spec needs the envelope schema.
```

---

## Plan 8 · OpenAPI / contract-first completion

```
Mode: implementation. Plan 8 of the MetalDocs refactor roadmap (can start after Plan 7 is green).

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 8. Confirm Plan 7 done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/architecture/api-contract.md — spec-as-source-of-truth, oapi-codegen v2, migration status.
5. wiki/decisions/0012-contract-first-api.md
6. references/oapi-codegen.md — how to regenerate, vendor-mode, include-tags.
7. wiki/modules/documents-tech-debt.md (T-002, T-004, T-010) + backlog R-002, R-004
8. wiki/modules/approval-tech-debt.md (T-002) + backlog R-002
9. wiki/modules/templates_v2-tech-debt.md (T-006) + backlog R-006
10. wiki/modules/registry-tech-debt.md (T-007, T-011) + backlog R-007, R-011
11. wiki/modules/audit-tech-debt.md (T-008) + backlog R-008
12. wiki/modules/taxonomy-tech-debt.md (T-009) + backlog R-009

Prerequisite: Plan 7 done (problem+json schema authored — referenced by spec error responses).

Goal: every HTTP route in openapi spec. Codegen wired. Raw mux.HandleFunc removed per module.

Per-module tasks (one PR each):

documents (T-002, T-004, T-010 / R-002, R-004):
- Add missing ops to api/openapi/v1/openapi.yaml: renameDocument, duplicateDocument,
  archiveDocument, comments CRUD, finalizeDocument operationId.
- Fix duplicate PATCH /api/v1/documents/{id} registration (handler.go:86 + :115).
- Regenerate internal/modules/documents/api/api.gen.go. Mount via ServerInterface.

approval (T-002 / R-002):
- Add to spec: POST /api/v1/documents/{id}/signoff, /signoffs, /cancel with request+response schemas.
- Generate stubs. Wire router.go to ServerInterface.

templates_v2 (T-006 / R-006):
- Add 12 missing routes to spec: /versions (POST), /schema (PUT), /submit, /review, /approve (POST×3),
  /autosave/presign, /autosave/commit, /archive, /approval-config (PUT), /docx-url (GET),
  /audit (GET), /placeholder-catalog (GET).
- Regenerate + wire handler.go to ServerInterface.

registry (T-007, T-011 / R-007, R-011):
- Add missing 422 template_invalid branch to spec response schema.
- Fix: move partial spec from api/openapi/v1/partials/registry.yaml to v1/ tree with correct
  path prefix documentation (routes are /api/v1/ per Plan 10 anchor).
- Regenerate.

audit (T-008 / R-008):
- Add operationId to GET /api/v1/audit/events path in openapi.yaml.
- Wire handler to generated ServerInterface (or at minimum add operationId so frontend codegen binds).

taxonomy (T-009 / R-009):
- Full spec authoring: 16 routes under /api/v1/taxonomy/* — families CRUD, profiles CRUD, areas CRUD.
- This is the largest sub-task: write request/response schemas for all 16 routes.
- Regenerate internal/modules/taxonomy/api/api.gen.go (new file). Wire handler.go to ServerInterface.

Regenerate frontend types after all backend specs merged:
- cd frontend && npm run generate (or equivalent openapi-typescript command).
- Confirm frontend/apps/web/src/lib/api-types/ updates cleanly.

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-08-openapi.md.
2. Confirm plan.
3. go test ./... after each module PR.
4. After each regeneration: api-lint (if wired) + confirm server starts cleanly.
5. Frontend codegen: verify generated types compile (npm run build or tsc --noEmit).
6. Dispatch wiki-curator. Update roadmap.md.

/simplify rules:
- Do NOT redesign handler logic while adding spec. Spec describes current behavior.
- Do NOT add new endpoints that do not already have a handler.
- Taxonomy spec: match current handler input/output shapes exactly, no new fields.
- Do NOT start Plan 9 (tx/idempotency) while Plan 8 is open — spec + codegen changes conflict.

Push back if I try to:
- Add new business logic while "adding routes to the spec."
- Redesign taxonomy API shape during spec authoring.
- Skip taxonomy (it's the hardest) — it must be completed, not deferred.
```

---

## Plan 9 · Transactional + idempotency hardening

```
Mode: implementation. Plan 9 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 9. Confirm Plans 7 + 8 done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/decisions/0011-cd-atomic-create.md — Stripe-style idempotency pattern already in platform.
5. wiki/modules/auth-tech-debt.md (T-004) + backlog R-004
6. wiki/modules/documents-tech-debt.md (T-006, T-009) + backlog R-006, R-009
7. wiki/modules/templates_v2-tech-debt.md (T-007, T-009, T-010) + backlog R-007, R-009, R-010
8. wiki/modules/taxonomy-tech-debt.md (T-007, T-011) + backlog R-007, R-011
9. Grep internal/platform/idempotency/ — confirm the platform helper from Plan 2 exists.

Goal: atomic multi-step writes. HTTP Idempotency-Key on POST create/mutate. Optimistic-lock enforced.

Workstream A — auth T-004 / R-004 (CreateUser atomicity):
- auth/application/service.go:305 CreateUser: wrap repo.CreateUser + roleAdmin.ReplaceUserRoles in
  one outer *sql.Tx. Both calls must use the same tx. No partial commit on failure.
- Tests: write test that fails TX-B and asserts auth_identities row is also rolled back.

Workstream B — documents T-006 / R-006 (finalize Idempotency-Key):
- handler.go:316: read Idempotency-Key header. Pass to submit_service.
- submit_service.go: store key in metaldocs.idempotency_keys table (use platform idempotency helper).
- On replay: return cached 201 not a new 409.
- documents T-009 / R-009 (placeholder FK): write migration fixing
  document_placeholder_values.revision_id REFERENCES to document_revisions(id).

Workstream C — templates_v2 T-007, T-009, T-010 / R-007, R-009, R-010:
- T-007: Wrap lifecycle.go:265 PublishTemplateVersion 3–5 ExecContext calls in a single *sql.Tx.
  Repository methods need a WithTx(tx) variant or accept *sql.Tx argument.
  Same for Service.Approve and Service.CreateTemplate where multi-step.
- T-010: lifecycle.go UpdateVersion — add WHERE lock_version = $X predicate.
  If 0 rows affected → return 409 PreconditionFailed. HTTP handler maps to 412.
- T-009: Read Idempotency-Key on POST /templates, /publish, /submit, /approve.
  Use platform/idempotency helper. Dedupe window: 24h.

Workstream D — taxonomy T-007 / R-007 (Deactivate atomicity + HasActiveProfiles tenant fix):
- family_service.go:48: wrap GetByCode + HasActiveProfiles + Update in one *sql.Tx
  with SELECT ... FOR UPDATE on the family row.
- HasActiveProfiles query: add tenant_id = $1 predicate so cross-tenant probe is closed.
- taxonomy T-011 / R-011: duplicate POST /profiles etc. returning 500 on PK violation —
  map 23505 uniqueness error to 409 conflict in writeProfileError (and family/area mirrors).

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-09-tx-idempotency.md.
2. Confirm plan.
3. For each workstream: write a failing test first (nexus:test-driven-development), then implement.
4. go test ./... green before PR.
5. Start API. Test: duplicate POST (expect cached 2xx or 409 not 500). Test concurrent autosave
   (simulate with two requests — expect 412 on stale lock_version). Test CreateUser failure midway.
6. Dispatch wiki-curator. Update roadmap.md.

/simplify rules:
- tx wrapping: pass *sql.Tx down, do not over-architect a UoW pattern.
- Idempotency: use the existing platform/idempotency helper. No new dedup infrastructure.
- Do NOT add pagination (minor debt) in this plan. Taxonomy T-012 stays open until Plan 13 or later.

Push back if I try to:
- Add new business features while "touching the tx boundary."
- Implement distributed saga / outbox pattern — not needed here.
- Bundle Plan 10 cleanup into Plan 9 PRs.
```

---

## Plan 10 · Legacy purge + rename sweep

```
Mode: implementation. Plan 10 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 10. Confirm Plans 4, 5, 6, 7, 8 done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/decisions/0002-zone-purge.md (editable_zones context)
5. wiki/modules/templates_v2-tech-debt.md (T-012) + backlog R-100, R-101
6. wiki/modules/approval-tech-debt.md (T-007, T-008, T-009, T-011) + backlog R-007..R-009, R-011
7. wiki/modules/registry-tech-debt.md (T-010) + backlog R-100, R-010
8. wiki/modules/auth-tech-debt.md (T-012) + backlog R-012
9. wiki/backlog/editor-ui-eigenpal-refactor.md (R-004, R-005, R-006)
10. wiki/modules/taxonomy-tech-debt.md (T-013, T-015) + backlog R-013, R-015
11. wiki/modules/documents-tech-debt.md + backlog R-100 (retire documents-v2.md stub)

Anchor decisions (locked 2026-05-11, must be applied here):
- Module dir: internal/modules/templates_v2/ → internal/modules/templates/
- URL prefix: all /api/v2/* → /api/v1/*. Every frontend call-site updated.
- Column: approval_instances.document_v2_id → document_id (with migration + constraint rename).

Workstream A — templates_v2 rename (R-100, R-101):
- Move internal/modules/templates_v2/ → internal/modules/templates/ (update all Go imports).
- URL: handler.go route registrations /api/v2/templates/* → /api/v1/templates/*.
- Update frontend/apps/web/src/lib/api/ references.
- Retire wiki/modules/templates-v2.md (predecessor doc) — delete file, remove from README.
- Retire wiki/modules/documents-v2.md stub (documents R-100) — same.
- After rename: confirm go build ./... and npm run build both pass.

Workstream B — URL prefix sweep (all /api/v2/* → /api/v1/*):
- Documents: /api/v2/documents/* → /api/v1/documents/*
- Registry: /api/v2/controlled-documents/* → /api/v1/controlled-documents/*
- Taxonomy: /api/v2/taxonomy/* → /api/v1/taxonomy/*
- Approval doc-scoped: /api/v2/documents/{id}/signoff etc. → /api/v1/documents/{id}/signoff
- Update api/openapi/v1/openapi.yaml paths accordingly.
- Update every frontend feature api call site and generated types.
- Update acceptance tests (wiki/tests/system-acceptance-test.md URLs) — or note they need update.

Workstream C — Column + schema cleanup:
- approval_instances.document_v2_id: write migration renaming column to document_id.
  Update UNIQUE constraint name. Update repository/postgres_approval_repository.go:36.
- templates_v2_template_version.editable_zones: write migration dropping the column (T-012 / R-012).
- registry R-100: drop or archive profile_sequence_counters legacy table if unused (confirm first).

Workstream D — Dead surfaces:
- editor-ui-eigenpal T-004 / R-004: delete createOutlinePlugin export from packages/editor-ui/src/index.ts.
  Delete OutlinePlugin.tsx if no other consumer (grep first).
- editor-ui-eigenpal T-005 / R-005: rename packages/editor-ui/src/plugins/mergefieldPlugin.ts to
  sidebarModelData.ts (or similar). Update import in sidebarModelBridge.ts.
- editor-ui-eigenpal T-006 / R-006: delete onLockLost prop from types.ts.
- auth T-012 / R-012: OriginProtection field in config.go:116 — wire it (add middleware check) or
  delete the config field + test. Simplest: delete if no enforcement plan exists.
- approval T-011 / R-011: collapse WithMembershipContext + setAuthzGUC to one helper.
  Delete the unused one. Update call sites.
- approval T-007 / R-007: rename infra/signature/ → infrastructure/signature/. Update imports.
- approval T-009 / R-009: write migration VALIDATE CONSTRAINT on NOT VALID FKs in approval_instances
  (submitted_by → iam_users, actor_user_id → iam_users).
- registry T-010 / R-010: expose repository via Module struct instead of constructing a second
  PostgresControlledDocumentRepository in main.go:224.
- taxonomy T-015 / R-015: write migration dropping redundant PK on code alone from document_profiles
  and document_process_areas, leaving only the (tenant_id, code) unique index as the key.

Workstream E — Naming:
- taxonomy T-013 / R-013: add DB BEFORE-UPDATE trigger on document_families.code (mirror 0122/0123).
  [Note: may already be done in Plan 5 — check first; skip if closed.]
- iam T-010 / R-010: auth↔IAM bidirectional dep — document as minor ADR stub (not a code change;
  falls to Plan 13 for full ADR; add // NOTE: non-circular today, see T-010 comment at import site).

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-10-purge-rename.md.
   Identify which Workstream items can be batched per PR (e.g. all column drops in one migration PR).
2. Confirm plan.
3. Start with the rename (Workstream A + B) — biggest blast radius. Run go test + npm test after.
4. Column + dead-surface PRs are independent — can be parallel PRs on separate branches.
5. Full acceptance test run (wiki/tests/system-acceptance-test.md) after all PRs merged.
6. Dispatch wiki-curator (will need to update many docs with new file paths).
7. Update roadmap.md.

/simplify rules:
- Rename is mechanical. Do NOT redesign module structure during rename.
- v1/v2 URL sweep: sed-like find-replace across frontend + spec + handler files. No architecture changes.
- If OriginProtection wiring is complex (requires CSRF token flow), just delete the dead config field.
- Do NOT add new features while in delete/rename mode.
- taxonomy T-015 migration: confirm no application code does WHERE code = $X across all tenants
  before dropping the cross-tenant PK. If uncertain, add a comment and defer.

Push back if I try to:
- Redesign the module structure while renaming.
- Keep any /api/v2/* path after this plan (the anchor decision is v1 canonical — no exceptions).
- Bundle Plan 12 screen work into Plan 10.
```

---

## Plan 12 · Screen finalization × 7

```
Mode: implementation. Plan 12 of the MetalDocs refactor roadmap.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 12. Confirm Plans 7, 8, 9, 11 done.
2. wiki/README.md
3. CLAUDE.md — metaldocs-frontend + metaldocs-screen-implementation skills mandatory.
4. wiki/concepts/design-workflow-audit.md — MANDATORY pre-implementation audit for each screen.
5. wiki/architecture/frontend-structure.md — canonical layout rules.
6. wiki/architecture/api-contract.md — codegen types available after Plan 8.
7. Per-screen backlog files (read before each screen's implementation session):
   - wiki/backlog/library-screen.md
   - wiki/backlog/novo-documento.md
   - wiki/backlog/templates.md
   - wiki/backlog/caixa-aprovacao.md
   - wiki/backlog/documento-publicado.md
   - wiki/backlog/template-editor.md
   - wiki/backlog/novo-template-wizard.md

Prerequisite: Plans 7 + 8 done (stable envelope + codegen). Plan 9 done (idempotency on POST
creates used by wizards). Plan 11 done (editor chrome stable for template-editor + documento-publicado).

Goal: every design-source/<slug>/ mockup either implemented (passing design-workflow-audit) or
rejected with rationale in NOTES.md. Close ~44 deferred backlog items across 7 screens.

Screen order (recommended — do one at a time, each = one PR):
1. templates — smallest deferred set (5 items), warms up the workflow.
2. caixa-aprovacao — 7 items, approval API now stable after Plans 7+8.
3. novo-documento — 6 items, wizardReducer + visibility controls.
4. novo-template-wizard — 9 items (largest wizard), Steps 3–5 deferred items.
5. template-editor — 7 items, requires Plan 11 editor-chrome stable.
6. documento-publicado — 9 items, requires registry + documents codegen from Plan 8.
7. library — 7 items, ActivityPanel + audit wiring (audit gated GET from Plan 6).

Per-screen workflow (mandatory):
1. Run metaldocs-screen-implementation skill (6-phase: Audit → Map → Pre-flight → Page assembly
   → Verify → Document).
2. Phase 1 Audit: compare every UI widget in design-source/<slug>/ against real document states,
   RBAC roles, and personas per wiki/concepts/design-workflow-audit.md. Record Keep/Cut/Defer in
   screen NOTES.md BEFORE touching any code.
3. Do NOT implement widgets flagged as Cut/Defer. Do NOT add mocked stat cards, fake data,
   or UI implying unsupported backend behavior.
4. After implementing: start dev server, exercise the screen, screenshot for PR description.
5. Run frontend-screen-reviewer agent to confirm visual + architectural parity before merging.

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-12-screens.md.
   List all 7 screens, estimated deferred-item count, implementation order rationale.
2. Confirm plan.
3. One PR per screen. Do not bundle two screens into one PR.
4. Dispatch wiki-curator after each screen PR to update the backlog file + module doc.
5. On completion: update roadmap.md Plan 12 → done. All 7 backlog files should have 0 open items.

/simplify rules:
- If a deferred item requires a backend endpoint that does not exist: mark it Defer in NOTES.md,
  do NOT mock or fake the data. Open a specific ticket/note for the backend gap.
- Do NOT add new design language or invent UI patterns not in design-source/.
- Do NOT refactor unrelated frontend code while implementing a screen.
- Do NOT start a screen without completing the design-workflow-audit Phase 1 first.

Push back if I try to:
- Implement two screens in one session/PR.
- Skip the design-workflow-audit step.
- Add mocked/hardcoded data to pass visual review.
- Start this plan before Plans 7 + 8 are confirmed done.
```

---

## Plan 13 · Doc-comment + ADR sweep

```
Mode: implementation. Plan 13 of the MetalDocs refactor roadmap. Final plan.

Execution model:
- Codex (codex:rescue skill): implements all Workstream code changes.
- Sonnet/Haiku: writes + runs tests, reviews diffs, makes commits.
- Opus: one grouped final review after ALL workstream commits land (nexus:requesting-code-review). NOT per-PR.

Rules:
- NO FALLBACK. If information needed to implement a fix is missing or unclear, STOP and report. Never guess, never use a default value, never assume a safe fallback. A wrong fallback is worse than no implementation.
- Evidence-based artifact reading: when a tech-debt row's Evidence field cites `_artifacts/XX.md §section`, read that exact file+section before implementing that fix. Do not read artifacts that no debt row in this plan cites.

Read first (in order):
1. wiki/backlog/roadmap.md — locate Plan 13. Confirm Plans 3–12 all done.
2. wiki/README.md
3. CLAUDE.md
4. wiki/decisions/ — all existing ADRs (0001, 0002, 0003, 0007, 0008, 0011, 0012). Note numbering.
5. Per-module tech-debt files for their missing-ADR + doc-comment rows:
   - iam T-011..T-012 (missing-ADR rows), T-013 (doc comments via R-013)
   - auth T-010, T-011 (R-011)
   - approval T-010 (R-010)
   - audit T-011, T-012 (R-011, R-012)
   - templates_v2 T-014 (R-014)
   - registry T-012 (R-012)
   - taxonomy T-014, T-016 (R-014, R-016)
   - editor-ui-eigenpal T-007, T-008 (R-007, R-008)
   - editor-chrome T-008 (R-008)

Goal: Go doc comments on every exported symbol. Every undocumented architectural decision
becomes a standalone ADR or explicitly accepted-as-docs-only with a // ADR-ACCEPTED comment.

Workstream A — ADRs to author (next available number after existing max):
- Tenant-resolution rule (locked in Plan 3): tenant from authn context, not header.
- IAM-table tier-2 + tripwire coverage (Plan 5): why IAM admin tables now have 3-layer defense.
- Canonical audit sink (Plan 6): metaldocs.audit_events is the only sink; governance_events + templates_v2_audit_log are deprecated.
- RFC 9457 rollout policy (Plan 7): per-module rollout sequence, frontend compat shim.
- templates_v2 → templates rename + v2→v1 URL migration (Plan 10).
- templatePlugin mode-gating rule (editor-ui-eigenpal T-007): why template-draft gates the plugin.
- ACL wrapper-only consumption rule (editor-ui-eigenpal T-008): all @eigenpal imports via @metaldocs/editor-ui only.
- EditorChrome slot-API shape (editor-chrome T-008): why slot over compound-components.
- document_families global-vs-tenant decision (taxonomy T-002 resolution from Plan 5): document what was decided.
- Area-hierarchy cycle-prevention rule (taxonomy T-016): application-layer check vs DB trigger.

Workstream B — Go doc comments:
- Add leading // SymbolName ... doc comment to every exported symbol missing one.
- Batch by module: one PR per module (auth, approval, audit, templates_v2, registry, taxonomy,
  editor-ui-eigenpal, editor-chrome, iam, documents).
- Rule: one short line max per symbol. Do NOT write multi-paragraph docstrings.
  The comment says WHY or WHAT CONTRACT, not what the code does line-by-line.

Workstream C — Close remaining minor open items:
- Audit: T-006 PK collision-prone event ID (replace timestamp-based id with uuid.New()).
  T-009 explicit GRANT SELECT on audit_events (add to a new migration).
- Auth: T-007 bidirectional dep note — add // NOTE: non-circular today, see T-007 comment.
- Approval: T-009 NOT VALID FKs — if not done in Plan 10, do here.
- Taxonomy: T-012 pagination — if cardinality is still small, accept as-is with a // TODO:pagination
  comment + note in tech-debt that this is intentionally deferred.
- Registry: T-011 v1/v2 partial path — if not fixed in Plan 10, fix here.
- editor-chrome: T-003 version-pin assertion — add a CI snapshot or comment guard.

Process:
1. nexus:writing-plans → docs/superpowers/specs/2026-05-NN-plan-13-docs-adrs.md.
   List all ADRs with their number assignments. List all modules needing doc-comment PRs.
2. Confirm plan.
3. ADR authoring: one PR for all ADRs (doc-only, no code). Review for accuracy against
   what was actually implemented in Plans 3–12 — do not ratify decisions that were changed mid-plan.
4. Doc-comment PRs: go doc output for each package should show non-empty for every export.
   Run golint or revive exported-comment rule if available.
5. Minor closes: one PR.
6. Dispatch wiki-curator (many wiki docs will need Last verified bumps + new ADR links added).
7. Final: update roadmap.md Plan 13 → done. Flip ALL remaining open plans to done (or note residuals).
   This is the last plan — the roadmap is complete.

/simplify rules:
- ADRs: one ADR per decision, ~20–40 lines each. No 300-line architecture essays.
- Doc comments: one line per symbol. No prose blocks.
- Do NOT reopen closed debt rows to "improve" the implementation. This plan documents, not refactors.
- Taxonomy T-012 pagination: if still small cardinality, accept and note. Do not implement just to close a debt row.

Push back if I try to:
- Use Plan 13 as an opportunity to refactor code that was already shipped in Plans 3–12.
- Author ADRs that contradict what was actually implemented.
- Leave the roadmap doc in a partially-done state — every plan must be marked done or residual before closing this session.
```
