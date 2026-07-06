# F-R2 Evidence — ADR status-field CI gate (Dim 10 → CONFIRMED)

Closes the Dim-10 DEBT: the ADR status-field rule (F9.1) is now enforced by a blocking CI gate, not
just a documented convention.

## Changes

| File | Change |
|------|--------|
| `scripts/check-adr-status.sh` (new) | Single-source sweep — awk budget check (≤3 lines / ≤400 chars) lifted verbatim from the governance doc's one-liner; `[dir]` arg (default `wiki/decisions`); exit 0 `adr-status: clean` / exit 1 offender list / exit 2 usage. |
| `.github/workflows/governance-check.yml` | Blocking `ADR status-field budget gate` step in the `check` job invoking the script (no `continue-on-error`). |
| `wiki/standards/documentation-governance.md` | "optional future extension" text replaced with the CI-enforced statement; manual sweep now points at the script. |

## Negative proof (gate FAILS on an over-budget status block)

Synthetic ADR with a 4-line / 846-char status block, in a temp dir (never committed):

```
=== NEGATIVE: synthetic over-budget ADR in temp dir (expect flagged, exit 1) ===
::error::ADR status-field budget exceeded (>3 physical lines OR >400 chars):
9999-synthetic.md: 4 lines, 846 chars
Fix: move execution history / amendment narrative OUT of the status block into
the ADR body (## Status history) or a companion NNNN-execution-history.md.
See wiki/standards/documentation-governance.md (ADR status field).
exit=1
```

## Positive proof (gate PASSES on the clean tree)

```
=== POSITIVE: sweep the real wiki/decisions (expect clean, exit 0) ===
adr-status: clean
exit=0
```

The real `wiki/decisions/` tree is clean (post-F9.1), so the gate produces no false positive.

## Blocking-ness proof

```
$ grep -n -A1 'ADR status-field budget gate' .github/workflows/governance-check.yml
      - name: ADR status-field budget gate
        run: bash scripts/check-adr-status.sh
```

The step has no `continue-on-error`, so a non-zero exit fails the `check` job and blocks the PR. It
runs under `on: pull_request: branches: [main]`.

## Doc-truth proof

The "Wiring this sweep into CI ... is an optional future extension, not required" sentence is removed
from `documentation-governance.md`; replaced by a **CI-enforced (F-R2)** paragraph naming the blocking
step and the single-source script.

## Defers / follow-ups

None. Gate is live logic + CI wiring; single source prevents local/CI drift.
