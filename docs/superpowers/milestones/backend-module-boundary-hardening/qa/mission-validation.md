# Mission Terminal Acceptance — Verdict

> Written by: the mission-validator subagent (separation of powers). Validates against: ../mission.md §8.
> Run: 2026-06-21 · Code HEAD: `44b83071` (mission tip; M0–M4 complete, local, not merged/pushed) · Verdict: see bottom.
> Re-audit artifact judged: `wiki/backend/_artifacts/architecture-re-audit-2026-06-21-post-boundary-hardening.md`.

This judge re-ran every deterministic §8 check itself at HEAD `44b83071` (did not trust milestone evidence
docs or the re-audit transcript) and independently spot-checked the re-audit's "0 remaining" claims by
re-grepping a sample of the §5 sites and re-running the named parity tests. Commands and real output below.

## Per-criterion results

| # | §8 criterion | Method run (command) | Real evidence | Pass? |
|---|--------------|----------------------|---------------|-------|
| 1 | ADR-0039 exists, Accepted, unambiguous definition + exemption list | `Read wiki/decisions/0039-cross-module-base-table-read-boundary.md` | Status line: **"Accepted 2026-06-20 (amended 2026-06-20 — D3(d)–(f) exemptions)"**. D1 one-sentence rule (raw read of another module's base table = violation); D2 base/view refinement; D3(a)–(f) exemption list; worked classification table of all §5 sites (0 unclassified). Mechanical, reviewer-decidable. | ✅ |
| 2 | cilint/CI H-G grep-guard green on full tree | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | **exit 0**. Source-verified `tools/cilint/internal/analyzers/hgcrossmodule.go`: `hgPendingRemediation` slice is **EMPTY** (only comments narrating ported sites remain); `hgExempt` holds **only** the permanent D3(d)–(f) carve-outs (X1–X8: audit_events / auth_identities / auth_sessions platform+auth reads, jobs watchdog approval_* worker-layer). | ✅ |
| 3 | Every §5 inventory item (rows 3–16) ported or ADR-0039-exempt | Direct `grep` of each cited site for its foreign base-table read | All deleted: item 4 (documents→controlled_documents) 0; items 5/6/7 (CD→documents/document_revisions/approval_instances) 0; items 8/9 (documents JOIN controlled_documents) 0; item 10 (documents→document_process_areas) 0; item 11 (iam→document_process_areas) 0 (comment-only mention); items 12/13/14 (→user_process_areas) replaced by `metaldocs.v_active_user_areas`; item 15 (search) reads only `v_document_search_facts`/`v_cd_search_facts`/`v_cd_grantee`; item 16/N1 (documents→templates_template_version) 0; item 3 resolution.go: **0 bare `"published"`/`"obsolete"` literals**. Views `0242`/`0243`/`0244` present. | ✅ |
| 4 | Re-audit grades module-boundaries = A and H-G = 0 both readings | Judge the artifact §2/§6 **and** independently re-run both readings | Artifact scorecard: dim 8 module-boundaries **B+→A**; code-quality **B+→A−**; no dimension below post-M9 floor. Independent re-run — **canonical §6 greps:** document_profiles outside taxonomy = 0; iam_users/iam_user_roles `FROM` outside /iam/ = 0. **Broad sweep** over all 8 debt tokens: every remaining `FROM/JOIN` hit is same-module (owner reading own table, D4c) **except** `jobs/stuck_instance_watchdog/job.go:147` (approval_instances), which is the recorded D3(f) X8 allowlist site. **0 violations outside the allowlist** under both readings. | ✅ |
| 5 | `go build ./...` + `go test ./...` green; integration green where docker PG available, else explicit not-run | `go build ./...`; `go test ./...`; targeted `-tags integration` parity tests | `go build ./...` **exit 0**; `go test ./...` (unit) **exit 0** (no FAIL). testdb-clone integration packages green. Named mission parity tests independently re-run and **PASS**: `TestActiveInstanceReader_ParityWithRawGetActiveInstance` (4 scopes), `TestCDFieldReader_ProfileCode/ProcessAreaCode_ParityWithRawSQL`, `TestCDVisibilityContract_ComposedDecisionParityWithRaw`, `TestCanRead_ViewParityWithRaw`, `TestList_ViewParityWithRaw`. Bounded not-run `_Live` class → see statement below. | ✅ (with bounded not-run) |
| 6 | 0 new Critical/Major, each skeptic-confirmed | Judge artifact §3/§4 | Artifact: 0 Critical/Major proposed by any of the 10 dimension readers; refute-by-default skeptic stage had nothing to refute; verified count **0/0**. Only minor comment/cosmetic observations, none action-required. No criterion in this judge's independent checks surfaced a new Critical/Major introduced by the remediation. | ✅ |

## Bounded not-run — raw-base-DSN `_Live` integration tests (HS-3, NOT a mission regression)

I independently reproduced this class at HEAD. `*_Live` / raw-base-DSN tests error at **seed/setup** with
`relation "metaldocs.<table>" does not exist (SQLSTATE 42P01)` — observed live:
`TestSequenceAllocatorNextAndIncrement_Concurrent` (`insert profile: relation "metaldocs.document_profiles" does not exist`)
and `TestPostgresLimiter_Live` (`relation "public.auth_failure_counters" does not exist`), same class as
`TestSecurityRepository_PortParity_Live` and `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema`. Cause is
**environmental**: the `:5434` test container's base `metaldocs` DB is intentionally empty — schema
materializes only inside per-test testdb-framework clones. These are **setup-stage `relation does not exist`
failures, not parity mismatches and not boundary regressions**. The M4 milestone-validator independently
reproduced this same class at the pre-M4 worktree (`e147f33e^`) and dispositioned it pre-existing
(milestone-4 `qa/milestone-qa.md` C4). The mission's actual seam-parity proofs run under testdb clones and
passed (criterion 5). Recorded **not-run (HS-3), no false green** — does not fail the mission.

## Pass bar
- Bar (§8): "in a fresh re-run of the F5.1 10-dimension architecture re-audit, **module-boundaries / DDD = A**
  … and **H-G = 0 under BOTH readings** — the canonical §6 greps AND the broad 'any cross-module owned-**base**-table
  read' reading as defined by ADR-0039 — with **0 skeptic-confirmed new Critical/Major** introduced by the
  remediation, and no regression in any other dimension (all remain ≥ their post-M9 grade)."
- Met? **Yes.** Deciding evidence: dim 8 = A (artifact §2, corroborated by my independent greps showing 0
  cross-module base-table reads outside the D3(f) allowlist site under both readings); cilint exit 0 with an
  empty remediation ledger; 0/0 Critical/Major; code-quality recovered B+→A− and no dimension below its
  post-M9 floor.

## Forbidden-list (any hit = FAIL)
- [ ] Fixture/mock passed off as real-provider proof — NOT hit: parity tests run against real Postgres testdb clones; the not-run `_Live` class is explicitly recorded, not false-greened.
- [ ] A criterion marked pass without a command actually run — NOT hit: every criterion backed by a command run this session with real output.
- [ ] Split-brain / guessed contract surfaced in the aggregate diff — NOT hit: consumers read owner-published views/ports (`v_active_user_areas`, `v_cd_*`, `v_document_search_facts`, CDFieldReader/ActiveInstanceReader/AreaCatalogReader/TemplateVersionPort); ADR-0039 + cilint guard make the contract mechanical.
- [ ] Self-judged / validator edited or fixed code — NOT hit: this judge wrote only this verdict file; no source/test/spec touched.

## Verdict
- VERDICT: PASS
- On PASS — handed back to the main session for the operator's final Grade-A sign-off on the parent
  `grade-a-completion` (HELD pending this PASS) + §12 program close-out. This judge does **not** flip mission,
  program, or roadmap status, and does not declare the mission done — that is the operator's action.
