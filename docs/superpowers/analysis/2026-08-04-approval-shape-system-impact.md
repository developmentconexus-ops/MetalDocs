# System-impact analysis — Approval shape: one decision surface for review and signature

**Date:** 2026-08-04
**Intent (one line):** Collapse the duplicated review/signature expression in `approval` into one
decision surface, make "signing includes conversing" a mechanical invariant instead of a grant, and
give the two rejection paths a single unambiguous meaning.
**Work type:** feature (architecture refactor inside an existing module)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow — proceed to design under the locked constraints in §10; an ADR is mandatory.

---

## 0. What triggered this

The operator's objection, restated precisely: a route variant that only signs makes no sense,
because a signer who finds a defect must review and return the document anyway. Investigation
confirmed the objection is real but mislocated — the missing thing is not the *behaviour*, it is the
*expression*. Findings, each read from code:

- **Route shape does not require review anywhere.** `assert_route_shape` (current text:
  `db/migrations/0316_livre_zero_stage_route.sql:73`, superseding the baseline copy at
  `db/baseline/0001_current_schema.sql:184`) enforces: `livre` ⇒ zero stages; every other class ⇒
  ≥1 stage; `controlado` and any unknown class ⇒ ≥1 `approval`-kind stage. No arm ever requires a
  `review`-kind stage. An approval-only `controlado` route is legal by construction.
- **An approval stage can already return the document.** `review_verdict_service.go:177` permits
  `request_changes` on an `approval`-kind stage and blocks only `ready` there
  (`domain.ErrVerdictReadyOnApprovalStage`). So R3 ("signing ⊃ conversing") is implemented.
- **Two rejections, one stage, no rule.** On the same active approval stage an actor may record
  `request_changes` (→ instance `changes_requested`, document back to `draft`,
  `revision_version + 1`, non-terminal — `review_verdict_service.go:365-412`) *or* a signoff with
  `decision='reject'` (→ stage `rejected_here`, instance `rejected`, terminal —
  `decision_service.go:638-694`). Nothing in the contract, the DB, or the UI says which is correct
  when. This is the actual ambiguity.
- **The two records are near-identical.** `approval_review_verdicts`
  (`db/baseline/0001_current_schema.sql:2062`) and `approval_signoffs` (`:2151`) share
  instance/stage/actor/tenant/comment/display-name-snapshot/on-behalf-of, the same
  `(stage_instance_id, actor_user_id)` uniqueness, and the same `enforce_approval_sod` trigger
  (`:4024`). The signoff adds exactly five columns: `decision`, `signature_method`,
  `signature_payload`, `content_hash`, `signature_meaning`. **A verdict is a signoff without the
  e-signature ceremony** — yet it carries two services, two endpoints, two DTO families and two
  screens.
- **The duplication is literal, not conceptual.** `ReviewVerdictService` and `DecisionService` each
  call the same `domain.ResolveEligibleIdentity`, the same `domain.CheckSoD`, and each carry their
  own `emitEligibilityRejection` with identical bodies
  (`review_verdict_service.go:448`, `decision_service.go:764`).
- **R3 is a grant, not an invariant.** `RecordVerdict` hard-requires
  `authz.Require(CapApprovalReview)` even on an approval-kind stage
  (`review_verdict_service.go:164`). `approval.review` is granted per profile
  (`internal/modules/iam/domain/model.go:94`). A tenant may therefore configure a signer profile
  holding `document.signoff` without `approval.review`, whose only "no" is the terminal reject. That
  is the operator's scenario, reachable today, guarded by nothing.

Correction of record: an earlier reading of the *baseline* text reported a permissive
`simples`/unknown fall-through. Migration 0316 already closed it — unknown classes fail closed onto
the `controlado` rule. Runtime truth is baseline + migrations.

## 1. Classify & own

- **Work type:** feature. No new module is born; the change is internal shape plus contract.
- **Owning module:** `approval` — it owns `approval_review_verdicts`, `approval_signoffs`,
  `approval_stage_instances`, `stage_kind`, the SoD trigger and the route-shape function's callers.
  ADR 0082 made it the 15th top-level module precisely so this kind of ruling has one home.
- **Explicitly NOT owning:**
  - `documents` — it owns the document status machine and the `under_review → draft` edge, but the
    *decision* that triggers that edge is approval's. Approval must keep driving it through the
    existing gate (the `metaldocs.cancel_in_progress` GUC + `CanTransitionDocumentStatus`), never by
    widening documents' rules.
  - `taxonomy` — it owns `governance_class`. Any change to what a class *requires of a route* is
    approval's policy expressed in approval's DB function; taxonomy's enum is untouched.
  - `iam` — it owns the capability catalog. Approval may need a capability changed, but the
    implication rule ("signing includes conversing") is approval's semantics; where it is *encoded*
    is decided in §4.
- **Cross-module edges (direction: A → B = A depends on B):**
  - `approval → iam` — `authz.Require`, capability constants. Published; unchanged in kind.
  - `approval → documents` — `docsdomain.CanTransitionDocumentStatus`, lifecycle event envelope.
    Published; unchanged in kind.
  - `approval → controlleddocuments` — area-code resolution via `docapp.LoadDocumentAreaCode`.
  - `approval → notifications` — via the approval-owned notification envelope (in flight on
    `feat/approval-accountability-loop`). Consumer-side only.
  - `jobs → approval` — SLA surfacer through `SLAOverdueReader` / `SLASurfaceWriter`. Untouched.
- **Ambiguity?** None. **AS-3 does not fire.**

## 2. Foundation verdict

- **Base:** two parallel decision pipelines that a single pipeline with a discriminator would
  express, plus one product rule (R3) held only by tenant configuration.
- **Judgement: the base is a patch, and the operator's read is correct.** The evidence is the
  literal duplication in §0 — same eligibility helper, same SoD predicate, byte-equivalent
  `emitEligibilityRejection`, near-equal tables, one uniqueness rule per table where a single record
  would need one. Duplication of *predicates* is the signature of a shape that grew by accretion
  rather than by decision. The concepts (review ≠ approval) are sound and regulatorily required
  (ISO 13485 §4.2.4 asks for review *and* approval; 21 CFR Part 11 §11.50 requires meaning-of-
  signature). What is legacy is the belief that two sound concepts need two mechanisms.
- **Global-maximum structure (named, for the design to work within):** *one decision record, one
  decision service, one endpoint, discriminated by an explicit outcome enum, with the e-signature as
  an optional attached ceremony required by the stage kind — not by the endpoint the caller
  happened to pick.* Concretely: outcome ∈ {`approve`, `return_for_changes`, `reject`} on a single
  record; `stage_kind='approval'` ⇒ signature block MUST be present (DB constraint, not app check);
  `stage_kind='review'` ⇒ signature block MUST be absent, `approve` narrows to "review passed".
  The 21 CFR Part 11 columns stay exactly where they are — nothing about unification weakens the
  signature evidence; it only stops a second, ceremony-free table from existing beside it.
- **Does the work optimize *inside* the patch?** No — it replaces it. **AS-2 does not fire.** (AS-2
  would fire on the opposite proposal: "add a rule about when to use reject vs request_changes",
  which cements the two-mechanism shape.)
- **Trade-off to state honestly:** the unification is a destructive contract change. Two endpoints,
  two DTO families and two frontend screens collapse into one. Per the operator's standing
  extermination rule the old wire fields are dropped clean, not relaxed-to-optional — which means
  frontend and any external consumer break in the same release, deliberately.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes** | The whole R3 question is a capability question. The design must express "a signer may return" as a capability implication or a stage-kind rule — never as "the approver role can". Tier-1 row for the unified endpoint; tier-2 `authz.Require` in-tx per outcome. | `authz.Require` (`iam/authz/authz.go:76`), `authz.SeedTxIdentity` |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** | The unified endpoint replaces `/review-verdict` and `/signoffs` by editing `api/openapi/v1/openapi.yaml` and regenerating. Note the embedded-spec churn rule: full regen, never partial. | `api/cfg.yaml` + `gen.go` |
| Multi-tenant pooled | Yes (inherited) | The unified table keeps `actor_tenant_id`, its `tenant_isolation` policy and `FORCE ROW LEVEL SECURITY` — both current tables already have them (`:2077`, `:2173`). No new tenancy surface. | `tenant.FromContext`, `authz.SeedTxIdentity` |
| Async = transactional outbox | Yes (inherited) | Governance events and lifecycle/notification enqueues stay in the business tx via the existing enqueuers. No new network call. | outbox repo, approval lifecycle enqueuer |
| DB enforces invariants | **Yes — load-bearing** | Three rules move from app code to constraints: (a) signature block present ⟺ `stage_kind='approval'`; (b) `ready`-equivalent outcome forbidden on an approval stage without a signature; (c) whatever route-shape rule §7 settles. `enforce_approval_sod` must be re-pointed at the unified table. | `assert_route_shape`, `enforce_approval_sod`, CHECK constraints |
| Cross-module via published interface only | Yes (inherited) | All four outbound edges in §1 already go through published interfaces; unification does not add an edge. The `documents` status flip continues through `CanTransitionDocumentStatus` + the cancel GUC — approval must not start writing documents' rules. | `docsdomain.CanTransitionDocumentStatus` |

**AS-1 does not fire** — no invariant is violated by the proposed direction. One invariant is
*strengthened*: rules currently held only in Go move to the DB line.

## 4. Capability wiring

Not a straightforward "add a capability" — the likely shape is a *narrowing or an implication*, so
the 10 touchpoints are walked as a change-impact list rather than an addition list.

1. **const + `validCapabilities`** — `approval.review` (`iam/domain/model.go:94`) and
   `document.signoff` both survive; the design decides whether a third is needed or whether the
   implication is encoded without one. Prefer **no new capability**: an implication is cheaper and
   less error-prone than a cap nobody remembers to grant.
2. **scope classify** — unchanged (`capability_scope.go`); if any cap is added it must be classified
   or `TestEveryCapabilityClassified` reds.
3. **tier-1 route→cap** — the unified endpoint needs its own explicit row in
   `apps/api/cmd/metaldocs-api/permissions.go`, and the two retired routes' rows must be **deleted
   in the same commit**. The generic `/api/v1/approval/` prefix fallback no longer exists, so a
   forgotten row falls through to `VisibilitySessionRequired` — silent privilege escalation.
4. **tier-2 `authz.Require` in-tx** — one call site instead of two, with the required capability
   selected by `(stage_kind, outcome)`. This is where R3 stops being a grant: on an approval-kind
   stage, `return_for_changes` must be authorised by the *signature* capability, not by
   `approval.review`. That single change closes the live gap in §0.
5. **seed grants** — `db/reference-data/0001_product_reference_data.sql` re-checked so no shipped
   profile can hold signature power without return power.
6. **DB tripwire** — `ck_cap_format` / `ck_cap_not_legacy` accept any name used; the
   subject-discriminated tripwire arms (ADR 0083) must be re-generated if a route/cap pair changes.
7. **guard tests** — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`, and the
   tripwire-arm parity lint must stay green.
8. **`TestCapabilityRegistrySize`** — currently `const want = 39`
   (`iam/domain/model_test.go:96`) **in the working tree, which already contains another session's
   in-flight `approval.sla_extend` work**. Re-verify against the merge base at design time; do not
   quote 39 as the committed baseline.
9. **CI capability-coherence (REQ-AUTHZ-5)** — the const/classify/tier-1/seed/test surfaces must
   agree after the route retirement.
10. **H-PRE-1** — the unified service performs `authz.Require` inside the decision tx exactly as
    both services do today. That tx takes no advisory lock, so the constraint is satisfied by not
    introducing one.

## 5. Module wiring

**N/A** — no module is born. `approval` already exists as the 15th top-level module (ADR 0082) with
its own `domain` / `application` / `infrastructure` / `http` / `api` / `jobs` layering; this work
changes shape inside those folders.

## 6. Frameworks to reuse, not reinvent

| Primitive | Reuse commitment |
|-----------|------------------|
| `TxRunner` (`platform/db/runner.go:21`) | The unified decision service takes the tx port, as both current services do. `Do` (writable) — `authz.Require` needs a writable tx (G1). |
| `authz.SeedTxIdentity` / `authz.Require` | Single tier-2 call site. No hand-rolled capability check. |
| `tenant.FromContext` | Unchanged. |
| `problem.New` / `problem.Write` + `problem.Register` (ADR 0089) | Retired error codes are **deleted**, not aliased; new codes registered under the `approval` module namespace. Duplicate registration panics at init — that is the mechanical guard against a half-migrated error surface. |
| `httpresponse` | Unchanged. |
| `audit.RecordTx` | Governance events keep recording inside the business tx. |
| Outbox / approval lifecycle + notification enqueuers | Reused as-is; the unified outcome maps to the existing event types, with retired types deleted. |
| Idempotency store (`signoff_idemp.go`, `platform/idempotency/postgres_store.go`) | The unified endpoint keeps per-handler idempotency — it is not a chain link. `route_admin_idemp.go` is the sibling pattern. |
| `testdb` factory (`tests/integration/testdb/`) | Every new integration test uses it. Test-framework hard gate. |
| `strictjson` | Existing module-private decoder; do not import cross-module. |

No new cross-cutting concern appears, so no new platform framework is proposed.

## 7. Contract & data

**OpenAPI.** One decision endpoint replaces `POST …/review-verdict` and `POST …/signoffs`.
`…/fast-forward` must be re-derived, not left behind: R5 "Aprovar já" exists precisely because two
records had to be written in one call — with one record it may collapse into the ordinary endpoint,
which is a design question, not a foregone conclusion. Regeneration is full-repo (`swaggerSpec` is
embedded per module; partial regen is forbidden drift).

**Migration.** Expand/contract, three steps, because the data is regulated evidence and must never
be reconstructed:

1. *Expand* — create the unified decision table (or widen `approval_signoffs` with a nullable
   signature block plus the outcome enum), add constraints as `NOT VALID`, backfill every
   `approval_review_verdicts` row with full provenance preserved (`verdict_at`, comment,
   display-name snapshot, on-behalf-of), re-point `enforce_approval_sod`.
2. *Validate* — `VALIDATE CONSTRAINT`; run a row-count and per-instance parity check as a
   migration-time assertion, not a manual step.
3. *Contract* — drop `approval_review_verdicts` only after the read paths are gone. Signed rows are
   Part 11 evidence: the migration copies them, it never derives or re-signs them, and
   `content_hash` / `signature_payload` are carried verbatim.

**Route shape.** The operator's original complaint returns here as a policy choice the design must
answer explicitly rather than inherit: does `controlado` gain a "≥1 review-kind stage" floor in
`assert_route_shape`? Under the unified shape the answer may legitimately be *no* — if every
approval stage provably carries return power, an approval-only route is no longer a broken promise.
Whichever way it goes, it is a DB-line rule with a test, not a convention.

**Destructive?** Yes, deliberately: two endpoints and two DTO families are removed in one release
(legacy-fallback-extermination). The frontend's two screens converge into one decision surface in
the same change set.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory, `//go:build integration`, R1–R4 discipline
  (`scripts/check-test-discipline.sh`). Remember the integration-tag compile gap: after any seam
  signature change run `go vet -tags integration` before committing.
- **QA gates that apply:** contract (regenerated routes match the spec), authz (tier-1 rows for the
  new route + deletion of the retired rows; tier-2 per outcome; tripwire arms), DB-invariant (the
  new CHECK constraints actually reject the forbidden shapes), migration/data-integrity (Part 11
  evidence preserved byte-for-byte), docs. **N/A:** multi-tenant isolation adds no new surface
  (inherited policies re-asserted by the existing suites); async/idempotency is unchanged in kind.
- **Mechanical validation — the operator's core requirement.** The design is not done unless each of
  these is a failing-first test or a constraint:
  1. A DB constraint rejecting a decision row on an `approval`-kind stage with no signature block.
  2. A DB constraint rejecting a signature block on a `review`-kind stage.
  3. A test proving a profile holding signature power can always return for changes — the direct
     regression guard for the §0 gap.
  4. A guard proving the retired routes are absent from the spec, from `permissions.go`, and from
     the generated code (the three-surface deletion).
  5. A migration-time parity assertion on the verdict backfill.
  6. `enforce_approval_sod` proven active on the unified table.
- **Evidence shape:** `go build ./...`, `go test ./...`, `go vet -tags integration ./...`,
  `go test -tags=integration ./...`, `.\scripts\check-system-runnable.ps1`, plus live QA driving the
  unified endpoint end-to-end through the container stack, with outcomes and any bounded defers
  reported before closure.

## 9. Docs / ADR

- **Wiki:** `wiki/modules/approval.md` and `wiki/modules/approval-tech-debt.md` updated with
  `Last verified` refreshed; `wiki/architecture/api-contract.md` for the route retirement.
- **REQ IDs:** cite from `wiki/architecture/backend-target-architecture.md` at design time —
  REQ-AUTHZ-* (tier-1/tier-2 coherence, REQ-AUTHZ-5 five-surface), the contract-first REQ, and the
  DB-invariant REQ. Confirm exact IDs against the doc rather than quoting from memory.
- **ADR required: yes, mandatory.** This changes a standing policy and supersedes prior decisions:
  - the 5-rule review/approval model ratified 2026-07-10 — R1 (per-profile signature policy) and R3
    (signing ⊃ conversing) are re-expressed; R5 ("Aprovar já") may be dissolved into the ordinary
    path;
  - ADR 0082's stage-kind expression (the module promotion itself stands);
  - ADR 0087's `livre` ruling only if the route-shape rule in §7 changes — otherwise untouched;
  - ADR 0083's subject-discriminated tripwire arms, if route/cap pairs change.
  The ADR states the unified record, the outcome enum, where the signature ceremony is required, and
  the regulatory mapping (Part 11 §11.50, ISO 13485 §4.2.4) that proves unification does not weaken
  evidence.

## 10. Verdict & locked constraints

**Verdict: 🟡 Yellow** — the direction fits the architecture and improves it; proceed to design.
Yellow rather than Green for three named reasons: a mandatory ADR superseding a ratified model, a
destructive contract change crossing into the frontend, and a data migration carrying 21 CFR Part 11
evidence.

**Open hard-stops:** none. AS-1 no (no invariant violated; two are strengthened). AS-2 no (the work
replaces the patch rather than optimizing inside it). AS-3 no (`approval` owns it unambiguously).

**Locked constraints handed to brainstorming:**

1. One decision record, one service, one endpoint, discriminated by an explicit outcome enum. Two
   parallel pipelines is the defect being removed — do not design a rule that keeps both.
2. The e-signature ceremony is required by `stage_kind`, enforced by a **DB constraint**, never by
   which endpoint the caller chose.
3. "Signing includes returning for changes" becomes mechanical: on an approval-kind stage the return
   outcome is authorised by the signature capability, not by `approval.review`. A test must prove a
   signature-only profile can return.
4. 21 CFR Part 11 evidence is carried verbatim across the migration — copied, never derived, never
   re-signed. `content_hash`, `signature_payload`, `signature_method`, `signature_meaning`,
   `actor_display_name_snapshot`, `on_behalf_of_user_id` all survive.
5. Legacy is exterminated, not relaxed: retired routes, DTOs, error codes and screens are deleted in
   the same change set. No optional-for-compat fields, no aliases. ADR 0089 duplicate-registration
   panic is the guard for the error surface.
6. The document status flip stays behind `documents`' published transition rule and the existing
   cancel GUC. Approval never writes documents' rules.
7. Whatever `assert_route_shape` ends up requiring for `controlado` is a DB-line rule with a test —
   the design must answer the question explicitly rather than inherit today's silence.
8. `TestCapabilityRegistrySize` is re-verified against the merge base, not the working tree — another
   session's `approval.sla_extend` work is in flight and has already moved it.
9. Prefer **no new capability**. If one proves necessary, walk all 10 touchpoints in
   `capability-wiring.md`, including the registry-size bump.
10. Full OpenAPI regeneration only. Partial regen is forbidden drift.
