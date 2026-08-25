# T11 — B11 Access Administration — P8 R6 Operator Partial Re-LOCK

> **Status:** OPERATOR-APPROVED / PARTIAL RE-LOCK.  
> **Artifact:** `docs/work/current/t11-b11-access-administration-p8-r6.html`  
> **Git blob:** `26e8905c5c5012aba59280b1001f62529ed4dfd0`  
> **Scope:** only the pagination / identity-picker surfaces reopened by the PR #173 review finding.

## Trigger

PR #173 review correctly identified that R5/P9 had not proven visible continuation for:

```text
op27 GroupMemberPage in the selected Group member list
op6 UserPage in grant User selection
op22 GroupPage in grant Group selection
op16 AreaPage in grant Area selection
```

The add-member User picker already had a visible pager, but R6 keeps it inside the same bounded continuation proof so op6 behavior is consistent across both User-selection jobs.

## Preserved LOCK scope

The material review finding did **not** reopen:

```text
/admin/access route
Por Área / Grupos / Funções IA
R4/R5 low-fidelity frame
Group multi-Area / Company access footprint semantics
Area-specific vs Company-wide separation
fixed Role meaning
membership add/remove consequence semantics
contextual Area/Group grant entry
Subject × Role × Scope final review
exact revoke
ambiguous same-key retry
B11-F1 op31 precision
Authorization authority
P10 pattern consolidation
89-operation census
```

## R6 delta operated

The operator was asked to judge only:

```text
1. Group member continuation to a later page and later-page removal target.
2. Continuation failure preserving the loaded member page.
3. Grant User selection from a later page.
4. Grant Group selection from a later page.
5. Grant Area selection from a later page.
6. Supporting-read continuation failure preserving loaded page/draft.
7. Contextual Group/Area grants preserving their exact preselection.
```

The exact R6 candidate had already been verified on the same bytes:

```text
structural verifier   12 / 12 PASS
Chromium behavior     23 / 23 PASS
JavaScript parse      PASS
Git blob              26e8905c5c5012aba59280b1001f62529ed4dfd0
```

## Operator disposition

**APPROVED.**

This approval is interpreted according to the explicit gate presented to the operator: **partial re-LOCK of the bounded R6 pagination delta only**.

Therefore:

```text
R6-01 Group member pagination          LOCKED
R6-02 Add-member User pagination       LOCKED
R6-03 Grant User pagination            LOCKED
R6-04 Grant Group pagination           LOCKED
R6-05 Grant Area pagination            LOCKED
```

The prior R5 LOCK remains the accepted basis for all unaffected B11 interaction structure; R6 becomes the canonical full B11 P8 reconstruction artifact because it contains the preserved R5 structure plus the re-locked pagination corrections.

## Reopen triggers

Reopen only on material Evidence such as:

```text
real scale proves visible cursor traversal itself insufficient for a proven job;
a new accepted search/filter requirement is proven;
implementation cannot realize the opaque cursor/history law without violating current contracts;
P11 integration exposes a cross-block contradiction in these collection-selection paths.
```

Preference for a richer search console or preloading all identities is not a reopen trigger.
