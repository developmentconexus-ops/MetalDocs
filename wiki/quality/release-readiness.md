# Release Readiness

> **Last verified:** 2026-05-27
> **Scope:** Canonical release gate and Go/No-Go procedure for MetalDocs.

## Purpose

Release readiness is the final quality gate before merge or release approval.

It exists to answer:

- did governance checks pass
- did hardening and validation checks pass
- is the current change fit to merge or release

## Canonical checks

- `check-governance`
- `phase3-hardening-gate`

## Canonical execution

Official:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/phase3-release-readiness.ps1 -BaseRef origin/main
```

Local without remote baseline:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/phase3-release-readiness.ps1 -BaseRef HEAD~1
```

## GitHub Actions entrypoint

Workflow:

- `release-readiness`

Inputs:

- `base_ref`
- `skip_govulncheck`

## Acceptance rule

Approve merge or release only when:

- final status is `approved`
- `governance_check = approved`
- `hardening_gate = approved`

## Evidence

Primary artifact:

- `non_git/release/phase3_release_readiness_<timestamp>.json`

Supporting evidence:

- `non_git/hardening/*.json`
- `non_git/contract/*.json`
- `non_git/security/*.json`

## Failure handling

If the gate fails:

1. fix the failing boundary
2. rerun release readiness
3. do not approve merge or release until the final status returns to `approved`

## Source normalization note

This page is the canonical wiki home for release-quality governance.
The older draft runbook under `docs/runbooks/release-readiness.md` should be treated as staging input until references are fully normalized.
