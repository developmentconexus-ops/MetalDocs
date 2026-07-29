# ADR 0086 — Template approval routes: config-first, profile-keyed

> **Status:** Accepted 2026-07-29 (operator ruling D6 of 2026-07-28: "rota de
> template na configuração"; Codex-aligned rev 4 — rounds 1–3 NOT-ALIGN
> findings incorporated, round 4 ALIGN)
> **Supersedes:** the per-instance template-route keying (`subject_key =
> template_id`) introduced by ADR 0082's P2.S3 wire mapping, and migration
> 0297's `approval_routes_template_subject_projection_check` (profile_code
> forced NULL on template routes).
> **Extends:** [ADR 0082](0082-approval-kernel-extraction.md) (subject-generic
> kernel — two-level keying preserved), [ADR 0081](0081-per-profile-signature-policy.md)
> (per-profile RoutePolicy now applies to template routes too),
> [ADR 0083](0083-subject-discriminated-capability-tripwire.md) (tripwire arms
> regenerate for the new projection).
> **Relates to:** usability tracker 2026-07-28 D2 gate (documents) — this ADR
> is its template-side symmetry; Etapa 4.1 journey evidence (2026-07-29).
> **Scope:** usability-remediation Etapa 4 (D6). Implementation is a separate
> unit gated on this ADR's acceptance.

## Context

A template approval route is keyed **per template instance**
(`subject_key = template_id`, `internal/modules/approval/domain/subject.go:47-57`).
Consequences, all proven live in the Etapa 4.1 journey (tracker 2026-07-29):

- Nobody creates the route automatically. For **every** template, an admin must
  `POST /approval/routes` with `subject_kind=template` and the template's UUID
  — and the FE route-builder cannot express that (`routeDraft.ts:12-16,114`
  requires `profileCode`), so the only path is raw API. A tenant configuring
  MetalDocs from zero cannot make templates approvable through the UI at all.
- The gate exists only at submit (`409 APPROVAL_ROUTE_MISSING`), not at
  creation — asymmetric with documents, which since Etapa 1 hard-block
  creation without an active route (D2, `90a6fae3`). An author can build a
  template that is structurally unapprovable and only finds out at submit.
- The route can only be created **after** the template exists (its UUID is the
  key), which is backwards for a governance object: approval governance is
  tenant configuration, not per-artifact ceremony. This is the direct cause of
  the operator's D6 ruling.
- Template routes skip per-profile RoutePolicy resolution entirely
  (`route_admin_service.go:258-268` — "a template subject has no profile"),
  so the 5-rule signature-policy model (ADR 0081) does not govern template
  approval shapes. Migration 0297's CHECK **forbids** `profile_code` on
  template routes, hard-wiring the gap.

Meanwhile `templates_template.doc_type_code` is already the taxonomy class key
(== profile code) used by `SetDefaultTemplate` + `IsPublished`
(`profile_service.go:277-333`, `template_version_reader.go:15-62`) and by the
new-document wizard's template filter. The natural config-first key already
exists on every governed template.

Patching the FE route-builder to speak `template_id` keying would lock in the
local maximum. The global-maximum structure — what Qualio/MasterControl do —
is **document-type-level template governance**: one template approval route per
profile, configured before any template exists.

## Decision

1. **Re-key template routes to the profile.** `subject_kind = 'template'`,
   `subject_key = doc_type_code` (the profile code). `profile_code` becomes
   **NOT NULL** on template routes (FK to `document_profiles`), identical to
   document routes; migration 0297's projection CHECK is replaced by the
   symmetric constraint (`subject_kind='template'` ⇒ `profile_code = subject_key`,
   mirroring the document-subject invariant in `resolveCreateRouteSubject`).
2. **RoutePolicy applies — on create AND update.** Template routes resolve
   the same per-profile signature policy as document routes (ADR 0081). Both
   template-branch skips in `route_admin_service.go` are deleted: the create
   path (`:266-271`) and the update path (`:449-457`), which today also
   bypasses policy. One kernel, one policy source, both write paths.
3. **Creation is hard-gated.** `POST /templates` requires an active template
   route for the declared `doc_type_code`, checked in-tx (same mechanism and
   error as documents: `409 APPROVAL_ROUTE_MISSING`). The submit-time check
   remains as the fail-closed backstop.
4. **`doc_type_code` becomes required** on template creation (422 when
   absent). Generic user templates are exterminated, not grandfathered.
   This is a deliberate breaking product change, not dev-data cleanup:
   today creation accepts an omitted `doc_type_code`
   (`internal/modules/templates/delivery/http/routes_generated.go:54-65`)
   and list-by-profile **deliberately includes** `doc_type_code=''` generic
   templates for every profile
   (`internal/modules/templates/infrastructure/postgres.go:163-186`). Both
   branches die with the cutover: creation 422s without a profile, and the
   generic-inclusion `OR` branch is removed from the list query. Rationale:
   a template that applies "to every profile" has no single approval
   authority, which is exactly the ambiguity config-first keying exists to
   kill (extermination principle — no relax-to-optional). The system blank
   template (`__system_blank__`, `doc_type_code = 'system'`) is
   system-owned, never enters tenant approval, and is exempt from the gate.
5. **Two-level keying is preserved** (ADR 0082): `approval_instances` remain
   keyed by the template version's identity. Only the **route** subject key
   changes.
6. **Resolution follows the key on every surface.** All three template-route
   resolution sites move from `template_id` to `doc_type_code`:
   - submit: `internal/modules/approval/application/template_submit_service.go:221-241`
   - approval preview: `internal/modules/approval/application/route_preview.go:104-130`
     — its input gains the profile key resolved through the **same** read
     port (version → `doc_type_code`), not a handler-side shortcut: the
     preview handler (`routes_approval_preview.go:43-48`) today discards
     the resolved version before calling preview; preview and submit MUST
     resolve identically or drift is guaranteed
   - HTTP submit handler selector:
     `internal/modules/templates/delivery/http/routes_approval_kernel.go:79-86`
     (today supplies only the template UUID).

   The approval-owned template read port
   (`internal/modules/templates/infrastructure/approval_version_reader.go:38-72`)
   exposes only status/content today; it is **extended to project
   `doc_type_code`** so the submit tx resolves the route in-tx from the
   subject itself — via the templates module's published interface, never
   its repository (module-boundary rule). Fail closed: a version whose
   template lacks a `doc_type_code` (impossible post-cutover, constraint-
   backed) is a defect, not a fallback path.
7. **Contract-first changes are part of the decision** (spec is route truth):
   - **Route-admin contract:** the **create-route** schema and validation
     (`internal/modules/approval/http/contracts/route.go:150-159`) today
     **require** `profile_code` to be absent on template routes — inverted
     post-cutover: `profile_code` required and validated `== subject_key`
     for `subject_kind=template`, same as documents. OpenAPI schema,
     generated types, request validation, and handler mapping all move
     together (full regen, no partial). `UpdateRouteRequest` carries no
     profile/subject fields (immutable identity, `route.go:160-194`) and
     stays that way — the update path's change is service-side only:
     RoutePolicy validated against the route's **persisted** profile
     (Decision 2).
   - **Create-template contract:** `POST /templates` in
     `api/openapi/v1/openapi.yaml` makes `doc_type_code` required and
     declares the 422 Problem response; regen; handler maps missing/blank
     to the declared 422 (no undeclared status codes). The creation-gate
     409 (`APPROVAL_ROUTE_MISSING`) is likewise declared.
8. **FE follows config-first.** Route-builder v2 gains a template-route mode
   keyed by profile (under Administração → Rotas, same screen family as
   document routes); the template creation UI gets the same
   readiness-gate messaging/badge pattern documents got in Etapa 1.

## Hard cutover (no fallback)

Per the extermination principle (no dual-read, no relax-to-optional):

- Migration deletes existing `subject_kind='template'` routes and their
  stages. This is dev-only data: live DB holds exactly the two Etapa 4 QA
  routes (`71c374c5`, `8139a455`) plus test leftovers; no production tenants
  exist (F-18 fresh-repo-at-release posture).
- The two generic dev templates (`qa1-template-approval`, `BDOCX-verify-409`,
  both QA leftovers with no `doc_type_code`) are deleted by the same
  migration. No reserved "generic" subject key is minted.
- Route resolution has exactly one lookup path post-cutover. A template whose
  profile has no active template route cannot be created (new) or submitted
  (pre-existing drafts) — fail closed, friendly 409 first, DB constraint as
  backstop.
- ADR 0083 tripwire arms are regenerated from the Go registry for the new
  projection (generation, not hand-sync, per GMR M2). The signoff-parent
  discriminator is **unchanged**: instances stay version-keyed and template
  signoff keeps its template-specific capability
  (`db/migrations/0299_tripwire_subject_discriminated_arms.sql:47-65`,
  `0300_tripwire_signoff_parent_discriminator.sql:275`) — only the route
  projection arm (0297's NULL-profile CHECK successor) changes.

## Consequences

- A tenant configures template governance once per profile, at config time,
  before any template exists — symmetric with document routes and with the
  D1 config-first onboarding ruling.
- The FE route-builder needs no per-template UUID plumbing; its existing
  profile-keyed model extends to a subject-kind toggle.
- Existing dev templates scoped to a profile (`it`) keep working: their
  profile's template route governs all their future versions.
- Template approval shapes come under the per-profile signature policy —
  closing the ADR 0081 coverage gap.
- Etapa 4.1's journey pain (raw-API route per template) disappears; the
  journey's admin ceremony reduces to: configure route once → authors create
  templates freely under it.

## Evidence

Etapa 4.1 live journey (tracker 2026-07-28, entries of 2026-07-29): template
creation ungated; approval-preview `route_id: null` until an admin hand-built
route `71c374c5` via raw API keyed to template UUID `0bad8489`; second
template required a second hand-built route `8139a455`; FE cannot express
either. Full E2E succeeded only with three seeded actors and two raw-API
route creations.
