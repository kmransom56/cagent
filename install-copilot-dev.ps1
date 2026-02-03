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
$dotnetVersion = $null
try {
    $dotnetVersion = (Get-Command dotnet -ErrorAction Stop).Version.ToString() 2>$null
} catch {
    # Try to find dotnet in common locations
    $dotnetPaths = @(
        "${env:ProgramFiles}\dotnet\dotnet.exe",
        "${env:ProgramFiles(x86)}\dotnet\dotnet.exe",
        "$env:USERPROFILE\.dotnet\dotnet.exe"
    )
    foreach ($path in $dotnetPaths) {
        if (Test-Path $path) {
            $dotnetVersion = & $path --version 2>$null
            if ($dotnetVersion) { break }
        }
    }
}

if ($null -eq $dotnetVersion -or $dotnetVersion -eq "") {
    Write-Host "Installing .NET 8 SDK..." -ForegroundColor Yellow
    choco install dotnet-8.0-sdk -y 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ .NET 8 SDK installed successfully" -ForegroundColor Green
        # Refresh PATH from registry
        $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
        # Try to find dotnet after installation
        Start-Sleep -Seconds 2
        $dotnetPaths = @(
            "${env:ProgramFiles}\dotnet\dotnet.exe",
            "${env:ProgramFiles(x86)}\dotnet\dotnet.exe"
        )
        foreach ($path in $dotnetPaths) {
            if (Test-Path $path) {
                $dotnetVersion = & $path --version 2>$null
                if ($dotnetVersion) {
                    Write-Host "  Installed version: $dotnetVersion" -ForegroundColor Cyan
                    break
                }
            }
        }
    } else {
        Write-Host "⚠ .NET 8 SDK installation may have failed. Check Chocolatey logs." -ForegroundColor Yellow
        Write-Host "  You may need to install manually from: https://dotnet.microsoft.com/download/dotnet/8.0" -ForegroundColor Yellow
    }
} else {
    Write-Host "✓ .NET already installed (version $dotnetVersion)" -ForegroundColor Green
}

# Verify required .NET workloads (only if dotnet is available)
if ($dotnetVersion) {
    Write-Host "Installing .NET workloads..." -ForegroundColor Yellow
    $dotnetCmd = (Get-Command dotnet -ErrorAction SilentlyContinue).Source
    if (-not $dotnetCmd) {
        $dotnetPaths = @(
            "${env:ProgramFiles}\dotnet\dotnet.exe",
            "${env:ProgramFiles(x86)}\dotnet\dotnet.exe"
        )
        foreach ($path in $dotnetPaths) {
            if (Test-Path $path) {
                $dotnetCmd = $path
                break
            }
        }
    }
    if ($dotnetCmd) {
        & $dotnetCmd workload restore 2>$null
        Write-Host "✓ .NET workloads ready" -ForegroundColor Green
    }
}

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
$yarnVersion = $null
try {
    $yarnVersion = yarn --version 2>$null
} catch {
    # Yarn not found
}

if ($null -eq $yarnVersion -or $yarnVersion -eq "") {
    Write-Host "Installing Yarn (Classic)..." -ForegroundColor Yellow
    try {
        npm install -g yarn --loglevel=error 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            # Refresh PATH
            $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH","User")
            Start-Sleep -Seconds 1
            try {
                yarn set version classic 2>&1 | Out-Null
                $yarnVersion = yarn --version 2>$null
                if ($yarnVersion) {
                    Write-Host "✓ Yarn installed successfully (version $yarnVersion)" -ForegroundColor Green
                } else {
                    Write-Host "⚠ Yarn installed but version check failed" -ForegroundColor Yellow
                }
            } catch {
                Write-Host "⚠ Yarn installed but classic version setup had issues" -ForegroundColor Yellow
                Write-Host "  You may need to run: yarn set version classic" -ForegroundColor Gray
            }
        } else {
            Write-Host "⚠ Yarn installation may have failed" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "⚠ Failed to install Yarn: $_" -ForegroundColor Yellow
    }
} else {
    Write-Host "✓ Yarn already installed (version $yarnVersion)" -ForegroundColor Green
    # Ensure classic version
    try {
        yarn set version classic 2>&1 | Out-Null
    } catch {
        # Ignore errors if already classic
    }
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
    $codeCmd = $null
    try {
        $codeCmd = Get-Command code -ErrorAction Stop
    } catch {
        # Check common VS Code installation paths
        $codePaths = @(
            "${env:ProgramFiles}\Microsoft VS Code\bin\code.cmd",
            "${env:ProgramFiles(x86)}\Microsoft VS Code\bin\code.cmd",
            "${env:LOCALAPPDATA}\Programs\Microsoft VS Code\bin\code.cmd",
            "$env:USERPROFILE\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd"
        )
        foreach ($path in $codePaths) {
            if (Test-Path $path) {
                $codeCmd = $path
                break
            }
        }
    }
    
    if ($null -ne $codeCmd) {
        Write-Host "Installing recommended extensions..." -ForegroundColor Yellow
        if ($codeCmd -is [string]) {
            & $codeCmd --install-extension ms-dotnettools.csharp 2>&1 | Out-Null
            & $codeCmd --install-extension ms-dotnettools.vscode-dotnet-runtime 2>&1 | Out-Null
            & $codeCmd --install-extension ms-python.python 2>&1 | Out-Null
            & $codeCmd --install-extension ms-vscode.makefile-tools 2>&1 | Out-Null
            & $codeCmd --install-extension eamodio.gitlens 2>&1 | Out-Null
            & $codeCmd --install-extension ms-vscode-remote.remote-wsl 2>&1 | Out-Null
        } else {
            code --install-extension ms-dotnettools.csharp 2>&1 | Out-Null
            code --install-extension ms-dotnettools.vscode-dotnet-runtime 2>&1 | Out-Null
            code --install-extension ms-python.python 2>&1 | Out-Null
            code --install-extension ms-vscode.makefile-tools 2>&1 | Out-Null
            code --install-extension eamodio.gitlens 2>&1 | Out-Null
            code --install-extension ms-vscode-remote.remote-wsl 2>&1 | Out-Null
        }
        Write-Host "✓ VS Code extensions installed" -ForegroundColor Green
    } else {
        Write-Host "⊘ VS Code not found or not in PATH. Skipping extensions." -ForegroundColor Yellow
        Write-Host "  Install VS Code from: https://code.visualstudio.com/" -ForegroundColor Gray
    }
}

# Final verification
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Installation Summary                                          ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Write-Host ""
Write-Host "Installed versions:" -ForegroundColor Green
# Safely get versions with error handling
$dotnetVer = "Not installed"
try {
    $dotnetCmd = (Get-Command dotnet -ErrorAction SilentlyContinue).Source
    if (-not $dotnetCmd) {
        $dotnetPaths = @(
            "${env:ProgramFiles}\dotnet\dotnet.exe",
            "${env:ProgramFiles(x86)}\dotnet\dotnet.exe"
        )
        foreach ($path in $dotnetPaths) {
            if (Test-Path $path) {
                $dotnetCmd = $path
                break
            }
        }
    }
    if ($dotnetCmd) {
        $dotnetVer = & $dotnetCmd --version 2>$null
    }
} catch {
    $dotnetVer = "Not found"
}

$yarnVer = "Not installed"
try {
    $yarnVer = yarn --version 2>$null
} catch {
    $yarnVer = "Not found"
}

Write-Host "  .NET SDK:    $dotnetVer" -ForegroundColor White
Write-Host "  Node.js:     $(node --version)" -ForegroundColor White
Write-Host "  Yarn:        $yarnVer" -ForegroundColor White
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
# KX7zUQIBAAIBAAIBAAIBAAIBADAxMA0GCWCGSAFlAwQCAQUABCC+X8gyrM/SdJdg
# iANw8JlZB5rtixd4EPRqiQeKfCcjOqCCAxwwggMYMIICAKADAgECAhAucOo6j/Tz
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
# ILWBrfmuYhejT13ntWouVuSZOvN0aX+JEloORfAiiGVhMA0GCSqGSIb3DQEBAQUA
# BIIBAJenzso9gMHO0ri4xRwFmu7p50/kt2vSBwjn4n+ZkiiUZhMSNr0TbhzDxz7A
# /WQYm+EqzOcu/kCQddTWA3il9QF/U+3y4Orrefms4T1AT2/DKJjFFgUnx0nX6mLP
# DEhQnKJ12bJ478OilTqYsZdzvJhSjB9/TA5NcfzHIa0RgAGBgY/XOnoI41QXW4yJ
# f2A4N+hUu5+vF4LfHFu8slutoR6gbCeVFLCQYCaPi9pY8zN8gh/79pk6i5bBE/1p
# XeeseZGJT6cRodQ7Htb7Zwodbf75Jl8WwmI6zLoCP/lvV85C+7myBcC65IvGIvKZ
# mWj7XQRvT/kstO8x+Bo0EUq4jiM=
# SIG # End signature block
