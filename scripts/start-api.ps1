param(
    [switch]$Build,
    [switch]$NoWorker
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$apiCriticalPaths = @(
    'apps/api/cmd/metaldocs-api',
    'internal/modules',
    'internal/platform',
    'db',
    'scripts/start-api.ps1'
)

$workerCriticalPaths = @(
    'apps/worker/cmd/metaldocs-worker',
    'internal/modules',
    'internal/platform',
    'db',
    'scripts/start-api.ps1'
)

function Get-LatestWriteTimeUtc {
    param(
        [string[]]$RelativePaths
    )

    $latest = [DateTime]::MinValue

    foreach ($relativePath in $RelativePaths) {
        $fullPath = Join-Path $root $relativePath
        if (-not (Test-Path $fullPath)) {
            throw "Critical path not found: $relativePath"
        }

        $item = Get-Item $fullPath -ErrorAction Stop
        $candidates = if ($item.PSIsContainer) {
            Get-ChildItem -Path $fullPath -Recurse -File -Force -ErrorAction Stop
        } else {
            @($item)
        }

        foreach ($candidate in $candidates) {
            if ($candidate.LastWriteTimeUtc -gt $latest) {
                $latest = $candidate.LastWriteTimeUtc
            }
        }
    }

    return $latest
}

function Get-BinaryFreshness {
    param(
        [string]$BinaryPath,
        [string[]]$CriticalPaths
    )

    $latestSourceWriteTime = Get-LatestWriteTimeUtc -RelativePaths $CriticalPaths

    if (-not (Test-Path $BinaryPath)) {
        return @{
            IsStale = $true
            Reason = "missing"
            LatestSourceWriteTime = $latestSourceWriteTime
            BinaryWriteTime = $null
        }
    }

    $binaryWriteTime = (Get-Item $BinaryPath -ErrorAction Stop).LastWriteTimeUtc

    return @{
        IsStale = ($latestSourceWriteTime -gt $binaryWriteTime)
        Reason = "timestamp"
        LatestSourceWriteTime = $latestSourceWriteTime
        BinaryWriteTime = $binaryWriteTime
    }
}

function Build-GoBinary {
    param(
        [string]$Name,
        [string]$OutputName,
        [string]$PackagePattern
    )

    Write-Host "Building $OutputName..."
    go build -o $OutputName $PackagePattern
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for $Name"
        exit 1
    }
}

function Get-ProcessEvidence {
    param(
        [int]$ProcessId
    )

    $process = Get-Process -Id $ProcessId -ErrorAction SilentlyContinue
    if (-not $process) {
        return $null
    }

    $path = $null
    try {
        $path = $process.Path
    } catch {
        $path = $null
    }

    return @{
        Process = $process
        ProcessName = $process.ProcessName
        Path = $path
    }
}

function Test-IsOwnedApiProcess {
    param(
        [hashtable]$ProcessEvidence,
        [string]$ExpectedBinaryPath
    )

    if (-not $ProcessEvidence) {
        return $false
    }

    if ($ProcessEvidence.Path) {
        try {
            $resolvedHeldPath = (Resolve-Path $ProcessEvidence.Path -ErrorAction Stop).Path
            $resolvedExpectedPath = (Resolve-Path $ExpectedBinaryPath -ErrorAction Stop).Path
            if ($resolvedHeldPath -ieq $resolvedExpectedPath) {
                return $true
            }
        } catch {
        }
    }

    return $false
}

function Stop-ExistingWorkerProcesses {
    $existingWorkers = @(Get-Process -Name 'metaldocs-worker' -ErrorAction SilentlyContinue)
    if ($existingWorkers.Count -eq 0) {
        return
    }

    $workerIds = ($existingWorkers | Select-Object -ExpandProperty Id) -join ', '
    Write-Host "Detected existing metaldocs-worker process(es); stopping before starting a new worker (PID $workerIds)"
    $existingWorkers | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
}

function Ensure-GoBinary {
    param(
        [string]$Name,
        [string]$BinaryPath,
        [string]$OutputName,
        [string]$PackagePattern,
        [string[]]$CriticalPaths,
        [switch]$ForceBuild
    )

    if ($ForceBuild) {
        Write-Host "Forced rebuild requested for $OutputName"
        Build-GoBinary -Name $Name -OutputName $OutputName -PackagePattern $PackagePattern
        return
    }

    $freshness = Get-BinaryFreshness -BinaryPath $BinaryPath -CriticalPaths $CriticalPaths
    if (-not $freshness.IsStale) {
        return
    }

    if ($freshness.Reason -eq "missing") {
        Write-Host "$OutputName not found; building it now"
    } else {
        Write-Host (
            "$OutputName is stale; auto-rebuilding because tracked source timestamps are newer " +
            "(binary=$($freshness.BinaryWriteTime.ToString('u').Trim()) source=$($freshness.LatestSourceWriteTime.ToString('u').Trim()))"
        )
    }

    Build-GoBinary -Name $Name -OutputName $OutputName -PackagePattern $PackagePattern
}

Get-Content ".env" | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), 'Process')
}

[System.Environment]::SetEnvironmentVariable('APP_PORT', '8081', 'Process')

$binary = Join-Path $root "metaldocs-api.exe"
$workerBinary = Join-Path $root "metaldocs-worker.exe"

$held = netstat -ano 2>$null | Select-String ":8081 " | ForEach-Object { ($_ -split '\s+')[5] } | Select-Object -First 1
if ($held) {
    $heldId = [int]$held
    $heldProcess = Get-ProcessEvidence -ProcessId $heldId
    if (Test-IsOwnedApiProcess -ProcessEvidence $heldProcess -ExpectedBinaryPath $binary) {
        $heldDesc = if ($heldProcess.Path) {
            "$($heldProcess.ProcessName) ($($heldProcess.Path))"
        } else {
            $heldProcess.ProcessName
        }
        Write-Host "Detected the current MetalDocs API process on :8081; stopping $heldDesc (PID $heldId) before starting a replacement"
        Stop-Process -Id $heldId -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    } else {
        $heldName = if ($heldProcess) { $heldProcess.ProcessName } else { 'unknown-process' }
        $heldPath = if ($heldProcess -and $heldProcess.Path) { $heldProcess.Path } else { 'path unavailable' }
        Write-Error "Port :8081 is occupied by another process (PID $heldId, name=$heldName, path=$heldPath). Resolve the port conflict before starting MetalDocs API."
        exit 1
    }
}

Ensure-GoBinary `
    -Name "MetalDocs API" `
    -BinaryPath $binary `
    -OutputName "metaldocs-api.exe" `
    -PackagePattern "./apps/api/cmd/metaldocs-api/..." `
    -CriticalPaths $apiCriticalPaths `
    -ForceBuild:$Build

if (-not $NoWorker) {
    Ensure-GoBinary `
        -Name "MetalDocs Worker" `
        -BinaryPath $workerBinary `
        -OutputName "metaldocs-worker.exe" `
        -PackagePattern "./apps/worker/cmd/metaldocs-worker/..." `
        -CriticalPaths $workerCriticalPaths `
        -ForceBuild:$Build
}

if (-not $NoWorker) {
    Stop-ExistingWorkerProcesses

    Write-Host "Starting MetalDocs Worker in background..."
    $workerProc = Start-Process -FilePath $workerBinary -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 1
    if ($workerProc.HasExited) {
        Write-Error "Worker process exited immediately after startup"
        exit 1
    }
    Write-Host ("Worker process launched and remained alive for the initial check window (PID: " + $workerProc.Id + ")")
}

Write-Host "Starting MetalDocs API on :8081 after timestamp-based binary checks"
& $binary
