# Evidence — unit blank-docx Option B (template docx read-path honesty)

**Date:** 2026-07-16 · **Branch:** `claude/determined-driscoll-2e3140` · **Base:** `4a91aadf`
(worktree forked from main HEAD; dispatch named base `60d0bf12` — `4a91aadf` is the docs-only
roadmap commit directly on top of it, contains it)
**Charter:** `docs/superpowers/analysis/2026-06-29-blank-template-docx-provision-system-impact.md` (Option B, operator-locked)

## Anchor re-verification (charter predates 3.1a/4.2)

| Charter anchor | Verified location 2026-07-16 |
|---|---|
| `queries.go:46` GetDocxURL | `internal/modules/templates/application/queries.go:59` |
| `verified_store.go:113` Exists | `internal/platform/objectstore/verified_store.go:235` (StatObject, unguarded DB-sourced-key trust model, same as PresignGet) |
| `errors.go:52` ErrUploadMissing/CodeUploadMissing | `domain/errors.go:13` + `delivery/http/errors.go:67` (→409, pinned by errors_test.go:28) |
| `keys.go:8` templateDocxKey | `application/keys.go:17` |
| `autosave.go:110` | `application/autosave.go` CommitAutosave→Confirm path intact |
| create sets key | `application/create.go:63` |

Intent unchanged; drift is line-number only + one structural fact: production Presigner is
`*objectstore.VerifiedStore` injected directly (main.go:567→1065), so adding `Exists` to the
templates `Presigner` port needs zero adapter code; two test fakes gain the method.

## Model-matrix corrections (hub, 2026-07-16)

- Correction #1 (superseded): codex retired all roles; dual gate = Opus + sonnet. Rows 2–8 ran
  under this ruling.
- Correction #2 (binding, profile 0dd1381e): retirement NARROWED to planner+implementer; planner
  = Opus; **dual gate = cold Opus + GPT-5.6 Sol medium via codex** (gate-only codex). Per hub
  rule, this unit's sonnet-produced plan (row 2) STANDS — already consumed + reviewed. The
  completed Opus+sonnet dual gate (rows 7–8) ran under correction #1 while it was binding; the
  §9 delta re-review of the contract-lock commit runs on the corrected matrix (rows 9–10).

## Dispatch ledger

| # | Role | Model / effort | Path | Prompt pack | Output artifact | Verdict/result |
|---|---|---|---|---|---|---|
| 1 | Feature planner | gpt-5.6-sol / medium | OS-process codex exec | scratchpad `prompt-planner.md` | — | KILLED mid-run — hub correction 2026-07-16 (profile §10 amendment 1070e94c): codex retired for MetalDocs, Claude-only matrix; no output consumed |
| 2 | Feature planner | claude sonnet subagent | Agent tool (sync) | scratchpad `prompt-planner.md` (same pack) | scratchpad `plan-blank-docx-b.md` | DONE — 2 slices (A code+tests, B wiki); contract finding: docx-url route lacks declared 409 (openapi.yaml:1731-1753) → REQUEST contract-lock sent to hub |
| 3 | Implement worker (slice A) | claude sonnet subagent | Agent tool (sync) | scratchpad `plan-blank-docx-b.md` slice A | commit `291bce1c` | DONE — TDD (failing test first confirmed), 5 new GetDocxURL tests, build/vet/gofmt/module tests green |
| 4 | Per-slice reviewer (slice A) | claude sonnet subagent | Agent tool (sync) | REVIEW-STANDARD §14 prompt-pack + diff 291bce1c + L0 report | verdict below | APPROVE, zero findings (G1-G3 clean; fail-closed, cross-tenant, no-test-theater receipts verified) |
| 5 | Implement worker (slice B, wiki) | claude sonnet subagent | Agent tool (sync) | scratchpad `plan-blank-docx-b.md` slice B | commit `ac9b8391` | DONE — templates.md §8.9a + Last-verified + anchor fix (queries.go:41→:51); tech-debt T-016 appended-closed |
| 6 | Per-slice reviewer (slice B) | claude sonnet subagent | Agent tool (sync) | §14 prompt-pack + diff ac9b8391 | verdict below | APPROVE — all anchors/claims verified exact vs code; honesty + idiom checks pass; 1 trivial non-blocking (openapi range 1731-1753 vs actual 1751) |
| 7 | Dual gate — Claude arm | COLD Opus subagent (model=opus, clean context) | Agent tool (sync, parallel with #8) | gate prompt-pack, range 4a91aadf..ac9b8391 | verdict below | APPROVE — 1 suggestion + 1 nit, both optional |
| 8 | Dual gate — second arm | independent claude sonnet subagent (clean context) | Agent tool (sync, parallel with #7) | same gate prompt-pack, same fixed SHA | verdict below | APPROVE — zero findings |
| 9 | §9 delta re-review — Claude arm | COLD Opus subagent (model=opus, clean context) | Agent tool (sync) | scratchpad `gate-delta-prompt.md`, commit 44a1960c | verdict below | APPROVE — zero findings, 4/4 checks pass with receipts |
| 10 | §9 delta re-review — GPT arm | gpt-5.6-sol / medium via codex (corrected matrix, gate-only) | OS-process codex exec (stdin closed, teed log) | same `gate-delta-prompt.md` | scratchpad `agent__gate-delta-gpt.last.md` + `.log` | APPROVE — zero findings; independent SHA-256 marker-hash proof that 9 modules' non-swaggerSpec content is byte-identical |

## Ladder results

- **L0:** `go build ./...` clean · `go vet ./...` clean · gofmt clean · api-lint `-strict` 0 violations ·
  module-boundaries OK. Pre-existing (NOT this diff, verified file-disjoint): check-test-discipline
  7 violations (approval module ×4, tests/integration/approval ×1, tests/integration/migrations ×2);
  cilint hgcrossmodule 4 hits (approval/auth/iam + templates/infrastructure/postgres.go:731 — untouched file).
- **L1 unit:** `go test ./...` full sweep — zero FAIL.
- **L1 integration (selective per profile §2; pg_stat_activity window verified clear):**
  `.\scripts\test-integration.ps1` templates pkg + guard suites — ALL ok:
  templates application 5.2s · delivery/http 7.8s · domain 1.4s · infrastructure 85.9s · jobs 17.4s ·
  tenantdata 18.0s · scenarios 286.8s · iam 141.3s. PASS.
- No migration/platform touch → no full integration sweep (profile §2 L1 selective policy).

## Dual gate

**Fixed range:** `4a91aadf..ac9b8391` (commits 291bce1c code, ac9b8391 wiki). Both arms cold,
clean-context, git read-only, same prompt-pack (REVIEW-STANDARD order + G1-G3 + anti-slop
checklist + severity schema).

- **Opus arm (row 7): APPROVE** (LGTM-with-comments). G1 PASS (in-bounds, no new surface,
  no-fallback honored, single caller routes_query.go:218, blast radius contained); G2 PASS
  (Option A rejection, unguarded-Exists trust model, integration-omission all documented);
  G3 NO (409 spec gap = disclosed pending contract-lock). Findings: 1 suggestion
  (CrossTenant test lands on GetVersion tenant guard, not the new Exists gate — legitimate
  boundary test, optional rename), 1 nit (fakes_test.go:299-305 comment over-claims what
  zero-value fakePresigner sites do). Anti-slop: clean, all receipts quoted (fail-closed
  queries.go:72-77 → MapErr default 500; absent-path :78-80 → errors.go:67 → 409 pinned
  errors_test.go:28; production satisfaction main.go:567→buildTemplatesModule).
- **Sonnet arm (row 8): APPROVE.** G1/G2/G3 all pass; ZERO findings survive verification.
  Receipts: queries.go:61-80 gate logic exact vs charter; verified_store.go:235-244 unchanged
  pre-existing; diff --stat 7 files all in templates module + 2 wiki, no openapi/migration/FE/
  parallel-owned test file; 5 tests each hit distinct branch incl. PresignGet-not-called
  assertion on error path.

**§8 merge + reconciliation:** APPROVE ∧ APPROVE → **APPROVE**. No blocking/important on
either arm → nothing to reconcile. Opus's suggestion+nit are advisory (test naming, comment
wording); recorded here, not applied — neither changes behavior, coverage, or a documented
claim, and both arms independently confirmed no-test-theater and comment-warranted.

## Contract-lock execution + §9 delta re-review

Hub GRANTED additive contract-lock (409 on docx-url) with conditions (a)-(d). Condition (d)
triggered: `embedded-spec: true` in every module cfg.yaml means ANY spec edit re-encodes the
base64 swaggerSpec in all modules' api.gen.go — reported to hub with per-file classification
(templates = 16 non-swaggerSpec lines, exactly the generated 409 type+visitor; 9 modules =
zero non-swaggerSpec lines; approval regen no-op). Hub ratified **option A** (spec + all 11
regens, embedded-spec parity; churn expected). Commit `44a1960c`: openapi.yaml +2 lines
(`'409': $ref Conflict` on GET /templates/{id}/versions/{n}/docx-url, idiom-matched to
siblings) + canonical `go generate` regens only, no hand-edits. Post-commit: build clean,
api-lint -strict 0 violations, templates tests green.

**§9 delta gate (corrected matrix per hub correction #2) — both arms APPROVE, zero findings:**
- Opus arm (row 9): 4/4 checks pass — additive-only spec edit verified against numstat,
  machine-plausible generated hunks, no forbidden surface (11 files exact), contract semantics
  = handler truth (errors.go:67 + errors_test.go:28 → Problem/RFC 9457).
- GPT-5.6 Sol arm (row 10, via codex read-only): independent proof — replaced each embedded
  swaggerSpec block with a marker and SHA-256'd parent vs current for the 9 non-templates
  modules: hashes identical → zero non-embedded-spec change; `git diff --check` clean;
  Conflict ref resolves to problem+json Problem schema.
- Merge: APPROVE ∧ APPROVE → **APPROVE**; nothing to reconcile. Contract-lock RELEASED at
  CLOSED per condition (b).

## Runtime verify (charter §8)

**Window GRANTED by hub (2026-07-16); journey EXECUTED — live baseline captured; fix
verification remains post-merge (stack runs pre-fix code).**

Hub granted the verify-window with stack truth "rebuilt from main @ 1070e94c" — that commit
does NOT contain this branch's fix (291bce1c, unmerged/unpushed), so the live stack exercises
the OLD read path by construction. Journey run anyway (conditions honored: `BDOCX-` prefix,
no container restart/reseed/rebuild, network evidence captured):

1. `POST /api/v1/auth/login` (admin dev-seed, Origin header per origin_protection) → session OK.
2. `POST /api/v1/templates` (Idempotency-Key, `{"key":"BDOCX-verify-409","name":"BDOCX runtime
   verify 409"}`) → 201; template `6f326aea-385a-4754-968b-7b22a39e9f95`, version 1 draft,
   `docx_storage_key` = `tenants/ffffffff-.../templates/6f326aea-.../versions/1.docx`,
   `content_hash: null` (no object yet) — confirms lazy-provision contract live.
3. `GET .../versions/1/docx-url` → **200** `{"data":{"url":"http://127.0.0.1:9000/metaldocs-attachments/tenants/.../versions/1.docx?X-Amz-..."}}`
   — the pre-fix URL-to-nowhere, reproduced live.
4. `GET <presigned URL>` → **404** `<Code>NoSuchKey</Code>` from MinIO (full XML captured).

**Verdict:** live DEFECT-REPRO baseline PASS — proves the bug this unit fixes is real on the
current main-built stack and that step 3 flips to `409 problem+json CodeUploadMissing` only
once this branch merges. The 409 network capture requested by the hub is unsatisfiable
pre-merge (reported to hub); fix-side live verification defers to post-merge QA:
same journey → step 3 = 409 CodeUploadMissing → autosave Confirm → docx-url → 200.
Debris: one template `BDOCX-verify-409` (id 6f326aea-385a-4754-968b-7b22a39e9f95), left for
QA-1 triage per prefix convention.

## MinIO orphan cleanup

**Result: NOT PRESENT — nothing deleted.** Presence-check 2026-07-16 against container
`metaldocs-minio` (volume `compose_metaldocs_minio_data`, mounted `/data`): full recursive
listing of bucket `metaldocs-attachments` contains exactly ONE object —
`tenants/ffffffff-ffff-ffff-ffff-ffffffffffff/documents/5b8a8db4-.../revisions/4dcd3ee6....docx`
(system-tenant document revision, unrelated, untouched). No `templates/` prefix exists; no key
matching `a5e1be9f*` or `ef374718*` anywhere. Volume dir dated Jul 10 — rebuilt after the
charter's 2026-06-29 snapshot; orphans went with the old volume.

**FINDING (field, for hub ratification):** `minio/minio:latest` container has NO `find`
binary — `docker exec ... sh -c "find ... 2>/dev/null"` exits 127 silently and reads as
"no results" (false-empty). Presence checks against MinIO data dirs must use `ls -R` (exists)
or `mc`; never trust a redirected `find`.

## Defers

1. **Runtime verify vs :80 stack** — see section above. Bounded to post-merge hub QA (stack
   runs pre-fix image; behavior fully pinned at L1). REQUEST verify-window remains open with
   hub — if granted pre-merge WITH a rebuild from this branch, chip (or hub) drives the
   409→autosave→200 sequence live.
2. ~~OpenAPI additive `'409'` on docx-url route~~ — **DISCHARGED in-unit**: contract-lock
   granted (option A), commit 44a1960c, §9 delta gate APPROVE×2 (see section above). No
   longer a defer.

## Close-out summary (final)

- **Branch:** `claude/determined-driscoll-2e3140` · **Final range:** `4a91aadf..HEAD` —
  291bce1c (code+tests) · ac9b8391 (wiki) · c3d5be9f/665b2bae (evidence) · 44a1960c
  (contract-lock 409 + full regen) · final evidence commit on top.
- **Planner provenance (hub correction #2 disclosure):** plan produced by sonnet subagent
  under then-binding correction #1; stands per hub rule (consumed + reviewed before #2).
- **Ladder:** L0 all green pre- and post-regen (build/vet/gofmt/api-lint -strict 0/module-
  boundaries; pre-existing file-disjoint findings listed above, not this diff); L1 unit full
  sweep zero FAIL; L1 selective integration ALL ok.
- **Dual gate (base range, correction-#1 matrix):** Opus APPROVE + sonnet APPROVE → APPROVE.
- **§9 delta gate (44a1960c, correction-#2 matrix):** Opus APPROVE + GPT-5.6 Sol via codex
  APPROVE → APPROVE. Contract-lock released.
- **Runtime verify:** window granted + journey executed — live defect-repro baseline PASS
  (200+URL → MinIO 404 NoSuchKey on pre-fix stack); 409-capture re-scoped by hub as
  post-merge QA journey (hub-owned). Debris: BDOCX-verify-409 template, disclosed.
- **MinIO orphan cleanup:** NOT PRESENT, nothing deleted (volume rebuilt Jul 10); `find`
  false-empty FINDING ratified by hub.
- **QA verdict:** PASS — L0/L1 + dual gate + delta gate + live baseline.
- **Remaining defer:** exactly one — post-merge 409 live journey (hub-owned).
- **Constraints honored:** no migration; openapi touched ONLY under granted contract-lock
  (additive, canonical regen); no FE; no new capability/module/async surface; tests confined
  to internal/modules/templates; parallel-owned tests/docx_v2/templates_integration_test.go
  untouched; nothing pushed.
