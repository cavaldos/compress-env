#!/usr/bin/env bash
# Checks run by CI before anything is built or released: formatting, vet, tests, build.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> gofmt"
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "These files need gofmt:" >&2
  echo "$unformatted" >&2
  exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test ./...

echo "==> go build"
go build -o /dev/null .

echo "OK"
