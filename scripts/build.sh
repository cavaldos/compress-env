#!/usr/bin/env bash
# Cross-compiles compress-env into dist/ and writes dist/checksums.txt.
#
#   scripts/build.sh <version>     e.g. scripts/build.sh v0.0.1
#
# Binaries are named compress-env-<os>-<arch>[.exe].
set -euo pipefail
cd "$(dirname "$0")/.."

version="${1:?usage: scripts/build.sh <version>}"
version="${version#v}"

targets=(
  linux/amd64
  linux/arm64
  darwin/amd64
  darwin/arm64
  windows/amd64
  windows/arm64
)

mkdir -p dist
for target in "${targets[@]}"; do
  goos="${target%/*}"
  goarch="${target#*/}"
  out="dist/compress-env-${goos}-${goarch}"
  [ "$goos" = "windows" ] && out+=".exe"

  echo "==> ${out}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "-s -w -X main.version=${version}" -o "$out" .
done

echo "==> checksums"
(
  cd dist
  if command -v sha256sum >/dev/null; then
    sha256sum compress-env-* > checksums.txt
  else
    shasum -a 256 compress-env-* > checksums.txt
  fi
  cat checksums.txt
)
