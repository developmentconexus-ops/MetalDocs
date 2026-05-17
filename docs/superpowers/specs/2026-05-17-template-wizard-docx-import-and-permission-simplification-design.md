# Template Wizard DOCX Import and Permission Simplification

> Date: 2026-05-17
> Status: approved direction, pending implementation plan
> Scope: template wizard, templates API/contract, templates persistence cleanup, Eigenpal DOCX import handoff, module wiki sync

## 1. Problem

The template wizard currently models template "permissions" as if a template creator can decide who may use a template when creating documents. That creates a product and authorization problem.

Example: a user is authorized to create document type `DC`. There are four published templates linked to `DC`. If one template creator restricts one of those templates to a single area, the user may lose access to a template that is valid for their document workflow even though their role already allows document creation.

That makes template visibility an accidental second authorization layer. It can conflict with the existing role/capability system and can make template selection feel arbitrary.

The wizard also needs to make DOCX import real. Selecting a `.docx` in the wizard should not only store file name metadata. It must upload the DOCX and commit it to the first template version so Eigenpal opens with the imported document rendered.

## 2. Decision

Remove template-use permissions from the template wizard and from the public product contract for now.

MetalDocs should use this separation:

- Roles and capabilities decide who can create, edit, review, approve, publish, archive, and manage templates and documents.
- Document type/profile decides which templates are valid choices for a new document.
- Template creator preference must not hide otherwise valid templates from authorized document authors.
- Future template restrictions, if needed for compliance, should be modeled as controlled "applicability" rules owned by governance/admin flows, not as creator-level wizard permissions.

This keeps authorization simple and auditable while avoiding a second hidden gate inside template metadata.

## 3. Evidence and Rationale

NIST RBAC describes permissions through users, roles, operations, and objects. That matches the MetalDocs role/capability model for actions such as `template.create`, `template.edit`, `template.review`, `template.approve`, and document creation.

NIST ABAC is useful for contextual decisions that combine subject, object, action, and environment attributes. It remains a possible future model for governed template applicability, but it should not be introduced casually through a creator-scoped wizard field.

OWASP authorization guidance recommends explicit authorization and validating permissions on every request. Hiding important policy only in frontend template filters would weaken that model.

Sources:

- NIST RBAC: https://csrc.nist.gov/Projects/Role-Based-Access-Control
- NIST ABAC SP 800-162: https://csrc.nist.gov/pubs/sp/800/162/upd2/final
- OWASP Authorization Cheat Sheet: https://docs.devnetexperttraining.com/static-docs/OWASP-Cheat-Sheet-Series/cheatsheets/Authorization_Cheat_Sheet.html

## 4. User Experience

The template wizard should keep a professional authoring flow:

1. Identify the template.
2. Link it to the intended document type/profile.
3. Choose a starting point:
   - blank document
   - import `.docx`
4. Confirm and create.
5. Open Eigenpal with the correct starting content.

The wizard should not show a "who can use this template" step.

If the user chooses blank, the editor opens with an Eigenpal blank document.

If the user imports `.docx`, the wizard uploads and commits the DOCX before navigating to the template editor. The editor then fetches the stored DOCX for version 1 and Eigenpal renders it.

## 5. Backend and API Design

The templates API should accept template creation data needed for identity and document-type/profile linkage, but not creator-scoped use permissions.

Current template visibility fields must be audited before removal:

- domain constants such as `public`, `internal`, and `specific`
- create-template command fields such as `visibility`, `areas`, and `specific_areas`
- OpenAPI request/response schema fields
- repository list filters that hide templates by visibility or area
- database columns and indexes, if present
- frontend generated types and wrappers

Target behavior:

- Listing templates for authoring/selection should be driven by tenant, lifecycle state, archived state, and document type/profile filters.
- Mutation permissions remain enforced through existing role/capability checks.
- The backend must not grant document creation based on template metadata.
- The backend must not hide valid document-type templates based on creator-scoped visibility metadata.

If database columns remain temporarily for migration safety, they should become inert compatibility fields with a clear removal path. The preferred final state is to remove unused visibility persistence once schema impact is understood.

## 6. DOCX Import Design

The wizard must store the selected DOCX as a real `File` in local wizard state until submit.

On submit with `startingPoint = docx`:

1. Create the template and version 1 through the templates API.
2. Request the existing template DOCX autosave/upload presign endpoint for version 1.
3. Upload the selected DOCX bytes directly to object storage through the presigned URL.
4. Compute or obtain the required content hash according to the existing autosave contract.
5. Commit the autosave/import to version 1.
6. Navigate to `/templates/{templateId}/versions/1`.
7. The editor uses the existing version DOCX URL flow and renders the imported file in Eigenpal.

On submit with `startingPoint = blank`:

1. Create the template and version 1.
2. Navigate to `/templates/{templateId}/versions/1`.
3. Eigenpal receives the blank document fallback already implemented in `packages/editor-ui`.

Failure behavior:

- If template creation fails, stay in the wizard and show the API error.
- If DOCX upload or commit fails after template creation, show a recoverable import error and do not silently open a blank editor as if import succeeded.
- The user should be able to retry the import or open the created template as blank only through an explicit action.

## 7. Frontend Design

Template wizard changes live under `frontend/apps/web/src/features/templates/`.

Expected changes:

- Remove the permissions/disponibilidade wizard step from the template creation flow.
- Remove reducer state that represents creator-scoped template usage permissions.
- Replace current DOCX placeholder metadata with real selected file state.
- Keep file name/size for display only.
- Add a submit workflow that branches between blank creation and DOCX import creation.
- Use feature API wrappers under `features/templates/api/`.
- Use generated API types from `lib/api-types/`.
- Use `lib/api/client.ts` / `apiFetch`; do not call authenticated API routes directly with raw `fetch`.
- Use TanStack Query only where server state hooks are needed; wizard submit can remain a mutation-style action if that matches the current feature pattern.

The document wizard visibility controls can remain in documents. They should not be copied into the template wizard for this feature.

## 8. Database Design

Database work should remove or neutralize template-use visibility persistence only after runtime and contract truth are mapped.

Implementation must classify the database change before editing:

- If visibility columns are unused after API/frontend cleanup, add a forward migration that removes them and update dictionary/wiki coverage.
- If removal is too risky because downstream code still reads them, add a forward migration or code change that makes the fields default/inert, then create a follow-up debt item for physical removal.

Historical migrations must not be rewritten.

## 9. Out of Scope

- Document wizard visibility changes.
- New ABAC policy engine.
- New admin UI for governed template applicability.
- Role/capability redesign.
- Approval lifecycle redesign.
- Template publishing workflow changes except where required to keep imported DOCX bytes visible in Eigenpal.

## 10. Verification

Backend/API:

- `npx @redocly/cli lint api/openapi/v1/openapi.yaml`
- `$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/templates/api/...`
- `go test ./internal/modules/templates/... -count=1`
- `go build ./...`

Frontend:

- `cd frontend/apps/web`
- `pnpm gen:api`
- `pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- targeted Vitest tests for template wizard reducer/page/API wrapper

Editor:

- Keep existing `packages/editor-ui` blank Eigenpal tests passing.
- Validate imported DOCX opens in the in-app browser at `/templates/{id}/versions/1`.

Database/wiki:

- Run the smallest applicable database dictionary/bootstrap checks if a migration or dictionary page changes.
- Sync `wiki/modules/templates.md` and related template artifacts after implementation.

## 11. Open Implementation Questions

These questions must be answered by runtime/code inspection during planning:

1. Are template visibility fields still required by any downstream module, seed, or list query?
2. Does the current autosave commit contract require the frontend to compute SHA-256, or can the backend verify from object storage?
3. Is the existing wizard submit path already structured as a mutation that can compose create, presign, upload, commit, and navigate?
4. Which OpenAPI schemas expose `visibility`, `areas`, or `specific_areas`, and are they safe to remove immediately?

## 12. Success Criteria

- Template wizard no longer asks who can use the template.
- Template creation remains protected by role/capability authorization.
- Document authors see valid published templates by document type/profile, not by creator-scoped template visibility.
- Imported DOCX files selected in the wizard open rendered in Eigenpal on the created template version.
- Blank templates still open as a blank Eigenpal document.
- OpenAPI, generated backend code, generated frontend types, runtime handlers, database schema, frontend code, and wiki memory agree.
