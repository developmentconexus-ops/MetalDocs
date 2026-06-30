param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env",
  [string]$StorageKey = "system/templates/blank.docx",
  [string]$Bucket,
  [switch]$VerifyOnly
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before seeding the system blank template."
}

Get-Content $EnvFile | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') {
    return
  }

  $name, $value = $_ -split '=', 2
  [System.Environment]::SetEnvironmentVariable($name, $value, 'Process')
}

if ([string]::IsNullOrWhiteSpace($Bucket)) {
  $Bucket = $env:METALDOCS_MINIO_BUCKET
}

if ([string]::IsNullOrWhiteSpace($Bucket)) {
  throw "METALDOCS_MINIO_BUCKET is required in $EnvFile or via -Bucket."
}

if ([string]::IsNullOrWhiteSpace($env:MINIO_ROOT_USER) -or [string]::IsNullOrWhiteSpace($env:MINIO_ROOT_PASSWORD)) {
  throw "MINIO_ROOT_USER and MINIO_ROOT_PASSWORD are required in $EnvFile."
}

$mcImage = "minio/mc:RELEASE.2024-04-18T16-45-29Z"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("metaldocs-system-blank-template-" + [System.Guid]::NewGuid().ToString("N"))
$docxPath = Join-Path $tempDir "blank.docx"
# Canonical container name (deploy/compose/docker-compose.yml pins
# container_name: metaldocs-minio), so we reference it directly for the network
# namespace instead of a hang-prone `docker compose ps -q` round-trip.
$minioContainerName = 'metaldocs-minio'

# Bounded docker invoker: runs the docker CLI inside a background job and
# enforces a hard timeout, so a stalled daemon or orchestration call (joining a
# container network namespace, an unresponsive Docker Desktop, etc.) fails fast
# with a clear message instead of hanging the script. $DockerArgs is passed
# in-memory to the job; when it carries MinIO creds (-e MC_HOST_local=...) those
# are never logged — only the docker command's own stdout/stderr is surfaced.
function Invoke-DockerBounded {
  param(
    [Parameter(Mandatory = $true)][string[]]$DockerArgs,
    [int]$TimeoutSeconds = 30,
    [string]$Label = 'docker command'
  )

  $job = Start-Job -ScriptBlock {
    param($innerArgs)
    $ErrorActionPreference = 'Continue'
    $out = & docker @innerArgs 2>&1 | Out-String
    [pscustomobject]@{ ExitCode = $LASTEXITCODE; Output = $out }
  } -ArgumentList (, $DockerArgs)

  try {
    if (-not (Wait-Job -Job $job -Timeout $TimeoutSeconds)) {
      throw "timed out after ${TimeoutSeconds}s on ${Label}; Docker appears stalled (check 'docker ps' and Docker Desktop)"
    }
    return Receive-Job -Job $job
  }
  finally {
    Stop-Job -Job $job -ErrorAction SilentlyContinue
    Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
  }
}

# True only when the canonical metaldocs-minio container exists and is running.
# Bounded so a stalled daemon cannot hang the probe.
function Test-MinioRunning {
  try {
    $res = Invoke-DockerBounded `
      -DockerArgs @('inspect', '-f', '{{.State.Running}}', $minioContainerName) `
      -TimeoutSeconds 10 `
      -Label "docker inspect $minioContainerName"
  }
  catch {
    Write-Host "[seed-system-blank-template] $($_.Exception.Message)"
    return $false
  }

  return ($res.ExitCode -eq 0 -and ([string]$res.Output).Trim() -eq 'true')
}

# Ensure the minio + minio-init services are up (bounded compose orchestration).
function Ensure-MinioUp {
  Write-Host "[seed-system-blank-template] Ensuring MinIO services are available..."
  $res = Invoke-DockerBounded `
    -DockerArgs @('compose', '-f', $ComposeFile, '--env-file', $EnvFile, 'up', '-d', 'minio', 'minio-init') `
    -TimeoutSeconds 120 `
    -Label 'docker compose up -d minio minio-init'
  if ($res.Output) { Write-Host $res.Output }
  if ($res.ExitCode -ne 0) {
    throw "failed to start minio/minio-init (docker compose up exited $($res.ExitCode))"
  }
}

function Invoke-McCommand {
  param(
    [string[]]$Arguments,
    [string[]]$VolumeArgs = @(),
    [int]$TimeoutSeconds = 30,
    [string]$Label
  )

  if ([string]::IsNullOrWhiteSpace($Label)) {
    $Label = "mc $($Arguments -join ' ')"
  }

  $dockerArgs = @(
    "run",
    "--rm",
    "--network", "container:$minioContainerName"
  ) + $VolumeArgs + @(
    "-e", "MC_HOST_local=http://$($env:MINIO_ROOT_USER):$($env:MINIO_ROOT_PASSWORD)@127.0.0.1:9000",
    $mcImage
  ) + $Arguments

  $res = Invoke-DockerBounded -DockerArgs $dockerArgs -TimeoutSeconds $TimeoutSeconds -Label $Label
  if ($res.Output) { Write-Host $res.Output }
  if ($res.ExitCode -ne 0) {
    throw "mc command failed: $($Arguments -join ' ')"
  }
}

# Read-only mc probe that runs *inside* the already-running minio container via
# `docker exec`, reusing its network namespace. This avoids the per-call
# `docker run` container create/destroy + network-namespace-join orchestration
# (observed at 6-30s and the source of the verify hang); exec lands in ~0.3s.
# Creds are passed via -e MC_HOST_local and are never logged.
function Invoke-McExec {
  param(
    [string[]]$Arguments,
    [int]$TimeoutSeconds = 30,
    [string]$Label
  )

  if ([string]::IsNullOrWhiteSpace($Label)) {
    $Label = "mc $($Arguments -join ' ')"
  }

  $dockerArgs = @(
    "exec",
    "-e", "MC_HOST_local=http://$($env:MINIO_ROOT_USER):$($env:MINIO_ROOT_PASSWORD)@127.0.0.1:9000",
    $minioContainerName,
    "mc"
  ) + $Arguments

  $res = Invoke-DockerBounded -DockerArgs $dockerArgs -TimeoutSeconds $TimeoutSeconds -Label $Label
  if ($res.Output) { Write-Host $res.Output }
  if ($res.ExitCode -ne 0) {
    throw "mc command failed: $($Arguments -join ' ')"
  }
}

function New-DeterministicBlankDocx {
  param(
    [string]$Path
  )

  Add-Type -AssemblyName System.IO.Compression
  Add-Type -AssemblyName System.IO.Compression.FileSystem

  $fixedTimestamp = [DateTimeOffset]::Parse("2026-05-17T00:00:00Z")
  $entries = [ordered]@{
    "[Content_Types].xml" = @'
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>
'@
    "_rels/.rels" = @'
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>
'@
    "word/document.xml" = @'
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p/>
    <w:sectPr/>
  </w:body>
</w:document>
'@
  }

  [System.IO.Directory]::CreateDirectory([System.IO.Path]::GetDirectoryName($Path)) | Out-Null
  if (Test-Path $Path) {
    Remove-Item -LiteralPath $Path -Force
  }

  $archive = [System.IO.Compression.ZipFile]::Open($Path, [System.IO.Compression.ZipArchiveMode]::Create)
  try {
    foreach ($entryPath in $entries.Keys) {
      $entry = $archive.CreateEntry($entryPath, [System.IO.Compression.CompressionLevel]::NoCompression)
      $entry.LastWriteTime = $fixedTimestamp
      $writer = New-Object System.IO.StreamWriter($entry.Open(), [System.Text.UTF8Encoding]::new($false))
      try {
        $writer.NewLine = "`n"
        $writer.Write($entries[$entryPath])
      } finally {
        $writer.Dispose()
      }
    }
  } finally {
    $archive.Dispose()
  }
}

try {
  if ($VerifyOnly) {
    # Fast, bounded verify: only bring MinIO up when it is not already running,
    # then probe the object by reusing the running container's network namespace.
    if (Test-MinioRunning) {
      Write-Host "[seed-system-blank-template] metaldocs-minio already running; skipping 'compose up'"
    }
    else {
      Write-Host "[seed-system-blank-template] metaldocs-minio not running; bringing MinIO up first"
      Ensure-MinioUp
    }

    Write-Host "[seed-system-blank-template] Verifying local/$Bucket/$StorageKey"
    Invoke-McExec -Arguments @("stat", "local/$Bucket/$StorageKey") -TimeoutSeconds 30
    Write-Host "[seed-system-blank-template] Verified local/$Bucket/$StorageKey"
    exit 0
  }

  # Seed path always ensures MinIO is up before writing the object.
  Ensure-MinioUp

  Write-Host "[seed-system-blank-template] Generating deterministic blank DOCX at $docxPath"
  New-DeterministicBlankDocx -Path $docxPath

  Write-Host "[seed-system-blank-template] Uploading local/$Bucket/$StorageKey"
  Invoke-McCommand -Arguments @("mb", "--ignore-existing", "local/$Bucket")
  Invoke-McCommand `
    -VolumeArgs @("-v", "${tempDir}:/seed:ro") `
    -Arguments @("cp", "/seed/blank.docx", "local/$Bucket/$StorageKey")
  Invoke-McCommand -Arguments @("stat", "local/$Bucket/$StorageKey")

  Write-Host "[seed-system-blank-template] Seeded local/$Bucket/$StorageKey"
}
finally {
  if (Test-Path $tempDir) {
    Remove-Item -LiteralPath $tempDir -Recurse -Force
  }
}
