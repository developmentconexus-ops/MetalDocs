# Architecture: Data Model — LEGACY CURRENT-STATE REFERENCE

> **Status:** CURRENT DATABASE EVIDENCE ONLY — target model not designed yet
> **Marked:** 2026-08-14

The current Postgres schema remains authoritative for **what the running application stores today**. It is not the target domain/data model for the Cohesive Platform Redesign.

Current target authority:

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md)
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

The redesign is explicitly reconsidering:

- `documents` / `controlled_documents` / template table families;
- Document vs DocumentRevision identity/lifecycle;
- Organization/Area/Group/RoleAssignment storage;
- ApprovalPolicy/Step/Instance/evidence storage;
- DocumentType / Family / GovernanceClass;
- numbering, renditions, release/effectivity, periodic review, distribution/read-ack and supporting projections.

Do not design migrations or preserve current tables merely because they exist. The final schema/table-ownership/transaction model will be specified only after the product/domain design closes.

For current table truth, use `wiki/database/index.md`, `db/baseline/` and live migrations. The former detailed content of this page remains available in Git history.
