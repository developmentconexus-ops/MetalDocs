# F2.2 — tier-1 ↔ tier-2 capability-name coherence

> **Milestone:** M2 · **Governing contract:** `../validation-contract.md §2` (binding; HS-7 on drift).
> **Approved for implementation:** 2026-07-03.

## Consumer contract

The **consumer** is CI + future maintainers. Runtime truth (contract §2): the two divergences the
2026-07-03 review named are **already code-resolved in ADR 0022 Phase 11 F4**; the review read stale
follow-up lines. F2.2 therefore delivers **verify + pin + doc-truth restore**, not "close two live
divergences":

1. **Verify (source evidence)** tier-1 == tier-2 capability names for both sites:
   - force-release: tier-1 `permissions.go` row → `CapMembershipManage`; tier-2
     `documents/repository/repository.go:798,828` assert `CapMembershipManage`.
   - approval-route management: tier-1 explicit `/approval/routes*` rows → `CapRouteManage`; tier-2
     route-management service asserts `route.manage`.
2. **Regression pin** — a test binding, for **exactly these two reconciled routes**, tier-1 route→cap
   == tier-2 asserted cap, so a future re-divergence reddens CI. NOT a blanket tier-1==tier-2 assertion
   (coarse/fine legitimately differ elsewhere). Prefer extending `permissions_test.go` /
   `permissions_authz_scope_test.go`.
3. **Doc-truth restore** — back-annotate the stale ADR 0022 Phase 7/8 ⚠️-follow-up lines (≈198,
   236–237, 250) as **RESOLVED in Phase 11 F4** (xref lines 349–351); correct any wiki page still
   describing either site as an open divergence. No code-behavior change.

## Non-goals

- **Do not "align" the deliberate coarse/fine `/approval/` differences** (tier-1 `document.submit` gate
  vs tier-2 `document.signoff` on decision, etc.) — that is the intended PDP split (review Dimension 2),
  recorded as an intentional written exception, not a defect.
- No change to route behavior, capabilities, or the PDP structure. No new capability.

## Validation gate

Per `../validation-contract.md §2.3`: POSITIVE (source excerpts show alignment; pin GREEN on HEAD; 5
authz lints green; ADR 0022 has no un-annotated open-divergence claim) + NEGATIVE (synthetic
re-divergence → pin RED, captured). Copied as the acceptance checklist into `evidence.md`.
