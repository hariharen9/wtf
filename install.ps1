$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

# =========================================================
# CONFIG
# =========================================================

$RepoOwner = 'hariharen9'
$RepoName  = 'wtf'
$ToolName  = 'wtf'
$Version   = 'latest'

$InstallRoot = Join-Path $HOME '.wtf'
$BinDir      = Join-Path $InstallRoot 'bin'
$TempDir     = Join-Path ([System.IO.Path]::GetTempPath()) 'wtf-installer'

# =========================================================
# STYLING
# =========================================================

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[ OK ] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host "[FAIL] $Message" -ForegroundColor Red
}

function Show-Banner {
    Write-Host ''
    Write-Host '██╗    ██╗████████╗███████╗' -ForegroundColor Green
    Write-Host '██║    ██║╚══██╔══╝██╔════╝' -ForegroundColor Green
    Write-Host '██║ █╗ ██║   ██║   █████╗  ' -ForegroundColor Green
    Write-Host '██║███╗██║   ██║   ██╔══╝  ' -ForegroundColor Green
    Write-Host '╚███╔███╔╝   ██║   ██║     ' -ForegroundColor Green
    Write-Host ' ╚══╝╚══╝    ╚═╝   ╚═╝     ' -ForegroundColor Green
    Write-Host ''
    Write-Host 'Where''s The File? - Native lightning-fast file search' -ForegroundColor DarkGray
    Write-Host ''
}

# =========================================================
# HELPERS
# =========================================================

function Ensure-Directory {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
}

function Test-Command {
    param([string]$Command)

    return $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
}

function Get-Architecture {

    # Reliable architecture detection for all Windows PowerShell versions

    $arch = $env:PROCESSOR_ARCHITECTURE

    # Handle WOW64
    if ($env:PROCESSOR_ARCHITEW6432) {
        $arch = $env:PROCESSOR_ARCHITEW6432
    }

    switch ($arch.ToUpper()) {

        'AMD64' {
            return 'windows-amd64.zip'
        }

        'ARM64' {
            return 'windows-arm64.zip'
        }

        default {
            throw "Unsupported architecture: $arch"
        }
    }
}

function Get-LatestReleaseUrl {
    param([string]$AssetName)

    return "https://github.com/$RepoOwner/$RepoName/releases/latest/download/$ToolName-$AssetName"
}

function Remove-IfExists {
    param([string]$Path)

    if (Test-Path $Path) {
        Remove-Item $Path -Force -Recurse -ErrorAction SilentlyContinue
    }
}

function Download-File {
    param(
        [string]$Url,
        [string]$Output
    )

    Write-Info "Downloading package..."
    Write-Host "       $Url" -ForegroundColor DarkGray

    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

    $methodsTried = @()

    # -----------------------------------------------------
    # Method 1: Invoke-WebRequest
    # -----------------------------------------------------

    try {
        $methodsTried += 'Invoke-WebRequest'

        Invoke-WebRequest `
            -Uri $Url `
            -OutFile $Output `
            -UseBasicParsing `
            -Headers @{ 'User-Agent' = 'wtf-installer' }

        if ((Get-Item $Output).Length -gt 0) {
            Write-Success 'Download completed using Invoke-WebRequest'
            return
        }
    }
    catch {
        Write-Warn "Invoke-WebRequest failed: $($_.Exception.Message)"
    }

    # -----------------------------------------------------
    # Method 2: curl.exe
    # -----------------------------------------------------

    if (Test-Command 'curl.exe') {
        try {
            $methodsTried += 'curl.exe'

            & curl.exe -L --fail --silent --show-error $Url -o $Output

            if ((Get-Item $Output).Length -gt 0) {
                Write-Success 'Download completed using curl.exe'
                return
            }
        }
        catch {
            Write-Warn "curl.exe failed: $($_.Exception.Message)"
        }
    }

    # -----------------------------------------------------
    # Method 3: BITS
    # -----------------------------------------------------

    if (Get-Command Start-BitsTransfer -ErrorAction SilentlyContinue) {
        try {
            $methodsTried += 'BITS'

            Start-BitsTransfer -Source $Url -Destination $Output

            if ((Get-Item $Output).Length -gt 0) {
                Write-Success 'Download completed using BITS'
                return
            }
        }
        catch {
            Write-Warn "BITS failed: $($_.Exception.Message)"
        }
    }

    throw "All download methods failed. Methods attempted: $($methodsTried -join ', ')"
}

function Add-ToPath {
    param([string]$PathToAdd)

    $CurrentUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')

    $PathItems = @()

    if ($CurrentUserPath) {
        $PathItems = $CurrentUserPath.Split(';')
    }

    $AlreadyExists = $false

    foreach ($item in $PathItems) {
        if ($item.TrimEnd('\\') -ieq $PathToAdd.TrimEnd('\\')) {
            $AlreadyExists = $true
            break
        }
    }

    if ($AlreadyExists) {
        Write-Success 'PATH already contains WTF binary directory'
        return
    }

    Write-Info 'Adding WTF to user PATH...'

    $NewPath = ($PathItems + $PathToAdd | Where-Object { $_ -and $_.Trim() -ne '' }) -join ';'

    [Environment]::SetEnvironmentVariable('Path', $NewPath, 'User')

    # Update current session immediately
    $env:Path += ";$PathToAdd"

    Write-Success 'PATH updated successfully'
}

function Test-Admin {
    $currentIdentity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentIdentity)

    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# =========================================================
# MAIN
# =========================================================

try {
    Clear-Host
    # Show-Banner

    Write-Info 'Preparing installer environment...'

    Ensure-Directory $InstallRoot
    Ensure-Directory $BinDir
    Ensure-Directory $TempDir

    if (Test-Admin) {
        Write-Info 'Running with Administrator privileges'
    }
    else {
        Write-Warn 'Not running as Administrator (this is okay for user install)'
    }

    # -----------------------------------------------------
    # Detect architecture
    # -----------------------------------------------------

    $AssetName = Get-Architecture

    Write-Info "Detected package: $AssetName"

    $DownloadUrl = Get-LatestReleaseUrl -AssetName $AssetName

    # -----------------------------------------------------
    # Download
    # -----------------------------------------------------

    $ZipPath = Join-Path $TempDir 'wtf.zip'

    Remove-IfExists $ZipPath

    Download-File -Url $DownloadUrl -Output $ZipPath

    if (-not (Test-Path $ZipPath)) {
        throw 'Downloaded ZIP file not found'
    }

    $ZipSize = [Math]::Round((Get-Item $ZipPath).Length / 1MB, 2)
    Write-Info "Downloaded archive size: ${ZipSize} MB"

    # -----------------------------------------------------
    # Extract
    # -----------------------------------------------------

    Write-Info 'Extracting package...'

    try {
        Expand-Archive -Path $ZipPath -DestinationPath $BinDir -Force
    }
    catch {
        throw "Extraction failed: $($_.Exception.Message)"
    }

    # -----------------------------------------------------
    # Validate install
    # -----------------------------------------------------

    $ExePath = Join-Path $BinDir 'wtf.exe'

    if (-not (Test-Path $ExePath)) {
        throw 'wtf.exe was not found after extraction'
    }

    Write-Success 'Binary extracted successfully'

    # -----------------------------------------------------
    # PATH setup
    # -----------------------------------------------------

    Add-ToPath -PathToAdd $BinDir

    # -----------------------------------------------------
    # Cleanup
    # -----------------------------------------------------

    Remove-IfExists $TempDir

    # -----------------------------------------------------
    # Final output
    # -----------------------------------------------------

    Write-Host ''
    Write-Success 'WTF installed successfully!'
    Write-Host ''

    Write-Host 'Installation Directory:' -ForegroundColor Cyan
    Write-Host "  $InstallRoot"
    Write-Host ''

    Write-Host 'Binary:' -ForegroundColor Cyan
    Write-Host "  $ExePath"
    Write-Host ''

    Write-Host 'Next Steps:' -ForegroundColor Green
    Write-Host '  1. Restart your terminal'
    Write-Host '  2. Run: wtf update'
    Write-Host '  3. Run: wtf'
    Write-Host ''

    # -----------------------------------------------------
    # Version check
    # -----------------------------------------------------

    try {
        Write-Info 'Verifying installation...'

        $versionOutput = & $ExePath -v 2>$null

        if ($LASTEXITCODE -eq 0) {
            Write-Success "Installed version: $versionOutput"
        }
    }
    catch {
        Write-Warn 'Installed successfully, but version verification failed'
    }
}
catch {
    Write-Host ''
    Write-Err $_.Exception.Message
    Write-Host ''

    Write-Host 'Troubleshooting:' -ForegroundColor Yellow
    Write-Host '  • Ensure internet access is available'
    Write-Host '  • Ensure GitHub is reachable'
    Write-Host '  • Try running PowerShell as Administrator'
    Write-Host '  • Check antivirus or Defender restrictions'
    Write-Host '  • Verify the release asset exists on GitHub'
    Write-Host ''

    exit 1
}
