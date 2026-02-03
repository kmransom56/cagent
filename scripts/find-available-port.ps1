# Find Available Port in 11000-12000 Range
# Usage: .\scripts\find-available-port.ps1 [startPort] [endPort]

param(
    [int]$StartPort = 11000,
    [int]$EndPort = 12000
)

Write-Host "Finding available port in range $StartPort-$EndPort..." -ForegroundColor Cyan

$port = $StartPort
$foundPort = $null

while ($port -le $EndPort -and -not $foundPort) {
    $inUse = Test-NetConnection -ComputerName localhost -Port $port -InformationLevel Quiet -WarningAction SilentlyContinue
    if (-not $inUse) {
        $foundPort = $port
        Write-Host "✓ Found available port: $foundPort" -ForegroundColor Green
    } else {
        $port++
    }
}

if ($foundPort) {
    Write-Output $foundPort
    exit 0
} else {
    Write-Host "✗ No available ports in range $StartPort-$EndPort" -ForegroundColor Red
    exit 1
}

