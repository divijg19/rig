#!/bin/sh
set -eu

REPO="divijg19/rig"
VERSION="${RIG_VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

case "$os" in
  linux|darwin) : ;;
  *)
    echo "unsupported OS: $os" >&2
    echo "On Windows, invoke: rig run / rig check / rig dev / rig start directly." >&2
    exit 1
    ;;
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
if [ "$VERSION" = "latest" ]; then
  base_url="https://github.com/${REPO}/releases/latest/download"
else
  base_url="https://github.com/${REPO}/releases/download/${VERSION}"
fi
url="${base_url}/${asset}"
checksum_url="${base_url}/${asset}.sha256"

echo "downloading ${url}" >&2

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp/$asset"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$tmp/$asset" "$url"
else
  echo "need curl or wget" >&2
  exit 1
fi

if [ ! -s "$tmp/$asset" ]; then
  echo "download failed or empty asset: $asset" >&2
  exit 1
fi

checksum_file="$tmp/${asset}.sha256"
checksum_present=0
if command -v curl >/dev/null 2>&1; then
  if curl -fsSL "$checksum_url" -o "$checksum_file"; then
    checksum_present=1
  fi
elif command -v wget >/dev/null 2>&1; then
  if wget -qO "$checksum_file" "$checksum_url"; then
    checksum_present=1
  fi
fi

if [ "$checksum_present" -eq 1 ]; then
  expected=$(awk '{print $1}' "$checksum_file" | head -n 1)
  if [ -z "$expected" ]; then
    echo "checksum file is empty or invalid: ${asset}.sha256" >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$tmp/$asset" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')
  else
    echo "checksum verification requested but sha256 tool unavailable (need sha256sum or shasum)" >&2
    exit 1
  fi

  if [ "$actual" != "$expected" ]; then
    echo "checksum mismatch for $asset" >&2
    echo "expected: $expected" >&2
    echo "actual:   $actual" >&2
    exit 1
  fi
  echo "verified: ${asset}.sha256" >&2
fi

# Release artifact contract: archive must contain exactly one file named 'rig'.
entries=$(tar -tzf "$tmp/$asset" | sed 's|^\./||' | sed 's|/$||')
count=$(printf "%s\n" "$entries" | sed '/^$/d' | wc -l | tr -d ' ')
if [ "$count" -ne 1 ] || [ "$entries" != "rig" ]; then
  echo "archive must contain exactly one file named 'rig'" >&2
  echo "found entries:" >&2
  printf "  %s\n" $entries >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"

if [ ! -f "$tmp/rig" ]; then
  echo "archive did not contain rig binary" >&2
  exit 1
fi

chmod +x "$tmp/rig"
tmp_out="$install_dir/rig.tmp.$$"
cp "$tmp/rig" "$tmp_out"
chmod +x "$tmp_out"
mv "$tmp_out" "$install_dir/rig"

echo "installed: $install_dir/rig" >&2
path_hint=1
old_ifs=$IFS
IFS=:
for p in $PATH; do
  if [ "$p" = "$install_dir" ]; then
    path_hint=0
    break
  fi
done
IFS=$old_ifs
if [ "$path_hint" -eq 1 ]; then
  echo "ensure it's on PATH: export PATH=\"$install_dir:\$PATH\"" >&2
fi

# Create alias symlinks on Unix only.
# Rules:
# - Use ln -sf
# - Do not fail if already exists
# - Do not overwrite non-symlink files
# - If creation fails, warn and continue
make_alias() {
  name="$1"
  link="$install_dir/$name"

  if [ -e "$link" ] && [ ! -L "$link" ]; then
    echo "warning: not overwriting non-symlink file: $link" >&2
    return 0
  fi

  if ! ln -sf "$install_dir/rig" "$link" 2>/dev/null; then
    echo "warning: failed to create symlink alias: $link" >&2
    return 0
  fi
}

make_alias "rir"
make_alias "ric"
make_alias "rid"
make_alias "ris"

echo "aliases: rir/ric/rid/ris (symlinks)" >&2
