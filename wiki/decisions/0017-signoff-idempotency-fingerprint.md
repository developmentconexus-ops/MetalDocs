# ADR 0017 — Signoff idempotency fingerprint = client-stable inputs only

**Status:** accepted
**Date:** 2026-06-01
**Supersedes the F-002 fix in:** commit `aefb23ea8` (reorder + `stageIDForHash` empty-string fallback) — that change was merged but did not work end-to-end; see "Context" below.
**Related:** [0007-two-tier-authz.md](0007-two-tier-authz.md), `internal/platform/idempotency`, `wiki/modules/approval.md`

## Context

Signoff replay is backed by `metaldocs.idempotency_keys`. The store follows the
Stripe / IETF `Idempotency-Key` model:

- **Slot identity** is the table primary key `(tenant_id, actor_user_id, route_template, key)` (`db/baseline/0001_current_schema.sql`). The route template is the *literal* template (`POST /api/v1/documents/{id}/signoff`) — it carries no document ID.
- **`payload_hash` is a separate misuse guard.** On a same-slot re-claim the store compares the stored hash to the incoming hash; a mismatch returns `idempotency.ErrConflict` (`internal/platform/idempotency/postgres_store.go:134`), signalling "this key was reused for a different request."

The first F-002 fix reordered `BeginDocumentReplay` to run before state/eligibility
validation, but left the document-scoped `payloadHash` deriving its `stageID`
(and `inst.ID`) from the **loaded, mutable instance**. Because the recording
call ran while the stage was active (real stage ID) and the replay call ran after
the instance went terminal (`activeStage == nil` → empty stage ID), an identical
client retry produced a **different** fingerprint, tripped the misuse guard, and
returned `ErrConflict`. The document-scoped handler did not map `ErrConflict`, so
the client saw `500 internal.unknown` instead of the cached replay. Live QA on
2026-06-01 reproduced this exactly (PO-RH-006). The unit tests passed only because
the fake store ignored the payload hash and pre-seeded a terminal-state slot,
never exercising the active→terminal recording transition.

## Decision

**A signoff replay fingerprint MUST be derived only from client-stable request
inputs — never from server-resolved mutable state.**

- **Document-scoped** (`SignoffByDocumentHandler`): `payloadHash = sha256(docID, decision, reason, contentHash)`. `docID` comes from the path; the rest from the body. Instance ID and stage ID are **excluded** — the client neither supplies nor controls them, and they rotate/terminate between attempts. `docID` is retained because the route template has no document ID, so the hash is also what isolates distinct documents under a reused key. `contentHash` gives free cross-revision protection: a different revision has different content → different fingerprint → legitimate conflict.
- **Stage-scoped** (`SignoffHandler`): unchanged. Its `instanceID`/`stageID` come from the **URL path**, so they are client-stable and correctly part of the fingerprint.
- **Replay runs before the instance load.** With the fingerprint independent of instance state, the handler computes the hash and calls `BeginDocumentReplay` before touching the DB for the instance, then loads + validates state only on a cache miss. Symmetric with the stage-scoped handler.
- **`idempotency.ErrConflict` maps to `409 idempotency.key_conflict`** in the approval error mapper, so a genuine key-reuse-with-different-body returns 409 (not 500) for every store caller.

## Consequences

- Retries of an identical signoff replay the cached outcome regardless of intervening terminal transition or stage rotation. Clients **must rotate `Idempotency-Key`** for a genuinely new attempt; reusing a key with a changed body returns `409 idempotency.key_conflict`.
- First calls to a terminal-state instance still return `409 state.instance_completed` and the slot is released (`Fail`), not committed — a later eligible attempt under the same key is not blocked by a phantom success.
- The replay seam now returns a `SignoffReplayCommitter` interface (not the Postgres-bound concrete handle), so the record→replay lifecycle is unit-testable with a faithful in-memory double that enforces the payload-hash guard. The prior hash-agnostic fake is what let the bug ship.

## Alternatives rejected

- **Keep stage ID via a record-time captured identity** — fragile; after terminal there is no active stage to recover, and it re-introduces server state into the fingerprint.
- **Drop `payload_hash` for signoff entirely** — loses misuse detection *and* the cross-document isolation the PK does not provide.
- **Only map `ErrConflict`→409, keep the hash as-is** — turns the 500 into a 409 but still never replays after terminal; fixes the symptom, not the cause.
