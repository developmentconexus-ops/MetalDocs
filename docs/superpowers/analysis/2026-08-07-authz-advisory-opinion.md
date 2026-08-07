# AuthZ grant-unification — independent advisory opinion (2026-08-07)

**Advisor:** Fable 5 (high effort), commissioned by the operator as specialist reviewer.
**Standing:** advisory. Not a ratification and not a gate. Recorded verbatim in substance so the
design spec can dispose of each point on the record rather than absorbing it silently.

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
| DD-3 — send back for restructure (revision-scoped floor) | **operator decision — open** |
| Role catalog: reference data vs tenant data | **operator decision — open** |
| ME-08 immediate deletion of the coverage endpoint | **operator decision — open** (issue #81) |
| ME-07 adopt-post-v1 recommendation + vendor split | **recorded**, no decision now (issue #80) |
| Phantom-control instrumentation defect class | **accepted** for the next catalog pass |
