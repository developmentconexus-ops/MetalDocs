# System-impact analysis — F2d.5b `f5b-pdf-read-canvas` (PDF-driven viewing for in-approval documents)

**Date:** 2026-07-09
**Intent (one line):** Serve the frozen-revision PDF rendition to in-approval viewers so read modes (approving / observing / author-waiting / lifecycle) render PDF, not the docx/TipTap editor; approver signs over PDF rendition bytes; enables the real editor lazy split.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🔴 **Red** — AS-2 (the foundation the feature is premised on does not exist at the target lifecycle point). *(see §10)*

> Same ten sections for module and feature work. Module-only rows marked **N/A**.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature.
- **Owning module(s):** `documents` — the view surface (`internal/modules/documents/application/view_service.go`, `GET /documents/{id}/view`) and the freeze/materialize pipeline (`freeze_service.go`, `documents/approval/application/decision_service.go`) both live here. FE `documents` feature slice owns `PdfCanvas`.
- **Contributing module:** `render` — the fanout + PDF outbox pipeline (`internal/modules/render/fanout/`) is the artifact producer. Any *new* rendition trigger touches render.
- **Explicitly NOT owning:** `controlleddocuments`, `templates` — they carry no view-surface or rendition responsibility. `iam/authz` is consumed (CapDocumentView) but does not own.
- **Cross-module edges (with direction):** `documents → render` (materialize/PDF outbox, through the fanout client/outbox repo — a published seam, not render's tables). `documents → iam/authz` (`authz.Require`, published). No repo/SQL reach-through introduced.
- **Ambiguity?** No — owning module is unambiguous. Not AS-3.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** the freeze/materialize → PDF pipeline (ADR 0015 + ADR 0009), and `ViewService.GetViewURL` which serves `documents.final_pdf_s3_key` gated to `{approved, scheduled, published}` (`view_service.go:43-47`).
- **Sound, or legacy/patch/workaround?** The pipeline itself is sound. **The premise binding it to this feature is false.** The feature was ratified on the claim *"rendition already materialized at freeze — approver signs over the existing rendition bytes."* Verified against source, that is not how freeze works:
  - **Freeze is terminal, not per-stage.** `Pin` (the in-tx freeze half) fires **only** inside `RecordSignoff` when `instance.Status == domain.InstanceApproved` — i.e. *all* stages complete (`decision_service.go:408,427-435`). The same tx flips the document `under_review → approved` (`:442-449`). The PDF is then produced **asynchronously, after** that terminal signoff (materialize outbox → docgen `/convert/pdf` → `WritePDF` stamps `final_pdf_s3_key`, freeze-and-fanout.md steps 12-17).
  - **Therefore an in-approval document has no PDF.** While a document is `under_review` (any review or approval stage still active), `final_pdf_s3_key` is NULL, `content_hash`/`values_hash` are unwritten, and no `frozen.docx`/`final.pdf` object exists. The `viewableStatuses` gate excludes exactly these states — correctly, because there is nothing to serve.
  - **Content is still mutable during approval.** The `author-changes-requested` edit path (built in F2d.5 S2b) lets the author rewrite the docx mid-workflow. A "frozen" PDF minted before that edit would be immediately stale. So freeze *cannot* simply be moved earlier (to submit or to first-stage) without breaking the mutability the review loop depends on.
  - **Even the final approver has no PDF at signing time.** Their signoff *produces* the freeze; the PDF is materialized async by their own decision. They cannot "view and sign over" a rendition that does not yet exist when they act.
- **Conclusion:** the feature would not be "serve the existing frozen PDF" — it would require **producing a new derived artifact for a lifecycle window that has none, and reconciling that with the signature-subject contract.** That is a foundation change, not an optimization inside the current one → **AS-2**.

**Global-maximum options for the operator (the decision AS-2 surfaces):**

- **Option A — Ephemeral preview rendition (decouple view surface from signature subject).** Add an on-demand docx→PDF *preview* path for in-approval docs (derived, non-immutable, cached by current `content_hash`). Read modes render this preview PDF. The **signature subject stays what it is today** — the approver signs over the source `content_hash` (If-Match / content_hash contract, unchanged, Part-11-clean), *not* over the preview bytes. Cost: a new preview render trigger + cache in `render`; a new `/view` (or `/preview`) branch that serves preview status for `under_review`. Honest: "you view a faithful PDF preview; your signature binds the immutable source hash."
- **Option B — Freeze-at-approval-entry (semantics change).** Move `Pin` to fire when the instance *enters* its first approval-kind stage (review fully done, content quiescent), so an immutable rendition exists for the approval window and approvers genuinely sign over it. Cost: real change to freeze semantics + ISO 9001 immutability timing (ADR 0015/0016 territory → **ADR required**); must prove content is immutable from that point (no `changes_requested` re-open after approval entry) — and it still leaves *review*-stage viewers with no PDF.
- **Option C — Scope PDF viewing to terminal states only (no backend change).** Keep the docx read-only canvas (`DocumentShell`) for in-approval read modes; PDF viewing applies only where a frozen rendition exists (`approved/scheduled/published`, already served). Contradicts the "all viewing is PDF" intent for the in-approval window, but is zero new pipeline and ships the editor lazy split independently.

Recommendation for the operator to weigh: **Option A** is the global maximum for the stated intent ("all viewing is PDF") without corrupting the signature-subject invariant or freeze immutability — it names the honest boundary (view surface ≠ signature subject) instead of pretending an immutable rendition exists mid-approval. Option B is a heavier, ADR-bearing semantics change with a review-stage gap. Option C is the minimal truthful fallback.

## 3. Invariant alignment
*(the 6 non-negotiables)*

| Invariant | Touched? | How satisfied / at risk | Helper to reuse |
|-----------|----------|-------------------------|-----------------|
| AuthZ = capabilities, never roles | Yes | In-approval PDF read still gated by `authz.Require(ctx, tx, CapDocumentView, "tenant")` as today; **new question:** does an *approving/observing* viewer hold `CapDocumentView` while `under_review`? Must verify grants, not assume. | `authz.Require` (`view_service.go:74`) |
| Contract-first (OpenAPI + oapi-codegen) | Yes | Any new preview status/route or widened `/view` response is an OpenAPI-first change (edit spec → regenerate), never hand-added. | `api/openapi/v1/openapi.yaml` + module `cfg.yaml`/`gen.go` |
| Multi-tenant pooled | Yes | Preview object keyed `tenants/{tenantID}/…`; presign via `AssertedPresignGet` (already tenant-guarded); cross-tenant → 404. | `AssertedPresignGet` (`view_service.go:21-23`) |
| Async = transactional outbox | Yes (Option A/B) | A new rendition trigger is a network side effect → must go through the outbox, never an inline docgen call in a handler tx. | `staging_outbox.go:29` |
| DB enforces invariants | Maybe | If Option B changes freeze timing, the `values_frozen_at` / status-transition triggers must move with it — not app-only. | `db/baseline/0001_current_schema.sql` |
| Cross-module via published interface only | Yes | `documents → render` stays on the fanout/outbox seam; no reach into render tables. | render fanout published surface |

No invariant is *violated by construction*, but the **signature-subject risk** (approver signing "over the PDF" when the PDF is a mutable preview) is a Part-11 §11.50 meaning-of-signature hazard — Option A resolves it by keeping the source hash as the subject; Option B by making the PDF genuinely immutable. This is why the feature cannot proceed as premised.

## 4. Capability wiring
**N/A** — no new capability. Feature reuses `CapDocumentView` (view) and the existing signoff caps. Open verify item (§7): confirm in-approval personas actually hold `CapDocumentView` in the seed grants before relying on it.

## 5. Module wiring
**N/A** — feature, not a new module. `documents` + `render` already exist with their ports.

## 6. Frameworks to reuse, not reinvent

- `TxRunner.Do` — the `/view` read tx (already used, `view_service.go:54`).
- `authz.Require` / cap cache — authz gate (already used).
- `AssertedPresignGet` — tenant-guarded presign for the preview object.
- Outbox repo (`staging_outbox.go`) — mandatory for any new rendition trigger (Option A/B).
- `problem.Write` — error responses (pending / failed / forbidden).
- `useDocumentPdfStatus` (FE) — existing polling hook is the exact pattern for `PdfCanvas`; reuse, don't rebuild.
- `testdb` factory — integration coverage for the new `/view` branch.

No hand-rolled equivalents planned.

## 7. Contract & data

- **OpenAPI-first:** Option A adds a preview-status branch to `/documents/{id}/view` (or a sibling `/preview`) — spec edit + regenerate. Option C: no contract change. Option B: no view-contract change but a freeze-timing change behind it.
- **Migration:** Option A — likely a preview-cache/outbox row (keyed by revision + content_hash), `tenant_id` present. Option B — migration to move/duplicate the freeze trigger timing. Option C — none.
- **Destructive change?** None to the live `/view` contract — additive (new status value / new route). Widening must not break current `{pdf_status, signed_url}` consumers.
- **Verify-before-build (the AS-2 crux, already done):** ✅ confirmed freeze is terminal-only (`decision_service.go:408`), ✅ `final_pdf_s3_key` NULL during `under_review`, ✅ content mutable during review (S2b author-edit path). Remaining verify for whichever option is chosen: in-approval `CapDocumentView` grants; whether any preview render path already exists (`render` preview endpoints) before adding one.

## 8. Test & QA plan
*(applies once an option is chosen and re-gated)*

- **Canonical framework:** `testdb` integration (`//go:build integration`) for the new `/view` branch + authz; component tests (vitest) for `PdfCanvas` and the read-mode canvas swap.
- **QA gates that apply:** contract (new response/route), authz (in-approval viewer × CapDocumentView), multi-tenant isolation (cross-tenant preview → 404), async/idempotency (Option A/B rendition trigger). DB-invariant gate applies only to Option B. Docs gate always.
- **Evidence shape:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...`, `.\scripts\check-system-runnable.ps1`, plus UI-driven live QA on the read modes (curl-only = FAIL, per the M2c/F8 lesson) — outcomes + review disposition + bounded defers.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/documents.md` (view surface + PDF-status section), `wiki/workflows/freeze-and-fanout.md` (if the trigger changes), refresh `Last verified`.
- **REQ IDs to cite (once designed):** the view/rendition + approval-lifecycle REQ IDs in `wiki/architecture/backend-target-architecture.md`.
- **ADR required?** **Option A** — likely yes (new derived-preview artifact class + the explicit "view surface ≠ signature subject" boundary is a policy statement worth pinning). **Option B** — yes (freeze-timing semantics change; amends/relates ADR 0015 + the immutability/hashing concept). **Option C** — no.

## 10. Verdict & locked constraints

- **Verdict:** 🔴 **Red.** The feature as ratified ("serve the already-materialized frozen PDF to in-approval viewers; approver signs over rendition bytes") is **infeasible on the current foundation** — freeze is terminal-only, so no PDF rendition exists during the approval window, and content remains mutable through review. Design cannot begin against a premise the code contradicts.
- **Open hard-stops:** **AS-2** (foundation the feature assumes does not exist at the target lifecycle point) — **unresolved; operator must choose A / B / C.** No AS-1 (no invariant violated by the *idea*, but Option B carries an ADR-bearing immutability change and any "sign over the PDF" framing carries a Part-11 §11.50 hazard that Option A/B must each resolve). No AS-3.
- **What must happen before brainstorming:** operator selects the redesign direction (A ephemeral-preview / B freeze-at-approval-entry / C terminal-only PDF). Whichever is chosen, re-run this gate for the *chosen* shape — the locked constraints differ per option (A: new preview artifact + keep source-hash as signature subject; B: ADR amending freeze timing + prove post-entry immutability + accept review-stage gap; C: no backend change, ship lazy split only).
- **Not blocked by this:** the **editor lazy split** does not depend on the PDF question — it can proceed independently under any option (the read canvas simply stops statically importing TipTap). That is the one piece of F2d.5b's value that is unconditionally shippable.

**STOP — Red verdict. No brainstorming handoff. No code. Operator decision required on §2 options A/B/C.**
