---
name: runtime-contract-prereq
description: Use when a MetalDocs task discovers startup drift, migration drift, auth/session failure, route mismatch, or runtime/OpenAPI/generated/frontend-wrapper mismatch that must be repaired before feature work continues.
---

# Runtime + Contract Prerequisite Audit

Read and follow `.claude/skills/runtime-contract-prereq/SKILL.md`.

If `.claude/skills/runtime-contract-prereq/SKILL.md` is missing, moved, or unreadable, stop feature work and treat this as a `workflow/tooling gap`. Surface the missing bridge path and continue only with the required sources in this file plus `wiki/architecture/*` truth docs until the bridge is repaired.

## Required sources

- `.claude/skills/runtime-contract-prereq/SKILL.md`
- `wiki/architecture/backend-api-structure.md`
- `wiki/architecture/api-contract.md`
- `wiki/architecture/frontend-structure.md` when frontend wrappers are involved

This bridge exists so Codex sessions can discover the canonical prerequisite workflow and source of truth.

## Freeze checks (required)

- Generated package present for each touched generated public module surface (example: controlled-documents at `internal/modules/controlleddocuments/api/` with `controlleddocumentsapi` artifacts).
- Runtime mounted through generated wrapper/`HandlerWithOptions` ownership for each touched generated public route surface (example: `/api/v1/controlled-documents*`), not raw public mux ownership.
- Runtime/spec/generated/backend/generated-frontend-wrapper alignment is provable for touched endpoints before feature work proceeds.
- Canonical module/product naming remains controlled-documents unless a legacy name is explicitly labeled as historical/migration context.

## Required mismatch classification

Before continue/stop, classify each mismatch at minimum as one of:

- runtime prerequisite
- shared contract prerequisite
- module-local implementation
- screen-local implementation
- wiki-memory drift
- workflow/tooling gap
- defer

## Required drift-class stop checks

Stop and surface prerequisite work when any of these are present:

- duplicate OpenAPI namespaces for the same public capability
- runtime/spec path mismatch for the same endpoint surface
- generated backend package exists but wrapper/`HandlerWithOptions` is not mounted for that public module surface
- frontend uses handwritten types for generated routes where generated wrappers/types should own the contract
- permission guard references a non-existent public route
- wiki status contradiction that changes planning/prerequisite decisions

## Required evidence checks (execute before continue/stop decision)

Run these checks and attach concrete evidence (command output, file path, or route table row) to each mismatch classification:

- Duplicate OpenAPI namespaces:
  `rg -n "^[[:space:]]*/api/v1/" api/openapi/v1/openapi.yaml`
  and validate conflicting ownership/tag blocks.
- Runtime/spec path mismatch:
  `rg -n "Handle|HandleFunc|/api/v1/" internal/modules/*/delivery/http`
  then compare against `api/openapi/v1/openapi.yaml` path entries for the same endpoint.
- Generated package present but not mounted:
  `rg -n "package .*api|type ServerInterface|HandlerWithOptions|ServerInterfaceWrapper" internal/modules/*/api internal/modules/*/delivery/http`
  and prove generated package exists without generated boundary mount.
- Frontend handwritten type drift on generated routes:
  `rg -n "type .*Response|interface .*Response|as .*paths\\[" frontend/apps/web/src/features frontend/apps/web/src/lib`
  then compare to generated `frontend/apps/web/src/lib/api-types/index.d.ts` route types.
- Permission guard on non-existent public route:
  `rg -n "permission|authorize|authz|guard|/api/v1/" internal/modules`
  and prove referenced route is absent from runtime route registration and OpenAPI.
- Wiki status contradiction affecting planning:
  collect the conflicting module wiki status/doc section and pair it with runtime/spec/generated evidence above.

Stop if feature work is trying to continue through a failing prerequisite boundary.
Stop if canonical guidance is missing or conflicts with the required sources.
Stop if contradictions exist across runtime ownership, OpenAPI contract ownership, generated package identity, or frontend wrapper/type expectations for the same endpoint surface.
