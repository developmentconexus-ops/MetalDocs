# ADR 0039 — Cross-module read boundary: a non-owner may read another module's **published view / read-port**, never its **base table** (H-G definition)

> **Status:** Accepted 2026-06-20 (amended 2026-06-20 — M0/F0.2 HS-6 ruling added exemptions D3(d)–(f) for platform-sink / parent-ADR-dispositioned-auth / worker-layer reads; see Amendment section)
> **Last verified:** 2026-06-20
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `backend-module-boundary-hardening` · Milestone M0 · Feature F0.1. This mission GATES the parent `grade-a-completion` Grade-A sign-off (met-on-bar, Grade-A HELD).
> **Supersedes:** none.
> **Related ADRs:** [0022 — Authz capability coherence](./0022-authz-capability-coherence.md) (never symptom-patch authz visibility); [0037 — Membership temporal model](./0037-membership-temporal-model.md) (active-now ⟺ `effective_to IS NULL`); [0038 — `FamilyCodeResolver` port](./0038-family-code-resolver-port.md) and [0029](./0029-user-display-name-reader-port.md)/[0030](./0030-template-version-state-port.md)/[0031](./0031-tenant-user-reader-port.md) (owner-published read-port precedent — the compliant mechanism this ADR generalizes).
> **Related code (Last verified 2026-06-21):**
> - `tools/cilint/internal/analyzers/` — the H-G analyzer (`hgcrossmodule`) that mechanizes this ADR (M0/F0.3), sibling of `noresponsemap.go` (H-D).
> - The in-scope sites this ADR classifies: mission `docs/superpowers/milestones/backend-module-boundary-hardening/mission.md` §5, evidence base `…/discovery-brief.md`.
> - `db/migrations/0242_iam_v_active_user_areas_view.sql` — the D3(a)/D4 iam-published active-membership view `metaldocs.v_active_user_areas` (`effective_to IS NULL`; columns `tenant_id, user_id, area_code, role`), built in M3/F3.1. The C1–C3 consumers (mission §5 rows 12–14) read this view.
> - `db/migrations/0243_cd_search_visibility_contract.sql` — the D3(a)/D4 controlleddocuments-published search read contract, built in M4/F4.1: `metaldocs.v_cd_search_facts` (1 row/CD: projection cols `code, department_code, profile_code, sequence_num` + `is_company` + `owner_user_id`) and `metaldocs.v_cd_grantee` (bounded restricted-CD visibility edges via `controlled_document_area_grants` ⋈ `v_active_user_areas` ∪ `controlled_document_user_grants`). The C4b/C4c/C4e consumer (search `v2documents/reader.go`, mission §5) reads these in M4/F4.3 instead of CD's base tables.
> - `db/migrations/0244_documents_search_projection.sql` — the D3(a)/D4 documents-published search projection contract, built in M4/F4.2: `metaldocs.v_document_search_facts`, a pure 1:1 projection of `public.documents` (the 14 columns search reads; no WHERE, no COALESCE — `archived_at` exposed, consumer applies its own active filter). The C4a consumer (search `v2documents/reader.go`, mission §5) reads this in M4/F4.3 instead of `public.documents`.
> - `db/migrations/0245_cd_obligated_readers_view.sql` — the D3(a)/D4 controlleddocuments-published obligated-reader read contract, built in mission `frontend-screen-completion` M2/F2.1a (ADR-0040): `metaldocs.v_cd_obligated_readers` (one row per `(tenant_id, controlled_document_id, user_id)` obligated to read a CD, with `area_code` + `source ∈ {user_grant, area_grant, company_scope}`; three legs UNION'd, DISTINCT BY `(tenant_id, cd, user_id)` with source precedence `user_grant > area_grant > company_scope`). Consumer module: `distribution` (built in M2/F2.1c + F2.2) — reads THIS view instead of CD's grant base tables. `metaldocs.v_cd_grantee` remains restricted-only (search semantics, migration 0243) — see ADR-0040 for the new-sibling-vs-extend rationale.
> - `db/migrations/0246_taxonomy_process_area_name_view.sql` — the D3(a)/D4 taxonomy-published per-area label read contract, built in mission `frontend-screen-completion` M2/F2.1b (ADR-0041): `metaldocs.v_process_area_name` (one `(tenant_id, area_code, area_name)` row per `(tenant_id, area_code)`, 1:1 projection of `metaldocs.document_process_areas` with `code → area_code` + `name → area_name`; no `is_active` / `archived_at` filter — parity with the existing in-Go `AreaCatalogReader` port). Consumer module: `distribution` (built in M2/F2.1c + F2.2) — joins this view to `v_cd_obligated_readers` on `(tenant_id, area_code)` to populate `DistributionRecipient.area_name` + `DistributionAreaCoverage.area_name`.

---

## Context

The parent `grade-a-completion` post-M9 terminal re-audit
(`wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m9.md` §6/§10) is **met-on-bar** under the
canonical §6 greps, but surfaced **module-boundary debt the §8 gate never measured**: a module reaching
into another module's **owned base table** with raw SQL (`SELECT`/`JOIN`/`EXISTS`), plus hardcoded foreign
domain-state literals. This is the **H-G** class — the same family ADRs 0029/0030/0031/0038 each closed for a
single table by introducing an owner-published port.

Two problems blocked remediation and motivate this ADR:

1. **The done-bar was a judgement call, not a rule.** The re-audit counted "14" H-G sites under a broad
   sweep; an exhaustive census (`…/discovery-brief.md`) found **~20** and *corrected two ownership facts*
   (`document_process_areas` is **taxonomy**-owned, not iam; `search/reader.go` crosses **two** modules).
   Remediating against an undercounted, disputed bar would repeat the five-consecutive Contract/API miss that
   caused this mission to exist.

2. **"Owned table" was ambiguous on the compliant side.** The strict reading — "no cross-module SQL at all"
   — would forbid the very mechanism ADRs 0029–0038 established (an owner publishes a *thing* the consumer
   reads). And the Category-C membership reads are **set-based** `EXISTS` predicates inside list/search
   queries (ADR 0037, `effective_to IS NULL`); replacing them with per-row Go calls is an N+1 regression and,
   for the in-tx site, brushes the H-PRE-1 advisory-lock hazard.

This ADR makes the bar **mechanical**: a named definition + a named exemption list, so "compliant" vs
"violation" is decidable by a reviewer or a linter with **no judgement**.

## Decision

### D1 — The boundary rule (one sentence)

> A module reading **another module's owned base table** with raw SQL (`SELECT` / `JOIN` / subquery /
> `EXISTS` against the table) is an **H-G violation**. Reading the owner's **published, versioned view** or
> calling the owner's **published read-port**, or reading **one's own** tables, is **compliant**.

"Owner" = the module that holds the table's writes (`INSERT`/`UPDATE`/`DELETE`). Ownership is a fact about
where the mutations live, established in source — not where the type is declared.

### D2 — "Owned table" is refined to "owned **base** table"

The broad re-audit reading ("any cross-module owned-table read") and the canonical §6 greps are reconciled by
refining **owned table → owned base table**. A **base table** is a physical table the owner mutates. A
**published view** (or read-model / projection) is *not* a base table: it is a deliberate, versioned read
contract the owner exposes. Under this refinement, **H-G = 0 holds under both readings simultaneously** —
the strict "no cross-module base-table read" and the canonical greps agree, because the only sanctioned
cross-module SQL is against published views, which neither reading counts as a violation.

### D3 — Exemption list (the compliant mechanisms)

A cross-module read is **compliant** iff it is one of:

- **(a) Published versioned view / read-model.** A `JOIN`/`SELECT` against another module's *published* view
  (e.g. iam's `metaldocs.v_active_user_areas`, a CD visibility projection). The view is the owner's published
  contract; its name is stable and versioned; the owner may refactor base tables beneath it. This keeps
  set-based SQL (no N+1) and, being `SELECT`-only, is tx-structure-neutral (H-PRE-1 safe for free — only the
  object name changes, base table → view; the caller's tx/lock structure is untouched).
- **(b) Owner-published read-port.** A Go call through an interface the owner publishes in its `domain`
  package, implemented in the owner's `infrastructure`, wired at the composition root (the ADR
  0029/0030/0031/0038 pattern). The consumer imports the interface, never the owner's table.
- **(c) Own tables.** A module reading the tables it itself owns is never an H-G concern.
- **(d) Platform append-sink read.** A read projection of a **cross-cutting platform sink** — a table the
  owning module exposes as an append-only telemetry/audit surface that *every* module writes via a published
  append API (`audit_events`, written through `AppendAudit[Tx]`). The sink is architecturally a shared
  platform surface, not a domain module's private state; cross-module *read* projections of it (metrics,
  audit-history lists) are a distinct class from reading a domain module's base table. Exempt by **recorded
  allowlist** (the F0.3 guard enumerates the sites), not by blanket rule — a new sink read must be added to
  the allowlist with justification.
- **(e) Parent-ADR-dispositioned auth read.** A cross-module read of `auth_identities` / `auth_sessions` that
  was **audited and accepted by the parent `grade-a-completion` M4** under **ADR 0029/0031** — specifically
  ADR 0031, which sanctions security/iam scoping `auth_identities` via `= ANY(ids)` precisely because that
  table carries no `tenant_id`. These are *dispositioned* (a closed parent decision), not unported; re-porting
  them re-litigates ADR 0031. Exempt by recorded allowlist; the guard carries the ADR reference per site.
- **(f) Worker-layer read.** A read by the `internal/modules/jobs/**` worker layer (e.g. the stuck-instance
  watchdog reading `approval_*` tables). The jobs layer is *infrastructure operating on* a domain, not a peer
  domain module; whether it is subject to the peer-module boundary rule is **deferred to a future
  jobs-boundary pass**. Exempt-with-note (recorded), not silently dropped.

Anything else — a raw `SELECT`/`JOIN`/`EXISTS` naming another module's **base table**, outside (a)–(f) — is a
**violation** and must be ported to (a) or (b) before this mission's terminal acceptance.

> **Honest scope of "H-G = 0 under both readings" (D2) once (d)–(f) exist.** Exemptions (d)–(f) are *recorded
> carve-outs for raw base-table reads*, so the strict broad reading is **not** literally 0 — those sites are
> still raw SQL. The terminal bar (§8) therefore means **0 violations outside the recorded allowlist**: every
> cross-module SQL read is either (i) against a published view / read-port / own table, or (ii) one of the
> finite, enumerated, justified (d)–(f) sites in the F0.3 guard's allowlist. The allowlist is the honest
> reconciliation — it does not pretend the reads are absent; it records *why each is principled*. A read that
> is neither (a)–(c) nor on the allowlist fails the guard.

### D4 — The active-now membership-view contract (no reinterpretation of ADR 0037)

The iam-published active-membership view (mechanism C-α, exemption (a), built in M3) encodes **exactly**
`effective_to IS NULL` — the Model-A active-now predicate fixed by **ADR 0037**. The view is a renamed,
published surface over that predicate; it introduces **no** interval reinterpretation
(`effective_to > now()` remains refuted per ADR 0037 D2). As-of/history reads keep the parameterized interval
form (ADR 0037 D3) and are out of this ADR's scope.

### D5 — H-PRE-1 (advisory-lock hazard) is preserved, not introduced

A membership/visibility read that executes **inside** a caller-supplied lock-holding transaction (the
`approval … ResolveEligibleActors` site, mission §5 row 14) must remain a plain `SELECT` (writes no
`authz_access_log` / is non-recording) and must stay tx-aware. A **published view** satisfies this for free:
the read is still `SELECT`-only; only the object name changes. **No port for any H-G site may place an
authz-recording read inside a lock-holding atomic tx** (HS-PRE-1). This ADR records the rule; M3 consumes it.

### D6 — The boundary is CI-enforced, not discipline-enforced

A cilint analyzer (`hgcrossmodule`, M0/F0.3) mechanizes D1–D3: it flags a raw cross-module base-table read in
a non-owner package outside the published-contract allowlist, and **fails the build** (`go run ./tools/cilint
./...` exits non-zero). The class cannot silently re-open. The analyzer is the sibling of the H-D
`noresponsemap` guard; like it, it matches literal table tokens (the residual — dynamically-assembled or
aliased SQL — is recorded as a census assumption, not engineered away).

## Worked classification — every mission §5 in-scope site (proof the rule is unambiguous)

Each site below is classified by **D1/D3 alone**, mechanically, with the deciding clause named. Owner per the
discovery census (`…/discovery-brief.md`). This table is the demonstration that the definition admits **no
judgement** — and the seed F0.2 reconciles its binding census against. (Category A rows read **no** SQL — they
are in-memory foreign-state literals; they are an H-G-adjacent coupling the mission remediates with typed
constants, classified here as "not a base-table read" for completeness.)

| # (mission §5) | Site | Reads (owner) | Reader module | Verdict | Deciding clause |
|---|------|---------------|---------------|---------|-----------------|
| 3 | `controlleddocuments/domain/resolution.go:42,55,58` | templates vocab (in-memory `*string`, **no SQL**) | CD | not a base-table read (Category A literal coupling) | D1 ranges over **SQL** reads; fixed via typed constants (M1), not a port |
| 4 | `documents/repository/repository.go:1701` `SELECT profile_code FROM controlled_documents` | **CD** | documents | **violation** | D1 — raw `SELECT` against another module's base table |
| 5 | `controlleddocuments/infrastructure/repository.go:532` `FROM document_revisions` | **documents** | CD | **violation** | D1 — raw subquery against another module's base table |
| 6 | `controlleddocuments/infrastructure/repository.go:539,545` `FROM documents` | **documents** | CD | **violation** | D1 — raw `JOIN` against another module's base table |
| 7 | `controlleddocuments/infrastructure/repository.go:593` `FROM approval_instances` | **approval** | CD | **violation** | D1 — raw `FROM` against another module's base table |
| 8 | `documents/application/document_area.go:37` `JOIN controlled_documents` (in-tx) | **CD** | documents | **violation** | D1 — raw `JOIN` against another module's base table; port must be tx-aware (in-tx snapshot) |
| 9 | `documents/approval/application/read_service.go:355` `JOIN controlled_documents` (in-tx) | **CD** | documents/approval | **violation** | D1 — raw `JOIN` against another module's base table; tx-aware port |
| 10 | `documents/repository/repository.go:154` `FROM metaldocs.document_process_areas` (in-tx) | **taxonomy** | documents | **violation** | D1 — raw `SELECT` against another module's base table (owner = taxonomy, census correction) |
| 11 | `iam/infrastructure/postgres/area_catalog_reader.go:28` `EXISTS(… document_process_areas)` | **taxonomy** | iam | **violation** | D1 — raw `EXISTS` against another module's base table (owner = taxonomy) |
| 12 | `controlleddocuments/infrastructure/repository.go:150` `FROM user_process_areas` EXISTS (list) | **iam** | CD | **violation** → compliant via exemption (a) | D1 violation today; D3(a) — port to the iam published active-membership view (M3) |
| 13 | `controlleddocuments/infrastructure/repository.go:492` `FROM user_process_areas` EXISTS (CanRead) | **iam** | CD | **violation** → compliant via (a) | D1 today; D3(a) — published view (M3) |
| 14 | `documents/approval/repository/postgres_approval_repository.go:1136` `FROM metaldocs.user_process_areas` (in-tx) | **iam** | approval | **violation** → compliant via (a), H-PRE-1 | D1 today; D3(a)+D5 — published view read **inside the existing tx**, non-recording (M3) |
| 15 | `search/infrastructure/v2documents/reader.go:70,97,102,111` `controlled_documents` + `controlled_document_*_grants` + `user_process_areas` | **CD + iam** | search | **violation** (worst offender — inlines CD's whole visibility predicate) | D1 — multiple raw reads of two foreign modules' base tables; D3(a) — CD-published visibility projection + iam membership view (M4) |

**Result: 0 sites unclassified.** Every SQL site is a **violation** under D1 today and becomes **compliant**
under exactly one D3 exemption after its mission-assigned port/view lands. The one non-SQL site (row 3) is
out of D1's range (literal coupling, fixed via typed constants).

## Amendment 2026-06-20 — M0/F0.2 HS-6 ruling (widen census surfaced sites beyond the §5 inventory)

The F0.2 binding census (`docs/superpowers/milestones/backend-module-boundary-hardening/milestone-0-adr-and-census/f0.2-binding-census/census.md`),
widening from the named §5 tokens to the **full owned-base-table set**, surfaced sites the original §5
inventory did not enumerate. Per mission §9 HS-6 these were surfaced to the operator
(`…/f0.2-binding-census/hs-6-scope-decision.md`); the operator ruled **fold N1 into M2** + **resolve X1–X8 via
the principled exemptions D3(d)–(f)** (this amendment adds them). Classification of the new sites:

| Site | Reads (owner) | Reader | Verdict | Deciding clause |
|------|---------------|--------|---------|-----------------|
| N1 — `documents/application/fillin_service.go:225` | `templates_template_version` (templates) | documents | **violation** → compliant via D3(b) | D1 today; port to a templates `placeholder_schema` read-port. **Folded into M2** (operator HS-6 ruling; ADR 0030 precedent). |
| X1 — `security/infrastructure/postgres/repository.go:121,185,236` | `auth_identities` (auth) | security | **exempt** D3(e) | Parent grade-a M4 F4.6; ADR 0031 `= ANY(ids)` scoping. Allowlisted. |
| X2 — `security/infrastructure/postgres/repository.go:262,269` | `auth_sessions` (auth) | security | **exempt** D3(e) | Parent M4 F4.6 session projection. Allowlisted. |
| X3 — `security/infrastructure/postgres/repository.go:340` | `audit_events` (audit) | security | **exempt** D3(d) | Platform append-sink read projection. Allowlisted. |
| X4 — `iam/infrastructure/postgres/observability_repository.go:63,75,93,152` | `audit_events` (audit) | iam | **exempt** D3(d) | Platform sink metrics COUNT. Allowlisted. |
| X5 — `iam/infrastructure/postgres/observability_repository.go:133,174` | `auth_identities` (auth) | iam | **exempt** D3(e) | active-users/lockout metrics; ADR 0029/0031 family. Allowlisted. |
| X6 — `iam/presence/repository.go:64–65` | `auth_identities` (auth) | iam | **exempt** D3(e) | presence username JOIN; auth-read disposition. Allowlisted. |
| X7 — `templates/repository/postgres.go:695` (`ListAudit`) | `audit_events` (audit) | templates | **exempt** D3(d) | Platform sink audit-history read. Allowlisted. |
| X8 — `jobs/stuck_instance_watchdog/job.go:147,148` | `approval_instances`, `approval_stage_instances` (approval) | jobs | **exempt** D3(f) | Worker-layer read; jobs-boundary deferred. Allowlisted-with-note. |

**Still 0 sites unclassified** after the widen: N1 is a violation routed to M2; X1–X8 are exempt under the
named D3(d)–(f) clauses and enumerated in the F0.3 guard allowlist. The `auth_failure_counters` candidate is a
**false positive** (documents/approval owns it — INSERT at `postgres_auth_failure_rate_limiter.go:64`, DELETE
at `:86`; same-module read), recorded and dropped.

## Consequences

### Positive
- The H-G done-bar is **mechanical**: a reviewer or the cilint guard classifies any cross-module read with no
  judgement. The undercount root cause (no machine-checkable definition) is fixed.
- "H-G = 0 under both readings" is **coherent** — the base/view refinement (D2) makes the strict reading and
  the canonical greps agree.
- The sanctioned mechanism (published view / read-port) **preserves set-based SQL** (no N+1) and is
  **H-PRE-1-safe** (SELECT-only views, tx-structure-neutral).
- Generalizes the per-table ADRs (0029/0030/0031/0038) into one durable, enforced rule.

### Negative / cost
- Each ported site adds one published view or one port round-trip (small bounded reads; see the per-ADR cost
  notes). Accepted — correctness/boundary integrity outranks a bounded extra read.
- The guard matches literal table tokens; an aliased or dynamically-assembled cross-module read could evade
  it. Recorded as a census assumption (F0.2 coverage statement), not engineered away in M0.

### Neutral
- **No behavior change, no migration, no query-result change** from *this ADR* — it is a definition. The views
  and ports it sanctions are built (with parity tests) in M1–M4.

## Verification

A cross-module read is **ADR-0039-compliant** when:
- It reads the owner's **published view** (D3a), or calls the owner's **published read-port** (D3b), or reads
  the reader's **own** tables (D3c); **and**
- it names **no** other module's **base table** in raw SQL; **and**
- if it runs inside a lock-holding tx, it is a non-recording `SELECT` (D5, HS-PRE-1).

Enforced by `go run ./tools/cilint ./...` (the `hgcrossmodule` analyzer, M0/F0.3): green ⟺ every cross-module
SQL read is against a view/own-table or carries a recorded exemption. The mission's terminal acceptance
(`mission.md` §8) requires this guard green on the full tree and the re-run re-audit grading
module-boundaries = A with H-G = 0 under both readings.
