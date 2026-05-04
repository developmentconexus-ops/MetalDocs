# System Acceptance Test — MetalDocs QMS

> **Last verified:** 2026-05-04
> **Scope:** Full end-to-end manual acceptance run for the MetalDocs regulatory-grade QMS. Covers taxonomy bootstrap → template authoring → controlled-document creation → ISO-segregated approval → freeze → PDF fanout → archive. Exercises Groups A–E shipped fixes.
> **Out of scope:** IAM admin flows (B1/A8), route-stage editing (A7/E8), multi-tenant isolation (B5/B6), Playwright e2e automation (see `references/how-to-run-tests.md`).
> **Audience:** QA engineers, release approvers, on-call engineers validating a hot-fix deploy.
> **Key files:**
> - `frontend/apps/web/src/features/approval/components/RegistryDetailPanel.tsx:314` — "Submeter para revisão" button label
> - `frontend/apps/web/src/features/documents/v2/hooks/useDocumentPdfStatus.ts:12-13` — POLL_INTERVAL_MS=3000, TIMEOUT_MS=60000 (E11)
> - `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx:172` — isEditable gate: `docStatus === 'draft'` (E1)
> - `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx:203` — "Aguardando..." submit button when auth not ready (E12)
> - `frontend/apps/web/src/features/registry/RegistryDetailPage.tsx:174` — "Nova Revisão" button (E10)
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx:98` — area filter first option "Todas as áreas", options from taxonomy (E7)
> - `internal/modules/documents/repository/repository.go:206,233,892-910` — `archived_at IS NULL` filter (C1)

**Predecessor doc:** `wiki/workflows/user-onboarding.md` — read for conceptual context on each step before running this checklist.

**Expected run time:** 20–35 min (full run, clean environment).

---

## What's new since 2026-05-02 (Groups C/D/E)

For testers who have run prior smoke runs — key behavior deltas:

- **PDF link appears without F5** — auto-polling every 3s (E11).
- **Token `{author}` / `{approvers}` resolve to display names, not UUIDs** (D1).
- **Token `{controlled_by_area}` resolves to area NAME, not code** — e.g. "Recursos Humanos" not "RH" (D2).
- **Archive removes document from default list** — soft-delete via `archived_at` (C1).
- **Editor is read-only on `under_review`, `approved`, `published`** — status-gated, not session-phase-gated (E1).
- **Submit button label is "Submeter para revisão"** — "Enviar" verb removed.

---

## Prerequisites

- API: `.\scripts\start-api.ps1` (port 8081). Frontend running or built.
- Postgres + MinIO + Gotenberg up via Docker Compose.
- Admin login: `admin` / `AdminMetalDocs123!`
- Clean environment **or** existing tenant with at least one family, area, profile, published template.
- Bootstrap triggers when `metaldocs.iam_user_roles` has no row with `role_code = 'system_admin'`.

---

## Setup — Test users (one-time per environment)

Three distinct accounts are required to cover ISO segregation:

- `admin` — taxonomy, template publish, profile binding.
- `author-test` — template and document author.
- `approver-test` — document approver (must not be the submitter).

Create via **Configurações → Usuários** (admin UI). Then grant process-area roles via SQL (required for `authz.Require` checks — without these rows all approval capability checks return 403):

```sql
-- author-test: author role in area RH
INSERT INTO metaldocs.user_process_areas (user_id, area_code, role, effective_from)
SELECT u.id, 'RH', 'author', now()
FROM metaldocs.iam_users u
JOIN metaldocs.auth_identities ai ON ai.user_id = u.id
WHERE ai.identifier = 'author-test'
ON CONFLICT DO NOTHING;

-- approver-test: reviewer role in area RH
INSERT INTO metaldocs.user_process_areas (user_id, area_code, role, effective_from)
SELECT u.id, 'RH', 'reviewer', now()
FROM metaldocs.iam_users u
JOIN metaldocs.auth_identities ai ON ai.user_id = u.id
WHERE ai.identifier = 'approver-test'
ON CONFLICT DO NOTHING;
```

Migration `0158` must be applied before this INSERT (widens the role check constraint). Full capability matrix: `references/local-dev-credentials.md`.

See `wiki/workflows/user-onboarding.md:229-253` for the canonical SQL and pitfalls.

---

## Routine A — Bootstrap (admin)

> Maps to Steps 1 + 4 of `user-onboarding.md`.

### A0. Famílias (precondition — run before A2)

- [ ] Login as `admin`.
- [ ] **Tipos Documentais → Famílias → Nova Família**: code `qualidade`, name `Qualidade`.
- [ ] Confirm family listed.
- [ ] **Expected:** no error toast. Families are globally scoped (shared across tenants). ProfileEditDialog requires a valid `familyCode` — A2 will fail with a validation error if no family exists.

### A1. Áreas

- [ ] **Tipos Documentais → Áreas → Nova Área**: code `RH`, name `Recursos Humanos`.
- [ ] Confirm area appears in list.
- [ ] **Expected:** no error toast. Persists after F5.

### A2. Perfis

- [ ] **Tipos Documentais → Perfis → Novo Perfil**: code `DC`, name `Descrição de Cargo`. Family: `qualidade`.
- [ ] Confirm profile listed.

### A3. (Deferred — return after Routine B)

> Bind profile to published template. See A4 below.

---

## Routine B — Template authoring (author-test + admin)

> Maps to Steps 2 + 3.

### B1. Create template

- [ ] Logout admin → login as `author-test`.
- [ ] **Templates → Novo Template**: title `DC — Descrição de Cargo`, target profile `DC`.
- [ ] Confirm creation. eigenpal author opens. State `v1 draft`.

### B2. Edit content

- [ ] Insert header with `{doc_code}` chip.
- [ ] Insert title block with `{doc_title}`.
- [ ] Insert signature line with `{author}` + `{effective_date}`.
- [ ] Insert block "Aprovado por: `{approvers}`".
- [ ] Insert area line "`{controlled_by_area}`".
- [ ] Body: 2–3 lorem ipsum paragraphs.
- [ ] **Save**. F5 → content persists.
- [ ] **Expected:** all 7 tokens recognized. No red/error chips.

### B3. Submit → Approve → Publish

- [ ] Version panel: **Submeter para revisão** → state `in_review`.
- [ ] Logout `author-test` → login `admin` (or template approver).
- [ ] **Templates → DC → v1 → Aprovar**. State `approved`.
- [ ] **Publicar**. State `published`.
- [ ] **Expected:** Publicar button enabled only after approve. Template list shows `v1 published`.

### B4. Tentar editar versão publicada (negative test)

- [ ] Attempt to edit content of `v1 published`.
- [ ] **Expected:** editor read-only or Save button disabled. Cannot mutate published version.

---

## Routine A4 — Bind template to profile (CRITICAL)

> Maps to Step 4. **Without this, Routine C fails.**

- [ ] Logged as `admin`.
- [ ] **Tipos Documentais → Perfis → DC → Editar**.
- [ ] **Template padrão**: select `DC — Descrição de Cargo (v1)`.
- [ ] Save.
- [ ] Reopen profile → confirm binding persisted.

---

## Routine C — Document creation (author-test)

> Maps to Steps 5 + 6 + 7.

### C1. Register Documento Controlado

- [ ] Login `author-test`.
- [ ] **Documentos Controlados → Novo Documento Controlado**.
- [ ] Profile: `DC`. Area: `RH`. Title: `Descrição de Cargo — Analista Fiscal`.
- [ ] Confirm.
- [ ] **Expected:** new CD with code `DC-RH-001` (or next in sequence).
- [ ] **E12 regression:** open the dialog immediately after page load (before auth fully resolves) → submit button should be disabled with label **"Aguardando..."**. (`RegistryCreateDialog.tsx:203`)

### C2. Generate working version

- [ ] Open CD → **Gerar Documento**.
- [ ] Wizard shows template already selected (binding from A4).
- [ ] **Gerar**. Editor opens.
- [ ] **Expected:** layout matches template. Tokens visible as chips, **not** resolved yet.

### C3. Fill content

- [ ] Add 2–3 paragraphs describing the role.
- [ ] Insert 2×3 table with responsibilities (validates table rendering).
- [ ] **Do not touch** the 7 fixed token chips.
- [ ] **Save**. F5 → persists.

### C4. Import DOCX (optional, gate sanity)

- [ ] In a separate test CD, generate document and import an external `.docx`.
- [ ] **Expected:** import without critical formatting loss. (eigenpal table-in-header bug is parked — note if it reappears.)

### C5. Renomear documento (E9 regression)

- [ ] In the editor, change the title → confirm persistence (F5 keeps new name).
- [ ] **Negative path:** stop `metaldocs-api.exe` → attempt rename again.
- [ ] **Expected:** UI reverts to last saved title (optimistic rollback). Structured Portuguese error toast via `resolveErrorMessage`.
- [ ] Restart API before continuing.

---

## Routine D — Approval (author-test + approver-test)

> Maps to Step 8. Validates ISO segregation.

### D0. Sem rota de aprovação (E3 negative path)

- [ ] As `admin`, ensure profile `DC` has **no active** approval route (deactivate all in Approval Routes).
- [ ] As `author-test`, create a new CD + generate doc + click **Finalizar**.
- [ ] **Expected:** specific Portuguese toast "Não há rota de aprovação ativa para este perfil." — not "Failed to finalize".
- [ ] Re-activate the route before continuing with D1.

### D1. Submit

- [ ] In editor → **Finalizar**.
- [ ] **Expected:** state `under_review`. Success toast.

### D2. ISO segregation (negative test)

- [ ] Still as `author-test`, go to **Caixa de Entrada de Aprovação**.
- [ ] **Expected:** submitted document does **not appear** for approval by the submitter. If it appears, critical bug.

### D3. Approve

- [ ] Logout → login `approver-test`.
- [ ] **Caixa de Entrada de Aprovação** → document `DC-RH-001` listed.
- [ ] Open → review content → **Aprovar**.
- [ ] Modal asks for password → enter `approver-test` password → confirm.
- [ ] **Expected:** signoff recorded. If route is `any_1`, state goes to `approved` immediately. If `m_of_n`, repeat with another approver.

### D4. Área filter (E7 regression)

- [ ] Still in **Caixa de Entrada de Aprovação**: open the area dropdown filter.
- [ ] **Expected:** first option is **"Todas as áreas"**. Remaining options come from taxonomy areas (e.g. `RH — Recursos Humanos`) — not a hardcoded `JUR/RH/FIN/TI/COM/ENG` set. (`InboxPage.tsx:98-100`)

### D5. Rejeição (alternative path — separate CD)

- [ ] Repeat D1–D3 on a separate CD, but click **Rejeitar** with a reason.
- [ ] **Expected:** state goes to `rejected`. Author receives notification.

---

## Routine E — Freeze + Fanout (automatic, observable)

> Maps to Step 9.

### E1. Freeze automático — status-gated read-only (regression E1)

- [ ] Immediately after last signoff in D3, open document detail.
- [ ] **Expected:** state `approved` (`frozen` terminology removed — `approved` is the post-signoff immutable state). Hashes `content_hash` visible, `revision_version` incremented.
- [ ] Attempt to edit → **Expected:** editor read-only. Edit button and Save disabled. Gate: `docStatus === 'draft'` — independent of session phase. (`DocumentEditorPage.tsx:172`)

### E2. Token resolution (D1, D2, D3, D8 regressions)

- [ ] Download the frozen DOCX → open in Word/LibreOffice.
- [ ] **Expected for each token:**
  - `{doc_code}` → real code (`DC-RH-001`).
  - `{doc_title}` → CD title.
  - `{revision_number}` → `1`.
  - `{author}` → **display name** of author (not UUID). (D1)
  - `{effective_date}` → fixed date (same across retries/replays; D3 ensures determinism).
  - `{approvers}` → list of **names** of approvers who signed **this** instance (D1, D8). No approvers from previous revisions.
  - `{controlled_by_area}` → **name** of area (`Recursos Humanos`), not the code (`RH`). (D2)

### E3. Fanout (PDF) — auto-polling (E11 regression)

- [ ] Stay on the document page after final approval.
- [ ] **Expected:** within ~30s the PDF link appears **without manual refresh** — hook `useDocumentPdfStatus` polls every 3s, timeout 60s. (`useDocumentPdfStatus.ts:12-13`)
- [ ] Visible states: `pending` → `ready`. If `failed` after >60s, check worker logs (`metaldocs-worker.exe`).
- [ ] Open PDF → tokens resolved, layout faithful to DOCX.

### E4. Idempotency

- [ ] Attempt re-approve / re-freeze via API or re-click → **Expected:** silent no-op, state not corrupted.
- [ ] **Known note (verify post-D4/D5 outbox refactor):** replay of the same `Idempotency-Key` after instance already approved may return HTTP 500 instead of `was_replay: true`. State is not corrupted, but the error response is incorrect. Confirm behavior has not regressed.

---

## Routine F — Revision (second iteration)

> Smoke the revision flow. Reuse the CD from Routine C.

- [ ] On CD `DC-RH-001` (already approved) → **Nova Revisão**.
- [ ] **E10 regression:** "Nova Revisão" button must appear even when a published version exists and no active draft is in progress. (`RegistryDetailPage.tsx:174`)
- [ ] **Expected:** new version `v2 draft` cloned from v1.
- [ ] Edit a difference (1 paragraph).
- [ ] Submit → Approve → Freeze → Fanout (Routines D + E).
- [ ] **Expected:** `{revision_number}` now resolves to `2`. Previous version remains accessible in history.
- [ ] **D5 regression:** v2's PDF is a **new, distinct artifact** from v1's PDF — not the same file.
- [ ] **D8 regression:** approvers listed in the v2 DOCX/PDF are only v2's approvers — no cross-contamination from v1.

---

## Routine G — Arquivamento (C1 regression)

- [ ] On CD `DC-RH-001` (any state) → **Arquivar** (or via API `POST /api/v1/documents/{id}/archive`).
- [ ] Return to **Documentos Controlados** list.
- [ ] **Expected:** the CD does **not appear** in the default list. (`archived_at IS NULL` filter — `repository.go:206,233,892-910`, ADR 0008.)
- [ ] (If UI supports it) toggle "incluir arquivados" filter → CD reappears with `arquivado` badge.
- [ ] Confirm via DB:
  ```sql
  SELECT id, archived_at FROM metaldocs.documents WHERE id = '<id>';
  ```
  → `archived_at` populated (not NULL). Document is soft-deleted, not hard-deleted.

---

## Pass / Fail Criteria

**Pass:** All Routines A–G complete without errors. Negative-test checkboxes (B4, D2, D0) confirm blocking. PDF generated with tokens resolved to human-readable values.

**Fail (any one of these blocks release):**

- Wizard in C2 offers no template (A4 binding or publish bug).
- Author can approve own document (D2 regression).
- Token does not resolve in frozen DOCX/PDF (E2/E3).
- `{author}` or `{approvers}` shows UUID instead of display name (D1 regression).
- `{controlled_by_area}` shows code instead of name (D2 regression).
- PDF does not appear after 60s of auto-polling (E11 regression).
- PDF link requires manual F5 to appear (E11 regression — hook not polling).
- Editor allows editing when `docStatus` is `under_review`, `approved`, or `published` (E1 regression).
- CD appears in default list after archive (C1 regression).
- Fanout does not produce PDF within 5 min.

---

## Quick commands

```powershell
# Start API
.\scripts\start-api.ps1

# Rebuild + start
.\scripts\start-api.ps1 -Build

# Frontend
cd frontend\apps\web; npm.cmd run build
cd frontend\apps\web; npm.cmd run dev

# Compose stack
docker compose up -d
docker compose logs -f api
docker compose logs -f docgen-v2
```

```bash
# Login check
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"admin","password":"AdminMetalDocs123!"}'
```

---

## Cross-refs

- Conceptual walkthrough → `wiki/workflows/user-onboarding.md`
- Token catalog (7 fixed tokens) → `wiki/concepts/placeholders.md`
- Freeze + fanout pipeline → `wiki/workflows/freeze-and-fanout.md`
- Approval routing details → `wiki/workflows/approval.md`
- ISO segregation rationale → `wiki/concepts/iso-segregation.md`
- Error UX (resolveErrorMessage, toasts) → `wiki/concepts/error-ux.md`
- Archive ADR → `wiki/decisions/0008-placeholder-fixed-catalog.md`
- How to run automated tests → `wiki/references/how-to-run-tests.md`
- Local dev credentials + capability matrix → `wiki/references/local-dev-credentials.md`
