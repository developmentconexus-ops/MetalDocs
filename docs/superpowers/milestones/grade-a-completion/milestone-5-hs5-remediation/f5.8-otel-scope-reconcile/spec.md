# Feature F5.8 — OTel scope reconcile (HS-4 fix feature from milestone-validator FAIL)

> **Milestone:** 5 — HS-5 remediation · **Feature:** `f5.8-otel-scope-reconcile`
> **Status:** **APPROVED 2026-06-19** — opened under HS-4 after the milestone-validator returned
> `VERDICT: FAIL` (C6 scope-drift). Operator decision (2026-06-19): **Ratify + re-home (keep)** — no
> code reverted. Documentation/reconciliation feature only.

## Trigger (validator finding)

The milestone-validator FAILed M5 on C6/C3: commit `61389120` ("F5.1+F5.2+F5.3+F5.4") bundled
**24,170 insertions** of out-of-scope work — vendored `go.opentelemetry.io/otel/semconv/v1.24.0` +
`v1.30.0` plus two `*_otel_test.go` files — with no rationale in the commit message or any M5 artifact,
against the M5 spec's "Do not touch composition/observability (already A−)" constraint.

## Investigation — the 24,170 lines are two different things, only one is scope

| Payload | Lines | Classification | Evidence |
|---------|-------|----------------|----------|
| `vendor/go.opentelemetry.io/otel/semconv/v1.24.0/*`, `v1.30.0/*` | ~23,800 | **NOT scope drift — build-prerequisite vendor repair.** At the parent commit `3d71b3e6`, `vendor/modules.txt` **already declared** both packages (lines 316–317) as transitive deps of `otelsql v0.42.0` / `otel v1.44.0` (M2/F2.3, ADR 0036), but the vendored **files were missing** → broken vendor tree, `go build -mod=vendor` fails. The commit materialized required-but-missing files. | `git show 3d71b3e6:vendor/modules.txt` lists v1.24.0/v1.30.0; `git ls-tree 3d71b3e6 -- …/semconv/v1.24.0` is **empty** (dir absent); commit `61389120` touched **no** `go.mod`/`go.sum`/`modules.txt`. |
| `controlleddocuments/application/service_otel_test.go` (+179), `documents/approval/application/decision_otel_test.go` (+167) | ~346 | **Genuinely out of M5 scope — observability test coverage (M2/F2.3, ADR 0036).** Net-new span-assertion tests (`cd.create`, `signoff.record` spans via `tracetest`). Additive, green, harmless — but they belong to the observability lineage M5 said not to touch, and were bundled undisclosed. | File bodies assert OTel spans/attributes; no M5 feature requires them. |

> The validator labeled the test paths as `iam/application` / `approval/application`; the actual paths
> are `controlleddocuments/application` / `documents/approval/application` (a path-elision misread in
> the `--stat` output). The substance of the finding — undisclosed observability work in an M5 commit —
> is correct; the magnitude (24k) is not (≈99% is required vendor repair, not scope).

## Decision (operator, 2026-06-19): Ratify + re-home (keep)

1. **Semconv vendor files:** reclassify as a **build prerequisite repair** (HS-3 class), not scope
   drift. Disclose in this evidence; no revert (removing them re-breaks `-mod=vendor`).
2. **The two `*_otel_test.go` files:** **retained** (green, additive coverage of ADR-0036
   instrumentation), **disclosed**, and **attributed to M2/F2.3 lineage**. The M5 spec's
   observability-freeze constraint is amended with a single **ratified, disclosed exception** for this
   bundled coverage. No new observability *production* code is in M5 — only tests pinning existing
   F2.3 spans.

## Non-goals

- No code reverted; no test deleted; no vendor file removed.
- No new observability production code, no new spans, no instrumentation changes.
- No git history rewrite of `61389120` (forward-only disclosure via this feature + milestone amendment).

## Validation Gate

1. **Evidence recorded** — this feature's `evidence.md` documents the reclassification with the exact
   git commands proving (a) semconv was declared-but-missing at parent, (b) the commit touched no dep
   manifest, (c) the two test files are additive span-assertions.
2. **Milestone spec amended** — `milestone.md` carries (a) the ratified disclosed-observability-test
   exception, (b) the semconv reclassification note, (c) an F5.8 row + HS-4 record.
3. **Build/tests still green** — `go build ./...` clean; `go test -count=1 ./...` 0 FAIL (unchanged —
   no code touched).
4. **Re-dispatch validator** — milestone-validator re-run returns `VERDICT: PASS` (C6 satisfied: the
   only remaining "drift" is disclosed + ratified; the 24k payload is reclassified as prerequisite
   repair with evidence).
