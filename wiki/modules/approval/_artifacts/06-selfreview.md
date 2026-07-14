# Phase 6.75 — Self-review (approval)

**Date:** 2026-05-10. Re-read of `wiki/modules/approval.md` + `wiki/modules/approval-tech-debt.md` + `wiki/backlog/approval-refactor.md` against artifacts 00–05.

## Checklist

1. **Severity rubric application.** Re-rated every Critical and Major row.
   - T-001 RFC 9457 drift → Critical (trigger: "Contract violation downstream consumers rely on"). Confirmed — frontend `ApprovalError` parser depends on legacy shape.
   - T-002 OpenAPI absence → Critical (same trigger, contract). Confirmed — codegen blocked, hand-rolled types drift.
   - T-003 substring classifier → Major (trigger: "Documented contract not followed; measurable user impact"). Borderline-Critical because regulated path returns 500; sticking with Major because a 500 vs 409 mis-classification is not data-loss / not authz bypass / not audit gap. Justification recorded in T-003 Observation.
   - T-004 deprecated PDF dispatcher → Major (trigger: "Duplicated write surfaces with different semantics for the same use case"). Confirmed.
   - T-005 inbox snapshot drift → Major (trigger: "Documented contract not followed"). Borderline Minor; pagination total inconsistency under concurrent signoff is user-visible. Stays Major.
   - T-006 tripwire pairing audit incomplete → Major (trigger: "Defense-in-depth gap"). Stays Major because cancel/cutover not yet verified; if verified clean later, downgrade.

2. **Mermaid box ↔ prose.** Walked every box in the C4 Context (§3) and Container (§5.1) diagrams.
   - Context: `approver`, `author`, `approval`, `documents`, `iam`, `render`, `jobsED`, `jobsSW`, `pg` — all named in §3.2 or §7.
   - Container: `http`, `app`, `domain`, `repo`, `infra1` (`infra/signature`), `infra2` (`infrastructure`), `pg`, `iam`, `render` — all named in §5.2 or §8.
   - PASS.

3. **Top-3 in §11.** Ordered by severity (both Criticals first), then blast-radius (T-001 affects every error path, T-002 affects every type-gen surface, T-003 single endpoint).
   - PASS.

4. **Cross-link existence.** Sampled:
   - `wiki/decisions/0007-two-tier-authz.md` — exists (read in Phase 0).
   - `wiki/concepts/iso-segregation.md` — exists (read in Phase 0).
   - `wiki/concepts/authz-tiers.md` — exists (read in Phase 0).
   - `wiki/workflows/approval.md` — exists (referenced in Phase 0).
   - `wiki/modules/iam-tech-debt.md` — exists (read in Phase 0).
   - `wiki/backlog/caixa-aprovacao.md` — referenced in superseded stub; assumed-exists; wiki-curator (Phase 7) will verify.
   - `wiki/concepts/error-ux.md` — referenced in superseded stub; assumed-exists.
   - `wiki/architecture/data-model.md` — referenced in tech-debt T-005; assumed-exists; wiki-curator to verify.
   - PASS with two assumed-exists items deferred to wiki-curator.

5. **Key Files freshness.** Sampled three anchors against current code:
   - `application/decision_service.go:88` (`RecordSignoff`) — verified in Phase 2 read.
   - `application/submit_service.go:43` (`SubmitRevisionForReview`) — verified in Phase 2 read.
   - `repository/postgres_approval_repository.go:316` (`FOR UPDATE` line) — Phase 4 + signoff trace cite `:314` and `:316`. Doc cites `:316`. Match.
   - PASS.

6. **Backlog ↔ debt linkage.** All 12 T-NNN have R-NNN counterparts (R-001..R-012). Tally script Check 3 confirmed no orphans printed. PASS.

7. **Industry citations.** `_artifacts/05-industry.md` cites IP-001, IP-002, IP-003, IP-004, IP-006, IP-007, IP-008 — all present in `references/industry-patterns-index.md`. Transactional outbox explicitly NOT cited as industry pattern (no row added). PASS.

8. **Subagent purity.** Re-skimmed:
   - `02-flow-inbox.md` — facts only. PASS.
   - `02-flow-signoff.md` — facts only. PASS.
   - `02-flow-submit.md` — facts only. PASS.
   - `03-deps.md` — facts only. PASS.
   - `04-persistence.md` — facts only. PASS.
   - No "should/recommend/professional/industry-standard" found.

## Adjustments applied during self-review

None — all checks passed without doc edits.

## Open items deferred to Phase 7 (wiki-curator)

- Verify `wiki/backlog/caixa-aprovacao.md`, `wiki/concepts/error-ux.md`, `wiki/architecture/data-model.md` exist; if not, downgrade to factual statement or remove cross-link.
- Refresh `Last verified` stamps on: `wiki/decisions/0007-two-tier-authz.md` (cross-linked), `wiki/concepts/iso-segregation.md`, `wiki/concepts/authz-tiers.md`, `wiki/modules/iam-tech-debt.md` (T-003/T-004 referenced).
- Update `wiki/README.md` index: replace prior approval stub entry with new doc + tech-debt + backlog anchors.
