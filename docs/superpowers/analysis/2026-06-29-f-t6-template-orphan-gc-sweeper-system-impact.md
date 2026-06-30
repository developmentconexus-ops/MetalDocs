# System-impact analysis — F-T6(a): template docx orphan GC sweeper

**Date:** 2026-06-29
**Intent (one line):** Add a reconciliation janitor that deletes orphaned template docx objects (left in the object store by `spawnNextDraft`'s pre-tx Copy when the surrounding tx rolls back) older than 24h.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

---

## 1. Classify & own

- **Work type:** feature (a new background janitor goroutine + one repo query + object-store reconciliation; no new module).
- **Owning module(s):** `templates` — owns `spawnNextDraft`, the `docx_storage_key` column, and the template object prefix. The sweeper belongs in a **new `internal/modules/templates/jobs/` subpackage** mirroring `internal/modules/documents/jobs/`.
- **Explicitly NOT owning:** the central `internal/modules/jobs` module — runtime truth: the canonical sibling `StartOrphanPendingSweeper` is **not** a River job there; it is a goroutine sweeper in `internal/modules/documents/jobs/` started from `apps/api/cmd/metaldocs-api/main.go:592-593`. The audit's loose "em internal/modules/jobs" is superseded by the sibling's actual location (CLAUDE.md runtime-truth rule). `metaldocs-jobs` (River) hosts *queue* jobs (scheduler, idempotency_janitor, watchdog) — not prefix-reconciliation sweepers.
- **Cross-module edges (with direction):** `templates → iam/authz` (background-bypass context, already a templates dependency). `templates → objectstore kernel` (the `VerifiedStore`/presign port templates already uses for Copy). No new cross-module edge; no repo/SQL reach-in.
- **Ambiguity?** Initially yes (audit said central `jobs`); resolved by targeted-verify of the sibling's location and wiring. **AS-3 raised and resolved** → owner is `templates/jobs`, wired in API main.

## 2. Foundation verdict

- **Base you'd build on:** the canonical sibling `documents/jobs/orphan_pending_sweeper.go` — a ticker goroutine under `authz.WithBackgroundBypass`, calling a repo cleanup method, started/stopped from API main. That base is sound and proven.
- **Sound, or legacy/patch/workaround?** Sound for the *lifecycle/wiring* shape. **But there is a real foundation gap:** the documents sibling deletes **DB rows** (`DeleteExpiredPending`). F-T6 must reconcile the **object store** (delete blobs that have *no* DB row) — there is **no drop-in sibling** for object-store prefix reconciliation, and the objectstore kernel (`VerifiedStore`) may not yet expose a `List(prefix)` primitive.
- **Global-maximum structure + trade-off:** the reconciliation must be built on the **hardened objectstore kernel from the F-O wave** (tenant-namespaced prefix List + Delete with the tenant-guard helpers), not on the legacy presign client. Building it on the pre-F-O object store would lock in a local maximum (an unguarded List/Delete). **Trade-off / sequencing constraint:** F-T6 must land **after** the F-O objectstore kernel helpers (esp. tenant-guarded List/Delete) so it consumes the validated primitive. AS-2 not triggered (we are *not* optimizing inside a patch — we are deferring until the correct primitive exists, then mirroring the sibling's lifecycle).

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | YES (background) | Runs under `authz.WithBackgroundBypass` (fail-closed off any HTTP path), exactly as the documents sibling (`orphan_pending_sweeper.go:15`). No role reasoning. | `authz.WithBackgroundBypass` |
| Contract-first (OpenAPI + oapi-codegen) | No | No HTTP route; pure background process. | — |
| Multi-tenant pooled | **YES** | Must reconcile **per tenant** with tenant-namespaced prefixes; never list/delete across tenant boundary. Delete only keys under the correct tenant prefix. | tenant-namespaced key builder; F-O tenant-guard helpers |
| Async = transactional outbox | No (read-then-delete reconciliation, not a state-write + network side effect in one tx) | The sweeper reads DB keys, lists object store, deletes orphans — no business-tx coupling a network call. It is a janitor, not an outbox producer. Object delete is the janitor's whole purpose, done outside any business tx. | — |
| DB enforces invariants | Partial | The orphan exists *because* there is no DB row (rollback). The 0250 `UNIQUE(docx_storage_key)` already prevents concurrent-spawn collisions; the sweeper is cleanup, not enforcement. DB membership query is the source of truth for "referenced". | 0250 UNIQUE (existing) |
| Cross-module via published interface only | YES (already) | templates uses its own repo + the objectstore port it already holds; no reach into another module's tables. | templates repo; objectstore port |

No invariant violated. AS-1 not triggered. **Locked:** per-tenant prefix scoping is mandatory (multi-tenant).

## 4. Capability wiring

**N/A** — adds no capability. Background janitor authorized via `authz.WithBackgroundBypass` (the established sibling pattern), not a new cap.

## 5. Module wiring

**N/A** — no new module. New `internal/modules/templates/jobs/` subpackage is an internal package under an existing module (mirrors `documents/jobs/`), not a bounded-context birth.

## 6. Frameworks to reuse, not reinvent

- **Sibling lifecycle:** `documents/jobs/orphan_pending_sweeper.go` — copy its ticker/`context.WithCancel`/`WithBackgroundBypass`/`stop func()` shape exactly. Do **not** hand-roll a different scheduler.
- **`authz.WithBackgroundBypass`** — background root auth. Reuse.
- **Objectstore kernel (`VerifiedStore` + F-O tenant-guard helpers)** — for List(prefix)+Delete. Reuse the hardened primitive; do **not** inline raw S3/MinIO calls. *(If List does not exist, it is added to the objectstore kernel in the F-O wave, not ad-hoc here.)*
- **`TxRunner.DoReadOnly`/`Do`** — for the DB membership query (list known `docx_storage_key`s per tenant). Reuse.
- **`slog`** — structured logging exactly as the sibling (`orphan_pending_sweeper deleted`).

## 7. Contract & data

- **OpenAPI-first:** N/A (no route).
- **Migration:** none required. Optionally a repo read method (`ListTemplateDocxKeys(ctx, tenantID)` or an orphan-candidate query) — a new query, not a schema change. The `docx_storage_key` column and 0250 UNIQUE already exist.
- **Destructive change?** The sweeper **deletes object-store blobs** — destructive by nature. Guard rails (locked): (a) only delete keys **older than 24h** (mtime/age threshold, matching the audit), (b) only under the correct **tenant prefix**, (c) only keys **absent** from the DB `docx_storage_key` set, (d) log every deletion. Age threshold prevents racing a Copy that is mid-flight before its tx commits.

## 8. Test & QA plan

- **Canonical framework:** integration test under `tests/integration/testdb/` (`//go:build integration`) with a fake/in-memory object store: seed (1) a referenced key (DB row present) → must survive; (2) an orphan key >24h, no DB row → must be deleted; (3) an orphan key <24h → must survive; (4) a key under a *different* tenant prefix → must never be touched. Mirror the documents `jobs_test.go` style.
- **QA gates that apply:** multi-tenant isolation (cross-tenant non-deletion), async/janitor idempotency (re-run deletes nothing new), DB-invariant (membership query correctness). Contract/authz-route gates **N/A**.
- **Evidence shape:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...` for the new sweeper test, `.\scripts\check-system-runnable.ps1` (startup wires the new goroutine). Report outcomes + two-stage review + the sequencing defer (after F-O).

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/templates.md` (lifecycle / object-store section) noting the orphan-GC janitor + refresh `Last verified`; add the deferred-then-done item to `wiki/modules/templates-tech-debt.md` (or remove it from tech-debt if it tracked F-T6).
- **REQ IDs cited:** the multi-tenant blob-key REQ and the async/janitor REQ from `wiki/architecture/backend-target-architecture.md`.
- **ADR required?** **No.** It mirrors an existing janitor pattern, introduces no capability, no contract change, and no policy/MUST-deviation. (The `internal/modules/jobs` → `templates/jobs` location correction is a runtime-truth alignment, not a policy decision.)

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — fits the proven janitor pattern; proceeds with a sequencing constraint and destructive-op guard rails.
- **Open hard-stops:** AS-3 raised (owning module ambiguous: audit said central `jobs`) and **resolved** → owner is `internal/modules/templates/jobs/`, wired in `metaldocs-api` main, mirroring `documents/jobs`. AS-1/AS-2 not triggered.
- **Locked constraints handed to implementation:**
  1. **Sequencing:** implement F-T6 **after** the F-O objectstore kernel wave so it consumes the tenant-guarded List/Delete primitive — not the legacy presign client.
  2. **Location/wiring:** new `internal/modules/templates/jobs/` package; `StartTemplateOrphanSweeper(ctx, repo, store, interval, maxAge)` mirroring `StartOrphanPendingSweeper`; started + stopped in `apps/api/cmd/metaldocs-api/main.go` beside the existing two sweepers.
  3. **Auth:** `authz.WithBackgroundBypass` background root (sibling pattern).
  4. **Destructive guard rails:** delete only keys >24h, under the correct tenant prefix, absent from the DB key set; log every deletion.
  5. **Multi-tenant:** reconcile per tenant; never cross the tenant prefix boundary.
  6. **Test:** integration test covering referenced-survives / old-orphan-deleted / young-orphan-survives / cross-tenant-untouched, on the canonical `testdb` framework.
