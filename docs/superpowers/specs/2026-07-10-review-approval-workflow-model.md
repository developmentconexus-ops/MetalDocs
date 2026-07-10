# Review/Approval Workflow Model — ratified domain spec

**Status:** Ratified by operator 2026-07-10 (interview-driven; "Ficou coerente sim" + "Okay confirmado").
**Scope:** Domain model + UX model only. No code in this spec. Backend gaps listed in §4 are
post-milestone feature work; each must run the `developing-new-work` gate before design.
**Owning modules:** `documents/approval` (kernel), `documents` (profiles/routes), frontend `features/approval`.

---

## 1. The five rules (ratified)

### R1 — Route shape: `review* → approval*`, ≥1 stage total; signature minimum is per-profile policy
A route is zero-or-more review stages followed by zero-or-more approval stages. Whether at
least one approval (signature) stage is REQUIRED is a **document-profile policy**, not a
system-global invariant:

| Profile | Signature policy | Example |
|---|---|---|
| Controlado | ≥1 approval stage required | POP, IT, desenho, FMEA |
| Simples | review-only route allowed (ends "Conferido") | nota fiscal, orçamento, relatório de rotina |
| Livre | no route at all | rascunhos, material não governado |

Backend truth supporting this: `AdvanceStage()` is kind-agnostic — a review-only route already
terminates in `InstanceApproved` → `documents.status='approved'`. That is a FEATURE under the
Simples profile, not a bug.

### R2 — Rounds are sequential filters; people in a round are parallel with quorum
Multiple review stages = rounds in sequence (e.g. revisão técnica → revisão qualidade); next
round activates only when the previous completes. People inside one round act in parallel;
completion is quorum-governed: `any_1_of` / `all_of` / `m_of_n`. Same quorum vocabulary on
approval stages.

### R3 — "Quem pode assinar também pode conversar" (approval ⊃ review powers)
On an approval stage the actor gets exactly two actions:
- **Assinar e aprovar** — e-signature ceremony (password re-auth, meaning-of-signature, legal
  checkbox; 21 CFR Part 11 shape already implemented in `ArtifactDecisionPanel`).
- **Solicitar mudanças** — no password, required comment, returns document to author.

The signed "Assinar e devolver" (signed rejection) leaves the UI; the kernel keeps the
capability (ledger/API can still record a signed reject where regulation demands it).
Consequence: an approver never needs a review stage inserted "for himself" — the power to
converse travels with the power to sign. Review stages exist to put OTHER people (who don't
sign) in front of the signers.

### R4 — Author never reviews/approves own document (SoD)
Author is auto-excluded from every stage of the route for his own document. Operator decision:
no exceptions.

### R5 — Overlap is natural; fast-forward "Aprovar já"; never pre-sign
Same person may appear in review and approval lists. Fast-forward is offered iff BOTH:
(a) this person's review verdict COMPLETES the review stage (quorum satisfied by it), and
(b) the person is eligible on the now-active approval stage.
Then one gesture records TWO ledger entries — the review verdict and the signature — separate
audit records, single UX ceremony. No signing before content freezes; freeze boundary stays at
the end of the last review-kind stage (approver-only route ⇒ freeze at submit).

## 2. Behavior consequences

- **Approver-only route (0 reviews):** content freezes at submit; approver has both R3 actions;
  request_changes thaws, returns to author; resubmit re-runs route from the top.
- **Review-only route (Simples):** terminal state reads "Conferido" in UX (status remains
  `approved` in the kernel — label only).
- **Ledger vs ceremony:** audit records never collapse; only UX ceremonies do (R5).
- **Request-changes from approval stage:** exits freeze, returns to author, resubmission
  restarts the route from the first stage.

## 3. Real-world anchors (operator-validated)

- Só aprovador: ata de análise crítica; política de diretoria; concessão/desvio; certificado interno.
- Revisor + aprovador: POP (téc → qualidade → gerente); desenho (par técnico → eng-chefe); PPAP/FMEA; manual cross-área (RH → jurídico → diretor).
- Só revisão: nota fiscal, orçamento, relatório de rotina.
- Separação conceitual: **revisor = quem entende do conteúdo; aprovador = quem responde por ele.**

## 4. Backend gaps (future feature work, each needs `developing-new-work` gate)

| # | Gap | Today | Target |
|---|---|---|---|
| G1 | Per-profile signature policy | No validation links document profile → route shape (`Route.Validate` has no stage_kind rule) | `Route.Validate` parameterized by profile policy (Controlado ⇒ ≥1 approval stage) |
| G2 | request_changes on approval stage | `ErrVerdictWrongStageKind` blocks ANY verdict on approval-kind stage (`review_verdict_service.go:128`) | Relax guard: allow ONLY `request_changes` (never `ready`) on approval-kind stages |
| G3 | Fast-forward "Aprovar já" | Nothing | Eligibility detection (verdict-completes-stage ∧ eligible-on-next-approval) + two-entry ledger write |

Frontend follow-ups: route builder per §mock `route_builder_mock_v2` (profile-driven approval
requirement, quorum pills, flow preview, overlap note); approver execution panel per
`approver_execution_screen_mock` (two actions, ceremony split).

## 5. Known adjacent finding

`ArtifactDecisionPanel` receives `defaultOptionKey="reject"` — reject preselected as default
decision. Dangerous default; unresolved; carried for HS-1.
