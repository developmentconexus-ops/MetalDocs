# System-impact analysis — Blank-template docx object provisioning

**Date:** 2026-06-29
**Intent (one line):** A blank template's committed version row points at a `versions/1.docx` object that does not exist until the first autosave — close the dangling-object integrity gap (provision an empty `.docx` via the outbox, OR formalize the lazy-provision contract with a guard/test). Also clean up orphaned MinIO objects (`a5e1be9f*`, `ef374718*`) from a prior dev DB rebuild.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

---

## 1. Classify & own

- **Work type:** feature (hardening of an existing create/read path; no new module).
- **Owning module(s):** `templates` — owns template versions, the docx storage key (`templateDocxKey`, `keys.go:8`), the create tx (`create.go:30`), the autosave write path (`autosave.go`), and the `docx-url` read path (`queries.go:46`).
- **Explicitly NOT owning:**
  - `render` / `docx-renderer` — reads only *published* template bytes; never touches a blank draft's key.
  - `documents` / `controlleddocuments` — snapshot published `placeholder_schema`; no dependency on the blank object.
  - `objectstore` platform pkg — provides the store capability; would gain a method under Option A but does not *own* the feature.
- **Cross-module edges (with direction):**
  - `templates → objectstore (platform)` — via the `application.Presigner` port (`ports.go:41`). Option A adds a server-side `Put`; this is a platform primitive, not a module's internals. ✅ in-bounds.
  - `templates → worker (platform)` — Option A only. New event type + runner registered in `apps/worker/cmd/metaldocs-worker/main.go`; dispatch via `internal/platform/worker/service.go`. ✅ published wiring pattern.
- **Ambiguity?** None. No AS-3.

## 2. Foundation verdict

- **Base you'd build on:** client-side presigned upload. The backend never writes docx bytes — the browser PUTs to a presigned URL, then `CommitAutosave` (`autosave.go:110`) verifies via `Confirm` (SHA-256). On create, the version row is committed with `DocxStorageKey` set but **no object behind it**. `GetDocxURL` (`queries.go:46`) hands out a presigned GET URL **without an existence check** (`PresignGet` only signs; MinIO never validates the key), so the endpoint returns `200 + URL-to-nowhere`; the 404 surfaces only when the browser fetches it.
- **Sound, or legacy/patch/workaround?** The presigned-upload design is **sound** (deliberate, avoids multi-MB docx through the API). The genuine defect is narrow: **the read path emits a presigned URL for an object it has not confirmed exists.** That is the dangling reference, not the lazy upload itself.
- **Global-maximum framing.** Two structurally different fixes:
  - **Option A — provision empty bytes via outbox.** Requires (a) a *new* server-side write capability on `VerifiedStore` (none exists today), (b) templates' *first* transactional-outbox usage (new table + allowlist + migration + tripwire review), (c) a new worker event type + `TemplateDocxJobRunner` + main.go wiring. Heavy async machinery whose only product is writing a constant empty `.docx`. **Risk: this is a local maximum dressed as global** — it builds a pipeline to paper over a read path that simply should not emit unverified URLs.
  - **Option B — make the read path honest + pin the contract.** `GetDocxURL` already has the modeled empty-state: `domain.ErrUploadMissing` → `problem.CodeUploadMissing` (`errors.go:52`), which the FE already tolerates (empty editor, by design). Gate the URL behind `store.Exists()` (already implemented, `verified_store.go:113`): when the object is absent, return the empty-state contract instead of a broken URL. Document "lazy-provision-on-first-autosave" as intended; add a guard/test pinning it. **Cheaper, removes the broken URL at the source, adds no async surface.**
- **Recommendation:** Option B is the global maximum here; Option A is over-engineering for a blank file with no server-side consumer. Decision belongs to the operator (task framed A as primary) — surfaced before design. No AS-2 (we are not optimizing *inside* a patch; the base is sound).

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | No (read path already `template.view`; create already `CapTemplateCreate` in-tx) | unchanged | `authz.Require` (already wired `create.go:67`) |
| Contract-first (OpenAPI + oapi-codegen) | **Maybe (B)** — if `docx-url` response shape gains an explicit empty-state flag, edit `api/openapi/v1/partials/templates.yaml` + regen. If we reuse existing `ErrUploadMissing`/409 shape, **no contract change.** | spec-first, regenerate | `api/cfg.yaml` + `go generate` |
| Multi-tenant pooled | Yes | key already tenant-scoped (`tenants/{tid}/...`); `Exists`/`Put` must keep the `assertTenant` prefix guard (`verified_store.go:46`) | `tenant.FromContext`, `assertTenant` |
| Async = transactional outbox | **Option A only** — provisioning is an external side effect; MUST enqueue in the create tx and let an idempotent consumer Put. **Must NOT** call objectstore inline in the create commit tx. | outbox enqueue in-tx + idempotent consumer | `staging_outbox.go:52` Enqueue pattern; worker runner pattern |
| DB enforces invariants | Option A — new outbox table needs `tenant_id` + the standard outbox columns/constraints; tripwire allowlist review. | migration + constraints | baseline schema conventions |
| Cross-module via published interface only | Yes — extend the `Presigner` port (A) / use existing `Exists` (B); worker wiring via published registration. No reaching into another module. | port extension | `application/ports.go:41` |

No invariant **violated** by either option. AS-1: none. (Option A is *constrained by* the outbox invariant, not in violation of it.)

## 4. Capability wiring

**N/A** — no new IAM capability. Create uses `CapTemplateCreate`, read uses `template.view`; both already wired. A worker consumer (Option A) runs out-of-band and is not capability-gated at the HTTP edge.

## 5. Module wiring

**N/A** — no new module. (Option A adds a worker *consumer* to an existing binary, not a bounded-context module.)

## 6. Frameworks to reuse, not reinvent

- `TxRunner.Do` — already owns the create tx (`create.go:63`); Option A enqueue goes inside it. ✅
- `tenant.FromContext` / `assertTenant` — tenant guard on any new store method. ✅
- Outbox repo (`staging_outbox.go`) — Option A reuses the `StagingOutboxRepository` shape + factory; **do not hand-roll** a templates outbox. ✅
- Worker `Service` dispatch + `.WithXxxRunner()` registration — Option A reuses this pattern (`internal/platform/worker/service.go`, `apps/worker/.../main.go`). ✅
- `problem.New`/`Write` + existing `CodeUploadMissing` — Option B reuses the modeled empty-state; no new error code. ✅
- `VerifiedStore.Exists` — Option B reuses (already implemented, currently unused on this path). ✅
- `testdb` factory — for any integration guard test. ✅

No genuinely-new cross-cutting concern. (Option A's only new primitive is a server-side `Put` on `VerifiedStore` — a legitimate platform extension, not a one-off.)

## 7. Contract & data

- **OpenAPI-first:** Option B with reused 409/`ErrUploadMissing` = **no contract change**. Option B with an explicit `{ empty: true }` response, or Option A's behavior, may warrant a doc note but no new route. Any response-shape change → `partials/templates.yaml` + regen.
- **Migration:** Option A only — `db/migrations/0NNN_templates_docx_outbox.sql` (table + `tenant_id` + outbox columns + indexes; tripwire allowlist). Option B — **no migration.**
- **Destructive change?** None. Read-path behavior tightens (URL→empty-state) but the FE already tolerates the empty state, so no contract break.
- **Orphan cleanup** (both options): delete MinIO objects under the templates prefix with keys starting `a5e1be9f` / `ef374718` if still present (dev-DB-rebuild leftovers). One-time `Delete` via store/`mc`; not a code change. Verify presence first; report what was removed.

## 8. Test & QA plan

- **Canonical framework:** unit — existing `application/*_test.go` + `fakes_test.go` fakes (extend the `Presigner` fake with `Exists`/`Put` as needed). Integration — `tests/docx_v2/templates_integration_test.go` or a `testdb`-backed guard, `//go:build integration`.
- **Guard/test (the pin):**
  - Option B: a test asserting (i) a freshly-created blank template's version has `DocxStorageKey` set, (ii) `GetDocxURL` returns the empty-state contract (`ErrUploadMissing`/409) — **not** a presigned URL — while the object is absent, and (iii) after a `Confirm`/autosave the same call returns a URL.
  - Option A: consumer idempotency test (Put-if-absent; re-run no-ops) + outbox enqueue-in-tx test.
- **QA gates that apply (feature subset):** contract (if response shape changes), multi-tenant isolation (key prefix guard on the new store method), async/idempotency (Option A consumer), DB-invariant (Option A migration). Authz gate **N/A** (unchanged). Docs gate applies.
- **Evidence shape:** `go build ./...`, `go test ./...`, `.\scripts\check-system-runnable.ps1`; runtime verify against the running app (create blank template → open before autosave → confirm `docx-url` resolves to the documented empty-state or a valid object, no broken 404). Capture network/logs.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/templates.md` (record the empty-state read-path contract; refresh `Last verified`) and `wiki/modules/templates-tech-debt.md` (close/append the dangling-object item). Option A also documents the new outbox + worker consumer.
- **REQ IDs cited:** the async-outbox REQ (REQ-ASYNC-*) governs Option A's machinery; the multi-tenant blob-key REQ governs the store-method tenant guard. (Pull exact IDs from `wiki/architecture/backend-target-architecture.md` at design time.)
- **ADR required?**
  - Option B: **no ADR** — in-bounds read-path correctness, reuses a modeled contract.
  - Option A: **likely yes (Yellow)** — introduces templates' first transactional outbox + a new server-side object-store write capability; that is a standing-design addition worth an ADR (or at minimum an explicit note that it follows the outbox invariant). This is the flagged risk carried into design.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to design, but the approach choice (A vs B) materially changes scope (~10×) and Option A carries an ADR flag. No hard-stop.
- **Open hard-stops:** AS-1 none · AS-2 none (base is sound) · AS-3 none.
- **Locked constraints handed to brainstorming / design:**
  1. Owning module is `templates`; no new IAM capability; no new bounded-context module.
  2. **If Option A:** provisioning is an external side effect → enqueue in the create tx, Put from an **idempotent** worker consumer; **never** call the object store inline in the commit tx (outbox invariant). New outbox table needs `tenant_id` + tripwire allowlist review; new `VerifiedStore.Put` must keep the `assertTenant` tenant-prefix guard. Likely ADR.
  3. **If Option B (recommended):** gate `GetDocxURL` on `store.Exists()`; return the existing `ErrUploadMissing`/`CodeUploadMissing` empty-state instead of a presigned URL-to-nowhere; document the lazy-provision-on-first-autosave contract; add the guard/test in §8. No migration, no new async surface, reuses modeled error.
  4. Both: orphan MinIO cleanup (`a5e1be9f*`, `ef374718*`) is a one-time op gated on presence-check; report what was removed.
  5. Evidence-before-closure: build + tests + runtime verify against the running app before claiming done.
