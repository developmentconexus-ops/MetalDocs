# F3.3 — Plan (how)

1. **Fixture prep (test-only).** The spy's `Create` returns `ID:"cd-1"` (not a UUID) + empty
   `TenantID` — fine for the raw-map handler, but the typed mappers `uuid.Parse` both. Make the spy
   configurable: add `createResult *application.CreateResult`; default to a UUID-valid result
   (`sampleControlledDocument()` + a UUID `DocumentRef`). This keeps `TestAtomicCreate_ForwardsGeneratedOnlyFields`
   green through the typed path.
2. **TDD red.** Add `TestAtomicCreate_UsesGeneratedResponse`: configure `createResult` with a doc that
   has **nil** `DepartmentCode`/`SequenceNum`/`OverrideTemplateVersionID` (valid UUID `Id`/`TenantId`).
   Assert 201, body unmarshals into `AtomicCreateResponse`, ids match, and the body string contains
   **no** `department_code`/`override_template_version_id`/`sequence_num`. Fails on current raw map
   (domain emits `…:null`).
3. **Implement.** Swap `routes.go:123` map for `controlleddocumentsapi.AtomicCreateResponse` via
   `controlledDocumentResponse(*res.ControlledDocument)` + `documentRefResponse(res.DocumentRef)`;
   mapper error → `WriteError(500,"INTERNAL_ERROR",…)` (mirror Get).
4. **Green + gates.** Run G1–G6 from spec.md.
5. **evidence.md.** Record commands + real output; label fixture vs real; note no FE regen.
