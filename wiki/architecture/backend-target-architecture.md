# Backend Target Architecture — HISTORICAL PRIOR TARGET

> **Status:** HISTORICAL / no longer normative for product/domain topology
> **Marked:** 2026-08-14

This page previously served as the normative target for the backend professionalization program. The Cohesive Platform Redesign now re-adjudicates product/domain boundaries, module ownership, authorization, approval, Controlled Information and the target data model.

Current target authority:

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md)
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

The former detailed REQ-* target remains available in Git history for infrastructure/engineering archaeology. Stable concerns such as contract-first API behavior, RFC 9457 errors, multi-tenant isolation, RLS defense-in-depth, transactional outbox, observability and DB-enforced invariants continue through their owning architecture/standard pages unless the active redesign explicitly changes them.

Do **not** infer from the former “15 bounded-context modules” topology, historical IAM/approval wording or previous module ownership that those boundaries survive the redesign.

A new technical target architecture will be written only after the whole-product domain model, final authorization matrix, supporting concerns and migration map are closed.
