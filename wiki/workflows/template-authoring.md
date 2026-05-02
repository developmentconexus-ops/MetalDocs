# Workflow: Template Authoring

> **Last verified:** 2026-05-01
> **Status:** Stub. See `workflows/user-onboarding.md` Steps 2–4 for the user-facing version. Expand here with edge cases (versioning while drafts in flight, retroactive binding changes, etc.).
> **Scope:** create → edit → submit → approve → publish → bind to profile.
> **Out of scope:** End-user document fill-in (see `workflows/document-fillin.md`).

## Quick summary

1. Author clicks **Novo Template**, fills the dialog, lands in eigenpal author with `v1 draft`.
2. Author edits content (uses the 7 fixed tokens literally as chips), saves, submits for review.
3. Reviewer (different user — ISO segregation) approves.
4. Admin publishes the version.
5. Admin binds the published version as the **default template** of one or more profiles via taxonomy.

See [workflows/user-onboarding.md](user-onboarding.md) for the full step-by-step click path.

## Edge cases (TBD — fill in after testing)

- Editing a published version → must create a new version, not in-place edit.
- Multiple drafts coexisting → which one is currently editable, which is approved.
- Re-binding a profile to a different template version → effect on existing vs new documents.
- Deprecating an old template version → does it auto-unbind from profiles?

## See also

- [modules/templates-v2.md](../modules/templates-v2.md)
- [modules/editor-ui-eigenpal.md](../modules/editor-ui-eigenpal.md)
- [concepts/placeholders.md](../concepts/placeholders.md)
