# Workflow: User Onboarding (End-to-End)

> **Last verified:** 2026-05-02
> **Note (smoke test 2026-05-01):** Non-admin users require `user_process_areas` entries to exercise approval authz. See "Common pitfalls" below.
> **Scope:** The full MetalDocs user journey from a brand-new tenant to a published, frozen document with PDF fanout. Written from the user's seat — what to click, in what order, why each step matters. Non-technical: no code, no API endpoints, no DB tables.
> **Out of scope:** Implementation details (see `modules/*.md`), API reference, deployment.
> **Audience:** New users, QA testers running smoke flows, product/support staff explaining the product.
> **Key UI files (for cross-ref only — do not need to read them as a user):**
> - `frontend/apps/web/src/features/taxonomy/TaxonomyAdminPage.tsx` — Tipos Documentais admin
> - `frontend/apps/web/src/features/taxonomy/ProfileEditDialog.tsx` — perfil editor + template binding
> - `frontend/apps/web/src/features/templates/v2/TemplatesListPage.tsx` — Templates list
> - `frontend/apps/web/src/features/templates/v2/TemplateAuthorPage.tsx` — eigenpal author
> - `frontend/apps/web/src/features/registry/RegistryListPage.tsx` — Documentos Controlados list
> - `frontend/apps/web/src/features/documents/v2/DocumentCreatePage.tsx` — wizard
> - `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` — eigenpal fill-in editor
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` — Caixa de Entrada de Aprovação

---

## Mental model (read this first)

MetalDocs is an ISO-bound controlled-document platform. Three nouns drive everything:

1. **Template** — a reusable DOCX-style layout with placeholders. Authored once, versioned, approved, published.
2. **Profile / Tipo Documental / Perfil Documental** — *the same thing under three names*. A category like "Procedimento Operacional", "Descrição de Cargo", "Política de Qualidade". A profile decides which **template** new documents of that type start from. One template can serve multiple profiles.
3. **Controlled Document (Documento Controlado)** — a unique, code-numbered slot in the catalog (e.g. `POP-RH-001`). It binds a profile + an area + a sequence number, then spawns one or more **document versions** (the actual editable instance the end user fills in).

Lifecycle terms (don't confuse them):

| Term | What it means |
|------|---------------|
| **Publish** | Template version becomes selectable when creating documents. Template-only term. |
| **Approve** (template) | Template version reviewed and approved by required signers. Precedes publish. |
| **Approve** (document) | Document version got all required signoffs. Triggers freeze automatically. |
| **Freeze** | Document version becomes immutable. All 7 fixed tokens resolve to final values. DOCX rewritten with substitutions. |
| **Fanout** | After freeze, async PDF generation via Gotenberg. PDF stored in S3/MinIO. |

The 7 fixed tokens (memorize): `{doc_code}`, `{doc_title}`, `{revision_number}`, `{author}`, `{effective_date}`, `{approvers}`, `{controlled_by_area}`. Templates use these literal names. They auto-resolve at freeze — users never type their values.

---

## The full flow (9 steps)

### Step 1 — Admin sets up taxonomy

**Who:** Admin (or whoever has `taxonomy:manage` capability).
**Where:** Sidebar → **Tipos Documentais**.
**What you do:**

1. The page opens on the **Famílias** tab (default). Create at least one document family (e.g. `qualidade — Qualidade`, `rh — Recursos Humanos`). Families are globally scoped — shared across all tenants. A profile cannot be created without a valid family.
2. Switch to the **Áreas** tab. Create the areas your org uses (e.g. `RH`, `Qualidade`, `Produção`). Each area gets a code + name. The code becomes part of the controlled-document number.
3. Switch to the **Perfis** tab. Create profiles like `POP — Procedimento Operacional`, `DC — Descrição de Cargo`. Each profile requires a family (selected from the dropdown). The profile's code prefix drives `{doc_code}`.

**Why:** Without at least one family, one area, and one profile, nobody can create a controlled document downstream. This is bootstrap work — usually done once per tenant (families may already exist from prior setup since they are global).

**Validation:** All three tabs list what you just created. No error toasts.

---

### Step 2 — Template author creates a template

**Who:** Anyone with `templates:author` capability.
**Where:** Sidebar → **Templates**.
**What you do:**

1. Click **Novo Template** (or "New Template").
2. Fill the dialog: title (e.g. "DC — Descrição de Cargo"), description, target profile (optional but recommended).
3. Confirm. The system creates the template + a first **version** in `draft` state and opens the eigenpal author.

**In the author (eigenpal):**

- Paste or type the layout. Use the 7 fixed tokens literally where you want substitutions: `{doc_code}` in the header, `{doc_title}` in the title block, `{author}` in the signature line, etc.
- You can import a starting `.docx` if you have one.
- Save often — the **Save** button persists the current version's content as `draft`.

**Why:** This is the reusable layout. Every document of this profile will start as a clone of this template's published version.

**Validation:** Refresh the page — your edits persist. The version list shows `v1 (draft)`.

---

### Step 3 — Template version goes through lifecycle

**Where:** Same template author page → **Versão** panel on the right.

**Lifecycle:** `draft → in_review → approved → published`

1. **Submit for review** — moves `draft → in_review`. Author can no longer edit content.
2. **Approve** — reviewer with `templates:approve` capability clicks Aprovar. State becomes `approved`.
3. **Publish** — admin clicks Publicar. State becomes `published`. Now this version is the one new documents will clone.

**Why:** Same ISO-segregation logic as documents — the author of a template version cannot also be its approver. Publishing is the gate that makes a template usable; before publish, downstream document creation cannot pick this version.

**Validation:** On the Templates list, the row shows `v1 published`. The "create document" wizard later will offer this template.

---

### Step 4 — Admin links template to profile (THE CRITICAL STEP)

**This is the step most users miss.** Without it, the document wizard cannot pre-fill a template for a given profile.

**Who:** Admin / taxonomy manager.
**Where:** **Tipos Documentais** → **Perfis** tab → click a profile row → **Editar**.
**What you do:**

1. The profile edit dialog opens.
2. Find the **Template padrão** (default template) selector.
3. Pick the template version you just published in Step 3.
4. Save.

**Why:** This wires `Profile → Template`. When a user later picks a profile during document creation, the wizard knows which template to clone. Without this binding the user gets a blank-slate document or an error.

**Important nuance:** Re-binding a profile to a different template version only affects **future** documents. Existing controlled documents keep their original template binding (snapshot is taken at document creation).

**Validation:** Edit the profile again — the template selector shows the bound template + version.

---

### Step 5 — User registers a Controlled Document (the catalog slot)

**Who:** Document author (any user with `registry:create` capability for the target area).
**Where:** Sidebar → **Documentos Controlados** (Registry).
**What you do:**

1. Click **Novo Documento Controlado**.
2. Pick the **Perfil** (Profile) — e.g. `DC — Descrição de Cargo`.
3. Pick the **Área** — e.g. `RH`.
4. Provide a title (free text) — e.g. "Descrição de Cargo — Analista Fiscal".
5. Confirm.

**What happens behind the scenes (worth knowing):**

- The system generates the unique code: `{profile-prefix}-{area-code}-{sequence}`, e.g. `DC-RH-001`.
- The sequence counter is per `(profile, area)` pair — so `DC-RH-002` is the next one in RH, but `DC-QUA-001` is the first in Qualidade.
- The CD is now in the catalog but has **no editable version yet**.

**Validation:** The Documentos Controlados list shows the new row with its code, profile, area, no current revision.

---

### Step 6 — User generates the working document version

**Where:** From the Documentos Controlados list → click the row → **Gerar Documento** (Generate).

**What you do:** The wizard opens. You confirm the template (already selected via the profile binding from Step 4). Click **Gerar**.

**What happens:** A new document **version** is created in `draft` state, cloned from the published template version. The eigenpal editor opens.

**Validation:** Editor URL has a version ID. The layout matches the template. Tokens like `{doc_code}` show as placeholder chips, not yet resolved.

---

### Step 7 — User edits content in the eigenpal editor

**Where:** Document editor page (opened automatically from Step 6).
**What you do:**

- Type / paste the actual content for this document.
- Apply formatting (headings, bold, lists, tables) as needed.
- The 7 fixed tokens stay as chips — **do not type values for them**. They resolve automatically at freeze.
- Save often.

**What you do NOT do:**

- Do not edit the template structure for one-off changes — that requires authoring a new template version (Step 2 again).
- Do not type the doc code, author name, effective date, etc. by hand — that's what the tokens are for.

**Validation:** Refresh the page; edits persist. Editor toolbar shows save status.

---

### Step 8 — Submit for approval

**Where:** Document editor → **Finalizar** (or Submit) button.

**What happens:**

- The backend resolves the active approval route for the document's profile, then in a single transaction: document state moves `draft → under_review` **and** an approval instance is created with pre-populated eligible actor lists.
- Approvers can immediately see the request in their **Caixa de Entrada** inbox — no manual DB patching required.

**Approver side:**

1. Approver opens **Caixa de Entrada de Aprovação** (Approval Inbox).
2. Sees the pending request with document title + code.
3. Clicks the row → reviews content → clicks **Aprovar** or **Rejeitar**.
4. **Aprovar** prompts for password confirmation (anti-mistake guard).
5. After all required signoffs are collected (per the route's quorum: `any_1`, `m_of_n`, etc.), state moves to `approved` and **freeze fires automatically**.

**ISO segregation:** The user who submitted cannot also approve. The system enforces this — approvals from the submitter are blocked at the API.

**Validation:** Document detail page shows `approved` state, signoff list with timestamps + approver names. Inbox no longer shows the request.

---

### Step 9 — Freeze and fanout (automatic, but observable)

**No user action needed.** Freeze fires inside the same transaction as the final signoff.

**What freeze does:**

- Resolves the 7 fixed tokens to final values (`{doc_code}` → actual code, `{author}` → first signed-off author, `{effective_date}` → freeze timestamp, `{approvers}` → comma-separated approvers, etc.).
- Rewrites the DOCX with all substitutions baked in.
- Stores the frozen DOCX artifact in S3/MinIO.
- Marks the document version `frozen` and immutable.

**Fanout (async, seconds-to-minutes after freeze):**

- An outbox worker picks up the freeze event.
- Calls Gotenberg to convert the frozen DOCX → PDF.
- Stores the PDF in S3/MinIO alongside the DOCX.
- The document detail page shows download links for both DOCX and PDF.

**Validation:**

- Document detail page shows state `frozen`, `values_frozen_at` timestamp, hash fingerprints.
- After ~30s the **PDF** download link appears (refresh if needed).
- Open the PDF — all 7 tokens are real values, not chips.

---

## Common pitfalls

1. **Skipping Step 4 (profile-to-template binding).** Symptom: wizard shows no template options or generates a blank document. Fix: bind the published template to the profile.
2. **Trying to publish a template version while still in `draft`.** Submit for review first, then approve, then publish. The button is disabled until prerequisites are met.
3. **Submitter trying to approve their own document.** The Aprovar button is hidden / API rejects. ISO segregation is non-negotiable.
4. **Typing values for fixed tokens.** Don't type `DC-RH-001` for `{doc_code}` — just leave the chip. Freeze resolves it.
5. **Editing a template after documents are already in flight.** Existing documents keep their snapshot. New ones use the new published version.
6. **Waiting forever for the PDF.** Fanout is async. If after 2 min there's no PDF, check the outbox worker logs (technical — out of scope here, see `workflows/freeze-and-fanout.md`).
7. **Non-admin users cannot submit or approve without `user_process_areas` entries.** The approval authz system (`authz.Require`) resolves capabilities via `user_process_areas` joined to `role_capabilities`. If a user has no row in `user_process_areas` for the document's area, all capability checks fail silently (403). Required setup for smoke testing with dedicated accounts:
   - `author-test` needs role `author` in area `RH` (grants `doc.submit`, `doc.edit_draft`).
   - `approver-test` needs role `reviewer` in area `RH` (grants `doc.signoff`, `doc.submit`).

   Migration `0158` widens the `user_process_areas_role_check` constraint to accept these role values. If `0158` has not been applied, the INSERT below will fail with a check-constraint violation — run migrations first.

   ```sql
   -- author-test: author role in RH
   INSERT INTO metaldocs.user_process_areas (user_id, area_code, role, effective_from)
   SELECT u.id, 'RH', 'author', now()
   FROM metaldocs.iam_users u
   JOIN metaldocs.auth_identities ai ON ai.user_id = u.id
   WHERE ai.identifier = 'author-test'
   ON CONFLICT DO NOTHING;

   -- approver-test: reviewer role in RH
   INSERT INTO metaldocs.user_process_areas (user_id, area_code, role, effective_from)
   SELECT u.id, 'RH', 'reviewer', now()
   FROM metaldocs.iam_users u
   JOIN metaldocs.auth_identities ai ON ai.user_id = u.id
   WHERE ai.identifier = 'approver-test'
   ON CONFLICT DO NOTHING;
   ```

   See `references/local-dev-credentials.md` for the full capability matrix.

---

## Cross-refs

- Templates internals → `modules/templates-v2.md`
- Documents internals → `modules/documents-v2.md`
- Token catalog → `concepts/placeholders.md`
- Approval routing & signoffs → `workflows/approval.md` (TBD) and `modules/approval.md` (TBD)
- Freeze + fanout pipeline → `workflows/freeze-and-fanout.md`
- ISO segregation rationale → `concepts/iso-segregation.md`
- Code generation rules → `concepts/controlled-documents.md`
- Local dev startup → `references/local-dev-startup.md`
