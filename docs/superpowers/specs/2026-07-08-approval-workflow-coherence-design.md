# Approval Workflow Coherence — Governing Design Spec

- **Date:** 2026-07-08
- **Status:** DRAFT — pending operator ratification
- **Program:** approval-remediation (extends the existing program; adds milestones M2d, M3, M4)
- **Supersedes/refines:** M2c F4 sidebar contract (honored, was violated in implementation); ADR 0072 (approval nested in documents — to be superseded by Milestone C's extraction ADR)
- **Related:** ADR 0022 (capabilities), ADR 0074 (route versioning), ADR 0077 (delegation), M2b kernel (F1–F10)

## 1. Problem — evidence, not assertion

### 1.1 The confirmed defect (M2c, live-QA)

Document signoff on a review stage returns `412 precondition.content_hash_mismatch`.
Root cause chain, all verified on runtime:

| Layer | Fact | Evidence |
|---|---|---|
| Contract | F4 spec: footer variant = `activeStage.stage_kind === 'approval'` | `milestone-2c-approval-screen-fe/f4-approval-sidebar-ia/spec.md:60-62` |
| Implementation | `DecisionFooter` branches on `decision != null`; `decision` offered via `policy.actions.signoff` keyed ONLY on document status (`under_review → signoff: true`) | `DecisionFooter.tsx:224`, `useDocumentApprovalArtifact.ts:205-206`, `approvalWorkflow.ts:46-49` |
| Backend | Signoff requires pinned `frozen_content_hash`; review stages have none by design (freeze fires at last review verdict). NULL pin at approval-kind stage is documented "impossible state" | `decision_service.go:245-269`, `freeze.go` |
| Why QA passed | F8 live QA drove review→approval lifecycle via curl; UI only observed on an approval-only route; unit tests passed `decision={null}` fixtures | `f8-close/qa/live-qa-log.md:84-105` |

Net: review stages show the signature panel → 412; `ReviewModeFooter` (verdict CTAs) is
dead code on any route with a review stage. Works only on approval-only routes.

### 1.2 The structural local maxima (audit findings A–F)

- **A — split-brain state:** 4 parallel derivations of "what can the viewer do":
  `TRANSITION_POLICY` (status-only), `resolveEditorMode` (`ApprovalCockpitPage.tsx:41-55`,
  correct), `signoffOffered` (`useDocumentApprovalArtifact.ts:205`), `DecisionFooter` branch.
- **B — multiple destinations:** the two `DocumentShell` **working** surfaces —
  `DocumentEditorPage` (661 lines, `/documents/:id/edit`) + `ApprovalCockpitPage` (348 lines,
  `/approvals/:documentId`) — both mount `DocumentShell` for the same artifact. Separately, the
  document **record** surface `DocumentDetailLayout`/`DocumentDetailRoute` + `distribution` child
  lives at `/documents/:id` (a different altitude — revisions/distribution/lineage/metadata).
  **Destination pinned (operator, 2026-07-08):** the working screen takes the canonical URL
  `/documents/:id`; the record surface survives unchanged at `/documents/:id/details`
  (+ `distribution` child); `/edit` and `/approvals/:documentId` redirect. See milestone-2d
  `## Destination` for the full rationale (canonical-artifact-URL evidence + surface-vs-domain
  ratification).
- **C — hollowed shell:** cockpit renders `ArtifactApprovalScreen` with
  `decision: undefined, approvalChain: null, actions: []` just to reuse its grid
  (`ApprovalCockpitPage.tsx:196`).
- **D — instance state outside react-query:** `useState`/`useEffect` imperative fetch, 1s
  `setInterval` staleness clock, dead `isStale`, `QK.approval` invalidations don't reach the
  instance → `onRefetchInstance` threading + `refetchInstanceRef` ordering hack.
- **E — eligibility invisible to the client:** delegation is resolved in-tx at signoff
  (`decision_service.go:288`); delegates are NOT in the DTO's `actors[]`. No client-side
  derivation can be correct → server must expose viewer facts.
- **F — machine conflation:** `TRANSITION_POLICY` mixes the document-lifecycle machine
  (publish/cancel — correctly keyed by status) with the stage machine (signoff/verdict —
  wrongly keyed by status).

### 1.3 The second approval system (coherence debt)

Templates module runs a complete parallel approval mini-system:
`ApprovalConfig{ReviewerRole *string, ApproverRole string}` per template, own SoD
(`CheckSegregation`), own state machine (`CanTransition(hasReviewer)`), own audit vocabulary
(`templates/domain/approval.go`, `templates/application/approval_config.go`,
`templates/domain/version.go`). No routes, no quorum, no delegation, no freeze, no instances,
no e-signature meaning, invisible to the worklist. Two SoD implementations + two state
machines + two audit vocabularies = split-brain. Milestone B (selectors) would widen the
divergence if templates stay behind.

### 1.4 Domain gaps vs. market (research round)

MetalDocs stage actor model today: `RequiredRole × fixed AreaCode`, resolved by
`ResolveEligibleActors` (`postgres_approval_repository.go:1568-1606`) into the
`eligible_actor_ids` snapshot at submit (`submit_service.go:204,221`). Gaps:

1. No named-user assignment for document routes.
2. `AreaCode` is a fixed literal — cannot express "approver of the document's OWN area";
   one route per profile forces RH and Quality docs through the same fixed-area pool.

Market evidence (decisive): **BPMN 2.0 user task = `assignee` + `candidateUsers` +
`candidateGroups`, combinable on the same task** — the union-of-selectors model is the
industry standard (Camunda docs). Veeva: initiator-selected participants constrained to an
allowed role. Power Automate approval types map exactly to our quorum
(everyone-must-approve = `all_of`, first-to-respond = `any_1_of`, + `m_of_n`). Qualio
anti-pattern to avoid: cannot hot-swap an unavailable approver without reverting to draft —
our drift policy + delegation (ADR 0077) already solve this better; preserve them.

## 2. Ratified decisions

| # | Decision | Rationale anchor |
|---|---|---|
| D1 | Server-derived `viewer` block on the instance DTO; FE never derives eligibility | §1.2-E; GitHub `viewerCanX` pattern |
| D2 | ONE pure FE selector `deriveWorkspaceMode` → discriminated union of workspace modes | kills §1.2-A/F |
| D3 | 1 route, 1 screen: `/documents/:id` single destination; cockpit dies | kills §1.2-B/C; M2c milestone objective text |
| D4 | Instance state moves into react-query | kills §1.2-D |
| D5 | Stage actor spec generalizes to a union of `ActorSelector`s | §1.4; BPMN standard |
| D6 | Approval kernel extracted to top-level `approval` module when the 2nd consumer (templates) arrives; templates rewired onto it | §1.3; supersedes ADR 0072 |
| D7 | Order A → C → B: fix defect first; extract kernel before selectors so selectors are built once, in their final home, serving both consumers on day 1 | avoids double-build |
| D8 | No defers of coherence debt. Non-goals are written refusals, not deferred work | operator directive 2026-07-08 |

## 3. Target architecture (end state)

```
                    ┌──────────────────────────────────────────┐
                    │   approval (top-level bounded context)   │
                    │  Route(subject_kind, subject_key, ver)   │
                    │  Stage{ selectors: []ActorSelector,      │
                    │         kind, quorum, drift, due }       │
                    │  resolver → eligible_actor_ids snapshot  │
                    │  instance · signoff · verdict · freeze   │
                    │  quorum · SoD · delegation · worklist    │
                    └───────────▲──────────────▲───────────────┘
                                │              │
                 documents ─────┘              └───── templates
             (subject_kind=document,        (subject_kind=template,
              subject_key=profile_code)      subject_key=doc_type)

FE: GET instance → { stages, actors(display names), verdicts[], viewer{...} }
    deriveWorkspaceMode(doc, instance, viewer) →
      author-editing | author-waiting | author-changes-requested |
      reviewing | approving | observing | lifecycle
    ONE screen (/documents/:id) adapts by mode (Google-Docs editing/suggesting/viewing model)
```

Invariants preserved untouched: capabilities-not-roles (ADR 0022), contract-first
(OpenAPI + oapi-codegen only), tx-local GUC tenancy, outbox async, DB-enforced invariants,
route versioned-immutable (ADR 0074), freeze boundary semantics (M2b F5), delegation
(ADR 0077), H-PRE-1 (no authz-recording reads inside lock-holding tx; display-name lookups
off-tx).

## 4. Milestone A (program id M2d) — FE workflow coherence + viewer contract

Closes the 412 defect and the single-screen objective. Frontend + contract + thin read-path
backend. Kernel untouched.

### A1 — Contract (contract-first)
- `ApprovalInstance` DTO gains:
  - `viewer` block (server-derived, single source of eligibility truth):
    `is_author`, `eligible_for_active_stage` (snapshot ∪ active delegation, minus SoD
    exclusion), `via_delegation_from` (nullable), `has_signed_active_stage`.
  - `verdicts[]` — review verdict history: actor id + display name, verdict, reason,
    timestamp (closes the review-timeline gap).
  - Actor display names on stage actors (closes `approvalWorkflow.ts:120` TODO).
- Computed in the view-read path (`GetInstanceByDocumentHandler` /
  `LoadInstanceByDocumentForView`); display-name joins off-tx per H-PRE-1.
- OpenAPI edit + `oapi-codegen` regen. No hand-written DTO consumers (ADR 0035 discipline).

### A2 — FE state machine
- ONE pure selector `deriveWorkspaceMode(doc, instance, viewer)` → discriminated union:
  `author-editing | author-waiting | author-changes-requested | reviewing | approving |
  observing | lifecycle`.
- `reviewing`/`approving` derive from `activeStage.stage_kind` + `viewer.eligible_for_active_stage`.
- `TRANSITION_POLICY` shrinks to document-lifecycle-only (publish/cancel by status);
  stage actions leave it entirely.

### A3 — Single screen
- `/documents/:id` is the only working destination (canonical artifact URL — pinned by operator
  2026-07-08); `/approvals/:documentId` AND `/documents/:id/edit` become redirects to it.
- The record surface (`DocumentDetailRoute`/`DocumentDetailLayout` + `distribution`) moves
  unchanged to `/documents/:id/details` — different altitude, kept, not absorbed.
- `ApprovalCockpitPage` + hollowed-shell composition deleted; `DocumentEditorPage` adapts
  by workspace mode; `DecisionFooter` variant = `stage_kind` (F4 contract honored).
- Worklist (`/approvals`) stays; its deep links target `/documents/:id`.
- Visual/UX contract: `2026-07-08-single-screen-design-brief.md` (ratified 2026-07-08),
  including two operator decisions: (1) author may reply to / resolve instance comments in
  `author-waiting` (authz surface verified at feature-spec time; a gap is a contract item,
  not a client workaround); (2) `author-editing` adopts the unified right-sidebar shell —
  `ArtifactMetaSidebar` composition retired.

### A4 — Instance state → react-query
- `useApprovalInstanceQuery` under `QK.approval`; invalidations propagate; delete imperative
  fetch state, 1s staleness `setInterval`, dead `isStale`, `refetchInstanceRef` hack.
- ETag seeding for If-Match preserved inside the query fetcher.

### A5 — Validation gate
- Unit: every `deriveWorkspaceMode` branch, including delegation and SoD-excluded author.
- Real-DB: `viewer` block correctness (author / eligible / delegate / signed / observer).
- Live QA **UI-driven** (browser preview tools, not curl) on a review+approval route AND an
  approval-only route — the exact gap that let the defect through. Review stage must show
  verdict CTAs; approval stage must show signature panel; full lifecycle to published.

## 5. Milestone C (program id M3) — approval kernel extraction + templates unification

### C0 — Gate
- Run `developing-new-work` (system-impact analysis, Green/Yellow/Red) before this
  milestone's `milestone.md` — new top-level module boundary (CLAUDE.md rule).

### C1 — Extraction
- `internal/modules/documents/approval` → `internal/modules/approval` (15th bounded
  context). Cross-module access via application service / published interface only.
- Route generalized: keyed by `(subject_kind, subject_key)` — `document+profile_code`
  (existing rows migrate), `template+doc_type` (new). Versioned-immutable unchanged.
- ADR: kernel extraction + supersede ADR 0072 (rationale: second consumer arrived).

### C2 — Templates rewired onto the kernel
- `ApprovalConfig{ReviewerRole?, ApproverRole}` migrates to a 2-stage (or 1-stage when no
  reviewer) route per doc_type; `CheckSegregation` + `CanTransition(hasReviewer)` +
  template-local approval audit vocabulary retired in favor of kernel SoD, instance state
  machine, and kernel audit events.
- Template versions gain: instances, worklist visibility, delegation, drift policy,
  e-signature meaning on signoff — same eQMS rigor as documents (templates ARE controlled
  documents under ISO).
- Data migration: existing configs → routes; in-flight template versions get a defined
  cutover rule (specified in the milestone, no silent state loss).

### C3 — Validation gate
- `go build ./...` + full test suite; kernel real-DB suites re-run green post-move.
- Template lifecycle live QA: config→route migration verified, review+approve+publish a
  template version through the kernel, worklist shows it, SoD + delegation enforced.
- Regression: document approval lifecycle unchanged (M2b F8-class walkthrough re-run).

## 6. Milestone B (program id M4) — actor selectors

Built once, in the extracted kernel, serving documents AND templates.

### B1 — Domain
`Stage.Selectors []ActorSelector`, discriminated union (BPMN
assignee/candidateUsers/candidateGroups pattern):

| Selector | Fields | Market equivalent |
|---|---|---|
| `named_user` | user_id | DocuSign/Jira named assignee |
| `role_in_fixed_area` | role, area_code | current model, preserved |
| `role_in_document_area` | role | pool follows subject's area |
| `submit_choice` | role (+ area scope) | Veeva initiator-select, constrained |

Validate: ≥1 selector per stage; quorum consistent with minimum resolvable pool;
`submit_choice` requires constraint fields.

### B2 — DB
- Child table `approval_route_stage_selectors`, CHECK constraints per selector type
  (DB enforces invariants). Migration backfills existing stages →
  `role_in_fixed_area(required_role, area_code)`. `RequiredRole`/`AreaCode` columns retired
  after backfill (no dual source of truth).

### B3 — Resolver
- `ResolveEligibleActors` → union of per-selector queries; `submit_choice` picks validated
  server-side at submit (chosen user must satisfy role/area constraint).
- Snapshot `eligible_actor_ids` + quorum/drift/SoD/delegation/freeze **untouched**.
- Drift: pool selectors re-evaluate per drift policy; `named_user` exempt from drift.

### B4 — Contract
- Route DTOs (GET/PUT) carry selectors; submit request gains optional `chosen_actors`
  per stage. OpenAPI + regen.

### B5 — FE
- `RouteEditorDialog` selector builder (replaces raw role/area text fields at :88-105).
- Submit dialog picker when the route demands a choice.

### B6 — Validation gate
- Domain validate tests per selector type + combinations; real-DB resolver tests
  (union, dedup, submit_choice constraint rejection); live QA: build route with mixed
  selectors in UI, submit with picker, full lifecycle for a document AND a template.

## 7. Non-goals (written refusals — not deferred work)

| Item | Why refused | Reopen trigger |
|---|---|---|
| Dynamic rule selectors (manager-of, field-driven) | YAGNI; union is extensible — a new selector type is a new row kind in the child table, no structural migration | A real customer rule that named/pool/choice cannot express |
| W12 parallel (DAG) stages | Serial is correct and consistent system-wide; intra-stage parallelism already exists via quorum (`all_of`/`m_of_n` = concurrent signers); cross-stage DAG rewrites route model, freeze join semantics, drift/cancel policy, and UI for zero present user scenarios. Operator ruled OUT 2026-07-08 | A concrete customer process requiring two simultaneous stages with distinct pools/SLAs |
| Rebuilding the worklist | C3/C5 shipped and passed; only deep-link targets change in A3 | — |

## 8. ADRs to write

1. **Viewer-facts contract** (A): server-derived `viewer` block as the single eligibility
   truth for clients; FE forbidden from deriving eligibility.
2. **Single artifact destination** (A): one screen per artifact, mode-adaptive; cockpit
   pattern retired.
3. **Approval kernel extraction** (C): 15th bounded context, subject-generalized routes;
   supersedes ADR 0072.
4. **ActorSelector model** (B): union-of-selectors as the stage assignment standard
   (BPMN-aligned); drift semantics per selector class.

## 9. Program mechanics

- Milestones join the **approval-remediation** program: A=M2d, C=M3, B=M4, in that order.
- M2c disposition: HS-1 recorded with the confirmed contract violation (§1.1) → remediated
  by M2d. M2c stays PASSED-with-deviation; the deviation closes at M2d's gate.
- Each milestone: `milestone` skill lifecycle — milestone.md spec up front, per-feature
  consumer-contract-first specs, TDD, evidence, independent `milestone-validator`, HS-1
  operator gate. No push without explicit permission.
- Single-screen visual design ratified via `/impeccable shape` →
  `2026-07-08-single-screen-design-brief.md` (feeds A3's feature specs).

## 10. Risks

| Risk | Mitigation |
|---|---|
| C's module move churns many imports | Mechanical move gated by full build+suite green; no behavior change in the move commit itself |
| Template in-flight versions at cutover | Explicit cutover rule in C's milestone.md; no silent state loss |
| A's viewer block leaks authz semantics into DTO | viewer facts are *display* truth; enforcement stays tier-2 `authz.Require` in-tx — DTO never gates the server |
| Selector backfill drift (B2) | Backfill + retire in one migration with verification query; parity test old-resolver vs new-resolver on existing routes |
| UI-driven QA cost | A5/C3/B6 gates name it explicitly; curl-only QA is a validator FAIL condition for FE-facing features |
