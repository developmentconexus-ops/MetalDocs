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
# Allow-model (REQ-TOP-1, realigned F9.5 — see wiki/decisions/<ADR>
# approval-nested-exception-and-boundary-model.md).
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

# documents/approval is a nested exception (ADR <NNNN>): it lives inside the
# documents bounded context, not as its own top-level module. Edges between
# documents and documents/approval (either direction) are INTERNAL to the
# documents module and are never flagged. External consumers (any other
# module, or documents' own delivery/http reaching into approval) may import
# ONLY approval's published surface -- domain/application/api are already
# covered by $allowedLayers; approval has no separate http/contracts package
# today (verified 2026-07-06), so the general allow-list already suffices.
# This list is where a NEW published approval package would be added if one
# is introduced later (promotion trigger recorded in the ADR).
$approvalPublishedExtra = @()

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

  # documents/approval is internal to the documents module for edge purposes.
  $currentIdentity = $currentModule
  if ($currentModule -eq "documents" -and $currentRest -match '^approval(/|$)') {
    $currentIdentity = "documents/approval"
  }

  $content = Get-Content $file.FullName -Raw
  $importMatches = [regex]::Matches($content, '"metaldocs/internal/modules/([^"]+)"')
  foreach ($match in $importMatches) {
    $importSuffix = $match.Groups[1].Value
    $importPath = "metaldocs/internal/modules/$importSuffix"

    # Determine target module identity, honoring the documents/approval nest.
    if ($importSuffix -match '^documents/approval(/|$)') {
      $targetIdentity = "documents/approval"
      $targetLayerPath = $importSuffix -replace '^documents/approval/?', ''
    } else {
      $parts = $importSuffix -split '/', 2
      $targetIdentity = $parts[0]
      $targetLayerPath = if ($parts.Length -gt 1) { $parts[1] } else { "" }
    }

    # Intra-module (including the documents<->approval nested exception): allowed.
    if ($targetIdentity -eq $currentIdentity) {
      continue
    }

    # documents <-> documents/approval is the nested exception: internal, allowed
    # in both directions regardless of layer.
    $bothInDocumentsFamily = (
      ($currentIdentity -eq "documents" -or $currentIdentity -eq "documents/approval") -and
      ($targetIdentity -eq "documents" -or $targetIdentity -eq "documents/approval")
    )
    if ($bothInDocumentsFamily) {
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
