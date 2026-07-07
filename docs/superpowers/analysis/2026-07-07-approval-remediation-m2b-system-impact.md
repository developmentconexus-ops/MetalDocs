# System-impact analysis — Approval system remediation (M2b program input)

**Date:** 2026-07-07
**Intent (one line):** Fix the approval system to professional-eQMS grade: workflow-model defects (W1–W13), permission gaps (P1–P8), FE cockpit redesign as editor-shell + approval sidebar, review/suggestion mode, supervisory oversight capability.
**Work type:** feature (multi-feature remediation inside existing modules; no new top-level module)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

> Source evidence: 6-front adversarial review (2026-07-07, this session) — access map, backend
> domain map, DB/route evidence, contracts/wiki/ADR sweep, FE audit (impeccable), industry
> benchmark. Finding IDs W1–W13 (workflow), P1–P8 (permissions), FE P0–P2 reference that review.

---

## 1. Classify & own

- **Work type:** feature (remediation program). No module birth.
- **Owning module(s):**
  - `documents/approval` (nested exception, ADR 0072) — workflow model: route versioning (W1),
    content freeze/invalidation (W2), stage capability enforcement (W3), SLA/escalation (W4),
    delegation (W5), pool validation (W6), SoD unification (W7), signature meaning (W8),
    minors (W9–W13), permission fixes P1/P3/P4/P6.
  - `iam` — new supervisory capability (working name `approval.oversee`, P5) + grants; consumed by
    approval via `authz.Require` and by FE route gating (P2/P8).
  - `frontend/apps/web` `features/documents` + `features/approvals` — cockpit = editor shell +
    approval sidebar; worklist consolidation; FE P0–P2 fixes.
  - `docx-v2` workspace (+ `third_party/eigenpal` vendor) — suggestion/comment mode surface the
    approval sidebar drives.
  - `notifications` / `jobs` — reminder + escalation events (W4) ride the existing fanout/River
    maintenance machinery.
- **Explicitly NOT owning:** `controlleddocuments` (lifecycle states stay as-is; only transition
  *triggers* change), `templates` (template approval is a separate consumer; parity later),
  `distribution`, `render` (PDF path untouched).
- **Cross-module edges (direction):**
  - `documents/approval → iam` via `authz.Require` + capability constants (published surface) — existing edge.
  - `documents/approval → notifications` via outbox events (published event contract) — existing edge, new event types.
  - `jobs → documents/approval` application service for SLA scan (published service iface) — pattern of `document_review_surfacer`.
  - FE `features/approvals → features/documents` editor shell — FE-internal composition, plus docx-v2 adapter → eigenpal editor API.
  - No new Go module-boundary edges; ADR 0072 intra-family exception covers documents↔approval.
- **Ambiguity?** None — AS-3 not triggered. (Suggestion-persistence entity ownership is a design
  question inside `documents/approval` vs `documents`; both are the same ADR 0072 family, so not a
  boundary ambiguity.)

## 2. Foundation verdict

- **Base:** signature/snapshot/quorum core is Grade-A quality — real Part 11 re-auth (bcrypt,
  rate-limited, fail-closed), frozen eligibility snapshots with drift policies, expressive quorum,
  two-layer SoD, OCC + idempotency throughout, in-tx governance events. This core is **sound** and
  is kept.
- **Defective substrate the work must replace (not patch):**
  1. **W1** — `enforce_route_immutable` + non-partial `UNIQUE(tenant_id, profile_code)` make a
     route permanently frozen after first use; ADR 0018 promises the opposite. Global-maximum
     structure: **versioned route definitions** (immutable version rows; instances pin a version;
     head version editable/replaceable). Patching the trigger condition alone would be a local
     maximum — rejected.
  2. **W2** — signoff hash pin floats to head revision (`content_hash_at_submit` has no production
     writer) while edits are allowed during `under_review`. Global-maximum structure: **content
     freeze during active workflow + review-layer (suggestions/comments) for approver input**, one
     canonical hash chain submit→signatures→publish. Patching the COALESCE would be a local
     maximum — rejected.
  3. **FE cockpit** — standalone page, alien slate palette, hand-rolled polling adapter, duplicated
     derivations. Global-maximum structure: **one artifact surface** — reuse the editor shell
     (document canvas + sidebar slot), sidebar mode = author vs approver (doctrine R5 of the
     lifecycle-ux spec: one implementation, N entry points). Polishing the current standalone page
     would be AS-2 territory — rejected; the redesign IS the work.
- **AS-2 status:** not triggered — the proposal is the global-maximum structure in all three spots;
  no optimization inside a patch is planned.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes, heavily** | New `approval.oversee` cap (not a ROOT role — operator's "super user" request reframed as capability, §4); W3 makes per-stage `required_capability` an `authz.Require` input; P1 adds explicit tier-1 rows for stage-scoped signoff/cancel; visibility gating (P2/P3/P8) expressed as eligibility ∪ oversee capability | `authz.Require`, `permissions.go` rules, `capability_scope.go` |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** | Every new/changed route (route versions admin, delegation, suggestions, oversight list, due-date fields) spec-first in `api/openapi/v1/openapi.yaml` → regenerate | `oapi-codegen` per-module `cfg.yaml`/`gen.go` |
| Multi-tenant pooled | **Yes** | New tenant tables (route versions, delegations, suggestions, SLA/due columns) all carry `tenant_id` + RLS per 0285-class pattern; cross-tenant → 404 | `tenant.FromContext`, `authz.SeedTxIdentity` |
| Async = transactional outbox | **Yes** | Reminder/escalation notifications (W4) enqueue in business tx; River periodic job scans due stages (pattern: `document_review_surfacer`); no inline network calls | outbox repo, `jobs/maintenance/periodic.go` dual-define (ADR 0067) |
| DB enforces invariants | **Yes** | Route-version immutability trigger (replaces `enforce_route_immutable`); content-freeze backstop (block content-revision writes while active instance exists, or invalidation trigger); SoD unified rule mirrored in `enforce_signoff_sod`; non-empty-pool CHECK at instance insert; partial unique on active route version | baseline trigger patterns; migration `0NNN` |
| Cross-module via published interface only | **Yes** | approval→iam and jobs→approval stay on published surfaces; ADR 0072 family exception covers documents↔approval; eigenpal is FE-vendor only (no Go edge) | existing ports |

**AS-1:** none — the remediation *restores* invariant coherence (W3 currently violates
capabilities-not-roles in spirit; P1 violates tier coherence). No planned violation.

## 4. Capability wiring — `approval.oversee` (name final at design)

1. const + `validCapabilities` — add in `iam/domain/model.go`.
2. scope classify — decision at design: **ScopeTenant** recommended (supervisory cross-area
   visibility is its point); revisit if per-area QA leads are wanted.
3. tier-1 route→cap — new oversight list route mapped explicitly; **plus P1 fix**: explicit tier-1
   rows for `POST /approval/instances/*/stages/*/signoffs` (`document.signoff`) and
   `POST /approval/instances/*/cancel` (`document.edit`) replacing the generic `/approval/` prefix
   fallback (same recipe as BE-9/ADR 0022 F4).
4. tier-2 — `authz.Require` in read/oversight service; instance-read visibility becomes
   (eligible-actor ∪ submitter ∪ oversee) instead of bare `document.view` tenant sentinel (P3).
5. seed grants — grant to `qms_admin`-class roles in reference data; fix dev-seed `approver`
   superset drift (P7) same pass.
6. DB tripwire — cap name passes `ck_cap_format`/`ck_cap_not_legacy` (dot-namespaced, fine).
7. guard tests — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` updated.
8. bump `TestCapabilityRegistrySize` — targeted-verify current count at implementation time.
9. CI capability-coherence (REQ-AUTHZ-5) — generated tripwire arms (GMR M2 machinery) pick the new
   cap up from the Go registry; drift lints must stay green.
10. H-PRE-1 — oversight reads that record audit stay off any lock-holding tx (standing constraint).

W3 note: per-stage `required_capability` enforcement reuses **existing** caps as `authz.Require`
input (no second new cap unless design adds `document.review` distinct-reviewer stage kind — if so,
it walks these 10 touchpoints too and bumps the registry again).

## 5. Module wiring

**N/A** — no module birth. `documents/approval` nested exception (ADR 0072) already hosts the
workflow; new tables live in its migrations; new routes in its OpenAPI tag.

## 6. Frameworks to reuse, not reinvent

- `TxRunner` (`Do`) — all new services (delegation, route-version admin, suggestion persistence, SLA scan writes). Note `authz.Require` needs writable tx (G1).
- `tenant.FromContext` + `authz.SeedTxIdentity` — every business tx.
- `authz.Require` — tier-2 everywhere incl. W3 per-stage input.
- `problem.New/Write` — all new errors; extend `signoffErrors` FE mapping (and unify its mixed code conventions while touching it — FE P2).
- `audit.RecordTx` + governance events — delegation grants/uses, route version publishes, suggestion resolution, escalation events.
- Outbox repo — reminder/escalation/notification side effects.
- River periodic job (ADR 0067 dual-define) — due-date/SLA scanner alongside `stuck-instance-watchdog` (which stays, alert-only per ADR 0068).
- `testdb` factory — all integration tests.
- **FE:** TanStack Query replaces the hand-rolled `setInterval` adapter (FE P0); shared `Dialog`; wine tokens from `src/styles/tokens.css` (slate palette dies); editor shell composition from `features/documents`.
- **Eigenpal editor (docx-v2)** — suggestion/comment mode: **capability unverified** (tarball not
  expanded in `third_party/eigenpal/`). Design-time verification REQUIRED before committing to the
  suggestion-mode scope (locked constraint, §10).

## 7. Contract & data

- **OpenAPI-first:** new/changed under the `approval` tag — route-version CRUD (replaces frozen
  single-row model), delegation endpoints, suggestion/comment endpoints (or reuse of existing
  comment surface — design decision), oversight instance list, due-date fields on instance/stage
  DTOs, meaning-of-signature field in signoff request/record. All spec → regen, never hand-added.
- **Migrations (expand/contract, tenant-scoped, RLS):**
  - `approval_route_versions` (immutable rows) + repoint `approval_instances.route_version` FK
    semantics; partial `UNIQUE (tenant_id, profile_code) WHERE active` for head; retire
    `enforce_route_immutable` in favor of version-row immutability trigger. **Expand/contract:**
    backfill existing routes as version 1; instances already store `route_version_snapshot`.
  - Content-freeze enforcement: trigger blocking `document_revisions`/working-content writes while
    an `in_progress` instance exists (or signoff-invalidation table) — design decides which; DB
    backstop mandatory either way.
  - `approval_delegations` (tenant, delegator, delegate, window, reason; audited).
  - `due_at` on `approval_stage_instances` (+ policy columns on route-version stages).
  - Suggestion persistence (entity TBD at design; tenant_id + document/instance keys).
  - Signature payload extension: `meaning` + signer role snapshot (additive JSONB — no contract break).
  - SoD: exclude author+submitter from `eligible_actor_ids` at snapshot; align
    `enforce_signoff_sod` to the unified rule; non-empty-pool check at submit.
- **Destructive?** Route single-row → versioned is the one structural change; expand/contract with
  backfill, no live-contract break in one step.

## 8. Test & QA plan

- **Framework:** `testdb` factory, `//go:build integration`, R1–R4 discipline. Real-DB suites for:
  route versioning lifecycle, freeze/invalidation, SoD pool exclusion, empty-pool submit rejection,
  delegation windows, SLA scanner, oversee visibility (positive + negative + cross-tenant 404).
- **QA gates applying (feature subset):** contract (spec↔handlers parity), authz (tier-1/tier-2
  coherence + tripwire parity lints), multi-tenant isolation (new tables, RLS-truth with
  `metaldocs_ci` non-owner role per M7 sweep), async/idempotency (reminder consumer idempotent),
  DB-invariant (new triggers), docs. All six in scope given breadth.
- **Evidence:** `go build ./...`, `go test ./...`, integration suite, FE `make test`,
  `check-system-runnable.ps1`, live docker QA driving the redesigned cockpit as a user
  (per operator's standing verify-live discipline), impeccable re-audit of the new screen.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/approval.md` (+ tech-debt sister: close the divergence rows this
  work fixes, open rows for anything deferred), `wiki/concepts/approval-routes.md` (versioning
  semantics replace frozen-route text), `wiki/workflows/approval.md` (TBD edge cases resolved:
  reassignment, mid-route edit policy), FE structure doc; refresh `Last verified` stamps.
- **REQ IDs:** REQ-AUTHZ-5 (capability coherence), plus route/versioning + tenancy REQ rows cited
  at review time from `wiki/architecture/backend-target-architecture.md`.
- **ADRs required (Yellow driver):**
  1. **Route definition versioning** — supersedes/amends **ADR 0018** (whose §1/§3 text the current
     schema already contradicts).
  2. **Content freeze + review-layer during active workflow** — policy change to document editing
     semantics (touches ADR 0015 pin chain, ADR 0069 reason-for-change adjacency).
  3. **`approval.oversee` capability + visibility model** — extends ADR 0022 catalogue; replaces
     the tenant-sentinel read (P3) with eligibility∪oversee.
  4. Delegation-of-authority model (may fold into #3 or stand alone — design decides).

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceeds to brainstorming with flagged risks; no unresolved hard-stop.
- **Open hard-stops:** none. AS-1 none (work restores invariants), AS-2 none (global-maximum
  structures named and chosen), AS-3 none.
- **Yellow risks carried into design:**
  1. **Eigenpal suggestion-mode capability is UNVERIFIED** (vendor tarball not expanded). Design
     must verify before locking suggestion-mode scope; fallback = MetalDocs-owned comment/suggestion
     layer over the docx viewer.
  2. **Scope is program-scale** — W1–W13 + P1–P8 + FE redesign will not fit one milestone-sized
     change safely; brainstorming must decompose (operator asked for "M2b" — expect M2b + siblings
     or a staged milestone).
  3. Three-to-four **ADRs required** (§9) — design output includes them.
- **Locked constraints handed to brainstorming:**
  - Oversight is a **capability** (`approval.oversee`-class), never a ROOT role (ADR 0022).
  - Route change = **versioned immutable definitions**; no trigger-condition patch.
  - W2 fix = **freeze + review layer**, one canonical hash chain; no COALESCE patch.
  - FE = **editor-shell reuse** (one artifact surface, sidebar slot by mode); no standalone-cockpit polish.
  - New caps ⇒ registry-size bump + generated-tripwire/lint parity (GMR M2 machinery).
  - Contract-first for every route; expand/contract for the route-versioning migration.
  - Suggestion mode contingent on eigenpal verification (risk 1).
  - Approver worklist single destination = redesigned cockpit (kills the 3-destination drift).
  - Watchdog stays alert-only (ADR 0068); SLA/escalation is new machinery beside it, not a change to it.
