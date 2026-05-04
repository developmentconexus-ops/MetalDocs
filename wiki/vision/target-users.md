# Target Users

> **Last verified:** 2026-05-01
> **Status:** Stub. Expand with concrete personas + research findings when available.
> **Scope:** Who uses MetalDocs and what they need.

## Primary personas

### 1. Quality Engineer / Document Controller

- Owns the controlled-document catalog.
- Bootstraps taxonomy (areas, profiles, template-to-profile bindings).
- Approves templates. Often part of approval routes for documents.
- Cares about: traceability, ISO compliance evidence, no silent mutations.

### 2. Process / Procedure Author

- Subject-matter expert. Writes the actual content of a controlled document.
- Picks a controlled-document slot, fills the editor, submits for approval.
- Cares about: fast editing, reusable templates, knowing when their doc is approved.

### 3. Approver / Manager

- Reviews submitted documents in the approval inbox.
- Signs off (with password confirmation) or rejects.
- Cannot approve their own submissions (ISO segregation).
- Cares about: clear inbox, full content visibility, audit trail of their decisions.

### 4. Operator / End consumer

- Reads the published PDF to execute the procedure.
- Does not log into MetalDocs (typically).
- Receives PDFs via distribution channels.

## Org context

- ISO-bound: 9001, 14001, 45001, 27001 typically.
- Industry: manufacturing, healthcare, finance, regulated services.
- Sizes from 50–5000 employees with 100–10000 controlled documents.

## See also

- [vision/product-vision.md](product-vision.md)
- [modules/iam-rbac.md](../modules/iam-rbac.md) — capabilities that gate each role
- [concepts/iso-segregation.md](../concepts/iso-segregation.md) — segregation of duties enforcement
