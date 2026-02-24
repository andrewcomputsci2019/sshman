#!/usr/bin/env bash
set -e

REPO="andrewcomputsci2019/sshman"
BINARY="ssh-man"

OS="$(uname -s)"
ARCH="$(uname -m)"

# Normalize OS
case "$OS" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  *)        echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Normalize Arch
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

FILE="${BINARY}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${FILE}"

TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${FILE}"

echo "Creating Tmp extraction location $TMP_DIR"

echo "Downloading $URL..."
curl -fsSL "$URL" -o "$ARCHIVE_PATH"

echo "Extracting..."
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

echo "Installing..."

INSTALL_DIR="/usr/local/bin"

if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/${BINARY}" "$INSTALL_DIR/${BINARY}"
else
    sudo mv "${TMP_DIR}/${BINARY}" "$INSTALL_DIR/${BINARY}"
fi

chmod +x "$INSTALL_DIR/${BINARY}"

echo "Cleaning up ${TMP_DIR}"
rm -r "$TMP_DIR"


read -p "Run ssh-man init [y/n]" userConf

case $userConf in
  [Yy]* ) echo "running ssh-man init"; ssh-man --init ;;
  * ) echo "skipping ssh-man init step make sure to run it before first use of the program" ;;
esac

echo "ssh-man installed successfully"
