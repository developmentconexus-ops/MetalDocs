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
$minioContainerId = $null

function Resolve-MinioContainerId {
  $containerId = & docker compose -f $ComposeFile --env-file $EnvFile ps -q minio
  if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($containerId)) {
    throw "failed to resolve running minio container id from docker compose"
  }

  return $containerId.Trim()
}

function Invoke-McCommand {
  param(
    [string[]]$Arguments,
    [string[]]$VolumeArgs = @()
  )

  $dockerArgs = @(
    "run",
    "--rm",
    "--network", "container:$minioContainerId"
  ) + $VolumeArgs + @(
    "-e", "MC_HOST_local=http://$($env:MINIO_ROOT_USER):$($env:MINIO_ROOT_PASSWORD)@127.0.0.1:9000",
    $mcImage
  ) + $Arguments

  & docker @dockerArgs | Out-Host
  if ($LASTEXITCODE -ne 0) {
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
  Write-Host "[seed-system-blank-template] Ensuring MinIO services are available..."
  docker compose -f $ComposeFile --env-file $EnvFile up -d minio minio-init | Out-Host
  if ($LASTEXITCODE -ne 0) {
    throw "failed to start minio/minio-init"
  }
  $minioContainerId = Resolve-MinioContainerId

  if ($VerifyOnly) {
    Write-Host "[seed-system-blank-template] Verifying local/$Bucket/$StorageKey"
    Invoke-McCommand -Arguments @("stat", "local/$Bucket/$StorageKey")
    Write-Host "[seed-system-blank-template] Verified local/$Bucket/$StorageKey"
    exit 0
  }

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
