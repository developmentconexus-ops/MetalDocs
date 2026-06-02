# Route Admin — Design Notes

**Slug:** `route-admin`
**Design file:** `index.html`
**Route:** `/workspace/approval-routes` (gated by `requiresAdmin: true`)
**Status:** Implemented — PR-4 (FE rewrite)

## Purpose
Tenant administrators configure approval routes per document profile:
- Create / edit named routes with N stages.
- Each stage: required role, area, quorum kind (any_1_of, all_of, m_of_n), drift policy.
- Deactivate route — registers a governance event with a required reason (≥4 chars).

## Design decisions
- **Container/presentational split** — `RouteAdminPage` owns data + mutations; `RouteListTable`, `RouteEditorDialog`, `DeactivateRouteDialog` are pure children.
- **Native `<dialog>`** — focus trap + Esc handled by the browser; `Dialog` primitive in `components/ui/` wraps it. No `@radix-ui` dependency added.
- **Status badge** — visually distinct: green pill (Ativa) vs slate pill (Inativa) with leading dot.
- **Edit disabled tooltip** — cause-based copy (`Rota inativa — desativada e somente leitura.`), not the legacy "referenced by active instance" misnomer (active routes are always editable).
- **Role select** — labels come from `useIamRolesQuery` (ADR 0018 Option B fallback to `lib/iam/roles.ts`); never a frozen import.
- **Optimistic mutations** — create inserts a placeholder row with a `optimistic-*` id; update rewrites in place; deactivate flips `active=false`. All three roll back on error and `invalidateQueries` on settle.
- **Error mapping** — 412 / 409 / 422 / 403 each map to a distinct PT-BR message via `messageForRouteError(problem.code, status)`.
- **Stable stage keys** — every draft stage carries a `uid` (uuidv4) so React keys survive reorder / insert / delete.

## Keep / Cut / Defer
- **Keep:** ordering by `updated_at` desc in the list (server-side default).
- **Cut:** legacy `Members` field on stage payload (removed in PR-1 / PR-2).
- **Defer:** route detail page with version history; bulk import; per-stage approver overrides.

## Test coverage
- `RouteAdminPage.test.tsx` — happy path, validation cascade, m_of_n bound check, optimistic 409 rollback, deactivate reason gating, Esc to close, IAM role label source.
- Manual preview drive script in PR body.

## Out of scope
- Backend handler changes (PR-2 owns).
- OpenAPI shape changes (PR-1 owns).
- Tier-1 cap split for `ControlledDocumentDetailPanel.submit` (F-001).
