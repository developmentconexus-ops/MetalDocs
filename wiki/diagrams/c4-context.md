# C4 Level 1 — System Context

> **Last verified:** 2026-06-01
> **Scope:** External actors + MetalDocs as a single system.
> **Source of truth for:** [`wiki/architecture/system-overview.md`](../architecture/system-overview.md).

```mermaid
C4Context
    title MetalDocs — System Context (ISO 9001 controlled-document QMS)

    Person(author, "Document author", "Drafts, fills in, edits controlled documents")
    Person(approver, "Approver", "Reviews and signs off documents per the approval route")
    Person(admin, "Tenant admin", "Manages users, roles, templates, taxonomy")
    Person(reader, "Reader", "Consumes approved documents and exports PDFs")

    System(metaldocs, "MetalDocs", "Multi-tenant SaaS for controlled-document lifecycle: templates → drafts → approval → frozen artifact → PDF")

    System_Ext(idp, "Identity (future)", "External IdP for SSO; today MetalDocs owns its own auth")
    System_Ext(browser, "User's browser", "Loads the SPA, talks directly to MetalDocs object storage via presigned URLs")

    Rel(author, metaldocs, "Creates and edits documents in the browser editor")
    Rel(approver, metaldocs, "Signs off via approval inbox")
    Rel(admin, metaldocs, "Authors templates, manages users + areas")
    Rel(reader, metaldocs, "Views/exports approved documents")
    Rel(metaldocs, browser, "Serves SPA + issues short-lived presigned PUT/GET URLs")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

## What this diagram answers

- **Who uses it?** Four user roles: author, approver, admin, reader.
- **What's outside?** Only the user's browser (which talks to object storage directly via presigned URLs) and a future SSO IdP. Everything else is internal.
- **What's the system's job?** Take a controlled document from draft → approved → archived artifact, with ISO 9001 traceability.

## Key invariants visible here

- **Browser talks to object storage directly.** The API issues presigned URLs; multi-megabyte docx bytes never pass through the API. This is the scaling pattern. See [sequence-edit-autosave.md](sequence-edit-autosave.md).
- **No external IdP today.** MetalDocs is the identity authority for v1.

For the moving parts inside the box, see [c4-container-backend.md](c4-container-backend.md).
