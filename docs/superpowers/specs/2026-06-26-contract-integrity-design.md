# Design Spec — API Contract Integrity (Workstream A)

- **Date:** 2026-06-26 (amended after adversarial review)
- **Mission:** [Document Contract & Storage Integrity](2026-06-26-mission-document-integrity.md)
- **Workstream:** A of 2 (sequenced first; unblocks the template editor and E2E)
- **Sibling spec:** [Storage Integrity](2026-06-26-storage-integrity-design.md)
- **Status:** Approved design, amended — pending implementation plan

---

## 1. Problem

The HTTP contract is **asserted, not enforced**. A live defect proves it: the
template editor crashes with `Cannot read properties of undefined (reading 'version')`
because `getTemplateSchemas` reads a `{ data: { version } }` envelope from an
endpoint that ADR-0035 already migrated to a **flat** body.

Root facts (all pre-existing on `main`, unrelated to the eigenpal ACL work):

1. **The spec is half-migrated.** ADR-0035 ("flat typed bodies") was applied to
   *some* endpoints only. Within one module, same resource family:
   - `GET /templates/{id}` → **enveloped** (`resp.Data.Template`), `routes_query.go:118`
   - `GET /templates/{id}/versions/{n}` → **flat** (`writeJSON(dto)`), `routes_query.go:158`
2. **Two coexisting frontend client styles.** A generated typed client
   (`createClient<paths>`, `client.ts:130`, backed by `openapi-fetch ^0.14.1`,
   already used in ~10 files) coexists with legacy hand-rolled
   `apiFetch<InlineLiteral>` calls. The crash lives in the legacy style —
   `getTemplateSchemas` hard-codes the wrong shape (`templates.ts:344-356`).
3. **The client blind-casts.** `apiFetch` does `return (await res.json()) as T`
   (`client.ts:107`) — zero runtime decode, so drift is invisible until a deref.
4. **Generated types present but bypassed.** `templates.ts:3` imports the generated
   `paths`/`components`, then ignores them via inline generics; it even re-declares
   ~13 `VersionDTO` fields as a hand-override (`templates.ts:26-55`) that *lies*
   about the contract (see §3, A-finding 4).

The same file mixes conventions throughout — `getTemplate` enveloped, `getVersion`
flat, `submit`/`review`/`approve`/`getDocxURL` enveloped, `listTemplates`
defensively probing `data?.templates ?? data?.items`, plus legacy shims
(`saveDraft`, `presignDocxUpload`, `presignSchemaUpload`).

`getTemplateSchemas` is **legacy redundancy** — it fetches the same endpoint
`getVersion` covers, extracting `placeholder_schema` (already a `VersionDTO` field)
plus `lock_version` (also already in-contract — confirmed `openapi.yaml:5425`,
`index.d.ts:2903`). It survived only because it was written against the old
enveloped shape and nothing forced it to match the contract when `getVersion` went
flat.

## 2. Goal

One response-shape law for **2xx resource bodies**, enforced at compile time on the
frontend AND by a true shape-conformance gate on the backend, with no hand-rolled
shape assertions anywhere. A wrong shape must fail a gate, not crash at runtime.

## 3. Decisions (with rationale; amended)

| # | Decision | Rationale |
|---|----------|-----------|
| A1 | **Envelope law (scoped to 2xx resource bodies):** single-resource → flat body; collection/paginated → `{ data:[...], meta }`; **resource + navigation/side-data → enveloped** (it is a one-plus-meta case, e.g. `approve` = version + `next_draft`). Error responses are always RFC 9457 `problem+json`; 204/201/presign bodies keep their existing flat shapes. These non-2xx-resource classes are **carved out** of the law. | Removes the ambiguity the reviewer found for `approve`/`submit`/`review` (which legitimately carry side-data or use `TemplateVersionEnvelope`). Errors/presign are already consistent. |
| A2 | **Enforcement = compile-time typed client + envelope-pattern lint.** Route every call through the generated `paths` types. Lint must ban the **`{ data: ... }` envelope-read pattern** (inline object/`data`-shaped type literals on the transport), NOT all generics — `apiFetch<VersionDTO>` is legitimate and must stay legal. | Converts "shouldn't" into "can't compile." The lint must target the *class* (envelope reads), not just the one inline-generic instance, or `getTemplate`/`getDocxURL`/`submitForReview` survive (reviewer finding). |
| A3 | **No frontend runtime decode (no zod). Instead add a backend shape-conformance gate.** A per-endpoint contract test (or response-vs-OpenAPI-schema validator) asserts handlers *emit* the declared shape — not just key presence. | **Correction:** the earlier claim "backend already gates conformance" was false. `cilint noresponsemap` only bans `map` literals; `typed_response_test` checks key-set only; oapi-codegen drift only proves `api.gen.go` matches the spec, not that a handler emits it faithfully (omitempty / nil-pointer / enum drift slip through). Frontend zod is still rejected (YAGNI, bundle cost); the honest fix is a real backend gate, owned here. |
| A4 | **Delete legacy duplicates/shims via refactor, not blind delete.** `getTemplateSchemas` is replaced by a pure domain-mapper `getVersionSchemas(v: VersionDTO) → { schemas, lockVersion }` (reusing `placeholderFromWire`), fed by the single typed `getVersion`; `useTemplateSchemas` and its 6 mock sites are rewritten accordingly. | **Correction:** `getTemplateSchemas` returns a *transformed* domain shape with a live consumer hook + mocks; it is not interchangeable with raw `getVersion`. The fix is a mapper + consumer rewrite, not a one-line delete. |

## 4. Scope

### 4.1 Contract policy (codify the law)
- Extend ADR-0035 to state the A1 law explicitly, including the side-data and
  carve-out rules. Update `wiki/architecture/api-contract.md`.

### 4.2 Backend (multi-module migration — hard-scoped)
- **Inventory deliverable:** enumerate **every** enveloped 2xx resource writer.
  Known surface (reviewer-counted ~15 `.Data` writers / 9 `data`-wrapper schemas):
  `templates/{routes_query,routes_lifecycle,routes_schema,routes_generated}.go`,
  `documents/{handler,fillin_handler}.go`, `controlleddocuments/routes.go`, plus an
  audit of taxonomy/approval/etc. This inventory is the load-bearing task; ship it
  as an explicit checklist.
- **Migrate single-resource stragglers to flat:** `GET /templates/{id}`; the
  docx-url endpoint (`getDocxURL`, migrated here per the mission coupling); and any
  other single-resource reads the inventory surfaces. Keep `approve`/`submit`/
  `review` enveloped (A1 side-data rule) — adjust only if the inventory shows a
  pure single-resource return.
- Per endpoint: edit `openapi.yaml` → `go generate ./...` → fix handler → update
  contract test. Each endpoint's change is atomic (BE shape + regenerated FE types
  in one release — see mission §3 deploy rules).
- **Pre-declare B's docx-url not-ready/not-found typed response** in the OpenAPI
  contract now, so Workstream B implements its handler against a frozen shape.
- Add the A3 backend shape-conformance gate.

### 4.3 Frontend
- Route **all** calls through the generated typed client; replace every
  `apiFetch<InlineLiteral>` and every `body.data` envelope read with a typed
  operation sourced from `paths`.
- **Remove the `VersionDTO` hand-override block** (`templates.ts:26-55`) and consume
  the generated type. Fix the live type-drift: generated `placeholder_schema` is an
  **array** (`{[k]:unknown}[] | null`, `openapi.yaml:5407`), not the `Record<...>`
  the override declares (reviewer finding — a Disease-A instance inside the audited
  file).
- **Replace** `getTemplateSchemas` with the `getVersionSchemas` mapper (A4); rewrite
  `useTemplateSchemas` + mocks. **Delete** `presignDocxUpload`, `presignSchemaUpload`,
  `saveDraft`, and `listTemplates` defensive probing.
- **Keep** the placeholder wire↔domain mapper (`placeholderFromWire`/`ToWire`) — a
  legitimate ACL, not legacy.
- **Lint rule** per A2 (envelope-pattern, not all-generics).

## 5. Non-goals
- Frontend runtime validation / zod (A3).
- Rewriting the snake_case↔camelCase mapping layer.
- Any storage-key write/read work (Workstream B) — except *declaring* B's docx-url
  response shape in the contract (mission coupling).
- Touching the eigenpal ACL walls (ADR 0046).

## 6. Acceptance criteria
1. Template editor mounts end-to-end (the `getTemplateSchemas` crash class is gone).
2. No `apiFetch<InlineLiteral>` shape assertions and no inline `{data:...}` envelope
   reads remain; the lint rule enforces it and is proven to flag a reintroduced one.
3. Every single-resource 2xx endpoint returns flat; every collection returns
   enveloped; side-data endpoints follow the A1 rule; spec, generated types, and
   handlers agree (CI drift guard green).
4. The A3 backend shape-conformance gate exists and passes for migrated endpoints.
5. `VersionDTO` hand-override removed; `placeholder_schema` typed as the wire array.
6. `go build ./...`, `go test ./...`, frontend tests, and typecheck green.
7. Legacy duplicates/shims removed via refactor (consumers + mocks updated).

## 7. Risks
- **Backend blast radius:** multi-module migration (~15 writers), not templates-
  local. Mitigate: inventory checklist, endpoint-by-endpoint atomic changes, CI
  drift guard.
- **Breaking wire change / deploy hazard:** see mission §3 — lockstep BE+FE release
  or tolerant-reader-first; never new BE shape against old FE bundle.
- **Test fallout:** enveloped-shape mocks break. Per CLAUDE.md: delete one-off
  scaffolding, repair only contract/invariant guards; canonical framework for new.

## 8. Test plan
- Backend: per-endpoint shape-conformance test (A3) + the oapi-codegen drift guard.
- Frontend: typed-client smoke per migrated call; lint negative-test (a reintroduced
  envelope read fails lint).
- The editor-load + full pipeline is proven by the **mission terminal E2E gate**
  (mission §4), not here.

## 9. Evidence / references
- ADR 0035 (`wiki/decisions/0035-flat-typed-responses-and-presign-status.md`)
- `routes_query.go:118,158`; `templates.ts` (mixed conventions; override `:26-55`);
  `client.ts:107,130`; `openapi.yaml:5407,5425`; `index.d.ts:2881,2903`
- Gates: `.github/workflows/api-contract.yml`; `tools/cilint/internal/analyzers/noresponsemap.go`
