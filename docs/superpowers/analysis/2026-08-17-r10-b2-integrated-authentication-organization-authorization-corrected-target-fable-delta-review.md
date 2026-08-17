# MetalDocs R10-B2 — Integrated Adjudicated Corrected Target — Bounded Delta Review

> **Status:** BOUNDED DELTA REVIEW — EVIDENCE ONLY / **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed corrected target:** `docs/superpowers/analysis/2026-08-17-r10-b2-integrated-authentication-organization-authorization-adjudicated-corrected-target.md` @ `2908a884c49f6cc81a637384b3646914ea556b1f`
> **Prior full independent review:** `...-independent-fable-review.md` @ `34a567fd` — APPROVE WITH MATERIAL FIXES (0/3/5)
> **Integrated candidate:** `...-fable-review-request.md` @ `b814f672`
> **Promoted architecture baseline:** `71791dfecd4cd185684373ffcdccbf256138b741`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — this review changes nothing.**

Scope discipline: this is one bounded delta review of the adjudicated corrections (M1–M3, L1–L5, D15, A1) and any contradiction they introduce. No broad B2 re-review, no B2-2/B2-3/B2-4 microreviews, no S1–S18 re-run except where a delta directly touches a scenario.

---

# 1. VERDICT

```text
APPROVE R10-B2 INTEGRATED ADJUDICATED CORRECTED TARGET

BLOCKER = 0
MAJOR   = 0
LOW     = 2   (N1, N2 — non-blocking notes; neither conditions promotion)

prior findings closed          = 3/3 MAJOR + 5/5 LOW + D15 note
A1 tenant-owner-only access administration = APPROVE
exact 5×43 bundle verification = MATCH (mechanically verified against archived R9 authority)
deadlock under corrected law   = NONE
new material contradiction     = NONE
B2-1 reopen                    = NO
reopen outside B2              = NO
broad review required          = NO
promotable after operator promotion gate = YES — entire R10-B2 batch
```

---

# 2. Delta 1 — M1 role↔scope CHECK — **CLOSED**

**Q1 — does the CHECK make `tenant_owner@AreaScope` and `area_manager@TenantScope` structurally impossible?** Yes. Corrected §5.3 CHECK, evaluated with the scope XOR (§5.1, exactly one scope column non-NULL):

- `tenant_owner` + `area_scope_id` set → XOR forces `tenant_scope_id` NULL → branch 1 false; branch 2 requires `area_manager`; branch 3 excludes `tenant_owner` → **row rejected**.
- `area_manager` + `tenant_scope_id` set → symmetric → **rejected**.
- The three unrestricted roles pass through branch 3 with either (exactly one) scope. Truth table over all 5 roles × 2 scope kinds reproduces the accepted matrix exactly.

**Q2 — NULL/XOR escape?** None. Every predicate in the three CHECKs is built from `IS NOT NULL` tests and an `IN` list over a `NOT NULL` column — none can evaluate to SQL `NULL`, so the CHECK-passes-on-UNKNOWN loophole is unreachable. Both-NULL and both-set scope states are independently rejected by the scope XOR before the matrix CHECK is even relevant. Subject XOR is orthogonal and unaffected.

**Q3 — authority duplication?** No. The matrix has one semantic home (static catalog §4.2/§5.3); the CHECK is enforcement vocabulary, exactly the B1 TEXT+CHECK pattern, and L3's parity proof (§4.3) mechanically pins catalog == role CHECK == role↔scope CHECK with drift failing verification. Enforcement, not a second catalog.

**Q4 — matrix still coherent with frozen bundles?** Yes, and more strongly than before: with the exact bundles recovered, `tenant_owner` is the only whole-company-administration bundle (TenantScope-only is forced) and archived R9-04 defines `area_manager` as operational with an explicit administration negative (§3 below) — AreaScope-only is the only coherent reading. The three operational roles at both scopes remain legitimate.

**Q5 — can a future check site still degrade a whole-company permission into an Area permission?** The residual risk is real and correctly handled: §5.3/§14 bind every B3–B5/R10-E check site to declare Tenant-wide vs Area-targeted, with tenant-owner-only whole-company families fixed as Tenant-wide. Because the CHECK already makes `tenant_owner@AreaScope` unrepresentable, a mis-declared Area-targeted check for a whole-company permission could only ever be satisfied by a TenantScope `tenant_owner` grant anyway — the M1 escalation cannot be reconstructed through check-site error. Defense-in-depth is correctly layered: structural CHECK (cannot hold the role at the wrong scope) + declaration law (cannot evaluate at the wrong altitude).

M1 **CLOSED**. The prior review's escalation example (`tenant_owner@AreaScope` self-grant) is now doubly dead: A1 removed the request-path actor (no AreaScope `access.manage` holder exists), and the CHECK removes the state itself. The corrected target's reclassification of the threat from request-path escalation to structural-invalid-state backstop (§1.1) is accurate, and retaining the CHECK despite A1 is the right call — bugs, migrations and maintenance paths are exactly what schema backstops exist for.

---

# 3. Delta 2 — A1 tenant-owner-only access administration — **APPROVE**

Verified against archived frozen authority, not the corrected target's own claims.

**Q1 — does the 5×43 matrix confirm only `tenant_owner` holds `access.manage`?** Yes, from the archived R9 ledger (`git show cea29ae7`, section R9-04 — Role bundles):

- `viewer` = `document.read_effective` only; `author` = viewer + 7 document permissions; `approver` = viewer + `approval.act`; `area_manager` = author + 9 (document.cancel_revision/obsolete/owner.manage, 4 approval.*, 2 distribution.*), closing with the explicit sentence: **"No tenant IAM/config/audit/session/lifecycle administration."**
- `tenant_owner` = "All 29 tenant Permissions via ordinary Authorizer. Still no bypass."

`access.manage` appears in no bundle except tenant_owner's all-29. R9.5's 16 additions (current ledger §1) add no administration permission to any role. Confirmed.

**Q2 — any frozen requirement for area_manager to administer RoleAssignments?** No — the opposite: the archived bundle text is an explicit negative. There is no frozen sentence anywhere in R1–R9.5 granting Area-local RBAC administration to anyone.

**Q3 — operational manager or RBAC manager?** Operational, by frozen text: the area_manager bundle is entirely document/approval/distribution/evidence/dossier operations. RBAC/config/audit/session administration is bundle-excluded.

**Q4 — losing a needed V1 capability?** No. The original candidate's `AreaScope(A) access.manage → manage AreaScope(A) grants` path was **structurally unreachable** in V1: the only role carrying `access.manage` is `tenant_owner`, and the M1 CHECK pins `tenant_owner` to TenantScope — so no principal could ever hold `access.manage` effective at only an AreaScope. Removing the sentence deletes a dead path, not a capability. This is a correct subtractive refinement, and note R9-03 (archived): self-session listing/revocation is relationship/self authority, not `session.manage` — so tenant-owner-only administration does not strip ordinary users of self-service session hygiene.

**Q5 — any Approval/Document/Distribution flow needing Area-local grant administration?** No. Approval actor resolution (`NamedUser | Group | RoleInArea`) *reads* current grants; it never creates them. Distribution snapshots audiences. Delegation of Area work is expressed by the Tenant owner granting AreaScope roles to Users/Groups — grant *use* is area-local, grant *administration* is not, and nothing frozen requires the latter.

**Q6 — does this reduce privilege-escalation surface?** Yes, materially. The entire S16 attack class (area-tier actor manipulating access configuration) loses its actor: no non-tenant-owner principal can mutate GroupMembership or RoleAssignment through the request path at all. What remains is single-tier and fully Audit-covered.

**Required sub-verdict: TENANT-OWNER-ONLY ACCESS ADMINISTRATION V1 = APPROVE.**

**Consistency with the prior full review (honest disclosure).** The prior review's §5.7 rationale contained the sentence "Area-local delegation remains available through direct per-User AreaScope grants, which stay entirely inside the Area admin's authority" — written under the then-unverified assumption that AreaScope `access.manage` holders could exist. A1 supersedes that assumption. No disposition changes: D22 (TenantScope-only membership administration) now holds *a fortiori*; D23's law ("access.manage effective at the target scope") is formally unchanged — TenantScope `access.manage` is effective at every scope under §6.2 — the holder set is simply smaller. This is a rationale correction, not a material contradiction, and the corrected target discloses it properly (§1.1) rather than promoting it silently.

---

# 4. Delta 3 — exact 5×43 matrix — **MATCH**

Mechanical verification, permission by permission, against three sources: archived R9-02 base catalog + R9-04 base bundles (`cea29ae7`), the single-company removal (current ledger §1: exactly `tenant.export` + `tenant.deletion.request` removed), and the locked R9.5 additions (current ledger §1).

**Catalog arithmetic.** R9-02 lists 29 (7 config + 11 document + 4 approval + 2 distribution + 5 compliance incl. the two later-removed). 29 − 2 = 27 base; + 16 R9.5 = 43. The corrected target's tenant_owner list enumerates exactly those 43; name-by-name comparison against the current ledger's 27+16 lists: **no missing, no extra, no renamed permission; neither removed permission reappears; no invented permission.**

**Per-bundle verification:**

| Role | Base (R9-04) | + R9.5 (ledger) | Expected | Corrected target | Verdict |
|---|---|---|---|---|---|
| viewer | 1 (`document.read_effective`) | +2 (`evidence.read`, `dossier.read`) | 3 | 3, names exact | **MATCH** |
| author | 8 (viewer + read_history/read_working/create/edit/comment/submit/review_periodic) | +7 (evidence.read/create/edit/capture, dossier.read/create/manage) | 15 | 15, names exact | **MATCH** |
| approver | 2 (viewer + `approval.act`) | +2 (`evidence.read`, `dossier.read`) | 4 | 4, names exact | **MATCH** |
| area_manager | 17 (author + cancel_revision/obsolete/owner.manage + 4 approval.* + 2 distribution.*) | +8 (all author additions + `evidence.void`) | 25 | 25, names exact | **MATCH** |
| tenant_owner | 27 (all base after removal) | +16 (all) | 43 | 43, names exact | **MATCH** |

**Specific proofs demanded:**

- `area_manager` has no `access.manage` — confirmed (not in author base, not in the 9-permission delta, excluded by the archived negative sentence).
- `area_manager` has no `organization.manage` — confirmed, same grounds.
- `approver` has no working/history blanket — confirmed: bundle is exactly {read_effective, approval.act, evidence.read, dossier.read}; the archived "Approver gets no blanket working/history access; Approval participation opens the exact case/Submission" sentence is carried verbatim into the corrected target (§4.2).
- `tenant_owner` = exactly all 43 — confirmed by enumeration.
- `tenant.export` / `tenant.deletion.request` absent everywhere — confirmed.
- No new permission invented — confirmed (bijection with ledger catalog).
- No role missing a frozen R9.5 addition — confirmed against the ledger's per-role addition lists, including `area_manager` receiving all author additions + `evidence.void` and `tenant_owner` receiving all 16.

**Result: EXACT MATCH — zero mismatches.** The provenance note ("29 originally, −2 single-company, +16 R9.5") is historically accurate.

---

# 5. Delta 4 — M2 lock order / deadlock — **CLOSED, NONE**

The corrected §8.1/§8.2 law is the prior review's M2 fix recorded faithfully: class order User ≺ Binding(asc id) ≺ Area ≺ child sets (Session asc id, GroupMembership asc (user_id,group_id), RoleAssignment asc id); Group deletion as an isolated class that never subsequently acquires User/Binding/Area; FOR SHARE for eligibility/acceptance readers, FOR UPDATE for lifecycle mutators; FOR KEY SHARE explicitly rejected (non-key `disabled_at` updates take FOR NO KEY UPDATE, which FOR KEY SHARE does not conflict with — the C1/C5 serialization would silently vanish).

Pairwise wait-for re-walk under the codified law:

- **Issuance vs offboarding** — both start at class 1 (User S vs X); direct conflict, total order, C1 outcomes only.
- **Binding disable vs offboarding** — disable holds Binding(2), proceeds to Sessions(4, asc id); offboarding holds User(1), then wants Binding(2). Disable never wants class 1 → cannot wait on offboarding → completes; offboarding then proceeds. Both order session updates asc id, so even the multi-row overlap (the prior review's concrete counterexample) is deadlock-free. No cycle.
- **Membership add / direct grant vs offboarding** — class-1 S vs X conflict; C3/C4 total order.
- **Area grant vs retirement** — grant: User(1,S) → Area(3,S); retirement: Area(3,X) only. Retirement holds nothing the grant holds beyond Area itself and wants nothing further → completes; grant re-reads `disabled_at` post-lock → C5 outcomes only.
- **Group delete vs membership add** — delete: Group FOR UPDATE → memberships asc user_id; add: User(1,S) → INSERT (FK takes KEY SHARE on the Group row, conflicting with FOR UPDATE). Add waits on the Group row; delete never waits on User or on the add's uncommitted row → delete commits → add's FK fails → fail closed. No cycle.
- **Group delete vs Group RoleAssignment mutation** — group-subject RA INSERT's FK likewise takes KEY SHARE on the Group row and serializes against the deletion's FOR UPDATE; RA revoke is a single-row delete that holds nothing else. No cycle.

Exhaustive composition over the operation set yields a wait-for graph that respects one partial order with the Group class disconnected from classes 1–3: **DEADLOCK CYCLE UNDER CORRECTED LAW = NONE.**

**Overfreezing check:** the law fixes only lock classes, child-row ordering and mode selection — the minimum needed to make the no-cycle proof and the C1/C5 conflict guarantee reproducible. SQL syntax, index selection and query shapes remain free. Not overfrozen; the FOR-KEY-SHARE rejection in particular is load-bearing correctness, not incidental mechanics.

---

# 6. Delta 5 — M3 Group deletion — **CLOSED**

- **Hard delete vs `Group.disabled_at`:** hard delete remains correct. A retirement state would add lifecycle machinery with no consumer; a Group that cannot be deleted because it is still referenced is *supposed* to stay, and an emptied-but-referenced Group is harmless in B2 (zero members → zero conferred access; an ApprovalPolicy step resolving an empty Group to zero actors is B4's existing no-eligible-actor fail-closed concern, not a B2 state problem). YAGNI holds.
- **Typed FK RESTRICT sufficiency:** yes for everything persisted — and the corrected law closes the only loophole by construction: "every persisted live Group reference must be an ordinary typed FK to Group(id) with RESTRICT/NO ACTION" is stated as a binding design law on consumers (§3.5), so a B4 design that buried a Group id in a JSONB policy document would violate B2 law, not merely miss an FK. Dangling-reference paths: none representable.
- **Approval and Distribution both B4:** correct routing — B1 §6.11 assigns Approval + CI-owned Rendition/Release + Distribution to B4. Both named consumers (ApprovalPolicy Step actor_rule = Group; live Distribution group audience pre-snapshot) are listed with the FK law in §14/B4.
- **B5 obligation:** correctly conditional only — "no speculative Group requirement; same reference law only if a real B5 Group reference appears". No speculative requirement added. Correct.
- **Historical safety:** snapshots resolve concrete Users (unchanged frozen law), so deletion of an unreferenced Group breaks no history.

---

# 7. Delta 6 — offboarding / re-enable — **CLOSED**

- **L1:** §7.1 now reads "revoke all ApplicationSessions … terminal revoked_at mutation; DO NOT erase Session rows here". Revoke/erase separation explicit; physical erasure lives only in the privacy lifecycle (§11). **Closed.**
- **No stale privilege restoration:** re-enable clears `disabled_at` only; memberships/direct grants were deleted at offboard and nothing re-creates them; revoked Sessions are terminal per promoted B2-1. Default deny until fresh grants. Structurally sound.
- **Audit:** offboarding, re-enable, membership removal and grant revocation are all in the §10 same-commit list, and §10 adds the forensic field floor (assignment id, subject, role, scope, actor, operation, trusted time) so transition evidence survives current-row deletion. Sufficient.
- **Provider side:** durable intent in the local commit, execution strictly post-commit via R10-D; binding correlation retained. Consistent with promoted B2-1 §7.6/§7.8.
- **No accidental suspension promise:** §7.2 states no temporary-suspension state exists V1 and names intentional-restoration suspension as a reopen trigger (§16). Correctly fenced.

---

# 8. Delta 7 — L2 / L3 / L5 / D15 — **ALL CLOSED**

- **L2 (§12):** first `tenant_owner` seed + lockout recovery on the distinct non-serving maintenance trust surface; never request-reachable, never a bypass; R10-F owns the procedure; R10-E UX guard is defense-in-depth only. Matches B1 §6.7. **Closed.**
- **L3 (§4.3):** mechanical parity proof — static role codes == role_code CHECK == role↔scope CHECK, drift fails CI/migration verification; CHECK named enforcement, not second authority. **Closed.**
- **L5 (§13):** blank/trim/casing are implementation-spec; explicitly no case-insensitive `Group.name` uniqueness, `Tenant.display_name` not a routing key, `Area.name` not identity. **Closed.**
- **D15 (§5.6):** UUID rationale recorded correctly — XOR union has no NULL-free natural column set; PostgreSQL PKs cannot contain nullable columns; the four partial uniques prove duplicate rejection but cannot jointly be a PK. GroupMembership's NULL-free pair PK correctly remains the contrast case. **Closed.**

---

# 9. Delta 8 — successor routing — **CONSISTENT**

§14 routes exactly and only what the accepted design implies: B3–B5/R10-E per-check-site Tenant-wide/Area-targeted declaration (with tenant-owner-only families pinned Tenant-wide); B4 = retired-Area new-policy rejection + both Group FKs + concrete-User snapshots + bounded fresh-auth per promoted B2-1; B5 = conditional-only; B6 = Audit skeleton privacy classification + grant/revoke forensic fields; R10-C = restore non-resurrection; R10-D = durable provider execution; R10-E = journeys/TTL/scope-check consumption/neutral fallback/optional guards; R10-F = bootstrap seeding, lockout recovery, legacy 8-role + dual grant-table + tenant_id/RLS cutover, parity gate. No successor item is smuggled into current B2 scope; implementation remains blocked and §15 keeps all proofs future-tense.

---

# 10. Global delta check

1. **Second authority introduced?** No. Matrix, vocabulary and lock law each have one semantic home; CHECKs are enforcement under L3 parity.
2. **Tenant partitioning/RLS reintroduced?** No. Only `tenant_scope_id` on RoleAssignment, nullable, XOR-bound, semantic (§5.7).
3. **B2-1 reopened?** No. Both families, binding retention, session terminality, C2 discipline and reconciliation semantics consumed unchanged.
4. **Exact bundle recovery contradicts the prior integrated review?** No disposition or verdict is contradicted. One prior *rationale sentence* (assumed Area-admin delegation in §5.7 of the full review) is superseded by A1 — disclosed in §3 above; conclusions hold a fortiori.
5. **A1 removed a required capability?** No — it removed a structurally unreachable dead path; archived R9-04 is an explicit negative for Area-local administration; self-session management remains relationship authority (R9-03).
6. **B2 remains Global Maximum?** Yes. Every delta is enforcement, subtraction or evidence-pinning inside the confirmed structure; A1 makes the target strictly smaller.
7. **Anything else deletable?** Nothing found. Re-checked the A1 neighborhood: the §6.2 scope-application law and the four partial uniques all remain load-bearing; the target sits at its subtractive floor.
8. **Another broad review?** No. Findings are exhausted; altitude is zero (two non-blocking notes below).

## Non-blocking notes (LOW)

**N1 — Group-subject RoleAssignment inserts vs Group deletion rely on the FK KEY-SHARE conflict.** §8.1's note names concurrent *membership* creation as the FK-supplied race; group-subject grant inserts use the identical mechanism (FK on `group_id` vs the deletion's FOR UPDATE) but are not named. The proof in §5 above covers it; one clarifying clause at implementation-spec time suffices. Non-material.

**N2 — On promotion, name the durable R10 authority page as the single current-bundle home.** The full 5×43 matrix now exists in this corrected target and (as provenance deltas) in the ledger. When the operator promotes B2 into `wiki/architecture/r10-technical-architecture.md`, that page should carry the current matrix and the ledger remain provenance — one current-truth home, avoiding hand-synced drift between documents. Promotion mechanics, not a design defect.

---

# 11. Closure

```text
VERDICT: APPROVE R10-B2 INTEGRATED ADJUDICATED CORRECTED TARGET

BLOCKER = 0
MAJOR   = 0
LOW     = 2 (N1, N2 — notes only)

M1 role↔scope CHECK                      CLOSED (structurally exhaustive; no NULL escape)
A1 tenant-owner-only access admin        APPROVE (archived R9-04 explicit; dead path removed)
exact 5×43 bundle verification           MATCH — zero mismatches vs cea29ae7 R9-02/R9-04 + ledger
M2 lock order / deadlock                 CLOSED — no wait-for cycle under codified law
M3 Group deletion                        CLOSED — fail-closed reference law, B4 routed, B5 conditional
offboarding / re-enable                  CLOSED (L1 revoke/erase separation explicit)
L2 / L3 / L5 / D15                       CLOSED
new material contradiction               NONE
B2-1 reopen                              NO
reopen outside B2                        NO
Global Maximum                           CONFIRMED (target strictly smaller after A1)
another broad review                     NOT REQUIRED
```

**Remaining promotion conditions (exact):**

1. Operator promotes the entire R10-B2 batch into R10 authority: update `wiki/architecture/r10-technical-architecture.md` (carrying the corrected target's laws and the full 5×43 matrix as the single current-bundle home — N2), `wiki/references/current-agent-handoff.md`, and the program mirror; record the evidence chain (candidate `b814f672` → full review `34a567fd` → corrected target `2908a884` → this delta review).
2. Successor obligations (§14 of the corrected target) recorded in the promoted authority as B3–B6/R10-C/D/E/F gates.
3. Nothing else: no further review cycle is required before promotion; N1 is an implementation-spec clause, N2 is promotion mechanics.

**Entire R10-B2 is promotable: YES.** After promotion, the next stage is R10-B3. No product implementation is authorized by this review.
