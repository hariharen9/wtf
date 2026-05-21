#!/bin/sh
# WTF Installer - macOS and Linux

set -e

# Styling colors
RESET="\033[0m"
BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RED="\033[31m"

print_step() {
  printf "${CYAN}🌀 %s${RESET}\n" "$1"
}

print_success() {
  printf "${GREEN}✨ %s${RESET}\n" "$1"
}

print_error() {
  printf "${RED}❌ Error: %s${RESET}\n" "$1" >&2
}

print_warning() {
  printf "${YELLOW}⚠️  %s${RESET}\n" "$1"
}

# ASCII Logo
printf "${BOLD}${GREEN}"
printf "  _    _  _____  ____ \n"
printf " | |  | ||_   _||  __| \n"
printf " | |  | |  | |  | |_   \n"
printf " | |/\\| |  | |  |  _|  \n"
printf " |  /\\  |  | |  | |    \n"
printf " |_/  \\_|  |_|  |_|    \n"
printf "${RESET}"
printf "${BOLD}Where's The File? - Sub-millisecond File Locator${RESET}\n\n"

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map OS & Arch to Release Archive Name
VERSION="1.0.0"
FILENAME=""

if [ "$OS" = "darwin" ]; then
  if [ "$ARCH" = "arm64" ]; then
    FILENAME="wtf-darwin-arm64.tar.gz"
  else
    FILENAME="wtf-darwin-amd64.tar.gz"
  fi
elif [ "$OS" = "linux" ]; then
  if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then
    FILENAME="wtf-linux-amd64.tar.gz"
  else
    print_error "Unsupported Linux architecture: $ARCH"
    exit 1
  fi
else
  print_error "Unsupported OS: $OS. Sticking to macOS and Linux for shell installer."
  exit 1
fi

# Define Install Paths
WTF_DIR="$HOME/.wtf"
BIN_DIR="$WTF_DIR/bin"
BINARY_PATH="$BIN_DIR/wtf"

# Create directories
mkdir -p "$BIN_DIR"

# Download URL
DOWNLOAD_URL="https://github.com/hariharen9/wtf/releases/download/v${VERSION}/${FILENAME}"
TEMP_ARCHIVE="/tmp/${FILENAME}"

print_step "Downloading native binary for ${OS}-${ARCH}..."
printf "   Source: %s\n" "$DOWNLOAD_URL"

# Check download tools
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_ARCHIVE"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TEMP_ARCHIVE" "$DOWNLOAD_URL"
else
  print_error "Could not find 'curl' or 'wget' in PATH. Please install one of them."
  exit 1
fi

print_step "Extracting archive to ${BIN_DIR}..."
tar -xzf "$TEMP_ARCHIVE" -C "$BIN_DIR"
rm -f "$TEMP_ARCHIVE"

# Mark binary executable
chmod +x "$BINARY_PATH"

print_success "WTF has been successfully installed!"

# Path checking and shell configuration recommendation
printf "\n"
case "$SHELL" in
  */zsh)
    SHELL_PROFILE="$HOME/.zshrc"
    ADD_PATH_CMD="export PATH=\"\$HOME/.wtf/bin:\$PATH\""
    ;;
  */bash)
    SHELL_PROFILE="$HOME/.bashrc"
    if [ "$OS" = "darwin" ]; then
      SHELL_PROFILE="$HOME/.bash_profile"
    fi
    ADD_PATH_CMD="export PATH=\"\$HOME/.wtf/bin:\$PATH\""
    ;;
  */fish)
    SHELL_PROFILE="$HOME/.config/fish/config.fish"
    ADD_PATH_CMD="fish_add_path \$HOME/.wtf/bin"
    ;;
  *)
    SHELL_PROFILE="your shell profile file"
    ADD_PATH_CMD="export PATH=\"\$HOME/.wtf/bin:\$PATH\""
    ;;
esac

# Check if PATH is already added
if echo "$PATH" | grep -q "$BIN_DIR"; then
  print_success "WTF binary directory is already in your PATH!"
else
  print_warning "WTF binary directory is NOT yet in your PATH."
  printf "\nTo add it to your PATH, run the following command:\n"
  printf "   ${BOLD}${GREEN}echo '%s' >> %s${RESET}\n" "$ADD_PATH_CMD" "$SHELL_PROFILE"
  printf "Then, restart your terminal or run:\n"
  printf "   ${BOLD}${GREEN}source %s${RESET}\n" "$SHELL_PROFILE"
fi

printf "\n⚡ ${BOLD}Next Steps:${RESET}\n"
printf "  1. Run ${BOLD}${GREEN}wtf update${RESET} to index your filesystem.\n"
printf "  2. Run ${BOLD}${GREEN}wtf${RESET} to launch the gorgeous interactive finder!\n\n"
