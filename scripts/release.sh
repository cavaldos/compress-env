#!/usr/bin/env bash
# Tags the current commit and pushes the tag. The release workflow
# (.github/workflows/release.yml) then tests, builds and publishes it.
#
#   scripts/release.sh v0.0.1            ask before pushing
#   scripts/release.sh v0.0.1 --yes      push without asking
#   scripts/release.sh v0.0.1 --dry-run  run every check, change nothing
set -euo pipefail
cd "$(dirname "$0")/.."

tag="${1:?usage: scripts/release.sh vX.Y.Z [--yes|--dry-run]}"
mode="${2:-}"

if ! [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]]; then
  echo "error: '$tag' is not a version like v1.2.3" >&2
  exit 1
fi

branch=$(git branch --show-current)
if [ "$branch" != "main" ]; then
  echo "error: releases are cut from main, you are on '$branch'" >&2
  exit 1
fi

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  echo "error: uncommitted changes, commit or stash them first" >&2
  exit 1
fi

git fetch --quiet origin main --tags
if [ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]; then
  echo "error: main is not in sync with origin/main, push or pull first" >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
  echo "error: tag $tag already exists" >&2
  exit 1
fi

scripts/test.sh

if [ "$mode" = "--dry-run" ]; then
  echo "dry run: would tag $(git rev-parse --short HEAD) as $tag and push it"
  exit 0
fi

if [ "$mode" != "--yes" ]; then
  read -r -p "Tag $(git rev-parse --short HEAD) as $tag and push it? [y/N] " answer
  [[ "$answer" =~ ^[Yy]$ ]] || { echo "aborted"; exit 1; }
fi

git tag -a "$tag" -m "Release $tag"
git push origin "$tag"
echo "Pushed $tag. Follow the release at: $(git remote get-url origin | sed 's/\.git$//')/actions"
