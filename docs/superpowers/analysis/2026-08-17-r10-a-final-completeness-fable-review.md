# R10-A Final Completeness — Independent Mechanical Sweep (Fable)

> **Status:** REVIEW EVIDENCE — INDEPENDENT COMPLETENESS VERDICT / AWAITING OPERATOR ADJUDICATION
> **Independent verdict:** `APPROVE R10-A COMPLETENESS CLOSURE WITH MATERIAL FIXES`
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Artifact under proof:** `docs/superpowers/analysis/2026-08-17-r10-a-final-completeness-correction.md` @ `5cb350d5`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Authority note:** this file is review evidence only. It adjudicates nothing, promotes nothing, and does not amend R9.5 or open R10-B.
> **Implementation gate:** **CLOSED — nothing implemented or proposed for implementation here.**

## 0. Scope and posture

This review executes only the narrow mechanical proof gate defined in the artifact's §9. It is **not** a third broad architecture review; the 8+3 topology already survived two independent adversarial passes and was not re-litigated. Frozen authority was reconstructed via the `AGENTS.md` read order at remote HEAD `5cb350d5` (fetched and verified; the only delta since the cold delta review is the completeness-correction artifact itself — single commit, single file). Authority docs verified unchanged across the whole R10-A review chain: `git diff --stat f51f6bfa..HEAD -- wiki/ docs/engineering/ AGENTS.md` is empty.

## 1. Verdict summary

```text
VERDICT: APPROVE R10-A COMPLETENESS CLOSURE WITH MATERIAL FIXES

BLOCKER findings: none
MAJOR findings:   FC-F1 (Document owner/responsibility relationship has no §4 fact row)
LOW findings:     FC-F2 (tenant settings subsumption undeclared)
RetentionPolicy entity test: PASS — no such V1 entity introduced
Duplicate owners: none found
Invented facts:   none found
CD-F2 / CD-F3 / CD-F4 corrections: VERIFIED CLOSED
Surface supersession: VERIFIED
R9.5 reopen set:  EMPTY
```

Both findings are mechanical inventory additions/clarifications. Neither questions the topology, an owner boundary, a seam, or any frozen semantic.

## 2. Sweep method and coverage

Every frozen durable/business fact family was enumerated directly from the active ledger (`docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`) section by section and matched against §4 of the artifact under proof:

| Frozen source | §4 target | Result |
|---|---|---|
| Ledger §1 AuthN/Org/AuthZ + 29+16 permission catalog + role bundles | §4.1–§4.3 | covered, except FC-F1/FC-F2 (below) |
| Ledger §2 Approval V1 (policy/steps/instance/participants/decision/attestation/SoD/withdraw-cancel-reassign) | §4.5 | **fully covered — CD-F1 verified closed** |
| Ledger §3 CI foundations (DocumentType, Document, Revision, Submission, numbering, Template/TemplateUse/TemplateSpec, reason-for-change) | §4.4 | covered |
| Ledger §4 Periodic Review / Rendition / Release / OfficialRepresentationPolicy | §4.4 | covered |
| Ledger §5 Distribution / Tenant Dictionary / System Value Catalog / Audit / Notifications / Search | §4.4, §4.8, §4.10, §4.12 | covered, incl. value-snapshot facts |
| Ledger §6 tenant lifecycle / deletion request / erasure / DEK custody / tombstones / restore / PlatformOperator-SystemPrincipal | §4.2, §4.7, §4.12 | covered |
| Ledger §8 R9.5-1 Artifact / ContentFormat / Evidence / EvidenceType / external-world cancellation / primary-Artifact relationships | §4.6, §4.9 | covered |
| Ledger §9 R9.5-2 storage/staging/relocation/managed keys/restore | §4.9, §4.12 | covered (providers correctly mechanism-classified) |
| Ledger §10 R9.5-3 WorkingContent / OCC / WorkingSnapshot / EditorSession / EditorialComment | §4.4 | covered |
| Ledger §11 R9.5-4 Dossier / DossierType / ExternalReference / links / ACTIVE↔ARCHIVED | §4.6 | covered, incl. Dossier lifecycle |
| Ledger §12 R9.5-5 retention rule vocabulary / bindings / anchors / extension / holds / materialization / disposition | §4.4, §4.6, §4.7 | covered; rule-value vs meaning split clean |
| Ledger §13 R9.5-6 migration / imported truth / export contracts / manifests | §4.4, §4.7, §4.11 | covered; imported-fact rule unambiguous |
| Ledger §14 R9.5-7 attestation evidence / basic content safety / manifestation renditions | §4.5, §4.9, §4.4 cross-owner manifestations | covered |
| Ledger §15–§16 non-goals / R9.5-8 refinements / permission delta | §4 overall | no invented capability; no excluded concept resurrected |

The eight author-sweep additions listed in the artifact's §8 (Dossier lifecycle, external-world cancellation, primary-Artifact relationships, type-owned retention-rule values, RetentionPolicy removal, PlatformOperator/SystemPrincipal, confirmation-stage validation facts, imported-history split) were each independently traced to a frozen ledger anchor. **None is invented; all are frozen facts that were previously implicit.**

## 3. Findings

### FC-F1 — MAJOR — the Document owner/responsibility relationship has no §4 fact row

**Claim.** The frozen R9 permission catalog contains `document.owner.manage` (active ledger, §1 catalog, line 117). A permission to manage a document's owner entails a durable business fact family — the Document owner/responsibility relationship — which must be classifiable by the normative inventory. §4.4's identity row enumerates "Document stable identity / code / type / Area relationship" (artifact line 158) and omits the owner relationship. The artifact's own §4.3 predicate table already names "Document/Revision **ownership**, Area and lifecycle relationships" as a Controlled Information relationship class (artifact line 141) — so the relationship is acknowledged on the predicate side while missing on the fact side. By the inventory's own completeness rule, this durable fact family is currently unclassifiable.

**Receipts.** Ledger line 117 (`document.owner.manage` frozen permission); artifact line 141 (predicate class acknowledges ownership) vs line 158 (fact row omits it).

**Adjudication classification.** Missing owner row — the exact defect class this gate exists to catch. Not a topology question: the owner is unambiguously Controlled Information.

**Required change.** Add one §4.4 row: Document owner/responsibility relationship → Controlled Information (boundary: an owner relationship never bypasses capability-based authorization; `document.owner.manage` governs changes).

**R9.5 reopen?** NO.

**Proof to close.** Row present; re-run of this sweep's §1-catalog cross-check finds every permission's object fact family classifiable.

### FC-F2 — LOW — tenant settings subsumption is undeclared

**Claim.** The frozen catalog contains `tenant.settings.manage` (ledger §1, line 99), implying durable tenant-scoped settings/configuration facts. §4.2's "Tenant" row plausibly subsumes them, but nothing says so; every other manage-permission object has an explicit family (DocumentType, EvidenceType, DossierType, ApprovalPolicy, TemplateUse, dictionary values, distribution, holds, etc.), making Tenant settings the one silent case.

**Required change.** Either state in the §4.2 Tenant row's boundary that tenant settings/configuration facts are part of the Tenant family, or add an explicit row. One line either way.

**R9.5 reopen?** NO.

## 4. Specific mandated tests

### 4.1 No V1 `RetentionPolicy` entity — **PASS**

Verified in all three places retention configuration appears: §4.4 DocumentType row ("direct frozen value, not a `RetentionPolicy` entity"), §4.6 EvidenceType row (same), §4.7 header ("retention-rule semantics, not a separately versioned `RetentionPolicy` aggregate"). This matches frozen authority: the ledger (§12) defines rule values chosen directly by DocumentType/EvidenceType (`NoMinimum | KeepFor | Indefinite`, anchor `CAPTURED_AT | OCCURRED_AT`) and snapshots them into RetentionBinding; no standalone policy aggregate is frozen anywhere. The subtraction is a correct YAGNI application, not an R9.5 reopen — no frozen semantic is weakened, and the reopen trigger (real independent lifecycle/reuse requirement) is properly stated.

### 4.2 Duplicate-owner sweep — **NONE FOUND**

Adversarially checked every fact that appears in two tables: primary-Artifact relationships (CI owns the Revision-side relationship, Documentary Context the Evidence-side, Artifact the byte identity — three distinct facts); retention-rule configuration (type owners own configured values, Records Governance owns vocabulary/meaning/bindings — reference-vs-referent split, declared on both sides); ContentFormat (Artifact sole authority; Documentary Context/CI rows reference only); fresh-auth (Authentication owns assurance state, Approval owns the Step requirement — declared both sides); tombstones (Organization) vs backup transport (platform); session truth (Authentication) vs erasure coordination; imported facts (single owner per fact via the §4.11 rule). No fact family has two independent writers/authorities.

### 4.3 Invented-entity sweep — **NONE FOUND**

Every §4 family traces to a frozen ledger anchor; no row introduces an entity, capability, or lifecycle absent from R3–R9.5. Explicitly re-checked the newest additions (confirmation-stage validation facts → ledger §14 basic content safety; external-world cancellation → ledger §8; EditorSession → ledger §10; PlatformOperator/SystemPrincipal → ledger §6). No excluded non-goal (§15) is resurrected.

### 4.4 CD-F2 / CD-F3 / CD-F4 closures — **VERIFIED, no new cycle or second authority**

- **CD-F2:** the §5 manifestation seam is correctly generalized (any cross-owner manifestation via published read/composition contract). Cycle check: Approval → CI is reference/read; the manifestation flow Approval-facts → published seam → CI representation adds a read-contract in the reverse direction, resolvable by composition/interface inversion without semantic cycle — explicitly constrained as ownership-only with shape deferred to R10-B. No authority moves: CI owns the representation, Approval the native fact.
- **CD-F3:** §6 supersedes the stale packet mapping; `tokens/value administration → Controlled Information`, `notifications → Notifications support`, plus coherent splits for `security` and `render`. No `Dictionary` or `notifications projection` classification survives.
- **CD-F4:** the §4.11 imported-fact ownership rule assigns historical Revision-attached governance evidence to Controlled Information by attachment object, with the never-fabricate-native rule intact and Interchange retaining only process truth. Unambiguous.

### 4.5 Surface supersession — **VERIFIED**

§1 supersession list + §6 classification remove every stale row; §6 correctly limits itself to owner classification and defers paths/DTOs/journeys to R10-E.

## 5. R9.5 reopen set

**EMPTY.** Neither finding, nor the RetentionPolicy subtraction, invalidates any frozen invariant/authority; nothing satisfies (or needs) the five-part reopen contract.

## 6. Final verdict

```text
VERDICT: APPROVE R10-A COMPLETENESS CLOSURE WITH MATERIAL FIXES
```

Required to close: **FC-F1** (one §4.4 row for the Document owner/responsibility relationship). **FC-F2** closes with a one-line subsumption note. Both are mechanical inventory edits; no topology, seam, boundary, or frozen semantic is touched. After these edits, this sweep found the inventory complete: every remaining frozen durable/business fact family has exactly one owner, no duplicates, no inventions, and the closure instrument finally demonstrates what it asserts.

**Residual method note for adjudication:** the permission catalog proved to be the most productive completeness oracle in this sweep (both findings derive from it). The promotion step may want to record the cross-check "every frozen permission's object maps to a classifiable fact family" as part of the R10-A closure evidence, so R10-B inherits it as an invariant rather than a review habit.
