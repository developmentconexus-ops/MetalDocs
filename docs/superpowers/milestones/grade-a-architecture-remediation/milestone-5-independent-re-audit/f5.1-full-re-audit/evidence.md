# F5.1 — Evidence

> **Feature:** F5.1 full independent re-audit  ·  **Status:** Done (deliverable produced). Verdict triggers F5.2 (HS-5).
> **Run:** 2026-06-15  ·  **Audited HEAD:** `02ed1c24`  ·  **Report HEAD:** written at `8a9759ac` working tree.

## What ran

| Step | Command / mechanism | Result |
|------|---------------------|--------|
| Re-audit fan-out | `Workflow` run `wf_b0109977-23a` — 42 agents, ~2.10M tokens, 1021 tool-uses, ~28 min | completed |
| Method | 10 dimension auditors (sonnet) read actual code at HEAD → adversarial skeptic (sonnet, refute-by-default) per Critical/Major → 2 grep-backed class-count agents (H-D, H-G) → synthesis | as designed |
| Deliverable | `wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md` (23,221 chars) | written |

## Acceptance (per F5.1 spec Validation Gate)

- ✅ Written report with 10-row scorecard vs 2026-06-13 baseline.
- ✅ Explicit ≥A− yes/no for each of the 3 formerly-C dimensions (all **No**).
- ✅ Every Critical/Major listed with skeptic verdict + `file:line`; refuted/downgraded recorded separately (so not re-raised).
- ✅ H-D and H-G re-measured with exact reproducible grep/build commands (report §6).
- ✅ §6 pass-bar verdict rendered mechanically.

F5.1's job was to produce an independent, cited, adversarially-verified audit. It did. The feature is complete regardless of the verdict's polarity.

## Verdict: MICRO-WAVE NEEDED (all 4 §6 checks FAIL)

| § Check | Bar | Result |
|---------|-----|--------|
| (1) 3 formerly-C dims ≥ A− | all ≥ A− | **FAIL** — module-boundaries **B+**, contract-api **C+**, composition **B+** |
| (2) 0 new confirmed Critical/Major | 0 | **FAIL** — **21** confirmed (1 Critical + 20 Major) per §4 detail (report §3 headline says 18 — undercount; §4 table is authoritative) |
| (3) H-D = 0 | 0 | **FAIL** — **4** drift sites (templates routes ×3 + taxonomy `routes_profiles` ×1) |
| (4) H-G = 0 | 0 | **FAIL** — **1** (`documents/application/service.go:282` hardcoded `"published"`); +1 NEW boundary site `security ListOffHoursAdminActions` (Major #11, JOINs `iam_user_roles`, not in M4 census) |

## Highest-priority confirmed findings (correctness/security, not hygiene)

- **#3 Major (authz):** `iam/authz/authz.go:123` — `authz.Require` checks only `effective_to IS NULL`, ignores `effective_from <= now()` → **future-dated memberships grant access prematurely**. (Sibling `ResolveEligibleActors` has the correct dual predicate.)
- **#1 Major (authz):** `controlleddocuments/application/service.go:173` — manual-code CD-create branch never seeds tx identity → `MustActorID` returns `ErrActorContextMissing` before the system-admin bypass fires → **all non-system-admin manual-code creates fail**.
- **#2 Major (authz):** `documents/approval/application/read_service.go:68` — `CapDocumentView` tenant-grade silently narrowed to area-grade for documents with a real area code.
- **#13 Critical (contract):** `documents/.../handler.go:881` — checkpoint endpoints serialize an untagged `domain.Checkpoint` → **PascalCase on the wire, breaks the FE-generated snake_case contract**.

## Disposition

Per **HS-5**, the verdict is presented to the operator for the continue-vs-replan decision before any F5.2 micro-wave begins. F5.2 scope (if continued) is the four triggers in report §3 "Overall Verdict": lift the 3 dims to A−, resolve the confirmed Critical/Major by root-cause family, drive H-D and H-G to 0.
