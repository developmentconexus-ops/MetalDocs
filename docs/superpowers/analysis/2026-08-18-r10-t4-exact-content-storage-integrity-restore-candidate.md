# R10-T4 — Exact Content, Storage Integrity & Restore — Reconciled Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **T3 authority:** `wiki/architecture/r10-t3-authorization-audit-enforcement.md`  
> **Implementation:** BLOCKED

T4 derives the smallest sustainable exact-content/storage/restore design from the ratified kernel. It deliberately reuses physical-safety discoveries from the old R10-C only where the Decision Registry preserves them and deliberately removes the old `Artifact` semantic ownership model.

T4 does not define final SQL/table/index syntax, Go package layout, public API/frontend routes, worker/retry topology or historical-migration execution.

---

# 1. T4 decision question

> **How does MetalDocs persist, retrieve, verify and restore the exact bytes owned/frozen by WorkingContent, Submission, OfficialRendition and later imported-history facts, while storage remains replaceable mechanism, untrusted content cannot enter governed truth unsafely, DRAFT autosave remains practical, and restore fails closed on missing/corrupt content or privacy resurrection?**

---

# 2. Registry baseline — not open for aesthetic redesign

T4 MUST consume:

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
provider-neutral storage/integrity seam is required
SHA-256 is a strong preserved exact-byte digest candidate
governed immutable bytes should not overwrite in place
production malware inspection before governed admission of untrusted bytes is a preserved candidate
Object Lock/WORM/versioning never becomes lifecycle authority
restore with missing/corrupt required bytes is not healthy
restore must not silently resurrect lawfully erased UserProfile PII
```

T4 MUST NOT revive:

```text
Artifact aggregate/owner
Artifact retention-root model
provider bucket/key/version as business identity
content-address hash as semantic entity id
confirmed orphan Artifact library
provider-specific business permissions
Object Lock/WORM as document lifecycle
multi-cloud/BYOS/active-active platform
confirmed-governed-content delete/disposition Launch API
```

---

# 3. Evidence reused from old R10-C

The old physical-integrity candidate exposed real failure modes that survive removal of Artifact ownership:

```text
provider PUT success != semantic admission
client hash/type != authoritative content facts
bytes scanned and then overwritten != safe admission
DB rollback after upload must not create governed history
every fast DRAFT autosave must not become immutable governed-history state
GC must never delete by provider age/listing alone
restore DB truth without matching bytes must fail closed
```

Useful external mechanism evidence remains:

- Amazon S3 supports conditional create/no-overwrite with `If-None-Match: *`, and bucket policy can require that condition;
- S3 provides object-integrity/checksum facilities and strong object read-after-write consistency;
- ClamAV `clamd` supports streamed scanning through `INSTREAM` and has an official container distribution.

Those are provider/mechanism facts only; they do not define MetalDocs semantics.

---

# 4. Credible architectural alternatives

## A — resurrect `Artifact` as exact-byte aggregate

```text
WorkingContent → Artifact
Submission → Artifact
Rendition → Artifact
```

**Reject — superseded by current GCR/T1.** It reintroduces an intermediate semantic owner whose meaning is already owned by WorkingContent/Submission/Rendition/imported facts.

## B — put provider key/path directly on every semantic record

```text
Submission.bucket
Submission.object_key
Submission.version_id
```

**Reject — mechanism becomes identity.** Provider migration/restore would rewrite semantic content identity or force provider-specific domain contracts.

## C — semantic exact descriptor + opaque technical content handle + provider-neutral store

```text
semantic owner
  exact descriptor
  opaque managed-content handle
        ↓
provider-neutral mechanism
        ↓
Local / S3 / future conforming provider
```

**Recommended Global Maximum.**

Semantic records own exact byte meaning. The handle exists only to retrieve those bytes. Provider location remains replaceable mechanism.

---

# 5. Exact content descriptor

T4 recommends one small semantic value shape wherever a semantic record owns/freezes exact bytes:

```text
ExactContentDescriptor
  sha256
  size_bytes
  content_format
```

Properties:

```text
SHA-256 = hash of exact raw bytes
size_bytes = exact raw-byte length
content_format = closed MetalDocs format vocabulary
```

`media_type`, filename, provider ETag/checksum, bucket/key/version and upload metadata are not exact-content identity. They may be derived/mechanism/provenance fields where a real consumer requires them.

The descriptor is copied/frozen on the semantic record that needs historical truth:

```text
WorkingContent      → current DRAFT descriptor
Submission          → immutable submitted descriptor
OfficialRendition   → immutable rendition descriptor
imported Revision   → later T7 exact imported-content descriptor
```

The same physical handle may appear on WorkingContent and its Submission, but each semantic record owns its own exact descriptor truth.

---

# 6. No mandatory whole-Submission canonical digest in Launch

Prior B3 proposed RFC8785/JCS + SHA-256 over an entire Submission manifest.

T4 recommends **not** making that a Launch requirement.

Reason:

```text
Submission already has stable identity
+
Submission fields/snapshots are immutable
+
exact source bytes have SHA-256 descriptor
+
governance/Release bind exact Submission id
```

A second canonical serialization/digest stack currently has no named Launch consumer and would create an additional compatibility contract.

Future digital-signature/non-repudiation/export requirements may define a versioned canonical package/digest over the already-preserved immutable facts without rewriting old Submission meaning.

Therefore:

```text
exact byte SHA-256 = ACCEPT baseline
whole Submission JCS digest = DEFER unless concrete consumer appears
```

---

# 7. Opaque managed-content handle — mechanism only

Semantic records may carry an opaque technical handle required to resolve bytes:

```text
managed_content_id UUID
```

Laws:

```text
managed_content_id != semantic content identity
managed_content_id != SHA-256
managed_content_id != provider key/version
provider migration may preserve same handle
semantic descriptor remains authority for exactness
```

The mechanism may maintain durable state needed for admission/cleanup, but that state is `DURABLE MECHANISM`, not a fifth business content owner.

Conceptual mechanism lifecycle:

```text
OPEN
→ READY
→ GC_PENDING
```

There is deliberately no semantic `CONFIRMED Artifact` state.

A READY managed-content object becomes preserved governed content **because an immutable semantic record references it**, not because the mechanism acquires business meaning.

---

# 8. One provider-neutral ManagedContentStore contract

Launch uses one active managed-content store per deployment.

Minimum provider-neutral properties:

```text
opaque UUID-addressed handle
create-once / no overwrite
exact full read
optional bounded/range read
copy existing authorized managed content to a new handle
explicit missing/failure
unreferenced DRAFT/stage cleanup
```

Conceptual mechanism operations:

```text
PresignCreate(handle, constraints)
Stat(handle)
OpenExact(handle)
OpenRange(handle, ...)
CopyToNewHandle(source, destination)
DeleteReclaimable(handle)
```

There is no ordinary Launch serving/domain `DeleteGovernedContent` capability.

One active store avoids a provider registry/replica-routing platform. A deployment needing multiple simultaneously active stores is a reopen trigger.

---

# 9. Provider profiles

## Local profile

Retain a real local provider for dev/test/conformance, not a fake in-memory storage path.

It must prove the same core contract:

```text
create-once
no overwrite
exact read
copy to distinct handle
explicit missing/failure
reclaimable-object cleanup
```

## AWS S3 reference production profile

Retain AWS S3 as the reference production profile, but the product invariant remains provider-neutral.

The reference profile must prove at least:

```text
TLS
private bucket/object posture
provider encryption at rest
conditional create/no overwrite
least-privilege read/write credentials
browser presign limited to one create handle
cleanup capability not exposed to browser/domain write path
exact server read for validation
```

S3 ETag/checksums/Versioning are supporting mechanism evidence only. MetalDocs SHA-256 remains semantic exact-byte authority. Versioning and Object Lock are not required Launch semantics.

A different production provider is allowed only when it satisfies the same conformance contract; no provider-specific semantic branch is created.

---

# 10. Upload / staging / READY validation

Browser/external ingress never writes semantic governed truth directly.

Conceptual path:

```text
allocate opaque managed_content_id
→ OPEN mechanism state
→ short-lived create-only direct upload
→ server independently reads exact uploaded bytes
→ derive authoritative descriptor
→ READY
```

Server derives:

```text
full SHA-256
size
actual ContentFormat
basic non-executing structural validity
```

Client filename, Content-Type and hash are hints only.

Once READY, admitted application paths cannot overwrite the physical bytes behind that handle. Replacement always creates a new handle.

Multipart upload orchestration is not a Launch baseline; reopen only when actual supported file-size evidence requires it.

---

# 11. Admission binding — prevent confused-deputy handle reuse

A READY upload handle must not be attachable to an arbitrary semantic root merely because a caller knows its UUID.

T4 retains the useful old R10-C concept as a **mechanism-only opaque admission binding**:

```text
managed content allocation
→ server binds handle to intended operation/root through an opaque unforgeable value
→ owning use case must present/re-prove that binding before attaching the handle
```

The mechanism does not parse or own `DocumentRevision`, `Submission`, Evidence or other business types. No generic `owner_type/owner_id` registry is introduced.

Legitimate cross-root copy/template-seeding flows create a new handle through an authorized server-side copy operation. Same-Revision resubmit may reuse the already-authorized handle when the content is unchanged.

---

# 12. Malware admission gate

T4 accepts the preserved safety property:

> **In production, untrusted external bytes cannot become immutable governed MetalDocs content without a successful malware inspection of those exact immutable bytes.**

Mechanism trust classes are server-selected, never client-selected:

```text
UNTRUSTED_EXTERNAL
TRUSTED_MANAGED_COPY
TRUSTED_INTERNAL_DERIVATION
```

Examples:

```text
browser upload                → UNTRUSTED_EXTERNAL
external repository import    → UNTRUSTED_EXTERNAL
historical migration file     → UNTRUSTED_EXTERNAL
authorized copy of already admitted managed bytes → TRUSTED_MANAGED_COPY
renderer-produced rendition   → TRUSTED_INTERNAL_DERIVATION
```

Production policy:

```text
UNTRUSTED_EXTERNAL → CLEAN required before immutable governed reference is committed
TRUSTED_MANAGED_COPY / TRUSTED_INTERNAL_DERIVATION → no mandatory rescan baseline
```

This means fast DRAFT autosaves may reach READY without scanning every debounce. The exact READY handle is scanned when it is about to cross a governed boundary such as SUBMIT or imported-content admission.

Scanner unavailable/incomplete → visible retriable failure; no semantic admission.
Malicious result → handle rejected/not attachable to immutable governed truth.

`MalwareInspector` remains a provider-neutral port. ClamAV/`clamd` is retained as the reference mechanism, not semantic authority.

Explicit dev/test profiles may disable malware inspection, but scanner-disabled data cannot silently be declared production-ready. Production promotion must re-enter the production admission proof.

---

# 13. DRAFT autosave without immutable-history explosion

A DRAFT autosave does not create immutable governed content history.

```text
allocate new handle
→ upload
→ READY descriptor
→ WorkingContent OCC transaction
     expected generation
     validate handle/admission binding
     copy descriptor + handle to WorkingContent
     generation++
→ previous unreferenced DRAFT handle becomes reclaimable candidate
```

WorkingContent remains the only current DRAFT authority.

No `WorkingSnapshot` business-history family is introduced.

Crash/reload recovery baseline is simply:

```text
reload current persisted WorkingContent
→ resolve its current handle
→ verify/use exact current DRAFT bytes
```

A richer undo/checkpoint/recovery history is added only if a real UX/recovery consumer proves it. `EditorSession` also remains optional UX mechanism for T6; OCC remains correctness authority.

---

# 14. SUBMIT / immutable semantic admission

When the current WorkingContent handle is about to become Submission truth:

```text
preflight exact malware scan when required
→ open local semantic transaction
→ serialize Document / WorkingContent under T2
→ prove expected WorkingContent generation
→ prove current handle is READY, immutable and admission-bound correctly
→ prove mechanism descriptor == WorkingContent descriptor
→ prove required malware result belongs to same exact handle/bytes
→ create immutable Submission copying exact descriptor + handle
→ freeze governance/representation snapshot
→ Revision DRAFT → SUBMITTED
→ required T3 Audit
→ COMMIT
```

No provider call occurs inside this local transaction.

If the transaction rolls back, the READY handle remains non-governed mechanism state and can be retried/reclaimed later. Provider upload success never creates Submission truth by itself.

No separate Artifact promotion is needed.

The same pattern applies to required OfficialRendition admission: renderer/provider work completes outside the semantic transaction; the final transaction revalidates the exact immutable READY handle and creates the semantic Rendition record.

Historical Migration uses the same eventual admission seam in T7 for untrusted exact imported files.

---

# 15. Physical immutability / no-overwrite law

Every admitted managed-content handle maps to create-once bytes.

```text
same handle + second PUT → rejected
DRAFT replacement        → new handle
semantic immutable record → descriptor + existing create-once handle
```

The normal browser/domain path never receives a capability that can overwrite an existing handle.

For S3, the reference profile uses conditional creation (`If-None-Match: *`) and may enforce that condition with bucket policy. Other providers must prove an equivalent create-once property.

This law makes malware results and SHA-256 verification stable: the scanned/hashed bytes cannot be changed afterward through admitted application paths.

---

# 16. Reclaimable DRAFT mechanism cleanup

Launch has no governed-content physical disposition.

Only mechanism objects that never became required immutable governed history and are no longer the current live DRAFT source may be reclaimed.

Eligibility fence:

```text
BEGIN
lock mechanism object
prove handle is not current WorkingContent
prove no immutable Submission/Rendition/imported semantic record references it
mark GC_PENDING
record cleanup intent only if T5 proves durable async cleanup is needed
COMMIT
```

Physical delete happens outside the semantic transaction and T5 owns worker/retry design.

Immediately before provider deletion, cleanup must re-read/re-prove `GC_PENDING` and absence of semantic/live references. Provider age/listing alone never authorizes deletion.

There is no Launch delete path for bytes referenced by Submission, OfficialRendition or imported governed history, including when the Revision later becomes SUPERSEDED, OBSOLETE or CANCELLED.

GC is mechanism reclamation, not Records disposition.

---

# 17. Backup set correctness

T4 defines backup correctness/readiness, not a custom backup product.

A backup capable of restoring a serving MetalDocs deployment must capture:

```text
one consistent product-state DB recovery point
+
all managed-content handles required by that DB snapshot
+
exact bytes for those handles
+
a backup content manifest containing handle + expected descriptor
```

Required handles include at least:

```text
all immutable Submission content
all required OfficialRendition content
all imported governed content represented at that snapshot
current WorkingContent handle for every open Revision
```

The backup content manifest is operations/recovery metadata, never semantic product authority.

Because DRAFT handles may later become reclaimable, an in-progress backup must prevent a handle selected from its DB snapshot from disappearing before the backup copy completes. The exact mechanism may be a short backup pin/lease or an equivalent bounded GC exclusion; it is not business retention.

Provider-native backup/snapshot tooling may satisfy this contract when it can prove the same complete set and exact descriptor integrity.

---

# 18. Restore readiness — fail closed

A restored deployment remains non-serving until verification proves every required semantic reference resolves to exact bytes.

For each semantic content reference in the restored DB:

```text
managed handle exists
exact bytes are readable
size == semantic descriptor.size_bytes
SHA256(bytes) == semantic descriptor.sha256
format is coherent with semantic descriptor.content_format
```

For current DRAFT WorkingContent, the restored handle must also exist and match its current descriptor.

Unreferenced stale/reclaimable mechanism objects may be absent without failing restore.

Any missing/corrupt required content:

```text
restore readiness = FAIL
serving = BLOCKED
```

Do not silently drop a history record, substitute another provider object, or recalculate semantic identity from whatever bytes happen to exist.

---

# 19. Privacy non-resurrection restore barrier

Byte-integrity success alone is not sufficient to serve a historical restore.

A DB backup taken before a lawful `UserProfile` erasure may still contain the erased profile. Therefore T4 establishes this restore law:

> **A restored recovery point may not enter serving mode until every lawful profile-erasure fact known after that recovery point has been reconciled onto the restored state.**

Acceptable operational shapes are bounded by this invariant:

```text
restore point at/after latest independently known erasure barrier
OR
replay/apply post-snapshot erasure facts from an independently retained recovery journal/control-plane source
```

If the deployment cannot prove the erasure barrier/journal is complete, serving restored human-readable profile data fails closed.

This is a narrow restore-safety mechanism, not a generic PrivacyCase platform and not a reason to introduce mandatory per-user/application encryption.

T7/operations later own the concrete recovery/cutover choreography; T4 owns the readiness invariant.

---

# 20. Future capability attack

| Future capability | T4 seam preserved |
|---|---|
| Distribution | released Submission/Rendition content resolves through stable descriptor+handle; Distribution does not own bytes |
| Periodic Review | reads exact current EFFECTIVE content without owning storage |
| Dossier | context references stable Document; no storage ownership |
| Evidence | future independent owner can freeze its own ExactContentDescriptor and reuse ManagedContentStore mechanism |
| Records/Hold/Disposition | can later add business disposition policy over semantic subjects; provider deletion/WORM remains enforcement only |
| Governed Export | can compute manifest/digests from stable semantic descriptors without exposing provider keys |
| Repository connector | import/publish copies use exact descriptors; external object identity never MetalDocs identity |
| Training/LMS | consumes effective content; no storage authority |
| Change Control | orchestration references Document/Revision; content mechanism unchanged |
| pooled tenancy | provider isolation/profile may reopen; semantic descriptors/handles remain provider-neutral |
| CRDT | replaces WorkingContent collaboration mechanics; immutable Submission/content descriptor boundary remains |

No future feature requires restoration of an Artifact semantic owner.

---

# 21. Proof obligations before implementation

Later implementation design/tests must falsifiably prove at least:

```text
client-provided hash/type never becomes authority
same managed handle cannot be overwritten through admitted paths
S3 reference profile enforces create-once/no-overwrite
provider relocation does not change semantic descriptor
DRAFT autosaves do not create immutable Submission/history
READY untrusted bytes without CLEAN cannot become immutable governed content in production
malware scan result applies to the exact immutable bytes later referenced
scanner unavailable cannot silently weaken production admission
SUBMIT rollback creates no semantic Submission even if provider upload succeeded
arbitrary guessed handle cannot be attached to another semantic root/operation
WorkingContent OCC remains sole DRAFT race arbiter
reclaimable GC cannot delete current WorkingContent or immutable governed content
GC worker stale intent cannot delete after reference/preservation state changes
Local and S3 profiles pass the same core store conformance
restore with missing/corrupt Submission/Rendition/imported/current-DRAFT bytes fails closed
backup captures exact handles required by its DB recovery point
backup/GC race cannot lose a selected live DRAFT handle before backup capture
historical restore cannot serve lawfully erased profile PII without erasure reconciliation
Object Lock/Versioning/provider checksum never becomes semantic lifecycle/content authority
```

---

# 22. Explicit non-decisions

T4 does not decide:

```text
final table names/SQL/indexes
specific Go SDK/client/package structure
exact upload size limits
multipart upload implementation
exact ContentFormat detector/parser library
worker/lease/retry/DLQ topology for cleanup/rendering
Search technology
public HTTP upload/download API
frontend editor/viewer UX
Historical Migration batch/cutover implementation
future Records disposition/WORM policy
multi-store routing/BYOS/active-active
```

---

# 23. Reopen triggers

Reopen the implicated T4 decision only on material evidence that:

```text
actual supported files require multipart upload
selected production provider cannot prove create-once/no-overwrite + exact read
one deployment genuinely needs multiple simultaneously active managed-content stores
trusted derived output must also be malware scanned
regulation/customer requirement promotes WORM/retention/disposition
provider migration becomes live product/runtime capability rather than operations event
DRAFT recovery requires user-visible checkpoint/undo history beyond current WorkingContent
future signature/non-repudiation requires canonical whole-Submission digest
backup provider cannot satisfy exact-handle manifest/pin semantics economically
privacy/legal requirements demand a stronger erasure mechanism than restore reconciliation
```

---

# 24. Operator adjudication packet

Recommended dispositions:

```text
T4-A ACCEPT — exact semantic byte descriptor = SHA-256 + exact size + closed ContentFormat on every semantic record that owns/freezes bytes.
T4-B ACCEPT — no mandatory whole-Submission RFC8785/JCS digest in Launch; immutable Submission fields + exact byte descriptor are sufficient until a named signing/export/non-repudiation consumer exists.
T4-C ACCEPT — semantic records carry an opaque managed-content handle for retrieval, but handle/provider identity is mechanism only and never semantic identity.
T4-D ACCEPT — one provider-neutral ManagedContentStore contract + one active store per deployment; Local is first-class dev/test/conformance and AWS S3 remains reference production profile.
T4-E ACCEPT — upload/admission uses OPEN→READY mechanism state; server independently derives SHA-256/size/format from exact bytes; client metadata is never authoritative.
T4-F ACCEPT — READY handle carries opaque server-owned admission binding so guessed/reused handles cannot cross semantic roots/operations; legitimate copy creates a new authorized handle except same-Revision unchanged reuse.
T4-G ACCEPT — production untrusted external bytes require CLEAN malware result on the exact immutable READY bytes before immutable governed semantic admission; autosave does not scan every debounce; ClamAV/clamd remains reference mechanism behind MalwareInspector port.
T4-H ACCEPT — every handle is create-once/no-overwrite; DRAFT replacement creates new handle; S3 reference uses/enforces conditional create; Versioning/ObjectLock remain optional mechanism evidence only.
T4-I ACCEPT — DRAFT autosave persists only current WorkingContent + READY handle under OCC; no WorkingSnapshot business history and no EditorSession correctness dependency.
T4-J ACCEPT — SUBMIT/Rendition semantic transaction performs no provider call; it revalidates pre-uploaded immutable READY handle, descriptor, admission binding and required malware proof, then freezes descriptor+handle on the semantic record. Rollback leaves reclaimable mechanism state only.
T4-K ACCEPT — only unreferenced/non-governed DRAFT mechanism objects may be reclaimed in Launch; GC requires DB eligibility fence + immediate recheck before provider delete; immutable governed content has no Launch delete path.
T4-L ACCEPT — every restorable backup pairs a consistent DB recovery point with a manifest/copy of every content handle required by that point, including current DRAFT WorkingContent; backup capture must exclude GC races through a bounded backup pin/equivalent mechanism.
T4-M ACCEPT — restored deployment stays non-serving until every required handle exists and exact size/SHA-256/format match semantic descriptors; missing/corrupt required content fails closed.
T4-N ACCEPT — restore also fails closed until post-snapshot lawful UserProfile erasures are reconciled through an independently retained erasure barrier/journal or equivalent recovery proof; no generic privacy workflow/mandatory crypto-erasure platform is introduced.
T4-O ACCEPT — future Evidence/Records/Export/Repository capabilities reuse semantic descriptors + shared managed-content mechanism without restoring Artifact ownership; future material requirements reopen only their implicated seam.
```

T4 remains non-authoritative until operator adjudication. After technical adjudication, **T5 still must not open**: the mandatory platform-facing T4 summary must be presented and explicitly ratified first.
