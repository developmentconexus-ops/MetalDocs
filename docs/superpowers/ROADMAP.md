# MetalDocs Consolidated Roadmap

**Reset date:** 2026-07-10 (operator-ordered re-plan; supersedes per-program READMEs as the ORDER of work — program folders remain the source of per-milestone detail).
**Rule:** order follows the real dependency DAG, NOT row number. Independent units run as **concurrent tracks** per HARNESS §9 (parallel-track execution, ratified 2026-07-14) — the hub builds a collision matrix (files · FE surface · contract · migration · DB shape · module) and dispatches disjoint units in parallel, serial merge-back through the one acceptance gate (smallest-first). A milestone may `SPLIT-REQUEST` its own independent slices into sibling tracks. Each unit lists its EXACT context files and a main-session token budget — do not read beyond the listed files without a named reason; delegate bulk reads to sonnet subagents (never fable workers).

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
| 3.1 | **M3 kernel extraction** ✅ CLOSED 2026-07-13 (milestone-validator PASS, HS-1 approved; merged `8685b85f` + shared-DB test remediation merged `b54de089` — integration restored to exactly the 9-accepted baseline; NOT pushed) | `documents/approval` → top-level `internal/modules/approval` (15th module); routes generalize to `(subject_kind, subject_key)`; templates kernel routes ADDED as backend truth (`/submit-for-approval` + `/signoff`, contract-first, integration-tested); supersedes ADR 0072. Legacy retirement DEFERRED to 3.1a (operator Option A 2026-07-12). Validator verdict: `docs/superpowers/reports/2026-07-12-m3-milestone-validation.md`. Follow-on debt: idempotency-key drift (`submit_service.go:81` raw-header passthrough, `ComputeIdempotencyKey` gone), testdb per-pid template-DB leak, stale module-count strings. Spec §5 `2026-07-08-approval-workflow-coherence-design.md:177-205`. | MUST run `developing-new-work` first (new module boundary) | ≤500k (milestone, multi-session) |
| 3.1a | **template legacy-approval retirement** ✅ CLOSED 2026-07-13 (reviewer ACCEPT 7/7; merged `96961e81`; ladder green — integration == 9-accepted baseline (+1 pre-existing OCC flake, green isolated); 0301/0302 applied live (`templates_approval_config` dropped, pre-check 0 rows); containers rebuilt + :80 smoke green; L3 browser QA = operator defer; HS-1 items in evidence `2026-07-13-unit-3.1a-evidence.md` §CLOSE-OUT; NOT pushed) | Distinct from 3.2. Sequenced: (a) migrate `CreateTemplate` (stop role-seeding) + `PublishTemplateVersion` (kernel-driven, drop role-SoD) off role model; (b) rebuild template-approval FE onto kernel routes (`TemplateApprovalRoute.tsx`/`TemplateEditorPage.tsx` call only legacy today; 0 kernel consumers); (c) delete legacy path (4 routes + handlers + `Service.SubmitForReview/Review/Approve/UpsertApprovalConfig` + `approval_config.go` + `GetApprovalConfig`/non-Tx `UpsertApprovalConfig` repo + legacy tests + FE) + DROP `templates_approval_config` (pre-drop emptiness assert) + `CapTemplateReview`/`domain.ApprovalConfig` orphan cleanup. Blocked because table/type shared with non-legacy create+publish (ADR 0082 §Transitional coexistence). Kernel + legacy coexist until this lands. | `developing-new-work`; depends 3.1; FE tooling (see `fe-node-modules-junction-drift`) + browser-QA (operator) | ≤400k |
| 3.2 | **M4 ActorSelector** | BPMN-aligned selector union + `approval_route_stage_selectors` table + RouteEditorDialog selector builder + submit-choice picker. Spec §6 (`:207-249`). Depends 3.1. | `developing-new-work` | ≤400k |

Ordering rationale (locked): G1–G3 land INSIDE the nested kernel BEFORE extraction — extraction then moves complete, ratified-model-conformant code once. Reverse order would extract a kernel that immediately needs domain surgery.

## 4. LATER — lifecycle tail + hygiene (PARALLELIZED 2026-07-14, HARNESS §9)

**Dependency truth (not row order):** 4.2 ⊥ 4.3 ⊥ 4.4 — three near-orthogonal tracks, verified
disjoint on all six collision axes (only 4.2 touches FE; only 4.2 touches `api/openapi`; only 4.4
adds a migration). Old "after 4.1/4.2 / after 4.3" ordering was serialization-by-decree, now
retired. Run concurrently (operator-approved 3 tracks, 2026-07-14):

```
Track A (feature)     4.2 template inbox        FE approval/InboxPage + approval kernel worklist query + api/openapi
Track B (test-infra)  4.3 E-PROD zero-baseline  Go tests + testdb-clone; owns TenantDataPort registration exclusively
Track C (mechanical)  4.4 hygiene (−leak item)  migration 0306 + submit_service idempotency + string sweep
                              └─ testdb-leak sub-item defers until B merges (or B absorbs it)
```

**Collision-matrix ownership (written into each chip):** migration block — C owns `0306+`, A/B add
none (max on main = 0305). Exclusive files — B owns `TenantDataPort` registration; A `REQUEST`s if
its worklist query needs a new tenant-table port (a read query should not). Contract lock — only A
edits `api/openapi`. Merge-back — serial through the one acceptance gate, **smallest-first C→B→A**;
later track rebases on merged main only if shared infra moved. Zero-new-RED holds per track on the
integrated main.

| # | Unit | Scope | Depends on |
|---|------|-------|------------|
| 4.1 | **M3 journey closure** ✅ CLOSED 2026-07-14 (branch `unit-4.1-journey-closure`, base `a684cb4e`, range `a684cb4e..daf471a5`; 4 commits, net 17 files +177/−380; L0 tsc EXIT=0, L1 vitest 81 files/546 tests; each slice + final net diff independently reviewed "No issues"; NOT pushed) | Deep links (cockpit↔detail, notifications, fanout CTA); delete dead FE affordances (findings 9–12, 20). Spec: `2026-07-06-lifecycle-ux-coherence-design.md`. **Runtime-truth:** F9/F10 already satisfied by ADR 0080 (evidence-only); F11 notification deep-link + F12 fanout-CTA route fix BUILT; F20 dead affordances DELETED. One defect found+fixed: slice-A `git add -A` on incomplete worktree checkout committed 2237 spurious deletions → restored `daf471a5`. L3 browser QA → operator (login-password prohibition). Evidence: `docs/superpowers/reports/2026-07-14-unit-4.1-evidence.md`. | 1.1 (M2 HS-1). Independent otherwise — may interleave after §2 screens to avoid same-surface collisions. |
| 4.2 | **M4 template inbox** (Track A) | Template reviews in single approver worklist, contract-first (finding 15). Consumes the EXTRACTED kernel worklist (3.1 done, legacy retired in 3.1a) — not any legacy template-review endpoint. Owns `api/openapi` edits this batch. | **3.1** ✅ (kernel extracted). Parallel-safe with 4.3/4.4 (disjoint surfaces). |

| 4.3 | **E-PROD zero-baseline + testdb-clone isolation** ✅ CLOSED 2026-07-14 (Track B, branch `claude/priceless-kilby-cc7a88`, range `7e046056..165347cd`, merge `b626538d`; baseline 9→1 exactly per revised exit; **dual-reviewer parity**: Claude cavecrew ACCEPT 0-findings + GPT-5.6 SOL CLEAN — GPT ran HUB-side via stdin evidence-pack because `codex-windows-sandbox-setup.exe` is missing machine-wide (unit-side codex cannot exec; operator fix = interactive `/codex:setup`); post-merge ladder: go build OK, scenarios PASS 326s, tenantdata PASS 14.6s on integrated main; residual 1 RED = seq-counter = unit 4.5 documented baseline; NOT pushed) | Kill the 9-accepted-RED baseline. Sequenced: (1) migrate the 4 non-isolated packages (`tests/integration/iam`-style shared-DB suites: `controlleddocuments/application`, `jobs/approval_sla_surfacer`, `tests/integration/scenarios`, `tests/integration/tenantdata`) to per-test testdb-clone (ADR 0034 framework, already canonical elsewhere); (2) re-run on clean DB — surviving RED = genuine product defect; (3) fix those at root (incl. registering `approval_delegations` + `approval_review_verdicts` TenantDataPorts). Exit: baseline = 0, gate = binary suite-green. Also kills `OCC_Race_N50` flake. Named trigger for wider sweep: any OTHER package showing pollution-flake migrates immediately. **(4) Suite-runtime budget (operator 2026-07-13: "50min é piada"):** audit the 3 worst packages — `tests/integration/approval` 1471s, `tests/integration/migrations` 1055s, `internal/platform/idempotency` 1041s — for real-sleep waits (janitor/TTL/poll loops) → fake-clock or short-TTL injection; add `t.Parallel()` where per-test clones make it safe; exit budget: full integration ≤ 15 min wall. Platform levers already landed hub-side 2026-07-13: `synchronous_commit=off` on dev postgres (compose) + selective-integration ladder policy + Go test cache re-enabled (HARNESS §5 L1). **Track B: owns `TenantDataPort` registration exclusively this batch; adds no migration.** **PREMISE CORRECTED 2026-07-14 (Track B ESCALATION + hub investigation):** step (1) is a NO-OP — all 4 target packages ALREADY use per-test testdb.Open clones at `7e046056`. The 9-RED baseline is NOT shared-DB pollution; they are GENUINE defects exposed on clean clones, root-diagnosed. Revised 4.3 exit = drive baseline 9→1 by fixing the genuine defects B can own: tenantdata ports (2), TriggerBypassBlocked (CI-role non-superuser), **sla_surfacer ambiguous-column SQL bug** (hub granted B exclusive ownership of `sla_overdue_reader.go`+`sla_surface_writer.go` — live River SLA job broken, 42702, qualify `asi.*`; 4 tests), **grant_area_membership dead-fn** (0 product callers, fn regex `^[A-Z0-9_]+$` ⊥ table CHECK lowercase = unsatisfiable → delete the 2 dead tests + fix `e2e_seed.go:579` call). Residual 1 RED = seq-counter → **unit 4.5** (below). Runtime-budget step (4) → **unit 4.6** (descoped, out of B's surface). | none — parallel-safe with 4.2/4.4 (Go test-infra, disjoint from FE + feature surfaces). |
| 4.4 | **Schema/symbol hygiene** (operator-approved 2026-07-13) | Mechanical sweep: DROP dead `templates_template_version.pending_reviewer_role`/`pending_approver_role` (write-never/read-never since 3.1a, migration + pre-drop assert); M3 idempotency-key drift (`submit_service.go:81` raw-header passthrough — restore derivation or ratify); stale "14 modules" count strings → 15. Cheap, haiku/sonnet workers. **Track C: owns migration `0306+`; the testdb per-pid template-DB leak sub-item defers until Track B (4.3) merges or absorbs it — the other 3 items run now.** **+ micro-item (hub-routed 2026-07-14): DROP the dead `grant_area_membership` SQL function (migration `0307`; 0 product callers, self-contradictory regex; Track B deletes its tests/e2e_seed call, C drops the fn).** | none for the 3 now-items — parallel-safe with 4.2/4.3 (migration number pre-reserved, dropped cols have 0 readers). Leak sub-item: after 4.3. |
| 4.5 | **`document_profiles` tenant-scoped PK** (filed 2026-07-14 from 4.3 escalation; needs operator prioritization) | Real multi-tenancy schema defect: `document_profiles` has `tenant_id NOT NULL` + `UNIQUE(tenant_id,code)` + 3 core children already FK to `(tenant_id,code)` (approval_routes, cd_sequence_counters, controlled_documents), and all reads filter tenant_id (per-tenant data) — yet **PK = `(code)` global**, which forbids two tenants sharing a profile code (breaks the pooled-tenant invariant). Fork resolved (a): migrate PK → `(tenant_id, code)`. Blast radius = the 4 children that FK to `code` alone (document_profile_governance, document_profile_schema_versions, document_sequences, template_drafts) — each needs a tenant_id-aware FK; design pass required. Unblocks `TestTenantIsolation_SequenceCounters_CrossTenant` (correct test, schema-forbidden today — kept RED as named defect until this lands). | `developing-new-work` (schema/invariant change); after 4.3 (which documents the residual RED). Operator: prioritize vs push. |
| 4.6 | **Integration suite-runtime budget** (descoped from 4.3 step-4, 2026-07-14) | Audit the 3 worst packages — `tests/integration/approval` 1471s, `tests/integration/migrations` 1055s, `internal/platform/idempotency` 1041s — for real-sleep waits (janitor/TTL/poll loops) → fake-clock/short-TTL injection; add `t.Parallel()` where per-test clones make it safe. Exit: full integration ≤ 15 min wall. Platform levers already landed (`synchronous_commit=off`, selective-integration policy, Go test cache — HARNESS §5 L1). | after 4.3/4.4 (touches migrations pkg = Track C overlap; sequence to avoid collision). |

Deferred register (findings 18/19/21/22/23): `docs/superpowers/milestones/lifecycle-ux-coherence/README.md`.

## 5. Operator decision queue (no agent work)

1. ~~HS-1: lifecycle M2~~ **APPROVED 2026-07-13** (91195a76).
2. ~~Final sign-off: GMR program~~ **SEALED 2026-07-13** (91195a76).
3. Final architecture review P1–P4 remediation list — pending operator OK (memory `final-architecture-review-2026-07-03`).
4. ~~Push consolidation~~ **DONE 2026-07-13**: 621 commits pushed, origin/main == main @ 91195a76. Standing rule unchanged: every future push needs explicit operator permission.
5. F-18 fresh-repo re-baseline at v1 (memory `f18-history-fresh-repo-at-release`) — unchanged by the push; history purge happens at re-baseline.
6. Consolidated L3 browser QA session (operator): F2d.8 walkthrough + route builder v2 + rebuilt template approval flow — :80 stack rebuilt 2026-07-13, images current.
7. ~~Debt units scope-OK~~ **APPROVED 2026-07-13** → rows 4.3 (E-PROD zero-baseline + testdb-clone) and 4.4 (schema/symbol hygiene).

## 6. Token discipline (how to work each unit)

- Open ONLY the unit's listed context files. The program folder's milestone.md/spec sections are the boundary — no tree crawls.
- Bulk inventory/reading → sonnet subagent returning compressed report (this roadmap was built that way: 3 research agents + 1 inventory agent, ~300k subagent tokens, main session stayed lean).
- Self-compact at 200k main-session tokens: flush durable state to the unit's evidence.md / this roadmap, then compact.
- Budgets above are main-session ceilings per unit; blowing one = stop, flush state, split the unit.
