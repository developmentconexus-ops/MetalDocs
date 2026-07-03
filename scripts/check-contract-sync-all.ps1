# check-contract-sync-all.ps1 — F1.3 blocking CI wrapper.
#
# Iterates the four gated modules (templates, documents, controlleddocuments,
# taxonomy — see validation-contract.md §F1.3 for the approval carve-out:
# approval is UsesGeneratedBoundary=$false with ownership questions entangled
# in the M9 F9.5 approval-promotion decision, so it is intentionally excluded
# from this blocking set) and fails (non-zero exit) if any module reports
# [DRIFT]. Prints a per-module summary naming which module(s) drifted.

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$gatedModules = @('templates', 'documents', 'controlleddocuments', 'taxonomy')

$results = @()
$anyDrift = $false

foreach ($module in $gatedModules) {
    Write-Host "----- contract-sync: $module -----"
    $output = & pwsh -File (Join-Path $PSScriptRoot 'check-module-contract-sync.ps1') -Module $module 2>&1
    $exitCode = $LASTEXITCODE
    $output | ForEach-Object { Write-Host $_ }

    $driftLines = @($output | Where-Object { $_ -match '^\[DRIFT\]' })
    $moduleDrifted = ($exitCode -ne 0) -or ($driftLines.Count -gt 0)

    if ($moduleDrifted) {
        $anyDrift = $true
    }

    $results += [pscustomobject]@{
        Module     = $module
        ExitCode   = $exitCode
        DriftCount = $driftLines.Count
        DriftLines = $driftLines
    }
    Write-Host ""
}

Write-Host "===== contract-sync summary ====="
foreach ($r in $results) {
    $status = if (($r.ExitCode -ne 0) -or ($r.DriftCount -gt 0)) { 'DRIFT' } else { 'OK' }
    Write-Host ("[{0}] {1} (exit {2}, {3} drift line(s))" -f $status, $r.Module, $r.ExitCode, $r.DriftCount)
    foreach ($line in $r.DriftLines) {
        Write-Host ("    $line")
    }
}

if ($anyDrift) {
    $driftedModules = ($results | Where-Object { ($_.ExitCode -ne 0) -or ($_.DriftCount -gt 0) } | ForEach-Object { $_.Module }) -join ', '
    Write-Host "RESULT: contract-sync FAILED — drifted module(s): $driftedModules"
    exit 1
}

Write-Host "RESULT: contract-sync OK — zero DRIFT across $($gatedModules -join ', ')"
exit 0
