#!/bin/sh
# Installs compress-env on macOS / Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh
#
# Project mode downloads the binary into the current folder instead, to run it
# as ./compress-env:
#
#   curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh -s -- --project
#
# Environment:
#   VERSION      release to install, e.g. v0.0.1 (default: latest)
#   INSTALL_DIR  where to put the binary
#                (default: /usr/local/bin if writable, else ~/.local/bin; project mode: .)
#   PROJECT=1    same as --project
set -eu

repo="cavaldos/compress-env"
base="${COMPRESS_ENV_BASE_URL:-https://github.com/$repo/releases}"
version="${VERSION:-latest}"
project="${PROJECT:-}"

for arg in "$@"; do
  case "$arg" in
    --project) project=1 ;;
    *) echo "error: unknown option '$arg'" >&2; exit 1 ;;
  esac
done

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
elif [ -n "$project" ]; then
  dir=.
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
if [ -n "$project" ]; then
  run="$dir/compress-env"
  case "$dir" in /*) ;; *) run="./${run#./}" ;; esac
  echo "Run it with: $run"
else
  case ":$PATH:" in
    *":$dir:"*) ;;
    *) echo "Note: $dir is not in your PATH. Add it, e.g.: export PATH=\"$dir:\$PATH\"" ;;
  esac
fi
"$dir/compress-env" -v | head -1
