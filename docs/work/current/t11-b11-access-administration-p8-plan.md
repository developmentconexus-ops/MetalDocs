# T11 — B11 Access Administration — P8 Functional Evidence Plan

> **For agentic workers:** execute inline against the approved B11 P6/P7 specification; this is temporary frontend-planning Evidence, not Product implementation.

**Goal:** Build one browser-operable low-fidelity B11 prototype that makes the approved Access workspace operable and actively falsifies B11-A1 findability, B11-A2 membership-consequence sufficiency, and B11-A3 access-explanation sufficiency.

**Architecture:** One self-contained HTML/CSS/vanilla-JS artifact under `docs/work/current/`. Deterministic local fixtures simulate only accepted reads/writes and material failures. The browser never calculates complete effective access, never fabricates global search/filter completeness, and never introduces production dependencies.

**Tech Stack:** HTML5, CSS, vanilla JavaScript, deterministic fixtures, Chromium/Playwright verifier outside the repository.

**Spec:** `docs/work/current/t11-b11-access-administration-p6-p7.md`

## Global constraints

- Preserve the accepted B01/B01N shell and stable `/admin/access` route.
- B11 local structure remains `Memberships` + `Role grants`.
- Primary B11 operations are exactly 27–33; B10 Organization reads may appear only as supporting identity reads.
- No custom Role/Permission editor, effective-permission engine, invented search/filter/explanation API, or client-side fake global search.
- `Subject × Role × Scope` must be explicit before `createRoleAssignment` and repeated in a review state.
- Membership add/remove and grant revoke copy must not claim removal/addition of all effective access.
- `createRoleAssignment` ambiguous transport outcome must preserve one logical command / one Idempotency-Key and expose safe same-command retry.
- P8 is disposable Evidence only; no Product/frontend runtime implementation.
- B12, FP2/P11, T12 and Product implementation remain blocked.

---

## Task 1 — Test-first verifier

**Files:**
- Create outside repository: `/tmp/b11_p8_check.py`
- Target outside repository during RED/GREEN: `/tmp/t11-b11-access-administration-p8.html`

- [ ] Write a Playwright verifier that first requires the artifact to exist and then checks real DOM behavior for route/shell, local navigation, pagination, membership add/remove, grant composition/review/create/revoke, safe consequence language, ambiguous create retry, failure scenarios, accessibility focusability and narrow-viewport local navigation.
- [ ] Run before the HTML exists and require the expected RED failure `artifact absent`.
- [ ] Keep the verifier outside the repository; it is a proof harness, not durable Product code.

## Task 2 — Minimum P8 structure

**File:**
- Create: `/tmp/t11-b11-access-administration-p8.html`

- [ ] Build inherited shell/chrome with `/admin/access` active.
- [ ] Add Evidence-only header/scenario controls clearly outside Product semantics.
- [ ] Add local `Memberships` and `Role grants` navigation.
- [ ] Add deterministic scale that requires pagination for Groups, Users, Areas and RoleAssignments.
- [ ] Do not render any search/filter input that could be mistaken for global collection search.

## Task 3 — Memberships behavior

- [ ] Group browse/select using deterministic paginated fixtures.
- [ ] Current-member browse using accepted `GroupMemberPage` semantics.
- [ ] Add-member flow uses supporting paginated User truth, makes disabled users unavailable, and confirms only bounded consequence.
- [ ] Remove-member flow targets one exact membership and states that other direct/group access may remain.
- [ ] Include a security-bearing Group fixture to exercise B11-A2 without calculating its complete effective permission result.

## Task 4 — Role grants behavior

- [ ] Paginated assignment ledger, no fake completeness/filtering.
- [ ] Grant composer: select USER|GROUP subject, fixed server-returned Role, COMPANY|AREA scope admitted by `RoleView.allowed_scope_kinds`.
- [ ] Role explanation uses only fixture `RoleView.permissions` corresponding to the accepted fixed role vocabulary.
- [ ] Review repeats Subject × Role × Scope and says the grant is additive.
- [ ] Successful grant creates one fixture assignment and returns to the ledger.
- [ ] Revoke targets one exact assignment and states that other access may remain.
- [ ] No edit-assignment abstraction.

## Task 5 — Failure/retry and falsification evidence

- [ ] Operable 403, 404, 409, 422 and ambiguous-create scenarios.
- [ ] Ambiguous `createRoleAssignment` state exposes `Retry same command`; retry preserves the same logical command/key and creates at most one assignment.
- [ ] Evidence tasks explicitly ask the operator to locate later-page User/Group/Area/assignment without fake search (A1), judge membership consequence copy (A2), and judge whether canonical grants can be administered safely without a complete explanation surface (A3).
- [ ] The product surface itself never claims complete effective access or “why access” truth.

## Task 6 — Accessibility/responsive proof

- [ ] Semantic buttons/forms/dialogs, visible focus, deterministic focus entry/return where practical, non-color-only state.
- [ ] At narrow width, global shell and B11 local nav remain operable; list/detail and grant composer stack vertically.
- [ ] No material security action depends on hover.

## Task 7 — Verification and publication

- [ ] Run the real-DOM verifier against the unchanged local HTML until all checks pass.
- [ ] Compute SHA-256 and Git blob identity for the exact verified bytes.
- [ ] Publish those exact bytes to `docs/work/current/t11-b11-access-administration-p8.html` on the B11 branch.
- [ ] Confirm the remote Git blob matches the locally computed Git blob.
- [ ] Update roadmap to `P8 CANDIDATE / OPERATOR REVIEW` with verification count and A1/A2/A3 still falsifiable.
- [ ] Stop for operator walkthrough; only the operator may disposition LOCK / REVISE / UPSTREAM FINDING.
