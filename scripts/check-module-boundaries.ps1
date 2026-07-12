param()

$ErrorActionPreference = "Stop"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  if (Test-Path "C:\Program Files\Go\bin\go.exe") {
    $env:Path = "C:\Program Files\Go\bin;" + $env:Path
  }
}

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
  throw "Go toolchain nao encontrada no PATH."
}

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$moduleRoot = Join-Path $root "internal/modules"
if (-not (Test-Path $moduleRoot)) {
  throw "Diretorio internal/modules nao encontrado."
}

# ---------------------------------------------------------------------------
# Allow-model (REQ-TOP-1, realigned F9.5; approval promoted to first-class
# module by ADR 0082 which supersedes ADR 0072 — see
# wiki/decisions/0082-approval-kernel-extraction.md).
#
# REQ-TOP-1: cross-module access goes through a module's application service
# or published Go interface -- NEVER another module's repository, SQL, or
# domain internals. A cross-module import may target ONLY the owning
# module's PUBLISHED SURFACE:
#   - <module>/domain
#   - <module>/application
#   - <module>/api
#   - explicit published tool packages (interfaces meant for cross-module
#     consumption even though they are not literally named domain/application/api)
#
# Everything else -- <module>/repository, <module>/infrastructure,
# <module>/delivery, <module>/http, <module>/jobs -- is FORBIDDEN across a
# module boundary. This is stricter-or-equal to the old (domain-only) model:
# every violation class the old model could catch, this model still catches,
# plus it now also catches iam/authz-style false negatives are re-classified
# as explicitly sanctioned (see $publishedPackages) rather than silently
# passing because they didn't match "domain".
# ---------------------------------------------------------------------------

# Layer names (first path segment under a module) that are always an allowed
# cross-module target.
$allowedLayers = @("domain", "application", "api")

# Explicit published packages: real, verified interface-style packages that
# live one level deeper than <module>/<layer> but are intentionally exposed
# for cross-module consumption. Each entry is the module-relative import
# suffix (i.e. what follows "metaldocs/internal/modules/<module>/").
$publishedPackages = @(
  "authz",              # iam/authz -- tier-2 in-tx capability check, ADR 0022
  "fanout",             # render/fanout -- outbox client for async render dispatch
  "fanout/dispatchjobs", # render/fanout/dispatchjobs -- River job registrations
  "resolvers"           # render/resolvers -- field-resolver registry consumed by render callers
)

# NOTE (ADR 0082, supersedes ADR 0072): approval is now a first-class
# top-level module (internal/modules/approval), extracted from documents.
# The former documents/approval nested-exception special-casing is retired --
# approval is treated exactly like any other module. Its published surface is
# domain/application/api (covered by $allowedLayers). This is STRICTER than the
# old nested model: documents may no longer reach approval/http or
# approval/infrastructure, only its published layers, and vice-versa. If a NEW
# published approval package (one level deeper than a layer) is ever introduced,
# add it to $publishedPackages above, not via any documents-family bypass.

# Explicit debt allow-list: the ONLY sanctioned suppression mechanism for a
# true violation that cannot be mechanically fixed without a port/interface
# redesign (HS-2). Each entry MUST cite the ADR anchor recording why it
# exists and its fix trigger. As of F9.5 this list is EMPTY -- the violation
# sweep found zero true cross-module violations in production code once the
# repository/infrastructure rename landed (see ADR debt table for the audit
# trail). Do not add entries here without an ADR row.
#
# Shape: @{ From = "<relative file path, forward slashes>"; To = "metaldocs/internal/modules/<module>/<layer>" }
$debtAllowList = @(
  # (empty -- see wiki/decisions/<ADR> debt table)
)

function Test-DebtAllowed {
  param($relativePath, $importPath)
  foreach ($entry in $debtAllowList) {
    if ($entry.From -eq $relativePath -and $entry.To -eq $importPath) {
      return $true
    }
  }
  return $false
}

$violations = New-Object System.Collections.Generic.List[string]

$goFiles = Get-ChildItem -Path $moduleRoot -Recurse -Filter *.go -File `
  | Where-Object { $_.FullName -notmatch '_test\.go$' }

foreach ($file in $goFiles) {
  $fullName = (Resolve-Path $file.FullName).Path
  $rootWithSep = $root.TrimEnd('\') + '\'
  if ($fullName.StartsWith($rootWithSep, [System.StringComparison]::OrdinalIgnoreCase)) {
    $relativePath = $fullName.Substring($rootWithSep.Length).Replace("\", "/")
  } else {
    $relativePath = $fullName.Replace("\", "/")
  }
  if ($relativePath -notmatch '^internal/modules/([^/]+)/(.*)$') {
    continue
  }
  $currentModule = $Matches[1]
  $currentRest = $Matches[2]

  $currentIdentity = $currentModule

  $content = Get-Content $file.FullName -Raw
  $importMatches = [regex]::Matches($content, '"metaldocs/internal/modules/([^"]+)"')
  foreach ($match in $importMatches) {
    $importSuffix = $match.Groups[1].Value
    $importPath = "metaldocs/internal/modules/$importSuffix"

    # Determine target module identity.
    $parts = $importSuffix -split '/', 2
    $targetIdentity = $parts[0]
    $targetLayerPath = if ($parts.Length -gt 1) { $parts[1] } else { "" }

    # Intra-module: allowed.
    if ($targetIdentity -eq $currentIdentity) {
      continue
    }

    $targetLayerTop = ($targetLayerPath -split '/')[0]

    $isAllowedLayer = $allowedLayers -contains $targetLayerTop
    $isPublishedPackage = $publishedPackages -contains $targetLayerPath -or ($publishedPackages | Where-Object { $targetLayerPath -eq $_ -or $targetLayerPath.StartsWith("$_/") }).Count -gt 0

    if ($isAllowedLayer -or $isPublishedPackage) {
      continue
    }

    if (Test-DebtAllowed -relativePath $relativePath -importPath $importPath) {
      continue
    }

    $violations.Add("$relativePath -> $importPath")
  }
}

if ($violations.Count -gt 0) {
  Write-Host "[module-boundaries] FAIL"
  Write-Host "Violacoes encontradas:"
  foreach ($v in $violations) {
    Write-Host (" - " + $v)
  }
  exit 1
}

Write-Host "[module-boundaries] OK"
