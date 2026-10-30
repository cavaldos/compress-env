#!/bin/sh
# Installs compress-env on macOS / Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh
#
# Project mode puts the binaries inside the current project (default ./bin),
# so they can be committed and teammates can run ./bin/compress-env without
# installing anything:
#
#   curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh -s -- --project
#
# Environment:
#   VERSION      release to install, e.g. v0.0.1 (default: latest)
#   INSTALL_DIR  where to put the binary
#                (default: /usr/local/bin if writable, else ~/.local/bin; project mode: ./bin)
#   PROJECT=1    same as --project
#   PLATFORMS    project mode only: platforms to include, e.g. "darwin-arm64 linux-amd64"
#                (default: all of them, so it runs for everyone on the team)
set -eu

repo="cavaldos/compress-env"
base="${COMPRESS_ENV_BASE_URL:-https://github.com/$repo/releases}"
version="${VERSION:-latest}"
project="${PROJECT:-}"
all_platforms="darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64"

for arg in "$@"; do
  case "$arg" in
    --project) project=1 ;;
    *) echo "error: unknown option '$arg'" >&2; exit 1 ;;
  esac
done

if [ "$version" = "latest" ]; then
  url="$base/latest/download"
else
  url="$base/download/$version"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d ' ' -f 1
  else
    shasum -a 256 "$1" | cut -d ' ' -f 1
  fi
}

# fetch <file>: downloads a release file into $tmp and checks its checksum.
fetch() {
  echo "Downloading $1 ($version)..."
  curl -fsSL "$url/$1" -o "$tmp/$1"
  expected=$(grep " $1\$" "$tmp/checksums.txt" | cut -d ' ' -f 1)
  if [ -z "$expected" ] || [ "$expected" != "$(sha256 "$tmp/$1")" ]; then
    echo "error: checksum mismatch for $1, not installing" >&2
    exit 1
  fi
}

curl -fsSL "$url/checksums.txt" -o "$tmp/checksums.txt"

# --- project mode -----------------------------------------------------------

if [ -n "$project" ]; then
  dir="${INSTALL_DIR:-bin}"
  mkdir -p "$dir"

  for p in ${PLATFORMS:-$all_platforms}; do
    file="compress-env-$p"
    case "$p" in windows-*) file="$file.exe" ;; esac
    fetch "$file"
    chmod +x "$tmp/$file"
    mv "$tmp/$file" "$dir/$file"
  done

  # Launcher for macOS / Linux / Git Bash: picks the binary for this machine.
  cat > "$dir/compress-env" <<'EOF'
#!/bin/sh
# Runs the compress-env binary for this OS / CPU from this folder.
# Installed by: https://github.com/cavaldos/compress-env (install.sh --project)
set -eu
dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  MINGW* | MSYS* | CYGWIN*) os=windows ;;
  *) echo "error: unsupported OS '$(uname -s)'" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "error: unsupported architecture '$(uname -m)'" >&2; exit 1 ;;
esac
bin="$dir/compress-env-$os-$arch"
[ "$os" = windows ] && bin="$bin.exe"
if [ ! -f "$bin" ]; then
  echo "error: no compress-env binary for $os/$arch in $dir" >&2
  echo "reinstall with: curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | sh -s -- --project" >&2
  exit 1
fi
[ -x "$bin" ] || chmod +x "$bin" 2>/dev/null || true
exec "$bin" "$@"
EOF
  chmod +x "$dir/compress-env"

  # Launcher for Windows cmd / PowerShell.
  printf '%s\r\n' \
    '@echo off' \
    'rem Runs the compress-env binary for this CPU from this folder.' \
    'setlocal' \
    'set "ARCH=amd64"' \
    'if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"' \
    'if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64" set "ARCH=arm64"' \
    'set "BIN=%~dp0compress-env-windows-%ARCH%.exe"' \
    'if exist "%BIN%" goto run' \
    'echo error: no compress-env binary for windows/%ARCH% in %~dp0 1>&2' \
    'exit /b 1' \
    ':run' \
    '"%BIN%" %*' \
    'exit /b %ERRORLEVEL%' > "$dir/compress-env.cmd"

  # Keep line endings right whatever core.autocrlf is set to on each machine.
  printf '%s\n' \
    'compress-env text eol=lf' \
    'compress-env.cmd text eol=crlf' \
    'compress-env-* binary' > "$dir/.gitattributes"

  echo
  echo "Installed into $dir/ :"
  ls -1 "$dir" | sed 's/^/  /'
  echo
  echo "Commit it so your team can use it without installing anything:"
  echo "  git add $dir && git commit -m \"chore: add compress-env\""
  echo
  run="$dir/compress-env"
  case "$dir" in /*) ;; *) run="./$run" ;; esac
  echo "Run it from the project root:"
  echo "  $run                 (macOS / Linux / Git Bash)"
  printf '  %s   (Windows)\n' "$(printf '%s' "$run" | tr '/' '\\')"
  echo
  "$dir/compress-env" -v 2>/dev/null | head -1 || true
  exit 0
fi

# --- global install ---------------------------------------------------------

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
fetch "$file"

mkdir -p "$dir"
chmod +x "$tmp/$file"
mv "$tmp/$file" "$dir/compress-env"

echo "Installed: $dir/compress-env"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Note: $dir is not in your PATH. Add it, e.g.: export PATH=\"$dir:\$PATH\"" ;;
esac
"$dir/compress-env" -v | head -1
