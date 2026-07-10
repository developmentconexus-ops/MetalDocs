# F2d.8 — Rendered-UI Live QA (browser, gateway :80)

> **Date:** 2026-07-10
> **Operator:** QA session (fresh, zero inherited context)
> **Scope:** F2d.8 close — UI-driven live QA of the single-screen approval workflow.
> **Stack:** full container stack via gateway `http://localhost` (:80). All observations are
> from the actually-rendered SPA on :80, driven through the app's own React handlers — **not** curl.
> **Verdict:** **NOT ALL GREEN — Journey 2 is RED.** Milestone-validator NOT run (all-green gate
> unmet). No fixes attempted (separation of powers). Artifacts committed, not pushed.

---

## Preflight

- **Browser tooling present** — `mcp Claude_Browser` preview tools available and functional
  (unblocks prior session's F-UI-1 "no browser tooling" defer).
- **Stack healthy** — `metaldocs-{api,worker,jobs,web,postgres,redis,minio,gotenberg}` +
  `docx-renderer` all up; gateway `nginx` on :80.
- **Served-FE provenance pinned** — `metaldocs-web:dev` is an nginx static build. Image `Created`
  `2026-07-10T20:11:25Z` (17:11 local). Main-checkout reflog shows HEAD was **`6ae3d6e5`** at that
  time (next commit `610d4a70` at 17:27). `6ae3d6e5` = this worktree's exact commit; the commits
  after it to current `main` (`de277a11`) are docs-only. **Therefore the served FE = the source
  read in this session, byte-for-byte.** Every code-truth claim below matches the running build.

### Evidence-method note (read before judging evidence class)

The `computer{action:"screenshot"}` tool **consistently timed out** in this environment. Per the
prior session's determination, rendered evidence is captured as **live-DOM reads** (via
`javascript_tool` against the :80 SPA) + **network** + **server logs**. This is the DOM/a11y-snapshot
class the acceptance calls for ("per-step UI evidence (DOM/a11y snapshots)"); it is rendered-UI
proof, not curl. Screenshot capture is an **evidence-tooling gap**, not a coverage gap.

### Fixture note (why the draft was advanced via API)

Journey 2's submit affordance is **missing from the rendered UI** (Finding **RED-1**). To let the
graded review/approval **screens** be observed as a real user would see them, the draft was advanced
to `under_review` via `POST /documents/{id}/submit` **purely as fixture plumbing**. Every *graded
observation* (J1, signature-ceremony positive control, J3 SoD) is a rendered-UI read. The
review→approval verdict was driven by a **real UI click**, not the API.

### Test artifact

- Document `PO-RH-001` (`bb4c58f4-b098-4ec5-b764-fca965a154fe`), profile `po`, created via the UI
  New-Document wizard, `created_by = author-test`.
- **Route A (review + approval)** instance `5164ae58-5428-423d-83d7-a056dfef6765`:
  stage 1 `Revisao` (`stage_kind=review`), stage 2 `Aprovacao` (`stage_kind=approval`).
- Personas (dev seed fixtures, `db/dev-seeds/0001_local_dev_seed.sql`): `author-test` (author /
  submitter), `admin` (review-stage actor), `approver-test` (approval-stage actor).

---

## Journey 1 — Review-stage verdict screen shows NO signature panel → **GREEN**

Persona **admin** (review-stage actor; `viewer = {is_author:false, eligible_for_active_stage:true,
has_signed_active_stage:false}`), rendered workspace at `/documents/bb4c58f4…`:

| Observed (live DOM) | Value |
|---|---|
| Verdict CTAs | **"Pronto para aprovação"**, **"Solicitar mudanças"** |
| `input[type=password]` present | **false** |
| Legal-effect checkbox count | **0** |
| Signing affordance (`Assinar…`) | **none** |
| Signature timeline label | "Sem assinaturas registradas." (read-only, not a control) |

→ Review-kind active stage offers a **verdict**, never a signature panel. **The M2c F4 deviation
(signature panel offered on a review stage) is CLOSED.** Meets acceptance line 20 / line 115(b).

---

## Journey 2 — Full lifecycle draft→submit→verdict→signoff→publish through UI, zero 412 → **RED**

### RED-1 (critical, FE) — no submit-for-approval affordance in the rendered UI

The author cannot start the lifecycle from the UI.

- Persona **author-test** (the document's creator; `/auth/me` confirms capability
  **`document.submit`**), rendered draft workspace (`author-editing` mode):
  buttons = editor toolbar (File/Format/Insert/Help/zoom/font) + "Tentar novamente" + "Ver ficha
  completa do documento". **`submitish = []`** — no "Submeter"/"Enviar"/"Revisão"/"Aprovação"
  control anywhere in the DOM.
- **Code truth (matches served build).** App-wide, the **only** path that POSTs
  `/documents/{id}/submit` is `approvalApi.submit` (`approval/api/approvalApi.ts:127`). Its **sole
  live importer is `DocumentEditorPage.tsx`** (line 17; the "Submeter para revisão" button lives at
  line 508). `DocumentEditorPage` is **UNROUTED**: F2d.7 (`a90275e7`) deleted its route wrapper
  `DocumentEditorRoutePage`; `routes.tsx` maps `/documents/:id/edit` → `<Navigate>` to the
  workspace; **no route mounts `DocumentEditorPage`** (every remaining reference is a comment,
  mirror-note, or test). The workspace extracted `EditorCanvas` for editing but **never re-homed
  the submit control**. `NewDocumentWizardPage` leaves the new doc as `draft` and navigates to the
  workspace — it does not submit.
- **Repro:** log in as `author-test`; open `/documents/{a-draft-id}`; observe author-editing mode
  with a working editor but **no submit action**. There is no UI route to advance
  `draft → under_review`.
- **Consequence:** the lifecycle **cannot be initiated in the rendered UI**. This holds for **both**
  route shapes — submit is a single route-agnostic code path — so a separate approval-only route
  walkthrough would hit the identical blocker. Acceptance line 57 ("full lifecycle
  draft→submit→…") is **FAILED**.

### Downstream (post-submit, once state set via fixture plumbing) — partial GREEN

- **Review→approval advance, driven by a real UI click** ("Pronto para aprovação" as admin):
  stage 1 `review` → **passed**, stage 2 `approval` → **active**, 1 verdict recorded,
  **zero 412 / zero errors**.
- Approval-stage signature ceremony correctly presented (positive control below).

These prove the *mid-stream* flow is UI-coherent; they do **not** flip J2, which is graded on the
start-to-finish UI lifecycle and is broken at step 1.

### Signoff execution — deliberately NOT performed

Completing the cryptographic signoff requires **entering an approver password into a field**, which
is a safety boundary this operator does not cross. The FE's responsibility — **correctly presenting
the gated ceremony** — is proven (below). The actual signoff + publish can be completed by the
operator; it is not required to reach the J2 verdict (already RED at submit).

---

## Positive control — signature ceremony IS present on the approval stage → **GREEN**

Persona **approver-test** (approval-stage actor; `viewer.eligible_for_active_stage = true`),
rendered workspace (`approving` mode):

| Observed (live DOM) | Value |
|---|---|
| `input[type=password]` present | **true** |
| Legal-effect checkbox count | **1** |
| Legal-effect copy | "Confirmo que revisei o conteúdo integralmente e que esta decisão tem efeito de assinatura eletrônica conforme a MP 2.200-2/2001." |
| Sign options | **"Assinar e aprovar"**, **"Assinar e devolver"** (return requires justification) |
| Submit gating | "Selecione uma decisão" (disabled until a decision is picked) |

→ Approval-kind stage renders the signature panel **only for an eligible actor**, with password
re-auth + legal-effect attestation. Correct contrast to J1. Meets acceptance line 24 / line 115.

---

## Journey 3 — Negatives → **GREEN** (with one out-of-scope backend finding)

### SoD — the submitter cannot sign → GREEN (rendered)

Persona **author-test** (the submitter; `viewer = {is_author:true, eligible_for_active_stage:false}`)
on the **active approval stage**, rendered workspace:

- `input[type=password]` = **false**, legal checkbox = **0**, signing affordance = **none**.
- Only "Cancelar instância" (the author may cancel their own instance).

Contrast with the positive control (an *independent* approver got the full ceremony) →
**Separation of Duties is enforced in the rendered UI**: the person who submitted cannot approve;
a different approver can.

### Wrong-stage action is rejected → GREEN (integrity), with Finding B-1

- The rendered UI **never offers** the wrong action (J1 vs positive control: review stage offers
  only verdicts; approval stage offers only signoffs), so a normal user cannot reach it.
- Crafted API probe (bypassing the UI): `POST /approval/instances/{id}/stages/{approvalStageId}/
  review-verdict {verdict:"ready"}` as the eligible approver → **rejected**; domain guard fired
  (`"review_verdict: stage must be kind 'review'"`); **no verdict recorded** (verdicts still 1,
  approval stage still active, 0 signoffs). Integrity holds.

#### Finding B-1 (backend; OUT OF SCOPE for this FE milestone; medium)

The wrong-stage guard maps to **HTTP 500 `internal.unknown`** ("internal error") instead of a
client **4xx**. The action is correctly rejected and integrity is preserved, but the error
**classification** is wrong: a caller error surfaces as a server error, polluting 5xx/error budgets
and leaking a generic "internal error" to the client. **Not reachable via the rendered UI** (the UI
never offers the action), so it does not affect the FE milestone verdict.
Repro: server log `{"level":"ERROR","msg":"approval handler error","status":500,
"code":"internal.unknown","error":"review_verdict: stage must be kind 'review'"}`.
Owning boundary: `documents/approval` application-service error mapping.

---

## Secondary FE finding

### F-2 (FE; medium UX) — draft shows an approval-load error instead of a clean empty state

A never-submitted draft has no active instance; `GET /documents/{id}/approval-instance` correctly
returns **404 `not_found.instance`** (proper `application/problem+json`, `status:404`). But the
rendered draft workspace sidebar "Dossiê governado" section shows an **error state**:
"Não foi possível carregar os dados de aprovação." + "Tentar novamente" — rather than a clean
draft/empty state.

- Source **appears** to guard this: `useApprovalInstanceQuery` catches `status === 404` and returns
  `null` (no error) — yet the rendered UI (built from the same commit `6ae3d6e5`) still errors
  (`DocumentWorkspacePage.tsx:173` ties `instanceError` to `instanceQuery.isError`). The runtime
  behavior contradicts the apparent source guard → **needs developer root-cause** (thrown
  `ApiError` shape vs the `.status === 404` guard at runtime, or query wiring). RED = repro only;
  no fix attempted.
- Repro: log in as `author-test`; open a fresh draft's `/documents/{id}`; the sidebar shows the
  approval-data load error + retry.

---

## Route shape B (approval-only) — not separately exercised (justified)

Only Route A (review + approval) was instantiated. Route B (approval-only) was **not** separately
walked because: (a) the graded J2 blocker **RED-1 is route-shape-independent** — submit is one
route-agnostic code path, so Route B fails identically at step 1; (b) the graded screen behaviors
are **stage-kind driven** (review→verdict-only; approval→ceremony-when-eligible), both already
observed on Route A's two stage kinds. A dedicated approval-only walkthrough would add no new
signal to the current verdict. Flagged as a bounded, explicit gap.

---

## Disposition

| Item | Verdict |
|---|---|
| J1 — review stage, no signature panel | **GREEN** |
| J2 — full lifecycle through UI, zero 412 | **RED** (RED-1: no submit affordance in UI) |
| Positive control — approval-stage signature ceremony | **GREEN** |
| J3 — SoD (submitter cannot sign) | **GREEN** |
| J3 — wrong-stage rejected (integrity) | **GREEN** (Finding B-1: 500 vs 4xx, backend, out of scope) |
| M2c F4 deviation closure (review ≠ signature) | **CLOSED / GREEN** |

**Findings ledger**
- **RED-1** (critical, FE, blocks J2): draft→submit affordance absent from the rendered UI; only
  live submit path lives in the unrouted `DocumentEditorPage`.
- **F-2** (medium, FE): draft shows approval-load error instead of clean empty state; source guards
  404 yet runtime errors — needs root-cause.
- **B-1** (medium, backend, out of scope): wrong-stage review-verdict → 500 `internal.unknown`
  instead of a 4xx; integrity preserved, not UI-reachable.

**Gate outcome:** because J2 is **RED**, F2d.8 does **not** pass and the milestone is **not**
all-green. Per separation of powers, **no fixes were attempted** and **milestone-validator was NOT
run** (its precondition is all features closed / all-green). QA artifacts committed; **not pushed**.
