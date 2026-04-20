param(
    [int]$Port = 3000
)

$ErrorActionPreference = "Stop"

function Invoke-Pnpm {
    param(
        [string[]]$Arguments
    )

    $pnpmCommand = Get-Command pnpm -ErrorAction SilentlyContinue
    if ($pnpmCommand) {
        & $pnpmCommand.Source @Arguments
        return
    }

    $corepackCommand = Get-Command corepack -ErrorAction SilentlyContinue
    if ($corepackCommand) {
        & $corepackCommand.Source pnpm @Arguments
        return
    }

    throw "pnpm not found. Enable it with Corepack (`corepack enable pnpm`) or install pnpm globally."
}

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
Set-Location -LiteralPath (Join-Path $repoRoot "frontend")

Write-Host "Starting SteelLink admin frontend..." -ForegroundColor Cyan
Write-Host "Frontend: http://localhost:$Port" -ForegroundColor DarkGray
Write-Host "Proxy target: http://localhost:8080" -ForegroundColor DarkGray

Invoke-Pnpm -Arguments @("dev", "--", "--host", "0.0.0.0", "--strictPort", "--port", "$Port")
