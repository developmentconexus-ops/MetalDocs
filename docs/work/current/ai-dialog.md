# T11 B01–B09 Checkpoint — Fresh Independent Fable Adversarial Review (R1)

> **Role:** Challenger (fresh independent whole-checkpoint review).
> **Nature:** Evidence only — never authority.
> **Date:** 2026-08-24.
> **Candidate:** PR #162 `arch/t11-implementation-program` @ `30486a93602622ed4eed5828584aa59a5441a9ce` against `main` @ `cae6ba48df5d611959c0390e0f2b9b8194d62a9d`.
> **Review branch:** `review/t11-b01-b09-checkpoint-fable-r1` (this dialogue is the only delta; it never merges).

## 1. Revalidation of the handoff claims

All handoff claims were revalidated against remote authority before review:

```text
candidate HEAD                     30486a93 CONFIRMED (origin/arch/t11-implementation-program)
base                               cae6ba48 CONFIRMED (origin/main)
PR #162 state                      OPEN / DRAFT / MERGEABLE / UNMERGED CONFIRMED
net diff                           21 files / 5880 additions / 280 deletions CONFIRMED
net file list                      exactly the 21 listed files CONFIRMED
Draft CI run #1490                 SUCCESS on exact HEAD 30486a93 CONFIRMED
required job 97540685520           name=required, SUCCESS, exact HEAD CONFIRMED
docs/work/**                       0 tracked files at candidate HEAD CONFIRMED
candidate tree                     51 tracked files; all .md except ci.yml/.gitignore/.gitattributes CONFIRMED
runtime/schema/OpenAPI/manifests   0 CONFIRMED
evidence ref                       evidence/t11-pr162-b01-b09-locks-20260824
                                   resolves to adf58e448bc5bd3a20cae5b7228d729c031f94ac CONFIRMED
```

## 2. Lock Evidence locator verification (byte-exact)

Every blob identity in `docs/decisions/t11-b01-b09-lock-evidence.md` §3 and §4 was verified by `git ls-tree` against the exact Evidence commit `adf58e44`:

```text
B01  6d4f5c25...  MATCH        B06  785473b8...  MATCH
B01N 17dd3570...  MATCH        B07  20ec64d3...  MATCH
B02  3ac7217f...  MATCH        B08  bb130535...  MATCH
B03  6bffa8c2...  MATCH        B09  7daa6054...  MATCH
B04  8da45d99...  MATCH
B05  0942dfb9...  MATCH
B09 P9 screen contract  ece854b4...  MATCH
B09 P10 consolidation   8d8328be...  MATCH
```

B09 closure-proof spot-checks against the Evidence blobs:

```text
P9 material controls   33 control rows counted; "33 / 33 traced / 0 unbound" stated in blob  CONFIRMED
P9 invented operations 0; "operation 90+ absent" stated in blob                              CONFIRMED
P7 exit                "unresolved upstream findings 0" stated in blob                       CONFIRMED
P10                    B09 pattern pass names locked P8 blob 7daa6054; reuse-only outcome    CONFIRMED
```

## 3. Census falsification result

Attacked arithmetically and semantically across all affected durable documents:

```text
76 (journeys) + 2 (T8-E read symmetry) + 8 (Discussion/Notifications) + 3 (Audit) = 89   HOLDS
Idempotency 10 (wire-contract "exact accepted 10") + 1 (createDocumentDiscussionMessage) = 11  HOLDS
ETag 13/13, exact-byte 4: unchanged from base wire-contract census                        HOLDS
PermissionCode: 15 enumerated in authorization-and-audit.md §4 + document.discuss = 16    HOLDS
Routes: 10 (base journeys) + /notifications = 11; discussion doc lists exactly 11         HOLDS
Owners: 4 + 1 (base ownership) + Notifications = 4 + 2; ownership.md now states 4+2       HOLDS
op numbering: 47/53/55/67/78 verified against the base wire ledger rows                   HOLDS
maxItems=37 for operation_codes matches wire-contract's 37-entry discriminator mapping    HOLDS
```

No numeric or semantic contradiction between current authorities was found. Stage-snapshot "78"/"86" statements are addressed by the census supersession law (see F2 for a residual clarity gap).

## 4. Cross-layer attack results (summary)

- **Discussion/Mention/Notifications:** ownership not duplicated; Notification never source authority (ownership §7 non-ownership list, domain-model binding laws); Mention explicitly never an access grant; Notification READ explicitly ≠ Read & Acknowledge; no event bus/broker/Redis; same-Scope atomicity is local-transaction, not platform; presentability recheck server-side before paging/counts. HOLDS.
- **Document Official action hints:** `allowed_actions` UX-guidance-only with structural inversion test; commands recheck truth; rejected screen-shaped `GET /actions`. HOLDS.
- **My Work:** projection-only, B06 remains case authority; F4 order + four bounded presets; cursor anchor server-owned. HOLDS.
- **Governance deadlines:** Controlled Documents owns config/snapshot/frozen due_at; breach = presentation only, no worker/notification/lifecycle effect; B06 read never inherits queue due_at. HOLDS.
- **Review-layer seam:** zero current capability delta; explicitly forbids dormant/"coming soon" controls; GOV-12 added to forward obligations. HOLDS.
- **Document History:** History ≠ Audit preserved both directions; frozen step_label from attempt snapshot; no current-resource resolver; no compare/restore authority. HOLDS.
- **B09 Audit:** op78 sole traversal authority, no detail endpoint, Query Assist purpose-built + Audit-visible-only + guidance-never-authorization; recognition never filters/orders/enters cursor; owner handoffs recheck destination disclosure; admin deep-links deferred to B10–B12; no frontend AuthZ matrix; dependent-filter invalid states closed (actor/resource laws → 400). Operation 90+ absent. HOLDS.
- **Boundary/regression:** T11 remains OPEN; B10/B11/B12/FP2/P11/T12 NOT OPEN; implementation BLOCKED with explicit no-implied-authorization hard stop; no methodology adoption smuggled (AGENTS still cites Method/Repository Standard v1.0.0; method doc is the pre-existing local v2.3 authority). HOLDS.
- **Repository/CI:** README landing-only; roadmap sole mutable status authority; forward-obligation counts 21/3/27=51 verified by grep and CI-enforced; all durable docs routed (CI link walk); review-branch isolation lane, Draft-only flat `docs/work/current/*.html` allowance with negative self-tests, non-Draft rejection of `docs/work/**` both via allowlist and the explicit ls-files gate; Ready transition has no hidden contradiction (candidate tracks zero docs/work files). HOLDS.

## 5. Findings

### F1 — IMPORTANT — Evidence-ref preservation has no firing mechanism and is absent from the provenance-ref register

- **Exact claim under attack:** locator §2 "The Evidence ref must not be moved or deleted while T11/P11 still depends on these LOCK artifacts", and the handoff claim "exact operator-LOCKED bytes survive candidate cleanup / future P11 can deterministically retrieve them".
- **Evidence / counterexample:**
  - `evidence/t11-pr162-b01-b09-locks-20260824` @ `adf58e44` is reachable today only through the candidate branch history and the evidence ref. Integration is squash merge (AGENTS.md), so after merge + normal candidate-branch deletion the evidence ref becomes the **sole** keeper of every locked P8 blob.
  - The repository's own doctrine makes this insufficient as-is: `docs/development/engineering-rules.md` — "A control counts only when its negative path is demonstrably capable of firing."
  - The repository already has the exact mechanism for this artifact class and does not apply it here: `.github/workflows/ci.yml` pins `archive/r10-pr131-pre-reset-20260820` and `archive/repository-governance-pr132-20260820` to exact SHAs on every run; the new evidence ref has no pin.
  - `AGENTS.md`: "Required unmerged provenance refs are recorded in `docs/decisions/repository-reset.md`" — `repository-reset.md` §"Durable unmerged provenance refs" records only the two PR #131/#132 archive refs; the new evidence ref, which is the same class (durable unmerged provenance with a named consumer, P11), is recorded only in the locator.
- **Why it matters:** accidental deletion of the evidence ref (most likely during exactly the post-merge branch-cleanup window this checkpoint creates) makes `adf58e44` unreachable and eventually GC-able; the operator-LOCKED P8 bytes the whole checkpoint exists to preserve would be unrecoverable, and the locator's §6 P11 retrieval law would be unsatisfiable. This is the "lost P8 evidence" underengineering failure mode of the review question.
- **Smallest implicated authority:** `.github/workflows/ci.yml` (ref-verification block) + `docs/decisions/repository-reset.md` §"Durable unmerged provenance refs" (register row). The locator itself needs no semantic change.
- **Falsifier / smallest disposition:** add the evidence ref to the existing CI exact-SHA ref-verification block (`evidence/t11-pr162-b01-b09-locks-20260824` == `adf58e448bc5bd3a20cae5b7228d729c031f94ac`) and register the ref in `repository-reset.md` alongside the two archive refs. The locator §7 retirement rule then governs future removal of the pin under explicit authority, exactly as for the archive refs. Alternatively, prove an equivalent existing mechanical protection (e.g. a ruleset that blocks `evidence/*` deletion) — no such protection is visible from repository content.

### F2 — MINOR — Current-tense "census = 86" statements inside documents added by this same PR

`discussion-notifications-launch.md` (§12 "Current application census after promotion: 86", §19 proof obligation "application operation census = 86", §21) and the four B05–B08 precision documents ("Current accepted census remains: 86 operations / 11 routes / 16 PermissionCode values") carry current-tense totals that were stage-accurate at their ratification (2026-08-23) but conflict verbally with the current 89. `api-operation-census.md` claims supremacy and explicitly lists "86" statements as superseded stage snapshots, so the conflict is resolvable by a careful reader; however its supersession category list ("Product/T6 authority, T4/T8/T9/T10 pages, ratification records, precision provenance, other pre-current-T11 closure snapshots") does not unambiguously name T11's own decision documents, and §19 of the discussion authority is a forward-looking proof list that a literal implementer could execute against 86. Smallest disposition: one clarifying line in `api-operation-census.md` naming the census blocks of current T11 bounded decisions as stage snapshots, or no action if the operator judges the existing supremacy language sufficient.

### F3 — MINOR — Discharged B01 re-LOCK obligation reads as pending

`discussion-notifications-launch.md` §11: "A smallest-scope rendered B01 P8 delta still requires operator re-LOCK before the structural frontend block is closed again." The roadmap (sole status authority) records `B01N … LOCKED / OPERATOR-RATIFIED` and the locator preserves the B01N P8 blob, so the obligation is discharged. The sentence is stage residue, not a contradiction of authority; roadmap wins. Smallest disposition: none required; optionally reword on the next lawful touch of that document.

### F4 — UNSUPPORTED PREFERENCE — "sole numeric census" label wider than the census document's coverage

`docs/roadmap.md` labels `api-operation-census.md` "the sole numeric census", but that document enumerates only operations/Idempotency/ETag/exact-byte (89/11/13/13/4); owners, routes and PermissionCode counts live in the roadmap census block and the bounded decisions. No contradiction exists; this is a wording-precision preference only.

## 6. Verdict

```text
MATERIAL               0
IMPORTANT              1   (F1)
MINOR                  2   (F2, F3)
UNSUPPORTED PREFERENCE 1   (F4)

VERDICT                LEAD RESPONSE REQUIRED
```

The cleaned 21-file candidate is otherwise a coherent, truth-preserving, YAGNI-bounded B01–B09 acceptance checkpoint: it does not falsely close T11, creates no Product/runtime authority, preserves every lock byte-exactly today, and its census/authority routing is internally consistent. The single IMPORTANT finding concerns the durability mechanism of the lock evidence after integration, not the content of any settled Product decision. No settled decision is reopened by this review.
