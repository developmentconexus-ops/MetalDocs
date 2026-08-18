# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + R10-A/B1/B2 PROMOTED + R10-B3/B4/B5 ACCEPTED FOR R10 INTEGRATION / NON-FINAL + R10-B6 NEXT**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file — current checkpoint / exact next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program/global-coherence authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 authority
6. `wiki/architecture/r10-technical-architecture.md` — promoted R10 authority through integrated B2
7. `docs/superpowers/analysis/2026-08-17-r10-b3-controlled-information-artifact-integrated-candidate.md` — accepted B3 working target
8. `docs/superpowers/analysis/2026-08-18-r10-b4-approval-rendition-release-distribution-integrated-candidate.md` — accepted corrected B4 working target
9. `docs/superpowers/analysis/2026-08-18-r10-b4-integration-acceptance.md` — B4 adjudication / bounded refinement
10. `docs/superpowers/analysis/2026-08-18-r10-b5-documentary-context-records-governance-artifact-closure-integrated-candidate.md` — accepted corrected B5 working target
11. `docs/superpowers/analysis/2026-08-18-r10-b5-integration-acceptance.md` — B5 adjudication + B3 bounded refinements
12. older review/evidence artifacts only when auditing an already-promoted/accepted decision

Git history is archive. Current code/schema/OpenAPI/module docs remain current-state/migration evidence only.

---

## Current checkpoint

```text
R3–R9.5 = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED

R10-A   = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2  = CLOSED / APPROVED / INTEGRATED
R10-B3  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B4  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B5  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
R10-B6  = NEXT / DESIGN ONLY
R10-C   = NOT STARTED
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Promoted authority remains R3–R9.5 + R10-A/B1/B2. B3/B4/B5 are current non-final integration authority and remain challengeable only by a material later-stage counterexample.

Whole-R10 Global Coherence Review + cold independent review occur before final R10 ratification unless a truly exceptional material trust-boundary/irreversible/cross-repository blocker requires earlier independent review.

---

## R10 working-mode rule

From B3 onward:

```text
repo authority
→ evidence + directed external research where useful
→ DevelopmentConexus Method
→ integrated candidate
→ internal adversarial challenge
→ cross-stage coherence
→ operator understanding/adjudication
→ ACCEPTED FOR R10 INTEGRATION / NON-FINAL
```

Do not return to per-table/per-field/microdecision review gates. Do not reopen an accepted decision for preference or implementation convenience.

---

## Accepted B3/B4 working kernel

```text
Document
→ business DocumentRevision
→ WorkingContent + monotonic OCC generation
→ immutable RevisionSubmission
→ one sequential Approval/Governance route
→ optional immutable Rendition
→ automatic system-owned Release
→ concrete Distribution obligations / explicit acknowledgement
```

Key current overlays:

- exact Submission is the downstream governance identity;
- SUBMIT consumes WorkingContent generation; returned same-REV work never mutates old Submission;
- B4 removed the historical `review|approval` Step discriminator: one governance Step with actor rule, ANY|ALL, optional fresh-auth/due date, universal comment/annotation/suggestion surface;
- Approval requirement + PolicyVersion and representation requirement are snapshotted per Submission;
- Release is sole effectivity transition and snapshots Distribution obligations in the winning transaction;
- viewer/editor/render/storage providers never become business identity/permission authority.

---

## Accepted B5 working outcome

### Documentary Context / Evidence

- Dossier = stable documentary context only; never folder/storage owner/access grant/retention authority/ERP-PLM master.
- Dossier↔Document is M:N over stable Document identity and never changes target lifecycle or AuthZ.
- Evidence is a separate small capture lifecycle `DRAFT → CAPTURED → VOIDED`; CAPTURE freezes immutable payload/metadata + exactly one primary Artifact and creates RetentionBinding.
- every CAPTURED Evidence has exactly one immutable primary Dossier and may have secondary contextual Dossier links; Evidence reuses primary-Dossier scope.
- no generic Dossier fields/forms/workflow/ACL/completeness engine and no generic Record declaration.

### Records Governance

- first DocumentRevision Submission / Evidence CAPTURE automatically create immutable typed `RetentionBinding`;
- type policy = `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`;
- anchors derive from owner lifecycle facts; Audit never becomes retention-clock authority;
- RetentionExtension only lengthens and is forbidden after DispositionFence;
- LegalHold scopes = Evidence | stable Document | Dossier and materializes exact RetentionBindings, including newly entering live scope;
- Hold activation is fail-closed/all-or-nothing when current scope contains an active disposition fence;
- expiry only means disposition eligibility; no automatic delete;
- semantic DispositionFence precedes external removal; immutable DispositionRecord exists only after verified physical + semantic completion;
- business lifecycle state and records disposition remain separate axes.

### Artifact closure

- every confirmed Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence;
- multiple references inside that root are allowed; cross-root Artifact-row reuse is rejected;
- same SHA across different roots may use separate Artifact semantic rows; physical dedupe remains mechanism freedom;
- Artifact has no independent retention policy/ref-count/owner registry.

### B5-approved bounded B3 refinements

1. `DocumentOrigin`: replace strong source-Submission FK with source Revision identity FK + immutable exact Submission/digest/hash provenance snapshots so lawful template-source disposition does not become impossible.
2. `DocumentRevision`: add canonical one-shot `cancelled_at` / `obsoleted_at`; native superseded anchor remains B4 `ReleaseRecord.released_at` and is not duplicated.
3. move immutable `dictionary_snapshot` off the permanent revision identity skeleton into `RevisionDictionarySnapshot`, which belongs to the Revision retention unit and may be removed by completed lawful disposition.

No B4 semantic was reopened by B5.

---

## Exact next step — R10-B6 Audit + Interchange + Cross-owner Atomicity

Open **R10-B6** as one integrated batch from promoted B1/B2 + accepted non-final B3/B4/B5.

B6 must jointly close at minimum:

```text
Audit
  minimum append-only AuditEvent semantic skeleton
  actor/resource/operation/trusted-time fields
  field-by-field PII classification and erasable enrichment boundary
  grant/revoke forensic reconstruction after current-state rows are deleted
  tamper-evidence/query/export contract boundaries
  Audit's separate retention regime

Interchange
  Historical Migration batch/plan/dry-run/item outcome/reconciliation persistence
  imported-history truth vs native domain facts
  Governed Subject Export process/package facts
  External Repository IMPORT_COPY / PUBLISH_COPY process truth
  connection/reference ownership required for Dossier ExternalReference / repository publication
  no Tenant Portability Export V1

Cross-owner atomicity
  final B1–B5 same-local-commit matrix
  required Audit append points
  required durable async intent points routed to R10-D
  transaction-composition seams without repository imports/nested commits
  final lock-order/deadlock challenge under READ COMMITTED
  imported-history DB coherence with native target constraints
```

B6 must explicitly challenge:

```text
B2 privacy/offboarding vs immutable Audit evidence
B3/B4/B5 semantic mutations requiring same-commit Audit
B5 Disposition vs Audit's separate retention regime
B5 surviving post-disposition skeleton privacy
Historical Migration without fabricating Approval/Release/native actors
Export completeness without granting access through context/projection
External Repository effects without provider atomicity
whole B1–B5 lock graph and cross-owner transaction order
```

Keep outside B6:

```text
physical storage/object-lock/malware/restore       → R10-C
async worker/retry/lease/projections/effects       → R10-D
API/frontend/viewer/editor user journeys           → R10-E
historical cutover/legacy deletion/bootstrap ops   → R10-F
```

Current implementation remains evidence only. Implementation remains **BLOCKED**.