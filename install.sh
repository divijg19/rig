#!/bin/sh
set -eu

REPO="divijg19/rig"
VERSION="${RIG_VERSION:-v0.3.0}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

case "$os" in
  linux|darwin) : ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

if [ -n "${RIG_INSTALL_DIR:-}" ]; then
  install_dir="$RIG_INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
  install_dir="/usr/local/bin"
else
  install_dir="$HOME/.local/bin"
fi

mkdir -p "$install_dir"

tmp=$(mktemp -d)
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

asset="rig_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"

echo "downloading ${url}" >&2

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp/$asset"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$tmp/$asset" "$url"
else
  echo "need curl or wget" >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"

if [ ! -f "$tmp/rig" ]; then
  echo "archive did not contain rig binary" >&2
  exit 1
fi

chmod +x "$tmp/rig"
mv "$tmp/rig" "$install_dir/rig"

echo "installed: $install_dir/rig" >&2
echo "optional aliases: run 'rig alias'" >&2
