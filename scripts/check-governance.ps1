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

# Line-comment syntax per file type. The exemption below is opt-in: a type
# absent from this table has no marker, so its diff always counts as real.
# Only types that occur under the trees the rules police need an entry.
$CommentMarker = @{
  '.go'         = '//'
  '.js'         = '//'
  '.mjs'        = '//'
  '.ps1'        = '#'
  '.sh'         = '#'
  '.yaml'       = '#'
  '.yml'        = '#'
  '.jq'         = '#'
  '.dockerfile' = '#'
  '.sql'        = '--'
}

# True when the file's diff contains a real code change; false when every
# changed line differs only in line comments (e.g. an appended
# "// #nosec Gxxx -- reason" suppression, or a rewritten rationale block).
# A comment cannot change an HTTP contract or an operating procedure —
# failing on one is the same cry-wolf class as the whitespace and _test.go
# exemptions.
#
# The marker is chosen by file type, not fixed to Go: the claim being made
# is about comments, and a `.jq` or `.ps1` comment is no more capable of
# changing an operating procedure than a `.go` one. Restricting it to "//"
# meant the rule fired on a diff that provably could not violate it — the
# exact failure this function exists to prevent, reintroduced by treating
# one language's syntax as the concept. A marker inside a string literal is
# also stripped, which can only make a rule MISS a change hidden in such a
# line; any other difference still fails closed, as does any file type with
# no entry in the table above.
function Test-RealCodeChange([string]$file) {
  $ext = [System.IO.Path]::GetExtension($file).ToLowerInvariant()
  if ($file -match '(^|/)Dockerfile(\.|$)') { $ext = '.dockerfile' }
  $marker = $CommentMarker[$ext]
  if (-not $marker) { return $true }
  $stripPattern = [regex]::Escape($marker) + '.*$'
  $diffLines = git diff -w "$BaseRef...HEAD" -- $file |
    Where-Object { ($_ -match '^[+-]') -and ($_ -notmatch '^(\+\+\+|---)') }
  $strip = { param($s) ($s.Substring(1) -replace $stripPattern, '').TrimEnd() }
  $added   = @($diffLines | Where-Object { $_ -match '^\+' } | ForEach-Object { & $strip $_ } | Where-Object { $_ -ne '' } | Sort-Object)
  $removed = @($diffLines | Where-Object { $_ -match '^-' } | ForEach-Object { & $strip $_ } | Where-Object { $_ -ne '' } | Sort-Object)
  return (($added -join "`n") -ne ($removed -join "`n"))
}

# A rule below asks a human-judgment question ("did this change the
# contract?"), so it needs a human-answer channel that is not a bypass:
# an auditable waiver. A waiver is valid ONLY for the diff that ADDS it —
# this checks `git diff $BaseRef...HEAD` of the waiver file, never its merged
# contents, so merged entries are permanent audit records that waive nothing
# for later PRs. The reviewer sees the waiver claim in the PR diff and can
# contest it. Default stays fail-closed: no added entry, rule fails as before.
function Test-Waived([string]$ruleId) {
  $added = git diff -w "$BaseRef...HEAD" -- scripts/check-governance-waivers.txt |
    Where-Object { ($_ -match '^\+') -and ($_ -notmatch '^\+\+\+') } |
    ForEach-Object { $_.Substring(1) }
  $entries = @($added | Where-Object { $_ -match "^\s*$([regex]::Escape($ruleId))\s*\|" })
  if ($entries.Count -gt 0) {
    Write-Host "[governance-check] rule '$ruleId' waived by entry added in this diff:"
    $entries | ForEach-Object { Write-Host "  $_" }
    return $true
  }
  return $false
}

# API contract-impacting changes must update OpenAPI.
# We intentionally scope to delivery/http handlers and API spec files to avoid false positives
# for non-contract bootstrap changes in apps/api.
# _test.go is excluded deliberately: a test exercises the contract, it cannot
# change it. Demanding an openapi.yaml edit because a handler TEST moved is
# the same cry-wolf failure as the whitespace case above.
#
# TRANSITIONAL RULE (labelled per CLAUDE.md "Global Maximum, Not Local
# Maximum"). This is a heuristic proxy: "handler file changed" ≈ "contract may
# have changed". It exists because only 4 of 15 modules sit behind the
# generated boundary (the contract-sync gated set); in every other module a
# hand-written handler CAN change wire behaviour (status, shape) without
# touching the spec, and no deterministic check catches that. Global-maximum
# structure: generated boundary in every module — typed responses make wire
# drift unrepresentable, contract-sync goes total, and this rule plus its
# waiver file are DELETED. Deletion milestone: contract-sync covers all 15
# modules (ROADMAP: generated-boundary expansion). Because the rule asks a
# human-judgment question, it carries a human-answer channel — see
# Test-Waived and scripts/check-governance-waivers.txt.
if ($changedText -match '(?m)^internal/modules/.+/delivery/http/.+(?<!_test)\.go$') {
  if ($changedText -notmatch '(?m)^api/openapi/v1/openapi.yaml$') {
    # Comment-only exemption — see Test-RealCodeChange (found for real: the
    # gosec triage appended "// #nosec" comments in routes_mapping.go and
    # tripped this rule).
    $handlerFiles = $changed | Where-Object {
      $_ -match '^internal/modules/.+/delivery/http/' -and $_ -match '\.go$' -and $_ -notmatch '_test\.go$'
    }
    $realChange = @($handlerFiles | Where-Object { Test-RealCodeChange $_ })
    if ($realChange.Count -gt 0 -and -not (Test-Waived 'api-contract-openapi')) {
      Fail "API contract change detected without OpenAPI update. If the handler change provably does not alter the wire contract, add a justified waiver line to scripts/check-governance-waivers.txt (see its header)."
    }
  }
}

# Same judgment-question class as the contract rule above ("does this domain
# change need new tests?"), so it shares the same adjudication channel. A
# mechanical refactor (lint sweep, rename, decomposition) whose behaviour is
# already pinned by the existing suites has no honest tests/ edit to make.
if ($changedText -match '(?m)^internal/modules/') {
  if ($changedText -notmatch '(?m)^tests/') {
    if (-not (Test-Waived 'domain-tests')) {
      Fail "Domain change detected without test updates under tests/. If the change provably needs no new tests (behaviour pinned by existing suites), add a justified waiver line to scripts/check-governance-waivers.txt (see its header)."
    }
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
  ($_ -match '^deploy/' -or
    ($_ -match '^scripts/' -and $_ -notmatch '^scripts/(check-|api-lint/|req-trace/|testdata/)')) -and
  (Test-RealCodeChange $_)  # comment-only edits cannot change an operating procedure
}
if ($opsChanged) {
  if ($changedText -notmatch '(?m)^docs/runbooks/') {
    Fail ("Infra/ops change detected without runbook update: " + ($opsChanged -join ", "))
  }
}

Write-Host "[governance-check] OK"
