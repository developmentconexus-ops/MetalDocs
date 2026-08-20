---
id: work-ai-dialog
kind: work
status: active
owner: architecture
summary: Temporary Lead and Fable review record for the repository documentation and agent-context governance decision.
---

# AI dialogue

> **TEMPORARY / NON-AUTHORITATIVE / DELETE BEFORE MERGE**

## Review context

```text
Repository: developmentconexus-ops/MetalDocs
Branch: docs/repository-information-architecture
PR: #132
Expected review HEAD: REVALIDATE REMOTELY BEFORE REVIEW
Current PR purpose: decide repository documentation and agent-context governance only
Product implementation: not authorized
Legacy deletion: not started
PR #131: frozen provenance only
```

Review target:

```text
docs/development/documentation.md
```

Supporting non-authoritative work:

```text
docs/work/current/proposal.md
docs/work/current/plan.md
```

Canonical engineering method and independent-review workflow:

```text
developmentconexus-ops/conexus-methodology/METHOD.md
developmentconexus-ops/conexus-methodology/README.md
```

Current MetalDocs repository files are evidence of present failure modes and current tool consumers; their existing documentation shape is not target authority.

## Review request

Perform one independent adversarial review of the proposed repository documentation profile.

Attack the following questions:

1. Does selecting one `docs/` root materially reduce authority ambiguity, context bloat, and Git conflict risk compared with retaining `wiki/` + `docs/`?
2. Are semantic filenames, frontmatter, an intent-based index, and explicit MkDocs navigation the smallest sustainable information architecture?
3. Does the proposal preserve accepted Product/R10 truth while deleting process/history artifacts from the live tree?
4. Is Git/closed-PR history sufficient as the archive, or is any named current consumer missing?
5. Is the proposed `AGENTS.md` model small enough for routine LLM orientation without hiding load-bearing authority?
6. Does the one-proposal/one-AI-dialog lifecycle eliminate review-artifact bloat without weakening independent challenge, Lead adjudication, or operator ratification?
7. Is one coherent ratifiable gate per PR the correct unit, and are S0/G0/G1/T8-E boundaries coherent and merge-safe?
8. Does the execution plan accidentally let a Writer invent a material Product/R10 decision during consolidation?
9. Is the allowlist deletion rule strong enough to remove legacy while preserving current runtime safety rails and runbooks?
10. Is the proposed docs-hygiene verifier structurally enforceable with the current Go verifier/negative-fixture spine?
11. Does any proposed mechanism add unnecessary framework/tooling complexity, especially MkDocs, Goldmark, frontmatter, or PR-draft-aware checks?
12. Would this profile transfer cleanly to Marketplace Central and other Conexus products without centralizing their product truth?
13. What can be removed from the proposal or plan without weakening a distinct material property?
14. Is there a materially better Global Maximum?

Required classification for each material finding:

```text
BLOCKER | MAJOR | LOW
claim
repo/source evidence
root cause
property at risk
smallest correction
upstream/product reopen required: yes/no
```

Required primary verdict — exactly one:

```text
APPROVE REPOSITORY DOCUMENTATION PROFILE
APPROVE REPOSITORY DOCUMENTATION PROFILE WITH MATERIAL FIXES
DO NOT APPROVE REPOSITORY DOCUMENTATION PROFILE
```

Explicitly report:

```text
Global Maximum confirmed: yes/no
one docs root confirmed: yes/no
naming/navigation model confirmed: yes/no
agent-context model confirmed: yes/no
AI-dialog/Fable lifecycle confirmed: yes/no
PR lifecycle confirmed: yes/no
allowlist deletion safe: yes/no
execution plan implementable: yes/no
another review round materially required: yes/no
Lead adjudication may proceed: yes/no
```

## Fable review

Fable writes its review here and modifies no other file.

## Lead adjudication

Lead confronts every material finding here. Reviewer output is evidence, never authority.

## Bounded round 2

Use only if a real material contradiction survives Lead adjudication.

## Operator decision

Record final operator ratification or the exact bounded reopen here.
