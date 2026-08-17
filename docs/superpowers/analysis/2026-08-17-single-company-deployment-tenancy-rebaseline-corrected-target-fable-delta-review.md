# MetalDocs — Single-Company Deployment / Tenancy Rebaseline — Corrected Target — Bounded Delta Review

> **Status:** INDEPENDENT BOUNDED DELTA REVIEW — evidence, **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** independent cold session (Fable), delta scope only
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed HEAD:** `31a57e5b46b9245a8a1466fff66d9f1b1ace20cd`
> **Reviewed artifact:** `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-adjudicated-corrected-target.md`
> **Prior independent review:** `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-independent-fable-review.md` @ `1acd5128`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Scope:** bounded — verify M1–M3 / L1–L6 closure and hunt new material contradictions. No whole-platform re-review.
> **Implementation gate:** CLOSED — this review changes no authority, code, schema, OpenAPI, frontend or deployment.

---

## 0. Delta basis

Verified `git diff 1acd5128..31a57e5b`: exactly one new file (the adjudicated corrected target, 752 lines). All four authorities (`AGENTS.md` chain: Method mirror, handoff, program authority, ledger, R10 technical authority) are byte-unchanged since the prior review — correct sequencing; promotion is properly gated behind this delta review. Additional evidence gathered this round: a mechanical residue grep across the four authority documents for tenant-machinery wording (`erasure|ERASED|tombstone|Portability|tenant.export|tenant.deletion|RLS|tenant_id|per-Tenant|tenant context`), used in D-L1.

---

## 1. Verdict

```text
VERDICT = APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY
          ADJUDICATED CORRECTED TARGET

BLOCKER = 0
MAJOR   = 0
LOW     = 3   (all promotion-mechanics riders; none reopens an adjudicated decision)

prior findings closed        = 9/9  (M1–M3, L1–L6)
new material contradiction   = NONE
```

Promotion may proceed after this review, subject to the three LOW riders being executed **inside the promotion change set** (§5). No further broad review is required.

---

## 2. Job 1 — prior-finding disposition

```text
M1. CLOSED — corrected target §5 — singleton Tenant root UUID + configured
    expected_tenant_id + startup/readiness compare; mismatch/missing/multiple
    roots = FAIL CLOSED; no Deployment aggregate/control plane/per-row column;
    company-root identity kept distinct from the GCR-R3 security-profile
    declaration (a legitimate NARROW: the profile-integrity property already has
    its own single-sourced mechanism owned by R10-C; two properties, two
    mechanisms, no gap).
M2. CLOSED — §11 + §18 — customer lifecycle deferred whole; user offboarding/
    session revocation, erasable enrichment, PII-minimized immutable Audit
    skeleton, B6 field-by-field classification, restore non-resurrection proof
    and the GCR-R4 crypto-erasure trigger all re-anchored to user/data-subject
    level; anti-overbuild guard added (no PrivacyCase/generic privacy platform);
    Retention/Hold/Disposition/Backup untouched.
M3. CLOSED — §12 — base catalog 29→27 (`tenant.export`, `tenant.deletion.request`
    removed), R9.5 delta 16, total 43, five roles unchanged; arithmetic verified
    against the frozen ledger §1 enumeration (7+11+4+2+5=29 → 27);
    reinstatement routed through the owning feature's formal reopen.
L1. CLOSED — adjudication row L1 + §5 + §17 — exactly-one-root enforced
    structurally with a runtime fail-closed check and an explicit
    can-be-shown-to-fire proof obligation; ACTIVE/one-state lifecycle correctly
    refused (row existence IS aliveness).
L2. CLOSED — §4 — Tenant/TenantScope/tenant_owner/tenant.settings.manage kept as
    one consistent vocabulary; binding redefinition recorded; partial rename
    forbidden; anti-inertia list explicit; deliberate future rename trigger.
L3. CLOSED — §7 + §17 — non-owner + NOSUPERUSER serving role and separate
    non-serving maintenance trust surface retained as ordinary least privilege;
    NOBYPASSRLS/per-Tenant-iteration wording dropped as vacuous.
L4. CLOSED — §15 — opaque/immutable/no-overwrite/SHA-256 laws retained as the
    real invariants; tenant prefix demoted from invariant; prefix removal not
    mandated; bucket/account isolation = ops configuration.
L5. CLOSED — §11 + adjudication L1 correction — whole lifecycle deferred; no
    vestigial SUSPENDED; deployment stop/maintenance = operations.
L6. CLOSED — §16 — B2–B5 explicit uniqueness-rederivation obligation with
    examples; final constraints correctly not invented in advance.
```

All nine dispositions are faithful to the independent review's required fixes; the two operator narrowings (M1 profile separation; L1 no-ACTIVE-state) are improvements, not weakenings.

---

## 3. Job 2 — new findings (attack on the corrected material)

### D-L1 — LOW — §22's amendment enumeration is incomplete; promotion must execute a mechanical residue sweep

The corrected target's amendment map (§22) enumerates ledger §1/§6/§9/§13 plus generic R10-A/R10-B1/mirror lines. The residue grep proves tenant-machinery wording also lives at sites the enumeration does not name. Known residual sites that would contradict or mis-anchor the promoted substrate if left verbatim:

```text
ledger 2026-08-14 §5   (line ~313)  Audit-skeleton obligation worded against
                                    "Tenant erasure" — must re-anchor per M2
ledger §6              (line ~351)  "Backup/restore must reapply erasure
                                    tombstones" — tenant-level tombstones deferred;
                                    becomes user-level non-resurrection proof
ledger §17 R10 list    (items 11, 18) "retention-aware tenant erasure/restore
                                    tombstone reconciliation" + post-erasure
                                    skeleton wording — re-anchor
r10 §2.2 Records       (line ~251)  "retention/hold blocker facts used by Tenant
                                    erasure" — consumer becomes deferred
r10 §2.2 Audit         (line ~309)  GCR-R4 paragraph worded against Tenant erasure
r10 §3 seam #9         (line ~396)  Tenant erasure/restore coordination seam
r10 §9.1–9.10, §9.13–9.14, §10      composite-PK/RLS/tenant-context/discovery laws
                                    and their proof obligations (already §22-listed
                                    generically; listed here for sweep completeness)
program authority §4 supporting list ("Tenant Lifecycle / Security"), §5 Retention
                                    bullets (~240–242), §7 Organization ruling (~319),
                                    §12 B2 checklist (~403–410)
handoff: checkpoint B1 law block (~228–262), B2 checklist (~320–329),
                                    B6 successor obligation (~361), GCR-R4 lines (~159–164)
```

Adjudicated intent is unambiguous (§11/§17/§18 state the re-anchored semantics), so this is promotion mechanics, not a decision defect. **Rider:** the promotion change set must (a) amend by residue sweep (grep-driven), not by the §22 section list alone, and (b) record the sweep pattern + result in the promotion commit so completeness is falsifiable. Left unexecuted, the promoted authorities would carry sentences treating Tenant erasure/tombstones as live V1 concepts — exactly the two-authorities/self-contradiction defect class this program exists to prevent.

### D-L2 — LOW — §19's four-package B2 scope is a summary and must not replace the surviving detailed checklist

The current handoff/R10 §10 B2 checklist contains surviving line items that §19 compresses or omits: per-family semantic-persistence-class × mutation-law classification; provider-side disable vs already-live Session posture; the six enumerated provider-reconciliation cases (subject absent / binding absent / removed-or-disabled / duplicate attempt / provider unavailable / uncertain-response retry). **Rider:** promotion must carry the surviving detailed items forward into the four-package structure; replacing the checklist with §19's summary would silently shrink B2's proof surface.

### D-L3 — LOW — Tenant root `id` must be classified immutable in B2

The handshake's trust anchor is the root row's UUID. B2's mutation-law classification (obligation unchanged per §17) must declare Tenant root identity immutable (settings mutable, id fixed), so a future mismatch incident is resolved by fixing configuration or restoring the right database — never by mutating the anchor. One classification line; no new mechanism.

No other new finding. Specifically checked and clean: no new root/company abstraction, no dual vocabulary, no Deployment aggregate/control-plane requirement, no compensating Area/role/permission RLS, no reintroduced partition column under another name, no contradiction with any §21 non-reopened decision, no weakening of the prior review's §10 must-survive list (all eight items verified present).

---

## 4. Required checks 1–18

| # | Check | Result |
|---|---|---|
| 1 | One company per deployment; no Metal Nobre hardcode; same build/migrations; second customer ≠ pooled | **PASS** — §3: config/data vary per deployment; forks forbidden; second customer → economics review only |
| 2 | Singleton Tenant root justified, separated from partitioning | **PASS** — §4: five real consumers; binding definition; anti-inertia list |
| 3 | Exactly-one-root coherent without ACTIVE lifecycle | **PASS** — §5 + adjudication L1: existence + fail-closed check; no state machine |
| 4 | Handshake preserves wrong-DB safety without Deployment aggregate | **PASS** — §5: mismatch/missing/multiple fail closed; profile identity separate (R10-C/GCR-R3) |
| 5 | id-only PK/FK preserves pooling-independent laws | **PASS** — §6: UUID identity, business-ID-never-PK, authority-neutral RESTRICT/NO ACTION, cascade restraint, no polymorphic registry all survive; anti-inertia law blocks `company_id`-by-reflex |
| 6 | RLS removal preserves real V1 security | **PASS** — §7: deployment boundary isolates the only dataset; no compensating Area/role RLS; non-owner/NOSUPERUSER + maintenance trust surface retained |
| 7 | Keycloak Organizations deferred without harming federation | **PASS** — §8: realm-level federation stays provider configuration; issuer explicit |
| 8 | Binding/session tenant dimension removed; boundaries intact | **PASS** — §9: `UNIQUE(issuer,subject)`; no AuthZ snapshot; anti-corruption retained; 1↔N bindings stays a legitimate B2 decision |
| 9 | TenantScope necessary and non-partitioning | **PASS** — §10: whole-company vs Area distinction real; representation freed of redundant payload |
| 10 | Lifecycle deferred without losing data-subject privacy | **PASS** — §11: M2 re-anchor complete incl. restore non-resurrection + GCR-R4 trigger; anti-overbuild guard |
| 11 | Only the two permissions orphaned; catalog arithmetic | **PASS** — §12: 27+16=43 verified; sweep of remaining 43 found no other permission whose capability was deferred |
| 12 | Portability deferred; Backup/GSE/PUBLISH_COPY/Historical Migration complete | **PASS** — §13: completeness law retained; stamp moves via backup/restore |
| 13 | Async: routing out, correctness mechanisms in | **PASS** — §14: outbox/intent/idempotency/lease/retry/DLQ retained; same-commit law survives |
| 14 | Key layout freed; opaque/immutable/no-overwrite/hash retained; no forced prefix removal | **PASS** — §15 |
| 15 | Uniqueness rederivation routed to B2–B5; no premature final constraints | **PASS** — §16 |
| 16 | No new root abstraction / dual vocabulary / control plane | **PASS** — §3 attack found none |
| 17 | Pooled triggers concrete | **PASS** — §20: five measured-evidence triggers + deliberate design stage + backfill anchor |
| 18 | B2 restart after promotion without another broad review | **PASS** — with §5 riders; see below |

---

## 5. Promotion conditions

Promotion may proceed **after this review** with no further broad review, provided the promotion change set:

```text
1. amends the four authorities per corrected target §22 PLUS the D-L1 residue
   sweep (grep-driven; pattern + result recorded in the promotion commit);
2. carries the surviving detailed B2 checklist items into the §19 four-package
   scope (D-L2), removing only the lines the rebaseline actually deleted;
3. adds the Tenant-root-id immutability classification obligation to B2 scope (D-L3);
4. embeds the §20 pooled-tenancy trigger set in amended B1 reopen triggers
   (already declared in §22);
5. records the decision as RESTRUCTURE NOW with this evidence chain
   (candidate @ cba89d9d → independent review @ 1acd5128 → adjudicated corrected
   target @ 31a57e5b → this delta review);
6. keeps B2 paused until promotion lands and keeps implementation BLOCKED.
```

Operator ratification of the promotion commit is the final gate; the riders are text-completeness obligations, not new decisions requiring re-adjudication.

---

## 6. Convergence statement

Prior findings 9/9 closed. New findings: 0 BLOCKER, 0 MAJOR, 3 LOW — all mechanical text-completeness riders at strictly lower altitude than the prior round (substrate decisions → promotion mechanics). Altitude is descending and the remaining work is exactly the kind a promotion diff can prove. **Stop condition met; review loop converged.**

```text
VERDICT = APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY ADJUDICATED CORRECTED TARGET
```
