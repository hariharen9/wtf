const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const REPO = 'hariharen9/wtf';
const VERSION = '1.0.0'; // Falls back to latest or matches npm package version

function getTargetAsset() {
  const platform = process.platform;
  const arch = process.arch;

  if (platform === 'darwin') {
    if (arch === 'arm64') return `wtf-darwin-arm64.tar.gz`;
    return `wtf-darwin-amd64.tar.gz`;
  }
  if (platform === 'linux') {
    if (arch === 'x64') return `wtf-linux-amd64.tar.gz`;
  }
  if (platform === 'win32') {
    if (arch === 'x64' || arch === 'ia32') return `wtf-windows-amd64.zip`;
  }

  throw new Error(`Unsupported platform/architecture: ${platform}/${arch}`);
}

function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      // Handle HTTP redirects (vital for GitHub Releases)
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return downloadFile(res.headers.location, destPath).then(resolve).catch(reject);
      }

      if (res.statusCode !== 200) {
        return reject(new Error(`Failed to download binary: Status Code ${res.statusCode}`));
      }

      const file = fs.createWriteStream(destPath);
      res.pipe(file);

      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(destPath, () => {});
      reject(err);
    });
  });
}

async function main() {
  const binDir = path.join(__dirname, 'bin');
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  try {
    const assetName = getTargetAsset();
    const tempFile = path.join(__dirname, assetName);
    
    // Construct download URL
    // e.g. https://github.com/hariharen9/wtf/releases/download/v1.0.0/wtf-windows-amd64.zip
    const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;

    console.log(`🌀 Downloading WTF binary for ${process.platform}-${process.arch}...`);
    console.log(`   Source: ${downloadUrl}`);
    
    await downloadFile(downloadUrl, tempFile);
    console.log(`📦 Extracting archive to ${binDir}...`);

    // Use system extraction tools (tar works on Win10+, macOS, and Linux)
    if (assetName.endsWith('.tar.gz')) {
      execSync(`tar -xzf "${tempFile}" -C "${binDir}"`);
    } else if (assetName.endsWith('.zip')) {
      // Windows tar natively handles zip files
      execSync(`tar -xf "${tempFile}" -C "${binDir}"`);
    }

    // Clean up temporary archive file
    fs.unlinkSync(tempFile);

    // Verify binary exists and set permissions on Unix systems
    const isWindows = process.platform === 'win32';
    const binaryName = isWindows ? 'wtf.exe' : 'wtf';
    const binaryPath = path.join(binDir, binaryName);

    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary file was not found in extracted archive at ${binaryPath}`);
    }

    if (!isWindows) {
      fs.chmodSync(binaryPath, 0o755); // make executable
    }

    console.log('✨ WTF native binary installed successfully!');
  } catch (err) {
    console.error('❌ Failed to install WTF binary.');
    console.error(`   Error details: ${err.message}`);
    console.warn('\n💡 Running in offline mode or dev environment?');
    console.warn('   Ensure you build the binary locally: "go build -o npm/bin/wtf main.go"');
    
    // We exit with 0 so npm install doesn't crash completely in offline/dev workspaces
    process.exit(0);
  }
}

main();
