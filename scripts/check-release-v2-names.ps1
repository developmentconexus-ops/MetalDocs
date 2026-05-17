param(
    [switch]$FailOnUnexpected,
    [string]$OutputPath = "docs/release/v2-name-inventory.md",
    [int]$MaxRows = 300
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$outputFullPath = [System.IO.Path]::GetFullPath((Join-Path $root $OutputPath))

$targets = @(
    "api",
    "apps",
    "db",
    "docs",
    "frontend/apps/web/src",
    "internal",
    "migrations",
    "wiki"
)

$matchRegex = "(?i)\bV2\b|_v2\b|/v2/|api/v2|templatesV2|templates_v2|documents_v2|docgenv2|docgen-v2|docgen_v2"
$excludedExtensions = @(
    ".exe", ".dll", ".so", ".dylib", ".class", ".jar",
    ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".pdf", ".zip", ".gz", ".tar", ".7z",
    ".woff", ".woff2", ".ttf", ".eot", ".mp3", ".mp4", ".mov", ".avi", ".bin"
)
$allowListRegexes = @(
    "(?i)github\.com/oapi-codegen/oapi-codegen/v2",
    "(?i)\boapi-codegen v2\b",
    "(?i)\bopenapi-typescript v7\b",
    "(?i)\bMDDM/v2\b"
)

function Get-AllowStatus {
    param([string]$Line)
    foreach ($rule in $allowListRegexes) {
        if ($Line -match $rule) { return $true }
    }
    return $false
}

$hits = New-Object System.Collections.Generic.List[object]

foreach ($target in $targets) {
    if (-not (Test-Path $target)) { continue }
    $files = Get-ChildItem -Path $target -Recurse -File -ErrorAction SilentlyContinue
    foreach ($file in $files) {
        if ([System.IO.Path]::GetFullPath($file.FullName) -eq $outputFullPath) { continue }
        if ($excludedExtensions -contains $file.Extension.ToLowerInvariant()) { continue }
        $matches = Select-String -Path $file.FullName -Pattern $matchRegex -AllMatches -SimpleMatch:$false -ErrorAction SilentlyContinue
        foreach ($m in $matches) {
            $lineText = $m.Line.Trim()
            $isAllowed = Get-AllowStatus -Line $lineText
            $relative = $file.FullName
            if ($relative.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
                $relative = $relative.Substring($root.Length).TrimStart('\')
            }
            $relative = $relative -replace "\\", "/"
            $hits.Add([pscustomobject]@{
                    Path    = $relative
                    Line    = $m.LineNumber
                    Allowed = $isAllowed
                    Text    = $lineText
                })
        }
    }
}

$sorted = $hits | Sort-Object Path, Line
$unexpected = $sorted | Where-Object { -not $_.Allowed }

$outDir = Split-Path -Parent $OutputPath
if ($outDir -and -not (Test-Path $outDir)) {
    New-Item -ItemType Directory -Path $outDir -Force | Out-Null
}

$lines = New-Object System.Collections.Generic.List[string]
$lines.Add("# Release V2 Name Inventory")
$lines.Add("")
$lines.Add("- Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss zzz")")
$lines.Add("- Scope: $($targets -join ', ')")
$lines.Add("- Match regex: " + $matchRegex)
$lines.Add("- Total hits: $($sorted.Count)")
$lines.Add("- Allowed hits: $(($sorted | Where-Object { $_.Allowed }).Count)")
$lines.Add("- Unexpected hits: $($unexpected.Count)")
$lines.Add("- Allowlist rules: github.com/oapi-codegen/oapi-codegen/v2, oapi-codegen v2, openapi-typescript v7, MDDM/v2")
$lines.Add("")
$lines.Add("## Unexpected Hits")
$lines.Add("")
$lines.Add("| Status | File | Line | Snippet |")
$lines.Add("|---|---|---:|---|")

$toWrite = $unexpected
if ($MaxRows -gt 0 -and $toWrite.Count -gt $MaxRows) {
    $toWrite = $toWrite | Select-Object -First $MaxRows
}

foreach ($hit in $toWrite) {
    $status = "unexpected"
    $snippet = ($hit.Text -replace "\|", "\\|" -replace "`r|`n", " ")
    $lines.Add("| $status | $($hit.Path) | $($hit.Line) | $snippet |")
}

if ($MaxRows -gt 0 -and $unexpected.Count -gt $MaxRows) {
    $lines.Add("")
    $lines.Add("_Output truncated to first $MaxRows unexpected rows. Use a larger -MaxRows for full inventory._")
}

[System.IO.File]::WriteAllLines((Join-Path $root $OutputPath), $lines, (New-Object System.Text.UTF8Encoding($false)))

Write-Host "V2 inventory written to $OutputPath"
Write-Host "Total hits: $($sorted.Count)"
Write-Host "Unexpected hits: $($unexpected.Count)"

if ($FailOnUnexpected -and $unexpected.Count -gt 0) {
    Write-Error "Unexpected V2 naming references found."
}
