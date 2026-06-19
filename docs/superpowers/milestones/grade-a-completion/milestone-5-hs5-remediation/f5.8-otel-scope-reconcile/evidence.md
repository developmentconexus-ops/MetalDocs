# Evidence — F5.8 OTel scope reconcile (HS-4 fix feature)

> **Status:** CLOSED 2026-06-19 · Validator FAIL reconciled by **reclassification + disclosed
> ratification**, no code reverted. Awaiting validator re-dispatch to confirm PASS.

## Disposition

The milestone-validator's C6 scope-drift FAIL on commit `61389120` is resolved by splitting the flagged
24,170-line payload into its two true parts and handling each correctly: the semconv vendor files are a
**build-prerequisite repair** (reclassified, kept), and the two `*_otel_test.go` files are **disclosed
and ratified** as additive F2.3/ADR-0036 coverage (kept, per operator decision "Ratify + re-home").

## Proof — reclassification of the semconv payload (NOT scope drift)

```
# (1) The two semconv packages were already DECLARED at the parent commit:
$ git show 3d71b3e6:vendor/modules.txt | grep -n 'semconv/v1.24.0\|semconv/v1.30.0'
316:go.opentelemetry.io/otel/semconv/v1.24.0
317:go.opentelemetry.io/otel/semconv/v1.30.0

# (2) ...but the vendored FILES were MISSING at the parent (dir absent → broken -mod=vendor):
$ git ls-tree 3d71b3e6 -- vendor/go.opentelemetry.io/otel/semconv/v1.24.0
(empty)

# (3) The commit added the files but touched NO dependency manifest (not a new dep — a vendor fill):
$ git show 61389120 --stat -- go.mod go.sum vendor/modules.txt
(empty — none of these files changed)
```

Conclusion: `61389120` materialized **required-but-missing** vendor files for transitive deps of
`otelsql v0.42.0` / `otel v1.44.0` (introduced by M2/F2.3, ADR 0036). This is HS-3-class prerequisite
hygiene (the build must be green under `-mod=vendor`), **not** new M5 observability scope. ≈23.8k of the
24.17k flagged lines fall here.

## Proof — the two otel tests are additive span-assertions (disclosed + ratified, kept)

```
$ git show 61389120 --name-only --diff-filter=A | grep otel_test.go
internal/modules/controlleddocuments/application/service_otel_test.go
internal/modules/documents/approval/application/decision_otel_test.go
```

- `service_otel_test.go` — asserts the CD create flow emits a `cd.create` span with
  `document.profile_code`, and sets span status `Error` on failure (`tracetest.SpanRecorder`).
- `decision_otel_test.go` — asserts the approval flow emits a `signoff.record` span with
  `signoff.verdict`.

Both pin **existing** F2.3 instrumentation (ADR 0036); neither adds production code or new spans. They
are green and harmless. Operator (2026-06-19) **ratified keeping** them; the M5 spec's
observability-freeze constraint now carries a single disclosed exception (see `milestone.md`). The
validator's path labels (`iam/application`, `approval/application`) were a `--stat` path-elision
misread; actual modules are `controlleddocuments` and `documents/approval`.

## Change (documentation only — no source/test/vendor file touched)

| File | Change |
|------|--------|
| `f5.8-otel-scope-reconcile/{spec,plan,evidence}.md` | This feature's investigation, decision, proof. |
| `milestone.md` | F5.8 row added; observability-freeze ratified exception + semconv reclassification note; HS-4 record; status line updated. |

## Validation Gate results (real output)

1. **Reclassification evidence** — captured above (declared-but-missing at parent; no manifest change).
2. **Milestone spec amended** — `milestone.md` carries the ratified exception, reclassification note,
   F5.8 row, and HS-4 record.
3. **Build/tests unchanged & green** — `go build ./...` → exit 0;
   `go test -count=1 ./internal/modules/controlleddocuments/application/...
   ./internal/modules/documents/approval/application/...` → both `ok` (the ratified otel tests pass).
4. **Validator re-dispatch** — pending; recorded here on return.

## Defers

None. The OTel tests are now attributed to F2.3/ADR 0036; any future move of that coverage into an
M2-lineage location is a documentation tidy, not a defer of F5.8.
