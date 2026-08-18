# R10-C — Artifact Physical Integrity — Integrated Candidate

> **Status:** NON-AUTHORITATIVE — SELF-REVIEWED CORRECTED CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input baseline:** R10-B complete/non-final + operator-approved Launch V1 scope rebaseline  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This is staging analysis. It does not independently ratify R10-C. It records one bounded B3 correction exposed by a real physical-integrity/autosave counterexample and otherwise preserves accepted R10-B ownership/lifecycle semantics.

---

# 1. Authority and evidence boundary

Authority order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. frozen R9.5 ledger
6. promoted R10 technical authority through B2
7. accepted non-final B3/B4/B5/B6 candidates + acceptance records
8. `wiki/architecture/launch-v1-scope-rebaseline.md` — current Launch overlay

Current code/module docs remain current-state evidence only.

Directed external evidence only:

- AWS S3 conditional create/no-overwrite: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/conditional-writes.html>
- AWS S3 conditional-write enforcement: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/conditional-writes-enforce.html>
- AWS S3 checksum/integrity support: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html>
- ClamAV `clamd` streaming protocol: <https://docs.clamav.net/manual/Usage/ClamdProtocol.html>
- official ClamAV container guidance: <https://docs.clamav.net/manual/Installing/Docker.html>

Useful current-state evidence:

- editor autosave is approximately every 1.5 seconds;
- browser bytes already go directly to object storage using presigned upload;
- server re-reads uploaded bytes and computes SHA-256 before accepting an autosave;
- provider upload can succeed while later DB work fails, leaving physical orphan data.

The target preserves the proven direct-upload/server-verification shape while removing legacy MinIO/tenant-key/current-table semantics as authority.

---

# 2. Launch boundary

Launch invariant from the active scope overlay:

> **Confirmed governed history is preserved and Launch V1 exposes no governed physical disposition. SUPERSEDED, OBSOLETE, CANCELLED and VOIDED never mean delete. Draft/temporary state that never became preserved governed history may be reclaimed.**

Not Launch C:

```text
RetentionBinding / retention clock
LegalHold
DispositionFence / DispositionRecord
Records-driven Object Lock/WORM
confirmed governed-history delete
multi-cloud/BYOS/active-active
permanent dual-write
runtime provider-migration product workflow
content-addressed Artifact identity
```

---

# 3. Root cause / target invariant

Failure class: **physical mechanism becoming semantic truth**.

Bad states:

```text
provider PUT success == Artifact confirmed
provider key/path == Artifact identity
client hash/type == authoritative facts
bytes scanned, then overwritten before confirmation
DB rollback but upload is treated as governed
restore DB says Artifact exists while bytes are missing/corrupt
every ~1.5s DRAFT autosave becomes preserved Artifact history
```

Target invariant:

> **A confirmed Artifact exists only after MetalDocs independently verifies one immutable physical candidate and the candidate is promoted, inside the same PostgreSQL transaction as its first typed semantic ownership relation, into immutable Artifact facts. Mutable DRAFT bytes remain reclaimable pre-confirmation state. Provider location is mechanism only. Every confirmed Artifact and every live draft candidate required by product state must resolve to exact bytes matching MetalDocs SHA-256/size facts.**

```text
ArtifactStage != Artifact
upload success != confirmation
DRAFT autosave != immutable governed history
provider checksum != canonical authority
GC != Records disposition
```

---

# 4. Global Maximum decision

## A — confirmed Artifact per autosave

Reject. With autosave around 1.5 seconds, either browser-produced content is malware-scanned at that frequency or an admitted browser upload path bypasses the frozen production confirmation gate. It also creates unnecessary immutable Artifact churn.

## B — storage-binding/replica/quarantine platform

Reject. `ArtifactStorageBinding`, provider registry, replicas, quarantine aggregate, multi-cloud migration and ObjectLock policy machinery have no Launch consumer.

## C — pre-confirmation ArtifactStage + Artifact only at immutable governed boundary

```text
ArtifactStage
→ READY mutable-draft candidate
→ required malware gate at governed confirmation
→ Artifact promotion on SUBMIT / CAPTURE / imported-content / Rendition confirmation
```

One active store maps MetalDocs UUIDs to opaque physical objects. Superseded draft stages are reclaimable; confirmed governed history is preserved.

**Recommended Global Maximum.**

---

# 5. B3-R7 bounded correction — DRAFT bytes use ArtifactStage

Accepted B3 used `WorkingContent.primary_artifact_id` while DRAFT. R10-C exposes a real scaling/safety counterexample.

Corrected current-R10 candidate:

```text
WorkingContent
  revision_id
  current_stage_id UUID NULL FK ArtifactStage
  current_artifact_id UUID NULL FK Artifact
  governed_metadata
  structured_authoring?
  working_version

CHECK exactly one current byte source
```

Rules:

- ordinary autosave/replacement produces a server-validated READY Stage and OCC-swaps WorkingContent to it;
- a same-Revision confirmed Artifact may remain the current DRAFT seed after return-for-changes when no later edit occurred;
- any later edit creates a new Stage;
- cross-Revision/new-document/template seeding creates a new Stage for the new target root; it does not reuse another semantic root's Artifact row;
- SUBMIT of a Stage performs the final required confirmation gate, promotes that same Stage UUID to Artifact, switches WorkingContent to the Artifact, and creates RevisionSubmission atomically;
- same-Revision resubmit with unchanged already-confirmed Artifact may reuse that Artifact.

Native `EvidenceDraft` likewise points to `ArtifactStage`; CAPTURE promotes exact current Stage to Artifact atomically with EvidenceCapture.

This changes only mutable DRAFT byte representation. Document/Revision/OCC/Submission/Approval/Release semantics remain unchanged.

---

# 6. ArtifactStage — Artifact-owned pre-confirmation state

R10-A already requires that Artifact not import CI/Documentary Context and that confirmation use an opaque owner reference. Therefore ArtifactStage has **no FK back to DocumentRevision or Evidence**.

```text
ArtifactStage
  id UUID PRIMARY KEY                 // becomes Artifact.id if promoted

  semantic_root_binding BYTEA NOT NULL
    CHECK octet_length(semantic_root_binding)=32

  origin_kind TEXT NOT NULL CHECK
    UNTRUSTED_EXTERNAL |
    CONFIRMED_ARTIFACT_COPY |
    TRUSTED_INTERNAL_DERIVATION

  expected_content_format TEXT NOT NULL CHECK closed ContentFormat

  state TEXT NOT NULL CHECK
    OPEN | READY | REJECTED | GC_PENDING

  sha256 BYTEA NULL CHECK NULL OR octet_length(sha256)=32
  size_bytes BIGINT NULL CHECK NULL OR size_bytes >= 0
  detected_content_format TEXT NULL
  detected_media_type TEXT NULL

  malware_status TEXT NOT NULL CHECK
    NOT_SCANNED | NOT_REQUIRED | CLEAN

  scanner_name TEXT NULL
  scanner_version TEXT NULL
  scanner_definitions_version TEXT NULL
  scanned_at TIMESTAMPTZ NULL

  created_at TIMESTAMPTZ NOT NULL
  expires_at TIMESTAMPTZ NOT NULL
  ready_at TIMESTAMPTZ NULL
  rejected_at TIMESTAMPTZ NULL
  rejection_code TEXT NULL
```

`semantic_root_binding` is a caller-supplied **opaque domain-separated digest** of the candidate semantic root. Artifact stores/compares it but does not parse/import CI/DC tables. Final Artifact root remains derived from the closed typed semantic-reference catalog, not from this Stage field.

Properties:

```text
OPEN       → no authoritative byte facts yet
READY      → SHA/size/detected format/media fixed
REJECTED   → never promotable
GC_PENDING → never promotable
```

After READY, root/origin/byte facts are immutable. Malware facts may advance only according to the closed confirmation policy.

The opaque root binding prevents one Stage from being attached/promoted under a different DocumentRevision/Evidence root without creating a polymorphic owner registry or a reverse business FK.

---

# 7. ManagedArtifactStore — smallest physical contract

One active store per deployment. Artifact/provider location mapping is derivable by the adapter from MetalDocs UUID; no Launch per-Artifact location table.

Core properties:

```text
UUID-addressed opaque object identity
create-once / no overwrite
exact full read
optional bounded/range read
exact confirmed-Artifact → new Stage copy
explicit missing/failure
stage cleanup
```

Conceptual consumer operations:

```text
PresignCreate(stageID, constraints)
Stat(id)
OpenExact(id)
OpenRange(id,...)
CopyArtifactToStage(sourceArtifactID,destinationStageID)
```

Cleanup receives a narrower capability:

```text
DeleteStage(stageID)
```

There is deliberately no Launch serving/domain `DeleteArtifact` operation.

Provider-private key/path/bucket/version is absent from Artifact and Stage DB state. If a future deployment must simultaneously address confirmed Artifacts across multiple active stores, that real requirement reopens a durable physical-binding model.

---

# 8. Upload → READY

## Allocate

Owning composition creates Stage with expected format, opaque root binding and **server-selected** origin kind. External clients cannot request trusted origin classes.

## Direct upload

Normal browser/external ingress uses a short-lived one-object create URL.

S3 reference profile uses create-if-absent conditional write (`If-None-Match: *`) and may enforce the condition at bucket policy. Replacement is a new Stage UUID; admitted upload paths cannot overwrite an existing Stage/Artifact object.

Launch baseline is bounded one-shot upload; multipart orchestration is deferred until actual file-size evidence requires it.

## READY validation

After upload completion signal, server derives from exact bytes:

```text
full SHA-256
size
actual ContentFormat
media type
basic non-executing structural validity
```

Client filename/content-type/hash are hints only.

`OPEN → READY` freezes these byte facts. Full SHA-256 remains MetalDocs canonical byte identity; provider checksums may support but never replace that authority.

---

# 9. Malware gate without scanning every autosave

Reference mechanism:

```text
MalwareInspector port
→ ClamAV clamd reference implementation
```

Production confirmation policy:

```text
UNTRUSTED_EXTERNAL          → CLEAN required
CONFIRMED_ARTIFACT_COPY     → NOT_REQUIRED
TRUSTED_INTERNAL_DERIVATION → NOT_REQUIRED
```

`UNTRUSTED_EXTERNAL` includes direct browser bytes, External Repository imports and Historical Migration exact-file ingress.

READY DRAFT autosaves are not confirmed governed Artifacts, so they do not require a malware scan on every debounce. When the exact Stage is about to cross SUBMIT/CAPTURE/import confirmation, scanner unavailable/incomplete means visible retriable failure; malicious result means REJECTED and never promotable.

The frozen launch claim remains exactly: untrusted bytes cannot become confirmed Artifact in production without CLEAN.

Dev/test may explicitly disable scanner. Scanner-disabled data is non-production data; Launch does not support silently switching such a dataset into production readiness. A future promotion/migration must re-enter the production confirmation proof rather than changing a flag.

Bounded scanner/version/definitions/time evidence is copied into Artifact technical provenance on promotion. No quarantine/CDR/rescan/security workflow platform.

---

# 10. Promotion — all-path confirmation guard

R10-C must not rely on “developers normally call the right function.”

Target enforcement:

- ordinary serving write trust has no unrestricted direct Artifact insertion path;
- Artifact exposes one narrow transactionally composable `PromoteStage` seam (DB-owned primitive or equivalently strong all-path control);
- the seam locks Stage and requires READY, matching opaque root binding, exact byte facts and required CLEAN/NOT_REQUIRED policy;
- it inserts Artifact with the same UUID and consumes Stage within the caller transaction;
- B5's closed typed-reference catalog + deferred DB enforcement remains the final committed-state backstop:

```text
committed Artifact with zero semantic refs   → fail
refs resolving to >1 semantic root           → fail
```

Therefore a caller cannot commit a confirmed Artifact merely by writing an Artifact row, nor can promotion commit without a typed semantic owner by transaction end.

Promotion transaction examples:

### SUBMIT

```text
preflight immutable Stage malware inspection if required
→ lock WorkingContent / CAS generation
→ lock Stage
→ re-prove Stage is current source + root binding
→ PromoteStage
→ WorkingContent current source = Artifact
→ RevisionSubmission exact Artifact
→ B4 requirement snapshots / Approval init
→ required intents
→ AuditChainHead last
→ COMMIT
```

### Evidence CAPTURE

```text
current EvidenceDraft Stage
→ required malware gate
→ PromoteStage
→ EvidenceCapture Artifact
→ Evidence CAPTURED
→ delete EvidenceDraft
→ Audit last
```

### Rendition

Trusted renderer output Stage is hash/format validated and promoted atomically with Rendition. Provider write alone never means Rendition success.

### Historical Migration

Untrusted exact file passes the same Stage/CLEAN gate before Artifact + imported target-owner content commit in the migration semantic unit.

No confirmed Artifact exists if the DB transaction rolls back; the Stage remains retry/cleanup state.

---

# 11. Physical immutability / delete trust boundary

Once a Stage is READY, every **admitted application delete** must first cross DB state `GC_PENDING`. Promotion and GC serialize on the same Stage row.

The normal serving API has no object-delete credential/capability for confirmed governed content. A separate cleanup execution surface may delete Stage objects only after DB eligibility.

This is not a claim that a hostile cloud/account administrator can never destroy bytes. Operator/provider corruption is detected by integrity/restore verification; Launch does not invent WORM/HSM infrastructure without requirement.

---

# 12. DRAFT autosave and superseded-stage cleanup

```text
allocate S2 for same semantic root
→ direct upload
→ READY
→ WorkingContent OCC transaction
     expected working_version
     assert Stage root binding
     current source = S2
     working_version++
→ prior Stage, when no longer live, becomes cleanup candidate
```

A prior confirmed Artifact used as same-Revision return-to-DRAFT seed is preserved because it may already be pinned by immutable Submission history.

Superseded Stage is not governed immutable history. Cleanup is technical state reclamation, not Records disposition.

---

# 13. GC race law

GC never deletes from object-store age/listing alone.

Fence:

```text
BEGIN
  lock ArtifactStage
  prove Stage is not current in any admitted live Stage consumer
  prove no Artifact with same UUID exists
  Stage → GC_PENDING
  insert cleanup intent
  Audit only if later census says this mechanism transition is audit-worthy
COMMIT
```

R10-D worker must **re-read/re-prove GC_PENDING and no same-id Artifact immediately before physical DeleteStage**. A stale/malformed cleanup intent cannot delete by ID without that DB proof.

After physical absence is confirmed:

```text
BEGIN
  lock GC_PENDING Stage
  re-prove no same-id Artifact
  delete Stage row
COMMIT
```

Any operation attaching/promoting a Stage requires READY, so GC_PENDING is an irreversible technical fence.

---

# 14. Local + AWS S3 profiles

## Local

First-class conformance provider, not fake storage. Must prove:

```text
UUID mapping
atomic create-if-absent/no overwrite
exact read
copy to distinct Stage id
stage cleanup
explicit missing/failure
```

Unsupported provider features remain unsupported.

## AWS S3 reference production

Core application invariant remains provider-neutral. Reference production should prove:

```text
TLS
provider encryption at rest
private bucket/object posture
conditional create/no overwrite
least-privilege read/write credentials
browser presign limited to one Stage create
cleanup credential not exposed to browser/serving domain
exact server read for validation
```

S3 ETag/checksums/optional Versioning are mechanism evidence only. Versioning is not an application dependency. Object Lock/WORM is outside Launch after Records-Governance defer.

---

# 15. Restore integrity

R10-C defines readiness property, not a custom backup platform.

A restored deployment remains non-serving until:

```text
for every Artifact:
  derived physical object exists
  size matches
  SHA256(exact bytes) == Artifact.sha256

for every live Stage referenced by WorkingContent/EvidenceDraft/other admitted live candidate:
  Stage row exists
  physical object exists
  if READY: size/SHA/format match Stage facts
```

Unreferenced stale Stage loss is cleanup state, not restore failure.

Privacy/erasure reconciliation is an additional later R10-F readiness gate. Hash success alone cannot put a historical restore into serving mode.

No Tenant-DEK/crypto-shred platform is introduced.

---

# 16. Current-state coherence benefit

Current legacy evidence already proves direct browser upload + server-side hash is operationally useful, but also shows the editor may create byte snapshots at roughly 1.5-second cadence. The Stage model preserves the direct-upload scaling advantage while moving immutable Artifact confirmation to the actual governed boundary.

This is a simplification, not a second storage system:

```text
DRAFT exact bytes  → ArtifactStage
immutable governed exact bytes → Artifact
same ManagedArtifactStore underneath
```

---

# 17. Cross-stage transaction consequences

Launch B6 matrix changes only where R10-C exposes the physical candidate:

### autosave

```text
external Stage upload/READY validation
→ WorkingContent OCC swap
→ old-Stage cleanup intent when applicable
```

Ordinary autosave is not semantic Audit by default.

### SUBMIT / CAPTURE / imported content / Rendition

Artifact promotion occurs inside the owning semantic transaction before final AuditChainHead acquisition.

No Records-Governance locks/intents exist after the Launch rebaseline.

Global law remains:

```text
owner/config locks
→ Stage/Artifact promotion
→ durable real-effect intents
→ AuditChainHead LAST
→ COMMIT
```

---

# 18. Proof strategy before implementation

Required falsification tests:

1. same Stage object cannot be overwritten through admitted upload paths;
2. client lies about hash/type → server-derived facts win/reject;
3. production untrusted Stage without CLEAN cannot become Artifact through **any** admitted write path;
4. repeated DRAFT autosaves create Stage/OCC state, not immutable Artifact history;
5. upload succeeds + semantic transaction rolls back → no Artifact row, Stage remains retry/cleanup state;
6. Stage opaque root binding mismatch rejects cross-root attach/promotion;
7. direct Artifact insert/bypass is rejected by DB/grant/control;
8. committed Artifact orphan or multi-root reference is rejected by deferred all-path guard;
9. GC-vs-promotion race yields exactly one winner; GC worker cannot delete a promoted same-ID Artifact;
10. corrupted/missing Artifact or live READY Stage prevents restore readiness;
11. Local and S3 pass the same core conformance suite;
12. scanner-disabled deployment cannot present as production-ready;
13. admitted B1→C write lock graph remains acyclic.

---

# 19. Explicit Launch non-goals

```text
ArtifactStorageBinding table
replica/ref-count authority
multipart orchestration
quarantine aggregate
periodic malware rescans
CDR/sandbox platform
GuardDuty-specific domain state
mandatory S3 Versioning
Object Lock/WORM
runtime provider relocation workflow
multi-cloud/BYOS/active-active
confirmed governed Artifact delete API
content-addressed Artifact identity
hash-every-range-viewer request
custom backup platform
```

---

# 20. Reopen triggers

Reopen only on material evidence:

- actual file sizes require multipart;
- one deployment must simultaneously address confirmed Artifacts across multiple stores;
- regulatory/customer requirement brings WORM/retention/disposition back;
- trusted derived outputs prove they require malware inspection;
- selected provider cannot satisfy create-once/no-overwrite + exact read;
- runtime provider relocation becomes a real operational consumer;
- implementation proves Stage/WorkingContent return-to-DRAFT semantics do not preserve the invariant;
- implementation proves Stage/GC/Audit lock graph cannot be linearized without a cycle.

---

# 21. Candidate status

If accepted:

```text
R10-B = COMPLETE / NON-FINAL / LAUNCH-SCOPE-REBASELINED
R10-C = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED
         including B3-R7 DRAFT ArtifactStage correction
R10-D = NEXT / SIMPLIFIED DESIGN ONLY
implementation = BLOCKED
```

Whole-R10 Global Coherence Review + cold independent review remain required before final ratification.