# QA-1 Browser QA Verdict — Consolidated Milestone Tester

- **Date:** 2026-07-16
- **Tester:** QA-1 persona chip (fresh USER-level browser tester, zero implementation context)
- **Target:** `http://localhost:80` — full container stack (gateway → web/api/worker/jobs/docx-renderer), main @ `1070e94c`
- **Method:** Exercised through the real browser (in-app browser pane). API asserts used only to confirm error *shapes* after a UI observation, never as a substitute for UI testing.
- **Personas (dev seed, not secrets):** `admin`, `approver`, `author-test`, `approver-test`.
- **Separation of powers:** QA-1 JUDGES. Every RED finding was recorded and testing CONTINUED. No source was fixed. Shared stack is HUB-owned — no restart/rebuild/reseed performed; the two direct DB writes below are additive, `QA1-`-scoped test *setup* (documented), not remediation.

---

## Overall Verdict: **FAIL (ship-blocked)**

The core document lifecycle, the review/signature two-actions split (F2d.8), authz negative paths, taxonomy admin, and the inbox worklist are all **live-green**. But four **CRITICAL** defects block release:

- **F2** — route builder saves are impossible (capability id mismatch `doc.signoff` vs `document.signoff`).
- **F9** — the PDF/docx materialization pipeline is dead (consumer never asserts its capability GUC), so *no* document ever gets an official PDF, yet publish is still allowed (F13).
- **F18** — template approval routes cannot be created through the API at all (contract requires `profile_code`, the DB forbids it for templates).
- **F22** — template sign-off always 500s (`main.go` rebuilds the Decision service and drops the template ports).

F18 + F22 together mean the **entire template approval journey (J3) is non-functional end-to-end**. F9 + F13 mean the **document PDF chain is non-functional** even though the UI reports success. These are misleading-success and dead-pipeline defects, i.e. critical by the QA operating-system severity model.

---

## Per-Journey Results

| # | Journey | Verdict | Notes |
|---|---------|---------|-------|
| J1 | Login personas | **PASS** | All four personas authenticate after the hub reseed. Login = `POST /api/v1/auth/login {identifier,password}`. |
| J2 | Document lifecycle e2e (draft→submit→review→verdict→signoff→publish) | **PASS (UI)** / **FAIL (PDF chain)** | Full UI lifecycle incl. publish → PUBLICADO proven live. PDF never materializes (F9); publish gate ignores dead PDF (F13). |
| J3 | Template approval flow (kernel-routed) | **FAIL** | Route creation impossible (F18); template sign-off always 500 (F22). Journey cannot complete through the product. |
| J4 | Template inbox worklist (unit 4.2) | **PASS (core)** | Template row surfaces + deep-link works. Row shows raw template UUID instead of a kind badge/code (F19). |
| J5 | Route builder v2 | **FAIL** | Every save 400s — builder sends `doc.signoff`, registry expects `document.signoff` (F2). UI cannot author template routes at all (F17). |
| J6 | Approver execution panel (two actions) | **PASS** | Verified inline during J8 — verdict vs signature actions render per stage. |
| J7 | Taxonomy admin (TabBar / HeroHeader) | **PASS** | HeroHeader "TAXONOMIA / Tipos Documentais" + 3 working tabs (Famílias / Perfis / Áreas), correct tables per tab, console clean. Screenshot captured. |
| J8 | F2d.8 — review stage no signature panel + no 412 | **PASS** | Review stage renders verdict actions only; signature only at approval stage; **zero HTTP 412 across the whole journey**. |
| J9 | Negative paths (SoD, wrong-stage, cross-tenant, RFC 9457) | **PASS** | All four sub-paths behave correctly (details below). One UX honesty note (F23). |

---

## J9 — Negative Path Evidence (this session)

- **SoD (author cannot sign own document):** Built route `QA1-Route-PO v3` (review→approver-test, signature→author-test as named signers). `author-test` authored `PO-RH-003` (`d18fbfdf-…`). After the review stage passed, `author-test` is the *named* signer of the signature stage **and** the author. The approval-instance viewer returns `eligible_for_active_stage=false` with `is_author=true` — SoD strips eligibility despite the named-signer selector. Sign attempt → **403 `AUTH_FORBIDDEN`** (`application/problem+json`). UI shows **no signature action** for the author (only "Cancelar instância"). ✅ API + UI both enforce SoD.
- **Wrong-stage action:** `author-test` posting a `review-verdict` on the active review stage → **403 `AUTH_FORBIDDEN`**. Sign-off on an already-terminal instance → **409 `state.instance_completed`**. ✅
- **Cross-tenant isolation:** Foreign-tenant document `e12ea2ce-…` (tenant `ae99096e…` ≠ admin's `ffffffff…`) — dossier fetch → **404 `NOT_FOUND`** `application/problem+json`; the UI silently redirects to home rather than rendering a 404 page (minor UX note, not a leak). ✅
- **RFC 9457 shapes:** Clean `problem+json` with `title`/`status`/`code` observed for `404 NOT_FOUND`, `400 validation.header_required` (Idempotency-Key, If-Match), `400 validation.if_match_malformed`, `400 validation.request_invalid` (content_hash rules), `403 AUTH_FORBIDDEN`, `409 state.instance_completed`. ✅ **Counter-example:** the F14/F18 **500** responses leak internal error chains and DB constraint names into `problem+json.title` — that is the negative-shape defect feeding this journey.

**QA test-setup writes (additive, tenant-GUC + capability-asserted, `QA1`-scoped):** (1) direct INSERT of the template approval route for J4 (F18 blocks the API path); (2) `QA1-Route-PO v3` with `named_user` selectors to construct the SoD scenario. The DB capability tripwire (`enforce_capability_asserted`, P0001) fired on the first un-asserted attempt — last-line enforcement confirmed live.

---

## Findings Ledger (QA1-F1 … QA1-F23)

Severity per `wiki/quality/qa-operating-system.md`.

### CRITICAL

- **F2 — Route builder capability id mismatch.** Builder posts `doc.signoff`; the capability registry expects `document.signoff`. Every route save 400s. Handler tests mock the wrong value, so this is false-green in CI. *Evidence:* J5 network 400s; grep of route-admin feature.
- **F9 — docx materialization consumer never asserts capability GUC.** The materialize consumer hits the DB tripwire (`P0001 ErrCapabilityNotAsserted`), `attempt_count` exhausts at 5, PDF pipeline dead for every document. *Evidence:* worker outbox rows; stuck `pdf_status=pending`.
- **F13 — Publish allowed with PDF never materialized.** Publish gate does not require a materialized PDF and surfaces no failed-materialization state; `PO-RH-002` reached PUBLICADO with `pdf_status` stuck "pending" forever. Misleading success. *Evidence:* J2 publish → PUBLICADO; "Gerando o PDF oficial…" never resolves.
- **F18 — Template approval route creation impossible via API.** `CreateRouteRequest.Validate()` unconditionally requires `profile_code`, but the DB check `approval_routes_template_subject_projection_check` forbids `profile_code` for `subject_kind='template'`. `POST` with the field → 500 (DB projection check); without → 400 "profile_code is required". Contract ⊥ constraint. The handler test posts `profile_code:"ops"` through a **mock** service → false-green. *Evidence:* `internal/modules/approval/http/contracts/route.go` Validate vs DB constraint; `route_admin_handler_test.go:289`.
- **F22 — Template sign-off always 500.** `apps/api/cmd/metaldocs-api/main.go` sets `WithTemplateVersionReader` / `WithTemplateCompletionWriter` (~line 645–649), then **rebuilds** `approvalServices.Decision` (~line 737) via a fresh `NewDecisionService(...).With…()` chain that omits the template ports, and wires that crippled instance into the templates module (~line 765). Template sign-off hits `recordSignoff: template version reader not configured` → 500. The comment "Decision is finalized above" is false. *Evidence:* J3 sign-off 500; `decision_service.go:351-353`.

### HIGH

- **F3 — (operator) Edit vs approval workspaces disconnected + ugly.** Document edit and document approval screens are visually and structurally unrelated; Dossiê stacking and sidebar feel unplanned. Feeds the design-unification brief below.
- **F4 — (operator) Route admin exposes raw 1/N/M quorum concepts.** The route builder surfaces low-level quorum arithmetic instead of the intended human abstraction; person selection is hard.
- **F11 — One-click review verdict, no guard rail.** Review verdict is a single click with no confirm / comment-required / undo, and silently flips the panel into signature mode. Risk of accidental irreversible verdict.
- **F13** — (listed under CRITICAL).
- **F16 — Notifications 403 poll storm.** `author-test` / `approver-test` lack the notifications capability, but the frontend keeps polling → repeated 403s.
- **F17 — Route builder UI cannot author template routes.** No `subject_kind` control in route-admin (it exists only in inbox components); authors dead-end at a 409 `APPROVAL_ROUTE_MISSING`. Compounds F18.
- **F21 — (operator) Three divergent visual languages.** Document-edit, document-approval, and template screens are three different design idioms. Design-unification strategy required (brief below).

### MEDIUM

- **F1 — Fresh-draft dossier 404 rendered as an error.** A never-submitted draft has no approval instance; the 404 is treated as a hard error instead of an empty/expected state.
- **F5 — (operator) Rejection path invisible.** The review UI offers sign-and-return but no visible reject / request-changes affordance at the point of decision.
- **F8 — `/view` poll storm without backoff.** Document view polls with no backoff.
- **F14 — Blank-template submit 500 + internal leak.** Submitting a template with empty content 500s (`chk_template_version_content_hash_non_draft`), and the raw constraint name / error chain leaks into `problem+json.title`. (Becomes 409 `APPROVAL_ROUTE_MISSING` once content is typed.)
- **F19 — Inbox row shows raw template UUID.** Worklist renders the bare template UUID instead of a kind badge + human code.
- **F20 — Decision vocabulary / signature semantics inconsistent doc vs template.** "Solicitar mudanças" vs "Rejeitar"; the template flow lacks the MP 2.200-2/2001 legal-confirmation checkbox the document flow has.
- **F23 — Header sub-status stale after stage advance.** After the review stage passes and the signature stage goes active, the document header still reads "EM REVISÃO / **Aguardando revisão**" (survives reload). The badge is derived from `document.status = under_review` alone and ignores the active approval stage, so it tells the user the doc awaits *review* when it actually awaits *signature*. The dossier panel below shows the correct state (Revisão = Aprovado, Assinatura = Em andamento). *Evidence:* J9, doc `d18fbfdf-…`.

### LOW

- **F6 — Identificação "---" placeholders.** TIPO / ÁREA / VISIBILIDADE render "---" on the view page.
- **F7 — Pre-submit 404 poll storm.** Repeated dossier 404s before first submit.
- **F10 — "Sem assinaturas registradas." on review stage.** Copy implies a missing signature on a stage that has none by design.
- **F15 — Template editor header title overlaps buttons.**
- **F19 / F20** — (listed under MEDIUM).
- **Cross-tenant UI redirect** — a foreign-tenant URL silently redirects home instead of showing a 404 page (correct isolation, minor UX).

*(F12 was reclassified as a browser-tooling limitation, not a product defect.)*

---

## Design-Unification Brief (for claude.ai/design — requested by operator)

**Problem to state to the designer.** MetalDocs today ships three unrelated visual languages for what is one continuous job-to-be-done ("get a controlled document/template from draft to published"): (1) the **document editor**, (2) the **document approval / dossier** view, and (3) the **template edit + template approval** screens. They differ in layout grid, panel hierarchy, typography rhythm, action placement, and status vocabulary. A user crossing from authoring into approval feels like they changed apps.

**The anchor pattern (do not invent a new one).** MetalDocs already has the right idea in **ADR 0080 — the mode-adaptive workspace**: one workspace shell whose *content and available actions adapt to the current mode/stage*, rather than separate pages per activity. Unification should **generalize ADR 0080's mode-adaptive workspace to cover documents AND templates AND every lifecycle stage**, not add a fourth style.

**What to instruct the designer to produce:**

1. **One workspace shell, mode-adaptive by lifecycle stage.** A single frame — masthead (code · title · status) + primary canvas + a right-hand governed rail — that stays visually constant while the *canvas* (editor / read-only frozen view) and the *rail* (identification, revisions, dossier, next-approvers, actions) swap by stage: `draft → under_review(review) → under_review(signature) → approved → published`. The user should never feel a page-type change, only a stage change.
2. **One status system.** A single badge + sub-status component driven by *both* `document.status` and the *active approval stage* (fixes F23-class drift): e.g. "Em revisão · Aguardando revisão" then "Em revisão · Aguardando assinatura". Define the exact copy for every (status × active-stage) pair, for documents and templates identically.
3. **One action model, capability-shaped not role-shaped.** Actions in the rail render from the viewer's eligibility for the active stage (the API already returns `eligible_for_active_stage` / `is_author` / `has_signed_active_stage`). Design the empty/blocked states explicitly: author-viewing-own-doc-at-signature shows *why* they can't sign (SoD), not just a missing button. Design the review two-actions (verdict) vs signature (sign) split as the *same* control family, differing only by stage (this is the F2d.8 behaviour — keep it, standardize its look).
4. **Documents and templates share the identical shell.** The only legitimate differences are labels ("documento" vs "modelo") and that a template's subject key is a version. Everything else — grid, rail, dossier timeline, approver chips — must be one component set. This directly retires F3/F17/F19/F20/F21's divergence.
5. **Human approval vocabulary, not raw arithmetic (fixes F4).** The route/quorum concepts (1-of / all-of / M-of-N, named user vs role-in-area) must be presented as plain-language rules with person/role pickers, not exposed as raw `quorum`/`quorum_m` fields.
6. **Governed rail = the dossier is the spine.** The approval dossier (submitted-by, per-stage verdicts, signatures, reason-for-change) is the one persistent element that tells the truth across every stage; design it once as the authoritative timeline and reuse it in edit and approval modes (read-only while editing, active while approving).

**Deliverable to ask for:** a single component/layout spec (masthead, canvas, governed rail, status system, action states, approver/quorum presentation) plus the copy matrix for every status × stage, applied identically to documents and templates — explicitly framed as *generalizing ADR 0080*, with the empty/blocked/error states drawn (not just the happy path).

**Guard rails for the designer:** capability-driven, never role-driven, actions (ADR 0022); the review stage must never show a signature panel (F2d.8); status copy must never overstate success (no "published/approved" affordance while the PDF has not materialized — tie to F13). These are product invariants, not style choices.

---

## Environment & Evidence Notes

- Stack containers observed running: `metaldocs-gateway` (:80), `metaldocs-api` (:8081), `metaldocs-web`, `metaldocs-worker`, `metaldocs-jobs`, `metaldocs-postgres` (:5433), `metaldocs-redis`, `metaldocs-minio`, `metaldocs-gotenberg`, `compose-docx-renderer-1`.
- `metaldocs_app` DB role is superuser + BYPASSRLS in dev, so RLS is inert locally — cross-tenant isolation here is enforced by the application 404 path, not by RLS. Noted for reproduction fidelity.
- Browser-tooling caveats encountered (not product defects): password fields fought browser autofill (worked around with the native value setter + direct `/auth/login` fetch to establish session); `form_input` does not fire React `onChange` on some inputs; screenshot tool intermittently timed out (fell back to `read_page`/`get_page_text`); the document wizard's "Avançar/Criar" sometimes needed a re-click.

## Recommended Release Gate

- **Block release** until F2, F9, F13, F18, F22 are root-caused and re-verified live through the browser.
- F18 + F22 should be fixed as **one family** (the approval-kernel template wiring): the contract/DB projection contradiction and the `main.go` Decision-rebuild drop are the same "template ports not honored" root boundary.
- F2 + F17 are the **route-builder capability/subject family** — fix the capability id at source and add the `subject_kind` control; delete the mock-only handler tests that masked both.
- Re-run J2 (PDF chain) and J3 (template e2e) after those fixes; everything else can proceed to a design pass using the brief above.
