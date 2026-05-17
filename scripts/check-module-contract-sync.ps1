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
                RuntimeFile = 'internal/modules/templates/delivery/http/handler.go'
                RuntimePatterns = @('/api/v1/templates', 'generated.ListTemplates', 'generated.GetTemplate')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/api/v1/templates:', '/api/v1/templates/{id}/versions/{n}/submit:', '/api/v1/templates/placeholder-catalog:')
                BackendFile = 'internal/modules/templates/api/api.gen.go'
                BackendPatterns = @('ListTemplates', 'CreateTemplate', 'GetTemplate', 'SubmitTemplateVersion')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/api/v1/templates":', '"/api/v1/templates/{id}/versions/{n}/submit":', '"/api/v1/templates/placeholder-catalog":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/templates/api/templates.ts'
                FrontendWrapperPatterns = @('/api/v1/templates', 'submitForReview', 'approveVersion', 'putTemplateSchemas')
            }
        }
        'approval' {
            return @{
                RuntimeFile = 'internal/modules/documents/approval/http/router.go'
                RuntimePatterns = @('/api/v1/approval/inbox', '/api/v1/documents/{id}/signoff', '/api/v1/approval/routes')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/api/v1/approval/inbox:', '/api/v1/documents/{id}/signoff:', '/api/v1/documents/{id}/cancel:', '/api/v1/approval/routes:')
                BackendFile = 'internal/modules/documents/approval/api/api.gen.go'
                BackendPatterns = @('ListApprovalInbox', 'ListApprovalRoutes', 'CreateApprovalRoute')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/api/v1/approval/inbox":', '"/api/v1/documents/{id}/signoff":', '"/api/v1/documents/{id}/cancel":', '"/api/v1/approval/routes":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/approval/api/approvalApi.ts'
                FrontendWrapperPatterns = @('/approval/inbox', '/documents/${documentId}/signoff', '/documents/${documentId}/cancel', '/approval/routes', 'listInbox', 'signoff', 'cancel', 'createRoute', 'listRoutes')
            }
        }
        'documents' {
            return @{
                RuntimeFile = 'internal/modules/documents/delivery/http/handler.go'
                RuntimePatterns = @('/api/v1/documents', 'h.listDocuments', 'h.getDocument')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/api/v1/documents:', '/api/v1/documents/{id}:', '/api/v1/documents/{id}/finalize:')
                BackendFile = 'internal/modules/documents/api/api.gen.go'
                BackendPatterns = @('ListDocuments', 'GetDocument', 'FinalizeDocument')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/api/v1/documents":', '"/api/v1/documents/{id}":', '"/api/v1/documents/{id}/finalize":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/documents/api/documents.ts'
                FrontendWrapperPatterns = @('/api/v1/documents', 'listDocuments', 'getDocument', 'finalizeDocument')
            }
        }
        'registry' {
            return @{
                RuntimeFile = 'internal/modules/registry/delivery/http/handler.go'
                RuntimePatterns = @('/api/v1/controlled-documents', 'generated.ListControlledDocuments', 'generated.GetControlledDocument')
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @('/api/v1/controlled-documents:', '/api/v1/controlled-documents/{id}:', '/api/v1/controlled-documents/{id}/revisions:')
                BackendFile = 'internal/modules/registry/api/api.gen.go'
                BackendPatterns = @('ListControlledDocuments', 'AtomicCreateControlledDocument', 'GetControlledDocument')
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @('"/api/v1/controlled-documents":', '"/api/v1/controlled-documents/{id}":', '"/api/v1/controlled-documents/{id}/revisions":')
                FrontendWrapperFile = 'frontend/apps/web/src/features/registry/api/controlledDocuments.ts'
                FrontendWrapperPatterns = @('/api/v1/controlled-documents', 'fetchControlledDocuments', 'createControlledDocumentAtomic')
            }
        }
        default {
            $basePath = "/api/v1/$ModuleName"
            return @{
                RuntimeFile = "internal/modules/$ModuleName/delivery/http/handler.go"
                RuntimePatterns = @($basePath)
                OpenApiFile = 'api/openapi/v1/openapi.yaml'
                OpenApiPatterns = @("${basePath}:")
                BackendFile = "internal/modules/$ModuleName/api/api.gen.go"
                BackendPatterns = @($basePath)
                FrontendTypesFile = 'frontend/apps/web/src/lib/api-types/index.d.ts'
                FrontendTypesPatterns = @("""$basePath"":")
                FrontendWrapperFile = "frontend/apps/web/src/features/$ModuleName/api/${ModuleName}V2.ts"
                FrontendWrapperPatterns = @($basePath)
            }
        }
    }
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

if ($hasIssues) {
    Write-Host 'RESULT: shared contract prerequisite'
    exit 1
}

Write-Host 'RESULT: contract surfaces present; manual drift review still required'
