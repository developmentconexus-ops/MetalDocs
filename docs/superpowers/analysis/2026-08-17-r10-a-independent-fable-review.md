# R10-A Ownership Topology — Independent Fable Adversarial Review

> **Status:** REVIEW EVIDENCE — INDEPENDENT VERDICT / AWAITING OPERATOR ADJUDICATION
> **Independent verdict:** `APPROVE R10-A WITH MATERIAL FIXES`
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Packet under challenge:** `docs/superpowers/analysis/2026-08-17-r10-a-ownership-topology-fable-review-request.md` @ `f51f6bfa`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Authority note:** this file is review evidence only. It does not amend R9.5 or R10 authority, and it adjudicates nothing. Accepted outcomes must be adjudicated separately against the Method and promoted deliberately.
> **Implementation gate:** **CLOSED — no product implementation performed or proposed here.**

## 0. Independence and evidence statement

State was reconstructed exclusively from the repository at `f51f6bfa`, following the `AGENTS.md` read order: method mirror → current handoff → redesign authority page → active ledger → R9.5-8 freeze evidence → review packet. Prior conversation memory was not used as authority. Current runtime/code was consulted only as migration/current-state evidence to falsify or support specific claims (cited below per finding); no conclusion in this review depends on current code shape. The Structural Inversion Test was applied to every candidate owner: each CONFIRM below was checked to still follow if the current 15 modules were shaped oppositely.

## 1. Authority reconstruction — SUCCESS

- `AGENTS.md` routing intact; method mirror `docs/engineering/standards/root-cause-global-maximum-method.md` present at v1.0.0.
- Handoff, redesign page and active ledger agree: R3–R9 LOCKED, R9.5-1…7 LOCKED, R9.5-8 CLOSED/APPROVED, R9.5 FROZEN, R10 NEXT/DESIGN ONLY, implementation BLOCKED.
- R9.5-8 freeze evidence consistent with ledger §16 (reopen set EMPTY, verdict `APPROVE / FREEZE R9.5`, operator-ratified 2026-08-17).
- No material contradiction found between authorities before judging the candidate. The packet's stage-gate table (§1) matches the handoff exactly.

The candidate was treated as hypothesis, not authority, per packet §0.

## 2. Material findings

### F1 — MAJOR — Tenant key custody (DEK) is an ownerless frozen durable fact

**Claim.** Frozen erasure semantics require durable per-tenant DEK facts with a lawfulness-gated lifecycle: "destroy no-longer-needed Tenant DEK"; "A DEK needed to preserve legally retained intelligible content is not destroyed while that obligation remains"; backup/restore must reconcile these facts (active ledger §6). The candidate assigns this fact class to no owner: it appears in the tenant-erasure composition ("DEK destruction only when lawful", packet §6) but is absent from every owned-facts list in packet §4.1/§4.2. The legacy disposition row for `security` (packet §9) routes "pure operational security to platform/supporting mechanics", which would silently demote a governance-coupled durable business fact to commodity mechanics — exactly the "commodity machinery must not become authority by convenience / differentiated semantics must not become commodity by accident" failure the Method names.

**Evidence anchors.** Active ledger §6 (erasure flow, DEK preservation rule); packet §4.1-B (Organization owned list — no key facts), §6 (erasure composition), §9 (`security` row). Current-state evidence (migration evidence only, not authority): `internal/modules/security/domain/tenant_crypto.go:10-15` — per-tenant DEK wrapped under deployment KEK in `tenant_keys`, crypto-shred via `DestroyTenantKeyTx`; proves the fact class is real, durable, and transactional today.

**Root cause.** The candidate derived owners top-down from business nouns and dispositioned `security` wholesale; key custody sits between platform mechanics (wrap/unwrap crypto) and governance (lawful destruction gate), so it fell through the classification seam.

**Invariant affected.** Packet §11.1 owner completeness and §11.2 owner uniqueness; R10 proof obligation 11 (retention-aware tenant erasure and restore tombstone reconciliation).

**Strongest credible alternatives.**
(a) Organization owns tenant key-custody facts (provision-at-onboarding, destroyed-at-erasure), with wrap/unwrap crypto remaining a platform mechanism; lawfulness input arrives from Records Governance blocker evaluation, which is already an existing edge.
(b) A dedicated narrow supporting owner (`support/tenantcrypto`) owning key lifecycle facts.
(c) Records Governance owns destruction gating while platform owns key rows — **rejected**: splits one fact's lifecycle across two authorities.

**Global Maximum analysis.** (a) is the smallest sustainable structure: Organization already owns Tenant lifecycle including ERASED and the erasure composition terminates there; DEK destruction is a component of that same lifecycle transition. (b) adds a module for one fact family with one writer — accidental structure.

**YAGNI.** No new capability; this is an assignment, not a construction.

**Required change.** Name the owner of tenant key-custody facts (recommend Organization, crypto mechanics in platform), add the fact family to the ownership inventory, and have the tenant-erasure and restore-reconciliation compositions cite that owner.

**R9.5 reopen?** NO — the frozen semantics are honored; only the R10-A assignment is missing.

**Proof to close.** Updated R10-A ownership inventory includes DEK/key-custody facts with exactly one owner; erasure/restore composition text names it.

### F2 — MAJOR — Authorization's "canonical resource/case relationship filtering contract" is ambiguous and risks migrating domain semantics into Authorization

**Claim.** Candidate Authorization owns "canonical grant evaluation" **and** "canonical resource/case relationship filtering contract" (packet §4.1-C). The frozen equation is `Permission + required resource/case relationship + Domain Governance constraints = ALLOW` (ledger §1). The relationship predicate is domain semantics (a Document belongs to an Area; CAPTURED Evidence reuses its primary Dossier scope; Dossier links never grant). If Authorization owns predicate **implementations**, it becomes a second business authority — a god policy engine every projection/export depends on, which the frozen anti-goals (no generic ACL/ReBAC) exclude. If it owns only the **contract shape**, then domains must own and supply their predicates — but the packet does not say which reading is intended. Packet §12.F poses exactly this question; the candidate text does not answer it.

**Evidence anchors.** Packet §4.1-C, §5 (Search "may consume published query/authz contracts"), §12.F; ledger §1 (equation), §11 (links never grant, projections reapply canonical AuthZ).

**Root cause.** "Canonical filtering contract" conflates two ownerships: the composition/evaluation authority (Authorization) and the per-relationship predicate semantics (each owning domain).

**Invariant affected.** §11.5 mechanism separation; §11.2 owner uniqueness; frozen "no generic ACL/ReBAC graph".

**Strongest credible alternatives.**
(a) Authorization owns grant evaluation plus the composable filter **contract shape**; each domain owns its relationship predicate and supplies it through that published contract; Search/export/timeline consume the composed result and never re-derive policy.
(b) Authorization owns all predicates — **rejected**: business semantics leak into Authorization; every domain visibility change becomes an Authorization change.
(c) Each surface reimplements visibility — **rejected**: duplicates the policy engine across projections; violates canonical-AuthZ obligation (R10 obligation 12).

**Global Maximum analysis.** (a) is the only shape in which both frozen requirements — one canonical evaluation and domain-owned relationship meaning — hold simultaneously.

**Required change.** R10-A must state (a) explicitly: predicate ownership stays per-domain; Authorization owns evaluation and the contract; projections are consumers only.

**R9.5 reopen?** NO.

**Proof to close.** Revised R10-A text names the predicate owner per relationship class and declares the Authorization contract as the single composition point.

### F3 — MAJOR — owner-completeness inventory is incomplete; several frozen durable facts have no assigned owner

**Claim.** Packet §11.1 requires demonstrating that "every frozen durable/business fact has exactly one semantic owner". The candidate provides owner noun-lists but no exhaustive fact inventory. An independent sweep of the frozen ledger against those lists found unassigned facts:

1. **TenantDeletionRequest / TenantErasureRecord / erasure tombstones** (ledger §6) — presumably Organization, but absent from §4.1-B;
2. **EditorSession** authoring lease (ledger §10) — presumably Controlled Information, absent from §4.1-D;
3. **Closed `ContentFormat` catalog** (ledger §8) — consumed by Artifact facts, DocumentType/EvidenceType allowed-format policy and `OfficialRepresentationPolicy`; no owner named; Dictionary's "bounded product/system value catalog" (§4.2-J) is ambiguous about whether it includes it;
4. **Restore-reconciliation process truth** (ledger §6: restore reapplies tombstones and reconciles retention/hold facts before service resumes) — Artifact owns byte "restore integrity facts" (§4.2-I) but the erasure/retention reconciliation process truth is unowned;
5. DEK facts — F1, kept separate for severity.

**Root cause.** Top-down noun derivation without a closing fact-inventory sweep; assertion where §11.1 demands demonstration.

**Invariant affected.** §11.1 owner completeness.

**Global Maximum analysis / required change.** The fix is structural for the closure artifact, not for the topology: R10-A closure must include a complete frozen-fact → owner inventory table. Recommended assignments: (1) Organization; (2) Controlled Information; (3) decide one authority — either explicitly inside Dictionary's product catalog or a platform-owned closed enum, never both; (4) Organization erasure/restore process truth, with Artifact keeping byte-integrity facts.

**R9.5 reopen?** NO.

**Proof to close.** Inventory table exists; an independent sweep finds no unassigned frozen durable fact.

### F4 — LOW — Dictionary standalone-owner status fails the Structural Inversion Test as presented

**Claim.** Frozen semantics evidence exactly one consuming lifecycle for tenant dictionary values: snapshot when a new REV is created (ledger §5) — a Controlled Information consumer. Current-state evidence (evidence only): the dictionary reader has exactly one business consumer, `documents` (`internal/modules/documents/application/service.go:141-143`). The System Value Catalog is product-owned with a different mutation lifecycle (code-shipped vs tenant-managed). Inverting the current tree: without a pre-existing `tokens` module, a fresh derivation would more likely classify tenant dictionary as tenant configuration consumed by Controlled Information than mint a peer supporting owner. Grouping the two value classes because "both produce values" is the same similarity-grouping defect the packet itself warns against for Interchange (§4.2-L). The packet's mandated challenge (§4.2-J) is upheld.

**Counterweight.** No invariant is broken either way; a separate owner does isolate the `dictionary.manage` admin surface; granularity is cheap to change at design time. Hence LOW.

**Required change.** Either (i) evidence a real second consumer from frozen semantics and keep Dictionary with the two value classes explicitly separated (tenant-managed values vs product-shipped catalog), or (ii) reclassify tenant dictionary as Controlled Information-owned configuration and place the product/system catalog with its single authority. Disposition below is UNKNOWN pending that adjudication.

**R9.5 reopen?** NO.

### F5 — LOW — dual-provenance risk at the Interchange boundary

**Claim.** Object-level creation/source provenance is owned by target domains (Dossier creation provenance, ledger §11; migrated-object source provenance, ledger §13), while Interchange owns "source provenance / transfer attempt / reconciliation identity" (packet §4.2-L). Both are called "provenance". Without an explicit split rule, a second provenance authority will emerge during R10-B/R10-F.

**Required change.** One binding sentence: object-level provenance facts live on the owning domain's objects; transfer/process/attempt truth lives in Interchange; neither duplicates the other.

**R9.5 reopen?** NO.

### F6 — LOW — Notifications is not fully rebuildable; the `projections/` classification overstates

**Claim.** Delivery/read-state is durable, user-facing and not rederivable from producers; Search is genuinely rebuildable. The frozen "delivery projection only" semantics are honored (no business meaning; acknowledgement never derives from it), but the package class `projections/` (packet §8) implies a rebuildability property Notifications lacks.

**Required change.** State explicitly that Notifications owns non-rebuildable delivery/read state as attributed non-business state (or classify it under `support/`). Classification precision only; no invariant broken.

**R9.5 reopen?** NO.

### F7 — LOW — the Audit same-commit durability seam must be explicit in R10-A

**Claim.** Frozen authority: critical governed mutation cannot report success without durable audit intent/event in the same commit boundary (ledger §5). With Audit as a separate owner, every mutating owner must append an AuditEvent inside its own transaction — which crosses the ownership seam unless Audit publishes a **transactional append port** (mechanism) distinct from its semantic ownership (timeline meaning, tamper-evidence, export/query, separate retention regime). The packet instructs the reviewer to challenge this distinction (§4.2-K) but does not declare the seam.

**Required change.** Declare the owners→Audit in-tx append edge as a published mechanism port; Audit remains the single timeline authority; producers never write the audit table directly.

**R9.5 reopen?** NO.

## 3. Candidate disposition table

| Candidate owner | Disposition | Basis |
|---|---|---|
| Authentication | **CONFIRM** | Frozen mandates separate AuthN behind the external-IdP seam; passes inversion. |
| Organization | **CONFIRM** | Correct cut; must additionally list deletion/erasure records (F3) and recommended key-custody facts (F1). |
| Authorization | **CONFIRM** | Correct cut conditional on F2 predicate-ownership clarification. |
| Controlled Information | **CONFIRM** | Large but single lifecycle authority; splitting authoring/release/rendition would recreate duplicate lifecycle authority (the current `documents`/`controlleddocuments` defect class); internal decomposition belongs to R10-B. |
| Approval | **CONFIRM** | Binds exact Submission, owns no Document lifecycle; bidirectional relation resolved by composition (DAG edge 1). |
| Documentary Context | **CONFIRM** | Evidence reuses primary Dossier scope; splitting Evidence from Dossier would put the most basic capture operation across two owners. |
| Records Governance | **CONFIRM** | Retention + hold + disposition cohere: disposition needs both eligibility and blockers; splitting strands it. |
| Distribution | **CONFIRM** | Distinct compliance facts (obligation, snapshot denominator, AcknowledgementRecord); frozen defines them independently of current code — passes inversion. |
| Artifact (supporting) | **CONFIRM** | Two-consumer derivation (Revision + Evidence) follows from frozen semantics alone; collapsing creates wrongful cross-domain content ownership; duplication duplicates integrity semantics. |
| Dictionary (supporting) | **UNKNOWN** | F4 — single evidenced consumer; standalone status and the tenant-vs-product merge require adjudication. |
| Audit (supporting) | **CONFIRM** | Distinct durable meaning (timeline/tamper-evidence/export + separate retention regime); conditional on F7 seam declaration. |
| Interchange (supporting) | **CONFIRM** | The four contracts share genuine transfer-truth machinery (batch/plan/dry-run/reconciliation, manifests, completeness, attempt/idempotency identity); splitting duplicates process-truth ownership or leaves thin wrappers around one shared mechanism — evidence the single owner is right. Anti-platform guard = enumerated contracts (exactly four) + stated negative boundary; F5 provenance rule required. |

Coordination compositions (§6): **CONFIRMED** — none conceals a missing first-class owner; each names the invariant-owning terminal authority (Submission→CI, Release→CI, Disposition→RG, Erasure→Organization). The R10 decomposition order A→B→C→D→E→F: **CONFIRMED** — no failure class is split incorrectly; no owner decision required prematurely deciding a later mechanism.

## 4. DAG verdict

**Acyclic is achievable with published seams.** No unavoidable semantic or package cycle was found. The strongest candidate cycles all resolve to composition-mediated or reference-only relations. Material edges that MUST be explicit in R10-A before closure:

1. **CI ↔ Approval via composition.** Submission coordination creates the ApprovalInstance; release coordination reads approval satisfaction and calls CI's atomic effectivity transition. Neither module imports the other; precondition evaluation lives in composition.
2. **In-tx published application ports.** Owners publish transactional ports; the composition layer opens one local DB transaction across owners (Submission + RetentionBinding + ApprovalInstance + audit + outbox intent). This is the load-bearing mechanism claim behind "no distributed transactions" — it must be a declared R10-A seam commitment, not an implicit assumption, or the transaction-boundary preview (§12.E) is unsubstantiated.
3. **Audit in-tx append port** (F7).
4. **Artifact confirmation takes an opaque owner reference.** Artifact never validates against CI/DC (no back-edge); no-confirmed-orphan is enforced by owner-driven confirmation choreography plus an R10-B DB backstop.
5. **RG prospective materialization** consumes CI/DC subject-entered-scope events plus published read ports; no RG back-edge into business lifecycle (holds block disposal only).
6. **Interchange requires each business owner to publish a privileged migration-grade import surface** (imported provenance, historical states without fabricated native approval). Interchange calls owners; owners never depend on Interchange.
7. **Producers resolve notification recipients** using Organization/Authorization published contracts before handing delivery intent to Notifications; Notifications never queries policy.
8. **Per-edge semantic vs compile-time resolution.** The §7 sketch arrows read as "provides-to"; final R10-A must state each edge as consumes/depends-on and resolve interface-inversion per edge, as the packet itself acknowledges (§7 last paragraph).

Transaction-boundary preview (§12.E): no frozen atomicity becomes impossible under this split, contingent on edge 2. Verified disposition and external publish remain two-phase (outbox) by frozen design; tenant erasure is a long-running composition whose process truth needs the F3-4 owner assignment.

## 5. Legacy disposition verdict

All 15 current modules have a target disposition; no indefinite compatibility status. Attack results:

- **`security` row — MATERIAL DEFECT (F1):** "pure operational security to platform" silently drops DEK/key-custody meaning into commodity mechanics. Must be corrected with F1's owner assignment.
- **`jobs` row — SOUND**, with a completeness note: the current periodic jobs re-home to semantic owners (document-review surfacer → Controlled Information; approval-SLA surfacer → Approval; release-hold reconciler → Controlled Information/Records Governance; audit-integrity validator → Audit; notifications fanout → Notifications; idempotency-janitor/outbox purge → platform). The R10-F map should name per-job destinations so no owner-attributed work is orphaned.
- **`tokens` row — PENDING F4** adjudication (converge to Dictionary vs reclassify into Controlled Information).
- All other rows (`approval`, `audit`, `auth`, `controlleddocuments`, `distribution`, `documents`, `iam`, `notifications`, `render`, `search`, `taxonomy`, `templates`) — **SOUND**: no retained hidden meaning, no dropped meaning, no duplicate owner, no inertia-driven preservation found. The `documents`/`controlleddocuments`/`templates` triple-delete into one Controlled Information owner is the correct root-cause resolution of the duplicate-lifecycle defect rather than a rename of it.

## 6. Structural Inversion and subtractive/YAGNI pass

**Inversion.** Every business bounded context (A–H) and Artifact/Audit/Interchange derive from frozen R3–R9.5 semantics and survive inversion of the current tree. **Dictionary fails as presented** (F4). Provider reversal (MinIO/S3/local, EigenPal/other editor, Gotenberg/other renderer, SharePoint/other repository) alters no ownership conclusion — no provider leakage found (§12.G clean).

**Subtractive pass.** Removal was attempted for every candidate owner:

- Dictionary — removable into Controlled Information without weakening a distinct property, unless a second consumer is evidenced (F4, UNKNOWN);
- Interchange — not removable: splitting duplicates process-truth ownership; per-domain import/export recreates a hidden orchestrator for cross-domain packages;
- Distribution — not removable: distinct compliance facts; folding grows a god CI;
- Audit — not removable: transversal timeline/export/retention-regime semantics need one owner;
- Documentary Context — not removable into CI: frozen separates Evidence from REV/Approval lifecycle;
- Authentication/Organization/Authorization merges — rejected: frozen separations and the evidenced IdP seam;
- Artifact — not removable: two-consumer byte truth.

No merged owner increases semantic ambiguity; Controlled Information's size is essential complexity (one lifecycle authority) with its negative-boundary list as the God-context guard. **No candidate owner exists solely for imagined future consumers except possibly Dictionary (F4).**

**Evidenced future seams (§12.J).** External IdP → Authentication adapter seam; realtime collaboration → WorkingContent provider seam inside CI; SharePoint Embedded → future content profile behind Artifact/Interchange; advanced content security → Artifact staging→validation→confirmation seam; PKI/signature level → Approval attestation evidence; ArtifactPackage → Artifact extension; second records/export/integration role → Authorization bundle change only. None requires rewriting core identity/authority. PASS.

## 7. R9.5 reopen set

**EMPTY.** F1 and F3 are R10-A assignment gaps against frozen semantics, not invalidations of any frozen invariant/authority. No finding satisfies (or needs) the five-part reopen contract.

## 8. Final verdict

```text
VERDICT: APPROVE R10-A WITH MATERIAL FIXES
```

Required before R10-A closure: F1 (DEK/key-custody owner), F2 (Authorization predicate-ownership statement), F3 (complete fact→owner inventory), plus the eight explicit DAG edges in §4. F4–F7 are LOW and may close with one-line adjudications.

**Single biggest remaining structural risk:** F2 — the ambiguity of Authorization's "canonical resource/case relationship filtering contract". It is the one place this topology can silently grow a second business-policy authority: if domain relationship predicates migrate into Authorization (or projections re-derive them), the frozen `Permission + relationship + Domain Governance` equation ends up owned by no one and implemented by everyone.
