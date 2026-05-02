# Workflow: Document Fill-In

> **Last verified:** 2026-05-01
> **Status:** Stub. See `workflows/user-onboarding.md` Steps 5–7 for user-facing version. Expand with editor-specific quirks + DOCX import behavior.
> **Scope:** Pick CD → wizard → editor → fill content → save.
> **Out of scope:** Submitting + approval (see `workflows/approval.md`), template authoring (see `workflows/template-authoring.md`).

## Quick summary

1. User opens **Documentos Controlados**.
2. Clicks **Novo Documento Controlado** → picks profile + area + title → CD created with auto-generated code.
3. From the CD detail → **Gerar Documento** → wizard confirms the bound template → first version cloned in `draft` state, eigenpal editor opens.
4. User edits content. Does NOT touch the 7 fixed-token chips.
5. Save persists the draft. F5 confirms persistence.

See [workflows/user-onboarding.md](user-onboarding.md) for the full step-by-step click path.

## Edge cases (TBD)

- Importing a `.docx` into the editor — known eigenpal bugs around table-in-header (parked).
- Switching templates mid-edit — not supported; would need a new version.
- Auto-save behavior + conflict resolution (ETag / optimistic concurrency).

## See also

- [modules/documents-v2.md](../modules/documents-v2.md)
- [modules/editor-ui-eigenpal.md](../modules/editor-ui-eigenpal.md)
- [concepts/controlled-documents.md](../concepts/controlled-documents.md)
