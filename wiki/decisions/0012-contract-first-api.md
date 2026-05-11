# ADR 0012 — Contract-First API via oapi-codegen

> **Last verified:** 2026-05-11
> **Scope:** Decision to migrate MetalDocs backend HTTP handlers to spec-generated types; root cause analysis of the `documents.name` bug; migration scope and constraints.
> **Out of scope:** Frontend codegen (covered in `architecture/api-contract.md §6`); specific handler implementation patterns (covered in `architecture/api-contract.md §3`).
> **Key files:**
> - `api/openapi/v1/openapi.yaml:1` — spec that governs the decision
> - `internal/modules/registry/delivery/http/routes.go:43` — `AtomicCreateControlledDocument` — first handler to receive the root-cause fix
> - `migrations/0183_documents_name_not_empty.sql:1` — DB invariant floor added as part of this change
> - `wiki/backlog/contract-first-followups.md` — deferred scope (approval/taxonomy/iam/platform)

---

## Status

Accepted — 2026-05-08

Commits: `aa867b6c` (tooling bootstrap), `c968b8e0` (root-cause fix), `9fccd8e7` (registry full migration), `f7f9c58d` (templates_v2 full migration), `81e7ec23` (documents bootstrap).

---

## Context

### Root cause: the `documents.name` bug

`POST /api/v2/controlled-documents` atomically creates a controlled-document slot and a first draft document revision. The frontend correctly sent `documentName` in the request body. The backend silently dropped it.

The cause was a hand-written Go request struct that did not include a `DocumentName` field. The struct decoded incoming JSON and left `documents.name` as an empty string. The spec (`openapi.yaml`) correctly defined `documentName` as a required property — the spec was never wrong. The bug lived entirely in the gap between the spec and the Go struct.

This class of bug (spec-correct, struct-missing) is undetectable by reviewing the spec or the handler in isolation. It surfaces only at runtime when a field is silently ignored.

### Pre-migration state

- Every HTTP module maintained its own hand-written request/response Go structs alongside the spec.
- No compile-time or CI check verified that structs matched the spec.
- Adding or renaming a field in the spec required a separate, manually-tracked change to the Go struct. Any missed sync produced a silent data loss or incorrect behaviour.

---

## Decision

Migrate all HTTP handler boundaries to types generated from `api/openapi/v1/openapi.yaml` via oapi-codegen v2.

Principles:
- The spec is the **only** authoritative definition of request/response shapes.
- Hand-written request/response structs at the HTTP boundary are prohibited for migrated modules.
- New endpoints must be authored in the spec first; handler implementation follows.
- CI enforces that generated files match the spec at every PR.

---

## Alternatives considered

**Strict-decode helper only (rejected).** Adding `DisallowUnknownFields` and a required-field check to the existing hand-written handler would have fixed the specific `documents.name` bug. This was rejected because it fixes one instance while leaving the root cause — structural drift between spec and handler — intact. The next hand-written struct omission would produce the same class of bug with no prevention mechanism.

**Full StrictServerInterface pattern (deferred).** oapi-codegen also generates a `StrictServerInterface` where handlers return typed structs instead of writing to `http.ResponseWriter`. This provides stronger type safety for responses. It was not adopted in this rollout because it requires invasive changes to existing handler method signatures. The current `ServerInterfaceWrapper` pattern is a stepping stone; the strict pattern can be adopted per-module in a future pass.

---

## Consequences

### Positive

- **Drift impossible at type level** for migrated modules: if the spec changes and `go generate` is not rerun, CI fails before merge.
- **New endpoints are spec-first by workflow constraint**, not just policy.
- **Request struct omissions surface at compile time**: if a field exists in the spec but is missing in the handler implementation, the Go compiler rejects the build.

### Negative / residual risks

- **Pre-codegen modules still drift-prone.** `approval`, `taxonomy`, `iam`, and `platform` have no spec coverage and continue to use hand-written structs. They carry the same risk class as the `documents.name` bug until migrated. See `wiki/backlog/contract-first-followups.md`.
- **Required-field enforcement is not automatic.** oapi-codegen generates value vs pointer fields for required vs optional, but does not emit runtime 400 responses for missing required fields. Handlers must check explicitly. See `architecture/api-contract.md §4`.
- **Unknown-field rejection is not automatic.** Handlers must call `contracts.Decode` (`internal/modules/documents/approval/http/contracts/strictjson.go:23`) at boundaries where unknown-field rejection matters.
- **Documents handler migration blocked.** The documents module has codegen bootstrapped but handler migration is deferred due to spec-handler drift (missing spec ops, orphaned spec ops). See `wiki/backlog/contract-first-followups.md`.

---

## See also

- `wiki/architecture/api-contract.md` — operational reference: codegen commands, wiring pattern, module status table, how to add a new module
- `wiki/backlog/contract-first-followups.md` — deferred items and migration template
- [`wiki/modules/documents.md`](../modules/documents.md) — documents module: codegen bootstrap-only status; handler migration blocked by spec drift
- [`wiki/modules/documents-tech-debt.md T-002`](../modules/documents-tech-debt.md#t-002--openapi-spec-drift-on-apiv2documents-routes) — spec/handler drift detail: ops without spec, spec ops without handler, missing `operationId` on `finalizeDocument`
- [`wiki/modules/taxonomy.md §8.1`](../modules/taxonomy.md#81-authentication--authorization) — taxonomy is the residual unmigrated module: raw `net/http.ServeMux`, no spec coverage for any of its 16 routes, no oapi-codegen — explicitly called out in Consequences §"Pre-codegen modules still drift-prone" (taxonomy T-009)
