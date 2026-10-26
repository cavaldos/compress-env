#!/bin/sh
# Installs compress-env on macOS / Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh
#
# Environment:
#   VERSION      release to install, e.g. v0.0.1 (default: latest)
#   INSTALL_DIR  where to put the binary (default: /usr/local/bin if writable, else ~/.local/bin)
set -eu

repo="cavaldos/compress-env"
base="${COMPRESS_ENV_BASE_URL:-https://github.com/$repo/releases}"
version="${VERSION:-latest}"

case "$(uname -s)" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *) echo "error: unsupported OS '$(uname -s)' (on Windows use install.ps1)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "error: unsupported architecture '$(uname -m)'" >&2; exit 1 ;;
esac

if [ -n "${INSTALL_DIR:-}" ]; then
  dir="$INSTALL_DIR"
elif [ -w /usr/local/bin ]; then
  dir=/usr/local/bin
else
  dir="$HOME/.local/bin"
fi

file="compress-env-$os-$arch"
if [ "$version" = "latest" ]; then
  url="$base/latest/download"
else
  url="$base/download/$version"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $file ($version)..."
curl -fsSL "$url/$file" -o "$tmp/$file"
curl -fsSL "$url/checksums.txt" -o "$tmp/checksums.txt"

expected=$(grep " $file\$" "$tmp/checksums.txt" | cut -d ' ' -f 1)
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$file" | cut -d ' ' -f 1)
else
  actual=$(shasum -a 256 "$tmp/$file" | cut -d ' ' -f 1)
fi
if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
  echo "error: checksum mismatch, not installing" >&2
  exit 1
fi

mkdir -p "$dir"
chmod +x "$tmp/$file"
mv "$tmp/$file" "$dir/compress-env"

echo "Installed: $dir/compress-env"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Note: $dir is not in your PATH. Add it, e.g.: export PATH=\"$dir:\$PATH\"" ;;
esac
"$dir/compress-env" -v | head -1
