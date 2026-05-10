# Phase 6.75 — Self-review (`documents`)

**Date:** 2026-05-10
**Reviewer:** main agent (Opus 4.7), fresh-eyes pass after Phase 6.5 `[tally] PASS`.

## 1. Severity rubric application

Each Critical / Major row re-mapped to the rubric in `templates/tech-debt-register.md`.

| Item | Severity | Trigger fired | Verdict |
|---|---|---|---|
| T-001 RFC 9457 envelope | Major | "Documented contract not followed by this module yet — measurable user impact via tooling that depends on the contract" | OK — Major (no audit-trail gap, no auth bypass) |
| T-002 OpenAPI spec drift | Critical | "Contract violation that downstream consumers rely on (e.g. error shape, idempotency replay window)" + applies to regulated paths | OK — Critical |
| T-003 documents-table defense-in-depth gap | Major | "Defense-in-depth gap: only one layer protects a mutation that the spec calls for multiple layers on" | OK — Major (no bypass exploit known; tier-1 still gates) |
| T-005 rename audit outside tx | Major | "Defense-in-depth gap" + atomic-pair broken; **not** Critical because audit IS written (just not atomically) — fails the consistency goal, not the trail-existence goal | OK — Major. If audit were skipped, this would be Critical. |
| T-006 finalize idempotency | Major | "Documented contract not followed yet (idempotency)" + partial unique index mitigates the worst outcome | OK — Major |
| T-009 placeholder_values FK | Major | "Defense-in-depth gap" (FK is a referential-integrity layer); latent today | OK — Major |

No row was upgraded or downgraded during this pass.

## 2. Mermaid box ↔ prose

Walked each C4 box.

- §3 Context: every box (`docs`, `registry`, `templates`, `iam`, `render`, `idemp`, `objstore`, `db`, `user`) is named in §3.2 Technical Context or §3.1 prose. **OK.**
- §5.1 Container: every box (`http`, `approvalHttp`, `app`, `domain`, `repo`, `jobs`, `api`, `db`, `iam`, `registry`, `templates`, `render`, `idemp`) named in §5.2 Public surface or §8 prose. **OK.**
- §6 sequence diagrams: every participant (`C`, `H`, `S`, `R`, `A`, `DB`, `SS`, `AR`, `EM`) is the natural HTTP/service shorthand and is explained by the surrounding step text. **OK.**

## 3. Top-3 in §11

Ordered by severity then blast-radius:
1. T-002 (Critical, surface = all /api/v2/documents/* routes, blocks RFC 9457 migration)
2. T-003 (Major, surface = every documents-table mutation, defense-in-depth)
3. T-005 (Major, surface = rename code path, atomicity broken on QMS rename audit)

Authorship-order would have led with T-001 (RFC 9457). Re-ordered correctly. **OK.**

## 4. Cross-link existence

Sampled all wiki cross-links in the composed doc:

- `wiki/decisions/0001-eigenpal-adoption.md` — exists
- `wiki/decisions/0007-two-tier-authz.md` — exists
- `wiki/decisions/0011-cd-atomic-create.md` — exists
- `wiki/decisions/0012-contract-first-api.md` — exists
- `wiki/concepts/placeholders.md` — exists
- `wiki/concepts/token-syntax.md` — exists
- `wiki/modules/iam-tech-debt.md` — exists
- `wiki/backlog/contract-first-followups.md` — exists
- `wiki/backlog/documents-refactor.md` — created this run
- `wiki/modules/documents-tech-debt.md` — created this run
- `wiki/modules/documents-v2.md` — exists, retired by R-100 in Phase 7
- `wiki/architecture/frontend-structure.md` — exists (per `wiki/README.md` index)
- `wiki/architecture/persistence.md` — exists (per `wiki/README.md` index)
- `wiki/architecture/api-contract.md` — referenced indirectly via ADR 0012; not directly linked. **OK.**

All concrete `wiki/**` paths in the doc resolve. **OK.**

## 5. Key Files freshness

Sampled three anchors during this pass (live code reads):

- `internal/modules/documents/delivery/http/handler.go:285` → `func (h *Handler) renameDocument` — confirmed.
- `internal/modules/documents/delivery/http/handler.go:869` → `func (h *Handler) authorizeDocumentScope` — confirmed, role gate at `:870`.
- `internal/modules/documents/approval/application/submit_service.go:85` → `authz.Require(ctx, tx, "doc.submit", areaCode)` — confirmed verbatim.

**Correction landed:** Codex artifact `02-flow-renameDocument.md` claimed handler returns 200 with `*domain.Document` after a `GetDocument` re-fetch. Actual code returns `204 No Content` (`handler.go:307`); no re-fetch. §6.2 sequence diagram and prose corrected during this self-review. **Also found** double `httpErr` call at `handler.go:303-304` — noted as latent bug in §6.2 prose, not promoted to its own tech-debt row (cosmetic; same status / message; visible only as "superfluous WriteHeader" log).

## 6. Backlog ↔ debt linkage

`tally_check.sh` already validated debt_id grammar and matched T-NNN ↔ backlog rows. Open WARN items:

- **T-007** (audit-domain decoupling, Minor, latent) — no backlog row by design (rubric allows latent Minor without row).
- **T-010** (legacy mux ↔ codegen drift, Minor) — explicitly gated on T-002 closure (R-002) per backlog notes; no row yet by design.

Both are documented as intentional in `documents-tech-debt.md`. **OK.**

## 7. Industry citations

§5 industry comparison artifact (`05-industry.md`) cites only existing rows in `references/industry-patterns-index.md`:
- IP-001, IP-002, IP-003, IP-004, IP-005, IP-006, IP-007, IP-008.

No new pattern added. No "industry standard X" sentence without a row. **OK.**

## 8. Subagent purity

Re-skim of Codex artifacts for "should / recommend / professional / industry-standard":

- `_artifacts/01-surface.md` — clean.
- `_artifacts/02-flow-listDocuments.md` — clean (uses `(unclear: …)` for missing spec ops, no prescriptive prose).
- `_artifacts/02-flow-renameDocument.md` — clean. Codex did make one **factual error** (claimed re-fetch + 200 response — see §5 above); fixed in composed doc, **not** the artifact (artifacts are research record, kept verbatim).
- `_artifacts/02-flow-finalizeDocument.md` — clean.
- `_artifacts/03-deps.md` — clean.
- `_artifacts/04-persistence.md` — Codex did use the word **"violation"** in its tripwire-pairing audit for `InsertInstance`/`InsertSignoff`/`UpdateInstanceStatus`. Main-agent verdict: these are not violations — the **caller** sets the GUC before the INSERT runs, so the tripwire passes; absence of `authz.Require` in the repo file is a layering observation. Reclassified verbally in composed doc §8.1 and tech-debt T-003 surface scoping (only `documents`-table mutations carry the gap; approval-table writes are fully gated by tier-2 at the caller). Not a "should" violation per se, but a borderline prescriptive word; logged here for skill changelog input.

---

## Result

Self-review complete. One material correction applied (§6.2 response shape). No re-run of Phase 6.5 needed: severity counts unchanged, ADR-link count unchanged, debt↔backlog linkage unchanged. Doc set is publish-ready.
