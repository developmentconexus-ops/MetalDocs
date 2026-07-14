# Mission (DESIGN-ONLY, parked): Document Distribution & Read-Tracking

> **Status:** DESIGN-ONLY — parked backlog. **Not started. Not specced via `/mission`.**
> **Authored:** 2026-06-21 (recon during frontend-screen-completion M2 scoping).
> **Execution gate:** Execute **only after the `frontend-screen-completion` mission completes.**
> When that gate clears, promote this design through the `/mission` skill (Phase 1 → governing
> spec → milestone tree). Until then it is a designed-but-dormant backlog entry, nothing more.
> **Baseline:** HEAD `d477e9f0` (backend Grade-A, 2026-06-21). Latest migration
> `db/migrations/0244_documents_search_projection.sql` → next free index `0245`.
> **Provenance:** consolidates the fanout / read-tracking design that used to live in
> [`distribuicao.md`](distribuicao.md) plus the M2 scoping recon. `distribuicao.md` now points here
> for everything write-path; it keeps only the M2 derive-on-read coverage-scope follow-ups.

---

## 1. Why this exists (the gap)

The Distribuição & Cobertura screen (`/documents/:documentId/distribution`) wants to answer a
compliance question: **"of everyone obligated to read this controlled document, who actually read
and acknowledged it, by when, and who is overdue?"**

That question has a **denominator** (who is obligated) and a **numerator** (who read / acknowledged).
Verified runtime truth at baseline:

- **Denominator — derivable, exists today.** The obligated set is computable from visibility grants
  + area membership: `controlled_document_area_grants`, `controlled_document_user_grants`,
  `document_process_areas`, `user_process_areas`, and the active-membership view
  `metaldocs.v_active_user_areas`. **This is what frontend-screen-completion M2 builds** (derive-on-read
  coverage-*scope*).
- **Numerator — does not exist anywhere.** There is **no** read event, no acknowledgement event, no
  per-recipient distribution target, no reminder job. `approval_signoffs` is **pre-publish approval
  only** — it is not a reader-side read/ack signal and must not be repurposed as one. So a coverage
  endpoint built today has a real denominator and **nothing real to count** for the numerator.
- **Action layer — does not exist.** The 4 hero CTAs (mass reminder, export report, add recipients,
  fanout policy) and all row/bulk actions require a fanout + dispatch module that does not exist.

This mission owns the **numerator + the action layer**: the evidence write-path that produces real
read/ack data, the reader surface that emits it, and the actions that operate on it. It is
**redesign-grade** (worker/outbox semantics, recipient resolution, idempotent dispatch, ack
semantics) — exactly the class the QA operating system says is *reported, not patched* inside a
screen-wiring milestone.

## 2. The 3-layer coverage model (boundary of ownership)

| Layer | What it is | Exists at baseline? | Owner |
|---|---|---|---|
| **Obligation** (denominator) | Who must read this doc — derived from grants + area membership | **Yes** (derivable) | **frontend-screen-completion M2** (derive-on-read) |
| **Evidence** (numerator) | Per-recipient read + acknowledge events over time | **No** | **THIS mission** |
| **Serving / rollup** | The 4 read endpoints that join obligation + evidence into KPI / recipients / coverage / timeline | Partial — M2 ships the obligation-only subset | M2 ships denominator subset; **this mission** completes the numerator-bearing fields + the timeline endpoint |
| **Action** | Reminders, export, fanout/recipient add, policy | **No** | **THIS mission** |

**The clean cut between M2 and this mission:** M2 ships honest denominator data and renders read/ack
as an explicit *"tracking pending"* state (not a fake `0%`, not illustrative). The moment a real
numerator is wanted, it comes from here — M2 cannot manufacture it without fabricating data.

## 3. Crux decision to resolve FIRST: snapshot-at-publish vs derive-on-read

This is the fork that shapes the whole data model. Resolve it at mission kickoff **before any table
is written** — it is an architecture decision worthy of an ADR.

- **Derive-on-read (recompute the obligated set live on every query).**
  - *Pro:* no obligation snapshot to maintain; always reflects current grants/membership; zero change
    to the publish path.
  - *Con:* the "obligated set" drifts as people join/leave areas — historical compliance ("who was
    obligated *when this revision published*") becomes unanswerable; auditors usually want the frozen
    set.
- **Snapshot-at-publish (freeze the obligated target list into `document_distribution_targets` at
  publish time).**
  - *Pro:* stable denominator, real historical compliance, natural anchor for reminder jobs and
    overdue math.
  - *Con:* **touches the Grade-A atomic `PublishApproved()` path**
    (`internal/modules/documents/approval/application/publish_service.go:52-148`). That is an **HS-2
    boundary** — a publish-path mutation must not be smuggled in as a side-effect. It needs its own
    ADR, its own tests, and must respect **H-PRE-1** (never call an authz-recording read inside the
    lock-holding atomic publish tx — resolve targets *outside* the critical section, or via an outbox
    event consumed by a worker after commit).

**Design lean (not binding):** snapshot the target set, but produce it via a **post-commit outbox
event** (publish emits `document.published`; a distribution worker resolves + writes targets), so the
atomic publish tx is untouched and H-PRE-1 holds. The read/ack numerator stays event-sourced
regardless of which way the denominator goes. The kickoff ADR settles it.

## 4. Data model (sketch — settle column-level in the kickoff ADR)

Four new tables, indices `0245+`, owned by a new module (§6). Schemas honor the existing migration
and audit conventions (`metaldocs.audit_events` stays the canonical audit sink; these are
*operational* tables, not an audit duplicate).

- **`document_distribution_targets`** — the obligated set (one row per recipient per published
  revision). `document_id`, `revision_id`, `user_id`, `source` (area-grant | user-grant), `area_code`
  (nullable), `read_deadline`, `created_at`. Populated at publish (snapshot path) or materialized on
  demand (derive path). Denominator of record.
- **`document_read_events`** — append-only read signal. `id`, `document_id`, `revision_id`,
  `user_id`, `read_at`, `source` (web-view | api), `actor_context`. First-read vs re-read derived by
  `MIN(read_at)`.
- **`document_acknowledgement_events`** — append-only explicit acknowledgement (stronger than read;
  "I have read and understood"). `id`, `document_id`, `revision_id`, `user_id`, `acknowledged_at`,
  `ack_method`, `signature_ref` (nullable, if a typed/sign affirmation is required by policy).
- **`document_reminder_jobs`** — reminder dispatch ledger (idempotent). `id`, `document_id`,
  `revision_id`, `recipient_id`, `kind` (initial | reminder | overdue), `scheduled_at`, `sent_at`
  (nullable), `dedupe_key` (unique), `status`. Backed by the outbox/worker pattern.

> Numerator = `read_events` / `ack_events` joined against `distribution_targets`. Overdue =
> `target.read_deadline < now AND no ack_event`. Adoption curve = cumulative distinct
> `user_id` over `read_at` / `acknowledged_at`.

## 5. Evidence write-path (where the numerator is born)

The numerator is **un-isolatable without a reader-side producer.** Two producers:

1. **Implicit read** — when an obligated user opens the published document (the Publicado screen,
   frontend-screen-completion **M4**), emit a `document_read_events` row (first-open, debounced).
   This couples directly to M4: the read producer lives on M4's screen, which is why "real read %"
   cannot be finished inside M2.
2. **Explicit acknowledge** — a **reader-side "Li e estou ciente" affirmation** on the Publicado
   screen writes `document_acknowledgement_events`. This is net-new UI **and** a net-new
   write-endpoint — it does not exist in any milestone yet and is owned here.

Both pass through tier-2 `authz.Require` with a new capability (§7) and must satisfy the
`trg_require_cap_asserted` DB tripwire (the cap assertion is recorded before the write). Read/ack are
high-frequency, low-risk writes → they must **not** sit inside any lock-holding atomic tx (H-PRE-1).

## 6. Module + serving layer

- **New module `internal/modules/distribution`** (hexagonal, mirroring the existing module shape:
  `domain/` `application/` `delivery/` `infrastructure/` `repository/`). Owns target resolution,
  read/ack ingestion, reminder dispatch (worker + outbox), and the rollup queries.
- **Contract-first** — author the OpenAPI surface, run `oapi-codegen`, regenerate FE types
  (`lib/api-types/`), keep `api-lint -strict = 0` and all 6 CI guards green. RFC 9457 problem+json on
  every error path.
- **The 4 read endpoints** (consolidated from `distribuicao.md`; M2 ships the denominator-only subset
  of the first three, this mission completes the numerator fields + the timeline endpoint):

| Endpoint | Purpose | M2 ships | This mission completes |
|---|---|---|---|
| `GET /documents/:id/distribution` | KPI rollup (`total_targets`, `acknowledged`, `read`, `pending`, `overdue`) + sidebar facts | `total_targets`, `pending`, deadline, by-area facts (denominator) | `acknowledged`, `read`, `overdue` (numerator) |
| `GET /documents/:id/distribution/recipients?page=&pageSize=&status=` | Paginated recipient table; `X-Total-Count` + `Link` headers | recipient identity + area + obligated status | `read_at`, `acknowledged_at`, `last_event_at`, status filters |
| `GET /documents/:id/distribution/coverage` | By-area breakdown | `area_code`, `area_name`, `total` (denominator) | `read`, `acknowledged` per area (numerator) |
| `GET /documents/:id/distribution/timeline?granularity=day` | Cumulative read/ack adoption curve | — (numerator-only; not in M2) | entire endpoint |

## 7. Authz / caps

- New capability family for distribution reads + reader writes (e.g. `CapDistributionRead` for the
  rollup endpoints, `CapDocumentAcknowledge` for the reader-side ack write). Wire through tier-2
  `authz.Require` + the `trg_require_cap_asserted` tripwire. Do **not** overload an existing cap —
  reader-acknowledge is a distinct authority from document admin.
- Reminder/export/fanout actions (§8) gate on a stronger admin cap (e.g. `CapDistributionManage`).
- **H-PRE-1 hard constraint:** any authz-recording read must stay **off** the lock-holding atomic tx
  (publish path, reminder-dispatch critical section). Resolve targets and record caps outside the
  critical section or in the post-commit worker.

## 8. Action layer (the 4 CTAs + row/bulk actions)

All require the dispatch module + a notification channel. **Couples to frontend-screen-completion
M3 (Notifications)** — reminders are delivered through that channel; do not build a parallel one.

| CTA / action | Endpoint | Needs |
|---|---|---|
| Lembrete em massa | `POST /documents/:id/fanout/reminders/bulk` | reminder worker + M3 channel |
| Exportar relatório | `GET /documents/:id/distribution/export` | rollup query → CSV/PDF stream |
| Adicionar destinatários | `POST /documents/:id/fanout/recipients` | target write + re-resolution |
| Política de fanout | settings surface (**no design yet** — own UX task) | policy model + UI |
| Row action — enviar lembrete | `POST /documents/:id/fanout/reminders/:recipientId` | idempotent single dispatch |
| Bulk action bar | bulk reminder / reassign / export | batch dispatch + dedupe |

Reminder dispatch is **idempotent** (the `document_reminder_jobs.dedupe_key` unique constraint +
outbox) — a double-click or a worker retry must never double-send.

## 9. Dependencies & coupling

- **Grade-A publish path (HS-2):** snapshot-at-publish touches `PublishApproved()`. Outbox-after-commit
  keeps the atomic tx untouched — the only safe way in.
- **frontend-screen-completion M3 (Notifications):** the reminder delivery channel. Build reminders
  *on* M3's channel, not beside it.
- **frontend-screen-completion M4 (Publicado):** hosts both numerator producers (implicit read on
  open + explicit acknowledge affirmation). The read/ack UI lives there; this mission supplies the
  endpoints it writes to.
- **`render-fanout` outbox pattern** (`modules/render-fanout.md`): reuse its worker/outbox semantics
  for reminder dispatch rather than inventing a second pattern.

## 10. Sequencing & sizing

Not one endpoint — a **3–4 milestone mini-mission**, runnable only after the frontend mission closes
(M3 + M4 must exist for the action layer + read producer to have homes):

1. **M-D1 — evidence schema + ingestion.** Kickoff ADR (snapshot vs derive), the 4 tables, read/ack
   write endpoints, caps. No UI beyond what M4 already exposes.
2. **M-D2 — reader producers.** Wire implicit-read emit + explicit acknowledge affirmation on the M4
   Publicado screen; backfill/derive targets for already-published docs.
3. **M-D3 — serving completion.** Finish the numerator fields on the 3 rollup endpoints + build the
   timeline endpoint; re-wire the Distribuição screen from "tracking pending" to live read/ack;
   pick a chart library (candidates `recharts` / `visx` — tree-shakeable, accessible, token-compatible).
4. **M-D4 — action layer.** Reminder worker + the CTAs/row/bulk actions on M3's notification channel;
   export; fanout policy UX.

## 11. Blast radius

New module + 4 tables + a publish-path outbox event + reader-side UI on M4 + reminder worker on M3's
channel + 4 (eventually 6+) endpoints + regenerated contract/types. Touches: publish (read-only via
outbox), notifications channel, Publicado screen, contract surface, CI guards. **High** — hence its
own mission, not a milestone bolt-on.

## 12. Terminal acceptance sketch (fill at kickoff)

- Real read/ack numerator rendered on the Distribuição screen — **zero** illustrative / "em breve"
  on that screen; the M2 "tracking pending" state is fully replaced by live data.
- Reader-side acknowledge produces a verifiable `document_acknowledgement_events` row; implicit read
  produces a `document_read_events` row on first open.
- Reminder dispatch idempotent under retry/double-fire (dedupe constraint proven).
- New backend `api-lint -strict = 0`, all 6 CI guards green, `go build ./...` + `go test ./...` clean,
  publish path unchanged except the post-commit outbox emit (its atomic tx diff = 0).
- Both frontend reviewer agents APPROVE the Distribuição + Publicado acknowledge surfaces.

## 13. Hard-stops (carry into the real mission spec)

| ID | Trigger here |
|----|--------------|
| HS-2 | Snapshot-at-publish or any change that mutates the Grade-A atomic `PublishApproved()` tx. Stop; outbox-after-commit only; ADR required. |
| HS-2 | Repurposing `approval_signoffs` as a read/ack signal (wrong domain — pre-publish approval). |
| HS-3 | H-PRE-1 violation: an authz-recording read placed inside a lock-holding atomic tx. |
| HS-6 | Building a second notification channel beside M3's, or a second outbox beside `render-fanout`'s. |

## 14. Open decisions for kickoff (do not pre-bake)

1. Snapshot-at-publish **vs** derive-on-read for the denominator (§3) — the ADR.
2. Does acknowledge require a typed/signed affirmation (`signature_ref`) or is a click sufficient?
   (compliance-policy call — ask the operator).
3. Is read-deadline per-document, per-area, or per-policy?
4. Export format(s): CSV only, or PDF compliance report too?
5. Reminder cadence policy: fixed schedule vs configurable per document.
