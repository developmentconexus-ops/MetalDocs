# F2 — Submit contract: optionalize governance fields + add revision_title

> **Milestone:** M1 canonical-submit-backend · **Findings:** 2 · **Status:** spec approved (ADR 0073-ratified)
> **Approved:** 2026-07-06 — contract distilled from ADR 0073 §2 + governing spec §1.3; no operator interview needed (contract fully specified by ratified ADR).

## Consumer contract (who consumes, required shape)

**Consumer:** the FE editor (`documents.ts`, dirty-tree M2) POSTs `/documents/{id}/submit`
with `If-Match: "v<N>"` (N≥0), `Idempotency-Key`, and body `{ revision_title?, reason_for_change?, reason_category? }` — **no `route_id`, no `content_hash`**. Integrations MAY still send `route_id`/`content_hash`.

Handwritten `contracts.SubmitRequest.Validate()` must therefore:
- Treat `route_id` **optional**: empty → OK (server resolves in-tx, F1); present → must be valid UUID.
- Treat `content_hash` **optional**: empty → OK (server binds head hash, F1); present → 64 hex.
- Accept `revision_title` (string, optional) and thread it to `application.SubmitRequest.RevisionTitle`.
- REV≥1 requiredness of title/reason stays enforced **downstream** in the service (governed rev number).

The OpenAPI spec + generated code are **already on disk** (dirty tree) — align the handwritten contract/handler to them; do not re-edit the spec.

## Non-goals
- No change to reason_for_change/reason_category handling (already present).
- No ETag/response-shape change (YAGNI §4).
- No service-layer logic here (that is F1).

## Validation Gate
- `contracts.SubmitRequest{}` (all empty) `.Validate()` → nil.
- `{route_id:"not-a-uuid"}` → validation error (400 upstream).
- `{content_hash:"xyz"}` → validation error (400 upstream).
- `revision_title` decoded from body and passed into `application.SubmitRequest`.
- Tests: `contracts` unit test for optional/format matrix; `go build ./...` green.
