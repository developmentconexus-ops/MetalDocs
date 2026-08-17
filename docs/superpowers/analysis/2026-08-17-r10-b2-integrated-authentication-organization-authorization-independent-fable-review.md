# MetalDocs R10-B2 — Integrated Authentication / Organization / Authorization — Independent Cold Adversarial Review

> **Status:** INDEPENDENT REVIEW — EVIDENCE ONLY / **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed candidate:** `docs/superpowers/analysis/2026-08-17-r10-b2-integrated-authentication-organization-authorization-fable-review-request.md` @ `b814f67284badd00182ff3c0abb77a66b448d7c9`
> **Architecture authority baseline:** `71791dfecd4cd185684373ffcdccbf256138b741` (R10-B2-1 promotion)
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — this review changes no authority, code, schema, OpenAPI, frontend or provider configuration.**

---

# 0. Bootstrap and evidence base

Read in authority order before reviewing: `AGENTS.md` → Method mirror → `wiki/references/current-agent-handoff.md` → `wiki/architecture/cohesive-platform-redesign.md` → frozen ledger `2026-08-14-cohesive-platform-redesign-ledger.md` → `wiki/architecture/r10-technical-architecture.md` (B1 + promoted B2-1). The full integrated packet was read end-to-end. The B2-2 candidate/independent review and commit `8e6634cc` (whole-system conformance review request) were treated strictly as evidence, never authority. Current implementation was consulted as evidence only:

```text
db/baseline/0001_current_schema.sql:1067  auth_sessions (session_id/user_id TEXT, ip, user_agent, last_seen_at, tenant_id)
db/baseline/0001_current_schema.sql:1255  iam_groups (id, tenant_id sentinel default, name, description)
db/baseline/0001_current_schema.sql:1270  iam_user_roles (global role, CHECK of 5 legacy codes incl. system_admin)
db/baseline/0001_current_schema.sql:~1285 iam_users (user_id TEXT, display_name inline, is_active + deactivated_at, last_login_*, mfa_*)
db/baseline/0001_current_schema.sql:1651  user_process_areas (7 legacy area roles, effective_from/effective_to intervals, granted_by/revoked_by)
internal/platform/iamtypes/role.go:28-35  legacy role enum
internal/modules/iam/api/api.gen.go:33-39,147-154  8-role legacy vocabulary (system_admin/area_admin/qms_admin/editor/signer/author/approver/viewer)
```

Method compliance: every material finding below carries evidence → root cause → target invariant → local-vs-global evaluation → enforcement → proof strategy → decision. The four mandated passes (Structural Inversion, Subtractive, Authority Duplication, Failure Class) are §4. All 18 scenarios are walked in §7; all 8 alternatives in §8; all 30 proof questions in §9; all 40 decisions in §6.

---

# 1. VERDICT

```text
APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 3   (M1 role↔scope matrix must be structurally enforced;
               M2 deterministic lock-order law must be codified — reachable deadlock otherwise;
               M3 Group hard-deletion law incomplete against live cross-owner references)
LOW     = 5   (L1–L5)
```

No finding invalidates the integrated architecture. All three MAJORs are bounded additions/corrections inside the candidate's own structure; none forces a family, table or ownership change. Batching did **not** hide a local maximum — it surfaced M1 and M3, which the per-family review cadence would likely have missed (both are cross-family composition defects).

---

# 2. Material findings

## M1 — MAJOR — The role↔scope compatibility matrix is correct but structurally unenforced; without a DB CHECK a real privilege-escalation path exists

**Evidence.** Candidate §5.4 proposes `tenant_owner → TenantScope only`, `area_manager → AreaScope only`, other three roles → both, and explicitly asks for verification. Candidate §5.7 grants `AreaScope(A) access.manage → may manage AreaScope(A) grants` **with no role restriction**. The RoleAssignment DDL (§5.2) carries XOR checks and a `role_code` CHECK but **no role↔scope compatibility constraint**, and no section states where the §5.4 matrix is enforced.

**Attack that fires.** An Area-scoped access administrator (holder of `access.manage` effective at AreaScope(A)) issues `RoleAssignment(subject=self, role_code='tenant_owner', area_scope_id=A)`. Under §5.7 the write is authorized (target scope = A). Under §5.6 evaluation, every **Area-targeted** check in A is now satisfied by the tenant_owner bundle — which contains permissions frozen as *tenant-owner-only* (`legal_hold.manage`, `disposition.manage`, `retention.extend`, `document_type.manage`, …, ledger §1). Whether those leak depends on an **unstated per-permission check-site classification** (Tenant-wide vs Area-targeted check). Ledger §12 makes LegalHold scope-able to Document/Dossier — i.e. plausibly evaluated as Area-targeted for a Document in A. The candidate currently relies on nothing preventing `tenant_owner@AreaScope` from existing.

**Root cause.** The matrix was proposed as descriptive vocabulary, not as an enforced invariant; grant-write authorization (§5.7) and grant-shape legality (§5.4) were designed on separate pages and never composed. This is precisely the cross-family defect class the integrated batch exists to catch.

**Target invariant.** A RoleAssignment whose `(role_code, scope type)` pair is outside the accepted matrix must be **unrepresentable**, independent of any application-path check and independent of any future per-permission check-site classification.

**Local vs global maximum.** Enforcing the matrix only in the grant service is the local maximum (every future write path must re-know the rule; DB backstop principle of B1 violated). The global maximum is schema-level unrepresentability plus the same rule at the write path for friendly errors.

**Required fix (exact).** Add to RoleAssignment:

```text
CHECK (
     (role_code = 'tenant_owner' AND tenant_scope_id IS NOT NULL)
  OR (role_code = 'area_manager' AND area_scope_id   IS NOT NULL)
  OR (role_code IN ('author','approver','viewer'))
)
```

(the XOR scope check already guarantees exactly one scope column is set, so the two restricted arms fully pin the scope type). Additionally state one successor law: **every permission-checking consumer (B3–B5/R10-E) must declare, per check site, whether the check target is Tenant-wide or Area-targeted; permissions frozen as tenant-owner-only whole-company families are Tenant-wide checks.** With the CHECK in place this classification is defense-in-depth rather than the only barrier.

**Matrix verification result.** The proposed matrix itself is **confirmed correct** against frozen authority: `tenant_owner` is a whole-company bundle (all 43 permissions; area-scoping it is semantically void and, as shown, dangerous); `area_manager` at TenantScope would constitute a de-facto sixth role ("manager of all areas") that the frozen five-role catalog does not contain — the company-wide need is served by `tenant_owner` or per-Area grants; `author/approver/viewer` at both scopes are legitimate (company-wide read/author/approve are real product situations). **Exact corrected matrix = the proposed matrix, unchanged, plus the CHECK above.**

**Proof strategy.** Negative fixture: INSERT `tenant_owner` + `area_scope_id` must fail at constraint level; grant-service test proving the friendly-path rejection; escalation test S16 extended with the self-grant `tenant_owner@AreaScope` attempt.

**Decision.** ACCEPT WITH FIX (IB2-D19). Bounded schema/text delta; no structural change.

## M2 — MAJOR — A deterministic lock-order law exists and must be codified; without it the candidate's own transactions contain a reachable deadlock

**Evidence.** Candidate §7.8 declines to freeze SQL but requires B2-4 to leave "an implementable deterministic lock-order law", and asks the reviewer to confirm or correct. The batch itself **is** B2-4, so the law must land in the corrected target, not remain hand-waved.

**Concrete counterexample under the candidate as written.** Offboarding (§6.2) revokes *all ApplicationSessions of the User*; binding disable/replacement (§7.2, promoted B2-1) revokes *all Sessions of the binding*. For a User with one enabled binding these are the **same multi-row set**. Two concurrent transactions updating the same multiple session rows in different physical orders is the classic multi-row deadlock; nothing in the candidate fixes an order. A second instance: group deletion (S8) deletes memberships of Group g while offboarding deletes memberships of User u — single shared row `(u,g)` per pair, so this pair alone cannot cycle, but combined with unordered session/RA child deletes the general hazard stands.

**Root cause.** Child-row mutation ordering was treated as implementation noise rather than part of the concurrency law. Parent-row ordering (User/Binding/Area) was implied by the per-invariant sections but never composed into one acquisition order.

**Target invariant.** All B2 transactions acquire locks consistently with one global partial order; therefore no cycle is constructible and READ COMMITTED + narrow locks suffice, exactly as B1 demands.

**Confirmed answer to §7.8: option 1 — a simple deterministic order exists.** Required fix (exact law to record):

```text
Canonical B2 lock acquisition order (partial order; classes may be skipped, never revisited backwards):

  1. User row
  2. ProviderSubjectBinding rows of that User, ascending id
  3. Area row
  4. child-row sets, each in ascending primary-key order:
       ApplicationSession by id
       GroupMembership by (user_id, group_id)
       RoleAssignment by id
  Group row is its own isolated class: group deletion locks the Group row first (FOR UPDATE),
  then its memberships (ascending user_id), then its RoleAssignments (ascending id); group
  operations touch no User/Binding/Area locks and therefore cannot join a cycle with classes 1–3.

Lock modes:
  eligibility/acceptance readers (issuance, membership add, grant insert) take FOR SHARE;
  lifecycle mutators (offboard, re-enable, retire, binding disable/replace) take FOR UPDATE.
  FOR KEY SHARE is insufficient: User.disabled_at / Area.disabled_at updates are non-key
  updates (FOR NO KEY UPDATE), which FOR KEY SHARE does not conflict with — the C1/C5
  serialization would silently not exist.
```

Verification against every candidate operation: issuance = User(S)→Binding(S); offboarding = User(X)→Bindings(asc id)→Sessions(asc id)→Memberships(asc)→RoleAssignments(asc); binding disable/replace = Binding(X)→Sessions(asc id) (skipping class 1 is legal — it never *returns* to class 1); direct Area grant = User(S)→Area(S)→insert; Area retire/re-enable = Area(X) only; membership add = User(S)→insert (FK takes KEY SHARE on Group row; group deletion's FOR UPDATE on the Group row conflicts, giving clean serialization with FK fail-closed after commit). Exhaustive pairwise composition yields no cycle. The FOR SHARE / FOR NO KEY UPDATE conflict-matrix subtlety is the one place a naive implementation would silently lose C1; the mode law above closes it.

**Decision.** ACCEPT WITH FIX (IB2-D33): record the order + mode law verbatim in the corrected target. Deadlock question (proof Q19): **no cycle under the codified law; reachable deadlock without it.**

## M3 — MAJOR — Group hard-deletion law is incomplete: live cross-owner references to Group identity exist outside the batch and S8 as written is unsafe

**Evidence.** Candidate §4.5 gives Group no lifecycle; S8 says deletion requires first removing memberships and RoleAssignments, "then delete Group", and claims historical safety because Approval/Distribution snapshot concrete Users. That claim covers **historical** references only. Frozen ledger §2 makes `Group` a live **ApprovalPolicy Step actor_rule** target (`NamedUser | Group | RoleInArea`), and Distribution audience configuration is group-expressed before release-time snapshotting. Those are *forward-looking live* references: an ApprovalPolicy in force that names Group g must not silently dangle when g is deleted. Note the asymmetry the candidate itself creates: `NamedUser` targets can never dangle (User is never hard-deleted — skeleton retained), `RoleInArea` targets cannot dangle (Area is never deleted, only retired), but `Group` — the only hard-deletable actor_rule target — has no protection.

**Root cause.** The batch boundary (AuthN/Org/AuthZ) excluded Approval/Distribution, and S8 asserted deletion safety from inside the boundary only. Same defect class as M1: composition across the batch edge.

**Target invariant.** Group deletion fails closed while any live cross-owner configuration references the Group. Structural realization under B1 reference law: consumers that persist a live Group reference hold an ordinary typed FK → `group(id)` with RESTRICT/NO ACTION, so deletion is structurally blocked; consumers that snapshot (Approval participants, Distribution denominators) need and get nothing.

**Required fix (exact).** Extend the Group law: *"Group deletion additionally fails closed while any live cross-owner reference to the Group exists (known V1 consumers: ApprovalPolicy Step actor_rule, Distribution audience configuration). Realization is the B1 typed-FK RESTRICT law at each consumer; B4/B5 must declare their Group-referencing persistent families and confirm the FK, as a named successor obligation."* S8 gains one sentence: emptying memberships/grants is necessary but not sufficient; cross-owner references must be absent or the DELETE fails closed.

**Decision.** ACCEPT WITH FIX (IB2-D10). No new lifecycle state is introduced — deletion stays hard, YAGNI on Group retirement holds; the fix is a reference law, not a state machine.

---

# 3. Low findings

**L1 — Offboarding must say "revoke", not "revoke/delete", for Sessions.** §6.2 writes "revoke/delete all ApplicationSessions". Promoted B2-1 makes `revoked_at` terminal and row erasure a *privacy-lifecycle* operation (§7.9/§10). Offboarding must set `revoked_at`; deletion inside offboarding would conflate eligibility lifecycle with privacy erasure and erase evidence prematurely. One-word fix.

**L2 — First-grant bootstrap and last-admin lockout need one explicit sentence.** Default-deny + no-bypass + destructive offboarding means: (a) a fresh deployment has zero RoleAssignments and no request-path actor can create the first one; (b) the sole TenantScope `access.manage` holder can be offboarded (including self-offboard), leaving no request-path recovery. Neither is a structural dead end — B1 §6.7's non-serving maintenance trust surface is the correct recovery/bootstrap path — but the target must *say* that initial `tenant_owner` seeding and admin-lockout recovery are maintenance-surface operations (R10-F/ops), and R10-E may add a friendly last-admin guard. Absent this sentence, implementers will invent a bootstrap bypass.

**L3 — Static-catalog ↔ DB-CHECK parity is a named proof obligation.** The role list now exists in exactly two places: the product catalog (authority) and the `role_code` CHECK (+ the M1 matrix CHECK) as enforcement. That is B1's TEXT+CHECK law, not a second authority — but it is also the exact hand-synced-enumeration defect class this repo has already catalogued. A parity guard (catalog ↔ CHECK drift fails CI/migration test) must be listed as an implementation-stage proof obligation.

**L4 — The exact 5×43 role→permission bundle matrix must be pinned in durable authority at promotion.** The ledger states the 27+16 permission codes and the R9.5 per-role additions, but the complete base-27 bundle mapping lives only in archived R9 material. The candidate correctly refuses to redesign bundles, but B2 promotion should restate the full matrix once in the promoted target (or the ledger) so B2-3 evaluation and B3+ consumers implement from durable authority, not from archaeology or legacy code.

**L5 — Name-normalization residue (carried from B2-2 review L5).** The candidate absorbed the Area-code normalization sentence but remains silent on `Group.name` / `Tenant.display_name` / `Area.name` blank/case handling. One sentence ("implementation-spec detail; no case-insensitive uniqueness as accidental identity semantics") prevents drift. Non-semantic.

**B2-2 absorption check (evidence):** L1 absorbed (§4.2 code-verbatim sentence); L2 **closed by decision** (destructive offboarding + no-restore re-enable — the integrated candidate resolves the deferred restore-vs-regrant gate exactly as the B2-2 review demanded); L3 absorbed (§4.6 pair-PK ruling recorded); L4 absorbed (§4.4 profile-presence law); L5 partially absorbed → carried here.

---

# 4. Mandated Method passes

## 4.1 Structural Inversion Test

*If MetalDocs had been designed from day one with Keycloak, one company per deployment, separate Organization/AuthZ and no legacy IAM tables, what B2 state and transaction laws would still necessarily exist?*

Necessarily re-derived from requirements alone: a stable person identity separate from provider identity (`User.id`); an erasable human-enrichment satellite (privacy law forces the split — you cannot erase what is also your FK anchor); a provider correlation fact keyed `(issuer,subject)` with acceptance state and uniqueness both ways; an opaque local session with digest-only storage, finite expiry, terminal revocation; an organizational Area as scope target; flat Groups + explicit membership; **one** additive grant fact (subject × role × scope) with default-deny evaluation; a static role/permission vocabulary (the product defines the roles — a from-scratch single-company product would never build configurable RBAC tables for five fixed roles); same-commit audit for access mutations; an offboarding transaction that revokes sessions and removes grants; durable provider intent instead of cross-DB atomicity. Even the singleton Tenant row survives inversion — not as tenancy, but as the FK target TenantScope needs and the deployment↔DB identity anchor the fail-closed handshake needs.

The inversion reproduces the candidate almost field-for-field. Nothing in the candidate exists merely because legacy IAM existed; conversely the legacy schema (dual-source `iam_user_roles`+`user_process_areas`, interval-retaining grants, monolithic `iam_users` with MFA/last-login mirrors, `tenant_id` sentinel defaults) fails the inversion at every point. **The candidate is inversion-stable; the incumbent is not.**

## 4.2 Subtractive pass

Attacked every §15 target. Result: **the integrated target is at its subtractive floor** — no element can be deleted without weakening a distinct invariant:

| Element | Verdict |
|---|---|
| `Tenant.display_name` | KEEP — the ledger requires editable company identity as a mutable fact; deleting it leaves `tenant.settings.manage` with no object and company name homeless |
| `Area.disabled_at` | KEEP — the one real lifecycle requirement (B2-2 F1), consumed by C5 and the fail-closed new-reference law |
| Area reversibility | KEEP — terminal retirement + total `UNIQUE(code)` would permanently consume the code namespace (an org could never recreate "ENG"); re-enable restains nothing because retirement disables nothing (§5.5 below) |
| `User.disabled_at` | KEEP — the single eligibility authority; C1 depends on it |
| `UserProfile` table / `email` | KEEP — erasure boundary requires the split; Notifications (delivery projection) is the real email consumer |
| `Group` / lifecycle omission | KEEP — flat identity only; M3 adds a reference law, not a lifecycle |
| `GroupMembership` explicit row | KEEP — sole membership truth; nothing else can express it |
| `RoleAssignment.id` | KEEP — see D15: the XOR shape has no NULL-free natural key, so a PK requires the surrogate |
| RoleAssignment current-only | KEEP — Audit owns history; no domain consumer replays grants (Approval/Distribution snapshot) |
| Static catalogs | KEEP — DB tables would be a second authority + unsupported configurability |
| Four partial uniques | KEEP — collapsing to COALESCE expression indexes would be polymorphic-fallback modeling and lose typed honesty; four explicit backstops are the enforcement floor |
| Explicit `tenant_scope_id` FK | KEEP — a sentinel/NULL-means-Tenant encoding would be a magic value, banned by frozen law |
| Membership Tenant-only admin | KEEP — §5.7 analysis: no safe area-local alternative exists without scoped Groups |
| Offboarding grant/membership deletion | KEEP — deleting the deletion resurrects the silent-restore hazard |
| Provider disable intent | KEEP — conditional ("when required"); without it provider truth drifts unbounded |
| Audit same-commit categories | KEEP all listed — each mutates identity/eligibility/access; even Tenant display rename is identity evidence (name appears on governed renditions) |

Nothing further is deletable. The additions required are exactly M1's CHECK, M2's law text, M3's reference law, and L1/L2/L5 sentences — all enforcement/wording, no new state.

## 4.3 Authority duplication pass

Searched all ten mandated concepts for dual authority:

```text
person identity            → User only                                     SINGLE
provider identity          → ProviderSubjectBinding (issuer,subject)       SINGLE
User eligibility           → User.disabled_at only                         SINGLE (issuance additionally requires accepted binding — conjunction, not duplication)
group membership           → GroupMembership row presence                  SINGLE
role grant                 → RoleAssignment row presence                   SINGLE
permission bundle          → static product catalog                        SINGLE (DB CHECKs are enforcement, not authority — L3 parity guard required)
scope                      → typed tenant_scope_id/area_scope_id           SINGLE
Area retirement            → Area.disabled_at                              SINGLE
Session validity           → expires_at + revoked_at (no status enum)      SINGLE
historical actor display   → owning families' snapshots + Audit skeleton;  NO DUPLICATION
                             UserProfile is current display only — different
                             meaning (point-in-time evidence vs current), allowed
```

Contrast with the incumbent, which duplicates grant authority (`iam_user_roles` vs `user_process_areas`, different role vocabularies at `role.go:28-35` vs `user_process_areas` CHECK), duplicates eligibility (`is_active` + `deactivated_at`), and mirrors provider state (`mfa_enabled/mfa_enrolled_at`). The target removes every one. **No duplicate authority found in the candidate.**

## 4.4 Failure-class pass

```text
privilege escalation           → ONE reachable path found and closed (M1 tenant_owner@AreaScope self-grant);
                                 Group-indirection escalation blocked by Tenant-only membership admin (S16)
stale grants                   → impossible as authority: evaluation reads live rows; no snapshot/cache exists (D20)
stale memberships              → same; Session carries no AuthZ state (B2-1)
silent privilege restoration   → structurally impossible: restored User has no rows to reactivate (deleted at offboard);
                                 Area re-enable restores nothing because retirement disabled nothing
offboarding partial commit     → impossible: single local tx (S18); composition seam per B1 §6.8
login/offboarding race         → C1 + M2 lock modes: total order proven
grant/offboarding race         → C4: same User-lock discipline
membership/offboarding race    → C3: same
Area retirement/grant race     → C5: Area row serialization; no grant born after retirement from stale check
binding/session race           → promoted B2-1 C3 retained unchanged
deadlock                       → reachable without M2's law; none under it
provider uncertainty           → never fabricates binding truth (B2-1 six outcomes retained); intents reconcile via R10-D
privacy erasure/restore        → profile/binding/session erasable; User skeleton + PII-minimized Audit survive;
                                 restore non-resurrection is a named R10-C proof
historical reference breakage  → User never deleted; Area never deleted; Group = M3 fix; snapshots make
                                 Approval/Distribution independent of live Org rows
generic IAM creep              → none: no configurable roles/permissions, no polymorphic subject/scope, no deny
                                 engine, no ReBAC, no effective-permission store, no privacy platform
```

---

# 5. Required sub-verdicts

## 5.1 Method / Global Maximum — **PASS**

The integrated candidate is the global maximum for the real constraints: fixed product semantics → static catalogs; single company → no partition plumbing; privacy law → User/Profile split; fail-secure law → destructive offboarding; B1 substrate → typed FKs, READ COMMITTED + narrow locks. Batching produced net review value: M1 and M3 are batch-edge composition defects invisible to per-family review. No incorrect local maximum is hidden; the three MAJORs are corrections *inside* the winning structure, not alternative structures.

## 5.2 Organization — **APPROVE**

Six families (Tenant, Area, User, UserProfile, Group, GroupMembership) remain correct after Authorization and transactions are included. Constant-expression singleton index + promoted handshake = exactly-one Tenant, proven both directions. Area identity/retirement law is right (M3 touches Group, not Area). User/UserProfile split survives the integrated re-attack (§8-H). GroupMembership pair-PK ruling stands.

## 5.3 Authorization relational model — **APPROVE WITH M1**

One XOR-typed RoleAssignment table with real FKs to all four possible endpoints, four partial-unique duplicate backstops, current-only INSERT/DELETE semantics. Superior to every alternative compared (§8). Requires M1's compatibility CHECK to be complete.

## 5.4 Role/Permission static-catalog — **APPROVE**

Five roles and 43 permissions are frozen product vocabulary, not deployment data. DB tables would (a) create a second authority against product semantics, (b) manufacture unsupported custom-role capability, (c) reintroduce the hand-synced-enumeration defect class at higher blast radius. No deployment-specific consumer requiring rows was found: Admin UI reads the catalog through the API; Audit references `role_code` text. L3 (parity guard) and L4 (pin the full bundle matrix in durable authority) attach.

## 5.5 RoleAssignment UUID / current-history — **APPROVE, with the honest rationale recorded**

The UUID is retained, but the load-bearing reason is structural, not attributional: the XOR shape means every natural-key column set contains NULLs, and PostgreSQL primary keys cannot include nullable columns — the four partial unique indexes cannot be a PK. A table with no PK violates B1's identity law; therefore the surrogate UUID is **necessary**, not stylistic. (GroupMembership's NULL-free pair key is why it legitimately differs.) Current-only + same-commit Audit is sufficient evidence: no frozen consumer replays historical grants — Approval snapshots participants, Distribution snapshots audiences, at their own boundaries. Record this rationale in the promoted target so the asymmetry with GroupMembership does not look arbitrary later.

## 5.6 Role↔scope compatibility — **matrix CONFIRMED; enforcement M1**

Exact corrected matrix = the proposed matrix (no change). Enforcement must be the M1 CHECK + write-path validation + the per-check-site scope-classification successor law.

## 5.7 GroupMembership administration / security — **APPROVE**

TenantScope-only membership administration is correctly conservative, not over-restrictive. The adversarial attempt to beat it: "allow AreaScope `access.manage` to manage membership of Groups whose entire grant surface lies within that Area" fails structurally — the grant surface is mutable, so a membership admitted when G was A-local silently escalates the moment any TenantScope admin grants G a wider role; making it safe requires freezing Group scope (= scoped Groups, alternative F, rejected) or making membership legality depend on cross-time assumptions no lock can express. Area-local delegation remains available through direct per-User AreaScope grants, which stay entirely inside the Area admin's authority. The composition risk (membership add activates all of G's grants) is placed at the only tier that can see G's whole grant surface. Correct.

## 5.8 Offboarding / re-enable — **APPROVE (L1 wording)**

Destructive offboarding is the right fail-secure semantics. Regrant friction after re-enable is bounded and evidence-assisted (Audit shows exactly what existed); silent privilege restoration after a two-year rehire is an unbounded latent escalation. No frozen product requirement for temporary suspension distinct from offboarding exists; if one materializes it is a reopen with its own state, never a reinterpretation of `disabled_at`. Rehire safely reuses the same User (S6): identity, governed history and binding correlation all survive; access does not. Binding retention (D26) is correct — the `(issuer,subject)→User` correlation remains *true* after employment ends; deleting it would be falsifying history, and eligibility gating already makes it inert. Same-commit Audit + conditional provider-disable intent close the truth loop.

## 5.9 Area retirement / re-enable — **APPROVE**

Retirement blocks new references only; existing Documents, grants and historical evidence remain valid — this is what prevents retirement from becoming an accidental deny-all (S14). Preserving existing grants is correct: revoking them would orphan historical content access, and the frozen requirement that authorized actors keep reading governed history would break. Re-enable is sound precisely because retirement never disabled anything: nothing is restored, only future references become legal again. The asymmetry with User offboarding is principled, not inconsistent — User disable is an *access-termination* event (person left), Area disable is a *structural-freeze* event (org chart changed); their re-enables therefore differ correctly. Terminal retirement fails the subtractive test (§4.2): with immutable total-unique codes it permanently consumes the code namespace and would push real orgs toward code mutation or duplicate-code hacks — strictly worse.

## 5.10 Canonical evaluation / default-deny — **APPROVE**

Live evaluation over current rows + static bundles + owner relationship predicates + governance constraints preserves the frozen equation exactly. No persisted effective permissions, no Session snapshot, no cache authority. Query shape at V1 scale (one company, hundreds of users, two-level union) is trivially indexable; no structural dead end. A rebuildable projection remains a legal future *mechanism* if measured evidence ever demands it.

## 5.11 Transaction / concurrency / deadlock — **APPROVE WITH M2**

All declared invariants (C1–C7) are cheaply implementable under READ COMMITTED + narrow row locks + the existing UNIQUE/CHECK/FK backstops. No SERIALIZABLE, no advisory-lock framework needed. Deadlock: reachable as written; none under M2's codified acquisition order and lock-mode law. In-flight request posture (§7.9/D37) is honest and sufficient: fail-closed future resolution after commit is the industry-correct contract; frozen law demands no linearization of already-authorized in-flight work, and pretending otherwise would be a false claim the Method forbids.

## 5.12 Audit / durable-intent — **APPROVE**

The same-commit list (§9.1) is neither too broad nor too narrow: every listed operation mutates identity, eligibility, acceptance or effective access. Ordinary login/logout correctly stay out of semantic Audit (they are authentication traffic, not governed mutation; security telemetry is an R10-E/ops concern). Provider effects correctly never share a transaction; durable intent + R10-D reconciliation matches B1 §6.9 and promoted B2-1 §7.8. The revocation-history question is resolved correctly: no domain consumer needs grant history as *domain evidence* (Approval/Distribution snapshot), so Audit-as-timeline is the right and only home — "domain evidence records remain authorities" is not violated because no domain authority for grant history is required to exist.

## 5.13 Privacy — **APPROVE**

Erasable surface (UserProfile, Binding, Sessions) vs retained skeleton (`User.id` + `disabled_at` + timestamps — non-PII) is exactly the frozen posture. Bare UUID + eligibility timestamps survive lawful erasure without carrying PII. Binding-erasure surrendering structural no-recorrelation is correctly inherited from promoted B2-1. Neutral-actor display fallback covers historical UI. No privacy platform introduced. B6 field-by-field Audit proof and R10-C restore non-resurrection proof remain correctly routed successors.

---

# 6. Decision disposition — IB2-D01..IB2-D40

| # | Decision | Disposition |
|---|---|---|
| D01 | Tenant = id + display_name only | **ACCEPT** |
| D02 | singleton = constant-expression unique + readiness at-least-one | **ACCEPT** |
| D03 | Area = id/code/name/disabled_at; no hierarchy | **ACCEPT** |
| D04 | Area code immutable; retirement blocks new references only | **ACCEPT** |
| D05 | Area disabled_at reversible | **ACCEPT** |
| D06 | User = id + disabled_at only | **ACCEPT** |
| D07 | User/UserProfile split | **ACCEPT** |
| D08 | email/username never identity; no UNIQUE(email) | **ACCEPT** |
| D09 | no User home_area | **ACCEPT** |
| D10 | Group = id + unique name; flat/company-wide/no lifecycle | **ACCEPT WITH FIX — M3** (deletion fail-closed vs live cross-owner refs; B4/B5 successor declaration) |
| D11 | GroupMembership pair PK/current-only/no UUID/history | **ACCEPT** |
| D12 | Role/Permission catalogs static product authority | **ACCEPT** (L3 parity guard; L4 pin bundle matrix) |
| D13 | one persisted Authorization family = RoleAssignment | **ACCEPT** |
| D14 | typed subject XOR + typed scope XOR | **ACCEPT** |
| D15 | RoleAssignment UUID retained | **ACCEPT** (record structural rationale — §5.5) |
| D16 | current-only INSERT/DELETE; Audit owns revoke history | **ACCEPT** |
| D17 | four duplicate-grant uniqueness backstops | **ACCEPT** |
| D18 | explicit Tenant FK = semantic TenantScope | **ACCEPT** |
| D19 | proposed role↔scope compatibility matrix | **ACCEPT WITH FIX — M1** (matrix confirmed; DB CHECK + check-site scope law required) |
| D20 | no persisted effective permissions | **ACCEPT** |
| D21 | TenantScope satisfies matching area-level checks; AreaScope only matching Area | **ACCEPT** |
| D22 | GroupMembership mutation requires TenantScope access.manage | **ACCEPT** |
| D23 | RoleAssignment mutation requires access.manage at target scope | **ACCEPT** (escalation hole closed by D19's CHECK, not by weakening this rule) |
| D24 | Group identity organization.manage; membership access.manage | **ACCEPT** |
| D25 | offboarding deletes Sessions+Memberships+direct RAs atomically | **ACCEPT WITH FIX — L1** (Sessions are *revoked*, terminally; never deleted in offboarding) |
| D26 | offboarding retains ProviderSubjectBinding correlation | **ACCEPT** |
| D27 | re-enable never restores deleted access configuration | **ACCEPT** |
| D28 | privacy cleanup deletes enrichment/auth state, retains User skeleton | **ACCEPT** |
| D29 | User lock serializes issuance/member-add/direct-grant/offboarding | **ACCEPT** (mode law in M2 binds it) |
| D30 | Area lock serializes retirement/re-enable vs new AreaScope grants | **ACCEPT** |
| D31 | Binding lock discipline remains promoted B2-1 law | **ACCEPT** |
| D32 | no tx couples Group grant + GroupMembership; conjunction is live truth | **ACCEPT** |
| D33 | deterministic narrow-lock ordering sufficient under READ COMMITTED | **ACCEPT WITH FIX — M2** (order exists; must be codified verbatim) |
| D34 | same-commit Audit for material B2 mutations | **ACCEPT** |
| D35 | provider durable intent shares local commit | **ACCEPT** |
| D36 | no cross-provider transaction | **ACCEPT** |
| D37 | bounded in-flight request semantics | **ACCEPT** |
| D38 | canonical evaluation = live state + static bundles + domain predicates | **ACCEPT** |
| D39 | retired Area existing grants/content usable; only new refs blocked | **ACCEPT** |
| D40 | no generic IAM/privacy/RBAC platform | **ACCEPT** |

No REJECT. No DEFER. 35 ACCEPT + 5 ACCEPT WITH FIX (D10, D19, D25, D33 + D15's recorded-rationale note being editorial).

---

# 7. End-to-end scenario walk — S1–S18

**S1 provisioning** — PASS. Two-transaction choreography matches promoted B2-1 §7.8; no email auto-binding path exists anywhere in the state model (no email column participates in identity); Session requires eligible User ∧ accepted binding.

**S2 direct Area grant** — PASS. `access.manage@A` + eligible User (FOR SHARE) + active Area (FOR SHARE) + insert + same-commit Audit. M1's CHECK additionally bounds which roles are grantable at AreaScope.

**S3 group-mediated grant, both orders** — PASS. Membership-first: grant activates for members at grant commit. Grant-first: membership add activates at membership commit — and the actor for that add is TenantScope `access.manage`, the tier able to see G's full grant surface before committing. No order yields access neither authorized tier intended.

**S4 Area retirement** — PASS. Existing references valid; new Document assignment / AreaScope grant / policy reference fail closed at their owning boundaries. Re-enable restores nothing (nothing was disabled) and only re-legalizes future references after Area-lock commit.

**S5 offboarding, full blast radius** — PASS with L1 wording (revoke, not delete). Single tx: disable + revoke Sessions + delete memberships + delete direct RAs + Audit + conditional provider intent. Group RAs correctly survive (they are Group facts). No resurrection path exists afterward.

**S6 rehire much later** — PASS. Same `User.id`; `disabled_at` cleared; binding row re-usable per B2-1 (same subject, same row); zero memberships/grants exist → default deny until fresh explicit grants. This is the scenario that proves destructive offboarding right.

**S7 privacy cleanup** — PASS. Session rows (already terminal) erasable → Binding erasable (RESTRICT order per B2-1 §7.9) → Profile erasable; `User.id` skeleton + governed history intact; UI neutral fallback (§4.4 law); restore non-resurrection = named R10-C proof.

**S8 Group deletion** — **FAILS AS WRITTEN → M3.** Memberships+RAs removal is insufficient; live ApprovalPolicy actor_rule / Distribution audience references must fail the deletion closed. With M3's reference law: PASS.

**S9 provider outage** — PASS. Established Sessions continue locally; new login/reauth fail visibly; intents retry via R10-D; Authorization is entirely local so no degradation of decisions.

**S10 grant revoke with live Session** — PASS. Next evaluation sees no row → deny. No session revocation needed; C4 holds because Session has no AuthZ snapshot.

**S11 membership removal with live Session** — PASS. Same mechanics.

**S12 concurrent offboarding vs login/member-add/direct-grant** — PASS *under M2's lock-mode law*. All three racing operations FOR SHARE the User row; offboarding FOR UPDATEs it; READ COMMITTED total order + post-lock re-read of `disabled_at` gives exactly the two legal outcomes of C1/C3/C4. Without the mode law a KEY-SHARE implementation silently loses the conflict — this is why M2 is MAJOR.

**S13 concurrent Area retirement vs AreaScope grant** — PASS under the same discipline (grant FOR SHARE Area; retirement FOR UPDATE Area; grant re-checks `disabled_at` after lock).

**S14 retired-Area existing access** — PASS. Grants and content untouched by retirement; only new references blocked. No deny-all.

**S15 duplicate grants** — PASS. The four partial uniques cover all four subject×scope quadrants; XOR checks guarantee every row lands in exactly one quadrant, so no duplicate escapes coverage. (Same subject+role at both Tenant and Area scope is two distinct facts, correctly permitted.)

**S16 escalation via Group** — PASS for the membership vector (Area admin cannot touch membership; TenantScope-only). **The packet's scenario is however incomplete: the self-grant vector `tenant_owner@AreaScope(A)` was open (M1).** With M1's CHECK: unrepresentable. S16 must be extended with that fixture at implementation time.

**S17 binding disabled/re-enabled while User eligible** — PASS. Pure promoted B2-1 territory, consumed unchanged; Authorization contributes no alternate identity path (only `Session → Binding → User` exists).

**S18 partial offboarding under failure** — PASS. One local transaction; any failure rolls back all of it; no state where User is disabled but grants/sessions survive (or vice versa) is committable. Composition-seam execution per B1 §6.8.

---

# 8. Alternatives — mandated comparison

**1. DB-configurable Role/Permission/RolePermission** — REJECTED. No V1 consumer configures roles; tables would be dormant configurability = second authority + custom-role capability the frozen catalog forbids. Classic generic-IAM local maximum.

**2. Separate UserRoleAssignment/GroupRoleAssignment** — REJECTED. Two tables still carry the scope XOR (only a 4-table split removes all NULLs, quadrupling every consumer query, the admin surface, offboarding deletes and Audit attribution). One XOR table + 2 CHECKs + 4 partial uniques is less total surface with identical structural strength.

**3. Generic subject_type/subject_id + scope_type/scope_id** — REJECTED. Loses real FKs (dangling subjects/scopes become representable), invites ReBAC-graph creep, violates B1's typed-reference law. Saves nothing.

**4. Retained revoked grants / effective intervals** — REJECTED. This is the incumbent (`user_process_areas.effective_from/effective_to`, `revoked_by` — `0001_current_schema.sql:1651`) and it is an authorization event store: every current-truth query needs `WHERE effective_to IS NULL`, uniqueness needs partial predicates over liveness, and a second history authority competes with Audit. No frozen consumer reads historical intervals; Approval/Distribution snapshot. Audit-only history is strictly simpler and sufficient.

**5. Preserve memberships/grants across offboarding** — REJECTED. Dormant-but-restorable access is a latent escalation with an unbounded fuse (B2-2 review L2 named it; the integrated candidate closes it). Regrant friction is bounded, visible and Audit-assisted; silent restoration is invisible and unbounded. Fail-secure wins.

**6. Scoped / Area-local Groups** — REJECTED. Introduces Group scope semantics duplicating RoleAssignment's scope authority (two places would then answer "where does this collective act?"), plus scope-change migration semantics, for the sole benefit of area-local membership admin — which direct AreaScope User grants already provide safely.

**7. Materialized effective-permission projection** — REJECTED for V1. Live evaluation is a small indexed union at single-company scale; a projection adds invalidation correctness (grant/membership/offboard/retire all become dual-write) — pure accidental complexity now. Legal later only as rebuildable mechanism on measured evidence; never authority.

**8. Monolithic User with nullable/scrubbed PII** — RE-ATTACKED and REJECTED again, more strongly under the integrated model: erasure would become in-place UPDATE of the identity row that every governed FK anchors (mutating semantic-authority identity state), NOT NULL contracts would collapse to nullable-everything, and "erased vs never-provisioned" becomes indistinguishable. Row-delete of a satellite is structurally cleaner in every dimension. The integrated Authorization/offboarding semantics change nothing in favor of the monolith.

The candidate beats the strongest credible alternative in every family.

---

# 9. Required proof outputs — 30 answers

1. **Global Maximum: yes** — batching surfaced (not hid) the two cross-edge defects; structure survives (§5.1).
2. **Six-family Organization shape: correct** after integration (§5.2).
3. **User+UserProfile beats monolith: yes** (§8-8).
4. **Reversible Area retirement: correct**; terminal retirement fails the code-namespace test (§5.9).
5. **Group lifecycle omission: genuinely YAGNI**; M3 adds a reference law, not lifecycle.
6. **Pair-keyed GroupMembership: still correct** under access-management/offboarding semantics.
7. **Static catalogs: yes**, product authority, not DB tables (§5.4).
8. **RoleAssignment UUID: yes** — structurally necessary (no NULL-free natural PK) (§5.5).
9. **One XOR table > separate typed tables: yes** (§8-2).
10. **Current-only + Audit: sufficient**; no domain consumer of grant history exists (§5.12).
11. **Role↔scope matrix: correct as proposed**; must be CHECK-enforced (M1).
12. **Explicit tenant_scope_id: correct** TenantScope representation (real FK, no sentinel).
13. **Area-admin Group escalation: blocked** on the membership vector; self-grant vector required M1; both closed.
14. **TenantScope-only membership admin: correctly conservative**; every relaxation examined is unsafe (§5.7).
15. **Destructive offboarding: right** fail-secure behavior (§5.8).
16. **Re-enable default-deny: no operational dead end** — regrant is normal administration; L2 names the only true dead ends (bootstrap/last-admin) and routes them to the maintenance surface.
17. **No accepted temporary-suspension requirement exists**; future need = reopen with own state.
18. **All concurrency invariants cheaply implementable under READ COMMITTED: yes**, with M2's mode law.
19. **Deadlock cycle: exists as written** (unordered multi-row session revocation overlap); **none** under M2's codified order.
20. **In-flight posture: honest and sufficient** (§5.11).
21. **Area retirement preserves existing grants/content, blocks new refs: yes** (S4/S14).
22. **Audit same-commit scope: right-sized** (§5.12).
23. **Provider intent coupling: correct**; provider truth never enters Organization (§5.12).
24. **Privacy cleanup preserves governed history without a privacy engine: yes** (S7).
25. **No accidental tenant_id-like partition column**: `tenant_scope_id` is a semantic scope FK on one table, nullable, XOR-bound — not a universal partition dimension. Verified against every proposed table.
26. **Nothing further deletable** (§4.2 subtractive floor).
27. **No current-implementation consumer proves a missing target fact** — one candidate-side sweep note: legacy `iam_groups.description` is dropped; classified deliberate (non-semantic UX convenience; R10-E may reintroduce as display metadata without semantic law). See §10.
28. **43-permission/five-role semantics preserved**; no sixth role, no new permission invented anywhere in the candidate.
29. **Default-deny and no-bypass preserved**; offboarding-under-`organization.manage` is not a bypass — access/session teardown is a mandatory consequence of the authorized operation, not discretionary access administration.
30. **B2 closes cleanly**: one adjudication + one bounded delta suffices (§13).

---

# 10. Current-implementation evidence classification

| Current fact (evidence anchor) | Classification | Target disposition |
|---|---|---|
| Dual-source grants: `iam_user_roles` (global, 5 legacy codes) + `user_process_areas` (area, 7 legacy codes) with divergent vocabularies | **LEGACY MECHANISM** (ADR 0007's unfinished migration) | replaced by single RoleAssignment family |
| `user_process_areas.effective_from/effective_to` + `granted_by/revoked_by` in-row history | **LEGACY MECHANISM** | history moves to Audit; current-only rows (alternative 4 rejected) |
| 8-role legacy vocabulary (`api.gen.go:33-39,147-154`) incl. `system_admin` short-circuit | **LEGACY MECHANISM** | five frozen roles; no bypass role |
| `iam_users` monolith: `display_name` inline, `is_active` **and** `deactivated_at` (dual eligibility), `last_login_*` | **ACCIDENTAL COMPLEXITY** | User/UserProfile split; single `disabled_at`; telemetry out of semantic state |
| `iam_users.mfa_enabled/mfa_enrolled_at` | **LEGACY MECHANISM** (provider-state mirror) | deleted; Keycloak owns MFA |
| `iam_users.user_id` TEXT PK | **LEGACY MECHANISM** | UUID PK per B1 |
| `auth_sessions` with `ip_address/user_agent/last_seen_at` + `tenant_id` | **LEGACY MECHANISM** | ApplicationSession per promoted B2-1; telemetry only via a future R10-E consumer |
| `tenant_id uuid DEFAULT 'ffffffff-…'` sentinel defaults + FORCE RLS surfaces | **LEGACY MECHANISM** | no partition column; no RLS (B1) |
| `iam_groups.description` | **ACCIDENTAL COMPLEXITY** (deliberate drop; non-semantic) | R10-E may reintroduce as UX display metadata if a real admin-journey consumer appears |
| Dev-seed/first-admin bootstrap exists in current runtime | **KNOWN REQUIREMENT** | target must name its successor → **L2** (maintenance-surface seeding) |
| `document_process_areas.archived_at/is_active` archive semantics | **KNOWN REQUIREMENT** | already absorbed as `Area.disabled_at` (B2-2 F1) |

The sweep found **no** current fact the candidate accidentally dropped beyond L2's bootstrap sentence; every other divergence from the incumbent is deliberate and Method-justified.

---

# 11. Exact corrected integrated target — required deltas

All deltas are bounded text/DDL additions inside the candidate's structure:

1. **M1** — add to RoleAssignment DDL the role↔scope compatibility CHECK (§2-M1 text verbatim); add the write-path validation sentence; add the successor law: every B3–B5/R10-E permission check site declares Tenant-wide vs Area-targeted, with tenant-owner-only families being Tenant-wide checks.
2. **M2** — insert the canonical lock acquisition order + lock-mode law (§2-M2 block verbatim) into the B2-4 section as binding law.
3. **M3** — extend §4.5/S8 with the Group-deletion fail-closed law against live cross-owner references + the B4/B5 successor declaration obligation.
4. **L1** — §6.2: "revoke all ApplicationSessions" (delete "delete"); one sentence: row erasure is privacy-lifecycle only.
5. **L2** — one paragraph: initial `tenant_owner` grant seeding and admin-lockout recovery are non-serving maintenance-surface operations (B1 §6.7); R10-E may add a last-admin/self-offboard guard as UX hardening.
6. **L3** — proof-obligation list gains: static-catalog ↔ DB-CHECK parity guard must demonstrably fire.
7. **L4** — promotion restates the complete 5×43 role→permission bundle matrix in durable authority.
8. **L5** — one sentence on Group/Tenant/Area name normalization as implementation-spec detail.
9. **Editorial (D15)** — record the structural UUID rationale (no NULL-free natural PK) beside the GroupMembership pair-PK ruling.

---

# 12. Reopen determinations

**B2-1 reopen: NONE.** The integrated candidate consumes promoted B2-1 without contradiction; binding retention at offboarding, re-enable serialization, session terminality and the six reconciliation outcomes are all honored verbatim.

**Reopen outside B2: NONE required.** Successor obligations (not reopens) surfaced:

```text
B4   — declare Group-referencing persistent families (ApprovalPolicy actor_rule) + FK RESTRICT (M3);
       consume Area-retirement law for new policy references
B5   — same declaration for any Documentary-Context group/audience reference (M3)
B6   — field-by-field surviving Audit skeleton proof (already routed); Audit grant/revocation event
       content must carry subject/role/scope facts sufficient for forensic reconstruction
R10-C — restore non-resurrection proof (already routed)
R10-E — Session TTL value; per-check-site scope classification consumption (M1); optional last-admin
       guard; optional session/group display telemetry with named consumers
R10-F — maintenance-surface bootstrap seeding of the first tenant_owner grant (L2); legacy IAM
       deletion map (both grant tables, MFA/telemetry columns, sentinel tenant defaults, RLS surfaces)
```

**No material contradiction with any frozen R3–R9.5 decision was found.** One frozen-text tension was examined and resolved: ledger §5 "AuditEvent is transversal timeline only" vs grant history living only in Audit — resolved because no domain authority for grant history is *required to exist* (no consumer), so nothing semantic was moved into Audit (§5.12).

---

# 13. Closure

**Another broad review required: NO.** Findings converged in one round to three bounded, altitude-mechanical fixes inside a confirmed structure; the four Method passes and 18 scenarios produced no structural counterexample.

**Bounded delta review after corrections: SUFFICIENT.** The delta reviewer must verify: M1 CHECK text + matrix unchanged; M2 law recorded verbatim and consistent with C1–C7; M3 law + successor declarations; L1–L5 sentences; D15 rationale. Nothing else reopens.

**Promotable after operator adjudication: YES** — R10-B2 (B2-2 + B2-3 + B2-4 semantics as integrated here) is promotable once the §11 deltas are adjudicated into the corrected target and the bounded delta review closes green.

```text
VERDICT: APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
         WITH MATERIAL FIXES
BLOCKER = 0
MAJOR   = 3  (M1, M2, M3)
LOW     = 5  (L1–L5)
static Role/Permission catalog        SURVIVED
RoleAssignment XOR shape + UUID       SURVIVED
destructive offboarding               SURVIVED
Area reversible retirement            SURVIVED
deadlock/concurrency                  CORRECTION REQUIRED (M2 — order codified, then clean)
B2-1 reopen                           NONE
broad re-review                       NOT REQUIRED
bounded delta after corrections       SUFFICIENT
promotable after adjudication         YES
```
