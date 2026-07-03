<#
.SYNOPSIS
  Mechanical tally check for a wiki module doc + its tech-debt register (PowerShell-native).

.DESCRIPTION
  Re-implements the retired bash `tally_check.sh` (formerly
  .claude/skills/metaldocs-module-doc/scripts/tally_check.sh, deleted in
  02ed1c24 alongside the legacy metaldocs-* skill trees) without shelling
  out to Git Bash. The original PowerShell wrapper
  (.claude/skills/metaldocs-module-doc-sync/scripts/wiki_sync_preflight.ps1,
  also deleted in 02ed1c24) invoked the bash script via Git Bash, which
  fails under this repo path with `CreateFileMapping ... Win32 error 5`
  (recorded across documents/iam sync-log.md — TST-02). This script performs
  the same checks natively in PowerShell so the gate no longer depends on
  Git Bash at all.

  Checks performed:
    1. Severity counts stated in `wiki/modules/<module>.md` section 11
       (Critical/Major/Minor) match the count of `- **Severity:** <word>`
       lines in `wiki/modules/<module>-tech-debt.md` (ALL rows counted,
       including closed — this matches current register convention, see
       documents.md:527 "all rows counted by tally including closed").
    2. "Decisions without ADR link: N" stated in the module doc matches the
       count of `missing-ADR` occurrences in the tech-debt register.

  Backlog-linkage checks (3/4 in the original script) are intentionally
  DROPPED: the `wiki/backlog/<module>-refactor.md` file with a `T-NNN`
  debt_id column no longer exists for most modules (current backlog/ dir
  holds feature-shaped docs, not a debt_id-keyed table) — that check would
  false-fail on every module today. If backlog-linkage tracking is wanted
  again it needs a fresh spec against the current backlog conventions;
  out of scope for TST-02 (tooling repair only, no register-format changes).

.PARAMETER Module
  Module slug, e.g. "documents", "iam". Must resolve to:
    wiki/modules/<Module>.md
    wiki/modules/<Module>-tech-debt.md

.PARAMETER RepoRoot
  Repo root to resolve paths from. Defaults to current directory.

.EXAMPLE
  ./scripts/wiki-tally-check.ps1 -Module documents

.EXAMPLE
  ./scripts/wiki-tally-check.ps1 -Module iam -RepoRoot C:\path\to\MetalDocs
#>
param(
    [Parameter(Mandatory = $true)]
    [string]$Module,

    [string]$RepoRoot = (Get-Location).Path
)

$ErrorActionPreference = "Stop"

# --- Exempt modules ----------------------------------------------------------
# Modules whose register deliberately does NOT use the T-NNN + `- **Severity:**`
# row shape this script tallies. Exemption != drift: forcing a fake severity
# taxonomy onto these would be format-washing, and renumbering would break
# existing cross-references (e.g. the grade-A register cites "security #7").
# This list MUST only shrink; additions require operator approval + reason.
$ExemptModules = @{
    # security-tech-debt.md is a numbered defer-table (| # | Item | Reason for
    # defer | Promote when |) and security.md is a narrow single-topic doc with
    # no section 11 — by design (2026-07-02 Phase-5 reconciliation verdict).
    'security' = 'defer-table register shape; no section-11 severity taxonomy by design'
}
if ($ExemptModules.ContainsKey($Module)) {
    Write-Host "[tally] SKIP (exempt: $($ExemptModules[$Module]))"
    exit 0
}

$doc = Join-Path $RepoRoot "wiki/modules/$Module.md"
$debt = Join-Path $RepoRoot "wiki/modules/$Module-tech-debt.md"

$fail = $false

foreach ($f in @($doc, $debt)) {
    if (-not (Test-Path -LiteralPath $f -PathType Leaf)) {
        Write-Host "FAIL: missing file $f"
        exit 1
    }
}

$docLines = Get-Content -LiteralPath $doc -Encoding UTF8
$debtLines = Get-Content -LiteralPath $debt -Encoding UTF8

# --- Check 1: severity counts ----------------------------------------------
# Register: lines like "- **Severity:** critical" / "- **Severity:** ~~minor~~ closed (...)"
# Count ALL rows regardless of open/closed suffix (matches documented convention).
$sevPattern = '^-\s+\*\*Severity:\*\*\s+~{0,2}(critical|major|minor)\b'
$sevMatches = $debtLines | Select-String -Pattern $sevPattern
$critActual = ($sevMatches | Where-Object { $_.Matches[0].Groups[1].Value -eq 'critical' }).Count
$majActual  = ($sevMatches | Where-Object { $_.Matches[0].Groups[1].Value -eq 'major' }).Count
$minActual  = ($sevMatches | Where-Object { $_.Matches[0].Groups[1].Value -eq 'minor' }).Count

$critStated = $null
$majStated  = $null
$minStated  = $null
foreach ($line in $docLines) {
    if ($line -match '^-\s*Critical:\s*(\d+)') { $critStated = [int]$Matches[1] }
    elseif ($line -match '^-\s*Major:\s*(\d+)') { $majStated = [int]$Matches[1] }
    elseif ($line -match '^-\s*Minor:\s*(\d+)') { $minStated = [int]$Matches[1] }
}

$critStatedDisplay = if ($null -ne $critStated) { $critStated } else { '?' }
$majStatedDisplay  = if ($null -ne $majStated)  { $majStated }  else { '?' }
$minStatedDisplay  = if ($null -ne $minStated)  { $minStated }  else { '?' }

Write-Host "[tally] severities -- actual C/M/m: $critActual/$majActual/$minActual - stated: $critStatedDisplay/$majStatedDisplay/$minStatedDisplay"

if ($critStated -ne $critActual -or $majStated -ne $majActual -or $minStated -ne $minActual) {
    Write-Host "FAIL: section 11 severity counts in $doc do not match register $debt"
    $fail = $true
}

# --- Check 2: missing-ADR count --------------------------------------------
$adrMissingCount = ($debtLines | Select-String -Pattern 'missing-ADR').Count
$adrStated = $null
foreach ($line in $docLines) {
    if ($line -match 'Decisions without ADR link:\s*(\d+)') {
        $adrStated = [int]$Matches[1]
        break
    }
}
$adrStatedDisplay = if ($null -ne $adrStated) { $adrStated } else { '?' }
Write-Host "[tally] missing-ADR -- register count: $adrMissingCount - doc stated: $adrStatedDisplay"

if ($null -ne $adrStated -and $adrStated -ne $adrMissingCount) {
    Write-Host "FAIL: missing-ADR count in $doc coverage-stats disagrees with register $debt"
    $fail = $true
}

# --- Summary ----------------------------------------------------------------
if (-not $fail) {
    Write-Host "[tally] PASS"
    exit 0
}
Write-Host "[tally] FAIL -- fix doc or register before closing sync"
exit 1
