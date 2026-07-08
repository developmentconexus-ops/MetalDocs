# M2c F8 — Live QA log (real stack, honestly labeled)

Real stack: API `metaldocs-api` :8081 (built `-Build`, migrations 33), FE vite dev :4173.
Auth: 4 real actors logged in via cookie session. Tenant `ffffffff-…-ffff`.

Actors / caps (from live `/auth/me`):
- author-test — document.create/edit/submit/view (AUTHOR); rh author membership.
- admin — document.review + document.signoff + document.publish + approval.oversee; rh qms_admin.
- approver — document.signoff (APPROVER, SoD ≠ author-test).
- approver-test — document.signoff; rh membership.

## Screens proven live (real backend, real session)

### C3/C5 — Worklist inbox (`/approvals`), admin
- Renders live: breadcrumb APROVAÇÕES › Caixa de entrada; Filtros (Estágio / Vencimento / Supervisão);
  Foco ↔ Linha do tempo view switch; "SUA FILA 0 · decisões pendentes"; empty-state copy PT-BR.
- Wine tokens render (F7). a11y snapshot clean.

### C1/C2 — Approval cockpit (`/approvals/{documentId}`), admin on PO-RH-003 (in_progress)
- DocumentShell (C1) renders live: header PO-RH-003, StateBadge "EM REVISÃO", v0 · REV00,
  editor toolbar (Documento/Comentários/File/Format/Insert/Help).
- Sidebar IA (C2) renders live: approval timeline "Submetido / Aprovacao / Outras ações";
  F7 LockBadge "🔒 Documento em revisão por outro usuário" (admin ≠ current stage actor → buffer locked);
  DecisionFooter "Assinar e aprovar" / "Assinar e devolver (requer justificativa)" / "Selecione uma decisão";
  "Cancelar instância" (CancelInstanceDialog trigger).
- Note: screenshot tool times out on the heavy docx editor iframe; proof captured via DOM introspection
  (document.body.innerText + querySelectorAll on sidebar classes) — honestly labeled non-screenshot.

## REAL BUG found by live QA (backend, HS-3 prerequisite repair)

Submitting a fresh doc (PO-RH-004, author-test) into a route whose active stage has **no SLA
due-in-days** fails:

```
POST /documents/{id}/submit → 500
approval handler error: submit: approval: unknown database error:
could not determine data type of parameter $16 (SQLSTATE 42P08)
```

Root cause: `InsertStageInstances` (postgres_approval_repository.go:94) computes `due_at` via
`CASE WHEN $13='active' AND $16 IS NOT NULL THEN now()+($16||' days')::interval ELSE NULL END`.
`$16` = `DueInDaysSnapshot *int`; when `nil` it is an **untyped NULL** referenced only inside the
CASE (never bound to a typed column) → Postgres cannot infer its type → 42P08. Any route stage
without an SLA can never be submitted. Sibling `UpdateStageStatus` (:1084) is safe — it uses the
typed COLUMN, not a bind param.

Disposition: real production defect, out of the M2c FE boundary (M2b F8 SLA backend code) but
HARD-BLOCKS the M2c live-QA walkthrough → HS-3. Fixed via type-cast on the param (`$16::int`),
TDD real-DB regression test added. Recorded as a deviation for HS-1.

## BUG #1 fixed — LIVE RED→GREEN confirmed
After the `$16::int` fix + `start-api.ps1 -Build` rebuild, resubmit of PO-RH-004:
`POST /documents/df236a53…/submit → 201` (was 500/42P08). Instance `98bd362d` in_progress,
stage 1 "Revisao" (stage_kind=review, active, actors admin+approver), stage 2 "Aprovacao" pending
(approver-test). **Live RED (500) → GREEN (201) proven on the real stack.**

Then admin `review-verdict {verdict:request_changes}` → `HTTP 200 outcome=changes_requested`;
document reverted to `draft` (review_verdict_service.go), instance → `changes_requested`.

## REAL BUG #2 found by live QA (backend contract-conformance, HS-3)
`GET /documents/{id}/approval-instance` returns `404 not_found.instance` for the
`changes_requested` instance:
```
repo LoadActiveInstanceByDocument WHERE ai.status IN ('in_progress','approved')   -- excludes changes_requested
```
But OpenAPI `ApprovalInstanceByDocumentResponse.status` enum INCLUDES `changes_requested`, and the FE
author panel (C5/F6, DocumentEditorPage.tsx:381-393) gates on
`getApprovalInstance().status === 'changes_requested'`. Net effect: **the C5 author
request-changes panel never renders on the real stack** — a core M2c deliverable, non-functional
live despite passing fixture-mocked unit tests (compile≠work).

Seam: the narrow repo method is shared by GET (FE read), publish, and mutation lookups. Broadening it
in place would corrupt publish/submit-re-entry semantics. Fix = a dedicated view-read method
(`LoadInstanceByDocumentForView`, status set + `changes_requested`) wired ONLY to
`GetInstanceByDocumentHandler`; publish + mutation stay narrow. TDD real-DB test; live RED already
captured (GET 404). Dispatched to a fresh sonnet implementer. Recorded as a deviation for HS-1.

## BUG #2 fixed — LIVE RED→GREEN confirmed
After `LoadInstanceByDocumentForView` (dedicated view-read incl. `changes_requested`) wired to
`GetInstanceByDocumentHandler` + rebuild: `GET /documents/{id}/approval-instance` for the
`changes_requested` instance → **200** (was 404). C5 author panel now has its data on the real stack.
Live RED (404) → GREEN (200) proven. Publish + mutation lookups kept on the narrow method (unchanged).

## Full lifecycle walkthrough — LIVE on real stack (PO-RH-004 df236a53 + others)

| Step | Call | Result |
|---|---|---|
| route review→approval | PUT `/approval/routes/{id}` v4, stage1 review role qms_admin, stage2 approval | route ready |
| submit | POST `/documents/df236a53/submit` If-Match "v?" | **201** instance in_progress (BUG#1 fix) |
| reviewer verdict request_changes | admin POST `review-verdict {request_changes}` | **200** doc→draft, instance→changes_requested |
| C5 author panel data | GET `/approval-instance` (changes_requested) | **200** (BUG#2 fix) — panel renders live |
| clean-buffer resubmit | author-test POST `/submit` | **201** new instance fcfe7a70 |
| review ready + freeze | admin POST `review-verdict {ready}` | **200** stage_completed, Aprovacao active, frozen_content_hash set |
| signoff w/ meaning | approver-test POST `/signoff {approve,password,content_hash}` If-Match "v3" | **200** outcome=approved; doc+instance approved |
| publish | admin POST `/documents/df236a53/publish` If-Match "v4" | **200** new_status=published |
| visibility 404 | admin GET bogus doc `/approval-instance` | **404** not_found.instance |
| oversee (has cap) | admin GET `/approval/inbox?scope=oversee` | **200** tenant-wide in-progress list |
| oversee denied | author-test GET `?scope=oversee` | **403** authz.capability_denied approval.oversee |
| cancel-with-reason | admin POST `/documents/564c8c75/cancel {reason}` If-Match "v1" | **200** doc under_review→draft, instance cancelled (dropped from oversee, view→404) |

All steps executed against `metaldocs-api` :8081 (real DB, real cookie sessions), honestly labeled —
curl-driven mutations with Origin + Idempotency-Key + If-Match; C1/C2/C3/C5 screen renders confirmed
via preview DOM introspection (screenshot tool times out on docx iframe). Two real backend defects
(BUG#1 SLA-null submit 42P08, BUG#2 changes_requested view 404) found by live QA, root-caused,
TDD-tested, live RED→GREEN — recorded as HS-1 deviations.
