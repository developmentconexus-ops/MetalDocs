# ADR 0022 — Execution history

> Relocated from the ADR 0022 `> **Status:**` and `> **Last verified:**` fields on 2026-07-06 (F9.1
> adr-hygiene, M9 governance-hygiene). The status field had grown to a single 2757-character line
> covering all 13 execution phases plus Wave Z — a "mega-status" that made the ADR's decision state
> unreadable at a glance (architecture review `778f494a` finding 105). This document is the relocated
> history, content preserved with **zero information loss**, restructured into dated phase entries.
> The ADR's own `> **Status:**` field now carries only the current state + a pointer here; the full
> `## Implementation — phased plan` narrative in the ADR body (Context/Decision/Consequences-adjacent
> content) is untouched by this relocation — only the header-field changelog moved.

See [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) for the ADR itself.

## Direction + governance model approved — 2026-06-03

Approved 2026-06-03.

## Phase 1 — complete 2026-06-03

Registry hygiene, behavior-neutral.

## Phase 2 — complete 2026-06-03

Typed scope classification + lifecycle grant seed; spec annotation scoped + lint-activation deferred —
see Phase 2 close-out in the ADR body.

## Phase 3 — complete 2026-06-03

Membership area-scoping — the original `area_admin` 403 closed at root: real `areaCode` passed to
tier-2, role-string gate deleted, `ErrCapDenied` → 403, R1 system_admin bypass asserted.

## Phase 4 — complete 2026-06-03

Directory/list area-scoping — `isMembershipDirectoryAdmin` `RoleSystemAdmin` gate replaced by
data-layer capability/area scope: system_admin → tenant-wide, area_admin → managed areas (filtered
IN SQL, R3), others → self-only; last handler role-name literal removed (ADR 0021).

## Phase 5 — complete 2026-06-03

CI binding — new `scripts/api-lint` rules `no-inline-capability` / `seed-registry-parity` /
`wiki-capability-parity`, area-grade annotation parity test, OpenAPI access-policy enum reconciled as
a separate vocabulary, RLS hardening verified: transaction-local GUCs PASS + native-RLS N/A
trigger-tripwire model.

## Phase 7 — complete 2026-06-03

Typed scope bound to runtime — the 5 declared-but-unenforced area-grade caps
`document.create`/`document.edit`/`controlled_documents.create`/`obsolete`/`supersede` now pass the
resource's real area to tier-2; new `authz-area-scope-binding` AST guard bans
`Require(<areaGradeCap>,"tenant")` (proven to bite on all gaps, then green); `areaEnforcedOps`
self-maintained from `IsAreaGrade`; review MED/LOW fixes folded — `BypassSystem` fail-closed
background bridge (CWE-269), `MembershipDirectoryScope` clock param, shared `SystemAdminExistsSQL`
(DRY), golden seed-row-count test.

## Phase 8–13 — complete 2026-06-04

Raw-string dialect closed, documents handler legacy role dialect decommissioned, registry
minimization, coherence cleanup F4–F8, naming consolidation `doc.*` → `document.*`, CI net revived —
blocking gate `0 blocking, 397 reported`.

## Wave Z Z-6 — complete

`area_membership` governance now in-tx via `LogTx` — T-007 closed, grant/revoke governance writes
atomic with the mutation.

## Phase 6 — complete 2026-06-13

Wiki sync — `authz-tiers.md` updated: scope classification, area-grade enforcement table,
`BypassSystem` background-only restriction; `wiki/modules/iam.md` + `wiki/modules/auth.md` already
current via prior wave syncs.

**Supersedes nothing; extends** ADR [`0007`](0007-two-tier-authz.md) (two-tier model + lint harness)
and ADR [`0021`](0021-tenant-vs-platform-admin-separation.md) (capabilities, not role names, are the
enforced boundary).

## Last-verified changelog (relocated from the `> **Last verified:**` field)

### 2026-06-13 — Phase 6 wiki sync complete

`authz-tiers.md` scope classification, area-grade enforcement table, `BypassSystem` background-only
restriction updated; `iam.md` + `auth.md` already current via Wave 2.12 syncs.

### 2026-06-08 — Phase F FD-1

`code_rules.go` anchor updated — `authz-call-present` deleted, `RunCodeRules` now at `:34`.

### 2026-06-07 — Phase C

Dead-path prune.

### 2026-06-04 — Phase 13 complete: CI coherence-net revival

Principle 5 now ACTUALLY enforced. Two compounding bugs fixed — CI passed `./internal/modules` not the
repo root, double-nesting every binding helper's join so the file-rooted guards
(`no-rolestring-in-delivery`, `authz-area-scope-binding`, `seed-registry-parity`,
`wiki-capability-parity`) silently loaded nothing via their `os.IsNotExist→empty` branch; AND the job
was `continue-on-error: true`. Fix: root arg = repo root (CI passes `.`, `main.go` normalizes via
`filepath.Abs`+`Clean`); new `-strict` flag turns a missing core file into a HARD error (kills the
silent-empty swallow) while preserving the testdata-fixture skip; `-enforce=blocking|reported|all`
splits the gate — binding/dialect/tripwire guards BLOCK (0 on clean tree, exit 0), the ~397 spec-drift
hits stay reported-only/non-blocking; new `e2e_test.go` drives the REAL entrypoint with a planted
violation per AST guard + strict-wrong-root regression.

**Gates:** build/vet/gofmt clean, `go test ./scripts/api-lint/... -count=1` green, blocking gate =
`0 blocking, 397 reported` exit 0, binary-level reddening proven, go-reviewer no CRITICAL/HIGH. Phase 6
wiki sync still runs LAST.

▸ Earlier: Phase 11 complete — coherence cleanup F4·F5·F6·F7 + F8-backend: F4 force-release tier-2 →
`membership.manage` (operator ruling a) + approval-route tier-1 rows + spec
security/pagination-exempt (8 lint hits cleared); F5 tripwire `SkipDir` + allow-list → live tripwire 0,
honest lint total 455→397 after the premise correction (455 = 405 spec-drift + 50 tripwire, NOT ~447
tripwire); F6 dead file/comment deleted + `authz-call-present` dormancy documented; F7 one shared
`LoadDocumentAreaCode` + reserve `area_code='tenant'` (migration 0228 + baseline); F8 deny-default
matrix + `system_admin`/`BypassSystem` audit (atomic sink, fail-closed). Review: no CRITICAL/HIGH. DB
live-bootstrap gate PASS (0228 applies + idempotent; CHECK rejects `area_code='tenant'`).

Phase 12 (F3) complete — `doc.*`→`document.*` cap-value rename (registry+seed+migration 0229),
audit-event vocabulary proven decoupled (BACKEND-ONLY, no FE/history touch), single `document` Admin
Center group, `controlled_documents.*` symmetry flagged-not-actioned; Phase 6 wiki sync NOW UNBLOCKED
(last code phase done).

Phase 10 complete — F2 registry minimization: 4 redundant phantom caps merged into canonical
equivalents; registry 33→29, area-grade 14→11; migration 0227 + baseline mirror.

Phase 9 complete — legacy "docx v2" role dialect decommissioned; post-migration audit added Phases
9-12.

### Prior: 2026-06-04 — Phase 13 was preceded by

(See the "2026-06-04 — Phase 13 complete" entry above for the full compounded-bug fix; the "▸ Earlier"
sub-entries under it — Phase 11, Phase 12, Phase 10, Phase 9 — are the immediately preceding
Last-verified states, listed in the original field in reverse-chronological "prior:" form and
normalized here into forward-chronological order for readability. No content was dropped in that
reordering.)
