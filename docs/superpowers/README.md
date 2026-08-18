# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 + T3 + Decision Registry CLOSED / OPERATOR-RATIFIED; T4 DECISIONS ACCEPTED / PLATFORM SUMMARY RATIFICATION NEXT.**

Durable accepted truth belongs in `wiki/`. Active, not-yet-promoted design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Current active staging

- `analysis/2026-08-18-r10-t4-exact-content-storage-integrity-restore-candidate.md` — **T4 accepted technical candidate; not yet promoted.**
- `analysis/2026-08-18-r10-t4-operator-adjudication.md` — **T4-A→T4-O accepted; platform summary ratification next.**

Completed T3 candidate/adjudication staging was removed after durable promotion. Git history is the archive.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

For every remaining T-stage:

```text
CURRENT / PRESERVE / REFINED → baseline
REOPEN                       → design in owning T-stage
DEFERRED                     → future seam/counterexample only
SUPERSEDED                   → reject inheritance absent explicit material reopen
```

## T4 preserved baseline examples

```text
no standalone Artifact semantic owner
exact-content facts belong to the semantic record that owns/freezes them
storage/provider identity never semantic identity
WorkingContent OCC is correctness authority
Submission immutable exact candidate
OfficialRendition binds exact Submission
provider calls do not join semantic transaction
Object Lock/WORM/versioning are mechanism/enforcement only
restore with missing/corrupt required bytes fails closed
```

## T4 adjudicated headline

```text
ExactContentDescriptor = SHA-256 + size + ContentFormat
no mandatory whole-Submission JCS digest Launch
opaque managed-content handle = mechanism only
provider-neutral ManagedContentStore; one active store/deployment
Local + AWS S3 reference profiles
OPEN→READY admission with server-derived byte facts
opaque admission binding
malware CLEAN required for untrusted governed admission; no scan every autosave
create-once/no-overwrite
WorkingContent current state = DRAFT recovery baseline
SUBMIT/Rendition zero provider/scanner calls inside semantic tx
only unreferenced/non-governed DRAFT mechanism content is reclaimable Launch
backup couples DB recovery point + exact required-content manifest/copy + GC fence
restore verifies required size/SHA-256/format before serving
older restore reconciles post-snapshot UserProfile erasures via minimum independent barrier/journal
future content capabilities reuse descriptor+mechanism without Artifact owner
```

## Mandatory T-stage closure protocol

```text
read registry
→ candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promotion/closure
→ update Decision Registry
→ remove completed staging
→ only then Tn+1
```

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         DECISIONS ACCEPTED / SUMMARY RATIFICATION PENDING
T5 Durable Async, Search & External Effects           NOT OPEN
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ Integrated Whole-R10 GCR
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

## Hard stop

No product implementation or implementation plan is authorized while active design gates remain open.
