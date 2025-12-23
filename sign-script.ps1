# PowerShell Script Signer Helper
# Uses the codesign API to sign PowerShell scripts

param(
    [Parameter(Mandatory=$true)]
    [string]$ScriptPath,
    
    [string]$CertThumbprint,
    [string]$PfxPath,
    [string]$PfxPassword,
    [string]$TimestampServer,
    [string]$ApiUrl = "http://localhost:20000"
)

# Check if script exists
if (-not (Test-Path $ScriptPath)) {
    Write-Host "ERROR: Script not found: $ScriptPath" -ForegroundColor Red
    exit 1
}

# Resolve full path
$ScriptPath = (Resolve-Path $ScriptPath).Path

Write-Host "Signing script: $ScriptPath" -ForegroundColor Cyan
Write-Host ""

# Check if API server is running
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/check-powershell" -Method Get -TimeoutSec 2
    if (-not $response.available) {
        Write-Host "ERROR: PowerShell is not available on the server" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "ERROR: Cannot connect to codesign API at $ApiUrl" -ForegroundColor Red
    Write-Host "Please start the server first:" -ForegroundColor Yellow
    Write-Host "  cd `"C:\Users\Keith Ransom\codesign`"" -ForegroundColor Yellow
    Write-Host "  .\start-powershell-signer.ps1" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Or use the batch file:" -ForegroundColor Yellow
    Write-Host "  `"C:\Users\Keith Ransom\codesign\start-powershell-signer.bat`"" -ForegroundColor Yellow
    exit 1
}

# If no certificate specified, list available certificates
if (-not $CertThumbprint -and -not $PfxPath) {
    Write-Host "No certificate specified. Listing available certificates..." -ForegroundColor Yellow
    Write-Host ""
    
    try {
        $certs = Invoke-RestMethod -Uri "$ApiUrl/api/list-certificates" -Method Get
        if ($certs.certificates.Count -eq 0) {
            Write-Host "No code signing certificates found in Cert:\CurrentUser\My" -ForegroundColor Red
            Write-Host ""
            Write-Host "You can create a self-signed certificate using the web UI at:" -ForegroundColor Yellow
            Write-Host "  $ApiUrl" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "Or specify a PFX file with -PfxPath parameter" -ForegroundColor Yellow
            exit 1
        }
        
        Write-Host "Available certificates:" -ForegroundColor Green
        for ($i = 0; $i -lt $certs.certificates.Count; $i++) {
            $cert = $certs.certificates[$i]
            Write-Host "  [$i] $($cert.Subject)" -ForegroundColor White
            Write-Host "      Thumbprint: $($cert.Thumbprint)" -ForegroundColor Gray
            Write-Host "      Expires: $($cert.NotAfter)" -ForegroundColor Gray
            Write-Host ""
        }
        
        Write-Host "Please specify a certificate using -CertThumbprint parameter" -ForegroundColor Yellow
        Write-Host "Example: .\sign-script.ps1 -ScriptPath `"$ScriptPath`" -CertThumbprint `"$($certs.certificates[0].Thumbprint)`"" -ForegroundColor Cyan
        exit 0
    } catch {
        Write-Host "ERROR: Failed to list certificates: $_" -ForegroundColor Red
        exit 1
    }
}

# Prepare request body
$body = @{
    scriptPath = $ScriptPath
}

if ($CertThumbprint) {
    $body.certThumbprint = $CertThumbprint
}

if ($PfxPath) {
    if (-not (Test-Path $PfxPath)) {
        Write-Host "ERROR: PFX file not found: $PfxPath" -ForegroundColor Red
        exit 1
    }
    $body.pfxPath = (Resolve-Path $PfxPath).Path
    if ($PfxPassword) {
        $body.pfxPassword = $PfxPassword
    }
}

if ($TimestampServer) {
    $body.timestampServer = $TimestampServer
}

# Sign the script
Write-Host "Signing script..." -ForegroundColor Yellow
try {
    $result = Invoke-RestMethod -Uri "$ApiUrl/api/sign-script" -Method Post -Body ($body | ConvertTo-Json) -ContentType "application/json"
    
    if ($result.success) {
        Write-Host "✓ Script signed successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Signature details:" -ForegroundColor Cyan
        Write-Host "  Status: $($result.data.Status)" -ForegroundColor White
        Write-Host "  Signed by: $($result.data.SignedBy)" -ForegroundColor White
        if ($result.data.TimeStamper) {
            Write-Host "  Timestamped by: $($result.data.TimeStamper)" -ForegroundColor White
        }
        Write-Host "  Signature Type: $($result.data.SignatureType)" -ForegroundColor White
        Write-Host ""
        
        # Verify the signature
        Write-Host "Verifying signature..." -ForegroundColor Yellow
        $verifyBody = @{ scriptPath = $ScriptPath } | ConvertTo-Json
        $verifyResult = Invoke-RestMethod -Uri "$ApiUrl/api/verify-signature" -Method Post -Body $verifyBody -ContentType "application/json"
        
        if ($verifyResult.success) {
            Write-Host "✓ Signature verified!" -ForegroundColor Green
            Write-Host "  Status: $($verifyResult.data.Status)" -ForegroundColor White
        } else {
            Write-Host "⚠ Verification warning: $($verifyResult.error)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "ERROR: Failed to sign script: $($result.error)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "ERROR: Failed to sign script: $_" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        try {
            $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Host "  Details: $($errorDetails.error)" -ForegroundColor Red
        } catch {
            Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
    }
    exit 1
}

