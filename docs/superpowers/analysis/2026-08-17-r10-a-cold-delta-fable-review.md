# R10-A Corrected Target — Cold Delta + Global Coherence Review (Fable)

> **Status:** REVIEW EVIDENCE — INDEPENDENT COLD DELTA VERDICT / AWAITING OPERATOR ADJUDICATION
> **Independent verdict:** `APPROVE R10-A CORRECTED TARGET WITH MATERIAL FIXES`
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Artifact under challenge:** `docs/superpowers/analysis/2026-08-17-r10-a-fable-adjudication-corrected-target.md` @ `74c1ba80`
> **Prior independent review (evidence only):** `docs/superpowers/analysis/2026-08-17-r10-a-independent-fable-review.md` @ `c0bde261`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Authority note:** this file is review evidence only. It adjudicates nothing, promotes nothing, and does not amend R9.5 or open R10-B.
> **Implementation gate:** **CLOSED — no product implementation performed or proposed here.**

## 0. Cold posture and evidence statement

State was reconstructed from the repository at remote HEAD `74c1ba80` (fetched and verified before review). The `AGENTS.md` chain, method mirror, handoff, redesign authority page, active ledger and R9.5-8 freeze evidence were re-verified: the only change since the prior review base is the adjudication/corrected-target artifact itself (single commit, single file — confirmed by diff). The corrected target was treated as hypothesis, not authority. The prior independent review's findings were **not** presumed correct: each adjudicated correction was re-derived from frozen semantics before its coherence was judged, and one prior-review-aligned claim was actively tested against current-state evidence and **falsified in the corrected target's favor** (§3, receipt R2). Current runtime/code was consulted only as evidence to falsify or support specific claims, cited per finding.

## 1. Verdict summary

```text
VERDICT: APPROVE R10-A CORRECTED TARGET WITH MATERIAL FIXES

BLOCKER findings: none
MAJOR findings:   CD-F1 (normative §4 inventory omits the Approval fact family)
LOW findings:     CD-F2, CD-F3, CD-F4
R9.5 reopen set:  EMPTY
```

The corrected 8+3 topology, the seven finding adjudications, the predicate-ownership rule, the key-custody split, the Dictionary deletion, the ContentFormat assignment, the Notifications reclassification and the explicit seam/DAG set are structurally sound and mutually coherent. The single material defect is that the §4 inventory — the artifact's own instrument for closing F3's defect class — fails its own completeness rule for one entire business bounded context.

## 2. Proof-obligation results (§9.1–§9.11)

### §9.1 — any frozen durable/business fact still without exactly one owner? — **FALSIFIED: YES (CD-F1)**

The §4 inventory is declared "normative for the corrected candidate" with a completeness rule making any durable fact "not classifiable by this inventory … a material R10-A contradiction". Sweeping the frozen ledger against the §4 table: **the entire Approval fact family has no row.** Missing fact families:

```text
ApprovalPolicy(version) + ordered ApprovalStep configuration (actor rule, ANY|ALL, requires_reauthentication, due_in_days)
ApprovalInstance + binding to exact RevisionSubmission
activated participant snapshot
ApprovalDecision + attestation evidence (actor, Step, policy version, decision, trusted server time, AuthN assurance/fresh-auth)
reassignment / cancellation / oversight facts
```

These are frozen durable facts (active ledger §2, §14.1–2) and include the compliance-critical attestation record. §3.1-E assigns them narratively, so this is not a topology defect — but §4 is the normative closure instrument, and by its own rule ApprovalDecision is currently "a material R10-A contradiction". Every other owner in §3.1/§3.2 has §4 rows; only Approval was skipped. See CD-F1.

Also caught by the same sweep, minor: Dossier `ACTIVE ↔ ARCHIVED` lifecycle state (ledger §11) is at best implicit in the "Dossier identity/stable key/scope" row — the fix sweep should make it explicit (folded into CD-F1's required re-sweep, not a separate finding).

### §9.2 — any fact assigned to two authorities? — **NO**

Attacked pairs: ContentFormat (Artifact) vs System Value Catalog (CI) — disjoint by definition and by the §4 boundary notes; retention policy reference (type owner) vs policy meaning/bindings (Records Governance) — clean FK-vs-referent split; Evidence naming policy (Documentary Context) vs ContentFormat catalog (Artifact) — explicit boundary note; object-level provenance (target owner) vs transfer provenance (Interchange) — explicit split rule; key-custody lifecycle facts (Organization) vs crypto mechanics (platform) — lifecycle-vs-mechanism, no shared fact. No dual authority found.

### §9.3 — falsify Organization key-custody ownership without promoting crypto to authority? — **COULD NOT FALSIFY**

Re-derived independently: the frozen erasure/restore semantics (ledger §6) make DEK preservation/destruction a component of the Tenant lifecycle transition Organization already owns; Records Governance supplies lawfulness inputs without owning key state; wrap/unwrap/KEK remain mechanism. Alternatives re-tested: a dedicated `support/tenantcrypto` owner adds a module for one fact family with one writer (accidental structure); Records Governance ownership splits one fact's lifecycle across two authorities. The chosen split survives Structural Inversion (erasure semantics alone force key lifecycle facts onto the tenant lifecycle owner, regardless of the current `security` module). Current-state receipt that the fact family is real and transactional: `internal/modules/security/domain/tenant_crypto.go:10-15`.

### §9.4 — domain relationship semantics Authorization would still need centrally? — **NONE FOUND**

Stress-tested against the hardest frozen surfaces: cross-scope search (Organization membership + CI area/lifecycle predicates composed); export completeness fail-closed (Documentary Context link predicates + per-target predicates — representable without centralization); approver read of the exact Submission under decision (`approval.act` + Approval-owned participant-snapshot relationship, consuming a CI read contract — fits the §3.1-C table row); Distribution audiences (never grant). The predicate-ownership table covers every frozen relationship class encountered; no predicate must migrate into Authorization.

### §9.5 — frozen second consumer requiring standalone Dictionary now? — **NONE FOUND**

Independent sweep of frozen semantics for value consumers: REV-creation snapshot (CI, ledger §5); Evidence naming tokens `{TYPE}/{DOSSIER}/{REF}/{SEQ}` are EvidenceType naming policy (Documentary Context, ledger §8) — not Dictionary; numbering literals are CI. No frozen consumer outside the governed authoring/revision lifecycle exists. The RESTRUCTURE NOW deletion is supported; the reopen trigger (real second consumer / independent lifecycle) is correctly stated.

### §9.6 — Notifications move creating or hiding business authority? — **NO**

Delivery/attempt/inbox/read state is durable but non-business; acknowledgement, approval and distribution truth remain with their owners; producers resolve recipients. `support/` placement now truthfully signals non-rebuildable-but-non-authoritative, and `projections/` regains its invariant (everything under it is rebuildable). Coherent.

### §9.7 — unavoidable semantic/package cycle? — **NONE**

Strongest candidates re-attacked: Organization ⇄ Authentication (erasure-time session revocation is composition-mediated); Organization ⇄ Authorization (every delivery surface consumes Authorization evaluation — a universal request-lifecycle mechanism edge, resolvable by interface inversion; Authorization's semantic dependency on Organization subjects/scopes stays one-way); CI ⇄ Approval (both directions composition/read-contract mediated per seam 1); CI ⇄ Records Governance (submission-time binding via composition; disposition reads CI facts one-way); Artifact confirmation (opaque owner reference, no back-import); Interchange (calls owners, never depended on); Audit (append-only inbound seam). Acyclic with the declared seams. One concrete data-flow deserves naming under seam 1 — see CD-F2.

### §9.8 — transactional-composability constraint: prejudges R10-B or too weak? — **NEITHER**

The constraint requires seams to *permit* one local DB transaction where frozen atomicity demands it (same-commit audit append, coherent Submission creation, atomic Release). That is forced by frozen invariants plus the modular-monolith premise, not a mechanism choice; UnitOfWork/port/DB shape remains genuinely open to R10-B. It is also not too weak: every frozen atomic case in §5.1's flows is expressible under it. Balanced as stated.

### §9.9 — provenance duplication between Interchange and target owners? — **NONE FOUND**

Per-item migration outcomes / reconciliation reports (process truth) vs object source provenance (target owner) vs ExternalReference (Documentary Context business correlation) are three distinct fact classes with no shared writer. One normative row is ambiguously worded — see CD-F4.

### §9.10 — Structural Inversion + subtractive/YAGNI on the final 8+3 — **PASS**

Re-run independently, not inherited: every business BC and Artifact/Audit/Interchange re-derive from frozen semantics under inversion of the current tree. Subtractive attempts: Notifications→platform (rejected — durable product-facing inbox state is attributed, not commodity); Interchange split (rejected — duplicates process-truth ownership or leaves thin wrappers over one shared mechanism); Audit→platform (rejected — timeline/export/retention-regime meaning needs one owner); Artifact collapse (rejected — wrongful cross-domain content ownership); Authentication/Organization/Authorization merges (rejected — frozen separations + IdP seam); Distribution fold (rejected — distinct compliance facts). No owner exists for an imagined future consumer — the one that did (Dictionary) was deleted by this correction. Watch item, not a finding: Controlled Information now concentrates ~19 of ~45 inventory fact families; this is essential complexity (one governed lifecycle authority) but makes CI's negative-boundary list (§3.1-D) the load-bearing God-context guard for R10-B.

### §9.11 — R9.5 reopen set — **EMPTY**

No finding invalidates a frozen invariant/authority; none satisfies (or needs) the five-part reopen contract. CD-F1 is an inventory-closure defect against R10-A's own rule, not an R9.5 semantic gap.

## 3. Global coherence of the corrections — receipts

**R1 — corrections are mutually consistent.** Key-custody (F1 fix) integrates with erasure/restore flows (§5.1) and the `security` legacy row (§7); predicate rule (F2 fix) integrates with Search/export seams (§5.2-8) and the §3.1-C table; Dictionary deletion (F4 fix) integrates with §3.1-D internals, §6 layout, §7 `tokens`/`render` rows and §8's subtractive statement; provenance split (F5 fix), Notifications reclass (F6 fix) and the Audit seam (F7 fix) each appear consistently in §3, §4, §5 and §6. No correction contradicts another.

**R2 — one adversarial probe falsified in the corrected target's favor.** The §7 `render` row's claim that "value-catalog semantics" currently live in `render` looked suspect (prior review had located dictionary values in `tokens`). Evidence: `internal/modules/render/domain/computed_catalog.go:1-5` — render's `ComputedCatalog` is documented as "the SINGLE SOURCE OF TRUTH for computed (system) tokens" (ADR 0050). The row is accurate current-state mapping; the System Value Catalog's current home is `render`, and its target re-home to Controlled Information is coherently stated in both §3.1-D and §7.

**R3 — the same evidence exposes CD-F2.** The current computed catalog includes `approvers` and `approval_date` (`computed_catalog.go:26-28`) — Approval-sourced values. At target level this is frozen-endorsed: approval/effectivity may be manifested in a human-readable Rendition (ledger §14.5). See CD-F2.

## 4. Material findings

### CD-F1 — MAJOR — the normative §4 inventory omits the entire Approval fact family

**Claim.** §4 presents itself as the normative exactly-one-owner closure inventory with a completeness rule declaring unclassifiable durable facts a material contradiction. ApprovalPolicy/Step configuration, ApprovalInstance, activated participant snapshots, ApprovalDecision with its attestation evidence, and reassignment/cancellation/oversight facts — all frozen durable facts (ledger §2, §14.1–2) — have no row. The instrument created to CLOSE F3's defect class ("full fact-to-owner inventory … not a patch for only the examples found by the reviewer", §2 F3 adjudication) reproduces the defect for one whole bounded context.

**Root cause.** The inventory was built delta-first — rows generated from the corrections and the prior review's named examples — instead of from a full frozen-ledger sweep. Same root cause as original F3 (assertion where demonstration is required), recurring at smaller scale.

**Invariant affected.** §9.1 owner completeness; §4's own completeness rule.

**Global Maximum / YAGNI.** No structural change needed; the topology (§3.1-E) already assigns these facts. The fix is mechanical: add the Approval rows (boundary notes: binds exact `RevisionSubmission`; never owns effectivity; attestation evidence is Approval authority, manifestation Renditions are CI representations of it) and make Dossier `ACTIVE↔ARCHIVED` explicit.

**Required change.** Add the missing rows, then rerun one independent completeness sweep of the full frozen ledger against §4 — the closing proof F3's adjudication promised.

**R9.5 reopen?** NO.

**Proof to close.** Post-fix sweep finds zero frozen durable fact families without a §4 row.

### CD-F2 — LOW — Approval-sourced computed values are an undeclared concrete data-flow under seam 1

**Claim.** The System Value Catalog (now CI-owned, §3.1-D) includes Approval-sourced values — current evidence `internal/modules/render/domain/computed_catalog.go:26-28` (`approvers`, `approval_date`); frozen target evidence ledger §14.5 (approval/effectivity manifestation Renditions). Producing such a Rendition requires Approval evidence as input. Seam 1 ("CI ↔ Approval … composition-mediated and/or reference/read-contract based") generically covers this, but the concrete flow is nowhere named, and it is exactly the kind of edge R10-B could naively wire as a CI→Approval import.

**Required change.** One sentence under seam 1 or §3.1-D: manifestation Renditions and Approval-sourced computed values receive Approval evidence via composition/read contract; CI never imports Approval to resolve them; approval facts in a Rendition are representations of Approval authority, never a second record of it.

**R9.5 reopen?** NO.

### CD-F3 — LOW — the candidate packet's §10 surface classification is now partially stale with no supersession note

**Claim.** The original packet's OpenAPI/tag direction (`tokens → dictionary`; `notifications → notifications projection`) predates the Dictionary deletion and Notifications reclassification. The corrected target never states that its §3/§6/§7 supersede the packet's §10 rows. Two live staging artifacts now answer the same classification question differently — a stale-second-authority seed for R10-E.

**Required change.** One supersession line in the corrected target (or its promotion commit): surface classification follows the corrected topology; packet §10 rows are historical.

**R9.5 reopen?** NO.

### CD-F4 — LOW — "target semantic owner of the imported object/fact" is ambiguous for imported approval-history evidence

**Claim.** The §4 row for historical imported object-level facts can be read two ways for imported approval-history evidence: owner = Approval (the fact's *kind*) or owner = Controlled Information (the *object the evidence attaches to*). Frozen semantics require the latter — imported governance evidence is revision-attached history, never native ApprovalDecision (ledger §13).

**Required change.** Reword to "the semantic owner of the object the imported fact attaches to", with the existing never-fabricated-as-native note.

**R9.5 reopen?** NO.

## 5. Final verdict

```text
VERDICT: APPROVE R10-A CORRECTED TARGET WITH MATERIAL FIXES
```

Required before closure/promotion: **CD-F1** (add Approval fact-family rows + one independent full-sweep proof). CD-F2–CD-F4 are LOW and close with one-line edits at adjudication.

**Single biggest remaining structural risk:** inventory-closure discipline, not topology. The 8+3 ownership cut is sound and survived re-derivation; what failed twice now (F3, then CD-F1) is the *demonstration* step — building the completeness proof from deltas instead of from the full frozen surface. If promotion accepts §4 without the independent full sweep, R10-B inherits a normative inventory that silently classifies an entire bounded context's facts "by narrative", and the completeness rule fires as a contradiction mid-stage instead of closing here.
