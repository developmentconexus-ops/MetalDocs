# QA-3 — Browser UI QA pós-remediação (2026-07-22)

Operator-authorized hub browser run (stack :80 @ 01e1a785 + docs commits).
Auth method: dev-seed login via curl (shell), session cookie injected into
browser — **no password ever typed into a browser form** (standing prohibition;
same boundary that deferred unit 2.5 L3). The signoff password step was
therefore executed via API shell with identical payload to the UI form; all
other steps driven in the real UI.

## Journey (document PO-RH-004, id 45c9e784, profile po, blank template)

| Step | Surface | Result |
|---|---|---|
| Login author-test | cookie injection | OK (UI shows Author Test) |
| Novo documento wizard (perfil po → área rh → título → Em branco → confirmar) | UI | PASS — PO-RH-004 created, editor opens |
| Type body content, autosave | UI (eigenpal editor) | PASS — "Salvo", revision 7 uploaded (1497 B, page_count 1) |
| Submeter para revisão | UI | PASS — toast "Documento submetido para revisão", EM REVISÃO, timeline stage "Revisão e Assinatura em andamento" |
| Approver inbox | UI (cookie swap) | PASS — 1 decisão pendente, card correct (autor, área, estágio 0/1) |
| "Aprovar e assinar →" | UI | PASS — workspace "Aprovando" mode, "Conteúdo verificado ✓", decision panel (decision, justificativa, password, legal checkbox) |
| Signoff (password step) | API shell (boundary) | 200 outcome approved |
| Document details | UI | PASS — REV00 · APROVADO, "aprovado em 22 de julho de 2026", Visualizar/Baixar PDF/Publicar buttons |
| Visualizar documento | UI | iframe loads presigned MinIO `final.pdf` (200, 6118 B) — pipeline surfaced to user |

Pipeline (DB/MinIO): materialize + pdf events published, staging row with
`final_docx_s3_key`, `final.pdf` + `frozen.docx` objects present. Plumbing
(QR-B/QR-C scope) fully green.

## F-QA3-1 — SHIP-BLOCKER: approved/signed PDF contains NO authored content

`final.pdf` (6118 B) is a genuinely blank 1-page PDF: page content stream has
no text operators, font dict is `<<>>` empty. `frozen.docx` is the empty
skeleton (`<w:p/>` only, 1109 B). Same for QA-2's document d18fbfdf (frozen.docx
identical skeleton) — **QA-2 proved plumbing, not content**.

Root cause chain (evidence-verified):

1. Editor content DOES persist: `document_revisions` rev 7 = real uploaded docx
   (1497 B, hash `07e00336…`), `documents.current_revision_id` → rev 7.
2. `FreezeService.Materialize` (`internal/modules/documents/application/freeze_service.go:277-284`)
   sends `BodyDocxS3Key: snap.BodyDocxS3Key` = **template snapshot**
   (`documents.body_docx_snapshot_s3_key` = `system/templates/blank.docx`,
   `body_docx_hash` = `e3b0c442…` = SHA-256 of empty input).
3. Renderer fanout renders template + pinned placeholder values (blank template,
   zero placeholders → blank output), writes `frozen.docx`, → Gotenberg → blank
   `final.pdf`.
4. The editor-authored revision docx is **never consumed** by the freeze
   pipeline. Any free-typed editor content is silently absent from the official
   frozen/signed/published artifact.

Integrity corollary: the approver reviewed the editor content in the workspace
("Conteúdo verificado ✓" over the live document) but the instance's
`frozen_content_hash` (`d408946e…`) matches neither revision hash — it hashes
the blank render. Signature attests to an artifact the approver never saw.

Classification: **architecture contradiction between two content models** —
docgen v2 (template + placeholder pinning = frozen truth) vs eigenpal editor
free-edit (revision docx = edited truth). Per CLAUDE.md this is a stop-and-
surface, not a patch. Options for operator ruling:

- **(a) Fix within model:** Materialize consumes `current_revision_id`'s
  revision docx as body (editor truth wins), placeholder resolution applied to
  it. Template snapshot only seeds the initial clone.
- **(b) Refactor/model split:** two governed document classes — template-
  governed (placeholder-only; editor must NOT allow free body edits) vs
  free-form (freeze = revision docx verbatim). Requires ADR + editor lockdown.

## Verdict

QA-3 **FAIL** (ship-blocker F-QA3-1). UI journey itself (wizard, editor,
submit, inbox, approve flow, status surfacing, PDF wiring) all PASS — the
defect is content-model, upstream of QR-C's (correct and live-proven) event
contract. M5 HS-1 remains blocked per operator condition ("funcionando
totalmente como um usuário utilizaria").

Minor observation: workspace right panel shows "Não foi possível carregar os
dados de aprovação" + Tentar novamente on a never-submitted draft (expected
404 surfaced as retryable error) — cosmetic, register with F-QA2 defers.
