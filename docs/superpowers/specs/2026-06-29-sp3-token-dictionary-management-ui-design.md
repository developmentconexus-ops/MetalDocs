# SP-3 — Token Dictionary Management UI — Design

**Date:** 2026-06-29
**Status:** Approved design (increment SP-3 of the Template Tokens program).
**Program north-star:** `2026-06-27-template-tokens-north-star.md` (§5 SP-3).
**Depends on:** SP-1 (tenant dictionary backend — domain, storage, IAM caps), SP-2 (freeze substitution + tag validation). Both landed.
**Governing ADRs:** 0048 (tenant token dictionary), 0049 (creation-time dictionary substitution).
**Canonical reference:** `wiki/architecture/frontend-structure.md` (folder layout, API, query, design-system rules). Grade-A exemplars to follow: `features/approval`, `features/templates` (June 2026). **`features/taxonomy` is legacy (2026-06-09) and is explicitly NOT a structural model.**

---

## 1. Problem & intent

SP-1 delivered the tenant token dictionary backend: domain + tenant-scoped storage +
capability-gated CRUD routes (`GET/POST /tokens`, `GET/PUT/DELETE /tokens/{id}`) under
the new IAM capabilities `token.view` (read) and `token_dictionary.manage` (write). SP-2
wired freeze-time substitution and template tag validation. There is **no authoring
surface** for the dictionary yet — entries can only be created via direct API calls.

SP-3 delivers the **management UI**: a capability-gated CRUD screen where a holder of
`token.view` can list dictionary entries and a holder of `token_dictionary.manage` can
create, edit, and delete them. Screen-level slice only — the in-editor token palette
(SP-4), import/export (SP-4), and draft-time preview (SP-5) are out of scope.

## 2. Scope

**In scope**
- A capability-gated CRUD screen for tenant token dictionary entries.
- Full client-side validation of the create/edit form (friendly first line; server
  remains authoritative).
- Route registration, navigation entry point, query/mutation wiring.
- Regenerating the frontend API types so the token dictionary DTOs are available.
- Component / hook / route tests using the canonical frontend test framework.

**Out of scope**
- Any backend change (SP-1 CRUD + caps already exist; the OpenAPI spec already defines
  the routes and schemas).
- In-editor token palette / authoring-surface integration (SP-4).
- Bulk import/export (SP-4).
- Draft-time token preview (SP-5).
- A `design-source/<slug>/` claude.design HTML mock + `metaldocs-screen-implementation`
  6-phase workflow. SP-3 is built **directly to the canonical structure** (operator
  decision 2026-06-29), consistent with how SP-0/1/2 built the tokens UI. Visual/code
  review is by `frontend-code-reviewer`.

## 3. Placement & authorization

Tokens are a **document/template platform concept**, not an admin-role artifact. Placing
the screen under `admin/*` would imply role-based access, which violates the non-negotiable
**AuthZ = capabilities, never roles** invariant — any holder of `token_dictionary.manage`
(which may be granted to authors/editors, not only admins) must be able to manage entries.
The screen therefore lives in the **templates** workspace.

- **Route:** `templates/tokens` — a flat `RouteObject` (the `templates/*` routes are flat
  siblings under `AppShell`, not nested children), added via a `tokenRoutes` array spread
  into `app/AppRouter.tsx`.
  `handle: { workspaceView: "templates", requiresCapability: "token.view" }`, lazy page.
  Route-level gating is enforced by the existing `requiresCapability` handle mechanism
  (`AppShell` route guard), exactly as `approval`/`iam` routes do.
- **Entry point:** a "Token Dictionary" link/button on `TemplatesListRoutePage`, rendered
  only for holders of `token.view`.
- **Write gating:** Create / Edit / Delete controls are gated with the canonical
  `useHasCapability('token_dictionary.manage')` hook (`features/iam/hooks/useHasCapability`,
  the pattern used by `iam` and `documents` pages) — never a role check. The backend
  (tier-1 route→capability + tier-2 in-tx) remains the authoritative gate; the UI gate is
  the friendly first line.

## 4. Components & feature tree

New feature module under `frontend/apps/web/src/features/tokens/`, following the canonical
layout in `frontend-structure.md` §2 and the `features/approval` / `features/templates`
exemplars (generated types, `lib/api` transport, central query keys, CSS-module styling):

```
features/tokens/
  api/
    tokens.ts            # apiFetch over lib/api; paths/components from lib/api-types
    tokensTypes.ts       # aliases of generated DTOs (see §7), e.g.
                         #   export type TokenDictionaryEntry = components['schemas']['TokenDictionaryEntry']
  queries/
    useTokensQuery.ts    # list query; QK.tokens.list()
    useTokenMutations.ts # create/update/delete; invalidate QK.tokens.list() on success
  components/
    TokenList.tsx + .module.css + .test.tsx        # table: name, label, value(trunc), description, updated_at, row actions
    TokenEditDialog.tsx + .module.css + .test.tsx  # create+edit form, full client validation (§5)
  pages/
    TokensRoutePage.tsx  # route entry; composes list + dialog; capability-gated controls
  routes.tsx             # tokenRoutes: RouteObject[]
  index.ts               # barrel: only the feature's public API
```

Cross-feature reuse: the computed placeholder catalog (for the §5 D4 collision check) is
**owned by `templates`**. Reuse `fetchPlaceholderCatalog` / `usePlaceholderCatalogQuery`
by **publishing them through the `features/templates` barrel** (`index.ts`) and importing
from `features/templates` — do **not** duplicate the catalog fetch/query in `tokens`. (If a
future increment needs the catalog in a third domain, lift it to a shared module then.)

Each unit has one purpose and a defined interface: `api/` speaks HTTP over the shared
transport, `queries/` owns caching/invalidation, the `components/` are presentational +
form units, the page composes them, `routes.tsx` declares the gated route, `index.ts`
exposes only the public surface.

## 5. Form validation (full mirror — friendly first line)

The server is always authoritative; the form mirrors the rules for instant feedback.
Validation uses `@metaldocs/shared-tokens` (already an `apps/web` workspace dependency —
`isValidIdent`, `isReservedIdent`, `IDENT_RE`):

| Field | Rules |
|---|---|
| `name` | required; length 1–64; `isValidIdent` (grammar `^[A-Za-z_][A-Za-z0-9_]*$`); `!isReservedIdent` (JS/DB reserved set); **not present in the computed placeholder catalog** (D4 collision guard — block names equal to any computed/native key) |
| `value` | required; length 1–4096 |
| `label` | required; length 1–256 |
| `description` | optional; length ≤1024 |

The computed-catalog collision set comes from `usePlaceholderCatalogQuery` (reused from
`templates`, §4) and is checked client-side. Server-side this is enforced by the SP-1/SP-2
reserved-name guard (ADR 0049 D4/D5); the UI check only spares the user a round-trip.
Server `409`/`422` (RFC 9457 `problem+json`) are mapped back to field errors / toast as the
fallback.

## 6. Data flow & error handling

- **Transport:** `apiFetch` from `lib/api` (credentials, RFC 9457 Problem decoding,
  `authn.expired` dispatch). `BASE = '/api/v1'`; calls use `'/api/v1/tokens'` etc. No
  custom per-feature `request<T>` wrapper.
- **List:** `useTokensQuery` reads `GET /tokens` → `ListTokenDictionaryEntriesResponse.items`.
  No pagination (the endpoint is `x-pagination-exempt`: bounded per-tenant dictionary).
  Query key `QK.tokens.list()`.
- **Mutations:** create (`POST /tokens`), update (`PUT /tokens/{id}`), delete
  (`DELETE /tokens/{id}`) invalidate `QK.tokens.list()` on success.
- **Query keys:** add a `tokens` namespace to the central `lib/queryKeys.ts` `QK`
  (`QK.tokens.list()`); never inline key arrays.
- **Errors:** RFC 9457 `problem+json` parsed by the existing `ApiError`/`parseProblem`
  path; surfaced as inline field errors (validation problems) or a toast (unexpected
  failures). Delete is confirmed before firing.

## 7. Types (generated, not hand-written)

The canonical API type source is the generated `frontend/apps/web/src/lib/api-types`
(`openapi-typescript` from `api/openapi/v1/openapi.yaml`). The token dictionary schemas
(`TokenDictionaryEntry`, `CreateTokenDictionaryEntryRequest`,
`UpdateTokenDictionaryEntryRequest`, `ListTokenDictionaryEntriesResponse`) are defined in
the OpenAPI spec but **absent from the currently-generated `index.d.ts`** — so SP-3 must
**regenerate** it:

```
cd frontend/apps/web && npm run gen:api    # openapi-typescript -> src/lib/api-types/index.d.ts
```

Note `gen:api` **rewrites the whole** `index.d.ts` from the current spec (global regen, not
additive); the diff is expected to add only the token schemas if the spec is otherwise in
sync — review the diff to confirm no unrelated drift. The feature aliases the generated
types in `api/tokensTypes.ts` (e.g. `components['schemas']['TokenDictionaryEntry']`,
`paths['/tokens']['post']['requestBody']…`), mirroring `templates/api/catalog.ts` and
`approval/api/approvalTypes.ts`. No hand-written DTO shapes.

## 8. Styling

CSS modules per component (`Component.module.css`), using the design tokens in
`styles/tokens.css` (wine palette `--brand-*`, `--bg`, `--accent`, …). No inline styles, no
duplicate catalog styling. Match the visual language of the `templates` / `approval`
screens.

## 9. Testing

Canonical frontend test framework (no bespoke harness), mirroring the `approval`/`templates`
test style:

- `TokenList` render test (rows, value truncation, action visibility per capability).
- `TokenEditDialog` validation tests: valid submit, grammar failure, reserved-ident
  failure, computed-catalog collision failure, length bounds, server-error mapping.
- Query/mutation hook tests: list fetch, create/update/delete invalidation of
  `QK.tokens.list()`.
- Route capability-gating test: `token.view` required to reach the route; write controls
  hidden without `token_dictionary.manage`.

## 10. ADR 0049 forensic-reconstruction gate

ADR 0049 §129–140 flags that `source='dictionary'` pinned rows store the value but not the
originating dictionary entry name, and states a "post-SP-2 owner must be named before
shipping SP-3." That gate concerns **forensic-audit** features (provenance reconstruction),
which SP-3 (management CRUD UI) does not touch and does not depend on. Resolution: update
ADR 0049 §140 to record that forensic reconstruction is owned by a future forensic-audit
epic, explicitly decoupled from SP-3. No SP-3 work item depends on it.

## 11. Out of scope

(See §2.) In-editor token palette (SP-4), import/export (SP-4), draft preview (SP-5),
backend changes, and the design-source screen workflow.

## 12. Verification

- `npm run gen:api` diff reviewed (only token schemas added).
- `npm run typecheck` / frontend build clean (regenerated types compile; aliases resolve).
- `make test` (frontend) green, including the new SP-3 tests.
- Route reachable only with `token.view`; write controls present only with
  `token_dictionary.manage`.
- Create → list reflects entry; edit → row updates; delete → row removed; invalid name
  (grammar/reserved/collision) blocked client-side with a clear message; server error
  surfaced on forced failure.

## 13. Relationship to the program

SP-3 consumes the SP-1 contract unchanged and adds no new capability, route, or schema. It
publishes the existing `templates` placeholder-catalog query through that feature's barrel
(no duplication) and unblocks SP-4 (the authoring palette can later add the dictionary as a
second data source) without coupling to it.
