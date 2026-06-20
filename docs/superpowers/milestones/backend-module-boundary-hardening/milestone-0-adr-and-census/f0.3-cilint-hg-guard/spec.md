# Feature F0.3 — Spec — cilint `hgcrossmodule` H-G guard

> **Milestone:** 0 · **Folder:** `f0.3-cilint-hg-guard` · **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-20 / leandrotca — *contract locked by `../mission.md` §5 row 1 + ADR-0039
> D6 + the F0.2 census (`../f0.2-binding-census/census.md`) it consumes; no implementation begins until this
> line is filled.*

## Interview record (fail-closed gate)

Scope and contract are locked by ADR-0039 D6 (CI enforcement) and the F0.2 census (owner map + in-scope/exempt
sets). The one design-open question — *how does the guard stay green during M1–M4 while the in-scope sites are
still un-ported?* — is answered by the baseline-ledger pattern below. Recorded; no fresh interview needed.

| # | Question | Answer |
|---|----------|--------|
| 1 | What is the violation the guard flags? | **Locked (ADR-0039 D1):** a raw `FROM`/`JOIN` (incl. subquery/`EXISTS`) against **another top-level module's owned base table**, in a non-owner package, outside the allowlist. |
| 2 | Where does "who owns which table" come from? | **Locked (F0.2 census table→owner map):** encoded as `hgOwnerByTable` data in the analyzer. Owner = the **top-level** module dir under `internal/modules/<module>/` (so `documents/approval` ⊂ `documents`; `iam/presence` ⊂ `iam`). |
| 3 | How does the guard avoid comment/identifier false positives? | **Locked:** scan **only `*ast.BasicLit` STRING nodes** (SQL lives in string literals). The census found real comments naming foreign tables (`people_service.go:690`, `observability_repository.go:164`) — a raw-source grep would false-positive; the AST scope structurally avoids them. |
| 4 | How is the guard green now when B/C/C4/N1 are still un-ported? | **Locked (baseline-ledger):** two allowlists — `hgPendingRemediation` (the in-scope debt: B1–B8, C1–C4, N1) which **M1–M4 drain to empty** as each site is ported, and `hgExempt` (X1–X8, permanent, ADR-0039 D3(d)–(f)). Wired green today; any **new/regressed** cross-module read (on neither list) fails the build immediately. Terminal (§8): `hgPendingRemediation` is empty. |
| 5 | Inline escape hatch? | **Locked:** `//cilint:allow-hgcrossmodule` on the offending line (sibling of `//cilint:allow-responsemap`), for a deliberate, rationale-commented exception. |
| 6 | Residual / not-covered? | **Locked (census coverage statement):** dynamically-assembled or aliased table names behind Go vars are invisible to a literal-token scan (same residual as the H-D guard). Recorded, not engineered away. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  1. **CI / `go run ./tools/cilint ./...`** — the build gate. Exits non-zero iff a cross-module base-table
     read exists outside `hgPendingRemediation ∪ hgExempt ∪ inline-allow`. This is how "the class can't
     silently re-open" (mission goal) is mechanized.
  2. **M1–M4** — each milestone, on porting its sites, **removes those entries from `hgPendingRemediation`**;
     the shrinking ledger is the mission's machine-checked progress meter.
  3. **The terminal mission-validator (§8)** — asserts `hgPendingRemediation` is empty and the full-tree run
     is green under both readings (0 violations outside the recorded permanent exempt allowlist).
- **Contract (exact shape consumers rely on):**
  - A new analyzer `HGCrossModule(files []string) []Finding` in `tools/cilint/internal/analyzers/`, wired into
    `RunAll`, `Analyzer: "hgcrossmodule"`, message naming the reader module, owner module, table, and ADR-0039.
  - `hgOwnerByTable` — the census table→owner map (every owned base table → top-level module).
  - `hgPendingRemediation` — `{fileSuffix, table}` entries = the exact F0.2 in-scope sites (B/C/C4/N1).
  - `hgExempt` — `{fileSuffix, table}` entries = X1–X8 with per-site ADR-0039 D3(d)/(e)/(f) rationale comments.
  - `//cilint:allow-hgcrossmodule` inline directive.
- **Source of truth:** `wiki/decisions/0039-*.md` (rule + exemptions), `../f0.2-binding-census/census.md`
  (owner map + in-scope/exempt sites), the live tree under `internal/modules/`.

## What this feature implements

The `hgcrossmodule` cilint analyzer (sibling of `noresponsemap`) + its bite/green unit tests, wired into
`RunAll`, green on the full tree today via the documented baseline ledger, failing the build on any new
cross-module owned-base-table read.

## Non-goals (mandatory)

- **No porting.** The guard *detects*; M1–M4 *fix*. Zero production SQL edited by this feature.
- **No new classification rule.** The guard mechanizes ADR-0039 D1/D3 exactly; it invents no verdict.
- **No emptying of `hgPendingRemediation` here.** That ledger starts full (all in-scope debt) and is drained
  by M1–M4. Emptying it in M0 would be a false green.
- **No flow/data-type analysis** — literal-token scan inside string literals only; dynamic/aliased SQL is the
  recorded residual.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Flags a cross-module base-table read (e.g. CD reading `documents`) | `TestHGCrossModule_Positive_CrossModuleRead` | fixture (bite) |
| Does NOT flag an own-table read (CD reading `controlled_documents`) | `TestHGCrossModule_Negative_OwnTable` | fixture |
| Does NOT flag a sub-package as cross-module (`documents/approval` reading `documents`) | `TestHGCrossModule_Negative_SubpackageSameModule` | fixture |
| Does NOT flag a foreign-table name appearing only in a **comment** | `TestHGCrossModule_Negative_CommentMention` | fixture |
| `hgPendingRemediation` entry suppresses a known in-scope site | `TestHGCrossModule_Negative_PendingBaseline` | fixture |
| `hgExempt` entry suppresses an X-site (platform sink / auth / jobs) | `TestHGCrossModule_Negative_Exempt` | fixture |
| Inline `//cilint:allow-hgcrossmodule` suppresses | `TestHGCrossModule_Negative_AllowDirective` | fixture |
| **Full-tree run is green today** (all real violations ∈ pending ∪ exempt) | `go run ./tools/cilint ./...` exit 0; recorded in `evidence.md` | **real** (live tree) |
| Analyzer + suite compile and pass | `go test ./tools/cilint/...` | real |

> The bite tests prove the guard *catches*; the green tests prove it *doesn't over-flag*; the real full-tree
> run proves the baseline ledger is accurate (every live cross-module read is accounted for, none orphaned).

## ADR needed?

- [x] **No new ADR** — the guard *mechanizes* ADR-0039 D6. (If the wired run surfaces a live cross-module read
  that is **neither** a known in-scope site **nor** a justifiable exemption, that is a census gap → HS-6 back
  to F0.2, not a new ADR here.)
