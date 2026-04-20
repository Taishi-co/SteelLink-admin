param(
    [int]$Port = 8080,
    [string]$GinMode = "debug",
    [string]$JwtSecret = "steellink-admin-dev-secret",
    [string]$AdminUsername = "admin",
    [string]$AdminPassword = "admin123",
    [string]$AdminRealName = "Platform Admin"
)

$ErrorActionPreference = "Stop"

function Resolve-GoExe {
    $candidates = @(
        (Join-Path $env:LOCALAPPDATA "Programs\Go-1.26.2\go\bin\go.exe"),
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Go executable not found. Install Go locally or place it under $env:LOCALAPPDATA\Programs\Go-1.26.2."
}

$goExe = Resolve-GoExe
$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Set-Location -LiteralPath (Join-Path $repoRoot "backend")

$env:PORT = "$Port"
$env:GIN_MODE = $GinMode
$env:JWT_SECRET = $JwtSecret
$env:ADMIN_DEFAULT_USERNAME = $AdminUsername
$env:ADMIN_DEFAULT_PASSWORD = $AdminPassword
$env:ADMIN_DEFAULT_REAL_NAME = $AdminRealName

Write-Host "Starting SteelLink admin backend..." -ForegroundColor Cyan
Write-Host "Backend: http://localhost:$Port" -ForegroundColor DarkGray
Write-Host "Default admin: $AdminUsername / $AdminPassword" -ForegroundColor DarkGray

& $goExe run .\cmd\server\main.go
