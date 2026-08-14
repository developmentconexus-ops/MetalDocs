# Product Vision

> **Last verified:** 2026-08-14
> **Status:** Active product intent; domain details still being finalized by the Cohesive Platform Redesign.

## One-liner

MetalDocs is a governed operational-information platform for organizations that need controlled, versioned, reviewable and auditable procedures, policies, instructions, forms and records with trustworthy publication/distribution evidence.

## Problem

Organizations operating under quality, safety, information-security or other controlled-process regimes need to know and prove:

- what the current official information is;
- who authored/reviewed/approved it;
- exactly which content those decisions referred to;
- what changed between revisions and why;
- which revision is effective;
- which template/revision a derived document came from;
- who is allowed to create/edit/approve/administer each class of information;
- whether required readers received/read/acknowledged released information where policy requires it;
- that historical evidence cannot be silently rewritten.

Generic editors/file shares can store documents but do not by themselves provide a coherent governed lifecycle, business numbering, approval evidence, effectivity, immutable provenance and organization-scoped authority.

## Product direction

The active redesign is converging on:

- `Document` as stable governed identity;
- `DocumentRevision` as versioned governed content;
- templates as an **exact released revision designated to seed other documents**, not a parallel lifecycle;
- `DocumentType` as the document-classification/configuration anchor (exact policy shape still being finalized);
- Organization with Tenant/Area/User/Group;
- scoped role/permission authorization;
- a small versioned human Approval engine specialized for governed information;
- immutable submission/decision evidence tied to exact revision/content hashes;
- a Release/effectivity boundary that makes the official revision unambiguous;
- supporting audit, rendering/renditions, periodic review, distribution/read-ack, notifications, search and token/value services consuming the same canonical truths.

## Critical content-truth rule

A human approval, freeze and official rendition must always refer to the **same exact revision/content digest**. MetalDocs must never allow an approver to review one piece of content while the system signs/releases another derived source.

## What MetalDocs is not

- Not a general-purpose word processor; the editor is an authoring surface inside a governed lifecycle.
- Not merely a file share/DMS folder tree.
- Not a generic BPM/workflow engine for arbitrary business processes.
- Not an IAM platform; identity/authorization exist to serve MetalDocs product operations.
- Not an architecture showcase: infrastructure such as Keycloak, OpenFGA, BPMN engines or policy languages is introduced only when a real requirement justifies it.

## Active design authority

For domain/architecture details read:

- [../architecture/cohesive-platform-redesign.md](../architecture/cohesive-platform-redesign.md)
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
