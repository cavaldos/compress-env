# Installs compress-env on Windows.
#
#   irm https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.ps1 | iex
#
# Project mode puts the binaries inside the current project (default .\bin),
# so they can be committed and teammates can run .\bin\compress-env without
# installing anything:
#
#   $env:PROJECT='1'; irm https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.ps1 | iex
#
# Environment:
#   VERSION      release to install, e.g. v0.0.1 (default: latest)
#   INSTALL_DIR  where to put the binary
#                (default: %LOCALAPPDATA%\Programs\compress-env; project mode: .\bin)
#   PROJECT=1    install into the current project
#   PLATFORMS    project mode only: platforms to include, e.g. "windows-amd64 darwin-arm64"
#                (default: all of them, so it runs for everyone on the team)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = 'cavaldos/compress-env'
$base = if ($env:COMPRESS_ENV_BASE_URL) { $env:COMPRESS_ENV_BASE_URL } else { "https://github.com/$repo/releases" }
$version = if ($env:VERSION) { $env:VERSION } else { 'latest' }
$url = if ($version -eq 'latest') { "$base/latest/download" } else { "$base/download/$version" }
$project = [bool]$env:PROJECT

$tmp = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null

# Downloads a release file into $tmp, checks its checksum and returns its path.
function Get-ReleaseFile($file) {
    Write-Host "Downloading $file ($version)..."
    $path = Join-Path $tmp $file
    Invoke-WebRequest -UseBasicParsing "$url/$file" -OutFile $path
    $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($file))$" } | Select-Object -First 1
    $expected = if ($line) { ($line -split '\s+')[0] } else { '' }
    $actual = (Get-FileHash -Algorithm SHA256 $path).Hash.ToLower()
    if (-not $expected -or $expected -ne $actual) { throw "Checksum mismatch for $file, not installing" }
    return $path
}

try {
    Invoke-WebRequest -UseBasicParsing "$url/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

    if ($project) {
        $dir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { 'bin' }
        $platforms = 'darwin-amd64', 'darwin-arm64', 'linux-amd64', 'linux-arm64', 'windows-amd64', 'windows-arm64'
        if ($env:PLATFORMS) { $platforms = $env:PLATFORMS -split '\s+' | Where-Object { $_ } }
        New-Item -ItemType Directory -Force -Path $dir | Out-Null

        foreach ($p in $platforms) {
            $file = "compress-env-$p"
            if ($p -like 'windows-*') { $file += '.exe' }
            Move-Item -Force (Get-ReleaseFile $file) (Join-Path $dir $file)
        }

        # Launcher for macOS / Linux / Git Bash (LF line endings).
        $sh = @'
#!/bin/sh
# Runs the compress-env binary for this OS / CPU from this folder.
# Installed by: https://github.com/cavaldos/compress-env (install.ps1, project mode)
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
'@
        [IO.File]::WriteAllText((Join-Path (Resolve-Path $dir) 'compress-env'), ($sh -replace "`r`n", "`n"))

        # Launcher for Windows cmd / PowerShell (CRLF line endings).
        $cmd = @(
            '@echo off'
            'rem Runs the compress-env binary for this CPU from this folder.'
            'setlocal'
            'set "ARCH=amd64"'
            'if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"'
            'if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64" set "ARCH=arm64"'
            'set "BIN=%~dp0compress-env-windows-%ARCH%.exe"'
            'if exist "%BIN%" goto run'
            'echo error: no compress-env binary for windows/%ARCH% in %~dp0 1>&2'
            'exit /b 1'
            ':run'
            '"%BIN%" %*'
            'exit /b %ERRORLEVEL%'
        ) -join "`r`n"
        [IO.File]::WriteAllText((Join-Path (Resolve-Path $dir) 'compress-env.cmd'), "$cmd`r`n")

        # Keep line endings right whatever core.autocrlf is set to on each machine.
        [IO.File]::WriteAllText((Join-Path (Resolve-Path $dir) '.gitattributes'),
            "compress-env text eol=lf`ncompress-env.cmd text eol=crlf`ncompress-env-* binary`n")

        # Windows has no executable bit; record it in git so the launcher runs on macOS / Linux.
        if ((Get-Command git -ErrorAction SilentlyContinue) -and (git rev-parse --is-inside-work-tree 2>$null)) {
            git add (Join-Path $dir '*') | Out-Null
            git update-index --chmod=+x (Join-Path $dir 'compress-env') | Out-Null
            Get-ChildItem $dir -Filter 'compress-env-*' | Where-Object { $_.Extension -ne '.exe' } |
                ForEach-Object { git update-index --chmod=+x (Join-Path $dir $_.Name) | Out-Null }
        }

        Write-Host ''
        Write-Host "Installed into $dir\ :"
        Get-ChildItem $dir | ForEach-Object { Write-Host "  $($_.Name)" }
        Write-Host ''
        Write-Host 'Commit it so your team can use it without installing anything:'
        Write-Host "  git add $dir; git commit -m `"chore: add compress-env`""
        Write-Host ''
        Write-Host 'Run it from the project root:'
        Write-Host "  .\$dir\compress-env          (Windows)"
        Write-Host "  ./$($dir -replace '\\', '/')/compress-env          (macOS / Linux / Git Bash)"
        return
    }

    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { $arch = 'amd64' }
        'ARM64' { $arch = 'arm64' }
        default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }

    $dir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\compress-env' }
    $file = "compress-env-windows-$arch.exe"
    $path = Get-ReleaseFile $file

    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Move-Item -Force $path (Join-Path $dir 'compress-env.exe')

    Write-Host "Installed: $dir\compress-env.exe"

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (($userPath -split ';') -notcontains $dir) {
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User')
        Write-Host "Added $dir to your PATH. Open a new terminal to use compress-env."
    }
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
