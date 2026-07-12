# System-impact analysis — Approval kernel extraction (M3 / ROADMAP unit 3.1)

**Date:** 2026-07-12
**Intent (one line):** Extract `internal/modules/documents/approval` → top-level `internal/modules/approval` (15th bounded context); generalize kernel to `(subject_kind, subject_key)`; rewire templates approval onto the kernel; supersede ADR 0072.
**Work type:** module (new top-level bounded context)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow — proceed; ADR supersession + contract-generalization ratification flagged as locked constraints (§10).

> This is a **promotion** of an existing, mature, ratified subtree, not a greenfield module. ADR 0072
> explicitly gated this move behind a named trigger; the trigger has fired (see §2). The ten sections
> are answered against the extraction, not a from-scratch build.

---

## 1. Classify & own

- **Work type:** module — births `internal/modules/approval` as the 15th top-level bounded context.
- **Owning module(s):** the new `approval` module owns the approval *kernel*: routes, instances,
  stages, signoffs, verdicts, delegation, SoD, quorum, freeze, fast-forward, scheduler, SLA surfacing.
  Generalized subject key = `(subject_kind, subject_key)` — no longer document-specific.
- **Explicitly NOT owning:**
  - `documents` — no longer owns the approval workflow; becomes a **consumer** (`subject_kind=document`,
    `subject_key=profile_code`). Retains document domain (versions, families, areas).
  - `templates` — becomes the **second consumer** (`subject_kind=template`, `subject_key=doc_type`);
    its parallel template-local approval path (`templates/application/approval_config.go`,
    `templates/domain/approval.go`) is **retired**, not extended.
- **Cross-module edges (with direction)** — `A → B` = A depends on B:
  - `documents → approval` (application + domain). Production edges today: exactly **2 files** —
    `documents/delivery/http/handler.go` (→ `approval/application`) and
    `documents/infrastructure/active_instance_reader.go` (→ `approval/domain`). Both already hit
    published layers; after extraction these are import-path renames, boundary-clean.
  - `approval → documents` (application + domain). **24 production edges** (17 domain, 7 application),
    all on the layer allow-list (`domain`/`application`) → allowed cross-module direction post-extraction.
  - `templates → approval` (application + api) — **new** consumer edge, published-surface only.
  - `audit → approval` — `audit/delivery/http/handler.go → approval/http/router`. **The single genuine
    violation-class edge**: `http/router` is NOT published surface. Must be re-ported (see §7).
  - `jobs/approval_sla_surfacer → approval/domain`, `jobs/stuck_instance_watchdog → approval/application`
    — already published-surface; import-path rename only.
  - **Direction constraint (locked):** `render`, `search`, `notifications` must NOT gain a dependency
    on `approval`; approval reaches them (if at all) via their published ports / the outbox.
- **Ambiguity?** None. Owning module unambiguous (the new `approval` module). No **AS-3**.

## 2. Foundation verdict

- **Base you'd build on:** the existing `internal/modules/documents/approval` subtree — 164 Go files,
  full-layer (`api application domain http infrastructure jobs`). Grade-A backend (signed off
  2026-06-21); G1 (per-profile signature policy), G2 (request_changes on approval stages), G3
  (fast-forward two-entries-one-tx) **merged and ratified**. This is the largest, most mature module
  in the codebase.
- **Sound, or legacy/patch/workaround?** **SOUND.** Not a patch. The subtree is the ratified approval
  engine; extraction moves it intact.
- **Global-maximum judgement:** the extraction *is itself* the global-maximum move, and ADR 0072
  named it as such. ADR 0072 (Accepted 2026-07-06) **rejected** promotion *at that time* on a
  hygiene-milestone boundary, but recorded a **promotion trigger** verbatim: *"if approval ever needs
  an independent lifecycle … or a **second bounded-context consumer that isn't `documents`**, the
  promotion plan starts from this ADR's coupling-edge inventory."* Spec C1 confirms the trigger has
  fired: *"ADR: kernel extraction + supersede ADR 0072 (rationale: **second consumer arrived**)"* —
  the second consumer is templates. Extracting now is executing the ADR-gated plan, not optimizing
  inside a patch. **No AS-2.**
  - ADR 0072's "dense bidirectional 100+ edges" cost estimate (2026-07-06) counted test edges and is
    now stale: the live production coupling is 2 files inbound + 24 outbound-on-allowed-layers + 1 true
    re-port (`audit → approval/http/router`). The extraction is materially more tractable than the ADR
    feared — but that does not lower the bar: behavior must stay byte-equal (§8, §10).

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes (moves intact) | Existing tier-2 `authz.Require` calls move with the code; tripwire arms follow the module (generated arms per GMR M2 — regenerate for the new module path). No new capability, no role reasoning introduced. | `authz.Require`, `authz.SeedTxIdentity`; generated tripwire registry |
| Contract-first (OpenAPI + oapi-codegen) | **Yes — decisive** | Routes generalize to `(subject_kind, subject_key)`. Spec edit + `oapi-codegen` regenerate is the ONLY way. **Route generalization is a contract decision → operator ratification BEFORE implementing** (P1 SPEC escalation). | `api/openapi/v1/openapi.yaml`, module `cfg.yaml` + `gen.go` |
| Multi-tenant pooled | Yes (moves intact) | Every approval table already carries `tenant_id`; tx-local GUCs preserved; cross-tenant → 404. Generalization adds `subject_kind`/`subject_key` columns, not tenancy change. | `tenant.FromContext`, `authz.SeedTxIdentity` |
| Async = transactional outbox | Yes (moves intact) | Lifecycle-event enqueuer + scheduled-publish jobs move with the module; outbox pattern unchanged; consumers stay idempotent. | outbox repo; existing `approval/jobs/*` |
| DB enforces invariants | Yes | Existing triggers/constraints move; generalization migration adds CHECK/constraints on `(subject_kind, subject_key)`; **idempotency route templates (fast-forward + signoff) must keep working** across the rename. | migration + existing tripwire SQL |
| Cross-module via published interface only | **Yes — this work STRENGTHENS it** | Post-extraction, documents + templates + audit + jobs consume approval via `domain`/`application`/`api` only. The one violation-class edge (`audit → approval/http/router`) is re-ported to a published surface. Module-boundary lint is the core proof (L0). | `application/ports.go`, `domain/port.go`; `check-module-boundaries.ps1` |

No invariant violated. **No AS-1.** Invariant 6 is the *point* of the unit.

## 4. Capability wiring

**N/A** — extraction introduces **no new capability**. Existing capabilities (approval/signoff/route-admin/
review/publish/etc.) and their tripwire arms move with the code. Tripwire arms are generated from the
Go registry (GMR M2) → regenerate so arms reference the new module path; `TestCapabilityRegistrySize`
unchanged (count is stable, only import paths move). If templates-onto-kernel rewiring surfaces a
capability gap (e.g. template approval currently uses a template-scoped cap), that is a **design item to
surface in brainstorming**, not a silent add — flag for the 10-touchpoint walk then.

## 5. Module wiring

Births `internal/modules/approval` — ordered birth steps (per `references/module-wiring.md`), adapted
for a **move** (folders + code already exist; the work is relocation + boundary-cleaning):

1. **Folders** — `{api, application, domain, http, infrastructure, jobs}` relocate under
   `internal/modules/approval/` (drop the `documents/` prefix).
2. **Domain** — entities + `domain/port.go` (provider ports) move intact; generalize document-specific
   value objects to `(subject_kind, subject_key)`.
3. **Application** — service + `application/ports.go` (consumer ports approval needs from documents/
   templates) move; add the published application-service surface documents + templates consume.
4. **Infrastructure** — repository moves; touches only approval's own tables. `subject_kind`/`subject_key`
   columns added.
5. **Delivery** — `http/router.go` + handlers move; router becomes the module's `RegisterRoutes`.
6. **api codegen** — new `api/cfg.yaml` `include-tags` for the generalized routes + `gen.go`; regenerate.
7. **OpenAPI** — retag routes under the extracted module; generalized `(subject_kind, subject_key)` path.
8. **`module.go` (optional)** — constructor with panic-on-nil-deps if the composition benefits (follow
   taxonomy shape).
9. **Composition root** — rewire `apps/api/cmd/metaldocs-api/main.go` (+ worker/jobs binaries for the
   scheduled-publish + lifecycle-enqueuer + sla-surfacer + watchdog consumers) to the new module path.
10. **Migration** — `db/migrations/0NNN_*.sql`: add `subject_kind`/`subject_key`, backfill existing
    rows to `document+profile_code`, keep `tenant_id` + all invariant triggers; idempotency route
    templates preserved.
11. **Docs** — `wiki/modules/approval.md` (12-section) + `approval-tech-debt.md` +
    `wiki/modules/index.md` entry; supersede ADR 0072 with a new ADR (wiki-curator).

**Boundary rule:** documents + templates depend on approval's published ports only; approval depends on
documents' `domain`/`application` published layers only. `check-module-boundaries.ps1` GREEN is the
gate.

## 6. Frameworks to reuse, not reinvent

All already used by the subtree; they move intact — **no hand-rolled equivalents introduced**:
`TxRunner` (`Do`/`DoReadOnly`), `tenant.FromContext`, `authz.SeedTxIdentity`/`Require`,
`problem.New`/`Write`, `httpresponse`, `audit.RecordTx`, outbox repo, `testdb` factory.
**One catalog item to resolve during the move:** the module-private strict-JSON decoder
`documents/approval/http/contracts/strictjson.go` — post-extraction it must not be imported
cross-module; either it stays module-private under `approval/http/contracts` (fine — only approval
uses it) or promote to `internal/platform/strictjson` if a second module needs it. Default: keep
module-private (no new cross-module dependency).

## 7. Contract & data

- **OpenAPI-first:** routes generalize `document`-scoped paths to `(subject_kind, subject_key)`.
  **This is the ratification gate** — surface the exact route shape (path template, params, backward-compat
  for existing `document+profile_code` rows) in a P1 SPEC ESCALATION to hub/operator BEFORE editing the
  spec or any Go. New tags for the extracted module; DTOs regenerate.
- **Migration:** `db/migrations/0NNN_*.sql` — add `subject_kind` + `subject_key` columns, backfill
  existing routes/instances to `('document', profile_code)`, add CHECK constraints (DB enforces
  invariants). **Idempotency route templates (fast-forward + signoff) MUST keep working** across the
  move — verify idempotency keys still resolve. Expand/contract: add generalized columns alongside,
  backfill, then constrain — never break the live contract in one step.
- **Templates data migration:** `ApprovalConfig{ReviewerRole?, ApproverRole}` → kernel routes
  (2-stage, or 1-stage when no reviewer) per `doc_type`; in-flight template versions get a **defined
  cutover rule** (spec C2 — no silent state loss). This is a substantive migration, not a rename.
- **Destructive change:** retiring the templates-local approval path (`CheckSegregation`,
  `CanTransition(hasReviewer)`, template approval audit vocabulary) — gate behind the kernel being live
  for templates first (expand/contract).
- **KNOWN open defect — DO NOT TOUCH:** E-PROD-2 `document_profiles` PK=(code) not tenant-scoped
  (operator decision pending). If the generalization migration touches profile-code keying, **note it,
  do not repair** — collision risk with the pending operator decision.

## 8. Test & QA plan

- **Canonical framework:** existing approval tests move with the code; **new** integration tests
  (generalization + templates-on-kernel) use `tests/integration/testdb` (`Open`, `seedWithCapsIdentity`
  per `authz_ctx.go`), `//go:build integration`, R1–R4 discipline.
- **QA gates (module → all 6 apply):** contract (generalized routes), authz (tripwire arms regenerated +
  green), multi-tenant isolation (cross-tenant 404 on generalized keys), async/idempotency (fast-forward
  + signoff idempotency templates), DB-invariant (new CHECK constraints), docs (wiki + ADR).
- **Behavior-byte-equal bar:** G1/G2/G3 semantics unchanged — document approval lifecycle (M2b F8-class
  walkthrough) re-run green; template lifecycle live QA (config→route migration, review+approve+publish a
  template version through the kernel, worklist + SoD + delegation).
- **Verification ladder:** L0 `go build ./...` + `api-lint -strict` + **`check-module-boundaries.ps1`
  (core proof)**; L1 `go test ./...` + `test-integration.ps1` (accepted RED = exactly 9 tests/4 pkgs
  E-PROD-1..5; bar = zero NEW failures). FE: `pnpm exec tsc --noEmit` clean if api-types shift.
- **Evidence shape:** commands + outcomes + review/QA disposition + bounded defers; dispatch ledger in
  `docs/superpowers/reports/2026-07-12-m3-kernel-extraction-evidence.md`. milestone-validator PASS before
  close.

## 9. Docs / ADR

- **Wiki:** new `wiki/modules/approval.md` (12-section) + `wiki/modules/approval-tech-debt.md` +
  `wiki/modules/index.md` entry; update `documents.md` + `templates.md` (approval now consumed, not
  owned); refresh `backend-target-architecture.md` REQ references (14→15 modules). Dispatch **wiki-curator**.
- **REQ IDs cited:** REQ-TOP-1 (cross-module published-surface), REQ-AUTHZ-* (capabilities), the
  module-count REQ in `backend-target-architecture.md` (bump 14→15). Confirm exact IDs at wiki-sync time.
- **ADR required? YES.** New ADR **supersedes ADR 0072** (policy change: the nested-exception ruling is
  reversed because its own named promotion trigger — second bounded-context consumer — fired). ADR records:
  extraction rationale, generalized subject model, templates-unification, superseded-0072 linkage, and the
  boundary-guard realignment (drop the `documents/approval` nested-family exception; `approval` is now a
  first-class module).

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to brainstorming/planning. Fits the architecture cleanly and is
  the ADR-0072-gated global-maximum move; Yellow (not Green) because it carries a mandatory ADR
  supersession and a contract-generalization decision that requires operator ratification before
  implementation.
- **Open hard-stops:** AS-1 none · AS-2 none · AS-3 none. Foundation sound; owning module unambiguous;
  no invariant violated.
- **Locked constraints handed to brainstorming/planning:**
  1. **Contract-first ratification gate** — route generalization to `(subject_kind, subject_key)` is a
     contract decision; send a P1 SPEC ESCALATION to hub/operator and WAIT for the answer BEFORE editing
     `openapi.yaml` or any route Go.
  2. **Supersede ADR 0072** with a new ADR; realign `check-module-boundaries.ps1` (drop the nested-family
     exception; `approval` becomes a first-class module) — the boundary lint GREEN is the core proof.
  3. **Behavior byte-equal** — G1/G2/G3 semantics move intact; behavior change = defect. Document
     lifecycle (M2b F8-class) + template lifecycle re-run green.
  4. **AuthZ capabilities never roles (ADR 0022)** — no new cap unless a real gap surfaces (then full
     10-touchpoint walk); tripwire arms regenerated for the new module path.
  5. **H-PRE-1** — authz-recording reads stay off lock-holding tx during the move.
  6. **Migration discipline** — expand/contract; `tenant_id` preserved; idempotency route templates
     (fast-forward + signoff) keep working; existing rows backfill to `('document', profile_code)`.
  7. **E-PROD-2 do-not-touch** — `document_profiles` PK=(code) not tenant-scoped is an open operator
     decision; note if the generalization migration approaches it, do not repair, do not collide.
  8. **Direction constraint** — `render`/`search`/`notifications` must not depend on `approval`.
  9. **The one real re-port** — `audit/delivery/http/handler.go → approval/http/router` must move to a
     published surface; everything else is import-path rename on already-allowed layers.
