# ADR 0087 — Livre governance class: configured no-approval route

> **Status:** Accepted 2026-08-03 (operator ruling 2026-07-29: "livre seria uma
> rota de aprovação que não precisa de aprovação — ele estaria configurado
> como livre, não desconfigurado"; ratified by operator 2026-08-03)
> **Supersedes:** the livre arm of [ADR 0081](0081-per-profile-signature-policy.md)
> (`RoutePolicyNoRoutePermitted` / migration 0295's "no active route may exist
> for a livre profile" — now folded into the baseline).
> **Extends:** [ADR 0086](0086-template-routes-config-first-profile-keyed.md)
> (config-first universal — this ADR closes its open item "livre profiles can
> never own templates"), [ADR 0082](0082-approval-kernel-extraction.md)
> (subject-generic kernel unchanged), [ADR 0085](0085-release-coordinator-approval-driven-publication.md)
> (post-approval publication pipeline is reused as-is).
> **Scope:** design only; implementation is a separate unit gated on this
> ADR's acceptance. System-impact gate: Yellow
> (`docs/superpowers/analysis/2026-08-03-livre-route-class-system-impact.md`).

## Context

`governance_class='livre'` today means **no route permitted**: the taxonomy
domain derives `RoutePolicyNoRoutePermitted`
(`internal/modules/taxonomy/domain/governance_class.go:61-62`) and the
DEFERRABLE both-directions policy trigger (ex-migration 0295, folded into the
baseline 2026-07-29) rejects any active route on a livre profile.

Since config-first became universal, that stance is self-contradictory:

- **Documents** hard-block creation without an active route (D2 gate) — so a
  livre profile cannot own documents at all.
- **Templates** (ADR 0086) hard-block creation on `has_active_template_route`
  — so a livre profile can never own templates either.
- Submit resolution binds only active routes and fails for livre
  (`internal/modules/approval/application/submit_service.go:204-216`).

"Livre" therefore degenerated from "ungoverned material" into "profile that
can own nothing". The operator ruled the opposite model: livre is a profile
whose route is **explicitly configured to require no approval**. Absence of a
route is *always* misconfiguration, for every governance class.

## Decision

1. **Config-first is universal, no exceptions.** Every profile — controlado,
   simples, livre — requires an explicit active route per subject kind before
   it can own documents/templates. Creation gates are unchanged; they simply
   see an active route.

2. **A livre route is a configured zero-stage route.** Its shape is derived,
   not flagged: `stages = []` on a route whose profile is
   `governance_class='livre'`. No `auto_approve` boolean exists anywhere —
   redundant state that can drift is forbidden (no-fallback principle).

3. **RoutePolicy vocabulary.** `RoutePolicyNoRoutePermitted` is replaced by
   `RoutePolicyNoApprovalStages`: an active route is required and MUST carry
   zero stages. The three-way published language becomes:
   - `controlado` → `require_approval_stage` (≥1 approval-kind stage) — unchanged
   - `simples` → `approval_optional` (review-only or staged) — unchanged
   - `livre` → `no_approval_stages` (route required, zero stages)
   `GovernanceClass` still never crosses the taxonomy→approval boundary; only
   the policy value does. Fail-closed default remains
   `require_approval_stage` on unknown/unset class.

4. **DB is the authoritative line — forward migration 0316** (first
   post-fold forward migration; baseline is never edited):
   - `assert_route_shape` livre arm: active route on a livre profile MUST have
     zero stages (any stage ⇒ P0001); replaces "no active route may exist".
   - Zero-stage is **DB-exclusive to livre**: controlado AND simples active
     routes MUST carry ≥1 stage at the DB line (previously an app-only
     structural floor). A zero-stage route auto-approves — that shape may
     never exist on a governed class.
   - Both directions preserved, DEFERRABLE INITIALLY DEFERRED preserved:
     route writes AND profile reclassification re-validate
     (reclassify controlado→livre with a staged active route ⇒ P0001;
     livre→controlado with a zero-stage active route ⇒ P0001).
   - No data purge needed (livre profiles have zero routes today by
     construction). If any purge is ever added, the deferred policy triggers
     must be disabled around it (SQLSTATE 55006 lesson, ADR 0086).

5. **Submit = normal path, instantly complete.** No side door. Submitting a
   subject bound to a livre (zero-stage) route runs the standard submit
   service: instance created, and — having no stages to satisfy — transitions
   to `approved` in the same transaction, with a governance event
   (`governance_events`, the approval module's canonical in-tx event log —
   same vocabulary and store as `approval_submitted`/`signoff_recorded`;
   auto-approval is recorded exactly as much as a normal approval, no more,
   no less) carrying the route/version that authorized it. This applies to
   **both subjects** — documents and templates (ADR 0086 template routes on a
   livre profile are zero-stage too; template submit auto-approves the
   template version through its normal completion path).
   Existing submit idempotency semantics apply unchanged. Post-approval
   behavior (release coordinator, publication, ADR 0085) is exactly that of a
   normally-approved instance. No new capability is minted; the submit
   capability set is unchanged — livre-ness is a route shape, not a
   permission.

6. **Route admin.** The write boundary (`Validate(policy)` off-tx) accepts a
   zero-stage route iff the profile policy is `no_approval_stages`, and
   rejects adding stages to it with a dedicated problem+json code. Creating,
   versioning, activating and deactivating livre routes uses the existing
   route-admin surface — a livre route is a first-class configuration object
   with the same audit trail.

## Consequences

- Livre profiles can own documents and templates through the same gates as
  everyone else; onboarding is uniform ("configure a route" is the single
  universal step).
- The approval kernel keeps one code path; reviewers reason about one
  lifecycle. Auto-approved instances are ordinary approved instances with an
  audit event and empty signoff set.
- Contract surface change is minimal: stages already serialize as a list;
  expected additions are problem codes (stages on a livre route; possibly a
  distinct submit response nuance) — all via OpenAPI + full regen.
- Tests: both-directions trigger matrix, submit instant-approve integration
  (incl. idempotent retry), creation-gate flip for livre profiles, contract
  tests for new problem codes. `testdb` factory, `//go:build integration`.

## Rejected alternatives

- **Keep livre route-less and special-case the gates** ("livre skips
  config") — reinstates a bypass branch in every gate and in submit; the
  exact non-uniformity the ruling kills.
- **`auto_approve` flag on routes** — redundant with (profile class × zero
  stages); two sources of truth that can disagree.
- **Auto-provision a hidden system route for livre profiles** — invisible
  configuration; violates "absence of route = misconfiguration" observability
  and the no-fallback principle.
