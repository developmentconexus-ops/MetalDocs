# ADR 0035 — Flat typed responses + presign endpoint status convention

> **Status:** Accepted
> **Date:** 2026-06-15
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Grade-A completion program · Milestone 1 (Contract / API integrity) · Feature F1.2 (status & body conformance)
> **Supersedes:** none
> **Related ADRs:** [0012 — contract-first API](./0012-contract-first-api.md)
> **Related code (Last verified 2026-06-15):**
> - `internal/modules/templates/delivery/http/routes_create.go:36` — `createNextVersion`
> - `internal/modules/templates/delivery/http/routes_autosave.go:42,90` — presign / commit autosave
> - `internal/modules/templates/delivery/http/routes_query.go:160` — `getVersion`
> - `internal/modules/documents/delivery/http/handler.go:881,954` — `listCheckpoints` / `createCheckpoint` (F1.1 flat precedent)
> - `api/openapi/v1/openapi.yaml` — component schemas `TemplateVersion`, `TemplatePresignAutosaveResponse`, `DocumentCheckpoint`

---

## Context

The MetalDocs backend has historically wrapped public-route responses in a `{ "data": { … } }`
envelope and emitted bodies as `map[string]any` literals. Two consequences fall out of that:

1. **OpenAPI declares no schema** for most endpoints (`responses: '200': { description: ok }`), so
   the wire shape is undeclared and the generated FE/BE type clients have nothing to bind to. The
   FE adapters compensate by reading `body.data.…` with hand-typed result structs.
2. **Public-route handlers serialize untyped Go maps.** The H-D class in the Grade-A audit
   (mission §5, report §6) is exactly this: `map[string]any` emits on public routes plus
   ad-hoc envelope wrappers. The fix is structural, not cosmetic — only typed responses survive
   `oapi-codegen` regeneration and only declared schemas can be validated end-to-end.

The Grade-A Milestone 1 program (F1.1 / F1.2 / F1.3 / F1.4) is sunsetting the H-D class
endpoint-by-endpoint. Each feature lands a small cluster of related endpoints to keep the contract
diff coherent and the validator audit trail readable.

Two additional architectural questions surfaced during F1.2 and need a durable answer because they
will shape every future endpoint added to this codebase:

- **(a) Envelope vs flat body:** when typing a previously-untyped endpoint, do we declare the
  legacy `{ "data": { … } }` shape (keep wire-compatibility with old FE adapters), or do we drop
  the envelope and declare a flat top-level schema (matching the F1.1 precedent —
  `DocumentCheckpoint` ships flat)?
- **(b) Status code for presign endpoints:** several presign endpoints currently return `201
  Created` with a body. The mission report flagged the `201` as part of the H-D class. But what
  is the **canonical** status code for a presign endpoint, and what is the rule for a genuine
  resource-create endpoint that legitimately materializes a row?

## Decision

### D1. Modern endpoints return flat typed bodies. The `{ "data": { … } }` envelope is sunset.

Every public-route response touched by a Grade-A milestone (or by any new feature going forward)
**declares a flat OpenAPI schema** and the handler emits the generated typed struct directly. The
`{ "data": { … } }` envelope is **not preserved** at the contract boundary.

The sunset is **incremental, per-endpoint, per-feature** — not a single repo-wide PR. Each feature
removes the envelope from the endpoints in its named cluster and leaves the rest alone. The
legacy `toTemplateResponse` / `toVersionResponse` `map[string]any` mappers are deleted as soon as
their last call site has migrated (F1.3 retires `toVersionResponse`).

Acceptable transitional state: half-typed module. The repo will go through a window where some
endpoints in the same module emit a flat typed body and others still emit `{ "data": { … } }`. The
window is closed by the feature plan (M1 closes the cluster covered by F1.1–F1.4; subsequent
milestones close the rest of the templates module). This is preferable to a single mega-PR that
ships all endpoints at once and entangles validator audit trails across unrelated concerns.

### D2. Presign endpoints return `200 OK + body`. Genuine resource creates return `201 Created + body`.

A **presign** endpoint computes an ephemeral signed token. No server-side resource is materialized
— the row, blob, queue entry, or upload doesn't exist yet; the token is a permission slip the
client uses to materialize it. The industry convention (AWS S3, GCS, Azure Blob, every major SaaS
upload SDK) is `200 OK + body` for presign. We adopt that convention.

A **resource-create** endpoint that actually inserts a row, writes a blob, enqueues a job, or
otherwise materializes state on the server returns `201 Created + body` with the created
resource (or its canonical handle). `POST /templates/{id}/versions` (`createNextVersion`) is a
genuine create — it inserts a new `template_versions` row — and stays `201` even though the
mission report originally framed it as a `201 → 200` fix. The H-D class on that endpoint is the
`map[string]any` body, not the status code; the status code was already canonical.

### D3. Each feature plan must name its consumer contract before any code.

Per the Milestone workflow's per-feature contract gate, the consumer contract (FE adapter, downstream
service, generated client) is read from the consumer **first**, then the producer is built to
match. ADR 0035 does not relax that requirement; it codifies the **defaults** (flat, `200` for
presign, `201` for create) that the contract should adopt when the consumer has no opinion or
where the legacy contract is wrong.

## Consequences

### Positive

- **One typed shape per resource.** Once a resource is typed (e.g. `TemplateVersion`), every
  endpoint that emits it shares the same component schema in `openapi.yaml` and the same
  generated Go struct + TypeScript type. No more parallel hand-typed FE `VersionDTO` definitions
  drifting from the wire.
- **`oapi-codegen` produces useful types.** The FE codegen pipeline finally has a declared shape
  to compile against, which unblocks future contract-test guards in CI (declared shape ⊆ wire
  shape) and lets us delete hand-written adapter types as they migrate.
- **Generated clients match the wire.** Future Grade-A milestones (mobile, partner clients, public
  SDK) can consume the OpenAPI directly without an FE-side translation layer.
- **Status codes match the industry vocabulary.** Presign 200 / create 201 means new engineers
  read the code with the right intuition; the audit grep against `201` no longer false-flags
  legitimate creates.

### Negative

- **Half-typed transitional state.** Until the full templates module migrates (F1.2 → F1.3 →
  future), some endpoints in the module emit flat typed bodies and others emit `{ "data": { … } }`.
  Reviewers must check the OpenAPI declaration per endpoint, not assume modular consistency.
- **FE adapter churn.** Each migrated endpoint forces a same-PR FE adapter edit (drop `body.data.…`
  indirection). Cheap per endpoint, but it's churn distributed across multiple PRs over multiple
  weeks.
- **Mission report deviation.** The original H-D-class framing on `createNextVersion` said
  `201 → 200`. ADR 0035 overrides that — the canonical status is `201`. The deviation is
  recorded in `docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.2-status-and-body-conformance/spec.md`
  (interview Q2) and the milestone validator was notified at PASS time.

### Neutral

- **Domain models unchanged.** ADR 0035 governs wire shape only. `domain.Document`,
  `domain.TemplateVersion`, and the rest of the domain layer keep their existing Go shape; the
  wire mapping happens at the handler boundary (`toAPI*` mappers, mirroring the F1.1
  `toAPICheckpoint` precedent).

## Alternatives considered

- **Keep the `{ "data": { … } }` envelope and declare it in OpenAPI.** Rejected: every endpoint
  would need a one-off wrapper schema (`CreateTemplateVersionResponse: { data: { version:
  TemplateVersion } }`), the codegen pipeline would produce nested generated types nobody wants,
  and the FE adapter layer would survive forever as an indirection that exists for no reason.
  The F1.1 precedent (flat `DocumentCheckpoint`) had already settled the question for one
  resource type; we extend that to the rest.
- **Single repo-wide envelope-sunset PR.** Rejected: would couple unrelated modules
  (templates / documents / taxonomy / audit) into one mega-diff, blow up the per-feature
  validator gate, and force every reviewer to context-switch across the entire H-D class at
  once. The per-feature incremental sunset preserves the milestone workflow's per-feature
  evidence trail.
- **Use `200` for both presign and resource-create.** Rejected: collapses two distinct REST
  semantics (no side effect vs material resource creation) into one status code; loses signal
  in audit logs and monitoring; conflicts with the industry vocabulary that future SDK
  consumers will expect.

## Adoption sequence

| Milestone / Feature | Endpoints migrated to flat typed bodies | Mapper deletions |
|---------------------|-----------------------------------------|------------------|
| F1.1 | `listCheckpoints`, `createCheckpoint` | (none — checkpoints had no legacy mapper) |
| F1.2 (this ADR's adoption) | `renameDocument` (no body), `createNextVersion`, `presignTemplateAutosave`, `commitTemplateAutosave`, `getTemplateVersion` | `toVersionResponse` survives — still used by `createTemplate` (F1.3) |
| F1.3 | `createTemplate` (drop undeclared `id` / `version_id`) | **`toVersionResponse` retired**; `toTemplateResponse` may survive |
| F1.4 | documents-module `restoreCheckpoint`, `documents/handler.go:317` (stats), taxonomy module, audit 405 | n/a (different mappers) |
| Future templates milestone | `listTemplates`, `listVersions`, lifecycle (`submit` / `review` / `approve` / `publish`), `getTemplate`, `getDocxURL`, `presignTemplateUpload`, `placeholderCatalog` | `toTemplateResponse` retired |

The table is the canonical adoption ledger. Updating it is the responsibility of any future
feature that migrates an endpoint covered by ADR 0035.

## Verification

A change is ADR-0035-compliant when:

- The endpoint's response is declared in `openapi.yaml` with a flat top-level schema (no `data`
  wrapper).
- The handler emits the generated typed struct (no `map[string]any` literal, no
  `httpresponse.WriteJSON(w, …, doc)` against a `domain.*` value on a public route).
- The FE adapter consumes the flat shape (no `body.data.…` indirection).
- The status code is `200` for ephemeral computations (presign, search, idempotent reads), `201`
  for genuine resource creates, `204` only where the response **truly has no useful payload**
  (e.g. delete / pure rename — see F1.2 `renameDocument`, which uses `200` empty-body as a
  spec-literal pragma rather than `204`).
