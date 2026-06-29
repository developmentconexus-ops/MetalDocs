# SP-3 — Token Dictionary Management UI — Design

**Date:** 2026-06-29
**Status:** Approved design (increment SP-3 of the Template Tokens program).
**Program north-star:** `2026-06-27-template-tokens-north-star.md` (§5 SP-3).
**Depends on:** SP-1 (tenant dictionary backend — domain, storage, IAM caps), SP-2 (freeze substitution + tag validation). Both landed.
**Governing ADRs:** 0048 (tenant token dictionary), 0049 (creation-time dictionary substitution).

---

## 1. Problem & intent

SP-1 delivered the tenant token dictionary backend: domain + tenant-scoped storage +
capability-gated CRUD routes (`GET/POST /tokens`, `GET/PUT/DELETE /tokens/{id}`) under
the new IAM capabilities `token.view` (read) and `token_dictionary.manage` (write). SP-2
wired freeze-time substitution and template tag validation. There is **no authoring
surface** for the dictionary yet — entries can only be created via direct API calls.

SP-3 delivers the **management UI**: a capability-gated CRUD screen where a holder of
`token.view` can list dictionary entries and a holder of `token_dictionary.manage` can
create, edit, and delete them. This is the screen-level slice only — the in-editor token
palette (SP-4), import/export (SP-4), and draft-time preview (SP-5) are out of scope.

## 2. Scope

**In scope**
- A capability-gated CRUD screen for tenant token dictionary entries.
- Full client-side validation of the create/edit form (friendly first line; server
  remains authoritative).
- Route registration, navigation entry point, query/mutation wiring.
- Regenerating the frontend API types to include the token dictionary DTOs.
- Component / hook / route tests using the canonical frontend test framework.

**Out of scope**
- Any backend change (SP-1 CRUD + caps already exist; the OpenAPI spec already defines
  the routes and schemas).
- In-editor token palette / authoring-surface integration (SP-4).
- Bulk import/export (SP-4).
- Draft-time token preview (SP-5).

## 3. Placement & authorization

Tokens are a **document/template platform concept**, not an admin-role artifact. Placing
the screen under `admin/*` would imply role-based access, which violates the non-negotiable
**AuthZ = capabilities, never roles** invariant — any holder of `token_dictionary.manage`
(which may be granted to authors/editors, not only admins) must be able to manage entries.
The screen therefore lives in the **templates** workspace.

- **Route:** `templates/tokens`
  `handle: { workspaceView: "templates", requiresCapability: "token.view" }`, lazy page.
- **Entry point:** a "Token Dictionary" link/button on `TemplatesListRoutePage`, rendered
  only for holders of `token.view`.
- **Write gating:** Create / Edit / Delete controls are gated on `token_dictionary.manage`
  via the existing capability-check mechanism (the same one used by neighbor admin screens),
  never on a role. The backend (tier-1 route→capability + tier-2 in-tx) remains the
  authoritative gate; the UI gate is the friendly first line.

## 4. Components & feature tree

New feature module under `frontend/apps/web/src/features/tokens/`, mirroring the
`features/taxonomy/` pattern (the closest existing capability-gated tenant-CRUD admin
screen):

```
features/tokens/
  api/
    tokens.ts          # BASE /api/v1/tokens — list/get/create/update/delete over lib/api/client.request<T>
    catalog.ts         # computed placeholder catalog for the D4 collision check
                       #   (reuse GET /templates/placeholder-catalog; import the templates client if it exposes one)
  queries/
    constants.ts       # query keys
    useTokensQuery.ts  # list query (no pagination — bounded per-tenant)
    useTokenMutations.ts # create/update/delete; invalidate list key on success
  types.ts             # re-export generated TokenDictionaryEntry + request DTOs (see §7)
  TokenList.tsx        # table: name, label, value (truncated), description, updated_at, row actions
  TokenEditDialog.tsx  # create + edit form with full client validation (§5)
  pages/
    TokensRoutePage.tsx
  routes.tsx           # tokenRoutes (registered in app/AppRouter.tsx)
  index.ts
```

Each unit has one purpose and a defined interface: the `api/` layer speaks HTTP, the
`queries/` layer owns caching/invalidation, `TokenList`/`TokenEditDialog` are presentational
+ form units, the page composes them, `routes.tsx` declares the gated route.

## 5. Form validation (full mirror — friendly first line)

The server is always authoritative; the form mirrors the rules for instant feedback.
Validation uses `@metaldocs/shared-tokens` (`isValidIdent`, `isReservedIdent`, `IDENT_RE`):

| Field | Rules |
|---|---|
| `name` | required; length 1–64; `isValidIdent` (grammar `^[A-Za-z_][A-Za-z0-9_]*$`); `!isReservedIdent` (JS/DB reserved set); **not present in the computed placeholder catalog** (D4 collision guard — block names equal to any of the 8 computed/native keys) |
| `value` | required; length 1–4096 |
| `label` | required; length 1–256 |
| `description` | optional; length ≤1024 |

The computed-catalog collision set is fetched once (§4 `catalog.ts`) and checked
client-side. Server-side this is enforced by the SP-1/SP-2 reserved-name guard (ADR 0049
D4/D5); the UI check only spares the user a round-trip. Server `409`/`422` (RFC 9457
`problem+json`) are mapped back to field errors / toast as the fallback.

## 6. Data flow & error handling

- **List:** `useTokensQuery` reads `GET /tokens` → `ListTokenDictionaryEntriesResponse.items`.
  No pagination (the endpoint is `x-pagination-exempt`: bounded per-tenant dictionary).
- **Mutations:** create (`POST /tokens`), update (`PUT /tokens/{id}`), delete
  (`DELETE /tokens/{id}`) invalidate the list query key on success.
- **Collision set:** computed catalog fetched once and cached for the form's collision check.
- **Errors:** RFC 9457 `problem+json` parsed by the existing error mapper; surfaced as
  inline field errors (validation problems) or a toast (unexpected failures). Delete is
  confirmed before firing (mirror taxonomy's destructive-action confirm).

## 7. Types

Regenerate the frontend API types (`frontend/apps/web/src/lib/api-types`) from the
OpenAPI spec so `TokenDictionaryEntry`, `CreateTokenDictionaryEntryRequest`, and
`UpdateTokenDictionaryEntryRequest` are present (they are defined in
`api/openapi/v1/openapi.yaml` but missing from the currently-generated types). The
feature's `types.ts` re-exports these generated DTOs. This follows the generated-DTO
preference (ADR 0035 / flat-envelope guidance) over hand-written body shapes.

Fallback (only if a full regen is out of band for this increment): hand-write `types.ts`
mirroring the OpenAPI schema exactly, and file a follow-up to regenerate.

## 8. Testing

Using the canonical frontend test framework (no bespoke harness), mirroring taxonomy's
test style:

- `TokenList` render test (rows, truncation, action visibility under each capability).
- `TokenEditDialog` validation tests: valid submit, grammar failure, reserved-ident
  failure, computed-catalog collision failure, length bounds, server-error mapping.
- Query/mutation hook tests: list fetch, create/update/delete invalidation.
- Route capability-gating test: `token.view` required to reach the route;
  write controls hidden without `token_dictionary.manage`.

## 9. ADR 0049 forensic-reconstruction gate

ADR 0049 §129–140 flags that `source='dictionary'` pinned rows store the value but not
the originating dictionary entry name, and states a "post-SP-2 owner must be named before
shipping SP-3." That gate concerns **forensic-audit** features (provenance reconstruction),
which SP-3 (management CRUD UI) does not touch and does not depend on. Resolution: update
ADR 0049 §140 to record that forensic reconstruction is owned by a future forensic-audit
epic, explicitly decoupled from SP-3. No SP-3 work item depends on it.

## 10. Verification

- `npm run typecheck` / frontend build clean (regenerated types compile).
- `make test` (frontend) green, including the new SP-3 tests.
- Route reachable only with `token.view`; write controls present only with
  `token_dictionary.manage`.
- Create → list reflects entry; edit → row updates; delete → row removed; invalid name
  (grammar/reserved/collision) blocked client-side with a clear message; server error
  surfaced on forced failure.

## 11. Relationship to the program

SP-3 consumes the SP-1 contract unchanged and adds no new capability, route, or schema.
It unblocks SP-4 (the authoring palette can later add the dictionary as a second data
source) without coupling to it.
