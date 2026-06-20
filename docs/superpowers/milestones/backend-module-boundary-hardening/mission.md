# Mission: Backend Module-Boundary Hardening

> **Status:** Drafting
> **Date:** 2026-06-20  ·  **Branch of record:** `main` (forks from HEAD `d7d53590`)
> **Type:** remediation
> **Slug:** `backend-module-boundary-hardening`  ·  **Owner / operator:** leandrotca
> **Evidence base:** `./discovery-brief.md`  ·  **Program index:** `./README.md`
> **Governs:** Milestones M0..M4 below. Each milestone gets its own plan via the `milestone` skill,
> executed in a fresh session. This file is the **stable governing contract** — it says *what* the
> mission is and *what proves it done*; it contains **no execution detail**.

---

## 1. Problem / why now

The parent `grade-a-completion` mission's post-M9 terminal re-audit
(`wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m9.md` §6/§10) is **met-on-bar** under the
canonical §6 greps (Contract/API B+→A−, H-D=0, the five-repeat miss closed), but it surfaced **pre-existing
module-boundary debt the §8 gate never measured**: cross-module raw SQL reads of another module's owned base
table, plus hardcoded foreign domain-state literals. The operator's decision: **Grade-A is HELD** until that
debt is eliminated — a genuinely high-level backend, not a label.

An exhaustive census (`./discovery-brief.md`) found the re-audit's "14" was itself an undercount: the real
set is **~20 sites in 3 categories**, and the re-audit mis-attributed two of them (`document_process_areas`
is taxonomy-owned, not iam; `search/reader.go` crosses two modules). Decomposing on that flawed "14" would
repeat the under-counting failure that caused the five Contract/API misses. This mission stands on the
corrected census instead.

## 2. Goals / Non-Goals

**Goals**
- module-boundaries / DDD dimension reaches **A** in the re-run F5.1 re-audit.
- **H-G = 0** under BOTH readings: the canonical §6 greps AND the broad "any cross-module owned-**base**-table
  read" reading, as reconciled by ADR-0039's published-contract exemption.
- An ADR (0039) that makes the done-bar **unambiguous**: a named definition + exemption list, so "compliant"
  vs "violation" is mechanical, not judgement.
- A CI/cilint grep-guard that fails the build on any new raw cross-module base-table read (the class can't
  silently re-open).
- Every ported site has a **parity test** proving identical results before its raw SQL is deleted.

**Non-Goals**
- No change to runtime behavior, query results, or visibility/authz semantics — this is a **seam** change,
  not a logic change. Parity tests enforce this.
- No new features, no schema migrations beyond the published view(s) the C-α mechanism requires.
- No touching auth/security cross-module reads — already ported (M4 work holds; confirmed in discovery).
- No widening to owned-table classes outside the H-G scope **until** M0's census re-scopes against ADR-0039
  (then in-scope items are addressed; genuinely out-of-scope ones are recorded as such).
- Not flipping the parent `grade-a-completion` to Grade-A — that sign-off is the operator's, gated on this
  mission's terminal PASS.

## 3. Locked decisions (operator-approved)

| # | Decision | Value |
|---|----------|-------|
| D1 | Category C mechanism | **C-α — published view + exemption.** iam publishes a versioned active-membership view; CD publishes a visibility projection for search. ADR-0039 declares: reading another module's **published view / read-model = compliant**; reading its **base table = violation**. Preserves set-based SQL (no N+1), sidesteps H-PRE-1 (object name changes, tx/lock structure does not), makes "H-G=0 under both readings" coherent by refining "owned table" → "owned **base** table". |
| D2 | Program placement | **Sibling mission that GATES** the parent `grade-a-completion`'s Grade-A sign-off. Parent terminal acceptance is met-on-bar; Grade-A is HELD pending this mission's terminal PASS. |
| D3 | C4 scope | **In-scope, as its own risk-isolated milestone (M4), sequenced last.** search consumes a CD-published visibility contract instead of inlining CD's whole predicate. Highest HS-2 risk; isolated so it cannot regress M0–M3. |
| D4 | ADR-0039 exemption list | Compliant: (a) JOIN/read of another module's **published, versioned view**; (b) call through an owner-**published read-port**; (c) reading one's **own** tables. Violation: raw `SELECT`/`JOIN`/`EXISTS` against another module's **base table**. Active-now membership view encodes exactly `effective_to IS NULL` (ADR 0037), no reinterpretation. |
| D5 | Sequencing | **Definition first, risk last.** M0 (ADR+census) → M1 (trivial typed constants) → M2 (contained read-ports) → M3 (contained membership view + consumption incl. H-PRE-1 site) → M4 (search visibility-contract, the redesign-risk item). Late milestones cannot regress the grade because the bar-defining work is already locked. |
| D6 | Parity discipline | Per-site **parity test** (raw-SQL result == ported-path result) committed and green **before** the raw SQL is deleted. Non-negotiable per site. |

## 4. Discovery summary

An exhaustive 3-agent census plus a definitive main-session token grep (`./discovery-brief.md`) mapped
**~20 cross-module reads in 3 categories** across `internal/modules/**` (non-test), with each site's owner
and tx/lock context **verified in source** this session. Confidence is high on the inventory and ownership;
**assumed** (deferred to M0's binding census) is only that no read hides behind a dynamic/aliased table name
and that owned-table classes beyond the named H-G tokens are absent. Auth/security confirmed already ported.
Every milestone below traces to a Category A/B/C finding in the brief.

## 5. Work / requirement inventory

| # | Item (site) | Class / kind | Milestone |
|---|-------------|--------------|-----------|
| 1 | ADR-0039: H-G definition + exemption list + CI grep-guard | definition / gate | M0 |
| 2 | Binding re-census against ADR-0039 (drain check, re-scope owned-table set) | census | M0 |
| 3 | `controlleddocuments/domain/resolution.go:42,55,58` — `"published"`/`"obsolete"` literals | A — stringly-typed (no SQL) | M1 |
| 4 | `documents/repository/repository.go:1701` — reads `controlled_documents` (profile_code) | B — foreign point-read | M2 |
| 5 | `controlleddocuments/infrastructure/repository.go:532` — reads `document_revisions` | B — foreign point-read | M2 |
| 6 | `controlleddocuments/infrastructure/repository.go:539,545` — reads `documents` (+literals :542,:548) | B — foreign point-read | M2 |
| 7 | `controlleddocuments/infrastructure/repository.go:593` — reads `approval_instances` (+literal :596) | B — foreign point-read | M2 |
| 8 | `documents/application/document_area.go:37` — reads `controlled_documents` (in-tx) | B — foreign point-read (tx-aware) | M2 |
| 9 | `documents/approval/application/read_service.go:355` — reads `controlled_documents` (in-tx) | B — foreign point-read (tx-aware) | M2 |
| 10 | `documents/repository/repository.go:154` — reads `document_process_areas` (taxonomy, in-tx) | B — foreign point-read (tx-aware) | M2 |
| 11 | `iam/infrastructure/postgres/area_catalog_reader.go:28` — reads `document_process_areas` (taxonomy) | B — foreign point-read | M2 |
| 12 | `controlleddocuments/infrastructure/repository.go:150` — `user_process_areas` EXISTS (list visibility) | C — authz-visibility membership | M3 |
| 13 | `controlleddocuments/infrastructure/repository.go:492` — `user_process_areas` EXISTS (CanRead) | C — authz-visibility membership | M3 |
| 14 | `documents/approval/repository/postgres_approval_repository.go:1136` — `user_process_areas` (in-tx, H-PRE-1) | C — authz-visibility membership (tx-aware) | M3 |
| 15 | `search/infrastructure/v2documents/reader.go:70,97,102,111` — inlines CD's whole visibility predicate (CD+iam) (+literals :92,:94) | C4 — visibility-contract redesign | M4 |

Out-of-scope (recorded, with reason): auth/security cross-module reads — **already ported** (no live SQL,
discovery-confirmed). Any owned-table read **outside** the H-G token set is deferred to M0's re-scope; if M0
surfaces new in-scope sites, they are added to the relevant milestone (HS-6 if it changes milestone shape).

## 6. Program architecture (by reference)

This mission executes via the `milestone` skill. The per-feature close-out loop, the per-feature
consumer-contract spec gate, and the per-milestone `milestone-validator` gate are defined there — **not
duplicated here**. See `.claude/skills/milestone/SKILL.md`. Program-scale shape only:

```
Mission: backend-module-boundary-hardening
└── Milestone (M0..M4)          ── each: features → milestone-validator gate → HS-1 operator gate
    └── Feature (Fx.y)          ── each: spec(consumer-contract) → plan → TDD → evidence
Terminal acceptance (§8)        ── main-session re-audit fan-out + independent mission-validator judge
```

## 7. Milestones

### M0 — ADR-0039: lock the definition + binding census
**Objective:** Make the done-bar mechanical before any code moves — author the ADR that defines H-G + the
exemption list, then re-census against that definition so the work inventory is authoritative (no third
undercount).

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F0.1 adr-0039 | ADR `wiki/decisions/0039-*.md`: H-G definition (raw read of another module's **base table** = violation), exemption list (D4), the active-now membership view contract, H-PRE-1 note. Supersession/links to ADR 0022/0037/0038. | ADR exists, status Accepted; definition + exemption list are unambiguous (a reviewer can classify any site mechanically); operator-ratified in the M0 HS-1 gate. |
| F0.2 binding-census | Re-run the cross-module read census against ADR-0039's definition; produce the authoritative in-scope site list + a coverage statement (re-scope owned-table set beyond the named tokens). | Census reproduces the ~20 brief sites; any new in-scope site is added to its milestone or recorded out-of-scope with reason; 0 sites left unclassified. |
| F0.3 cilint-guard | A cilint/CI grep-guard analyzer that fails on a raw cross-module base-table read outside the exemption allowlist (the H-G analyzer, sibling to the H-D noresponsemap guard). | Guard flags every current in-scope site (red on today's tree); allowlist holds the exemptions; `go run ./tools/cilint ./...` wired. |

**Milestone gate:** Contract/architecture-truth checklist. Proves the definition is locked and the census is
exhaustive **before** remediation — the root-cause fix for the undercount. Validated by `milestone-validator`.

### M1 — Category A: typed status constants
**Objective:** Eliminate the 3 stringly-typed foreign domain-state literals in CD domain logic (no SQL).

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F1.1 resolution-constants | `controlleddocuments/domain/resolution.go:42,55,58` use `templates/domain` `VersionStatus*` constants (precedent: `service.go:283`). | `go build ./...` green; existing resolution unit tests green; grep shows 0 bare `"published"`/`"obsolete"` literals in `resolution.go`; cilint guard unaffected. |

**Milestone gate:** Code-quality checklist; no behavior change (unit tests identical). `milestone-validator`.

### M2 — Category B: owner-published read-ports
**Objective:** Replace the 8 clean foreign point-reads with owner-published read-ports; tx-aware where the
read is in-tx. Each port consumed behind a parity test, then raw SQL deleted.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F2.1 cd-read-port | CD publishes a read-port for the profile/CD point-reads consumed by documents (items 4, 8, 9). tx-aware variant for the in-tx callers. | Parity test per site (raw == port) green **before** deletion; `go build`/`go test ./...` green; cilint guard clears items 4,8,9. |
| F2.2 documents-read-port | documents/approval publish read-ports for the instance/revision/approval reads consumed by CD (items 5, 6, 7); foreign status literals replaced with documents/approval typed constants. | Parity test per site green before deletion; build/tests green; guard clears 5,6,7; 0 bare status literals in those queries. |
| F2.3 taxonomy-read-port | taxonomy publishes an area read-port (name + existence) consumed by documents and iam (items 10, 11); tx-aware for the in-tx caller. | Parity test per site green before deletion; build/tests green; guard clears 10,11. |

**Milestone gate:** Persistence + module-boundaries checklists; parity proven per site; no behavior change.
`milestone-validator`.

### M3 — Category C contained: published membership view + consumption
**Objective:** iam publishes the active-membership view (C-α); CD list/CanRead and approval (in-tx, H-PRE-1)
consume it instead of `user_process_areas` base table. search (C4) is deliberately deferred to M4.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F3.1 membership-view | iam publishes `metaldocs.v_active_user_areas` (migration) encoding exactly `effective_to IS NULL` (ADR 0037); documented as the published contract per ADR-0039. | View migration applies; view rows == base-table active-now rows (parity query); ADR-0039 exemption references it. |
| F3.2 cd-consume-view | CD `repository.go:150` (list) and `:492` (CanRead) JOIN the published view instead of `user_process_areas`. | Parity test (list + CanRead results identical pre/post) green before deletion; build/tests green; guard clears 12,13; set-based SQL preserved (no per-row loop). |
| F3.3 approval-consume-view | approval `postgres_approval_repository.go:1136` `ResolveEligibleActors` reads the published view **in the existing tx** (tx-aware, no recording call — H-PRE-1 safe). | Parity test green before deletion; build/integration tests green; guard clears 14; reviewer confirms no authz-recording read added inside the lock-holding tx. |

**Milestone gate:** module-boundaries + authz-invariant (H-PRE-1) checklists; parity per site; set-based +
tx/lock structure unchanged. `milestone-validator`.

### M4 — C4: search consumes CD-published visibility contract (risk-isolated, last)
**Objective:** search stops inlining CD's entire visibility predicate; it consumes a CD-published visibility
contract (view/read-model) joined with the iam membership view. Sequenced last so its redesign risk cannot
regress M0–M3.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F4.1 cd-visibility-contract | CD publishes a visibility projection/view expressing "actor can see CD" (company/restricted/owner/area-grant/user-grant), per ADR-0039. | View migration applies; visibility rows == current inline-predicate results (parity query across scopes). |
| F4.2 search-consume | search `reader.go` JOINs the CD visibility contract + iam membership view instead of `controlled_documents`/`controlled_document_*_grants`/`user_process_areas` base tables; foreign literals :92,:94 removed. | Parity test (search results identical pre/post across visibility scopes) green before deletion; build/integration tests green; guard clears item 15; **if** this requires a cross-module API redesign beyond a contained view-consume → **HS-2 stop** and surface before building. |

**Milestone gate:** module-boundaries checklist at full scope; search result parity; HS-2 boundary respected.
`milestone-validator`.

## 8. ★ Terminal acceptance — definition of done (written up front)

- **Pass bar (the mission shall be X):** in a fresh re-run of the F5.1 10-dimension architecture re-audit,
  **module-boundaries / DDD = A** (up from the post-M9 A−/B+ variance), and **H-G = 0 under BOTH readings** —
  the canonical §6 greps AND the broad "any cross-module owned-**base**-table read" reading as defined by
  ADR-0039 — with **0 skeptic-confirmed new Critical/Major** introduced by the remediation, and no regression
  in any other dimension (all remain ≥ their post-M9 grade).
- **What to validate:**
  1. ADR-0039 exists, Accepted, with an unambiguous definition + exemption list.
  2. The cilint/CI H-G grep-guard is green on the full tree (0 violations outside the published-contract
     allowlist).
  3. Every inventory item (§5 rows 3–15) is either ported (raw base-table SQL deleted, parity test green) or
     explicitly exempted by ADR-0039 (published view / read-port).
  4. The re-run re-audit grades module-boundaries = A and reports H-G = 0 under both readings.
  5. `go build ./...` and `go test ./...` green; integration suite green where a docker Postgres (:5433) is
     available, else explicitly noted as not-run (no false green).
  6. 0 new Critical/Major in the re-audit, each candidate skeptic-confirmed.
- **How to validate:**
  - **Fan-out (main session):** re-run the F5.1 10-dimension re-audit Workflow from clean state, capturing
    `wiki/backend/_artifacts/architecture-re-audit-<date>-post-boundary-hardening.md`, with the per-finding
    skeptic rule. Then dispatch `mission-validator` to **judge that artifact** against the pass bar.
  - **Deterministic (mission-validator runs itself):** `go run ./tools/cilint ./...` (H-G guard green);
    `git grep` the §5 site tokens to confirm 0 raw base-table reads remain outside the allowlist; re-run the
    named parity tests; `go build ./...` + `go test ./...`. The validator independently spot-checks the
    re-audit's "0 remaining" claims by re-grepping a sample of the §5 sites.
- **Who validates:** the independent `mission-validator` subagent (`.claude/agents/mission-validator.md`). It
  judges and writes `qa/mission-validation.md` only — never edits code or flips status. The Grade-A sign-off
  on the parent remains the operator's.
- **On miss (HS-5):** the missed criteria become a bounded remediation micro-milestone run through
  `milestone`, then `mission-validator` is re-dispatched. The operator decides continue vs replan at each
  loop. **Note:** the parent's HS-5 six-consecutive-miss stop rule is inherited — a terminal miss here
  surfaces to the operator and does **not** auto-open further work.

## 9. Hard-stop catalog

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary (M0..M4) | Operator review gate; no next milestone / no merge without approval |
| HS-2 | A fix implies redesign outside the assigned boundary (esp. M4 search visibility-contract, or any port that turns out to need a shared-API redesign) | Stop; report the boundary + minimum prerequisite plan; no symptom-patch |
| HS-3 | A prerequisite boundary fails (build/runnable; docker Postgres :5433 unavailable for an integration parity test) | Repair / note the gap; rerun the checkpoint; resume. Never false-green a skipped parity test |
| HS-4 | A `milestone-validator` returns FAIL | Open the named fix feature; re-run its lifecycle; re-dispatch the validator |
| HS-5 | The terminal `mission-validator` misses the §8 bar | Bounded remediation micro-milestone; re-dispatch; operator decides continue vs replan |
| HS-6 | Scope drift — M0 census surfaces in-scope sites that change a milestone's shape, or a port balloons | Stop; surface the deviation; re-interview / replan before continuing |
| HS-PRE-1 | A port would place an authz-recording read inside a lock-holding atomic tx | Forbidden; keep the read off the lock-holding tx or use the published view (SELECT-only) |

## 10. Constraints respected

- **ADR 0022** (authz root cause — never symptom-patch authz visibility), **ADR 0037** (membership temporal
  model, `effective_to IS NULL` active-now; the published view encodes exactly this), **ADR 0038**
  (FamilyCodeResolver port — precedent for owner-published ports). New ADR = **0039**.
- **H-PRE-1** advisory-lock rule (memory): no authz-recording read inside a lock-holding tx; the C-α view is
  SELECT-only and tx-structure-neutral.
- **Parity-before-delete** (D6): no raw SQL removed without a green parity test.
- House rules: PowerShell for any local startup (never bash/`source .env`); never read/print `.env`;
  contract-first regen order where api.gen is touched; **no merge by agent**; **never push** (commits local
  only, CLAUDE.md §5.0 standing authorization to commit).
- Skill routing: backend/contract work reads `wiki/architecture/backend-api-structure.md` +
  `api-contract.md`; persistence work reads `wiki/database/index.md`; QA close-out reads
  `wiki/quality/qa-operating-system.md`.

## 11. Execution model

One `mission.md` governs all milestones. Each milestone → its own plan via `milestone`, executed in a
**fresh session**, subagent-driven. Operator gate between every milestone (HS-1); **no merge by the agent**.
Token discipline: parallel fan-out only where it pays (M0 census, terminal re-audit); everything else direct
tools. Model policy: sonnet analysis/review, haiku mechanical, never fable workers, ≤15 concurrent.

## 12. End-state / reconciliation

Fill only when the last milestone has passed and the terminal gate is green:
- [ ] Every planned feature (M0..M4) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance (§8) passed — link `qa/mission-validation.md`.
- [ ] Parent `grade-a-completion` Grade-A sign-off unblocked and presented to operator.
- [ ] Operator sign-off: <date / name>
