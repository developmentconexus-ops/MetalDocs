# Design Spec — Storage Referential Integrity (Workstream B)

- **Date:** 2026-06-26
- **Mission:** Document Contract & Storage Integrity (Grade-A)
- **Workstream:** B of 2 (sequenced after [Contract Integrity](2026-06-26-contract-integrity-design.md))
- **Status:** Approved design — pending implementation plan

---

## 1. Problem

A persisted `storage_key` is a **declared intent, not a receipt of existence**.

`internal/modules/templates/application/create.go:54` sets
`docx_storage_key = fmt.Sprintf("templates/%s/versions/1.docx", id)` **eagerly at
row-create**. The object bytes are written **lazily** on the first autosave — or
never. Between create and first autosave the DB advertises a docx that does not
exist, so a presign GET against it 404s. `internal/modules/templates/application/queries.go:54`
presigns GET for that key without first checking existence.

Metadata-row write and object write are decoupled on several paths; the key is a
derived constant, never reconciled against storage reality.

## 2. The invariant we want

> A persisted `storage_key` always points to an object that exists.

## 3. Evidence — this is a solved, standard pattern (and already in-repo)

The canonical industry solution is the **two-phase pending → confirmed upload**:

1. DB row starts in a *pending* state holding the *intended* key. Client uploads
   directly to object storage via a presigned URL.
2. A **confirm** step — server-side, not client-trusted — flips the row to
   *available* after verifying the object (HEAD/Stat + ETag/hash). The rest of the
   system only ever reads confirmed pointers. S3 now gives strong read-after-write
   consistency, so a HEAD immediately after upload is reliable.
3. A janitor / lifecycle rule reconciles orphans (uploaded-but-unconfirmed, or
   dangling pointers).

Named implementations: **Rails ActiveStorage** (direct-upload `Blob`, attached only
after confirmation, orphan blobs purged by a scheduled job); **AWS reference
architecture** (store DB metadata only after the `s3:ObjectCreated` event);
GitHub / Slack / Stripe file uploads; Cloudinary / Uploadcare; Uppy + tus.

Sources:
- AWS — S3 Event Notifications: https://aws.amazon.com/blogs/aws/s3-event-notification/
- AWS — S3 strong read-after-write consistency: https://aws.amazon.com/s3/consistency/
- Two-phase presigned upload write-ups: https://brightinventions.pl/blog/efficient-S3-file-uploads-with-async-processing/ ; https://medium.com/@Games24x7Tech/a-complete-guide-to-s3-file-upload-using-pre-signed-post-urls-9cb2d6cfc0ab

**Decisive fact:** MetalDocs already implements this on the **documents** path —
`autosave_pending_uploads` (pending, intended key) → `CommitUpload` after
server-side hash verification (confirm, records the real `storage_key`),
`internal/modules/documents/repository/repository.go:919-926,1094-1116`. The
**templates** path simply never adopted it.

## 4. Decision

**Write-verified pointer, unified on the documents pattern, plus a read-side guard
as defense-in-depth.** Approved 2026-06-26.

Rationale: makes the invariant structurally *true* (not merely handled-gracefully-
when-false); reuses a pattern already proven in this repo (no new invention or
dependency); matches AWS + ActiveStorage. Option "read-guard only" was rejected —
it leaves a known-lie in the DB, which no reference architecture endorses.

## 5. Scope

### 5.1 Templates (adopt pending → confirmed)
- Stop setting `docx_storage_key` eagerly at `create.go:54`.
- `docx_storage_key` column becomes **nullable** — semantics: "no docx committed
  yet" for a freshly created version.
- The key is recorded/confirmed only at the autosave commit, which **already**
  hash-verifies the uploaded object (templates presigner does a Stat in
  `HeadContentHash`; `updateVersionDraftCAS` persists under optimistic lock). The
  confirm phase therefore already exists on this path — the fix is to stop writing
  the eager key and let commit be the sole writer of a verified key.

### 5.2 Read-side defense (fail-closed)
- Before any presign GET that can hit an unverified key (notably the document
  export / docx-url path), Stat/HEAD the object and fail-closed (typed
  not-found / not-ready) if absent. Cheap belt-and-suspenders; the column should
  never be wrong after 5.1.

### 5.3 Frontend
- Render a nullable / not-yet-committed docx (fresh version, no bytes) as an
  explicit **empty-editor** state — never an error or crash. (Today
  `useTemplateDraft` already guards `if (res.ok)`; make the empty state intentional
  and visible, not an accidental blank.)

## 6. Non-goals (explicitly deferred, not silently dropped)
- **PDF write hardening:** `WritePDF` is fire-and-forget (standalone UPDATE, no tx,
  no existence pre-check) — `internal/modules/documents/repository/snapshot_repository.go:154-165`.
  A real but separate output-side consistency smell. Defer unless evidence demands.
- **Orphan-reconciliation janitor / lifecycle expiry** for abandoned pending
  uploads. Add only if orphan accumulation is observed. (Documents' pending table
  is the natural place to sweep.)
- The materialize/freeze path is already clean (in-tx Pin + outbox + atomic
  `WriteFinalDocx`) — out of scope.

## 7. Acceptance criteria
1. No persisted `storage_key` / `docx_storage_key` can reference a non-existent
   object: a freshly created template version has a null key until a verified
   autosave commit.
2. Presign GET on a missing/unverified object fails closed with a typed response,
   never a raw 404 leaking to a deref.
3. Fresh-version editor shows an intentional empty state, not an error.
4. `go build ./...`, `go test ./...`, and frontend tests green.
5. The templates path and documents path share the same pending → confirmed
   discipline (no divergent storage-write conventions remain).

## 8. Risks
- DB column nullability change needs a migration; existing rows with eager keys may
  point at objects that do exist (post-autosave) or not (never-saved). Plan a
  backfill/audit step: keys with a verified object stay; unverified keys null out.
- Sequencing: depends on Workstream A only insofar as both touch the templates
  surface; B can proceed once A's frontend templates calls are typed.

## 9. Evidence / references
- `internal/modules/templates/application/create.go:54` (eager key)
- `internal/modules/templates/application/queries.go:54` (presign GET, no existence check)
- `internal/modules/documents/repository/repository.go:919-926,1094-1116` (the in-repo pending → confirm pattern to mirror)
- `internal/modules/documents/repository/snapshot_repository.go:154-165` (deferred PDF write smell)
- ADR 0015 — async freeze/pin/materialize (`wiki/decisions/0015-async-freeze-pin-materialize.md`)
