# MetalDocs Consolidated Roadmap

**Reset date:** 2026-07-10 (operator-ordered re-plan; supersedes per-program READMEs as the ORDER of work — program folders remain the source of per-milestone detail).
**Rule:** work top-to-bottom. One unit at a time. Each unit lists its EXACT context files and a main-session token budget — do not read beyond the listed files without a named reason; delegate bulk reads to sonnet subagents (never fable workers).

---

## 0. Standing constraints (apply to every unit)

- **Execution harness is binding:** `docs/superpowers/HARNESS.md` — model matrix, unit loop P0–P7, verification ladder L0–L4, fresh-session browser-QA protocol.
- Gates: `developing-new-work` before any new feature/module design; milestone-validator at milestone end; operator HS-1 before status flips; evidence.md before closure.
- Contract-first (`api/openapi` + oapi-codegen), capabilities-not-roles (ADR 0022), testdb factory, outbox, DB-enforces-invariants, H-PRE-1.
- Commit after verified work; NEVER push without operator permission. Many sealed programs are local-only — push consolidation is an operator decision (see §5).
- Workers: sonnet implement/review, haiku mechanical, ≤15 concurrent. Fable = synthesis/design only.

---

## 1. NOW — close the nearly-closed (unblocks 3 programs)

| # | Unit | What's left | Blocker type | Context files (read ONLY these) | Budget |
|---|------|-------------|--------------|--------------------------------|--------|
| 1.1 | lifecycle-ux-coherence **M2 HS-1** | Nothing to implement. Operator reviews validator PASS and approves; then flip statuses, scaffold M3. | **Operator action** | `docs/superpowers/milestones/lifecycle-ux-coherence/milestone-2-fe-surface-ownership/qa/milestone-qa.md` | ≤20k |
| 1.2 | approval-remediation **M2d F2d.8** | UI-rendered live QA walkthrough (browser preview session): review stage shows NO signature panel; full draft→submit→verdict→signoff→publish through UI, no 412. Closes M2c's F4 deviation. Then milestone-validator + HS-1. | Needs browser-preview session | `.../milestone-2d-workflow-coherence-fe/milestone.md`, `.../f8-close-live-qa/qa/live-qa-log.md`, spec §4 `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` | ≤120k |
| 1.3 | frontend-screen-completion **M5 F5.2** | **CLOSED 2026-07-10 (verify-only).** Restyle was already delivered by `2a371d60 (FE-14, 2026-07-02)` — page is now `pages/TaxonomyAdminPage.tsx` + tokenized `.module.css`, **0 inline `style=`** (row premise was stale, predated FE-14). Verified: grep=0, tsc EXIT=0, vitest 23/23, rendered UI GREEN, both reviewers APPROVE. Evidence: `.../f5.2-taxonomy-restyle/evidence.md`. **NEXT:** milestone-validator on M5 → HS-1 → mission terminal acceptance (mission-validator, per-screen re-audit fan-out per mission.md §8). | None | `.../frontend-screen-completion/milestone-5-signoff-taxonomy/milestone.md`, `.../f5.2-taxonomy-restyle/evidence.md`, mission.md §7–§8 | ≤100k |
| 1.4 | **GMR final sign-off** | Terminal acceptance PASSED (run-2 10/10). Operator final sign-off only. | **Operator action** | memory `gmr-program-terminal-acceptance-passed` | — |

## 2. NEXT — review/approval workflow model (ratified 2026-07-10, gate passed Yellow)

Governing spec (LOCKED): `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md`.
System-impact analysis (committed): `docs/superpowers/analysis/2026-07-10-review-approval-workflow-model-system-impact.md`.
Research-validated design decisions (operator-locked 2026-07-10): taxonomy owns `GovernanceClass{controlado,simples,livre}` + pure `RoutePolicy()` derivation; approval consumes narrow `RoutePolicy` via NEW narrow `approval→taxonomy` read port (NOT CDFieldReader); Livre ACTIVELY blocks route creation; DB deferrable trigger BOTH directions (route writes + profile reclassification guard); expand-only migration default `controlado`; ADR required (per-profile-never-global + predicate-rule traceability + type-object-table evolution note).

| # | Unit | Scope | Context files | Budget |
|---|------|-------|---------------|--------|
| 2.1 | **G1** per-profile signature policy | Spec → plan → TDD → evidence → commit. Taxonomy column+domain+contract; approval RoutePolicy + Route.Validate(policy) at route-admin+submit; trigger; ADR; module-boundaries.yml edge. | analysis doc above; `internal/modules/taxonomy/domain/profile.go`, `domain/port.go`, `application/profile_service.go`, `infrastructure/repository.go`; `internal/modules/documents/approval/domain/route.go`, `application/route_admin_service.go`, `application/submit_service.go`; `db/migrations/0286_*.sql`, `0287_*.sql` | ≤300k |
| 2.2 | **G2** request_changes on approval stages | Relax `ErrVerdictWrongStageKind` → allow ONLY `request_changes` (never `ready`), in service pre-check `review_verdict_service.go:128` AND `domain.NewVerdict`. No DB change (verified: 0286 has no verdict×stage CHECK). Kernel keeps signed-reject (`signature_meaning='rejection'`). | `.../application/review_verdict_service.go`, `.../domain/` verdict files, spec R3 | ≤150k |
| 2.3 | **G3** fast-forward "Aprovar já" | Eligibility detection (verdict completes review stage by quorum ∧ eligible on now-active approval stage) + TWO ledger entries one tx (H-PRE-1: authz reads off-tx). Contract addition. Freeze boundary unchanged. | `.../application/review_verdict_service.go`, `decision_service.go` (quorum eval), signoff service, spec R5 | ≤300k |
| 2.4 | **Screen: approver execution panel** + reject-default fix | Two actions ("Assinar e aprovar" password+legal; "Solicitar mudanças" comment-only). Remove signed "Assinar e devolver" from UI. Fix `ArtifactDecisionPanel` `defaultOptionKey="reject"` → no preselection. Mock: `approver_execution_screen_mock`. UI-driven QA on :80. | FE `features/approval` decision-panel components; G2/G3 generated API types | ≤200k |
| 2.5 | **Screen: route builder v2** | Profile-governed approval section (Controlado badge "obrigatório ≥1 assinatura" / Simples "opcional"), quorum pills, sequential rounds, live flow preview, overlap→"Aprovar já" note, author-excluded note. Mock: `route_builder_mock_v2`. Needs G1 contract (governance_class). NOTE: approval-remediation M4 will EXTEND this with ActorSelector builder — build v2 so selector area is a clean extension slot. | FE route-editor components; G1 generated API types | ≤250k |

## 3. THEN — approval-remediation architecture tail

| # | Unit | Scope | Gate | Budget |
|---|------|-------|------|--------|
| 3.1 | **M3 kernel extraction** ✅ CLOSED 2026-07-12 (milestone-validator PASS, HS-1 approved; HEAD `ff806655`, NOT pushed) | `documents/approval` → top-level `internal/modules/approval` (15th module); routes generalize to `(subject_kind, subject_key)`; templates kernel routes ADDED as backend truth (`/submit-for-approval` + `/signoff`, contract-first, integration-tested); supersedes ADR 0072. Legacy retirement DEFERRED to 3.1a (operator Option A 2026-07-12). Validator verdict: `docs/superpowers/reports/2026-07-12-m3-milestone-validation.md`. Follow-on debt: idempotency-key drift (`submit_service.go:81` raw-header passthrough, `ComputeIdempotencyKey` gone), testdb per-pid template-DB leak, stale module-count strings. Spec §5 `2026-07-08-approval-workflow-coherence-design.md:177-205`. | MUST run `developing-new-work` first (new module boundary) | ≤500k (milestone, multi-session) |
| 3.1a | **template legacy-approval retirement** | Distinct from 3.2. Sequenced: (a) migrate `CreateTemplate` (stop role-seeding) + `PublishTemplateVersion` (kernel-driven, drop role-SoD) off role model; (b) rebuild template-approval FE onto kernel routes (`TemplateApprovalRoute.tsx`/`TemplateEditorPage.tsx` call only legacy today; 0 kernel consumers); (c) delete legacy path (4 routes + handlers + `Service.SubmitForReview/Review/Approve/UpsertApprovalConfig` + `approval_config.go` + `GetApprovalConfig`/non-Tx `UpsertApprovalConfig` repo + legacy tests + FE) + DROP `templates_approval_config` (pre-drop emptiness assert) + `CapTemplateReview`/`domain.ApprovalConfig` orphan cleanup. Blocked because table/type shared with non-legacy create+publish (ADR 0082 §Transitional coexistence). Kernel + legacy coexist until this lands. | `developing-new-work`; depends 3.1; FE tooling (see `fe-node-modules-junction-drift`) + browser-QA (operator) | ≤400k |
| 3.2 | **M4 ActorSelector** | BPMN-aligned selector union + `approval_route_stage_selectors` table + RouteEditorDialog selector builder + submit-choice picker. Spec §6 (`:207-249`). Depends 3.1. | `developing-new-work` | ≤400k |

Ordering rationale (locked): G1–G3 land INSIDE the nested kernel BEFORE extraction — extraction then moves complete, ratified-model-conformant code once. Reverse order would extract a kernel that immediately needs domain surgery.

## 4. LATER — lifecycle-ux-coherence tail

| # | Unit | Scope | Depends on |
|---|------|-------|------------|
| 4.1 | **M3 journey closure** | Deep links (cockpit↔detail, notifications, fanout CTA); delete dead FE affordances (findings 9–12, 20). Spec: `2026-07-06-lifecycle-ux-coherence-design.md`. | 1.1 (M2 HS-1). Independent otherwise — may interleave after §2 screens to avoid same-surface collisions. |
| 4.2 | **M4 template inbox** | Template reviews in single approver worklist, contract-first (finding 15). | **3.1** — templates rewire onto extracted kernel first; building this before M3 = guaranteed rework. |

Deferred register (findings 18/19/21/22/23): `docs/superpowers/milestones/lifecycle-ux-coherence/README.md`.

## 5. Operator decision queue (no agent work)

1. HS-1: lifecycle M2 (§1.1) — validator PASS waiting.
2. Final sign-off: GMR program (§1.4).
3. Final architecture review P1–P4 remediation list — pending operator OK (memory `final-architecture-review-2026-07-03`).
4. **Push consolidation**: GMR HEAD, lifecycle M1–M2, frontend-screen M0–M5, approval-remediation M2b–M2d, workflow-model commits — all local-only. Decide when to push.
5. F-18 fresh-repo re-baseline at v1 (memory `f18-history-fresh-repo-at-release`).

## 6. Token discipline (how to work each unit)

- Open ONLY the unit's listed context files. The program folder's milestone.md/spec sections are the boundary — no tree crawls.
- Bulk inventory/reading → sonnet subagent returning compressed report (this roadmap was built that way: 3 research agents + 1 inventory agent, ~300k subagent tokens, main session stayed lean).
- Self-compact at 200k main-session tokens: flush durable state to the unit's evidence.md / this roadmap, then compact.
- Budgets above are main-session ceilings per unit; blowing one = stop, flush state, split the unit.
