# Template docx key integrity & publish hardening (Phase 1)

- **Date:** 2026-06-29
- **Owning module:** `templates`
- **Status:** design approved (direction + spawn-content decision), pending written-spec review
- **Supersedes framing of:** `docs/superpowers/analysis/2026-06-29-blank-template-docx-provision-system-impact.md` (Yellow; the "provision empty bytes" premise is re-scoped to Phase 2)
- **Evidence base:** workflow `wf_7fcdc63c-b0d` (4-agent op-map + cited research + greenfield synthesis + adversarial critique)

## 1. Problem

A template version row's `docx_storage_key` is supposed to point at exactly one object, and that object is supposed to be that version's **own** immutable content. Three distinct issues break this — two are real damage, one is cosmetic:

| Issue | What it is | Severity |
|---|---|---|
| **Shared-key corruption** | `buildNextDraftVersion` spawns the next draft with `next.DocxStorageKey = published.DocxStorageKey` (lifecycle.go:437-448). Editing/autosaving the draft PUTs to the **published** object's key → overwrites an approved, immutable document. Affects the **live** Approve path (lifecycle.go:250) and the Publish path (lifecycle.go:399). | **Corruption** |
| **Publish key injection** | `PublishTemplateVersion` sets `version.DocxStorageKey = cmd.DocxKey` — client-supplied, never `assertTenant`-guarded, never validated (lifecycle.go:388; openapi.yaml:1279-1282 `required: [docx_key, schema_key]`). A crafted publish body can point a published version at another tenant's object; `PresignGet` is not tenant-guarded (verified_store.go:105). | **Security (multi-tenant invariant)** |
| **D2 blank-create 404** | Blank create commits a v1 row with a key but no object; `GetDocxURL` presigns a GET to a key with nothing behind it → 404 (queries.go:54-57). The FE already swallows this as "open empty" (useTemplateDraft.ts:52-56). | **Cosmetic** |

**This spec is Phase 1: close the corruption + the security hole.** D2 is deferred to Phase 2 (it corrupts nothing and is already absorbed harmlessly by the FE).

## 2. Why this design (global maximum, not a patch)

The cross-store consistency literature is unambiguous (workflow research, cited in §8):

- Atomicity across a DB row and an object store is impossible without 2PC, which this class of system rejects. The goal is a **recoverable asymmetry + reconciliation**, never a network PUT inside the state-write tx (Kleppmann; Richardson, transactional outbox).
- **Store-then-reference**: write the blob first, commit the reference second, so the only crash outcome is an orphan (safe, GC-able), never a dangling reference (Thom Wright).
- The server must **never synthesize docx bytes** — the vendored client engine `createEmptyDocument()` is the only producer (MetalDocsEditor.tsx). This is the hard constraint that rejects the "server writes an embedded blank `.docx`" approach (it makes the Go server a second docx producer, permanently coupled to the vendored binary format).

`CopyObject` (server-side object duplication) does **not** violate the producer invariant: it moves already-verified, client-produced bytes object-to-object; no docx bytes are authored by or stream through the Go server. This is the distinction that makes copy-on-spawn correct where an embedded-blank `Put` would be wrong.

## 3. Design

### 3.1 New storage primitive: `Copy`

Add to `objectstore.VerifiedStore` (internal/platform/objectstore/verified_store.go):

```go
// Copy duplicates an existing object to a new tenant-scoped key, server-side
// (no bytes stream through the app). The DESTINATION is tenant-prefix guarded;
// the source is a DB-sourced/server-trusted key (read path, not guarded).
func (s *VerifiedStore) Copy(ctx context.Context, tenantID, srcKey, dstKey string) error
```

- Wraps `minio.Client.CopyObject(ctx, dstOpts, srcOpts)`.
- `assertTenant(tenantID, dstKey)` on the destination (same guard as `PresignPut`).
- Add `Copy` to the templates `Presigner` port (ports.go:41-46) and to `fakePresigner` (fakes_test.go). Zero production wiring change — the concrete `*VerifiedStore` gains the method; `apps/api/.../main.go` already injects `*VerifiedStore`.

### 3.2 Copy-on-spawn: every spawned version owns a distinct key + its own object

A version that has a source (v2, v3, …) is born as a **byte-copy of its source at its own canonical key**. The very first blank v1 has no source — it stays blank (D2, Phase 2).

Change `buildNextDraftVersion` to take `tenantID` and key the draft at its own slot:

```go
// before: published.DocxStorageKey  (SHARED — the bug)
// after:  templateDocxKey(tenantID, templateID, nextNum)  (DISTINCT)
```

Both spawn paths perform the copy **pre-tx** (network call kept out of the state-write tx, per the outbox rule), then commit the row:

1. compute `nextNum` (see §3.3)
2. `dstKey := templateDocxKey(tenantID, templateID, nextNum)`
3. `if err := s.presign.Copy(ctx, tenantID, source.DocxStorageKey, dstKey); err != nil { return err }` — **abort the spawn before opening the tx** if the copy fails (store-then-reference: object exists before the referencing row commits)
4. build the draft with `dstKey`; **leave `ContentHash` empty** so the publish gate (`ContentHash != ""`, lifecycle.go:350) still forces a real edit before the new revision can publish
5. commit the row in the existing tx

Apply at:
- **Approve** accept branch (lifecycle.go:250) — the live path.
- **PublishTemplateVersion** (lifecycle.go:399).
- **CreateNextVersion** (create.go:140-150) — currently fresh key but no copy → opens blank; route it through the same copy so explicit new-version also starts from source content. Removes the divergence between the two spawn families.

Crash between Copy and row-commit → orphan object (unreferenced, safe, GC-able by the Phase 2 reaper). Never a dangling reference.

### 3.3 Unify `nextNum` allocation

The three paths diverge: Approve uses `max(LatestVersion+1, VersionNumber+1)` (lifecycle.go:246-248); Publish and CreateNextVersion use `LatestVersion+1` (lifecycle.go:398, create.go:140). Extract one helper so all paths allocate identically.

The existing `UNIQUE(template_id, version_number)` (db/baseline:3015) already makes concurrent same-`nextNum` inserts hard-fail today — this change neither introduces nor worsens that hazard. Unification is for correctness/consistency, not a new constraint. (Full serialization/retry of concurrent spawns is a pre-existing concern, explicitly out of scope here.)

### 3.4 Close the publish key-injection hole

- Remove `docx_key` (and assess `schema_key`) from the `/templates/{id}/versions/{n}/publish` request body in `api/openapi/v1/openapi.yaml:1277-1282`; regenerate via `oapi-codegen` (contract-first — no hand-edited routes).
- `PublishTemplateVersion` keeps the version's **own canonical** `DocxStorageKey` (the value it was born with) instead of `cmd.DocxKey` — i.e. delete lifecycle.go:388. This matches `Approve`, which never re-keys on publish.
- Drop `DocxKey` from `PublishTemplateVersionCmd` (lifecycle.go:320-326) and the handler wiring (routes_generated.go:192).
- **No FE cutover:** `canActOnVersion.ts:87` documents that the editor screen does not call `/publish`; the FE publishes via the Approve path.

### 3.5 DB last line of defense

After the data migration (§4), add `UNIQUE(docx_storage_key)` on `public.templates_template_version`. With de-shared positional keys this is consistent with the existing `(template_id, version_number)` unique, and it makes a future re-introduction of shared keys structurally impossible ("DB enforces invariants").

## 4. Data migration

Existing rows may already share keys (spawned drafts pointing at a published key). The `UNIQUE(docx_storage_key)` add will fail unless those are resolved first. One forward migration:

1. **Re-key** every non-published row whose `docx_storage_key` equals another row's key to its canonical `tenants/{tenant_id}/templates/{template_id}/versions/{version_number}.docx`.
2. **Object copy** (companion step, using the new `Copy` primitive or `mc cp`): copy the published object to each re-keyed draft's new key so existing in-flight drafts retain their content. If a draft's object already diverges from the published one (it was edited — meaning corruption already occurred), flag it for manual review rather than silently copying.
3. Add `UNIQUE(docx_storage_key)`.
4. **Audit** published rows for non-canonical keys (artifacts of the §3.4 injection path) before trusting the constraint.

Check row counts first — the feature is recent, so the affected set is expected to be small or empty; if empty, steps 1–2 are no-ops and only the constraint is added.

## 5. Invariants satisfied

- **Multi-tenant pooled:** `Copy` guards the destination prefix; removing the client-supplied publish key closes a cross-tenant pointer hole.
- **Async = no network in state-write tx:** `Copy` runs pre-tx, exactly like the existing `GetTemplateByKey` check and `CommitAutosave`'s `Confirm`.
- **Contract-first:** publish body change goes through openapi + `oapi-codegen`.
- **DB enforces invariants:** `UNIQUE(docx_storage_key)` is the last line; app de-sharing is the friendly first line.
- **Producer invariant:** `Copy` moves client-produced bytes; the server still synthesizes no docx.
- **Capabilities, not roles:** no authz change; existing `CapTemplateEdit/Approve/Publish` tier-2 checks are untouched.

## 6. Testing (canonical framework per test-discipline)

- **Unit (service):** spawned draft gets a distinct canonical key; `Copy` invoked with `(published key → new key)` before the tx; `ContentHash` left empty; Copy failure aborts the spawn with no row written. Use the existing `fakePresigner` (extended with `Copy`).
- **Integration (testdb factory):** publish/approve → next draft has its own key and its own object; editing the draft does not mutate the published object; `UNIQUE(docx_storage_key)` rejects a duplicate.
- **Contract:** `/publish` no longer accepts `docx_key`; published version keeps its canonical key.
- **Verification (runtime, `verify` skill):** drive Approve-publish in the running app, confirm the published docx is byte-stable after the spawned draft is edited.

## 7. Out of scope — Phase 2 (deferred)

D2 read-path honesty for the origin blank v1: explicit `provisioning` lifecycle state, `GetDocxURL` returning typed not-ready instead of a URL-to-nothing, and a level-triggered reconciler/janitor (with grace tuned to the **max PUT→Confirm wall-clock gap**, not the presign TTL). Before any content-addressed-storage work, fix `documents/service.go:332-337` (`content_hash = sha256(docxKey)` welds document identity to the template key string). Tracked separately; none of it is required to close the corruption or the security hole.

## 8. References

- Store-then-reference (orphan-vs-dangling asymmetry): https://thomwright.co.uk/failure-patterns/store-then-reference/
- Transactional outbox / dual-write: https://microservices.io/patterns/data/transactional-outbox.html ; https://martin.kleppmann.com/2015/05/27/logs-for-data-infrastructure.html
- Model "exists but no blob yet" as a first-class state (Google Docs create→batchUpdate): https://developers.google.com/workspace/docs/api/reference/rest/v1/documents/create
- Level-triggered reconciliation (Phase 2): https://book.kubebuilder.io/reference/good-practices.html
- Orphan-blob GC (Phase 2): https://learn.microsoft.com/en-us/previous-versions/office/developer/sharepoint-2010/bb862262(v=office.14)
