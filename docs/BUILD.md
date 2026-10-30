# Build from source

How to build, test and release compress-env yourself. To just use it, see the [README](../README.md#install).

## Contents

- [Requirements](#requirements)
- [Build and run](#build-and-run)
- [Install your build](#install-your-build)
- [Build for every platform](#build-for-every-platform)
- [Use your build inside a project](#use-your-build-inside-a-project)
- [Tests](#tests)
- [Project layout](#project-layout)
- [Release](#release)
- [Install script options](#install-script-options)

## Requirements

- [Go](https://go.dev/dl/) **1.25 or newer** (`go version`)
- Git
- Bash, for the scripts in `scripts/` (on Windows: Git Bash or WSL)

No C compiler is needed: builds use `CGO_ENABLED=0`.

## Build and run

```bash
git clone https://github.com/cavaldos/compress-env.git
cd compress-env
go build -o compress-env .        # Windows: go build -o compress-env.exe .
./compress-env -v
```

Run without building a binary:

```bash
go run .                          # interactive menu
go run . compress -p my-secret
```

To try it on sample files without touching your own, use a `sandbox/` folder (ignored by git):

```bash
mkdir -p sandbox/api && cd sandbox
echo "API_KEY=123" > api/.env
go run ..
```

### Set the version

The version shown by `compress-env -v` comes from `main.version`. Set it at build time:

```bash
go build -trimpath -ldflags "-s -w -X main.version=0.3.0" -o compress-env .
```

`-s -w` strips debug info (smaller binary), `-trimpath` removes local paths from it.

## Install your build

**With `go install`** (puts it in `$(go env GOPATH)/bin`, usually `~/go/bin`):

```bash
go install github.com/cavaldos/compress-env@latest   # from GitHub
go install .                                         # from your clone
```

Make sure `~/go/bin` is in your `PATH`. Builds made this way show the default version, not the release tag.

**Or copy the binary** to a folder in your `PATH`:

```bash
go build -o compress-env .
sudo mv compress-env /usr/local/bin/        # or ~/.local/bin without sudo
```

On Windows, move `compress-env.exe` to a folder such as `%LOCALAPPDATA%\Programs\compress-env` and add it to your user `Path`.

## Build for every platform

```bash
scripts/build.sh v0.3.0
```

It writes to `dist/`:

```
dist/
├── compress-env-darwin-amd64        macOS Intel
├── compress-env-darwin-arm64        macOS Apple Silicon
├── compress-env-linux-amd64
├── compress-env-linux-arm64
├── compress-env-windows-amd64.exe
├── compress-env-windows-arm64.exe
└── checksums.txt                    SHA-256 of each binary
```

One platform only:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o compress-env-linux-arm64 .
```

| `GOOS`    | `GOARCH`         |
| --------- | ---------------- |
| `darwin`  | `amd64`, `arm64` |
| `linux`   | `amd64`, `arm64` |
| `windows` | `amd64`, `arm64` |

## Use your build inside a project

The project install (`install.sh --project`) downloads release binaries. To put your own build in a project instead, point the script at your `dist/` folder. It expects the release layout `<base>/download/<version>/`:

```bash
scripts/build.sh v0.3.0-dev
mkdir -p /tmp/ce-release/download/v0.3.0-dev
cp dist/* /tmp/ce-release/download/v0.3.0-dev/

cd ~/code/my-app
COMPRESS_ENV_BASE_URL=file:///tmp/ce-release VERSION=v0.3.0-dev \
  sh ~/code/compress-env/scripts/install.sh --project
```

This also writes the launchers (`bin/compress-env`, `bin/compress-env.cmd`) that pick the right binary for each machine.

## Tests

Run everything CI runs (gofmt, go vet, tests, build):

```bash
scripts/test.sh
```

Or separately:

```bash
gofmt -l .                  # lists files that need formatting
go vet ./...
go test ./...
go test ./internal/tui -run Picker -v
```

CI (`.github/workflows/ci.yml`) runs `scripts/test.sh` on Linux, macOS and Windows for every push to `main` and every pull request.

## Project layout

```
main.go                  commands (compress, restore, -h, -v)
internal/
├── compress/            finds .env files and packs them into one text
├── restore/             unpacks that text back into files
├── hash/                password encoding
├── pubkey/              public / private key encryption (age)
├── config/              .compress-env.yaml and the list of candidate files
├── tui/                 interactive menu, file tree picker, tab completion
└── help/                help and version text
scripts/
├── build.sh             cross-compile into dist/
├── test.sh              checks run by CI
├── release.sh           tag and push a release
├── install.sh           installer for macOS / Linux
└── install.ps1          installer for Windows
docs/
├── BUILD.md             this file
└── images/              README screenshots
```

## Release

Releases are built by GitHub Actions when a `v*` tag is pushed (`.github/workflows/release.yml`): it runs the tests, builds every platform and publishes `dist/*` as a GitHub release. Tags with a `-` (e.g. `v0.3.0-rc.1`) are marked as pre-releases.

```bash
scripts/release.sh v0.3.0 --dry-run   # run every check, change nothing
scripts/release.sh v0.3.0             # ask, then tag and push
scripts/release.sh v0.3.0 --yes       # tag and push without asking
```

The script refuses to release unless you are on `main`, with no uncommitted changes, in sync with `origin/main`, the tag is new and the tests pass.

The install scripts download from the latest release, so a new version is available to everyone once the workflow finishes.

## Install script options

Both `scripts/install.sh` and `scripts/install.ps1` read these environment variables.

| Variable                | Meaning                                                    | Default |
| ----------------------- | ---------------------------------------------------------- | ------- |
| `VERSION`               | Release to install, e.g. `v0.2.0`                          | latest |
| `INSTALL_DIR`           | Where to put the binary                                    | machine: `/usr/local/bin` if writable, else `~/.local/bin` (Windows: `%LOCALAPPDATA%\Programs\compress-env`) · project: `bin` |
| `PROJECT=1`             | Install into the current project (same as `--project`)     | off |
| `PLATFORMS`             | Project mode: platforms to include, space separated        | all six |
| `COMPRESS_ENV_BASE_URL` | Where releases are downloaded from                         | `https://github.com/cavaldos/compress-env/releases` |

Every download is checked against `checksums.txt`; nothing is installed on a mismatch.

Examples:

```bash
# a specific version, into ~/bin
curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | VERSION=v0.2.0 INSTALL_DIR=~/bin sh

# project mode, macOS Apple Silicon and Windows only, into tools/
curl -fsSL https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.sh | PLATFORMS="darwin-arm64 windows-amd64" INSTALL_DIR=tools sh -s -- --project
```

```powershell
$env:VERSION='v0.2.0'; $env:PROJECT='1'; irm https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.ps1 | iex
```
