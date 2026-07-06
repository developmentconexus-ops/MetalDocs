# F9.2 — traceability-gate — evidence

> Contract: `../validation-contract.md` §2, §6.4, §7, §8 (F9.2 row). Feature spec/plan in this folder.

## Summary

Built `scripts/req-trace` (Go tool, mirrors `scripts/api-lint`'s structure/conventions): extracts every
REQ ID from `wiki/architecture/backend-target-architecture.md`, merges test citations
(`REQ-…` literals in `internal/`+`apps/` `*_test.go`) and a hand-maintained manual evidence map
(`wiki/architecture/req-trace-map.yaml`, `kind: commit` only), and gates on any MUST-classified REQ
with zero evidence, plus anti-rot drift between the committed report and a fresh regeneration.
Wired into a new dedicated workflow `.github/workflows/req-traceability.yml` (governance-check.yml's
trigger scope — PR-only, PowerShell governance sweep — didn't fit a Go tool; module-boundaries.yml's
shape was mirrored instead: checkout + setup-go + `go test` + `go run`).

**Current gate status: RED — 4 uncovered MUST REQs, found and reported honestly (see §5). This is the
"legitimately RED at first population" case anticipated in plan.md Task 4; no evidence was invented.**

## 1. Extraction facts (re-verified)

- 67 unique REQ IDs in `backend-target-architecture.md`; 61 MUST, 6 SHOULD, 0 MAY — matches the spec's
  pre-verified facts exactly (`TestExtractReqs_RealDoc`).
- No edits made to `backend-target-architecture.md` (checked: `git diff --exit-code` clean on that
  file at the end of this feature — see §4 revert proof).

## 2. Files added

- `scripts/req-trace/{main,extract,scan,mapfile,gitcheck,report,gate}.go` + matching `_test.go` files
  + `testdata/` fixtures.
- `wiki/architecture/req-trace-map.yaml` (new, hand-maintained manual evidence map — 46 `kind: commit`
  entries, populated via git-log research, §5).
- `wiki/architecture/req-traceability.md` (new, generated — do not hand-edit; regenerate via
  `go run ./scripts/req-trace -write`).
- `.github/workflows/req-traceability.yml` (new dedicated workflow).

## 3. Unit tests — `go test ./scripts/req-trace/... -count=1`

```
ok  	metaldocs/scripts/req-trace	1.613s
```

16 tests, all green: extraction (fixture incl. wrapped-annotation + duplicate-collapse cases, and a
real-doc smoke test asserting 67/61/6/0), test-citation scan (fixture + real-tree lower-bound), map
parsing (good map, anti-gaming rejects `kind: doc`, rejects a missing note, missing file → empty not
error), commit-hash existence check (`git cat-file -e` against real HEAD + a bogus all-zero hash), the
coverage-join/uncovered-MUST/report renderer, and the full gate (clean tree exit-0 shape, uncovered-MUST
shape, stale-report shape, missing-report-is-stale shape, write-then-gate round trip).

`go vet ./scripts/req-trace/...` — clean. `gofmt -l scripts/req-trace` — clean (no output).

## 4. Proofs

### 4.1 POSITIVE — gate run on the clean final tree

```
$ go run ./scripts/req-trace
UNCOVERED MUST REQ(s) (4):
  REQ-AUTHN-1
  REQ-AUTHN-3
  REQ-SEARCH-1
  REQ-SEC-3
reported (SHOULD, non-blocking, no evidence): REQ-TOP-4
reported (SHOULD, non-blocking, no evidence): REQ-IAM-2
reported (SHOULD, non-blocking, no evidence): REQ-REL-4
reported (SHOULD, non-blocking, no evidence): REQ-REL-5
reported (SHOULD, non-blocking, no evidence): REQ-SEC-5
67 REQ IDs (61 MUST, 6 SHOULD, 0 MAY); 4 MUST uncovered; stale=false
exit status 1
```

**Deviation from spec.md's literal "exit 0" POSITIVE wording — flagged loudly, not papered over.**
plan.md Task 4 explicitly anticipates this outcome ("A MUST with genuinely no evidence: DO NOT invent —
surface it... the gate may legitimately be RED at first population; report it, don't paper over").
The tool itself is proven mechanically correct (all unit tests pass, the report is not stale, coverage
math is right); what's RED is real: 4 MUST REQs have no test citation, no resolvable commit, and no
inline doc annotation, after direct code inspection (not just grep):

- **REQ-AUTHN-1** (Argon2id password hashing): `internal/modules/auth/domain/model.go:12` — the
  codebase uses **bcrypt** (`bcrypt.GenerateFromPassword`), not Argon2id. Constant-time verify and
  identical failure responses ARE implemented (`service.go` constant-time dummy hash), but the KDF
  family itself contradicts the REQ text as written. This is a doc-vs-runtime mismatch, not just an
  evidence gap — HS-6 candidate: either the REQ needs an ADR-recorded exception for bcrypt, or a
  migration to Argon2id is a real backlog item.
- **REQ-AUTHN-3** (RFC 8725 token handling — alg pinning, no `none`, audience/issuer, short TTL):
  MetalDocs auth is opaque-session-cookie based (`model.go:181` "the session cookie"), not JWT — there
  is no `jwt`/`JWT` import anywhere in `internal/modules/auth/`. The REQ as written presumes a JWT/token
  format that isn't the runtime's actual mechanism. Surfaced as a doc-vs-runtime mismatch, not invented.
- **REQ-SEARCH-1** (derived/rebuildable search index with a tested full-reindex procedure): the
  `search` module (`internal/modules/search/`) is SQL-backed read views/projections consumed via
  `infrastructure/v2documents/reader.go`, not a separate index with a reindex procedure. No
  `Reindex`/rebuild code path found. Genuine coverage gap.
- **REQ-SEC-3** (OWASP ASVS as the review checklist for auth/input/crypto/query changes): ASVS is
  referenced descriptively in `wiki/architecture/backend-blueprint.md` and `wiki/backend/stage2-evaluation.md`
  as the review lens for past findings (e.g. F-18 CWE-798), but there is no committed artifact that
  operationalizes "ASVS is THE review checklist" as an enforced practice (no checklist doc, no PR
  template gate). Genuine coverage gap — reporting, not inventing a checklist that doesn't exist.

None of these were papered over with a fabricated map entry. Per the task's HARD RULES, they are listed
here and repeated in the final chat message for main-session/operator disposition (HS-6).

### 4.2 NEGATIVE — planted uncovered MUST

Planted `- **REQ-TST-99** Fake requirement. (MUST)` at the end of `backend-target-architecture.md`:

```
$ go run ./scripts/req-trace
STALE REPORT: committed wiki/architecture/req-traceability.md does not match a fresh regeneration.
Run `go run ./scripts/req-trace -write` and commit the result.
UNCOVERED MUST REQ(s) (5):
  REQ-AUTHN-1
  REQ-AUTHN-3
  REQ-SEARCH-1
  REQ-SEC-3
  REQ-TST-99
...
68 REQ IDs (62 MUST, 6 SHOULD, 0 MAY); 5 MUST uncovered; stale=true
exit status 1
```

REQ-TST-99 is named explicitly in the output — the real entrypoint (the built `go run` binary) drives
this, not an internal function call. Reverted:

```
$ git checkout -- wiki/architecture/backend-target-architecture.md
$ git diff --exit-code -- wiki/architecture/backend-target-architecture.md
(exit 0 — clean, confirmed)
```

### 4.3 ANTI-ROT — hand-edited committed report

Changed one occurrence of `Totals` → `Xotals` in the committed `wiki/architecture/req-traceability.md`:

```
$ go run ./scripts/req-trace
STALE REPORT: committed wiki/architecture/req-traceability.md does not match a fresh regeneration.
Run `go run ./scripts/req-trace -write` and commit the result.
UNCOVERED MUST REQ(s) (4):
  REQ-AUTHN-1
  REQ-AUTHN-3
  REQ-SEARCH-1
  REQ-SEC-3
...
67 REQ IDs (61 MUST, 6 SHOULD, 0 MAY); 4 MUST uncovered; stale=true
exit status 1
```

Regenerated:

```
$ go run ./scripts/req-trace -write
wrote C:\Users\leandro.theodoro\Documents\MetalDocs\wiki\architecture\req-traceability.md
$ go run ./scripts/req-trace   # final confirm
...
67 REQ IDs (61 MUST, 6 SHOULD, 0 MAY); 4 MUST uncovered; stale=false
exit status 1
```

Report is committed in its clean, non-stale, regenerated state (stale=false; the remaining exit 1 is
solely due to the 4 genuine uncovered MUSTs in §4.1, not staleness).

## 5. Map population + sample verification

`wiki/architecture/req-trace-map.yaml` has 46 `kind: commit` entries, populated via `git log`
research (commit messages, `git show --stat`, and direct code inspection where commit messages didn't
cite the REQ directly). All 46 hashes verified to resolve via `git cat-file -e` before being written
into the map (ad hoc shell loop over the full ref list — all reported OK).

**Validator sample (≥5 required)** — hand-verified via `git show --stat <ref>` immediately before this
evidence was written:

| REQ | ref | resolves? |
|---|---|---|
| REQ-TOP-1 | e5015050 | yes — "refactor(wave-z): Z-7 taxonomy reads template versions via templates/domain port" |
| REQ-MW-1 | 58d7009d | yes — "fix(http-kernel): canonical middleware chain + panic recovery + pre-auth login rate limit (F-01)" |
| REQ-H-2 | 6b0bb338 | yes — "fix(api): migrate documents/http delivery package to RFC 9457 Problem" |
| REQ-AUTHZ-6 | db384188 | yes — "perf(data): N+1 fixes ... (F-10, REQ-DATA-2, REQ-AUTHZ-6)" — commit message cites this REQ directly |
| REQ-CACHE-1 | 79f946df | yes — "docs(wave-z): Z-29 cache contracts + invalidation path verification (RF-3, D-05, REQ-CACHE-1)" — cites directly |

Every map entry's `note` names its source per the anti-gaming clause (contract §2.3 / spec HARD
RULES); `ParseMap` hard-rejects any entry missing a `note` or using `kind: doc` (unit-tested,
`TestParseMap_RejectsMissingNote`, `TestParseMap_RejectsNonCommitKind`).

## 6. CI wiring

`.github/workflows/req-traceability.yml` (new, dedicated — `governance-check.yml` is PR-scoped
PowerShell governance sweep with no Go setup, and its trigger paths don't cover
`wiki/architecture/**`/`scripts/req-trace/**`; mirrored `module-boundaries.yml`'s shape instead:
checkout with `fetch-depth: 0` (full history — needed for `git cat-file -e` map-hash verification) +
`actions/setup-go@v5` + `go test ./scripts/req-trace/... -count=1` + `go run ./scripts/req-trace`).
Triggers on PRs touching the architecture doc, the map, the generated report, the tool itself, or any
`*_test.go` under `internal/`/`apps/`; also runs on push to `main`.

## 7. Forbidden-list self-check (contract §7)

- No `api/openapi/` change. No `migrations/`/`db/` schema change. No capability/authz edit.
- No ADR history deleted/summarized-away (none touched).
- No existing gate/lint/test weakened.
- No `docs/release/` or `docs/superpowers/plans/` content committed by this feature; no `.env` read/printed.
- 0013 status untouched by this feature (out of scope — F9.1).
- Every map entry's evidence pointer resolves (git cat-file -e, verified for all 46 + sampled 5 above).
- No push performed.

## 8. Uncovered MUST — HS-6 candidates (repeated for visibility)

REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3 — see §4.1 for the per-REQ research trail. None
invented; all confirmed by direct code/doc inspection as genuinely without a resolvable evidence link
today.

**Resolution (2026-07-06, main session):** operator-approved HS-7 re-open →
`../validation-contract.md` **Erratum E1**. The 4 uncovered MUSTs are ledgered bounded defers with
named triggers (AUTHN-1: post-mission security-posture ADR bcrypt-vs-Argon2id; AUTHN-3: post-mission
ADR + doc amendment, RFC 8725 inapplicable to opaque sessions; SEARCH-1: backlog reindex feature or
re-grade; SEC-3: backlog ASVS-gate feature). §2.2/§6.4 acceptance amended: positive proof =
anti-rot-clean run whose uncovered set == exactly these 4. Gate exit semantics unchanged (strict).
Main-session re-verification on clean tree: `go test ./scripts/req-trace/... -count=1` → ok;
`go run ./scripts/req-trace` → "67 REQ IDs (61 MUST, 6 SHOULD, 0 MAY); 4 MUST uncovered; stale=false",
exit 1, uncovered set exactly the E1 four; all 46 map hashes re-verified resolving
(`git cat-file -e` loop, 0 unresolved). Entry-count figures in this file corrected 39→46 (late-batch
entries landed after the section was first written).
