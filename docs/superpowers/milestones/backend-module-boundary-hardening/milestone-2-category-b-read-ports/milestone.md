# Milestone 2 — Category B: owner-published read-ports

> **Program:** backend-module-boundary-hardening  ·  **Governing spec:** `../mission.md` (§5 rows 4–11 + 16, §7 M2)
> **Status:** Executing (operator-approved + HS-3 cleared 2026-06-20). Operator authorized bringing up a
> live test Postgres directly; a throwaway tmpfs container runs on `:5434` (non-secret CI creds
> `metaldocs:metaldocs`, role `metaldocs_app` created, isolated from the dev DB on `:5433`). The
> `tests/integration/testdb` harness is green against it (`go test -tags integration
> ./tests/integration/testdb/...` → ok), so D6 parity-before-delete can run on a real DB.
> **Validator PASS 2026-06-21** (F2.1–F2.4 closed + F2.5 HS-4 guard-test-realign fix;
> `qa/milestone-qa.md`). **HS-1 operator-approved 2026-06-21** — M3 dispatched to a fresh session.
> **Authored:** 2026-06-20 — *before any feature in this milestone began.* Test-PG DSN:
> `postgres://metaldocs:metaldocs@localhost:5434/metaldocs?sslmode=disable` (ephemeral; never `.env`).

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The end-of-milestone
> QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

Eliminate the **9 clean foreign point-reads** (census Category B, B1–B8 + N1) where a module reaches
into another module's **owned base table** with raw SQL to fetch a small fact. After this milestone, a
module that needs a fact owned by another module obtains it through that owner's **published read-port**
— a Go interface declared in the owner's `domain` package, implemented in the owner's `infrastructure`,
wired at the composition root — and **never** names the owner's base table in its own SQL. This is the
exact mechanism ADRs 0029/0030/0031/0038 each established for one table; M2 applies it to the remaining
Category-B set.

**Consumer-experienced change:** for each ported pair, the consumer module's query stops carrying the
owner's table name and storage shape; the owner becomes the single home for that read. Observable via:
the `hgcrossmodule` cilint guard clears each site's `hgPendingRemediation` ledger entry (build still
green), and a per-site **parity test** proves the ported path returns byte-identical results to the raw
SQL it replaces.

**Bar moved:** the Category-B class → **0 remaining sites** in `hgPendingRemediation` for the B/N1
rows. After M2 the only `hgPendingRemediation` entries left are the Category-C (M3) and C4 (M4) rows.

This is a **seam** change, **not** a logic change (mission §2 Non-Goals; D6 parity). No behavior,
visibility, or authz semantics change — parity tests are the lock.

## Appetite

- **Appetite:** 4 features, one read-port per owner. No new tables, no migrations, no view objects
  (views are M3/M4). A port is a Go interface + one Postgres adapter + composition-root wiring.
- **Rabbit holes (do not chase):**
  - **Denormalizing a snapshot column** onto the consumer's table to avoid the read — rejected; that is
    a schema/migration change, out of M2's appetite and a different design (mission §2: no schema
    changes beyond the published view M3/M4 need). Reads stay **live**.
  - **Collapsing several owner reads into one "god" query port** that returns more than the consumer
    needs — port shape is **consumer-driven**; return exactly the fact(s) the call site consumes, no
    speculative fields.
  - **Re-porting the Category-C membership `user_process_areas` reads** that happen to sit in the same
    files (e.g. `controlleddocuments/infrastructure/repository.go:150,492`) — those are **M3** (published
    view), not M2. Touch only the B-row reads in those files.
  - **Refactoring the consumer's surrounding logic** (the COALESCE fallback chains in `document_area.go`
    / `read_service.go`, the `GetActiveInstance` projection assembly) beyond the minimum needed to route
    the foreign read through the port. Parity forbids behavior drift; keep the surrounding Go shape.
  - **Migrating taxonomy `IsPublished` or the ADR-0030 port to explicit-tenant** signature cleanups —
    unrelated bounded defers; not triggered here.

## Features

Order is intentional: F2.1 (CD port) and F2.2 (documents port) are the two highest-traffic seams and
are mutually independent; F2.3 (taxonomy) and F2.4 (templates) are smaller, single-method ports. Each
feature is self-contained — its own owner publishes, its own consumer(s) adopt, parity-before-delete
per site.

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1 | `f2.1-cd-read-port` | **controlleddocuments (CD)** publishes a read-port for the CD facts consumed by **documents** / **documents/approval**: (i) `profile_code` for a controlled-document id — B1 (`documents/repository/repository.go:1701`, off-tx); (ii) `process_area_code` for a controlled-document id, used as the LEFT-JOIN fallback in the area-resolver COALESCE chain — B5 (`documents/application/document_area.go:37`, **in-tx**) and B6 (`documents/approval/application/read_service.go:355`, **in-tx**). Port interface in `controlleddocuments/domain`, adapter in `controlleddocuments/infrastructure`, wired at composition root; **tx-aware** variant (accepts the caller's `*sql.Tx`) for B5/B6. Consumer keeps its own-table SQL; only the `controlled_documents` read leaves the consumer's query. | Per-site **parity test** (raw SQL result == port result, incl. the COALESCE NULL-vs-empty ordering for B5/B6 and the cd-row-absent case) committed and **green BEFORE** the raw read is deleted; `go build ./...` and `go test ./...` green; `go run ./tools/cilint ./...` exit 0 with the B1/B5/B6 entries removed from `hgPendingRemediation`; `git grep` shows 0 `controlled_documents` reads remain under `documents/`. |
| F2.2 | `f2.2-documents-read-port` | **documents** (incl. its **approval** sub-context) publishes a read-port for the active-instance projection consumed by **CD** in `GetActiveInstance` (`controlleddocuments/infrastructure/repository.go`): the active/published `documents` rows + `document_revisions` content-hash (B2 `:532`, B3 `:539,545`) and the in-progress `approval_instances` id (B4 `:593`). Port interface(s) in `documents/domain` (and/or `documents/approval/domain`), adapter(s) in the owner's `infrastructure`, wired at composition root; CD consumes the interface and maps to its own `ActiveDocumentInstance`. **Foreign status literals** (`'draft'`,`'under_review'`,`'approved'`,`'rejected'`,`'scheduled'`,`'published'`,`'in_progress'`) in the moved reads use the **owner's** typed constants, not bare strings. | Per-site parity test (raw == port, across active/published/under-review/none cases) green **before** deletion; `go build`/`go test ./...` green; cilint guard clears B2/B3/B4 (entries removed from `hgPendingRemediation`); `git grep` shows 0 `documents`/`document_revisions`/`approval_instances` reads remain under `controlleddocuments/`; 0 bare foreign status literals in the ported queries. |
| F2.3 | `f2.3-taxonomy-read-port` | **taxonomy** publishes a process-area read-port over `document_process_areas` (taxonomy-owned): (i) area **name** for a `(tenant, code)` — B7 (`documents/repository/repository.go:154`, **in-tx**); (ii) area **existence** for a `(tenant, code)` — B8 (`iam/infrastructure/postgres/area_catalog_reader.go:28`, off-tx). Port interface in `taxonomy/domain`, adapter in `taxonomy/infrastructure`, wired into both consumers at the composition root; **tx-aware** for the in-tx B7 caller. | Per-site parity test (raw == port; name-found / name-absent for B7, exists / not-exists for B8) green **before** deletion; `go build`/`go test ./...` green; cilint guard clears B7/B8; `git grep` shows 0 `document_process_areas` reads remain under `documents/` or `iam/`. |
| F2.4 | `f2.4-templates-read-port` | **templates** publishes a read of `templates_template_version.placeholder_schema` for a document's template version, consumed by **documents** `fillin_service.go` (N1 `:225`). Per ADR 0030 precedent, **extend the existing** templates-owned `TemplateVersionPort` (do not introduce a parallel reader) with a `placeholder_schema` accessor keyed on version id; the consumer reads its own `documents.template_version_id` (own table) and calls the port. | Per-site parity test (raw == port; schema-present / null / no-row cases) green **before** deletion; `go build`/`go test ./...` green; cilint guard clears N1 (entry removed from `hgPendingRemediation`); `git grep` shows 0 `templates_template_version` reads remain under `documents/`. |

For each feature, "what to validate" is objectively checkable: a named parity test green before a named
raw read is deleted, a clean build/test, a guard exit code, and a grep returning 0.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M2 it enforces:

1. **Per-feature acceptance** — F2.1–F2.4 each meet every cell of their "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored: the port shape was read **from the consumer
   call site** (returns exactly the fact the site consumes), the interface lives in the **owner's**
   `domain` package (not the consumer's), and the consumer imports the interface — never the owner's
   table.
2. **Workflow-class QA checklist** — persistence (`wiki/quality/*` persistence/db checklist) +
   module-boundaries (DDD ownership: producer = owner, consumer imports interface only). Test-framework
   discipline for any new parity tests (canonical fixture framework, ADR 0034).
3. **Regression** — M0 + M1 still pass their gates: ADR-0039 unchanged; M1's typed-constants result
   intact; the `hgPendingRemediation` ledger contains **only** the Category-C (M3) + C4 (M4) rows after
   M2 (the 9 B/N1 entries removed, none of the C/C4 entries touched); `go run ./tools/cilint ./...` exit
   0 on the full tree throughout.
4. **Quality-bar / root-cause check** — the Category-B class is re-measured: `git grep` of each B/N1
   table token under its consumer module returns **0** raw reads, and each was replaced by a call to the
   **owner's** published port (root cause = consumer encoding the owner's storage shape), **not** by a
   consumer-local copy of the query or a duplicate adapter over the same table (symptom patch — the
   ADR-0030 "no parallel reader" rule).
5. **No unplanned scope** — only the 9 named reads leave their consumers; the Category-C reads in the
   same files are untouched (M3); no migrations, no views, no schema changes; no consumer logic
   refactored beyond routing the foreign read. Any other touched file is drift and must be recorded with
   rationale.
6. **Parity-before-delete (D6) audit** — for every deleted raw read, its parity test exists, asserts
   raw-path == port-path equality, and the feature's `evidence.md` shows it **green before** the
   deletion commit (or, if Docker Postgres :5433 was unavailable, the integration parity step is marked
   **not-run / HS-3** explicitly — never false-green).

## Dependencies & constraints

- **Depends on:** M0 (ADR-0039 definition + binding census + `hgcrossmodule` guard with the
  `hgPendingRemediation` ledger this milestone drains) and M1 (Category-A constants — independent, no
  overlap). Addresses census rows **B1–B8 + N1** exactly.
- **Quality goals (ranked):** **1. correctness/parity** (byte-identical results — D6 is non-negotiable)
  > **2. boundary integrity** (owner is the single home for the read; consumer imports an interface) >
  **3. simplicity** (smallest port that satisfies the consumer; no speculative fields) > performance
  (one extra bounded round-trip per read is accepted, per ADR 0038 cost note).
- **Architectural constraints (hard rules):**
  - **No migrations, no views, no schema changes.** M2 is Go interfaces + adapters + wiring only.
    (Views = M3/M4.) Reads stay **live** (no snapshot/denormalization).
  - **Owner-published read-port pattern (ADR 0029/0030/0031/0038):** interface in owner `domain`,
    impl in owner `infrastructure`, wired at composition root; consumer depends on the interface. No
    **parallel reader** over an already-ported owner table (ADR 0030 rule) — **extend** the existing
    port where one exists (F2.4 extends ADR-0030 `TemplateVersionPort`).
  - **tx-aware where in-tx (B5, B6, B7):** the port variant for these callers must execute inside the
    caller's existing `*sql.Tx` (the read currently runs in-tx). **HS-PRE-1:** these are plain,
    **non-recording** `SELECT`s — they must stay non-recording; no port may add an authz-recording read
    inside a lock-holding tx.
  - **No import cycle:** owner `domain` must not import the consumer module. Verify per feature
    (taxonomy/CD/documents/templates `domain` import nothing from their consumers).
  - **D6 parity-before-delete:** no raw read deleted without its green parity test (constraint the C6
    audit binds on).
  - **Test discipline:** new parity/unit tests use the canonical fixture framework (ADR 0034); integration
    parity needs live PG (:5433) — if down, mark **not-run (HS-3)**, never false-green.
- **Risks (named):**
  - *COALESCE NULL-vs-empty parity (B5/B6).* `COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')`
    treats a NULL snapshot differently from an empty-string snapshot; the Go split must preserve that the
    non-NULL own snapshot wins even when "". **Mitigation:** parity test seeds NULL-snapshot,
    empty-snapshot, and cd-absent cases and asserts identical output before deletion.
  - *`GetActiveInstance` is a multi-table projection, not a point read (F2.2).* The FULL OUTER JOIN +
    derived approval lookup is the largest port. **Mitigation:** the port returns the exact projection CD
    already maps to `ActiveDocumentInstance`; parity test covers active/published/under-review/none.
  - *A port turns out to need a cross-module shared-API redesign* (not a contained read-port).
    **Mitigation:** HS-2 stop — surface the boundary + minimum prerequisite plan; do not symptom-patch.
  - *Docker Postgres :5433 down* → integration parity not runnable. **Mitigation:** HS-3 — mark the
    integration step not-run explicitly in evidence; never false-green a skipped parity test.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | **The M2 boundary itself.** On validator PASS, the main session flips status and **stops** for operator review. No M3 start, no merge, without explicit approval. |
| HS-2 | A Category-B port turns out to require a cross-module **API redesign** beyond a contained read-port (a shared-API change, a consumer-side contract reshape, a new write path). **Stop**; surface the boundary + minimum prerequisite plan; do not symptom-patch the read. (Not expected — all 9 are clean point/projection reads.) |
| HS-3 | A prerequisite boundary fails: `go build ./...` red, or Docker Postgres (:5433) unavailable for an integration parity test. Repair / **note the gap explicitly** (mark the parity step not-run), rerun the checkpoint, resume. **Never false-green a skipped parity test.** |
| HS-4 | The `milestone-validator` returns FAIL (a parity gap, a symptom-patch / parallel-reader, an unported B row, a guard regression, scope drift). Open the named fix feature; re-run its lifecycle; re-dispatch the validator. |
| HS-6 | A Category-B site the census missed surfaces mid-milestone and changes M2's shape (e.g. another consumer of one of these owner tables). **Stop**; surface to the operator; replan before continuing. |
| HS-PRE-1 | A port would place an **authz-recording** read inside a lock-holding atomic tx. **Forbidden.** The B5/B6/B7 in-tx reads are plain non-recording `SELECT`s — they must stay so; keep the port read off any lock-holding/recording path. |
