# System-impact analysis — F2d.6 author comment replies

**Date:** 2026-07-09
**Intent (one line):** Surface the author's reply-to / resolve affordance on instance comments in the
`author-waiting` workspace mode (design brief §9.1) — author responds to reviewer comment threads
during review without editing document content.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟢 Green *(see §10)*

> FE-only surfacing of an already-authorized backend capability. Same ten sections; module-birth rows
> marked **N/A** with reason.

---

## 1. Classify & own

- **Work type:** feature (FE-only).
- **Owning module(s):** frontend `features/documents` (workspace screen + comment panel). The backend
  comment endpoints are already owned by the `documents` module (`delivery/http/handler.go`) and ship
  today — this feature adds no backend behavior, only a FE surface.
- **Explicitly NOT owning:** `documents/approval` — comments are document-scoped, not
  approval-instance-scoped; the reply/resolve endpoints live in `documents/delivery/http`, not in the
  approval handler (the milestone-row path citation `approval/…/handler.go:1083` was a drift; the real
  anchor is `documents/delivery/http/handler.go:1083`). `iam` — no new capability, so no IAM change.
- **Cross-module edges (with direction):** none new. FE → existing `documents` comment endpoints
  (`POST/PATCH /documents/{id}/comments…`) via the existing FE API client. No new Go cross-module edge.
- **Ambiguity?** None. AS-3 not triggered.

## 2. Foundation verdict

- **Base you'd build on:** the existing comment stack — backend `AddDocumentComment` /
  `UpdateDocumentComment` (reply = create with `ParentLibraryID`; resolve = update with `Done`), and
  the existing FE hooks `useDocumentComments` / `useDocumentCommentsQuery` and the comment thread UI
  already used in the editor/review canvases.
- **Sound, or legacy/patch/workaround?** Sound. The reply/resolve operations are first-class,
  already-authorized, already-tested backend behavior; the FE hooks are the canonical comment
  primitives. Surfacing them in a new mode reuses the platform, it does not patch around it.
- **If patchy:** N/A. No AS-2.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | No (verified) | Reply/resolve gate on `authorizeDocumentScope` → `CapDocumentView` tenant-grade (`handler.go:1086,1119,1201-1206`), which the author already holds on their own document. No new cap, no new gate. | existing `authorizeDocumentScope` (unchanged) |
| Contract-first (OpenAPI + oapi-codegen) | No | `createComment`/`updateComment` routes already exist in the spec; reply is `createComment` with `ParentLibraryID`, resolve is `updateComment` with `Done`. No route/DTO change. | existing generated comment DTOs |
| Multi-tenant pooled | No | No new query/table; existing endpoints carry tenant scope. | — |
| Async = transactional outbox | No | No external side effect; comment write is a synchronous in-tx state write. | — |
| DB enforces invariants | No | No new invariant/table. | — |
| Cross-module via published interface only | No | FE-only; no new Go cross-module access. | — |

No violation. AS-1 not triggered.

## 4. Capability wiring

**N/A** — no capability added or changed. Reply/resolve are exercises of the existing `CapDocumentView`
tenant-grade gate. `TestCapabilityRegistrySize` unchanged.

## 5. Module wiring

**N/A** — no module born.

## 6. Frameworks to reuse, not reinvent

FE primitives only: existing comment hooks (`useDocumentComments`, `useDocumentCommentsQuery`), the
existing comment-thread component(s) already mounted in the editor/review canvases, `WorkspaceSidebar`'s
`contextualPanel` slot (the same slot F2d.5 S2b used to thread `RequestedChangesPanel`), and
`deriveWorkspaceMode`'s `author-waiting` branch. No new hook, no hand-rolled fetch/mutation. Backend:
no change — the existing `documents` comment service/handler is reused as-is.

## 7. Contract & data

- **OpenAPI-first:** no route added/changed. Endpoints already in the spec.
- **Migration:** none.
- **Destructive change?** none.

## 8. Test & QA plan

- **Canonical framework:** FE = vitest + Testing Library (the class's canonical framework). Backend =
  zero change, so no new Go test required; the milestone row's "real-DB authz test confirming existing
  behavior" is **optional corroboration** (author can reply/resolve on own document under review;
  non-`CapDocumentView` actor cannot) — will include only if it lands cleanly on the `testdb` factory,
  else mark a bounded defer since it asserts already-covered backend behavior.
- **QA gates that apply:** the FE component-behavior gate. Contract / multi-tenant / async / DB-invariant
  gates = N/A (nothing backend touched). Docs gate = light (see §9).
- **Evidence shape:** `vitest run` (comment-panel + workspace author-waiting tests) + `tsc --noEmit 0`
  + zero-backend `git status --porcelain` gate + independent cavecrew-review + evidence.md.

## 9. Docs / ADR

- **Wiki:** light — a note in `wiki/architecture/frontend-structure.md` only if a genuinely new panel
  pattern is introduced; if it reuses the existing comment-thread component in a new mode, refresh the
  affected `Last verified` stamp rather than adding a doc.
- **REQ IDs cited:** design brief §9.1 (author comment replies) — a milestone-brief item, not a
  backend REQ; no backend-target-architecture REQ touched.
- **ADR required?** No. No MUST-deviation, no policy change.

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green — proceed to design. FE-only surfacing of an existing, already-authorized
  capability; no invariant touched; no backend/contract/migration change.
- **Open hard-stops:** none (AS-1/AS-2/AS-3 all clear; the authz claim was targeted-verified at
  `handler.go:1083-1206`).
- **Locked constraints handed to design:**
  1. **Zero backend diff** — reuse `createComment`(reply)/`updateComment`(resolve) verbatim; `git
     status`/grep gate enforces it.
  2. **No content editing** — the author-waiting affordance is reply + resolve on comment threads ONLY;
     it must NOT expose document-body editing (that is `author-editing`/`author-changes-requested`,
     gated by document status, not this mode).
  3. **Reuse canonical comment primitives** — existing `useDocumentComments` /
     `useDocumentCommentsQuery` + existing comment-thread component + `WorkspaceSidebar.contextualPanel`
     slot; no new hook or bespoke fetch.
  4. **Mode-gated surface** — the reply/resolve panel appears in `author-waiting` (and remains available
     wherever comment threads already render), never introduces a new eligibility derivation outside
     `deriveWorkspaceMode`.
