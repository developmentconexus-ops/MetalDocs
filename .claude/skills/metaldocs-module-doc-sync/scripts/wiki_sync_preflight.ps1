param(
    [Parameter(Mandatory = $true)]
    [string]$Module,

    [string]$RepoRoot = (Get-Location).Path,

    [switch]$SkipTally
)

$ErrorActionPreference = "Stop"

function Test-File($Path) {
    return Test-Path -LiteralPath $Path -PathType Leaf
}

function Test-Dir($Path) {
    return Test-Path -LiteralPath $Path -PathType Container
}

$doc = Join-Path $RepoRoot "wiki/modules/$Module.md"
$debt = Join-Path $RepoRoot "wiki/modules/$Module-tech-debt.md"
$backlog = Join-Path $RepoRoot "wiki/backlog/$Module-refactor.md"
$artifacts = Join-Path $RepoRoot "wiki/modules/$Module/_artifacts"
$syncLog = Join-Path $artifacts "sync-log.md"
$tally = Join-Path $RepoRoot ".claude/skills/metaldocs-module-doc/scripts/tally_check.sh"

$gitBashCandidates = @(
    "C:\Program Files\Git\bin\bash.exe",
    "C:\Program Files\Git\usr\bin\bash.exe",
    "C:\Program Files (x86)\Git\bin\bash.exe"
)

$gitBash = $gitBashCandidates | Where-Object { Test-File $_ } | Select-Object -First 1

$result = [ordered]@{
    module = $Module
    doc = Test-File $doc
    techDebt = Test-File $debt
    backlog = Test-File $backlog
    artifacts = Test-Dir $artifacts
    syncLog = Test-File $syncLog
    tallyScript = Test-File $tally
    gitBash = $gitBash
    readyForSync = $false
    tallyStatus = "skipped"
    tallyOutput = @()
}

$result.readyForSync = $result.doc -and $result.techDebt -and $result.backlog

if (-not $SkipTally -and $result.readyForSync -and $result.tallyScript -and $gitBash) {
    $oldLocation = Get-Location
    try {
        Set-Location -LiteralPath $RepoRoot
        $output = & $gitBash ".claude/skills/metaldocs-module-doc/scripts/tally_check.sh" $Module 2>&1
        $result.tallyOutput = @($output)
        if ($LASTEXITCODE -eq 0) {
            $result.tallyStatus = "PASS"
        } else {
            $result.tallyStatus = "FAIL"
        }
    } finally {
        Set-Location -LiteralPath $oldLocation
    }
}

$result | ConvertTo-Json -Depth 4
