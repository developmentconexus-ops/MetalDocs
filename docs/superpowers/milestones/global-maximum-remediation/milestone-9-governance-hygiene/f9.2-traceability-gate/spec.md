# F9.2 — traceability-gate (feature spec)

> **Milestone:** M9 governance-hygiene · **Contract:** `../validation-contract.md` §2 (binding)
> **Approved:** 2026-07-06 — approved against mission.md M9 row + validation-contract §2 (operator-locked
> sources; autonomous session per mission D2). Code may start.

## Consumer contract (first)

**Consumer 1 — CI:** a job (in `.github/workflows/`) runs one command; exit 0 ⇔ every MUST-classified
REQ ID extracted from `wiki/architecture/backend-target-architecture.md` has ≥1 resolvable evidence
link. Exit ≠0 names the uncovered REQ id(s) in output. SHOULD/MAY: reported lines, never exit-affecting.

**Consumer 2 — reviewer citing REQ IDs (CLAUDE.md rule):** a committed, human-readable coverage map
(`wiki/architecture/req-traceability.md`, generated) showing per-REQ: classification, evidence kind
(test | commit | doc-annotation), pointer. Regenerable by one documented command; stale map (drift vs
regeneration) fails the gate — the map cannot rot.

**Consumer 3 — mission Terminal Acceptance (§8):** this gate joins the mission gate inventory —
POSITIVE (green on clean tree) + NEGATIVE (red on planted uncovered MUST) proofs recorded in evidence.

## Interview record (B1.5 — resolved from normative sources)

| Q | A | Source |
|---|---|--------|
| Extraction source & shape? | `- **REQ-<AREA>-<n>** … (MUST/SHOULD/MAY …)` lines in backend-target-architecture.md; 67 unique IDs, 61 MUST-line / 6 SHOULD-line today; classification is inside the trailing parens. Extraction is scripted — no hand list. | Doc grep 2026-07-06; contract §2.1 |
| Evidence kinds? | (a) `REQ-…` literal in any `*_test.go` under `internal/` or `apps/`; (b) manual map entry with commit hash; (c) satisfied-by annotation inside the REQ line's parens (e.g. "satisfied Wave 1 (F-01)"). | Contract §2.1 |
| Where do manual entries live? | `wiki/architecture/req-trace-map.yaml` (hand-maintained, schema-checked by the tool): `req`, `kind: commit|doc`, `ref`, `note`. Tool merges auto-scan + manual. | Design; contract §2.1 "committed map, regenerable" |
| Tool form? | Go tool `scripts/req-trace/` (mirrors `scripts/api-lint` precedent: testable, cross-platform, no new toolchain). Modes: default = gate (exit code); `-write` regenerates the committed report; gate fails if committed report ≠ regenerated (anti-rot). | Repo precedent |
| CI wiring? | Extend `.github/workflows/governance-check.yml` with a req-trace job/step (or a dedicated workflow if governance-check's triggers don't fit). Local command documented in the report header. | Contract §2.4 |
| Anti-gaming? | Manual `commit` entries: tool verifies hash exists when running in a full clone (`git cat-file -e`), skipping-with-warning on shallow CI; validator samples ≥5 map entries by hand. Bulk "doc" entries without a real annotation in the doc line are impossible — kind (c) is auto-derived from the doc text, not writable in the map. | Contract §2.3 |
| Does F9.2 edit the architecture doc? | No new REQs, no re-grading. Only mechanical fixes if extraction needs them (none expected — format is regular). | Milestone rabbit-hole list |

## Non-goals (mandatory)

- No new REQ IDs; no MUST↔SHOULD re-grading; no rewriting REQ text.
- No enforcement for SHOULD/MAY (report-only).
- No frontend REQ sources; backend-target-architecture.md only.
- No retrofit of test citations into test files beyond what honesty requires — coverage comes from the
  three legitimate evidence kinds; writing new tests for uncovered REQs is OUT of scope (the map's
  commit/doc evidence covers the historically-satisfied ones; a genuinely evidence-less MUST is a
  finding to surface, not to paper over).
- No blocking on map-note quality (pointer resolvability is the bar).

## Validation Gate

1. **POSITIVE:** `go run ./scripts/req-trace` on clean tree → exit 0; output + committed report
   captured in evidence.
2. **NEGATIVE:** plant `- **REQ-TST-99** Fake requirement. (MUST)` in the doc → gate exit ≠0 naming
   REQ-TST-99; revert; both outputs captured. (Negative drives the real entrypoint.)
3. **Anti-rot:** hand-edit the committed report (one char) → gate exit ≠0 (stale-report detection);
   revert. Captured.
4. **Unit tests:** `go test ./scripts/req-trace/...` green — extraction (incl. annotation kind),
   map merge, uncovered-MUST detection, stale-report detection.
5. **CI:** workflow file contains the gate invocation; YAML anchors referenced in evidence.
6. **Map integrity:** every MUST REQ appears in the report with a pointer; validator samples ≥5.
