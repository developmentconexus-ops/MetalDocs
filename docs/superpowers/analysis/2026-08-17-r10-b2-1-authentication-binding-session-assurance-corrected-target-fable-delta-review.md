# MetalDocs R10-B2-1 — Authentication Binding / Session / Assurance — Corrected Target — Independent Bounded Delta Review

> **Status:** INDEPENDENT BOUNDED DELTA REVIEW — EVIDENCE ONLY — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** Claude Fable 5 (cold bounded pass; repository truth at `ee0a0ce0`; no prior-conversation authority used)
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Review target:** `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-adjudicated-corrected-target.md` @ `ee0a0ce0`
> **Prior independent review:** `...-independent-fable-review.md` @ `361f6c8b` — `APPROVE WITH MATERIAL FIXES` (0 BLOCKER / 3 MAJOR / 5 LOW)
> **Original candidate:** `...-fable-review-request.md` @ `9cba3acd`
> **Authority baseline:** `e1bd83ce9a0a9b70135e4bd6984d990a54ba6377`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** delta only — F1/F2/F3 closures, the new reversible-`disabled_at` representation, and any new contradiction the corrections introduce. Whole-B2/whole-platform review deliberately NOT restarted.
> **Authority note:** findings are evidence for operator adjudication; nothing here amends authority.

Delta verified: `git log 361f6c8b..ee0a0ce0` contains exactly one commit adding the corrected-target
artifact; the authority chain (AGENTS.md, Method, handoff, program page, ledger, R10 page) is unchanged
against the state already verified by the prior review. All non-delta conclusions of the prior review
therefore carry over and were only re-touched where a correction could disturb them.

---

# 1. VERDICT

```text
APPROVE R10-B2-1 ADJUDICATED CORRECTED TARGET
```

```text
BLOCKER = 0
MAJOR   = 0
LOW     = 3   (DL1–DL3 — successor-obligation notes; none requires amending the corrected target before promotion)

prior findings closed = 8/8  (F1, F2, F3, L1–L5)
new material contradiction = NONE
new concurrency counterexample = NONE
```

---

# 2. F1 closure — total `UNIQUE(issuer,subject)` + reversible `disabled_at`

## 2.1 Required explicit sub-verdict

```text
TOTAL UNIQUE(issuer,subject) + REVERSIBLE disabled_at = APPROVE
```

The operator's modified fix was attacked on its own merits, not against the previously proposed
ACTIVE-scoped alternative. Conclusion: the modified fix **eliminates the original structural dead-end and
is stronger than the fix the prior review proposed.** The prior review's proposal made "one User per
subject at a time" a procedural property while leaving cross-User handover representable; the corrected
target makes "one User per subject *while the correlation record exists*" a structural DB guarantee. For
a governed attestation system, subject↔User attribution stability is the more valuable property, and the
prior review's "confirmed account handover" scenario is correctly reclassified as reopen-tier: deliberate
re-pointing of one provider account to a different human is an identity-hygiene anti-pattern that a QMS
should refuse by default, not a normal path. The reopen trigger in corrected §5.4/§20 names it precisely.

## 2.2 Dead-end elimination — verified case by case

| Scenario | Under corrected target | Dead-end? |
|---|---|---|
| Security unlink → relink (same User, same subject) | clear `disabled_at` on same row; mapping untouched; Audit records both transitions | No |
| Accidental disable → re-enable | same | No |
| Provider/realm migration (issuer changes) | new `(issuer,subject)` → new row; §11 replacement tx disables old + enables new + revokes old-bound Sessions atomically; total unique unaffected (different issuer) | No |
| Identity replacement (same User, new subject) | new row, different subject; old row retained disabled | No |
| Rollback of a replacement (return to old subject) | disable new row, re-enable old row — both same-mapping operations | No |
| Upstream federation change behind Keycloak | MetalDocs-facing issuer+subject unchanged; no binding operation at all | No |
| Historical Migration / R10-F cutover | fresh provider subjects → fresh rows; no reuse pressure | No |
| Same subject → different User while row retained | structurally impossible — **deliberate**, reopen-gated (§20) | By design |
| Same subject → different User after lawful erasure of the old row | possible as a NEW §6-governed correlation decision (see §8/DL1) | No |

The original defect (relink structurally impossible without violating uniqueness, immutability, or the
one-way lifecycle) is closed: reversibility removes the one-way lifecycle from the contradiction set, and
every legitimate V1 path lands on either same-row re-enable or new-row-new-subject.

## 2.3 Is total uniqueness safe against provider subject semantics?

External claim checked: OIDC Core requires `sub` to be locally unique and **never reassigned** within an
issuer; Keycloak's `sub` is the immutable provider-user UUID, and delete+recreate yields a new UUID. So
"same `(issuer,subject)`, different human" is provider misbehavior, not an expected class — and if a
misbehaving/rebuilt provider ever did reassign a subject, total uniqueness plus immutable mapping **fails
closed in the safest available direction** (the recycled subject cannot silently acquire a second User;
it resolves to the original correlation, and fixing it requires explicit disable/erasure decisions).
Partial-scoped uniqueness would share the same provider-trust exposure while guaranteeing less. No
accepted V1 requirement needs the same stable subject to become a different MetalDocs User; none of the
frozen flows (login, reauth, offboarding, replacement, reconciliation, migration) requires it.

## 2.4 Historical-truth model after reversibility

Clearing `disabled_at` makes the row **current-acceptance-only** state; disable/re-enable *history* moves
wholly to Audit (append-only, tamper-evident — frozen authority). This is cleaner than the original
candidate, where retained replacement rows doubled as history: now exactly one authority per meaning
(row = current acceptance; Audit = transition history). No V1 consumer needs point-in-time binding-state
reconstruction outside Audit forensics: Approval snapshots its own decision evidence, and Sessions bound
to a binding are revoked in the disabling transaction, so no downstream fact depends on "was the binding
enabled at time T" being answerable from the row. Sufficient.

## 2.5 Concurrency of the reversible representation (the new attack surface)

`UNIQUE(user_id) WHERE disabled_at IS NULL` is a partial unique **index** — PostgreSQL enforces it at
write time under every isolation level, independent of application locking. Attacked orderings:

1. **Replacement tx (disable old + enable/create new) vs concurrent re-enable of old:** both mutate the
   old row → row-lock serialization; the loser either observes the new state or attempts a second enabled
   row for the User and hits the partial unique index → ERROR, fail closed.
2. **Two concurrent re-enables of two different disabled rows (same User):** second UPDATE's index
   insertion waits on the first in-flight transaction, then fails on the unique index → fail closed.
3. **Login issuance vs disable/replacement (C3):** issuance validates acceptance against the binding row
   inside its serialization boundary (row share-lock or equivalent); disable takes the exclusive lock and
   revokes bound Sessions in the same tx. Whichever commits second observes the other (READ COMMITTED
   lock-requeue re-evaluation). Identical pattern to C1; cheap; no isolation escalation.
4. **Login vs re-enable:** both orders yield a legal state (refused login or valid Session on an enabled
   binding).
5. **Session on a re-enabled binding:** re-enable never resurrects Sessions (revocation terminal, §7.6);
   it only permits new issuance. No orphan acceptance.

**No sequence of disable / create-replacement / re-enable / login can commit two concurrently enabled
bindings for one User or leave a valid Session attached to a non-accepted binding**, provided B2-4
implements C3's declared serialization — and the data shape makes that enforcement trivially possible
(one row to lock, one partial index as DB backstop). The DB-backstop invariant (one enabled binding/User)
holds even if application locking is buggy, which satisfies the DB-enforces-invariants law.

**DL2 (LOW, successor note for B2-4):** re-enable must be explicitly treated as an *acceptance mutation*
under the same C3 lock discipline as disable/replacement (it is implied by "accepted binding state" but
worth naming, since it is the one mutation the original candidate did not have).

**F1 = CLOSED.**

---

# 3. F2 closure — subject correlation authority

Corrected §6 was attacked for residual attribute-authority paths:

- **Email via reconciliation:** the permitted path 2 ("unique provider-side correlation/idempotency
  mechanism belonging to that exact intent") cannot be satisfied by email/username uniqueness: those are
  human identity attributes explicitly enumerated as never-authority, and "belonging to that exact
  intent" requires the correlation marker to be produced by/for the intent execution itself, not to
  pre-exist as a human attribute. The forbidden list ("already exists + matching attribute") closes the
  conflict-adoption hole found in the original review. No path found where an attribute silently selects
  a subject.
- **Precision vs over-specification:** path 2 states the property (causal, unique, intent-bound) without
  freezing R10-D mechanics (provider attribute stamped at create, idempotency key, search-by-marker on
  retry — all remain legal realizations). Correct altitude.
- **Create-timeout reconciliation:** timeout → uncertain → reconciler searches for the intent's own
  correlation marker; found ⇒ causally proven subject ⇒ bind; not found ⇒ retry create. Safe without
  attribute matching, without a third semantic family.
- **Conflict/already-exists:** no automatic binding; pending explicit correlation (path 3). Adoption of a
  pre-existing provider account is possible **only** through an explicit trusted human decision. Correct.
- **Third semantic state?** The pending decision lives as R10-D mechanism/intent state plus an eventual
  audited human decision that produces the binding. No semantic table is needed and none was added.

**DL3 (LOW, optional tightening — no promotion blocker):** if a future editor touches §6, path 2 could
say "a correlation marker created by the execution of that exact intent" to make the already-implied
exclusion of pre-existing attributes textually self-evident. The current text is not exploitable when
read with its forbidden list.

**F2 = CLOSED.**

---

# 4. F3 closure — fresh-auth satisfaction contract

Corrected §9 was attacked for surviving "sudo forever" and for hidden state needs:

- **Sudo-forever:** dead. Bare non-NULL is explicitly forbidden as satisfaction; consumers must apply
  one-shot or a deliberately configured window; no implicit/unbounded window exists; initial login does
  not satisfy. The degenerate reading the original review exposed is now unwritable.
- **Does operation binding need to persist in Authentication?** No. One-shot linkage is realizable as
  (a) flow/mechanism state bridging the redirect (explicitly routed to journey design, §9.4) plus
  (b) the consumer's own decision record as the durable consumption evidence — ApprovalDecision already
  snapshots the assurance it consumed (frozen ledger attestation law). A dedicated one-shot semantic
  table in Authentication would duplicate the consumer's evidence authority; correctly absent.
- **Two concurrent operations consuming the same evidence:** if the consumer's policy is a freshness
  window, double-consumption inside the window is legal by definition. If the policy is one-shot,
  single-consumption enforcement belongs to the consuming policy authority (B4 for Approval), with B2-4
  supplying ordinary transactional means. B2-1 correctly owns neither — its law ("never bare non-NULL")
  is consumer-agnostic. Ownership routing in §18 is correct.
- **Are `latest_*` fields still needed?** Yes: option-B (window) policies read persisted state across the
  callback→action request gap, and `auth_time/acr/amr` are the frozen provider-contract facts consumers
  snapshot. No field deletable (see §9 below); `FreshAuthEvidence` remains correctly transient.

**F3 = CLOSED.**

---

# 5. ApplicationSession — re-verification under the delta

Nothing in F1/F2/F3 closures disturbs the Session verdicts; re-attacked briefly:

- The provider-JWT-plus-minimal-revocation alternative was re-run against the corrected target: unchanged
  defeat. The "minimum revocation machinery" is a per-request-consulted server-side record keyed by
  session identity — i.e., the Session table re-created minus its benefits, plus provider-lifetime
  coupling, refresh-token custody, and a claims-bearing artifact past the login boundary. Reversible
  bindings change nothing in this trade. **Local opaque ApplicationSession remains the Global Maximum.**
- Field omissions re-verified: no `user_id` duplicate (resolution chain single-authority; §8's runtime
  context `user_id` is resolved, not persisted — consistent), no tenant/AuthZ snapshot (paired invariant
  with C4 explicitly recorded), no provider tokens (transient `id_token_hint` mechanism door correctly
  bounded), no IP/UA/LastSeen semantics (support-state door correctly bounded), no idle timeout (absolute
  TTL suffices; L3 routing of the staleness consequence to the TTL chooser is present in §7.5), multiple
  Sessions/User, finite expiry, digest-only storage with row-disclosure non-replayability (live-proven
  property). All hold.

---

# 6. Provider disable / offboarding — re-verification

Bounded staleness is now an **explicit bounded property**: staleness ≤ remaining lifetime of the affected
Session ≤ configured finite absolute TTL, with reconciliation as an earlier optional bound and local
revocation always available. The corrected target states the consequence at the point where the TTL value
will be chosen (§7.5) — exactly what the prior L3 required. Offboarding remains authoritative, local,
one-coherent-transaction (B2-4-finalized), provider-independent — it never inherits provider staleness.
No TTL number selected, correctly. No change to the prior ACCEPT.

---

# 7. Provider reconciliation — six cases under reversible `disabled_at`

Reversibility changes case 3 for the better (provider account disabled → binding disabled + Sessions
revoked; provider account later restored → same-row re-enable instead of the old model's blocked new-row
path) and leaves the rest intact. A retried intent observing an existing row with the **same** mapping is
an idempotent no-op, distinct from case 4's conflicting-duplicate rejection — both behaviors fall out of
total uniqueness naturally. No case requires a `provider_sync` semantic FSM; attempt/retry/lease/error
state remains R10-D mechanism state; uncertain provider truth never becomes binding truth (§12.6 + §6);
every provider interaction remains outside local DB atomicity (§17 orderings re-walked — no hidden XA).

---

# 8. Privacy — including post-erasure subject reuse

- Erasure order (`ApplicationSession` → `ProviderSubjectBinding`) matches the RESTRICT reference law;
  Audit's surviving skeleton is PII-minimized and has no FK dependence on Authentication rows; governed
  evidence references the MetalDocs User, never the provider subject — so binding/session erasure cannot
  orphan or rewrite governed evidence.
- Reversible history does **not** become a retention system — the opposite: same-row reversibility
  retains *fewer* rows than the original new-row-per-relink model; transition history lives in Audit,
  which already has its own retention regime.
- **Total uniqueness does not force permanent retention.** The structural no-handover guarantee is scoped
  to *retained* rows. After lawful erasure of a binding row, a later reappearance of the same
  `(issuer,subject)` can safely create a **new** row — this is not a "handover" but a new correlation
  decision, still gated by §6 (causal correlation or explicit trusted human decision; never attribute
  matching), and historical attribution of pre-erasure governed evidence is unaffected because that
  evidence references the User, not the subject. What lawful erasure genuinely surrenders is the
  structural DB guarantee against re-correlation of that subject — deliberate and inherent to erasure,
  with Audit's PII-minimized skeleton (per B6's field-by-field proof) as the remaining historical record.

**DL1 (LOW, successor note for B2-4/B6):** record this post-erasure semantics explicitly in the B2-4
persistence/privacy classification (one sentence: "erasing a binding row lawfully surrenders the
structural no-re-correlation guarantee for that subject; any later re-binding is a new §6-governed
decision"). No target amendment required now; no contradiction exists.

---

# 9. Subtractive pass — "what can still be deleted from the corrected target?"

| Element | Deletable? | Consuming property |
|---|---|---|
| `disabled_at` | No | acceptance gate; partial-unique predicate; the F1 relink fix itself |
| Binding `created_at` | No (weakest keeper) | correlation-establishment instant on the row itself; cheap; forensic ordering without Audit join |
| `latest_reauthenticated_at` | No | provider-independent local verification anchor for window policies (F3 option B) |
| `latest_provider_auth_time` / `latest_acr` / `latest_amr` | No | frozen provider-contract facts; Approval snapshot inputs across the callback→action gap |
| `UNIQUE(credential_digest)` | No | lookup index + collision/duplicate backstop in one constraint |
| Session `id` | No | stable session identity independent of credential material; FK/evidence target (`FreshAuthEvidence.session_id`); digest-as-PK would make security material a key |
| Binding `id` | No | Session FK target; B1 forbids external identifiers as technical PKs |
| `state`/`state_changed_at` | Already deleted | L1 adopted — enum gone |
| Any lifecycle rule | No | terminal revocation → C2; same-tx revoke → C3/D18; finite expiry → staleness bound; reversibility → F1 |
| Any reconciliation rule | No | each of the six maps to a distinct reachable failure class |

Nothing further is deletable without losing a named property (**P31 confirmed**). The corrected target is
at its subtractive floor.

---

# 10. Proof-obligation verification — corrected target §19 P1–P31

| P | Result | P | Result | P | Result |
|---|---|---|---|---|---|
| P1 | HOLDS | P12 | HOLDS (live-proven pattern) | P22 | HOLDS (row-lock realization, L5/§13) |
| P2 | HOLDS | P13 | HOLDS | P23 | HOLDS (conditional final update) |
| P3 | **HOLDS — no internal contradiction** (§2.2; reversibility is a declared mutation law of a non-mapping field) | P14 | HOLDS | P24 | HOLDS (§2.5 — incl. re-enable races) |
| P4 | **HOLDS under concurrency** (partial unique index enforced at write time; §2.5 cases 1–2) | P15 | HOLDS | P25 | HOLDS (paired invariant §8) |
| P5 | HOLDS (same-row re-enable; mapping untouched) | P16 | HOLDS | P26 | HOLDS (§7) |
| P6 | HOLDS (total unique + immutable mapping) | P17 | HOLDS | P27 | HOLDS |
| P7 | HOLDS (§3) | P18 | HOLDS | P28 | HOLDS |
| P8 | HOLDS (§3 conflict path) | P19 | HOLDS | P29 | HOLDS (§8) |
| P9 | HOLDS | P20 | HOLDS (TTL-bounded, explicit) | P30 | HOLDS (§17 re-walked) |
| P10 | HOLDS | P21 | HOLDS | P31 | HOLDS (§9) |
| P11 | HOLDS | | | | |

---

# 11. Required outputs

- **Verdict:** `APPROVE R10-B2-1 ADJUDICATED CORRECTED TARGET`.
- **Counts:** BLOCKER 0 / MAJOR 0 / LOW 3 (DL1 post-erasure semantics note → B2-4/B6; DL2 re-enable
  under C3 lock discipline → B2-4; DL3 optional §6 path-2 wording tightening → only if the text is
  edited again anyway).
- **F1 closure:** CLOSED — modified fix approved; explicit sub-verdict `TOTAL UNIQUE(issuer,subject) +
  REVERSIBLE disabled_at = APPROVE`; stronger than the reviewer's original proposal.
- **F2 closure:** CLOSED — no attribute-authority path survives; correct altitude vs R10-D.
- **F3 closure:** CLOSED — degenerate "non-NULL = fresh" reading unwritable; ownership routing correct.
- **Total issuer+subject uniqueness:** SURVIVED (approved as structural attribution guarantee).
- **Reversible `disabled_at`:** SURVIVED (eliminates dead-end; concurrency-safe with DB backstop).
- **One-enabled-binding/User:** SURVIVED (partial unique index enforces under all interleavings).
- **Local ApplicationSession:** SURVIVED (stateless alternative re-attacked and re-defeated).
- **Assurance model:** SURVIVED (bounded latest inputs + transient evidence + consumer-owned bounded
  satisfaction; no third table).
- **Deletable fields:** NONE (subtractive floor reached; binding `created_at` is the weakest keeper and
  still consumed).
- **New concurrency counterexample:** NONE (including the new disable/replace/re-enable/login surface).
- **Privacy consequences:** erasure order sound; no retention coercion; post-erasure subject reuse is a
  new §6-governed decision (DL1 note).
- **Material reopen outside B2-1:** NONE. No finding crosses the Authentication boundary; B2-2/B2-3/B2-4
  remain unpromoted; the successor obligations recorded here (DL1, DL2, C1/C3 enforcement, B4 freshness
  policy) are already routed by the corrected target's §18.
- **Exact promotion conditions:** none outstanding — the corrected target is promotable **as written**.
  DL1–DL3 are successor-stage notes, not pre-promotion amendments. Operator promotion of B2-1 to target
  authority is the next gate.
- **Another broad review:** NOT REQUIRED. The delta introduced no new authority, no topology movement,
  and no contradiction with frozen R9.5/GCR/single-company/R10-A/B1 authority.

**Convergence:** count 0 BLOCKER / 0 MAJOR; altitude fell from constraint-level defects to
successor-stage notes; stop condition met. Findings remain evidence; operator adjudication remains the
authority gate.
