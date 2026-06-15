# Feature F4c.4 — Plan

> **Spec:** [`spec.md`](spec.md) (Approved 2026-06-15). This plan is the **how**; spec is the
> **what**. Evidence is recorded in [`evidence.md`](evidence.md) at close.

Linear plan. One commit per numbered step unless noted. No fan-out — F4c.4 is mechanical.

## Step 0 — Pre-flight (no code)

- Read HEAD `.github/workflows/module-boundaries.yml` (target of Q2 edit).
- Confirm operator DSN reachable for AC6/AC7: `.\scripts\start-api.ps1` smoke OK.
- Record HEAD SHA in evidence (the F4c.3-closed baseline).

## Step 1 — `pgtest` audit (AC5 branch decision)

```bash
grep -rn "pgtest\." --include="*.go" | grep -v "internal/testsupport/pgtest/"
grep -rn "testsupport/pgtest" --include="*.go" | grep -v "internal/testsupport/pgtest/"
```

Branch:
- **Zero callers** → Step 2 deletes the dir.
- **Callers exist** → classify; if **any stateful-write** caller → HS-6 STOP, surface to operator.
  If **all no-write** → Step 2 keeps + adds `// no-write only` header.

Record both grep outputs in evidence verbatim (before/after).

## Step 2 — `pgtest` resolution

Per Step 1 branch:
- Delete: `git rm -r internal/testsupport/pgtest/`. Run `go build ./...` + `go vet ./...` to prove
  no orphan import remains. Commit: `chore(milestone-4c): F4c.4 step 2 — retire dead pgtest harness`.
- Keep: edit `internal/testsupport/pgtest/pgtest.go` to add a `// no-write only` package doc; list
  the audited surviving callers in the comment. Same build/vet. Commit:
  `chore(milestone-4c): F4c.4 step 2 — scope pgtest to no-write callers`.

## Step 3 — TDD RED: plant per-rule violations

Create a throwaway `internal/scratch/f4c4_red/f4c4_red_test.go` with first line
`//go:build integration` and four planted lines (one per rule R1–R4). Run intended guard script
(not yet written) — confirms script absence by `command not found`, which is the trivial RED.
That's insufficient — instead:

**Real RED process:** for each rule R1..R4 in turn:
1. Plant violation in the throwaway file.
2. Implement the script's *single* rule for that rule only.
3. Run script → confirm exit non-zero + `<path>:<line>: <rule-id>: <text>` printed.
4. Remove the violation line → rerun → exit zero.
5. Move to the next rule.

This is the TDD shape that proves each rule independently. Captured in evidence per rule.

## Step 4 — Implement `scripts/check-test-discipline.sh`

Bash 3.2-compatible (macOS default), no GNU-grep-only flags. Layout:

```bash
#!/usr/bin/env bash
# F4c.4 — Test-discipline grep guard. See docs/superpowers/milestones/.../f4c.4-ci-grep-guards/spec.md
set -euo pipefail

# Discover integration test files (first line == //go:build integration), exclude tests/integration/testdb/**
mapfile -t FILES < <(
  git ls-files '*_test.go' \
    | grep -v '^tests/integration/testdb/' \
    | while read -r f; do
        head -1 "$f" | grep -q '^//go:build integration' && echo "$f"
      done
)

violations=0
report() { echo "$1:$2: $3: $4"; violations=$((violations+1)); }

# R1 inline metaldocs.asserted_caps set_config
# R2 is_local=false
# R3 hardcoded dev-tenant UUID literal (the testdb.DevTenant constant value)
# R4 bare unqualified `documents` table reference

for f in "${FILES[@]}"; do
  while IFS=: read -r line text; do
    report "$f" "$line" "R1" "$text"
  done < <(grep -nE "set_config\('metaldocs\.asserted_caps'" "$f" || true)
  while IFS=: read -r line text; do
    report "$f" "$line" "R2" "$text"
  done < <(grep -nE "set_config\([^)]*,[[:space:]]*false[[:space:]]*\)" "$f" || true)
  # R3 — DevTenantID literal pinned to the one canonical const value (read from testdb at runtime).
  if [[ -n "${DEV_TENANT_UUID:-}" ]]; then
    while IFS=: read -r line text; do
      report "$f" "$line" "R3" "$text"
    done < <(grep -nF "\"${DEV_TENANT_UUID}\"" "$f" || true)
  fi
  while IFS=: read -r line text; do
    report "$f" "$line" "R4" "$text"
  done < <(grep -nE "(FROM|JOIN|INTO|UPDATE)[[:space:]]+documents([^_a-zA-Z]|\$)" "$f" || true)
done

if (( violations > 0 )); then
  echo "test-discipline: $violations violation(s)"; exit 1
fi
echo "test-discipline: clean"
```

`DEV_TENANT_UUID` is exported by the script itself by reading `tests/integration/testdb/factory.go`
(grep the const), so the guard is self-contained — no env-dependence in CI.

Commit: `feat(milestone-4c): F4c.4 step 4 — test-discipline grep guard (R1–R4)`.

## Step 5 — Extend `module-boundaries.yml`

Append the documented YAML step from spec §C. No reorder. Validate locally with
`yamllint .github/workflows/module-boundaries.yml` if available; otherwise eyeball + `actionlint`
if present. Commit: `ci(milestone-4c): F4c.4 step 5 — wire test-discipline guard into module-boundaries workflow`.

## Step 6 — Guard reference doc

Create `wiki/quality/test-discipline.md` (R1–R4 + allow-list + sanctioned pattern + dev-onboarding
example). Add one line to `wiki/README.md` index. Bump `Last verified:` per CLAUDE.md drift policy.
Dispatch `wiki-curator` to refresh related anchors. Commit: `docs(milestone-4c): F4c.4 step 6 —
test-discipline reference doc + wiki index`.

## Step 7 — Clean-tree GREEN proof

```bash
bash scripts/check-test-discipline.sh; echo $?
```

Expect `test-discipline: clean` + exit 0. Capture verbatim. **AC1 green.**

## Step 8 — Full integration suite + regression

```powershell
$env:DATABASE_URL = "<operator DSN>"
go test -tags integration -count=1 ./... 2>&1 | Tee-Object -FilePath docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4c-test-fixture-framework/f4c.4-ci-grep-guards/clean-baseline.log
```

Compare RED set against F4c.3 close baseline (M4b `tests/integration/scenarios/` debt expected; no
new failures permitted). **AC6 + AC7.**

Plus targeted regression run from AC7's `-run` filter; capture verbatim.

## Step 9 — `evidence.md` + close

Fill the per-AC table with real commands + real output (planted-violation snippets per rule, before/
after greps, workflow-step diff, doc paths, clean-baseline log path, regression run). Bounded
defers carried forward (M4b scenarios debt). Commit: `docs(milestone-4c): F4c.4 close — evidence.md
+ guard live`.

## Hard-stop triggers

- **HS-2:** if any rule cannot be expressed without editing factory API or production-source.
- **HS-3:** if operator DSN fails / build fails before Step 8.
- **HS-6:** if Step 1 finds a stateful-write `pgtest` caller. Surface, do not silently migrate.

## Out of scope (deferred to F4c.5)

- Framework-level wiki page + ADR (F4c.5 owns the full architecture decision record).
- Any policy beyond the four R-rules.
