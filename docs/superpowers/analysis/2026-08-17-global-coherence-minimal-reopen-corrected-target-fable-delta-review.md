# MetalDocs Global Coherence Review — Corrected Target / Independent Bounded Delta Review

> **Status:** INDEPENDENT BOUNDED DELTA REVIEW — **EVIDENCE ONLY, NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed HEAD:** `31470deb1f12f401d0483539b6e0e9442779097c`
> **Reviewed artifact:** `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-adjudicated-corrected-target.md`
> **Prior review of reference:** `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-independent-fable-review.md` @ `ccf97578`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** bounded delta only — closure of the 6 MAJOR / 5 LOW prior findings and the corrected-target §12 questions. No whole-platform re-review performed or required.
> **Implementation gate:** CLOSED — this review authorizes nothing.

Findings are evidence, not authority.

---

# 1. Verdict

```text
APPROVE GCR ADJUDICATED CORRECTED TARGET
```

```text
BLOCKER = 0
MAJOR   = 0
LOW     = 5 (promotion-precision notes; none blocks promotion)

prior findings closed        = 11/11 (M1–M6, L1–L5)
new material contradiction   = NONE
fifth material local maximum = NONE
promotion after this check   = MAY PROCEED — no further broad review required
```

The delta since `ccf97578` is exactly one artifact (the adjudicated corrected target); R9.5/R10-A/R10-B1/handoff authorities are untouched, matching the declared gate.

---

# 2. Closure audit of prior findings

| Prior | Adjudication | Closure verdict |
|---|---|---|
| M1 — R10-A Authentication amendment required | ACCEPT; §3.1 + §10 amend the owned-fact family (binding / app Session / assurance / anti-corruption contract in; credential storage/policy/activation/lockout out); topology stays 8+3 | **CLOSED** |
| M2 — no cross-DB atomicity; provisioning choreography | ACCEPT; §3.5 states the law, enumerates six reconciliation cases (user-no-subject, subject-no-binding, orphaned binding, duplicate `iss+sub`, provider down, uncertain-response retry), forbids XA/2PC, routes B2/R10-D | **CLOSED** — the case matrix is complete enough to be operable, not underspecified |
| M3 — structural anti-corruption contract | ACCEPT/STRENGTHEN; §3.3 enumerates the only crossing facts (`issuer, subject, authenticated_at, auth_time, acr?, amr?`), bans the generic claims map, bans provider role/group/org/permission consumption, bans mapping table and claim-to-permission bridge | **CLOSED** — structural, not disciplinary |
| M4 — audit-PII family + fail-closed + removal cascade | ACCEPT/ROOT-CAUSE RESTRUCTURE; see §3 below | **CLOSED** — stronger than requested, coherently |
| M5 — non-vacuous, secure-by-default inspection | ACCEPT CAUSE / REJECT PRODUCTION OPT-OUT; production always requires inspection, fail-closed; only explicit dev/test profiles may disable; no tenant-facing authority | **CLOSED** — stricter than the requested fix; strictness adds safety, removes the vacuous-profile hole entirely |
| M6 — port + conformance as first-class surface | ACCEPT; §4 promotes port+conformance, keeps Local dev/test + AWS S3 reference profile, freezes no self-hosted provider, bounds the transitional MinIO dev/CI endpoint as mechanism **with a required deletion/replacement condition**, routes conformance execution + client library to R10-C | **CLOSED** |
| L1 — S3 client library | routed R10-C; `minio-go/v7` explicitly evidence-only | **CLOSED** |
| L2 — scanner/parser order + validator hardening | routed R10-C; sandbox-platform explicitly excluded absent evidence | **CLOSED** |
| L3 — credential-journey deletion | routed R10-E; provider-hosted/themed journeys; no rebuild via admin APIs | **CLOSED** |
| L4 — cross-DB atomicity wording in C2 | ACCEPT; §8 states the invariant verbatim | **CLOSED** |
| L5 — fail-open configuration evidence | ACCEPT as implementation evidence; later proof must show production configuration cannot silently disable a required property | **CLOSED** (proof obligation must land in R10-C/implementation — DL4 below) |

No adjudication converted a review finding into a new framework, product requirement, or duplicate authority. No finding created a new bounded context or semantic owner.

---

# 3. M4 special audit — REMOVE-now instead of prove-or-remove

The prior review required the audit-PII family to be adjudicated **before** the REMOVE arm. The operator instead decided the arm now by restructuring the target: immutable Audit = PII-minimized/non-PII skeleton; human enrichment resolves through separately erasable state; mandatory DEK/KEK/wrap-unwrap/crypto-shred removed from V1.

Delta verdict: **legitimate and sound under the Method** — this is choosing the smaller structure with a falsifiable proof strategy rather than deferring the choice, because:

1. **The erasure invariant survives intact.** §6.2 preserves the full frozen sequence (retention/hold blockers → revoke access → erase eligible substantive rows/blobs → preserve allowed non-PII skeleton → tombstone → ERASED → restore reconciliation). Nothing in the frozen invariant depends on key destruction; frozen authority already required post-erasure survival of only a **non-PII** skeleton, so a non-PII skeleton design is the direct realization of the frozen wording, not a new claim.
2. **Removal weakens no real backup-erasure property.** The DEK only ever protected `audit_events.payload` (prior review §0.2); every other tenant fact was always plaintext in product rows and backups, with erasure resting on verified deletion + backup expiry + restore tombstone reconciliation. Removing the DEK changes the posture of exactly one family, and the skeleton design removes PII from that family at the root instead of encrypting it — root cause over mechanism.
3. **No false-security residue.** §6.3 forbids any cryptographic-erasure claim without a named Target Data family and fail-closed enforcement — which also retires the observed fail-open shape (KEK unset ⇒ silent no-op shred) as a target possibility.
4. **The cascade is complete.** §6.2 + §10 remove, together: ledger §6 crypto-shred step, R9.5-2 DEK statement, Organization key-custody fact family, B2 key-custody scope, mandatory platform KEK/wrap-unwrap mechanism. This is exactly the prior review's Q27 cascade (abstractions that existed only because the DEK existed).
5. **The reopen trigger is explicit and testable.** If B6's field-by-field classification (§6.3) proves an immutable Target Data family that must remain stored yet become unintelligible, R4 reopens with R10-C/R10-F backup/restore proof. Proof-before-implementation is preserved: B6 cannot close without the classification.

Coherence with append-only/tamper-evident Audit: an opaque-reference skeleton (actor/resource as UUIDs whose linking records live in erasable Organization state) is chain-stable — erasure touches the erasable side, immutable rows and their tamper evidence never change. ApprovalDecision and other domain evidence records are unaffected: they are tenant substantive state, erased wholesale at tenant erasure, so PII-minimization applies only to the surviving audit skeleton. No contradiction with any frozen semantic found.

---

# 4. Answers to the corrected target's §12 questions

1. **M1/topology:** yes — ownership amendment is bounded; 8+3 unchanged; no fact family is orphaned (each removed Authentication fact has a named provider owner; each added fact is Authentication-owned).
2. **Structural isolation:** yes — enumerated-facts-only contract + banned representations make provider role/group/org consumption unrepresentable, not merely prohibited.
3. **Operability of non-atomic provisioning:** yes — the six-case matrix plus idempotency/reconciliation requirement is the correct specification altitude for GCR; concrete choreography is B2/R10-D work.
4. **Opaque Session:** coherent — the BFF architecture requires an app session distinct from the IdP session; §3.4 carries authentication context only and explicitly refuses role/permission/group/Area-grant snapshots, so authorization changes take effect without session invalidation and the session never becomes a second AuthZ surface.
5. **M6:** yes — the entitlement is now the port + conformance contract; no replacement provider was named; the provider-name coupling class is removed, not re-instantiated.
6. **Transitional dev/CI endpoint:** correctly mechanism-only, with an explicit deletion/replacement condition, routed to R10-C.
7. **Inspection gate:** non-vacuous (production unconditional), fail-closed (three failure postures enumerated), bounded (exclusion list retained: no quarantine aggregate/CDR/rescan/intel platform/sandbox cluster/security domain).
8. **DEK removal vs erasure invariant:** preserved — §3 above.
9. **Non-PII skeleton coherence:** coherent — §3 above; B6 classification obligation makes it falsifiable.
10. **R4 reopen trigger:** present, explicit, testable (§6.3).
11. **C1/C2:** ambiguity removed, no new authority created; C2 adds only a restricting law (no cross-DB atomicity), which narrows rather than expands invariant surface.
12. **New contradiction / fifth local maximum from adjudication:** none found — the adjudication is net-subtractive (removes DEK machinery, removes provider-name entitlement, removes credential machinery) and adds only enumerated seams. Checked specifically: provider lockout vs app pre-auth rate limiting (mechanism overlap only, no authority overlap); Approval `requires_reauthentication` (assurance facts flow through the contract; frozen Approval semantics untouched); R9.5-2 encrypted-transport/provider-encryption-at-rest baseline (unchanged, still the confidentiality baseline post-DEK); Distribution/read-acknowledge, Records, Interchange (untouched).
13. **B2 resumption:** yes — §11 scope plus surviving handoff obligations plus R10-A fact families plus B1 laws are complete enough to resume after promotion (see DL2 for the one promotion-mechanics precision).

---

# 5. LOW findings (promotion-precision notes; none blocks approval)

| ID | Note |
|---|---|
| DL1 | **Binding uniqueness — name the tenant dimension.** §3.2 defers cardinality/uniqueness to B2 and names the one-User-many-subjects direction, but not the more likely inverse: the same provider `(issuer, subject)` bound to Users in **multiple Tenants** (consultant/operator personas). Global `UNIQUE(issuer, subject)` would make cross-tenant membership unrepresentable; per-tenant uniqueness (`UNIQUE(tenant_id, issuer, subject)`) is the shape B1's tenant-owned law suggests. B2 must decide this explicitly. |
| DL2 | **Promotion must amend the handoff B2 list line-by-line, not replace it with §11.** §11 is an area map. The handoff's existing must-decide list carries named obligations absent from §11 (tenant settings/configuration persistence, grant/revocation evidence, canonical grant-evaluation read model, per-family persistence/mutation classification, same-Tenant FK/RLS application, transaction boundaries). All are independently anchored in R10-A fact families and B1 law, so no authority is lost either way — but promotion should remove/add specific lines (credential/DEK lines out; binding/choreography/no-provider-role-proof lines in) to avoid scope-regression ambiguity. |
| DL3 | **Stale-mirror sweep at promotion.** `wiki/architecture/cohesive-platform-redesign.md` §5 mirrors frozen lines that the deltas amend — at least the Retention bullet "required decryption capability remains until preserved content may lawfully be destroyed", the Authentication bullet ("current MetalDocs authentication/session approach is acceptable for V1"), the build-vs-buy Keycloak ruling (§9), and the storage bullet naming MinIO (§5). Promotion must sweep every mirrored statement in the program authority in the same change, or the program page contradicts the amended ledger. |
| DL4 | **Profile-leak proof needs a named owner.** "Explicit dev/test profiles may disable inspection" is safe only if profile declaration is structurally single-sourced (an inspection-disabled deployment cannot present itself as production). R10-C already owns gate-bypass proof; add profile-declaration integrity to that proof obligation explicitly, completing the L5 disposition. |
| DL5 | **Provider-side disable vs live app session.** §3.5 covers the binding state for a removed/disabled provider subject; B2 should also state the app-session posture when the provider disables a subject mid-session (accept bounded staleness until expiry/revocation, or consume back-channel logout later). Authentication owns application-session revocation, so this is a B2 line item, not a new seam. |

---

# 6. Gate check

Verified at `31470deb`:

```text
delta since prior review  = 1 new artifact only (corrected target)
R9.5 ledger               = untouched
R10-A authority           = untouched
R10-B1 authority          = untouched
current-agent-handoff     = untouched
code / schema / OpenAPI   = untouched
implementation gate       = CLOSED, respected
```

The corrected target correctly declares itself non-authority pending this check.

---

# 7. Resulting state

```text
VERDICT: APPROVE GCR ADJUDICATED CORRECTED TARGET

BLOCKER = 0
MAJOR   = 0
LOW     = 5 (DL1–DL5, promotion-precision only)

prior findings M1–M6, L1–L5 = 11/11 CLOSED
new material contradiction   = NONE
new fifth local maximum      = NONE

PROMOTION = MAY PROCEED on this corrected target without another broad review,
            executing the §10 bounded amendments plus DL2/DL3 promotion mechanics,
            with DL1/DL4/DL5 carried as named B2/R10-C line items.

R10-B2 = MAY RESUME after promotion with the corrected scope.
```

The GCR review loop terminates here: candidate → independent cold review → operator adjudication/corrected target → this bounded delta check. Further review re-enters only through the stage gates of B2+ or a genuine new-evidence reopen trigger.
