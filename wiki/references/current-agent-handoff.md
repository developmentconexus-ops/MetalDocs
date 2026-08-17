# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5-8 IN REVIEW**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Follow `AGENTS.md` first. The active route is:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md` for the current independent-review packet

Git history is archive. Do not revive historical specs/ADRs/runtime concepts by inertia.

## Current checkpoint

R3–R9 principal governance is locked. R9.5 whole-product completion has locked:

1. **Content Model** — format-agnostic Document; immutable Artifact; semantic Evidence/EvidenceType; one primary Artifact; canonical naming; format-aware official representation.
2. **Storage/Repositories** — one Managed Artifact Store/deployment; Local/MinIO/AWS S3 adapters; provider-independent hashes/keys; external repositories use explicit import/publish copies; SharePoint Embedded future profile.
3. **Authoring/EigenPal** — persisted WorkingContent + technical working_version/OCC; one writer V1; immutable Submission boundary; Approval read-only; realtime collaboration deferred behind seam.
4. **Dossier/Context** — stable context key/scope; small DossierType; M:N Document links; CAPTURED Evidence has immutable primary Dossier; ExternalReferences; explicit ERP/PLM/PM/CMMS boundaries.
5. **Retention/Records/Legal Hold** — RetentionBinding on CAPTURED Evidence / first submitted Revision; type-scoped policy snapshots; explicit disposition; LegalHold over Evidence/Document/Dossier; WORM/Purview enforcement only; tenant erasure blocked while obligations survive.
6. **Import/Migration/Export** — ordinary import distinguished from privileged Historical Migration; external history never fabricated as native MetalDocs events; current-state/full-history modes; deterministic revision/code mapping; MigrationBatch dry-run/idempotency/reconciliation; provider-independent portability/governed-subject manifests with hashes and no secrets/runtime internals.
7. **Launch Attestation + Basic Content Safety** — approval binds exact Submission/digest with actor/step/policy/time/assurance evidence; return-for-changes requires reason; application approval does not claim ICP-Brasil/qualified signature; approval may be manifested in a Rendition without mutating source bytes; launch upload only needs supported-format allowlist, basic size/type coherence and non-execution/download-safe behavior.

**R9.5-8 is now IN REVIEW, not closed.** The integrated candidate packet exists and currently recommends `APPROVE / FREEZE R9.5`, but the DevelopmentConexus Engineering Method requires an independent/fresh adversarial challenger before this major architecture boundary is ratified.

Previous R10-A topology remains **not approved** and blocked.

## Candidate R9.5-8 deltas under independent challenge

The review packet contains the complete reasoning. The only material candidate deltas are:

- Submission must freeze one coherent WorkingContent version across Artifact + governed metadata + decision-relevant structured/template provenance.
- WorkingContent is format-agnostic DRAFT authority; EigenPal is a provider, not its definition.
- **Bounded reopen of R9.5-3 external-edit stale-base wording:** while a Revision is DRAFT, an authorized actor may deliberately replace current WorkingContent; OCC prevents overlapping-operation last-write-wins, but MetalDocs does not track/infer arbitrary file ancestry from prior downloads. `SUBMIT` is the freeze boundary and SUBMITTED rejects mutation.
- Approval review resolves only the exact Submission; unsupported/unproven preview paths fall back to a supported inspection path for that same Submission rather than creating a second content authority.
- Active Document/Dossier LegalHold scope continues materializing newly entering retention subjects while the hold remains active; removal/unlink does not release already-held subjects.
- cross-scope Dossier/Document/Evidence links are context only and never grants; projections and exports reapply canonical AuthZ.
- export completeness must be explicit and authorization-safe.
- candidate bounded R9.5 authorization delta contains **16 semantic permissions** and no new role/provider-specific permission engine.

These are candidate outcomes until the independent review is dispositioned; the active ledger remains target authority until promotion.

## Mandatory adversarial coverage already staged

The current packet attacks all required cases:

1. in-app DOCX procedure from DRAFT → Submission → Approval → Release;
2. externally edited/uploaded XLSX governed Document;
3. native PDF/SVG/CAD-style controlled source without universal PDF assumption;
4. Evidence upload/classification/naming inside Sale Dossier;
5. Product Dossier with mechanical/electrical/manual Documents and inspection Evidence without drifting into PLM;
6. stale autosave / concurrent DRAFT replacement;
7. return-for-changes + same REV + new immutable Submission;
8. MinIO→S3 physical relocation with unchanged business identity;
9. SharePoint IMPORT_COPY/PUBLISH_COPY and external drift without silent mutation;
10. historical migration with incomplete bytes/dates/approval evidence;
11. retention expiry + LegalHold + explicit disposition;
12. tenant deletion blocked by retained/held content;
13. external ERP/PLM object disappearing while Dossier history remains;
14. cross-scope AuthZ, case visibility and strict Approval SoD;
15. renderer/storage/job failure without corrupting business truth.

## Explicitly deferred from launch

These remain future triggers, **not hidden V1 TODOs**:

```text
malware scanning / ClamAV / quarantine / periodic rescans
ArtifactSecurityAssessment/CDR/advanced active-content analysis
ICP-Brasil/PKI/DocuSign/Adobe Sign/RFC3161/TSA/HSM
cryptographically signed export packages / package-signing-key lifecycle
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery/ESI preservation
realtime coauthoring / WOPI-style collaboration
true indivisible multi-file ArtifactPackage without a real supported format requiring it
```

## Important launch truths

- `DRAFT` content is mutable persisted WorkingContent; edit/upload/replacement is allowed to authorized actors while it remains DRAFT.
- every DRAFT-mutating path participates in the same `working_version` / OCC control; concurrency protection prevents races, not deliberate later replacement after the actor observes current state.
- `Submission` is the immutable review boundary; Approval/Rendition/Release bind exact Submission/digest.
- preserve actor, Step, ApprovalPolicy version, server time and required AuthN/fresh-auth evidence.
- MetalDocs V1 claims authenticated application approval, not a legal-signature level it has not implemented.
- no stamping/mutation of approved source bytes; human-readable approval manifestation is derived.
- DocumentType/EvidenceType only accept explicitly supported ContentFormats.
- client filename is provenance only; canonical naming is MetalDocs-owned.
- unsupported/complex formats need not be rendered inline; safe exact-source download is sufficient where the product supports that inspection path.
- basic validation is required; a security platform is not.

## Exact next step — independent R9.5-8 adversarial challenge

Run an independent/fresh READ-ONLY challenge of:

`docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md`

The challenger must reconstruct authority first, attack the candidate findings and all 15 mandatory scenarios, run the subtractive/YAGNI pass, and end with exactly one material verdict:

```text
VERDICT: APPROVE / FREEZE R9.5
```

or

```text
VERDICT: REOPEN — <minimal materially implicated decisions>
```

Reviewer findings are evidence, not requirement authority. Reopen only on a concrete invariant/authority/failure counterexample.

**Only after that review is dispositioned and the operator explicitly ratifies the freeze may R9.5-8 become CLOSED and R10 bounded contexts/filesystem/data model begin.**
