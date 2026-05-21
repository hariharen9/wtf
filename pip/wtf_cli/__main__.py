import os
import sys
import shutil
import urllib.request
import zipfile
import tarfile
import platform
import subprocess

REPO = "hariharen9/wtf"
VERSION = "0.0.1"

def get_target_details():
    system = platform.system().lower()
    machine = platform.machine().lower()

    if system == "darwin":
        if "arm" in machine or "aarch" in machine:
            return "wtf-darwin-arm64.tar.gz", True
        return "wtf-darwin-amd64.tar.gz", True
    elif system == "linux":
        if "amd64" in machine or "x86_64" in machine:
            return "wtf-linux-amd64.tar.gz", True
    elif system == "windows":
        if "amd64" in machine or "x86_64" in machine or "x86" in machine:
            return "wtf-windows-amd64.zip", False

    raise RuntimeError(f"Unsupported system/architecture: {system}/{machine}")

def download_and_extract(download_url, archive_name, bin_dir, is_tar):
    temp_archive = os.path.join(bin_dir, archive_name)
    
    print(f"🌀 Downloading WTF native binary...")
    print(f"   Source: {download_url}")
    
    try:
        # Download archive
        with urllib.request.urlopen(download_url) as response, open(temp_archive, 'wb') as out_file:
            shutil.copyfileobj(response, out_file)
            
        print(f"📦 Extracting archive to {bin_dir}...")
        
        # Extract archive using Python built-in modules
        if is_tar:
            with tarfile.open(temp_archive, "r:gz") as tar:
                tar.extractall(path=bin_dir)
        else:
            with zipfile.ZipFile(temp_archive, 'r') as zip_ref:
                zip_ref.extractall(bin_dir)
                
        # Remove temporary archive
        os.remove(temp_archive)
        
    except Exception as e:
        if os.path.exists(temp_archive):
            os.remove(temp_archive)
        raise RuntimeError(f"Failed to fetch binary: {e}")

def main():
    is_windows = platform.system().lower() == "windows"
    binary_name = "wtf.exe" if is_windows else "wtf"
    
    # Store binary in user home folder: ~/.wtf/bin/
    user_home = os.path.expanduser("~")
    bin_dir = os.path.join(user_home, ".wtf", "bin")
    binary_path = os.path.join(bin_dir, binary_name)

    # Resolve arguments
    cli_args = sys.argv[1:]

    # Download binary on-demand if missing
    if not os.path.exists(binary_path):
        os.makedirs(bin_dir, exist_ok=True)
        try:
            archive_name, is_tar = get_target_details()
            download_url = f"https://github.com/{REPO}/releases/latest/download/{archive_name}"
            download_and_extract(download_url, archive_name, bin_dir, is_tar)
            
            # Set executable bit on Unix-based OS
            if not is_windows:
                os.chmod(binary_path, 0o755)
            print("✨ WTF installed successfully!\n")
        except Exception as err:
            print(f"❌ Error setting up WTF native binary: {err}", file=sys.stderr)
            print("\n💡 Running in developer mode? Please compile the binary locally:")
            print(f"   go build -o {binary_path} main.go")
            sys.exit(1)

    # Launch native binary
    if is_windows:
        # On Windows, spawn a subprocess and forward arguments
        try:
            res = subprocess.run([binary_path] + cli_args)
            sys.exit(res.returncode)
        except KeyboardInterrupt:
            # Handle user interrupt gracefully
            sys.exit(130)
    else:
        # On macOS and Linux, replace the current Python process completely (os.execv)
        # This keeps job control, signals, and pipes 100% native with zero overhead!
        try:
            os.execv(binary_path, [binary_path] + cli_args)
        except Exception as e:
            print(f"❌ Failed to overlay process: {e}", file=sys.stderr)
            sys.exit(1)

if __name__ == "__main__":
    main()
