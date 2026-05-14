param(
    [string]$TargetRoute = '/api/v1/health/ready',
    [switch]$StartApi
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Add-Type -AssemblyName System.Net.Http

$baseUrl = 'http://127.0.0.1:8081'
$loginRoute = '/api/v1/auth/login'
$meRoute = '/api/v1/auth/me'
$loginIdentifier = 'admin'
$loginPassword = 'AdminMetalDocs123!'

function Fail-Checkpoint {
    param(
        [string]$Name,
        [string]$Message
    )

    Write-Host "FAIL $Name - $Message"
    throw "checkpoint failure"
}

function Pass-Checkpoint {
    param(
        [string]$Name,
        [string]$Message
    )

    Write-Host "PASS $Name - $Message"
}

function Invoke-CheckedRequest {
    param(
        [System.Net.Http.HttpClient]$Client,
        [string]$Method,
        [string]$Route,
        [string]$JsonBody
    )

    $uri = $baseUrl + $Route
    $httpMethod = New-Object System.Net.Http.HttpMethod($Method.ToUpperInvariant())
    $request = New-Object System.Net.Http.HttpRequestMessage($httpMethod, $uri)

    if ($JsonBody) {
        $request.Content = New-Object System.Net.Http.StringContent(
            $JsonBody,
            [System.Text.Encoding]::UTF8,
            'application/json'
        )
    }

    try {
        return $Client.SendAsync($request).GetAwaiter().GetResult()
    } catch {
        throw "network error calling $Method $Route - $($_.Exception.Message)"
    }
}

function Wait-ForReady {
    param(
        [System.Net.Http.HttpClient]$Client,
        [int]$TimeoutSeconds = 180,
        [System.Diagnostics.Process]$StartupProcess,
        [int]$ConsecutiveSuccesses = 2
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    $lastStatusCode = $null
    $lastBody = $null
    $successCount = 0
    do {
        try {
            $response = Invoke-CheckedRequest -Client $Client -Method Get -Route '/api/v1/health/ready'
            if ($response.IsSuccessStatusCode) {
                $successCount += 1
                if ($successCount -ge $ConsecutiveSuccesses) {
                    return $response
                }

                Start-Sleep -Seconds 1
                continue
            }

            $lastStatusCode = [int]$response.StatusCode
            $lastBody = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
            $successCount = 0
        } catch {
            $successCount = 0
        }

        if ($StartupProcess -and $StartupProcess.HasExited) {
            if ($lastStatusCode -ne $null) {
                throw "start-api.ps1 exited before readiness succeeded; last response HTTP $lastStatusCode; body=$lastBody"
            }
            throw "start-api.ps1 exited before readiness succeeded"
        }

        Start-Sleep -Seconds 1
    } while ((Get-Date) -lt $deadline)

    if ($lastStatusCode -ne $null) {
        throw "timed out waiting for /api/v1/health/ready success; last response HTTP $lastStatusCode; body=$lastBody"
    }

    throw "timed out waiting for /api/v1/health/ready"
}

$cookieContainer = New-Object System.Net.CookieContainer
$handler = New-Object System.Net.Http.HttpClientHandler
$handler.CookieContainer = $cookieContainer
$handler.UseCookies = $true
$client = New-Object System.Net.Http.HttpClient($handler)
$client.Timeout = [TimeSpan]::FromSeconds(15)

try {
    $startupProcess = $null
    if ($StartApi) {
        Write-Host "Starting API via scripts/start-api.ps1 -Build -NoWorker"
        $startupProcess = Start-Process `
            -FilePath "powershell" `
            -ArgumentList @(
                '-NoProfile',
                '-ExecutionPolicy',
                'Bypass',
                '-File',
                (Join-Path $PSScriptRoot 'start-api.ps1'),
                '-Build',
                '-NoWorker'
            ) `
            -PassThru `
            -WindowStyle Hidden

        $readyResponse = Wait-ForReady -Client $client -StartupProcess $startupProcess
        Pass-Checkpoint -Name 'startup' -Message "/api/v1/health/ready returned HTTP $([int]$readyResponse.StatusCode) after start-api.ps1"
    }

    $loginPayload = @{
        identifier = $loginIdentifier
        password = $loginPassword
    } | ConvertTo-Json -Compress

    $loginResponse = Invoke-CheckedRequest -Client $client -Method Post -Route $loginRoute -JsonBody $loginPayload
    if (-not $loginResponse.IsSuccessStatusCode) {
        $loginBody = $loginResponse.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        Fail-Checkpoint -Name 'login-endpoint' -Message "POST $loginRoute returned HTTP $([int]$loginResponse.StatusCode); body=$loginBody"
    }
    Pass-Checkpoint -Name 'login-endpoint' -Message "POST $loginRoute returned HTTP $([int]$loginResponse.StatusCode)"

    $cookieScope = [Uri]$baseUrl
    $cookies = $cookieContainer.GetCookies($cookieScope)
    if ($cookies.Count -lt 1) {
        Fail-Checkpoint -Name 'login-session' -Message 'login succeeded but no session cookie was stored'
    }
    Pass-Checkpoint -Name 'login-session' -Message "received $($cookies.Count) session cookie(s)"

    $meResponse = Invoke-CheckedRequest -Client $client -Method Get -Route $meRoute
    if (-not $meResponse.IsSuccessStatusCode) {
        $meBody = $meResponse.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        Fail-Checkpoint -Name 'auth-me' -Message "GET $meRoute returned HTTP $([int]$meResponse.StatusCode); body=$meBody"
    }
    Pass-Checkpoint -Name 'auth-me' -Message "GET $meRoute returned HTTP $([int]$meResponse.StatusCode)"

    $targetResponse = Invoke-CheckedRequest -Client $client -Method Get -Route $TargetRoute
    if (-not $targetResponse.IsSuccessStatusCode) {
        $targetBody = $targetResponse.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        Fail-Checkpoint -Name 'target-route' -Message "GET $TargetRoute returned HTTP $([int]$targetResponse.StatusCode); body=$targetBody"
    }
    Pass-Checkpoint -Name 'target-route' -Message "GET $TargetRoute returned HTTP $([int]$targetResponse.StatusCode)"
}
catch {
    if ($_.Exception.Message -ne 'checkpoint failure') {
        Write-Host "FAIL unexpected - $($_.Exception.Message)"
    }
    exit 1
}
finally {
    $client.Dispose()
    $handler.Dispose()
}
