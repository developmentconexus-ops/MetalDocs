# R10-T4 — Operator Adjudication / Summary Ratification Gate

> **Status:** ACTIVE STAGING — T4 DECISIONS ADJUDICATED / PLATFORM SUMMARY RATIFICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Candidate:** `docs/superpowers/analysis/2026-08-18-r10-t4-exact-content-storage-integrity-restore-candidate.md`  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This record captures operator adjudication of T4 after the operator requested and received a product-level explanation of why exact-byte identity, immutable content admission and restore verification are essential rather than overengineering. It does **not** close T4 or open T5. T4 closes only after explicit operator ratification of the required platform-facing T4 summary.

## 1. Operator comprehension basis

The operator approved T4 after the architecture was explained in terms of four product promises:

```text
1. “You approved this exact content.”
2. “This is the exact official content now in effect.”
3. “Governed history did not silently change underneath its business identity.”
4. “Backup/restore truly recovers both semantic state and the exact required content.”
```

The accepted minimum is therefore about **proving exact governed content**, not building a generic byte-management/content platform.

## 2. T4 adjudication

The operator accepted all recommendations T4-A→T4-O as written:

```text
T4-A ACCEPT — ExactContentDescriptor = SHA-256 + exact size + closed ContentFormat.
T4-B ACCEPT — no mandatory RFC8785/JCS whole-Submission composite digest in Launch; exact-byte digest is required, whole-Submission canonical digest is deferred until a named consumer exists.
T4-C ACCEPT — opaque managed-content UUID handle is retrieval mechanism only; it is not semantic identity and is not the hash/provider key.
T4-D ACCEPT — one provider-neutral ManagedContentStore contract and one active store per deployment; Local is dev/test/conformance, AWS S3 is the reference production profile.
T4-E ACCEPT — upload/admission mechanism uses OPEN→READY; server derives/verifies exact SHA-256, size and format from exact stored bytes; client claims are hints only.
T4-F ACCEPT — opaque admission/root binding prevents arbitrary/cross-root handle reuse while keeping storage unaware of Document/Revision semantics; legitimate copy produces a new handle.
T4-G ACCEPT — UNTRUSTED_EXTERNAL content requires CLEAN malware proof before immutable governed admission in production; do not scan every DRAFT autosave; ClamAV/clamd is the reference MalwareInspector mechanism.
T4-H ACCEPT — admitted managed content is create-once/no-overwrite; DRAFT replacement creates a new handle.
T4-I ACCEPT — current WorkingContent is the Launch DRAFT recovery authority; no mandatory WorkingSnapshot/business autosave-history model.
T4-J ACCEPT — SUBMIT and OfficialRendition semantic transactions make zero provider/scanner calls; preflight READY/descriptor/malware proof is revalidated and the exact handle+descriptor are frozen atomically with semantic truth.
T4-K ACCEPT — only unreferenced/non-governed DRAFT mechanism objects are reclaimable in Launch; immutable governed content has no ordinary Launch delete/disposition capability.
T4-L ACCEPT — backup couples one DB recovery point with an exact manifest/copy of every managed-content handle required by that recovery point, including current referenced DRAFT WorkingContent; backup capture fences GC for selected handles.
T4-M ACCEPT — restore remains non-serving until every required handle exists, is readable, and matches exact size/SHA-256/format.
T4-N ACCEPT — restore from an older recovery point must reconcile later lawful UserProfile erasures through the minimum independently retained erasure barrier/journal before restored profile data may serve; this is not a generic privacy workflow platform.
T4-O ACCEPT — future Evidence/Records/Export/Repository capabilities reuse stable semantic identities, ExactContentDescriptor and the provider-neutral content mechanism; standalone Artifact ownership must not return by convenience.
```

## 3. Why this was judged not overengineering

Every accepted Launch mechanism has a current named consumer:

```text
SHA-256                    → prove exact submitted/released/rendition content
size                       → cheap integrity + restore validation
ContentFormat              → governed format/view/render behavior
opaque handle              → retrieve bytes without provider identity becoming domain identity
create-once/no-overwrite   → preserve hash/malware/Submission exactness
OPEN→READY                 → provider PUT success is not semantic admission
malware gate               → untrusted bytes cannot become governed truth unsafely
restore verification       → Product Contract backup/restore correctness
```

Explicitly rejected/deferred as overengineering absent a consumer:

```text
Artifact semantic owner / ArtifactStage business model
whole-Submission JCS digest
multi-cloud/BYOS/active-active
replication/dedup/content-addressed identity
Object Lock/WORM as lifecycle authority
quarantine/CDR/rescan platform
PKI/signature infrastructure
confirmed governed-content delete/disposition Launch API
generic privacy-case platform / per-user crypto-erasure machinery
```

## 4. Current gate

```text
T4 material decisions       = ADJUDICATED / ACCEPTED
T4 platform summary         = NEXT
T4 final closure/promotion  = PENDING SUMMARY RATIFICATION
Decision Registry update    = PENDING T4 CLOSURE
T5                          = NOT OPEN
implementation              = BLOCKED
```

Per the operator-approved stage protocol:

```text
T4 design
→ T4 adjudication ✅
→ platform-facing T4 summary NEXT
→ explicit operator summary ratification
→ promote/close T4
→ update Decision Registry
→ remove completed T4 staging
→ only then open T5
```
