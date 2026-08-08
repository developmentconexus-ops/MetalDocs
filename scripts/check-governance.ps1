param(
  # Default reads GITHUB_BASE_REF (GitHub Actions sets this automatically on
  # pull_request events to the PR's base branch name) rather than the
  # workflow passing a templated `${{ github.base_ref }}` argument — a
  # registry Check's Argv is compile-time literals only, no shell/YAML
  # interpolation (tools/verify/registry.go doc comment on Check.Argv), so
  # anything the check needs at runtime has to come from an env var. Outside
  # Actions (a laptop, or `push`) GITHUB_BASE_REF is unset, so this falls
  # back to "origin/main" — the same default this param always had.
  [string]$BaseRef = $(if ($env:GITHUB_BASE_REF) { "origin/$env:GITHUB_BASE_REF" } else { "origin/main" })
)

$ErrorActionPreference = "Stop"

# -w omits files whose diff is whitespace-only. Every rule below is of the
# form "if you touched X you must also touch Y", which is a proxy for "you
# changed behaviour". A reformat is provably not a behaviour change, so
# counting it makes the gate fire on a diff that cannot possibly violate what
# the rule protects. A gate that cries wolf teaches people to force past it,
# which costs more than the rule was worth. (Found for real: the A1 gofmt
# sweep touched 96 files and tripped the OpenAPI rule.)
$changed = git diff -w --name-only "$BaseRef...HEAD"
$changedText = ($changed -join "`n")

Write-Host "Changed files:"
Write-Host $changedText

function Fail([string]$msg) {
  Write-Error "[governance-check] $msg"
  exit 1
}

# API contract-impacting changes must update OpenAPI.
# We intentionally scope to delivery/http handlers and API spec files to avoid false positives
# for non-contract bootstrap changes in apps/api.
# _test.go is excluded deliberately: a test exercises the contract, it cannot
# change it. Demanding an openapi.yaml edit because a handler TEST moved is
# the same cry-wolf failure as the whitespace case above.
if ($changedText -match '(?m)^internal/modules/.+/delivery/http/.+(?<!_test)\.go$') {
  if ($changedText -notmatch '(?m)^api/openapi/v1/openapi.yaml$') {
    Fail "API contract change detected without OpenAPI update."
  }
}

if ($changedText -match '(?m)^internal/modules/') {
  if ($changedText -notmatch '(?m)^tests/') {
    Fail "Domain change detected without test updates under tests/."
  }
}

# The runbook rule protects one thing: if you change how the system is
# deployed or operated, the runbook must follow. scripts/ is no longer only
# ops — most of it is now CI verification code (check-*, api-lint, req-trace,
# cilint drivers), which changes nothing an operator does at 3am. Matching
# those made the rule demand a runbook edit for adding a linter, so the rule
# is scoped to what it actually claims to guard.
#
# This is a narrowing, not a relaxation: every path that could alter operating
# procedure still matches. If a future scripts/ file does affect operations,
# it belongs under one of the ops-shaped names below or the rule is wrong
# again — say so rather than adding an exception.
#
# Saying so, 2026-08-08. The rule was wrong again, and the shape of the
# mistake matters more than the fix. `check-|api-lint/|req-trace/` is a
# hand-maintained list of things that are NOT ops, so every new non-ops thing
# under scripts/ has to be *remembered* into it. CI Phase 0 added
# scripts/testdata/test-discipline/*.go.txt — fixtures for check-test-discipline
# — and the rule demanded a runbook edit for three static text files.
#
# `testdata/` is added as a category, not as a fourth name. It is Go's
# reserved directory name, ignored by the toolchain by definition: a path
# under testdata/ is input to a test, never code that runs in production and
# never something an operator can do at 3am. That is a property of the path,
# not a fact someone has to keep in sync.
#
# The enumeration is still inverted and still drifts — a new non-ops script
# that is neither check-, api-lint/, req-trace/ nor testdata/ will trip this
# again. Filed as an instance of the hand-synced-enumeration class in
# docs/engineering/mechanical-enforcement-register.md; the global maximum is
# deriving ops-ness from something declared, not pattern-matching paths.
$opsChanged = $changed | Where-Object {
  $_ -match '^deploy/' -or
  ($_ -match '^scripts/' -and $_ -notmatch '^scripts/(check-|api-lint/|req-trace/|testdata/)')
}
if ($opsChanged) {
  if ($changedText -notmatch '(?m)^docs/runbooks/') {
    Fail ("Infra/ops change detected without runbook update: " + ($opsChanged -join ", "))
  }
}

Write-Host "[governance-check] OK"
