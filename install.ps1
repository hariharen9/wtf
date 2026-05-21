# WTF Installer - Windows PowerShell
$ErrorActionPreference = 'Stop'

# Styling colors
$RESET = "[0m"
$BOLD = "[1m"
$GREEN = "[32m"
$CYAN = "[36m"
$YELLOW = "[33m"
$RED = "[31m"

function Write-Host-Color($text, $colorEsc) {
    Write-Host "$([char]27)$colorEsc$text$([char]27)$RESET"
}

# ASCII Logo
Write-Host-Color "  _    _  _____  ____ " $GREEN
Write-Host-Color " | |  | ||_   _||  __| " $GREEN
Write-Host-Color " | |  | |  | |  | |_   " $GREEN
Write-Host-Color " | |/\ |  | |  |  _|  " $GREEN
Write-Host-Color " |  /\  |  | |  | |    " $GREEN
Write-Host-Color " |_/  \_|  |_|  |_|    " $GREEN
Write-Host "$([char]27)${BOLD}Where's The File? - Sub-millisecond File Locator$([char]27)$RESET`n"

# Verify Architecture
if ($env:PROCESSOR_ARCHITECTURE -ne "AMD64" -and $env:PROCESSOR_ARCHITEW6432 -ne "AMD64") {
    Write-Host-Color "❌ Error: WTF only supports 64-bit Windows architecture." $RED
    exit 1
}

$VERSION = "0.0.1"
$FILENAME = "wtf-windows-amd64.zip"
$WTF_DIR = Join-Path $HOME ".wtf"
$BIN_DIR = Join-Path $WTF_DIR "bin"
$BINARY_PATH = Join-Path $BIN_DIR "wtf.exe"

# Create target directory
if (-not (Test-Path $BIN_DIR)) {
    New-Item -ItemType Directory -Path $BIN_DIR -Force | Out-Null
}

$DOWNLOAD_URL = "https://github.com/hariharen9/wtf/releases/latest/download/$FILENAME"
$TEMP_ZIP = [System.IO.Path]::GetTempFileName() + ".zip"

Write-Host-Color "🌀 Downloading native binary for Windows-x64..." $CYAN
Write-Host "   Source: $DOWNLOAD_URL"

# Download Zip file using a multi-method resilient download block
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

$downloaded = $false
$lastError = ""

# Method 1: Invoke-WebRequest (Standard PS cmdlet, handles proxies and TLS perfectly)
if (-not $downloaded) {
    try {
        Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $TEMP_ZIP -UseBasicParsing -ErrorAction Stop
        $downloaded = $true
    }
    catch {
        $lastError = $_.Exception.Message
    }
}

# Method 2: System.Net.Http.HttpClient (.NET fallback)
if (-not $downloaded) {
    try {
        $httpClient = New-Object System.Net.Http.HttpClient
        $responseTask = $httpClient.GetByteArrayAsync($DOWNLOAD_URL)
        $responseTask.Wait()
        [System.IO.File]::WriteAllBytes($TEMP_ZIP, $responseTask.Result)
        $downloaded = $true
    }
    catch {
        $lastError = $_.Exception.Message
        if ($_.Exception.InnerException) {
            $lastError += " -> " + $_.Exception.InnerException.Message
        }
    }
}

# Method 3: Start-BitsTransfer (Windows BITS fallback)
if (-not $downloaded) {
    try {
        Start-BitsTransfer -Source $DOWNLOAD_URL -Destination $TEMP_ZIP -ErrorAction Stop
        $downloaded = $true
    }
    catch {
        $lastError = $_.Exception.Message
    }
}

if (-not $downloaded) {
    Write-Host-Color "❌ Failed to download binary. Check your connection or the release link." $RED
    Write-Host-Color "   Error details: $lastError" $RED
    exit 1
}

Write-Host-Color "📦 Extracting archive to $BIN_DIR..." $CYAN
try {
    Expand-Archive -Path $TEMP_ZIP -DestinationPath $BIN_DIR -Force
    Remove-Item $TEMP_ZIP -Force
}
catch {
    Write-Host-Color "❌ Failed to extract zip archive." $RED
    if (Test-Path $TEMP_ZIP) { Remove-Item $TEMP_ZIP -Force }
    exit 1
}

Write-Host-Color "✨ WTF has been successfully installed!" $GREEN

# Check and automate PATH environment variable updates
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$binDirExpanded = [System.IO.Path]::GetFullPath($BIN_DIR)

if ($userPath -like "*$binDirExpanded*") {
    Write-Host-Color "✨ WTF binary directory is already in your PATH!" $GREEN
}
else {
    Write-Host-Color "⚠️  WTF binary directory is NOT yet in your PATH." $YELLOW
    Write-Host "🌀 Automatically adding it to your User PATH variable..." -NoNewline
    try {
        $newUserPath = $userPath + ";" + $binDirExpanded
        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
        $env:Path += ";" + $binDirExpanded
        Write-Host-Color " Done!" $GREEN
    }
    catch {
        Write-Host-Color " Failed!" $RED
        Write-Host "Please add the directory manually to your PATH environment variable:"
        Write-Host "   Directory: $binDirExpanded"
    }
}

Write-Host "`n⚡ Next Steps:"
Write-Host-Color "  1. Restart your terminal (so the new PATH environment takes full effect)." $GREEN
Write-Host-Color "  2. Run 'wtf update' to index your filesystem." $GREEN
Write-Host-Color "  3. Run 'wtf' to launch the gorgeous interactive finder!" $GREEN
Write-Host ""
