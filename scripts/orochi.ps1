# Orochi BitTorrent Client Launcher for Windows PowerShell
# This script ensures the application runs properly on Windows

$ErrorActionPreference = "Stop"

# Get the script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$orochiExe = Join-Path (Split-Path -Parent $scriptDir) "orochi.exe"

# Check if orochi.exe exists
if (!(Test-Path $orochiExe)) {
    Write-Host "Error: orochi.exe not found at $orochiExe" -ForegroundColor Red
    Write-Host "Please ensure the binary is in the correct location." -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Display startup message
Write-Host "Starting Orochi BitTorrent Client..." -ForegroundColor Green
Write-Host ""

# Run Orochi with any passed arguments
try {
    & $orochiExe $args
    $exitCode = $LASTEXITCODE
} catch {
    Write-Host "Error running Orochi: $_" -ForegroundColor Red
    $exitCode = 1
}

# Keep console open if there was an error
if ($exitCode -ne 0) {
    Write-Host ""
    Write-Host "Orochi exited with error code: $exitCode" -ForegroundColor Red
    Read-Host "Press Enter to exit"
}