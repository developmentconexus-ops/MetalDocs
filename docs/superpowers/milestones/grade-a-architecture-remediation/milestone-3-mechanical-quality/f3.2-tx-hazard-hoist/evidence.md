# Feature F3.2 — Evidence (verify-already-done)

> **Milestone:** 3  ·  **Feature:** `f3.2-tx-hazard-hoist`  ·  **Closed:** 2026-06-14
> **Contract:** `milestone.md` F3.2 row (amended). The H-PRE-1 deadlock root cause was already hoisted
> off-tx in a prior wave; this row records that the deadlock-bearing read is off-tx and that the
> residual in-tx reads are non-authz (no hazard). Memory: `[[advisory-lock-deadlock-constraint]]`.

## What was implemented

Nothing. F3.2 is a verify-already-done row. The genuine hazard — an authz-recording taxonomy read
(`GetByCode` via `ensureTemplateArtifact`) running inside the audit-hash-chain advisory-lock tx — was
already moved to a pre-flight off-tx read. The spec's literal target lines (`service.go:308,331`) now
hold **plain non-authz SELECTs**, which are not an H-PRE-1 hazard.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Deadlock read is hoisted off-tx | Read `controlleddocuments/application/service.go:255–292` | Auto path pre-flights `ensureTemplateArtifact` + `ResolveTemplateVersionID` **before** `s.runner.Do` opens the tx (line 292). Explicit comment (`:278–283`): "Pre-flight OFF-TX … running them inside the tx — which holds the audit hash-chain advisory lock once authz.Require records the system_admin bypass — self-deadlocks." | real |
| Residual in-tx reads are non-authz | Read `controlleddocuments/infrastructure/repository.go:702` (`GetTemplateVersionState`) and `:72` (`CodeExists`) | `GetTemplateVersionState` = `SELECT v.status, t.doc_type_code FROM templates_template_version …` on `c.db`; `CodeExists` = `SELECT EXISTS(… controlled_documents …)` on `r.db`. **Neither calls `authz.Require`/`GetByCode`** → no audit-lock acquisition → no self-deadlock. | real |
| Runtime proof of no deadlock (prior wave) | Audit H-PRE-1 runtime evidence | `POST /api/v1/controlled-documents` (system_admin) returned in **83 ms**; `pg_locks` advisory count = **0** before and after — previously a 60–90 s hang then 500. Source: `wiki/backend/_artifacts/architecture-audit-2026-06-13.md` line 109. | real (prior-wave runtime) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from milestone.md F3.2) | Met? | Evidence |
|-------------------------------------|------|----------|
| H-PRE-1 deadlock read is off-tx (comment + runtime proof) | yes | rows 1, 3 |
| Residual in-tx `GetTemplateVersionState`/`CodeExists` shown non-authz → no hazard | yes | row 2 |
| Port refactor of `GetTemplateVersionState` deferred to M4 F4.2 | yes | recorded in defers below + governing spec F4.2 |

## Review disposition

- Spec-compliance review: n/a — no code change; reconciliation reviewed at HS-6 operator gate.
- Code-quality review: n/a — no code change.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `GetTemplateVersionState` reach-without-a-port → `TemplateVersionStateReader` | Not a deadlock concern (read is non-authz); it is the H-G cross-module-reach concern owned by M4 F4.2, which already names this exact site and requires the read stay off the lock-holding tx | M4 F4.2 (governing spec §6); owner: backend agent at M4 |
