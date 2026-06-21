# Feature F4.1 — Spec: controlleddocuments publishes a search visibility + projection read contract

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.1-cd-visibility-contract`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator ratified Shape 2 via F4.1 consumer-contract interview) — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine: `superpowers:brainstorming`, seeded with milestone.md F4.1 row. Consumer-contract-first
dialog; the contract shape (the HS-2 risk item) was the decisive question.

| # | Question | Answer |
|---|----------|--------|
| 1 | What does the search consumer actually need from CD — visibility only, or projection too? | **Both.** `reader.go` LEFT JOINs `controlled_documents` for *display projection* (`cd.code, cd.department_code, cd.profile_code, cd.sequence_num` — reader.go:60,62,64,65) AND for the *visibility predicate* (`visibility_scope`, `owner_user_id`, area-grants, user-grants). So the CD contract serves two jobs; C4b (`controlled_documents`) covers both, plus C4c/C4e (grant tables). |
| 2 | Can CD publish the visibility decision as a single `(cd × actor)` view search equi-joins on `$13`? | **No — trips HS-2.** Company-scope CDs are visible to *every* tenant user; enumerating them requires JOINing `iam_users` (an iam base table) → reintroduces an H-G violation. The unbounded legs (company, owner) must stay scalar; only the *bounded* grant edges may enumerate. |
| 3 | How to keep the projection LEFT JOIN at 1 row/CD while enumerating bounded grantees? | **Two views (Shape 2).** A 1:1 facts view is the projection JOIN target (carries projection cols + `is_company` + `owner_user_id`); a separate bounded grantee view carries one row per real grant edge, consumed by search as an `EXISTS`. A single fan-out view would multiply document rows through the projection JOIN → rejected. |
| 4 | How is the `'company'`/`'restricted'` literal coupling (reader.go:92,94, Category-A class) removed? | The facts view exposes **`is_company boolean`** (= `visibility_scope = 'company'`), so search never names CD's scope enum. The `restricted` leg is implicit: a non-company CD is visible only via owner or a grantee `EXISTS`; `restricted` need not be named at all. |
| 5 | How does the bounded grantee view stay zero-drift vs the inline predicate (esp. revoked membership)? | `v_cd_grantee` derives area-membership grantees by joining `controlled_document_area_grants` → iam's `metaldocs.v_active_user_areas` (which encodes `effective_to IS NULL`, ADR 0037 D1 — revoked rows already excluded), UNION the direct `controlled_document_user_grants`. Identical set to the inline correlated-EXISTS legs. Parity test seeds revoked-member + ungranted-user discriminators. |
| 6 | Owner of these views / which module's migration? | controlleddocuments — the views read only CD-owned base tables (`controlled_documents`, `controlled_document_area_grants`, `controlled_document_user_grants`) + iam's *published* `v_active_user_areas` (compliant D3a). No third module's base table. Migration 0243. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `internal/modules/search/infrastructure/v2documents/reader.go` `ListDocuments` (the v2
  documents search list query). It is the *only* consumer of these views in M4 (F4.3 wires it).
- **Contract — two published, versioned views (ADR-0039 D3a), schema `metaldocs`:**

  **(1) `metaldocs.v_cd_search_facts`** — exactly 1 row per controlled document. Projection JOIN target
  AND scalar visibility legs:
  | column | type | meaning |
  |--------|------|---------|
  | `tenant_id` | uuid | CD tenant |
  | `controlled_document_id` | uuid | = `controlled_documents.id` (search joins on `d.controlled_document_id`) |
  | `code` | text | `controlled_documents.code` (projection) |
  | `department_code` | text | `controlled_documents.department_code` (projection) |
  | `profile_code` | text | `controlled_documents.profile_code` (projection — search's `cd.profile_code` COALESCE source) |
  | `sequence_num` | bigint/int | `controlled_documents.sequence_num` (projection) |
  | `is_company` | boolean | `visibility_scope = 'company'` (replaces the `'company'` literal) |
  | `owner_user_id` | uuid | `controlled_documents.owner_user_id` (owner visibility leg) |

  **(2) `metaldocs.v_cd_grantee`** — bounded; 1 row per (CD, actor) grant edge for restricted CDs only:
  | column | type | meaning |
  |--------|------|---------|
  | `tenant_id` | uuid | CD tenant |
  | `controlled_document_id` | uuid | the restricted CD |
  | `grantee_user_id` | uuid | an actor who may see it via an **active** area-grant membership (`v_active_user_areas`) **or** a direct user-grant |

- **The visibility decision search computes** (equivalent to today's inline predicate), for a CD row `cd`
  from `v_cd_search_facts` and actor `$13`:
  `cd.is_company OR cd.owner_user_id = $13 OR EXISTS(SELECT 1 FROM metaldocs.v_cd_grantee g WHERE g.tenant_id = cd.tenant_id AND g.controlled_document_id = cd.controlled_document_id AND g.grantee_user_id = $13)`
- **Source of truth for the contract:** the current CD canonical predicate
  `internal/modules/controlleddocuments/infrastructure/repository.go:145-178` (post-M3, already consuming
  `v_active_user_areas`) and the inlined copy in `reader.go:89-118`. ADR-0039 D3a/D4.

## What this feature implements

A single migration (`db/migrations/0243_*.sql`) creating the two `metaldocs` views above over CD's own
base tables plus iam's published `v_active_user_areas`, with `COMMENT ON VIEW` documenting each as a
published ADR-0039 D3a read contract. ADR-0039's `Related code` note is extended to reference them. No
Go code, no change to CD's own repository, no change to search (that is F4.3).

## Non-goals (mandatory)

- **No search-side change** — `reader.go` is untouched in F4.1 (F4.3 consumes the views).
- **No documents projection** — the `public.documents` read (C4a) is F4.2's view, not here.
- **No change to CD's own List/CanRead repository predicate** — F4.1 only *publishes* views; CD's
  internal queries keep their current shape (already M3-compliant).
- **No (cd × actor) cross-product / company-scope enumeration** — rejected in interview Q2 (HS-2).
- **No new authz semantics** — seam only; the views encode exactly the existing predicate's set.
- **No grantee fan-out through the projection JOIN** — facts view is strictly 1:1.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration 0243 applies cleanly on real PG (test DB :5434) | `go test -tags integration ./internal/modules/controlleddocuments/infrastructure/...` runs against migrated schema; or migration runner applies 0243 | real |
| `v_cd_search_facts` yields exactly 1 row per CD with correct projection cols + `is_company` + `owner_user_id` == base-table values | new integration parity test `cd_visibility_contract_parity_integration_test.go` — facts-view row == `controlled_documents` row across company + restricted CDs | real |
| `v_cd_grantee` set == inline grant-leg set, **excluding revoked members and ungranted users** | same parity test — grantee set for restricted CD == {active area members ∪ user-grants}; asserts revoked-member ∉ set, ungranted-user ∉ set | real |
| The composed decision `is_company OR owner=$13 OR EXISTS(grantee=$13)` == the verbatim pre-M4 inline predicate, across all 5 actor scopes | parity test composes the decision over the views and compares to a verbatim raw copy of `reader.go:89-118` predicate for {owner, areaMember, revokedMem, userGrant, none} × {company, restricted} | real |
| ADR-0039 references the two new views | grep ADR-0039 for `v_cd_search_facts` / `v_cd_grantee` | real |

> TDD: write the failing parity test (against the not-yet-created views) first, then add the migration to
> green. The discriminator rows (revoked member, ungranted user) are the anti-drift guard.
> **HS-3:** if test PG :5434 is down, these steps are **not-run**, never false-green.

## ADR needed?

- [x] No *new* ADR — this feature is an instance of the already-Accepted **ADR-0039 D3a** mechanism
  (published view). Action: extend ADR-0039's `Related code` note to list migration 0243 and the two
  views (as M3/F3.1 did for `v_active_user_areas`). No new decision.
