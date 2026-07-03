param(
    [Parameter(Mandatory = $true)]
    [string]$Module
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

function New-ModuleConfig {
    param(
        [string]$ModuleName
    )

    switch ($ModuleName.ToLowerInvariant()) {
        'templates' {
            return @{
                ModuleName = 'templates'
                RuntimeFile = 'internal/modules/templates/delivery/http/handler.go'
                RuntimeOwnerFiles = @('internal/modules/templates/delivery/http/handler.go')
                RuntimePatterns = @('templatesapi.HandlerWithOptions', '/api/v1/templates')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                # AD-1: spec path keys are relative (servers.url already carries /api/v1).
                OpenApiPatterns = @('/templates:', '/templates/{id}/versions/{n}/submit:', '/templates/placeholder-catalog:')
                BackendFile = 'internal/modules/templates/api/api.gen.go'
                BackendPatterns = @('ListTemplates', 'CreateTemplate', 'GetTemplate', 'SubmitTemplateVersion')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                # Generated index.d.ts keys are also relative (same AD-1 migration).
                FrontendTypesPatterns = @('"/templates":', '"/templates/{id}/versions/{n}/submit":', '"/templates/placeholder-catalog":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/templates/api/templates.ts'
                FrontendWrapperPatterns = @('/api/v1/templates', 'submitForReview', 'approveVersion', 'putTemplateSchemas')
                RouteFamilies = @('/api/v1/templates')
                WikiFile = 'wiki/modules/templates.md'
                UsesGeneratedBoundary = $true
                OwnedOpenApiPaths = @('/api/v1/templates')
                # Finite, named, shrink-only allowlist (same idiom as css-token-discipline /
                # test-discipline) for local wire-adapter types that the F1.4 contract
                # (docs/superpowers/milestones/global-maximum-remediation/milestone-1-contract-fe-gates/
                # validation-contract.md, F1.4 Part B) explicitly adjudicated as genuine local
                # view types, NOT overrides of a generated schema — the generated DTO types
                # placeholder_schema as untyped JSON, so a local wire-format view type is
                # required at that one call site. NOT a general escape hatch: adding an entry
                # here requires the same explicit adjudication F1.4 gave these two. Owner:
                # Leandro; trigger: remove an entry only when the type is deleted or replaced
                # by a generated shape (never add without an equivalent recorded decision).
                FrontendTypeAllowlist = @('TemplateSchemas', 'WirePlaceholder')
            }
        }
        'approval' {
            return @{
                ModuleName = 'approval'
                RuntimeFile = 'internal/modules/documents/approval/http/router.go'
                RuntimeOwnerFiles = @('internal/modules/documents/approval/http/router.go')
                RuntimePatterns = @('/api/v1/approval/inbox', '/api/v1/documents/{id}/signoff', '/api/v1/approval/routes', 'approvalapi.ServerInterfaceWrapper')
                RuntimeForbiddenPatterns = @('h.SubmitHandler', 'h.SignoffHandler', 'h.PublishHandler', 'h.SchedulePublishHandler')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/api/v1/approval/inbox:', '/api/v1/documents/{id}/signoff:', '/api/v1/documents/{id}/cancel:', '/api/v1/approval/routes:')
                BackendFile = 'internal/modules/documents/approval/api/api.gen.go'
                BackendPatterns = @('ListApprovalInbox', 'ListApprovalRoutes', 'CreateApprovalRoute')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/api/v1/approval/inbox":', '"/api/v1/documents/{id}/signoff":', '"/api/v1/documents/{id}/cancel":', '"/api/v1/approval/routes":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/approval/api/approvalApi.ts'
                FrontendWrapperPatterns = @('/approval/inbox', '/documents/${documentId}/signoff', '/documents/${documentId}/cancel', '/approval/routes', 'listInbox', 'signoff', 'cancel', 'createRoute', 'listRoutes')
                RouteFamilies = @('/api/v1/approval', '/api/v1/documents')
                OwnedOpenApiPaths = @(
                    '/api/v1/approval/inbox',
                    '/api/v1/approval/instances/{instance_id}',
                    '/api/v1/approval/instances/{instance_id}/cancel',
                    '/api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs',
                    '/api/v1/approval/routes',
                    '/api/v1/approval/routes/{id}',
                    '/api/v1/documents/{id}/approval-instance',
                    '/api/v1/documents/{id}/cancel',
                    '/api/v1/documents/{id}/obsolete',
                    '/api/v1/documents/{id}/publish',
                    '/api/v1/documents/{id}/schedule-publish',
                    '/api/v1/documents/{id}/signoff',
                    '/api/v1/documents/{id}/submit',
                    '/api/v1/documents/{id}/supersede'
                )
                UsesGeneratedBoundary = $false
                WikiFile = 'wiki/modules/approval.md'
            }
        }
        'documents' {
            return @{
                ModuleName = 'documents'
                # Runtime-truth (verified): the generated-boundary mount call
                # (documentsapi.HandlerWithOptions) lives in delivery/http/handler.go,
                # not module.go (module.go only wires deps + calls RegisterRoutes).
                RuntimeFile = 'internal/modules/documents/delivery/http/handler.go'
                RuntimeOwnerFiles = @(
                    'internal/modules/documents/delivery/http/handler.go',
                    'internal/modules/documents/delivery/http/export_handler.go',
                    'internal/modules/documents/http/fillin_handler.go',
                    'internal/modules/documents/http/placeholder_options_handler.go',
                    'internal/modules/documents/http/reconstruct_handler.go',
                    'internal/modules/documents/http/view_handler.go'
                )
                RuntimePatterns = @('documentsapi.HandlerWithOptions')
                RuntimeForbiddenPatterns = @('mux.HandleFunc("GET /api/v1/documents"', 'mux.HandleFunc("POST /api/v1/documents/{id}/finalize"')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                # AD-1: spec path keys are relative.
                OpenApiPatterns = @('/documents:', '/documents/{id}:', '/documents/{id}/finalize:')
                # The old forbid ("`n  /documents:") matched the CORRECT relative key
                # post-AD-1, inverting the check (it would always DRIFT on a clean spec).
                # No legacy-duplicate hazard remains: PATH-BASE-PREFIX (scripts/api-lint)
                # already rejects any /api/v1-prefixed path key, so an absolute-form
                # duplicate cannot exist in a passing spec. Forbid retired.
                OpenApiForbiddenPatterns = @()
                BackendFile = 'internal/modules/documents/api/api.gen.go'
                BackendPatterns = @('ListDocuments', 'GetDocument', 'FinalizeDocument')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/documents":', '"/documents/{id}":', '"/documents/{id}/finalize":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/documents/api/documents.ts'
                FrontendWrapperPatterns = @('/api/v1/documents', 'listDocuments', 'getDocument', 'finalizeDocument')
                RouteFamilies = @('/api/v1/documents')
                OwnedOpenApiPaths = @('/api/v1/documents', '/api/v1/documents/{id}', '/api/v1/documents/{id}/finalize')
                UsesGeneratedBoundary = $true
                WikiFile = 'wiki/modules/documents.md'
            }
        }
        'controlleddocuments' {
            return @{
                ModuleName = 'controlleddocuments'
                RuntimeFile = 'internal/modules/controlleddocuments/delivery/http/handler.go'
                RuntimeOwnerFiles = @('internal/modules/controlleddocuments/delivery/http/handler.go')
                RuntimePatterns = @('/api/v1/controlled-documents', 'controlleddocumentsapi.HandlerWithOptions', 'ErrorHandlerFunc')
                RuntimeForbiddenPatterns = @('mux.Handle("GET /api/v1/controlled-documents", generated.ListControlledDocuments', 'mux.Handle("GET /api/v1/controlled-documents/{id}", generated.GetControlledDocument')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                # AD-1: spec path keys are relative.
                OpenApiPatterns = @('/controlled-documents:', '/controlled-documents/{id}:', '/controlled-documents/{id}/revisions:')
                BackendFile = 'internal/modules/controlleddocuments/api/api.gen.go'
                BackendPatterns = @('ListControlledDocuments', 'AtomicCreateControlledDocument', 'GetControlledDocument')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/controlled-documents":', '"/controlled-documents/{id}":', '"/controlled-documents/{id}/revisions":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts'
                FrontendWrapperPatterns = @('/api/v1/controlled-documents', 'fetchControlledDocuments', 'createControlledDocumentAtomic')
                RouteFamilies = @('/api/v1/controlled-documents')
                OwnedOpenApiPaths = @('/api/v1/controlled-documents', '/api/v1/controlled-documents/{id}', '/api/v1/controlled-documents/{id}/revisions')
                UsesGeneratedBoundary = $true
                WikiFile = 'wiki/modules/controlled-documents.md'
            }
        }
        'controlled-documents' {
            return @{
                ModuleName = 'controlled-documents'
                RuntimeFile = 'internal/modules/controlleddocuments/delivery/http/handler.go'
                RuntimeOwnerFiles = @('internal/modules/controlleddocuments/delivery/http/handler.go')
                RuntimePatterns = @('/api/v1/controlled-documents', 'controlleddocumentsapi.HandlerWithOptions', 'ErrorHandlerFunc')
                RuntimeForbiddenPatterns = @('mux.Handle("GET /api/v1/controlled-documents", generated.ListControlledDocuments', 'mux.Handle("GET /api/v1/controlled-documents/{id}", generated.GetControlledDocument')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/controlled-documents:', '/controlled-documents/{id}:', '/controlled-documents/{id}/revisions:')
                BackendFile = 'internal/modules/controlleddocuments/api/api.gen.go'
                BackendPatterns = @('ListControlledDocuments', 'AtomicCreateControlledDocument', 'GetControlledDocument')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/controlled-documents":', '"/controlled-documents/{id}":', '"/controlled-documents/{id}/revisions":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts'
                FrontendWrapperPatterns = @('/api/v1/controlled-documents', 'fetchControlledDocuments', 'createControlledDocumentAtomic')
                RouteFamilies = @('/api/v1/controlled-documents')
                OwnedOpenApiPaths = @('/api/v1/controlled-documents', '/api/v1/controlled-documents/{id}', '/api/v1/controlled-documents/{id}/revisions')
                UsesGeneratedBoundary = $true
                WikiFile = 'wiki/modules/controlled-documents.md'
            }
        }
        'taxonomy' {
            return @{
                ModuleName = 'taxonomy'
                RuntimeFile = 'internal/modules/taxonomy/delivery/http/handler.go'
                RuntimeOwnerFiles = @('internal/modules/taxonomy/delivery/http/handler.go')
                RuntimePatterns = @('taxonomyapi.HandlerWithOptions', 'BaseRouter: mux')
                RuntimeForbiddenPatterns = @('mux.HandleFunc("GET /api/v1/taxonomy/profiles"', 'mux.HandleFunc("POST /api/v1/taxonomy/families"')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                # AD-1: spec path keys are relative.
                OpenApiPatterns = @('/taxonomy/profiles:', '/taxonomy/areas:', '/taxonomy/families:')
                BackendFile = 'internal/modules/taxonomy/api/api.gen.go'
                BackendPatterns = @('ListTaxonomyProfiles', 'ListTaxonomyAreas', 'ListTaxonomyFamilies')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/taxonomy/profiles":', '"/taxonomy/areas":', '"/taxonomy/families":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/taxonomy/api/taxonomy.ts'
                # Wrapper now calls the generated `api` client (openapi-fetch) with
                # relative operation paths directly (api.GET("/taxonomy/profiles", ...))
                # instead of a hand-rolled BASE constant.
                FrontendWrapperPatterns = @('api.GET("/taxonomy/profiles"', 'api.GET("/taxonomy/areas"', 'api.GET("/taxonomy/families"')
                RouteFamilies = @('/api/v1/taxonomy')
                OwnedOpenApiPaths = @('/api/v1/taxonomy/profiles', '/api/v1/taxonomy/areas', '/api/v1/taxonomy/families')
                UsesGeneratedBoundary = $true
                WikiFile = 'wiki/modules/taxonomy.md'
            }
        }
        'registry' {
            throw "Module 'registry' was renamed to 'controlleddocuments'. Re-run with -Module controlleddocuments."
        }
        default {
            $basePath = "/api/v1/$ModuleName"
            return @{
                ModuleName = $ModuleName
                RuntimeFile = "internal/modules/$ModuleName/delivery/http/handler.go"
                RuntimeOwnerFiles = @("internal/modules/$ModuleName/delivery/http/handler.go")
                RuntimePatterns = @($basePath)
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @("${basePath}:")
                BackendFile = "internal/modules/$ModuleName/api/api.gen.go"
                BackendPatterns = @($basePath)
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @("""$basePath"":")
                FrontendWrapperFile = "frontend/apps/web/src/features/$ModuleName/api/${ModuleName}V2.ts"
                FrontendWrapperPatterns = @($basePath)
                RouteFamilies = @($basePath)
                OwnedOpenApiPaths = @($basePath)
                UsesGeneratedBoundary = $false
                WikiFile = "wiki/modules/$ModuleName.md"
            }
        }
    }
}

function Get-OpenApiPaths {
    param([string]$RelativePath)
    $fullPath = Join-Path $root $RelativePath
    if (-not (Test-Path $fullPath)) { return @() }
    $content = Get-Content $fullPath -Raw
    $matches = [regex]::Matches($content, '(?m)^\s{2}(/[^:]+):\s*$')
    $paths = New-Object System.Collections.Generic.HashSet[string]
    foreach ($m in $matches) { [void]$paths.Add($m.Groups[1].Value.Trim()) }
    return @($paths)
}

function Get-RuntimePaths {
    param([string[]]$RelativePaths)
    $paths = New-Object System.Collections.Generic.HashSet[string]
    foreach ($relativePath in $RelativePaths) {
        $fullPath = Join-Path $root $relativePath
        if (-not (Test-Path $fullPath)) { continue }
        $content = Get-Content $fullPath -Raw
        $matches = [regex]::Matches($content, '/api/v1/[A-Za-z0-9\-/\{\}_]+')
        foreach ($m in $matches) {
            $candidate = $m.Value
            if ($candidate.Length -gt 1 -and $candidate.EndsWith('/')) {
                $candidate = $candidate.TrimEnd('/')
            }
            [void]$paths.Add($candidate)
        }
    }
    return @($paths)
}

function Get-OwnedOpenApiPaths {
    param(
        [string[]]$ConfiguredOwnedPaths,
        [string]$OpenApiFile,
        [string[]]$Families
    )
    if ($ConfiguredOwnedPaths -and $ConfiguredOwnedPaths.Count -gt 0) {
        return @($ConfiguredOwnedPaths)
    }
    return Select-ModuleFamilyPaths -AllPaths (Get-OpenApiPaths -RelativePath $OpenApiFile) -Families $Families
}

function Select-ModuleFamilyPaths {
    param(
        [string[]]$AllPaths,
        [string[]]$Families
    )
    $selected = New-Object System.Collections.Generic.HashSet[string]
    foreach ($path in $AllPaths) {
        foreach ($family in $Families) {
            if ($path.StartsWith($family, [System.StringComparison]::Ordinal)) {
                [void]$selected.Add($path)
                break
            }
        }
    }
    return @($selected)
}

function Test-DuplicateCanonicalFamilies {
    param(
        [string]$OpenApiFile,
        [string[]]$Families
    )
    $openApiPaths = Get-OpenApiPaths -RelativePath $OpenApiFile
    $familyPaths = Select-ModuleFamilyPaths -AllPaths $openApiPaths -Families $Families
    $allOpenApi = New-Object System.Collections.Generic.HashSet[string]
    foreach ($p in $openApiPaths) { [void]$allOpenApi.Add($p) }

    $duplicates = New-Object System.Collections.Generic.List[string]
    foreach ($path in $familyPaths) {
        $canonicalSuffix = $path.Substring('/api/v1'.Length)
        $legacyPath = $canonicalSuffix
        if ($legacyPath.Length -gt 0 -and $allOpenApi.Contains($legacyPath)) {
            $duplicates.Add("$path <-> $legacyPath")
        }
    }

    if ($duplicates.Count -gt 0) {
        return @{ Status = 'DRIFT'; Detail = "duplicate path families: $($duplicates -join '; ')" }
    }
    return @{ Status = 'OK'; Detail = "families checked: $($Families -join ', ')" }
}

function Test-OpenApiPathsHaveRuntimeOwners {
    param(
        [string]$OpenApiFile,
        [string[]]$RuntimeFiles,
        [string[]]$Families,
        [string[]]$OwnedOpenApiPaths,
        [bool]$UsesGeneratedBoundary
    )
    $openApiPaths = Get-OwnedOpenApiPaths -ConfiguredOwnedPaths $OwnedOpenApiPaths -OpenApiFile $OpenApiFile -Families $Families
    if ($UsesGeneratedBoundary) {
        return @{ Status = 'OK'; Detail = "generated boundary owns configured OpenAPI paths ($($openApiPaths.Count))" }
    }
    $runtimePaths = Select-ModuleFamilyPaths -AllPaths (Get-RuntimePaths -RelativePaths $RuntimeFiles) -Families $Families
    $runtimeSet = New-Object System.Collections.Generic.HashSet[string]
    foreach ($path in $runtimePaths) { [void]$runtimeSet.Add($path) }

    $missingOwners = New-Object System.Collections.Generic.List[string]
    foreach ($path in $openApiPaths) {
        if (-not $runtimeSet.Contains($path)) { $missingOwners.Add($path) }
    }

    if ($missingOwners.Count -gt 0) {
        return @{ Status = 'DRIFT'; Detail = "OpenAPI paths without runtime owners: $($missingOwners -join ', ')" }
    }
    return @{ Status = 'OK'; Detail = "OpenAPI paths covered in runtime owners ($($openApiPaths.Count))" }
}

function Test-RuntimeRoutesHaveOpenApiPaths {
    param(
        [string]$OpenApiFile,
        [string[]]$RuntimeFiles,
        [string[]]$Families,
        [string[]]$OwnedOpenApiPaths,
        [bool]$UsesGeneratedBoundary
    )
    $openApiPaths = Get-OwnedOpenApiPaths -ConfiguredOwnedPaths $OwnedOpenApiPaths -OpenApiFile $OpenApiFile -Families $Families
    $openApiSet = New-Object System.Collections.Generic.HashSet[string]
    foreach ($path in $openApiPaths) { [void]$openApiSet.Add($path) }

    if ($UsesGeneratedBoundary) {
        return @{ Status = 'OK'; Detail = "generated boundary runtime routes mapped via OpenAPI contract ($($openApiPaths.Count))" }
    }

    $runtimePaths = Select-ModuleFamilyPaths -AllPaths (Get-RuntimePaths -RelativePaths $RuntimeFiles) -Families $Families
    $missingSpec = New-Object System.Collections.Generic.List[string]
    foreach ($path in $runtimePaths) {
        if (-not $openApiSet.Contains($path)) { $missingSpec.Add($path) }
    }

    if ($missingSpec.Count -gt 0) {
        return @{ Status = 'DRIFT'; Detail = "runtime routes without OpenAPI paths: $($missingSpec -join ', ')" }
    }
    return @{ Status = 'OK'; Detail = "runtime routes mapped in OpenAPI ($($runtimePaths.Count))" }
}

function Test-FrontendGeneratedTypeUsage {
    param(
        [string]$RelativePath,
        [string[]]$Allowlist = @()
    )
    $fullPath = Join-Path $root $RelativePath
    if (-not (Test-Path $fullPath)) {
        return @{ Status = 'MISSING'; Detail = $RelativePath }
    }

    $content = Get-Content $fullPath -Raw
    $customInterfaces = [regex]::Matches($content, '(?m)^\s*(export\s+)?interface\s+([A-Za-z_][A-Za-z0-9_]*)\b')
    # Allow-list for generated-derived type aliases (F1.3 reconciliation):
    #   - `paths[...]` / `components[...]` / `operations[...]` — direct generated-index access.
    #   - `NonNullable<...>` wrapping any of the above (e.g. NonNullable<operations['x']['parameters']['query']>).
    #   - `<GeneratedAlias>['prop']` — an indexed-access off another local alias (itself expected
    #     to be generated-derived, e.g. `VersionDTO['status']` where VersionDTO = components[...]).
    #     The regex can't chase the transitive definition, so it allow-lists the *shape*
    #     (bare identifier immediately followed by an indexed-access); this still requires a
    #     literal `['...']` / `[...]` on the RHS, so it does not blanket-exempt arbitrary aliases.
    $genericAllow = '(?:NonNullable<\s*)?(?:paths|components|operations)\[|[A-Za-z_][A-Za-z0-9_]*\[[''"]'
    $customTypeAliases = [regex]::Matches($content, "(?m)^\s*(export\s+)?type\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(?!.*(?:$genericAllow)).+`$")
    $driftItems = New-Object System.Collections.Generic.List[string]
    $allowSet = New-Object System.Collections.Generic.HashSet[string]
    foreach ($name in $Allowlist) { [void]$allowSet.Add($name) }

    foreach ($m in $customInterfaces) {
        $name = $m.Groups[2].Value
        if ($allowSet.Contains($name)) { continue }
        $driftItems.Add("interface $name")
    }
    foreach ($m in $customTypeAliases) {
        $name = $m.Groups[2].Value
        if ($allowSet.Contains($name)) { continue }
        $driftItems.Add("type $name")
    }

    if ($driftItems.Count -gt 0) {
        return @{ Status = 'DRIFT'; Detail = "$RelativePath handwritten frontend types: $($driftItems -join ', ')" }
    }
    return @{ Status = 'OK'; Detail = "$RelativePath generated type usage looks aligned" }
}

function Test-WikiStatusContradiction {
    param(
        [string]$WikiFile,
        [string]$RuntimeFile,
        [string[]]$OwnedOpenApiPaths = @()
    )
    $wikiPath = Join-Path $root $WikiFile
    $runtimePath = Join-Path $root $RuntimeFile
    if (-not (Test-Path $wikiPath)) {
        return @{ Status = 'MISSING'; Detail = $WikiFile }
    }
    if (-not (Test-Path $runtimePath)) {
        return @{ Status = 'MISSING'; Detail = $RuntimeFile }
    }

    $wikiRaw = Get-Content $wikiPath -Raw
    $wiki = $wikiRaw.ToLowerInvariant()
    $runtime = Get-Content $runtimePath -Raw
    $runtimeGeneratedMounted = (
        $runtime.IndexOf('HandlerWithOptions', [System.StringComparison]::Ordinal) -ge 0 -or
        $runtime.IndexOf('ServerInterfaceWrapper', [System.StringComparison]::Ordinal) -ge 0
    )

    $wikiSaysRawOrWrapperOnly = (
        $wiki.IndexOf('module contract status: raw', [System.StringComparison]::Ordinal) -ge 0 -or
        $wiki.IndexOf('module contract status: wrapper-only', [System.StringComparison]::Ordinal) -ge 0
    )
    $wikiSaysGeneratedBoundary = (
        $wiki.IndexOf('generated boundary mounted', [System.StringComparison]::Ordinal) -ge 0 -or
        $wiki.IndexOf('generated-wrapper mount', [System.StringComparison]::Ordinal) -ge 0
    )
    $wikiSaysContracted = (
        $wiki.IndexOf('module contract status: contracted', [System.StringComparison]::Ordinal) -ge 0
    )

    if ($wikiSaysRawOrWrapperOnly -and $runtimeGeneratedMounted) {
        return @{ Status = 'DRIFT'; Detail = "$WikiFile says raw/wrapper-only while runtime mounts generated boundary" }
    }
    if ($wikiSaysGeneratedBoundary -and -not $runtimeGeneratedMounted) {
        return @{ Status = 'DRIFT'; Detail = "$WikiFile says generated boundary mounted but runtime mount does not indicate wrapper boundary" }
    }

    # Detect stale bootstrap-only / not-mounted phrasing in active wiki when runtime
    # is actually wrapper-mounted. These claims were valid pre-freeze; post-freeze
    # they are false and the gate must reject them.
    if ($runtimeGeneratedMounted) {
        $stalePhrases = @(
            'bootstrap only',
            'still mounted via mux.handlefunc',
            'still mounted via stdlib mux',
            'still mounted via `mux.handlefunc',
            'not mounted)'
        )
        foreach ($phrase in $stalePhrases) {
            if ($wiki.IndexOf($phrase, [System.StringComparison]::Ordinal) -ge 0) {
                return @{ Status = 'DRIFT'; Detail = "$WikiFile contains stale phrase '$phrase' while runtime mounts generated boundary" }
            }
        }
    }

    # When module status is Contracted or Generated boundary mounted, a route-truth
    # row marked 'Spec missing' for an OwnedOpenApiPaths entry is a contradiction.
    if (($wikiSaysContracted -or $wikiSaysGeneratedBoundary) -and $null -ne $OwnedOpenApiPaths -and $OwnedOpenApiPaths.Count -gt 0) {
        $lines = $wikiRaw -split "`n"
        foreach ($line in $lines) {
            if ($line.IndexOf('Spec missing', [System.StringComparison]::Ordinal) -lt 0) { continue }
            foreach ($ownedPath in $OwnedOpenApiPaths) {
                $needle = '`' + $ownedPath + '`'
                if ($line.IndexOf($needle, [System.StringComparison]::Ordinal) -ge 0) {
                    return @{ Status = 'DRIFT'; Detail = "$WikiFile marks owned path $ownedPath as 'Spec missing' while module is contracted/generated-mounted" }
                }
            }
        }
    }

    return @{ Status = 'OK'; Detail = "$WikiFile status aligned with runtime mount pattern" }
}

function Test-Surface {
    param(
        [string]$Label,
        [string]$RelativePath,
        [string[]]$Patterns
    )

    $fullPath = Join-Path $root $RelativePath
    if (-not (Test-Path $fullPath)) {
        return @{
            Status = 'MISSING'
            Detail = $RelativePath
        }
    }

    $content = Get-Content $fullPath -Raw
    $missingPatterns = New-Object System.Collections.Generic.List[string]

    foreach ($pattern in $Patterns) {
        if ($content.IndexOf($pattern, [System.StringComparison]::Ordinal) -lt 0) {
            $missingPatterns.Add($pattern)
        }
    }

    if ($missingPatterns.Count -gt 0) {
        return @{
            Status = 'DRIFT'
            Detail = "$RelativePath missing: $($missingPatterns -join ', ')"
        }
    }

    return @{
        Status = 'OK'
        Detail = $RelativePath
    }
}

function Test-ForbiddenPatterns {
    param(
        [string]$Label,
        [string]$RelativePath,
        [string[]]$Patterns
    )

    if (-not $Patterns -or $Patterns.Count -eq 0) {
        return @{
            Status = 'OK'
            Detail = $RelativePath
        }
    }

    $fullPath = Join-Path $root $RelativePath
    if (-not (Test-Path $fullPath)) {
        return @{
            Status = 'MISSING'
            Detail = $RelativePath
        }
    }

    $content = Get-Content $fullPath -Raw
    $foundPatterns = New-Object System.Collections.Generic.List[string]

    foreach ($pattern in $Patterns) {
        if ($content.IndexOf($pattern, [System.StringComparison]::Ordinal) -ge 0) {
            $foundPatterns.Add($pattern)
        }
    }

    if ($foundPatterns.Count -gt 0) {
        return @{
            Status = 'DRIFT'
            Detail = "$RelativePath forbidden: $($foundPatterns -join ', ')"
        }
    }

    return @{
        Status = 'OK'
        Detail = $RelativePath
    }
}

$config = New-ModuleConfig -ModuleName $Module

$checks = @(
    @{
        Label = 'runtime route ownership files'
        File = $config.RuntimeFile
        Patterns = $config.RuntimePatterns
    },
    @{
        Label = 'OpenAPI path presence'
        File = $config.OpenApiFile
        Patterns = $config.OpenApiPatterns
    },
    @{
        Label = 'generated backend package presence'
        File = $config.BackendFile
        Patterns = $config.BackendPatterns
    },
    @{
        Label = 'generated frontend path/type presence'
        File = $config.FrontendTypesFile
        Patterns = $config.FrontendTypesPatterns
    },
    @{
        Label = 'feature API wrapper presence'
        File = $config.FrontendWrapperFile
        Patterns = $config.FrontendWrapperPatterns
    }
)

$hasIssues = $false

Write-Host "Module: $Module"

foreach ($check in $checks) {
    $result = Test-Surface -Label $check.Label -RelativePath $check.File -Patterns $check.Patterns
    if ($result.Status -ne 'OK') {
        $hasIssues = $true
    }

    Write-Host ("[{0}] {1} - {2}" -f $result.Status, $check.Label, $result.Detail)
}

$runtimeForbiddenResult = Test-ForbiddenPatterns -Label 'runtime forbidden patterns' -RelativePath $config.RuntimeFile -Patterns $config.RuntimeForbiddenPatterns
if ($runtimeForbiddenResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $runtimeForbiddenResult.Status, 'runtime forbidden patterns', $runtimeForbiddenResult.Detail)

$openApiForbiddenResult = Test-ForbiddenPatterns -Label 'OpenAPI forbidden patterns' -RelativePath $config.OpenApiFile -Patterns $config.OpenApiForbiddenPatterns
if ($openApiForbiddenResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $openApiForbiddenResult.Status, 'OpenAPI forbidden patterns', $openApiForbiddenResult.Detail)

$duplicateFamiliesResult = Test-DuplicateCanonicalFamilies -OpenApiFile $config.OpenApiFile -Families $config.RouteFamilies
if ($duplicateFamiliesResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $duplicateFamiliesResult.Status, 'freeze: duplicate canonical OpenAPI path families', $duplicateFamiliesResult.Detail)

$openApiOwnerResult = Test-OpenApiPathsHaveRuntimeOwners `
    -OpenApiFile $config.OpenApiFile `
    -RuntimeFiles $config.RuntimeOwnerFiles `
    -Families $config.RouteFamilies `
    -OwnedOpenApiPaths $config.OwnedOpenApiPaths `
    -UsesGeneratedBoundary $config.UsesGeneratedBoundary
if ($openApiOwnerResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $openApiOwnerResult.Status, 'freeze: OpenAPI paths without runtime owners', $openApiOwnerResult.Detail)

$runtimeOwnerResult = Test-RuntimeRoutesHaveOpenApiPaths `
    -OpenApiFile $config.OpenApiFile `
    -RuntimeFiles $config.RuntimeOwnerFiles `
    -Families $config.RouteFamilies `
    -OwnedOpenApiPaths $config.OwnedOpenApiPaths `
    -UsesGeneratedBoundary $config.UsesGeneratedBoundary
if ($runtimeOwnerResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $runtimeOwnerResult.Status, 'freeze: runtime routes without OpenAPI paths', $runtimeOwnerResult.Detail)

$frontendTypeDriftResult = Test-FrontendGeneratedTypeUsage -RelativePath $config.FrontendWrapperFile -Allowlist $config.FrontendTypeAllowlist
if ($frontendTypeDriftResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $frontendTypeDriftResult.Status, 'freeze: frontend handwritten type drift', $frontendTypeDriftResult.Detail)

$wikiStatusResult = Test-WikiStatusContradiction `
    -WikiFile $config.WikiFile `
    -RuntimeFile $config.RuntimeFile `
    -OwnedOpenApiPaths $config.OwnedOpenApiPaths
if ($wikiStatusResult.Status -ne 'OK') {
    $hasIssues = $true
}
Write-Host ("[{0}] {1} - {2}" -f $wikiStatusResult.Status, 'freeze: wiki status contradiction', $wikiStatusResult.Detail)

if ($hasIssues) {
    Write-Host 'RESULT: shared contract prerequisite'
    exit 1
}

Write-Host 'RESULT: contract surfaces present; manual drift review still required'

