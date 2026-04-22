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
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($goCommand) {
        return $goCommand.Source
    }

    $candidates = @()

    $localProgramRoot = Join-Path $env:LOCALAPPDATA "Programs"
    if (Test-Path $localProgramRoot) {
        $candidates += Get-ChildItem -Path $localProgramRoot -Directory -Filter "Go*" -ErrorAction SilentlyContinue |
            ForEach-Object { Join-Path $_.FullName "go\bin\go.exe" }
    }

    $candidates += @(
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Go executable not found. Install Go and ensure `go` is on PATH, or install it into a standard Windows location."
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
