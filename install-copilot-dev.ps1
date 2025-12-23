# Chat Copilot Development Environment Setup Script
# For Windows (PowerShell 7+)
# Installs: .NET 8 SDK, Node.js, Yarn, Git, Docker Desktop, and VS Code extensions
#
# To run this script, use one of the following methods:
#   1. From PowerShell: pwsh -ExecutionPolicy Bypass -File .\install-copilot-dev.ps1
#   2. From Command Prompt: pwsh -ExecutionPolicy Bypass -File install-copilot-dev.ps1
#   3. Right-click script > Run with PowerShell (may require execution policy change)

param(
    [switch]$SkipChocolatey = $false,
    [switch]$SkipDocker = $false,
    [switch]$SkipVSCode = $false,
    [switch]$VSCodeExtensions = $true
)

# Requires PowerShell 7+ (Core)
if ($PSVersionTable.PSVersion.Major -lt 7) {
    Write-Host "ERROR: This script requires PowerShell 7 or higher (PowerShell Core)." -ForegroundColor Red
    Write-Host "Download from: https://github.com/PowerShell/PowerShell/releases" -ForegroundColor Yellow
    exit 1
}

# Check if running as Administrator
if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "ERROR: This script must be run as Administrator." -ForegroundColor Red
    Write-Host "Please right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    exit 1
}

Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Chat Copilot Development Environment Setup                   ║" -ForegroundColor Cyan
Write-Host "║  Installing prerequisites for Semantic Kernel + Chat Copilot  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Step 1: Install Chocolatey (if not already installed)
Write-Host "[1/6] Checking Chocolatey..." -ForegroundColor Green
if (-not (Get-Command choco -ErrorAction SilentlyContinue)) {
    if ($SkipChocolatey) {
        Write-Host "⊘ Skipping Chocolatey installation" -ForegroundColor Yellow
    } else {
        Write-Host "Installing Chocolatey package manager..." -ForegroundColor Yellow
        Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process -Force
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
        iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        Write-Host "✓ Chocolatey installed successfully" -ForegroundColor Green
    }
} else {
    Write-Host "✓ Chocolatey already installed" -ForegroundColor Green
}

# Step 2: Install .NET 8 SDK
Write-Host ""
Write-Host "[2/6] Checking .NET 8 SDK..." -ForegroundColor Green
$dotnetVersion = dotnet --version 2>$null
if ($null -eq $dotnetVersion) {
    Write-Host "Installing .NET 8 SDK..." -ForegroundColor Yellow
    choco install dotnet-sdk -y --version=8.0
    Write-Host "✓ .NET 8 SDK installed successfully" -ForegroundColor Green
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
    dotnet --version
} else {
    Write-Host "✓ .NET already installed (version $dotnetVersion)" -ForegroundColor Green
}

# Verify required .NET workloads
Write-Host "Installing .NET workloads..." -ForegroundColor Yellow
dotnet workload restore 2>$null
Write-Host "✓ .NET workloads ready" -ForegroundColor Green

# Step 3: Install Node.js
Write-Host ""
Write-Host "[3/6] Checking Node.js..." -ForegroundColor Green
$nodeVersion = node --version 2>$null
if ($null -eq $nodeVersion) {
    Write-Host "Installing Node.js (LTS)..." -ForegroundColor Yellow
    choco install nodejs --lts -y
    Write-Host "✓ Node.js installed successfully" -ForegroundColor Green
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
    node --version
} else {
    Write-Host "✓ Node.js already installed (version $nodeVersion)" -ForegroundColor Green
}

# Step 4: Install Yarn
Write-Host ""
Write-Host "[4/6] Checking Yarn..." -ForegroundColor Green
$yarnVersion = yarn --version 2>$null
if ($null -eq $yarnVersion) {
    Write-Host "Installing Yarn (Classic)..." -ForegroundColor Yellow
    npm install -g yarn
    yarn set version classic
    Write-Host "✓ Yarn installed successfully" -ForegroundColor Green
    yarn --version
} else {
    Write-Host "✓ Yarn already installed (version $yarnVersion)" -ForegroundColor Green
    # Ensure classic version
    yarn set version classic 2>$null
}

# Step 5: Install Git (if not already installed)
Write-Host ""
Write-Host "[5/6] Checking Git..." -ForegroundColor Green
$gitVersion = git --version 2>$null
if ($null -eq $gitVersion) {
    Write-Host "Installing Git..." -ForegroundColor Yellow
    choco install git -y
    Write-Host "✓ Git installed successfully" -ForegroundColor Green
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
    git --version
} else {
    Write-Host "✓ Git already installed" -ForegroundColor Green
}

# Step 6: Install Docker Desktop (Optional)
Write-Host ""
Write-Host "[6/6] Checking Docker..." -ForegroundColor Green
$dockerVersion = docker --version 2>$null
if ($null -eq $dockerVersion) {
    if ($SkipDocker) {
        Write-Host "⊘ Skipping Docker installation (optional)" -ForegroundColor Yellow
    } else {
        Write-Host "Installing Docker Desktop..." -ForegroundColor Yellow
        choco install docker-desktop -y
        Write-Host "✓ Docker Desktop installed successfully" -ForegroundColor Green
        Write-Host "⚠ Docker Desktop requires restart. Please restart your computer and run Docker Desktop." -ForegroundColor Yellow
    }
} else {
    Write-Host "✓ Docker already installed" -ForegroundColor Green
}

# Optional: Install VS Code extensions
if ($VSCodeExtensions) {
    Write-Host ""
    Write-Host "[Optional] Installing VS Code Extensions..." -ForegroundColor Green
    $codeCmd = code 2>$null
    if ($null -ne $codeCmd) {
        Write-Host "Installing recommended extensions..." -ForegroundColor Yellow
        code --install-extension ms-dotnettools.csharp
        code --install-extension ms-dotnettools.vscode-dotnet-runtime
        code --install-extension ms-python.python
        code --install-extension ms-vscode.makefile-tools
        code --install-extension eamodio.gitlens
        code --install-extension ms-vscode-remote.remote-wsl
        Write-Host "✓ VS Code extensions installed" -ForegroundColor Green
    } else {
        Write-Host "⊘ VS Code not found or not in PATH. Skipping extensions." -ForegroundColor Yellow
    }
}

# Final verification
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Installation Summary                                          ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Write-Host ""
Write-Host "Installed versions:" -ForegroundColor Green
Write-Host "  .NET SDK:    $(dotnet --version)" -ForegroundColor White
Write-Host "  Node.js:     $(node --version)" -ForegroundColor White
Write-Host "  Yarn:        $(yarn --version)" -ForegroundColor White
Write-Host "  Git:         $(git --version)" -ForegroundColor White
if ($null -ne $dockerVersion) {
    Write-Host "  Docker:      $(docker --version)" -ForegroundColor White
}

Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Green
Write-Host "  1. Clone the Chat Copilot repository:" -ForegroundColor White
Write-Host "     git clone https://github.com/microsoft/chat-copilot" -ForegroundColor Yellow
Write-Host ""
Write-Host "  2. Navigate to the chat-copilot directory:" -ForegroundColor White
Write-Host "     cd chat-copilot" -ForegroundColor Yellow
Write-Host ""
Write-Host "  3. Run the configuration script:" -ForegroundColor White
Write-Host "     .\scripts\Configure.ps1 -AIService OpenAI -APIKey YOUR_API_KEY" -ForegroundColor Yellow
Write-Host "     (or use AzureOpenAI with -Endpoint parameter)" -ForegroundColor Yellow
Write-Host ""
Write-Host "  4. Start Chat Copilot:" -ForegroundColor White
Write-Host "     .\scripts\Start.ps1" -ForegroundColor Yellow
Write-Host ""

Write-Host "Documentation:" -ForegroundColor Green
Write-Host "  • Chat Copilot: https://github.com/microsoft/chat-copilot" -ForegroundColor Cyan
Write-Host "  • Semantic Kernel: https://github.com/microsoft/semantic-kernel" -ForegroundColor Cyan
Write-Host "  • .NET 8 Docs: https://learn.microsoft.com/en-us/dotnet/" -ForegroundColor Cyan
Write-Host ""

if (-not $SkipDocker) {
    Write-Host "⚠ Important: If you installed Docker Desktop, please restart your computer" -ForegroundColor Yellow
    Write-Host "   and launch Docker Desktop before running Chat Copilot." -ForegroundColor Yellow
    Write-Host ""
}

Write-Host "✓ Setup complete! Your development environment is ready." -ForegroundColor Green

# SIG # Begin signature block
# MIIFsAYJKoZIhvcNAQcCoIIFoTCCBZ0CAQExDzANBglghkgBZQMEAgEFADB5Bgor
# BgEEAYI3AgEEoGswaTA0BgorBgEEAYI3AgEeMCYCAwEAAAQQH8w7YFlLCE63JNLG
# KX7zUQIBAAIBAAIBAAIBAAIBADAxMA0GCWCGSAFlAwQCAQUABCA6Og/RWzUZIWXx
# M1axCeuIm9QNnzVq4CBcBGQmyh7AeaCCAxwwggMYMIICAKADAgECAhAucOo6j/Tz
# gURLf5KSeLjZMA0GCSqGSIb3DQEBCwUAMCQxIjAgBgNVBAMMGUtlaXRoIFJhbnNv
# bSBDb2RlIFNpZ25pbmcwHhcNMjUxMjIyMjMzNzMyWhcNMzAxMjIyMjM0NzMyWjAk
# MSIwIAYDVQQDDBlLZWl0aCBSYW5zb20gQ29kZSBTaWduaW5nMIIBIjANBgkqhkiG
# 9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp5cy/OEQOv+frZJkC9jk1gTwEk0BvPFlVHuP
# +ib8Fxu81knk0t+6FSiUH/m+Nr/unW2MzMAK5keoP74r7mI5BWSpcUfLF+gfGL3M
# LAP/q9LtIUHQheO9lLnK4tFEFUwMCdRE3gOf0O38P4oZuQwvRTC665qdtfUNyDnz
# sYg59WfSEX3cA8QBzm4aNmXNW0YgeCuOdiwsXVEkMZeeHs5JazSo8UyBaF+xGeID
# EPzJi0XCvVPxL2LVixaDnH/0QeiBjpgA7apNbL/Rp7F1j1sr8KRWtmCDhjtLIn0N
# /cZVWslisMmPZpecWUwHA09ury7KDEHdAggXYitqrI1lkA5YkQIDAQABo0YwRDAO
# BgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYBBQUHAwMwHQYDVR0OBBYEFC2g
# QZgQ/iiOXSDfhpilwiFXuL4IMA0GCSqGSIb3DQEBCwUAA4IBAQA2yWvH11uBD3v2
# TaJ1bb34CF0eLwxRUvY7GAirHLkpeVleXMX8Lcr/DAru9LFxz1wKR+jq86seUvMD
# yFxK/Y43nrUWpDFW3tGELahaKllHR5SMi6ZpQ6b0i4MeJXHu5A2cRfg7IvTrt+zO
# +NJyf5hdkLMTpFHlAQzJx7TYb2kwUVuUqz/YgjLH2bDHLodo/NiyC/NOTGn9N3EU
# a8ch8q3W4mXbZHYdzr3QiN1wuQVYTZfOZwALl8fTu6vU+7Fm8m1073DJJCqVOmTu
# DVVNdYklhhqcja6I3l4Uc5kfqCSWCTTlIjCqUiUW84BhrX8j0CNgP8lRisjl3a5E
# YqyqP0WXMYIB6jCCAeYCAQEwODAkMSIwIAYDVQQDDBlLZWl0aCBSYW5zb20gQ29k
# ZSBTaWduaW5nAhAucOo6j/TzgURLf5KSeLjZMA0GCWCGSAFlAwQCAQUAoIGEMBgG
# CisGAQQBgjcCAQwxCjAIoAKAAKECgAAwGQYJKoZIhvcNAQkDMQwGCisGAQQBgjcC
# AQQwHAYKKwYBBAGCNwIBCzEOMAwGCisGAQQBgjcCARUwLwYJKoZIhvcNAQkEMSIE
# IJhjUXkF+36ZqaSkKp++y+BoPqu5zBOiPzTx4Q+9Tef3MA0GCSqGSIb3DQEBAQUA
# BIIBAB65GKcMmZ66bE+KM462aUMJ+P+WCXlFxRpEqEpRdUFNeTUaYDPOWqJxFBgl
# 7sRai0ZHoC/KtKSCs9X+u5ibLm4EE5HW3IzXyKjb2K/o8fivBe4iAKgii+wvPeBi
# CiMOFp0FF73ds9uiFx2fTpNlmbyPWFxsOYoNR+X8kBrr5UqVzTyYYIpWgdtjwx6v
# XS0k+Whidh+bmDkNrP0rBXk48feQF+ae4UzQUQDigXl/7RHN0rfQ09p3TvpZACr9
# VSD8rJuX0REw9+bE75FNr+xKwDqG37p5bDWjdX3ClDCFKMPKtuFu79PtwsTcXszM
# EWY3roc84lOJBlafLu4Na9gQ+8I=
# SIG # End signature block
