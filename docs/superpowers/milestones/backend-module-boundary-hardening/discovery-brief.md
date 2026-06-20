# Discovery Brief: Backend Module-Boundary Hardening

> **Mission slug:** `backend-module-boundary-hardening`  ·  **Type:** remediation
> **Date:** 2026-06-20  ·  **Branch:** `main` (HEAD `d7d53590`)
> **Agents / models used:** 3× Explore census agents (per-module sweep) + main-session definitive grep as the skeptic/reconciliation pass.
> This is the **evidence base** the mission stands on. Every claim in `mission.md` traces to a finding here.

## Why this brief exists (and why it supersedes "the 14")

The post-M9 grade-a-completion terminal re-audit
(`wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m9.md` §6/§10) named **14** H-G
cross-module sites under a broad sweep. That count was itself a **sample, not a census** — the same
under-counting failure mode that produced the five-consecutive Contract/API misses. So this mission opened
with an **exhaustive census** (3 module-scoped agents) followed by a **definitive token grep** of every
foreign owned-table name across `internal/modules`. The census expanded the set to **~20 sites** and, more
importantly, **corrected two ownership facts** the re-audit got wrong:

1. `document_process_areas` is **taxonomy-owned** (`taxonomy/infrastructure/repository.go` holds all
   INSERT/UPDATE/SELECT), **not** iam-owned. Both iam (`area_catalog_reader.go:28`) and documents
   (`repository.go:154`) read **taxonomy** cross-module.
2. `search/.../reader.go` crosses **two** foreign modules in a single visibility query (CD-owned grant
   tables **and** iam-owned `user_process_areas`) and re-derives CD's **entire** visibility predicate — the
   single worst offender, and it has **sibling** EXISTS legs (`:97`, `:111`) the sample missed.

## Method

| Agent / lens | Scope swept | Verified how |
|--------------|-------------|--------------|
| Census A — controlleddocuments | `internal/modules/controlleddocuments/**` | Read each SQL string; classified self-owned vs foreign |
| Census B — documents + approval | `internal/modules/documents/**` (incl. `approval/**`) | Read each SQL string; flagged in-tx vs off-tx |
| Census C — search / iam / security / taxonomy | `internal/modules/{search,iam,security,taxonomy}/**` | Read each SQL string; identified table owner |
| Reconciliation grep (main session) | `grep -rn` of `user_process_areas`, `controlled_document*`, `document_process_areas`, `FROM documents`/`approval_instances` across all of `internal/modules`, excluding `_test.go` | Definitive token sweep — caught siblings (`reader.go:111`) and corrected ownership (`document_process_areas` → taxonomy) |

**Verified vs assumed.** Every `file:line` below was **read** in source this session (the SQL string and its
tx/lock context inspected) — not inferred from the re-audit. The **owner** of each table was verified by
locating the module that holds its writes (INSERT/UPDATE). What is **assumed** (and deferred to the
mission's M0 binding census against ADR-0039): that no cross-module read hides behind a non-obvious table
alias or a dynamically-built query string — the grep matched literal table names only.

**Skeptic outcome.** The reconciliation grep is the skeptic pass. It *added* sites (`reader.go:111`,
confirmed `:97`), *moved* sites (`document_process_areas` owner), and *kept* every site the census named.
No census finding was dropped as hallucinated. The re-audit's "14" is **downgraded** as an undercount and is
**not** the work inventory; this brief's inventory is.

## Findings

Ownership legend: **CD** = controlleddocuments · **DOC** = documents · **APP** = documents/approval ·
**IAM** = iam · **TAX** = taxonomy · **SR** = search. "in-tx" = read executes on a caller-supplied
`*sql.Tx`/`db.Tx` (consistency- and H-PRE-1-relevant); "off-tx" = read on the pooled `r.db`.

### Category A — hardcoded foreign domain-state literals (typed-constant fix, no port)

| # | Finding (citation) | Reads owner | Severity / kind | Confidence | Proposed home |
|---|--------------------|-------------|-----------------|------------|---------------|
| A1 | `controlleddocuments/domain/resolution.go:42` — `*candidate.Status != "published"` | templates vocab (no SQL; in-memory) | minor / stringly-typed | verified | M1 |
| A2 | `controlleddocuments/domain/resolution.go:55` — `*candidate.Status == "obsolete"` | templates vocab | minor / stringly-typed | verified | M1 |
| A3 | `controlleddocuments/domain/resolution.go:58` — `*candidate.Status != "published"` | templates vocab | minor / stringly-typed | verified | M1 |

Fix: import `templates/domain` `VersionStatusPublished` / `VersionStatusObsolete` constants. Precedent:
`controlleddocuments/application/service.go:283` already uses `string(templatesdomain.VersionStatusPublished)`.
These three are **pure** Category A — no SQL, just in-memory comparison of a `*string`.

> **Literal-coupling note:** other hardcoded foreign-state literals exist **inside** the cross-module SQL of
> Category B/C sites and travel with those ports, not with A: CD `repository.go:542` (`'draft','under_review',
> 'approved','rejected','scheduled'`), `:548` (`'published'`), `:596` (`'in_progress'`); search
> `reader.go:92,94` (`'company'`/`'restricted'`). They are fixed **when the surrounding read becomes a port**,
> not as standalone constant swaps. Recorded here so the inventory is complete; homed under B/C below.

### Category B — clean foreign point-reads → owner-published read-ports

No authz-visibility logic; simple lookups/joins of another module's base table. Mechanism: owner publishes a
read-port; caller consumes it; raw SQL deleted behind a parity test.

| # | Finding (citation) | Reads owner | tx | Confidence | Proposed home |
|---|--------------------|-------------|----|------------|---------------|
| B1 | `documents/repository/repository.go:1701` — `SELECT profile_code FROM controlled_documents WHERE id=$1 AND tenant_id=$2` | CD | off-tx | verified | M2 |
| B2 | `controlleddocuments/infrastructure/repository.go:532` — `SELECT content_hash FROM document_revisions …` (subquery in GetActiveInstance) | DOC | off-tx | verified | M2 |
| B3 | `controlleddocuments/infrastructure/repository.go:539,545` — `FROM documents …` active/published instance (FULL OUTER JOIN); literals `:542,:548` | DOC | off-tx | verified | M2 |
| B4 | `controlleddocuments/infrastructure/repository.go:593` — `FROM approval_instances …`; literal `:596 'in_progress'` | APP | off-tx | verified | M2 |
| B5 | `documents/application/document_area.go:37` — `LEFT JOIN controlled_documents cd …` (area-code snapshot) **[NEW — not in the 14]** | CD | **in-tx** | verified | M2 |
| B6 | `documents/approval/application/read_service.go:355` — `LEFT JOIN controlled_documents cd …` (instance area-code) **[NEW — not in the 14]** | CD | **in-tx** | verified | M2 |
| B7 | `documents/repository/repository.go:154` — `SELECT name FROM metaldocs.document_process_areas …` (area-name snapshot) | **TAX** | **in-tx** | verified | M2 |
| B8 | `iam/infrastructure/postgres/area_catalog_reader.go:28` — `EXISTS(… FROM metaldocs.document_process_areas)` (area existence) | **TAX** | off-tx | verified | M2 |

> B5/B6 are **in-tx** reads on a `*sql.Tx` and produce a value the surrounding write transaction snapshots
> (area-code). The port for these must be **tx-aware** (accept the caller's tx) to preserve snapshot
> consistency — it cannot become an off-tx call without changing semantics. B7 likewise in-tx.

### Category C — authz-visibility membership reads (design-first; tx-aware port; H-PRE-1)

These re-derive another module's **visibility predicate** and read iam's `user_process_areas` active-now
membership (`effective_to IS NULL`, ADR 0037). Set-based; naive Go-side replacement risks N+1. Mechanism is
the **load-bearing Phase-2 decision** (see Open Questions).

| # | Finding (citation) | Foreign owners crossed | tx | Confidence | Proposed home |
|---|--------------------|------------------------|----|------------|---------------|
| C1 | `controlleddocuments/infrastructure/repository.go:150` — `FROM user_process_areas upa` EXISTS, in **ListControlledDocuments** visibility subquery | IAM (CD owns the grant tables + predicate; only the membership leg is foreign) | off-tx | verified | M3 |
| C2 | `controlleddocuments/infrastructure/repository.go:492` — `FROM user_process_areas upa` EXISTS, in **CanRead** visibility subquery | IAM | off-tx | verified | M3 |
| C3 | `documents/approval/repository/postgres_approval_repository.go:1136` — `FROM metaldocs.user_process_areas` in `ResolveEligibleActors(ctx, tx, …)` | IAM | **in-tx (H-PRE-1)** | verified | M3 |
| C4 | `search/infrastructure/v2documents/reader.go:70,97,102,111` — single visibility query: `LEFT JOIN public.controlled_documents` (:70), `controlled_document_area_grants` EXISTS (:97), `user_process_areas` EXISTS (:102), `controlled_document_user_grants` EXISTS (:111); literals `:92,94 'company'/'restricted'` | **CD + IAM** (re-derives CD's *entire* visibility predicate) | off-tx | verified | M3 |

> **C4 is the worst offender.** Search does not just borrow iam's membership leg — it **inlines CD's whole
> visibility rule** (company/restricted/owner/area-grant/user-grant). Swapping only the `user_process_areas`
> leg leaves search still re-deriving CD's predicate. The clean fix is for search to consume a **CD-published
> visibility/ACL contract** (so the rule lives in one place), with iam membership underneath it. This makes
> C4 materially larger than C1–C3 and may warrant its own feature.

## Constraints & risks surfaced

- **H-PRE-1 (advisory-lock deadlock memory):** C3 reads iam membership **inside** a `tx`. The rule forbids an
  authz-**recording** read inside a lock-holding atomic tx. A plain membership `SELECT` is *not* recording
  (writes no `authz_access_log`), so C3 is not a deadlock today — but any C-mechanism port for C3 **must stay
  tx-aware and must not introduce a recording call**. A **published-view** mechanism satisfies this for free:
  only the object name changes (base table → view), tx/lock structure is untouched.
- **No N+1 regression:** C1/C2/C4 are set-based EXISTS predicates inside list/search queries. A Go-side
  per-row membership loop would be a perf regression. The mechanism must preserve set-based evaluation.
- **ADR 0037 active-now predicate:** every membership read gates on `upa.effective_to IS NULL`. Any published
  view/port must encode exactly this predicate — no interval reinterpretation.
- **Parity-before-delete:** each ported site needs a parity test proving identical results before the raw SQL
  is removed (mission constraint).
- **HS-2 redesign boundary:** C4 (search consuming a CD visibility contract) touches the shared
  visibility/ACL surface across CD + search + iam. If the mechanism turns out to require a cross-module API
  redesign rather than a contained port, that is an HS-2 stop — surface before building.
- **Ownership corrections are load-bearing:** `document_process_areas` = **taxonomy**, not iam (B7/B8). The
  ADR exemption list and M0 census must use the corrected owner, or B7/B8 get mis-homed.

## Open questions for the operator (Phase-2 Locked Decisions)

1. **D1 — Category C mechanism (load-bearing).** Pick the H-G-compliant seam, recorded in ADR-0039:
   - **Option C-α (published view + exemption):** iam publishes `metaldocs.v_active_user_areas` (and CD
     publishes a visibility projection for C4); the ADR declares **JOIN/read of another module's published,
     versioned view = compliant; raw read of a base table = violation**. Preserves set-based SQL, sidesteps
     H-PRE-1, cheapest. "H-G=0 under both readings" holds because the broad reading's "owned table" is
     refined to "owned **base** table."
   - **Option C-β (tx-aware Go ports, no views):** iam exposes a `MembershipPort.HasActiveArea(ctx, tx, …)` /
     `ActiveAreaCodes(…)`; callers pre-resolve membership then filter. Strict "no cross-module SQL at all,"
     but risks N+1 / requires query restructuring, and C4 still needs a CD visibility port.
   - *Recommendation:* **C-α** — it is the only option that makes both §8 readings agree without a perf
     regression and without touching the H-PRE-1 lock structure. C-β is the strict-purist fallback.
2. **D2 — Program placement / Grade-A gating.** Confirm this mission is a **sibling that gates** the parent
   grade-a-completion's Grade-A sign-off (parent terminal acceptance is met-on-bar but Grade-A is HELD).
3. **D3 — C4 scope.** Is "search consumes a CD-published visibility contract" **in** this mission, or a
   bounded follow-up? It is the largest single item and the one most likely to hit HS-2.
4. **D4 — Exemption list contents.** Ratify the ADR-0039 named exemptions (published view JOIN = compliant;
   taxonomy/iam published read-ports = compliant; what else?).

## Coverage statement

**Swept:** all of `internal/modules/**` non-test `.go` for the foreign owned-table tokens `user_process_areas`,
`controlled_document*`, `document_process_areas`, and `FROM documents`/`approval_instances` reads by
non-owners. Auth/security confirmed **already ported** (iam ports, no live cross-module SQL) — M4 work holds.

**NOT swept (deferred to mission M0 binding census):**
- Dynamically-assembled SQL or table names behind aliases/variables — grep matched literal tokens only.
- Cross-module reads via **other** owned tables not in the token list above (e.g. template/version tables,
  governance/event tables) — the audit scope was the H-G classes named in the re-audit; M0 must widen to the
  full owned-table set once ADR-0039 fixes the definition.
- `_test.go` files — excluded by design; tests may legitimately read across modules.
- Integration verification (docker Postgres :5433) was **not run** this session (Docker down); all findings
  are static-read, not runtime-reproduced.

No silent caps: the inventory above is **~20 sites in 3 categories**, expanded and ownership-corrected from
the re-audit's 14. The mission's M0 re-runs this census against the ADR-0039 definition as its first gate.
