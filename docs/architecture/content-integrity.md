# R10-T4 — Exact Content, Storage Integrity & Restore

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Post-T5 Fable bounded amendment:** 2026-08-18 — restore security non-resurrection + admission-claim GC liveness  
> **T8-E bounded correction:** 2026-08-21 — already-PDF required rendition reuses admitted bytes; rendering only when transformation is required
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **T3 authority:** `wiki/architecture/r10-t3-authorization-audit-enforcement.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T4 architecture plus bounded completeness amendments ratified through the post-T5 independent-review checkpoint. T4 defines how MetalDocs proves, retrieves and restores the exact content owned/frozen by WorkingContent, Submission, OfficialRendition and later imported-history facts while storage remains replaceable mechanism.

The operator accepted T4-A→T4-O and explicitly ratified the platform-facing T4 summary on 2026-08-18 after a dedicated comprehension check proving why exact-content identity is essential to a controlled-document product rather than overengineering.

---

## 1. Product promises T4 protects

T4 exists to make four Launch claims falsifiable:

```text
1. “You approved this exact content.”
2. “This is the exact official content now in effect.”
3. “Governed history did not silently change underneath its business identity.”
4. “Backup/restore truly recovers both semantic state and the exact required content without silently resurrecting later-invalid access/privacy state.”
```

The minimum architecture is therefore about proving exact governed content and safe restore readiness, not building a generic byte-management/content/security-recovery platform.

---

## 2. Preserved baseline

```text
no standalone Artifact semantic owner
exact-content facts belong to the semantic record that owns/freezes them
storage/provider identity never becomes semantic identity
WorkingContent = sole mutable DRAFT authority
WorkingContent OCC/CAS = DRAFT race authority
Submission = immutable exact governed attempt
OfficialRendition binds exact Submission
native/imported history remain distinguishable
provider/external calls never join local semantic transaction
Object Lock/WORM/provider versioning never becomes lifecycle authority
restore with missing/corrupt required bytes is not healthy
restore must not silently resurrect lawfully erased UserProfile PII
restore must not silently reactivate restored ApplicationSessions or known post-snapshot access teardown
```

T4 must not be used to resurrect:

```text
Artifact aggregate/owner
ArtifactStage business model
Artifact retention-root model
provider key/bucket/version as semantic identity
content-addressed semantic entity identity
confirmed orphan Artifact library
multi-cloud/BYOS/active-active product platform
confirmed-governed-content delete/disposition Launch API
generic per-grant security teardown journal/platform
```

---

## 3. ExactContentDescriptor — T4-A

Every semantic record that owns/freezes exact bytes uses the minimal semantic descriptor:

```text
ExactContentDescriptor
  sha256
  size_bytes
  content_format
```

Laws:

```text
sha256        = SHA-256 of exact raw bytes
size_bytes    = exact raw-byte length
content_format= closed MetalDocs ContentFormat vocabulary
```

Not exact-content identity:

```text
filename
client Content-Type
provider ETag/checksum
bucket/key/version
URL
upload metadata
```

The descriptor is owned by the semantic fact that needs historical truth:

```text
WorkingContent    → current DRAFT descriptor
Submission        → immutable submitted descriptor
OfficialRendition → immutable rendition descriptor
imported history  → T7 target-owned exact descriptor
```

The same physical handle may be referenced by WorkingContent and the Submission created from it, but each semantic record owns its own descriptor truth.

---

## 4. No mandatory whole-Submission canonical digest — T4-B

Launch does not require RFC8785/JCS or another canonical whole-Submission composite digest.

Current proof is already provided by:

```text
stable immutable Submission identity
+ immutable Submission fields/snapshots
+ exact source-byte SHA-256 descriptor
+ governance/Release binding exact Submission id
```

A versioned canonical package/digest is deferred until a named consumer such as digital signature, external non-repudiation or signed export requires it.

---

## 5. Opaque managed-content handle — T4-C

Semantic records may carry:

```text
managed_content_id UUID
```

It is retrieval mechanism only.

```text
managed_content_id != semantic content identity
managed_content_id != SHA-256
managed_content_id != provider key/version
```

Provider migration may preserve the same handle while changing physical location. Semantic exactness remains the descriptor.

The durable mechanism may use the technical lifecycle:

```text
OPEN
→ READY
→ GC_PENDING
```

There is no semantic `CONFIRMED Artifact` state. A READY object becomes preserved governed content because a semantic record references it, not because the storage mechanism becomes a business owner.

---

## 6. ManagedContentStore contract and provider profiles — T4-D

Launch uses one active provider-neutral `ManagedContentStore` per deployment.

Minimum contract:

```text
opaque UUID-addressed handle
create-once / no overwrite
exact full read
optional bounded/range read
copy authorized existing managed content to a new handle
explicit missing/failure
cleanup of reclaimable mechanism content
```

Conceptual operations:

```text
PresignCreate(handle, constraints)
Stat(handle)
OpenExact(handle)
OpenRange(handle, ...)
CopyToNewHandle(source, destination)
DeleteReclaimable(handle)
```

There is no ordinary Launch serving/domain `DeleteGovernedContent` operation.

Provider profiles:

```text
Local   → first-class dev/test/conformance provider
AWS S3  → reference production provider
```

Both must prove the same core contract. S3-specific checksums, ETags, Versioning and Object Lock remain mechanism evidence only.

A deployment requiring multiple simultaneously active stores is a T4 reopen trigger.

---

## 7. Upload/admission OPEN→READY — T4-E

Browser/external ingress never writes governed semantic truth directly.

```text
allocate managed_content_id
→ OPEN
→ short-lived create-only upload
→ server independently reads exact stored bytes
→ derive SHA-256 + size + actual ContentFormat
→ basic non-executing structural validation
→ READY
```

Client filename, Content-Type and hash are hints only.

`provider PUT success != semantic admission`.

Multipart upload is not baseline; it reopens only when real supported file-size evidence requires it.

---

## 8. Opaque admission binding — T4-F

Knowing a handle UUID never authorizes attaching it to an arbitrary semantic operation/root.

```text
managed-content allocation
→ server binds handle to intended operation/root through an opaque unforgeable binding/claim
→ owning use case must re-prove that binding before attachment
```

The mechanism does not parse or own Document/Revision/Submission semantics. No generic `owner_type/owner_id` registry is introduced.

A **live admission claim/binding reserved for an in-flight attachment** protects that READY handle from GC eligibility until the claim is consumed, explicitly released or reaches a bounded mechanism expiry. The claim is technical liveness/authorization state, not business ownership or retention.

Legitimate cross-root copy/template seeding creates a new authorized handle. Same-Revision unchanged resubmission may reuse the already-authorized handle.

---

## 9. Malware governed-boundary gate — T4-G

Production safety invariant:

> **Untrusted external bytes cannot become immutable governed MetalDocs content without successful malware inspection of those exact immutable bytes.**

Server-selected trust classes:

```text
UNTRUSTED_EXTERNAL
TRUSTED_MANAGED_COPY
TRUSTED_INTERNAL_DERIVATION
```

Baseline examples:

```text
browser upload             → UNTRUSTED_EXTERNAL
external-repository import → UNTRUSTED_EXTERNAL
historical migration file  → UNTRUSTED_EXTERNAL
authorized admitted copy   → TRUSTED_MANAGED_COPY
renderer output            → TRUSTED_INTERNAL_DERIVATION
```

Production policy:

```text
UNTRUSTED_EXTERNAL → CLEAN required before immutable governed reference commit
trusted copy/internal derivation → no mandatory rescan baseline
```

Fast DRAFT autosaves may reach READY without scanning every debounce. The exact READY handle is scanned before crossing a governed boundary such as SUBMIT/imported-content admission.

Scanner unavailable/incomplete → visible retriable failure; no governed admission.
Malicious result → content is not attachable to immutable governed truth.

`MalwareInspector` is provider-neutral; ClamAV/`clamd` is the reference mechanism.

Scanner-disabled dev/test data cannot silently be declared production-ready.

---

## 10. Create-once / no-overwrite — T4-H

Every admitted handle maps to create-once bytes:

```text
same handle + second PUT → rejected
DRAFT replacement        → new handle
immutable semantic fact  → descriptor + create-once handle
```

Browser/domain write paths receive no overwrite capability.

The S3 reference profile uses/enforces an equivalent conditional-create property. Versioning/Object Lock are not required semantic dependencies.

This makes SHA-256 and malware evidence stable because the verified/scanned bytes cannot be changed afterward through admitted application paths.

---

## 11. DRAFT recovery — T4-I

DRAFT autosave persists only current WorkingContent plus its READY handle/descriptor under T2 OCC:

```text
new handle
→ upload
→ READY
→ WorkingContent CAS/OCC swap
→ generation++
```

WorkingContent remains the current DRAFT recovery authority.

Launch introduces no mandatory:

```text
WorkingSnapshot business history
business autosave history
EditorSession correctness dependency
```

Crash/reload recovery uses the persisted current WorkingContent. Rich undo/checkpoint history reopens only on a real UX/recovery requirement. OCC remains correctness authority.

---

## 12. SUBMIT / OfficialRendition semantic admission — T4-J

Provider/scanner work finishes before the local semantic transaction.

For SUBMIT:

```text
preflight malware scan when required
→ BEGIN local semantic transaction
→ serialize Document/WorkingContent under T2
→ prove expected WorkingContent generation
→ prove current handle READY/create-once/admission-bound
→ prove mechanism descriptor == WorkingContent descriptor
→ prove required malware proof belongs to exact same handle/bytes
→ create immutable Submission copying descriptor + handle
→ freeze governance/representation snapshot
→ Revision DRAFT → SUBMITTED
→ required T3 Audit
→ COMMIT
```

No provider/scanner call occurs inside the semantic transaction.

Rollback creates no Submission truth even when provider upload already succeeded. The READY handle remains retry/reclaim mechanism state once any live admission claim is released/expires.

Required OfficialRendition has exactly two Launch realization paths:

```text
submitted source already PDF + required format PDF
  -> revalidate the exact already-admitted Submission handle + descriptor
  -> create OfficialRendition semantic fact over that same handle + descriptor
  -> no provider copy
  -> no renderer execution
  -> no durable rendition intent

submitted source DOCX + required format PDF
  -> render outside the semantic transaction
  -> admit/verify the produced READY PDF through T4
  -> final admission revalidates exact READY content and Submission eligibility
  -> create OfficialRendition semantic fact
```

The first path is not a semantic downgrade: OfficialRendition still exists as the required immutable fact, but byte-for-byte PDF duplication/transformation is removed because it proves no additional property. Target proof must show same-PDF rendition preserves the exact Submission handle/descriptor and invokes neither `CopyToNewHandle` nor renderer/durable-intent machinery.

T7 Historical Migration uses the same admission seam for untrusted imported exact content.

---

## 13. Reclaimable DRAFT cleanup / no governed-content delete — T4-K

Launch has no governed-content physical disposition.

Only non-governed mechanism content that is no longer required may be reclaimed.

Eligibility concept:

```text
BEGIN
lock mechanism object
prove not current WorkingContent
prove no immutable Submission/Rendition/imported fact references handle
prove no live admission claim/binding reserves handle for an in-flight attachment
prove no backup exclusion/pin protects handle
mark GC_PENDING
COMMIT
```

Physical delete is outside the semantic transaction and T5 owns worker/retry mechanics.

Immediately before provider delete, execution must re-read/re-prove `GC_PENDING` and absence of semantic/live references, live admission claims/bindings and backup protection.

Provider age/listing alone never authorizes deletion.

Bytes referenced by Submission, OfficialRendition or imported governed history have no ordinary Launch delete path, including when the business Revision later becomes SUPERSEDED, OBSOLETE or CANCELLED.

GC is technical reclamation, not Records disposition.

---

## 14. Backup set correctness — T4-L

A restorable recovery point consists of:

```text
one consistent product-state DB recovery point
+ all managed-content handles required by that DB snapshot
+ exact bytes for those handles
+ backup manifest of handle + expected ExactContentDescriptor
```

Required content includes at least:

```text
all immutable Submission content
all required OfficialRendition content
all imported governed content represented at that snapshot
current WorkingContent content for every open Revision
```

The backup manifest is recovery/operations metadata, never semantic product authority.

An in-progress backup must prevent selected live DRAFT content from becoming physically reclaimable before capture completes through a bounded backup pin/lease or equivalent GC exclusion. This is not business retention.

Because the selected durable-job mechanism stores required job intents in the same PostgreSQL product-state database, a DB recovery point is transactionally coherent between semantic facts and the durable intents committed with them: pre-snapshot semantic requirements and their intents restore together; post-snapshot facts and their intents are both absent. Restored pending work may re-run under T5 idempotency. A future move to a separate job substrate must re-prove this recovery coherence rather than assume it.

Provider-native backup/snapshot mechanisms may satisfy the contract if they prove the same complete set and descriptor integrity.

---

## 15. Restore exact-content and session readiness — T4-M

A restored deployment remains non-serving until every required semantic content reference proves:

```text
handle exists
exact bytes are readable
actual size == semantic size_bytes
SHA256(actual bytes) == semantic sha256
format is coherent with semantic content_format
```

Current DRAFT WorkingContent is included.

Unreferenced stale/reclaimable mechanism content may be absent without failing restore.

Any missing/corrupt required content:

```text
restore readiness = FAIL
serving = BLOCKED
```

Do not silently drop history, substitute another provider object or recalculate semantic truth from whatever bytes happen to be present.

**All ApplicationSessions restored from the recovery point are invalidated before ordinary serving resumes.** A user must establish a fresh post-restore session. Restore never treats a historical session row/token as proof of current authentication eligibility.

---

## 16. Privacy + security non-resurrection restore barrier — T4-N

Byte integrity and session invalidation are necessary but insufficient for a historical restore.

A restored recovery point may not enter ordinary authenticated serving mode until:

```text
lawful UserProfile erasures known after the recovery point are reconciled
AND
required known post-snapshot User offboardings / access teardown / security revocations
have been reconciled or otherwise proven safe for the restored serving state
```

Acceptable privacy shapes remain bounded:

```text
restore point at/after latest independently known erasure barrier
OR
replay/apply post-snapshot erasure facts from an independently retained recovery journal/control-plane source
```

For security teardown, T4 freezes the **readiness invariant**, not a generic journal design. T7/operations must choose the smallest recovery evidence/choreography that can prove and reapply the required post-snapshot offboardings/revocations for the chosen recovery model. This may reuse an independently retained control-plane/recovery source where appropriate; T4 does not require journaling every grant mutation as a new product subsystem.

If the chosen recovery proof cannot establish a safe current access state, ordinary authenticated serving remains fail-closed until reconciliation is completed. Non-serving recovery/maintenance operations may continue through an explicit operations trust surface; they do not become ordinary product access.

If completeness of the privacy erasure barrier/journal cannot be proven, restored human-readable profile data also fails closed for serving.

This is a narrow restore-safety mechanism, not a generic privacy/security workflow and not a reason to introduce mandatory per-user/application encryption.

T7/operations own concrete restore/cutover choreography; T4 owns the readiness invariant.

---

## 17. Future-evolution law — T4-O

Known future capabilities attach without restoring Artifact ownership:

```text
Distribution      → consumes released exact content
Periodic Review   → reads exact current EFFECTIVE content
Dossier           → references stable Document identity
Evidence          → may freeze its own ExactContentDescriptor + managed-content handle
Records           → attaches policy to semantic subjects; provider WORM/delete stays enforcement
Governed Export   → derives manifests/digests from semantic descriptors
Repository        → copies exact content; external identity never MetalDocs identity
Training/LMS      → consumes released/effective content
Change Control    → orchestrates stable Document/Revision identities
pooled tenancy    → may reopen isolation/profile, not exact-content ownership
CRDT              → may replace DRAFT collaboration, not immutable Submission boundary
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

No future capability requires a standalone Artifact semantic owner.

---

## 18. Why this is the Global Maximum rather than overengineering

Every accepted Launch mechanism has a named current consumer:

```text
SHA-256                  → exact submitted/released/rendition truth
size                     → cheap integrity + restore validation
ContentFormat            → governed format/view/render behavior
opaque handle            → retrieve bytes without provider identity becoming domain identity
create-once/no-overwrite → stable hash/malware/Submission exactness
OPEN→READY               → provider upload success is not semantic admission
admission binding        → authorize in-flight attachment without generic owner registry
malware gate             → untrusted bytes cannot become governed truth unsafely
restore verification     → Product Contract backup/restore correctness
session invalidation     → historical restore cannot reactivate bearer access
security reconciliation  → restored old access state cannot silently become current truth
```

Explicitly absent/deferred because no current consumer proves them:

```text
Artifact/ArtifactStage business ownership
whole-Submission JCS digest
multi-cloud/BYOS/active-active
replication/dedup/content-addressed semantic identity
Object Lock/WORM lifecycle
quarantine/CDR/rescan platform
PKI/signature infrastructure
confirmed governed-content delete/disposition Launch API
generic privacy-case platform/per-user crypto-erasure machinery
generic per-grant security teardown journal
```

---

## 19. Proof obligations before implementation

Implementation design/tests must falsifiably prove at least:

```text
client-provided hash/type never becomes authority
same managed handle cannot be overwritten through admitted paths
Local and S3 profiles pass the same core conformance suite
provider relocation does not change semantic descriptor
DRAFT autosaves do not create immutable Submission/history
READY untrusted bytes without CLEAN cannot become governed content in production
malware result applies to the exact immutable bytes later referenced
scanner failure cannot silently weaken production admission
SUBMIT rollback creates no semantic Submission despite successful prior upload
guessed handle cannot be attached to another root/operation
live admission claim/binding protects an in-flight READY handle from GC until bounded consume/release/expiry
WorkingContent OCC remains sole DRAFT race arbiter
GC cannot delete current WorkingContent or immutable governed content
stale cleanup intent cannot delete after eligibility/reference/claim changes
backup captures exact content required by its DB recovery point
backup/GC race cannot lose selected DRAFT content before capture
same-DB restore preserves coherence between committed semantic requirements and transaction-coupled durable intents
restore missing/corrupt required content fails closed
all restored ApplicationSessions are invalid before ordinary serving
historical restore cannot serve erased UserProfile PII without erasure reconciliation
historical restore cannot silently re-enable known post-snapshot offboarded/revoked access state
provider Object Lock/Versioning/checksum never becomes semantic authority
```

---

## 20. Explicit non-decisions

T4 does not decide:

```text
final SQL/table/index names
specific Go SDK/package structure
exact upload size limits
multipart implementation
exact ContentFormat detector/parser library
worker/lease/retry/DLQ topology for cleanup/rendering
Search technology
public upload/download API
frontend editor/viewer UX
Historical Migration batch/cutover implementation
exact recovery evidence/source for post-snapshot security teardown
future Records disposition/WORM policy
multi-store routing/BYOS/active-active
```

---

## 21. Reopen triggers

Reopen only the implicated T4 seam on material evidence that:

```text
supported files require multipart upload
production provider cannot prove create-once/no-overwrite + exact read
one deployment genuinely needs multiple simultaneously active stores
trusted derived output must also be malware scanned
regulation/customer requirement promotes WORM/retention/disposition
provider migration becomes live product/runtime capability
DRAFT recovery needs user-visible checkpoint/undo history
signature/non-repudiation needs canonical whole-Submission digest
backup provider cannot satisfy exact-content capture/GC exclusion economically
privacy/legal requirements demand stronger erasure than restore reconciliation
the selected recovery model cannot prove post-snapshot access teardown without a stronger dedicated mechanism
```

Implementation remains **BLOCKED** until the remaining R10 stages, integrated GCR, cold review and final operator ratification close.
