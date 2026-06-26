# Design Spec — Storage Referential Integrity (Workstream B)

- **Date:** 2026-06-26 (amended after adversarial review)
- **Mission:** [Document Contract & Storage Integrity](2026-06-26-mission-document-integrity.md)
- **Workstream:** B of 2 (sequenced after [Contract Integrity](2026-06-26-contract-integrity-design.md))
- **Status:** Approved design, amended — pending implementation plan

---

## 1. Problem

A persisted `storage_key` is a **declared intent, not a receipt of existence**.

`internal/modules/templates/application/create.go:54` sets
`docx_storage_key = fmt.Sprintf("templates/%s/versions/1.docx", id)` **eagerly at
row-create**. The object bytes are written **lazily** on first autosave — or never.
Between create and first verified write, the DB advertises a docx that does not
exist.

`queries.go:54` (`GetDocxURL`) presigns GET after only a **non-empty** check
(`if v.DocxStorageKey == "" { return ErrUploadMissing }`) — it does **not** Stat
that the object exists. (Correction vs the first draft of this spec, which said it
had no check at all.) So the residual read-side risk is precisely "key non-empty,
object never written."

Metadata-row write and object write are decoupled on several paths; the key is a
derived constant, never reconciled against storage reality.

## 2. The invariant we want

> A persisted, non-null `storage_key` always points to an object that exists, and
> `content_hash != ''` is its proof-of-verified-commit.

## 3. Evidence — a solved, standard pattern (and already in-repo)

Canonical solution: **two-phase pending → confirmed upload**. DB row starts pending
with the *intended* key; client uploads via presigned URL; a **server-side confirm**
(HEAD/Stat + hash) flips it to confirmed; only confirmed pointers are read; a janitor
reconciles orphans. S3 now gives strong read-after-write consistency, so a HEAD right
after upload is reliable.

Named implementations: **Rails ActiveStorage** (direct-upload `Blob`, attached only
after confirmation, orphan blobs purged by a scheduled job); **AWS reference
architecture** (store DB metadata only after `s3:ObjectCreated`); GitHub / Slack /
Stripe file uploads; Cloudinary / Uploadcare; Uppy + tus.

Sources: AWS S3 Event Notifications (https://aws.amazon.com/blogs/aws/s3-event-notification/);
AWS S3 strong consistency (https://aws.amazon.com/s3/consistency/);
two-phase presigned upload (https://brightinventions.pl/blog/efficient-S3-file-uploads-with-async-processing/).

**Decisive fact — verified in code:** the **templates** `CommitAutosave` path already
does a true server-side confirm: `autosave.go:156 HeadContentHash` →
`templates_presigner.go:47-62` does `GetObject` + `Stat` (returns `ErrUploadMissing`
on NoSuchKey) AND streams+hashes the bytes against `ExpectedContentHash`. The confirm
phase exists; templates simply also writes an eager key elsewhere. The **documents**
path uses the same discipline (`repository.go:1094-1116 CommitUpload`).

## 4. Decision

**Write-verified pointer, unified on the confirm pattern, plus a read-side guard as
defense-in-depth.** Approved 2026-06-26. Makes the invariant structurally true,
reuses an in-repo pattern, matches AWS + ActiveStorage. "Read-guard only" was
rejected (leaves a known-lie in the DB).

## 5. Scope

### 5.1 Templates: one verified writer, nullable-until-confirmed
- Stop setting `docx_storage_key` eagerly at `create.go:54` (and the v(n) bump at
  `lifecycle.go:145`). A fresh version has **no committed docx**.
- `docx_storage_key` column becomes **nullable**.
- **Go-type cascade (Critical, reviewer):** the domain field `version.go:23
  DocxStorageKey string` and the scan at `mappers.go:67` (`Scan(&v.DocxStorageKey)`
  into a `string`) will **error on NULL** (`converting NULL to string`), breaking
  `GetVersion` — which runs at the head of nearly every templates op. Therefore:
  change the domain field to `*string` (or `sql.NullString`) and **audit ~20 read
  sites** (`GetVersion`, `GetPublishedVersion` `templates_reader.go:33`, presign,
  submit, publish, queries) for null-handling. This is a code cascade, not a
  one-line column change.
- The verified `CommitAutosave` (§3) becomes the writer of a non-null key.

### 5.2 Delete the second, unverified writer (Major, reviewer)
- `SaveTemplateDraft` (`autosave.go:98-141`, route `PUT .../versions/{n}/draft`,
  `handler.go:50`) writes `docx_storage_key` + client-supplied `content_hash` via
  `UpdateVersionDraftCASTx` with **no** Stat — a live path that can set a non-empty
  hash for a non-existent object, falsely satisfying the publish gate (§5.4). It is
  the storage-side twin of Crack A's "two coexisting styles."
- **Decision (approved 2026-06-26): DELETE it.** Verified dead — no frontend caller
  (editor uses `presign`+`commit`), the `schema_storage_key`/`schema_content_hash`
  fields it carries exist nowhere else (schema saves go through `PUT .../schema`),
  and its repo methods have no other caller. Full removal scope:
  - **OpenAPI:** remove the `PUT /templates/{id}/versions/{n}/draft` path; drop the
    `.../draft` references in the `lock_version` description; `go generate ./...` to
    regenerate `api.gen.go` (drops the handler iface, `SaveTemplateDraftJSONBody`,
    and the `schema_*` request fields).
  - **Delivery:** `routes_generated.go:164 SaveTemplateDraft` + `missingSaveTemplateDraftField`;
    route mount `handler.go:50`.
  - **Application:** `SaveTemplateDraft` + `SaveTemplateDraftCmd` (`autosave.go:88-141`).
  - **Repository:** `UpdateVersionDraftCAS` + `UpdateVersionDraftCASTx`
    (`postgres.go:433,437`) and the two `ports.go:26-27` interface methods (no other
    caller).
  - **Tests:** `TestSaveTemplateDraft_*` (`autosave_test.go:298-367`) and the
    `UpdateVersionDraftCAS*` fakes (`fakes_test.go`, `routes_create_test.go`).
  - **Frontend:** the `saveDraft` shim (`templates.ts:220`) — already slated for
    deletion in Workstream A (it only forwards to `commitAutosave`).
  - Low deploy risk: removing an unused endpoint; forward-only contract change, no
    client depends on it (mission §3).

### 5.3 Document instance create/clone path (Critical, reviewer — elevated into scope)
- `service.go:264 cloneIntoTx` reads the template's published key
  (`GetPublishedVersion`, scans into non-nullable `string` `templates_reader.go:33`)
  and passes it verbatim into the new document's `storage_key` (`:305`) with **no
  existence check**. This propagates the eager-vs-lazy gap *transitively to
  documents*, and seeds `body_docx_s3_key` (the render/fanout input).
  **Decision:** the clone may only inherit a key whose source version is verified
  (`content_hash != ''`); `GetPublishedVersion` must handle the nullable key; publish
  (§5.4) guarantees a published version's key is non-null+verified, which makes the
  clone safe by construction.

### 5.4 Publish-time guard (Major, reviewer)
- Publish currently gates on `content_hash == ''` (`lifecycle.go:35-37,350-352`),
  never on the key. Make the coupling explicit and tested: **a version may publish
  only with a non-null, verified key** (`content_hash != ''` ⟹ key set at a verified
  commit). A null key with non-empty hash must be impossible. This is what makes §5.3
  sound.

### 5.5 Read-side defense (fail-closed, typed)
- Before any presign GET that can hit an unverified/absent object (`GetDocxURL`,
  document export), Stat/HEAD and fail closed with a **typed not-ready/not-found
  response** (struct, satisfies `cilint noresponsemap`). Its shape is declared in
  Workstream A's contract (mission §2 coupling). Update the existing non-empty check
  to also handle null.

### 5.6 Frontend
- Render a null / not-yet-committed docx (fresh version) as an explicit
  **empty-editor** state — intentional and visible, never an error or accidental
  blank. (`useTemplateDraft` already guards; make the empty state deliberate.)

### 5.7 Stated exceptions (document, don't silently allow)
- `buildNextDraftVersion` (`lifecycle.go:437-448`) carries the *published* key into
  the new v(n+1) draft with `content_hash=''`. Under the "null until commit" model
  this is an **intentional passthrough** of an already-existing, verified object —
  document it as a reasoned exception, not a violation.

## 6. Non-goals (explicitly deferred, not silently dropped)
- **PDF write hardening:** `WritePDF` (`snapshot_repository.go:154-165`) is a
  standalone UPDATE, no tx, no existence pre-check — the same lie on the output side.
  Deferred, **but** the defer carries a caveat: PDF reads must fail-closed (verify
  before serving) until hardened. Revisit if evidence of dangling PDF pointers
  appears.
- **Orphan-reconciliation janitor / lifecycle expiry** for abandoned pending uploads.
  Add only if orphan accumulation is observed.
- Materialize/freeze path is already clean (in-tx Pin + outbox + atomic
  `WriteFinalDocx`) — out of scope.
- **Tenant-prefixed keys:** template keys lack a tenant segment
  (`templates/{id}/versions/{n}.docx`); UUIDs prevent collision but per-tenant
  lifecycle/quota is weaker. Noted, not in scope.

## 7. Acceptance criteria
1. A freshly created template version has a **null** `docx_storage_key` until a
   verified autosave commit; `GetVersion` and all ~20 readers handle null without
   error (the Go-type cascade is resolved).
2. No persisted non-null key can reference a non-existent object: the only writer of
   a non-null key is the Stat-verified commit (§5.1); the `/draft` second writer is
   fully deleted (§5.2) — endpoint, handler, service, repo methods, and tests gone,
   `go build`/`go test` green without them.
3. Publish is gated such that a published version always has a non-null verified key
   (§5.4); document clone inherits only verified keys (§5.3).
4. Presign GET on a missing/unverified object fails closed with a typed response
   (§5.5), never a raw 404 leaking to a deref.
5. Fresh-version editor shows an intentional empty state (§5.6).
6. Backfill nulls `docx_storage_key` where `content_hash = ''` (DB-only, no S3 scan);
   migration is forward-only (mission §3).
7. `go build ./...`, `go test ./...`, and frontend tests green.

## 8. Migration / backfill
- Add nullable migration for `docx_storage_key`.
- **Backfill pivots on the in-DB signal:** null the key where `content_hash = ''`
  (never committed); keep it where `content_hash != ''` (verified). **No per-row S3
  Stat sweep** (the first draft's "Stat every row" plan was unsound and unbounded).
- Forward-only; no down-migration to NOT NULL (mission §3).

## 9. Risks
- The Go-type cascade (§5.1) is the largest under-estimated item; treat the
  read-site audit as a first-class plan task, not a cleanup.
- §5.3 elevates the document-clone path into scope; if plan effort balloons, it may
  split into its own milestone — but it cannot be silently deferred, as it shares the
  exact bug class the mission exists to kill.

## 10. Evidence / references
- `create.go:54`, `lifecycle.go:145,35-37,350-352,437-448`; `queries.go:54`
- `autosave.go:98-141,143-,156`; `templates_presigner.go:47-62`;
  `repository/postgres.go:433-460`; `mappers.go:67`; `version.go:23`
- `documents/repository/repository.go:1094-1116`; `documents/application/service.go:264,305`;
  `templates_reader.go:33`; `snapshot_repository.go:154-165`; `render/fanout/client.go`
- ADR 0015 (`wiki/decisions/0015-async-freeze-pin-materialize.md`); new ADR to author (mission §5).
