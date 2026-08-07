# AuthZ grant-unification — independent advisory opinion (2026-08-07)

**Advisor:** Fable 5 (high effort), commissioned by the operator as specialist reviewer.
**Standing:** advisory. Not a ratification and not a gate. Recorded verbatim in substance so the
design spec can dispose of each point on the record rather than absorbing it silently.

> **Round 2 (2026-08-07).** The spec written from round 1 was audited back against the operator's
> standard — *does this actually delete the manual and the legacy, or does it just read well?* Verdict
> **(b): a large real de-legacying, but not yet mechanically self-defending.** Nine residues, two
> guards claiming more than they enforce, one module-boundary violation, four missing sections. All
> closed in spec **v2**; findings summarized in §R2 at the end of this document. Two of them generalized
> into register entries ME-11 (#84) and ME-12 (#85).

**Scope reviewed:** `2026-08-06-authz-grant-unification-decisions.md` (D1–D4),
`2026-08-06-authz-grant-unification-system-impact.md` (the gate artifact),
`docs/engineering/mechanical-enforcement-register.md`, and the ratified design decisions DD-1..DD-4.

**Verification the advisor did first-hand** (stated, and consistent with what this session verified
independently): all four `CapabilityService.CanDo` query arms plus `IsSystemAdmin` and
`CapsByUserID`; `authz.Require`'s bypass short-circuit, `recordBypass`, and the asserted-caps GUC
plumbing; the documents-handler visibility binary; `area_capability_reader.go`;
`user_area_repository.go`; the tripwire function in the baseline; the auth service (4,378 LOC
confirmed, `Authenticate` has no MFA step); the `MfaUserReader` port.

---

## Severity-ordered summary

1. **DD-2's in-migration equivalence assertion is largely ceremonial and can pass vacuously.** The
   real evidence must be a seeded Go-level parity test. (High)
2. **DD-3 contradicts D3's own ratified reasoning** — the authorship floor is a second grant regime,
   the exact structure D3 rejects — and it is keyed on document identity while content is not frozen
   at authorship. Restructure, don't defend. (High)
3. **DD-1's claim "the ordinary path already records it" is factually wrong today.** Deleting
   `recordBypass` deletes audit fidelity nothing replaces. (High)
4. **DD-1's bypass inventory undercounts: five, not three.** The kill list is itself a hand-synced
   enumeration. (Medium, cheap)
5. **DD-4's theorem covers the two tiers only.** The read models can re-implement the predicate and
   drift; needs a sole-reader lint. (Medium)
6. **ME-08: delete the coverage endpoint now.** The defect class is real and generalizable. (High
   standalone; the fix is a day)
7. **ME-07: recommend adopt, timed post-v1; gate the MFA build on that decision.** (Strategic)
8. **Undecided: is the role catalog product reference data or tenant data?** "Adding a role is an
   INSERT" and contract-first generated role enums cannot both be fully true. (Medium — operator
   decision, blocks design)

---

## DD-1 — kill all three bypasses

Direction endorsed. Kubernetes' `cluster-admin` is a `ClusterRole` bound like any other, not an
`if isAdmin` branch; Zanzibar has no admin bit anywhere. Re-expressing document visibility as "holds
`document.read` at `scope_kind='tenant'`" is exactly the ME-03 fix. Three problems.

### P1 — the audit claim is wrong as written

D2 says "make it a grant and the ordinary path already records it." The code says otherwise. The
ordinary grant path in `authz.Require` (`internal/modules/iam/authz/authz.go:144-166`) records
**nothing** — it sets the asserted-caps GUC and returns. Only the bypass arm INSERTs an audit event
(`authz.go:125-133`, fail-closed in-tx, per ADR 0022 Phase 11 F8, whose rationale was that a
tenant-wide bypass must be visible in `audit.read` at the same fidelity as a normal grant).

Delete `recordBypass` and `system_admin`'s tenant-wide privilege use goes from per-use audited to
not audited at all.

The honest defense exists: after unification the *grant itself* is durable, provenance-carrying data
in `capability_bindings`, so "why was this actor authorized" is reconstructable from
(operation audit event) ⋈ (binding history) — which the old model could not do. That defense must be
argued in the mandatory ADR as an **explicit disposition of F8, with the reconstruction query
shown**. Silently dropping the INSERT is how a Part 11 finding is born.

Corollary: dropping the INSERT is also what makes `authz.Require` legal in a read-only tx. The gate
artifact §6 is right that relaxing the `authz-require-rw-tx` lint is a separate decision — make it in
the same ADR, not by drift.

### P2 — the "full capability set" bundle is a new hand-sync

If `system_admin`'s bundle is seeded once as N `role_capabilities` rows, capability N+1 added to the
Go registry never appears in it and `system_admin` silently lacks it — the disjoint-tables defect
reborn as a stale bundle. Gate artifact §4.5 gestures at "derived from the registry" but names no
firing mechanism.

Required: either the seed is generator output with a blocking parity lint (level 3), or a boot check
asserts `bundle(system_admin) == registry` (level 2). GMR M2's tripwire generator + parity lints is
the in-house template. Without one, D2 trades a visible bypass for an invisible drift.

### P3 — the inventory is wrong, and that matters more than the count

Five bypasses in production code, not three:

| # | Site | Named in DD-1? |
|---|---|---|
| a | tier-1 `CanDo` UNION arms (`capability_service.go:56-68`) | yes |
| b | tier-2 `Require` short-circuit (`authz.go:113-136`) | yes |
| c | documents visibility (`handler.go:436`) + `area_capability_reader.go:51` | yes |
| d | `UserAreaRepository.MembershipDirectoryScope` (`user_area_repository.go:159`) embeds `SystemAdminExistsSQL` as its tenant-wide branch | **no** |
| e | `CapsByUserID` (`capability_service.go:132-139`) returns `AllCapabilities()` on the admin branch — the `/auth/me` surface | **no** |

(e) matters beyond the count: leave it and the frontend renders capabilities for admins by a
different rule than enforcement uses.

The meta-point is the register's own doctrine: **a hand-listed kill list is a hand-synced
enumeration.** The completion criterion must be mechanical — a lint asserting the literal
`'system_admin'` appears in no production SQL or Go outside the role-catalog seed. The SOLE-RLS lint
is the in-house template for exactly this shape. That converts "we think we got them all" into a red
build.

---

## DD-2 — single-tx hard break

Correct call. With no production tenant, expand/contract is pure ceremony and a dual-write window is
two sources of truth — the defect as a rollout procedure. Mechanically sound: Postgres DDL is
transactional, `now()` is tx-stable so effective-interval evaluation cannot skew mid-migration, and
lock contention is irrelevant with no traffic.

**But the assertion is weaker than the prose believes.** Three ways it passes while the model is
still wrong:

1. **Vacuous truth.** The dev DB was wiped 2026-07-29. An equivalence replay over empty or
   near-empty grant tables proves nothing — old-empty equals new-empty for *any* backfill, including
   a completely broken one. The real evidence is a **Go integration test** seeding a corpus that
   deliberately covers every path: direct role, group-derived role, `iam_group_roles` junk strings
   (no CHECK — rows may reference roles that exist nowhere), expired `effective_to` rows, area grants
   for each of the 7 area roles, the 3 roles the `iam_user_roles` CHECK omits, `system_admin` via
   both direct and group — then diffing old-evaluator against new-evaluator output for every
   (actor, cap, area). Keep the in-migration assert as a belt; the seeded parity test is the
   acceptance evidence.
2. **Transcription risk.** The old evaluators are Go-embedded SQL. An in-migration assert must
   re-express them in migration SQL — a hand-copy. Drop one predicate in transcription
   (`g.tenant_id = $2` on the group join, `effective_to IS NULL`, the `IsValidCapability` pre-filter)
   and the assert compares the new model against a *wrong rendering* of the old one, and passes. The
   Go-level test dodges this entirely because it calls the actual old functions.
3. **Quantifier direction.** "The new source reproduces every grant the old evaluators returned" is
   old ⊆ new, and silent about new ∖ old — while the design *deliberately* creates new grants
   (`iam_user_roles` rows become `scope_kind='tenant'`, newly honored by tier-2; §7 names this as a
   privilege change). A one-directional assert lets unintended escalation ride in alongside the
   intended one. The assert must be three-part: **tier-1 exact equality; tier-2 old ⊆ new; and
   new ∖ old at tier-2 equals exactly the computed expected-escalation set**, enumerated and
   materialized. If the expected-diff set is not compared, the assertion is decoration.

**Migration hazard not in the artifact:** `iam_group_roles` junk rows (no CHECK) will violate the new
FK to `role_catalog`. They grant nothing today — they join to nothing in `role_capabilities` — so
skipping them is semantically safe, but it must be an explicit `WHERE role IN (SELECT ...)` with a
raised NOTICE and count, not an accidental FK failure or a silent drop.

**Atomicity:** the `SeedWithCaps` re-point (gate artifact §10 constraint 6) must land in the same
commit as 0318 or the entire integration estate breaks against the template databases.

---

## DD-4 — one evaluator, one query builder

The theorem framing is genuinely good: `tier1(cap) := ∃area: tier2(cap, area)` makes the
strict-relaxation invariant structural. Rejecting the SQL-function alternative is correct on the
evidence — `enforce_capability_asserted` (`db/baseline/0001_current_schema.sql:304+`) reads the
`metaldocs.asserted_caps` GUC, so it is proof-of-passage, not a grant reader, and there is no
third SQL-side evaluator forcing the shared source into the DB. Three gaps:

1. **The theorem covers two callers; the table will have more readers.** `AreaCapabilityReader`
   (`area_capability_reader.go:49-64`), `MembershipDirectoryScope`, and the future document-list
   visibility query all embed the grant join today. Nothing in DD-4 stops them re-implementing the
   predicate by hand against `capability_bindings` — at which point the read models drift from the
   evaluator and ME-03 recurs as "UI shows areas the evaluator would deny." Required, level 3: the
   table name `capability_bindings` may appear only inside the one builder package; every read model
   composes the builder. Same lint shape as SOLE-RLS.
2. **Dangling-scope unreachability.** `tier1 = ∃area: tier2` reintroduces the unreachable state
   through the one door the formula does not guard: a binding whose `scope_ref` names a retired or
   deleted area. Tier-1 admits, tier-2 denies for every live area — a per-user instance of exactly
   the defect this program removes. `scope_ref` must FK to areas with explicit retire semantics
   (revoke bindings on retirement, or exclude dangling scopes from tier-1's existential). Two-line
   decision now, incident later.
3. **Keep the property test anyway.** "Theorem rather than test" holds only while the code shape
   holds. The theorem argues the test is cheap, not absent.

Non-issue, stated once so it is not re-raised: tier-1 runs on the pool, tier-2 in-tx; read-skew
between them within one request is inherent and harmless, because tier-2 is the binding decision.

---

## DD-3 — the strongest case against the permanent authorship read floor

First, deflating the scary version: terminated employees are handled by `is_active`. The floor only
matters for *active* users whose grants were revoked. With that said, the case against is strong, in
four prongs.

1. **The floor is keyed on document identity, but content is not frozen at authorship.** "Documents
   I created" means the author of revision 1 reads revision 12 — written by others, after they left
   the area, including content whose classification rose after they authored it. The traceability
   justification covers *their own record*; the implementation grants *the living document*. Those
   are different objects. This is the technically decisive attack.
2. **The traceability need is already served by better instruments.** In a Part 11 audit, "an author
   must find their history" is answered by the immutable audit trail, their signature manifestations,
   and their signed PDFs — records frozen at the moment of their involvement, which is exactly what
   an audit wants. A standing live-read ACL shows current state, not their historical act.
3. **§11.10(d) and data-integrity investigations.** When an author is removed from an area *for
   cause* — the ALCOA+ data-integrity investigation, the FDA's favorite topic this decade — the
   subject of the investigation retains read access to the records at issue, and no administrator
   action short of deactivating the whole account removes it. "Show me that removing X from area Y
   removes X's access to area Y's records" gets the answer "except the ones they wrote, permanently,
   by design." Defensible on paper; a finding generator in the room.
4. **It contradicts D3's own ratified reasoning — the internal-consistency kill shot.** D3 rejects
   direct per-user grants because "a direct grant is a second grant regime living beside the role
   regime" and "who holds `document.read`?" stops having a cheap complete answer. An authorship floor
   is precisely a per-user, per-document implicit grant living outside `capability_bindings`. That is
   D3 argument 1, verbatim.

**Advisor recommendation:** replace the permanent live-document floor with a **revision-scoped
floor** — read of the revisions the actor authored or signed — or drop the floor and serve the need
through the audit/signoff record surface.

**If the operator's ruling stands as-is**, two conditions are non-negotiable: (a) express the floor
*inside* the single evaluator as an explicit relation the policy engine unions, so one query still
answers "who can read D" — never a `WHERE created_by` bolted into the list handler; and (b) record it
in the ADR as a named, justified exception to D3, or the next program inherits an unlabelled second
regime from the program whose charter was deleting second regimes. The advisor adds: a floor an
administrator can sever with a recorded reason survives an audit; an unseverable one is the finding.

---

## ME-07 — authentication build-vs-adopt

**Recommendation: adopt, timed after v1 ships; freeze the hand-rolled auth now; let the MFA decision
force the timing.**

- **The category call is settled** — authN commodity, authZ domain. The code is good: the
  per-identity login lock closing the lockout TOCTOU and the bcrypt-time equalization on every
  failure path (`service.go:283-318`) is better than most commercial products' first attempt. Quality
  is not the question. The question is who carries the *next* 4,000 lines: TOTP, WebAuthn, recovery
  codes, SSO/SAML, SCIM, session revocation under federation. That backlog is inevitable for a
  product sold into pharma, and every line of it expands the Part 11 §11.10/§11.300 validation
  surface instead of shrinking it.
- **ME-08 is the forcing function.** MFA must exist. Building TOTP + WebAuthn in-house is the moment
  the sunk cost doubles. This converts ME-07 from "study someday" to "decide before the MFA work
  starts."
- **The integration is unusually tractable** because the module boundary is already cut correctly:
  `auth` owns credentials and sessions, `iam` owns profile and grants, joined by a stable subject id.
  Adoption means `auth` becomes an OIDC relying party, the `authn` chain link is the single seam, and
  `iam` does not move. The sync worry is real but bounded — the IdP owns credentials and factors,
  `iam_users` stays as profile + grants keyed by the OIDC `sub`. A reference, not a replica.
- **Vendor choice depends on the one thing the code cannot say — deployment model.** On-prem /
  single-tenant (common in pharma) → **Keycloak**: boring, self-hosted, enormous compliance paper
  trail, auditors have seen it a hundred times. Central pooled SaaS → **Zitadel**: multi-tenancy
  (organizations) is first-class, where Keycloak's realm-per-tenant degrades past low hundreds. Ory
  is a component kit (more integration LOC, partially defeats the purpose). WorkOS is
  SSO-as-a-SaaS-API — wrong for self-hosted regulated deployments.
- **Still needed to decide:** when the first SSO customer requirement lands; whether a formal CSV
  package is intended (vendor SOC2/validation evidence is the strongest adopt argument); ops appetite
  for one more stateful service in the compose stack; and the release state — the release is on HOLD
  with three blockers, and inserting an IdP migration into a held release would be malpractice.
- **Sequence:** v1 ships on the frozen hand-rolled path → 1-week spike (Keycloak vs Zitadel against
  the tenancy model) → adopt, with MFA as the first delivered feature of the new path.

---

## ME-08 — MFA as a dashboard

**Severity confirmed, including "most serious."** The nuance: the *absence* of MFA is a product gap,
arguably not even non-compliance under a strict §11.300 reading. The *dashboard* is the dangerous
half. A coverage percentage over a mechanism that does not exist manufactures evidence — that screen
will be cited as proof MFA exists ("coverage 0%, rollout in progress"), and when the auditor asks to
see enrollment, the finding is no longer "missing feature" but **"misrepresented control"**, which in
GxP vocabulary sits next to data-integrity findings.

**Immediate action — this week, not this program:**

1. Delete the coverage endpoint and its route, with the OpenAPI operation removed (contract-first
   cuts both ways). Not feature-flagged off; deleted.
2. Delete `MfaUserReader` and its implementation. A port whose only consumer is gone is dead weight
   that reads as capability — ME-04's second gap, same class.
3. `iam_users.mfa_enabled` / `mfa_enrolled_at`: columns nothing reads are inert, but under
   legacy-extermination doctrine, drop them in the next migration and re-add when MFA is real.
4. Route the control itself to the ME-07 decision — MFA should be the IdP's first delivered feature,
   not an in-house build.

**Defect class: genuine, generalizable, add it.** Distinct from Class 12 (debt written up as policy)
by mechanism: no human ever claimed the control worked — *the artifact makes the claim structurally*.
It generalizes well past MFA: metrics for a retry path that never executes, an audit dashboard over a
disabled trail, SLO panels over unmeasured SLIs, a backup-status page with no restore path.

Suggested name: **phantom-control instrumentation**. The generalized firing mechanism is the
interesting part and belongs in the catalog entry: *a reporting surface must declare its enforcement
anchor* — the coverage handler must reference (import, or register against) the code path that
enforces the control, so a report over nothing is uncompilable or fails a lint. That is a
level-1/level-3 mechanism for the whole class, not just this instance.

---

## Cross-cutting risks

1. **Undecided contract shape for the role catalog.** D3 argument 5 justifies rejecting direct grants
   by promising cheap roles ("a new role is an INSERT, not a migration"). Gate artifact §7 says the
   OpenAPI `UserRole`/`AreaRole` enums become derived from the role catalog — i.e. roles remain
   compile-time contract surface: catalog INSERT → spec regen → FE regen → deploy. **Both cannot be
   fully true.** Either (a) the catalog is **product-global reference data**, generated enums are
   correct, and "cheap roles" is overstated so D3's escape valve is weaker than ratified; or (b) it is
   **tenant-definable data** (which "make the right path cheap" quietly implies, and which serious
   eQMS competitors offer), in which case role codes must leave the enum space entirely — strings
   validated against the catalog — and the contract changes shape. Deciding this by default, inside
   the migration, is exactly the "semantic decision made by accident" the gate artifact warns about.
   **Operator decision, before design.**
2. **Binding revocation must be soft.** `capability_bindings` should follow the ADR 0037 soft-delete
   model — revocation sets `effective_to`, never deletes — so grant history is durable Part 11
   evidence and "who could do X on date D" is answerable retroactively. Note this also serves the
   DD-3 justification: "this person was the author and held access at the time" lives in binding
   history + audit trail, further weakening the case for a live read floor.
3. **The completion criterion must include the mechanical bypass-extinction lint** (DD-1 P3).
   Otherwise "all bypasses die" is verified by the same hand inspection that undercounted them at
   three.
4. **RF-3 (cache invalidation): close it inside this program.** After unification exactly one table's
   mutation invalidates capability caches. `WithCapCache` is per-request so live risk is small, but
   the contract is worth writing while there is one table to watch instead of three.

## Where the design is right

Recorded so the corrections above are not mistaken for a rejection. The Kubernetes-shaped model
(scope on the binding) is the correct industry anchor, and correctly not-Zanzibar — ME-06's recorded
trigger is the right way to hold that door. The red CI lane as acceptance criterion is unusual but
sound for a team of one with the program queued next. Leaving `role_capabilities` untouched is
correct. The six-surfaces correction and ME-02 — "a guard covering the wrong noun reads exactly like
one that works" — should graduate into the defect-class catalog alongside the phantom-control class.

---

## Disposition

| Point | Status |
|---|---|
| DD-1 P1 — F8 audit disposition in the ADR | **accepted**, carried into the spec |
| DD-1 P2 — registry-parity guard for the system_admin bundle | **accepted** |
| DD-1 P3 — mechanically derived kill list + extinction lint | **accepted**; filed as ME-09 |
| DD-2 — seeded Go parity test as acceptance evidence, three-part directional assert | **accepted** |
| DD-2 — `iam_group_roles` junk rows explicit, NOTICE + count | **accepted** |
| DD-4 — sole-reader lint on `capability_bindings` | **accepted** |
| DD-4 — dangling-scope FK + retire semantics | **accepted** |
| DD-3 — send back for restructure (revision-scoped floor) | **accepted 2026-08-07** — see DD-5 below |
| Role catalog: reference data vs tenant data | **decided 2026-08-07** — see DD-6 below |
| ME-08 immediate deletion of the coverage endpoint | **operator decision — open** (issue #81) |
| ME-07 adopt-post-v1 recommendation + vendor split | **recorded**, no decision now (issue #80) |
| Phantom-control instrumentation defect class | **accepted** for the next catalog pass |

---

## Operator rulings on the two open points (2026-08-07)

### DD-5 — the authorship floor is revision-scoped, not document-scoped

**Supersedes DD-3 as ratified 2026-08-06.** The floor is *read of the revisions the actor authored
or signed*, never the living document.

The advisor's prong 1 is the reason: keyed on document identity, the floor gave the author of
revision 1 read access to revision 12 — written by others, after they left the area, at a
classification they never saw. The eQMS traceability need is about *their own act*, and a
revision-scoped floor is that act, frozen. A live-document floor was a larger grant wearing the
justification of a smaller one.

Two consequences, both good:
- **The D3 contradiction dissolves.** A revision the actor authored is not a per-user ACL row; it is
  a fact already recorded in revision provenance. "Who can read D?" stays answerable from
  `capability_bindings` plus a join the evaluator owns — no second grant regime, so **no named ADR
  exception to D3 is needed**. DD-3's fallback conditions (a)/(b)/(c) are moot.
- **It composes with soft revocation** (cross-cutting risk 2). Binding history answers "could they
  read it at the time"; the revision floor answers "can they retrieve what they wrote". Different
  questions, different instruments, neither faked by the other.

Design constraint carried forward: the floor is still expressed **inside the single evaluator** as a
relation the policy engine unions. Never a `WHERE created_by` in a list handler — that is how ME-03
was born.

### DD-6 — the role catalog is product reference data, with two corrections

Roles stay a product-controlled vocabulary. The OpenAPI `UserRole` / `AreaRole` enums remain
generated, and `role_catalog` becomes the single upstream they are generated *from* — which is
ME-01's fix and changes nothing about the contract's shape.

**Decided on merit, not on what exists.** Two grounds:

1. **Validated-configuration burden.** In a validated eQMS, every tenant-defined role is
   un-validated configuration inside the customer's CSV package — the customer must justify their
   own bundle in their own audit, alone. A product-controlled vocabulary is a compliance asset.
2. **The frontend guard is real and would be destroyed.** `frontend/apps/web/src/lib/iam/role-vocabulary.ts`
   single-sources both role types from the generated contract and proves every runtime list against
   them with `as const satisfies Record<UserRole, …>` — miss a role or invent one and `tsc` fails.
   18 files depend on it. Under tenant-definable roles the type degrades to `string` and that proof
   evaluates to nothing. This is a **positive template** in the register's sense and belongs beside
   `assertSurface`.

**Correction 1 — D3 argument 5 is overstated and must be reworded.** "A new role is an INSERT, not a
migration" is half true: no DDL, but a spec regen, an FE regen and a **release**. D3's case against
direct per-user grants stands on auditability and reviewability, which are untouched — but it may not
lean on a cheapness it does not deliver. The escape valve for a genuine one-off is narrower than
ratified, and the record must say so.

**Correction 2 — the inversion trigger is recorded** in the register as ME-10, in the ME-06 mold, so
this is revisited on evidence rather than by inertia. The trigger is the first customer requiring
their own role vocabulary; at that point the industry shape (one type, `role_code` as a string FK,
product roles merely seeded rows — Kubernetes `cluster-admin`, Keycloak realm roles, AWS managed
policies) becomes the global maximum and the migration cost includes a role-composition admin UI and
labels moving from code to data.

---

## §R2 — Second-pass audit of the design spec (2026-08-07)

**Question put to the advisor, in the operator's words:** *"identificar se isso estamos realmente
refatorando tirando tudo de manual, legacy, validando para não acontecer novamente, deixando nível
profissional para não ter problemas novamente, arquitetura clean, sem redundância, sem hardcode,
gambiarra e etc."*

**Verdict: (b).** The spine is professional-grade and round 1's corrections were absorbed for real,
not cosmetically. But as written, the post-program system still needed discipline in nine places, two
of seven guards claimed more than they enforced, one section broke a module boundary, and four
sections a professional spec must contain were absent.

### Findings, and where each is closed in spec v2

| # | Finding | Severity | Closed in |
|---|---|---|---|
| R2-1 | **F5's "unrepresentable" was false.** Areas are soft-archived (ADR 0010, `taxonomy/domain/area.go:120`), so the FK it relied on fires *never*; the whole claim rested on a rule a human must remember in the archive service. Level 5 in level 1's clothing — ME-02 inside the anti-ME-02 program. | High | §3.4 — the invariant moves into the evaluator: tier-1's existential joins live areas, so a dangling binding grants nothing at either tier by construction |
| R2-2 | **F3 proves the string is gone, not the bypass.** F1 generates role constants, so generated code contains the literal and the lint must allowlist it; `if role == iamtypes.RoleSystemAdmin` is then a new bypass with zero literals and a green lint. v1 §9.6 — the section that exists to state non-claims — committed the error it warns about. | High | §8 F3 non-claim column + §14 item 2; generalized as ME-12 / #85 |
| R2-3 | **§5 violated the module boundary.** Revision provenance is `documents`, signoffs are `approval` (ADR 0082). Putting the visibility union "in the evaluator" put iam's SQL across two other modules' tables. v1 conflated *not in the handler* (right) with *in iam* (wrong module). | High | §5 rewritten — visibility is `documents`' question; iam publishes scope resolution only; the signed-revision leg needs an `approval` port or is already denormalized, and determining which is implementation step 1 |
| R2-4 | **`recordBypass` deletion was overbroad.** `BypassKindBackground` is live (`authz.go:213`); `BypassSystem` records every scheduler bypass through the same function. Wholesale deletion unaudits the path nobody watches — an F8-class regression. | High | §6.1 — delete the arm, keep the function; the rw-tx relaxation narrows to `Require` call sites only |
| R2-5 | **Tripwire re-point absent.** Baseline arms guard `iam_user_roles` (`:355`), `user_process_areas` (`:358`), `iam_group_members` (`:414`). The migration drops two of those tables and v1 left the arms orphaned and `capability_bindings` unarmed — grant mutations *less* DB-guarded after a hardening program. | High | §6.3, incl. the scope-discriminated capability ruling (tenant ⇒ `user.manage`, area ⇒ `membership.manage`) |
| R2-6 | **`iam_group_members` never dispositioned.** The backfill reads it, the drop list omits it, no section states its fate — while it becomes half of every group-derived decision. | High | §3.5 |
| R2-7 | **`role_capabilities` "unchanged" was wrong.** Hundreds of hand-typed INSERTs; `.role` has no FK to the catalog; `.capability` has no parity against the Go registry — a typo'd capability is a dead row that reads as a grant. | Medium | §3.3 + F8 |
| R2-8 | **D2 silently disarms `TestEveryCapSeededOrDeferred`.** With system_admin holding every capability, `seededCaps` contains everything and the guard goes green forever — deleting one of the few mechanisms already catching this defect class. *(Found independently by the coordinator, same session.)* | High | §3.3.1 + F9 |
| R2-9 | **HTTP contract surface absent** in a contract-first repo, where the spec *is* route truth. Frontend re-point (18 files on the role enums, two membership screens) likewise unspecified. | High | §9 |
| R2-10 | **`subject_id text` chose polymorphism over referential integrity** — a binding to a deleted or non-existent subject was representable and guard-invisible. The one place the *new* schema picked convenience over the register's own doctrine. | Medium | §3.2 — discriminated column pair with real FKs |
| R2-11 | **The transaction seam was pseudocode** where it should be specified; tier-1 staying a pooled read was never stated. | Medium | §4.2 |
| R2-12 | **Expected-escalation set was "enumerated"** — if hand-listed, a hand-sync against the backfill. Parity corpus likewise hand-maintained against the role list. | Medium | §7.1 — computed in the migration; corpus derived from `role_catalog` |
| R2-13 | **No rollback story, and `scope_kind='tenant'` is where old sloppiness hides** — the backfill converts every legacy row to tenant scope, now honored by tier-2. Visible via the assert, but reviewed by nobody. | Medium | §7 archive schema, §7.2 one-time review, §11 item 8 |
| R2-14 | **F1's physical upstream unnamed.** A DB table cannot be a codegen input; left unstated, F1 degrades to a generator with a hand-maintained input nobody declared. | Medium | §3.1 — `db/reference-data/role_catalog.yaml` is the artifact |
| R2-15 | **`deferredCaps` mirrored by hand** between `permissions_test.go:609` and `scripts/api-lint/registry_rules.go`, by comment. A hand-synced enumeration inside an anti-drift mechanism. *(Found independently by the coordinator.)* | Low | Out of scope, §12; filed as ME-11 / #84 |

### Verified for this audit

`taxonomy/domain/area.go:24,115,120` (soft archive); `authz.go:126,197,213` (both bypass kinds);
`0001_current_schema.sql:355,358,414` (tripwire arms), `:141,877,952` (the three `user_process_areas`
protections needing re-homing); `permissions_test.go:504,597,609` (the seed guard and its allow-list);
`0001_product_reference_data.sql:156+` (the hand-written matrix); `role-vocabulary.ts:34-36` (the
subset proof).

### What the audit said not to touch

The hard-break single-tx migration and its three-part directional assert; the seeded Go parity test as
acceptance evidence; F7 as definition of done; `role_catalog`'s shape including `area_assignable`
replacing subset-by-coincidence; DD-6's grounds; keeping `role_capabilities` as a separate bundle
table (its *shape*, not its disposition); §6.2's F8/rw-tx coupling; ME-09's framing; the junk-row
NOTICE handling; `subject_user_id` as TEXT — the type is right per the tokens lesson, it was the
missing FK that was not.

### Standing lesson

Two of the three worst findings (R2-1, R2-2) are **wrong-noun failures inside mechanisms written
because of ME-02**, by an author who had just written ME-02 down. That is the evidence that wrong-noun
is not a mistake one learns past — it is the default outcome whenever a mechanism targets a syntactic
proxy for a semantic property. ME-12 records it as a class.
