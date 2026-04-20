param(
    [int]$Port = 3000
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
Set-Location -LiteralPath (Join-Path $repoRoot "frontend")

$env:PATH = "C:\Program Files\nodejs;C:\Program Files\nodejs\node_modules\corepack\shims;$env:PATH"

Write-Host "Starting SteelLink admin frontend..." -ForegroundColor Cyan
Write-Host "Frontend: http://localhost:$Port" -ForegroundColor DarkGray
Write-Host "Proxy target: http://localhost:8080" -ForegroundColor DarkGray

cmd /c "pnpm dev -- --host 0.0.0.0 --strictPort --port $Port"
